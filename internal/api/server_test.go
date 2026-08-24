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
		body, err := json.Marshal(contracts.Evidence{Platform: platform, Contract: "v1", State: contracts.Validated, Source: "test"})
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
