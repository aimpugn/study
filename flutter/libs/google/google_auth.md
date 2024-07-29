# Google Auth

## web

### 메타 태그 추가

- 웹의 경우 메타 태그로 클라이언트 아이디 추가 필요

```html
  <meta name="google-signin-client_id"
    content="12345678910-abcdefghijklmnopqrstu.apps.googleusercontent.com" />
```

### 웹 SDK 구성

### signin 메서드 사용은 권장되지 않음

- 이 프로세스는 ID 토큰을 반환하지 않으므로 권장하지 않는다

```dart
GoogleSignIn googleSignIn = GoogleSignIn(
    clientId: const String.fromEnvironment('GOOGLE_CLIENT_ID'),
);

await googleSignIn.signIn();
```

### 사용자의 ID 토큰에 액세스할 수 없음 로그

- 이 메시지는 사용자의 ID 토큰에 액세스할 수 없음을 알려주는 역할을 하지만 안전하게 무시해도 된다

```log
The OAuth token was not passed to gapi.client, since the gapi.client library is not loaded in your page.
```

### SHA-1 키 등록

```shell
keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
별칭 이름: androiddebugkey
생성 날짜: 2022. 9. 14.
항목 유형: PrivateKeyEntry
인증서 체인 길이: 1
인증서[1]:
소유자: C=US, O=Android, CN=Android Debug
발행자: C=US, O=Android, CN=Android Debug
일련 번호: 1
적합한 시작 날짜: Wed Sep 14 00:54:09 KST 2022 종료 날짜: Fri Sep 06 00:54:09 KST 2052
인증서 지문:
     SHA1: 3F:48:D2:E6:4A:2F:5D:FA:DB:33:4A:46:75:47:33:2E:72:D8:67:3C
     SHA256: 42:DE:13:B2:66:1B:29:41:E1:0E:6A:74:AA:CB:FC:E0:E4:67:EE:17:09:74:7B:81:80:49:1E:6C:9D:48:8A:DE
서명 알고리즘 이름: SHA1withRSA(약함)
주체 공용 키 알고리즘: 2048비트 RSA 키
버전: 1

Warning:
인증서 uses the SHA1withRSA signature algorithm which is considered a security risk.
```

### 사용하는 메서드가 다르다

Flutter에서 Google Sign-In을 구현할 때 웹과 네이티브 플랫폼(안드로이드) 간의 차이는 주로 인증 흐름의 차이 때문입니다:

1. 네이티브 플랫폼 (안드로이드)에서의 Google Sign-In:
   - 네이티브 플랫폼에서는 `google_sign_in` 플러그인을 사용하여 인증 흐름을 구현
   - 이 플러그인은 Google 로그인 프로세스를 트리거하고, 사용자 인증 정보를 얻은 후 `GoogleAuthProvider.credential`을 생성하여 Firebase 인증과 통합
2. 웹 플랫폼에서의 Google Sign-In:
   - 웹에서는 Firebase SDK가 인증 흐름을 자동으로 처리
   - `GoogleAuthProvider` 객체를 생성하고, 필요한 추가적인 권한 스코프를 설정한 후 `signInWithPopup` 메서드를 사용하여 사용자에게 로그인을 요청
   - 이 메서드는 새 창을 열어 사용자에게 로그인을 유도

이러한 차이는 각 플랫폼의 특성과 사용자 인터페이스에 최적화된 인증 경험을 제공하기 위함
네이티브 플랫폼에서는 **네이티브 Google 인증 흐름**을 사용하는 반면, 웹에서는 **팝업 또는 리디렉션 기반의 인증 흐름**을 제공
따라서 `gs.signIn();`을 웹에서 사용하는 것은 권장되지 않으며, 각 플랫폼에 적합한 방법을 사용하는 것이 바람직하다

### reCAPTCHA

```text
# 사이트에서 사용자에게 제공하는 HTML 코드에 이 사이트 키를 사용
# 클라이언트 측 통합 알아보기: https://developers.google.com/recaptcha/docs/v3
siteKey: 6LeiziopAAAAAKTg739rVpEsD4khRTjfjtsw_IKM

# 사이트와 reCAPTCHA 커뮤니케이션 위해 비밀키 사용
# 서버 측 통합 알아보기: https://developers.google.com/recaptcha/docs/verify?hl=ko
secretKey: 6LeiziopAAAAAN8oHoo_0QLtaYN3qRTMwKkFN59P
```

#### [reCAPTCHA v3](https://developers.google.com/recaptcha/docs/v3?hl=ko)

```html
<script src="https://www.google.com/recaptcha/api.js"></script>

<script>
    // 토큰을 처리할 콜백 함수
    function onSubmit(token) {
        document.getElementById("demo-form").submit();
    }
</script>

<!-- html 버튼에 속성을 추가 -->
<button class="g-recaptcha" 
        data-sitekey="reCAPTCHA_site_key" 
        data-callback='onSubmit' 
        data-action='submit'>Submit</button>
```

## [people API](https://console.cloud.google.com/apis/library/people.googleapis.com?project=dilectio-e5f94) 활성화

## 토큰

```log
[GSI_LOGGER-TOKEN_CLIENT]: The OAuth token was not passed to gapi.client, since the gapi.client library is not loaded in your page.
response.message is [firebase_auth/invalid-credential] Invalid Idp Response: access_token audience is not for this project
```
