package features

import (
	"fmt"
	"sort"
	"time"
)

type RegimeState string

const (
	RegimeStateTradeable RegimeState = "TRADEABLE"
	RegimeStateWatch     RegimeState = "WATCH"
	RegimeStateNoOperate RegimeState = "NO-OPERATE"
)

type RegimeTransitionKind string

const (
	RegimeTransitionEnter     RegimeTransitionKind = "enter"
	RegimeTransitionHold      RegimeTransitionKind = "hold"
	RegimeTransitionDowngrade RegimeTransitionKind = "downgrade"
	RegimeTransitionRecover   RegimeTransitionKind = "recover"
)

type RegimeReasonCode string

const (
	RegimeReasonHealthy                    RegimeReasonCode = "healthy"
	RegimeReasonCompositeUnavailable       RegimeReasonCode = "composite-unavailable"
	RegimeReasonLateWindowIncomplete       RegimeReasonCode = "late-window-incomplete"
	RegimeReasonFragmentationSevere        RegimeReasonCode = "fragmentation-severe"
	RegimeReasonFragmentationModerate      RegimeReasonCode = "fragmentation-moderate"
	RegimeReasonCoverageLow                RegimeReasonCode = "coverage-low"
	RegimeReasonTimestampTrustLoss         RegimeReasonCode = "timestamp-trust-loss"
	RegimeReasonMarketQualityCap           RegimeReasonCode = "market-quality-cap"
	RegimeReasonUSDMInfluenceCap           RegimeReasonCode = "usdm-influence-cap"
	RegimeReasonRecoveryPersistencePending RegimeReasonCode = "recovery-persistence-pending"
	RegimeReasonGlobalSharedNoOperate      RegimeReasonCode = "global-shared-no-operate"
	RegimeReasonGlobalSharedWatch          RegimeReasonCode = "global-shared-watch"
)

type RegimeTriggerMetric struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value,omitempty"`
	Text    string  `json:"text,omitempty"`
	Integer int     `json:"integer,omitempty"`
}

type SymbolRegimeThresholds struct {
	CoverageWatchMax             float64 `json:"coverageWatchMax"`
	CoverageNoOperateMax         float64 `json:"coverageNoOperateMax"`
	CombinedTrustCapWatchMax     float64 `json:"combinedTrustCapWatchMax"`
	CombinedTrustCapNoOperateMax float64 `json:"combinedTrustCapNoOperateMax"`
	TimestampFallbackWatchRatio  float64 `json:"timestampFallbackWatchRatio"`
	TimestampFallbackNoOpRatio   float64 `json:"timestampFallbackNoOperateRatio"`
	NoOperateToWatchWindows      int     `json:"noOperateToWatchWindows"`
	WatchToTradeableWindows      int     `json:"watchToTradeableWindows"`
}

type GlobalRegimeThresholds struct {
	NoOperateToWatchWindows int `json:"noOperateToWatchWindows"`
	WatchToTradeableWindows int `json:"watchToTradeableWindows"`
}

type RegimeConfig struct {
	SchemaVersion    string                 `json:"schemaVersion"`
	ConfigVersion    string                 `json:"configVersion"`
	AlgorithmVersion string                 `json:"algorithmVersion"`
	Symbols          []string               `json:"symbols"`
	Symbol           SymbolRegimeThresholds `json:"symbol"`
	Global           GlobalRegimeThresholds `json:"global"`
}

type SymbolRegimeSnapshot struct {
	SchemaVersion         string                `json:"schemaVersion"`
	Symbol                string                `json:"symbol"`
	State                 RegimeState           `json:"state"`
	EffectiveBucketEnd    string                `json:"effectiveBucketEnd"`
	Reasons               []RegimeReasonCode    `json:"reasons"`
	PrimaryReason         RegimeReasonCode      `json:"primaryReason"`
	TriggerMetrics        []RegimeTriggerMetric `json:"triggerMetrics,omitempty"`
	PreviousState         RegimeState           `json:"previousState,omitempty"`
	TransitionKind        RegimeTransitionKind  `json:"transitionKind"`
	ConfigVersion         string                `json:"configVersion"`
	AlgorithmVersion      string                `json:"algorithmVersion"`
	RecoveryCandidate     RegimeState           `json:"recoveryCandidate,omitempty"`
	RecoveryWindowCount   int                   `json:"recoveryWindowCount,omitempty"`
	RecoveryWindowTarget  int                   `json:"recoveryWindowTarget,omitempty"`
	ObservedInstantaneous RegimeState           `json:"observedInstantaneousState"`
}

