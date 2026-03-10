package venuebinance

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type SpotDepthBootstrapState string

const (
	SpotDepthBootstrapIdle              SpotDepthBootstrapState = "idle"
	SpotDepthBootstrapBuffering         SpotDepthBootstrapState = "buffering"
	SpotDepthBootstrapSnapshotRequested SpotDepthBootstrapState = "snapshot-requested"
	SpotDepthBootstrapSynchronized      SpotDepthBootstrapState = "synchronized"
	SpotDepthBootstrapFailed            SpotDepthBootstrapState = "bootstrap-failed"
)

type SpotDepthBootstrapFailureReason string

const (
	SpotDepthBootstrapFailureNone                SpotDepthBootstrapFailureReason = ""
	SpotDepthBootstrapFailureBufferedDeltaNeeded SpotDepthBootstrapFailureReason = "buffered-delta-required"
	SpotDepthBootstrapFailureSnapshotFetch       SpotDepthBootstrapFailureReason = "snapshot-fetch-failed"
	SpotDepthBootstrapFailureSnapshotParse       SpotDepthBootstrapFailureReason = "snapshot-parse-failed"
	SpotDepthBootstrapFailureSymbolMismatch      SpotDepthBootstrapFailureReason = "source-symbol-mismatch"
	SpotDepthBootstrapFailureBridgeMissing       SpotDepthBootstrapFailureReason = "bridging-delta-missing"
)

type SpotDepthSnapshotResponse struct {
	Payload  []byte
	RecvTime time.Time
}

type SpotDepthSnapshotFetcher interface {
	FetchSpotDepthSnapshot(ctx context.Context, sourceSymbol string) (SpotDepthSnapshotResponse, error)
}

type SpotDepthBootstrapSync struct {
	SourceSymbol string
	Snapshot     ParsedOrderBook
	Deltas       []ParsedOrderBook
}

type SpotDepthBootstrapStatus struct {
	State               SpotDepthBootstrapState
	SourceSymbol        string
	BufferedDeltaCount  int
	SnapshotRequestedAt time.Time
	FailureReason       SpotDepthBootstrapFailureReason
	Synchronized        bool
}

type SpotDepthBootstrapOwner struct {
	runtime             *Runtime
	fetcher             SpotDepthSnapshotFetcher
	state               SpotDepthBootstrapState
	sourceSymbol        string
	buffered            []ParsedOrderBook
	snapshotRequestedAt time.Time
	failureReason       SpotDepthBootstrapFailureReason
}

func NewSpotDepthBootstrapOwner(runtime *Runtime, fetcher SpotDepthSnapshotFetcher) (*SpotDepthBootstrapOwner, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}
	if fetcher == nil {
		return nil, fmt.Errorf("snapshot fetcher is required")
	}
	if !hasSpotOrderBookStream(runtime.config.Adapter.Streams) {
		return nil, fmt.Errorf("binance spot depth bootstrap requires spot order-book stream")
	}

	return &SpotDepthBootstrapOwner{
		runtime: runtime,
		fetcher: fetcher,
		state:   SpotDepthBootstrapIdle,
	}, nil
}

func (o *SpotDepthBootstrapOwner) Status() SpotDepthBootstrapStatus {
	if o == nil {
		return SpotDepthBootstrapStatus{}
	}
	return SpotDepthBootstrapStatus{
		State:               o.state,
		SourceSymbol:        o.sourceSymbol,
		BufferedDeltaCount:  len(o.buffered),
		SnapshotRequestedAt: o.snapshotRequestedAt,
		FailureReason:       o.failureReason,
		Synchronized:        o.state == SpotDepthBootstrapSynchronized,
	}
}

func (o *SpotDepthBootstrapOwner) BufferDelta(frame SpotRawFrame) error {
	if o == nil {
		return fmt.Errorf("spot depth bootstrap owner is required")
	}
	if o.state == SpotDepthBootstrapSynchronized {
		return fmt.Errorf("spot depth bootstrap is already synchronized")
	}
	if o.state == SpotDepthBootstrapFailed {
		return fmt.Errorf("spot depth bootstrap failed: %s", o.failureReason)
	}

	parsed, err := ParseOrderBookFrame(frame)
	if err != nil {
		return err
	}
	if o.sourceSymbol == "" {
		o.sourceSymbol = parsed.SourceSymbol
	}
	if parsed.SourceSymbol != o.sourceSymbol {
		return fmt.Errorf("depth bootstrap source symbol = %q, want %q", parsed.SourceSymbol, o.sourceSymbol)
	}

	o.buffered = append(o.buffered, parsed)
	o.state = SpotDepthBootstrapBuffering
	return nil
}

