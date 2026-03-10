package venuebinance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedOrderBook struct {
	SourceSymbol  string
	FirstSequence int64
	FinalSequence int64
	Message       ingestion.OrderBookMessage
}

type bookLevel [2]string

type orderBookSnapshotPayload struct {
	LastUpdateID int64       `json:"lastUpdateId"`
	SourceSymbol string      `json:"symbol"`
	Bids         []bookLevel `json:"bids"`
	Asks         []bookLevel `json:"asks"`
}

type orderBookDeltaPayload struct {
	EventType     string      `json:"e"`
	EventTimeMs   int64       `json:"E"`
	SourceSymbol  string      `json:"s"`
	FirstUpdateID int64       `json:"U"`
	FinalUpdateID int64       `json:"u"`
	Bids          []bookLevel `json:"b"`
	Asks          []bookLevel `json:"a"`
}

func ParseOrderBookSnapshot(raw []byte, recvTime time.Time) (ParsedOrderBook, error) {
	return parseOrderBookSnapshot(raw, "", recvTime)
}

func ParseOrderBookSnapshotWithSourceSymbol(raw []byte, sourceSymbol string, recvTime time.Time) (ParsedOrderBook, error) {
	return parseOrderBookSnapshot(raw, sourceSymbol, recvTime)
}

func parseOrderBookSnapshot(raw []byte, sourceSymbol string, recvTime time.Time) (ParsedOrderBook, error) {
	if recvTime.IsZero() {
		return ParsedOrderBook{}, fmt.Errorf("recv time is required")
	}

	var payload orderBookSnapshotPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("decode binance order-book snapshot: %w", err)
	}
	resolvedSourceSymbol := payload.SourceSymbol
	if resolvedSourceSymbol == "" {
		resolvedSourceSymbol = sourceSymbol
	}
	if payload.SourceSymbol != "" && sourceSymbol != "" && payload.SourceSymbol != sourceSymbol {
		return ParsedOrderBook{}, fmt.Errorf("snapshot source symbol = %q, want %q", payload.SourceSymbol, sourceSymbol)
	}
	if resolvedSourceSymbol == "" {
		return ParsedOrderBook{}, fmt.Errorf("source symbol is required")
	}
	if payload.LastUpdateID <= 0 {
		return ParsedOrderBook{}, fmt.Errorf("last update id must be positive")
	}

	bestBidPrice, err := bestBookPrice(payload.Bids, "bid")
	if err != nil {
		return ParsedOrderBook{}, err
	}
	bestAskPrice, err := bestBookPrice(payload.Asks, "ask")
	if err != nil {
		return ParsedOrderBook{}, err
	}

	return ParsedOrderBook{
		SourceSymbol:  resolvedSourceSymbol,
		FirstSequence: payload.LastUpdateID,
		FinalSequence: payload.LastUpdateID,
		Message: ingestion.OrderBookMessage{
			Type:          "snapshot",
			FirstSequence: payload.LastUpdateID,
			Sequence:      payload.LastUpdateID,
			BestBidPrice:  bestBidPrice,
			BestAskPrice:  bestAskPrice,
			RecvTs:        recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func ParseOrderBookDelta(raw []byte, recvTime time.Time) (ParsedOrderBook, error) {
	if recvTime.IsZero() {
		return ParsedOrderBook{}, fmt.Errorf("recv time is required")
	}

	var payload orderBookDeltaPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("decode binance order-book delta: %w", err)
	}
	if payload.EventType != "depthUpdate" {
		return ParsedOrderBook{}, fmt.Errorf("unsupported binance event type %q", payload.EventType)
	}
	if payload.SourceSymbol == "" {
		return ParsedOrderBook{}, fmt.Errorf("source symbol is required")
	}
	if payload.FirstUpdateID <= 0 {
		return ParsedOrderBook{}, fmt.Errorf("first update id must be positive")
	}
	if payload.FinalUpdateID < payload.FirstUpdateID {
		return ParsedOrderBook{}, fmt.Errorf("final update id must be greater than or equal to first update id")
	}

	bestBidPrice, err := bestBookPrice(payload.Bids, "bid")
	if err != nil {
		return ParsedOrderBook{}, err
	}
	bestAskPrice, err := bestBookPrice(payload.Asks, "ask")
	if err != nil {
		return ParsedOrderBook{}, err
	}
	exchangeTimestamp, err := formatUnixMilliTimestamp(payload.EventTimeMs)
	if err != nil {
		return ParsedOrderBook{}, err
	}

	return ParsedOrderBook{
		SourceSymbol:  payload.SourceSymbol,
		FirstSequence: payload.FirstUpdateID,
		FinalSequence: payload.FinalUpdateID,
		Message: ingestion.OrderBookMessage{
			Type:          "delta",
			FirstSequence: payload.FirstUpdateID,
			Sequence:      payload.FinalUpdateID,
			BestBidPrice:  bestBidPrice,
			BestAskPrice:  bestAskPrice,
			ExchangeTs:    exchangeTimestamp,
			RecvTs:        recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func bestBookPrice(levels []bookLevel, side string) (string, error) {
	if len(levels) == 0 {
		return "", fmt.Errorf("%s levels are required", side)
	}
	if levels[0][0] == "" {
		return "", fmt.Errorf("%s price is required", side)
	}
	return levels[0][0], nil
}

func formatUnixMilliTimestamp(value int64) (string, error) {
	if value <= 0 {
		return "", fmt.Errorf("event time must be present")
	}
	return time.UnixMilli(value).UTC().Format(time.RFC3339Nano), nil
}

func ParseOrderBookFrame(frame SpotRawFrame) (ParsedOrderBook, error) {
	return ParseOrderBookDelta(frame.Payload, frame.RecvTime)
}
