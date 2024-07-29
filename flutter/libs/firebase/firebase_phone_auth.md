# Firebase Phone Auth

## [전화 인증](https://firebase.google.com/docs/auth/flutter/phone-auth?hl=ko)

- 전화 인증을 사용하면 사용자가 전화를 인증자로 사용하여 **Firebase에 로그인**할 수 있다
- 제공된 전화번호를 사용하여 고유 코드가 포함된 SMS 메시지가 사용자에게 전송되고, 코드가 승인되면 사용자가 Firebase에 로그인할 수 있다

> NOTE:
>
> 악용을 방지하기 위해 새 프로젝트에 현재 50의 SMS 일일 할당량이 적용됩니다. 할당량을 상향 조정하려면 프로젝트에 결제 계정을 추가하세요.

- 플랫폼별로 설정이 다르다
    - [Android](https://firebase.google.com/docs/auth/android/phone-auth?hl=ko&authuser=0)
        - Firebase Console에서 앱의 SHA-1 해시를 설정
        - 앱의 SHA-1 해시를 찾는 방법은 [클라이언트 인증](https://developers.google.com/android/guides/client-auth?authuser=0&hl=ko)을 참조
    - [Web](https://firebase.google.com/docs/auth/web/phone-auth?hl=ko&authuser=0)
        - `Firebase Console > Authentication > Settings > 도메인 > 승인된 도메인`에서 도메인 설정 필요
- [flutter 전화 인증](https://firebase.google.com/docs/auth/flutter/phone-auth?hl=ko&authuser=0)

## flutter 전화 인증

- 네이티브(예: Android 및 iOS) 플랫폼은 전화번호 검증에 서로 다른 기능을 제공
- 각 플랫폼에 독점적으로 사용할 수 있는 두 가지 방법이 있다
    - 네이티브 플랫폼: `verifyPhoneNumber`.
    - 웹 플랫폼: `signInWithPhoneNumber`.

### [네이티브: `verifyPhoneNumber`](https://firebase.google.com/docs/auth/flutter/phone-auth?hl=ko&authuser=0#native_verifyphonenumber)

- 사용자의 전화번호가 먼저 인증된 후
    - 사용자가 로그인
    - 또는 계정을 `PhoneAuthCredential`에 연결
- 따라서 먼저 사용자에게 전화번호를 요청해야 한다. 전화번호가 제공되면 `verifyPhoneNumber()` 메서드를 호출

```dart
await FirebaseAuth.instance.verifyPhoneNumber(
  phoneNumber: '+44 7123 123 456',
  verificationCompleted: (PhoneAuthCredential credential) {},
  verificationFailed: (FirebaseAuthException e) {},
  codeSent: (String verificationId, int? resendToken) {},
  codeAutoRetrievalTimeout: (String verificationId) {},
);
```

- 4개의 콜백이 있다
    - `verificationCompleted`: Android 기기의 SMS 코드 자동 처리.
    - `verificationFailed`: 잘못된 전화번호나 SMS 할당량 초과 여부 등의 실패 이벤트를 처리
    - `codeSent`: Firebase에서 기기로 코드가 전송된 경우를 처리하며 사용자에게 코드를 입력하라는 메시지를 표시하는 데 사용
    - `codeAutoRetrievalTimeout`: 자동 SMS 코드 처리에 실패한 경우 시간 초과를 처리

### [웹: signInWithPhoneNumber](https://firebase.google.com/docs/auth/flutter/phone-auth?hl=ko&authuser=0#web_signinwithphonenumber)

- 웹 플랫폼에서 사용자는 **제공된 전화번호로 전송된 SMS 코드를 입력하는 방식**으로 휴대전화에 액세스할 수 있는지 확인하여 로그인 가능
- 보안 및 스팸 방지를 강화하기 위해 사용자는 Google reCAPTCHA 위젯을 완료하여 실제 사람임을 증명해야 하고, 확인이 완료되면 SMS 코드가 전송된다
- Flutter용 Firebase 인증 SDK는 기본적으로 reCAPTCHA 위젯을 관리하지만, 필요한 경우 표시 및 구성 방법을 제어할 수 있다
- 시작하려면 전화번호로 `signInWithPhoneNumber` 메서드를 호출
    1. 메서드를 호출하면 먼저 reCAPTCHA 위젯이 표시되고, 사용자는 SMS 코드가 전송되기 전에 테스트를 완료해야 한다
    2. 완료되면 확인된 `ConfirmationResult` 응답의 confirm 메서드에 SMS 코드를 제공하여 사용자를 로그인 처리할 수 있다

```dart
FirebaseAuth auth = FirebaseAuth.instance;

// Wait for the user to complete the reCAPTCHA & for an SMS code to be sent.
ConfirmationResult confirmationResult = await auth.signInWithPhoneNumber('+44 7123 123 456');

// SMS 코드 확인
UserCredential userCredential = await confirmationResult.confirm('123456');
```

- 다른 로그인 흐름과 마찬가지로, 성공적으로 로그인하면 애플리케이션 전체에서 구독한 인증 상태 리스너가 트리거된다
