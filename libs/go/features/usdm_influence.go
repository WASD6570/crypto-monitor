package features

import (
	"fmt"
	"slices"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const (
	USDMInfluenceInputSchema  = "v1"
	USDMInfluenceSignalSchema = "v1"

	USDMInfluenceSurfaceWebsocket = "websocket"
	USDMInfluenceSurfaceRESTPoll  = "rest-poll"
)

var usdmInfluenceTrackedSymbols = []string{"BTC-USD", "ETH-USD"}

type USDMInfluencePosture string

const (
	USDMInfluencePostureAuxiliary       USDMInfluencePosture = "AUXILIARY"
	USDMInfluencePostureDegradeCap      USDMInfluencePosture = "DEGRADE_CAP"
	USDMInfluencePostureNoContext       USDMInfluencePosture = "NO_CONTEXT"
	USDMInfluencePostureDegradedContext USDMInfluencePosture = "DEGRADED_CONTEXT"
)

type USDMInfluenceReasonCode string

const (
	USDMInfluenceReasonHealthy              USDMInfluenceReasonCode = "healthy"
	USDMInfluenceReasonNoContext            USDMInfluenceReasonCode = "no-context"
	USDMInfluenceReasonFundingMissing       USDMInfluenceReasonCode = "funding-missing"
	USDMInfluenceReasonMarkIndexMissing     USDMInfluenceReasonCode = "mark-index-missing"
	USDMInfluenceReasonOpenInterestMissing  USDMInfluenceReasonCode = "open-interest-missing"
	USDMInfluenceReasonWebsocketStale       USDMInfluenceReasonCode = "websocket-stale"
	USDMInfluenceReasonWebsocketDegraded    USDMInfluenceReasonCode = "websocket-degraded"
	USDMInfluenceReasonOpenInterestStale    USDMInfluenceReasonCode = "open-interest-stale"
	USDMInfluenceReasonOpenInterestDegraded USDMInfluenceReasonCode = "open-interest-degraded"
	USDMInfluenceReasonTimestampDegraded    USDMInfluenceReasonCode = "timestamp-degraded"
	USDMInfluenceReasonBasisWide            USDMInfluenceReasonCode = "basis-wide"
	USDMInfluenceReasonFundingElevated      USDMInfluenceReasonCode = "funding-elevated"
	USDMInfluenceReasonRecentLiquidation    USDMInfluenceReasonCode = "recent-liquidation"
)

type USDMInfluenceFreshness string

const (
	USDMInfluenceFreshnessUnavailable USDMInfluenceFreshness = "UNAVAILABLE"
	USDMInfluenceFreshnessFresh       USDMInfluenceFreshness = "FRESH"
	USDMInfluenceFreshnessDegraded    USDMInfluenceFreshness = "DEGRADED"
	USDMInfluenceFreshnessStale       USDMInfluenceFreshness = "STALE"
)

type USDMInfluenceMetric struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value,omitempty"`
	Text    string  `json:"text,omitempty"`
	Integer int64   `json:"integer,omitempty"`
}

type USDMInfluenceInputMetadata struct {
	Surface           string                             `json:"surface"`
	Available         bool                               `json:"available"`
	Freshness         USDMInfluenceFreshness             `json:"freshness"`
	ExchangeTs        string                             `json:"exchangeTs,omitempty"`
	RecvTs            string                             `json:"recvTs,omitempty"`
	AgeMillis         int64                              `json:"ageMillis,omitempty"`
	TimestampStatus   ingestion.CanonicalTimestampStatus `json:"timestampStatus,omitempty"`
	FeedHealthState   ingestion.FeedHealthState          `json:"feedHealthState,omitempty"`
	FeedHealthReasons []ingestion.DegradationReason      `json:"feedHealthReasons,omitempty"`
	SourceRecordID    string                             `json:"sourceRecordId,omitempty"`
}

type USDMFundingInput struct {
	Metadata      USDMInfluenceInputMetadata `json:"metadata"`
	FundingRate   string                     `json:"fundingRate,omitempty"`
	NextFundingTs string                     `json:"nextFundingTs,omitempty"`
}

type USDMMarkIndexInput struct {
	Metadata   USDMInfluenceInputMetadata `json:"metadata"`
	MarkPrice  string                     `json:"markPrice,omitempty"`
	IndexPrice string                     `json:"indexPrice,omitempty"`
}

