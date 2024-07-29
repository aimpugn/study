# Either

- [Either](#either)
    - [`foldIdentity` 확장 함수 정의해보기](#foldidentity-확장-함수-정의해보기)
        - [왜 직접 `fold` 사용시와 `extension`에서 `fold` 사용시 다르게 동작?](#왜-직접-fold-사용시와-extension에서-fold-사용시-다르게-동작)

## `foldIdentity` 확장 함수 정의해보기

```dart
extension EitherExtension<L, R> on Either<L, R> {
  Object foldIdentity() {
    return fold(identity, identity);
                ^^^^^^^^  ^^^^^^^^ 에러 발생
                Couldn't infer type parameter 'T'.

                Tried to infer 'L' for 'T' which doesn't work:
                Function type declared as 'T Function<T>(T)'
                                used where  'Object Function(L)' is required.

                Consider passing explicit type argument(s) to the generic.
  }
}
```

근데 `fold(identity, identity)`를 직접 사용하면 `Object`로 추론이 된다

```dart
final signInResult = signInEither.fold(identity, identity);
      ^^^^^^^^^^^^
      Object 타입
```

직접 사용시에는 `fold` 함수가 `L`과 `R` 중 어떤 타입을 반환할지 명확하지 않기 때문에 가장 상위 타입인 `Object`로 추론된다.
하지만, `extension`에서 `fold`를 사용할 때는 타입 추론이 더 엄격하게 적용되어, `L`과 `R`이 명확히 구분되지 않으면 오류가 발생한다.

### 왜 직접 `fold` 사용시와 `extension`에서 `fold` 사용시 다르게 동작?

- Dart에서 `extension` 내부에서 `fold`를 사용할 때와 직접 `fold`를 사용할 때 타입 추론이 다르게 동작하는 것은 언어의 타입 추론 메커니즘과 관련이 있다.
    - 특히, 제네릭 타입과 관련된 타입 추론에서 복잡한 상황이 발생할 수 있다
    - 그러나, 일반적으로 Dart의 타입 시스템은 상황에 따라 일관성 있게 타입을 추론하려고 한다
- `fold(identity, identity)` 사용 시 직접 호출하면 Dart 컴파일러는 `Either<L, R>` 타입에서 `L`과 `R` 중 어떤 타입이 반환될지 확실하지 않기 때문에 가장 상위 타입인 `Object`로 추론
    - 이 경우 컴파일러는 `identity` 함수의 타입 파라미터 `T`를 `Object`로 추론하게 됩니다.
- 반면, `extension` 내에서 `fold`를 사용할 때 타입 추론은 더 엄격해질 수 있다
    - `extension`은 기존 타입에 새로운 메서드를 추가하는 것으로, `extension` 내부에서 `fold`를 호출하면 `L`과 `R`이 분명히 구분되어야 한다.
    - `identity` 함수를 사용할 때 `L`과 `R` 중 어느 쪽 타입을 사용해야 할지 명확하지 않으므로, 타입 추론 오류가 발생

근데 에러를 보면 Dart 컴파일러가 `Either<L, R>`의 `fold` 메서드에서 `L`과 `R`을 각각 구별하기는 한다.

```log
# Left 경우
Couldn't infer type parameter 'T'.

Tried to infer 'L' for 'T' which doesn't work:
  Function type declared as 'T Function<T>(T)'
                used where  'Object Function(L)' is required.

Consider passing explicit type argument(s) to the generic.

# Right 경우
Couldn't infer type parameter 'T'.

Tried to infer 'R' for 'T' which doesn't work:
  Function type declared as 'T Function<T>(T)'
                used where  'Object Function(R)' is required.

Consider passing explicit type argument(s) to the generic.
```

그러나 문제는 `identity` 함수의 타입 추론과 관련이 있다. `identity` 함수는 제네릭 함수로, 타입 매개변수 `T`를 받아 동일한 타입 `T`를 반환한다.

```dart
T identity<T>(T value) => value;
```

`fold(identity, identity)`를 사용할 때, `fold` 메서드의 **각 매개변수**에 대해 `identity` 함수의 타입 `T`를 추론해야 한다.
즉, `fold`의 첫 번째 매개변수(Left 처리)에 대해서는 `T`를 `L`로 추론하려고 하고, 두 번째 매개변수(Right 처리)에 대해서는 `T`를 `R`로 추론하려고 한다.

하지만 `identity` 함수의 타입 `T`는 **함수 호출 시점에서 결정**되어야 한다. `Either<L, R>`의 `fold` 메서드에서 `L`과 `R`이 다른 타입일 경우, 컴파일러는 한 함수(`identity`)에 대해 두 가지 다른 타입을 추론하려고 시도하지만, 이는 타입 시스템의 한계로 인해 실패하게 된다. 결과적으로, 컴파일러는 `identity` 함수의 타입 `T`를 `L` 또는 `R`로 적절하게 추론할 수 없어 오류를 발생시킨다.

이러한 타입 추론 문제를 해결하기 위해서는 `fold` 내에서 각각의 경우에 대해 명시적인 타입 처리를 제공하는 것이 좋다.

```dart
extension EitherExtension<L, R> on Either<L, R> {
  Object foldIdentity() {
    // `L`과 `R` 각각에 대해 명시적으로 타입을 처리하게 되며, 타입 추론 문제를 해결
    return fold((L l) => l as Object, (R r) => r as Object);
  }
}
```
