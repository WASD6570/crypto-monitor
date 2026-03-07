package features

import (
	"fmt"
	"sort"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type BucketFamily string

const (
	BucketFamily30s BucketFamily = "30s"
	BucketFamily2m  BucketFamily = "2m"
	BucketFamily5m  BucketFamily = "5m"
)

type BucketSource string

const (
	BucketSourceExchangeTs  BucketSource = "exchangeTs"
	BucketSourceRecvTs      BucketSource = "recvTs"
	BucketSourceUnavailable BucketSource = "unavailable"
)

type LateEventDisposition string

const (
	LateEventOnTime          LateEventDisposition = "on-time"
	LateEventWithinWatermark LateEventDisposition = "within-watermark"
	LateEventAfterWatermark  LateEventDisposition = "after-watermark"
)

type DirectionAgreement string

const (
	DirectionAgreementAligned DirectionAgreement = "aligned"
	DirectionAgreementMixed   DirectionAgreement = "mixed"
	DirectionAgreementOpposed DirectionAgreement = "opposed"
	DirectionAgreementFlat    DirectionAgreement = "flat"
	DirectionAgreementUnknown DirectionAgreement = "unknown"
)

type FragmentationSeverity string

const (
	FragmentationSeverityLow      FragmentationSeverity = "low"
	FragmentationSeverityModerate FragmentationSeverity = "moderate"
	FragmentationSeveritySevere   FragmentationSeverity = "severe"
)

type BucketConfig struct {
	SchemaVersion        string                              `json:"schemaVersion"`
	ConfigVersion        string                              `json:"configVersion"`
	AlgorithmVersion     string                              `json:"algorithmVersion"`
	TimestampSkewSeconds int                                 `json:"timestampSkewSeconds"`
	Families             map[BucketFamily]BucketFamilyConfig `json:"families"`
	Thresholds           BucketThresholdConfig               `json:"thresholds"`
}

type BucketFamilyConfig struct {
	IntervalSeconds     int     `json:"intervalSeconds"`
	WatermarkSeconds    int     `json:"watermarkSeconds"`
	MinimumCompleteness float64 `json:"minimumCompleteness"`
	ExpectedBucketCount int     `json:"expectedBucketCount,omitempty"`
}

type BucketThresholdConfig struct {
	Divergence map[BucketFamily]DivergenceThresholds `json:"divergence"`
	Quality    MarketQualityThresholds               `json:"quality"`
}

type DivergenceThresholds struct {
	PriceDistanceModerateBps  float64 `json:"priceDistanceModerateBps"`
	PriceDistanceSevereBps    float64 `json:"priceDistanceSevereBps"`
	CoverageGapModerate       float64 `json:"coverageGapModerate"`
	CoverageGapSevere         float64 `json:"coverageGapSevere"`
	TimestampFallbackRatioCap float64 `json:"timestampFallbackRatioCap"`
}

type MarketQualityThresholds struct {
	ConcentrationSoftCap float64 `json:"concentrationSoftCap"`
	ModerateCap          float64 `json:"moderateCap"`
	SevereCap            float64 `json:"severeCap"`
	TimestampTrustCap    float64 `json:"timestampTrustCap"`
	IncompleteCap        float64 `json:"incompleteCap"`
}

type BucketAssignment struct {
	Symbol             string               `json:"symbol"`
	Family             BucketFamily         `json:"family"`
	BucketStart        string               `json:"bucketStart"`
	BucketEnd          string               `json:"bucketEnd"`
	BucketSource       BucketSource         `json:"bucketSource"`
	TimestampDegraded  bool                 `json:"timestampDegraded"`
	LateDisposition    LateEventDisposition `json:"lateDisposition"`
	ExchangeRecvSkewMs int64                `json:"exchangeRecvSkewMs"`
}

type BucketWindowSummary struct {
	Family                          BucketFamily `json:"family"`
	Start                           string       `json:"start"`
	End                             string       `json:"end"`
	ClosedBucketCount               int          `json:"closedBucketCount"`
	ExpectedBucketCount             int          `json:"expectedBucketCount"`
	MissingBucketCount              int          `json:"missingBucketCount"`
	ContainsAfterWatermarkLateEvent bool         `json:"containsAfterWatermarkLateEvent"`
	ConfigVersion                   string       `json:"configVersion"`
	AlgorithmVersion                string       `json:"algorithmVersion"`
}

type CompositeBucketSide struct {
	Available                         bool    `json:"available"`
	CoverageRatio                     float64 `json:"coverageRatio"`
	HealthScore                       float64 `json:"healthScore"`
	MaxContributorWeight              float64 `json:"maxContributorWeight"`
	TimestampFallbackContributorCount int     `json:"timestampFallbackContributorCount"`
	ConfiguredContributorCount        int     `json:"configuredContributorCount"`
	EligibleContributorCount          int     `json:"eligibleContributorCount"`
	ContributingContributorCount      int     `json:"contributingContributorCount"`
	Unavailable                       bool    `json:"unavailable"`
}

type DivergenceSummary struct {
	Available           bool               `json:"available"`
	PriceDistanceBps    float64            `json:"priceDistanceBps"`
	DirectionAgreement  DirectionAgreement `json:"directionAgreement"`
	ParticipationGap    float64            `json:"participationGap"`
	LeaderChurnDetected bool               `json:"leaderChurnDetected"`
	ReasonCodes         []ReasonCode       `json:"reasonCodes,omitempty"`
}

type FragmentationSummary struct {
	Severity             FragmentationSeverity `json:"severity"`
	PersistenceCount     int                   `json:"persistenceCount"`
	PrimaryCauses        []ReasonCode          `json:"primaryCauses,omitempty"`
	UnavailableSideCount int                   `json:"unavailableSideCount"`
}

type TimestampTrustSummary struct {
	ExchangeBucketCount int     `json:"exchangeBucketCount"`
	RecvFallbackCount   int     `json:"recvFallbackCount"`
	FallbackRatio       float64 `json:"fallbackRatio"`
	WorldFallbackCount  int     `json:"worldFallbackCount"`
	USAFallbackCount    int     `json:"usaFallbackCount"`
	OneSidedFallback    bool    `json:"oneSidedFallback"`
	TrustCap            bool    `json:"trustCap"`
}

type MarketQualitySummary struct {
	WorldQualityCap   float64      `json:"worldQualityCap"`
	USAQualityCap     float64      `json:"usaQualityCap"`
	CombinedTrustCap  float64      `json:"combinedTrustCap"`
	DowngradedReasons []ReasonCode `json:"downgradedReasons,omitempty"`
	ReplayProvenance  string       `json:"replayProvenance"`
}

type MarketQualityBucket struct {
	SchemaVersion  string                `json:"schemaVersion"`
	Symbol         string                `json:"symbol"`
	Window         BucketWindowSummary   `json:"window"`
	Assignment     BucketAssignment      `json:"assignment"`
	World          CompositeBucketSide   `json:"world"`
	USA            CompositeBucketSide   `json:"usa"`
	Divergence     DivergenceSummary     `json:"divergence"`
	Fragmentation  FragmentationSummary  `json:"fragmentation"`
	TimestampTrust TimestampTrustSummary `json:"timestampTrust"`
	MarketQuality  MarketQualitySummary  `json:"marketQuality"`

	worldPrice                 *float64
	usaPrice                   *float64
	worldLeaderKey             string
	usaLeaderKey               string
	containsAfterWatermarkLate bool
}

type WorldUSAObservation struct {
	Symbol     string            `json:"symbol"`
	ExchangeTs time.Time         `json:"exchangeTs"`
	RecvTs     time.Time         `json:"recvTs"`
	World      CompositeSnapshot `json:"world"`
	USA        CompositeSnapshot `json:"usa"`
	Now        time.Time         `json:"now"`
}

type ObservationResult struct {
	Assignment BucketAssignment      `json:"assignment"`
	Accepted   bool                  `json:"accepted"`
	Emitted    []MarketQualityBucket `json:"emitted"`
}

type WorldUSABucketProcessor struct {
	config       BucketConfig
	pending      map[string]map[time.Time]MarketQualityBucket
	closed30s    map[string][]MarketQualityBucket
	lastClosed30 map[string]time.Time
	lateSeen     map[string]map[time.Time]bool
}

func (c BucketConfig) Validate() error {
	if c.SchemaVersion == "" || c.ConfigVersion == "" || c.AlgorithmVersion == "" {
		return fmt.Errorf("bucket config version fields are required")
	}
	if c.TimestampSkewSeconds <= 0 {
		return fmt.Errorf("timestampSkewSeconds must be positive")
	}
	for _, family := range []BucketFamily{BucketFamily30s, BucketFamily2m, BucketFamily5m} {
		cfg, ok := c.Families[family]
		if !ok {
			return fmt.Errorf("family %q config is required", family)
		}
		if cfg.IntervalSeconds <= 0 || cfg.WatermarkSeconds <= 0 {
			return fmt.Errorf("family %q interval and watermark must be positive", family)
		}
		if cfg.MinimumCompleteness <= 0 || cfg.MinimumCompleteness > 1 {
			return fmt.Errorf("family %q minimumCompleteness must be within (0,1]", family)
		}
	}
	for _, family := range []BucketFamily{BucketFamily30s, BucketFamily2m, BucketFamily5m} {
		thresholds, ok := c.Thresholds.Divergence[family]
		if !ok {
			return fmt.Errorf("family %q divergence thresholds are required", family)
		}
		if thresholds.PriceDistanceModerateBps > thresholds.PriceDistanceSevereBps {
			return fmt.Errorf("family %q moderate distance threshold must be <= severe threshold", family)
		}
	}
	return nil
}

func NewWorldUSABucketProcessor(config BucketConfig) (*WorldUSABucketProcessor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &WorldUSABucketProcessor{
		config:       config,
		pending:      map[string]map[time.Time]MarketQualityBucket{},
		closed30s:    map[string][]MarketQualityBucket{},
		lastClosed30: map[string]time.Time{},
		lateSeen:     map[string]map[time.Time]bool{},
	}, nil
}

func AssignBucket(config BucketConfig, symbol string, family BucketFamily, exchangeTs, recvTs, now time.Time) (BucketAssignment, error) {
	if err := config.Validate(); err != nil {
		return BucketAssignment{}, err
	}
	if symbol == "" {
		return BucketAssignment{}, fmt.Errorf("symbol is required")
	}
	familyConfig, ok := config.Families[family]
	if !ok {
		return BucketAssignment{}, fmt.Errorf("family %q config missing", family)
	}
	canonical, err := ingestion.ResolveCanonicalTimestamp(exchangeTs, recvTs, ingestion.TimestampPolicy{MaxExchangeRecvSkew: time.Duration(config.TimestampSkewSeconds) * time.Second})
	if err != nil {
		return BucketAssignment{}, err
	}
	bucketStart := floorTime(canonical.EventTime.UTC(), time.Duration(familyConfig.IntervalSeconds)*time.Second)
	bucketEnd := bucketStart.Add(time.Duration(familyConfig.IntervalSeconds) * time.Second)
	disposition := LateEventOnTime
	if now.After(bucketEnd) {
		if now.After(bucketEnd.Add(time.Duration(familyConfig.WatermarkSeconds) * time.Second)) {
			disposition = LateEventAfterWatermark
		} else {
			disposition = LateEventWithinWatermark
		}
	}
	source := BucketSourceExchangeTs
	if canonical.Status == ingestion.TimestampStatusDegraded {
		source = BucketSourceRecvTs
	}
	return BucketAssignment{
		Symbol:             symbol,
		Family:             family,
		BucketStart:        bucketStart.Format(time.RFC3339Nano),
		BucketEnd:          bucketEnd.Format(time.RFC3339Nano),
		BucketSource:       source,
		TimestampDegraded:  canonical.Status == ingestion.TimestampStatusDegraded,
		LateDisposition:    disposition,
		ExchangeRecvSkewMs: canonical.ExchangeRecvSkew.Milliseconds(),
	}, nil
}

func (p *WorldUSABucketProcessor) Observe(observation WorldUSAObservation) (ObservationResult, error) {
	if p == nil {
		return ObservationResult{}, fmt.Errorf("bucket processor is required")
	}
	assignment, err := AssignBucket(p.config, observation.Symbol, BucketFamily30s, observation.ExchangeTs, observation.RecvTs, observation.Now)
	if err != nil {
		return ObservationResult{}, err
	}
	result := ObservationResult{Assignment: assignment}
	bucketStart, err := time.Parse(time.RFC3339Nano, assignment.BucketStart)
	if err != nil {
		return ObservationResult{}, err
	}
	if assignment.LateDisposition == LateEventAfterWatermark {
		if p.lateSeen[observation.Symbol] == nil {
			p.lateSeen[observation.Symbol] = map[time.Time]bool{}
		}
		p.lateSeen[observation.Symbol][bucketStart] = true
		result.Accepted = false
		result.Emitted = p.advance(observation.Symbol, observation.Now)
		return result, nil
	}
	if p.pending[observation.Symbol] == nil {
		p.pending[observation.Symbol] = map[time.Time]MarketQualityBucket{}
	}
	bucket := summarize30sBucket(p.config, assignment, observation.World, observation.USA)
	p.pending[observation.Symbol][bucketStart] = bucket
	result.Accepted = true
	result.Emitted = p.advance(observation.Symbol, observation.Now)
	return result, nil
}

func (p *WorldUSABucketProcessor) Advance(symbol string, now time.Time) []MarketQualityBucket {
	if p == nil {
		return nil
	}
	return p.advance(symbol, now)
}

func (p *WorldUSABucketProcessor) advance(symbol string, now time.Time) []MarketQualityBucket {
	pending := p.pending[symbol]
	if len(pending) == 0 {
		return nil
	}
	starts := make([]time.Time, 0, len(pending))
	for start := range pending {
		starts = append(starts, start)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].Before(starts[j]) })
	emitted := make([]MarketQualityBucket, 0)
	for _, start := range starts {
		bucket := pending[start]
		end, _ := time.Parse(time.RFC3339Nano, bucket.Window.End)
		watermark := time.Duration(p.config.Families[BucketFamily30s].WatermarkSeconds) * time.Second
		if now.Before(end.Add(watermark)) {
			continue
		}
		lastClosed := p.lastClosed30[symbol]
		for !lastClosed.IsZero() && lastClosed.Add(30*time.Second).Before(start) {
			missingStart := lastClosed.Add(30 * time.Second)
			placeholder := missing30sBucket(p.config, symbol, missingStart, p.lateSeen[symbol][missingStart])
			emitted = append(emitted, p.close30s(symbol, missingStart, placeholder)...)
			lastClosed = missingStart
		}
		bucket.containsAfterWatermarkLate = p.lateSeen[symbol][start]
		emitted = append(emitted, p.close30s(symbol, start, bucket)...)
		delete(p.pending[symbol], start)
	}
	return emitted
}

