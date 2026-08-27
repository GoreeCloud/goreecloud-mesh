package contracts

import (
	"errors"
	"strings"
	"time"
)

const (
	EvidenceRefreshIntentVersion    = "goreecloud.evidence-refresh-intent.v1"
	EvidenceRefreshIntentContract   = "contracts/mesh.evidence-refresh-intent.schema.json"
	EvidenceRefreshResponseVersion  = "goreecloud.evidence-refresh-response.v1"
	EvidenceRefreshResponseContract = "contracts/mesh.evidence-refresh-response.schema.json"
	EvidenceRefreshCoordinator      = "goreecloud-mesh"
	EvidenceRefreshRepository       = "GoreeCloud/goreecloud-mesh"
)

type EvidenceRefreshReason string

const (
	EvidenceRefreshStale  EvidenceRefreshReason = "stale"
	EvidenceRefreshEmpty  EvidenceRefreshReason = "empty"
	EvidenceRefreshManual EvidenceRefreshReason = "manual"
)

type EvidenceRefreshResponseStatus string

const (
	EvidenceRefreshReceived    EvidenceRefreshResponseStatus = "received"
	EvidenceRefreshCompleted   EvidenceRefreshResponseStatus = "completed"
	EvidenceRefreshDeclined    EvidenceRefreshResponseStatus = "declined"
	EvidenceRefreshUnavailable EvidenceRefreshResponseStatus = "unavailable"
)

