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
)

func TestSourceAttestationAPIFailsClosedAndRecordsValidatedSource(t *testing.T) {
	h := testHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/platforms/source-attestations", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d body=%s", response.Code, response.Body.String())
	}
	var initial struct {
		AllValidated bool                          `json:"all_validated"`
		Attestations []contracts.SourceAttestation `json:"attestations"`
		Note         string                        `json:"note"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	if initial.AllValidated || len(initial.Attestations) != 0 || initial.Note == "" {
		t.Fatalf("unexpected initial source attestation state: %+v", initial)
	}

	entry, ok := contracts.CatalogFor(contracts.GlazeUI)
	if !ok {
		t.Fatal("Glaze UI catalog entry missing")
	}
	attestation := contracts.SourceAttestation{
		Platform:           contracts.GlazeUI,
		Repository:         entry.Repository,
		Revision:           strings.Repeat("a", 40),
		ManifestPath:       entry.IntegrationManifest,
		ManifestSHA256:     strings.Repeat("b", 64),
		State:              contracts.Validated,
		ValidationWorkflow: "Glaze UI CI",
		ValidationRunID:    1,
		ObservedAt:         time.Now().UTC(),
	}
	body, err := json.Marshal(attestation)
	if err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/platforms/source-attestations", bytes.NewReader(body))
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("record status = %d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/platforms/status", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	var status struct {
		StableEligible              bool `json:"stable_eligible"`
		SourceAttestationsValidated bool `json:"source_attestations_validated"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.StableEligible {
		t.Fatal("source attestation must not satisfy runtime Stable eligibility")
	}
	if status.SourceAttestationsValidated {
		t.Fatal("one of four mandatory source attestations cannot satisfy all-source validation")
	}
}

func TestSourceAttestationAPIRejectsImpliedStableAcceptance(t *testing.T) {
	h := testHandler(t)
	entry, ok := contracts.CatalogFor(contracts.Everkeep)
	if !ok {
		t.Fatal("Everkeep catalog entry missing")
	}
	attestation := contracts.SourceAttestation{
		Platform:                contracts.Everkeep,
		Repository:              entry.Repository,
		Revision:                strings.Repeat("a", 40),
		ManifestPath:            entry.IntegrationManifest,
		ManifestSHA256:          strings.Repeat("b", 64),
		State:                   contracts.Pending,
		ObservedAt:              time.Now().UTC(),
		StableAcceptanceImplied: true,
	}
	body, err := json.Marshal(attestation)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/platforms/source-attestations", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}
