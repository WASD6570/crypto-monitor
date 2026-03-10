package venuebinance

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

const usdmOpenInterestFeedHealthRecordPrefix = "runtime:binance-usdm-open-interest:"

type ParsedUSDMOpenInterest struct {
	SourceSymbol string
	Message      ingestion.OpenInterestMessage
}

type USDMOpenInterestPollRequest struct {
	Symbol       string
	SourceSymbol string
	Path         string
}

type USDMOpenInterestPollPlan struct {
	Request USDMOpenInterestPollRequest
	Allowed bool
	RetryAt time.Time
	DueAt   time.Time
}

type USDMOpenInterestPollerState struct {
	Symbol           string
	SourceSymbol     string
	QuoteCurrency    string
	LastAttemptAt    time.Time
	LastSuccessAt    time.Time
	NextPollAt       time.Time
	RateLimitUntil   time.Time
	RecentAttemptIDs int
	RecentAttempts   []time.Time
}

type USDMOpenInterestPoller struct {
	runtime             *Runtime
	symbols             []usdmSymbolBinding
	pollInterval        time.Duration
	pollsPerMinuteLimit int
	localClockOffset    time.Duration
	states              map[string]*usdmOpenInterestSymbolState
}

type usdmOpenInterestPayload struct {
	SourceSymbol string `json:"symbol"`
	OpenInterest string `json:"openInterest"`
	TimeMs       int64  `json:"time"`
}

type usdmOpenInterestSymbolState struct {
	binding        usdmSymbolBinding
	lastAttemptAt  time.Time
	lastSuccessAt  time.Time
	nextPollAt     time.Time
	rateLimitUntil time.Time
	recentAttempts []time.Time
}

