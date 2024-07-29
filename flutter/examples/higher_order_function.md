# higher order function

## 고차 함수

## 예제

### `guardWebExceptions<R>` 함수

이 함수는 주어진 콜백 함수 `cb`를 실행하고, 특정 유형의 예외(FirebaseException)를 처리하며, 다른 예외는 그대로 전달합니다.
함수형 프로그래밍의 한 예로서 예외 처리를 함수로 추상화하여 재사용성과 코드의 간결성을 높이는 것을 목표로 하는 코드입니다.

이런 고차 함수는 다음과 같은 장점이 있습니다:
- **코드 재사용성**: 여러 곳에서 동일한 예외 처리를 반복하는 대신, `guardWebExceptions` 함수를 사용하여 예외 처리 로직을 일관성 있게 유지할 수 있습니다.
- **코드 간결성**: 예외 처리 코드가 함수 하나로 추상화되어, 코드가 간결하고 읽기 쉬워집니다.
- **유연성**: 콜백 함수와 커스텀 파서 함수를 매개변수로 받아 다양한 상황에 유연하게 대응할 수 있습니다.

```dart
R guardWebExceptions<R>(
  R Function() cb, {
  required String plugin,
  required String Function(String) codeParser,
  String Function(String code, String message)? messageParser,
}) {
  try {
    final value = cb();

    if (value is Future) {
      return value.catchError(
        (err, stack) => Error.throwWithStackTrace(
          _mapException(
            err,
            plugin: plugin,
            codeParser: codeParser,
            messageParser: messageParser,
          ),
          stack,
        ),
        test: _testException,
      ) as R;
    } else if (value is Stream) {
      return value.handleError(
        (err, stack) => Error.throwWithStackTrace(
          _mapException(
            err,
            plugin: plugin,
            codeParser: codeParser,
            messageParser: messageParser,
          ),
          stack,
        ),
        test: _testException,
      ) as R;
    }

    return value;
  } catch (error, stack) {
    if (!_testException(error)) {
      // Make sure to preserve the stacktrace
      rethrow;
    }

    Error.throwWithStackTrace(
      _mapException(
        error,
        plugin: plugin,
        codeParser: codeParser,
        messageParser: messageParser,
      ),
      stack,
    );
  }
}
```

1. **제네릭 타입 `R`**:
   - 이 함수는 제네릭 타입 `R`을 사용하여 반환값의 타입을 호출 시점에 결정할 수 있도록 합니다.
   - 이는 함수가 다양한 타입의 반환값을 처리할 수 있게 합니다.

2. **매개변수 `cb`**:
   - `R Function()` 타입의 콜백 함수입니다. 이 함수는 예외 처리가 적용될 코드 블록을 전달받습니다.

3. **필수 매개변수**:
   - `plugin`: 문자열 타입으로, 오류가 발생한 플러그인의 이름을 나타냅니다.
   - `codeParser`: 문자열을 입력으로 받아 오류 코드를 파싱하는 함수입니다.

4. **선택적 매개변수**:
   - `messageParser`: 오류 코드와 메시지를 입력으로 받아 커스텀 메시지를 생성하는 함수입니다.

5. **예외 처리 로직**:
   - `try` 블록에서 콜백 함수 `cb`를 실행합니다.
   - `Future` 타입의 반환값일 경우 `catchError` 메서드를 사용하여 오류를 잡습니다.
   - `Stream` 타입의 반환값일 경우 `handleError` 메서드를 사용하여 오류를 잡습니다.
   - 위의 경우가 아닌 일반 값일 경우 바로 반환합니다.
   - `catch` 블록에서 예외가 발생하면 `_testException` 함수로 예외 유형을 검사합니다.
   - 예외 유형이 처리 대상일 경우 `_mapException` 함수로 예외를 변환하고, 스택 트레이스를 보존한 채 다시 던집니다.

6. **헬퍼 함수**:
   - `_testException`: 예외가 처리 가능한 유형인지 검사합니다.
   - `_mapException`: 웹 오류를 `FirebaseException`으로 변환합니다.

이를 다음과 같이 사용할 수 있습니다.

```dart
void main() {
  // 예제 콜백 함수
  String exampleFunction() {
    // 웹 오류가 발생한다고 가정
    throw WebError('404', 'Not Found');
  }

  // 예외 처리 적용
  var result = guardWebExceptions<String>(
    exampleFunction,
    plugin: 'examplePlugin',
    codeParser: (code) => 'plugin-$code',
    messageParser: (code, message) => 'Error: $message',
  );

  print(result);
}
```

이 예제에서 `exampleFunction`이 웹 오류를 발생시키고, `guardWebExceptions` 함수가 이를 `FirebaseException`으로 변환하여 처리합니다.

---

```dart
/// Will return a [FirebaseException] from a thrown web error.
/// Any other errors will be propagated as normal.
R guardWebExceptions<R>(
  R Function() cb, {
  required String plugin,
  required String Function(String) codeParser,
  String Function(String code, String message)? messageParser,
}) {
  try {
    final value = cb();

    if (value is Future) {
      return value.catchError(
        (err, stack) => Error.throwWithStackTrace(
          _mapException(
            err,
            plugin: plugin,
            codeParser: codeParser,
            messageParser: messageParser,
          ),
          stack,
        ),
        test: _testException,
      ) as R;
    } else if (value is Stream) {
      return value.handleError(
        (err, stack) => Error.throwWithStackTrace(
          _mapException(
            err,
            plugin: plugin,
            codeParser: codeParser,
            messageParser: messageParser,
          ),
          stack,
        ),
        test: _testException,
      ) as R;
    }

    return value;
  } catch (error, stack) {
    if (!_testException(error)) {
      // Make sure to preserve the stacktrace
      rethrow;
    }

    Error.throwWithStackTrace(
      _mapException(
        error,
        plugin: plugin,
        codeParser: codeParser,
        messageParser: messageParser,
      ),
      stack,
    );
  }
}
```
