package venuebinance

import (
	"fmt"
	"sync"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type USDMInfluenceInputOwner struct {
	mu     sync.RWMutex
	order  []usdmSymbolBinding
	states map[string]*usdmInfluenceSymbolState
}

type usdmInfluenceSymbolState struct {
	binding                usdmSymbolBinding
	funding                *ingestion.CanonicalFundingRateEvent
	markIndex              *ingestion.CanonicalMarkIndexEvent
	liquidation            *ingestion.CanonicalLiquidationPrintEvent
	openInterest           *ingestion.CanonicalOpenInterestSnapshotEvent
	websocketFeedHealth    *ingestion.CanonicalFeedHealthEvent
	openInterestFeedHealth *ingestion.CanonicalFeedHealthEvent
}

func NewUSDMInfluenceInputOwner(runtime *Runtime) (*USDMInfluenceInputOwner, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}
	bindings, err := usdmSymbolBindings(runtime.config.Symbols)
	if err != nil {
		return nil, err
	}
	if err := features.ValidateUSDMInfluenceSymbols(symbolsFromBindings(bindings)); err != nil {
		return nil, err
	}
	states := make(map[string]*usdmInfluenceSymbolState, len(bindings))
	for _, binding := range bindings {
		binding := binding
		states[binding.Symbol] = &usdmInfluenceSymbolState{binding: binding}
	}
	return &USDMInfluenceInputOwner{order: append([]usdmSymbolBinding(nil), bindings...), states: states}, nil
}

func (o *USDMInfluenceInputOwner) AcceptFunding(event ingestion.CanonicalFundingRateEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.funding, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	state.funding = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) AcceptMarkIndex(event ingestion.CanonicalMarkIndexEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.markIndex, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	state.markIndex = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) AcceptLiquidation(event ingestion.CanonicalLiquidationPrintEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.liquidation, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	state.liquidation = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) AcceptOpenInterest(event ingestion.CanonicalOpenInterestSnapshotEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.openInterest, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	state.openInterest = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) ObserveWebsocketFeedHealth(event ingestion.CanonicalFeedHealthEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.websocketFeedHealth, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	copy.DegradationReasons = append([]ingestion.DegradationReason(nil), event.DegradationReasons...)
	state.websocketFeedHealth = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) ObserveOpenInterestFeedHealth(event ingestion.CanonicalFeedHealthEvent) error {
	if o == nil {
		return fmt.Errorf("usdm influence input owner is required")
	}
	state, err := o.lockedStateForSymbol(event.Symbol)
	if err != nil {
		return err
	}
	replace, err := shouldReplaceCanonicalEvent(state.openInterestFeedHealth, &event)
	if err != nil {
		o.mu.Unlock()
		return err
	}
	if !replace {
		o.mu.Unlock()
		return nil
	}
	copy := event
	copy.DegradationReasons = append([]ingestion.DegradationReason(nil), event.DegradationReasons...)
	state.openInterestFeedHealth = &copy
	o.mu.Unlock()
	return nil
}

func (o *USDMInfluenceInputOwner) Snapshot(now time.Time) (features.USDMInfluenceEvaluatorInput, error) {
	if o == nil {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("usdm influence input owner is required")
	}
	if now.IsZero() {
		return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("snapshot time is required")
	}
	now = now.UTC()

	o.mu.RLock()
	defer o.mu.RUnlock()

	symbols := make([]features.USDMSymbolInfluenceInput, 0, len(o.order))
	for _, binding := range o.order {
		state := o.states[binding.Symbol]
		if state == nil {
			return features.USDMInfluenceEvaluatorInput{}, fmt.Errorf("missing usdm influence state for %q", binding.Symbol)
		}
		symbol, err := state.snapshotSymbol(now)
		if err != nil {
			return features.USDMInfluenceEvaluatorInput{}, err
		}
		symbols = append(symbols, symbol)
	}
	input := features.USDMInfluenceEvaluatorInput{
		SchemaVersion: features.USDMInfluenceInputSchema,
		ObservedAt:    now.Format(time.RFC3339Nano),
		Symbols:       symbols,
	}
	if err := input.Validate(); err != nil {
		return features.USDMInfluenceEvaluatorInput{}, err
	}
	return input, nil
}

