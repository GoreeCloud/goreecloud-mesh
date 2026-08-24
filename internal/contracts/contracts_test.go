package contracts

import "testing"

func TestStableEligibilityFailsClosed(t *testing.T) {
	r := NewRegistry()
	if r.StableEligible() {
		t.Fatal("empty evidence registry must not be Stable eligible")
	}
	for _, p := range Mandatory() {
		if _, err := r.Record(Evidence{Platform: p, Contract: "v1", State: Validated, Source: "test"}); err != nil {
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
		if _, err := r.Record(Evidence{Platform: p, Contract: "v1", State: state}); err != nil {
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
