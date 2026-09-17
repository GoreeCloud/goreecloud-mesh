package mesh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDurableEventJournalRejectsAdvancedOffsetWithoutRetainedHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	body := []byte(`{
  "schema": "goreecloud.mesh.event-journal.v1",
  "next_offset": 2,
  "max_entries": 8,
  "entries": []
}
`)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := NewDurableEventJournal(path, 8); err == nil {
		t.Fatal("advanced journal offset without retained history must fail closed")
	}
}

func TestDurableEventJournalAcceptsPristineEmptyPersistedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	body := []byte(`{
  "schema": "goreecloud.mesh.event-journal.v1",
  "next_offset": 1,
  "max_entries": 8,
  "entries": []
}
`)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	journal, err := NewDurableEventJournal(path, 8)
	if err != nil {
		t.Fatalf("pristine persisted journal must remain valid: %v", err)
	}
	first, last, ok := journal.Bounds()
	if ok || first != 0 || last != 0 {
		t.Fatalf("unexpected pristine bounds: first=%d last=%d ok=%v", first, last, ok)
	}
}
