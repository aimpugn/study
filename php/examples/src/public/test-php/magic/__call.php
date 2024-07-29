<?php
require_once dirname(__DIR__) . '/autoload.php';

class __Call {

    public function __call($method, $args)
    {
        println(print_r([
            'method' => $method,
            'args' => $args
        ], true), 'when __call');
    }
}

$tmp = new __Call('when construct', [0, 1, 2]);
/*
[when __call] Array
(
    [method] => test
    [args] => Array
        (
            [0] => test is not exists
            [1] => Array
                (
                    [0] => 3
                    [1] => 4
                    [2] => 5
                )

        )

)
 */
$tmp->test('test is not exists', [3, 4, 5]);