func (o *SpotDepthBootstrapOwner) Synchronize(ctx context.Context) (SpotDepthBootstrapSync, error) {
	if o == nil {
		return SpotDepthBootstrapSync{}, fmt.Errorf("spot depth bootstrap owner is required")
	}
	if ctx == nil {
		return SpotDepthBootstrapSync{}, fmt.Errorf("context is required")
	}
	if o.sourceSymbol == "" || len(o.buffered) == 0 {
		o.state = SpotDepthBootstrapFailed
		o.failureReason = SpotDepthBootstrapFailureBufferedDeltaNeeded
		return SpotDepthBootstrapSync{}, fmt.Errorf("buffered depth delta is required before snapshot bootstrap")
	}

	o.state = SpotDepthBootstrapSnapshotRequested
	snapshotResponse, err := o.fetcher.FetchSpotDepthSnapshot(ctx, o.sourceSymbol)
	if err != nil {
		o.state = SpotDepthBootstrapFailed
		o.failureReason = SpotDepthBootstrapFailureSnapshotFetch
		return SpotDepthBootstrapSync{}, fmt.Errorf("fetch spot depth snapshot: %w", err)
	}
	o.snapshotRequestedAt = snapshotResponse.RecvTime

	snapshot, err := ParseOrderBookSnapshotWithSourceSymbol(snapshotResponse.Payload, o.sourceSymbol, snapshotResponse.RecvTime)
	if err != nil {
		o.state = SpotDepthBootstrapFailed
		o.failureReason = SpotDepthBootstrapFailureSnapshotParse
		return SpotDepthBootstrapSync{}, fmt.Errorf("parse spot depth snapshot: %w", err)
	}
	if snapshot.SourceSymbol != o.sourceSymbol {
		o.state = SpotDepthBootstrapFailed
		o.failureReason = SpotDepthBootstrapFailureSymbolMismatch
		return SpotDepthBootstrapSync{}, fmt.Errorf("snapshot source symbol = %q, want %q", snapshot.SourceSymbol, o.sourceSymbol)
	}

	aligned, err := alignBufferedDeltas(snapshot.FinalSequence, o.buffered)
	if err != nil {
		o.state = SpotDepthBootstrapFailed
		o.failureReason = SpotDepthBootstrapFailureBridgeMissing
		return SpotDepthBootstrapSync{}, err
	}

	o.state = SpotDepthBootstrapSynchronized
	o.failureReason = SpotDepthBootstrapFailureNone
	return SpotDepthBootstrapSync{
		SourceSymbol: o.sourceSymbol,
		Snapshot:     snapshot,
		Deltas:       aligned,
	}, nil
}

func alignBufferedDeltas(snapshotSequence int64, buffered []ParsedOrderBook) ([]ParsedOrderBook, error) {
	if snapshotSequence <= 0 {
		return nil, fmt.Errorf("snapshot sequence must be positive")
	}
	if len(buffered) == 0 {
		return nil, fmt.Errorf("buffered depth delta is required before snapshot bootstrap")
	}

	expected := snapshotSequence + 1
	for idx, delta := range buffered {
		if delta.FirstSequence <= 0 || delta.FinalSequence <= 0 {
			return nil, fmt.Errorf("buffered delta sequence window is required")
		}
		if delta.FinalSequence <= snapshotSequence {
			continue
		}
		if delta.FirstSequence <= expected && delta.FinalSequence >= expected {
			return slices.Clone(buffered[idx:]), nil
		}
	}

	return nil, fmt.Errorf("no buffered depth delta bridges snapshot sequence %d", snapshotSequence)
}

func hasSpotOrderBookStream(streams []ingestion.StreamDefinition) bool {
	for _, stream := range streams {
		if stream.MarketType == "spot" && stream.Kind == ingestion.StreamOrderBook {
			return true
		}
	}
	return false
}
