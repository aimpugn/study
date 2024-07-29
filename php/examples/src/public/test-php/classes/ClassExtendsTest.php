<?php
require_once dirname(__DIR__) . '/autoload.php';

class TestParent {
    private $field1 = '$field1';
    protected $field2 = '$field2';

    public function getFieldFromParent($fieldName)
    {
        return $this->{$fieldName};
    }
}

class TestChild extends TestParent {

    public function __construct($f1, $f2)
    {
        $this->field1 = $f1;
        $this->field2 = $f2;
    }

    public function getFieldFromChild($fieldName)
    {
        return $this->{$fieldName};
    }
}

$child = new TestChild('f1', 'f2');
println($child->getFieldFromChild('field1'), 'getFieldFromChild field1');
println($child->getFieldFromChild('field2'), 'getFieldFromChild field2');
println($child->getFieldFromParent('field1'), 'getFieldFromParent field1');
println($child->getFieldFromParent('field2'), 'getFieldFromParent field2');

println(json_encode($child));
