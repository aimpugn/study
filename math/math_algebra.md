# Math algebra

- [Math algebra](#math-algebra)
    - [1. 기초 대수학 (Elementary Algebra)](#1-기초-대수학-elementary-algebra)
        - [1.1. 수와 연산](#11-수와-연산)
            - [자연수 (Natural Numbers)](#자연수-natural-numbers)
            - [정수 (Integers)](#정수-integers)
                - [예제: 두 정수의 최대공약수(GCD) 계산](#예제-두-정수의-최대공약수gcd-계산)
                - [암호학에서의 역할](#암호학에서의-역할)
                - [해시 함수에서의 역할](#해시-함수에서의-역할)
                - [난수 생성에서의 역할](#난수-생성에서의-역할)
            - [유리수 (Rational Numbers)](#유리수-rational-numbers)
        - [1.2. 방정식과 부등식](#12-방정식과-부등식)
        - [1.3. 다항식](#13-다항식)
        - [1.4. 인수분해](#14-인수분해)
        - [1.5. 유리식과 무리식](#15-유리식과-무리식)
    - [2. 선형대수학 (Linear Algebra)](#2-선형대수학-linear-algebra)
        - [2.1. 벡터와 벡터 공간](#21-벡터와-벡터-공간)
        - [2.2. 행렬과 행렬 연산](#22-행렬과-행렬-연산)
        - [2.3. 선형 변환](#23-선형-변환)
        - [2.4. 고유값과 고유벡터](#24-고유값과-고유벡터)
        - [2.5. 내적 공간](#25-내적-공간)
        - [2.6. 직교성과 정규직교기저](#26-직교성과-정규직교기저)
        - [2.7. 행렬 분해 (LU, QR, SVD 등)](#27-행렬-분해-lu-qr-svd-등)
    - [3. 추상대수학 (Abstract Algebra)](#3-추상대수학-abstract-algebra)
        - [3.1. 군론 (Group Theory)](#31-군론-group-theory)
            - [3.1.1. 군의 정의와 성질](#311-군의-정의와-성질)
            - [3.1.2. 부분군과 정규부분군](#312-부분군과-정규부분군)
            - [3.1.3. 준동형사상과 동형사상](#313-준동형사상과-동형사상)
            - [3.1.4. 군 작용](#314-군-작용)
            - [3.1.5. Sylow 정리](#315-sylow-정리)
        - [3.2. 환론 (Ring Theory)](#32-환론-ring-theory)
            - [3.2.1. 환의 정의와 성질](#321-환의-정의와-성질)
            - [3.2.2. 이데알과 몫환](#322-이데알과-몫환)
            - [3.2.3. 정역과 체](#323-정역과-체)
            - [3.2.4. 다항식 환](#324-다항식-환)
        - [3.3. 체론 (Field Theory)](#33-체론-field-theory)
            - [3.3.1. 체의 정의와 성질](#331-체의-정의와-성질)
            - [3.3.2. 체의 확대](#332-체의-확대)
            - [3.3.3. 갈루아 이론](#333-갈루아-이론)
        - [3.4. 모듈과 대수 (Modules and Algebras)](#34-모듈과-대수-modules-and-algebras)
            - [3.4.1. 모듈의 정의와 성질](#341-모듈의-정의와-성질)
            - [3.4.2. 대수의 정의와 성질](#342-대수의-정의와-성질)
    - [4. 범주론 (Category Theory)](#4-범주론-category-theory)
        - [4.1. 범주의 정의와 예](#41-범주의-정의와-예)
        - [4.2. 함자 (Functors)](#42-함자-functors)
        - [4.3. 자연변환 (Natural Transformations)](#43-자연변환-natural-transformations)
        - [4.4. 극한과 여극한 (Limits and Colimits)](#44-극한과-여극한-limits-and-colimits)
        - [4.5. 아벨 범주 (Abelian Categories)](#45-아벨-범주-abelian-categories)
    - [5. 대수기하학 (Algebraic Geometry)](#5-대수기하학-algebraic-geometry)
        - [5.1. 아핀 대수다양체 (Affine Algebraic Varieties)](#51-아핀-대수다양체-affine-algebraic-varieties)
        - [5.2. 사영 대수다양체 (Projective Algebraic Varieties)](#52-사영-대수다양체-projective-algebraic-varieties)
        - [5.3. 대수곡선 (Algebraic Curves)](#53-대수곡선-algebraic-curves)
        - [5.4. 대수곡면 (Algebraic Surfaces)](#54-대수곡면-algebraic-surfaces)
        - [5.5. 스킴 이론 (Scheme Theory)](#55-스킴-이론-scheme-theory)
    - [6. 대수적 위상수학 (Algebraic Topology)](#6-대수적-위상수학-algebraic-topology)
        - [6.1. 기본군 (Fundamental Group)](#61-기본군-fundamental-group)
        - [6.2. 호몰로지 이론 (Homology Theory)](#62-호몰로지-이론-homology-theory)
        - [6.3. 코호몰로지 이론 (Cohomology Theory)](#63-코호몰로지-이론-cohomology-theory)
        - [6.4. 호모토피 이론 (Homotopy Theory)](#64-호모토피-이론-homotopy-theory)
    - [7. 수론 (Number Theory)](#7-수론-number-theory)
        - [7.1. 초등 수론 (Elementary Number Theory)](#71-초등-수론-elementary-number-theory)
        - [7.2. 해석적 수론 (Analytic Number Theory)](#72-해석적-수론-analytic-number-theory)
        - [7.3. 대수적 수론 (Algebraic Number Theory)](#73-대수적-수론-algebraic-number-theory)
        - [7.4. 타원곡선 (Elliptic Curves)](#74-타원곡선-elliptic-curves)
    - [8. 표현론 (Representation Theory)](#8-표현론-representation-theory)
        - [8.1. 군의 표현 (Group Representations)](#81-군의-표현-group-representations)
        - [8.2. 리 대수의 표현 (Lie Algebra Representations)](#82-리-대수의-표현-lie-algebra-representations)
        - [8.3. 호프 대수의 표현 (Hopf Algebra Representations)](#83-호프-대수의-표현-hopf-algebra-representations)
    - [9. 호모로지 대수 (Homological Algebra)](#9-호모로지-대수-homological-algebra)
        - [9.1. 복체와 호모로지 (Complexes and Homology)](#91-복체와-호모로지-complexes-and-homology)
        - [9.2. 유도 함자 (Derived Functors)](#92-유도-함자-derived-functors)
        - [9.3. 스펙트럴 수열 (Spectral Sequences)](#93-스펙트럴-수열-spectral-sequences)
    - [10. 교환대수 (Commutative Algebra)](#10-교환대수-commutative-algebra)
        - [10.1. 노터 환과 아르틴 환 (Noetherian and Artinian Rings)](#101-노터-환과-아르틴-환-noetherian-and-artinian-rings)
        - [10.2. 정역 분해 (Integral Domain Factorization)](#102-정역-분해-integral-domain-factorization)
        - [10.3. 국소화 (Localization)](#103-국소화-localization)
        - [10.4. 완비화 (Completion)](#104-완비화-completion)
    - [11. 비교환대수 (Non-commutative Algebra)](#11-비교환대수-non-commutative-algebra)
        - [11.1. 분할대수 (Division Algebras)](#111-분할대수-division-algebras)
        - [11.2. 리 대수 (Lie Algebras)](#112-리-대수-lie-algebras)
        - [11.3. C\*-대수 (C\*-algebras)](#113-c-대수-c-algebras)

## 1. 기초 대수학 (Elementary Algebra)

### 1.1. 수와 연산

기초 대수학의 수와 연산 부분은 수학과 프로그래밍의 기본을 이루는 핵심 개념입니다.
여기서는 자연수, 정수, 유리수의 성질과 이들의 연산을 살펴보고, 이를 Golang을 사용하여 구현해보겠습니다.

#### 자연수 (Natural Numbers)

자연수는 $1, 2, 3, \dots$ 같은 양의 정수를 의미합니다.
자연수의 핵심 성질 중 하나는 연속성입니다. 즉, 어떤 자연수 다음에는 항상 그보다 1 큰 수가 존재합니다.
이 성질은 반복문과 재귀 알고리즘의 기초가 됩니다.

프로그래밍에서 자연수는 주로 반복문의 카운터, 슬라이스의 인덱스, 또는 양의 정수 값을 나타내는 데 사용됩니다.

이처럼 수학적 성질을 이해하고 활용하면 더 효율적인 알고리즘을 설계할 수 있습니다.
이는 대규모 데이터를 다루는 현대 프로그래밍에서 매우 중요한 기술입니다.

다음은 1부터 n까지의 자연수 합을 계산하는 Golang 함수입니다:

```go
package main

import (
    "fmt"
)

func sumOfNaturalNumbers(n int) int {
    sum := 0
    for i := 1; i <= n; i++ {
        sum += i
    }
    return sum
}

func main() {
    fmt.Println(sumOfNaturalNumbers(10)) // 출력: 55
}
```

이 함수는 가우스의 합 공식 $(n(n+1))/2$를 직접 구현한 것입니다. 이 공식과 알고리즘은 자연수의 여러 기본적인 성질을 이용합니다:

1. 순서성: 자연수는 1부터 시작하여 순서대로 증가합니다.
2. 덧셈의 교환법칙: 더하는 순서를 바꿔도 결과는 같습니다.
3. 덧셈의 결합법칙: 어떤 순서로 덧셈을 수행해도 결과는 같습니다.

이 알고리즘은 이러한 성질들을 다음과 같이 활용합니다:

1. 순서성: `for i := 1; i <= n; i++`는 1부터 n까지의 모든 자연수를 순서대로 탐색합니다.
2. 덧셈의 교환법칙과 결합법칙: `sum += i`는 각 숫자를 누적하여 더합니다. 이 과정에서 덧셈의 순서는 중요하지 않습니다.

가우스의 합 공식을 직접 구현하는 대신 수학적 공식을 사용하면 더 효율적인 알고리즘을 만들 수 있습니다:

```go
func sumOfNaturalNumbersOptimized(n int) int {
    return n * (n + 1) / 2
}
```

이 최적화된 버전은 가우스가 발견한 수학적 패턴을 이용합니다. 이 패턴은 다음과 같은 자연수의 성질에 기반합니다:

1. 대칭성: 1부터 n까지의 합은 n부터 1까지의 합과 같습니다.
2. 곱셈의 분배법칙: n * (n + 1)은 n^2 + n과 같습니다.

가우스는 이 성질들을 이용하여 1부터 n까지의 합이 항상 $(n(n+1))/2$임을 증명했습니다.
이 공식을 사용하면 반복문 없이 단 한 번의 계산으로 결과를 얻을 수 있어, 시간 복잡도가 O(n)에서 O(1)로 개선됩니다.

#### 정수 (Integers)

정수는 양의 정수, 0, 음의 정수를 모두 포함하는 수 체계입니다.
즉, $\ldots, -2, -1, 0, 1, 2, \ldots$ 와 같은 수들의 집합입니다.
정수는 자연수보다 더 넓은 범위의 수를 다룰 수 있게 해주며, 뺄셈 연산을 자유롭게 할 수 있게 합니다.

정수의 핵심 성질에는 다음과 같은 것들이 있습니다:
- **덧셈과 뺄셈에 대해 닫혀있음**: 두 정수를 더하거나 빼면 항상 정수가 나옵니다.
- **곱셈에 대해 닫혀있음**: 두 정수를 곱하면 항상 정수가 나옵니다.
- **순서 관계**: 임의의 두 정수는 항상 크기를 비교할 수 있습니다.

정수론의 개념들은 현대 컴퓨터 과학의 핵심적인 알고리즘과 데이터 구조의 기반이 되고 있습니다.
정수의 기본적인 성질을 이해하고 활용하면, 다양한 수학적 문제를 효율적으로 해결할 수 있는 알고리즘을 개발할 수 있습니다.
대부분의 프로그래밍 언어에서 기본 데이터 타입으로 제공되며, 다양한 용도로 활용됩니다:
- 카운터
- 인덱스
- 크기
- ID 등

암호학, 해시 함수, 난수 생성 등의 분야에서 중요한 역할을 하며, 실제 시스템의 보안, 효율성, 신뢰성을 보장하는 데 직접적으로 기여하고 있습니다.
- 암호학: RSA 암호화와 타원곡선 암호화에서 큰 소수의 성질을 이용하여 안전한 키 생성과 메시지 암호화를 수행합니다.
- 해시 함수: 모듈러 연산과 소수를 활용하여 효율적이고 균일한 해시 분포를 얻습니다.
- 난수 생성: 선형 합동 생성기와 메르센 트위스터 등에서 정수론적 개념을 활용하여 긴 주기의 난수를 생성합니다.

##### 예제: 두 정수의 최대공약수(GCD) 계산

아래 함수는 유클리드 알고리즘을 구현한 것으로, 정수의 여러 중요한 성질을 활용합니다:

1. 나눗셈의 성질:

    임의의 두 정수 $a$와 $b$에 대해 다음과 같이 표현할 수 있습니다.

    $a = (b\times{q}) + r$

    - $q$: 몫
    - $r$: 나머지. ($0 \leq{r <{|b|}}$)

2. 최대공약수의 성질

    $gcd(a,b) = gcd(b, a\mod{b})$, 이때 $b \neq 0$

3. 교환법칙:

    $gcd(a,b) = gcd(b,a)$

```go
package main

import (
    "fmt"
    "math"
)

func gcd(a, b int) int {
    // 반복문은 b가 0이 될 때까지 계속됩니다. 
    // 이는 나눗셈의 성질을 이용하여 두 수의 최대공약수를 점점 작은 수의 최대공약수 문제로 축소시키는 과정입니다.
    for b != 0 {
        // gcd(a,b) = gcd(b, a mod b) 성질을 직접 구현한 것입니다
        a, b = b, a%b
    }
    // 최종적으로 b가 0이 되면, a가 최대공약수가 됩니다.
    return int(math.Abs(float64(a)))
}

func main() {
    fmt.Println(gcd(48, 18)) // 출력: 6
}
```

이 유클리드 알고리즘의 시간 복잡도:
- $\mathcal{O}(\log(\min(a,b)))$

이 복잡도는 입력 값의 크기에 대해 로그 스케일로 증가하므로, 매우 큰 수에 대해서도 효율적으로 동작함을 의미합니다.
이는 각 단계에서 최소한 숫자의 크기가 절반 이상 줄어들기 때문입니다.

1. 알고리즘의 각 단계 분석:

    각 단계에서 $a$를 $b$로 나눈 나머지를 계산합니다:
    - $a\mod{b}$

    이후 $b$가 새로운 $a$가 되고, 나머지가 새로운 $b$가 됩니다.

2. 나머지 연산의 성질:

    나머지는 항상 나누는 수보다 작습니다.
    - $(a \mod{b}) < b$

    이는 새로운 b가 이전 b의 절반보다 항상 작다는 것을 의미합니다:
    - $(a \mod{b}) < b/2$

3. 최악의 경우 분석:

    최악의 경우, 각 단계에서 새로운 $b$가 이전 $b$의 정확히 절반이라고 가정합니다.
    이 경우, $k$번의 단계 후 $b$의 값은 다음과 같이 됩니다:
    - $b\times{(1/2)^{k}}$

4. 종료 조건까지의 단계 수 계산:

    알고리즘은 b가 0이 될 때 종료됩니다.
    - $b\times{(\frac{1}{2})^{k}} < 1$ (즉, b가 1보다 작아질 때)

    양변에 로그를 취하면
    - $\log{b} - k\times{\log{2}} < 0$

    $k$에 대해 정리하면:
    - $k > \frac{\log{b}}{\log{2}} = \log_{2}{b}$

5. 시간 복잡도 도출:

   위 분석에 따르면, 알고리즘이 종료되기 위해 필요한 최대 단계 수는 $log_{2}{b}$입니다.

   로그의 밑을 변경해도 복잡도의 차수는 변하지 않으므로, 이는 다음과 같습니다.
   - $\mathcal{O}(\log(b))$

   $a$와 $b$ 중 작은 수를 기준으로 하므로, 최종적으로
   - $\mathcal{O}(\log(\min(a,b)))$

6. 구체적인 예시:

    $a = 1000, b = 24$의 경우를 고려해봅시다.

    1. 1000 % 24 = 16  (1000 = 41 * 24 + 16)
    2. 24 % 16 = 8     (24 = 1 * 16 + 8)
    3. 16 % 8 = 0      (16 = 2 * 8 + 0)

    총 3단계만에 종료되었습니다.
    $log_{2}{24}\approx{4.58}$이므로, 이론적 상한과 일치함을 볼 수 있습니다.

유클리드 알고리즘은 단순히 최대공약수를 구하는 것 외에도 다양한 응용이 있습니다:

1. 분수의 약분: 분자와 분모의 최대공약수로 나누어 기약분수를 만들 수 있습니다.
2. 디오판토스 방정식 해결: $ax + by = c$ 형태의 방정식 해결에 사용됩니다.
3. 모듈러 역원 계산: 암호학에서 중요하게 사용되는 개념입니다.

##### 암호학에서의 역할

- RSA 암호화
    RSA 알고리즘은 큰 소수의 곱으로 이루어진 합성수의 인수분해가 어렵다는 사실을 이용합니다.

    두 큰 소수 $p$와 $q$의 곱 $n = p\times{q}$를 알더라도, $n$으로부터 $p$와 $q$를 효율적으로 찾는 알고리즘이 현재까지 알려져 있지 않기 때문입니다.

    공개키 $e$와 비밀키 $d$를 선택할 때 다음 조건을 사용합니다:

    - $(e\times{d})\mod{\phi(n)} = 1$

    여기서 $\phi(n)$은 *오일러 토티언트 함수*로, 정수론의 중요한 개념입니다.

    예를 들어, 메시지 `m`을
    - 암호화할 때 $c = m^{e}\mod{n}$을 계산
    - 복호화할 때 $m = c^d\mod{n}$을 계산

    이 과정에서 *모듈러 지수 연산*이 핵심적인 역할을 합니다.

- 타원곡선 암호화: 타원곡선 상의 점들의 덧셈 연산이 정수의 곱셈과 유사한 성질을 가진다는 점을 이용합니다.
    - 왜: 타원곡선 상의 이산로그 문제가 일반적인 이산로그 문제보다 더 어렵다고 알려져 있어, 더 짧은 키로 동일한 수준의 보안을 제공할 수 있기 때문입니다.
    - 어떻게: y^2 = x^3 + ax + b (mod p) 형태의 방정식을 사용하며, 여기서 p는 큰 소수입니다. 점의 덧셈 연산은 정수론의 개념을 바탕으로 정의됩니다.
    - 구체적 예: 공개키 Q = dG를 생성할 때, d는 비밀키이고 G는 타원곡선 상의 기준점입니다. 이 연산은 타원곡선 상에서 점 G를 d번 더하는 것과 같습니다.

##### 해시 함수에서의 역할

   a) 모듈러 연산: 많은 해시 함수들이 모듈러 연산을 사용하여 해시 값을 계산합니다.
      - 왜: 모듈러 연산은 큰 수를 작은 범위로 매핑하는 데 효과적이며, 결과의 분포가 균일하기 때문입니다.
      - 어떻게: h(x) = x mod m 형태의 간단한 해시 함수에서부터, 더 복잡한 형태의 모듈러 연산을 포함하는 해시 함수까지 다양하게 사용됩니다.
      - 구체적 예: 문자열 "hello"를 해시할 때, 각 문자의 ASCII 값을 더한 후 테이블 크기로 모듈러 연산을 수행할 수 있습니다. (104 + 101 + 108 + 108 + 111) mod 100 = 32

   b) 소수를 이용한 해시 테이블 크기 선정: 해시 테이블의 크기로 소수를 선택하면 충돌을 줄일 수 있습니다.
      - 왜: 소수를 모듈러로 사용하면 해시 값의 분포가 더 균일해지기 때문입니다.
      - 어떻게: 해시 테이블의 크기를 가장 가까운 소수로 선택합니다.
      - 구체적 예: 10000개의 원소를 저장할 해시 테이블을 만들 때, 크기를 10000이 아닌 10007(다음으로 큰 소수)로 선택합니다.

##### 난수 생성에서의 역할

   a) 선형 합동 생성기(Linear Congruential Generator): (a*X + c) mod m 형태의 수열을 이용하여 난수를 생성합니다.
      - 왜: 이 방법은 계산이 빠르고, 적절한 매개변수를 선택하면 긴 주기의 난수를 생성할 수 있기 때문입니다.
      - 어떻게: a, c, m을 적절히 선택하여 최대 주기(m)의 난수 수열을 생성합니다.
      - 구체적 예: X_{n+1} = (1664525* X_n + 1013904223) mod 2^32 (이것은 실제로 사용되는 매개변수입니다)

   b) 메르센 트위스터(Mersenne Twister): 메르센 소수(2^p - 1 형태의 소수)를 이용하여 긴 주기의 난수를 생성합니다.
      - 왜: 메르센 소수를 사용하면 매우 긴 주기(2^19937-1)의 난수를 효율적으로 생성할 수 있기 때문입니다.
      - 어떻게: 메르센 소수를 모듈러로 사용하여 선형 되먹임 시프트 레지스터(LFSR)를 구현합니다.
      - 구체적 예: MT19937 알고리즘은 2^19937-1을 주기로 사용하며, 이는 메르센 소수입니다.

#### 유리수 (Rational Numbers)

유리수는 두 정수의 비로 표현할 수 있는 수입니다. Golang에서는 유리수를 위한 내장 타입이 없으므로, 구조체를 사용하여 구현할 수 있습니다.

Golang 예제 (분수 구조체 구현):

```go
package main

import (
    "fmt"
    "math"
)

// Fraction 구조체는 분자와 분모를 가진 유리수를 표현합니다.
type Fraction struct {
    Numerator   int
    Denominator int
}

// NewFraction 함수는 새로운 Fraction을 생성하고 약분합니다.
func NewFraction(numerator, denominator int) Fraction {
    if denominator == 0 {
        panic("Denominator cannot be zero")
    }
    
    // 부호 정규화
    if denominator < 0 {
        numerator, denominator = -numerator, -denominator
    }
    
    // 최대공약수를 이용한 약분
    g := gcd(numerator, denominator)
    return Fraction{
        Numerator:   numerator / g,
        Denominator: denominator / g,
    }
}

// Add 메서드는 두 분수를 더합니다.
func (f Fraction) Add(other Fraction) Fraction {
    return NewFraction(
        f.Numerator*other.Denominator + other.Numerator*f.Denominator,
        f.Denominator*other.Denominator,
    )
}

// String 메서드는 분수를 문자열로 표현합니다.
func (f Fraction) String() string {
    return fmt.Sprintf("%d/%d", f.Numerator, f.Denominator)
}

func main() {
    f1 := NewFraction(1, 2)
    f2 := NewFraction(1, 3)
    sum := f1.Add(f2)
    fmt.Printf("%s + %s = %s\n", f1, f2, sum) // 출력: 1/2 + 1/3 = 5/6
}
```

이 예제는 유리수의 기본 연산인 덧셈을 구현하고 있습니다. 또한 최대공약수(GCD)를 이용하여 분수를 약분하는 과정을 보여줍니다. 이는 유리수의 기본 성질을 프로그래밍으로 구현한 것입니다.

### 1.2. 방정식과 부등식

### 1.3. 다항식

### 1.4. 인수분해

### 1.5. 유리식과 무리식

## 2. 선형대수학 (Linear Algebra)

### 2.1. 벡터와 벡터 공간

### 2.2. 행렬과 행렬 연산

### 2.3. 선형 변환

### 2.4. 고유값과 고유벡터

### 2.5. 내적 공간

### 2.6. 직교성과 정규직교기저

### 2.7. 행렬 분해 (LU, QR, SVD 등)

## 3. 추상대수학 (Abstract Algebra)

### 3.1. 군론 (Group Theory)

#### 3.1.1. 군의 정의와 성질

#### 3.1.2. 부분군과 정규부분군

#### 3.1.3. 준동형사상과 동형사상

#### 3.1.4. 군 작용

#### 3.1.5. Sylow 정리

### 3.2. 환론 (Ring Theory)

#### 3.2.1. 환의 정의와 성질

#### 3.2.2. 이데알과 몫환

#### 3.2.3. 정역과 체

#### 3.2.4. 다항식 환

### 3.3. 체론 (Field Theory)

#### 3.3.1. 체의 정의와 성질

#### 3.3.2. 체의 확대

#### 3.3.3. 갈루아 이론

### 3.4. 모듈과 대수 (Modules and Algebras)

#### 3.4.1. 모듈의 정의와 성질

#### 3.4.2. 대수의 정의와 성질

## 4. 범주론 (Category Theory)

### 4.1. 범주의 정의와 예

### 4.2. 함자 (Functors)

### 4.3. 자연변환 (Natural Transformations)

### 4.4. 극한과 여극한 (Limits and Colimits)

### 4.5. 아벨 범주 (Abelian Categories)

## 5. 대수기하학 (Algebraic Geometry)

### 5.1. 아핀 대수다양체 (Affine Algebraic Varieties)

### 5.2. 사영 대수다양체 (Projective Algebraic Varieties)

### 5.3. 대수곡선 (Algebraic Curves)

### 5.4. 대수곡면 (Algebraic Surfaces)

### 5.5. 스킴 이론 (Scheme Theory)

## 6. 대수적 위상수학 (Algebraic Topology)

### 6.1. 기본군 (Fundamental Group)

### 6.2. 호몰로지 이론 (Homology Theory)

### 6.3. 코호몰로지 이론 (Cohomology Theory)

### 6.4. 호모토피 이론 (Homotopy Theory)

## 7. 수론 (Number Theory)

### 7.1. 초등 수론 (Elementary Number Theory)

### 7.2. 해석적 수론 (Analytic Number Theory)

### 7.3. 대수적 수론 (Algebraic Number Theory)

### 7.4. 타원곡선 (Elliptic Curves)

## 8. 표현론 (Representation Theory)

### 8.1. 군의 표현 (Group Representations)

### 8.2. 리 대수의 표현 (Lie Algebra Representations)

### 8.3. 호프 대수의 표현 (Hopf Algebra Representations)

## 9. 호모로지 대수 (Homological Algebra)

### 9.1. 복체와 호모로지 (Complexes and Homology)

### 9.2. 유도 함자 (Derived Functors)

### 9.3. 스펙트럴 수열 (Spectral Sequences)

## 10. 교환대수 (Commutative Algebra)

### 10.1. 노터 환과 아르틴 환 (Noetherian and Artinian Rings)

### 10.2. 정역 분해 (Integral Domain Factorization)

### 10.3. 국소화 (Localization)

### 10.4. 완비화 (Completion)

## 11. 비교환대수 (Non-commutative Algebra)

### 11.1. 분할대수 (Division Algebras)

### 11.2. 리 대수 (Lie Algebras)

### 11.3. C*-대수 (C*-algebras)
