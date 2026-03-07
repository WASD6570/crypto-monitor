package ingestion

import "testing"

func TestOrderBookSequencerAcceptsSnapshotThenSequentialDelta(t *testing.T) {
	var sequencer OrderBookSequencer

	snapshotResult, err := sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateSnapshot, Sequence: 500})
	if err != nil {
		t.Fatalf("apply snapshot: %v", err)
	}
	if snapshotResult.Action != SequenceAcceptedSnapshot {
		t.Fatalf("snapshot action = %q, want %q", snapshotResult.Action, SequenceAcceptedSnapshot)
	}

	deltaResult, err := sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateDelta, Sequence: 501})
	if err != nil {
		t.Fatalf("apply delta: %v", err)
	}
	if deltaResult.Action != SequenceAcceptedDelta {
		t.Fatalf("delta action = %q, want %q", deltaResult.Action, SequenceAcceptedDelta)
	}
	if sequencer.LastSequence() != 501 {
		t.Fatalf("last sequence = %d, want %d", sequencer.LastSequence(), 501)
	}
	if !sequencer.Ready() {
		t.Fatal("expected sequencer to remain ready after in-order delta")
	}
}

func TestOrderBookSequencerIgnoresStaleDelta(t *testing.T) {
	var sequencer OrderBookSequencer
	_, _ = sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateSnapshot, Sequence: 500})
	_, _ = sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateDelta, Sequence: 501})

	result, err := sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateDelta, Sequence: 501})
	if err != nil {
		t.Fatalf("apply stale delta: %v", err)
	}
	if result.Action != SequenceIgnoredStale {
		t.Fatalf("action = %q, want %q", result.Action, SequenceIgnoredStale)
	}
	if result.Reason != "stale-or-out-of-order-delta" {
		t.Fatalf("reason = %q, want stale/out-of-order reason", result.Reason)
	}
	if sequencer.ResyncRequired() {
		t.Fatal("stale delta should not force resync")
	}
}

func TestOrderBookSequencerRequiresResyncOnGap(t *testing.T) {
	var sequencer OrderBookSequencer
	_, _ = sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateSnapshot, Sequence: 900})

	result, err := sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateDelta, Sequence: 902})
	if err != nil {
		t.Fatalf("apply gap delta: %v", err)
	}
	if result.Action != SequenceRequiresResync {
		t.Fatalf("action = %q, want %q", result.Action, SequenceRequiresResync)
	}
	if result.Reason != "sequence-gap-detected" {
		t.Fatalf("reason = %q, want %q", result.Reason, "sequence-gap-detected")
	}
	if !sequencer.ResyncRequired() {
		t.Fatal("expected resync to be required after gap")
	}
	if sequencer.Ready() {
		t.Fatal("sequencer should not be ready after a gap")
	}
}

func TestOrderBookSequencerSnapshotClearsResyncState(t *testing.T) {
	var sequencer OrderBookSequencer
	_, _ = sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateSnapshot, Sequence: 900})
	_, _ = sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateDelta, Sequence: 902})

	result, err := sequencer.Apply(SequencedBookUpdate{Kind: BookUpdateSnapshot, Sequence: 950})
	if err != nil {
		t.Fatalf("apply resync snapshot: %v", err)
	}
	if result.Action != SequenceAcceptedSnapshot {
		t.Fatalf("action = %q, want %q", result.Action, SequenceAcceptedSnapshot)
	}
	if sequencer.ResyncRequired() {
		t.Fatal("snapshot should clear resync-required state")
	}
	if !sequencer.Ready() {
		t.Fatal("sequencer should be ready after replacement snapshot")
	}
}
