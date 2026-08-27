package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func TestRequireScopeFailsClosedWithoutVerifier(t *testing.T) {
	h := requireScope(nil, ScopeServicesWrite, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run without identity verifier")
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireScopeRejectsVerificationFailure(t *testing.T) {
	verifier := staticVerifier{err: errors.New("invalid subject")}
	h := requireScope(verifier, ScopeServicesWrite, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run after verification failure")
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireScopeRejectsInvalidPrincipal(t *testing.T) {
	verifier := staticVerifier{principal: trust.Principal{ServiceID: "client", Scopes: []string{ScopeServicesWrite}}}
	h := requireScope(verifier, ScopeServicesWrite, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run for principal without issuer")
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireScopeRejectsUntrustedIssuer(t *testing.T) {
	verifier := staticVerifier{principal: trust.Principal{ServiceID: "client", Issuer: "example-identity", Subject: "service:client", Scopes: []string{ScopeServicesWrite}}}
	h := requireScope(verifier, ScopeServicesWrite, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run for principal from untrusted issuer")
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireScopeRejectsInsufficientScope(t *testing.T) {
	verifier := staticVerifier{principal: trust.Principal{ServiceID: "client", Issuer: TrustedIdentityIssuer, Subject: "service:client", Scopes: []string{ScopePolicyEvaluate}}}
	h := requireScope(verifier, ScopeServicesWrite, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run without required scope")
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireScopePropagatesVerifiedPrincipal(t *testing.T) {
	verifier := staticVerifier{principal: trust.Principal{ServiceID: " client ", Issuer: " goreecloud-identity ", Subject: " service:client ", Scopes: []string{ScopeServicesWrite}}}
	called := false
	h := requireScope(verifier, ScopeServicesWrite, func(w http.ResponseWriter, r *http.Request) {
		called = true
		principal, ok := trust.PrincipalFromContext(r.Context())
		if !ok || principal.ServiceID != "client" || principal.Issuer != TrustedIdentityIssuer {
			t.Fatalf("unexpected principal: %#v", principal)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	response := httptest.NewRecorder()
	h(response, httptest.NewRequest(http.MethodPost, "/v1/services", nil))
	if !called || response.Code != http.StatusNoContent {
		t.Fatalf("called=%v status=%d body=%s", called, response.Code, response.Body.String())
	}
}
