package features

import (
	"fmt"
	"sort"
	"time"
)

const (
	MarketStateCurrentResponseSchema = "market_state_current_response_v1"
	MarketStateCurrentSymbolSchema   = "market_state_current_symbol_v1"
	MarketStateCurrentGlobalSchema   = "market_state_current_global_v1"
	MarketStateRecentContextSchema   = "market_state_recent_context_v1"

	currentStateRecentContextLimit = 1
)

type CurrentStateAvailability string

const (
	CurrentStateAvailabilityAvailable   CurrentStateAvailability = "available"
	CurrentStateAvailabilityDegraded    CurrentStateAvailability = "degraded"
	CurrentStateAvailabilityPartial     CurrentStateAvailability = "partial"
	CurrentStateAvailabilityUnavailable CurrentStateAvailability = "unavailable"
)

type MarketStateCurrentVersion struct {
	SchemaFamilyVersion string `json:"schemaFamilyVersion"`
	ConfigVersion       string `json:"configVersion"`
	AlgorithmVersion    string `json:"algorithmVersion"`
}

type MarketStateCurrentBucketRef struct {
	Family           BucketFamily `json:"family"`
	BucketStart      string       `json:"bucketStart"`
	BucketEnd        string       `json:"bucketEnd"`
	ConfigVersion    string       `json:"configVersion,omitempty"`
	AlgorithmVersion string       `json:"algorithmVersion,omitempty"`
}

type MarketStateHistoryAuditSeam struct {
	ReservedSchemaFamily string                        `json:"reservedSchemaFamily"`
	BucketRefs           []MarketStateCurrentBucketRef `json:"bucketRefs,omitempty"`
}

type MarketStateCurrentProvenance struct {
	CompositeBucketTs []string                      `json:"compositeBucketTs,omitempty"`
	BucketRefs        []MarketStateCurrentBucketRef `json:"bucketRefs,omitempty"`
	SymbolBucketEnd   string                        `json:"symbolBucketEnd,omitempty"`
	GlobalBucketEnd   string                        `json:"globalBucketEnd,omitempty"`
	USDMInfluence     *MarketStateCurrentUSDMInfluenceProvenance `json:"usdmInfluence,omitempty"`
	HistorySeam       MarketStateHistoryAuditSeam   `json:"historySeam"`
}

type MarketStateCurrentUSDMInfluenceProvenance struct {
	Evaluated        bool                    `json:"evaluated"`
	Posture          USDMInfluencePosture    `json:"posture,omitempty"`
	PrimaryReason    USDMInfluenceReasonCode `json:"primaryReason,omitempty"`
	AppliedCap       bool                    `json:"appliedCap,omitempty"`
	ObservedAt       string                  `json:"observedAt,omitempty"`
	ConfigVersion    string                  `json:"configVersion,omitempty"`
	AlgorithmVersion string                  `json:"algorithmVersion,omitempty"`
}

type MarketStateCurrentCompositeSection struct {
	Availability CurrentStateAvailability `json:"availability"`
	ReasonCodes  []ReasonCode             `json:"reasonCodes,omitempty"`
	World        CompositeSnapshot        `json:"world"`
	USA          CompositeSnapshot        `json:"usa"`
}

type MarketStateCurrentBucketSection struct {
	Availability CurrentStateAvailability `json:"availability"`
	ReasonCodes  []ReasonCode             `json:"reasonCodes,omitempty"`
	Bucket       MarketQualityBucket      `json:"bucket"`
}

type MarketStateCurrentBucketSummaries struct {
	ThirtySeconds MarketStateCurrentBucketSection `json:"thirtySeconds"`
	TwoMinutes    MarketStateCurrentBucketSection `json:"twoMinutes"`
	FiveMinutes   MarketStateCurrentBucketSection `json:"fiveMinutes"`
}

