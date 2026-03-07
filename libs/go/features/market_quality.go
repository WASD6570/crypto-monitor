package features

func summarizeDivergence(buckets []*MarketQualityBucket) DivergenceSummary {
	if len(buckets) == 0 {
		return DivergenceSummary{Available: false, DirectionAgreement: DirectionAgreementUnknown}
	}
	latest := buckets[len(buckets)-1]
	available := latest.World.Available && latest.USA.Available && latest.worldPrice != nil && latest.usaPrice != nil
	if !available {
		return DivergenceSummary{
			Available:          false,
			DirectionAgreement: DirectionAgreementUnknown,
			ReasonCodes:        []ReasonCode{ReasonCompositeUnavailable},
		}
	}
	priceDistanceBps := roundMetric(priceDistance(*latest.worldPrice, *latest.usaPrice))
	participationGap := roundMetric(abs(latest.World.CoverageRatio - latest.USA.CoverageRatio))
	worldTrend := sideTrend(buckets, true)
	usaTrend := sideTrend(buckets, false)
	leaderChurn := leaderChurn(buckets, true) || leaderChurn(buckets, false)
	reasons := make([]ReasonCode, 0, 4)
	if priceDistanceBps > 0 {
		reasons = append(reasons, ReasonPriceDivergence)
	}
	switch directionAgreement(worldTrend, usaTrend) {
	case DirectionAgreementMixed, DirectionAgreementOpposed:
		reasons = appendReason(reasons, ReasonDirectionalDisagreement)
	}
	if participationGap > 0 {
		reasons = appendReason(reasons, ReasonCoverageAsymmetry)
	}
	if leaderChurn {
		reasons = appendReason(reasons, ReasonContributorChurn)
	}
	return DivergenceSummary{
		Available:           true,
		PriceDistanceBps:    priceDistanceBps,
		DirectionAgreement:  directionAgreement(worldTrend, usaTrend),
		ParticipationGap:    participationGap,
		LeaderChurnDetected: leaderChurn,
		ReasonCodes:         reasons,
	}
}

func summarizeTimestampTrust(config BucketConfig, buckets []MarketQualityBucket) TimestampTrustSummary {
	summary := TimestampTrustSummary{}
	configuredContributors := 0
	for _, bucket := range buckets {
		if bucket.Assignment.BucketSource == BucketSourceRecvTs {
			summary.RecvFallbackCount++
		} else if bucket.Assignment.BucketSource == BucketSourceExchangeTs {
			summary.ExchangeBucketCount++
		}
		summary.WorldFallbackCount += bucket.World.TimestampFallbackContributorCount
		summary.USAFallbackCount += bucket.USA.TimestampFallbackContributorCount
		configuredContributors += bucket.World.ConfiguredContributorCount + bucket.USA.ConfiguredContributorCount
	}
	fallbackContributors := summary.WorldFallbackCount + summary.USAFallbackCount
	if configuredContributors > 0 {
		summary.FallbackRatio = roundMetric(float64(fallbackContributors) / float64(configuredContributors))
	}
	summary.OneSidedFallback = (summary.WorldFallbackCount > 0) != (summary.USAFallbackCount > 0)
	thresholds := config.Thresholds.Divergence[buckets[len(buckets)-1].Window.Family]
	summary.TrustCap = summary.FallbackRatio >= thresholds.TimestampFallbackRatioCap || summary.OneSidedFallback || summary.RecvFallbackCount > 0
	return summary
}

func summarizeFragmentation(config BucketConfig, family BucketFamily, divergence DivergenceSummary, timestampTrust TimestampTrustSummary, persistence int) FragmentationSummary {
	thresholds := config.Thresholds.Divergence[family]
	reasons := append([]ReasonCode(nil), divergence.ReasonCodes...)
	unavailableSides := 0
	if !divergence.Available {
		unavailableSides = 1
		reasons = appendReason(reasons, ReasonCompositeUnavailable)
	}
	severity := FragmentationSeverityLow
	if !divergence.Available || divergence.PriceDistanceBps >= thresholds.PriceDistanceSevereBps || divergence.DirectionAgreement == DirectionAgreementOpposed || divergence.ParticipationGap >= thresholds.CoverageGapSevere || timestampTrust.TrustCap {
		severity = FragmentationSeveritySevere
	} else if divergence.PriceDistanceBps >= thresholds.PriceDistanceModerateBps || divergence.DirectionAgreement == DirectionAgreementMixed || divergence.ParticipationGap >= thresholds.CoverageGapModerate || divergence.LeaderChurnDetected {
		severity = FragmentationSeverityModerate
	}
	if timestampTrust.TrustCap {
		reasons = appendReason(reasons, ReasonTimestampTrustLoss)
	}
	if len(reasons) == 0 {
		reasons = []ReasonCode{ReasonNoContributors}
	}
	return FragmentationSummary{
		Severity:             severity,
		PersistenceCount:     persistence,
		PrimaryCauses:        reasons,
		UnavailableSideCount: unavailableSides,
	}
}

