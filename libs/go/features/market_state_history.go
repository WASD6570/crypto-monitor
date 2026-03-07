package features

import (
	"fmt"
	"sort"
)

const (
	MarketStateHistorySymbolSchema    = "market_state_history_symbol_v1"
	MarketStateHistoryGlobalSchema    = "market_state_history_global_v1"
	MarketStateAuditProvenanceSchema  = "market_state_audit_provenance_v1"
	MarketStateAuditLineageReplayKind = "market-state-audit-lineage.v1"
)

type MarketStateHistoryResolutionStatus string

const (
	MarketStateHistoryResolutionExact       MarketStateHistoryResolutionStatus = "exact"
	MarketStateHistoryResolutionUnavailable MarketStateHistoryResolutionStatus = "unavailable"
	MarketStateHistoryResolutionSuperseded  MarketStateHistoryResolutionStatus = "superseded"
)

type MarketStateHistoryAvailabilityCode string

const (
	MarketStateHistoryAvailabilityOK                MarketStateHistoryAvailabilityCode = "ok"
	MarketStateHistoryAvailabilityArtifactMissing   MarketStateHistoryAvailabilityCode = "artifact-unavailable"
	MarketStateHistoryAvailabilityPinMismatch       MarketStateHistoryAvailabilityCode = "pin-mismatch"
	MarketStateHistoryAvailabilityIncompleteLineage MarketStateHistoryAvailabilityCode = "incomplete-lineage"
)

type MarketStateAuditStatus string

const (
	MarketStateAuditStatusAuthoritativeOriginal MarketStateAuditStatus = "authoritative_original"
	MarketStateAuditStatusReplayCorrected       MarketStateAuditStatus = "replay_corrected"
	MarketStateAuditStatusUnavailable           MarketStateAuditStatus = "unavailable"
)

type MarketStateHistoryLookupQuery struct {
	Scope             string
	Symbol            string
	BucketFamily      BucketFamily
	BucketStart       string
	BucketEnd         string
	AsOf              string
	ConfigVersion     string
	AlgorithmVersion  string
	ReplayRunID       string
	ReplayManifestRef string
}

type MarketStateHistoryLookup struct {
	Scope             string                             `json:"scope"`
	Symbol            string                             `json:"symbol,omitempty"`
	BucketFamily      BucketFamily                       `json:"bucketFamily"`
	BucketStart       string                             `json:"bucketStart,omitempty"`
	BucketEnd         string                             `json:"bucketEnd,omitempty"`
	AsOf              string                             `json:"asOf,omitempty"`
	ConfigVersion     string                             `json:"configVersion,omitempty"`
	AlgorithmVersion  string                             `json:"algorithmVersion,omitempty"`
	ReplayRunID       string                             `json:"replayRunId,omitempty"`
	ReplayManifestRef string                             `json:"replayManifestRef,omitempty"`
	ResolutionStatus  MarketStateHistoryResolutionStatus `json:"resolutionStatus"`
}

type MarketStateHistoryAvailability struct {
	Code        MarketStateHistoryAvailabilityCode `json:"code"`
	ReasonCodes []string                           `json:"reasonCodes,omitempty"`
}

type MarketStateAuditCorrection struct {
	CorrectionCause                string
	ReasonCodes                    []string
	AuthoritativeReplayRunID       string
	AuthoritativeReplayManifestRef string
	AuthoritativeArtifactIDs       []string
	SupersededReplayRunID          string
	SupersededReplayManifestRef    string
	SupersededArtifactIDs          []string
}

type MarketStateAuditLineage struct {
	ReplayRunID       string                        `json:"replayRunId,omitempty"`
	ReplayManifestRef string                        `json:"replayManifestRef,omitempty"`
	ArtifactIDs       []string                      `json:"artifactIds,omitempty"`
	CompositeBucketTs []string                      `json:"compositeBucketTs,omitempty"`
	BucketRefs        []MarketStateCurrentBucketRef `json:"bucketRefs,omitempty"`
	SymbolBucketEnd   string                        `json:"symbolBucketEnd,omitempty"`
	GlobalBucketEnd   string                        `json:"globalBucketEnd,omitempty"`
}

