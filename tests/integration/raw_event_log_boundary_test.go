package integration

import (
	"reflect"
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
	replayengine "github.com/crypto-market-copilot/alerts/services/replay-engine"
)

func TestRawManifestContinuityAcrossTierTransition(t *testing.T) {
	logicalPartition := ingestion.RawPartitionKey{
		UTCDate: "2026-03-06",
		Symbol:  "BTC-USD",
		Venue:   ingestion.VenueCoinbase,
	}
	hotIndex, err := ingestion.NewRawPartitionManifestIndex([]ingestion.RawPartitionManifestRecord{{
		SchemaVersion:         "v1",
		LogicalPartition:      logicalPartition,
		StorageState:          ingestion.RawStorageStateHot,
		Location:              "hot://raw/BTC-USD/2026-03-06",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            3,
		FirstCanonicalEventID: "ce:first",
		LastCanonicalEventID:  "ce:last",
		ContinuityChecksum:    "sha256:123",
	}})
	if err != nil {
		t.Fatalf("new hot index: %v", err)
	}
	coldIndex, err := ingestion.NewRawPartitionManifestIndex([]ingestion.RawPartitionManifestRecord{{
		SchemaVersion:         "v1",
		LogicalPartition:      logicalPartition,
		StorageState:          ingestion.RawStorageStateCold,
		Location:              "cold://raw/BTC-USD/2026-03-06.zst",
		HotRetentionUntil:     "2026-04-05T00:00:00Z",
		ColdRetentionUntil:    "2027-03-06T00:00:00Z",
		EntryCount:            3,
		FirstCanonicalEventID: "ce:first",
		LastCanonicalEventID:  "ce:last",
		ContinuityChecksum:    "sha256:123",
	}})
	if err != nil {
		t.Fatalf("new cold index: %v", err)
	}

	scope := ingestion.RawPartitionLookupScope{
		Symbol: "BTC-USD",
		Venue:  ingestion.VenueCoinbase,
		Start:  mustRawBoundaryTime(t, "2026-03-06T00:00:00Z"),
		End:    mustRawBoundaryTime(t, "2026-03-06T23:59:59Z"),
	}

	hotResolver, err := replayengine.NewResolver(hotIndex)
	if err != nil {
		t.Fatalf("new hot resolver: %v", err)
	}
	coldResolver, err := replayengine.NewResolver(coldIndex)
	if err != nil {
		t.Fatalf("new cold resolver: %v", err)
	}

	hotRecords, err := hotResolver.ResolvePartitions(scope)
	if err != nil {
		t.Fatalf("resolve hot partitions: %v", err)
	}
	coldRecords, err := coldResolver.ResolvePartitions(scope)
	if err != nil {
		t.Fatalf("resolve cold partitions: %v", err)
	}
	if len(hotRecords) != 1 || len(coldRecords) != 1 {
		t.Fatalf("resolved counts = %d hot, %d cold, want 1 each", len(hotRecords), len(coldRecords))
	}
	if hotRecords[0].LogicalPartition != coldRecords[0].LogicalPartition {
		t.Fatalf("logical partitions differ: %+v vs %+v", hotRecords[0].LogicalPartition, coldRecords[0].LogicalPartition)
	}
	if !reflect.DeepEqual(
		[]any{hotRecords[0].EntryCount, hotRecords[0].FirstCanonicalEventID, hotRecords[0].LastCanonicalEventID, hotRecords[0].ContinuityChecksum},
		[]any{coldRecords[0].EntryCount, coldRecords[0].FirstCanonicalEventID, coldRecords[0].LastCanonicalEventID, coldRecords[0].ContinuityChecksum},
	) {
		t.Fatalf("continuity markers differ: hot=%+v cold=%+v", hotRecords[0], coldRecords[0])
	}
}

func mustRawBoundaryTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse boundary time %q: %v", value, err)
	}
	return parsed
}
