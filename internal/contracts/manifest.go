package contracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	IntegrationManifestSchemaVersion = 2
	MeshRepository                   = "GoreeCloud/goreecloud-mesh"
	ReadOnlyCoordination             = "read-only-coordination"
	ProducerOrApplicablePolicy       = "producer-or-applicable-policy"
)

// EvidenceValidity declares who is authoritative for runtime-evidence validity.
// Mesh may evaluate the declared deadline but may not extend or override it.
type EvidenceValidity struct {
	Authority                     string `json:"authority"`
	ValidUntilRequiredForValidated bool   `json:"valid_until_required_for_validated"`
	ConsumerOverrideAllowed        bool   `json:"consumer_override_allowed"`
}

// IntegrationManifest is the bilateral source contract published by each
// mandatory platform repository for GoreeCloud Mesh coordination.
type IntegrationManifest struct {
	SchemaVersion               int              `json:"schema_version"`
	Platform                    Platform         `json:"platform"`
	ProducerRepository          string           `json:"producer_repository"`
	ConsumerRepository          string           `json:"consumer_repository"`
	AuthoritativeContract       string           `json:"authoritative_contract"`
	IntegrationMode             string           `json:"integration_mode"`
	AuthorityTransfer           bool             `json:"authority_transfer"`
	RuntimeAcceptanceImplied    bool             `json:"runtime_acceptance_implied"`
	StableAcceptanceImplied     bool             `json:"stable_acceptance_implied"`
	SensitiveInformationAllowed bool             `json:"sensitive_information_allowed"`
	EvidenceValidity            EvidenceValidity `json:"evidence_validity"`
	Purpose                     string           `json:"purpose"`
}

// DecodeIntegrationManifest rejects unknown fields and trailing JSON so the
// boundary cannot silently broaden without an explicit Mesh source change.
func DecodeIntegrationManifest(data []byte) (IntegrationManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var manifest IntegrationManifest
	if err := decoder.Decode(&manifest); err != nil {
		return IntegrationManifest{}, fmt.Errorf("decode integration manifest: %w", err)
	}
	if decoder.More() {
		return IntegrationManifest{}, errors.New("integration manifest contains trailing JSON")
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return IntegrationManifest{}, errors.New("integration manifest contains trailing JSON")
	}
	return manifest, nil
}

// ValidateIntegrationManifest confirms that a platform's bilateral manifest
// matches the canonical Mesh catalog and preserves fail-closed authority,
// privacy, evidence-validity, runtime-acceptance, and Stable-acceptance boundaries.
func ValidateIntegrationManifest(entry CatalogEntry, manifest IntegrationManifest) error {
	if !entry.Required || !IsMandatory(entry.Platform) {
		return errors.New("catalog entry is not a mandatory Mesh platform")
	}
	if manifest.SchemaVersion != IntegrationManifestSchemaVersion {
		return errors.New("unsupported integration manifest schema version")
	}
	if manifest.Platform != entry.Platform {
		return errors.New("integration manifest platform does not match catalog")
	}
	if strings.TrimSpace(manifest.ProducerRepository) != entry.Repository {
		return errors.New("integration manifest producer repository does not match catalog")
	}
	if strings.TrimSpace(manifest.ConsumerRepository) != MeshRepository {
		return errors.New("integration manifest consumer repository must be GoreeCloud Mesh")
	}
	if strings.TrimSpace(manifest.AuthoritativeContract) != entry.ContractSource {
		return errors.New("integration manifest authoritative contract does not match catalog")
	}
	if strings.TrimSpace(manifest.IntegrationMode) != ReadOnlyCoordination {
		return errors.New("integration manifest must use read-only coordination")
	}
	if manifest.AuthorityTransfer {
		return errors.New("integration manifest cannot transfer platform authority to Mesh")
	}
	if manifest.RuntimeAcceptanceImplied {
		return errors.New("integration manifest cannot imply runtime acceptance")
	}
	if manifest.StableAcceptanceImplied {
		return errors.New("integration manifest cannot imply Stable acceptance")
	}
	if manifest.SensitiveInformationAllowed {
		return errors.New("integration manifest cannot permit sensitive information")
	}
	if strings.TrimSpace(manifest.EvidenceValidity.Authority) != ProducerOrApplicablePolicy {
		return errors.New("integration manifest evidence validity must remain producer or applicable-policy authoritative")
	}
	if !manifest.EvidenceValidity.ValidUntilRequiredForValidated {
		return errors.New("integration manifest must require valid_until for validated evidence")
	}
	if manifest.EvidenceValidity.ConsumerOverrideAllowed {
		return errors.New("integration manifest cannot allow Mesh to override producer-declared evidence validity")
	}
	if strings.TrimSpace(manifest.Purpose) == "" {
		return errors.New("integration manifest purpose is required")
	}
	return nil
}
