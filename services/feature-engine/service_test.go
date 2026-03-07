package featureengine

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestWorldUSACompositeConstruction(t *testing.T) {
	service := newService(t)
	snapshots, err := service.BuildWorldUSASnapshots("BTC-USD", time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "BTC-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 64000, LiquidityScore: 140, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 64010, LiquidityScore: 120, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 64005, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "BTC-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 64002, LiquidityScore: 80, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
	})
	if err != nil {
		t.Fatalf("build snapshots: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("snapshot count = %d, want 2", len(snapshots))
	}
	if snapshots[0].CompositeGroup != features.CompositeGroupWorld || snapshots[1].CompositeGroup != features.CompositeGroupUSA {
		t.Fatalf("groups = %v, want WORLD then USA", []features.CompositeGroup{snapshots[0].CompositeGroup, snapshots[1].CompositeGroup})
	}
	if snapshots[0].CompositePrice == nil || snapshots[1].CompositePrice == nil {
		t.Fatal("expected composite prices for both snapshots")
	}
}

func TestCompositeDegradedVenueHandling(t *testing.T) {
	service := newService(t)
	snapshot, err := service.BuildCompositeSnapshot(features.CompositeGroupWorld, "ETH-USD", time.Date(2026, 3, 6, 12, 3, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3500, LiquidityScore: 120, TimestampStatus: ingestion.TimestampStatusDegraded, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueBybit, MarketType: "perpetual", QuoteCurrency: "USDT", Price: 3502, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded, FeedHealthReasons: []ingestion.DegradationReason{ingestion.ReasonReconnectLoop}},
	})
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	if !snapshot.Degraded {
		t.Fatal("expected degraded snapshot")
	}
	if snapshot.TimestampFallbackContributorCount != 1 {
		t.Fatalf("timestamp fallback contributors = %d, want 1", snapshot.TimestampFallbackContributorCount)
	}
	if !containsReason(snapshot.DegradedReasons, features.ReasonTimestampFallback) {
		t.Fatalf("degraded reasons = %v, want %q", snapshot.DegradedReasons, features.ReasonTimestampFallback)
	}
	if !containsReason(snapshot.DegradedReasons, features.ReasonFeedHealthDegraded) {
		t.Fatalf("degraded reasons = %v, want %q", snapshot.DegradedReasons, features.ReasonFeedHealthDegraded)
	}
}

func TestCompositeAllContributorsExcluded(t *testing.T) {
	service := newService(t)
	snapshot, err := service.BuildCompositeSnapshot(features.CompositeGroupUSA, "BTC-USD", time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 64000, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthStale},
		{Symbol: "BTC-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 64020, LiquidityScore: 100, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthDegraded, FeedHealthReasons: []ingestion.DegradationReason{ingestion.ReasonSequenceGap}},
	})
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	if !snapshot.Unavailable {
		t.Fatal("expected unavailable snapshot")
	}
	if snapshot.CompositePrice != nil {
		t.Fatalf("composite price = %v, want nil", *snapshot.CompositePrice)
	}
	if !reflect.DeepEqual(snapshot.DegradedReasons, []features.ReasonCode{features.ReasonNoContributors}) {
		t.Fatalf("degraded reasons = %v, want no-contributors only", snapshot.DegradedReasons)
	}
}

func TestWorldUSACompositeSnapshotShape(t *testing.T) {
	service := newService(t)
	snapshot, err := service.BuildCompositeSnapshot(features.CompositeGroupUSA, "ETH-USD", time.Date(2026, 3, 6, 12, 4, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueCoinbase, MarketType: "spot", QuoteCurrency: "USD", Price: 3501, LiquidityScore: 95, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
		{Symbol: "ETH-USD", Venue: ingestion.VenueKraken, MarketType: "spot", QuoteCurrency: "USD", Price: 3500, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy},
	})
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	if snapshot.SchemaVersion == "" || snapshot.ConfigVersion == "" || snapshot.AlgorithmVersion == "" {
		t.Fatalf("version fields missing: %+v", snapshot)
	}
	if snapshot.PriceBasis == "" || snapshot.QuoteNormalizationMode == "" {
		t.Fatalf("boundary fields missing: %+v", snapshot)
	}
	if len(snapshot.Contributors) != 2 {
		t.Fatalf("contributors = %d, want 2", len(snapshot.Contributors))
	}
	if snapshot.CoverageRatio != 1 || snapshot.MaxContributorWeight == 0 {
		t.Fatalf("coverage or concentration seam missing: %+v", snapshot)
	}
}

