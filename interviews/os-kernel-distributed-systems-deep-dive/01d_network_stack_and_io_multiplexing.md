# 01d. Network Stack and I/O Multiplexing

분산 시스템에서 request는 갑자기 application method 안으로 생겨나지 않습니다. NIC가 frame을 받고, driver가 DMA ring의 buffer를 확인하고, interrupt 또는 polling 경로가 kernel network stack을 깨우고, TCP가 byte stream을 조립하고, socket receive buffer에 data를 넣고, epoll이 "이 fd에서 읽을 수 있다"고 application thread를 깨워야 합니다. 이 경로를 모르면 Kafka fetch latency, Cassandra read timeout, Spark shuffle fetch failure를 "네트워크 문제"라는 한 덩어리로만 말하게 됩니다.

이 문서는 packet이 request가 되고 response가 다시 packet으로 나가는 길을 먼저 그린 뒤, blocking/non-blocking, synchronous/asynchronous, readiness/completion, multiplexing을 분리합니다.

## NIC에서 application thread까지

NIC(network interface card)는 network frame을 받으면 host memory에 packet data를 놓아야 합니다. 현대 장치는 CPU가 byte마다 복사하게 하지 않고 DMA(direct memory access)를 사용해 장치가 memory buffer에 직접 쓰도록 합니다. driver와 NIC는 보통 receive ring이라는 descriptor 배열을 공유합니다. ring에는 "이 memory buffer를 packet 수신에 써도 된다"는 정보와 수신 완료 상태가 들어갑니다.

```
wire
  -> NIC receives Ethernet frame
  -> DMA writes packet bytes into RX ring buffer
  -> NIC raises interrupt or waits for polling
  -> driver/NAPI poll collects packets
  -> kernel builds skb-like packet representation
  -> IP layer checks address/routing
  -> TCP layer orders bytes, handles ACK/window/retransmission
  -> socket receive queue
  -> waiting epoll/read thread becomes runnable
```

이 trace의 핵심은 "packet 도착"과 "application read" 사이에 driver, interrupt, NAPI, softirq, TCP, socket queue, scheduler가 있다는 점입니다. NIC가 packet을 받았어도 application thread가 CPU를 받지 못하면 request handling은 늦어집니다. TCP receive queue에 byte가 쌓여도 application이 `read()`를 호출하지 않으면 backlog가 커지고 sender의 window가 줄어들 수 있습니다.

## interrupt만으로는 고속 network를 감당하기 어렵다

packet마다 interrupt를 하나씩 받으면 고속 network에서 CPU가 interrupt 처리에 압도됩니다. Linux networking은 NAPI라는 event handling 방식을 사용합니다. 기본 아이디어는 packet이 도착하면 interrupt가 NAPI poll을 schedule하고, 이후 일정 budget 안에서 여러 packet을 polling으로 처리하는 것입니다. 작업이 남으면 다시 poll되고, 끝나면 interrupt를 다시 열 수 있습니다.

```
packet arrival
  -> hardware interrupt
  -> driver schedules NAPI
  -> interrupts masked for that queue
  -> softirq or NAPI thread polls packets up to budget
  -> packets enter network stack
  -> when work complete, interrupts unmasked
```

이 구조는 latency와 throughput의 tradeoff입니다. interrupt를 줄이면 batch 처리로 throughput이 좋아질 수 있지만, packet이 application에 보이기까지 기다리는 시간이 늘 수 있습니다. busy polling은 application이 packet 도착을 기다리며 CPU를 태우는 대신 latency를 줄이려는 선택입니다. 일반 서버에서는 무조건 켜는 최적화가 아니라 workload와 CPU 여유를 보고 판단해야 합니다.

## TCP는 packet이 아니라 신뢰할 수 있는 byte stream을 제공하려 한다

application은 보통 TCP socket에서 byte stream을 읽습니다. TCP는 packet loss를 재전송하고, sequence number로 순서를 맞추고, receiver window와 congestion control로 속도를 조절합니다. 따라서 application이 `send()`한 한 덩어리가 상대 `read()` 한 번과 1:1로 대응된다고 생각하면 안 됩니다. TCP는 message boundary를 보존하지 않습니다.

```
sender write 10KB
  -> TCP segments into packets
  -> network may reorder/drop packets
  -> receiver TCP reorders and acknowledges
  -> application read may receive 4KB, then 6KB
```

