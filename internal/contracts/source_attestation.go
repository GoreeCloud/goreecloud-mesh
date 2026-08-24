package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// SourceAttestation binds a platform integration manifest to the exact producer
// source revision and digest that was reviewed. It is source-provenance evidence
// only and cannot imply runtime or Stable acceptance.
type SourceAttestation struct {
	Platform                 Platform  `json:"platform"`
	Repository               string    `json:"repository"`
	Revision                 string    `json:"revision"`
	ManifestPath             string    `json:"manifest_path"`
	ManifestSHA256           string    `json:"manifest_sha256"`
	State                    State     `json:"state"`
	ValidationWorkflow       string    `json:"validation_workflow,omitempty"`
	ValidationRunID          int64     `json:"validation_run_id,omitempty"`
	ObservedAt               time.Time `json:"observed_at"`
	RuntimeAcceptanceImplied bool      `json:"runtime_acceptance_implied"`
	StableAcceptanceImplied  bool      `json:"stable_acceptance_implied"`
}

// ManifestSHA256 returns the lowercase SHA-256 digest used by source
// attestations to bind evidence to exact manifest bytes.
func ManifestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ValidateSourceAttestation verifies that source provenance matches the
// canonical platform catalog and remains separate from runtime/Stable claims.
func ValidateSourceAttestation(entry CatalogEntry, attestation SourceAttestation) error {
	if !entry.Required || !IsMandatory(entry.Platform) {
		return errors.New("catalog entry is not a mandatory Mesh platform")
	}
	if attestation.Platform != entry.Platform {
		return errors.New("source attestation platform does not match catalog")
	}
	if strings.TrimSpace(attestation.Repository) != entry.Repository {
		return errors.New("source attestation repository does not match catalog")
	}
	if strings.TrimSpace(attestation.ManifestPath) != entry.IntegrationManifest {
		return errors.New("source attestation manifest path does not match catalog")
	}
	if !isLowerHex(attestation.Revision, 40) {
		return errors.New("source attestation revision must be a full lowercase Git SHA-1")
	}
	if !isLowerHex(attestation.ManifestSHA256, 64) {
		return errors.New("source attestation manifest digest must be a lowercase SHA-256")
	}
	if attestation.State != Pending && attestation.State != Validated && attestation.State != Blocked {
		return errors.New("invalid source attestation state")
	}
	if attestation.ObservedAt.IsZero() {
		return errors.New("source attestation observed_at is required")
	}
	if attestation.RuntimeAcceptanceImplied {
		return errors.New("source attestation cannot imply runtime acceptance")
	}
	if attestation.StableAcceptanceImplied {
		return errors.New("source attestation cannot imply Stable acceptance")
	}
	if attestation.State == Validated {
		if strings.TrimSpace(attestation.ValidationWorkflow) == "" {
			return errors.New("validated source attestation requires a validation workflow")
		}
		if attestation.ValidationRunID <= 0 {
			return errors.New("validated source attestation requires a positive validation run id")
		}
	}
	return nil
}

func isLowerHex(value string, length int) bool {
	if len(value) != length || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
