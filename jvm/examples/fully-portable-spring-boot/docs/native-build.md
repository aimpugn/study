# Native Build.

Java 25 + Spring Boot 4.0.0 (Java 25 1급 지원) + GraalVM for JDK 25 Native Image를 사용하고, Linux 쪽은 musl 기반 정적 링크(native-image `--static --libc=musl`)로 “libc조차 포함된 단일 바이너리”를 만들고, macOS/Windows는 OS 기본 라이브러리에만 의존하는 “셀프 컨테이너 JVM 없는 바이너리”로 가는 식입니다.

그 위에 picocli로 `fpsb boot`, `fpsb config`, `fpsb stat` 같은 서브커맨드를 구현하고, `boot` 서브커맨드가 실제 Spring Boot 애플리케이션(HTTP 서버 + TCP 전문 서버)을 띄우는 구조로 잡으면 됩니다. TCP 전문 서버는 Netty 없이도, Spring이 관리하는 하나의 `@Service` 안에서 `ServerSocket` 루프를 돌리는 방식으로 쉽게 PoC가 가능합니다.

아래에서 단계별로 구조와 코드 예시를 보여 드리겠습니다.

---

### 1. 현실적인 “fully portable” 정의를 먼저 정리

완전히 OS 무관한 단일 바이너리 하나로 Linux, macOS, Windows, x86, aarch64를 모두 커버하는 것은 불가능합니다. GraalVM Native Image로 빌드된 실행 파일은 “빌드된 OS/아키텍처에서만” 동작합니다.([위키백과])

그래서 현실적인 목표는 이렇게 잡는 것이 좋습니다.

1. OS/아키텍처별로 한 개씩의 바이너리를 만든다.
   예를 들어 `fpsb-linux-x64`, `fpsb-linux-aarch64`, `fpsb-macos-aarch64`, `fpsb-win-x64.exe` 같은 식입니다.

2. Linux에서는 musl 정적 링크로 glibc 버전 의존성까지 제거한다.
   GraalVM Native Image는 `--static --libc=musl`로 libc까지 포함한 완전 정적 바이너리를 만들 수 있습니다.([graalvm.org]) 이러면 런타임에 필요한 것은 사실상 커널뿐이고, `scratch` 같은 빈 컨테이너에서도 실행이 가능합니다.

3. macOS/Windows에서는 “JDK 없이 실행되는 단일 실행 파일”을 목표로 한다.
   이 쪽은 OS 정책상 완전 정적 링크는 어렵지만, 어쨌든 별도 JRE/JDK 설치 없이 바이너리만 두고 실행할 수 있으니, 사용자 입장에서는 충분히 “fully portable”에 가깝습니다.

질문에서 말씀하신 “OS 버전 무관하고 libc 동적 링크에 영향 받지 않는 fat 바이너리”는 Linux + musl 조합에서 잘 만족시킬 수 있고, 나머지 OS는 “최소 OS 의존” 정도로 보는 게 좋습니다.

---

### 2. PoC용 전체 구조 개략

PoC는 최대한 단순하게, 단일 Gradle 프로젝트 하나로 잡는 것이 좋습니다.

논리적인 구성은 다음과 같습니다.

* 하나의 Java 애플리케이션이 있다. 이 애플리케이션의 진입점은 picocli CLI이다.
  사용자가 `fpsb ...` 명령을 치면 picocli가 인자를 파싱하고, `boot`, `config`, `stat` 같은 서브커맨드를 실행한다.

* `fpsb boot` 서브커맨드는 Spring Boot 애플리케이션을 구동한다.
  내부적으로 `SpringApplication.run(FpsbServerApplication.class, ...)`를 호출해서 HTTP 서버와 TCP 전문 서버를 같이 띄운다.

* `fpsb config`, `fpsb stat` 서브커맨드는 별도의 서버를 띄우지 않고, 설정 파일(YAML 등)을 읽어 출력하거나, 간단한 상태 파일이나 헬스 엔드포인트를 조회하는 정도로만 PoC를 구성한다.

* Spring Boot 애플리케이션은 두 가지 역할을 동시에 한다.
  하나는 익숙한 HTTP 서버(`/health`, `/echo` 같은 간단한 엔드포인트).
  다른 하나는 특정 포트(예: 5000)에서 고정 길이 전문(예: 300 byte)를 받는 TCP 서버이다.

* TCP 서버는 Netty 없이 구현한다.
  Spring이 관리하는 `@Service` 안에서 `ServerSocket` 또는 `ServerSocketChannel`을 사용하여 accept 루프를 돌리고, 고정 길이만큼 바이트를 읽어서 공통 도메인 서비스에 위임하는 구조로 한다. 이 정도는 순수 JDK + Spring만으로 쉽게 구현 가능합니다.

* 빌드는 GraalVM Native Image Gradle 플러그인을 사용해 `./gradlew nativeCompile`로 바이너리를 만들고, 그 결과물을 `bin/fpsb` 또는 `fpsb`라는 이름으로 패키징한다.([graalvm.org])

* 프로젝트 디렉터리 안에 `bin/jdk-25`를 두고, 빌드 시에는 이 JDK를 사용하도록 스크립트를 작성할 수 있다. 런타임에는 필요 없고, “빌드 환경을 외부에 의존하지 않는다”라는 의미의 포터블 빌드 환경을 만들 수 있다.

---

### 3. Gradle 설정 예시 (Java 25 + Spring Boot 4 + GraalVM Native + picocli)

Java 25와 Spring Boot 4.0.0, 그리고 GraalVM Native Build Tools 0.11.3을 사용하는 Gradle Kotlin DSL 예시입니다.([Home])

```kotlin
// build.gradle.kts

import org.graalvm.buildtools.gradle.dsl.GraalVMExtension

plugins {
    // Spring Boot 4.0.0 (Java 25 1급 지원)
    id("org.springframework.boot") version "4.0.0"
    id("io.spring.dependency-management") version "1.1.6"

    // 일반 Java + 애플리케이션
    java
    application

    // GraalVM Native Image Gradle 플러그인
    id("org.graalvm.buildtools.native") version "0.11.3"
}

// main() 이 들어있는 클래스의 FQCN
application {
    mainClass.set("com.example.fpsb.FpsbCliApplication")
}

java {
    // Java 25 toolchain 사용
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(25))
    }
}

repositories {
    mavenCentral()
}

dependencies {
    // Spring Boot 웹 서버 (Tomcat 기반, HTTP 엔드포인트용)
    implementation("org.springframework.boot:spring-boot-starter-web")

    // 필요하다면 actuator로 상태 노출 (PoC에서는 선택)
    // implementation("org.springframework.boot:spring-boot-starter-actuator")

    // picocli CLI
    implementation("info.picocli:picocli:4.7.6")

    // GraalVM native-image에서 reflection 메타데이터 생성을 위한 annotation processor
    annotationProcessor("info.picocli:picocli-codegen:4.7.6")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
}

// GraalVM Native 빌드 설정
configure<GraalVMExtension> {
    binaries {
        named("main") {
            // 생성될 실행 파일 이름
            imageName.set("fpsb")

            // Spring Boot 4 + Native Image는 기본적으로 AOT 리소스 탐지 활성화
            resources.autodetect()

            // Linux에서 완전 정적 musl 바이너리로 만들고 싶다면 빌드 아규먼트를 추가
            // (musl 지원 GraalVM 25 이미지 또는 네이티브 설치 필요)
            buildArgs.add("--static")
            buildArgs.add("--libc=musl")
        }
    }
}

// GraalVM 25 툴체인을 명시적으로 선택하고 싶다면 (예: Oracle GraalVM 25)
// docs에서 안내하는 대로 vendor, version을 지정할 수 있음:contentReference[oaicite:5]{index=5}
tasks.withType<org.graalvm.buildtools.gradle.tasks.BuildNativeImageTask> {
    javaLauncher.set(
        javaToolchains.launcherFor {
            languageVersion.set(JavaLanguageVersion.of(25))
            // vendor.set(JvmVendorSpec.matching("Oracle Corporation"))
        }
    )
}
```

이 상태에서 개발 머신에 GraalVM for JDK 25가 설치되어 있다면, 다음 명령 한 번으로 네이티브 바이너리가 생성됩니다.

```bash
./gradlew nativeCompile
# 결과 바이너리: build/native/nativeCompile/fpsb
```

Spring Boot 공식 문서도 Gradle 플러그인을 통해 `nativeCompile`로 네이티브 이미지를 만드는 방식을 권장하고 있습니다.([graalvm.org])

---

### 4. Picocli CLI + Spring Boot 서버를 한 프로세스 안에 넣는 코드 스켈레톤

