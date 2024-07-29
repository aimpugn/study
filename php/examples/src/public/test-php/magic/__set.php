<?php
require_once dirname(__DIR__) . '/autoload.php';

trait __Set {

    public function __set($name, $value)
    {
        $method = 'set' . ucwords($name);
        println($method, 'method');
        if (method_exists($this, $method)) {
            $this->{$method}($value);
        }
    }

    public function test($instance, $arr)
    {
        foreach ($arr as $val) {
            $instance->__set($val, $val);
        }
    }
}

class Caller {
    use __Set;

    public function call()
    {
        $callee = new Callee();
        (new Caller)->test($callee, ['var1', 'var2', 'var3']);
        return $callee;
    }
}

class Callee {
    private $var1;
    private $var2;

    public function setVar1($var1)
    {
        $this->var1 = $var1;
    }
    public function setVar2($var2)
    {
        $this->var2 = $var2;
    }
}

$tmp = new Caller();
println($tmp->call());
