<?php
require_once dirname(__DIR__) . '/autoload.php';

$expiry = "2023-10";
println(substr($expiry, 2, 2), 'substr($expiry, 2, 2)');
println(substr($expiry, 4), 'substr($expiry, 4)');

$strToSplit = "123abc";

println(str_split($strToSplit), 'str_split');
