package features

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestBucketAssignment(t *testing.T) {
	config := testBucketConfig()
	exchange := time.Date(2026, 3, 6, 12, 0, 17, 0, time.UTC)
	recv := exchange.Add(500 * time.Millisecond)
	for _, family := range []BucketFamily{BucketFamily30s, BucketFamily2m, BucketFamily5m} {
		assignment, err := AssignBucket(config, "BTC-USD", family, exchange, recv, recv)
		if err != nil {
			t.Fatalf("assign %s bucket: %v", family, err)
		}
		if assignment.BucketSource != BucketSourceExchangeTs {
			t.Fatalf("family %s source = %q, want %q", family, assignment.BucketSource, BucketSourceExchangeTs)
		}
		if assignment.LateDisposition != LateEventOnTime {
			t.Fatalf("family %s late disposition = %q, want %q", family, assignment.LateDisposition, LateEventOnTime)
		}
	}
	assignment2m, _ := AssignBucket(config, "BTC-USD", BucketFamily2m, exchange, recv, recv)
	assignment5m, _ := AssignBucket(config, "BTC-USD", BucketFamily5m, exchange, recv, recv)
	if assignment2m.BucketStart != "2026-03-06T12:00:00Z" {
		t.Fatalf("2m bucket start = %s, want 2026-03-06T12:00:00Z", assignment2m.BucketStart)
	}
	if assignment5m.BucketEnd != "2026-03-06T12:05:00Z" {
		t.Fatalf("5m bucket end = %s, want 2026-03-06T12:05:00Z", assignment5m.BucketEnd)
	}
}

func TestBucketAssignmentRecvTsFallback(t *testing.T) {
	config := testBucketConfig()
	recv := time.Date(2026, 3, 6, 12, 0, 31, 0, time.UTC)
	assignment, err := AssignBucket(config, "ETH-USD", BucketFamily30s, time.Time{}, recv, recv)
	if err != nil {
		t.Fatalf("assign fallback bucket: %v", err)
	}
	if assignment.BucketSource != BucketSourceRecvTs {
		t.Fatalf("bucket source = %q, want %q", assignment.BucketSource, BucketSourceRecvTs)
	}
	if !assignment.TimestampDegraded {
		t.Fatal("expected timestamp-degraded fallback")
	}
}

func TestBucketRollupFrom30sClosures(t *testing.T) {
	processor, err := NewWorldUSABucketProcessor(testBucketConfig())
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	var emitted []MarketQualityBucket
	for index := 0; index < 10; index++ {
		bucketStart := start.Add(time.Duration(index) * 30 * time.Second)
		result, observeErr := processor.Observe(WorldUSAObservation{
			Symbol:     "BTC-USD",
			ExchangeTs: bucketStart.Add(10 * time.Second),
			RecvTs:     bucketStart.Add(11 * time.Second),
			Now:        bucketStart.Add(11 * time.Second),
			World:      testSnapshot(CompositeGroupWorld, 64000+float64(index), 1, 0.99, 0.55, 0, false, "binance:spot"),
			USA:        testSnapshot(CompositeGroupUSA, 64001+float64(index), 1, 0.98, 0.56, 0, false, "coinbase:spot"),
		})
		if observeErr != nil {
			t.Fatalf("observe bucket %d: %v", index, observeErr)
		}
		emitted = append(emitted, result.Emitted...)
	}
	emitted = append(emitted, processor.Advance("BTC-USD", start.Add(5*time.Minute+33*time.Second))...)
	rollup2m := findBucket(emitted, BucketFamily2m, "2026-03-06T12:02:00Z")
	if rollup2m == nil {
		t.Fatal("expected 2m rollup at 12:02:00Z")
	}
	rollup5m := findBucket(emitted, BucketFamily5m, "2026-03-06T12:05:00Z")
	if rollup5m == nil {
		t.Fatal("expected 5m rollup at 12:05:00Z")
	}
	if rollup5m.Window.ClosedBucketCount != 10 || rollup5m.Window.MissingBucketCount != 0 {
		t.Fatalf("5m rollup counts = %+v, want 10 closed and 0 missing", rollup5m.Window)
	}
}

