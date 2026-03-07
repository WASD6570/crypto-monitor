package ingestion

import (
	"fmt"
	"time"
)

type TradeMetadata struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	Venue         Venue
	MarketType    string
}

type TradeMessage struct {
	Type       string `json:"type"`
	TradeID    string `json:"tradeId"`
	Price      string `json:"price"`
	Size       string `json:"size"`
	Side       string `json:"side"`
	ExchangeTs string `json:"exchangeTs"`
	RecvTs     string `json:"recvTs"`
}

type CanonicalTradeEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	Symbol                  string                   `json:"symbol"`
	SourceSymbol            string                   `json:"sourceSymbol"`
	QuoteCurrency           string                   `json:"quoteCurrency"`
	Venue                   Venue                    `json:"venue"`
	MarketType              string                   `json:"marketType"`
	ExchangeTs              string                   `json:"exchangeTs"`
	RecvTs                  string                   `json:"recvTs"`
	TimestampStatus         CanonicalTimestampStatus `json:"timestampStatus"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

func NormalizeTradeMessage(metadata TradeMetadata, message TradeMessage, policy TimestampPolicy) (CanonicalTradeEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalTradeEvent{}, fmt.Errorf("trade metadata is incomplete")
	}
	if message.Type != "trade" {
		return CanonicalTradeEvent{}, fmt.Errorf("unsupported trade message type %q", message.Type)
	}
	if message.TradeID == "" {
		return CanonicalTradeEvent{}, fmt.Errorf("trade id is required")
	}

	recvTime, err := time.Parse(time.RFC3339Nano, message.RecvTs)
	if err != nil {
		return CanonicalTradeEvent{}, fmt.Errorf("parse recv timestamp: %w", err)
	}

	var exchangeTime time.Time
	if message.ExchangeTs != "" {
		exchangeTime, err = time.Parse(time.RFC3339Nano, message.ExchangeTs)
		if err != nil {
			exchangeTime = time.Time{}
		}
	}

	resolved, err := ResolveCanonicalTimestamp(exchangeTime, recvTime, policy)
	if err != nil {
		return CanonicalTradeEvent{}, err
	}

	return CanonicalTradeEvent{
		SchemaVersion:           "v1",
		EventType:               "market-trade",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolved.Status,
		SourceRecordID:          "trade:" + message.TradeID,
		CanonicalEventTime:      resolved.EventTime,
		TimestampFallbackReason: resolved.FallbackReason,
	}, nil
}
