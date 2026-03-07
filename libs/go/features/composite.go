package features

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type CompositeGroup string

const (
	CompositeGroupWorld CompositeGroup = "WORLD"
	CompositeGroupUSA   CompositeGroup = "USA"
)

type ContributorStatus string

const (
	ContributorStatusIncluded  ContributorStatus = "included"
	ContributorStatusPenalized ContributorStatus = "penalized"
	ContributorStatusClamped   ContributorStatus = "clamped"
	ContributorStatusExcluded  ContributorStatus = "excluded"
)

type QuoteNormalizationMode string

const (
	QuoteNormalizationModeDirectUSD QuoteNormalizationMode = "direct-usd"
	QuoteNormalizationModeProxy     QuoteNormalizationMode = "config-proxy"
	QuoteNormalizationModeMixed     QuoteNormalizationMode = "mixed"
	QuoteNormalizationModeNone      QuoteNormalizationMode = "unavailable"
)

type ReasonCode string

const (
	ReasonMissingInput            ReasonCode = "missing-input"
	ReasonQuoteProxyDenied        ReasonCode = "quote-proxy-denied"
	ReasonQuoteConfidenceLoss     ReasonCode = "quote-confidence-loss"
	ReasonQuoteProxyPenalty       ReasonCode = "quote-proxy-penalized"
	ReasonTimestampFallback       ReasonCode = "timestamp-fallback"
	ReasonFeedHealthDegraded      ReasonCode = "feed-health-degraded"
	ReasonFeedHealthExcluded      ReasonCode = "feed-health-excluded"
	ReasonContributorClamped      ReasonCode = "contributor-clamped"
	ReasonMissingConfiguredPeer   ReasonCode = "missing-configured-peer"
	ReasonNoContributors          ReasonCode = "no-contributors"
	ReasonCompositeUnavailable    ReasonCode = "composite-unavailable"
	ReasonPriceDivergence         ReasonCode = "price-divergence"
	ReasonDirectionalDisagreement ReasonCode = "directional-disagreement"
	ReasonCoverageAsymmetry       ReasonCode = "coverage-asymmetry"
	ReasonTimestampTrustLoss      ReasonCode = "timestamp-trust-loss"
	ReasonContributorChurn        ReasonCode = "contributor-churn"
	ReasonUnusableInput           ReasonCode = "unusable-input"
)

type CompositeConfig struct {
	SchemaVersion    string                         `json:"schemaVersion"`
	ConfigVersion    string                         `json:"configVersion"`
	AlgorithmVersion string                         `json:"algorithmVersion"`
	Penalties        PenaltyConfig                  `json:"penalties"`
	QuoteProxies     map[string]QuoteProxyRule      `json:"quoteProxies"`
	Groups           map[CompositeGroup]GroupConfig `json:"groups"`
}

type PenaltyConfig struct {
	FeedHealthDegradedMultiplier float64 `json:"feedHealthDegradedMultiplier"`
	TimestampDegradedMultiplier  float64 `json:"timestampDegradedMultiplier"`
}

type QuoteProxyRule struct {
	Enabled           bool    `json:"enabled"`
	PenaltyMultiplier float64 `json:"penaltyMultiplier"`
}

type GroupConfig struct {
	Members []MemberConfig `json:"members"`
	Clamp   ClampConfig    `json:"clamp"`
}

type MemberConfig struct {
	Venue      ingestion.Venue `json:"venue"`
	MarketType string          `json:"marketType"`
	Symbols    []string        `json:"symbols"`
}

type ClampConfig struct {
	MinWeight float64 `json:"minWeight"`
	MaxWeight float64 `json:"maxWeight"`
}

type ContributorInput struct {
	Symbol                 string                             `json:"symbol"`
	Venue                  ingestion.Venue                    `json:"venue"`
	MarketType             string                             `json:"marketType"`
	QuoteCurrency          string                             `json:"quoteCurrency"`
	Price                  float64                            `json:"price"`
	LiquidityScore         float64                            `json:"liquidityScore"`
	TimestampStatus        ingestion.CanonicalTimestampStatus `json:"timestampStatus"`
	FeedHealthState        ingestion.FeedHealthState          `json:"feedHealthState"`
	FeedHealthReasons      []ingestion.DegradationReason      `json:"feedHealthReasons,omitempty"`
	QuoteConfidenceDropped bool                               `json:"quoteConfidenceDropped,omitempty"`
}

