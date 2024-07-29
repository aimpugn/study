# AsyncValue

- [AsyncValue](#asyncvalue)
    - [`AsyncValue`의 상태 변화 이해](#asyncvalue의-상태-변화-이해)
    - [`AsyncValue` 타입 사용법 배우기](#asyncvalue-타입-사용법-배우기)
        - [`AsyncValue.when` 경우 로딩만 로깅되는 경우](#asyncvaluewhen-경우-로딩만-로깅되는-경우)
            - [문제](#문제)
            - [원인](#원인)
            - [해결](#해결)

## `AsyncValue`의 상태 변화 이해

`AsyncValue`는 초기 상태가 `loading`이며, 비동기 작업의 결과에 따라 `data` 또는 `error` 상태로 전환됩니다. `read` 메서드는 상태 변화를 구독하지 않으므로, `startPhoneNumberVerification`이 호출될 때 `phoneAuthFactoryProvider`의 상태가 여전히 `loading`일 가능성이 높습니다.

## `AsyncValue` 타입 사용법 배우기

### `AsyncValue.when` 경우 로딩만 로깅되는 경우

#### 문제

아래 코드의 경우 `AsyncValue` 타입의 결과 `phoneAuth`에서 `when`을 사용할 때 로딩에 대한 로그만 남는다

Presentation Layer

```dart
void _startVerifyPhoneNumber() async {
  final phonNumber = _state.phoneNumber;
  if (phonNumber != null) {
    final phoneNumber = "${phonNumber.countryCode}${phonNumber.number}";
    final verificationId = await ref
        .read(authApplicationProvider.notifier)
        ^^^^^^ 이 부분이 문제가 된다
        .startPhoneNumberVerification(phoneNumber);
    _updateState((signUpScreenState) async {
      signUpScreenState.smsCodeSent = verificationId.isNotEmpty;
      signUpScreenState.verificationId = verificationId;
    });
  }
}
```

Application Layer

```dart
Future<String> startPhoneNumberVerification(String phoneNumber) async {
  final phoneAuth = ref.read(phoneAuthFactoryProvider);
                       ^^^^^^ 이 부분이 문제가 된다
  return phoneAuth.when(
    loading: () {
      log.i("startPhoneNumberVerification when loading");
      return "";
    }, // 로딩 중일 때 처리
    error: (err, stack) {
      log.e(err, stackTrace: stack);
      return "";
    }, // 에러 상태일 때 처리
    data: (phoneAuth) async {
      log.i("phonAuth is $phoneAuth");
      if (phoneAuth == null) {
        throw UnsupportedError("Unsupported Platform");
      }
      final either = await phoneAuth
          .startPhoneNumberVerification(
            phoneNumber,
            (verificationId, resendToken) {},
            log.e,
          )
          .run();
      return either.getOrElse((_) => "");
    },
  );
}
```

Infrastructure Layer

```dart
@Riverpod(keepAlive: true)
class PhoneAuthFactory extends _$PhoneAuthFactory {
  final log = Logger();

  @override
  Future<PhoneAuthRepositoryInterface?> build() async {
  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 상태가 비동기로 업데이트 되며, AsyncValue 타입으로 리턴된다
    if (state.value != null) {
      return state.value;
    }
    if (kIsWeb) {
      log.i("when web");
      return PhoneAuthWeb(auth: FirebaseAuth.instance);
    } else if (Platform.isAndroid) {
      log.i("when android");
      return PhoneAuthAndroid(auth: FirebaseAuth.instance);
    } else if (Platform.isIOS) {
      log.i("when ios");
      return PhoneAuthIOS(auth: FirebaseAuth.instance);
    }

    return null;
  }
}
```

#### 원인

> `read`는 **현재 상태의 스냅샷**을 반환하는 반면, `watch`는 상태의 **변화를 지속적으로 관찰하고 반응**한다. 따라서 비동기 작업을 다룰 때는 상태 변화에 대응하기 위해 `watch` 사용이 더 적합하다.

`PhoneAuthFactory`에서 `Future<PhoneAuthRepositoryInterface?> build() async {` 보면 알 수 있듯이, 상태를 비동기로 업데이트 한다.
따라서 **현재 상태 읽기**를 하는 `read`를 사용하면 나중에 비동기로 업데이트 된 상태의 변화를 감지하지 못한다.

그렇다고 Application 레이어의 `startPhoneNumberVerification`에서만 `watch`를 사용하면 안 된다. 프리젠테이션에서도 이 상태의 변화를 알아야 하기 때문에 Presentation 레이어의 `_startVerifyPhoneNumber`에서도 `read`가 아닌 `watch`를 사용해야 한다

#### 해결

Presentation Layer

```dart
void _startVerifyPhoneNumber() async {
  final phonNumber = _state.phoneNumber;
  if (phonNumber != null) {
    final phoneNumber = "${phonNumber.countryCode}${phonNumber.number}";
    final verificationId = await ref
        .watch(authApplicationProvider.notifier)
        ^^^^^^ read -> watch
        .startPhoneNumberVerification(phoneNumber);
```

Application Layer

```dart
Future<String> startPhoneNumberVerification(String phoneNumber) async {
  final phoneAuth = ref.watch(phoneAuthFactoryProvider);
                       ^^^^^^ read -> watch
}
```

추가적으로 다음과 같은 사항들도 나중에 고려할 수 있다

- 상태 변화에 따른 UI 처리
    - `watch`를 사용하면 상태 변화에 따라 UI가 업데이트될 수 있으므로, 이를 위해 `AsyncValue`의 상태에 따라 다른 UI를 표시하는 로직을 추가해야 할 수도 있다.
    - 예를 들어, `loading` 상태일 때 로딩 인디케이터를 보여주는 등의 처리가 필요
- 비동기 작업의 완료 확인
    - `watch`를 사용할 경우, 비동기 작업이 완료되고 `data` 상태에 도달했을 때만 특정 액션을 수행하도록 구현해야 한다.
    - `data` 상태가 될 때까지 기다리는 로직을 추가할 필요가 있다
    - BUT 현재 모든 상태를 애플리케이션 레이어에서 관리하고 프리젠테이션 레이어는 결과만 받도록 구현되어 있으므로, 굳이 `_startVerifyPhoneNumber`에서 로딩시 처리를 할 필요는 없다
- 에러 핸들링
    - 비동기 작업 중 에러가 발생할 경우, 이를 적절히 처리하는 로직이 필요하다
    - 에러 상태를 UI에 반영하거나 사용자에게 알리는 방법을 고려해야 한다
