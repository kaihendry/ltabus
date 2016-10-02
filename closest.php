<?php


header('Content-Type: application/json');

$lat = $_REQUEST["lat"];
$lon = $_REQUEST["lon"];

if (empty($lat)) {
	$lat = 1.2994217;
}

if (empty($lon)) {
	$lon = 103.8555408;
}

$creds = parse_ini_file(".creds.ini");

$url = 'https://www.googleapis.com/fusiontables/v2/query?sql=SELECT+*+FROM+1kJscQXsc0jVMvrn2x5J93-5PpPXwc5zowERJSv8w+ORDER+BY+ST_DISTANCE' .
       '(geometry%2C+LATLNG(' . $lat . '%2C' . $lon . '))+LIMIT+10&key=' . $creds["fusionapikey"];


$ch = curl_init();
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_URL,$url);
$json=curl_exec($ch);
$info = curl_getinfo($ch);
$errinfo = curl_error($ch);
curl_close($ch);

echo $errinfo;

$r = json_decode($json);

$closest = array();
foreach ($r->rows as $calc) {
	array_push($closest, array("name" => $calc[0], "id" => sprintf("%05d", $calc[1])));
}
echo (json_encode($closest));


?>
