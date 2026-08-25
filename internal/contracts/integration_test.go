package contracts

import "testing"

func TestIntegrationStatusesFailClosedWithoutEvidence(t *testing.T) {
	r := NewRegistry()
	statuses := r.IntegrationStatuses()
	if len(statuses) != len(Mandatory()) {
		t.Fatalf("statuses = %d mandatory = %d", len(statuses), len(Mandatory()))
	}
	for _, status := range statuses {
		if status.EvidencePresent {
			t.Fatalf("%s unexpectedly has evidence", status.Platform)
		}
		if status.EvidenceState != Pending {
			t.Fatalf("%s state = %q, want pending", status.Platform, status.EvidenceState)
		}
		if status.StableGateSatisfied {
			t.Fatalf("%s unexpectedly satisfies Stable gate", status.Platform)
		}
		if status.Evidence != nil {
			t.Fatalf("%s unexpectedly exposes evidence", status.Platform)
		}
	}
}

func TestIntegrationStatusesReflectRecordedEvidence(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Record(canonicalEvidence(t, Wardveil, Validated)); err != nil {
		t.Fatal(err)
	}
	privacy := canonicalEvidence(t, PrivacyShield, Blocked)
	if _, err := r.Record(privacy); err != nil {
		t.Fatal(err)
	}

	byPlatform := map[Platform]IntegrationStatus{}
	for _, status := range r.IntegrationStatuses() {
		byPlatform[status.Platform] = status
	}

	wardveil := byPlatform[Wardveil]
	if !wardveil.EvidencePresent || wardveil.EvidenceState != Validated || !wardveil.StableGateSatisfied || wardveil.Evidence == nil {
		t.Fatalf("unexpected Wardveil status: %+v", wardveil)
	}
	privacyStatus := byPlatform[PrivacyShield]
	if !privacyStatus.EvidencePresent || privacyStatus.EvidenceState != Blocked || privacyStatus.StableGateSatisfied || privacyStatus.Evidence == nil {
		t.Fatalf("unexpected Privacy Shield status: %+v", privacyStatus)
	}
	glaze := byPlatform[GlazeUI]
	if glaze.EvidencePresent || glaze.EvidenceState != Pending || glaze.StableGateSatisfied {
		t.Fatalf("missing Glaze UI evidence must fail closed: %+v", glaze)
	}
}
