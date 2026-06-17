# Maven은 의존성을 어떻게 찾고 해석하나 — 좌표와 로컬 저장소

빌드가 "이 라이브러리를 못 찾겠다"며 멈추거나, 반대로 "받을 수 없는 라이브러리를 내가 만든 것으로 대신 채우는" 일을 하려면, 먼저 Maven이 *무엇을 기준으로* 의존성을 식별하고 *어디서* 찾는지를 정확히 알아야 한다. 이 노트는 그 두 가지 — 좌표(coordinate)와 저장소(repository) — 를 중심으로 Maven의 의존성 해석을 처음부터 쌓는다. 특정 프로젝트와 무관한 일반 메커니즘만 다룬다.

읽고 나면 다음을 스스로 설명할 수 있어야 한다 — 좌표가 무엇이고 저장소의 어느 경로에 대응하는가, Maven이 어떤 순서로 찾는가, transitive 의존성과 scope가 클래스패스를 어떻게 바꾸는가, 그리고 임의의 JAR을 저장소에 직접 넣는 방법.

## 목차

- [1. 직답](#1-직답)
- [2. 좌표와 저장소 레이아웃](#2-좌표와-저장소-레이아웃)
- [3. 어디서 찾나: 해석 순서](#3-어디서-찾나-해석-순서)
- [4. POM이 의존성을 선언하는 법](#4-pom이-의존성을-선언하는-법)
- [5. 저장소에 넣는 법: install / deploy / install-file](#5-저장소에-넣는-법-install--deploy--install-file)
- [6. 다중 모듈 빌드(reactor)](#6-다중-모듈-빌드reactor)
- [7. 혼동하기 쉬운 점](#7-혼동하기-쉬운-점)
- [8. 실패와 검증](#8-실패와-검증)
- [9. 스스로 확인하기](#9-스스로-확인하기)
- [10. 확인한 사실과 추론](#10-확인한-사실과-추론)
- [관련 노트](#관련-노트)

## 1. 직답

Maven은 의존성을 **좌표(coordinate)** 로 식별한다. 좌표는 `groupId:artifactId:version`이고, 필요하면 packaging(기본 `jar`)과 classifier가 붙는다. 의존성을 해석할 때는 **로컬 저장소를 먼저** 보고, 없으면 **원격 저장소**에서 받아 로컬에 캐시한다.

핵심 성질 하나: 저장소는 좌표를 키로 하는 *디렉터리 트리*일 뿐이고, Maven은 그 좌표 자리에 있는 아티팩트가 "진짜"인지 내용을 검증하지 않는다. 좌표만 맞으면 해석된다.

## 2. 좌표와 저장소 레이아웃

좌표는 저장소 안의 한 경로로 1:1 대응한다. 규칙은 단순하다 — `groupId`의 점을 슬래시로 바꾸고, `artifactId`, `version` 디렉터리를 차례로 내려가면, 그 안에 `artifactId-version.packaging`과 같은 이름의 `.pom`이 있다.

좌표 `com.google.code.gson:gson:2.10.1`은 다음 경로다.

```text
<repo>/com/google/code/gson/gson/2.10.1/gson-2.10.1.jar
<repo>/com/google/code/gson/gson/2.10.1/gson-2.10.1.pom
```

`.pom`이 `.jar` 옆에 함께 있다는 점이 중요하다. 그 아티팩트가 *자기 의존성*을 무엇으로 가지는지는 이 `.pom`에 적혀 있고, Maven은 그걸 읽어 transitive 의존성을 따라간다(4절). classifier가 있으면 `artifactId-version-classifier.jar`(예: `...-sources.jar`)가 된다.

즉 "저장소"라는 건 특별한 서버가 아니라 이 레이아웃을 지키는 디렉터리(또는 그 디렉터리를 HTTP로 노출한 것)다. 그래서 좌표 자리에 파일을 맞춰 넣기만 하면 그 좌표는 "존재"하게 된다.

## 3. 어디서 찾나: 해석 순서

의존성 하나를 해석할 때 Maven이 보는 순서는 다음과 같다.

1. **로컬 저장소** — 기본 위치는 `~/.m2/repository`. 여기 좌표가 있으면 그대로 쓴다.
2. **원격 저장소** — 없으면 설정된 원격(기본은 Maven Central, 추가로 회사 Nexus 등)에서 받아 **로컬에 캐시**한 뒤 쓴다. 다음부터는 1번에서 바로 찾는다.

이 경로들은 설정으로 바뀐다.

- `~/.m2/settings.xml` — `<localRepository>`, 원격 저장소·미러(mirror), 인증.
- `-Dmaven.repo.local=<경로>` — 이번 실행에서만 로컬 저장소 위치를 바꾼다. 전역 `~/.m2`를 건드리지 않고 프로젝트 전용 저장소를 쓸 때 유용하다.
- 프로젝트 루트의 `.mvn/maven.config` — 그 디렉터리에서 실행되는 모든 `mvn`에 인자를 자동으로 붙인다. 예를 들어 한 줄 `-Dmaven.repo.local=.codex-m2`만 적어 두면 그 프로젝트는 항상 로컬 폴더를 저장소로 쓴다.
- 오프라인 모드 `mvn -o` — 원격 접근을 막고 **이미 로컬에 있는 것만** 쓴다. 없는 좌표를 만들어 주지는 않는다(5절과 구분).

## 4. POM이 의존성을 선언하는 법

`pom.xml`의 `<dependencies>`는 좌표 목록이다. 여기서 알아야 할 동작이 네 가지다.

1. **transitive(전이) 의존성** — 내가 A를 의존하면, A의 `.pom`이 선언한 의존성(B, C…)도 자동으로 끌려온다. 그래서 직접 적지 않은 라이브러리가 클래스패스에 들어온다. 어떤 좌표를 "대신 채울" 때, 그 좌표의 `.pom`이 transitive로 무엇을 끌어와야 하는지까지 맞춰야 소비자가 정상 컴파일된다.
2. **scope** — 의존성이 어느 클래스패스에, 어떻게 전이되는지를 정한다.
    - `compile`(기본): 컴파일·테스트·런타임 모두. 전이된다.
    - `provided`: 컴파일·테스트 클래스패스엔 있지만 런타임 패키지에는 포함 안 됨(서버/컨테이너가 제공한다고 가정). 전이 안 됨.
    - `runtime`: 컴파일엔 없고 테스트·런타임에만.
    - `test`: 테스트 컴파일·실행에만. 전이 안 됨.
    - `system`: 로컬 파일 경로를 직접 가리킴(`<systemPath>`). 전이·이식성이 없어 거의 쓰지 않는다.
3. **`<dependencyManagement>`** — 버전·scope를 *중앙에서 고정*하되, 그 자체로 의존성을 추가하진 않는다. 하위 모듈이 같은 좌표를 version 없이 선언하면 여기 버전이 적용된다.
4. **충돌 조정과 제외** — 같은 좌표의 다른 버전이 그래프에 둘 이상이면 Maven은 "가장 가까운 정의(nearest-wins)"를 택한다. 원치 않는 전이 의존성은 `<exclusions>`로 끊고, 전이시키고 싶지 않은 내 의존성은 `<optional>true</optional>`로 둔다.

## 5. 저장소에 넣는 법: install / deploy / install-file

아티팩트가 로컬 저장소에 *들어가는* 경로는 세 가지다.

- **`mvn install`** — 현재 프로젝트를 빌드해 그 산출물(JAR + POM)을 로컬 저장소에 자기 좌표로 넣는다. 같은 머신의 다른 빌드가 이 좌표를 의존할 수 있게 된다.
- **`mvn deploy`** — 원격 저장소(Nexus 등)에 올린다. 팀이 공유한다.
- **`mvn install:install-file`** — *빌드 없이* 임의의 JAR을 원하는 좌표로 로컬에 넣는다. `-Dfile=x.jar`로 파일을, `-DpomFile=x.pom`(또는 `-DgroupId/-DartifactId/-Dversion`)으로 좌표를 준다. 빌드 산출물이 아닌 외부 JAR을 특정 좌표로 등록할 때 쓴다.

버전 종류도 알아야 한다. `1.0.0` 같은 **release**는 불변으로 캐시된다(한 번 받으면 다시 안 받음). `1.0.0-SNAPSHOT`은 **개발 중** 버전이라 타임스탬프로 갱신되고 더 자주 재해석된다. SNAPSHOT을 의존하면 "어제와 다른 내용"이 들어올 수 있다는 뜻이다.

## 6. 다중 모듈 빌드(reactor)

여러 모듈을 한 번에 빌드하는 구조를 reactor라고 한다. 부모 `pom.xml`(packaging `pom`)이 `<modules>`로 자식들을 묶고, `mvn install`을 부모에서 실행하면 Maven이 **모듈 간 의존 그래프를 계산해 올바른 순서로** 빌드한다. 모듈 A가 모듈 B를 의존하면 B가 먼저 빌드된다.

reactor 안에서 모듈끼리는 보통 `SNAPSHOT` 좌표로 서로를 참조하며, 같은 실행(reactor) 안에서 해석된다. 그래서 "B를 먼저 install하고 A를 빌드"를 수동으로 안 해도 된다.

## 7. 혼동하기 쉬운 점

- **로컬 저장소 vs 오프라인 모드** — 로컬 저장소는 "어디서 찾고 캐시하나"의 위치이고, `-o`는 "원격을 아예 막는다"는 모드다. `-o`는 *비어 있는* 좌표를 채워 주지 않는다.
- **컴파일 클래스패스 vs 런타임** — 좌표가 해석돼 컴파일 클래스패스에 올라간다는 것과, 그 안의 클래스가 런타임에 실제로 동작한다는 것은 별개다. 컴파일은 시그니처만 보기 때문이다(자세히는 [`./javac_classpath_and_release.md`](./javac_classpath_and_release.md)).
- **"not found in central"의 의미** — 그 좌표를 *닿을 수 있는 어떤 저장소에서도* 못 찾았다는 것이지, 코드가 틀렸다는 게 아니다. 좌표 오타거나, 사내 전용이라 Central에 없거나, 원격 설정이 빠진 경우다.
- **SNAPSHOT 의외성** — 빌드가 "왜 어제와 다르지"의 흔한 원인. release로 고정하면 재현성이 올라간다.

## 8. 실패와 검증

- **좌표 해석 확인**: `mvn dependency:resolve`(직접 의존성), `mvn dependency:tree`(transitive 그래프 + 충돌 조정 결과). "왜 이 라이브러리가 들어왔지"는 거의 항상 `dependency:tree`로 답한다.
- **로컬에 있는지 확인**: `<repo>/<groupId를 /로>/<artifactId>/<version>/`에 `.jar`과 `.pom`이 있는지 본다.
- **`Could not resolve dependencies`** = 좌표를 어디서도 못 찾음 → 좌표 철자, 원격 저장소/미러 설정, 또는 그 좌표를 로컬에 직접 넣었는지(5절)를 점검.
- **강제 로컬 재현**: `mvn -o`로 돌려 보면 "지금 로컬에 실제로 있는 것만으로 빌드되는가"를 확인할 수 있다.

`확인해 보면 된다`가 아니라, 위처럼 무엇을 실행하고 무엇이 PASS/FAIL인지가 분명해야 한다.

## 9. 스스로 확인하기

- 좌표 `org.example:foo:2.1`은 로컬 저장소의 어느 경로에 놓이나? 그 옆에 `.pom`이 함께 있어야 하는 이유는?
- 내가 직접 적지 않은 라이브러리가 클래스패스에 들어왔다. 무엇으로 출처를 추적하나?
- `provided`와 `compile` scope는 패키징(런타임 산출물)에서 어떻게 다른가?
- `mvn -o`로는 해결되지 않지만 `install:install-file`로는 해결되는 상황은 무엇인가?
- 같은 좌표의 두 버전이 전이로 들어올 때 Maven은 어느 것을 택하나?

## 10. 확인한 사실과 추론

사실: 좌표·저장소 레이아웃·해석 순서·scope·`install-file`·reactor·SNAPSHOT/release는 Maven의 표준 동작이며 공식 문서로 확인 가능하다(Maven: Introduction to the Dependency Mechanism, Maven Lifecycle Reference).

추론·주의: 충돌 조정의 "nearest-wins"는 기본 규칙이지만 `dependencyManagement`·`exclusions`로 자주 덮이므로, 실제 결과는 항상 `dependency:tree`로 확인하는 편이 안전하다. 미러·인증·프록시가 끼면 해석 순서의 체감 동작이 달라질 수 있다.

## 관련 노트

- 컴파일이 왜 시그니처만 보는가, 컴파일 에러와 런타임 에러의 차이: [`./javac_classpath_and_release.md`](./javac_classpath_and_release.md)
- 페이즈·플러그인·테스트 실행이 이 위에서 어떻게 도는가: [`./maven_build_lifecycle.md`](./maven_build_lifecycle.md)
- 이 원리의 실제 적용(받을 수 없는 의존성을 로컬 좌표로 채우기): [`./maven_local_shim_repo.md`](./maven_local_shim_repo.md)
