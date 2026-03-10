package venuebinance

import (
	"fmt"
	"strings"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const (
	spotWebsocketRawConnectionRef       = "binance-spot-ws"
	usdmWebsocketRawConnectionRef       = "binance-usdm-ws"
	usdmOpenInterestRawConnectionRef    = "binance-usdm-rest"
	usdmOpenInterestRawSessionPrefix    = "binance-usdm-rest-session-"
	usdmWebsocketFeedHealthStreamKey    = string(ingestion.StreamMarkIndex)
	usdmOpenInterestFeedHealthStreamKey = string(ingestion.StreamOpenInterest)
)

func SpotWebsocketRawWriteContext(supervisor *SpotWebsocketSupervisor) (ingestion.RawWriteContext, error) {
	if supervisor == nil {
		return ingestion.RawWriteContext{}, fmt.Errorf("spot websocket supervisor is required")
	}
	state := supervisor.State()
	if state.SessionRef == "" {
		return ingestion.RawWriteContext{}, fmt.Errorf("spot websocket session ref is required")
	}
	return ingestion.RawWriteContext{ConnectionRef: spotWebsocketRawConnectionRef, SessionRef: state.SessionRef}, nil
}

func SpotDepthRecoveryRawWriteContext(supervisor *SpotWebsocketSupervisor, degradedFeedRef string) (ingestion.RawWriteContext, error) {
	context, err := SpotWebsocketRawWriteContext(supervisor)
	if err != nil {
		return ingestion.RawWriteContext{}, err
	}
	context.DegradedFeedRef = degradedFeedRef
	return context, nil
}

func USDMWebsocketRawWriteContext(runtime *USDMRuntime) (ingestion.RawWriteContext, error) {
	if runtime == nil {
		return ingestion.RawWriteContext{}, fmt.Errorf("usdm runtime is required")
	}
	state := runtime.State()
	if state.SessionRef == "" {
		return ingestion.RawWriteContext{}, fmt.Errorf("usdm websocket session ref is required")
	}
	return ingestion.RawWriteContext{ConnectionRef: usdmWebsocketRawConnectionRef, SessionRef: state.SessionRef}, nil
}

func USDMOpenInterestRawWriteContext(poller *USDMOpenInterestPoller, sourceSymbol string) (ingestion.RawWriteContext, error) {
	if poller == nil {
		return ingestion.RawWriteContext{}, fmt.Errorf("open interest poller is required")
	}
	if sourceSymbol == "" {
		return ingestion.RawWriteContext{}, fmt.Errorf("source symbol is required")
	}
	if _, ok := poller.states[sourceSymbol]; !ok {
		return ingestion.RawWriteContext{}, fmt.Errorf("unsupported open interest source symbol %q", sourceSymbol)
	}
	return ingestion.RawWriteContext{
		ConnectionRef: usdmOpenInterestRawConnectionRef,
		SessionRef:    usdmOpenInterestRawSessionPrefix + strings.ToLower(sourceSymbol),
	}, nil
}

func BuildSpotTradeRawAppendEntry(supervisor *SpotWebsocketSupervisor, event ingestion.CanonicalTradeEvent, metadata ingestion.TradeMetadata, message ingestion.TradeMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := SpotWebsocketRawWriteContext(supervisor)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromTrade(event, metadata, message, context, options)
}

func BuildSpotOrderBookRawAppendEntry(supervisor *SpotWebsocketSupervisor, event ingestion.CanonicalOrderBookEvent, metadata ingestion.BookMetadata, message ingestion.OrderBookMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := SpotWebsocketRawWriteContext(supervisor)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromOrderBook(event, metadata, message, context, options)
}

func BuildSpotDepthFeedHealthRawAppendEntry(supervisor *SpotWebsocketSupervisor, event ingestion.CanonicalFeedHealthEvent, sourceInstrumentID string, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := SpotDepthRecoveryRawWriteContext(supervisor, degradedFeedRefForEvent(event))
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromFeedHealth(event, sourceInstrumentID, string(ingestion.StreamOrderBook), context, options)
}

func BuildUSDMFundingRawAppendEntry(runtime *USDMRuntime, event ingestion.CanonicalFundingRateEvent, metadata ingestion.DerivativesMetadata, message ingestion.FundingRateMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMWebsocketRawWriteContext(runtime)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromFundingRate(event, metadata, message, context, options)
}

func BuildUSDMMarkIndexRawAppendEntry(runtime *USDMRuntime, event ingestion.CanonicalMarkIndexEvent, metadata ingestion.DerivativesMetadata, message ingestion.MarkIndexMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMWebsocketRawWriteContext(runtime)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromMarkIndex(event, metadata, message, context, options)
}

func BuildUSDMOpenInterestRawAppendEntry(poller *USDMOpenInterestPoller, event ingestion.CanonicalOpenInterestSnapshotEvent, metadata ingestion.DerivativesMetadata, message ingestion.OpenInterestMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMOpenInterestRawWriteContext(poller, metadata.SourceSymbol)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromOpenInterest(event, metadata, message, context, options)
}

func BuildUSDMLiquidationRawAppendEntry(runtime *USDMRuntime, event ingestion.CanonicalLiquidationPrintEvent, metadata ingestion.DerivativesMetadata, message ingestion.LiquidationMessage, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMWebsocketRawWriteContext(runtime)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	return ingestion.BuildRawAppendEntryFromLiquidation(event, metadata, message, context, options)
}

func BuildUSDMWebsocketFeedHealthRawAppendEntry(runtime *USDMRuntime, event ingestion.CanonicalFeedHealthEvent, sourceInstrumentID string, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMWebsocketRawWriteContext(runtime)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	context.DegradedFeedRef = degradedFeedRefForEvent(event)
	return ingestion.BuildRawAppendEntryFromFeedHealth(event, sourceInstrumentID, usdmWebsocketFeedHealthStreamKey, context, options)
}

func BuildUSDMOpenInterestFeedHealthRawAppendEntry(poller *USDMOpenInterestPoller, event ingestion.CanonicalFeedHealthEvent, sourceInstrumentID string, options ingestion.RawWriteOptions) (ingestion.RawAppendEntry, error) {
	context, err := USDMOpenInterestRawWriteContext(poller, sourceInstrumentID)
	if err != nil {
		return ingestion.RawAppendEntry{}, err
	}
	context.DegradedFeedRef = degradedFeedRefForEvent(event)
	return ingestion.BuildRawAppendEntryFromFeedHealth(event, sourceInstrumentID, usdmOpenInterestFeedHealthStreamKey, context, options)
}

func degradedFeedRefForEvent(event ingestion.CanonicalFeedHealthEvent) string {
	if len(event.DegradationReasons) == 0 {
		return ""
	}
	return event.SourceRecordID
}
