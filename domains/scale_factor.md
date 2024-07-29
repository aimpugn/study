# Scale Factor

- [Scale Factor](#scale-factor)
    - [Scale Factor란?](#scale-factor란)
    - [주요 통화의 Scale Factor](#주요-통화의-scale-factor)
    - [Scale Factor의 역할](#scale-factor의-역할)
    - [정확한 계산을 위한 방법](#정확한-계산을-위한-방법)
    - [`round` 함수의 사용](#round-함수의-사용)
    - [주의사항](#주의사항)
    - [부동소수점 오차 없도록 계산하기](#부동소수점-오차-없도록-계산하기)
        - [1. BC Math 라이브러리 사용](#1-bc-math-라이브러리-사용)
        - [2. 정수로 변환 후 계산](#2-정수로-변환-후-계산)
        - [3. GMP(GNU Multiple Precision) 라이브러리 사용](#3-gmpgnu-multiple-precision-라이브러리-사용)
        - [4. 문자열 조작과 정수 연산 조합](#4-문자열-조작과-정수-연산-조합)
        - [소수점 이하 둘째자리로 유지하기](#소수점-이하-둘째자리로-유지하기)

## Scale Factor란?

scale factor는 통화 단위의 정밀도를 나타내는 중요한 개념입니다.
Scale factor는 통화의 가장 작은 단위를 나타내는 데 사용되는 10의 거듭제곱 값입니다.
이는 통화의 소수점 이하 자릿수를 결정합니다.

## 주요 통화의 Scale Factor

- KRW (대한민국 원): 0 (1원이 가장 작은 단위)
- USD (미국 달러): 2 (1센트 = $0.01이 가장 작은 단위)
- JPY (일본 엔): 0 (1엔이 가장 작은 단위)
- EUR (유로): 2 (1센트 = €0.01이 가장 작은 단위)
- BTC (비트코인): 8 (1 사토시 = 0.00000001 BTC가 가장 작은 단위)

## Scale Factor의 역할

- 정확성: 통화 금액을 정확하게 표현하고 계산할 수 있게 해줍니다.
- 저장: 데이터베이스나 시스템에서 통화 금액을 효율적으로 저장할 수 있게 합니다.
- 연산: 서로 다른 통화 간의 변환이나 계산을 정확하게 수행할 수 있게 합니다.

## 정확한 계산을 위한 방법

a. 정수형으로 저장:
    금액을 가장 작은 단위로 변환하여 정수형으로 저장합니다.
    예: $10.50 → 1050 (센트 단위)

b. 계산 시 Scale Factor 고려:
    모든 계산은 정수형으로 수행한 후, 결과를 다시 원래 단위로 변환합니다.

c. 반올림 주의:
    나눗셈 등의 연산 후에는 적절한 반올림 처리가 필요합니다.

```php
class Currency {
    private $amount;
    private $scaleFactor;

    public function __construct($amount, $scaleFactor) {
        $this->amount = $amount * (10 ** $scaleFactor);
        $this->scaleFactor = $scaleFactor;
    }

    public function add(Currency $other) {
        if ($this->scaleFactor !== $other->scaleFactor) {
            throw new Exception("Cannot add currencies with different scale factors");
        }
        $this->amount += $other->amount;
    }

    public function getFormattedAmount() {
        return number_format($this->amount / (10 ** $this->scaleFactor), $this->scaleFactor);
    }
}

// 사용 예
$usd = new Currency(10.50, 2);  // $10.50
$krw = new Currency(1000, 0);   // ₩1000

echo $usd->getFormattedAmount();  // 출력: 10.50
echo $krw->getFormattedAmount();  // 출력: 1000
```

## `round` 함수의 사용

네, 말씀하신 내용은 정확합니다. PHP에서 `round(1.111, 3)`을 사용하면 소수점 셋째 자리에서 반올림하여 소수점 셋째 자리까지 표시합니다. 그러나 여기서 주의할 점이 있습니다:

1. `round()` 함수의 동작:
   - `round(1.111, 3)`의 결과는 1.111입니다.
   - `round(1.1115, 3)`의 결과는 1.112입니다.

2. `round()` vs `number_format()`:
   - `round()`: 숫자를 반올림하지만, 결과를 항상 지정된 소수점 자릿수로 표시하지는 않습니다.
   - `number_format()`: 숫자를 지정된 형식으로 포맷팅하며, 항상 지정된 소수점 자릿수를 표시합니다.

3. 사용 시 고려사항:
   - 정확성: `round()`는 수학적으로 정확한 반올림을 수행합니다.
   - 표시: `number_format()`은 결과를 항상 일관된 형식으로 표시합니다.

4. 예시:

    ```php
    $num = 1.111;

    echo round($num, 3);  // 출력: 1.111
    echo "\n";
    echo number_format($num, 3);  // 출력: 1.111

    $num = 1.0;

    echo round($num, 3);  // 출력: 1
    echo "\n";
    echo number_format($num, 3);  // 출력: 1.000
    ```

5. 금융 계산에서의 사용:
   금융 계산에서는 `round()`를 사용하는 것이 적절할 수 있습니다. 그 이유는:
   - 정확한 수학적 반올림을 수행합니다.
   - 불필요한 trailing zeros를 제거합니다.

   그러나 결과를 표시할 때는 `number_format()`을 사용하여 일관된 형식을 유지하는 것이 좋습니다.

6. 제안하는 방식:
   계산과 표시를 분리하는 것이 좋습니다:

```php
function calculateAmount($amount, $precision) {
    return round($amount, $precision);
}

function formatAmount($amount, $precision) {
    return number_format($amount, $precision);
}

$amount = 1.1115;
$calculatedAmount = calculateAmount($amount, 3);
echo "계산된 금액: " . $calculatedAmount . "\n"; // 계산된 금액: 1.112
echo "표시될 금액: " . formatAmount($calculatedAmount, 3); // 표시될 금액: 1.112
```

결론적으로, `round()`를 사용하여 계산하는 것은 괜찮습니다. 그러나 최종 결과를 사용자에게 표시하거나 저장할 때는 `number_format()`을 사용하여 일관된 형식을 유지하는 것이 좋습니다. 이렇게 하면 정확성과 일관성을 모두 유지할 수 있습니다.

## 주의사항

- 통화 변환: 서로 다른 scale factor를 가진 통화 간 변환 시 주의가 필요합니다.
- 부동소수점 오차: 정수형으로 저장하더라도 변환 과정에서 부동소수점 오차가 발생할 수 있으므로 주의해야 합니다.
- 반올림 정책: 금융 거래에서는 반올림 정책(올림, 내림, 반올림)이 중요하므로, 상황에 맞는 정책을 사용해야 합니다.

## 부동소수점 오차 없도록 계산하기

PHP 5.5와 5.6에서 USD나 EUR 같은 통화의 부동소수점 금액(예: 10.123)을 정확하게 처리하는 방법에 대해 설명드리겠습니다.
부동소수점 연산의 문제점을 피하고 정확한 계산을 수행하기 위한 여러 방법이 있습니다.

다음 소개할 방법들은 모두 부동소수점 표현의 한계를 우회하여 정확한 계산을 수행합니다.
*부동소수점 숫자*는 컴퓨터에서 *이진 분수로 표현*되므로 일부 십진 소수를 정확히 표현할 수 없습니다.
이 문제를 해결하기 위해 문자열 표현, 정수 연산, 또는 특수 라이브러리를 사용하여 정확한 십진 표현과 연산을 가능하게 합니다.

실제 금융 애플리케이션에서는 이러한 방법들을 사용하여 정확한 금액 계산을 보장합니다.
선택한 방법에 따라 성능과 정확성 사이의 균형을 고려해야 하며, 애플리케이션의 요구사항에 가장 적합한 방법을 선택해야 합니다.

### 1. BC Math 라이브러리 사용

BC Math 라이브러리는 임의 정밀도 수학 함수를 제공합니다.
이 라이브러리를 사용하면 정확한 소수점 계산이 가능합니다.

```php
$a = '10.123';
$b = '20.456';
$sum = bcadd($a, $b, 3);  // 30.579

echo $sum;
```

- BC Math는 문자열로 숫자를 처리하여 내부적으로 정확한 십진 표현을 유지합니다.
- 부동소수점 표현의 한계를 우회하여 정확한 십진 연산을 수행합니다.
- 원하는 정밀도를 지정할 수 있어 반올림 오류를 제어할 수 있습니다.

### 2. 정수로 변환 후 계산

금액을 센트 단위의 정수로 변환하여 계산한 후, 다시 달러 단위로 변환하는 방법입니다.

```php
function addMoney($a, $b) {
    $a_cents = (int)($a * 100);
    $b_cents = (int)($b * 100);
    $sum_cents = $a_cents + $b_cents;
    return $sum_cents / 100;
}

$result = addMoney(10.123, 20.456);
echo number_format($result, 3);  // 30.579
```

- 정수 연산은 부동소수점 연산보다 정확합니다.
- 센트 단위로 변환함으로써 소수점 이하의 정확성을 유지합니다.
- 최종 결과를 100으로 나눌 때 발생할 수 있는 오차는 매우 작습니다.

### 3. GMP(GNU Multiple Precision) 라이브러리 사용

GMP 확장을 사용하면 매우 큰 숫자나 높은 정밀도가 필요한 계산을 수행할 수 있습니다.

```php
$a = gmp_init('10123', 10);  // 10.123을 1000배한 값
$b = gmp_init('20456', 10);  // 20.456을 1000배한 값
$sum = gmp_add($a, $b);
$result = gmp_strval($sum) / 1000;

echo number_format($result, 3);  // 30.579
```

- GMP는 임의 정밀도 정수 연산을 지원합니다.
- 내부적으로 정확한 정수 표현을 사용하여 계산합니다.
- 스케일링(여기서는 1000배)을 통해 소수점 정확도를 유지합니다.

### 4. 문자열 조작과 정수 연산 조합

문자열로 숫자를 처리하고, 소수점을 제거한 후 정수 연산을 수행하는 방법입니다.

```php
function addPreciseFloat($a, $b) {
    $a_parts = explode('.', $a);
    $b_parts = explode('.', $b);
    
    $a_int = $a_parts[0] . str_pad(isset($a_parts[1]) ? $a_parts[1] : '', 3, '0', STR_PAD_RIGHT);
    $b_int = $b_parts[0] . str_pad(isset($b_parts[1]) ? $b_parts[1] : '', 3, '0', STR_PAD_RIGHT);
    
    $sum = $a_int + $b_int;
    return substr($sum, 0, -3) . '.' . substr($sum, -3);
}

$result = addPreciseFloat('10.123', '20.456');
echo $result;  // 30.579
```

- 문자열 조작을 통해 소수점 위치를 정확하게 제어합니다.
- 정수 연산을 사용하여 부동소수점 오차를 방지합니다.
- 결과를 문자열로 조작하여 정확한 소수점 표현을 유지합니다.

### 소수점 이하 둘째자리로 유지하기

```php
function processAmount($amount) {
    // 1. 문자열로 변환하여 부동소수점 오차 방지
    $amount = strval($amount);

    // 2. 소수점 위치 찾기
    $decimalPos = strpos($amount, '.');

    if ($decimalPos === false) {
        // 3. 소수점이 없는 경우, 소수점과 두 개의 0을 추가
        return $amount . '.00';
    } else {
        // 4. 소수점 이하 부분 추출
        $fractional = substr($amount, $decimalPos + 1);
        
        // 5. 소수점 이하 자릿수에 따른 처리
        if (strlen($fractional) <= 2) {
            // 6. 2자리 미만인 경우, 0으로 채움
            return str_pad($amount, strlen($amount) + (2 - strlen($fractional)), '0');
        } else {
            // 7. 3자리 이상인 경우, 반올림 처리
            $rounded = bcadd($amount, '0', 2);
            
            // 8. 결과가 정수인 경우 .00 추가
            return strpos($rounded, '.') === false ? $rounded . '.00' : $rounded;
        }
    }
}
```
