# build context

- [build context](#build-context)
    - [BuildContext?](#buildcontext)
    - [Don't use `BuildContext`s across async gaps](#dont-use-buildcontexts-across-async-gaps)

## BuildContext?

`BuildContext`는 Flutter의 UI 요소나 위젯 트리에 대한 참조를 유지하며, 특정 시점의 위젯 트리 상태에 대한 정보를 제공

## Don't use `BuildContext`s across async gaps

- "Don't use `BuildContext`s across async gaps"?
    - 비동기 작업이 수행되는 동안 BuildContext의 참조가 변경될 가능성이 있음을 나타낸다
    - 주로 `BuildContext`를 비동기 작업의 콜백 또는 `Future`에 전달할 때 발생

```dart
final goRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    // .... 생략 ....
    redirect: (context, state) async {
        // top level redirection 경우 특정 라우트에 연결되지 않기 때문에
        // `state.path`, `state.name`이 `null`이 될 수 있다
        log.d("state.fullPath is ${state.fullPath}");
        // public
        if ([
            AppRoute.home.fullPath,
            AppRoute.signUp.fullPath,
            AppRoute.signIn.fullPath,
        ].contains(state.fullPath)) {
            return null;
        }

        final currentAuthInfo = await ref.read(authApplicationProvider.notifier).getCurrentAuthInfo();
        if (currentAuthInfo == null) {
            GoRouter.of(context).goNamed(AppRoute.account.name);
        }

        return null;
    },
```

BuildContext를 비동기 작업을 수행할 때 건너뛰는 것은 위험할 수 있다.
비동기 작업이 완료되는 동안 위젯 트리가 변경될 수 있으며, 이로 인해 이전 BuildContext가 무효화될 수 있다.
이러한 상황은 주로 위젯이 rebuild 되거나 제거되는 경우에 발생합니다.
