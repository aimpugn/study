# Lesson 01 Line-By-Line Walkthrough

이 문서는 첫 강의의 코드를 **위에서 아래로 따라가며 줄 단위로 설명**하기 위한 companion walkthrough입니다.
소스 파일에는 핵심 의도와 비교 포인트를 짧게 남기고, 여기서는 그 코드를 처음 읽는 사람도 놓치지 않도록 한 줄씩 풀어 설명합니다.

빈 줄은 의미를 가진 실행 줄이 아니라 구조를 끊는 호흡이므로 따로 설명하지 않습니다.
대신 한 줄로 떼어 설명하면 오히려 이해가 나빠지는 곳은 `작은 contiguous 묶음`으로 다룹니다.

## [BlockingEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/blocking/BlockingEchoServer.java)

### 패키지, import, 클래스 선언

1. `1` `package io.aimpugn.learn.netty.lesson1.blocking;`
   이 파일이 첫 강의의 `blocking` 구현이라는 사실을 패키지 단계에서 먼저 고정합니다.
2. `3-5` `Lesson1Support`, `LessonServer`, `ObservationLog` import
   세 구현이 같은 실행 틀과 같은 로그 형식을 공유하도록 하는 공통 수단을 가져옵니다.
3. `7-19` Java 표준 라이브러리 import
   blocking 서버가 실제로 기대는 재료가 무엇인지 보여 줍니다. 소켓, 문자 스트림, thread pool, 원자 카운터가 핵심입니다.
4. `21-27` 클래스 KDoc
   이 구현의 학습 질문을 먼저 밝힙니다. "연결마다 기다리는 모델"이 무엇인지, 그리고 thread 수가 함께 늘어난다는 관측 포인트가 여기 들어 있습니다.
5. `28` `public final class BlockingEchoServer implements LessonServer {`
   첫 강의의 blocking 서버 구현을 시작하고, 공통 계약은 `LessonServer`로 맞춥니다.

### 필드와 생성자

6. `30` `private final ObservationLog observationLog = new ObservationLog("blocking");`
   이 서버가 남기는 모든 로그에 `blocking`이라는 이름표를 붙여, 다른 구현과 섞여도 바로 구분되게 합니다.
7. `31-32` `connectionIds` 앞 주석
   왜 이 카운터가 필요한지 먼저 말합니다. 로그에서 같은 연결을 계속 따라가기 위한 번호표라는 뜻입니다.
8. `33` `private final AtomicInteger connectionIds = new AtomicInteger();`
   새 연결이 들어올 때마다 `client-1`, `client-2`처럼 번호를 안정적으로 올리기 위한 원자 카운터입니다.
9. `34-35` `workerThreadIds` 앞 주석
   blocking 모델의 핵심 차이가 "연결마다 worker thread가 붙는다"는 점이라는 사실을 먼저 박아 둡니다.
10. `36` `private final AtomicInteger workerThreadIds = new AtomicInteger();`
    새 worker thread 이름에 번호를 붙여, 연결 수가 늘 때 thread도 늘어난다는 사실을 로그에서 바로 보이게 합니다.
11. `37` `private final int requestedPort;`
    사용자가 원하는 포트 번호를 잠깐 들고 있다가 실제 바인딩 때 씁니다.
12. `38` `private final ExecutorService connectionExecutor = Executors.newCachedThreadPool(task ->`
    연결마다 일을 넘길 executor를 만들기 시작합니다. `cachedThreadPool`을 고른 이유는 blocking 예제에서 worker가 연결 수에 따라 늘어나는 모습을 숨기지 않기 위해서입니다.
13. `39` `Thread.ofPlatform()`
    virtual thread가 아니라 platform thread를 써서, 운영체제 thread가 실제로 늘어나는 감각을 더 직접적으로 보여 줍니다.
14. `40` `.name("blocking-client-" + workerThreadIds.incrementAndGet())`
    worker thread 이름에 증가 번호를 붙여 `blocking-client-1`처럼 읽히게 만듭니다.
15. `41` `.unstarted(task));`
    executor가 필요할 때 시작할 thread 객체를 반환하고 필드 초기화를 닫습니다.
16. `43` `private volatile boolean running;`
    accept loop를 멈출 때 다른 thread에서도 변경을 바로 볼 수 있게 `volatile` 플래그를 둡니다.
17. `44` `private ServerSocket serverSocket;`
    실제로 새 연결을 받는 서버 소켓을 나중에 초기화하려고 필드로 둡니다.
18. `45` `private Thread acceptThread;`
    accept만 전담하는 thread를 보관해 두었다가 종료 시 join하기 위해 둡니다.
19. `47` `private BlockingEchoServer(int requestedPort) {`
    외부에서 바로 `new` 하지 못하게 생성자를 숨기고, 시작 로직은 `start(...)` 팩토리로 몰아 둡니다.
20. `48` `this.requestedPort = requestedPort;`
    생성자 입력으로 받은 포트 번호를 필드에 저장합니다.
21. `49` `}`
    생성자를 닫습니다.

### 시작 경로

22. `51` `public static BlockingEchoServer start(int port) throws IOException {`
    외부에서 이 서버를 시작할 때 들어오는 공개 진입점입니다.
23. `52` `BlockingEchoServer server = new BlockingEchoServer(port);`
    아직 소켓을 열지 않은 객체를 먼저 만듭니다.
24. `53` `server.startInternal();`
    실제 바인딩과 thread 시작은 내부 메서드로 넘깁니다.
25. `54` `return server;`
    준비가 끝난 서버 객체를 호출자에게 돌려줍니다.
26. `55` `}`
    `start(...)`를 닫습니다.
27. `57` `private void startInternal() throws IOException {`
    서버 내부 시작 절차를 정의합니다.
28. `58` `serverSocket = new ServerSocket(requestedPort);`
    지정된 포트에 서버 소켓을 열어 실제 listen을 시작합니다.
29. `59` `running = true;`
    accept loop가 돌 수 있게 실행 플래그를 올립니다.
30. `60` `observationLog.event("server-start", "server", "bound to port=" + serverSocket.getLocalPort());`
    서버가 실제 어떤 포트에 묶였는지 로그로 남깁니다. 테스트에서 `0` 포트를 쓰면 여기서 실제 포트가 드러납니다.
31. `61` accept thread 설명 주석 첫 줄
    왜 accept를 별도 thread로 떼는지 먼저 말합니다.
32. `62` `acceptThread = Thread.ofPlatform().name("blocking-accept").start(this::acceptLoop);`
    `blocking-accept`라는 이름의 platform thread를 만들어 새 연결 수락만 전담하게 합니다.
33. `63` `}`
    시작 절차를 닫습니다.

### accept loop

34. `65` `private void acceptLoop() {`
    새 연결을 계속 받는 루프 메서드를 엽니다.
35. `66` `while (running) {`
    종료 전까지 accept를 반복하겠다는 뜻입니다.
36. `67` `try {`
    소켓 수락 중 생길 수 있는 예외를 한곳에서 처리하려고 보호 구역을 엽니다.
37. `68` `Socket socket = serverSocket.accept();`
    여기서 accept thread가 실제로 block됩니다. 새 연결이 들어오기 전까지 이 줄에서 기다립니다.
