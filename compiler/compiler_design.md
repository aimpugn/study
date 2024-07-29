# Compiler design

- [Compiler design](#compiler-design)
    - [개요](#개요)
    - [중간 표현(IR)과 SSA(Static Single Assignment)](#중간-표현ir과-ssastatic-single-assignment)
        - [중간 표현(IR)이란?](#중간-표현ir이란)
        - [SSA(Static Single Assignment)란?](#ssastatic-single-assignment란)
        - [왜 SSA 형태의 중간 표현을 사용하는가?](#왜-ssa-형태의-중간-표현을-사용하는가)
        - [중간 표현(IR)과 SSA(Static Single Assignment)와 관련된 과목](#중간-표현ir과-ssastatic-single-assignment와-관련된-과목)

## 개요

1. 컴파일러 구조 개요
   1. 컴파일러의 단계
   2. 프론트엔드, 미들엔드, 백엔드의 역할

2. 프론트엔드 설계
   1. 어휘 분석기 (Lexical Analyzer)
   2. 구문 분석기 (Syntax Analyzer)
   3. 의미 분석기 (Semantic Analyzer)
   4. 심볼 테이블 관리

3. 미들엔드 설계
    1. 중간 표현 (Intermediate Representation)
        1. IR의 목적과 요구사항
        2. IR 형식의 선택
            - 추상 구문 트리 (Abstract Syntax Tree, AST)
            - 삼주소 코드 (Three-Address Code)
            - 사중식 (Quadruple)
            - 정적 단일 할당 형식 (Static Single Assignment, SSA)
        3. IR 설계 시 고려사항
        4. IR 생성 알고리즘
        5. IR과 프로그래밍 언어 설계
            1. IR과 언어 표현력
            2. IR과 타입 시스템
            3. IR과 메모리 모델
            4. IR과 병렬성/동시성 표현

    2. 데이터 흐름 분석 (Data Flow Analysis)

        변수의 값과 사용 패턴을 추적하여 다양한 최적화 기회를 식별합니다.
        불필요한 연산을 제거하고 코드의 효율성을 향상시킵니다.

        1. 도달 정의 분석 (Reaching Definitions Analysis)
        2. 활성 변수 분석 (Live Variable Analysis)
        3. 사용-정의 체인 (Use-Def Chains) 구축
        4. 정의-사용 체인 (Def-Use Chains) 구축
        5. 상수 전파 분석 (Constant Propagation Analysis)
        6. 사용되지 않는 코드 분석 (Dead Code Analysis)

    3. 제어 흐름 분석 (Control Flow Analysis)

        프로그램의 실행 경로를 분석하여 최적화 기회를 식별합니다.
        루프와 같은 중요한 프로그램 구조를 식별하여 특화된 최적화를 적용할 수 있게 합니다.

        1. 제어 흐름 그래프 (Control Flow Graph, CFG) 구축
        2. 기본 블록 (Basic Blocks) 식별
        3. 지배자 트리 (Dominator Tree) 생성
        4. 루프 분석 (Loop Analysis)
            1. 루프 식별
            2. 루프 불변 코드 탐지
            3. 귀납 변수 분석

    4. 정적 단일 할당 (Static Single Assignment, SSA)

        SSA 형식은 많은 기계 독립적 최적화를 더 효과적으로 만들지만, 모든 최적화가 SSA를 필요로 하는 것은 아닙니다.
        SSA는 주로 미들엔드에서 사용되며, 백엔드로 넘어가기 전에 일반적으로 SSA 형식에서 벗어납니다(SSA destruction).

        1. SSA의 정의와 특징
        2. SSA 형식으로의 변환
            1. 변환 알고리즘
            2. 변수 재명명 (Variable Renaming)
            3. φ(Phi) 함수 삽입
        3. SSA의 장점
            1. 데이터 흐름 분석 용이성
            2. 최적화 기회 증가
        4. SSA 기반 최적화 기법
            1. 상수 전파 (Constant Propagation)
            2. 사용되지 않는 코드 제거 (Dead Code Elimination)
            3. 공통 부분식 제거 (Common Subexpression Elimination)
        5. SSA의 한계와 확장
            1. SSA 폼 해체 (SSA Destruction)
            2. 확장된 SSA 형식 (e.g., Memory SSA)

    5. 중간 코드 최적화
        1. 기계 독립적 최적화

            미들엔드의 중간 코드 최적화 단계에 포함됩니다.
            이 최적화들은 특정 하드웨어에 종속되지 않고 중간 표현(IR)에서 수행됩니다.
            SSA 형식을 사용하면 이러한 최적화들의 효과가 더욱 증대될 수 있지만, SSA가 필수 조건은 아닙니다.

            1. 상수 폴딩과 전파 (Constant Folding and Propagation)
            2. 복사 전파 (Copy Propagation)
            3. 죽은 코드 제거 (Dead Code Elimination)
            4. 공통 부분식 제거 (Common Subexpression Elimination)
            5. 코드 이동 (Code Motion)
            6. 부분 중복 제거 (Partial Redundancy Elimination)
            7. 루프 최적화 (Loop Optimization)
                - 루프 불변 코드 이동 (Loop Invariant Code Motion)
                - 루프 펼치기 (Loop Unrolling)
            8. 함수 인라인화 (Function Inlining)

        2. 기계 종속적 최적화

            백엔드의 목적 코드 최적화 단계에 포함됩니다.
            이 최적화들은 특정 타겟 아키텍처의 특성을 고려하여 수행됩니다.

            1. 명령어 선택 (Instruction Selection)
            2. 명령어 스케줄링 (Instruction Scheduling)
            3. 레지스터 할당 (Register Allocation)

        3. 데이터 흐름 분석 기반 최적화
        4. 제어 흐름 기반 최적화
        5. 로컬 최적화
        6. 전역 최적화
        7. 인터프로시저 최적화

    6. 최적화 패스 관리
        1. 최적화 패스 순서 결정
        2. 최적화 간 상호작용 관리
        3. 디버그 정보 유지

4. 백엔드 설계
    1. 목표 기계 코드 생성
        1. 중간 표현에서 목표 기계 코드로의 변환
        2. 목표 아키텍처 특성 고려

    2. 명령어 선택
        1. 트리 패턴 매칭
        2. 동적 프로그래밍 기반 최적 선택
        3. 피킹(Peephole) 패턴 적용

    3. 명령어 스케줄링
        1. 기본 블록 내 스케줄링
        2. 전역 명령어 스케줄링
        3. 소프트웨어 파이프라이닝

    4. 레지스터 할당
        1. 그래프 색칠 알고리즘
        2. 선형 스캔 할당
        3. 스필 코드 생성

    5. 목적 코드 최적화
        1. 기계 종속적 최적화
            1. 명령어 선택 (Instruction Selection)
            2. 명령어 스케줄링 (Instruction Scheduling)
            3. 레지스터 할당 (Register Allocation)
            4. 피킹 최적화 (Peephole Optimization)
            5. 강도 감소 (Strength Reduction)
            6. 주소 계산 최적화 (Address Calculation Optimization)
        2. 프로세서 특화 최적화
            a. SIMD 명령어 활용
            b. 분기 예측 최적화
        3. 링커 수준 최적화
            a. 함수 인라이닝
            b. 코드 레이아웃 최적화
        4. 프로파일 기반 최적화 (PGO)

    6. 목적 파일 생성
        1. 재배치 정보 생성
        2. 심볼 테이블 생성
        3. 디버깅 정보 포함

    7. 아키텍처 특화 최적화
        1. 캐시 친화적 코드 생성
        2. 분기 예측 고려
        3. 특수 명령어 셋 활용 (예: AVX, SSE)

5. 특수 컴파일러 기술
   1. JIT (Just-In-Time) 컴파일러
   2. AOT (Ahead-Of-Time) 컴파일러
   3. 크로스 컴파일러

6. 컴파일러 인프라 설계
    1. 모듈화와 확장성
    2. 오류 처리 및 복구 메커니즘
    3. 최적화 프레임워크 (예: LLVM Pass)
    4. 프로그램 분석 프레임워크
        1. 추상 해석 (Abstract Interpretation)
        2. 포인트-투 분석 (Points-to Analysis)
        3. 별칭 분석 (Alias Analysis)
        4. 타입 추론 (Type Inference)

7. 컴파일러 테스팅 및 검증
    1. 단위 테스트 및 통합 테스트
    2. 회귀 테스트
    3. 정확성 검증 (Correctness Verification)

8. 성능 및 메트릭
    A. 컴파일 시간 최적화
    B. 생성된 코드의 성능 측정
    C. 코드 크기 vs 실행 속도 트레이드오프

9. 최신 컴파일러 기술 동향
   1. JIT (Just-In-Time) 컴파일러에서의 IR 활용
   2. LLVM과 MLIR (Multi-Level IR)
   3. 기계학습과 컴파일러 최적화
   4. 프로그램 합성과 IR
   5. 병렬 및 분산 컴파일
   6. 도메인 특화 언어(DSL) 지원

## 중간 표현(IR)과 SSA(Static Single Assignment)

### 중간 표현(IR)이란?

중간 표현은 *소스 코드와 목표 기계어 사이의 중간 단계 표현*입니다.

이는 다음과 같은 목적으로 사용됩니다:
- 컴파일러 최적화를 쉽게 적용할 수 있는 형태로 코드를 변환
- 다양한 소스 언어와 목표 아키텍처 사이의 추상화 계층 제공
- 코드 분석과 변환을 용이하게 함

```go
func add(a, b int) int {
    c := a + b
    return c
}
```

이 코드의 간단한 중간 표현은 다음과 같을 수 있습니다:

```plaintext
1: PARAM a
2: PARAM b
3: ADD temp1, a, b
4: ASSIGN c, temp1
5: RETURN c
```

### SSA(Static Single Assignment)란?

SSA는 *각 변수가 프로그램 내에서 정확히 한 번만 할당되는 특별한 형태의 중간 표현(Intermediate Representation, IR)*입니다.

SSA의 주요 특징은:
- 각 변수에 대한 정의가 유일함
- 변수 사용 전에 반드시 정의가 선행됨
- 제어 흐름이 합쳐지는 지점에서 φ(phi) 함수를 사용하여 여러 정의를 통합

SSA 형태로 변환하면 위의 중간 표현은 다음과 같이 바뀝니다:

```plaintext
1: a1 = PARAM
2: b1 = PARAM
3: temp1 = ADD a1, b1
4: c1 = ASSIGN temp1
5: RETURN c1
```

### 왜 SSA 형태의 중간 표현을 사용하는가?

SSA는 다음과 같은 이점을 제공합니다:

- 데이터 흐름 분석 용이성:
   *각 변수가 한 번만 할당*되므로, 변수의 정의와 사용 관계를 쉽게 추적할 수 있습니다.

- 최적화 기회 증가:
   불필요한 할당 제거, 상수 전파 등의 최적화를 더 쉽게 적용할 수 있습니다.

- 모호성 제거:
   각 변수 사용이 단일 정의와 연결되므로, 변수의 값이 어디서 오는지 명확합니다.

```go
func example(x int) int {
    y := x + 1
    if x > 10 {
        y = x - 1
    }
    return y
}
```

이 코드의 SSA 형태 중간 표현은 다음과 같을 수 있습니다:

```plaintext
1:  x1 = PARAM
2:  y1 = ADD x1, 1
3:  cond1 = GT x1, 10
4:  BR cond1, true_block, false_block
5: true_block:
6:  y2 = SUB x1, 1
7:  JMP merge_block
8: false_block:
9:  JMP merge_block
10: merge_block:
11: y3 = PHI(y1, y2)
12: RETURN y3
```

여기서 `y3 = PHI(y1, y2)`는 φ 함수로, 제어 흐름이 합쳐지는 지점에서 `y`의 값이 `y1` 또는 `y2` 중 하나가 될 수 있음을 나타냅니다.

이러한 SSA 형태는 다음과 같은 최적화를 용이하게 합니다:

1. 불필요한 코드 제거: `y1`이 사용되지 않는다면 2번 라인을 제거할 수 있습니다.
2. 상수 전파: `x1`이 상수라면, 그 값을 `y1`과 `y2` 계산에 직접 사용할 수 있습니다.
3. 공통 부분식 제거: `x1 + 1`과 `x1 - 1`이 여러 번 계산된다면, 결과를 재사용할 수 있습니다.

또한, SSA는 포인터 분석, 별칭 분석 등 고급 컴파일러 최적화 기법의 기반이 됩니다.
이를 통해 더 효율적인 코드 생성이 가능해지며, 결과적으로 프로그램의 성능이 향상됩니다.

### 중간 표현(IR)과 SSA(Static Single Assignment)와 관련된 과목

1. 컴파일러 설계 (Compiler Design)
    - 교과 내용:
        - 컴파일러의 구조
        - 어휘 분석과 구문 분석
        - 의미 분석
        - 중간 코드 생성
        - 코드 최적화
        - 코드 생성

    - 관련 챕터:
        - "중간 표현" (Intermediate Representations)
        - "SSA와 데이터 흐름 분석" (SSA and Data Flow Analysis)

2. 프로그래밍 언어 이론 (Programming Language Theory)
    - 교과 내용:
        - 형식 의미론 (Formal Semantics)
        - 타입 시스템
        - 프로그램 분석
        - 언어 설계 원칙

    - 관련 챕터:
        - "프로그램 표현과 분석" (Program Representation and Analysis)

3. 소프트웨어 공학 (Software Engineering)
    - 교과 내용:
        - 소프트웨어 개발 생명 주기
        - 소프트웨어 아키텍처
        - 코드 품질과 리팩토링
        - 정적 분석 도구

    - 관련 챕터:
        - "코드 분석과 변환" (Code Analysis and Transformation)

4. 고급 알고리즘 (Advanced Algorithms)
    - 교과 내용:
        - 그래프 알고리즘
        - 최적화 기법
        - 프로그램 분석 알고리즘

    - 관련 챕터:
        - "프로그램 분석을 위한 알고리즘" (Algorithms for Program Analysis)

5. 컴퓨터 아키텍처 (Computer Architecture)
    - 교과 내용:
        - 프로세서 설계
        - 명령어 집합 아키텍처 (ISA)
        - 파이프라이닝과 병렬 처리

    - 관련 챕터:
        - "컴파일러 최적화와 하드웨어 상호작용" (Compiler Optimizations and Hardware Interactions)
