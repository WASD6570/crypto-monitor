package features

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const (
	SpotTradeFlowBucketSchemaVersion = "v1"
	SpotTradeFlowMarketType          = "spot"
	SpotTradeFlowSideBuy             = "buy"
	SpotTradeFlowSideSell            = "sell"
)

var (
	spotTradeFlowTrackedSymbols = []string{"BTC-USD", "ETH-USD"}
	spotTradeFlowSourceSymbols  = map[string]string{"BTC-USD": "BTCUSDT", "ETH-USD": "ETHUSDT"}
	spotTradeFlowFamilyOrder    = []BucketFamily{BucketFamily30s, BucketFamily2m, BucketFamily5m}
)

type SpotTradeFlowConfig struct {
	SchemaVersion        string                              `json:"schemaVersion"`
	ConfigVersion        string                              `json:"configVersion"`
	AlgorithmVersion     string                              `json:"algorithmVersion"`
	TimestampSkewSeconds int                                 `json:"timestampSkewSeconds"`
	Symbols              []string                            `json:"symbols"`
	SourceSymbols        map[string]string                   `json:"sourceSymbols"`
	Families             map[BucketFamily]BucketFamilyConfig `json:"families"`
}

type SpotTradeFlowObservation struct {
	Symbol            string                             `json:"symbol"`
	Venue             ingestion.Venue                    `json:"venue"`
	MarketType        string                             `json:"marketType"`
	SourceSymbol      string                             `json:"sourceSymbol"`
	SourceRecordID    string                             `json:"sourceRecordId"`
	Side              string                             `json:"side"`
	Price             float64                            `json:"price"`
	Size              float64                            `json:"size"`
	ExchangeTs        time.Time                          `json:"exchangeTs"`
	RecvTs            time.Time                          `json:"recvTs"`
	ObservedAt        time.Time                          `json:"observedAt,omitempty"`
	TimestampStatus   ingestion.CanonicalTimestampStatus `json:"timestampStatus"`
	FeedHealthState   ingestion.FeedHealthState          `json:"feedHealthState"`
	FeedHealthReasons []ingestion.DegradationReason      `json:"feedHealthReasons,omitempty"`
}

type SpotTradeFlowBucket struct {
	SchemaVersion          string                        `json:"schemaVersion"`
	Symbol                 string                        `json:"symbol"`
	Family                 BucketFamily                  `json:"family"`
	BucketStart            string                        `json:"bucketStart"`
	BucketEnd              string                        `json:"bucketEnd"`
	BucketSource           BucketSource                  `json:"bucketSource"`
	LateDisposition        LateEventDisposition          `json:"lateDisposition"`
	TradeCount             int                           `json:"tradeCount"`
	DuplicateCount         int                           `json:"duplicateCount"`
	BuyTradeCount          int                           `json:"buyTradeCount"`
	SellTradeCount         int                           `json:"sellTradeCount"`
	BuyNotional            float64                       `json:"buyNotional"`
	SellNotional           float64                       `json:"sellNotional"`
	NetAggressorNotional   float64                       `json:"netAggressorNotional"`
	TotalNotional          float64                       `json:"totalNotional"`
	VWAP                   float64                       `json:"vwap"`
	FirstPrice             float64                       `json:"firstPrice"`
	LastPrice              float64                       `json:"lastPrice"`
	PriceChangeBps         float64                       `json:"priceChangeBps"`
	TimestampFallbackCount int                           `json:"timestampFallbackCount"`
	FeedHealthState        ingestion.FeedHealthState     `json:"feedHealthState"`
	FeedHealthReasons      []ingestion.DegradationReason `json:"feedHealthReasons,omitempty"`
	ConfigVersion          string                        `json:"configVersion"`
	AlgorithmVersion       string                        `json:"algorithmVersion"`
}

type SpotTradeFlowObservationResult struct {
	Assignment BucketAssignment      `json:"assignment"`
	Accepted   bool                  `json:"accepted"`
	Duplicate  bool                  `json:"duplicate"`
	Buckets    []SpotTradeFlowBucket `json:"buckets"`
}

type SpotTradeFlowProcessor struct {
	config      SpotTradeFlowConfig
	buckets     map[spotTradeFlowBucketKey]*spotTradeFlowBucketState
	sourceIndex map[string][]spotTradeFlowBucketKey
}

type spotTradeFlowBucketKey struct {
	Symbol string
	Family BucketFamily
	Start  time.Time
}

type spotTradeFlowBucketState struct {
	assignment     BucketAssignment
	trades         map[string]SpotTradeFlowObservation
	duplicateCount int
}

