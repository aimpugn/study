<?php
require_once dirname(__DIR__) . '/autoload.php';

$status = 'SUCCESS';

if ($status !== 'CANCELED' && $status !== 'PARTIAL_CANCELED') {
	print_r("NOT CANCELED" . PHP_EOL);
}
$arr1 = [];
var_dump(filter_var($arr1['escrow_confirmed'], FILTER_VALIDATE_BOOLEAN));

function testControl($bool = true, $innerBool = true)
{
    if($bool) {
        if ($innerBool) {
            println('When $bool is true, $innerBool is true');
        } else {
            println('When $bool is true, $innerBool is false');
        }
    } else {
        println('When $bool is false');
        if ($innerBool) {
            println('When $bool is false, $innerBool is true');
        } else {
            println('When $bool is false, $innerBool is false');
        }
    }

    println('At the end');
}

testControl(true, true);
