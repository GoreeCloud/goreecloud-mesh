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

	want := Evidence{
		Platform:   Wardveil,
		Contract:   "wardveil-runtime-v1",
		State:      Validated,
		Source:     "wardveil-adapter",
		Revision:   "example-revision",
		ObservedAt: time.Date(2026, 8, 24, 23, 0, 0, 0, time.UTC),
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
	data := `[{"platform":"wardveil-security","contract":"wardveil-runtime-v1","state":"accepted","observed_at":"2026-08-24T23:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write invalid persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected invalid persisted runtime evidence to fail closed")
	}
}

func TestPersistentRuntimeEvidenceRejectsDuplicatePlatform(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-evidence.json")
	data := `[
		{"platform":"everkeep","contract":"everkeep-runtime-v1","state":"pending","observed_at":"2026-08-24T23:00:00Z"},
		{"platform":"everkeep","contract":"everkeep-runtime-v1","state":"blocked","observed_at":"2026-08-24T23:01:00Z"}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write duplicate persisted evidence: %v", err)
	}
	if _, err := NewPersistentRegistry(path); err == nil {
		t.Fatal("expected duplicate platform runtime evidence to fail closed")
	}
}
