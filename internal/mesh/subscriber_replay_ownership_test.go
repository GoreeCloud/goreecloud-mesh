//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package mesh

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReplayOwnershipSerializesIndependentCheckpointStores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subscriber-checkpoints.json")
	first, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatalf("NewDurableSubscriberCheckpoints(first): %v", err)
	}
	second, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatalf("NewDurableSubscriberCheckpoints(second): %v", err)
	}

	firstLease, err := first.acquireReplayOwnership(context.Background())
	if err != nil {
		t.Fatalf("first replay ownership: %v", err)
	}

	blockedCtx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	if _, err := second.acquireReplayOwnership(blockedCtx); !errors.Is(err, context.DeadlineExceeded) {
		_ = first.releaseReplayOwnership(firstLease)
		t.Fatalf("second store acquired ownership while first held kernel lock: %v", err)
	}

	if err := first.releaseReplayOwnership(firstLease); err != nil {
		t.Fatalf("release first replay ownership: %v", err)
	}

	secondCtx, secondCancel := context.WithTimeout(context.Background(), time.Second)
	defer secondCancel()
	secondLease, err := second.acquireReplayOwnership(secondCtx)
	if err != nil {
		t.Fatalf("second replay ownership after release: %v", err)
	}
	if err := second.releaseReplayOwnership(secondLease); err != nil {
		t.Fatalf("release second replay ownership: %v", err)
	}

	info, err := os.Lstat(path + replayProcessLockSuffix)
	if err != nil {
		t.Fatalf("persistent replay lock file missing: %v", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("replay lock mode = %v, want private regular 0600 file", info.Mode())
	}
}

func TestReplayOwnershipRejectsSymlinkBackedLockPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subscriber-checkpoints.json")
	store, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatalf("NewDurableSubscriberCheckpoints: %v", err)
	}

	target := filepath.Join(dir, "unexpected-lock-target")
	if err := os.WriteFile(target, []byte("not a Mesh replay lock\n"), 0o600); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(target, path+replayProcessLockSuffix); err != nil {
		t.Fatalf("create replay lock symlink: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := store.acquireReplayOwnership(ctx); err == nil {
		t.Fatal("symlink-backed replay ownership lock must fail closed")
	}

	lease, err := store.acquireReplay(context.Background())
	if err != nil {
		t.Fatalf("failed process-lock acquisition must release in-process ownership: %v", err)
	}
	store.releaseReplay()
}

func TestReplayOwnershipRejectsNonPrivateExistingLockFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subscriber-checkpoints.json")
	store, err := NewDurableSubscriberCheckpoints(path)
	if err != nil {
		t.Fatalf("NewDurableSubscriberCheckpoints: %v", err)
	}
	lockPath := path + replayProcessLockSuffix
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatalf("write non-private replay lock: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := store.acquireReplayOwnership(ctx); err == nil {
		t.Fatal("non-private replay ownership lock must fail closed")
	}
}
