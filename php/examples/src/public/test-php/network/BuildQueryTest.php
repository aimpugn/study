<?php
require_once dirname(__DIR__) . '/autoload.php';

$query1 = [
    'key1' => 'value1',
    'key2' => 'val ue2',
    'key3' => '',
    'key4' => 'value3',
];

println(http_build_query($query1, PHP_QUERY_RFC3986), 'PHP_QUERY_RFC3986');
println(http_build_query($query1, PHP_QUERY_RFC1738), 'PHP_QUERY_RFC1738');
$query2 = [];
foreach($query1 as $key => $value) {
    $query2[$key] = urlencode($value);
}
println(http_build_query($query2, PHP_QUERY_RFC3986), 'rawurlencoded');

$query3 = [
    'imp_uid' => 'imp_897072122501',
    'request_id' => 'req_1660748071590',
    'user_code' => 'imp68124833'
];
println(http_build_query($query3, PHP_QUERY_RFC3986), '$query3 by PHP_QUERY_RFC3986');
println(http_build_query($query3), '$query3 by default');
