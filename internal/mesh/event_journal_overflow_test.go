package mesh

import (
	"errors"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

func TestDurableEventJournalFailsClosedBeforeOffsetWraparound(t *testing.T) {
	journal := &DurableEventJournal{
		path:       filepath.Join(t.TempDir(), "events.json"),
		maxEntries: 4,
		nextOffset: math.MaxUint64,
		entries:    []DurableEventRecord{},
	}
	event := model.Event{
		Schema:    EventSchemaV1,
		ID:        "evt-1",
		Type:      EventServiceUpsertedV1,
		Source:    "identity",
		Subject:   "identity",
		Data:      map[string]any{"health": "healthy"},
		CreatedAt: time.Now().UTC(),
	}

	_, err := journal.Append(event)
	if !errors.Is(err, ErrEventOffsetExhausted) {
		t.Fatalf("Append error = %v, want %v", err, ErrEventOffsetExhausted)
	}
	if journal.nextOffset != math.MaxUint64 {
		t.Fatalf("next offset changed after rejected append: %d", journal.nextOffset)
	}
	if len(journal.entries) != 0 {
		t.Fatalf("journal mutated after rejected append: %+v", journal.entries)
	}
}