type GlobalRegimeSnapshot struct {
	SchemaVersion           string               `json:"schemaVersion"`
	State                   RegimeState          `json:"state"`
	EffectiveBucketEnd      string               `json:"effectiveBucketEnd"`
	Reasons                 []RegimeReasonCode   `json:"reasons"`
	PrimaryReason           RegimeReasonCode     `json:"primaryReason"`
	AffectedSymbols         []string             `json:"affectedSymbols,omitempty"`
	AppliedCeilingToSymbols []string             `json:"appliedCeilingToSymbols,omitempty"`
	PreviousState           RegimeState          `json:"previousState,omitempty"`
	TransitionKind          RegimeTransitionKind `json:"transitionKind"`
	ConfigVersion           string               `json:"configVersion"`
	AlgorithmVersion        string               `json:"algorithmVersion"`
	RecoveryCandidate       RegimeState          `json:"recoveryCandidate,omitempty"`
	RecoveryWindowCount     int                  `json:"recoveryWindowCount,omitempty"`
	RecoveryWindowTarget    int                  `json:"recoveryWindowTarget,omitempty"`
	ObservedInstantaneous   RegimeState          `json:"observedInstantaneousState"`
}

type RegimeEvaluation struct {
	Symbol         SymbolRegimeSnapshot   `json:"symbol"`
	Global         GlobalRegimeSnapshot   `json:"global"`
	EffectiveState map[string]RegimeState `json:"effectiveState"`
}

func (c RegimeConfig) Validate() error {
	if c.SchemaVersion == "" || c.ConfigVersion == "" || c.AlgorithmVersion == "" {
		return fmt.Errorf("regime config version fields are required")
	}
	if len(c.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}
	if c.Symbol.NoOperateToWatchWindows <= 0 || c.Symbol.WatchToTradeableWindows <= 0 {
		return fmt.Errorf("symbol recovery windows must be positive")
	}
	if c.Global.NoOperateToWatchWindows <= 0 || c.Global.WatchToTradeableWindows <= 0 {
		return fmt.Errorf("global recovery windows must be positive")
	}
	return nil
}

func EvaluateSymbolRegime(config RegimeConfig, bucket MarketQualityBucket, previous *SymbolRegimeSnapshot) (SymbolRegimeSnapshot, error) {
	if err := config.Validate(); err != nil {
		return SymbolRegimeSnapshot{}, err
	}
	if bucket.Window.Family != BucketFamily5m {
		return SymbolRegimeSnapshot{}, fmt.Errorf("symbol regime requires 5m bucket input")
	}
	instantaneous, reasons, metrics := classifySymbolInstantaneous(config.Symbol, bucket)
	nextState, transition, recoveryCandidate, recoveryCount, recoveryTarget, outputReasons := applyRecovery(previous, instantaneous, reasons, config.Symbol.NoOperateToWatchWindows, config.Symbol.WatchToTradeableWindows)
	return SymbolRegimeSnapshot{
		SchemaVersion:         config.SchemaVersion,
		Symbol:                bucket.Symbol,
		State:                 nextState,
		EffectiveBucketEnd:    bucket.Window.End,
		Reasons:               outputReasons,
		PrimaryReason:         outputReasons[0],
		TriggerMetrics:        metrics,
		PreviousState:         previousState(previous),
		TransitionKind:        transition,
		ConfigVersion:         config.ConfigVersion,
		AlgorithmVersion:      config.AlgorithmVersion,
		RecoveryCandidate:     recoveryCandidate,
		RecoveryWindowCount:   recoveryCount,
		RecoveryWindowTarget:  recoveryTarget,
		ObservedInstantaneous: instantaneous,
	}, nil
}

