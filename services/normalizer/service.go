package normalizer

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type Service struct {
	policy     ingestion.TimestampPolicy
	rawWriter  ingestion.RawEventWriter
	rawOptions ingestion.RawWriteOptions
}

type Option func(*Service)

type TradeInput struct {
	Metadata ingestion.TradeMetadata
	Message  ingestion.TradeMessage
	Raw      ingestion.RawWriteContext
}

type OrderBookInput struct {
	Metadata  ingestion.BookMetadata
	Message   ingestion.OrderBookMessage
	Sequencer *ingestion.OrderBookSequencer
	Raw       ingestion.RawWriteContext
}

type FeedHealthInput struct {
	Metadata ingestion.FeedHealthMetadata
	Message  ingestion.FeedHealthMessage
	Raw      ingestion.RawWriteContext
}

type FundingInput struct {
	Metadata ingestion.DerivativesMetadata
	Message  ingestion.FundingRateMessage
	Raw      ingestion.RawWriteContext
}

type MarkIndexInput struct {
	Metadata ingestion.DerivativesMetadata
	Message  ingestion.MarkIndexMessage
	Raw      ingestion.RawWriteContext
}

type OpenInterestInput struct {
	Metadata ingestion.DerivativesMetadata
	Message  ingestion.OpenInterestMessage
	Raw      ingestion.RawWriteContext
}

type LiquidationInput struct {
	Metadata ingestion.DerivativesMetadata
	Message  ingestion.LiquidationMessage
	Raw      ingestion.RawWriteContext
}

func WithRawEventWriter(writer ingestion.RawEventWriter, options ingestion.RawWriteOptions) Option {
	return func(service *Service) {
		service.rawWriter = writer
		service.rawOptions = options
	}
}

func NewService(policy ingestion.TimestampPolicy, options ...Option) (*Service, error) {
	if policy.MaxExchangeRecvSkew <= 0 {
		return nil, fmt.Errorf("timestamp policy is required")
	}
	service := &Service{policy: policy, rawOptions: ingestion.DefaultRawWriteOptions()}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service, nil
}

func (s *Service) NormalizeTrade(input TradeInput) (ingestion.CanonicalTradeEvent, error) {
	if s == nil {
		return ingestion.CanonicalTradeEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeTradeMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalTradeEvent{}, err
	}
	if err := s.appendTradeRaw(input, event); err != nil {
		return ingestion.CanonicalTradeEvent{}, err
	}
	return event, nil
}

func (s *Service) NormalizeOrderBook(input OrderBookInput) (ingestion.OrderBookNormalizationResult, error) {
	if s == nil {
		return ingestion.OrderBookNormalizationResult{}, fmt.Errorf("normalizer service is required")
	}
	result, err := ingestion.NormalizeOrderBookMessage(input.Metadata, input.Message, input.Sequencer, s.policy)
	if err != nil {
		return ingestion.OrderBookNormalizationResult{}, err
	}
	if err := s.appendOrderBookRaw(input, result); err != nil {
		return ingestion.OrderBookNormalizationResult{}, err
	}
	return result, nil
}

func (s *Service) NormalizeFeedHealth(input FeedHealthInput) (ingestion.CanonicalFeedHealthEvent, error) {
	if s == nil {
		return ingestion.CanonicalFeedHealthEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeFeedHealthMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalFeedHealthEvent{}, err
	}
	if err := s.appendFeedHealthRaw(input, event); err != nil {
		return ingestion.CanonicalFeedHealthEvent{}, err
	}
	return event, nil
}

func (s *Service) NormalizeFunding(input FundingInput) (ingestion.CanonicalFundingRateEvent, error) {
	if s == nil {
		return ingestion.CanonicalFundingRateEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeFundingMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalFundingRateEvent{}, err
	}
	if err := s.appendFundingRaw(input, event); err != nil {
		return ingestion.CanonicalFundingRateEvent{}, err
	}
	return event, nil
}

func (s *Service) NormalizeMarkIndex(input MarkIndexInput) (ingestion.CanonicalMarkIndexEvent, error) {
	if s == nil {
		return ingestion.CanonicalMarkIndexEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeMarkIndexMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalMarkIndexEvent{}, err
	}
	if err := s.appendMarkIndexRaw(input, event); err != nil {
		return ingestion.CanonicalMarkIndexEvent{}, err
	}
	return event, nil
}

func (s *Service) NormalizeOpenInterest(input OpenInterestInput) (ingestion.CanonicalOpenInterestSnapshotEvent, error) {
	if s == nil {
		return ingestion.CanonicalOpenInterestSnapshotEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeOpenInterestMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalOpenInterestSnapshotEvent{}, err
	}
	if err := s.appendOpenInterestRaw(input, event); err != nil {
		return ingestion.CanonicalOpenInterestSnapshotEvent{}, err
	}
	return event, nil
}

func (s *Service) NormalizeLiquidation(input LiquidationInput) (ingestion.CanonicalLiquidationPrintEvent, error) {
	if s == nil {
		return ingestion.CanonicalLiquidationPrintEvent{}, fmt.Errorf("normalizer service is required")
	}
	event, err := ingestion.NormalizeLiquidationMessage(input.Metadata, input.Message, s.policy)
	if err != nil {
		return ingestion.CanonicalLiquidationPrintEvent{}, err
	}
	if err := s.appendLiquidationRaw(input, event); err != nil {
		return ingestion.CanonicalLiquidationPrintEvent{}, err
	}
	return event, nil
}

func (s *Service) appendTradeRaw(input TradeInput, event ingestion.CanonicalTradeEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromTrade(event, input.Metadata, input.Message, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}

func (s *Service) appendOrderBookRaw(input OrderBookInput, result ingestion.OrderBookNormalizationResult) error {
	if s.rawWriter == nil {
		return nil
	}
	if result.OrderBookEvent != nil {
		entry, err := ingestion.BuildRawAppendEntryFromOrderBook(*result.OrderBookEvent, input.Metadata, input.Message, input.Raw, s.rawOptions)
		if err != nil {
			return err
		}
		if err := s.rawWriter.Append(entry); err != nil {
			return err
		}
	}
	if result.FeedHealthEvent != nil {
		entry, err := ingestion.BuildRawAppendEntryFromFeedHealth(*result.FeedHealthEvent, input.Metadata.SourceSymbol, string(ingestion.StreamOrderBook), input.Raw, s.rawOptions)
		if err != nil {
			return err
		}
		if err := s.rawWriter.Append(entry); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) appendFeedHealthRaw(input FeedHealthInput, event ingestion.CanonicalFeedHealthEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromFeedHealth(event, input.Metadata.SourceSymbol, ingestion.RawStreamFamilyFeedHealth, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}

func (s *Service) appendFundingRaw(input FundingInput, event ingestion.CanonicalFundingRateEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromFundingRate(event, input.Metadata, input.Message, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}

func (s *Service) appendMarkIndexRaw(input MarkIndexInput, event ingestion.CanonicalMarkIndexEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromMarkIndex(event, input.Metadata, input.Message, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}

func (s *Service) appendOpenInterestRaw(input OpenInterestInput, event ingestion.CanonicalOpenInterestSnapshotEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromOpenInterest(event, input.Metadata, input.Message, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}

func (s *Service) appendLiquidationRaw(input LiquidationInput, event ingestion.CanonicalLiquidationPrintEvent) error {
	if s.rawWriter == nil {
		return nil
	}
	entry, err := ingestion.BuildRawAppendEntryFromLiquidation(event, input.Metadata, input.Message, input.Raw, s.rawOptions)
	if err != nil {
		return err
	}
	return s.rawWriter.Append(entry)
}