38. `69` `int connectionId = connectionIds.incrementAndGet();`
    막 들어온 연결에 새 번호를 하나 할당합니다.
39. `70` `String channel = channelName(connectionId);`
    로그에서 읽기 쉬운 `client-N` 형태의 이름을 만듭니다.
40. `71` `observationLog.event("accept", channel, "accepted " + socket.getRemoteSocketAddress());`
    어느 원격 주소에서 연결이 들어왔는지 accept 시점 로그를 찍습니다.
41. `72-73` worker 설명 주석
    blocking 모델의 핵심이 바로 다음 줄이라는 점을 강조합니다.
42. `74` `connectionExecutor.submit(() -> handleConnection(socket, connectionId));`
    새 연결 하나를 전용 worker에게 넘기고, 그 worker가 이후 read/write를 독점하게 만듭니다.
43. `75` `} catch (SocketException socketException) {`
    서버 소켓이 닫힐 때 흔히 나오는 소켓 예외를 따로 받습니다.
44. `76` `if (running) {`
    종료 중이 아니라 진짜 오류일 때만 에러 로그를 남기겠다는 분기입니다.
45. `77` `observationLog.event("accept-error", "server", socketException.toString());`
    예기치 못한 accept 오류를 기록합니다.
46. `78` `}`
    `running` 검사 분기를 닫습니다.
47. `79` `return;`
    소켓 예외가 나오면 accept loop를 끝냅니다.
48. `80` `} catch (IOException ioException) {`
    일반 입출력 예외도 따로 받습니다.
49. `81` `observationLog.event("accept-error", "server", ioException.toString());`
    일반 입출력 오류를 기록합니다.
50. `82` `}`
    `IOException` 처리 블록을 닫습니다.
51. `83` `}`
    `while (running)` 루프를 닫습니다.
52. `84` `}`
    `acceptLoop()`를 닫습니다.

### 연결 처리

53. `86` `private void handleConnection(Socket socket, int connectionId) {`
    worker thread가 실제 한 연결을 오래 붙잡고 처리하는 메서드를 엽니다.
54. `87` `String channel = channelName(connectionId);`
    이 연결에 대응하는 로그용 이름을 다시 만듭니다.
55. `89-90` try-with-resources 설명 주석
    socket과 그 위 래퍼들을 한 scope 안에서 같이 닫는 이유를 먼저 설명합니다.
56. `91` `try (socket;`
    worker가 끝날 때 원본 소켓부터 자동으로 닫겠다는 선언입니다.
57. `92` `BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), StandardCharsets.UTF_8));`
    소켓 입력 바이트를 UTF-8 문자 줄로 읽기 쉽게 감싼 reader를 만듭니다.
58. `93` `BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), StandardCharsets.UTF_8))) {`
    소켓 출력 바이트를 UTF-8 문자 줄로 쓰기 쉽게 감싼 writer를 만들고 try 헤더를 닫습니다.
59. `95` `observationLog.event("connection-start", channel, "worker attached");`
    이 연결에 worker가 실제로 붙었다는 사실을 로그로 남깁니다.
60. `97` `String line;`
    `readLine()` 결과를 담을 변수를 미리 선언합니다.
61. `98-99` loop 설명 주석
    blocking 모델과 NIO/Netty 모델의 가장 직접적인 차이가 바로 다음 줄이라는 사실을 정리합니다.
62. `100` `while ((line = reader.readLine()) != null) {`
    worker thread가 여기서 줄이 올 때까지 block되고, 한 줄이 오면 루프 본문으로 내려갑니다.
63. `101` `observationLog.event("read", channel, "line=\"" + ObservationLog.previewText(line) + "\"");`
    읽은 줄을 로그에 남깁니다. 너무 긴 텍스트는 preview로 잘라 가독성을 지킵니다.
64. `102` `writer.write(line);`
    입력으로 받은 문자열 본문을 그대로 씁니다.
65. `103` `writer.newLine();`
    line protocol을 유지하려고 줄바꿈을 다시 붙입니다.
66. `104` `writer.flush();`
    버퍼에만 남기지 않고 즉시 소켓으로 내보냅니다.
67. `105` `observationLog.event("write", channel, "echoed line=\"" + ObservationLog.previewText(line) + "\"");`
    echo 응답을 보냈다는 로그를 남깁니다.
68. `106` `}`
    한 줄 처리 루프를 닫습니다.
69. `108` `observationLog.event("client-close", channel, "peer closed connection");`
    상대가 연결을 정상 종료했음을 기록합니다.
70. `109` `} catch (IOException ioException) {`
    읽기/쓰기 중 생기는 입출력 예외를 받습니다.
71. `110` `observationLog.event("connection-error", channel, ioException.toString());`
    연결 처리 중 오류를 기록합니다.
72. `111` `}`
    예외 처리 블록을 닫습니다.
73. `112` `}`
    `handleConnection(...)`을 닫습니다.

### 공통 계약과 종료

74. `114-117` `serverName()`
    이 구현의 공통 이름을 `blocking`으로 돌려줍니다.
75. `119-122` `port()`
    실제 바인딩된 서버 포트를 돌려줍니다.
76. `124` `@Override`
    공통 종료 계약 구현이라는 뜻을 다시 밝힙니다.
77. `125` `public void close() throws Exception {`
    서버 정리 절차를 엽니다.
78. `126` `running = false;`
    accept loop가 더 이상 새 연결을 받지 않게 합니다.
79. `128` `if (serverSocket != null && !serverSocket.isClosed()) {`
    서버 소켓이 실제로 열려 있을 때만 닫으려는 방어 분기입니다.
80. `129` `serverSocket.close();`
    accept를 깨우기 위해 서버 소켓을 닫습니다.
81. `130` `}`
    서버 소켓 정리 분기를 닫습니다.
82. `131` `if (acceptThread != null && acceptThread != Thread.currentThread()) {`
    현재 thread가 자기 자신을 join하지 않도록 막으면서 accept thread 종료를 기다립니다.
83. `132` `acceptThread.join(2_000);`
    최대 2초 동안 accept thread가 정리되기를 기다립니다.
84. `133` `}`
    accept thread join 분기를 닫습니다.
85. `135` `connectionExecutor.shutdownNow();`
    연결 worker들에게 즉시 종료 신호를 보냅니다.
86. `136` `connectionExecutor.awaitTermination(2, TimeUnit.SECONDS);`
    worker들이 실제로 정리될 시간을 조금 줍니다.
87. `137` `observationLog.event("server-stop", "server", "closed");`
    서버 종료 로그를 남깁니다.
88. `138` `}`
    `close()`를 닫습니다.
89. `140-142` `channelName(...)`
    숫자 id를 `client-N` 문자열로 바꾸는 작은 helper입니다.
90. `144` `public static void main(String[] args) throws Exception {`
    로컬에서 이 예제 서버를 직접 실행할 때 들어오는 main 진입점입니다.
91. `145` `int port = Lesson1Support.parsePort(args, 9001);`
    인자가 없으면 기본 포트 `9001`을 쓰고, 있으면 사용자가 준 포트를 읽습니다.
92. `146` `try (BlockingEchoServer server = BlockingEchoServer.start(port)) {`
    서버를 시작하고, main이 끝날 때 자동으로 닫히게 try-with-resources에 넣습니다.
