# Java file

- [Java file](#java-file)
    - [시스템 콜과 버퍼링](#시스템-콜과-버퍼링)
    - [효율적인 파일 생성 및 쓰기 방법](#효율적인-파일-생성-및-쓰기-방법)
        - [`Files.write()` / `Files.writeString()` (NIO)](#fileswrite--fileswritestring-nio)
        - [`BufferedWriter`](#bufferedwriter)
        - [`FileChannel`과 `ByteBuffer` (NIO)](#filechannel과-bytebuffer-nio)
        - [`MappedByteBuffer`](#mappedbytebuffer)
    - [Java `File`과 `Path`](#java-file과-path)
    - [예제](#예제)
        - [create directory and file](#create-directory-and-file)

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
    //   - CREATE: 파일이 존재하지 않으면 새로 생성합니다[2].
    //   - TRUNCATE_EXISTING: 파일이 존재하면 기존 내용을 모두 지우고 새로 씁니다. (덮어쓰기)
    //   - APPEND: 파일이 존재하면 기존 내용의 끝에 데이터를 추가합니다[2].
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
    // try-with-resources를 사용하면 명시적 flush() 호출이 필수는 아닙니다[3].
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
    // FileChannel.open()을 통해 쓰기, 생성 옵션을 지정하여 채널을 엽니다[4].
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
    // MapMode.READ_WRITE: 읽고 쓸 수 있도록 매핑합니다[5].
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
