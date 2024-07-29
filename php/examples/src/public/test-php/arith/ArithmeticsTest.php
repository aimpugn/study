<?php
require_once dirname(__DIR__) . '/autoload.php';

error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

$totalCount = 110;
$successCount = 100;
$failCount = 10;
println($successCount);
println($failCount);

$tmp = sprintf("Success : %d, Fail : %d (%s)\r", $successCount, $failCount, number_format($successCount + $failCount / $totalCount * 100, 2) . "%");
println($tmp);

$tmp = sprintf("Success : %d, Fail : %d (%s)\r", $successCount, $failCount, "tmp");
println(number_format($failCount / $totalCount, 2), 'test title');
println(number_format(($successCount + $failCount) / $totalCount * 100, 2) . "%");

$tmp1 = 1000;
$tmp2 = 500;
$tmp3 = null;
unset($undef);

println([
    1000 - 500 - null,
    gettype(1000 - 500 - null)
], "1000 - 500 - null");
println([
    1000.00 - 500.00 - null,
    gettype(1000.00 - 500.00 - null)
], "1000.00 - 500.00 - null");

//println([
//    $tmp1 - $tmp2 - $tmp3 - $undef, // PHP Warning:  Undefined variable $undef
//    gettype($tmp1 - $tmp2 - $tmp3 - $undef)
//], "$tmp1 - $tmp2 - $tmp3 - $undef");
