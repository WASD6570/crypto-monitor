# Replay And Analytics UI Testing

## Testing Goals

- prove the UI can answer the time-to-trust questions for a recent alert using only service-owned review data
- prove stale, empty, degraded, and partial-data paths remain usable and honest
- prove the SPA does not rederive alert, outcome, replay, or analytics business logic in the client
- prove performance-sensitive paths defer heavy replay and analytics work until needed

## Output Artifact

- Write the implementation test results to `plans/epics/replay-and-analytics-ui/testing-report.md`.

## Test Environment Assumptions

- `apps/web` has fixture-backed component tests and a browser smoke harness such as Vitest plus Testing Library and/or Playwright.
- Mock or fixture responses can represent recent alert detail, replay coverage variants, aggregate analytics variants, and request failures.
- Service-owned fixtures pin timestamps, statuses, and ordering so the UI can be checked for faithful rendering.

## Validation Commands

- `pnpm --dir apps/web test -- --runInBand replay analytics`
- `pnpm --dir apps/web test:e2e --grep "alert review|replay window|analytics slices"`
- `pnpm --dir apps/web build`
- `pnpm --dir apps/web exec playwright test tests/review-smoke.spec.ts --project=chromium`

## Smoke Matrix

### 1. Recent Alert Review Happy Path

- Inputs: recent alert list fixture, full alert detail fixture, full replay coverage fixture, completed outcome fixture, saved simulation-summary fixture
- Verify:
  - alert stream renders recent items in stable order from service payload
  - selecting an alert shows why-fired explanation, outcome summary, replay coverage, and simulation summary
  - outcome labels and decisive status match fixture values exactly
  - simulation summary is visibly separate from outcome summary

### 2. Empty And First-Run States

- Inputs: empty recent alert list fixture, empty analytics slice fixture
- Verify:
  - page shell renders without crash
  - empty copy is explicit and filter-aware
  - no fake zero-performance metrics or invented replay placeholders appear

### 3. Stale Data And Partial Availability

- Inputs: stale alert detail fixture, stale analytics fixture, delayed simulation-summary fixture
- Verify:
  - stale badges show per panel using supplied timestamps
  - old detail remains visible during retry or refresh failure
  - analytics age is independent from alert-detail age
  - missing simulation summary does not block outcome review

### 4. Degraded Replay Coverage

- Inputs: partial replay coverage fixture, unavailable replay fixture, degraded timestamp-basis fixture
- Verify:
  - replay panel shows exact partial or unavailable status from payload
  - missing segments are labeled, not interpolated
  - alert detail remains usable when replay is partial or absent

### 5. Regime Slices And Best/Avoid Summaries

- Inputs: analytics aggregate fixture with regime cohorts, fragmentation cohorts, degraded cohorts, low-sample best/avoid entries
- Verify:
  - regime panels preserve service ordering and labels
  - low-sample entries carry caution state
  - best/avoid items render sample counts and freshness
  - client does not reorder entries into a different ranking than the payload

### 6. Mobile And Disclosure Behavior

- Inputs: same fixtures as happy path plus small viewport
- Verify:
  - primary alert summary and why-fired fields are visible without horizontal scroll
  - secondary panels stay collapsed or drawer-based until requested
  - tap targets and keyboard navigation remain usable

## Negative Cases

- Alert detail payload omits optional simulation fields: UI must show simulation absence, not derived net-viability claims.
- Outcome payload contains `UNKNOWN` or pending status: UI must render that exact neutral state, not green/red success styling by default.
- Replay payload includes missing coverage spans or retention gap reason: UI must label the gap and avoid drawing uninterrupted history.
- Analytics payload has stale timestamp but fresh alert list exists: page must show different ages per panel and avoid single-page freshness claims.
- Analytics payload returns low sample counts: best/avoid summaries must show caution markers and sample counts instead of strong conclusions.
- Service returns request error after previous success: panel keeps last successful data when safe and shows scoped retry/error state.
- Unknown future enum value for a service-owned reason or state: UI shows safe fallback label without crashing or mapping it to an existing business meaning.

## Fidelity Checks Against Trust Boundary

- Inspect tests for any client-side computation of regime, severity, outcome ordering, simulation result, or best/avoid ranking; only formatting and display transforms are allowed.
- Assert rendered status text and rank order come directly from fixtures for at least one detail view and one analytics view.
- Assert replay coverage UI uses explicit coverage metadata, not inferred continuity from timestamps alone.

## Performance And Bundling Checks

- Verify replay payload is not requested until the replay panel is opened.
- Verify analytics bundle or route can load without blocking the base review shell.
- Verify `pnpm --dir apps/web build` completes without introducing SSR-specific dependencies or Next.js assumptions.
- If bundle analysis is available, check that any added heavy visualization dependency is justified; otherwise prefer native tables and lightweight primitives.

## Manual Review Notes

- Open a recent alert and confirm the first viewport answers: what fired, why it was allowed, whether outcome exists, and whether replay is fully covered.
- Toggle to a stale or degraded fixture and confirm the page reduces trust explicitly without hiding data.
- Compare a best-condition row and avoid-condition row against fixture JSON to ensure labels, sample sizes, and ordering are preserved.

## Exit Criteria

- all listed validation commands pass
- happy path and negative-path fixtures render without crashes
- stale, empty, degraded, and partial states are explicit and conservative
- review UI remains within Vite SPA constraints and avoids client-owned business logic
