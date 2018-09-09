Request for API access from <https://www.mytransport.sg/content/mytransport/home/dataMall/request-for-api.html>

Assuming `export accountkey="SECRET"`

# How to update Singapore bus stop information

	curl -X GET \
	  http://datamall2.mytransport.sg/ltaodataservice/BusStops \
	  -H 'AccountKey: SECRET'

# License

MIT

# Icon generator

Font is from https://dejavu-fonts.github.io/

## Related

* <https://github.com/cheeaun/arrivelah>
* <https://cheeaun.github.io/busrouter-sg/>