func EvaluateGlobalRegime(config RegimeConfig, symbols map[string]SymbolRegimeSnapshot, previous *GlobalRegimeSnapshot) (GlobalRegimeSnapshot, error) {
	if err := config.Validate(); err != nil {
		return GlobalRegimeSnapshot{}, err
	}
	instantaneous, reasons, affected, end := classifyGlobalInstantaneous(config.Symbols, symbols)
	nextState, transition, recoveryCandidate, recoveryCount, recoveryTarget, outputReasons := applyRecoveryGlobal(previous, instantaneous, reasons, config.Global.NoOperateToWatchWindows, config.Global.WatchToTradeableWindows)
	applied := appliedCeiling(config.Symbols, symbols, nextState)
	return GlobalRegimeSnapshot{
		SchemaVersion:           config.SchemaVersion,
		State:                   nextState,
		EffectiveBucketEnd:      end,
		Reasons:                 outputReasons,
		PrimaryReason:           outputReasons[0],
		AffectedSymbols:         affected,
		AppliedCeilingToSymbols: applied,
		PreviousState:           previousGlobalState(previous),
		TransitionKind:          transition,
		ConfigVersion:           config.ConfigVersion,
		AlgorithmVersion:        config.AlgorithmVersion,
		RecoveryCandidate:       recoveryCandidate,
		RecoveryWindowCount:     recoveryCount,
		RecoveryWindowTarget:    recoveryTarget,
		ObservedInstantaneous:   instantaneous,
	}, nil
}

func EffectiveStates(symbols []string, snapshots map[string]SymbolRegimeSnapshot, global GlobalRegimeSnapshot) map[string]RegimeState {
	result := make(map[string]RegimeState, len(symbols))
	for _, symbol := range symbols {
		snapshot, ok := snapshots[symbol]
		if !ok {
			continue
		}
		result[symbol] = minState(snapshot.State, global.State)
	}
	return result
}

