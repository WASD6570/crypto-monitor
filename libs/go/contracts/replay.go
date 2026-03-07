package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type ReplayRuntimeMode string

const (
	ReplayRuntimeModeInspect ReplayRuntimeMode = "inspect"
	ReplayRuntimeModeRebuild ReplayRuntimeMode = "rebuild"
	ReplayRuntimeModeCompare ReplayRuntimeMode = "compare"
)

type ReplayManifestStatus string

const (
	ReplayManifestStatusPlanned   ReplayManifestStatus = "planned"
	ReplayManifestStatusRunning   ReplayManifestStatus = "running"
	ReplayManifestStatusCompleted ReplayManifestStatus = "completed"
	ReplayManifestStatusFailed    ReplayManifestStatus = "failed"
)

type ReplayRunRequest struct {
	SchemaVersion    string                  `json:"schemaVersion"`
	RunID            string                  `json:"runId"`
	RequestID        string                  `json:"requestId,omitempty"`
	RequestDigest    string                  `json:"requestDigest,omitempty"`
	Scope            ReplayScope             `json:"scope"`
	ScopeKey         string                  `json:"scopeKey,omitempty"`
	ConflictKey      string                  `json:"conflictKey,omitempty"`
	RuntimeMode      ReplayRuntimeMode       `json:"runtimeMode"`
	RequestSource    string                  `json:"requestSource,omitempty"`
	OperatorNote     string                  `json:"operatorNote,omitempty"`
	ApplyIntent      bool                    `json:"applyIntent,omitempty"`
	ApprovalContext  *ReplayApprovalContext  `json:"approvalContext,omitempty"`
	ConfigSnapshot   ReplaySnapshotRef       `json:"configSnapshot"`
	ContractVersions []ReplaySnapshotRef     `json:"contractVersions"`
	Initiator        ReplayInitiatorMetadata `json:"initiator"`
}

type ReplayScope struct {
	Symbol         string   `json:"symbol"`
	Venues         []string `json:"venues"`
	StreamFamilies []string `json:"streamFamilies,omitempty"`
	WindowStart    string   `json:"windowStart"`
	WindowEnd      string   `json:"windowEnd"`
}

type ReplaySnapshotRef struct {
	SchemaVersion string `json:"schemaVersion"`
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	Version       string `json:"version,omitempty"`
	Digest        string `json:"digest"`
}

type ReplayPartitionRef struct {
	SchemaVersion         string `json:"schemaVersion"`
	LogicalPartition      string `json:"logicalPartition"`
	StorageState          string `json:"storageState,omitempty"`
	Location              string `json:"location"`
	EntryCount            int    `json:"entryCount"`
	FirstCanonicalEventID string `json:"firstCanonicalEventId"`
	LastCanonicalEventID  string `json:"lastCanonicalEventId"`
	ContinuityChecksum    string `json:"continuityChecksum"`
}

type ReplayBuildProvenance struct {
	Service           string `json:"service"`
	GitSHA            string `json:"gitSha,omitempty"`
	ReleaseID         string `json:"releaseId,omitempty"`
	BuildTimestamp    string `json:"buildTimestamp,omitempty"`
	FeatureFlagDigest string `json:"featureFlagDigest,omitempty"`
}

type ReplayApprovalContext struct {
	AuthorizationContext string `json:"authorizationContext"`
	ApprovalRef          string `json:"approvalRef"`
	PromotionToken       string `json:"promotionToken"`
}

type ReplayInitiatorMetadata struct {
	Actor       string `json:"actor"`
	ReasonCode  string `json:"reasonCode"`
	RequestedAt string `json:"requestedAt"`
}

