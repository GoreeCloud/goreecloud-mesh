package mesh

import (
	"context"
	"errors"
)

// SubscriberEventHandler processes one replayed durable event. Returning an
// error stops replay before that event is acknowledged, so a later replay may
// present the same event again. This deliberately does not claim exactly-once
// processing or delivery.
type SubscriberEventHandler func(context.Context, DurableEventRecord) error

// ReplaySubscriber replays a bounded batch after a subscriber's durable
// checkpoint and persists progress only after each handler call succeeds.
//
// The returned processed count includes only events whose handler completed and
// whose durable checkpoint was successfully advanced. A handler or checkpoint
// persistence failure stops the batch immediately.
func ReplaySubscriber(
	ctx context.Context,
	subscriberID string,
	limit int,
	journal *DurableEventJournal,
	checkpoints *DurableSubscriberCheckpoints,
	handler SubscriberEventHandler,
) (processed int, checkpoint uint64, err error) {
	if err := validateSubscriberID(subscriberID); err != nil {
		return 0, 0, err
	}
	if journal == nil {
		return 0, 0, errors.New("durable event journal is required")
	}
	if checkpoints == nil {
		return 0, 0, errors.New("subscriber checkpoint store is required")
	}
	if handler == nil {
		return 0, 0, errors.New("subscriber event handler is required")
	}
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	checkpoint, found, err := checkpoints.Checkpoint(subscriberID)
	if err != nil {
		return 0, 0, err
	}
	if !found {
		checkpoint = 0
	}

	records, err := journal.ReplayAfter(checkpoint, limit)
	if err != nil {
		return 0, checkpoint, err
	}
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return processed, checkpoint, err
		}
		if err := handler(ctx, cloneDurableEventRecord(record)); err != nil {
			return processed, checkpoint, err
		}
		if err := checkpoints.Acknowledge(subscriberID, record.Offset, journal); err != nil {
			return processed, checkpoint, err
		}
		checkpoint = record.Offset
		processed++
	}
	return processed, checkpoint, nil
}
