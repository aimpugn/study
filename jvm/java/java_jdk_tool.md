# Java JDK Tool

- [Java JDK Tool](#java-jdk-tool)
    - [JDK Tool Specification](#jdk-tool-specification)
    - [`java`](#java)
    - [`javac`](#javac)
        - [Synopsys](#synopsys)
        - [사용법](#사용법)
        - [옵션](#옵션)
        - [java 명령어](#java-명령어)
        - [Java 소스 파일 직접 실행](#java-소스-파일-직접-실행)
            - [사용 방법](#사용-방법)
            - [장점](#장점)
            - [제한사항](#제한사항)
            - [내부 동작 원리](#내부-동작-원리)
            - [성능 고려사항](#성능-고려사항)
            - [대안: jshell](#대안-jshell)
        - [결론](#결론)
    - [패키지 구조](#패키지-구조)
        - [`src/main/java`](#srcmainjava)
        - [패키지 구조는 파일 시스템의 디렉토리 구조와 일치해야 한다](#패키지-구조는-파일-시스템의-디렉토리-구조와-일치해야-한다)

## [JDK Tool Specification](https://docs.oracle.com/en/java/javase/17/docs/specs/man/)

```plaintext
https://docs.oracle.com/en/java/javase/<VERSION>/docs/specs/man/java.html
```

## [`java`](https://docs.oracle.com/en/java/javase/17/docs/specs/man/java.html)

## [`javac`](https://docs.oracle.com/en/java/javase/17/docs/specs/man/javac.html)

### Synopsys

```sh
javac <options> <source files>
```

### 사용법

```sh
# .class 바이트코드로 컴파일합니다.
❯ javac HelloWorld.java
❯ ll
total 16
-rw-r--r--@ 1 rody  staff   426B  9  2 23:44 HelloWorld.class
-rw-r--r--@ 1 rody  staff   179B  9  2 23:45 HelloWorld.java

#
❯ java HelloWorld
Hello World!
```

### 옵션

- `@<filename>`

    Read options and filenames from file

- `-Akey[=value]`

    Options to pass to annotation processors

- `--add-modules <module>(,<module>)*`

    Root modules to resolve in addition to the initial modules,
    or all modules on the module path if `<module>` is ALL-MODULE-PATH.

- `--boot-class-path <path>`, `-bootclasspath <path>`

    Override location of bootstrap class files

- `--class-path <path>`, `-classpath <path>`, `-cp <path>`

    Specify where to find user class files and annotation processors

- `-d <directory>`

    Specify where to place generated class files

- `-deprecation`

    Output source locations where deprecated APIs are used

- `--enable-preview`

    Enable preview language features.
    To be used in conjunction with either -source or --release.

- `-encoding <encoding>`

    Specify character encoding used by source files

- `-endorseddirs <dirs>`

    Override location of endorsed standards path

- `-extdirs <dirs>`

    Override location of installed extensions

- `-g`

    Generate all debugging info

- `-g:{lines,vars,source}`

    Generate only some debugging info

- `-g:none`

    Generate no debugging info

- `-h <directory>`
        Specify where to place generated native header files

- `--help`, `-help`, `-?`: Print this help message

- `--help-extra`, `-X`

    Print help on extra options

- `-implicit:{none,class}`

    Specify whether to generate class files for implicitly referenced files

- `-J<flag>`

    Pass `<flag>` directly to the runtime system

- `--limit-modules <module>(,<module>)*`

    Limit the universe of observable modules

- `--module <module>(,<module>)*`, `-m <module>(,<module>)*`

    Compile only the specified module(s), check timestamps

- `--module-path <path>`, `-p <path>`

    Specify where to find application modules

- `--module-source-path <module-source-path>`

    Specify where to find input source files for multiple modules

- `--module-version <version>`

    Specify version of modules that are being compiled

- `-nowarn`

    Generate no warnings

- `-parameters`

    Generate metadata for reflection on method parameters

- `-proc:{none,only,full}`

    Control whether annotation processing and/or compilation is done.

- `-processor <class1>[,<class2>,<class3>...]`

    Names of the annotation processors to run;
    bypasses default discovery process

- `--processor-module-path <path>`

    Specify a module path where to find annotation processors

- `--processor-path <path>`, `-processorpath <path>`

    Specify where to find annotation processors

- `-profile <profile>`

    Check that API used is available in the specified profile.
    This option is deprecated and may be removed in a future release.

- `--release <release>`

    Compile for the specified Java SE release.

    Supported releases:
    8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21

- `-s <directory>`
    Specify where to place generated source files

- `--source <release>`, `-source <release>`
    Provide source compatibility with the specified Java SE release.

    Supported releases:
    8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21

- `--source-path <path>`, `-sourcepath <path>`
    Specify where to find input source files

- `--system <jdk>|none`
    Override location of system modules

- `--target <release>`, `-target <release>`
    Generate class files suitable for the specified Java SE release.

    Supported releases:
    8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21

- `--upgrade-module-path <path>`
    Override location of upgradeable modules

- `-verbose`
    Output messages about what the compiler is doing

- `--version`, `-version`: Version information
- `-Werror`: Terminate compilation if warnings occur

### java 명령어

Java에는 실제로 컴파일과 실행을 한 번에 수행하고 .class 파일을 자동으로 삭제하는 내장 명령어는 없습니다. 하지만 Java 11부터 도입된 `java` 명령어의 새로운 기능을 사용하면 비슷한 효과를 얻을 수 있습니다. 이에 대해 자세히 설명드리겠습니다.

### Java 소스 파일 직접 실행

Java 11 이상에서는 소스 파일을 직접 실행할 수 있는 기능이 추가되었습니다. 이 기능을 사용하면 명시적인 컴파일 단계 없이 Java 프로그램을 실행할 수 있습니다.

#### 사용 방법

```bash
java HelloWorld.java
```

이 명령어는 다음과 같은 과정을 거칩니다:

1. HelloWorld.java 파일을 메모리 내에서 컴파일합니다.
2. 컴파일된 바이트코드를 메모리에 로드합니다.
3. 프로그램을 실행합니다.
4. 실행이 끝나면 메모리에서 바이트코드를 제거합니다.

#### 장점

1. 명시적인 컴파일 단계가 필요 없습니다.
2. .class 파일이 디스크에 생성되지 않습니다.
3. 한 줄의 명령어로 컴파일과 실행을 모두 수행할 수 있습니다.

#### 제한사항

1. 이 방식은 단일 소스 파일에만 적용 가능합니다.
2. 복잡한 프로젝트나 여러 파일로 구성된 프로그램에는 적합하지 않습니다.
3. 클래스패스 설정 등 일부 고급 옵션 사용에 제한이 있을 수 있습니다.

#### 내부 동작 원리

1. 소스 코드 파싱: `java` 명령어는 먼저 .java 파일을 읽고 파싱합니다.
2. 메모리 내 컴파일: 파싱된 소스 코드는 메모리 상에서 바이트코드로 컴파일됩니다. 이 과정에서 javac와 유사한 내부 컴파일러가 사용됩니다.
3. 클래스 로딩: 컴파일된 바이트코드는 JVM의 클래스 로더에 의해 메모리에 로드됩니다.
4. 실행: 로드된 클래스의 main 메소드가 실행됩니다.
5. 정리: 프로그램 실행이 종료되면, 사용된 메모리가 해제되고 로드된 클래스 정보가 제거됩니다.

#### 성능 고려사항

- 이 방식은 매번 소스 코드를 컴파일하므로, 미리 컴파일된 .class 파일을 실행하는 것보다는 약간의 오버헤드가 있을 수 있습니다.
- 그러나 작은 프로그램이나 스크립트와 같은 용도로 사용할 때는 이 차이가 크게 느껴지지 않을 것입니다.

#### 대안: jshell

Java 9부터 도입된 `jshell`이라는 REPL(Read-Eval-Print Loop) 도구도 있습니다. 이를 사용하면 Java 코드를 대화식으로 실행할 수 있습니다.

```bash
jshell
```

jshell 프롬프트에서:

```java
System.out.println("Hello World!");
```

이 방식은 임시 테스트나 빠른 프로토타이핑에 유용할 수 있습니다.

### 결론

Java 11 이상을 사용한다면, `java HelloWorld.java` 명령어를 사용하여 PHP나 Go와 유사한 방식으로 Java 프로그램을 빠르게 실행할 수 있습니다. 이 방식은 .class 파일을 생성하지 않으므로, 임시 파일 관리에 대한 걱정 없이 간단한 Java 프로그램을 실행할 수 있습니다.

## 패키지 구조

### `src/main/java`

`src/main/java` 구조는 Java 언어 자체의 스펙은 아니지만, Java 프로젝트의 표준 디렉토리 구조로 널리 사용되는 컨벤션입니다.

이 구조는 Apache Maven이 도입한 표준 디렉토리 레이아웃입니다.
Gradle 등 다른 빌드 도구들도 이를 채택했습니다.

- `src`: 소스 코드와 리소스를 포함
    - `main`: 메인 애플리케이션 코드
        - `java`: Java 소스 파일
        - `resources`: 리소스 파일
    - `test`: 테스트 코드

빌드 도구와의 호환되고 소스 코드, 테스트, 리소스의 명확한 분리가 가능합니다.

### 패키지 구조는 파일 시스템의 디렉토리 구조와 일치해야 한다

Java에서 패키지 구조는 파일 시스템의 디렉토리 구조와 일치해야 합니다.

예를 들어, `package some;`라고 선언한 클래스는 `some` 디렉토리 안에 있어야 합니다.

Java 프로젝트에서 '소스 루트'는 패키지 구조의 시작점을 정의합니다.
일반적으로 `src` 또는 `src/main/java` 디렉토리가 소스 루트로 설정됩니다.

IDE나 빌드 도구는 소스 루트 설정을 기반으로 패키지 구조를 해석합니다.

Java 패키지 이름은 일반적으로 모두 소문자로 작성하며, 숫자로 시작하지 않는 것이 관례입니다.

해결 방안:

1. 패키지 선언 제거
   가장 간단한 방법은 패키지 선언을 제거하는 것입니다. 각 `Main.java` 파일에서 `package p258712;` 라인을 삭제하세요.

2. 프로젝트 구조 변경
    패키지를 유지하고 싶다면, 디렉토리 구조를 변경해야 합니다:

    ```sh
    .
    └── src
        └── main
            └── java
                ├── p258712
                │   └── Main.java
                ├── p258713
                │   └── Main.java
                └── p258714
                    └── Main.java
    ```

    그리고 각 `Main.java` 파일의 패키지 선언을 유지합니다.

3. VSCode 설정 조정
    `.vscode/settings.json` 파일에 다음 설정을 추가합니다:

    ```json
    {
        "java.project.sourcePaths": ["."],
        "java.project.outputPath": "bin",
        "java.project.referencedLibraries": ["lib/**/*.jar"]
    }
    ```

    이 설정은 현재 디렉토리를 소스 루트로 인식하게 합니다.

4. Java 프로젝트 파일 생성
   프로젝트 루트에 `.classpath` 파일을 생성하고 다음 내용을 추가합니다:

   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <classpath>
       <classpathentry kind="src" path="258712"/>
       <classpathentry kind="src" path="258713"/>
       <classpathentry kind="src" path="258714"/>
       <classpathentry kind="output" path="bin"/>
   </classpath>
   ```

   이 파일은 각 문제 디렉토리를 개별 소스 루트로 인식하게 합니다.

5. 패키지 이름 변경 (선택사항)
   Java 명명 규칙을 따르려면 패키지 이름을 `p258712` 대신 `problem258712`와 같이 변경할 수 있습니다.

이러한 방법들 중 하나를 선택하여 적용하면 VSCode에서 Java 파일을 올바르게 인식하고 오류가 해결될 것입니다. 프로젝트의 규모와 목적에 따라 가장 적합한 방법을 선택하시면 됩니다.
