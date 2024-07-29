<?php
require_once dirname(__DIR__) . '/autoload.php';
error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

$number = 135.79;

function toString($decimalData) {
    $NANO_FACTOR = 1000000000;
    $units = $decimalData['units'];
    $nanos = $decimalData['nanos'];

    if (!empty($decimalData)) {
        if (!empty($units) || $units == "0") {
            if (!empty($nanos)) {
                return $units + ($nanos / $NANO_FACTOR / 100);
            }
            return $units;
        }
    }
    return '0';
}

function toDouble($decimalData) {
    $NANO_FACTOR = 1000000000;
    $units = $decimalData['units'];
    $nanos = $decimalData['nanos'];

    if (!empty($decimalData)) {
        if (!empty($units) || $units == "0") {
            if (!empty($nanos)) {
                return $units + ($nanos / $NANO_FACTOR / 100);
            }
            return floatval($units);
        }
    }
    return floatval(0);
}

println(round($number, 3) . ' and type is ' . gettype(round($number, 3)));
println(round($number, 2) . ' and type is ' . gettype(round($number, 2)));
println(round($number, 1) . ' and type is ' . gettype(round($number, 1)));
println(round($number, 0) . ' and type is ' . gettype(round($number, 0)));
println(round($number, -1) . ' and type is ' . gettype(round($number, -1)));
println(round($number, -2) . ' and type is ' . gettype(round($number, -2)));
println(round($number, -3) . ' and type is ' . gettype(round($number, -3)));
$decimalData1 = ['units' => "50000", 'nanos' => 0];
$decimalDataString1 = toString($decimalData1);
println('1. ' . $decimalDataString1 . ' and type is ' . gettype($decimalDataString1));
$decimalData2 = ['units' => "50000", 'nanos' => 10];
$decimalDataString2 = toString($decimalData2);
println('2. ' . $decimalDataString2 . ' and type is ' . gettype($decimalDataString2));
$decimalData3 = ['units' => "50000", 'nanos' => 500000000];
$decimalDataString3 = toString($decimalData3);
println('3. ' . $decimalDataString3 . ' and type is ' . gettype($decimalDataString3));

$rounded = round(floatval(50000), 2);
println($rounded . ' and type is ' . gettype($rounded));
println(var_export($rounded === toString(['units' => "50000", 'nanos' => 0]), true));

println(0.0 + "50000", '0.0 + "50000"');
println(gettype(0.0 + "50000"), 'type of 0.0 + "50000"');

println(round(floatval('1234.567'), 2), "round(floatval('1234.567'), 2)");
println(floatval('1234.56'), "floatval('1234.56')");
println([
    "round(floatval('1234.00'), 2)" => round(floatval('1234.00'), 2),
    "type" => gettype(round(floatval('1234.00'), 2))
], "round(floatval('1234.00'), 2)");
println([
    "floatval('1234.00')" => floatval('1234.00'),
    "type" => gettype(floatval('1234.00'))
], "floatval('1234.00')");
println(floatval('1234'), "floatval('1234')");

$test1 = toDouble([
    'units' => "0",
    'nanos' => 5 * 1000000000
]);
println([
    'toDouble()' => $test1,
    'type' => gettype($test1),
], 'unit 0, nano 500 * 1000000000');
println(round(floatval('127.543xfre'), 2), "round(floatval('127.543xfre'), 2)");
println(floatval('0'), "floatval('0')");
println(floatval('1'), "floatval('1')");
println(floatval('1.'), "floatval('1.')");
println(floatval('1.1'), "floatval('1.1')");
println(floatval('2.13'), "floatval('2.13')");


function equals($double, $other)
{
    if (is_numeric($other)) {
        return $double === floor(floatval($other) * 100) / 100;
    }

    return false;
}

function toDouble2($val, $precision = 2)
{
    return floor(floatval($val) * 100) / 100;
}


println([
    'val' => toDouble2('123.4567'),
    'type' => gettype(toDouble2('123.4567'))
], 1);

