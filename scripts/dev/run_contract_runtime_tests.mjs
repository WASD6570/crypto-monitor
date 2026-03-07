import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import {
  loadJson,
  validateFixtureForTsConsumer,
  validateReplaySeedForTsConsumer,
} from '../../libs/ts/contracts/runtime.mjs';

const mode = process.argv[2];

if (mode !== 'contracts') {
  console.error('Only the contracts test mode is supported. Run: pnpm test contracts');
  process.exit(1);
}

const currentDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(currentDir, '..', '..');

const fixtureManifest = loadJson(join(repoRoot, 'tests/fixtures/manifest.v1.json'));
for (const entry of fixtureManifest.fixtures) {
  validateFixtureForTsConsumer(loadJson(join(repoRoot, entry.path)));
}

const replayManifest = loadJson(join(repoRoot, 'tests/replay/manifest.v1.json'));
for (const entry of replayManifest.seeds) {
  validateReplaySeedForTsConsumer(loadJson(join(repoRoot, entry.path)));
}

console.log('TypeScript runtime contract tests passed.');
