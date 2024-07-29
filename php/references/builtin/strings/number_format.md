# number_format

## number_format

```php
number_format(
    float $num,
    int $decimals = 0,
    ?string $decimal_separator = ".",
    ?string $thousands_separator = ","
): string
```

### Decimal separator의 필요성

- Decimal separator는 수의 정수부와 소수부를 구분하는 기호입니다.
- 문화적 차이로 인해 다양한 decimal separator가 사용됩니다. (예: 미국은 점(.), 유럽 일부 국가는 쉼표(,))
- 프로그래밍에서 decimal separator를 지정할 수 있게 하는 이유:
  1. 국제화(i18n)와 지역화(l10n) 지원
  2. 다양한 형식의 입력 데이터 처리
  3. 사용자 정의 출력 형식 지원

## `strval` + `explode` vs `number_format`, 그리고 반올림

### `strval` + `explode` 방식

```php
$num = 1234.5678;
$parts = explode('.', strval($num));
$formatted = $parts[0] . '.' . substr($parts[1], 0, 2);
```

### number_format 방식

```php
$num = 1234.5678;
$formatted = number_format($num, 2, '.', ',');
```

### 차이점

1. 정밀도:
    - `number_format`: 내부적으로 반올림을 수행
    - `explode`: 방식은 단순 절삭
2. 천단위 구분:
    - `number_format`: 자동으로 처리
    - `explode`: 추가 작업이 필요
3. 유연성:
    - `explode`: 더 많은 제어를 제공하지만, 복잡성도 증가
4. 성능:
    - 일반적으로 `number_format`이 더 빠르고 최적화되어 있습니다.

### 금액 처리와 반올림

- 금액을 다룰 때 반올림(round up)은 상황에 따라 적절할 수 있지만, 주의가 필요합니다.
- 재무회계에서는 일반적으로 "bankers rounding" (반올림 5/4)을 사용합니다.
- 반올림은 누적 오차를 발생시킬 수 있으므로, 중요한 금융 계산에서는 정확한 십진 연산(예: BC Math)을 사용하는 것이 좋습니다.
