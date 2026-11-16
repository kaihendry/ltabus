# Stop arrivals

A user enters a five-digit bus stop number and sees the stop's name and, for each bus service, up to three countdowns to its next arrivals. Services are sorted by the soonest bus, and each countdown is coloured by how full the bus is.

## Sub-features

- `arrivals-lookup`: the form at `/` submits to `/?id=<stop>` and renders the stop name and its services.
- `arrivals-order`: services are listed soonest first.
- `arrivals-load`: countdowns carry `load-seats`, `load-standing` or `load-full`.
- `arrivals-validate`: anything other than five digits is not submitted.
- `arrivals-icon`: each stop gets a 200×200 PNG icon at `/icon?stop=<stop>`, used for the favicon and the home-screen icon.
- `arrivals-nojs`: the countdown is correct without JavaScript.
- `arrivals-live`: a real stop shows live datamall data.

## How to get to it (user POV)

- Type a stop number into "Bus stop #" on any page and press "Bus due times".
- Open a link such as `/?id=99999` directly (a shared link or a home-screen bookmark).
- Tap a stop in the visit history list (see [visit-history](./visit-history.md)).

## Driving it with bin/drive

Preconditions:

- `bin/doctor $RUN` passes.
- `arrivals-live` additionally needs `ACCOUNTKEY` set at launch.

Run `bin/drive $RUN stop-arrivals`.

- **Look up.** Go to `/`. Fill `getByLabel("Bus stop #")` with `99999` and click `getByRole("button", { name: "Bus due times" })`. The URL becomes `/?id=99999`, the title is `Singapore bus arrival times for stop 99999`, and `#namedBusStop` reads `Test Bus Stop`.
- **Order and countdowns.** `.buses li strong` reads `42`, `666` and `.buses time` reads `1m`, `5m`, `17m`, `32m`.
- **Load colours.** `.buses time` has the classes `load-seats`, `load-standing`, `load-full`, `load-seats`. The service link `42` points to `https://busrouter.sg/#/services/42`.
- **Validate.** Fill `123` and submit. The URL stays at `/` and no `.buses li` appears.
- **Icon.** `link[rel=icon]` is `/icon?stop=99999`. That URL returns a 200 `image/png` of 200×200, which is saved as `icon-99999.png`. `/icon?stop=abc` returns 400.
- **No JS.** With JavaScript disabled, the same lookup shows the same four countdowns.
- **Live.** `[needs ACCOUNTKEY]` Go to `/?id=01012`. Status 200 and `#namedBusStop` reads `Hotel Grand Pacific`.
- **Proof.** Check `1-entered` and `2-arrivals` (`.png` and `.aria.yml`), `rejected`, `nojs`, `icon-99999.png`, and each test's `trace.zip`.

## Gotchas

- Without `ACCOUNTKEY`, a real stop still makes an outbound call to datamall (about 300ms) and then shows a plain-text 424 page reading `datamall API is returning, bad response: …`. That page loads no app.js, so the stop is not added to history.
- Do not assert a bus count for a real stop. Late at night in Singapore (roughly 00:30–05:30 SGT) a stop can legitimately have no buses.
- The countdowns for `99999` change 30s after load: `1m` becomes `60s`, then `59s`, and so on. Assert straight away, or compare against the `datetime` attribute.
- `static/` and the HTML template are embedded at build time. After editing `static/index.html`, `app.js` or `app.css`, relaunch. Doctor flags this as `STALE`.
