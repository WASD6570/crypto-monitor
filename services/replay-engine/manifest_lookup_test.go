package replayengine

import (
	"testing"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

func TestReplayPartitionLookupDoesNotGuessStoragePaths(t *testing.T) {
	reader := &stubManifestReader{
		records: []ingestion.RawPartitionManifestRecord{{
			SchemaVersion:         "v1",
			LogicalPartition:      ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase},
			StorageState:          ingestion.RawStorageStateHot,
			Location:              "hot://raw/BTC-USD/2026-03-06",
			HotRetentionUntil:     "2026-04-05T00:00:00Z",
			ColdRetentionUntil:    "2027-03-06T00:00:00Z",
			EntryCount:            2,
			FirstCanonicalEventID: "ce:first",
			LastCanonicalEventID:  "ce:last",
			ContinuityChecksum:    "sha256:abc",
		}},
	}
	resolver, err := NewResolver(reader)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	scope := ingestion.RawPartitionLookupScope{
		Symbol: "BTC-USD",
		Venue:  ingestion.VenueCoinbase,
		Start:  mustReplayTime(t, "2026-03-06T00:00:00Z"),
		End:    mustReplayTime(t, "2026-03-06T23:59:59Z"),
	}

	records, err := resolver.ResolvePartitions(scope)
	if err != nil {
		t.Fatalf("resolve partitions: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("resolved records = %d, want 1", len(records))
	}
	if reader.scope != scope {
		t.Fatalf("resolver scope = %+v, want %+v", reader.scope, scope)
	}
	if records[0].Location != "hot://raw/BTC-USD/2026-03-06" {
		t.Fatalf("location = %q, want manifest-provided location", records[0].Location)
	}
}

func TestReplayScopeResolutionDoesNotGuessTierPaths(t *testing.T) {
	reader := &stubManifestReader{
		records: []ingestion.RawPartitionManifestRecord{{
			SchemaVersion:         "v1",
			LogicalPartition:      ingestion.RawPartitionKey{UTCDate: "2026-03-06", Symbol: "BTC-USD", Venue: ingestion.VenueCoinbase, StreamFamily: "trades"},
			StorageState:          ingestion.RawStorageStateCold,
			Location:              "cold://archive/BTC-USD/2026-03-06/trades",
			HotRetentionUntil:     "2026-04-05T00:00:00Z",
			ColdRetentionUntil:    "2027-03-06T00:00:00Z",
			EntryCount:            2,
			FirstCanonicalEventID: "ce:first",
			LastCanonicalEventID:  "ce:last",
			ContinuityChecksum:    "sha256:abc",
		}},
	}
	resolver, err := NewResolver(reader)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	records, err := resolver.ResolvePartitions(ingestion.RawPartitionLookupScope{
		Symbol:       "BTC-USD",
		Venue:        ingestion.VenueCoinbase,
		StreamFamily: "trades",
		Start:        mustReplayTime(t, "2026-03-06T00:00:00Z"),
		End:          mustReplayTime(t, "2026-03-06T23:59:59Z"),
	})
	if err != nil {
		t.Fatalf("resolve partitions: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("resolved records = %d, want 1", len(records))
	}
	if records[0].StorageState != ingestion.RawStorageStateCold {
		t.Fatalf("storage state = %q, want cold", records[0].StorageState)
	}
	if records[0].Location != "cold://archive/BTC-USD/2026-03-06/trades" {
		t.Fatalf("location = %q, want manifest-provided cold location", records[0].Location)
	}
}

type stubManifestReader struct {
	scope   ingestion.RawPartitionLookupScope
	records []ingestion.RawPartitionManifestRecord
}

func (r *stubManifestReader) ResolveRawPartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error) {
	r.scope = scope
	resolved := make([]ingestion.RawPartitionManifestRecord, 0, len(r.records))
	for _, record := range r.records {
		if record.LogicalPartition.Symbol != scope.Symbol || record.LogicalPartition.Venue != scope.Venue {
			continue
		}
		if scope.StreamFamily != "" && record.LogicalPartition.StreamFamily != scope.StreamFamily {
			continue
		}
		resolved = append(resolved, record)
	}
	return resolved, nil
}

func mustReplayTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse replay time %q: %v", value, err)
	}
	return parsed
}