func ParseUSDMOpenInterest(raw []byte, recvTime time.Time) (ParsedUSDMOpenInterest, error) {
	if recvTime.IsZero() {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("recv time is required")
	}

	var payload usdmOpenInterestPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("decode binance usdm open interest payload: %w", err)
	}
	if payload.SourceSymbol == "" {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("source symbol is required")
	}
	if payload.OpenInterest == "" {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("open interest is required")
	}

	exchangeTimestamp := ""
	if payload.TimeMs > 0 {
		formatted, err := formatUnixMilliTimestamp(payload.TimeMs)
		if err != nil {
			return ParsedUSDMOpenInterest{}, err
		}
		exchangeTimestamp = formatted
	}

	return ParsedUSDMOpenInterest{
		SourceSymbol: payload.SourceSymbol,
		Message: ingestion.OpenInterestMessage{
			Type:         "open-interest",
			OpenInterest: payload.OpenInterest,
			ExchangeTs:   exchangeTimestamp,
			RecvTs:       recvTime.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

func NewUSDMOpenInterestPoller(runtime *Runtime) (*USDMOpenInterestPoller, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime is required")
	}
	config, err := usdmOpenInterestConfig(runtime.config)
	if err != nil {
		return nil, err
	}
	bindings, err := usdmSymbolBindings(config.Symbols)
	if err != nil {
		return nil, err
	}
	states := make(map[string]*usdmOpenInterestSymbolState, len(bindings))
	for _, binding := range bindings {
		binding := binding
		states[binding.SourceSymbol] = &usdmOpenInterestSymbolState{binding: binding}
	}
	return &USDMOpenInterestPoller{
		runtime:             runtime,
		symbols:             bindings,
		pollInterval:        config.OpenInterestPollInterval,
		pollsPerMinuteLimit: config.OpenInterestPollsPerMinuteLimit,
		states:              states,
	}, nil
}

func (p *USDMOpenInterestPoller) State() []USDMOpenInterestPollerState {
	if p == nil {
		return nil
	}
	states := make([]USDMOpenInterestPollerState, 0, len(p.symbols))
	for _, binding := range p.symbols {
		state := p.states[binding.SourceSymbol]
		states = append(states, USDMOpenInterestPollerState{
			Symbol:           binding.Symbol,
			SourceSymbol:     binding.SourceSymbol,
			QuoteCurrency:    binding.QuoteCurrency,
			LastAttemptAt:    state.lastAttemptAt,
			LastSuccessAt:    state.lastSuccessAt,
			NextPollAt:       state.nextPollAt,
			RateLimitUntil:   state.rateLimitUntil,
			RecentAttemptIDs: len(state.recentAttempts),
			RecentAttempts:   append([]time.Time(nil), state.recentAttempts...),
		})
	}
	return states
}

func (p *USDMOpenInterestPoller) UpdateClockOffset(offset time.Duration) {
	if p == nil {
		return
	}
	p.localClockOffset = offset
}

func (p *USDMOpenInterestPoller) DuePolls(now time.Time) ([]USDMOpenInterestPollPlan, error) {
	if p == nil {
		return nil, fmt.Errorf("open interest poller is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("current time is required")
	}
	plans := make([]USDMOpenInterestPollPlan, 0, len(p.symbols))
	for _, binding := range p.symbols {
		state := p.states[binding.SourceSymbol]
		if err := pruneOpenInterestAttempts(state, now); err != nil {
			return nil, err
		}
		if !state.nextPollAt.IsZero() && now.Before(state.nextPollAt) {
			continue
		}
		plan := USDMOpenInterestPollPlan{
			Request: USDMOpenInterestPollRequest{
				Symbol:       binding.Symbol,
				SourceSymbol: binding.SourceSymbol,
				Path:         buildUSDMOpenInterestPath(binding.SourceSymbol),
			},
			Allowed: true,
			DueAt:   state.nextPollAt,
		}
		if len(state.recentAttempts) >= p.pollsPerMinuteLimit {
			plan.Allowed = false
			plan.RetryAt = oldestOpenInterestAttempt(state.recentAttempts).Add(time.Minute)
		}
		plans = append(plans, plan)
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Request.SourceSymbol < plans[j].Request.SourceSymbol
	})
	return plans, nil
}

func (p *USDMOpenInterestPoller) BeginPoll(sourceSymbol string, now time.Time) (USDMOpenInterestPollRequest, error) {
	if p == nil {
		return USDMOpenInterestPollRequest{}, fmt.Errorf("open interest poller is required")
	}
	if now.IsZero() {
		return USDMOpenInterestPollRequest{}, fmt.Errorf("poll time is required")
	}
	state, ok := p.states[sourceSymbol]
	if !ok {
		return USDMOpenInterestPollRequest{}, fmt.Errorf("unsupported open interest source symbol %q", sourceSymbol)
	}
	if err := pruneOpenInterestAttempts(state, now); err != nil {
		return USDMOpenInterestPollRequest{}, err
	}
	if !state.nextPollAt.IsZero() && now.Before(state.nextPollAt) {
		return USDMOpenInterestPollRequest{}, fmt.Errorf("open interest poll for %s is not due until %s", sourceSymbol, state.nextPollAt.UTC().Format(time.RFC3339Nano))
	}
	if len(state.recentAttempts) >= p.pollsPerMinuteLimit {
		state.rateLimitUntil = oldestOpenInterestAttempt(state.recentAttempts).Add(time.Minute)
		return USDMOpenInterestPollRequest{}, fmt.Errorf("open interest poll rate limit reached for %s until %s", sourceSymbol, state.rateLimitUntil.UTC().Format(time.RFC3339Nano))
	}
	state.lastAttemptAt = now
	state.nextPollAt = now.Add(p.pollInterval)
	state.recentAttempts = append(state.recentAttempts, now)
	if !state.rateLimitUntil.IsZero() && !now.Before(state.rateLimitUntil) {
		state.rateLimitUntil = time.Time{}
	}
	return USDMOpenInterestPollRequest{
		Symbol:       state.binding.Symbol,
		SourceSymbol: state.binding.SourceSymbol,
		Path:         buildUSDMOpenInterestPath(state.binding.SourceSymbol),
	}, nil
}

func (p *USDMOpenInterestPoller) AcceptSnapshot(raw []byte, now time.Time) (ParsedUSDMOpenInterest, error) {
	if p == nil {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("open interest poller is required")
	}
	parsed, err := ParseUSDMOpenInterest(raw, now)
	if err != nil {
		return ParsedUSDMOpenInterest{}, err
	}
	state, ok := p.states[parsed.SourceSymbol]
	if !ok {
		return ParsedUSDMOpenInterest{}, fmt.Errorf("unsupported open interest source symbol %q", parsed.SourceSymbol)
	}
	state.lastSuccessAt = now
	if !state.rateLimitUntil.IsZero() && !now.Before(state.rateLimitUntil) {
		state.rateLimitUntil = time.Time{}
	}
	return parsed, nil
}

func (p *USDMOpenInterestPoller) FeedHealthInputs(now time.Time) ([]USDMFeedHealthInput, error) {
	if p == nil {
		return nil, fmt.Errorf("open interest poller is required")
	}
	if now.IsZero() {
		return nil, fmt.Errorf("feed health time is required")
	}
	inputs := make([]USDMFeedHealthInput, 0, len(p.symbols))
	for _, binding := range p.symbols {
		status, err := p.HealthStatus(binding.SourceSymbol, now)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, USDMFeedHealthInput{
			Metadata: ingestion.FeedHealthMetadata{
				Symbol:        binding.Symbol,
				SourceSymbol:  binding.SourceSymbol,
				QuoteCurrency: binding.QuoteCurrency,
				Venue:         ingestion.VenueBinance,
				MarketType:    "perpetual",
			},
			Message: ingestion.FeedHealthMessage{
				ExchangeTs:     now.UTC().Format(time.RFC3339Nano),
				RecvTs:         now.UTC().Format(time.RFC3339Nano),
				SourceRecordID: usdmOpenInterestFeedHealthRecordPrefix + binding.SourceSymbol,
				Status:         status,
			},
		})
	}
	return inputs, nil
}

