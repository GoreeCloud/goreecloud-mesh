//go:build !(linux || darwin || freebsd || openbsd || netbsd || dragonfly)

package mesh

import (
	"context"
	"errors"
)

func acquireReplayProcessLock(context.Context, string) (replayProcessLease, error) {
	return nil, errors.New("cross-process durable replay ownership is unsupported on this platform")
}
