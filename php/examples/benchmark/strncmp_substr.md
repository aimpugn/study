# `strncmp` vs `substr`

- [`strncmp` vs `substr`](#strncmp-vs-substr)
    - [strncmp vs substr 비교: 거의 차이 없다](#strncmp-vs-substr-비교-거의-차이-없다)
    - [`strncmp` 사용](#strncmp-사용)
    - [`substr` 사용](#substr-사용)
    - [벤치마크](#벤치마크)

## strncmp vs substr 비교: 거의 차이 없다

`strncmp` 함수와 `substr` 함수를 사용한 문자열 비교의 효율성과 속도를 비교할 때, 일반적으로 `strncmp` 함수가 더 효율적이고 빠릅니다. 이유는 다음과 같습니다:

1. **메모리 할당**: `strncmp` 함수는 추가적인 메모리 할당 없이 직접 원본 문자열에서 비교를 수행합니다. 반면, `substr` 함수는 새로운 문자열을 생성하기 위해 메모리를 할당하고, 이후에 생성된 문자열을 대상 문자열과 비교합니다. 이 과정에서 발생하는 메모리 할당과 해제는 성능에 부담을 줄 수 있습니다.

2. **연산 과정**: `strncmp`는 지정된 길이만큼 문자열을 비교하고 즉시 결과를 반환합니다. 이는 비교 과정이 단순하고 빠르다는 것을 의미합니다. 반면, `substr`을 사용할 경우, 먼저 부분 문자열을 추출하고, 추출된 문자열을 다시 비교하는 두 단계의 연산이 필요합니다. 이는 `strncmp`에 비해 추가적인 연산 단계를 포함합니다.

3. **함수의 목적**: `strncmp`는 문자열의 특정 부분만을 비교하는 것이 목적이므로, 이러한 비교가 필요할 때 최적화되어 있습니다. `substr`은 문자열의 일부를 추출하는 것이 주 목적이므로, 단순 비교를 위해서는 오버헤드가 더 클 수 있습니다.

**메모리 사용**:
- `strncmp`는 새로운 문자열을 생성하지 않고 직접 비교하므로 메모리 사용이 적습니다.
- `substr`는 새로운 문자열을 생성하므로 추가 메모리 할당이 발생합니다.

**연산 속도**:
- `strncmp`는 불필요한 메모리 복사가 없고, 필요한 만큼만 비교하여 빠르게 종료할 수 있습니다.
- `substr`는 새로운 문자열을 생성하고, 이 문자열과 비교하는 과정을 거치므로 더 많은 시간이 소요됩니다.

따라서, **`strncmp`가 더 효율적이고 빠릅니다**. 이는 추가 메모리 할당이 없고, 직접 비교를 수행하기 때문입니다.

결론적으로, 문자열의 시작 부분이 특정 패턴과 일치하는지만을 확인하고자 할 때는 `strncmp` 함수를 사용하는 것이 더 효율적이며 성능상의 이점이 있습니다. 이는 함수의 내부 구현과 연산의 복잡성에서 기인합니다. 따라서, 성능을 중시하는 상황에서는 `strncmp`의 사용을 권장합니다.

## `strncmp` 사용

```php
strncmp('abcd-efgh-ij-123e4567-e89b-12d3-a456-426614174000', 'abcd-efgh-ij-', 13) === 0
```

- **작동 방식**: `strncmp` 함수는 두 문자열의 앞부분을 비교합니다. 여기서는 첫 번째 문자열의 앞 13자와 두 번째 문자열을 비교합니다.
- **효율성**: `strncmp`는 비교를 수행할 때 새로운 문자열을 생성하지 않고, 원래 문자열의 부분을 직접 비교합니다. 이는 메모리 할당이 필요 없으므로 효율적입니다.
- **속도**: `strncmp`는 필요할 때까지 비교를 중단할 수 있으며, 불필요한 메모리 복사가 없으므로 빠릅니다.

## `substr` 사용

```php
substr('abcd-efgh-ij-123e4567-e89b-12d3-a456-426614174000', 0, 13) === 'abcd-efgh-ij-'
```

- **작동 방식**: `substr` 함수는 첫 번째 문자열의 앞 13자를 추출하여 새로운 문자열을 생성합니다. 그런 다음 새 문자열과 두 번째 문자열을 비교합니다.
- **효율성**: `substr`는 새로운 문자열을 생성하는 데 추가 메모리를 사용합니다. 이는 메모리 할당과 해제를 필요로 하므로 더 비효율적입니다.
- **속도**: `substr`는 새로운 문자열을 생성하는 데 시간이 걸리므로, `strncmp`보다 느립니다.

## 벤치마크

```php
<?php
class StrBenchMark
{
    public function main()
    {
        // 패턴과 반복 횟수 설정
        $pattern = 'abcd-efgh-ij-';
        $patternLength = strlen($pattern);
        $iterations = 100 * 1000; // 반복 횟수

        // 벤치마크 실행
        $this->benchmark($iterations, $pattern, $patternLength);
    }

    // 벤치마크 테스트 함수
    public function benchmark($iterations, $pattern, $patternLength)
    {
        $strncmpTime = 0;
        $substrTime = 0;

        for ($i = 0; $i < $iterations; $i++) {
            $testString = $this->generateRandomString(rand(20, 100)) . $pattern; // 패턴이 확실히 포함되도록 함

            // strncmp 테스트
            $start = microtime(true);
            $result = strncmp($testString, $pattern, $patternLength) === 0;
            $strncmpTime += microtime(true) - $start;

            // substr 테스트
            $start = microtime(true);
            $result = substr($testString, 0, $patternLength) === $pattern;
            $substrTime += microtime(true) - $start;
        }

        // 7.6461553573608 × 10^-7 seconds, 0.00000076461553573608 seconds, 0.765 마이크로초
        echo 'Average time for strncmp: ' . ($strncmpTime / $iterations) . " seconds\n";
        // 7.7724695205688 × 10^-7 seconds, 0.00000077724695205688 seconds, 0.777 마이크로초
        echo 'Average time for substr: ' . ($substrTime / $iterations) . " seconds\n";
    }

    // 문자열 생성 함수
    public function generateRandomString($length)
    {
        $characters = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
        $charactersLength = strlen($characters);
        $randomString = '';
        for ($i = 0; $i < $length; $i++) {
            $randomString .= $characters[rand(0, $charactersLength - 1)];
        }
        return $randomString;
    }
}
?>
```