func classifySymbolInstantaneous(config SymbolRegimeThresholds, bucket MarketQualityBucket) (RegimeState, []RegimeReasonCode, []RegimeTriggerMetric) {
	reasons := make([]RegimeReasonCode, 0, 6)
	metrics := make([]RegimeTriggerMetric, 0, 6)
	state := RegimeStateTradeable
	if bucket.World.Unavailable || bucket.USA.Unavailable || !bucket.Divergence.Available {
		state = minState(state, RegimeStateNoOperate)
		reasons = append(reasons, RegimeReasonCompositeUnavailable)
		metrics = append(metrics, RegimeTriggerMetric{Name: "unavailableSideCount", Integer: max(bucket.Fragmentation.UnavailableSideCount, 1)})
	}
	if bucket.Window.MissingBucketCount > 0 || bucket.Window.ClosedBucketCount < bucket.Window.ExpectedBucketCount {
		state = minState(state, RegimeStateNoOperate)
		reasons = append(reasons, RegimeReasonLateWindowIncomplete)
		metrics = append(metrics, RegimeTriggerMetric{Name: "missingBucketCount", Integer: bucket.Window.MissingBucketCount})
	}
	switch bucket.Fragmentation.Severity {
	case FragmentationSeveritySevere:
		state = minState(state, RegimeStateNoOperate)
		reasons = append(reasons, RegimeReasonFragmentationSevere)
		metrics = append(metrics, RegimeTriggerMetric{Name: "fragmentationPersistence", Integer: bucket.Fragmentation.PersistenceCount})
	case FragmentationSeverityModerate:
		state = minState(state, RegimeStateWatch)
		reasons = append(reasons, RegimeReasonFragmentationModerate)
		metrics = append(metrics, RegimeTriggerMetric{Name: "fragmentationPersistence", Integer: bucket.Fragmentation.PersistenceCount})
	}
	coverage := roundMetric(minFloat(bucket.World.CoverageRatio, bucket.USA.CoverageRatio))
	if coverage <= config.CoverageNoOperateMax {
		state = minState(state, RegimeStateNoOperate)
		reasons = append(reasons, RegimeReasonCoverageLow)
		metrics = append(metrics, RegimeTriggerMetric{Name: "coverageRatio", Value: coverage})
	} else if coverage <= config.CoverageWatchMax {
		state = minState(state, RegimeStateWatch)
		reasons = append(reasons, RegimeReasonCoverageLow)
		metrics = append(metrics, RegimeTriggerMetric{Name: "coverageRatio", Value: coverage})
	}
	if bucket.TimestampTrust.TrustCap || bucket.TimestampTrust.FallbackRatio >= config.TimestampFallbackWatchRatio || bucket.Assignment.BucketSource == BucketSourceRecvTs {
		reasons = append(reasons, RegimeReasonTimestampTrustLoss)
		metrics = append(metrics, RegimeTriggerMetric{Name: "timestampFallbackRatio", Value: roundMetric(bucket.TimestampTrust.FallbackRatio)})
		if bucket.TimestampTrust.FallbackRatio >= config.TimestampFallbackNoOpRatio || bucket.Assignment.BucketSource == BucketSourceRecvTs {
			state = minState(state, RegimeStateNoOperate)
		} else {
			state = minState(state, RegimeStateWatch)
		}
	}
	if bucket.MarketQuality.CombinedTrustCap <= config.CombinedTrustCapNoOperateMax {
		state = minState(state, RegimeStateNoOperate)
		reasons = append(reasons, RegimeReasonMarketQualityCap)
		metrics = append(metrics, RegimeTriggerMetric{Name: "combinedTrustCap", Value: bucket.MarketQuality.CombinedTrustCap})
	} else if bucket.MarketQuality.CombinedTrustCap <= config.CombinedTrustCapWatchMax {
		state = minState(state, RegimeStateWatch)
		reasons = append(reasons, RegimeReasonMarketQualityCap)
		metrics = append(metrics, RegimeTriggerMetric{Name: "combinedTrustCap", Value: bucket.MarketQuality.CombinedTrustCap})
	}
	reasons = stableReasonSet(reasons)
	metrics = stableMetrics(metrics)
	if len(reasons) == 0 {
		return RegimeStateTradeable, []RegimeReasonCode{RegimeReasonHealthy}, []RegimeTriggerMetric{{Name: "combinedTrustCap", Value: bucket.MarketQuality.CombinedTrustCap}}
	}
	return state, reasons, metrics
}

func classifyGlobalInstantaneous(expectedSymbols []string, symbols map[string]SymbolRegimeSnapshot) (RegimeState, []RegimeReasonCode, []string, string) {
	if len(symbols) < len(expectedSymbols) {
		return RegimeStateTradeable, []RegimeReasonCode{RegimeReasonHealthy}, nil, latestBucketEnd(symbols)
	}
	ordered := make([]string, 0, len(expectedSymbols))
	for _, symbol := range expectedSymbols {
		if _, ok := symbols[symbol]; ok {
			ordered = append(ordered, symbol)
		}
	}
	shared := sharedReasons(symbols, ordered)
	allWatch := true
	allNoOperate := true
	for _, symbol := range ordered {
		state := symbols[symbol].State
		if state == RegimeStateTradeable {
			allWatch = false
			allNoOperate = false
			break
		}
		if state != RegimeStateNoOperate {
			allNoOperate = false
		}
	}
	end := latestBucketEnd(symbols)
	if len(shared) > 0 && allNoOperate {
		return RegimeStateNoOperate, append([]RegimeReasonCode{RegimeReasonGlobalSharedNoOperate}, shared...), ordered, end
	}
	if len(shared) > 0 && allWatch {
		return RegimeStateWatch, append([]RegimeReasonCode{RegimeReasonGlobalSharedWatch}, shared...), ordered, end
	}
	return RegimeStateTradeable, []RegimeReasonCode{RegimeReasonHealthy}, nil, end
}