func (o *USDMInfluenceInputOwner) lockedStateForSymbol(symbol string) (*usdmInfluenceSymbolState, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	o.mu.Lock()
	state, ok := o.states[symbol]
	if !ok {
		o.mu.Unlock()
		return nil, fmt.Errorf("unsupported usdm influence symbol %q", symbol)
	}
	return state, nil
}

func (s *usdmInfluenceSymbolState) snapshotSymbol(now time.Time) (features.USDMSymbolInfluenceInput, error) {
	if s == nil {
		return features.USDMSymbolInfluenceInput{}, fmt.Errorf("usdm influence symbol state is required")
	}
	return features.USDMSymbolInfluenceInput{
		Symbol:        s.binding.Symbol,
		SourceSymbol:  s.binding.SourceSymbol,
		QuoteCurrency: s.binding.QuoteCurrency,
		Funding:       fundingInputFromEvent(now, s.funding, s.websocketFeedHealth),
		MarkIndex:     markIndexInputFromEvent(now, s.markIndex, s.websocketFeedHealth),
		Liquidation:   liquidationInputFromEvent(now, s.liquidation, s.websocketFeedHealth),
		OpenInterest:  openInterestInputFromEvent(now, s.openInterest, s.openInterestFeedHealth),
	}, nil
}

func fundingInputFromEvent(now time.Time, event *ingestion.CanonicalFundingRateEvent, health *ingestion.CanonicalFeedHealthEvent) features.USDMFundingInput {
	metadata := inputMetadataFromEvent(now, features.USDMInfluenceSurfaceWebsocket, eventTimeFields(event), eventHealthFields(health), eventTimestampStatus(event), eventSourceRecordID(event), event != nil)
	if event == nil {
		return features.USDMFundingInput{Metadata: metadata}
	}
	return features.USDMFundingInput{Metadata: metadata, FundingRate: event.FundingRate, NextFundingTs: event.NextFundingTs}
}

func markIndexInputFromEvent(now time.Time, event *ingestion.CanonicalMarkIndexEvent, health *ingestion.CanonicalFeedHealthEvent) features.USDMMarkIndexInput {
	metadata := inputMetadataFromEvent(now, features.USDMInfluenceSurfaceWebsocket, eventTimeFields(event), eventHealthFields(health), eventTimestampStatus(event), eventSourceRecordID(event), event != nil)
	if event == nil {
		return features.USDMMarkIndexInput{Metadata: metadata}
	}
	return features.USDMMarkIndexInput{Metadata: metadata, MarkPrice: event.MarkPrice, IndexPrice: event.IndexPrice}
}

func liquidationInputFromEvent(now time.Time, event *ingestion.CanonicalLiquidationPrintEvent, health *ingestion.CanonicalFeedHealthEvent) features.USDMLiquidationInput {
	metadata := inputMetadataFromEvent(now, features.USDMInfluenceSurfaceWebsocket, eventTimeFields(event), eventHealthFields(health), eventTimestampStatus(event), eventSourceRecordID(event), event != nil)
	if event == nil {
		return features.USDMLiquidationInput{Metadata: metadata}
	}
	return features.USDMLiquidationInput{Metadata: metadata, Side: event.Side, Price: event.Price, Size: event.Size}
}

func openInterestInputFromEvent(now time.Time, event *ingestion.CanonicalOpenInterestSnapshotEvent, health *ingestion.CanonicalFeedHealthEvent) features.USDMOpenInterestInput {
	metadata := inputMetadataFromEvent(now, features.USDMInfluenceSurfaceRESTPoll, eventTimeFields(event), eventHealthFields(health), eventTimestampStatus(event), eventSourceRecordID(event), event != nil)
	if event == nil {
		return features.USDMOpenInterestInput{Metadata: metadata}
	}
	return features.USDMOpenInterestInput{Metadata: metadata, OpenInterest: event.OpenInterest}
}

