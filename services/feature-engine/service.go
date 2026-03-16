package featureengine

import (
	"fmt"
	"strings"
	"time"

	"github.com/crypto-market-copilot/alerts/libs/go/features"
	slowcontext "github.com/crypto-market-copilot/alerts/services/slow-context"
)

type Service struct {
	config          features.CompositeConfig
	bucketConfig    *features.BucketConfig
	bucketProcessor *features.WorldUSABucketProcessor
	slowContext     SlowContextReader
	usdmInfluence   features.USDMInfluenceConfig
}

type ServiceOption func(*Service) error

type SlowContextReader interface {
	QueryAsset(query slowcontext.AssetQuery) (slowcontext.AssetContextResponse, error)
}

type CurrentStateWithSlowContextResponse struct {
	CurrentState features.MarketStateCurrentResponse
	SlowContext  slowcontext.AssetContextResponse
}

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

func WithSlowContextReader(reader SlowContextReader) ServiceOption {
	return func(service *Service) error {
		service.slowContext = reader
		return nil
	}
}

func WithUSDMInfluenceConfig(config features.USDMInfluenceConfig) ServiceOption {
	return func(service *Service) error {
		if err := config.Validate(); err != nil {
			return err
		}
		service.usdmInfluence = config
		return nil
	}
}

func NewService(config features.CompositeConfig, options ...ServiceOption) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	service := &Service{config: config, usdmInfluence: features.DefaultUSDMInfluenceConfig()}
	for _, option := range options {
		if err := option(service); err != nil {
			return nil, err
		}
	}
	if err := service.usdmInfluence.Validate(); err != nil {
		return nil, err
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

func (s *Service) QueryCurrentStateWithSlowContext(query features.SymbolCurrentStateQuery, slowQuery slowcontext.AssetQuery) (CurrentStateWithSlowContextResponse, error) {
	currentState, err := s.QueryCurrentState(query)
	if err != nil {
		return CurrentStateWithSlowContextResponse{}, err
	}
	asset := slowQuery.Asset
	if asset == "" {
		asset = assetFromSymbol(query.Symbol)
	}
	if asset == "" {
		return CurrentStateWithSlowContextResponse{CurrentState: currentState, SlowContext: slowcontext.NewUnavailableAssetContextResponse("", slowQuery.MetricFamilies, slowQuery.Now, "missing slow-context asset")}, nil
	}
	if slowQuery.Now.IsZero() {
		slowQuery.Now = time.Now().UTC()
	}
	slowQuery.Asset = asset
	if s.slowContext == nil {
		return CurrentStateWithSlowContextResponse{CurrentState: currentState, SlowContext: slowcontext.NewUnavailableAssetContextResponse(asset, slowQuery.MetricFamilies, slowQuery.Now, "slow context reader unavailable")}, nil
	}
	slowResponse, err := s.slowContext.QueryAsset(slowQuery)
	if err != nil {
		return CurrentStateWithSlowContextResponse{CurrentState: currentState, SlowContext: slowcontext.NewUnavailableAssetContextResponse(asset, slowQuery.MetricFamilies, slowQuery.Now, err.Error())}, nil
	}
	return CurrentStateWithSlowContextResponse{CurrentState: currentState, SlowContext: slowResponse}, nil
}

func assetFromSymbol(symbol string) string {
	if symbol == "" {
		return ""
	}
	parts := strings.Split(symbol, "-")
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}
