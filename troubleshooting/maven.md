# Maven

- [Maven](#maven)
    - [Plugin org.apache.maven.plugins:maven-archetype-plugin:3.4.0 or one of its dependencies could not be resolved](#plugin-orgapachemavenpluginsmaven-archetype-plugin340-or-one-of-its-dependencies-could-not-be-resolved)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
            - [Nexus 저장소 캐시 무효화](#nexus-저장소-캐시-무효화)
            - [`mvn dependency:get`을 통한 단일 아티팩트 다운로드해보기](#mvn-dependencyget을-통한-단일-아티팩트-다운로드해보기)
            - [수동으로 아티팩트 설치](#수동으로-아티팩트-설치)

## Plugin org.apache.maven.plugins:maven-archetype-plugin:3.4.0 or one of its dependencies could not be resolved

### 문제

```log
standalone-pom: Execution default-cli of goal org.apache.maven.plugins:maven-archetype-plugin:3.4.0:generate failed: Plugin org.apache.maven.plugins:maven-archetype-plugin:3.4.0 or one of its dependencies could not be resolved: The following artifacts could not be resolved: org.apache.groovy:groovy:jar:4.0.26 (present, but unavailable), org.apache.ivy:ivy:jar:2.5.3 (present, but unavailable): Could not find artifact org.apache.groovy:groovy:jar:4.0.26 in public
```

### 원인

Maven은 의존성을 찾기 위해 다음과 같은 계층적 구조를 사용합니다.

1. 로컬 저장소 (`.m2/repository`): 개발자 머신에 캐시된 아티팩트가 있는지 먼저 확인합니다.
2. 원격 저장소 (Remote Repositories): 로컬에 없다면 `pom.xml`이나 `settings.xml`에 설정된 원격 저장소(기본값은 Maven Central)에서 다운로드를 시도합니다.

`(present, but unavailable)`이라는 메시지가 나타나는 것은 Nexus Maven Repository와 같은 프록시 저장소를 사용할 때 발생할 수 있으며, Nexus 서버(프록시) 측의 문제일 가능성이 높다고 합니다.

- `present`:
    - Nexus 저장소가 해당 아티팩트(`groovy:jar:4.0.26`)의 존재 자체는 인지하고 있음을 의미
    - 보통 Nexus가 원격 저장소(Maven Central)로부터 메타데이터 파일(주로 `.pom` 파일)은 성공적으로 가져왔을 때 발생
- `but unavailable`:
    - 하지만 정작 아티팩트의 핵심인 `.jar` 파일 자체를 제공할 수 없다는 의미
    - Nexus가 `.jar` 파일을 다운로드하려고 시도했지만 실패했거나, 다운로드 과정이 불완전하게 끝난 상태

Nexus의 Not Found Cache (NFC) 메커니즘 때문일 수 있다고 합니다.

Nexus는 성능 향상을 위해 프록시 역할을 수행합니다.
어떤 아티팩트 요청이 들어왔을 때, 원격 저장소(Maven Central)에 해당 파일이 없어서 다운로드에 실패하면, Nexus는 이 "찾을 수 없음" 상태를 일정 시간 동안 캐시(Not Found Cache)에 저장합니다.

1. 어떤 이유로든(일시적인 네트워크 오류, Maven Central에 아직 파일이 완전히 배포되지 않은 시점 등) Nexus가 `groovy-4.0.26.jar` 파일을 다운로드하려다 실패합니다.
2. Nexus는 "이 파일은 존재하지 않는다"는 정보를 자신의 NFC에 기록합니다.
3. 이후 PC에서 다시 해당 아티팩트를 요청하면, Nexus는 원격 저장소에 다시 확인하러 가지 않고, NFC에 기록된 "없음" 정보를 즉시 반환합니다.

`mvn -U` 옵션을 사용하면 "로컬 캐시(`~/.m2/repository`)를 무시하고 원격 저장소(Nexus)에 다시 물어보라"고 지시합니다.
하지만 요청을 받은 Nexus 서버는 여전히 자신의 NFC를 신뢰합니다.
Nexus 서버의 NFC가 비워지지 않는 한 "찾아봤지만 없었다"라고 응답하며 실제 Maven Central에는 접근하지 않습니다.

결국, 클라이언트 측에서 아무리 강제로 업데이트를 요청해도 서버 측의 캐시가 그대로라면 문제는 해결되지 않습니다.

### 해결

#### Nexus 저장소 캐시 무효화

Nexus 관리자 권한이 있다면, 웹 UI를 통해 직접 캐시를 비울 수 있습니다.

1. Nexus Repository Manager에 관리자로 로그인 후 Repositories로 이동
2. 문제가 발생하는 프록시 저장소(보통 `maven-central`)를 선택
3. 설정 화면 하단으로 스크롤하여 "Invalidate Cache" 버튼 클릭하여 Nexus가 해당 저장소에 대해 가지고 있는 모든 캐시(NFC 포함)를 강제로 비움

Nexus 서버의 캐시를 비운 후, 클라이언트 측 캐시도 지웁니다.

1. 로컬 `.m2` 저장소에서 문제되는 아티팩트의 디렉토리를 통째로 삭제하여 관련 캐시를 완전히 제거
    - `~/.m2/repository/org/apache/groovy/`
    - `~/.m2/repository/org/apache/ivy/`

2. `mvn` 명령어 재실행: `-U` 플래그와 함께 다시 명령어를 실행

    ```bash
    mvn -U archetype:generate ... (나머지 옵션 동일)
    ```

#### `mvn dependency:get`을 통한 단일 아티팩트 다운로드해보기

```bash
mvn dependency:get -Dartifact=org.apache.groovy:groovy:4.0.23
```

#### 수동으로 아티팩트 설치

만약 어떤 방법으로도 Maven이 아티팩트를 자동으로 다운로드하지 못한다면,
Maven Central에서 `.jar` 파일과 `.pom` 파일을 직접 다운로드하여 로컬 저장소에 수동으로 설치할 수도 있습니다.

1. Maven Central에서 `groovy-4.0.23.jar`와 `groovy-4.0.23.pom` 파일을 다운로드합니다.
2. 다음 명령어를 사용하여 로컬 저장소에 설치합니다.

    ```bash
    mvn install:install-file \
        -Dfile=/path/to/groovy-4.0.23.jar \
        -DpomFile=/path/to/groovy-4.0.23.pom \
        -DgroupId=org.apache.groovy \
        -DartifactId=groovy \
        -Dversion=4.0.23 \
        -Dpackaging=jar
    ```