type ReplayRunManifest struct {
	SchemaVersion    string                  `json:"schemaVersion"`
	RunID            string                  `json:"runId"`
	RequestID        string                  `json:"requestId"`
	RequestDigest    string                  `json:"requestDigest"`
	Scope            ReplayScope             `json:"scope"`
	ScopeKey         string                  `json:"scopeKey"`
	ConflictKey      string                  `json:"conflictKey"`
	RuntimeMode      ReplayRuntimeMode       `json:"runtimeMode"`
	RequestSource    string                  `json:"requestSource,omitempty"`
	OperatorNote     string                  `json:"operatorNote,omitempty"`
	ApplyIntent      bool                    `json:"applyIntent,omitempty"`
	ApprovalContext  *ReplayApprovalContext  `json:"approvalContext,omitempty"`
	Status           ReplayManifestStatus    `json:"status"`
	OrderingPolicyID string                  `json:"orderingPolicyId"`
	RawPartitions    []ReplayPartitionRef    `json:"rawPartitions"`
	ConfigSnapshot   ReplaySnapshotRef       `json:"configSnapshot"`
	ContractVersions []ReplaySnapshotRef     `json:"contractVersions"`
	Build            ReplayBuildProvenance   `json:"build"`
	Initiator        ReplayInitiatorMetadata `json:"initiator"`
}

type ReplayInputCounters struct {
	Partitions              int `json:"partitions"`
	Events                  int `json:"events"`
	Duplicates              int `json:"duplicates"`
	LateEvents              int `json:"lateEvents"`
	DegradedTimestampEvents int `json:"degradedTimestampEvents"`
}

type ReplayArtifactRef struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Digest    string `json:"digest"`
}

type ReplayRunResult struct {
	SchemaVersion        string              `json:"schemaVersion"`
	RunID                string              `json:"runId"`
	Status               string              `json:"status"`
	FailureCategory      string              `json:"failureCategory,omitempty"`
	InputCounters        ReplayInputCounters `json:"inputCounters"`
	ArtifactRefs         []ReplayArtifactRef `json:"artifactRefs,omitempty"`
	OutputDigest         string              `json:"outputDigest,omitempty"`
	ComparisonSummaryRef *ReplayArtifactRef  `json:"comparisonSummaryRef,omitempty"`
	StartedAt            string              `json:"startedAt"`
	FinishedAt           string              `json:"finishedAt"`
	ManifestDigest       string              `json:"manifestDigest"`
}

type ReplayCompareSummary struct {
	SchemaVersion        string            `json:"schemaVersion"`
	RunID                string            `json:"runId"`
	ComparedTargetID     string            `json:"comparedTargetId"`
	ChangedArtifactCount int               `json:"changedArtifactCount"`
	FirstMismatch        string            `json:"firstMismatch,omitempty"`
	UnchangedCount       int               `json:"unchangedCount"`
	DriftClassification  string            `json:"driftClassification"`
	ArtifactDigests      map[string]string `json:"artifactDigests"`
	AuditOnly            bool              `json:"auditOnly"`
}

type ReplayCheckpoint struct {
	SchemaVersion           string            `json:"schemaVersion"`
	RequestID               string            `json:"requestId"`
	RunID                   string            `json:"runId"`
	ScopeKey                string            `json:"scopeKey"`
	LogicalPartition        string            `json:"logicalPartition"`
	LastMaterializedEventID string            `json:"lastMaterializedEventId"`
	LastScannedEventID      string            `json:"lastScannedEventId,omitempty"`
	ConfigSnapshot          ReplaySnapshotRef `json:"configSnapshot"`
	ContractSetDigest       string            `json:"contractSetDigest"`
	BuildDigest             string            `json:"buildDigest"`
	OutputMode              ReplayRuntimeMode `json:"outputMode"`
	Sequence                int               `json:"sequence"`
	RetryCount              int               `json:"retryCount"`
	LastErrorClass          string            `json:"lastErrorClass,omitempty"`
	RecordedAt              string            `json:"recordedAt"`
}

type ReplayRequestAuditRecord struct {
	SchemaVersion string            `json:"schemaVersion"`
	RequestID     string            `json:"requestId"`
	RunID         string            `json:"runId"`
	Initiator     string            `json:"initiator"`
	RequestSource string            `json:"requestSource,omitempty"`
	ReasonCode    string            `json:"reasonCode"`
	OperatorNote  string            `json:"operatorNote,omitempty"`
	ScopeKey      string            `json:"scopeKey"`
	RuntimeMode   ReplayRuntimeMode `json:"runtimeMode"`
	ApplyIntent   bool              `json:"applyIntent"`
	RecordedAt    string            `json:"recordedAt"`
}

