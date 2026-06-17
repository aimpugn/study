# 컴파일은 시그니처만 본다 — classpath, 컴파일 vs 런타임 에러, --release

자바 빌드에서 사람을 자주 놀라게 하는 사실 두 가지가 있다. 첫째, **컴파일은 의존성의 *시그니처*(타입·메서드 모양)만 보고, 구현 내용은 보지 않는다.** 그래서 모양만 같은 JAR이면 내용이 무엇이든 컴파일이 된다. 둘째, **JDK는 자기보다 *낮은* 버전의 바이트코드는 만들 수 있어도, *높은* 버전은 못 만든다.** JDK 17로는 "Java 21로 컴파일"이 안 된다. 이 두 사실이 빌드 동작의 많은 부분을 설명한다. 특정 프로젝트와 무관한 일반 메커니즘만 다룬다.

읽고 나면 다음을 스스로 설명할 수 있어야 한다 — `cannot find symbol`(컴파일)과 `NoClassDefFoundError`/`NoSuchMethodError`(런타임)가 왜 다른 단계의 문제인가, 왜 모양만 맞는 JAR이 컴파일을 통과시키는가, 그리고 `release version 21 not supported`가 무슨 뜻인가.

## 목차

- [1. 직답](#1-직답)
- [2. classpath란 무엇인가](#2-classpath란-무엇인가)
- [3. 컴파일은 시그니처만 본다](#3-컴파일은-시그니처만-본다)
- [4. 컴파일타임 에러 vs 런타임 에러](#4-컴파일타임-에러-vs-런타임-에러)
- [5. --release, source/target, 바이트코드 버전](#5---release-sourcetarget-바이트코드-버전)
- [6. JAR은 zip이다](#6-jar은-zip이다)
- [7. 혼동하기 쉬운 점](#7-혼동하기-쉬운-점)
- [8. 실패와 검증](#8-실패와-검증)
- [9. 스스로 확인하기](#9-스스로-확인하기)
- [10. 확인한 사실과 추론](#10-확인한-사실과-추론)
- [관련 노트](#관련-노트)

## 1. 직답

`javac`는 코드가 참조하는 타입·메서드·필드를 **classpath 위의 시그니처**에 맞춰 검사한다. 그 의존성의 메서드 *내부 구현*은 실행하지도, 들여다보지도 않는다. 그래서 같은 패키지·클래스명·메서드 시그니처를 가진 클래스라면 그 안이 무엇이든 컴파일을 만족시킨다.

그리고 컴파일러는 자기 버전 이하의 클래스 파일만 만들 수 있다. `--release N`은 "Java N의 API에 맞춰 컴파일하고 N용 바이트코드를 내라"는 뜻인데, N이 그 JDK보다 높으면 컴파일러가 그 릴리스를 모르므로 거부한다.

## 2. classpath란 무엇인가

classpath는 JVM·컴파일러가 클래스를 찾는 위치 목록이다. 디렉터리와 JAR 파일들의 나열이며, `-cp`(또는 `-classpath`)로 준다. 자바의 패키지는 경로로 대응한다 — `com.acme.Crypto`는 classpath의 어느 항목 안 `com/acme/Crypto.class`로 찾는다.

JAR은 그 자체가 `.class`들을 담은 압축 파일(zip)이므로, classpath에 JAR을 올리면 그 안의 클래스들이 보인다(6절). Maven은 이 classpath를 의존성 해석 결과로 자동 구성해 `javac`/`java`에 넘긴다(의존성 해석 자체는 [`./maven_dependency_resolution.md`](./maven_dependency_resolution.md)).

## 3. 컴파일은 시그니처만 본다

`javac`가 `String s = Crypto.enc(x);`를 컴파일할 때 하는 일은, classpath에서 `com.acme.Crypto`를 찾아 `enc(...)`라는 메서드가 *그 시그니처로 존재하는지*, 반환형이 호환되는지를 확인하는 것뿐이다. `enc`가 내부에서 무엇을 하는지는 컴파일과 무관하다.

그래서 다음 둘은 컴파일러 입장에서 완전히 동등하다.

```java
// 진짜 라이브러리
public static String enc(String s) { /* 복잡한 암호화 */ }
// 모양만 같은 대체 구현
public static String enc(String s) { return s; }   // 아무 내용
```

둘 다 `com.acme.Crypto.enc(String): String` 시그니처를 제공하므로, 이 메서드를 호출하는 코드는 어느 쪽으로도 컴파일된다. 이 성질이 "받을 수 없는 의존성을 모양만 맞춘 대체물로 채워 컴파일을 통과시키는" 기법의 근거다([`./maven_local_shim_repo.md`](./maven_local_shim_repo.md)). 반대로, 호출하는 메서드가 classpath의 타입에 *없으면* 컴파일이 막힌다 — 이게 `cannot find symbol`이다.

## 4. 컴파일타임 에러 vs 런타임 에러

같은 "없음"이라도 *언제* 드러나는지가 다르고, 그게 원인을 가리킨다.

- **`cannot find symbol`** (javac, 컴파일 단계) — 코드가 참조하는 타입·메서드·필드가 *컴파일 classpath의 시그니처에 없다*. 의존성이 빠졌거나, 버전이 그 심볼을 아직/이미 안 가졌거나, 오타다.
- **`NoClassDefFoundError`** (런타임) — 컴파일은 됐는데 *실행 시점 classpath에 그 클래스가 없다*. 컴파일과 런타임 classpath가 다를 때(예: `provided`인데 런타임에 그 제공자가 없음).
- **`NoSuchMethodError` / `NoSuchFieldError`** (런타임) — 클래스는 있는데 *그 메서드/필드가 없다*. 시그니처 X로 컴파일했는데 실행 시점 JAR은 다른 버전이라 그 메서드가 사라졌거나 시그니처가 바뀐 경우. 전형적인 "버전 스큐(skew)".

핵심 결론: **"컴파일됐으니 실행도 될 것"은 거짓이다.** 컴파일은 시그니처에 대한 약속일 뿐이고, 런타임에 그 약속을 지키는 실제 클래스가 같은 모양으로 있어야 한다. 그래서 모양만 맞춘 대체물로 컴파일을 통과시켜도, 런타임에 그 동작 결과까지 필요하면 내용도 맞아야 한다.

## 5. --release, source/target, 바이트코드 버전

모든 `.class` 파일에는 그것을 만든 자바 버전을 나타내는 **major version** 번호가 박힌다(예: Java 17 = 61, Java 21 = 65). JVM은 자기보다 *높은* major version의 클래스를 실행하지 못한다 — 실행하면 `UnsupportedClassVersionError`다.

컴파일러 쪽 옵션은 세 가지다.

- `--release N` — Java N의 표준 API로만 컴파일하고 N용 바이트코드를 낸다. 가장 정확한 크로스 컴파일 방식(상위 JDK에서 하위 타깃을 안전하게).
- `-source N` / `-target N` — 각각 소스 문법 레벨과 바이트코드 레벨. `--release`와 달리 API 제한이 없어, 상위 JDK에서 빌드하면 하위 타깃인데도 상위 API를 우연히 써 버리는 함정이 있다.

여기서 비대칭이 핵심이다 — 컴파일러는 **자기 버전 이하**의 릴리스만 안다. 그래서 JDK 17로 `--release 21`을 주면 `error: release version 21 not supported`가 난다. 21을 타깃하려면 JDK 21 이상으로 컴파일해야 한다. 반대로 JDK 21로 `--release 17`은 정상이다(하위 타깃).

실무 함의: 한 빌드 안에서도 모듈마다 요구 릴리스가 다르면, *가장 높은 릴리스를 감당할 JDK*로 그 단계를 돌려야 한다. 상위 JDK는 하위 타깃을 만들 수 있으니, 여러 JDK를 두고 단계별로 `JAVA_HOME`을 바꿔 끼우는 게 일반적인 해법이다(Maven toolchains로 모듈별 JDK를 지정하는 방법도 있다).

## 6. JAR은 zip이다

JAR(Java ARchive)은 `.class`와 리소스를 담은 zip 파일에 `META-INF/MANIFEST.MF`(아카이브 전체의 메타데이터를 적는 파일) 등을 더한 것이다. 그래서 내용은 zip 도구나 `jar` 명령으로 그대로 들여다볼 수 있다.

```bash
jar tf some-lib.jar | grep Crypto      # 어떤 클래스가 들어 있나
#  com/acme/Crypto.class
```

어떤 JAR이 기대한 클래스를 담고 있는지, 또는 좌표 자리에 엉뚱한 JAR이 깔렸는지 확인할 때 이 한 줄이 결정적이다. (Multi-Release JAR은 `META-INF/versions/N/`에 버전별 클래스를 따로 담아 한 JAR이 여러 런타임을 지원하기도 한다.)

## 7. 혼동하기 쉬운 점

- **컴파일 통과 ≠ 런타임 안전** — 4절. 시그니처 약속과 실제 구현 존재는 별개다.
- **`provided` scope의 런타임 부재** — 컴파일엔 있으나 패키지에 안 들어가, 실행 환경이 그 클래스를 안 주면 `NoClassDefFoundError`.
- **`UnsupportedClassVersionError`는 5절 규칙의 런타임 거울** — "상위 바이트코드를 하위 JVM에서 실행"하면 난다. 컴파일은 됐는데 더 낮은 JVM에서 돌릴 때.
- **`-source/-target`의 함정** — 상위 JDK에서 하위 타깃을 만들 때 상위 API가 새어 들어가면, 컴파일은 되는데 하위 런타임에서 `NoSuchMethodError`가 난다. `--release`가 이를 막는다.

## 8. 실패와 검증

에러 메시지가 곧 단계를 가리킨다.

- `cannot find symbol` → 컴파일 classpath/시그니처 문제. 의존성·버전·오타 점검. `mvn dependency:tree`로 그 좌표가 실제로 들어왔는지 확인.
- `release version N not supported` → 컴파일 JDK가 N보다 낮음. `java -version`/`JAVA_HOME` 확인, N 이상 JDK로 그 단계 빌드.
- `UnsupportedClassVersionError` → 실행 JVM이 클래스보다 낮음. 실행 JDK를 올리거나 더 낮은 타깃으로 컴파일.
- `NoSuchMethodError`/`NoClassDefFoundError` → 컴파일·런타임 classpath 버전 스큐. 런타임에 실제 올라온 JAR 버전을 확인.
- 시그니처·버전 직접 확인: `javap -p <클래스>`(메서드 시그니처), `javap -verbose <클래스> | findstr "major"`(바이트코드 major version).

## 9. 스스로 확인하기

- `javac`가 의존성의 메서드 *구현*을 보지 않는다는 사실이, 왜 "모양만 같은 JAR이 컴파일을 통과"시키는 걸 가능하게 하나?
- `cannot find symbol`과 `NoSuchMethodError`는 각각 어느 단계의 문제이고, 무엇이 다른가?
- JDK 17에서 `--release 21`이 거부되는 이유는? JDK 21에서 `--release 17`은 왜 괜찮나?
- "컴파일됐으니 실행도 된다"가 거짓인 시나리오를 하나 들어 보라.
- 좌표 자리에 엉뚱한 JAR이 깔렸는지 어떻게 한 줄로 확인하나?

## 10. 확인한 사실과 추론

사실: classpath 기반 시그니처 컴파일, `cannot find symbol` vs 런타임 링크 에러, 클래스 파일 major version과 `UnsupportedClassVersionError`, `--release`의 크로스 컴파일 의미는 javac/JVM 명세와 공식 문서로 확인되는 표준 동작이다.

추론·주의: 특정 major version 번호(17=61, 21=65)는 사실이지만 버전이 늘면 갱신해야 한다. Multi-Release JAR이나 모듈 시스템(JPMS)이 끼면 classpath 동작이 더 복잡해질 수 있어, 실제 문제는 `javap`·`dependency:tree`로 확인하는 편이 안전하다.

## 관련 노트

- 의존성이 어떻게 classpath에 올라오나(좌표·해석): [`./maven_dependency_resolution.md`](./maven_dependency_resolution.md)
- 이 성질의 실제 적용(시그니처만 맞춘 shim으로 컴파일 통과): [`./maven_local_shim_repo.md`](./maven_local_shim_repo.md)
- 빌드 페이즈에서 컴파일·테스트가 도는 순서: [`./maven_build_lifecycle.md`](./maven_build_lifecycle.md)
