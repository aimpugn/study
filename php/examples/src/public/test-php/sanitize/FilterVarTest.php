<?php
require_once dirname(__DIR__) . '/autoload.php';

println([
    var_export(filter_var('true', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE), true),
    gettype(filter_var('true', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE))
]);
println([
    var_export(filter_var('Yes', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE), true),
    gettype(filter_var('Yes', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE))
], 'When filter_var Yes');

println([
    var_export((boolean)'true', true),
    gettype((boolean)'true')
]);
$false = (boolean)'false';
println([
    var_export($false, true),
    gettype($false)
]);

$false = 'false';
println([
    var_export(settype($false, 'boolean'), true),
    gettype(settype($false, 'boolean'))
], "settype(false, 'boolean')");

println([
    var_export(json_decode('true'), true),
    gettype(json_decode('true'))
], "json_decode('true')");
println([
    var_export(json_decode('false'), true),
    gettype(json_decode('false'))
], "json_decode('false')");
println([
    var_export(json_decode('fase'), true),
    gettype(json_decode('fase'))
], "json_decode('fase')");
println([
    var_export(json_decode('tru'), true),
    gettype(json_decode('tru'))
], "json_decode('tru')");
println([
    var_export(json_decode('Yes'), true),
    gettype(json_decode('Yes'))
], "json_decode('Yes')");