func (p *WorldUSABucketProcessor) close30s(symbol string, start time.Time, bucket MarketQualityBucket) []MarketQualityBucket {
	bucket.Window.ContainsAfterWatermarkLateEvent = bucket.containsAfterWatermarkLate
	p.closed30s[symbol] = append(p.closed30s[symbol], bucket)
	p.lastClosed30[symbol] = start
	emitted := []MarketQualityBucket{bucket}
	if isRollupBoundary(start.Add(30*time.Second), BucketFamily2m) {
		emitted = append(emitted, buildRollupFromClosed(p.config, symbol, BucketFamily2m, p.closed30s[symbol]))
	}
	if isRollupBoundary(start.Add(30*time.Second), BucketFamily5m) {
		emitted = append(emitted, buildRollupFromClosed(p.config, symbol, BucketFamily5m, p.closed30s[symbol]))
	}
	return emitted
}

func summarize30sBucket(config BucketConfig, assignment BucketAssignment, world CompositeSnapshot, usa CompositeSnapshot) MarketQualityBucket {
	window := BucketWindowSummary{
		Family:              BucketFamily30s,
		Start:               assignment.BucketStart,
		End:                 assignment.BucketEnd,
		ClosedBucketCount:   1,
		ExpectedBucketCount: 1,
		ConfigVersion:       config.ConfigVersion,
		AlgorithmVersion:    config.AlgorithmVersion,
	}
	bucket := MarketQualityBucket{
		SchemaVersion:  config.SchemaVersion,
		Symbol:         assignment.Symbol,
		Window:         window,
		Assignment:     assignment,
		World:          summarizeSide(world),
		USA:            summarizeSide(usa),
		worldPrice:     world.CompositePrice,
		usaPrice:       usa.CompositePrice,
		worldLeaderKey: leaderKey(world),
		usaLeaderKey:   leaderKey(usa),
	}
	applyDerivedSummaries(config, BucketFamily30s, &bucket)
	return bucket
}

