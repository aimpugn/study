# Spring Boot executable jar startup

- [개요](#개요)
- [작은 실행 흐름 먼저 보기](#작은-실행-흐름-먼저-보기)
- [일반 실행형 JAR와 Spring Boot executable jar](#일반-실행형-jar와-spring-boot-executable-jar)
- [셸과 운영 체제는 어디까지 담당할까](#셸과-운영-체제는-어디까지-담당할까)
- [Java launcher와 manifest는 무엇을 결정할까](#java-launcher와-manifest는-무엇을-결정할까)
- [JarLauncher는 실제로 무엇을 할까](#jarlauncher는-실제로-무엇을-할까)
- [실제 애플리케이션 main과 SpringApplication.run](#실제-애플리케이션-main과-springapplicationrun)
- [언제 애플리케이션이 떴다고 볼 수 있을까](#언제-애플리케이션이-떴다고-볼-수-있을까)
- [직접 확인해 보기](#직접-확인해-보기)
- [정리](#정리)
- [근거와 추가 읽기](#근거와-추가-읽기)

## 개요

여기서 터미널과 쉘을 통해 `java -jar SpringBootApp.jar`를 명령하면 실제로 누가 무엇을 읽고, 어느 순간부터 "이 애플리케이션이 떴다"라고 판단할 수 있는지 정리합니다.

운영 체제는 곧바로 내 애플리케이션의 `main()`을 실행하지 않습니다.
먼저 `java`라는 네이티브 프로그램이 시작되고, Java launcher가 JAR의 manifest를 읽고, Spring Boot executable jar라면 `JarLauncher`가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 해석한 뒤 실제 애플리케이션 시작 클래스까지 연결합니다.
그 다음에야 `SpringApplication.run()`이 실행되고, `ApplicationContext`, 내장 웹 서버, readiness 단계가 차례로 닫힙니다.

## 작은 실행 흐름 먼저 보기

작게 줄이면 실행 흐름은 다음과 같습니다.

```text
shell
    -> OS가 java 실행 파일을 시작
    -> Java launcher가 -jar 대상 JAR을 연다
    -> META-INF/MANIFEST.MF에서 Main-Class를 읽는다
    -> Spring Boot jar라면 JarLauncher가 BOOT-INF/classes/와 BOOT-INF/lib/를 기준으로 classpath를 준비한다
    -> Start-Class의 main()을 호출한다
    -> SpringApplication.run()
    -> ApplicationContext refresh
    -> 내장 웹 서버 시작
    -> runner / readiness 단계
```

이 문서에서는 이 흐름을 기준으로, 터미널에서 `java -jar SpringBootApp.jar`를 입력했을 때 어떤 계층이 어디까지 담당하고 어떤 시점부터 "애플리케이션이 실행되었다"고 볼 수 있는지 따라갑니다.

## 일반 실행형 JAR와 Spring Boot executable jar

우선 중요한 것은 `java -jar app.jar`라고 명령할 때 `app.jar` 안에 있는 내가 개발한 클래스 하나를 곧바로 실행하지 않는다는 점입니다.
먼저 `java` 프로그램이 실행되고, 그 프로그램이 JAR 안의 manifest를 읽어 시작점을 결정합니다.

이때 일반적인 실행형 JAR와 Spring Boot executable jar로 나뉩니다.

1. 일반 실행형 JAR

    manifest의 `Main-Class`가 곧 애플리케이션 시작 클래스입니다.
    Java launcher는 이 클래스를 로드하고, 그 클래스의 `main()`을 호출합니다.

2. Spring Boot executable jar

    manifest의 `Main-Class`는 보통 `org.springframework.boot.loader.launch.JarLauncher`입니다.
    실제 애플리케이션 시작 클래스는 `Start-Class`에 들어 있습니다.

    즉 Java launcher가 곧바로 내 애플리케이션 클래스로 들어가는 것이 아니라, 먼저 `JarLauncher`를 실행합니다.
    `JarLauncher`는 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 읽어 classpath를 준비한 뒤 `Start-Class`의 `main()`을 호출합니다.

    이 흐름을 짧게 쓰면 다음과 같습니다.

    ```text
    Java launcher
        -> Main-Class = JarLauncher
        -> JarLauncher가 BOOT-INF/classes/와 BOOT-INF/lib/를 읽음
        -> Start-Class 확인
        -> Start-Class의 main() 호출
    ```

일반 실행형 JAR는 `Main-Class`가 바로 애플리케이션 시작 클래스입니다.
반면 Spring Boot executable jar는 `Main-Class`와 실제 애플리케이션 시작 클래스가 다릅니다.
즉 `Main-Class`인 `JarLauncher`가 실행된 뒤, `Start-Class`의 `main()`이 호출되기 전에는 Spring Boot loader가 내부 JAR 구조를 해석하고 classpath를 준비하는 단계가 실제로 존재합니다.

## 셸과 운영 체제는 어디까지 담당할까

유닉스 계열 환경에서 셸은 보통 자식 프로세스를 만들고 그 자식에게 새 프로그램 이미지를 덮어씌우는 방식으로 명령을 실행합니다.
주로 `fork + exec` 패턴으로 실행됩니다.

여기서 핵심은 `execve(2)`입니다.
`execve(2)`는 현재 프로세스 이미지를 새 프로그램 이미지로 바꾸는 시스템 콜입니다.
[`execve(2)` man page](https://man7.org/linux/man-pages/man2/execve.2.html)에서는 "current process image를 new process image로 replace한다"고 설명합니다.

이 관점에서 보면 운영 체제가 직접 이해하는 것은 Spring Boot가 아니라 `java` 실행 파일입니다.
운영 체제는 `java`라는 네이티브 프로그램을 시작할 뿐이고, `-jar`, `META-INF/MANIFEST.MF`, `Main-Class`, `JarLauncher`, `Start-Class` 같은 규칙은 그 뒤에 실행되는 Java launcher와 Spring Boot loader가 처리합니다.

즉 운영 체제 관점에서의 경계는 다음처럼 정리할 수 있습니다.

1. 셸이 명령을 해석합니다.
2. 운영 체제가 `java` 실행 파일을 시작합니다.
3. 그 다음부터는 `java` 프로세스가 JAR을 해석하며 실행을 이어갑니다.

여기서 운영 체제가 JAR 내부 구조까지 이해한다고 생각하면 뒤 설명이 전부 흐려집니다.

## Java launcher와 manifest는 무엇을 결정할까

로컬에서 `java --help`를 보면 다음처럼 나옵니다.

```text
Usage: java [options] -jar <jarfile>.jar [args...]
       (to execute a jar file)
```

즉 `-jar`는 "클래스 이름을 직접 주는 대신, JAR 파일을 실행 단위로 삼겠다"는 뜻입니다.
이 모드에서 Java launcher는 JAR 안의 `META-INF/MANIFEST.MF`를 읽고 시작 클래스를 결정합니다.

일반 실행형 JAR라면 여기서 `Main-Class`를 읽고 바로 그 클래스의 `main()`으로 들어갑니다.
Spring Boot executable jar라면 여기서 읽히는 `Main-Class`가 `JarLauncher`이고, 실제 애플리케이션 클래스는 `Start-Class`에 들어 있습니다.

즉 manifest는 단순 메모가 아니라 시작 경로를 결정하는 실제 실행 메타데이터입니다.
Spring Boot executable jar에서는 다음 두 항목을 같이 봐야 전체 흐름이 보입니다.

1. `Main-Class`

    Java launcher가 가장 먼저 실행할 클래스입니다.
    Spring Boot executable jar에서는 보통 `JarLauncher`입니다.

2. `Start-Class`

    `JarLauncher`가 classpath를 준비한 뒤 최종적으로 넘겨 주는 실제 애플리케이션 시작 클래스입니다.

이 두 항목을 구분하지 않으면 "왜 `java -jar`를 했는데 내 앱 클래스가 아니라 loader가 먼저 실행되느냐"를 이해하기 어렵습니다.

## JarLauncher는 실제로 무엇을 할까

Spring Boot executable jar의 핵심은 "실행 가능한 단일 JAR 파일"이라는 배포 형식입니다.
애플리케이션 클래스와 라이브러리들을 한 파일에 묶고, `java -jar app.jar` 한 줄로 실행할 수 있어야 합니다.

문제는 라이브러리들이 JAR 안의 `BOOT-INF/lib/` 밑에 nested jar 형태로 들어 있다는 점입니다.
Java launcher는 이 구조를 일반 classpath처럼 바로 풀어 주지 않습니다.
그래서 Spring Boot는 자기 loader를 함께 넣습니다.

저장소 안의 실제 예제 jar를 열어 보면 이 구조를 바로 볼 수 있습니다.

```sh
jar tf jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar | sed -n '1,80p'
```

앞부분에는 `org/springframework/boot/loader/...`가 있고, 뒤에는 `BOOT-INF/classes/`, `BOOT-INF/lib/`가 함께 보입니다.
즉 이 JAR 파일은 "내 앱 클래스만 들어 있는 JAR"이 아니라 "loader + app classes + nested libs" 구조입니다.

manifest도 직접 확인할 수 있습니다.

```sh
unzip -p jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
```

이 저장소의 샘플 jar에서는 실제로 다음과 같이 나옵니다.

```text
Main-Class: org.springframework.boot.loader.launch.JarLauncher
Start-Class: com.example.portable.FullyPortableApplication
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
```

여기서 `JarLauncher`가 실제로 하는 일은 크게 세 가지입니다.

1. 실행 중인 JAR가 Spring Boot executable jar 구조인지 전제하고 내부 경로를 해석합니다.

2. 애플리케이션 클래스는 `BOOT-INF/classes/`에서, 라이브러리는 `BOOT-INF/lib/`에서 찾도록 classpath를 준비합니다.

3. `Start-Class`를 찾아 그 클래스의 `main()`을 호출합니다.

즉 `JarLauncher`는 "실제 애플리케이션을 실행하기 전에 필요한 실행 환경을 맞추는 준비 담당자"입니다.
이 준비가 끝나야 비로소 우리가 평소 보던 `SpringApplication.run()`이 호출됩니다.

## 실제 애플리케이션 main과 SpringApplication.run

이제부터는 Spring Boot framework 쪽 설명입니다.
보통 애플리케이션 시작 클래스의 `main()`은 다음처럼 생겼습니다.

```java
public static void main(String[] args) {
    SpringApplication.run(MyApplication.class, args);
}
```

이 코드는 시작점처럼 보이지만, Spring Boot executable jar에서는 이미 앞단의 `java launcher -> manifest -> JarLauncher -> Start-Class 연결`이 끝난 뒤에 실행됩니다.

그 다음 `SpringApplication.run()`이 이어집니다.
저장소의 예제 코드 [SpringMain.kt](../examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt) 주석을 기준으로 줄이면 다음 순서로 이해할 수 있습니다.

1. bootstrap context, listener, initializer를 준비합니다.

2. environment를 준비합니다.

3. `ApplicationContext`를 생성합니다.

4. bean definition을 로드하고 context를 준비합니다.

5. `ApplicationContext`를 refresh 합니다.

6. 웹 애플리케이션이면 refresh 과정에서 내장 웹 서버를 초기화하고 시작합니다.

이 흐름에서 중요한 것은 `JarLauncher`와 `SpringApplication.run()`의 역할이 다르다는 점입니다.

1. `JarLauncher`

    Spring Boot executable jar 내부 구조를 해석하고, 실행 가능한 classpath를 준비합니다.

2. `SpringApplication.run()`

    실제 Spring 애플리케이션 세계를 세웁니다.
    environment, configuration, bean, `ApplicationContext`, 내장 웹 서버, runner, readiness가 여기서 연결됩니다.

즉 `JarLauncher`는 "실행 파일 구조를 푸는 단계"이고, `SpringApplication.run()`은 "애플리케이션을 실제로 기동하는 단계"입니다.

## 언제 애플리케이션이 떴다고 볼 수 있을까

이 질문은 어느 계층에서 보느냐에 따라 답이 다릅니다.

1. 운영 체제 관점

    `java` 프로세스가 시작된 순간 이미 실행 중이라고 볼 수 있습니다.
    다만 이 시점은 Spring 애플리케이션이 준비되었다는 뜻은 아닙니다.

2. Java launcher 관점

    manifest를 읽고 시작 경로를 따라 들어가기 시작한 시점부터 JAR 실행이 진행 중이라고 볼 수 있습니다.
    역시 이 시점은 애플리케이션 준비 완료와는 다릅니다.

3. Spring context 관점

    `ApplicationContext`가 refresh 되면 Spring 컨테이너는 살아 있다고 볼 수 있습니다.
    하지만 웹 서버와 runner까지 다 끝났는지는 따로 봐야 합니다.

4. 트래픽 수신 관점

    내장 웹 서버가 포트를 바인드하고, startup runner까지 끝나 readiness가 올라와야 "요청을 받아도 된다"고 말하는 편이 더 정확합니다.

Spring Boot 공식 문서도 started와 ready를 구분합니다.
문서에 따르면 `ApplicationStartedEvent`는 context refresh 뒤, 하지만 `ApplicationRunner`와 `CommandLineRunner` 호출 전입니다.
반면 `ApplicationReadyEvent`는 runner 호출 뒤이고, readiness는 이 뒤에 `ReadinessState.ACCEPTING_TRAFFIC`로 올라갑니다.

즉 다음 네 문장은 서로 같은 말이 아닙니다.

1. 프로세스가 떴다.
2. 컨텍스트가 refresh 됐다.
3. 웹 서버가 포트를 바인드했다.
4. 트래픽을 받아도 된다.

이 차이를 구분해야 startup hook, warm-up, health check, readiness probe를 정확하게 다룰 수 있습니다.

실제로 저장소 안 샘플 jar를 실행하면 이런 로그가 보입니다.

```text
Tomcat initialized with port 0 (http)
Starting Servlet engine: [Apache Tomcat/10.1.30]
Root WebApplicationContext: initialization completed in 539 ms
Tomcat started on port 60324 (http) with context path '/'
Started FullyPortableApplication in 1.123 seconds
```

이 로그는 적어도 다음 두 가지를 보여 줍니다.

1. `Tomcat started on port ...`

    내장 웹 서버가 실제로 포트를 바인드했다는 강한 관측 신호입니다.

2. `Started FullyPortableApplication ...`

    Spring Boot가 startup을 거의 마무리했다는 높은 수준의 관측 신호입니다.

다만 이 로그 한 줄만 보고 "이제 startup 후처리까지 완전히 끝났다"고 단정하면 안 됩니다.
runner와 readiness는 별도로 구분해서 보는 편이 안전합니다.

## 직접 확인해 보기

이 주제는 문서만 읽기보다 직접 한 번 확인하는 편이 훨씬 잘 남습니다.

1. `-jar`가 JAR 실행 모드인지 확인합니다.

    ```sh
    java --help | sed -n '1,12p'
    ```

    PASS:
    `java [options] -jar <jarfile>.jar`가 보입니다.

    FAIL:
    `-jar` 항목이 보이지 않거나, 기대한 JDK/JRE가 아닙니다.

2. Spring Boot executable jar 구조를 확인합니다.

    ```sh
    jar tf jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar | sed -n '1,80p'
    ```

    PASS:
    `org/springframework/boot/loader/`와 `BOOT-INF/classes/`, `BOOT-INF/lib/`가 같이 보입니다.

    FAIL:
    일반 JAR처럼 앱 클래스만 보이고 loader와 `BOOT-INF` 구조가 없습니다.

3. manifest에서 `Main-Class`와 `Start-Class`를 확인합니다.

    ```sh
    unzip -p jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
    ```

    PASS:
    `Main-Class: org.springframework.boot.loader.launch.JarLauncher`와 `Start-Class: ...`가 함께 보입니다.

    FAIL:
    `Main-Class`만 있고 `Start-Class`가 없거나, 기대한 Spring Boot executable jar가 아닙니다.

4. 실제 startup 로그를 확인합니다.

    ```sh
    java -jar jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar --server.port=0
    ```

    PASS:
    `Tomcat started on port ...`와 `Started ...` 로그를 순서대로 볼 수 있습니다.

    FAIL:
    startup 중 예외가 나거나, 포트 바인드 이전에 종료됩니다.

## 정리

`java -jar SpringBootApp.jar`는 "JAR 안의 내 `main()`을 그냥 실행한다"는 단순한 동작이 아닙니다.
운영 체제는 먼저 `java`라는 네이티브 프로그램을 시작합니다.
그 다음 Java launcher가 manifest를 읽고, Spring Boot executable jar라면 `JarLauncher`가 `BOOT-INF/classes/`와 `BOOT-INF/lib/`를 바탕으로 classpath를 준비한 뒤 `Start-Class`의 `main()`을 호출합니다.
그리고 나서 `SpringApplication.run()`이 environment, `ApplicationContext`, 내장 웹 서버, runner, readiness를 차례로 닫아 갑니다.

이 흐름을 이해하면 다음 질문들이 훨씬 쉬워집니다.
왜 Spring Boot jar 안에 loader 클래스가 들어 있는지, 왜 일반 실행형 JAR와 구조가 다른지, 왜 startup 로그의 각 줄이 서로 다른 계층의 신호인지, 그리고 언제부터 health check나 readiness probe를 신뢰할 수 있는지 자연스럽게 이어집니다.

## 근거와 추가 읽기

- source reservoir: `.tmp/interviews2.md`
- repo example: [SpringMain.kt](../examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt)
- related hub: [spring_init.md](./spring_init.md)
- Spring Boot executable jar format: [docs.spring.io](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)
- Spring Boot application lifecycle and availability: [docs.spring.io](https://docs.spring.io/spring-boot/4.1/reference/features/spring-application.html)
- `execve(2)` background: [man7.org](https://man7.org/linux/man-pages/man2/execve.2.html)