93. `147` `Lesson1Support.keepRunning(server);`
    Ctrl+C가 올 때까지 프로세스를 유지합니다.
94. `148` `}`
    main의 try 블록을 닫습니다.
95. `149-150` 파일 종료
    `main`과 클래스 본문을 모두 닫습니다.

## [NioSelectorEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/nio/NioSelectorEchoServer.java)

### 패키지, import, 클래스 선언

1. `1` package 선언
   이 파일이 첫 강의의 `nio` 구현임을 명시합니다.
2. `3-5` 공통 lesson import
   실행 방식과 로그 형식을 다른 서버와 맞추기 위한 공통 도구입니다.
3. `7-19` Java NIO 관련 import
   이 구현이 selector, channel, buffer, attachment 상태를 직접 다룬다는 사실이 import만 봐도 드러납니다.
4. `21-27` 클래스 KDoc
   한 selector thread가 여러 연결을 번갈아 처리한다는 핵심 비교 포인트를 먼저 말합니다.
5. `28` 클래스 선언
   NIO echo 서버 구현을 시작하고 공통 `LessonServer` 계약을 맞춥니다.

### 필드와 생성자

6. `30` `ObservationLog` 필드
   이 구현의 로그 이름표를 `nio`로 고정합니다.
7. `31` `connectionIds`
   selector 기반이어도 연결마다 구분 번호는 필요하므로 원자 카운터를 둡니다.
8. `32` `requestedPort`
   바인딩 전까지 원하는 포트 번호를 들고 있습니다.
9. `34` `running`
   selector loop를 멈추기 위한 실행 플래그입니다.
10. `35` `selector`
    준비된 채널을 골라 주는 selector 객체를 담습니다.
11. `36` `serverSocketChannel`
    새 연결을 받는 NIO 서버 채널입니다.
12. `37` `selectorThread`
    selector loop를 실제로 돌리는 thread를 보관합니다.
13. `39-41` 생성자
    외부에서는 `start(...)`만 쓰게 하고 내부에서 requested port만 저장합니다.

### 시작 경로

14. `43-47` `start(...)`
    blocking 예제와 같은 패턴으로 객체를 만들고 내부 시작 절차를 호출한 뒤 반환합니다.
15. `49` `startInternal()` 선언
    실제 selector와 channel 초기화를 여는 메서드입니다.
16. `50` `selector = Selector.open();`
    준비된 IO 이벤트를 받을 selector를 엽니다.
17. `51` `serverSocketChannel = ServerSocketChannel.open();`
    서버용 채널을 엽니다.
18. `52` `serverSocketChannel.configureBlocking(false);`
    이 채널을 non-blocking 모드로 바꿉니다. 이 줄이 blocking 예제와 갈라지는 첫 번째 핵심 지점입니다.
19. `53` `serverSocketChannel.bind(new InetSocketAddress(requestedPort));`
    실제 포트에 바인딩합니다.
20. `54` `serverSocketChannel.register(selector, SelectionKey.OP_ACCEPT);`
    이 서버 채널에서 accept 준비가 생기면 selector가 알려 달라고 등록합니다.
21. `55` `running = true;`
    event loop 실행 플래그를 올립니다.
22. `56` `server-start` 로그
    실제 바인딩된 포트를 기록합니다.
23. `57` `selectorThread = Thread.ofPlatform().name("nio-selector").start(this::runEventLoop);`
    `nio-selector`라는 이름의 thread 하나가 event loop 전체를 맡게 합니다.
24. `58` 시작 절차 종료
    `startInternal()`을 닫습니다.

### selector event loop

25. `60` `runEventLoop()` 선언
    selector 기반 event loop를 엽니다.
26. `61-62` read buffer 설명 주석
    첫 강의에서는 thread가 하나뿐이라 buffer 하나 재사용으로 단순화해도 된다는 뜻입니다.
27. `63` `ByteBuffer readBuffer = ByteBuffer.allocate(1024);`
    읽기용 임시 버퍼 하나를 만듭니다.
28. `65` `while (running) {`
    종료 전까지 selector loop를 반복합니다.
29. `66` `try {`
    selector 관련 예외를 처리하려고 보호 구역을 엽니다.
30. `67-68` `selector.select()` 설명 주석
    blocking 모델처럼 연결마다 잠드는 것이 아니라, 준비된 일만 골라 받는다는 점을 먼저 설명합니다.
31. `69` `selector.select();`
    처리 가능한 채널이 생길 때까지 이 thread가 기다립니다.
32. `70` `if (!running || !selector.isOpen()) {`
    종료 중이거나 selector가 이미 닫혔으면 더 진행하지 않겠다는 방어 분기입니다.
33. `71` `return;`
    event loop를 끝냅니다.
34. `72` 분기 종료
    종료 검사 블록을 닫습니다.
35. `73` `Set<SelectionKey> selectedKeys = selector.selectedKeys();`
    방금 준비되었다고 표시된 키 집합을 꺼냅니다.
36. `74` `Iterator<SelectionKey> iterator = selectedKeys.iterator();`
    선택된 키를 하나씩 소비할 iterator를 만듭니다.
37. `75` `while (iterator.hasNext()) {`
    준비된 키를 모두 처리할 때까지 반복합니다.
38. `76` `SelectionKey key = iterator.next();`
    다음 키 하나를 꺼냅니다.
39. `77` `iterator.remove();`
    같은 키를 다음 루프에서 또 처리하지 않도록 선택 집합에서 제거합니다.
40. `79` `if (!key.isValid()) {`
    이미 취소되었거나 무효가 된 키면 건너뜁니다.
41. `80` `continue;`
    다음 키로 넘어갑니다.
42. `81` 무효 키 분기 종료
    첫 번째 필터를 닫습니다.
43. `82` `if (key.isAcceptable()) {`
    이 키가 "새 연결을 받을 준비가 됨"을 뜻하는지 확인합니다.
44. `83` `acceptClient();`
    새 연결 수락 전용 메서드로 넘깁니다.
45. `84` `continue;`
    accept를 처리했으면 같은 키로 read까지 하지 않고 다음 키로 갑니다.
46. `85` accept 분기 종료
    accept 처리 분기를 닫습니다.
47. `86` `if (key.isReadable()) {`
    이 키가 "읽을 데이터가 있음"을 뜻하는지 확인합니다.
48. `87` `readClient(key, readBuffer);`
    읽기 처리 메서드로 넘깁니다.
49. `88` read 분기 종료
    read 분기를 닫습니다.
50. `89` inner while 종료
    선택된 키 처리 루프를 닫습니다.
51. `90` `} catch (IOException ioException) {`
    selector loop 중 일반 입출력 예외를 받습니다.
52. `91` `if (running) {`
    종료 중이 아닌 진짜 오류일 때만 로그를 찍습니다.
53. `92` `selector-error` 로그
    selector 관련 오류를 기록합니다.
54. `93` 조건 분기 종료
    오류 로그 분기를 닫습니다.
55. `94` `return;`
    오류가 났으면 event loop를 끝냅니다.
56. `95` `} catch (ClosedSelectorException ignored) {`
    종료 과정에서 selector가 닫히며 생길 수 있는 예외는 조용히 받습니다.
57. `96` `return;`
    selector가 닫혔으면 정상 종료로 봅니다.
