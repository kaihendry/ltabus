---
name: verify-ltabus
description: Launch, drive and prove the ltabus Singapore bus-arrival web app (Go server rendering HTML, plus static/app.js) in a real Chromium browser. Use to confirm that a change to stop arrivals, the countdown or auto-refresh, the closest-stop redirect, the visit history, or the visitor cookie logging works end to end. Each run leaves screenshots, ARIA snapshots, Playwright traces and the server log as evidence.
---

# Verify ltabus

The user touches one surface: the web page at `/` and `/?id=<5-digit stop>`. Other surfaces are `/closest?lat=&lng=` (redirect), `/icon?stop=` (PNG) and `/static/*`. Production is AWS Lambda at bus.dabase.com. Never verify against production: every visit there is counted in the real visitor numbers.

Stop `99999` is a fictional test stop. Its buses are made up relative to the server clock (42 in 1m30s; 666 in 5m30s, 17m30s and 32m30s), and it needs no network or key. Real stops call LTA datamall and need `ACCOUNTKEY` in the environment. Without it they answer HTTP 424.

All helpers live in `.claude/skills/verify-ltabus/bin/` (written `bin/` below). Run them from the repo root.

## Launch

```sh
.claude/skills/verify-ltabus/bin/launch        # prints RUN=<dir> and URL=<url>
```

Launch builds this checkout with `go build`, starts the binary on a free port bound to 127.0.0.1, and returns once `/?id=99999` renders its four `<time>` elements. That takes about 3s. `ACCOUNTKEY` is passed through when your environment has it. The repo's `.envrc` fetches the key from SSM, so it needs the `mine` AWS profile. If your shell doesn't have it (doctor says so), launch with `direnv exec . .claude/skills/verify-ltabus/bin/launch`. Never print or read the key's value. Every later command takes the printed `RUN` dir. Pass it explicitly, because shell state does not persist between tool calls. Each launch is isolated: its own port and its own `RUN` dir under `.verify/` (gitignored). Concurrent runs do not collide, and there is no database or other shared state.

Restart after any change to `*.go` or `static/`, because both are embedded in the binary at build time. To restart, run `bin/cleanup $RUN` and then `bin/launch`. Doctor reports a forgotten restart as `STALE`.

## Doctor

```sh
.claude/skills/verify-ltabus/bin/doctor $RUN
```

Doctor is read-only. It checks that the process is up, that the port belongs to that process, that the `X-Version` header matches the launched `git describe`, that the source has not changed since the build, and that the test stop renders. It also notes whether `ACCOUNTKEY` is set. Run it first whenever something looks off. `bin/drive` runs it automatically and refuses to continue on any `FAIL`.

## Drive

```sh
.claude/skills/verify-ltabus/bin/drive $RUN                              # journeys and e2e/ (~60s)
.claude/skills/verify-ltabus/bin/drive $RUN --project journeys           # user-path proof only
.claude/skills/verify-ltabus/bin/drive $RUN stop-arrivals nojs           # one feature (file-name filters)
.claude/skills/verify-ltabus/bin/drive $RUN --grep-invert ACCOUNTKEY     # skip live-datamall journeys
```

The config is `playwright.config.js` in this skill. It never starts a server and never reuses one. It has two projects:

- **`journeys`** are the specs in `journeys/`, one per feature in `features/`. They drive real user paths and are the proof.
- **`e2e`** is the repo's own regression suite in `e2e/`, the same tests CI runs before deploy. Some of these tests fake the clock, the tab's visibility or localStorage, so report them as regression results, not as user-path proof.

Each journey covers only what `e2e/` and the Go tests do not. The feature files say which test covers each sub-feature. Any arguments after `$RUN` go to `npx playwright test`, so `--headed`, `--grep`, `--project` and file filters all work. Do not use the repo's `make browsertest` for proof: its config reuses whatever is already listening on :8081, which may be the user's own stale `go run`.

To prove a new behaviour, add a test to the matching journey, or a new journey plus a feature file. Drive through user handles only: `getByLabel("Bus stop #")`, `getByRole("button", { name: "Bus due times" })`, `getByRole("link", { name: "Find closest bus stop" })`, `#namedBusStop`, `.buses li`, `.buses time`, `#lastupdated`, `#stations li`. Call `proof(page, testInfo, name)` from `journeys/proof.js` before and after each action that matters. Use `serverLog()` to check side effects.

## Evidence

Each drive writes to `$RUN/drive-<HHMMSS>/`:

- `report.txt` holds the pass/fail list.
- `artifacts/<test>/` holds `*.png` and `*.aria.yml` from `proof()`, the final `test-*.png`, and `trace.zip`. Open a trace with `npx playwright show-trace <trace.zip>` to see every action alongside its DOM snapshot.
- `$RUN/server.log` holds one `msg=response` line per request, with status, path, duration and `unique=<visitor id>` once a cookie came back.

Proof standards:

- Drive the real user path: type into the form, click the 🚏 link, wait for the real 20s refresh. Do not use `page.clock`, override `document.hidden` or seed localStorage. Those belong in `e2e/` regression tests, not in proof.
- Capture the state before and after the action, not just the final screen.
- Check side effects alongside the screen: the server log for cookies and status codes, and `localStorage` only through what the page renders from it.
- Datamall is the only external boundary. Do not mock it. A journey that needs it is tagged `[needs ACCOUNTKEY]` and fails without the key. Report that failure as such, and never weaken the journey so it passes.

## Cleanup

```sh
.claude/skills/verify-ltabus/bin/cleanup $RUN
```

Cleanup kills only the pid this run recorded, and only if that pid is still this run's binary. It then deletes the binary. Everything else in `$RUN` stays: evidence, `server.log`, version and fingerprint. Delete old runs with `rm -rf .verify/<run>` only when you no longer need the proof. Run cleanup after a failed attempt too, so you do not strand servers.

## Features

See [features/README.md](features/README.md). A change to a mapped feature is proven only when every entry point in its file has been driven.
