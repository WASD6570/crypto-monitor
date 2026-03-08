package slowcontext

import "testing"

func TestSlowContextFreshnessClassification(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	cmeResult, err := service.ParsePoll(SourceFamilyCME, loadFixture(t, "tests/fixtures/slow-context/cme-published.fixture.v1.json"), "2026-03-09T20:35:00Z")
	if err != nil {
		returnError(t, "parse cme fixture", err)
	}
	if _, err := service.RecordAccepted(cmeResult, "session", mustTime(t, "2026-03-09T20:35:00Z")); err != nil {
		returnError(t, "record cme accepted result", err)
	}
	cmeFresh, err := service.QueryAsset(AssetQuery{Asset: "BTC", MetricFamilies: []MetricFamily{MetricFamilyCMEVolume}, Now: mustTime(t, "2026-03-10T20:59:00Z")})
	if err != nil {
		returnError(t, "query fresh cme context", err)
	}
	assertFreshness(t, cmeFresh, MetricFamilyCMEVolume, FreshnessFresh)

	cmeDelayed, err := service.QueryAsset(AssetQuery{Asset: "BTC", MetricFamilies: []MetricFamily{MetricFamilyCMEVolume}, Now: mustTime(t, "2026-03-10T21:05:00Z")})
	if err != nil {
		returnError(t, "query delayed cme context", err)
	}
	assertFreshness(t, cmeDelayed, MetricFamilyCMEVolume, FreshnessDelayed)

	cmeStale, err := service.QueryAsset(AssetQuery{Asset: "BTC", MetricFamilies: []MetricFamily{MetricFamilyCMEVolume}, Now: mustTime(t, "2026-03-12T09:05:00Z")})
	if err != nil {
		returnError(t, "query stale cme context", err)
	}
	assertFreshness(t, cmeStale, MetricFamilyCMEVolume, FreshnessStale)

	etfResult, err := service.ParsePoll(SourceFamilyETF, loadFixture(t, "tests/fixtures/slow-context/etf-published.fixture.v1.json"), "2026-03-09T22:20:00Z")
	if err != nil {
		returnError(t, "parse etf fixture", err)
	}
	if _, err := service.RecordAccepted(etfResult, "daily", mustTime(t, "2026-03-09T22:20:00Z")); err != nil {
		returnError(t, "record etf accepted result", err)
	}
	etfStale, err := service.QueryAsset(AssetQuery{Asset: "BTC", MetricFamilies: []MetricFamily{MetricFamilyETFDailyFlow}, Now: mustTime(t, "2026-03-13T23:10:00Z")})
	if err != nil {
		returnError(t, "query stale etf context", err)
	}
	assertFreshness(t, etfStale, MetricFamilyETFDailyFlow, FreshnessStale)
}

func TestSlowContextUnavailableState(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	response, err := service.QueryAsset(AssetQuery{Asset: "ETH", MetricFamilies: []MetricFamily{MetricFamilyCMEOpenInterest}, Now: mustTime(t, "2026-03-10T12:00:00Z")})
	if err != nil {
		returnError(t, "query unavailable state", err)
	}
	context, ok := response.Context(MetricFamilyCMEOpenInterest)
	if !ok {
		t.Fatal("expected cme open interest context")
	}
	if context.Availability != AvailabilityUnavailable {
		t.Fatalf("availability = %q, want %q", context.Availability, AvailabilityUnavailable)
	}
	if context.Freshness != FreshnessUnavailable {
		t.Fatalf("freshness = %q, want %q", context.Freshness, FreshnessUnavailable)
	}
	if context.MessageKey != "cme_open_interest_unavailable" {
		t.Fatalf("message key = %q", context.MessageKey)
	}
}

func TestSlowContextLatestRevisionSelection(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	baseline, err := service.ParsePoll(SourceFamilyCME, loadFixture(t, "tests/fixtures/slow-context/cme-same-asof.fixture.v1.json"), "2026-03-09T20:50:00Z")
	if err != nil {
		returnError(t, "parse baseline fixture", err)
	}
	corrected, err := service.ParsePoll(SourceFamilyCME, loadFixture(t, "tests/fixtures/slow-context/cme-corrected-same-asof.fixture.v1.json"), "2026-03-09T21:10:00Z")
	if err != nil {
		returnError(t, "parse corrected fixture", err)
	}
	if _, err := service.RecordAccepted(baseline, "session", mustTime(t, "2026-03-09T20:50:00Z")); err != nil {
		returnError(t, "record baseline accepted result", err)
	}
	if _, err := service.RecordAccepted(corrected, "session", mustTime(t, "2026-03-09T21:10:00Z")); err != nil {
		returnError(t, "record corrected accepted result", err)
	}

	response, err := service.QueryAsset(AssetQuery{Asset: "ETH", MetricFamilies: []MetricFamily{MetricFamilyCMEOpenInterest}, Now: mustTime(t, "2026-03-10T12:00:00Z")})
	if err != nil {
		returnError(t, "query latest revision", err)
	}
	context, ok := response.Context(MetricFamilyCMEOpenInterest)
	if !ok {
		t.Fatal("expected cme open interest context")
	}
	if context.Revision != "2" {
		t.Fatalf("revision = %q, want %q", context.Revision, "2")
	}
	if context.Value == nil || context.Value.Amount != "9310.00" {
		t.Fatalf("value = %+v, want amount 9310.00", context.Value)
	}
	if context.PreviousValue == nil || context.PreviousValue.Amount != "9284.00" {
		t.Fatalf("previous value = %+v, want amount 9284.00", context.PreviousValue)
	}
	if context.Availability != AvailabilityAvailable {
		t.Fatalf("availability = %q, want %q", context.Availability, AvailabilityAvailable)
	}
}

func assertFreshness(t *testing.T, response AssetContextResponse, metricFamily MetricFamily, freshness FreshnessState) {
	t.Helper()
	context, ok := response.Context(metricFamily)
	if !ok {
		t.Fatalf("expected %q context", metricFamily)
	}
	if context.Freshness != freshness {
		t.Fatalf("freshness = %q, want %q", context.Freshness, freshness)
	}
	if context.Availability != AvailabilityAvailable {
		t.Fatalf("availability = %q, want %q", context.Availability, AvailabilityAvailable)
	}
	if context.ThresholdBasis.ExpectedCadence == "" {
		t.Fatalf("threshold basis missing expected cadence: %+v", context.ThresholdBasis)
	}
}