type ReplayExecutionAuditRecord struct {
	SchemaVersion      string              `json:"schemaVersion"`
	RequestID          string              `json:"requestId"`
	RunID              string              `json:"runId"`
	ResolvedPartitions []string            `json:"resolvedPartitions"`
	ConfigSnapshot     ReplaySnapshotRef   `json:"configSnapshot"`
	ContractSetDigest  string              `json:"contractSetDigest"`
	BuildDigest        string              `json:"buildDigest"`
	StartedAt          string              `json:"startedAt"`
	FinishedAt         string              `json:"finishedAt"`
	InputCounters      ReplayInputCounters `json:"inputCounters"`
	FailureCategory    string              `json:"failureCategory,omitempty"`
}

type ReplayCheckpointAuditRecord struct {
	SchemaVersion           string `json:"schemaVersion"`
	RequestID               string `json:"requestId"`
	RunID                   string `json:"runId"`
	CheckpointSequence      int    `json:"checkpointSequence"`
	LogicalPartition        string `json:"logicalPartition"`
	LastMaterializedEventID string `json:"lastMaterializedEventId"`
	RetryCount              int    `json:"retryCount"`
	FailureClass            string `json:"failureClass,omitempty"`
	RecordedAt              string `json:"recordedAt"`
}

type ReplayOutcomeAuditRecord struct {
	SchemaVersion        string              `json:"schemaVersion"`
	RequestID            string              `json:"requestId"`
	RunID                string              `json:"runId"`
	TerminalStatus       string              `json:"terminalStatus"`
	ArtifactRefs         []ReplayArtifactRef `json:"artifactRefs,omitempty"`
	PriorOutputReference string              `json:"priorOutputReference,omitempty"`
	ReplacementOutputRef string              `json:"replacementOutputRef,omitempty"`
	PromotionToken       string              `json:"promotionToken,omitempty"`
	CrossedApplyGate     bool                `json:"crossedApplyGate"`
	RecordedAt           string              `json:"recordedAt"`
}

func ValidateReplayRuntimeMode(mode ReplayRuntimeMode) error {
	switch mode {
	case ReplayRuntimeModeInspect, ReplayRuntimeModeRebuild, ReplayRuntimeModeCompare:
		return nil
	default:
		return fmt.Errorf("unsupported replay runtime mode %q", mode)
	}
}

func ValidateReplayApprovalContext(ctx *ReplayApprovalContext) error {
	if ctx == nil {
		return fmt.Errorf("replay approval context is required")
	}
	if ctx.AuthorizationContext == "" || ctx.ApprovalRef == "" || ctx.PromotionToken == "" {
		return fmt.Errorf("replay approval context is incomplete")
	}
	return nil
}

func ValidateReplaySnapshotRef(ref ReplaySnapshotRef) error {
	if ref.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay snapshot schema version %q", ref.SchemaVersion)
	}
	if ref.Kind == "" || ref.ID == "" || ref.Digest == "" {
		return fmt.Errorf("replay snapshot ref is incomplete")
	}
	return nil
}

func ValidateReplayRunRequest(request ReplayRunRequest) error {
	if request.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay run request schema version %q", request.SchemaVersion)
	}
	if request.RunID == "" {
		return fmt.Errorf("run id is required")
	}
	if err := ValidateReplayRuntimeMode(request.RuntimeMode); err != nil {
		return err
	}
	if err := validateReplayScope(request.Scope); err != nil {
		return err
	}
	if err := ValidateReplaySnapshotRef(request.ConfigSnapshot); err != nil {
		return err
	}
	for _, ref := range request.ContractVersions {
		if err := ValidateReplaySnapshotRef(ref); err != nil {
			return err
		}
	}
	if request.Initiator.Actor == "" || request.Initiator.ReasonCode == "" || request.Initiator.RequestedAt == "" {
		return fmt.Errorf("replay initiator metadata is incomplete")
	}
	return nil
}