Kafka protocol, Cassandra native protocol, Spark RPC/shuffle protocol은 이 byte stream 위에 자기 message framing을 얹습니다. 그래서 request parser는 "한 번 read했더니 요청 하나가 온다"가 아니라, buffer에 byte를 모아 length field나 protocol delimiter를 보고 message boundary를 찾아야 합니다.

## blocking과 non-blocking은 thread가 기다리는 방식의 차이다

blocking socket에서 `read()`를 호출했는데 받을 byte가 없으면, calling thread는 kernel 안에서 sleep 상태로 들어갈 수 있습니다. CPU는 다른 thread에게 넘어갑니다. 나중에 socket receive queue에 data가 들어오면 kernel은 기다리던 thread를 깨워 runnable로 만듭니다.

non-blocking socket에서는 받을 byte가 없을 때 thread를 sleep시키지 않고 `EAGAIN`/`EWOULDBLOCK` 같은 결과로 바로 돌아옵니다. 그러면 application은 다른 fd를 보거나 event loop로 돌아갈 수 있습니다.

```
blocking read:
  read(fd)
    -> no data
    -> thread sleeps in kernel wait queue
    -> data arrives
    -> thread wakes and continues

non-blocking read:
  read(fd)
    -> no data
    -> return EAGAIN
    -> application decides what to do next
```

non-blocking은 "항상 빠르다"가 아닙니다. 기다리는 책임이 kernel에서 application event loop로 이동합니다. 잘못 쓰면 busy loop로 CPU를 태우거나, readiness event를 놓치거나, partial read/write를 처리하지 못해 protocol bug가 생깁니다.

## synchronous/asynchronous는 완료를 누가 언제 알려 주는가의 축이다

blocking/non-blocking은 호출한 thread가 기다리느냐의 축입니다. synchronous/asynchronous는 작업 완료를 어떤 방식으로 다루느냐의 축입니다. synchronous call은 호출 흐름이 그 작업의 완료 또는 실패를 직접 확인하고 다음으로 갑니다. asynchronous model은 요청을 제출하고, 완료는 callback, future, completion queue, event 등 별도 경로로 받습니다.

이 두 축은 섞이기 쉽습니다.

| 축 | 질문 | 예 |
|---|---|---|
| blocking vs non-blocking | 지금 호출한 thread가 준비될 때까지 잠드는가? | blocking `read()`, non-blocking `read()` returning `EAGAIN` |
| synchronous vs asynchronous | 작업 완료를 호출 흐름에서 직접 받는가, 나중에 별도 경로로 받는가? | synchronous `write()` result, async completion callback |
| readiness vs completion | "할 수 있음"을 알려 주는가, "끝났음"을 알려 주는가? | epoll readiness, io_uring completion |

`epoll`은 흔히 async I/O처럼 말하지만 정확히는 readiness notification입니다. epoll은 "이 fd에서 읽거나 쓸 수 있을 가능성이 있다"를 알려 줍니다. 실제 `read()`/`write()`는 application이 호출해야 하고, partial read/write와 `EAGAIN`을 처리해야 합니다. 반대로 completion-based I/O는 작업을 제출하고 완료 queue에서 "이 작업이 끝났다"를 받는 모델에 가깝습니다. Linux의 io_uring은 여러 작업을 submission queue에 넣고 completion queue로 결과를 받는 구조를 제공합니다. 다만 모든 작업이 언제나 완전한 비동기 device I/O로 내려가는 것은 아니며 kernel thread나 fallback 경로가 개입할 수 있으므로 세부는 operation과 kernel version을 확인해야 합니다.

## multiplexing은 많은 fd를 적은 thread로 기다리는 방법이다

select, poll, epoll, kqueue는 여러 file descriptor의 준비 상태를 기다리는 도구입니다. 오래된 `select()`는 fd set 크기 제한과 매 호출마다 set을 넘기는 비용이 있고, `poll()`은 배열을 훑습니다. Linux `epoll`은 관심 fd 집합을 kernel 객체에 등록하고 ready list를 통해 ready event를 받는 방식으로 많은 connection에서 더 유리합니다. macOS/BSD 계열에는 kqueue가 있습니다.

```
event loop
  epoll_wait()
    -> returns fd 7 readable, fd 9 writable
  read(fd 7) until EAGAIN or protocol frame complete
  write(fd 9) until send buffer full or response done
  update interest set
  epoll_wait()
```

