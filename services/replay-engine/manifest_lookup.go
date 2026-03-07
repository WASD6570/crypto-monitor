package replayengine

import (
	"fmt"

	"github.com/crypto-market-copilot/alerts/libs/go/ingestion"
)

type ManifestReader interface {
	ResolveRawPartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error)
}

type Resolver struct {
	reader ManifestReader
}

func NewResolver(reader ManifestReader) (*Resolver, error) {
	if reader == nil {
		return nil, fmt.Errorf("manifest reader is required")
	}
	return &Resolver{reader: reader}, nil
}

func (r *Resolver) ResolvePartitions(scope ingestion.RawPartitionLookupScope) ([]ingestion.RawPartitionManifestRecord, error) {
	if r == nil {
		return nil, fmt.Errorf("replay resolver is required")
	}
	return r.reader.ResolveRawPartitions(scope)
}
