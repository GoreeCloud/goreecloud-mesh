package mesh

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestReplaySubscriberAcknowledgesOnlySuccessfulEvents(t *testing.T) {
	dir := t.TempDir()
	journal, err := NewDurableEventJournal(filepath.Join(dir, "events.json"), 8)
	if err != nil {
		t.Fatal(err)
	}
	for sequence := uint64(1); sequence <= 3; sequence++ {
		if _, err := journal.Append(journalEvent(t, sequence)); err != nil {
			t.Fatal(err)
		}
	}
	checkpoints, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "checkpoints.json"))
	if err != nil {
		t.Fatal(err)
	}

	seen := []uint64{}
	stopErr := errors.New("subscriber processing failed")
	processed, checkpoint, err := ReplaySubscriber(
		context.Background(),
		"manager-event-consumer",
		3,
		journal,
		checkpoints,
		func(_ context.Context, record DurableEventRecord) error {
			seen = append(seen, record.Offset)
			if record.Offset == 2 {
				return stopErr
			}
			return nil
		},
	)
	if !errors.Is(err, stopErr) {
		t.Fatalf("expected handler failure, got %v", err)
	}
	if processed != 1 || checkpoint != 1 {
		t.Fatalf("processed=%d checkpoint=%d, want 1/1", processed, checkpoint)
	}
	stored, ok, err := checkpoints.Checkpoint("manager-event-consumer")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || stored != 1 {
		t.Fatalf("stored checkpoint=%d ok=%v, want 1/true", stored, ok)
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("unexpected handler sequence: %v", seen)
	}

	seen = nil
	processed, checkpoint, err = ReplaySubscriber(
		context.Background(),
		"manager-event-consumer",
		3,
		journal,
		checkpoints,
		func(_ context.Context, record DurableEventRecord) error {
			seen = append(seen, record.Offset)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if processed != 2 || checkpoint != 3 {
		t.Fatalf("processed=%d checkpoint=%d, want 2/3", processed, checkpoint)
	}
	if len(seen) != 2 || seen[0] != 2 || seen[1] != 3 {
		t.Fatalf("failed event was not replayed before later event: %v", seen)
	}
}

func TestReplaySubscriberFailsClosedWhenCheckpointFallsBehindRetention(t *testing.T) {
	dir := t.TempDir()
	journal, err := NewDurableEventJournal(filepath.Join(dir, "events.json"), 2)
	if err != nil {
		t.Fatal(err)
	}
	for sequence := uint64(1); sequence <= 3; sequence++ {
		if _, err := journal.Append(journalEvent(t, sequence)); err != nil {
			t.Fatal(err)
		}
	}
	checkpoints, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "checkpoints.json"))
	if err != nil {
		t.Fatal(err)
	}

	called := false
	_, _, err = ReplaySubscriber(
		context.Background(),
		"stale-consumer",
		2,
		journal,
		checkpoints,
		func(context.Context, DurableEventRecord) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, ErrEventCheckpointTooOld) {
		t.Fatalf("expected checkpoint-too-old failure, got %v", err)
	}
	if called {
		t.Fatal("handler ran despite missing retained history")
	}
}

func TestReplaySubscriberSerializesCheckpointOwnershipWithinRuntime(t *testing.T) {
	dir := t.TempDir()
	journal, err := NewDurableEventJournal(filepath.Join(dir, "events.json"), 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.Append(journalEvent(t, 1)); err != nil {
		t.Fatal(err)
	}
	checkpoints, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "checkpoints.json"))
	if err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, _, err := ReplaySubscriber(
			context.Background(),
			"manager-event-consumer",
			1,
			journal,
			checkpoints,
			func(context.Context, DurableEventRecord) error {
				close(started)
				<-release
				return nil
			},
		)
		firstDone <- err
	}()

	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	secondCalled := false
	processed, checkpoint, err := ReplaySubscriber(
		ctx,
		"manager-event-consumer",
		1,
		journal,
		checkpoints,
		func(context.Context, DurableEventRecord) error {
			secondCalled = true
			return nil
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		close(release)
		t.Fatalf("expected context deadline while replay ownership was held, got %v", err)
	}
	if processed != 0 || checkpoint != 0 {
		close(release)
		t.Fatalf("processed=%d checkpoint=%d, want 0/0 while replay ownership is held", processed, checkpoint)
	}
	if secondCalled {
		close(release)
		t.Fatal("second replay handler ran while another replay pass owned checkpoint progression")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first replay failed: %v", err)
	}
	stored, ok, err := checkpoints.Checkpoint("manager-event-consumer")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || stored != 1 {
		t.Fatalf("stored checkpoint=%d ok=%v, want 1/true", stored, ok)
	}
}