PoC 용으로 최대한 단순하게, 진입점은 picocli의 `FpsbCliApplication.main` 하나로 하고, `boot`, `config`, `stat` 서브커맨드를 구현하는 예시입니다.

설명에 집중하기 위해 Java 코드만 보여줍니다.

```java
// src/main/java/com/example/fpsb/FpsbCliApplication.java
package com.example.fpsb;

import org.springframework.boot.SpringApplication;
import org.springframework.context.ConfigurableApplicationContext;
import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

// 루트 커맨드: `fpsb` 자체를 의미
@Command(
        name = "fpsb",
        mixinStandardHelpOptions = true,
        version = "fpsb 0.1.0",
        description = "Fully Portable Spring Boot PoC",
        subcommands = {
                FpsbCliApplication.BootCommand.class,
                FpsbCliApplication.ConfigCommand.class,
                FpsbCliApplication.StatCommand.class
        }
)
public class FpsbCliApplication implements Runnable {

    // 루트 커맨드가 직접 할 일은 없고, help/usage 정도만 제공
    @Override
    public void run() {
        System.out.println("Use `fpsb boot`, `fpsb config`, or `fpsb stat`.");
    }

    public static void main(String[] args) {
        int exitCode = new CommandLine(new FpsbCliApplication()).execute(args);
        System.exit(exitCode);
    }

    // `fpsb boot` 서브커맨드: Spring Boot 서버를 구동
    @Command(
            name = "boot",
            description = "Start Spring Boot HTTP + TCP server"
    )
    static class BootCommand implements Runnable {

        @Option(names = "--http-port", description = "HTTP server port", defaultValue = "8080")
        int httpPort;

        @Option(names = "--tcp-port", description = "Fixed-length TCP server port", defaultValue = "5000")
        int tcpPort;

        @Option(names = "--profile", description = "Spring profile", defaultValue = "default")
        String profile;

        @Override
        public void run() {
            // Spring Boot 애플리케이션을 프로그래밍 방식으로 구동
            SpringApplication app = new SpringApplication(FpsbServerApplication.class);

            // 환경 설정을 커맨드 라인에서 주입
            app.setAdditionalProfiles(profile);

            // server.port와 커스텀 프로퍼티로 TCP 포트를 전달
            String[] springArgs = new String[]{
                    "--server.port=" + httpPort,
                    "--fpsb.tcp.port=" + tcpPort
            };

            ConfigurableApplicationContext ctx = app.run(springArgs);

            // 여기서는 단순히 서버가 종료될 때까지 대기
            // 프로덕션에서는 shutdown 훅, signal 처리 등을 더 정교하게 구성
            ctx.registerShutdownHook();
        }
    }

    // `fpsb config` 서브커맨드: 설정 조회/수정 PoC
    @Command(
            name = "config",
            description = "Show or edit configuration (PoC: just print)"
    )
    static class ConfigCommand implements Runnable {

        @Option(names = "--show", description = "Show current config", defaultValue = "true")
        boolean show;

        @Override
        public void run() {
            // PoC에서는 그냥 하드코딩된 설정을 출력
            // 실제로는 conf/application.yml 파일을 읽어서 출력하거나,
            // JSON/YAML을 편집하는 식으로 구현 가능
            if (show) {
                System.out.println("fpsb config (PoC)");
                System.out.println("  http.port = 8080");
                System.out.println("  tcp.port  = 5000");
            }
        }
    }

    // `fpsb stat` 서브커맨드: 상태 조회 PoC
    @Command(
            name = "stat",
            description = "Show server status (PoC: static output)"
    )
    static class StatCommand implements Runnable {

        @Override
        public void run() {
            // 실제로는 HTTP /actuator/health 를 호출하거나,
            // PID 파일 등을 읽어 현재 동작 중인 인스턴스 상태를 확인할 수 있음
            System.out.println("fpsb stat (PoC)");
            System.out.println("  http: UNKNOWN (PoC)");
            System.out.println("  tcp : UNKNOWN (PoC)");
        }
    }
}
```

여기서는 picocli와 Spring Boot를 강하게 엮지 않고, `boot` 서브커맨드가 단순히 `SpringApplication.run`을 호출하는 형태로 유지했습니다. 이 방식은 네이티브 이미지에서도 비교적 문제가 적고, CLI와 서버 생명주기를 명확히 분리하기 때문에 PoC로는 이해하기 쉽습니다.

`picocli-spring-boot-starter`를 사용하면 Spring DI 안에서 picocli 커맨드를 관리할 수도 있지만, 필요 이상으로 복잡해지고 네이티브 이미지 구성도 까다로워질 수 있습니다.([Maven Repository])

---

### 5. Spring Boot 서버 (HTTP + 고정 길이 TCP 전문 서버) 구현 예시

Spring Boot 애플리케이션 본체와 고정 길이 전문 서버를 구현한 예시입니다.

먼저 일반적인 Spring Boot `@SpringBootApplication`입니다.

```java
// src/main/java/com/example/fpsb/FpsbServerApplication.java
package com.example.fpsb;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class FpsbServerApplication {

    public static void main(String[] args) {
        SpringApplication.run(FpsbServerApplication.class, args);
    }
}
```

간단한 HTTP 컨트롤러를 하나 둡니다.

```java
// src/main/java/com/example/fpsb/http/HelloController.java
package com.example.fpsb.http;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HelloController {

    @GetMapping("/health")
    public String health() {
        return "OK";
    }

    @GetMapping("/echo")
    public String echo(String msg) {
        return "echo: " + msg;
    }
}
```

이제 고정 길이 TCP 전문 서버입니다. PoC에서는 “연결마다 300바이트를 딱 한 번 읽고 처리”하는 아주 단순한 형태로 두겠습니다.

```java
// src/main/java/com/example/fpsb/tcp/FixedLengthTcpServer.java
package com.example.fpsb.tcp;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.InputStream;
import java.io.OutputStream;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

// Spring이 관리하는 TCP 서버 컴포넌트
@Service
public class FixedLengthTcpServer {

    // application.properties 또는 커맨드라인에서 주입하는 TCP 포트
    @Value("${fpsb.tcp.port:5000}")
    private int tcpPort;

    // PoC에서는 고정 길이 300 바이트로 가정
    private static final int MESSAGE_LENGTH = 300;

    private final ExecutorService clientExecutor = Executors.newCachedThreadPool();

    private volatile boolean running = true;

    private ServerSocket serverSocket;

    private final FixedLengthMessageHandler handler;

    // 메시지 처리 로직을 분리해 놓으면 테스트와 재사용이 쉬움
    public FixedLengthTcpServer(FixedLengthMessageHandler handler) {
        this.handler = handler;
    }

    // Spring 컨텍스트 초기화 후 TCP 서버를 시작
    @PostConstruct
    public void start() {
        clientExecutor.submit(this::acceptLoop);
    }

    // 애플리케이션 종료 시 TCP 서버도 정리
    @PreDestroy
    public void stop() {
        running = false;
        try {
            if (serverSocket != null) {
                serverSocket.close();
            }
        } catch (Exception ignored) {
        }
        clientExecutor.shutdownNow();
    }

    private void acceptLoop() {
        try (ServerSocket ss = new ServerSocket(tcpPort)) {
            this.serverSocket = ss;
            System.out.println("FixedLengthTcpServer listening on port " + tcpPort);

            while (running) {
                Socket client = ss.accept();
                clientExecutor.submit(() -> handleClient(client));
            }
        } catch (Exception e) {
            if (running) {
                e.printStackTrace();
            }
        }
    }

    private void handleClient(Socket client) {
        try (Socket c = client;
             InputStream in = c.getInputStream();
             OutputStream out = c.getOutputStream()) {

            byte[] buf = in.readNBytes(MESSAGE_LENGTH);
            if (buf.length != MESSAGE_LENGTH) {
                // 메시지 길이가 맞지 않으면 PoC에서는 그냥 무시
                return;
            }

            byte[] response = handler.handle(buf);

            out.write(response);
            out.flush();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
```

그리고 실제 비즈니스 로직을 담당하는 핸들러입니다. 여기서는 단순히 전문의 앞 10바이트만 잘라서 로그 찍고, “OK”라는 문자열을 돌려주는 수준으로 둡니다.

```java
// src/main/java/com/example/fpsb/tcp/FixedLengthMessageHandler.java
package com.example.fpsb.tcp;

import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

@Service
public class FixedLengthMessageHandler {

    public byte[] handle(byte[] message) {
        // PoC: 앞 10바이트만 로그용으로 출력
        byte[] header = Arrays.copyOfRange(message, 0, 10);
        String headerStr = new String(header, StandardCharsets.US_ASCII);
        System.out.println("Received fixed-length message, header = [" + headerStr + "]");

        // 실제로는 이 바이트 배열을 파싱해 도메인 객체로 변환하고
        // 공통 서비스 로직을 호출한 뒤, 응답 전문을 다시 직렬화하는 구조가 될 것
        String response = "OK";
        return response.getBytes(StandardCharsets.US_ASCII);
    }
}
```

