<?php
error_reporting(E_ALL);
$path = __DIR__ . '/';
$className = 'ToBeRequired';

// 그런데 이러면, 클래스 인스턴스가 호출한 클래스 필드에 설정이 안 된다. 왜????
if(file_exists($path) && (require_once $path) === 1 && class_exists($className)) {
    println(__DIR__, '__DIR__');
    println(__FILE__, '__FILE__');
    println("$className class exists!!!!!!");
    $instance = new $className();
    println(var_export(is_callable([$instance, 'printMe']), true), "is_callable([instance, 'printMe'])");
    println(var_export(is_callable([$instance, 'printMe'], true), true), "is_callable([instance, 'printMe'], true)");
    println(var_export(is_callable(['printMe']), true), "is_callable(['printMe'])");
    $instance->printMe();
    println(var_export($instance, true), '$instance');
}


// impossible to catch require_once
//try {
//    $require_once2 = require_once(__DIR__ . '/' . 'NotExists.php');
//    println(var_export($require_once2, true), '$require_once2');
//    println(gettype($require_once2), 'type of $require_once2');
//} catch(Exception $exception) {
//    println(gettype($exception), 'type of $exception');
//    println($exception->getMessage(), 'message of $exception');
//}