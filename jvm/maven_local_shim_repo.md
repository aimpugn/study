# 받을 수 없는 사내 의존성으로 묶인 Maven 빌드를 로컬에서 되살리기 (shim 로컬 저장소)

금융권처럼 망이 분리된 환경의 코드는 보통 사내 Nexus나 내부망에만 있는 프레임워크 JAR(예: `kpn.co.kr:sfm5-core`, `kpn.coocon:CcSecurity_jdk1.6`)에 의존한다. 이 코드를 망 밖(개인 PC, 외부 CI, AI 에이전트 샌드박스)으로 가져오면 `mvn`이 의존성 해석 단계에서 바로 멈춘다. 받을 수 없는 JAR이기 때문이다. 그러면 컴파일도, 테스트도, AI를 통한 분석·수정도 할 수 없다.

이 문서는 그 빌드를 **사내 JAR 없이** 로컬에서 되살리는 기법을 설명한다. 핵심 도구는 "**shim 모듈을 만들어 회사와 똑같은 좌표로 프로젝트-로컬 Maven 저장소에 설치**"하는 것이다. 실제 사례로는 한 펌뱅킹 서비스(이하 *대상 프로젝트*)를 외부에서 빌드·테스트 가능하게 만든 작업을 인용하되, 기법 자체는 좌표 기반 의존성을 쓰는 모든 빌드 시스템으로 전이된다.

읽고 나면 다음을 스스로 설명할 수 있어야 한다 — *왜* 가짜 JAR이 진짜 의존성을 만족시키는가, alias POM이 *왜* 따로 필요한가, shim을 *어디까지* 만들어야 하는가(컴파일 vs 동작), 그리고 이 빌드가 깨질 때 *어디를* 봐야 하는가.

## 목차

