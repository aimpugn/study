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

blocking socket에서 `read()`를 호출했는데 받을 byte가 없으면, calling thread는 kernel 안에서 sleep 상태로 들어갈 수 있습니다. CPU는 다른 thread에게 넘어갑니다. socket receive queue에 data가 들어오면 kernel은 기다리던 thread를 깨워 runnable로 만듭니다.

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
| synchronous vs asynchronous | 작업 완료를 호출 흐름에서 직접 받는가, 별도 경로로 받는가? | synchronous `write()` result, async completion callback |
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

## NIC queue, RSS, interrupt affinity는 request가 처음 줄을 서는 자리다

네트워크 병목을 application thread 수로만 보면 첫 대기열을 놓칩니다. Packet은 먼저 NIC에 도착합니다. 현대 NIC는 여러 RX/TX queue를 가지고, RSS(receive-side scaling) 같은 방식으로 flow를 여러 queue에 나눌 수 있습니다. 각 queue는 특정 CPU interrupt나 NAPI poll context와 연결될 수 있고, 이 연결이 어긋나면 한 CPU만 softirq로 바쁘고 application worker는 다른 CPU에서 cache miss와 wakeup 비용을 겪을 수 있습니다.

```
incoming flow
  -> NIC hash chooses RX queue
  -> DMA writes frame into memory ring
  -> interrupt or polling schedules NAPI
  -> CPU runs driver poll
  -> packet enters kernel network stack
```

Interrupt affinity는 어느 CPU가 device interrupt를 처리할지 정하는 설정입니다. Receive packet steering이나 transmit packet steering 같은 kernel 기능은 packet processing을 여러 CPU로 분산하려고 합니다. 하지만 무조건 분산이 좋은 것은 아닙니다. Packet을 처리하는 CPU, application thread가 실행되는 CPU, memory가 놓인 NUMA node가 너무 멀어지면 data가 계속 CPU 사이를 이동합니다. Low-latency 환경에서는 NIC queue, IRQ affinity, NAPI budget, application worker affinity를 함께 맞추려는 이유가 여기 있습니다.

Kafka broker는 producer/fetch traffic과 replica fetch traffic이 동시에 들어옵니다. Cassandra node는 client read/write, gossip, repair streaming을 함께 받습니다. Spark executor는 shuffle fetch가 특정 host/partition에 몰릴 수 있습니다. 이 traffic들이 NIC queue와 softirq CPU를 공유하면 application metric에는 "request timeout"이나 "fetch failure"로 올라오지만, 첫 병목은 packet processing queue일 수 있습니다. `ethtool -S`, `/proc/interrupts`, `softnet_stat`, eBPF network trace, `sar -n`, NIC driver stats는 이 첫 줄을 보는 도구입니다.

## TCP는 순서를 만들지만 지연도 만든다

TCP는 application에게 reliable byte stream을 제공합니다. Packet이 유실되면 재전송하고, 순서가 바뀌면 다시 정렬하며, receiver가 감당할 수 있는 만큼만 보내도록 flow control을 합니다. Congestion control은 네트워크 전체 혼잡을 추정해 sending rate를 조절합니다. 덕분에 application은 UDP packet loss를 직접 다루지 않아도 되지만, 대신 head-of-line blocking과 retransmission delay, buffer growth, delayed ACK, congestion window 변화 같은 현상을 만나게 됩니다.

```
sender writes bytes
  -> TCP segments them
  -> packets travel and may be lost/reordered
  -> receiver ACKs contiguous byte sequence
  -> missing segment delays later bytes delivery to application
  -> retransmission recovers loss
```

TCP는 message boundary를 보존하지 않습니다. Application이 100 byte를 한 번 쓰고 200 byte를 한 번 썼다고 receiver가 같은 두 번의 read로 받는다는 보장은 없습니다. Kafka, Cassandra, Spark는 각자 protocol framing을 두어 byte stream 안에서 request/response boundary를 찾습니다. 그래서 `read()` 한 번이 곧 하나의 request라는 코드는 깨집니다. Partial read와 partial write를 처리하는 parser와 output buffer가 필요합니다.

Timeout을 볼 때도 TCP 계층을 건너뛰면 안 됩니다. 상대가 죽었는지, network packet이 유실되었는지, receiver socket buffer가 꽉 찼는지, sender congestion window가 줄었는지, application thread가 읽지 못해 receive buffer가 차는지 모두 timeout으로 보일 수 있습니다. Cassandra read timeout은 replica process crash가 아니라 GC pause나 compaction I/O, socket backlog, retransmission과도 양립합니다. Spark shuffle fetch timeout도 remote executor loss뿐 아니라 network congestion과 local disk stall을 함께 봐야 합니다.

## Listen backlog와 accept queue는 서버가 연결을 받기 전의 대기실이다

