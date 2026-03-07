package integration

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestWorldUSABucketReplayWindow(t *testing.T) {
	service := newBucketIntegrationService(t)
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	firstRun := runBucketWindow(t, service, start)
	service = newBucketIntegrationService(t)
	secondRun := runBucketWindow(t, service, start)
	if !reflect.DeepEqual(firstRun, secondRun) {
		t.Fatalf("bucket replay outputs differ\nfirst: %+v\nsecond: %+v", firstRun, secondRun)
	}
}

func TestWorldUSABucketSummarySeams(t *testing.T) {
	service := newBucketIntegrationService(t)
	buckets := runBucketWindow(t, service, time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC))
	var found5m bool
	for _, bucket := range buckets {
		if bucket.Window.Family != features.BucketFamily5m {
			continue
		}
		found5m = true
		if bucket.Window.ConfigVersion == "" || bucket.Window.AlgorithmVersion == "" {
			t.Fatalf("version seams missing: %+v", bucket.Window)
		}
		if bucket.MarketQuality.ReplayProvenance == "" {
			t.Fatalf("replay provenance missing: %+v", bucket.MarketQuality)
		}
		if bucket.Divergence.DirectionAgreement == "" || bucket.Fragmentation.Severity == "" {
			t.Fatalf("summary seams missing: %+v", bucket)
		}
	}
	if !found5m {
		t.Fatal("expected 5m summary seam")
	}
}

func newBucketIntegrationService(t *testing.T) *featureengine.Service {
	t.Helper()
	service, err := featureengine.NewService(features.CompositeConfig{
		SchemaVersion:    "v1",
		ConfigVersion:    "composite-config.v1",
		AlgorithmVersion: "world-usa-composite.v1",
		Penalties:        features.PenaltyConfig{FeedHealthDegradedMultiplier: 0.8, TimestampDegradedMultiplier: 0.75},
		QuoteProxies:     map[string]features.QuoteProxyRule{"USDT": {Enabled: true, PenaltyMultiplier: 1}, "USDC": {Enabled: true, PenaltyMultiplier: 0.98}},
		Groups: map[features.CompositeGroup]features.GroupConfig{
			features.CompositeGroupWorld: {Members: []features.MemberConfig{{Venue: ingestion.VenueBinance, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueBybit, MarketType: "perpetual", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.8}},
			features.CompositeGroupUSA:   {Members: []features.MemberConfig{{Venue: ingestion.VenueCoinbase, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}, {Venue: ingestion.VenueKraken, MarketType: "spot", Symbols: []string{"BTC-USD", "ETH-USD"}}}, Clamp: features.ClampConfig{MinWeight: 0.2, MaxWeight: 0.7}},
		},
	}, featureengine.WithBucketConfig(features.BucketConfig{
		SchemaVersion:        "v1",
		ConfigVersion:        "market-quality.v1",
		AlgorithmVersion:     "market-quality-buckets.v1",
		TimestampSkewSeconds: 2,
		Families:             map[features.BucketFamily]features.BucketFamilyConfig{features.BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1}, features.BucketFamily2m: {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 0.75}, features.BucketFamily5m: {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 0.8}},
		Thresholds:           features.BucketThresholdConfig{Divergence: map[features.BucketFamily]features.DivergenceThresholds{features.BucketFamily30s: {PriceDistanceModerateBps: 2, PriceDistanceSevereBps: 8, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}, features.BucketFamily2m: {PriceDistanceModerateBps: 3, PriceDistanceSevereBps: 10, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}, features.BucketFamily5m: {PriceDistanceModerateBps: 4, PriceDistanceSevereBps: 12, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}}, Quality: features.MarketQualityThresholds{ConcentrationSoftCap: 0.7, ModerateCap: 0.65, SevereCap: 0.35, TimestampTrustCap: 0.55, IncompleteCap: 0.6}},
	}))
	if err != nil {
		t.Fatalf("new bucket integration service: %v", err)
	}
	return service
}

func runBucketWindow(t *testing.T, service *featureengine.Service, start time.Time) []features.MarketQualityBucket {
	t.Helper()
	var emitted []features.MarketQualityBucket
	for index := 0; index < 10; index++ {
		bucketStart := start.Add(time.Duration(index) * 30 * time.Second)
		result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
			Symbol:     "BTC-USD",
			ExchangeTs: bucketStart.Add(10 * time.Second),
			RecvTs:     bucketStart.Add(11 * time.Second),
			Now:        bucketStart.Add(11 * time.Second),
			World:      integrationSnapshot(features.CompositeGroupWorld, 64000+float64(index), 1, 0.99, 0.6, 0, false, ingestion.VenueBinance),
			USA:        integrationSnapshot(features.CompositeGroupUSA, 64001+float64(index), 1, 0.98, 0.6, 0, false, ingestion.VenueCoinbase),
		})
		if err != nil {
			t.Fatalf("observe bucket %d: %v", index, err)
		}
		emitted = append(emitted, result.Emitted...)
	}
	advanced, err := service.AdvanceWorldUSABuckets("BTC-USD", start.Add(5*time.Minute+33*time.Second))
	if err != nil {
		t.Fatalf("advance integration window: %v", err)
	}
	emitted = append(emitted, advanced...)
	return emitted
}

func integrationSnapshot(group features.CompositeGroup, price float64, coverage float64, health float64, maxWeight float64, fallbackCount int, unavailable bool, venue ingestion.Venue) features.CompositeSnapshot {
	contributors := []features.SnapshotContributor{}
	if !unavailable {
		contributors = append(contributors, features.SnapshotContributor{Venue: venue, MarketType: "spot", FinalWeight: maxWeight})
	}
	snapshot := features.CompositeSnapshot{CompositeGroup: group, Contributors: contributors, ConfiguredContributorCount: 2, EligibleContributorCount: 2, ContributingContributorCount: 2, CoverageRatio: coverage, HealthScore: health, MaxContributorWeight: maxWeight, TimestampFallbackContributorCount: fallbackCount, Unavailable: unavailable}
	if !unavailable {
		snapshot.CompositePrice = &price
	}
	return snapshot
}