type MarketStateAuditProvenanceResponse struct {
	SchemaVersion        string                   `json:"schemaVersion"`
	Lookup               MarketStateHistoryLookup `json:"lookup"`
	Status               MarketStateAuditStatus   `json:"status"`
	CorrectionCause      string                   `json:"correctionCause,omitempty"`
	ReasonCodes          []string                 `json:"reasonCodes,omitempty"`
	AuthoritativeLineage MarketStateAuditLineage  `json:"authoritativeLineage"`
	SupersededLineage    *MarketStateAuditLineage `json:"supersededLineage,omitempty"`
}

type MarketStateHistorySymbolResponse struct {
	SchemaVersion string                             `json:"schemaVersion"`
	Lookup        MarketStateHistoryLookup           `json:"lookup"`
	Availability  MarketStateHistoryAvailability     `json:"availability"`
	State         *MarketStateCurrentResponse        `json:"state,omitempty"`
	Audit         MarketStateAuditProvenanceResponse `json:"audit"`
}

type MarketStateHistoryGlobalResponse struct {
	SchemaVersion string                             `json:"schemaVersion"`
	Lookup        MarketStateHistoryLookup           `json:"lookup"`
	Availability  MarketStateHistoryAvailability     `json:"availability"`
	State         *MarketStateCurrentGlobalResponse  `json:"state,omitempty"`
	Audit         MarketStateAuditProvenanceResponse `json:"audit"`
}

type SymbolHistoricalStateQuery struct {
	Lookup     MarketStateHistoryLookupQuery
	Current    MarketStateCurrentResponse
	Correction *MarketStateAuditCorrection
}

type GlobalHistoricalStateQuery struct {
	Lookup     MarketStateHistoryLookupQuery
	Current    MarketStateCurrentGlobalResponse
	Correction *MarketStateAuditCorrection
}

type MarketStateAuditQuery struct {
	Lookup     MarketStateHistoryLookupQuery
	Provenance MarketStateCurrentProvenance
	Correction *MarketStateAuditCorrection
	Available  bool
}

func BuildMarketStateHistorySymbolResponse(query SymbolHistoricalStateQuery) (MarketStateHistorySymbolResponse, error) {
	lookup, err := normalizeHistoryLookup(query.Lookup, "symbol")
	if err != nil {
		return MarketStateHistorySymbolResponse{}, err
	}
	availability := validateSymbolHistoryLookup(query.Lookup, query.Current)
	lookup.ResolutionStatus = historyResolutionStatus(availability.Code, query.Correction)
	audit, err := BuildMarketStateAuditProvenance(MarketStateAuditQuery{
		Lookup:     query.Lookup,
		Provenance: query.Current.Provenance,
		Correction: query.Correction,
		Available:  availability.Code == MarketStateHistoryAvailabilityOK,
	})
	if err != nil {
		return MarketStateHistorySymbolResponse{}, err
	}
	audit.Lookup = lookup
	response := MarketStateHistorySymbolResponse{
		SchemaVersion: MarketStateHistorySymbolSchema,
		Lookup:        lookup,
		Availability:  availability,
		Audit:         audit,
	}
	if availability.Code == MarketStateHistoryAvailabilityOK {
		copy := query.Current
		response.State = &copy
	}
	return response, nil
}

func BuildMarketStateHistoryGlobalResponse(query GlobalHistoricalStateQuery) (MarketStateHistoryGlobalResponse, error) {
	lookup, err := normalizeHistoryLookup(query.Lookup, "global")
	if err != nil {
		return MarketStateHistoryGlobalResponse{}, err
	}
	availability := validateGlobalHistoryLookup(query.Lookup, query.Current)
	lookup.ResolutionStatus = historyResolutionStatus(availability.Code, query.Correction)
	audit, err := BuildMarketStateAuditProvenance(MarketStateAuditQuery{
		Lookup:     query.Lookup,
		Provenance: query.Current.Provenance,
		Correction: query.Correction,
		Available:  availability.Code == MarketStateHistoryAvailabilityOK,
	})
	if err != nil {
		return MarketStateHistoryGlobalResponse{}, err
	}
	audit.Lookup = lookup
	response := MarketStateHistoryGlobalResponse{
		SchemaVersion: MarketStateHistoryGlobalSchema,
		Lookup:        lookup,
		Availability:  availability,
		Audit:         audit,
	}
	if availability.Code == MarketStateHistoryAvailabilityOK {
		copy := query.Current
		response.State = &copy
	}
	return response, nil
}

