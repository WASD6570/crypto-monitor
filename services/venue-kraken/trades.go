package venuekraken

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedTrade struct {
	SourceSymbol string
	Message      ingestion.TradeMessage
}

type tradePayload struct {
	Channel    string `json:"channel"`
	Pair       string `json:"pair"`
	TradeID    string `json:"tradeId"`
	Price      string `json:"price"`
	Volume     string `json:"volume"`
	Side       string `json:"side"`
	ExchangeTs string `json:"exchangeTs"`
}

func ParseTradeEvent(raw []byte, recvTime time.Time) (ParsedTrade, error) {
	if recvTime.IsZero() {
		return ParsedTrade{}, fmt.Errorf("recv time is required")
	}

	var payload tradePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTrade{}, fmt.Errorf("decode kraken trade payload: %w", err)
	}
	if payload.Channel != "trade" {
		return ParsedTrade{}, fmt.Errorf("unsupported kraken channel %q", payload.Channel)
	}
	if payload.Pair == "" {
		return ParsedTrade{}, fmt.Errorf("source symbol is required")
	}
	if payload.TradeID == "" {
		return ParsedTrade{}, fmt.Errorf("trade id is required")
	}
	if payload.Price == "" {
		return ParsedTrade{}, fmt.Errorf("price is required")
	}
	if payload.Volume == "" {
		return ParsedTrade{}, fmt.Errorf("volume is required")
	}
	if payload.Side != "buy" && payload.Side != "sell" {
		return ParsedTrade{}, fmt.Errorf("unsupported trade side %q", payload.Side)
	}
	if _, err := time.Parse(time.RFC3339Nano, payload.ExchangeTs); err != nil {
		return ParsedTrade{}, fmt.Errorf("parse exchange timestamp: %w", err)
	}

	return ParsedTrade{
		SourceSymbol: payload.Pair,
		Message: ingestion.TradeMessage{
			Type:       "trade",
			TradeID:    payload.TradeID,
			Price:      payload.Price,
			Size:       payload.Volume,
			Side:       payload.Side,
			ExchangeTs: payload.ExchangeTs,
			RecvTs:     recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}
