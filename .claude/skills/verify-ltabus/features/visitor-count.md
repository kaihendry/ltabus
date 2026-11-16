# Visitor count

The site estimates its audience without any analytics service. A page visit hands the browser a random `visitor` cookie, and every request that sends it back is logged with `unique=<id>`. `visitors.sh` and the README's CloudWatch queries count those IDs.

## Sub-features

- `visitor-set`: a page request without the cookie gets `Set-Cookie: visitor=visitor-<random>` for one year.
- `visitor-count`: requests carrying the cookie log `unique=<id>` on their `msg=response` line. A new browser does this immediately for the page's CSS, JS, icon and manifest.
- `visitor-keep`: a return visit keeps the same ID and sets no new cookie.
- `visitor-assets`: `/static/*` and `/icon` never hand out an ID.

## How to get to it (user POV)

- Open any page in a new browser. The operator later sees one visitor in `visitors.sh`.

## Driving it with bin/drive

Preconditions:

- `bin/doctor $RUN` passes.
- The browser context is fresh, with no cookies.

Run `bin/drive $RUN visitor-count`.

- **Set.** Go to `/?id=99999`. The document response has `set-cookie` matching `^visitor=visitor-[A-Z0-9]+;`.
- **Count.** `server.log` lines after the visit include `unique=<that id>` on the asset requests. The first document line has no `unique=`.
- **Keep.** Reload. There is no `set-cookie`, and the context cookie is unchanged.
- **Assets.** A cookie-less request for `/static/app.css` has no `set-cookie`.
- **Proof.** Check `server-log.txt` in the test's artifacts, which holds exactly the lines this visit caused.

## Gotchas

- Locally the logger writes text (`unique=visitor-…`). On Lambda it writes JSON (`"unique":"visitor-…"`). Match the local format here.
- `bin/doctor` and `curl` requests send no cookie, so they never add `unique=` lines. They do add `msg=response` lines, so slice the log from an offset taken before the action.
- On Lambda, cookies must appear only in the API Gateway v2 `cookies` field (`lambdaHandler`). That path is unit-tested in `http_test.go` and cannot be driven locally.
