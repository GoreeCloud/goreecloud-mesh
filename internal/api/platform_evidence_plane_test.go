package api

import (
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
)

func TestPlatformEvidencePlaneSeparatesFreshnessFromDomainOutcome(t *testing.T) {
	base := time.Date(2026, 8, 26, 20, 0, 0, 0, time.UTC)
	evaluatedAt := base.Add(2 * time.Hour)

	current := apiEnvelope(base)
	current.Subject = contracts.EvidenceEnvelopeSubject{Kind: "service", ID: "goreecloud-drive", Scope: "runtime"}
	current.ObservedAt = base.Add(90 * time.Minute)
	current.ValidUntil = base.Add(3 * time.Hour)
	current.Outcome = "blocked"

	stale := everkeepAPIEnvelope(base)
	stale.ID = "everkeep-stale-evidence-001"
	stale.Subject = current.Subject
	stale.ObservedAt = base
	stale.ValidUntil = base.Add(time.Hour)
	stale.Outcome = "pass"

	view := buildEvidenceSubjectView([]contracts.EvidenceEnvelope{current, stale}, "service", "goreecloud-drive", "runtime", evaluatedAt)
	if view.Transport.State != "current" || view.Transport.CurrentCount != 1 || view.Transport.StaleCount != 1 || view.Transport.RefreshRequired {
		t.Fatalf("unexpected transport view: %#v", view.Transport)
	}

	var sawBlockedCurrent, sawStalePass bool
	for _, authority := range view.Authorities {
		for _, assertion := range authority.Assertions {
			if authority.Producer == "wardveil-security" && assertion.Latest.Outcome == "blocked" && assertion.Latest.Fresh {
				sawBlockedCurrent = true
			}
			if authority.Producer == "everkeep" && assertion.Latest.Outcome == "pass" && !assertion.Latest.Fresh && assertion.LatestCurrent == nil {
				sawStalePass = true
			}
		}
	}
	if !sawBlockedCurrent {
		t.Fatal("current transport evidence must preserve a negative producer outcome")
	}
	if !sawStalePass {
		t.Fatal("expired producer evidence must remain auditable but not current")
	}
}
