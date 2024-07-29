<?php
require_once dirname(__DIR__) . '/autoload.php';

$arr1 = [
    'key' => 'value',
    'key2' => 'value2'
];

println(isset($arr1['key'], $arr1['key2']), "isset(arr1['key'], arr1['key2'])");
println(isset($arr1['key'], $arr1['not_exists']), "isset(arr1['key'], arr1['not_exists'])");
println(isset($arr1['not_exists1'], $arr1['not_exists2']), "isset(arr1['not_exists1'], arr1['not_exists2'])");

$arr2 = null;
println(isset($arr2['key']), "isset(null['key'])");

$params1 = [
    'key1' => 'value1',
    'key2' => 'value2',
    'key3' => 'value3',
    'null' => null,
];
$missed1 = [];

# https://stackoverflow.com/a/72561794
# https://modernpug.github.io/php-the-right-way/pages/The-Basics.html
# true ?: when false
isset($params1['key1']) ?: $missed1[] = 'key1';
isset($params1['key2']) ?: $missed1[] = 'key2';
isset($params1['key3']) ?: $missed1[] = 'key3';
isset($params1['key4']) ?: $missed1[] = 'key4';
isset($params1['key5']) ?: $missed1[] = 'key5';
println(implode(', ', $missed1), 'missed');
println(isset($params1['null']), 'when value is null');