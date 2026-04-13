# Spring Boot executable jar startup

`java -jar SpringBootApp.jar`를 입력하면, 운영 체제가 곧바로 내 애플리케이션의 `main()`을 실행하는 것은 아닙니다.
먼저 셸이 `java` 실행 파일을 띄우고, Java launcher가 JAR의 `META-INF/MANIFEST.MF`를 읽어 어디를 시작점으로 삼을지 결정합니다.
일반 실행형 JAR이라면 여기서 바로 애플리케이션의 `Main-Class`로 들어가지만, Spring Boot executable jar는 한 단계가 더 있습니다.
manifest의 `Main-Class`가 보통 내 애플리케이션 클래스가 아니라 `org.springframework.boot.loader.launch.JarLauncher`이고, 이 loader가 `BOOT-INF/classes`와 `BOOT-INF/lib`를 읽어 실제 애플리케이션 시작 클래스(`Start-Class`)를 찾아 실행합니다.

짧게 줄이면 흐름은 이렇습니다.

```text
shell
  -> OS가 java 실행 파일을 시작
  -> Java launcher가 -jar 대상 JAR을 연다
  -> MANIFEST.MF에서 Main-Class를 읽는다
  -> Spring Boot jar면 JarLauncher가 BOOT-INF/...를 classpath로 조립한다
  -> Start-Class의 main() 호출
  -> SpringApplication.run()
  -> ApplicationContext refresh
  -> 내장 웹 서버 시작
  -> runners / readiness 단계
```

이 문서의 핵심 질문은 하나입니다.
`java -jar SpringBootApp.jar`를 치면 실제로 누가 무엇을 읽고, 어느 순간부터 "이 애플리케이션이 떴다"라고 말할 수 있을까?

## 가장 작은 모델부터 잡기

가장 먼저 머릿속에 고정해야 할 것은 `java -jar app.jar`가 "`app.jar` 안의 내 클래스 하나를 그냥 실행한다"는 모델이 아니라는 점입니다.
정확한 모델은 "`java` 프로그램이 먼저 뜨고, 그 프로그램이 JAR을 해석해서 시작점을 찾는다"에 더 가깝습니다.

여기서 일반 실행형 JAR과 Spring Boot executable jar가 갈립니다.

- 일반 실행형 JAR
  - manifest의 `Main-Class`가 곧 애플리케이션 시작 클래스다.
- Spring Boot executable jar
  - manifest의 `Main-Class`는 대개 `JarLauncher`다.
  - 실제 애플리케이션 시작 클래스는 `Start-Class`에 들어 있다.
  - 그 사이에 Spring Boot loader가 nested jar와 classpath를 정리하는 bootstrap 단계가 하나 더 있다.

이 작은 차이를 이해하지 못하면, 왜 Spring Boot jar 안에 `org/springframework/boot/loader/...`가 들어 있는지, 왜 `BOOT-INF/lib` 밑에 라이브러리 jar가 또 들어 있는지, 왜 단순한 fat jar와 Spring Boot executable jar를 같은 것으로 보면 안 되는지가 전부 흐려집니다.

## 셸과 운영 체제는 어디까지 담당할까

유닉스 계열 환경에서 셸은 보통 자식 프로세스를 만들고 그 자식에게 새 프로그램 이미지를 덮어씌우는 방식으로 명령을 실행합니다.
실무에서 자주 함께 언급되는 말이 `fork + exec`인 이유가 이것입니다.
다만 여기서 정말 핵심인 것은 `fork`보다 `execve`입니다.
`execve(2)`는 현재 프로세스 이미지를 새 프로그램 이미지로 바꾸는 시스템 콜이고, man page도 이를 "current process image를 new process image로 replace한다"는 식으로 설명합니다.

즉 `java -jar ...`의 핵심 OS 경계는 다음처럼 이해하는 편이 안전합니다.

1. 셸이 명령을 해석한다.
2. 운영 체제가 `java` 실행 파일을 시작한다.
3. 그 시점부터는 "JAR이 실행되는 것"이 아니라 "java launcher 프로세스가 JAR을 해석하며 실행을 이어가는 것"이다.

이 차이가 중요한 이유는, 운영 체제는 여기서 "JAR 안에 있는 Spring Boot 구조"를 이해하지 못하기 때문입니다.
운영 체제는 그저 `java`라는 네이티브 프로그램을 실행할 뿐이고, `-jar`, manifest, `Main-Class`, `JarLauncher`, `Start-Class` 같은 규칙은 JVM launcher와 Spring Boot loader가 담당합니다.

## `-jar`가 실제로 뜻하는 것

로컬에서 `java --help`를 보면 다음처럼 나옵니다.

```text
Usage: java [options] -jar <jarfile>.jar [args...]
       (to execute a jar file)
```