이 정도 구성만으로도 다음이 가능합니다.

* `fpsb boot --http-port=8080 --tcp-port=5000`
  → HTTP `/health`, `/echo` 엔드포인트가 뜬다.
  → 5000 포트에서 300바이트 전문을 받는 TCP 서버가 뜬다.

* 논리적으로 HTTP 요청과 TCP 전문 모두 같은 `FixedLengthMessageHandler` 또는 더 일반화된 도메인 서비스 계층을 공유하도록 설계할 수 있다. 예를 들어 HTTP `/process`는 JSON 바디를 받아 도메인 객체로 만든 뒤, TCP 전문이 사용하는 것과 같은 서비스 메서드를 호출하도록 만들 수 있다.

Netty나 Spring Integration 없이도 이 정도 raw TCP 서버는 Spring만으로 충분히 구현이 가능합니다.

Spring Integration TCP/UDP 모듈(`spring-integration-ip`)을 쓰면 length-header 프로토콜, connection factory, 메세지 채널 등 많은 기능을 얻을 수 있지만, PoC에서는 오히려 학습 비용이 더 크기 때문에 이렇게 “Spring이 관리하는 순수 JDK 소켓 서버”로 시작하는 편이 직관적입니다.

---

### 6. GraalVM Native Image 빌드와 정적 링크 전략

빌드 측면에서는 크게 두 단계로 생각하면 됩니다.

첫째, 개발/빌드 환경에서 GraalVM for JDK 25를 사용해 네이티브 이미지를 만든다.
Spring Boot 공식 문서는 Native Build Tools 플러그인을 이용해 Gradle에서 `nativeCompile`을 실행하는 방식을 설명하고 있습니다.([graalvm.org])

둘째, Linux에서 glibc 의존성을 없애고 싶다면 musl 기반 빌드를 사용한다.
GraalVM Native Image 문서에는 완전 동적, 대부분 정적(단 libc만 동적), 완전 정적( musl-libc 사용) 세 가지 링크 방식이 설명돼 있습니다. 정적 바이너리를 만들려면 musl을 포함한 툴체인 또는 `native-image-community:25.0.1-muslib-ol9-20251021` 같은 컨테이너 이미지를 사용하는 것이 일반적입니다.([graalvm.org])

PoC에서는 일단 다음 순서로 진행하는 것이 안전합니다.

1. 로컬에서 GraalVM 25 + Spring Boot 4 조합으로 `./gradlew nativeCompile`이 성공하도록 만든다.
2. 결과 바이너리를 직접 실행하면서 HTTP/TCP 서버가 제대로 뜨는지 확인한다.
3. 그 이후에 Docker 기반 musl 빌드(완전 정적 바이너리)를 적용해 본다.

---

### 7. `bin/jdk-25`와 `fpsb` 런처 스크립트 구조

질문에서 “bin 디렉토리 이하에 jdk 25가 있어서 로컬 머신 환경에 무관하게 빌드하고 실행”하는 목표를 말씀하셨습니다. 이 요구는 두 레벨로 나뉩니다.

* 런타임: 네이티브 바이너리만 필요하도록 한다.
  즉, 운영 서버에는 `fpsb` 실행 파일 하나만 두고 실행할 수 있도록 한다. (Linux라면 musl 정적 링크를 사용해 libc 의존성까지 제거)

* 빌드 타임: 프로젝트 안에 포함된 JDK 25로 빌드하게 한다.
  예를 들어 `bin/jdk-25`에 JDK를 압축 풀어 두고, `bin/fpsb-build` 스크립트가 이 JDK를 사용해서 Gradle과 native-image를 실행하게 할 수 있습니다.

아주 단순한 런처 스크립트 예시는 다음과 같습니다.

```bash
#!/usr/bin/env bash
# bin/fpsb (Linux/macOS용 예시)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_HOME="${SCRIPT_DIR}/.."

# 네이티브 바이너리가 여기 있다고 가정
NATIVE_BIN="${APP_HOME}/build/native/nativeCompile/fpsb"

exec "${NATIVE_BIN}" "$@"
```

배포용 아티팩트를 만들 때는 `build/native/nativeCompile/fpsb`를 `dist/fpsb` 같은 위치로 복사하고, 위 스크립트를 조금 수정해 “빌드 경로를 제거한 순수 배포 디렉터리”로 만드는 식으로 정리하면 됩니다.

빌드용 스크립트는 다음과 같이 `bin/jdk-25`를 사용하도록 할 수 있습니다.

```bash
#!/usr/bin/env bash
# bin/fpsb-build (PoC 예시)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_HOME="${SCRIPT_DIR}/.."

JDK_HOME="${APP_HOME}/bin/jdk-25"
export JAVA_HOME="${JDK_HOME}"
export PATH="${JAVA_HOME}/bin:${PATH}"

cd "${APP_HOME}" || exit 1

./gradlew clean nativeCompile
```

이렇게 하면 개발자 로컬에 무엇이 설치되어 있든 상관 없이, 프로젝트 안에 포함된 JDK 25를 사용해 항상 동일한 버전으로 빌드할 수 있습니다. 이후 CI에서도 동일한 구조를 사용하면 “빌드 환경의 JDK 버전이 뒤섞이는 문제”를 줄일 수 있습니다.

---

### 8. 정리

정리하면, 질문하신 PoC 목표는 다음 방식으로 충분히 달성할 수 있습니다.

* Java 25 + Spring Boot 4.0.0 + GraalVM 25 Native Image 조합으로, OS/아키텍처별 단일 실행 파일을 생성한다.
* Linux에서는 musl 기반 완전 정적 링크로 libc 의존성까지 제거된 바이너리를 만들 수 있다.
* picocli를 사용해 `fpsb boot`, `fpsb config`, `fpsb stat` 같은 CLI 인터페이스를 제공하고, `boot` 서브커맨드가 Spring Boot 애플리케이션(HTTP 서버 + 고정 길이 TCP 전문 서버)을 구동하는 구조로 만든다.
* TCP 전문 서버는 Netty나 Spring Integration 없이도, Spring이 관리하는 `@Service` 안에서 순수 `ServerSocket` 루프를 돌리는 방식으로 충분히 PoC 구현이 가능하다.
* 프로젝트 내부에 `bin/jdk-25`를 두고 빌드 스크립트가 이 JDK와 GraalVM native-image를 사용하도록 하면, 빌드 환경 역시 “portable”하게 만들 수 있다.

이 구조를 그대로 코드로 옮겨서 한 번 nativeCompile까지 돌려보시면, 어느 지점에서 네이티브 이미지 이슈가 발생하는지까지 확인할 수 있고, 그 다음 단계부터는 GraalVM의 리플렉션/리소스 설정을 조금씩 조정해 가며 완성도를 올려 가면 됩니다.


---

• 좋은 질문입니다. 핵심은 “GraalVM JDK”랑 “GraalVM native-image로 만든 바이너리”를 구분해서 생각해야 한다는 점입니다.

1. GraalVM JDK vs native-image

- GraalVM JDK를 그냥 java로 쓰면 사실상 HotSpot이랑 거의 동일하게 동작합니다.
  → 이 모드에서는 리플렉션도 평소처럼 다 됩니다.
- 지금 문제를 일으키는 건 ./target/fpsb 같은 native-image로 미리 컴파일된 바이너리입니다.
  → 이건 더 이상 JVM 위에서 .class를 읽어오는 런타임이 아니라, AOT(미리 컴파일된) 네이티브 실행 파일입니다.

2. 왜 native-image에서 리플렉션이 “안 보이는” 것처럼 보이나

native-image는 “닫힌 세계(closed‑world)” 가정으로 빌드합니다:

- 빌드 시점에 “어떤 클래스/메서드/필드가 실제로 쓰이는지” 전부 분석해서, 그 정보만 바이너리에 집어넣고 나머지는 과감히 버립니다.
  → 안 그러면 dead code 제거/최적화/스타트업 이점이 거의 사라지기 때문입니다.
- 그런데 리플렉션은 코드상에 new StatCommand() 같은 정적 참조가 없어도,
  런타임에 Class.forName("...StatCommand").getDeclaredConstructor()로 **“마음대로 호출할 수 있다”**고 가정하는 API입니다.
