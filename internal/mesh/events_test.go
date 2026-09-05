package mesh

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

func TestServiceLifecycleEventUsesVersionedBoundedContract(t *testing.T) {
	m := testMesh(t)
	events, cancel := m.Subscribe(1)
	defer cancel()

	_, err := m.RegisterService(model.Service{
		ID:     "identity",
		Name:   "GoreeCloud Identity",
		Kind:   "platform",
		Health: model.HealthHealthy,
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if err := ValidateEvent(event); err != nil {
			t.Fatalf("published event must satisfy the event contract: %v", err)
		}
		if event.Schema != EventSchemaV1 || event.Type != EventServiceUpsertedV1 {
			t.Fatalf("unexpected event identity: %#v", event)
		}
		if event.Source != "identity" || event.Subject != "identity" {
			t.Fatalf("unexpected source/subject: %#v", event)
		}
		if event.AuthorityTransfer {
			t.Fatal("Mesh lifecycle event must not transfer authority")
		}
		if got := event.Data["health"]; got != string(model.HealthHealthy) {
			t.Fatalf("unexpected health payload: %#v", event.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("expected service lifecycle event")
	}
}

func TestRelationshipLifecycleEventUsesClosedPayload(t *testing.T) {
	m := testMesh(t)
	events, cancel := m.Subscribe(4)
	defer cancel()

	_, _ = m.RegisterService(model.Service{ID: "manager", Name: "GoreeCloud Manager", Kind: "application"})
	_, _ = m.RegisterService(model.Service{ID: "identity", Name: "GoreeCloud Identity", Kind: "platform"})
	<-events
	<-events

	_, err := m.AddRelationship(model.Relationship{
		ID:      "manager-identity-auth",
		From:    "manager",
		To:      "identity",
		Type:    "consumes",
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	event := <-events
	if err := ValidateEvent(event); err != nil {
		t.Fatalf("published relationship event must satisfy the event contract: %v", err)
	}
	if event.Type != EventRelationshipUpsertedV1 {
		t.Fatalf("unexpected event type: %q", event.Type)
	}
	if len(event.Data) != 2 || event.Data["target"] != "identity" || event.Data["type"] != "consumes" {
		t.Fatalf("unexpected relationship payload: %#v", event.Data)
	}
}

func TestSubscribeTypesMinimizesLocalDelivery(t *testing.T) {
	m := testMesh(t)
	events, cancel, err := m.SubscribeTypes(4, EventRelationshipUpsertedV1)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	_, err = m.RegisterService(model.Service{ID: "manager", Name: "GoreeCloud Manager", Kind: "application"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.RegisterService(model.Service{ID: "identity", Name: "GoreeCloud Identity", Kind: "platform"})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		t.Fatalf("relationship-only subscriber received unrelated service event: %#v", event)
	default:
	}

	_, err = m.AddRelationship(model.Relationship{
		ID:      "manager-identity-auth",
		From:    "manager",
		To:      "identity",
		Type:    "consumes",
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Type != EventRelationshipUpsertedV1 {
			t.Fatalf("unexpected filtered event type: %q", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("expected relationship event for relationship-only subscriber")
	}
}

func TestSubscribeTypesRejectsUnknownEventType(t *testing.T) {
	m := testMesh(t)
	if _, _, err := m.SubscribeTypes(4, "mesh.application.payload.v1"); err == nil {
		t.Fatal("filtered subscription must reject unregistered event types")
	}
	if _, _, err := m.SubscribeTypes(4); err == nil {
		t.Fatal("filtered subscription must require at least one event type")
	}
}

func TestSubscribeClampsLocalBuffer(t *testing.T) {
	m := testMesh(t)
	events, cancel := m.Subscribe(maxEventSubscriberBuffer + 1000)
	defer cancel()
	if got := cap(events); got != maxEventSubscriberBuffer {
		t.Fatalf("subscriber buffer = %d, want hard ceiling %d", got, maxEventSubscriberBuffer)
	}
}

func TestEventContractRejectsAuthorityTransferAndArbitraryPayloads(t *testing.T) {
	base := model.Event{
		Schema:    EventSchemaV1,
		ID:        "evt-1",
		Type:      EventServiceUpsertedV1,
		Source:    "identity",
		Subject:   "identity",
		Data:      map[string]any{"health": "healthy"},
		CreatedAt: time.Now().UTC(),
	}
	if err := ValidateEvent(base); err != nil {
		t.Fatalf("valid fixture rejected: %v", err)
	}

	withSecretField := base
	withSecretField.Data = map[string]any{"health": "healthy", "authorization": "Bearer secret"}
	if err := ValidateEvent(withSecretField); err == nil {
		t.Fatal("event contract must reject arbitrary credential-bearing fields")
	}

	withAuthorityTransfer := base
	withAuthorityTransfer.AuthorityTransfer = true
	if err := ValidateEvent(withAuthorityTransfer); err == nil {
		t.Fatal("event contract must reject authority transfer")
	}

	withUnknownType := base
	withUnknownType.Type = "mesh.application.payload.v1"
	if err := ValidateEvent(withUnknownType); err == nil {
		t.Fatal("event contract must reject unregistered event types")
	}
}

func TestEventSchemaArtifactIsValidJSONAndNamesCurrentContract(t *testing.T) {
	raw, err := os.ReadFile("../../contracts/mesh.event.v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("event schema must be valid JSON: %v", err)
	}
	if got := schema["$id"]; got != "https://mesh.goreecloud.com/contracts/mesh.event.v1.schema.json" {
		t.Fatalf("unexpected event schema id: %#v", got)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("event schema properties are required")
	}
	schemaProperty, ok := properties["schema"].(map[string]any)
	if !ok || schemaProperty["const"] != EventSchemaV1 {
		t.Fatal("event schema artifact must bind the Go event schema identifier")
	}
}
