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
$info = curl_getinfo($ch);
$errinfo = curl_error($ch);
curl_close($ch);

$j = json_decode($result, true);

if (empty($j)) {
	echo "<h1>ERROR</h1><pre>";
	print_r($info);
	print_r($errinfo);
	echo "</pre>";
	die("<h1>No result from LTA API</h1>");
}

if (isset($j["odata.error"])) { die ($j["odata.error"]["message"]["value"]); }

function my_sort($a, $b) {
    if ($a["NextBus"]["EstimatedArrival"] < $b["NextBus"]["EstimatedArrival"]) {
        return -1;
    } else if ($a["NextBus"]["EstimatedArrival"] > $b["NextBus"]["EstimatedArrival"]) {
        return 1;
    } else {
        return 0;
    }
}

usort($j["Services"], 'my_sort');

?>
<!DOCTYPE html>
<html>
<head>
<title><?php echo $j["BusStopID"] . " " . $_GET["name"]; ?></title>
<meta name=viewport content="width=device-width, initial-scale=1">
<link rel='icon' href='data:;base64,iVBORw0KGgo='>
<style>
body { padding: 5px; font-size: 120%; }
a { font-size: 110%; }
ul,ol { padding-left: 0; }
.buses li { list-style: none; }
.buses li:before {
  content: "\1F68C";
  padding-right: 8px;
}

ol { padding-left: 0; list-style: none; }
#stations li:before {
  content: "🚏";
  padding-right: 8px;
}

.busstopid { white-space: nowrap; display:inline-block; border-bottom: thin solid black; padding-bottom:2px; margin: 0 }
</style>
</head>
<body>
<h3 class=busstopid><?php echo "🚏" . $j["BusStopID"] . " " . $_GET["name"]; ?></h3>
<ul class=buses>
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
	echo "<strong><a href='https://busrouter.sg/#/services/" . $service["ServiceNo"] . "'>" . $service["ServiceNo"] . "</a></strong> ";
	echo tmark($service["NextBus"]) . ", ";
	if (isset($service["SubsequentBus"]["EstimatedArrival"])) {
		echo tmark($service["SubsequentBus"]);
	}
	echo "</li>\n";
}
?>
</ul>
<h4>Last updated: <span id=lastupdated></span></h4>
<form>
<input required type=number inputmode="numeric" pattern="[0-9]*" value=<?php echo $id;?> name=id>
<input type=submit>
</form>

<p><a href=/close.html>Experimental: Closest stops</a></p>

<ol id=stations></ol>

<p><a href=/map.html>Map of Singapore bus stops</a></p>

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
	var count = 1;
	if (typeof slog['<?php echo $id;?>'] == "number") {
		count = slog['<?php echo $id;?>'];
	}
	try { count = slog['<?php echo $id;?>'].count + 1; } catch(e) { console.log(e); }
	<?php if ($_GET["name"]) { ?>
	slog['<?php echo $id;?>'] = { "name": "<?php echo $_GET["name"]; ?>", "x": "<?php echo $_GET["lat"]; ?>", "y": "<?php echo $_GET["lon"]; ?>", "count": count };
<?php } else { ?>
	slog['<?php echo $id;?>'].count = count;
<?php } ?>
	//console.debug(slog);
	localStorage.setItem('history', JSON.stringify(slog));
	//console.debug(localStorage['history']);

	var sortable = [];
	for (var station in slog) {
		sortable.push([station, slog[station]])
	}
	sortable.sort(function(a, b) {return a[1] - b[1]})
	//console.debug(sortable);
	var ul = document.getElementById("stations");
	for (var i = sortable.length - 1; i >= 0; i--) {
		var key = sortable[i][0];
		var value = sortable[i][1];
		//console.log(key, value);

		var li = document.createElement("li");
		var link = document.createElement('a');
		link.setAttribute('href', '/?id=' + key + '&name=' + encodeURI(value.name));
		if (value.name) {
		link.appendChild(document.createTextNode(key + " " + value.name));
		} else {
		link.appendChild(document.createTextNode(key));
		}
		li.appendChild(link);
		ul.appendChild(li);
	}

}, false);
</script>
<footer><a href=https://github.com/kaihendry/ltabus>Source code</a>&diam;<a href="mailto:hendry+bus@iki.fi">Email feedback</a></footer>
</body>
</html>
