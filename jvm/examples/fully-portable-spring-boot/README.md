# Fully Portable Spring Boot (PoC)

Spring Boot + Picocli 기반으로 만든 완전 이식 가능한(native-image) PoC 프로젝트입니다.

핵심 포인트:

- Spring Boot 3 (AOT + GraalVM native-image)
- Picocli 기반 CLI (`start`, `stat`, `config`)
- HTTP 서버 + 고정 길이 바이너리 전문(예: 300 byte) 수신용 TCP 서버
- GraalVM + musl 조합으로 Linux 환경용 정적 링크 바이너리 프로파일 제공

> 참고: 현재 `pom.xml` 의 `java.version` 은 21로 설정되어 있습니다.  
> 실제 Java 25 JDK 환경이 준비되면 `-Djava.version=25` 를 넘기거나 `pom.xml` 속성을 25로 변경해서 빌드하면 됩니다.

## 사전 준비

Graal VM JDK 25를 설치합니다.

```shell
brew install --cask graalvm-jdk@25
```

## 빌드 개요

이 프로젝트는 세 가지 실행 형태를 제공합니다.

1. **JAR + 내장 JDK (JVM 모드, 개발/디버깅용)**  
2. **네이티브 서버 바이너리 (`fpsb-server`, Spring Boot 전용)**  
3. **네이티브 CLI 바이너리 (`fpsb-cli`, Picocli 전용 도구)**  
4. **JVM OCI 이미지 + udocker 런처 (`run.sh`, 동적 플러그인/JAR 로딩용)**

운영 환경에서는 2, 3번처럼 **JDK 없이도 실행 가능한 네이티브 바이너리**를 사용하는 것을 목표로 합니다.

## 1) JAR + 내장 JDK 25 (개발/디버깅용)

일반 JVM용 실행 JAR:

```bash
mvn clean package
```

생성물: `target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar`

이 JAR 는 프로젝트에 함께 포함된 GraalVM JDK 25 (`bin/graalvm/...`) 또는  
`JAVA_HOME` 에 설정된 JDK 로 실행할 수 있습니다.

## JAR + 내장 JDK 25 (PoC)

1. JAR 빌드

   ```bash
   mvn clean package
   ```

2. `bin/jdk-25` 디렉터리에 JDK 25 배포판을 풀어둡니다. 예:

   ```bash
   mkdir -p bin
   tar -xf openjdk-25_linux-x64.tar.gz -C bin
   mv bin/jdk-25* bin/jdk-25
   ```

3. 실행(Unix 계열, Picocli CLI + Spring Boot 를 한 프로세스에서 실행)

   ```bash
   chmod +x bin/fpsb
   ./bin/fpsb boot --port 8080 --binary-port 9090
   ```

   - `bin/fpsb` 는 우선순위대로 다음을 사용합니다.
     1) `bin/jdk-25/bin/java`  
     2) `JAVA_HOME` 환경변수  
     3) PATH 상의 `java`

기본 엔드포인트:

- HTTP: `GET http://localhost:8080/hello`
- Actuator Health: `GET http://localhost:8080/actuator/health`
- 바이너리 전문 서버: `tcp://localhost:9090` (300 byte 고정 메시지)

예: `nc` 로 테스트

```bash
printf '%-300s' 'TEST MESSAGE' | nc localhost 9090
```

## 2) GraalVM Native Image – 서버용 (`fpsb-server`)

### 사전 준비

- GraalVM (또는 Oracle/GraalVM CE) 설치
- `native-image` 유틸리티 설치
- Linux + musl 정적 링크를 사용할 경우:
  - `musl-gcc` 등 musl 툴체인 설치

### 일반 native 빌드 (현재 OS/아키텍처용 서버 바이너리)

```bash
mvn -Pnative-server -DskipTests package
```

생성물 (예시):

- `target/fpsb-server` (플랫폼/OS 에 따라 확장자/이름이 달라질 수 있음)

### Linux + musl 정적 링크 빌드

동적 libc(glibc) 의존성을 없애서 최대한 “fat binary”에 가깝게 만드는 프로파일입니다.

```bash
mvn -Pnative-musl -DskipTests package
```

- 생성물: `target/fpsb-server-musl`
- 빌드 시 GraalVM 이 `--static --libc=musl` 옵션으로 native-image 를 실행합니다.  
  실제로는 빌드 환경에 musl 툴체인이 제대로 설치되어 있어야 합니다 (예: `musl-gcc`).

### 서버 바이너리 실행 예시

```bash
# native-image 빌드 후 (예: Linux/x86_64 환경)
./target/fpsb-server --server.port=8080 --binary.server.port=9090
```

## 3) 네이티브 CLI (`fpsb-cli`)

서버와 통신하는 Picocli 기반 CLI 를 네이티브로 빌드할 수도 있습니다.

```bash
mvn -Pnative-cli -DskipTests package
```

생성물 (예시):

- `target/fpsb-cli`

사용 예시:

```bash
# 서버 상태 확인
./target/fpsb-cli stat --host localhost --port 8080 --binary-port 9090

# 설정 도움말
./target/fpsb-cli config
```

