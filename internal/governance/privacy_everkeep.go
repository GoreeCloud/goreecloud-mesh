package governance

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

type DataClass string

const (
	DataCoordination DataClass = "coordination-metadata"
	DataOperational  DataClass = "operational-evidence"
	DataSensitive    DataClass = "sensitive-metadata"
)

type Retention string

const (
	RetentionEphemeral Retention = "ephemeral"
	RetentionBounded   Retention = "bounded"
	RetentionPreserved Retention = "preserved"
)

type PrivacyRecord struct {
	Name            string        `json:"name"`
	Class           DataClass     `json:"class"`
	Purpose         string        `json:"purpose"`
	Retention       Retention     `json:"retention"`
	MaxAge          time.Duration `json:"max_age,omitempty"`
	ContainsPayload bool          `json:"contains_payload"`
	Exportable      bool          `json:"exportable"`
}

func (r PrivacyRecord) Validate() error {
	if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Purpose) == "" {
		return errors.New("name and purpose are required")
	}
	if r.Class != DataCoordination && r.Class != DataOperational && r.Class != DataSensitive {
		return errors.New("invalid data class")
	}
	if r.Retention != RetentionEphemeral && r.Retention != RetentionBounded && r.Retention != RetentionPreserved {
		return errors.New("invalid retention class")
	}
	if r.ContainsPayload {
		return errors.New("application payload content is not permitted in Mesh governance records")
	}
	if r.Retention == RetentionBounded && r.MaxAge <= 0 {
		return errors.New("bounded retention requires a positive max age")
	}
	if r.Retention != RetentionBounded && r.MaxAge != 0 {
		return errors.New("max age is only valid for bounded retention")
	}
	return nil
}

type RecoveryDimension string

const (
	RecoveryBackupCoverage   RecoveryDimension = "backup_coverage"
	RecoveryRestoreCapability RecoveryDimension = "restore_capability"
	RecoveryPortability       RecoveryDimension = "portability"
	RecoveryDocumentation     RecoveryDimension = "documentation"
	RecoveryProvenance        RecoveryDimension = "provenance"
)

var exactRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type RecoveryEvidence struct {
	Dimension  RecoveryDimension `json:"dimension"`
	State      string            `json:"state"`
	Source     string            `json:"source"`
	Revision   string            `json:"revision"`
	ObservedAt time.Time         `json:"observed_at"`
	ValidUntil time.Time         `json:"valid_until"`
}

func (e RecoveryEvidence) Validate(at time.Time) error {
	if !isRequiredRecoveryDimension(e.Dimension) {
		return errors.New("invalid recovery dimension")
	}
	if e.State != "validated" && e.State != "degraded" && e.State != "unknown" {
		return errors.New("invalid recovery evidence state")
	}
	if strings.TrimSpace(e.Source) == "" {
		return errors.New("source is required")
	}
	if !exactRevisionPattern.MatchString(e.Revision) {
		return errors.New("exact lowercase 40-character source revision is required")
	}
	if e.ObservedAt.IsZero() || e.ValidUntil.IsZero() {
		return errors.New("observed and validity times are required")
	}
	if e.ObservedAt.After(at) {
		return errors.New("future-dated recovery evidence is invalid")
	}
	if !e.ValidUntil.After(e.ObservedAt) {
		return errors.New("validity must extend beyond observation time")
	}
	return nil
}

func RecoveryReady(evidence []RecoveryEvidence, at time.Time) bool {
	required := map[RecoveryDimension]bool{}
	for _, dimension := range RequiredRecoveryDimensions() {
		required[dimension] = false
	}
	for _, item := range evidence {
		if item.Validate(at) != nil || item.State != "validated" || !item.ValidUntil.After(at) {
			continue
		}
		required[item.Dimension] = true
	}
	for _, ok := range required {
		if !ok {
			return false
		}
	}
	return true
}

func RequiredRecoveryDimensions() []RecoveryDimension {
	out := []RecoveryDimension{
		RecoveryBackupCoverage,
		RecoveryRestoreCapability,
		RecoveryPortability,
		RecoveryDocumentation,
		RecoveryProvenance,
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func isRequiredRecoveryDimension(dimension RecoveryDimension) bool {
	for _, required := range RequiredRecoveryDimensions() {
		if dimension == required {
			return true
		}
	}
	return false
}
