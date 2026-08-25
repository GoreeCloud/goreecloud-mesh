package contracts

import "testing"

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
		Revision: "test-revision",
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

func TestValidatedRuntimeEvidenceRequiresSourceAndRevision(t *testing.T) {
	r := NewRegistry()
	evidence := canonicalEvidence(t, Everkeep, Validated)
	evidence.Source = ""
	if _, err := r.Record(evidence); err == nil {
		t.Fatal("expected validated evidence without source to be rejected")
	}

	evidence = canonicalEvidence(t, Everkeep, Validated)
	evidence.Revision = ""
	if _, err := r.Record(evidence); err == nil {
		t.Fatal("expected validated evidence without revision to be rejected")
	}
}
