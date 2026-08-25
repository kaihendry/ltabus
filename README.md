# Singapore bus arrival Web application

Proving that a Web application sucks less than a native App to get bus arrival time given a bus stop!

# Test bus stop

Stop `99999` makes up its own buses, so it never calls datamall and needs no
ACCOUNTKEY: <https://bus.dabase.com/?id=99999>. Handy for `go run .` locally,
and it is what the tests and the deploy check drive.

# Accountkey

Request for API access from <https://www.mytransport.sg/content/mytransport/home/dataMall/request-for-api.html>

# Related

- <https://github.com/cheeaun/arrivelah>
- <https://cheeaun.github.io/busrouter-sg/>
