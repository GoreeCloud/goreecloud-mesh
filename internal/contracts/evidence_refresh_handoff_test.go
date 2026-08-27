package contracts

import (
	"strings"
	"testing"
	"time"
)

func refreshEvidenceEnvelope(requestedAt time.Time) EvidenceEnvelope {
	return EvidenceEnvelope{
		Version: EvidenceEnvelopeVersion,
		ID:      "wardveil-evidence-002",
		Producer: EvidenceEnvelopeProducer{
			System:     WardveilProducer,
			Repository: "GoreeCloud/goreecloud-wardveil-security",
			Revision:   strings.Repeat("b", 40),
			Contract:   "contracts/wardveil.status.schema.json",
		},
		AuthorityDomain: "security",
		Subject: EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "goreecloud-drive",
			Scope: "runtime",
		},
		Assertion:              "security-status",
		Outcome:                "protected",
		Source:                 "wardveil://records/evidence-002",
		ObservedAt:             requestedAt.Add(30 * time.Second),
		ValidUntil:             requestedAt.Add(10 * time.Minute),
		DataClass:              EvidenceDerived,
		Summary:                "Wardveil security-status evidence.",
		ContainsUserContent:    false,
		ContainsSecretMaterial: false,
	}
}

func TestNormalizeEvidenceRefreshHandoffEstablishesCurrentEvidenceOnlyAfterEnvelopeValidation(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)
	handoff, err := normalizeEvidenceRefreshHandoffAt(
		refreshResponse(requestedAt),
		refreshIntent(requestedAt),
		refreshEvidenceEnvelope(requestedAt),
		evaluatedAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !handoff.CurrentEvidenceEstablished || !handoff.ProducerAuthorityPreserved || handoff.ExecutionAuthorized {
		t.Fatalf("invalid handoff boundaries: %#v", handoff)
	}
	if handoff.Evidence.Outcome != "protected" || handoff.Response.EvidenceEnvelopeID != handoff.Evidence.ID {
		t.Fatalf("producer evidence truth/reference was not preserved: %#v", handoff)
	}
}

func TestNormalizeEvidenceRefreshHandoffRejectsUnboundEvidence(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)
	intent := refreshIntent(requestedAt)
	response := refreshResponse(requestedAt)

	cases := map[string]func(*EvidenceEnvelope){
		"wrong envelope id": func(v *EvidenceEnvelope) { v.ID = "wardveil-evidence-other" },
		"wrong producer revision": func(v *EvidenceEnvelope) { v.Producer.Revision = strings.Repeat("c", 40) },
		"wrong authority domain": func(v *EvidenceEnvelope) { v.AuthorityDomain = "privacy" },
		"wrong subject": func(v *EvidenceEnvelope) { v.Subject.ID = "another-service" },
		"wrong assertion": func(v *EvidenceEnvelope) { v.Assertion = "different-status" },
		"pre-request observation": func(v *EvidenceEnvelope) { v.ObservedAt = requestedAt.Add(-time.Second) },
		"post-response observation": func(v *EvidenceEnvelope) { v.ObservedAt = response.RespondedAt.Add(time.Second) },
		"expired evidence": func(v *EvidenceEnvelope) { v.ValidUntil = evaluatedAt.Add(-time.Second) },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			evidence := refreshEvidenceEnvelope(requestedAt)
			mutate(&evidence)
			if _, err := normalizeEvidenceRefreshHandoffAt(response, intent, evidence, evaluatedAt); err == nil {
				t.Fatalf("%s must be rejected", name)
			}
		})
	}
}

func TestNormalizeEvidenceRefreshHandoffRequiresCompletedEvidenceReceipt(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)
	response := refreshResponse(requestedAt)
	response.Status = EvidenceRefreshReceived
	response.EvidenceProduced = false
	response.EvidenceEnvelopeID = ""
	if _, err := normalizeEvidenceRefreshHandoffAt(response, refreshIntent(requestedAt), refreshEvidenceEnvelope(requestedAt), evaluatedAt); err == nil {
		t.Fatal("handoff must require completed handling with an evidence reference")
	}
}
