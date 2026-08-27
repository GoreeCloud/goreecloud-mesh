package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
)

func TestPlatformEvidencePlaneSeparatesFreshnessFromDomainOutcome(t *testing.T) {
	now := time.Now().UTC()
	registry := contracts.NewEvidenceEnvelopeRegistry()

	current := apiEnvelope(now)
	current.Subject = contracts.EvidenceEnvelopeSubject{Kind: "service", ID: "goreecloud-drive", Scope: "runtime"}
	current.Outcome = "blocked"
	if _, err := registry.Record(current); err != nil {
		t.Fatal(err)
	}

	stale := everkeepAPIEnvelope(now)
	stale.ID = "everkeep-stale-evidence-001"
	stale.Subject = current.Subject
	stale.ObservedAt = now.Add(-2 * time.Hour)
	stale.ValidUntil = now.Add(-time.Hour)
	stale.Outcome = "pass"
	if _, err := registry.Record(stale); err != nil {
		t.Fatal(err)
	}

	h := evidenceAPIHandlerAs(t, "mesh-console", registry, ScopeEvidenceRead)
	req := httptest.NewRequest(http.MethodGet, "/v1/evidence/subjects/service/goreecloud-drive?scope=runtime", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}

	var view struct {
		Transport struct {
			State        string `json:"state"`
			CurrentCount int    `json:"current_count"`
			StaleCount   int    `json:"stale_count"`
		} `json:"transport"`
		Authorities []struct {
			Producer   string `json:"producer"`
			Assertions []struct {
				Latest struct {
					Outcome string `json:"outcome"`
					Fresh   bool   `json:"fresh"`
				} `json:"latest"`
				LatestCurrent *struct {
					Outcome string `json:"outcome"`
				} `json:"latest_current"`
			} `json:"assertions"`
		} `json:"authorities"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.Transport.State != "available" || view.Transport.CurrentCount != 1 || view.Transport.StaleCount != 1 {
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
