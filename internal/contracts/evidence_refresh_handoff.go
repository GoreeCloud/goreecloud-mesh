package contracts

import (
	"errors"
	"time"
)

// EvidenceRefreshHandoff is the point where a bounded producer handling receipt
// is proven to reference a separately validated, current producer-authoritative
// evidence envelope. The handoff establishes evidence lifecycle state only; the
// producer's outcome remains the domain truth and Mesh gains no producer
// authority or execution capability.
type EvidenceRefreshHandoff struct {
	Response                   EvidenceRefreshResponse `json:"response"`
	Evidence                   EvidenceEnvelope        `json:"evidence"`
	CurrentEvidenceEstablished bool                    `json:"current_evidence_established"`
	ProducerAuthorityPreserved bool                    `json:"producer_authority_preserved"`
	ExecutionAuthorized        bool                    `json:"execution_authorized"`
}

func NormalizeEvidenceRefreshHandoff(response EvidenceRefreshResponse, intent EvidenceRefreshIntent, evidence EvidenceEnvelope) (EvidenceRefreshHandoff, error) {
	return normalizeEvidenceRefreshHandoffAt(response, intent, evidence, time.Now().UTC())
}

func normalizeEvidenceRefreshHandoffAt(response EvidenceRefreshResponse, intent EvidenceRefreshIntent, evidence EvidenceEnvelope, evaluatedAt time.Time) (EvidenceRefreshHandoff, error) {
	normalizedIntent, err := normalizeEvidenceRefreshIntentAt(intent, evaluatedAt)
	if err != nil {
		return EvidenceRefreshHandoff{}, err
	}
	normalizedResponse, err := normalizeEvidenceRefreshResponseForIntentAt(response, normalizedIntent, evaluatedAt)
	if err != nil {
		return EvidenceRefreshHandoff{}, err
	}
	if normalizedResponse.Status != EvidenceRefreshCompleted || !normalizedResponse.EvidenceProduced || normalizedResponse.EvidenceEnvelopeID == "" {
		return EvidenceRefreshHandoff{}, errors.New("evidence refresh handoff requires a completed response that references produced evidence")
	}

	normalizedEvidence, err := normalizeEvidenceEnvelopeAt(evidence, evaluatedAt)
	if err != nil {
		return EvidenceRefreshHandoff{}, err
	}
	if normalizedEvidence.ID != normalizedResponse.EvidenceEnvelopeID {
		return EvidenceRefreshHandoff{}, errors.New("refresh response evidence reference does not match the supplied evidence envelope")
	}
	if normalizedEvidence.Producer.System != normalizedResponse.Producer.System ||
		normalizedEvidence.Producer.Repository != normalizedResponse.Producer.Repository ||
		normalizedEvidence.Producer.Revision != normalizedResponse.Producer.Revision {
		return EvidenceRefreshHandoff{}, errors.New("refresh evidence producer provenance does not match the response producer")
	}
	if normalizedEvidence.AuthorityDomain != normalizedResponse.AuthorityDomain ||
		normalizedEvidence.Subject != normalizedResponse.Subject ||
		normalizedEvidence.Assertion != normalizedResponse.Assertion {
		return EvidenceRefreshHandoff{}, errors.New("refresh evidence changed the response authority, subject, or assertion")
	}
	if normalizedEvidence.ObservedAt.Before(normalizedIntent.RequestedAt) {
		return EvidenceRefreshHandoff{}, errors.New("refresh evidence observation predates the refresh request")
	}
	if normalizedEvidence.ObservedAt.After(normalizedResponse.RespondedAt) {
		return EvidenceRefreshHandoff{}, errors.New("refresh response cannot reference evidence observed after the response")
	}

	return EvidenceRefreshHandoff{
		Response:                   normalizedResponse,
		Evidence:                   normalizedEvidence,
		CurrentEvidenceEstablished: true,
		ProducerAuthorityPreserved: true,
		ExecutionAuthorized:        false,
	}, nil
}
