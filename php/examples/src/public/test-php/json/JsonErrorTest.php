<?php
require_once dirname(__DIR__) . '/autoload.php';

$invalid =  '\xB1\x31';
$invalid = json_encode($invalid);
println(json_last_error(), "json_last_error");
println($invalid, 'test');

$text = "\xB1\x31";

$json  = json_encode($text);
$error = json_last_error();

var_dump($json, $error === JSON_ERROR_UTF8);
println([
    'json_last_error_msg()' => json_last_error_msg(),
    'json_last_error()' => json_last_error(),
], 'json_last_error_message1');


$invalid = '{"test":"value}';
try {
    $try1 = json_decode($invalid);
    print_r(var_export($try1, true) . PHP_EOL);
} catch (Exception $_) {
    print_r('exception?? ' . $_->getMessage());
}
print_r([
    'json_last_error_msg()' => json_last_error_msg(),
    'json_last_error()' => json_last_error(),
]);