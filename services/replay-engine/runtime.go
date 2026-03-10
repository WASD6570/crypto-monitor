package replayengine

import (
	"fmt"
	"sort"
	"strings"
	"time"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const orderingPolicyID = "event-time-sequence-canonical-id.v1"

type ConfigSnapshot struct {
	Ref              contracts.ReplaySnapshotRef
	OrderingPolicyID string
}

type ConfigSnapshotLoader interface {
	LoadConfigSnapshot(ref contracts.ReplaySnapshotRef) (ConfigSnapshot, error)
}

type RawEntryLoader interface {
	LoadRawEntries(partitions []contracts.ReplayPartitionRef) ([]ingestion.RawAppendEntry, error)
}

type ArtifactWriter interface {
	WriteArtifact(runID string, mode contracts.ReplayRuntimeMode, kind string, payload any) (contracts.ReplayArtifactRef, error)
}

type Clock interface {
	Now() time.Time
}

type CompareTarget struct {
	ID     string
	Digest string
}

type ManifestBuilder struct {
	reader         ManifestReader
	snapshotLoader ConfigSnapshotLoader
	build          contracts.ReplayBuildProvenance
}

type Engine struct {
	reader         ManifestReader
	snapshotLoader ConfigSnapshotLoader
	entryLoader    RawEntryLoader
	artifactWriter ArtifactWriter
	clock          Clock
}

type ExecutionOutput struct {
	OrderedEntries []ingestion.RawAppendEntry
	Result         contracts.ReplayRunResult
	CompareSummary *contracts.ReplayCompareSummary
	Checkpoint     *contracts.ReplayCheckpoint
	AuditTrail     ReplayAuditTrail
}

type realClock struct{}

type replayLookupScope struct {
	streamFamily string
	sharedOnly   bool
}

func (realClock) Now() time.Time { return time.Now().UTC() }

func NewManifestBuilder(reader ManifestReader, snapshotLoader ConfigSnapshotLoader, build contracts.ReplayBuildProvenance) (*ManifestBuilder, error) {
	if reader == nil {
		return nil, fmt.Errorf("manifest reader is required")
	}
	if snapshotLoader == nil {
		return nil, fmt.Errorf("snapshot loader is required")
	}
	if build.Service == "" {
		return nil, fmt.Errorf("build provenance service is required")
	}
	return &ManifestBuilder{reader: reader, snapshotLoader: snapshotLoader, build: build}, nil
}

func (b *ManifestBuilder) Build(request contracts.ReplayRunRequest) (contracts.ReplayRunManifest, error) {
	if b == nil {
		return contracts.ReplayRunManifest{}, fmt.Errorf("manifest builder is required")
	}
	prepared, err := prepareReplayRunRequest(request, b.build, DefaultReplayRequestWindowPolicy())
	if err != nil {
		return contracts.ReplayRunManifest{}, err
	}
	request = prepared
	configSnapshot, err := b.snapshotLoader.LoadConfigSnapshot(request.ConfigSnapshot)
	if err != nil {
		return contracts.ReplayRunManifest{}, err
	}
	if err := contracts.ReplayDigestMatches(request.ConfigSnapshot.Digest, configSnapshot.Ref.Digest); err != nil {
		return contracts.ReplayRunManifest{}, fmt.Errorf("config snapshot digest mismatch: %w", err)
	}

	start, err := time.Parse(time.RFC3339Nano, request.Scope.WindowStart)
	if err != nil {
		return contracts.ReplayRunManifest{}, fmt.Errorf("parse replay window start: %w", err)
	}
	end, err := time.Parse(time.RFC3339Nano, request.Scope.WindowEnd)
	if err != nil {
		return contracts.ReplayRunManifest{}, fmt.Errorf("parse replay window end: %w", err)
	}

	partitionRefs := make([]contracts.ReplayPartitionRef, 0)
	resolvedPartitions := make(map[string]contracts.ReplayPartitionRef)
	venues := append([]string(nil), request.Scope.Venues...)
	sort.Strings(venues)
	lookupScopes := replayLookupScopes(request.Scope.StreamFamilies)
	for _, venue := range venues {
		for _, lookupScope := range lookupScopes {
			records, err := b.reader.ResolveRawPartitions(ingestion.RawPartitionLookupScope{
				Symbol:       request.Scope.Symbol,
				Venue:        replayLookupVenue(venue),
				StreamFamily: lookupScope.streamFamily,
				Start:        start,
				End:          end,
			})
			if err != nil {
				return contracts.ReplayRunManifest{}, err
			}
			for _, record := range records {
				if lookupScope.sharedOnly && record.LogicalPartition.StreamFamily != "" {
					continue
				}
				ref := contracts.ReplayPartitionRef{
					SchemaVersion:         "v1",
					LogicalPartition:      record.LogicalPartition.String(),
					StorageState:          string(record.StorageState),
					Location:              record.Location,
					EntryCount:            record.EntryCount,
					FirstCanonicalEventID: record.FirstCanonicalEventID,
					LastCanonicalEventID:  record.LastCanonicalEventID,
					ContinuityChecksum:    record.ContinuityChecksum,
				}
				if existing, ok := resolvedPartitions[ref.LogicalPartition]; ok {
					if existing != ref {
						return contracts.ReplayRunManifest{}, fmt.Errorf("resolved raw partition drift for %q", ref.LogicalPartition)
					}
					continue
				}
				resolvedPartitions[ref.LogicalPartition] = ref
				partitionRefs = append(partitionRefs, ref)
			}
		}
	}
	if len(partitionRefs) == 0 {
		return contracts.ReplayRunManifest{}, fmt.Errorf("replay scope resolved no raw partitions")
	}
	sort.Slice(partitionRefs, func(i, j int) bool {
		return partitionRefs[i].LogicalPartition < partitionRefs[j].LogicalPartition
	})

	requestDigest := request.RequestDigest
	if requestDigest == "" {
		requestDigest, err = contracts.ReplayRequestDigest(request)
		if err != nil {
			return contracts.ReplayRunManifest{}, err
		}
	}

	manifest := contracts.ReplayRunManifest{
		SchemaVersion:    "v1",
		RunID:            request.RunID,
		RequestID:        request.RequestID,
		RequestDigest:    requestDigest,
		Scope:            request.Scope,
		ScopeKey:         request.ScopeKey,
		ConflictKey:      request.ConflictKey,
		RuntimeMode:      request.RuntimeMode,
		RequestSource:    request.RequestSource,
		OperatorNote:     request.OperatorNote,
		ApplyIntent:      request.ApplyIntent,
		ApprovalContext:  request.ApprovalContext,
		Status:           contracts.ReplayManifestStatusPlanned,
		OrderingPolicyID: configSnapshot.OrderingPolicyID,
		RawPartitions:    partitionRefs,
		ConfigSnapshot:   configSnapshot.Ref,
		ContractVersions: append([]contracts.ReplaySnapshotRef(nil), request.ContractVersions...),
		Build:            b.build,
		Initiator:        request.Initiator,
	}
	if manifest.OrderingPolicyID == "" {
		manifest.OrderingPolicyID = orderingPolicyID
	}
	return manifest, contracts.ValidateReplayRunManifest(manifest)
}

func NewEngine(reader ManifestReader, snapshotLoader ConfigSnapshotLoader, entryLoader RawEntryLoader, artifactWriter ArtifactWriter) (*Engine, error) {
	if reader == nil {
		return nil, fmt.Errorf("manifest reader is required")
	}
	if snapshotLoader == nil {
		return nil, fmt.Errorf("snapshot loader is required")
	}
	if entryLoader == nil {
		return nil, fmt.Errorf("raw entry loader is required")
	}
	if artifactWriter == nil {
		return nil, fmt.Errorf("artifact writer is required")
	}
	return &Engine{reader: reader, snapshotLoader: snapshotLoader, entryLoader: entryLoader, artifactWriter: artifactWriter, clock: realClock{}}, nil
}

func (e *Engine) Execute(manifest contracts.ReplayRunManifest, compareTarget *CompareTarget) (ExecutionOutput, error) {
	if e == nil {
		return ExecutionOutput{}, fmt.Errorf("replay engine is required")
	}
	startedAt := e.clock.Now().UTC().Format(time.RFC3339Nano)
	auditTrail := ReplayAuditTrail{
		Request: contracts.ReplayRequestAuditRecord{
			SchemaVersion: "v1",
			RequestID:     manifest.RequestID,
			RunID:         manifest.RunID,
			Initiator:     manifest.Initiator.Actor,
			RequestSource: manifest.RequestSource,
			ReasonCode:    manifest.Initiator.ReasonCode,
			OperatorNote:  manifest.OperatorNote,
			ScopeKey:      manifest.ScopeKey,
			RuntimeMode:   manifest.RuntimeMode,
			ApplyIntent:   manifest.ApplyIntent,
			RecordedAt:    startedAt,
		},
	}
	result := contracts.ReplayRunResult{
		SchemaVersion: "v1",
		RunID:         manifest.RunID,
		Status:        "failed",
		StartedAt:     startedAt,
		FinishedAt:    startedAt,
	}
	manifestDigest, err := contracts.ReplayManifestDigest(manifest)
	if err == nil {
		result.ManifestDigest = manifestDigest
	}
	if err := contracts.ValidateReplayRunManifest(manifest); err != nil {
		result.FailureCategory = "invalid-manifest"
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, startedAt, false)
		return output, err
	}
	if err := contracts.ValidateReplayRequestAuditRecord(auditTrail.Request); err != nil {
		result.FailureCategory = "invalid-manifest"
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, startedAt, false)
		return output, err
	}
	if manifest.ApplyIntent {
		if err := contracts.ValidateReplayApprovalContext(manifest.ApprovalContext); err != nil {
			result.FailureCategory = "approval-required"
			result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
			output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
			output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
			output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
			return output, err
		}
	}

	configSnapshot, err := e.snapshotLoader.LoadConfigSnapshot(manifest.ConfigSnapshot)
	if err != nil {
		result.FailureCategory = "missing-snapshot"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		return ExecutionOutput{Result: result}, err
	}
	if err := contracts.ReplayDigestMatches(manifest.ConfigSnapshot.Digest, configSnapshot.Ref.Digest); err != nil {
		result.FailureCategory = "snapshot-drift"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, err
	}
	if configSnapshot.OrderingPolicyID != "" && manifest.OrderingPolicyID != configSnapshot.OrderingPolicyID {
		result.FailureCategory = "snapshot-drift"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, fmt.Errorf("ordering policy drift: manifest=%q snapshot=%q", manifest.OrderingPolicyID, configSnapshot.OrderingPolicyID)
	}
	if err := validateResolvedPartitions(e.reader, manifest); err != nil {
		result.FailureCategory = "snapshot-drift"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, err
	}
	if manifest.RuntimeMode == contracts.ReplayRuntimeModeCompare && compareTarget == nil {
		result.FailureCategory = "missing-compare-target"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, fmt.Errorf("compare target is required")
	}

	entries, err := e.entryLoader.LoadRawEntries(manifest.RawPartitions)
	if err != nil {
		result.FailureCategory = "input-load-failed"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output := ExecutionOutput{Result: result, AuditTrail: auditTrail}
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, err
	}
	ordered := sortReplayEntries(entries)
	result.InputCounters = replayCounters(ordered, manifest.RawPartitions)

	outputDigest, err := digestEntries(ordered)
	if err != nil {
		result.FailureCategory = "artifact-digest-failed"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		return ExecutionOutput{Result: result}, err
	}
	result.OutputDigest = outputDigest

	output := ExecutionOutput{OrderedEntries: ordered, Result: result, AuditTrail: auditTrail}
	switch manifest.RuntimeMode {
	case contracts.ReplayRuntimeModeInspect:
		// No rebuilt artifacts for inspect mode.
	case contracts.ReplayRuntimeModeRebuild:
		artifactRef, err := e.artifactWriter.WriteArtifact(manifest.RunID, manifest.RuntimeMode, "rebuild-output", replayArtifactPayload(ordered))
		if err != nil {
			result.FailureCategory = "artifact-write-failed"
			result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
			output.Result = result
			output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
			output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
			return output, err
		}
		result.ArtifactRefs = []contracts.ReplayArtifactRef{artifactRef}
	case contracts.ReplayRuntimeModeCompare:
		summary := compareSummary(manifest.RunID, outputDigest, compareTarget)
		ref, err := e.artifactWriter.WriteArtifact(manifest.RunID, manifest.RuntimeMode, "compare-summary", summary)
		if err != nil {
			result.FailureCategory = "artifact-write-failed"
			result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
			output.Result = result
			output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
			output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
			return output, err
		}
		result.ComparisonSummaryRef = &ref
		output.CompareSummary = &summary
	default:
		result.FailureCategory = "unsupported-mode"
		result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
		output.Result = result
		output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
		output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
		return output, fmt.Errorf("unsupported replay runtime mode %q", manifest.RuntimeMode)
	}

	result.Status = "completed"
	result.FinishedAt = e.clock.Now().UTC().Format(time.RFC3339Nano)
	output.Result = result
	output.Checkpoint = replayCheckpoint(manifest, ordered, result.FinishedAt)
	output.AuditTrail.Execution = replayExecutionAuditRecord(manifest, result, startedAt, result.FinishedAt)
	output.AuditTrail.Outcome = replayOutcomeAuditRecord(manifest, result, result.FinishedAt, false)
	if output.Checkpoint != nil {
		output.AuditTrail.Checkpoints = []contracts.ReplayCheckpointAuditRecord{replayCheckpointAuditRecord(*output.Checkpoint)}
	}
	return output, contracts.ValidateReplayRunResult(output.Result)
}

