package replayengine

import (
	"fmt"
	"strings"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ReplayAuditTrail struct {
	Request     contracts.ReplayRequestAuditRecord
	Execution   contracts.ReplayExecutionAuditRecord
	Checkpoints []contracts.ReplayCheckpointAuditRecord
	Outcome     contracts.ReplayOutcomeAuditRecord
}

type ReplayRequestWindowPolicy struct {
	MaxInspectWindow     time.Duration
	MaxMaterializeWindow time.Duration
	MaxApplyWindow       time.Duration
}

type ReplayResumeDecision struct {
	RequestID              string
	RunID                  string
	ResumeFromEventID      string
	ResumeFromScanEventID  string
	LogicalPartition       string
	NextCheckpointSequence int
	RetryCount             int
}

func DefaultReplayRequestWindowPolicy() ReplayRequestWindowPolicy {
	return ReplayRequestWindowPolicy{
		MaxInspectWindow:     24 * time.Hour,
		MaxMaterializeWindow: 24 * time.Hour,
		MaxApplyWindow:       12 * time.Hour,
	}
}

func prepareReplayRunRequest(request contracts.ReplayRunRequest, build contracts.ReplayBuildProvenance, policy ReplayRequestWindowPolicy) (contracts.ReplayRunRequest, error) {
	if err := contracts.ValidateReplayRunRequest(request); err != nil {
		return contracts.ReplayRunRequest{}, err
	}
	normalizedScope, scopeKey, err := contracts.NormalizeReplayScope(request.Scope)
	if err != nil {
		return contracts.ReplayRunRequest{}, err
	}
	start, err := time.Parse(time.RFC3339Nano, normalizedScope.WindowStart)
	if err != nil {
		return contracts.ReplayRunRequest{}, fmt.Errorf("parse replay window start: %w", err)
	}
	end, err := time.Parse(time.RFC3339Nano, normalizedScope.WindowEnd)
	if err != nil {
		return contracts.ReplayRunRequest{}, fmt.Errorf("parse replay window end: %w", err)
	}
	window := end.Sub(start)
	maxWindow := policy.MaxInspectWindow
	if request.RuntimeMode == contracts.ReplayRuntimeModeRebuild {
		maxWindow = policy.MaxMaterializeWindow
	}
	if request.ApplyIntent {
		maxWindow = policy.MaxApplyWindow
	}
	if maxWindow > 0 && window > maxWindow {
		return contracts.ReplayRunRequest{}, fmt.Errorf("replay scope window %s exceeds max %s", window, maxWindow)
	}
	request.Scope = normalizedScope
	request.ScopeKey = scopeKey
	request.RequestSource = strings.TrimSpace(request.RequestSource)
	request.OperatorNote = strings.TrimSpace(request.OperatorNote)
	requestID, _, conflictKey, err := contracts.ReplayRequestIdentity(request, build)
	if err != nil {
		return contracts.ReplayRunRequest{}, err
	}
	request.RequestID = requestID
	request.ConflictKey = conflictKey
	request.RequestDigest, err = contracts.ReplayRequestDigest(request)
	if err != nil {
		return contracts.ReplayRunRequest{}, err
	}
	return request, nil
}

func ReplayRequestsConflict(candidate contracts.ReplayRunManifest, active []contracts.ReplayRunManifest) (string, bool, error) {
	candidateStart, candidateEnd, err := replayScopeWindow(candidate.Scope)
	if err != nil {
		return "", false, err
	}
	for _, current := range active {
		if current.RequestID == candidate.RequestID {
			continue
		}
		currentStart, currentEnd, err := replayScopeWindow(current.Scope)
		if err != nil {
			return "", false, err
		}
		if !windowsOverlap(candidateStart, candidateEnd, currentStart, currentEnd) {
			continue
		}
		if candidate.Scope.Symbol != current.Scope.Symbol || !stringSetsOverlap(candidate.Scope.Venues, current.Scope.Venues) || !streamFamiliesOverlap(candidate.Scope.StreamFamilies, current.Scope.StreamFamilies) {
			continue
		}
		candidateSurface := replaySurface(candidate.RuntimeMode, candidate.ApplyIntent)
		currentSurface := replaySurface(current.RuntimeMode, current.ApplyIntent)
		if candidateSurface == "read-only" && currentSurface == "read-only" {
			continue
		}
		if candidateSurface == "rebuild" && currentSurface == "read-only" {
			continue
		}
		if candidateSurface == "read-only" && currentSurface == "rebuild" {
			continue
		}
		return current.ConflictKey, true, nil
	}
	return "", false, nil
}

func ReplayResumeFromCheckpoint(manifest contracts.ReplayRunManifest, checkpoint contracts.ReplayCheckpoint) (ReplayResumeDecision, error) {
	if err := contracts.ValidateReplayRunManifest(manifest); err != nil {
		return ReplayResumeDecision{}, err
	}
	if err := contracts.ValidateReplayCheckpoint(checkpoint); err != nil {
		return ReplayResumeDecision{}, err
	}
	if checkpoint.RequestID != manifest.RequestID {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint request drift: request=%q manifest=%q", checkpoint.RequestID, manifest.RequestID)
	}
	if checkpoint.ScopeKey != manifest.ScopeKey {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint scope drift: checkpoint=%q manifest=%q", checkpoint.ScopeKey, manifest.ScopeKey)
	}
	if checkpoint.ConfigSnapshot.Digest != manifest.ConfigSnapshot.Digest {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint config snapshot drift: checkpoint=%q manifest=%q", checkpoint.ConfigSnapshot.Digest, manifest.ConfigSnapshot.Digest)
	}
	contractSetDigest, err := contracts.ReplayContractSetDigest(manifest.ContractVersions)
	if err != nil {
		return ReplayResumeDecision{}, err
	}
	if checkpoint.ContractSetDigest != contractSetDigest {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint contract set drift: checkpoint=%q manifest=%q", checkpoint.ContractSetDigest, contractSetDigest)
	}
	buildDigest, err := contracts.ReplayBuildDigest(manifest.Build)
	if err != nil {
		return ReplayResumeDecision{}, err
	}
	if checkpoint.BuildDigest != buildDigest {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint build drift: checkpoint=%q manifest=%q", checkpoint.BuildDigest, buildDigest)
	}
	if checkpoint.OutputMode != manifest.RuntimeMode {
		return ReplayResumeDecision{}, fmt.Errorf("checkpoint output mode drift: checkpoint=%q manifest=%q", checkpoint.OutputMode, manifest.RuntimeMode)
	}
	return ReplayResumeDecision{
		RequestID:              checkpoint.RequestID,
		RunID:                  manifest.RunID,
		ResumeFromEventID:      checkpoint.LastMaterializedEventID,
		ResumeFromScanEventID:  checkpoint.LastScannedEventID,
		LogicalPartition:       checkpoint.LogicalPartition,
		NextCheckpointSequence: checkpoint.Sequence + 1,
		RetryCount:             checkpoint.RetryCount + 1,
	}, nil
}

func ReplayApplyGate(manifest contracts.ReplayRunManifest, artifacts []contracts.ReplayArtifactRef, priorOutput string, outcomes map[string]contracts.ReplayOutcomeAuditRecord, recordedAt time.Time) (contracts.ReplayOutcomeAuditRecord, error) {
	if !manifest.ApplyIntent {
		return contracts.ReplayOutcomeAuditRecord{}, fmt.Errorf("apply intent is required")
	}
	if err := contracts.ValidateReplayApprovalContext(manifest.ApprovalContext); err != nil {
		return contracts.ReplayOutcomeAuditRecord{}, err
	}
	if outcomes == nil {
		return contracts.ReplayOutcomeAuditRecord{}, fmt.Errorf("apply outcome store is required")
	}
	key, err := contracts.ReplayApplyGateIdempotencyKey(manifest.RequestID, manifest.ApprovalContext.PromotionToken)
	if err != nil {
		return contracts.ReplayOutcomeAuditRecord{}, err
	}
	if existing, ok := outcomes[key]; ok {
		return existing, nil
	}
	outcome := contracts.ReplayOutcomeAuditRecord{
		SchemaVersion:        "v1",
		RequestID:            manifest.RequestID,
		RunID:                manifest.RunID,
		TerminalStatus:       "promoted-correction",
		ArtifactRefs:         append([]contracts.ReplayArtifactRef(nil), artifacts...),
		PriorOutputReference: priorOutput,
		PromotionToken:       manifest.ApprovalContext.PromotionToken,
		CrossedApplyGate:     true,
		RecordedAt:           recordedAt.UTC().Format(time.RFC3339Nano),
	}
	if len(artifacts) > 0 {
		outcome.ReplacementOutputRef = artifacts[0].Digest
	}
	if err := contracts.ValidateReplayOutcomeAuditRecord(outcome); err != nil {
		return contracts.ReplayOutcomeAuditRecord{}, err
	}
	outcomes[key] = outcome
	return outcome, nil
}

func replayExecutionAuditRecord(manifest contracts.ReplayRunManifest, result contracts.ReplayRunResult, startedAt string, finishedAt string) contracts.ReplayExecutionAuditRecord {
	partitions := make([]string, 0, len(manifest.RawPartitions))
	for _, partition := range manifest.RawPartitions {
		partitions = append(partitions, partition.LogicalPartition)
	}
	contractSetDigest, _ := contracts.ReplayContractSetDigest(manifest.ContractVersions)
	buildDigest, _ := contracts.ReplayBuildDigest(manifest.Build)
	return contracts.ReplayExecutionAuditRecord{
		SchemaVersion:      "v1",
		RequestID:          manifest.RequestID,
		RunID:              manifest.RunID,
		ResolvedPartitions: partitions,
		ConfigSnapshot:     manifest.ConfigSnapshot,
		ContractSetDigest:  contractSetDigest,
		BuildDigest:        buildDigest,
		StartedAt:          startedAt,
		FinishedAt:         finishedAt,
		InputCounters:      result.InputCounters,
		FailureCategory:    result.FailureCategory,
	}
}

func replayOutcomeAuditRecord(manifest contracts.ReplayRunManifest, result contracts.ReplayRunResult, recordedAt string, crossedApplyGate bool) contracts.ReplayOutcomeAuditRecord {
	status := "no-op"
	switch {
	case result.FailureCategory == "approval-required":
		status = "rejected-apply"
	case manifest.RuntimeMode == contracts.ReplayRuntimeModeCompare:
		status = "compare-only"
	case manifest.RuntimeMode == contracts.ReplayRuntimeModeRebuild:
		status = "isolated-rebuild"
	}
	promotionToken := ""
	if manifest.ApprovalContext != nil {
		promotionToken = manifest.ApprovalContext.PromotionToken
	}
	return contracts.ReplayOutcomeAuditRecord{
		SchemaVersion:    "v1",
		RequestID:        manifest.RequestID,
		RunID:            manifest.RunID,
		TerminalStatus:   status,
		ArtifactRefs:     append([]contracts.ReplayArtifactRef(nil), result.ArtifactRefs...),
		PromotionToken:   promotionToken,
		CrossedApplyGate: crossedApplyGate,
		RecordedAt:       recordedAt,
	}
}

func replayCheckpoint(manifest contracts.ReplayRunManifest, ordered []ingestion.RawAppendEntry, recordedAt string) *contracts.ReplayCheckpoint {
	if len(ordered) == 0 || len(manifest.RawPartitions) == 0 {
		return nil
	}
	contractSetDigest, _ := contracts.ReplayContractSetDigest(manifest.ContractVersions)
	buildDigest, _ := contracts.ReplayBuildDigest(manifest.Build)
	checkpoint := contracts.ReplayCheckpoint{
		SchemaVersion:           "v1",
		RequestID:               manifest.RequestID,
		RunID:                   manifest.RunID,
		ScopeKey:                manifest.ScopeKey,
		LogicalPartition:        manifest.RawPartitions[len(manifest.RawPartitions)-1].LogicalPartition,
		LastMaterializedEventID: ordered[len(ordered)-1].CanonicalEventID,
		LastScannedEventID:      ordered[len(ordered)-1].CanonicalEventID,
		ConfigSnapshot:          manifest.ConfigSnapshot,
		ContractSetDigest:       contractSetDigest,
		BuildDigest:             buildDigest,
		OutputMode:              manifest.RuntimeMode,
		Sequence:                1,
		RetryCount:              0,
		RecordedAt:              recordedAt,
	}
	return &checkpoint
}

func replayCheckpointAuditRecord(checkpoint contracts.ReplayCheckpoint) contracts.ReplayCheckpointAuditRecord {
	return contracts.ReplayCheckpointAuditRecord{
		SchemaVersion:           "v1",
		RequestID:               checkpoint.RequestID,
		RunID:                   checkpoint.RunID,
		CheckpointSequence:      checkpoint.Sequence,
		LogicalPartition:        checkpoint.LogicalPartition,
		LastMaterializedEventID: checkpoint.LastMaterializedEventID,
		RetryCount:              checkpoint.RetryCount,
		FailureClass:            checkpoint.LastErrorClass,
		RecordedAt:              checkpoint.RecordedAt,
	}
}

func replayScopeWindow(scope contracts.ReplayScope) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339Nano, scope.WindowStart)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse(time.RFC3339Nano, scope.WindowEnd)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func windowsOverlap(leftStart, leftEnd, rightStart, rightEnd time.Time) bool {
	return leftStart.Before(rightEnd) && rightStart.Before(leftEnd)
}

func stringSetsOverlap(left []string, right []string) bool {
	seen := make(map[string]struct{}, len(left))
	for _, item := range left {
		seen[item] = struct{}{}
	}
	for _, item := range right {
		if _, ok := seen[item]; ok {
			return true
		}
	}
	return false
}

func streamFamiliesOverlap(left []string, right []string) bool {
	if len(left) == 0 || len(right) == 0 {
		return true
	}
	return stringSetsOverlap(left, right)
}

func replaySurface(mode contracts.ReplayRuntimeMode, applyIntent bool) string {
	if applyIntent {
		return "apply"
	}
	if mode == contracts.ReplayRuntimeModeRebuild {
		return "rebuild"
	}
	return "read-only"
}