func missing30sBucket(config BucketConfig, symbol string, start time.Time, containsLate bool) MarketQualityBucket {
	end := start.Add(30 * time.Second)
	bucket := MarketQualityBucket{
		SchemaVersion: config.SchemaVersion,
		Symbol:        symbol,
		Window: BucketWindowSummary{
			Family:                          BucketFamily30s,
			Start:                           start.Format(time.RFC3339Nano),
			End:                             end.Format(time.RFC3339Nano),
			ClosedBucketCount:               0,
			ExpectedBucketCount:             1,
			MissingBucketCount:              1,
			ContainsAfterWatermarkLateEvent: containsLate,
			ConfigVersion:                   config.ConfigVersion,
			AlgorithmVersion:                config.AlgorithmVersion,
		},
		Assignment: BucketAssignment{
			Symbol:            symbol,
			Family:            BucketFamily30s,
			BucketStart:       start.Format(time.RFC3339Nano),
			BucketEnd:         end.Format(time.RFC3339Nano),
			BucketSource:      BucketSourceUnavailable,
			LateDisposition:   LateEventOnTime,
			TimestampDegraded: false,
		},
		World: CompositeBucketSide{Unavailable: true},
		USA:   CompositeBucketSide{Unavailable: true},
		Divergence: DivergenceSummary{
			Available:          false,
			DirectionAgreement: DirectionAgreementUnknown,
			ReasonCodes:        []ReasonCode{ReasonCompositeUnavailable},
		},
		Fragmentation: FragmentationSummary{
			Severity:             FragmentationSeveritySevere,
			PersistenceCount:     1,
			PrimaryCauses:        []ReasonCode{ReasonCompositeUnavailable},
			UnavailableSideCount: 2,
		},
		MarketQuality: MarketQualitySummary{
			WorldQualityCap:   0,
			USAQualityCap:     0,
			CombinedTrustCap:  0,
			DowngradedReasons: []ReasonCode{ReasonCompositeUnavailable},
			ReplayProvenance:  config.ConfigVersion,
		},
		containsAfterWatermarkLate: containsLate,
	}
	return bucket
}