type USDMLiquidationInput struct {
	Metadata USDMInfluenceInputMetadata `json:"metadata"`
	Side     string                     `json:"side,omitempty"`
	Price    string                     `json:"price,omitempty"`
	Size     string                     `json:"size,omitempty"`
}

type USDMOpenInterestInput struct {
	Metadata     USDMInfluenceInputMetadata `json:"metadata"`
	OpenInterest string                     `json:"openInterest,omitempty"`
}

type USDMSymbolInfluenceInput struct {
	Symbol        string                `json:"symbol"`
	SourceSymbol  string                `json:"sourceSymbol"`
	QuoteCurrency string                `json:"quoteCurrency"`
	Funding       USDMFundingInput      `json:"funding"`
	MarkIndex     USDMMarkIndexInput    `json:"markIndex"`
	Liquidation   USDMLiquidationInput  `json:"liquidation"`
	OpenInterest  USDMOpenInterestInput `json:"openInterest"`
}

type USDMInfluenceEvaluatorInput struct {
	SchemaVersion string                     `json:"schemaVersion"`
	ObservedAt    string                     `json:"observedAt"`
	Symbols       []USDMSymbolInfluenceInput `json:"symbols"`
}

type USDMInfluenceConfig struct {
	SchemaVersion                    string   `json:"schemaVersion"`
	ConfigVersion                    string   `json:"configVersion"`
	AlgorithmVersion                 string   `json:"algorithmVersion"`
	Symbols                          []string `json:"symbols"`
	FundingAbsBpsDegradeCap          float64  `json:"fundingAbsBpsDegradeCap"`
	BasisAbsBpsDegradeCap            float64  `json:"basisAbsBpsDegradeCap"`
	LiquidationRecentWindowSeconds   int      `json:"liquidationRecentWindowSeconds"`
	LiquidationNotionalDegradeCapUSD float64  `json:"liquidationNotionalDegradeCapUsd"`
}

type USDMSymbolInfluenceSignal struct {
	SchemaVersion    string                    `json:"schemaVersion"`
	Symbol           string                    `json:"symbol"`
	SourceSymbol     string                    `json:"sourceSymbol"`
	QuoteCurrency    string                    `json:"quoteCurrency"`
	Posture          USDMInfluencePosture      `json:"posture"`
	Reasons          []USDMInfluenceReasonCode `json:"reasons"`
	PrimaryReason    USDMInfluenceReasonCode   `json:"primaryReason"`
	TriggerMetrics   []USDMInfluenceMetric     `json:"triggerMetrics,omitempty"`
	ConfigVersion    string                    `json:"configVersion"`
	AlgorithmVersion string                    `json:"algorithmVersion"`
	ObservedAt       string                    `json:"observedAt"`
}

type USDMInfluenceSignalSet struct {
	SchemaVersion    string                      `json:"schemaVersion"`
	ObservedAt       string                      `json:"observedAt"`
	Signals          []USDMSymbolInfluenceSignal `json:"signals"`
	ConfigVersion    string                      `json:"configVersion"`
	AlgorithmVersion string                      `json:"algorithmVersion"`
}

func DefaultUSDMInfluenceConfig() USDMInfluenceConfig {
	return USDMInfluenceConfig{
		SchemaVersion:                    "v1",
		ConfigVersion:                    "feature-engine.binance-usdm-influence.v1",
		AlgorithmVersion:                 "binance-usdm-policy.v1",
		Symbols:                          USDMInfluenceTrackedSymbols(),
		FundingAbsBpsDegradeCap:          5,
		BasisAbsBpsDegradeCap:            8,
		LiquidationRecentWindowSeconds:   300,
		LiquidationNotionalDegradeCapUSD: 100000,
	}
}

func USDMInfluenceTrackedSymbols() []string {
	return append([]string(nil), usdmInfluenceTrackedSymbols...)
}

func ValidateUSDMInfluenceSymbols(symbols []string) error {
	if !slices.Equal(symbols, usdmInfluenceTrackedSymbols) {
		return fmt.Errorf("usdm influence symbols must equal %v", usdmInfluenceTrackedSymbols)
	}
	return nil
}