58. `97-99` loop와 메서드 종료
    event loop와 메서드를 닫습니다.

### accept와 read 처리

59. `101` `acceptClient()` 선언
    새 연결을 받아 등록하는 절차입니다.
60. `102` `SocketChannel socketChannel = serverSocketChannel.accept();`
    서버 채널에서 새 연결 하나를 받아 옵니다.
61. `103` `if (socketChannel == null) {`
    non-blocking accept는 준비 신호가 있어도 실제 채널이 아직 없을 수 있어 null 방어가 필요합니다.
62. `104` `return;`
    실제 연결이 없으면 그냥 끝냅니다.
63. `105` 분기 종료
    null 방어 분기를 닫습니다.
64. `107` `socketChannel.configureBlocking(false);`
    새 연결 채널도 non-blocking 모드로 둡니다.
65. `108` `connectionId` 증가
    새 연결에 번호를 붙입니다.
66. `109` `ClientState state = new ClientState(connectionId);`
    이 연결이 아직 다 받지 못한 줄 조각을 저장할 상태 객체를 만듭니다.
67. `110` `socketChannel.register(selector, SelectionKey.OP_READ, state);`
    이 채널이 읽을 준비가 되면 selector가 알려 주도록 등록하고, attachment로 `state`를 같이 달아 둡니다.
68. `111` `accept` 로그
    새 연결이 들어왔음을 기록합니다.
69. `112` 메서드 종료
    accept 처리를 닫습니다.
70. `114` `readClient(...)` 선언
    선택된 읽기 키 하나를 실제로 소비하는 메서드입니다.
71. `115` `SocketChannel channel = (SocketChannel) key.channel();`
    selection key에서 실제 채널을 꺼냅니다.
72. `116` `ClientState state = (ClientState) key.attachment();`
    이 채널과 함께 저장해 둔 줄 조립 상태를 꺼냅니다.
73. `118` `readBuffer.clear();`
    재사용하는 버퍼를 새 읽기용 상태로 되돌립니다.
74. `119` `int bytesRead = channel.read(readBuffer);`
    지금 읽을 수 있는 바이트를 버퍼에 채웁니다.
75. `120` `if (bytesRead == -1) {`
    상대가 연결을 닫아 EOF가 왔는지 확인합니다.
76. `121` `client-close` 로그
    정상 종료를 기록합니다.
77. `122` `closeChannel(channel, key);`
    키를 취소하고 채널을 닫습니다.
78. `123` `return;`
    EOF 처리 후 메서드를 끝냅니다.
79. `124` EOF 분기 종료
    첫 번째 종료 조건을 닫습니다.
80. `126` `if (bytesRead == 0) {`
    읽을 준비는 되었지만 실제로 읽은 바이트가 없으면 더 할 일이 없습니다.
81. `127` `return;`
    그대로 끝냅니다.
82. `128` zero-read 분기 종료
    두 번째 빠른 종료를 닫습니다.
83. `130` `read-bytes` 로그
    몇 바이트를 읽었는지 먼저 기록합니다.
84. `131` `readBuffer.flip();`
    방금 채운 버퍼를 읽기 모드로 뒤집습니다.
85. `133-134` framing 설명 주석
    non-blocking read에서는 한 줄이 한 번에 안 올 수 있으므로, 줄 경계를 직접 조립해야 한다는 사실을 요약합니다.
86. `135` `while (readBuffer.hasRemaining()) {`
    버퍼 안의 바이트를 하나씩 소비합니다.
87. `136` `byte next = readBuffer.get();`
    다음 바이트 하나를 꺼냅니다.
88. `137` `if (next == '\n') {`
    줄바꿈을 만나면 한 줄이 완성되었다고 봅니다.
89. `138` `String line = state.takeLine();`
    지금까지 모은 바이트를 문자열 한 줄로 바꿉니다.
90. `139` `read-line` 로그
    애플리케이션이 이해할 한 줄 메시지를 만들었다는 사실을 기록합니다.
91. `140` `writeLine(channel, state, line);`
    echo 응답을 씁니다.
92. `141` `continue;`
    줄 하나를 처리했으면 다음 바이트로 갑니다.
93. `142` 줄바꿈 분기 종료
    첫 번째 바이트 분기를 닫습니다.
94. `143` `if (next != '\r') {`
    CRLF에서 carriage return은 내용에 넣지 않으려고 건너뜁니다.
95. `144` `state.append(next);`
    줄바꿈이 아닌 바이트를 현재 줄 버퍼에 붙입니다.
96. `145` carriage return 분기 종료
    두 번째 바이트 분기를 닫습니다.
97. `146-147` while와 메서드 종료
    버퍼 소비 루프와 `readClient(...)`를 닫습니다.

### write, close, main, ClientState

98. `149` `writeLine(...)` 선언
    한 줄 문자열을 다시 소켓으로 밀어 넣는 메서드입니다.
99. `150` `ByteBuffer response = StandardCharsets.UTF_8.encode(line + "\n");`
    응답 문자열을 UTF-8 바이트 버퍼로 바꿉니다.
100. `151-152` write loop 설명 주석
     NIO에서는 한 번의 `write(...)`로 모두 보내지 못할 수 있어 직접 반복해야 한다는 점을 강조합니다.
101. `153` `while (response.hasRemaining()) {`
     아직 안 써진 바이트가 남아 있는 동안 계속 씁니다.
102. `154` `channel.write(response);`
     실제 소켓 채널에 남은 바이트를 보냅니다.
103. `155` write loop 종료
     모든 바이트를 보냈으면 루프를 닫습니다.
104. `156` `write` 로그
     echo 응답 완료를 기록합니다.
105. `157` 메서드 종료
     `writeLine(...)`을 닫습니다.
106. `159-162` `closeChannel(...)`
     selection key를 취소하고 채널을 닫는 작은 정리 메서드입니다.
107. `164-167` `serverName()`
     공통 이름을 `nio`로 돌려줍니다.
108. `169-176` `port()`
     로컬 주소에서 실제 바인딩 포트를 읽고, 실패하면 의미 있는 예외로 바꿉니다.
109. `178-194` `close()`
     running을 내리고 selector를 깨운 뒤, 서버 채널과 selector를 닫고 selector thread 정리까지 마무리합니다.
110. `196-200` `main(...)`
     기본 포트 `9002`를 사용해 이 예제를 직접 띄우는 entry point입니다.
111. `203-206` `ClientState` KDoc
     `readLine()`이 숨겨 주던 줄 조립 상태를 NIO에서는 직접 들고 있어야 한다는 점을 설명합니다.
112. `207` `private static final class ClientState {`
     연결별 읽기 상태를 담는 작은 내부 타입을 엽니다.
113. `208` `connectionId`
     이 상태가 어느 연결용인지 기억합니다.
114. `209` `lineBuffer`
     아직 줄바꿈을 만나지 못한 바이트를 임시로 모아 둡니다.
115. `211-213` 생성자
     연결 번호를 저장합니다.
116. `215-217` `append(...)`
     바이트 하나를 현재 줄 버퍼 끝에 붙입니다.
117. `219` `takeLine()` 선언
     모은 바이트를 문자열 한 줄로 꺼내는 메서드를 엽니다.
118. `220` `String line = lineBuffer.toString(StandardCharsets.UTF_8);`
     누적 바이트를 UTF-8 문자열로 해석합니다.
