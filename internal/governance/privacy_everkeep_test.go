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

func TestRecoveryReadyFailsClosed(t *testing.T) {
	now := time.Now().UTC()
	evidence := []RecoveryEvidence{
		{Capability: RecoveryExport, State: "validated", Source: "everkeep", ObservedAt: now},
		{Capability: RecoveryBackup, State: "validated", Source: "everkeep", ObservedAt: now},
		{Capability: RecoveryRestore, State: "validated", Source: "everkeep", ObservedAt: now},
	}
	if RecoveryReady(evidence) {
		t.Fatal("missing verification evidence must fail closed")
	}
	evidence = append(evidence, RecoveryEvidence{Capability: RecoveryVerify, State: "validated", Source: "everkeep", ObservedAt: now})
	if !RecoveryReady(evidence) {
		t.Fatal("all validated recovery capabilities should satisfy source-level recovery readiness")
	}
}