func buildRollupFromClosed(config BucketConfig, symbol string, family BucketFamily, closed []MarketQualityBucket) MarketQualityBucket {
	expected := rollupBucketCount(family)
	components := lastClosedBuckets(closed, expected)
	rollup := MarketQualityBucket{
		SchemaVersion: config.SchemaVersion,
		Symbol:        symbol,
		Window: BucketWindowSummary{
			Family:              family,
			ClosedBucketCount:   len(components),
			ExpectedBucketCount: expected,
			ConfigVersion:       config.ConfigVersion,
			AlgorithmVersion:    config.AlgorithmVersion,
		},
		Assignment: BucketAssignment{Symbol: symbol, Family: family},
	}
	if len(components) == 0 {
		return rollup
	}
	start := components[0].Window.Start
	end := components[len(components)-1].Window.End
	rollup.Window.Start = start
	rollup.Window.End = end
	rollup.Assignment.BucketStart = start
	rollup.Assignment.BucketEnd = end
	rollup.Assignment.BucketSource = dominantSource(components)
	applyDerivedSummaries(config, family, bucketPointers(components, &rollup)...)
	return rollup
}

func bucketPointers(components []MarketQualityBucket, rollup *MarketQualityBucket) []*MarketQualityBucket {
	pointers := make([]*MarketQualityBucket, 0, len(components)+1)
	for index := range components {
		component := components[index]
		pointers = append(pointers, &component)
	}
	pointers = append(pointers, rollup)
	return pointers
}

