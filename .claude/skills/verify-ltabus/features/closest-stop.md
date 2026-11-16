# Closest stop

A user who opens the home page, or taps the 🚏 sign, is sent to the bus stop nearest to their location. The browser's geolocation is passed to `/closest`, which redirects to that stop's arrivals page.

## Sub-features

- `closest-home`: loading `/` with location permission redirects to `/?id=<nearest stop>` via a 302 from `/closest`.
- `closest-sign`: the 🚏 link (`aria-label` "Find closest bus stop") on any page does the same.
- `closest-arrivals`: the destination shows that stop's live arrivals.

## How to get to it (user POV)

- Open `https://<host>/` and allow location.
- Tap 🚏 at the top of any page.

## Driving it with bin/drive

Preconditions:

- `bin/doctor $RUN` passes.
- The browser context grants `geolocation` at `1.29684825487647, 103.85253591654006`, which is exactly stop `01012` (Hotel Grand Pacific).
- `closest-arrivals` needs `ACCOUNTKEY`.

Run `bin/drive $RUN closest-stop`.

- **Home.** Go to `/`. The response for `?id=01012` was redirected from `/closest` with status 302, and the page URL ends in `/?id=01012`.
- **Sign.** Go to `/?id=99999` and click `getByRole("heading").getByRole("link", { name: "Find closest bus stop" })`. The page URL ends in `/?id=01012`.
- **Arrivals.** `[needs ACCOUNTKEY]` After the home redirect, the status is 200 and `#namedBusStop` reads `Hotel Grand Pacific`.
- **Proof.** Check `arrived`, `relocated` and `closest-arrivals`. The trace shows the `/closest?lat=…&lng=…` request. `server.log` has `req_path="/closest?..." res_status=302`.

## Gotchas

- Without `ACCOUNTKEY` the redirect still works, but the destination is a 424 page. `arrived` and `relocated` screenshots then show that error text, which is expected.
- Without the `geolocation` permission, `/` simply stays put. Headless Chromium does not show a prompt.
- `/closest` uses squared lat/lng distance, not real distance. Stand exactly on a stop's coordinates from `static/all.json` to get a deterministic answer.
- `static/all.json` is replaced by the weekly bus stop update. If `01012` moves or disappears, pick another stop from the file and update both the journey and this file.