func TestCompositeUnavailableState(t *testing.T) {
	service := newService(t)
	snapshot, err := service.BuildCompositeSnapshot(features.CompositeGroupWorld, "ETH-USD", time.Date(2026, 3, 6, 12, 5, 0, 0, time.UTC), []features.ContributorInput{
		{Symbol: "ETH-USD", Venue: ingestion.VenueBinance, MarketType: "spot", QuoteCurrency: "USDT", Price: 3501, LiquidityScore: 90, TimestampStatus: ingestion.TimestampStatusNormal, FeedHealthState: ingestion.FeedHealthHealthy, QuoteConfidenceDropped: true},
	})
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	if !snapshot.Unavailable {
		t.Fatal("expected unavailable snapshot")
	}
	if !snapshot.Degraded {
		t.Fatal("expected unavailable snapshot to remain degraded")
	}
}

func TestBucketAssignment(t *testing.T) {
	service := newBucketService(t)
	result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "BTC-USD",
		ExchangeTs: time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC),
		RecvTs:     time.Date(2026, 3, 6, 12, 0, 10, 500000000, time.UTC),
		Now:        time.Date(2026, 3, 6, 12, 0, 10, 500000000, time.UTC),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 64000, 1, 0.99, 0.6, 0, false, "binance"),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 64001, 1, 0.99, 0.6, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe bucket: %v", err)
	}
	if result.Assignment.BucketStart != "2026-03-06T12:00:00Z" {
		t.Fatalf("bucket start = %s, want 2026-03-06T12:00:00Z", result.Assignment.BucketStart)
	}
	emitted, err := service.AdvanceWorldUSABuckets("BTC-USD", time.Date(2026, 3, 6, 12, 0, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance buckets: %v", err)
	}
	if len(emitted) == 0 || emitted[0].Window.Family != features.BucketFamily30s {
		t.Fatalf("emitted families = %+v, want 30s output", emitted)
	}
}

func TestMissing30sBucketPropagation(t *testing.T) {
	service := newBucketService(t)
	starts := []time.Time{
		time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 6, 12, 0, 30, 0, time.UTC),
		time.Date(2026, 3, 6, 12, 1, 30, 0, time.UTC),
	}
	var emitted []features.MarketQualityBucket
	for _, start := range starts {
		result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
			Symbol:     "ETH-USD",
			ExchangeTs: start.Add(10 * time.Second),
			RecvTs:     start.Add(11 * time.Second),
			Now:        start.Add(11 * time.Second),
			World:      testBucketSnapshot(features.CompositeGroupWorld, 3500, 1, 0.99, 0.6, 0, false, "binance"),
			USA:        testBucketSnapshot(features.CompositeGroupUSA, 3501, 1, 0.99, 0.6, 0, false, "coinbase"),
		})
		if err != nil {
			t.Fatalf("observe bucket at %s: %v", start, err)
		}
		emitted = append(emitted, result.Emitted...)
	}
	advanced, err := service.AdvanceWorldUSABuckets("ETH-USD", time.Date(2026, 3, 6, 12, 2, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance missing bucket rollup: %v", err)
	}
	emitted = append(emitted, advanced...)
	rollup := findServiceBucket(emitted, features.BucketFamily2m, "2026-03-06T12:02:00Z")
	if rollup == nil {
		t.Fatal("expected 2m rollup")
	}
	if rollup.Window.MissingBucketCount != 1 {
		t.Fatalf("missing bucket count = %d, want 1", rollup.Window.MissingBucketCount)
	}
}

func TestLateEventHandling(t *testing.T) {
	service := newBucketService(t)
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	first, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "BTC-USD",
		ExchangeTs: start.Add(10 * time.Second),
		RecvTs:     start.Add(11 * time.Second),
		Now:        start.Add(11 * time.Second),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 64000, 1, 0.99, 0.6, 0, false, "binance"),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 64001, 1, 0.99, 0.6, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe first bucket: %v", err)
	}
	if _, err := service.AdvanceWorldUSABuckets("BTC-USD", start.Add(33*time.Second)); err != nil {
		t.Fatalf("advance first bucket: %v", err)
	}
	late, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "BTC-USD",
		ExchangeTs: start.Add(29 * time.Second),
		RecvTs:     start.Add(29 * time.Second),
		Now:        start.Add(50 * time.Second),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 65000, 1, 0.99, 0.6, 0, false, "binance"),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 65001, 1, 0.99, 0.6, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe late bucket: %v", err)
	}
	if !first.Accepted || late.Accepted {
		t.Fatalf("accepted flags = %v/%v, want true/false", first.Accepted, late.Accepted)
	}
	if len(late.Emitted) != 0 {
		t.Fatalf("late emitted = %+v, want no mutation", late.Emitted)
	}
}

