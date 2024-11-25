# troubleshootings

- [troubleshootings](#troubleshootings)
    - [Unable to generate package graph, no `/path/to/.dart_tool/flutter_gen/pubspec.yaml` found](#unable-to-generate-package-graph-no-pathtodart_toolflutter_genpubspecyaml-found)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [type 'Null' is not a subtype of type 'YamlMap' in type cast](#type-null-is-not-a-subtype-of-type-yamlmap-in-type-cast)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [The class 'SomeThing' can't be extended, implemented, or mixed in outside of its library because it's a sealed class](#the-class-something-cant-be-extended-implemented-or-mixed-in-outside-of-its-library-because-its-a-sealed-class)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [Error: Bad state: Future already completed](#error-bad-state-future-already-completed)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [상위 클래스로 상태 정의하고 하위 클래스를 값으로 넣을 때, 나중에 하위 클래스로 인지하지 못함](#상위-클래스로-상태-정의하고-하위-클래스를-값으로-넣을-때-나중에-하위-클래스로-인지하지-못함)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [Flutter - 'initialValue == null || controller == null': is not true. error](#flutter---initialvalue--null--controller--null-is-not-true-error)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [state가 업데이트 되지 않는 이슈](#state가-업데이트-되지-않는-이슈)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - [`read` 사용해야 할 곳에서 `watch` 사용 시 에러 발생한다](#read-사용해야-할-곳에서-watch-사용-시-에러-발생한다)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-7)
    - [firebase firestore 저장 에러](#firebase-firestore-저장-에러)
        - [문제](#문제-8)
        - [원인](#원인-8)
        - [해결](#해결-8)
    - [RenderFlex children have non-zero flex but incoming height constraints are unbounded](#renderflex-children-have-non-zero-flex-but-incoming-height-constraints-are-unbounded)
        - [문제](#문제-9)
        - [원인](#원인-9)
        - [플러터 레이아웃 이해](#플러터-레이아웃-이해)
        - [해결](#해결-9)
    - [Vertical viewport was given unbounded height](#vertical-viewport-was-given-unbounded-height)
        - [문제](#문제-10)
        - [원인](#원인-10)
        - [해결](#해결-10)
    - [`TextFormField`의 `validator`가 두 번 호출됨](#textformfield의-validator가-두-번-호출됨)
        - [문제](#문제-11)
        - [원인](#원인-11)
            - [Key 매개변수 추가해보기](#key-매개변수-추가해보기)
            - [왜 내부 `TextFormField`에 별도의 `key`를 할당하지 않아도 되는가?](#왜-내부-textformfield에-별도의-key를-할당하지-않아도-되는가)
        - [해결](#해결-11)
    - [The return type 'Null' isn't a 'AppUser', as required by the closure's context](#the-return-type-null-isnt-a-appuser-as-required-by-the-closures-context)
        - [문제](#문제-12)
        - [원인](#원인-12)
            - [`_` 함수](#_-함수)
        - [해결](#해결-12)
    - [Error: MissingPluginException(No implementation found for method getAll on channel plugins.flutter.io/shared\_preferences)](#error-missingpluginexceptionno-implementation-found-for-method-getall-on-channel-pluginsflutterioshared_preferences)
        - [문제](#문제-13)
        - [원인](#원인-13)
        - [해결](#해결-13)
    - [로그가 두 번 찍히는 문제](#로그가-두-번-찍히는-문제)
        - [문제](#문제-14)
        - [원인](#원인-14)
            - [상태 변경에 의한 재빌드?](#상태-변경에-의한-재빌드)
            - [Flutter 프레임워크의 최적화 동작?](#flutter-프레임워크의-최적화-동작)
            - [외부 요인에 의한 재렌더링?](#외부-요인에-의한-재렌더링)
            - [GoRouter의 페이지 빌더 동작?](#gorouter의-페이지-빌더-동작)
        - [해결](#해결-14)
    - [The following ImageCodecException was thrown resolving an image codec](#the-following-imagecodecexception-was-thrown-resolving-an-image-codec)
        - [문제](#문제-15)
        - [원인](#원인-15)
        - [해결](#해결-15)
    - [Plugin \[id: 'com.android.application', version: '7.2.0', apply: false\] was not found in any of the following sources](#plugin-id-comandroidapplication-version-720-apply-false-was-not-found-in-any-of-the-following-sources)
        - [문제](#문제-16)
        - [원인](#원인-16)
        - [해결](#해결-16)
    - [Dependency 'androidx.activity:activity:1.8.0' requires libraries and applications that depend on it to compile against version 34 or later of the Android APIs](#dependency-androidxactivityactivity180-requires-libraries-and-applications-that-depend-on-it-to-compile-against-version-34-or-later-of-the-android-apis)
        - [문제](#문제-17)
        - [원인](#원인-17)
        - [해결](#해결-17)
    - [Querying the mapped value of provider(java.util.Set) before task ':app:processDebugGoogleServices' has completed is not supported](#querying-the-mapped-value-of-providerjavautilset-before-task-appprocessdebuggoogleservices-has-completed-is-not-supported)
        - [문제](#문제-18)
        - [원인](#원인-18)
        - [해결](#해결-18)
    - [The number of method references in a `.dex` file cannot exceed 64K](#the-number-of-method-references-in-a-dex-file-cannot-exceed-64k)
        - [문제](#문제-19)
        - [원인](#원인-19)
        - [해결](#해결-19)
    - [Unhandled Exception: Binding has not yet been initialized](#unhandled-exception-binding-has-not-yet-been-initialized)
        - [문제](#문제-20)
        - [원인](#원인-20)
        - [해결](#해결-20)
    - [PlatformException(sign\_in\_failed, com.google.android.gms.common.api.ApiException: 10: , null, null)](#platformexceptionsign_in_failed-comgoogleandroidgmscommonapiapiexception-10--null-null)
        - [문제](#문제-21)
        - [원인](#원인-21)
        - [해결](#해결-21)

## Unable to generate package graph, no `/path/to/.dart_tool/flutter_gen/pubspec.yaml` found

### 문제

```log
❯ flutter pub run build_runner watch
Deprecated. Use `dart run` instead.
Building package executable... (3.4s)
Built build_runner:build_runner.
Unhandled exception:
Bad state: Unable to generate package graph, no `/Users/rody/VscodeProjects/dilectio/dilectio_app/.dart_tool/flutter_gen/pubspec.yaml` found.
#0      _pubspecForPath (package:build_runner_core/src/package_graph/package_graph.dart:253:5)
#1      _parsePackageDependencies (package:build_runner_core/src/package_graph/package_graph.dart:227:21)
#2      PackageGraph.forPath (package:build_runner_core/src/package_graph/package_graph.dart:105:33)
<asynchronous suspension>
#3      main (file:///Users/rody/.pub-cache/hosted/pub.dev/build_runner-2.4.6/bin/build_runner.dart:27:30)
<asynchronous suspension>
```

### 원인

### 해결

- [github issue](https://github.com/dart-lang/build/issues/2835#issuecomment-697931824)

```shell
flutter clean

flutter packages pub get
```

## type 'Null' is not a subtype of type 'YamlMap' in type cast

### 문제

```log
❯ dart run build_runner watch
Building package executable... (3.6s)
Built build_runner:build_runner.
Unhandled exception:
type 'Null' is not a subtype of type 'YamlMap' in type cast
#0      _pubspecForPath (package:build_runner_core/src/package_graph/package_graph.dart:256:47)
#1      _parsePackageDependencies (package:build_runner_core/src/package_graph/package_graph.dart:227:21)
#2      PackageGraph.forPath (package:build_runner_core/src/package_graph/package_graph.dart:105:33)
<asynchronous suspension>
#3      main (file:///Users/rody/.pub-cache/hosted/pub.dev/build_runner-2.4.6/bin/build_runner.dart:27:30)
<asynchronous suspension>
```

### 원인

### 해결

지우고 다시 실행하면 된다

```shell
flutter clean

dart run build_runner watch
```

## The class 'SomeThing' can't be extended, implemented, or mixed in outside of its library because it's a sealed class

### 문제

```log
❯ dart run build_runner watch
Resolving dependencies in /Users/rody/VscodeProjects/flutter_ddd_riverpod_example...
Got dependencies in /Users/rody/VscodeProjects/flutter_ddd_riverpod_example.
Building package executable... (3.5s)
Failed to build build_runner:build_runner:
../../.pub-cache/hosted/pub.dev/watcher-1.0.1/lib/src/constructable_file_system_event.dart:7:57: Error: The class 'FileSystemEvent' can't be extended, implemented, or mixed in outside of its library because it's a sealed class.
abstract class _ConstructableFileSystemEvent implements FileSystemEvent {
                                                        ^
../../.pub-cache/hosted/pub.dev/pub_semver-2.1.1/lib/src/version_constraint.dart:96:13: Error: Method not found: 'FallThroughError'.
      throw FallThroughError();
            ^^^^^^^^^^^^^^^^
Failed to build build_runner:build_runner:
../../.pub-cache/hosted/pub.dev/watcher-1.0.1/lib/src/constructable_file_system_event.dart:7:57: Error: The class 'FileSystemEvent' can't be extended, implemented, or mixed in outside of its library because it's a sealed class.
abstract class _ConstructableFileSystemEvent implements FileSystemEvent {
                                                        ^
../../.pub-cache/hosted/pub.dev/pub_semver-2.1.1/lib/src/version_constraint.dart:96:13: Error: Method not found: 'FallThroughError'.
      throw FallThroughError();
```

### 원인

### 해결

- [Watcher fails with class 'FileSystemEvent' can't be extended, for dart, not for flutter.](https://github.com/dart-lang/sdk/issues/52570#issuecomment-1570610278)

```shell
dart pub upgrade
```

## Error: Bad state: Future already completed

### 문제

```log
Error: Bad state: Future already completed
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49     throw_
dart-sdk/lib/async/future_impl.dart 41:79                                        complete
packages/riverpod/src/async_notifier/base.dart 273:16                            onData
packages/riverpod/src/common.dart 324:16                                         map
packages/riverpod/src/async_notifier/base.dart 206:13                            set state
packages/riverpod/src/async_notifier.dart 56:14                                  set state
packages/dilectio_app/src/subprojects/user/domain/auth_domain_service.dart 45:9  <fn>
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 45:50               <fn>
dart-sdk/lib/async/zone.dart 1661:54                                             runUnary
dart-sdk/lib/async/future_impl.dart 156:18                                       handleValue
dart-sdk/lib/async/future_impl.dart 840:44                                       handleValueCallback
dart-sdk/lib/async/future_impl.dart 869:13                                       _propagateToListeners
dart-sdk/lib/async/future_impl.dart 641:5                                        [_completeWithValue]
dart-sdk/lib/async/future_impl.dart 715:7                                        callback
dart-sdk/lib/async/schedule_microtask.dart 40:11                                 _microtaskLoop
dart-sdk/lib/async/schedule_microtask.dart 49:5                                  _startMicrotaskLoop
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 181:15              <fn>
```

### 원인

```dart
part 'auth_domain_service.g.dart';
part 'auth_domain_service.freezed.dart';

@riverpod
class AuthDomainService extends _$AuthDomainService {
  final log = Logger('AuthDomainService');

  @override
  FutureOr<AuthDomainState> build() {
    return AuthDomainSuccess(loggedIn: false);
  }

  ReaderTask<AuthRepositoryInterface, void> loginByEmailAndPassword(
          SignInCommand cmd) =>
      ReaderTask<AuthRepositoryInterface, void>.Do((_) async {
        state = const AsyncValue.loading();

        final taskEither = await _(ReaderTask(
          (userRepo) async =>
              userRepo.signInByEmailAndPassword(cmd.email, cmd.password),
        ));

        final either = await taskEither.run();

        final signInState = either.fold(
          (l) => AuthDomainFailure.internalServerError(
            error: l.error,
            stackTrace: l.stackTrace,
          ),
          (r) => AuthDomainSuccess(
            loggedIn: true,
            appUser: r,
          ),
        );

        state = AsyncValue.data(signInState);
      });
}
```

1. `Future`와 `FutureOr`:
    - `Future<T>`: Dart에서 비동기 연산의 결과를 나타내는 타입입니다. 이는 미래에 어떤 값을 생성할 것이라는 약속을 나타냅니다.
    - `FutureOr<T>`: `T` 혹은 `Future<T>` 타입의 값을 가질 수 있는 Dart의 특수 타입입니다. 이는 동기 및 비동기 코드를 더 유연하게 작성할 수 있게 돕습니다.
2. 오류의 원인:
    - "Future already completed" 오류는 일반적으로 `Completer` 객체에 대해 `complete` 메서드를 두 번 호출하거나,
    - 이미 완료된 `Future` 객체에 대해 `complete` 메서드를 호출하는 경우에 발생합니다.
3. build 메서드와 StateNotifier의 state 프로퍼티:
    - 원본 build 메서드의 시그니처는 `FutureOr<AuthDomainState> build()`로, 이는 `Future<AuthDomainState>` 또는 `AuthDomainState`를 반환할 수 있습니다.
    - `StateNotifier`의 `state` 프로퍼티는 `AsyncValue<AuthDomainState>` 타입으로, 이는 비동기 작업의 현재 상태를 나타냅니다.
4. 문제 발생 상황:
    - `build` 메서드가 `AuthDomainState`를 동기적으로 반환하고, 이 후에 `state` 프로퍼티를 설정할 때, 이미 완료된 `Future` 객체에 연결될 수 있어서 오류가 발생할 수 있습니다.

```dart
FutureOr<AuthDomainState> build() {
  // 동기적으로 AuthDomainState 객체를 반환
  return AuthDomainSuccess(loggedIn: false);
}

void someMethod() {
  // state 프로퍼티를 설정
  state = AsyncValue.data(build());
}
```

- 여기서 build 메서드가 동기적으로 `AuthDomainState` 객체를 반환하므로, 이 `AuthDomainState` 객체는 이미 완료된 `Future` 객체에 연결될 수 있다
- `state` 프로퍼티는 `AsyncValue<AuthDomainState>` 타입이므로, `AsyncValue.data` 메서드를 사용하여 `AuthDomainState` 객체를 `AsyncValue<AuthDomainState>` 객체로 변환
- 그러나 이 변환 과정에서 Future already completed 오류가 발생할 수 있다

### 해결

- `FutureOr<AuthDomainState> build()` 가 아닌 `Future<AuthDomainState> build() async`가 되어야 한다

## 상위 클래스로 상태 정의하고 하위 클래스를 값으로 넣을 때, 나중에 하위 클래스로 인지하지 못함

### 문제

### 원인

```log
AsyncData<AuthDomainState> (AsyncData<AuthDomainState>(value: Instance of 'AuthDomainSuccess'))
    error: null
    isLoading: false,
    stackTrace: null,
    value: AuthDomainSuccess
        appUser: _$36_AppUser (AppUser(defaultInfo: Instance of 'DefaultInfo', educations: null, workplaces: null))
        loggedin: true
```

근데, 아래처럼 수정하면 된다. 왜?

```dart
part 'auth_domain_service.g.dart';
part 'auth_domain_service.freezed.dart';

@Riverpod(keepAlive: true)
class AuthDomainService extends _$AuthDomainService {
  final _log = Logger();

  @override
  Future<AuthDomainState> build() async {
    return AuthDomainState();
  }

  ReaderTask<AuthRepositoryInterface, void> createUserByEmailAndPassword(
          CreateUserByEmailCommand cmd) =>
      authHandler(
        (userRepo) async => userRepo.createUserWithEmailAndPassword(cmd),
      );

  ReaderTask<AuthRepositoryInterface, void> loginByEmailAndPassword(
          SignInByEmailCommand cmd) =>
      authHandler(
        (userRepo) async => userRepo.signInByEmailAndPassword(cmd),
      );

  ReaderTask<AuthRepositoryInterface, void> authHandler(
          Future<TaskEither<AuthDomainFailure, AppUser>> Function(
                  AuthRepositoryInterface)
              authFunction) =>
      ReaderTask<AuthRepositoryInterface, void>.Do((_) async {
        state = const AsyncValue.loading();

        final taskEither = await _(ReaderTask(authFunction));

        final either = await taskEither.run();

        final result = either.match(
          (l) => AuthDomainFailure.internalServerError(
            error: l.error,
            stackTrace: l.stackTrace,
          ),
          (r) => AuthDomainSuccess(
            appUser: r,
          ),
        );
        _log.d(
            "result's type is AuthDomainSuccess? ${result is AuthDomainSuccess}");

        if (result is AuthDomainSuccess) {
          _log.d("[AuthDomainSuccess]");
          if (state.value?.changed != null) {
            _log.d("[AuthDomainSuccess] state.value?.changed != null");
            state = AsyncData(AuthDomainState(changed: true));
          }
          state = AsyncData(result);
          _log.d(
              "[AuthDomainSuccess] state.value.runtimeType is ${state.value.runtimeType}");
          _log.d(
              "[AuthDomainSuccess] state.value is AuthDomainSuccess? ${state.value is AuthDomainSuccess}");
          _log.d("[AuthDomainSuccess] state.value is ${state.value?.changed}");
        }
        if (result is AuthDomainFailure) {
          _log.d("[AuthDomainFailure]");
          state = AsyncError(result, result.stackTrace);
          _log.d(
              "[AuthDomainFailure] state.value.runtimeType is ${state.value.runtimeType}");
        }
      });
}

class AuthDomainState {
  final bool? changed;
  AuthDomainState({this.changed});
}

class AuthDomainSuccess extends AuthDomainState {
  final AppUser appUser;

  AuthDomainSuccess({
    required this.appUser,
  });
}

@freezed
class AuthDomainFailure extends AuthDomainState
    with _$AuthDomainFailure
    implements Exception {
  AuthDomainFailure._();
  @override
  // TODO: implement changed
  bool? get changed => super.changed;

  factory AuthDomainFailure(String message,
      {Object? error,
      required StackTrace stackTrace,
      bool? changed}) = _AuthDomainFailure;

  factory AuthDomainFailure.internalServerError(
          {Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Internal Server Error",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.notFound({Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Not Found",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.conflict({Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Conflict",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.fromFirebase(String? message,
          {Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        message ?? "Error Form Firebase",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );
}
```

상위 클래스의 생성자만 만들어도 된다. 생성자를 지우면?

```dart
class AuthDomainState {
  AuthDomainState();
}
```

### 해결

근데 타입 추론 같은 문제가 아니라, freezed나 riverpod_generator 통해서 만들어진 코드의 문제였던 걸로 보인다. 특히 `freezed`.

`flutter clean` 후에 새로 생성 했을 때도 문제가 재발했었는데, 다음과 같은 순서로 State 클래스의 코드에 `changed` 속성 추가하고 `AuthDomainFailure` 클래스도 수정하다 보니 문제가 해결됐다.
`const factory AuthDomainFailure` -> `factory AuthDomainFailure`로 수정해보기도 했다. 그래서 아마 `freezed` 통해서 생성된 코드에 문제가 있던 게 아닐까...

아니면 나중에는 vscode를 재시작해보는 것도 좋을지도.

```dart
class AuthDomainState {
  final bool? changed;
  AuthDomainState({this.changed});
}

class AuthDomainSuccess extends AuthDomainState {
  final AppUser appUser;

  AuthDomainSuccess({
    required this.appUser,
  });
}

@freezed
class AuthDomainFailure extends AuthDomainState
    with _$AuthDomainFailure
    implements Exception {
  AuthDomainFailure._();
  @override
  // TODO: implement changed
  bool? get changed => super.changed;

  factory AuthDomainFailure(String message,
      {Object? error,
      required StackTrace stackTrace,
      bool? changed}) = _AuthDomainFailure;

  factory AuthDomainFailure.internalServerError(
          {Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Internal Server Error",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.notFound({Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Not Found",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.conflict({Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        "Conflict",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );

  factory AuthDomainFailure.fromFirebase(String? message,
          {Object? error, StackTrace? stackTrace}) =>
      _$_AuthDomainFailure(
        message ?? "Error Form Firebase",
        error: error,
        stackTrace: stackTrace ?? StackTrace.current,
      );
}
```

## Flutter - 'initialValue == null || controller == null': is not true. error

### 문제

```log
'initialValue == null || controller == null': is not true
```

### 원인

```dart
TextFormField(
    key: SignUpScreen.emailKey,
    // controller: _emailController, // TODO uncomment
    decoration: InputDecoration(
    labelText: l10n.labelEmail,
    hintText: l10n.hintEmail,
    ),
    autovalidateMode: AutovalidateMode.onUserInteraction,
    validator: _validateEmail,
    autocorrect: false,
    textInputAction: TextInputAction.next,
    keyboardType: TextInputType.emailAddress,
    keyboardAppearance: Brightness.light,
    inputFormatters: <TextInputFormatter>[
    defaultNoSpaceTextInputValidator
    ],
    initialValue: "test@test.com", // TODO delete
),
```

- `controller`와 `initialValue`를 동시에 사용할 수 없다

```dart
class TextFormField extends FormField<String> {
  /// Creates a [FormField] that contains a [TextField].
  ///
  /// When a [controller] is specified, [initialValue] must be null (the
  /// default). If [controller] is null, then a [TextEditingController]
  /// will be constructed automatically and its `text` will be initialized
  /// to [initialValue] or the empty string.
  ///
  /// For documentation about the various parameters, see the [TextField] class
  /// and [TextField.new], the constructor.
  TextFormField({
    ... 생략 ...
    bool canRequestFocus = true,
  }) : assert(initialValue == null || controller == null),
```

### 해결

- 초기값 사용 위해 `controller`를 주석 처리

## state가 업데이트 되지 않는 이슈

### 문제

```dart
@Riverpod()
class AuthDomainService extends _$AuthDomainService {
  final log = Logger();

  @override
  Future<AuthDomainState> build() async {
    return AuthDomainState();
  }

  ReaderTask<AuthRepositoryInterface, void> createUserByEmailAndPassword(
          CreateUserByEmailCommand cmd) =>
      ReaderTask.Do((_) async {
        state = const AsyncValue.loading();
        final taskEither = await _(ReaderTask(
          (userRepo) async => userRepo.createUserWithEmailAndPassword(cmd),
        ));

        final either = await taskEither.run();

        state = either.fold(
            (left) =>
                AsyncValue.error(AuthDomainFailure(left), left.stackTrace),
            (right) => AsyncValue.data(AuthDomainSuccess(appUser: right)));
      });
}
```

아주 신기하게도 위와 같이 도메인 레이어에 메서드를 정의하고, 아래와 같이 사용할 때,

```dart
  Future<void> signUp(SignUpInfo signUpInfoDTO) async {
    state = const AsyncValue.loading();

    final readerTask = ref
        .watch(authDomainServiceProvider.notifier)
        .createUserByEmailAndPassword(CreateUserByEmailCommand(
          name: signUpInfoDTO.name,
          email: signUpInfoDTO.email,
          phone: signUpInfoDTO.phone,
          phoneVerified: signUpInfoDTO.phoneVerified,
          password: signUpInfoDTO.password,
        ));

    await readerTask.run(ref.read(authRepositoryProvider));

    final currentState = ref.watch(authDomainServiceProvider);
    if (currentState.hasError) {
      final failState = currentState.error as AuthDomainFailure;
      state = AsyncError(failState.error, failState.error.stackTrace);
      return;
    }

    final successState = currentState.value as AuthDomainSuccess;
    //                                     ^^^^
    //                                     Error: Expected a value of type 'AuthDomainSuccess',
    //                                     but got one of type 'Null'
    final result = await ref
        .read(userRepositoryProvider)
        .upsertAppUser(successState.appUser)
        .run();

    state = result.fold(
      (left) => AsyncError(
        AuthApplicationFailure(ApplicationErrors.fromDomain(left)),
        left.stackTrace,
      ),
      (right) => AsyncData(
        AuthApplicationSuccess(appUserDTO: AppUserDTO.fromDomain(right)),
      ),
    );
  }
```

`currentState.value as AuthDomainSuccess;` 타입 캐스팅이 실패한다

```log
Error: Expected a value of type 'AuthDomainSuccess', but got one of type 'Null'
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49        throw_
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 127:3         castError
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/operations.dart 742:12    cast
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/classes.dart 660:14       as_C
packages/dilectio_app/src/subprojects/user/application/auth_application.dart 88:45  signUp
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 45:50                  <fn>
dart-sdk/lib/async/zone.dart 1661:54                                                runUnary
dart-sdk/lib/async/future_impl.dart 156:18                                          handleValue
dart-sdk/lib/async/future_impl.dart 840:44                                          handleValueCallback
dart-sdk/lib/async/future_impl.dart 869:13                                          _propagateToListeners
dart-sdk/lib/async/future_impl.dart 641:5                                           [_completeWithValue]
dart-sdk/lib/async/future_impl.dart 715:7                                           callback
dart-sdk/lib/async/schedule_microtask.dart 40:11                                    _microtaskLoop
dart-sdk/lib/async/schedule_microtask.dart 49:5                                     _startMicrotaskLoop
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 181:15                 <fn>
```

### 원인

- 정확하진 않지만, `read`가 현재 상태를 읽기만 하므로, 그 내부적으로 `state`를 업데이트할 때, 실제로 수정이 이뤄지지 않는 거 같다

### 해결

```dart
  Future<void> signUp(SignUpInfo signUpInfoDTO) async {
    state = const AsyncValue.loading();

    final readerTask = ref
        //.read(authDomainServiceProvider.notifier) // convert `read` to `watch`
        .watch(authDomainServiceProvider.notifier)
        .createUserByEmailAndPassword(CreateUserByEmailCommand(
          name: signUpInfoDTO.name,
          email: signUpInfoDTO.email,
          phone: signUpInfoDTO.phone,
          phoneVerified: signUpInfoDTO.phoneVerified,
          password: signUpInfoDTO.password,
        ));
```

- `read`가 아닌 `watch`로 바꾸면 상태가 업데이트되고, 원하는대로 작동한다.

## `read` 사용해야 할 곳에서 `watch` 사용 시 에러 발생한다

- Provider의 의존성이 변경되었지만 아직 Provider가 다시 빌드되지 않았을 때 ref 함수를 사용하면 발생하는 에러

### 문제

로그인 화면에서 `state` 상태가 변경되면 `account` 페이지로 이동시킨다. 이때 `GoRouter`의 `redirect` 속성에서 `ref.watch` 사용하면 익셉션이 발생한다

```dart
class _SignInContentsState extends ConsumerState<SignInContents>
    with TextValidator {

  @override
  Widget build(BuildContext context) {
    ref.listen<AsyncValue>(
      signInControllerProvider,
      (_, state) => state.when(
        data: (data) {
          if (data is SignInSuccess) {
            GoRouter.of(context).goNamed(AppRoute.account.name);
          } else {
            state.showAlertDialogOnError(context);
          }
        },
        .. 생략 ..
    );
```

```dart
final goRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    ... 생략 ...
    redirect: (context, state) {
      final loggedInState = ref.watch(loggedInStateProvider);
      //                    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
      ... 생략 ...
      return null;
    }
```

```log
══╡ EXCEPTION CAUGHT BY FOUNDATION LIBRARY ╞════════════════════════════════════════════════════════
The following assertion was thrown while dispatching notifications for GoRouteInformationProvider:
Assertion failed:
file:///Users/rody/.pub-cache/hosted/pub.dev/riverpod-2.4.0/lib/src/framework/element.dart:673:7
!_didChangeDependency
"Cannot use ref functions after the dependency of a provider changed but before the provider
rebuilt"
```

### 원인

- Riverpod의 Provider는 주어진 BuildContext에 대한 Provider의 현재 상태를 제공한다
- `watch`를 사용하면 Provider의 상태가 변경될 때마다 이를 감지하고 연결된 위젯을 다시 빌드한다
- `watch`를 `redirect`에서 사용하면 상태 변경이 일어날 때마다 리스닝이 발생한다.
    - 로그인 상태가 변경될 때마다 라우터가 다시 평가한다.
    - 즉, 로그인 상태가 변할 때마다 해당 Provider가 감지하여 redirect 함수를 다시 실행할 수 있다.
    - **이렇게 되면 라우터 설정이 변경되기 전에 redirect 함수가 다시 호출될 가능성이 높다**
    - 또한 go_route의 `redirect` 함수는 라우팅 로직을 처리하면서 특정 조건에 따라 다른 경로로 리디렉션을 수행한다. 이 과정에서 상태 변경을 트리거할 수 있으며, 이로 인해 Riverpod의 `watch` 함수가 문제를 일으킬 수 있다. 이미 **실행 중인 함수 스택이 있을 때 상태 변경**이 일어나게 된다.
- 즉, Riverpod의 `watch` 메서드는 상태가 변경될 때마다 이를 감지하여 관련된 위젯을 재구성(rebuild)하게 되는데, 이러한 재구성이 함수(`redirect`)의 실행 중간에 발생하면 그 함수(`redirect`) 스택은 아직 완료되지 않은 상태가 된다. 이 때 `watch` 메서드를 호출하면, "Cannot use ref functions after the dependency of a provider changed but before the provider rebuilt"(provider가 재구성되기 전에 ref 함수를 사용할 수 없다)는 에러가 발생하게 된다.

### 해결

`read`가 아닌 `watch`를 사용

## firebase firestore 저장 에러

### 문제

```dart
_hasValue: true
invalidValue: _$36_DefaultInfo (DefaultInfo(uid: eIMPZKK1h6Mf5SnP0Kw892D7gKt1, email: aimpugn@gmail.com, emailVerified: false, name: Some(김상현), phone: Some(01091851202), phoneVerified: Some(true), profileUris: None))
<error>: <Unexpected DWDS error for getObject: Unexpected error from chrome devtools:>
message: "Could not convert"
name: "dartObject"
```

```log
┌───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
│ #0   packages/dilectio_app/src/subprojects/user/application/auth_application.dart 80:19  <fn>
├┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
│ 💡 result is Left(DomainErrors(message: Internal Server Error, error: Invalid argument (dartObject): Could not convert: Instance of '_$36_DefaultInfo', stackTrace: dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49                            throw_
│ 💡 packages/firebase_core_web/src/interop/utils/utils.dart 95:5                                            jsify
│ 💡 packages/firebase_core_web/src/interop/utils/utils.dart 83:36                                           <fn>
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/private/linked_hash_map.dart 21:13                                forEach
│ 💡 packages/firebase_core_web/src/interop/utils/utils.dart 82:15                                           jsify
│ 💡 packages/cloud_firestore_web/src/interop/utils/utils.dart 42:23                                         jsify
│ 💡 packages/cloud_firestore_web/src/interop/firestore.dart 370:46                                          set
│ 💡 packages/cloud_firestore_web/src/document_reference_web.dart 33:23                                      <fn>
│ 💡 packages/_flutterfire_internals/_flutterfire_internals.dart 112:21                                      guardWebExceptions
│ 💡 packages/cloud_firestore_web/src/internals.dart 12:20                                                   convertWebExceptions
│ 💡 packages/cloud_firestore_web/src/document_reference_web.dart 32:12                                      set
│ 💡 packages/cloud_firestore/src/document_reference.dart 168:22                                             set
│ 💡 packages/dilectio_app/src/subprojects/user/infrastructure/repositories/user_repository_impl.dart 29:27  <fn>
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
│ 💡 packages/dilectio_app/src/subprojects/user/infrastructure/repositories/user_repository_impl.dart 26:9   <fn>
│ 💡 packages/fpdart/src/task_either.dart 280:39                                                             <fn>
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
│ 💡 packages/fpdart/src/task_either.dart 278:24                                                             <fn>
│ 💡 packages/fpdart/src/task_either.dart 183:37                                                             run
│ 💡 packages/dilectio_app/src/subprojects/user/application/auth_application.dart 78:45                      <fn>
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
│ 💡 packages/dilectio_app/src/subprojects/user/application/auth_application.dart 74:18                      <fn>
│ 💡 packages/riverpod/src/common.dart 686:16                                                                AsyncValueX.when
│ 💡 packages/dilectio_app/src/subprojects/user/application/auth_application.dart 73:41                      signUp
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 45:50                                      <fn>
│ 💡 dart-sdk/lib/async/zone.dart 1661:54                                                                    runUnary
│ 💡 dart-sdk/lib/async/future_impl.dart 156:18                                                              handleValue
│ 💡 dart-sdk/lib/async/future_impl.dart 840:44                                                              handleValueCallback
│ 💡 dart-sdk/lib/async/future_impl.dart 869:13                                                              _propagateToListeners
│ 💡 dart-sdk/lib/async/future_impl.dart 641:5                                                               [_completeWithValue]
│ 💡 dart-sdk/lib/async/future_impl.dart 715:7                                                               callback
│ 💡 dart-sdk/lib/async/schedule_microtask.dart 40:11                                                        _microtaskLoop
│ 💡 dart-sdk/lib/async/schedule_microtask.dart 49:5                                                         _startMicrotaskLoop
│ 💡 dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 181:15                                     <fn>
│ 💡 ))
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
Instance of 'AuthApplicationFailure', dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49                            throw_
packages/firebase_core_web/src/interop/utils/utils.dart 95:5                                            jsify
packages/firebase_core_web/src/interop/utils/utils.dart 83:36                                           <fn>
dart-sdk/lib/_internal/js_dev_runtime/private/linked_hash_map.dart 21:13                                forEach
packages/firebase_core_web/src/interop/utils/utils.dart 82:15                                           jsify
packages/cloud_firestore_web/src/interop/utils/utils.dart 42:23                                         jsify
packages/cloud_firestore_web/src/interop/firestore.dart 370:46                                          set
packages/cloud_firestore_web/src/document_reference_web.dart 33:23                                      <fn>
packages/_flutterfire_internals/_flutterfire_internals.dart 112:21                                      guardWebExceptions
packages/cloud_firestore_web/src/internals.dart 12:20                                                   convertWebExceptions
packages/cloud_firestore_web/src/document_reference_web.dart 32:12                                      set
packages/cloud_firestore/src/document_reference.dart 168:22                                             set
packages/dilectio_app/src/subprojects/user/infrastructure/repositories/user_repository_impl.dart 29:27  <fn>
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
packages/dilectio_app/src/subprojects/user/infrastructure/repositories/user_repository_impl.dart 26:9   <fn>
packages/fpdart/src/task_either.dart 280:39                                                             <fn>
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
packages/fpdart/src/task_either.dart 278:24                                                             <fn>
packages/fpdart/src/task_either.dart 183:37                                                             run
packages/dilectio_app/src/subprojects/user/application/auth_application.dart 78:45                      <fn>
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54                                      runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5                                      _async
packages/dilectio_app/src/subprojects/user/application/auth_application.dart 74:18                      <fn>
packages/riverpod/src/common.dart 686:16                                                                AsyncValueX.when
packages/dilectio_app/src/subprojects/user/application/auth_application.dart 73:41                      signUp
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 45:50                                      <fn>
dart-sdk/lib/async/zone.dart 1661:54                                                                    runUnary
dart-sdk/lib/async/future_impl.dart 156:18                                                              handleValue
dart-sdk/lib/async/future_impl.dart 840:44                                                              handleValueCallback
dart-sdk/lib/async/future_impl.dart 869:13                                                              _propagateToListeners
dart-sdk/lib/async/future_impl.dart 641:5                                                               [_completeWithValue]
dart-sdk/lib/async/future_impl.dart 715:7                                                               callback
dart-sdk/lib/async/schedule_microtask.dart 40:11                                                        _microtaskLoop
dart-sdk/lib/async/schedule_microtask.dart 49:5                                                         _startMicrotaskLoop
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 181:15                                     <fn>
```

```dart
// /Users/rody/.pub-cache/hosted/pub.dev/firebase_core_web-2.8.1/lib/src/interop/utils/utils.dart
/// Returns the JS implementation from Dart Object.
///
/// The optional [customJsify] function may return `null` to indicate,
/// that it could not handle the given Dart Object.
dynamic jsify(
  Object? dartObject, [
  Object? Function(Object? object)? customJsify,
]) {
  if (_isBasicType(dartObject)) {
    if (dartObject == null) {
      return null;
    }
    return dartObject;
  }

  if (dartObject is Iterable) {
    return jsifyList(dartObject, customJsify);
  }

  if (dartObject is Map) {
    var jsMap = util.newObject();
    dartObject.forEach((key, value) {
      util.setProperty(jsMap, key, jsify(value, customJsify));
    });
    return jsMap;
  }

  if (dartObject is Function) {
    return allowInterop(dartObject);
  }

  Object? value = customJsify?.call(dartObject);

  if (value == null) {
    throw ArgumentError.value(dartObject, 'dartObject', 'Could not convert');
  }

  return value;
}
```

### 원인

- [Datatype conversion of Dart "color" to "string" when in a "list" #1201](https://github.com/FlutterFlow/flutterflow-issues/issues/1201)

```dart
import 'package:fpdart/fpdart.dart';
import 'package:freezed_annotation/freezed_annotation.dart';

part 'default_info.freezed.dart';
part 'default_info.g.dart';

@freezed
class DefaultInfo with _$DefaultInfo {
  const factory DefaultInfo({
    required String uid,
    required String email,
    required bool emailVerified,
    required Option<String> name,
    required Option<String> phone,
    required Option<bool> phoneVerified,
    required Option<List<Uri>>? profileUris,
  }) = _DefaultInfo;

  factory DefaultInfo.fromJson(Map<String, dynamic> json) =>
      _$DefaultInfoFromJson(json);
}
```

- `DefaultInfo`는 nested 오브젝트 `@JsonSerializable(explicitToJson: true)`가 추가되어야 한다

> Note:
> In order to serialize **nested** lists of freezed objects, you are supposed to either specify a `@JsonSerializable(explicitToJson: true)` or change `explicit_to_json` inside your `build.yaml` file (see the [documentation](https://github.com/google/json_serializable.dart/tree/master/json_serializable#build-configuration)).

### 해결

```dart
import 'package:fpdart/fpdart.dart';
import 'package:freezed_annotation/freezed_annotation.dart';

part 'default_info.freezed.dart';
part 'default_info.g.dart';

@freezed
class DefaultInfo with _$DefaultInfo {
  @JsonSerializable(explicitToJson: true)
  const factory DefaultInfo({
    required String uid,
    required String email,
    required bool emailVerified,
    required Option<String> name,
    required Option<String> phone,
    required Option<bool> phoneVerified,
    required Option<List<Uri>>? profileUris,
  }) = _DefaultInfo;

  factory DefaultInfo.fromJson(Map<String, dynamic> json) =>
      _$DefaultInfoFromJson(json);
}
```

## RenderFlex children have non-zero flex but incoming height constraints are unbounded

### 문제

```log
... 원래 정상적인 로그 생략 ...
══╡ EXCEPTION CAUGHT BY RENDERING LIBRARY ╞═════════════════════════════════════════════════════════
The following assertion was thrown during performLayout():
RenderFlex children have non-zero flex but incoming height constraints are unbounded.
When a column is in a parent that does not provide a finite height constraint, for example if it is
in a vertical scrollable, it will try to shrink-wrap its children along the vertical axis. Setting a
flex on a child (e.g. using Expanded) indicates that the child is to expand to fill the remaining
space in the vertical direction.
These two directives are mutually exclusive. If a parent is to shrink-wrap its child, the child
cannot simultaneously expand to fit its parent.
Consider setting mainAxisSize to MainAxisSize.min and using FlexFit.loose fits for the flexible
children (using Flexible rather than Expanded). This will allow the flexible children to size
themselves to less than the infinite remaining space they would otherwise be forced to take, and
then will cause the RenderFlex to shrink-wrap the children rather than expanding to fit the maximum
constraints provided by the parent.
If this message did not help you determine the problem, consider using debugDumpRenderTree():
  https://flutter.dev/debugging/#rendering-layer
  http://api.flutter.dev/flutter/rendering/debugDumpRenderTree.html
The affected RenderFlex is:
  RenderFlex#3c228 relayoutBoundary=up27 NEEDS-LAYOUT NEEDS-PAINT NEEDS-COMPOSITING-BITS-UPDATE(creator: Column ← _FormScope ← WillPopScope ← Form-[LabeledGlobalKey<FormState>#e5ee0] ← _FocusInheritedScope ← Semantics ← FocusScope ← Padding ← Semantics ← DefaultTextStyle ← AnimatedDefaultTextStyle ← _InkFeatures-[GlobalKey#606ec ink renderer] ← ⋯, parentData: <none> (can use size), constraints: BoxConstraints(w=528.0, 0.0<=h<=Infinity), size: MISSING, direction: vertical, mainAxisAlignment: start, mainAxisSize: max, crossAxisAlignment: stretch, verticalDirection: down)
The creator information is set to:
  Column ← _FormScope ← WillPopScope ← Form-[LabeledGlobalKey<FormState>#e5ee0] ←
  _FocusInheritedScope ← Semantics ← FocusScope ← Padding ← Semantics ← DefaultTextStyle ←
  AnimatedDefaultTextStyle ← _InkFeatures-[GlobalKey#606ec ink renderer] ← ⋯
The nearest ancestor providing an unbounded width constraint is: _RenderSingleChildViewport#6ed73 relayoutBoundary=up14 NEEDS-LAYOUT NEEDS-PAINT NEEDS-COMPOSITING-BITS-UPDATE:
  needs compositing
  creator: _SingleChildViewport ← IgnorePointer-[GlobalKey#c8522] ← Semantics ← Listener ←
    _GestureSemantics ← RawGestureDetector-[LabeledGlobalKey<RawGestureDetectorState>#4929c] ←
    Listener ← _ScrollableScope ← _ScrollSemantics-[GlobalKey#e2ec0] ←
    NotificationListener<ScrollMetricsNotification> ← RepaintBoundary ← CustomPaint-[GlobalKey#c75d8]
    ← ⋯
  parentData: <none> (can use size)
  constraints: BoxConstraints(0.0<=w<=1200.0, 0.0<=h<=744.0)
  size: MISSING
  offset: Offset(0.0, -0.0)
See also: https://flutter.dev/layout/
If none of the above helps enough to fix this problem, please don't hesitate to file a bug:
  https://github.com/flutter/flutter/issues/new?template=2_bug.yml
The relevant error-causing widget was:
  Column
Column:file:///Users/rody/VscodeProjects/dilectio/dilectio_app/lib/src/subprojects/user/presentation/sign_in/sign_in_screen.dart:140:18
When the exception was thrown, this was the stack:
... stack trace ...
```

- 위의 에러는 Flutter 레이아웃 시스템에서 발생한 것으로, `RenderFlex` 관련 에러
- 이 에러의 주요 원인은 `Column` 위젯 내에 `Spacer` 또는 `Expanded` 같은 유연한 위젯을 사용했으나, `Column`이 무한 높이(`unbounded height`)를 가진 부모 위젯(예: `SingleChildScrollView`) 내에 있기 때문

### 원인

- 위 에러는 `Column`이 무한 높이를 가진 환경에서 `Expanded` 또는 `Spacer`와 같은 유연한 위젯이 사용되었음을 나타낸다
- `Expanded`나 `Spacer`는 사용 가능한 공간을 채우려고 하지만, `SingleChildScrollView`와 같은 스크롤 가능한 위젯은 자식에게 제한 없는 높이를 제공하기 때문에, `Expanded` 또는 `Spacer`가 얼마나 많은 공간을 차지해야 할지 결정할 수 없다

### 플러터 레이아웃 이해

- **제한 조건 (Constraints):** 부모 위젯이 자식에게 주는 크기 제한. 부모가 무한 높이나 너비를 제공할 수도 있음.
- **유연성 (Flexibility):** `Expanded`와 `Flexible` 위젯은 주어진 공간 내에서 자식의 크기를 어떻게 조절할지 결정한다
- **스크롤 가능한 컨텍스트:** `SingleChildScrollView` 같은 스크롤 가능한 위젯은 자식에게 무한 높이를 제공할 수 있으며, 이는 레이아웃 문제를 일으킬 수 있다

### 해결

1. **`mainAxisSize` 설정:**
   `Column`의 `mainAxisSize` 속성을 `MainAxisSize.min`으로 설정하여, `Column`이 자식들의 크기에 맞게 크기를 조절하도록 한다. 이렇게 하면 `Spacer`나 `Expanded`가 무한 공간을 차지하려는 시도를 하지 않는다

   ```dart
   Column(
     mainAxisSize: MainAxisSize.min,
     children: [...],
   )
   ```

2. **`Flexible` 위젯 사용:**
   `Expanded` 대신 `Flexible` 위접을 사용하고, `fit` 속성을 `FlexFit.loose`로 설정한다. 이렇게 하면 자식 위젯이 더 유연하게 크기를 조절할 수 있다

   ```dart
   Flexible(
     fit: FlexFit.loose,
     child: YourWidget(),
   )
   ```

3. **레이아웃 구조 재고려:**
   현재 레이아웃 구조가 스크롤 가능한 위젯 내에 유연한 위젯을 사용하는 것이 적합한지 고려합니다. 필요에 따라 다른 레이아웃 위젯을 사용하거나, 스크롤 동작을 다르게 처리하는 방식을 고려해 볼 수 있습니다.

## Vertical viewport was given unbounded height

### 문제

```log
The following assertion was thrown during performResize():
Vertical viewport was given unbounded height.
Viewports expand in the scrolling direction to fill their container. In this case, a vertical
viewport was given an unlimited amount of vertical space in which to expand. This situation
typically happens when a scrollable widget is nested inside another scrollable widget.
If this widget is always nested in a scrollable widget there is no need to use a viewport because
there will always be enough vertical space for the children. In this case, consider using a Column
or Wrap instead. Otherwise, consider using a CustomScrollView to concatenate arbitrary slivers into
a single scrollable.
The relevant error-causing widget was:
  ListView
ListView:file:///Users/rody/VscodeProjects/dilectio/dilectio_app/lib/src/subprojects/user/presentation/sign_in/sign_in_screen.dart:140:18
sign_in_screen.dart:140
When the exception was thrown, this was the stack:
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49  throw_
```

### 원인

- 이 에러 메시지는 Flutter 레이아웃 시스템에서 `ListView` 위젯이 무제한 높이의 공간에 배치되었을 때 발생
- 특히 "Vertical viewport was given unbounded height"라는 부분이 핵심
- 이는 `ListView` (또는 다른 스크롤 가능한 위젯)가 **스크롤 방향으로 무한한 공간을 확장하려고 할 때 발생**
- 이 문제는 일반적으로 스크롤 가능한 위젯이 다른 스크롤 가능한 위젯 안에 중첩되었을 때 나타난다
- **무한한 확장:**
    - 스크롤 가능한 위젯들 (`ListView`, `SingleChildScrollView` 등)은 자신이 차지할 수 있는 공간을 무한대로 가정한다
    - 즉, 이들은 자신을 포함하는 부모 위젯이 제공하는 공간을 기준으로 크기를 결정한다
- **중첩된 스크롤:**
    - 만약 `ListView`가 다른 스크롤 가능한 위젯 안에 위치한다면, `ListView`는 사용 가능한 공간의 한계를 알 수 없으므로, 얼마나 많은 공간을 차지해야 할지 결정할 수 없다

### 해결

1. **중첩된 스크롤 구조 재고려:**
    - 만약 `ListView`가 다른 스크롤 가능한 위젯 안에 있지 않도록 구조를 변경할 수 있다면, 이 문제를 해결할 수 있다
    - 예를 들어, `ListView`를 `Column`이나 `Wrap` 같은 정적인 레이아웃 위젯으로 대체할 수 있다
2. **`CustomScrollView` 사용:**
    - 여러 스크롤 가능한 위젯을 연결해야 한다면, `CustomScrollView`를 사용하여 다양한 'slivers' (예: `SliverList`, `SliverAppBar` 등)를 하나의 스크롤 가능한 뷰로 결합할 수 있다
    - `CustomScrollView`는 중첩된 스크롤 상황에서 더 잘 동작한다
3. **제한된 공간 제공:**
    - `ListView`를 포함하는 위젯에 명확한 높이 제약을 제공한다
    - 예를 들어, `ListView`를 `Container` 또는 `SizedBox`로 감싸서 특정 높이를 지정할 수 있다
4. **중첩된 스크롤 가능한 위젯 피하기:**
    - 가능하다면, 스크롤 가능한 위젯을 다른 스크롤 가능한 위젯 안에 넣는 것을 피해야 한다.
    - 대신, 전체 페이지에 하나의 스크롤 가능한 위젯만 사용하는 것이 좋다

## `TextFormField`의 `validator`가 두 번 호출됨

### 문제

### 원인

```dart
  const GrayFilledTextFormField({
    required this.textFormFieldKey,
    required this.controller,
    required this.keyboardType,
    this.textInputAction,
    this.labelName,
    this.labelText,
    this.obscureText,
    this.labelTextStyle,
    this.hintText,
    this.inputFormatters,
    this.validator,
    this.suffix,
    this.suffixIcon,
    this.decoration,
    this.autocorrect,
    this.enableSuggestions,
    this.inputDecorationEnabled,
    this.onEditingComplete,
  }) : super(key: textFormFieldKey);
```

- Flutter에서 위젯의 `key`를 사용하는 방식을 보면, `GrayFilledTextFormField` 위젯에 전달된 `textFormFieldKey`는 두 가지 용도로 사용되고 있다
    - 하나는 `GrayFilledTextFormField`의 생성자를 통해 `super(key: textFormFieldKey)`로 전달
    - 다른 하나는 내부의 `TextFormField` 위젯에도 동일한 `key`가 할당
- Flutter에서 `key`는 특정 위젯의 인스턴스를 유일하게 식별하는 데 사용된다. 따라서 Flutter 프레임워크가 위젯 트리에서 위젯을 올바르게 식별하고 추적하는 데 혼란을 줄 수 있기 때문에, 하나의 `key`를 여러 위젯에 동시에 사용하는 것은 권장되지 않는다.
- `GrayFilledTextFormField`의 `super(key: textFormFieldKey)` 호출은 `GrayFilledTextFormField` 위젯 자체를 식별하는 데 사용되며, 이것이 `TextFormField`에도 동일하게 사용되면, Flutter 프레임워크는 두 위젯을 동일한 것으로 오인할 수 있다. 이로 인해 위젯의 상태 관리, 렌더링, 그리고 트리 재구성(rebuilding) 과정에서 예기치 않은 문제가 발생할 수 있다.
- 이 문제를 해결하기 위해서는 각 위젯에 고유한 `key`를 할당하는 것이 좋다. 예를 들어, `GrayFilledTextFormField`와 내부의 `TextFormField`에 각각 다른 `key`를 사용하면 이러한 문제를 방지할 수 있다

#### Key 매개변수 추가해보기

- Key 매개변수 추가하기
    - Flutter에서 `key`는 위젯의 고유성을 보장하고, 위젯 트리에서 위젯의 위치를 추적하는 데 사용된다.
    - `key`의 주요 목적은 Flutter 프레임워크가 위젯을 올바르게 식별하고, 상태 관리 및 렌더링 최적화를 위해 위젯의 생명 주기를 관리하는 데 있다.
    - 따라서 위젯의 생성자에 `Key` 타입의 선택적(named) 매개변수를 추가하고, **이 매개변수는 `super` 호출에 전달되어야** 한다.
- 즉 **내부 위젯에 Key 전달하지 않아야** 한다. **내부 `TextFormField`에** 다른 `key`를 사용하거나, **아예 `key`를 할당하지 않아야** 한다
- `GrayFilledTextFormField` 위젯은 `Key` 매개변수를 받고, 이 `key`는 위젯을 식별하는 데 사용되며, 내부 `TextFormField`에는 별도의 `key`를 할당하지 않는다. 이렇게 하면 Flutter 프레임워크가 위젯의 생명 주기와 상태를 올바르게 관리할 수 있다. 이러한 접근 방식은 Flutter의 권장사항을 따르며, 위젯의 고유성을 보장하는 동시에, 내부 위젯 간의 잠재적인 충돌을 방지한다

```dart
class GrayFilledTextFormField extends StatelessWidget {
  // 나머지 매개변수들...

  // 생성자에 Key 매개변수 추가
  const GrayFilledTextFormField({
    Key? key, // Key 매개변수 추가
    required this.controller,
    required this.keyboardType,
    // 나머지 매개변수들...
  }) : super(key: key); // super 호출에 key 전달

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // ...
        Expanded(
          child: TextFormField(
            // key: textFormFieldKey, // 내부 TextFormField에는 별도의 key를 할당하지 않음
            // 나머지 코드...
          ),
        )
      ],
    );
  }
}
```

#### 왜 내부 `TextFormField`에 별도의 `key`를 할당하지 않아도 되는가?

- **위젯 식별**
    - `TextFormField`와 같은 폼 필드 위젯은 대부분의 경우 고유한 `key`가 필요하지 않다.
    - 이는 `TextFormField`의 상태(예: 입력된 텍스트)가 주로 `TextEditingController`를 통해 관리되기 때문
    - 따라서, 위젯 트리에서 해당 위젯의 위치나 고유성을 식별할 필요가 거의 없다
- **상태 관리**
    - `TextEditingController`와 같은 컨트롤러를 사용하는 경우, 위젯의 상태가 컨트롤러에 의해 관리된다
    - 이 경우 위젯 자체의 `key`보다는 컨트롤러가 중요한 역할을 하게 된다
- **위젯 트리의 변화**
    - Flutter에서 위젯의 `key`는 위젯이 위젯 트리에서 재배치되거나, 위젯 트리 내에서 동일한 유형의 다른 위젯과 구별될 필요가 있을 때 중요하다.
    - `TextFormField`와 같은 단순한 폼 필드의 경우, 위젯 트리에서 자주 재배치되거나, 고유성이 필수적인 상황이 드물기 때문에 `key`가 필수적이지 않다
- **성능 최적화**
    - 모든 위젯에 `key`를 할당하면 Flutter의 성능 최적화 기능을 방해할 수 있다.
    - 불필요하게 `key`를 사용하면 Flutter가 위젯의 불필요한 재생성을 피하기 위해 추가적인 작업을 수행해야 할 수 있다

즉, `key`는 위젯의 고유성을 관리하고 위젯 트리에서의 위치 변화를 추적하는 데 중요하지만, 모든 위젯에 `key`를 사용할 필요는 없다.
특히 `TextFormField`와 같은 폼 필드에서는 상태 관리가 컨트롤러를 통해 이루어지기 때문에 `key`의 사용이 필수적이지 않다.
`key`의 사용은 위젯의 특정 요구사항과 상황에 따라 결정되어야 하며, 필요하지 않은 경우에는 생략하는 것이 좋다.

### 해결

로직이 잘못된 거였다

```diff
  Future<void> _login() async {
    setState(() => _submitted = true);
    // only submit the form if validation passes
    if (_signInFormKey.currentState!.validate()) {
      await ref
          .read(authApplicationProvider.notifier)
          .loginByEmail(email, password);
+     setState(() => _submitted = false);
    }
-   setState(() => _submitted = false);
  }
```

```dart
validator: (email) =>
    !_submitted ? null : emailErrorText(email),
```

`_submitted` 상태가 login이 되지도 않았는데 false로 바뀌면서 `!_submitted`은 true가 되고 null이 리턴되면서 제대로 에러 메시지가 출력되지 않았던 것

## The return type 'Null' isn't a 'AppUser', as required by the closure's context

### 문제

```dart
  TaskEither<AuthDomainFailure, AuthDomainSuccess> createUserByEmailAndPassword(
          CreateUserByEmailCommand cmd) =>
      TaskEither.Do((_) async {
        // 첫 번째 단계: 가능한 사용자 ID인지 확인
        final isPossibleUserIDEither =
            await userRepo.isPossibleUserID(cmd.email).run();
        if (isPossibleUserIDEither.isLeft()) {
          return isPossibleUserIDEither;
        }
        final isPossibleUserID = isPossibleUserIDEither.getOrElse((_) => false);
        if (!isPossibleUserID) {
          return Left(DomainErrors.conflict(
              error: Object(), stackTrace: StackTrace.current));
        }

        // 두 번째 단계: 사용자 생성
        final createdUserEither =
            await authRepo.createUserWithEmailAndPassword(cmd).run();
        if (createdUserEither.isLeft()) {
          return createdUserEither.mapLeft(identity);
        }
        final appUser = createdUserEither.getOrElse((_) => null);
                                          ^^^^^^^^^^^^^^^^^^^^^^ Error!
        if (appUser == null) {
          return Left(DomainErrors.fromFirebase(
              error: Object(), stackTrace: StackTrace.current));
        }

        return Right(appUser);
      }).mapLeft(AuthDomainFailure.new).map(AuthDomainSuccess.new);
                                            ^^^^^^^^^^^^^^^^^^^^^^ Error!
```

그래서 아래와 같이 `TaskEither<DomainErrors, AppUser>.Do((_) async {` 타입을 명시하고, `Left`, `Right`로 리턴을 해봤다

```dart
  TaskEither<AuthDomainFailure, AuthDomainSuccess> createUserByEmailAndPassword(
          CreateUserByEmailCommand cmd) =>
      TaskEither<DomainErrors, AppUser>.Do((_) async {
        // 첫 번째 단계: 가능한 사용자 ID인지 확인
        final isPossibleUserIDEither =
            await userRepo.isPossibleUserID(cmd.email).run();

        final errorOrIsPossibleUserId =
            isPossibleUserIDEither.fold(identity, identity);
        if (errorOrIsPossibleUserId is DomainErrors) {
          return Left(errorOrIsPossibleUserId);
        }

        final isPossibleUserID = errorOrIsPossibleUserId as bool;
        if (!isPossibleUserID) {
          return Left(DomainErrors.conflict(
              error: Object(), stackTrace: StackTrace.current));
        }

        // 두 번째 단계: 사용자 생성
        final createdUserEither =
            await authRepo.createUserWithEmailAndPassword(cmd).run();

        final errorOrAppUser = createdUserEither.fold(identity, identity);
        if (errorOrAppUser is DomainErrors) {
          return Left(errorOrAppUser);
        }

        final appUser = errorOrAppUser as AppUser;

        final marked = await userRepo.upsertUserID(cmd.email).run();
        final errorOrmarked = marked.fold(identity, identity);
        if (errorOrmarked is DomainErrors) {
          return Left(errorOrmarked);
        }

        return appUser;
      }).mapLeft(AuthDomainFailure.new).map(AuthDomainSuccess.new);
```

그랬더니 `Left(errorOrIsPossibleUserId)`, `Left(DomainErrors.conflict(error: Object(), stackTrace: StackTrace.current))`, `Left(errorOrAppUser)`, `Left(errorOrmarked)`에서 아래와 같은 에러가 발생한다

```log
The return type 'Left<DomainErrors, dynamic>' isn't a 'Future<AppUser>', as required by the closure's context.
```

Do 표기법에서 사용되는 클로저 컨텍스트에서 return은 `Future<AppUser>`이길 기대하는 것으로 보이는데,`TaskEither<DomainErrors, AppUser>.Do`에서 Left는`DomainErrors` 타입이라고 제가 지정을 했다. 뭐가 잘못된 걸까?

### 원인

- `getOrElse` 함수는 `Either`의 `Right` 값이 없을 경우 사용할 기본값을 제공하는데, 이 기본값은 `Right` 타입과 일치해야 한다
- `createdUserEither`의 `Right` 값은 `?`가 없는 `AppUser`인데, 기본값이 `null`로 리턴되기 때문에 dart의 타입 시스템이 이를 허용하지 않는 것
- 그리고 `Do` 표기법 내에서 `Left`, `Right`로 리턴하는 것은 문제 해결이 되지 않는다. 주어지는 `_` 함수를 사용해야 한다.

#### `_` 함수

- `_` 함수를 사용하는 것이 문제를 해결한 이유는 이 함수가 `TaskEither`의 제네릭 타입을 유지하면서 `Future`를 반환하는 데 필요한 **브리징 역할**을 하기 때문이다
- Dart에서는 타입이 매우 엄격하게 적용되므로, `Future<A> Function<A>(TaskEither<DomainErrors, A>) _`와 같은 함수는 `TaskEither`를 `Future`로 변환하면서도 제네릭 타입을 유지하는 역할한다
- 이 함수의 작동 원리는
    - **타입 유지**
        - 이 함수는 `TaskEither<DomainErrors, A>`의 인스턴스를 받고, 같은 `TaskEither<DomainErrors, A>` 타입을 반환한다
        - 이는 `TaskEither.Do` 블록 안에서 각 반환값이 `TaskEither<DomainErrors, AppUser>` 타입을 유지해야 한다는 요구 사항을 충족시킨다
    - **`Future` 반환**
        - Dart의 `async` 함수는 항상 `Future`를 반환해야 한다
        - 이 함수는 `TaskEither` 인스턴스를 `Future`로 래핑하여 이 요구 사항을 충족한다
        - 즉, `TaskEither.Do` 블록 안의 각 단계에서 `Future<TaskEither<DomainErrors, AppUser>>`를 반환할 수 있게 해준다
    - **제네릭 타입**
        - 함수의 제네릭 타입 `A`는 `TaskEither`의 `Right` 값의 타입을 나타낸다
        - 이를 통해 `TaskEither`의 `Left`와 `Right` 값 모두에 대한 타입 안정성을 제공한다
- 결론적으로, `_` 함수는
    - `TaskEither.Do` 블록 내에서 `TaskEither`의 타입을 유지하면서 비동기 함수가 요구하는 `Future`를 반환하는 중간자 역할을 하고,
    - 이를 통해 타입 시스템의 요구 사항을 충족시키고,
    - 코드의 정확성과 타입 안정성을 보장한다

이러한 함수는 함수형 프로그래밍에서 자주 볼 수 있는 패턴으로, 복잡한 타입 체계를 가진 언어에서 비동기 작업과 타입 안정성을 동시에 유지하기 위해 사용된다

### 해결

```dart
  TaskEither<AuthDomainFailure, AuthDomainSuccess> createUserByEmailAndPassword(
          CreateUserByEmailCommand cmd) =>
      TaskEither<DomainErrors, AppUser>.Do((_) async {
        // 첫 번째 단계: 가능한 사용자 ID인지 확인
        final isPossibleUserIDEither =
            await userRepo.isPossibleUserID(cmd.email).run();

        final errorOrIsPossibleUserId =
            isPossibleUserIDEither.fold(identity, identity);
        if (errorOrIsPossibleUserId is DomainErrors) {
          return _(TaskEither.left(errorOrIsPossibleUserId));
        }

        final isPossibleUserID = errorOrIsPossibleUserId as bool;
        if (!isPossibleUserID) {
          return _(TaskEither.left(DomainErrors.conflict(
              error: Object(), stackTrace: StackTrace.current)));
        }

        // 두 번째 단계: 사용자 생성
        final createdUserEither =
            await authRepo.createUserWithEmailAndPassword(cmd).run();

        final errorOrAppUser = createdUserEither.fold(identity, identity);
        if (errorOrAppUser is DomainErrors) {
          return _(TaskEither.left(errorOrAppUser));
        }

        final appUser = errorOrAppUser as AppUser;

        final marked = await userRepo.upsertUserID(cmd.email).run();
        final errorOrmarked = marked.fold(identity, identity);
        if (errorOrmarked is DomainErrors) {
          return _(TaskEither.left(errorOrmarked));
        }

        return _(TaskEither.right(appUser));
      }).mapLeft(AuthDomainFailure.new).map(AuthDomainSuccess.new);
```

## Error: MissingPluginException(No implementation found for method getAll on channel plugins.flutter.io/shared_preferences)

### 문제

```log
Error: MissingPluginException(No implementation found for method getAll on channel plugins.flutter.io/shared_preferences)
```

### 원인

아마도 패키지 다운로드가 제대로 안 됐던 게 아닌가 싶다

### 해결

- `flutter clean` 후 재실행

## 로그가 두 번 찍히는 문제

### 문제

```dart
class AccountScreen extends ConsumerStatefulWidget {
  const AccountScreen({super.key});

  @override
  ConsumerState<ConsumerStatefulWidget> createState() =>
      AccountScreenContents();
}

/// Simple user data table showing the uid and email
class AccountScreenContents extends ConsumerState<AccountScreen> {
  final log = Logger();

  @override
  Widget build(BuildContext context) {
    log.i("AccountScreenContents");
    final l10n = AppLocalizations.of(context)!;
    final state = ref.watch(userManageApplicationProvider);

    return Scaffold(
```

```dart
          GoRoute(
            path: AppRoute.account.path,
            name: AppRoute.account.name,
            pageBuilder: (context, state) => CustomTransitionPage(
              name: AppRoute.account.name,
              fullscreenDialog: true,
              child: const AccountScreen(),
```

```log
┌───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
│ #0   packages/dilectio_app/src/subprojects/user/presentation/account/account_screen.dart 32:9  build
│ #1   packages/flutter/src/widgets/framework.dart 5409:27                                       build
├┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
│ 💡 AccountScreenContents
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
┌───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
│ #0   packages/dilectio_app/src/subprojects/user/presentation/account/account_screen.dart 32:9  build
│ #1   packages/flutter/src/widgets/framework.dart 5409:27                                       build
├┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
│ 💡 AccountScreenContents
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
```

이렇게 로그가 두번 출력되는데... `AccountScreenContents`는 한번만 생성하고 있고, `AccountScreen`도 const로 한번만 생성되고 있다

### 원인

디버깅 방법

- 상태 관리 로직 검토: `userManageApplicationProvider`와 관련된 상태 관리 로직을 검토하여, 상태 변화가 의도대로 이루어지는지 확인
- 부모 위젯의 영향 분석: `AccountScreenContents`의 부모 위젯이 어떻게 구성되어 있는지 검토하여, 부모 위젯의 상태 변화가 하위 위젯에 어떤 영향을 미치는지 파악
- Flutter 위젯 라이프사이클 이해: Flutter의 위젯 라이프사이클과 빌드 과정을 이해하고, 어떤 단계에서 `build` 메서드가 호출되는지 분석
- 라우팅 로직 확인: `GoRouter`와 관련된 코드를 검토하여, 라우팅 과정에서 위젯의 빌드가 어떻게 관리되는지 확인

#### 상태 변경에 의한 재빌드?

`ref.watch(userManageApplicationProvider)` 호출로 인해, `userManageApplicationProvider`의 상태가 변경되면 `AccountScreenContents` 위젯은 다시 빌드됩니다. 상태가 초기화되거나 업데이트될 때 `build` 메서드가 여러 번 호출될 수 있습니다.

#### Flutter 프레임워크의 최적화 동작?

Flutter 프레임워크는 성능 최적화를 위해 위젯 트리의 빌드 과정에서 때때로 위젯을 불필요하게 재빌드할 수 있습니다. 특히, 복잡한 위젯 트리나 상태 관리 로직이 있는 경우, Flutter는 더 나은 사용자 경험을 위해 추가적인 빌드를 수행할 수 있습니다.

#### 외부 요인에 의한 재렌더링?

외부 요인, 예를 들어 네트워크 요청, 타이머, 사용자 상호작용 등으로 인해 `AccountScreenContents`가 부모 위젯에 의해 재빌드될 수 있습니다. 부모 위젯의 상태 변화가 하위 위젯의 재빌드를 유발하는 것입니다.

#### GoRouter의 페이지 빌더 동작?

`GoRoute`에서 사용하는 `CustomTransitionPage` 또는 `GoRouter`의 페이지 빌더 로직에 의해 `AccountScreen`이 재생성되고, 결과적으로 `AccountScreenContents`가 재빌드될 수 있습니다. 라우팅 로직이 페이지의 빌드 타이밍에 영향을 줄 수 있습니다.

### 해결

## The following ImageCodecException was thrown resolving an image codec

### 문제

```log
══╡ EXCEPTION CAUGHT BY IMAGE RESOURCE SERVICE ╞════════════════════════════════════════════════════
The following ImageCodecException was thrown resolving an image codec:
Failed to detect image file format using the file header.
File header was [0x3c 0x21 0x44 0x4f 0x43 0x54 0x59 0x50 0x45 0x20].
Image source: encoded image bytes
When the exception was thrown, this was the stack:
dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:3  throw_
lib/_engine/engine/canvaskit/image_web_codecs.dart 37:7                      create
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54           runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5           _async
lib/_engine/engine/canvaskit/image_web_codecs.dart 22:46                     create
lib/_engine/engine/canvaskit/image.dart 13:34                                skiaInstantiateImageCodec
lib/_engine/engine/canvaskit/renderer.dart 198:15                            instantiateImageCodec
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54           runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5           _async
lib/_engine/engine/canvaskit/renderer.dart 193:41                            instantiateImageCodec
lib/ui/painting.dart 493:28                                                  instantiateImageCodecWithSize
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 84:54           runBody
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 127:5           _async
lib/ui/painting.dart 488:44                                                  instantiateImageCodecWithSize
packages/flutter/src/painting/binding.dart 141:15                            instantiateImageCodecWithSize
packages/flutter/src/painting/_network_image_web.dart 164:20                 _loadAsync
dart-sdk/lib/_internal/js_dev_runtime/patch/async_patch.dart 45:50           <fn>
```

### 원인

- 웹 이미지 피커의 `XFile`이 제대로 처리가 안 됐다

### 해결

```dart
class WebImagePickerService implements ImagePickerService {
  @override
  ImagePicker picker = ImagePicker();

  // imagePath: blob:http://localhost:57167/6460f49e-46cb-4501-9afd-a130cf69b5a7
  @override
  Widget getImageWidget(String imagePath) => Image.network(imagePath);

  @override
  Future<XFile?> pickImage() async =>
      await picker.pickImage(source: ImageSource.gallery);
}
```

## Plugin [id: 'com.android.application', version: '7.2.0', apply: false] was not found in any of the following sources

### 문제

```log
Launching lib/main.dart on sdk gphone64 arm64 in debug mode...
main.dart:1

FAILURE: Build failed with an exception.

* Where:
Build file '/Users/rody/VscodeProjects/dilectio/dilectio_app/android/build.gradle' line: 18

* What went wrong:
Plugin [id: 'com.android.application', version: '7.2.0', apply: false] was not found in any of the following sources:

- Gradle Core Plugins (plugin is not in 'org.gradle' namespace)
- Plugin Repositories (could not resolve plugin artifact 'com.android.application:com.android.application.gradle.plugin:7.2.0')
  Searched in the following repositories:
    Gradle Central Plugin Repository

* Try:
> Run with --stacktrace option to get the stack trace.
> Run with --info or --debug option to get more log output.
> Run with --scan to get full insights.

* Get more help at https://help.gradle.org

BUILD FAILED in 737ms
Exception: Gradle task assembleDebug failed with exit code 1
Exited (1)
```

### 원인

- 문제는 plugins 섹션에서 발생. Flutter 프로젝트의 Android 빌드 구성에서 `com.android.application` 플러그인을 선언하는 방식에 문제
- `com.android.application` 플러그인은 일반적으로 `build.gradle`의 `plugins` 섹션에 `id 'com.android.application'` 형태로 선언되며, 별도의 버전 지정은 필요 없다.
- 플러그인 버전은 `buildscript` 섹션에서 `classpath`를 통해 지정된 `com.android.tools.build:gradle`의 버전에 의해 결정된다

### 해결

루트 프로젝트 수준에서는 `plugins`을 제거한다

```diff
-plugins {
-    id 'com.android.application' version '7.2.0' apply false
-    id 'com.google.gms.google-services' version '4.4.0' apply false
-}
```

`com.android.application` 플러그인은 각 모듈의 `build.gradle` 파일에서 적용한다.
예를 들어, `android/app/build.gradle` 파일에는 다음과 같이 플러그인을 적용

```diff
+plugins {
+    id 'com.android.application'
+    id 'kotlin-android'
+}
+
+android {
+    // 안드로이드 설정
+}
+
+dependencies {
+    // 종속성
+}
```

- `com.android.application` 플러그인을 루트 수준의 `build.gradle` 파일에 직접 버전을 지정하여 선언하는 방식은 Flutter의 표준 구성 다르다.
- Flutter 프로젝트에서의 일반적인 구성은
    1. `com.android.application` 플러그인
        - Flutter 프로젝트에서 이 플러그인은 보통 **앱 수준**의 `build.gradle` 파일(`android/app/build.gradle`)에서 선언된다
        - 루트 수준에서는 버전을 명시적으로 지정하지 않는다
        - `com.android.application` 플러그인 버전은 안드로이드 그래들 플러그인(`com.android.tools.build:gradle`)의 버전에 의해 결정
        - 일반적으로 앱 수준에서 적용되며, 루트 수준에서 `apply false`와 함께 선언하는 것은 일반적인 Flutter 프로젝트 구성이 아니다
    2. `com.google.gms.google-services` 플러그인:
        - 이 플러그인은 Firebase와 같은 Google 서비스를 사용할 때 필요하다
        - 일반적으로 루트 수준의 `build.gradle` 파일에서 종속성을 추가(`classpath` 키워드 사용)
        - 앱 수준의 `build.gradle` 파일에서 실제로 적용(`apply plugin` 또는 `plugins` 블록 사용)
- 좋은 프랙티스?
    - 루트 수준의 `build.gradle`:
        - `buildscript` 블록에서 안드로이드 그래들 플러그인과 필요한 다른 플러그인의 종속성을 선언
        - `allprojects` 블록에서 리포지토리를 정의
    - 앱 수준의 `build.gradle`:
        - `plugins` 또는 `apply plugin`을 사용하여 필요한 플러그인을 적용
        - 앱의 SDK 버전, 종속성, 빌드 설정 등을 정의

## Dependency 'androidx.activity:activity:1.8.0' requires libraries and applications that depend on it to compile against version 34 or later of the Android APIs

### 문제

```log
Launching lib/main.dart on sdk gphone64 arm64 in debug mode...

FAILURE: Build failed with an exception.

* What went wrong:
Execution failed for task ':app:checkDebugAarMetadata'.
> A failure occurred while executing com.android.build.gradle.internal.tasks.CheckAarMetadataWorkAction
   > An issue was found when checking AAR metadata:

       1.  Dependency 'androidx.activity:activity:1.8.0' requires libraries and applications that
           depend on it to compile against version 34 or later of the
           Android APIs.

           :app is currently compiled against android-33.

           Also, the maximum recommended compile SDK version for Android Gradle
           plugin 7.3.0 is 33.

           Recommended action: Update this project's version of the Android Gradle
           plugin to one that supports 34, then update this project to use
           compileSdkVerion of at least 34.

           Note that updating a library or application's compileSdkVersion (which
           allows newer APIs to be used) can be done separately from updating
           targetSdkVersion (which opts the app in to new runtime behavior) and
           minSdkVersion (which determines which devices the app can be installed
           on).

* Try:
> Run with --stacktrace option to get the stack trace.
> Run with --info or --debug option to get more log output.
> Run with --scan to get full insights.

* Get more help at https://help.gradle.org

BUILD FAILED in 4s
Exception: Gradle task assembleDebug failed with exit code 1
Exited (1)

```

### 원인

너무 높은 버전의 material 라이브러리 버전을 사용했기 때문

```gradle
dependencies {
  // Import the Firebase BoM
  implementation platform('com.google.firebase:firebase-bom:32.6.0')
  implementation 'com.google.android.material:material:1.10.0'
                                                       ^^^^^^
                                                       버전이 너무 높으면 activity 관련 에러 발생
```

### 해결

- `1.8.0` 정도로 낮춤

## Querying the mapped value of provider(java.util.Set) before task ':app:processDebugGoogleServices' has completed is not supported

### 문제

```log
Launching lib/main.dart on sdk gphone64 arm64 in debug mode...
main.dart:1

FAILURE: Build failed with an exception.

* What went wrong:
Execution failed for task ':app:mapDebugSourceSetPaths'.
> Error while evaluating property 'extraGeneratedResDir' of task ':app:mapDebugSourceSetPaths'
   > Failed to calculate the value of task ':app:mapDebugSourceSetPaths' property 'extraGeneratedResDir'.
      > Querying the mapped value of provider(java.util.Set) before task ':app:processDebugGoogleServices' has completed is not supported

* Try:
> Run with --stacktrace option to get the stack trace.
> Run with --info or --debug option to get more log output.
> Run with --scan to get full insights.

* Get more help at https://help.gradle.org

BUILD FAILED in 4s
Exception: Gradle task assembleDebug failed with exit code 1
Exited (1)
```

### 원인

[스택오버플로우 답변](https://stackoverflow.com/a/73774667)에 따르면, 이 문제는 `com.android.tools.build:gradle` 버전을 `7.3.0`으로 업데이트 한 후에 발생했고, `com.google.gms:google-services` 버전을 `4.3.14`로 업데이트 한 이후 해결 됐다고 한다

### 해결

[메이븐 리파지토리](https://mvnrepository.com/artifact/com.google.gms/google-services/4.4.0) 확인해보면 최신이 `4.4.0`이어서 이거로 사용

## The number of method references in a `.dex` file cannot exceed 64K

### 문제

```log
RROR:D8: Cannot fit requested classes in a single dex file (# methods: 149621 > 65536)
com.android.builder.dexing.DexArchiveMergerException: Error while merging dex archives:
The number of method references in a .dex file cannot exceed 64K.
Learn how to resolve this issue at https://developer.android.com/tools/building/multidex.html
    at com.android.builder.dexing.D8DexArchiveMerger.getExceptionToRethrow(D8DexArchiveMerger.java:151)
    at com.android.builder.dexing.D8DexArchiveMerger.mergeDexArchives(D8DexArchiveMerger.java:138)
    at com.android.build.gradle.internal.tasks.DexMergingWorkAction.merge(DexMergingTask.kt:859)
    at com.android.build.gradle.internal.tasks.DexMergingWorkAction.run(DexMergingTask.kt:805)
    at com.android.build.gradle.internal.profile.ProfileAwareWorkAction.execute(ProfileAwareWorkAction.kt:74)
    at org.gradle.workers.internal.DefaultWorkerServer.execute(DefaultWorkerServer.java:63)
    at org.gradle.workers.internal.NoIsolationWorkerFactory$1$1.create(NoIsolationWorkerFactory.java:66)
    at org.gradle.workers.internal.NoIsolationWorkerFactory$1$1.create(NoIsolationWorkerFactory.java:62)
    at org.gradle.internal.classloader.ClassLoaderUtils.executeInClassloader(ClassLoaderUtils.java:100)
    at org.gradle.workers.internal.NoIsolationWorkerFactory$1.lambda$execute$0(NoIsolationWorkerFactory.java:62)
    at org.gradle.workers.internal.AbstractWorker$1.call(AbstractWorker.java:44)
    at org.gradle.workers.internal.AbstractWorker$1.call(AbstractWorker.java:41)
    ... 생략 ...
FAILURE: Build completed with 2 failures.

1: Task failed with an exception.
-----------
* What went wrong:
Execution failed for task ':app:compileDebugKotlin'.
> A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
   > Compilation error. See log for more details

* Try:
> Run with --stacktrace option to get the stack trace.
> Run with --info or --debug option to get more log output.
> Run with --scan to get full insights.
==============================================================================

2: Task failed with an exception.
-----------
* What went wrong:
Execution failed for task ':app:mergeExtDexDebug'.
> A failure occurred while executing com.android.build.gradle.internal.tasks.DexMergingTaskDelegate
   > There was a failure while executing work items
      > A failure occurred while executing com.android.build.gradle.internal.tasks.DexMergingWorkAction
         > com.android.builder.dexing.DexArchiveMergerException: Error while merging dex archives:
           The number of method references in a .dex file cannot exceed 64K.
           Learn how to resolve this issue at https://developer.android.com/tools/building/multidex.html

* Try:
> Run with --stacktrace option to get the stack trace.
> Run with --info or --debug option to get more log output.
> Run with --scan to get full insights.
==============================================================================

* Get more help at https://help.gradle.org

BUILD FAILED in 1m 2s
┌─ Flutter Fix ──────────────────────────────────────────────────────────────────────────────┐
│ [!] Your project requires a newer version of the Kotlin Gradle plugin.                     │
│ Find the latest version on https://kotlinlang.org/docs/releases.html#release-details, then │
│ update /Users/rody/VscodeProjects/dilectio/dilectio_app/android/build.gradle:              │
│ ext.kotlin_version = '<latest-version>'                                                    │
└────────────────────────────────────────────────────────────────────────────────────────────┘
```

### 원인

[메서드가 64K개를 초과하는 앱에 관해 멀티덱스 사용 설정](https://developer.android.com/build/multidex?hl=ko)을 참고하라는데...

앱에 API 20 이하 minSdk가 있고 앱과 앱에서 참조하는 라이브러리가 65,536개 메서드를 초과하면 앱이 Android 빌드 아키텍처의 제한에 도달했음을 나타내는 다음과 같은 빌드 오류가 발생

```log
trouble writing output:
Too many field references: 131000; max is 65536.
You may try using --multi-dex option.
```

또는

```log
Conversion to Dalvik format failed:
Unable to execute dex: method ID not in [0, 0xffff]: 65536
```

### 해결

> minSdkVersion이 21 이상으로 설정되면 멀티덱스가 기본적으로 사용 설정되며 멀티덱스 라이브러리가 필요하지 않습니다.

`minSdkVersion` 버전을 21로 지정했다

```gradle
    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId "com.bmd.dilectio"
        // You can update the following values to match your application needs.
        // For more information, see: https://docs.flutter.dev/deployment/android#reviewing-the-gradle-build-configuration.
        // minSdkVersion flutter.minSdkVersion
        minSdkVersion 21
        ^^^^^^^^^^^^^^^^ 버전을 지정
        targetSdkVersion flutter.targetSdkVersion
        versionCode flutterVersionCode.toInteger()
        versionName flutterVersionName
    }
```

## Unhandled Exception: Binding has not yet been initialized

### 문제

```log
E/flutter ( 7024): [ERROR:flutter/runtime/dart_vm_initializer.cc(41)] Unhandled Exception: Binding has not yet been initialized.
E/flutter ( 7024): The "instance" getter on the ServicesBinding binding mixin is only available once that binding has been initialized.
E/flutter ( 7024): Typically, this is done by calling "WidgetsFlutterBinding.ensureInitialized()" or "runApp()" (the latter calls the former). Typically this call is done in the "void main()" method. The "ensureInitialized" method is idempotent; calling it multiple times is not harmful. After calling that method, the "instance" getter will return the binding.
E/flutter ( 7024): In a test, one can call "TestWidgetsFlutterBinding.ensureInitialized()" as the first line in the test's "main()" method to initialize the binding.
E/flutter ( 7024): If ServicesBinding is a custom binding mixin, there must also be a custom binding class, like WidgetsFlutterBinding, but that mixes in the selected binding, and that is the class that must be constructed before using the "instance" getter.
```

### 원인

```dart
void main() async {
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  Logger.root.level = Level.ALL; // defaults to Level.INFO
  Logger.root.onRecord.listen((record) {
    if (kDebugMode) {
      print('${record.level.name}: ${record.time}: ${record.message}');
    }
  });

  final bootstrap = Bootstrap();
  final providerContainer = await bootstrap.createProviderContainer();

  WidgetsFlutterBinding.ensureInitialized();

  runApp(
    bootstrap.rootWidget(
      container: providerContainer,
    ),
  );
}
```

- `Firebase.initializeApp`을 호출하기 전에 `WidgetsFlutterBinding.ensureInitialized()`를 호출하지 않아서 발생
- Flutter 애플리케이션에서 `main` 함수는 앱의 진입점이며, 여기서 필요한 초기화 작업을 수행한다. 따라서 Flutter 프레임워크와 Firebase 등의 외부 플러그인을 사용하기 전에 반드시 필요한 초기화 작업을 완료해야 한다.
- 즉, `Firebase.initializeApp`은 **Flutter의 서비스 바인딩이 초기화된 후에 호출**되어야 한다. 그렇지 않으면, **앱이 필요한 Flutter 서비스**에 접근할 수 없어서 위와 같은 오류가 발생한다.
- `WidgetsFlutterBinding.ensureInitialized()`
    - Flutter 프레임워크가 각종 서비스와 바인딩을 초기화하는 데 필요하다
    - 특히, 비동기 작업을 수행하기 전이나 외부 라이브러리(예: Firebase)를 초기화하기 전에 이 메서드를 호출해야 한다

### 해결

```dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 위치를 가장 상단으로 이동

  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  Logger.root.level = Level.ALL; // defaults to Level.INFO
  Logger.root.onRecord.listen((record) {
    if (kDebugMode) {
      print('${record.level.name}: ${record.time}: ${record.message}');
    }
  });

  final bootstrap = Bootstrap();
  final providerContainer = await bootstrap.createProviderContainer();


  runApp(
    bootstrap.rootWidget(
      container: providerContainer,
    ),
  );
}
```

## PlatformException(sign_in_failed, com.google.android.gms.common.api.ApiException: 10: , null, null)

### 문제

```log
I/flutter (26524): response.message is PlatformException(sign_in_failed, com.google.android.gms.common.api.ApiException: 10: , null, null)
I/flutter (26524): id from _getUser is null
D/OpenGLRenderer(26524): setSurface called with nullptr
D/OpenGLRenderer(26524): setSurface() destroyed EGLSurface
D/OpenGLRenderer(26524): destroyEglSurface
I/ViewRootImpl@fb4bfb8[SignInHubActivity](26524): dispatchDetachedFromWindow
D/InputTransport(26524): Input channel destroyed: '7a75c9d', fd=128
E/flutter (26524): [ERROR:flutter/runtime/dart_vm_initializer.cc(41)] Unhandled Exception: 'package:cloud_firestore/src/collection_reference.dart': Failed assertion: line 116 pos 14: 'path.isNotEmpty': a document path must be a non-empty string
E/flutter (26524): #0      _AssertionError._doThrowNew (dart:core-patch/errors_patch.dart:51:61)
E/flutter (26524): #1      _AssertionError._throwNew (dart:core-patch/errors_patch.dart:40:5)
E/flutter (26524): #2      _JsonCollectionReference.doc (package:cloud_firestore/src/collection_reference.dart:116:14)
E/flutter (26524): #3      FirebaseDatabaseSource.getUser (package:soltalk/data/db/remote/firebase_database_source.dart:63:49)
E/flutter (26524): #4      UserProvider._getUser (package:soltalk/data/provider/user_provider.dart:121:63)
E/flutter (26524): <asynchronous suspension>
E/flutter (26524): #5      UserProvider.loginUserGoogle (package:soltalk/data/provider/user_provider.dart:82:21)
E/flutter (26524): <asynchronous suspension>
E/flutter (26524): #6      LoginScreenState.loginGooglePressed.<anonymous closure> (package:soltalk/ui/screens/login_screen.dart:56:60)
E/flutter (26524): <asynchronous suspension>
E/flutter (26524): #7      LoginScreenState.loginGooglePressed (package:soltalk/ui/screens/login_screen.dart:56:5)
E/flutter (26524): <asynchronous suspension>
E/flutter (26524):
```

### 원인

- [Firebase 프로젝트 설정에서 안드로이드앱 중 SHA-1 추가해주지 않아서](https://kyungsnim.net/200)라고 한다
- 무식하게 기존 앱의 SHA-1 값을 그대로 붙여 넣었는데, 그 탓인가...

### 해결
