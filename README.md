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

# Rough visitor count

The existing `visitor` cookie is logged as `unique` on subsequent requests,
including requests for CSS and JavaScript on a first page load. In CloudWatch
Logs Insights, select `/aws/lambda/ltabus` in `ap-southeast-1`, choose a time
range, and run:

```text
filter msg = "return visitor"
| stats countDistinct(unique) as visitors
```

Add `by bin(1d)` for daily counts. This counts cookie IDs, so clearing cookies
or using another browser counts again; clients that never send the cookie back
are missed. It is a rough audience estimate without another service or database.
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

# Accountkey

Request for API access from <https://www.mytransport.sg/content/mytransport/home/dataMall/request-for-api.html>

# Related

- <https://github.com/cheeaun/arrivelah>
- <https://cheeaun.github.io/busrouter-sg/>
