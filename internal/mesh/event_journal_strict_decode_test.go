package mesh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeValidJournalForStrictDecodeTest(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.json")
	journal, err := NewDurableEventJournal(path, 8)
	if err != nil {
		t.Fatalf("NewDurableEventJournal: %v", err)
	}
	event, err := newEvent(
		1,
		EventServiceUpsertedV1,
		"goreecloud-sync",
		"goreecloud-sync",
		map[string]any{"health": "healthy"},
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("newEvent: %v", err)
	}
	if _, err := journal.Append(event); err != nil {
		t.Fatalf("Append: %v", err)
	}
	return path
}

func TestDurableEventJournalRejectsUnknownPersistedFields(t *testing.T) {
	path := writeValidJournalForStrictDecodeTest(t)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	modified := strings.Replace(string(body), "{\n", "{\n  \"unexpected\": true,\n", 1)
	if err := os.WriteFile(path, []byte(modified), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := NewDurableEventJournal(path, 8); err == nil {
		t.Fatal("journal with unknown persisted field must fail closed")
	}
}

func TestDurableEventJournalRejectsTrailingJSONValue(t *testing.T) {
	path := writeValidJournalForStrictDecodeTest(t)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if _, err := file.WriteString("\n{}\n"); err != nil {
		_ = file.Close()
		t.Fatalf("append trailing JSON: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := NewDurableEventJournal(path, 8); err == nil {
		t.Fatal("journal with trailing JSON value must fail closed")
	}
}
