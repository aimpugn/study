# callback

## closure

### 클로저란?

클로저(Closure)는 익명 함수로, 자신이 정의된 환경(스코프)의 변수를 캡처하고 이를 함수 내부에서 사용할 수 있는 기능을 갖습니다.
클로저는 변수를 캡처함으로써 상태를 유지할 수 있습니다.
이는 클로저가 *정의된 시점의 변수를 기억하고, 나중에 실행될 때도 그 변수를 사용*할 수 있게 합니다.

PHP에서 클로저는 `Closure` 클래스의 인스턴스로, 익명 함수와 동일합니다.
클로저는 다음과 같은 특징을 가집니다:

- **익명 함수**: 이름이 없는 함수입니다.
- **변수 캡처**: 클로저는 자신이 정의된 스코프의 변수를 캡처하여 사용할 수 있습니다.
- **`use` 키워드**: 클로저가 외부 변수를 사용할 수 있도록 합니다.

```php
$message = "Hello, World!";

// `$greet` 클로저는 `$message` 변수를 캡처하여 나중에 호출될 때 사용
$greet = function() use ($message) {
    echo $message;
};

$greet(); // Output: Hello, World!
```

### PHP의 콜백과 클로저

- **PHP에서 콜백**

    PHP에서 콜백은 호출 가능한 모든 것을 의미합니다
    이는 함수 이름, 객체 메서드, 그리고 클로저를 포함합니다.

- **클로저**:

    익명 함수로, `Closure` 클래스의 인스턴스입니다.
    클로저는 콜백의 한 형태입니다.

```php
function callFunc1(Closure $closure) {
    $closure();
}

function callFunc2(callable $callback) {
    $callback();
}

$function = function() {
    echo 'Hello, World!';
};

callFunc1($function); // Hello, World!
callFunc2($function); // Hello, World!
```

위 예제에서 `callFunc1`은 `Closure` 타입의 인자를 받으며, `callFunc2`는 `callable` 타입의 인자를 받습니다.
`callable`은 클로저뿐만 아니라 일반 함수나 객체 메서드도 받을 수 있습니다.

따라서 PHP의 콜백은 클로저를 포함하며, 클로저는 콜백의 한 형태로 볼 수 있습니다.

### `use` 키워드

`function($param) use ($expectedSubset)` 구문에서 `use ($expectedSubset)`는 익명 함수(클로저) 내부에서 외부 변수 `$expectedSubset`을 사용할 수 있도록 하는 역할을 합니다.
PHP의 클로저는 *기본적으로 자신이 정의된 스코프 외부의 변수를 직접 접근할 수 없기* 때문에, `use` 키워드를 사용하여 외부 변수를 클로저 내부로 가져와야 합니다.

만약 `use`를 사용하지 않고 바로 `$expectedSubset`을 사용하려고 하면, 클로저 내부에서 해당 변수를 인식하지 못해 에러가 발생합니다. 예를 들어, 다음과 같은 코드는 작동하지 않습니다:

```php
$expectedSubset = [
    'A' => $data['A'],
    'B' => $data['B'],
    'C' => $data['C'],
    'D' => $data['D'],
    'E' => $data['E'],
    'F' => $data['F'],
];

$mockControler->MyModel
    ->expects($this->once())
    ->method('set')
    ->with($this->callback(function($param) {
        foreach ($expectedSubset as $key => $value) { // 에러 발생
            if (!array_key_exists($key, $param) || $param[$key] !== $value) {
                return false;
            }
        }
        return true;
    }))
    ->will($this->returnValue(null));
```

위 코드에서 `$expectedSubset` 변수를 클로저 내부에서 접근하려고 하면, 해당 변수가 정의되지 않았다는 에러가 발생합니다. 따라서 `use` 키워드를 사용하여 외부 변수를 클로저 내부로 가져와야 합니다.

올바른 코드는 다음과 같습니다:

```php
$expectedSubset = [
    'A' => $data['A'],
    'B' => $data['B'],
    'C' => $data['C'],
    'D' => $data['D'],
    'E' => $data['E'],
    'F' => $data['F'],
];

$mockControler->MyModel
    ->expects($this->once())
    ->method('set')
    ->with($this->callback(function($param) use ($expectedSubset) {
        foreach ($expectedSubset as $key => $value) {
            if (!array_key_exists($key, $param) || $param[$key] !== $value) {
                return false;
            }
        }
        return true;
    }))
    ->will($this->returnValue(null));
```

이렇게 하면 클로저 내부에서 `$expectedSubset` 변수를 사용할 수 있게 되어, 원하는 검증을 수행할 수 있습니다.
