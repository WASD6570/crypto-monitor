package replay

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	featureengine "github.com/crypto-market-copilot/alerts/services/feature-engine"
)

func TestWorldUSABucketDeterminism(t *testing.T) {
	first := replayBucketWindow(t, newReplayBucketService(t), "market-quality.v1")
	second := replayBucketWindow(t, newReplayBucketService(t), "market-quality.v1")
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("bucket outputs differ across replay runs\nfirst: %+v\nsecond: %+v", first, second)
	}
}

func TestWorldUSABucketConfigVersionPinning(t *testing.T) {
	baseline := replayBucketWindow(t, buildReplayBucketService(t, "market-quality.v1"), "market-quality.v1")
	updated := replayBucketWindow(t, buildReplayBucketService(t, "market-quality.v2"), "market-quality.v2")
	if reflect.DeepEqual(baseline, updated) {
		t.Fatal("expected config version change to alter bucket outputs")
	}
	if baseline[len(baseline)-1].Window.ConfigVersion != "market-quality.v1" {
		t.Fatalf("baseline config version = %q, want market-quality.v1", baseline[len(baseline)-1].Window.ConfigVersion)
	}
	if updated[len(updated)-1].Window.ConfigVersion != "market-quality.v2" {
		t.Fatalf("updated config version = %q, want market-quality.v2", updated[len(updated)-1].Window.ConfigVersion)
	}
}

func newReplayBucketService(t *testing.T) *featureengine.Service {
	t.Helper()
	return buildReplayBucketService(t, "market-quality.v1")
}

func buildReplayBucketService(t *testing.T, configVersion string) *featureengine.Service {
	t.Helper()
	timestampCap := 0.55
	if configVersion == "market-quality.v2" {
		timestampCap = 0.45
	}
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
		ConfigVersion:        configVersion,
		AlgorithmVersion:     "market-quality-buckets.v1",
		TimestampSkewSeconds: 2,
		Families:             map[features.BucketFamily]features.BucketFamilyConfig{features.BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1}, features.BucketFamily2m: {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 0.75}, features.BucketFamily5m: {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 0.8}},
		Thresholds:           features.BucketThresholdConfig{Divergence: map[features.BucketFamily]features.DivergenceThresholds{features.BucketFamily30s: {PriceDistanceModerateBps: 2, PriceDistanceSevereBps: 8, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}, features.BucketFamily2m: {PriceDistanceModerateBps: 3, PriceDistanceSevereBps: 10, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}, features.BucketFamily5m: {PriceDistanceModerateBps: 4, PriceDistanceSevereBps: 12, CoverageGapModerate: 0.2, CoverageGapSevere: 0.4, TimestampFallbackRatioCap: 0.2}}, Quality: features.MarketQualityThresholds{ConcentrationSoftCap: 0.7, ModerateCap: 0.65, SevereCap: 0.35, TimestampTrustCap: timestampCap, IncompleteCap: 0.6}},
	}))
	if err != nil {
		t.Fatalf("new replay bucket service: %v", err)
	}
	return service
}

func replayBucketWindow(t *testing.T, service *featureengine.Service, _ string) []features.MarketQualityBucket {
	t.Helper()
	start := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	var emitted []features.MarketQualityBucket
	for index := 0; index < 10; index++ {
		bucketStart := start.Add(time.Duration(index) * 30 * time.Second)
		exchangeTs := bucketStart.Add(10 * time.Second)
		if index == 2 {
			exchangeTs = time.Time{}
		}
		result, err := service.ObserveWorldUSABucket(features.WorldUSAObservation{
			Symbol:     "ETH-USD",
			ExchangeTs: exchangeTs,
			RecvTs:     bucketStart.Add(11 * time.Second),
			Now:        bucketStart.Add(11 * time.Second),
			World:      replaySnapshot(features.CompositeGroupWorld, 3500+float64(index), 1, 0.99, 0.6, 1, false, ingestion.VenueBinance),
			USA:        replaySnapshot(features.CompositeGroupUSA, 3501+float64(index), 0.5+float64(index%2)/2, 0.97, 0.6, 0, false, ingestion.VenueCoinbase),
		})
		if err != nil {
			t.Fatalf("observe replay bucket %d: %v", index, err)
		}
		emitted = append(emitted, result.Emitted...)
	}
	advanced, err := service.AdvanceWorldUSABuckets("ETH-USD", start.Add(5*time.Minute+33*time.Second))
	if err != nil {
		t.Fatalf("advance replay window: %v", err)
	}
	emitted = append(emitted, advanced...)
	return emitted
}

func replaySnapshot(group features.CompositeGroup, price float64, coverage float64, health float64, maxWeight float64, fallbackCount int, unavailable bool, venue ingestion.Venue) features.CompositeSnapshot {
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
