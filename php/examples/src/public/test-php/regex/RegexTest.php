<?php
require_once dirname(__DIR__) . '/autoload.php';
error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

$string = ' test     string  $  ';
$result = preg_replace('/t\s*s/', ' ', $string);
println($result);

$cardBins = [
	'123456',
	'1234567',
	'1234567 ',
	' 1234567 ',
	'12345678',
	'123456**',
	'123456ds',
	'123456&`',
	'123456&`',
	'12345🗅`', # https://www.compart.com/en/unicode/U+1F5C5
	'12345🗌`', # https://www.compart.com/en/unicode/U+1F5CC
	'**123456**',
];

foreach($cardBins as $cardBin) {
	$converted = preg_replace('/[^0-9]/', '', $cardBin);
	$convertedLen = strlen($converted);
	println("$cardBin to $converted(length: $convertedLen)");
}

$test1 = preg_replace('/\D/', '', 1234.01);
println(preg_replace('/\D/', '', 1234), 'preg replace number');
$test2 = preg_replace('/\D/', '', 1234.01);
println(preg_replace('/\D/', '', null), 'preg replace null');
$test3 = preg_replace('/\D/', '', 1234.01);
println(preg_replace('/\D/', '', 1234.01), 'preg replace float');