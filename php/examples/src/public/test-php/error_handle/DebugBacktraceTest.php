<?php
require_once dirname(__DIR__) . '/autoload.php';
error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

$debugBacktrace1 = null;
$debugBacktrace2 = null;
$debugBacktrace3 = null;

/**
 * @throws Exception
 */
function func1(): string
{
    return func2();
}

/**
 * @throws Exception
 */
function func2($arg1 = 'default'): string
{
    return func3('test1', 'test2');
}

function func3($arg1 = 'default', $arg2 = null): string
{
    global $debugBacktrace1;
    global $debugBacktrace2;

    $debugBacktrace1 =debug_backtrace(DEBUG_BACKTRACE_PROVIDE_OBJECT);
    $debugBacktrace2 = debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS);
    $debugBacktrace2 = debug_backtrace(DEBUG_BACKTRACE_PROVIDE_OBJECT | DEBUG_BACKTRACE_IGNORE_ARGS);
    return "$arg1 and $arg2";
}

try {
    $result = func1();
} catch(Exception $e) {
    println($e);
}

println($debugBacktrace1, 'DEBUG_BACKTRACE_PROVIDE_OBJECT');
println($debugBacktrace2, 'DEBUG_BACKTRACE_IGNORE_ARGS');