## 구조

- `ApplicationMain` – 단일 바이너리/프로세스의 진입점, Picocli 기반 CLI (`fpsb`)
- `FullyPortableApplication` – Spring Boot @SpringBootApplication
- `binary.BinaryProtocolServer` – 고정 길이(기본 300 byte) 바이너리 전문 서버
- `binary.BinaryServerProperties` – 포트/전문 길이 설정용 @ConfigurationProperties
- `web.HelloController` – HTTP 예제 엔드포인트
- `cli.*Command` – `start`, `stat`, `config` Picocli 커맨드들

## 지원 플랫폼 / 런타임 요구사항(예시 안내)

실제 배포 시에는 OS/아키텍처 별로 각각 빌드된 바이너리를 사용해야 합니다.

- macOS (arm64)
  - 예: macOS 14 (Sonoma) 이상에서 빌드/테스트
  - 추가 JDK 설치 없이 `fpsb-server`, `fpsb-cli` 실행 가능
- Linux (x86_64)
  - 일반 빌드: glibc 기반 배포판 (예: Ubuntu 20.04+, CentOS 8+ 등)
  - `native-musl` 조합: musl 정적 링크로 대부분의 최신 Linux 배포판에서 실행 가능
- Windows (x86_64)
  - Windows 10 / Windows Server 2019 이상에서 빌드/테스트 (예시)
  - `.exe` 형태로 배포 가능

위 요구사항은 예시이며, 실제 운영 환경에서는 각각의 타깃 OS/아키텍처에서  
`-Pnative-server`, `-Pnative-cli` 프로파일을 사용해 바이너리를 빌드한 후,  
해당 환경에서 테스트한 결과를 기반으로 “지원 OS/버전”을 문서화하는 것을 권장합니다.

## 4) JVM OCI 이미지 + udocker 런처

native-image + musl 방식으로는 리플렉션/동적 클래스 로딩/플러그인 JAR 로딩에 제약이 크기 때문에,
일반 JVM 모드(HotSpot)로 실행되는 OCI 이미지를 만들고, Docker 없이도 실행 가능하도록
`udocker` + `run.sh` 조합을 제공합니다.

구조 요약:

- `Dockerfile.jvm` – `mvn package` 로 생성한 fat jar를 기반으로 하는 JVM용 OCI 이미지 정의
- `bin/build-oci-image` – Docker로 이미지를 빌드하고 `download/fpsb-jvm-latest.tar` 로 export
- `bin/udocker/udocker` – 로컬 `bin/udocker` Python 모듈을 사용하는 udocker 래퍼
- `run.sh` – udocker를 초기화하고, OCI 이미지를 import/실행하는 단일 진입점
  - 컨테이너 안에서 `/opt/app/app.jar` 를 `java -jar` 로 실행
  - 호스트의 `./plugins` 디렉터리를 컨테이너 `/opt/app/plugins` 로 마운트
  - 환경 변수 `FPSB_PLUGINS_DIR=/opt/app/plugins` 를 설정 (애플리케이션이 플러그인 디렉터리로 사용할 수 있음)

### 빌드 및 실행 흐름 (요약)

1. JVM fat jar 빌드 및 OCI 이미지 생성 (Docker가 있는 환경에서):

   ```bash
   bin/build-oci-image
   ```

   - 기본 이미지 이름: `fpsb-jvm:latest`
   - export 경로: `download/fpsb-jvm-latest.tar`

2. 배포 대상(리눅스 x86_64 등)에 다음 파일을 배치:

   - `run.sh`
   - `bin/udocker/**` (udocker Python 모듈)
   - `download/fpsb-jvm-latest.tar` (또는 `FPSB_IMAGE_TAR` 로 지정한 tar 파일)

3. 배포 대상에서 실행:

   ```bash
   chmod +x run.sh bin/udocker/udocker bin/build-oci-image
   ./run.sh
   ```

   - `run.sh` 는 최초 실행 시 `.udocker/` 아래에 로컬 udocker 리포지토리를 초기화합니다.
   - `download/fpsb-jvm-latest.tar` 가 로컬 udocker 이미지로 import 됩니다.
   - 컨테이너는 호스트의 `./plugins` 디렉터리를 `/opt/app/plugins` 로 마운트합니다.

4. 플러그인/동적 JAR 로딩

   - 애플리케이션이 `/opt/app/plugins` (또는 `FPSB_PLUGINS_DIR`) 를 플러그인 디렉터리로 사용하도록 설계되어 있다면,
     호스트에서 `./plugins` 아래에 JAR을 추가/갱신하는 것만으로 컨테이너 안에서 해당 JAR을 읽어 동적 클래스 로딩을 할 수 있습니다.


## 다음 단계 아이디어

- Picocli 서브커맨드에 실제 운영/상태/설정 정보를 더 풍부하게 노출
- 전문 포맷(예: 고정 필드 오프셋/길이)을 별도 DTO/파서로 분리
- 빌드 파이프라인(CI)에서 linux-x86_64 / aarch64 등 타깃별 native-image 빌드 매트릭스 구성
