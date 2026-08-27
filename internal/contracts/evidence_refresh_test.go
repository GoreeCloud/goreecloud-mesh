package contracts

import (
	"strings"
	"testing"
	"time"
)

func refreshIntent(now time.Time) EvidenceRefreshIntent {
	latest := now.Add(-2 * time.Hour)
	return EvidenceRefreshIntent{
		Version: EvidenceRefreshIntentVersion,
		ID:      "refresh-wardveil-drive-001",
		Coordinator: EvidenceRefreshCoordinatorIdentity{
			System:     EvidenceRefreshCoordinator,
			Repository: EvidenceRefreshRepository,
			Revision:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Contract:   EvidenceRefreshIntentContract,
		},
		Producer:        WardveilProducer,
		AuthorityDomain: "security",
		Subject: EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "goreecloud-drive",
			Scope: "runtime",
		},
		Assertion:        "security-status",
		Reason:           EvidenceRefreshStale,
		RequestedAt:      now,
		LatestObservedAt: &latest,
	}
}

func refreshResponse(now time.Time) EvidenceRefreshResponse {
	return EvidenceRefreshResponse{
		Version: EvidenceRefreshResponseVersion,
		ID:      "refresh-response-wardveil-drive-001",
		Intent: EvidenceRefreshIntentReference{
			ID:                  "refresh-wardveil-drive-001",
			CoordinatorRevision: strings.Repeat("a", 40),
			Reason:              EvidenceRefreshStale,
			RequestedAt:         now,
		},
		Producer: EvidenceRefreshResponseProducer{
			System:     WardveilProducer,
			Repository: "GoreeCloud/goreecloud-wardveil-security",
			Revision:   strings.Repeat("b", 40),
		},
		AuthorityDomain: "security",
		Subject: EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "goreecloud-drive",
			Scope: "runtime",
		},
		Assertion:              "security-status",
		Status:                 EvidenceRefreshCompleted,
		ReasonCode:             "evidence-issued",
		RespondedAt:            now.Add(time.Minute),
		EvidenceProduced:       true,
		EvidenceEnvelopeID:     "wardveil-evidence-002",
		ContainsUserContent:    false,
		ContainsSecretMaterial: false,
		AuthorityTransferred:   false,
		ExecutionAuthorized:    false,
	}
}

func TestNormalizeEvidenceRefreshIntentPreservesProducerAuthority(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	intent, err := normalizeEvidenceRefreshIntentAt(refreshIntent(now), now)
	if err != nil {
		t.Fatal(err)
	}
	if intent.Producer != WardveilProducer || intent.AuthorityDomain != "security" {
		t.Fatalf("producer authority changed: %#v", intent)
	}
	if intent.AuthorityTransferred || intent.ExecutionAuthorized {
		t.Fatalf("refresh intent granted authority or execution: %#v", intent)
	}
}

func TestNormalizeEvidenceRefreshIntentRejectsAuthorityMismatch(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	intent := refreshIntent(now)
	intent.AuthorityDomain = "privacy"
	if _, err := normalizeEvidenceRefreshIntentAt(intent, now); err == nil {
		t.Fatal("cross-authority refresh intent must be rejected")
	}
}

func TestNormalizeEvidenceRefreshIntentRejectsExecutionOrAuthorityTransfer(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	intent := refreshIntent(now)
	intent.ExecutionAuthorized = true
	if _, err := normalizeEvidenceRefreshIntentAt(intent, now); err == nil {
		t.Fatal("refresh intent must not authorize execution")
	}
	intent = refreshIntent(now)
	intent.AuthorityTransferred = true
	if _, err := normalizeEvidenceRefreshIntentAt(intent, now); err == nil {
		t.Fatal("refresh intent must not transfer authority")
	}
}

