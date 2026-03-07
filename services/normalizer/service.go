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
