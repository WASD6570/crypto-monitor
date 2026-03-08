package slowcontext

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestServiceRegistersSourceAdapters(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	if _, err := service.AdapterFor(SourceFamilyCME); err != nil {
		returnError(t, "cme adapter", err)
	}
	if _, err := service.AdapterFor(SourceFamilyETF); err != nil {
		returnError(t, "etf adapter", err)
	}
	if _, err := service.AdapterFor(SourceFamily("MISSING")); err == nil {
		t.Fatal("expected unsupported source family to fail")
	}
	if _, err := NewService(NewCMEAdapter(), NewCMEAdapter()); err == nil {
		t.Fatal("expected duplicate source adapters to fail")
	}
}

func TestCMEAdapterParsesPublishedFixture(t *testing.T) {
	raw := loadFixture(t, "tests/fixtures/slow-context/cme-published.fixture.v1.json")
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	result, err := service.ParsePoll(SourceFamilyCME, raw, "2026-03-09T20:35:00Z")
	if err != nil {
		returnError(t, "parse cme published fixture", err)
	}

	assertResult(t, result, PollResult{
		SourceFamily: SourceFamilyCME,
		MetricFamily: MetricFamilyCMEVolume,
		Status:       PublicationStatusNew,
		SourceKey:    "cme.daily.btc",
		Asset:        "BTC",
		AsOfTs:       mustTime(t, "2026-03-09T00:00:00Z"),
		PublishedTs:  mustTime(t, "2026-03-09T20:30:00Z"),
		IngestTs:     mustTime(t, "2026-03-09T20:35:00Z"),
		DedupeKey:    "CME|cme_volume|BTC|cme.daily.btc|2026-03-09T00:00:00Z",
		Revision:     "1",
		Value:        &CandidateValue{Amount: "18342.55", Unit: "contracts"},
	})
}

func TestETFAdapterParsesPublishedFixture(t *testing.T) {
	raw := loadFixture(t, "tests/fixtures/slow-context/etf-published.fixture.v1.json")
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	result, err := service.ParsePoll(SourceFamilyETF, raw, "2026-03-09T22:20:00Z")
	if err != nil {
		returnError(t, "parse etf published fixture", err)
	}

	assertResult(t, result, PollResult{
		SourceFamily: SourceFamilyETF,
		MetricFamily: MetricFamilyETFDailyFlow,
		Status:       PublicationStatusNew,
		SourceKey:    "etf.daily.flow",
		Asset:        "BTC",
		AsOfTs:       mustTime(t, "2026-03-09T00:00:00Z"),
		PublishedTs:  mustTime(t, "2026-03-09T22:15:00Z"),
		IngestTs:     mustTime(t, "2026-03-09T22:20:00Z"),
		DedupeKey:    "ETF|etf_daily_flow|BTC|etf.daily.flow|2026-03-09T00:00:00Z",
		Revision:     "1",
		Value:        &CandidateValue{Amount: "245000000.00", Unit: "usd"},
	})
}

func TestCMEAdapterClassifiesRepeatedSameAsOfFixture(t *testing.T) {
	raw := loadFixture(t, "tests/fixtures/slow-context/cme-same-asof.fixture.v1.json")
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	result, err := service.ParsePoll(SourceFamilyCME, raw, "2026-03-09T20:50:00Z")
	if err != nil {
		returnError(t, "parse cme same-asof fixture", err)
	}

	if result.Status != PublicationStatusSameAsOf {
		t.Fatalf("status = %q, want %q", result.Status, PublicationStatusSameAsOf)
	}
	if result.DedupeKey != "CME|cme_open_interest|ETH|cme.daily.eth|2026-03-09T00:00:00Z" {
		t.Fatalf("dedupe key = %q", result.DedupeKey)
	}
	if result.Value == nil || result.Value.Amount != "9284.00" {
		t.Fatalf("value = %+v, want amount 9284.00", result.Value)
	}
	if result.PublishedTs.IsZero() {
		t.Fatal("expected published timestamp to be preserved")
	}
}

func TestCMEAdapterClassifiesNotYetPublishedFixture(t *testing.T) {
	raw := loadFixture(t, "tests/fixtures/slow-context/cme-not-yet-published.fixture.v1.json")
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	result, err := service.ParsePoll(SourceFamilyCME, raw, "2026-03-10T13:00:00Z")
	if err != nil {
		returnError(t, "parse cme not-yet fixture", err)
	}

	if result.Status != PublicationStatusNotYet {
		t.Fatalf("status = %q, want %q", result.Status, PublicationStatusNotYet)
	}
	if result.Value != nil {
		t.Fatalf("value = %+v, want nil", result.Value)
	}
	if result.PublishedTs != (time.Time{}) {
		t.Fatalf("published timestamp = %s, want zero", result.PublishedTs)
	}
	if result.DedupeKey != "CME|cme_open_interest|ETH|cme.daily.eth|2026-03-10T00:00:00Z" {
		t.Fatalf("dedupe key = %q", result.DedupeKey)
	}
}

func TestCMEAdapterReturnsExplicitParseFailure(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	_, err = service.ParsePoll(SourceFamilyCME, []byte(`{"sourceKey":1`), "2026-03-09T20:35:00Z")
	if err == nil {
		t.Fatal("expected malformed payload to fail")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected parse error, got %T", err)
	}
	if parseErr.SourceFamily != SourceFamilyCME {
		t.Fatalf("parse error source family = %q, want %q", parseErr.SourceFamily, SourceFamilyCME)
	}
}

func assertResult(t *testing.T, actual PollResult, want PollResult) {
	t.Helper()
	if actual.SourceFamily != want.SourceFamily {
		t.Fatalf("source family = %q, want %q", actual.SourceFamily, want.SourceFamily)
	}
	if actual.MetricFamily != want.MetricFamily {
		t.Fatalf("metric family = %q, want %q", actual.MetricFamily, want.MetricFamily)
	}
	if actual.Status != want.Status {
		t.Fatalf("status = %q, want %q", actual.Status, want.Status)
	}
	if actual.SourceKey != want.SourceKey {
		t.Fatalf("source key = %q, want %q", actual.SourceKey, want.SourceKey)
	}
	if actual.Asset != want.Asset {
		t.Fatalf("asset = %q, want %q", actual.Asset, want.Asset)
	}
	if !actual.AsOfTs.Equal(want.AsOfTs) {
		t.Fatalf("as-of = %s, want %s", actual.AsOfTs, want.AsOfTs)
	}
	if !actual.PublishedTs.Equal(want.PublishedTs) {
		t.Fatalf("published = %s, want %s", actual.PublishedTs, want.PublishedTs)
	}
	if !actual.IngestTs.Equal(want.IngestTs) {
		t.Fatalf("ingest = %s, want %s", actual.IngestTs, want.IngestTs)
	}
	if actual.DedupeKey != want.DedupeKey {
		t.Fatalf("dedupe key = %q, want %q", actual.DedupeKey, want.DedupeKey)
	}
	if actual.Revision != want.Revision {
		t.Fatalf("revision = %q, want %q", actual.Revision, want.Revision)
	}
	if actual.Value == nil || want.Value == nil {
		t.Fatalf("expected non-nil values, got actual=%+v want=%+v", actual.Value, want.Value)
	}
	if *actual.Value != *want.Value {
		t.Fatalf("value = %+v, want %+v", *actual.Value, *want.Value)
	}
}

func returnError(t *testing.T, action string, err error) {
	t.Helper()
	t.Fatalf("%s: %v", action, err)
}

func loadFixture(t *testing.T, relativePath string) []byte {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot(t), relativePath))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relativePath, err)
	}
	return contents
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}
