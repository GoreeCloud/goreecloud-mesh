package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

const (
	ScopeServicesWrite      = "mesh.services.write"
	ScopeRelationshipsWrite = "mesh.relationships.write"
	ScopePolicyEvaluate     = "mesh.policy.evaluate"
	ScopeAttestationsWrite  = "mesh.attestations.write"
	ScopeContractsWrite     = "mesh.contracts.write"
	ScopeEvidenceRead       = "mesh.evidence.read"
	ScopeEvidenceWrite      = "mesh.evidence.write"
	ScopeRecoveryWrite      = "mesh.everkeep.recovery.write"

	TrustedIdentityIssuer = "goreecloud-identity"
)

func requireScope(verifier trust.Verifier, scope string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if verifier == nil {
			writeError(w, http.StatusUnauthorized, errors.New("identity verification is unavailable"))
			return
		}
		principal, err := verifier.Verify(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, errors.New("identity verification failed"))
			return
		}
		if err := trust.Validate(principal); err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		if strings.TrimSpace(principal.Issuer) != TrustedIdentityIssuer {
			writeError(w, http.StatusUnauthorized, errors.New("untrusted identity issuer"))
			return
		}
		if !trust.HasScope(principal, scope) {
			writeError(w, http.StatusForbidden, errors.New("insufficient scope"))
			return
		}
		next(w, r.WithContext(trust.WithPrincipal(r.Context(), principal)))
	}
}
