Request for API access from <http://www.mytransport.sg/content/mytransport/home/dataMall.html>

Create `.creds.ini` and populate the values for:

	uniqueuserid=""
	accountkey=""

Make sure your Webserver does not serve dotfiles, e.g. <http://bus.dabase.com/.creds.ini> is 403 Forbidden.

# TODO

Name bus stops and link to map to find bus stop ids.

	wget http://www.mytransport.sg/content/dam/mytransport/DataMall_StaticData/Geospatial/BusStops.zip
	unzip BusStops.zip
	ogr2ogr -f KML [Jan2015.kml](http://s.natalian.org/2015-04-28/Jan2015.kml) BusStop.shp

ogr2ogr binary is from the [gdap package](https://www.archlinux.org/packages/community/i686/gdal/)
