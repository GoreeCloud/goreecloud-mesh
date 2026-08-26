package contracts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const evidenceEnvelopeRegistryVersion = "mesh.evidence-registry.v1"

type persistedEvidenceEnvelopeRegistry struct {
	Version   string             `json:"version"`
	Envelopes []EvidenceEnvelope `json:"envelopes"`
}

// NewPersistentEvidenceEnvelopeRegistry loads immutable evidence envelopes from
// path. Expired envelopes are retained as historical evidence; only malformed,
// future-dated, non-canonical, authority-invalid, or duplicate records block
// startup. An empty path keeps the registry process-local.
func NewPersistentEvidenceEnvelopeRegistry(path string) (*EvidenceEnvelopeRegistry, error) {
	r := NewEvidenceEnvelopeRegistry()
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

	var persisted persistedEvidenceEnvelopeRegistry
	if err := json.Unmarshal(b, &persisted); err != nil {
		return nil, err
	}
	if persisted.Version != evidenceEnvelopeRegistryVersion {
		return nil, errors.New("unsupported persisted evidence envelope registry version")
	}

	now := time.Now().UTC()
	for _, envelope := range persisted.Envelopes {
		validated, err := normalizeStoredEvidenceEnvelope(envelope, now)
		if err != nil {
			return nil, err
		}
		if _, exists := r.envelopes[validated.ID]; exists {
			return nil, errors.New("persisted evidence envelope registry contains duplicate id")
		}
		r.envelopes[validated.ID] = validated
	}
	return r, nil
}

func (r *EvidenceEnvelopeRegistry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil && filepath.Dir(r.path) != "." {
		return err
	}

	payload := persistedEvidenceEnvelopeRegistry{
		Version:   evidenceEnvelopeRegistryVersion,
		Envelopes: make([]EvidenceEnvelope, 0, len(r.envelopes)),
	}
	for _, envelope := range r.envelopes {
		payload.Envelopes = append(payload.Envelopes, envelope)
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}
