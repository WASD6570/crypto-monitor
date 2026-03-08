package slowcontext

import "fmt"

type Adapter interface {
	SourceFamily() SourceFamily
	ParsePoll(raw []byte, ingestTs string) (PollResult, error)
}

type Service struct {
	adapters map[SourceFamily]Adapter
	store    Store
}

func NewService(adapters ...Adapter) (*Service, error) {
	registered := adapters
	if len(registered) == 0 {
		registered = []Adapter{
			NewCMEAdapter(),
			NewETFAdapter(),
		}
	}

	service := &Service{adapters: make(map[SourceFamily]Adapter, len(registered)), store: NewInMemoryStore()}
	for _, adapter := range registered {
		if adapter == nil {
			return nil, fmt.Errorf("adapter is required")
		}
		sourceFamily := adapter.SourceFamily()
		if err := validateSourceFamily(sourceFamily); err != nil {
			return nil, err
		}
		if _, exists := service.adapters[sourceFamily]; exists {
			return nil, fmt.Errorf("duplicate adapter for source family %q", sourceFamily)
		}
		service.adapters[sourceFamily] = adapter
	}

	return service, nil
}

func (s *Service) AdapterFor(sourceFamily SourceFamily) (Adapter, error) {
	if s == nil {
		return nil, fmt.Errorf("slow context service is required")
	}
	if err := validateSourceFamily(sourceFamily); err != nil {
		return nil, err
	}
	adapter, ok := s.adapters[sourceFamily]
	if !ok {
		return nil, fmt.Errorf("adapter for source family %q is not registered", sourceFamily)
	}
	return adapter, nil
}

func (s *Service) ParsePoll(sourceFamily SourceFamily, raw []byte, ingestTs string) (PollResult, error) {
	adapter, err := s.AdapterFor(sourceFamily)
	if err != nil {
		return PollResult{}, err
	}
	return adapter.ParsePoll(raw, ingestTs)
}