func validateResolvedPartitions(reader ManifestReader, manifest contracts.ReplayRunManifest) error {
	start, err := time.Parse(time.RFC3339Nano, manifest.Scope.WindowStart)
	if err != nil {
		return fmt.Errorf("parse replay window start: %w", err)
	}
	end, err := time.Parse(time.RFC3339Nano, manifest.Scope.WindowEnd)
	if err != nil {
		return fmt.Errorf("parse replay window end: %w", err)
	}
	expected := make(map[string]contracts.ReplayPartitionRef, len(manifest.RawPartitions))
	for _, ref := range manifest.RawPartitions {
		expected[ref.LogicalPartition] = ref
	}
	lookupScopes := replayLookupScopes(manifest.Scope.StreamFamilies)
	for _, venue := range manifest.Scope.Venues {
		for _, lookupScope := range lookupScopes {
			records, err := reader.ResolveRawPartitions(ingestion.RawPartitionLookupScope{
				Symbol:       manifest.Scope.Symbol,
				Venue:        replayLookupVenue(venue),
				StreamFamily: lookupScope.streamFamily,
				Start:        start,
				End:          end,
			})
			if err != nil {
				return err
			}
			for _, record := range records {
				if lookupScope.sharedOnly && record.LogicalPartition.StreamFamily != "" {
					continue
				}
				ref, ok := expected[record.LogicalPartition.String()]
				if !ok {
					return fmt.Errorf("resolved unexpected raw partition %q", record.LogicalPartition.String())
				}
				if ref.ContinuityChecksum != record.ContinuityChecksum || ref.EntryCount != record.EntryCount || ref.Location != record.Location || (ref.StorageState != "" && ref.StorageState != string(record.StorageState)) {
					return fmt.Errorf("resolved raw partition drift for %q", record.LogicalPartition.String())
				}
				delete(expected, record.LogicalPartition.String())
			}
		}
	}
	if len(expected) > 0 {
		missing := make([]string, 0, len(expected))
		for key := range expected {
			missing = append(missing, key)
		}
		sort.Strings(missing)
		return fmt.Errorf("resolved raw partitions missing manifest refs: %s", strings.Join(missing, ", "))
	}
	return nil
}

