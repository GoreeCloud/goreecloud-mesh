package contracts

// IntegrationStatus combines Mesh's read-only authority catalog with the
// latest bounded evidence for one mandatory integral platform system.
type IntegrationStatus struct {
	CatalogEntry
	EvidencePresent    bool      `json:"evidence_present"`
	EvidenceState      State     `json:"evidence_state"`
	StableGateSatisfied bool     `json:"stable_gate_satisfied"`
	Evidence           *Evidence `json:"evidence,omitempty"`
}

// IntegrationStatuses returns one deterministic status record for every
// mandatory platform system. Missing evidence is represented as pending and
// fails closed rather than being omitted or interpreted as healthy.
func (r *Registry) IntegrationStatuses() []IntegrationStatus {
	entries := Catalog()
	out := make([]IntegrationStatus, 0, len(entries))
	for _, entry := range entries {
		status := IntegrationStatus{
			CatalogEntry:        entry,
			EvidenceState:       Pending,
			StableGateSatisfied: false,
		}
		if evidence, ok := r.Get(entry.Platform); ok {
			copy := evidence
			status.EvidencePresent = true
			status.EvidenceState = evidence.State
			status.StableGateSatisfied = evidence.State == Validated
			status.Evidence = &copy
		}
		out = append(out, status)
	}
	return out
}
