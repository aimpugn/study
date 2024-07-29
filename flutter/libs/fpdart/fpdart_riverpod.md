# fpdart_riverpod

- [fpdart\_riverpod](#fpdart_riverpod)
    - [📔 API 정의: `StorageService`](#-api-정의-storageservice)
    - [🔗 `riverpod`과 `fpdart`를 연결](#-riverpod과-fpdart를-연결)
    - [`fpdart`와 `Do` 표기법](#fpdart와-do-표기법)
        - [Business logic with fpdart and the Do notation 참고](#business-logic-with-fpdart-and-the-do-notation-참고)
        - [`monadic bind` 또는 `binding function`](#monadic-bind-또는-binding-function)
            - [`_()`?](#_)

## 📔 API 정의: `StorageService`

`abstract class`을 사용하면:

1. 구체적인 구현 없이 API 메서드를 정의할 수 있다

    ```dart
    /// `ReaderTaskEither`의 주 의존성
    abstract class StorageService {
    /// 👇 Future를 리턴하므로 호출 시 `Either` 아닌 `TaskEither` 사용 필요 
        Future<List<EventEntity>> get getAll; /// Returns the list of all `EventEntity` in storage
        Future<EventEntity> put(String title); ///  Create and store a new `EventEntity` given its title, and return the new instance of `EventEntity`
    }
    ```

2. 유지보수성에 좋다. 특정 데이터 소스에 종속되지 않고, 구현체를 바꿈으로써 저장소를 바꿀 수 있다.

    ```dart
    /// All of the classes below are valid `StorageService`
    ///
    /// You can use any of them to store the data ✅
    class SharedPreferencesStorage implements StorageService { ... }
    class LocalStorage implements StorageService { ... }
    class RemoteDatabaseStorage implements StorageService { ... }
    ```

3. 테스트에 좋다

    ```dart
    // Mock `StorageService` for testing with ease 🤝
    class TestStorageService extends Mock implements StorageService {}
    ```

## 🔗 `riverpod`과 `fpdart`를 연결

```dart
@riverpod
Future<GetAllEventState> eventList(EventListRef ref) async {
  final service = ref.watch(storageServiceProvider);
  return getAllEvent.run(service);
}
```

## `fpdart`와 `Do` 표기법

### [Business logic with fpdart and the Do notation](https://www.sandromaglione.com/techblog/fpdart-riverpod-functional-programming-flutter-part-5) 참고

- ✅ 모든 `fpdart` 요청에 제공하려면 유효한 `StorageService` 구현체에 접근해야 하고, 이를 위해 `StorageService`를 위한 프로바이더 필요
- ✅ `getAllEvent`를 실행하고 UI에 목록을 리턴할 책임이 있는 프로바이더 필요

```dart
/// https://www.sandromaglione.com/techblog/fpdart-riverpod-functional-programming-flutter-part-3#readertaskeither-dependencies--errors--success
///              👇 Dependency   👇 Either 실패(on left)
ReaderTaskEither<StorageService, Errors, Success> getAllEvent = ReaderTaskEither(/* TODO */);
///                                      👆 Either 성공(on right)
```

문제는 `riverpod` 자체적으로 에러를 핸들링하는 방법을 갖고 있다는 것.
`FutureProvider`는 `AsyncValue`를 리턴하고, `AsyncValue`는 기본적으로 로딩과 에러 상태를 제공한다.
이때 `Either`를 사용하는 것은 불편하고 에러 핸들링 하는 코드가 중복된다.

```dart
// duplicate
eventList.map(
  loading: (_) => ...,
  error: (error) => ..., // 👈 Error from `riverpod`'s `AsyncValue`
  data: (either) => either.match(
    (error) => switch(error) { ... }, // 👈 Pattern match error from `fpdart`'s `Either`
    (success) => switch(success) { ... }, // Pattern match success value from `fpdart`'s `Either`
  ),
)

// 우리가 원하는 것
eventList.map(
  loading: (_) => ...,
  error: (error) => ..., // 👈 Unexpected errors
  data: (data) => switch(data) { ... }, // 👈 One `switch` for all expected errors and success value
)
```

- ✅ 예상치 못한 에러는 `Either` 대신 `AsyncValue` 사용해서 처리
- ✅ `ReaderTaskEither`를 사용하는 대신 `Either`가 없는 `ReaderTask` 사용해서
    - 예상되는 에러 처리
    - 성공 값 처리

```dart
/// without `Either`                  👇 모든 가능한 에러와 성공 값을 하나의 sealed class로 인코딩(encode)해야 한다
final ReaderTask<StorageService, GetAllEventState> getAllEvent 
///                                                   `fpdart` 내부의 모든 타입에는 `Do` 표기법(notation) 함수를 초기화할 수 있는 `Do` 생성자가 있다
///                                                👇 `Do` 생성자는 함수에 대한 접근을 제공한다. (그 함수는 관례상 `_`라고 한다).
    = ReaderTask<StorageService, GetAllEventState>.Do(
///                   👆👇                👆👇
///          (storageService) async => GetAllEventState 타입 결과
/// 
/// `ReaderTask`는 여기서 `StorageService`의 인스턴스에 대한 접근을 제공하며,
/// 여기서 `StorageService.getAll`을 호출하고 `List<EventEntity>`를 얻기 위해 사용한다

/// (_) async {
///  👆 에러를 함수형 스타일로 처리하면서, 모든 `fpdart`의 타입으로부터 결과 값을 추출하고 사용할 수 있도록 해준다.
///     `_`은 컨벤션(called `_` by convention).
/// },
///
/// < `_` 함수의 시그니처 >
/// `Future<A> Function<A>(ReaderTask<StorageService, A>) _` 
/// => `Future<GetAllEventState> Function<GetAllEventState>(ReaderTask<StorageService, GetAllEventState>) _`
/// => `ReaderTask<StorageService, GetAllEventState>`를 인자로 받고, `Future<GetAllEventState>` 타입을 리턴한다

        (_) async {
///     👆 storageService
///                                                                                  `Do` 생성자로 do 표기법 함수 초기화
///                                                                                  `Do` 표기법 안에서 `_` 사용하면 `run` 호출하지 않고 리턴 값들을 추출할 수 있다
///                                                                                  `Do` 표기법 내 `_` 함수는 `ReaderTask`을 필요로 한다  
///                                                                                  👇 
            TaskEither<QueryGetAllEventError, List<EventEntity>> executeQuery = await _(
///         👆 `getAll` 함수 호출하기 위해 사용
                ReaderTask(
                    (storageService) async => TaskEither.tryCatch(
///                                                      👆 가능한 에러를 처리하기 위해 `tryCatch` 생성자 사용
                        () => storageService.getAll,
                        QueryGetAllEventError.new, // (object, stackTrace) => QueryGetAllEventError(object, stackTrace), 와 같다
                    ),
                ),
            );
///         `run` should only be called at the very end!
///         이 예제의 경우 결과 값을 사용하기 전 마지막 단계에서 `riverpod` 프로바이더에서 `getAllEvent.run(service);` 실행

            return _(
                ReaderTask(
                    (_) => executeQuery
///                     👇 에러와 성공 값들을 `GetAllEventState`에 매핑하기 위해 호출
                        .match(
                            identity, // 주어진 입력 값을 반환하는 `fpdart`의 함수. `T identity<T>(T a) => a;`
///                         👆 실패 경우. `A Function(QueryGetAllEventError) onLeft`
                            SuccessGetAllEventState.new,
///                         👆 성공 경우. `A Function(List<EventEntity>) onRight`
                        )
///                     👆 `match`의 결과 `Task<GetAllEventState>`가 리턴되며,
                        .run(),
///                     👆 `Task`를 실행하고 `GetAllEventState`를 추출하기 위해 `run` 실행 필요
                ),
            );
        },
);
```

`Do` 표기법 사용하지 않으면서 같은 기능하는 코드

```dart
/// Chain of method calls instead of a series of step
final getAllEventChain = ReaderTask(
  (StorageService storageService) => TaskEither.tryCatch(
    () => storageService.getAll,
    QueryGetAllEventError.new,
  )
      .match(
        identity,
        SuccessGetAllEventState.new,
      )
      .run(),
);
```

모든 가능한 에러와 성공 값을 하나의 sealed class로 인코딩(encode)하기 위해 `GetAllEventState` 정의

```dart
/// get_all_event_state.dart
import 'package:fpdart_riverpod/entities/event_entity.dart';

part 'get_all_event_error.dart';
 
sealed class GetAllEventState {
  const GetAllEventState();
}
 
class SuccessGetAllEventState extends GetAllEventState {
  final List<EventEntity> eventEntity;
  const SuccessGetAllEventState(this.eventEntity);
}
```

코드를 깔끔하게 구조화하기 위해 에러를 별도 파일로 생성한다. 하지만 `sealed class`는 같은 라이브러리에서만 상속이 되므로 `part`, `part of`를 사용한다

```dart
/// get_all_event_error.dart
part of 'get_all_event_state.dart';
 
sealed class GetAllEventError extends GetAllEventState {
  const GetAllEventError();
}
 
class QueryGetAllEventError extends GetAllEventError {
  final Object object;
  final StackTrace stackTrace;
  const QueryGetAllEventError(this.object, this.stackTrace);
}
```

### `monadic bind` 또는 `binding function`

```dart
  TaskEither<ApplicationError, JWTApplicationSuccess> getJWT(
      String email, String password) {
    return TaskEither<ApplicationError, JWTApplicationSuccess>.Do((_) async {
          await getDefaultUserInfo(email, password).run(userRepository);
      if (appUserState is AppUserError) {
        return _(TaskEither.left(const ApplicationError.notFound("")));
      }

      return _(TaskEither.right(JWTApplicationSuccess(JWT({}))));
    });
  }
```

#### `_()`?

- `monadic bind` 또는 `binding function`을 수행하는 함수. 함수형 프로그래밍에서 이러한 함수는 모나드의 값을 "풀어주는" 역할
- `TaskEither`의 값을 `Future`로 변환해주는 함수. 즉, `TaskEither<ApplicationError, A>` 타입의 값을 받아 `Future<A>` 타입의 값을 반환
    - 타입을 보면 `Future<A> Function<A>(TaskEither<AppUserError, A>)` 라고 나온다
    - `TaskEither<AppUserError, A>`를 인자로 받아서 `Future<A>`를 리턴하는 함수
    - `TaskEither`의 내부 값을 추출하여 Future로 변환. `TaskEither` 내부의 실제 값(성공 또는 실패)을 가져와 다음 연산을 수행할 수 있다.

    ```dart
    final result = await _(TaskEither.right(42)); // result는 Future<42>
    ```

> `monadic bind`?

- 모나드의 값을 다른 모나드로 변환하는 연산
- `TaskEither`는 실패 또는 성공의 상태와 함께 값을 가진 비동기 작업을 나타내는 모나드. `TaskEither`의 값을 다루려면 내부의 `Either` 값을 추출해야 하는데, 이 작업을 바인딩 함수가 해준다
