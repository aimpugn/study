<?php
require_once dirname(__DIR__) . '/autoload.php';
error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

println(intval(0.01), 'intval 0.01');
println(intval(1.01), 'intval 1.01');
println(intval('1.02'), "intval '1.02'");
println(intval('ddd3'), "intval 'ddd3'");
println(intval('1ddd3'), "intval 'ddd3'");
println((int) 0.01, '(int) 0.01');
println((int) 1.01, '(int) 1.01');
println((int) '1.02',  "(int) '1.02'");
println((int) 'ddd3',  "(int) 'ddd3'");
println((int) '1ddd3',  "(int) '1ddd3'");
$arr1 = null;
println(is_numeric($arr1),  "is_numeric(null)");
println(isset($arr1['test']) && is_numeric($arr1['test']),  "arr1['test'] && is_numeric(null['test'])");
println(((int) ' 1 2 3'),  "((int) ' 1 2 3')");
println(((int) trim(' 1 2 3')),  "((int) trim(' 1 2 3'))");
println(intval(' 1 2 3'),  "intval(' 1 2 3')");
$tmp1 = ' 1 2 3';
$tmp1 = (int)preg_replace('/\D/', '', $tmp1);
println([
    '$tmp1' => $tmp1,
    'type' => gettype($tmp1)
], 'replace not number to empty');