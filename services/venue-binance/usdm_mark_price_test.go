package venuebinance

import (
	"testing"
	"time"
)

func TestUSDMMarkPriceParsesFundingAndMarkIndexCandidates(t *testing.T) {
	recvTime := time.UnixMilli(1772798405040).UTC()
	parsed, err := ParseUSDMMarkPrice([]byte(`{"e":"markPriceUpdate","E":1772798405000,"s":"BTCUSDT","p":"11794.15000000","i":"11784.62659091","r":"0.00038167","T":1772812800000}`), recvTime)
	if err != nil {
		t.Fatalf("parse usdm mark price: %v", err)
	}
	if parsed.SourceSymbol != "BTCUSDT" {
		t.Fatalf("source symbol = %q", parsed.SourceSymbol)
	}
	if parsed.Funding.Type != "funding-rate" || parsed.MarkIndex.Type != "mark-index" {
		t.Fatalf("unexpected parsed messages: %+v", parsed)
	}
	if parsed.Funding.ExchangeTs != "2026-03-06T12:00:05Z" {
		t.Fatalf("funding exchangeTs = %q", parsed.Funding.ExchangeTs)
	}
	if parsed.MarkIndex.MarkPrice != "11794.15000000" || parsed.MarkIndex.IndexPrice != "11784.62659091" {
		t.Fatalf("unexpected mark/index values: %+v", parsed.MarkIndex)
	}
	if parsed.Funding.NextFundingTs != "2026-03-06T16:00:00Z" {
		t.Fatalf("next funding ts = %q", parsed.Funding.NextFundingTs)
	}
}

func TestUSDMMarkPriceRejectsMalformedPayload(t *testing.T) {
	_, err := ParseUSDMMarkPrice([]byte(`{"e":"markPriceUpdate","E":1772798405000,"s":"BTCUSDT","p":"11794.15000000","i":"11784.62659091"}`), time.UnixMilli(1772798405040).UTC())
	if err == nil {
		t.Fatal("expected missing funding rate to fail")
	}
}
