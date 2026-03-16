package features

import "fmt"

func ApplyUSDMInfluenceToSymbolRegime(symbol SymbolRegimeSnapshot, signal *USDMSymbolInfluenceSignal) (SymbolRegimeSnapshot, *MarketStateCurrentUSDMInfluenceProvenance, error) {
	if signal == nil {
		return symbol, nil, nil
	}
	if symbol.Symbol == "" {
		return SymbolRegimeSnapshot{}, nil, fmt.Errorf("symbol regime symbol is required")
	}
	if signal.Symbol == "" {
		return SymbolRegimeSnapshot{}, nil, fmt.Errorf("usdm influence signal symbol is required")
	}
	if symbol.Symbol != signal.Symbol {
		return SymbolRegimeSnapshot{}, nil, fmt.Errorf("usdm influence signal symbol %q does not match regime symbol %q", signal.Symbol, symbol.Symbol)
	}
	provenance := &MarketStateCurrentUSDMInfluenceProvenance{
		Evaluated:        true,
		Posture:          signal.Posture,
		PrimaryReason:    signal.PrimaryReason,
		ObservedAt:       signal.ObservedAt,
		ConfigVersion:    signal.ConfigVersion,
		AlgorithmVersion: signal.AlgorithmVersion,
	}
	if signal.Posture != USDMInfluencePostureDegradeCap || symbol.State != RegimeStateTradeable {
		return symbol, provenance, nil
	}
	adjusted := symbol
	adjusted.State = RegimeStateWatch
	adjusted.PreviousState = symbol.State
	adjusted.TransitionKind = RegimeTransitionDowngrade
	adjusted.Reasons = stableReasonSet(append([]RegimeReasonCode{RegimeReasonUSDMInfluenceCap}, symbol.Reasons...))
	adjusted.PrimaryReason = adjusted.Reasons[0]
	observed := symbol.ObservedInstantaneous
	if observed == "" {
		observed = symbol.State
	}
	adjusted.ObservedInstantaneous = minState(observed, RegimeStateWatch)
	adjusted.RecoveryCandidate = ""
	adjusted.RecoveryWindowCount = 0
	adjusted.RecoveryWindowTarget = 0
	provenance.AppliedCap = true
	return adjusted, provenance, nil
}
