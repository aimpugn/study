<?php
require_once dirname(__DIR__) . '/autoload.php';


$arr1 = [
	'key1' => 'value1',
	'key5' => 'value5',
];

$arr2 = [
	'key1',
	'key2',
	'key3',
	'key5'
];
$arr3 = array_intersect_key($arr1, array_flip($arr2));
$arr4 = [
	'user' => []
];
$arr4['user'] += $arr3;

$arr5 = [
	[
		'id' => 2,
		'pg_tid' => null
	],
	[
		'id' => 1,
		'pg_tid' => 'some valid pg tid'
	],
	[
		'id' => 0,
		'pg_tid' => ' '
	],
];

$arr6 = array_filter($arr5, function ($var) {
	return !empty($var['pg_tid']) && trim($var['pg_tid']);
});

var_dump(! empty(trim($arr6['test'])));

$arr7 = ['copy_test'];
$arr8 = $arr7;

$arr8[0] .= '_modified';

println($arr7, 'arr7');
println($arr8, 'arr8');

$tmp = [];
foreach($tmp as $value) {
	println("value: $value" );
}

$tmp2 = "1234";
foreach($tmp2 as $value) {
	println("value: $value" );
}

$arr9 = [1, 2, 3, 4];
$arr10 = [1, 2, 3, 4];
println(array_merge($arr9, $arr10), 'array_merge($arr9, $arr10)');

$arr11 = [];
println($arr10 + $arr11);

println(var_export(in_array(1, $arr10), true));
println(['x'] + ['a', 'bfd', 'd']);
println(array_merge(['x'], ['a', 'bfd', 'd']));
$arr12 = ['x', 'a', 'bfd', 'd'];
array_shift($arr12);
println($arr12);
$arr13 = ['x', 'a', '', 'bfd', "", 'd', false, true, "false"];

println(array_filter($arr13));

$arr14 = [
    'key' => '',
];

println(var_export(isset($arr14['key2']['key3']) && is_string($arr14['key2']), true));

$arr15 = [
    'key' => [1, 2,3, ],
];

println(var_export(isset($arr15['key']) && $arr15['key'], true));

$arr16 = [
    'key1' => 'val1',
];
println(array_merge($arr16, ['key1' => 'val1_modified']), "array_merge arr16, ['key2' => 'val2_modified']");
println($arr16 + ['key1' => 'val1_modified'], "arr16 + ['key2' => 'val2_modified']");

$extendedData = [
    'card_number' => '123456*****'
];
$extension = array_intersect_key($extendedData, array_flip(array(
    'card_number',
    'card_type',
    'card_issue_code',
    'is_free_interest',
    'emb_pg_provider',
)));
println($extension, '$extension');
println(var_export($extension['card_number'] && $extension['card_type'] === null, true), '$extension[card_number] && $extension[card_type] === null');

$assoc1 = [
    'k1' => 'v1',
    'k2' => 'v2',
    'k3' => 'v3',
    'k4' => 'v4',
];
$assoc2 = [
    'k1' => 'v1_modified',
    'k5' => 'v5',
];
$assoc3 = [];
println($assoc1 + $assoc2, '$assoc1 + $assoc2');
println($assoc1 + $assoc3, '$assoc1 + $assoc3');

println(array_filter([
    'country_code' => 'KR',
    'CC_BRAND' => null,
    'PAYMENT_SYSTEM_NAME' => '',
], function ($el) {
    return !empty(strval($el));
}), 'array_filter with callback');

println(array_filter([
    'country_code' => 'KR',
    'CC_BRAND' => null,
    'PAYMENT_SYSTEM_NAME' => '',
]), 'array_filter');