package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/governance"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func evidenceAPIHandler(t *testing.T, scopes ...string) http.Handler {
	t.Helper()
	return evidenceAPIHandlerAs(t, "wardveil-security", contracts.NewEvidenceEnvelopeRegistry(), scopes...)
}

func evidenceAPIHandlerAs(t *testing.T, serviceID string, envelopes *contracts.EvidenceEnvelopeRegistry, scopes ...string) http.Handler {
	t.Helper()
	state, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	verifier := staticVerifier{principal: trust.Principal{
		ServiceID: serviceID,
		Issuer:    "goreecloud-identity",
		Subject:   "service:" + serviceID,
		Scopes:    scopes,
	}}
	return NewAuthorizedWithRecoveryAndEvidence(
		mesh.New(state),
		contracts.NewRegistry(),
		contracts.NewSourceAttestationRegistry(),
		governance.NewRecoveryRegistry(),
		envelopes,
		verifier,
		nil,
	)
}

func apiEnvelope(now time.Time) contracts.EvidenceEnvelope {
	return contracts.EvidenceEnvelope{
		Version: contracts.EvidenceEnvelopeVersion,
		ID:      "wardveil-api-evidence-001",
		Producer: contracts.EvidenceEnvelopeProducer{
			System:     contracts.WardveilProducer,
			Repository: "GoreeCloud/goreecloud-wardveil-security",
			Revision:   strings.Repeat("a", 40),
			Contract:   "contracts/wardveil.status.schema.json",
		},
		AuthorityDomain: "security",
		Subject: contracts.EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "wardveil-security",
			Scope: "runtime",
		},
		Assertion:              "security-status",
		Outcome:                "protected",
		Source:                 "wardveil://evidence/api-001",
		ObservedAt:             now.Add(-time.Minute),
		ValidUntil:             now.Add(time.Hour),
		DataClass:              contracts.EvidenceDerived,
		Summary:                "Derived security status.",
		ContainsUserContent:    false,
		ContainsSecretMaterial: false,
	}
}

func everkeepAPIEnvelope(now time.Time) contracts.EvidenceEnvelope {
	v := apiEnvelope(now)
	v.ID = "everkeep-api-evidence-001"
	v.Producer = contracts.EvidenceEnvelopeProducer{
		System:     contracts.EverkeepProducer,
		Repository: "GoreeCloud/goreecloud-everkeep",
		Revision:   strings.Repeat("b", 40),
		Contract:   "contracts/everkeep.restore-verification.schema.json",
	}
	v.AuthorityDomain = "recovery"
	v.Assertion = "restore-verification"
	v.Outcome = "pass"
	v.Source = "everkeep://restore-verification/api-001"
	v.Summary = "Restore verification passed."
	return v
}

func TestEvidenceEnvelopeAPIRequiresWriteScope(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeEvidenceRead)
	body, _ := json.Marshal(apiEnvelope(time.Now().UTC()))
	request := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEvidenceEnvelopeAPIRequiresReadScope(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeEvidenceWrite)
	request := httptest.NewRequest(http.MethodGet, "/v1/evidence/status", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEvidenceEnvelopeAPIRejectsProducerIdentityMismatch(t *testing.T) {
	h := evidenceAPIHandlerAs(t, "privacy-shield", contracts.NewEvidenceEnvelopeRegistry(), ScopeEvidenceWrite)
	body, _ := json.Marshal(apiEnvelope(time.Now().UTC()))
	request := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEvidenceEnvelopeAPIRecordsReplaysFiltersAndReadsEvidence(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeEvidenceWrite, ScopeEvidenceRead)
	envelope := apiEnvelope(time.Now().UTC())
	body, _ := json.Marshal(envelope)
	request := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("record status = %d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("replay status = %d body=%s", response.Code, response.Body.String())
	}
	var replay struct {
		Replayed bool `json:"replayed"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &replay); err != nil {
		t.Fatal(err)
	}
	if !replay.Replayed {
		t.Fatal("expected exact replay to be reported as replayed")
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/evidence/envelopes?current=true&producer=wardveil-security&authority_domain=security", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", response.Code, response.Body.String())
	}
	var listed struct {
		Count     int `json:"count"`
		Envelopes []struct {
			ID    string `json:"id"`
			Fresh bool   `json:"fresh"`
		} `json:"envelopes"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if listed.Count != 1 || len(listed.Envelopes) != 1 || listed.Envelopes[0].ID != envelope.ID || !listed.Envelopes[0].Fresh {
		t.Fatalf("unexpected evidence list: %#v", listed)
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/evidence/envelopes/"+envelope.ID, nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/evidence/status", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status summary = %d body=%s", response.Code, response.Body.String())
	}
	var status struct {
		Current int `json:"current"`
		Stale   int `json:"stale"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Current != 1 || status.Stale != 0 {
		t.Fatalf("unexpected evidence status: %#v", status)
	}
}

func TestEvidenceSubjectViewPreservesAuthorityBoundaries(t *testing.T) {
	now := time.Now().UTC()
	envelopes := contracts.NewEvidenceEnvelopeRegistry()
	wardveil := apiEnvelope(now)
	wardveil.Subject = contracts.EvidenceEnvelopeSubject{Kind: "service", ID: "goreecloud-drive", Scope: "runtime"}
	everkeep := everkeepAPIEnvelope(now)
	everkeep.Subject = wardveil.Subject
	if _, err := envelopes.Record(wardveil); err != nil {
		t.Fatal(err)
	}
	if _, err := envelopes.Record(everkeep); err != nil {
		t.Fatal(err)
	}

	h := evidenceAPIHandlerAs(t, "mesh-console", envelopes, ScopeEvidenceRead)
	request := httptest.NewRequest(http.MethodGet, "/v1/evidence/subjects/service/goreecloud-drive?scope=runtime", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("subject status = %d body=%s", response.Code, response.Body.String())
	}
	var view struct {
		Transport struct {
			State        string `json:"state"`
			CurrentCount int    `json:"current_count"`
		} `json:"transport"`
		Authorities []struct {
			Producer        string `json:"producer"`
			AuthorityDomain string `json:"authority_domain"`
		} `json:"authorities"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.Transport.State != "available" || view.Transport.CurrentCount != 2 {
		t.Fatalf("unexpected transport view: %#v", view.Transport)
	}
	if len(view.Authorities) != 2 {
		t.Fatalf("authority groups = %d; want 2", len(view.Authorities))
	}
	if view.Authorities[0].Producer == view.Authorities[1].Producer || view.Authorities[0].AuthorityDomain == view.Authorities[1].AuthorityDomain {
		t.Fatalf("authority boundaries collapsed: %#v", view.Authorities)
	}
}

func TestEvidenceEnvelopeAPIRejectsInvalidCurrentFilter(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeEvidenceRead)
	request := httptest.NewRequest(http.MethodGet, "/v1/evidence/envelopes?current=maybe", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}