type ContributorPenalties struct {
	FeedHealthMultiplier         float64 `json:"feedHealthMultiplier"`
	TimestampDegradedMultiplier  float64 `json:"timestampDegradedMultiplier"`
	QuoteNormalizationMultiplier float64 `json:"quoteNormalizationMultiplier"`
	TotalMultiplier              float64 `json:"totalMultiplier"`
}

type SnapshotContributor struct {
	Venue                  ingestion.Venue                    `json:"venue"`
	MarketType             string                             `json:"marketType"`
	QuoteCurrency          string                             `json:"quoteCurrency"`
	Price                  float64                            `json:"price,omitempty"`
	RawWeight              float64                            `json:"rawWeight"`
	FinalWeight            float64                            `json:"finalWeight"`
	Status                 ContributorStatus                  `json:"status"`
	ReasonCodes            []ReasonCode                       `json:"reasonCodes,omitempty"`
	Penalties              ContributorPenalties               `json:"penalties"`
	QuoteNormalizationMode QuoteNormalizationMode             `json:"quoteNormalizationMode"`
	TimestampStatus        ingestion.CanonicalTimestampStatus `json:"timestampStatus"`
	FeedHealthState        ingestion.FeedHealthState          `json:"feedHealthState"`
}

type CompositeSnapshot struct {
	SchemaVersion                     string                 `json:"schemaVersion"`
	Symbol                            string                 `json:"symbol"`
	BucketTs                          string                 `json:"bucketTs"`
	CompositeGroup                    CompositeGroup         `json:"compositeGroup"`
	PriceBasis                        string                 `json:"priceBasis"`
	QuoteNormalizationMode            QuoteNormalizationMode `json:"quoteNormalizationMode"`
	CompositePrice                    *float64               `json:"compositePrice,omitempty"`
	Contributors                      []SnapshotContributor  `json:"contributors"`
	ConfiguredContributorCount        int                    `json:"configuredContributorCount"`
	EligibleContributorCount          int                    `json:"eligibleContributorCount"`
	ContributingContributorCount      int                    `json:"contributingContributorCount"`
	CoverageRatio                     float64                `json:"coverageRatio"`
	HealthScore                       float64                `json:"healthScore"`
	MaxContributorWeight              float64                `json:"maxContributorWeight"`
	TimestampFallbackContributorCount int                    `json:"timestampFallbackContributorCount"`
	Degraded                          bool                   `json:"degraded"`
	DegradedReasons                   []ReasonCode           `json:"degradedReasons,omitempty"`
	Unavailable                       bool                   `json:"unavailable"`
	ConfigVersion                     string                 `json:"configVersion"`
	AlgorithmVersion                  string                 `json:"algorithmVersion"`
}

func (c CompositeConfig) Validate() error {
	if c.SchemaVersion == "" {
		return fmt.Errorf("schemaVersion is required")
	}
	if c.ConfigVersion == "" {
		return fmt.Errorf("configVersion is required")
	}
	if c.AlgorithmVersion == "" {
		return fmt.Errorf("algorithmVersion is required")
	}
	if c.Penalties.FeedHealthDegradedMultiplier <= 0 || c.Penalties.FeedHealthDegradedMultiplier > 1 {
		return fmt.Errorf("feed health degraded multiplier must be within (0,1]")
	}
	if c.Penalties.TimestampDegradedMultiplier <= 0 || c.Penalties.TimestampDegradedMultiplier > 1 {
		return fmt.Errorf("timestamp degraded multiplier must be within (0,1]")
	}
	for quote, rule := range c.QuoteProxies {
		if rule.PenaltyMultiplier <= 0 || rule.PenaltyMultiplier > 1 {
			return fmt.Errorf("quote proxy %q penalty multiplier must be within (0,1]", quote)
		}
	}
	for _, group := range []CompositeGroup{CompositeGroupWorld, CompositeGroupUSA} {
		groupConfig, ok := c.Groups[group]
		if !ok {
			return fmt.Errorf("group %q is required", group)
		}
		if len(groupConfig.Members) == 0 {
			return fmt.Errorf("group %q requires at least one member", group)
		}
		if groupConfig.Clamp.MinWeight < 0 || groupConfig.Clamp.MinWeight >= 1 {
			return fmt.Errorf("group %q clamp minWeight must be within [0,1)", group)
		}
		if groupConfig.Clamp.MaxWeight <= 0 || groupConfig.Clamp.MaxWeight > 1 {
			return fmt.Errorf("group %q clamp maxWeight must be within (0,1]", group)
		}
		for _, member := range groupConfig.Members {
			if member.Venue == "" || member.MarketType == "" || len(member.Symbols) == 0 {
				return fmt.Errorf("group %q has incomplete member config", group)
			}
		}
	}
	return nil
}

