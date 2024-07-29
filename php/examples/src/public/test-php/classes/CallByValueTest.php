<?php
require_once dirname(__DIR__) . '/autoload.php';

class TestSettee {
    public $field;
}

class TestSetter {
    /**
     * @param TestSettee $obj
     * @param mixed $value
     * @return void
     */
    public function setValue(TestSettee $obj, $value)
    {
        $obj->field = $value;
    }
}

$settee = new TestSettee();
$setter = new TestSetter();

$setter->setValue($settee, 'testValue1');
println($settee, 'when set testValue1');
$setter->setValue($settee, ['key1' => 'val1']);
println($settee, 'when set assoc array');




