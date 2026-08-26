package api

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

type evidenceDeliveryReceipt struct {
	Envelope          evidenceEnvelopeView `json:"envelope"`
	Replayed          bool                 `json:"replayed"`
	AcceptedAt        time.Time            `json:"accepted_at"`
	ProducerServiceID string               `json:"producer_service_id"`
	Note              string               `json:"note"`
}

type evidenceTransportView struct {
	State        string `json:"state"`
	CurrentCount int    `json:"current_count"`
	StaleCount   int    `json:"stale_count"`
}

type evidenceAssertionView struct {
	Assertion     string                `json:"assertion"`
	Latest        evidenceEnvelopeView  `json:"latest"`
	LatestCurrent *evidenceEnvelopeView `json:"latest_current,omitempty"`
	HistoryCount  int                   `json:"history_count"`
}

type evidenceAuthorityView struct {
	Producer        contracts.EvidenceProducerID `json:"producer"`
	AuthorityDomain string                       `json:"authority_domain"`
	Assertions      []evidenceAssertionView      `json:"assertions"`
}

type evidenceSubjectView struct {
	Subject     contracts.EvidenceEnvelopeSubject `json:"subject"`
	Transport   evidenceTransportView             `json:"transport"`
	Authorities []evidenceAuthorityView           `json:"authorities"`
	Note        string                            `json:"note"`
}

func validateEvidenceProducerPrincipal(r *http.Request, envelope contracts.EvidenceEnvelope) (trust.Principal, error) {
	principal, ok := trust.PrincipalFromContext(r.Context())
	if !ok {
		return trust.Principal{}, errors.New("verified producer identity is unavailable")
	}
	expected := string(envelope.Producer.System)
	if principal.ServiceID != expected {
		return trust.Principal{}, errors.New("authenticated service identity does not match evidence producer")
	}
	if strings.TrimSpace(principal.Subject) == "" {
		return trust.Principal{}, errors.New("authenticated producer subject is required")
	}
	return principal, nil
}

func buildEvidenceSubjectView(envelopes []contracts.EvidenceEnvelope, kind, id, scope string, evaluatedAt time.Time) evidenceSubjectView {
	kind = strings.TrimSpace(kind)
	id = strings.TrimSpace(id)
	scope = strings.TrimSpace(scope)
	evaluatedAt = evaluatedAt.UTC()

	type assertionBucket struct {
		latest        *evidenceEnvelopeView
		latestCurrent *evidenceEnvelopeView
		count         int
	}
	type authorityBucket struct {
		producer contracts.EvidenceProducerID
		domain   string
		items    map[string]*assertionBucket
	}

	buckets := map[string]*authorityBucket{}
	current := 0
	stale := 0
	resolvedScope := scope

	for _, envelope := range envelopes {
		if envelope.Subject.Kind != kind || envelope.Subject.ID != id {
			continue
		}
		if scope != "" && envelope.Subject.Scope != scope {
			continue
		}
		fresh := envelope.FreshAt(evaluatedAt)
		if fresh {
			current++
		} else {
			stale++
		}
		if resolvedScope == "" && envelope.Subject.Scope != "" {
			resolvedScope = envelope.Subject.Scope
		}

		key := string(envelope.Producer.System) + "\x00" + envelope.AuthorityDomain
		bucket := buckets[key]
		if bucket == nil {
			bucket = &authorityBucket{
				producer: envelope.Producer.System,
				domain:   envelope.AuthorityDomain,
				items:    map[string]*assertionBucket{},
			}
			buckets[key] = bucket
		}
		item := bucket.items[envelope.Assertion]
		if item == nil {
			item = &assertionBucket{}
			bucket.items[envelope.Assertion] = item
		}
		item.count++
		view := evidenceEnvelopeView{EvidenceEnvelope: envelope, Fresh: fresh}
		if item.latest == nil || envelope.ObservedAt.After(item.latest.ObservedAt) {
			copy := view
			item.latest = &copy
		}
		if fresh && (item.latestCurrent == nil || envelope.ObservedAt.After(item.latestCurrent.ObservedAt)) {
			copy := view
			item.latestCurrent = &copy
		}
	}

	authorities := make([]evidenceAuthorityView, 0, len(buckets))
	for _, bucket := range buckets {
		assertions := make([]evidenceAssertionView, 0, len(bucket.items))
		for assertion, item := range bucket.items {
			if item.latest == nil {
				continue
			}
			assertions = append(assertions, evidenceAssertionView{
				Assertion:     assertion,
				Latest:        *item.latest,
				LatestCurrent: item.latestCurrent,
				HistoryCount:  item.count,
			})
		}
		sort.Slice(assertions, func(i, j int) bool { return assertions[i].Assertion < assertions[j].Assertion })
		authorities = append(authorities, evidenceAuthorityView{
			Producer:        bucket.producer,
			AuthorityDomain: bucket.domain,
			Assertions:      assertions,
		})
	}
	sort.Slice(authorities, func(i, j int) bool {
		if authorities[i].Producer == authorities[j].Producer {
			return authorities[i].AuthorityDomain < authorities[j].AuthorityDomain
		}
		return authorities[i].Producer < authorities[j].Producer
	})

	return evidenceSubjectView{
		Subject: contracts.EvidenceEnvelopeSubject{Kind: kind, ID: id, Scope: resolvedScope},
		Transport: evidenceTransportView{
			State:        "available",
			CurrentCount: current,
			StaleCount:   stale,
		},
		Authorities: authorities,
		Note:        "Mesh preserves producer outcomes by authority. This view does not combine security, privacy, recovery, continuity, or design-conformance evidence into a single verdict.",
	}
}
