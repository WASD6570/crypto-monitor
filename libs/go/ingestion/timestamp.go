package ingestion

import (
	"fmt"
	"time"
)

const StrictExchangeTimestampMaxSkew = 2 * time.Second

type CanonicalTimestampStatus string

const (
	TimestampStatusNormal   CanonicalTimestampStatus = "normal"
	TimestampStatusDegraded CanonicalTimestampStatus = "degraded"
)

type TimestampFallbackReason string

const (
	TimestampReasonNone                     TimestampFallbackReason = ""
	TimestampReasonExchangeMissingOrInvalid TimestampFallbackReason = "exchange-ts-missing-or-invalid"
	TimestampReasonExchangeSkewExceeded     TimestampFallbackReason = "exchange-ts-skew-exceeded"
)

type CanonicalTimestamp struct {
	EventTime        time.Time
	ExchangeTime     time.Time
	RecvTime         time.Time
	Status           CanonicalTimestampStatus
	FallbackReason   TimestampFallbackReason
	ExchangeRecvSkew time.Duration
}

type TimestampPolicy struct {
	MaxExchangeRecvSkew time.Duration
}

func StrictTimestampPolicy() TimestampPolicy {
	return TimestampPolicy{MaxExchangeRecvSkew: StrictExchangeTimestampMaxSkew}
}

func ResolveCanonicalTimestamp(exchangeTime, recvTime time.Time, policy TimestampPolicy) (CanonicalTimestamp, error) {
	if recvTime.IsZero() {
		return CanonicalTimestamp{}, fmt.Errorf("recv time is required")
	}
	if policy.MaxExchangeRecvSkew <= 0 {
		return CanonicalTimestamp{}, fmt.Errorf("max exchange/recv skew must be positive")
	}

	result := CanonicalTimestamp{
		ExchangeTime: exchangeTime,
		RecvTime:     recvTime,
	}

	if exchangeTime.IsZero() {
		result.EventTime = recvTime
		result.Status = TimestampStatusDegraded
		result.FallbackReason = TimestampReasonExchangeMissingOrInvalid
		return result, nil
	}

	skew := recvTime.Sub(exchangeTime)
	if skew < 0 {
		skew = -skew
	}
	result.ExchangeRecvSkew = skew

	if skew > policy.MaxExchangeRecvSkew {
		result.EventTime = recvTime
		result.Status = TimestampStatusDegraded
		result.FallbackReason = TimestampReasonExchangeSkewExceeded
		return result, nil
	}

	result.EventTime = exchangeTime
	result.Status = TimestampStatusNormal
	result.FallbackReason = TimestampReasonNone
	return result, nil
}
