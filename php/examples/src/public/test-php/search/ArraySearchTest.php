<?php
require_once dirname(__DIR__) . '/autoload.php';

$arr1 = [
    'key0' => 0,
    'key1' => 1,
    'key2' => 2,
    'key3' => 3,
    'key4' => 4,
    'key5' => 5,
];

println(array_search(5, $arr1));