func TestNormalizeEvidenceRefreshIntentEnforcesLifecycleEvidence(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	stale := refreshIntent(now)
	stale.LatestObservedAt = nil
	if _, err := normalizeEvidenceRefreshIntentAt(stale, now); err == nil {
		t.Fatal("stale refresh must identify the latest historical observation")
	}

	empty := refreshIntent(now)
	empty.Reason = EvidenceRefreshEmpty
	if _, err := normalizeEvidenceRefreshIntentAt(empty, now); err == nil {
		t.Fatal("empty refresh must not claim an existing observation")
	}
	empty.LatestObservedAt = nil
	if _, err := normalizeEvidenceRefreshIntentAt(empty, now); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeEvidenceRefreshIntentRejectsSensitiveContentFlags(t *testing.T) {
	now := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	intent := refreshIntent(now)
	intent.ContainsUserContent = true
	if _, err := normalizeEvidenceRefreshIntentAt(intent, now); err == nil {
		t.Fatal("user-content-bearing refresh intent must be rejected")
	}
	intent = refreshIntent(now)
	intent.ContainsSecretMaterial = true
	if _, err := normalizeEvidenceRefreshIntentAt(intent, now); err == nil {
		t.Fatal("secret-bearing refresh intent must be rejected")
	}
}

func TestNormalizeEvidenceRefreshResponseBindsExactIntentWithoutCreatingTruth(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)
	response, err := normalizeEvidenceRefreshResponseForIntentAt(refreshResponse(requestedAt), refreshIntent(requestedAt), evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != EvidenceRefreshCompleted || !response.EvidenceProduced || response.EvidenceEnvelopeID == "" {
		t.Fatalf("refresh response lost handling state: %#v", response)
	}
	if response.AuthorityTransferred || response.ExecutionAuthorized {
		t.Fatalf("refresh response granted authority or execution: %#v", response)
	}
}

func TestNormalizeEvidenceRefreshResponseRejectsIntentOrAuthorityMutation(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)

	wrongIntent := refreshResponse(requestedAt)
	wrongIntent.Intent.ID = "different-intent"
	if _, err := normalizeEvidenceRefreshResponseForIntentAt(wrongIntent, refreshIntent(requestedAt), evaluatedAt); err == nil {
		t.Fatal("response bound to a different intent must be rejected")
	}

	wrongAuthority := refreshResponse(requestedAt)
	wrongAuthority.AuthorityDomain = "privacy"
	if _, err := normalizeEvidenceRefreshResponseForIntentAt(wrongAuthority, refreshIntent(requestedAt), evaluatedAt); err == nil {
		t.Fatal("response that changes producer authority must be rejected")
	}
}

func TestNormalizeEvidenceRefreshResponseSeparatesReceiptFromEvidence(t *testing.T) {
	requestedAt := time.Date(2026, 8, 27, 18, 50, 0, 0, time.UTC)
	evaluatedAt := requestedAt.Add(2 * time.Minute)

	received := refreshResponse(requestedAt)
	received.Status = EvidenceRefreshReceived
	received.EvidenceProduced = false
	received.EvidenceEnvelopeID = ""
	if _, err := normalizeEvidenceRefreshResponseForIntentAt(received, refreshIntent(requestedAt), evaluatedAt); err != nil {
		t.Fatalf("received receipt without evidence should be valid: %v", err)
	}

	invalid := received
	invalid.EvidenceProduced = true
	invalid.EvidenceEnvelopeID = "wardveil-evidence-002"
	if _, err := normalizeEvidenceRefreshResponseForIntentAt(invalid, refreshIntent(requestedAt), evaluatedAt); err == nil {
		t.Fatal("non-completed response must not claim produced evidence")
	}

	orphan := refreshResponse(requestedAt)
	orphan.EvidenceProduced = false
	if _, err := normalizeEvidenceRefreshResponseForIntentAt(orphan, refreshIntent(requestedAt), evaluatedAt); err == nil {
		t.Fatal("response must not carry evidence id when evidence_produced is false")
	}
}

func TestGlazeEvidenceProducerUsesCanonicalRepositoryIdentity(t *testing.T) {
	if evidenceProducerRepositories[GlazeUIProducer] != "GoreeCloud/goreecloud-glaze-ui" {
		t.Fatalf("Glaze UI evidence producer repository is stale: %q", evidenceProducerRepositories[GlazeUIProducer])
	}
}
