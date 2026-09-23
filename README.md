# Singapore bus arrival Web application

Proving that a Web application sucks less than a native App to get bus arrival time given a bus stop!

# Test bus stop

Stop `99999` makes up its own buses, so it never calls datamall and needs no
ACCOUNTKEY: <https://bus.dabase.com/?id=99999>. Handy for `go run .` locally,
and it is what the tests and the deploy check drive.

# Local development

Run `PORT=8081 go run .` and open <http://localhost:8081/?id=99999>.
CSS and JavaScript are embedded directly from `static/`; no asset build is needed.
Restart Go after editing them. Node is only needed for `make browsertest`.

The functional browser tests run in CI and block deployment on failure.
The throttled performance measurement is opt-in because it reports timings
without a pass/fail threshold:

```sh
BENCHMARK=1 npx playwright test e2e/throttled.spec.js
CACHE=warm BENCHMARK=1 npx playwright test e2e/throttled.spec.js
```

# Arrival-page reliability

In CloudWatch Logs Insights, select `/aws/lambda/ltabus` in `ap-southeast-1`
and a recent time range. This query measures real stop-page request latency
and HTTP failures, excluding the fictional test stop:

```text
filter msg = "response"
| filter req_path like /^\/\?/ and req_path like /[?&]id=[0-9]{5}(&|$)/
| filter req_path not like /[?&]id=99999(&|$)/
| stats count(*) as requests,
    pct(duration, 95) as p95_ms,
    pct(duration, 99) as p99_ms,
    sum(if(res_status = 424, 1, 0)) as upstream_failures,
    sum(if(res_status >= 500 and res_status < 600, 1, 0)) as server_failures,
    100 * sum(if(res_status = 424 or (res_status >= 500 and res_status < 600), 1, 0)) / count(*) as failure_pct
  by bin(1d)
```

Duration includes upstream access, decoding, and rendering; it does not isolate
Datamall latency. These are HTTP failures, not a measure of arrival prediction
accuracy. Browser tests and the deployment smoke check use stop `99999`, so
they do not measure live Datamall reliability.

A seven-day query run on 23 September 2026 found 128,838 matching requests,
15 HTTP 424 responses (0.012%), and no 5xx responses. Daily p95 durations
were 96–120 ms and p99 durations were 215–249 ms; the first and last daily
buckets were partial days. This sample does not indicate an urgent HTTP
reliability problem.

# Rough visitor count

The `visitor` cookie is logged as `unique` on every page request, including
the first visit. Existing cookie IDs are preserved; static assets and icons
neither create nor log visitor identities. In CloudWatch
Logs Insights, select `/aws/lambda/ltabus` in `ap-southeast-1`, choose a time
range, and run:

```text
filter msg = "return visitor"
| stats countDistinct(unique) as visitors
```

Add `by bin(1d)` for daily counts. This counts cookie IDs, so clearing cookies
or using another browser counts again; clients that never send the cookie back
can be counted again on each page request. It is a rough audience estimate without another service or database.
See [CloudWatch aggregation functions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_QuerySyntax-Stats.html).

For CSV from the CLI, install `aws` and `jq` and configure AWS credentials with
`logs:StartQuery`, `logs:GetQueryResults`, and `logs:StopQuery` access:

```sh
./visitors.sh > visitors.csv                 # last 30 complete UTC days
AWS_PROFILE=your-profile ./visitors.sh 90 > visitors.csv
```

The script uses `ap-southeast-1` by default (`AWS_REGION` overrides it), waits
for the query to complete, and writes `date,visitors` columns in date order.
Days with no matching log entries get zero, including days outside log retention.
Daily counts are distinct within each day; summing them does not give unique
visitors for the whole period.

For weekly growth and retention:

```sh
AWS_PROFILE=mine ./visitors.sh --weekly 8 > weekly-visitors.csv
```

This reports eight complete Monday-to-Sunday UTC weeks, querying one additional
baseline week. `growth_pct` compares unique visitors with the previous week.
`returning_visitors` counts cookies present in both weeks, and `retention_pct`
is that count divided by the previous week's visitors. Percentages are blank
when the previous week had no visitors. This measures return visits by active
visitors, rather than retention of newly acquired users. Weekly calculations
happen in CloudWatch; individual cookie IDs are not exported.
Keep the report and its baseline week within the log group's retention period
to avoid comparing incomplete weeks.

# Accountkey

Request for API access from <https://www.mytransport.sg/content/mytransport/home/dataMall/request-for-api.html>

# Related

- <https://github.com/cheeaun/arrivelah>
- <https://cheeaun.github.io/busrouter-sg/>
