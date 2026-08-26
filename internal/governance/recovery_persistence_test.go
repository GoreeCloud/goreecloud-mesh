package governance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const recoveryPersistenceRevision = "0123456789abcdef0123456789abcdef01234567"

func TestPersistentRecoveryRegistrySurvivesRestart(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	path := filepath.Join(t.TempDir(), "recovery", "evidence.json")

	registry, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("create persistent recovery registry: %v", err)
	}

	evidence := RecoveryEvidence{
		Dimension:  RecoveryRestoreCapability,
		State:      "validated",
		Source:     "GoreeCloud/goreecloud-everkeep",
		Revision:   recoveryPersistenceRevision,
		ObservedAt: now.Add(-time.Minute),
		ValidUntil: now.Add(time.Hour),
	}
	if _, err := registry.Record(evidence, now); err != nil {
		t.Fatalf("record recovery evidence: %v", err)
	}

	reloaded, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("reload persistent recovery registry: %v", err)
	}
	items := reloaded.List()
	if len(items) != 1 {
		t.Fatalf("reloaded evidence count = %d, want 1", len(items))
	}
	if items[0] != evidence {
		t.Fatalf("reloaded evidence = %#v, want %#v", items[0], evidence)
	}
}

func TestPersistentRecoveryRegistryKeepsExpiredEvidenceForAuditButFailsReadiness(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	path := filepath.Join(t.TempDir(), "recovery.json")

	registry, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("create persistent recovery registry: %v", err)
	}

	for _, dimension := range RequiredRecoveryDimensions() {
		_, err := registry.Record(RecoveryEvidence{
			Dimension:  dimension,
			State:      "validated",
			Source:     "GoreeCloud/goreecloud-everkeep",
			Revision:   recoveryPersistenceRevision,
			ObservedAt: now.Add(-2 * time.Hour),
			ValidUntil: now.Add(-time.Hour),
		}, now)
		if err != nil {
			t.Fatalf("record expired %s evidence: %v", dimension, err)
		}
	}

	reloaded, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("reload expired recovery evidence: %v", err)
	}
	if got := len(reloaded.List()); got != len(RequiredRecoveryDimensions()) {
		t.Fatalf("reloaded evidence count = %d, want %d", got, len(RequiredRecoveryDimensions()))
	}
	if reloaded.Ready(now) {
		t.Fatal("expired persisted recovery evidence must fail closed")
	}
}

func TestPersistentRecoveryRegistryRejectsDuplicatePersistedDimensions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	path := filepath.Join(t.TempDir(), "recovery.json")
	evidence := `[
  {
    "dimension": "backup_coverage",
    "state": "validated",
    "source": "GoreeCloud/goreecloud-everkeep",
    "revision": "0123456789abcdef0123456789abcdef01234567",
    "observed_at": "2026-08-26T12:00:00Z",
    "valid_until": "2026-08-27T12:00:00Z"
  },
  {
    "dimension": "backup_coverage",
    "state": "validated",
    "source": "GoreeCloud/goreecloud-everkeep",
    "revision": "0123456789abcdef0123456789abcdef01234567",
    "observed_at": "2026-08-26T12:00:00Z",
    "valid_until": "2026-08-27T12:00:00Z"
  }
]`
	if err := os.WriteFile(path, []byte(evidence), 0o600); err != nil {
		t.Fatalf("write duplicate persistence fixture: %v", err)
	}

	if _, err := NewPersistentRecoveryRegistry(path, now); err == nil {
		t.Fatal("duplicate persisted recovery dimensions must be rejected")
	}
}
