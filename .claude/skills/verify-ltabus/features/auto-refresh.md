# Auto-refresh

A stop page stays current without a tap. The countdowns tick every second, "Last updated" counts up from the load, and the page reloads itself every 20 seconds while it is visible. It does not reload while the user is editing the stop number.

## Sub-features

- `refresh-tick`: the `<time>` countdowns and `#lastupdated` recompute from the clock every second.
- `refresh-reload`: a visible, untouched stop page reloads after 20s, with no extra browser history entry.
- `refresh-edit-guard`: an edited "Bus stop #" value is not wiped by the reload.
- `refresh-return`: returning to a hidden tab whose data is at least 20s old reloads it.

## How to get to it (user POV)

- Open any stop page and leave it open.
- Switch away from the tab or lock the phone, then come back after 20s or more.

## Driving it with bin/drive

Preconditions:

- `bin/doctor $RUN` passes.

Run `bin/drive $RUN auto-refresh refresh`. That runs the journey plus `e2e/refresh.spec.js`, and takes about 45s of real time.

- **Tick.** Go to `/?id=99999`. `#lastupdated` changes text within a few seconds.
- **Reload.** Wait for the next `load` event, up to 30s. The URL is still `/?id=99999`, the first `.buses time` has a new `datetime`, and `history.length` is unchanged.
- **Edit guard.** Covered by `e2e/refresh.spec.js` ("a partially entered stop survives…") in real time. "Bus stop #" is filled with `123`. After 21s the value is still `123` and there was only one navigation.
- **Proof.** Compare `1-before` with `2-reloaded`. The traces show the reload navigation.

## Gotchas

- `refresh-return` has no user-path journey. Headless Chromium is always visible. `e2e/refresh.spec.js` covers it as a regression test by faking `document.hidden` and the clock. To prove it for real, run with `--headed` and switch tabs by hand, or use a phone.
- `#lastupdated` counts negative (`0s`, `-1s`, `-2s`, …) on purpose: it shows how long ago the times were fetched, for reference. Do not report it as a bug.
- The page reloads only if "Bus stop #" still equals the loaded stop. Clicking into the field does not block the reload, but typing does.
