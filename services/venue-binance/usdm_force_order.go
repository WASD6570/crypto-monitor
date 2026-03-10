package venuebinance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedUSDMLiquidation struct {
	SourceSymbol string
	Message      ingestion.LiquidationMessage
}

type usdmForceOrderPayload struct {
	EventType   string `json:"e"`
	EventTimeMs int64  `json:"E"`
	Order       struct {
		SourceSymbol string `json:"s"`
		Side         string `json:"S"`
		Price        string `json:"p"`
		Size         string `json:"q"`
		TradeTimeMs  int64  `json:"T"`
	} `json:"o"`
}

func ParseUSDMForceOrder(raw []byte, recvTime time.Time) (ParsedUSDMLiquidation, error) {
	if recvTime.IsZero() {
		return ParsedUSDMLiquidation{}, fmt.Errorf("recv time is required")
	}

	var payload usdmForceOrderPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedUSDMLiquidation{}, fmt.Errorf("decode binance usdm force-order payload: %w", err)
	}
	if payload.EventType != "forceOrder" {
		return ParsedUSDMLiquidation{}, fmt.Errorf("unsupported binance usdm event type %q", payload.EventType)
	}
	if payload.Order.SourceSymbol == "" {
		return ParsedUSDMLiquidation{}, fmt.Errorf("source symbol is required")
	}
	if payload.Order.Side == "" || payload.Order.Price == "" || payload.Order.Size == "" {
		return ParsedUSDMLiquidation{}, fmt.Errorf("liquidation side, price, and size are required")
	}
	exchangeTimestamp, err := formatForceOrderTimestamp(payload.Order.TradeTimeMs, payload.EventTimeMs)
	if err != nil {
		return ParsedUSDMLiquidation{}, err
	}
	return ParsedUSDMLiquidation{
		SourceSymbol: payload.Order.SourceSymbol,
		Message: ingestion.LiquidationMessage{
			Type:       "liquidation-print",
			Side:       strings.ToLower(payload.Order.Side),
			Price:      payload.Order.Price,
			Size:       payload.Order.Size,
			ExchangeTs: exchangeTimestamp,
			RecvTs:     recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func formatForceOrderTimestamp(tradeTimeMs, eventTimeMs int64) (string, error) {
	selected := tradeTimeMs
	if selected <= 0 {
		selected = eventTimeMs
	}
	return formatUnixMilliTimestamp(selected)
}
