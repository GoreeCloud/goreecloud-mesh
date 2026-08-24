package contracts

import (
	"strings"
	"testing"
	"time"
)

func TestManifestSHA256(t *testing.T) {
	got := ManifestSHA256([]byte("mesh"))
	want := "d30d52968e57011204e4b43aaf78b5e7021d6c3f6f47a827473373135d3cd771"
	if got != want {
		t.Fatalf("unexpected digest %q", got)
	}
}

func TestValidateSourceAttestation(t *testing.T) {
	entry, ok := CatalogFor(GlazeUI)
	if !ok {
		t.Fatal("Glaze UI catalog entry missing")
	}
	attestation := SourceAttestation{
		Platform:           GlazeUI,
		Repository:         entry.Repository,
		Revision:           strings.Repeat("a", 40),
		ManifestPath:       entry.IntegrationManifest,
		ManifestSHA256:     strings.Repeat("b", 64),
		State:              Validated,
		ValidationWorkflow: "Glaze UI CI",
		ValidationRunID:    32783342119,
		ObservedAt:         time.Date(2026, 8, 24, 23, 45, 0, 0, time.UTC),
	}
	if err := ValidateSourceAttestation(entry, attestation); err != nil {
		t.Fatalf("valid attestation rejected: %v", err)
	}
}

func TestValidateSourceAttestationFailsClosed(t *testing.T) {
	entry, ok := CatalogFor(PrivacyShield)
	if !ok {
		t.Fatal("Privacy Shield catalog entry missing")
	}
	base := SourceAttestation{
		Platform:           PrivacyShield,
		Repository:         entry.Repository,
		Revision:           strings.Repeat("c", 40),
		ManifestPath:       entry.IntegrationManifest,
		ManifestSHA256:     strings.Repeat("d", 64),
		State:              Validated,
		ValidationWorkflow: "Validate",
		ValidationRunID:    1,
		ObservedAt:         time.Now().UTC(),
	}

	tests := []struct {
		name   string
		mutate func(*SourceAttestation)
	}{
		{"repository mismatch", func(v *SourceAttestation) { v.Repository = "GoreeCloud/other" }},
		{"manifest mismatch", func(v *SourceAttestation) { v.ManifestPath = "other.json" }},
		{"short revision", func(v *SourceAttestation) { v.Revision = "abc" }},
		{"uppercase revision", func(v *SourceAttestation) { v.Revision = strings.Repeat("A", 40) }},
		{"invalid digest", func(v *SourceAttestation) { v.ManifestSHA256 = strings.Repeat("z", 64) }},
		{"missing observed time", func(v *SourceAttestation) { v.ObservedAt = time.Time{} }},
		{"runtime claim", func(v *SourceAttestation) { v.RuntimeAcceptanceImplied = true }},
		{"stable claim", func(v *SourceAttestation) { v.StableAcceptanceImplied = true }},
		{"validated without workflow", func(v *SourceAttestation) { v.ValidationWorkflow = "" }},
		{"validated without run", func(v *SourceAttestation) { v.ValidationRunID = 0 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := base
			tc.mutate(&candidate)
			if err := ValidateSourceAttestation(entry, candidate); err == nil {
				t.Fatal("expected validation failure")
			}
		})
	}
}

func TestPendingSourceAttestationDoesNotRequireWorkflowEvidence(t *testing.T) {
	entry, ok := CatalogFor(Everkeep)
	if !ok {
		t.Fatal("Everkeep catalog entry missing")
	}
	attestation := SourceAttestation{
		Platform:       Everkeep,
		Repository:     entry.Repository,
		Revision:       strings.Repeat("e", 40),
		ManifestPath:   entry.IntegrationManifest,
		ManifestSHA256: strings.Repeat("f", 64),
		State:          Pending,
		ObservedAt:     time.Now().UTC(),
	}
	if err := ValidateSourceAttestation(entry, attestation); err != nil {
		t.Fatalf("pending attestation should be valid without workflow evidence: %v", err)
	}
}
