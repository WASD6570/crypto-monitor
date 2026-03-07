package normalizer

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type Service struct {
	policy ingestion.TimestampPolicy
}

type TradeInput struct {
	Metadata ingestion.TradeMetadata
	Message  ingestion.TradeMessage
}

type OrderBookInput struct {
	Metadata  ingestion.BookMetadata
	Message   ingestion.OrderBookMessage
	Sequencer *ingestion.OrderBookSequencer
}

type FeedHealthInput struct {
	Metadata ingestion.FeedHealthMetadata
	Message  ingestion.FeedHealthMessage
}

func NewService(policy ingestion.TimestampPolicy) (*Service, error) {
	if policy.MaxExchangeRecvSkew <= 0 {
		return nil, fmt.Errorf("timestamp policy is required")
	}
	return &Service{policy: policy}, nil
}

func (s *Service) NormalizeTrade(input TradeInput) (ingestion.CanonicalTradeEvent, error) {
	if s == nil {
		return ingestion.CanonicalTradeEvent{}, fmt.Errorf("normalizer service is required")
	}
	return ingestion.NormalizeTradeMessage(input.Metadata, input.Message, s.policy)
}

func (s *Service) NormalizeOrderBook(input OrderBookInput) (ingestion.OrderBookNormalizationResult, error) {
	if s == nil {
		return ingestion.OrderBookNormalizationResult{}, fmt.Errorf("normalizer service is required")
	}
	return ingestion.NormalizeOrderBookMessage(input.Metadata, input.Message, input.Sequencer, s.policy)
}

func (s *Service) NormalizeFeedHealth(input FeedHealthInput) (ingestion.CanonicalFeedHealthEvent, error) {
	if s == nil {
		return ingestion.CanonicalFeedHealthEvent{}, fmt.Errorf("normalizer service is required")
	}
	return ingestion.NormalizeFeedHealthMessage(input.Metadata, input.Message, s.policy)
}
