package ingestion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const (
	RawEventLogSchemaVersion = "v1"

	DefaultRawLateAfter       = 2 * time.Second
	DefaultRawHotRetention    = 30 * 24 * time.Hour
	DefaultRawColdRetention   = 365 * 24 * time.Hour
	RawStreamFamilyFeedHealth = "feed-health"
)

type RawBucketTimestampSource string

const (
	RawBucketTimestampSourceExchange RawBucketTimestampSource = "exchangeTs"
	RawBucketTimestampSourceRecv     RawBucketTimestampSource = "recvTs"
)

type RawStorageState string

const (
	RawStorageStateHot        RawStorageState = "hot"
	RawStorageStateCold       RawStorageState = "cold"
	RawStorageStateTransition RawStorageState = "transition"
)

type RawWriteOptions struct {
	NormalizerService string
	BuildVersion      string
	LateAfter         time.Duration
	HotRetention      time.Duration
	ColdRetention     time.Duration
}

func DefaultRawWriteOptions() RawWriteOptions {
	return RawWriteOptions{
		NormalizerService: "services/normalizer",
		BuildVersion:      "dev",
		LateAfter:         DefaultRawLateAfter,
		HotRetention:      DefaultRawHotRetention,
		ColdRetention:     DefaultRawColdRetention,
	}
}

func (o RawWriteOptions) normalized() RawWriteOptions {
	defaults := DefaultRawWriteOptions()
	if o.NormalizerService == "" {
		o.NormalizerService = defaults.NormalizerService
	}
	if o.BuildVersion == "" {
		o.BuildVersion = defaults.BuildVersion
	}
	if o.LateAfter <= 0 {
		o.LateAfter = defaults.LateAfter
	}
	if o.HotRetention <= 0 {
		o.HotRetention = defaults.HotRetention
	}
	if o.ColdRetention <= 0 {
		o.ColdRetention = defaults.ColdRetention
	}
	return o
}

type RawWriteContext struct {
	ConnectionRef   string
	SessionRef      string
	DegradedFeedRef string
}

func (c RawWriteContext) normalized() RawWriteContext {
	if c.ConnectionRef == "" {
		c.ConnectionRef = "unknown-connection"
	}
	if c.SessionRef == "" {
		c.SessionRef = "unknown-session"
	}
	return c
}

type RawDuplicateAudit struct {
	IdentityKey string `json:"identityKey"`
	Occurrence  int    `json:"occurrence"`
	Duplicate   bool   `json:"duplicate"`
}

type RawPartitionKey struct {
	UTCDate      string `json:"utcDate"`
	Symbol       string `json:"symbol"`
	Venue        Venue  `json:"venue"`
	StreamFamily string `json:"streamFamily,omitempty"`
}

func (k RawPartitionKey) String() string {
	if k.StreamFamily == "" {
		return fmt.Sprintf("%s/%s/%s", k.UTCDate, k.Symbol, k.Venue)
	}
	return fmt.Sprintf("%s/%s/%s/%s", k.UTCDate, k.Symbol, k.Venue, k.StreamFamily)
}

type RawAppendEntry struct {
	SchemaVersion              string                   `json:"schemaVersion"`
	CanonicalSchemaVersion     string                   `json:"canonicalSchemaVersion"`
	CanonicalEventType         string                   `json:"canonicalEventType"`
	CanonicalEventID           string                   `json:"canonicalEventId"`
	CanonicalPayload           json.RawMessage          `json:"canonicalPayload"`
	VenueMessageID             string                   `json:"venueMessageId,omitempty"`
	VenueSequence              int64                    `json:"venueSequence,omitempty"`
	StreamKey                  string                   `json:"streamKey"`
	Symbol                     string                   `json:"symbol"`
	Venue                      Venue                    `json:"venue"`
	MarketType                 string                   `json:"marketType"`
	StreamFamily               string                   `json:"streamFamily"`
	SourceInstrumentID         string                   `json:"sourceInstrumentId,omitempty"`
	ExchangeTs                 string                   `json:"exchangeTs"`
	RecvTs                     string                   `json:"recvTs"`
	BucketTimestamp            string                   `json:"bucketTimestamp"`
	BucketTimestampSource      RawBucketTimestampSource `json:"bucketTimestampSource"`
	TimestampDegradationReason TimestampFallbackReason  `json:"timestampDegradationReason,omitempty"`
	Late                       bool                     `json:"late"`
	NormalizerService          string                   `json:"normalizerService"`
	ConnectionRef              string                   `json:"connectionRef"`
	SessionRef                 string                   `json:"sessionRef"`
	BuildVersion               string                   `json:"buildVersion"`
	DegradedFeedRef            string                   `json:"degradedFeedRef,omitempty"`
	DuplicateAudit             RawDuplicateAudit        `json:"duplicateAudit"`
	PartitionKey               RawPartitionKey          `json:"partitionKey"`
}

