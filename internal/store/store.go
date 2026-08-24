package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

type Store struct {
	mu   sync.RWMutex
	path string
	data model.State
}

func New(path string) (*Store, error) {
	s := &Store{path: path, data: model.State{Services: map[string]model.Service{}, Relationships: map[string]model.Relationship{}}}
	if path == "" {
		return s, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, err
	}
	if s.data.Services == nil {
		s.data.Services = map[string]model.Service{}
	}
	if s.data.Relationships == nil {
		s.data.Relationships = map[string]model.Relationship{}
	}
	return s, nil
}

func (s *Store) Snapshot() model.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := model.State{Services: map[string]model.Service{}, Relationships: map[string]model.Relationship{}}
	for k, v := range s.data.Services {
		out.Services[k] = v
	}
	for k, v := range s.data.Relationships {
		out.Relationships[k] = v
	}
	return out
}

func (s *Store) Service(id string) (model.Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data.Services[id]
	return v, ok
}

func (s *Store) PutService(v model.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Services[v.ID] = v
	return s.persistLocked()
}

func (s *Store) PutRelationship(v model.Relationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Relationships[v.ID] = v
	return s.persistLocked()
}

func (s *Store) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil && filepath.Dir(s.path) != "." {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
