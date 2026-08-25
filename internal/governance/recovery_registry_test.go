package governance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testRecoveryRevision = "0123456789abcdef0123456789abcdef01234567"

func registryRecoveryEvidence(dimension RecoveryDimension, now time.Time) RecoveryEvidence {
	return RecoveryEvidence{
		Dimension:  dimension,
		State:      "validated",
		Source:     "GoreeCloud/goreecloud-everkeep",
		Revision:   testRecoveryRevision,
		ObservedAt: now.Add(-time.Minute),
		ValidUntil: now.Add(time.Hour),
	}
}

func TestRecoveryRegistryFailsClosedUntilComplete(t *testing.T) {
	now := time.Now().UTC()
	r := NewRecoveryRegistry()
	for _, dimension := range RequiredRecoveryDimensions()[:4] {
		if _, err := r.Record(registryRecoveryEvidence(dimension, now), now); err != nil {
			t.Fatalf("record evidence: %v", err)
		}
	}
	if r.Ready(now) {
		t.Fatal("incomplete recovery evidence must not be ready")
	}
	if _, err := r.Record(registryRecoveryEvidence(RecoveryProvenance, now), now); err != nil {
		t.Fatalf("record provenance: %v", err)
	}
	if !r.Ready(now) {
		t.Fatal("complete valid recovery evidence should be ready")
	}
}

func TestPersistentRecoveryRegistrySurvivesRestartAndExpires(t *testing.T) {
	now := time.Now().UTC()
	path := filepath.Join(t.TempDir(), "recovery.json")
	r, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	for _, dimension := range RequiredRecoveryDimensions() {
		if _, err := r.Record(registryRecoveryEvidence(dimension, now), now); err != nil {
			t.Fatalf("record %s: %v", dimension, err)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat persisted evidence: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("persisted evidence permissions = %o, want 600", info.Mode().Perm())
	}

	reloaded, err := NewPersistentRecoveryRegistry(path, now)
	if err != nil {
		t.Fatalf("reload registry: %v", err)
	}
	if !reloaded.Ready(now) {
		t.Fatal("reloaded complete evidence should be ready before validity expiry")
	}
	if reloaded.Ready(now.Add(2 * time.Hour)) {
		t.Fatal("expired recovery evidence must fail closed")
	}
}