type RawEventWriter interface {
	Append(entry RawAppendEntry) error
}

type InMemoryRawEventWriter struct {
	entries          []RawAppendEntry
	duplicateCounter map[string]int
}

func NewInMemoryRawEventWriter() *InMemoryRawEventWriter {
	return &InMemoryRawEventWriter{duplicateCounter: map[string]int{}}
}

func (w *InMemoryRawEventWriter) Append(entry RawAppendEntry) error {
	if w == nil {
		return fmt.Errorf("raw event writer is required")
	}
	if err := ValidateRawAppendEntry(entry); err != nil {
		return err
	}

	stored := cloneRawAppendEntry(entry)
	identityKey := stored.DuplicateAudit.IdentityKey
	if identityKey == "" {
		identityKey = rawDuplicateIdentityKey(stored)
	}
	occurrence := w.duplicateCounter[identityKey] + 1
	w.duplicateCounter[identityKey] = occurrence
	stored.DuplicateAudit = RawDuplicateAudit{
		IdentityKey: identityKey,
		Occurrence:  occurrence,
		Duplicate:   occurrence > 1,
	}
	stored.PartitionKey = RouteRawPartition(stored)
	w.entries = append(w.entries, stored)
	return nil
}

func (w *InMemoryRawEventWriter) Entries() []RawAppendEntry {
	if w == nil {
		return nil
	}
	entries := make([]RawAppendEntry, 0, len(w.entries))
	for _, entry := range w.entries {
		entries = append(entries, cloneRawAppendEntry(entry))
	}
	return entries
}

type RawPartitionManifestRecord struct {
	SchemaVersion         string          `json:"schemaVersion"`
	LogicalPartition      RawPartitionKey `json:"logicalPartition"`
	StorageState          RawStorageState `json:"storageState"`
	Location              string          `json:"location"`
	HotRetentionUntil     string          `json:"hotRetentionUntil"`
	ColdRetentionUntil    string          `json:"coldRetentionUntil"`
	EntryCount            int             `json:"entryCount"`
	FirstCanonicalEventID string          `json:"firstCanonicalEventId"`
	LastCanonicalEventID  string          `json:"lastCanonicalEventId"`
	ContinuityChecksum    string          `json:"continuityChecksum"`
}

type RawPartitionLookupScope struct {
	Symbol       string
	Venue        Venue
	StreamFamily string
	Start        time.Time
	End          time.Time
}

type RawPartitionResolver interface {
	ResolveRawPartitions(scope RawPartitionLookupScope) ([]RawPartitionManifestRecord, error)
}

type RawPartitionManifestIndex struct {
	records []RawPartitionManifestRecord
}

func NewRawPartitionManifestIndex(records []RawPartitionManifestRecord) (*RawPartitionManifestIndex, error) {
	index := &RawPartitionManifestIndex{records: make([]RawPartitionManifestRecord, 0, len(records))}
	seen := map[string]struct{}{}
	for _, record := range records {
		if err := ValidateRawPartitionManifestRecord(record); err != nil {
			return nil, err
		}
		key := record.LogicalPartition.String()
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate manifest record for logical partition %q", key)
		}
		seen[key] = struct{}{}
		index.records = append(index.records, record)
	}
	return index, nil
}

func (i *RawPartitionManifestIndex) ResolveRawPartitions(scope RawPartitionLookupScope) ([]RawPartitionManifestRecord, error) {
	if i == nil {
		return nil, fmt.Errorf("raw partition manifest index is required")
	}
	if scope.Symbol == "" || scope.Venue == "" {
		return nil, fmt.Errorf("symbol and venue are required")
	}
	if scope.Start.IsZero() || scope.End.IsZero() || scope.End.Before(scope.Start) {
		return nil, fmt.Errorf("valid lookup window is required")
	}

	startDate := scope.Start.UTC().Format("2006-01-02")
	endDate := scope.End.UTC().Format("2006-01-02")
	resolved := make([]RawPartitionManifestRecord, 0)
	for _, record := range i.records {
		if record.LogicalPartition.Symbol != scope.Symbol || record.LogicalPartition.Venue != scope.Venue {
			continue
		}
		if scope.StreamFamily != "" && record.LogicalPartition.StreamFamily != scope.StreamFamily {
			continue
		}
		if record.LogicalPartition.UTCDate < startDate || record.LogicalPartition.UTCDate > endDate {
			continue
		}
		resolved = append(resolved, record)
	}
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].LogicalPartition.String() < resolved[j].LogicalPartition.String()
	})
	return append([]RawPartitionManifestRecord(nil), resolved...), nil
}

