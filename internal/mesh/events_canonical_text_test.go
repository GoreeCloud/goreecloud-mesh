package mesh

import (
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

func canonicalServiceEvent() model.Event {
	return model.Event{
		Schema:    EventSchemaV1,
		ID:        "evt-1",
		Type:      EventServiceUpsertedV1,
		Source:    "identity",
		Subject:   "identity",
		Data:      map[string]any{"health": "healthy"},
		CreatedAt: time.Now().UTC(),
	}
}

func TestEventContractRejectsSurroundingWhitespaceInIdentityFields(t *testing.T) {
	for name, mutate := range map[string]func(*model.Event){
		"source":  func(event *model.Event) { event.Source = " identity" },
		"subject": func(event *model.Event) { event.Subject = "identity " },
	} {
		t.Run(name, func(t *testing.T) {
			event := canonicalServiceEvent()
			mutate(&event)
			if err := ValidateEvent(event); err == nil {
				t.Fatal("event identity with surrounding whitespace must fail closed")
			}
		})
	}
}

func TestEventContractRejectsSurroundingWhitespaceInPayloadValues(t *testing.T) {
	service := canonicalServiceEvent()
	service.Data = map[string]any{"health": " healthy"}
	if err := ValidateEvent(service); err == nil {
		t.Fatal("service event health with surrounding whitespace must fail closed")
	}

	relationship := model.Event{
		Schema:    EventSchemaV1,
		ID:        "evt-2",
		Type:      EventRelationshipUpsertedV1,
		Source:    "manager",
		Subject:   "relationship-1",
		Data:      map[string]any{"target": " identity", "type": "consumes"},
		CreatedAt: time.Now().UTC(),
	}
	if err := ValidateEvent(relationship); err == nil {
		t.Fatal("relationship target with surrounding whitespace must fail closed")
	}

	relationship.Data = map[string]any{"target": "identity", "type": "consumes "}
	if err := ValidateEvent(relationship); err == nil {
		t.Fatal("relationship type with surrounding whitespace must fail closed")
	}
}
