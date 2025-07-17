# Maven Plugin exec

- [Maven Plugin exec](#maven-plugin-exec)
    - [Maven으로 `main` 실행](#maven으로-main-실행)
        - [CLI에서 직접 실행](#cli에서-직접-실행)
        - [`pom.xml`에 플러그인 설정](#pomxml에-플러그인-설정)

## Maven으로 `main` 실행

[Exec Maven Plugin](https://www.mojohaus.org/exec-maven-plugin/)을 사용합니다.
Maven 빌드 프로세스 내에서 외부 프로세스나 Java 프로그램을 실행할 수 있게 합니다.

- `exec:java`: 프로젝트의 클래스패스(의존성 포함)를 사용하여 JVM 내에서 직접 Java 클래스를 실행합니다.

    다음과 같은 과정을 거칩니다:

    1. 프로젝트의 모든 의존성을 클래스패스에 추가
    2. 필요시 자동으로 컴파일 단계 실행
    3. 지정된 `main` 클래스를 현재 JVM에서 로드
    4. `main` 메서드를 호출하여 애플리케이션 실행

- `exec:exec`: 일반적인 시스템 명령어나 스크립트를 실행합니다.

주요 매개변수는 다음과 같습니다:
- `exec.mainClass`: 실행할 main 클래스의 완전한 이름. e.g. `"com.example.App"`
- `exec.args`: main 메서드에 전달할 인수. e.g. `"arg1 arg2"`
- `exec.cleanupDaemonThreads`: 데몬 스레드 정리 여부. e.g. `false`

### CLI에서 직접 실행

`pom.xml` 파일을 수정하지 않고, 커맨드 라인에서 직접 `main` 메서드를 실행합니다.

1. `mvn compile` 명령을 실행하여 `App.java`를 `.class` 파일로 컴파일합니다.
2. `exec:java` 실행합니다.

```bash
mvn compile exec:java \
    -Dexec.mainClass="io.aimpugn.jv.file.App"
```

- `compile`: `exec:java`를 실행하기 전 컴파일합니다.
- `exec:java`: Exec Maven Plugin의 `java` 목표(goal)를 호출합니다.
- `-Dexec.mainClass`: 실행할 `main` 메서드가 포함된 클래스의 전체 패키지 경로(Fully Qualified Class Name)를 지정합니다.

만약 `main` 메서드에 인자를 전달해야 한다면 `-Dexec.args`를 사용합니다.

```bash
# "arg1"과 "arg2"를 인자로 전달
mvn compile exec:java \
    -Dexec.mainClass="io.aimpugn.jv.file.App" \
    -Dexec.args="arg1 arg2"
```

### `pom.xml`에 플러그인 설정

 `pom.xml` 파일에 `exec-maven-plugin` 설정을 추가합니다.

1. `pom.xml` 파일 열기: 프로젝트 루트의 `pom.xml` 파일을 엽니다.
2. `` 섹션에 플러그인 추가: `` → `` 섹션 안에 다음 내용을 추가합니다.

    ```xml
    <plugin>
        <groupId>org.codehaus.mojo</groupId>
        <artifactId>exec-maven-plugin</artifactId>
        <version>3.2.0</version>
        <configuration>
            <mainClass>io.aimpugn.jv.file.App</mainClass>
            <arguments>
                <argument>arg1</argument>
                <argument>arg2</argument>
            </arguments>
        </configuration>
    </plugin>
    ```

3. 실행: 설정이 추가된 후에는 터미널에서 더 간단한 명령으로 실행할 수 있습니다.

```bash
mvn compile exec:java
```

Maven 프로파일을 활용해 실행을 더욱 간소화할 수 있습니다.

```xml
<profiles>
    <profile>
        <id>run</id>
        <build>
            <defaultGoal>compile</defaultGoal>
            <plugins>
                <plugin>
                    <groupId>org.codehaus.mojo</groupId>
                    <artifactId>exec-maven-plugin</artifactId>
                    <version>3.1.0</version>
                    <executions>
                        <execution>
                            <phase>compile</phase>
                            <goals>
                                <goal>java</goal>
                            </goals>
                        </execution>
                    </executions>
                    <configuration>
                        <mainClass>io.aimpugn.jv.file.App</mainClass>
                    </configuration>
                </plugin>
            </plugins>
        </build>
    </profile>
</profiles>
```

설정 후 간단한 명령으로 실행:

```bash
mvn -Prun
```
