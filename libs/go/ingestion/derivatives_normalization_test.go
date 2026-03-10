package ingestion

import (
	"testing"
	"time"
)

func TestNormalizeFunding(t *testing.T) {
	actual, err := NormalizeFundingMessage(derivativesMetadata(), FundingRateMessage{
		Type:          "funding-rate",
		FundingRate:   "0.00038167",
		NextFundingTs: "2026-03-06T16:00:00Z",
		ExchangeTs:    "2026-03-06T12:00:05Z",
		RecvTs:        "2026-03-06T12:00:05.040Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize funding: %v", err)
	}
	if actual.EventType != "funding-rate" || actual.Symbol != "BTC-USD" {
		t.Fatalf("unexpected funding event: %+v", actual)
	}
	if actual.TimestampStatus != TimestampStatusNormal {
		t.Fatalf("timestamp status = %q, want %q", actual.TimestampStatus, TimestampStatusNormal)
	}
	if actual.SourceRecordID != "funding:2026-03-06T12:00:05Z" {
		t.Fatalf("sourceRecordId = %q", actual.SourceRecordID)
	}
}

func TestNormalizeMarkIndex(t *testing.T) {
	actual, err := NormalizeMarkIndexMessage(derivativesMetadata(), MarkIndexMessage{
		Type:       "mark-index",
		MarkPrice:  "11794.15000000",
		IndexPrice: "11784.62659091",
		ExchangeTs: "2026-03-06T12:00:05Z",
		RecvTs:     "2026-03-06T12:00:05.040Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize mark index: %v", err)
	}
	if actual.EventType != "mark-index" || actual.MarkPrice == "" || actual.IndexPrice == "" {
		t.Fatalf("unexpected mark index event: %+v", actual)
	}
	if actual.SourceRecordID != "mark-index:2026-03-06T12:00:05Z" {
		t.Fatalf("sourceRecordId = %q", actual.SourceRecordID)
	}
}

func TestNormalizeOpenInterest(t *testing.T) {
	actual, err := NormalizeOpenInterestMessage(derivativesMetadata(), OpenInterestMessage{
		Type:         "open-interest",
		OpenInterest: "10659.509",
		ExchangeTs:   "2026-03-06T12:00:04Z",
		RecvTs:       "2026-03-06T12:00:04.060Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize open interest: %v", err)
	}
	if actual.EventType != "open-interest-snapshot" || actual.OpenInterest != "10659.509" {
		t.Fatalf("unexpected open interest event: %+v", actual)
	}
	if actual.SourceRecordID != "oi:2026-03-06T12:00:04Z" {
		t.Fatalf("sourceRecordId = %q", actual.SourceRecordID)
	}
}

func TestNormalizeLiquidation(t *testing.T) {
	actual, err := NormalizeLiquidationMessage(derivativesMetadata(), LiquidationMessage{
		Type:          "liquidation-print",
		LiquidationID: "liq-1",
		Side:          "sell",
		Price:         "9910",
		Size:          "0.014",
		ExchangeTs:    "2026-03-06T12:00:06Z",
		RecvTs:        "2026-03-06T12:00:06.020Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize liquidation: %v", err)
	}
	if actual.EventType != "liquidation-print" || actual.Side != "sell" {
		t.Fatalf("unexpected liquidation event: %+v", actual)
	}
	if actual.SourceRecordID != "liquidation:liq-1" {
		t.Fatalf("sourceRecordId = %q", actual.SourceRecordID)
	}
}

func TestNormalizeFundingFallsBackWhenExchangeTimestampSkews(t *testing.T) {
	actual, err := NormalizeFundingMessage(derivativesMetadata(), FundingRateMessage{
		Type:        "funding-rate",
		FundingRate: "0.00038167",
		ExchangeTs:  "2026-03-06T12:00:00Z",
		RecvTs:      "2026-03-06T12:00:05.040Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize funding: %v", err)
	}
	if actual.TimestampStatus != TimestampStatusDegraded {
		t.Fatalf("timestamp status = %q, want %q", actual.TimestampStatus, TimestampStatusDegraded)
	}
	if actual.TimestampFallbackReason != TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, TimestampReasonExchangeSkewExceeded)
	}
	if !actual.CanonicalEventTime.Equal(mustDerivativesTime(t, "2026-03-06T12:00:05.040Z")) {
		t.Fatalf("canonical event time = %s", actual.CanonicalEventTime)
	}
}

func TestNormalizeOpenInterestFallsBackWhenExchangeTimestampMissing(t *testing.T) {
	actual, err := NormalizeOpenInterestMessage(derivativesMetadata(), OpenInterestMessage{
		Type:         "open-interest",
		OpenInterest: "10659.509",
		RecvTs:       "2026-03-06T12:00:04.060Z",
	}, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize open interest: %v", err)
	}
	if actual.TimestampStatus != TimestampStatusDegraded {
		t.Fatalf("timestamp status = %q, want %q", actual.TimestampStatus, TimestampStatusDegraded)
	}
	if actual.TimestampFallbackReason != TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", actual.TimestampFallbackReason, TimestampReasonExchangeMissingOrInvalid)
	}
	if actual.SourceRecordID != "oi:2026-03-06T12:00:04.060Z" {
		t.Fatalf("sourceRecordId = %q", actual.SourceRecordID)
	}
}

func TestNormalizeLiquidationUsesDeterministicDerivedIDWithoutLiquidationID(t *testing.T) {
	message := LiquidationMessage{
		Type:       "liquidation-print",
		Side:       "sell",
		Price:      "9910",
		Size:       "0.014",
		ExchangeTs: "2026-03-06T12:00:06Z",
		RecvTs:     "2026-03-06T12:00:06.020Z",
	}
	first, err := NormalizeLiquidationMessage(derivativesMetadata(), message, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize first liquidation: %v", err)
	}
	second, err := NormalizeLiquidationMessage(derivativesMetadata(), message, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("normalize second liquidation: %v", err)
	}
	if first.SourceRecordID != second.SourceRecordID {
		t.Fatalf("sourceRecordId mismatch: %q vs %q", first.SourceRecordID, second.SourceRecordID)
	}
}

func derivativesMetadata() DerivativesMetadata {
	return DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "perpetual"}
}

func mustDerivativesTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}
