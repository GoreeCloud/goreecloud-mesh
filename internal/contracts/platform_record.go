package contracts

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	PlatformManifestSchemaVersion   = "0.2"
	PlatformResultSchemaVersion     = "0.2"
	PlatformEvaluatorRepository     = "GoreeCloud/GoreeCloud"
	PlatformAggregationRoleReadOnly = "read-only"
)

var platformComponentIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var platformRepositoryPattern = regexp.MustCompile(`^GoreeCloud/[A-Za-z0-9._-]+$`)

var platformLifecycleValues = map[string]struct{}{
	"concept":           {},
	"experimental":      {},
	"development":       {},
	"release-candidate": {},
	"stable":            {},
	"deprecated":        {},
	"retired":           {},
}

var platformConformanceValues = map[string]struct{}{
	"conformant":    {},
	"nonconformant": {},
	"unverified":    {},
}

var platformSystemResultValues = map[string]struct{}{
	"applicable-conformant":         {},
	"applicable-migration-required": {},
	"applicable-blocked":            {},
	"applicable-nonconformant":      {},
	"not-applicable-justified":      {},
}

var platformSystemKeys = []string{
	"manager",
	"privacy_shield",
	"wardveil_security",
	"everkeep",
	"glaze_ui",
	"mesh",
	"identity",
}

