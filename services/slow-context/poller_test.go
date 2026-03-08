package slowcontext

import (
	"errors"
	"testing"
	"time"
)

func TestSlowContextAdapterParsesPublishedFixtures(t *testing.T) {
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	for _, testCase := range []struct {
		name         string
		sourceFamily SourceFamily
		fixturePath  string
		ingestTs     string
		metricFamily MetricFamily
	}{
		{name: "cme", sourceFamily: SourceFamilyCME, fixturePath: "tests/fixtures/slow-context/cme-published.fixture.v1.json", ingestTs: "2026-03-09T20:35:00Z", metricFamily: MetricFamilyCMEVolume},
		{name: "etf", sourceFamily: SourceFamilyETF, fixturePath: "tests/fixtures/slow-context/etf-published.fixture.v1.json", ingestTs: "2026-03-09T22:20:00Z", metricFamily: MetricFamilyETFDailyFlow},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := service.ParsePoll(testCase.sourceFamily, loadFixture(t, testCase.fixturePath), testCase.ingestTs)
			if err != nil {
				returnError(t, "parse published fixture", err)
			}
			if result.Status != PublicationStatusNew {
				t.Fatalf("status = %q, want %q", result.Status, PublicationStatusNew)
			}
			if result.MetricFamily != testCase.metricFamily {
				t.Fatalf("metric family = %q, want %q", result.MetricFamily, testCase.metricFamily)
			}
			if result.PublishedTs.IsZero() || result.IngestTs.IsZero() || result.AsOfTs.IsZero() {
				t.Fatalf("expected timestamps to be preserved: %+v", result)
			}
			if result.DedupeKey == "" {
				t.Fatal("expected dedupe key to be populated")
			}
		})
	}
}

func TestSlowContextRepeatedPollingIsIdempotent(t *testing.T) {
	tracker, err := NewTracker()
	if err != nil {
		returnError(t, "new tracker", err)
	}
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}

	result, err := service.ParsePoll(SourceFamilyCME, loadFixture(t, "tests/fixtures/slow-context/cme-same-asof.fixture.v1.json"), "2026-03-09T20:50:00Z")
	if err != nil {
		returnError(t, "parse repeated fixture", err)
	}

	firstStatus, firstDecision, err := tracker.ObservePoll(mustTime(t, "2026-03-09T20:50:00Z"), result)
	if err != nil {
		returnError(t, "observe first same-asof poll", err)
	}
	secondStatus, secondDecision, err := tracker.ObservePoll(mustTime(t, "2026-03-09T20:55:00Z"), result)
	if err != nil {
		returnError(t, "observe second same-asof poll", err)
	}

	if !firstDecision.Accepted {
		t.Fatal("expected first repeated poll fixture to seed accepted state")
	}
	if secondDecision.Accepted {
		t.Fatal("expected second repeated same-asof poll to avoid duplicate acceptance")
	}
	if !secondDecision.Duplicate {
		t.Fatal("expected second repeated same-asof poll to be marked duplicate")
	}
	if secondStatus.HealthState != SourceHealthHealthy {
		t.Fatalf("health state = %q, want %q", secondStatus.HealthState, SourceHealthHealthy)
	}
	if !secondStatus.LastSuccessfulPublicationAt.Equal(mustTime(t, "2026-03-09T20:55:00Z")) {
		t.Fatalf("last successful publication at = %s", secondStatus.LastSuccessfulPublicationAt)
	}
	if firstStatus.LastAcceptedDedupeKey != secondStatus.LastAcceptedDedupeKey {
		t.Fatalf("accepted dedupe key changed across duplicate polls: %q vs %q", firstStatus.LastAcceptedDedupeKey, secondStatus.LastAcceptedDedupeKey)
	}
}

func TestSlowContextDelayedPublicationClassification(t *testing.T) {
	tracker, err := NewTracker()
	if err != nil {
		returnError(t, "new tracker", err)
	}
	service, err := NewService()
	if err != nil {
		returnError(t, "new service", err)
	}
	result, err := service.ParsePoll(SourceFamilyCME, loadFixture(t, "tests/fixtures/slow-context/cme-not-yet-published.fixture.v1.json"), "2026-03-10T20:30:00Z")
	if err != nil {
		returnError(t, "parse not-yet-published fixture", err)
	}

	insideWindow, _, err := tracker.ObservePoll(mustTime(t, "2026-03-10T20:30:00Z"), result)
	if err != nil {
		returnError(t, "observe inside-window poll", err)
	}
	if insideWindow.HealthState != SourceHealthHealthy {
		t.Fatalf("inside-window health state = %q, want %q", insideWindow.HealthState, SourceHealthHealthy)
	}

	delayed, _, err := tracker.ObservePoll(mustTime(t, "2026-03-10T21:05:00Z"), result)
	if err != nil {
		returnError(t, "observe delayed poll", err)
	}
	if delayed.HealthState != SourceHealthDelayedPublication {
		t.Fatalf("delayed health state = %q, want %q", delayed.HealthState, SourceHealthDelayedPublication)
	}
	if delayed.ConsecutiveDelayedPolls != 1 {
		t.Fatalf("delayed poll count = %d, want 1", delayed.ConsecutiveDelayedPolls)
	}

	repeatedDelayed, _, err := tracker.ObservePoll(mustTime(t, "2026-03-10T21:20:00Z"), result)
	if err != nil {
		returnError(t, "observe repeated delayed poll", err)
	}
	if repeatedDelayed.ConsecutiveDelayedPolls != 2 {
		t.Fatalf("repeated delayed poll count = %d, want 2", repeatedDelayed.ConsecutiveDelayedPolls)
	}
}

