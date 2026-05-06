# 핵심 인터뷰 정리

이 문서는 `interviews` 디렉터리의 세부 대주제 문서를 대체하지 않는다.
역할은 더 좁다.
면접에서 짧은 질문 하나가 들어왔을 때, 여러 기술 단위를 한 번에 엮어 답변할 수 있도록 핵심 질문 묶음과 꼬리 질문 경로를 빠르게 복원하는 데 있다.

세부 문서는 각각 한 주제를 깊게 파고든다.
이 문서는 그 세부 주제를 실전 답변 순서로 다시 묶는다.
좋은 답변은 처음 30초 안에 핵심 판단을 말하고, 그 뒤에 운영체제, 런타임, 프레임워크, 네트워크, 데이터베이스, 분산 시스템, 운영 관측으로 필요한 만큼 내려간다.

## 목차

- [빠른 사용법](#빠른-사용법)
- [시간이 없을 때 먼저 볼 질문](#시간이-없을-때-먼저-볼-질문)
- [모든 답변에 깔리는 기본 프레임](#모든-답변에-깔리는-기본-프레임)
- [1. 웹 요청 하나는 OS, 프록시, 톰캣, 스프링, DB를 어떻게 지나가는가](#1-웹-요청-하나는-os-프록시-톰캣-스프링-db를-어떻게-지나가는가)
- [2. 트랜잭션은 ACID에서 MVCC, 격리 수준, 전파, 2PC, 보상 트랜잭션까지 어떻게 이어지는가](#2-트랜잭션은-acid에서-mvcc-격리-수준-전파-2pc-보상-트랜잭션까지-어떻게-이어지는가)
- [3. 인덱스는 왜 B+Tree를 쓰고 실제 데이터 읽기와 어떻게 연결되는가](#3-인덱스는-왜-btree를-쓰고-실제-데이터-읽기와-어떻게-연결되는가)
- [4. 수억 건 테이블은 인덱스만으로 해결되지 않을 때 어떻게 운영하는가](#4-수억-건-테이블은-인덱스만으로-해결되지-않을-때-어떻게-운영하는가)
- [5. 동시성 문제는 race condition, visibility, ordering, deadlock으로 어떻게 나뉘는가](#5-동시성-문제는-race-condition-visibility-ordering-deadlock으로-어떻게-나뉘는가)
- [6. blocking, non-blocking, async, event loop는 서로 어떻게 다른가](#6-blocking-non-blocking-async-event-loop는-서로-어떻게-다른가)
- [7. `java -jar` 이후 JVM, class loader, Spring Boot, AOP proxy는 어떤 순서로 동작하는가](#7-java--jar-이후-jvm-class-loader-spring-boot-aop-proxy는-어떤-순서로-동작하는가)
- [8. Kafka, RabbitMQ, outbox, idempotent consumer는 왜 함께 물리는가](#8-kafka-rabbitmq-outbox-idempotent-consumer는-왜-함께-물리는가)
- [9. 분산 시스템에서 strong consistency, eventual consistency, replication, quorum은 어떻게 엮이는가](#9-분산-시스템에서-strong-consistency-eventual-consistency-replication-quorum은-어떻게-엮이는가)
- [10. TCP, HTTP, keep-alive, TLS, reverse proxy는 한 요청 경로에서 어떻게 만나는가](#10-tcp-http-keep-alive-tls-reverse-proxy는-한-요청-경로에서-어떻게-만나는가)
- [11. TLS와 암호학 질문은 인증, 키 교환, 암호화, 해시를 어떻게 분리해야 하는가](#11-tls와-암호학-질문은-인증-키-교환-암호화-해시를-어떻게-분리해야-하는가)
- [12. JVM GC와 런타임 성능은 heap, allocation, pause, throughput을 어떻게 이어 말하는가](#12-jvm-gc와-런타임-성능은-heap-allocation-pause-throughput을-어떻게-이어-말하는가)
- [13. 장애 분석 질문은 관측 지표에서 병목 계층으로 어떻게 좁혀 가는가](#13-장애-분석-질문은-관측-지표에서-병목-계층으로-어떻게-좁혀-가는가)
- [14. 알고리즘과 코드 품질 질문은 복잡도, 자료구조, 불변식, 테스트 가능성으로 어떻게 답하는가](#14-알고리즘과-코드-품질-질문은-복잡도-자료구조-불변식-테스트-가능성으로-어떻게-답하는가)
- [복합 질문으로 연습하기](#복합-질문으로-연습하기)
- [마지막 점검 질문](#마지막-점검-질문)

## 빠른 사용법

이 문서는 암기 문장 모음이 아니라 답변 경로 지도다.
한 질문을 받으면 아래 순서로 말한다.

1. 먼저 해결하려는 문제를 말한다.

    예를 들어 트랜잭션 질문이면 "여러 변경을 하나의 논리적 작업으로 묶고, 동시에 실행되는 작업이 서로 망치지 않게 하며, 장애 뒤에도 복구 가능한 상태를 남기는 장치입니다"처럼 답한다.
    용어 정의보다 문제를 먼저 잡으면 꼬리 질문이 들어와도 답변 방향을 잃지 않는다.

2. 핵심 불변식을 말한다.

    불변식은 깨지면 안 되는 약속이다.
    인덱스라면 "정렬된 작은 구조를 통해 전체 테이블을 매번 읽지 않는다"가 불변식이고, 메시징이라면 "생산자와 소비자의 시간, 장애, 처리 속도를 분리한다"가 불변식이다.

3. 실제 경로를 낮은 층부터 위로 올린다.

    면접관이 깊게 들어오면 "애플리케이션이 이렇게 합니다"에서 끝내지 말고, 소켓, 파일 디스크립터, 커널 대기열, 런타임 스레드, 커넥션 풀, WAL, 인덱스 페이지처럼 실제로 상태가 바뀌는 지점을 짚는다.

4. 트레이드오프와 실패 신호로 마무리한다.

    좋은 답변은 "항상 좋다"가 아니라 "이 조건에서는 이기고, 이 조건에서는 비용이 생깁니다"까지 말한다.
    마지막에는 `EXPLAIN`, GC log, thread dump, slow query log, packet capture, broker lag, p99 latency처럼 확인 경로를 붙인다.

## 시간이 없을 때 먼저 볼 질문

시간이 정말 촉박하면 아래 8개를 먼저 복습한다.
이 8개는 서로 연결되어 있어서, 하나를 제대로 답하면 여러 꼬리 질문으로 확장할 수 있다.

1. 웹 요청 하나가 들어와 DB에 저장되고 응답되기까지 어떤 계층을 지나가는가.
2. 트랜잭션 격리 수준은 MVCC, lock, phantom read, Spring 전파, 2PC, 보상 트랜잭션과 어떻게 이어지는가.
3. 인덱스는 왜 B+Tree이고, 커버링 인덱스와 실제 row fetch는 어떻게 다른가.
4. blocking, non-blocking, async, event loop, epoll은 각각 무엇을 해결하고 무엇을 해결하지 못하는가.
5. `@Transactional`은 왜 프록시를 거쳐야 하고, self-invocation에서는 왜 깨질 수 있는가.
6. Kafka와 RabbitMQ는 로그와 큐라는 모델 차이 때문에 순서, 재처리, 확장 방식이 어떻게 달라지는가.
7. TLS 핸드셰이크에서 인증서와 ECDHE가 각각 무엇을 해결하는가.
8. 장애가 났을 때 p99 latency, CPU, memory, GC, thread, DB, network, broker lag 중 어디서부터 좁혀 가는가.

## 모든 답변에 깔리는 기본 프레임

면접 답변은 보통 아래 다섯 문장 뼈대로 정리하면 덜 흔들린다.

```text
이 질문은 결국 [문제]를 해결하는 방식에 관한 질문입니다.
핵심 약속은 [불변식]입니다.
실제로는 [낮은 계층]에서 [중간 계층]을 거쳐 [애플리케이션 계층]으로 올라옵니다.
대신 [트레이드오프]가 생기므로 [조건]에서는 다른 선택이 낫습니다.
확인은 [관측 지표, 명령, 로그, 실행 계획]으로 합니다.
```

이 뼈대는 답변을 기계적으로 만들기 위한 문장이 아니다.
면접관이 꼬리 질문을 던졌을 때 "어느 계층의 질문인지" 빠르게 분류하기 위한 기준이다.

## 1. 웹 요청 하나는 OS, 프록시, 톰캣, 스프링, DB를 어떻게 지나가는가

**첫 30초 답변**

웹 요청은 브라우저나 클라이언트 코드에서 HTTP 메시지로 만들어지고, TCP 연결과 필요하면 TLS 핸드셰이크를 거쳐 서버 커널의 소켓 버퍼에 도착한다.
Nginx 같은 리버스 프록시는 그 요청을 파싱해 upstream인 Tomcat이나 Spring Boot 내장 서버로 다시 전달하고, 서블릿 컨테이너는 요청을 `DispatcherServlet`으로 넘긴다.
그 뒤 컨트롤러, 서비스, 트랜잭션, 커넥션 풀, DB 실행 계획을 거쳐 결과가 반대 방향으로 돌아간다.

**이어 말할 순서**

1. 클라이언트는 DNS, TCP 3-way handshake, TLS 핸드셰이크를 거쳐 서버와 연결을 만든다.
2. 서버 커널은 NIC에서 받은 패킷을 TCP/IP 스택으로 처리하고, 포트에 바인딩된 리스닝 소켓과 연결별 소켓으로 나눈다.
3. Nginx는 `accept()`된 연결에서 HTTP 요청을 읽고, 설정에 따라 upstream으로 새 HTTP 요청을 만든다.
4. Tomcat은 connector thread나 executor를 통해 요청을 받고, 서블릿 컨테이너 안에서 `DispatcherServlet`까지 전달한다.
5. Spring MVC는 handler mapping, argument binding, validation, controller, service, repository 흐름으로 이동한다.
6. DB 접근이 필요하면 HikariCP 같은 커넥션 풀에서 커넥션을 빌리고, 트랜잭션 경계 안에서 SQL을 실행한다.
7. 응답은 DB result set, domain object, DTO, HTTP response, proxy response, TCP segment 순서로 돌아간다.

**꼬리 질문 지도**

- Nginx는 왜 필요한가: TLS 종료, static file, connection buffering, rate limit, upstream health check, L7 routing을 애플리케이션 앞에서 처리할 수 있다.
- Tomcat은 왜 필요한가: Java 웹 애플리케이션이 HTTP 요청을 서블릿 API와 스레드 모델 위에서 처리할 수 있게 하는 컨테이너다.
- keep-alive는 왜 중요한가: 요청마다 TCP와 TLS 연결을 새로 만들지 않아 handshake 비용과 `TIME_WAIT` 부담을 줄인다.
- 병목은 어디서 생기는가: listen backlog, file descriptor limit, Nginx worker, Tomcat thread pool, DB connection pool, slow query, lock wait, 외부 API timeout 중 하나가 먼저 포화된다.
- WebFlux나 Netty면 무엇이 달라지는가: thread-per-request 모델보다 적은 이벤트 루프 스레드로 많은 연결을 다루지만, CPU 작업이나 blocking call을 event loop에서 실행하면 오히려 전체 처리가 막힌다.

**실무 답변 포인트**

대용량 트래픽 질문에서는 "서버를 늘립니다"보다 먼저 어떤 자원이 포화되는지 말한다.
연결 수가 문제면 backlog, file descriptor, worker, keep-alive, idle timeout을 본다.
처리 시간이 문제면 애플리케이션 thread dump, DB slow query, p95/p99 latency, upstream timeout을 본다.
DB connection pool이 작아서 기다리는지, 너무 커서 DB를 압박하는지도 함께 본다.

**확인 경로**

- 관련 문서: [network-web-protocols.md](network-web-protocols.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [concurrency-async-io.md](concurrency-async-io.md)
- 확인 명령 후보: `ss -antp`, `lsof -i :80`, Nginx access/error log, Tomcat access log, Spring actuator metrics, DB slow query log

## 2. 트랜잭션은 ACID에서 MVCC, 격리 수준, 전파, 2PC, 보상 트랜잭션까지 어떻게 이어지는가

**첫 30초 답변**

트랜잭션은 여러 읽기와 쓰기를 하나의 논리적 작업으로 묶는 장치다.
ACID는 그 작업이 전부 성공하거나 실패해야 한다는 원자성, 무결성 제약을 깨지 않아야 한다는 일관성, 동시에 실행되어도 서로 망치지 않아야 한다는 격리성, 커밋 뒤 장애가 나도 남아야 한다는 지속성을 말한다.
MVCC는 특히 격리성을 구현하는 대표 방식으로, 읽는 쪽에는 일정 시점의 스냅샷을 보여 주고 쓰는 쪽은 새 버전을 만들게 해서 읽기와 쓰기 충돌을 줄인다.

**이어 말할 순서**

1. ACID는 트랜잭션의 목표이고, MVCC와 lock은 그 목표를 구현하는 수단이다.
2. 격리 수준은 동시 트랜잭션이 서로의 중간 상태를 얼마나 볼 수 있는지 정하는 정책이다.
3. `READ COMMITTED`는 보통 statement마다 커밋된 최신 상태를 보고, `REPEATABLE READ`는 한 트랜잭션 안에서 일관된 스냅샷을 본다.
4. phantom read는 같은 조건으로 다시 조회했을 때 새 row가 보이는 현상인데, SQL 표준 설명과 실제 DB 엔진 구현이 다를 수 있다.
5. PostgreSQL은 `REPEATABLE READ`를 스냅샷 격리에 가깝게 구현하므로 phantom까지 막지만, serialization anomaly는 `SERIALIZABLE`에서 다룬다.
6. Spring의 `@Transactional`은 DB 격리 수준 자체가 아니라, 애플리케이션 메서드 호출을 어떤 트랜잭션 경계에 넣을지 정한다.
7. 여러 DB나 여러 서비스가 하나의 business transaction에 참여하면 2PC, saga, outbox, idempotency 같은 분산 일관성 문제가 된다.

**꼬리 질문 지도**

- 2PC와 2PL은 무엇이 다른가: 2PC는 여러 참여자에게 commit/rollback 결정을 합의시키는 분산 원자성 프로토콜이고, 2PL은 lock 획득과 해제 단계를 제한해 직렬 가능성을 얻는 동시성 제어 프로토콜이다.
- PostgreSQL에서 2PC를 왜 쓰는가: 외부 트랜잭션 매니저가 여러 자원에 대한 commit 결정을 조율할 때, DB가 prepared transaction 상태를 durable하게 보관해야 하기 때문이다.
- 2PC의 위험은 무엇인가: coordinator 장애나 prepared transaction 방치가 lock과 자원을 오래 잡을 수 있고, 운영 복구 절차가 필요하다.
- 보상 트랜잭션은 rollback과 같은가: 아니다. 이미 커밋된 로컬 트랜잭션을 의미상 되돌리는 별도 비즈니스 행위다. 외부 결제 취소, 포인트 환불, 재고 복구처럼 정확한 과거 상태 복원이 아니라 반대 효과를 내는 작업일 수 있다.
- Spring 전파는 왜 중요한가: service 단위의 원자성을 잡아야 하는데 repository 메서드마다 트랜잭션을 나누면 주문 저장과 결제 저장이 따로 커밋될 수 있다.
- checked exception은 왜 주의해야 하는가: Spring 기본 정책에서는 runtime exception 중심으로 rollback 되므로, checked exception을 rollback하려면 정책을 명시해야 한다.

**실무 답변 포인트**

결제나 주문 질문에서는 "무조건 강한 트랜잭션"이라고 말하지 않는다.
한 DB 안의 짧은 작업은 DB 트랜잭션으로 묶는다.
외부 PG, 메시지 발행, 다른 서비스 DB까지 걸리는 긴 흐름은 2PC가 가능한지, 운영 부담이 감당되는지, 실패 보상과 중복 처리 방지가 가능한지로 나누어 답한다.
실무에서는 local transaction, outbox table, idempotency key, retry, dead-letter queue, compensation을 함께 설계하는 경우가 많다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md), [distributed-systems-architecture.md](distributed-systems-architecture.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 실험 후보: 두 DB 세션에서 같은 row와 range를 읽고 수정하며 격리 수준별 결과를 비교한다. Spring에서는 service 레벨 트랜잭션과 repository 레벨 트랜잭션의 commit 경계를 테스트로 나눈다.

## 3. 인덱스는 왜 B+Tree를 쓰고 실제 데이터 읽기와 어떻게 연결되는가

**첫 30초 답변**

DB 인덱스는 테이블 전체를 매번 읽지 않기 위해 만든 정렬된 보조 구조다.
B+Tree는 한 페이지 안에 많은 key를 담아 fanout을 크게 만들고, 모든 leaf를 같은 깊이에 둬서 최악의 탐색 비용을 작게 유지한다.
내부 노드는 길 안내용 key를 담고, leaf는 key와 row 위치 또는 실제 row를 담기 때문에, point lookup과 range scan을 모두 안정적으로 처리할 수 있다.

**이어 말할 순서**

1. DB는 보통 디스크와 메모리 page 단위로 읽기 때문에, 자료구조도 page I/O를 줄이는 방향이어야 한다.
2. B+Tree의 internal page는 key와 child pointer를 많이 담아 한 번 내려갈 때 후보 범위를 크게 줄인다.
3. 모든 leaf가 같은 깊이에 있으므로 특정 key만 유난히 오래 걸리는 편향을 줄인다.
4. leaf page들이 key 순서로 이어져 있어 range scan이 자연스럽다.
5. secondary index를 타면 index leaf에서 row id나 primary key를 얻고, 다시 table이나 clustered index로 실제 row를 찾아갈 수 있다.
6. covering index는 쿼리에 필요한 column이 모두 index 안에 있어 table row fetch를 생략한다.
7. 인덱스는 read를 빠르게 하지만 write, update, delete 때 tree 유지 비용과 page split 비용을 만든다.

**꼬리 질문 지도**

- 왜 balanced여야 하는가: tree가 한쪽으로 치우치면 최악의 탐색 비용이 커지고, 특정 key 범위가 계속 느려진다.
- 왜 깊이가 작아야 하는가: depth 하나가 늘 때마다 page read 후보가 늘어나며, 디스크 I/O나 buffer pool miss가 latency를 크게 키운다.
- 깊이가 왜 보통 몇 단계에 머무는가: page 하나에 key를 수백 개 담을 수 있으면 fanout이 커져 수억 건도 root, internal, leaf 몇 단계로 좁혀진다.
- B Tree와 B+Tree는 무엇이 다른가: B Tree는 내부 노드에도 data를 둘 수 있고, B+Tree는 실제 data 또는 row reference를 leaf에 모아 range scan과 높은 fanout에 유리하다.
- clustered index와 heap table은 무엇이 다른가: clustered 구조에서는 primary key 순서가 실제 row 저장 위치와 강하게 연결되고, heap 구조에서는 index가 별도 row 위치를 가리킨다.
- index scan이 항상 빠른가: 아니다. selectivity가 낮아 많은 row를 가져오면 random row fetch 비용 때문에 sequential scan이 더 나을 수 있다.
- composite index에서 순서가 왜 중요한가: `(a, b)` index는 `a`로 먼저 좁히는 쿼리에 강하고, `b`만 조건인 쿼리에는 기대한 만큼 도움이 안 될 수 있다.

**실무 답변 포인트**

인덱스 질문에서는 "B+Tree라서 빠릅니다"에서 멈추면 약하다.
page, fanout, buffer pool, selectivity, covering, row fetch, write amplification까지 이어 말해야 한다.
성능 질문이면 반드시 `EXPLAIN`을 언급한다.
좋은 답변은 "이 쿼리가 index를 탔는가"보다 "index를 탄 뒤 실제로 몇 row와 몇 page를 읽었는가"까지 본다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md)
- 확인 실험 후보: 같은 조건으로 `SELECT indexed_column`, `SELECT *`, 낮은 selectivity 조건, 높은 selectivity 조건을 나눠 `EXPLAIN`과 실제 읽은 row 수를 비교한다.

## 4. 수억 건 테이블은 인덱스만으로 해결되지 않을 때 어떻게 운영하는가

**첫 30초 답변**

수억 건 테이블 문제는 인덱스 하나를 추가하는 문제가 아니라 working set, write cost, migration risk, query routing을 함께 다루는 운영 문제다.
먼저 실제 조회 패턴과 보존 정책을 확인하고, 인덱스 정리, range scan 최적화, archive, partitioning, replica, batch migration을 조합한다.
서비스 중단이 불가능하면 한 번에 옮기지 않고 작은 batch와 검증 가능한 cutover 경로로 나눠야 한다.

**이어 말할 순서**

1. 현재 느린 이유를 먼저 분리한다. full scan인지, index selectivity 문제인지, lock wait인지, buffer pool miss인지, network 전송량인지 확인한다.
2. 자주 쓰는 쿼리와 오래된 데이터 접근 빈도를 나눠 hot data와 cold data를 분리한다.
3. 중복 인덱스와 거의 쓰지 않는 인덱스를 정리해 write 비용을 줄인다.
4. ID나 날짜처럼 라우팅 가능한 기준이 있으면 partitioning이나 table split 후보가 된다.
5. 기존 인터페이스를 보존하려면 view, application routing, compatibility API 중 하나를 선택한다.
6. 데이터 이동은 chunk 단위로 하고, 각 chunk마다 row count, checksum, 누락 범위, rollback 방법을 확인한다.
7. 최종적으로 `EXPLAIN`, slow query, replication lag, lock wait, batch duration으로 성공 여부를 본다.

**꼬리 질문 지도**

- archive와 partitioning은 무엇이 다른가: archive는 덜 쓰는 데이터를 별도 저장소나 테이블로 이동해 현재 working set을 줄이는 것이고, partitioning은 조회와 저장 경로를 기준에 따라 여러 물리 단위로 나누는 것이다.
- view로 통합하면 끝인가: 아니다. view 자체에는 보통 별도 index가 없고, optimizer가 각 underlying table의 index를 잘 쓰는지 확인해야 한다.
- application routing은 왜 위험한가: table 선택 규칙이 애플리케이션 계약이 되므로, id range나 date range 경계가 틀리면 누락과 중복이 생긴다.
- batch migration에서 무엇을 확인하는가: source count, target count, min/max id, checksum, failed batch retry, lock time, replication lag, 새 write 경로를 확인한다.
- 온라인 DDL을 못 쓰는 구버전이면 어떻게 하는가: shadow table, trigger나 dual write, chunk copy, read routing 전환 같은 보수적인 절차가 필요하지만, 구현 복잡도와 장애 시 복구 비용도 함께 설명해야 한다.

**실무 답변 포인트**

대용량 테이블 질문의 강한 답변은 "새 테이블로 나누겠습니다"가 아니다.
어떤 조회를 빠르게 만들고, 어떤 쓰기 비용을 감수하며, 어떤 기간 동안 어떤 fallback이 가능한지를 말해야 한다.
정산, 결제, 감사 데이터라면 삭제보다 보존과 재조회 가능성이 먼저다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md)
- 확인 실험 후보: 대표 쿼리별 `EXPLAIN`, index cardinality, slow query log, batch copy dry-run count, source/target checksum 비교

## 5. 동시성 문제는 race condition, visibility, ordering, deadlock으로 어떻게 나뉘는가

**첫 30초 답변**

동시성 문제는 단순히 두 스레드가 동시에 실행되는 문제가 아니라, 공유 상태를 여러 실행 흐름이 어떤 순서로 읽고 쓰는지의 문제다.
race condition은 실행 순서에 따라 결과가 달라지는 문제이고, visibility는 한 스레드의 변경을 다른 스레드가 언제 보느냐의 문제이며, ordering은 CPU나 컴파일러 최적화로 실행 순서가 관측상 달라지는 문제다.
lock, `volatile`, atomic operation, immutable object는 각각 다른 문제를 해결한다.

**이어 말할 순서**

1. 먼저 공유 mutable state가 있는지 확인한다.
2. `count++` 같은 read-modify-write는 원자적이지 않으므로 race condition이 생길 수 있다.
3. 한 스레드가 쓴 값을 다른 스레드가 못 보면 visibility 문제가 된다.
4. Java에서는 `synchronized`, `volatile`, `Lock`, atomic class가 happens-before 관계나 원자성을 제공한다.
5. `synchronized`는 mutual exclusion과 visibility를 함께 제공하지만, `volatile`은 단일 변수 가시성과 ordering에는 강해도 compound operation 원자성은 보장하지 않는다.
6. lock 순서가 뒤엉키면 deadlock이 생기고, lock을 너무 크게 잡으면 contention 때문에 throughput이 떨어진다.
7. immutable object와 thread confinement는 공유 변경 자체를 줄여 문제를 단순하게 만든다.

**꼬리 질문 지도**

- 읽기 전용 객체는 왜 안전한가: 객체가 완전히 초기화된 뒤 변경되지 않고 안전하게 publish 되면 여러 스레드가 같은 값을 읽어도 경쟁이 없다.
- `volatile int count`로 `count++`가 안전한가: 아니다. `count++`는 읽기, 증가, 쓰기 세 단계라 visibility만으로 원자성이 생기지 않는다.
- deadlock을 어떻게 줄이는가: lock ordering, timeout, lock scope 축소, try-lock, 더 높은 수준의 concurrent collection을 사용한다.
- synchronized와 ReentrantLock은 무엇이 다른가: 둘 다 mutual exclusion을 제공하지만, ReentrantLock은 tryLock, interruptible lock, fair lock 같은 제어 옵션을 제공한다.
- optimistic locking은 DB lock과 같은가: 보통 version이나 CAS 값을 비교해 충돌을 나중에 감지하는 방식이며, 자원을 미리 막는 pessimistic lock과 다르다.

**실무 답변 포인트**

동시성 답변에서는 용어를 한꺼번에 섞지 않는다.
"이 문제는 원자성 문제인지, 가시성 문제인지, 순서 문제인지, 진행 보장 문제인지 먼저 나눕니다"라고 말하면 답변이 단단해진다.
그 뒤에 코드 레벨 수단과 DB 레벨 수단이 같은 이름의 lock이어도 적용 계층과 비용이 다르다는 점을 설명한다.

**확인 경로**

- 관련 문서: [concurrency-async-io.md](concurrency-async-io.md), [language-runtime.md](language-runtime.md)
- 확인 실험 후보: 의도적으로 `count++`를 여러 스레드에서 반복해 누락을 보고, `AtomicInteger`, `synchronized`, `LongAdder`로 바꿔 결과와 성능을 비교한다.

## 6. blocking, non-blocking, async, event loop는 서로 어떻게 다른가

**첫 30초 답변**

blocking과 non-blocking은 호출한 스레드가 기다리는 방식의 차이고, sync와 async는 결과를 받는 방식의 차이다.
event loop는 준비된 I/O 이벤트를 적은 수의 스레드가 반복 처리하는 실행 모델이다.
Linux의 `epoll`은 많은 file descriptor 중 준비된 소켓만 알려 주는 readiness notification이고, 이것만으로 CPU 작업이나 모든 파일 I/O가 자동으로 빨라지는 것은 아니다.

**이어 말할 순서**

1. blocking I/O는 syscall이나 library call이 완료될 때까지 현재 스레드가 기다린다.
2. non-blocking I/O는 지금 준비되지 않았으면 `EAGAIN` 같은 신호로 즉시 돌아오게 한다.
3. busy loop로 계속 물어보면 CPU를 낭비하므로, `select`, `poll`, `epoll`, `kqueue`, IOCP 같은 감시 구조가 필요해진다.
4. `epoll_wait()`은 등록된 fd 중 ready 상태가 된 항목을 돌려준다.
5. 네트워크 소켓은 ready 상태 변화가 의미 있지만, 일반 파일은 보통 항상 읽기 가능한 것처럼 보이므로 event loop와 맞지 않는 경우가 많다.
6. Node.js의 libuv는 네트워크 소켓은 event loop로, 파일 I/O나 일부 DNS, crypto 작업은 thread pool로 처리할 수 있다.
7. Netty와 WebFlux는 event loop thread를 막지 않는 코드를 전제로 높은 동시 연결 수를 다룬다.

**꼬리 질문 지도**

- non-blocking이면 async인가: 아니다. non-blocking read를 직접 반복 호출하면 호출자는 여전히 결과 확인을 직접 관리한다. async는 완료 통지를 callback, future, coroutine continuation 같은 방식으로 받는다.
- event loop가 왜 적은 스레드로 많은 연결을 처리하는가: 대부분의 연결이 CPU를 쓰는 시간이 아니라 I/O 대기 시간이기 때문에, 준비된 소켓만 처리하면 된다.
- event loop가 언제 망가지는가: handler 안에서 blocking DB call, file read, sleep, CPU heavy 작업을 하면 같은 event loop에 묶인 다른 연결이 밀린다.
- coroutine은 thread인가: 아니다. coroutine은 중단과 재개 가능한 실행 단위이고, 실제 실행은 thread 위에서 일어난다.
- io_uring은 epoll과 무엇이 다른가: epoll은 준비 여부를 알려 주는 모델에 가깝고, io_uring은 제출 큐와 완료 큐로 실제 I/O 요청과 완료를 더 직접적으로 다루는 모델이다.

**실무 답변 포인트**

비동기 질문에서 "스레드를 적게 씁니다"만 말하면 부족하다.
대기 시간이 긴 I/O를 다룰 때 event loop가 유리하지만, blocking library를 섞으면 이 장점이 사라진다고 말해야 한다.
그래서 실무에서는 non-blocking stack을 쓰더라도 DB driver, HTTP client, file I/O, serialization, logging appender까지 함께 확인해야 한다.

**확인 경로**

- 관련 문서: [concurrency-async-io.md](concurrency-async-io.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 실험 후보: event loop handler에 `sleep`이나 blocking call을 넣고 p99 latency가 어떻게 튀는지 본다. Linux에서는 `strace`나 thread dump로 blocking syscall을 확인한다.

## 7. `java -jar` 이후 JVM, class loader, Spring Boot, AOP proxy는 어떤 순서로 동작하는가

**첫 30초 답변**

`java -jar`는 먼저 OS가 `java` 실행 파일을 프로세스로 만들고, JVM이 JAR의 manifest와 classpath를 해석해 main class를 찾은 뒤 실행한다.
Spring Boot에서는 `main()`이 `SpringApplication.run()`을 호출하고, application context가 bean definition을 읽어 bean을 만들고 의존성을 연결한다.
웹 애플리케이션이면 내장 Tomcat이 뜨고, AOP와 `@Transactional` 같은 기능은 실제 객체 앞에 proxy를 세워 메서드 호출을 감싼다.

**이어 말할 순서**

1. shell은 `execve` 계열 호출로 `java` 바이너리를 실행한다.
2. OS loader는 실행 파일과 shared library를 메모리에 매핑하고 프로세스 주소 공간을 준비한다.
3. JVM은 옵션, classpath, jar manifest, class loader 계층을 준비한다.
4. class loader는 bootstrap, platform, application, 그리고 Spring Boot의 launcher class loader 같은 계층을 통해 class를 찾는다.
5. `main()`이 호출되고 `SpringApplication.run()`이 environment, bean definition, auto-configuration, application context를 초기화한다.
6. 웹이면 embedded servlet container가 port를 열고 `DispatcherServlet`을 등록한다.
7. AOP proxy는 bean을 감싸 transaction, security, logging 같은 cross-cutting concern을 메서드 호출 앞뒤에 삽입한다.

**꼬리 질문 지도**

- 부모 위임 모델은 왜 필요한가: JDK 핵심 class를 애플리케이션이 임의로 덮어쓰지 못하게 하고, 공통 class 충돌을 줄인다.
- Spring Boot fat jar는 왜 별도 loader가 필요한가: nested jar 구조를 일반 classpath처럼 읽으려면 Boot loader가 내부 lib와 application class를 해석해야 한다.
- IoC와 DI는 어떻게 다르나: IoC는 객체 생성과 연결의 제어권이 컨테이너로 넘어간 구조이고, DI는 그 제어권을 이용해 필요한 의존성을 주입하는 방식이다.
- `@Transactional`은 왜 self-invocation에서 안 먹을 수 있나: 같은 객체 안의 `this.method()` 호출은 proxy를 거치지 않기 때문이다.
- AOP proxy가 JDK dynamic proxy인지 CGLIB인지 왜 중요한가: interface 기반인지 class subclassing 기반인지에 따라 final method, visibility, type casting 같은 제약이 달라진다.

**실무 답변 포인트**

Spring 질문은 annotation 이름을 외우는 질문처럼 보여도 실제로는 "누가 객체를 만들고, 누가 호출을 감싸고, 어떤 호출 경로에서 그 감싸기가 빠지는가"를 보는 경우가 많다.
따라서 `@Transactional` 답변은 항상 proxy, transaction manager, connection binding, propagation, rollback policy, self-invocation을 함께 묶어야 한다.

**확인 경로**

- 관련 문서: [language-runtime.md](language-runtime.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md)
- 확인 실험 후보: self-invocation 메서드와 외부 bean 호출 메서드를 나누고, transaction active 여부와 rollback 여부를 테스트한다.

## 8. Kafka, RabbitMQ, outbox, idempotent consumer는 왜 함께 물리는가

**첫 30초 답변**

메시징은 서비스 사이의 시간, 장애, 처리 속도를 분리하기 위한 구조다.
Kafka는 append-only log와 partition을 중심으로 여러 consumer group이 같은 데이터를 각자 offset으로 읽는 모델에 강하고, RabbitMQ는 exchange, queue, routing, ack를 중심으로 작업 전달과 라우팅에 강하다.
분산 시스템에서는 메시지를 보냈는데 DB commit이 실패하거나, DB commit은 됐는데 메시지 발행이 실패하는 문제가 생기므로 outbox와 idempotent consumer가 함께 나온다.

**이어 말할 순서**

1. producer는 이벤트나 command를 broker로 보낸다.
2. broker는 durable log나 queue에 메시지를 보관하고 consumer에게 전달한다.
3. consumer는 처리 뒤 ack나 offset commit으로 "어디까지 처리했는지"를 남긴다.
4. 장애가 나면 같은 메시지가 다시 전달될 수 있으므로 consumer는 idempotent해야 한다.
5. DB 변경과 메시지 발행을 원자적으로 묶기 어렵기 때문에 outbox table에 이벤트를 같은 DB transaction으로 저장하고, 별도 publisher가 broker로 내보내는 방식이 자주 쓰인다.
6. 실패 메시지는 retry, backoff, dead-letter queue, manual replay 경로로 관리한다.

**꼬리 질문 지도**

- Kafka는 왜 throughput이 높은가: partitioned append-only log, sequential I/O, batching, page cache 활용이 잘 맞기 때문이다.
- Kafka의 순서는 어디까지 보장되는가: partition 안에서는 순서가 있지만, topic 전체나 여러 partition 전체 순서는 별도 설계 없이는 보장되지 않는다.
- RabbitMQ는 왜 라우팅에 강한가: exchange type과 binding key로 direct, fanout, topic 같은 전달 패턴을 표현할 수 있다.
- exactly once는 믿어도 되는가: broker 내부 처리 보장과 외부 DB side effect의 exactly once는 다르다. 실무 답변에서는 idempotency key와 transaction boundary를 함께 말해야 한다.
- outbox의 단점은 무엇인가: outbox table 증가, publisher lag, 중복 발행, cleanup, ordering 처리를 운영해야 한다.

**실무 답변 포인트**

메시징 답변에서 가장 중요한 것은 "한 번만 처리됩니다"라고 쉽게 말하지 않는 것이다.
네트워크와 프로세스 장애가 있으면 at-least-once 재전달이 흔하고, 그래서 consumer 쪽 중복 처리 방지가 사실상 필수다.
결제나 주문에서는 event id, business idempotency key, unique constraint, 처리 상태 machine을 같이 둔다.

**확인 경로**

- 관련 문서: [messaging-event-driven.md](messaging-event-driven.md), [distributed-systems-architecture.md](distributed-systems-architecture.md)
- 확인 실험 후보: consumer가 처리 후 ack 전에 죽는 상황, ack 후 DB write 전에 죽는 상황을 나누어 재전달과 중복 처리를 확인한다.

## 9. 분산 시스템에서 strong consistency, eventual consistency, replication, quorum은 어떻게 엮이는가

**첫 30초 답변**

분산 시스템은 여러 노드가 같은 상태를 다루기 때문에, 장애와 네트워크 지연 속에서 어떤 읽기와 쓰기를 허용할지 정해야 한다.
strong consistency는 읽는 쪽이 최신 쓰기와 모순되지 않게 보는 성질을 강하게 요구하고, eventual consistency는 더 이상 업데이트가 없다면 복제본들이 결국 같은 값으로 수렴한다는 모델이다.
replication, quorum, leader election, consensus는 이 선택을 구현하는 수단이다.

**이어 말할 순서**

1. 단일 DB 안에서는 lock, MVCC, WAL 같은 메커니즘으로 일관성을 다룬다.
2. 여러 노드로 복제하면 leader, follower, replication log, lag가 생긴다.
3. follower read를 허용하면 latency와 read capacity는 좋아질 수 있지만 stale read가 생길 수 있다.
4. quorum read/write는 여러 replica 중 일정 수 이상의 응답을 요구해 최신성 확률이나 보장을 높인다.
5. network partition이 생기면 모든 노드가 항상 응답하는 가용성과 모든 응답이 최신이라는 일관성을 동시에 지키기 어렵다.
6. consensus는 여러 노드가 하나의 순서나 leader 결정에 합의하도록 도와주지만, latency와 운영 복잡도가 든다.
7. business flow에서는 완전한 strong consistency보다 invariant별로 강한 구간과 eventual 구간을 나누는 경우가 많다.

**꼬리 질문 지도**

- CAP를 어떻게 말해야 하나: 평소에는 latency와 consistency tradeoff가 더 많이 보이고, partition 상황에서는 consistency와 availability 중 무엇을 희생할지 명시해야 한다.
- replication lag가 실제로 어떤 문제를 만드나: 방금 쓴 주문이 조회 API에서 안 보이거나, 중복 요청 방지 조회가 stale data를 보고 잘못 통과할 수 있다.
- read-your-writes는 무엇인가: 자신이 방금 쓴 값은 이후 읽기에서 볼 수 있어야 한다는 사용자 관점 보장이다.
- leader election은 왜 필요한가: 같은 데이터를 여러 노드가 동시에 leader처럼 쓰면 split brain과 충돌이 생긴다.
- consensus를 모든 곳에 쓰면 되나: strong ordering이 필요한 metadata나 critical state에는 좋지만, 모든 business event에 적용하면 latency와 가용성 비용이 커질 수 있다.

**실무 답변 포인트**

분산 일관성 질문에서는 "정합성이 중요합니다"라는 말보다 어떤 invariant를 반드시 지켜야 하는지 말해야 한다.
예를 들어 결제 승인 중복 방지, 재고 음수 방지, 포인트 중복 적립 방지처럼 깨지면 돈이나 신뢰가 깨지는 invariant는 강하게 잡고, 알림이나 검색 색인 반영은 eventual로 둘 수 있다.

**확인 경로**

- 관련 문서: [distributed-systems-architecture.md](distributed-systems-architecture.md), [database-storage-search-nosql.md](database-storage-search-nosql.md), [messaging-event-driven.md](messaging-event-driven.md)
- 확인 실험 후보: primary write 후 replica read를 강제로 보내 lag를 관측하고, stale read가 business invariant를 깨는지 시나리오 테스트를 만든다.

## 10. TCP, HTTP, keep-alive, TLS, reverse proxy는 한 요청 경로에서 어떻게 만나는가

**첫 30초 답변**

TCP는 신뢰성 있는 byte stream을 제공하고, HTTP는 그 byte stream 위에서 request와 response 의미를 정의한다.
TLS는 HTTP 메시지가 오가기 전에 인증과 키 교환을 수행해 암호화된 채널을 만들며, HTTPS는 HTTP를 TLS 위에 올린 것이다.
reverse proxy는 클라이언트와 서버 사이에서 연결을 받아 backend로 다시 연결을 만들고, routing, TLS termination, buffering, compression, timeout 같은 정책을 적용한다.

**이어 말할 순서**

1. TCP는 SYN, SYN-ACK, ACK로 연결을 수립하고, sequence number와 ack로 순서와 재전송을 관리한다.
2. TLS를 쓰면 HTTP 요청 전 `ClientHello`, `ServerHello`, certificate, key exchange, finished 흐름으로 session key를 합의한다.
3. HTTP/1.1 keep-alive는 같은 TCP 연결에서 여러 요청을 재사용해 handshake 비용을 줄인다.
4. HTTP/2는 하나의 연결 안에서 stream multiplexing을 제공하지만, TCP head-of-line blocking 영향은 남을 수 있다.
5. reverse proxy는 client-side connection과 upstream connection을 분리한다.
6. timeout은 client, proxy, app server, DB, external API마다 따로 있고, 서로 맞지 않으면 중간 계층에서 끊긴다.

**꼬리 질문 지도**

- keep-alive를 너무 길게 두면 무엇이 문제인가: idle connection이 file descriptor와 memory를 잡고, load balancer나 proxy의 connection table을 압박할 수 있다.
- TCP와 HTTP의 timeout은 어떻게 다른가: TCP 연결 자체의 상태와 HTTP request 처리 시간은 다른 계층의 정책이다.
- TLS termination은 어디서 하는가: Nginx나 load balancer에서 끝낼 수도 있고, 애플리케이션까지 end-to-end로 가져갈 수도 있다. 보안 경계와 운영 편의가 tradeoff다.
- proxy와 reverse proxy는 무엇이 다른가: forward proxy는 client를 대신해 외부로 나가고, reverse proxy는 server 앞에서 client 요청을 받아 backend로 전달한다.
- streaming 응답은 무엇이 어려운가: 전체 파일을 memory에 올리지 않아야 하고, backpressure, client disconnect, proxy buffering, timeout을 함께 봐야 한다.

**실무 답변 포인트**

네트워크 답변에서는 "HTTP는 TCP 위에서 동작합니다"에서 멈추지 않는다.
실제 장애는 TIME_WAIT 증가, TLS handshake 비용, keep-alive timeout mismatch, proxy buffering, slow client, upstream timeout처럼 계층 사이 설정 불일치에서 자주 나온다.

**확인 경로**

- 관련 문서: [network-web-protocols.md](network-web-protocols.md), [security-cryptography.md](security-cryptography.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 명령 후보: `curl -v`, `openssl s_client`, `tcpdump`, `ss -s`, Nginx upstream timing log

## 11. TLS와 암호학 질문은 인증, 키 교환, 암호화, 해시를 어떻게 분리해야 하는가

**첫 30초 답변**

TLS는 단순히 "공개키로 암호화한다"가 아니다.
인증서는 서버가 정말 그 도메인의 주체인지 확인하는 데 쓰이고, ECDHE 같은 키 교환은 클라이언트와 서버가 네트워크에 비밀 키를 직접 보내지 않고 shared secret을 만들기 위해 쓰인다.
실제 데이터 암호화는 핸드셰이크 뒤 만들어진 대칭키로 처리하고, 무결성은 AEAD나 MAC 계열 메커니즘이 담당한다.

**이어 말할 순서**

1. 클라이언트는 `ClientHello`로 지원 TLS 버전과 cipher suite 후보를 보낸다.
2. 서버는 인증서와 key exchange 정보를 보내 자신의 신원을 증명하고 session key 재료를 제공한다.
3. 클라이언트는 인증서 chain, hostname, validity, trust anchor를 확인한다.
4. ECDHE에서는 양쪽이 임시 개인키와 공개값으로 같은 shared secret을 계산한다.
5. 양쪽은 shared secret과 hello random들을 사용해 traffic key를 만든다.
6. 이후 HTTP 데이터는 대칭키 기반으로 빠르게 암호화되고 무결성 검증된다.

**꼬리 질문 지도**

- 전방 비밀성은 무엇인가: 서버 장기 private key가 나중에 유출되어도 과거 세션의 트래픽을 해독하기 어렵게 하는 성질이다. 세션마다 임시 key exchange를 쓰기 때문에 가능하다.
- 암호화와 해시는 무엇이 다른가: 암호화는 key로 원문과 암호문을 변환하고 복호화가 가능하다. 해시는 단방향 요약값이다.
- 서명은 무엇을 보장하나: private key 소유자가 특정 데이터에 서명했음을 검증하고, 데이터 변조를 탐지하게 해 준다.
- password는 암호화하면 되나: 보통 복호화가 필요 없으므로 salt와 느린 password hashing을 사용한다.
- 인증서가 있어도 중간자 공격이 가능한가: trust store, hostname verification, private key 보호, 잘못된 인증서 수락 정책이 깨지면 가능하다.

**실무 답변 포인트**

TLS 질문에서 가장 흔한 약한 답변은 "공개키로 데이터를 암호화합니다"다.
현대 TLS에서는 인증, 키 교환, traffic key derivation, 대칭 암호화가 역할을 나눠 가진다고 설명해야 한다.
그리고 인증서 검증을 꺼 버리는 `trust all` 코드가 왜 위험한지 바로 이어 말할 수 있어야 한다.

**확인 경로**

- 관련 문서: [security-cryptography.md](security-cryptography.md), [network-web-protocols.md](network-web-protocols.md)
- 확인 명령 후보: `openssl s_client -connect host:443 -servername host`, `curl -v https://host`

## 12. JVM GC와 런타임 성능은 heap, allocation, pause, throughput을 어떻게 이어 말하는가

**첫 30초 답변**

GC는 더 이상 접근할 수 없는 객체를 찾아 heap memory를 회수하는 런타임 기능이다.
JVM은 객체 allocation이 많고 heap이 커질수록 GC 비용과 pause가 성능에 영향을 줄 수 있으므로, young generation, old generation, region, root scanning, marking, compaction 같은 메커니즘을 사용한다.
튜닝의 목표는 "GC를 없애는 것"이 아니라 latency, throughput, memory footprint 중 서비스 목표에 맞는 균형을 잡는 것이다.

**이어 말할 순서**

1. 객체는 대부분 heap에 할당되고, thread마다 stack frame에는 지역 변수와 호출 상태가 쌓인다.
2. 많은 객체는 금방 죽으므로 generational GC는 young 영역을 자주 빠르게 청소한다.
3. 오래 살아남은 객체는 old 영역이나 region으로 이동하고, 여기서 회수 비용이 더 커질 수 있다.
4. GC는 root에서 시작해 reachable object graph를 찾고, unreachable object를 회수한다.
5. Stop-the-world는 GC가 일관된 object graph를 다루기 위해 application thread를 잠시 멈추는 구간이다.
6. G1, ZGC 같은 collector는 pause time, heap size, throughput 목표가 다르다.
7. GC 문제는 log, allocation rate, promotion rate, pause time, heap occupancy로 확인한다.

**꼬리 질문 지도**

- VM과 runtime과 GC는 같은가: 아니다. VM은 bytecode 같은 중간 표현을 실행하는 환경이고, runtime은 memory, scheduling, library support 같은 실행 지원 전체를 말하며, GC는 그중 memory 회수 기능이다.
- heap을 크게 하면 해결되나: OOM 가능성은 줄 수 있지만, pause와 memory footprint가 늘 수 있다. 반대로 너무 작으면 GC가 너무 자주 돈다.
- memory leak은 GC 언어에서 가능한가: 가능하다. 더 이상 business 의미가 없는 객체라도 static map, cache, listener, ThreadLocal 등에서 reachable하면 GC가 회수할 수 없다.
- `finalize`나 명시적 GC 호출은 좋은가: 보통 예측 가능성과 성능을 해치므로 일반 해법으로 삼지 않는다.
- GC 튜닝보다 먼저 볼 것은 무엇인가: 불필요한 allocation, 큰 object graph, cache 정책, batch size, serialization buffer, query result size를 먼저 본다.

**실무 답변 포인트**

GC 답변은 collector 이름 나열보다 "서비스 목표"로 시작한다.
p99 latency가 중요하면 pause를 보고, batch throughput이 중요하면 전체 처리량을 보며, memory 비용이 중요하면 heap occupancy와 object lifetime을 본다.
GC log 없이 collector만 바꾸는 것은 진단이 아니라 추측에 가깝다.

**확인 경로**

- 관련 문서: [language-runtime.md](language-runtime.md), [concurrency-async-io.md](concurrency-async-io.md)
- 확인 명령 후보: GC log, Java Flight Recorder, heap dump, allocation profiler, thread dump

## 13. 장애 분석 질문은 관측 지표에서 병목 계층으로 어떻게 좁혀 가는가

**첫 30초 답변**

장애 분석은 증상에서 바로 원인을 찍는 일이 아니라, 어느 계층의 병목인지 좁혀 가는 일이다.
먼저 사용자 영향과 시간 범위를 고정하고, p50/p95/p99 latency, error rate, traffic, saturation을 본다.
그 다음 CPU, memory, GC, thread, connection pool, DB, network, broker lag, external dependency를 순서대로 확인해 병목 계층을 찾는다.

**이어 말할 순서**

1. 증상과 영향 범위를 먼저 고정한다. 몇 시부터, 어떤 endpoint, 어떤 사용자, 어떤 error code인지 본다.
2. latency와 error가 같이 올랐는지, latency만 올랐는지, traffic 증가와 함께인지 확인한다.
3. 애플리케이션에서는 thread pool queue, active thread, blocked thread, GC pause, exception rate를 본다.
4. DB에서는 connection pool wait, lock wait, slow query, execution plan, replication lag를 본다.
5. 네트워크에서는 connect timeout, read timeout, retransmission, DNS, TLS handshake, upstream timing을 본다.
6. 메시징에서는 consumer lag, retry storm, DLQ 증가, partition skew를 본다.
7. 원인을 찾은 뒤에는 재발 방지 지표와 alert threshold를 남긴다.

**꼬리 질문 지도**

- CPU가 낮은데 느린 이유는 무엇인가: I/O wait, lock wait, connection pool wait, external API wait처럼 CPU를 쓰지 않고 기다리는 병목일 수 있다.
- p99만 튀는 이유는 무엇인가: 평균은 정상이어도 일부 요청이 lock, GC, cold cache, slow query, retry, queueing delay를 만날 수 있다.
- thread dump에서 무엇을 보나: 많은 thread가 같은 lock, socket read, DB driver, pool borrow, file I/O에서 멈춰 있는지 본다.
- slow query가 없는데 DB가 느릴 수 있나: connection pool wait, lock wait, network, result set 과다 전송, buffer pool miss, replica lag가 원인일 수 있다.
- 장애 대응과 근본 해결은 어떻게 나누나: 우선 우회나 rate limit으로 영향도를 줄이고, 이후 원인 메커니즘과 재발 방지 장치를 닫는다.

**실무 답변 포인트**

면접에서 장애 경험을 말할 때는 "로그를 봤습니다"보다 "어떤 가설을 세우고 어떤 관측으로 배제했는지"가 중요하다.
좋은 답변은 직접 원인과 근본 원인을 나눈다.
예를 들어 "DB connection pool exhausted"는 직접 원인이고, 그 앞의 slow query, 외부 API 지연, transaction scope 과다, thread leak이 근본 원인일 수 있다.

**확인 경로**

- 관련 문서: [problem-solving-code-quality.md](problem-solving-code-quality.md), [network-web-protocols.md](network-web-protocols.md), [database-storage-search-nosql.md](database-storage-search-nosql.md), [messaging-event-driven.md](messaging-event-driven.md)
- 확인 명령 후보: dashboard, log correlation id, thread dump, heap/GC log, DB slow query log, `EXPLAIN`, broker lag metrics

## 14. 알고리즘과 코드 품질 질문은 복잡도, 자료구조, 불변식, 테스트 가능성으로 어떻게 답하는가

**첫 30초 답변**

알고리즘 질문은 정답 코드를 외우는 문제가 아니라, 입력 크기와 제약에서 어떤 자료구조와 불변식을 선택해야 하는지 설명하는 문제다.
먼저 완전 탐색이 왜 비싼지 말하고, 중복 계산, 정렬된 성질, 그래프 구조, greedy choice, dynamic programming state 같은 개선 축을 찾는다.
코드 품질 질문에서는 그 선택이 테스트 가능하고 변경에 강하며, 실패했을 때 어디서 깨졌는지 드러나는지도 함께 본다.

**이어 말할 순서**

1. 입력 크기와 시간/메모리 제약을 먼저 확인한다.
2. 가장 단순한 baseline을 말하고, 그 복잡도가 왜 부족한지 계산한다.
3. 중복 계산이면 memoization이나 DP를, 빠른 lookup이면 hash map이나 set을, 순서가 중요하면 heap, tree, sort, two pointer를 생각한다.
4. greedy는 "지금 최선의 선택이 전체 최선으로 이어지는 불변식"을 설명해야 한다.
5. DP는 state, transition, base case, iteration order를 설명해야 한다.
6. graph는 node, edge, weight, direction, visited invariant를 먼저 고정한다.
7. 구현 뒤에는 boundary case와 failure case로 검증한다.

**꼬리 질문 지도**

- 시간 복잡도만 보면 충분한가: 아니다. memory, constant factor, cache locality, I/O, DB round trip 같은 실제 비용도 봐야 한다.
- hash map은 항상 O(1)인가: 평균적으로 빠르지만 hash collision, resizing, memory overhead가 있고, 순서 조건에는 맞지 않는다.
- DP와 greedy를 어떻게 구분하나: greedy는 한 번의 선택을 되돌리지 않아도 되는 증명이 필요하고, DP는 여러 부분 문제의 최적 결과를 조합한다.
- 테스트 더블은 왜 필요한가: 외부 시스템이나 느린 의존성을 대체해 단위 테스트가 빠르고 결정적으로 실패하게 만든다.
- 좋은 코드 설계는 무엇인가: 책임 경계가 분명하고, 도메인 불변식이 코드 구조에 드러나며, 테스트가 business rule을 설명해야 한다.

**실무 답변 포인트**

알고리즘과 코드 품질은 별개가 아니다.
실무에서는 빠른 알고리즘을 고르는 것만큼, 그 알고리즘의 전제와 실패 조건을 코드와 테스트에 남기는 것이 중요하다.
면접에서는 "왜 이 자료구조가 이 입력 제약에서 맞는지"와 "나중에 요구가 바뀌면 어디를 바꾸면 되는지"를 함께 답하면 강하다.

**확인 경로**

- 관련 문서: [problem-solving-code-quality.md](problem-solving-code-quality.md), [dynamic_programming.md](../algorithms/dynamic_programming.md)
- 확인 실험 후보: baseline과 개선안을 같은 입력에서 실행해 시간 증가율을 비교하고, boundary case를 먼저 테스트한다.

## 복합 질문으로 연습하기

이 절은 실제 면접처럼 한 질문에서 여러 주제가 한꺼번에 따라오는 연습용이다.
각 질문은 외울 답이 아니라 말하는 순서를 익히기 위한 것이다.

### 결제 승인 API를 설계해 보세요

**첫 답변**

결제 승인 API는 중복 승인 방지, 외부 PG 장애 처리, DB 원자성, 메시지 발행, 운영 추적성을 함께 만족해야 한다.
요청이 들어오면 idempotency key로 중복을 먼저 막고, local DB transaction 안에서 결제 시도 상태와 outbox event를 함께 기록한다.
외부 PG 호출은 timeout과 retry 정책을 명확히 두고, 성공이나 실패 결과를 상태 machine으로 반영하며, 후속 알림이나 정산 이벤트는 outbox publisher가 broker로 내보낸다.

**이어갈 꼬리**

1. HTTP 요청 경로: Nginx, Tomcat, Spring controller, service, HikariCP
2. 트랜잭션 경계: 결제 시도 row, idempotency key unique constraint, outbox write
3. 외부 PG: timeout, retry, circuit breaker, duplicate request 방지
4. 분산 일관성: 2PC를 쓰기 어려운 이유, saga와 compensation
5. 메시징: outbox, broker, consumer idempotency, DLQ
6. 인덱스: `payment_id`, `idempotency_key`, `status`, `created_at`
7. 운영 관측: correlation id, PG latency, approval failure rate, retry count, outbox lag

### 대용량 조회 API가 갑자기 느려졌습니다

**첫 답변**

먼저 어느 계층에서 느린지 좁힌다.
API p99가 튀는지, DB slow query가 늘었는지, connection pool wait가 생겼는지, GC나 thread block이 있는지 확인한다.
DB 쪽이면 실행 계획이 index scan에서 full scan으로 바뀌었는지, selectivity가 나빠졌는지, covering index가 깨졌는지, lock wait나 buffer pool miss가 있는지 본다.

**이어갈 꼬리**

1. 관측: p99, error, QPS, pool wait, slow query
2. DB: `EXPLAIN`, rows examined, index cardinality, filesort, temporary table
3. 인덱스: composite order, covering, range condition, row fetch
4. 테이블 운영: archive, partitioning, hot/cold data
5. 애플리케이션: pagination, N+1 query, serialization cost, response size
6. 네트워크: large payload, proxy buffering, client disconnect

### WebFlux로 바꾸면 서버가 더 빨라지나요

**첫 답변**

항상 빨라지는 것은 아니다.
WebFlux나 Netty 기반 non-blocking stack은 많은 I/O 대기 연결을 적은 event loop thread로 다룰 때 유리하다.
하지만 DB driver, file I/O, 외부 library, JSON 처리, 암호화 같은 blocking 또는 CPU-heavy 작업을 event loop에서 실행하면 오히려 전체 연결이 같이 밀릴 수 있다.

**이어갈 꼬리**

1. 개념 구분: blocking vs non-blocking, sync vs async
2. OS: socket fd, `epoll`, readiness event
3. 런타임: event loop, worker pool, backpressure
4. Spring: MVC thread-per-request와 WebFlux event loop 모델
5. DB: R2DBC 같은 non-blocking driver가 필요한 이유
6. 운영: thread dump 대신 event loop latency, queue depth, blocking call detection을 본다

### 로그인과 HTTPS 보안을 설명해 보세요

**첫 답변**

HTTPS는 TLS로 통신 채널을 보호하고, 로그인은 사용자의 인증 정보를 검증해 application session이나 token을 발급하는 별도 문제다.
TLS에서는 인증서로 서버 신원을 확인하고, ECDHE 같은 키 교환으로 session key를 만든 뒤, 실제 HTTP 데이터는 대칭키로 암호화한다.
비밀번호는 복호화할 필요가 없으므로 salt와 느린 password hashing으로 저장하고, session cookie나 token은 탈취와 재사용을 막기 위해 만료, secure flag, same-site, rotation 같은 정책을 둔다.

**이어갈 꼬리**

1. TCP/TLS: handshake, certificate, ECDHE, traffic key
2. 암호학: encryption, hash, signature, MAC 구분
3. 웹 보안: cookie, SameSite, CSRF, XSS, CORS
4. 운영: certificate renewal, private key 보호, TLS termination 위치
5. 장애: handshake failure, expired certificate, wrong SNI, trust store 문제

## 마지막 점검 질문

문서를 덮고 아래 질문에 자기 말로 답해 본다.
막히는 질문이 있으면 해당 대주제 문서로 돌아가 세부를 다시 읽는다.

1. `@Transactional` 질문에서 DB 격리 수준, Spring proxy, propagation, rollback policy, distributed transaction을 한 흐름으로 설명할 수 있는가.
2. B+Tree 질문에서 page, fanout, depth, leaf link, row fetch, covering index, write cost를 빠뜨리지 않고 말할 수 있는가.
3. event loop 질문에서 `epoll`이 "완료"가 아니라 "준비됨"을 알려 주는 모델이라는 점을 설명할 수 있는가.
4. 메시징 질문에서 broker의 전달 보장과 외부 DB side effect의 정확히 한 번 처리가 다른 문제임을 말할 수 있는가.
5. 장애 질문에서 직접 원인과 근본 원인을 분리하고, 어떤 관측으로 가설을 배제했는지 설명할 수 있는가.
6. 보안 질문에서 인증서, 키 교환, 세션 키, 대칭 암호, 해시, 서명을 서로 섞지 않고 설명할 수 있는가.
7. 시스템 설계 질문에서 작은 기술 단위의 제약을 큰 아키텍처 선택으로 연결할 수 있는가.
