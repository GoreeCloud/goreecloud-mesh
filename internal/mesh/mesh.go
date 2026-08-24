package mesh

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
)

type Mesh struct {
	store *store.Store
	seq   atomic.Uint64
	mu    sync.RWMutex
	subs  map[chan model.Event]struct{}
}

func New(s *store.Store) *Mesh {
	return &Mesh{store: s, subs: map[chan model.Event]struct{}{}}
}

func (m *Mesh) State() model.State { return m.store.Snapshot() }

func (m *Mesh) RegisterService(v model.Service) (model.Service, error) {
	v.ID = strings.TrimSpace(v.ID)
	v.Name = strings.TrimSpace(v.Name)
	v.Kind = strings.TrimSpace(v.Kind)
	if v.ID == "" || v.Name == "" || v.Kind == "" {
		return model.Service{}, errors.New("id, name, and kind are required")
	}
	if v.Health == "" {
		v.Health = model.HealthUnknown
	}
	if !validHealth(v.Health) {
		return model.Service{}, fmt.Errorf("invalid health state %q", v.Health)
	}
	v.Capabilities = unique(v.Capabilities)
	v.Dependencies = unique(v.Dependencies)
	v.UpdatedAt = time.Now().UTC()
	if err := m.store.PutService(v); err != nil {
		return model.Service{}, err
	}
	m.publish("mesh.service.upserted", v.ID, v.ID, map[string]any{"health": v.Health})
	return v, nil
}

func (m *Mesh) Service(id string) (model.Service, bool) { return m.store.Service(id) }

func (m *Mesh) Services() []model.Service {
	s := m.store.Snapshot()
	out := make([]model.Service, 0, len(s.Services))
	for _, v := range s.Services {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *Mesh) Discover(capability string) []model.Service {
	capability = strings.TrimSpace(capability)
	out := []model.Service{}
	for _, s := range m.Services() {
		if s.Health == model.HealthUnavailable {
			continue
		}
		for _, c := range s.Capabilities {
			if c == capability {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func (m *Mesh) AddRelationship(r model.Relationship) (model.Relationship, error) {
	r.ID = strings.TrimSpace(r.ID)
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
	r.Type = strings.TrimSpace(r.Type)
	if r.ID == "" || r.From == "" || r.To == "" || r.Type == "" {
		return model.Relationship{}, errors.New("id, from, to, and type are required")
	}
	if r.From == r.To {
		return model.Relationship{}, errors.New("self relationships are not allowed")
	}
	if _, ok := m.store.Service(r.From); !ok {
		return model.Relationship{}, fmt.Errorf("source service %q is not registered", r.From)
	}
	if _, ok := m.store.Service(r.To); !ok {
		return model.Relationship{}, fmt.Errorf("target service %q is not registered", r.To)
	}
	r.UpdatedAt = time.Now().UTC()
	if err := m.store.PutRelationship(r); err != nil {
		return model.Relationship{}, err
	}
	m.publish("mesh.relationship.upserted", r.From, r.ID, map[string]any{"target": r.To, "type": r.Type})
	return r, nil
}

func (m *Mesh) Evaluate(req model.PolicyRequest) model.PolicyDecision {
	source, ok := m.store.Service(strings.TrimSpace(req.Source))
	if !ok {
		return model.PolicyDecision{Reason: "source is not registered"}
	}
	target, ok := m.store.Service(strings.TrimSpace(req.Target))
	if !ok {
		return model.PolicyDecision{Reason: "target is not registered"}
	}
	if target.Health == model.HealthUnavailable {
		return model.PolicyDecision{Reason: "target is unavailable"}
	}
	if req.Capability == "" {
		return model.PolicyDecision{Reason: "capability is required"}
	}
	capable := false
	for _, c := range target.Capabilities {
		if c == req.Capability {
			capable = true
			break
		}
	}
	if !capable {
		return model.PolicyDecision{Reason: "target does not advertise requested capability"}
	}
	state := m.store.Snapshot()
	for _, r := range state.Relationships {
		if r.Enabled && r.From == source.ID && r.To == target.ID && (r.Capability == "" || r.Capability == req.Capability) {
			return model.PolicyDecision{Allowed: true, Reason: "enabled relationship authorizes capability"}
		}
	}
	return model.PolicyDecision{Reason: "no enabled relationship authorizes capability"}
}

// Impact returns registered services that directly or transitively depend on id.
func (m *Mesh) Impact(id string) []string {
	state := m.store.Snapshot()
	reverse := map[string][]string{}
	for _, s := range state.Services {
		for _, dependency := range s.Dependencies {
			reverse[dependency] = append(reverse[dependency], s.ID)
		}
	}
	for _, r := range state.Relationships {
		if r.Enabled && r.Required {
			reverse[r.To] = append(reverse[r.To], r.From)
		}
	}
	seen := map[string]bool{id: true}
	queue := []string{id}
	out := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range reverse[current] {
			if seen[dependent] {
				continue
			}
			seen[dependent] = true
			out = append(out, dependent)
			queue = append(queue, dependent)
		}
	}
	sort.Strings(out)
	return out
}

func (m *Mesh) Subscribe(buffer int) (<-chan model.Event, func()) {
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan model.Event, buffer)
	m.mu.Lock()
	m.subs[ch] = struct{}{}
	m.mu.Unlock()
	cancel := func() {
		m.mu.Lock()
		if _, ok := m.subs[ch]; ok {
			delete(m.subs, ch)
			close(ch)
		}
		m.mu.Unlock()
	}
	return ch, cancel
}

func (m *Mesh) publish(kind, source, subject string, data map[string]any) {
	e := model.Event{ID: fmt.Sprintf("evt-%d", m.seq.Add(1)), Type: kind, Source: source, Subject: subject, Data: data, CreatedAt: time.Now().UTC()}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for ch := range m.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func validHealth(v model.HealthState) bool {
	switch v {
	case model.HealthUnknown, model.HealthHealthy, model.HealthDegraded, model.HealthUnavailable:
		return true
	default:
		return false
	}
}
