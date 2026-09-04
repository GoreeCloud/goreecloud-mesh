package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func postPlatformRecord(t *testing.T, h http.Handler, record any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/platform-registry", bytes.NewReader(body)))
	return response
}

func TestPlatformRegistryAPIRejectsRegressiveAuthenticatedWrite(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-tasks",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-tasks",
		Scopes:    []string{ScopePlatformRegistryRead, ScopePlatformRegistryWrite},
	}
	h, registry := platformRegistryHandler(principal)
	current := platformRecordFixture()
	if response := postPlatformRecord(t, h, current); response.Code != http.StatusCreated {
		t.Fatalf("initial status = %d body=%s", response.Code, response.Body.String())
	}

	stale := platformRecordFixture()
	stale.ObservedAt = stale.ObservedAt.Add(-time.Minute)
	stale.Health.RuntimeState = "healthy"
	response := postPlatformRecord(t, h, stale)
	if response.Code != http.StatusConflict {
		t.Fatalf("stale status = %d body=%s", response.Code, response.Body.String())
	}
	stored, _ := registry.Get(current.Component.ID)
	if stored.Health.RuntimeState != current.Health.RuntimeState || !stored.ObservedAt.Equal(current.ObservedAt) {
		t.Fatal("regressive authenticated write changed stored producer state")
	}
}

func TestPlatformRegistryAPIAllowsIndependentMonotonicClocks(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-tasks",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-tasks",
		Scopes:    []string{ScopePlatformRegistryRead, ScopePlatformRegistryWrite},
	}
	h, registry := platformRegistryHandler(principal)
	current := platformRecordFixture()
	if response := postPlatformRecord(t, h, current); response.Code != http.StatusCreated {
		t.Fatalf("initial status = %d body=%s", response.Code, response.Body.String())
	}

	producerAdvance := platformRecordFixture()
	producerAdvance.ObservedAt = producerAdvance.ObservedAt.Add(time.Minute)
	producerAdvance.Health.RuntimeState = "healthy"
	if response := postPlatformRecord(t, h, producerAdvance); response.Code != http.StatusCreated {
		t.Fatalf("producer advance status = %d body=%s", response.Code, response.Body.String())
	}

	evaluatorAdvance := producerAdvance
	evaluatorAdvance.Conformance.EvaluatedAt = evaluatorAdvance.Conformance.EvaluatedAt.Add(2 * time.Minute)
	evaluatorAdvance.Conformance.Blockers = []string{"later canonical evaluation"}
	if response := postPlatformRecord(t, h, evaluatorAdvance); response.Code != http.StatusCreated {
		t.Fatalf("evaluator advance status = %d body=%s", response.Code, response.Body.String())
	}

	stored, _ := registry.Get(current.Component.ID)
	if stored.Health.RuntimeState != "healthy" {
		t.Fatal("newer producer observation was not retained")
	}
	if len(stored.Conformance.Blockers) != 1 || stored.Conformance.Blockers[0] != "later canonical evaluation" {
		t.Fatal("later canonical evaluation was not retained")
	}
}

func TestPlatformRegistryAPIAllowsExactRetryButRejectsSameTimeMutation(t *testing.T) {
	principal := trust.Principal{
		ServiceID: "goreecloud-tasks",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:goreecloud-tasks",
		Scopes:    []string{ScopePlatformRegistryWrite},
	}
	h, _ := platformRegistryHandler(principal)
	current := platformRecordFixture()
	if response := postPlatformRecord(t, h, current); response.Code != http.StatusCreated {
		t.Fatalf("initial status = %d body=%s", response.Code, response.Body.String())
	}
	if response := postPlatformRecord(t, h, current); response.Code != http.StatusCreated {
		t.Fatalf("idempotent retry status = %d body=%s", response.Code, response.Body.String())
	}

	mutated := platformRecordFixture()
	mutated.Health.Readiness = "ready"
	response := postPlatformRecord(t, h, mutated)
	if response.Code != http.StatusConflict {
		t.Fatalf("same-time mutation status = %d body=%s", response.Code, response.Body.String())
	}
}
