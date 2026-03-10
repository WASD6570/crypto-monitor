package venuebinance

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
	EventType    string `json:"e"`
	EventTimeMs  int64  `json:"E"`
	SourceSymbol string `json:"s"`
	TradeID      int64  `json:"t"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	TradeTimeMs  int64  `json:"T"`
	BuyerMaker   bool   `json:"m"`
	Ignore       bool   `json:"M"`
}

func ParseTradeEvent(raw []byte, recvTime time.Time) (ParsedTrade, error) {
	if recvTime.IsZero() {
		return ParsedTrade{}, fmt.Errorf("recv time is required")
	}

	var payload tradePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTrade{}, fmt.Errorf("decode binance trade payload: %w", err)
	}
	if payload.EventType != "trade" {
		return ParsedTrade{}, fmt.Errorf("unsupported binance event type %q", payload.EventType)
	}
	if payload.SourceSymbol == "" {
		return ParsedTrade{}, fmt.Errorf("source symbol is required")
	}
	if payload.TradeID <= 0 {
		return ParsedTrade{}, fmt.Errorf("trade id must be positive")
	}

	exchangeTimestamp, err := formatExchangeTimestamp(payload.TradeTimeMs, payload.EventTimeMs)
	if err != nil {
		return ParsedTrade{}, err
	}

	return ParsedTrade{
		SourceSymbol: payload.SourceSymbol,
		Message: ingestion.TradeMessage{
			Type:       "trade",
			TradeID:    fmt.Sprintf("%d", payload.TradeID),
			Price:      payload.Price,
			Size:       payload.Quantity,
			Side:       deriveAggressorSide(payload.BuyerMaker),
			ExchangeTs: exchangeTimestamp,
			RecvTs:     recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func ParseTradeFrame(frame SpotRawFrame) (ParsedTrade, error) {
	return ParseTradeEvent(frame.Payload, frame.RecvTime)
}

func deriveAggressorSide(buyerMaker bool) string {
	if buyerMaker {
		return "sell"
	}
	return "buy"
}

func formatExchangeTimestamp(tradeTimeMs, eventTimeMs int64) (string, error) {
	selected := tradeTimeMs
	if selected <= 0 {
		selected = eventTimeMs
	}
	if selected <= 0 {
		return "", fmt.Errorf("trade time or event time must be present")
	}
	return time.UnixMilli(selected).UTC().Format("2006-01-02T15:04:05.000Z07:00"), nil
}