func BuildMarketStateAuditProvenance(query MarketStateAuditQuery) (MarketStateAuditProvenanceResponse, error) {
	lookup, err := normalizeHistoryLookup(query.Lookup, query.Lookup.Scope)
	if err != nil {
		return MarketStateAuditProvenanceResponse{}, err
	}
	status := MarketStateAuditStatusUnavailable
	if query.Available {
		status = MarketStateAuditStatusAuthoritativeOriginal
		if query.Correction != nil && query.Correction.CorrectionCause != "" {
			status = MarketStateAuditStatusReplayCorrected
		}
	}
	response := MarketStateAuditProvenanceResponse{
		SchemaVersion: MarketStateAuditProvenanceSchema,
		Lookup:        lookup,
		Status:        status,
		AuthoritativeLineage: MarketStateAuditLineage{
			ReplayRunID:       firstNonEmpty(correctionValue(query.Correction, true, false), query.Lookup.ReplayRunID),
			ReplayManifestRef: firstNonEmpty(correctionManifest(query.Correction, true, false), query.Lookup.ReplayManifestRef),
			ArtifactIDs:       sortedStrings(correctionArtifacts(query.Correction, true, false)),
			CompositeBucketTs: append([]string(nil), query.Provenance.CompositeBucketTs...),
			BucketRefs:        append([]MarketStateCurrentBucketRef(nil), query.Provenance.BucketRefs...),
			SymbolBucketEnd:   query.Provenance.SymbolBucketEnd,
			GlobalBucketEnd:   query.Provenance.GlobalBucketEnd,
		},
	}
	if query.Correction != nil {
		response.CorrectionCause = query.Correction.CorrectionCause
		response.ReasonCodes = sortedStrings(query.Correction.ReasonCodes)
		if query.Correction.SupersededReplayRunID != "" || query.Correction.SupersededReplayManifestRef != "" || len(query.Correction.SupersededArtifactIDs) > 0 {
			response.SupersededLineage = &MarketStateAuditLineage{
				ReplayRunID:       query.Correction.SupersededReplayRunID,
				ReplayManifestRef: query.Correction.SupersededReplayManifestRef,
				ArtifactIDs:       sortedStrings(query.Correction.SupersededArtifactIDs),
				CompositeBucketTs: append([]string(nil), query.Provenance.CompositeBucketTs...),
				BucketRefs:        append([]MarketStateCurrentBucketRef(nil), query.Provenance.BucketRefs...),
				SymbolBucketEnd:   query.Provenance.SymbolBucketEnd,
				GlobalBucketEnd:   query.Provenance.GlobalBucketEnd,
			}
		}
	}
	if !query.Available {
		response.AuthoritativeLineage = MarketStateAuditLineage{}
		response.SupersededLineage = nil
	}
	return response, nil
}

func normalizeHistoryLookup(lookup MarketStateHistoryLookupQuery, fallbackScope string) (MarketStateHistoryLookup, error) {
	scope := firstNonEmpty(lookup.Scope, fallbackScope)
	if scope != "symbol" && scope != "global" {
		return MarketStateHistoryLookup{}, fmt.Errorf("history scope must be symbol or global")
	}
	if lookup.BucketFamily == "" {
		return MarketStateHistoryLookup{}, fmt.Errorf("bucket family is required")
	}
	if lookup.BucketEnd == "" && lookup.AsOf == "" {
		return MarketStateHistoryLookup{}, fmt.Errorf("bucket end or asOf is required")
	}
	if scope == "symbol" && lookup.Symbol == "" {
		return MarketStateHistoryLookup{}, fmt.Errorf("symbol history lookup requires symbol")
	}
	return MarketStateHistoryLookup{
		Scope:             scope,
		Symbol:            lookup.Symbol,
		BucketFamily:      lookup.BucketFamily,
		BucketStart:       lookup.BucketStart,
		BucketEnd:         lookup.BucketEnd,
		AsOf:              lookup.AsOf,
		ConfigVersion:     lookup.ConfigVersion,
		AlgorithmVersion:  lookup.AlgorithmVersion,
		ReplayRunID:       lookup.ReplayRunID,
		ReplayManifestRef: lookup.ReplayManifestRef,
	}, nil
}