func TestSlowContextCorrectionHandling(t *testing.T) {
	tracker, err := NewTracker()
	if err != nil {
		returnError(t, "new tracker", err)
	}
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

	baselineStatus, baselineDecision, err := tracker.ObservePoll(mustTime(t, "2026-03-09T20:50:00Z"), baseline)
	if err != nil {
		returnError(t, "observe baseline poll", err)
	}
	correctedStatus, correctedDecision, err := tracker.ObservePoll(mustTime(t, "2026-03-09T21:10:00Z"), corrected)
	if err != nil {
		returnError(t, "observe corrected poll", err)
	}

	if !baselineDecision.Accepted {
		t.Fatal("expected baseline same-asof result to seed accepted state")
	}
	if !correctedDecision.Accepted {
		t.Fatal("expected corrected same-asof result to be accepted")
	}
	if !correctedDecision.Correction {
		t.Fatal("expected corrected same-asof result to be marked as correction")
	}
	if correctedStatus.LastAcceptedRevision != "2" {
		t.Fatalf("last accepted revision = %q, want %q", correctedStatus.LastAcceptedRevision, "2")
	}
	if correctedStatus.LastPublishedValue == nil || correctedStatus.LastPublishedValue.Amount != "9310.00" {
		t.Fatalf("last published value = %+v, want amount 9310.00", correctedStatus.LastPublishedValue)
	}
	if baselineStatus.LastAcceptedDedupeKey != correctedStatus.LastAcceptedDedupeKey {
		t.Fatalf("correction should preserve dedupe key identity: %q vs %q", baselineStatus.LastAcceptedDedupeKey, correctedStatus.LastAcceptedDedupeKey)
	}
}

func TestSlowContextSourceFailuresStayIsolated(t *testing.T) {
	tracker, err := NewTracker()
	if err != nil {
		returnError(t, "new tracker", err)
	}

	identity := SourceIdentity{
		SourceFamily: SourceFamilyETF,
		MetricFamily: MetricFamilyETFDailyFlow,
		SourceKey:    "etf.daily.flow",
		Asset:        "BTC",
	}
	status, err := tracker.ObserveFailure(identity, mustTime(t, "2026-03-09T22:30:00Z"), errors.New("source timeout"))
	if err != nil {
		returnError(t, "observe source-unavailable failure", err)
	}
	if status.HealthState != SourceHealthSourceUnavailable {
		t.Fatalf("health state = %q, want %q", status.HealthState, SourceHealthSourceUnavailable)
	}
	if status.ConsecutiveFailures != 1 {
		t.Fatalf("consecutive failures = %d, want 1", status.ConsecutiveFailures)
	}

	parseFailure, err := tracker.ObserveFailure(identity, mustTime(t, "2026-03-09T22:31:00Z"), &ParseError{SourceFamily: SourceFamilyETF, Err: errors.New("malformed payload")})
	if err != nil {
		returnError(t, "observe parse failure", err)
	}
	if parseFailure.HealthState != SourceHealthParseFailed {
		t.Fatalf("health state = %q, want %q", parseFailure.HealthState, SourceHealthParseFailed)
	}
	if parseFailure.ConsecutiveFailures != 2 {
		t.Fatalf("consecutive failures = %d, want 2", parseFailure.ConsecutiveFailures)
	}
	if parseFailure.LastError == "" {
		t.Fatal("expected last error to be recorded")
	}
}

func TestScheduleUsesPublishWindowCadence(t *testing.T) {
	tracker, err := NewTracker()
	if err != nil {
		returnError(t, "new tracker", err)
	}
	inWindow, err := tracker.PollCadence(SourceFamilyCME, mustTime(t, "2026-03-09T20:10:00Z"))
	if err != nil {
		returnError(t, "poll cadence in window", err)
	}
	outsideWindow, err := tracker.PollCadence(SourceFamilyCME, mustTime(t, "2026-03-09T10:10:00Z"))
	if err != nil {
		returnError(t, "poll cadence outside window", err)
	}
	if inWindow != 15*time.Minute {
		t.Fatalf("in-window cadence = %s, want %s", inWindow, 15*time.Minute)
	}
	if outsideWindow != time.Hour {
		t.Fatalf("outside-window cadence = %s, want %s", outsideWindow, time.Hour)
	}
}
