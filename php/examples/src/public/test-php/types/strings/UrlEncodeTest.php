<?php
require_once dirname(__DIR__) . '/autoload.php';

$urls = [
    'http://www.example.com:8800',
    'http://www.example.com/a/b/c/d/e/f/g/h/i.html',
    'http://www.test.com?pageid=123&testid=1524',
    'http://www.test.com/do.html#A',
    'http://stackoverflow.com/users/9999999/not a-real-user',
    'http//:stackoverflow.com/questions/9715606/bad-url-test-cases',
    'http://stackoverflow.com/users/9999999/not-a-real-user',
    'https://example.com/this"should-be-quoted',
    'https://example.com/thisisa"quote/helloworld/',
    'https://example.com?key="value"',
];

foreach($urls as $url) {
    println(urlencode($url), "Original: $url");
    println(rawurldecode($url), "Original: $url");
    println(mysqli_escape_string($url), "Original: $url");
}