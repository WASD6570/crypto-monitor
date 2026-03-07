package venuebybit

import (
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type Runtime struct {
	config ingestion.VenueRuntimeConfig
}

type AdapterHealthInput struct {
	ConnectionState       ingestion.ConnectionState
	Now                   time.Time
	LastMessageAt         time.Time
	LastSnapshotAt        time.Time
	SequenceGapDetected   bool
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	ResyncCount           int
}

type AdapterLoopState struct {
	ConnectionState       ingestion.ConnectionState
	LastMessageAt         time.Time
	LastSnapshotAt        time.Time
	SequenceGapDetected   bool
	LocalClockOffset      time.Duration
	ConsecutiveReconnects int
	ResyncCount           int
}

func NewRuntime(config ingestion.VenueRuntimeConfig) (*Runtime, error) {
	if config.Venue != ingestion.VenueBybit {
		return nil, fmt.Errorf("bybit runtime requires %q venue config", ingestion.VenueBybit)
	}
	if err := config.Adapter.Validate(); err != nil {
		return nil, err
	}
	if config.ServicePath == "" {
		return nil, fmt.Errorf("service path is required")
	}
	return &Runtime{config: config}, nil
}

func (s *AdapterLoopState) SetConnectionState(state ingestion.ConnectionState) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if err := validateConnectionState(state); err != nil {
		return err
	}
	s.ConnectionState = state
	return nil
}

func (s *AdapterLoopState) RecordMessage(at time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if at.IsZero() {
		return fmt.Errorf("message time is required")
	}
	s.LastMessageAt = at
	return nil
}

func (s *AdapterLoopState) RecordSnapshot(at time.Time) error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	if at.IsZero() {
		return fmt.Errorf("snapshot time is required")
	}
	s.LastSnapshotAt = at
	return nil
}

func (s *AdapterLoopState) MarkSequenceGap() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.SequenceGapDetected = true
	return nil
}

func (s *AdapterLoopState) IncrementReconnectCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ConsecutiveReconnects++
	return nil
}

func (s *AdapterLoopState) IncrementResyncCount() error {
	if s == nil {
		return fmt.Errorf("loop state is required")
	}
	s.ResyncCount++
	return nil
}

func (s AdapterLoopState) HealthInput(now time.Time) AdapterHealthInput {
	return AdapterHealthInput{
		ConnectionState:       s.ConnectionState,
		Now:                   now,
		LastMessageAt:         s.LastMessageAt,
		LastSnapshotAt:        s.LastSnapshotAt,
		SequenceGapDetected:   s.SequenceGapDetected,
		LocalClockOffset:      s.LocalClockOffset,
		ConsecutiveReconnects: s.ConsecutiveReconnects,
		ResyncCount:           s.ResyncCount,
	}
}

func (r *Runtime) ReconnectDelay(attempt int) (time.Duration, error) {
	if r == nil {
		return 0, fmt.Errorf("runtime is required")
	}
	return ingestion.ReconnectDelay(attempt, r.config.Adapter.ReconnectBackoffMin, r.config.Adapter.ReconnectBackoffMax)
}

func (r *Runtime) BuildHealthSnapshot(input AdapterHealthInput) (ingestion.HealthSnapshot, error) {
	if r == nil {
		return ingestion.HealthSnapshot{}, fmt.Errorf("runtime is required")
	}
	if err := validateConnectionState(input.ConnectionState); err != nil {
		return ingestion.HealthSnapshot{}, err
	}
	if input.Now.IsZero() {
		return ingestion.HealthSnapshot{}, fmt.Errorf("current time is required")
	}
	if !input.LastMessageAt.IsZero() && input.LastMessageAt.After(input.Now) {
		return ingestion.HealthSnapshot{}, fmt.Errorf("last message time cannot be in the future")
	}
	if !input.LastSnapshotAt.IsZero() && input.LastSnapshotAt.After(input.Now) {
		return ingestion.HealthSnapshot{}, fmt.Errorf("last snapshot time cannot be in the future")
	}

	return ingestion.HealthSnapshot{
		ConnectionState:       input.ConnectionState,
		Now:                   input.Now,
		LastMessageAt:         input.LastMessageAt,
		LastSnapshotAt:        input.LastSnapshotAt,
		SequenceGapDetected:   input.SequenceGapDetected,
		ConsecutiveReconnects: input.ConsecutiveReconnects,
		ResyncCount:           input.ResyncCount,
		LocalClockOffset:      input.LocalClockOffset,
	}, nil
}

func (r *Runtime) EvaluateAdapterInput(input AdapterHealthInput) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("runtime is required")
	}
	snapshot, err := r.BuildHealthSnapshot(input)
	if err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	return ingestion.EvaluateHealth(r.config.Adapter, snapshot)
}

func (r *Runtime) EvaluateLoopState(state AdapterLoopState, now time.Time) (ingestion.FeedHealthStatus, error) {
	if r == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("runtime is required")
	}
	return r.EvaluateAdapterInput(state.HealthInput(now))
}

func validateConnectionState(state ingestion.ConnectionState) error {
	switch state {
	case ingestion.ConnectionDisconnected, ingestion.ConnectionConnecting, ingestion.ConnectionConnected, ingestion.ConnectionReconnecting, ingestion.ConnectionResyncing:
		return nil
	default:
		return fmt.Errorf("unsupported connection state %q", state)
	}
}
