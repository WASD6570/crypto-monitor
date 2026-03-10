package venuebinance

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSpotDepthBootstrapOwnerBuffersDeltasAndSurfacesPendingState(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthBootstrapOwner(runtime, stubSpotDepthSnapshotFetcher{})
	if err != nil {
		t.Fatalf("new bootstrap owner: %v", err)
	}

	if err := owner.BufferDelta(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798520100).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798520100,"s":"BTCUSDT","U":699,"u":700,"b":[["64020.00","1.10"]],"a":[["64020.50","0.80"]]}`),
	}); err != nil {
		t.Fatalf("buffer first delta: %v", err)
	}
	if err := owner.BufferDelta(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798520200).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798520200,"s":"BTCUSDT","U":701,"u":702,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`),
	}); err != nil {
		t.Fatalf("buffer second delta: %v", err)
	}

	status := owner.Status()
	if status.State != SpotDepthBootstrapBuffering {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthBootstrapBuffering)
	}
	if status.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", status.SourceSymbol, "BTCUSDT")
	}
	if status.BufferedDeltaCount != 2 {
		t.Fatalf("buffered delta count = %d, want %d", status.BufferedDeltaCount, 2)
	}
	if status.Synchronized {
		t.Fatal("did not expect synchronized status before snapshot bootstrap")
	}
}

func TestSpotDepthBootstrapOwnerRejectsMalformedSnapshotState(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthBootstrapOwner(runtime, stubSpotDepthSnapshotFetcher{
		response: SpotDepthSnapshotResponse{
			RecvTime: time.UnixMilli(1772798520300).UTC(),
			Payload:  []byte(`{"lastUpdateId":0,"bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`),
		},
	})
	if err != nil {
		t.Fatalf("new bootstrap owner: %v", err)
	}
	if err := owner.BufferDelta(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798520200).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798520200,"s":"BTCUSDT","U":701,"u":701,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`),
	}); err != nil {
		t.Fatalf("buffer delta: %v", err)
	}

	if _, err := owner.Synchronize(context.Background()); err == nil {
		t.Fatal("expected malformed snapshot to fail")
	}
	status := owner.Status()
	if status.State != SpotDepthBootstrapFailed {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthBootstrapFailed)
	}
	if status.FailureReason != SpotDepthBootstrapFailureSnapshotParse {
		t.Fatalf("failure reason = %q, want %q", status.FailureReason, SpotDepthBootstrapFailureSnapshotParse)
	}
}

func TestSpotDepthBootstrapOwnerSynchronizeReturnsSnapshotAndAlignedDeltas(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthBootstrapOwner(runtime, stubSpotDepthSnapshotFetcher{
		response: SpotDepthSnapshotResponse{
			RecvTime: time.UnixMilli(1772798520250).UTC(),
			Payload:  []byte(`{"lastUpdateId":700,"bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`),
		},
	})
	if err != nil {
		t.Fatalf("new bootstrap owner: %v", err)
	}
	for _, frame := range []SpotRawFrame{
		{
			RecvTime: time.UnixMilli(1772798520100).UTC(),
			Payload:  []byte(`{"e":"depthUpdate","E":1772798520100,"s":"BTCUSDT","U":698,"u":700,"b":[["64019.90","1.15"]],"a":[["64020.60","0.85"]]}`),
		},
		{
			RecvTime: time.UnixMilli(1772798520200).UTC(),
			Payload:  []byte(`{"e":"depthUpdate","E":1772798520200,"s":"BTCUSDT","U":700,"u":701,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`),
		},
		{
			RecvTime: time.UnixMilli(1772798520300).UTC(),
			Payload:  []byte(`{"e":"depthUpdate","E":1772798520300,"s":"BTCUSDT","U":702,"u":702,"b":[["64020.20","1.00"]],"a":[["64020.30","0.70"]]}`),
		},
	} {
		if err := owner.BufferDelta(frame); err != nil {
			t.Fatalf("buffer delta: %v", err)
		}
	}

	sync, err := owner.Synchronize(context.Background())
	if err != nil {
		t.Fatalf("synchronize bootstrap owner: %v", err)
	}
	if sync.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q, want %q", sync.SourceSymbol, "BTCUSDT")
	}
	if sync.Snapshot.Message.Sequence != 700 {
		t.Fatalf("snapshot sequence = %d, want %d", sync.Snapshot.Message.Sequence, 700)
	}
	if len(sync.Deltas) != 2 {
		t.Fatalf("aligned delta count = %d, want %d", len(sync.Deltas), 2)
	}
	if sync.Deltas[0].FirstSequence != 700 || sync.Deltas[0].FinalSequence != 701 {
		t.Fatalf("first aligned window = %d-%d, want 700-701", sync.Deltas[0].FirstSequence, sync.Deltas[0].FinalSequence)
	}
	if sync.Deltas[1].FirstSequence != 702 || sync.Deltas[1].FinalSequence != 702 {
		t.Fatalf("second aligned window = %d-%d, want 702-702", sync.Deltas[1].FirstSequence, sync.Deltas[1].FinalSequence)
	}
	status := owner.Status()
	if status.State != SpotDepthBootstrapSynchronized {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthBootstrapSynchronized)
	}
	if !status.Synchronized {
		t.Fatal("expected synchronized status after successful bootstrap")
	}
	if !status.SnapshotRequestedAt.Equal(time.UnixMilli(1772798520250).UTC()) {
		t.Fatalf("snapshot requested at = %s, want %s", status.SnapshotRequestedAt, time.UnixMilli(1772798520250).UTC())
	}
}

