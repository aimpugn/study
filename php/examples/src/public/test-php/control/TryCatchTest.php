<?php
require_once dirname(__DIR__) . '/autoload.php';

$tmp1 = [
    'k1' => 'v1',
];

try {
    $tmp2 = [
        'k2' => 'v2',
    ];
    throw new Exception('Test');
} catch (Exception $e) {
    println($tmp1);
    println($tmp2);
}