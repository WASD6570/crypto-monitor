package ingestion

import (
	"testing"
	"time"
)

func TestResolveCanonicalTimestampKeepsPlausibleExchangeTime(t *testing.T) {
	exchangeTime := time.Date(2026, time.March, 6, 12, 0, 2, 100000000, time.UTC)
	recvTime := time.Date(2026, time.March, 6, 12, 0, 2, 130000000, time.UTC)

	resolved, err := ResolveCanonicalTimestamp(exchangeTime, recvTime, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("resolve canonical timestamp: %v", err)
	}
	if !resolved.EventTime.Equal(exchangeTime) {
		t.Fatalf("event time = %s, want %s", resolved.EventTime, exchangeTime)
	}
	if resolved.Status != TimestampStatusNormal {
		t.Fatalf("status = %q, want %q", resolved.Status, TimestampStatusNormal)
	}
	if resolved.FallbackReason != TimestampReasonNone {
		t.Fatalf("fallback reason = %q, want none", resolved.FallbackReason)
	}
}

func TestResolveCanonicalTimestampFallsBackWhenExchangeMissing(t *testing.T) {
	recvTime := time.Date(2026, time.March, 6, 12, 2, 20, 30000000, time.UTC)

	resolved, err := ResolveCanonicalTimestamp(time.Time{}, recvTime, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("resolve canonical timestamp: %v", err)
	}
	if !resolved.EventTime.Equal(recvTime) {
		t.Fatalf("event time = %s, want %s", resolved.EventTime, recvTime)
	}
	if resolved.Status != TimestampStatusDegraded {
		t.Fatalf("status = %q, want %q", resolved.Status, TimestampStatusDegraded)
	}
	if resolved.FallbackReason != TimestampReasonExchangeMissingOrInvalid {
		t.Fatalf("fallback reason = %q, want %q", resolved.FallbackReason, TimestampReasonExchangeMissingOrInvalid)
	}
}

func TestResolveCanonicalTimestampFallsBackWhenStrictSkewExceeded(t *testing.T) {
	exchangeTime := time.Date(2026, time.March, 6, 12, 2, 17, 0, time.UTC)
	recvTime := time.Date(2026, time.March, 6, 12, 2, 20, 30000000, time.UTC)

	resolved, err := ResolveCanonicalTimestamp(exchangeTime, recvTime, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("resolve canonical timestamp: %v", err)
	}
	if !resolved.EventTime.Equal(recvTime) {
		t.Fatalf("event time = %s, want %s", resolved.EventTime, recvTime)
	}
	if resolved.Status != TimestampStatusDegraded {
		t.Fatalf("status = %q, want %q", resolved.Status, TimestampStatusDegraded)
	}
	if resolved.FallbackReason != TimestampReasonExchangeSkewExceeded {
		t.Fatalf("fallback reason = %q, want %q", resolved.FallbackReason, TimestampReasonExchangeSkewExceeded)
	}
	if resolved.ExchangeRecvSkew <= StrictExchangeTimestampMaxSkew {
		t.Fatalf("skew = %s, want > %s", resolved.ExchangeRecvSkew, StrictExchangeTimestampMaxSkew)
	}
}

func TestResolveCanonicalTimestampAcceptsExactStrictBoundary(t *testing.T) {
	exchangeTime := time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC)
	recvTime := exchangeTime.Add(StrictExchangeTimestampMaxSkew)

	resolved, err := ResolveCanonicalTimestamp(exchangeTime, recvTime, StrictTimestampPolicy())
	if err != nil {
		t.Fatalf("resolve canonical timestamp: %v", err)
	}
	if !resolved.EventTime.Equal(exchangeTime) {
		t.Fatalf("event time = %s, want %s", resolved.EventTime, exchangeTime)
	}
	if resolved.Status != TimestampStatusNormal {
		t.Fatalf("status = %q, want %q", resolved.Status, TimestampStatusNormal)
	}
}