서버가 `listen(fd, backlog)`를 호출하면 kernel은 incoming connection을 관리할 queue를 준비합니다. TCP handshake 과정의 SYN backlog와 handshake가 끝난 뒤 application이 `accept()`하기 전의 accept queue를 구분해야 합니다. Application이 바쁘거나 single accept loop가 밀리면 handshake가 끝난 connection도 application socket으로 넘어가지 못하고 queue에 머물 수 있습니다. Queue가 차면 client는 connection timeout이나 reset을 볼 수 있습니다.

```
client connect()
  -> SYN arrives
  -> SYN backlog / handshake state
  -> established connection enters accept queue
  -> application accept()
  -> new connected socket fd
```

Backlog 문제는 "thread pool을 늘리자"로 바로 해결되지 않을 수 있습니다. Accept loop가 CPU를 못 받는지, TLS handshake가 user-space에서 오래 걸리는지, connection rate가 burst인지, SYN flood mitigation이 작동하는지, file descriptor limit이 낮은지, per-process accept가 느린지 봐야 합니다. Kafka broker listener, Cassandra native transport, Spark driver/executor RPC endpoint 모두 이 queue를 지나며, Kubernetes service나 load balancer 앞에서는 또 다른 connection queue와 health check가 붙습니다.

관측은 `ss -ltn`, `ss -s`, kernel TCP counters, application accept latency, load balancer metrics를 함께 봅니다. Established connection 수가 많다고 모두 application이 처리 중인 것은 아닙니다. Accept queue에 쌓여 있거나, application이 accepted했지만 protocol handshake를 기다리거나, TLS handshake에서 CPU를 쓰고 있을 수 있습니다.

## Readiness와 completion은 다른 약속이다

`select`, `poll`, `epoll`, `kqueue` 같은 readiness API는 "지금 read/write를 시도하면 blocking 없이 어느 정도 진행될 가능성이 있다"는 상태를 알려 줍니다. 이것은 작업이 완료되었다는 뜻이 아닙니다. Readable event를 받고 `read()`를 호출했더니 partial data만 받을 수 있고, writable event를 받고 `write()`를 호출했더니 send buffer에 일부만 들어갈 수 있습니다. Application은 남은 data와 parser state를 직접 관리해야 합니다.

Completion API는 "요청한 작업이 완료되었다"는 결과를 돌려주는 모델입니다. Linux `io_uring`은 file I/O와 socket I/O에서 submission queue와 completion queue를 통해 더 넓은 asynchronous I/O 모델을 제공합니다. 하지만 `io_uring`을 쓴다고 모든 비용이 사라지는 것은 아닙니다. Buffer lifetime, cancellation, ordering, backpressure, kernel version, operation support, security policy가 모두 중요합니다.

```
readiness model
  epoll says fd readable
  -> application calls read
  -> application handles bytes and EAGAIN

completion model
  application submits operation
  -> kernel completes later
  -> completion queue entry reports result
```

`kqueue`는 BSD/macOS 계열에서 쓰이는 event notification interface입니다. Linux의 `epoll`과 이름과 세부 semantics가 다릅니다. 따라서 portability가 필요한 runtime은 플랫폼별 backend를 감춥니다. Java NIO selector, Netty event loop, libuv, tokio 같은 runtime은 OS별 readiness/completion interface 위에 추상화를 올립니다. 하지만 추상화가 있다고 해서 partial read/write, backpressure, event loop blocking 문제가 사라지지 않습니다.

## Event loop는 기다림을 줄이지만 막히면 전체가 막힌다

Event loop 기반 서버는 적은 thread로 많은 connection을 다룹니다. 핵심은 "할 일이 생긴 fd"만 깨워 처리하고, I/O가 준비되지 않은 connection 때문에 thread가 잠들지 않게 하는 것입니다. 하지만 event loop thread 안에서 CPU를 오래 쓰거나 blocking call을 호출하면 같은 loop가 담당하는 다른 connection도 기다립니다. 그래서 Netty나 Kafka network thread 같은 event loop 계열 구조에서는 loop 안의 작업을 짧게 유지하고, 무거운 작업을 worker pool이나 별도 queue로 넘깁니다.

```
event loop tick
  -> epoll_wait returns readable/writable fds
  -> read bytes
  -> parse frames
  -> enqueue request to worker
  -> write ready responses
  -> return to epoll_wait quickly
```

Kafka broker의 network thread는 socket read/write와 request queue enqueue를 담당하고, request handler가 실제 처리의 많은 부분을 맡습니다. 이 분리가 깨져 network thread가 오래 CPU를 쓰면 모든 connection의 latency가 오릅니다. Cassandra native transport도 event loop와 request executor 사이의 분리를 갖습니다. Spark RPC나 shuffle service도 event loop가 block되면 많은 fetch와 control message가 함께 느려질 수 있습니다.

