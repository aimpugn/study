# Bloom filter

- [Bloom filter](#bloom-filter)
    - [블룸 필터(Bloom Filter)](#블룸-필터bloom-filter)
    - [블룸 필터 특성](#블룸-필터-특성)
    - [블룸 필터의 구성 요소](#블룸-필터의-구성-요소)
        - [해시 함수(Hash Functions)](#해시-함수hash-functions)
    - [블룸 필터 작동 원리](#블룸-필터-작동-원리)
    - [블룸 필터의 수학적 분석](#블룸-필터의-수학적-분석)
    - [블룸 필터 구현](#블룸-필터-구현)
        - [Bloom Filter의 Golang 구현](#bloom-filter의-golang-구현)
    - [블룸 필터의 응용 분야](#블룸-필터의-응용-분야)
    - [블룸 필터의 한계](#블룸-필터의-한계)
    - [심화 학습 내용 및 관련 개념](#심화-학습-내용-및-관련-개념)

## 블룸 필터(Bloom Filter)

블룸 필터(Bloom Filter)는 확률적 자료 구조로, *요소가 집합에 속하는지 여부를 빠르게 확인*할 수 있도록 해줍니다.
블룸 필터는 일부 오차를 허용하는 대신, 매우 효율적인 메모리 사용과 빠른 검색 속도를 제공합니다.
- "이 요소는 집합에 속하지 않는다"는 확실한 결과를 제공합니다.
- 하지만 "이 요소는 집합에 속한다"는 결과에는 일부 오류(위양성, False Positive)가 있을 수 있습니다.

있다고 잘못 긍정할 수는 있지만, 있는 것을 없다고 하지는 않는다.

- **False Positive**: 집합에 실제로 속하지 않는 요소를 속한다고 잘못 긍정할 수 있습니다.
- **False Negative**: 집합에 실제로 속해 있는 요소를 없다고 하지는 않습니다. 블룸 필터는 이 오류를 방지합니다.

블룸 필터를 **문지기**에 비유할 수 있습니다.
이 문지기는 매우 효율적으로 특정 사람이 어떤 방에 들어갔는지 기억하려고 합니다.
그러나 이 문지기에게는 한 가지 한계가 있습니다:

1. **문지기의 판단이 틀릴 수 있는 경우 (False Positive)**: 문지기의 실수

    어떤 사람이 들어가지 않았지만, 문지기는 그 사람이 들어갔다고 잘못 기억할 수 있습니다.

    예를 들어, 실제로는 A와 B가 들어갔지만, 문지기는 A와 B 외에 C도 들어갔다고 생각할 수 있습니다.

2. **문지기의 확실한 판단 (False Negative 없음)**: 문지기의 확실한 판단

    하지만 어떤 사람이 실제로 방에 들어갔다면, 문지기가 그 사실을 모르는 일은 없습니다.
    즉, 문지기가 특정 사람이 들어가지 않았다고 말한다면, 그 사람은 정말로 방에 들어가지 않은 것입니다.

따라서 문지기가 그 사람이 방에 있다고 할 때, 실제로 그 사람이 방에 있을 가능성은 매우 높지만, 드물게 잘못된 판단일 수도 있습니다.
하지만 문지기가 어떤 사람이 방에 없다고 확신할 때, 그 사람은 진짜 방에 없습니다.

블룸 필터는 1970년에 Burton Howard Bloom에 의해 제안되었습니다.
블룸 필터의 효율성과 오차 허용 가능성 덕분에, 인터넷 검색 엔진, 데이터베이스 시스템, 캐시 관리 등 다양한 분야에서 폭넓게 사용되고 있습니다.

## 블룸 필터 특성

1. **공간 효율성**: 매우 큰 집합을 표현하는 데 적은 메모리를 사용합니다.
2. **시간 복잡도**: 요소 추가와 검사가 O(k) 시간에 수행됩니다 (k는 해시 함수의 수).
3. **확률적 특성**: 거짓 음성은 절대 발생하지 않지만, 거짓 양성이 발생할 수 있습니다.
4. **한계점**: 요소를 삭제할 수 없습니다. 삭제가 필요한 경우 Counting Bloom Filter 등의 변형을 사용해야 합니다.

## 블룸 필터의 구성 요소

1. **비트 배열(Bit Array)**: 길이가 $m$인 비트 배열로, 초기에는 모든 비트가 0으로 설정됩니다.
2. **해시 함수(Hash Functions)**: $k$개의 독립적인 해시 함수가 필요합니다. 각 해시 함수는 입력 데이터를 받아서 $m$ 길이의 비트 배열의 인덱스를 반환합니다.

### 해시 함수(Hash Functions)

1. 블룸 필터의 해시 함수 요구사항:

    - 빠른 계산 속도: 블룸 필터는 빠른 검색을 위한 것이므로, 해시 함수도 빠르게 계산될 수 있어야 합니다.
    - 균등 분포: 해시 값이 비트 배열 전체에 고르게 분포되어야 합니다.
    - 서로 독립적: 여러 해시 함수는 서로 독립적인 결과를 생성해야 합니다.

2. 일반적으로 사용되는 해시 함수 유형:

    - 비암호화 해시 함수:
        - MurmurHash
        - FNV (Fowler-Noll-Vo)
        - CityHash
        - xxHash

    - 단순한 수학적 함수:

        곱셈 해시 함수: $h(x) = (a \times x + b) \mod m$
        - $a$와 $b$: 소수
        - $m$: 비트 배열의 크기

    - 암호화 해시 함수(주의: 이들은 계산 비용이 높아 블룸 필터에는 과도할 수 있습니다)

        - MD5
        - SHA1
        - SHA256 등

3. 복수의 해시 함수 구현 방법:

    - 실제로 여러 개의 다른 해시 함수 사용합니다. 예를 들어, MurmurHash, FNV, CityHash를 각각 사용합니다.

    - 더블 해싱: 두 개의 해시 함수를 사용하여 k개의 해시 값 생성합니다.

      $h_i(x) = (h1(x) + i \times h2(x)) \mod m$

      여기서 $i = 0, 1, ..., k-1$

    - 단일 해시 함수의 출력을 분할:
      하나의 해시 함수 출력을 여러 부분으로 나누어 사용

4. 해시 함수 선택 시 고려사항:

   - 성능: 블룸 필터의 주요 장점은 속도이므로, 빠른 해시 함수가 중요합니다.
   - 충돌 저항성: 균등한 분포를 위해 충돌이 적은 해시 함수가 좋습니다.
   - 구현 복잡성: 간단한 해시 함수로도 충분한 경우가 많습니다.
   - 보안: 일반적인 블룸 필터 용도에서는 암호학적 안전성이 필요하지 않습니다.

## 블룸 필터 작동 원리

블룸 필터는 *원소가 집합에 속하는지 여부를 검사하는데 사용되는 **확률적** 자료 구조*입니다.
공간 효율성이 매우 높지만, 거짓 양성이 발생할 수 있습니다.

1. **삽입**:
    - 요소를 삽입할 때, $k$개의 해시 함수를 사용하여 비트 배열에서 $k$개의 위치를 계산합니다.
    - 해당 위치의 비트를 1로 설정합니다.

2. **검사**:
    - 요소가 집합에 있는지 검사할 때, 동일한 $k$개의 해시 함수로 비트 배열의 $k$개의 위치를 계산합니다.
    - 계산된 **모든** 위치의 비트가 1이라면, 해당 요소가 집합에 있을 가능성이 있습니다. (하지만 False Positive일 수도 있음)
    - 하나라도 0이 있다면, 해당 요소는 집합에 없다고 확신할 수 있습니다.

구체적으로 다음과 같은 방식으로 동작하게 됩니다:

1. 초기 상태

    **비트 배열**:
    블룸 필터는 모두 0으로 초기화된 비트 배열로 시작합니다.
    예시에서는 16비트를 사용하지만, 실제로는 훨씬 더 큽니다.

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 0 0 0 0 0 0 0 0 0  0  0  0  0  0
    ```

    **해시 함수**:
    블룸 필터는 여러 개의 해시 함수를 사용합니다.
    아래 예시에서는 3개를 사용하지만, 실제로는 더 많이 사용할 수 있습니다.

2. 요소 추가: "cat"

    "cat"을 추가할 때, 각 해시 함수가 다음 인덱스를 반환한다고 가정합니다:
    - 해시 함수 1: 2
    - 해시 함수 2: 5
    - 해시 함수 3: 11

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 1 0 0 1 0 0 0 0 0  1  0  0  0  0
        ↑     ↑            ↑
        │     │            │
        │     │            └────────────── "cat" 해시 3
        │     └─────────────────────────── "cat" 해시 2
        └───────────────────────────────── "cat" 해시 1
    ```

3. 요소 추가: "dog"

    "dog"를 추가할 때의 해시 결과:
    - 해시 함수 1: 3
    - 해시 함수 2: 7
    - 해시 함수 3: 14

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 1 1 0 1 0 1 0 0 0  1  0  0  1  0
        ↑ ↑   ↑   ↑        ↑        ↑
        │ │   │   │        │        └───── "dog" 해시 3
        │ │   │   │        └────────────── "cat" 해시 3       
        │ │   │   └─────────────────────── "dog" 해시 2  
        │ │   └─────────────────────────── "cat" 해시 2
        │ └─────────────────────────────── "dog" 해시 1
        └───────────────────────────────── "cat" 해시 1
    ```

4. 요소 검사: "fish"

    "fish"가 존재하는지 검사합니다. 해시 결과가 다음과 같다고 가정합니다:
    - 해시 함수 1: 1
    - 해시 함수 2: 4
    - 해시 함수 3: 12

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 1 1 0 1 0 1 0 0 0  1  0  0  1  0
      ↓ ↑ ↑ ↓ ↑   ↑        ↑  ↓     ↑
      │ │ │ │ │   │        │  │     └───── "dog" 해시 3
      │ │ │ │ │   │        └──│─────────── "cat" 해시 3       
      │ │ │ │ │   └───────────│─────────── "dog" 해시 2  
      │ │ │ │ └───────────────│─────────── "cat" 해시 2
      │ │ └─│─────────────────│─────────── "dog" 해시 1
      │ └───│─────────────────│─────────── "cat" 해시 1
      │     │                 │
      │     │                 ↓
      │     │                 "fish" 해시 3 (0)    
      │     ↓
      │     "fish" 해시 2 (0)
      ↓
      "fish" 해시 1 (0)
    ```

    `fish`는 존재하지 않습니다.
    하나 이상의 비트가 0이므로, 블룸 필터는 `fish`가 집합에 없다고 정확하게 판단합니다.

5. 요소 검사: "cat" (기존 요소)

    "cat"이 존재하는지 다시 검사합니다:

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 1 1 0 1 0 1 0 0 0  1  0  0  1  0
        ↑ ↑   ↑   ↑        ↑       ↑
        │ │   │   │        │       └────── "dog" 해시 3
        │ │   │   │        └────────────── "cat" 해시 3 (1)
        │ │   │   └─────────────────────── "dog" 해시 2
        │ │   └─────────────────────────── "cat" 해시 2 (1)
        │ └─────────────────────────────── "dog" 해시 1
        └───────────────────────────────── "cat" 해시 1 (1)
    ```

    **모든** 해당 비트가 1이므로, 블룸 필터는 `cat`이 집합에 있다고 판단합니다.

6. 거짓 양성 예시: "cow"

    "cow"가 존재하는지 검사합니다. 해시 결과가 다음과 같다고 가정합니다:
    - 해시 함수 1: 2
    - 해시 함수 2: 7
    - 해시 함수 3: 11

    ```sh
    0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15
    0 0 1 1 0 1 0 1 0 0 0  1  0  0  1  0
        ↑ ↑   ↑   ↑        ↑       ↑
        │ │   │   │        │       └────── "dog" 해시 3
        │ │   │   │        └────────────── "cat" 해시 3, "cow" 해시 3 (1)
        │ │   │   └─────────────────────── "dog" 해시 2
        │ │   └─────────────────────────── "cat" 해시 2, "cow" 해시 2 (1)
        │ └─────────────────────────────── "dog" 해시 1
        └───────────────────────────────── "cat" 해시 1, "cow" 해시 1 (1)
    ```

     **모든** 해당 비트가 1이므로, 블룸 필터는 `cow`가 존재한다고 판단합니다.
     그러나 `cow`는 실제로 추가된 적이 없으므로, 이는 거짓 양성입니다.

## 블룸 필터의 수학적 분석

**오류 확률 (False Positive Rate)**:

블룸 필터의 False Positive Rate는 해시 함수의 개수 $k$, 비트 배열의 길이 $m$, 삽입된 요소의 개수 $n$에 따라 결정됩니다.

이 확률은 다음과 같이 계산할 수 있습니다:

$$
P(\text{False Positive}) = \left(1 - \left(1 - \frac{1}{m}\right)^{kn}\right)^k
$$

- 여기서 $kn$은 비트 배열에서 1로 설정된 비트의 평균 개수를 나타냅니다.
- $m$과 $k$가 클수록, 즉 비트 배열이 길고 해시 함수가 많을수록 False Positive 확률은 낮아집니다.

## 블룸 필터 구현

### Bloom Filter의 Golang 구현

Bloom Filter는 특정 요소가 집합에 속하는지 여부를 확인하는 *확률적* 자료구조입니다.
이 자료구조는 공간 효율적이지만, 실제로는 포함되지 않은 요소를 포함한다고 잘못 판단하는 **거짓 긍정(false positive)**이 발생할 수 있습니다.
그러나 **거짓 부정(false negative)**은 발생하지 않으므로, 부정(negative)는 항상 옳은 부정(negative)입니다.
따라서 포함되지 않은 요소를 포함하지 않았다고 정확하게 판단할 수 있습니다.

1. **비트 배열(Bit Array)**: 요소가 집합에 속하는지를 나타내는 비트를 저장하는 배열입니다.
2. **해시 함수(Hash Functions)**: 입력된 데이터를 해싱하여 비트 배열의 특정 비트를 설정하는 데 사용되는 여러 개의 해시 함수입니다.

```go
package main

import (
    "hash/fnv"    // 해시 함수로 FNV-1a 해시 알고리즘을 사용하기 위해 임포트합니다.
    "math"        // 수학적 계산을 위해 math 패키지를 임포트합니다.
)

// Bloom Filter의 작동 원리:
// 1. 요소를 추가할 때, 여러 해시 함수를 사용하여 비트 배열의 여러 위치를 1(true)로 설정합니다.
// 2. 요소의 존재 여부를 확인할 때, 동일한 해시 함수들을 사용하여 모든 해당 위치가 1인지 확인합니다.
// 3. 만약 하나라도 0(false)이면, 그 요소는 확실히 집합에 없습니다.
// 4. 모두 1이면, 요소가 집합에 있을 가능성이 높지만, 거짓 양성의 가능성이 있습니다.

// 이 구현의 장점:
// - 공간 효율성: 비트 배열을 사용하여 메모리 사용을 최소화합니다.
// - 시간 효율성: 요소 추가와 검사가 O(k) 시간 복잡도를 가집니다. (k는 해시 함수의 수)
// - 확장성: n과 p를 조절하여 다양한 크기와 정확도의 Bloom Filter를 생성할 수 있습니다.

// 주의사항:
// - Bloom Filter는 요소 삭제를 지원하지 않습니다.
// - 거짓 양성의 가능성이 있으므로, 절대적인 정확성이 필요한 경우에는 부적합할 수 있습니다.

// BloomFilter 구조체는 Bloom Filter의 핵심 구성 요소를 정의합니다.
type BloomFilter struct {
    bitArray []bool  // 비트 배열: 각 요소의 존재 여부를 표시하는 불리언 배열입니다.
    k        int     // 해시 함수의 개수: 더 많은 해시 함수를 사용하면 정확도가 높아지지만,
                     //               계산 비용과 메모리 사용량이 증가합니다.
    m        int     // 비트 배열의 크기: 더 큰 배열은 정확도를 높이지만, 메모리 사용량이 증가합니다.
}

// NewBloomFilter 함수는 Bloom Filter를 초기화합니다.
// n: 예상되는 요소의 수
// p: 허용 가능한 거짓 양성 비율
func NewBloomFilter(n int, p float64) *BloomFilter {
    // 비트 배열의 최적 크기(m)를 계산합니다.
    // 1. 거짓 양성 확률 계산
    //     
    //     Bloom Filter에서 특정 비트가 0으로 남아있을 확률:
    //     - 각 해시 함수가 특정 비트를 1로 남길 확률: 
    // 
    //         1/m
    // 
    //     - 각 해시 함수가 특정 비트를 0으로 남길 확률: 
    //     
    //         (1 - 1/m)
    //     
    //     - k개의 해시 함수가 모두 해당 비트를 0으로 남길 확률(k번 곱한다): 
    //
    //         (1 - 1/m)^k
    //  
    //     - n개의 요소가 모두 해당 비트를 0으로 남길 확률(n번 곱한다).
    //       즉, Bloom Filter에서 특정 비트가 0으로 남아있을 확률: 
    //           
    //         ((1 - 1/m)^k)^n = (1 - 1/m)^(kn)
    //     
    //     Bloom Filter에서 특정 비트가 1로 설정될 확률:
    //     
    //         1 - (1 - 1/m)^(kn)
    //     
    //     거짓 양성 확률 p는 실제로는 존재하지 않는 요소가 Bloom Filter에서 마치 존재하는 것처럼 잘못 판단되는 확률을 의미합니다.
    //     이는 새로 검사하는 요소가 k개의 해시 함수에 의해 선택된 모든 비트들이 이미 1로 설정되어 있는 경우 발생합니다.
    //       
    //         거짓 양성 확률 p = (1 - (1 - 1/m)^(kn))^k
    // 
    // 2. 근사
    // 
    //     m이 충분히 크고 1/m이 작다고 가정하면, 다음 근사를 사용할 수 있습니다: (1 - 1/m)^(kn) ≈ e^(-kn/m)
    //     이를 적용하면: p ≈ (1 - e^(-kn/m))^k
    // 
    // 3. 최적화
    // 
    //     최적의 k 값은 m/n * ln(2)입니다(이는 별도로 유도됩니다). 이는 거짓 양성 확률을 최소화하는 k 값입니다.
    //     이를 대입하면:
    //
    //         p ≈ (1 - e^(-ln(2)))^(m/n * ln(2))
    //         = (1/2)^(m/n * ln(2))
    //         = (0.6185)^(m/n)
    // 
    // 4. m에 대해 해결
    // 
    //     위 식을 m에 대해 해결하면: m = -n * ln(p) / (ln(2)^2)
    // 
    // 공식: m = -n * ln(p) / (ln(2)^2)
    // 
    // 이 공식은 주어진 n과 p에 대해 최적의 비트 배열 크기를 제공합니다.
    //  자연로그를 계산하는 Go의 math.Log 함수를 사용합니다.
    m := int(math.Ceil(float64(-n) * math.Log(p) / math.Pow(math.Log(2), 2)))
    
    // 최적의 해시 함수 개수(k)를 계산합니다.
    // 공식: k = (m/n) * ln(2)
    // 이 공식은 주어진 m과 n에 대해 최적의 해시 함수 개수를 제공합니다.
    k := int(math.Round(float64(m) / float64(n) * math.Log(2)))
    
    // BloomFilter 구조체를 생성하고 초기화합니다.
    return &BloomFilter{
        bitArray: make([]bool, m),  // m 크기의 불리언 배열을 생성합니다. 초기값은 모두 false입니다.
        k:        k,                // 계산된 최적의 해시 함수 개수를 저장합니다.
        m:        m,                // 계산된 비트 배열의 크기를 저장합니다.
    }
}

// hash 메서드는 주어진 문자열에 대해 k개의 서로 다른 해시 값을 생성합니다.
func (bf *BloomFilter) hash(item string) []int {
    hashes := make([]int, bf.k)   // k개의 해시 값을 저장할 정수 슬라이스를 생성합니다.
    for i := 0; i < bf.k; i++ {   // k번 반복하여 각각 다른 해시 값을 생성합니다.
        h := fnv.New64a()         // FNV-1a 해시 알고리즘 객체를 생성합니다.
                                  // FNV-1a는 빠르고 균일한 분포를 제공하는 해시 알고리즘입니다.
        h.Write([]byte(item))     // 문자열을 바이트 슬라이스로 변환하여 해시 함수에 입력합니다.
        
        // 해시 값을 계산하고, 비트 배열의 크기로 모듈러 연산을 수행합니다.
        // 이는 해시 값을 0부터 m-1 사이의 값으로 매핑합니다.
        // 'i*7'을 더하는 이유는 각 해시 함수가 서로 다른 값을 생성하도록 하기 위함입니다.
        // 7은 소수이며, 이를 통해 해시 값들이 고르게 분포되도록 합니다.
        hashes[i] = int(h.Sum64()%uint64(bf.m)) + i*7
    }
    return hashes  // 생성된 k개의 해시 값을 반환합니다.
}

// Add 메서드는 주어진 요소를 Bloom Filter에 추가합니다.
func (bf *BloomFilter) Add(item string) {
    hashes := bf.hash(item)       // 주어진 요소에 대해 k개의 해시 값을 생성합니다.
    for _, h := range hashes {    // 각 해시 값에 대해
        bf.bitArray[h%bf.m] = true // 해당 인덱스의 비트를 true로 설정합니다.
        // 모듈러 연산(h%bf.m)을 사용하는 이유는 해시 값을 비트 배열의 유효한 인덱스로 매핑하기 위함입니다.
    }
}

// Test 메서드는 주어진 요소가 Bloom Filter에 존재할 가능성이 있는지 검사합니다.
func (bf *BloomFilter) Test(item string) bool {
    hashes := bf.hash(item)       // 주어진 요소에 대해 k개의 해시 값을 생성합니다.
    for _, h := range hashes {    // 각 해시 값에 대해
        if !bf.bitArray[h%bf.m] { // 해당 인덱스의 비트가 false라면
            return false          // 요소가 존재하지 않음을 확신할 수 있으므로 false를 반환합니다.
        }
    }
    // 모든 해시 값에 해당하는 비트가 true라면, 요소가 존재할 가능성이 있습니다.
    // 단, 이는 거짓 양성(false positive)의 가능성을 내포하고 있습니다.
    return true
}

func main() {
    // 사용 예시
    n := 1000           // 예상되는 요소의 개수
    p := 0.01           // 허용 가능한 거짓 긍정 비율
    bf := NewBloomFilter(n, p) // Bloom Filter를 초기화합니다.

    bf.Add("foo")  // "foo"를 Bloom Filter에 추가합니다.
    bf.Add("bar")  // "bar"를 Bloom Filter에 추가합니다.

    println(bf.Test("foo")) // 예상 출력: true (추가한 요소이므로)
    println(bf.Test("baz")) // 예상 출력: false (추가하지 않은 요소이므로)
}
```

```go
package main

import (
    "fmt"
    "hash/fnv"  // fnv는 Go의 표준 라이브러리에서 제공하는 해시 함수입니다.
    "math"      // math 패키지는 수학적 연산을 위해 사용됩니다.
)

// BloomFilter 구조체는 Bloom filter의 기본 구조를 정의합니다.
type BloomFilter struct {
    bitArray []bool      // 비트 배열: 각 요소가 추가되었는지 추적합니다.
    size     int         // 비트 배열의 크기
    hashFuncs []hash.Hash64  // 해시 함수들: 각 요소에 대해 여러 해시를 생성합니다.
}

// NewBloomFilter 함수는 새로운 Bloom filter를 생성합니다.
// size: 비트 배열의 크기, numHash: 사용할 해시 함수의 수
func NewBloomFilter(size int, numHash int) *BloomFilter {
    bf := &BloomFilter{
        bitArray: make([]bool, size),  // size 크기의 bool 슬라이스를 생성합니다.
        size:     size,
        hashFuncs: make([]hash.Hash64, numHash),  // numHash 개수만큼의 해시 함수를 저장할 슬라이스를 생성합니다.
    }
    // 각 해시 함수를 초기화합니다. 여기서는 fnv.New64()를 사용하지만, 실제로는 다양한 해시 함수를 사용하는 것이 좋습니다.
    for i := 0; i < numHash; i++ {
        bf.hashFuncs[i] = fnv.New64()
    }
    return bf
}

// Add 메서드는 Bloom filter에 항목을 추가합니다.
func (bf *BloomFilter) Add(item string) {
    for _, h := range bf.hashFuncs {
        h.Reset()  // 해시 함수를 초기화합니다.
        h.Write([]byte(item))  // 항목을 바이트 슬라이스로 변환하여 해시 함수에 입력합니다.
        index := h.Sum64() % uint64(bf.size)  // 해시 값을 비트 배열의 인덱스로 변환합니다.
        bf.bitArray[index] = true  // 해당 인덱스의 비트를 1로 설정합니다.
    }
}

// Contains 메서드는 항목이 Bloom filter에 존재할 가능성을 확인합니다.
func (bf *BloomFilter) Contains(item string) bool {
    for _, h := range bf.hashFuncs {
        h.Reset()
        h.Write([]byte(item))
        index := h.Sum64() % uint64(bf.size)
        if !bf.bitArray[index] {
            return false  // 하나라도 0이면 항목이 없다고 확신할 수 있습니다.
        }
    }
    return true  // 모든 비트가 1이면 항목이 있을 가능성이 있습니다.
}

// FalsePositiveRate 메서드는 추정 오탐률을 계산합니다.
func (bf *BloomFilter) FalsePositiveRate(numItems int) float64 {
    k := float64(len(bf.hashFuncs))  // 해시 함수의 수
    m := float64(bf.size)            // 비트 배열의 크기
    n := float64(numItems)           // 추가된 항목의 수
    // 오탐률 공식: (1 - e^(-kn/m))^k
    return math.Pow(1 - math.Exp(-k*n/m), k)
}

func main() {
    // 1000 비트의 크기와 3개의 해시 함수를 가진 Bloom filter를 생성합니다.
    bf := NewBloomFilter(1000, 3)

    // 몇 가지 항목을 추가합니다.
    bf.Add("apple")
    bf.Add("banana")
    bf.Add("cherry")

    // 항목의 존재 여부를 확인합니다.
    fmt.Println("Contains 'apple':", bf.Contains("apple"))    // true가 예상됩니다.
    fmt.Println("Contains 'durian':", bf.Contains("durian")) // false가 예상되지만, 오탐의 가능성이 있습니다.

    // 추정 오탐률을 계산합니다.
    fmt.Printf("Estimated false positive rate: %.4f\n", bf.FalsePositiveRate(3))
}
```

## 블룸 필터의 응용 분야

> 거짓 양성의 가능성을 감수할 수 있는 상황들일 것
>
> - 초기 빠른 필터링이 중요함: 거짓 양성으로 인한 추가 검증 비용보다 전체적인 성능 향상이 더 큽니다.
> - 메모리 효율성: 완전한 데이터 세트를 저장하는 것보다 블룸 필터 사용이 훨씬 메모리 효율적입니다.
> - 거짓 음성 회피: 이러한 응용 분야에서는 거짓 음성(있는 것을 없다고 판단)이 더 큰 문제를 일으킬 수 있습니다.
> - 추가 검증 가능: 거짓 양성이 발생하더라도 추가 검증을 통해 문제를 해결할 수 있습니다.

1. 데이터베이스 시스템에서 디스크 접근 최소화:

    - 양성: 데이터가 실제로 디스크에 존재함
    - 음성: 데이터가 디스크에 존재하지 않음
    - 거짓 양성: 디스크에 없는 데이터를 있다고 판단
    - 거짓 음성: 디스크에 있는 데이터를 없다고 판단 (블룸 필터에서는 발생하지 않음)

    거짓 음성이 발생하면 실제로 존재하는 데이터를 찾지 못해 시스템 오류가 발생할 수 있습니다.

    거짓 양성 처리 방법:
    1. 블룸 필터가 "데이터가 존재한다"고 판단하면 실제 디스크 접근을 수행합니다.
    2. 실제로 데이터가 디스크에 없다면 (거짓 양성), 추가 처리 로직을 실행합니다.

    블룸 필터는 대부분의 "디스크에 없는 데이터"를 빠르게 필터링합니다.
    거짓 양성으로 인한 추가 디스크 접근이 발생하지만, 이는 전체 디스크 접근 횟수에 비해 적습니다.

    예를 들어, 1,000,000개의 데이터 요청 중 200,000개만 실제로 디스크에 존재하고,
    블룸 필터의 거짓 양성률이 1%라고 가정합니다. 그 결과:

    - 800,000개의 디스크에 없는 데이터 중 약 8,000개가 거짓 양성으로 판단됩니다.
    - 나머지 792,000개는 정확하게 "디스크에 없음"으로 판단되어 792,000번의 불필요한 디스크 접근을 막습니다.
    - 208,000번의 디스크 접근이 발생합니다 (실제 존재하는 데이터 200,000 + 거짓 양성 8,000).
    - 거짓 양성으로 인한 8,000번의 추가 접근이 있지만, 792,000번의 불필요한 접근을 막아 전체적인 성능이 크게 향상됩니다.

2. 네트워크 라우터에서 루프 감지:

    - 양성: 네트워크 경로에 루프가 실제로 존재함
    - 음성: 네트워크 경로에 루프가 존재하지 않음
    - 거짓 양성: 루프가 없는 경로를 루프가 있다고 판단
    - 거짓 음성: 실제 루프를 감지하지 못함 (블룸 필터에서는 발생하지 않음)

    거짓 음성의 리스크:
    - 거짓 음성이 발생하면 네트워크 루프로 인한 패킷 폭주와 네트워크 마비가 발생할 수 있습니다.

    거짓 양성 처리 방법:
    1. 블룸 필터가 "루프가 존재한다"고 판단하면 추가적인 경로 검증을 수행합니다.
    2. 실제로 루프가 없다면 (거짓 양성), 정상적인 라우팅을 진행합니다.

    블룸 필터는 대부분의 "루프가 없는 경로"를 빠르게 필터링합니다.
    거짓 양성으로 인한 추가 검증이 발생하지만, 이는 전체 경로 검사 횟수에 비해 적습니다.

    예를 들어, 100,000개의 네트워크 경로 중 100개만 실제로 루프가 있고,
    블룸 필터의 거짓 양성률이 0.1%라고 가정합니다. 그 결과:

    - 99,900개의 루프가 없는 경로 중 약 100개가 거짓 양성으로 판단됩니다.
    - 나머지 99,800개는 정확하게 "루프 없음"으로 판단되므로, 99,800개의 경로에 대해 불필요한 추가 검증을 피합니다.
    - 200개의 경로에 대해 추가 검증이 필요합니다 (실제 루프 100 + 거짓 양성 100).
    - 거짓 양성으로 인한 100번의 추가 검증이 있지만, 99,800개의 경로를 빠르게 처리하여 전체적인 라우팅 성능이 크게 향상됩니다.

3. 웹 크롤러에서 이미 방문한 URL 필터링:

    - 양성: URL을 이미 방문함
    - 음성: URL을 아직 방문하지 않음
    - 거짓 양성: 방문하지 않은 URL을 방문했다고 판단
    - 거짓 음성: 방문한 URL을 방문하지 않았다고 판단 (블룸 필터에서는 발생하지 않음)

    거짓 음성의 리스크:
    - 거짓 음성이 발생하면 같은 페이지를 반복해서 크롤링하여 리소스 낭비와 웹사이트에 불필요한 부하를 줄 수 있습니다.

    거짓 양성 처리 방법:
    1. 블룸 필터가 "URL을 방문했다"고 판단하면 해당 URL을 크롤링 대상에서 제외합니다.
    2. 이로 인해 일부 URL을 방문하지 못할 수 있지만, 크롤링 효율성이 크게 향상됩니다.

    블룸 필터는 대부분의 "방문하지 않은 URL"을 빠르게 식별합니다.
    거짓 양성으로 인해 일부 URL을 놓치지만, 전체적인 크롤링 속도와 효율성이 크게 향상됩니다.

    예를 들어, 10,000,000개의 URL 중 3,000,000개를 이미 방문했고,
    블룸 필터의 거짓 양성률이 1%라고 가정합니다. 그 결과:

    - 7,000,000개의 방문하지 않은 URL 중 약 70,000개가 거짓 양성으로 판단됩니다.
    - 나머지 6,930,000개는 정확하게 "방문하지 않음"으로 판단되므로, 6,930,000개의 새로운 URL을 크롤링할 수 있습니다.
    - 70,000개의 URL을 놓치지만 (거짓 양성으로 인해), 3,000,000개의 이미 방문한 URL에 대한 불필요한 재방문을 피합니다.
    - 전체적으로 크롤링 효율성이 크게 향상되며, 웹사이트에 대한 부하도 줄어듭니다.

4. 스펠링 체커:

    - 양성: 단어가 사전에 존재함
    - 음성: 단어가 사전에 존재하지 않음
    - 거짓 양성: 올바른 단어를 잘못된 것으로 판단
    - 거짓 음성: 잘못된 단어를 올바른 것으로 판단 (블룸 필터에서는 발생하지 않음)

    거짓 음성의 리스크:
    - 거짓 음성이 발생하면 잘못된 철자를 감지하지 못해 문서의 품질이 저하될 수 있습니다.

    거짓 양성 처리 방법:
    1. 블룸 필터가 "단어가 사전에 없다"고 판단하면 사용자에게 잠재적인 철자 오류로 표시합니다.
    2. 사용자는 이를 무시하거나 수정할 수 있습니다.

    블룸 필터는 대부분의 올바른 단어를 빠르게 확인합니다.
    거대한 사전 대신 작은 크기의 블룸 필터로 효율적인 철자 검사가 가능합니다.

    예를 들어, 100,000개의 고유한 단어를 포함하는 문서를 검사하고,
    블룸 필터의 거짓 양성률이 0.1%라고 가정합니다. 그 결과:

    - 100,000개의 단어 중 약 100개가 거짓 양성으로 판단됩니다.
    - 나머지 99,900개는 정확하게 판단되므로, 99,900개의 단어를 정확하게 검사합니다.
    - 100개의 올바른 단어가 잠재적 오류로 표시되지만, 사용자가 쉽게 무시할 수 있습니다.
    - 100,000개의 단어를 포함한 완전한 사전 대신 10MB 크기의 블룸 필터로 효율적인 검사가 가능합니다.

5. 캐시 시스템에서 키 존재 여부 빠른 확인:

    - 양성: 키가 캐시에 존재함
    - 음성: 키가 캐시에 존재하지 않음
    - 거짓 양성: 캐시에 없는 키를 있다고 잘못 판단
    - 거짓 음성: 캐시에 있는 키를 없다고 판단 (블룸 필터에서는 발생하지 않음)

    거짓 음성의 리스크:
    - 거짓 음성이 발생하면 캐시에 실제로 존재하는 데이터를 찾지 못해 불필요한 데이터베이스 접근이 발생할 수 있습니다.

    거짓 양성 처리 방법:
    1. 블룸 필터가 "키가 존재한다"고 판단하면 실제 캐시 조회를 수행합니다.
    2. 이때 실제로 키가 캐시에 없다면 (거짓 양성), 데이터베이스에서 데이터를 가져와야 합니다.

    블룸 필터는 대부분의 "캐시에 없는 키"를 빠르게 필터링합니다.
    거짓 양성으로 인한 추가 캐시 조회가 발생하지만, 이는 전체 캐시 미스 횟수에 비해 적습니다.

    예를 들어, 1,000,000개의 키 중 100,000개가 캐시에 있고,
    블룸 필터의 거짓 양성률이 1%라고 가정합니다. 그 결과:

    - 캐시에 없는 900,000개의 키 중 1%인 약 9,000개가 거짓 양성으로 판단됩니다.
    - 나머지 891,000개는 정확하게 "캐시에 없음"으로 판단되므로, 891,000번의 불필요한 캐시 접근을 막습니다.
    - 109,000번의 캐시 조회가 발생합니다 (실제 캐시 히트 100,000 + 거짓 양성 9,000).
    - 거짓 양성으로 인한 9,000번의 추가 조회가 있지만, 891,000번의 불필요한 접근을 막아 전체적인 성능이 향상됩니다.

## 블룸 필터의 한계

- **False Positive**:

    블룸 필터는 False Positive를 발생시킬 수 있습니다.
    즉, 존재하지 않는 요소를 있다고 잘못 판단할 수 있습니다.
    따라서 거짓 양성의 가능성을 감수할 수 있는 상황에서 사용되어야 합니다.

- **삭제 불가**

    블룸 필터에서는 요소를 삭제하는 것이 불가능합니다.
    이는 *요소를 제거할 때 다른 요소들의 위치 정보도 손상시킬 수 있기 때문*입니다.
    이러한 문제를 해결하기 위해 Counting Bloom Filter와 같은 변형이 제안되었습니다.

## 심화 학습 내용 및 관련 개념

- **Counting Bloom Filter**: 요소의 삽입과 삭제가 가능한 블룸 필터의 변형.
- **Cuckoo Filter**: 블룸 필터의 False Positive 문제를 해결하고, 더 높은 성능을 제공하는 변형.
- **Spectral Bloom Filter**: 주파수 정보를 고려하여 요소가 얼마나 자주 발생하는지를 추적할 수 있는 블룸 필터의 변형.