func sortReplayEntries(entries []ingestion.RawAppendEntry) []ingestion.RawAppendEntry {
	ordered := append([]ingestion.RawAppendEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := ordered[i]
		right := ordered[j]
		if left.BucketTimestamp != right.BucketTimestamp {
			return left.BucketTimestamp < right.BucketTimestamp
		}
		leftHasSequence := left.VenueSequence > 0
		rightHasSequence := right.VenueSequence > 0
		if leftHasSequence != rightHasSequence {
			return leftHasSequence
		}
		if leftHasSequence && left.VenueSequence != right.VenueSequence {
			return left.VenueSequence < right.VenueSequence
		}
		return left.CanonicalEventID < right.CanonicalEventID
	})
	return ordered
}

func replayCounters(entries []ingestion.RawAppendEntry, partitions []contracts.ReplayPartitionRef) contracts.ReplayInputCounters {
	counters := contracts.ReplayInputCounters{Partitions: len(partitions), Events: len(entries)}
	for _, entry := range entries {
		if entry.DuplicateAudit.Duplicate {
			counters.Duplicates++
		}
		if entry.Late {
			counters.LateEvents++
		}
		if entry.BucketTimestampSource == ingestion.RawBucketTimestampSourceRecv || entry.TimestampDegradationReason != ingestion.TimestampReasonNone {
			counters.DegradedTimestampEvents++
		}
	}
	return counters
}