즉 `-jar`는 "클래스 이름을 직접 주는 대신, JAR 파일을 실행 단위로 삼겠다"는 뜻입니다.
이 모드에서 Java launcher는 JAR 내부의 `META-INF/MANIFEST.MF`를 읽고 시작 클래스를 결정합니다.

일반 실행형 JAR의 mental model은 여기서 끝나도 됩니다.
manifest의 `Main-Class`가 곧 `public static void main(String[] args)`를 가진 애플리케이션 클래스이기 때문입니다.

하지만 Spring Boot executable jar는 이 지점에서 한 번 더 꺾입니다.
manifest의 시작 클래스가 내 앱이 아니라 loader이기 때문입니다.

## Spring Boot jar는 왜 loader가 하나 더 필요한가

Spring Boot executable jar는 "클래스 파일 몇 개와 라이브러리 몇 개를 나란히 압축한 단순 JAR"이 아닙니다.
애플리케이션 클래스는 `BOOT-INF/classes/` 아래에, 의존 라이브러리는 `BOOT-INF/lib/` 아래에 다시 JAR 형태로 들어갑니다.

이 구조가 필요한 이유는 배포 편의성입니다.
실행에 필요한 것을 JAR 하나에 묶어 두고 `java -jar app.jar` 한 줄로 실행하고 싶기 때문입니다.
문제는 Java 표준 launcher가 "JAR 안에 또 들어 있는 nested JAR들"을 일반 classpath처럼 바로 다뤄 주지는 않는다는 점입니다.
그래서 Spring Boot는 자기 loader를 같이 넣습니다.

Spring Boot 공식 executable jar 문서도 이 구조를 전제로 설명합니다.
`JarLauncher`는 `BOOT-INF/lib/`를 보고, 애플리케이션 클래스는 `BOOT-INF/classes/`에서 찾도록 설계되어 있습니다.

저장소 안의 실제 예제 jar를 확인해 보면 이 차이가 아주 선명합니다.

```sh
jar tf jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar | sed -n '1,40p'
```

앞부분을 보면 `org/springframework/boot/loader/...`와 `BOOT-INF/...`가 같이 들어 있습니다.
즉 "내 앱 클래스만 들어 있는 JAR"이 아니라 "loader + app classes + nested libs"를 한 번에 묶은 구조입니다.

manifest도 직접 확인할 수 있습니다.

```sh
unzip -p jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
```

이 예제 jar에서는 실제로 이렇게 나옵니다.

```text
Main-Class: org.springframework.boot.loader.launch.JarLauncher
Start-Class: com.example.portable.FullyPortableApplication
Spring-Boot-Classes: BOOT-INF/classes/
Spring-Boot-Lib: BOOT-INF/lib/
```

여기서 꼭 구분해야 할 것이 있습니다.

- `Main-Class`
  - Java launcher가 가장 먼저 호출하는 진입점
  - Spring Boot executable jar에서는 보통 `JarLauncher`
- `Start-Class`
  - loader가 classpath를 준비한 뒤 최종적으로 넘겨 주는 실제 애플리케이션 시작 클래스

즉 Spring Boot jar에서 "누가 main을 결정하느냐"는 질문의 답은 한 줄이 아닙니다.
먼저 Java launcher가 manifest의 `Main-Class`를 읽고, 그 `Main-Class`인 `JarLauncher`가 다시 `Start-Class`를 읽어 실제 앱의 `main()`으로 연결합니다.

## 실제 실행 경로를 단계별로 쌓아 보기

이제 `java -jar SpringBootApp.jar`가 실제로 어떻게 올라가는지를 아래 순서로 보면 됩니다.

### 1. 셸이 `java`를 실행한다

이 단계의 주인공은 Spring이 아니라 운영 체제와 Java launcher입니다.
셸은 사용자가 친 `java -jar app.jar ...`를 해석하고, 운영 체제는 `java`라는 네이티브 실행 파일을 시작합니다.

이 시점에서 생기는 프로세스는 "Spring Boot 프로세스"가 아니라 "JVM launcher 프로세스"라고 보는 편이 더 정확합니다.
Spring Boot는 아직 시작하지도 않았습니다.

### 2. Java launcher가 `-jar` 대상과 manifest를 읽는다

`-jar` 모드에서는 launcher가 JAR 파일을 실행 단위로 취급합니다.
그리고 `META-INF/MANIFEST.MF`에서 시작 클래스를 확인합니다.

일반 JAR이면 여기서 바로 앱 클래스로 들어가지만, Spring Boot executable jar면 `JarLauncher`로 들어갑니다.

### 3. `JarLauncher`가 Spring Boot jar 구조를 해석한다

이 단계가 Spring Boot executable jar의 핵심 차별점입니다.
`JarLauncher`는 다음 두 가지를 해결합니다.