func validateSymbolHistoryLookup(lookup MarketStateHistoryLookupQuery, current MarketStateCurrentResponse) MarketStateHistoryAvailability {
	if current.SchemaVersion == "" {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if current.Symbol != lookup.Symbol {
		return unavailableHistory(MarketStateHistoryAvailabilityPinMismatch, "pin-mismatch")
	}
	if current.Version.ConfigVersion != lookup.ConfigVersion || current.Version.AlgorithmVersion != lookup.AlgorithmVersion {
		return unavailableHistory(MarketStateHistoryAvailabilityPinMismatch, "pin-mismatch")
	}
	if lookup.AsOf != "" && current.AsOf != lookup.AsOf {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if !matchHistoryBucketRef(current.Provenance.BucketRefs, lookup) {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if current.Provenance.HistorySeam.ReservedSchemaFamily == "" {
		return unavailableHistory(MarketStateHistoryAvailabilityIncompleteLineage, "incomplete-lineage")
	}
	return MarketStateHistoryAvailability{Code: MarketStateHistoryAvailabilityOK}
}

func validateGlobalHistoryLookup(lookup MarketStateHistoryLookupQuery, current MarketStateCurrentGlobalResponse) MarketStateHistoryAvailability {
	if current.SchemaVersion == "" {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if current.Version.ConfigVersion != lookup.ConfigVersion || current.Version.AlgorithmVersion != lookup.AlgorithmVersion {
		return unavailableHistory(MarketStateHistoryAvailabilityPinMismatch, "pin-mismatch")
	}
	if lookup.AsOf != "" && current.AsOf != lookup.AsOf {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if !matchHistoryBucketRef(current.Provenance.BucketRefs, lookup) {
		return unavailableHistory(MarketStateHistoryAvailabilityArtifactMissing, "artifact-unavailable")
	}
	if current.Provenance.HistorySeam.ReservedSchemaFamily == "" {
		return unavailableHistory(MarketStateHistoryAvailabilityIncompleteLineage, "incomplete-lineage")
	}
	return MarketStateHistoryAvailability{Code: MarketStateHistoryAvailabilityOK}
}

func matchHistoryBucketRef(refs []MarketStateCurrentBucketRef, lookup MarketStateHistoryLookupQuery) bool {
	targetEnd := firstNonEmpty(lookup.BucketEnd, lookup.AsOf)
	for _, ref := range refs {
		if ref.Family != lookup.BucketFamily || ref.BucketEnd != targetEnd {
			continue
		}
		if lookup.BucketStart != "" && ref.BucketStart != lookup.BucketStart {
			continue
		}
		return true
	}
	return false
}

func unavailableHistory(code MarketStateHistoryAvailabilityCode, reason string) MarketStateHistoryAvailability {
	return MarketStateHistoryAvailability{Code: code, ReasonCodes: []string{reason}}
}

func historyResolutionStatus(code MarketStateHistoryAvailabilityCode, correction *MarketStateAuditCorrection) MarketStateHistoryResolutionStatus {
	if code != MarketStateHistoryAvailabilityOK {
		return MarketStateHistoryResolutionUnavailable
	}
	if correction != nil && correction.CorrectionCause != "" {
		return MarketStateHistoryResolutionSuperseded
	}
	return MarketStateHistoryResolutionExact
}

func correctionValue(correction *MarketStateAuditCorrection, authoritative bool, superseded bool) string {
	if correction == nil {
		return ""
	}
	if authoritative && !superseded {
		return correction.AuthoritativeReplayRunID
	}
	if superseded {
		return correction.SupersededReplayRunID
	}
	return ""
}

func correctionManifest(correction *MarketStateAuditCorrection, authoritative bool, superseded bool) string {
	if correction == nil {
		return ""
	}
	if authoritative && !superseded {
		return correction.AuthoritativeReplayManifestRef
	}
	if superseded {
		return correction.SupersededReplayManifestRef
	}
	return ""
}

func correctionArtifacts(correction *MarketStateAuditCorrection, authoritative bool, superseded bool) []string {
	if correction == nil {
		return nil
	}
	if authoritative && !superseded {
		return correction.AuthoritativeArtifactIDs
	}
	if superseded {
		return correction.SupersededArtifactIDs
	}
	return nil
}

func sortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	ordered := append([]string(nil), values...)
	sort.Strings(ordered)
	return ordered
}
