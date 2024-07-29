<?php
require_once dirname(__DIR__) . '/autoload.php';

class ToBeRequired {
    public function printMe() {
        println(__FILE__, "When it's required, I'm " . __CLASS__);
    }
}

