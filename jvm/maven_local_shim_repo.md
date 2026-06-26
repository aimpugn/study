# 받을 수 없는 비공개 Maven 의존성으로 묶인 Java 빌드를 로컬에서 되살리기

접근 권한이 없는 Maven 저장소에 있는 JAR에 의존하는 Java 프로젝트를 밖으로 가져오면, `mvn`은 소스 컴파일 전에 멈춘다. 예를 들어 프로젝트가 `com.example.legacy:workflow-core:1.1.3`이나 `com.example.security:crypto-legacy-jdk6:1.0.4` 같은 비공개 좌표를 요구하는데 현재 PC나 CI가 그 저장소에 닿을 수 없다면, Maven은 해당 좌표를 해석하지 못한다. 그러면 컴파일도, 테스트도, 리팩터링 검증도 시작할 수 없다.

이 문서는 그 상황에서 **원래 `pom.xml`의 의존성 좌표는 유지하면서**, 프로젝트 전용 로컬 Maven 저장소에 대체 JAR을 설치해 빌드를 되살리는 방법을 설명한다. 핵심은 Maven이 의존성을 `groupId:artifactId:version`이라는 좌표로 찾는다는 점과, Java 컴파일러가 컴파일 단계에서 구현 내용이 아니라 클래스와 메서드의 시그니처를 본다는 점이다.

읽고 나면 다음 내용을 자기 말로 설명할 수 있어야 한다.

