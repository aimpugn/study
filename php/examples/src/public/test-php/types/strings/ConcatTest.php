<?php
require_once dirname(__DIR__) . '/autoload.php';

$toggle = true;
if ($toggle) {
    $message .= "test";
} else {
    $message .= "test2";
}
println($message, '$message');