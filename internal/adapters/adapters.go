package adapters

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

// EvidenceSource identifies the specialized GoreeCloud authority that produced an observation.
type EvidenceSource string

const (
	SourceMonitoring EvidenceSource = "goreecloud-monitoring"
	SourceGateway    EvidenceSource = "goreecloud-gateway"
	SourceNetwork    EvidenceSource = "goreecloud-network"
)

// HealthObservation is approved health evidence consumed from GoreeCloud Monitoring.
type HealthObservation struct {
	ServiceID  string            `json:"service_id"`
	State      model.HealthState `json:"state"`
	Source     EvidenceSource    `json:"source"`
	ObservedAt time.Time         `json:"observed_at"`
	Revision   string            `json:"revision,omitempty"`
	Detail     string            `json:"detail,omitempty"`
}

// EnforcementObservation describes whether a declared relationship is represented by a specialized enforcement authority.
type EnforcementObservation struct {
	RelationshipID string         `json:"relationship_id"`
	Source         EvidenceSource `json:"source"`
	Configured     bool           `json:"configured"`
	ObservedAt     time.Time      `json:"observed_at"`
	Revision       string         `json:"revision,omitempty"`
	Detail         string         `json:"detail,omitempty"`
}

type Monitoring interface {
	Health(context.Context, string) (HealthObservation, error)
}

type Gateway interface {
	Enforcement(context.Context, model.Relationship) (EnforcementObservation, error)
}

type Network interface {
	Enforcement(context.Context, model.Relationship) (EnforcementObservation, error)
}

func ValidateHealthObservation(v HealthObservation) error {
	if strings.TrimSpace(v.ServiceID) == "" {
		return errors.New("service_id is required")
	}
	if v.Source != SourceMonitoring {
		return errors.New("health evidence must come from goreecloud-monitoring")
	}
	switch v.State {
	case model.HealthUnknown, model.HealthHealthy, model.HealthDegraded, model.HealthUnavailable:
	default:
		return errors.New("invalid health state")
	}
	if v.ObservedAt.IsZero() {
		return errors.New("observed_at is required")
	}
	return nil
}

func ValidateEnforcementObservation(v EnforcementObservation, expected EvidenceSource) error {
	if strings.TrimSpace(v.RelationshipID) == "" {
		return errors.New("relationship_id is required")
	}
	if expected != SourceGateway && expected != SourceNetwork {
		return errors.New("expected source must be gateway or network")
	}
	if v.Source != expected {
		return errors.New("enforcement evidence source does not match authority")
	}
	if v.ObservedAt.IsZero() {
		return errors.New("observed_at is required")
	}
	return nil
}

// Fresh reports whether evidence remains within the caller-selected maximum age.
// A non-positive maximum age intentionally fails closed.
func Fresh(observedAt, now time.Time, maxAge time.Duration) bool {
	if observedAt.IsZero() || maxAge <= 0 {
		return false
	}
	age := now.UTC().Sub(observedAt.UTC())
	return age >= 0 && age <= maxAge
}
