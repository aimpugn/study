# [functions](https://www.php.net/manual/en/language.functions.php)

- [functions](#functions)
    - [function arguments](#function-arguments)
        - [passing arguments by value, pass by value (default)](#passing-arguments-by-value-pass-by-value-default)

## [function arguments](https://www.php.net/manual/en/functions.arguments.php#functions.arguments)

> PHP supports
> - *passing arguments by value* (the default)
> - [passing by reference](https://www.php.net/manual/en/functions.arguments.php#functions.arguments.by-reference)
> - [default argument values](https://www.php.net/manual/en/functions.arguments.php#functions.arguments.default)
> - [Variable-length argument lists](https://www.php.net/manual/en/functions.arguments.php#functions.variable-arg-list)
> - [Named Arguments](https://www.php.net/manual/en/functions.arguments.php#functions.named-arguments)

### passing arguments by value, pass by value (default)

PHP에서 기본적으로 인자 전달 방식은 "pass by value"입니다.
이는 함수에 변수를 전달할 때 변수의 값이 복사되어 전달된다는 것을 의미합니다.

하지만 PHP에서는 *복사된 변수가 원래 변수와 동일한 데이터를 가리키는 메커니즘*을 가지고 있기 때문에, 큰 데이터 구조를 함수로 전달할 때 실제로 전체 데이터를 복사하지 않고도 효율적으로 전달할 수 있습니다.

이를 설명하기 위해 PHP의 배열 동작 방식을 이해해야 합니다.
PHP에서는 배열이 실제로는 내부적으로 `copy-on-write`(`COW`) 메커니즘을 사용합니다.
즉, *배열을 함수에 전달할 때는 배열의 참조(포인터)가 복사*됩니다.
이 참조는 새로운 배열을 생성하지 않으며, 원래 배열을 참조합니다.
실제로 배열이 변경되기 전까지는 데이터의 실제 복사가 일어나지 않습니다.
이는 성능을 향상시키기 위한 최적화입니다.

다음은 예시 코드입니다:

```php
function callFunction($array) {
    // 이 함수 안에서 $array를 변경하지 않는 한 원래 배열과 동일한 데이터를 참조합니다.
    // $array가 여기서 변경되면 그 시점에서 실제 복사가 일어납니다.
    echo $array['Model1']['prop1']; // 'aaaaaaaaaaaaaaaaaaa' 출력
}

$largeAssociativeArray = [
    'Model1' => [
        'prop1' => 'aaaaaaaaaaaaaaaaaaa',
        'prop2' => 'bbbbbbbbbb',
        'prop3' => 'cccccccccccc',
        'prop4' => 'ddd',
        'prop5' => 'eeeeeee',
        'prop6' => 'vvvvvvvvvvvvvvvv',
        'prop7' => 'xxxx',
    ],
    'Model2' => [
        'prop1' => 'aaaaaaaaaaaaaaaaaaa',
        'prop2' => 'bbbbbbbbbb',
        'prop3' => 'cccccccccccc',
        'prop4' => 'ddd',
        'prop5' => 'eeeeeee',
        'prop6' => 'vvvvvvvvvvvvvvvv',
        'prop7' => 'xxxx',
    ],
    // ... more and more ...
];

callFunction($largeAssociativeArray);
```

위 코드에서 `callFunction` 함수는 `$largeAssociativeArray`를 인자로 받습니다.
PHP는 이 배열을 참조로 전달하며, 배열의 데이터를 변경하지 않는 한 실제 데이터의 복사가 발생하지 않습니다.

그러나 함수 내에서 배열의 데이터를 변경하는 순간 실제 데이터의 복사가 발생하여 변경된 배열이 독립적으로 존재하게 됩니다.

따라서, PHP에서 배열을 함수에 전달할 때 기본적으로 배열 전체가 복사되지 않고 참조가 전달되기 때문에 성능상의 부담이 줄어듭니다.
이는 PHP의 효율적인 메모리 관리 전략 중 하나입니다.
