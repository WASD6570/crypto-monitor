package venuebinance

import (
	"context"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestSpotDepthRecoveryOwnerMarksSequenceGapAndBlocksOnCooldown(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798520250, 701, 1772798520300)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}

	if err := owner.MarkSequenceGap(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798520600).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798520600,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.40","0.95"]],"a":[["64020.70","0.80"]]}`),
	}); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	status := owner.Status()
	if status.State != SpotDepthRecoveryCooldownBlocked {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthRecoveryCooldownBlocked)
	}
	if !status.SequenceGapDetected {
		t.Fatal("expected sequence gap to remain visible")
	}
	if status.RemainingCooldown != 650*time.Millisecond {
		t.Fatalf("remaining cooldown = %s, want %s", status.RemainingCooldown, 650*time.Millisecond)
	}
	if status.Synchronized {
		t.Fatal("did not expect synchronized state after gap")
	}
	if _, err := owner.Recover(context.Background(), time.UnixMilli(1772798520600).UTC()); err == nil {
		t.Fatal("expected cooldown-blocked recovery to fail")
	}
}

func TestSpotDepthRecoveryOwnerBlocksRecoveryAtRateLimit(t *testing.T) {
	runtime := newBinanceRuntime(t)
	runtime.config.SnapshotRecoveryPerMinuteLimit = 1
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798466000, 701, 1772798466100)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	if err := owner.MarkSequenceGap(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798521000).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798521000,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.40","0.95"]],"a":[["64020.70","0.80"]]}`),
	}); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	status := owner.Status()
	if status.State != SpotDepthRecoveryRateLimitBlocked {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthRecoveryRateLimitBlocked)
	}
	if status.RetryAfter != 5*time.Second {
		t.Fatalf("retry after = %s, want %s", status.RetryAfter, 5*time.Second)
	}
	feedStatus, err := owner.HealthStatus(time.UnixMilli(1772798521000).UTC(), ingestion.ConnectionConnected, 0, 0)
	if err != nil {
		t.Fatalf("health status: %v", err)
	}
	if !containsDepthRecoveryReason(feedStatus.Reasons, ingestion.ReasonRateLimit) {
		t.Fatalf("reasons = %v, want %q", feedStatus.Reasons, ingestion.ReasonRateLimit)
	}
}

func TestSpotDepthRecoveryOwnerRecoversWithReplacementSnapshot(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{
		response: SpotDepthSnapshotResponse{
			RecvTime: time.UnixMilli(1772798521700).UTC(),
			Payload:  []byte(`{"lastUpdateId":704,"bids":[["64020.40","0.95"]],"asks":[["64020.70","0.80"]]}`),
		},
	})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798520000, 701, 1772798520100)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	if err := owner.MarkSequenceGap(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798521200).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798521200,"s":"BTCUSDT","U":703,"u":704,"b":[["64020.40","0.95"]],"a":[["64020.70","0.80"]]}`),
	}); err != nil {
		t.Fatalf("mark sequence gap: %v", err)
	}
	if err := owner.BufferRecoveryDelta(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798521800).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798521800,"s":"BTCUSDT","U":705,"u":706,"b":[["64020.50","0.90"]],"a":[["64020.80","0.75"]]}`),
	}); err != nil {
		t.Fatalf("buffer recovery delta: %v", err)
	}

	sync, err := owner.Recover(context.Background(), time.UnixMilli(1772798521600).UTC())
	if err != nil {
		t.Fatalf("recover: %v", err)
	}
	if sync.Snapshot.FinalSequence != 704 {
		t.Fatalf("snapshot final sequence = %d, want %d", sync.Snapshot.FinalSequence, 704)
	}
	if len(sync.Deltas) != 1 {
		t.Fatalf("aligned delta count = %d, want %d", len(sync.Deltas), 1)
	}
	status := owner.Status()
	if status.State != SpotDepthRecoverySynchronized {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthRecoverySynchronized)
	}
	if status.SequenceGapDetected {
		t.Fatal("expected sequence gap to clear after recovery")
	}
	if status.ResyncCount != 0 {
		t.Fatalf("resync count = %d, want %d", status.ResyncCount, 0)
	}
	if status.LastAcceptedSequence != 706 {
		t.Fatalf("last accepted sequence = %d, want %d", status.LastAcceptedSequence, 706)
	}
}

