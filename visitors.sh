#!/usr/bin/env bash
set -euo pipefail

# Distinct visitor cookies for complete UTC days or Monday-to-Sunday weeks.
mode=daily
default=30
max=9999
if [[ ${1:-} == --weekly ]]; then
    mode=weekly
    default=8
    max=52
    shift
fi
count=${1:-$default}
if (( $# > 1 )) || ! [[ $count =~ ^[1-9][0-9]{0,3}$ ]] || (( count > max )); then
    echo "Usage: $0 [days: 1-9999] | --weekly [weeks: 1-52]" >&2
    exit 2
fi

export AWS_PAGER="" AWS_REGION="${AWS_REGION:-ap-southeast-1}"
end=$(( $(date +%s) / 86400 * 86400 ))
if [[ $mode == weekly ]]; then
    # 1970-01-05 (four days after the epoch) was a Monday.
    end=$(( (end - 4 * 86400) / 604800 * 604800 + 4 * 86400 ))
    start=$(( end - (count + 1) * 604800 ))
    # Include a baseline week; one flag per cookie per week ignores repeat hits.
    query="filter msg = \"return visitor\" and toMillis(@timestamp) < $((end * 1000))
        | fields floor((toMillis(@timestamp) - $((start * 1000))) / 604800000) as week
        | stats "
    separator=""
    for ((i=0; i<=count; i++)); do
        query+="${separator}max(if(week = $i, 1, 0)) as w$i"
        separator=", "
    done
    query+=" by unique | stats sum(w0) as visitors0"
    for ((i=1; i<=count; i++)); do
        query+=", sum(w$i) as visitors$i, sum(w$((i-1)) * w$i) as returning$i"
    done
else
    start=$(( end - count * 86400 ))
    query="filter msg = \"return visitor\" and toMillis(@timestamp) < $((end * 1000))
        | stats countDistinct(unique) as visitors by bin(1d) as day
        | sort day asc"
fi

query_id=$(aws logs start-query --log-group-name /aws/lambda/ltabus \
    --start-time "$start" --end-time "$end" --limit 10000 \
    --query-string "$query" \
    --query queryId --output text)
trap 'aws logs stop-query --query-id "$query_id" >/dev/null 2>&1 || true' EXIT

while :; do
    result=$(aws logs get-query-results --query-id "$query_id" --output json)
    status=$(jq -er .status <<< "$result")
    case $status in
        Complete) break ;;
        Scheduled|Running) sleep 1 ;;
        *) echo "Query $query_id ended with status $status" >&2; exit 1 ;;
    esac
done
trap - EXIT

if [[ $mode == weekly ]]; then
    jq -r --argjson start "$start" --argjson weeks "$count" '
        def percent(n; d): if d > 0 then (1000 * n / d | round) / 10 else null end;
        (.results[0] // [] | map({key: .field, value: (.value | tonumber)}) | from_entries) as $c
        | ["week_start", "visitors", "previous_visitors", "growth_pct", "returning_visitors", "retention_pct"],
          (range(1; $weeks + 1) as $i
            | ($c["visitors\($i)"] // 0) as $current
            | ($c["visitors\($i - 1)"] // 0) as $previous
            | ($c["returning\($i)"] // 0) as $returning
            | [($start + $i * 604800 | strftime("%Y-%m-%d")),
               $current, $previous, percent($current - $previous; $previous),
               $returning, percent($returning; $previous)])
        | @csv
    ' <<< "$result"
    exit
fi

jq -r --argjson start "$start" --argjson end "$end" '
    [.results[] | map({key: .field, value: .value}) | from_entries
        | {key: .day[:10], value: (.visitors | tonumber)}]
    | from_entries as $counts
    | ["date", "visitors"],
      (range($start; $end; 86400) | strftime("%Y-%m-%d") | [., $counts[.] // 0])
    | @csv
' <<< "$result"
