package ingestion

import (
	"fmt"
	"time"
)

type BookMetadata struct {
	Symbol        string
	SourceSymbol  string
	QuoteCurrency string
	Venue         Venue
	MarketType    string
}

type OrderBookMessage struct {
	Type         string `json:"type"`
	Sequence     int64  `json:"sequence"`
	BestBidPrice string `json:"bestBidPrice"`
	BestAskPrice string `json:"bestAskPrice"`
	ExchangeTs   string `json:"exchangeTs"`
	RecvTs       string `json:"recvTs"`
}

type CanonicalOrderBookEvent struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	EventType               string                   `json:"eventType"`
	BookAction              string                   `json:"bookAction"`
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

type CanonicalFeedHealthEvent struct {
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
	FeedHealthState         FeedHealthState          `json:"feedHealthState"`
	DegradationReasons      []DegradationReason      `json:"degradationReasons,omitempty"`
	SourceRecordID          string                   `json:"sourceRecordId"`
	CanonicalEventTime      time.Time                `json:"-"`
	TimestampFallbackReason TimestampFallbackReason  `json:"-"`
}

type OrderBookNormalizationResult struct {
	SequenceResult  SequenceResult
	OrderBookEvent  *CanonicalOrderBookEvent
	FeedHealthEvent *CanonicalFeedHealthEvent
}

func NormalizeOrderBookMessage(metadata BookMetadata, message OrderBookMessage, sequencer *OrderBookSequencer, policy TimestampPolicy) (OrderBookNormalizationResult, error) {
	if metadata.Symbol == "" || metadata.SourceSymbol == "" || metadata.QuoteCurrency == "" || metadata.Venue == "" || metadata.MarketType == "" {
		return OrderBookNormalizationResult{}, fmt.Errorf("book metadata is incomplete")
	}
	if sequencer == nil {
		return OrderBookNormalizationResult{}, fmt.Errorf("order book sequencer is required")
	}

	kind := BookUpdateKind(message.Type)
	if kind != BookUpdateSnapshot && kind != BookUpdateDelta && kind != BookUpdateTopOfBook {
		return OrderBookNormalizationResult{}, fmt.Errorf("unsupported book message type %q", message.Type)
	}

	resolvedTs, err := resolveMessageTimestamps(message.ExchangeTs, message.RecvTs, policy)
	if err != nil {
		return OrderBookNormalizationResult{}, err
	}

	if kind == BookUpdateTopOfBook {
		return OrderBookNormalizationResult{
			SequenceResult: SequenceResult{Action: SequenceAcceptedTopOfBook},
			OrderBookEvent: &CanonicalOrderBookEvent{
				SchemaVersion:           "v1",
				EventType:               "order-book-top",
				BookAction:              string(kind),
				Symbol:                  metadata.Symbol,
				SourceSymbol:            metadata.SourceSymbol,
				QuoteCurrency:           metadata.QuoteCurrency,
				Venue:                   metadata.Venue,
				MarketType:              metadata.MarketType,
				ExchangeTs:              message.ExchangeTs,
				RecvTs:                  message.RecvTs,
				TimestampStatus:         resolvedTs.Status,
				SourceRecordID:          topOfBookSourceRecordID(message),
				CanonicalEventTime:      resolvedTs.EventTime,
				TimestampFallbackReason: resolvedTs.FallbackReason,
			},
		}, nil
	}

	sequenceResult, err := sequencer.Apply(SequencedBookUpdate{Kind: kind, Sequence: message.Sequence})
	if err != nil {
		return OrderBookNormalizationResult{}, err
	}

	result := OrderBookNormalizationResult{SequenceResult: sequenceResult}
	if sequenceResult.Action == SequenceAcceptedSnapshot || sequenceResult.Action == SequenceAcceptedDelta {
		result.OrderBookEvent = &CanonicalOrderBookEvent{
			SchemaVersion:           "v1",
			EventType:               "order-book-top",
			BookAction:              message.Type,
			Symbol:                  metadata.Symbol,
			SourceSymbol:            metadata.SourceSymbol,
			QuoteCurrency:           metadata.QuoteCurrency,
			Venue:                   metadata.Venue,
			MarketType:              metadata.MarketType,
			ExchangeTs:              message.ExchangeTs,
			RecvTs:                  message.RecvTs,
			TimestampStatus:         resolvedTs.Status,
			SourceRecordID:          fmt.Sprintf("book:%d", message.Sequence),
			CanonicalEventTime:      resolvedTs.EventTime,
			TimestampFallbackReason: resolvedTs.FallbackReason,
		}
		return result, nil
	}

	if sequenceResult.Action == SequenceRequiresResync {
		result.FeedHealthEvent = &CanonicalFeedHealthEvent{
			SchemaVersion:           "v1",
			EventType:               "feed-health",
			Symbol:                  metadata.Symbol,
			SourceSymbol:            metadata.SourceSymbol,
			QuoteCurrency:           metadata.QuoteCurrency,
			Venue:                   metadata.Venue,
			MarketType:              metadata.MarketType,
			ExchangeTs:              message.ExchangeTs,
			RecvTs:                  message.RecvTs,
			TimestampStatus:         resolvedTs.Status,
			FeedHealthState:         FeedHealthDegraded,
			DegradationReasons:      []DegradationReason{ReasonSequenceGap},
			SourceRecordID:          gapSourceRecordID(sequenceResult, message.Sequence),
			CanonicalEventTime:      resolvedTs.EventTime,
			TimestampFallbackReason: resolvedTs.FallbackReason,
		}
	}

	return result, nil
}

func resolveMessageTimestamps(exchangeTimestamp, recvTimestamp string, policy TimestampPolicy) (CanonicalTimestamp, error) {
	recvTime, err := time.Parse(time.RFC3339Nano, recvTimestamp)
	if err != nil {
		return CanonicalTimestamp{}, fmt.Errorf("parse recv timestamp: %w", err)
	}

	var exchangeTime time.Time
	if exchangeTimestamp != "" {
		exchangeTime, err = time.Parse(time.RFC3339Nano, exchangeTimestamp)
		if err != nil {
			exchangeTime = time.Time{}
		}
	}

	return ResolveCanonicalTimestamp(exchangeTime, recvTime, policy)
}

func gapSourceRecordID(sequenceResult SequenceResult, sequence int64) string {
	if sequenceResult.LastSequence > 0 {
		return fmt.Sprintf("gap:%d-%d", sequenceResult.LastSequence, sequence)
	}
	return fmt.Sprintf("gap:unknown-%d", sequence)
}

func topOfBookSourceRecordID(message OrderBookMessage) string {
	timestamp := message.ExchangeTs
	if timestamp == "" {
		timestamp = message.RecvTs
	}
	return "ticker:" + timestamp
}
