<?php
require_once dirname(__DIR__) . '/autoload.php';

println(sprintf("%03d", 11));

println(sprintf("'%s'", addslashes('"2022-09-01 22:53:00"')));
println(sprintf("'%s'", addslashes("2022-09-01 '22:53:00")));
println(sprintf("'%s'", addslashes('paid')));
$tmp = date("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'");
println($tmp);

println(sprintf('%03d', 22));
println(sprintf('%03d', 322));
