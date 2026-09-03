package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func (s *Server) platformRegistryList(w http.ResponseWriter, _ *http.Request) {
	records := s.platformRecords.List()
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(records),
		"records": records,
		"note":    "Mesh indexes source-attributed Platform Contract declarations and computed conformance results for coordination. Repository and Platform-System authorities remain unchanged.",
	})
}

func (s *Server) platformRegistryRecord(w http.ResponseWriter, r *http.Request) {
	var record contracts.PlatformRecord
	if err := decodeJSON(r, &record); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	principal, ok := trust.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, errors.New("verified platform record producer identity is unavailable"))
		return
	}
	if strings.TrimSpace(principal.ServiceID) != strings.TrimSpace(record.ComponentID) {
		writeError(w, http.StatusForbidden, errors.New("authenticated service identity does not match platform record component"))
		return
	}
	recorded, err := s.platformRecords.Record(record)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"record": recorded,
		"note":   "Mesh accepted a normalized coordination record from the authenticated component. Acceptance does not upgrade the repository declaration or computed conformance verdict.",
	})
}

func (s *Server) platformRegistryGet(w http.ResponseWriter, r *http.Request) {
	componentID := strings.TrimSpace(r.PathValue("id"))
	if componentID == "" {
		writeError(w, http.StatusBadRequest, errors.New("platform component id is required"))
		return
	}
	record, ok := s.platformRecords.Get(componentID)
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("platform component record not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"record": record,
		"note":   "This is a source-attributed coordination view. Mesh presentation does not transfer authority from the repository or producer systems.",
	})
}
