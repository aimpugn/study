# RiskyTruthyFalsyComparison

## `RiskyTruthyFalsyComparison`?

> `RiskyTruthyFalsyComparison`: `null|string` 타입의 피연산자는 string 타입을 포함하고 있으며, 이는 참(truthy) 또는 거짓(falsy)으로 평가될 수 있습니다. 이는 예상치 못한 동작을 초래할 수 있습니다. 엄격한 비교를 사용하십시오.

Psalm에서 발생하는 `RiskyTruthyFalsyComparison` 경고는 `null|string` 타입의 피연산자가 `string` 타입을 포함하고 있으며, 이는 참(truthy) 또는 거짓(falsy)으로 평가될 수 있기 때문에 발생합니다.
이는 예상치 못한 동작을 초래할 수 있으므로 엄격한 비교를 사용하라는 경고입니다.

## 설명

PHP에서 빈 문자열(`""`)은 `false`로 평가되기 때문에, `null`과 빈 문자열을 구분하지 않고 비교하면 의도하지 않은 결과를 초래할 수 있습니다.

예를 들어, `null`과 빈 문자열 모두 `false`로 평가되므로, `if ($value)`와 같은 조건문에서 두 값이 동일하게 처리됩니다.
이는 코드의 의도를 명확히 하고 버그를 방지하기 위해 엄격한 비교(`===` 또는 `!==`)를 사용해야 하는 이유입니다.

```php
/**
 * @param null|string $value
 * @return bool
 */
function isValueSet($value) {
    return $value; // 이 부분에서 경고 발생
}
```

위 코드에서 `$value`는 `null` 또는 `string` 타입을 가질 수 있습니다.
PHP에서 빈 문자열(`""`)은 `false`로 평가되며, 이는 예상치 못한 동작을 초래할 수 있습니다.
따라서 엄격한 비교를 사용하여 이러한 문제를 방지해야 합니다.

## 코드 수정 예시

다음은 경고를 해결하기 위해 엄격한 비교를 사용하는 예시입니다:

```php
/**
 * @param null|string $value
 * @return bool
 */
function isValueSet($value) {
    return $value !== null && $value !== ''; // 엄격한 비교 사용
}
```

이렇게 하면 `null`과 빈 문자열을 명확히 구분할 수 있으며, 예상치 못한 동작을 방지할 수 있습니다.