119. `221` `lineBuffer.reset();`
     다음 줄을 위해 버퍼를 비웁니다.
120. `222` `return line;`
     완성된 줄을 돌려줍니다.
121. `223` 메서드 종료
     `takeLine()`을 닫습니다.
122. `225-227` `channelName()`
     connection id를 `client-N` 형식으로 바꿉니다.
123. `228-229` 타입 종료
     `ClientState`와 바깥 클래스를 닫습니다.

## [NettyEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/netty/NettyEchoServer.java)

### 패키지, import, 클래스 선언

1. `1` package 선언
   이 파일이 첫 강의의 Netty 구현임을 나타냅니다.
2. `3-5` lesson 공통 import
   공통 실행 틀과 공통 로그를 재사용합니다.
3. `6-19` Netty import
   이 구현이 `ServerBootstrap`, `EventLoopGroup`, `ChannelPipeline`, codec, handler라는 추상화를 쓴다는 사실이 여기서 드러납니다.
4. `21-24` Java 표준 import
   포트 읽기, UTF-8, 종료 대기, 연결 id 카운터에 필요한 표준 라이브러리입니다.
5. `26-32` 클래스 KDoc
   NIO에서 직접 하던 일을 Netty가 어떤 추상화 경계로 나눠 주는지라는 학습 질문을 먼저 고정합니다.
6. `33` 클래스 선언
   Netty echo 서버 구현을 시작합니다.

### 필드와 생성자

7. `35` `private static final AttributeKey<Integer> CONNECTION_ID = AttributeKey.valueOf("lesson1.connectionId");`
   child channel마다 연결 번호를 붙여 두기 위한 Netty attribute key를 미리 만듭니다.
8. `37` `ObservationLog` 필드
   이 구현의 로그 이름표를 `netty`로 고정합니다.
9. `38` `connectionIds`
   channel마다 번호를 붙일 원자 카운터입니다.
10. `39` `requestedPort`
    사용자가 원하는 포트 번호를 저장합니다.
11. `41` `bossGroup`
    새 연결 accept를 맡는 EventLoopGroup 필드입니다.
12. `42` `workerGroup`
    child channel IO와 handler 실행을 맡는 EventLoopGroup 필드입니다.
13. `43` `serverChannel`
    서버 바인딩 후 얻는 Netty channel을 보관합니다.
14. `45-47` 생성자
    requested port만 저장하고 실제 부팅은 `startInternal()`에 맡깁니다.

### 시작 경로와 bootstrap

15. `49-53` `start(...)`
    blocking/NIO 예제와 같은 패턴으로 객체 생성 -> 내부 시작 -> 반환 순서를 유지합니다.
16. `55` `startInternal()` 선언
    Netty bootstrap과 bind를 실제로 수행하는 메서드입니다.
17. `56` boss/worker 설명 주석
    accept와 client IO를 분리해 thread 이름만 봐도 역할이 갈리는 모습을 보여 주겠다는 선언입니다.
18. `57` `bossGroup = new NioEventLoopGroup(1, new DefaultThreadFactory("netty-boss"));`
    accept 전용 boss group을 스레드 1개로 만들고 이름 접두사를 `netty-boss`로 줍니다.
19. `58` `workerGroup = new NioEventLoopGroup(0, new DefaultThreadFactory("netty-worker"));`
    worker group은 Netty 기본 크기를 따르되 이름 접두사는 `netty-worker`로 줍니다.
20. `60` `ServerBootstrap bootstrap = new ServerBootstrap()`
    서버 bootstrap 설정을 builder 방식으로 시작합니다.
21. `61` `.group(bossGroup, workerGroup)`
    boss/worker 두 그룹을 bootstrap에 연결합니다.
22. `62` `.channel(NioServerSocketChannel.class)`
    서버 채널 구현으로 NIO 기반 server socket channel을 쓰겠다고 지정합니다.
23. `63` `.handler(new BossAcceptLogHandler())`
    서버 채널 레벨에서 accept 관측 로그를 찍는 handler를 붙입니다.
24. `64` `.childHandler(new ChannelInitializer<SocketChannel>() {`
    새 child channel이 생길 때마다 pipeline을 초기화할 규칙을 등록합니다.
25. `65` `@Override`
    부모 클래스의 `initChannel(...)` 구현임을 표시합니다.
26. `66` `protected void initChannel(SocketChannel channel) {`
    child channel pipeline 구성을 시작합니다.
27. `67-69` pipeline 설명 주석
    NIO에서 직접 하던 framing/encoding을 Netty는 handler 체인으로 나눈다는 비교 포인트를 요약합니다.
28. `70` `channel.pipeline()`
    이 child channel의 pipeline 구성을 시작합니다.
29. `71` `.addLast("lineDecoder", new LineBasedFrameDecoder(1024))`
    바이트 흐름을 줄 단위 프레임으로 자르는 decoder를 먼저 붙입니다.
30. `72` `.addLast("stringDecoder", new StringDecoder(StandardCharsets.UTF_8))`
    프레임 바이트를 UTF-8 문자열로 바꾸는 decoder를 붙입니다.
31. `73` `.addLast("stringEncoder", new StringEncoder(StandardCharsets.UTF_8))`
    문자열을 다시 UTF-8 바이트로 바꾸는 encoder를 붙입니다.
32. `74` `.addLast("echoHandler", new EchoHandler());`
    마지막에 실제 비즈니스 동작인 echo handler를 붙입니다.
33. `75` `}`
    `initChannel(...)`을 닫습니다.
34. `76` `});`
    child handler 설정 전체를 닫습니다.
35. `78` `var bindFuture = bootstrap.bind(requestedPort).awaitUninterruptibly();`
    지정된 포트에 bind를 시도하고 완료될 때까지 기다립니다.
36. `79` `if (!bindFuture.isSuccess()) {`
    bind 실패 여부를 검사합니다.
37. `80` `Throwable cause = bindFuture.cause();`
    실패 원인을 꺼냅니다.
38. `81` `close();`
    부분적으로 열린 자원을 먼저 정리합니다.
39. `82` `if (cause instanceof Exception exception) {`
    원인이 `Exception`이면 그대로 다시 던질 수 있는지 확인합니다.
40. `83` `throw exception;`
    원래 예외를 유지해 다시 던집니다.
41. `84` `}`
    예외 타입 검사 분기를 닫습니다.
42. `85` `throw new IllegalStateException("Netty 서버 바인딩에 실패했습니다.", cause);`
    체크 예외가 아닌 경우 의미 있는 래핑 예외로 바꿉니다.
43. `86` bind 실패 분기 종료
    실패 처리 블록을 닫습니다.
44. `88` `serverChannel = bindFuture.channel();`
    성공한 bind 결과에서 서버 채널을 보관합니다.
45. `89` `server-start` 로그
    실제 포트를 기록합니다.
46. `90` 메서드 종료
    `startInternal()`을 닫습니다.

### 공통 계약과 종료

47. `92-95` `serverName()`
    공통 이름을 `netty`로 돌려줍니다.
48. `97-100` `port()`
    서버 채널의 로컬 주소에서 실제 포트를 읽습니다.
49. `102` `@Override`
    공통 종료 계약 구현임을 표시합니다.
