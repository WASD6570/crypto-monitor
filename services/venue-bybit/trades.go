package venuebybit

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
	Topic string      `json:"topic"`
	Type  string      `json:"type"`
	TS    int64       `json:"ts"`
	Data  []tradeData `json:"data"`
}

type tradeData struct {
	Symbol       string `json:"s"`
	TradeID      string `json:"i"`
	Side         string `json:"S"`
	Price        string `json:"p"`
	Size         string `json:"v"`
	TradeTimeMs  int64  `json:"T"`
	TimestampMs  int64  `json:"ts"`
	MarketType   string `json:"marketType"`
	ProductType  string `json:"productType"`
	IsBuyerMaker bool   `json:"m"`
}

func ParseTradeEvent(raw []byte, recvTime time.Time) (ParsedTrade, error) {
	if recvTime.IsZero() {
		return ParsedTrade{}, fmt.Errorf("recv time is required")
	}

	var payload tradePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTrade{}, fmt.Errorf("decode bybit trade payload: %w", err)
	}
	if payload.Topic == "" {
		return ParsedTrade{}, fmt.Errorf("trade topic is required")
	}
	if len(payload.Data) != 1 {
		return ParsedTrade{}, fmt.Errorf("expected exactly one bybit trade record")
	}

	record := payload.Data[0]
	if record.Symbol == "" {
		return ParsedTrade{}, fmt.Errorf("source symbol is required")
	}
	if record.TradeID == "" {
		return ParsedTrade{}, fmt.Errorf("trade id is required")
	}
	if record.Price == "" {
		return ParsedTrade{}, fmt.Errorf("price is required")
	}
	if record.Size == "" {
		return ParsedTrade{}, fmt.Errorf("size is required")
	}
	if record.Side != "Buy" && record.Side != "Sell" {
		return ParsedTrade{}, fmt.Errorf("unsupported bybit trade side %q", record.Side)
	}

	exchangeTimestamp, err := formatExchangeTimestamp(record.TradeTimeMs, record.TimestampMs, payload.TS)
	if err != nil {
		return ParsedTrade{}, err
	}

	return ParsedTrade{
		SourceSymbol: record.Symbol,
		Message: ingestion.TradeMessage{
			Type:       "trade",
			TradeID:    record.TradeID,
			Price:      record.Price,
			Size:       record.Size,
			Side:       normalizeTradeSide(record.Side),
			ExchangeTs: exchangeTimestamp,
			RecvTs:     recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func normalizeTradeSide(side string) string {
	if side == "Buy" {
		return "buy"
	}
	return "sell"
}

func formatExchangeTimestamp(values ...int64) (string, error) {
	for _, value := range values {
		if value <= 0 {
			continue
		}
		return time.UnixMilli(value).UTC().Format(time.RFC3339Nano), nil
	}
	return "", fmt.Errorf("trade timestamp is required")
}