func ValidateReplayRunManifest(manifest ReplayRunManifest) error {
	if manifest.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay run manifest schema version %q", manifest.SchemaVersion)
	}
	if manifest.RunID == "" || manifest.RequestID == "" || manifest.RequestDigest == "" || manifest.ScopeKey == "" || manifest.ConflictKey == "" || manifest.OrderingPolicyID == "" {
		return fmt.Errorf("replay run manifest metadata is incomplete")
	}
	if err := ValidateReplayRuntimeMode(manifest.RuntimeMode); err != nil {
		return err
	}
	if err := validateReplayScope(manifest.Scope); err != nil {
		return err
	}
	if len(manifest.RawPartitions) == 0 {
		return fmt.Errorf("replay run manifest must include at least one raw partition")
	}
	for _, ref := range manifest.RawPartitions {
		if ref.SchemaVersion != "v1" {
			return fmt.Errorf("unsupported replay partition schema version %q", ref.SchemaVersion)
		}
		if ref.LogicalPartition == "" || ref.Location == "" || ref.ContinuityChecksum == "" {
			return fmt.Errorf("replay partition ref is incomplete")
		}
		if ref.StorageState != "" && ref.StorageState != "hot" && ref.StorageState != "cold" && ref.StorageState != "transition" {
			return fmt.Errorf("unsupported replay partition storage state %q", ref.StorageState)
		}
	}
	if err := ValidateReplaySnapshotRef(manifest.ConfigSnapshot); err != nil {
		return err
	}
	for _, ref := range manifest.ContractVersions {
		if err := ValidateReplaySnapshotRef(ref); err != nil {
			return err
		}
	}
	if manifest.Build.Service == "" {
		return fmt.Errorf("replay build provenance service is required")
	}
	if manifest.Initiator.Actor == "" || manifest.Initiator.ReasonCode == "" || manifest.Initiator.RequestedAt == "" {
		return fmt.Errorf("replay initiator metadata is incomplete")
	}
	return nil
}

func ValidateReplayRunResult(result ReplayRunResult) error {
	if result.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay run result schema version %q", result.SchemaVersion)
	}
	if result.RunID == "" || result.Status == "" || result.StartedAt == "" || result.FinishedAt == "" || result.ManifestDigest == "" {
		return fmt.Errorf("replay run result metadata is incomplete")
	}
	return nil
}

func ValidateReplayCompareSummary(summary ReplayCompareSummary) error {
	if summary.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay compare summary schema version %q", summary.SchemaVersion)
	}
	if summary.RunID == "" || summary.ComparedTargetID == "" || summary.DriftClassification == "" {
		return fmt.Errorf("replay compare summary metadata is incomplete")
	}
	return nil
}

func ValidateReplayCheckpoint(checkpoint ReplayCheckpoint) error {
	if checkpoint.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay checkpoint schema version %q", checkpoint.SchemaVersion)
	}
	if checkpoint.RequestID == "" || checkpoint.RunID == "" || checkpoint.ScopeKey == "" || checkpoint.LogicalPartition == "" || checkpoint.LastMaterializedEventID == "" || checkpoint.ContractSetDigest == "" || checkpoint.BuildDigest == "" || checkpoint.Sequence < 1 || checkpoint.RecordedAt == "" {
		return fmt.Errorf("replay checkpoint metadata is incomplete")
	}
	if err := ValidateReplayRuntimeMode(checkpoint.OutputMode); err != nil {
		return err
	}
	if err := ValidateReplaySnapshotRef(checkpoint.ConfigSnapshot); err != nil {
		return err
	}
	return nil
}

func ValidateReplayRequestAuditRecord(record ReplayRequestAuditRecord) error {
	if record.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay request audit schema version %q", record.SchemaVersion)
	}
	if record.RequestID == "" || record.RunID == "" || record.Initiator == "" || record.ReasonCode == "" || record.ScopeKey == "" || record.RecordedAt == "" {
		return fmt.Errorf("replay request audit metadata is incomplete")
	}
	return ValidateReplayRuntimeMode(record.RuntimeMode)
}

func ValidateReplayExecutionAuditRecord(record ReplayExecutionAuditRecord) error {
	if record.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay execution audit schema version %q", record.SchemaVersion)
	}
	if record.RequestID == "" || record.RunID == "" || len(record.ResolvedPartitions) == 0 || record.ContractSetDigest == "" || record.BuildDigest == "" || record.StartedAt == "" || record.FinishedAt == "" {
		return fmt.Errorf("replay execution audit metadata is incomplete")
	}
	return ValidateReplaySnapshotRef(record.ConfigSnapshot)
}