func BuildCompositeSnapshot(config CompositeConfig, group CompositeGroup, symbol string, bucketTs time.Time, inputs []ContributorInput) (CompositeSnapshot, error) {
	if err := config.Validate(); err != nil {
		return CompositeSnapshot{}, err
	}
	if symbol == "" {
		return CompositeSnapshot{}, fmt.Errorf("symbol is required")
	}
	if bucketTs.IsZero() {
		return CompositeSnapshot{}, fmt.Errorf("bucket timestamp is required")
	}
	groupConfig, ok := config.Groups[group]
	if !ok {
		return CompositeSnapshot{}, fmt.Errorf("unknown composite group %q", group)
	}

	members := membersForSymbol(groupConfig.Members, symbol)
	if len(members) == 0 {
		return CompositeSnapshot{}, fmt.Errorf("group %q has no configured members for symbol %q", group, symbol)
	}

	orderedInputs := make(map[string]ContributorInput, len(inputs))
	for _, input := range inputs {
		if input.Symbol != symbol {
			continue
		}
		orderedInputs[memberKey(input.Venue, input.MarketType)] = input
	}

	contributors := make([]SnapshotContributor, 0, len(members))
	eligibleIndexes := make([]int, 0, len(members))
	adjustedWeights := make([]float64, 0, len(members))
	trustScores := make([]float64, 0, len(members))
	anyProxy := false
	anyDirect := false
	timestampFallbackCount := 0

	for _, member := range members {
		input, ok := orderedInputs[memberKey(member.Venue, member.MarketType)]
		contributor := SnapshotContributor{
			Venue:                  member.Venue,
			MarketType:             member.MarketType,
			Status:                 ContributorStatusExcluded,
			ReasonCodes:            nil,
			Penalties:              ContributorPenalties{FeedHealthMultiplier: 1, TimestampDegradedMultiplier: 1, QuoteNormalizationMultiplier: 1, TotalMultiplier: 1},
			QuoteNormalizationMode: QuoteNormalizationModeNone,
		}
		if !ok {
			contributor.ReasonCodes = []ReasonCode{ReasonMissingInput}
			contributors = append(contributors, contributor)
			continue
		}

		contributor.QuoteCurrency = input.QuoteCurrency
		contributor.Price = input.Price
		contributor.TimestampStatus = input.TimestampStatus
		contributor.FeedHealthState = input.FeedHealthState

		if input.Price <= 0 || input.LiquidityScore <= 0 {
			contributor.ReasonCodes = []ReasonCode{ReasonUnusableInput}
			contributors = append(contributors, contributor)
			continue
		}

		quoteMode, quotePenalty, quoteReason, excluded := evaluateQuoteNormalization(config, group, input)
		contributor.QuoteNormalizationMode = quoteMode
		contributor.Penalties.QuoteNormalizationMultiplier = quotePenalty
		if quoteMode == QuoteNormalizationModeDirectUSD {
			anyDirect = true
		}
		if quoteMode == QuoteNormalizationModeProxy {
			anyProxy = true
		}
		if excluded {
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, quoteReason)
			contributors = append(contributors, contributor)
			continue
		}
		if quoteReason != "" {
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, quoteReason)
		}

		if shouldExcludeForHealth(input) {
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, ReasonFeedHealthExcluded)
			contributors = append(contributors, contributor)
			continue
		}

		if input.FeedHealthState == ingestion.FeedHealthDegraded {
			contributor.Penalties.FeedHealthMultiplier = config.Penalties.FeedHealthDegradedMultiplier
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, ReasonFeedHealthDegraded)
		}
		if input.TimestampStatus == ingestion.TimestampStatusDegraded {
			contributor.Penalties.TimestampDegradedMultiplier = config.Penalties.TimestampDegradedMultiplier
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, ReasonTimestampFallback)
			timestampFallbackCount++
		}

		contributor.RawWeight = input.LiquidityScore
		contributor.Status = contributorStatusForReasons(contributor.ReasonCodes)
		contributor.Penalties.TotalMultiplier = contributor.Penalties.FeedHealthMultiplier * contributor.Penalties.TimestampDegradedMultiplier * contributor.Penalties.QuoteNormalizationMultiplier
		contributors = append(contributors, contributor)
		eligibleIndexes = append(eligibleIndexes, len(contributors)-1)
		adjustedWeights = append(adjustedWeights, contributor.RawWeight*contributor.Penalties.TotalMultiplier)
		trustScores = append(trustScores, contributor.Penalties.TotalMultiplier)
	}

	snapshot := CompositeSnapshot{
		SchemaVersion:                     config.SchemaVersion,
		Symbol:                            symbol,
		BucketTs:                          bucketTs.UTC().Format(time.RFC3339Nano),
		CompositeGroup:                    group,
		PriceBasis:                        "weighted-price",
		QuoteNormalizationMode:            summarizeQuoteMode(anyDirect, anyProxy),
		Contributors:                      contributors,
		ConfiguredContributorCount:        len(members),
		EligibleContributorCount:          len(eligibleIndexes),
		TimestampFallbackContributorCount: timestampFallbackCount,
		ConfigVersion:                     config.ConfigVersion,
		AlgorithmVersion:                  config.AlgorithmVersion,
	}

	if len(eligibleIndexes) == 0 {
		snapshot.CoverageRatio = 0
		snapshot.HealthScore = 0
		snapshot.Degraded = true
		snapshot.Unavailable = true
		snapshot.DegradedReasons = []ReasonCode{ReasonNoContributors}
		return snapshot, nil
	}

	preClamp := normalize(adjustedWeights)
	postClamp, clamped := clampNormalized(preClamp, groupConfig.Clamp.MinWeight, groupConfig.Clamp.MaxWeight)

	var compositePrice float64
	var healthScore float64
	var maxWeight float64
	contributingCount := 0
	reasonSet := map[ReasonCode]struct{}{}
	for contributorIndex, eligibleIndex := range eligibleIndexes {
		contributor := &snapshot.Contributors[eligibleIndex]
		contributor.FinalWeight = postClamp[contributorIndex]
		if clamped[contributorIndex] {
			contributor.Status = ContributorStatusClamped
			contributor.ReasonCodes = appendReason(contributor.ReasonCodes, ReasonContributorClamped)
		}
		if contributor.FinalWeight > 0 {
			contributingCount++
			compositePrice += contributor.Price * contributor.FinalWeight
			healthScore += trustScores[contributorIndex] * contributor.FinalWeight
			if contributor.FinalWeight > maxWeight {
				maxWeight = contributor.FinalWeight
			}
		}
		for _, reason := range contributor.ReasonCodes {
			reasonSet[reason] = struct{}{}
		}
	}

	snapshot.ContributingContributorCount = contributingCount
	snapshot.CoverageRatio = roundMetric(float64(contributingCount) / float64(snapshot.ConfiguredContributorCount))
	snapshot.HealthScore = roundMetric(healthScore)
	snapshot.MaxContributorWeight = roundMetric(maxWeight)
	if contributingCount == 0 {
		snapshot.Degraded = true
		snapshot.Unavailable = true
		snapshot.DegradedReasons = []ReasonCode{ReasonNoContributors}
		return snapshot, nil
	}
	price := roundMetric(compositePrice)
	snapshot.CompositePrice = &price

	if contributingCount < snapshot.ConfiguredContributorCount {
		reasonSet[ReasonMissingConfiguredPeer] = struct{}{}
	}
	reasons := sortedReasons(reasonSet)
	snapshot.DegradedReasons = reasons
	snapshot.Degraded = len(reasons) > 0 || snapshot.CoverageRatio < 1 || snapshot.HealthScore < 1
	snapshot.Unavailable = false
	return snapshot, nil
}

