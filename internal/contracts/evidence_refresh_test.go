package contracts

import (
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
