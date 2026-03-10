package venuebinance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ParsedUSDMMarkPrice struct {
	SourceSymbol string
	Funding      ingestion.FundingRateMessage
	MarkIndex    ingestion.MarkIndexMessage
}

type usdmMarkPricePayload struct {
	EventType       string `json:"e"`
	EventTimeMs     int64  `json:"E"`
	SourceSymbol    string `json:"s"`
	MarkPrice       string `json:"p"`
	IndexPrice      string `json:"i"`
	EstimatedSettle string `json:"P"`
	FundingRate     string `json:"r"`
	NextFundingMs   int64  `json:"T"`
}

func ParseUSDMMarkPrice(raw []byte, recvTime time.Time) (ParsedUSDMMarkPrice, error) {
	if recvTime.IsZero() {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("recv time is required")
	}

	var payload usdmMarkPricePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("decode binance usdm mark price payload: %w", err)
	}
	if payload.EventType != "markPriceUpdate" {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("unsupported binance usdm event type %q", payload.EventType)
	}
	if payload.SourceSymbol == "" {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("source symbol is required")
	}
	if payload.MarkPrice == "" || payload.IndexPrice == "" {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("mark and index prices are required")
	}
	if payload.FundingRate == "" {
		return ParsedUSDMMarkPrice{}, fmt.Errorf("funding rate is required")
	}

	exchangeTimestamp, err := formatUnixMilliTimestamp(payload.EventTimeMs)
	if err != nil {
		return ParsedUSDMMarkPrice{}, err
	}
	nextFundingTimestamp := ""
	if payload.NextFundingMs > 0 {
		nextFundingTimestamp = time.UnixMilli(payload.NextFundingMs).UTC().Format(time.RFC3339Nano)
	}
	recvTimestamp := recvTime.UTC().Format(time.RFC3339Nano)

	return ParsedUSDMMarkPrice{
		SourceSymbol: payload.SourceSymbol,
		Funding: ingestion.FundingRateMessage{
			Type:          "funding-rate",
			FundingRate:   payload.FundingRate,
			NextFundingTs: nextFundingTimestamp,
			ExchangeTs:    exchangeTimestamp,
			RecvTs:        recvTimestamp,
		},
		MarkIndex: ingestion.MarkIndexMessage{
			Type:       "mark-index",
			MarkPrice:  payload.MarkPrice,
			IndexPrice: payload.IndexPrice,
			ExchangeTs: exchangeTimestamp,
			RecvTs:     recvTimestamp,
		},
	}, nil
}
