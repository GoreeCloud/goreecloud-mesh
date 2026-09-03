package contracts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// NewPersistentPlatformRecordRegistry loads source-attributed Platform Contract
// coordination records from path. An empty path keeps the registry process-local.
func NewPersistentPlatformRecordRegistry(path string) (*PlatformRecordRegistry, error) {
	r := NewPlatformRecordRegistry()
	r.path = path
	if path == "" {
		return r, nil
	}

	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return r, nil
	}

	var persisted []PlatformRecord
	if err := json.Unmarshal(b, &persisted); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	evaluatedAt := time.Now().UTC()
	for _, record := range persisted {
		validated, err := normalizePlatformRecord(record, evaluatedAt)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[validated.ComponentID]; ok {
			return nil, errors.New("persisted platform records contain duplicate component_id")
		}
		seen[validated.ComponentID] = struct{}{}
		r.records[validated.ComponentID] = validated
	}
	return r, nil
}

func (r *PlatformRecordRegistry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil && filepath.Dir(r.path) != "." {
		return err
	}
	items := make([]PlatformRecord, 0, len(r.records))
	for _, record := range r.records {
		items = append(items, copyPlatformRecord(record))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ComponentID < items[j].ComponentID })
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}
