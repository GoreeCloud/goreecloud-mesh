package mesh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDurableEventJournalRejectsDuplicateTopLevelJSONKey(t *testing.T) {
	path := writeValidJournalForStrictDecodeTest(t)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	needle := `"schema": "` + EventJournalSchemaV1 + `",`
	replacement := needle + "\n  " + needle
	modified := strings.Replace(string(body), needle, replacement, 1)
	if modified == string(body) {
		t.Fatal("test fixture did not contain expected schema key")
	}
	if err := os.WriteFile(path, []byte(modified), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewDurableEventJournal(path, 8); err == nil {
		t.Fatal("journal with duplicate top-level JSON key must fail closed")
	}
}

func TestDurableEventJournalRejectsDuplicateNestedJSONKey(t *testing.T) {
	path := writeValidJournalForStrictDecodeTest(t)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	needle := `"offset": 1,`
	replacement := needle + "\n      " + needle
	modified := strings.Replace(string(body), needle, replacement, 1)
	if modified == string(body) {
		t.Fatal("test fixture did not contain expected offset key")
	}
	if err := os.WriteFile(path, []byte(modified), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewDurableEventJournal(path, 8); err == nil {
		t.Fatal("journal with duplicate nested JSON key must fail closed")
	}
}

func TestDurableSubscriberCheckpointsRejectDuplicateSubscriberKey(t *testing.T) {
	dir := t.TempDir()
	journal, err := NewDurableEventJournal(filepath.Join(dir, "events.json"), 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.Append(journalEvent(t, 1)); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "subscribers.json")
	store, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Acknowledge("goreecloud-manager", 1, journal); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	needle := `"goreecloud-manager": 1`
	replacement := needle + ",\n    " + needle
	modified := strings.Replace(string(body), needle, replacement, 1)
	if modified == string(body) {
		t.Fatal("test fixture did not contain expected subscriber key")
	}
	if err := os.WriteFile(path, []byte(modified), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewDurableSubscriberCheckpoints(path); err == nil {
		t.Fatal("checkpoint state with duplicate subscriber key must fail closed")
	}
}