type MarketStateCurrentRegimeSection struct {
	Availability   CurrentStateAvailability `json:"availability"`
	ReasonCodes    []RegimeReasonCode       `json:"reasonCodes,omitempty"`
	EffectiveState RegimeState              `json:"effectiveState"`
	Symbol         SymbolRegimeSnapshot     `json:"symbol"`
	Global         GlobalRegimeSnapshot     `json:"global"`
}

type MarketStateRecentContextFamily struct {
	Family             BucketFamily             `json:"family"`
	Availability       CurrentStateAvailability `json:"availability"`
	Complete           bool                     `json:"complete"`
	Buckets            []MarketQualityBucket    `json:"buckets"`
	MissingBucketCount int                      `json:"missingBucketCount"`
}

type MarketStateRecentContext struct {
	SchemaVersion string                         `json:"schemaVersion"`
	ThirtySeconds MarketStateRecentContextFamily `json:"thirtySeconds"`
	TwoMinutes    MarketStateRecentContextFamily `json:"twoMinutes"`
	FiveMinutes   MarketStateRecentContextFamily `json:"fiveMinutes"`
}

type MarketStateCurrentResponse struct {
	SchemaVersion string                             `json:"schemaVersion"`
	Symbol        string                             `json:"symbol"`
	AsOf          string                             `json:"asOf"`
	Version       MarketStateCurrentVersion          `json:"version"`
	Composite     MarketStateCurrentCompositeSection `json:"composite"`
	Buckets       MarketStateCurrentBucketSummaries  `json:"buckets"`
	Regime        MarketStateCurrentRegimeSection    `json:"regime"`
	RecentContext MarketStateRecentContext           `json:"recentContext"`
	Provenance    MarketStateCurrentProvenance       `json:"provenance"`
}

type MarketStateCurrentSymbolSummary struct {
	Symbol           string                   `json:"symbol"`
	AsOf             string                   `json:"asOf"`
	EffectiveState   RegimeState              `json:"effectiveState"`
	SymbolState      RegimeState              `json:"symbolState"`
	GlobalState      RegimeState              `json:"globalState"`
	ReasonCodes      []RegimeReasonCode       `json:"reasonCodes,omitempty"`
	Availability     CurrentStateAvailability `json:"availability"`
	USDMInfluence    *MarketStateCurrentUSDMInfluenceProvenance `json:"usdmInfluence,omitempty"`
	ConfigVersion    string                   `json:"configVersion,omitempty"`
	AlgorithmVersion string                   `json:"algorithmVersion,omitempty"`
}

type MarketStateCurrentGlobalResponse struct {
	SchemaVersion string                            `json:"schemaVersion"`
	AsOf          string                            `json:"asOf"`
	Version       MarketStateCurrentVersion         `json:"version"`
	Global        GlobalRegimeSnapshot              `json:"global"`
	Symbols       []MarketStateCurrentSymbolSummary `json:"symbols"`
	Provenance    MarketStateCurrentProvenance      `json:"provenance"`
}

type SymbolCurrentStateQuery struct {
	Symbol        string
	AsOf          time.Time
	World         CompositeSnapshot
	USA           CompositeSnapshot
	Buckets       []MarketQualityBucket
	SymbolRegime  SymbolRegimeSnapshot
	GlobalRegime  GlobalRegimeSnapshot
	RecentContext []MarketQualityBucket
	USDMInfluence  *MarketStateCurrentUSDMInfluenceProvenance
}

type GlobalCurrentStateQuery struct {
	AsOf         time.Time
	GlobalRegime GlobalRegimeSnapshot
	Symbols      []MarketStateCurrentResponse
}

