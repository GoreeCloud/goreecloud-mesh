package contracts

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"
)

// EvidenceEnvelopeRegistry stores immutable producer-authoritative envelopes.
// Mesh owns registry durability and transport semantics, not the truth of the
// assertion contained in an accepted producer envelope.
type EvidenceEnvelopeRegistry struct {
	mu        sync.RWMutex
	path      string
	envelopes map[string]EvidenceEnvelope
}

func NewEvidenceEnvelopeRegistry() *EvidenceEnvelopeRegistry {
	return &EvidenceEnvelopeRegistry{envelopes: map[string]EvidenceEnvelope{}}
}

func (r *EvidenceEnvelopeRegistry) Record(v EvidenceEnvelope) (EvidenceEnvelope, error) {
	return r.recordAt(v, time.Now().UTC())
}

func (r *EvidenceEnvelopeRegistry) recordAt(v EvidenceEnvelope, evaluatedAt time.Time) (EvidenceEnvelope, error) {
	v, err := normalizeEvidenceEnvelopeAt(v, evaluatedAt)
	if err != nil {
		return EvidenceEnvelope{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.envelopes[v.ID]; ok {
		if reflect.DeepEqual(existing, v) {
			return existing, nil
		}
		return EvidenceEnvelope{}, errors.New("evidence envelope id is immutable and already exists with different content")
	}

	r.envelopes[v.ID] = v
	if err := r.persistLocked(); err != nil {
		delete(r.envelopes, v.ID)
		return EvidenceEnvelope{}, err
	}
	return v, nil
}

func (r *EvidenceEnvelopeRegistry) Get(id string) (EvidenceEnvelope, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.envelopes[id]
	return v, ok
}

func (r *EvidenceEnvelopeRegistry) List() []EvidenceEnvelope {
	r.mu.RLock()
	out := make([]EvidenceEnvelope, 0, len(r.envelopes))
	for _, v := range r.envelopes {
		out = append(out, v)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].ObservedAt.Equal(out[j].ObservedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].ObservedAt.After(out[j].ObservedAt)
	})
	return out
}

func (r *EvidenceEnvelopeRegistry) CurrentAt(evaluatedAt time.Time) []EvidenceEnvelope {
	items := r.List()
	out := make([]EvidenceEnvelope, 0, len(items))
	for _, v := range items {
		if v.FreshAt(evaluatedAt) {
			out = append(out, v)
		}
	}
	return out
}

func (r *EvidenceEnvelopeRegistry) CountsAt(evaluatedAt time.Time) (current int, stale int) {
	for _, v := range r.List() {
		if v.FreshAt(evaluatedAt) {
			current++
		} else {
			stale++
		}
	}
	return current, stale
}
