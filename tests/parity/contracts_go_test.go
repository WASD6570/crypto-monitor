package parity

import (
	"path/filepath"
	"runtime"
	"testing"

	contracts "github.com/crypto-market-copilot/alerts/libs/go/contracts"
)

func TestGoConsumerValidatesFixtures(t *testing.T) {
	repoRoot := repoRoot(t)
	manifest, err := contracts.LoadFixtureManifest(filepath.Join(repoRoot, "tests/fixtures/manifest.v1.json"))
	if err != nil {
		t.Fatalf("load fixture manifest: %v", err)
	}

	for _, entry := range manifest.Fixtures {
		fixture, err := contracts.LoadFixture(filepath.Join(repoRoot, entry.Path))
		if err != nil {
			t.Fatalf("load fixture %s: %v", entry.ID, err)
		}
		if err := contracts.ValidateFixture(fixture); err != nil {
			t.Fatalf("validate fixture %s: %v", entry.ID, err)
		}
	}
}

func TestGoConsumerValidatesReplaySeeds(t *testing.T) {
	repoRoot := repoRoot(t)
	manifest, err := contracts.LoadReplayManifest(filepath.Join(repoRoot, "tests/replay/manifest.v1.json"))
	if err != nil {
		t.Fatalf("load replay manifest: %v", err)
	}

	for _, entry := range manifest.Seeds {
		seed, err := contracts.LoadReplaySeed(filepath.Join(repoRoot, entry.Path))
		if err != nil {
			t.Fatalf("load replay seed %s: %v", entry.ID, err)
		}
		if err := contracts.ValidateReplaySeed(seed); err != nil {
			t.Fatalf("validate replay seed %s: %v", entry.ID, err)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}
