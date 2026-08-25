package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/governance"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func recoveryAPIHandler(t *testing.T, scopes ...string) http.Handler {
	t.Helper()
	state, err := store.New("")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	verifier := staticVerifier{principal: trust.Principal{
		ServiceID: "everkeep-test-client",
		Issuer:    "goreecloud-identity",
		Subject:   "service:everkeep-test-client",
		Scopes:    scopes,
	}}
	return NewAuthorizedWithRecovery(mesh.New(state), contracts.NewRegistry(), contracts.NewSourceAttestationRegistry(), governance.NewRecoveryRegistry(), verifier, nil)
}

func TestRecoveryAPIRequiresDedicatedWriteScope(t *testing.T) {
	h := recoveryAPIHandler(t, ScopeContractsWrite)
	now := time.Now().UTC()
	body, _ := json.Marshal(governance.RecoveryEvidence{
		Dimension:  governance.RecoveryBackupCoverage,
		State:      "validated",
		Source:     "GoreeCloud/goreecloud-everkeep",
		Revision:   testRuntimeRevision,
		ObservedAt: now.Add(-time.Minute),
		ValidUntil: now.Add(time.Hour),
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/everkeep/recovery-evidence", bytes.NewReader(body))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRecoveryReadinessAPIFailsClosedAndCompletes(t *testing.T) {
	h := recoveryAPIHandler(t, ScopeRecoveryWrite)
	now := time.Now().UTC()

	request := httptest.NewRequest(http.MethodGet, "/v1/everkeep/recovery-readiness", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	var initial map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	if ready, _ := initial["ready"].(bool); ready {
		t.Fatal("empty recovery evidence must fail closed")
	}

	for _, dimension := range governance.RequiredRecoveryDimensions() {
		body, _ := json.Marshal(governance.RecoveryEvidence{
			Dimension:  dimension,
			State:      "validated",
			Source:     "GoreeCloud/goreecloud-everkeep",
			Revision:   testRuntimeRevision,
			ObservedAt: now.Add(-time.Minute),
			ValidUntil: now.Add(time.Hour),
		})
		request := httptest.NewRequest(http.MethodPost, "/v1/everkeep/recovery-evidence", bytes.NewReader(body))
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("record %s status = %d body=%s", dimension, response.Code, response.Body.String())
		}
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/everkeep/recovery-readiness", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	var final map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &final); err != nil {
		t.Fatal(err)
	}
	if ready, _ := final["ready"].(bool); !ready {
		t.Fatal("complete current recovery evidence should satisfy source-level readiness")
	}
}