func TestDivergenceMetrics(t *testing.T) {
	service := newBucketService(t)
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	var emitted []features.MarketQualityBucket
	for index, usaPrice := range []float64{110, 108, 106, 104} {
		result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
			Symbol:     "BTC-USD",
			ExchangeTs: start.Add(time.Duration(index)*30*time.Second + 10*time.Second),
			RecvTs:     start.Add(time.Duration(index)*30*time.Second + 11*time.Second),
			Now:        start.Add(time.Duration(index)*30*time.Second + 11*time.Second),
			World:      testBucketSnapshot(features.CompositeGroupWorld, 100+float64(index*2), 1, 0.99, 0.6, 0, false, "binance"),
			USA:        testBucketSnapshot(features.CompositeGroupUSA, usaPrice, 1, 0.99, 0.6, 0, false, "coinbase"),
		})
		if err != nil {
			t.Fatalf("observe divergence bucket %d: %v", index, err)
		}
		emitted = append(emitted, result.Emitted...)
	}
	advanced, err := service.AdvanceWorldUSABuckets("BTC-USD", time.Date(2026, 3, 6, 12, 2, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance divergence rollup: %v", err)
	}
	emitted = append(emitted, advanced...)
	rollup := findServiceBucket(emitted, features.BucketFamily2m, "2026-03-06T12:02:00Z")
	if rollup == nil {
		t.Fatal("expected 2m rollup")
	}
	if rollup.Divergence.DirectionAgreement != features.DirectionAgreementOpposed {
		t.Fatalf("direction agreement = %q, want %q", rollup.Divergence.DirectionAgreement, features.DirectionAgreementOpposed)
	}
}

func TestFragmentationOutputs(t *testing.T) {
	service := newBucketService(t)
	_, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "ETH-USD",
		ExchangeTs: time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC),
		RecvTs:     time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		Now:        time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 0, 0, 0, 0, 0, true, ""),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 3500, 1, 0.99, 0.6, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe fragmented bucket: %v", err)
	}
	emitted, err := service.AdvanceWorldUSABuckets("ETH-USD", time.Date(2026, 3, 6, 12, 0, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance fragmented bucket: %v", err)
	}
	if emitted[0].Fragmentation.Severity != features.FragmentationSeveritySevere {
		t.Fatalf("fragmentation severity = %q, want %q", emitted[0].Fragmentation.Severity, features.FragmentationSeveritySevere)
	}
	if len(emitted[0].Fragmentation.PrimaryCauses) == 0 {
		t.Fatal("expected fragmentation causes")
	}
}

func TestMarketQualitySummary(t *testing.T) {
	service := newBucketService(t)
	_, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "BTC-USD",
		ExchangeTs: time.Date(2026, 3, 6, 12, 0, 10, 0, time.UTC),
		RecvTs:     time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		Now:        time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 64000, 1, 0.92, 0.9, 2, false, "binance"),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 64020, 0.5, 0.72, 0.85, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe market quality bucket: %v", err)
	}
	emitted, err := service.AdvanceWorldUSABuckets("BTC-USD", time.Date(2026, 3, 6, 12, 0, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance market quality bucket: %v", err)
	}
	bucket := emitted[0]
	if bucket.MarketQuality.CombinedTrustCap <= 0 || bucket.MarketQuality.CombinedTrustCap > 0.55 {
		t.Fatalf("combined trust cap = %v, want conservative non-zero cap", bucket.MarketQuality.CombinedTrustCap)
	}
}

