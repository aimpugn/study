<?php
require_once dirname(__DIR__) . '/autoload.php';

class Test {
    var $field1;
    var $field2;
    var $field3;
}


/**
 * @param Test $test
 * @return void
 */
function callByRef($test, $fieldName, $fieldValue)
{
    if (property_exists($test, $fieldName) && ! is_null($fieldValue)) {
        $test->{$fieldName} = $fieldValue;
    }
}

$test1 = new Test();
println($test1, 'when created');
callByRef($test1, 'field1', 'value1');
println($test1, 'after callBy field1 and value1');