Backpressure는 event loop에서 특히 중요합니다. Downstream worker queue가 가득 찼는데 event loop가 계속 socket에서 읽기만 하면 memory가 늘고 timeout이 뒤늦게 터집니다. 반대로 읽기를 멈추거나 interest를 조절하면 TCP receive window가 줄어 sender가 자연스럽게 느려질 수 있습니다. 좋은 서버는 "받을 수 있는 만큼만 읽고, 보낼 수 있는 만큼만 쓰고, downstream queue 상태를 network interest와 연결"합니다.

## Send path와 qdisc는 응답도 줄을 세운다

응답을 보낼 때도 application이 `write()`를 호출한다고 곧장 wire로 나가지 않습니다. Bytes는 socket send buffer에 들어가고, TCP가 segment를 만들고, qdisc(queueing discipline)가 packet을 queueing/shaping할 수 있으며, NIC TX queue와 DMA를 거쳐 wire로 나갑니다. Receiver가 느리거나 network가 혼잡하면 send buffer가 차고, non-blocking write는 `EAGAIN`을 반환합니다. Blocking write는 잠들 수 있습니다.

```
application response
  -> write/send
  -> socket send buffer
  -> TCP segmentation and congestion window
  -> qdisc / traffic control
  -> NIC TX ring
  -> DMA to device
  -> wire
```

Qdisc는 traffic shaping, fairness, queue management를 위해 존재합니다. Kubernetes나 service mesh, host-level traffic control, cloud networking layer가 붙으면 application은 local socket write만 보지만 실제 path에는 더 많은 queue가 있습니다. Kafka replication traffic이 client traffic과 같은 NIC를 공유하고, Cassandra repair streaming이 foreground reads와 경쟁하고, Spark shuffle이 executor RPC를 밀어내는 상황에서 send path queue를 봐야 하는 이유입니다.

Zero-copy 전송도 이 자리에서 이해할 수 있습니다. `sendfile()` 같은 경로는 file data를 user-space로 복사했다가 다시 kernel socket buffer로 복사하는 일을 줄일 수 있습니다. Kafka가 page cache와 file-to-socket 전송을 중요하게 보는 이유입니다. 하지만 TLS를 적용하면 암호화 때문에 data가 user-space나 crypto path를 거쳐야 할 수 있고, compression이나 record transformation도 zero-copy 이점을 줄입니다. "zero-copy니까 CPU가 없다"가 아니라 "불필요한 copy를 줄이는 특정 경로"로 이해해야 합니다.

## Network backpressure는 timeout보다 먼저 읽어야 한다

Backpressure는 downstream이 감당할 수 없을 때 upstream이 속도를 줄이게 만드는 피드백입니다. 네트워크에서는 receiver window, send buffer fullness, application read pause, event interest 조절, queue limit, request quota가 모두 backpressure가 될 수 있습니다. Backpressure가 없으면 queue가 계속 커지고, memory가 늘고, timeout과 retry가 더 많은 traffic을 만들어 장애를 키웁니다.

```
downstream disk is slow
  -> request handler queue grows
  -> network thread should slow reads or reject early
  -> client sees bounded latency/error

without backpressure
  -> network keeps accepting
  -> memory buffers grow
  -> timeouts trigger retries
  -> more traffic arrives
```

Kafka는 producer quota, request queue, socket buffer, replication lag, consumer lag가 backpressure 신호가 됩니다. Cassandra는 coordinator timeout, pending tasks, dropped mutations, stream throughput, compaction backlog가 신호입니다. Spark는 shuffle fetch wait, executor lost, task retry, streaming backpressure와 source rate limit이 관련됩니다. 네트워크 문서는 socket API에서 끝나는 것이 아니라, product-level queue와 retry policy까지 이어져야 합니다.

## Interview replay: packet에서 request까지, response에서 timeout까지

네트워크 면접 답변은 먼저 짧게 시작할 수 있습니다. "Packet은 NIC에서 DMA로 memory ring에 들어오고, interrupt나 NAPI polling을 통해 kernel network stack으로 올라오며, TCP는 byte stream을 재조립해 socket receive buffer에 넣습니다. `epoll`은 fd가 읽을 수 있음을 알려 주고, application thread는 scheduler에게 CPU를 받은 뒤 `read()`로 bytes를 가져와 protocol frame을 만듭니다." 이 문장에 send path를 붙이면 "응답 bytes는 socket send buffer, TCP congestion/flow control, qdisc, NIC TX queue를 거쳐 나갑니다"가 됩니다.