type usdmEventTimeFields struct {
	exchangeTs string
	recvTs     string
}

type usdmHealthFields struct {
	state   ingestion.FeedHealthState
	reasons []ingestion.DegradationReason
	ok      bool
}

func inputMetadataFromEvent(now time.Time, surface string, times usdmEventTimeFields, health usdmHealthFields, timestampStatus ingestion.CanonicalTimestampStatus, sourceRecordID string, available bool) features.USDMInfluenceInputMetadata {
	metadata := features.USDMInfluenceInputMetadata{
		Surface:         surface,
		Available:       available,
		Freshness:       features.USDMInfluenceFreshnessUnavailable,
		ExchangeTs:      times.exchangeTs,
		RecvTs:          times.recvTs,
		AgeMillis:       -1,
		TimestampStatus: timestampStatus,
		SourceRecordID:  sourceRecordID,
	}
	if !available {
		if health.ok {
			metadata.FeedHealthState = health.state
			metadata.FeedHealthReasons = append([]ingestion.DegradationReason(nil), health.reasons...)
		}
		return metadata
	}
	if recvAt, err := time.Parse(time.RFC3339Nano, times.recvTs); err == nil && !recvAt.After(now) {
		metadata.AgeMillis = now.Sub(recvAt).Milliseconds()
	}
	if !health.ok {
		metadata.FeedHealthState = ingestion.FeedHealthHealthy
	} else {
		metadata.FeedHealthState = health.state
		metadata.FeedHealthReasons = append([]ingestion.DegradationReason(nil), health.reasons...)
	}
	metadata.Freshness = freshnessFromInput(metadata.FeedHealthState, timestampStatus)
	return metadata
}

func freshnessFromInput(feedHealthState ingestion.FeedHealthState, timestampStatus ingestion.CanonicalTimestampStatus) features.USDMInfluenceFreshness {
	if feedHealthState == ingestion.FeedHealthStale {
		return features.USDMInfluenceFreshnessStale
	}
	if feedHealthState == ingestion.FeedHealthDegraded || timestampStatus == ingestion.TimestampStatusDegraded {
		return features.USDMInfluenceFreshnessDegraded
	}
	return features.USDMInfluenceFreshnessFresh
}

func eventHealthFields(event *ingestion.CanonicalFeedHealthEvent) usdmHealthFields {
	if event == nil {
		return usdmHealthFields{}
	}
	return usdmHealthFields{state: event.FeedHealthState, reasons: append([]ingestion.DegradationReason(nil), event.DegradationReasons...), ok: true}
}

func eventTimeFields(event any) usdmEventTimeFields {
	switch typed := event.(type) {
	case *ingestion.CanonicalFundingRateEvent:
		if typed == nil {
			return usdmEventTimeFields{}
		}
		return usdmEventTimeFields{exchangeTs: typed.ExchangeTs, recvTs: typed.RecvTs}
	case *ingestion.CanonicalMarkIndexEvent:
		if typed == nil {
			return usdmEventTimeFields{}
		}
		return usdmEventTimeFields{exchangeTs: typed.ExchangeTs, recvTs: typed.RecvTs}
	case *ingestion.CanonicalLiquidationPrintEvent:
		if typed == nil {
			return usdmEventTimeFields{}
		}
		return usdmEventTimeFields{exchangeTs: typed.ExchangeTs, recvTs: typed.RecvTs}
	case *ingestion.CanonicalOpenInterestSnapshotEvent:
		if typed == nil {
			return usdmEventTimeFields{}
		}
		return usdmEventTimeFields{exchangeTs: typed.ExchangeTs, recvTs: typed.RecvTs}
	default:
		return usdmEventTimeFields{}
	}
}

