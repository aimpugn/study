# fpdart

- [fpdart](#fpdart)
    - [`Either` vs `try ~ catch`](#either-vs-try--catch)
    - [`TaskEither`](#taskeither)
        - [`TaskEither`와 `run`](#taskeither와-run)
        - [`TaskEither.tryCatch`](#taskeithertrycatch)
        - [`TaskEither.tryCatchK`](#taskeithertrycatchk)
    - [ReaderTaskEither](#readertaskeither)
    - [`ReaderTask`](#readertask)
    - [`ReaderTask`와 `TaskEither` 그리고 레이어 선택](#readertask와-taskeither-그리고-레이어-선택)

## `Either` vs `try ~ catch`

- 에러를 명확히 할 수 있다
- `try ~ catch`는 구조가 깊어지면 어디서 이 catch가 이뤄지는지 파악하기가 어렵다
    - 조무보 - 부모 - 자식에서 여러 함수가 있다고 할 때, 조부모에서 catch 하는 경우 있고 부모에서 catch 하는 경우 있고 이렇게 된다면 try ~ catch의 범위가 명확치 않아서 나중에 에러 처리가 힘들다

## `TaskEither`

- 실패 또는 성공을 나타내는 `Either`를 반환하는 비동기 작업을 추상화
- `Monad`와 `Functor`의 특성을 가지고 있다
- 두 개의 모나드로 구성
    - `Task`: 비동기 작업 의미
    - `Either`: 두 가지 경우 의미. Left는 에러, Right는 정상
- Functor 특성
    - `map`, `flatMap` 등의 함수를 사용 가능
    - 비동기 작업을 연쇄적으로 수행할 때 유용

```dart
// infra: users_repository_impl
  @override
  Future<Either<UserError, AppUser>> signInByEmailAndPassword(
      String email, String password) async {
    final response = await _auth.signInWithEmailAndPassword(
      email: email,
      password: password,
    );

    final user = response.user;
    if (user == null) {
      return left(const UserError.notFound());
    }

    return right(user.toAppUser);
  }
```

```dart
// application: users_service
  TaskEither<UserError, AppUser> signInByEmailAndPassword(
          String email, String password) =>
      TaskEither(
          () => usersRepository.signInByEmailAndPassword(email, password));
```

```dart
// presentation: sign_up_controller
    // TaskEither<UserError, AppUser>
    final taskEither = userService.signInByEmailAndPassword(
        signUpInfoDTO.email ?? '', signUpInfoDTO.password ?? '');
    // Future<Either<UserError, AppUser>>
    final result1 = taskEither.run();
    // Either<UserError, AppUser>
    final result2 = await taskEither.run();
```

### `TaskEither`와 `run`

```dart
final class TaskEither<L, R> extends HKT2<_TaskEitherHKT, L, R>
    with
        ... 생략 ...
    final Future<Either<L, R>> Function() _run;

    /// Run the task and return a `Future<Either<L, R>>`.
    Future<Either<L, R>> run() => _run();
```

```dart
state = await AsyncValue.guard(() async {
  final either = await userApplication.getJWT(email, password).run();

  return either.fold(
    (l) => SignInError.of(
      l.message,
      error: l.error,
      stackTrace: l.stackTrace,
    ),
    (r) => SignInSuccess(jwt: r.jwt),
  );
});
```

1. `getJWT`는 `TaskEither<ApplicationError, JWTApplicationSuccess>`를 리턴
    - 여기서 `TaskEither`는 실패(`ApplicationError`) 또는 성공(`JWTApplicationSuccess`)을 나타내는 `Either`를 반환하는 비동기 작업을 추상화
2. 리턴된 `TaskEither`에 `run`을 실행하면 이 비동기 작업을 실행하여 `Future<Either<ApplicationError, JWTApplicationSuccess>>`를 리턴
    - 즉, `run`은 `Task`라는 비동기 작업을 실행해서 `Future`로 바꾼다. `Future`는 Dart에서 비동기 작업을 처리하기 위한 기본적인 방법
    - `await`로 `Future`에서 `Either` 결과를 추출. `Either`는 결과를 `Left`나 `Right`로 명시적으로 나타내어 에러 처리를 간편하게 한다.

즉, `run`은 `Task` 모나드를 `Future`로 만들고, `Future`로부터 추출할 값을 `Either`로 바꿔준다.

### `TaskEither.tryCatch`

### `TaskEither.tryCatchK`

## ReaderTaskEither

아래와 같은 성질을 갖는 함수를 인코딩한다(`ReaderTaskEither` encodes a function that:)

1. (Reader) 의존성을 필요로 하고(Requires some dependencies)
2. (Either) 어떤 에러와 함께 실패할 수 있고(May fail with some error)
3. (Task) 비동기 성공 값을 리터하는 (Returns an async success value)

## `ReaderTask`

- 함수형 프로그래밍에서 사용되는 개념 중 하나. 다음 두 가지 작업을 함께 처리할 수 있는 강력한 도구. 여러 함수나 비동기 작업을 조합하여 -> 새로운 함수나 비동기 작업을 생성하는 데 사용
    - **의존성 주입**(Reader)
    - **비동기 작업**(Task)
- `Reader Monad`와 `Task Monad`의 조합
    - `Reader Monad`: 의존성 주입을 위한 패턴으로, 어떤 컨텍스트 `E`를 받아서 결과 `A`를 생성하는 **함수를 캡슐화**
    - `Task Monad`: 비동기 작업을 표현하는 패턴으로, `Future`나 `Promise`와 유사

## `ReaderTask`와 `TaskEither` 그리고 레이어 선택

- Domain Layer:
    - `ReaderTask`를 사용하는 것이 좋을 거 같다.
    - 도메인 레이어에서는 리포지토리 같은 의존성과 비동기 연산을 다루기 때문. `ReaderTask`를 사용하면 이 두 가지를 효과적으로 처리할 수 있다
    - `*RepositoryImpl` 같은 구현체를 애플리케이션에서 주입 받아 사용할 수 있다.
- Application Layer
    - `TaskEither`를 사용하는 것이 좋을 수 있다
    - 애플리케이션 레이어에서는 주로 비즈니스 로직의 성공 또는 실패를 다룸
    - TaskEither를 사용하면 비동기 연산과 실패 처리를 효과적으로 할 수 있음
