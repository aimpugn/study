<?php
require_once dirname(__DIR__) . '/autoload.php';

$null1 = null;
print_r(intval($null1));

$number = 1000;
$powOfTen = pow(10, 1);

println([
    'number' => $number,
    'divide' => ($number / $powOfTen),
    'divided type' => gettype($number / $powOfTen),
    'epsilon' => (($number / $powOfTen) / 100),
]);

