package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type contract struct {
	Project          string `json:"project"`
	DevelopmentModel string `json:"development_model"`
	StableEligible   bool   `json:"stable_eligible"`
	Systems          map[string]struct {
		Required bool     `json:"required"`
		State    string   `json:"state"`
		Evidence []string `json:"evidence"`
	} `json:"systems"`
	AcceptanceGates []string `json:"acceptance_gates"`
}

func TestPlatformConformanceTruthContract(t *testing.T) {
	_, here, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(here), "..", "..", "docs", "platform-conformance.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c contract
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	if c.Project != "GoreeCloud Mesh" {
		t.Fatalf("unexpected project %q", c.Project)
	}
	if c.DevelopmentModel != "original-native-goreecloud" {
		t.Fatalf("unexpected development model %q", c.DevelopmentModel)
	}
	required := []string{"glaze_ui", "wardveil_security", "privacy_shield", "everkeep"}
	pending := false
	for _, name := range required {
		system, ok := c.Systems[name]
		if !ok || !system.Required {
			t.Fatalf("mandatory system %q missing or optional", name)
		}
		if len(system.Evidence) == 0 {
			t.Fatalf("mandatory system %q lacks evidence", name)
		}
		switch system.State {
		case "pending", "blocked":
			pending = true
		case "validated":
		default:
			t.Fatalf("mandatory system %q has invalid state %q", name, system.State)
		}
	}
	if pending && c.StableEligible {
		t.Fatal("Stable eligibility must be false while any mandatory platform system is pending or blocked")
	}
	if len(c.AcceptanceGates) == 0 {
		t.Fatal("acceptance gates must be explicit")
	}
}