func DefaultSpotTradeFlowConfig() SpotTradeFlowConfig {
	return SpotTradeFlowConfig{
		SchemaVersion:        SpotTradeFlowBucketSchemaVersion,
		ConfigVersion:        "feature-engine.binance-spot-trade-flow.v1",
		AlgorithmVersion:     "binance-spot-trade-flow-buckets.v1",
		TimestampSkewSeconds: 2,
		Symbols:              SpotTradeFlowTrackedSymbols(),
		SourceSymbols:        SpotTradeFlowSourceSymbols(),
		Families: map[BucketFamily]BucketFamilyConfig{
			BucketFamily30s: {IntervalSeconds: 30, WatermarkSeconds: 2, MinimumCompleteness: 1},
			BucketFamily2m:  {IntervalSeconds: 120, WatermarkSeconds: 5, MinimumCompleteness: 1},
			BucketFamily5m:  {IntervalSeconds: 300, WatermarkSeconds: 10, MinimumCompleteness: 1},
		},
	}
}

func SpotTradeFlowTrackedSymbols() []string {
	return append([]string(nil), spotTradeFlowTrackedSymbols...)
}

func SpotTradeFlowSourceSymbols() map[string]string {
	result := make(map[string]string, len(spotTradeFlowSourceSymbols))
	for symbol, source := range spotTradeFlowSourceSymbols {
		result[symbol] = source
	}
	return result
}

func ValidateSpotTradeFlowSymbols(symbols []string) error {
	if !slices.Equal(symbols, spotTradeFlowTrackedSymbols) {
		return fmt.Errorf("spot trade-flow symbols must equal %v", spotTradeFlowTrackedSymbols)
	}
	return nil
}

func (c SpotTradeFlowConfig) Validate() error {
	if c.SchemaVersion == "" || c.ConfigVersion == "" || c.AlgorithmVersion == "" {
		return fmt.Errorf("spot trade-flow config version fields are required")
	}
	if c.TimestampSkewSeconds <= 0 {
		return fmt.Errorf("spot trade-flow timestampSkewSeconds must be positive")
	}
	if err := ValidateSpotTradeFlowSymbols(c.Symbols); err != nil {
		return err
	}
	for _, symbol := range spotTradeFlowTrackedSymbols {
		if c.SourceSymbols[symbol] != spotTradeFlowSourceSymbols[symbol] {
			return fmt.Errorf("spot trade-flow source symbol for %q must be %q", symbol, spotTradeFlowSourceSymbols[symbol])
		}
	}
	for _, family := range spotTradeFlowFamilyOrder {
		familyConfig, ok := c.Families[family]
		if !ok {
			return fmt.Errorf("spot trade-flow family %q config is required", family)
		}
		if familyConfig.IntervalSeconds <= 0 || familyConfig.WatermarkSeconds <= 0 {
			return fmt.Errorf("spot trade-flow family %q interval and watermark must be positive", family)
		}
	}
	return nil
}

func NewSpotTradeFlowProcessor(config SpotTradeFlowConfig) (*SpotTradeFlowProcessor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &SpotTradeFlowProcessor{
		config:      config,
		buckets:     map[spotTradeFlowBucketKey]*spotTradeFlowBucketState{},
		sourceIndex: map[string][]spotTradeFlowBucketKey{},
	}, nil
}

