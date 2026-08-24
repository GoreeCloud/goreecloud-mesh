package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func persistentTestAttestation(t *testing.T, platform Platform) SourceAttestation {
	t.Helper()
	entry, ok := CatalogFor(platform)
	if !ok {
		t.Fatalf("catalog entry missing for %s", platform)
	}
	return SourceAttestation{
		Platform:           platform,
		Repository:         entry.Repository,
		Revision:           strings.Repeat("a", 40),
		ManifestPath:       entry.IntegrationManifest,
		ManifestSHA256:     strings.Repeat("b", 64),
		State:              Validated,
		ValidationWorkflow: "test-ci",
		ValidationRunID:    1,
		ObservedAt:         time.Date(2026, 8, 24, 23, 55, 0, 0, time.UTC),
	}
}

func TestPersistentSourceAttestationRegistrySurvivesReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-attestations.json")
	registry, err := NewPersistentSourceAttestationRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	want := persistentTestAttestation(t, Wardveil)
	if _, err := registry.Record(want); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewPersistentSourceAttestationRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.Get(Wardveil)
	if !ok {
		t.Fatal("persisted attestation missing after reload")
	}
	if got.Revision != want.Revision || got.ManifestSHA256 != want.ManifestSHA256 {
		t.Fatalf("reloaded attestation mismatch: %+v", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("source attestation state must not be group/world accessible: %o", info.Mode().Perm())
	}
}

func TestPersistentSourceAttestationRegistryRejectsInvalidStoredEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-attestations.json")
	if err := os.WriteFile(path, []byte(`[{"platform":"glaze-ui","repository":"GoreeCloud/other"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPersistentSourceAttestationRegistry(path); err == nil {
		t.Fatal("invalid persisted source attestation must fail closed")
	}
}