func (i USDMInfluenceEvaluatorInput) Validate() error {
	if i.SchemaVersion != USDMInfluenceInputSchema {
		return fmt.Errorf("unsupported usdm influence input schema %q", i.SchemaVersion)
	}
	if _, err := time.Parse(time.RFC3339Nano, i.ObservedAt); err != nil {
		return fmt.Errorf("observedAt must be RFC3339Nano: %w", err)
	}
	if err := ValidateUSDMInfluenceSymbols(symbolsFromInputs(i.Symbols)); err != nil {
		return err
	}
	for _, symbol := range i.Symbols {
		if symbol.SourceSymbol == "" || symbol.QuoteCurrency == "" {
			return fmt.Errorf("usdm influence input identity is incomplete for %q", symbol.Symbol)
		}
		for _, metadata := range []USDMInfluenceInputMetadata{symbol.Funding.Metadata, symbol.MarkIndex.Metadata, symbol.Liquidation.Metadata, symbol.OpenInterest.Metadata} {
			if metadata.Surface == "" {
				return fmt.Errorf("input surface is required for %q", symbol.Symbol)
			}
			if metadata.Freshness == "" {
				return fmt.Errorf("input freshness is required for %q", symbol.Symbol)
			}
		}
	}
	return nil
}

func (c USDMInfluenceConfig) Validate() error {
	if c.SchemaVersion == "" || c.ConfigVersion == "" || c.AlgorithmVersion == "" {
		return fmt.Errorf("usdm influence config version fields are required")
	}
	if err := ValidateUSDMInfluenceSymbols(c.Symbols); err != nil {
		return err
	}
	if c.FundingAbsBpsDegradeCap <= 0 {
		return fmt.Errorf("funding abs bps degrade cap must be positive")
	}
	if c.BasisAbsBpsDegradeCap <= 0 {
		return fmt.Errorf("basis abs bps degrade cap must be positive")
	}
	if c.LiquidationRecentWindowSeconds <= 0 {
		return fmt.Errorf("liquidation recent window must be positive")
	}
	if c.LiquidationNotionalDegradeCapUSD <= 0 {
		return fmt.Errorf("liquidation notional degrade cap must be positive")
	}
	return nil
}

func (s USDMInfluenceSignalSet) Validate() error {
	if s.SchemaVersion != USDMInfluenceSignalSchema {
		return fmt.Errorf("unsupported usdm influence signal schema %q", s.SchemaVersion)
	}
	if s.ConfigVersion == "" || s.AlgorithmVersion == "" {
		return fmt.Errorf("usdm influence signal version fields are required")
	}
	if _, err := time.Parse(time.RFC3339Nano, s.ObservedAt); err != nil {
		return fmt.Errorf("observedAt must be RFC3339Nano: %w", err)
	}
	if err := ValidateUSDMInfluenceSymbols(symbolsFromSignals(s.Signals)); err != nil {
		return err
	}
	for _, signal := range s.Signals {
		if signal.SchemaVersion != USDMInfluenceSignalSchema {
			return fmt.Errorf("unsupported symbol signal schema %q", signal.SchemaVersion)
		}
		if signal.Posture == "" {
			return fmt.Errorf("signal posture is required for %q", signal.Symbol)
		}
		if signal.SourceSymbol == "" || signal.QuoteCurrency == "" {
			return fmt.Errorf("signal identity is incomplete for %q", signal.Symbol)
		}
		if len(signal.Reasons) == 0 {
			return fmt.Errorf("signal reasons are required for %q", signal.Symbol)
		}
		if signal.PrimaryReason != signal.Reasons[0] {
			return fmt.Errorf("signal primary reason must match first reason for %q", signal.Symbol)
		}
		if signal.ConfigVersion == "" || signal.AlgorithmVersion == "" {
			return fmt.Errorf("signal version fields are required for %q", signal.Symbol)
		}
	}
	return nil
}

func symbolsFromInputs(inputs []USDMSymbolInfluenceInput) []string {
	symbols := make([]string, 0, len(inputs))
	for _, input := range inputs {
		symbols = append(symbols, input.Symbol)
	}
	return symbols
}

func symbolsFromSignals(signals []USDMSymbolInfluenceSignal) []string {
	symbols := make([]string, 0, len(signals))
	for _, signal := range signals {
		symbols = append(symbols, signal.Symbol)
	}
	return symbols
}
