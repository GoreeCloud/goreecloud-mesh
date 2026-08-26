package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func registryEnvelope(now time.Time) EvidenceEnvelope {
	return EvidenceEnvelope{
		Version: EvidenceEnvelopeVersion,
		ID:      "wardveil-evidence-001",
		Producer: EvidenceEnvelopeProducer{
			System:     WardveilProducer,
			Repository: "GoreeCloud/goreecloud-wardveil-security",
			Revision:   strings.Repeat("a", 40),
			Contract:   "contracts/wardveil.status.schema.json",
		},
		AuthorityDomain: "security",
		Subject: EvidenceEnvelopeSubject{
			Kind:  "service",
			ID:    "wardveil-security",
			Scope: "runtime",
		},
		Assertion:              "security-status",
		Outcome:                "protected",
		Source:                 "wardveil://evidence/001",
		ObservedAt:             now.Add(-time.Minute),
		ValidUntil:             now.Add(time.Hour),
		DataClass:              EvidenceDerived,
		Summary:                "Derived security status.",
		ContainsUserContent:    false,
		ContainsSecretMaterial: false,
	}
}

func TestEvidenceEnvelopeRegistryRecordIsImmutableAndIdempotent(t *testing.T) {
	now := time.Date(2026, time.August, 26, 23, 0, 0, 0, time.UTC)
	r := NewEvidenceEnvelopeRegistry()
	v := registryEnvelope(now)
	if _, err := r.recordAt(v, now); err != nil {
		t.Fatalf("record: %v", err)
	}
	if _, err := r.recordAt(v, now); err != nil {
		t.Fatalf("idempotent replay should succeed: %v", err)
	}
	changed := v
	changed.Outcome = "attention"
	if _, err := r.recordAt(changed, now); err == nil {
		t.Fatal("expected immutable evidence id conflict")
	}
}

func TestEvidenceEnvelopeRegistrySeparatesCurrentFromStale(t *testing.T) {
	now := time.Date(2026, time.August, 26, 23, 0, 0, 0, time.UTC)
	r := NewEvidenceEnvelopeRegistry()
	if _, err := r.recordAt(registryEnvelope(now), now); err != nil {
		t.Fatal(err)
	}
	if got := len(r.CurrentAt(now)); got != 1 {
		t.Fatalf("current count = %d", got)
	}
	current, stale := r.CountsAt(now.Add(2 * time.Hour))
	if current != 0 || stale != 1 {
		t.Fatalf("counts = current:%d stale:%d", current, stale)
	}
}

func TestPersistentEvidenceEnvelopeRegistryReloadsExpiredEvidence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	now := time.Now().UTC()
	r, err := NewPersistentEvidenceEnvelopeRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	v := registryEnvelope(now)
	v.ValidUntil = now.Add(50 * time.Millisecond)
	if _, err := r.Record(v); err != nil {
		t.Fatal(err)
	}

	time.Sleep(80 * time.Millisecond)
	reloaded, err := NewPersistentEvidenceEnvelopeRegistry(path)
	if err != nil {
		t.Fatalf("expired evidence must not block restart: %v", err)
	}
	if _, ok := reloaded.Get(v.ID); !ok {
		t.Fatal("expected expired evidence to remain available for audit")
	}
	if got := len(reloaded.CurrentAt(time.Now().UTC())); got != 0 {
		t.Fatalf("expired evidence must not be current; got %d", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("evidence registry permissions too broad: %o", info.Mode().Perm())
	}
}

func TestEvidenceEnvelopeRejectsCrossProducerContract(t *testing.T) {
	now := time.Date(2026, time.August, 26, 23, 0, 0, 0, time.UTC)
	v := registryEnvelope(now)
	v.Producer.Contract = "contracts/privacy-shield.status.schema.json"
	if _, err := normalizeEvidenceEnvelopeAt(v, now); err == nil {
		t.Fatal("expected cross-producer contract authority mismatch")
	}
}