- native-image 입장에서는 “정적 분석으로는 이 생성자가 쓰이는지 알 수가 없으니,
  개발자가 별도로 ‘이 클래스를 리플렉션으로 쓸 거다’라고 알려줘야 합니다.”
    - 그게 Graal의 reflect-config.json 또는 Spring의 RuntimeHints, 혹은 Picocli의 @NativeImageHint 같은 것들이고,
    - 안 알려주면, 그 생성자는 메타데이터에서 빠지거나 접근 불가 처리되고,
      결국 Class.getDeclaredConstructor()가 **NoSuchMethodException**을 던지면서
      “the class has no constructor”라는 메시지로 보입니다.

즉, Java 소스/바이트코드에는 분명 기본 생성자가 있는데,
native-image가 만든 내부 리플렉션 메타데이터에는 그 정보가 없어서 “없는 것처럼” 보이는 겁니다.

3. 그래서 지금 프로젝트에서는

- ./target/fpsb는 이전에 빌드된 native 바이너리라서, 그 안에는
    - Picocli의 기본 DefaultFactory가 있고
    - Graal에 StatCommand 생성자를 리플렉션으로 쓴다는 힌트가 전혀 없어서
      → DefaultFactory가 newInstance()를 하려다 실패 → 지금 그 에러가 납니다.
- 해결 방법은 두 가지 계열입니다:
    - (A) 리플렉션을 안 쓰도록: 우리가 바꾼 것처럼 ApplicationMain에서 CommandLine(IFactory)를 써서 new StatCommand()를 직접 해
      준다.
    - (B) 리플렉션 정보를 native-image에 알려주도록:
        - Graal reflect-config.json에 StatCommand의 생성자 사용을 등록하거나,
        - Spring RuntimeHintsRegistrar를 구현해서 reflection().registerType(StatCommand.class,
          MemberCategory.INVOKE_DECLARED_CONSTRUCTORS) 같은 힌트를 넣는다,
        - 혹은 Picocli의 @NativeImageHint 등을 사용한다.

그리고 이 모든 건 “native-image로 컴파일할 때만” 필요한 제약입니다.
그냥 GraalVM JDK를 java로 쓰거나, OpenJDK/HotSpot을 쓰는 JVM 모드에서는 지금 코드 그대로 잘 동작합니다.

요약하면:

- “왜 생성자를 못 보냐?” → native-image가 리플렉션 메타데이터를 기본으로 다 안 넣기 때문입니다.
- “그냥 OpenJDK처럼 쓰면 안 되나?” → JDK로 쓸 때는 그렇고,
  지금처럼 “native 바이너리로 빌드해서 JVM 없이 실행하고 싶다”면 그 대가로 이런 제약(AOT + 힌트/설정)이 따라옵니다.

여기까지 이해되셨으면, 원하시는 방향(리플렉션 힌트 vs IFactory 방식) 기준으로 깔끔한 최종 구조도 같이 정리해 드릴게요.

---

• 요약부터 적을게요.

1. “JDK 25를 쓴다”와 “classfile 25로 컴파일한다”는 다른 개념입니다.
    - 지금 pom.xml은 java.version=21이라 바이트코드는 Java 21(major 65) 로 컴파일됩니다.
    - 하지만 실행하는 JDK는 GraalVM JDK 25를 쓰도록 맞춰둘 수 있고, 실제로 그게 맞는 방식입니다.
    - Spring Boot 3.3.4 가 사용하는 ASM 이 아직 classfile 69(Java 25) 를 못 읽어서, 25로 컴파일하면 아까 보신 Unsupported class file
      major version 69 에러가 납니다.
    - 그래서 “JDK 25로 실행하지만, classfile은 21로 컴파일” 구조가 현재 Spring Boot 3.3.x 에서 권장되는 패턴입니다.
2. GraalVM JDK 25 사용 설정
    - bin/build-native 는 이미 GRAAL_HOME="./bin/graalvm/macos" 를 잡아서 GraalVM JDK 25를 씁니다.
    - 방금 bin/fpsb도 이렇게 바꿔두었습니다 (bin/fpsb):

      # 1순위: bin/graalvm/macos
      if [ -d "${SCRIPT_DIR}/graalvm/macos" ]; then
      export JAVA_HOME="${SCRIPT_DIR}/graalvm/macos"
      JAVA_BIN="${JAVA_HOME}/bin/java"
      # 2순위: 예전 레이아웃 bin/jdk-25
      elif [ -d "${SCRIPT_DIR}/jdk-25" ]; then
      export JAVA_HOME="${SCRIPT_DIR}/jdk-25"
      JAVA_BIN="${JAVA_HOME}/bin/java"
      ...
    - 즉, 프로젝트 루트에서:

      ./bin/build-native    # GraalVM JDK 25 + native-image 빌드
      ./target/fpsb boot ... # 빌드 후 네이티브 바이너리 실행

      를 하면 전부 GraalVM JDK 25 기반으로 돌아갑니다.
    - mvn 자체도 GraalVM JDK 25로 돌리려면, 쉘에서 한 번만:

      export JAVA_HOME="$PWD/bin/graalvm/macos"
      export PATH="$JAVA_HOME/bin:$PATH"

      해두고 mvn clean package 등을 실행하시면 됩니다.
3. virtual thread 사용 설정

   지금 코드에는 두 군데에 virtual thread를 연결해 두었습니다.
    - HTTP(Spring MVC/Tomcat) 쪽

      src/main/resources/application.yml 에 다음을 추가했습니다:

      spring:
      threads:
      virtual:
      enabled: true

      이건 Spring Boot 3.2+ 에서 공식으로 제공하는 설정이라,
      JDK 21+ 환경이면 HTTP 요청 처리용 쓰레드 풀을 virtual thread 기반으로 붙여 줍니다.
      (실행 JDK는 25지만, API는 21에서 이미 정식이라 문제 없습니다.)
    - 바이너리 TCP 서버 쪽

      src/main/java/com/example/portable/binary/BinaryProtocolServer.java 의 실행 풀을:

      private final ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();

      로 변경했습니다.
      이제 각 TCP 연결 처리가 플랫폼 쓰레드가 아니라 virtual thread 위에서 돌아갑니다.
4. 왜 classformat 에러를 피하려면 여전히 21로 컴파일해야 하나?
    - Spring Boot 3.3.4 내부의 org.springframework.asm.ClassReader 가 classfile major 69(Java 25)를 아직 지원하지 않습니다.
    - 우리가 java.version을 25로 돌리면, ApplicationMain$1 같은 내부 클래스가 major 69로 컴파일되고,
      Spring AOT(spring-boot:process-aot)가 클래스패스 스캐닝을 하다가 그 클래스를 읽지 못해 지금 보신 에러를 냅니다.
    - 그래서 실제 실행 JDK는 25, 컴파일 타깃은 21로 유지하는 게 현재로서는 가장 현실적인 조합입니다.
      이렇게 해도:
        - GraalVM JDK 25의 런타임, GC, Loom(virtual threads) 전부 사용 가능하고
        - Spring Boot 3.3.x 의 AOT + native-image 빌드도 깨지지 않습니다.

정리하면:

- 지금 설정으로 실행 JDK는 bin/graalvm/macos 안의 GraalVM JDK 25를 사용하게 되어 있고,
- 애플리케이션은 HTTP/TCP 둘 다 virtual threads를 사용하도록 수동 설정해 두었습니다.
- 다만 Spring Boot 3.3.4 한계 때문에 classfile은 Java 21 수준으로 컴파일해야 해서 pom.xml 의 <java.version>21</java.version>은 그대
  로 두는 게 안전합니다.

이 상태에서 한 번:

export JAVA_HOME="$PWD/bin/graalvm/macos"
export PATH="$JAVA_HOME/bin:$PATH"

mvn -Pnative -DskipTests package   # 또는 ./bin/build-native
./target/fpsb boot --port 8080 --binary-port 9090
./target/fpsb stat

---

• 짧게 말하면: “Spring Boot 4.0로 올리면 해결된다”는 보장은 현재(제가 아는 시점 기준) 없습니다.
그리고 지금 겪는 문제는 “버전 숫자”보다는 JDK 25 classfile 지원 여부(ASM 버전) 문제라서 그렇습니다.

핵심만 정리하면:

- Spring Boot 3.3.x는 내부적으로 Spring ASM을 써서 클래스를 스캔(AOT, 컴포넌트 스캔 등)합니다.
- 이 ASM 버전이 아직 classfile major 69(Java 25) 를 못 읽어서,
  JDK 25 타깃으로 컴파일된 ApplicationMain$1.class를 보면 Unsupported class file major version 69 예외를 던집니다.
- 이건 “Boot 3 vs 4”의 문제가 아니라, **“Spring Framework(+ASM)가 Java 25 바이트코드를 지원하느냐”**의 문제입니다.
- 제가 학습한 시점(2024-10) 기준으로는 Spring Boot 4.0 정식은 아직 없거나, Java 25 지원 얘기가 공식화되지 않았습니다.
  그러니 “4.0 쓰면 된다”고 가정하고 올라가는 건 지금으로선 불확실합니다.

