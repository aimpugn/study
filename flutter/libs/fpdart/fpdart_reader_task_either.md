# fpdart ReaderTaskEither

## ReaderTaskEither?

- `Reader`
    - 의존성 주입을 함수형 스타일로 처리한다
    - 특정 타입을 받아 다른 타입을 반환하는 함수를 담고 있다
- `Task`
    - 비동기 작업을 나타낸다
    - 이 작업은 **실행 시점에 평가**된다
- `Either<L, R>`
    - 두 가지 타입 중 하나를 가질 수 있는 컨테이너
    - Left는 실패, Right는 성공

## 구현

### 같은 기능, 여러 바리에이션

```dart
ReaderTask<AuthRepositoryInterface, void> createUserByEmailAndPassword(
        CreateUserByEmailCommand cmd) =>
    ReaderTask.Do((_) async {
    state = const AsyncValue.loading();
    final taskEither = await _(ReaderTask(
        (authRepo) async => authRepo.createUserWithEmailAndPassword(cmd),
    ));

    final either = await taskEither.run();

    state = either.fold(
        (left) => AsyncValue.error(AuthDomainFailure(left), left.stackTrace),
        (right) => AsyncValue.data(AuthDomainSuccess(right)),
    );
});
```

#### `taskEither`를 실행하고 결과를 리턴

```dart
ReaderTask<AuthRepositoryInterface, AuthDomainState>
    createUserByEmailAndPassword2(CreateUserByEmailCommand cmd) =>
        ReaderTask.Do((_) async {
        final taskEither = await _(ReaderTask(
            (authRepo) async => authRepo.createUserWithEmailAndPassword(cmd),
        ));

        final either = await taskEither.run();

        return either.fold(
            (left) => AuthDomainFailure(left),
            (right) => AuthDomainSuccess(right),
        );
    });

```

#### `Do` 표기법 사용하지 않고 `ReaderTaskEither` 리턴

- `ReaderTaskEither` 함수가 실행될 때 `TaskEither`도 즉시 실행
- 명령형 스타일에 가깝다

```dart
ReaderTaskEither<AuthRepositoryInterface, AuthDomainFailure,
        AuthDomainSuccess>
    createUserByEmailAndPassword3(CreateUserByEmailCommand cmd) {
    return ReaderTaskEither((authRepo) async {
        final either =
            await authRepo.createUserWithEmailAndPassword(cmd).run();

        return either.fold(
            (left) => Left(AuthDomainFailure(left)),
            (right) => Right(AuthDomainSuccess(right)),
        );
    });
}
```

#### `taskEither` 생성만 하고 `ReaderTaskEither` 안에서 실행

- `TaskEither`의 실행이 늦춰진다
- 함수형 프로그래밍의 `지연 평가(lazy evaluation)`에 가깝다

```dart
ReaderTaskEither<AuthRepositoryInterface, AuthDomainFailure,
    AuthDomainSuccess> createUserByEmailAndPassword4(
        CreateUserByEmailCommand cmd) =>
    ReaderTaskEither.Do((_) async {
    final taskEither = await _(ReaderTaskEither.fromReader(Reader(
        (authRespo) => authRespo.createUserWithEmailAndPassword(cmd),
    )));

    return _(ReaderTaskEither(
        (_) async => taskEither
            .map(AuthDomainSuccess.new)
            .mapLeft(AuthDomainFailure.new)
            .run(),
    ));
});
```

#### `taskEither.run` 후 결과를 `Either.fold`

```dart
ReaderTaskEither<AuthRepositoryInterface, AuthDomainFailure,
    AuthDomainSuccess> createUserByEmailAndPassword5(
        CreateUserByEmailCommand cmd) =>
    ReaderTaskEither.Do((_) async {
    final taskEither = await _(ReaderTaskEither.fromReader(Reader(
        (authRespo) => authRespo.createUserWithEmailAndPassword(cmd),
    )));

    return _(ReaderTaskEither(
        (authRespo) => taskEither.run().then((value) => value.fold(
            (left) => Either.left(AuthDomainFailure(left)),
            (right) => Either.right(AuthDomainSuccess(right)),
        )),
    ));
});
```
