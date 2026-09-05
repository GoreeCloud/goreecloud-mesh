package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

const (
	defaultExternalEventBuffer       = 8
	maxExternalEventBuffer           = 64
	defaultEventStreamWindowSeconds  = 5
	maximumEventStreamWindowSeconds  = 10
)

type eventStreamServer struct {
	mesh *mesh.Mesh
}

// WithEventStream adds the authenticated, live-only external event-consumer
// adapter without changing the underlying best-effort in-process event bus.
// The stream is intentionally time-bounded so this source milestone does not
// weaken the server's existing write-timeout protection or imply a durable
// subscription/session contract.
func WithEventStream(base http.Handler, m *mesh.Mesh, verifier trust.Verifier, logger *slog.Logger) http.Handler {
	if base == nil {
		base = http.NotFoundHandler()
	}
	if logger == nil {
		logger = slog.Default()
	}

	s := &eventStreamServer{mesh: m}
	stream := http.HandlerFunc(requireScope(verifier, ScopeEventsRead, s.stream))
	mux := http.NewServeMux()
	mux.Handle("GET /v1/events/stream", requestLog(logger, stream))
	mux.Handle("/", base)
	return mux
}

func (s *eventStreamServer) stream(w http.ResponseWriter, r *http.Request) {
	if s.mesh == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("Mesh event service is unavailable"))
		return
	}
	if strings.TrimSpace(r.Header.Get("Last-Event-ID")) != "" {
		writeError(w, http.StatusConflict, errors.New("event replay is unavailable; Last-Event-ID is not accepted"))
		return
	}

	q := r.URL.Query()
	for key := range q {
		switch key {
		case "type", "buffer", "window_seconds":
		default:
			writeError(w, http.StatusBadRequest, fmt.Errorf("unsupported event stream query parameter %q", key))
			return
		}
	}

	eventTypes := q["type"]
	if len(eventTypes) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("at least one event type filter is required"))
		return
	}

	buffer, err := boundedEventStreamInteger(q["buffer"], defaultExternalEventBuffer, 1, maxExternalEventBuffer, "buffer")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	windowSeconds, err := boundedEventStreamInteger(q["window_seconds"], defaultEventStreamWindowSeconds, 1, maximumEventStreamWindowSeconds, "window_seconds")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	events, cancel, err := s.mesh.SubscribeTypes(buffer, eventTypes...)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer cancel()

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, errors.New("streaming response is unavailable"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, ": goreecloud-mesh live-only best-effort event stream\n\n")
	flusher.Flush()

	timer := time.NewTimer(time.Duration(windowSeconds) * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timer.C:
			_, _ = fmt.Fprint(w, ": stream-window-complete; reconnect without replay guarantees\n\n")
			flusher.Flush()
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			payload, err := json.Marshal(event)
			if err != nil {
				return
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload)
			flusher.Flush()
		}
	}
}

func boundedEventStreamInteger(values []string, fallback, min, max int, name string) (int, error) {
	if len(values) == 0 {
		return fallback, nil
	}
	if len(values) != 1 {
		return 0, fmt.Errorf("%s must be provided at most once", name)
	}
	value := strings.TrimSpace(values[0])
	if value == "" {
		return 0, fmt.Errorf("%s must be an integer from %d to %d", name, min, max)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		return 0, fmt.Errorf("%s must be an integer from %d to %d", name, min, max)
	}
	return parsed, nil
}
