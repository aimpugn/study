# Tinder imitation

- [Tinder imitation](#tinder-imitation)
    - [google auth](#google-auth)
        - [sign key](#sign-key)
    - [flutter create](#flutter-create)
    - [web integration](#web-integration)
        - [add meta tag](#add-meta-tag)
        - [Authorized JavaScript origins](#authorized-javascript-origins)
        - [firebase config](#firebase-config)
    - [Firestore](#firestore)
        - [enable](#enable)
        - [database](#database)
    - [기타](#기타)
        - [참고](#참고)

## google auth

### sign key

- 모드
    - debug: 앱 개발 중에 사용되는 키스토어. Flutter 또는 Android Studio가 자동으로 생성
    - release: 앱이 사용자에게 배포될 때 사용되는 키스토어. 개발자가 직접 생성해야 하며, 앱이 릴리즈 모드로 빌드될 때 이 키스토어를 사용하여 앱을 서명

```shell
keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
```

- [SHA-1 fingerprint of keystore certificate](https://stackoverflow.com/questions/15727912/sha-1-fingerprint-of-keystore-certificate/35308827#35308827)

```log
❯ keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
Alias name: androiddebugkey
Creation date: Sep 14, 2022
Entry type: PrivateKeyEntry
Certificate chain length: 1
Certificate[1]:
Owner: C=US, O=Android, CN=Android Debug
Issuer: C=US, O=Android, CN=Android Debug
Serial number: 1
Valid from: Wed Sep 14 00:54:09 KST 2022 until: Fri Sep 06 00:54:09 KST 2052
Certificate fingerprints:
         SHA1: 3F:48:D2:E6:4A:2F:5D:FA:DB:33:4A:46:75:47:33:2E:72:D8:67:3C
         SHA256: 42:DE:13:B2:66:1B:29:41:E1:0E:6A:74:AA:CB:FC:E0:E4:67:EE:17:09:74:7B:81:80:49:1E:6C:9D:48:8A:DE
Signature algorithm name: SHA1withRSA (weak)
Subject Public Key Algorithm: 2048-bit RSA key
Version: 1

Warning:
The certificate uses the SHA1withRSA signature algorithm which is considered a security risk. This algorithm will be disabled in a future update.
```

```diff
      "oauth_client": [
        {
+          "client_id": "425142454101-e4kqk56qmj399oioo451egcl9i58erc5.apps.googleusercontent.com",
+          "client_type": 1,
+          "android_info": {
+            "package_name": "com.fensterseifer.tinder_new",
+            "certificate_hash": "3f48d2e64a2f5dfadb334a467547332e72d8673c"
+          }
+        },
+        {
+          "client_id": "425142454101-1s2v2dp5b5cgk30p86h1b495qlfbnvtk.apps.googleusercontent.com",
          "client_type": 3
        }
      ],

```

## flutter create

특정 플랫폼만 생성할 수도 있다
- ios (default),
- android (default),
- windows (default),
- linux (default),
- macos (default),
- web (default)

```shell
flutter create --platform web .
```

```shell
❯ flutter create --platform web .
Recreating project ....
  test/widget_test.dart (created)
  tinder_new.iml (created)
  web/favicon.png (created)
  web/index.html (created)
  web/manifest.json (created)
  web/icons/Icon-maskable-512.png (created)
  web/icons/Icon-192.png (created)
  web/icons/Icon-maskable-192.png (created)
  web/icons/Icon-512.png (created)
  analysis_options.yaml (created)
  .idea/runConfigurations/main_dart.xml (created)
  .idea/libraries/Dart_SDK.xml (created)
  .idea/libraries/KotlinJavaRuntime.xml (created)
  .idea/modules.xml (created)
  .idea/workspace.xml (created)
Running "flutter pub get" in tinder_new...
Resolving dependencies... (2.4s)
  async 2.10.0 (2.11.0 available)
  characters 1.2.1 (1.3.0 available)
  collection 1.17.0 (1.17.2 available)
  flutter_plugin_android_lifecycle 2.0.14 (2.0.15 available)
  image_picker_android 0.8.6+14 (0.8.6+15 available)
  js 0.6.5 (0.6.7 available)
  lints 2.0.1 (2.1.0 available)
  matcher 0.12.13 (0.12.16 available)
  material_color_utilities 0.2.0 (0.5.0 available)
  meta 1.8.0 (1.9.1 available)
  path 1.8.2 (1.8.3 available)
  source_span 1.9.1 (1.10.0 available)
  test_api 0.4.16 (0.6.0 available)
  win32 4.1.4 (5.0.2 available)
Got dependencies!
Wrote 15 files.

All done!
You can find general documentation for Flutter at: https://docs.flutter.dev/
Detailed API documentation is available at: https://api.flutter.dev/
If you prefer video documentation, consider: https://www.youtube.com/c/flutterdev

In order to run your application, type:

  $ cd .
  $ flutter run

Your application code is in ./lib/main.dart.
```

## [web integration](https://pub.dev/documentation/google_sign_in_web/latest/)

### add meta tag

```html
<meta name="google-signin-client_id" content="YOUR_GOOGLE_SIGN_IN_OAUTH_CLIENT_ID.apps.googleusercontent.com">
```

### Authorized JavaScript origins

- go to [Credentials page](https://console.developers.google.com/apis/credentials)
    1. Credentials 좌측 메뉴 클릭
    2. CREATE CREDENTIALS 버튼 클릭
    3. Create OAuth client ID 선택
    4. consent screen 구성
        1. user type: Internal(x), External(o)
        2. app info
            1. app name: tinder-imitation
            2. email
    5. 테스트 유저: `Testing` 상태 동안 테스트 유저들만 앱에 접근 가능.
    6. 완료하면 앱이 생성되고, 다시 Create OAuth client ID 생성...?

### firebase config

```js
<script type="module">
  // Import the functions you need from the SDKs you need
  import { initializeApp } from "https://www.gstatic.com/firebasejs/9.22.0/firebase-app.js";
  import { getAnalytics } from "https://www.gstatic.com/firebasejs/9.22.0/firebase-analytics.js";
  // TODO: Add SDKs for Firebase products that you want to use
  // https://firebase.google.com/docs/web/setup#available-libraries

  // Your web app's Firebase configuration
  // For Firebase JS SDK v7.20.0 and later, measurementId is optional
  const firebaseConfig = {
    apiKey: "<API_KEY>",
    authDomain: "tinder-imitation.firebaseapp.com",
    projectId: "tinder-imitation",
    storageBucket: "tinder-imitation.appspot.com",
    messagingSenderId: "<messagingSenderId>",
    appId: "1:4251....:web:b6be52077.....",
    measurementId: "G-G91..."
  };

  // Initialize Firebase
  const app = initializeApp(firebaseConfig);
  const analytics = getAnalytics(app);
</script>
```

## Firestore

### enable

```log
base.js:24 [2023-05-19T18:29:06.406Z]  @firebase/firestore: Firestore (9.18.0): Could not reach Cloud Firestore backend. Connection failed 1 times. Most recent error: FirebaseError: [code=permission-denied]: Cloud Firestore API has not been used in project 425142454101 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/firestore.googleapis.com/overview?project=425142454101 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry.
This typically indicates that your device does not have a healthy Internet connection at the moment. The client will operate in offline mode until it is able to successfully connect to the backend
```

- enable `Google Cloud Firestore API`

> Cloud Firestore is a NoSQL document database that simplifies storing, syncing, and querying data for your mobile and web apps at global scale. Its client libraries provide live synchronization and offline support, while its security features and integrations with the Firebase and Google Cloud platforms accelerate building truly serverless apps.

### database

```log
base.js:24 [2023-05-19T18:34:43.360Z]  @firebase/firestore: Firestore (9.18.0): Could not reach Cloud Firestore backend. Connection failed 1 times. Most recent error: FirebaseError: [code=not-found]: The database (default) does not exist for project tinder-imitation Please visit https://console.cloud.google.com/datastore/setup?project=tinder-imitation to add a Cloud Datastore or Cloud Firestore database. 
This typically indicates that your device does not have a healthy Internet connection at the moment. The client will operate in offline mode until it is able to successfully connect to the backend.
```

- mode
    - Native mode(O)
    - Datastore mode
- location(**Your location selection is permanent**)
    - asia-northeast3 (Seoul)

## 기타

### 참고

- [Flutter: Implementing Google Sign In](https://medium.com/flutter-community/flutter-implementing-google-sign-in-71888bca24ed)
- [Flutter web: Firebase Authentication and Google Sign-In](https://blog.codemagic.io/flutter-web-firebase-authentication-and-google-sign-in/)
- [SHA-1 fingerprint of keystore certificate](https://stackoverflow.com/questions/15727912/sha-1-fingerprint-of-keystore-certificate/35308827#35308827)
- Firebase
    - [호스팅 동작 구성](https://firebase.google.com/docs/hosting/full-config?hl=ko)
    - [Get started with Cloud Storage on Web](https://firebase.google.com/docs/storage/web/start)
    - [Learn the core syntax of the Firebase Security Rules for Cloud Storage language](https://firebase.google.com/docs/storage/security/core-syntax)
    - [Use conditions in Firebase Cloud Storage Security Rules](https://firebase.google.com/docs/storage/security/rules-conditions#public)
- [Bouncing Button Animation In Flutter](https://medium.flutterdevs.com/bouncing-button-animation-in-flutter-d0296442f3c5)
- [Custom Swatch for Material App Theme – primarySwatch](https://dev.to/rohanjsh/custom-swatch-for-material-app-theme-primaryswatch-3kic)
