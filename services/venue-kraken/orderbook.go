package venuekraken

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedOrderBook struct {
	SourceSymbol string
	Message      ingestion.OrderBookMessage
}

type L2IntegrityState struct {
	sequencer ingestion.OrderBookSequencer
}

type orderBookPayload struct {
	Channel    string      `json:"channel"`
	Type       string      `json:"type"`
	Pair       string      `json:"pair"`
	Sequence   int64       `json:"sequence"`
	Bids       []bookLevel `json:"bids"`
	Asks       []bookLevel `json:"asks"`
	ExchangeTs string      `json:"exchangeTs"`
}

type bookLevel [2]string

func ParseOrderBookEvent(raw []byte, recvTime string) (ParsedOrderBook, error) {
	if recvTime == "" {
		return ParsedOrderBook{}, fmt.Errorf("recv time is required")
	}

	var payload orderBookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("decode kraken order-book payload: %w", err)
	}
	if payload.Channel != "book" {
		return ParsedOrderBook{}, fmt.Errorf("unsupported kraken channel %q", payload.Channel)
	}
	if payload.Type != "snapshot" && payload.Type != "delta" {
		return ParsedOrderBook{}, fmt.Errorf("unsupported kraken order-book type %q", payload.Type)
	}
	if payload.Pair == "" {
		return ParsedOrderBook{}, fmt.Errorf("source symbol is required")
	}
	if payload.Sequence <= 0 {
		return ParsedOrderBook{}, fmt.Errorf("sequence is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, payload.ExchangeTs); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("parse exchange timestamp: %w", err)
	}
	if _, err := time.Parse(time.RFC3339Nano, recvTime); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("parse recv timestamp: %w", err)
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
		SourceSymbol: payload.Pair,
		Message: ingestion.OrderBookMessage{
			Type:         payload.Type,
			Sequence:     payload.Sequence,
			BestBidPrice: bestBidPrice,
			BestAskPrice: bestAskPrice,
			ExchangeTs:   payload.ExchangeTs,
			RecvTs:       recvTime,
		},
	}, nil
}

func (s *L2IntegrityState) Normalize(metadata ingestion.BookMetadata, message ingestion.OrderBookMessage) (ingestion.OrderBookNormalizationResult, error) {
	if s == nil {
		return ingestion.OrderBookNormalizationResult{}, fmt.Errorf("l2 integrity state is required")
	}
	return ingestion.NormalizeOrderBookMessage(metadata, message, &s.sequencer, ingestion.StrictTimestampPolicy())
}

func (s *L2IntegrityState) ResyncRequired() bool {
	if s == nil {
		return false
	}
	return s.sequencer.ResyncRequired()
}

func (s *L2IntegrityState) Ready() bool {
	if s == nil {
		return false
	}
	return s.sequencer.Ready()
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