func applyDerivedSummaries(config BucketConfig, family BucketFamily, buckets ...*MarketQualityBucket) {
	if len(buckets) == 0 {
		return
	}
	if len(buckets) == 1 {
		bucket := buckets[0]
		bucket.Divergence = summarizeDivergence([]*MarketQualityBucket{bucket})
		bucket.TimestampTrust = summarizeTimestampTrust(config, []MarketQualityBucket{*bucket})
		bucket.Fragmentation = summarizeFragmentation(config, family, bucket.Divergence, bucket.TimestampTrust, 1)
		bucket.MarketQuality = summarizeMarketQuality(config, bucket.World, bucket.USA, bucket.Fragmentation, bucket.TimestampTrust, bucket.Window)
		return
	}
	rollup := buckets[len(buckets)-1]
	components := make([]MarketQualityBucket, 0, len(buckets)-1)
	for _, bucket := range buckets[:len(buckets)-1] {
		components = append(components, *bucket)
	}
	rollup.World = aggregateSide(components, true)
	rollup.USA = aggregateSide(components, false)
	rollup.Divergence = summarizeDivergence(buckets[:len(buckets)-1])
	rollup.TimestampTrust = summarizeTimestampTrust(config, components)
	rollup.Window.MissingBucketCount = missingBucketCount(components)
	rollup.Window.ClosedBucketCount = len(components) - rollup.Window.MissingBucketCount
	rollup.Window.ContainsAfterWatermarkLateEvent = containsLate(components)
	rollup.Fragmentation = summarizeFragmentation(config, family, rollup.Divergence, rollup.TimestampTrust, fragmentationPersistence(components))
	if rollup.Window.MissingBucketCount > 0 {
		rollup.Fragmentation.Severity = FragmentationSeveritySevere
		rollup.Fragmentation.PrimaryCauses = appendReason(rollup.Fragmentation.PrimaryCauses, ReasonMissingInput)
	}
	rollup.MarketQuality = summarizeMarketQuality(config, rollup.World, rollup.USA, rollup.Fragmentation, rollup.TimestampTrust, rollup.Window)
	rollup.Assignment.TimestampDegraded = rollup.TimestampTrust.RecvFallbackCount > 0
	rollup.Assignment.LateDisposition = LateEventOnTime
	if rollup.TimestampTrust.RecvFallbackCount > 0 {
		rollup.Assignment.BucketSource = BucketSourceRecvTs
	}
	rollup.worldPrice = latestPrice(components, true)
	rollup.usaPrice = latestPrice(components, false)
	rollup.worldLeaderKey = latestLeader(components, true)
	rollup.usaLeaderKey = latestLeader(components, false)
}