func eventTimestampStatus(event any) ingestion.CanonicalTimestampStatus {
	switch typed := event.(type) {
	case *ingestion.CanonicalFundingRateEvent:
		if typed != nil {
			return typed.TimestampStatus
		}
	case *ingestion.CanonicalMarkIndexEvent:
		if typed != nil {
			return typed.TimestampStatus
		}
	case *ingestion.CanonicalLiquidationPrintEvent:
		if typed != nil {
			return typed.TimestampStatus
		}
	case *ingestion.CanonicalOpenInterestSnapshotEvent:
		if typed != nil {
			return typed.TimestampStatus
		}
	}
	return ""
}

func eventSourceRecordID(event any) string {
	switch typed := event.(type) {
	case *ingestion.CanonicalFundingRateEvent:
		if typed != nil {
			return typed.SourceRecordID
		}
	case *ingestion.CanonicalMarkIndexEvent:
		if typed != nil {
			return typed.SourceRecordID
		}
	case *ingestion.CanonicalLiquidationPrintEvent:
		if typed != nil {
			return typed.SourceRecordID
		}
	case *ingestion.CanonicalOpenInterestSnapshotEvent:
		if typed != nil {
			return typed.SourceRecordID
		}
	}
	return ""
}

func symbolsFromBindings(bindings []usdmSymbolBinding) []string {
	symbols := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		symbols = append(symbols, binding.Symbol)
	}
	return symbols
}

func shouldReplaceCanonicalEvent(current any, candidate any) (bool, error) {
	candidateRecv, candidateExchange, err := canonicalEventOrderingTimes(candidate)
	if err != nil {
		return false, err
	}
	currentRecv, currentExchange, err := canonicalEventOrderingTimes(current)
	if err != nil {
		return false, err
	}
	if currentRecv.IsZero() {
		return true, nil
	}
	if candidateRecv.After(currentRecv) {
		return true, nil
	}
	if candidateRecv.Before(currentRecv) {
		return false, nil
	}
	if currentExchange.IsZero() {
		return !candidateExchange.IsZero(), nil
	}
	if candidateExchange.After(currentExchange) {
		return true, nil
	}
	return false, nil
}

func canonicalEventOrderingTimes(event any) (time.Time, time.Time, error) {
	switch typed := event.(type) {
	case *ingestion.CanonicalFundingRateEvent:
		if typed == nil {
			return time.Time{}, time.Time{}, nil
		}
		recv, err := parseOptionalTimestamp(typed.RecvTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		exchange, err := parseOptionalTimestamp(typed.ExchangeTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return recv, exchange, nil
	case *ingestion.CanonicalMarkIndexEvent:
		if typed == nil {
			return time.Time{}, time.Time{}, nil
		}
		recv, err := parseOptionalTimestamp(typed.RecvTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		exchange, err := parseOptionalTimestamp(typed.ExchangeTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return recv, exchange, nil
	case *ingestion.CanonicalLiquidationPrintEvent:
		if typed == nil {
			return time.Time{}, time.Time{}, nil
		}
		recv, err := parseOptionalTimestamp(typed.RecvTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		exchange, err := parseOptionalTimestamp(typed.ExchangeTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return recv, exchange, nil
	case *ingestion.CanonicalOpenInterestSnapshotEvent:
		if typed == nil {
			return time.Time{}, time.Time{}, nil
		}
		recv, err := parseOptionalTimestamp(typed.RecvTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		exchange, err := parseOptionalTimestamp(typed.ExchangeTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return recv, exchange, nil
	case *ingestion.CanonicalFeedHealthEvent:
		if typed == nil {
			return time.Time{}, time.Time{}, nil
		}
		recv, err := parseOptionalTimestamp(typed.RecvTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		exchange, err := parseOptionalTimestamp(typed.ExchangeTs)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return recv, exchange, nil
	case nil:
		return time.Time{}, time.Time{}, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported canonical event type %T", event)
	}
}

func parseOptionalTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", value, err)
	}
	return parsed.UTC(), nil
}
