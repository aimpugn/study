<?php

require_once dirname(__DIR__) . '/autoload.php';

$requestUrl = 'http://sub_b.localhost/users/getToken';
$headers = [
    'Content-Type: application/json'
];
$postStr = json_encode([
    "apiKey" => "apiKeyValue",
    "apiSecret" => "apiSecretValue"
]);
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $requestUrl);
curl_setopt($ch, CURLOPT_PORT, 8001);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
curl_setopt($ch, CURLOPT_POSTFIELDS, $postStr);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_TIMEOUT, 60); //60s timeout(naver recommended)

//execute post
$body = curl_exec($ch);
$errno = curl_errno($ch);
$error = curl_error($ch);
$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$contentLength = curl_getinfo($ch, CURLINFO_CONTENT_LENGTH_DOWNLOAD);

println($body);
println('strlen: ' . strlen($body));
println($errno);
println($error);
println($status);
println('contentLength: ' . $contentLength);
