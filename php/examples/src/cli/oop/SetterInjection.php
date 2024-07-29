<?php

class Container
{
    public $property1;
    public $property2;
}

class Property1
{
    public $value = __CLASS__;
}

class Property2
{
    public $value = __CLASS__;
}

$prop1 = new Property1();
$prop2 = new Property2();
$container = new Container();

$container->property1 = $prop1;
$container->property2 = $prop2;

print_r('===============================================' . PHP_EOL);
print_r(json_encode($container, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_PRETTY_PRINT) . PHP_EOL);

$prop1->value = 'prop1 value is now changed';
$prop2->value = 'prop2 value is now changed';
print_r('===============================================' . PHP_EOL);
print_r(json_encode($container, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_PRETTY_PRINT) . PHP_EOL);

// 출력:
// ===============================================
// {
//     "property1": {
//         "value": "Property1"
//     },
//     "property2": {
//         "value": "Property2"
//     }
// }
// ===============================================
// {
//     "property1": {
//         "value": "prop1 value is now changed"
//     },
//     "property2": {
//         "value": "prop2 value is now changed"
//     }
// }