func TestMissing30sBucketPropagation(t *testing.T) {
	processor, err := NewWorldUSABucketProcessor(testBucketConfig())
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	indices := []int{0, 1, 3}
	var emitted []MarketQualityBucket
	for _, index := range indices {
		bucketStart := start.Add(time.Duration(index) * 30 * time.Second)
		result, observeErr := processor.Observe(WorldUSAObservation{
			Symbol:     "ETH-USD",
			ExchangeTs: bucketStart.Add(10 * time.Second),
			RecvTs:     bucketStart.Add(11 * time.Second),
			Now:        bucketStart.Add(11 * time.Second),
			World:      testSnapshot(CompositeGroupWorld, 3500+float64(index), 1, 0.97, 0.54, 0, false, "binance:spot"),
			USA:        testSnapshot(CompositeGroupUSA, 3499+float64(index), 1, 0.96, 0.53, 0, false, "coinbase:spot"),
		})
		if observeErr != nil {
			t.Fatalf("observe bucket %d: %v", index, observeErr)
		}
		emitted = append(emitted, result.Emitted...)
	}
	emitted = append(emitted, processor.Advance("ETH-USD", start.Add(2*time.Minute+33*time.Second))...)
	rollup := findBucket(emitted, BucketFamily2m, "2026-03-06T12:02:00Z")
	if rollup == nil {
		t.Fatal("expected 2m rollup at 12:02:00Z")
	}
	if rollup.Window.MissingBucketCount != 1 {
		t.Fatalf("missing bucket count = %d, want 1", rollup.Window.MissingBucketCount)
	}
	if rollup.Fragmentation.Severity != FragmentationSeveritySevere {
		t.Fatalf("fragmentation severity = %q, want %q", rollup.Fragmentation.Severity, FragmentationSeveritySevere)
	}
}

func testBucketConfig() BucketConfig {
	return BucketConfig{
		SchemaVersion:        "v1",
		ConfigVersion:        "market-quality.v1",
		AlgorithmVersion:     "market-quality-buckets.v1",
		TimestampSkewSeconds: 2,
		Families: map[BucketFamily]BucketFamilyConfig{
			BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1},
			BucketFamily2m:  {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 0.75},
			BucketFamily5m:  {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 0.8},
		},
		Thresholds: BucketThresholdConfig{
			Divergence: map[BucketFamily]DivergenceThresholds{
				BucketFamily30s: {PriceDistanceModerateBps: 2, PriceDistanceSevereBps: 8, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				BucketFamily2m:  {PriceDistanceModerateBps: 3, PriceDistanceSevereBps: 10, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				BucketFamily5m:  {PriceDistanceModerateBps: 4, PriceDistanceSevereBps: 12, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
			},
			Quality: MarketQualityThresholds{ConcentrationSoftCap: 0.7, ModerateCap: 0.65, SevereCap: 0.35, TimestampTrustCap: 0.55, IncompleteCap: 0.6},
		},
	}
}

func testSnapshot(group CompositeGroup, price float64, coverage float64, health float64, maxWeight float64, fallbackCount int, unavailable bool, leader string) CompositeSnapshot {
	contributors := []SnapshotContributor{}
	if !unavailable {
		contributors = append(contributors, SnapshotContributor{Venue: ingestion.VenueBinance, MarketType: "spot", FinalWeight: maxWeight})
		if leader == "coinbase:spot" {
			contributors = []SnapshotContributor{{Venue: ingestion.VenueCoinbase, MarketType: "spot", FinalWeight: maxWeight}}
		}
	}
	snapshot := CompositeSnapshot{
		SchemaVersion:                     "v1",
		Symbol:                            "BTC-USD",
		BucketTs:                          "2026-03-06T12:00:00Z",
		CompositeGroup:                    group,
		Contributors:                      contributors,
		ConfiguredContributorCount:        2,
		EligibleContributorCount:          2,
		ContributingContributorCount:      2,
		CoverageRatio:                     coverage,
		HealthScore:                       health,
		MaxContributorWeight:              maxWeight,
		TimestampFallbackContributorCount: fallbackCount,
		Unavailable:                       unavailable,
	}
	if unavailable {
		return snapshot
	}
	snapshot.CompositePrice = &price
	return snapshot
}

func findBucket(buckets []MarketQualityBucket, family BucketFamily, end string) *MarketQualityBucket {
	for index := range buckets {
		if buckets[index].Window.Family == family && buckets[index].Window.End == end {
			return &buckets[index]
		}
	}
	return nil
}
