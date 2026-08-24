package contracts

import (
	"errors"
	"sort"
	"sync"
)

// SourceAttestationRegistry stores the latest validated source-provenance state
// for each mandatory platform. It is intentionally separate from runtime
// contract evidence so source provenance cannot satisfy runtime or Stable gates.
type SourceAttestationRegistry struct {
	mu           sync.RWMutex
	attestations map[Platform]SourceAttestation
}

func NewSourceAttestationRegistry() *SourceAttestationRegistry {
	return &SourceAttestationRegistry{attestations: map[Platform]SourceAttestation{}}
}

func (r *SourceAttestationRegistry) Record(attestation SourceAttestation) (SourceAttestation, error) {
	entry, ok := CatalogFor(attestation.Platform)
	if !ok {
		return SourceAttestation{}, errors.New("platform is not present in the Mesh platform catalog")
	}
	if err := ValidateSourceAttestation(entry, attestation); err != nil {
		return SourceAttestation{}, err
	}

	r.mu.Lock()
	r.attestations[attestation.Platform] = attestation
	r.mu.Unlock()
	return attestation, nil
}

func (r *SourceAttestationRegistry) Get(platform Platform) (SourceAttestation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	attestation, ok := r.attestations[platform]
	return attestation, ok
}

func (r *SourceAttestationRegistry) List() []SourceAttestation {
	r.mu.RLock()
	out := make([]SourceAttestation, 0, len(r.attestations))
	for _, attestation := range r.attestations {
		out = append(out, attestation)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Platform < out[j].Platform })
	return out
}

// AllValidated reports source-provenance completeness only. It deliberately
// does not imply runtime acceptance or Stable eligibility.
func (r *SourceAttestationRegistry) AllValidated() bool {
	for _, platform := range Mandatory() {
		attestation, ok := r.Get(platform)
		if !ok || attestation.State != Validated {
			return false
		}
	}
	return true
}
