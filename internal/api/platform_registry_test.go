package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/platformregistry"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func platformRecordFixture() platformregistry.Record {
	now := time.Date(2026, 9, 3, 21, 0, 0, 0, time.UTC)
	return platformregistry.Record{
		Schema: platformregistry.Schema,
		Source: platformregistry.Source{
			Repository:            "GoreeCloud/goreecloud-tasks",
			Revision:              "0123456789abcdef0123456789abcdef01234567",
			ContractSchemaVersion: "0.2",
			AuthorityTransfer:     false,
		},
		Component: platformregistry.Component{
			ID:                 "goreecloud-tasks",
			ProductName:        "GoreeCloud Tasks",
			Kind:               "application",
			Repository:         "GoreeCloud/goreecloud-tasks",
			Lifecycle:          "development",
			Version:            "0.1",
			SupportedPlatforms: []string{"web", "linux-server"},
		},
		Dependencies: []string{"goreecloud-mesh"},
		PlatformSystems: map[string]platformregistry.PlatformSystem{
			"manager":           {Result: "applicable-nonconformant"},
			"identity":          {Result: "applicable-migration-required"},
			"wardveil_security": {Result: "applicable-nonconformant"},
			"privacy_shield":    {Result: "applicable-nonconformant"},
			"everkeep":          {Result: "applicable-nonconformant"},
			"mesh":              {Result: "applicable-nonconformant"},
			"glaze_ui":          {Result: "applicable-migration-required"},
		},
		Health:      platformregistry.Health{RuntimeState: "unknown", HealthState: "unknown", Readiness: "unknown"},
		Recovery:    platformregistry.Recovery{BackupStatus: "required_missing", RestoreStatus: "required_missing"},
		Portability: platformregistry.Portability{ExportStatus: "implemented_unverified"},
		Conformance: platformregistry.Conformance{
			DeclaredResult:      "nonconformant",
			ComputedResult:      "nonconformant",
			StableEligible:      false,
			EvaluatorRepository: "GoreeCloud/GoreeCloud",
			EvaluatorRevision:   "abcdef0123456789abcdef0123456789abcdef01",
			EvaluatedAt:         now,
		},
		ObservedAt: now,
	}
}

func platformRegistryHandler(principal trust.Principal) (http.Handler, *platformregistry.Registry) {
	registry := platformregistry.New()
	return WithPlatformRegistry(http.NotFoundHandler(), registry, staticVerifier{principal: principal}), registry
}

func TestPlatformRegistryAPIFailsClosedWithoutVerifier(t *testing.T) {
	h := WithPlatformRegistry(http.NotFoundHandler(), platformregistry.New(), nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/platform-registry", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestPlatformRegistryAPIRecordsAuthenticatedProducer(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-tasks",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-tasks",
		Scopes:    []string{ScopePlatformRegistryRead, ScopePlatformRegistryWrite},
	}
	h, registry := platformRegistryHandler(principal)
	record := platformRecordFixture()
	body, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/platform-registry", bytes.NewReader(body)))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if _, ok := registry.Get(record.Component.ID); !ok {
		t.Fatal("record was not stored")
	}

	response = httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/platform-registry/goreecloud-tasks", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("read status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestPlatformRegistryAPIRejectsCrossProducerWrite(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-manager",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-manager",
		Scopes:    []string{ScopePlatformRegistryWrite},
	}
	h, registry := platformRegistryHandler(principal)
	record := platformRecordFixture()
	body, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/platform-registry", bytes.NewReader(body)))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if _, ok := registry.Get(record.Component.ID); ok {
		t.Fatal("cross-producer write changed registry")
	}
}

func TestPlatformRegistryAPIRequiresReadScope(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-tasks",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-tasks",
		Scopes:    []string{ScopePlatformRegistryWrite},
	}
	h, _ := platformRegistryHandler(principal)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/platform-registry", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}