func BuildMarketStateCurrentResponse(query SymbolCurrentStateQuery) (MarketStateCurrentResponse, error) {
	if query.Symbol == "" {
		return MarketStateCurrentResponse{}, fmt.Errorf("symbol is required")
	}
	buckets := latestBucketsByFamily(query.Buckets)
	response := MarketStateCurrentResponse{
		SchemaVersion: MarketStateCurrentResponseSchema,
		Symbol:        query.Symbol,
		Version: MarketStateCurrentVersion{
			SchemaFamilyVersion: MarketStateCurrentResponseSchema,
			ConfigVersion:       firstNonEmpty(query.SymbolRegime.ConfigVersion, versionFromBucket(buckets[BucketFamily5m], true), versionFromComposite(query.World, true), versionFromComposite(query.USA, true)),
			AlgorithmVersion:    firstNonEmpty(query.SymbolRegime.AlgorithmVersion, versionFromBucket(buckets[BucketFamily5m], false), versionFromComposite(query.World, false), versionFromComposite(query.USA, false)),
		},
	}
	response.Composite = buildCompositeSection(query.World, query.USA)
	response.Buckets = MarketStateCurrentBucketSummaries{
		ThirtySeconds: buildBucketSection(buckets[BucketFamily30s]),
		TwoMinutes:    buildBucketSection(buckets[BucketFamily2m]),
		FiveMinutes:   buildBucketSection(buckets[BucketFamily5m]),
	}
	effectiveState := minState(query.SymbolRegime.State, query.GlobalRegime.State)
	response.Regime = MarketStateCurrentRegimeSection{
		Availability:   regimeAvailability(query.SymbolRegime, query.GlobalRegime),
		ReasonCodes:    stableRegimeReasonSet(query.SymbolRegime.Reasons, query.GlobalRegime.Reasons),
		EffectiveState: effectiveState,
		Symbol:         query.SymbolRegime,
		Global:         query.GlobalRegime,
	}
	response.RecentContext = buildRecentContext(query.RecentContext)
	response.AsOf = computeCurrentStateAsOf(query.AsOf, query.World.BucketTs, query.USA.BucketTs, query.SymbolRegime.EffectiveBucketEnd, query.GlobalRegime.EffectiveBucketEnd, bucketEnd(buckets[BucketFamily30s]), bucketEnd(buckets[BucketFamily2m]), bucketEnd(buckets[BucketFamily5m]))
	response.Provenance = buildProvenance(query.World, query.USA, buckets, query.SymbolRegime, query.GlobalRegime, query.RecentContext, query.USDMInfluence)
	return response, nil
}

func BuildMarketStateCurrentGlobalResponse(query GlobalCurrentStateQuery) (MarketStateCurrentGlobalResponse, error) {
	response := MarketStateCurrentGlobalResponse{
		SchemaVersion: MarketStateCurrentGlobalSchema,
		Version: MarketStateCurrentVersion{
			SchemaFamilyVersion: MarketStateCurrentGlobalSchema,
			ConfigVersion:       query.GlobalRegime.ConfigVersion,
			AlgorithmVersion:    query.GlobalRegime.AlgorithmVersion,
		},
		Global: query.GlobalRegime,
	}
	symbols := append([]MarketStateCurrentResponse(nil), query.Symbols...)
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Symbol < symbols[j].Symbol })
	response.Symbols = make([]MarketStateCurrentSymbolSummary, 0, len(symbols))
	for _, symbol := range symbols {
		response.Symbols = append(response.Symbols, MarketStateCurrentSymbolSummary{
			Symbol:           symbol.Symbol,
			AsOf:             symbol.AsOf,
			EffectiveState:   symbol.Regime.EffectiveState,
			SymbolState:      symbol.Regime.Symbol.State,
			GlobalState:      symbol.Regime.Global.State,
			ReasonCodes:      stableRegimeReasonSet(symbol.Regime.Symbol.Reasons, symbol.Regime.Global.Reasons),
			Availability:     symbol.Regime.Availability,
			USDMInfluence:    cloneUSDMInfluenceProvenance(symbol.Provenance.USDMInfluence),
			ConfigVersion:    symbol.Version.ConfigVersion,
			AlgorithmVersion: symbol.Version.AlgorithmVersion,
		})
	}
	response.AsOf = computeCurrentStateAsOf(query.AsOf, query.GlobalRegime.EffectiveBucketEnd)
	response.Provenance = MarketStateCurrentProvenance{
		SymbolBucketEnd: query.GlobalRegime.EffectiveBucketEnd,
		GlobalBucketEnd: query.GlobalRegime.EffectiveBucketEnd,
		HistorySeam: MarketStateHistoryAuditSeam{
			ReservedSchemaFamily: "market-state-history-and-audit-reads",
		},
	}
	for _, symbol := range symbols {
		response.Provenance.BucketRefs = append(response.Provenance.BucketRefs, symbol.Provenance.BucketRefs...)
	}
	response.Provenance.BucketRefs = dedupeBucketRefs(response.Provenance.BucketRefs)
	response.Provenance.HistorySeam.BucketRefs = response.Provenance.BucketRefs
	return response, nil
}

