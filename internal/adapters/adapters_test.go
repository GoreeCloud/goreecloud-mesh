package adapters

import (
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

func TestValidateHealthObservation(t *testing.T) {
	now := time.Now().UTC()
	if err := ValidateHealthObservation(HealthObservation{ServiceID: "goreecloud-identity", State: model.HealthHealthy, Source: SourceMonitoring, ObservedAt: now}); err != nil {
		t.Fatalf("valid observation rejected: %v", err)
	}
	if err := ValidateHealthObservation(HealthObservation{ServiceID: "goreecloud-identity", State: model.HealthHealthy, Source: SourceGateway, ObservedAt: now}); err == nil {
		t.Fatal("gateway must not be accepted as health authority")
	}
}

func TestValidateEnforcementObservation(t *testing.T) {
	now := time.Now().UTC()
	v := EnforcementObservation{RelationshipID: "manager-identity", Source: SourceGateway, Configured: true, ObservedAt: now}
	if err := ValidateEnforcementObservation(v, SourceGateway); err != nil {
		t.Fatalf("valid gateway evidence rejected: %v", err)
	}
	if err := ValidateEnforcementObservation(v, SourceNetwork); err == nil {
		t.Fatal("mismatched enforcement authority must be rejected")
	}
}

func TestFreshFailsClosed(t *testing.T) {
	now := time.Now().UTC()
	if Fresh(now.Add(-time.Minute), now, 0) {
		t.Fatal("non-positive maximum age must fail closed")
	}
	if !Fresh(now.Add(-time.Minute), now, 5*time.Minute) {
		t.Fatal("recent evidence should be fresh")
	}
	if Fresh(now.Add(-10*time.Minute), now, 5*time.Minute) {
		t.Fatal("stale evidence must not be fresh")
	}
	if Fresh(now.Add(time.Minute), now, 5*time.Minute) {
		t.Fatal("future evidence must not be fresh")
	}
}
