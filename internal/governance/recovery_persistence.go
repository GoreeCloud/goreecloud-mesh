package governance

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// NewPersistentRecoveryRegistry loads Everkeep recovery evidence from path.
// An empty path keeps recovery evidence process-local.
func NewPersistentRecoveryRegistry(path string, at time.Time) (*RecoveryRegistry, error) {
	r := NewRecoveryRegistry()
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

	var persisted []RecoveryEvidence
	if err := json.Unmarshal(b, &persisted); err != nil {
		return nil, err
	}
	seen := map[RecoveryDimension]struct{}{}
	for _, evidence := range persisted {
		if _, ok := seen[evidence.Dimension]; ok {
			return nil, errors.New("persisted recovery evidence contains duplicate dimension")
		}
		seen[evidence.Dimension] = struct{}{}
		if err := evidence.Validate(at); err != nil {
			return nil, err
		}
		r.evidence[evidence.Dimension] = evidence
	}
	return r, nil
}

func (r *RecoveryRegistry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil && filepath.Dir(r.path) != "." {
		return err
	}
	items := make([]RecoveryEvidence, 0, len(r.evidence))
	for _, dimension := range RequiredRecoveryDimensions() {
		if evidence, ok := r.evidence[dimension]; ok {
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
