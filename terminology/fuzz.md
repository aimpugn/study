# fuzz

- [fuzz](#fuzz)
    - [Fuzz](#fuzz-1)
    - [Fuzzy](#fuzzy)
        - [Fuzzy Logic (퍼지 논리)](#fuzzy-logic-퍼지-논리)
        - [Fuzzy Set (퍼지 집합)](#fuzzy-set-퍼지-집합)
        - [Fuzzy Search (퍼지 검색)](#fuzzy-search-퍼지-검색)
        - [Fuzzy Sort (퍼지 정렬)](#fuzzy-sort-퍼지-정렬)
        - [Fuzzy Test](#fuzzy-test)
    - [Fuzzing](#fuzzing)
        - [Fuzzing의 핵심 요소](#fuzzing의-핵심-요소)
        - [Fuzzing과 관련된 기술 및 도구](#fuzzing과-관련된-기술-및-도구)
            - [Go Fuzzing supported by OSS-Fuzz](#go-fuzzing-supported-by-oss-fuzz)
            - [OSS-Fuzz](#oss-fuzz)
            - [FZF (Fuzzy Finder)](#fzf-fuzzy-finder)
    - [심화 학습 내용 및 관련 개념](#심화-학습-내용-및-관련-개념)
    - [예제](#예제)
        - [Fuzzing과 Fuzzy test](#fuzzing과-fuzzy-test)
- [Fuzz, Fuzzy, 그리고 Fuzzing: 개념과 응용](#fuzz-fuzzy-그리고-fuzzing-개념과-응용)
    - [목차](#목차)
    - [서론](#서론)
    - [Fuzz](#fuzz-2)
    - [Fuzzy](#fuzzy-1)
        - [Fuzzy Logic (퍼지 논리)](#fuzzy-logic-퍼지-논리-1)
        - [Fuzzy Set (퍼지 집합)](#fuzzy-set-퍼지-집합-1)
        - [Fuzzy Search (퍼지 검색)](#fuzzy-search-퍼지-검색-1)
        - [Fuzzy Sort (퍼지 정렬)](#fuzzy-sort-퍼지-정렬-1)
    - [Fuzzing](#fuzzing-1)
        - [Fuzzing의 유형](#fuzzing의-유형)
        - [Fuzzy Test vs Fuzzing](#fuzzy-test-vs-fuzzing)
    - [관련 기술 및 도구](#관련-기술-및-도구)
    - [결론](#결론)

## Fuzz

"Fuzz"는 본래 "보풀" 또는 "잡음"을 의미하는 일반적인 영어 단어입니다.
컴퓨터 과학 분야에서는 소프트웨어 테스팅에서 사용되는 *무작위 또는 비정상적인 데이터*를 의미합니다.

이러한 무작위/비정상 데이터의 특징은 다음과 같습니다:
- 예측 불가능하고 구조화되지 않은 데이터
- 소프트웨어의 예외 처리 능력을 테스트하는 데 사용
- 일반적으로 자동화된 도구에 의해 생성됨

이러한 데이터의 용도는:
- 버그 발견
- 보안 취약점 탐지
- 소프트웨어의 견고성 테스트

아래 코드는 무작위의 "fuzz" 데이터를 생성합니다.

```python
import random
import string

def generate_fuzz(length):
    return ''.join(random.choice(string.printable) for _ in range(length))

def process_input(data):
    # 예를 들어, 소프트웨어가 이 입력을 처리한다고 가정합니다.
    if not isinstance(data, str):
        raise ValueError("Invalid input type")
    if len(data) > 100:
        raise ValueError("Input too long")
    return data.upper()

# 무작위 fuzz 데이터 생성
fuzz_data = generate_fuzz(100)

# 생성된 fuzz 데이터로 함수 테스트
try:
    result = process_input(fuzz_data)
    print(f"Processed result: {result}")
except Exception as e:
    print(f"Error occurred: {e}")
```

이러한 데이터는 소프트웨어 테스팅에서 예상치 못한 입력을 시뮬레이션하는 데 사용될 수 있습니다.

## Fuzzy

"모호한" 또는 "애매한"이라는 뜻을 가진 영어 단어

"Fuzzy"라는 단어의 어원:
- 15세기 중반 네덜란드어 "voos" (스펀지 같은, 부드러운)에서 유래
- 17세기 영어로 유입되어 "fluffy" (보풀이 있는, 부드러운)의 의미로 사용
- 현대에 와서 "불명확한", "모호한"의 의미로 확장

일반적 의미:
- 경계가 명확하지 않은 상태
- 이진법적 (예/아니오)으로 구분하기 어려운 상황
- 정도의 차이가 있는 상태를 표현

'Fuzzy'라는 용어는 1965년 아제르바이잔 출신의 수학자 로트피 자데(Lotfi A. Zadeh)가 'Fuzzy Set Theory(퍼지 집합 이론)'를 제안하면서 널리 알려지게 되었습니다.
자데는 현실 세계의 많은 문제들이 명확한 경계나 이진적인 답을 가지지 않으며, 이를 모델링하기 위해서는 *불확실성을 다룰 수 있는 논리*가 필요하다고 주장했습니다.
이에 따라 'Fuzzy Logic(퍼지 논리)'와 같은 개념이 발전하였고, 이는 다양한 분야에 걸쳐 응용되고 있습니다.

즉, 'Fuzzy'는 고전적인 이진 논리(True/False)와는 달리, *정보나 데이터가 명확하지 않거나 불확실한 상태에서 처리될 수 있는 방법론*을 나타냅니다.

### Fuzzy Logic (퍼지 논리)

퍼지 논리는 1965년 Lotfi A. Zadeh가 제안한 개념으로, 전통적인 이진 논리를 확장한 다치 논리 시스템입니다.

특징:
- 0과 1 사이의 연속적인 값을 사용하여 부분적 진리를 표현
- 자연어의 애매모호함을 수학적으로 모델링
- 복잡한 시스템을 간단한 IF-THEN 규칙으로 표현 가능

응용 분야:
- 제어 시스템 (예: 에어컨, 세탁기의 자동 제어)
- 패턴 인식 및 이미지 처리
- 의사 결정 지원 시스템

### Fuzzy Set (퍼지 집합)

퍼지 집합은 전통적인 집합 이론을 확장한 개념으로, 원소의 소속 정도를 0과 1 사이의 실수로 표현합니다.

특징:
- 부분적 소속을 허용
- 소속 함수(membership function)를 통해 원소의 소속 정도를 정의

수학적 표현:
퍼지 집합 A는 다음과 같이 정의됩니다:
A = {(x, μA(x)) | x ∈ X}
여기서 X는 전체 집합, μA(x)는 x의 A에 대한 소속도 함수 (0 ≤ μA(x) ≤ 1)

예시:
"젊은 사람"이라는 퍼지 집합에서:
- 20세: 소속도 1.0
- 30세: 소속도 0.8
- 40세: 소속도 0.5
- 50세: 소속도 0.2

### Fuzzy Search (퍼지 검색)

퍼지 검색은 정확한 일치가 아닌 *유사성을 기반으로 검색*을 수행하는 기법입니다.
'Fuzzy Search'는 검색어와 정확히 일치하지 않는 경우에도 관련된 결과를 반환하는 검색 방법을 의미합니다.
이는 오타, 철자 오류, 혹은 유사한 표현들을 포함하여 검색 결과를 더 유연하게 제공하는 데 사용됩니다.

예를 들어, 'apple'을 검색했을 때 'appel'과 같은 유사한 단어도 결과에 포함될 수 있습니다.
이는 편집 거리(Edit Distance)와 같은 알고리즘을 사용하여 구현됩니다.

특징:
- 철자 오류, 유사한 발음, 약간의 변형을 허용
- 검색의 유연성을 높이고 사용자 의도를 더 잘 반영

알고리즘:
- Levenshtein 거리
- Soundex
- n-gram

```python
from fuzzywuzzy import fuzz

def fuzzy_search(query, items, threshold=70):
    results = []
    for item in items:
        ratio = fuzz.ratio(query.lower(), item.lower())
        if ratio >= threshold:
            results.append((item, ratio))
    return sorted(results, key=lambda x: x[1], reverse=True)

# 사용 예
items = ["apple", "banana", "cherry", "date", "elderberry"]
query = "banan"
print(fuzzy_search(query, items))
# 출력: [('banana', 91), ('apple', 70)]
```

### Fuzzy Sort (퍼지 정렬)

퍼지 정렬은 퍼지 논리를 사용하여 항목들을 정렬하는 방법입니다.
전통적인 정렬과 달리, 데이터가 완벽하게 정렬되지 않았거나 항목들 간의 정렬 기준이 모호한 상황에서 사용됩니다.

예를 들어, 검색 결과를 '관련성'에 따라 정렬하는데, 이 관련성은 정량적으로 정의하기 어렵고 모호한 기준에 의해 정렬됩니다.
이와 같은 정렬 알고리즘은 입력 데이터의 불확실성을 고려하여 유연하게 동작합니다.

특징:
- 부분적 순서 관계를 허용
- 여러 기준에 따른 복합적인 정렬이 가능

알고리즘:
1. 각 항목에 대해 퍼지 멤버십 함수를 정의
2. 항목들 간의 퍼지 관계를 계산
3. 퍼지 관계를 기반으로 순서를 결정

```python
def fuzzy_sort(items, criteria):
    membership_values = []
    for item in items:
        values = [criterion(item) for criterion in criteria]
        membership_values.append(values)

    sorted_items = []
    while membership_values:
        max_index = find_max_membership(membership_values)
        sorted_items.append(items[max_index])
        del items[max_index]
        del membership_values[max_index]

    return sorted_items

# 사용 예
items = [{"name": "A", "price": 100, "quality": 0.8},
         {"name": "B", "price": 80, "quality": 0.9},
         {"name": "C", "price": 120, "quality": 0.7}]

def price_criterion(item):
    return 1 - (item["price"] / 150)  # 가격이 낮을수록 높은 값

def quality_criterion(item):
    return item["quality"]

sorted_items = fuzzy_sort(items, [price_criterion, quality_criterion])
```

### Fuzzy Test

```python
import random
import string

def generate_fuzz(length):
    return ''.join(random.choice(string.printable) for _ in range(length))

def fuzz_test(function_to_test, num_tests=1000, max_length=100):
    for _ in range(num_tests):
        fuzz_input = generate_fuzz(random.randint(1, max_length))
        try:
            function_to_test(fuzz_input)
        except Exception as e:
            print(f"Exception caught: {e}")
            print(f"Input that caused exception: {fuzz_input}")

# 테스트할 함수
def example_function(input_string):
    if len(input_string) > 50:
        raise ValueError("Input too long")
    return input_string.upper()

# Fuzzing 실행
fuzz_test(example_function)
```

## Fuzzing

"Fuzzing"은 자동화된 소프트웨어 테스팅 기법으로, 무작위 또는 변형된 입력 데이터를 프로그램에 제공하여 버그, 취약점, 충돌 등을 발견하는 과정입니다.

특징:
- 자동화된 테스트 프로세스
- 대량의 테스트 케이스 생성 및 실행
- 예상치 못한 동작이나 취약점 발견에 효과적

Fuzzing의 유형:
- 무작위(Random) Fuzzing: 완전히 무작위한 데이터 사용
- 변형(Mutation) Fuzzing: 유효한 입력을 변형하여 테스트
- 생성(Generation) Fuzzing: 특정 형식이나 프로토콜에 맞는 데이터 생성

Fuzzing의 개념은 1980년대 초반에 벨 연구소의 Barton Miller 교수와 그의 학생들이 처음 개발했습니다.
그들은 유닉스 운영 체제의 명령줄 도구들에 임의의 입력을 제공하여 프로그램이 어떻게 반응하는지 관찰했습니다.
소프트웨어가 예기치 않은 입력이나 극단적인 상황에 어떻게 반응하는지를 이해하고, 그 결과를 통해 소프트웨어의 안정성과 신뢰성을 개선하고자 하는 것이 목적이었다.
이 초기 실험에서 많은 표준 유틸리티들이 비정상적으로 종료되거나 실패하는 것을 발견했고, 이로부터 fuzzing이라는 개념이 탄생했다.

[Fuzzing](https://en.wikipedia.org/wiki/Fuzzing)은 소프트웨어 테스팅의 한 방법으로, 예측 불가능한 상황을 다루거나 프로그램에 무작위 데이터(퍼지 데이터)를 입력하여 실행시켜 오류, 취약점, 버그 등을 찾아내는 과정입니다.
일반적인 테스트 케이스로는 발견하기 어려운 예외적인 상황이나 입력에 대한 프로그램의 반응을 탐색합니다.

### Fuzzing의 핵심 요소

- 자동화된 테스팅:

    Fuzzing은 주로 자동화된 도구를 사용하여 수행됩니다.
    이 도구들은 무작위 데이터를 생성하고, 이를 프로그램에 입력한 후 결과를 관찰합니다.

- 무작위 데이터 생성:

    Fuzzing의 중요한 특징은 무작위성입니다.
    테스팅 도구는 일반적인 사용 사례와는 다른, 예측 불가능한 데이터를 생성하여 프로그램을 테스트합니다.

- 에러 및 취약점 탐지:

    Fuzzing의 주된 목적은 프로그램이 무작위 데이터를 처리하면서 발생할 수 있는 에러나 취약점을 찾아내는 것입니다.
    무작위 데이터를 입력함으로써 개발자는 예상치 못한 상황에서의 시스템 반응을 확인할 수 있으며, 이를 통해 보안 허점이 있는지 확인합니다.

- 커버리지 기반 접근법:

    많은 현대의 fuzzing 도구들은 코드 커버리지 정보를 사용하여 보다 효율적으로 데이터를 생성합니다.
    이를 통해 코드의 다양한 부분이 테스트될 수 있도록 합니다.

### Fuzzing과 관련된 기술 및 도구

#### [Go Fuzzing](https://go.dev/doc/security/fuzz/) supported by [OSS-Fuzz](https://google.github.io/oss-fuzz/getting-started/new-project-guide/go-lang/#native-go-fuzzing-support)

Go 프로그래밍 언어에서는 Fuzzing을 사용하여 코드의 특정 부분을 테스트하고, 예외적인 경우에 대한 대응을 확인한다.
Go Fuzzing은 코드 커버리지를 기반으로 하여 효율적으로 테스트 케이스를 생성합니다.

#### [OSS-Fuzz](https://google.github.io/oss-fuzz/)

#### FZF (Fuzzy Finder)

이는 명령줄에서 파일, 명령어 기록, 프로세스 등을 빠르게 검색하고 선택하는 도구.
여기서 'fuzzy'는 정확한 문자 일치가 아닌 유사한 결과를 반환하는 검색 방식을 의미한다.

## 심화 학습 내용 및 관련 개념

- 퍼지 제어 시스템: 퍼지 논리를 기반으로 한 제어 시스템으로, 복잡한 비선형 시스템을 효과적으로 제어할 수 있습니다.
- 뉴로-퍼지 시스템: 신경망과 퍼지 논리를 결합한 하이브리드 시스템으로, 학습 능력과 해석 가능성을 동시에 갖춥니다.
- 유전 알고리즘을 이용한 퍼지 규칙 최적화: 유전 알고리즘을 사용하여 퍼지 시스템의 규칙과 멤버십 함수를 자동으로 최적화하는 기법입니다.
- 퍼지 클러스터링: 데이터 포인트가 여러 클러스터에 부분적으로 속할 수 있는 클러스터링 기법입니다.
- 퍼지 데이터베이스: 불확실한 정보를 저장하고 검색할 수 있는 데이터베이스 시스템입니다.
- Symbolic Fuzzing: 프로그램의 심볼릭 실행을 통해 더 효과적인 테스트 케이스를 생성하는 고급 Fuzzing 기법입니다.

- **퍼지 논리(Fuzzy Logic)**: 전통적인 논리보다 더 많은 유연성을 제공하며, 불확실한 상황에서도 논리를 적용할 수 있는 기법.
- **편집 거리(Edit Distance)**: 문자열 유사성을 측정하는 알고리즘으로, 퍼지 검색에서 자주 사용됩니다.
- **퍼지 집합 이론(Fuzzy Set Theory)**: 퍼지 집합을 수학적으로 정의하고 분석하는 이론으로, 퍼지 개념의 기초가 됩니다.

## 예제

### Fuzzing과 Fuzzy test

```php
<?php

class AmountUtil
{
    const DECIMAL_PRECISION_LIMIT = 2;

    public static function getAmountWithPrecision($amount, $precision = self::DECIMAL_PRECISION_LIMIT)
    {
        if ($precision < 0) {
            $precision = 0;
        }
        $multiplier = pow(10, $precision);
        return floor($amount * $multiplier) / $multiplier;
    }
}

class AmountUtilTest extends PHPUnit\Framework\TestCase
{
    // Fuzzing approach
    public function testGetAmountWithPrecisionFuzzing()
    {
        $numTests = 1000; // Number of random tests to run

        for ($i = 0; $i < $numTests; $i++) {
            // Generate random input
            $amount = mt_rand() / mt_getrandmax() * 1000000; // Random float between 0 and 1,000,000
            $precision = mt_rand(0, 10); // Random integer between 0 and 10

            $result = AmountUtil::getAmountWithPrecision($amount, $precision);

            // Assertions
            $this->assertLessThanOrEqual($amount, $result, "Result should not be greater than input");
            $this->assertGreaterThan($amount - pow(10, -$precision), $result, "Result should not be too small");
            $this->assertEquals(
                $result,
                round($result, $precision),
                "Result should have correct precision"
            );
        }
    }

    // Fuzzy Test approach
    public function testGetAmountWithPrecisionFuzzyTest()
    {
        $testCases = [
            ['amount' => 100.126, 'precision' => 2, 'expectedRange' => [100.12, 100.13]],
            ['amount' => 99.995, 'precision' => 2, 'expectedRange' => [99.99, 100.00]],
            ['amount' => 0.001, 'precision' => 3, 'expectedRange' => [0.001, 0.002]],
            ['amount' => 999.999, 'precision' => 1, 'expectedRange' => [999.9, 1000.0]],
            ['amount' => 0.0001, 'precision' => 4, 'expectedRange' => [0.0001, 0.0002]],
        ];

        foreach ($testCases as $case) {
            $result = AmountUtil::getAmountWithPrecision($case['amount'], $case['precision']);

            $this->assertGreaterThanOrEqual(
                $case['expectedRange'][0],
                $result,
                "Result should be within the expected fuzzy range"
            );
            $this->assertLessThan(
                $case['expectedRange'][1],
                $result,
                "Result should be within the expected fuzzy range"
            );
        }
    }
}
```

1. Fuzzing (testGetAmountWithPrecisionFuzzing):
   - 무작위로 생성된 대량의 입력 데이터를 사용합니다.
   - 예상치 못한 에지 케이스나 버그를 발견하는 데 효과적입니다.(함수의 견고성을 테스트)
   - 함수의 일반적인 동작과 견고성을 테스트합니다.
   - 각 테스트 케이스에 대해 일반적인 규칙(예: 결과가 입력보다 크지 않아야 함)을 검증합니다.

2. Fuzzy Test (testGetAmountWithPrecisionFuzzyTest):
   - 특정한 테스트 케이스를 사용하지만, 정확한 값 대신 허용 가능한 범위를 정의합니다.
   - 함수의 동작이 "대략적으로" 올바른지 확인합니다.
   - 경계값이나 특별한 상황(예: 반올림이 필요한 경우)에 중점을 둡니다.
   - 각 테스트 케이스에 대해 결과가 예상 범위 내에 있는지 확인합니다.

---

# Fuzz, Fuzzy, 그리고 Fuzzing: 개념과 응용

## 목차

- [fuzz](#fuzz)
    - [Fuzz](#fuzz-1)
    - [Fuzzy](#fuzzy)
        - [Fuzzy Logic (퍼지 논리)](#fuzzy-logic-퍼지-논리)
        - [Fuzzy Set (퍼지 집합)](#fuzzy-set-퍼지-집합)
        - [Fuzzy Search (퍼지 검색)](#fuzzy-search-퍼지-검색)
        - [Fuzzy Sort (퍼지 정렬)](#fuzzy-sort-퍼지-정렬)
        - [Fuzzy Test](#fuzzy-test)
    - [Fuzzing](#fuzzing)
        - [Fuzzing의 핵심 요소](#fuzzing의-핵심-요소)
        - [Fuzzing과 관련된 기술 및 도구](#fuzzing과-관련된-기술-및-도구)
            - [Go Fuzzing supported by OSS-Fuzz](#go-fuzzing-supported-by-oss-fuzz)
            - [OSS-Fuzz](#oss-fuzz)
            - [FZF (Fuzzy Finder)](#fzf-fuzzy-finder)
    - [심화 학습 내용 및 관련 개념](#심화-학습-내용-및-관련-개념)
    - [예제](#예제)
        - [Fuzzing과 Fuzzy test](#fuzzing과-fuzzy-test)
- [Fuzz, Fuzzy, 그리고 Fuzzing: 개념과 응용](#fuzz-fuzzy-그리고-fuzzing-개념과-응용)
    - [목차](#목차)
    - [서론](#서론)
    - [Fuzz](#fuzz-2)
    - [Fuzzy](#fuzzy-1)
        - [Fuzzy Logic (퍼지 논리)](#fuzzy-logic-퍼지-논리-1)
        - [Fuzzy Set (퍼지 집합)](#fuzzy-set-퍼지-집합-1)
        - [Fuzzy Search (퍼지 검색)](#fuzzy-search-퍼지-검색-1)
        - [Fuzzy Sort (퍼지 정렬)](#fuzzy-sort-퍼지-정렬-1)
    - [Fuzzing](#fuzzing-1)
        - [Fuzzing의 유형](#fuzzing의-유형)
        - [Fuzzy Test vs Fuzzing](#fuzzy-test-vs-fuzzing)
    - [관련 기술 및 도구](#관련-기술-및-도구)
    - [결론](#결론)

## 서론

"Fuzz", "Fuzzy", "Fuzzing"은 컴퓨터 과학과 소프트웨어 공학 분야에서 중요한 개념들입니다. 이 문서에서는 이 세 가지 개념의 정의, 특징, 응용 분야, 그리고 서로 간의 관계에 대해 상세히 설명합니다.

## Fuzz

## Fuzzy

"Fuzzy"는 "모호한" 또는 "불분명한"이라는 의미를 가진 형용사입니다. 컴퓨터 과학에서 이 개념은 불확실성을 다루는 방법론을 나타냅니다.

1. **어원과 역사**:
   - 15세기 중반 네덜란드어 "voos" (스펀지 같은, 부드러운)에서 유래
   - 1965년 Lotfi A. Zadeh가 'Fuzzy Set Theory(퍼지 집합 이론)'를 제안하면서 컴퓨터 과학 분야에 도입

2. **특징**:
   - 이진법적(예/아니오)으로 구분하기 어려운 상황을 다룸
   - 연속적인 값을 사용하여 부분적 진리를 표현
   - 자연어의 애매모호함을 수학적으로 모델링

3. **주요 개념**:

### Fuzzy Logic (퍼지 논리)

퍼지 논리는 전통적인 이진 논리를 확장한 다치 논리 시스템입니다.

- **특징**:
    - 0과 1 사이의 연속적인 값을 사용
    - 복잡한 시스템을 간단한 IF-THEN 규칙으로 표현 가능

- **응용 분야**:
    - 제어 시스템 (예: 에어컨, 세탁기의 자동 제어)
    - 패턴 인식 및 이미지 처리
    - 의사 결정 지원 시스템

- **예시**:

  ```python
  def fuzzy_temperature(temp):
      if temp < 0:
          return "매우 추움"
      elif 0 <= temp < 10:
          return "추움"
      elif 10 <= temp < 20:
          return "선선함"
      elif 20 <= temp < 30:
          return "따뜻함"
      else:
          return "더움"

  print(fuzzy_temperature(15))  # 출력: 선선함
  ```

### Fuzzy Set (퍼지 집합)

퍼지 집합은 원소의 소속 정도를 0과 1 사이의 실수로 표현합니다.

- **수학적 표현**:
  A = {(x, μA(x)) | x ∈ X}
  여기서 X는 전체 집합, μA(x)는 x의 A에 대한 소속도 함수 (0 ≤ μA(x) ≤ 1)

- **예시**:
  "젊은 사람" 퍼지 집합:
    - 20세: 소속도 1.0
    - 30세: 소속도 0.8
    - 40세: 소속도 0.5
    - 50세: 소속도 0.2

### Fuzzy Search (퍼지 검색)

퍼지 검색은 정확한 일치가 아닌 유사성을 기반으로 검색을 수행하는 기법입니다.

- **특징**:
    - 철자 오류, 유사한 발음, 약간의 변형을 허용
    - 검색의 유연성을 높이고 사용자 의도를 더 잘 반영

- **알고리즘**:
    - Levenshtein 거리
    - Soundex
    - n-gram

- **예시**:

  ```python
  from fuzzywuzzy import fuzz

  def fuzzy_search(query, items, threshold=70):
      results = []
      for item in items:
          ratio = fuzz.ratio(query.lower(), item.lower())
          if ratio >= threshold:
              results.append((item, ratio))
      return sorted(results, key=lambda x: x[1], reverse=True)

  items = ["apple", "banana", "cherry", "date", "elderberry"]
  query = "banan"
  print(fuzzy_search(query, items))
  # 출력: [('banana', 91), ('apple', 70)]
  ```

### Fuzzy Sort (퍼지 정렬)

퍼지 정렬은 퍼지 논리를 사용하여 항목들을 정렬하는 방법입니다.

- **특징**:
    - 부분적 순서 관계를 허용
    - 여러 기준에 따른 복합적인 정렬이 가능

- **알고리즘**:
  1. 각 항목에 대해 퍼지 멤버십 함수를 정의
  2. 항목들 간의 퍼지 관계를 계산
  3. 퍼지 관계를 기반으로 순서를 결정

- **예시**:

  ```python
  def fuzzy_sort(items, criteria):
      membership_values = []
      for item in items:
          values = [criterion(item) for criterion in criteria]
          membership_values.append(values)

      sorted_items = []
      while membership_values:
          max_index = find_max_membership(membership_values)
          sorted_items.append(items[max_index])
          del items[max_index]
          del membership_values[max_index]

      return sorted_items

  # 사용 예
  items = [{"name": "A", "price": 100, "quality": 0.8},
           {"name": "B", "price": 80, "quality": 0.9},
           {"name": "C", "price": 120, "quality": 0.7}]

  def price_criterion(item):
      return 1 - (item["price"] / 150)  # 가격이 낮을수록 높은 값

  def quality_criterion(item):
      return item["quality"]

  sorted_items = fuzzy_sort(items, [price_criterion, quality_criterion])
  ```

## Fuzzing

Fuzzing은 자동화된 소프트웨어 테스팅 기법으로, 무작위 또는 변형된 입력 데이터를 프로그램에 제공하여 버그, 취약점, 충돌 등을 발견하는 과정입니다.

1. **역사**:
   - 1980년대 초반 벨 연구소의 Barton Miller 교수와 그의 학생들이 개발
   - 유닉스 운영 체제의 명령줄 도구들에 임의의 입력을 제공하여 프로그램의 반응을 관찰하는 것으로 시작

2. **특징**:
   - 자동화된 테스트 프로세스
   - 대량의 테스트 케이스 생성 및 실행
   - 예상치 못한 동작이나 취약점 발견에 효과적

3. **핵심 요소**:
   - 자동화된 테스팅
   - 무작위 데이터 생성
   - 에러 및 취약점 탐지
   - 커버리지 기반 접근법

### Fuzzing의 유형

1. **무작위(Random) Fuzzing**:
   완전히 무작위한 데이터를 생성하여 테스트합니다.

2. **변형(Mutation) Fuzzing**:
   유효한 입력 데이터를 변형하여 테스트합니다.

3. **생성(Generation) Fuzzing**:
   특정 형식이나 프로토콜에 맞는 데이터를 생성하여 테스트합니다.

**예시 코드**:

```python
import random
import string

def generate_fuzz(length):
    return ''.join(random.choice(string.printable) for _ in range(length))

def fuzz_test(function_to_test, num_tests=1000, max_length=100):
    for _ in range(num_tests):
        fuzz_input = generate_fuzz(random.randint(1, max_length))
        try:
            function_to_test(fuzz_input)
        except Exception as e:
            print(f"Exception caught: {e}")
            print(f"Input that caused exception: {fuzz_input}")

# 테스트할 함수
def example_function(input_string):
    if len(input_string) > 50:
        raise ValueError("Input too long")
    return input_string.upper()

# Fuzzing 실행
fuzz_test(example_function)
```

### Fuzzy Test vs Fuzzing

Fuzzy Test와 Fuzzing은 서로 다른 개념입니다:

1. **Fuzzy Test**:
   - 불확실하거나 모호한 입력 조건에서 시스템의 동작을 테스트
   - 퍼지 논리의 개념을 테스팅에 적용
   - 주로 설계 단계에서의 요구사항을 검증

2. **Fuzzing**:
   - 무작위 또는 비정상적인 입력을 사용하여 소프트웨어의 취약점을 찾는 기법
   - 주로 구현 단계에서 사용
   - 보안 취약점과 예상치 못한 버그를 찾는 데 중점

## 관련 기술 및 도구

1. **Go Fuzzing**:
   Go 프로그래밍 언어에서 제공하는 내장 퍼징 도구

2. **OSS-Fuzz**:
   Google이 운영하는 오픈소스 프로젝트를 위한 지속적 퍼징 플랫폼

3. **FZF (Fuzzy Finder)**:
   명령줄에서 파일, 명령어 기록, 프로세스 등을 빠르게 검색하고 선택하는 도구

## 결론

Fuzz, Fuzzy, Fuzzing은 각각 다른 개념이지만 모두 불확실성이나 예측 불가능성을 다룬
