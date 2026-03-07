package ingestion

import (
	"fmt"
	"time"
)

func ReconnectDelay(attempt int, minBackoff, maxBackoff time.Duration) (time.Duration, error) {
	if attempt <= 0 {
		return 0, fmt.Errorf("attempt must be positive")
	}
	if minBackoff <= 0 || maxBackoff <= 0 {
		return 0, fmt.Errorf("backoff thresholds must be positive")
	}
	if maxBackoff < minBackoff {
		return 0, fmt.Errorf("max backoff must be greater than or equal to min backoff")
	}

	delay := minBackoff
	for i := 1; i < attempt; i++ {
		if delay >= maxBackoff {
			return maxBackoff, nil
		}
		if delay > maxBackoff/2 {
			delay = maxBackoff
			continue
		}
		delay *= 2
	}

	if delay > maxBackoff {
		return maxBackoff, nil
	}
	return delay, nil
}
