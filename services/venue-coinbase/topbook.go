package venuecoinbase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedTopOfBook struct {
	SourceSymbol string
	Message      ingestion.OrderBookMessage
}

type topOfBookPayload struct {
	Type         string `json:"type"`
	ProductID    string `json:"product_id"`
	SourceSymbol string `json:"sourceSymbol"`
	Sequence     int64  `json:"sequence"`
	BestBid      string `json:"best_bid"`
	BestBidPrice string `json:"bestBidPrice"`
	BestAsk      string `json:"best_ask"`
	BestAskPrice string `json:"bestAskPrice"`
	Time         string `json:"time"`
	ExchangeTs   string `json:"exchangeTs"`
}

func ParseTopOfBookEvent(raw []byte, recvTime time.Time) (ParsedTopOfBook, error) {
	if recvTime.IsZero() {
		return ParsedTopOfBook{}, fmt.Errorf("recv time is required")
	}

	var payload topOfBookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTopOfBook{}, fmt.Errorf("decode coinbase top-of-book payload: %w", err)
	}
	if payload.Type != "ticker" {
		return ParsedTopOfBook{}, fmt.Errorf("unsupported coinbase event type %q", payload.Type)
	}

	sourceSymbol := payload.ProductID
	if sourceSymbol == "" {
		sourceSymbol = payload.SourceSymbol
	}
	if sourceSymbol == "" {
		return ParsedTopOfBook{}, fmt.Errorf("source symbol is required")
	}
	bestBidPrice := payload.BestBid
	if bestBidPrice == "" {
		bestBidPrice = payload.BestBidPrice
	}
	if bestBidPrice == "" {
		return ParsedTopOfBook{}, fmt.Errorf("best bid price is required")
	}
	bestAskPrice := payload.BestAsk
	if bestAskPrice == "" {
		bestAskPrice = payload.BestAskPrice
	}
	if bestAskPrice == "" {
		return ParsedTopOfBook{}, fmt.Errorf("best ask price is required")
	}

	exchangeTimestamp, err := parseExchangeTimestamp(payload.Time, payload.ExchangeTs)
	if err != nil {
		return ParsedTopOfBook{}, err
	}

	return ParsedTopOfBook{
		SourceSymbol: sourceSymbol,
		Message: ingestion.OrderBookMessage{
			Type:         string(ingestion.BookUpdateTopOfBook),
			Sequence:     payload.Sequence,
			BestBidPrice: bestBidPrice,
			BestAskPrice: bestAskPrice,
			ExchangeTs:   exchangeTimestamp,
			RecvTs:       recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}
