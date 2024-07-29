# Riverpod

- [Riverpod](#riverpod)
    - [life cycle](#life-cycle)
        - [`Disposed` / `Uninitialized`](#disposed--uninitialized)
        - [`Creating` -\> `Alive`](#creating---alive)
        - [`Alive`](#alive)
        - [`Alive` -\> `Paused`](#alive---paused)
        - [`Alive` -\> `Disposing`](#alive---disposing)
    - [상태 변경과 build](#상태-변경과-build)
    - [riverpod\_generator](#riverpod_generator)
        - [installation](#installation)
        - [`@riverpod` on class -\> `AutoDisposeAsyncNotifier`](#riverpod-on-class---autodisposeasyncnotifier)
        - [`@Riverpod(keepAlive: true)` on function -\> `Provider`](#riverpodkeepalive-true-on-function---provider)
        - [`@riverpod` on function -\> `AutoDisposeProvider`](#riverpod-on-function---autodisposeprovider)
        - [watch 사용법](#watch-사용법)
    - [classes](#classes)
        - [`AsyncValue`](#asyncvalue)
        - [`WidgetRef`](#widgetref)
            - [`listen`](#listen)
        - [`StateNotifer`](#statenotifer)
        - [`ChangeNotifier`](#changenotifier)
        - [`FutureProvider`](#futureprovider)
        - [`StreamProvider`](#streamprovider)
        - [`ScopedProvider`](#scopedprovider)
    - [`read`와 `watch`](#read와-watch)
    - [`listen`과 `watch`](#listen과-watch)
    - [기타](#기타)

## [life cycle](https://docs-v2.riverpod.dev/docs/concepts/provider_lifecycles)

- `Disposed`
- `Uninitialized`
- `Alive`
- `Paused`

### `Disposed` / `Uninitialized`

> - does not take up any memory since its state is not initialized.
> - just a definition of how to create the provider's state when you need it.

### `Creating` -> `Alive`

> - When an `Uninitialized` provider is read, listened to or watched it's state will be created.
> - If there are any circular dependencies during this creation process Riverpod will throw an error.
> - The provider's state is stored in a `ProviderContainer`. In a Flutter app this container is in a [`ProviderScope`](https://pub.dev/documentation/flutter_riverpod/latest/flutter_riverpod/ProviderScope-class.html) widget.
> - All Flutter applications using Riverpod must contain a `ProviderScope` at the root of their widget tree.
> - This is very similar to how flutter widgets work. You only pay for the definition once, but **can reuse the state in different parts of the tree** as needed.

### `Alive`

> - When your provider is `Alive`, changes to its state will cause **dependent providers and/or the dependent UI to rebuild**.
> - you can `watch` other providers to **have the provider recreate itself** whenever one of it's dependencies changes.
> - If you need to have some long-lived state that depends on other state you can use Ref's `listen` method to **subscribe for changes on another provider without causing a rebuild of the provider**.
> - If you need to use the state from another provider in a side-effect, you can use Ref's `read` method **to obtain the current state from another provider**.
> - Typically when constructing a `StateNotifier` or `ChangeNotifier` class, you should pass in the `ref` to **allow the `Notifier` to obtain the current value of dependencies as needed**.

### `Alive` -> `Paused`

### `Alive` -> `Disposing`

## 상태 변경과 build

```dart
class DilectioApp extends ConsumerWidget {
  const DilectioApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp.router(
      routerConfig: ref.watch(goRouterProvider),
      debugShowCheckedModeBanner: false,
      restorationScopeId: toString(),
      ... 생략 ...
    );
  }
}
```

- `ref.watch(goRouterProvider)`를 사용하면 `goRouterProvider`의 상태가 변경될 때 위젯이 다시 빌드된다. 그러면 앱 전체를 다시 빌드하게 되는 걸까? 아니라고 한다
    - Flutter의 위젯 트리는 매우 효율적으로 설계되어 있어, 변경된 부분만 다시 빌드. `MaterialApp`을 다시 빌드하더라도, 그 하위에 있는 위젯들은 그대로 유지될 수 있고, 실제로 변경되어야 할 부분만 다시 빌드된다
    - `MaterialApp.router`의 `routerConfig` 속성이 변경되면, 라우터와 관련된 부분은 새로운 상태로 업데이트 될 것이다. 예를 들어, 현재 화면이 새로운 라우트로 변경되어야 한다면 해당 화면은 새로 빌드된다.

## [riverpod_generator](https://pub.dev/packages/riverpod_generator)

> - to offer a different syntax for defining "providers" by relying on code generation.

### installation

```yaml
dependencies:
  # or flutter_riverpod/hooks_riverpod as per https://riverpod.dev/docs/getting_started
  # riverpod: ^2.3.6 # dart only no flutter
  hooks_riverpod: ^2.3.7
  # the annotation package containing @riverpod
  riverpod_annotation: ^2.1.1

dev_dependencies:
  # a tool for running code generators
  build_runner: ^2.4.6
  # the code generator
  riverpod_generator: ^2.2.4
```

### `@riverpod` on class -> `AutoDisposeAsyncNotifier`

```dart
// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_application.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$userApplicationHash() => r'8438d571d278c24f8c2143857f1e4fa78b40a1d1';

/// See also [UserApplication].
@ProviderFor(UserApplication)
final userApplicationProvider =
    AutoDisposeAsyncNotifierProvider<UserApplication, void>.internal(
  UserApplication.new,
  name: r'userApplicationProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$userApplicationHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef _$UserApplication = AutoDisposeAsyncNotifier<void>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member
```

### `@Riverpod(keepAlive: true)` on function -> `Provider`

```dart
// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_application.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$userApplicationHash() => r'cbd3a8d1f3b2c73f468302a0ef14f573e60cfcc3';

/// See also [userApplication].
@ProviderFor(userApplication)
final userApplicationProvider = Provider<UserApplication>.internal(
  userApplication,
  name: r'userApplicationProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$userApplicationHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef UserApplicationRef = ProviderRef<UserApplication>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member
```

### `@riverpod` on function -> `AutoDisposeProvider`

```dart
// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_application.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$userApplicationHash() => r'd4cd32eb232e6b9c97a1530ee728d2ea34e82430';

/// See also [userApplication].
@ProviderFor(userApplication)
final userApplicationProvider = AutoDisposeProvider<UserApplication>.internal(
  userApplication,
  name: r'userApplicationProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$userApplicationHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

typedef UserApplicationRef = AutoDisposeProviderRef<UserApplication>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member
```

### watch 사용법

```dart
@riverpod
class UsersController extends _$UsersController {
  @override
  FutureOr<void> build() {}

  Future<bool> loginByEmail({
    required String email,
    required String password,
  }) async {
    state = const AsyncValue.loading();

    final signInService = ref.watch(signInSerivceProvider);
    state = await AsyncValue.guard(
        () => signInService.byEmailAndPassword(email, password));

    return state.hasError == false;
  }
}
```

```dart
final asyncValueState = ref.watch(usersControllerProvider);
final usersController = ref.watch(usersControllerProvider.notifier);
```

- `usersControllerProvider`를 watch 하면 `AsyncValue<void>` 타입의 값이 리턴된다
- `usersControllerProvider.notifier)`를 watch 하면 `UsersController` 타입의 값이 리턴된다

## classes

### `AsyncValue`

> a class that allows you to perform actions on distinct UI states such as loading, error, or success.

### `WidgetRef`

#### `listen`

```dart
ref.listen<AsyncValue>(
    signInControllerProvider,
    (_, state) => state.showAlertDialogOnError(context),
);
```

- `signInControllerProvider`는 감시하고자 하는 프로바이더로, 이 프로바이더의 상태 변화를 수신
- 그 중에서도 `AsyncValue` 타입 사용하므로 비동기 값의 변화를 감지
- 프로바이더 상태가 변할 때마다 `(_, state) => state.showAlertDialogOnError(context)` 콜백 호출

### `StateNotifer`

- 상태 변경 감지: `StateNotifier`에서는 상태를 직접 변경하면 자동으로 알림이 전달된다
- 상태 관리: 상태는 하나의 불변 객체로 관리되며, `state` 속성을 통해 접근할 수 있다
- 캡슐화: `StateNotifier`는 상태를 외부에서 직접 변경할 수 없으므로 상태의 무결성이 더 잘 보장된다
- 메모리 관리: `StateNotifier`는 리스너 관리가 자동으로 이루어진다
- 플랫폼 중립: `StateNotifier`는 Flutter에 종속되지 않은 독립적인 패키지
- `freezed`와 상성이 좋아 같이 사용. Provider패키지를 만든 [Remi Rousselet](https://twitter.com/remi_rousselet)가 만들었다.
- [flutter state_notifier란?](https://negabaro.github.io/archive/flutter-state_notifier) 참고

```dart
class LoggedInStateNotifier extends StateNotifier<LoggedInState> {
  LoggedInStateNotifier() : super(LoggedInState());

  void setLoggedInInfo(String uid, String email) {
    state.uid = uid;
    state.email = email;
  }
}

final loggedInStateProvider =
    StateNotifierProvider<LoggedInStateNotifier, LoggedInState>((_) {
  return LoggedInStateNotifier();
});


final goRouterProvider = Provider<GoRouter>((ref) {
  final loggedInState = ref.watch(loggedInStateProvider);
  final userRepository = ref.watch(userRepositoryProvider);

  return GoRouter(
    observers: [RouterObserver(loggedInState: loggedInState)],
    //                                        ^^^^^^^^^^^^^ flutter에서 객체는 기본적으로 참조로 전달. 
  )
});

class RouterObserver extends NavigatorObserver {
  final LoggedInState loggedInState;
  //                  ^^^^^^^^^^^^^^
  //                  따라서 처음 `LoggedInStateNotifier`에 초기화된 **원본** `LoggedInState` 객체의 상태가 변경되면
  //                  여기서도 **동일한 객체를 참조**하고 있기 때문에 자동으로 변경된 상태가 반영된다.
  RouterObserver({required this.loggedInState});
}
```

### `ChangeNotifier`

- 상태 변경 감지: `ChangeNotifier`에서 상태를 변경하려면 `notifyListeners()` 메서드를 호출해야 한다
- 상태 관리: 상태는 `ChangeNotifier` 내부의 여러 변수로 관리된다
- 캡슐화: `notifyListeners()`를 임의로 호출할 수 있으므로 상태의 무결성을 보장하기 어렵다
- 메모리 관리: `addListener`와 `removeListener`를 수동으로 관리해야 할 수도 있다
- Flutter 기반: `ChangeNotifier는` Flutter의 foundation 라이브러리에 포함되어 있다

### `FutureProvider`

### `StreamProvider`

### `ScopedProvider`

## `read`와 `watch`

```dart
class _SignInContentsState extends ConsumerState<SignInContents>
    with TextValidator {

  Future<void> _login() async {
    setState(() => _submitted = true);
    // only submit the form if validation passes
    if (_formKey.currentState!.validate()) {
      final controller = ref.read(signInControllerProvider.notifier);
      final signInState = await controller.loginByEmail(
        email: email,
        password: password,
      );

      if (signInState is SignInError) {
        log.warning(signInState);
      }

      final success = signInState is SignInSuccess;

      final currentState = ref.read(logInStateProvider.notifier);
      log.info("when _login success, ${currentState.jwt!.payload}");
    }
    setState(() => _submitted = false);
  }

  @override
  Widget build(BuildContext context) {
    ref.listen<AsyncValue>(
      signInControllerProvider,
      (_, state) => state.showAlertDialogOnError(context),
    );
    final state = ref.watch(signInControllerProvider);
    final currentState = ref.watch(logInStateProvider.notifier);
    if (currentState != null && currentState.jwt != null) {
      log.info("when watch, currentState is ${currentState.jwt!.payload}");
    } else {
      log.info("when watch, currentState is null");
    }

    ... 생략 ...
  }
}
```

위의 로그를 출력해보면, 아래와 같다

```log
[_SignInContentsState] when watch, currentState is null
[UserApplication] currentState is Instance of 'LogInState'
[UserApplication] after is Instance of 'LogInState'
[_SignInContentsState] when _login success, {id: igf3AWU0XCMmXBOVJCl4Ec7kd1i2, email: aimpugn@gmail.com}
[_SignInContentsState] when watch, currentState is {id: igf3AWU0XCMmXBOVJCl4Ec7kd1i2, email: aimpugn@gmail.com}
```

1. `build` 시에 `currentState` 확인
2. 로그인 후 jwt 업데이트 전/후
3. `_login` 성공 후 내부 로그
4. 상태가 변경되면 `watch`가 이를 감지하고 위젯을 다시 `build`

## `listen`과 `watch`

- `ref.listen`
    - 상태 변화를 감지하고, 이에 따라 부가 작업을 수행
    - 예를 들어 로그인이 성공적으로 이루어진 경우 홈 화면으로 이동하는 작업을 수행할 수 있음
    - listen은 주로 화면을 다시 그리지 않고 부가적인 작업을 수행할 때 사용
- `ref.watch`:
    - 이 메서드는 상태 변화를 감지하고, 이에 따라 위젯을 다시 그린다
    - 주로 UI를 업데이트하는 데 사용됩니다.

> 왜 두 번 호출하는가?
>
> `ref.listen`과 `ref.watch`는 각각 다른 목적으로 설계되었다.
> `ref.listen`은 상태 변화에 따른 부가 작업을 처리하고, `ref.watch`는 상태 변화에 따라 UI를 업데이트.
> 따라서 두 메서드를 같이 사용하는 것은 일반적인 패턴이며, 각각의 목적에 맞게 사용되어야 한다.

```dart
  Widget build(BuildContext context) {
    ref.listen<AsyncValue>(
      signInControllerProvider,
      (_, state) => state.when(data: (data) {
        log.info("state is $state");
        log.info("data is $data");
      }, error: (error, stackTrace) {
        log.info("error is $error");
        log.info("stackTrace is $stackTrace");
      }, loading: () {
        log.info("when loading");
      }),
    )

    final state = ref.watch(signInControllerProvider);
    
    ... 생략 ...
  }
```

이를 찍어 보면, 로그인 성공 시 아래와 같이 나온다

```log
[_SignInContentsState] when loading
[UserApplication] currentState is Instance of 'LogInState'
[UserApplication] after is Instance of 'LogInState'
[_SignInContentsState] state is AsyncData<SignInState>(value: Instance of 'SignInSuccess')
[_SignInContentsState] data is Instance of 'SignInSuccess'
```

## 기타

- [Riverpod Data Caching and Providers Lifecycle: Full Guide](https://codewithandrea.com/articles/flutter-riverpod-data-caching-providers-lifecycle/)
- [Flutter 인기 아키텍처 라이브러리 3종 비교 분석 - GetX vs BLoC vs Provider](https://engineering.linecorp.com/ko/blog/flutter-architecture-getx-bloc-provider)
- [Riverpod vs Bloc: Comparison of basic features](https://medium.com/snapp-x/riverpod-vs-bloc-comparison-of-basic-features-71f6d8732d82)