func TestSpotDepthRecoveryOwnerRefreshDueAndSnapshotStaleHealth(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{
		response: SpotDepthSnapshotResponse{
			RecvTime: time.UnixMilli(1772798821000).UTC(),
			Payload:  []byte(`{"lastUpdateId":706,"bids":[["64020.60","0.85"]],"asks":[["64020.90","0.70"]]}`),
		},
	})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798520000, 706, 1772798520100)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	refreshStatus, err := owner.SnapshotRefreshStatus(time.UnixMilli(1772798820000).UTC())
	if err != nil {
		t.Fatalf("snapshot refresh status: %v", err)
	}
	if !refreshStatus.Due {
		t.Fatal("expected snapshot refresh to be due after configured interval")
	}
	feedStatus, err := owner.HealthStatus(time.UnixMilli(1772798851000).UTC(), ingestion.ConnectionConnected, 0, 0)
	if err != nil {
		t.Fatalf("health status: %v", err)
	}
	if feedStatus.State != ingestion.FeedHealthStale {
		t.Fatalf("feed health state = %q, want %q", feedStatus.State, ingestion.FeedHealthStale)
	}
	if !containsDepthRecoveryReason(feedStatus.Reasons, ingestion.ReasonSnapshotStale) {
		t.Fatalf("reasons = %v, want %q", feedStatus.Reasons, ingestion.ReasonSnapshotStale)
	}
	if _, err := owner.Refresh(context.Background(), time.UnixMilli(1772798820000).UTC()); err != nil {
		t.Fatalf("refresh snapshot: %v", err)
	}
	status := owner.Status()
	if status.RefreshDue {
		t.Fatal("did not expect refresh to remain due after replacement snapshot")
	}
	if status.LastAcceptedSequence != 706 {
		t.Fatalf("last accepted sequence = %d, want %d", status.LastAcceptedSequence, 706)
	}
	if status.State != SpotDepthRecoverySynchronized {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthRecoverySynchronized)
	}
}

func TestSpotDepthRecoveryOwnerFeedHealthInputUsesDepthPrefix(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798520000, 701, 1772798520100)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	input, err := owner.FeedHealthInput(SpotDepthFeedHealthOptions{
		Symbol:          "BTC-USD",
		QuoteCurrency:   "USDT",
		Now:             time.UnixMilli(1772798520200).UTC(),
		ConnectionState: ingestion.ConnectionConnected,
	})
	if err != nil {
		t.Fatalf("feed health input: %v", err)
	}
	if input.Message.SourceRecordID != "runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("source record id = %q, want %q", input.Message.SourceRecordID, "runtime:binance-spot-depth:BTCUSDT")
	}
	if input.Message.Status.State != ingestion.FeedHealthHealthy {
		t.Fatalf("feed health state = %q, want %q", input.Message.Status.State, ingestion.FeedHealthHealthy)
	}
}

func TestSpotDepthRecoveryOwnerAcceptSynchronizedDeltaKeepsMessageFresh(t *testing.T) {
	runtime := newBinanceRuntime(t)
	owner, err := NewSpotDepthRecoveryOwner(runtime, stubSpotDepthSnapshotFetcher{})
	if err != nil {
		t.Fatalf("new recovery owner: %v", err)
	}
	if err := owner.StartSynchronized(bootstrapSync(t, 700, 1772798520000, 701, 1772798520100)); err != nil {
		t.Fatalf("start synchronized: %v", err)
	}
	if err := owner.AcceptSynchronizedDelta(SpotRawFrame{
		RecvTime: time.UnixMilli(1772798550000).UTC(),
		Payload:  []byte(`{"e":"depthUpdate","E":1772798550000,"s":"BTCUSDT","U":702,"u":702,"b":[["64020.30","0.98"]],"a":[["64020.60","0.78"]]}`),
	}); err != nil {
		t.Fatalf("accept synchronized delta: %v", err)
	}
	status := owner.Status()
	if status.LastAcceptedSequence != 702 {
		t.Fatalf("last accepted sequence = %d, want %d", status.LastAcceptedSequence, 702)
	}
	if !status.LastMessageAt.Equal(time.UnixMilli(1772798550000).UTC()) {
		t.Fatalf("last message at = %s, want %s", status.LastMessageAt, time.UnixMilli(1772798550000).UTC())
	}
	if status.State != SpotDepthRecoverySynchronized {
		t.Fatalf("state = %q, want %q", status.State, SpotDepthRecoverySynchronized)
	}
}

func bootstrapSync(t *testing.T, snapshotSequence int64, snapshotRecvMs int64, finalSequence int64, finalRecvMs int64) SpotDepthBootstrapSync {
	t.Helper()
	snapshotRecv := time.UnixMilli(snapshotRecvMs).UTC()
	deltaRecv := time.UnixMilli(finalRecvMs).UTC()
	snapshot, err := ParseOrderBookSnapshotWithSourceSymbol([]byte(`{"lastUpdateId":700,"bids":[["64020.00","1.10"]],"asks":[["64020.50","0.80"]]}`), "BTCUSDT", snapshotRecv)
	if err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	snapshot.FirstSequence = snapshotSequence
	snapshot.FinalSequence = snapshotSequence
	snapshot.Message.FirstSequence = snapshotSequence
	snapshot.Message.Sequence = snapshotSequence
	delta, err := ParseOrderBookDelta([]byte(`{"e":"depthUpdate","E":1772798520300,"s":"BTCUSDT","U":700,"u":701,"b":[["64020.10","1.05"]],"a":[["64020.40","0.75"]]}`), deltaRecv)
	if err != nil {
		t.Fatalf("parse delta: %v", err)
	}
	delta.FirstSequence = snapshotSequence
	delta.FinalSequence = finalSequence
	delta.Message.FirstSequence = snapshotSequence
	delta.Message.Sequence = finalSequence
	return SpotDepthBootstrapSync{SourceSymbol: "BTCUSDT", Snapshot: snapshot, Deltas: []ParsedOrderBook{delta}}
}
