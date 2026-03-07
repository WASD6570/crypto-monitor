package ingestion

import (
	"testing"
	"time"
)

func TestReconnectDelayReturnsDeterministicExponentialBackoff(t *testing.T) {
	delay, err := ReconnectDelay(3, time.Second, 10*time.Second)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}

	if delay != 4*time.Second {
		t.Fatalf("delay = %s, want %s", delay, 4*time.Second)
	}
}

func TestReconnectDelayClampsAtConfiguredMaximum(t *testing.T) {
	delay, err := ReconnectDelay(6, time.Second, 10*time.Second)
	if err != nil {
		t.Fatalf("reconnect delay: %v", err)
	}

	if delay != 10*time.Second {
		t.Fatalf("delay = %s, want %s", delay, 10*time.Second)
	}
}

func TestReconnectDelayRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		minBackoff time.Duration
		maxBackoff time.Duration
	}{
		{name: "non-positive attempt", attempt: 0, minBackoff: time.Second, maxBackoff: 2 * time.Second},
		{name: "non-positive min", attempt: 1, minBackoff: 0, maxBackoff: 2 * time.Second},
		{name: "max below min", attempt: 1, minBackoff: 2 * time.Second, maxBackoff: time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ReconnectDelay(tc.attempt, tc.minBackoff, tc.maxBackoff); err == nil {
				t.Fatal("expected invalid reconnect delay input to fail")
			}
		})
	}
}
