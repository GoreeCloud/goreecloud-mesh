package governance

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type RecoveryRegistry struct {
	mu       sync.RWMutex
	evidence map[RecoveryDimension]RecoveryEvidence
	path     string
}

func NewRecoveryRegistry() *RecoveryRegistry {
	return &RecoveryRegistry{evidence: map[RecoveryDimension]RecoveryEvidence{}}
}

func (r *RecoveryRegistry) Record(evidence RecoveryEvidence, at time.Time) (RecoveryEvidence, error) {
	if r == nil {
		return RecoveryEvidence{}, errors.New("recovery registry is unavailable")
	}
	if err := evidence.Validate(at); err != nil {
		return RecoveryEvidence{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	previous, existed := r.evidence[evidence.Dimension]
	r.evidence[evidence.Dimension] = evidence
	if err := r.persistLocked(); err != nil {
		if existed {
			r.evidence[evidence.Dimension] = previous
		} else {
			delete(r.evidence, evidence.Dimension)
		}
		return RecoveryEvidence{}, err
	}
	return evidence, nil
}

func (r *RecoveryRegistry) List() []RecoveryEvidence {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RecoveryEvidence, 0, len(r.evidence))
	for _, evidence := range r.evidence {
		out = append(out, evidence)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dimension < out[j].Dimension })
	return out
}

func (r *RecoveryRegistry) Ready(at time.Time) bool {
	return RecoveryReady(r.List(), at)
}
