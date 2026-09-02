package platformregistry

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const Schema = "goreecloud.mesh.platform-record.v1"

var revisionPattern = regexp.MustCompile(`^[a-fA-F0-9]{40,64}$`)

type Source struct {
	Repository            string `json:"repository"`
	Revision              string `json:"revision"`
	ContractSchemaVersion string `json:"contract_schema_version"`
	AuthorityTransfer     bool   `json:"authority_transfer"`
}

type Component struct {
	ID                 string   `json:"id"`
	ProductName        string   `json:"product_name"`
	Kind               string   `json:"kind"`
	Repository         string   `json:"repository"`
	Lifecycle          string   `json:"lifecycle"`
	Version            string   `json:"version"`
	SupportedPlatforms []string `json:"supported_platforms"`
}

type Relationship struct {
	Target   string `json:"target"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type PlatformSystem struct {
	Status       string   `json:"status"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

type Health struct {
	RuntimeState string `json:"runtime_state"`
	HealthState  string `json:"health_state"`
	Readiness    string `json:"readiness"`
}

type Recovery struct {
	BackupStatus        string     `json:"backup_status"`
	RestoreStatus       string     `json:"restore_status"`
	LastVerifiedRestore *time.Time `json:"last_verified_restore,omitempty"`
}

type Portability struct {
	ExportStatus string   `json:"export_status"`
	Formats      []string `json:"formats,omitempty"`
}

type Conformance struct {
	OverallResult            string   `json:"overall_result"`
	StableEligible           bool     `json:"stable_eligible"`
	MissingMandatoryEvidence []string `json:"missing_mandatory_evidence,omitempty"`
}

type EvidenceRef struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Location string `json:"location"`
	Producer string `json:"producer"`
}

type Record struct {
	Schema          string                    `json:"schema"`
	Source          Source                    `json:"source"`
	Component       Component                 `json:"component"`
	Capabilities    []string                  `json:"capabilities,omitempty"`
	Dependencies    []string                  `json:"dependencies,omitempty"`
	Relationships   []Relationship            `json:"relationships,omitempty"`
	PlatformSystems map[string]PlatformSystem `json:"platform_systems"`
	Health          Health                    `json:"health"`
	Recovery        Recovery                  `json:"recovery"`
	Portability     Portability               `json:"portability"`
	Conformance     Conformance               `json:"conformance"`
	EvidenceRefs    []EvidenceRef             `json:"evidence_refs,omitempty"`
	ObservedAt      time.Time                 `json:"observed_at"`
}

type Registry struct {
	mu      sync.RWMutex
	records map[string]Record
}

func New() *Registry {
	return &Registry{records: map[string]Record{}}
}

func Validate(record Record) error {
	if record.Schema != Schema {
		return fmt.Errorf("unsupported platform record schema %q", record.Schema)
	}
	if record.Source.AuthorityTransfer {
		return errors.New("authority_transfer must remain false")
	}
	if strings.TrimSpace(record.Source.Repository) == "" || record.Source.Repository != record.Component.Repository {
		return errors.New("source repository must exactly match component repository")
	}
	if !revisionPattern.MatchString(record.Source.Revision) {
		return errors.New("source revision must be a 40-64 character hexadecimal immutable revision")
	}
	if record.Source.ContractSchemaVersion != "1.0" {
		return fmt.Errorf("unsupported platform contract schema %q", record.Source.ContractSchemaVersion)
	}
	if strings.TrimSpace(record.Component.ID) == "" || strings.TrimSpace(record.Component.ProductName) == "" || strings.TrimSpace(record.Component.Kind) == "" || strings.TrimSpace(record.Component.Version) == "" {
		return errors.New("component id, product_name, kind, and version are required")
	}
	if !validLifecycle(record.Component.Lifecycle) {
		return fmt.Errorf("invalid lifecycle %q", record.Component.Lifecycle)
	}
	if len(unique(record.Component.SupportedPlatforms)) == 0 {
		return errors.New("at least one supported platform is required")
	}
	if record.ObservedAt.IsZero() {
		return errors.New("observed_at is required")
	}
	if record.Conformance.OverallResult != "conformant" && record.Conformance.OverallResult != "non-conformant" {
		return fmt.Errorf("invalid conformance result %q", record.Conformance.OverallResult)
	}
	if record.Conformance.OverallResult != "conformant" && record.Conformance.StableEligible {
		return errors.New("non-conformant records cannot be Stable-eligible")
	}
	if record.Component.Lifecycle == "Stable" && !record.Conformance.StableEligible {
		return errors.New("Stable component record must carry current Stable eligibility")
	}
	if record.Recovery.RestoreStatus == "verified" && record.Recovery.LastVerifiedRestore == nil {
		return errors.New("verified restore requires last_verified_restore")
	}
	if record.Recovery.RestoreStatus != "verified" && record.Recovery.LastVerifiedRestore != nil {
		return errors.New("last_verified_restore cannot be set when restore_status is not verified")
	}
	if err := validatePlatformSystems(record.PlatformSystems); err != nil {
		return err
	}
	if err := validateRelationships(record.Component.ID, record.Relationships); err != nil {
		return err
	}
	if err := validateEvidence(record.EvidenceRefs); err != nil {
		return err
	}
	if hasDuplicates(record.Capabilities) || hasDuplicates(record.Dependencies) || hasDuplicates(record.Conformance.MissingMandatoryEvidence) {
		return errors.New("capabilities, dependencies, and missing evidence identifiers must be unique")
	}
	return nil
}

