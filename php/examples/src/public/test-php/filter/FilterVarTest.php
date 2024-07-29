<?php
require_once dirname(__DIR__) . '/autoload.php';

println(var_export(filter_var('true'), true), "filter_var('true')");
println(var_export(filter_var('Y'), true), "filter_var('Y')");
println(var_export(filter_var('true', FILTER_VALIDATE_BOOLEAN), true),
    "filter_var('true', FILTER_VALIDATE_BOOLEAN)");
println(var_export(filter_var('Y', FILTER_VALIDATE_BOOLEAN), true),
    "filter_var('Y', FILTER_VALIDATE_BOOLEAN)");
println(var_export(filter_var('Yes', FILTER_VALIDATE_BOOLEAN), true),
    "filter_var('Yes', FILTER_VALIDATE_BOOLEAN)");