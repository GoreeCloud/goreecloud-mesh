package api

import (
	"encoding/json"
	"os"
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

func TestPlatformEvidencePlaneRegistersIntegralProducerAuthorities(t *testing.T) {
	type planeSystem struct {
		System     string   `json:"system"`
		Repository string   `json:"repository"`
		Profile    string   `json:"profile"`
		Domains    []string `json:"authority_domains"`
	}
	type planeContract struct {
		Authority         string        `json:"authority"`
		AuthorityTransfer bool          `json:"authority_transfer"`
		Systems           []planeSystem `json:"systems"`
	}

	raw, err := os.ReadFile("../../contracts/mesh.platform-evidence-plane.v1.json")
	if err != nil {
		t.Fatalf("read platform evidence-plane contract: %v", err)
	}

	var plane planeContract
	if err := json.Unmarshal(raw, &plane); err != nil {
		t.Fatalf("decode platform evidence-plane contract: %v", err)
	}
	if plane.Authority != "coordination-only" {
		t.Fatalf("Mesh evidence-plane authority must remain coordination-only, got %q", plane.Authority)
	}
	if plane.AuthorityTransfer {
		t.Fatal("Mesh evidence plane must never transfer producer authority")
	}

	expected := map[string]string{
		"wardveil-security":   "GoreeCloud/goreecloud-wardveil-security",
		"privacy-shield":      "GoreeCloud/goreecloud-privacy-shield",
		"everkeep":            "GoreeCloud/goreecloud-everkeep",
		"glaze-ui":            "GoreeCloud/goreecloud-glaze-ui",
		"goreecloud-identity": "GoreeCloud/goreecloud-identity",
	}
	if len(plane.Systems) != len(expected) {
		t.Fatalf("expected %d integral producer authorities, got %d", len(expected), len(plane.Systems))
	}

	for _, system := range plane.Systems {
		repository, ok := expected[system.System]
		if !ok {
			t.Fatalf("unexpected evidence-plane producer system %q", system.System)
		}
		if system.Repository != repository {
			t.Fatalf("producer %q must use canonical repository %q, got %q", system.System, repository, system.Repository)
		}
		if system.Profile == "" {
			t.Fatalf("producer %q must declare its producer-owned Mesh evidence profile", system.System)
		}
		if len(system.Domains) == 0 {
			t.Fatalf("producer %q must declare at least one authority domain", system.System)
		}
		delete(expected, system.System)
	}
	if len(expected) != 0 {
		t.Fatalf("missing integral producer authorities: %#v", expected)
	}
}
