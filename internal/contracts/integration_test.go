package contracts

import (
	"testing"
	"time"
)

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
		if status.EvidenceFresh {
			t.Fatalf("%s unexpectedly has fresh evidence", status.Platform)
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
	evaluatedAt := time.Date(2026, 8, 25, 1, 0, 0, 0, time.UTC)
	r := NewRegistry()

	wardveil := canonicalEvidence(t, Wardveil, Validated)
	wardveil.ObservedAt = evaluatedAt
	validatedWardveil, err := normalizeEvidenceAt(wardveil, evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	r.evidence[Wardveil] = validatedWardveil

	privacy := canonicalEvidence(t, PrivacyShield, Blocked)
	privacy.ObservedAt = evaluatedAt
	validatedPrivacy, err := normalizeEvidenceAt(privacy, evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	r.evidence[PrivacyShield] = validatedPrivacy

	byPlatform := map[Platform]IntegrationStatus{}
	for _, status := range r.integrationStatusesAt(evaluatedAt) {
		byPlatform[status.Platform] = status
	}

	wardveilStatus := byPlatform[Wardveil]
	if !wardveilStatus.EvidencePresent || wardveilStatus.EvidenceState != Validated || !wardveilStatus.EvidenceFresh || !wardveilStatus.StableGateSatisfied || wardveilStatus.Evidence == nil {
		t.Fatalf("unexpected Wardveil status: %+v", wardveilStatus)
	}
	privacyStatus := byPlatform[PrivacyShield]
	if !privacyStatus.EvidencePresent || privacyStatus.EvidenceState != Blocked || privacyStatus.EvidenceFresh || privacyStatus.StableGateSatisfied || privacyStatus.Evidence == nil {
		t.Fatalf("unexpected Privacy Shield status: %+v", privacyStatus)
	}
	glaze := byPlatform[GlazeUI]
	if glaze.EvidencePresent || glaze.EvidenceState != Pending || glaze.EvidenceFresh || glaze.StableGateSatisfied {
		t.Fatalf("missing Glaze UI evidence must fail closed: %+v", glaze)
	}
}

func TestIntegrationStatusesExposeStaleEvidenceWithoutSatisfyingGate(t *testing.T) {
	evaluatedAt := time.Date(2026, 8, 25, 1, 0, 0, 0, time.UTC)
	r := NewRegistry()
	evidence := canonicalEvidence(t, Everkeep, Validated)
	evidence.ObservedAt = evaluatedAt.Add(-RuntimeEvidenceMaxAge - time.Second)
	validated, err := normalizeEvidenceAt(evidence, evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	r.evidence[Everkeep] = validated

	for _, status := range r.integrationStatusesAt(evaluatedAt) {
		if status.Platform != Everkeep {
			continue
		}
		if !status.EvidencePresent || status.EvidenceState != Validated || status.EvidenceFresh || status.StableGateSatisfied || status.Evidence == nil {
			t.Fatalf("stale Everkeep evidence must remain visible but fail closed: %+v", status)
		}
		return
	}
	t.Fatal("Everkeep status missing")
}
