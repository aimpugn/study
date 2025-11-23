# Java file

- [Java file](#java-file)
    - [Java 파일 I/O](#java-파일-io)
        - [작은 파일의 간단한 입출력: NIO `Files` 이용](#작은-파일의-간단한-입출력-nio-files-이용)
        - [대용량 또는 스트리밍 I/O: 버퍼드 스트림 (`BufferedReader/Writer`, `BufferedInputStream/OutputStream`)](#대용량-또는-스트리밍-io-버퍼드-스트림-bufferedreaderwriter-bufferedinputstreamoutputstream)
        - [대용량 파일의 효율적 복사 및 전송: `FileChannel.transferTo()/transferFrom()` (Zero-Copy)](#대용량-파일의-효율적-복사-및-전송-filechanneltransfertotransferfrom-zero-copy)
    - [4. 랜덤 액세스 및 초고속 파일 I/O – 메모리 맵핑 (`MappedByteBuffer`)](#4-랜덤-액세스-및-초고속-파일-io--메모리-맵핑-mappedbytebuffer)
    - [5. 논블로킹/비동기 파일 I/O – `AsynchronousFileChannel` + `CompletableFuture` (NIO2의 AIO)](#5-논블로킹비동기-파일-io--asynchronousfilechannel--completablefuture-nio2의-aio)
    - [6. 파일 변경 감시 – `WatchService` (NIO2)](#6-파일-변경-감시--watchservice-nio2)
    - [7. ZIP/JAR 등 압축 파일 시스템 및 커스텀 파일시스템 – `FileSystems.newFileSystem(URI, ...)`](#7-zipjar-등-압축-파일-시스템-및-커스텀-파일시스템--filesystemsnewfilesystemuri-)
    - [8. 외부 라이브러리를 활용한 파일 I/O 보조 도구](#8-외부-라이브러리를-활용한-파일-io-보조-도구)
        - [Apache Commons IO – `FileUtils`, `IOUtils` 등](#apache-commons-io--fileutils-ioutils-등)
        - [8.2 Google Guava – **Files** \& **MoreFiles** 유틸](#82-google-guava--files--morefiles-유틸)
        - [8.3 Square **Okio** – 현대적인 I/O 프레임워크](#83-square-okio--현대적인-io-프레임워크)
        - [8.4 Apache Commons **VFS** / Schlichtherle **TrueVFS** – 가상 파일시스템 통합](#84-apache-commons-vfs--schlichtherle-truevfs--가상-파일시스템-통합)
        - [8.5 Google **Jimfs** – 자바용 인메모리 파일시스템 (테스트용)](#85-google-jimfs--자바용-인메모리-파일시스템-테스트용)
    - [시스템 콜과 버퍼링](#시스템-콜과-버퍼링)
    - [효율적인 파일 생성 및 쓰기 방법](#효율적인-파일-생성-및-쓰기-방법)
        - [`Files.write()` / `Files.writeString()` (NIO)](#fileswrite--fileswritestring-nio)
        - [`BufferedWriter`](#bufferedwriter)
        - [`FileChannel`과 `ByteBuffer` (NIO)](#filechannel과-bytebuffer-nio)
        - [`MappedByteBuffer`](#mappedbytebuffer)
    - [Java `File`과 `Path`](#java-file과-path)
    - [예제](#예제)
        - [create directory and file](#create-directory-and-file)
- [Java 17과 리눅스 서버에서의 고가용성 파일 처리: 완전한 기술적 분석](#java-17과-리눅스-서버에서의-고가용성-파일-처리-완전한-기술적-분석)
    - [파일 쓰기 동작의 완전한 흐름](#파일-쓰기-동작의-완전한-흐름)
        - [1. Java 애플리케이션 레이어에서의 파일 처리](#1-java-애플리케이션-레이어에서의-파일-처리)
        - [2. JVM 레이어: 네이티브 호출과 메모리 관리](#2-jvm-레이어-네이티브-호출과-메모리-관리)
        - [3. 운영체제 레이어: 시스템 콜과 페이지 캐시](#3-운영체제-레이어-시스템-콜과-페이지-캐시)
        - [4. 파일 시스템 레이어: ext4와 저널링](#4-파일-시스템-레이어-ext4와-저널링)
        - [5. 블록 레이어와 I/O 스케줄링](#5-블록-레이어와-io-스케줄링)
        - [6. 스토리지 디바이스: 물리적 쓰기와 내부 캐시](#6-스토리지-디바이스-물리적-쓰기와-내부-캐시)
    - [동기화 메커니즘과 내구성 보장](#동기화-메커니즘과-내구성-보장)
        - [fsync()와 fdatasync()의 차이점](#fsync와-fdatasync의-차이점)
        - [O\_DIRECT와 페이지 캐시 우회](#o_direct와-페이지-캐시-우회)
    - [스트리밍과 대용량 파일 처리](#스트리밍과-대용량-파일-처리)
        - [스트리밍의 본질](#스트리밍의-본질)
        - [디스크 단편화와 성능 영향](#디스크-단편화와-성능-영향)
        - [병렬 처리와 블록 쓰기 순서](#병렬-처리와-블록-쓰기-순서)
    - [고가용성과 내결함성 패턴](#고가용성과-내결함성-패턴)
        - [Write-Ahead Logging (WAL) 패턴](#write-ahead-logging-wal-패턴)
        - [원자적 파일 쓰기 패턴](#원자적-파일-쓰기-패턴)
        - [파일 잠금과 동시성 제어](#파일-잠금과-동시성-제어)
    - [메모리 관리와 성능 최적화](#메모리-관리와-성능-최적화)
        - [JVM 메모리 모델과 파일 I/O](#jvm-메모리-모델과-파일-io)
        - [대용량 파일 처리 최적화](#대용량-파일-처리-최적화)
        - [모니터링과 성능 측정](#모니터링과-성능-측정)
        - [오류 처리와 복구 전략](#오류-처리와-복구-전략)
    - [결론](#결론)
- [Java 17과 리눅스 서버에서의 고가용성 파일 처리: 심층적 기술 분석](#java-17과-리눅스-서버에서의-고가용성-파일-처리-심층적-기술-분석)
    - [물리적 저장장치: 데이터가 실제로 저장되는 곳](#물리적-저장장치-데이터가-실제로-저장되는-곳)
        - [섹터와 블록: 저장의 기본 단위](#섹터와-블록-저장의-기본-단위)
        - [디스크 기하학적 구조와 성능](#디스크-기하학적-구조와-성능)
    - [리눅스 커널: 복잡한 I/O 시스템의 조율자](#리눅스-커널-복잡한-io-시스템의-조율자)
        - [가상 파일 시스템(VFS): 통일된 인터페이스](#가상-파일-시스템vfs-통일된-인터페이스)
        - [블록 I/O 계층: 성능 최적화의 핵심](#블록-io-계층-성능-최적화의-핵심)
        - [멀티큐 블록 레이어(blk-mq): 현대적 병렬 처리](#멀티큐-블록-레이어blk-mq-현대적-병렬-처리)
    - [메모리 관리: 페이지 캐시와 버퍼 캐시](#메모리-관리-페이지-캐시와-버퍼-캐시)
        - [페이지 캐시: 파일 데이터의 메모리 저장소](#페이지-캐시-파일-데이터의-메모리-저장소)
        - [버퍼 캐시: 메타데이터 관리](#버퍼-캐시-메타데이터-관리)
        - [메모리 매핑된 I/O: 가상 메모리의 활용](#메모리-매핑된-io-가상-메모리의-활용)
    - [ext4 파일시스템: 고성능 저널링 시스템](#ext4-파일시스템-고성능-저널링-시스템)
        - [Extent 기반 블록 할당](#extent-기반-블록-할당)
        - [멀티블록 할당자(Multi-block Allocator)](#멀티블록-할당자multi-block-allocator)
        - [지연 할당(Delayed Allocation)](#지연-할당delayed-allocation)
        - [저널링: 데이터 일관성 보장](#저널링-데이터-일관성-보장)
    - [JVM과 Native 메소드: Java에서 시스템 호출까지](#jvm과-native-메소드-java에서-시스템-호출까지)
        - [JVM 아키텍처와 메모리 구조](#jvm-아키텍처와-메모리-구조)
        - [ByteBuffer: 메모리 관리의 핵심](#bytebuffer-메모리-관리의-핵심)
        - [DirectByteBuffer의 성능 특성](#directbytebuffer의-성능-특성)
        - [JNI와 Native 메소드 호출](#jni와-native-메소드-호출)
    - [동기화와 데이터 일관성: 안전한 파일 처리](#동기화와-데이터-일관성-안전한-파일-처리)
        - [fsync와 fdatasync: 영속성 보장](#fsync와-fdatasync-영속성-보장)
        - [Write Barrier: 순서 보장](#write-barrier-순서-보장)
        - [스케줄러 선택 기준](#스케줄러-선택-기준)
    - [메모리 관리와 페이지 폴트](#메모리-관리와-페이지-폴트)
        - [페이지 폴트 처리 메커니즘](#페이지-폴트-처리-메커니즘)
        - [메모리 매핑과 성능](#메모리-매핑과-성능)
    - [실제 구현: 고가용성 파일 처리 시스템](#실제-구현-고가용성-파일-처리-시스템)
        - [완전한 트랜잭션 파일 시스템](#완전한-트랜잭션-파일-시스템)
        - [성능 모니터링과 최적화](#성능-모니터링과-최적화)
    - [결론과 실무 권장사항](#결론과-실무-권장사항)
        - [시스템 설정 최적화](#시스템-설정-최적화)
        - [모니터링 지표](#모니터링-지표)
        - [오류 처리와 복구 전략](#오류-처리와-복구-전략-1)

## Java 파일 I/O

은 텍스트 파일 처리부터 대용량 파일 복사, 메모리 맵핑, 비동기 I/O, 파일 시스템 감시, ZIP 파일시스템 및 가상 파일시스템에 등 Java에는 다양한 파일 입출력 방법들이 존재합니다.

### 작은 파일의 간단한 입출력: NIO `Files` 이용

과거 `java.io.File`과 스트림을 직접 열던 방식은 파일 존재 여부 확인, 스트림 닫기 등을 개발자가 직접 관리해야 했습니다.
NIO의 `Files` 유틸은 boilerplate 없이 예외 처리와 자원 관리를 간편화합니다.
작은 크기 또는 크기가 예측 가능한 파일에 적합하며, 전체 내용을 메모리에 올려도 부담되지 않는 경우에 효율적입니다.
하지만 매우 큰 파일에 이 방식을 쓰면 메모리 부족이 발생할 수 있습니다.

Java 7부터 도입된 NIO.2 (`java.nio.file.Files`) 유틸리티 메서드는 한두 줄의 코드로 파일을 읽거나 쓸 수 있는 고수준 API를 제공합니다. 예를 들어,
- `Files.readString()`
- `Files.writeString()`
- `Files.copy()` 등

이러한 메서드는 내부적으로 스트림을 열고 닫는 과정을 숨기며, 경로는 `Path` 객체로 표현합니다.
`Path`는 OS마다 다른 경로 구문을 추상화하여 운영체제에 독립적인 경로 표현을 가능하게 합니다.

`Files.readString(path)` 등을 호출하면 JVM은 내부적으로 해당 경로의 파일을 한 번에 통째로 읽기 위해 필요한 시스템 호출을 수행합니다.
예를 들어 파일을 읽을 때는 `open()`으로 파일 디스크립터를 얻고 `read()` 시스템 콜로 내용을 커널로부터 읽어옵니다.
`Files` 메서드는 파일 크기만큼 바이트 버퍼를 확보한 뒤 한꺼번에 읽기 때문에, 작은 파일의 경우 불필요한 반복 I/O를 줄여주는 효과가 있습니다.

파일 쓰기인 `Files.writeString()`도 마찬가지로 내부적으로 버퍼를 사용하여 한번에 기록합니다.
파일이 없으면 `StandardOpenOption.CREATE` 옵션으로 자동 생성하고, 쓰기 후에는 스트림을 자동으로 닫아 자원을 반환합니다.

시스템 콜 및 버퍼링 관점에서 본다면 작은 파일은 한두 번의 `read()`/`write()` 시스템 콜로도 처리 가능하므로, `Files` 유틸은 내부 버퍼를 이용해 내용을 모았다가 최소한의 호출로 처리합니다.
예를 들어 100바이트 텍스트 파일을 `Files.readAllBytes()`로 읽으면, OS에 한 번의 `read()` 요청만 보내서 100바이트를 한꺼번에 가져옵니다.
이러한 방식은 I/O 호출 횟수를 줄여*CPU 컨텍스트 스위칭 오버헤드를 감소시킵니다.

`Files` 메서드는 내부에서 `FileChannel`이나 스트림을 열고 닫기 때문에, 개발자가 `try-with-resources`를 직접 쓰지 않아도 자원 누수 없이 안전합니다.

```java
var filePath = Paths.get("example.txt");

// 파일에 기록할 내용 준비 (여러 줄의 문자열)
var lines = Arrays.asList(
    "Java 17: NIO Files.write() Example",
    "한 줄의 코드로 파일 읽기/쓰기가 가능합니다.",
    "내부적으로 스트림을 열고 닫아주므로 자원 누수를 걱정하지 않아도 됩니다."
);

try {
    // Files.write()는 고수준 파일 쓰기 API입니다.
    // 이 한 줄 호출로 파일 생성/열기, 데이터 쓰기, 스트림 닫기가 모두 처리됩니다.
    Files.write(filePath, lines, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
    // 옵션 설명:
    // CREATE            : 파일이 없으면 새로 생성합니다
    // TRUNCATE_EXISTING : 파일이 이미 존재하면 기존 내용을 지우고 새로 씁니다
    // (※ APPEND 옵션을 주면 기존 내용 뒤에 추가할 수도 있음)

    System.out.println("파일 쓰기 완료: " + filePath.toAbsolutePath());
} catch (IOException e) {
    System.err.println("I/O 오류 발생: " + e.getMessage());
}
```

내부적으로 Java는 파일을 열고 (`open()` 시스템 콜), 모든 `lines` 내용을 한꺼번에 모아서 (`write()` 시스템 콜) 커널의 페이지 캐시로 보냅니다.
메서드 호출 한 번에 다수의 데이터를 보내므로, 반복적인 작은 쓰기 호출을 피하여 불필요한 시스템 콜이 줄어들어 성능이 향상될 수 있습니다.

또한 `Path`를 사용하므로 NIO가 플랫폼별로 파일 경로를 알아서 처리하기 때문에 Windows면 `"C:\\...\\example.txt"`, 리눅스면 `"/home/../example.txt"`처럼 운영체제별 경로 표기 차이를 신경쓰지 않아도 됩니다.

`Files.write`는 쓰려는 데이터 양이 메모리에 무리 없이 올라갈 정도로 작거나 적당할 때 유용합니다.
예컨대 설정 파일, 로그 파일, 간단한 텍스트 출력 등에 많이 쓰이며, 코드가 매우 간결해집니다.

하지만 파일 크기가 매우 큰 경우 (수백 MB 이상) 한꺼번에 메모리에 올리면 오히려 메모리 부족이나 GC 부하를 야기할 수 있기 때문에 큰 파일에는 적합하지 않습니다.
또한 입출력 중간 과정을 세밀하게 제어하기 어렵고, 모든 데이터를 메모리에 올리기 때문에 스트리밍 처리에는 부적합합니다.

`Files.write()`와 같은 고수준 API로 쓰기를 수행하면, JVM은 내부 버퍼에 데이터를 쓴 후 커널 모드로 전환하여 `write()` 시스템 콜을 호출합니다.
이때 데이터는 '유저 공간에서 커널 공간으로' 복사되어 커널의 페이지 캐시에 기록됩니다.
`Files.write()`가 리턴되면 커널 버퍼(페이지 캐시)에는 데이터가 있지만, 실제 디스크에는 아직 기록되지 않았을 수 있습니다.
OS는 지연 쓰기(lazy write)를 하기 때문에 일정 시간이나 버퍼 임계치에 도달하면 그때 비로소 디스크에 쓰기를 합니다.
따라서 데이터의 영속성을 바로 보장하려면 추가로 `FileChannel.force(true)` 또는 `Files.getFileAttributeView(...).flush()` 등을 호출해 커널 버퍼를 디스크에 강제 flush 해야 합니다.
가령 데이터베이스 등에서는 이 작업을 통해 ACID의 Durability(내구성)를 확보합니다.

### 대용량 또는 스트리밍 I/O: 버퍼드 스트림 (`BufferedReader/Writer`, `BufferedInputStream/OutputStream`)

자바의 전통적 I/O API (`java.io` 패키지)에서는 스트림(Stream) 개념을 사용해 파일 내용을 순차적으로 읽거나 씁니다.
특히 버퍼링(Buffering, 작은 조각 여러 개를 한 번에 모아 크게 주고받기)이 적용된 다양한 클래스를 통해 I/O 효율을 높입니다.
- `BufferedReader`
- `BufferedWriter`
- `BufferedInputStream`
- `BufferedOutputStream` 등

즉, 매 바이트마다 디스크에 접근하지 말고 일정 크기 메모리 버퍼에 담았다가 한 번에 읽고 쓰면 시스템 콜 횟수가 줄어들어 전체 성능이 개선될 수 있습니다.

예를 들어 파일에서 1바이트를 1만 번 읽는 코드를 가정합니다.
버퍼 없이 매번 읽으면 1만 번의 `read()` 시스템 호출과 1만 번의 사용자 및 커널 간의 전환(컨텍스트 스위치)이 발생하여 CPU와 I/O 자원을 낭비하고 속도를 크게 떨어뜨립니다.
버퍼링된 입력 스트림은 내부에 (예: 8KB 정도의) 배열 버퍼를 두고, `read()` 호출 시 한꺼번에 8KB를 미리 읽어둔 뒤 사용자에게는 1바이트씩 제공하는 방식입니다.
이렇게 하면 실제 파일 읽기 시스템 콜은 8KB당 1번만 일어나므로, 1만 바이트 읽기에 약 2번 정도의 시스템 콜만 발생할 수도 있습니다.

**Buffered 스트림 클래스 동작:**

`BufferedInputStream`/`BufferedReader`는 데코레이터 패턴으로 구현되어, 기존 `InputStream`/`Reader`에 버퍼링 기능을 추가합니다.
마찬가지로 `BufferedOutputStream`/`BufferedWriter`는 출력 시 버퍼를 채운 뒤 한꺼번에 `write()` 시스템 콜을 호출합니다.

예를 들어 `BufferedWriter`로 텍스트를 쓸 때, `write()` 메서드를 호출해도 실제 디스크에 바로 가지 않고 일단 JVM 힙 메모리의 버퍼에 저장됩니다.
버퍼가 가득 차거나 명시적으로 `flush()`가 호출되면 그때서야 커널에 데이터를 보내는 `write()` 시스템 콜이 발생하기 때문에 시스템 콜 횟수가 크게 감소할 수 있습니다.

대용량 텍스트 파일 처리 또는 스트리밍 데이터 처리에 자주 사용됩니다.
예를 들어 로그 파일을 지속적으로 쓰는 경우, 한 줄 쓸 때마다 디스크 I/O를 하면 비효율적이므로 `BufferedWriter`로 줄들을 모았다가 쓰는 것이 좋습니다.
반대로 소켓이나 파일에서 오는 연속적인 입력 스트림을 처리할 때도 `BufferedReader`로 감싸서 행 단위(`readLine()` 활용)로 읽으면 편리합니다.

버퍼를 사용한다는 점을 제외하면 기본적으로 OS와의 상호작용은 NIO와 비슷합니다.
데이터를 읽을 때는 버퍼가 비어있는 경우에만 실제 `read()` 시스템 콜이 발생하여 커널 페이지 캐시에서 데이터를 덩어리로 가져옵니다.
쓸 때는 버퍼가 찰 때까지 `write()` 시스템 콜을 유예하다가, 버퍼가 꽉 차면 그제서야 커널에 한꺼번에 전달합니다.

이 과정에서 '페이지 캐시'와 '스케줄링'은 OS가 관리합니다.
버퍼링된 쓰기의 경우, 사용자 공간에서 모은 데이터가 `write()` 시스템 콜로 커널에 보내지면 커널 공간의 페이지 캐시에 쌓입니다.
이후 OS는 자체 스케줄에 따라 디스크에 기록합니다.
만약 프로그램이 종료되거나 강제로 flush하지 않은 상태라면, 남은 데이터는 유실될 위험이 있으므로 항상 `bw.close()` 또는 `bw.flush()`를 호출해주는 것이 중요합니다.

기본 버퍼 크기는 클래스별로 8KB 등으로 정해져 있지만, 생성자에서 지정할 수도 있습니다.
버퍼가 너무 작으면 시스템 콜이 자주 일어나고, 너무 크면 메모리를 낭비하거나 불필요하게 많은 데이터를 선읽기/선쓰기하게 됩니다.
적절한 크기는 워크로드에 따라 경험적으로 결정합니다.

```java
Path inputPath = Paths.get("large_input.txt");
Path outputPath = Paths.get("large_output.txt");

try (
    // 파일 입력 스트림에 BufferedReader 장착 (디코레이터)
    BufferedReader reader = Files.newBufferedReader(inputPath, StandardCharsets.UTF_8);
    // 파일 출력 스트림에 BufferedWriter 장착
    BufferedWriter writer = Files.newBufferedWriter(outputPath, StandardCharsets.UTF_8);
) {
    String line;
    // 한 줄씩 읽어서 쓴다. BufferedReader가 내부적으로 많은 텍스트를 한번에 읽어둠.
    while ((line = reader.readLine()) != null) {
        writer.write(line);
        writer.newLine();  // 줄바꿈 추가
        // (여기서 writer.flush()를 호출하지 않으면 버퍼가 찰 때까지 실제 디스크에 안 씀)
    }
    // try-with-resources가 끝나면 자동으로 writer.close() -> 내부에서 flush() 수행
} catch (IOException e) {
    e.printStackTrace();
}
```

라인 단위 처리를 하지만, 실제로 일어나는 I/O는 훨씬 큰 덩어리 단위로 이뤄집니다.
`BufferedReader`는 파일에서 미리 많은 데이터를 읽어 `reader`의 버퍼에 저장해두고, `readLine()`은 메모리 버퍼에서 한 줄씩 뽑아갑니다.
만약 버퍼가 고갈되면 그때 추가로 `read()` 시스템 콜을 발생시켜 다음 데이터 덩어리를 가져옵니다.

반대로 `BufferedWriter`는 `write()` 호출 시마다 내용을 자체 버퍼에 쌓아두다가, 버퍼가 가득 차거나 명시적으로 `flush()`/`close()` 될 때 비로소 한꺼번에 파일로 출력합니다.

예를 들어 위 코드에서 `large_input.txt` 파일이 10MB이고 버퍼 크기가 8KB라고 가정합시다. 일반 `FileReader`/`FileWriter`로 한 문자씩 처리하면 수백만 번의 I/O 호출이 필요할 수 있지만, `BufferedReader/Writer`를 쓰면 10MB/8KB ≈ 약 1280번 정도의 시스템 콜로 모든 작업을 끝낼 수 있습니다. 시스템 콜 횟수가 수백만→천여 개로 급감하므로 **속도 향상이 현격**합니다.

`writer.close()`는 내부적으로 `flush()`를 호출하여 버퍼에 남은 데이터를 모두 출력합니다.

1. 저 스페이스 버퍼: 예컨대 8KB 버퍼가 가득 차면,
2. 시스템 콜 발생: JVM이 `write(fd, buffer, 8192)` 시스템 콜을 호출하여 커널 모드로 전환,
3. 커널 공간 복사: 8192바이트 데이터가 커널의 페이지 캐시로 복사되고 이 순간 호출은 성공 반환. (애플리케이션은 이 시점을 파일 기록 완료로 인식하지만 실제로는 디스크엔 아직 전달되지 않음)
4. 디스크 기록: OS가 나중에 페이지 캐시의 dirty 페이지들을 디스크로 쓰기.

결국 `InputStream`이 `AutoClosable` 인터페이스를 구현하므로 수동으로 `flush()`하지 않아도`try-with-resources` 마지막에 자동으로 처리됩니다.
다만, 장기간 데이터가 쌓이는 실시간 로그의 경우 주기적으로 `flush()`를 호출하여 실시간성을 확보하기도 합니다.

- 장점:
    - 기존 Java I/O와의 호환성이 높고 (`Reader`/`Writer`를 그대로 래핑) 구현이 간단함
    - 메모리-디스크 왕복 횟수를 줄여서 대용량 처리 성능이 뛰어남
    - 라인 단위, 바이트 단위 등 다양한 편의 메서드를 제공하기 때문에 스트리밍 처리에 용이함
- 단점:
    - 버퍼 크기를 잘못 설정하면 효율이 떨어질 수 있고, 지나치게 큰 버퍼는 메모리 낭비
    - 또한 파일 전체를 한 번에 다루는 고수준 API(`Files.readAllBytes` 등)에 비하면 코드가 장황할 수 있음

단, 버퍼를 flush하지 않고 프로그램이 비정상 종료되면 데이터 유실 위험이 있다는 것을 주의해야 합니다.
버퍼링은 사용자-커널 전환 비용을 줄이지만 데이터 완전성은 OS의 flush에 맡기므로, 중요한 데이터는 flush, 파일닫기 또는 `fsync`를 호출해야 합니다.

### 대용량 파일의 효율적 복사 및 전송: `FileChannel.transferTo()/transferFrom()` (Zero-Copy)

**철학 및 배경:** 수 GB 이상의 *매우 큰 파일*을 다룰 때는, 기존 방식으로 일일이 읽어서 쓰는 것조차 CPU와 메모리에 큰 부담이 됩니다. **Zero-Copy**란 데이터 복사 과정을 최소화하여 효율을 극대화하는 기술로, Unix 계열 OS의 `sendfile(2)` 시스템 콜 등이 대표적입니다. Java NIO의 `FileChannel.transferTo()`와 `transferFrom()` 메서드는 이런 \*\*커널 지원 "Zero-Copy"\*\*를 이용하여 두 채널 간 데이터를 직접 전송합니다. 이를 통해 파일을 네트워크로 보내거나 파일 간 복사할 때 **사용자 공간으로 데이터가 올라오지 않고** OS가 커널 공간에서 바로 처리하게 합니다.

**해결하는 문제:** 전통적으로 Java에서 파일을 소켓으로 보내려면 (예: 파일 다운로드 구현) 다음과 같은 흐름이었습니다:

1. 파일에서 `read()` → 커널에서 사용자 공간으로 데이터 복사
2. 소켓에 `write()` → 사용자 공간에서 다시 커널 네트워크 버퍼로 데이터 복사
3. 커널 네트워크 버퍼 → NIC 전송

이 과정에서 **두 번의 데이터 복사와 컨텍스트 스위치**가 발생합니다. Zero-copy 기법을 쓰면 **디스크 -> 커널 페이지캐시 -> 네트워크 버퍼**로 바로 복사가 일어나고, 사용자 공간엔 데이터가 내려오지 않으므로 CPU 메모리 복사 비용과 컨텍스트 스위칭이 크게 줄어듭니다. 즉, **큰 파일 전송에서의 병목** (CPU 복사 비용, 메모리 대역폭 사용)을 완화합니다.

`FileChannel.transferTo()`는 바로 이러한 zero-copy 전송을 Java에서 가능케 하는 메서드입니다. 실제 구현에서, **전송 대상이 SocketChannel인 경우** Linux라면 `sendfile` 시스템 콜을 내부적으로 호출하여 **커널 레벨에서 파일 내용을 직접 네트워크로 보냅니다**. 대상이 또 다른 FileChannel일 때도, 운영체제가 지원한다면 커널이 파일->파일 복사를 최적화합니다 (일부 OS에서는 file->file일 때도 `sendfile` 또는 유사 기능 사용). 결과적으로 사용자 코드에서 루프 돌며 버퍼 복사하는 것보다 훨씬 빠르게 처리됩니다.

**내부 동작 (저수준):** `transferTo(position, count, targetChannel)`을 호출하면 JVM은 OS 시스템 콜 `sendfile` 또는 유사 기능을 사용해 **파일 디스크립터 간 데이터 이동**을 지시합니다. 예컨대 Linux에서는 다음과 같이 동작합니다:

- 커널이 파일(Page Cache)에 있는 데이터를 직접 **네트워크 소켓 버퍼**로 복사 (DMA 엔진 이용 가능).
- 데이터가 **사용자 공간으로 한 번도 올라오지 않고**, 커널 내에서 처리되므로 *user-kernel* 전환 및 복사 비용 0에 수렴 (그래서 **Zero-Copy**라 부름).

단, "zero-copy"라 해도 엄밀히는 \*\*커널 내에서의 데이터 복사(copy)\*\*는 발생합니다 (디스크에서 읽어 온 페이지 캐시 -> 소켓 송신 버퍼). 다만 이 복사는 커널 영역 내에서 일어나기 때문에 사용자 영역 ↔ 커널 영역 간 복사보다는 훨씬 효율적입니다. 또한 최신 DMA 기술로 커널 복사마저 줄이는 경우도 있습니다. 핵심은 **사용자 프로세스가 데이터를 한 바이트도 읽지 않고 OS에게 "저 파일을 저쪽으로 보내"라고 위임**한다는 점입니다.

**사용 예시:** 아래 코드는 `FileChannel.transferTo()`를 이용해 대용량 파일을 복사하는 예입니다. 먼저 전통적인 방식(버퍼를 이용한 수동 복사)과 비교하고, 이어서 `transferTo` 사용을 보여줍니다.

```java
// 기존 방식: 직접 버퍼를 할당하여 파일을 복사
try (FileChannel src = FileChannel.open(Paths.get("source.bin"), StandardOpenOption.READ);
     FileChannel dest = FileChannel.open(Paths.get("dest.bin"), StandardOpenOption.CREATE_NEW, StandardOpenOption.WRITE)) {

    ByteBuffer buffer = ByteBuffer.allocate(1024 * 1024); // 1MB 버퍼
    while (src.read(buffer) != -1) {  // 소스에서 버퍼로 읽기
        buffer.flip();                // 읽은 데이터를 버퍼에서 꺼낼 준비
        dest.write(buffer);           // 버퍼에서 대상 파일로 쓰기
        buffer.clear();               // 버퍼 비우기
    }
}

// 개선된 방식: transferTo 이용 (Zero-Copy)
try (FileChannel src = FileChannel.open(Paths.get("source.bin"), StandardOpenOption.READ);
     FileChannel dest = FileChannel.open(Paths.get("dest2.bin"), StandardOpenOption.CREATE_NEW, StandardOpenOption.WRITE)) {

    long fileSize = src.size();
    long transferred = src.transferTo(0, fileSize, dest);
    System.out.println("전송 바이트: " + transferred + " / 총 파일 크기: " + fileSize);
}
```

첫 번째 (전통적인) 방식은 1MB 메모리 버퍼를 만들어 루프를 돌면서 파일 내용을 복사했습니다. 이 경우 **사용자 공간에서 1MB씩 데이터를 읽어 썼기 때문에** 100GB 파일이라면 약 100번 \* 1000 = 100,000번의 read/write 시스템 콜과 동일한 양의 메모리 복사가 발생할 수 있습니다 (물론 실제론 read 100k + write 100k = 200k syscalls). 반면 두 번째 `transferTo` 방식은 **딱 한 번의 시스템 콜**로 OS에게 100GB 전송을 맡깁니다. OS는 내부적으로 고속 복사를 수행하며, 자바 애플리케이션은 CPU를 거의 사용하지 않고 전송이 완료되길 기다리기만 하면 됩니다.

**효과:** 실무에서 `transferTo`/`transferFrom`은 파일 서버나 미디어 스트리밍 서버에서 애용됩니다. 예컨대 **Apache Kafka**는 디스크의 로그 세그먼트를 네트워크로 전송할 때 zero-copy를 활용하여 **페이지 캐시 + zero-copy 조합으로 소비자에게 매우 빠르게 데이터를 전송**합니다. 성능상 수십 퍼센트 이상의 I/O 향상을 볼 수 있고, CPU 사용률은 크게 낮아집니다.

**주의점:** 모든 경우에 zero-copy가 가능한 것은 아닙니다. OS에 따라 제약이 있는데:

- **플랫폼 지원:** Linux/Unix 계열은 `transferTo`가 소켓에 대해 `sendfile`을 사용하지만, Windows는 Java에서 소켓에 대해 overlapped I/O 등을 내부적으로 사용할 수 있으나 과거에는 제약이 있었습니다. (Java 17에서는 대부분 플랫폼에서 투명하게 동작하지만, 만약 zero-copy를 사용할 수 없으면 JVM이 알아서 일반 방식으로 fallback합니다.)
- **파일->소켓 vs 소켓->파일:** 일반적으로 **파일→소켓** 방향의 전송에 최적화되어 있습니다. 소켓에서 파일로 받는 경우는 효과가 제한적이라, 이때는 전통적 방법과 성능 차이가 크지 않을 수 있습니다.
- **전송 크기 제한:** OS 마다 한 번의 `sendfile`로 보낼 수 있는 최대 바이트 수 제한이 있을 수 있습니다. 매우 큰 파일은 루프에서 `transferTo`를 여러 번 호출해야 할 수도 있습니다. (`transferTo`는 반환값으로 보낸 바이트 수를 리턴하므로 루프로 처리 가능)
- **메모리 맵 충돌:** `MappedByteBuffer` 등과 동시에 사용할 경우, 혹은 파일 잠금(lock) 상태에서는 사용에 제약이 있을 수 있습니다.

**장점:** 대용량 파일 복사/전송에 있어서 **최상의 성능**을 제공합니다. CPU 개입이 적고, 데이터 복사 횟수가 줄어들어 **메모리 대역폭**을 아낄 수 있습니다. 결과적으로 동일 하드웨어에서 더 많은 스루풋을 낼 수 있습니다. 특히 파일을 네트워크로 보낼 때 효과적입니다.

**단점:** 코드가 비교적 low-level이고, **전송 실패나 부분 전송에 대한 처리를 신경써야** 합니다. (예를 들어 네트워크 소켓의 경우 송신 버퍼 부족 등으로 한 번에 다 못 보낼 수 있으므로 루프 돌며 전송해야 함.) 또한 Zero-Copy라고 해도 커널 내 버퍼 복사는 있기 때문에, 아주 작은 파일(예: 몇 KB)에서는 오히려 이점이 미미하거나 함수 호출 오버헤드 때문에 느릴 수 있습니다. 일반적으로 **몇 MB 이상의 전송부터 효과가 뚜렷**해집니다.

---

## 4. 랜덤 액세스 및 초고속 파일 I/O – 메모리 맵핑 (`MappedByteBuffer`)

**철학 및 배경:** \*\*메모리 맵 파일 (Memory-Mapped File)\*\*은 파일의 내용을 가상의 메모리에 맵핑하여 **마치 메모리를 다루듯** 파일을 읽고 쓰는 기법입니다. Java에서는 `FileChannel.map()` 메서드를 통해 `MappedByteBuffer`를 얻어 사용할 수 있습니다. 이 방식은 1970\~80년대부터 OS가 지원해온 기법으로, Java에는 NIO (JDK 1.4)부터 도입되었습니다. 철학은 "*디스크 I/O를 일일이 읽지 말고, OS에게 파일을 **메모리 페이지**로 취급시켜 달라고 하자*"는 것입니다. 그러면 읽기/쓰기는 일반 메모리 접근처럼 수행되고, **필요한 부분만 페이지 단위로 로드**되며, 수정된 페이지는 나중에 디스크에 반영(flush)됩니다.

**해결하는 문제:**

- ***임의 접근(Random Access)**:* 매우 큰 파일에서 일부 바이트만 자주 읽거나 쓰는 경우, 기존 스트림 방식으로는 원하는 위치로 `seek`한 후 `read`/`write`하고 또 `seek`해야 합니다. 반면 메모리맵은 파일 내용이 메모리에 올라온 것처럼 다루기 때문에 `buffer.get(index)`로 즉시 특정 위치 접근이 가능합니다.
- **성능:** 메모리 맵핑은 OS의 가상메모리 시스템을 활용하므로, **파일 내용을 프로세스 메모리에 직접 매핑**합니다. 읽기/쓰기 호출이 없고 대신 **페이지 폴트**에 의해 자동으로 필요한 조각만 디스크에서 불러옵니다. 또한 데이터를 수정하면 해당 메모리 페이지가 dirty로 표시되었다가, OS가 적절한 시점에 디스크에 쓰기 때문에, 커널이 **효율적인 캐싱과 쓰기 병합**을 해줄 수 있습니다. 이러한 특성 때문에 메모리맵은 **아주 빠른 I/O**를 요구하는 애플리케이션 (데이터베이스, 검색 엔진, 금융 트레이딩 시스템 등)에서 자주 쓰입니다. 실제로 "Java에서 가능한 가장 빠른 I/O" 중 하나로 꼽히며, 마이크로초 단위의 저지연이 필요한 경우 선택됩니다.

**내부 동작:** `FileChannel.map()`을 호출하면 운영체제의 `mmap` 시스템 콜이 수행됩니다. 이때 **파일의 특정 영역을 프로세스 가상 메모리 영역에 매핑**하게 되며, 해당 범위에 접근할 때 실제 디스크에서 데이터가 페이지 단위로 로드됩니다. 이를 \*\*페이지 폴트(Page Fault)\*\*라고 합니다. 예를 들어 1GB 파일을 맵핑하더라도 실제로는 아직 아무 것도 읽지 않고 있다가, 맵핑된 버퍼의 첫 바이트를 읽으려고 하면 그 시점에 OS가 해당 4KB 페이지를 디스크에서 읽어와 메모리에 올립니다. 이후 그 페이지 내에서는 메모리 읽기 속도로 접근 가능합니다. 만약 순차적으로 모든 페이지를 접근하면 결과적으로 파일 전체를 읽게 되지만, 중간에 일부만 건드리면 *건드린 부분만* 메모리에 올라옵니다.

쓰기의 경우도 유사합니다. `MappedByteBuffer.put()`으로 데이터를 쓰면 즉시 디스크에 가지 않고 **메모리 페이지에만 수정**되고, 해당 페이지는 dirty 상태가 됩니다. OS는 메모리 상황을 보아가며 또는 맵이 언맵되거나 `MappedByteBuffer.force()`가 호출될 때 해당 페이지를 디스크에 씁니다. 이는 앞서 설명한 OS 페이지 캐시의 작동과 본질적으로 동일하지만, 메모리맵은 **애플리케이션이 직접 페이지 캐시에 접근**하는 형태로 볼 수 있습니다. (일반 파일 I/O도 내부적으로는 페이지 캐시를 사용하지만, 그 과정이 사용자에게 투명함)

**메모리 맵의 장점:**

- 매우 빠른 데이터 접근: 메모리 읽기/쓰기에 준하는 속도입니다. 별도의 read/write 호출이 없고 CPU의 메모리 관리 유닛(MMU)이 디스크→메모리 로드를 관리하므로 오버헤드가 적습니다.
- 커널의 **지능형 캐싱** 활용: OS가 메모리 사용량을 보며 페이지를 캐시하거나 버립니다. 한번 읽은 페이지는 다시 접근 시 디스크 I/O 없이 빠르게 사용 가능합니다.
- 큰 파일 처리: 수백 MB\~수 GB 파일도 작은 메모리 사용으로 핸들링 가능. 예컨데 10GB 파일에서 100MB만 실제 접근하면, 그 부분만 메모리에 올라오고 나머지 9.9GB는 디스크에 그대로.
- 프로세스 간 공유: 메모리 맵은 **여러 프로세스 간 메모리 공유**도 가능합니다 (같은 파일을 맵핑하면 같은 물리 메모리를 바라봄). Java에서는 여러 프로세스가 동시에 mmap 할 일은 드물지만, C나 OS 레벨에서는 IPC로 응용합니다.

**사용 예시:** 다음 코드에서는 메모리 맵핑을 사용하여 파일을 읽고 쓰는 예를 보여줍니다. 100MB 크기의 파일을 메모리맵으로 열고, 일정 바이트를 수정한 뒤 읽어오는 시나리오입니다.

```java
int FILE_SIZE = 100 * 1024 * 1024; // 100MB
Path path = Paths.get("data.bin");

// 100MB 크기의 빈 파일 생성 (플랫폼에 따라서는 파일이 실제로 이 크기로 확장됨)
Files.write(path, new byte[FILE_SIZE], StandardOpenOption.CREATE, StandardOpenOption.WRITE);

try (RandomAccessFile raf = new RandomAccessFile(path.toFile(), "rw");
     FileChannel channel = raf.getChannel()) {

    // 파일을 메모리맵: 전체 범위를 READ_WRITE 모드로 매핑
    MappedByteBuffer mbuf = channel.map(FileChannel.MapMode.READ_WRITE, 0, FILE_SIZE);
    // 이제 mbuf의 0~99,999,999 인덱스가 파일 내용에 대응됩니다.

    // 1. 맵핑된 버퍼에 일부 데이터 써보기
    for (int i = 0; i < 1024; i++) {
        mbuf.put(i, (byte) 0xFF);  // 파일 처음 1024바이트를 0xFF로 설정
    }
    // ※ put() 호출 시 즉각 디스크에 쓰지 않고 메모리에만 반영합니다.

    // 2. 맵핑된 버퍼에서 특정 위치 읽기 (random access)
    int pos = 500;
    byte value = mbuf.get(pos);
    System.out.println(String.format("Position %d value: 0x%02X", pos, value));

    // 3. 변경 내용 강제 flush (옵션): 메모리의 변경분을 디스크에 확실히 반영
    mbuf.force();  // 호출하지 않아도 일정 시간 뒤 OS가 자동 flush하지만, 안전을 위해 호출 가능
} catch (IOException e) {
    e.printStackTrace();
}
```

위 예제에서 `channel.map(...)` 호출로 **100MB 파일을 프로세스 메모리에 맵핑**했습니다. 아직 아무 데이터 접근을 안 했으므로 실질적인 디스크 읽기는 일어나지 않았습니다. 그 다음 `mbuf.put(i, (byte)0xFF)`를 1024번 호출하여 파일의 0\~1023 바이트를 변경했습니다. 이때도 **디스크 쓰기 시스템 콜은 발생하지 않았습니다**. 단지 해당 페이지(예를 들어 첫 4KB 페이지와 그 다음 페이지 등)가 메모리에 올라오고 변경되었을 뿐입니다. (처음 `put`할 때 해당 페이지가 없으면 페이지 폴트로 디스크에서 로드 후 메모리 수정) 이렇게 수정된 페이지들은 dirty로 표시되어 OS가 배경에서 디스크에 쓸 준비를 합니다. 마지막으로 `mbuf.force()`를 호출하면 즉시 해당 페이지들을 디스크에 기록하는 `msync` 시스템 콜과 유사한 동작을 합니다. 이로써 변경 내용이 물리 디스크에 안전하게 저장됩니다.

**성능 측면:** 일반적인 파일 I/O로 100MB를 순차 쓰기하면 CPU가 100MB 데이터를 일일이 복사하지만, 메모리맵은 위 코드처럼 바이트 단위 랜덤 접근도 빠르게 수행됩니다. 특히, **랜덤 읽기 패턴의 경우** 메모리맵이 압도적으로 빠릅니다. 왜냐하면 임의 위치 접근 시 커널이 해당 페이지만 로드하면 되므로 시킹(seeking) 비용이 최소화되고, 사용자가 커널에 여러 번 read 요청을 보낼 필요 없이 하드웨어가 페이지를 로드해주기 때문입니다.

다만, **순차 접근**만 하는 용도라면 큰 이점이 없을 수도 있습니다. 어느 정도 이상 파일 크기에서는 오히려 전통적인 buffered I/O와 비슷하거나, 경우에 따라 **버퍼드 I/O가 더 빠른 경우도** 있습니다. (Buffered I/O는 read ahead 최적화 등이 있고, GC에 덜 영향을 미치며, 구현 간 오버헤드 차이 등이 원인입니다.) 따라서 메모리맵은 특히 **랜덤 액세스 빈도가 높거나 다중 쓰레드가 동시에 큰 파일의 부분들을 접근**하는 상황, 그리고 **메모리 공유**가 필요한 경우에 적합합니다.

**장점:**

- *빠른 속도:* 앞서 언급했듯 **가장 빠른 파일 I/O** 기법 중 하나입니다. 커널의 가상메모리 관리에 맡기므로 read/write 함수 호출 및 복사 오버헤드가 없습니다. 대용량 파일도 효율적으로 다룰 수 있어 고성능 시스템에서 애용됩니다.
- *임의 접근 효율:* 배열 인덱스 접근처럼 동작하므로, 파일의 특정 오프셋을 읽기 위해 일련의 스트림 연산을 하지 않아도 됩니다. 이는 데이터베이스 인덱스, 거대한 파일 내 검색 등의 시나리오에서 유용합니다 (DBMS도 내부적으로 메모리맵을 활용하는 경우가 많습니다).
- *시스템 메모리 활용:* JVM 힙 대신 **OS의 페이지 캐시**를 사용하므로, 힙 메모리 부담이 없습니다. 맵핑한 파일 크기가 JVM 힙보다 커도 문제가 없으며, OS가 메모리 상황에 따라 페이지를 스왑아웃하거나 캐시를 조절해주므로 **대용량 파일도 작은 메모리 Footprint로 처리**할 수 있습니다.
- *프로세스간 공유:* (Java 내부에서는 잘 쓰이지 않지만) 동일 파일을 여러 프로세스가 맵핑하면 물리 메모리를 공유하므로 IPC에 활용 가능하고, 한 프로세스의 변경을 다른 쪽이 곧바로 볼 수도 있습니다.

**단점 및 주의:**

- *메모리 사용 주의:* 맵핑은 가상 메모리를 점유합니다. 32-bit JVM에서는 주소 공간 제한으로 **최대 2GB까지** 한 영역을 맵핑할 수 있습니다. 64-bit에서는 실메모리만 충분하면 크기에 제약이 크진 않지만, **연속된 가상 메모리 확보**가 필요하므로 아주 조각난 상황에서는 실패할 수 있습니다. 또한 한번 맵핑한 파일은 GC로 바로 해제되지 않고, **JVM 프로세스가 종료되거나** `Cleaner`가 실행되어야 해제됩니다. 그래서 맵핑한 버퍼를 자주 만들고 버리면 OutOfMemoryError (map 영역 부족)가 날 수 있습니다. 적절히 `MappedByteBuffer.force()`와 `System.gc()`를 통해 언맵을 유도해야 하는 번거로움이 있습니다.
- *데이터 일관성:* 메모리맵으로 쓴 내용은 OS가 알아서 flush하지만, 프로세스 크래시나 OS 크래시 시 flush되지 못한 데이터는 유실될 수 있습니다. 데이터베이스처럼 **강제 동기화**가 필요하다면 `mbuf.force()`를 호출해줘야 합니다. 이는 일반 I/O에서 `fileChannel.force()`나 `fsync`를 호출하는 것과 같은 맥락입니다.
- *페이지 수명:* 맵핑된 페이지들이 다른 일반 메모리와 경쟁하므로, 매우 큰 파일을 맵핑하고 거의 사용하지 않으면 쓸데없이 메모리만 차지할 수 있습니다. (OS가 곧 스왑아웃하긴 하지만).
- *GC에 영향:* 비록 Java 힙을 쓰지 않아도, `MappedByteBuffer` 객체 자체는 Direct ByteBuffer로서 **JVM의 **Direct 메모리** 영역을 사용**합니다. 이 영역은 `-XX:MaxDirectMemorySize` 제한을 받으며, GC가 이 용량을 추적합니다. 너무 많은 메모리맵을 동시에 열면 Direct memory 부족이나 GC 부담 증가로 성능이 안 좋아질 수 있습니다. 실제로 메모리맵 남용 시 GC가 빈번히 일어나 성능저하 사례도 있습니다.
- *쓰레드 안전성:* `MappedByteBuffer` 자체는 thread-safe하지 않습니다. 동시에 같은 버퍼를 여러 스레드가 접근하면 별도 동기화가 필요합니다. 다만 일반적으로 하나의 맵을 한 스레드에서 쓰거나, 구간을 나눠 쓰면 큰 문제는 없습니다.

결론적으로, **MappedByteBuffer는 "한번 맵핑해 놓고 자유자재로 파일 데이터를 메모리처럼 활용"할 수 있게 해주는 강력한 도구**입니다. 적절히 활용하면 성능을 극한까지 끌어올릴 수 있지만, JVM의 관리 영역 밖을 다루는 만큼 주의를 요합니다. 주요 용도는 데이터베이스(예: LMDB, ChronicleMap), 대용량 파일 편집기, OS와 밀접히 통신하는 네이티브 연동 등에 있습니다.

---

## 5. 논블로킹/비동기 파일 I/O – `AsynchronousFileChannel` + `CompletableFuture` (NIO2의 AIO)

**철학 및 배경:** 전통적인 Java I/O (그리고 위의 FileChannel 등) API는 대부분 **블로킹 I/O**입니다. 즉, `read()`나 `write()` 메서드를 호출하면 작업이 완료될 때까지 현재 쓰레드가 멈춘 채 기다립니다. 그러나 서버나 백그라운드 작업에서는 **파일 I/O를 하는 동안 쓰레드가 놀고 있기보다는, 다른 일에 쓰이거나 완전히 쓰레드가 필요없게 만들고 싶을 때**가 있습니다. 이를 위해 Java 7의 NIO.2는 **비동기 파일 채널**을 도입했습니다. `AsynchronousFileChannel` 클래스가 그것으로, 이름 그대로 **비동기(async) 방식**으로 파일 읽기/쓰기를 수행합니다.

**해결하는 문제:**

- **다중 병렬 I/O 처리:** 예를 들어 여러 개의 큰 파일을 동시에 읽어야 한다면, 일반적으로는 스레드를 여러 개 만들어 각각 블로킹 I/O를 수행합니다. 그러나 스레드가 너무 많이 늘어나면 컨텍스트 스위치 비용이 커지고 관리가 어려워집니다. `AsynchronousFileChannel`은 OS나 JVM의 스레드풀을 활용하여 **백그라운드에서 I/O를 수행**하고, 완료되면 콜백(CompletionHandler)이나 Future로 결과를 알려줍니다. 이러면 개발자는 수십\~수백 I/O를 한두 개 쓰레드로도 돌릴 수 있고, 특히 I/O 대기 시간이 긴 HDD 작업에서 **스레드가 놀고있는 시간을 제거**할 수 있습니다.
- **메인 쓰레드 비블로킹:** GUI 애플리케이션이나 이벤트 루프 기반 서버에서, 큰 파일을 읽느라 메인 쓰레드가 멈추면 안 됩니다. AIO를 쓰면 **메인 쓰레드는 작업 요청만 던져놓고 곧바로 다음 일**을 할 수 있습니다. I/O 완료는 나중에 이벤트로 받아 처리하면 되므로, 애플리케이션의 응답성을 높일 수 있습니다.

**내부 동작:** `AsynchronousFileChannel`을 만들 때 **전담 스레드 풀**을 지정할 수 있습니다. 지정하지 않으면 자바가 내부적으로 기본 스레드풀 (CachedThreadPool 유사한)을 사용합니다. 자, 예를 들어 `asyncChannel.read(buffer, position, attachment, completionHandler)`를 호출하면:

1. **요청 큐잉:** 호출한 쓰레드는 즉시 반환받습니다 (`read()` 메서드는 `Future`를 리턴하거나 CompletionHandler를 통해 처리). 내부적으로, 파일 읽기 작업이 별도의 작업(task)으로 스레드풀에 제출됩니다.
2. **백그라운드 I/O:** 스레드풀의 스레드 중 하나가 실제 `pread()` (포지션 지정 비동기 read) 시스템 콜을 호출하거나, OS가 직접 지원하면 OS 이벤트를 등록합니다.

   - 윈도우의 경우 OS 레벨로 Overlapped I/O (IOCP) 지원이 있어, 스레드 없이 OS가 완료시까지 처리해줄 수 있습니다.
   - 리눅스의 경우 전통적으로는 파일의 비동기 I/O 지원이 부족하여, 사실상 **스레드풀이 blocking I/O를 대신 호출**하는 식으로 구현되었습니다. (리눅스 AIO는 O\_DIRECT 등 제한이 있어, Java는 주로 자체 스레드풀로 구현.)
3. **I/O 완료 시:** 읽기 작업이 완료되면, **CompletionHandler 콜백**이 호출되거나, Future의 상태가 완료로 바뀝니다. CompletionHandler는 `AsynchronousFileChannel` 생성 시 지정한 스레드풀의 스레드 중 하나가 수행합니다. 즉, 사용자 콜백은 별도 스레드에서 실행되므로, 거기서 무거운 로직을 처리해도 메인 흐름에 영향을 안 줍니다 (단, 주의: 기본 스레드풀은 다소 제한이 있을 수 있으니 너무 무거운 작업은 별도 처리가 필요).

Java 문서에 따르면, 하나의 AsynchronousFileChannel에서 **다중 읽기/쓰기 작업을 동시에 outstanding**하게 발행할 수 있습니다. 순서는 보장되지 않지만, 내부적으로는 OS나 스레드풀에서 여러 I/O를 병렬로 처리합니다. 예컨대 4개 파일에서 4군데를 동시에 읽어오라고 하면 4개 작업이 한꺼번에 진행될 수 있습니다 (HDD라면 실제로는 시킹으로 직렬화되겠지만, SSD나 OS 캐시가 있다면 병렬성이 더 유효할 수 있습니다).

**사용 예시:** 아래 코드는 `AsynchronousFileChannel`을 이용해 비동기 읽기를 수행하는 두 가지 방법 (Future와 CompletionHandler)을 보여줍니다.

```java
Path filePath = Paths.get("bigfile.log");
ByteBuffer buffer = ByteBuffer.allocate(1024 * 1024); // 1MB 버퍼

// (1) Future 방식 비동기 읽기
try (AsynchronousFileChannel asyncChannel = AsynchronousFileChannel.open(filePath, StandardOpenOption.READ)) {
    Future<Integer> result = asyncChannel.read(buffer, 0);  // non-blocking 호출
    // ... (다른 작업 수행 가능) ...
    // 필요할 때 결과 받기 (이때까지 완료되지 않았으면 여기서 블록됨)
    int bytesRead = result.get();
    System.out.println("읽은 바이트 수 (Future): " + bytesRead);
} catch (Exception e) {
    e.printStackTrace();
}

// (2) CompletionHandler 콜백 방식 비동기 읽기
try (AsynchronousFileChannel asyncChannel = AsynchronousFileChannel.open(filePath, StandardOpenOption.READ)) {
    asyncChannel.read(buffer, 0, "작업1", new CompletionHandler<Integer, String>() {
        @Override
        public void completed(Integer result, String attachment) {
            System.out.println(attachment + " 완료! 읽은 바이트: " + result);
        }
        @Override
        public void failed(Throwable exc, String attachment) {
            System.err.println(attachment + " 실패: " + exc.getMessage());
        }
    });
    // read 호출 직후 바로 반환, 아래 코드 즉시 실행됨
    System.out.println("메인 스레드는 블로킹 없이 계속 실행됨...");
    // ... (다른 작업 수행, I/O 완료는 별도 스레드에서 처리됨) ...
    Thread.sleep(2000); // 예제에서는 프로그램이 종료되지 않도록 잠시 대기
} catch (Exception e) {
    e.printStackTrace();
}
```

위 코드를 보면, `asyncChannel.read(buffer, position)` 호출 후 **바로 제어권이 돌아오기 때문에** 이후 코드를 곧바로 실행할 수 있습니다. (예제에서는 그 사실을 출력으로 확인하려고 `System.out.println`과 `Thread.sleep`을 넣었습니다.) I/O 완료는 `(1)`에서는 `Future.get()`을 호출하는 시점까지 지연될 수 있고, `(2)`에서는 `completed` 콜백으로 통지됩니다.

**AsynchronousFileChannel의 구현상 한계:** 언급했듯이, 리눅스에서 파일의 진짜 async I/O를 지원하지 않기 때문에 Java는 **뒤에서 쓰레드로 작업을 처리**하는 경우가 많습니다. 즉, 사용자 스레드를 덜 쓰는 대신 JVM 내부 스레드를 쓰는 격입니다. 그러므로 CPU 이용률 측면에서는 블로킹과 크게 다르지 않을 수 있습니다. 하지만 논블로킹의 이점은 **프로그래밍 모델**에 있습니다. 많은 수의 동시 파일 작업을 쉽게 관리하고, I/O 대기 시간 동안 애플리케이션의 다른 처리를 진행할 수 있게 해줍니다.

**장점:**

- 여러 파일 I/O를 동시에 처리하기 쉽습니다. **스레드 수를 최소화**하면서 다중 I/O 병행성을 확보할 수 있습니다. 예컨대, 스레드 하나로도 수십 파일을 동시에 읽도록 쭉 요청만 던져놓고 Future/콜백으로 관리 가능.
- 이벤트 지향 프로그래밍과 잘 맞습니다. 네트워크의 NIO `Selector`처럼, 파일도 CompletionHandler를 쓰면 **I/O 완료 이벤트**로 다룰 수 있습니다.
- UI/메인 쓰레드가 블로킹되지 않으므로 애플리케이션 응답성 향상.

**단점:**

- 코드가 복잡해집니다. Future나 콜백을 사용해야 하므로 동기식 코드보다 이해하기 어려울 수 있습니다. (Java 8 이후 `CompletableFuture`와 람다 등을 사용하면 조금 깔끔해졌지만 여전히 난이도가 있습니다.)
- **OS 지원 제약:** OS 차원에서 async I/O가 완벽하지 않아, 특히 디스크가 느린 경우 성능이 극적으로 향상되지 않을 수 있습니다. 예를 들어 HDD에서는 여러 I/O를 동시에 해봐야 디스크가 한 번에 하나씩 밖에 못하기 때문에, 논블로킹이더라도 내부적으로 순차 처리됩니다. SSD나 NVMe 환경에서는 병렬 처리 이점이 조금 있습니다만, 그럴 땐 차라리 스레드 여러 개가 동시 블로킹 읽기 해도 비슷한 효과일 수 있습니다.
- **자원 관리:** 백그라운드 스레드풀이 있으므로, 너무 많은 async 채널/작업을 내면 스레드풀이 과부하 걸릴 수 있습니다. 기본 스레드풀은 CPU 코어 수 등으로 제한될 수 있으며, 필요시 `AsynchronousChannelGroup`을 만들어 풀 크기를 조절해야 합니다.
- **순서 보장 어려움:** 여러 작업 결과의 순서를 보장받으려면 별도 동기화가 필요합니다.

정리하면, **AsynchronousFileChannel은 특정 상황에서 유용**합니다. 대용량 파일을 비동기적으로 복사한다든가, 여러 파일을 동시에 로딩해야 할 때 메인 로직은 진행하면서 I/O 결과를 나중에 취합하는 패턴 등에 좋습니다. 다만 일반적인 파일 처리에서는 오히려 과하게 복잡할 수 있어 많이 쓰이진 않습니다. (오히려 네트워크 소켓의 async는 더 자주 쓰입니다.) \*\*JDK 19부터는 Virtual Thread (Project Loom)\*\*이 도입되어, 수천 개의 파일 읽기도 가상쓰레드로 막 던져버리고 블로킹 해도 효율적으로 처리되기에, 미래에는 AsynchronousFileChannel의 존재감이 줄어들 가능성도 있습니다.

---

## 6. 파일 변경 감시 – `WatchService` (NIO2)

**철학 및 배경:** 운영체제는 파일시스템의 변경을 애플리케이션에 알려주는 기능(inotify, ReadDirectoryChangesW 등)을 제공합니다. Java 7의 NIO.2에 포함된 `WatchService` API는 이러한 OS 기능을 활용하여 **디렉터리 안의 파일 변화 이벤트를 모니터링**할 수 있게 합니다. 철학은 "*폴링(polling)으로 변화 체크하지 말고, 이벤트를 구독(subscribe)하라*"는 것입니다. 개발자는 특정 디렉터리에 "Watcher"를 등록해두고, 생성/삭제/수정 이벤트가 발생하면 통지받을 수 있습니다.

**해결하는 문제:**

- **폴링의 비효율성:** 어떤 설정 파일이 바뀌면 프로그램이 그것을 즉시 로드하도록 만들고 싶다고 합시다. 전통적으로는 5초마다 파일의 마지막 수정시각을 체크(polling)하거나, 사용자에게 "다시 불러오기" 버튼을 누르게 하는 방법이 있었습니다. 하지만 폴링은 변경이 없을 때도 불필요한 I/O를 발생시키고, 실시간성이 떨어집니다. WatchService를 사용하면 **이벤트 기반**으로 파일 변화를 감시할 수 있어, 변경 발생 직후 즉각 처리할 수 있습니다.
- **Cross-platform 방식:** 각 OS별 파일감시 API를 직접 쓰려면 네이티브 코드를 OS마다 작성해야 하지만, Java WatchService는 이를 추상화하여 **한 API로 Windows, Linux, macOS 등의 파일 이벤트를 다룰 수 있게** 해줍니다.

**내부 동작:** WatchService는 **운영체제 종속 구현**을 가지며, 자바는 기본 파일시스템에 대한 WatchService를 OS에 맞게 제공합니다.

- 리눅스: `inotify` API를 사용합니다. 커널이 디렉터리 FD에 대해 변화 이벤트(생성, 삭제, 수정 등)를 큐에 넣어주고, Java는 이 큐를 읽어와 `WatchKey`를 통해 전달합니다.
- Windows: NTFS Change Journal 또는 FindFirstChangeNotification 등의 WinAPI를 사용해 구현되어 있습니다.
- Mac: kqueue 또는 FSEvents 등을 사용.

Java에서 `FileSystems.getDefault().newWatchService()`로 WatchService 인스턴스를 생성하고, 감시할 `Path`에 대해 `register()`를 호출하며 어떤 이벤트를 감시할지 지정합니다. 그러면 **해당 디렉터리에 변경이 생길 때마다** WatchService 내부에서 `WatchKey` 큐에 이벤트를 적재합니다. 개발자는 `watchService.poll()` 또는 `take()` 메서드로 **이벤트를 가져와 처리**하면 됩니다. `poll()`은 비동기 (즉시 반환)이며, `take()`는 이벤트가 올 때까지 블로킹합니다. 이벤트 종류는 `ENTRY_CREATE`, `ENTRY_DELETE`, `ENTRY_MODIFY` 등이 있습니다.

**사용 예시:** 예를 들어 특정 폴더에 새로운 파일이 생기면 콘솔에 알려주는 코드는 다음과 같습니다:

```java
try (WatchService watchService = FileSystems.getDefault().newWatchService()) {
    Path dir = Paths.get("/var/log/myapp");
    // 디렉터리에 CREATE, DELETE, MODIFY 이벤트 감시 등록
    dir.register(watchService, StandardWatchEventKinds.ENTRY_CREATE,
                                StandardWatchEventKinds.ENTRY_DELETE,
                                StandardWatchEventKinds.ENTRY_MODIFY);
    System.out.println("파일 변화 감시 시작... (" + dir + ")");

    while (true) {
        WatchKey key;
        try {
            key = watchService.take(); // 이벤트 발생까지 블로킹 대기:contentReference[oaicite:98]{index=98}
        } catch (InterruptedException e) {
            System.out.println("감시 중단됨");
            break;
        }

        // 발생한 이벤트 처리
        for (WatchEvent<?> event : key.pollEvents()) {
            WatchEvent.Kind<?> kind = event.kind();
            Path changedFileName = (Path) event.context();
            // context()는 변경이 일어난 파일/디렉터리의 **이름(Path)** (전체경로 아님)
            System.out.println("이벤트 발생: " + kind + " - " + changedFileName);
        }

        // 키 reset (다음 변경 감시를 위해 필수)
        boolean valid = key.reset();
        if (!valid) {
            System.out.println("감시 키가 유효하지 않음, 디렉터리가 삭제되었을 수 있습니다.");
            break;
        }
    }
} catch (IOException e) {
    e.printStackTrace();
}
```

이 코드는 `/var/log/myapp` 디렉터리에서 파일이 생성/삭제/수정될 때마다 이벤트를 받아 출력합니다. `watchService.take()`는 **이벤트가 있을 때까지 스레드를 잠재우므로**, CPU polling 없이 효율적으로 기다릴 수 있습니다. 그리고 이벤트가 도착하면 루프를 돌며 `key.pollEvents()`로 해당 키에 쌓인 모든 이벤트를 처리합니다. 마지막에 `key.reset()`을 호출하는데, 이것은 **한 번 사용한 WatchKey를 재사용**하기 위함입니다. 만약 reset하지 않으면 해당 키는 비활성화되어 더 이상 이벤트를 받지 못합니다.

**한계와 특징:**

- WatchService는 **디렉터리 단위**로 감시합니다. 파일 하나만 감시하고 싶어도, 그 파일이 있는 디렉터리를 감시하고 이벤트 발생 시 해당 파일인지 체크해야 합니다.
- 이벤트 유실 가능성: 매우 짧은 시간에 수많은 변화가 일어나면 (예: 디렉터리에 수천 파일 생성) 이벤트 큐가 Overflow될 수 있습니다. 이 경우 `OVERFLOW` 키 이벤트가 발생하고 일부 이벤트는 구체적인 정보를 잃습니다. 따라서 그런 경우를 염두에 두어야 합니다.
- **플랫폼 차이:** 대부분의 OS는 **파일 내용 수정**을 `ENTRY_MODIFY`로 알려주지만, 어떤 OS는 파일이 열려 있다 닫힐 때만 이벤트를 주거나, 하위 디렉터리 생성은 상위에서 감지 안 되는 등 세부 동작이 다를 수 있습니다. (Java doc에 플랫폼별 제약이 언급되어 있음)
- 폴더 깊이: 기본적으로 **하위 디렉터리 변화는 자동 감시되지 않습니다**. 모든 하위 폴더를 재귀적으로 감시하려면 각각 register해야 합니다. (Commons IO에 DirectoryWatcher 유틸 등이 이런 걸 쉽게 해주는 기능 있음)

**주요 활용:**

- **로그 모니터링 및 실시간 로딩:** 애플리케이션이 설정 파일을 자동 리로드하거나 로그 파일 변경 시 UI에 즉각 반영하는 데 사용.
- **동기화 도구:** 예를 들어 클라우드 드라이브 클라이언트나 백업 솔루션이 로컬 폴더 변화를 감지하여 업로드/백업을 수행할 때.
- **개발 편의:** IDE의 **파일 시스템 Watching** (파일이 외부에서 수정되었을 때 알려주는 기능)에 활용됩니다.
- **보안 모니터링:** 랜섬웨어나 삭제 등 이상 행동을 탐지하려 특정 경로의 변경을 감시하는 도구 등.

Baeldung에서는 Java WatchService에 대해 \*"NIO.2의 WatchService API는 디렉터리 변화를 모니터링하기 위한 **확장성 있는 솔루션**을 제공하며, 성능을 위해 매우 최적화되어 있어 직접 구현할 필요가 없다"\*고 언급합니다. 즉, 직접 파일 변경을 감시하는 쓰레드를 돌리기보다는, 이 API를 사용하면 OS 수준 최적화에 기대어 효율적으로 처리할 수 있다는 것입니다.

**장점:** 이벤트 기반으로 파일 변화를 다룸으로써 **실시간성**이 높고, 불필요한 디스크 탐색을 안 해도 됩니다. 구현이 비교적 간단하며 OS 리소스를 직접 활용하므로 성능도 좋습니다.

**단점:** 매우 광범위한 변경(예: 전체 디렉토리 삭제 등)에는 한계가 있고, 플랫폼 의존적 세부사항이 있어 100% 이식성은 담보하기 어렵습니다 (그러나 일반적인 사용에는 문제 없음). 또한 자바의 WatchService가 모든 OS 기능을 다 지원하지는 않아, 예를 들어 macOS에서는 대량 파일 변경 감지에 이슈가 있었다는 보고도 있습니다. 그럼에도, \*"실시간 파일 감시"\*라는 큰 요구사항을 비교적 적은 코드로 구현할 수 있다는 것이 큰 강점입니다.

---

## 7. ZIP/JAR 등 압축 파일 시스템 및 커스텀 파일시스템 – `FileSystems.newFileSystem(URI, ...)`

**철학 및 배경:** Java NIO.2는 **플러거블 파일시스템** 아키텍처를 도입했습니다. 이를 통해 로컬 디스크가 아닌 다른 매체도 `Path`와 `FileSystem` API로 다룰 수 있습니다. 대표적인 내장 구현이 \*\*ZIP 파일시스템(zipfs)\*\*입니다. ZIP나 JAR 파일을 마치 하나의 디렉터리 트리인 것처럼 접근할 수 있게 해줍니다. 철학은 "*압축 파일 안에 있는 파일도 일종의 파일이다 – 일반 디렉터리처럼 다룰 수 있으면 편하지 않을까?*" 하는 것입니다. 개발자는 ZIP 파일 내부를 해제하지 않고도 `Path`로 경로를 만들어 `Files.readAllBytes(zippath)`처럼 읽을 수 있습니다.

**해결하는 문제:**

- **ZIP/JAR 처리의 편의성:** 전에는 ZIP 파일을 다루려면 `java.util.zip.ZipInputStream` 등 스트림을 써서 Entry를 루프 돌며 읽고, 출력하려면 ZipOutputStream에 개별 Entry로 써야 했습니다. NIO.2의 zipfs를 쓰면 **평범한 파일 복사/읽기 메서드로** ZIP 내용을 다룰 수 있습니다. 예를 들어, 설정 파일이 들어있는 ZIP을 프로그램이 직접 열어 내용을 수정하고 다시 저장하는 작업을 간단히 할 수 있습니다.
- **추상화:** ZIP 외에도, 네트워크 파일시스템이나 가상 드라이브 등을 자바에서 손쉽게 다룰 수 있도록 합니다. 예를 들어 회사에서 만든 `MemoryFileSystem`이나 클라우드 스토리지 파일시스템 플러그인을 장착하면, 로컬 파일 다루듯 구글 드라이브의 파일을 다룰 수도 있습니다. (이런 건 JDK 기본 제공은 아니고, 서드파티나 직접 구현해야 합니다.)
- **아카이브 탐색:** JAR 파일 내 클래스 탐색 등도 이 API로 구현됩니다. Java의 ClassLoader도 내부적으로는 zipfs를 써서 .jar 안 클래스를 읽습니다.

**내부 동작:** `FileSystems.newFileSystem(URI, env)` 메서드를 호출하면, 해당 URI scheme에 등록된 파일시스템 프로바이더를 통해 `FileSystem` 객체를 얻습니다. 예컨대 `URI.create("jar:file:/path/to/archive.zip")` URI를 넣으면, `jar` 스킴을 처리하는 zipfs provider가 불려 ZIP 파일을 열고 새로운 파일시스템 인스턴스를 만듭니다. 그러면 그 `FileSystem` 안에서는 ZIP 안의 각 엔트리가 하나의 `Path`로 표현됩니다. 이 Path를 가지고 `Files.copy`나 `Files.newInputStream` 등을 호출하면, zipfs 구현이 해당 엔트리의 데이터를 읽어서 돌려줍니다. **ZIP 파일을 쓴다면** zipfs가 임시 파일에 기록하다가 close 시 실제 ZIP 형식으로 만들어줍니다.

Oracle 공식 문서에 따르면, JDK7에 도입된 zipfs는 JDK의 첫 커스텀 파일시스템 구현 사례로, JDK에 소스 코드(데모로) 포함되어 있습니다. zipfs는 `jimfs` 같은 다른 FS와 구조가 비슷하며, `FileSystemProvider`라는 SPI를 구현합니다.

**사용 예시:** ZIP 파일을 새로 만들고, 그 안에 파일을 복사하는 예제를 보겠습니다:

```java
// zip 파일시스템 생성 (ZIP 파일이 없으면 새로 생성)
Map<String, String> env = new HashMap<>();
env.put("create", "true");  // 없으면 생성
URI zipUri = URI.create("jar:file:/tmp/example.zip");
try (FileSystem zipFs = FileSystems.newFileSystem(zipUri, env)) {
    Path externalFile = Paths.get("/home/user/data.txt");            // 외부 실제 파일
    Path pathInZip = zipFs.getPath("/data.txt");                     // ZIP 내부 경로
    // 외부 파일을 ZIP 내부로 복사
    Files.copy(externalFile, pathInZip, StandardCopyOption.REPLACE_EXISTING);
    // 복사 완료 - 이제 example.zip 안에 data.txt 파일이 생겼음
}
// zipFs 닫히는 순간 ZIP 파일이 최종 작성됩니다.
```

위 코드에서는 `jar:file:/tmp/example.zip` URI로 새로운 ZIP 파일시스템을 생성했습니다. 그런 다음 `zipFs.getPath("/data.txt")`로 ZIP 내부 경로를 획득하고, `Files.copy`를 이용해 로컬 파일을 그 경로로 복사했습니다. 이 한 줄의 `Files.copy`로 사실상 ZIP 엔트리에 데이터를 쓰는 작업이 완료되었습니다. 스트림을 열고 바이트를 쓰는 등 세부 작업은 모두 zipfs가 처리했습니다. try-with-resources 블록을 벗어나 `zipFs.close()`가 호출되면, zipfs는 ZIP 디렉토리 구조와 압축 형식을 완성하여 실제 파일 `/tmp/example.zip`에 기록합니다. (env에 `"create","true"`를 주었으므로 ZIP 파일이 자동 생성된 것입니다.)

**응용:** zipfs를 이용하면 **JAR 교체나 동적 모듈 업데이트**도 쉽게 구현 가능합니다. 예를 들어, 실행 중에 라이브러리 JAR의 일부 리소스를 수정해야 할 경우 (잘 있진 않지만) zipfs로 해당 JAR를 열어 Path로 접근해 고치고 닫을 수 있습니다. 또한 **일회성 압축 해제** 없이 파일을 읽으므로 메모리 낭비를 줄입니다.

**가상 파일시스템 (VFS) 플러그인:** zipfs 외에도 Java 생태계에는 여러 파일시스템 구현체가 있습니다:

- **Jimfs (Google)**: 메모리 기반 FS (다음 절에 설명).
- **Fuse-J / Dropbox VFS**: Java NIO API로 Dropbox 같은 서비스를 접근하는 구현 (Commons VFS 플러그인 등).
- **TrueVFS**: ZIP, TAR 등 여러 아카이브와 FTP 등을 하나의 가상 경로 공간으로 다루는 강력한 VFS 라이브러리.
- **Hadoop FileSystem (HDFS)**: Hadoop은 자체 `FileSystem` API를 가지고 있지만, NIO provider를 통해 `Path`로 HDFS 접근을 가능케 하는 어댑터도 존재합니다.

자바 표준에 포함되진 않았지만 **Commons VFS** 라이브러리는 다양한 프로토콜(SFTP, HTTP, S3 등)을 일관된 FileObject API로 다루는 유명한 라이브러리입니다. Apache Commons VFS의 목표는 "*여러 가지 다른 파일시스템을 위한 단일 API 제공*"이며, 로컬 디스크, HTTP 서버, ZIP 아카이브 내부 등을 **통일된 인터페이스**로 접근하도록 합니다. 실제로 Commons VFS에는 S3, FTP, 웹DAV 등 수많은 플러그인이 있어, 한 줄의 URL로 원격 자원에 접근할 수 있습니다. (예: `VFS.getManager().resolveFile("s3://bucket/name")`).

**장점:**

- 개발 편의성 폭발💥: ZIP을 다룰 때, 별도 라이브러리 호출 없이 표준 파일 API 그대로 쓰면 되니 러닝커브가 낮고 생산성이 높습니다.
- 다양한 매체 추상화: 코드 한 세트로 로컬 파일이든 ZIP 엔트리든, 심지어 원격 파일시스템이든 다룰 수 있게 됩니다. 예를 들어 `Files.copy(path1, path2)`에서 path1이 로컬이고 path2가 ZIP 내부 경로여도 잘 작동합니다. 이는 **전략 패턴**으로 `FileSystemProvider` 구현만 바뀌고 API는 동일한 덕분입니다.
- ZIP 경우 성능도 충분히 양호: 내부적으로 `Deflater`를 사용해 압축/해제하므로 Java 레벨에서 효율적으로 처리됩니다.

**단점:**

- 사용 시 유의: ZIP 파일을 Write 모드로 열면 동시 접근이 제한됩니다. 또 ZIP 파일시스템은 메모리에 일부 색인 정보를 들고 있을 수 있어, 큰 ZIP을 열면 메모리 사용이 있을 수 있습니다.
- zipfs 구현상의 제한: ZIP64 (4GB 이상 ZIP) 같은 특수한 ZIP도 지원은 되지만, 복잡한 압축 방식이나 손상 복원 등은 지원 안 합니다. **일반적인 ZIP 용도에는 문제없음**.
- Commons VFS 등의 외부 라이브러리는 종종 예외 처리가 복잡하거나 실시간성 부족(예: HTTP HEAD 요청) 등의 문제가 있을 수 있습니다. 그러나 이는 특정 구현의 문제일 뿐, 개념적 단점은 아닙니다.

정리하면, **자바의 FileSystem 추상화는 파일 자원의 근원을 추상화하여** 우리가 흔히 쓰는 `Path`, `Files` API로 **압축파일 내부나 원격파일 등도 투명하게** 접근하게 해줍니다. 이것은 상당히 세련된 구조이며, 필요할 때 굉장히 유용합니다. (사실 많은 사람들이 존재 자체를 모르는 기능이기도 합니다!)

---

## 8. 외부 라이브러리를 활용한 파일 I/O 보조 도구

자바 표준 API 외에도, 오픈소스 라이브러리들이 다양한 파일 I/O 유틸리티를 제공합니다. 이들은 주로 **편의성과 추가 기능**에 초점을 맞추고 있어, 표준 API와 조합하면 개발 생산성을 크게 높일 수 있습니다. 몇 가지 주요 라이브러리와 그 강점을 알아보겠습니다:

### Apache Commons IO – `FileUtils`, `IOUtils` 등

Apache Commons IO는 자바 I/O 관련 유틸 기능을 한데 모은 라이브러리입니다.
일상적인 파일 작업들을 위한 헬퍼 메서드를 다수 제공합니다. 예를 들어:

- `FileUtils.copyDirectory(srcDir, destDir)`: 디렉토리 전체 복사 (하위 폴더 재귀 복사, 파일 날짜 보존 등 옵션).
- `FileUtils.deleteDirectory(dir)`: 디렉토리 삭제 (비어있지 않아도 통째로 삭제).
- `FileUtils.readFileToString(file, UTF_8)`: 파일을 통째로 읽어 문자열로 리턴.
- `IOUtils.toString(inputStream, UTF_8)`: InputStream 전체를 String으로 간편 변환.
- `IOUtils.copy(InputStream, OutputStream)`: 스트림 복사 (바이트를 효율적으로 버퍼링하여 복사) 등.

이러한 기능 일부는 Java 7+에서 `Files` API로 제공되지만, Commons IO는 Java 초창기부터 존재하여 표준 API가 부족했던 영역을 채워왔습니다.
오늘날에도 남아있는 이유는 편의 메서드의 종합적인 모음과 일부 특화 기능 때문입니다:
- `FileUtils.listFiles()`는 와일드카드/확장자 필터로 파일 목록을 수집
- `FileUtils.sizeOfDirectory()`처럼 한눈에 알기 쉬운 메서드 등

수십 개의 헬퍼를 통해 반복적인 입출력 코드를 크게 줄여줍니다.
Commons IO 사용자 가이드는 URL에서 바이트 읽어 출력하기 예제를 보여주는데, 전통적인 방식으로 8줄 넘게 작성하던 코드를 IOUtils 쓰면 3줄로 끝냅니다.

**주요 유틸 클래스:**

- `IOUtils`: 앞서 언급한 스트림/리더/Writer 관련 읽기쓰기 복사 등의 정적 메서드 제공.
- `FileUtils`: 파일 및 디렉터리 조작 관련 메서드 집합. 특히 디렉터리 복사/삭제, 파일 비교(`contentEquals`), 디렉터리 크기 계산 등이 유용합니다. 자바 표준으로 구현하려면 해야 할 반복작업들을 캡슐화했습니다.
- `FilenameUtils`: 파일 경로 문자열 관련 유틸 (확장자 추출, 경로 normalize 등). 운영체제별 경로 구분자 차이를 신경쓰지 않고 쓸 수 있습니다.
- `FileMonitor` (`FileAlterationObserver` 등): Commons IO는 간이 파일 변경 모니터링 기능도 제공합니다. (NIO WatchService 등장 이전에 만들어진 것으로, 폴링 기반이긴 하지만 OS별 notifer도 부분 활용)
- 기타: `EndianUtils`, `LineIterator` 등.

```java
// 예: 전체 디렉토리 복사 (하위 폴더 포함)
File srcDir = new File("/home/user/documents");
File destDir = new File("/home/user/backup/documents");
try {
    FileUtils.copyDirectory(srcDir, destDir);
    System.out.println("디렉터리 복사 완료!");
} catch (IOException e) {
    e.printStackTrace();
}

// 예: InputStream을 문자열로 읽기 (encoding 지정)
try (InputStream in = new URL("https://commons.apache.org").openStream()) {
    String html = IOUtils.toString(in, StandardCharsets.UTF_8);
    System.out.println("사이트 HTML 길이: " + html.length());
}
```

위 첫 번째 예에서 `FileUtils.copyDirectory`는 내부적으로 소스 디렉터리를 탐색하여 디렉터리 구조를 만들고, 파일들을 하나씩 복사합니다. 기존의 `Files.walkFileTree` 등을 사용하면 직접 구현해야 할 로직이 헬퍼 메서드 하나로 해결된 것입니다. 두 번째 예의 `IOUtils.toString(in, UTF_8)`은 `BufferedReader`로 일일이 읽고 StringBuilder에 이어붙이는 코드 수십 줄을 한 줄로 대체했습니다. 이러한 간결함이 곧 **생산성 향상과 버그 감소**로 이어집니다.

**장단점:**

- *장점:* 코드가 대폭 간결해집니다. "몇 줄짜리 반복 코드 제거"라는 목표대로, 자잘한 I/O 처리 코드를 일일이 작성하지 않아도 됩니다. 또한 Commons IO는 아주 널리 쓰여서 **신뢰성이 높고 레퍼런스도 많습니다**.
- *단점:* JDK7+부터는 `Files`에 상당 부분 기능이 생겨서, 중복되는 부분이 있습니다. 예컨대 `FileUtils.writeStringToFile` vs `Files.writeString`처럼 하는 일이 같은 경우가 많습니다. 따라서 **중복 의존성**이라는 관점에서, 순수 JDK로 가능하면 굳이 외부 라이브러리를 안 쓰는 선택지도 있습니다.
- 또한 **성능**: 편의 메서드들은 범용적으로 작성되었기 때문에, 아주 큰 데이터에서는 비효율일 수 있습니다. (예: IOUtils.toString으로 1GB 읽으면 1GB 문자열 객체를 만들어야 하니 메모리 소모 커짐.) 이런 문제는 개발자가 상황에 맞게 사용하면 되므로 큰 단점은 아닙니다.

### 8.2 Google Guava – **Files** & **MoreFiles** 유틸

**소개:** Guava는 구글의 공용 자바 라이브러리로, 컬렉션, 문자열, 동시성 등 여러 유틸을 담고 있습니다. 그 중 `com.google.common.io.Files` 클래스와 Java 7 도입 후 추가된 `MoreFiles` 클래스가 파일 관련 도구들입니다. Commons IO와 유사하지만 약간 관점이 다릅니다:

- `Files` (Guava): 주로 **java.io.File** 객체를 다루는 정적 메서드 제공. (왜냐면 Guava 개발 당시엔 Path가 없었음) 예: `Files.asCharSource(File, Charset)` → CharSource 반환, `Files.touch(File)` → 존재하지 않으면 0바이트 파일 생성 또는 업데이트, `Files.hash(File, HashFunction)` → 파일 내용의 해시 계산 등.
- `MoreFiles` (Guava): Java7 Path를 위한 유틸. 예: `MoreFiles.asCharSource(Path, Charset)` (동일 기능 Path 버전), `MoreFiles.fileTreeTraverser()` (디렉터리 트리를 탐색하는 Traverser 객체 제공) 등.

Guava Files의 장점은 Guava의 다른 부분과 **시너지**입니다. 예를 들어 Guava의 `Preconditions`로 인수 검증, `Charsets` 상수, `Hashing` 라이브러리 등이 Files 유틸과 함께 자주 쓰입니다. (Commons IO에도 해시 계산 메서드가 있지만, Guava Hashing은 더 다양한 해시를 지원하기도 합니다.)

**사용 예시:** Guava의 CharSource/CharSink는 특히 편리합니다:

```java
File logFile = new File("app.log");
Charset utf8 = StandardCharsets.UTF_8;
CharSource source = Files.asCharSource(logFile, utf8);
CharSink sink = Files.asCharSink(new File("app-copy.log"), utf8);

// 파일 전체를 문자열 리스트로 읽기
List<String> lines = source.readLines();  // 한 줄씩 읽어 리스트로
// 파일 복사 (문자 단위)
source.copyTo(sink);
```

`CharSource`와 `CharSink`는 Guava IO의 핵심 개념으로, **데이터의 공급자와 수신자**를 추상화한 객체입니다. `Files.asCharSource`는 File -> CharSource로 래핑하여, `readLines()`, `read()` (전체 as String) 등의 메서드를 제공합니다. 또한 `copyTo`로 Sink에 바로 출력할 수도 있습니다. 이러한 API는 **스트림을 다루는 것을 객체지향적으로 캡슐화**했다는 점이 돋보입니다. Okio의 BufferedSource/Sink와도 맥이 통합니다.

Guava Files가 제공하는 기타 기능:

- `Files.createTempDir()`: 임시 디렉터리 생성. (JDK는 임시파일만 지원하므로 요긴)
- `Files.fileTreeTraverser()`: 파일 트리 순회용 Traverser (Guava graph traversal API 활용).
- `Files.getFileExtension(String path)` 또는 `getNameWithoutExtension` 등 경로 문자열 유틸 (Commons FilenameUtils와 비슷).
- `Files.isDirectory().apply(file)` 같은 Predicate (Guava predicate와 연계).

**장점:** Guava를 이미 쓰고 있다면 별도 라이브러리 없이 편의 I/O 기능을 쓸 수 있습니다. Guava의 장점인 **가독성 높은 API**가 적용되어, 예컨대 CharSource/CharSink 개념은 자바 표준엔 없는 세련된 접근입니다. 또한 Guava는 **안정적**이고 구글에서 광범위하게 사용되므로 신뢰가 갑니다.

**단점:** Guava Files의 많은 부분은 Java NIO.2가 나온 후 `MoreFiles`로 옮겨갔고, Java 자체에도 유사 기능이 생겨서 상대적으로 덜 주목받습니다. 또한 Guava는 라이브러리 규모가 커서, 만약 프로젝트에서 Guava를 안 쓰는데 단순히 파일 I/O 때문에 추가하기엔 부담일 수 있습니다. (요즘은 Guava 거의 필수수준으로 쓰이지만요.)

### 8.3 Square **Okio** – 현대적인 I/O 프레임워크

**소개 및 철학:** Okio는 Square사(스퀘어, 현재 Block)에서 만든 I/O 라이브러리입니다. 원래 OkHttp (HTTP 클라이언트)의 내부 컴포넌트로 시작했다가 독립 라이브러리가 되었습니다. Okio의 핵심은 \*\*소스(Source)와 싱크(Sink)\*\*라는 추상화입니다. 이는 `InputStream`/`OutputStream`과 유사하지만, Java의 전통 스트림보다 **간결하고 강력한 API**를 제공합니다. 또한 `Buffer`와 `ByteString`이라는 자료구조로 효율성을 극대화했습니다.

**철학:** *"Java I/O를 더 쉽게, 더 빠르게 만들자."* Okio는 Java의 스트림 디자인에서 부족했던 부분을 개선합니다:

- Timeout 지원: `Source`/`Sink`는 읽기/쓰기에 timeout을 걸 수 있습니다.
- 쉬운 구현: Source 인터페이스는 `read(Buffer)` 하나 정도로 단순해서, InputStream처럼 구현 복잡도가 낮습니다. (OutputStream의 `available()` 같은 함정 메서드 없음)
- 풍부한 API: Buffer를 사용하면 바이트, UTF-8 문자열, 숫자 등을 편하게 읽고 쓸 수 있습니다. C의 stdio처럼 `bufferedSink.writeInt()` 같은 식으로 다양한 형식을 지원합니다.
- 문자 스트림 통합: InputStream/Reader의 구분 없이, Okio는 전부 바이트로 다루되 필요 시 UTF-8 등 변환을 쉽게 제공합니다.
- Zero-copy 설계: 내부적으로 Buffer는 **세그먼트 리스트**로 구현돼 있어서, 데이터 복사 없이 버퍼를 나누고 합칠 수 있습니다. 예를 들어 한 Buffer에서 다른 Buffer로 데이터를 이동할 때 메모리 복사가 일어나지 않습니다. 이것은 Okio 성능의 큰 비결입니다. (Segment pool이라는 바이트 배열 풀도 있어, 객체 할당을 줄임)

**사용 예시:** Okio로 파일을 읽는 간단한 코드를 보겠습니다 (Okio 3.x API 기준):

```java
import okio.BufferedSource;
import okio.Okio;
import java.io.File;

File file = new File("example.txt");
try (BufferedSource source = Okio.buffer(Okio.source(file))) {
    String firstLine = source.readUtf8Line();
    ByteString first5Bytes = source.readByteString(5);
    System.out.println("첫 줄: " + firstLine);
    System.out.println("다음 5바이트 (hex): " + first5Bytes.hex());
}
```

여기서 `Okio.source(file)`는 파일을 읽는 Source를 생성하고, `Okio.buffer(source)`는 이를 BufferedSource로 감쌉니다. (이 BufferedSource는 Java의 BufferedReader+InputStream 조합과 비슷한 역할). 그런 다음 `readUtf8Line()`으로 UTF-8 인코딩된 한 줄을 문자열로 읽고, `readByteString(5)`로 다음 5바이트를 `ByteString` 객체로 읽었습니다. `ByteString`은 불변(byte\[]) 타입으로, hex, base64 출력 등의 편의 메서드를 갖고 있습니다.

Okio를 이용하면 **데이터 처리 로직이 매우 깔끔**해집니다. e.g. 크기 지정 없는 `source.readUtf8()`은 파일 전체를 UTF-8 문자열로 읽어오고, `sink.writeUtf8("hello")`는 바이트 변환과 쓰기를 한 번에 합니다. 또한 Okio Buffer는 **thread-safe 분리**가 가능해서, 하나의 Buffer에서 데이터를 잘라 다른 스레드에 넘겨 처리하는 등 고성능 파이프라인 구성도 용이합니다.

Okio는 또한 **압축, 암호화 코덱**도 지원합니다. GzipSource/Sink, Deflate, Base64, Hashing 등 일반적인 처리를 데코레이터로 제공하여, 쉽게 스트림에 끼워넣을 수 있습니다. 예를 들어 `new GzipSource(source)`로 감싸면 읽으면서 자동 풀압축됩니다. 이러한 설계는 **데코레이터 패턴**이자, Java FilterInputStream보다 훨씬 사용이 쉽습니다.

**성능:** Okio는 안드로이드 및 서버 모두에서 성능 호평을 받았습니다. Segment pool과 효율적인 API 덕분에 **java.io** 대비 오버헤드를 크게 줄였습니다. 어떤 벤치마크에서는 Okio가 비슷한 작업을 하는 java.io보다 3배 이상 빠르다는 결과도 있습니다 (특히 많은 소량 I/O 작업에서). Square 측 자료에 따르면 OkHttp에서 Okio 도입 후 전체 throughput이 개선되었다 합니다.

**Kotlin 및 멀티플랫폼:** Okio 2.x부터는 Kotlin으로 포팅되어, Kotlin 멀티플랫폼 (Android, iOS, JVM) 라이브러리가 되었습니다. 그래서 Kotlin 개발자에게 더 친숙하게 다듬어졌고, 코루틴과의 통합도 잘되어 있습니다. 예컨대 Okio 3.x는 suspend 함수로 Source/Sink 읽기 쓰기를 제공하여, 비동기 코드를 쉽게 작성할 수 있습니다.

**장점:**

- *최고 수준의 성능:* 세그먼트 풀, zero-copy 버퍼링 등의 기법으로 매우 빠릅니다. 특히 많은 작은 I/O 조각 처리에 강합니다.
- *모던 API:* 가독성 좋고, 함수형 스타일에 가깝습니다. Java I/O의 여러 단점을 보완했습니다.
- *풍부한 추가 기능:* Base64, Gzip, hashing 등 편의 기능 내장. 별도 라이브러리 의존 줄일 수 있음.
- *멀티플랫폼:* 자바 외에 Kotlin Native(iOS) 등에서 동일 코드 사용 가능 (멋진 부가이익).

**단점:**

- 학습 곡선: java.io에 익숙한 개발자에게는 개념이 생소할 수 있습니다. 하지만 문서와 사용 예제가 잘 되어 있어 큰 문제는 아닙니다.
- 의존성 추가: Okio는 가벼운 편이지만 (수백 KB), 프로젝트에 외부 lib 추가를 꺼리는 경우 고려.
- 파일 통합성: Okio는 **파일시스템 API 자체를 대체하진 않습니다**. 파일 여는 것(Okio.source)은 결국 File/Path 필요하고, NIO FileChannel 기능 (락, memory-map 등)은 Okio에 노출 안 되어 있습니다. 그러므로 Okio는 **스트림 처리 성능 향상** 용도로 쓰이고, 특별한 파일 제어는 여전히 표준 API를 사용해야 합니다.

Okio를 요약하면, \*"Java I/O의 부족함을 채운 세련된 대안"\*입니다. 특히 OkHttp, Moshi(JSON 파서) 등과 함께 Android/Java 네트워킹에서 사실상의 표준으로 자리잡았고, 현재도 활발히 유지보수되고 있습니다.

### 8.4 Apache Commons **VFS** / Schlichtherle **TrueVFS** – 가상 파일시스템 통합

**소개:** 앞서 7절에서 Java NIO의 FileSystem SPI와 zipfs를 봤는데, Commons VFS와 TrueVFS는 그 아이디어를 더 확장한 *외부 라이브러리*입니다. 이들은 **다양한 파일 저장소(S3, FTP, ZIP 등)를 하나의 추상화로 볼 수 있게** 해줍니다.

- **Apache Commons VFS:** 로컬, FTP, SFTP, HTTP, WebDAV, SMB, ZIP, gzipped tar 등 무수한 파일소스들을 통일된 `FileObject` API로 다룹니다. 예를 들어 `"sftp://user:pass@host/dir/file.txt"`를 `VFS.getManager().resolveFile()`로 `FileObject`를 얻으면, 이후 `fileObject.getContent().getInputStream()`으로 내용 읽는 등 로컬 파일처럼 다룰 수 있습니다. Commons VFS는 플러그인 구조로, 원하는 프로토콜에 대한 provider JAR만 추가하면 동작합니다. (예: vfs-s3, vfs-jdbc 등 별도 제공). Commons VFS의 슬로건은 "*여러 다른 소스의 파일을 단일 API로 볼 수 있게, 다양한 소스의 파일을 한눈에 다룬다*".

- **TrueVFS (옛 TrueZIP):** 이것도 Java7을 지원하는 가상 파일시스템 라이브러리로, 특히 **아카이브 파일 (zip, tar 등)을 마치 폴더처럼 다루는 것**에 중점을 둡니다. TrueVFS는 자체 API도 있지만 NIO `Path`와 통합도 잘되어 있어, 전역적으로 TrueVFS Kernel을 설정해두면 `.zip`파일 경로를 `Paths.get("archive.zip/entry.txt")` 식으로 접근할 수도 있습니다. (동작 원리는 custom FS provider를 등록하는 식) TrueVFS는 nested archives (ZIP 안에 ZIP)도 처리하고, 강력한 캐싱 및 성능 최적화 기법을 갖고 있어 엔터프라이즈에서 쓰였습니다. 다만 최근 크게 유지되진 않아 사용 시 주의가 필요할 수 있습니다.

**해결되는 문제:**

1. **다양한 프로토콜 파일 접근:** 예를 들어 AWS S3에 있는 파일을 다운로드하려면 일반적으로 AWS SDK를 써야 하지만, Commons VFS를 쓰면 `FileObject s3file = manager.resolveFile("s3://..."); s3file.copyFrom(localFile, Selectors.SELECT_SELF);`처럼 간단히 표현 가능합니다. 개발자는 각 API 상세를 모르더라도 VFS 추상화로 작업을 수행할 수 있습니다.
2. **아카이브의 투명한 처리:** TrueVFS는 zip, tar, 7z 등의 포맷을 모두 파일시스템화하기 때문에, 압축 해제/압축 과정을 수동으로 코딩할 필요 없이 `Files.copy(srcPath, destPath)`를 하면 srcPath가 zip내 경로고 destPath가 로컬 경로여도 잘 동작합니다.
3. **파일 위치의 추상화:** 응용프로그램이 입력 파일 경로를 받는데, 이게 로컬 파일일 수도 있고 원격 URL일 수도 있고 zip안 경로일 수도 있을 때, VFS를 사용하면 입력 경로 문자열 앞부분 (scheme)을 파싱해 자동으로 해당 FileObject를 만들고 처리 가능하므로 **유연성**이 높아집니다.

**사용 예시 (Commons VFS):**

```java
FileSystemManager fsManager = VFS.getManager();
FileObject ftpFile = fsManager.resolveFile("ftp://user:pass@ftp.example.com/pub/test.txt");
FileObject localFile = fsManager.resolveFile("file:///C:/temp/test.txt");

// FTP 파일을 로컬로 다운로드
localFile.copyFrom(ftpFile, Selectors.SELECT_SELF);
// FileObject 사용 후에는 닫기
ftpFile.close();
localFile.close();
```

이 코드에서 `resolveFile`에 준 URL의 scheme에 따라 내부적으로 FTP provider가 동작해 `ftpFile`을 생성했습니다. 이후 `copyFrom`으로 복사할 때, VFS가 **알아서** FTP 입력스트림을 열고 로컬 출력스트림을 열어 데이터를 복사합니다. 개발자는 복사 로직을 신경쓸 필요 없이 **한 줄 호출로 서로 다른 파일시스템 간 복사를 수행**한 것입니다. (사실 copyFrom 내부에서 Streams를 쓰겠지만 그것을 추상화했으니 장점이지요.)

**사용 예시 (TrueVFS):**

TrueVFS는 NIO와 통합되어, 예를 들어 TrueVFS를 클래스패스에 넣으면 `"zip:file:/path/archive.zip!/dir/entry.txt"` 같은 URI를 `Paths.get`으로 열 수 있습니다. (느낌표 `!`는 JDK zipfs와 비슷한 문법) 아니면 TrueVFS API로 `TFile archiveFile = new TFile("archive.zip/dir/entry.txt"); String content = new TFileReader(archiveFile).readLine();` 처럼 사용합니다. TrueVFS는 **Thread-safe**하고, Zip 엔트리를 부분 수정해도 다른 엔트리 다시 안쓰는 등 최적화가 뛰어납니다.

**장점:**

- *놀라운 범용성:* 한 번 익혀 두면 **거의 모든 종류의 파일 저장소를 통일되게** 다룰 수 있습니다. 코드를 변경하지 않고 입력 경로만 바꾸어 로컬->원격 대상이 바뀌어도 동작하게 할 수 있습니다.
- *개발 비용 감소:* 각 프로토콜별 API 습득 시간을 줄이고, 일관된 error handling 가능.
- *복잡한 시나리오 지원:* Commons VFS는 caching 옵션, refresh, polling 등 부가 기능도 있고, TrueVFS는 여러 nested archive, 갱신 flush 정책 등 세세한 튜닝이 가능하여 **복잡한 파일 접근 시나리오**를 커버합니다.

**단점:**

- *추가 의존성 및 러닝커브:* VFS, TrueVFS 모두 개념적으로 방대하고, 종종 발생하는 예외나 동작을 이해하려면 학습이 필요합니다. (예: VFS에서 `FileObject`는 쓰고 나면 반드시 `.close()`해야 리소스 누수 없음, TrueVFS에서는 마지막에 `Tvfs.umount()` 해줘야 변경사항 flush됨 등.)
- *성능 오버헤드:* 추상화 계층이 추가되므로 로컬 파일만 다루는 간단한 상황에서는 약간 성능 손해가 있을 수 있습니다. 하지만 대부분 I/O 자체가 병목이라 오버헤드는 미미합니다.
- *라이브러리 활성도:* Commons VFS는 그래도 유지되고 있으나 아주 활발하진 않습니다. TrueVFS는 2018 이후 업데이트가 없어, 최신 환경에서는 이슈 가능성도.
- *특수 상황 한계:* 100% 투명하지 않을 수 있음. 예를 들어 VFS로 Windows UNC path 같은 edge case를 다룰 때 등.

**언제 유용한가:** 파일 동기화 도구, 클라이언트에 입력 파일 경로 다양하게 받는 툴 등에서 유용합니다. 만약 애플리케이션이 "**파일경로 문자열**"을 받아 처리하는 구조라면, VFS를 도입해두면 사용자가 `http://` 경로를 넣든 `zip:`을 넣든 다 처리할 수 있게 만들어 확장성을 얻을 수 있죠.

한 줄 평으로: *"Commons VFS/TrueVFS는 **파일시스템의 경계를 허문 추상화**이며, 다양한 저장소를 단일 인터페이스로 제어케 한다"*.

### 8.5 Google **Jimfs** – 자바용 인메모리 파일시스템 (테스트용)

**소개 및 배경:** **Jimfs**는 구글이 만든 **메모리 기반 파일시스템 구현체**입니다. Java NIO의 `java.nio.file.FileSystem`을 구현하여, **RAM 속에 가상의 디스크**를 흉내냅니다. 주요 목적은 **단위 테스트** 등에서, 실제 디스크에 영향 없이 파일 I/O를 테스트하거나, Windows/Unix 간 경로 차이를 OS 상관없이 검증하는 것입니다. Jimfs는 Java 8 이상에서 동작하며, POSIX 동작을 충실히 모사합니다 (SymLink, permissions 등 지원).

**철학:** "*파일시스템 의존성을 격리*" – 개발자가 파일 입출력 로직을 테스트할 때 실제 파일을 생성/삭제하면 테스트가 느려지고, CI/CD 환경에 따라 동작이 다르거나, 깨끗한 환경을 유지하기 어렵습니다. Jimfs를 쓰면 **모든 파일 조작이 메모리에서 이뤄지므로** 빠르고, 테스트 간 간섭이 없으며, OS별 차이도 통제 가능합니다. Google에서는 이 라이브러리를 자신들의 코드베이스 테스트에 활용해왔습니다.

**사용 예시:** Jimfs 사용은 매우 간단합니다:

```java
// Jimfs 설정: 유닉스 호환 모드 파일시스템 생성
FileSystem memFs = Jimfs.newFileSystem(Configuration.unix());
Path foo = memFs.getPath("/foo");
Files.createDirectory(foo);
Path hello = foo.resolve("hello.txt");
Files.writeString(hello, "Hello Jimfs!", StandardOpenOption.CREATE);
System.out.println("Jimfs 파일 크기: " + Files.size(hello) + "바이트");  // 출력: 13바이트
// 파일 읽기
String content = Files.readString(hello);
System.out.println("내용: " + content);
```

위에서 `Jimfs.newFileSystem(Configuration.unix())`로 Unix 스타일 경로 (`/` 루트)와 동작을 갖는 인메모리 FS를 생성했습니다. Windows 스타일 (`Configuration.windows()`)로 만들면 드라이브 레터, 역슬래시 경로 등도 흉내냅니다. 그런 다음 보이는 것처럼 `memFs.getPath`로 경로를 만들고 `Files`의 모든 정적 메서드를 그대로 사용했습니다. Jimfs는 **java.nio.file.FileStore**도 구현해서 `Files.getFileStore(path)` 등도 동작하고, 기본적인 속성 (용량, 사용량은 의미 없으나 100% 사용 가능으로 리턴)도 지원합니다.

- **POSIX 호환성:** Jimfs는 실제 유닉스와 거의 유사하게 움직입니다. 예를 들어 **심볼릭 링크** 기능이 있어서 `Files.createSymbolicLink()` 사용 가능하고, 권한도 `Files.setPosixFilePermissions`로 설정하고 `Files.getPosixFilePermissions`로 조회 가능합니다. 다만 Jimfs이므로 권한은 보안 효과보다는 시뮬레이션 용도입니다.
- **성능:** 메모리 상에서 동작하므로 디스크 I/O보다 훨씬 빠릅니다. 수천 파일 생성/삭제도 지연이 없습니다. 물론 대용량 파일 데이터를 Jimfs에 넣으면 그만큼 JVM 메모리를 쓰니 주의가 필요합니다.
- **Thread-safety:** Jimfs는 스레드세이프하게 설계되어 동시에 여러 쓰레드가 파일조작해도 문제 없도록 되어 있습니다.

**활용 사례:**

- *단위 테스트:* 예를 들어 파일 쓰는 함수가 정상 작동하는지 테스트하려면, Jimfs를 써서 메모리 파일시스템에 파일을 쓰게 하고 결과를 확인한 뒤, test가 끝나면 FileSystem을 닫아(umount) 버립니다. **실제 디스크를 건드리지 않으니 테스트 격리가 완벽**하고, Windows/Unix 경로 다호환 가능해서 OS마다 테스트 케이스 작성 안 해도 됩니다.
- *애플리케이션 임시 스토리지:* 어떤 경우에는 실제 디스크보다 메모리 저장이 나을 때 Jimfs를 쓰기도 합니다. 예컨대 애플리케이션이 runtime에 임시 파일을 많이 쓰는데, 종료 시 휘발되어도 될 경우, Jimfs 위에 경로를 잡아 쓰면 IO 성능이 빨라집니다. (이건 주로 Java 프로그램이 /tmp를 램디스크로 쓰는 것과 유사한 아이디어)
- *학습 및 시뮬레이션:* 파일 시스템 관련 기능을 개발할 때, Jimfs로 각종 시나리오를 가짜로 만들어 볼 수 있습니다. (파일시스템 event, 읽기전용 속성 등)

**장점:**

- *테스트 용이성:* Mock 없이 실제와 유사한 환경을 만들어 주니 테스트 코드가 간단하고 확실합니다. Jimfs를 쓰면 \*"테스트에서 실제 FS 안 써도 충분히 검증할 수 있다"\*는 큰 이점이 있습니다.
- *운영체제 독립성:* Jimfs로 Windows path 규칙도 Linux에서 테스트 가능하니, 크로스플랫폼 개발에 도움 됩니다.
- *기능 완성도:* Google이 회사 내부에서 쓰던 것이어서 그런지, 상당히 완성도가 높습니다. (예: 기본 attribute view 다수 지원, WatchService까지 실험적으로 지원).
- *경량:* 의존성 하나 추가로 작은 라이브러리고, JDK API와 자연스럽게 통합됩니다.

**단점:**

- *메모리 사용:* 모든 데이터가 JVM 힙(정확히는 off-heap 아닐까? 보통 On-heap이겠네요) 에 있으므로, 큰 파일을 다량 저장하면 OutOfMemory 위험이 있습니다. 따라서 수십 MB 이상 데이터를 테스트로 올리는 건 지양해야 합니다.
- *지속성 없음:* 휘발성이라, 운영 용도로 쓰려면 데이터 손실 감수해야. (메모리 DB처럼, 캐시나 임시 목적 외엔 적합치 않음)
- *WatchService 실험적:* Jimfs가 WatchService도 구현했지만, 실무 이벤트 정확도까지는 보장 장담 어렵습니다. (SmallFS 특징)

**비교:** Java 자체도 기본 `RAMDisk`는 없으므로, Jimfs는 **Java 진영 거의 유일무이한 인메모리 FS**입니다. .NET엔 InMemory file system test double이 MS 제공되지만, Java엔 이 external library가 사실상 표준입니다. 대안으로 JUnit의 TempDir (임시 폴더) 정도 쓰겠지만, 그것은 결국 실디스크이므로 Jimfs만한 격리성은 없습니다.

---

**맺음말 (비교 정리):** 이제 다양한 파일 처리 방식과 라이브러리를 살펴봤습니다. 요점을 간략히 표로 정리하면:

| **카테고리**             | **대표 API/도구**                          | **주요 장점**                               | **주요 단점**                           | **언제 적합한가**                    |
| -------------------- | -------------------------------------- | --------------------------------------- | ----------------------------------- | ------------------------------ |
| **간단 파일 I/O**        | `Files.readString/writeString`         | 코드 간결, 자원 자동관리, 작은 I/O에 효율적             | 큰 파일 메모리 부담, 스트리밍 불가                | 작은 설정/로그/텍스트 파일 처리             |
| **버퍼/스트림 I/O**       | `BufferedReader/Writer`, etc.          | 시스템콜↓ 성능↑, 스트리밍 처리 가능                   | 코드 다소 장황, flush 누락 시 데이터 유실 가능      | 대용량 텍스트, 소켓/파일 스트리밍            |
| **Zero-copy 대용량 전송** | `FileChannel.transferTo/From`          | CPU부하↓, 대용량 파일 네트웍 전송 최적                | 소스-타겟 제약(주로 파일→소켓), 소규모엔 효과 적음      | 파일 서버, 대용량 로그 네트워크 전송          |
| **메모리 맵핑 I/O**       | `FileChannel.map` (`MappedByteBuffer`) | 랜덤액세스 빠름, 고성능 (주식/DB)                   | OOM 위험 관리 필요, flush 고려 필요           | 대용량 파일 랜덤 읽기, 메모리 공유 필요        |
| **비동기 파일 I/O**       | `AsynchronousFileChannel`              | 쓰레드 대기 없음, 병렬 I/O 처리                    | 구현 복잡, OS지원 한계로 드라마틱 효과 제한          | GUI메인쓰레드 I/O, 백그라운드 병렬 작업      |
| **파일 변경 감시**         | `WatchService`                         | 실시간 이벤트, 폴링불필요                          | 플랫폼 제약/overflow 가능, 하위폴더 재귀X        | 설정 변경 자동로딩, 폴더 동기화/감시          |
| **ZIP/커스텀 FS**       | `FileSystems.newFileSystem(zip)`       | ZIP 내부 투명 접근, 커스텀 FS 확장성                | 러닝커브, 쓰기 시 리소스 관리 필요                | 압축파일 읽기/쓰기, 통합 FS 접근           |
| **Commons IO**       | `FileUtils`, `IOUtils`                 | 수십개 편의 메서드, boilerplate 감소              | JDK 중복기능 존재, 대용량 사용시 주의             | 빠른 개발 필요시, JDK 부족기능 보완         |
| **Guava IO**         | `Files`, `MoreFiles` (Guava)           | Fluent API, CharSource 등 편리, Guava와 시너지 | Guava 미사용시 굳이 추가 필요?, 기능 JDK와 중복 일부 | Guava 사용자, CharSource/Sink 활용시 |
| **Okio**             | `BufferedSource/Sink`, `Buffer`        | 매우 빠름(세그먼트풀), API모던, 멀티플랫폼              | 익숙치 않을 수 있음, FS 제어기능은 없음            | 고성능 네트워크/로깅, 코틀린 멀티플랫폼         |
| **Commons VFS**      | `VFS.getManager().resolveFile`         | 다종 저장소 단일 API, 파일소스 추상화                 | 초기 학습 필요, 최신 유지성 보통                 | 다양한 입력/출력 소스 처리, 유연성 필요        |
| **TrueVFS**          | `TFile`, NIO Provider                  | nested archive까지 FS처럼, 성능 최적            | 유지보수 불확실 (2018), 전역 설정 필요           | 복잡한 아카이브 다룰 때 (특수)             |
| **Jimfs (메모리 FS)**   | `Jimfs.newFileSystem`                  | 테스트에 유용, OS독립, 빠름                       | 휘발성, 메모리 한정, 운영 데이터엔 부적합            | 단위테스트, 일시적 가상파일 활용             |

이 표를 참고하여, 상황에 맞는 도구를 선택할 수 있습니다:

- **예1:** "한 두 번 읽고 마는 설정파일 로드" → Files.readString (간단).
- **예2:** "실시간 로그 파일에 지속적으로 기록" → BufferedWriter (버퍼링).
- **예3:** "100GB 파일을 네트워크로 전송" → FileChannel.transferTo (zero-copy) 권장.
- **예4:** "대용량 바이너리 파일에서 인덱스 위치들만 찾아 읽기" → MappedByteBuffer 적합 (랜덤액세스 우수).
- **예5:** "GUI에서 큰 파일 열기 버튼 눌렀을 때 UI freeze 없이 내용 미리보기" → AsynchronousFileChannel로 비동기 읽기 or 가상쓰레드 I/O.
- **예6:** "사용자가 편집 중인 설정파일 외부에서 바뀌면 자동 재로딩" → WatchService로 변경 이벤트 감지.
- **예7:** "프로그램이 압축파일 (.zip) 내 데이터를 직접 편집" → zipfs (FileSystems.newFileSystem)로 JAR 처리.
- **예8:** "파일 복사, 삭제 등 보일러플레이트 줄이기" → Commons IO FileUtils 사용으로 생산성↑.
- **예9:** "여러 개의 서로 다른 저장소(S3, FTP)에서 파일 수집" → Commons VFS로 통합 처리 (구현 용이).
- **예10:** "파일 I/O 로직 유닛테스트" → Jimfs로 실제 디스크 없이 검증.

마지막으로, **기술 선택 시 주의할 점:** 너무 과하지 않은 방법을 고르는 것이 중요합니다. 작은 작업에 메모리맵이나 Async I/O를 쓰는 것은 \*\*과공능(Overengineering)\*\*일 수 있고, 반대로 대용량 처리를 평범한 방식으로 하면 성능 부족이 생깁니다. 각 접근방식의 역사와 내부를 이해했다면, 이제 실제 개발 전에 *"이 유즈케이스에는 이 방식이 옳겠군, 이건 과하겠군"* 하고 **근거를 가지고 판단**할 수 있을 것입니다. 항상 \*\*요구사항 (데이터 크기, 동시성, 성능, 복잡도)\*\*을 먼저 따져보고, 가장 적합한 도구를 선택하는 것이 **현명한 개발자의 길**일 것입니다. 🚀

**참고 자료:** Java 공식 문서와 다수 블로그에서 언급된 내용을 함께 참조하여 작성했습니다. 예컨대 Oracle 도움말의 NIO 패키지 개요, Baeldung 등 튜토리얼, Medium 기술 블로그, Google의 라이브러리 문서 등이 상세 설명을 제공합니다. 필요시 해당 출처를 확인해보면 더욱 깊은 이해에 도움이 될 것입니다.

## 시스템 콜과 버퍼링

파일 I/O 작업의 효율성은 시스템 콜(System Call) 횟수에 영향을 받습니다.
사용자 공간의 애플리케이션이 파일에 데이터를 쓸 때마다 JVM은 커널에 작업을 요청하는 시스템 콜(`write()`)을 발생시킵니다.

애플리케이션이 `write()`를 호출하면 데이터는 다음과 같은 여정을 거칩니다.

1. 사용자 공간 버퍼 (User-Space Buffer): `BufferedWriter` 사용 시 데이터는 JVM 힙 메모리에 있는 버퍼에 먼저 저장됩니다.
2. 시스템 콜 및 컨텍스트 스위칭: 버퍼가 차면 `write()` 시스템 콜이 발생하고, 커널 모드로 전환됩니다.
3. 커널 공간 버퍼 (Kernel-Space Buffer / Page Cache): 데이터는 사용자 공간 버퍼에서 커널의 페이지 캐시로 복사됩니다. 이 시점에서 애플리케이션은 쓰기 작업이 완료되었다고 인식하고 다음 작업을 수행합니다.
4. 디스크 쓰기 (Physical Write): 커널은 자체적인 스케줄링 정책(I/O 스케줄러)에 따라 페이지 캐시의 '더티(dirty)' 페이지들을 모아 디스크에 비동기적으로 기록합니다.

`write()` 시스템 콜이 성공적으로 반환되었다는 것은 데이터가 디스크에 물리적으로 기록되었음을 의미하지 않으며, 단지 커널의 페이지 캐시에 복사되었음을 의미할 뿐입니다.
따라서 만약 이 시점에 시스템 전원이 꺼지면 페이지 캐시에 있던 데이터는 유실됩니다.

데이터베이스의 ACID 중 내구성(Durability)를 보장하기 위해서는 커널에 페이지 캐시의 내용을 즉시 디스크에 동기화하라는 명령(리눅스의 `fsync` 시스템 콜에 해당)을 보내는 메서드들을 호출해야 합니다.
성능 저하가 있을 수 있지만, 데이터 무결성을 보장합니다.
- `FileChannel.force()`
- `RandomAccessFile.getFD().sync()`

`write` 같은 시스템 콜이 발생하면 사용자 모드에서 커널 모드로의 컨텍스트 스위칭(Context Switch)이 발생하고 상당한 오버헤드를 수반합니다.

예를 들어, 1바이트씩 10,000번을 쓰는 코드는 10,000번의 시스템 콜을 발생시킬 수 있습니다.
이는 CPU 자원을 낭비하고 디스크 I/O를 파편화시켜 전체 성능을 크게 저하시킵니다.

버퍼링(Buffering)은 통해 이런 문제를 해결할 수 있습니다.
데이터를 즉시 파일에 쓰는 대신, 메모리(User-Space Buffer)에 일정 크기만큼 모았다가 버퍼가 가득 차거나 특정 조건이 만족될 때 한 번에 큰 덩어리로 시스템 콜을 호출하면 시스템 콜 횟수를 줄이고 그에 따라 I/O 성능을 향상시킬 수 있습니다.

## 효율적인 파일 생성 및 쓰기 방법

### `Files.write()` / `Files.writeString()` (NIO)

`Files.write()` 및 `Files.writeString()`는 작은 크기의 파일을 간편하게 쓸 때 가장 효율적입니다.

이 메서드들은 내부적으로 버퍼링을 사용하며, 파일이 존재하지 않으면 `StandardOpenOption.CREATE` 옵션을 통해 자동으로 생성합니다.
파일 생성과 쓰기 작업이 원자적으로(atomically) 처리되지는 않지만, 하나의 메서드 호출로 통합되어 코드가 간결해집니다.

- 시스템 콜 관련: 적은 양의 데이터를 쓸 때 내부 버퍼링을 통해 불필요한 시스템 콜을 방지합니다.
- 자원 관리: 또한 `try-with-resources` 구문 없이도 내부적으로 스트림을 안전하게 열고 닫아 자원 누수를 방지합니다.

설정 파일, 로그 메시지 등 크기가 작거나 중간 정도인 파일 쓰기에 적합합니다.

```java
Path path = Paths.get("output.txt");
List lines = Arrays.asList("Java NIO File Writing", "Efficient Approach");

// 파일이 없으면 생성(CREATE), 이미 있으면 내용을 덮어씀(TRUNCATE_EXISTING)
Files.write(path, lines, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
```

```java
// Path 객체는 파일 시스템의 경로를 추상화합니다.
// OS에 독립적인 경로 표현을 가능하게 해줍니다.
Path filePath = Paths.get("simple-output.txt");

// 파일에 쓸 문자열 리스트를 준비합니다.
List lines = Arrays.asList(
        "Java 17: Files.write() Example",
        "This is a simple and modern way to write to a file.",
        "Each line in the list will be a new line in the file."
);

try {
    // Files.write()는 파일 쓰기를 위한 가장 직관적인 NIO API입니다.
    // 이 메서드 하나로 파일 열기, 쓰기, 닫기 작업이 모두 처리됩니다.
    // 내부적으로 스트림을 열고 닫으므로 자원 누수에 안전합니다.

    // filePath: 데이터를 쓸 파일의 경로입니다.
    // lines: 파일에 쓸 문자열 데이터(Iterable). 각 요소는 한 줄이 됩니다.
    // StandardOpenOption: 파일 열기 옵션을 지정합니다.
    //   - CREATE: 파일이 존재하지 않으면 새로 생성합니다.
    //   - TRUNCATE_EXISTING: 파일이 존재하면 기존 내용을 모두 지우고 새로 씁니다. (덮어쓰기)
    //   - APPEND: 파일이 존재하면 기존 내용의 끝에 데이터를 추가합니다.
    Files.write(filePath, lines, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);

    System.out.println("File written successfully using Files.write(): " + filePath.toAbsolutePath());

} catch (IOException e) {
    // 파일 쓰기 도중 디스크 공간 부족, 권한 없음 등의 I/O 예외가 발생할 수 있습니다.
    System.err.println("An I/O error occurred: " + e.getMessage());
    e.printStackTrace();
} catch (SecurityException e) {
    // 파일 쓰기 권한이 없는 경우 SecurityException이 발생할 수 있습니다.
    System.err.println("Write access to file was denied: " + e.getMessage());
    e.printStackTrace();
}
```

### `BufferedWriter`

대용량 텍스트 파일을 쓰거나 스트리밍 데이터를 처리할 때 사용합니다.

`FileWriter`와 같은 기본적인 `Writer` 객체를 감싸서(Decorator 패턴) 사용자 공간에 버퍼를 추가합니다.
`write()` 메서드가 호출되면 데이터는 파일이 아닌 메모리 버퍼에 먼저 쓰입니다.
버퍼가 가득 차거나 `flush()`, `close()`가 호출될 때만 실제 물리 파일에 대한 시스템 콜이 발생합니다.

- 시스템 콜 관련: 버퍼링을 통해 수많은 쓰기 요청을 단 몇 번의 시스템 콜로 통합하여 컨텍스트 스위칭 비용을 대폭 줄입니다.
- CPU 사용량 감소: 잦은 I/O 대기 상태를 방지하여 CPU가 다른 작업을 처리할 수 있도록 합니다.

대용량 CSV 파일 생성, 긴 텍스트 문서 저장, 실시간으로 생성되는 로그 데이터 기록 등에 적합합니다.

`try-with-resources`를 사용하거나 `finally` 블록에서 `close()`를 호출해야 버퍼에 남아있는 데이터가 파일에 완전히 쓰이는 것을 보장할 수 있습니다.
`close()`는 내부적으로 `flush()`를 호출합니다.

```java
try (BufferedWriter writer = new BufferedWriter(new FileWriter("large_output.txt"))) {
    for (int i = 0; i < 10000; i++) {
        writer.write("This is line " + i);
        writer.newLine();
    }
} catch (IOException e) {
    e.printStackTrace();
}
```

```java
String fileName = "buffered-output.log";

// Java 7부터 도입된 try-with-resources 구문입니다.
// 이 블록이 끝나면 괄호 안에 선언된 자원(writer)의 close() 메서드가 자동으로 호출됩니다.
// 이를 통해 자원 누수를 안정적으로 방지할 수 있습니다.
try (
    // FileWriter는 파일에 문자 기반 스트림을 생성합니다. 파일이 없으면 자동으로 생성됩니다.
    // BufferedWriter는 FileWriter를 감싸(Wrapping) 버퍼링 기능을 추가합니다.
    // 기본 버퍼 크기(보통 8192자)를 사용하여 I/O 성능을 향상시킵니다.
    BufferedWriter writer = new BufferedWriter(new FileWriter(fileName))
) {
    // 이제부터 writer.write() 호출은 실제 디스크에 쓰는 것이 아니라,
    // 메모리에 있는 내부 버퍼에 데이터를 쌓는 역할을 합니다.
    writer.write("--- Log Start ---");
    writer.newLine(); // OS에 맞는 개행 문자를 추가해주는 편리한 메서드입니다.

    for (int i = 1; i <= 1000; i++) {
        writer.write(String.format("Log entry #%d: This is a test log message.", i));
        writer.newLine();
        // 이 루프 동안 수많은 write 호출이 발생하지만, 실제 시스템 콜(디스크 쓰기)은
        // 버퍼가 가득 찼을 때만 간헐적으로 발생합니다. 이것이 버퍼링의 핵심입니다.
    }

    writer.write("--- Log End ---");
    writer.newLine();

    // 명시적으로 flush()를 호출하면 버퍼에 남아있는 모든 데이터를 즉시 디스크로 보냅니다.
    // close()가 호출될 때 내부적으로 flush()가 호출되므로,
    // try-with-resources를 사용하면 명시적 flush() 호출이 필수는 아닙니다.
    // 하지만 중요한 데이터 쓰기 직후에는 명시적으로 호출하는 것이 안전할 수 있습니다.
    writer.flush();

    System.out.println("Log file written successfully using BufferedWriter: " + fileName);

} catch (IOException e) {
    System.err.println("Failed to write to file: " + e.getMessage());
    e.printStackTrace();
}
// try 블록을 벗어나는 순간, writer.close()가 자동으로 호출되어
// 남아있는 모든 버퍼 내용이 파일에 쓰이고 파일 핸들이 해제됩니다.
```

### `FileChannel`과 `ByteBuffer` (NIO)

성능에 매우 민감한 애플리케이션이나 바이너리 데이터를 직접 다룰 때 사용하는 저수준(low-level) 고성능 API입니다.

`FileChannel`은 운영체제의 파일 디스크립터에 직접 연결되는 통로 역할을 합니다.
데이터를 `ByteBuffer`라는 메모리 버퍼에 담아 채널을 통해 전송합니다.
JVM 힙 메모리가 아닌 네이티브 메모리(Direct Buffer)를 사용할 수 있어 데이터 복사 오버헤드를 줄일 수 있습니다.

- 커널 버퍼와의 상호작용:
    - OS의 페이지 캐시(Page Cache)와 더 효율적으로 상호작용합니다.
    - `transferTo()`, `transferFrom()` 같은 메서드는 커널 수준에서 데이터를 복사(Zero-Copy)합니다.
- 데이터 내구성 제어:
    - `force(boolean metaData)` 메서드를 통해 데이터가 물리적 디스크에 기록되는 시점을 명시적으로 제어할 수 있습니다.
    - 데이터베이스의 커밋 로그처럼 데이터의 내구성(Durability)이 중요한 경우 필수적입니다.

데이터베이스 시스템, 대용량 파일 복사/전송, 고성능 네트워킹 애플리케이션 등에 적합합니다.

단, 코드 복잡성이 증가할 수 있습니다.

```java
Path filePath = Paths.get("channel-output.bin");

String text = "Hello FileChannel! This is raw byte data.";
byte[] data = text.getBytes(StandardCharsets.UTF_8);

// FileChannel은 기존 I/O 스트림과 달리 바이트 채널을 직접 다룹니다.
// try-with-resources로 FileChannel 객체의 자동 닫힘을 보장합니다.
try (
    // FileChannel.open()을 통해 쓰기, 생성 옵션을 지정하여 채널을 엽니다.
    FileChannel fileChannel = FileChannel.open(filePath, StandardOpenOption.CREATE, StandardOpenOption.WRITE)
) {
    // ByteBuffer는 데이터를 담는 컨테이너입니다. JVM 힙 메모리에 할당됩니다.
    // (Direct Buffer를 사용하려면 ByteBuffer.allocateDirect()를 사용합니다.)
    ByteBuffer buffer = ByteBuffer.allocate(data.length);

    // 1. 데이터를 버퍼에 넣습니다. (쓰기 모드)
    buffer.put(data);

    // 2. 버퍼를 읽기 모드로 전환합니다.
    //    position을 0으로, limit을 현재 position으로 설정하여
    //    버퍼에 쓰인 데이터를 처음부터 읽을 수 있도록 준비합니다.
    buffer.flip();

    // 3. 버퍼의 데이터를 채널을 통해 파일에 씁니다.
    //    channel.write()는 버퍼의 현재 position부터 limit까지의 데이터를 씁니다.
    //    이 작업은 내부적으로 OS의 write() 시스템 콜을 호출합니다.
    while (buffer.hasRemaining()) {
        fileChannel.write(buffer);
    }

    // (선택 사항) 데이터 내구성을 보장해야 할 때 사용합니다.
    // OS의 페이지 캐시에만 저장된 데이터를 물리적 디스크에 강제로 동기화합니다.
    // 데이터베이스의 트랜잭션 커밋과 유사한 역할을 하며, 성능 저하를 유발할 수 있습니다.
    // 첫 번째 인자 true는 파일 데이터와 메타데이터(파일 크기 등) 모두를 동기화하라는 의미입니다.
    fileChannel.force(true);

    System.out.println("File written successfully using FileChannel: " + filePath.toAbsolutePath());

} catch (IOException e) {
    System.err.println("An error occurred while writing with FileChannel: " + e.getMessage());
    e.printStackTrace();
}
```

### `MappedByteBuffer`

파일의 특정 영역을 애플리케이션의 가상 메모리 주소 공간에 직접 매핑하는 가장 빠른 파일 I/O 방법입니다.

`FileChannel.map()`을 통해 파일을 메모리에 매핑하면, 파일은 거대한 바이트 배열처럼 취급됩니다.
데이터를 메모리에 쓰는 즉시 운영체제가 이를 감지하고, 백그라운드에서 비동기적으로 파일에 반영합니다.

- 시스템 콜 부재:
    - 전통적인 `read()`/`write()` 시스템 콜을 사용하지 않습니다.
    - 데이터는 사용자 공간과 커널 공간 사이를 복사할 필요 없이 직접 처리됩니다.
- OS 레벨 최적화:
    - 파일 I/O가 아닌 메모리 접근으로 처리되므로, OS의 가상 메모리 관리자(VMM)가 페이지 폴트를 통해 I/O를 최적화합니다.

매우 큰 파일에 대한 무작위 접근(random access)이 빈번한 경우, 고성능 데이터 분석, 인메모리 데이터베이스 구현 등에 적합합니다.

단, 데이터가 디스크에 쓰이는 시점을 OS가 결정하므로, `force()`를 명시적으로 호출하지 않으면 애플리케이션 충돌 시 데이터가 유실될 수 있습니다.
또한 자원 관리가 복잡합니다.

```java
int FILE_SIZE = 1024; // 1 KB
String fileName = "mapped-file.dat";

// MappedByteBuffer는 try-with-resources로 직접 관리할 수 없으므로,
// FileChannel과 RandomAccessFile을 명시적으로 닫아주어야 합니다.
try (
    // RandomAccessFile은 파일의 어느 위치에나 읽고 쓸 수 있게 해줍니다.
    // "rw" 모드는 읽기/쓰기 모드를 의미합니다.
    RandomAccessFile file = new RandomAccessFile(fileName, "rw");
    FileChannel channel = file.getChannel()
) {
    // 파일의 특정 영역을 메모리에 매핑합니다.
    // MapMode.READ_WRITE: 읽고 쓸 수 있도록 매핑합니다.
    // 0: 파일의 시작 위치부터 매핑합니다.
    // FILE_SIZE: 매핑할 크기입니다. 이 크기만큼 파일 공간이 미리 확보됩니다.
    MappedByteBuffer mappedBuffer = channel.map(FileChannel.MapMode.READ_WRITE, 0, FILE_SIZE);

    // 이제 mappedBuffer는 파일과 직접 연결된 메모리 영역처럼 동작합니다.
    // 시스템 콜(write) 없이 직접 메모리에 데이터를 씁니다.
    String message1 = "Writing directly to memory-mapped file. ";
    mappedBuffer.put(message1.getBytes(StandardCharsets.UTF_8));

    String message2 = "This is extremely fast.";
    mappedBuffer.put(message2.getBytes(StandardCharsets.UTF_8));

    System.out.println("Data written to memory-mapped buffer.");

    // (선택 사항) 메모리 버퍼의 변경 사항을 디스크에 강제로 기록합니다.
    // 이 호출이 없으면 OS가 비동기적으로 언젠가 디스크에 쓰지만,
    // 시스템 장애 시 데이터가 유실될 수 있습니다.
    mappedBuffer.force();

    System.out.println("Forced buffer content to be written to the storage device.");

} catch (IOException e) {
    System.err.println("An error occurred with Memory-Mapped File I/O: " + e.getMessage());
    e.printStackTrace();
}
```

## Java `File`과 `Path`

`File` 클래스는 Java 1.0 시절에 설계된 API로, 파일과 디렉토리 경로를 문자열(`String`) 기반으로 처리했고, Java의 NIO(New Input/Output)이 추가된 Java 7 이전에는 `Path`가 없었습니다.

Java 7부터 NIO 패키지가 추가되었고, `java.nio.file.Path` 클래스를 포함한 새로운 파일 시스템 API가 도입되었습니다. 이를 통해 더 유연한 방식으로 파일 경로를 관리합니다.

`File` 객체와 `Path` 객체는 서로 변환이 가능합니다:
- `Path`를 `File`로 변환: `path.toFile()`
- `File`을 `Path`로 변환: `file.toPath()`

`Path` 객체와 NIO API를 사용하는 것이 권장됩니다.

```java
Path filePath = Paths.get("example.txt");
Files.createFile(filePath); // 파일 생성
```

`File` 클래스는 여전히 `String` 기반으로 경로를 처리합니다.

```java
File file = new File("example.txt");
file.createNewFile(); // 파일 생성
```

## 예제

### create directory and file

```java
/**
 * @param fileName
 *
 * @return
 */
public Path getReportFilePath(String fileName) {
    return Paths.get("target", "reports", fileName);
}

public Optional<Path> createFile(Path filePath) {
    try {
        Files.createDirectories(filePath.getParent());
        Files.createFile(filePath);
        return Optional.of(filePath);
    } catch (IOException e) {
        // 예외 처리
        System.err.println("Error occurred while creating file: " + e.getMessage());
        e.printStackTrace(); // 로그 출력 또는 추가 처리
    }

    return Optional.empty();
}
```

이에 대한 테스트:

```java
@Test
@DisplayName("파일 생성 전 경로 얻기")
void test_generateFilePath() {
    var path = mrs.getReportFilePath("filename.xlsx");

    assertEquals(path.toString(), "target\\reports\\filename.xlsx");
    assertEquals(
        path.toAbsolutePath().toString(),
        "C:\\Users\\%USERNAME%\\path\\to\\project\\target\\reports\\filename.xlsx"
    );
}
```

---

# Java 17과 리눅스 서버에서의 고가용성 파일 처리: 완전한 기술적 분석

배치 펌뱅킹 시스템과 같은 금융 서비스에서 파일 처리는 매우 중요한 요소입니다. 데이터 무결성과 고가용성이 보장되어야 하는 환경에서, Java 17과 리눅스 서버를 기반으로 한 파일 처리 시스템의 동작 원리를 하드웨어부터 애플리케이션 레이어까지 상세히 분석해보겠습니다.

## 파일 쓰기 동작의 완전한 흐름

### 1. Java 애플리케이션 레이어에서의 파일 처리

Java에서 파일을 처리할 때, 우리가 흔히 사용하는 `Files.write()`나 `FileChannel.write()` 호출은 다음과 같은 복잡한 과정을 거칩니다:

```java
// 일반적인 파일 쓰기
Files.write(path, data, StandardOpenOption.CREATE, StandardOpenOption.APPEND);

// NIO를 이용한 방식
try (FileChannel channel = FileChannel.open(path, StandardOpenOption.WRITE)) {
    ByteBuffer buffer = ByteBuffer.allocateDirect(1024);
    buffer.put(data);
    buffer.flip();
    channel.write(buffer);
}
```

이때 핵심적으로 이해해야 할 부분은 **ByteBuffer의 종류**입니다:

**HeapByteBuffer**: JVM 힙 메모리에 생성되며, 실제 I/O 작업 시 네이티브 메모리로 복사가 필요합니다. 이 과정에서 추가적인 메모리 할당과 복사 오버헤드가 발생합니다.

**DirectByteBuffer**: JVM 힙 외부의 네이티브 메모리에 직접 할당되어, 커널과의 데이터 교환 시 복사 과정을 생략할 수 있습니다. 대용량 파일 처리나 빈번한 I/O 작업에서 성능상 이점이 있습니다.

**MappedByteBuffer**: 파일을 메모리에 직접 매핑하여, 가상 메모리 시스템을 통해 파일 접근을 메모리 접근처럼 처리할 수 있습니다.

### 2. JVM 레이어: 네이티브 호출과 메모리 관리

Java 파일 I/O는 결국 JVM의 네이티브 코드를 통해 운영체제의 시스템 콜을 호출합니다. 이 과정에서 중요한 것은 **가비지 컬렉션과 파일 디스크립터 관리**입니다:

```java
// 파일 디스크립터 누수 방지를 위한 올바른 패턴
try (FileChannel channel = FileChannel.open(path, StandardOpenOption.WRITE)) {
    // 파일 작업 수행
} // try-with-resources가 자동으로 채널을 닫아 디스크립터 해제
```

JVM은 Direct Memory를 별도로 관리하며, 이는 힙 메모리와는 독립적입니다. 대용량 파일 처리 시 `-XX:MaxDirectMemorySize` 설정을 통해 이 영역의 크기를 조정해야 합니다.

### 3. 운영체제 레이어: 시스템 콜과 페이지 캐시

JVM이 `write()` 시스템 콜을 호출하면, 리눅스 커널의 **Virtual File System(VFS)**이 이를 처리합니다. 여기서 핵심적인 동작들이 일어납니다:

**페이지 캐시 활용**: 기본적으로 모든 쓰기 작업은 먼저 운영체제의 페이지 캐시에 저장됩니다. 이는 성능 향상을 위한 것이지만, 시스템 충돌 시 데이터 손실 위험이 있습니다.

**파일 디스크립터 관리**: 각 열린 파일은 고유한 파일 디스크립터를 가지며, 프로세스당 제한이 있습니다. 펌뱅킹과 같은 대용량 처리 시스템에서는 이 제한을 고려해야 합니다.

### 4. 파일 시스템 레이어: ext4와 저널링

ext4 파일 시스템은 **저널링 메커니즘**을 통해 데이터 무결성을 보장합니다. 이는 세 가지 모드로 동작합니다:

**Ordered 모드** (기본값): 메타데이터만 저널에 기록하되, 데이터 블록을 먼저 쓴 후 메타데이터를 저널에 기록합니다.

**Journal 모드**: 데이터와 메타데이터 모두를 저널에 기록하여 최고 수준의 안전성을 제공하지만, 성능은 가장 낮습니다.

**Writeback 모드**: 메타데이터만 저널에 기록하고 데이터와의 순서를 보장하지 않아 가장 빠르지만 위험합니다.

### 5. 블록 레이어와 I/O 스케줄링

리눅스 커널의 **블록 레이어**는 디스크 접근을 최적화하기 위해 여러 I/O 스케줄러를 제공합니다:

**mq-deadline**: 읽기와 쓰기 요청을 분리하여 처리하며, 응답시간을 보장합니다. 대부분의 워크로드에 적합합니다.

**none (NOOP)**: 스케줄링을 수행하지 않고 FIFO 방식으로 처리합니다. NVMe SSD와 같은 고성능 스토리지에 적합합니다.

**bfq**: 대화형 작업에 최적화되어 있어 데스크톱 환경에 적합합니다.

### 6. 스토리지 디바이스: 물리적 쓰기와 내부 캐시

최종적으로 데이터는 물리적 스토리지 디바이스에 기록됩니다. 이때 **디스크 내부 캐시**와 **Write Barrier** 개념이 중요합니다:

**디스크 캐시**: 대부분의 스토리지 디바이스는 내부 캐시를 가지고 있어, 쓰기 완료 신호를 실제 물리적 쓰기 완료 전에 보낼 수 있습니다.

**Write Barrier**: 파일 시스템이 메타데이터의 순서를 보장하기 위해 사용하는 메커니즘으로, 성능을 다소 희생하여 데이터 무결성을 보장합니다.

## 동기화 메커니즘과 내구성 보장

### fsync()와 fdatasync()의 차이점

펌뱅킹 시스템에서 매우 중요한 것은 **데이터의 영속성 보장**입니다:

```java
// Java에서의 동기화
FileChannel channel = FileChannel.open(path, StandardOpenOption.WRITE);
channel.write(buffer);
channel.force(false); // fdatasync() 호출 - 데이터만 동기화
channel.force(true);  // fsync() 호출 - 데이터 + 메타데이터 동기화
```

**fdatasync()**: 파일 데이터만 스토리지에 강제로 쓰고, 파일 크기 변경과 같은 중요한 메타데이터만 동기화합니다.

**fsync()**: 파일 데이터와 모든 메타데이터(접근 시간, 수정 시간 등)를 스토리지에 강제로 씁니다.

성능상으로는 `fdatasync()`가 더 빠르지만, 완전한 무결성이 필요한 금융 시스템에서는 `fsync()`를 사용하는 것이 권장됩니다.

### O_DIRECT와 페이지 캐시 우회

특별한 경우에는 **O_DIRECT** 플래그를 사용하여 운영체제의 페이지 캐시를 우회할 수 있습니다:

```java
// Java에서는 직접 O_DIRECT를 사용할 수 없지만,
// 특정 구현에서는 DirectByteBuffer와 함께 유사한 효과를 얻을 수 있습니다
```

O_DIRECT는 다음과 같은 특징을 가집니다:
- 메모리 복사 오버헤드를 줄여 CPU 사용률을 낮춥니다
- 예측 가능한 성능을 제공하지만, 작은 I/O에서는 오히려 느릴 수 있습니다
- 메모리 정렬 요구사항이 있어 구현이 복잡합니다

## 스트리밍과 대용량 파일 처리

### 스트리밍의 본질

스트리밍은 **전체 파일을 메모리에 로드하지 않고 조각조각 처리하는 방식**입니다. 이는 특히 대용량 파일 처리에서 메모리 사용량을 최적화하는 핵심 기법입니다:

```java
// 스트리밍 방식의 파일 처리
try (InputStream input = Files.newInputStream(sourcePath);
     OutputStream output = Files.newOutputStream(targetPath)) {

    byte[] buffer = new byte[8192]; // 8KB 버퍼
    int bytesRead;
    while ((bytesRead = input.read(buffer)) != -1) {
        output.write(buffer, 0, bytesRead);
        // 필요 시 여기서 fsync() 호출
    }
}
```

### 디스크 단편화와 성능 영향

파일이 디스크에 저장될 때 **단편화(fragmentation)**가 발생할 수 있습니다. ext4는 다음과 같은 메커니즘으로 이를 최소화합니다:

**Extents**: 연속된 블록들을 하나의 단위로 관리하여 단편화를 줄입니다.

**Delayed allocation**: 실제 블록 할당을 지연시켜 더 큰 연속 공간을 확보할 수 있도록 합니다.

**Multi-block allocator**: 여러 블록을 한 번에 할당하여 단편화를 방지합니다.

### 병렬 처리와 블록 쓰기 순서

현대의 스토리지 시스템은 **여러 블록에 동시에 쓰기**를 수행할 수 있습니다. 하지만 이때 중요한 것은 **쓰기 순서 보장**입니다:

1. **Write Barrier**: 중요한 메타데이터가 먼저 기록되도록 순서를 강제합니다
2. **Command Queuing**: NVMe와 같은 현대 인터페이스는 여러 명령을 동시에 처리할 수 있습니다
3. **Atomic Write**: 파일 시스템 레벨에서 원자적 쓰기를 보장합니다

## 고가용성과 내결함성 패턴

### Write-Ahead Logging (WAL) 패턴

데이터베이스와 같은 시스템에서 사용하는 **WAL 패턴**을 파일 처리에도 적용할 수 있습니다:

```java
// WAL 패턴의 구현 예시
public class WALFileWriter {
    private final Path walPath;
    private final Path dataPath;

    public void writeWithWAL(byte[] data) throws IOException {
        // 1. 먼저 WAL에 기록
        writeToWAL(data);

        // 2. WAL 동기화
        syncWAL();

        // 3. 실제 데이터 파일에 기록
        writeToDataFile(data);

        // 4. 데이터 파일 동기화
        syncDataFile();

        // 5. WAL 항목 제거 표시
        markWALComplete();
    }
}
```

### 원자적 파일 쓰기 패턴

펌뱅킹과 같은 중요한 시스템에서는 **원자적 파일 쓰기**가 필수입니다:

```java
public class AtomicFileWriter {
    public void atomicWrite(Path targetPath, byte[] data) throws IOException {
        Path tempPath = targetPath.resolveSibling(
            targetPath.getFileName() + ".tmp." + UUID.randomUUID());

        try {
            // 1. 임시 파일에 쓰기
            Files.write(tempPath, data);

            // 2. 동기화
            try (FileChannel channel = FileChannel.open(tempPath,
                    StandardOpenOption.WRITE)) {
                channel.force(true);
            }

            // 3. 원자적 이동 (rename)
            Files.move(tempPath, targetPath, StandardCopyOption.ATOMIC_MOVE);

        } finally {
            Files.deleteIfExists(tempPath);
        }
    }
}
```

### 파일 잠금과 동시성 제어

여러 프로세스가 동시에 같은 파일에 접근할 때는 **파일 잠금**이 필요합니다:

```java
try (FileChannel channel = FileChannel.open(path,
        StandardOpenOption.WRITE, StandardOpenOption.CREATE)) {

    // 배타적 잠금 획득
    FileLock lock = channel.lock();
    try {
        // 임계 영역에서의 파일 작업
        channel.write(buffer);
        channel.force(true);
    } finally {
        lock.release();
    }
}
```

## 메모리 관리와 성능 최적화

### JVM 메모리 모델과 파일 I/O

Java 17에서 파일 I/O 성능을 최적화하려면 **JVM 메모리 구조**를 이해해야 합니다:

**Heap Memory**: 일반적인 Java 객체들이 저장되는 영역으로, GC의 영향을 받습니다.

**Direct Memory**: JVM 힙 외부 영역으로, 네이티브 I/O에 최적화되어 있습니다.

**Memory-Mapped Files**: 파일을 가상 메모리에 직접 매핑하여 빠른 접근을 가능하게 합니다.

### 대용량 파일 처리 최적화

펌뱅킹 시스템에서 대용량 파일을 처리할 때는 다음과 같은 최적화가 필요합니다:

```java
// 대용량 파일 처리 최적화 예시
public class OptimizedFileProcessor {
    private static final int BUFFER_SIZE = 1024 * 1024; // 1MB 버퍼

    public void processLargeFile(Path inputPath, Path outputPath)
            throws IOException {

        try (FileChannel inputChannel = FileChannel.open(inputPath,
                StandardOpenOption.READ);
             FileChannel outputChannel = FileChannel.open(outputPath,
                StandardOpenOption.WRITE, StandardOpenOption.CREATE)) {

            // Direct ByteBuffer 사용으로 복사 오버헤드 최소화
            ByteBuffer buffer = ByteBuffer.allocateDirect(BUFFER_SIZE);

            long position = 0;
            long totalSize = inputChannel.size();

            while (position  /sys/block/sda/queue/scheduler
```

### 모니터링과 성능 측정

파일 I/O 성능을 모니터링하기 위한 핵심 지표들:

```bash
# I/O 성능 모니터링
iostat -x 1
iotop -o
dstat -d

# 파일 시스템 상태 확인
tune2fs -l /dev/sda1
dumpe2fs /dev/sda1 | grep -A1 "Free blocks"

# JVM 메모리 사용량 확인
jstat -gc
jcmd  VM.native_memory summary
```

### 오류 처리와 복구 전략

견고한 파일 처리 시스템을 위한 오류 처리 패턴:

```java
public class RobustFileProcessor {
    private static final int MAX_RETRY_ATTEMPTS = 3;
    private static final Duration RETRY_DELAY = Duration.ofSeconds(1);

    public void safeFileWrite(Path path, byte[] data) throws IOException {
        int attempts = 0;
        IOException lastException = null;

        while (attempts < MAX_RETRY_ATTEMPTS) {
            try {
                atomicWrite(path, data);
                return; // 성공

            } catch (IOException e) {
                lastException = e;
                attempts++;

                if (attempts < MAX_RETRY_ATTEMPTS) {
                    try {
                        Thread.sleep(RETRY_DELAY.toMillis());
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        throw new IOException("Interrupted during retry", ie);
                    }
                }
            }
        }

        throw new IOException("Failed after " + MAX_RETRY_ATTEMPTS +
                             " attempts", lastException);
    }
}
```

## 결론

Java 17과 리눅스 서버를 기반으로 한 고가용성 파일 처리 시스템은 하드웨어부터 애플리케이션까지 모든 레이어에서의 이해와 최적화가 필요합니다. 특히 펌뱅킹과 같은 금융 시스템에서는 성능과 안전성의 균형을 맞추는 것이 핵심입니다.

핵심 원칙들:
1. **적절한 동기화 수준 선택**: 비즈니스 요구사항에 따라 fsync/fdatasync 사용
2. **메모리 관리 최적화**: DirectByteBuffer와 적절한 버퍼 크기 사용
3. **원자적 연산 보장**: 임시 파일과 rename을 활용한 안전한 쓰기
4. **모니터링과 복구**: 적절한 오류 처리와 재시도 로직 구현
5. **시스템 레벨 튜닝**: I/O 스케줄러와 파일 시스템 옵션 최적화

이러한 이해를 바탕으로 구현한 파일 처리 시스템은 높은 신뢰성과 성능을 동시에 제공할 수 있으며, 금융 서비스와 같은 미션 크리티컬한 환경에서도 안전하게 운영될 수 있습니다.

---

# Java 17과 리눅스 서버에서의 고가용성 파일 처리: 심층적 기술 분석

배치 펌뱅킹 시스템에서 파일 처리는 단순한 읽기/쓰기 작업이 아닙니다. 이는 하드웨어 수준의 물리적 저장장치부터 Java 애플리케이션 레이어까지 복잡한 시스템들의 조화로운 상호작용입니다. 이 분석에서는 이러한 모든 레이어가 어떻게 작동하고 연결되는지 완전히 이해하여, 견고하고 안전한 파일 처리 시스템을 구축할 수 있는 지식을 제공하겠습니다.

## 물리적 저장장치: 데이터가 실제로 저장되는 곳

### 섹터와 블록: 저장의 기본 단위

모든 파일 처리는 물리적 저장장치의 **섹터(sector)**에서 시작됩니다. 섹터는 하드디스크에서 데이터를 저장하는 최소 물리적 단위로, 전통적으로 512바이트였지만, 현대의 하드디스크와 SSD는 4096바이트(4KB)의 **어드밴스드 포맷(Advanced Format)**을 사용합니다.

하지만 여기서 중요한 것은 **물리적 섹터와 논리적 블록의 차이**입니다. 운영체제는 직접 섹터를 다루지 않고, 대신 블록 단위로 작업합니다. 리눅스 파일시스템에서 블록은 보통 4KB 크기이며, 이는 메모리 페이지 크기와 동일합니다. 이러한 정렬은 성능상의 이유로 매우 중요합니다.

### 디스크 기하학적 구조와 성능

현대 하드디스크는 복잡한 기하학적 구조를 가지고 있습니다. 디스크는 여러 개의 플래터(platter)로 구성되며, 각 플래터는 여러 개의 트랙(track)을 가집니다. 트랙은 다시 여러 개의 섹터로 나뉩니다.

**Surface Serpentine**과 **Hybrid Serpentine** 같은 현대적 섹터 배치 기법은 성능을 최적화하기 위해 연속된 논리적 블록 주소(LBA)를 물리적으로 가까운 위치에 배치합니다. 이는 파일이 연속적으로 저장될 때 헤드의 이동을 최소화하여 성능을 향상시킵니다.

## 리눅스 커널: 복잡한 I/O 시스템의 조율자

### 가상 파일 시스템(VFS): 통일된 인터페이스

리눅스 커널의 **가상 파일 시스템(VFS)**은 다양한 파일시스템들을 통일된 인터페이스로 제공합니다. VFS는 사용자 공간의 시스템 콜과 구체적인 파일시스템 구현 사이의 추상화 계층 역할을 합니다.

```c
// VFS의 핵심 구조체들
struct super_block {
    struct list_head s_list;
    dev_t s_dev;
    struct file_system_type *s_type;
    struct super_operations *s_op;
    // 파일시스템의 전역 정보
};

struct inode {
    umode_t i_mode;
    uid_t i_uid;
    gid_t i_gid;
    loff_t i_size;
    struct timespec i_atime, i_mtime, i_ctime;
    struct inode_operations *i_op;
    // 파일의 메타데이터
};

struct dentry {
    struct inode *d_inode;
    struct dentry *d_parent;
    struct qstr d_name;
    // 디렉토리 항목 정보
};
```

### 블록 I/O 계층: 성능 최적화의 핵심

리눅스 커널의 **블록 I/O 계층**은 파일시스템과 블록 디바이스 드라이버 사이에 위치하여 I/O 요청을 최적화합니다. 이 계층은 여러 단계로 구성됩니다:

1. **BIO(Block I/O) 구조체**: 개별 I/O 요청을 표현합니다
2. **REQUEST 구조체**: 여러 BIO를 하나로 합쳐 효율성을 높입니다
3. **I/O 스케줄러**: 디스크 성능을 최적화하기 위해 요청을 재정렬합니다

```c
// 블록 I/O의 핵심 구조
struct bio {
    sector_t bi_sector;           // 시작 섹터
    struct bio *bi_next;          // 다음 BIO
    struct block_device *bi_bdev; // 타겟 디바이스
    unsigned int bi_size;         // 크기
    struct bio_vec *bi_io_vec;    // 메모리 벡터
    bio_end_io_t *bi_end_io;     // 완료 콜백
};
```

### 멀티큐 블록 레이어(blk-mq): 현대적 병렬 처리

현대 SSD와 NVMe 디바이스의 병렬 처리 능력을 활용하기 위해 리눅스는 **멀티큐 블록 레이어(blk-mq)**를 도입했습니다. 이는 각 CPU 코어마다 별도의 I/O 큐를 생성하여 락 경합을 최소화하고 병렬성을 극대화합니다.

```c
// 멀티큐 구조
struct blk_mq_ctx {
    struct {
        struct list_head rq_list;
    } ____cacheline_aligned_in_smp;

    unsigned int cpu;
    unsigned int index_hw;
    struct blk_mq_hw_ctx *hctx;
};

struct blk_mq_hw_ctx {
    struct {
        struct list_head dispatch;
        struct sbitmap ctx_map;
    } ____cacheline_aligned_in_smp;

    unsigned int queue_num;
    atomic_t nr_active;
    struct blk_mq_tags *tags;
};
```

## 메모리 관리: 페이지 캐시와 버퍼 캐시

### 페이지 캐시: 파일 데이터의 메모리 저장소

리눅스의 **페이지 캐시**는 파일 I/O 성능을 향상시키는 핵심 컴포넌트입니다. 모든 파일 읽기와 쓰기는 페이지 캐시를 통해 이루어지며, 이는 다음과 같은 이점을 제공합니다:

1. **읽기 성능 향상**: 한 번 읽은 데이터는 메모리에 캐시되어 후속 읽기가 빨라집니다
2. **쓰기 성능 향상**: 쓰기 작업은 먼저 메모리에 수행되고, 나중에 디스크에 플러시됩니다
3. **선행 읽기(Readahead)**: 파일의 일부를 읽을 때 뒤따르는 부분을 미리 읽어 성능을 향상시킵니다

### 버퍼 캐시: 메타데이터 관리

**버퍼 캐시**는 파일시스템의 메타데이터(슈퍼블록, 아이노드, 디렉토리 블록 등)를 캐시합니다. 현대 리눅스에서는 **Unified Buffer Cache**를 사용하여 페이지 캐시와 버퍼 캐시를 통합 관리합니다.

```c
// 페이지 캐시 구조
struct address_space {
    struct inode *host;
    struct radix_tree_root page_tree;
    spinlock_t tree_lock;
    atomic_t i_mmap_writable;
    struct rb_root i_mmap;
    struct rw_semaphore i_mmap_rwsem;
    unsigned long nrpages;
    pgoff_t writeback_index;
    const struct address_space_operations *a_ops;
    unsigned long flags;
    gfp_t gfp_mask;
    struct backing_dev_info *backing_dev_info;
};
```

### 메모리 매핑된 I/O: 가상 메모리의 활용

**메모리 매핑된 I/O(Memory-Mapped I/O)**는 파일을 가상 메모리에 직접 매핑하여 메모리 접근을 통해 파일 I/O를 수행하는 기법입니다. 이는 다음과 같은 장점을 제공합니다:

1. **시스템 콜 오버헤드 감소**: 일단 매핑되면 메모리 접근만으로 파일 I/O가 가능합니다
2. **Copy 오버헤드 제거**: 데이터를 사용자 공간으로 복사할 필요가 없습니다
3. **페이지 폴트 기반 지연 로딩**: 실제 접근될 때만 페이지가 메모리에 로드됩니다

## ext4 파일시스템: 고성능 저널링 시스템

### Extent 기반 블록 할당

ext4는 **Extent** 기반 블록 할당을 사용하여 파일 단편화를 크게 줄입니다. Extent는 연속된 블록들의 범위를 나타내며, 하나의 Extent는 최대 128MB의 연속된 공간을 표현할 수 있습니다.

```c
// ext4 Extent 구조
struct ext4_extent {
    __le32 ee_block;   // 논리적 블록 번호
    __le16 ee_len;     // 블록 개수
    __le16 ee_start_hi; // 물리적 블록 번호 상위 16비트
    __le32 ee_start_lo; // 물리적 블록 번호 하위 32비트
};

struct ext4_extent_header {
    __le16 eh_magic;   // 매직 넘버
    __le16 eh_entries; // 유효한 항목 수
    __le16 eh_max;     // 최대 항목 수
    __le16 eh_depth;   // 트리 깊이
    __le32 eh_generation; // 생성 번호
};
```

### 멀티블록 할당자(Multi-block Allocator)

ext4의 **멀티블록 할당자**는 한 번에 여러 블록을 할당하여 단편화를 방지하고 성능을 향상시킵니다. 이는 두 가지 사전 할당 공간을 유지합니다:

1. **Per-inode 사전 할당**: 큰 파일을 위한 파일별 사전 할당 공간
2. **Per-CPU 지역성 그룹**: 작은 파일들을 가깝게 배치하기 위한 CPU별 공간

### 지연 할당(Delayed Allocation)

**지연 할당**은 실제 디스크 쓰기가 발생할 때까지 블록 할당을 지연시키는 기법입니다. 이는 다음과 같은 이점을 제공합니다:

1. **더 나은 할당 결정**: 전체 쓰기 패턴을 파악한 후 최적의 블록을 할당할 수 있습니다
2. **단편화 감소**: 연속된 블록들을 한 번에 할당할 수 있습니다
3. **불필요한 할당 방지**: 단기간 파일의 경우 할당 자체를 피할 수 있습니다

```c
// 지연 할당 프로세스
write_begin() {
    // 블록 할당 대신 예약만 수행
    ext4_da_reserve_space(inode, blocks);
    // 버퍼 헤드에 지연 할당 마크
    set_buffer_delay(bh);
}

writepage() {
    // 실제 블록 할당 수행
    ext4_da_map_blocks(inode, sector, &map);
    // 예약 해제
    ext4_da_release_space(inode, reserved_blocks);
}
```

### 저널링: 데이터 일관성 보장

ext4의 **저널링 시스템**은 시스템 크래시 시에도 파일시스템의 일관성을 보장합니다. 세 가지 저널링 모드를 지원합니다:

1. **ordered 모드** (기본값): 메타데이터만 저널에 기록하되, 데이터 블록을 먼저 쓴 후 메타데이터를 저널에 기록합니다
2. **journal 모드**: 데이터와 메타데이터 모두를 저널에 기록하여 최고 수준의 안전성을 제공합니다
3. **writeback 모드**: 메타데이터만 저널에 기록하고 데이터와의 순서를 보장하지 않습니다

```c
// 저널링 프로세스
journal_start(handle, credits);
// 메타데이터 변경
ext4_mark_inode_dirty(handle, inode);
ext4_mark_iloc_dirty(handle, inode, &iloc);
// 저널 커밋
journal_stop(handle);
```

## JVM과 Native 메소드: Java에서 시스템 호출까지

### JVM 아키텍처와 메모리 구조

Java 17에서 파일 I/O를 수행할 때, JVM은 복잡한 메모리 관리 시스템을 사용합니다. JVM의 메모리는 여러 영역으로 구분됩니다:

1. **힙 메모리**: 일반적인 Java 객체들이 저장되는 영역
2. **직접 메모리(Direct Memory)**: JVM 힙 외부 영역으로, 네이티브 I/O에 최적화
3. **메타스페이스(Metaspace)**: 클래스 메타데이터가 저장되는 영역

### ByteBuffer: 메모리 관리의 핵심

Java NIO의 **ByteBuffer**는 파일 I/O의 핵심 컴포넌트입니다. 세 가지 주요 유형이 있습니다:

1. **HeapByteBuffer**: JVM 힙 메모리에 할당되며, 가비지 컬렉션의 대상입니다
2. **DirectByteBuffer**: 힙 외부 메모리에 할당되어 네이티브 I/O에 최적화되어 있습니다
3. **MappedByteBuffer**: 파일을 메모리에 직접 매핑하여 빠른 접근을 가능하게 합니다

```java
// ByteBuffer 사용 예시
public class OptimizedFileIO {
    private static final int BUFFER_SIZE = 1024 * 1024; // 1MB

    public void efficientFileCopy(Path source, Path target) throws IOException {
        try (FileChannel sourceChannel = FileChannel.open(source, StandardOpenOption.READ);
             FileChannel targetChannel = FileChannel.open(target,
                 StandardOpenOption.WRITE, StandardOpenOption.CREATE)) {

            // DirectByteBuffer 사용으로 복사 오버헤드 최소화
            ByteBuffer buffer = ByteBuffer.allocateDirect(BUFFER_SIZE);

            while (sourceChannel.read(buffer) > 0) {
                buffer.flip();
                targetChannel.write(buffer);
                buffer.clear();
            }
        }
    }
}
```

### DirectByteBuffer의 성능 특성

**DirectByteBuffer**는 네이티브 메모리를 사용하여 커널과의 데이터 교환에서 복사 오버헤드를 줄입니다. 하지만 할당 비용이 매우 높습니다:

- HeapByteBuffer 대비 약 25배 느린 할당 속도
- 페이지 정렬 요구사항으로 인한 메모리 낭비
- 가비지 컬렉션 대상이 아니므로 수동 관리 필요

```java
// DirectByteBuffer 재사용 패턴
public class DirectBufferPool {
    private final Queue pool = new ConcurrentLinkedQueue<>();
    private final int bufferSize;

    public DirectBufferPool(int bufferSize) {
        this.bufferSize = bufferSize;
    }

    public ByteBuffer acquire() {
        ByteBuffer buffer = pool.poll();
        if (buffer == null) {
            buffer = ByteBuffer.allocateDirect(bufferSize);
        }
        buffer.clear();
        return buffer;
    }

    public void release(ByteBuffer buffer) {
        if (buffer.isDirect() && buffer.capacity() == bufferSize) {
            pool.offer(buffer);
        }
    }
}
```

### JNI와 Native 메소드 호출

Java의 파일 I/O는 궁극적으로 **JNI(Java Native Interface)**를 통해 운영체제의 시스템 콜을 호출합니다. 이 과정은 다음과 같습니다:

1. **Java 메소드 호출**: `FileChannel.write()` 등의 Java API 호출
2. **JNI 바인딩**: 네이티브 메소드로 연결
3. **시스템 콜 호출**: 운영체제의 `write()`, `read()` 등의 시스템 콜 실행
4. **결과 반환**: 네이티브 코드에서 Java로 결과 반환

```java
// FileInputStream의 네이티브 메소드 예시
public class FileInputStream extends InputStream {
    private native int readBytes(byte b[], int off, int len) throws IOException;

    public int read(byte b[], int off, int len) throws IOException {
        return readBytes(b, off, len);
    }
}
```

## 동기화와 데이터 일관성: 안전한 파일 처리

### fsync와 fdatasync: 영속성 보장

**fsync()**와 **fdatasync()**는 데이터의 영속성을 보장하는 핵심 시스템 콜입니다. 둘 사이의 차이점은 다음과 같습니다:

- **fsync()**: 파일 데이터와 모든 메타데이터(접근 시간, 수정 시간 등)를 디스크에 강제로 씁니다
- **fdatasync()**: 파일 데이터와 중요한 메타데이터만 동기화합니다

```java
// Java에서의 동기화
public class SafeFileWriter {
    public void writeWithSync(Path path, byte[] data, boolean metadata)
            throws IOException {
        try (FileChannel channel = FileChannel.open(path,
                StandardOpenOption.WRITE, StandardOpenOption.CREATE)) {

            ByteBuffer buffer = ByteBuffer.wrap(data);
            channel.write(buffer);

            // 동기화 수행
            channel.force(metadata); // metadata=true면 fsync, false면 fdatasync
        }
    }
}
```

### Write Barrier: 순서 보장

**Write Barrier**는 메모리와 디스크 쓰기 순서를 보장하는 메커니즘입니다. 리눅스 커널에서는 저널링 파일시스템의 메타데이터 일관성을 보장하기 위해 사용됩니다.

```c
// 리눅스 커널의 write barrier 구현
void blkdev_issue_flush(struct block_device *bdev, gfp_t gfp_mask,
                       sector_t *error_sector) {
    struct request_queue *q;
    struct bio *bio;
    int ret = 0;

    if (bdev->bd_disk == NULL)
        return -ENXIO;

    q = bdev_get_queue(bdev);
    if (!q)
        return -ENXIO;

    bio = bio_alloc(gfp_mask, 0);
    bio->bi_bdev = bdev;
    bio->bi_flags |= 1  /sys/block/sda/queue/scheduler
```

### 스케줄러 선택 기준

펌뱅킹 시스템에서는 다음과 같은 기준으로 스케줄러를 선택해야 합니다:

- **HDD 사용 시**: mq-deadline 또는 bfq를 사용하여 헤드 이동을 최소화
- **SSD 사용 시**: none 또는 mq-deadline을 사용하여 스케줄링 오버헤드 최소화
- **혼합 워크로드**: mq-deadline을 사용하여 균형잡힌 성능 제공

## 메모리 관리와 페이지 폴트

### 페이지 폴트 처리 메커니즘

**페이지 폴트**는 프로세스가 현재 메모리에 로드되지 않은 페이지에 접근할 때 발생합니다. 이는 다음과 같은 과정으로 처리됩니다:

1. **MMU가 페이지 폴트 감지**: 접근하려는 페이지가 메모리에 없음을 감지
2. **커널 페이지 폴트 핸들러 호출**: 인터럽트가 발생하여 커널이 제어권을 가짐
3. **페이지 할당 및 로딩**: 필요한 페이지를 디스크에서 메모리로 로드
4. **페이지 테이블 업데이트**: 가상 주소와 물리 주소의 매핑 정보 업데이트

```c
// 페이지 폴트 핸들러 구조
static int __do_page_fault(struct mm_struct *mm, unsigned long addr,
                          unsigned int mm_flags, unsigned long vm_flags) {
    struct vm_area_struct *vma;
    int fault;

    vma = find_vma(mm, addr);
    if (unlikely(!vma)) {
        return VM_FAULT_BADMAP;
    }

    if (unlikely(vma->vm_start > addr)) {
        if (unlikely(!(vma->vm_flags & VM_GROWSDOWN))) {
            return VM_FAULT_BADMAP;
        }
    }

    fault = handle_mm_fault(vma, addr, mm_flags);
    return fault;
}
```

### 메모리 매핑과 성능

**메모리 매핑된 파일**은 페이지 폴트 메커니즘을 활용하여 필요한 부분만 메모리에 로드합니다. 이는 다음과 같은 장점을 제공합니다:

- **지연 로딩**: 실제 접근되는 부분만 메모리에 로드하여 메모리 사용량 최적화
- **자동 페이지 아웃**: 메모리 부족 시 자동으로 페이지가 스왑 아웃됨
- **여러 프로세스 간 공유**: 동일한 파일을 여러 프로세스가 효율적으로 공유

```java
// 메모리 매핑된 파일 사용
public class MemoryMappedFileProcessor {
    public void processLargeFile(Path filePath) throws IOException {
        try (FileChannel channel = FileChannel.open(filePath,
                StandardOpenOption.READ, StandardOpenOption.WRITE)) {

            long fileSize = channel.size();
            MappedByteBuffer buffer = channel.map(
                FileChannel.MapMode.READ_WRITE, 0, fileSize);

            // 메모리 접근을 통한 파일 처리
            while (buffer.hasRemaining()) {
                byte value = buffer.get();
                // 처리 로직
            }

            // 변경사항 강제 동기화
            buffer.force();
        }
    }
}
```

## 실제 구현: 고가용성 파일 처리 시스템

### 완전한 트랜잭션 파일 시스템

다음은 앞서 설명한 모든 개념을 통합한 완전한 파일 처리 시스템입니다:

```java
public class HighAvailabilityFileProcessor {
    private final Path dataDir;
    private final Path walDir;
    private final DirectBufferPool bufferPool;
    private final AtomicLong transactionId;
    private final ConcurrentHashMap activeTransactions;

    public HighAvailabilityFileProcessor(Path dataDir, Path walDir) {
        this.dataDir = dataDir;
        this.walDir = walDir;
        this.bufferPool = new DirectBufferPool(64 * 1024); // 64KB 버퍼
        this.transactionId = new AtomicLong(0);
        this.activeTransactions = new ConcurrentHashMap<>();
    }

    public long beginTransaction() {
        long txId = transactionId.incrementAndGet();
        Transaction tx = new Transaction(txId);
        activeTransactions.put(txId, tx);
        return txId;
    }

    public void writeFile(long txId, String filename, byte[] data) throws IOException {
        Transaction tx = activeTransactions.get(txId);
        if (tx == null) {
            throw new IllegalStateException("Transaction not found: " + txId);
        }

        // 1. WAL 레코드 생성
        WALRecord walRecord = new WALRecord(txId, filename, data);

        // 2. WAL에 기록
        writeWALRecord(walRecord);

        // 3. 트랜잭션에 기록 추가
        tx.addRecord(walRecord);
    }

    public void commitTransaction(long txId) throws IOException {
        Transaction tx = activeTransactions.remove(txId);
        if (tx == null) {
            throw new IllegalStateException("Transaction not found: " + txId);
        }

        try {
            // 1. WAL 커밋 레코드 쓰기
            writeWALCommit(txId);

            // 2. 실제 데이터 파일에 쓰기
            for (WALRecord record : tx.getRecords()) {
                writeDataFile(record.getFilename(), record.getData());
            }

            // 3. WAL 완료 표시
            markWALComplete(txId);

        } catch (IOException e) {
            // 롤백 처리
            rollbackTransaction(txId);
            throw e;
        }
    }

    public void rollbackTransaction(long txId) {
        Transaction tx = activeTransactions.remove(txId);
        if (tx != null) {
            // WAL에서 트랜잭션 롤백 기록
            try {
                writeWALRollback(txId);
            } catch (IOException e) {
                // 로그 기록 실패 시 처리
                logger.error("Failed to write rollback record for transaction " + txId, e);
            }
        }
    }

    private void writeWALRecord(WALRecord record) throws IOException {
        ByteBuffer buffer = bufferPool.acquire();
        try {
            // WAL 레코드 직렬화
            record.serialize(buffer);

            // WAL 파일에 쓰기
            try (FileChannel walChannel = FileChannel.open(
                    walDir.resolve("transaction.wal"),
                    StandardOpenOption.CREATE,
                    StandardOpenOption.WRITE,
                    StandardOpenOption.APPEND)) {

                walChannel.write(buffer);
                walChannel.force(true); // 즉시 동기화
            }
        } finally {
            bufferPool.release(buffer);
        }
    }

    private void writeDataFile(String filename, byte[] data) throws IOException {
        Path dataPath = dataDir.resolve(filename);
        Path tempPath = dataDir.resolve(filename + ".tmp." + System.nanoTime());

        try {
            // 임시 파일에 쓰기
            Files.write(tempPath, data, StandardOpenOption.CREATE);

            // 동기화
            try (FileChannel channel = FileChannel.open(tempPath,
                    StandardOpenOption.WRITE)) {
                channel.force(true);
            }

            // 원자적 이동
            Files.move(tempPath, dataPath, StandardCopyOption.ATOMIC_MOVE);

        } finally {
            Files.deleteIfExists(tempPath);
        }
    }

    // 복구 프로세스
    public void recover() throws IOException {
        // WAL 파일 스캔
        try (FileChannel walChannel = FileChannel.open(
                walDir.resolve("transaction.wal"),
                StandardOpenOption.READ)) {

            ByteBuffer buffer = ByteBuffer.allocate(64 * 1024);
            Map> transactions = new HashMap<>();
            Set committedTransactions = new HashSet<>();

            // WAL 레코드 읽기
            while (walChannel.read(buffer) > 0) {
                buffer.flip();

                while (buffer.hasRemaining()) {
                    WALRecord record = WALRecord.deserialize(buffer);

                    if (record.getType() == WALRecord.Type.DATA) {
                        transactions.computeIfAbsent(record.getTransactionId(),
                            k -> new ArrayList<>()).add(record);
                    } else if (record.getType() == WALRecord.Type.COMMIT) {
                        committedTransactions.add(record.getTransactionId());
                    }
                }

                buffer.clear();
            }

            // 커밋된 트랜잭션 재적용
            for (Long txId : committedTransactions) {
                List records = transactions.get(txId);
                if (records != null) {
                    for (WALRecord record : records) {
                        writeDataFile(record.getFilename(), record.getData());
                    }
                }
            }
        }
    }

    private static class Transaction {
        private final long id;
        private final List records;
        private final long startTime;

        public Transaction(long id) {
            this.id = id;
            this.records = new ArrayList<>();
            this.startTime = System.currentTimeMillis();
        }

        public void addRecord(WALRecord record) {
            records.add(record);
        }

        public List getRecords() {
            return Collections.unmodifiableList(records);
        }

        public long getId() {
            return id;
        }

        public long getStartTime() {
            return startTime;
        }
    }
}
```

### 성능 모니터링과 최적화

```java
public class FileIOMonitor {
    private final MeterRegistry meterRegistry;
    private final Timer writeTimer;
    private final Timer readTimer;
    private final Counter errorCounter;
    private final Gauge activeTransactions;

    public FileIOMonitor(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
        this.writeTimer = Timer.builder("file.write.duration")
            .description("File write operation duration")
            .register(meterRegistry);
        this.readTimer = Timer.builder("file.read.duration")
            .description("File read operation duration")
            .register(meterRegistry);
        this.errorCounter = Counter.builder("file.operation.errors")
            .description("File operation errors")
            .register(meterRegistry);
        this.activeTransactions = Gauge.builder("file.transactions.active")
            .description("Number of active transactions")
            .register(meterRegistry, this, FileIOMonitor::getActiveTransactionCount);
    }

    public  T monitorWrite(String operation, Supplier supplier) {
        return Timer.Sample.start(meterRegistry)
            .stop(writeTimer.tag("operation", operation))
            .recordCallable(() -> {
                try {
                    return supplier.get();
                } catch (Exception e) {
                    errorCounter.increment(Tags.of("operation", operation, "error", e.getClass().getSimpleName()));
                    throw e;
                }
            });
    }

    private double getActiveTransactionCount() {
        // 활성 트랜잭션 수 반환
        return activeTransactions.size();
    }
}
```

## 결론과 실무 권장사항

### 시스템 설정 최적화

펌뱅킹 시스템을 위한 최적화된 시스템 설정:

```bash
# JVM 옵션
-XX:+UseG1GC                           # G1 가비지 컬렉터 사용
-XX:MaxGCPauseMillis=100               # 최대 GC 일시정지 시간
-XX:MaxDirectMemorySize=4G             # 직접 메모리 크기
-XX:+UnlockExperimentalVMOptions       # 실험적 옵션 활성화
-XX:+UseLargePages                     # 대용량 페이지 사용

# 파일시스템 마운트 옵션
mount -o data=ordered,barrier=1,noatime /dev/sda1 /data

# I/O 스케줄러 설정
echo mq-deadline > /sys/block/sda/queue/scheduler

# 커널 파라미터 튜닝
echo 'vm.dirty_ratio = 5' >> /etc/sysctl.conf
echo 'vm.dirty_background_ratio = 2' >> /etc/sysctl.conf
echo 'vm.dirty_expire_centisecs = 3000' >> /etc/sysctl.conf
echo 'vm.dirty_writeback_centisecs = 100' >> /etc/sysctl.conf
```

### 모니터링 지표

중요한 모니터링 지표들:

```java
// 핵심 성능 지표
public class PerformanceMetrics {
    // I/O 관련 지표
    private final Timer fileWriteLatency;
    private final Timer fileReadLatency;
    private final Counter ioErrors;
    private final Gauge diskUtilization;

    // 메모리 관련 지표
    private final Gauge heapUsage;
    private final Gauge directMemoryUsage;
    private final Counter pageFaults;

    // 트랜잭션 관련 지표
    private final Timer transactionDuration;
    private final Counter transactionCommits;
    private final Counter transactionRollbacks;

    // 파일시스템 관련 지표
    private final Gauge availableDiskSpace;
    private final Counter fsyncCalls;
    private final Timer fsyncLatency;
}
```

### 오류 처리와 복구 전략

```java
public class ErrorHandlingStrategy {
    private static final int MAX_RETRY_ATTEMPTS = 3;
    private static final Duration RETRY_DELAY = Duration.ofSeconds(1);

    public  T executeWithRetry(Supplier operation, String operationName) {
        int attempts = 0;
        Exception lastException = null;

        while (attempts < MAX_RETRY_ATTEMPTS) {
            try {
                return operation.get();
            } catch (IOException e) {
                lastException = e;
                attempts++;

                if (attempts < MAX_RETRY_ATTEMPTS) {
                    try {
                        Thread.sleep(RETRY_DELAY.toMillis());
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        throw new RuntimeException("Interrupted during retry", ie);
                    }
                }
            }
        }

        throw new RuntimeException("Operation failed after " + MAX_RETRY_ATTEMPTS +
                                 " attempts: " + operationName, lastException);
    }
}
```

이 완전한 분석을 통해 Java 17과 리눅스 서버에서의 파일 처리가 단순한 읽기/쓰기 작업이 아니라, 하드웨어부터 애플리케이션까지 모든 레이어의 정교한 협력이라는 것을 이해할 수 있습니다. 각 레이어의 동작 원리와 최적화 기법을 이해함으로써, 펌뱅킹과 같은 미션 크리티컬한 시스템에서도 안전하고 효율적인 파일 처리 시스템을 구축할 수 있습니다.
