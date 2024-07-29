<?php
require_once dirname(__DIR__) . '/autoload.php';

$arr1 = [
	" val1 ",
	"  val2\r",
	" val3\n",
	" val4\r\n",
];
print_r($arr1);

$arr2 = array_map('trim', $arr1);

print_r($arr2);
