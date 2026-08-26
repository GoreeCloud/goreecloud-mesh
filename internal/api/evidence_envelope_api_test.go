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
	state, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	verifier := staticVerifier{principal: trust.Principal{
		ServiceID: "mesh-evidence-test-client",
		Issuer:    "goreecloud-identity",
		Subject:   "service:mesh-evidence-test-client",
		Scopes:    scopes,
	}}
	return NewAuthorizedWithRecoveryAndEvidence(
		mesh.New(state),
		contracts.NewRegistry(),
		contracts.NewSourceAttestationRegistry(),
		governance.NewRecoveryRegistry(),
		contracts.NewEvidenceEnvelopeRegistry(),
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

func TestEvidenceEnvelopeAPIRequiresWriteScope(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeContractsWrite)
	body, _ := json.Marshal(apiEnvelope(time.Now().UTC()))
	request := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEvidenceEnvelopeAPIRecordsFiltersAndReadsEvidence(t *testing.T) {
	h := evidenceAPIHandler(t, ScopeEvidenceWrite)
	envelope := apiEnvelope(time.Now().UTC())
	body, _ := json.Marshal(envelope)
	request := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("record status = %d body=%s", response.Code, response.Body.String())
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

func TestEvidenceEnvelopeAPIRejectsInvalidCurrentFilter(t *testing.T) {
	h := evidenceAPIHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/evidence/envelopes?current=maybe", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}