func applyRecovery(previous *SymbolRegimeSnapshot, instantaneous RegimeState, reasons []RegimeReasonCode, noOperateToWatch int, watchToTradeable int) (RegimeState, RegimeTransitionKind, RegimeState, int, int, []RegimeReasonCode) {
	if previous == nil || previous.State == "" {
		return instantaneous, RegimeTransitionEnter, "", 0, 0, reasons
	}
	if stateRank(instantaneous) > stateRank(previous.State) {
		return instantaneous, RegimeTransitionDowngrade, "", 0, 0, reasons
	}
	if instantaneous == previous.State {
		return instantaneous, RegimeTransitionHold, "", 0, 0, reasons
	}
	if previous.State == RegimeStateNoOperate {
		count := 1
		if previous.RecoveryCandidate == RegimeStateWatch {
			count = previous.RecoveryWindowCount + 1
		}
		if count >= noOperateToWatch {
			return RegimeStateWatch, RegimeTransitionRecover, "", 0, 0, reasons
		}
		return RegimeStateNoOperate, RegimeTransitionHold, RegimeStateWatch, count, noOperateToWatch, append([]RegimeReasonCode{RegimeReasonRecoveryPersistencePending}, reasons...)
	}
	count := 1
	if previous.RecoveryCandidate == RegimeStateTradeable {
		count = previous.RecoveryWindowCount + 1
	}
	if count >= watchToTradeable {
		return RegimeStateTradeable, RegimeTransitionRecover, "", 0, 0, reasons
	}
	return RegimeStateWatch, RegimeTransitionHold, RegimeStateTradeable, count, watchToTradeable, append([]RegimeReasonCode{RegimeReasonRecoveryPersistencePending}, reasons...)
}

func applyRecoveryGlobal(previous *GlobalRegimeSnapshot, instantaneous RegimeState, reasons []RegimeReasonCode, noOperateToWatch int, watchToTradeable int) (RegimeState, RegimeTransitionKind, RegimeState, int, int, []RegimeReasonCode) {
	if previous == nil || previous.State == "" {
		return instantaneous, RegimeTransitionEnter, "", 0, 0, reasons
	}
	if stateRank(instantaneous) > stateRank(previous.State) {
		return instantaneous, RegimeTransitionDowngrade, "", 0, 0, reasons
	}
	if instantaneous == previous.State {
		return instantaneous, RegimeTransitionHold, "", 0, 0, reasons
	}
	if previous.State == RegimeStateNoOperate {
		count := 1
		if previous.RecoveryCandidate == RegimeStateWatch {
			count = previous.RecoveryWindowCount + 1
		}
		if count >= noOperateToWatch {
			return RegimeStateWatch, RegimeTransitionRecover, "", 0, 0, reasons
		}
		return RegimeStateNoOperate, RegimeTransitionHold, RegimeStateWatch, count, noOperateToWatch, append([]RegimeReasonCode{RegimeReasonRecoveryPersistencePending}, reasons...)
	}
	count := 1
	if previous.RecoveryCandidate == RegimeStateTradeable {
		count = previous.RecoveryWindowCount + 1
	}
	if count >= watchToTradeable {
		return RegimeStateTradeable, RegimeTransitionRecover, "", 0, 0, reasons
	}
	return RegimeStateWatch, RegimeTransitionHold, RegimeStateTradeable, count, watchToTradeable, append([]RegimeReasonCode{RegimeReasonRecoveryPersistencePending}, reasons...)
}

func sharedReasons(symbols map[string]SymbolRegimeSnapshot, ordered []string) []RegimeReasonCode {
	shared := map[RegimeReasonCode]int{}
	for _, symbol := range ordered {
		seen := map[RegimeReasonCode]struct{}{}
		for _, reason := range symbols[symbol].Reasons {
			if !globalSharedReason(reason) {
				continue
			}
			if _, ok := seen[reason]; ok {
				continue
			}
			seen[reason] = struct{}{}
			shared[reason]++
		}
	}
	reasons := make([]RegimeReasonCode, 0, len(shared))
	for reason, count := range shared {
		if count == len(ordered) {
			reasons = append(reasons, reason)
		}
	}
	return stableReasonSet(reasons)
}

