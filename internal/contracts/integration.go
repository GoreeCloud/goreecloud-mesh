package contracts

import "time"

// IntegrationStatus combines Mesh's read-only authority catalog with the
// latest bounded evidence for one mandatory integral platform system.
type IntegrationStatus struct {
	CatalogEntry
	EvidencePresent     bool      `json:"evidence_present"`
	EvidenceState       State     `json:"evidence_state"`
	EvidenceFresh       bool      `json:"evidence_fresh"`
	StableGateSatisfied bool      `json:"stable_gate_satisfied"`
	Evidence            *Evidence `json:"evidence,omitempty"`
}

// IntegrationStatuses returns one deterministic status record for every
// mandatory platform system. Missing or stale evidence fails closed rather
// than being omitted or interpreted as healthy.
func (r *Registry) IntegrationStatuses() []IntegrationStatus {
	return r.integrationStatusesAt(time.Now().UTC())
}

func (r *Registry) integrationStatusesAt(evaluatedAt time.Time) []IntegrationStatus {
	entries := Catalog()
	out := make([]IntegrationStatus, 0, len(entries))
	for _, entry := range entries {
		status := IntegrationStatus{
			CatalogEntry:        entry,
			EvidenceState:       Pending,
			EvidenceFresh:       false,
			StableGateSatisfied: false,
		}
		if evidence, ok := r.Get(entry.Platform); ok {
			copy := evidence
			status.EvidencePresent = true
			status.EvidenceState = evidence.State
			status.EvidenceFresh = evidenceFreshAt(evidence, evaluatedAt)
			status.StableGateSatisfied = status.EvidenceFresh
			status.Evidence = &copy
		}
		out = append(out, status)
	}
	return out
}
