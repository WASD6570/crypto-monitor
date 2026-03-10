package ingestion

import (
	"fmt"
	"strings"
	"time"
)

type DerivativesMetadata struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	Venue         Venue
	MarketType    string
}

type FundingRateMessage struct {
	Type          string `json:"type"`
	FundingRate   string `json:"fundingRate,omitempty"`
	NextFundingTs string `json:"nextFundingTs,omitempty"`
	ExchangeTs    string `json:"exchangeTs"`
	RecvTs        string `json:"recvTs"`
}

type MarkIndexMessage struct {
	Type       string `json:"type"`
	MarkPrice  string `json:"markPrice,omitempty"`
	IndexPrice string `json:"indexPrice,omitempty"`
	ExchangeTs string `json:"exchangeTs"`
	RecvTs     string `json:"recvTs"`
}

type LiquidationMessage struct {
	Type          string `json:"type"`
	LiquidationID string `json:"liquidationId,omitempty"`
	Side          string `json:"side,omitempty"`
	Price         string `json:"price,omitempty"`
	Size          string `json:"size,omitempty"`
	ExchangeTs    string `json:"exchangeTs"`
	RecvTs        string `json:"recvTs"`
}

type OpenInterestMessage struct {
	Type         string `json:"type"`
	OpenInterest string `json:"openInterest,omitempty"`
	ExchangeTs   string `json:"exchangeTs"`
	RecvTs       string `json:"recvTs"`
}

type CanonicalFundingRateEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	Symbol                  string                   `json:"symbol"`
	SourceSymbol            string                   `json:"sourceSymbol"`
	QuoteCurrency           string                   `json:"quoteCurrency"`
	Venue                   Venue                    `json:"venue"`
	MarketType              string                   `json:"marketType"`
	FundingRate             string                   `json:"fundingRate,omitempty"`
	NextFundingTs           string                   `json:"nextFundingTs,omitempty"`
	ExchangeTs              string                   `json:"exchangeTs"`
	RecvTs                  string                   `json:"recvTs"`
	TimestampStatus         CanonicalTimestampStatus `json:"timestampStatus"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

type CanonicalMarkIndexEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	Symbol                  string                   `json:"symbol"`
	SourceSymbol            string                   `json:"sourceSymbol"`
	QuoteCurrency           string                   `json:"quoteCurrency"`
	Venue                   Venue                    `json:"venue"`
	MarketType              string                   `json:"marketType"`
	MarkPrice               string                   `json:"markPrice,omitempty"`
	IndexPrice              string                   `json:"indexPrice,omitempty"`
	ExchangeTs              string                   `json:"exchangeTs"`
	RecvTs                  string                   `json:"recvTs"`
	TimestampStatus         CanonicalTimestampStatus `json:"timestampStatus"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

type CanonicalLiquidationPrintEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	Symbol                  string                   `json:"symbol"`
	SourceSymbol            string                   `json:"sourceSymbol"`
	QuoteCurrency           string                   `json:"quoteCurrency"`
	Venue                   Venue                    `json:"venue"`
	MarketType              string                   `json:"marketType"`
	Side                    string                   `json:"side,omitempty"`
	Price                   string                   `json:"price,omitempty"`
	Size                    string                   `json:"size,omitempty"`
	ExchangeTs              string                   `json:"exchangeTs"`
	RecvTs                  string                   `json:"recvTs"`
	TimestampStatus         CanonicalTimestampStatus `json:"timestampStatus"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

type CanonicalOpenInterestSnapshotEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	Symbol                  string                   `json:"symbol"`
	SourceSymbol            string                   `json:"sourceSymbol"`
	QuoteCurrency           string                   `json:"quoteCurrency"`
	Venue                   Venue                    `json:"venue"`
	MarketType              string                   `json:"marketType"`
	OpenInterest            string                   `json:"openInterest,omitempty"`
	ExchangeTs              string                   `json:"exchangeTs"`
	RecvTs                  string                   `json:"recvTs"`
	TimestampStatus         CanonicalTimestampStatus `json:"timestampStatus"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

func NormalizeFundingMessage(metadata DerivativesMetadata, message FundingRateMessage, policy TimestampPolicy) (CanonicalFundingRateEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalFundingRateEvent{}, fmt.Errorf("derivatives metadata is incomplete")
	}
	if message.Type != "funding-rate" {
		return CanonicalFundingRateEvent{}, fmt.Errorf("unsupported funding message type %q", message.Type)
	}
	if message.FundingRate == "" {
		return CanonicalFundingRateEvent{}, fmt.Errorf("funding rate is required")
	}

	resolved, err := resolveDerivativesTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return CanonicalFundingRateEvent{}, err
	}

	return CanonicalFundingRateEvent{
		SchemaVersion:           "v1",
		EventType:               "funding-rate",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		FundingRate:             message.FundingRate,
		NextFundingTs:           message.NextFundingTs,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolved.Status,
		SourceRecordID:          derivativesSourceRecordID("funding", message.ExchangeTs, message.RecvTs),
		CanonicalEventTime:      resolved.EventTime,
		TimestampFallbackReason: resolved.FallbackReason,
	}, nil
}

