package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

const canonicalEverkeepAdoptionSchema = "GoreeCloud/goreecloud-everkeep:contracts/continuity.adoption.schema.json@v1"
const canonicalEverkeepAcceptanceSchema = "GoreeCloud/goreecloud-everkeep:contracts/continuity.acceptance.schema.json@v1"

type everkeepAdoption struct {
	SchemaVersion        int      `json:"schema_version"`
	Project              string   `json:"project"`
	Repository           string   `json:"repository"`
	Role                 string   `json:"role"`
	Dimensions           []string `json:"dimensions"`
	ReadOnly             bool     `json:"read_only"`
	FailClosed           bool     `json:"fail_closed"`
	StatusSchema         string   `json:"status_schema"`
	AuthoritativeSources []string `json:"authoritative_sources"`
}

type everkeepAcceptance struct {
	SchemaVersion int    `json:"schema_version"`
	Application   string `json:"application"`
	Producer      string `json:"producer"`
	StatusSchema  string `json:"status_schema"`
	Freshness     struct {
		RequiredForReady bool   `json:"required_for_ready"`
		Rule             string `json:"rule"`
	} `json:"freshness"`
	FailureBehavior map[string]string   `json:"failure_behavior"`
	RequiredReady   map[string][]string `json:"required_ready_evidence"`
	Sensitive       struct {
		Forbidden []string `json:"forbidden"`
	} `json:"sensitive_evidence"`
	Acceptance struct {
		Integrated            bool `json:"everkeep_integrated"`
		Ready                 bool `json:"everkeep_ready"`
		TargetRuntimeRequired bool `json:"target_runtime_acceptance_required"`
		ExactRevisionRequired bool `json:"exact_revision_acceptance_required"`
	} `json:"acceptance"`
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test path")
	}
	return filepath.Join(filepath.Dir(here), "..", "..")
}

func readJSON(t *testing.T, path string, out any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}

func TestEverkeepAdoptionManifestIsBoundedAndFailClosed(t *testing.T) {
	root := repositoryRoot(t)
	var adoption everkeepAdoption
	readJSON(t, filepath.Join(root, "docs", "everkeep.adoption.json"), &adoption)
	if adoption.SchemaVersion != 1 || adoption.Project != "goreecloud-mesh" || adoption.Repository != "GoreeCloud/goreecloud-mesh" {
		t.Fatalf("unexpected adoption identity: %+v", adoption)
	}
	if adoption.Role != "producer_consumer" || !adoption.ReadOnly || !adoption.FailClosed {
		t.Fatal("Mesh Everkeep adoption must remain read-only, fail-closed producer-consumer")
	}
	if adoption.StatusSchema != "contracts/continuity.status.schema.json" {
		t.Fatalf("unexpected status schema %q", adoption.StatusSchema)
	}
	got := append([]string(nil), adoption.Dimensions...)
	sort.Strings(got)
	want := []string{"backup_coverage", "documentation", "portability", "provenance", "restore_capability"}
	if len(got) != len(want) {
		t.Fatalf("unexpected dimensions %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected dimensions %v", got)
		}
	}
	for _, source := range adoption.AuthoritativeSources {
		if _, err := os.Stat(filepath.Join(root, source)); err != nil {
			t.Fatalf("missing authoritative source %q: %v", source, err)
		}
	}
}

func TestEverkeepAcceptancePolicyCannotClaimReady(t *testing.T) {
	root := repositoryRoot(t)
	var policy everkeepAcceptance
	readJSON(t, filepath.Join(root, "docs", "everkeep.acceptance.json"), &policy)
	if policy.SchemaVersion != 1 || policy.Application != "goreecloud-mesh" || policy.Producer != "GoreeCloud/goreecloud-mesh" {
		t.Fatalf("unexpected acceptance identity: %+v", policy)
	}
	if policy.StatusSchema != "contracts/continuity.status.schema.json" || !policy.Freshness.RequiredForReady || policy.Freshness.Rule == "" {
		t.Fatal("Everkeep acceptance must require canonical status and explicit freshness")
	}
	for _, key := range []string{"producer_unavailable", "malformed_evidence", "missing_required_evidence", "failed_restore_validation", "summary_rule"} {
		if policy.FailureBehavior[key] == "" {
			t.Fatalf("missing failure behavior %q", key)
		}
	}
	for _, dimension := range []string{"backup_coverage", "restore_capability", "portability", "documentation", "provenance"} {
		if len(policy.RequiredReady[dimension]) == 0 {
			t.Fatalf("missing required-ready evidence for %q", dimension)
		}
	}
	if policy.Acceptance.Integrated || policy.Acceptance.Ready {
		t.Fatal("source policy must not claim Everkeep integration or readiness")
	}
	if !policy.Acceptance.TargetRuntimeRequired || !policy.Acceptance.ExactRevisionRequired {
		t.Fatal("target-runtime and exact-revision acceptance must remain required")
	}
	forbidden := map[string]bool{}
	for _, item := range policy.Sensitive.Forbidden {
		forbidden[item] = true
	}
	for _, item := range []string{"passwords", "tokens", "private keys", "secret values", "application payload content"} {
		if !forbidden[item] {
			t.Fatalf("sensitive evidence %q must remain forbidden", item)
		}
	}
}

func TestPlatformConformanceBindsCanonicalEverkeepSchemas(t *testing.T) {
	root := repositoryRoot(t)
	var raw struct {
		StableEligible bool `json:"stable_eligible"`
		Systems        map[string]struct {
			State                     string `json:"state"`
			CanonicalAdoptionSchema   string `json:"canonical_adoption_schema"`
			CanonicalAcceptanceSchema string `json:"canonical_acceptance_schema"`
		} `json:"systems"`
	}
	readJSON(t, filepath.Join(root, "docs", "platform-conformance.json"), &raw)
	everkeep, ok := raw.Systems["everkeep"]
	if !ok {
		t.Fatal("Everkeep conformance entry missing")
	}
	if everkeep.State != "pending" || raw.StableEligible {
		t.Fatal("Everkeep must remain pending and Stable-ineligible")
	}
	if everkeep.CanonicalAdoptionSchema != canonicalEverkeepAdoptionSchema {
		t.Fatalf("unexpected adoption schema binding %q", everkeep.CanonicalAdoptionSchema)
	}
	if everkeep.CanonicalAcceptanceSchema != canonicalEverkeepAcceptanceSchema {
		t.Fatalf("unexpected acceptance schema binding %q", everkeep.CanonicalAcceptanceSchema)
	}
}