func membersForSymbol(members []MemberConfig, symbol string) []MemberConfig {
	filtered := make([]MemberConfig, 0, len(members))
	for _, member := range members {
		for _, configuredSymbol := range member.Symbols {
			if configuredSymbol == symbol {
				filtered = append(filtered, member)
				break
			}
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		left := memberKey(filtered[i].Venue, filtered[i].MarketType)
		right := memberKey(filtered[j].Venue, filtered[j].MarketType)
		return left < right
	})
	return filtered
}

func evaluateQuoteNormalization(config CompositeConfig, group CompositeGroup, input ContributorInput) (QuoteNormalizationMode, float64, ReasonCode, bool) {
	quote := strings.ToUpper(input.QuoteCurrency)
	if quote == "USD" {
		return QuoteNormalizationModeDirectUSD, 1, "", false
	}
	if group != CompositeGroupWorld {
		return QuoteNormalizationModeNone, 1, ReasonQuoteProxyDenied, true
	}
	rule, ok := config.QuoteProxies[quote]
	if !ok || !rule.Enabled {
		return QuoteNormalizationModeNone, 1, ReasonQuoteProxyDenied, true
	}
	if input.QuoteConfidenceDropped {
		return QuoteNormalizationModeNone, 1, ReasonQuoteConfidenceLoss, true
	}
	if rule.PenaltyMultiplier < 1 {
		return QuoteNormalizationModeProxy, rule.PenaltyMultiplier, ReasonQuoteProxyPenalty, false
	}
	return QuoteNormalizationModeProxy, 1, "", false
}