func (p *SpotTradeFlowProcessor) Observe(observation SpotTradeFlowObservation) (SpotTradeFlowObservationResult, error) {
	if p == nil {
		return SpotTradeFlowObservationResult{}, fmt.Errorf("spot trade-flow processor is required")
	}
	if err := observation.validate(p.config); err != nil {
		return SpotTradeFlowObservationResult{}, err
	}

	assignments := make(map[BucketFamily]BucketAssignment, len(spotTradeFlowFamilyOrder))
	for _, family := range spotTradeFlowFamilyOrder {
		assignment, err := assignSpotTradeFlowBucket(p.config, observation, family)
		if err != nil {
			return SpotTradeFlowObservationResult{}, err
		}
		assignments[family] = assignment
	}

	primary := assignments[BucketFamily30s]
	if primary.LateDisposition == LateEventAfterWatermark {
		return SpotTradeFlowObservationResult{Assignment: primary, Accepted: false, Buckets: p.Snapshot(observation.Symbol)}, nil
	}

	sourceKey := spotTradeFlowSourceKey(observation)
	if keys, ok := p.sourceIndex[sourceKey]; ok {
		for _, key := range keys {
			if bucket := p.buckets[key]; bucket != nil {
				bucket.duplicateCount++
			}
		}
		return SpotTradeFlowObservationResult{Assignment: primary, Accepted: false, Duplicate: true, Buckets: p.Snapshot(observation.Symbol)}, nil
	}

	keys := make([]spotTradeFlowBucketKey, 0, len(spotTradeFlowFamilyOrder))
	for _, family := range spotTradeFlowFamilyOrder {
		assignment := assignments[family]
		start, err := time.Parse(time.RFC3339Nano, assignment.BucketStart)
		if err != nil {
			return SpotTradeFlowObservationResult{}, fmt.Errorf("parse spot trade-flow bucket start: %w", err)
		}
		key := spotTradeFlowBucketKey{Symbol: observation.Symbol, Family: family, Start: start}
		bucket := p.buckets[key]
		if bucket == nil {
			bucket = &spotTradeFlowBucketState{assignment: assignment, trades: map[string]SpotTradeFlowObservation{}}
			p.buckets[key] = bucket
		}
		bucket.trades[observation.SourceRecordID] = observation
		keys = append(keys, key)
	}
	p.sourceIndex[sourceKey] = keys

	return SpotTradeFlowObservationResult{Assignment: primary, Accepted: true, Buckets: p.Snapshot(observation.Symbol)}, nil
}

func (p *SpotTradeFlowProcessor) Snapshot(symbols ...string) []SpotTradeFlowBucket {
	if p == nil || len(p.buckets) == 0 {
		return nil
	}
	selected := map[string]struct{}{}
	if len(symbols) == 0 {
		for _, symbol := range p.config.Symbols {
			selected[symbol] = struct{}{}
		}
	} else {
		for _, symbol := range symbols {
			selected[symbol] = struct{}{}
		}
	}

	buckets := make([]SpotTradeFlowBucket, 0, len(p.buckets))
	for key, state := range p.buckets {
		if _, ok := selected[key.Symbol]; !ok {
			continue
		}
		buckets = append(buckets, state.snapshot(p.config))
	}
	sort.Slice(buckets, func(i, j int) bool {
		return compareSpotTradeFlowBuckets(buckets[i], buckets[j]) < 0
	})
	return buckets
}

func (o SpotTradeFlowObservation) validate(config SpotTradeFlowConfig) error {
	if o.Symbol == "" {
		return fmt.Errorf("spot trade-flow symbol is required")
	}
	expectedSource, ok := config.SourceSymbols[o.Symbol]
	if !ok {
		return fmt.Errorf("unsupported spot trade-flow symbol %q", o.Symbol)
	}
	if o.SourceSymbol != expectedSource {
		return fmt.Errorf("spot trade-flow source symbol for %q must be %q", o.Symbol, expectedSource)
	}
	if o.Venue != ingestion.VenueBinance {
		return fmt.Errorf("spot trade-flow venue must be %q", ingestion.VenueBinance)
	}
	if o.MarketType != SpotTradeFlowMarketType {
		return fmt.Errorf("spot trade-flow market type must be %q", SpotTradeFlowMarketType)
	}
	if o.SourceRecordID == "" {
		return fmt.Errorf("spot trade-flow source record id is required")
	}
	if o.Side != SpotTradeFlowSideBuy && o.Side != SpotTradeFlowSideSell {
		return fmt.Errorf("unsupported spot trade-flow side %q", o.Side)
	}
	if o.Price <= 0 {
		return fmt.Errorf("spot trade-flow price must be positive")
	}
	if o.Size <= 0 {
		return fmt.Errorf("spot trade-flow size must be positive")
	}
	if o.RecvTs.IsZero() {
		return fmt.Errorf("spot trade-flow recv timestamp is required")
	}
	if o.TimestampStatus != ingestion.TimestampStatusNormal && o.TimestampStatus != ingestion.TimestampStatusDegraded {
		return fmt.Errorf("unsupported spot trade-flow timestamp status %q", o.TimestampStatus)
	}
	if o.TimestampStatus == ingestion.TimestampStatusNormal && o.ExchangeTs.IsZero() {
		return fmt.Errorf("spot trade-flow exchange timestamp is required for normal timestamp status")
	}
	if o.FeedHealthState != ingestion.FeedHealthHealthy && o.FeedHealthState != ingestion.FeedHealthDegraded && o.FeedHealthState != ingestion.FeedHealthStale {
		return fmt.Errorf("unsupported spot trade-flow feed health state %q", o.FeedHealthState)
	}
	return nil
}

