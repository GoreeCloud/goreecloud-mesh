package platformregistry

import (
	"testing"
	"time"
)

func fixture() Record {
	now := time.Date(2026, 9, 2, 21, 0, 0, 0, time.UTC)
	return Record{
		Schema: Schema,
		Source: Source{
			Repository:            "GoreeCloud/goreecloud-tasks",
			Revision:              "0123456789abcdef0123456789abcdef01234567",
			ContractSchemaVersion: "1.0",
			AuthorityTransfer:     false,
		},
		Component: Component{
			ID:                 "goreecloud.tasks",
			ProductName:        "GoreeCloud Tasks",
			Kind:               "application",
			Repository:         "GoreeCloud/goreecloud-tasks",
			Lifecycle:          "Development",
			Version:            "0.1",
			SupportedPlatforms: []string{"web", "linux-server"},
		},
		Capabilities: []string{"task-management", "portable-user-export"},
		Dependencies: []string{"goreecloud.mesh"},
		Relationships: []Relationship{
			{Target: "goreecloud.manager", Type: "observed-by", Required: false},
		},
		PlatformSystems: map[string]PlatformSystem{
			"manager":           {Status: "Applicable — Nonconformant"},
			"identity":          {Status: "Applicable — Migration Required"},
			"wardveil_security": {Status: "Applicable — Nonconformant"},
			"privacy_shield":    {Status: "Applicable — Nonconformant"},
			"everkeep":          {Status: "Applicable — Nonconformant"},
			"mesh":              {Status: "Applicable — Nonconformant"},
			"glaze_ui":          {Status: "Applicable — Migration Required"},
		},
		Health: Health{RuntimeState: "unknown", HealthState: "unknown", Readiness: "unknown"},
		Recovery: Recovery{
			BackupStatus:  "required_missing",
			RestoreStatus: "required_missing",
		},
		Portability: Portability{ExportStatus: "implemented_unverified", Formats: []string{"goreecloud-tasks-user-archive-json"}},
		Conformance: Conformance{
			OverallResult:            "non-conformant",
			StableEligible:           false,
			MissingMandatoryEvidence: []string{"restore", "mesh-registration", "glaze-ui"},
		},
		EvidenceRefs: []EvidenceRef{
			{ID: "tasks-readme", Kind: "documentation", Location: "README.md", Producer: "GoreeCloud/goreecloud-tasks"},
		},
		ObservedAt: now,
	}
}

func TestRegistryAcceptsAuthorityPreservingRecord(t *testing.T) {
	registry := New()
	record := fixture()
	if err := registry.Upsert(record); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	got, ok := registry.Get("goreecloud.tasks")
	if !ok {
		t.Fatal("expected registered component")
	}
	if got.Source.Repository != got.Component.Repository {
		t.Fatal("producer authority was not preserved")
	}
	if got.Conformance.StableEligible {
		t.Fatal("non-conformant component became Stable-eligible")
	}
}

func TestRegistryRejectsAuthorityTransfer(t *testing.T) {
	record := fixture()
	record.Source.AuthorityTransfer = true
	if err := New().Upsert(record); err == nil {
		t.Fatal("expected authority transfer to be rejected")
	}
}

func TestRegistryRejectsProducerRepositoryMismatch(t *testing.T) {
	record := fixture()
	record.Source.Repository = "GoreeCloud/goreecloud-manager"
	if err := New().Upsert(record); err == nil {
		t.Fatal("expected repository mismatch to be rejected")
	}
}

func TestRegistryRejectsFalseStableEligibility(t *testing.T) {
	record := fixture()
	record.Conformance.StableEligible = true
	if err := New().Upsert(record); err == nil {
		t.Fatal("expected false Stable eligibility to be rejected")
	}
}

func TestRegistryRequiresVerifiedRestoreTimestamp(t *testing.T) {
	record := fixture()
	record.Recovery.RestoreStatus = "verified"
	if err := New().Upsert(record); err == nil {
		t.Fatal("expected verified restore without timestamp to be rejected")
	}

	now := time.Now().UTC()
	record.Recovery.LastVerifiedRestore = &now
	if err := New().Upsert(record); err != nil {
		t.Fatalf("verified restore with timestamp rejected: %v", err)
	}
}

func TestRegistryDoesNotTurnBackupIntoRestoreEvidence(t *testing.T) {
	record := fixture()
	record.Recovery.BackupStatus = "verified"
	record.Recovery.RestoreStatus = "required_missing"
	if err := New().Upsert(record); err != nil {
		t.Fatalf("record rejected: %v", err)
	}
	got, _ := NewWith(record).Get("goreecloud.tasks")
	if got.Recovery.RestoreStatus == "verified" {
		t.Fatal("backup status manufactured a verified restore")
	}
}

func TestDependentsUsesOnlyDeclaredDependencies(t *testing.T) {
	registry := New()
	mesh := fixture()
	mesh.Component.ID = "goreecloud.mesh"
	mesh.Component.ProductName = "GoreeCloud Mesh"
	mesh.Component.Repository = "GoreeCloud/goreecloud-mesh"
	mesh.Source.Repository = mesh.Component.Repository
	mesh.Dependencies = nil
	mesh.Relationships = nil
	if err := registry.Upsert(mesh); err != nil {
		t.Fatalf("mesh Upsert() error = %v", err)
	}
	tasks := fixture()
	if err := registry.Upsert(tasks); err != nil {
		t.Fatalf("tasks Upsert() error = %v", err)
	}
	got := registry.Dependents("goreecloud.mesh")
	if len(got) != 1 || got[0] != "goreecloud.tasks" {
		t.Fatalf("Dependents() = %v", got)
	}
}

// NewWith is a test helper that keeps production registry construction explicit.
func NewWith(record Record) *Registry {
	registry := New()
	_ = registry.Upsert(record)
	return registry
}