func buildCompositeSection(world CompositeSnapshot, usa CompositeSnapshot) MarketStateCurrentCompositeSection {
	reasonSet := map[ReasonCode]struct{}{}
	for _, reason := range world.DegradedReasons {
		reasonSet[reason] = struct{}{}
	}
	for _, reason := range usa.DegradedReasons {
		reasonSet[reason] = struct{}{}
	}
	availability := CurrentStateAvailabilityAvailable
	if world.Unavailable && usa.Unavailable {
		availability = CurrentStateAvailabilityUnavailable
	} else if world.Unavailable || usa.Unavailable {
		availability = CurrentStateAvailabilityPartial
	} else if world.Degraded || usa.Degraded {
		availability = CurrentStateAvailabilityDegraded
	}
	return MarketStateCurrentCompositeSection{
		Availability: availability,
		ReasonCodes:  sortedReasons(reasonSet),
		World:        world,
		USA:          usa,
	}
}

func buildBucketSection(bucket MarketQualityBucket) MarketStateCurrentBucketSection {
	availability := bucketAvailability(bucket)
	return MarketStateCurrentBucketSection{
		Availability: availability,
		ReasonCodes:  bucketReasonCodes(bucket),
		Bucket:       bucket,
	}
}

func buildRecentContext(buckets []MarketQualityBucket) MarketStateRecentContext {
	grouped := map[BucketFamily][]MarketQualityBucket{}
	for _, bucket := range buckets {
		grouped[bucket.Window.Family] = append(grouped[bucket.Window.Family], bucket)
	}
	return MarketStateRecentContext{
		SchemaVersion: MarketStateRecentContextSchema,
		ThirtySeconds: buildRecentContextFamily(BucketFamily30s, grouped[BucketFamily30s]),
		TwoMinutes:    buildRecentContextFamily(BucketFamily2m, grouped[BucketFamily2m]),
		FiveMinutes:   buildRecentContextFamily(BucketFamily5m, grouped[BucketFamily5m]),
	}
}

func buildRecentContextFamily(family BucketFamily, buckets []MarketQualityBucket) MarketStateRecentContextFamily {
	sorted := append([]MarketQualityBucket(nil), buckets...)
	sort.Slice(sorted, func(i, j int) bool {
		left, _ := time.Parse(time.RFC3339Nano, sorted[i].Window.End)
		right, _ := time.Parse(time.RFC3339Nano, sorted[j].Window.End)
		return left.Before(right)
	})
	if len(sorted) > currentStateRecentContextLimit {
		sorted = sorted[len(sorted)-currentStateRecentContextLimit:]
	}
	missing := 0
	complete := len(sorted) > 0
	availability := CurrentStateAvailabilityUnavailable
	if len(sorted) > 0 {
		availability = bucketAvailability(sorted[len(sorted)-1])
		for _, bucket := range sorted {
			missing += bucket.Window.MissingBucketCount
			if bucket.Window.MissingBucketCount > 0 || bucket.Window.ClosedBucketCount < bucket.Window.ExpectedBucketCount {
				complete = false
			}
		}
	}
	return MarketStateRecentContextFamily{
		Family:             family,
		Availability:       availability,
		Complete:           complete,
		Buckets:            sorted,
		MissingBucketCount: missing,
	}
}