func (p *USDMOpenInterestPoller) HealthStatus(sourceSymbol string, now time.Time) (ingestion.FeedHealthStatus, error) {
	if p == nil {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("open interest poller is required")
	}
	if now.IsZero() {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("current time is required")
	}
	state, ok := p.states[sourceSymbol]
	if !ok {
		return ingestion.FeedHealthStatus{}, fmt.Errorf("unsupported open interest source symbol %q", sourceSymbol)
	}
	if err := pruneOpenInterestAttempts(state, now); err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	connectionState := ingestion.ConnectionConnected
	if state.lastSuccessAt.IsZero() {
		connectionState = ingestion.ConnectionConnecting
	}
	status, err := p.runtime.EvaluateAdapterInput(AdapterHealthInput{
		ConnectionState:  connectionState,
		Now:              now,
		LastMessageAt:    state.lastSuccessAt,
		LocalClockOffset: p.localClockOffset,
	})
	if err != nil {
		return ingestion.FeedHealthStatus{}, err
	}
	if !state.rateLimitUntil.IsZero() && now.Before(state.rateLimitUntil) {
		status.Reasons = appendReason(status.Reasons, ingestion.ReasonRateLimit)
		if status.State == ingestion.FeedHealthHealthy {
			status.State = ingestion.FeedHealthDegraded
		}
	}
	return status, nil
}

func usdmOpenInterestConfig(config ingestion.VenueRuntimeConfig) (ingestion.VenueRuntimeConfig, error) {
	filteredStreams := make([]ingestion.StreamDefinition, 0, len(config.Adapter.Streams))
	for _, stream := range config.Adapter.Streams {
		if stream.MarketType != "perpetual" || stream.Kind != ingestion.StreamOpenInterest {
			continue
		}
		filteredStreams = append(filteredStreams, ingestion.StreamDefinition{Kind: stream.Kind, MarketType: stream.MarketType, SnapshotRequired: false})
	}
	if len(filteredStreams) == 0 {
		return ingestion.VenueRuntimeConfig{}, fmt.Errorf("binance usdm open interest poller requires perpetual open-interest stream")
	}
	if config.OpenInterestPollInterval <= 0 {
		return ingestion.VenueRuntimeConfig{}, fmt.Errorf("binance usdm open interest poll interval is required")
	}
	if config.OpenInterestPollsPerMinuteLimit <= 0 {
		return ingestion.VenueRuntimeConfig{}, fmt.Errorf("binance usdm open interest polls per-minute limit is required")
	}
	filtered := config
	filtered.Adapter = config.Adapter
	filtered.Adapter.Streams = filteredStreams
	filtered.SnapshotRefreshRequired = false
	filtered.SnapshotRefreshInterval = 0
	return filtered, nil
}

func buildUSDMOpenInterestPath(sourceSymbol string) string {
	return fmt.Sprintf("/fapi/v1/openInterest?symbol=%s", sourceSymbol)
}

func pruneOpenInterestAttempts(state *usdmOpenInterestSymbolState, now time.Time) error {
	if state == nil {
		return fmt.Errorf("open interest state is required")
	}
	windowStart := now.Add(-time.Minute)
	pruned := state.recentAttempts[:0]
	for _, attempt := range state.recentAttempts {
		if attempt.IsZero() {
			return fmt.Errorf("open interest attempt time is required")
		}
		if attempt.After(now) {
			return fmt.Errorf("open interest attempt cannot be in the future")
		}
		if attempt.Before(windowStart) {
			continue
		}
		pruned = append(pruned, attempt)
	}
	state.recentAttempts = append([]time.Time(nil), pruned...)
	if !state.rateLimitUntil.IsZero() && !now.Before(state.rateLimitUntil) {
		state.rateLimitUntil = time.Time{}
	}
	return nil
}

func oldestOpenInterestAttempt(attempts []time.Time) time.Time {
	oldest := attempts[0]
	for _, attempt := range attempts[1:] {
		if attempt.Before(oldest) {
			oldest = attempt
		}
	}
	return oldest
}

func appendReason(reasons []ingestion.DegradationReason, reason ingestion.DegradationReason) []ingestion.DegradationReason {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	reasons = append(reasons, reason)
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i] < reasons[j]
	})
	return reasons
}
