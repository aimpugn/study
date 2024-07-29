# ListView

- [ListView](#listview)
    - [ListView](#listview-1)
    - [주요 특징](#주요-특징)
    - [Flexible 위젯들 사용 제한](#flexible-위젯들-사용-제한)

## ListView

- `ListView`는 일련의 자식 위젯들을 세로로 나열하는 스크롤 가능한 리스트를 만드는 데 사용된다

## 주요 특징

- 자체 스크롤 가능성: `ListView`는 기본적으로 스크롤 가능하며, 별도의 `SingleChildScrollView`가 필요하지 않다.
- 대용량 데이터 처리: `ListView.builder` 생성자를 사용하면 대량의 데이터를 효율적으로 처리할 수 있다. 이는 화면에 보이는 부분만 렌더링하여 메모리 사용을 최적화한다
- 높이 자동 조절: `ListView`는 부모 위젯으로부터 받은 높이 제약에 따라 자동으로 크기를 조절한다

## Flexible 위젯들 사용 제한

- `ListView`나 `SingleChildScrollView` 내에서 `Flexible` 또는 `Expanded`와 같은 위젯을 사용하는 것은 피해야 한다
- 이러한 위젯들은 부모 위젯의 제한된 크기에 맞춰 자식들의 크기를 조절하기 때문에, **무한 스크롤 영역에서 예상치 못한 문제를 일으킬 수** 있다
