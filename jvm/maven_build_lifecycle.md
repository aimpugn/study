# Maven 빌드 lifecycle — 페이즈, 플러그인, 그리고 테스트가 실제로 도는 법

`mvn test` 한 줄은 "테스트를 돌린다"가 전부가 아니다. Maven은 정해진 **페이즈(phase)의 순서열**을 실행하고, 각 페이즈에 묶인 플러그인 goal이 실제 일을 한다. 이 구조를 알면 왜 의존성 해석이 컴파일보다 먼저 일어나는지, 왜 `test-compile`과 `test`가 따로인지, surefire가 테스트를 *별도 JVM에서* 어떻게 돌리고 `-D`가 왜 안 먹을 때가 있는지, 프로파일이 어떻게 환경별로 빌드를 바꾸는지가 한 번에 설명된다. 특정 프로젝트와 무관한 일반 메커니즘만 다룬다.

읽고 나면 다음을 스스로 설명할 수 있어야 한다 — 페이즈가 누적 실행된다는 말의 의미, 페이즈와 plugin goal의 관계, surefire의 fork와 `argLine`, 그리고 프로파일 활성화 방식.

## 목차

- [1. 직답](#1-직답)
- [2. 기본 lifecycle: 페이즈의 순서열](#2-기본-lifecycle-페이즈의-순서열)
- [3. 페이즈는 이름, 일은 plugin goal이 한다](#3-페이즈는-이름-일은-plugin-goal이-한다)
- [4. surefire가 테스트를 도는 법](#4-surefire가-테스트를-도는-법)
- [5. 플러그인 설정과 pluginManagement](#5-플러그인-설정과-pluginmanagement)
- [6. 프로파일과 활성화](#6-프로파일과-활성화)
- [7. 혼동하기 쉬운 점](#7-혼동하기-쉬운-점)
- [8. 실패와 검증](#8-실패와-검증)
- [9. 스스로 확인하기](#9-스스로-확인하기)
- [10. 확인한 사실과 추론](#10-확인한-사실과-추론)
- [관련 노트](#관련-노트)

## 1. 직답

Maven 빌드는 페이즈의 **순서열**이다. `mvn <페이즈>`는 그 페이즈*까지*의 모든 이전 페이즈를 순서대로 실행한다(누적). 각 페이즈 자체는 빈 단계 이름일 뿐이고, 실제 작업은 그 페이즈에 *바인딩된* 플러그인 goal이 한다. 그래서 같은 `mvn test`라도 "테스트만"이 아니라 "검증→의존성 해석→컴파일→테스트 컴파일→테스트"를 차례로 거친다.

## 2. 기본 lifecycle: 페이즈의 순서열

기본(default) lifecycle의 주요 페이즈는 대략 이 순서다.

1. `validate` — 프로젝트 구조·정보 검증.
2. `compile` — 메인 소스 컴파일. (이 직전에 의존성이 해석된다 — 그래서 받을 수 없는 의존성은 *컴파일 한 줄 전에* 멈춘다. [`./maven_dependency_resolution.md`](./maven_dependency_resolution.md).)
3. `process-test-resources` — 테스트 리소스 복사·필터링.
4. `test-compile` — 테스트 소스 컴파일.
5. `test` — 단위 테스트 실행.
6. `package` — JAR/WAR 등으로 패키징.
7. `verify` — 통합 검증.
8. `install` — 로컬 저장소에 설치.
9. `deploy` — 원격 저장소에 배포.

누적성이 핵심이다. `mvn package`는 compile·test까지 다 거치고, `mvn install`은 그 위에 package·install을 더한다. 그래서 "왜 install이 테스트를 또 도나?"의 답은 "install 페이즈에 도달하려면 그 앞 test 페이즈를 지나야 하기 때문"이다.

`test-compile`과 `test`가 나뉜 이유도 여기서 보인다 — 테스트 소스를 *컴파일하는* 단계와 *실행하는* 단계가 별개라, "컴파일은 되는지"만 빠르게 보려면 `mvn -DskipTests test-compile`처럼 실행 없이 컴파일까지만 갈 수 있다.

## 3. 페이즈는 이름, 일은 plugin goal이 한다

페이즈에 실제 동작을 채우는 건 플러그인의 goal이다. 패키징 타입(jar/war 등)마다 기본 바인딩이 정해져 있다.

- `compile` → `maven-compiler-plugin:compile`
- `test-compile` → `maven-compiler-plugin:testCompile`
- `test` → `maven-surefire-plugin:test`
- `package` → `maven-jar-plugin:jar` 또는 `maven-war-plugin:war`
- `install` → `maven-install-plugin:install`

페이즈를 거치지 않고 goal을 직접 부를 수도 있다. `mvn dependency:tree`, `mvn help:describe`처럼 `plugin:goal` 형태다. 그래서 "compile 같은 lifecycle 호출"과 "goal 직접 호출"은 다른 두 방식이고, 한 명령에 섞어 쓸 수도 있다(`mvn clean test`).

## 4. surefire가 테스트를 도는 법

`test` 페이즈는 surefire 플러그인이 맡는다. 알아야 할 동작이 셋이다.

1. **별도 JVM fork** — surefire는 기본적으로 테스트를 위해 *새 JVM을 띄운다*. 그래서 부모 `mvn` 프로세스에 준 `-D` 시스템 프로퍼티가 그 자식 JVM에 **자동으로 전달되지 않는다**. 테스트 JVM에 옵션을 넣으려면 surefire의 `<argLine>`을 써야 한다.

    ```xml
    <plugin>
      <groupId>org.apache.maven.plugins</groupId>
      <artifactId>maven-surefire-plugin</artifactId>
      <configuration>
        <argLine>-Dfile.encoding=UTF-8</argLine>
      </configuration>
    </plugin>
    ```

    "명령줄에 `-Dfile.encoding=UTF-8`을 줬는데 테스트엔 안 먹더라"의 원인이 이것이다.
2. **테스트 탐색** — surefire는 이름 패턴으로 테스트 클래스를 찾는다(기본 `**/*Test.java`, `**/Test*.java`, `**/*Tests.java` 등). 이 규칙에 안 맞는 이름의 테스트는 *CLI 빌드에서 조용히 누락*된다(IDE에선 클릭으로 돌아가 안 보일 수 있다 — [`./ide_vs_build_tool.md`](./ide_vs_build_tool.md)).
3. **결과 리포트** — `target/surefire-reports/`에 클래스별 `.xml`(기계용)과 `.txt`가 남는다. 실패를 대량으로 분류·집계할 때는 사람용 렌더링이 아니라 이 XML을 읽는 게 권위 있는 원천이다([`../thinking/failure_triage.md`](../thinking/failure_triage.md)).

실행 제어 옵션: `-Dtest=ClassA,ClassB`(특정만), `-DskipTests`(컴파일은 하고 실행만 건너뜀), `-DtestFailureIgnore=true`(실패해도 빌드 계속 — 전체 실패 목록을 한 번에 모을 때 유용), `-Dmaven.test.skip=true`(테스트 컴파일까지 통째로 건너뜀).

## 5. 플러그인 설정과 pluginManagement

`<build><plugins>`에서 플러그인의 바인딩을 추가하거나 설정한다. compiler 플러그인이 `release`/`source`/`target`·소스 인코딩을 받는 자리이고, surefire가 위 옵션을 받는 자리다. `<pluginManagement>`는 플러그인 버전·설정을 *중앙에서 고정*하되 실제 적용은 자식이 그 플러그인을 쓸 때만 일어난다(의존성의 `dependencyManagement`와 같은 결).

추가 소스 루트를 빌드에 넣는 것도 플러그인이 한다(예: `build-helper-maven-plugin:add-test-source`). 기본 lifecycle엔 테스트 소스 루트가 하나라, 별도 루트를 더하려면 이런 플러그인 goal을 적당한 페이즈(`generate-test-sources`)에 바인딩한다.

## 6. 프로파일과 활성화

프로파일(`<profile>`)은 빌드의 변형이다. 환경에 따라 소스 루트·플러그인·프로퍼티·의존성을 켜고 끈다. 활성화 방법이 여러 가지다.

- 명시: `mvn -P프로파일id`.
- `<activation>` 조건: 특정 프로퍼티 존재(`-Dx`), **파일 존재/부재**(`<file><exists>...`), OS, JDK 버전, 또는 `<activeByDefault>`.

특히 "파일 존재"로 활성화하면, 어떤 환경 표식(예: 특정 디렉터리)이 있을 때만 프로파일이 켜진다. 이걸로 "로컬에만 있는 동작"을 로컬에서만 자동으로 켜는 식의 환경 분기를 만든다.

## 7. 혼동하기 쉬운 점

- **`-DskipTests` vs `-Dmaven.test.skip`** — 앞은 테스트를 *컴파일은 하고 실행만* 건너뛴다. 뒤는 테스트 *컴파일까지* 건너뛴다. "테스트 코드가 안 깨지는지"만 보고 싶으면 skip이 아니라 `test-compile`을 부른다.
- **페이즈 누적성** — `mvn install`이 테스트를 도는 건 install이 test '뒤'에 있기 때문이다. 특정 페이즈만 똑 떼서 실행할 수는 없다(goal 직접 호출은 예외).
- **`-D`가 테스트에 안 먹음** — 4절. surefire가 fork한 JVM엔 `argLine`으로 넘겨야 한다.
- **테스트 이름 규칙** — surefire include 패턴에 안 맞는 테스트는 CLI에서 안 돌아간다. "IDE에선 돌았는데 CI엔 없다"의 흔한 원인.

## 8. 실패와 검증

- 페이즈에 무엇이 바인딩됐는지: `mvn help:describe -Dcmd=test`(또는 `-Dcmd=package`).
- 무엇이 왜 실행/스킵됐는지: `mvn -X`(디버그 로그)로 plugin 실행과 프로파일 활성화를 추적.
- 테스트 실제 결과: `target/surefire-reports/*.xml`을 본다(콘솔 요약이 잘리면 더 그렇다).
- 특정 테스트만: `mvn -Dtest=클래스 test`. 활성 프로파일 확인: `mvn help:active-profiles`.

## 9. 스스로 확인하기

- `mvn package`가 테스트를 도는 이유를 페이즈 누적성으로 설명해 보라.
- 명령줄 `-Dfile.encoding=UTF-8`이 테스트 JVM엔 왜 안 전달되나? 어떻게 전달하나?
- `-DskipTests`와 `-Dmaven.test.skip`의 차이는? "테스트 컴파일은 되는지"만 보려면 무엇을 쓰나?
- 이름이 `FooSpec.java`인 테스트가 IDE에선 도는데 `mvn test`에선 안 돈다. 왜일까?
- 특정 디렉터리가 있을 때만 켜지는 빌드 동작은 어떻게 만드나?

## 10. 확인한 사실과 추론

사실: lifecycle 페이즈 순서·누적성, 기본 plugin 바인딩, surefire의 fork·argLine·리포트 위치, 프로파일 활성화 방식은 Maven 공식 문서(Lifecycle Reference, Surefire Plugin)로 확인되는 표준 동작이다.

추론·주의: surefire의 기본 include 패턴·fork 여부는 버전·설정에 따라 달라질 수 있어(예: `forkCount` 0이면 같은 JVM), 실제 동작은 해당 프로젝트의 surefire 설정과 버전을 확인하는 편이 안전하다.

## 관련 노트

- 의존성 해석이 compile 전에 일어나는 그 메커니즘: [`./maven_dependency_resolution.md`](./maven_dependency_resolution.md)
- compile/test-compile에서 일어나는 컴파일의 본질: [`./javac_classpath_and_release.md`](./javac_classpath_and_release.md)
- 같은 프로젝트가 IDE와 CLI에서 다르게 도는 이유: [`./ide_vs_build_tool.md`](./ide_vs_build_tool.md)
- surefire XML로 대량 실패를 집계·분류하기: [`../thinking/failure_triage.md`](../thinking/failure_triage.md)
