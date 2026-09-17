package mesh

import (
	"context"
	"errors"
)

const replayProcessLockSuffix = ".replay.lock"

type replayProcessLease interface {
	Release() error
}

func (s *DurableSubscriberCheckpoints) acquireReplayOwnership(ctx context.Context) (replayProcessLease, error) {
	if err := s.acquireReplay(ctx); err != nil {
		return nil, err
	}
	lease, err := acquireReplayProcessLock(ctx, s.path+replayProcessLockSuffix)
	if err != nil {
		s.releaseReplay()
		return nil, err
	}
	return lease, nil
}

func (s *DurableSubscriberCheckpoints) releaseReplayOwnership(lease replayProcessLease) error {
	defer s.releaseReplay()
	if lease == nil {
		return errors.New("replay process ownership lease is required")
	}
	return lease.Release()
}
