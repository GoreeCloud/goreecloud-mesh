package mesh

import (
	"testing"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
)

func testMesh(t *testing.T) *Mesh {
	t.Helper()
	s, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	return New(s)
}

func TestDiscoveryPolicyAndImpact(t *testing.T) {
	m := testMesh(t)
	_, err := m.RegisterService(model.Service{ID: "identity", Name: "GoreeCloud Identity", Kind: "platform", Health: model.HealthHealthy, Capabilities: []string{"identity.authenticate"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.RegisterService(model.Service{ID: "manager", Name: "GoreeCloud Manager", Kind: "application", Health: model.HealthHealthy, Dependencies: []string{"identity"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.RegisterService(model.Service{ID: "notes", Name: "GoreeCloud Notes", Kind: "application", Health: model.HealthHealthy, Dependencies: []string{"manager"}})
	if err != nil {
		t.Fatal(err)
	}

	found := m.Discover("identity.authenticate")
	if len(found) != 1 || found[0].ID != "identity" {
		t.Fatalf("unexpected discovery result: %#v", found)
	}

	denied := m.Evaluate(model.PolicyRequest{Source: "manager", Target: "identity", Capability: "identity.authenticate"})
	if denied.Allowed {
		t.Fatalf("request must fail closed without a relationship: %#v", denied)
	}

	_, err = m.AddRelationship(model.Relationship{ID: "manager-identity-auth", From: "manager", To: "identity", Type: "consumes", Capability: "identity.authenticate", Required: true, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	allowed := m.Evaluate(model.PolicyRequest{Source: "manager", Target: "identity", Capability: "identity.authenticate"})
	if !allowed.Allowed {
		t.Fatalf("expected allowed decision: %#v", allowed)
	}

	impact := m.Impact("identity")
	if len(impact) != 2 || impact[0] != "manager" || impact[1] != "notes" {
		t.Fatalf("unexpected impact: %#v", impact)
	}
}

func TestUnavailableTargetsAreExcludedAndDenied(t *testing.T) {
	m := testMesh(t)
	_, _ = m.RegisterService(model.Service{ID: "notify", Name: "GoreeCloud Notify", Kind: "application", Health: model.HealthUnavailable, Capabilities: []string{"notify.send"}})
	_, _ = m.RegisterService(model.Service{ID: "manager", Name: "GoreeCloud Manager", Kind: "application", Health: model.HealthHealthy})
	_, _ = m.AddRelationship(model.Relationship{ID: "manager-notify", From: "manager", To: "notify", Type: "consumes", Capability: "notify.send", Enabled: true})

	if got := m.Discover("notify.send"); len(got) != 0 {
		t.Fatalf("unavailable target must not be discovered: %#v", got)
	}
	decision := m.Evaluate(model.PolicyRequest{Source: "manager", Target: "notify", Capability: "notify.send"})
	if decision.Allowed || decision.Reason != "target is unavailable" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
}