func TestTimestampTrustPropagation(t *testing.T) {
	service := newBucketService(t)
	_, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
		Symbol:     "ETH-USD",
		ExchangeTs: time.Time{},
		RecvTs:     time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		Now:        time.Date(2026, 3, 6, 12, 0, 11, 0, time.UTC),
		World:      testBucketSnapshot(features.CompositeGroupWorld, 3500, 1, 0.99, 0.6, 2, false, "binance"),
		USA:        testBucketSnapshot(features.CompositeGroupUSA, 3501, 1, 0.99, 0.6, 0, false, "coinbase"),
	})
	if err != nil {
		t.Fatalf("observe fallback bucket: %v", err)
	}
	emitted, err := service.AdvanceWorldUSABuckets("ETH-USD", time.Date(2026, 3, 6, 12, 0, 33, 0, time.UTC))
	if err != nil {
		t.Fatalf("advance fallback bucket: %v", err)
	}
	bucket := emitted[0]
	if bucket.Assignment.BucketSource != features.BucketSourceRecvTs {
		t.Fatalf("bucket source = %q, want %q", bucket.Assignment.BucketSource, features.BucketSourceRecvTs)
	}
	if !bucket.TimestampTrust.TrustCap {
		t.Fatal("expected timestamp trust cap")
	}
}

func newService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties:        features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75},
		QuoteProxies: map[string]features.QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
			"USDC": {Enabled: true, PenaltyMultiplier: 0.98},
		},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {
				Members: []features.MemberConfig{
					{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			features.CompositeGroupUSA: {
				Members: []features.MemberConfig{
					{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
					{Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}},
				},
				Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.7},
			},
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}

func newBucketService(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties:        features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75},
		QuoteProxies: map[string]features.QuoteProxyRule{
			"USDT": {Enabled: true, PenaltyMultiplier: 1},
			"USDC": {Enabled: true, PenaltyMultiplier: 0.98},
		},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {
				Members: []features.MemberConfig{{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8},
			},
			features.CompositeGroupUSA: {
				Members: []features.MemberConfig{{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}},
				Clamp:   features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.7},
			},
		},
	}, WithBucketConfig(features.BucketConfig{
		SchemaVersion:        "v1",
		ConfigVersion:        "market-quality.v1",
		AlgorithmVersion:     "market-quality-buckets.v1",
		TimestampSkewSeconds: 2,
		Families: map[features.BucketFamily]features.BucketFamilyConfig{
			features.BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1},
			features.BucketFamily2m:  {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 0.75},
			features.BucketFamily5m:  {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 0.8},
		},
		Thresholds: features.BucketThresholdConfig{
			Divergence: map[features.BucketFamily]features.DivergenceThresholds{
				features.BucketFamily30s: {PriceDistanceModerateBps: 2, PriceDistanceSevereBps: 8, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				features.BucketFamily2m:  {PriceDistanceModerateBps: 3, PriceDistanceSevereBps: 10, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
				features.BucketFamily5m:  {PriceDistanceModerateBps: 4, PriceDistanceSevereBps: 12, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2},
			},
			Quality: features.MarketQualityThresholds{ConcentrationSoftCap: 0.7, ModerateCap: 0.65, SevereCap: 0.35, TimestampTrustCap: 0.55, IncompleteCap: 0.6},
		},
	}))
	if err != nil {
		t.Fatalf("new bucket service: %v", err)
	}
	return service
}

func testBucketSnapshot(group features.CompositeGroup, price float64, coverage float64, health float64, maxWeight float64, fallbackCount int, unavailable bool, leader string) features.CompositeSnapshot {
	contributors := []features.SnapshotContributor{}
	if !unavailable {
		venue := ingestion.VenueBinance
		if leader == "coinbase" {
			venue = ingestion.VenueCoinbase
		}
		contributors = append(contributors, features.SnapshotContributor{Venue: venue, MarketType: "spot", FinalWeight: maxWeight})
	}
	snapshot := features.CompositeSnapshot{CompositeGroup: group, Contributors: contributors, ConfiguredContributorCount: 2, EligibleContributorCount: 2, ContributingContributorCount: 2, CoverageRatio: coverage, HealthScore: health, MaxContributorWeight: maxWeight, TimestampFallbackContributorCount: fallbackCount, Unavailable: unavailable}
	if !unavailable {
		snapshot.CompositePrice = &price
	}
	return snapshot
}

func findServiceBucket(buckets []features.MarketQualityBucket, family features.BucketFamily, end string) *features.MarketQualityBucket {
	for index := range buckets {
		if buckets[index].Window.Family == family && buckets[index].Window.End == end {
			return &buckets[index]
		}
	}
	return nil
}

func containsReason(reasons []features.ReasonCode, target features.ReasonCode) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
