# sequence

- [sequence](#sequence)
    - [시퀀스 (Sequence)](#시퀀스-sequence)
    - [시퀀스의 일반적 특성](#시퀀스의-일반적-특성)
    - [컴퓨터 과학에서의 시퀀스](#컴퓨터-과학에서의-시퀀스)

## 시퀀스 (Sequence)

시퀀스(Sequence)는 "순서가 있는 일련의 항목들"을 의미하는 일반적인 개념입니다.
[Oxford English Dictionary](https://www.oed.com/view/Entry/176259)에 따르면,
시퀀스의 기본 정의는 "연속적으로 따라오는 것들, 연속체, 연속"입니다.

## 시퀀스의 일반적 특성

1. 순서(Order): 항목들이 특정한 순서로 배열되어 있습니다.
2. 연속성(Continuity): 항목들이 연속적으로 이어집니다.
3. 관계성(Relation): 각 항목은 이전 항목과 다음 항목과 관계를 가집니다.

예를 들어, 숫자의 시퀀스 (1, 2, 3, 4, 5), 알파벳의 시퀀스 (A, B, C, D, E) 등이 있습니다.

## 컴퓨터 과학에서의 시퀀스

프로그래밍에서 "시퀀스"라는 용어는 *순서가 있는 데이터의 집합*을 표현하는 데 사용됩니다.
URL 인코딩/디코딩 맥락에서 이는 문자나 바이트의 순서 있는 나열을 의미합니다.

1. 데이터 구조: 배열(array), 연결 리스트(linked list), 스택(stack), 큐(queue) 등은 시퀀스의 예입니다.
2. 알고리즘: 정렬, 검색 등의 알고리즘은 시퀀스를 다룹니다.
3. 프로그래밍 패러다임: 함수형 프로그래밍에서 시퀀스 처리는 핵심 개념입니다.

- [IEEE Computer Society 웹사이트](https://www.computer.org/education/bodies-of-knowledge/software-engineering) 문서
- [Data Structures: A Comprehensive Introduction](https://dev.to/m__mdy__m/data-structures-a-comprehensive-introduction-2o13)

"바이트 시퀀스"와 "퍼센트 인코딩된 시퀀스"라는 표현에서 시퀀스는 다음을 의미합니다:

1. 바이트 시퀀스:

    연속된 바이트들의 나열을 의미합니다.
    각 바이트는 8비트의 데이터를 나타냅니다.

    예를 들어, ASCII 문자열 "Hello"는 바이트 시퀀스 [72, 101, 108, 108, 111]로 표현됩니다.

2. 퍼센트 인코딩된 시퀀스:

    URL 인코딩된 문자들의 연속을 의미합니다.
    각 인코딩된 문자는 `%`로 시작하고 두 개의 16진수가 따라옵니다.

    예를 들어, "Hello, World!"의 퍼센트 인코딩된 시퀀스는 "Hello%2C%20World%21"입니다.
