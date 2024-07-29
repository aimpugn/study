# parsing theory

- [parsing theory](#parsing-theory)
    - [서론](#서론)
    - [파싱의 기본 개념](#파싱의-기본-개념)
    - [정규 표현식과 유한 상태 오토마타 (DFA)](#정규-표현식과-유한-상태-오토마타-dfa)
        - [정규 표현식 (Regular Expressions)](#정규-표현식-regular-expressions)
        - [유한 상태 오토마타 (DFA, Deterministic Finite Automata)](#유한-상태-오토마타-dfa-deterministic-finite-automata)
    - [문맥 자유 언어와 푸시다운 오토마타 (PDA)](#문맥-자유-언어와-푸시다운-오토마타-pda)
        - [문맥 자유 언어 (Context-Free Languages)](#문맥-자유-언어-context-free-languages)
        - [푸시다운 오토마타 (PDA, Pushdown Automata)](#푸시다운-오토마타-pda-pushdown-automata)
    - [문맥 민감 언어와 선형 유한 비결정적 튜링 머신 (LBA)](#문맥-민감-언어와-선형-유한-비결정적-튜링-머신-lba)
        - [문맥 민감 언어 (Context-Sensitive Languages)](#문맥-민감-언어-context-sensitive-languages)
        - [선형 유한 비결정적 튜링 머신 (LBA, Linear Bounded Automaton)](#선형-유한-비결정적-튜링-머신-lba-linear-bounded-automaton)
    - [재귀 하향 언어와 튜링 머신](#재귀-하향-언어와-튜링-머신)
        - [재귀 하향 언어 (Recursively Enumerable Languages)](#재귀-하향-언어-recursively-enumerable-languages)
        - [튜링 머신 (Turing Machine)](#튜링-머신-turing-machine)
    - [실제 사례](#실제-사례)
        - [정규 표현식의 한계](#정규-표현식의-한계)
        - [HTML 파싱의 어려움](#html-파싱의-어려움)
        - [JSON 파싱](#json-파싱)
    - [고급 파싱 기법](#고급-파싱-기법)
        - [Lookahead 파싱](#lookahead-파싱)
        - [Parsing Expression Grammar (PEG)](#parsing-expression-grammar-peg)
        - [ISGLR](#isglr)
        - [Pika Parsing](#pika-parsing)
        - [Earley Parser](#earley-parser)
        - [CYK Algorithm](#cyk-algorithm)
    - [결론](#결론)
    - [참고 문헌 및 자료](#참고-문헌-및-자료)

## 서론

- 문서의 목적: 파싱 이론의 중요성과 실용적 적용 방안을 이해하고 학습.
- 문서의 범위: 정규 표현식, 문맥 자유 언어, 문맥 민감 언어, 재귀 하향 언어 등 다양한 파싱 기법과 실제 사례를 포괄.

## 파싱의 기본 개념

- 파싱의 정의: 문자열을 구조화된 데이터로 변환하는 과정.
- 파싱의 필요성: 프로그래밍 언어 컴파일러, 데이터 검증, 형식 언어 처리 등.

## 정규 표현식과 유한 상태 오토마타 (DFA)

### 정규 표현식 (Regular Expressions)

- 정의: 문자열 패턴 매칭을 위한 표현식.
- 사용 예: 문자열 검색, 데이터 검증.
- 한계: 복잡한 중첩 구조 처리 불가능.

### 유한 상태 오토마타 (DFA, Deterministic Finite Automata)

- 정의: 상태 전이를 통해 입력 문자열을 처리하는 오토마타.
- 구성 요소: 상태, 전이 함수, 시작 상태, 종료 상태.
- 장점: 빠른 처리 속도, 메모리 효율성.
- 한계: 중첩 구조 처리 불가능.

## 문맥 자유 언어와 푸시다운 오토마타 (PDA)

### 문맥 자유 언어 (Context-Free Languages)

- 정의: PDA로 파싱 가능한 언어.
- 사용 예: 프로그래밍 언어의 구문 분석, XML, JSON.

### 푸시다운 오토마타 (PDA, Pushdown Automata)

- 정의: 스택을 사용하는 유한 상태 오토마타.
- 구성 요소: 상태, 입력 알파벳, 스택 알파벳, 전이 함수, 시작 상태, 종료 상태.
- 장점: 중첩 구조 처리 가능.
- 한계: 더 복잡한 문법 처리 불가능.

## 문맥 민감 언어와 선형 유한 비결정적 튜링 머신 (LBA)

### 문맥 민감 언어 (Context-Sensitive Languages)

- 정의: LBA로 파싱 가능한 언어.
- 사용 예: 프로그래밍 언어의 고급 구문.

### 선형 유한 비결정적 튜링 머신 (LBA, Linear Bounded Automaton)

- 정의: 테이프의 길이가 입력의 길이에 비례하는 제한된 튜링 머신.
- 특징: 중첩된 구조 및 문맥 의존적 구조 처리 가능.

## 재귀 하향 언어와 튜링 머신

### 재귀 하향 언어 (Recursively Enumerable Languages)

- 정의: 튜링 머신으로 파싱 가능한 언어.
- 사용 예: 고급 프로그래밍 언어.

### 튜링 머신 (Turing Machine)

- 정의: 무한 테이프와 상태 전이를 사용하여 계산을 수행하는 기계.
- 특징: 모든 계산 가능한 문제를 해결할 수 있으나 무한 루프 가능성 있음.

## 실제 사례

### 정규 표현식의 한계

- 예제: 괄호 짝 맞추기, HTML 태그 분석.

### HTML 파싱의 어려움

- 구체적인 예: HTML의 중첩 구조 파싱.

### JSON 파싱

- 구체적인 예: JSON의 계층적 데이터 구조 파싱.

## 고급 파싱 기법

### Lookahead 파싱

- 정의: 미래의 입력을 미리 읽고 파싱하는 방법.
- 기법: LR(0), LL(1), LR(1), SLR(n), LALR(n).

### Parsing Expression Grammar (PEG)

- 정의: CFG의 한계를 극복한 기법.
- 장점: 비결정성 문제 해결, 우선순위 선택.

### ISGLR

- 정의: 증분적 스캐너 없는 일반화된 LR 파싱.
- 장점: 성능 향상, 구문 트리 재사용.

### Pika Parsing

- 정의: 역방향 동적 프로그래밍을 통한 좌측 재귀와 오류 복구 해결.
- 장점: 오류 복구, 좌측 재귀 처리.

### Earley Parser

- 정의: 모든 CFG를 처리할 수 있는 파서.
- 특징: 효율적인 파싱, 특히 애매모호한 문법 처리에 유용.

### CYK Algorithm

- 정의: 특정 CFG를 위한 효율적인 파싱 알고리즘.
- 특징: 문법을 CNF(Chomsky Normal Form)로 변환하여 사용.

## 결론

- 요약: 각 파싱 방법의 장단점과 적절한 사용 사례.
- 추가 참고 자료: 최신 파싱 이론 논문 및 기술 자료.

## 참고 문헌 및 자료

- [Parsing Theory Overview](https://stereobooster.com/parsing-theory-overview)
- [Top Down Parsing Explained](https://edurev.in/course/lecture)
- [Comprehensive Taxonomy for Log Parsing](https://conf.researchr.org/icse-2024-posters)
- [Parsing Techniques for Logs](https://www.researchgate.net/publication/350284167_Parsing_Techniques_for_Logs)
