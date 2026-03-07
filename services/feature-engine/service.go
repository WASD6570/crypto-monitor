package featureengine

import (
	"fmt"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
)

type Service struct {
	config          features.CompositeConfig
	bucketConfig    *features.BucketConfig
	bucketProcessor *features.WorldUSABucketProcessor
}

type ServiceOption func(*Service) error

func WithBucketConfig(config features.BucketConfig) ServiceOption {
	return func(service *Service) error {
		processor, err := features.NewWorldUSABucketProcessor(config)
		if err != nil {
			return err
		}
		service.bucketConfig = &config
		service.bucketProcessor = processor
		return nil
	}
}

func NewService(config features.CompositeConfig, options ...ServiceOption) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	service := &Service{config: config}
	for _, option := range options {
		if err := option(service); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func (s *Service) BuildCompositeSnapshot(group features.CompositeGroup, symbol string, bucketTs time.Time, inputs []features.ContributorInput) (features.CompositeSnapshot, error) {
	if s == nil {
		return features.CompositeSnapshot{}, fmt.Errorf("feature engine service is required")
	}
	return features.BuildCompositeSnapshot(s.config, group, symbol, bucketTs, inputs)
}

func (s *Service) BuildWorldUSASnapshots(symbol string, bucketTs time.Time, inputs []features.ContributorInput) ([]features.CompositeSnapshot, error) {
	world, err := s.BuildCompositeSnapshot(features.CompositeGroupWorld, symbol, bucketTs, inputs)
	if err != nil {
		return nil, err
	}
	usa, err := s.BuildCompositeSnapshot(features.CompositeGroupUSA, symbol, bucketTs, inputs)
	if err != nil {
		return nil, err
	}
	return []features.CompositeSnapshot{world, usa}, nil
}

func (s *Service) ObserveWorldUSABucket(observation features.WorldUSAObservation) (features.ObservationResult, error) {
	if s == nil {
		return features.ObservationResult{}, fmt.Errorf("feature engine service is required")
	}
	if s.bucketProcessor == nil {
		return features.ObservationResult{}, fmt.Errorf("bucket processor is not configured")
	}
	return s.bucketProcessor.Observe(observation)
}

func (s *Service) AdvanceWorldUSABuckets(symbol string, now time.Time) ([]features.MarketQualityBucket, error) {
	if s == nil {
		return nil, fmt.Errorf("feature engine service is required")
	}
	if s.bucketProcessor == nil {
		return nil, fmt.Errorf("bucket processor is not configured")
	}
	return s.bucketProcessor.Advance(symbol, now), nil
}

func (s *Service) QueryCurrentState(query features.SymbolCurrentStateQuery) (features.MarketStateCurrentResponse, error) {
	if s == nil {
		return features.MarketStateCurrentResponse{}, fmt.Errorf("feature engine service is required")
	}
	return features.BuildMarketStateCurrentResponse(query)
}
