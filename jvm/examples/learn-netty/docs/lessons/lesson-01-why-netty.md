# Lesson 01 — 왜 blocking 서버만으로는 부족하고 Netty가 필요한가?

한 문장으로 먼저 말하면, Netty는 "비동기 서버를 더 빠르게 만드는 마법 라이브러리"가 아니라, **Java NIO로 직접 만들면 금방 커지는 selector, 상태 관리, 디코딩, 스레드 handoff, 연결 lifecycle 처리를 사람이 감당할 수 있는 구조로 정리해 주는 프레임워크**입니다.

이 강의에서는 그 말을 추상적으로 믿지 않고, 아래 세 가지를 직접 나란히 놓고 봅니다.

1. 연결마다 스레드를 잡고 기다리는 blocking echo 서버
2. selector 하나가 여러 연결을 번갈아 보는 Java NIO echo 서버
3. 같은 event-driven 모델을 Netty의 `EventLoop`, `ChannelPipeline`, codec으로 정리한 Netty echo 서버

## 이 강의에서 실제로 만드는 것

- blocking echo 서버
  - [`BlockingEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/blocking/BlockingEchoServer.java)
- Java NIO selector echo 서버
  - [`NioSelectorEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/nio/NioSelectorEchoServer.java)
- Netty echo 서버
  - [`NettyEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/netty/NettyEchoServer.java)
- 공통 관측 로그
  - [`ObservationLog.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/common/ObservationLog.java)
- 자동 검증 테스트 12개
  - [`Lesson1EchoServersTest.java`](../../src/test/java/io/aimpugn/learn/netty/lesson1/Lesson1EchoServersTest.java)

## 버전과 이번 강의의 의도적 단순화

- 빌드 target: Java 21
- 실제 검증 환경: Java 25.0.2
- Netty: 4.1.132.Final
- 프로토콜 단순화:
  - 줄바꿈(`\n`)으로 메시지 경계를 나누는 UTF-8 line protocol만 사용합니다.
  - `ByteBuf`, backpressure, partial write, allocator, flush 튜닝은 아직 깊게 다루지 않습니다.

이 단순화는 "실무 서버 그대로"를 만들기 위해서가 아니라, **스레드 모델과 이벤트 흐름 차이를 먼저 눈으로 보이게 만들기 위해서**입니다.

참고로 Java 25 환경에서는 Netty가 내부적으로 `Unsafe` 관련 경고를 한 번 출력할 수 있습니다. 이 경고는 이번 강의 코드의 버그라기보다, 현재 JDK와 Netty 내부 구현 사이에서 뜨는 런타임 경고로 읽는 편이 맞습니다.

## 먼저 큰 그림

blocking 서버는 이해가 가장 쉽습니다. 대신 연결 수가 늘수록 "기다리는 스레드"도 같이 늘어납니다.

```text
client
  -> accept thread
    -> connection thread (client-1)
      -> readLine() blocks
      -> write() blocks
```

Java NIO selector 서버는 스레드를 많이 만들지 않아도 됩니다. 대신 개발자가 직접 해야 하는 일이 많아집니다.

```text
client sockets
  -> selector thread
    -> accept ready?
    -> read ready?
    -> bytes -> line state
    -> write bytes back
```

Netty는 event-driven 모델 자체를 바꾸지 않습니다. 대신 사람이 직접 짜던 selector loop와 연결 상태 관리를 `EventLoop`, `Channel`, `ChannelPipeline`, codec으로 나눠 줍니다.

```text
client
  -> boss EventLoop
    -> child Channel handoff
      -> worker EventLoop
        -> LineBasedFrameDecoder
        -> StringDecoder
        -> EchoHandler
        -> StringEncoder
```

즉, Netty의 핵심은 "비동기를 처음 발명했다"가 아니라, **이미 존재하는 비동기/이벤트 기반 모델을 사람이 계속 설명하고 검증할 수 있는 구조로 정리했다**는 데 있습니다.

## 1. blocking 서버는 왜 쉬운가