- 왜 대체 JAR이 원래 의존성 좌표를 만족시킬 수 있는가
- alias POM이 왜 별도로 필요한가
- shim을 어디까지 구현해야 하는가: 컴파일 표면과 런타임 동작의 경계
- offline Maven 빌드가 깨질 때 어느 층을 먼저 확인해야 하는가
- 이 방법이 `systemPath`, mock, 오프라인 모드와 어떻게 다른가

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
- [13. 실제 Windows/Nushell 런북: public repository 없이 `mvn` 돌리기](#13-실제-windowsnushell-런북-public-repository-없이-mvn-돌리기)
- [관련 노트](#관련-노트)

## 1. 한 줄 원리

**Maven은 의존성을 좌표(coordinate), 즉 `groupId:artifactId:version`으로 해석한다.** 저장소의 그 좌표 위치에 JAR과 POM이 있으면, Maven은 그 의존성이 존재한다고 본다. 그래서 내가 만든 JAR이라도 `com.example.security:crypto-legacy-jdk6:1.0.4`라는 좌표로 프로젝트 전용 로컬 저장소에 설치하면, 그 좌표를 요구하는 `pom.xml`은 해석을 통과한다.

이 사실이 문제를 바꾼다. "접근할 수 없는 저장소에서 JAR을 받아와야 한다"가 아니라, "빌드가 요구하는 좌표를 현재 내가 통제하는 저장소에 어떤 아티팩트로 채울 것인가"가 된다.

여기서 **shim은 원래 라이브러리 전체가 아니라, 현재 프로젝트가 실제로 컴파일하거나 테스트하는 API 표면만 제공하는 대체 모듈**이다. 전체 라이브러리를 복제하는 것이 아니라, 빌드와 검증을 진행하기 위한 최소한의 Java 클래스, 메서드, 상수, 예외 타입을 제공한다.

## 2. 문제는 어떻게 나타나나

대상 Java 프로젝트의 `pom.xml`이 다음처럼 접근 제한 저장소에만 있는 좌표를 의존성으로 선언한다고 하자.

```xml
<dependency>
    <groupId>com.example.legacy</groupId>
    <artifactId>workflow-core</artifactId>
    <version>1.1.3</version>
</dependency>
<dependency>
    <groupId>com.example.security</groupId>
    <artifactId>crypto-legacy-jdk6</artifactId>
    <version>1.0.4</version>
</dependency>
```

현재 PC가 그 저장소에 접근하지 못하면 Maven은 의존성 해석 단계에서 멈춘다.

```text
[ERROR] Failed to execute goal ... on project app:
  Could not resolve dependencies for project ...:
  com.example.security:crypto-legacy-jdk6:jar:1.0.4 was not found in
  https://repo.maven.apache.org/maven2 ...
```

중요한 점은 이것이 컴파일 오류가 아니라 **컴파일 이전의 의존성 해석 실패**라는 사실이다. Java 소스는 아직 `javac`까지 가지 못했다. 따라서 소스 코드를 먼저 고치는 방식으로는 해결되지 않는다. Maven이 요구하는 좌표를 현재 빌드가 볼 수 있는 저장소에 채워야 한다.

원래 `pom.xml`의 좌표를 지우거나 공개 라이브러리 좌표로 바꾸는 방식은 보통 좋지 않다. 그 좌표는 운영 환경이나 원래 배포 환경에서 실제 의존성을 가리키는 계약이기 때문이다. 로컬 빌드를 위해 계약 자체를 바꾸면, 나중에 원래 저장소가 연결됐을 때 같은 빌드인지 검증하기 어렵다. 그래서 좌표는 그대로 두고, **로컬 저장소의 같은 좌표 자리에 대체 JAR을 설치**하는 쪽이 더 안전하다.

## 3. shim을 만들기 전에 알아야 할 세 가지

shim은 추측으로 만들면 안 된다. 대상 프로젝트가 실제로 요구하는 Maven 좌표와 Java 시그니처를 먼저 좁혀야 한다.

1. **어떤 좌표를 요구하는가**

    `pom.xml`, 부모 POM, `dependencyManagement`, profile을 읽어 접근할 수 없는 좌표 목록을 모은다. 같은 라이브러리라도 모듈마다 다른 버전을 요구하면, 그 버전도 각각 로컬 저장소에 재현해야 한다.

2. **그 라이브러리의 무엇을 실제로 호출하는가**

    전체 API가 아니라 쓰이는 표면만 필요하다. 시작은 import 검색이고, 그다음은 `mvn test-compile`이 알려 주는 `cannot find symbol`을 따라간다. 예를 들어 프로젝트가 보안 라이브러리에서 아래 세 메서드만 호출한다고 하자.

    ```java
    import com.example.security.LegacyCrypto;

    String encrypted = LegacyCrypto.encryptBase64(message, key, true);
    String mac = LegacyCrypto.hmacSha256(body, key, true);
    String plain = LegacyCrypto.decryptBase64(encrypted, key, "UTF-8", true);
    ```

    그러면 shim의 `LegacyCrypto`는 우선 이 세 메서드의 시그니처를 제공해야 한다. 컴파일을 다시 돌려 다음 빠진 타입이나 메서드를 확인하고, 더 이상 빠진 심볼이 없을 때까지 shim 표면을 넓힌다. 이 반복은 "원래 라이브러리를 복제하는 과정"이 아니라 "현재 프로젝트가 요구하는 계약을 관측하는 과정"이다.

3. **Maven이 어떤 로컬 저장소를 보는가**

    Maven의 기본 로컬 저장소는 `~/.m2/repository`다. 프로젝트마다 저장소를 분리하려면 프로젝트 루트의 `.mvn/maven.config`에 다음처럼 적는다.

    ```text
    -Dmaven.repo.local=.project-m2
    ```

    이렇게 하면 그 프로젝트에서 실행하는 `mvn`은 전역 `~/.m2` 대신 저장소 루트의 `.project-m2`를 로컬 저장소로 쓴다. shim도 이 저장소에 설치하면 원래 `pom.xml`을 바꾸지 않아도 의존성 해석이 통과한다.

컴파일과 런타임의 경계도 분명히 해야 한다. Java 컴파일러는 메서드 본문이 아니라 시그니처를 보고 컴파일한다. 따라서 컴파일만 목표라면 쓰이는 타입과 메서드 모양을 제공하면 된다. 그러나 테스트가 암호화 결과, 체크섬, 인코딩 결과처럼 **의존성의 출력값을 단언**한다면 shim은 그 동작까지 충실히 구현해야 한다. 반대로 외부 장비, 네트워크, 운영 환경이 필요한 기능은 가짜 성공값을 돌려주기보다 명확한 예외로 fail-fast 하는 편이 안전하다. 허구의 성공값은 테스트를 잘못된 정답에 맞춰 버린다.

## 4. 가장 작은 동작 예제

원리를 최소 형태로 보자. 어떤 프로젝트가 `com.example.private:secret-lib:1.0`에 의존하고 `com.example.privateapi.Crypto.enc(String)` 하나만 호출한다고 하자.

먼저 shim 소스를 만든다.

```java
// shim/src/main/java/com/example/privateapi/Crypto.java
package com.example.privateapi;

import java.nio.charset.StandardCharsets;
import java.util.Base64;

public final class Crypto {
    private Crypto() {
    }

    public static String enc(String value) {
        return Base64.getEncoder()
                .encodeToString(value.getBytes(StandardCharsets.UTF_8));
    }
}
```

이 소스를 JAR로 빌드한 뒤, 소비자가 원하는 좌표를 가진 POM을 함께 준비한다.

```xml
<!-- secret-lib-1.0.pom -->
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example.private</groupId>
  <artifactId>secret-lib</artifactId>
  <version>1.0</version>
</project>
```

그다음 `install-file`로 프로젝트 전용 로컬 저장소에 설치한다.

```bash
mvn org.apache.maven.plugins:maven-install-plugin:3.1.4:install-file \
  -Dmaven.repo.local=./.project-m2 \
  -Dfile=shim/target/secret-lib-shim.jar \
  -DpomFile=secret-lib-1.0.pom
```

이제 원래 프로젝트의 의존성 해석은 해당 좌표를 찾을 수 있다.

```bash
mvn -Dmaven.repo.local=./.project-m2 -DskipTests test-compile
```

이 예제에서 중요한 것은 `-DpomFile`이 선언한 좌표다. JAR 파일 이름이 `secret-lib-shim.jar`이어도, 설치된 좌표는 POM의 `com.example.private:secret-lib:1.0`이다. Maven은 소비자 `pom.xml`의 좌표와 로컬 저장소의 좌표가 일치하는지를 본다.

## 5. 실제 메커니즘: 좌표 alias 설치

규모가 커지면 shim 모듈의 빌드 좌표와 소비자가 요구하는 좌표를 분리하는 편이 좋다.

- shim 모듈 자체의 좌표: `local.shims:workflow-core-shim:0.1.0-SNAPSHOT`
- 소비자가 요구하는 좌표: `com.example.legacy:workflow-core:1.1.3`

둘은 같은 JAR을 가리킬 수 있지만, 의미는 다르다. shim 모듈의 좌표는 "내가 빌드하는 대체 구현"을 나타내고, 소비자 좌표는 "원래 프로젝트가 요구하는 의존성 계약"을 나타낸다. 이 둘을 연결하는 파일이 **alias POM**이다.

모듈을 처음부터 `com.example.legacy:workflow-core`로 빌드하지 않는 이유는 두 가지다.

1. 한 shim JAR을 소비자가 요구하는 여러 버전, 예를 들어 `1.1.3`, `1.1.9`, `1.1.12`에 동시에 매핑해야 할 수 있다. Maven 모듈 하나는 한 번에 버전 하나만 가진다.
2. 내 로컬 reactor가 외부 좌표를 자기 산출물 좌표처럼 주장하면, 나중에 실제 의존성이 같은 좌표로 들어왔을 때 어느 것이 진짜인지 헷갈리기 쉽다.

따라서 shim은 중립적인 로컬 좌표로 한 번 빌드하고, 설치 단계에서 소비자 좌표로 여러 번 alias 하는 편이 더 분명하다.

alias POM은 두 역할을 한다.

1. JAR이 설치될 좌표, 즉 `groupId`, `artifactId`, `version`을 선언한다.
2. 그 좌표가 소비자에게 전이시켜야 하는 공개 의존성을 선언한다. 예를 들어 원래 라이브러리의 POM이 Jackson, Spring, MyBatis 같은 공개 라이브러리를 전이 의존성으로 제공했다면, alias POM도 필요한 범위에서 그 정보를 담아야 한다.

설치 스크립트의 핵심은 다음처럼 단순해질 수 있다.

```bash
install_alias() {
  local jar="$1" pom="$2"

  mvn org.apache.maven.plugins:maven-install-plugin:3.1.4:install-file \
    -Dmaven.repo.local="$LOCAL_REPO" \
    -Dfile="$jar" \
    -DpomFile="$pom"
}

# 같은 shim JAR을 소비자가 요구하는 여러 버전으로 alias
for version in 1.1.3 1.1.9 1.1.12; do
  install_alias "$WORKFLOW_CORE_JAR" "$(write_alias_pom workflow-core "$version")"
done
```

설치가 끝나면 로컬 저장소에는 소비자 좌표 경로로 JAR과 POM이 놓인다.

```text
.project-m2/com/example/security/crypto-legacy-jdk6/1.0.4/crypto-legacy-jdk6-1.0.4.jar
.project-m2/com/example/legacy/workflow-core/1.1.3/workflow-core-1.1.3.jar
.project-m2/com/example/legacy/workflow-core/1.1.9/workflow-core-1.1.9.jar
.project-m2/com/example/legacy/legacy-util/1.0.6/legacy-util-1.0.6.jar
```

흐름을 한 줄로 따라가면 다음과 같다.

```text
shim 소스 -> shim JAR -> install-file + alias POM -> 로컬 저장소의 원래 좌표 -> 소비자 Maven 빌드
```

소비자 프로젝트는 자기 `pom.xml`을 유지한 채 빌드된다. 바뀐 것은 원격에서 받지 못하던 좌표를 프로젝트 전용 로컬 저장소에 미리 채운 점이다.

## 6. 규모를 키우는 자동화: reactor + 설치 스크립트

shim이 여러 개이고 서로 의존하면 수작업 설치는 금방 깨진다. 이때는 두 층으로 나누는 편이 안정적이다.

1. **reactor 부모 POM**

    reactor는 여러 Maven 모듈을 한 번에 빌드하는 다중 모듈 구조다. 부모 `pom.xml`이 `<modules>`로 shim 모듈을 묶으면, Maven이 모듈 간 의존 순서를 계산한다.

    ```text
    shims/pom.xml                  (packaging=pom, <modules> 나열)
    shims/workflow-core/pom.xml    (local.shims:workflow-core-shim)
    shims/legacy-util/pom.xml      (local.shims:legacy-util-shim -> workflow-core-shim 의존)
    shims/crypto/pom.xml           (local.shims:crypto-legacy-shim)
    ```

    부모에서 `mvn install`을 실행하면 각 shim JAR이 로컬 좌표로 만들어진다.

2. **alias 설치 스크립트**

    reactor 빌드가 끝난 뒤, 각 shim JAR을 소비자 좌표로 프로젝트 전용 로컬 저장소에 설치한다.

    ```bash
    bash scripts/install-local-shims.sh "$PWD/.project-m2"
    mvn -q -DskipTests=false test
    ```

JDK 버전도 별도로 확인해야 한다. 예를 들어 shim 중 하나가 `maven.compiler.release=21`을 요구하면 JDK 17로는 컴파일이 멈춘다.

```text
[ERROR] Fatal error compiling:
  error: release version 21 not supported
```

이 경우에는 shim 빌드는 JDK 21로 실행하고, 본 프로젝트가 Java 17 target을 요구한다면 본 프로젝트 빌드는 `--release 17` 또는 Maven compiler 설정에 맞춰 돌린다. 상위 JDK는 낮은 release target을 만들 수 있지만, 낮은 JDK는 높은 release target을 만들 수 없다. 단계별 `JAVA_HOME`을 명시하는 이유가 여기에 있다.

## 7. 혼동하기 쉬운 인접 개념

- **shim 저장소 vs `<systemPath>`/`system` scope**

    `system` scope는 로컬 파일 경로의 JAR을 직접 가리킨다. Maven POM Reference는 접근할 수 없는 JAR을 다루는 방법 중 로컬 설치를 가장 단순한 권장 방법으로 제시하고, `systemPath` 방식은 권장하지 않는다고 설명한다. shim 저장소 방식은 JAR과 POM을 가진 정상 Maven 아티팩트를 로컬 저장소에 설치하므로, 전이 의존성과 좌표 추적이 더 자연스럽다.

- **shim vs mock**

    mock은 테스트 안에서 객체 하나의 동작을 흉내 낸다. shim은 프로젝트 전체의 컴파일 클래스패스와 테스트 런타임 클래스패스를 채우는 의존성 대역이다. 층이 다르다. mock은 shim 위에서 쓸 수 있지만, Maven이 의존성을 해석하지 못하는 문제를 mock만으로 해결할 수는 없다.

- **로컬 저장소 vs 오프라인 모드(`-o`)**

    로컬 저장소는 Maven이 아티팩트를 찾는 디렉터리다. 오프라인 모드는 원격 저장소 접근을 막고 이미 로컬에 있는 것만 쓰게 하는 실행 모드다. `-o`는 없는 좌표를 만들어 주지 않는다. 먼저 shim이나 공개 의존성을 로컬 저장소에 채운 뒤 `-o`로 검증해야 한다.

- **컴파일 성공 vs 동작 충실도**

    컴파일 성공은 필요한 시그니처가 있다는 뜻이다. 테스트가 실제 결과값을 검증한다면 shim 구현도 그 결과값을 맞춰야 한다. 반대로 테스트가 닿지 않는 외부 부수효과는 임의 성공보다 명확한 예외가 낫다.

## 8. 실패 모드와 함정

- **alias POM에 전이 의존성 누락**

    원래 라이브러리 POM이 제공하던 공개 의존성이 빠지면, 좌표 자체는 해석되더라도 다른 클래스에서 `cannot find symbol`이 날 수 있다. 이때는 shim JAR만 볼 것이 아니라 alias POM의 `<dependencies>`도 확인한다.

- **shim 표면 부족**

    대상 프로젝트가 shim에 없는 클래스, 메서드, 필드, 예외 타입을 참조하면 컴파일 단계에서 실패한다. 이때 고쳐야 할 곳은 보통 대상 프로젝트 소스가 아니라 shim 표면이다.

- **JDK 버전 불일치**

    `release version N not supported`는 현재 Maven이 사용하는 JDK가 `N`보다 낮다는 신호다. `mvn --version`으로 Maven이 실제로 잡은 Java 버전을 확인한다.

- **동작 충실도 과소평가**

    암호화, 인코딩, 체크섬, 날짜 계산처럼 출력값이 테스트의 판단 대상이면 shim이 컴파일 표면만 제공해서는 부족하다. 반대로 외부 시스템 호출처럼 로컬에서 의미 있게 재현할 수 없는 경로는 테스트가 호출했을 때 즉시 식별 가능한 예외를 내는 편이 안전하다.

- **프로젝트별 shim 포크**

    같은 비공개 API를 여러 프로젝트가 사용한다면, 가능한 한 shim 소스는 하나의 공통 스냅샷으로 유지하고 프로젝트별 alias만 다르게 둔다. 프로젝트마다 독립적으로 흉내 내면 같은 API 표면이 조금씩 갈라져서 이후 검증 비용이 커진다.

## 9. 검증 경로

- **좌표 해석 확인**

    ```bash
    mvn -Dmaven.repo.local=.project-m2 -DskipTests test-compile
    ```

    PASS는 main/test 컴파일 성공이다. FAIL은 보통 `Could not resolve` 또는 `cannot find symbol`이다. 전자는 좌표나 저장소 문제이고, 후자는 shim 시그니처 문제다.

- **설치된 좌표 확인**

    ```bash
    find .project-m2/com/example -maxdepth 5 -type f | sort | head
    ```

    PASS는 요구한 `groupId/artifactId/version` 경로 아래 `.jar`과 `.pom`이 함께 있는 것이다.

- **shim JAR 내용 확인**

    ```bash
    jar tf shims/crypto/target/crypto-legacy-shim.jar | grep 'LegacyCrypto.class'
    ```

    PASS는 소비자 코드가 import하는 클래스 경로가 JAR 안에 보이는 것이다.

- **의존성 그래프 확인**

    ```bash
    mvn -Dmaven.repo.local=.project-m2 dependency:tree
    ```

    PASS는 접근 제한 좌표가 로컬 저장소에서 해석되고, alias POM이 의도한 전이 의존성을 제공하는 것이다.

- **동작 충실도 확인**

    ```bash
    mvn -Dmaven.repo.local=.project-m2 test
    ```

    PASS는 테스트가 기대하는 계산 결과와 예외 경계가 통과하는 것이다. 컴파일은 통과했는데 테스트가 실패하면, shim의 런타임 동작이 부족하거나 테스트가 실제 외부 환경을 요구하는지 구분한다.

## 10. 다른 생태계로의 전이

핵심 원리, 즉 "빌드가 요구하는 이름과 버전을 내가 통제하는 로컬 출처로 만족시킨다"는 Maven 밖으로도 옮겨진다. 다만 각 생태계의 공식 메커니즘은 다르므로 실제 적용 전에는 해당 도구의 1차 문서를 확인해야 한다.

- **npm/yarn**: 사설 레지스트리에 같은 `name@version`을 publish하거나, `file:` 의존성으로 로컬 패키지를 연결한다.
- **pip**: 로컬 wheelhouse를 `--find-links`로 지정하거나, 같은 패키지명과 버전의 wheel을 내부 인덱스에 둔다.
- **Go**: `replace` 지시어로 모듈 경로를 로컬 모듈로 돌린다.
- **Gradle**: Maven 로컬 저장소를 `maven { url = uri(...) }`로 지정하거나 composite build로 대체 구현을 연결한다.

공통 질문은 하나다. "빌드가 요구하는 이름과 버전을, 현재 접근 가능한 출처에서 무엇으로 채울 것인가?"

## 11. 스스로 확인하기

아래 질문에 답할 수 있으면 이 문서의 핵심을 잡은 것이다.

- Maven이 의존성을 좌표로 해석한다는 사실이 왜 shim 로컬 저장소를 가능하게 하나?
- shim 모듈의 빌드 좌표 `local.shims:workflow-core-shim`과 소비자 좌표 `com.example.legacy:workflow-core:1.1.3`를 잇는 것은 무엇인가?
- `mvn -o`와 `install:install-file`은 각각 어떤 문제를 풀고, 어떤 문제를 풀지 못하나?
- `cannot find symbol`과 `Could not resolve dependencies`는 각각 어느 층의 실패인가?
- 테스트가 암호화 결과값을 단언한다면 shim은 컴파일 표면 외에 무엇을 더 보장해야 하나?
- 외부 시스템 호출을 로컬 shim에서 무조건 성공시켰을 때 어떤 잘못된 테스트 신뢰가 생길 수 있나?

## 12. 확인한 사실과 추론

공식 Maven 문서로 직접 확인한 사실은 다음과 같다.

- Maven POM Reference는 의존성이 Maven 좌표로 기술되며, 중앙 저장소에서 받을 수 없는 JAR을 다루는 방법으로 Maven Install Plugin을 이용한 로컬 설치를 제시한다. 같은 문서에서 `systemPath` 방식은 권장하지 않는다고 설명한다. <https://maven.apache.org/pom.html>
- Maven Install Plugin의 `install-file` goal은 파일을 로컬 저장소에 설치하며, `pomFile`을 통해 main artifact와 함께 설치할 POM을 지정할 수 있다. <https://maven.apache.org/plugins/maven-install-plugin/install-file-mojo.html>
- Maven의 3rd party JAR 설치 가이드는 public repository에 없는 JAR을 로컬 저장소의 올바른 위치에 넣어 Maven이 찾게 만드는 흐름을 설명한다. <https://maven.apache.org/guides/mini/guide-3rd-party-jars-local.html>
- Maven 설정 문서는 `.mvn/maven.config`가 프로젝트 최상위 `.mvn` 디렉터리에 놓이며, `mvn` 명령줄 옵션을 프로젝트별로 담을 수 있다고 설명한다. <https://maven.apache.org/configure.html>
- Maven dependency mechanism 문서는 dependency scope가 transitivity와 classpath 포함 시점을 제한한다고 설명한다. <https://maven.apache.org/guides/introduction/introduction-to-dependency-mechanism.html>

이 문서의 `com.example...` 좌표, Windows 경로, 테스트 클래스명은 실제 조직이나 제품을 가리키지 않는 재현용 예시다. 실제 프로젝트에 적용할 때는 현재 `pom.xml`, `.mvn/maven.config`, `mvn --version`, `dependency:tree`, 로컬 저장소 경로를 직접 확인해야 한다.

다른 생태계로의 전이는 Maven의 동작과 유사한 "이름과 버전의 로컬 충족" 원리를 설명한 추론이다. npm, pip, Go, Gradle에서는 각 도구의 공식 문서와 현재 프로젝트 설정을 다시 확인해야 한다.

## 13. 실제 Windows/Nushell 런북: public repository 없이 `mvn` 돌리기

이 절은 위 원리를 Windows 개발 PC의 실제 명령 순서로 옮긴 것이다. 대상은 접근 제한 Maven 좌표를 가진 일반 Java 프로젝트다. 프로젝트는 `.mvn/maven.config`로 프로젝트 전용 로컬 저장소를 지정하고, `tools/local-shims` 같은 shim 모듈을 그 저장소에 설치해 public repository 없이도 컴파일과 일부 테스트를 돌린다.

전체 구조는 다음 한 문장으로 잡을 수 있다. `mvn`은 실행될 때 `JAVA_HOME`의 JDK를 쓰고, `.mvn/maven.config`의 `-Dmaven.repo.local=.project-m2`를 읽어 로컬 저장소 위치를 정하고, `-o`가 붙으면 원격 저장소를 보지 않고 `.project-m2` 안의 아티팩트만 사용한다. 따라서 `.project-m2`에 shim과 필요한 공개 의존성 캐시가 충분히 들어 있으면 public repository 없이도 빌드가 된다.

### 13.1 현재 PC에서 확인할 구성요소

아래 경로는 예시다. 다른 PC에서는 Maven과 JDK 설치 위치가 다르므로 먼저 파일 탐색기나 `ls`로 존재를 확인한다.

```text
C:\dev\apache-maven-3.9.9\bin\mvn.cmd
C:\dev\jdk\21.0.11+10
C:\Users\me\Workspace\private-java-app\.mvn\maven.config
C:\Users\me\Workspace\private-java-app\.project-m2
```

`.mvn/maven.config`에는 보통 다음처럼 프로젝트 전용 로컬 저장소 옵션을 둔다.

```text
-Dmaven.repo.local=.project-m2
```

이 한 줄 때문에 `mvn test`를 실행할 때마다 전역 `~/.m2/repository`가 아니라 프로젝트 루트의 `.project-m2`를 먼저 본다. Maven 실행 파일이 Apache Maven 단독 설치가 아니라 mvnd 번들 안에 들어 있는 경우도 있다. 그때도 `mvn --version`이 Apache Maven 버전과 Java 버전을 제대로 출력하면 같은 방식으로 쓸 수 있다.

### 13.2 Nushell에서 세션 환경 잡기

Nushell에서는 환경변수를 `$env.NAME = value`로 둔다. Windows의 `$env.PATH`는 보통 경로 리스트처럼 다룰 수 있으므로, 현재 PATH 앞에 JDK와 Maven bin을 추가한다.

```nu
cd C:\Users\me\Workspace\private-java-app

$env.JAVA_HOME = 'C:\dev\jdk\21.0.11+10'

let java_bin = ($env.JAVA_HOME | path join 'bin')
let maven_bin = 'C:\dev\apache-maven-3.9.9\bin'

$env.PATH = ($env.PATH | prepend $maven_bin | prepend $java_bin)
```

그다음 Maven이 실제로 어느 Java와 Maven으로 뜨는지 확인한다.

```nu
mvn --version
open .mvn/maven.config
ls .project-m2 | first 5
```

PASS 신호는 세 가지다. `mvn --version`이 Apache Maven 버전을 출력하고, Java 버전이 의도한 JDK로 보이며, `.mvn/maven.config`에서 `-Dmaven.repo.local=.project-m2`가 보이면 된다. `.project-m2`가 비어 있으면 offline 빌드는 실패할 수 있다. 이 경우 먼저 shim 설치와 필요한 공개 의존성 캐시 준비가 필요하다.

### 13.3 public repository 접근을 막고 빌드 확인하기

public repository를 실제로 쓰지 않는지 확인하려면 `-o`를 붙인다. `-o`는 offline mode다. Maven은 원격 저장소를 보지 않고 로컬 저장소에 이미 있는 아티팩트만 사용한다.

```nu
mvn -o -q '-DskipTests' test
```

이 명령은 테스트 실행을 건너뛰되 main/test 컴파일은 유지하는 compile gate로 쓸 수 있다. `-DskipTests`는 테스트 실행만 건너뛰고 테스트 컴파일은 남긴다. 테스트 컴파일까지 건너뛰는 `-Dmaven.test.skip=true`와 다르다.

특정 JUnit만 돌릴 때도 같은 원리다.

```nu
mvn -o '-Dtest=CryptoShimCompatibilityTest' test
```

여러 테스트 클래스를 한 번에 지정할 수도 있다.

```nu
mvn -o '-Dtest=CryptoShimCompatibilityTest,WorkflowCoreCompileTest,OfflineRepositorySmokeTest' test
```

PASS는 `BUILD SUCCESS`와 기대한 테스트 개수의 성공이다. FAIL이 `Cannot access central in offline mode`라면 로컬 저장소에 필요한 아티팩트가 빠진 것이다. FAIL이 테스트 assertion이라면 의존성 해석은 통과했지만 shim의 런타임 동작이나 테스트 전제가 부족한 것이다.

### 13.4 설치 스크립트는 무엇을 준비하나

`scripts/install-local-shims.sh` 같은 스크립트는 보통 두 가지를 한다.

1. `tools/local-shims` 같은 shim 소스를 빌드한다.
2. 그 JAR을 `install-file`과 alias POM으로 `.project-m2`의 소비자 좌표 경로에 설치한다.

즉 설치 스크립트는 Maven 자체를 설치하는 스크립트가 아니다. Maven이 나중에 접근 제한 좌표를 찾을 수 있도록 프로젝트 전용 로컬 저장소를 채우는 스크립트다. `mvn` 실행 파일이 PATH에 없으면 설치 스크립트도 실패한다. 스크립트 안에서도 결국 `mvn install` 또는 `mvn install:install-file`을 호출하기 때문이다.

Windows에서 실행 순서는 다음처럼 잡는 편이 안전하다.

```nu
cd C:\Users\me\Workspace\private-java-app

$env.JAVA_HOME = 'C:\dev\jdk\21.0.11+10'
let java_bin = ($env.JAVA_HOME | path join 'bin')
let maven_bin = 'C:\dev\apache-maven-3.9.9\bin'
$env.PATH = ($env.PATH | prepend $maven_bin | prepend $java_bin)

mvn --version
bash scripts/install-local-shims.sh .project-m2
mvn -o -q '-DskipTests' test
```

JDK는 프로젝트의 가장 높은 컴파일 요구에 맞춰 고른다. shim 모듈 중 하나가 release 21을 요구하면 JDK 21 이상이 필요하다. 대상 프로젝트 자체가 Java 17 target이라도 JDK 21은 Java 17 target을 컴파일할 수 있다. 반대로 JDK 17로 release 21 모듈을 컴파일할 수는 없다.

### 13.5 실패 신호를 원인별로 읽기

- `mvn: command not found` 또는 `The term 'mvn' is not recognized`

    Maven 실행 파일이 PATH에 없다. `C:\dev\apache-maven-3.9.9\bin` 또는 현재 PC의 Maven bin 경로를 PATH 앞에 추가한다.

- `release version 21 not supported`

    Maven이 낮은 JDK로 컴파일하고 있다. `JAVA_HOME`을 JDK 21 이상으로 바꾸고, `JAVA_HOME\bin`을 PATH 앞쪽에 둔다. `mvn --version`으로 Maven이 실제로 쓰는 Java를 확인한다.

- `Could not resolve dependencies ... was not found in ...`

    Maven은 실행됐지만 필요한 좌표가 로컬 저장소나 접근 가능한 원격 저장소에 없다. `.mvn/maven.config`가 `.project-m2`를 가리키는지 확인하고, shim 설치 스크립트가 해당 좌표의 `.jar`과 `.pom`을 설치했는지 본다.

- `Cannot access central in offline mode` 또는 `Cannot access ... in offline mode`

    `-o` 때문에 원격을 보지 못하는데 로컬 저장소에도 해당 아티팩트가 없다. public repository 없이 빌드하려면 그 좌표를 `.project-m2`에 미리 캐시하거나 shim으로 설치해야 한다.

- `scripts/install-local-shims.sh: line 2: set: pipefail\r: invalid option name`

    스크립트 파일의 줄끝이 CRLF라 bash가 `\r`까지 옵션 이름으로 읽은 것이다. 저장소가 LF를 요구한다면 스크립트를 LF로 정규화한다. 이 오류는 Maven 의존성 문제가 아니라 shell script 파일 형식 문제다.

이 실패 신호들은 서로 다른 층위에 있다. PATH 문제는 운영체제가 `mvn` 실행 파일을 찾지 못하는 문제이고, JDK 문제는 Maven이 어떤 `javac`를 쓰는지의 문제이며, `Could not resolve`는 Maven의 좌표와 저장소 문제이고, CRLF는 bash가 스크립트 문법을 읽기 전에 줄끝에서 넘어진 문제다. 같은 "빌드 안 됨"이라도 층위를 분리해야 빨리 고친다.

## 관련 노트

- 이 기법을 떠받치는 중립 원리: [Maven 의존성 해석](./maven_dependency_resolution.md), [컴파일은 시그니처만 본다](./javac_classpath_and_release.md), [빌드 lifecycle](./maven_build_lifecycle.md)
- Maven 일반 트러블슈팅: [`../troubleshooting/maven.md`](../troubleshooting/maven.md)
- 본 소스를 고치지 않고 현재 동작을 고정하는 테스트: [`../testing/characterization_testing.md`](../testing/characterization_testing.md)
