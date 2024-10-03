# Stream

## 스트리밍

스트리밍은 데이터의 연속적인 흐름을 처리하는 방법으로, *대량의 데이터를 작은 단위로 나누어 순차적으로 처리*합니다.

```plaintext
[Data Source]
     |
  [Stream]
     |
[Intermediate Operations]
     |
[Terminal Operation]
     |
[Result]
```

## 특징

- 데이터 파이프라인: 데이터 소스부터 최종 처리까지의 연속적인 흐름을 구축합니다.
- Lazy Evaluation: 필요할 때까지 계산을 지연하여 메모리 사용을 최소화합니다.
- 함수형 프로그래밍 패턴: `map`, `filter`, `reduce` 등의 함수로 데이터를 변환합니다.
- 효율적인 메모리 사용: 전체 데이터를 메모리에 로드하지 않고도 처리 가능합니다.
- 실시간 데이터 처리: 데이터가 생성됨과 동시에 즉시 처리할 수 있습니다.

## 예제

### Java

```java
List<String> names = Arrays.asList("Alice", "Bob", "Charlie", "David");

List<String> filteredNames = names.stream()
    .filter(name -> name.startsWith("C"))
    .map(String::toUpperCase)
    .collect(Collectors.toList());

System.out.println(filteredNames); // [CHARLIE]
```