[`BlockingEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/blocking/BlockingEchoServer.java)는 구조가 직관적입니다.

1. `ServerSocket.accept()`가 새 연결을 기다립니다.
2. 연결이 들어오면 전용 worker thread를 하나 잡습니다.
3. 그 thread가 `readLine()`으로 계속 기다립니다.
4. 줄 하나를 읽으면 바로 다시 써 줍니다.

이 구조가 쉬운 이유는 "기다리는 위치"가 코드에 그대로 드러나기 때문입니다. `accept()`에서 기다리고, `readLine()`에서 기다리고, `write()`에서 보냅니다. 초보자 입장에서는 가장 이해하기 쉬운 모델입니다.

하지만 비용도 분명합니다. 연결이 늘수록 기다리는 thread도 같이 늘어납니다. 한 연결이 오래 기다리면 그 thread는 계속 붙잡혀 있습니다. 즉, **코드 읽기는 쉬운데, 연결 수가 늘어나는 순간 비용 구조가 거칠게 드러나는 모델**입니다.

실행 로그도 그 사실을 그대로 보여 줍니다.

```text
#0002 ... | blocking | accept           | client-1 | thread=blocking-accept   | accepted /127.0.0.1:59505
#0003 ... | blocking | connection-start | client-1 | thread=blocking-client-1 | worker attached
#0004 ... | blocking | read             | client-1 | thread=blocking-client-1 | line="hello-blocking"
#0005 ... | blocking | write            | client-1 | thread=blocking-client-1 | echoed line="hello-blocking"
```

이 로그를 보면 새 연결을 받는 thread와, 실제로 읽고 쓰는 thread가 분리되어 있습니다. 그리고 각 연결은 자기 worker thread에 매달립니다. 이게 blocking 모델의 감각입니다.

## 2. Java NIO selector 서버는 무엇을 직접 떠안는가

[`NioSelectorEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/nio/NioSelectorEchoServer.java)는 thread 수를 확 줄입니다. 이 예제에서는 `nio-selector` thread 하나가 accept와 read를 번갈아 처리합니다.

좋은 점은 눈에 띕니다. 한 thread가 여러 연결을 돌아볼 수 있으니, blocking처럼 "연결 수 = 기다리는 thread 수"가 되지 않습니다.

하지만 개발자가 직접 떠안는 일도 바로 보입니다.

1. `Selector`를 열고 키를 등록해야 합니다.
2. `accept`와 `read` 이벤트를 직접 분기해야 합니다.
3. 연결마다 읽던 중간 상태를 attachment로 들고 다녀야 합니다.
4. 바이트를 line 단위 메시지로 조립하는 규칙을 직접 짜야 합니다.
5. 언제 채널을 닫고 키를 취소할지도 직접 관리해야 합니다.

즉, NIO는 "thread는 아끼지만 구조는 공짜가 아니다"를 보여 줍니다. 비동기 자체보다도, **비동기에서 생기는 상태 관리 비용**이 갑자기 눈앞에 나타납니다.

실행 로그를 보면 accept, read, decode, write가 모두 `nio-selector` thread 위에서 일어납니다.

```text
#0002 ... | nio | accept     | client-1 | thread=nio-selector | accepted /127.0.0.1:59509
#0003 ... | nio | read-bytes | client-1 | thread=nio-selector | bytes=10
#0004 ... | nio | read-line  | client-1 | thread=nio-selector | line="hello-nio"
#0005 ... | nio | write      | client-1 | thread=nio-selector | echoed line="hello-nio"
```

이 강의에서 아주 중요한 포인트가 여기 있습니다. `read-bytes`와 `read-line`이 분리돼 있다는 것은, "소켓에서 읽은 것"과 "애플리케이션이 이해할 메시지"가 원래는 다르다는 뜻입니다. 지금은 line protocol이라 간단하지만, 이 지점이 나중에 framing, partial read, codec 설계로 이어집니다.

## 3. Netty는 무엇을 정리해 주는가

[`NettyEchoServer.java`](../../src/main/java/io/aimpugn/learn/netty/lesson1/netty/NettyEchoServer.java)는 NIO의 event-driven 모델을 버리지 않습니다. 대신 사람이 계속 직접 다루던 경계들을 나눠 줍니다.

이 예제에서는 다음 구조를 사용합니다.

1. boss `EventLoopGroup`
   - 새 연결을 받습니다.
2. worker `EventLoopGroup`
   - 실제 읽기/쓰기와 handler 실행을 맡습니다.
3. `LineBasedFrameDecoder`
   - 바이트를 줄 단위 프레임으로 자릅니다.
4. `StringDecoder`, `StringEncoder`
   - 바이트와 문자열 변환을 분리합니다.
5. `EchoHandler`
   - "읽은 문자열을 다시 쓴다"는 핵심 비즈니스 동작만 남깁니다.

즉, Netty는 "selector를 숨긴다"기보다, **selector loop와 연결 상태, codec, handler 책임을 사람이 계속 추적 가능한 단위로 쪼개 줍니다.**

로그를 보면 이 handoff가 더 잘 보입니다.

```text
#0002 ... | netty | boss-accept    | client-pending | thread=netty-boss-6-1   | accepted child channel=[...]
#0003 ... | netty | channel-active | client-1       | thread=netty-worker-7-1 | remote=/127.0.0.1:59529
#0004 ... | netty | read           | client-1       | thread=netty-worker-7-1 | line="hello-netty"
#0005 ... | netty | write          | client-1       | thread=netty-worker-7-1 | echoed line="hello-netty"
```

