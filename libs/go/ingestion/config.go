package ingestion

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type EnvironmentConfig struct {
	SchemaVersion     string                       `json:"schemaVersion"`
	Environment       string                       `json:"environment"`
	Symbols           []string                     `json:"symbols"`
	NormalizerHandoff NormalizerHandoffConfig      `json:"normalizerHandoff"`
	Venues            map[Venue]VenueRuntimeSource `json:"venues"`
}

type NormalizerHandoffConfig struct {
	Service                  string `json:"service"`
	PreserveExchangeTs       bool   `json:"preserveExchangeTs"`
	PreserveRecvTs           bool   `json:"preserveRecvTs"`
	PropagateDegradedReasons bool   `json:"propagateDegradedReasons"`
}

type VenueRuntimeSource struct {
	ServicePath           string                `json:"servicePath"`
	Websocket             WebsocketConfig       `json:"websocket"`
	Rest                  RestConfig            `json:"rest"`
	Health                HealthThresholds      `json:"health"`
	SnapshotRefreshPolicy SnapshotRefreshPolicy `json:"snapshotRefreshPolicy"`
	Streams               []StreamDefinition    `json:"streams"`
}

type WebsocketConfig struct {
	HeartbeatTimeoutMs     int  `json:"heartbeatTimeoutMs"`
	ReconnectBackoffMinMs  int  `json:"reconnectBackoffMinMs"`
	ReconnectBackoffMaxMs  int  `json:"reconnectBackoffMaxMs"`
	ReconnectLoopThreshold int  `json:"reconnectLoopThreshold"`
	ConnectsPerMinuteLimit int  `json:"connectsPerMinuteLimit"`
	ResubscribeOnReconnect bool `json:"resubscribeOnReconnect"`
}

type RestConfig struct {
	SnapshotRecoveryPerMinuteLimit  int `json:"snapshotRecoveryPerMinuteLimit"`
	SnapshotCooldownMs              int `json:"snapshotCooldownMs"`
	OpenInterestPollIntervalMs      int `json:"openInterestPollIntervalMs"`
	OpenInterestPollsPerMinuteLimit int `json:"openInterestPollsPerMinuteLimit"`
}

type HealthThresholds struct {
	MessageStaleAfterMs   int `json:"messageStaleAfterMs"`
	SnapshotStaleAfterMs  int `json:"snapshotStaleAfterMs"`
	ResyncLoopThreshold   int `json:"resyncLoopThreshold"`
	ClockOffsetWarningMs  int `json:"clockOffsetWarningMs"`
	ClockOffsetDegradedMs int `json:"clockOffsetDegradedMs"`
}

type SnapshotRefreshPolicy struct {
	Required          bool `json:"required"`
	RefreshIntervalMs int  `json:"refreshIntervalMs"`
}

type VenueRuntimeConfig struct {
	Venue                           Venue
	ServicePath                     string
	Symbols                         []string
	NormalizerHandoff               NormalizerHandoffConfig
	Adapter                         AdapterConfig
	HeartbeatTimeout                time.Duration
	ConnectsPerMinuteLimit          int
	ResubscribeOnReconnect          bool
	SnapshotRecoveryPerMinuteLimit  int
	SnapshotCooldown                time.Duration
	OpenInterestPollInterval        time.Duration
	OpenInterestPollsPerMinuteLimit int
	SnapshotRefreshRequired         bool
	SnapshotRefreshInterval         time.Duration
}

func LoadEnvironmentConfig(path string) (EnvironmentConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return EnvironmentConfig{}, err
	}

	var config EnvironmentConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		return EnvironmentConfig{}, err
	}

	if err := config.Validate(); err != nil {
		return EnvironmentConfig{}, err
	}

	return config, nil
}

func (c EnvironmentConfig) Validate() error {
	if c.SchemaVersion != "v1" {
		return fmt.Errorf("schemaVersion must be v1")
	}
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if len(c.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}
	if c.NormalizerHandoff.Service == "" {
		return fmt.Errorf("normalizer handoff service is required")
	}
	if !c.NormalizerHandoff.PreserveExchangeTs || !c.NormalizerHandoff.PreserveRecvTs || !c.NormalizerHandoff.PropagateDegradedReasons {
		return fmt.Errorf("normalizer handoff must preserve timestamps and degraded reasons")
	}
	if len(c.Venues) == 0 {
		return fmt.Errorf("at least one venue config is required")
	}

	for venue, venueConfig := range c.Venues {
		if _, err := venueConfig.RuntimeConfig(venue, c.Symbols, c.NormalizerHandoff); err != nil {
			return fmt.Errorf("venue %s: %w", venue, err)
		}
	}

	return nil
}

func (c EnvironmentConfig) RuntimeConfigFor(venue Venue) (VenueRuntimeConfig, error) {
	venueConfig, ok := c.Venues[venue]
	if !ok {
		return VenueRuntimeConfig{}, fmt.Errorf("unknown venue %q", venue)
	}
	return venueConfig.RuntimeConfig(venue, c.Symbols, c.NormalizerHandoff)
}

