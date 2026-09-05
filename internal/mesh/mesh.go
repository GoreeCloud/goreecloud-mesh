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

const maxEventSubscriberBuffer = 64

type eventSubscriber struct {
	ch       chan model.Event
	allTypes bool
	types    map[string]struct{}
}

type Mesh struct {
	store *store.Store
	seq   atomic.Uint64
	mu    sync.RWMutex
	subs  map[chan model.Event]eventSubscriber
}

func New(s *store.Store) *Mesh {
	return &Mesh{store: s, subs: map[chan model.Event]eventSubscriber{}}
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
	event, err := newEvent(
		m.seq.Add(1),
		EventServiceUpsertedV1,
		v.ID,
		v.ID,
		map[string]any{"health": string(v.Health)},
		v.UpdatedAt,
	)
	if err != nil {
		return model.Service{}, fmt.Errorf("build service lifecycle event: %w", err)
	}
	if err := m.store.PutService(v); err != nil {
		return model.Service{}, err
	}
	m.dispatch(event)
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
	event, err := newEvent(
		m.seq.Add(1),
		EventRelationshipUpsertedV1,
		r.From,
		r.ID,
		map[string]any{"target": r.To, "type": r.Type},
		r.UpdatedAt,
	)
	if err != nil {
		return model.Relationship{}, fmt.Errorf("build relationship lifecycle event: %w", err)
	}
	if err := m.store.PutRelationship(r); err != nil {
		return model.Relationship{}, err
	}
	m.dispatch(event)
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

// Subscribe preserves the original local subscription behavior: the caller
// receives every currently registered lifecycle event type. The buffer is
// always clamped to the local event-bus ceiling so an in-process consumer
// cannot request unbounded channel memory.
func (m *Mesh) Subscribe(buffer int) (<-chan model.Event, func()) {
	return m.subscribe(buffer, true, nil)
}

// SubscribeTypes lets an in-process consumer minimize which lifecycle event
// types it receives. This is a privacy/minimization boundary, not consumer
// authentication: external or cross-process subscriptions still require a
// separate GoreeCloud Identity-authenticated transport milestone.
func (m *Mesh) SubscribeTypes(buffer int, eventTypes ...string) (<-chan model.Event, func(), error) {
	if len(eventTypes) == 0 {
		return nil, nil, errors.New("at least one event type is required")
	}
	types := make(map[string]struct{}, len(eventTypes))
	for _, eventType := range eventTypes {
		eventType = strings.TrimSpace(eventType)
		if !validEventType(eventType) {
			return nil, nil, fmt.Errorf("unsupported subscription event type %q", eventType)
		}
		types[eventType] = struct{}{}
	}
	ch, cancel := m.subscribe(buffer, false, types)
	return ch, cancel, nil
}

func (m *Mesh) subscribe(buffer int, allTypes bool, types map[string]struct{}) (<-chan model.Event, func()) {
	if buffer < 1 {
		buffer = 1
	}
	if buffer > maxEventSubscriberBuffer {
		buffer = maxEventSubscriberBuffer
	}
	ch := make(chan model.Event, buffer)
	m.mu.Lock()
	m.subs[ch] = eventSubscriber{ch: ch, allTypes: allTypes, types: types}
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

func (m *Mesh) dispatch(e model.Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, sub := range m.subs {
		if !sub.allTypes {
			if _, ok := sub.types[e.Type]; !ok {
				continue
			}
		}
		select {
		case sub.ch <- e:
		default:
		}
	}
}

func validEventType(eventType string) bool {
	switch eventType {
	case EventServiceUpsertedV1, EventRelationshipUpsertedV1:
		return true
	default:
		return false
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