이 로그를 읽을 때 핵심은 "Netty가 thread를 없앴다"가 아닙니다. 오히려 반대입니다. **새 연결을 받는 책임과, 이미 연결된 채널을 처리하는 책임을 더 선명하게 분리했다**고 읽어야 합니다.

그리고 더 중요한 차이는 코드 양보다 코드의 의미입니다. blocking/NIO에서는 "언제 기다리고, 어디에 상태를 저장하고, 바이트를 어떻게 메시지로 자를지"를 애플리케이션 코드가 직접 안고 있습니다. Netty에서는 그 중 상당 부분이 프레임워크 경계로 올라가고, handler는 더 작은 책임만 가질 수 있습니다.

## 그래서 왜 Netty가 필요한가

이제 세 예제를 한 줄로 비교해 보면 결론이 단단해집니다.

1. blocking 서버는 가장 읽기 쉽지만, 연결 수가 늘수록 기다리는 thread 비용이 그대로 드러납니다.
2. Java NIO는 thread를 아끼지만, selector loop, attachment state, framing, lifecycle 관리 부담을 개발자가 직접 짊어집니다.
3. Netty는 NIO의 모델을 유지하면서도, 그 부담을 `EventLoop`, `ChannelPipeline`, codec, handler라는 이름 있는 경계로 정리해 줍니다.

즉, Netty가 필요한 이유는 "NIO를 더 어렵게 포장해서"가 아니라, **NIO로 직접 만들면 금방 커지는 복잡도를 구조적으로 나눠 주기 때문**입니다.

이 말을 뒤집어서 읽어도 됩니다. Netty를 깊게 이해하려면 결국 다음 질문으로 내려가야 합니다.

1. boss와 worker는 왜 나뉘는가
2. `ChannelPipeline`은 inbound/outbound를 어떻게 흘리는가
3. `ByteBuf`는 왜 `byte[]`보다 복잡한가

이 강의는 바로 그 다음 단계로 넘어가기 위한 출발점입니다.

## 직접 실행

먼저 전체 예제가 깨지지 않는지 확인합니다.

```bash
mvn test
```

blocking 서버를 실행합니다.

```bash
mvn -q -DskipTests exec:java \
  -Dexec.mainClass=io.aimpugn.learn.netty.lesson1.blocking.BlockingEchoServer \
  -Dexec.args="9001"
printf 'hello-blocking\n' | nc 127.0.0.1 9001
```

Java NIO selector 서버를 실행합니다.

```bash
mvn -q -DskipTests exec:java \
  -Dexec.mainClass=io.aimpugn.learn.netty.lesson1.nio.NioSelectorEchoServer \
  -Dexec.args="9002"
printf 'hello-nio\n' | nc 127.0.0.1 9002
```

Netty 서버를 실행합니다.

```bash
mvn -q -DskipTests exec:java \
  -Dexec.mainClass=io.aimpugn.learn.netty.lesson1.netty.NettyEchoServer \
  -Dexec.args="9003"
printf 'hello-netty\n' | nc 127.0.0.1 9003
```

## PASS / FAIL 신호

PASS는 아래처럼 읽어야 합니다.

1. blocking
   - `accept`는 `blocking-accept` thread에서 보입니다.
   - 실제 `read`와 `write`는 `blocking-client-*` thread에서 보입니다.
2. NIO
   - `accept`, `read-bytes`, `read-line`, `write`가 모두 `nio-selector` thread에서 보입니다.
3. Netty
   - `boss-accept`는 `netty-boss-*` thread에서 보입니다.
   - `channel-active`, `read`, `write`는 `netty-worker-*` thread에서 보입니다.

FAIL은 아래처럼 읽으면 됩니다.

1. blocking인데 모든 연결이 하나의 worker thread처럼만 보이면 구현이 의도와 다릅니다.
2. NIO인데 연결마다 별도 worker thread가 계속 생기면 selector 모델을 제대로 쓰지 않은 것입니다.
3. Netty인데 boss/worker 구분이 전혀 보이지 않거나 codec 없이 handler가 바이트 파싱까지 전부 떠안고 있으면, Netty가 제공하는 구조적 이점을 아직 못 살린 것입니다.

## 이번 강의에서 일부러 남겨 둔 질문

이 강의는 아직 Netty의 속을 깊게 파지 않았습니다. 대신 "왜 굳이 Netty 같은 구조가 생겼는가"를 먼저 감각으로 붙였습니다.

다음 강의에서 자연스럽게 이어질 질문은 이것입니다.

**`NioEventLoopGroup(1)`의 1은 정확히 무엇을 뜻하고, boss와 worker는 실제로 어떻게 역할을 나누는가?**