func summarizeSide(snapshot CompositeSnapshot) CompositeBucketSide {
	return CompositeBucketSide{
		Available:                         !snapshot.Unavailable && snapshot.CompositePrice != nil,
		CoverageRatio:                     snapshot.CoverageRatio,
		HealthScore:                       snapshot.HealthScore,
		MaxContributorWeight:              snapshot.MaxContributorWeight,
		TimestampFallbackContributorCount: snapshot.TimestampFallbackContributorCount,
		ConfiguredContributorCount:        snapshot.ConfiguredContributorCount,
		EligibleContributorCount:          snapshot.EligibleContributorCount,
		ContributingContributorCount:      snapshot.ContributingContributorCount,
		Unavailable:                       snapshot.Unavailable,
	}
}

func leaderKey(snapshot CompositeSnapshot) string {
	bestWeight := -1.0
	best := ""
	for _, contributor := range snapshot.Contributors {
		if contributor.FinalWeight > bestWeight {
			bestWeight = contributor.FinalWeight
			best = memberKey(contributor.Venue, contributor.MarketType)
		}
	}
	return best
}

func floorTime(ts time.Time, interval time.Duration) time.Time {
	if interval <= 0 {
		return ts.UTC()
	}
	return ts.UTC().Truncate(interval)
}