func digestEntries(entries []ingestion.RawAppendEntry) (string, error) {
	type replayDigestEntry struct {
		CanonicalEventID           string                             `json:"canonicalEventId"`
		VenueMessageID             string                             `json:"venueMessageId"`
		VenueSequence              int64                              `json:"venueSequence"`
		StreamKey                  string                             `json:"streamKey"`
		StreamFamily               string                             `json:"streamFamily"`
		BucketTimestamp            string                             `json:"bucketTimestamp"`
		BucketTimestampSource      ingestion.RawBucketTimestampSource `json:"bucketTimestampSource"`
		TimestampDegradationReason ingestion.TimestampFallbackReason  `json:"timestampDegradationReason,omitempty"`
		Late                       bool                               `json:"late"`
		DegradedFeedRef            string                             `json:"degradedFeedRef,omitempty"`
		DuplicateIdentityKey       string                             `json:"duplicateIdentityKey,omitempty"`
		DuplicateOccurrence        int                                `json:"duplicateOccurrence,omitempty"`
		Duplicate                  bool                               `json:"duplicate"`
	}
	digestEntries := make([]replayDigestEntry, 0, len(entries))
	for _, entry := range entries {
		digestEntries = append(digestEntries, replayDigestEntry{
			CanonicalEventID:           entry.CanonicalEventID,
			VenueMessageID:             entry.VenueMessageID,
			VenueSequence:              entry.VenueSequence,
			StreamKey:                  entry.StreamKey,
			StreamFamily:               entry.StreamFamily,
			BucketTimestamp:            entry.BucketTimestamp,
			BucketTimestampSource:      entry.BucketTimestampSource,
			TimestampDegradationReason: entry.TimestampDegradationReason,
			Late:                       entry.Late,
			DegradedFeedRef:            entry.DegradedFeedRef,
			DuplicateIdentityKey:       entry.DuplicateAudit.IdentityKey,
			DuplicateOccurrence:        entry.DuplicateAudit.Occurrence,
			Duplicate:                  entry.DuplicateAudit.Duplicate,
		})
	}
	return contracts.ReplayValueDigest(struct {
		Entries []replayDigestEntry `json:"entries"`
	}{Entries: digestEntries})
}