50. `103` `public void close() {`
    Netty 자원 정리 절차를 엽니다.
51. `104` `if (serverChannel != null) {`
    서버 채널이 실제로 열려 있으면 닫습니다.
52. `105` `serverChannel.close().awaitUninterruptibly();`
    close future 완료까지 기다립니다.
53. `106` server channel 분기 종료
    첫 정리 분기를 닫습니다.
54. `107` `if (bossGroup != null) {`
    boss group이 있으면 종료합니다.
55. `108` `bossGroup.shutdownGracefully(0, 5, TimeUnit.SECONDS).awaitUninterruptibly();`
    boss event loop를 우아하게 종료하고 완료까지 기다립니다.
56. `109` boss 분기 종료
    boss group 정리를 닫습니다.
57. `110` `if (workerGroup != null) {`
    worker group이 있으면 종료합니다.
58. `111` `workerGroup.shutdownGracefully(0, 5, TimeUnit.SECONDS).awaitUninterruptibly();`
    worker event loop를 우아하게 종료하고 완료까지 기다립니다.
59. `112` worker 분기 종료
    worker group 정리를 닫습니다.
60. `113` `server-stop` 로그
    서버 종료를 기록합니다.
61. `114` `}`
    `close()`를 닫습니다.
62. `116-120` `main(...)`
    기본 포트 `9003`으로 이 예제를 직접 실행하는 entry point입니다.

### EchoHandler

63. `123-129` `EchoHandler` KDoc
    Netty에서는 비즈니스 handler가 이미 "한 줄 문자열"만 받는다는 사실을 먼저 설명합니다.
64. `130` `private final class EchoHandler extends SimpleChannelInboundHandler<String> {`
    문자열 한 줄만 읽어 처리하는 inbound handler를 엽니다.
65. `132` `@Override`
    channel 활성화 콜백 구현임을 표시합니다.
66. `133` `public void channelActive(ChannelHandlerContext context) {`
    child channel이 worker owner를 갖고 활성화될 때 호출되는 메서드입니다.
67. `134-135` 주석
    connection id를 boss accept 시점이 아니라 worker ownership이 잡힌 뒤에 붙이는 이유를 설명합니다.
68. `136` `int connectionId = connectionIds.incrementAndGet();`
    새 active channel에 번호를 하나 할당합니다.
69. `137` `context.channel().attr(CONNECTION_ID).set(connectionId);`
    방금 만든 번호를 channel attribute에 저장합니다.
70. `138` `channel-active` 로그
    worker thread 위에서 child channel이 활성화되었음을 기록합니다.
71. `139` `context.fireChannelActive();`
    다음 handler들에게도 active 이벤트를 계속 흘려 보냅니다.
72. `140` 메서드 종료
    `channelActive(...)`를 닫습니다.
73. `142` `@Override`
    메시지 읽기 콜백 구현임을 표시합니다.
74. `143` `protected void channelRead0(ChannelHandlerContext context, String message) {`
    decoder를 통과한 문자열 한 줄을 받는 메서드를 엽니다.
75. `144` `read` 로그
    읽은 한 줄 문자열을 기록합니다.
76. `145` `String response = message + "\n";`
    line protocol을 유지하려고 줄바꿈을 다시 붙입니다.
77. `146` `write` 로그
    echo 응답을 보낸다는 사실을 기록합니다.
78. `147-148` writeAndFlush 설명 주석
    Netty가 NIO의 write loop와 상태 관리를 프레임워크 경계 뒤로 감춘다는 점을 강조합니다.
79. `149` `context.writeAndFlush(response);`
    outbound 경로를 따라 응답을 보내고 즉시 flush합니다.
80. `150` 메서드 종료
    `channelRead0(...)`를 닫습니다.
81. `152-156` `channelInactive(...)`
    연결 종료를 로그로 남기고 다음 handler로도 inactive 이벤트를 흘려 보냅니다.
82. `158-162` `exceptionCaught(...)`
    채널 오류를 기록하고 해당 채널을 닫습니다.
83. `164` `private String channelName(ChannelHandlerContext context) {`
    채널에 저장된 connection id를 사람이 읽기 쉬운 이름으로 바꾸는 helper를 엽니다.
84. `165` `Integer connectionId = context.channel().attr(CONNECTION_ID).get();`
    channel attribute에서 연결 번호를 꺼냅니다.
85. `166` `if (connectionId == null) {`
    아직 번호를 붙이기 전이면 임시 이름을 써야 합니다.
86. `167` `return "client-pending";`
    boss accept 직후처럼 번호가 아직 없을 때 쓰는 이름입니다.
87. `168` null 분기 종료
    임시 이름 분기를 닫습니다.
88. `169` `return "client-" + connectionId;`
    번호가 있으면 최종 이름을 돌려줍니다.
89. `170-171` helper와 handler 종료
    `channelName(...)`과 `EchoHandler`를 닫습니다.

### BossAcceptLogHandler

90. `173-176` KDoc
    비즈니스에는 꼭 필요 없지만 학습용 관측에는 중요한 handler라는 점을 설명합니다.
91. `177` `private final class BossAcceptLogHandler extends ChannelInboundHandlerAdapter {`
    accept 관측 전용 handler를 엽니다.
92. `179` `@Override`
    inbound `channelRead(...)` 구현임을 표시합니다.
93. `180` `public void channelRead(ChannelHandlerContext context, Object message) throws Exception {`
    boss channel이 child channel을 읽어들였을 때 호출되는 메서드입니다.
94. `181` `boss-accept` 로그
    아직 worker handoff 전인 시점의 accept 이벤트를 기록합니다.
95. `182` `super.channelRead(context, message);`
    이벤트를 다음 단계로 넘겨 실제 child channel 등록과 handoff가 계속 일어나게 합니다.
96. `183-185` 종료
    `channelRead(...)`, `BossAcceptLogHandler`, 바깥 클래스를 닫습니다.

## 공통 support와 테스트 코드

### [ObservationLog.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/common/ObservationLog.java)

1. `1` package 선언
   첫 강의 공통 지원 코드 패키지에 이 타입을 둡니다.
2. `3-6` import
   시간 포맷, null 안전, event sequence 카운터에 필요한 표준 도구입니다.
3. `8-11` KDoc
   왜 세 구현이 같은 로그 형식을 공유해야 하는지 먼저 설명합니다.
4. `12` 클래스 선언
   공통 관측 로그 타입을 엽니다.
5. `14` `TIME_FORMATTER`
   시각을 `HH:mm:ss.SSS` 형식으로 맞춰 보기 쉽게 만듭니다.
6. `15` `eventSequence`
   이벤트 순서를 1, 2, 3...으로 붙여 로그를 시간만이 아니라 순서로도 비교하게 합니다.
7. `16` `serverName`
   blocking/NIO/Netty 중 어느 서버 로그인지 저장합니다.
8. `17` `startedAtNanos`
   서버 시작 시점을 저장해 경과 시간을 계산합니다.
9. `19-21` 생성자
   서버 이름을 필드에 저장합니다.
10. `23-26` `event(...)` KDoc
    로그 열 순서를 통일해 비교 학습을 쉽게 만들겠다는 의도입니다.
11. `27` `event(...)` 선언
    로그 한 줄을 찍는 공통 메서드를 엽니다.
