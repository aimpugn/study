# PHPUnit - match

- [PHPUnit - match](#phpunit---match)
    - [`with`](#with)
        - [중복되는 매개변수 검증 로직 공통화](#중복되는-매개변수-검증-로직-공통화)

## `with`

### 중복되는 매개변수 검증 로직 공통화

```php
->expects($this->once())
->method('someMethod')
->with(
    $this->callback(function ($param) {
        return !empty($param);
    }),
    $this->callback(function ($param) {
        return !empty($param);
    }),
    $this->callback(function ($param) use ($additional) {
        return $param === $additional;
    }),
    $this->callback(function ($param) {
        return !empty($param);
    }),
    $this->callback(function ($param) {
        return !empty($param);
    })
)
->will($this->returnValue($expected));
```

이 코드는 PHPUnit의 목(mock) 객체를 사용한 테스트 코드로 보입니다.
여러 매개변수에 대해 반복적으로 유사한 검증을 수행하고 있습니다.
이를 개선하기 위한 몇 가지 방법을 제안하겠습니다.

1. 공통 콜백 함수 추출

    먼저, 반복되는 검증 로직을 별도의 함수로 추출할 수 있습니다.

    ```php
    private function isNotEmpty($param)
    {
        return !empty($param);
    }

    private function isEqualTo($expected)
    {
        return function ($param) use ($expected) {
            return $param === $expected;
        };
    }
    ```

    간단하고 직관적이지만, 여전히 약간의 중복이 있습니다.

2. 배열을 사용한 콜백 정의

    그리고 이를 배열을 사용하여 간결하게 표현할 수 있습니다.

    ```php
    $callbacks = [
        [$this, 'isNotEmpty'],
        [$this, 'isNotEmpty'],
        $this->isEqualTo($additional),
        [$this, 'isNotEmpty'],
        [$this, 'isNotEmpty']
    ];

    $callbackArgs = array_map(function ($callback) {
        return $this->callback($callback);
    }, $callbacks);

    ->expects($this->once())
    ->method('someMethod')
    ->with(...$callbackArgs)
    ->will($this->returnValue($expected));
    ```

    중복을 줄이고 유연성을 제공하지만, 복잡성이 증가합니다.

3. 데이터 구조를 사용한 검증 로직 정의

    더 나아가, 각 매개변수에 대한 검증 로직을 데이터 구조로 정의할 수 있습니다.

    ```php
    private function validateParam($rule, $param)
    {
        switch ($rule['type']) {
            case 'notEmpty':
                return !empty($param);
            case 'equals':
                return $param === $rule['value'];
            // 필요에 따라 더 많은 규칙 추가
        }
    }

    // 테스트 메서드 내부
    $paramRules = [
        ['type' => 'notEmpty'],
        ['type' => 'notEmpty'],
        ['type' => 'equals', 'value' => $additional],
        ['type' => 'notEmpty'],
        ['type' => 'notEmpty']
    ];

    $callbacks = array_map(function ($rule) {
        return $this->callback(function ($param) use ($rule) {
            return $this->validateParam($rule, $param);
        });
    }, $paramRules);

    ->expects($this->once())
    ->method('someMethod')
    ->with(...$callbacks)
    ->will($this->returnValue($expected));
    ```

    매우 유연하고 확장 가능하지만, 초기 설정이 복잡할 수 있습니다.

4. 커스텀 제약 클래스 생성

    PHPUnit의 `Constraint` 클래스를 확장하여 커스텀 제약을 만들 수 있습니다.

    ```php
    use PHPUnit\Framework\Constraint\Constraint;

    class NotEmptyConstraint extends Constraint
    {
        public function matches($other): bool
        {
            return !empty($other);
        }

        public function toString(): string
        {
            return 'is not empty';
        }
    }

    class EqualsConstraint extends Constraint
    {
        private $value;

        public function __construct($value)
        {
            $this->value = $value;
        }

        public function matches($other): bool
        {
            return $this->value === $other;
        }

        public function toString(): string
        {
            return 'is equal to ' . $this->exporter()->export($this->value);
        }
    }

    // 테스트 메서드 내부
    ->expects($this->once())
    ->method('someMethod')
    ->with(
        new NotEmptyConstraint(),
        new NotEmptyConstraint(),
        new EqualsConstraint($additional),
        new NotEmptyConstraint(),
        new NotEmptyConstraint()
    )
    ->will($this->returnValue($expected));
    ```

    가장 깔끔하고 PHPUnit의 철학에 부합하지만, 추가 클래스 생성이 필요합니다.

5. Custom Matcher 클래스 생성

    가장 깔끔하고 PHPUnit의 철학에 부합합니다.

    ```php
    class NotEmptyConstraint extends PHPUnit_Framework_Constraint
    {
        public function matches($other)
        {
            return !empty($other);
        }

        public function toString()
        {
            return 'is not empty';
        }
    }

    class EqualsConstraint extends PHPUnit_Framework_Constraint
    {
        protected $value;

        public function __construct($value)
        {
            $this->value = $value;
        }

        public function matches($other)
        {
            return $this->value === $other;
        }

        public function toString()
        {
            return 'is equal to ' . var_export($this->value, true);
        }
    }

    // 테스트 클래스 내부
    public function getNotEmpty()
    {
        return new NotEmptyConstraint();
    }

    public function getEquals($value)
    {
        return new EqualsConstraint($value);
    }

    // 테스트 메서드 내부
    $mock->expects($this->once())
        ->method('someMethod')
        ->with(
            $this->getNotEmpty(),
            $this->getNotEmpty(),
            $this->getEquals($additional),
            $this->getNotEmpty(),
            $this->getNotEmpty()
        )
        ->will($this->returnValue($expected));
    ```

    하지만, 추가 클래스 생성이 필요하고, 초기 설정이 복잡할 수 있습니다
