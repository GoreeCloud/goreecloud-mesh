package governance

import (
	"testing"
	"time"
)

func TestPrivacyRecordRejectsPayloads(t *testing.T) {
	r := PrivacyRecord{
		Name:            "message-body",
		Class:           DataSensitive,
		Purpose:         "coordination",
		Retention:       RetentionEphemeral,
		ContainsPayload: true,
	}
	if err := r.Validate(); err == nil {
		t.Fatal("expected application payload content to be rejected")
	}
}

func TestBoundedRetentionRequiresMaxAge(t *testing.T) {
	r := PrivacyRecord{
		Name:      "health-evidence",
		Class:     DataOperational,
		Purpose:   "dependency impact",
		Retention: RetentionBounded,
	}
	if err := r.Validate(); err == nil {
		t.Fatal("expected bounded retention without max age to fail")
	}
	r.MaxAge = 24 * time.Hour
	if err := r.Validate(); err != nil {
		t.Fatalf("valid bounded record rejected: %v", err)
	}
}

func recoveryEvidence(dimension RecoveryDimension, now time.Time) RecoveryEvidence {
	return RecoveryEvidence{
		Dimension:  dimension,
		State:      "validated",
		Source:     "everkeep-runtime-acceptance",
		Revision:   "0123456789abcdef0123456789abcdef01234567",
		ObservedAt: now.Add(-time.Minute),
		ValidUntil: now.Add(time.Hour),
	}
}

func TestRecoveryReadyFailsClosed(t *testing.T) {
	now := time.Now().UTC()
	evidence := []RecoveryEvidence{
		recoveryEvidence(RecoveryBackupCoverage, now),
		recoveryEvidence(RecoveryRestoreCapability, now),
		recoveryEvidence(RecoveryPortability, now),
		recoveryEvidence(RecoveryDocumentation, now),
	}
	if RecoveryReady(evidence, now) {
		t.Fatal("missing provenance evidence must fail closed")
	}
	evidence = append(evidence, recoveryEvidence(RecoveryProvenance, now))
	if !RecoveryReady(evidence, now) {
		t.Fatal("all canonical validated recovery dimensions should satisfy source-level recovery readiness")
	}
}

func TestRecoveryReadyRejectsExpiredEvidence(t *testing.T) {
	now := time.Now().UTC()
	evidence := make([]RecoveryEvidence, 0, len(RequiredRecoveryDimensions()))
	for _, dimension := range RequiredRecoveryDimensions() {
		evidence = append(evidence, recoveryEvidence(dimension, now))
	}
	evidence[0].ValidUntil = now.Add(-time.Second)
	if RecoveryReady(evidence, now) {
		t.Fatal("expired recovery evidence must fail closed")
	}
}

func TestRecoveryEvidenceRequiresExactRevisionAndForwardValidity(t *testing.T) {
	now := time.Now().UTC()
	evidence := recoveryEvidence(RecoveryProvenance, now)
	evidence.Revision = "main"
	if err := evidence.Validate(now); err == nil {
		t.Fatal("non-exact revision must be rejected")
	}
	evidence = recoveryEvidence(RecoveryProvenance, now)
	evidence.ValidUntil = evidence.ObservedAt
	if err := evidence.Validate(now); err == nil {
		t.Fatal("non-forward validity must be rejected")
	}
}
