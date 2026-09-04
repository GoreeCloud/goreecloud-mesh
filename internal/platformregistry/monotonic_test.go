package platformregistry

import (
	"strings"
	"testing"
	"time"
)

func TestValidateUpdateAllowsExactIdempotentRetry(t *testing.T) {
	current := fixture()
	if err := ValidateUpdate(current, current); err != nil {
		t.Fatalf("exact retry rejected: %v", err)
	}
}

func TestValidateUpdateRejectsProducerObservationRegression(t *testing.T) {
	current := fixture()
	incoming := fixture()
	incoming.ObservedAt = incoming.ObservedAt.Add(-time.Minute)
	if err := ValidateUpdate(current, incoming); err == nil || !strings.Contains(err.Error(), "observed_at") {
		t.Fatalf("expected observed_at regression rejection, got %v", err)
	}
}

func TestValidateUpdateRejectsEvaluatorTimeRegression(t *testing.T) {
	current := fixture()
	incoming := fixture()
	incoming.Conformance.EvaluatedAt = incoming.Conformance.EvaluatedAt.Add(-time.Minute)
	if err := ValidateUpdate(current, incoming); err == nil || !strings.Contains(err.Error(), "evaluated_at") {
		t.Fatalf("expected evaluated_at regression rejection, got %v", err)
	}
}

func TestValidateUpdateRequiresNewObservationForProducerChange(t *testing.T) {
	current := fixture()
	incoming := fixture()
	incoming.Health.RuntimeState = "healthy"
	if err := ValidateUpdate(current, incoming); err == nil || !strings.Contains(err.Error(), "newer observed_at") {
		t.Fatalf("expected same-time producer mutation rejection, got %v", err)
	}

	incoming.ObservedAt = incoming.ObservedAt.Add(time.Minute)
	if err := ValidateUpdate(current, incoming); err != nil {
		t.Fatalf("newer producer observation rejected: %v", err)
	}
}

func TestValidateUpdateRequiresNewEvaluationForConformanceChange(t *testing.T) {
	current := fixture()
	incoming := fixture()
	incoming.Conformance.Blockers = []string{"new canonical evaluation blocker"}
	if err := ValidateUpdate(current, incoming); err == nil || !strings.Contains(err.Error(), "newer evaluated_at") {
		t.Fatalf("expected same-time conformance mutation rejection, got %v", err)
	}

	incoming.Conformance.EvaluatedAt = incoming.Conformance.EvaluatedAt.Add(time.Minute)
	if err := ValidateUpdate(current, incoming); err != nil {
		t.Fatalf("later canonical evaluation rejected: %v", err)
	}
}
