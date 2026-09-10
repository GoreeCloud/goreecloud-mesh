package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
)

func TestEventStreamExposesMachineReadableDeliveryBoundary(t *testing.T) {
	m := newEventStreamMesh(t)
	h := WithEventStream(http.NotFoundHandler(), m, eventReadVerifier(ScopeEventsRead), nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/events/stream?type="+mesh.EventServiceUpsertedV1+"&window_seconds=10",
		nil,
	).WithContext(ctx)
	response := newEventFlushRecorder()
	done := make(chan struct{})
	go func() {
		h.ServeHTTP(response, request)
		close(done)
	}()

	waitForEventFlush(t, response.flushes)
	if got := response.Header().Get(eventDeliveryHeader); got != "best-effort-live-only" {
		t.Fatalf("%s = %q", eventDeliveryHeader, got)
	}
	if got := response.Header().Get(eventReplayHeader); got != "unavailable" {
		t.Fatalf("%s = %q", eventReplayHeader, got)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("event stream did not stop after cancellation")
	}
}