type PlatformSystemRecord struct {
	Result   string   `json:"result"`
	Version  string   `json:"version,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
}

type PlatformRecordAuthority struct {
	DeclarationOwner  string `json:"declaration_owner"`
	AggregationRole   string `json:"aggregation_role"`
	AuthorityTransfer bool   `json:"authority_transfer"`
}

// PlatformRecord is a normalized, source-attributed coordination view derived
// from one repository's Platform Contract declaration and computed conformance
// result. Mesh may index and correlate this record but cannot upgrade the
// declaration, evaluator result, or producer-domain evidence represented by it.
type PlatformRecord struct {
	ComponentID           string                          `json:"component_id"`
	ProductName           string                          `json:"product_name"`
	ComponentType         string                          `json:"component_type"`
	Repository            string                          `json:"repository"`
	Lifecycle             string                          `json:"lifecycle"`
	Version               string                          `json:"version"`
	ManifestSchemaVersion string                          `json:"manifest_schema_version"`
	ResultSchemaVersion   string                          `json:"result_schema_version"`
	EvaluatedRevision     string                          `json:"evaluated_revision"`
	EvaluatorRepository   string                          `json:"evaluator_repository"`
	EvaluatorRevision     string                          `json:"evaluator_revision"`
	EvaluatedAt           time.Time                       `json:"evaluated_at"`
	SupportedPlatforms    []string                        `json:"supported_platforms"`
	PlatformSystems       map[string]PlatformSystemRecord `json:"platform_systems"`
	Capabilities          []string                        `json:"capabilities,omitempty"`
	Dependencies          []string                        `json:"dependencies,omitempty"`
	HealthEndpoint        string                          `json:"health_endpoint,omitempty"`
	ReadinessEndpoint     string                          `json:"readiness_endpoint,omitempty"`
	CompatibilityRequires []string                        `json:"compatibility_requires,omitempty"`
	ComputedConformance   string                          `json:"computed_conformance"`
	StableEligible        bool                            `json:"stable_eligible"`
	Blockers              []string                        `json:"blockers,omitempty"`
	Evidence              []string                        `json:"evidence,omitempty"`
	Authority             PlatformRecordAuthority         `json:"authority"`
}

type PlatformRecordRegistry struct {
	mu      sync.RWMutex
	path    string
	records map[string]PlatformRecord
}

func NewPlatformRecordRegistry() *PlatformRecordRegistry {
	return &PlatformRecordRegistry{records: map[string]PlatformRecord{}}
}

func normalizePlatformRecord(v PlatformRecord, evaluatedAt time.Time) (PlatformRecord, error) {
	v.ComponentID = strings.TrimSpace(v.ComponentID)
	v.ProductName = strings.TrimSpace(v.ProductName)
	v.ComponentType = strings.TrimSpace(v.ComponentType)
	v.Repository = strings.TrimSpace(v.Repository)
	v.Lifecycle = strings.TrimSpace(v.Lifecycle)
	v.Version = strings.TrimSpace(v.Version)
	v.ManifestSchemaVersion = strings.TrimSpace(v.ManifestSchemaVersion)
	v.ResultSchemaVersion = strings.TrimSpace(v.ResultSchemaVersion)
	v.EvaluatedRevision = strings.TrimSpace(v.EvaluatedRevision)
	v.EvaluatorRepository = strings.TrimSpace(v.EvaluatorRepository)
	v.EvaluatorRevision = strings.TrimSpace(v.EvaluatorRevision)
	v.HealthEndpoint = strings.TrimSpace(v.HealthEndpoint)
	v.ReadinessEndpoint = strings.TrimSpace(v.ReadinessEndpoint)
	v.ComputedConformance = strings.TrimSpace(v.ComputedConformance)
	v.Authority.DeclarationOwner = strings.TrimSpace(v.Authority.DeclarationOwner)
	v.Authority.AggregationRole = strings.TrimSpace(v.Authority.AggregationRole)
	evaluatedAt = evaluatedAt.UTC()

	if !platformComponentIDPattern.MatchString(v.ComponentID) {
		return PlatformRecord{}, errors.New("component_id must use lowercase kebab-case")
	}
	if v.ProductName == "" {
		return PlatformRecord{}, errors.New("product_name is required")
	}
	if v.ComponentType != "application" && v.ComponentType != "service" {
		return PlatformRecord{}, errors.New("component_type must be application or service")
	}
	if !platformRepositoryPattern.MatchString(v.Repository) {
		return PlatformRecord{}, errors.New("repository must identify a GoreeCloud repository")
	}
	if _, ok := platformLifecycleValues[v.Lifecycle]; !ok {
		return PlatformRecord{}, errors.New("invalid lifecycle")
	}
	if v.Version == "" {
		return PlatformRecord{}, errors.New("version is required")
	}
	if v.ManifestSchemaVersion != PlatformManifestSchemaVersion {
		return PlatformRecord{}, errors.New("unsupported Platform Contract manifest schema version")
	}
	if v.ResultSchemaVersion != PlatformResultSchemaVersion {
		return PlatformRecord{}, errors.New("unsupported platform conformance result schema version")
	}
	if !fullGitRevisionPattern.MatchString(v.EvaluatedRevision) {
		return PlatformRecord{}, errors.New("evaluated_revision must be an exact Git revision")
	}
	if v.EvaluatorRepository != PlatformEvaluatorRepository {
		return PlatformRecord{}, errors.New("evaluator_repository must identify the canonical Platform Contract repository")
	}
	if !fullGitRevisionPattern.MatchString(v.EvaluatorRevision) {
		return PlatformRecord{}, errors.New("evaluator_revision must be an exact Git revision")
	}
	if v.EvaluatedAt.IsZero() {
		return PlatformRecord{}, errors.New("evaluated_at is required")
	}
	v.EvaluatedAt = v.EvaluatedAt.UTC()
	if v.EvaluatedAt.After(evaluatedAt) {
		return PlatformRecord{}, errors.New("evaluated_at cannot be in the future")
	}
	if _, ok := platformConformanceValues[v.ComputedConformance]; !ok {
		return PlatformRecord{}, errors.New("invalid computed_conformance")
	}
	if v.Authority.DeclarationOwner != v.Repository {
		return PlatformRecord{}, errors.New("declaration_owner must match repository")
	}
	if v.Authority.AggregationRole != PlatformAggregationRoleReadOnly {
		return PlatformRecord{}, errors.New("aggregation_role must be read-only")
	}
	if v.Authority.AuthorityTransfer {
		return PlatformRecord{}, errors.New("authority_transfer must be false")
	}

	var err error
	if v.SupportedPlatforms, err = normalizePlatformStrings(v.SupportedPlatforms, true); err != nil {
		return PlatformRecord{}, errors.New("supported_platforms: " + err.Error())
	}
	if v.Capabilities, err = normalizePlatformStrings(v.Capabilities, false); err != nil {
		return PlatformRecord{}, errors.New("capabilities: " + err.Error())
	}
	if v.Dependencies, err = normalizePlatformStrings(v.Dependencies, false); err != nil {
		return PlatformRecord{}, errors.New("dependencies: " + err.Error())
	}
	if v.CompatibilityRequires, err = normalizePlatformStrings(v.CompatibilityRequires, false); err != nil {
		return PlatformRecord{}, errors.New("compatibility_requires: " + err.Error())
	}
	if v.Blockers, err = normalizePlatformStrings(v.Blockers, false); err != nil {
		return PlatformRecord{}, errors.New("blockers: " + err.Error())
	}
	if v.Evidence, err = normalizePlatformStrings(v.Evidence, false); err != nil {
		return PlatformRecord{}, errors.New("evidence: " + err.Error())
	}

	if len(v.PlatformSystems) != len(platformSystemKeys) {
		return PlatformRecord{}, errors.New("platform_systems must contain exactly seven Integral Platform Systems")
	}
	for _, key := range platformSystemKeys {
		system, ok := v.PlatformSystems[key]
		if !ok {
			return PlatformRecord{}, errors.New("platform_systems is missing " + key)
		}
		system.Result = strings.TrimSpace(system.Result)
		system.Version = strings.TrimSpace(system.Version)
		if _, ok := platformSystemResultValues[system.Result]; !ok {
			return PlatformRecord{}, errors.New("invalid platform system result for " + key)
		}
		system.Evidence, err = normalizePlatformStrings(system.Evidence, false)
		if err != nil {
			return PlatformRecord{}, errors.New("platform_systems." + key + ".evidence: " + err.Error())
		}
		if (system.Result == "applicable-conformant" || system.Result == "not-applicable-justified") && len(system.Evidence) == 0 {
			return PlatformRecord{}, errors.New("positive or not-applicable platform result requires evidence for " + key)
		}
		v.PlatformSystems[key] = system
	}
	for key := range v.PlatformSystems {
		known := false
		for _, expected := range platformSystemKeys {
			if key == expected {
				known = true
				break
			}
		}
		if !known {
			return PlatformRecord{}, errors.New("platform_systems contains unsupported key " + key)
		}
	}

	if v.StableEligible {
		if v.Lifecycle != "stable" || v.ComputedConformance != "conformant" || len(v.Blockers) != 0 {
			return PlatformRecord{}, errors.New("stable_eligible result is inconsistent with lifecycle, conformance, or blockers")
		}
		for _, key := range platformSystemKeys {
			result := v.PlatformSystems[key].Result
			if result != "applicable-conformant" && result != "not-applicable-justified" {
				return PlatformRecord{}, errors.New("stable_eligible result has unresolved platform system " + key)
			}
		}
	}

	return copyPlatformRecord(v), nil
}

func normalizePlatformStrings(values []string, required bool) ([]string, error) {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("values must not be empty")
		}
		if _, ok := seen[value]; ok {
			return nil, errors.New("duplicate value " + value)
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if required && len(out) == 0 {
		return nil, errors.New("at least one value is required")
	}
	return out, nil
}

func copyPlatformRecord(v PlatformRecord) PlatformRecord {
	v.SupportedPlatforms = append([]string(nil), v.SupportedPlatforms...)
	v.Capabilities = append([]string(nil), v.Capabilities...)
	v.Dependencies = append([]string(nil), v.Dependencies...)
	v.CompatibilityRequires = append([]string(nil), v.CompatibilityRequires...)
	v.Blockers = append([]string(nil), v.Blockers...)
	v.Evidence = append([]string(nil), v.Evidence...)
	copiedSystems := make(map[string]PlatformSystemRecord, len(v.PlatformSystems))
	for key, system := range v.PlatformSystems {
		system.Evidence = append([]string(nil), system.Evidence...)
		copiedSystems[key] = system
	}
	v.PlatformSystems = copiedSystems
	return v
}

func (r *PlatformRecordRegistry) Record(v PlatformRecord) (PlatformRecord, error) {
	normalized, err := normalizePlatformRecord(v, time.Now().UTC())
	if err != nil {
		return PlatformRecord{}, err
	}

	r.mu.Lock()
	previous, hadPrevious := r.records[normalized.ComponentID]
	r.records[normalized.ComponentID] = copyPlatformRecord(normalized)
	if err := r.persistLocked(); err != nil {
		if hadPrevious {
			r.records[normalized.ComponentID] = previous
		} else {
			delete(r.records, normalized.ComponentID)
		}
		r.mu.Unlock()
		return PlatformRecord{}, err
	}
	r.mu.Unlock()
	return copyPlatformRecord(normalized), nil
}

func (r *PlatformRecordRegistry) Get(componentID string) (PlatformRecord, bool) {
	componentID = strings.TrimSpace(componentID)
	r.mu.RLock()
	v, ok := r.records[componentID]
	r.mu.RUnlock()
	if !ok {
		return PlatformRecord{}, false
	}
	return copyPlatformRecord(v), true
}

func (r *PlatformRecordRegistry) List() []PlatformRecord {
	r.mu.RLock()
	out := make([]PlatformRecord, 0, len(r.records))
	for _, v := range r.records {
		out = append(out, copyPlatformRecord(v))
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ComponentID < out[j].ComponentID })
	return out
}