type EvidenceRefreshCoordinatorIdentity struct {
	System     string `json:"system"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Contract   string `json:"contract"`
}

type EvidenceRefreshIntent struct {
	Version                string                             `json:"version"`
	ID                     string                             `json:"id"`
	Coordinator            EvidenceRefreshCoordinatorIdentity `json:"coordinator"`
	Producer               EvidenceProducerID                 `json:"producer"`
	AuthorityDomain        string                             `json:"authority_domain"`
	Subject                EvidenceEnvelopeSubject            `json:"subject"`
	Assertion              string                             `json:"assertion"`
	Reason                 EvidenceRefreshReason              `json:"reason"`
	RequestedAt            time.Time                          `json:"requested_at"`
	LatestObservedAt       *time.Time                         `json:"latest_observed_at,omitempty"`
	ContainsUserContent    bool                               `json:"contains_user_content"`
	ContainsSecretMaterial bool                               `json:"contains_secret_material"`
	AuthorityTransferred   bool                               `json:"authority_transferred"`
	ExecutionAuthorized    bool                               `json:"execution_authorized"`
}

type EvidenceRefreshIntentReference struct {
	ID                  string                `json:"id"`
	CoordinatorRevision string                `json:"coordinator_revision"`
	Reason              EvidenceRefreshReason `json:"reason"`
	RequestedAt         time.Time             `json:"requested_at"`
}

type EvidenceRefreshResponseProducer struct {
	System     EvidenceProducerID `json:"system"`
	Repository string             `json:"repository"`
	Revision   string             `json:"revision"`
}

type EvidenceRefreshResponse struct {
	Version                string                          `json:"version"`
	ID                     string                          `json:"id"`
	Intent                 EvidenceRefreshIntentReference  `json:"intent"`
	Producer               EvidenceRefreshResponseProducer `json:"producer"`
	AuthorityDomain        string                          `json:"authority_domain"`
	Subject                EvidenceEnvelopeSubject         `json:"subject"`
	Assertion              string                          `json:"assertion"`
	Status                 EvidenceRefreshResponseStatus   `json:"status"`
	ReasonCode             string                          `json:"reason_code,omitempty"`
	RespondedAt            time.Time                       `json:"responded_at"`
	EvidenceProduced       bool                            `json:"evidence_produced"`
	EvidenceEnvelopeID     string                          `json:"evidence_envelope_id,omitempty"`
	ContainsUserContent    bool                            `json:"contains_user_content"`
	ContainsSecretMaterial bool                            `json:"contains_secret_material"`
	AuthorityTransferred   bool                            `json:"authority_transferred"`
	ExecutionAuthorized    bool                            `json:"execution_authorized"`
}

func NormalizeEvidenceRefreshIntent(v EvidenceRefreshIntent) (EvidenceRefreshIntent, error) {
	return normalizeEvidenceRefreshIntentAt(v, time.Now().UTC())
}

func NormalizeEvidenceRefreshResponse(v EvidenceRefreshResponse) (EvidenceRefreshResponse, error) {
	return normalizeEvidenceRefreshResponseAt(v, time.Now().UTC())
}

func NormalizeEvidenceRefreshResponseForIntent(v EvidenceRefreshResponse, intent EvidenceRefreshIntent) (EvidenceRefreshResponse, error) {
	return normalizeEvidenceRefreshResponseForIntentAt(v, intent, time.Now().UTC())
}

func normalizeEvidenceRefreshIntentAt(v EvidenceRefreshIntent, evaluatedAt time.Time) (EvidenceRefreshIntent, error) {
	v.Version = strings.TrimSpace(v.Version)
	v.ID = strings.TrimSpace(v.ID)
	v.Coordinator.System = strings.TrimSpace(v.Coordinator.System)
	v.Coordinator.Repository = strings.TrimSpace(v.Coordinator.Repository)
	v.Coordinator.Revision = strings.TrimSpace(v.Coordinator.Revision)
	v.Coordinator.Contract = strings.TrimSpace(v.Coordinator.Contract)
	v.AuthorityDomain = strings.TrimSpace(v.AuthorityDomain)
	v.Subject.Kind = strings.TrimSpace(v.Subject.Kind)
	v.Subject.ID = strings.TrimSpace(v.Subject.ID)
	v.Subject.Scope = strings.TrimSpace(v.Subject.Scope)
	v.Assertion = strings.TrimSpace(v.Assertion)
	evaluatedAt = evaluatedAt.UTC()

	if v.Version != EvidenceRefreshIntentVersion {
		return EvidenceRefreshIntent{}, errors.New("unsupported evidence refresh intent version")
	}
	if v.ID == "" || len(v.ID) > 128 {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh intent id is required and must be at most 128 characters")
	}
	if v.Coordinator.System != EvidenceRefreshCoordinator || v.Coordinator.Repository != EvidenceRefreshRepository {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh intent must be coordinated by canonical GoreeCloud Mesh")
	}
	if !fullGitRevisionPattern.MatchString(v.Coordinator.Revision) {
		return EvidenceRefreshIntent{}, errors.New("coordinator revision must be an exact 40-character lowercase Git revision")
	}
	if v.Coordinator.Contract != EvidenceRefreshIntentContract {
		return EvidenceRefreshIntent{}, errors.New("coordinator contract must be the canonical Mesh evidence refresh intent contract")
	}
	if v.Producer == MeshProducer {
		return EvidenceRefreshIntent{}, errors.New("Mesh cannot target itself as an evidence producer refresh authority")
	}
	producerDomains, ok := evidenceAuthorityDomains[v.Producer]
	if !ok || !producerDomains[v.AuthorityDomain] {
		return EvidenceRefreshIntent{}, errors.New("authority domain is not valid for the targeted evidence producer")
	}
	if v.Subject.Kind == "" || v.Subject.ID == "" {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh subject kind and id are required")
	}
	if len(v.Subject.Kind) > 64 || len(v.Subject.ID) > 256 || len(v.Subject.Scope) > 256 {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh subject fields exceed maximum length")
	}
	if v.Assertion == "" || len(v.Assertion) > 128 {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh assertion is required and must be at most 128 characters")
	}
	if !validEvidenceRefreshReason(v.Reason) {
		return EvidenceRefreshIntent{}, errors.New("invalid evidence refresh reason")
	}
	if v.RequestedAt.IsZero() {
		return EvidenceRefreshIntent{}, errors.New("requested_at is required")
	}
	v.RequestedAt = v.RequestedAt.UTC()
	if v.RequestedAt.After(evaluatedAt) {
		return EvidenceRefreshIntent{}, errors.New("requested_at cannot be after evaluation time")
	}
	if v.LatestObservedAt != nil {
		latest := v.LatestObservedAt.UTC()
		if latest.After(v.RequestedAt) {
			return EvidenceRefreshIntent{}, errors.New("latest_observed_at cannot be after requested_at")
		}
		v.LatestObservedAt = &latest
	}
	if v.Reason == EvidenceRefreshStale && v.LatestObservedAt == nil {
		return EvidenceRefreshIntent{}, errors.New("stale evidence refresh requires latest_observed_at")
	}
	if v.Reason == EvidenceRefreshEmpty && v.LatestObservedAt != nil {
		return EvidenceRefreshIntent{}, errors.New("empty evidence refresh must not claim an existing observation")
	}
	if v.ContainsUserContent || v.ContainsSecretMaterial {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh intent must not contain user content or secret material")
	}
	if v.AuthorityTransferred || v.ExecutionAuthorized {
		return EvidenceRefreshIntent{}, errors.New("evidence refresh intent cannot transfer producer authority or authorize execution")
	}
	return v, nil
}

func normalizeEvidenceRefreshResponseAt(v EvidenceRefreshResponse, evaluatedAt time.Time) (EvidenceRefreshResponse, error) {
	v.Version = strings.TrimSpace(v.Version)
	v.ID = strings.TrimSpace(v.ID)
	v.Intent.ID = strings.TrimSpace(v.Intent.ID)
	v.Intent.CoordinatorRevision = strings.TrimSpace(v.Intent.CoordinatorRevision)
	v.Producer.Repository = strings.TrimSpace(v.Producer.Repository)
	v.Producer.Revision = strings.TrimSpace(v.Producer.Revision)
	v.AuthorityDomain = strings.TrimSpace(v.AuthorityDomain)
	v.Subject.Kind = strings.TrimSpace(v.Subject.Kind)
	v.Subject.ID = strings.TrimSpace(v.Subject.ID)
	v.Subject.Scope = strings.TrimSpace(v.Subject.Scope)
	v.Assertion = strings.TrimSpace(v.Assertion)
	v.ReasonCode = strings.TrimSpace(v.ReasonCode)
	v.EvidenceEnvelopeID = strings.TrimSpace(v.EvidenceEnvelopeID)
	evaluatedAt = evaluatedAt.UTC()

	if v.Version != EvidenceRefreshResponseVersion {
		return EvidenceRefreshResponse{}, errors.New("unsupported evidence refresh response version")
	}
	if v.ID == "" || len(v.ID) > 128 {
		return EvidenceRefreshResponse{}, errors.New("evidence refresh response id is required and must be at most 128 characters")
	}
	if v.Intent.ID == "" || len(v.Intent.ID) > 128 {
		return EvidenceRefreshResponse{}, errors.New("referenced refresh intent id is required and must be at most 128 characters")
	}
	if !fullGitRevisionPattern.MatchString(v.Intent.CoordinatorRevision) {
		return EvidenceRefreshResponse{}, errors.New("referenced coordinator revision must be an exact 40-character lowercase Git revision")
	}
	if !validEvidenceRefreshReason(v.Intent.Reason) {
		return EvidenceRefreshResponse{}, errors.New("referenced refresh reason is invalid")
	}
	if v.Intent.RequestedAt.IsZero() {
		return EvidenceRefreshResponse{}, errors.New("referenced requested_at is required")
	}
	v.Intent.RequestedAt = v.Intent.RequestedAt.UTC()
	if v.Intent.RequestedAt.After(evaluatedAt) {
		return EvidenceRefreshResponse{}, errors.New("referenced requested_at cannot be after evaluation time")
	}
	if v.Producer.System == MeshProducer {
		return EvidenceRefreshResponse{}, errors.New("Mesh cannot be an evidence refresh response producer")
	}
	canonicalRepository, ok := evidenceProducerRepositories[v.Producer.System]
	if !ok || v.Producer.Repository != canonicalRepository {
		return EvidenceRefreshResponse{}, errors.New("response producer repository does not match canonical GoreeCloud repository")
	}
	if !fullGitRevisionPattern.MatchString(v.Producer.Revision) {
		return EvidenceRefreshResponse{}, errors.New("response producer revision must be an exact 40-character lowercase Git revision")
	}
	producerDomains, ok := evidenceAuthorityDomains[v.Producer.System]
	if !ok || !producerDomains[v.AuthorityDomain] {
		return EvidenceRefreshResponse{}, errors.New("authority domain is not valid for the response producer")
	}
	if v.Subject.Kind == "" || v.Subject.ID == "" {
		return EvidenceRefreshResponse{}, errors.New("refresh response subject kind and id are required")
	}
	if len(v.Subject.Kind) > 64 || len(v.Subject.ID) > 256 || len(v.Subject.Scope) > 256 {
		return EvidenceRefreshResponse{}, errors.New("refresh response subject fields exceed maximum length")
	}
	if v.Assertion == "" || len(v.Assertion) > 128 {
		return EvidenceRefreshResponse{}, errors.New("refresh response assertion is required and must be at most 128 characters")
	}
	if !validEvidenceRefreshResponseStatus(v.Status) {
		return EvidenceRefreshResponse{}, errors.New("invalid evidence refresh response status")
	}
	if len(v.ReasonCode) > 64 {
		return EvidenceRefreshResponse{}, errors.New("refresh response reason_code must be at most 64 characters")
	}
	if v.RespondedAt.IsZero() {
		return EvidenceRefreshResponse{}, errors.New("responded_at is required")
	}
	v.RespondedAt = v.RespondedAt.UTC()
	if v.RespondedAt.Before(v.Intent.RequestedAt) || v.RespondedAt.After(evaluatedAt) {
		return EvidenceRefreshResponse{}, errors.New("responded_at must be between requested_at and evaluation time")
	}
	if v.Status != EvidenceRefreshCompleted && v.EvidenceProduced {
		return EvidenceRefreshResponse{}, errors.New("only a completed refresh response may reference produced evidence")
	}
	if v.EvidenceProduced {
		if v.EvidenceEnvelopeID == "" || len(v.EvidenceEnvelopeID) > 128 {
			return EvidenceRefreshResponse{}, errors.New("produced evidence requires an evidence envelope id of at most 128 characters")
		}
	} else if v.EvidenceEnvelopeID != "" {
		return EvidenceRefreshResponse{}, errors.New("evidence envelope id is forbidden when evidence_produced is false")
	}
	if v.ContainsUserContent || v.ContainsSecretMaterial {
		return EvidenceRefreshResponse{}, errors.New("evidence refresh response must not contain user content or secret material")
	}
	if v.AuthorityTransferred || v.ExecutionAuthorized {
		return EvidenceRefreshResponse{}, errors.New("evidence refresh response cannot transfer producer authority or authorize execution")
	}
	return v, nil
}

func normalizeEvidenceRefreshResponseForIntentAt(v EvidenceRefreshResponse, intent EvidenceRefreshIntent, evaluatedAt time.Time) (EvidenceRefreshResponse, error) {
	normalizedIntent, err := normalizeEvidenceRefreshIntentAt(intent, evaluatedAt)
	if err != nil {
		return EvidenceRefreshResponse{}, err
	}
	normalizedResponse, err := normalizeEvidenceRefreshResponseAt(v, evaluatedAt)
	if err != nil {
		return EvidenceRefreshResponse{}, err
	}
	if normalizedResponse.Intent.ID != normalizedIntent.ID ||
		normalizedResponse.Intent.CoordinatorRevision != normalizedIntent.Coordinator.Revision ||
		normalizedResponse.Intent.Reason != normalizedIntent.Reason ||
		!normalizedResponse.Intent.RequestedAt.Equal(normalizedIntent.RequestedAt) {
		return EvidenceRefreshResponse{}, errors.New("refresh response does not bind the exact original intent")
	}
	if normalizedResponse.Producer.System != normalizedIntent.Producer ||
		normalizedResponse.AuthorityDomain != normalizedIntent.AuthorityDomain ||
		normalizedResponse.Subject != normalizedIntent.Subject ||
		normalizedResponse.Assertion != normalizedIntent.Assertion {
		return EvidenceRefreshResponse{}, errors.New("refresh response changed the original producer authority, subject, or assertion")
	}
	return normalizedResponse, nil
}

func validEvidenceRefreshReason(v EvidenceRefreshReason) bool {
	return v == EvidenceRefreshStale || v == EvidenceRefreshEmpty || v == EvidenceRefreshManual
}

func validEvidenceRefreshResponseStatus(v EvidenceRefreshResponseStatus) bool {
	return v == EvidenceRefreshReceived || v == EvidenceRefreshCompleted || v == EvidenceRefreshDeclined || v == EvidenceRefreshUnavailable
}
