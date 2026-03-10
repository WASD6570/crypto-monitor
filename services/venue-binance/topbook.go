package venuebinance

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
	UpdateID     int64  `json:"u"`
	EventTimeMs  int64  `json:"E"`
	SourceSymbol string `json:"s"`
	BestBidPrice string `json:"b"`
	BestBidSize  string `json:"B"`
	BestAskPrice string `json:"a"`
	BestAskSize  string `json:"A"`
}

func ParseTopOfBookEvent(raw []byte, recvTime time.Time) (ParsedTopOfBook, error) {
	if recvTime.IsZero() {
		return ParsedTopOfBook{}, fmt.Errorf("recv time is required")
	}

	var payload topOfBookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedTopOfBook{}, fmt.Errorf("decode binance top-of-book payload: %w", err)
	}
	if payload.SourceSymbol == "" {
		return ParsedTopOfBook{}, fmt.Errorf("source symbol is required")
	}
	if payload.UpdateID <= 0 {
		return ParsedTopOfBook{}, fmt.Errorf("update id must be positive")
	}
	if payload.BestBidPrice == "" {
		return ParsedTopOfBook{}, fmt.Errorf("best bid price is required")
	}
	if payload.BestAskPrice == "" {
		return ParsedTopOfBook{}, fmt.Errorf("best ask price is required")
	}

	exchangeTimestamp := ""
	if payload.EventTimeMs > 0 {
		var err error
		exchangeTimestamp, err = formatUnixMilliTimestamp(payload.EventTimeMs)
		if err != nil {
			return ParsedTopOfBook{}, err
		}
	}

	return ParsedTopOfBook{
		SourceSymbol: payload.SourceSymbol,
		Message: ingestion.OrderBookMessage{
			Type:         string(ingestion.BookUpdateTopOfBook),
			Sequence:     payload.UpdateID,
			BestBidPrice: payload.BestBidPrice,
			BestAskPrice: payload.BestAskPrice,
			ExchangeTs:   exchangeTimestamp,
			RecvTs:       recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func ParseTopOfBookFrame(frame SpotRawFrame) (ParsedTopOfBook, error) {
	return ParseTopOfBookEvent(frame.Payload, frame.RecvTime)
}
