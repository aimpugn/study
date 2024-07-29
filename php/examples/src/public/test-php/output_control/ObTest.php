<?php

require_once dirname(__DIR__) . '/autoload.php';

println(ob_get_level(), "when script start");

$obCount = 0;
ob_start();
$obCount++;
println(ob_get_level(), "after {$obCount} ob_start");
ob_start();
$obCount++;
println(ob_get_level(), "after {$obCount} ob_start");
ob_start();
$obCount++;
println(ob_get_level(), "after {$obCount} ob_start");

print "foo" . PHP_EOL; // NOT printed
ob_end_clean();$obCount--;
println(ob_get_level(), "after {$obCount} ob_end_clean");

print "bar" . PHP_EOL; // Printed
ob_end_flush();$obCount--;

println(ob_get_level(), "after {$obCount} ob_end_flush");

print "baz" . PHP_EOL;
println(str_replace('\n', ' ', json_encode(ob_get_contents())), "after {$obCount} ob_get_contents");
println(ob_get_level(), "after {$obCount} ob_get_level");

println(ob_get_clean(), "after {$obCount} ob_get_clean");
println(ob_get_level(), "after {$obCount} ob_get_level");