func TestSpotDepthBootstrapOwnerFailsWhenNoBufferedDeltaBridgesSnapshot(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthBootstrapOwner(runtime, stubSpotDepthSnapshotFetcher{
		response: SpotDepthSnapshotResponse{
			RecvTime: time.UnixMilli(1772798520250).UTC(),
			Payload:  []byte(`{"lastUpdateId":700,"bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`),
		},
	})
	if err != nil {
		t.Fatalf("new bootstrap owner: %v", err)
	}
	for _, frame := range []SpotRawFrame{
		{
			RecvTime: time.UnixMilli(1772798520100).UTC(),
			Payload:  []byte(`{"e":"depthUpdate","E":1772798520100,"s":"BTCUSDT","U":698,"u":699,"b":[["64019.90","1.15"]],"a":[["64020.60","0.85"]]}`),
		},
		{
			RecvTime: time.UnixMilli(1772798520200).UTC(),
			Payload:  []byte(`{"e":"depthUpdate","E":1772798520200,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`),
		},
	} {
		if err := owner.BufferDelta(frame); err != nil {
			t.Fatalf("buffer delta: %v", err)
		}
	}

	if _, err := owner.Synchronize(context.Background()); err == nil {
		t.Fatal("expected synchronize without bridging delta to fail")
	}
	status := owner.Status()
	if status.State != SpotDepthBootstrapFailed {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthBootstrapFailed)
	}
	if status.FailureReason != SpotDepthBootstrapFailureBridgeMissing {
		t.Fatalf("failure reason = %q, want %q", status.FailureReason, SpotDepthBootstrapFailureBridgeMissing)
	}
}

type stubSpotDepthSnapshotFetcher struct {
	response SpotDepthSnapshotResponse
	err      error
}

func (f stubSpotDepthSnapshotFetcher) FetchSpotDepthSnapshot(_ context.Context, sourceSymbol string) (SpotDepthSnapshotResponse, error) {
	if sourceSymbol == "" {
		return SpotDepthSnapshotResponse{}, fmt.Errorf("source symbol is required")
	}
	if f.err != nil {
		return SpotDepthSnapshotResponse{}, f.err
	}
	return f.response, nil
}

func newBinanceRuntime(t *testing.T) *Runtime {
	t.Helper()
	config := loadBinanceRuntimeConfig(t)
	runtime, err := NewRuntime(config)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	return runtime
}