1. 애플리케이션 클래스는 `BOOT-INF/classes/`에서 찾는다.
2. 의존 라이브러리는 `BOOT-INF/lib/` 아래 nested jar들에서 찾는다.

즉 여기서 classpath를 "한 덩어리 JAR 바깥에 lib 폴더를 나란히 두는 방식"이 아니라 "실행 JAR 내부 구조를 읽어 조립하는 방식"으로 만듭니다.
이 준비가 끝나야 비로소 실제 `Start-Class`의 `main()`을 호출할 수 있습니다.

### 4. 실제 애플리케이션 `main()`이 호출된다

여기서부터가 우리가 평소 보던 코드입니다.

```java
public static void main(String[] args) {
    SpringApplication.run(MyApplication.class, args);
}
```

많은 사람이 `java -jar`를 곧 이 코드와 동일시하지만, 실제로는 그 앞에 "launcher -> manifest -> loader -> classpath bootstrap" 단계가 이미 지나갔습니다.

### 5. `SpringApplication.run()`이 환경과 컨텍스트를 준비한다

이제 Spring Boot framework 레벨로 올라옵니다.
저장소의 예제 코드 [SpringMain.kt](../examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt) 주석이 잘 정리해 두었듯이, `SpringApplication.run()`은 대략 다음 순서로 생각하면 됩니다.

1. bootstrap / listener / initializer 준비
2. environment 준비
3. application context 생성
4. bean definition 로드와 context 준비
5. context refresh

이 단계에서 `application.yaml`, profile, 외부 설정, component scan, auto-configuration, bean creation 같은 것들이 본격적으로 엮입니다.
즉 loader가 "실행 가능하게 만드는 부트스트랩"이었다면, `SpringApplication.run()`은 "애플리케이션 의미를 가진 Spring 세계를 실제로 세우는 부트스트랩"입니다.

### 6. context refresh 중에 내장 웹 서버가 뜬다

Spring MVC 기반 웹 애플리케이션이면 `ApplicationContext`가 refresh 되는 동안 내장 Tomcat 같은 웹 서버가 초기화됩니다.
Spring Boot 공식 문서도 `ContextRefreshedEvent`와 `WebServerInitializedEvent`가 `ApplicationStartedEvent`보다 앞에서 발생한다고 설명합니다.

즉 "컨텍스트가 만들어졌다"와 "포트를 바인드한 웹 서버가 떴다"는 같은 말이 아닙니다.
웹 애플리케이션에서는 refresh 과정 안에서 웹 서버 준비가 끼어듭니다.

실제로 저장소 안 샘플 jar를 실행해 보면 이런 순서가 보입니다.

```text
Tomcat initialized with port 0 (http)
Starting Servlet engine: [Apache Tomcat/10.1.30]
Root WebApplicationContext: initialization completed in 539 ms
Tomcat started on port 60324 (http) with context path '/'
Started FullyPortableApplication in 1.123 seconds
```

여기서 볼 포인트는 두 가지입니다.

- `Tomcat started on port ...`
  - 내장 웹 서버가 포트를 바인드했다는 강한 관측 신호
- `Started FullyPortableApplication ...`
  - Spring Boot가 startup을 거의 마무리했다는 높은 수준의 관측 신호

다만 이 둘을 무조건 "이제 모든 startup 작업이 완전히 끝났다"와 동의어로 취급하면 안 됩니다.
Spring Boot는 startup 후반에 runner와 availability state도 별도로 다루기 때문입니다.

## 언제부터 "떠 있다"고 말할 수 있을까

이 질문은 의외로 계층마다 답이 다릅니다.

- 운영 체제 관점
  - `java` 프로세스가 생긴 순간 이미 "실행 중"입니다.
- Java launcher 관점
  - manifest를 읽고 시작점을 잡은 뒤부터 JAR 실행이 진행 중입니다.
- Spring context 관점
  - `ApplicationContext`가 refresh 되면 애플리케이션은 live한 상태로 볼 수 있습니다.
- 트래픽 수신 관점
  - 내장 웹 서버가 포트를 바인드하고, startup runner까지 끝나 readiness가 올라와야 정말 "요청을 받을 준비가 되었다"고 말하는 편이 정확합니다.

Spring Boot 공식 문서는 이 차이를 꽤 분명하게 나눕니다.
문서에 따르면:

- `ApplicationStartedEvent`
  - context refresh 뒤, 하지만 `ApplicationRunner` / `CommandLineRunner` 호출 전
- `ApplicationReadyEvent`
  - runner 호출 뒤
- readiness state
  - `ApplicationReadyEvent` 바로 뒤 `ReadinessState.ACCEPTING_TRAFFIC`

그래서 "프로세스가 떴다", "컨텍스트가 refresh 됐다", "웹 서버가 포트를 바인드했다", "트래픽을 받아도 된다"는 서로 다른 문장입니다.
운영 장애나 startup hook, warm-up, cache priming, runner 로직을 다룰 때 이 차이를 구분하지 않으면 굉장히 쉽게 오해합니다.

