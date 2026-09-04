package platformregistry

import (
	"errors"
	"reflect"
)

// ValidateUpdate prevents an accepted component record from moving backwards in
// producer observation time or canonical-evaluator time. Exact retries are
// idempotent. Producer-owned facts may change only with a later observed_at;
// canonical conformance may change only with a later evaluated_at.
func ValidateUpdate(current, incoming Record) error {
	current = normalized(current)
	incoming = normalized(incoming)

	if incoming.ObservedAt.Before(current.ObservedAt) {
		return errors.New("platform record observed_at cannot move backwards")
	}
	if incoming.Conformance.EvaluatedAt.Before(current.Conformance.EvaluatedAt) {
		return errors.New("platform record conformance evaluated_at cannot move backwards")
	}

	currentProducer := current
	incomingProducer := incoming
	currentProducer.Conformance = Conformance{}
	incomingProducer.Conformance = Conformance{}
	if !reflect.DeepEqual(currentProducer, incomingProducer) && !incoming.ObservedAt.After(current.ObservedAt) {
		return errors.New("producer-owned platform state changes require a newer observed_at")
	}

	if !reflect.DeepEqual(current.Conformance, incoming.Conformance) && !incoming.Conformance.EvaluatedAt.After(current.Conformance.EvaluatedAt) {
		return errors.New("canonical conformance changes require a newer evaluated_at")
	}

	return nil
}