func BuildRawAppendEntryFromTrade(event CanonicalTradeEvent, metadata TradeMetadata, message TradeMessage, context RawWriteContext, options RawWriteOptions) (RawAppendEntry, error) {
	if event.SchemaVersion != "v1" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical trade schema version %q", event.SchemaVersion)
	}
	if event.EventType != "market-trade" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical trade event type %q", event.EventType)
	}
	return newRawAppendEntry(rawAppendEntryArgs{
		canonicalSchemaVersion:  event.SchemaVersion,
		canonicalEventType:      event.EventType,
		canonicalPayload:        event,
		venueMessageID:          message.TradeID,
		streamKey:               string(StreamTrades),
		streamFamily:            string(StreamTrades),
		symbol:                  event.Symbol,
		venue:                   event.Venue,
		marketType:              event.MarketType,
		sourceInstrumentID:      metadata.SourceSymbol,
		exchangeTs:              event.ExchangeTs,
		recvTs:                  event.RecvTs,
		canonicalEventTime:      event.CanonicalEventTime,
		timestampFallbackReason: event.TimestampFallbackReason,
		timestampStatus:         event.TimestampStatus,
		degradedFeedRef:         context.DegradedFeedRef,
		context:                 context,
		options:                 options,
	})
}

func BuildRawAppendEntryFromOrderBook(event CanonicalOrderBookEvent, metadata BookMetadata, message OrderBookMessage, context RawWriteContext, options RawWriteOptions) (RawAppendEntry, error) {
	if event.SchemaVersion != "v1" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical order-book schema version %q", event.SchemaVersion)
	}
	if event.EventType != "order-book-top" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical order-book event type %q", event.EventType)
	}
	streamFamily := string(StreamOrderBook)
	if message.Type == string(BookUpdateTopOfBook) {
		streamFamily = string(StreamTopOfBook)
	}
	return newRawAppendEntry(rawAppendEntryArgs{
		canonicalSchemaVersion:  event.SchemaVersion,
		canonicalEventType:      event.EventType,
		canonicalPayload:        event,
		venueSequence:           message.Sequence,
		streamKey:               streamFamily,
		streamFamily:            streamFamily,
		symbol:                  event.Symbol,
		venue:                   event.Venue,
		marketType:              event.MarketType,
		sourceInstrumentID:      metadata.SourceSymbol,
		exchangeTs:              event.ExchangeTs,
		recvTs:                  event.RecvTs,
		canonicalEventTime:      event.CanonicalEventTime,
		timestampFallbackReason: event.TimestampFallbackReason,
		timestampStatus:         event.TimestampStatus,
		degradedFeedRef:         context.DegradedFeedRef,
		context:                 context,
		options:                 options,
	})
}

func BuildRawAppendEntryFromFeedHealth(event CanonicalFeedHealthEvent, sourceInstrumentID string, streamKey string, context RawWriteContext, options RawWriteOptions) (RawAppendEntry, error) {
	if event.SchemaVersion != "v1" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical feed-health schema version %q", event.SchemaVersion)
	}
	if event.EventType != "feed-health" {
		return RawAppendEntry{}, fmt.Errorf("unsupported canonical feed-health event type %q", event.EventType)
	}
	degradedRef := context.DegradedFeedRef
	if degradedRef == "" && len(event.DegradationReasons) > 0 {
		degradedRef = event.SourceRecordID
	}
	return newRawAppendEntry(rawAppendEntryArgs{
		canonicalSchemaVersion:  event.SchemaVersion,
		canonicalEventType:      event.EventType,
		canonicalPayload:        event,
		venueMessageID:          event.SourceRecordID,
		streamKey:               streamKey,
		streamFamily:            RawStreamFamilyFeedHealth,
		symbol:                  event.Symbol,
		venue:                   event.Venue,
		marketType:              event.MarketType,
		sourceInstrumentID:      sourceInstrumentID,
		exchangeTs:              event.ExchangeTs,
		recvTs:                  event.RecvTs,
		canonicalEventTime:      event.CanonicalEventTime,
		timestampFallbackReason: event.TimestampFallbackReason,
		timestampStatus:         event.TimestampStatus,
		degradedFeedRef:         degradedRef,
		context:                 context,
		options:                 options,
	})
}

