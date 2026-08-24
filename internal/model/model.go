package model

import "time"

type HealthState string

const (
	HealthUnknown   HealthState = "unknown"
	HealthHealthy   HealthState = "healthy"
	HealthDegraded  HealthState = "degraded"
	HealthUnavailable HealthState = "unavailable"
)

type ConformanceState string

const (
	ConformancePending   ConformanceState = "pending"
	ConformanceValidated ConformanceState = "validated"
	ConformanceBlocked   ConformanceState = "blocked"
)

type Service struct {
	ID           string                      `json:"id"`
	Name         string                      `json:"name"`
	Kind         string                      `json:"kind"`
	Version      string                      `json:"version,omitempty"`
	Endpoint     string                      `json:"endpoint,omitempty"`
	Capabilities []string                    `json:"capabilities,omitempty"`
	Dependencies []string                    `json:"dependencies,omitempty"`
	Health       HealthState                 `json:"health"`
	Labels       map[string]string           `json:"labels,omitempty"`
	Conformance  map[string]ConformanceState `json:"conformance,omitempty"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

type Relationship struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Type       string    `json:"type"`
	Capability string    `json:"capability,omitempty"`
	Contract   string    `json:"contract,omitempty"`
	Required   bool      `json:"required"`
	Enabled    bool      `json:"enabled"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Source    string         `json:"source"`
	Subject   string         `json:"subject,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type State struct {
	Services      map[string]Service      `json:"services"`
	Relationships map[string]Relationship `json:"relationships"`
}

type PolicyRequest struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	Capability string `json:"capability"`
}

type PolicyDecision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}