12. `28` `elapsedMillis`
    시작 후 얼마나 지났는지 밀리초로 계산합니다.
13. `29` `sequence`
    이 이벤트의 순번을 하나 올려 받습니다.
14. `31` `synchronized (System.out) {`
    여러 thread가 동시에 찍어도 로그 한 줄이 섞이지 않게 보호합니다.
15. `32-42` `System.out.printf(...)`
    순번, 현재 시각, 경과 시간, 서버 이름, phase, channel, thread 이름, 세부 내용을 같은 칼럼 순서로 출력합니다.
16. `43-44` 종료
    동기화 블록과 `event(...)`를 닫습니다.
17. `46` `previewText(...)` 선언
    너무 긴 메시지를 잘라 로그 가독성을 지키는 helper입니다.
18. `47-49` `escaped` 계산
    줄바꿈 문자들을 로그에서 보이도록 escape합니다.
19. `50` 길이 검사
    80자 이하인지 확인합니다.
20. `51` `return escaped;`
    짧으면 그대로 돌려줍니다.
21. `52-53` 긴 텍스트 처리
    너무 길면 앞 77자만 남기고 `...`를 붙입니다.
22. `54-55` 종료
    helper와 클래스를 닫습니다.

### [Lesson1Support.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/common/Lesson1Support.java)

1. `1` package 선언
   첫 강의 공통 지원 코드 패키지입니다.
2. `3` import
   프로세스를 계속 살려 두는 latch에 필요합니다.
3. `5-8` KDoc
   실행 방식 차이가 아니라 처리 모델 차이에 집중하게 하려는 공통 지원 코드라는 뜻입니다.
4. `9` 클래스 선언
   정적 helper만 모은 타입을 엽니다.
5. `11-12` private 생성자
   인스턴스화를 막습니다.
6. `14` `parsePort(...)` 선언
   main 메서드들이 공통으로 쓰는 포트 파서입니다.
7. `15` `if (args.length == 0) {`
   인자가 없으면 기본 포트를 쓰겠다는 분기입니다.
8. `16` `return defaultPort;`
   기본 포트를 돌려줍니다.
9. `17` 첫 분기 종료
   인자 없음 분기를 닫습니다.
10. `18` `if (args.length != 1) {`
    포트는 하나만 받겠다는 검사를 합니다.
11. `19` `throw new IllegalArgumentException(...)`
    잘못된 인자 개수면 이해하기 쉬운 예외를 던집니다.
12. `20` 두 번째 분기 종료
    인자 개수 검사를 닫습니다.
13. `21` `return Integer.parseInt(args[0]);`
    사용자가 준 포트를 정수로 바꿔 돌려줍니다.
14. `22` helper 종료
    `parsePort(...)`를 닫습니다.
15. `24` `keepRunning(...)` 선언
    서버 프로세스를 Ctrl+C까지 계속 살려 두는 공통 helper입니다.
16. `25-27` shutdown hook 등록
    프로세스가 내려갈 때 서버를 조용히 닫는 hook thread를 등록합니다.
17. `29-30` ready 메시지 주석
    세 구현이 같은 ready 형식을 쓰게 해 학습자가 로그 차이에 집중하도록 돕습니다.
18. `31` `System.out.printf(...)`
    서버 이름과 포트를 출력합니다.
19. `32` `new CountDownLatch(1).await();`
    카운트다운 되지 않는 latch로 현재 thread를 사실상 무기한 대기시킵니다.
20. `33` helper 종료
    `keepRunning(...)`을 닫습니다.
21. `35` `closeQuietly(...)` 선언
    shutdown hook에서 조용히 닫기 위한 helper입니다.
22. `36-37` `try`
    `close()` 호출을 시도합니다.
23. `38` `} catch (Exception ignored) {`
    종료 중 예외는 삼킵니다.
24. `39` 주석
    shutdown hook에서는 종료 자체를 막지 않는 것이 더 중요하다는 뜻입니다.
25. `40-42` 종료
    `catch`, helper, 클래스를 닫습니다.

### [LessonServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/common/LessonServer.java)

1. `1` package 선언
   첫 강의 공통 계약 패키지입니다.
2. `3-7` KDoc
   문제는 같게 두고 구현만 바꿔 본다는 비교 학습 목적을 설명합니다.
3. `8` 인터페이스 선언
   모든 서버가 맞춰야 할 최소 계약을 정의합니다.
4. `10` `String serverName();`
   로그와 출력에 쓸 서버 이름 계약입니다.
5. `12` `int port();`
   실제 바인딩된 포트를 읽는 계약입니다.
6. `14` `@Override`
   `AutoCloseable`의 `close()`를 그대로 약속하겠다는 뜻입니다.
7. `15` `void close() throws Exception;`
   서버 자원 정리 계약을 선언합니다.
8. `16` 종료
   인터페이스를 닫습니다.

### [EchoTestClient.java](../../src/test/java/io/aimpugn/learn/netty/lesson1/EchoTestClient.java)

1. `1` package 선언
   첫 강의 테스트 코드 패키지입니다.
2. `3-12` import
   테스트 클라이언트가 실제 TCP 소켓으로 서버와 이야기하는 데 필요한 표준 도구입니다.
3. `14` 클래스 선언
   테스트 전용 helper를 엽니다.
4. `16` `TIMEOUT_MILLIS`
   응답이 오지 않아 테스트가 영원히 멈추지 않도록 2초 타임아웃을 둡니다.
5. `18-19` private 생성자
   정적 helper만 쓰게 합니다.
6. `21` `exchangeSingleLine(...)` 선언
   한 줄 요청-응답을 검증하는 helper입니다.
7. `22` `Socket socket = connect(port);`
   테스트용 TCP 연결을 엽니다.
8. `23` `BufferedReader reader = ...`
   응답을 줄 단위로 읽기 위한 reader를 만듭니다.
9. `24` `BufferedWriter writer = ...`
   요청을 줄 단위로 쓰기 위한 writer를 만듭니다.
10. `26` `writer.write(line);`
    요청 문자열 본문을 씁니다.
11. `27` `writer.newLine();`
    줄바꿈으로 메시지 경계를 만듭니다.
12. `28` `writer.flush();`
    요청을 즉시 전송합니다.
13. `29` `return reader.readLine();`
    서버가 echo로 돌려준 한 줄을 읽어 반환합니다.
14. `30-31` 종료
    try-with-resources와 메서드를 닫습니다.
15. `33` `exchangeTwoLines(...)` 선언
    같은 연결에서 두 줄 연속 echo를 검증하는 helper입니다.
16. `34-36` socket, reader, writer 준비
    single-line helper와 같은 방식으로 연결과 스트림을 엽니다.
17. `38` `List<String> responses = new ArrayList<>();`
    두 응답을 차례로 담을 리스트를 만듭니다.
18. `39-42` 첫 번째 요청-응답
    첫 줄을 보내고 flush한 뒤 응답 한 줄을 리스트에 넣습니다.
19. `44-47` 두 번째 요청-응답
    같은 연결에서 두 번째 줄도 보내고 응답을 추가합니다.
20. `48` `return responses;`
    두 응답을 함께 돌려줍니다.
21. `49-50` 종료
    try와 메서드를 닫습니다.