func RouteRawPartition(entry RawAppendEntry) RawPartitionKey {
	bucketTime, err := time.Parse(time.RFC3339Nano, entry.BucketTimestamp)
	if err != nil {
		return RawPartitionKey{}
	}
	key := RawPartitionKey{
		UTCDate: bucketTime.UTC().Format("2006-01-02"),
		Symbol:  entry.Symbol,
		Venue:   entry.Venue,
	}
	if shouldPartitionByStreamFamily(entry.StreamFamily) {
		key.StreamFamily = entry.StreamFamily
	}
	return key
}

func ValidateRawAppendEntry(entry RawAppendEntry) error {
	if entry.SchemaVersion != RawEventLogSchemaVersion {
		return fmt.Errorf("raw append schemaVersion must be %q", RawEventLogSchemaVersion)
	}
	if entry.CanonicalSchemaVersion != "v1" {
		return fmt.Errorf("canonical schemaVersion must be v1")
	}
	if entry.CanonicalEventType == "" || entry.CanonicalEventID == "" {
		return fmt.Errorf("canonical event metadata is incomplete")
	}
	if len(entry.CanonicalPayload) == 0 {
		return fmt.Errorf("canonical payload is required")
	}
	if entry.StreamKey == "" || entry.Symbol == "" || entry.Venue == "" || entry.MarketType == "" || entry.StreamFamily == "" {
		return fmt.Errorf("raw append market provenance is incomplete")
	}
	if entry.RecvTs == "" || entry.BucketTimestamp == "" || entry.BucketTimestampSource == "" {
		return fmt.Errorf("raw append timestamp provenance is incomplete")
	}
	if entry.NormalizerService == "" || entry.ConnectionRef == "" || entry.SessionRef == "" || entry.BuildVersion == "" {
		return fmt.Errorf("raw append ingest provenance is incomplete")
	}
	if entry.PartitionKey.UTCDate == "" {
		return fmt.Errorf("raw append partition key is required")
	}
	return nil
}

func ValidateRawPartitionManifestRecord(record RawPartitionManifestRecord) error {
	if record.SchemaVersion != "v1" {
		return fmt.Errorf("manifest schemaVersion must be v1")
	}
	if record.LogicalPartition.UTCDate == "" || record.LogicalPartition.Symbol == "" || record.LogicalPartition.Venue == "" {
		return fmt.Errorf("manifest logical partition is incomplete")
	}
	if record.StorageState != RawStorageStateHot && record.StorageState != RawStorageStateCold && record.StorageState != RawStorageStateTransition {
		return fmt.Errorf("manifest storageState %q is unsupported", record.StorageState)
	}
	if record.Location == "" || record.EntryCount <= 0 {
		return fmt.Errorf("manifest location and entryCount are required")
	}
	if record.FirstCanonicalEventID == "" || record.LastCanonicalEventID == "" || record.ContinuityChecksum == "" {
		return fmt.Errorf("manifest continuity markers are required")
	}
	return nil
}

type rawAppendEntryArgs struct {
	canonicalSchemaVersion  string
	canonicalEventType      string
	canonicalPayload        any
	venueMessageID          string
	venueSequence           int64
	streamKey               string
	streamFamily            string
	symbol                  string
	venue                   Venue
	marketType              string
	sourceInstrumentID      string
	exchangeTs              string
	recvTs                  string
	canonicalEventTime      time.Time
	timestampFallbackReason TimestampFallbackReason
	timestampStatus         CanonicalTimestampStatus
	degradedFeedRef         string
	context                 RawWriteContext
	options                 RawWriteOptions
}

