Request for API access from <http://www.mytransport.sg/content/mytransport/home/dataMall.html>

Create `.creds.ini` and populate the values for:

	uniqueuserid=""
	accountkey=""

Make sure your Webserver does not serve dotfiles, e.g. <http://bus.dabase.com/.creds.ini> is 403 Forbidden.

# How to update Singapore bus stop information

	curl -X GET \
	  http://datamall2.mytransport.sg/ltaodataservice/BusStops \
	  -H 'AccountKey: SECRET'

# License

MIT

## Related

* <https://github.com/cheeaun/arrivelah>
* <https://cheeaun.github.io/busrouter-sg/>
