package contracts

import (
	"errors"
	"strings"
	"time"
)

const (
	EvidenceRefreshIntentVersion  = "goreecloud.evidence-refresh-intent.v1"
	EvidenceRefreshIntentContract = "contracts/mesh.evidence-refresh-intent.schema.json"
	EvidenceRefreshCoordinator    = "goreecloud-mesh"
	EvidenceRefreshRepository     = "GoreeCloud/goreecloud-mesh"
)

type EvidenceRefreshReason string

const (
	EvidenceRefreshStale  EvidenceRefreshReason = "stale"
	EvidenceRefreshEmpty  EvidenceRefreshReason = "empty"
	EvidenceRefreshManual EvidenceRefreshReason = "manual"
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

func NormalizeEvidenceRefreshIntent(v EvidenceRefreshIntent) (EvidenceRefreshIntent, error) {
	return normalizeEvidenceRefreshIntentAt(v, time.Now().UTC())
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
	if v.Reason != EvidenceRefreshStale && v.Reason != EvidenceRefreshEmpty && v.Reason != EvidenceRefreshManual {
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