func isRollupBoundary(end time.Time, family BucketFamily) bool {
	end = end.UTC()
	switch family {
	case BucketFamily2m:
		return end.Second() == 0 && end.Minute()%2 == 0
	case BucketFamily5m:
		return end.Second() == 0 && end.Minute()%5 == 0
	default:
		return false
	}
}

func rollupBucketCount(family BucketFamily) int {
	switch family {
	case BucketFamily2m:
		return 4
	case BucketFamily5m:
		return 10
	default:
		return 1
	}
}

func lastClosedBuckets(closed []MarketQualityBucket, count int) []MarketQualityBucket {
	if len(closed) <= count {
		return append([]MarketQualityBucket(nil), closed...)
	}
	return append([]MarketQualityBucket(nil), closed[len(closed)-count:]...)
}

func dominantSource(buckets []MarketQualityBucket) BucketSource {
	recv := 0
	exchange := 0
	for _, bucket := range buckets {
		switch bucket.Assignment.BucketSource {
		case BucketSourceRecvTs:
			recv++
		case BucketSourceExchangeTs:
			exchange++
		}
	}
	if recv > 0 && recv >= exchange {
		return BucketSourceRecvTs
	}
	if exchange > 0 {
		return BucketSourceExchangeTs
	}
	return BucketSourceUnavailable
}

func containsLate(buckets []MarketQualityBucket) bool {
	for _, bucket := range buckets {
		if bucket.Window.ContainsAfterWatermarkLateEvent {
			return true
		}
	}
	return false
}

func latestPrice(buckets []MarketQualityBucket, world bool) *float64 {
	for index := len(buckets) - 1; index >= 0; index-- {
		if world && buckets[index].worldPrice != nil {
			return buckets[index].worldPrice
		}
		if !world && buckets[index].usaPrice != nil {
			return buckets[index].usaPrice
		}
	}
	return nil
}

func latestLeader(buckets []MarketQualityBucket, world bool) string {
	for index := len(buckets) - 1; index >= 0; index-- {
		if world && buckets[index].worldLeaderKey != "" {
			return buckets[index].worldLeaderKey
		}
		if !world && buckets[index].usaLeaderKey != "" {
			return buckets[index].usaLeaderKey
		}
	}
	return ""
}

func aggregateSide(buckets []MarketQualityBucket, world bool) CompositeBucketSide {
	count := 0.0
	summary := CompositeBucketSide{}
	for _, bucket := range buckets {
		side := bucket.USA
		if world {
			side = bucket.World
		}
		if side.ConfiguredContributorCount > summary.ConfiguredContributorCount {
			summary.ConfiguredContributorCount = side.ConfiguredContributorCount
		}
		summary.EligibleContributorCount += side.EligibleContributorCount
		summary.ContributingContributorCount += side.ContributingContributorCount
		summary.TimestampFallbackContributorCount += side.TimestampFallbackContributorCount
		summary.Unavailable = summary.Unavailable || side.Unavailable
		if side.Available {
			summary.Available = true
			summary.CoverageRatio += side.CoverageRatio
			summary.HealthScore += side.HealthScore
			if side.MaxContributorWeight > summary.MaxContributorWeight {
				summary.MaxContributorWeight = side.MaxContributorWeight
			}
			count++
		}
	}
	if count > 0 {
		summary.CoverageRatio = roundMetric(summary.CoverageRatio / count)
		summary.HealthScore = roundMetric(summary.HealthScore / count)
	}
	return summary
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func missingBucketCount(buckets []MarketQualityBucket) int {
	count := 0
	for _, bucket := range buckets {
		count += bucket.Window.MissingBucketCount
	}
	return count
}
