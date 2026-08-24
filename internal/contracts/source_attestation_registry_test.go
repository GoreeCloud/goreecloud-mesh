package contracts

import (
	"strings"
	"testing"
	"time"
)

func TestSourceAttestationRegistryFailsClosedUntilAllValidated(t *testing.T) {
	registry := NewSourceAttestationRegistry()
	if registry.AllValidated() {
		t.Fatal("empty source attestation registry must fail closed")
	}

	for _, platform := range Mandatory() {
		entry, ok := CatalogFor(platform)
		if !ok {
			t.Fatalf("catalog entry missing for %s", platform)
		}
		attestation := SourceAttestation{
			Platform:           platform,
			Repository:         entry.Repository,
			Revision:           strings.Repeat("a", 40),
			ManifestPath:       entry.IntegrationManifest,
			ManifestSHA256:     strings.Repeat("b", 64),
			State:              Validated,
			ValidationWorkflow: "test-ci",
			ValidationRunID:    1,
			ObservedAt:         time.Now().UTC(),
		}
		if _, err := registry.Record(attestation); err != nil {
			t.Fatalf("record %s: %v", platform, err)
		}
	}

	if !registry.AllValidated() {
		t.Fatal("all mandatory source attestations should report validated source provenance")
	}
}

func TestSourceAttestationRegistryRejectsCatalogMismatch(t *testing.T) {
	registry := NewSourceAttestationRegistry()
	attestation := SourceAttestation{
		Platform:           GlazeUI,
		Repository:         "GoreeCloud/other",
		Revision:           strings.Repeat("a", 40),
		ManifestPath:       "contracts/mesh.integration.json",
		ManifestSHA256:     strings.Repeat("b", 64),
		State:              Validated,
		ValidationWorkflow: "test-ci",
		ValidationRunID:    1,
		ObservedAt:         time.Now().UTC(),
	}
	if _, err := registry.Record(attestation); err == nil {
		t.Fatal("expected mismatched repository to be rejected")
	}
}
