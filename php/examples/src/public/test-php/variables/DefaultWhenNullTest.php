<?php
require_once dirname(__DIR__) . '/autoload.php';

$var1 = ['key2' => 'test'];

$var2 = $var1['key'] ?: 1;

println($var2, '$var2');

$var3 = $var1['key2'] ?: 1;

println($var3, '$var3');