func shouldExcludeForHealth(input ContributorInput) bool {
	if input.FeedHealthState == ingestion.FeedHealthStale {
		return true
	}
	for _, reason := range input.FeedHealthReasons {
		if reason == ingestion.ReasonSequenceGap {
			return true
		}
	}
	return false
}

func contributorStatusForReasons(reasons []ReasonCode) ContributorStatus {
	if len(reasons) == 0 {
		return ContributorStatusIncluded
	}
	return ContributorStatusPenalized
}

func clampNormalized(weights []float64, minWeight float64, maxWeight float64) ([]float64, []bool) {
	bounded := append([]float64(nil), weights...)
	clamped := make([]bool, len(weights))
	if len(weights) <= 1 {
		return bounded, clamped
	}
	effectiveMin := minWeight
	if effectiveMin*float64(len(weights)) > 1 {
		effectiveMin = 0
	}
	effectiveMax := maxWeight
	if effectiveMax*float64(len(weights)) < 1 {
		effectiveMax = 1
	}
	original := append([]float64(nil), weights...)
	fixed := make([]bool, len(weights))

	for iteration := 0; iteration < len(weights)*2; iteration++ {
		remainingBudget := 1.0
		remainingOriginal := 0.0
		for index := range bounded {
			if fixed[index] {
				remainingBudget -= bounded[index]
				continue
			}
			remainingOriginal += original[index]
		}
		if remainingOriginal <= 0 {
			break
		}

		changed := false
		for index := range bounded {
			if fixed[index] {
				continue
			}
			candidate := remainingBudget * original[index] / remainingOriginal
			if effectiveMin > 0 && candidate < effectiveMin {
				bounded[index] = effectiveMin
				fixed[index] = true
				clamped[index] = true
				changed = true
				continue
			}
			if effectiveMax < 1 && candidate > effectiveMax {
				bounded[index] = effectiveMax
				fixed[index] = true
				clamped[index] = true
				changed = true
			}
		}
		if !changed {
			for index := range bounded {
				if fixed[index] {
					continue
				}
				bounded[index] = remainingBudget * original[index] / remainingOriginal
			}
			break
		}
	}

	return normalizeWithDrift(bounded), clamped
}

func normalize(weights []float64) []float64 {
	total := 0.0
	for _, weight := range weights {
		total += weight
	}
	if total <= 0 {
		return make([]float64, len(weights))
	}
	normalized := make([]float64, len(weights))
	for index, weight := range weights {
		normalized[index] = weight / total
	}
	return normalizeWithDrift(normalized)
}

func normalizeWithDrift(weights []float64) []float64 {
	normalized := append([]float64(nil), weights...)
	total := 0.0
	for _, weight := range normalized {
		total += weight
	}
	if total == 0 || len(normalized) == 0 {
		return normalized
	}
	drift := 1.0 - total
	normalized[len(normalized)-1] += drift
	for index := range normalized {
		normalized[index] = roundMetric(normalized[index])
	}
	total = 0
	for _, weight := range normalized {
		total += weight
	}
	normalized[len(normalized)-1] = roundMetric(normalized[len(normalized)-1] + (1 - total))
	return normalized
}

func summarizeQuoteMode(anyDirect, anyProxy bool) QuoteNormalizationMode {
	switch {
	case anyDirect && anyProxy:
		return QuoteNormalizationModeMixed
	case anyProxy:
		return QuoteNormalizationModeProxy
	case anyDirect:
		return QuoteNormalizationModeDirectUSD
	default:
		return QuoteNormalizationModeNone
	}
}

func memberKey(venue ingestion.Venue, marketType string) string {
	return string(venue) + ":" + marketType
}

func appendReason(reasons []ReasonCode, reason ReasonCode) []ReasonCode {
	if reason == "" {
		return reasons
	}
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}

func sortedReasons(reasonSet map[ReasonCode]struct{}) []ReasonCode {
	reasons := make([]ReasonCode, 0, len(reasonSet))
	for reason := range reasonSet {
		reasons = append(reasons, reason)
	}
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i] < reasons[j]
	})
	return reasons
}

func roundMetric(value float64) float64 {
	return float64(int(value*1_000_000+0.5)) / 1_000_000
}
