package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

const (
	SubscriberCheckpointSchemaV1      = "goreecloud.mesh.subscriber-checkpoints.v1"
	maxSubscriberCheckpointStateBytes = 4 << 20
	maxSubscriberCheckpoints          = 10000
	maxSubscriberIDLength             = 160
)

var ErrSubscriberCheckpointRegression = errors.New("subscriber checkpoint cannot move backwards")

type subscriberCheckpointState struct {
	Schema      string            `json:"schema"`
	Checkpoints map[string]uint64 `json:"checkpoints"`
}

// DurableSubscriberCheckpoints persists per-subscriber durable replay positions.
// It does not imply event delivery, retry, acknowledgement transport, or an
// at-least-once/exactly-once guarantee; it is only the durable checkpoint state
// primitive used by future delivery paths.
type DurableSubscriberCheckpoints struct {
	mu          sync.RWMutex
	path        string
	checkpoints map[string]uint64
	replayGate  chan struct{}
}

func NewDurableSubscriberCheckpoints(path string) (*DurableSubscriberCheckpoints, error) {
	if path == "" {
		return nil, errors.New("subscriber checkpoint path is required")
	}
	store := &DurableSubscriberCheckpoints{
		path:        path,
		checkpoints: map[string]uint64{},
		replayGate:  make(chan struct{}, 1),
	}
	store.replayGate <- struct{}{}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *DurableSubscriberCheckpoints) acquireReplay(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.replayGate:
		return nil
	}
}

func (s *DurableSubscriberCheckpoints) releaseReplay() {
	s.replayGate <- struct{}{}
}

func (s *DurableSubscriberCheckpoints) Acknowledge(subscriberID string, checkpoint uint64, journal *DurableEventJournal) error {
	if err := validateSubscriberID(subscriberID); err != nil {
		return err
	}
	if journal == nil {
		return errors.New("durable event journal is required")
	}

	first, last, hasEntries := journal.Bounds()
	if checkpoint > last {
		return ErrEventCheckpointAhead
	}
	if hasEntries && first > 0 && checkpoint < first-1 {
		return ErrEventCheckpointTooOld
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if previous, ok := s.checkpoints[subscriberID]; ok {
		if checkpoint < previous {
			return ErrSubscriberCheckpointRegression
		}
		if checkpoint == previous {
			return nil
		}
	} else if len(s.checkpoints) >= maxSubscriberCheckpoints {
		return fmt.Errorf("subscriber checkpoint store exceeds %d subscribers", maxSubscriberCheckpoints)
	}

	candidate := make(map[string]uint64, len(s.checkpoints)+1)
	for id, value := range s.checkpoints {
		candidate[id] = value
	}
	candidate[subscriberID] = checkpoint
	if err := s.persist(subscriberCheckpointState{
		Schema:      SubscriberCheckpointSchemaV1,
		Checkpoints: candidate,
	}); err != nil {
		return err
	}
	s.checkpoints = candidate
	return nil
}

func (s *DurableSubscriberCheckpoints) Checkpoint(subscriberID string) (uint64, bool, error) {
	if err := validateSubscriberID(subscriberID); err != nil {
		return 0, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	checkpoint, ok := s.checkpoints[subscriberID]
	return checkpoint, ok, nil
}

func validateSubscriberID(value string) error {
	if value == "" || value != strings.TrimSpace(value) {
		return errors.New("subscriber ID must be a non-empty canonical identifier")
	}
	if len(value) > maxSubscriberIDLength {
		return fmt.Errorf("subscriber ID exceeds %d characters", maxSubscriberIDLength)
	}
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return errors.New("subscriber ID contains control characters")
	}
	return nil
}

func (s *DurableSubscriberCheckpoints) load() error {
	info, err := os.Lstat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("subscriber checkpoint state must be a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		return fmt.Errorf("subscriber checkpoint permissions must be 0600, got %04o", info.Mode().Perm())
	}
	if info.Size() <= 0 {
		return errors.New("subscriber checkpoint state is empty or corrupt")
	}
	if info.Size() > maxSubscriberCheckpointStateBytes {
		return fmt.Errorf("subscriber checkpoint state exceeds %d-byte limit", maxSubscriberCheckpointStateBytes)
	}

	file, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(info, openedInfo) {
		return errors.New("subscriber checkpoint state changed while being opened")
	}
	body, err := io.ReadAll(io.LimitReader(file, maxSubscriberCheckpointStateBytes+1))
	if err != nil {
		return err
	}
	if len(body) == 0 || len(body) > maxSubscriberCheckpointStateBytes {
		return errors.New("subscriber checkpoint state is empty, oversized, or corrupt")
	}

	var state subscriberCheckpointState
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return fmt.Errorf("decode subscriber checkpoint state: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("decode subscriber checkpoint state: trailing JSON value is not allowed")
		}
		return fmt.Errorf("decode subscriber checkpoint trailing data: %w", err)
	}
	if state.Schema != SubscriberCheckpointSchemaV1 {
		return fmt.Errorf("unsupported subscriber checkpoint schema %q", state.Schema)
	}
	if state.Checkpoints == nil {
		return errors.New("subscriber checkpoint map is required")
	}
	if len(state.Checkpoints) > maxSubscriberCheckpoints {
		return fmt.Errorf("subscriber checkpoint store exceeds %d subscribers", maxSubscriberCheckpoints)
	}
	for id := range state.Checkpoints {
		if err := validateSubscriberID(id); err != nil {
			return fmt.Errorf("invalid persisted subscriber ID: %w", err)
		}
	}
	s.checkpoints = state.Checkpoints
	return nil
}

func (s *DurableSubscriberCheckpoints) persist(state subscriberCheckpointState) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".subscriber-checkpoints-*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(tmp)
	}
	if err := file.Chmod(0o600); err != nil {
		cleanup()
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		cleanup()
		return err
	}
	if stat, err := file.Stat(); err != nil {
		cleanup()
		return err
	} else if stat.Size() > maxSubscriberCheckpointStateBytes {
		cleanup()
		return fmt.Errorf("subscriber checkpoint state exceeds %d-byte limit", maxSubscriberCheckpointStateBytes)
	}
	if err := file.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync subscriber checkpoint directory: %w", err)
	}
	return nil
}
