package ingestion

import (
	"testing"
	"time"
)

func TestRawWriteBoundaryAppendsImmutableEntries(t *testing.T) {
	writer := NewInMemoryRawEventWriter()
	first, err := BuildRawAppendEntryFromTrade(
		CanonicalTradeEvent{
			SchemaVersion:      "v1",
			EventType:          "market-trade",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTC-USD",
			QuoteCurrency:      "USD",
			Venue:              VenueCoinbase,
			MarketType:         "spot",
			ExchangeTs:         "2026-03-06T12:00:00Z",
			RecvTs:             "2026-03-06T12:00:00.500Z",
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "trade:first",
			CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:00Z"),
		},
		TradeMetadata{Symbol: "BTC-USD", SourceSymbol: "BTC-USD", QuoteCurrency: "USD", Venue: VenueCoinbase, MarketType: "spot"},
		TradeMessage{TradeID: "first", ExchangeTs: "2026-03-06T12:00:00Z", RecvTs: "2026-03-06T12:00:00.500Z"},
		RawWriteContext{ConnectionRef: "conn-1", SessionRef: "sess-1"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build first entry: %v", err)
	}
	if err := writer.Append(first); err != nil {
		t.Fatalf("append first entry: %v", err)
	}

	second, err := BuildRawAppendEntryFromTrade(
		CanonicalTradeEvent{
			SchemaVersion:      "v1",
			EventType:          "market-trade",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTC-USD",
			QuoteCurrency:      "USD",
			Venue:              VenueCoinbase,
			MarketType:         "spot",
			ExchangeTs:         "2026-03-06T12:00:01Z",
			RecvTs:             "2026-03-06T12:00:01.500Z",
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "trade:second",
			CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:01Z"),
		},
		TradeMetadata{Symbol: "BTC-USD", SourceSymbol: "BTC-USD", QuoteCurrency: "USD", Venue: VenueCoinbase, MarketType: "spot"},
		TradeMessage{TradeID: "second", ExchangeTs: "2026-03-06T12:00:01Z", RecvTs: "2026-03-06T12:00:01.500Z"},
		RawWriteContext{ConnectionRef: "conn-1", SessionRef: "sess-1"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build second entry: %v", err)
	}
	if err := writer.Append(second); err != nil {
		t.Fatalf("append second entry: %v", err)
	}

	entries := writer.Entries()
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].CanonicalEventID != first.CanonicalEventID {
		t.Fatalf("first canonicalEventId = %q, want %q", entries[0].CanonicalEventID, first.CanonicalEventID)
	}
	if entries[0].DuplicateAudit.Occurrence != 1 {
		t.Fatalf("first duplicate occurrence = %d, want 1", entries[0].DuplicateAudit.Occurrence)
	}
	if entries[1].CanonicalEventID != second.CanonicalEventID {
		t.Fatalf("second canonicalEventId = %q, want %q", entries[1].CanonicalEventID, second.CanonicalEventID)
	}
	if entries[0].RecvTs != "2026-03-06T12:00:00.500Z" {
		t.Fatalf("first recvTs mutated to %q", entries[0].RecvTs)
	}
	if entries[0].PartitionKey.String() != "2026-03-06/BTC-USD/COINBASE" {
		t.Fatalf("first partition key = %q, want %q", entries[0].PartitionKey.String(), "2026-03-06/BTC-USD/COINBASE")
	}
}

func TestRawWriteBoundaryPersistsTimestampProvenance(t *testing.T) {
	entry, err := BuildRawAppendEntryFromTrade(
		CanonicalTradeEvent{
			SchemaVersion:           "v1",
			EventType:               "market-trade",
			Symbol:                  "ETH-USD",
			SourceSymbol:            "ETHUSDT",
			QuoteCurrency:           "USDT",
			Venue:                   VenueBinance,
			MarketType:              "spot",
			ExchangeTs:              "",
			RecvTs:                  "2026-03-06T23:59:59.900Z",
			TimestampStatus:         TimestampStatusDegraded,
			SourceRecordID:          "trade:degraded",
			CanonicalEventTime:      mustRawTime(t, "2026-03-06T23:59:59.900Z"),
			TimestampFallbackReason: TimestampReasonExchangeMissingOrInvalid,
		},
		TradeMetadata{Symbol: "ETH-USD", SourceSymbol: "ETHUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "spot"},
		TradeMessage{TradeID: "degraded", RecvTs: "2026-03-06T23:59:59.900Z"},
		RawWriteContext{ConnectionRef: "conn-2", SessionRef: "sess-2"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if entry.ExchangeTs != "" {
		t.Fatalf("exchangeTs = %q, want empty", entry.ExchangeTs)
	}
	if entry.RecvTs != "2026-03-06T23:59:59.900Z" {
		t.Fatalf("recvTs = %q, want recv timestamp", entry.RecvTs)
	}
	if entry.BucketTimestampSource != RawBucketTimestampSourceRecv {
		t.Fatalf("bucketTimestampSource = %q, want %q", entry.BucketTimestampSource, RawBucketTimestampSourceRecv)
	}
	if entry.BucketTimestamp != entry.RecvTs {
		t.Fatalf("bucketTimestamp = %q, want recv timestamp", entry.BucketTimestamp)
	}
	if entry.TimestampDegradationReason != TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("timestampDegradationReason = %q, want %q", entry.TimestampDegradationReason, TimestampReasonExchangeMissingOrInvalid)
	}
}

func TestRawWriteBoundaryPersistsOpenInterestStreamFamily(t *testing.T) {
	entry, err := BuildRawAppendEntryFromOpenInterest(
		CanonicalOpenInterestSnapshotEvent{
			SchemaVersion:      "v1",
			EventType:          "open-interest-snapshot",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              VenueBinance,
			MarketType:         "perpetual",
			OpenInterest:       "10659.509",
			ExchangeTs:         "2026-03-06T12:00:04Z",
			RecvTs:             "2026-03-06T12:00:04.060Z",
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "oi:2026-03-06T12:00:04Z",
			CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:04Z"),
		},
		DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "perpetual"},
		OpenInterestMessage{Type: "open-interest", OpenInterest: "10659.509", ExchangeTs: "2026-03-06T12:00:04Z", RecvTs: "2026-03-06T12:00:04.060Z"},
		RawWriteContext{ConnectionRef: "conn-oi", SessionRef: "sess-oi"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build open interest entry: %v", err)
	}
	if entry.StreamFamily != string(StreamOpenInterest) {
		t.Fatalf("stream family = %q, want %q", entry.StreamFamily, StreamOpenInterest)
	}
	if entry.PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/open-interest" {
		t.Fatalf("partition key = %q, want %q", entry.PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/open-interest")
	}
	if entry.VenueMessageID != "2026-03-06T12:00:04Z" {
		t.Fatalf("venueMessageId = %q, want %q", entry.VenueMessageID, "2026-03-06T12:00:04Z")
	}
}

func TestRawWriteBoundaryRecordsDuplicateAuditFacts(t *testing.T) {
	writer := NewInMemoryRawEventWriter()
	for i := 0; i < 2; i++ {
		entry, err := BuildRawAppendEntryFromTrade(
			CanonicalTradeEvent{
				SchemaVersion:      "v1",
				EventType:          "market-trade",
				Symbol:             "BTC-USD",
				SourceSymbol:       "BTCUSDT",
				QuoteCurrency:      "USDT",
				Venue:              VenueBinance,
				MarketType:         "spot",
				ExchangeTs:         "2026-03-06T12:00:00Z",
				RecvTs:             "2026-03-06T12:00:00.100Z",
				TimestampStatus:    TimestampStatusNormal,
				SourceRecordID:     "trade:dup",
				CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:00Z"),
			},
			TradeMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "spot"},
			TradeMessage{TradeID: "dup-trade", ExchangeTs: "2026-03-06T12:00:00Z", RecvTs: "2026-03-06T12:00:00.100Z"},
			RawWriteContext{ConnectionRef: "conn-dup", SessionRef: "sess-dup"},
			RawWriteOptions{BuildVersion: "test-build"},
		)
		if err != nil {
			t.Fatalf("build duplicate entry %d: %v", i, err)
		}
		if err := writer.Append(entry); err != nil {
			t.Fatalf("append duplicate entry %d: %v", i, err)
		}
	}

	entries := writer.Entries()
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].DuplicateAudit.Duplicate {
		t.Fatal("first append should not be marked duplicate")
	}
	if !entries[1].DuplicateAudit.Duplicate {
		t.Fatal("second append should be marked duplicate")
	}
	if entries[1].DuplicateAudit.Occurrence != 2 {
		t.Fatalf("second duplicate occurrence = %d, want 2", entries[1].DuplicateAudit.Occurrence)
	}
	if entries[0].DuplicateAudit.IdentityKey != "message:trades:dup-trade" {
		t.Fatalf("identity key = %q, want %q", entries[0].DuplicateAudit.IdentityKey, "message:trades:dup-trade")
	}
	if entries[0].PartitionKey != entries[1].PartitionKey {
		t.Fatalf("duplicate partitions differ: %+v vs %+v", entries[0].PartitionKey, entries[1].PartitionKey)
	}
}

func TestBuildRawAppendEntryFromFeedHealthRetainsDegradedFeedReference(t *testing.T) {
	entry, err := BuildRawAppendEntryFromFeedHealth(
		CanonicalFeedHealthEvent{
			SchemaVersion:      "v1",
			EventType:          "feed-health",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTCUSDT",
			QuoteCurrency:      "USDT",
			Venue:              VenueBinance,
			MarketType:         "spot",
			ExchangeTs:         "2026-03-06T12:04:00Z",
			RecvTs:             "2026-03-06T12:04:00.050Z",
			TimestampStatus:    TimestampStatusNormal,
			FeedHealthState:    FeedHealthDegraded,
			DegradationReasons: []DegradationReason{ReasonSequenceGap},
			SourceRecordID:     "runtime:binance-spot-depth:BTCUSDT",
			CanonicalEventTime: mustRawTime(t, "2026-03-06T12:04:00Z"),
		},
		"BTCUSDT",
		string(StreamOrderBook),
		RawWriteContext{ConnectionRef: "binance-spot-ws", SessionRef: "binance-spot-session-2", DegradedFeedRef: "runtime:binance-spot-depth:BTCUSDT"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build feed-health entry: %v", err)
	}
	if entry.StreamFamily != RawStreamFamilyFeedHealth {
		t.Fatalf("stream family = %q, want %q", entry.StreamFamily, RawStreamFamilyFeedHealth)
	}
	if entry.StreamKey != string(StreamOrderBook) {
		t.Fatalf("stream key = %q, want %q", entry.StreamKey, StreamOrderBook)
	}
	if entry.DegradedFeedRef != "runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("degradedFeedRef = %q, want %q", entry.DegradedFeedRef, "runtime:binance-spot-depth:BTCUSDT")
	}
	if entry.PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/feed-health" {
		t.Fatalf("partition key = %q, want %q", entry.PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/feed-health")
	}
	if entry.DuplicateAudit.IdentityKey != "message:order-book:runtime:binance-spot-depth:BTCUSDT" {
		t.Fatalf("identity key = %q, want %q", entry.DuplicateAudit.IdentityKey, "message:order-book:runtime:binance-spot-depth:BTCUSDT")
	}
}

func TestBuildRawAppendEntryKeepsUSDMFundingAndMarkIndexDistinct(t *testing.T) {
	writer := NewInMemoryRawEventWriter()
	metadata := DerivativesMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "perpetual"}
	context := RawWriteContext{ConnectionRef: "binance-usdm-ws", SessionRef: "binance-usdm-session-4"}
	options := RawWriteOptions{BuildVersion: "test-build"}
	fundingMessage := FundingRateMessage{Type: "funding-rate", FundingRate: "0.00038167", NextFundingTs: "2026-03-06T16:00:00Z", ExchangeTs: "2026-03-06T12:00:05Z", RecvTs: "2026-03-06T12:00:05.040Z"}
	markMessage := MarkIndexMessage{Type: "mark-index", MarkPrice: "64010.10", IndexPrice: "64009.90", ExchangeTs: "2026-03-06T12:00:05Z", RecvTs: "2026-03-06T12:00:05.040Z"}

	fundingEntry, err := BuildRawAppendEntryFromFundingRate(
		CanonicalFundingRateEvent{
			SchemaVersion:      "v1",
			EventType:          "funding-rate",
			Symbol:             metadata.Symbol,
			SourceSymbol:       metadata.SourceSymbol,
			QuoteCurrency:      metadata.QuoteCurrency,
			Venue:              metadata.Venue,
			MarketType:         metadata.MarketType,
			FundingRate:        fundingMessage.FundingRate,
			NextFundingTs:      fundingMessage.NextFundingTs,
			ExchangeTs:         fundingMessage.ExchangeTs,
			RecvTs:             fundingMessage.RecvTs,
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "funding:2026-03-06T12:00:05Z",
			CanonicalEventTime: mustRawTime(t, fundingMessage.ExchangeTs),
		},
		metadata,
		fundingMessage,
		context,
		options,
	)
	if err != nil {
		t.Fatalf("build funding entry: %v", err)
	}
	markEntry, err := BuildRawAppendEntryFromMarkIndex(
		CanonicalMarkIndexEvent{
			SchemaVersion:      "v1",
			EventType:          "mark-index",
			Symbol:             metadata.Symbol,
			SourceSymbol:       metadata.SourceSymbol,
			QuoteCurrency:      metadata.QuoteCurrency,
			Venue:              metadata.Venue,
			MarketType:         metadata.MarketType,
			MarkPrice:          markMessage.MarkPrice,
			IndexPrice:         markMessage.IndexPrice,
			ExchangeTs:         markMessage.ExchangeTs,
			RecvTs:             markMessage.RecvTs,
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "mark:2026-03-06T12:00:05Z",
			CanonicalEventTime: mustRawTime(t, markMessage.ExchangeTs),
		},
		metadata,
		markMessage,
		context,
		options,
	)
	if err != nil {
		t.Fatalf("build mark-index entry: %v", err)
	}
	if fundingEntry.DuplicateAudit.IdentityKey == markEntry.DuplicateAudit.IdentityKey {
		t.Fatalf("identity keys unexpectedly match: %q", fundingEntry.DuplicateAudit.IdentityKey)
	}
	if err := writer.Append(fundingEntry); err != nil {
		t.Fatalf("append funding entry: %v", err)
	}
	if err := writer.Append(markEntry); err != nil {
		t.Fatalf("append mark-index entry: %v", err)
	}
	entries := writer.Entries()
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].DuplicateAudit.Duplicate || entries[1].DuplicateAudit.Duplicate {
		t.Fatalf("cross-family entries should not be marked duplicate: %+v %+v", entries[0].DuplicateAudit, entries[1].DuplicateAudit)
	}
	if entries[0].DuplicateAudit.Occurrence != 1 || entries[1].DuplicateAudit.Occurrence != 1 {
		t.Fatalf("occurrences = %d and %d, want 1 and 1", entries[0].DuplicateAudit.Occurrence, entries[1].DuplicateAudit.Occurrence)
	}
	if entries[0].PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/funding-rate" {
		t.Fatalf("funding partition key = %q, want %q", entries[0].PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/funding-rate")
	}
	if entries[1].PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE/mark-index" {
		t.Fatalf("mark-index partition key = %q, want %q", entries[1].PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE/mark-index")
	}
}

