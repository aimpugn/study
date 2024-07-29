# Atomic Types

- [Atomic Types](#atomic-types)
    - [Atomic Types?](#atomic-types-1)
    - [Scalar types](#scalar-types)
    - [Object types](#object-types)
    - [Array Types](#array-types)
        - [Callable arrays](#callable-arrays)
        - [non-empty-array](#non-empty-array)

## [Atomic Types](https://psalm.dev/docs/annotating_code/type_syntax/atomic_types/)?

## Scalar types

`int`, `bool`, `float`, `string`은 스칼라 타입의 예입니다.
스칼라 타입은 PHP에서 스칼라 값을 나타냅니다.

## Object types

## [Array Types](https://psalm.dev/docs/annotating_code/type_syntax/array_types/)

- PHP에서 `array` 타입은 주로 3 가지 다른 자료 구조를 나타낸다
- PHP는 기본적으로 이러한 배열을 모두 동일하게 처리한다(첫 번째 경우에 대한 몇 가지 최적화 기능이 있지만).

```php
# List: https://en.wikipedia.org/wiki/List_(abstract_data_type)
$a = [1, 2, 3, 4, 5];

# Associative array: https://en.wikipedia.org/wiki/Associative_array
$b = [0 => 'hello', 5 => 'goodbye'];
$c = ['a' => 'AA', 'b' => 'BB', 'c' => 'CC'];

# Makeshift Structs: https://en.wikipedia.org/wiki/Struct_(C_programming_language)
$d = ['name' => 'Psalm', 'type' => 'tool'];
```

```php
<?php
["hello", "world", "foo" => new stdClass, 28 => false];
```

- 위의 코드는 psalm 내부적으로 아래처럼 처리 된다

    ```php
    array{0: string, 1: string, foo: stdClass, 28: false}
    ```

- 포맷을 직접 지정할 수 있다

    ```php
    /** @return array{foo: string, bar: int} */
    ```

- optional 키는 뒤에 오는 `?`로 표시할 수 있다

    ```php
    /** @return array{optional?: string, bar: int} */
    ```

- 한 줄 주석도 사용이 가능

    ```php
    /** @return array { // Array with comments.
     *     // Comments can be placed on their own line. 
     *     foo: string, // An array key description.
     *     bar: array {, // Another array key description.
     *         'foo//bar': string, // Array key with "//" in it's name.
     *     },
     * }
     */
    ```

### [Callable arrays](https://psalm.dev/docs/annotating_code/type_syntax/array_types/#callable-arrays)

- An array holding a callable, like PHP's native `call_user_func()` and friends supports it:

```php
<?php

$callable = ['myClass', 'aMethod'];
$callable = [$object, 'aMethod'];
```

### [non-empty-array](https://psalm.dev/docs/annotating_code/type_syntax/array_types/#non-empty-array)

- 비워둘 수 없는 배열
- [generic syntax](https://psalm.dev/docs/annotating_code/type_syntax/array_types/#generic-arrays)도 지원된다

```php
non-empty-array<string, int>
```
