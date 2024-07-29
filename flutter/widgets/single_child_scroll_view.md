# `SingleChildScrollView`

- [`SingleChildScrollView`](#singlechildscrollview)
    - [`SingleChildScrollView`](#singlechildscrollview-1)
    - [주요 특징](#주요-특징)
    - [Flexible 위젯들 사용 제한](#flexible-위젯들-사용-제한)

## `SingleChildScrollView`

- `SingleChildScrollView`는 단일 자식 위젯을 감싸 스크롤 가능하게 만드는 위젯

## 주요 특징

- 단일 자식 위젯: `SingleChildScrollView`는 하나의 자식 위젯만을 갖으며, 이 자식은 일반적으로 `Column`, `Row`, `Container` 등이 될 수 있다
- 무한 높이/너비: 자식 위젯에 무한한 높이 또는 너비를 제공하여, 내용이 화면 크기를 초과할 경우 스크롤이 가능하도록 한다
- 자식 크기에 따라 크기 결정: 자식 위젯의 크기에 따라 스크롤 영역의 크기가 결정된다

## Flexible 위젯들 사용 제한

- `ListView`나 `SingleChildScrollView` 내에서 `Flexible` 또는 `Expanded`와 같은 위젯을 사용하는 것은 피해야 한다
- 이러한 위젯들은 부모 위젯의 제한된 크기에 맞춰 자식들의 크기를 조절하기 때문에, **무한 스크롤 영역에서 예상치 못한 문제를 일으킬 수** 있다
