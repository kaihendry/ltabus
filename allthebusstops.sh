#!/bin/bash

test -f .env && source .env

if ! test "$ACCOUNTKEY"
then
	echo env \"ACCOUNTKEY\" is unset
	exit
fi

pwd=$PWD
cd "$(mktemp -d)" || exit

count=0
while :
do
	curl -s -f -X GET http://datamall2.mytransport.sg/ltaodataservice/BusStops?\$skip=$count -H "accountkey: $ACCOUNTKEY" |
	jq .value[] > $count.json
	test -s "$count.json" || break
	count=$((count+500))
done

jq . ./*.json | jq -s . > "$pwd/all.json"
busstopcount=$(jq 'length' "$pwd/all.json")
echo "Bus stop count: $busstopcount"

mv $pwd/all.json $pwd/static/

sed -i "s,{{ totalStops }},$busstopcount," "$pwd/static/index.html"
