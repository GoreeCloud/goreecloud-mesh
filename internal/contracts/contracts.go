package contracts

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

// Platform identifies a mandatory GoreeCloud integral platform system.
type Platform string

const (
	GlazeUI       Platform = "glaze-ui"
	Wardveil      Platform = "wardveil-security"
	PrivacyShield Platform = "privacy-shield"
	Everkeep      Platform = "everkeep"
)

var mandatory = []Platform{GlazeUI, Wardveil, PrivacyShield, Everkeep}

type State string

const (
	Pending   State = "pending"
	Validated State = "validated"
	Blocked   State = "blocked"
)

// Evidence records runtime-verifiable proof without embedding secrets or private user data.
type Evidence struct {
	Platform   Platform  `json:"platform"`
	Contract   string    `json:"contract"`
	State      State     `json:"state"`
	Source     string    `json:"source,omitempty"`
	Revision   string    `json:"revision,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
	Detail     string    `json:"detail,omitempty"`
}

type Registry struct {
	mu       sync.RWMutex
	path     string
	evidence map[Platform]Evidence
}

func NewRegistry() *Registry { return &Registry{evidence: map[Platform]Evidence{}} }

func normalizeEvidence(v Evidence) (Evidence, error) {
	v.Contract = strings.TrimSpace(v.Contract)
	v.Source = strings.TrimSpace(v.Source)
	v.Revision = strings.TrimSpace(v.Revision)
	v.Detail = strings.TrimSpace(v.Detail)
	entry, ok := CatalogFor(v.Platform)
	if !ok || !entry.Required {
		return Evidence{}, errors.New("platform is not a mandatory Mesh platform contract")
	}
	if v.Contract == "" {
		return Evidence{}, errors.New("contract is required")
	}
	if v.Contract != entry.ContractSource {
		return Evidence{}, errors.New("contract does not match the canonical Mesh platform catalog")
	}
	if v.State != Pending && v.State != Validated && v.State != Blocked {
		return Evidence{}, errors.New("invalid contract state")
	}
	if v.State == Validated {
		if v.Source == "" {
			return Evidence{}, errors.New("validated runtime evidence requires a source")
		}
		if v.Revision == "" {
			return Evidence{}, errors.New("validated runtime evidence requires a revision")
		}
	}
	if v.ObservedAt.IsZero() {
		v.ObservedAt = time.Now().UTC()
	} else {
		v.ObservedAt = v.ObservedAt.UTC()
	}
	return v, nil
}

func (r *Registry) Record(v Evidence) (Evidence, error) {
	v, err := normalizeEvidence(v)
	if err != nil {
		return Evidence{}, err
	}

	r.mu.Lock()
	previous, hadPrevious := r.evidence[v.Platform]
	r.evidence[v.Platform] = v
	if err := r.persistLocked(); err != nil {
		if hadPrevious {
			r.evidence[v.Platform] = previous
		} else {
			delete(r.evidence, v.Platform)
		}
		r.mu.Unlock()
		return Evidence{}, err
	}
	r.mu.Unlock()
	return v, nil
}

func (r *Registry) Get(p Platform) (Evidence, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.evidence[p]
	return v, ok
}

func (r *Registry) List() []Evidence {
	r.mu.RLock()
	out := make([]Evidence, 0, len(r.evidence))
	for _, v := range r.evidence {
		out = append(out, v)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Platform < out[j].Platform })
	return out
}

// StableEligible is deliberately fail-closed: all four mandatory contracts must have validated evidence.
func (r *Registry) StableEligible() bool {
	for _, p := range mandatory {
		v, ok := r.Get(p)
		if !ok || v.State != Validated {
			return false
		}
	}
	return true
}

func Mandatory() []Platform { return append([]Platform(nil), mandatory...) }

func IsMandatory(p Platform) bool {
	for _, candidate := range mandatory {
		if p == candidate {
			return true
		}
	}
	return false
}
