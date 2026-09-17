package mesh

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDurableSubscriberCheckpointPersistsAndSurvivesRestart(t *testing.T) {
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

	path := filepath.Join(dir, "subscribers.json")
	store, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Acknowledge("goreecloud-manager", 2, journal); err != nil {
		t.Fatal(err)
	}

	restarted, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, ok, err := restarted.Checkpoint("goreecloud-manager")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || checkpoint != 2 {
		t.Fatalf("checkpoint = %d ok=%v, want 2 true", checkpoint, ok)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("subscriber checkpoint state must be private regular file: %v %04o", info.Mode(), info.Mode().Perm())
	}
}

func TestDurableSubscriberCheckpointIsMonotonicAndIdempotent(t *testing.T) {
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
	store, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "subscribers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Acknowledge("privacy-shield", 2, journal); err != nil {
		t.Fatal(err)
	}
	if err := store.Acknowledge("privacy-shield", 2, journal); err != nil {
		t.Fatalf("idempotent acknowledgement failed: %v", err)
	}
	if err := store.Acknowledge("privacy-shield", 1, journal); !errors.Is(err, ErrSubscriberCheckpointRegression) {
		t.Fatalf("regression error = %v", err)
	}
}

func TestDurableSubscriberCheckpointRejectsAheadAndEvictedHistory(t *testing.T) {
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
	store, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "subscribers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Acknowledge("everkeep", 0, journal); !errors.Is(err, ErrEventCheckpointTooOld) {
		t.Fatalf("old checkpoint error = %v", err)
	}
	if err := store.Acknowledge("everkeep", 4, journal); !errors.Is(err, ErrEventCheckpointAhead) {
		t.Fatalf("ahead checkpoint error = %v", err)
	}
	if err := store.Acknowledge("everkeep", 1, journal); err != nil {
		t.Fatalf("checkpoint immediately before retained history must remain valid: %v", err)
	}
}

func TestDurableSubscriberCheckpointRejectsNoncanonicalSubscriberIDAndSymlinkState(t *testing.T) {
	dir := t.TempDir()
	journal, err := NewDurableEventJournal(filepath.Join(dir, "events.json"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.Append(journalEvent(t, 1)); err != nil {
		t.Fatal(err)
	}
	store, err := NewDurableSubscriberCheckpoints(filepath.Join(dir, "subscribers.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, subscriberID := range []string{"", " manager", "manager\n"} {
		if err := store.Acknowledge(subscriberID, 0, journal); err == nil {
			t.Fatalf("noncanonical subscriber ID %q must fail", subscriberID)
		}
	}

	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte(`{"schema":"goreecloud.mesh.subscriber-checkpoints.v1","checkpoints":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "linked-subscribers.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDurableSubscriberCheckpoints(link); err == nil {
		t.Fatal("subscriber checkpoint store must reject symlink state")
	}
}
