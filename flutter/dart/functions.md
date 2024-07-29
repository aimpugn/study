# Functions

- [Functions](#functions)
    - [Functions](#functions-1)
    - [화살표 함수 (Arrow Function)](#화살표-함수-arrow-function)
    - [중괄호를 사용한 함수 (Block Function)](#중괄호를-사용한-함수-block-function)
    - [기타](#기타)
        - [Arrow Function VS Block Function](#arrow-function-vs-block-function)

## [Functions](https://dart.dev/language/functions)

## 화살표 함수 (Arrow Function)

- `=> expr` 구문은 `{ return expr; }`의 약어
- 화살표(`=>`)와 세미콜론(`;`) 사이에는 문(statement)이 아닌 표현식(expression)만 표시할 수 있다. 예를 들어 if 문을 넣을 수는 없지만 [조건식](https://dart.dev/language/operators#conditional-expressions)은 사용할 수 있습니다.
- 화살표 함수는 `=>` 기호를 사용하여 단일 표현식을 간결하게 반환한다
- 이는 주로 한 줄짜리 간단한 함수에 적합하다

```dart
() => someExpression;
```

예시:

```dart
() => setState(() => _passwordVisible = !_passwordVisible);
```

- 이 경우, 화살표 함수는 `setState`를 호출하고, 그 내부에서 `_passwordVisible`의 값을 반전시킨다
- 이 함수는 단일 표현식을 실행하므로 화살표 함수로 작성하기 적합하다

## 중괄호를 사용한 함수 (Block Function)

- 중괄호를 사용하는 함수는 복수의 문장을 포함할 수 있으며, 보다 복잡한 로직을 수행할 수 있다

```dart
() {
  // 여러 문장 실행 가능
  someStatement;
  anotherStatement;
}
```

예시:

```dart
() {
  setState(() => _passwordVisible = !_passwordVisible);
}
```

- 이 경우, 중괄호 안에 `setState`를 호출하는 코드가 들어간다
- 중괄호를 사용하는 함수는 여러 줄의 코드를 포함할 수 있으므로, 복잡한 함수에 적합하다

## 고차 함수(Higher order function)

## 기타

### Arrow Function VS Block Function

1. **표현식 대 문장:**
    - 화살표 함수는 단일 표현식만을 반환
    - 중괄호를 사용한 함수는 여러 문장을 포함할 수 있음
2. **간결함 대 복잡성:**
    - 화살표 함수는 코드를 더 간결하게 만드는 데 적합하며, 주로 간단한 연산에 사용
    - 반면, 중괄호를 사용한 함수는 더 복잡한 로직을 포함할 수 있다
3. **반환 값:**
    - 화살표 함수는 자동으로 오른쪽의 표현식의 결과를 반환한다
    - 중괄호를 사용한 함수에서는 `return` 문을 명시적으로 사용해야 값을 반환할 수 있다