꼬리 질문은 구분으로 닫습니다. Blocking/non-blocking은 thread가 기다리는 방식입니다. Sync/async는 완료 결과를 받는 방식입니다. Readiness와 completion은 다른 interface입니다. TCP timeout은 node crash의 증명이 아닙니다. Packet 도착은 request 처리 완료가 아닙니다. Event loop는 많은 fd를 효율적으로 기다리지만, loop thread가 막히면 많은 connection이 함께 느려집니다. Backpressure가 없으면 timeout과 retry가 시스템을 더 아프게 만들 수 있습니다.

## UDP와 TCP를 제품 의미로 비교하기

UDP는 message boundary가 있는 datagram을 보내지만 delivery, ordering, retransmission을 보장하지 않습니다. TCP는 reliable ordered byte stream을 제공하지만 message boundary를 보존하지 않고 head-of-line blocking을 만들 수 있습니다. 둘 중 무엇이 "빠르다"가 아니라, 어떤 보장을 transport에 맡기고 어떤 보장을 application protocol이 맡을지의 선택입니다.

Kafka, Cassandra, Spark의 핵심 data path는 일반적으로 TCP 기반 protocol을 사용합니다. Record, mutation, shuffle block은 유실되면 안 되고, application protocol이 retry와 ordering, framing을 얹습니다. 반면 service discovery나 일부 telemetry, custom low-latency protocol은 UDP나 QUIC 같은 다른 선택을 할 수 있습니다. UDP를 쓰면 application이 loss, duplication, reordering, congestion control을 더 많이 책임져야 합니다. TCP를 쓰면 kernel TCP stack의 flow/congestion control과 buffer semantics를 이해해야 합니다.

```
TCP application protocol
  -> define frame length/header
  -> handle partial read/write
  -> rely on TCP ordering/retransmission
  -> still define request timeout and idempotency

UDP application protocol
  -> each datagram has boundary
  -> application handles loss/reorder/duplicate
  -> congestion responsibility moves upward
```

분산 시스템에서 transport timeout은 항상 불확실성을 남깁니다. TCP connection이 끊어졌다고 remote operation이 적용되지 않았다는 뜻은 아닙니다. UDP response가 사라졌다고 server가 처리하지 않았다는 뜻도 아닙니다. 그래서 네트워크 장은 idempotency, retry, deduplication, quorum, offset commit 같은 분산 시스템 장과 맞물립니다.

## TLS와 compression은 network path를 다시 CPU path로 끌어온다

Plain TCP send path에서는 file-to-socket 최적화나 큰 buffer가 유리할 수 있습니다. 하지만 TLS를 적용하면 bytes를 암호화하고 인증 tag를 계산해야 합니다. Compression도 CPU를 써서 network bytes를 줄이는 tradeoff입니다. Network가 병목일 때 compression은 도움이 될 수 있지만, CPU가 병목이면 오히려 latency를 키울 수 있습니다. TLS termination 위치가 application process인지 sidecar인지 load balancer인지에 따라 CPU 사용 위치와 관측 지점도 달라집니다.

```
response bytes
  -> optional compression
  -> TLS record encryption
  -> socket send buffer
  -> TCP segmentation
  -> NIC
```

Kafka는 compression batch와 TLS 설정에 따라 CPU/network/disk tradeoff가 바뀝니다. Cassandra internode encryption은 repair/streaming과 foreground traffic CPU를 증가시킬 수 있습니다. Spark shuffle encryption/compression은 network bytes를 줄이는 대신 executor CPU를 씁니다. 따라서 network latency를 볼 때 CPU profile을 함께 보지 않으면 "네트워크가 느리다"는 결론이 틀릴 수 있습니다.

## Packet capture와 socket 지표를 함께 읽기

`tcpdump`는 packet을 보여 주지만 application queue를 직접 보여 주지 않습니다. `ss`는 socket state와 send/receive queue를 보여 주지만 wire에서 실제로 어떤 retransmission이 일어났는지까지 모두 설명하지는 않습니다. Product metric은 request latency와 timeout을 보여 주지만 kernel queue를 숨깁니다. 세 표면을 함께 놓아야 흐름이 닫힙니다.

```
product timeout
  -> ss shows send queue growing?
  -> tcpdump shows retransmission or zero window?
  -> thread dump shows event loop blocked?
  -> iostat shows downstream disk slow?
```

Receiver zero window가 보이면 상대 application이 bytes를 빨리 읽지 못해 receive buffer가 찬 상황을 의심합니다. Retransmission이 많으면 network loss나 congestion을 의심합니다. Send queue가 크고 packet은 잘 나가는데 response가 늦으면 remote application이나 downstream storage를 봅니다. Receive queue가 쌓이면 local application thread가 읽지 못하는 이유를 봅니다. 이런 식으로 packet, socket, thread, disk를 연결하면 network debugging이 단일 도구 의존에서 벗어납니다.
