package contracts

import (
	"encoding/json"
	"fmt"
	"os"
)

type FixtureManifest struct {
	Fixtures []FixtureManifestEntry `json:"fixtures"`
}

type FixtureManifestEntry struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type ReplayManifest struct {
	Seeds []ReplayManifestEntry `json:"seeds"`
}

type ReplayManifestEntry struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type Fixture struct {
	FixtureVersion   string                   `json:"fixtureVersion"`
	ID               string                   `json:"id"`
	Family           string                   `json:"family"`
	Category         string                   `json:"category"`
	ScenarioClass    string                   `json:"scenarioClass"`
	Venue            string                   `json:"venue"`
	Symbol           string                   `json:"symbol"`
	QuoteCurrency    string                   `json:"quoteCurrency"`
	TargetSchema     string                   `json:"targetSchema"`
	Checks           []string                 `json:"checks"`
	ExpectedCanonical []map[string]any        `json:"expectedCanonical"`
}

type ReplaySeed struct {
	SchemaVersion       string              `json:"schemaVersion"`
	ID                  string              `json:"id"`
	Category            string              `json:"category"`
	Symbol              string              `json:"symbol"`
	TargetSchema        string              `json:"targetSchema"`
	Purpose             string              `json:"purpose"`
	FixtureRefs         []string            `json:"fixtureRefs"`
	Tags                []string            `json:"tags"`
	ExpectedDeterminism ExpectedDeterminism `json:"expectedDeterminism"`
}

type ExpectedDeterminism struct {
	EventCount           int      `json:"eventCount"`
	OrderedSourceRecordIDs []string `json:"orderedSourceRecordIds"`
}

var requiredFieldsBySchema = map[string][]string{
	"market-trade.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
	"order-book-top.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
	"feed-health.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"feedHealthState",
		"sourceRecordId",
	},
	"funding-rate.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
	"open-interest-snapshot.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
	"mark-index.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
	"liquidation-print.v1.schema.json": {
		"schemaVersion",
		"eventType",
		"symbol",
		"sourceSymbol",
		"quoteCurrency",
		"venue",
		"marketType",
		"exchangeTs",
		"recvTs",
		"timestampStatus",
		"sourceRecordId",
	},
}

var expectedEventTypeBySchema = map[string]string{
	"market-trade.v1.schema.json":           "market-trade",
	"order-book-top.v1.schema.json":         "order-book-top",
	"feed-health.v1.schema.json":            "feed-health",
	"funding-rate.v1.schema.json":           "funding-rate",
	"open-interest-snapshot.v1.schema.json": "open-interest-snapshot",
	"mark-index.v1.schema.json":             "mark-index",
	"liquidation-print.v1.schema.json":      "liquidation-print",
}

func LoadFixtureManifest(path string) (FixtureManifest, error) {
	var manifest FixtureManifest
	err := loadJSON(path, &manifest)
	return manifest, err
}

func LoadReplayManifest(path string) (ReplayManifest, error) {
	var manifest ReplayManifest
	err := loadJSON(path, &manifest)
	return manifest, err
}

func LoadFixture(path string) (Fixture, error) {
	var fixture Fixture
	err := loadJSON(path, &fixture)
	return fixture, err
}

func LoadReplaySeed(path string) (ReplaySeed, error) {
	var seed ReplaySeed
	err := loadJSON(path, &seed)
	return seed, err
}

func ValidateFixture(fixture Fixture) error {
	if fixture.FixtureVersion != "v1" {
		return fmt.Errorf("fixtureVersion must be v1")
	}
	if fixture.ID == "" || fixture.Family == "" || fixture.TargetSchema == "" {
		return fmt.Errorf("fixture metadata is incomplete")
	}
	if len(fixture.Checks) == 0 {
		return fmt.Errorf("fixture checks must not be empty")
	}
	if len(fixture.ExpectedCanonical) == 0 {
		return fmt.Errorf("expectedCanonical must not be empty")
	}

	requiredFields, ok := requiredFieldsBySchema[fixture.TargetSchema]
	if !ok {
		return fmt.Errorf("unsupported target schema %q", fixture.TargetSchema)
	}

	expectedEventType := expectedEventTypeBySchema[fixture.TargetSchema]
	for index, payload := range fixture.ExpectedCanonical {
		for _, field := range requiredFields {
			if _, ok := payload[field]; !ok {
				return fmt.Errorf("canonical payload %d missing field %q", index, field)
			}
		}

		schemaVersion, _ := payload["schemaVersion"].(string)
		if schemaVersion != "v1" {
			return fmt.Errorf("canonical payload %d has unsupported schemaVersion %q", index, schemaVersion)
		}

		eventType, _ := payload["eventType"].(string)
		if eventType != expectedEventType {
			return fmt.Errorf("canonical payload %d has eventType %q, expected %q", index, eventType, expectedEventType)
		}
	}

	return nil
}

func ValidateReplaySeed(seed ReplaySeed) error {
	if seed.SchemaVersion != "v1" {
		return fmt.Errorf("schemaVersion must be v1")
	}
	if seed.ID == "" || seed.TargetSchema != "replay-seed.v1.schema.json" {
		return fmt.Errorf("replay seed metadata is incomplete or targetSchema is unsupported")
	}
	if len(seed.FixtureRefs) == 0 || len(seed.Tags) == 0 {
		return fmt.Errorf("replay seed fixtureRefs and tags must not be empty")
	}
	if seed.ExpectedDeterminism.EventCount <= 0 {
		return fmt.Errorf("expectedDeterminism.eventCount must be positive")
	}
	if len(seed.ExpectedDeterminism.OrderedSourceRecordIDs) != seed.ExpectedDeterminism.EventCount {
		return fmt.Errorf("orderedSourceRecordIds length must match eventCount")
	}
	return nil
}

func loadJSON(path string, target any) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, target); err != nil {
		return err
	}
	return nil
}
