# Visit history

Every stop a user opens is remembered in this browser and listed below the arrivals on every page, with how many times it was opened. The most visited stops come first. Tapping one opens it.

## Sub-features

- `history-record`: opening a stop page increments its count and stores its name.
- `history-list`: `#stations` lists stops as `<code> <name> (<count>)` on every page, including `/`.
- `history-open`: a listed stop links to `/?id=<code>`.
- `history-order`: the most visited stops come first.

## How to get to it (user POV)

- Open a few stop pages, then open the home page or any other stop page and look below "Last updated".

## Driving it with bin/drive

Preconditions:

- `bin/doctor $RUN` passes.
- The browser context is fresh, with empty localStorage.
- `history-order` needs `ACCOUNTKEY`, because it needs a second stop that renders.

Run `bin/drive $RUN visit-history`.

- **Record.** Go to `/?id=99999`. `#stations li` reads `99999 Test Bus Stop (1)`.
- **List.** Open `/?id=99999` again, then go to `/`. `#stations li` reads `99999 Test Bus Stop (2)`.
- **Open.** Click the link `99999 Test Bus Stop (2)`. The URL ends in `/?id=99999` and the list reads `(3)`.
- **Order.** `[needs ACCOUNTKEY]` Open `01012` once and `99999` twice, then go to `/`. The list reads `99999 Test Bus Stop (2)` and then `01012 Hotel Grand Pacific (1)`.
- **Proof.** Check `1-home`, `2-revisited` and `order`.

## Gotchas

- History lives in this browser's localStorage only. The server never sees it, so `server.log` cannot prove it.
- A 424 page (a real stop without a key) runs no app.js and records nothing, which is why `history-order` needs the key.
- `e2e/stations.spec.js` covers ordering, `&` in names and corrupt storage by seeding localStorage. That is a regression test, not proof of the user path.
