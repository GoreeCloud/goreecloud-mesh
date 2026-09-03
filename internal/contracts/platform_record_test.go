package contracts

import (
	"strings"
	"testing"
	"time"
)

func validPlatformRecord() PlatformRecord {
	systems := map[string]PlatformSystemRecord{}
	for _, key := range platformSystemKeys {
		systems[key] = PlatformSystemRecord{
		Result:   "applicable-blocked",
		Evidence: []string{"evidence/" + key + ".json"},
	}
	return PlatformRecord{
		ComponentID:           "example-app",
		ProductName:           "Example App",
		ComponentType:         "application",
		Repository:            "GoreeCloud/example-app",
		Lifecycle:             "development",
		Version:               "0.1.0-dev",
		ManifestSchemaVersion: PlatformManifestSchemaVersion,
		ResultSchemaVersion:   PlatformResultSchemaVersion,
		EvaluatedRevision:     strings.Repeat("a", 40),
		EvaluatorRepository:   PlatformEvaluatorRepository,
		EvaluatorRevision:     strings.Repeat("b", 40),
		EvaluatedAt:           time.Date(2026, 9, 3, 18, 0, 0, 0, time.UTC),
		SupportedPlatforms:    []string{"web"},
		PlatformSystems:       systems,
		Capabilities:          []string{"example.read"},
		Dependencies:          []string{"goreecloud-identity"},
		CompatibilityRequires: []string{"goreecloud-platform-contract==0.2", "glaze-ui==1.1.0"},
		ComputedConformance:   "nonconformant",
		StableEligible:        false,
		Blockers:              []string{"acceptance evidence incomplete"},
		Evidence:              []string{"goreecloud.conformance-result.json"},
		Authority: PlatformRecordAuthority{
			DeclarationOwner:  "GoreeCloud/example-app",
			AggregationRole:   PlatformAggregationRoleReadOnly,
			AuthorityTransfer: false,
		},
	}
}

func TestPlatformRecordRegistryPreservesSourceAttributedState(t *testing.T) {
	r := NewPlatformRecordRegistry()
	record, err := r.Record(validPlatformRecord())
	if err != nil {
		t.Fatal(err)
	}
	if record.ComponentID != "example-app" || record.ComputedConformance != "nonconformant" {
		t.Fatalf("unexpected record: %+v", record)
	}
	got, ok := r.Get("example-app")
	if !ok {
		t.Fatal("record missing after insert")
	}
	if got.Authority.DeclarationOwner != got.Repository || got.Authority.AggregationRole != PlatformAggregationRoleReadOnly || got.Authority.AuthorityTransfer {
		t.Fatalf("authority boundary changed: %+v", got.Authority)
	}
}

func TestPlatformRecordRejectsAuthorityTransfer(t *testing.T) {
	v := validPlatformRecord()
	v.Authority.AuthorityTransfer = true
	if _, err := normalizePlatformRecord(v, time.Date(2026, 9, 3, 19, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("authority transfer must be rejected")
	}
}

func TestPlatformRecordRequiresExactEvaluatorProvenance(t *testing.T) {
	v := validPlatformRecord()
	v.EvaluatorRevision = "abc123"
	if _, err := normalizePlatformRecord(v, time.Date(2026, 9, 3, 19, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("short evaluator revision must be rejected")
	}

	v = validPlatformRecord()
	v.EvaluatorRepository = "GoreeCloud/other"
	if _, err := normalizePlatformRecord(v, time.Date(2026, 9, 3, 19, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("non-canonical evaluator repository must be rejected")
	}
}

func TestPlatformRecordRequiresSevenSystemResults(t *testing.T) {
	v := validPlatformRecord()
	delete(v.PlatformSystems, "identity")
	if _, err := normalizePlatformRecord(v, time.Date(2026, 9, 3, 19, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("missing Integral Platform System must be rejected")
	}
}

func TestPlatformRecordStableEligibilityFailsClosed(t *testing.T) {
	v := validPlatformRecord()
	v.Lifecycle = "stable"
	v.ComputedConformance = "conformant"
	v.StableEligible = true
	v.Blockers = nil
	if _, err := normalizePlatformRecord(v, time.Date(2026, 9, 3, 19, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("unresolved platform-system results must block Stable eligibility")
	}
}

func TestPlatformRecordRegistryReturnsCopies(t *testing.T) {
	r := NewPlatformRecordRegistry()
	if _, err := r.Record(validPlatformRecord()); err != nil {
		t.Fatal(err)
	}
	got, ok := r.Get("example-app")
	if !ok {
		t.Fatal("record missing")
	}
	got.SupportedPlatforms[0] = "tampered"
	got.PlatformSystems["manager"] = PlatformSystemRecord{Result: "applicable-conformant", Evidence: []string{"tampered"}}

	again, ok := r.Get("example-app")
	if !ok {
		t.Fatal("record missing")
	}
	if again.SupportedPlatforms[0] != "web" || again.PlatformSystems["manager"].Result != "applicable-blocked" {
		t.Fatalf("stored record mutated through caller copy: %+v", again)
	}
}
