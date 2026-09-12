package mesh

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func journalEvent(t *testing.T, sequence uint64) model.Event {
	t.Helper()
	event, err := newEvent(
		sequence,
		EventServiceUpsertedV1,
		"identity",
		"identity",
		map[string]any{"health": "healthy"},
		time.Date(2026, 9, 12, 4, int(sequence), 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func TestDurableEventJournalSurvivesRestartAndReplaysFromCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	journal, err := NewDurableEventJournal(path, 8)
	if err != nil {
		t.Fatal(err)
	}

	for sequence := uint64(1); sequence <= 3; sequence++ {
		record, err := journal.Append(journalEvent(t, sequence))
		if err != nil {
			t.Fatal(err)
		}
		if record.Offset != sequence {
			t.Fatalf("offset = %d, want %d", record.Offset, sequence)
		}
	}

	restarted, err := NewDurableEventJournal(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := restarted.ReplayAfter(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(replayed) != 2 || replayed[0].Offset != 2 || replayed[1].Offset != 3 {
		t.Fatalf("unexpected replay: %#v", replayed)
	}

	record, err := restarted.Append(journalEvent(t, 4))
	if err != nil {
		t.Fatal(err)
	}
	if record.Offset != 4 {
		t.Fatalf("restart reused or skipped durable offset: %d", record.Offset)
	}
}

func TestDurableEventJournalRetentionFailsClosedForOldCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	journal, err := NewDurableEventJournal(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	for sequence := uint64(1); sequence <= 3; sequence++ {
		if _, err := journal.Append(journalEvent(t, sequence)); err != nil {
			t.Fatal(err)
		}
	}

	first, last, ok := journal.Bounds()
	if !ok || first != 2 || last != 3 {
		t.Fatalf("unexpected retained bounds: first=%d last=%d ok=%v", first, last, ok)
	}
	if _, err := journal.ReplayAfter(0, 10); !errors.Is(err, ErrEventCheckpointTooOld) {
		t.Fatalf("old checkpoint error = %v", err)
	}
	replayed, err := journal.ReplayAfter(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(replayed) != 2 || replayed[0].Offset != 2 || replayed[1].Offset != 3 {
		t.Fatalf("unexpected retained replay: %#v", replayed)
	}
}

func TestDurableEventJournalRejectsAheadCheckpoint(t *testing.T) {
	journal, err := NewDurableEventJournal(filepath.Join(t.TempDir(), "events.json"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.Append(journalEvent(t, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := journal.ReplayAfter(2, 4); !errors.Is(err, ErrEventCheckpointAhead) {
		t.Fatalf("ahead checkpoint error = %v", err)
	}
}

func TestDurableEventJournalFailsClosedOnCorruptOrRetentionMismatchedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDurableEventJournal(path, 4); err == nil {
		t.Fatal("corrupt durable event state must fail closed")
	}

	path = filepath.Join(t.TempDir(), "retention.json")
	journal, err := NewDurableEventJournal(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.Append(journalEvent(t, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDurableEventJournal(path, 3); err == nil {
		t.Fatal("retention semantic drift must require explicit migration")
	}
}

func TestDurableEventJournalRejectsInvalidEventAndDefensiveCopiesData(t *testing.T) {
	journal, err := NewDurableEventJournal(filepath.Join(t.TempDir(), "events.json"), 4)
	if err != nil {
		t.Fatal(err)
	}

	invalid := journalEvent(t, 1)
	invalid.AuthorityTransfer = true
	if _, err := journal.Append(invalid); err == nil {
		t.Fatal("journal must reject event authority transfer")
	}

	valid := journalEvent(t, 1)
	record, err := journal.Append(valid)
	if err != nil {
		t.Fatal(err)
	}
	valid.Data["health"] = "unavailable"
	if record.Event.Data["health"] != "healthy" {
		t.Fatal("journal append result must not alias caller event data")
	}
	replayed, err := journal.ReplayAfter(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	replayed[0].Event.Data["health"] = "degraded"
	replayedAgain, err := journal.ReplayAfter(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if replayedAgain[0].Event.Data["health"] != "healthy" {
		t.Fatal("replay result must not mutate retained journal state")
	}
}
