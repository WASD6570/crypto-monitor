package venuebybit

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

type bookPayload struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type"`
	TS    int64           `json:"ts"`
	Data  bookPayloadData `json:"data"`
}

type bookPayloadData struct {
	Symbol   string      `json:"s"`
	Bids     []bookLevel `json:"b"`
	Asks     []bookLevel `json:"a"`
	UpdateID int64       `json:"u"`
	Seq      int64       `json:"seq"`
}

type bookLevel [2]string

func ParseOrderBookEvent(raw []byte, recvTime time.Time) (ParsedOrderBook, error) {
	if recvTime.IsZero() {
		return ParsedOrderBook{}, fmt.Errorf("recv time is required")
	}

	var payload bookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedOrderBook{}, fmt.Errorf("decode bybit order-book payload: %w", err)
	}
	if payload.Topic == "" {
		return ParsedOrderBook{}, fmt.Errorf("order-book topic is required")
	}
	if payload.Type != "snapshot" && payload.Type != "delta" {
		return ParsedOrderBook{}, fmt.Errorf("unsupported bybit order-book type %q", payload.Type)
	}
	if payload.Data.Symbol == "" {
		return ParsedOrderBook{}, fmt.Errorf("source symbol is required")
	}

	sequence := payload.Data.UpdateID
	if sequence <= 0 {
		sequence = payload.Data.Seq
	}
	if sequence <= 0 {
		return ParsedOrderBook{}, fmt.Errorf("sequence is required")
	}

	bestBidPrice, err := bestBookPrice(payload.Data.Bids, "bid")
	if err != nil {
		return ParsedOrderBook{}, err
	}
	bestAskPrice, err := bestBookPrice(payload.Data.Asks, "ask")
	if err != nil {
		return ParsedOrderBook{}, err
	}
	exchangeTimestamp, err := formatOrderBookTimestamp(payload.TS)
	if err != nil {
		return ParsedOrderBook{}, err
	}

	return ParsedOrderBook{
		SourceSymbol: payload.Data.Symbol,
		Message: ingestion.OrderBookMessage{
			Type:         payload.Type,
			Sequence:     sequence,
			BestBidPrice: bestBidPrice,
			BestAskPrice: bestAskPrice,
			ExchangeTs:   exchangeTimestamp,
			RecvTs:       recvTime.UTC().Format(time.RFC3339Nano),
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

func formatOrderBookTimestamp(ts int64) (string, error) {
	if ts <= 0 {
		return "", fmt.Errorf("order-book timestamp is required")
	}
	return time.UnixMilli(ts).UTC().Format(time.RFC3339Nano), nil
}