func buildProvenance(world CompositeSnapshot, usa CompositeSnapshot, buckets map[BucketFamily]MarketQualityBucket, symbol SymbolRegimeSnapshot, global GlobalRegimeSnapshot, recent []MarketQualityBucket, usdm *MarketStateCurrentUSDMInfluenceProvenance) MarketStateCurrentProvenance {
	refs := []MarketStateCurrentBucketRef{}
	for _, family := range []BucketFamily{BucketFamily30s, BucketFamily2m, BucketFamily5m} {
		bucket := buckets[family]
		if bucket.Window.End == "" {
			continue
		}
		refs = append(refs, MarketStateCurrentBucketRef{Family: family, BucketStart: bucket.Window.Start, BucketEnd: bucket.Window.End, ConfigVersion: bucket.Window.ConfigVersion, AlgorithmVersion: bucket.Window.AlgorithmVersion})
	}
	for _, bucket := range recent {
		if bucket.Window.End == "" {
			continue
		}
		refs = append(refs, MarketStateCurrentBucketRef{Family: bucket.Window.Family, BucketStart: bucket.Window.Start, BucketEnd: bucket.Window.End, ConfigVersion: bucket.Window.ConfigVersion, AlgorithmVersion: bucket.Window.AlgorithmVersion})
	}
	refs = dedupeBucketRefs(refs)
	compositeBucketTs := make([]string, 0, 2)
	if world.BucketTs != "" {
		compositeBucketTs = append(compositeBucketTs, world.BucketTs)
	}
	if usa.BucketTs != "" && usa.BucketTs != world.BucketTs {
		compositeBucketTs = append(compositeBucketTs, usa.BucketTs)
	}
	sort.Strings(compositeBucketTs)
	provenance := MarketStateCurrentProvenance{
		CompositeBucketTs: compositeBucketTs,
		BucketRefs:        refs,
		SymbolBucketEnd:   symbol.EffectiveBucketEnd,
		GlobalBucketEnd:   global.EffectiveBucketEnd,
		USDMInfluence:     cloneUSDMInfluenceProvenance(usdm),
		HistorySeam: MarketStateHistoryAuditSeam{
			ReservedSchemaFamily: "market-state-history-and-audit-reads",
			BucketRefs:           refs,
		},
	}
	return provenance
}