func globalSharedReason(reason RegimeReasonCode) bool {
	switch reason {
	case RegimeReasonCompositeUnavailable, RegimeReasonLateWindowIncomplete, RegimeReasonFragmentationSevere, RegimeReasonFragmentationModerate, RegimeReasonCoverageLow, RegimeReasonTimestampTrustLoss, RegimeReasonMarketQualityCap, RegimeReasonUSDMInfluenceCap:
		return true
	default:
		return false
	}
}

func appliedCeiling(expectedSymbols []string, symbols map[string]SymbolRegimeSnapshot, global RegimeState) []string {
	applied := make([]string, 0, len(expectedSymbols))
	if global == RegimeStateTradeable {
		return applied
	}
	for _, symbol := range expectedSymbols {
		if _, ok := symbols[symbol]; !ok {
			continue
		}
		applied = append(applied, symbol)
	}
	return applied
}

func latestBucketEnd(symbols map[string]SymbolRegimeSnapshot) string {
	latest := time.Time{}
	text := ""
	for _, snapshot := range symbols {
		ts, err := time.Parse(time.RFC3339Nano, snapshot.EffectiveBucketEnd)
		if err != nil {
			continue
		}
		if ts.After(latest) {
			latest = ts
			text = snapshot.EffectiveBucketEnd
		}
	}
	return text
}

func stableReasonSet(reasons []RegimeReasonCode) []RegimeReasonCode {
	if len(reasons) == 0 {
		return nil
	}
	seen := map[RegimeReasonCode]struct{}{}
	ordered := make([]RegimeReasonCode, 0, len(reasons))
	for _, reason := range reasons {
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		ordered = append(ordered, reason)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return reasonRank(ordered[i]) < reasonRank(ordered[j])
	})
	return ordered
}

func stableMetrics(metrics []RegimeTriggerMetric) []RegimeTriggerMetric {
	if len(metrics) == 0 {
		return nil
	}
	seen := map[string]RegimeTriggerMetric{}
	for _, metric := range metrics {
		if _, ok := seen[metric.Name]; !ok {
			seen[metric.Name] = metric
		}
	}
	ordered := make([]RegimeTriggerMetric, 0, len(seen))
	for _, metric := range seen {
		ordered = append(ordered, metric)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	return ordered
}

func stateRank(state RegimeState) int {
	switch state {
	case RegimeStateTradeable:
		return 1
	case RegimeStateWatch:
		return 2
	case RegimeStateNoOperate:
		return 3
	default:
		return 0
	}
}

func minState(left, right RegimeState) RegimeState {
	if stateRank(left) >= stateRank(right) {
		return left
	}
	return right
}

func reasonRank(reason RegimeReasonCode) int {
	switch reason {
	case RegimeReasonRecoveryPersistencePending:
		return 0
	case RegimeReasonGlobalSharedNoOperate:
		return 1
	case RegimeReasonGlobalSharedWatch:
		return 2
	case RegimeReasonCompositeUnavailable:
		return 3
	case RegimeReasonLateWindowIncomplete:
		return 4
	case RegimeReasonFragmentationSevere:
		return 5
	case RegimeReasonFragmentationModerate:
		return 6
	case RegimeReasonTimestampTrustLoss:
		return 7
	case RegimeReasonCoverageLow:
		return 8
	case RegimeReasonMarketQualityCap:
		return 9
	case RegimeReasonUSDMInfluenceCap:
		return 10
	case RegimeReasonHealthy:
		return 11
	default:
		return 99
	}
}

func previousState(previous *SymbolRegimeSnapshot) RegimeState {
	if previous == nil {
		return ""
	}
	return previous.State
}

func previousGlobalState(previous *GlobalRegimeSnapshot) RegimeState {
	if previous == nil {
		return ""
	}
	return previous.State
}
