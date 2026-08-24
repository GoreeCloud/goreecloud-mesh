package governance

import (
	"errors"
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

type RecoveryCapability string

const (
	RecoveryExport  RecoveryCapability = "export"
	RecoveryBackup  RecoveryCapability = "backup"
	RecoveryRestore RecoveryCapability = "restore"
	RecoveryVerify  RecoveryCapability = "verify"
)

type RecoveryEvidence struct {
	Capability RecoveryCapability `json:"capability"`
	State      string             `json:"state"`
	Source     string             `json:"source"`
	Revision   string             `json:"revision,omitempty"`
	ObservedAt time.Time          `json:"observed_at"`
}

func (e RecoveryEvidence) Validate() error {
	if e.Capability != RecoveryExport && e.Capability != RecoveryBackup && e.Capability != RecoveryRestore && e.Capability != RecoveryVerify {
		return errors.New("invalid recovery capability")
	}
	if strings.TrimSpace(e.Source) == "" || strings.TrimSpace(e.State) == "" {
		return errors.New("source and state are required")
	}
	if e.ObservedAt.IsZero() {
		return errors.New("observed time is required")
	}
	return nil
}

func RecoveryReady(evidence []RecoveryEvidence) bool {
	required := map[RecoveryCapability]bool{
		RecoveryExport: false,
		RecoveryBackup: false,
		RecoveryRestore: false,
		RecoveryVerify: false,
	}
	for _, item := range evidence {
		if item.Validate() != nil || item.State != "validated" {
			continue
		}
		if _, ok := required[item.Capability]; ok {
			required[item.Capability] = true
		}
	}
	for _, ok := range required {
		if !ok {
			return false
		}
	}
	return true
}

func RequiredRecoveryCapabilities() []RecoveryCapability {
	out := []RecoveryCapability{RecoveryExport, RecoveryBackup, RecoveryRestore, RecoveryVerify}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
