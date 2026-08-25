package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
)

const testRuntimeRevision = "0123456789abcdef0123456789abcdef01234567"

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	state, err := store.New("")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return New(mesh.New(state), contracts.NewRegistry(), nil)
}

func TestPlatformCatalogAPI(t *testing.T) {
	h := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/platforms", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}

	var body struct {
		Systems []contracts.CatalogEntry `json:"systems"`
		Note    string                   `json:"note"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Systems) != len(contracts.Mandatory()) {
		t.Fatalf("systems = %d mandatory = %d", len(body.Systems), len(contracts.Mandatory()))
	}
	if body.Note == "" {
		t.Fatal("authority-boundary note is required")
	}
}

func TestPlatformStatusAPIFailsClosed(t *testing.T) {
	h := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/platforms/status", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}

	var body struct {
		StableEligible bool                          `json:"stable_eligible"`
		Systems        []contracts.IntegrationStatus `json:"systems"`
		Note           string                        `json:"note"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.StableEligible {
		t.Fatal("empty runtime evidence must fail closed")
	}
	if len(body.Systems) != len(contracts.Mandatory()) {
		t.Fatalf("systems = %d mandatory = %d", len(body.Systems), len(contracts.Mandatory()))
	}
	for _, status := range body.Systems {
		if status.EvidencePresent || status.EvidenceState != contracts.Pending || status.StableGateSatisfied {
			t.Fatalf("unexpected initial status for %s: %+v", status.Platform, status)
		}
	}
	if body.Note == "" {
		t.Fatal("fail-closed boundary note is required")
	}
}

func TestContractStableEligibilityAPI(t *testing.T) {
	h := testHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/contracts/stable-eligibility", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("initial status = %d", response.Code)
	}
	var initial map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	if eligible, _ := initial["stable_eligible"].(bool); eligible {
		t.Fatal("empty contract registry must fail closed")
	}

	for _, platform := range contracts.Mandatory() {
		entry, ok := contracts.CatalogFor(platform)
		if !ok {
			t.Fatalf("catalog entry missing for %s", platform)
		}
		body, err := json.Marshal(contracts.Evidence{
			Platform: platform,
			Contract: entry.ContractSource,
			State:    contracts.Validated,
			Source:   "test-adapter",
			Revision: testRuntimeRevision,
		})
		if err != nil {
			t.Fatal(err)
		}
		request = httptest.NewRequest(http.MethodPost, "/v1/contracts", bytes.NewReader(body))
		response = httptest.NewRecorder()
		h.ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("record %s status = %d body=%s", platform, response.Code, response.Body.String())
		}
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/contracts/stable-eligibility", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	var final map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &final); err != nil {
		t.Fatal(err)
	}
	if eligible, _ := final["stable_eligible"].(bool); !eligible {
		t.Fatal("all validated mandatory contracts should satisfy source-level eligibility evaluation")
	}
}

func TestContractAPIRejectsUnknownPlatform(t *testing.T) {
	h := testHandler(t)
	body := []byte(`{"platform":"other","contract":"v1","state":"validated"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/contracts", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestContractAPIRejectsNonCanonicalValidatedEvidence(t *testing.T) {
	h := testHandler(t)
	body := []byte(`{"platform":"wardveil-security","contract":"wardveil-runtime-v1","state":"validated","source":"test-adapter","revision":"0123456789abcdef0123456789abcdef01234567"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/contracts", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestContractAPIRejectsShortValidatedRevision(t *testing.T) {
	h := testHandler(t)
	entry, ok := contracts.CatalogFor(contracts.Wardveil)
	if !ok {
		t.Fatal("Wardveil catalog entry missing")
	}
	body, err := json.Marshal(contracts.Evidence{
		Platform: contracts.Wardveil,
		Contract: entry.ContractSource,
		State:    contracts.Validated,
		Source:   "test-adapter",
		Revision: "abc123",
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/contracts", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}