func summarizeMarketQuality(config BucketConfig, world CompositeBucketSide, usa CompositeBucketSide, fragmentation FragmentationSummary, timestampTrust TimestampTrustSummary, window BucketWindowSummary) MarketQualitySummary {
	thresholds := config.Thresholds.Quality
	worldCap := sideQualityCap(world, thresholds)
	usaCap := sideQualityCap(usa, thresholds)
	combined := minFloat(worldCap, usaCap)
	reasons := append([]ReasonCode(nil), fragmentation.PrimaryCauses...)
	if window.ExpectedBucketCount > 0 {
		completeness := float64(window.ClosedBucketCount-window.MissingBucketCount) / float64(window.ExpectedBucketCount)
		if completeness < config.Families[window.Family].MinimumCompleteness {
			combined = minFloat(combined, thresholds.IncompleteCap)
			reasons = appendReason(reasons, ReasonMissingInput)
		}
	}
	if timestampTrust.TrustCap {
		combined = minFloat(combined, thresholds.TimestampTrustCap)
		reasons = appendReason(reasons, ReasonTimestampTrustLoss)
	}
	switch fragmentation.Severity {
	case FragmentationSeveritySevere:
		combined = minFloat(combined, thresholds.SevereCap)
	case FragmentationSeverityModerate:
		combined = minFloat(combined, thresholds.ModerateCap)
	}
	return MarketQualitySummary{
		WorldQualityCap:   worldCap,
		USAQualityCap:     usaCap,
		CombinedTrustCap:  roundMetric(combined),
		DowngradedReasons: reasons,
		ReplayProvenance:  window.ConfigVersion,
	}
}

func fragmentationPersistence(buckets []MarketQualityBucket) int {
	count := 0
	for index := len(buckets) - 1; index >= 0; index-- {
		if buckets[index].Fragmentation.Severity == FragmentationSeverityLow {
			break
		}
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}

func sideQualityCap(side CompositeBucketSide, thresholds MarketQualityThresholds) float64 {
	if side.Unavailable || !side.Available {
		return 0
	}
	cap := minFloat(side.HealthScore, side.CoverageRatio)
	if side.MaxContributorWeight > thresholds.ConcentrationSoftCap {
		cap = minFloat(cap, roundMetric(1-(side.MaxContributorWeight-thresholds.ConcentrationSoftCap)))
	}
	if cap < 0 {
		return 0
	}
	return roundMetric(cap)
}

func sideTrend(buckets []*MarketQualityBucket, world bool) int {
	var first *float64
	var last *float64
	for _, bucket := range buckets {
		price := bucket.usaPrice
		if world {
			price = bucket.worldPrice
		}
		if price == nil {
			continue
		}
		if first == nil {
			first = price
		}
		last = price
	}
	if first == nil || last == nil {
		return 0
	}
	switch {
	case *last > *first:
		return 1
	case *last < *first:
		return -1
	default:
		return 0
	}
}

func directionAgreement(worldTrend, usaTrend int) DirectionAgreement {
	switch {
	case worldTrend == 0 && usaTrend == 0:
		return DirectionAgreementFlat
	case worldTrend == 0 || usaTrend == 0:
		return DirectionAgreementMixed
	case worldTrend == usaTrend:
		return DirectionAgreementAligned
	default:
		return DirectionAgreementOpposed
	}
}

func leaderChurn(buckets []*MarketQualityBucket, world bool) bool {
	previous := ""
	for _, bucket := range buckets {
		current := bucket.usaLeaderKey
		if world {
			current = bucket.worldLeaderKey
		}
		if current == "" {
			continue
		}
		if previous != "" && current != previous {
			return true
		}
		previous = current
	}
	return false
}

func priceDistance(left, right float64) float64 {
	average := (left + right) / 2
	if average == 0 {
		return 0
	}
	return abs(left-right) / average * 10000
}

func minFloat(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
