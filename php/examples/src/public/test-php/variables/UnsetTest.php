<?php
require_once dirname(__DIR__) . '/autoload.php';

class A {
    function __destruct() {
        echo "cYa later!!\n";
    }
}

$a = new A();
$a -> a = $a;
#unset($a); # Just uncomment, and you'll see

echo "No Message ... hm, what now?\n";
unset($a -> a);
echo "Tried to unset a of object\n";
unset($a);
echo "Finally that thing is gone\n";