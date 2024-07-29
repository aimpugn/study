# state

- [state](#state)
    - [`mounted`](#mounted)
        - [`mounted` 속성의 주요 포인트](#mounted-속성의-주요-포인트)
        - [사용 예시](#사용-예시)
        - [`GoRouter.of(context)` 예제](#gorouterofcontext-예제)
            - [해결 방법](#해결-방법)
            - [코드 예시](#코드-예시)
    - [Stateful 위젯의 컨텐츠인 State에 값 전달하기](#stateful-위젯의-컨텐츠인-state에-값-전달하기)
    - [비동기 상태 업데이트](#비동기-상태-업데이트)
        - [`ConsumerState` 상속하는 화면의 클래스 속성 상태 업데이트](#consumerstate-상속하는-화면의-클래스-속성-상태-업데이트)
        - [`listenManual` 사용](#listenmanual-사용)

## `mounted`

- Flutter에서 `mounted` 속성은 `State` 객체와 관련이 있으며, 해당 `State` 객체가 현재 위젯 트리에 마운트되어 있는지 여부를 나타낸다.
- 즉, `mounted`는 `State` 객체가 현재 활성화되어 있고 사용 가능한 상태인지를 확인하는 데 사용된다.
- `mounted` 속성은 비동기 작업과 상태 업데이트를 수행하는 Flutter 앱에서 매우 중요한 역할을 한다. 이를 통해 위젯의 현재 상태를 안전하게 관리하고, 불필요한 에러를 방지할 수 있다.

### `mounted` 속성의 주요 포인트

1. 위젯의 생명주기: `mounted`는 `StatefulWidget`의 `State` 객체에 속한 속성입니다. 이 속성은 위젯이 위젯 트리에 마운트되어 있는 동안 `true`로 설정되며, 위젯이 트리에서 제거되면 `false`가 됩니다.

2. 안전한 상태 업데이트: `mounted`를 사용하면 비동기 작업 후에 위젯이 여전히 화면에 존재하는지 확인할 수 있습니다. 이는 `setState`를 호출하기 전에 위젯이 여전히 존재하는지 확인하는 데 유용합니다. `setState`를 호출할 때 위젯이 이미 제거되었다면, 에러가 발생할 수 있기 때문입니다.

3. 비동기 작업과의 관계: 비동기 작업(예: `Future`, `async` 함수 등)을 수행하는 동안 사용자가 위젯을 벗어나거나 위젯이 트리에서 제거될 수 있습니다. 이 경우, `mounted` 속성을 검사하여 위젯이 여전히 활성화되어 있는지 확인한 후에 상태 업데이트나 `context` 접근 등을 수행해야 합니다.

### 사용 예시

```dart
Future<void> fetchData() async {
  final response = await fetchSomeData();

  if (mounted) {
    setState(() {
      // 데이터를 상태에 저장
    });
  }
}
```

이 예시에서 `fetchData` 함수는 비동기 작업을 수행한 후, `setState`를 호출하기 전에 `mounted`를 검사합니다. 이렇게 하면 위젯이 화면에서 사라진 상태에서 상태 업데이트를 시도하는 것을 방지할 수 있습니다.

### `GoRouter.of(context)` 예제

```dart
  Future<void> _login() async {
    setState(() => _submitted = true);
    // only submit the form if validation passes
    if (_signInFormKey.currentState!.validate()) {
      await ref
          .read(authApplicationProvider.notifier)
          .loginByEmailAndPassword(email, password);
      final authState = ref.read(authApplicationProvider);
      if (authState.value?.isSuccess == true) {
        final uid = authState.value!.success!.uid;
        await ref.read(userApplicationProvider.notifier).updateLatestUser(uid);
        GoRouter.of(context).go(AppRoute.account.fullPath);
        ^^^^^^^^^^^^^^^^^^^^
        Don't use 'BuildContext's across async gaps.
        Try rewriting the code to not reference the 'BuildContext'
      }
    } else {
      setState(() => _submitted = false);
    }
  }
```

- `async` 작업이 완료되는 동안 사용자가 다른 화면으로 이동하거나 현재 위젯이 트리에서 제거될 수 있기 때문에, 비동기 작업 후에 `BuildContext`를 사용하는 것은 위젯이 더 이상 화면에 표시되지 않을 수 있기 때문에 문제가 될 수 있다
- 이럴 때 `BuildContext`를 사용하려고 하면, 이미 사라진 위젯의 `context`를 참조하게 되어 오류가 발생

#### 해결 방법

1. 상태 확인: 비동기 작업이 완료된 후에 현재 위젯이 여전히 마운트되어 있는지 확인합니다. `mounted` 속성을 사용하여 현재 위젯이 여전히 트리에 있는지 확인할 수 있습니다.

   ```dart
   if (mounted) {
     GoRouter.of(context).go(AppRoute.account.fullPath);
   }
   ```

2. 변수에 `BuildContext` 저장: 비동기 작업을 시작하기 전에 `BuildContext`를 로컬 변수에 저장하고, 비동기 작업 후에 이 변수를 사용한다. 이 방법은 여전히 `mounted` 속성으로 위젯의 상태를 검사해야 한다.

   ```dart
   final localContext = context;

   await ref.read(authApplicationProvider.notifier)
       .loginByEmailAndPassword(email, password);

   if (mounted) {
     GoRouter.of(localContext).go(AppRoute.account.fullPath);
   }
   ```

#### 코드 예시

```dart
Future<void> _login() async {
  setState(() => _submitted = true);

  if (_signInFormKey.currentState!.validate()) {
    final localContext = context; // Context를 로컬 변수에 저장

    await ref
        .read(authApplicationProvider.notifier)
        .loginByEmailAndPassword(email, password);

    if (mounted) { // 위젯이 여전히 마운트되어 있는지 확인
      final authState = ref.read(authApplicationProvider);
      if (authState.value?.isSuccess == true) {
        final uid = authState.value?.success?.uid;
        await ref
            .read(userApplicationProvider.notifier)
            .updateLatestUser(uid);
        
        GoRouter.of(localContext).go(AppRoute.account.fullPath);
      } else {
        // handle login failure
      }
    }
  } else {
    if (mounted) {
      setState(() => _submitted = false);
    }
  }
}
```

이 방법을 사용하면 비동기 작업 후에도 `BuildContext`를 안전하게 사용할 수 있으며, 위젯이 화면에서 사라진 후에는 `context`를 사용하지 않아 오류를 방지할 수 있습니다.

## Stateful 위젯의 컨텐츠인 State에 값 전달하기

- Flutter에서 `StatefulWidget`과 그에 대응하는 `State` 클래스 사이에서 데이터를 전달할 때, 일반적으로 `StatefulWidget`의 생성자를 통해 데이터를 전달하고, 이 데이터를 `State` 클래스에서 접근할 수 있도록 한다.
- `StatefulWidget`의 인스턴스 변수는 해당 위젯의 `State` 객체에서 `widget` 속성을 통해 접근할 수 있다

- 가령 `GreyFilledSpaceSeparatingTagsInput` 위젯에서 `helperText`와 `hintText`를 생성자를 통해 전달받아 `GreyFilledSpaceSeparatingTagsState` 상태에서 사용하는 방법은?
    1. `GreyFilledSpaceSeparatingTagsInput` 생성자에 `helperText`와 `hintText`를 위한 매개변수를 추가
    2. `GreyFilledSpaceSeparatingTagsState`에서 `widget.helperText`와 `widget.hintText`를 통해 이 값을 사용

```dart
class GreyFilledSpaceSeparatingTagsInput extends ConsumerStatefulWidget {
  final String helperText;
  final String hintText;

  const GreyFilledSpaceSeparatingTagsInput({
    super.key,
    this.helperText = 'Enter language...',
    this.hintText = 'Enter tag...',
  });

  @override
  ConsumerState<ConsumerStatefulWidget> createState() =>
      GreyFilledSpaceSeparatingTagsState();
}

class GreyFilledSpaceSeparatingTagsState
    extends ConsumerState<GreyFilledSpaceSeparatingTagsInput> {
  ... 생략 ...
  @override
  Widget build(BuildContext context) {
    return TextFieldTags(
      textfieldTagsController: _controller,
      ... 생략 ...
      inputfieldBuilder: (context, tec, fn, error, onChanged, onSubmitted) {
        return ((context, sc, tags, onTagDelete) {
          return Padding(
            padding: const EdgeInsets.all(10.0),
            child: TextField(
                ... 생략 ...
                helperText: widget.helperText,
                            ^^^^^^ 여기서 widget 속성 사용
                helperStyle: const TextStyle(
                  color: primaryBackgroundColor,
                ),
                hintText: _controller.hasTags ? '' : widget.hintText,
                                                     ^^^^^^
                                                     여기서 widget 속성 사용
              ),
              // ... 나머지 코드
```

## 비동기 상태 업데이트

```dart
  Future<void> _login() async {
    // only submit the form if validation passes
    if (!_signInFormKey.currentState!.validate()) {
      return;
    }
    await ref
        .read(authApplicationProvider.notifier)
        .loginByEmailAndPassword(email, password);
    final authState = ref.read(authApplicationProvider);

    if (authState.value?.isSuccess != true) {
      return;
    }
    final uid = authState.value!.success!.uid;
    await ref.read(userApplicationProvider.notifier).updateLatestUser(uid);
    setState(() => _submitted = true);
  }
```

```dart
  Future<void> updateLatestUser(String uid) async {
    state = const AsyncValue.loading();

    final appUser = await findUserByUID(uid);
    if (appUser == null) {
      state = AsyncValue.error(
        UserApplicationState(
          failure: UserApplicationFailure(
            DomainErrors.notFound(),
          ),
        ),
        StackTrace.current,
      );
      return;
    }

    state = AsyncValue.data(UserApplicationState(
      success: UserApplicationSuccess(appUser),
    ));
  }
```

`loginByEmailAndPassword` 후 애플리케이션 레이어의 user application에서 관리하는 `userState`를 최신으로 업데이트 하기 위해 `await ref.read(userApplicationProvider.notifier).updateLatestUser(uid);`를 호출하면 state가 업데이트 되기 전에 login 함수가 끝나버린다.

그래서 가령 로그인 함수에서 아래처럼 userState를 사용하려고 하면 여전히 `AsyncLoading`인 상태로 나온다. 하지만 `updateLatestUser`의 코드를 보면 `appUser`가 없으면 `AsyncValue.error`가 되고, `appUser`가 있으면 `AsyncValue.data(UserApplicationState`가 호출되어서 `AsyncValue.data`가 되어야 할 거 같은데, 그렇게 동작하지 않는다.

```dart
    final uid = authState.value!.success!.uid;
    await ref.read(userApplicationProvider.notifier).updateLatestUser(uid);
    final userState = ref.read(userApplicationProvider);
    // 여기서 `userState`는 AsyncLoading 상태
```

### `ConsumerState` 상속하는 화면의 클래스 속성 상태 업데이트

### `listenManual` 사용
