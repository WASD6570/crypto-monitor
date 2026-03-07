package venuecoinbase

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
	Type         string `json:"type"`
	ProductID    string `json:"product_id"`
	SourceSymbol string `json:"sourceSymbol"`
	TradeID      int64  `json:"trade_id"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	Side         string `json:"side"`
	Time         string `json:"time"`
	ExchangeTs   string `json:"exchangeTs"`
}

func ParseTradeEvent(raw []byte, recvTime time.Time) (ParsedTrade, error) {
	if recvTime.IsZero() {
		return ParsedTrade{}, fmt.Errorf("recv time is required")
	}

	var payload tradePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTrade{}, fmt.Errorf("decode coinbase trade payload: %w", err)
	}
	if payload.Type != "match" {
		return ParsedTrade{}, fmt.Errorf("unsupported coinbase event type %q", payload.Type)
	}

	sourceSymbol := payload.ProductID
	if sourceSymbol == "" {
		sourceSymbol = payload.SourceSymbol
	}
	if sourceSymbol == "" {
		return ParsedTrade{}, fmt.Errorf("source symbol is required")
	}
	if payload.TradeID <= 0 {
		return ParsedTrade{}, fmt.Errorf("trade id must be positive")
	}
	if payload.Price == "" {
		return ParsedTrade{}, fmt.Errorf("price is required")
	}
	if payload.Size == "" {
		return ParsedTrade{}, fmt.Errorf("size is required")
	}
	if payload.Side != "buy" && payload.Side != "sell" {
		return ParsedTrade{}, fmt.Errorf("unsupported trade side %q", payload.Side)
	}

	exchangeTimestamp, err := parseExchangeTimestamp(payload.Time, payload.ExchangeTs)
	if err != nil {
		return ParsedTrade{}, err
	}

	return ParsedTrade{
		SourceSymbol: sourceSymbol,
		Message: ingestion.TradeMessage{
			Type:       "trade",
			TradeID:    fmt.Sprintf("%d", payload.TradeID),
			Price:      payload.Price,
			Size:       payload.Size,
			Side:       payload.Side,
			ExchangeTs: exchangeTimestamp,
			RecvTs:     recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func parseExchangeTimestamp(values ...string) (string, error) {
	for _, value := range values {
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return "", fmt.Errorf("parse exchange timestamp %q: %w", value, err)
		}
		return parsed.UTC().Format(time.RFC3339Nano), nil
	}
	return "", fmt.Errorf("exchange timestamp is required")
}
