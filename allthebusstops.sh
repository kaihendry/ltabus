#!/bin/bash

if ! test "$ACCOUNTKEY"; then
	echo env \"ACCOUNTKEY\" is unset
	exit
fi

pwd=$PWD
cd "$(mktemp -d)" || exit

count=0
while :; do
	curl -s -f -X GET https://datamall2.mytransport.sg/ltaodataservice/BusStops?\$skip=$count -H "accountkey: $ACCOUNTKEY" |
		jq .value[] >$count.json
	test -s "$count.json" || break
	count=$((count + 500))
done

jq . ./*.json | jq -s . >"$pwd/all.json"
newBusCount=$(jq 'length' "$pwd/all.json")
echo "Bus stop count: $newBusCount"

mv $pwd/all.json $pwd/static/