func assignSpotTradeFlowBucket(config SpotTradeFlowConfig, observation SpotTradeFlowObservation, family BucketFamily) (BucketAssignment, error) {
	familyConfig, ok := config.Families[family]
	if !ok {
		return BucketAssignment{}, fmt.Errorf("spot trade-flow family %q config missing", family)
	}
	eventTime := observation.ExchangeTs
	bucketSource := BucketSourceExchangeTs
	timestampDegraded := false
	if observation.TimestampStatus == ingestion.TimestampStatusDegraded {
		eventTime = observation.RecvTs
		bucketSource = BucketSourceRecvTs
		timestampDegraded = true
	}
	if eventTime.IsZero() {
		return BucketAssignment{}, fmt.Errorf("spot trade-flow event timestamp is required")
	}
	now := observation.ObservedAt
	if now.IsZero() {
		now = observation.RecvTs
	}
	interval := time.Duration(familyConfig.IntervalSeconds) * time.Second
	bucketStart := floorTime(eventTime.UTC(), interval)
	bucketEnd := bucketStart.Add(interval)
	disposition := LateEventOnTime
	if now.After(bucketEnd) {
		if now.After(bucketEnd.Add(time.Duration(familyConfig.WatermarkSeconds) * time.Second)) {
			disposition = LateEventAfterWatermark
		} else {
			disposition = LateEventWithinWatermark
		}
	}
	return BucketAssignment{
		Symbol:             observation.Symbol,
		Family:             family,
		BucketStart:        bucketStart.Format(time.RFC3339Nano),
		BucketEnd:          bucketEnd.Format(time.RFC3339Nano),
		BucketSource:       bucketSource,
		TimestampDegraded:  timestampDegraded,
		LateDisposition:    disposition,
		ExchangeRecvSkewMs: exchangeRecvSkewMillis(observation.ExchangeTs, observation.RecvTs),
	}, nil
}

func (s *spotTradeFlowBucketState) snapshot(config SpotTradeFlowConfig) SpotTradeFlowBucket {
	trades := make([]SpotTradeFlowObservation, 0, len(s.trades))
	for _, trade := range s.trades {
		trades = append(trades, trade)
	}
	sort.Slice(trades, func(i, j int) bool {
		return compareSpotTradeFlowObservations(trades[i], trades[j]) < 0
	})

	bucket := SpotTradeFlowBucket{
		SchemaVersion:    config.SchemaVersion,
		Symbol:           s.assignment.Symbol,
		Family:           s.assignment.Family,
		BucketStart:      s.assignment.BucketStart,
		BucketEnd:        s.assignment.BucketEnd,
		BucketSource:     s.assignment.BucketSource,
		LateDisposition:  s.assignment.LateDisposition,
		DuplicateCount:   s.duplicateCount,
		FeedHealthState:  ingestion.FeedHealthHealthy,
		ConfigVersion:    config.ConfigVersion,
		AlgorithmVersion: config.AlgorithmVersion,
	}
	if len(trades) == 0 {
		return bucket
	}

	reasons := map[ingestion.DegradationReason]struct{}{}
	totalSize := 0.0
	for _, trade := range trades {
		notional := trade.Price * trade.Size
		bucket.TradeCount++
		totalSize += trade.Size
		bucket.TotalNotional += notional
		if trade.TimestampStatus == ingestion.TimestampStatusDegraded {
			bucket.TimestampFallbackCount++
		}
		switch trade.Side {
		case SpotTradeFlowSideBuy:
			bucket.BuyTradeCount++
			bucket.BuyNotional += notional
		case SpotTradeFlowSideSell:
			bucket.SellTradeCount++
			bucket.SellNotional += notional
		}
		bucket.FeedHealthState = worstFeedHealthState(bucket.FeedHealthState, trade.FeedHealthState)
		for _, reason := range trade.FeedHealthReasons {
			reasons[reason] = struct{}{}
		}
	}
	bucket.NetAggressorNotional = bucket.BuyNotional - bucket.SellNotional
	bucket.FirstPrice = trades[0].Price
	bucket.LastPrice = trades[len(trades)-1].Price
	if totalSize > 0 {
		bucket.VWAP = bucket.TotalNotional / totalSize
	}
	if len(trades) > 1 && bucket.FirstPrice > 0 {
		bucket.PriceChangeBps = ((bucket.LastPrice - bucket.FirstPrice) / bucket.FirstPrice) * 10000
	}
	bucket.BuyNotional = roundMetric(bucket.BuyNotional)
	bucket.SellNotional = roundMetric(bucket.SellNotional)
	bucket.NetAggressorNotional = roundMetric(bucket.NetAggressorNotional)
	bucket.TotalNotional = roundMetric(bucket.TotalNotional)
	bucket.VWAP = roundMetric(bucket.VWAP)
	bucket.FirstPrice = roundMetric(bucket.FirstPrice)
	bucket.LastPrice = roundMetric(bucket.LastPrice)
	bucket.PriceChangeBps = roundMetric(bucket.PriceChangeBps)
	bucket.FeedHealthReasons = sortedDegradationReasons(reasons)
	return bucket
}

