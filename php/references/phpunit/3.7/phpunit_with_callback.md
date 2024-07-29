# PHPUnit with callback

- [PHPUnit with callback](#phpunit-with-callback)
    - [`->with($this->callback(...))` 경우 콜백 내에서 불일치 내용을 상세히 출력](#-withthis-callback-경우-콜백-내에서-불일치-내용을-상세히-출력)
        - [1. PHPUnit의 assert 메서드 활용](#1-phpunit의-assert-메서드-활용)
        - [2. 예외 던지기](#2-예외-던지기)
        - [3. 로깅 활용](#3-로깅-활용)
        - [4. 커스텀 어서션 메서드 생성](#4-커스텀-어서션-메서드-생성)

## `->with($this->callback(...))` 경우 콜백 내에서 불일치 내용을 상세히 출력

PHPUnit 3.7.28 버전에서는 직접적으로 불일치 내용을 상세히 출력하는 내장 기능이 없습니다.
하지만 테스트의 가독성과 디버깅을 위해 몇 가지 방법을 제안할 수 있습니다:

### 1. PHPUnit의 assert 메서드 활용

```php
->method('someMethod')
->with(
    $this->callback(function ($param) {
        $this->assertEquals(
            $param['value'], 
            $param['another'], 
            "Value {$param['value']} does not match {$param['another']}"
        );
        return true;
    })
)
```

이 방법을 사용하면 테스트 실패 시 PHPUnit이 불일치 내용을 출력합니다.

### 2. 예외 던지기

```php
->method('someMethod')
->with(
    $this->callback(function ($param) {
        if ($param['value'] !== $param['another']) {
            throw new PHPUnit_Framework_ExpectationFailedException(
                "Value {$param['value']} does not match {$param['another']}"
            );
        }
        return true;
    })
)
```

이 방법은 불일치 시 예외를 발생시켜 상세 정보를 제공합니다.

### 3. 로깅 활용

```php
->method('someMethod')
->with(
    $this->callback(function ($param) use ($testCase) {
        if ($param['value'] !== $param['another']) {
            $testCase->getTestResultObject()->addError(
                $testCase,
                new Exception("Value {$param['value']} does not match {$param['another']}"),
                0
            );
            return false;
        }
        return true;
    })
)
```

이 방법은 테스트 결과 객체에 에러를 추가하여 상세 정보를 로깅합니다.

### 4. 커스텀 어서션 메서드 생성

```php
public function assertArrayKeysMatch($expected, $actual, $message = '') {
    foreach ($expected as $key => $value) {
        if (!isset($actual[$key]) || $actual[$key] !== $value) {
            $this->fail(
                $message ?: "Key '$key' mismatch: expected '$value', got '" 
                    . ($actual[$key] ?? 'undefined') 
                    . "'"
            );
        }
    }
    return true;
}

// 사용
->method('someMethod')
->with(
    $this->callback(function ($param) {
        return $this->assertArrayKeysMatch(
            ['value' => $param['value']], 
            ['value' => $param['another']]
        );
    })
)
```

이 방법은 재사용 가능한 커스텀 어서션을 만들어 상세한 불일치 정보를 제공합니다.