edge-triggered epoll은 상태 변화가 있을 때 event를 주므로, non-blocking fd를 끝까지 drain하지 않으면 다시 event를 못 받을 수 있습니다. level-triggered epoll은 조건이 계속 참이면 반복해서 event를 줍니다. edge-triggered가 더 고급이라 항상 좋다는 식으로 외우면 안 됩니다. application이 fd를 어떻게 drain하고 backpressure를 어떻게 표현하는지가 핵심입니다.

## accept queue와 listen backlog는 연결이 application에 도달하기 전의 대기열이다

server socket은 `listen()` 상태에서 connection을 받습니다. TCP handshake와 accept 대기열은 application이 `accept()`로 connection fd를 가져가기 전의 kernel 상태입니다. backlog가 가득 차거나 application accept loop가 느리면 client connection이 실패하거나 지연될 수 있습니다.

```
client SYN
  -> kernel TCP handshake state
  -> established connection enters accept queue
  -> application accept()
  -> new connected socket fd
  -> registered to epoll or handed to worker
```

이 경로 때문에 "server thread가 바쁘다"와 "kernel accept queue가 찼다"를 구분해야 합니다. Kafka broker, Cassandra node, Spark driver/executor RPC endpoint 모두 connection accept와 socket buffer, event loop 영향을 받습니다.

## response가 나가는 길도 queue와 flow control을 지난다

application이 response를 만들고 `write()`나 `send()`를 호출하면 byte는 socket send buffer에 들어갑니다. TCP는 receiver window와 congestion window를 보며 segment를 내보냅니다. qdisc(queueing discipline)와 NIC transmit queue를 지나 DMA로 NIC에 전달되고, NIC가 frame을 wire로 보냅니다.

```
application response bytes
  -> send()
  -> socket send buffer
  -> TCP segmentation / retransmission state
  -> qdisc / traffic shaping
  -> NIC TX ring
  -> DMA to device
  -> wire
```

상대가 느리게 읽으면 receiver window가 줄고, sender의 socket send buffer가 차며, non-blocking write는 `EAGAIN`을 반환할 수 있습니다. blocking write는 buffer가 비기를 기다리며 sleep할 수 있습니다. 이 상황을 application-level backpressure와 연결하지 않으면 retry와 thread 증가로 장애를 키우기 쉽습니다.

## backend request path 전체 trace

아래 trace를 하나로 말할 수 있어야 Kafka/Cassandra/Spark network 병목을 하부에서 설명할 수 있습니다.

```
1. client sends TCP segments
2. server NIC DMAs frames into RX ring
3. interrupt schedules NAPI polling
4. driver/NAPI hands packets to kernel network stack
5. TCP validates sequence, ACK/window, reassembles byte stream
6. bytes enter socket receive queue
7. epoll_wait returns "fd readable"
8. application thread becomes runnable and gets CPU
9. application read() pulls bytes and parser builds a request
10. user code reads/writes files, memory, locks, downstream systems
11. response bytes enter socket send buffer
12. TCP segments, qdisc/NIC TX queue sends frames
13. client receives response or timeout observes uncertainty
```

이 중 하나만 밀려도 application latency가 늘 수 있습니다. NIC interrupt affinity, NAPI budget, softirq backlog, socket receive queue, event loop CPU, parser allocation, downstream disk I/O, response send buffer가 모두 후보입니다.

## 제품으로 다시 연결하기

- Kafka producer request와 consumer fetch는 broker network path를 지나고, follower replica fetch도 같은 network stack과 disk path를 경쟁합니다. broker socket queue가 밀리면 produce/fetch timeout이 application-level 오류로 올라옵니다.
- Cassandra coordinator와 replica 사이의 read/write, gossip, repair streaming은 모두 TCP 위에 있습니다. read timeout이 곧 node crash라는 뜻은 아닙니다. socket queue, GC pause, compaction I/O, network retransmission이 모두 양립할 수 있습니다.
- Spark shuffle fetch는 executor 사이에서 많은 block을 가져오는 network + disk + serialization 작업입니다. fetch failure는 remote executor loss일 수도 있지만, network timeout, local disk spill, GC pause, connection pool pressure일 수도 있습니다.

문서를 덮고 다음 구분을 말해 보세요.

- blocking과 non-blocking은 호출 thread가 잠드는지의 축입니다.
- synchronous와 asynchronous는 완료를 어떤 경로로 받는지의 축입니다.
- epoll은 readiness를 알려 주며, 실제 read/write와 partial 처리 책임은 application에 남습니다.
- packet 도착과 request 처리 사이에는 driver, NAPI, TCP, socket queue, scheduler, event loop가 있습니다.
