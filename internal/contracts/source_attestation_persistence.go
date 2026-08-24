package contracts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// NewPersistentSourceAttestationRegistry loads a source-attestation registry
// from path. An empty path keeps the registry process-local.
func NewPersistentSourceAttestationRegistry(path string) (*SourceAttestationRegistry, error) {
	r := NewSourceAttestationRegistry()
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

	var persisted []SourceAttestation
	if err := json.Unmarshal(b, &persisted); err != nil {
		return nil, err
	}
	for _, attestation := range persisted {
		entry, ok := CatalogFor(attestation.Platform)
		if !ok {
			return nil, errors.New("persisted source attestation references unknown platform")
		}
		if err := ValidateSourceAttestation(entry, attestation); err != nil {
			return nil, err
		}
		r.attestations[attestation.Platform] = attestation
	}
	return r, nil
}

func (r *SourceAttestationRegistry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil && filepath.Dir(r.path) != "." {
		return err
	}
	items := make([]SourceAttestation, 0, len(r.attestations))
	for _, attestation := range r.attestations {
		items = append(items, attestation)
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
