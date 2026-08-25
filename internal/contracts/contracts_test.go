package contracts

import (
	"testing"
	"time"
)

const testRevision = "0123456789abcdef0123456789abcdef01234567"

func canonicalEvidence(t *testing.T, platform Platform, state State) Evidence {
	t.Helper()
	entry, ok := CatalogFor(platform)
	if !ok {
		t.Fatalf("catalog entry missing for %s", platform)
	}
	return Evidence{
		Platform: platform,
		Contract: entry.ContractSource,
		State:    state,
		Source:   "test-adapter",
		Revision: testRevision,
	}
}

func TestStableEligibilityFailsClosed(t *testing.T) {
	r := NewRegistry()
	if r.StableEligible() {
		t.Fatal("empty evidence registry must not be Stable eligible")
	}
	for _, p := range Mandatory() {
		if _, err := r.Record(canonicalEvidence(t, p, Validated)); err != nil {
			t.Fatalf("record %s: %v", p, err)
		}
	}
	if !r.StableEligible() {
		t.Fatal("all mandatory validated contracts should be Stable eligible")
	}
}

func TestBlockedContractPreventsEligibility(t *testing.T) {
	r := NewRegistry()
	for _, p := range Mandatory() {
		state := Validated
		if p == PrivacyShield {
			state = Blocked
		}
		if _, err := r.Record(canonicalEvidence(t, p, state)); err != nil {
			t.Fatal(err)
		}
	}
	if r.StableEligible() {
		t.Fatal("blocked mandatory contract must prevent Stable eligibility")
	}
}

func TestUnknownPlatformRejected(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Record(Evidence{Platform: "other", Contract: "v1", State: Validated}); err == nil {
		t.Fatal("expected unknown platform to be rejected")
	}
}

func TestRuntimeEvidenceRejectsNonCanonicalContract(t *testing.T) {
	r := NewRegistry()
	evidence := canonicalEvidence(t, Wardveil, Validated)
	evidence.Contract = "wardveil-runtime-v1"
	if _, err := r.Record(evidence); err == nil {
		t.Fatal("expected non-canonical contract to be rejected")
	}
}

func TestValidatedRuntimeEvidenceRequiresSourceAndExactRevision(t *testing.T) {
	r := NewRegistry()
	evidence := canonicalEvidence(t, Everkeep, Validated)
	evidence.Source = ""
	if _, err := r.Record(evidence); err == nil {
		t.Fatal("expected validated evidence without source to be rejected")
	}

	for _, revision := range []string{"", "abc123", "0123456789ABCDEF0123456789ABCDEF01234567", "g123456789abcdef0123456789abcdef01234567"} {
		evidence = canonicalEvidence(t, Everkeep, Validated)
		evidence.Revision = revision
		if _, err := r.Record(evidence); err == nil {
			t.Fatalf("expected validated evidence revision %q to be rejected", revision)
		}
	}
}

func TestRuntimeEvidenceRejectsFutureObservation(t *testing.T) {
	evaluatedAt := time.Date(2026, 8, 25, 0, 30, 0, 0, time.UTC)
	evidence := canonicalEvidence(t, PrivacyShield, Validated)
	evidence.ObservedAt = evaluatedAt.Add(time.Second)
	if _, err := normalizeEvidenceAt(evidence, evaluatedAt); err == nil {
		t.Fatal("expected future-dated runtime evidence to be rejected")
	}

	evidence.ObservedAt = evaluatedAt
	got, err := normalizeEvidenceAt(evidence, evaluatedAt)
	if err != nil {
		t.Fatalf("current observation rejected: %v", err)
	}
	if !got.ObservedAt.Equal(evaluatedAt) {
		t.Fatalf("observed_at = %s want %s", got.ObservedAt, evaluatedAt)
	}
}