func replayLookupScopes(streamFamilies []string) []replayLookupScope {
	if len(streamFamilies) == 0 {
		return []replayLookupScope{{streamFamily: ""}}
	}
	seen := make(map[replayLookupScope]struct{}, len(streamFamilies))
	lookupScopes := make([]replayLookupScope, 0, len(streamFamilies))
	for _, streamFamily := range streamFamilies {
		lookupFamily := replayLookupFamily(streamFamily)
		scope := replayLookupScope{streamFamily: lookupFamily, sharedOnly: lookupFamily == ""}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		lookupScopes = append(lookupScopes, scope)
	}
	sort.Slice(lookupScopes, func(i, j int) bool {
		return lookupScopes[i].streamFamily < lookupScopes[j].streamFamily
	})
	return lookupScopes
}

func replayLookupFamily(streamFamily string) string {
	switch streamFamily {
	case string(ingestion.StreamOrderBook), string(ingestion.StreamFundingRate), string(ingestion.StreamOpenInterest), string(ingestion.StreamMarkIndex), string(ingestion.StreamLiquidation), ingestion.RawStreamFamilyFeedHealth:
		return streamFamily
	default:
		return ""
	}
}

func replayLookupVenue(venue string) ingestion.Venue {
	return ingestion.Venue(strings.ToUpper(strings.TrimSpace(venue)))
}

func replayArtifactPayload(entries []ingestion.RawAppendEntry) map[string]any {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.CanonicalEventID)
	}
	return map[string]any{"orderedCanonicalEventIds": ids}
}

func compareSummary(runID string, outputDigest string, target *CompareTarget) contracts.ReplayCompareSummary {
	summary := contracts.ReplayCompareSummary{
		SchemaVersion:       "v1",
		RunID:               runID,
		ComparedTargetID:    target.ID,
		ArtifactDigests:     map[string]string{"rebuild-output": outputDigest, "target": target.Digest},
		AuditOnly:           true,
		DriftClassification: "match",
		UnchangedCount:      1,
	}
	if target.Digest != outputDigest {
		summary.ChangedArtifactCount = 1
		summary.UnchangedCount = 0
		summary.FirstMismatch = "rebuild-output"
		summary.DriftClassification = "drift"
	}
	return summary
}