반대로, 지금 구조는 이미 JDK 25 + virtual thread를 쓸 수 있습니다:

- JAVA_HOME을 bin/graalvm/macos로 두고 빌드/실행하면 GraalVM JDK 25 런타임을 쓰고 있고,
- pom.xml의 <java.version>21</java.version> 덕분에 classfile은 Java 21 수준이라 Spring AOT/ASM이 안전하게 읽을 수 있고,
- virtual threads는 JDK 21에서 이미 정식이기 때문에:
    - spring.threads.virtual.enabled=true (HTTP 요청용)
    - Executors.newVirtualThreadPerTaskExecutor() (바이너리 서버용)
      로 가상 스레드 기능은 그대로 사용 가능합니다.

그래서 현실적인 선택지는:

- 지금처럼 **“JDK 25로 실행 + classfile 21로 컴파일”**을 유지하고,
- 나중에 Spring 쪽에서 공식적으로 Java 25 classfile을 지원하는 버전(3.x든 4.x든)이 나오면,
  그때 <java.version>25</java.version> 로 올리는 게 안전합니다.

---

• 지금 상황을 한 줄로 요약하면:

- 네이티브 바이너리 안에서 Picocli 옵션 메타데이터가 제대로 안 들어가 있어서 boot/start 서브커맨드가 --port, --binary-port 를 “없는
  옵션”으로 보고 있습니다.
- config 커맨드가 보여주는 예시는 그냥 System.out.println(...) 이라, 실제로 옵션이 잘 동작한다는 뜻이 아닙니다.

이건 코드 버그라기보다 GraalVM native-image + Picocli 세팅 문제라서, Picocli의 GraalVM용 annotation processor (picocli-codegen)를 켜
서 리플렉션 메타데이터를 생성해 주면 해결됩니다.

———

1. pom.xml에 Picocli codegen 추가

info.picocli:picocli-codegen 의존성과 maven-compiler-plugin 설정을 넣어 주세요.

pom.xml 안에 이미 있는 Picocli 의존 바로 아래 정도에:

  <dependency>
      <groupId>info.picocli</groupId>
      <artifactId>picocli</artifactId>
      <version>${picocli.version}</version>
  </dependency>

  <!-- GraalVM native-image용 Picocli 메타데이터 생성기 -->
  <dependency>
      <groupId>info.picocli</groupId>
      <artifactId>picocli-codegen</artifactId>
      <version>${picocli.version}</version>
      <scope>provided</scope>
  </dependency>