func ValidateReplayCheckpointAuditRecord(record ReplayCheckpointAuditRecord) error {
	if record.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay checkpoint audit schema version %q", record.SchemaVersion)
	}
	if record.RequestID == "" || record.RunID == "" || record.CheckpointSequence < 1 || record.LogicalPartition == "" || record.LastMaterializedEventID == "" || record.RecordedAt == "" {
		return fmt.Errorf("replay checkpoint audit metadata is incomplete")
	}
	return nil
}

func ValidateReplayOutcomeAuditRecord(record ReplayOutcomeAuditRecord) error {
	if record.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported replay outcome audit schema version %q", record.SchemaVersion)
	}
	if record.RequestID == "" || record.RunID == "" || record.TerminalStatus == "" || record.RecordedAt == "" {
		return fmt.Errorf("replay outcome audit metadata is incomplete")
	}
	switch record.TerminalStatus {
	case "no-op", "compare-only", "isolated-rebuild", "rejected-apply", "promoted-correction":
		return nil
	default:
		return fmt.Errorf("unsupported replay outcome terminal status %q", record.TerminalStatus)
	}
}

func ReplayRequestDigest(request ReplayRunRequest) (string, error) {
	return canonicalJSONDigest(request)
}

func ReplayManifestDigest(manifest ReplayRunManifest) (string, error) {
	return canonicalJSONDigest(manifest)
}

func ReplayValueDigest(value any) (string, error) {
	return canonicalJSONDigest(value)
}

func ReplayContractSetDigest(refs []ReplaySnapshotRef) (string, error) {
	normalized := append([]ReplaySnapshotRef(nil), refs...)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Kind != normalized[j].Kind {
			return normalized[i].Kind < normalized[j].Kind
		}
		if normalized[i].ID != normalized[j].ID {
			return normalized[i].ID < normalized[j].ID
		}
		if normalized[i].Version != normalized[j].Version {
			return normalized[i].Version < normalized[j].Version
		}
		return normalized[i].Digest < normalized[j].Digest
	})
	return ReplayValueDigest(normalized)
}

func ReplayBuildDigest(build ReplayBuildProvenance) (string, error) {
	return ReplayValueDigest(build)
}

func NormalizeReplayScope(scope ReplayScope) (ReplayScope, string, error) {
	if err := validateReplayScope(scope); err != nil {
		return ReplayScope{}, "", err
	}
	normalized := ReplayScope{
		Symbol:      strings.ToUpper(strings.TrimSpace(scope.Symbol)),
		WindowStart: scope.WindowStart,
		WindowEnd:   scope.WindowEnd,
	}
	venues := make([]string, 0, len(scope.Venues))
	seenVenues := make(map[string]struct{}, len(scope.Venues))
	for _, venue := range scope.Venues {
		value := strings.ToLower(strings.TrimSpace(venue))
		if value == "" || value == "*" {
			return ReplayScope{}, "", fmt.Errorf("replay scope venues must be explicit")
		}
		if _, ok := seenVenues[value]; ok {
			continue
		}
		seenVenues[value] = struct{}{}
		venues = append(venues, value)
	}
	if len(venues) == 0 {
		return ReplayScope{}, "", fmt.Errorf("replay scope venues are required")
	}
	sort.Strings(venues)
	normalized.Venues = venues
	streams := make([]string, 0, len(scope.StreamFamilies))
	seenStreams := make(map[string]struct{}, len(scope.StreamFamilies))
	for _, family := range scope.StreamFamilies {
		value := strings.ToLower(strings.TrimSpace(family))
		if value == "" || value == "*" {
			return ReplayScope{}, "", fmt.Errorf("replay stream families must be explicit")
		}
		if _, ok := seenStreams[value]; ok {
			continue
		}
		seenStreams[value] = struct{}{}
		streams = append(streams, value)
	}
	if len(streams) > 0 {
		sort.Strings(streams)
		normalized.StreamFamilies = streams
	}
	start, err := time.Parse(time.RFC3339Nano, normalized.WindowStart)
	if err != nil {
		return ReplayScope{}, "", fmt.Errorf("parse replay window start: %w", err)
	}
	end, err := time.Parse(time.RFC3339Nano, normalized.WindowEnd)
	if err != nil {
		return ReplayScope{}, "", fmt.Errorf("parse replay window end: %w", err)
	}
	if !end.After(start) {
		return ReplayScope{}, "", fmt.Errorf("replay scope window end must be after start")
	}
	parts := []string{
		"symbol=" + normalized.Symbol,
		"venues=" + strings.Join(normalized.Venues, ","),
		"window=" + normalized.WindowStart + "/" + normalized.WindowEnd,
	}
	if len(normalized.StreamFamilies) > 0 {
		parts = append(parts, "streams="+strings.Join(normalized.StreamFamilies, ","))
	}
	sort.Strings(parts)
	return normalized, strings.Join(parts, "|"), nil
}

