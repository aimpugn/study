<?php
require_once dirname(__DIR__) . '/autoload.php';

$arr1 = [ // ! empty($arr[keyN])
	'key1' => true, // TRUE
	'key2' => false, // false
	'key3' => null, // false
	'key4' => '', // false
	'key5' => ' ', // TRUE
	'key6' => 0, // false
	'key7' => '0', // false
	'key8' => [], // false
	// undefined false
];
$tmp = [
	'space' => ' '
];
println("isset(undefined)", isset($undefined));
println("isset(undefined['undefined'])", isset($undefined['undefined']));
println("empty(trim(undefined['undefined']))", empty($undefined['undefined']));
println("check space", (! empty($tmp['space']) && trim($tmp['space'])));


println('trim 0', trim(0));
println('trim 0 gettype', gettype(trim(0)));
println('trim 1 gettype', gettype(trim(1)));

println("! empty(arr1['key1'])", ! empty(trim($arr1['key1'])));
println("! empty(arr1['key2'])", ! empty($arr1['key2']));
println("! empty(arr1['key3'])", ! empty($arr1['key3']));
println("! empty(arr1['key4'])", ! empty($arr1['key4']));
println("! empty(arr1['key5'])", ! empty($arr1['key5']));
println("! empty(arr1['key6'])", ! empty($arr1['key6']));
println("! empty(arr1['key7'])", ! empty($arr1['key7']));
println("! empty(arr1['key8'])", ! empty($arr1['key8']));
println("! empty(arr1['undefined'])", ! empty($arr1['undefined']));

$expected_array_got_string = 'somestring';
var_dump(empty($expected_array_got_string['some_key']));
var_dump(empty($expected_array_got_string[0]));
println("expected_array_got_string['0']", $expected_array_got_string['0']); // NOT empty
var_dump(empty($expected_array_got_string['0']));
println("expected_array_got_string[0.5]", $expected_array_got_string[0.5]); // NOT empty
var_dump(empty($expected_array_got_string[0.5]));
var_dump($expected_array_got_string[0.5]);
println("expected_array_got_string['0.5']", $expected_array_got_string['0.5']); // empty
var_dump(empty($expected_array_got_string['0.5']));
var_dump(empty($expected_array_got_string['0 Mostel']));