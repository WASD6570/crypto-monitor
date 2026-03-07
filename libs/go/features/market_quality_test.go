package features

import (
	"testing"
	"time"
)

func TestDivergenceMetrics(t *testing.T) {
	config := testBucketConfig()
	t.Run("aligned", func(t *testing.T) {
		rollup := buildRollupFromClosed(config, "BTC-USD", BucketFamily2m, []MarketQualityBucket{
			summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 100, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 101, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:00:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 102, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 103, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:01:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 104, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 105, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:01:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 106, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 107, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
		})
		if rollup.Divergence.DirectionAgreement != DirectionAgreementAligned {
			t.Fatalf("direction agreement = %q, want %q", rollup.Divergence.DirectionAgreement, DirectionAgreementAligned)
		}
	})
	t.Run("opposed", func(t *testing.T) {
		rollup := buildRollupFromClosed(config, "BTC-USD", BucketFamily2m, []MarketQualityBucket{
			summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 100, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 110, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:00:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 102, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 108, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:01:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 104, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 106, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
			summarize30sBucket(config, testAssignment("2026-03-06T12:01:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 106, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 104, 1, 0.98, 0.6, 0, false, "coinbase:spot")),
		})
		if rollup.Divergence.DirectionAgreement != DirectionAgreementOpposed {
			t.Fatalf("direction agreement = %q, want %q", rollup.Divergence.DirectionAgreement, DirectionAgreementOpposed)
		}
	})
}

func TestFragmentationSeverity(t *testing.T) {
	config := testBucketConfig()
	low := summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 100, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 100.01, 1, 0.99, 0.6, 0, false, "coinbase:spot"))
	if low.Fragmentation.Severity != FragmentationSeverityLow {
		t.Fatalf("low severity = %q, want %q", low.Fragmentation.Severity, FragmentationSeverityLow)
	}
	moderate := summarize30sBucket(config, testAssignment("2026-03-06T12:00:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 100, 1, 0.99, 0.6, 0, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 100.05, 0.75, 0.99, 0.6, 0, false, "coinbase:spot"))
	if moderate.Fragmentation.Severity != FragmentationSeverityModerate {
		t.Fatalf("moderate severity = %q, want %q", moderate.Fragmentation.Severity, FragmentationSeverityModerate)
	}
	severe := summarize30sBucket(config, testAssignment("2026-03-06T12:01:00Z", BucketSourceRecvTs), testSnapshot(CompositeGroupWorld, 0, 0, 0, 0, 0, true, ""), testSnapshot(CompositeGroupUSA, 100, 1, 0.99, 0.6, 0, false, "coinbase:spot"))
	if severe.Fragmentation.Severity != FragmentationSeveritySevere {
		t.Fatalf("severe severity = %q, want %q", severe.Fragmentation.Severity, FragmentationSeveritySevere)
	}
}

func TestCoverageAndTimestampTrustSummaries(t *testing.T) {
	config := testBucketConfig()
	rollup := buildRollupFromClosed(config, "ETH-USD", BucketFamily2m, []MarketQualityBucket{
		summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceRecvTs), testSnapshot(CompositeGroupWorld, 3500, 1, 0.99, 0.6, 2, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 3500.2, 0.5, 0.98, 0.6, 0, false, "coinbase:spot")),
		summarize30sBucket(config, testAssignment("2026-03-06T12:00:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 3500.3, 1, 0.99, 0.6, 2, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 3500.1, 0.5, 0.98, 0.6, 0, false, "coinbase:spot")),
		summarize30sBucket(config, testAssignment("2026-03-06T12:01:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 3500.4, 1, 0.99, 0.6, 1, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 3500.2, 0.5, 0.98, 0.6, 0, false, "coinbase:spot")),
		summarize30sBucket(config, testAssignment("2026-03-06T12:01:30Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 3500.5, 1, 0.99, 0.6, 1, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 3500.3, 0.5, 0.98, 0.6, 0, false, "coinbase:spot")),
	})
	if rollup.Divergence.ParticipationGap <= 0 {
		t.Fatalf("participation gap = %v, want positive", rollup.Divergence.ParticipationGap)
	}
	if !rollup.TimestampTrust.OneSidedFallback {
		t.Fatal("expected one-sided fallback")
	}
	if !rollup.TimestampTrust.TrustCap {
		t.Fatal("expected trust cap")
	}
}

func TestMarketQualityCaps(t *testing.T) {
	config := testBucketConfig()
	bucket := summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceRecvTs), testSnapshot(CompositeGroupWorld, 64000, 1, 0.92, 0.9, 2, false, "binance:spot"), testSnapshot(CompositeGroupUSA, 64020, 0.5, 0.7, 0.85, 0, false, "coinbase:spot"))
	if bucket.MarketQuality.CombinedTrustCap > config.Thresholds.Quality.TimestampTrustCap {
		t.Fatalf("combined trust cap = %v, want <= %v", bucket.MarketQuality.CombinedTrustCap, config.Thresholds.Quality.TimestampTrustCap)
	}
	if len(bucket.MarketQuality.DowngradedReasons) == 0 {
		t.Fatal("expected downgraded reasons")
	}
}

func TestUnavailableCompositeSummary(t *testing.T) {
	config := testBucketConfig()
	bucket := summarize30sBucket(config, testAssignment("2026-03-06T12:00:00Z", BucketSourceExchangeTs), testSnapshot(CompositeGroupWorld, 0, 0, 0, 0, 0, true, ""), testSnapshot(CompositeGroupUSA, 0, 0, 0, 0, 0, true, ""))
	if bucket.Divergence.Available {
		t.Fatal("expected unavailable divergence summary")
	}
	if bucket.Fragmentation.UnavailableSideCount == 0 {
		t.Fatal("expected unavailable-side marker")
	}
	if bucket.MarketQuality.CombinedTrustCap != 0 {
		t.Fatalf("combined trust cap = %v, want 0", bucket.MarketQuality.CombinedTrustCap)
	}
}

func testAssignment(start string, source BucketSource) BucketAssignment {
	startTs := mustTime(start)
	return BucketAssignment{
		Symbol:       "BTC-USD",
		Family:       BucketFamily30s,
		BucketStart:  startTs.Format(time.RFC3339Nano),
		BucketEnd:    startTs.Add(30 * time.Second).Format(time.RFC3339Nano),
		BucketSource: source,
	}
}

func mustTime(value string) time.Time {
	ts, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return ts
}
