# Spring Boot executable jar startup

- [개요](#개요)
- [먼저 눈으로 볼 구조와 공식 자료](#먼저-눈으로-볼-구조와-공식-자료)
- [전체 흐름 한 번에 보기](#전체-흐름-한-번에-보기)
- [셸은 입력을 어떻게 받아 명령으로 바꿀까](#셸은-입력을-어떻게-받아-명령으로-바꿀까)
- [운영 체제는 어떻게 java를 실행할까](#운영-체제는-어떻게-java를-실행할까)
- [`META-INF/MANIFEST.MF`는 무엇이고, 왜 생겼고, 언제부터 있었을까](#meta-infmanifestmf는-무엇이고-왜-생겼고-언제부터-있었을까)
- [일반 실행형 JAR와 Spring Boot executable jar](#일반-실행형-jar와-spring-boot-executable-jar)
- [JarLauncher는 실제로 무엇을 할까](#jarlauncher는-실제로-무엇을-할까)
- [`SpringApplication.run()`은 언제 호출될까](#springapplicationrun은-언제-호출될까)
- [언제 애플리케이션이 떴다고 볼 수 있을까](#언제-애플리케이션이-떴다고-볼-수-있을까)
- [직접 확인해 보기](#직접-확인해-보기)
- [정리](#정리)
- [근거와 추가 읽기](#근거와-추가-읽기)

## 개요

여기서 터미널과 셸을 통해 `java -jar SpringBootApp.jar`를 명령하면 실제로 누가 무엇을 읽고, 어느 순간부터 "이 애플리케이션이 떴다"라고 판단할 수 있는지 정리합니다.

이 문서는 두 질문을 닫는 데 초점을 둡니다.

1. `java -jar SpringBootApp.jar`를 입력하면 실제로 누가 무엇을 읽는가
2. 어느 시점부터 "이 애플리케이션이 떴다"고 말할 수 있는가

핵심은 운영 체제가 곧바로 내 애플리케이션의 `main()`을 실행하는 것이 아니라는 점입니다.
먼저 셸이 입력 줄을 해석하고 `java`라는 네이티브 실행 파일을 찾습니다.
그 다음 운영 체제가 그 실행 파일을 메모리에 올리고 프로세스를 시작합니다.
그 뒤에 실행된 Java launcher가 JAR의 `META-INF/MANIFEST.MF`를 읽고, Spring Boot executable jar라면 `JarLauncher`가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 해석한 뒤 실제 애플리케이션 시작 클래스까지 연결합니다.
그 다음에야 `SpringApplication.run()`이 실행되고, `ApplicationContext`, 내장 웹 서버, readiness 단계가 차례로 닫힙니다.

## 먼저 눈으로 볼 구조와 공식 자료

예를 들어, [fully-portable-spring-boot-0.0.1-SNAPSHOT.jar](/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar)을 기준으로 startup path를 이해하려면 먼저 두 가지를 확인하면 됩니다.

1. JAR 안에 어떤 클래스와 라이브러리가 들어 있는가
2. manifest가 누구를 첫 시작 클래스로 지정하는가

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

manifest 전체 출력에는 버전 정보와 index 정보도 더 들어 있습니다. startup path를 이해하는 데 직접 참여하는 필드만 보면 다음과 같습니다.

```sh
unzip -p /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
```

```text
Main-Class: org.springframework.boot.loader.launch.JarLauncher
Start-Class: com.example.portable.FullyPortableApplication
Spring-Boot-Version: 3.3.4
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
```

이제 이 예시에서 눈여겨볼 항목을 하나씩 보면 다음과 같습니다.

- `META-INF/MANIFEST.MF`

    JAR 전체에 공통으로 적용되는 메타데이터를 적는 파일입니다.
    `java -jar`로 실행할 때 Java launcher는 이 파일을 열어 `Main-Class`를 읽습니다.

- `org/springframework/boot/loader/launch/JarLauncher.class`

    Spring Boot loader의 시작 클래스입니다.
    Spring Boot executable jar에서는 manifest의 `Main-Class`가 보통 이 클래스를 가리킵니다.

    이 클래스가 `BOOT-INF/classes/`가 아니라 jar 루트에 들어 있는 이유도 중요합니다.
    아직 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 classpath로 풀기 전에도 먼저 로드되어야 하기 때문입니다.

- `BOOT-INF/classes/com/example/portable/FullyPortableApplication.class`

    우리가 만든 애플리케이션 클래스입니다.
    manifest의 `Start-Class`가 이 클래스를 가리키면, 결국 이 클래스의 `main()`으로 도달하게 됩니다.

- `BOOT-INF/lib/spring-boot-3.3.4.jar`

    `SpringApplication` 같은 Spring Boot 핵심 부팅 코드가 들어 있는 라이브러리입니다.
    애플리케이션 `main()`이 `SpringApplication.run()`을 호출하면, 실제 부팅 로직은 이 라이브러리 안의 코드가 담당합니다.

- `BOOT-INF/lib/spring-boot-autoconfigure-3.3.4.jar`

    자동 구성(auto-configuration) 로직이 들어 있는 라이브러리입니다.
    `ApplicationContext`를 올리는 과정에서 조건에 맞는 자동 구성을 적용할 때 이 라이브러리가 함께 참여합니다.

방금 본 manifest를 startup path 기준으로 필드별로 풀면 다음과 같습니다.

- `Main-Class`

    Java launcher가 가장 먼저 시작할 클래스를 가리킵니다.
    여기서는 `JarLauncher`가 들어 있으므로, Java launcher는 내 애플리케이션 클래스로 바로 들어가지 않습니다.

- `Start-Class`

    `JarLauncher`가 classpath를 준비한 뒤 최종적으로 넘겨 줄 실제 애플리케이션 시작 클래스입니다.

- `Spring-Boot-Classes`

    애플리케이션 클래스들을 어디서 찾아야 하는지 알려 줍니다.
    여기서는 `BOOT-INF/classes/`입니다.

- `Spring-Boot-Lib`

    라이브러리 jar들을 어디서 찾아야 하는지 알려 줍니다.
    여기서는 `BOOT-INF/lib/`입니다.

위에서 본 파일과 필드를 실행 순서로 다시 놓으면 다음과 같습니다.

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

이 구간에서 기대는 공식 규약은 두 층입니다.

- Java/JAR 규약

    [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)과 [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html)가 `META-INF/MANIFEST.MF`, `Main-Class`, `java -jar` 규칙을 정의합니다.

- Spring Boot executable jar 규약

    [Spring Boot Executable Jar Format](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)과 [Launching Executable Jars](https://docs.spring.io/spring-boot/specification/executable-jar/launching.html)가 `BOOT-INF/classes/`, `BOOT-INF/lib/`, `JarLauncher`, `Start-Class`를 설명합니다.

## 전체 흐름 한 번에 보기

앞에서 본 파일과 필드를 실행 순서로 다시 놓으면 큰 흐름은 다음과 같습니다.

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

이제 터미널 입력부터 이 흐름을 순서대로 정리합니다.

## 셸은 입력을 어떻게 받아 명령으로 바꿀까

터미널 창에 `java -jar SpringBootApp.jar --server.port=8081`를 입력했다고 가정해 보겠습니다.
이때 키 입력을 화면에 보여 주는 것은 터미널 프로그램이고, Enter를 눌러 한 줄이 완성되면 그 줄을 읽어 해석하는 것은 셸입니다.

셸은 이 입력 줄을 그대로 운영 체제에 넘기지 않습니다.
먼저 공백, 따옴표, 이스케이프, 변수 치환 같은 규칙을 적용해서 "실제로 실행할 프로그램 이름"과 "그 프로그램에 넘길 인자 목록"으로 정리합니다.
예를 들어 `"Spring Boot App.jar"`처럼 따옴표를 썼다면, 셸은 공백을 기준으로 둘로 쪼개지 않고 하나의 인자로 묶습니다.

셸이 하는 일을 순서대로 적으면 다음과 같습니다.

1. 입력 줄 읽기: 터미널에서 받은 문자열 한 줄을 셸이 읽습니다.
2. 토큰과 인자 생성: 공백, 따옴표, 이스케이프, 변수 치환 규칙을 적용해서 `argv` 형태의 인자 배열로 바꿉니다.
3. 실행할 대상 탐색: `java`가 셸 builtin인지, 함수인지, 외부 실행 파일인지 판단하고, 외부 실행 파일이면 `PATH`를 따라 실제 경로를 찾습니다.

실제로 셸이 인자를 어떤 모양으로 만들지 보이면 감이 더 잘 잡힙니다.

```text
input : java -jar "Spring Boot App.jar" --server.port=8081
argv  : ['java', '-jar', 'Spring Boot App.jar', '--server.port=8081']
```

즉 셸이 하는 일은 "문자열 한 줄"을 "실행 파일 + 인자 목록"으로 바꾸는 것입니다.
이 단계에서는 아직 JAR를 열지 않았고, Spring Boot를 이해하지도 않았습니다.
셸은 그저 `java`라는 프로그램을 어떤 인자와 함께 실행할지 정리하는 역할을 맡습니다.

로컬에서도 같은 사실을 바로 확인할 수 있습니다.

```sh
command -v java
```

```text
/usr/bin/java
```

이 머신에서는 셸이 먼저 `/usr/bin/java`를 찾습니다.

## 운영 체제는 어떻게 java를 실행할까

셸이 실행 대상을 정하면, 그다음에는 운영 체제가 네이티브 실행 파일을 시작합니다.
유닉스 계열 환경에서는 보통 `fork + exec` 패턴으로 설명합니다.
셸이 자식 프로세스 문맥을 준비하고, `execve(2)` 계열 호출로 새 프로그램 이미지를 덮어씌우는 방식입니다.

여기서 중요한 점은 운영 체제가 JAR 내부 구조를 이해하지 않는다는 것입니다.
운영 체제가 직접 다루는 것은 `SpringBootApp.jar`가 아니라 `java`라는 네이티브 실행 파일입니다.
이 파일이 운영 체제가 정한 실행 파일 포맷에 맞게 만들어져 있으면, 운영 체제는 그 파일을 메모리에 올리고 프로세스를 시작할 수 있습니다.

이 머신에서 직접 확인한 사실은 두 가지입니다.

1. 셸이 찾는 첫 진입점은 `/usr/bin/java`입니다.
2. 현재 JDK home 안에도 별도의 `bin/java`, `libjli.dylib`, `libjvm.dylib`가 있습니다.

```sh
file /usr/bin/java
otool -L /usr/bin/java | sed -n '1,8p'
JH=$(/usr/libexec/java_home)
file "$JH/bin/java"
otool -L "$JH/bin/java" | sed -n '1,8p'
```

```text
/usr/bin/java: Mach-O universal binary with 2 architectures: ...
/usr/bin/java:
    /System/Library/PrivateFrameworks/JavaLaunching.framework/Versions/A/JavaLaunching

/Library/Java/JavaVirtualMachines/graalvm-25.jdk/Contents/Home/bin/java: Mach-O 64-bit executable arm64
/Library/Java/JavaVirtualMachines/graalvm-25.jdk/Contents/Home/bin/java:
    @rpath/libjli.dylib
```

이 출력만 놓고도 계층을 나눠 볼 수 있습니다.

1. 셸은 `/usr/bin/java`를 찾습니다.
2. `/usr/bin/java`는 JAR가 아니라 macOS가 이해하는 Mach-O 실행 파일입니다.
3. 현재 JDK home 안에도 별도의 `bin/java` 실행 파일이 있고, 이 실행 파일은 `libjli.dylib` 같은 공유 라이브러리에 연결됩니다.
4. JDK home 아래에는 `lib/server/libjvm.dylib`도 따로 존재합니다.

즉 `java` 실행은 "JAR 파일 하나를 운영 체제가 곧바로 읽는다"는 그림이 아니라, 네이티브 실행 파일과 공유 라이브러리를 먼저 다루는 그림입니다.
`libjli`는 Java launcher 쪽 진입 라이브러리이고, `libjvm`은 JVM 엔진이 들어 있는 공유 라이브러리라고 보면 됩니다.
운영 체제가 이 라이브러리들을 동적 로딩으로 연결해 프로세스 주소 공간을 준비한 뒤에야, 그 위에서 돌아가는 Java launcher가 JAR와 manifest를 읽기 시작합니다.

여기서 현재 머신 관측과 플랫폼 일반 규칙은 구분해서 보는 편이 좋습니다.

이 실행 파일 포맷은 플랫폼마다 다릅니다.

- Linux에서는 보통 ELF(Executable and Linkable Format)를 봅니다.

    [`elf(5)`](https://man7.org/linux/man-pages/man5/elf.5.html) 문서를 보면 ELF 실행 파일은 ELF header와 program header table 같은 구조를 가진다고 설명합니다.
    운영 체제는 이런 헤더를 읽고 "어디를 메모리에 매핑하고 어디서 실행을 시작할지"를 판단합니다.

- macOS에서는 같은 역할을 Mach-O 포맷이 맡습니다. 이 문서를 작성한 현재 머신은 macOS이므로, 실제 관찰 결과도 ELF가 아니라 Mach-O로 나옵니다.

## `META-INF/MANIFEST.MF`는 무엇이고, 왜 생겼고, 언제부터 있었을까

Java 표준의 [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)은 JAR를 ZIP 기반 파일 포맷으로 설명합니다.
이 문서에 따르면 JAR는 `META-INF` 디렉터리를 가질 수 있고, 그 안의 `MANIFEST.MF`는 JAR 전체에 공통으로 적용되는 메타데이터를 적는 manifest 파일입니다.

즉 `META-INF/MANIFEST.MF`는 Spring Boot가 임의로 만든 파일이 아니라, JAR 규약 안에 있는 표준 위치입니다.
이 파일 안에는 `name: value` 형태의 속성이 들어가고, manifest의 main section에는 JAR 자체와 애플리케이션에 대한 보안 및 설정 정보가 들어갑니다.
executable JAR에서는 `Main-Class` 같은 속성이 시작 경로를 결정합니다.

여기서 manifest라는 말도 같이 정리해 두는 편이 좋습니다.
manifest는 archive 전체에 공통으로 적용되는 메타데이터를 적는 파일이라고 생각하면 됩니다.
archive는 여러 파일을 하나로 묶은 파일이라고 보면 됩니다.

보통은 `jar` 파일을 "클래스를 묶은 압축 파일" 정도로 이해하기 때문에 `META-INF/MANIFEST.MF`를 자세하게 살펴볼 일이 없습니다.
하지만 JAR는 단순 압축 파일이 아니라, 여러 클래스와 리소스를 하나의 배포 단위로 묶고 그 배포 단위에 대한 설명도 함께 담는 형식입니다.
`META-INF/MANIFEST.MF`는 바로 그 설명을 적는 자리입니다.

Oracle의 [JAR File Overview](https://docs.oracle.com/javase/6/docs/technotes/guides/jar/jarGuide.html)에 따르면 applet과 관련 리소스를 한 번의 HTTP transaction으로 내려받게 하여 여러 연결을 열지 않게 하고, 압축과 디지털 서명을 지원하기 위해 JAR 형식의 파일을 지원하게 됐습니다.
여기서 applet과 관련 리소스는, 웹에서 내려받아 실행하던 Java applet 클래스와 그 applet이 함께 필요로 하는 이미지, 오디오, 기타 보조 파일들을 뜻한다고 이해하면 됩니다.

그런데 많은 파일을 하나의 archive로 묶으려면 그 archive 전체에 대한 설명이 필요해집니다.
무엇이 시작 클래스인지, 무엇이 서명되었는지, 어떤 버전 정보가 있는지 같은 정보를 파일마다 흩어 적을 수는 없으니, archive 전체의 공통 메타데이터를 담는 표준 위치가 필요했고 그 자리가 manifest입니다.

`언제부터 있었는가`도 문서로 직접 확인되는 범위까지는 정리할 수 있습니다.

1. JDK 1.1 시점에는 이미 JAR가 보입니다.

    Oracle의 [Security Developer's Guide](https://docs.oracle.com/en/java/javase/21/security/security-developer-guide.pdf)는 "JDK 1.1 introduced the concept of a signed applet"이며, signed applet이 JAR format으로 전달된다고 설명합니다.
    즉 적어도 JDK 1.1 시점에는 JAR가 이미 중요한 배포 형식이었습니다.

2. Java API 차원에서는 `java.util.jar` 패키지가 `Since: 1.2`로 표시됩니다.

    [java.util.jar package summary](https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/jar/package-summary.html)를 보면 JAR 형식과 optional manifest file을 다루는 API가 `Since: 1.2`라고 나옵니다.

3. Java 8, 11, 17, 21, 25 문서에도 JAR File Specification이 계속 존재합니다.

    즉 [Java 8](https://docs.oracle.com/javase/8/docs/technotes/guides/jar/jar.html), [Java 11](https://docs.oracle.com/en/java/javase/11/docs/specs/jar/jar.html), [Java 17](https://docs.oracle.com/en/java/javase/17/docs/specs/jar/jar.html), [Java 21](https://docs.oracle.com/en/java/javase/21/docs/specs/jar/jar.html), [Java 25](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)까지 같은 계보의 규약이 이어집니다.

그러면 "그 전에는 어떻게 했을까?"라는 질문도 자연스럽게 생깁니다.
여기서는 두 층으로 나눠 보는 편이 좋습니다.

1. applet 배포 관점

    위의 JAR File Overview는 archive에서 찾지 못한 파일은 서버에서 `as in JDK1.0.2` 방식으로 다시 찾는다고 설명합니다.
    즉 JDK1.0 계열에서는 지금처럼 JAR archive 하나를 중심에 두기보다, CODEBASE 기준으로 클래스와 리소스를 개별적으로 찾는 감각이 더 강했습니다.
    JAR는 이런 방식을 archive 중심 배포로 묶어 주는 쪽으로 발전한 것입니다.

2. local application 실행 관점

    지금처럼 `java -jar app.jar`로 archive 자체를 실행 단위로 보는 방식은 manifest의 `Main-Class` 같은 규약이 있어야 자연스럽습니다.
    그 전에는 보통 classpath를 맞추고 main class 이름을 직접 지정하는 방식이 더 기본에 가까웠습니다.

Oracle의 [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html) 문서는 `java -jar jarfile`를 다음 뜻으로 설명합니다.

1. 실행 단위는 class 이름이 아니라 JAR 파일입니다.
2. 그 JAR의 manifest에는 `Main-Class`가 있어야 합니다.
3. `-jar`를 사용할 때는 지정한 JAR 파일이 user classes의 source가 됩니다.

즉 `java -jar app.jar`라고 입력했을 때 Java launcher는 "JAR를 열고, manifest를 읽고, `Main-Class`가 가리키는 클래스를 시작점으로 잡는 규칙"을 이미 알고 있습니다.
현재 JAR specification 문서도 `Main-Class`를 stand-alone application을 `java -jar x.jar`로 직접 실행할 때 쓰는 속성으로 설명합니다.

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

따라서 sample manifest 안에 `Main-Class`와 `Start-Class`가 둘 다 있는 이유는 Java launcher와 Spring Boot loader가 읽는 규칙이 다르기 때문입니다.

## 일반 실행형 JAR와 Spring Boot executable jar

여기까지 정리한 내용을 기준으로 일반 실행형 JAR와 Spring Boot executable jar를 나눠 보면 차이가 더 분명해집니다.

1. 일반 실행형 JAR

    manifest의 `Main-Class`가 곧 애플리케이션 시작 클래스입니다.
    Java launcher는 그 클래스를 로드하고, 그 클래스의 `main()`을 호출합니다.

2. Spring Boot executable jar

    manifest의 `Main-Class`는 보통 `org.springframework.boot.loader.launch.JarLauncher`입니다.
    실제 애플리케이션 시작 클래스는 `Start-Class`에 들어 있습니다.

    즉 Java launcher가 내 애플리케이션 클래스로 바로 들어가는 것이 아니라, 먼저 `JarLauncher`를 실행합니다.
    `JarLauncher`는 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 읽어 실행 가능한 classpath를 준비한 뒤 `Start-Class`의 `main()`을 호출합니다.

두 실행 경로를 나란히 놓으면 다음과 같습니다.

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
라이브러리들이 평평한 classpath 바깥이 아니라 `BOOT-INF/lib/` 아래 nested jar 형태로 들어 있기 때문에, Java launcher가 곧바로 애플리케이션 `main()`으로 들어가면 아직 실행 경로가 준비되지 않습니다.
그래서 `JarLauncher`가 중간에 들어가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 먼저 해석합니다.

## JarLauncher는 실제로 무엇을 할까

Spring Boot 공식 문서의 [Launching Executable Jars](https://docs.spring.io/spring-boot/specification/executable-jar/launching.html)는 `Launcher`를 executable jar의 main entry point라고 설명합니다.
즉 manifest의 `Main-Class` 자리에 들어가는 쪽이 `Launcher`이고, sample jar에서는 그 역할을 `JarLauncher`가 맡습니다.
같은 문서는 `JarLauncher`가 `BOOT-INF/lib/`를 보고 classpath를 구성하고, `Start-Class`를 실제 시작 클래스로 사용한다고 설명합니다.

즉 `JarLauncher`는 단순히 "한 번 거쳐 가는 클래스"가 아닙니다.
Spring Boot executable jar 구조를 실제로 실행 가능한 Java classpath로 번역하는 bootstrap 담당자입니다.

`JarLauncher`가 하는 일을 순서대로 적으면 다음과 같습니다.

1. 지금 실행 중인 archive가 Spring Boot executable jar 구조인지 전제하고 내부 경로를 해석합니다.

    `BOOT-INF/classes/`에 애플리케이션 클래스가 있고, `BOOT-INF/lib/`에 라이브러리 jar들이 있다는 규약을 사용합니다.

2. 그 구조를 바탕으로 classpath와 class loader를 준비합니다.

    Java launcher는 manifest의 `Main-Class`까지는 알지만, nested jar가 들어 있는 Spring Boot 내부 구조를 스스로 일반 classpath처럼 풀어 주지는 않습니다.
    이 부분을 `JarLauncher`와 Spring Boot loader가 맡습니다.

3. `Start-Class`를 찾아 실제 애플리케이션의 `main()`을 호출합니다.

    여기서부터 비로소 우리가 평소 보던 애플리케이션 시작 클래스 코드가 실행됩니다.

따라서 `JarLauncher`의 역할은 "내 애플리케이션을 대신 실행해 주는 프레임워크"라기보다, "Spring Boot jar 구조를 일반적인 Java 실행 흐름으로 연결하는 bootstrap bridge"에 가깝습니다.

## `SpringApplication.run()`은 언제 호출될까

보통 스프링 부트 애플리케이션 시작 클래스의 `main()`은 다음과 같이 생겼습니다.

```java
public static void main(String[] args) {
    SpringApplication.run(MyApplication.class, args);
}
```

애플리케이션 코드에서 처음 눈에 띄는 줄은 이 `main()`입니다.
하지만 프로세스 전체 타임라인에서 보면 이 줄이 맨 앞은 아닙니다.
Spring Boot executable jar에서는 `java launcher -> manifest -> JarLauncher -> Start-Class 연결`이 끝난 후 이 `main()`이 실행됩니다.

즉 `SpringApplication.run()`은 "프로세스가 처음 생기는 순간"이 아니라, 이미 Java 프로세스가 살아 있고, manifest가 해석되었고, Spring Boot loader가 classpath를 준비한 뒤에 호출됩니다.

여기서부터는 JAR 파일 구조를 푸는 단계가 아니라, 실제 Spring framework startup 단계입니다.
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

Spring Boot executable jar에서는 `main()`도 이미 loader bootstrap 뒤에 도착한 시작점입니다.
따라서 `main()`이 시작점처럼 보이더라도, 그 앞단에는 packaging/bootstrap 단계가 더 있습니다.

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

Spring Boot 공식 문서는 이 차이를 이벤트 순서로 더 잘게 나눠 보여 줍니다.

1. `ContextRefreshedEvent`

    `ApplicationContext` refresh가 끝난 시점입니다.

2. `WebServerInitializedEvent`

    내장 웹 서버가 준비된 시점입니다.

3. `ApplicationStartedEvent`

    context refresh 이후이지만 `ApplicationRunner`, `CommandLineRunner` 호출 전입니다.

4. `ApplicationReadyEvent`

    runner 호출까지 끝난 뒤입니다.

5. `AvailabilityChangeEvent(ReadinessState.ACCEPTING_TRAFFIC)`

    Spring Boot가 요청을 받아도 되는 상태라고 선언하는 시점입니다.

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
2026-04-14T23:03:41.481+09:00  INFO 73999 --- [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat initialized with port 0 (http)
2026-04-14T23:03:41.925+09:00  INFO 73999 --- [           main] o.s.b.w.embedded.tomcat.TomcatWebServer  : Tomcat started on port 63272 (http) with context path '/'
2026-04-14T23:03:41.936+09:00  INFO 73999 --- [           main] c.e.portable.FullyPortableApplication    : Started FullyPortableApplication in 1.458 seconds (process running for 1.862)
```

이 로그를 읽는 기준은 다음과 같습니다.

1. `Tomcat initialized with port 0`

    웹 서버 초기화가 시작되었다는 뜻입니다.

2. `Tomcat started on port ...`

    실제 포트 바인딩이 끝났다는 강한 관측 신호입니다.

3. `Started FullyPortableApplication in ...`

    Spring Boot가 애플리케이션 시작 완료 로그를 남긴 시점입니다.
    다만 이 로그 한 줄만으로 `ApplicationReadyEvent`를 직접 본 것은 아닙니다.

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

- [JAR File Overview](https://docs.oracle.com/javase/6/docs/technotes/guides/jar/jarGuide.html)
- [java.util.jar package summary](https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/jar/package-summary.html)
- [Security Developer's Guide](https://docs.oracle.com/en/java/javase/21/security/security-developer-guide.pdf)
- [JAR File Specification - Java 8](https://docs.oracle.com/javase/8/docs/technotes/guides/jar/jar.html)
- [JAR File Specification - Java 11](https://docs.oracle.com/en/java/javase/11/docs/specs/jar/jar.html)
- [JAR File Specification - Java 17](https://docs.oracle.com/en/java/javase/17/docs/specs/jar/jar.html)
- [JAR File Specification - Java 21](https://docs.oracle.com/en/java/javase/21/docs/specs/jar/jar.html)
- [JAR File Specification](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html)
- [The java Command](https://docs.oracle.com/en/java/javase/25/docs/specs/man/java.html)
- [Spring Boot Executable Jar Format](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)
- [Launching Executable Jars](https://docs.spring.io/spring-boot/specification/executable-jar/launching.html)
- [SpringApplication](https://docs.spring.io/spring-boot/reference/features/spring-application.html)
- [`execve(2)` man page](https://man7.org/linux/man-pages/man2/execve.2.html)
- [`elf(5)` man page](https://man7.org/linux/man-pages/man5/elf.5.html)
- 샘플 jar: [fully-portable-spring-boot-0.0.1-SNAPSHOT.jar](/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar)
