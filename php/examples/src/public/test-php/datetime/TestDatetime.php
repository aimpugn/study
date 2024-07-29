<?php
require_once dirname(__DIR__) . '/autoload.php';

println([
    'microtime(false)' => microtime(false),
    'type' => gettype(microtime(false))
]);
println(((double) explode(' ', microtime(false))[0]) * 1000, 'microtime(false) * 1000');
println(date('l, d-M-y H:i:s T'));
println(date('D, d M Y H:i:s \G\M\T'));
println(date('Y-m-d\TH:i:s.u'));
println(time(), 'time()');
println(microtime(true), 'microtime(true)');
$tmp = microtime(true) * 10000;
println([
    $tmp,
    (int) $tmp,
    gettype($tmp),
    strlen("$tmp")
], 'microtime(true) * 10000');

println(strtotime('2022-10-20T08:32:46+00:00'), 'strtotime 2022-10-20T08:32:46+00:00');
println(strtotime('2022-10-20T08:32:46Z'), 'strtotime 2022-10-20T08:32:46Z');
println(strtotime('20221020T083246Z'), 'strtotime 20221020T083246Z');
println(strtotime('2022-01-01T00:00:00.000'), 'strtotime 2022-01-01T00:00:00.000');
println(date('Y-m-d H:i:s', strtotime('2022-01-01T00:00:00.000')), 'strtotime 2022-01-01T00:00:00.000 to Y-m-d H:i:s');
