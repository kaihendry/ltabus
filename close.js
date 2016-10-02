var x = document.getElementById("output");

function geoError(data) {
	if (data.message) {
	x.innerHTML = data.message;
	} else {
	x.innerHTML = data;
	}
	console.log(data);
}

if(navigator.geolocation) {
	navigator.geolocation.getCurrentPosition(function(position) {
		closest(position.coords.latitude, position.coords.longitude);
	}, geoError);
} else { geoError("No Geolocation"); }

function closest(lat, lon) {

	
	console.debug(lat, lon);
	var xmlhttp = new XMLHttpRequest();
	var url = "closest.php?lat=" + lat + "&lon=" + lon;

	xmlhttp.onreadystatechange = function() {
		if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
			var myArr = JSON.parse(xmlhttp.responseText);
			console.log(myArr);
			myFunction(myArr);
		}
	}
	xmlhttp.open("GET", url, true);
	xmlhttp.send();

	function myFunction(arr) {
		var out = "<ul>";
		var i;
		for(i = 0; i < arr.length; i++) {
			console.log(arr[i]);
			out += '<li><a href=/?id=' + arr[i].id + '&name=' + encodeURI(arr[i].name) + '>' +
				arr[i].id + ' ' + arr[i].name + '</a></li>';
		}
		x.innerHTML = out + "</ul>";
	}
}