func compareSpotTradeFlowObservations(left, right SpotTradeFlowObservation) int {
	if cmp := compareTime(spotTradeFlowEventTime(left), spotTradeFlowEventTime(right)); cmp != 0 {
		return cmp
	}
	if cmp := compareTime(left.RecvTs, right.RecvTs); cmp != 0 {
		return cmp
	}
	if left.SourceRecordID < right.SourceRecordID {
		return -1
	}
	if left.SourceRecordID > right.SourceRecordID {
		return 1
	}
	return 0
}

func compareSpotTradeFlowBuckets(left, right SpotTradeFlowBucket) int {
	if cmp := compareSymbolOrder(left.Symbol, right.Symbol, spotTradeFlowTrackedSymbols); cmp != 0 {
		return cmp
	}
	if cmp := compareBucketFamilyOrder(left.Family, right.Family); cmp != 0 {
		return cmp
	}
	if left.BucketStart < right.BucketStart {
		return -1
	}
	if left.BucketStart > right.BucketStart {
		return 1
	}
	if left.BucketEnd < right.BucketEnd {
		return -1
	}
	if left.BucketEnd > right.BucketEnd {
		return 1
	}
	return 0
}

func spotTradeFlowEventTime(observation SpotTradeFlowObservation) time.Time {
	if observation.TimestampStatus == ingestion.TimestampStatusDegraded {
		return observation.RecvTs
	}
	return observation.ExchangeTs
}

func spotTradeFlowSourceKey(observation SpotTradeFlowObservation) string {
	return observation.Symbol + "\x00" + observation.SourceRecordID
}

func worstFeedHealthState(left, right ingestion.FeedHealthState) ingestion.FeedHealthState {
	if feedHealthRank(right) > feedHealthRank(left) {
		return right
	}
	return left
}

func feedHealthRank(state ingestion.FeedHealthState) int {
	switch state {
	case ingestion.FeedHealthStale:
		return 3
	case ingestion.FeedHealthDegraded:
		return 2
	case ingestion.FeedHealthHealthy:
		return 1
	default:
		return 0
	}
}

func sortedDegradationReasons(reasonSet map[ingestion.DegradationReason]struct{}) []ingestion.DegradationReason {
	reasons := make([]ingestion.DegradationReason, 0, len(reasonSet))
	for reason := range reasonSet {
		reasons = append(reasons, reason)
	}
	sort.Slice(reasons, func(i, j int) bool { return reasons[i] < reasons[j] })
	return reasons
}

func compareBucketFamilyOrder(left, right BucketFamily) int {
	leftIndex := bucketFamilyOrderIndex(left)
	rightIndex := bucketFamilyOrderIndex(right)
	if leftIndex < rightIndex {
		return -1
	}
	if leftIndex > rightIndex {
		return 1
	}
	return 0
}

func bucketFamilyOrderIndex(family BucketFamily) int {
	for index, candidate := range spotTradeFlowFamilyOrder {
		if candidate == family {
			return index
		}
	}
	return len(spotTradeFlowFamilyOrder)
}

func compareSymbolOrder(left, right string, order []string) int {
	leftIndex := symbolOrderIndex(left, order)
	rightIndex := symbolOrderIndex(right, order)
	if leftIndex < rightIndex {
		return -1
	}
	if leftIndex > rightIndex {
		return 1
	}
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func symbolOrderIndex(symbol string, order []string) int {
	for index, candidate := range order {
		if candidate == symbol {
			return index
		}
	}
	return len(order)
}

func compareTime(left, right time.Time) int {
	if left.Before(right) {
		return -1
	}
	if left.After(right) {
		return 1
	}
	return 0
}

func exchangeRecvSkewMillis(exchangeTs, recvTs time.Time) int64 {
	if exchangeTs.IsZero() || recvTs.IsZero() {
		return 0
	}
	skew := recvTs.Sub(exchangeTs)
	if skew < 0 {
		skew = -skew
	}
	return skew.Milliseconds()
}