func (v VenueRuntimeSource) RuntimeConfig(venue Venue, symbols []string, handoff NormalizerHandoffConfig) (VenueRuntimeConfig, error) {
	if v.ServicePath == "" {
		return VenueRuntimeConfig{}, fmt.Errorf("service path is required")
	}
	if v.Websocket.HeartbeatTimeoutMs <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("heartbeat timeout must be positive")
	}
	if v.Websocket.ReconnectBackoffMinMs <= 0 || v.Websocket.ReconnectBackoffMaxMs <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("reconnect backoff thresholds must be positive")
	}
	if v.Websocket.ReconnectBackoffMaxMs < v.Websocket.ReconnectBackoffMinMs {
		return VenueRuntimeConfig{}, fmt.Errorf("reconnect backoff max must be greater than or equal to min")
	}
	if v.Websocket.ReconnectLoopThreshold <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("reconnect loop threshold must be positive")
	}
	if v.Websocket.ConnectsPerMinuteLimit <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("connects per minute limit must be positive")
	}
	if v.Rest.SnapshotRecoveryPerMinuteLimit < 0 || v.Rest.SnapshotCooldownMs < 0 || v.Rest.OpenInterestPollIntervalMs < 0 || v.Rest.OpenInterestPollsPerMinuteLimit < 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("rest limits must be non-negative")
	}
	if v.Health.MessageStaleAfterMs <= 0 || v.Health.SnapshotStaleAfterMs <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("health stale thresholds must be positive")
	}
	if v.Health.ResyncLoopThreshold <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("resync loop threshold must be positive")
	}
	if v.Health.ClockOffsetWarningMs <= 0 || v.Health.ClockOffsetDegradedMs <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("clock offset thresholds must be positive")
	}
	if v.Health.ClockOffsetDegradedMs < v.Health.ClockOffsetWarningMs {
		return VenueRuntimeConfig{}, fmt.Errorf("clock degraded threshold must be greater than or equal to warning threshold")
	}

	adapter := AdapterConfig{
		Venue:                  venue,
		Streams:                v.Streams,
		MessageStaleAfter:      time.Duration(v.Health.MessageStaleAfterMs) * time.Millisecond,
		SnapshotStaleAfter:     time.Duration(v.Health.SnapshotStaleAfterMs) * time.Millisecond,
		ReconnectBackoffMin:    time.Duration(v.Websocket.ReconnectBackoffMinMs) * time.Millisecond,
		ReconnectBackoffMax:    time.Duration(v.Websocket.ReconnectBackoffMaxMs) * time.Millisecond,
		ReconnectLoopThreshold: v.Websocket.ReconnectLoopThreshold,
		ResyncLoopThreshold:    v.Health.ResyncLoopThreshold,
		ClockOffsetWarning:     time.Duration(v.Health.ClockOffsetWarningMs) * time.Millisecond,
		ClockOffsetDegraded:    time.Duration(v.Health.ClockOffsetDegradedMs) * time.Millisecond,
	}
	if err := adapter.Validate(); err != nil {
		return VenueRuntimeConfig{}, err
	}

	if v.SnapshotRefreshPolicy.Required && v.SnapshotRefreshPolicy.RefreshIntervalMs <= 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("snapshot refresh interval must be positive when snapshots are required")
	}
	if !v.SnapshotRefreshPolicy.Required && v.SnapshotRefreshPolicy.RefreshIntervalMs != 0 {
		return VenueRuntimeConfig{}, fmt.Errorf("snapshot refresh interval must be zero when snapshots are not required")
	}

	hasSnapshotStream := false
	hasOpenInterestStream := false
	for _, stream := range v.Streams {
		if stream.SnapshotRequired {
			hasSnapshotStream = true
		}
		if stream.Kind == StreamOpenInterest {
			hasOpenInterestStream = true
		}
	}
	if v.SnapshotRefreshPolicy.Required && !hasSnapshotStream {
		return VenueRuntimeConfig{}, fmt.Errorf("snapshot refresh is required but no stream declares snapshotRequired")
	}
	if !v.SnapshotRefreshPolicy.Required && hasSnapshotStream {
		return VenueRuntimeConfig{}, fmt.Errorf("snapshot refresh is disabled but snapshotRequired streams are configured")
	}
	if hasOpenInterestStream {
		if v.Rest.OpenInterestPollIntervalMs <= 0 {
			return VenueRuntimeConfig{}, fmt.Errorf("open interest poll interval must be positive when open-interest stream is configured")
		}
		if v.Rest.OpenInterestPollsPerMinuteLimit <= 0 {
			return VenueRuntimeConfig{}, fmt.Errorf("open interest polls per-minute limit must be positive when open-interest stream is configured")
		}
	}

	return VenueRuntimeConfig{
		Venue:                           venue,
		ServicePath:                     v.ServicePath,
		Symbols:                         append([]string(nil), symbols...),
		NormalizerHandoff:               handoff,
		Adapter:                         adapter,
		HeartbeatTimeout:                time.Duration(v.Websocket.HeartbeatTimeoutMs) * time.Millisecond,
		ConnectsPerMinuteLimit:          v.Websocket.ConnectsPerMinuteLimit,
		ResubscribeOnReconnect:          v.Websocket.ResubscribeOnReconnect,
		SnapshotRecoveryPerMinuteLimit:  v.Rest.SnapshotRecoveryPerMinuteLimit,
		SnapshotCooldown:                time.Duration(v.Rest.SnapshotCooldownMs) * time.Millisecond,
		OpenInterestPollInterval:        time.Duration(v.Rest.OpenInterestPollIntervalMs) * time.Millisecond,
		OpenInterestPollsPerMinuteLimit: v.Rest.OpenInterestPollsPerMinuteLimit,
		SnapshotRefreshRequired:         v.SnapshotRefreshPolicy.Required,
		SnapshotRefreshInterval:         time.Duration(v.SnapshotRefreshPolicy.RefreshIntervalMs) * time.Millisecond,
	}, nil
}
