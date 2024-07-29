# go_router

- [go\_router](#go_router)
    - [Build Complex Navigation Stack with Flutter](#build-complex-navigation-stack-with-flutter)
    - [Flutter Navigation with GoRouter: Go vs Push](#flutter-navigation-with-gorouter-go-vs-push)
    - [경로 이동 방식의 차이점](#경로-이동-방식의-차이점)
        - [`GoRouter.of(context).push("/${AppRoute.account.path}")`](#gorouterofcontextpushapprouteaccountpath)
        - [`GoRouter.of(context).go("/${AppRoute.account.path}")`](#gorouterofcontextgoapprouteaccountpath)
        - [`context.goNamed(AppRoute.account.path)`](#contextgonamedapprouteaccountpath)
    - [Provider read vs watch](#provider-read-vs-watch)
        - [read](#read)
        - [watch](#watch)
    - [classes](#classes)
        - [`GoRouter`](#gorouter)
        - [`GoRoute`](#goroute)
    - [기타](#기타)

## [Build Complex Navigation Stack with Flutter](https://auberginesolutions.com/blog/build-navigation-stack-with-flutter/)

## [Flutter Navigation with GoRouter: Go vs Push](https://codewithandrea.com/articles/flutter-navigation-gorouter-go-vs-push/)

## 경로 이동 방식의 차이점

```dart
enum AppRoute {
  home(path: '/'),
  signIn(path: 'signIn'),
  signUp(path: 'signUp'),
  account(path: 'account');

  const AppRoute({required this.path});

  final String path;
}
```

경로를 이런 enum으로 정의했습니다.

### `GoRouter.of(context).push("/${AppRoute.account.path}")`

- 새로운 화면을 현재 네비게이션 스택의 맨 위에 추가
- 사용자가 Android의 뒤로 가기 버튼이나 앱의 뒤로 가기 버튼을 누르면 이전 화면으로 돌아갈 수 있다

### `GoRouter.of(context).go("/${AppRoute.account.path}")`

- 네비게이션 스택을 완전히 교체하여 새로운 화면으로 이동
    - `go` 메서드는 현재 네비게이션 스택의 최상단 경로를 새로운 경로로 교체한다
    - 이는 기존의 경로를 완전히 대체하는 것을 의미
- 뒤로 가기 버튼을 눌렀을 때 이전 화면으로 돌아가지 않는다
    - 일반적으로 `go` 메서드를 사용하면 뒤로 가기 버튼을 눌렀을 때 이전 화면으로 돌아가지 않는다
    - 그러나 이는 네비게이션 스택과 관련된 상태에 따라 달라질 수 있다

### `context.goNamed(AppRoute.account.path)`

- 특별한 라이브러리나 패키지 없이 Flutter의 기본 Navigator를 사용하여 이름을 기반으로 한 라우팅을 수행

## Provider read vs watch

### read

- `build` 시에 한번 읽은 `goRouterProvider`를 계속 사용하게 된다

    ```dart
    class DilectioApp extends ConsumerWidget {
        const DilectioApp({super.key});

        @override
        Widget build(BuildContext context, WidgetRef ref) {
            return MaterialApp.router(
            routerConfig: ref.read(goRouterProvider),
            ... 생략 ...
            );
        }
    }
    ```

    ```dart
    final goRouterProvider = Provider<GoRouter>((ref) {
        final loggedInState = ref.read(loggedInStateProvider);
        log.info([
            loggedInState,
            loggedInState.uid,
        ]);
    ```

- `watch`를 사용해도 최신 로그인 상태를 다시 읽을 수 없다

    ```dart
    final goRouterProvider = Provider<GoRouter>((ref) {
        final loggedInState = ref.watch(loggedInStateProvider);
        log.info([
            loggedInState,
            loggedInState.uid,
        ]);
    ```

    ```log
    Restarted application in 99ms.
    The platformViewRegistry getter is deprecated and will be removed in a future release. Please import it from `dart:ui_web` instead.
    INFO: 2023-09-28 21:42:16.811: build is called
    INFO: 2023-09-28 21:42:16.812: [Instance of 'LoggedInState', null]
    INFO: 2023-09-28 21:42:16.813: Full paths for routes: ... 생략 ...
    INFO: 2023-09-28 21:42:16.814: setting initial location /
    INFO: 2023-09-28 21:42:16.823: Using MaterialApp configuration
    INFO: 2023-09-28 21:42:16.824: loggedInState is null, null
    <로그인 후>
    INFO: 2023-09-28 21:42:59.615: currentState is Instance of 'LoggedInStateNotifier'
    INFO: 2023-09-28 21:42:59.616: after is Instance of 'LoggedInStateNotifier'
    INFO: 2023-09-28 21:42:59.618: getting location for name: "account"
    INFO: 2023-09-28 21:42:59.618: going to /account
    INFO: 2023-09-28 21:42:59.623: loggedInState is null, null
    ```

### watch

- `goRouterProvider`를 `watch` 하더라도 `loggedInStateProvider`를 `read` 하기 때문에 결국 현재 업데이트 된 상태를 읽을 수 없다

    ```dart
    class DilectioApp extends ConsumerWidget {
        const DilectioApp({super.key});

        @override
        Widget build(BuildContext context, WidgetRef ref) {
            return MaterialApp.router(
            routerConfig: ref.watch(goRouterProvider),
            ... 생략 ...
            );
        }
    }
    ```

    ```dart
    final goRouterProvider = Provider<GoRouter>((ref) {
        final loggedInState = ref.read(loggedInStateProvider);
        log.info([
            loggedInState,
            loggedInState.uid,
        ]);
    ```

    - `watch`를 해야 현재 상태 값을 다시 읽어 들일 수 있다. 근데 이때 `DilectioApp`의 `build`가 다시 호출된다. 즉 전체 앱이 다시 빌드 된다

    ```dart
    final goRouterProvider = Provider<GoRouter>((ref) {
    final loggedInState = ref.read(loggedInStateProvider);
    log.info([
        loggedInState,
        loggedInState.uid,
    ]);
    ```

    ```log
    Restarted application in 101ms.
    The platformViewRegistry getter is deprecated and will be removed in a future release. Please import it from `dart:ui_web` instead.
    INFO: 2023-09-28 21:45:11.024: build is called
    INFO: 2023-09-28 21:45:11.025: [Instance of 'LoggedInState', null]
    INFO: 2023-09-28 21:45:11.026: Full paths for routes: ... 생략 ...
    INFO: 2023-09-28 21:45:11.027: setting initial location /
    INFO: 2023-09-28 21:45:11.037: Using MaterialApp configuration
    INFO: 2023-09-28 21:45:11.039: loggedInState is null, null

    <로그인 후>
    INFO: 2023-09-28 21:45:37.854: currentState is Instance of 'LoggedInStateNotifier'
    INFO: 2023-09-28 21:45:37.855: after is Instance of 'LoggedInStateNotifier'
    INFO: 2023-09-28 21:45:37.857: getting location for name: "account"
    INFO: 2023-09-28 21:45:37.857: going to /account
    INFO: 2023-09-28 21:45:37.859: [Instance of 'LoggedInState', igf3AWU0XCMmXBOVJCl4Ec7kd1i2]
    INFO: 2023-09-28 21:45:37.860: Full paths for routes: ... 생략 ...
    INFO: 2023-09-28 21:45:37.861: setting initial location /
    INFO: 2023-09-28 21:45:37.861: build is called
    INFO: 2023-09-28 21:45:37.871: Using MaterialApp configuration
    INFO: 2023-09-28 21:45:37.876: loggedInState is igf3AWU0XCMmXBOVJCl4Ec7kd1i2, aimpugn@gmail.com
    ```

## classes

### `GoRouter`

```dart
final goRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: AppRoute.home.path,
    routes: [
      GoRoute(
        path: AppRoute.home.path,
        name: AppRoute.home.name,
        builder: (context, state) => const SignInScreen(),
      )
    ],
    errorBuilder: (context, state) => const NotFoundScreen(),
    redirect: (context, state) {
      // top level redirection 경우 특정 라우트에 연결되지 않기 때문에
      // `state.path`, `state.name`이 `null`이 될 수 있다
      if ([
        AppRoute.home.fullPath,
        AppRoute.signUp.fullPath,
        AppRoute.signIn.fullPath,
      ].contains(state.fullPath)) {
        return null;
      }

      final loggedInState = ref.read(loggedInStateProvider);

      if (!loggedInState.isLoggedin) {
        return AppRoute.signIn.fullPath;
      }

      return null;
    },
  );
});
```

### `GoRoute`

- `builder`:
    - responsible for building the Widget to display on screen
    - 단순히 `Widget`을 반환하며, 페이지 전환 애니메이션을 제어할 수 없다
- `pageBuilder`:
    - customize the transition animation when that route becomes active
    - 페이지 전환 애니메이션을 제어할 수 있다
    - `Page` 객체를 반환하고, `Page` 객체는 Navigator에 의해 관리되며, 페이지 전환 애니메이션 등을 제어할 수 있다

## 기타

- [Flutter Authentication Flow with Go Router and Provider](https://blog.ishangavidusha.com/flutter-authentication-flow-with-go-router-and-provider)
