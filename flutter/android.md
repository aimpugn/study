# Android

- [Android](#android)
    - [Build and release an Android app](#build-and-release-an-android-app)
    - [Android 프로젝트에 Firebase 추가](#android-프로젝트에-firebase-추가)
        - [로컬 머신에 Flutter 환경 설정](#로컬-머신에-flutter-환경-설정)
        - [Firebase 프로젝트 설정](#firebase-프로젝트-설정)
        - [안드로이드 빌드 설정](#안드로이드-빌드-설정)
        - [앱 빌드 및 실행](#앱-빌드-및-실행)
        - [5. 플레이 스토어 배포 (선택적)](#5-플레이-스토어-배포-선택적)
    - [Flutter 앱에 Firebase 추가](#flutter-앱에-firebase-추가)
        - [CLI 설정](#cli-설정)
        - [Firebase 사용하도록 앱 구성](#firebase-사용하도록-앱-구성)
        - [앱에서 Firebase 초기화](#앱에서-firebase-초기화)
    - [Material Component 활성화](#material-component-활성화)
    - [SDK 버전 지정](#sdk-버전-지정)
    - [`DEX`](#dex)
        - [DEX (Dalvik Executable)](#dex-dalvik-executable)
        - [Dalvik 가상 머신 -\> ART로 대체](#dalvik-가상-머신---art로-대체)
        - [Android Runtime (ART)](#android-runtime-art)
        - [64K 참조 제한과 `Multi-DEX`의 관계](#64k-참조-제한과-multi-dex의-관계)
        - [메서드가 64K개를 초과하는 앱에 관해 멀티덱스 사용 설정](#메서드가-64k개를-초과하는-앱에-관해-멀티덱스-사용-설정)
    - [용어](#용어)
        - [`APK`(Android Package Kit)](#apkandroid-package-kit)

## [Build and release an Android app](https://docs.flutter.dev/deployment/android)

## [Android 프로젝트에 Firebase 추가](https://firebase.google.com/docs/android/setup?hl=ko&authuser=0)

### 로컬 머신에 Flutter 환경 설정

- Flutter SDK 설치
    - Flutter SDK가 설치되어 있지 않은 경우, [Flutter 공식 웹사이트](https://flutter.dev)에서 다운로드하여 설치
- Flutter 환경 확인
    - 터미널이나 커맨드 프롬프트에서 `flutter doctor` 명령어를 실행하여 Flutter 환경이 올바르게 설정되었는지 확인
    - 필요한 도구들이 모두 설치되어 있고, 설정이 올바른지 확인

### Firebase 프로젝트 설정

1. [Firebase 콘솔](https://console.firebase.google.com/)에서 새 프로젝트를 생성
2. Firebase 콘솔에서 안드로이드 앱을 프로젝트에 추가. 이 과정에서 **안드로이드 앱의 패키지 이름이 필요**하다.
3. Firebase 설정 과정에서 생성되는 `google-services.json` 파일을 다운로드하고, 이 파일을 Flutter 프로젝트의 `android/app` 디렉토리에 복사
4. `android/build.gradle` 파일과 `android/app/build.gradle` 파일에 Firebase 관련 종속성과 플러그인을 추가
   - `android/build.gradle` 파일에는 `classpath 'com.google.gms:google-services:4.x.x'`을 추가
   - `android/app/build.gradle` 파일 맨 아래에 `apply plugin: 'com.google.gms.google-services'`를 추가

```gradle
// 루트(프로젝트) 수준
// `<project>/build.gradle.kts` 또는 `<project>/build.gradle`
plugins {
  id 'com.android.application' version '7.2.0' apply false
  // ...

  // Add the dependency for the Google services Gradle plugin
  id 'com.google.gms.google-services' version '4.3.15' apply false
}
```

```gradle
// 모듈(앱 수준)
// `<project>/<app-module>/build.gradle.kts` 또는 `<project>/<app-module>/build.gradle`
plugins {
  id 'com.android.application' version '7.2.0' apply false
  // ...

  // Add the dependency for the Google services Gradle plugin
  id 'com.google.gms.google-services' version '4.3.15' apply false
}
```

### 안드로이드 빌드 설정

1. 안드로이드 매니페스트 설정: `android/app/src/main/AndroidManifest.xml` 파일에서 필요한 권한을 설정
2. 빌드 버전 설정: `android/app/build.gradle` 파일에서 `versionCode`와 `versionName`을 설정
3. 키스토어 설정:
    - 릴리스 빌드를 위해 안드로이드 키스토어를 설정
    - `key.properties` 파일을 생성하고, `android/app/build.gradle` 파일에 키스토어 정보를 추가

### 앱 빌드 및 실행

1. 디버그 모드로 실행: 개발 중에는 `flutter run` 명령어를 사용하여 디바이스나 에뮬레이터에서 앱을 실행합니다.

2. 릴리스 빌드 생성: 릴리스 버전의 APK 또는 App Bundle을 생성하기 위해 `flutter build apk` 또는 `flutter build appbundle` 명령어를 사용합니다.

3. 릴리스 빌드 테스트: 빌드된 릴리스 버전의 앱을 디바이스에 설치하여 테스트합니다.

### 5. 플레이 스토어 배포 (선택적)

1. Google Play Console 설정: Google Play Console에 앱을 등록하고, 릴리스 관리 섹션에서 APK 또는 App Bundle을 업로드합니다.

2. 스토어 상세 정보 입력: 앱의 상세 정보, 스크린샷, 설명 등을 입력합니다.

3. 리뷰 및 출시: 앱을 출시하기 전에 Google Play의 리뷰 프로세스를 거칩니다.

이러한 단계를 거쳐 Flutter 앱을 안드로이드 플랫폼으로 빌드하고 배포할 수 있습니다. Firebase를 사용

하는 경우, Firebase와 관련된 설정을 올바르게 완료하는 것이 중요합니다. Flutter와 Firebase의 버전이 서로 호환되는지도 확인해야 합니다.

## Flutter 앱에 Firebase 추가

- [Flutter 앱에 Firebase 추가](https://firebase.google.com/docs/flutter/setup?hl=ko&platform=android)

### CLI 설정

```shell
firebase login
```

### Firebase 사용하도록 앱 구성

```shell
flutterfire configure
```

- Flutter 앱에서 지원되는 플랫폼(iOS, Android, Web)을 선택하도록 요청
    - 선택한 각 플랫폼에 대해 FlutterFire CLI가 Firebase 프로젝트에서 새 Firebase 앱을 만든다
    - 기존 Firebase 프로젝트를 사용하거나 새 Firebase 프로젝트를 만들도록 선택할 수 있다
    - 기존 Firebase 프로젝트에 등록된 앱이 있으면 FlutterFire CLI가 현재 Flutter 프로젝트 구성을 기준으로 일치하는 항목을 찾으려고 시도한다

> 참고: 다음은 Firebase 프로젝트 설정 및 관리에 관한 몇 가지 팁입니다.
> - 여러 대안을 처리하는 방법을 비롯하여 앱을 Firebase 프로젝트에 추가하는 방법을 자세히 알아보려면 권장사항을 확인하세요.
> - 프로젝트에서 Google 애널리틱스를 사용 설정하세요. 그러면 Crashlytics 및 원격 구성과 같은 많은 Firebase 제품의 사용 환경을 최적화할 수 있습니다.

- Firebase 구성 파일(`firebase_options.dart`)을 만들고 이를 `lib/` 디렉터리에 추가

> 참고: 이 Firebase 구성 파일에는 고유하지만 선택한 각 플랫폼에 대해 보안 비밀은 아닌 식별자가 있습니다.
> 이 구성 파일에 대한 자세한 내용은 Firebase 프로젝트 이해를 참조하세요.

- (Android의 Crashlytics 또는 Performance Monitoring의 경우) 필요한 제품별 Gradle 플러그인을 Flutter 앱에 추가합니다.

> 참고: FlutterFire CLI로 적합한 Gradle 플러그인을 추가하려면 제품의 Flutter 플러그인을 Flutter 앱으로 이미 가져온 상태여야 합니다.

- 이렇게 `flutterfire configure`를 처음 실행한 후에는 **다음 경우**에 언제든지 명령어를 다시 실행해야 한다
    - Flutter 앱에서 새 플랫폼 지원을 시작하는 경우
    - 특히 `Google`, `Crashlytics`, `Performance Monitoring`, `Realtime Database`로 로그인을 시작할 때와 같이 Flutter 앱에서 새 Firebase 서비스 또는 제품을 사용하는 경우

### 앱에서 Firebase 초기화

```shell
# Flutter 프로젝트 디렉터리에서 다음 명령어를 실행하여 core 플러그인을 설치
flutter pub add firebase_core

# Flutter 앱의 Firebase 구성이 최신 상태인지 확인
flutterfire configure
```

이를 실제로 실행하면 아래와 같이 진행된다

```shell
╰─ flutterfire configure
i Found 6 Firebase projects.                                                                                                                                                                                 
✔ Select a Firebase project to configure your Flutter application with · dilectio-e5f94 (dilectio)                                                                                                           
✔ Which platforms should your configuration support (use arrow keys & space to select)? · android, web                                                                                                       
i Firebase android app com.bmd.dilectio registered.                                                                                                                                                          
i Firebase web app dilectio_app (web) registered.                                                                                                                                                            

Firebase configuration file lib/firebase_options.dart generated successfully with the following Firebase apps:

Platform  Firebase App Id
web       1:928175087752:web:ee6a879ff6dda1793cbf3d
android   1:928175087752:android:5b1aaa87db7f8ccc3cbf3d

Learn more about using this file and next steps from the documentation:
 > https://firebase.google.com/docs/flutter/setup
```

그리고 `lib/main.dart`에서 구성 파일 가져오도록 한다

```dart
import 'package:firebase_core/firebase_core.dart';
import 'firebase_options.dart';

// `lib/main.dart` 파일에서 구성 파일로 내보낸 `DefaultFirebaseOptions` 객체를 사용하여 Firebase를 초기화
await Firebase.initializeApp(
  options: DefaultFirebaseOptions.currentPlatform,
);
```

애플리케이션을 다시 빌드

```shell
flutter run
```

## [Material Component 활성화](https://docs.flutter.dev/deployment/android#enabling-material-components)

```gradle
// <my-app>/android/app/build.gradle:
dependencies {
    // ...
    implementation 'com.google.android.material:material:<version>'
    // ...
}
```

버전은 [Google Maven Repository](https://maven.google.com/web/index.html#com.google.android.material:material)에서 확인

## SDK 버전 지정

앱 수준 `android/app/build.gradle`에 설정

```gradle
android {
    ...

    defaultConfig {
        ...
        minSdkVersion 16
        targetSdkVersion 29
        ...
    }
}
```

## `DEX`

- `DEX` (Dalvik Executable):
    - Android 시스템에서 실행 가능한 바이트 코드 파일 형식
    - Java나 Kotlin으로 작성된 코드는 `DEX` 형식으로 컴파일되어 Android 디바이스에서 실행된다
    - Android 앱이 빌드될 때 Java/Kotlin 코드는 DEX 파일로 컴파일되며, 이 파일들은 APK 내에 포함된다
- `Multi-DEX`:
    - 단일 DEX 파일에서 허용되는 64K(65,536) 메서드 제한을 넘어서는 경우, 여러 개의 DEX 파일을 사용하여 이 제한을 우회하는 기술

### DEX (Dalvik Executable)

- `DEX` 파일은 Android 플랫폼에서 실행되는 애플리케이션을 위한 컴파일된 바이트코드 파일
- Java 클래스 파일을 Dalvik 가상 머신이 실행할 수 있는 형식으로 변환하여 포함한다
- `DEX` 파일 포맷은 메모리와 처리 효율을 최적화하기 위해 설계되었다. 이는 특히 메모리 리소스가 제한적인 모바일 디바이스에서 중요한 요소
- `DEX` 파일은 여러 Java 클래스를 하나의 DEX 파일 안에 효율적으로 압축하여 저장한다. 이는 실행 시 메모리 사용량을 줄이고, 애플리케이션의 성능을 향상시키는 데 도움을 준다.

### Dalvik 가상 머신 -> ART로 대체

- Dalvik은 Android 애플리케이션을 실행하기 위한 가상 머신. Android 4.4 KitKat 버전 이전까지의 Android 플랫폼에서 주로 사용되었다.
- Dalvik은 각 앱마다 별도의 프로세스로 실행되며, 각 프로세스에는 자체 Dalvik 인스턴스가 할당되고, 이는 앱 간의 격리와 보안을 강화한다.
- Dalvik은 Just-In-Time (JIT) 컴파일 방식을 사용했으며, 이는 애플리케이션을 실행할 때마다 필요한 코드를 실시간으로 컴파일하는 방식
- Dalvik 가상 머신은 Dan Bornstein에 의해 개발되었으며, 이름은 그가 한때 거주했던 아이슬란드의 마을 Dalvík에서 따왔다.
- Dalvik은 ART로 대체 되었으며, 이는 실제로 Android 운영 체제에서 애플리케이션을 실행하는 방식의 변화를 의미한다

### Android Runtime (ART)

- Android 5.0 (Lollipop) 이후, Dalvik은 Android Runtime (ART)으로 대체되었다.
- ART는 Dalvik에 비해 향상된 성능, 효율성, 가비지 컬렉션 메커니즘을 제공
- ART는 설치 시 앱의 `DEX` 파일을 기계어로 컴파일하는 `AOT` (Ahead-Of-Time) 컴파일 방식을 사용
    - 애플리케이션 설치 시 코드를 미리 컴파일
    - 앱의 실행 속도를 향상시키고, 배터리 수명을 연장하는 데 도움을 준다
- Dalvik과 ART 모두 DEX (Dalvik Executable) 파일 포맷을 사용. `DEX`는 Java 클래스 파일을 Android 디바이스에서 실행할 수 있는 형식으로 변환한 것이므로, Dalvik이 ART로 대체되었어도 DEX 파일 포맷 자체는 여전히 사용된다

### 64K 참조 제한과 `Multi-DEX`의 관계

- 64K 참조 제한
    - 단일 DEX 파일 내에서 참조될 수 있는 메서드의 최대 개수
    - 이 한계를 넘으면 앱이 빌드되지 않는다
- `Multi-DEX`와의 관계
    - Multi-DEX를 사용하면 하나의 앱이 여러 DEX 파일을 포함할 수 있다
    - 따라서 64K 참조 제한을 넘길 수 있다

### [메서드가 64K개를 초과하는 앱에 관해 멀티덱스 사용 설정](https://developer.android.com/build/multidex?hl=ko)

## 용어

### `APK`(Android Package Kit)

- Android 애플리케이션을 배포하기 위한 패키지 파일 형식
- APK 파일은 앱의 코드, 리소스, 자산, 인증서 등을 포함
