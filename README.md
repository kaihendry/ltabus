Request for API access from <http://www.mytransport.sg/content/mytransport/home/dataMall.html>

Create `.creds.ini` and populate the values for:

	uniqueuserid=""
	accountkey=""

Make sure your Webserver does not serve dotfiles, e.g. <http://bus.dabase.com/.creds.ini> is 403 Forbidden.

# How to update Singapore bus stop information

	wget http://www.mytransport.sg/content/dam/mytransport/DataMall_StaticData/Geospatial/BusStops.zip
	unzip BusStops.zip
	ogr2ogr -f GeoJSON -t_srs crs:84 bus-stops.json BusStops_Jan2015/BusStop.shp

ogr2ogr binary is from the [gdal package](https://www.archlinux.org/packages/community/i686/gdal/)

# License

MIT

## Related

* <https://github.com/cheeaun/arrivelah>
* <https://cheeaun.github.io/busrouter-sg/>
