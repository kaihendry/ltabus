#!/usr/bin/env bash
set -euo pipefail

# Daily distinct visitor cookies for the last N complete UTC days.
days=${1:-30}
if (( $# > 1 )) || ! [[ $days =~ ^[1-9][0-9]{0,3}$ ]]; then
    echo "Usage: $0 [days: 1-9999] > visitors.csv" >&2
    exit 2
fi

export AWS_PAGER="" AWS_REGION="${AWS_REGION:-ap-southeast-1}"
end=$(( $(date +%s) / 86400 * 86400 ))
start=$(( end - days * 86400 ))

query_id=$(aws logs start-query --log-group-name /aws/lambda/ltabus \
    --start-time "$start" --end-time "$end" --limit 10000 \
    --query-string "filter msg = \"return visitor\" and toMillis(@timestamp) < $((end * 1000))
        | stats countDistinct(unique) as visitors by bin(1d) as day
        | sort day asc" \
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

jq -r --argjson start "$start" --argjson end "$end" '
    [.results[] | map({key: .field, value: .value}) | from_entries
        | {key: .day[:10], value: (.visitors | tonumber)}]
    | from_entries as $counts
    | ["date", "visitors"],
      (range($start; $end; 86400) | strftime("%Y-%m-%d") | [., $counts[.] // 0])
    | @csv
' <<< "$result"