func ReplayRequestIdentity(request ReplayRunRequest, build ReplayBuildProvenance) (string, string, string, error) {
	normalized, scopeKey, err := NormalizeReplayScope(request.Scope)
	if err != nil {
		return "", "", "", err
	}
	contractSetDigest, err := ReplayContractSetDigest(request.ContractVersions)
	if err != nil {
		return "", "", "", err
	}
	buildDigest, err := ReplayBuildDigest(build)
	if err != nil {
		return "", "", "", err
	}
	requestID, err := ReplayValueDigest(struct {
		ScopeKey          string            `json:"scopeKey"`
		RuntimeMode       ReplayRuntimeMode `json:"runtimeMode"`
		ConfigDigest      string            `json:"configDigest"`
		ContractSetDigest string            `json:"contractSetDigest"`
		BuildDigest       string            `json:"buildDigest"`
		Initiator         string            `json:"initiator"`
		ApplyIntent       bool              `json:"applyIntent"`
	}{
		ScopeKey:          scopeKey,
		RuntimeMode:       request.RuntimeMode,
		ConfigDigest:      request.ConfigSnapshot.Digest,
		ContractSetDigest: contractSetDigest,
		BuildDigest:       buildDigest,
		Initiator:         request.Initiator.Actor,
		ApplyIntent:       request.ApplyIntent,
	})
	if err != nil {
		return "", "", "", err
	}
	conflictSurface := replayConflictSurface(request.RuntimeMode, request.ApplyIntent)
	conflictKey, err := ReplayValueDigest(struct {
		ScopeKey string `json:"scopeKey"`
		Surface  string `json:"surface"`
	}{ScopeKey: scopeKey, Surface: conflictSurface})
	if err != nil {
		return "", "", "", err
	}
	_ = normalized
	return requestID, scopeKey, conflictKey, nil
}

func ReplayApplyGateIdempotencyKey(requestID string, promotionToken string) (string, error) {
	if requestID == "" || promotionToken == "" {
		return "", fmt.Errorf("request id and promotion token are required")
	}
	return ReplayValueDigest(struct {
		RequestID      string `json:"requestId"`
		PromotionToken string `json:"promotionToken"`
	}{RequestID: requestID, PromotionToken: promotionToken})
}

func replayConflictSurface(mode ReplayRuntimeMode, applyIntent bool) string {
	if applyIntent {
		return "apply"
	}
	switch mode {
	case ReplayRuntimeModeRebuild:
		return "rebuild"
	default:
		return "read-only"
	}
}

func ReplayDigestMatches(expected string, actual string) error {
	if expected == "" || actual == "" {
		return fmt.Errorf("expected and actual digests are required")
	}
	if expected != actual {
		return fmt.Errorf("digest drift: expected %q, got %q", expected, actual)
	}
	return nil
}

func canonicalJSONDigest(value any) (string, error) {
	normalized, err := normalizeJSONValue(value)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func normalizeJSONValue(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	return sortJSONValue(decoded), nil
}

func sortJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		normalized := make(map[string]any, len(typed))
		for _, key := range keys {
			normalized[key] = sortJSONValue(typed[key])
		}
		return normalized
	case []any:
		normalized := make([]any, 0, len(typed))
		for _, item := range typed {
			normalized = append(normalized, sortJSONValue(item))
		}
		return normalized
	default:
		return typed
	}
}

func validateReplayScope(scope ReplayScope) error {
	if scope.Symbol == "" || scope.WindowStart == "" || scope.WindowEnd == "" {
		return fmt.Errorf("replay scope is incomplete")
	}
	if len(scope.Venues) == 0 {
		return fmt.Errorf("replay scope venues are required")
	}
	for _, venue := range scope.Venues {
		if strings.TrimSpace(venue) == "" {
			return fmt.Errorf("replay scope venues must not be empty")
		}
	}
	return nil
}
