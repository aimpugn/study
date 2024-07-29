<?php
require_once dirname(__DIR__) . '/autoload.php';

$o1 = new stdClass();

println($o1->test);
println($o1->getTest());