func cloneUSDMInfluenceProvenance(value *MarketStateCurrentUSDMInfluenceProvenance) *MarketStateCurrentUSDMInfluenceProvenance {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func latestBucketsByFamily(buckets []MarketQualityBucket) map[BucketFamily]MarketQualityBucket {
	latest := map[BucketFamily]MarketQualityBucket{}
	for _, bucket := range buckets {
		existing, ok := latest[bucket.Window.Family]
		if !ok || bucketIsLater(bucket, existing) {
			latest[bucket.Window.Family] = bucket
		}
	}
	return latest
}

func bucketIsLater(left, right MarketQualityBucket) bool {
	leftTs, leftErr := time.Parse(time.RFC3339Nano, left.Window.End)
	rightTs, rightErr := time.Parse(time.RFC3339Nano, right.Window.End)
	if leftErr != nil {
		return false
	}
	if rightErr != nil {
		return true
	}
	return leftTs.After(rightTs)
}

func bucketAvailability(bucket MarketQualityBucket) CurrentStateAvailability {
	if bucket.Window.End == "" {
		return CurrentStateAvailabilityUnavailable
	}
	if bucket.Window.MissingBucketCount >= bucket.Window.ExpectedBucketCount && bucket.Window.ExpectedBucketCount > 0 {
		return CurrentStateAvailabilityUnavailable
	}
	if bucket.World.Unavailable && bucket.USA.Unavailable {
		return CurrentStateAvailabilityUnavailable
	}
	if bucket.Window.MissingBucketCount > 0 || bucket.World.Unavailable || bucket.USA.Unavailable {
		return CurrentStateAvailabilityPartial
	}
	if bucket.TimestampTrust.TrustCap || bucket.Fragmentation.Severity != FragmentationSeverityLow || bucket.MarketQuality.CombinedTrustCap < 1 {
		return CurrentStateAvailabilityDegraded
	}
	return CurrentStateAvailabilityAvailable
}

func bucketReasonCodes(bucket MarketQualityBucket) []ReasonCode {
	reasonSet := map[ReasonCode]struct{}{}
	for _, reason := range bucket.Divergence.ReasonCodes {
		reasonSet[reason] = struct{}{}
	}
	for _, reason := range bucket.Fragmentation.PrimaryCauses {
		reasonSet[reason] = struct{}{}
	}
	for _, reason := range bucket.MarketQuality.DowngradedReasons {
		reasonSet[reason] = struct{}{}
	}
	if bucket.Window.MissingBucketCount > 0 {
		reasonSet[ReasonMissingInput] = struct{}{}
	}
	return sortedReasons(reasonSet)
}

func regimeAvailability(symbol SymbolRegimeSnapshot, global GlobalRegimeSnapshot) CurrentStateAvailability {
	if symbol.State == "" && global.State == "" {
		return CurrentStateAvailabilityUnavailable
	}
	if symbol.State == RegimeStateTradeable && global.State == RegimeStateTradeable {
		return CurrentStateAvailabilityAvailable
	}
	return CurrentStateAvailabilityDegraded
}

func stableRegimeReasonSet(groups ...[]RegimeReasonCode) []RegimeReasonCode {
	combined := make([]RegimeReasonCode, 0)
	for _, group := range groups {
		combined = append(combined, group...)
	}
	return stableReasonSet(combined)
}

func versionFromBucket(bucket MarketQualityBucket, config bool) string {
	if config {
		return bucket.Window.ConfigVersion
	}
	return bucket.Window.AlgorithmVersion
}

func versionFromComposite(snapshot CompositeSnapshot, config bool) string {
	if config {
		return snapshot.ConfigVersion
	}
	return snapshot.AlgorithmVersion
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func bucketEnd(bucket MarketQualityBucket) string {
	return bucket.Window.End
}

func computeCurrentStateAsOf(explicit time.Time, timestamps ...string) string {
	if !explicit.IsZero() {
		return explicit.UTC().Format(time.RFC3339Nano)
	}
	latest := time.Time{}
	for _, ts := range timestamps {
		parsed, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			continue
		}
		if parsed.After(latest) {
			latest = parsed
		}
	}
	if latest.IsZero() {
		return ""
	}
	return latest.UTC().Format(time.RFC3339Nano)
}

func dedupeBucketRefs(refs []MarketStateCurrentBucketRef) []MarketStateCurrentBucketRef {
	seen := map[string]MarketStateCurrentBucketRef{}
	for _, ref := range refs {
		key := string(ref.Family) + ":" + ref.BucketStart + ":" + ref.BucketEnd
		seen[key] = ref
	}
	ordered := make([]MarketStateCurrentBucketRef, 0, len(seen))
	for _, ref := range seen {
		ordered = append(ordered, ref)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Family != ordered[j].Family {
			return ordered[i].Family < ordered[j].Family
		}
		if ordered[i].BucketStart != ordered[j].BucketStart {
			return ordered[i].BucketStart < ordered[j].BucketStart
		}
		return ordered[i].BucketEnd < ordered[j].BucketEnd
	})
	return ordered
}