## 흔한 오해를 여기서 끊어 두기

### `java -jar`면 내 앱의 `main()`이 곧바로 실행된다

Spring Boot executable jar에서는 그렇지 않습니다.
Java launcher가 먼저 manifest의 `Main-Class`를 보고, Spring Boot loader가 한 번 더 실제 앱 시작 클래스를 찾아 연결합니다.

### Spring Boot fat jar는 "라이브러리를 그냥 한 폴더에 모은 JAR"이다

그렇게 보면 부족합니다.
핵심은 라이브러리가 `BOOT-INF/lib/` 아래 nested jar로 들어 있고, 이를 읽기 위한 loader가 같이 들어 있다는 점입니다.
즉 단순 압축이 아니라 실행 구조까지 포함한 packaging입니다.

### `Started ...` 로그가 보이면 모든 startup 작업이 끝난 것이다

실무에서는 종종 그렇게 읽지만, readiness와 완전히 같은 말로 취급하면 위험합니다.
Spring Boot는 started, ready, liveness, readiness를 구분하고, runner 작업은 ready 직전까지도 이어질 수 있습니다.

### `fork()`가 핵심이다

설명용으로 `fork + exec`를 함께 말하는 것은 유용하지만, 여기서 정말 load-bearing한 개념은 `execve` 쪽입니다.
우리가 궁금한 것은 "새 프로그램 이미지가 어떻게 시작되느냐"이지, 셸 구현 세부를 전부 배우는 것이 아니기 때문입니다.

## 직접 확인해 보기

이 주제는 문서만 읽고 끝내기보다 직접 관측해야 훨씬 빨리 머리에 남습니다.

### 1. `-jar`가 JAR 실행 모드라는 것 확인

```sh
java --help | sed -n '1,12p'
```

PASS:
`java [options] -jar <jarfile>.jar`가 보인다.

FAIL:
`-jar` 항목이 안 보이거나, 현재 사용 중인 `java`가 기대한 JDK/JRE가 아니다.

### 2. Spring Boot jar 내부 구조 확인

```sh
jar tf jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar | sed -n '1,80p'
```

PASS:
`org/springframework/boot/loader/`와 `BOOT-INF/classes/`, `BOOT-INF/lib/`가 같이 보인다.

FAIL:
일반 jar처럼 앱 클래스만 보이고 loader / `BOOT-INF` 구조가 없다.

### 3. manifest에서 loader와 실제 시작 클래스 확인

```sh
unzip -p jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF
```

PASS:
`Main-Class: org.springframework.boot.loader.launch.JarLauncher`와 `Start-Class: ...`가 함께 보인다.

FAIL:
`Main-Class`만 있고 `Start-Class`가 없거나, 기대한 Spring Boot executable jar가 아니다.

### 4. 실제 startup 로그 관측

```sh
java -jar jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar --server.port=0
```

PASS:
`Tomcat started on port ...`와 `Started ...` 로그를 순서대로 볼 수 있다.

FAIL:
startup 중 예외가 나거나, 포트 바인드 이전에 종료된다.

## 정리

`java -jar SpringBootApp.jar`는 "JAR 안의 내 main 하나를 그냥 실행한다"는 단순한 동작이 아닙니다.
운영 체제는 `java`라는 네이티브 프로그램을 시작하고, Java launcher는 manifest를 읽고, Spring Boot loader는 `BOOT-INF/...`를 바탕으로 classpath를 조립하고, 그 뒤에야 실제 애플리케이션 `main()`이 호출됩니다.
그리고 `SpringApplication.run()`이 environment, context, web server, runner, readiness를 차례로 닫아 가며 비로소 "트래픽을 받아도 되는 애플리케이션" 상태에 도달합니다.

이 흐름을 머릿속에 고정해 두면, 다음 질문들도 훨씬 쉬워집니다.
왜 Spring Boot jar 안에 loader 클래스가 들어 있는가, 왜 일반 jar와 layout이 다른가, 왜 startup 로그의 각 줄이 다른 층위의 신호인가, 그리고 언제부터 health check나 readiness probe를 신뢰해도 되는가.

## 근거와 추가 읽기

- source reservoir: `.tmp/interviews2.md`
- repo example: [SpringMain.kt](../examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt)
- related hub: [spring_init.md](./spring_init.md)
- Spring Boot executable jar format: [docs.spring.io](https://docs.spring.io/spring-boot/specification/executable-jar/index.html)
- Spring Boot application lifecycle and availability: [docs.spring.io](https://docs.spring.io/spring-boot/4.1/reference/features/spring-application.html)
- `execve(2)` background: [man7.org](https://man7.org/linux/man-pages/man2/execve.2.html)