func newRawAppendEntry(args rawAppendEntryArgs) (RawAppendEntry, error) {
	payload, err := json.Marshal(args.canonicalPayload)
	if err != nil {
		return RawAppendEntry{}, fmt.Errorf("marshal canonical payload: %w", err)
	}
	bucketTimestamp := args.exchangeTs
	bucketSource := RawBucketTimestampSourceExchange
	if args.timestampStatus == TimestampStatusDegraded || args.timestampFallbackReason != TimestampReasonNone || bucketTimestamp == "" {
		bucketTimestamp = args.recvTs
		bucketSource = RawBucketTimestampSourceRecv
	}
	context := args.context.normalized()
	options := args.options.normalized()
	entry := RawAppendEntry{
		SchemaVersion:              RawEventLogSchemaVersion,
		CanonicalSchemaVersion:     args.canonicalSchemaVersion,
		CanonicalEventType:         args.canonicalEventType,
		CanonicalEventID:           canonicalEventID(args.canonicalEventType, args.venue, args.symbol, args.marketType, args.venueMessageID, args.venueSequence, args.streamKey, args.exchangeTs, args.recvTs),
		CanonicalPayload:           payload,
		VenueMessageID:             args.venueMessageID,
		VenueSequence:              args.venueSequence,
		StreamKey:                  args.streamKey,
		Symbol:                     args.symbol,
		Venue:                      args.venue,
		MarketType:                 args.marketType,
		StreamFamily:               args.streamFamily,
		SourceInstrumentID:         args.sourceInstrumentID,
		ExchangeTs:                 args.exchangeTs,
		RecvTs:                     args.recvTs,
		BucketTimestamp:            bucketTimestamp,
		BucketTimestampSource:      bucketSource,
		TimestampDegradationReason: args.timestampFallbackReason,
		Late:                       isLateRawEvent(args.canonicalEventTime, args.recvTs, options.LateAfter),
		NormalizerService:          options.NormalizerService,
		ConnectionRef:              context.ConnectionRef,
		SessionRef:                 context.SessionRef,
		BuildVersion:               options.BuildVersion,
		DegradedFeedRef:            args.degradedFeedRef,
		DuplicateAudit: RawDuplicateAudit{
			IdentityKey: rawIdentityPrecedenceKey(args.venueMessageID, args.venueSequence, args.streamKey, canonicalEventID(args.canonicalEventType, args.venue, args.symbol, args.marketType, args.venueMessageID, args.venueSequence, args.streamKey, args.exchangeTs, args.recvTs)),
		},
	}
	entry.PartitionKey = RouteRawPartition(entry)
	if err := ValidateRawAppendEntry(entry); err != nil {
		return RawAppendEntry{}, err
	}
	return entry, nil
}

func canonicalEventID(eventType string, venue Venue, symbol, marketType, venueMessageID string, venueSequence int64, streamKey, exchangeTs, recvTs string) string {
	input := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s|%s|%s", eventType, venue, symbol, marketType, venueMessageID, venueSequence, streamKey, exchangeTs, recvTs)
	sum := sha256.Sum256([]byte(input))
	return "ce:" + hex.EncodeToString(sum[:16])
}

func isLateRawEvent(bucketTime time.Time, recvTimestamp string, lateAfter time.Duration) bool {
	if bucketTime.IsZero() || recvTimestamp == "" || lateAfter <= 0 {
		return false
	}
	recvTime, err := time.Parse(time.RFC3339Nano, recvTimestamp)
	if err != nil {
		return false
	}
	return recvTime.Sub(bucketTime) > lateAfter
}

func rawIdentityPrecedenceKey(venueMessageID string, venueSequence int64, streamKey, canonicalEventID string) string {
	if venueMessageID != "" {
		return "message:" + venueMessageID
	}
	if venueSequence > 0 && streamKey != "" {
		return fmt.Sprintf("sequence:%s:%d", streamKey, venueSequence)
	}
	return "canonical:" + canonicalEventID
}

func rawDuplicateIdentityKey(entry RawAppendEntry) string {
	return rawIdentityPrecedenceKey(entry.VenueMessageID, entry.VenueSequence, entry.StreamKey, entry.CanonicalEventID)
}

func shouldPartitionByStreamFamily(streamFamily string) bool {
	switch streamFamily {
	case string(StreamOrderBook), string(StreamFundingRate), string(StreamOpenInterest), string(StreamMarkIndex), string(StreamLiquidation), RawStreamFamilyFeedHealth:
		return true
	default:
		return false
	}
}

func cloneRawAppendEntry(entry RawAppendEntry) RawAppendEntry {
	entry.CanonicalPayload = append(json.RawMessage(nil), entry.CanonicalPayload...)
	return entry
}
