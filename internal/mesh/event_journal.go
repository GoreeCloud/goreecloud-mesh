package mesh

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

const EventJournalSchemaV1 = "goreecloud.mesh.event-journal.v1"

var (
	ErrEventCheckpointTooOld = errors.New("event checkpoint is older than retained history")
	ErrEventCheckpointAhead  = errors.New("event checkpoint is ahead of the journal")
)

type DurableEventRecord struct {
	Offset uint64      `json:"offset"`
	Event  model.Event `json:"event"`
}

type durableEventJournalState struct {
	Schema     string               `json:"schema"`
	NextOffset uint64               `json:"next_offset"`
	MaxEntries int                  `json:"max_entries"`
	Entries    []DurableEventRecord `json:"entries"`
}

type DurableEventJournal struct {
	mu         sync.RWMutex
	path       string
	maxEntries int
	nextOffset uint64
	entries    []DurableEventRecord
}

func NewDurableEventJournal(path string, maxEntries int) (*DurableEventJournal, error) {
	if path == "" {
		return nil, errors.New("durable event journal path is required")
	}
	if maxEntries < 1 || maxEntries > 100000 {
		return nil, errors.New("durable event journal max entries must be between 1 and 100000")
	}
	journal := &DurableEventJournal{
		path:       path,
		maxEntries: maxEntries,
		nextOffset: 1,
		entries:    []DurableEventRecord{},
	}
	if err := journal.load(); err != nil {
		return nil, err
	}
	return journal, nil
}

func (j *DurableEventJournal) load() error {
	body, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("durable event journal is empty or corrupt")
	}

	var state durableEventJournalState
	if err := json.Unmarshal(body, &state); err != nil {
		return fmt.Errorf("decode durable event journal: %w", err)
	}
	if state.Schema != EventJournalSchemaV1 {
		return fmt.Errorf("unsupported durable event journal schema %q", state.Schema)
	}
	if state.MaxEntries != j.maxEntries {
		return fmt.Errorf("durable event journal retention changed from %d to %d", state.MaxEntries, j.maxEntries)
	}
	if state.NextOffset == 0 {
		return errors.New("durable event journal next offset must be positive")
	}
	if len(state.Entries) > j.maxEntries {
		return errors.New("durable event journal exceeds configured retention")
	}

	var previous uint64
	for index, record := range state.Entries {
		if record.Offset == 0 || (index > 0 && record.Offset != previous+1) {
			return errors.New("durable event journal offsets are not contiguous")
		}
		if err := ValidateEvent(record.Event); err != nil {
			return fmt.Errorf("durable event journal contains invalid event at offset %d: %w", record.Offset, err)
		}
		previous = record.Offset
	}
	if len(state.Entries) > 0 && state.NextOffset != state.Entries[len(state.Entries)-1].Offset+1 {
		return errors.New("durable event journal next offset does not follow retained history")
	}

	j.nextOffset = state.NextOffset
	j.entries = append([]DurableEventRecord(nil), state.Entries...)
	return nil
}

func (j *DurableEventJournal) Append(event model.Event) (DurableEventRecord, error) {
	if err := ValidateEvent(event); err != nil {
		return DurableEventRecord{}, err
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	record := DurableEventRecord{Offset: j.nextOffset, Event: cloneEvent(event)}
	candidate := append(append([]DurableEventRecord(nil), j.entries...), record)
	if len(candidate) > j.maxEntries {
		candidate = append([]DurableEventRecord(nil), candidate[len(candidate)-j.maxEntries:]...)
	}
	state := durableEventJournalState{
		Schema:     EventJournalSchemaV1,
		NextOffset: j.nextOffset + 1,
		MaxEntries: j.maxEntries,
		Entries:    candidate,
	}
	if err := j.persist(state); err != nil {
		return DurableEventRecord{}, err
	}

	j.entries = candidate
	j.nextOffset++
	return cloneDurableEventRecord(record), nil
}

func (j *DurableEventJournal) ReplayAfter(checkpoint uint64, limit int) ([]DurableEventRecord, error) {
	if limit < 1 || limit > 1000 {
		return nil, errors.New("replay limit must be between 1 and 1000")
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	lastOffset := j.nextOffset - 1
	if checkpoint > lastOffset {
		return nil, ErrEventCheckpointAhead
	}
	if len(j.entries) == 0 {
		return []DurableEventRecord{}, nil
	}

	firstOffset := j.entries[0].Offset
	if checkpoint+1 < firstOffset {
		return nil, ErrEventCheckpointTooOld
	}

	results := make([]DurableEventRecord, 0, limit)
	for _, record := range j.entries {
		if record.Offset <= checkpoint {
			continue
		}
		results = append(results, cloneDurableEventRecord(record))
		if len(results) == limit {
			break
		}
	}
	return results, nil
}

func (j *DurableEventJournal) Bounds() (first uint64, last uint64, ok bool) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if len(j.entries) == 0 {
		return 0, j.nextOffset - 1, false
	}
	return j.entries[0].Offset, j.entries[len(j.entries)-1].Offset, true
}

func (j *DurableEventJournal) persist(state durableEventJournalState) error {
	dir := filepath.Dir(j.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	tmp := j.path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, j.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func cloneEvent(event model.Event) model.Event {
	cloned := event
	if event.Data != nil {
		cloned.Data = make(map[string]any, len(event.Data))
		for key, value := range event.Data {
			cloned.Data[key] = value
		}
	}
	return cloned
}

func cloneDurableEventRecord(record DurableEventRecord) DurableEventRecord {
	return DurableEventRecord{Offset: record.Offset, Event: cloneEvent(record.Event)}
}