func (r *Registry) Upsert(record Record) error {
	if err := Validate(record); err != nil {
		return err
	}
	record = normalized(record)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.Component.ID] = record
	return nil
}

func (r *Registry) Get(id string) (Record, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.records[strings.TrimSpace(id)]
	return clone(record), ok
}

func (r *Registry) List() []Record {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Record, 0, len(r.records))
	for _, record := range r.records {
		out = append(out, clone(record))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Component.ID < out[j].Component.ID })
	return out
}

// Dependents returns component IDs whose declared dependency graph directly or
// transitively depends on id. It never infers undeclared relationships.
func (r *Registry) Dependents(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reverse := map[string][]string{}
	for _, record := range r.records {
		for _, dependency := range record.Dependencies {
			reverse[dependency] = append(reverse[dependency], record.Component.ID)
		}
		for _, relationship := range record.Relationships {
			if relationship.Required {
				reverse[relationship.Target] = append(reverse[relationship.Target], record.Component.ID)
			}
		}
	}
	seen := map[string]bool{id: true}
	queue := []string{id}
	out := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range reverse[current] {
			if seen[dependent] {
				continue
			}
			seen[dependent] = true
			out = append(out, dependent)
			queue = append(queue, dependent)
		}
	}
	sort.Strings(out)
	return out
}

func validLifecycle(value string) bool {
	switch value {
	case "Concept", "Experimental", "Development", "Release Candidate", "Stable", "Deprecated", "Retired":
		return true
	default:
		return false
	}
}

func validatePlatformSystems(systems map[string]PlatformSystem) error {
	required := []string{"manager", "identity", "wardveil_security", "privacy_shield", "everkeep", "mesh", "glaze_ui"}
	for _, name := range required {
		value, ok := systems[name]
		if !ok {
			return fmt.Errorf("missing platform system evaluation %q", name)
		}
		if strings.TrimSpace(value.Status) == "" {
			return fmt.Errorf("platform system %q status is required", name)
		}
	}
	return nil
}

func validateRelationships(componentID string, relationships []Relationship) error {
	seen := map[string]struct{}{}
	for _, relationship := range relationships {
		if strings.TrimSpace(relationship.Target) == "" || strings.TrimSpace(relationship.Type) == "" {
			return errors.New("relationship target and type are required")
		}
		if relationship.Target == componentID {
			return errors.New("self relationships are not allowed")
		}
		key := relationship.Type + "\x00" + relationship.Target
		if _, ok := seen[key]; ok {
			return errors.New("duplicate relationship")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateEvidence(items []EvidenceRef) error {
	seen := map[string]struct{}{}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Kind) == "" || strings.TrimSpace(item.Location) == "" || strings.TrimSpace(item.Producer) == "" {
			return errors.New("evidence id, kind, location, and producer are required")
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate evidence id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}

func normalized(record Record) Record {
	record.Capabilities = unique(record.Capabilities)
	record.Dependencies = unique(record.Dependencies)
	record.Component.SupportedPlatforms = unique(record.Component.SupportedPlatforms)
	record.Conformance.MissingMandatoryEvidence = unique(record.Conformance.MissingMandatoryEvidence)
	return record
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func hasDuplicates(values []string) bool {
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func clone(record Record) Record {
	record.Capabilities = append([]string(nil), record.Capabilities...)
	record.Dependencies = append([]string(nil), record.Dependencies...)
	record.Relationships = append([]Relationship(nil), record.Relationships...)
	record.Component.SupportedPlatforms = append([]string(nil), record.Component.SupportedPlatforms...)
	record.Conformance.MissingMandatoryEvidence = append([]string(nil), record.Conformance.MissingMandatoryEvidence...)
	record.EvidenceRefs = append([]EvidenceRef(nil), record.EvidenceRefs...)
	if record.PlatformSystems != nil {
		copySystems := make(map[string]PlatformSystem, len(record.PlatformSystems))
		for key, value := range record.PlatformSystems {
			value.EvidenceRefs = append([]string(nil), value.EvidenceRefs...)
			copySystems[key] = value
		}
		record.PlatformSystems = copySystems
	}
	return record
}
