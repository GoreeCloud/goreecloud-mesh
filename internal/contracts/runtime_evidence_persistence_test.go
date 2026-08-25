package contracts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistentRuntimeEvidenceSurvivesReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	registry, err := NewPersistentRegistry(path)
	if err != nil {
		t.Fatalf("new persistent registry: %v", err)
	}

	entry, ok := CatalogFor(Wardveil)
	if !ok {
		t.Fatal("Wardveil catalog entry missing")
	}
	want := Evidence{
		Platform:   Wardveil,
		Repository: entry.Repository,
		Contract:   entry.ContractSource,
		State:      Validated,
		Source:     "wardveil-adapter",
		Revision:   testRevision,
		ObservedAt: time.Date(2026, 8, 24, 23, 0, 0, 0, time.UTC),
		ValidUntil: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
		Detail:     "validated bounded runtime evidence",
	}
	if _, err := registry.Record(want); err != nil {
		t.Fatalf("record evidence: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat persisted evidence: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("runtime evidence file must not be group/world accessible: %o", info.Mode().Perm())
	}

	reloaded, err := NewPersistentRegistry(path)
	if err != nil {
		t.Fatalf("reload persistent registry: %v", err)
	}
	got, ok := reloaded.Get(Wardveil)
	if !ok {
		t.Fatal("expected Wardveil evidence after reload")
	}
	if got != want {
		t.Fatalf("reloaded evidence mismatch: got %#v want %#v", got, want)
	}
}

func TestPersistentRuntimeEvidenceRejectsInvalidStoredState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	entry, ok := CatalogFor(Wardveil)
	if !ok {
		t.Fatal("Wardveil catalog entry missing")
	}
	data := `[{"platform":"wardveil-security","repository":"` + entry.Repository + `","contract":"` + entry.ContractSource + `","state":"accepted","observed_at":"2026-08-24T23:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write invalid persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected invalid persisted runtime evidence to fail closed")
	}
}

func TestPersistentRuntimeEvidenceRejectsDuplicatePlatform(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	entry, ok := CatalogFor(Everkeep)
	if !ok {
		t.Fatal("Everkeep catalog entry missing")
	}
	data := `[
		{"platform":"everkeep","repository":"` + entry.Repository + `","contract":"` + entry.ContractSource + `","state":"pending","observed_at":"2026-08-24T23:00:00Z"},
		{"platform":"everkeep","repository":"` + entry.Repository + `","contract":"` + entry.ContractSource + `","state":"blocked","observed_at":"2026-08-24T23:01:00Z"}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write duplicate persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected duplicate platform runtime evidence to fail closed")
	}
}

func TestPersistentRuntimeEvidenceRejectsNonCanonicalContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	data := `[{"platform":"privacy-shield","repository":"GoreeCloud/goreecloud-privacy-shield","contract":"privacy-runtime-v1","state":"pending","observed_at":"2026-08-24T23:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write invalid persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected non-canonical persisted runtime evidence to fail closed")
	}
}

func TestPersistentRuntimeEvidenceRejectsMismatchedRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	entry, ok := CatalogFor(Wardveil)
	if !ok {
		t.Fatal("Wardveil catalog entry missing")
	}
	data := `[{"platform":"wardveil-security","repository":"GoreeCloud/goreecloud-everkeep","contract":"` + entry.ContractSource + `","state":"pending","observed_at":"2026-08-24T23:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write mismatched persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected mismatched persisted runtime evidence repository to fail closed")
	}
}
