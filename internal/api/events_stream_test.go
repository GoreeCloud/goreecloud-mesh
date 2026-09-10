package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

type eventStreamVerifier struct {
	principal trust.Principal
}

func (v eventStreamVerifier) Verify(*http.Request) (trust.Principal, error) {
	return v.principal, nil
}

type eventFlushRecorder struct {
	*httptest.ResponseRecorder
	flushes chan struct{}
	mu      sync.Mutex
}

func newEventFlushRecorder() *eventFlushRecorder {
	return &eventFlushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		flushes:          make(chan struct{}, 8),
	}
}

func (r *eventFlushRecorder) Flush() {
	r.mu.Lock()
	r.ResponseRecorder.Flush()
	r.mu.Unlock()
	select {
	case r.flushes <- struct{}{}:
	default:
	}
}

func newEventStreamMesh(t *testing.T) *mesh.Mesh {
	t.Helper()
	state, err := store.New("")
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	return mesh.New(state)
}

func eventReadVerifier(scopes ...string) eventStreamVerifier {
	return eventStreamVerifier{principal: trust.Principal{
		ServiceID: "event-consumer",
		Issuer:    TrustedIdentityIssuer,
		Subject:   "service:event-consumer",
		Scopes:    scopes,
	}}
}

func TestEventStreamFailsClosedWithoutIdentityVerifier(t *testing.T) {
	h := WithEventStream(http.NotFoundHandler(), newEventStreamMesh(t), nil, nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/events/stream?type="+mesh.EventServiceUpsertedV1, nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEventStreamRequiresDedicatedReadScope(t *testing.T) {
	h := WithEventStream(http.NotFoundHandler(), newEventStreamMesh(t), eventReadVerifier(ScopeEvidenceRead), nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/events/stream?type="+mesh.EventServiceUpsertedV1, nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestEventStreamRejectsAmbiguousOrReplayRequests(t *testing.T) {
	h := WithEventStream(http.NotFoundHandler(), newEventStreamMesh(t), eventReadVerifier(ScopeEventsRead), nil)
	tests := []struct {
		name   string
		target string
		header string
		want   int
	}{
		{name: "missing type", target: "/v1/events/stream", want: http.StatusBadRequest},
		{name: "unknown type", target: "/v1/events/stream?type=mesh.unknown.v1", want: http.StatusBadRequest},
		{name: "oversized buffer", target: "/v1/events/stream?type=" + mesh.EventServiceUpsertedV1 + "&buffer=65", want: http.StatusBadRequest},
		{name: "duplicate buffer", target: "/v1/events/stream?type=" + mesh.EventServiceUpsertedV1 + "&buffer=2&buffer=3", want: http.StatusBadRequest},
		{name: "oversized window", target: "/v1/events/stream?type=" + mesh.EventServiceUpsertedV1 + "&window_seconds=11", want: http.StatusBadRequest},
		{name: "cursor", target: "/v1/events/stream?type=" + mesh.EventServiceUpsertedV1 + "&cursor=evt-1", want: http.StatusBadRequest},
		{name: "last event id", target: "/v1/events/stream?type=" + mesh.EventServiceUpsertedV1, header: "evt-1", want: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.target, nil)
			if tt.header != "" {
				request.Header.Set("Last-Event-ID", tt.header)
			}
			response := httptest.NewRecorder()
			h.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d want=%d body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestEventStreamDeliversOnlyRequestedLifecycleType(t *testing.T) {
	m := newEventStreamMesh(t)
	h := WithEventStream(http.NotFoundHandler(), m, eventReadVerifier(ScopeEventsRead), nil)

	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/events/stream?type="+mesh.EventRelationshipUpsertedV1+"&buffer=2&window_seconds=10",
		nil,
	).WithContext(ctx)
	response := newEventFlushRecorder()
	done := make(chan struct{})
	go func() {
		h.ServeHTTP(response, request)
		close(done)
	}()

	waitForEventFlush(t, response.flushes)

	if _, err := m.RegisterService(model.Service{ID: "source", Name: "Source", Kind: "service"}); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if _, err := m.RegisterService(model.Service{ID: "target", Name: "Target", Kind: "service"}); err != nil {
		t.Fatalf("register target: %v", err)
	}
	if _, err := m.AddRelationship(model.Relationship{ID: "rel-1", From: "source", To: "target", Type: "depends-on", Enabled: true}); err != nil {
		t.Fatalf("add relationship: %v", err)
	}

	waitForEventFlush(t, response.flushes)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("event stream did not stop after request cancellation")
	}

	body := response.Body.String()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, body)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("content type = %q", got)
	}
	if !strings.Contains(body, "event: "+mesh.EventRelationshipUpsertedV1) {
		t.Fatalf("relationship event missing from stream: %s", body)
	}
	if strings.Contains(body, "event: "+mesh.EventServiceUpsertedV1) {
		t.Fatalf("unrequested service event leaked into stream: %s", body)
	}
	if !strings.Contains(body, `"authority_transfer":false`) {
		t.Fatalf("authority-transfer boundary missing from streamed event: %s", body)
	}
	if strings.Contains(body, "\nid: ") {
		t.Fatalf("SSE id field must not imply replay support: %s", body)
	}
}

func waitForEventFlush(t *testing.T, flushes <-chan struct{}) {
	t.Helper()
	select {
	case <-flushes:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event stream flush")
	}
}
