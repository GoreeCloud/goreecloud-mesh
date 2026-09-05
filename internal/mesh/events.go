package mesh

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/GoreeCloud/goreecloud-mesh/internal/model"
)

const (
	EventSchemaV1               = "goreecloud.mesh.event.v1"
	EventServiceUpsertedV1      = "mesh.service.upserted.v1"
	EventRelationshipUpsertedV1 = "mesh.relationship.upserted.v1"

	maxEventIdentityRunes = 128
	maxEventValueRunes    = 256
)

var (
	eventIDPattern       = regexp.MustCompile(`^evt-[1-9][0-9]*$`)
	eventTypeNamePattern = regexp.MustCompile(
		`^mesh(?:\.[a-z][a-z0-9-]*){2,}\.v[1-9][0-9]*$`,
	)
)

func newEvent(seq uint64, kind, source, subject string, data map[string]any, createdAt time.Time) (model.Event, error) {
	e := model.Event{
		Schema:            EventSchemaV1,
		ID:                fmt.Sprintf("evt-%d", seq),
		Type:              kind,
		Source:            strings.TrimSpace(source),
		Subject:           strings.TrimSpace(subject),
		Data:              data,
		CreatedAt:         createdAt.UTC(),
		AuthorityTransfer: false,
	}
	if err := ValidateEvent(e); err != nil {
		return model.Event{}, err
	}
	return e, nil
}

// ValidateEvent enforces the current bounded in-process Mesh lifecycle-event
// contract. It is intentionally closed: adding event types or payload fields is
// a versioned contract change rather than an invitation to carry arbitrary
// application data through Mesh.
func ValidateEvent(e model.Event) error {
	if e.Schema != EventSchemaV1 {
		return fmt.Errorf("unsupported event schema %q", e.Schema)
	}
	if !eventTypeNamePattern.MatchString(e.Type) {
		return errors.New(
			"event type must use the mesh.<subject>.<action>.v<major> lowercase namespace",
		)
	}
	if !eventIDPattern.MatchString(e.ID) {
		return errors.New("event id must use the process-local evt-<sequence> form")
	}
	if err := validateEventText("source", e.Source, maxEventIdentityRunes); err != nil {
		return err
	}
	if err := validateEventText("subject", e.Subject, maxEventIdentityRunes); err != nil {
		return err
	}
	if e.CreatedAt.IsZero() {
		return errors.New("event created_at is required")
	}
	if e.AuthorityTransfer {
		return errors.New("Mesh events must not transfer producer authority")
	}

	switch e.Type {
	case EventServiceUpsertedV1:
		if e.Source != e.Subject {
			return errors.New("service lifecycle event source must match subject")
		}
		if err := requireEventDataKeys(e.Data, "health"); err != nil {
			return err
		}
		health, err := eventDataString(e.Data, "health")
		if err != nil {
			return err
		}
		if !validHealth(model.HealthState(health)) {
			return fmt.Errorf("invalid service event health %q", health)
		}
	case EventRelationshipUpsertedV1:
		if err := requireEventDataKeys(e.Data, "target", "type"); err != nil {
			return err
		}
		target, err := eventDataString(e.Data, "target")
		if err != nil {
			return err
		}
		if err := validateEventText("data.target", target, maxEventIdentityRunes); err != nil {
			return err
		}
		relationshipType, err := eventDataString(e.Data, "type")
		if err != nil {
			return err
		}
		if err := validateEventText("data.type", relationshipType, maxEventValueRunes); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported event type %q", e.Type)
	}

	return nil
}

func requireEventDataKeys(data map[string]any, required ...string) error {
	if len(data) != len(required) {
		return errors.New("event data contains an unsupported or missing field")
	}
	for _, key := range required {
		if _, ok := data[key]; !ok {
			return fmt.Errorf("event data field %q is required", key)
		}
	}
	return nil
}

func eventDataString(data map[string]any, key string) (string, error) {
	value, ok := data[key]
	if !ok {
		return "", fmt.Errorf("event data field %q is required", key)
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("event data field %q must be a string", key)
	}
	text = strings.TrimSpace(text)
	if err := validateEventText("data."+key, text, maxEventValueRunes); err != nil {
		return "", err
	}
	return text, nil
}

func validateEventText(name, value string, maxRunes int) error {
	if value == "" {
		return fmt.Errorf("event %s is required", name)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("event %s must be valid UTF-8", name)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("event %s exceeds %d characters", name, maxRunes)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("event %s contains control characters", name)
		}
	}
	return nil
}
