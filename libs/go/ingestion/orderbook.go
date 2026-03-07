package ingestion

import "fmt"

type BookUpdateKind string

const (
	BookUpdateSnapshot BookUpdateKind = "snapshot"
	BookUpdateDelta    BookUpdateKind = "delta"
)

type SequenceAction string

const (
	SequenceAcceptedSnapshot SequenceAction = "accepted-snapshot"
	SequenceAcceptedDelta    SequenceAction = "accepted-delta"
	SequenceIgnoredStale     SequenceAction = "ignored-stale"
	SequenceRequiresResync   SequenceAction = "requires-resync"
)

type SequencedBookUpdate struct {
	Kind     BookUpdateKind
	Sequence int64
}

type SequenceResult struct {
	Action         SequenceAction
	LastSequence   int64
	ExpectedNext   int64
	ResyncRequired bool
	Reason         string
}

type OrderBookSequencer struct {
	bootstrapped   bool
	lastSequence   int64
	resyncRequired bool
}

func (s *OrderBookSequencer) Apply(update SequencedBookUpdate) (SequenceResult, error) {
	if update.Kind == "" {
		return SequenceResult{}, fmt.Errorf("update kind is required")
	}
	if update.Sequence <= 0 {
		return SequenceResult{}, fmt.Errorf("sequence must be positive")
	}

	switch update.Kind {
	case BookUpdateSnapshot:
		s.bootstrapped = true
		s.lastSequence = update.Sequence
		s.resyncRequired = false
		return SequenceResult{
			Action:         SequenceAcceptedSnapshot,
			LastSequence:   s.lastSequence,
			ExpectedNext:   s.lastSequence + 1,
			ResyncRequired: false,
		}, nil
	case BookUpdateDelta:
		if !s.bootstrapped || s.resyncRequired {
			s.resyncRequired = true
			return SequenceResult{
				Action:         SequenceRequiresResync,
				LastSequence:   s.lastSequence,
				ExpectedNext:   s.lastSequence + 1,
				ResyncRequired: true,
				Reason:         "delta-received-without-ready-snapshot",
			}, nil
		}

		expected := s.lastSequence + 1
		if update.Sequence == expected {
			s.lastSequence = update.Sequence
			return SequenceResult{
				Action:         SequenceAcceptedDelta,
				LastSequence:   s.lastSequence,
				ExpectedNext:   s.lastSequence + 1,
				ResyncRequired: false,
			}, nil
		}
		if update.Sequence <= s.lastSequence {
			return SequenceResult{
				Action:         SequenceIgnoredStale,
				LastSequence:   s.lastSequence,
				ExpectedNext:   expected,
				ResyncRequired: false,
				Reason:         "stale-or-out-of-order-delta",
			}, nil
		}

		s.resyncRequired = true
		return SequenceResult{
			Action:         SequenceRequiresResync,
			LastSequence:   s.lastSequence,
			ExpectedNext:   expected,
			ResyncRequired: true,
			Reason:         "sequence-gap-detected",
		}, nil
	default:
		return SequenceResult{}, fmt.Errorf("unsupported update kind %q", update.Kind)
	}
}

func (s *OrderBookSequencer) LastSequence() int64 {
	return s.lastSequence
}

func (s *OrderBookSequencer) Ready() bool {
	return s.bootstrapped && !s.resyncRequired
}

func (s *OrderBookSequencer) ResyncRequired() bool {
	return s.resyncRequired
}
