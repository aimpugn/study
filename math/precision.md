# precision

## 반올림하지 않고 정밀도를 조절하는 방법

### 1. 문자열 조작 방법 (제시하신 방법)

```php
public static function getAmountWithPrecision($amount, $precision = 2)
{
    $formatted = number_format($amount, $precision + 1, '.', '');
    return floatval(substr_replace($formatted, '', -1, 1));
}
```

장점:
- 간단하고 직관적인 방법
- PHP의 내장 함수를 사용하여 구현이 쉬움

단점:
- 문자열 조작을 통한 방법이므로 대규모 데이터 처리 시 성능 이슈가 있을 수 있음
- floatval()을 사용하여 다시 float로 변환하는 과정에서 정밀도 손실 가능성 있음

#### floatval 사용시 정밀도 손실

floatval() 함수는 문자열을 부동소수점 수로 변환합니다.
이 과정에서 다음과 같은 문제가 발생할 수 있습니다:

```php
$str = "123.456789012345";
$float = floatval($str);
echo $float; // 출력: 123.45678901234
```

여기서 마지막 몇 자리가 정확하지 않게 표현됩니다.
이는 IEEE 754 표준에 따른 부동소수점 표현의 한계 때문입니다.

더 극단적인 예:

```php
$str = "0.1000000000000006";
$float = floatval($str);
echo $float; // 출력: 0.1
```

이 경우, 매우 작은 차이가 완전히 무시됩니다.

### 2. 수학적 접근 방법

```php
public static function getAmountWithPrecision($amount, $precision = 2)
{
    $multiplier = pow(10, $precision);
    return floor($amount * $multiplier) / $multiplier;
}
```

장점:
- 순수한 수학적 연산을 사용하므로 이해하기 쉬움
- 문자열 조작 없이 수치 연산만으로 처리

단점:
- 부동소수점 연산의 정밀도 문제가 발생할 수 있음

#### 부동소수점 연산의 정밀도 문제

부동소수점 연산은 이진 분수를 사용하여 수를 표현하기 때문에, 십진 분수를 정확히 표현하지 못하는 경우가 있습니다.

예를 들어:

```php
$a = 0.1;
$b = 0.2;
$c = $a + $b;
echo $c; // 출력: 0.3
echo $c == 0.3 ? 'true' : 'false'; // 출력: false
printf('%.20f', $c); // 출력: 0.30000000000000004441
```

여기서 0.1과 0.2의 합이 정확히 0.3이 되지 않습니다.
이는 0.1과 0.2가 이진 부동소수점으로 정확히 표현될 수 없기 때문입니다.

다른 예:

```php
$large = 10000000000000000;
$small = 0.00000000000000001;
$result = $large + $small;
echo $result == $large; // 출력: true
```

이 경우, 매우 큰 수에 매우 작은 수를 더했을 때, 작은 수가 완전히 무시됩니다.

### 3. BCMath 라이브러리 사용 (정밀한 십진 연산)

```php
public static function getAmountWithPrecision($amount, $precision = 2)
{
    return bcadd($amount, '0', $precision);
}
```

장점:
- 매우 정확한 십진 연산 가능
- 대규모의 금융 계산에 적합

단점:
- BCMath 확장이 필요함
- 큰 숫자나 많은 소수점을 다룰 때 성능이 약간 떨어질 수 있음

### 4. 정수로 변환 후 처리

```php
public static function getAmountWithPrecision($amount, $precision = 2)
{
    $multiplier = pow(10, $precision);
    $intValue = (int)($amount * $multiplier);
    return $intValue / $multiplier;
}
```

장점:
- 정수 연산을 사용하여 부동소수점 오차 최소화
- 비교적 빠른 성능

단점:
- 매우 큰 숫자나 높은 정밀도에서는 정수 오버플로우 가능성 있음

## 최적의 방법 선택

1. 일반적인 용도: 제시하신 문자열 조작 방법이나 수학적 접근 방법이 대부분의 경우에 충분히 좋습니다.

2. 고정밀 금융 계산: BCMath 라이브러리를 사용하는 방법이 가장 정확하고 안전합니다.

3. 성능이 중요한 경우: 정수로 변환 후 처리하는 방법이 좋은 선택일 수 있습니다.

4. 최적의 절충안:

```php
public static function getAmountWithPrecision($amount, $precision = 2)
{
    if (function_exists('bcadd')) {
        return bcadd($amount, '0', $precision);
    } else {
        $multiplier = pow(10, $precision);
        return floor($amount * $multiplier) / $multiplier;
    }
}
```

이 방법은 BCMath가 사용 가능할 경우 가장 정확한 결과를 제공하고, 그렇지 않은 경우 수학적 접근 방법을 사용합니다.
이는 정확성과 호환성, 성능 사이의 좋은 균형을 제공합니다.

결론적으로, 사용 환경과 요구사항(정확성, 성능, 호환성)을 고려하여 가장 적합한 방법을 선택해야 합니다. 금융 어플리케이션의 경우, 가능하다면 BCMath를 사용하는 것이 가장 안전하고 정확한 방법일 것입니다.

## 정밀도 문제

1. floatval() 사용 시 정밀도 손실 문제:

2. 부동소수점 연산의 정밀도 문제:

3. 반복적인 연산에서의 오차 누적:

```php
$sum = 0;
for ($i = 0; $i < 1000; $i++) {
    $sum += 0.1;
}
echo $sum; // 출력: 99.999999999998
```

이 예에서, 0.1을 1000번 더했을 때 정확히 100이 되지 않습니다. 작은 오차가 누적되어 결과에 영향을 미칩니다.

4. 큰 숫자와 작은 숫자의 연산:

```php
$large = 1e15;
$small = 1;
$result = ($large + $small) - $large;
echo $result; // 출력: 0 (예상값: 1)
```

이 경우, 큰 숫자에 작은 숫자를 더한 후 다시 큰 숫자를 빼면, 작은 숫자의 정보가 완전히 손실됩니다.

이러한 문제들은 금융 계산이나 과학적 계산에서 심각한 오류를 야기할 수 있습니다. 따라서 높은 정밀도가 요구되는 상황에서는 다음과 같은 방법을 고려해야 합니다:

1. BCMath 라이브러리 사용
2. 정수로 변환하여 계산 후 다시 변환
3. Decimal 타입을 지원하는 특수 라이브러리 사용
4. 가능한 경우 분수 연산 사용

이러한 방법들은 부동소수점 연산의 한계를 극복하고 더 정확한 결과를 얻을 수 있게 해줍니다.