func TestRawWriteBoundaryRejectsContractMismatch(t *testing.T) {
	_, err := BuildRawAppendEntryFromTrade(
		CanonicalTradeEvent{
			SchemaVersion:      "v2",
			EventType:          "market-trade",
			Symbol:             "BTC-USD",
			SourceSymbol:       "BTC-USD",
			QuoteCurrency:      "USD",
			Venue:              VenueCoinbase,
			MarketType:         "spot",
			ExchangeTs:         "2026-03-06T12:00:00Z",
			RecvTs:             "2026-03-06T12:00:00.010Z",
			TimestampStatus:    TimestampStatusNormal,
			SourceRecordID:     "trade:mismatch",
			CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:00Z"),
		},
		TradeMetadata{Symbol: "BTC-USD", SourceSymbol: "BTC-USD", QuoteCurrency: "USD", Venue: VenueCoinbase, MarketType: "spot"},
		TradeMessage{TradeID: "mismatch", ExchangeTs: "2026-03-06T12:00:00Z", RecvTs: "2026-03-06T12:00:00.010Z"},
		RawWriteContext{ConnectionRef: "conn-1", SessionRef: "sess-1"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err == nil {
		t.Fatal("expected contract mismatch error")
	}
}

func TestRawPartitionRoutingUsesPersistedBucketDecision(t *testing.T) {
	entry, err := BuildRawAppendEntryFromTrade(
		CanonicalTradeEvent{
			SchemaVersion:           "v1",
			EventType:               "market-trade",
			Symbol:                  "ETH-USD",
			SourceSymbol:            "ETHUSDT",
			QuoteCurrency:           "USDT",
			Venue:                   VenueBinance,
			MarketType:              "spot",
			ExchangeTs:              "2026-03-05T23:59:59Z",
			RecvTs:                  "2026-03-06T00:00:01Z",
			TimestampStatus:         TimestampStatusDegraded,
			SourceRecordID:          "trade:routing",
			CanonicalEventTime:      mustRawTime(t, "2026-03-06T00:00:01Z"),
			TimestampFallbackReason: TimestampReasonExchangeSkewExceeded,
		},
		TradeMetadata{Symbol: "ETH-USD", SourceSymbol: "ETHUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "spot"},
		TradeMessage{TradeID: "routing", ExchangeTs: "2026-03-05T23:59:59Z", RecvTs: "2026-03-06T00:00:01Z"},
		RawWriteContext{ConnectionRef: "conn-3", SessionRef: "sess-3"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if entry.PartitionKey.String() != "2026-03-06/ETH-USD/BINANCE" {
		t.Fatalf("partition key = %q, want %q", entry.PartitionKey.String(), "2026-03-06/ETH-USD/BINANCE")
	}
	if entry.BucketTimestampSource != RawBucketTimestampSourceRecv {
		t.Fatalf("bucketTimestampSource = %q, want %q", entry.BucketTimestampSource, RawBucketTimestampSourceRecv)
	}
}

func TestRawPartitionRoutingIsStableForDuplicateInputs(t *testing.T) {
	writer := NewInMemoryRawEventWriter()
	var keys []RawPartitionKey
	for _, recv := range []string{"2026-03-06T12:00:00.100Z", "2026-03-06T12:00:00.100Z"} {
		entry, err := BuildRawAppendEntryFromOrderBook(
			CanonicalOrderBookEvent{
				SchemaVersion:      "v1",
				EventType:          "order-book-top",
				BookAction:         string(BookUpdateDelta),
				Symbol:             "ETH-USD",
				SourceSymbol:       "ETH/USD",
				QuoteCurrency:      "USD",
				Venue:              VenueKraken,
				MarketType:         "spot",
				ExchangeTs:         "2026-03-06T12:00:00Z",
				RecvTs:             recv,
				TimestampStatus:    TimestampStatusNormal,
				SourceRecordID:     "book:901",
				CanonicalEventTime: mustRawTime(t, "2026-03-06T12:00:00Z"),
			},
			BookMetadata{Symbol: "ETH-USD", SourceSymbol: "ETH/USD", QuoteCurrency: "USD", Venue: VenueKraken, MarketType: "spot"},
			OrderBookMessage{Type: string(BookUpdateDelta), Sequence: 901, ExchangeTs: "2026-03-06T12:00:00Z", RecvTs: recv},
			RawWriteContext{ConnectionRef: "conn-book", SessionRef: "sess-book"},
			RawWriteOptions{BuildVersion: "test-build"},
		)
		if err != nil {
			t.Fatalf("build entry: %v", err)
		}
		if err := writer.Append(entry); err != nil {
			t.Fatalf("append entry: %v", err)
		}
		keys = append(keys, entry.PartitionKey)
	}
	if keys[0] != keys[1] {
		t.Fatalf("partition keys differ: %+v vs %+v", keys[0], keys[1])
	}
	if keys[0].String() != "2026-03-06/ETH-USD/KRAKEN/order-book" {
		t.Fatalf("partition key = %q, want %q", keys[0].String(), "2026-03-06/ETH-USD/KRAKEN/order-book")
	}
}

func TestBuildRawAppendEntryFromTopOfBookUsesTopOfBookStreamFamily(t *testing.T) {
	entry, err := BuildRawAppendEntryFromOrderBook(
		CanonicalOrderBookEvent{
			SchemaVersion:           "v1",
			EventType:               "order-book-top",
			BookAction:              string(BookUpdateTopOfBook),
			Symbol:                  "BTC-USD",
			SourceSymbol:            "BTCUSDT",
			QuoteCurrency:           "USDT",
			Venue:                   VenueBinance,
			MarketType:              "spot",
			RecvTs:                  "2026-03-06T12:00:00.18Z",
			TimestampStatus:         TimestampStatusDegraded,
			SourceRecordID:          "ticker:9001",
			CanonicalEventTime:      mustRawTime(t, "2026-03-06T12:00:00.18Z"),
			TimestampFallbackReason: TimestampReasonExchangeMissingOrInvalid,
		},
		BookMetadata{Symbol: "BTC-USD", SourceSymbol: "BTCUSDT", QuoteCurrency: "USDT", Venue: VenueBinance, MarketType: "spot"},
		OrderBookMessage{Type: string(BookUpdateTopOfBook), Sequence: 9001, RecvTs: "2026-03-06T12:00:00.18Z"},
		RawWriteContext{ConnectionRef: "binance-spot-ws", SessionRef: "spot-session-1"},
		RawWriteOptions{BuildVersion: "test-build"},
	)
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if entry.StreamFamily != string(StreamTopOfBook) {
		t.Fatalf("stream family = %q, want %q", entry.StreamFamily, StreamTopOfBook)
	}
	if entry.PartitionKey.String() != "2026-03-06/BTC-USD/BINANCE" {
		t.Fatalf("partition key = %q, want %q", entry.PartitionKey.String(), "2026-03-06/BTC-USD/BINANCE")
	}
	if entry.VenueSequence != 9001 {
		t.Fatalf("venue sequence = %d, want %d", entry.VenueSequence, 9001)
	}
}

func mustRawTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse raw time %q: %v", value, err)
	}
	return parsed
}
