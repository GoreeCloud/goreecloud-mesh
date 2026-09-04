package platformregistry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	StateSchema       = "goreecloud.mesh.platform-registry-state.v1"
	maxStateFileBytes = 4 << 20
)

type persistentState struct {
	Schema  string   `json:"schema"`
	Records []Record `json:"records"`
}

// PersistentRegistry stores only normalized platform coordination metadata.
// It does not persist producer payloads, credentials, tokens, or authority.
type PersistentRegistry struct {
	mu       sync.RWMutex
	path     string
	registry *Registry
}

func NewPersistent(path string) (*PersistentRegistry, error) {
	p := &PersistentRegistry{path: filepath.Clean(path), registry: New()}
	if path == "" {
		p.path = ""
		return p, nil
	}
	info, err := os.Lstat(p.path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, errors.New("platform registry state must be a regular non-symlink file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, errors.New("platform registry state must not be accessible to group or other users")
	}
	if info.Size() > maxStateFileBytes {
		return nil, errors.New("platform registry state exceeds the maximum supported size")
	}
	if info.Size() == 0 {
		return p, nil
	}
	raw, err := os.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	var state persistentState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode platform registry state: %w", err)
	}
	if state.Schema != StateSchema {
		return nil, fmt.Errorf("unsupported platform registry state schema %q", state.Schema)
	}
	seen := map[string]struct{}{}
	for _, record := range state.Records {
		if _, exists := seen[record.Component.ID]; exists {
			return nil, fmt.Errorf("duplicate persisted component id %q", record.Component.ID)
		}
		if err := p.registry.Upsert(record); err != nil {
			return nil, fmt.Errorf("invalid persisted platform record %q: %w", record.Component.ID, err)
		}
		seen[record.Component.ID] = struct{}{}
	}
	return p, nil
}

func (p *PersistentRegistry) Upsert(record Record) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	next := New()
	for _, existing := range p.registry.List() {
		if err := next.Upsert(existing); err != nil {
			return err
		}
	}
	if err := next.Upsert(record); err != nil {
		return err
	}
	if err := p.persist(next.List()); err != nil {
		return err
	}
	p.registry = next
	return nil
}

func (p *PersistentRegistry) Get(id string) (Record, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.registry.Get(id)
}

func (p *PersistentRegistry) List() []Record {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.registry.List()
}

func (p *PersistentRegistry) Dependents(id string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.registry.Dependents(id)
}

func (p *PersistentRegistry) persist(records []Record) error {
	if p.path == "" {
		return nil
	}
	dir := filepath.Dir(p.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(persistentState{Schema: StateSchema, Records: records}, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if len(raw) > maxStateFileBytes {
		return errors.New("platform registry state exceeds the maximum supported size")
	}
	temporary, err := os.CreateTemp(dir, ".mesh-platform-registry-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(raw); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, p.path); err != nil {
		return err
	}
	return os.Chmod(p.path, 0o600)
}