- [1. 한 줄 원리](#1-한-줄-원리)
- [2. 문제는 어떻게 나타나나](#2-문제는-어떻게-나타나나)
- [3. shim을 만들기 전에 알아야 할 세 가지](#3-shim을-만들기-전에-알아야-할-세-가지)
- [4. 가장 작은 동작 예제](#4-가장-작은-동작-예제)
- [5. 실제 메커니즘: 좌표 alias 설치](#5-실제-메커니즘-좌표-alias-설치)
- [6. 규모를 키우는 자동화: reactor + 설치 스크립트](#6-규모를-키우는-자동화-reactor--설치-스크립트)
- [7. 혼동하기 쉬운 인접 개념](#7-혼동하기-쉬운-인접-개념)
- [8. 실패 모드와 함정](#8-실패-모드와-함정)
- [9. 검증 경로](#9-검증-경로)
- [10. 다른 생태계로의 전이](#10-다른-생태계로의-전이)
- [11. 스스로 확인하기](#11-스스로-확인하기)
- [12. 확인한 사실과 추론](#12-확인한-사실과-추론)
- [관련 노트](#관련-노트)

## 1. 한 줄 원리

Maven은 의존성을 **좌표**(coordinate = `groupId:artifactId:version`)로 해석한다. JAR의 *내용*이 진짜인지, 누가 만들었는지, 원본 소스가 무엇인지는 따지지 않는다. 그래서 내가 만든 아무 JAR이라도 `kpn.coocon:CcSecurity_jdk1.6:1.0.4`라는 좌표로 로컬 저장소에 넣어 두면, 그 좌표를 요구하는 의존성은 만족된다.

이 한 가지 사실이 지렛대다. "없는 의존성을 받아오는" 문제를, "같은 *이름표*를 단 내 구현으로 그 자리를 채우는" 문제로 바꾼다.

여기서 "shim"은 진짜 라이브러리의 API 표면 중 **대상 프로젝트가 실제로 호출하는 부분만** 최소로 다시 구현한 가짜 모듈을 뜻한다. 어댑터/대역(stand-in)에 가깝고, 전체 라이브러리를 복제하는 것이 아니다.

## 2. 문제는 어떻게 나타나나

대상 프로젝트의 `pom.xml`은 사내 좌표를 의존성으로 선언한다.

```xml
<dependency>
    <groupId>kpn.co.kr</groupId>
    <artifactId>sfm5-core</artifactId>
    <version>1.1.3</version>
</dependency>
<dependency>
    <groupId>kpn.coocon</groupId>
    <artifactId>CcSecurity_jdk1.6</artifactId>
    <version>1.0.4</version>
</dependency>
```

이 좌표들은 Maven Central에 없다. 망 밖에서 빌드하면 의존성 해석에서 멈춘다(첫 번째 벽돌 — 독자가 붙잡을 실제 증상).

```text
[ERROR] Failed to execute goal ... on project app:
  Could not resolve dependencies for project ...:
  kpn.coocon:CcSecurity_jdk1.6:jar:1.0.4 was not found in
  https://repo.maven.apache.org/maven2 ...
```

중요한 점: 이건 컴파일 오류가 아니라 그 **이전 단계**인 의존성 해석 실패다. 소스 한 줄도 컴파일되기 전에 끝난다. 그래서 "코드를 고쳐서" 해결할 수 있는 문제가 아니라, **의존성 그래프를 채워야** 풀리는 문제다.

`pom.xml`의 이 좌표들은 함부로 지우면 안 된다. 운영 환경에서는 이 좌표가 진짜 사내 JAR을 가리키는 **계약**이기 때문이다. 로컬 빌드를 위해 좌표를 바꾸면 그 계약 증거가 사라진다. 그래서 좌표는 그대로 두고, 그 좌표를 *로컬에서만* 다른 JAR로 채우는 방향으로 간다.

## 3. shim을 만들기 전에 알아야 할 세 가지

shim은 추측으로 만들 수 없다. 대상 프로젝트의 구조를 먼저 읽어야 한다.

1. **어떤 좌표를 요구하는가** — `pom.xml`의 사내 의존성 좌표 목록. 이걸 그대로(같은 group/artifact/version) 재현해야 한다. 여러 모듈이 같은 라이브러리의 *다른 버전*을 요구하면, 그 버전들도 모두 재현 대상이다.

2. **그 라이브러리의 무엇을 실제로 호출하는가** — 전체 API가 아니라 *쓰이는 표면*만 필요하다. 찾는 법은 단순하다. import 문을 grep하고, 컴파일을 돌려 "cannot find symbol" 오류가 가리키는 클래스·메서드를 모은다. 예를 들어 대상 프로젝트는 보안 라이브러리에서 정적 메서드 세 개만 썼다.

    ```java
    import cc.checkpay.common.CcSecurityUtil;
    ...
    String ev  = CcSecurityUtil.EncryptAes256Base64(msg, key, true);
    String vv  = CcSecurityUtil.getHmacSha256(body, key, true);
    String dec = CcSecurityUtil.DecryptAes256Base64(rev, key, "UTF-8", true);
    ```

    그러면 shim의 `CcSecurityUtil`은 이 세 메서드만 가지면 컴파일을 만족시킨다. 이 표면 파악은 한 번에 끝나지 않고 반복 루프다 — 컴파일을 돌리면 `cannot find symbol`이 다음 빠진 심볼을 가리키고, 그걸 shim에 더해 다시 컴파일하기를 더 이상 빠진 게 없을 때까지 반복해 좁혀 간다.

3. **Maven이 로컬 저장소를 어떻게 고르는가** — 기본값은 `~/.m2/repository`다. 이걸 프로젝트별로 바꾸는 깔끔한 방법이 저장소 루트의 `.mvn/maven.config`다. 이 파일에 적은 인자는 그 디렉터리에서 실행되는 모든 `mvn`에 자동으로 붙는다.

    ```text
    -Dmaven.repo.local=.codex-m2
    ```

    이렇게 하면 전역 `~/.m2`를 건드리지 않고, 이 프로젝트만 저장소 루트의 `.codex-m2` 폴더를 로컬 저장소로 쓴다. shim도 여기로 설치하면 빌드가 자동으로 찾는다.

오해 교정: shim은 진짜 라이브러리의 동작을 *일반적으로* 똑같이 만들 필요가 없다. **컴파일에는 항상** 필요하지만, **런타임에는 테스트가 실제로 실행하는 경로만** 필요하다. 그래서 순수 계산(예: AES 암복호화)처럼 테스트가 결과값을 단언하는 부분은 진짜와 같은 알고리즘으로 충실히 구현하고, 외부 네트워크 호출(HSM·소켓·HTTP)처럼 부수효과만 있는 부분은 "부르면 즉시 명확한 예외(fail-fast)"로 닫는 편이 안전하다. 가짜 성공값을 만들어 두면 테스트가 그 허구를 정답으로 굳혀 버린다.

## 4. 가장 작은 동작 예제

원리를 최소 형태로 보자. 어떤 프로젝트가 `acme:secret-lib:1.0`에 의존하고 `com.acme.Crypto.enc(String)` 하나만 호출한다고 하자.

먼저 shim 소스를 만든다.

```java
// shim/src/main/java/com/acme/Crypto.java
package com.acme;
public final class Crypto {
    private Crypto() {}
    public static String enc(String s) {           // 쓰이는 표면만
        return java.util.Base64.getEncoder()
                 .encodeToString(s.getBytes(java.nio.charset.StandardCharsets.UTF_8));
    }
}
```

이걸 JAR로 만들고(`javac` + `jar`, 또는 작은 Maven 모듈), 진짜 좌표로 로컬 저장소에 설치한다. 좌표를 선언하는 최소 alias POM을 함께 준다.

```xml
<!-- secret-lib-1.0.pom -->
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>acme</groupId>
  <artifactId>secret-lib</artifactId>
  <version>1.0</version>
</project>
```

```bash
mvn install:install-file \
  -Dmaven.repo.local=./localrepo \
  -Dfile=shim/target/shim.jar \
  -DpomFile=secret-lib-1.0.pom
```

이제 원래 프로젝트의 의존성 해석이 통과한다. (이 `acme` 블록은 원리를 보여 주는 설명용 구성이며 직접 실행한 결과는 아니다. 실제로 실행해 확인한 데이터는 같은 패턴을 적용한 대상 프로젝트 기록으로 12절에 있다.)

```bash
mvn -Dmaven.repo.local=./localrepo -DskipTests test-compile
```

핵심은 `-DpomFile`이 선언한 좌표(`acme:secret-lib:1.0`)로 JAR이 설치된다는 점이다. JAR 안의 클래스가 무엇이든 Maven은 좌표만 보고 의존성을 만족시킨다.

## 5. 실제 메커니즘: 좌표 alias 설치

규모가 커지면 한 가지 분리가 중요해진다. shim 모듈은 보통 **자기 고유의 로컬 좌표**로 빌드된다(예: `local.kcf:local-sfm5-core:0.1.0-SNAPSHOT`). 하지만 대상 프로젝트가 원하는 좌표는 `kpn.co.kr:sfm5-core:1.1.3`이다. 둘은 다르다. 그래서 같은 JAR을, 소비자가 원하는 **진짜 좌표로 다시 설치**하는 단계가 필요하다. 이때 쓰는 별도 POM이 *alias POM*이다.

왜 모듈을 처음부터 진짜 좌표(`kpn.co.kr:sfm5-core`)로 빌드하지 않을까? 두 가지 이유다. 첫째, 한 shim JAR을 소비자가 요구하는 *여러 버전*(1.1.3, 1.1.9, …)에 동시에 매핑해야 하는데, Maven 모듈 하나는 버전 하나만 가진다. 둘째, 내 reactor가 남의 회사 좌표를 자기 산출물 좌표로 주장하면 빌드 구성이 헷갈리고, 나중에 진짜 라이브러리가 같은 좌표로 들어오면 충돌한다. 그래서 모듈은 중립적 로컬 좌표로 *한 번* 빌드하고, 설치 단계에서 진짜 좌표(들)로 *여러 번* alias 한다.

alias POM은 두 가지를 한다.

1. JAR이 설치될 좌표(`groupId`/`artifactId`/`version`)를 선언한다.
2. 그 좌표가 가져야 할 **전이 의존성**을 선언한다. 소비자가 이 의존성을 통해 Spring·MyBatis·Gson 같은 공개 라이브러리를 함께 끌어오기 때문이다. 이걸 빠뜨리면 컴파일은 되다가 다른 클래스에서 "cannot find symbol"이 난다.

대상 프로젝트의 설치 스크립트에서 일반화한 핵심 함수는 이렇다.

```bash
install_alias() {
  local jar="$1" artifact="$2" version="$3" pom="$4"
  mvn org.apache.maven.plugins:maven-install-plugin:3.1.1:install-file \
    -Dmaven.repo.local="$LOCAL_REPO" \
    -Dfile="$jar" -DpomFile="$pom"
}

# 같은 shim JAR을, 소비자가 요구하는 여러 버전으로 동시에 alias
for v in 1.1.3 1.1.9 1.1.11 1.1.12; do
  install_alias "$SFM5_CORE_JAR" sfm5-core "$v" "$(write_pom sfm5-core "$v")"
done
```

(`write_pom`은 위 alias POM을 버전만 바꿔 찍어내는 헬퍼다. 한 프로젝트 안에서도 모듈마다 요구 버전이 달라 같은 JAR을 여러 버전으로 깔아야 할 때가 있다.)

설치가 끝나면 로컬 저장소에는 진짜 좌표 경로로 JAR이 자리잡는다(이번 작업에서 실제로 확인한 출력).

```text
.codex-m2/kpn/coocon/CcSecurity_jdk1.6/1.0.4/CcSecurity_jdk1.6-1.0.4.jar
.codex-m2/kpn/co/kr/sfm5-core/{1.1.3,1.1.9,1.1.11,...}/...
.codex-m2/kpn/co/kr/kcf-util/{1.0.6,1.0.9,1.1.3,1.1.7}/...
```

흐름을 한 줄로 따라가면: *shim 모듈 컴파일 → target에 JAR → `install-file`이 진짜 좌표 경로로 복사 + alias POM 기록 → 소비자 `mvn`이 좌표로 발견*. 소비자(대상 프로젝트)는 자기 `pom.xml`을 한 글자도 바꾸지 않은 채 빌드된다.

## 6. 규모를 키우는 자동화: reactor + 설치 스크립트

shim이 여러 개고 서로 의존하면(예: util shim이 core shim을 쓴다) 수작업 설치는 무너진다. 두 층으로 자동화한다.

1. **reactor 부모 POM**(reactor = 여러 모듈을 한 번에 빌드·해석하는 Maven 다중 모듈 빌드) — 모든 shim 모듈을 `<modules>`로 묶는 부모 POM. `mvn install` 한 번이면 Maven이 모듈 간 의존 순서를 계산해 전부 빌드하고 로컬 좌표로 설치한다. shim 사이 의존(util → core)은 각 모듈 POM에 평범한 의존성으로 선언하면 reactor가 순서를 잡는다.

    ```text
    shims/pom.xml         (packaging=pom, <modules> 나열)
    shims/core/pom.xml    (local.kcf:local-core)
    shims/util/pom.xml    (local.kcf:local-util  → depends local-core)
    shims/vendor-x/pom.xml
    ```

2. **설치 스크립트** — reactor 빌드 후, 각 모듈 JAR을 5절의 `install_alias`로 진짜 좌표(들)에 깐다. 스크립트는 보통 `~/.m2`가 아니라 프로젝트-로컬 저장소(`.codex-m2`)를 인자로 받아 거기에만 설치한다.

전체 셋업은 "shim 설치 → 빌드/테스트"의 사다리가 된다.

```bash
# 1) shim 빌드 + alias 설치 (한 번)
bash scripts/install-local-shims.sh "$PWD/.codex-m2"
# 2) 이후 일반 빌드 (.mvn/maven.config 가 .codex-m2 를 가리킴)
mvn -q -DskipTests=false test
```

함정 하나가 여기서 드러났다(실측). shim 모듈마다 컴파일 JDK 요건이 다를 수 있다. 한 모듈이 `maven.compiler.release=21`을 요구하면 JDK 17로는 빌드가 멈춘다.

```text
[ERROR] ... project local-sfm6-core: Fatal error compiling:
  error: release version 21 not supported
```

그래서 **shim 빌드 JDK와 본 프로젝트 빌드 JDK를 분리**했다. shim 설치는 JDK 21로 돌리고(상위 호환), 본 프로젝트 컴파일·테스트는 요건대로 JDK 17로 돌린다. 같은 머신에 여러 JDK를 두고 단계별로 `JAVA_HOME`을 바꿔 끼우면 된다.

## 7. 혼동하기 쉬운 인접 개념

- **shim 저장소 vs `<systemPath>`/system scope** — system scope는 느슨한 JAR 파일 하나를 절대경로로 가리킨다. 전이 의존성이 없고 깨지기 쉬워 사실상 폐기된 방식이다. shim 저장소는 POM을 가진 *제대로 된 저장소 아티팩트*라 전이 해석이 정상 동작한다.
- **shim vs mock** — Mockito mock은 *테스트 안*에서 객체 하나의 동작을 흉내 낸다. shim은 *프로젝트 전체(main+test)*의 컴파일·실행을 위한 의존성 자체의 대역이다. 층이 다르다. mock은 shim 위에서 쓰인다.
- **로컬 저장소 vs 오프라인 모드(`-o`)** — `mvn -o`는 이미 채워진 저장소에서 네트워크 접근만 막는다. 없는 아티팩트를 *만들어 주지는* 않는다. shim 저장소는 받을 수 없는 좌표를 *채우는* 작업이라, 둘은 목적이 다르다(보완 관계).
- **정답지 경계** — shim은 본 프로젝트 소스에 맞춰 적응해야 하고, 그 반대로 본 소스를 shim에 맞춰 고치면 안 된다. 본 소스는 운영 코드(정답지)이고 shim은 그걸 돌리기 위한 가설이다. 이 방향을 거꾸로 하면 로컬 편의를 위해 운영 동작을 훼손하게 된다. (관련: [`../testing/characterization_testing.md`](../testing/characterization_testing.md).)

## 8. 실패 모드와 함정

- **alias POM에 전이 의존성 누락** → 좌표는 해석되는데 그 라이브러리가 쓰던 Spring/MyBatis 등이 안 끌려와 다른 곳에서 컴파일 실패. alias POM에 필요한 전이 의존성을 선언해 막는다.
- **shim 표면 부족** → 본 프로젝트가 shim에 없는 메서드를 부르면 *본 프로젝트가* 컴파일 실패한다(테스트가 아니라). 이때는 테스트가 아니라 shim을 넓혀야 한다. 즉 본 소스가 갱신돼 새 API를 부르기 시작하면 shim도 따라 커진다.
- **JDK 버전 불일치** → 7절의 `release version N not supported`. 모듈별 요건에 맞는 JDK로 빌드.
- **동작 충실도** → 테스트가 의존성의 *출력*을 단언하는 경우(암호화·인코딩·체크섬), shim은 컴파일만이 아니라 진짜와 같은 결과를 내야 한다. 예: 테스트가 fixture를 `Aes256`으로 암호화해 넣고 본 코드가 복호화해 비교하면, shim의 암복호화가 진짜와 byte 단위로 같아야 통과한다.
- **per-project 포크 금지** → 같은 회사 라이브러리를 쓰는 프로젝트가 여럿이면 shim 스냅샷을 하나로 공유하고 재동기화한다. 프로젝트마다 따로 흉내 내면 표면이 갈라진다.

## 9. 검증 경로

- **좌표 해석 확인**: `mvn -Dmaven.repo.local=<repo> -DskipTests test-compile`
  - PASS = 컴파일 성공(모든 좌표 해석됨).
  - FAIL = `Could not resolve`(좌표 누락/오타) 또는 `cannot find symbol`(shim에 쓰이는 메서드 없음).
- **설치된 좌표 확인**: `<repo>/<group을 /로>/<artifact>/<version>/`에 `.jar`과 `.pom`이 있는지 본다.
- **shim JAR이 기대 클래스를 담았는지**: `jar tf <jar> | grep CcSecurityUtil` → `cc/checkpay/common/CcSecurityUtil.class`.
- **동작 충실도**: 전체 `mvn test`. shim 출력이 진짜와 다르면 여기서 단언 실패로 드러난다.

`확인해 보면 된다`는 검증이 아니다. 위처럼 *무엇을 실행하고, PASS/FAIL이 무엇인지*가 있어야 한다.

## 10. 다른 생태계로의 전이

핵심 원리(이름@버전을 *내가 통제하는 로컬 출처*로 만족)는 좌표 기반 빌드 어디에나 옮겨진다.

- **npm/yarn**: 사설 레지스트리(Verdaccio)에 같은 `name@version`을 publish하거나, `file:`/`link:` 의존성으로 대체.
- **pip**: 로컬 인덱스(`--index-url`)나 wheelhouse(`--find-links`)에 같은 패키지명·버전의 wheel을 둔다.
- **Go**: `replace` 지시어로 모듈 경로를 로컬로 돌리거나 `GOFLAGS=-mod=mod` + 로컬 모듈 캐시, `GOPROXY=off`.
- **Gradle**: `mavenLocal()` 저장소 + 위와 같은 alias 설치, 또는 composite build(`includeBuild`)로 좌표를 로컬 빌드로 치환.

공통 질문은 늘 같다 — "빌드가 요구하는 *이름@버전*을, 진짜를 못 받는 상황에서 무엇으로 채울 것인가."

## 11. 스스로 확인하기

- Maven이 의존성을 좌표로만 해석한다는 사실이, 왜 "가짜 JAR로 진짜 의존성을 만족"시키는 걸 가능하게 하나?
- shim 모듈의 빌드 좌표(`local.kcf:local-sfm5-core`)와 소비자가 원하는 좌표(`kpn.co.kr:sfm5-core:1.1.3`)가 다른데, 둘을 잇는 것은 무엇인가? 그게 없으면 무엇이 깨지나?
- shim 저장소와 `mvn -o`(offline)는 무엇이 다른가? 한쪽만으로 안 되는 상황은?
- 테스트가 어떤 의존성의 암호화 *출력값*을 단언한다면, 그 의존성 shim에서 컴파일 외에 무엇을 더 보장해야 하나?
- 본 프로젝트 소스가 갱신돼 shim에 없는 메서드를 호출하기 시작하면, 빌드는 어느 단계에서 깨지고 무엇을 고쳐야 하나(테스트? shim? 본 소스?)?

## 12. 확인한 사실과 추론

이번 대상 프로젝트 작업에서 **직접 확인한 사실**: `.mvn/maven.config`의 `-Dmaven.repo.local=.codex-m2` 리다이렉트로 전역 `~/.m2`를 건드리지 않고 빌드됨, shim JAR을 `install-file` + alias POM으로 진짜 좌표(`kpn.coocon:CcSecurity_jdk1.6:1.0.4` 등)에 설치하면 본 프로젝트가 미수정 상태로 `test-compile` 통과(RC 0), shim 모듈 중 하나의 `release 21` 요건 때문에 JDK 17 빌드가 멈춰 shim은 JDK 21·본 빌드는 JDK 17로 분리한 것, `jar tf`로 alias JAR에 기대 클래스가 들어 있음을 확인한 것.

**일반 원리·추론**: 좌표 기반 해석은 Maven의 설계 사실이지만, 위 npm/pip/Go/Gradle 전이는 각 도구의 동등 메커니즘에 기댄 추론이며 프로젝트마다 세부가 다르다. system scope가 "사실상 폐기"라는 평가도 일반 통념 수준의 판단이다. 실제 적용 시에는 해당 생태계의 1차 문서로 한 번 더 확인하는 편이 안전하다.

## 관련 노트

- 이 기법을 떠받치는 중립 원리 — [Maven 의존성 해석](./maven_dependency_resolution.md), [컴파일은 시그니처만 본다](./javac_classpath_and_release.md), [빌드 lifecycle](./maven_build_lifecycle.md)
- 망분리 환경의 개발·배포 아키텍처(이 기법이 들어맞는 큰 그림): [`../ai/closed_network_ai_arch.md`](../ai/closed_network_ai_arch.md), [`../ai/network_separated_ai_agent_cicd.md`](../ai/network_separated_ai_agent_cicd.md)
- Maven 일반 트러블슈팅: [`../troubleshooting/maven.md`](../troubleshooting/maven.md)
- 본 소스를 고치지 않고 현재 동작을 고정하는 테스트(정답지 경계와 같은 원리): [`../testing/characterization_testing.md`](../testing/characterization_testing.md)
