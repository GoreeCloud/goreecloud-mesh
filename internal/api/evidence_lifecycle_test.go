package api

import (
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
)

func TestEvidenceTransportLifecycle(t *testing.T) {
	cases := []struct {
		name            string
		current         int
		stale           int
		wantState       string
		wantRefresh     bool
	}{
		{name: "current evidence wins", current: 2, stale: 3, wantState: "current", wantRefresh: false},
		{name: "stale history only", current: 0, stale: 2, wantState: "stale-only", wantRefresh: true},
		{name: "no evidence", current: 0, stale: 0, wantState: "empty", wantRefresh: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, refresh := evidenceTransportLifecycle(tc.current, tc.stale)
			if state != tc.wantState || refresh != tc.wantRefresh {
				t.Fatalf("lifecycle = (%q, %v), want (%q, %v)", state, refresh, tc.wantState, tc.wantRefresh)
			}
		})
	}
}

func TestEvidenceSubjectViewMarksStaleEvidenceAsHistoryOnly(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	envelope := contracts.EvidenceEnvelope{
		Version: contracts.EvidenceEnvelopeVersion,
		ID:      "wardveil-stale-001",
		Producer: contracts.EvidenceEnvelopeProducer{
			System:     contracts.WardveilProducer,
			Repository: "GoreeCloud/goreecloud-wardveil-security",
			Revision:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Contract:   "contracts/wardveil.status.schema.json",
		},
		AuthorityDomain: "security",
		Subject: contracts.EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "goreecloud-drive",
			Scope: "runtime",
		},
		Assertion:              "security-status",
		Outcome:                "protected",
		Source:                 "wardveil://evidence/stale-001",
		ObservedAt:             now.Add(-2 * time.Hour),
		ValidUntil:             now.Add(-time.Hour),
		DataClass:              contracts.EvidenceDerived,
		Summary:                "Historical security status.",
		ContainsUserContent:    false,
		ContainsSecretMaterial: false,
	}

	view := buildEvidenceSubjectView([]contracts.EvidenceEnvelope{envelope}, "service", "goreecloud-drive", "runtime", now)
	if view.Transport.State != "stale-only" || !view.Transport.RefreshRequired {
		t.Fatalf("unexpected transport lifecycle: %#v", view.Transport)
	}
	if view.Transport.CurrentCount != 0 || view.Transport.StaleCount != 1 {
		t.Fatalf("unexpected transport counts: %#v", view.Transport)
	}
	if len(view.Authorities) != 1 || len(view.Authorities[0].Assertions) != 1 {
		t.Fatalf("unexpected authority projection: %#v", view.Authorities)
	}
	assertion := view.Authorities[0].Assertions[0]
	if assertion.Latest.Outcome != "protected" {
		t.Fatalf("historical producer outcome was not preserved: %#v", assertion.Latest)
	}
	if assertion.LatestCurrent != nil {
		t.Fatalf("stale evidence must not be projected as current: %#v", assertion.LatestCurrent)
	}
}

func TestEvidenceSubjectViewMarksMissingEvidenceEmpty(t *testing.T) {
	view := buildEvidenceSubjectView(nil, "service", "goreecloud-drive", "runtime", time.Now().UTC())
	if view.Transport.State != "empty" || !view.Transport.RefreshRequired {
		t.Fatalf("unexpected empty lifecycle: %#v", view.Transport)
	}
	if len(view.Authorities) != 0 {
		t.Fatalf("empty evidence view must not create authorities: %#v", view.Authorities)
	}
}