그리고 <build><plugins> 안에 maven-compiler-plugin 설정을 추가합니다 (없으면 새로 추가, 있으면 configuration에 합쳐도 됩니다):

  <build>
      <plugins>
          <!-- Picocli annotation processing: META-INF/native-image/*/reflect-config.json 생성 -->
          <plugin>
              <groupId>org.apache.maven.plugins</groupId>
              <artifactId>maven-compiler-plugin</artifactId>
              <configuration>
                  <annotationProcessorPaths>
                      <path>
                          <groupId>info.picocli</groupId>
                          <artifactId>picocli-codegen</artifactId>
                          <version>${picocli.version}</version>
                      </path>
                  </annotationProcessorPaths>
                  <compilerArgs>
                      <!-- picocli-codegen이 native-image 메타데이터를 이 프로젝트 이름으로 생성하게 함 -->
                      <arg>-Aproject=${project.groupId}/${project.artifactId}</arg>
                  </compilerArgs>
              </configuration>
          </plugin>

          <!-- (이미 있던 spring-boot-maven-plugin, native-maven-plugin 등은 그대로 두면 됩니다) -->
      </plugins>
  </build>

이렇게 하면 컴파일 시 Picocli가 META-INF/native-image/com.example/fully-portable-spring-boot/reflect-config.json 같은 파일을 만들어
주고, GraalVM Maven 플러그인이 이걸 자동으로 native-image에 반영합니다.
그 결과, StartCommand 의 @Option(-p, --port, -b, --binary-port) 같은 정보가 네이티브 실행 파일 안에도 제대로 살아있게 됩니다.

———

2. 네이티브 바이너리 다시 빌드

이제 각자 환경(당신의 macOS)에서, GraalVM JDK 25를 쓰도록 설정한 다음 다시 빌드해야 합니다.

예를 들어:

# (선택) GraalVM JDK 25 지정
export JAVA_HOME="$PWD/bin/graalvm/macos"
export PATH="$JAVA_HOME/bin:$PATH"

# JAR + AOT 다시 생성
mvn clean package

# macOS용 네이티브 바이너리 생성
mvn -Pnative -DskipTests package
# 또는 프로젝트에 있는 스크립트 사용:
# ./bin/build-native

빌드가 끝나면:

./target/fpsb boot --port 8081 --binary-port 9100
# 또는
./target/fpsb start --port 8081 --binary-port 9100

를 실행했을 때 더 이상 Unknown options 가 나오지 않고, 서버가 뜨는 게 정상입니다.
./target/fpsb boot --help 를 해 보면 이제는 -p, --port, -b, --binary-port 옵션 설명이 Usage에 같이 나와야 합니다.

———

3. Spring Boot 버전 관련해서

- 이 Picocli/GraalVM 문제는 Boot 3.3.x / 4.0.0 둘 다 공통되는 부분이라, 위 설정은 어떤 쪽에서도 필요합니다.
- 다만 native + AOT 쪽 에러(예: ApplicationMain__ApplicationContextInitializer 찾을 수 없음)는 Boot 4.0에서 AOT 규칙이 바뀐 영향
  이라,
  지금 PoC 프로젝트를 빨리 안정적으로 돌리고 싶다면:
    - parent를 다시 3.3.4 로 두고
    - <java.version>21</java.version> 으로 유지한 채
    - 위 picocli-codegen 설정을 함께 쓰는 구성이 가장 덜 고생하는 조합입니다.

---

Edited pom.xml (+59 -1)
121      <profiles>
122 -        <!-- Standard GraalVM native-image build. Does NOT use musl / static linking. -->
122 +        <!-- Legacy: single native binary (CLI + server entry). Kept for backward-compat only. -->
123          <profile>
⋮
187          </profile>
188 +
189 +        <!-- Native server binary: Spring Boot application only (no CLI), OS/arch-specific. -->
190 +        <profile>
191 +            <id>native-server</id>
192 +            <build>
193 +                <plugins>
194 +                    <plugin>
195 +                        <groupId>org.graalvm.buildtools</groupId>
196 +                        <artifactId>native-maven-plugin</artifactId>
197 +                        <version>${native.maven.plugin.version}</version>
198 +                        <extensions>true</extensions>
199 +                        <configuration>
200 +                            <!-- Example output: target/fpsb-server -->
201 +                            <imageName>fpsb-server</imageName>
202 +                            <mainClass>com.example.portable.FullyPortableApplication</mainClass>
203 +                            <fallback>false</fallback>
204 +                        </configuration>
205 +                        <executions>
206 +                            <execution>
207 +                                <id>build-native-server</id>
208 +                                <goals>
209 +                                    <goal>compile-no-fork</goal>
210 +                                </goals>
211 +                            </execution>
212 +                        </executions>
213 +                    </plugin>
214 +                </plugins>
215 +            </build>
216 +        </profile>
217 +
218 +        <!-- Native CLI binary: Picocli commands (stat, config, etc.), can talk to server over HTTP/TCP. -->
219 +        <profile>
220 +            <id>native-cli</id>
221 +            <build>
222 +                <plugins>
223 +                    <plugin>
224 +                        <groupId>org.graalvm.buildtools</groupId>
225 +                        <artifactId>native-maven-plugin</artifactId>
226 +                        <version>${native.maven.plugin.version}</version>
227 +                        <extensions>true</extensions>
228 +                        <configuration>
229 +                            <!-- Example output: target/fpsb-cli -->
230 +                            <imageName>fpsb-cli</imageName>
231 +                            <mainClass>com.example.portable.ApplicationMain</mainClass>
232 +                            <fallback>false</fallback>
233 +                        </configuration>
234 +                        <executions>
235 +                            <execution>
236 +                                <id>build-native-cli</id>
237 +                                <goals>
238 +                                    <goal>compile-no-fork</goal>
239 +                                </goals>
240 +                            </execution>
241 +                        </executions>
242 +                    </plugin>
243 +                </plugins>
244 +            </build>
245 +        </profile>
246      </profiles>

─ Worked for 58s ─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 생각하고 계신 목표(“서버 쪽은 Java 25, libc, 동적 라이브러리 모른 채 그냥 바이너리 실행만 하면 되게”)는 충분히 현실적인 목표입니다.
다만 몇 가지 “물리 법칙” 같은 제약을 이해하고, 그 안에서 안내를 잘 해줘야 합니다.

아래 기준으로 정리하겠습니다.

———

1. 어떤 아티팩트를 만들게 되는지

지금 pom.xml에 두 개 프로파일을 추가해두었습니다:

- native-server
    - mainClass = com.example.portable.FullyPortableApplication
    - 출력: target/fpsb-server
    - Spring Boot 서버 전용 네이티브 바이너리 (CLI 없음)
- native-cli
    - mainClass = com.example.portable.ApplicationMain (Picocli)
    - 출력: target/fpsb-cli
    - stat, config 같은 CLI 전용 네이티브 바이너리
      (서버와는 HTTP/TCP로 통신하는 클라이언트 역할)

빌드 예시 (현재 macOS에서):

# 서버 네이티브
mvn -Pnative-server -DskipTests package

# CLI 네이티브
mvn -Pnative-cli -DskipTests package

이렇게 하면 둘 다 실행할 때 Java 25가 전혀 필요 없습니다.
(빌드할 때만 GraalVM JDK 25가 필요)

———

2. “완전 이식성”에서 현실적인 하한/상한

네이티브 바이너리는 결국 OS + CPU 아키텍처에 종속입니다.
완전한 의미의 “어디서나 한 파일”은 불가능하고, 현실적인 단위는:

- fpsb-server-macos-aarch64
- fpsb-server-linux-x86_64
- fpsb-server-linux-aarch64
- fpsb-server-windows-x86_64.exe
- fpsb-cli-... (각 OS별)

정도로 OS/아키텍처별 바이너리 하나씩입니다.

그리고 각 OS마다 “최소 환경”은 이렇게 안내할 수 있습니다.

macOS용 바이너리

- 필수: 동일 아키텍처 (예: M1/M2 → arm64)
- OS 버전: “빌드한 macOS 버전 이상”으로 보는 게 안전
    - 예: macOS 14에서 빌드했다면 “macOS 14 이상에서 테스트됨” 정도로 안내
- 추가 런타임 의존성: 없음 (Apple의 표준 libSystem 등은 OS에 기본 포함)

Linux용 바이너리

- 두 가지 전략:
    1. 일반 glibc 동적 링크
        - “glibc 최소 버전 + 커널 버전”을 어느 정도 가정해야 함
            - 보통 Ubuntu 20.04 / CentOS 8 수준의 glibc 이상이면 안전
        - 문서에 대략:
            - “Linux x86_64, glibc 2.28+, Linux kernel 4.x 이상” 정도로 표기
    2. --static --libc=musl 로 빌드한 “fat” 바이너리 (이미 native-musl 프로파일에서 사용 중인 방식)
        - glibc 의존이 없어져서, “대부분의 현대 Linux 배포판”에서 돌기 쉬움
        - 여전히 **커널 버전(아주 옛날 커널은 안 될 수 있음)**은 하한이 있으나, 실무에서는
            - “Linux x86_64, 커널 4.x 이상(또는 3.10+)” 정도로 안내

Windows용 바이너리

- fpsb-server.exe / fpsb-cli.exe
- 필수: 동일 아키텍처 (x86_64)
- OS 버전: “Windows Server 2019/Windows 10 이상에서 테스트” 정도로 안내
- 추가 DLL 의존성은 거의 없고, 기본 Win32 API에 링크됩니다.

정리하면 문서에 이렇게 적을 수 있습니다:

- “이 바이너리는 JVM이나 GraalVM 설치 없이 실행 가능하지만,
  OS/아키텍처별로 빌드된 버전 중 하나를 사용해야 합니다.”
- “Linux용 바이너리는 glibc 동적 링크 버전과, musl 기반 정적 링크 버전이 있으며,
  후자는 추가 라이브러리 설치 없이 대부분의 최신 리눅스에서 실행 가능합니다.”

———

3. “서버에서 아무것도 몰라도 된다”를 어떻게 보장하나

당신이 해야 할 일은:

- 빌드 시점에: 각 OS/아키텍처용 GraalVM JDK에서 -Pnative-server, -Pnative-cli 로 바이너리 생성
    - macOS에서는 macOS용만,
    - Linux/Windows는 각각 해당 OS에서, 혹은 그 OS용 컨테이너/VM 안에서 빌드
- 배포 시점에는:
    - 해당 OS용 fpsb-server 바이너리만 전달
    - 서버에는:
        - “이 파일에 실행 권한만 주고 실행하면 된다”라고 안내
        - 추가로 필요한 건 포트 열기, systemd 서비스 등록 같은 운영 이슈뿐

즉, 서버 팀/운영자는:

- “java가 몇 버전인지, GraalVM이 뭔지, libc가 뭔지”를 몰라도 되고,
- 문서에는 단순히:
    - “이 실행 파일은 Linux x86_64 (kernel ≥ 4.4) 에서 테스트 되었고,
      추가 라이브러리나 JVM 없이 실행 가능합니다.”
    - “macOS/Windows는 각각 이 파일을 사용하고, 똑같이 JVM 필요 없습니다.”

라고만 써주면 됩니다.

———

4. 지금 상태에서 다음 단계

당신의 목표를 기준으로 하면, 다음 순서가 이상적입니다.

1. pom.xml 정리
    - 이미 native-server, native-cli 프로파일 추가해 두었습니다.
    - Spring Integration 스냅샷/버전 문제는 Boot BOM에 맞춰 정리(앞에서 설명한 대로)하면 서버 부팅 에러는 사라집니다.
2. 각 OS에서 빌드
    - macOS: mvn -Pnative-server -DskipTests package → fpsb-server-macos-arm64
    - Linux: mvn -Pnative-server -Pnative-musl -DskipTests package → fpsb-server-linux-x86_64 (musl static)
    - Windows: mvn -Pnative-server -DskipTests package → fpsb-server.exe
3. 문서에 “지원 플랫폼 매트릭스”와 “최소 요구사항”을 표 형태로 정리

---

• 목표(서버 쪽에서는 Java 25, libc 등을 전혀 몰라도 되고, 그냥 바이너리만 돌리게 하는 것)에 맞게 정리해 두었습니다. 지금 상태 기준으로
정리하면 이렇습니다.

1. 두 개의 네이티브 바이너리 프로파일

pom.xml 에 다음 네이티브 프로파일이 있습니다.

- native-server
    - mainClass = com.example.portable.FullyPortableApplication
    - 출력: target/fpsb-server (OS/아키텍처별 바이너리)
    - 용도: HTTP + 바이너리 TCP 서버 전용
- native-cli
    - mainClass = com.example.portable.ApplicationMain (Picocli)
    - 출력: target/fpsb-cli
    - 용도: stat, config 같은 CLI 도구 (서버와 HTTP/TCP로 통신)

빌드 예시는:

# 서버 네이티브 (현 OS/아키텍처용)
mvn -Pnative-server -DskipTests package

# CLI 네이티브
mvn -Pnative-cli -DskipTests package

이렇게 만든 fpsb-server, fpsb-cli 는 실행할 때 JDK가 필요 없습니다.
(GraalVM JDK 25는 빌드할 때만 필요)

2. Spring Integration 버전 문제 해결

NoClassDefFoundError: MapAccessor 문제는 Spring Integration 7 스냅샷과 Boot 3.3.4 조합 때문이라:

- spring-snapshot 리포지토리 제거
- spring-integration-bom 7.0.0-SNAPSHOT 제거
- spring-integration-ip, spring-integration-test 는 버전 없이 두고, Boot BOM에 맡김

으로 맞춰 두었습니다. 이제 Boot 3.3.4와 호환되는 Integration 버전이 자동으로 선택됩니다.

3. README 정리 – 어떤 환경에서 어떻게 쓸 수 있는지

README.md 를 다음 관점으로 정리했습니다.

- 실행 형태 3가지
    - JAR + JDK (개발/디버깅용, ./bin/fpsb boot ...)
    - 네이티브 서버 (fpsb-server): Spring Boot 서버 전용
    - 네이티브 CLI (fpsb-cli): Picocli 도구
- 서버용 네이티브 빌드/실행 예시

  mvn -Pnative-server -DskipTests package
  ./target/fpsb-server --server.port=8080 --binary.server.port=9090
- CLI 네이티브 빌드/사용 예시

  mvn -Pnative-cli -DskipTests package

  ./target/fpsb-cli stat --host localhost --port 8080 --binary-port 9090
  ./target/fpsb-cli config
- 지원 플랫폼/하한 안내(예시)

  README에 “운영 환경에서 어떤 하한을 가정하는지”를 이렇게 적을 수 있도록 틀을 잡아 두었습니다:
    - macOS (arm64)
        - 예: macOS 14 이상에서 빌드/테스트
        - 추가 JDK 없이 fpsb-server, fpsb-cli 실행
    - Linux (x86_64)
        - 일반 빌드: glibc 기반 배포판 (Ubuntu 20.04+, CentOS 8+ 등)
        - native-musl 조합: musl 정적 링크로 대부분의 최신 Linux 배포판에서 실행 가능
    - Windows (x86_64)
        - Windows 10 / Server 2019 이상에서 빌드/테스트 (예시)
        - .exe 형태로 배포

실제로는 각 타깃 OS/아키텍처에서 -Pnative-server, -Pnative-cli로 바이너리를 빌드하고,
그 환경에서 테스트해 본 뒤 “이 바이너리는 ○○ OS/버전 이상에서 검증됐다”고 문서에 구체적으로 적어주시면 됩니다.

이제부터는:

- 서버 쪽: 대상 서버에는 그냥 fpsb-server 하나만 배포해도 되고, JDK는 전혀 신경 쓰지 않아도 됩니다.
- 운영/관리 CLI: 운영자의 노트북이나 Bastion 등에서 fpsb-cli (또는 JAR + bin/fpsb)를 사용해 서버 상태 점검/설정 안내를 하면 됩니다.


---

## 1. “plain JAR” vs Spring Boot 실행 JAR(fat JAR)

먼저 JAR 형태부터 정리해야 합니다.

### 1-1. plain JAR (얇은 JAR, thin JAR)

- 일반 Maven 기본 JAR가 이 형태입니다.
- 구조는 대략:

  META-INF/MANIFEST.MF
  com/example/portable/FullyPortableApplication.class
  com/example/portable/ApplicationMain.class
  ...
- 특징:
    - 내 클래스들만 들어있고, 의존 라이브러리들은 들어있지 않습니다.
    - 실행할 때는 보통:

      java -cp myapp.jar:lib1.jar:lib2.jar ... com.example.Main
    - 즉, -cp(classpath)에 JAR 여러 개를 나열해서 사용합니다.
    - GraalVM native-image도 이 모델을 상정합니다:
        - -cp에 JAR 여러 개를 주고,
        - 그 중에서 com.example.portable.FullyPortableApplication 같은 클래스를 classpath 루트에서 찾습니다.

### 1-2. Spring Boot 실행 JAR (fat JAR, repackage / bootJar)

Spring Boot는 “단일 실행 JAR”를 위해 자체 레이아웃을 씁니다. 아까 jar tf 할 때 보셨던 것처럼:

META-INF/MANIFEST.MF
org/springframework/boot/loader/... (부트 로더)
BOOT-INF/classes/com/example/portable/FullyPortableApplication.class
BOOT-INF/lib/dependency1.jar
BOOT-INF/lib/dependency2.jar
...

- 클래스가 BOOT-INF/classes/... 아래에 들어가 있고,
- 라이브러리들은 BOOT-INF/lib/* 아래에 중첩된 JAR로 들어갑니다.
- 이 구조는 Spring Boot 로더(org.springframework.boot.loader.)가 이해하는 포맷이지,
  일반적인 java -cp some.jar com.example.Main 이 곧바로 이해하는 포맷이 아닙니다.
    - 그래서 Boot는 “JarLauncher” 같은 로더 클래스를 통해 이 JAR를 실행합니다:

      java -jar fully-portable-spring-boot-0.0.1-SNAPSHOT.jar
      # 내부적으로는 JarLauncher가 BOOT-INF/classes, BOOT-INF/lib를 풀어서 classpath 를 구성
- GraalVM native-image 입장에서는:
    - 단순히 -cp fully-portable-spring-boot-0.0.1-SNAPSHOT.jar com.example.portable.FullyPortableApplication
    - 로 주어졌을 때 JAR 루트에서 com/example/...를 찾으려고 합니다.
    - 하지만 실제 클래스는 BOOT-INF/classes/com/example/... 아래 있기 때문에 “클래스가 없다”고 보는 겁니다.

이게 바로 에러 메시지의 의미였습니다:

> Main entry point class 'com.example.portable.FullyPortableApplication' neither found on classpath: '...fully-portable-spring-boot-
> 0.0.1-SNAPSHOT.jar ...'

= “이 JAR를 그냥 classpath에 올렸더니, 루트에 com/example/portable/FullyPortableApplication.class 가 없다.”

———

## 2. Maven 빌드 순서에서 무슨 일이 일어나는지

mvn package 를 할 때 Maven이 대략 이렇게 움직입니다:

1. maven-compiler-plugin: target/classes/... 에 .class 파일 생성
2. maven-jar-plugin (기본):
    - target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar (plain JAR) 생성
3. spring-boot-maven-plugin:repackage:
    - 방금 만든 plain JAR를 읽어서 Spring Boot 실행 JAR로 “다시 패키징(repackage)”
    - 원래 JAR는 ...jar.original 로 이름을 바꾸고,
    - 실행 JAR(fat JAR)를 ...jar 자리로 덮어씁니다.

그 결과:

- target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar : Boot 실행 JAR (fat JAR)
- target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar.original : plain JAR

그리고 native-maven-plugin 은 기본적으로 “메인 아티팩트”인 ...jar 를 인풋으로 씁니다.

- 우리가 mainClass = com.example.portable.FullyPortableApplication 으로 줬으니,
- 이 JAR를 평범한 classpath JAR 라고 생각하고 FullyPortableApplication 클래스를 찾으려다가 실패한 겁니다.

———

## 3. 왜 spring-boot.repackage.skip 으로 해결하는가

우리 목표: native-server 빌드를 할 때만 “plain JAR를 유지”하고 싶다.

그래서 pom.xml 에 이런 구조를 넣었습니다.

### 3-1. 기본값: 평소에는 repackage 켜두기

  <properties>
      ...
      <spring-boot.repackage.skip>false</spring-boot.repackage.skip>
  </properties>

  <build>
    <plugins>
      <plugin>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-maven-plugin</artifactId>
        <configuration>
          <mainClass>com.example.portable.FullyPortableApplication</mainClass>
          <skip>${spring-boot.repackage.skip}</skip>
        </configuration>
      </plugin>
    </plugins>
  </build>

- 평소에는 spring-boot.repackage.skip=false 이므로,
- spring-boot:repackage 가 실행되며,
  target/....jar 는 Boot 실행 JAR가 됩니다.
- 이 모드는:
    - ./bin/fpsb 로 JAR을 실행할 때
    - JVM 모드에서 개발/디버깅할 때
    - 즉, “Boot 방식 fat JAR”로 실행하고 싶은 경우에 필요합니다.

### 3-2. -Pnative-server 일 때만 repackage 끄기

native-server 프로파일 안에 이렇게 넣었습니다:

  <profile>
    <id>native-server</id>
    <properties>
      <!-- Skip Boot repackage so the main artifact is a plain JAR (classes at root). -->
      <spring-boot.repackage.skip>true</spring-boot.repackage.skip>
    </properties>
    <build>
      <plugins>
        <plugin>
          <groupId>org.graalvm.buildtools</groupId>
          <artifactId>native-maven-plugin</artifactId>
          <version>${native.maven.plugin.version}</version>
          <extensions>true</extensions>
          <configuration>
            <imageName>fpsb-server</imageName>
            <mainClass>com.example.portable.FullyPortableApplication</mainClass>
            <fallback>false</fallback>
          </configuration>
          ...
        </plugin>
      </plugins>
    </build>
  </profile>

- mvn -Pnative-server ... 로 빌드하면:
    - 이 프로파일의 <properties>가 활성화됩니다.
    - 그 결과, spring-boot.repackage.skip=true 가 되어,
    - spring-boot:repackage 플러그인이 실행을 건너뜁니다(skip).
- 그러면 package 후에 남는 target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar 는:
    - 더 이상 Boot 실행 JAR가 아니고,
    - 원래 jar 플러그인이 만든 plain JAR 그대로입니다.

이 plain JAR의 구조는:

META-INF/MANIFEST.MF
com/example/portable/FullyPortableApplication.class
com/example/portable/ApplicationMain.class
...
(의존 라이브러리는 별도의 JAR)

그래서 native-image 가 -cp 에 이 JAR를 올리고:

native-image ... -cp fully-portable-spring-boot-0.0.1-SNAPSHOT.jar:... \
com.example.portable.FullyPortableApplication

를 실행하면, 정상적인 classpath 규칙으로 FullyPortableApplication 클래스를 찾아냅니다.

요약하면:

- Plain JAR: 클래스들이 루트에 (com/example/...) 있고, 일반 classpath 규칙으로 접근 가능
- Boot fat JAR: 클래스가 BOOT-INF/classes/... 아래 있고, Boot 로더를 통해서만 접근 가능
- spring-boot.repackage.skip 을 프로파일별로 조절해서,
    - JVM/실행 JAR 모드에서는 fat JAR를 쓰고,
    - native-server 빌드에서는 plain JAR를 쓰도록 스위칭한 것입니다.

———

## 4. “plain JAR 말고 또 뭐가 있나요?”에 대한 간단 정리

JAR 쪽에서 흔히 보는 몇 가지 형태만 정리해보면:

- Plain JAR (thin JAR) – Maven 기본
    - 클래스/리소스만 들어있고, 의존성은 별도 JAR.
- Spring Boot 실행 JAR (fat JAR, repackage/bootJar) – 지금 이야기한 것
    - Boot 로더 + BOOT-INF 구조.
    - java -jar 로 단일 JAR 실행.
- Shaded JAR (maven-shade-plugin)
    - 의존 JAR들의 클래스까지 한 JAR 안으로 다 “복사/병합”해서 넣는 형태.
    - 충돌 방지를 위해 패키지 이름을 relocate 하는 경우도 있음.
- WAR
    - 톰캣/와일드플라이 등 서블릿 컨테이너 배포용 포맷.

GraalVM native-image 는 기본적으로 “plain JAR + 별도 의존 JAR” 조합을 classpath에 올리는 모델에 가장 잘 맞습니다.
Spring Boot fat JAR도 Graal이 인식하게 만들 수는 있지만(추가 작업/플러그인 연계 필요),
지금 이 프로젝트에서는 **“native-server 빌드 시에만 plain JAR로 유지”**하는 쪽이 훨씬 단순하고 명확합니다.

———

정리하면:

- native-server에서는 plain JAR를 사용해서 GraalVM이 FullyPortableApplication 을 제대로 찾도록 했고,
- JVM 모드나 ./bin/fpsb 같은 경로에서는 여전히 Spring Boot 실행 JAR(fat JAR) 를 사용할 수 있도록
  spring-boot.repackage.skip 를 프로파일별로 스위칭하고 있습니다.

---

## 1. 에러의 의미 정리

에러:

> AotInitializerNotFoundException: Startup with AOT mode enabled failed: AOT initializer
> com.example.portable.FullyPortableApplication__ApplicationContextInitializer could not be found

Spring Boot 3.3 + GraalVM 네이티브 통합 구조는 대략:

1. spring-boot-maven-plugin:process-aot 가
    - FullyPortableApplication__ApplicationContextInitializer 같은 AOT 초기화 클래스를 빌드 시점에 생성
2. 그 다음 native-maven-plugin 이
    - 이 AOT 클래스가 포함된 classpath 를 기반으로 native-image 를 호출
3. 네이티브 실행 시:
    - Boot 내부에서 spring.aot.enabled=true 인 걸 보고
      FullyPortableApplication__ApplicationContextInitializer 를 찾아서 실행

지금은 3번까지는 그대로인데, 1번에서 우리가 AOT 클래스를 안 만들도록 해놓은 상태라서,
실행 시점에 “AOT 모드를 켰는데 initializer가 없네?” → 이 예외가 터진 겁니다.

———

## 2. 왜 AOT 클래스가 안 만들어졌나 (우리 설정의 영향)

우리가 했던 변경:

- native-server 프로파일에서, plain JAR를 쓰기 위해 Boot의 repackage를 끄고 싶었습니다.
- 그래서 처음에는 spring-boot-maven-plugin 에서 <skip>${spring-boot.repackage.skip}</skip> 을 달고,
- native-server 프로파일 안에서:

  <properties>
    <spring-boot.repackage.skip>true</spring-boot.repackage.skip>
  </properties>

  를 넣었죠.

문제는:

- spring-boot.repackage.skip 이라는 프로퍼티는 원래 “repackage goal만” 끄려고 있는 속성인데,
- 우리가 <skip> 이라는 플러그인 전체 스킵 옵션에 묶어버려서,
    - native-server 때는 spring-boot-maven-plugin 의 모든 goal 이 통째로 꺼져 버렸습니다.
    - 그 안에는 repackage 뿐 아니라 process-aot 도 포함됩니다.
- 그래서 -Pnative-server 빌드에는:
    - spring-boot:process-aot 가 전혀 실행되지 않았고,
    - FullyPortableApplication__ApplicationContextInitializer 도 생성되지 않았던 겁니다.

그런데 native-maven-plugin 쪽에서는:

- 네이티브 빌드용으로 spring.aot.enabled=true 같은 설정을 자동으로 넣고,
- Boot가 “나는 AOT 모드로 실행 중이야”라고 인식하게 만듭니다.

결과:

- AOT 모드는 켜져 있는데 (spring.aot.enabled=true)
- AOT initializer 클래스는 존재하지 않아서 지금 예외가 터진 상태입니다.

———

## 3. 무엇을 고쳤는가

핵심은:

- repackage만 끄고,
- spring-boot:process-aot 같은 나머지 goal은 그대로 실행되게 하는 것.

그래서 pom.xml에서 다음을 했습니다.

1. <skip>${spring-boot.repackage.skip}</skip> 삭제

   <plugin>
     <groupId>org.springframework.boot</groupId>
     <artifactId>spring-boot-maven-plugin</artifactId>
     <configuration>
       <!-- AOT/bootJar용 메인 클래스 지정만 남김 -->
       <mainClass>com.example.portable.FullyPortableApplication</mainClass>
     </configuration>
   </plugin>
    - 이렇게 하면 <skip> 으로 플러그인 전체를 끄지 않습니다.
    - spring-boot.repackage.skip 프로퍼티는 이제 Boot 플러그인 내부에서
      정상적으로 “repackage goal만 skip” 하는 용도로 사용됩니다.
2. native-server 프로파일에서는 그대로:

   <profile>
     <id>native-server</id>
     <properties>
       <!-- 이 값이 true일 때, repackage goal만 skip 된다 -->
       <spring-boot.repackage.skip>true</spring-boot.repackage.skip>
     </properties>
     ...
   </profile>

이렇게 되면 -Pnative-server 빌드 순서는:

1. compile → jar 플러그인이 plain JAR 생성
2. spring-boot:process-aot 실행 →
    - FullyPortableApplication__ApplicationContextInitializer 등 AOT 클래스 생성
3. spring-boot:repackage 는 spring-boot.repackage.skip=true 덕분에 건너뜀
    - 그래서 최종 target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar 는 여전히 plain JAR
4. native-maven-plugin:compile-no-fork 가
    - 위 plain JAR + AOT 클래스들을 classpath에 올리고
    - GraalVM native-image 를 호출해 fpsb-server 생성

실행 시에는:

- Boot가 spring.aot.enabled=true 를 보고 AOT 모드로 부팅하고,
- 이제 classpath에 FullyPortableApplication__ApplicationContextInitializer 가 있기 때문에
  더 이상 AotInitializerNotFoundException 을 던지지 않습니다.

———

## 4. plain JAR / fat JAR / AOT / native-image 관계를 한 번에 요약하면

- Spring Boot fat JAR (repackage/bootJar)
    - 사람/운영자 입장: java -jar app.jar 한 번이면 되니까 편함
    - 하지만 native-image 인풋으로 쓰기는 까다로움(BOOT-INF 구조 등)
- plain JAR
    - GraalVM native-image 가 가장 좋아하는 형태 (정상 classpath 규칙)
    - Boot 3.3의 AOT(process-aot) 도 plain JAR 기반으로 생성된 클래스를 사용
- 우리 전략
    - 평소(JVM 모드)에는 Boot repackage를 써서 fat JAR(java -jar, ./bin/fpsb) 사용
    - -Pnative-server 때만:
        - spring-boot.repackage.skip=true 로 repackage(=fat JAR 생성)만 끄고
        - process-aot 는 그대로 실행시켜서 AOT initializer를 생성
        - 그 plain JAR + AOT 클래스를 가지고 GraalVM native-image 로 fpsb-server 생성

지금 구조는 이 흐름에 맞춰 정리해 둔 상태입니다.

———

실제 Mac 환경에서는:

export JAVA_HOME="$PWD/bin/graalvm/macos"
export PATH="$JAVA_HOME/bin:$PATH"

mvn clean -Pnative-server -DskipTests package
./target/fpsb-server --server.port=8080 --binary.server.port=9090

이렇게 했을 때 더 이상 AotInitializerNotFoundException 이 뜨지 않아야 합니다.
(만약 OS 쪽 tmp 권한 문제 같은 게 또 보이면, 그건 별도의 시스템 권한 이슈이고,
AOT initializer 관련 문제는 위처럼 정리된 상태라고 보시면 됩니다.)