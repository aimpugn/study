<?php
require_once dirname(__DIR__) . '/autoload.php';


print_r(json_encode([
	'key' => '어떤 값'
], JSON_UNESCAPED_UNICODE) . PHP_EOL);

$val = '\x22reason\x22:\x22\xEA\xB4\x80\xEB\xA6\xAC\xEC\x9E\x90\xED\x8E\x98\xEC\x9D\xB4\xEC\xA7\x80\xEC\xB7\xA8\xEC\x86\x8C\x22';

$json1 = '{"object":"secureInstruction"}';
$parsedJson1 = json_decode($json1);
$obj1 = new stdClass();
$obj1->status = 201;
$obj1->response = $parsedJson1;
print_r(gettype($parsedJson1));
print_r($obj1);
print_r($obj1->response->object);
// 가변길이 정수?
println("json encode bool", json_encode(["bool_true" => true, "bool_false" => false, "int_1" => 1], JSON_PRETTY_PRINT));

$tmp = [];

println(json_decode($tmp['null']), 'when encode undefined key');
println(json_decode(null), 'when encode null');