func NormalizeMarkIndexMessage(metadata DerivativesMetadata, message MarkIndexMessage, policy TimestampPolicy) (CanonicalMarkIndexEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalMarkIndexEvent{}, fmt.Errorf("derivatives metadata is incomplete")
	}
	if message.Type != "mark-index" {
		return CanonicalMarkIndexEvent{}, fmt.Errorf("unsupported mark index message type %q", message.Type)
	}
	if message.MarkPrice == "" || message.IndexPrice == "" {
		return CanonicalMarkIndexEvent{}, fmt.Errorf("mark and index prices are required")
	}

	resolved, err := resolveDerivativesTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return CanonicalMarkIndexEvent{}, err
	}

	return CanonicalMarkIndexEvent{
		SchemaVersion:           "v1",
		EventType:               "mark-index",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		MarkPrice:               message.MarkPrice,
		IndexPrice:              message.IndexPrice,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolved.Status,
		SourceRecordID:          derivativesSourceRecordID("mark-index", message.ExchangeTs, message.RecvTs),
		CanonicalEventTime:      resolved.EventTime,
		TimestampFallbackReason: resolved.FallbackReason,
	}, nil
}

func NormalizeOpenInterestMessage(metadata DerivativesMetadata, message OpenInterestMessage, policy TimestampPolicy) (CanonicalOpenInterestSnapshotEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalOpenInterestSnapshotEvent{}, fmt.Errorf("derivatives metadata is incomplete")
	}
	if message.Type != "open-interest" {
		return CanonicalOpenInterestSnapshotEvent{}, fmt.Errorf("unsupported open interest message type %q", message.Type)
	}
	if message.OpenInterest == "" {
		return CanonicalOpenInterestSnapshotEvent{}, fmt.Errorf("open interest is required")
	}

	resolved, err := resolveDerivativesTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return CanonicalOpenInterestSnapshotEvent{}, err
	}

	return CanonicalOpenInterestSnapshotEvent{
		SchemaVersion:           "v1",
		EventType:               "open-interest-snapshot",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		OpenInterest:            message.OpenInterest,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolved.Status,
		SourceRecordID:          derivativesSourceRecordID("oi", message.ExchangeTs, message.RecvTs),
		CanonicalEventTime:      resolved.EventTime,
		TimestampFallbackReason: resolved.FallbackReason,
	}, nil
}

func NormalizeLiquidationMessage(metadata DerivativesMetadata, message LiquidationMessage, policy TimestampPolicy) (CanonicalLiquidationPrintEvent, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return CanonicalLiquidationPrintEvent{}, fmt.Errorf("derivatives metadata is incomplete")
	}
	if message.Type != "liquidation-print" {
		return CanonicalLiquidationPrintEvent{}, fmt.Errorf("unsupported liquidation message type %q", message.Type)
	}
	if message.Side == "" || message.Price == "" || message.Size == "" {
		return CanonicalLiquidationPrintEvent{}, fmt.Errorf("liquidation side, price, and size are required")
	}

	resolved, err := resolveDerivativesTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return CanonicalLiquidationPrintEvent{}, err
	}

	return CanonicalLiquidationPrintEvent{
		SchemaVersion:           "v1",
		EventType:               "liquidation-print",
		Symbol:                  metadata.Symbol,
		SourceSymbol:            metadata.SourceSymbol,
		QuoteCurrency:           metadata.QuoteCurrency,
		Venue:                   metadata.Venue,
		MarketType:              metadata.MarketType,
		Side:                    message.Side,
		Price:                   message.Price,
		Size:                    message.Size,
		ExchangeTs:              message.ExchangeTs,
		RecvTs:                  message.RecvTs,
		TimestampStatus:         resolved.Status,
		SourceRecordID:          liquidationSourceRecordID(metadata.SourceSymbol, message),
		CanonicalEventTime:      resolved.EventTime,
		TimestampFallbackReason: resolved.FallbackReason,
	}, nil
}

func resolveDerivativesTimestamps(exchangeTimestamp, recvTimestamp string, policy TimestampPolicy) (CanonicalTimestamp, error) {
	return resolveMessageTimestamps(exchangeTimestamp, recvTimestamp, policy)
}

func derivativesSourceRecordID(prefix, exchangeTimestamp, recvTimestamp string) string {
	if exchangeTimestamp != "" {
		return prefix + ":" + exchangeTimestamp
	}
	return prefix + ":" + recvTimestamp
}

func liquidationSourceRecordID(sourceSymbol string, message LiquidationMessage) string {
	if message.LiquidationID != "" {
		return "liquidation:" + message.LiquidationID
	}
	parts := []string{"liquidation", sourceSymbol}
	if message.ExchangeTs != "" {
		parts = append(parts, message.ExchangeTs)
	} else {
		parts = append(parts, message.RecvTs)
	}
	parts = append(parts, strings.ToLower(message.Side), message.Price, message.Size)
	return strings.Join(parts, ":")
}
