package contracts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// NewPersistentRegistry loads runtime contract evidence from path. An empty
// path keeps runtime evidence process-local.
func NewPersistentRegistry(path string) (*Registry, error) {
	r := NewRegistry()
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

	var persisted []Evidence
	if err := json.Unmarshal(b, &persisted); err != nil {
		return nil, err
	}
	seen := map[Platform]struct{}{}
	for _, evidence := range persisted {
		if _, ok := seen[evidence.Platform]; ok {
			return nil, errors.New("persisted runtime evidence contains duplicate platform")
		}
		seen[evidence.Platform] = struct{}{}
		validated, err := normalizeEvidence(evidence)
		if err != nil {
			return nil, err
		}
		r.evidence[validated.Platform] = validated
	}
	return r, nil
}

func (r *Registry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil && filepath.Dir(r.path) != "." {
		return err
	}
	items := make([]Evidence, 0, len(r.evidence))
	for _, platform := range Mandatory() {
		if evidence, ok := r.evidence[platform]; ok {
			items = append(items, evidence)
		}
	}
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
