//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package mesh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const replayProcessLockRetry = 25 * time.Millisecond

type unixReplayProcessLock struct {
	file *os.File
}

func acquireReplayProcessLock(ctx context.Context, path string) (replayProcessLease, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("prepare replay ownership directory: %w", err)
	}

	fd, err := syscall.Open(
		path,
		syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC|syscall.O_NOFOLLOW,
		0o600,
	)
	if err != nil {
		return nil, fmt.Errorf("open replay ownership lock: %w", err)
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, errors.New("open replay ownership lock: invalid file descriptor")
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("inspect replay ownership lock: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("replay ownership lock must be a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		_ = file.Close()
		return nil, fmt.Errorf("replay ownership lock permissions must be 0600, got %04o", info.Mode().Perm())
	}

	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &unixReplayProcessLock{file: file}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = file.Close()
			return nil, fmt.Errorf("acquire replay ownership lock: %w", err)
		}

		timer := time.NewTimer(replayProcessLockRetry)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			_ = file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (l *unixReplayProcessLock) Release() error {
	if l == nil || l.file == nil {
		return errors.New("replay ownership lock is not held")
	}
	fd := int(l.file.Fd())
	unlockErr := syscall.Flock(fd, syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	return errors.Join(unlockErr, closeErr)
}
