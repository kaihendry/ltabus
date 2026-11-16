# ltabus verification map

This map is the maintained source for verifying the user-facing behaviour of ltabus. Read this index first, then use the matching feature file as the recipe. Each feature has a journey in `../journeys/<feature>.spec.js`. Sub-features that `e2e/` or the Go tests already cover say so instead of having a second journey. `bin/drive` runs `e2e/` too. The Go tests run with `go test ./...`.

## Baseline preconditions

- Launch a fresh instance with `bin/launch` and keep its `RUN`.
- `bin/doctor $RUN` reports no `FAIL`.
- Chromium for Playwright 1.62.1 is installed. If it is not, run `npx playwright install chromium`.
- `ACCOUNTKEY` is only needed for journeys tagged `[needs ACCOUNTKEY]`.
- Never drive an instance this run did not launch, and never drive bus.dabase.com.

## Driving conventions

- Every Playwright test gets a fresh browser context. That means no cookies and no localStorage, which matches a first visit.
- Prefer labels, roles and the stable ids listed in SKILL.md over CSS position.
- Stop `99999` arrival text is deterministic for 30s after each page load. Assert it straight after loading.
- Use real time. A test that needs the 20s refresh waits for it.

## Proof and skip reporting

- UI proof is a `proof()` screenshot plus an ARIA snapshot from before and after the action, plus the trace.
- Side-effect proof is the relevant `server.log` lines, saved into the test's artifacts.
- Report a `[needs ACCOUNTKEY]` failure as unverified for lack of a key. Do not count it as a pass, and do not claim it was covered by stop `99999`.
- Report every entry point you did not drive.

## Feature entry contract

Each feature file has an H1 title and one paragraph, followed by exactly these H2s: `Sub-features`, `How to get to it (user POV)`, `Driving it with bin/drive`, and `Gotchas`.

## Features

- [Stop arrivals](./stop-arrivals.md) covers looking up a stop, sorted countdowns, load colours, input validation, the stop icon, the no-JS page and live datamall stops.
- [Auto-refresh](./auto-refresh.md) covers the countdown ticking, the page reloading every 20s, and an edited stop number surviving the reload.
- [Closest stop](./closest-stop.md) covers geolocation from the home page and the 🚏 link redirecting through `/closest`.
- [Visit history](./visit-history.md) covers the list of visited stops with counts on every page.
- [Visitor count](./visitor-count.md) covers the `visitor` cookie and the `unique=` log field behind `visitors.sh`.

Not yet mapped: the "Share this Web app!" button (`navigator.share` is missing in headless Chromium and needs a phone) and PWA install from `static/manifest.json`.
