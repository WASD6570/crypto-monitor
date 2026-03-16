package main

import (
	"context"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	marketstateapi "github.com/crypto-market-copilot/alerts/services/market-state-api"
	venuebinance "github.com/crypto-market-copilot/alerts/services/venue-binance"
)

type binanceRuntimeHealthReadiness string

const (
	binanceRuntimeHealthNotReady binanceRuntimeHealthReadiness = "NOT_READY"
	binanceRuntimeHealthReady    binanceRuntimeHealthReadiness = "READY"
)

type binanceRuntimeHealthReader interface {
	RuntimeHealthSnapshot(ctx context.Context, now time.Time) (binanceRuntimeHealthSnapshot, error)
}

type binanceRuntimeHealthSnapshot struct {
	GeneratedAt time.Time
	Symbols     []binanceRuntimeHealthSymbolSnapshot
}

type binanceRuntimeHealthSymbolSnapshot struct {
	Symbol                string
	SourceSymbol          string
	QuoteCurrency         string
	Readiness             binanceRuntimeHealthReadiness
	FeedHealth            ingestion.FeedHealthStatus
	ConnectionState       ingestion.ConnectionState
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	DepthStatus           venuebinance.SpotDepthRecoveryStatus
	LastAcceptedExchange  time.Time
	LastAcceptedRecv      time.Time
	LastMessageAt         time.Time
	LastSnapshotAt        time.Time
}

func (o *binanceSpotRuntimeOwner) RuntimeHealthSnapshot(ctx context.Context, now time.Time) (binanceRuntimeHealthSnapshot, error) {
	if ctx == nil {
		return binanceRuntimeHealthSnapshot{}, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return binanceRuntimeHealthSnapshot{}, err
	}
	if now.IsZero() {
		return binanceRuntimeHealthSnapshot{}, fmt.Errorf("current time is required")
	}
	now = now.UTC()

	o.mu.RLock()
	defer o.mu.RUnlock()
	if !o.started {
		return binanceRuntimeHealthSnapshot{}, fmt.Errorf("binance spot runtime owner is not started")
	}

	symbols := make([]binanceRuntimeHealthSymbolSnapshot, 0, len(o.order))
	for _, symbol := range o.order {
		state := o.states[symbol]
		if state == nil {
			continue
		}
		entry, err := state.runtimeHealthSnapshot(now)
		if err != nil {
			return binanceRuntimeHealthSnapshot{}, err
		}
		symbols = append(symbols, entry)
	}

	return binanceRuntimeHealthSnapshot{GeneratedAt: now, Symbols: symbols}, nil
}

func (s *spotRuntimeState) runtimeHealthSnapshot(now time.Time) (binanceRuntimeHealthSymbolSnapshot, error) {
	if s == nil {
		return binanceRuntimeHealthSymbolSnapshot{}, fmt.Errorf("spot runtime state is required")
	}
	if now.IsZero() {
		return binanceRuntimeHealthSymbolSnapshot{}, fmt.Errorf("current time is required")
	}
	now = now.UTC()

	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.depth == nil {
		return binanceRuntimeHealthSymbolSnapshot{}, fmt.Errorf("spot depth recovery owner is required")
	}

	depthStatus, err := s.depth.StatusAt(now)
	if err != nil {
		return binanceRuntimeHealthSymbolSnapshot{}, err
	}
	feedHealth, err := s.depth.HealthStatus(now, s.connectionState, s.localClockOffset, s.consecutiveReconnects)
	if err != nil {
		return binanceRuntimeHealthSymbolSnapshot{}, err
	}

	readiness := binanceRuntimeHealthNotReady
	if s.haveObservation && s.publishable {
		readiness = binanceRuntimeHealthReady
	}

	return binanceRuntimeHealthSymbolSnapshot{
		Symbol:                s.binding.Symbol,
		SourceSymbol:          s.binding.SourceSymbol,
		QuoteCurrency:         s.binding.QuoteCurrency,
		Readiness:             readiness,
		FeedHealth:            feedHealth,
		ConnectionState:       s.connectionState,
		LocalClockOffset:      s.localClockOffset,
		ConsecutiveReconnects: s.consecutiveReconnects,
		DepthStatus:           depthStatus,
		LastAcceptedExchange:  s.lastAcceptedExchange,
		LastAcceptedRecv:      s.lastAcceptedRecv,
		LastMessageAt:         depthStatus.LastMessageAt,
		LastSnapshotAt:        depthStatus.LastSnapshotAt,
	}, nil
}

func (s binanceRuntimeHealthSnapshot) runtimeStatusResponse() marketstateapi.RuntimeStatusResponse {
	symbols := make([]marketstateapi.RuntimeStatusSymbolResponse, 0, len(s.Symbols))
	for _, symbol := range s.Symbols {
		symbols = append(symbols, symbol.runtimeStatusResponse())
	}
	return marketstateapi.RuntimeStatusResponse{
		GeneratedAt: s.GeneratedAt,
		Symbols:     symbols,
	}
}

func (s binanceRuntimeHealthSymbolSnapshot) runtimeStatusResponse() marketstateapi.RuntimeStatusSymbolResponse {
	return marketstateapi.RuntimeStatusSymbolResponse{
		Symbol:                 s.Symbol,
		SourceSymbol:           s.SourceSymbol,
		QuoteCurrency:          s.QuoteCurrency,
		Readiness:              runtimeStatusReadiness(s.Readiness),
		FeedHealth:             marketstateapi.NewRuntimeStatusFeedHealthResponse(s.FeedHealth),
		ConnectionState:        s.ConnectionState,
		LocalClockOffsetMillis: s.LocalClockOffset.Milliseconds(),
		ConsecutiveReconnects:  s.ConsecutiveReconnects,
		DepthStatus:            marketstateapi.NewRuntimeStatusDepthStatusResponse(s.DepthStatus),
		LastAcceptedExchange:   nullableRuntimeStatusTime(s.LastAcceptedExchange),
		LastAcceptedRecv:       nullableRuntimeStatusTime(s.LastAcceptedRecv),
		LastMessageAt:          nullableRuntimeStatusTime(s.LastMessageAt),
		LastSnapshotAt:         nullableRuntimeStatusTime(s.LastSnapshotAt),
	}
}

func runtimeStatusReadiness(readiness binanceRuntimeHealthReadiness) marketstateapi.RuntimeStatusReadiness {
	if readiness == binanceRuntimeHealthReady {
		return marketstateapi.RuntimeStatusReady
	}
	return marketstateapi.RuntimeStatusNotReady
}

func nullableRuntimeStatusTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	utc := value.UTC()
	return &utc
}