22. `52` `connectAndCloseWithoutSending(...)` 선언
    데이터를 안 보내고 연결만 열었다 닫는 클라이언트를 흉내 내는 helper입니다.
23. `53` `try (Socket socket = connect(port)) {`
    연결만 열고 자동으로 닫히게 합니다.
24. `54` 주석
    이 테스트가 일부러 "연결만 열고 바로 닫는" 상황을 만드는 것임을 설명합니다.
25. `55-56` 종료
    try와 helper를 닫습니다.
26. `58` `connect(...)` 선언
    공통 연결 생성 helper입니다.
27. `59` `Socket socket = new Socket("127.0.0.1", port);`
    로컬 서버에 TCP 연결을 엽니다.
28. `60` `socket.setSoTimeout(TIMEOUT_MILLIS);`
    읽기 타임아웃을 걸어 테스트가 무한 대기하지 않게 합니다.
29. `61` `return socket;`
    준비된 소켓을 돌려줍니다.
30. `62-63` 종료
    helper와 클래스를 닫습니다.

### [Lesson1EchoServersTest.java](../../src/test/java/io/aimpugn/learn/netty/lesson1/Lesson1EchoServersTest.java)

1. `1` package 선언
   첫 강의 테스트 코드 패키지입니다.
2. `3-9` import
   세 서버 구현과 JUnit 파라미터화 테스트 도구를 가져옵니다.
3. `11-12` import
   기대 결과 리스트와 서버 spec 스트림에 필요합니다.
4. `14` 클래스 선언
   첫 강의 통합 테스트 본문을 엽니다.
5. `16-17` 첫 번째 파라미터화 테스트 애노테이션
   모든 서버 구현이 "한 줄 echo"를 통과하는지 같은 문장으로 검증하겠다는 뜻입니다.
6. `18` `echoesSingleLine(...)` 선언
   한 줄 echo 성공 케이스를 테스트합니다.
7. `19` `try (LessonServer server = serverSpec.starter().start(0)) {`
   각 구현을 임의 포트로 띄우고, 테스트가 끝나면 자동으로 닫습니다.
8. `20` `String response = EchoTestClient.exchangeSingleLine(...);`
   실제 TCP 클라이언트로 한 줄 요청을 보내 응답을 받습니다.
9. `21` `Assertions.assertEquals(...)`
   응답이 정확히 같은 문자열인지 검증합니다.
10. `22-23` 종료
    첫 번째 테스트를 닫습니다.
11. `25-26` 두 번째 테스트 애노테이션
    같은 연결에서 두 줄을 처리하는지 검증하겠다는 뜻입니다.
12. `27` `echoesTwoLinesOnSameConnection(...)`
    두 줄 연속 echo 성공 케이스를 테스트합니다.
13. `28` 서버 시작
    각 구현을 임의 포트로 띄웁니다.
14. `29` 두 줄 요청
    테스트 클라이언트로 두 줄을 같은 연결에서 보냅니다.
15. `30` 기대 리스트 비교
    두 응답 순서와 내용이 모두 맞는지 검증합니다.
16. `31-32` 종료
    두 번째 테스트를 닫습니다.
17. `34-35` 세 번째 테스트 애노테이션
    데이터 없이 연결 후 종료하는 클라이언트도 서버를 망가뜨리지 않아야 한다는 케이스입니다.
18. `36` `survivesClientDisconnectWithoutData(...)`
    connect-close 실패 케이스 뒤 회복을 테스트합니다.
19. `37` 서버 시작
    각 구현을 임의 포트로 띄웁니다.
20. `38` `EchoTestClient.connectAndCloseWithoutSending(server.port());`
    데이터를 안 보내고 바로 닫는 클라이언트를 한 번 보냅니다.
21. `39` `exchangeSingleLine(...)`
    그 다음에도 서버가 정상 echo를 하는지 다시 확인합니다.
22. `40` `assertEquals(...)`
    회복 후 응답이 정확한지 검증합니다.
23. `41-42` 종료
    세 번째 테스트를 닫습니다.
24. `44-45` 네 번째 테스트 애노테이션
    이미 사용 중인 포트에 다시 바인딩하면 실패해야 한다는 케이스입니다.
25. `46` `failsToStartOnUsedPort(...)`
    포트 충돌 실패 케이스를 테스트합니다.
26. `47` `runningServer` 시작
    먼저 한 서버를 임의 포트에 띄웁니다.
27. `48` `Assertions.assertThrows(...)`
    같은 포트에 두 번째 서버를 띄우면 예외가 발생해야 함을 검증합니다.
28. `49-50` 종료
    네 번째 테스트를 닫습니다.
29. `52` `serverSpecs()` 선언
    파라미터화 테스트에 공급할 서버 목록을 만듭니다.
30. `53` `return Stream.of(`
    여러 구현을 하나의 스트림으로 묶기 시작합니다.
31. `54` blocking spec
    이름 `blocking`과 시작 함수 `BlockingEchoServer::start`를 묶습니다.
32. `55` nio spec
    이름 `nio`와 시작 함수를 묶습니다.
33. `56` netty spec
    이름 `netty`와 시작 함수를 묶습니다.
34. `57-58` 종료
    스트림 생성과 helper를 닫습니다.
35. `60` `private record ServerSpec(String name, ServerStarter starter) {`
    테스트가 공통으로 쓸 "이름 + 시작 함수" 묶음을 record로 정의합니다.
36. `61` `@Override`
    record 기본 `toString()`을 덮어쓰겠다는 뜻입니다.
37. `62` `public String toString() {`
    테스트 이름에 사람이 읽기 쉬운 값이 나오게 합니다.
38. `63` `return name;`
    record 이름만 출력합니다.
39. `64-65` 종료
    `toString()`과 record를 닫습니다.
40. `67` `@FunctionalInterface`
    함수 하나만 가진 인터페이스임을 표시합니다.
41. `68` `private interface ServerStarter {`
    각 서버의 `start(int)` 메서드를 같은 타입으로 받기 위한 함수형 인터페이스입니다.
42. `69` `LessonServer start(int port) throws Exception;`
    포트를 받아 서버를 시작하고 `LessonServer`를 돌려주는 계약입니다.
43. `70-71` 종료
    인터페이스와 테스트 클래스를 닫습니다.

## 이 문서를 읽는 순서

1. 먼저 [BlockingEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/blocking/BlockingEchoServer.java)에서 `accept()`와 `readLine()`이 어디서 block되는지 봅니다.
2. 그다음 [NioSelectorEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/nio/NioSelectorEchoServer.java)에서 `configureBlocking(false)`, `selector.select()`, `ClientState`를 따라가며 "thread는 줄었지만 상태 관리는 늘었다"는 점을 확인합니다.
3. 마지막으로 [NettyEchoServer.java](../../src/main/java/io/aimpugn/learn/netty/lesson1/netty/NettyEchoServer.java)에서 `LineBasedFrameDecoder`, `StringDecoder`, `StringEncoder`, `EchoHandler`로 책임이 어떻게 분리되는지 봅니다.

한 문장으로 기억하면 이렇습니다.

**blocking은 기다리는 위치가 코드에 그대로 보이고, NIO는 그 기다림을 selector와 상태 객체로 옮기며, Netty는 그 상태와 경계를 pipeline과 handler 체인으로 정리합니다.**
