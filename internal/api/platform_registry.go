package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/platformregistry"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

type PlatformRegistry interface {
	Upsert(platformregistry.Record) error
	Get(string) (platformregistry.Record, bool)
	List() []platformregistry.Record
	Dependents(string) []string
}

type platformRegistryAPI struct {
	registry PlatformRegistry
}

// WithPlatformRegistry adds authenticated platform-registry transport to the
// existing Mesh API. Authentication proves the submitting service identity and
// Mesh scope only; producer-domain assertions remain owned by the record source.
func WithPlatformRegistry(base http.Handler, registry PlatformRegistry, verifier trust.Verifier) http.Handler {
	if base == nil {
		base = http.NotFoundHandler()
	}
	if registry == nil {
		registry = platformregistry.New()
	}
	api := &platformRegistryAPI{registry: registry}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/platform-registry", requireScope(verifier, ScopePlatformRegistryRead, api.list))
	mux.HandleFunc("POST /v1/platform-registry", requireScope(verifier, ScopePlatformRegistryWrite, api.record))
	mux.HandleFunc("GET /v1/platform-registry/{id}", requireScope(verifier, ScopePlatformRegistryRead, api.get))
	mux.HandleFunc("GET /v1/platform-registry/{id}/dependents", requireScope(verifier, ScopePlatformRegistryRead, api.dependents))
	mux.Handle("/", base)
	return mux
}

func (a *platformRegistryAPI) list(w http.ResponseWriter, _ *http.Request) {
	records := a.registry.List()
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(records),
		"records": records,
		"note":    "Mesh stores authority-preserving coordination metadata. Authentication and registry acceptance do not upgrade producer conformance truth.",
	})
}

func (a *platformRegistryAPI) record(w http.ResponseWriter, r *http.Request) {
	var record platformregistry.Record
	if err := decodeJSON(r, &record); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	principal, ok := trust.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("verified service identity is missing from request context"))
		return
	}
	if strings.TrimSpace(principal.ServiceID) != strings.TrimSpace(record.Component.ID) {
		writeError(w, http.StatusForbidden, errors.New("authenticated service identity must match platform record component id"))
		return
	}
	if err := a.registry.Upsert(record); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	stored, _ := a.registry.Get(record.Component.ID)
	writeJSON(w, http.StatusCreated, map[string]any{
		"record":              stored,
		"accepted_at":         time.Now().UTC(),
		"producer_service_id": principal.ServiceID,
		"authority_transfer":  false,
		"note":                "Mesh accepted transport from the authenticated component service. The producer remains authoritative for the record's underlying facts.",
	})
}

func (a *platformRegistryAPI) get(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("component id is required"))
		return
	}
	record, ok := a.registry.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("platform record not found"))
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *platformRegistryAPI) dependents(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("component id is required"))
		return
	}
	if _, ok := a.registry.Get(id); !ok {
		writeError(w, http.StatusNotFound, errors.New("platform record not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"component_id": id,
		"dependents":   a.registry.Dependents(id),
		"note":         "Dependency impact uses only explicitly declared dependencies and required relationships.",
	})
}
