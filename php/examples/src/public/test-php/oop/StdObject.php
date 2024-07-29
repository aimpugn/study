<?php
require_once dirname(__DIR__) . '/autoload.php';

$test = [
	'status' => 200,
	'response' => [
		'result' => true
	]
];

//print_r($test->status); // PHP Notice:  Trying to get property of non-object in
$testObj = (object)$test;
print_r("(object)test, testObj->status: $testObj->status\n");

$genericObject = new stdClass();
$genericObject->status = 200;
print_r("genericObject = new stdClass(), genericObject->status: $genericObject->status\n");
