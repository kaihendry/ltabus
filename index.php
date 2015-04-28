<?php

if (isset($_GET["id"]) && is_numeric($_GET["id"])) {
	$id = $_GET["id"];
} else {
	$id = "07379";
}

$url = "http://datamall2.mytransport.sg/ltaodataservice/BusArrival?BusStopID=$id";

$creds = parse_ini_file(".creds.ini");

$ch = curl_init();
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_URL,$url);
$headers = array(
	'UniqueUserId:' . $creds["uniqueuserid"],
	'AccountKey:' . $creds["accountkey"],
);
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
$result=curl_exec($ch);
curl_close($ch);

$j = json_decode($result, true);
if (isset($j["odata.error"])) { die ($j["odata.error"]["message"]["value"]); }

?>
<!DOCTYPE html>
<html>
<head>
<title><?php echo $j["BusStopID"]; ?></title>
<meta name=viewport content="width=device-width, initial-scale=1">
<style>
ul#buses li { list-style: none; }
ul#buses li:before {
  content: "\1F68C";
}

#busstopid {
  transform: rotate(-90deg);
  -webkit-transform: rotate(-90deg);
  width: 5em;
  position: absolute;
  left: -2em;
  top: -1em;
}
</style>
</head>
<body>
<h1 id=busstopid><?php echo $j["BusStopID"]; ?></h1>
<ul id=buses>
<?php

function tmark($s) {
	// var_dump(trim($s["Load"]));

	switch (trim($s["Load"])) {
	case "Seats Available":
		$color = "green";
		break;
	case "Standing Available":
		$color = "orange";
		break;
	case "Limited Standing":
		$color = "red";
		break;
	default:
		$color = "black";
		break;
	}

	return '<time style="color: ' . $color . '" dateTime="' . $s["EstimatedArrival"] . '">' . $s["EstimatedArrival"] . '</time>';
}

foreach ($j["Services"] as $service) {
	if (empty($service["NextBus"]["EstimatedArrival"])) { continue; }
	echo "<li>";
	echo "<strong>" . $service["ServiceNo"] . "</strong> ";
	echo tmark($service["NextBus"]) . ", ";
	if (isset($service["SubsequentBus"]["EstimatedArrival"])) {
		echo tmark($service["SubsequentBus"]);
	}
	echo "</li>\n";
}
?>
</ul>
<h4>Last updated: <span id=lastupdated></span> ago</h4>
<form>
<input required type=number value=<?php echo $id;?> name=id>
<input type=submit>
</form>
<ol id=stations></ol>

<script>
function countdown(id, time) {
	// console.log(id,time);
	var seconds =  time / 1000;
	if (Math.abs(seconds) > 60) {
		id.innerHTML = parseInt(seconds / 60) + "m";
	} else {
		id.innerHTML = parseInt(seconds) + "s";
	}
	setTimeout(countdown,1000, id, time - 1000);
}

window.addEventListener('load', function() {
	var timings = document.getElementsByTagName("time");
	var now = new Date();
	for (var i = 0; i < timings.length; i++) {
		var arr = new Date(timings[i].getAttribute("datetime"));
		var elapsed = arr.getTime() - now.getTime();
		countdown(timings[i], elapsed);
	}
	var lastupdated = document.getElementById("lastupdated");
	countdown(lastupdated, Date.now() - now);


	slog = (JSON.parse(localStorage.getItem("history")) || {});
	slog['<?php echo $id;?>'] = 1 + (slog['<?php echo $id;?>'] || 0);
	localStorage.setItem('history', JSON.stringify(slog));
	console.debug(localStorage['history']);

	var sortable = [];
	for (var station in slog) {
		sortable.push([station, slog[station]])
	}
	sortable.sort(function(a, b) {return a[1] - b[1]})
	// console.debug(sortable);
	var ul = document.getElementById("stations");
	for (var i = sortable.length - 1; i >= 0; i--) {
		var key = sortable[i][0];
		var value = sortable[i][1];
		// console.log(key, value);

		var li = document.createElement("li");
		var link = document.createElement('a');
		link.setAttribute('href', '/?id=' + key);
		link.appendChild(document.createTextNode(key));
		li.appendChild(link);
		ul.appendChild(li);
	}

}, false);
</script>
<footer><a href=https://github.com/kaihendry/ltabus>Source code</a></footer>
</body>
</html>
