package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

type Server struct {
	mesh      *mesh.Mesh
	contracts *contracts.Registry
	logger    *slog.Logger
}

func New(m *mesh.Mesh, registry *contracts.Registry, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if registry == nil {
		registry = contracts.NewRegistry()
	}
	s := &Server{mesh: m, contracts: registry, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /v1/state", s.state)
	mux.HandleFunc("GET /v1/services", s.services)
	mux.HandleFunc("POST /v1/services", s.services)
	mux.HandleFunc("GET /v1/services/{id}", s.service)
	mux.HandleFunc("GET /v1/capabilities/{capability}", s.capability)
	mux.HandleFunc("POST /v1/relationships", s.relationships)
	mux.HandleFunc("GET /v1/graph/impact", s.impact)
	mux.HandleFunc("POST /v1/evaluate", s.evaluate)
	mux.HandleFunc("GET /v1/platforms", s.platforms)
	mux.HandleFunc("GET /v1/contracts", s.contractsList)
	mux.HandleFunc("POST /v1/contracts", s.contractsRecord)
	mux.HandleFunc("GET /v1/contracts/stable-eligibility", s.contractsStableEligibility)
	return requestLog(logger, mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "goreecloud-mesh"})
}

func (s *Server) state(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.mesh.State())
}

func (s *Server) services(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.mesh.Services())
		return
	}
	var v model.Service
	if err := decodeJSON(r, &v); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.mesh.RegisterService(v)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) service(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	v, ok := s.mesh.Service(id)
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("service not found"))
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) capability(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.mesh.Discover(r.PathValue("capability")))
}

func (s *Server) relationships(w http.ResponseWriter, r *http.Request) {
	var v model.Relationship
	if err := decodeJSON(r, &v); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.mesh.AddRelationship(v)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) impact(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	if _, ok := s.mesh.Service(id); !ok {
		writeError(w, http.StatusNotFound, errors.New("service not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"service": id, "affected": s.mesh.Impact(id)})
}

func (s *Server) evaluate(w http.ResponseWriter, r *http.Request) {
	var req model.PolicyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, s.mesh.Evaluate(req))
}

func (s *Server) platforms(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"systems": contracts.Catalog(),
		"note":    "Mesh coordinates these systems but does not assume their authority.",
	})
}

func (s *Server) contractsList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mandatory": contracts.Mandatory(),
		"evidence":  s.contracts.List(),
	})
}

func (s *Server) contractsRecord(w http.ResponseWriter, r *http.Request) {
	var evidence contracts.Evidence
	if err := decodeJSON(r, &evidence); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	recorded, err := s.contracts.Record(evidence)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, recorded)
}

func (s *Server) contractsStableEligibility(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"stable_eligible": s.contracts.StableEligible(),
		"mandatory":       contracts.Mandatory(),
		"evidence":        s.contracts.List(),
	})
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
