# Spring Boot executable jar startup

- [개요](#개요)
- [먼저 눈으로 볼 구조와 공식 자료](#먼저-눈으로-볼-구조와-공식-자료)
- [전체 흐름 한 번에 보기](#전체-흐름-한-번에-보기)
- [셸은 입력을 어떻게 받아 명령으로 바꿀까](#셸은-입력을-어떻게-받아-명령으로-바꿀까)
- [운영 체제는 어떻게 java를 실행할까](#운영-체제는-어떻게-java를-실행할까)
- [`META-INF/MANIFEST.MF`는 무엇이고 누가 정했을까](#meta-infmanifestmf는-무엇이고-누가-정했을까)
- [일반 실행형 JAR와 Spring Boot executable jar](#일반-실행형-jar와-spring-boot-executable-jar)
- [JarLauncher는 실제로 무엇을 할까](#jarlauncher는-실제로-무엇을-할까)
- [`SpringApplication.run()`은 언제 호출될까](#springapplicationrun은-언제-호출될까)
- [언제 애플리케이션이 떴다고 볼 수 있을까](#언제-애플리케이션이-떴다고-볼-수-있을까)
- [직접 확인해 보기](#직접-확인해-보기)
- [정리](#정리)
- [근거와 추가 읽기](#근거와-추가-읽기)

## 개요

여기서 터미널과 셸을 통해 `java -jar SpringBootApp.jar`를 명령하면 실제로 누가 무엇을 읽고, 어느 순간부터 "이 애플리케이션이 떴다"라고 판단할 수 있는지 정리합니다.

핵심은 운영 체제가 곧바로 내 애플리케이션의 `main()`을 실행하는 것이 아니라는 점입니다.
먼저 셸이 입력 줄을 해석하고 `java`라는 네이티브 실행 파일을 찾습니다.
그 다음 운영 체제가 그 실행 파일을 메모리에 올리고 프로세스를 시작합니다.
그 뒤에 실행된 Java launcher가 JAR의 `META-INF/MANIFEST.MF`를 읽고, Spring Boot executable jar라면 `JarLauncher`가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 해석한 뒤 실제 애플리케이션 시작 클래스까지 연결합니다.
그 다음에야 `SpringApplication.run()`이 실행되고, `ApplicationContext`, 내장 웹 서버, readiness 단계가 차례로 닫힙니다.

## 먼저 눈으로 볼 구조와 공식 자료

예를 들어, [fully-portable-spring-boot-0.0.1-SNAPSHOT.jar](/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar)을 보면 다음과 같이 실행에 참여하는 항목들을 확인할 수 있습니다.

```sh
jar tf /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar \
  | rg 'META-INF/MANIFEST.MF|JarLauncher.class|FullyPortableApplication.class|spring-boot-3.3.4.jar|spring-boot-autoconfigure-3.3.4.jar'
```

```text
META-INF/MANIFEST.MF
org/springframework/boot/loader/launch/JarLauncher.class
BOOT-INF/classes/com/example/portable/FullyPortableApplication.class
BOOT-INF/lib/spring-boot-3.3.4.jar
BOOT-INF/lib/spring-boot-autoconfigure-3.3.4.jar
```

이제 각 항목을 하나씩 보면 다음과 같습니다.

1. `META-INF/MANIFEST.MF`

    JAR의 manifest 파일입니다.
    `java -jar`로 실행할 때 Java launcher는 이 파일을 열어 `Main-Class`를 읽습니다.

2. `org/springframework/boot/loader/launch/JarLauncher.class`

    Spring Boot loader의 시작 클래스입니다.
    Spring Boot executable jar에서는 manifest의 `Main-Class`가 보통 이 클래스를 가리킵니다.

    이 클래스가 `BOOT-INF/classes/`가 아니라 jar 루트에 들어 있는 이유도 중요합니다.
    아직 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 classpath로 풀기 전에도 먼저 로드되어야 하기 때문입니다.

3. `BOOT-INF/classes/com/example/portable/FullyPortableApplication.class`

    우리가 만든 애플리케이션 클래스입니다.
    manifest의 `Start-Class`가 이 클래스를 가리키면, 결국 이 클래스의 `main()`으로 도달하게 됩니다.

4. `BOOT-INF/lib/spring-boot-3.3.4.jar`

    `SpringApplication` 같은 Spring Boot 핵심 부팅 코드가 들어 있는 라이브러리입니다.
    애플리케이션 `main()`이 `SpringApplication.run()`을 호출하면, 실제 부팅 로직은 이 라이브러리 안의 코드가 담당합니다.

5. `BOOT-INF/lib/spring-boot-autoconfigure-3.3.4.jar`

    자동 구성(auto-configuration) 로직이 들어 있는 라이브러리입니다.
    `ApplicationContext`를 올리는 과정에서 조건에 맞는 자동 구성을 적용할 때 이 라이브러리가 함께 참여합니다.

manifest도 바로 확인할 수 있습니다.

```sh
unzip -p /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
```

```text
Main-Class: org.springframework.boot.loader.launch.JarLauncher
Start-Class: com.example.portable.FullyPortableApplication
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
```

여기서 각 항목은 다음 뜻입니다.

1. `Main-Class`

    Java launcher가 가장 먼저 시작할 클래스를 가리킵니다.
    여기서는 `JarLauncher`가 들어 있으므로, Java launcher는 내 애플리케이션 클래스로 바로 들어가지 않습니다.

2. `Start-Class`

    `JarLauncher`가 classpath를 준비한 뒤 최종적으로 넘겨 줄 실제 애플리케이션 시작 클래스입니다.

3. `Spring-Boot-Classes`

    애플리케이션 클래스들을 어디서 찾아야 하는지 알려 줍니다.
    여기서는 `BOOT-INF/classes/`입니다.

4. `Spring-Boot-Lib`

    라이브러리 jar들을 어디서 찾아야 하는지 알려 줍니다.
    여기서는 `BOOT-INF/lib/`입니다.

이제 위 항목들이 실제 실행에서 어떻게 엮이는지 순서대로 보면 다음과 같습니다.

1. 운영 체제가 `java` 네이티브 실행 파일을 시작합니다.

    이 단계에서는 아직 jar 내부를 읽지 않습니다.

2. Java launcher가 jar를 열고 `META-INF/MANIFEST.MF`를 읽습니다.

    여기서 `Main-Class`를 보고 첫 시작 클래스를 결정합니다.

3. `Main-Class`가 `JarLauncher`이므로 Spring Boot loader가 먼저 시작합니다.

    이때 `JarLauncher.class`가 jar 루트에 있어야 하는 이유가 드러납니다.
    아직 `BOOT-INF/classes/`와 `BOOT-INF/lib/`가 classpath로 정리되기 전이므로, 먼저 로드될 수 있는 위치에 있어야 합니다.

4. `JarLauncher`가 `Spring-Boot-Classes`와 `Spring-Boot-Lib`를 읽어 classpath를 준비합니다.

    즉 `BOOT-INF/classes/`에 있는 애플리케이션 클래스와 `BOOT-INF/lib/`에 있는 라이브러리 jar들을 이제야 실행 가능한 형태로 연결합니다.

5. `JarLauncher`가 `Start-Class`를 찾아 애플리케이션 `main()`을 호출합니다.

    여기서 `BOOT-INF/classes/com/example/portable/FullyPortableApplication.class`가 실제 시작점으로 연결됩니다.

6. 애플리케이션 `main()` 안의 `SpringApplication.run()`이 Spring Boot 라이브러리들을 사용해 컨텍스트를 올립니다.

    이때 `spring-boot-3.3.4.jar` 안의 부팅 코드와 `spring-boot-autoconfigure-3.3.4.jar` 안의 자동 구성 로직이 함께 동작합니다.

같이 보면 좋은 공식 자료는 다음과 같습니다.

1. [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)

    `META-INF`, `MANIFEST.MF`, `Main-Class` 같은 규칙은 Java 표준 쪽에서 정의합니다.

2. [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html)

    `java -jar jarfile`가 정확히 무엇을 뜻하는지 Java launcher 관점에서 설명합니다.

3. [Spring Boot Executable Jar Format](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)

    `BOOT-INF/classes/`, `BOOT-INF/lib/`, `JarLauncher`, `Start-Class`는 Spring Boot executable jar 규약 쪽 이야기입니다.

4. [`execve(2)` man page](https://man7.org/linux/man-pages/man2/execve.2.html), [`elf(5)` man page](https://man7.org/linux/man-pages/man5/elf.5.html)

    셸이 명령을 실행할 때 운영 체제가 네이티브 실행 파일을 어떻게 시작하는지 이해할 때 도움이 됩니다. `elf(5)`는 Linux 실행 파일 포맷 예시입니다.

## 전체 흐름 한 번에 보기

먼저 큰 흐름을 한 번에 보면 다음과 같습니다.

```text
터미널 입력
    -> 셸이 입력 줄을 토큰으로 나누고 PATH에서 java를 찾음
    -> 운영 체제가 java 네이티브 실행 파일을 적재
    -> 실행된 java launcher가 -jar 옵션을 해석
    -> JAR 안의 META-INF/MANIFEST.MF를 읽음
    -> Main-Class = JarLauncher
    -> JarLauncher가 BOOT-INF/classes/와 BOOT-INF/lib/를 기준으로 classpath 준비
    -> Start-Class의 main() 호출
    -> SpringApplication.run()
    -> ApplicationContext refresh
    -> 내장 웹 서버 초기화
    -> ApplicationStartedEvent
    -> ApplicationReadyEvent
    -> ReadinessState.ACCEPTING_TRAFFIC
```

이제 위 흐름을 아래층부터 하나씩 올려 보겠습니다.

## 셸은 입력을 어떻게 받아 명령으로 바꿀까

터미널 창에 `java -jar SpringBootApp.jar --server.port=8081`를 입력했다고 가정해 보겠습니다.
이때 키 입력을 화면에 보여 주는 것은 터미널 프로그램이고, Enter를 눌러 한 줄이 완성되면 그 줄을 읽어 해석하는 것은 셸입니다.

셸은 이 입력 줄을 그대로 운영 체제에 넘기지 않습니다.
먼저 공백, 따옴표, 이스케이프, 변수 치환 같은 규칙을 적용해서 "실제로 실행할 프로그램 이름"과 "그 프로그램에 넘길 인자 목록"으로 정리합니다.
예를 들어 `"Spring Boot App.jar"`처럼 따옴표를 썼다면, 셸은 공백을 기준으로 둘로 쪼개지 않고 하나의 인자로 묶습니다.

작게 줄이면 셸이 하는 일은 다음과 같습니다.

1. 입력 줄을 읽습니다.

    터미널에서 받은 문자열 한 줄을 셸이 읽습니다.

2. 토큰과 인자를 만듭니다.

    공백, 따옴표, 이스케이프, 변수 치환 규칙을 적용해서 `argv` 형태의 인자 배열로 바꿉니다.

3. 실행할 대상을 찾습니다.

    `java`가 셸 builtin인지, 함수인지, 외부 실행 파일인지 판단하고, 외부 실행 파일이면 `PATH`를 따라 실제 경로를 찾습니다.

실제로 셸이 인자를 어떤 모양으로 만들지 감을 잡으려면 다음 작은 예시를 보면 됩니다.

```text
input : java -jar "Spring Boot App.jar" --server.port=8081
argv  : ['java', '-jar', 'Spring Boot App.jar', '--server.port=8081']
```

즉 셸이 하는 일은 "문자열 한 줄"을 "실행 파일 + 인자 목록"으로 바꾸는 것입니다.
이 단계에서는 아직 JAR를 열지 않았고, Spring Boot를 이해하지도 않았습니다.
셸은 그저 `java`라는 프로그램을 어떤 인자와 함께 실행할지 정리하는 역할을 맡습니다.

로컬에서도 `command -v java`를 실행해 보면 실제 실행 대상이 `/usr/bin/java`로 해석되는 것을 확인할 수 있습니다.

## 운영 체제는 어떻게 java를 실행할까

셸이 실행 대상을 정하면, 그다음에는 운영 체제가 네이티브 실행 파일을 시작합니다.
유닉스 계열 환경에서는 보통 `fork + exec` 흐름으로 설명합니다.
셸이 새 프로세스 문맥을 준비하고, `execve(2)` 계열 호출로 현재 실행 이미지를 새 프로그램 이미지로 바꾸는 식입니다.

여기서 중요한 것은 운영 체제가 JAR 내부 구조를 이해하지 않는다는 점입니다.
운영 체제가 직접 다루는 것은 `SpringBootApp.jar`가 아니라 `java`라는 네이티브 실행 파일입니다.
그 실행 파일이 운영 체제가 정한 실행 파일 포맷에 맞게 만들어져 있으면, 운영 체제는 그 파일을 메모리에 올리고 프로세스를 시작할 수 있습니다.

이 실행 파일 포맷은 플랫폼마다 다릅니다.

1. Linux에서는 보통 ELF(Executable and Linkable Format)를 봅니다.

    [`elf(5)`](https://man7.org/linux/man-pages/man5/elf.5.html) 문서를 보면 ELF 실행 파일은 ELF header와 program header table 같은 구조를 가진다고 설명합니다.
    운영 체제는 이런 헤더를 읽고 "어디를 메모리에 매핑하고 어디서 실행을 시작할지"를 판단합니다.

2. macOS에서는 같은 역할을 Mach-O 포맷이 맡습니다.

    이 문서를 작성한 현재 머신은 macOS이므로, 실제 관찰 결과도 ELF가 아니라 Mach-O로 나옵니다.

로컬에서 직접 확인하면 다음과 같습니다.

```sh
file /usr/bin/java
```

```text
/usr/bin/java: Mach-O universal binary with 2 architectures: ...
```

즉 이 머신에서 운영 체제가 실제로 시작하는 대상은 JAR가 아니라 Mach-O 형식의 `/usr/bin/java`입니다.

이 다음에 운영 체제와 런타임 로더는 이 네이티브 프로그램이 필요로 하는 공유 라이브러리도 함께 연결합니다.
로컬에서 `otool -L /usr/bin/java`를 보면 Java launching 관련 framework가 보이고, JDK 홈 아래를 보면 `libjli.dylib`, `libjvm.dylib` 같은 라이브러리도 확인할 수 있습니다.

```text
/usr/bin/java:
    /System/Library/PrivateFrameworks/JavaLaunching.framework/Versions/A/JavaLaunching

/Library/Java/JavaVirtualMachines/graalvm-25.jdk/Contents/Home/lib/libjli.dylib
/Library/Java/JavaVirtualMachines/graalvm-25.jdk/Contents/Home/lib/server/libjvm.dylib
```

여기서 `libjvm`은 JVM 엔진 자체가 들어 있는 공유 라이브러리라고 생각하면 됩니다.
즉 `java` 실행 파일 하나가 모든 것을 혼자 다 들고 있는 것이 아니라, 실행 도중 필요한 공유 라이브러리를 운영 체제의 동적 로딩 메커니즘으로 연결해 JVM을 구성합니다.

동적 로딩은 아주 거칠게 말해, "실행 파일 안에 모든 기계어를 한 덩어리로 박아 두지 않고, 필요한 라이브러리를 실행 시점에 프로세스 주소 공간에 연결하는 방식"입니다.
그래서 운영 체제는 네이티브 실행 파일과 공유 라이브러리를 다루고, 그 위에서 돌아가는 Java launcher가 비로소 JAR와 manifest를 읽기 시작합니다.

## `META-INF/MANIFEST.MF`는 무엇이고 누가 정했을까

이제 JAR 안으로 올라가 보겠습니다.

Java 표준의 [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)은 JAR를 ZIP 기반 파일 포맷으로 설명합니다.
이 문서에 따르면 JAR는 `META-INF` 디렉터리를 가질 수 있고, 그 안의 `MANIFEST.MF`는 JAR 자체와 애플리케이션에 대한 설정 정보를 담는 manifest 파일입니다.

즉 `META-INF/MANIFEST.MF`는 Spring Boot가 임의로 만든 파일이 아니라, JAR 규약 안에 있는 표준 위치입니다.
이 파일 안에는 `name: value` 형태의 속성이 들어가고, executable JAR에서는 `Main-Class` 같은 속성이 시작 경로를 결정합니다.

Oracle의 [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html) 문서는 `java -jar jarfile`를 다음 뜻으로 설명합니다.

1. 실행 단위는 class 이름이 아니라 JAR 파일입니다.
2. 그 JAR의 manifest에는 `Main-Class`가 있어야 합니다.
3. `-jar`를 사용할 때는 지정한 JAR 파일이 user classes의 source가 됩니다.

즉 `java -jar app.jar`라고 입력했을 때 Java launcher는 "JAR를 열고, manifest를 읽고, `Main-Class`가 가리키는 클래스를 시작점으로 잡는 규칙"을 이미 알고 있습니다.

이 문서에서 계속 쓰는 manifest도 다시 실제 파일로 보면 더 분명합니다.

```text
Main-Class: org.springframework.boot.loader.launch.JarLauncher
Start-Class: com.example.portable.FullyPortableApplication
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
```

여기서 `Main-Class`는 Java 표준 executable JAR 규약 쪽 항목입니다.
반면 `Start-Class`, `Spring-Boot-Classes`, `Spring-Boot-Lib`는 Spring Boot executable jar를 실행하기 위해 추가된 Spring Boot 쪽 메타데이터입니다.

즉 이 manifest 한 장 안에 두 층의 규칙이 같이 들어 있습니다.

1. Java launcher가 이해하는 규칙

    `Main-Class`

2. Spring Boot loader가 추가로 이해하는 규칙

    `Start-Class`, `Spring-Boot-Classes`, `Spring-Boot-Lib`

이 구분이 잡혀 있으면 "`Main-Class`와 `Start-Class`가 왜 둘 다 있지?"라는 의문이 자연스럽게 풀립니다.

## 일반 실행형 JAR와 Spring Boot executable jar

이제 일반적인 실행형 JAR와 Spring Boot executable jar를 나눠서 보면 구조가 더 선명합니다.

1. 일반 실행형 JAR

    manifest의 `Main-Class`가 곧 애플리케이션 시작 클래스입니다.
    Java launcher는 그 클래스를 로드하고, 그 클래스의 `main()`을 호출합니다.

2. Spring Boot executable jar

    manifest의 `Main-Class`는 보통 `org.springframework.boot.loader.launch.JarLauncher`입니다.
    실제 애플리케이션 시작 클래스는 `Start-Class`에 들어 있습니다.

    즉 Java launcher가 내 애플리케이션 클래스로 바로 들어가는 것이 아니라, 먼저 `JarLauncher`를 실행합니다.
    `JarLauncher`는 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 읽어 실행 가능한 classpath를 준비한 뒤 `Start-Class`의 `main()`을 호출합니다.

이 차이를 짧게 쓰면 다음과 같습니다.

```text
일반 실행형 JAR
    Java launcher
        -> Main-Class = 내 애플리케이션
        -> main()

Spring Boot executable jar
    Java launcher
        -> Main-Class = JarLauncher
        -> JarLauncher가 BOOT-INF/classes/와 BOOT-INF/lib/를 해석
        -> Start-Class의 main()
```

즉 Spring Boot executable jar에서는 `Main-Class`와 실제 애플리케이션 시작 클래스가 다릅니다.
이 둘 사이에 `JarLauncher`가 들어가 있는 이유는, 라이브러리들이 평평한 classpath 바깥이 아니라 `BOOT-INF/lib/` 아래 nested jar 형태로 들어 있기 때문입니다.

## JarLauncher는 실제로 무엇을 할까

Spring Boot 공식 문서의 [Launching Executable Jars](https://docs.spring.io/spring-boot/specification/executable-jar/launching.html)는 `Launcher`를 executable jar의 main entry point라고 설명합니다.
같은 문서에서 `JarLauncher`는 `BOOT-INF/lib/`를 보고 classpath를 구성하고, `Start-Class`를 실제 시작 클래스로 사용한다고 설명합니다.

즉 `JarLauncher`는 단순히 "한 번 거쳐 가는 클래스"가 아닙니다.
Spring Boot executable jar 구조를 실제로 실행 가능한 Java classpath로 번역하는 bootstrap 담당자입니다.

작게 줄이면 `JarLauncher`가 하는 일은 다음과 같습니다.

1. 지금 실행 중인 archive가 Spring Boot executable jar 구조인지 전제하고 내부 경로를 해석합니다.

    `BOOT-INF/classes/`에 애플리케이션 클래스가 있고, `BOOT-INF/lib/`에 라이브러리 jar들이 있다는 규약을 사용합니다.

2. 그 구조를 바탕으로 classpath와 class loader를 준비합니다.

    Java launcher는 manifest의 `Main-Class`까지는 알지만, nested jar가 들어 있는 Spring Boot 내부 구조를 스스로 일반 classpath처럼 풀어 주지는 않습니다.
    이 부분을 `JarLauncher`와 Spring Boot loader가 맡습니다.

3. `Start-Class`를 찾아 실제 애플리케이션의 `main()`을 호출합니다.

    여기서부터 비로소 우리가 평소 보던 애플리케이션 시작 클래스 코드가 실행됩니다.

이렇게 보면 `JarLauncher`는 "내 애플리케이션을 대신 실행하는 프레임워크"라기보다, "Spring Boot jar 구조를 일반적인 Java 실행 흐름으로 연결하는 다리"에 가깝습니다.

## `SpringApplication.run()`은 언제 호출될까

보통 스프링 부트 애플리케이션 시작 클래스의 `main()`은 다음과 같이 생겼습니다.

```java
public static void main(String[] args) {
    SpringApplication.run(MyApplication.class, args);
}
```

이 코드는 시작점처럼 보입니다.
하지만 Spring Boot executable jar 경우 `java launcher -> manifest -> JarLauncher -> Start-Class 연결`이 끝난 후 이 `main()`이 실행됩니다.

즉 `SpringApplication.run()`은 "프로세스가 처음 생기는 순간"이 아니라, 이미 Java 프로세스가 살아 있고, manifest가 해석되었고, Spring Boot loader가 classpath를 준비한 뒤에 호출됩니다.

이제부터는 Spring Boot framework에 대해 설명합니다.
`SpringApplication.run()`이 호출되면 대략 다음 흐름이 이어집니다.

1. SpringApplication이 bootstrap 환경을 준비합니다.
2. Environment를 만들고 설정을 읽습니다.
3. `ApplicationContext`를 생성하고 bean definition을 로드합니다.
4. context refresh 과정에서 필요한 bean을 만들고 초기화합니다.
5. 웹 애플리케이션이면 내장 웹 서버를 초기화하고 포트를 바인드합니다.
6. `ApplicationRunner`, `CommandLineRunner`를 호출합니다.
7. 준비 완료 이벤트와 readiness 상태가 올라갑니다.

즉 `JarLauncher`와 `SpringApplication.run()`은 담당 계층이 다릅니다.

1. `JarLauncher`

    Spring Boot executable jar 파일 구조를 실행 가능한 classpath로 푸는 단계입니다.

2. `SpringApplication.run()`

    실제 Spring 애플리케이션을 세우는 단계입니다.
    environment, bean, context, 웹 서버, runner, readiness가 여기서 연결됩니다.

이 두 단계를 섞어 생각하면 "`main()`이 시작점 아닌가?"라는 질문이 계속 남습니다.
Spring Boot executable jar에서는 `main()`도 이미 loader bootstrap 뒤에 도착한 시작점입니다.

## 언제 애플리케이션이 떴다고 볼 수 있을까

이 질문은 어느 계층에서 보느냐에 따라 답이 달라집니다.

1. 운영 체제 관점

    `java` 프로세스가 시작된 순간 이미 실행 중이라고 볼 수 있습니다.
    하지만 이 시점은 Spring 애플리케이션 준비 완료를 뜻하지는 않습니다.

2. Java launcher 관점

    `-jar` 대상 JAR을 열고 manifest를 읽기 시작한 시점부터 JAR 실행이 진행 중이라고 볼 수 있습니다.
    역시 이 시점은 아직 readiness와는 거리가 있습니다.

3. Spring container 관점

    `ApplicationContext`가 refresh 되면 Spring 컨테이너는 살아 있다고 볼 수 있습니다.
    그래도 runner가 끝났는지, 웹 서버가 포트를 열었는지는 더 봐야 합니다.

4. 트래픽 수신 관점

    내장 웹 서버가 실제로 포트를 바인드하고, runner까지 끝나고, readiness가 올라와야 "이제 요청을 받아도 된다"고 말하는 편이 더 정확합니다.

Spring Boot 공식 문서는 이 차이를 이벤트 순서로 보여 줍니다.
문서에 따르면 `ApplicationStartedEvent`는 context refresh 이후, runner 호출 전입니다.
그 다음 `ApplicationReadyEvent`가 오고, 이어서 `ReadinessState.ACCEPTING_TRAFFIC` 상태가 올라갑니다.

즉 다음 네 문장은 서로 같은 말이 아닙니다.

1. 프로세스가 떴다.
2. manifest를 읽기 시작했다.
3. 컨텍스트가 refresh 됐다.
4. 트래픽을 받아도 된다.

이 차이를 구분해야 startup hook, warm-up, health check, readiness probe를 정확하게 이해할 수 있습니다.

## 직접 확인해 보기

이 저장소의 샘플 jar를 실제로 실행하면 다음과 같이 확인할 수 있습니다.

```sh
java -jar /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar --server.port=0
```

이번 실행에서는 다음 줄이 보였습니다. 포트 번호와 소요 시간은 매번 달라질 수 있습니다.

```text
2026-04-14T00:36:39.058+09:00  INFO 22876 --- [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat initialized with port 0 (http)
2026-04-14T00:36:39.464+09:00  INFO 22876 --- [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat started on port 61937 (http) with context path '/'
2026-04-14T00:36:39.474+09:00  INFO 22876 --- [           main] c.e.portable.FullyPortableApplication    : Started FullyPortableApplication in 1.15 seconds (process running for 1.504)
```

이 로그를 읽는 기준은 다음과 같습니다.

1. `Tomcat initialized with port 0`

    웹 서버 초기화가 시작되었다는 뜻입니다.

2. `Tomcat started on port ...`

    실제 포트 바인딩이 끝났다는 강한 관측 신호입니다.

3. `Started FullyPortableApplication in ...`

    Spring Boot가 애플리케이션 시작 완료 로그를 남긴 시점입니다.

만약 더 엄밀하게 readiness를 보고 싶다면, 로그만 보지 말고 `ApplicationReadyEvent`나 `ReadinessState.ACCEPTING_TRAFFIC`까지 확인하는 편이 더 정확합니다.

## 정리

`java -jar SpringBootApp.jar`를 입력했을 때 실제로 이어지는 흐름은 다음처럼 정리할 수 있습니다.

1. 셸이 입력 줄을 실행 파일과 인자 목록으로 정리합니다.
2. 운영 체제가 네이티브 `java` 실행 파일을 시작합니다.
3. 실행된 Java launcher가 `-jar` 옵션과 `META-INF/MANIFEST.MF`를 읽습니다.
4. `Main-Class`가 `JarLauncher`이므로 Spring Boot loader가 먼저 실행됩니다.
5. `JarLauncher`가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 기준으로 classpath를 준비합니다.
6. `Start-Class`의 `main()`이 호출됩니다.
7. 그 안의 `SpringApplication.run()`이 `ApplicationContext`, 웹 서버, runner, readiness를 차례로 세웁니다.

즉 운영 체제가 직접 이해하는 것은 JAR가 아니라 `java`라는 네이티브 실행 파일입니다.
JAR 규약은 Java launcher가 이해하고, Spring Boot executable jar 규약은 `JarLauncher`가 이어서 이해합니다.
이 층을 나눠서 보면 "`java -jar`를 쳤는데 왜 내 `main()`이 바로 시작되지 않지?"라는 질문이 자연스럽게 풀립니다.

## 근거와 추가 읽기

- [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)
- [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html)
- [Spring Boot Executable Jar Format](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)
- [Launching Executable Jars](https://docs.spring.io/spring-boot/specification/executable-jar/launching.html)
- [SpringApplication](https://docs.spring.io/spring-boot/4.1/reference/features/spring-application.html)
- [`execve(2)` man page](https://man7.org/linux/man-pages/man2/execve.2.html)
- [`elf(5)` man page](https://man7.org/linux/man-pages/man5/elf.5.html)
- 샘플 jar: [fully-portable-spring-boot-0.0.1-SNAPSHOT.jar](/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar)
