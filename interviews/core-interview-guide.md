# 핵심 인터뷰 정리

이 문서는 `interviews` 디렉터리의 세부 대주제 문서를 대체하지 않습니다.
역할은 더 좁습니다.
면접에서 짧은 질문 하나가 들어왔을 때, 여러 기술 단위를 한 번에 엮어 답변할 수 있도록 핵심 질문 묶음과 꼬리 질문 경로를 빠르게 복원하는 데 있습니다.

세부 문서는 각각 한 주제를 깊게 파고듭니다.
이 문서는 그 세부 주제를 실전 답변 순서로 다시 묶습니다.
좋은 답변은 처음 30초 안에 핵심 판단을 말하고, 그 뒤에 운영체제, 런타임, 프레임워크, 네트워크, 데이터베이스, 분산 시스템, 운영 관측으로 필요한 만큼 내려갑니다.

## 목차

- [핵심 인터뷰 정리](#핵심-인터뷰-정리)
  - [목차](#목차)
  - [빠른 사용법](#빠른-사용법)
  - [시간이 없을 때 먼저 볼 질문](#시간이-없을-때-먼저-볼-질문)
  - [모든 답변에 깔리는 기본 프레임](#모든-답변에-깔리는-기본-프레임)
  - [이 문서에 주제를 추가하는 기준](#이-문서에-주제를-추가하는-기준)
  - [1. 웹 요청 하나는 OS, 프록시, 톰캣, 스프링, DB를 어떻게 지나가나요](#1-웹-요청-하나는-os-프록시-톰캣-스프링-db를-어떻게-지나가나요)
  - [2. 트랜잭션은 ACID에서 MVCC, 격리 수준, 전파, 2PC, 보상 트랜잭션까지 어떻게 이어지나요](#2-트랜잭션은-acid에서-mvcc-격리-수준-전파-2pc-보상-트랜잭션까지-어떻게-이어지나요)
  - [3. 인덱스는 왜 B+Tree를 쓰고 실제 데이터 읽기와 어떻게 연결되나요](#3-인덱스는-왜-btree를-쓰고-실제-데이터-읽기와-어떻게-연결되나요)
  - [4. 수억 건 테이블은 인덱스만으로 해결되지 않을 때 어떻게 운영하나요](#4-수억-건-테이블은-인덱스만으로-해결되지-않을-때-어떻게-운영하나요)
  - [5. 동시성 문제는 race condition, visibility, ordering, deadlock으로 어떻게 나뉘나요](#5-동시성-문제는-race-condition-visibility-ordering-deadlock으로-어떻게-나뉘나요)
  - [6. blocking, non-blocking, async, event loop는 서로 어떻게 다른가요](#6-blocking-non-blocking-async-event-loop는-서로-어떻게-다른가요)
  - [7. `java -jar` 이후 JVM, class loader, Spring Boot, AOP proxy는 어떤 순서로 동작하나요](#7-java--jar-이후-jvm-class-loader-spring-boot-aop-proxy는-어떤-순서로-동작하나요)
  - [8. Kafka, RabbitMQ, outbox, idempotent consumer는 왜 함께 물리나요](#8-kafka-rabbitmq-outbox-idempotent-consumer는-왜-함께-물리나요)
  - [9. 분산 시스템에서 strong consistency, eventual consistency, replication, quorum은 어떻게 엮이나요](#9-분산-시스템에서-strong-consistency-eventual-consistency-replication-quorum은-어떻게-엮이나요)
  - [10. TCP, HTTP, keep-alive, TLS, reverse proxy는 한 요청 경로에서 어떻게 만나나요](#10-tcp-http-keep-alive-tls-reverse-proxy는-한-요청-경로에서-어떻게-만나나요)
  - [11. TLS와 암호학 질문은 인증, 키 교환, 암호화, 해시를 어떻게 분리해야 하나요](#11-tls와-암호학-질문은-인증-키-교환-암호화-해시를-어떻게-분리해야-하나요)
  - [12. JVM GC와 런타임 성능은 heap, allocation, pause, throughput을 어떻게 이어 말하나요](#12-jvm-gc와-런타임-성능은-heap-allocation-pause-throughput을-어떻게-이어-말하나요)
  - [13. 장애 분석 질문은 관측 지표에서 병목 계층으로 어떻게 좁혀 가나요](#13-장애-분석-질문은-관측-지표에서-병목-계층으로-어떻게-좁혀-가나요)
  - [14. 알고리즘과 코드 품질 질문은 복잡도, 자료구조, 불변식, 테스트 가능성으로 어떻게 답하나요](#14-알고리즘과-코드-품질-질문은-복잡도-자료구조-불변식-테스트-가능성으로-어떻게-답하나요)
  - [15. 검색과 NoSQL은 shard, replica, query fan-out, heap을 어떻게 함께 다루나요](#15-검색과-nosql은-shard-replica-query-fan-out-heap을-어떻게-함께-다루나요)
  - [16. 조회 API 성능은 애플리케이션과 DB 경계를 어떻게 나누어 보나요](#16-조회-api-성능은-애플리케이션과-db-경계를-어떻게-나누어-보나요)
  - [17. 파일 스트리밍은 HTTP, TCP, page cache, backpressure를 어떻게 지나가나요](#17-파일-스트리밍은-http-tcp-page-cache-backpressure를-어떻게-지나가나요)
  - [18. 로그인, 세션, 토큰, OAuth는 인증과 인가를 어떻게 분리해 말하나요](#18-로그인-세션-토큰-oauth는-인증과-인가를-어떻게-분리해-말하나요)
  - [19. 컨테이너에서 Spring 서비스가 느리거나 죽을 때 무엇을 확인하나요](#19-컨테이너에서-spring-서비스가-느리거나-죽을-때-무엇을-확인하나요)
  - [20. 고가용성은 replication, 장애 격리, health check, failover를 어떻게 설계하나요](#20-고가용성은-replication-장애-격리-health-check-failover를-어떻게-설계하나요)
  - [21. Spring의 IoC, DI, AOP proxy, test double은 객체 경계와 테스트 가능성을 어떻게 만드나요](#21-spring의-ioc-di-aop-proxy-test-double은-객체-경계와-테스트-가능성을-어떻게-만드나요)
  - [22. 숫자 표현, CPU cache, memory barrier는 정확성과 성능 질문에서 어떻게 등장하나요](#22-숫자-표현-cpu-cache-memory-barrier는-정확성과-성능-질문에서-어떻게-등장하나요)
  - [23. VM, runtime, GC, scheduler는 언어 실행 모델 비교에서 어떻게 나뉘나요](#23-vm-runtime-gc-scheduler는-언어-실행-모델-비교에서-어떻게-나뉘나요)
  - [복합 질문으로 연습하기](#복합-질문으로-연습하기)
    - [결제 승인 API를 설계해 보세요](#결제-승인-api를-설계해-보세요)
    - [대용량 조회 API가 갑자기 느려졌습니다](#대용량-조회-api가-갑자기-느려졌습니다)
    - [WebFlux로 바꾸면 서버가 더 빨라지나요](#webflux로-바꾸면-서버가-더-빨라지나요)
    - [로그인과 HTTPS 보안을 설명해 보세요](#로그인과-https-보안을-설명해-보세요)
    - [검색 API를 설계해 보세요](#검색-api를-설계해-보세요)
    - [100GB 파일을 다운로드하게 해야 합니다](#100gb-파일을-다운로드하게-해야-합니다)
    - [컨테이너 배포 후 서비스가 OOMKilled로 재시작됩니다](#컨테이너-배포-후-서비스가-oomkilled로-재시작됩니다)
    - [OAuth 기반 외부 API 연동을 설계해 보세요](#oauth-기반-외부-api-연동을-설계해-보세요)
  - [낯선 복합 질문을 받았을 때 복구 루틴](#낯선-복합-질문을-받았을-때-복구-루틴)
  - [마지막 점검 질문](#마지막-점검-질문)

## 빠른 사용법

이 문서는 암기 문장 모음이 아니라 답변 경로 지도입니다.
한 질문을 받으면 아래 순서로 말합니다.

1. 먼저 해결하려는 문제를 말합니다.

    예를 들어 트랜잭션 질문이면 "여러 변경을 하나의 논리적 작업으로 묶고, 동시에 실행되는 작업이 서로 망치지 않게 하며, 장애 뒤에도 복구 가능한 상태를 남기는 장치입니다"처럼 답합니다.
    용어 정의보다 문제를 먼저 잡으면 꼬리 질문이 들어와도 답변 방향을 잃지 않습니다.

2. 핵심 불변식을 말합니다.

    불변식은 깨지면 안 되는 약속입니다.
    인덱스라면 "정렬된 작은 구조를 통해 전체 테이블을 매번 읽지 않는다"가 불변식이고, 메시징이라면 "생산자와 소비자의 시간, 장애, 처리 속도를 분리한다"가 불변식입니다.

3. 실제 경로를 낮은 층부터 위로 올립니다.

    면접관이 깊게 들어오면 "애플리케이션이 이렇게 합니다"에서 끝내지 말고, 소켓, 파일 디스크립터, 커널 대기열, 런타임 스레드, 커넥션 풀, WAL, 인덱스 페이지처럼 실제로 상태가 바뀌는 지점을 짚습니다.

4. 트레이드오프와 실패 신호로 마무리합니다.

    좋은 답변은 "항상 좋다"가 아니라 "이 조건에서는 이기고, 이 조건에서는 비용이 생깁니다"까지 말합니다.
    마지막에는 `EXPLAIN`, GC log, thread dump, slow query log, packet capture, broker lag, p99 latency처럼 확인 경로를 붙입니다.

## 시간이 없을 때 먼저 볼 질문

시간이 정말 촉박하면 아래 8개를 먼저 복습합니다.
이 8개는 서로 연결되어 있어서, 하나를 제대로 답하면 여러 꼬리 질문으로 확장할 수 있습니다.

1. 웹 요청 하나가 들어와 DB에 저장되고 응답되기까지 어떤 계층을 지나가나요.
2. 트랜잭션 격리 수준은 MVCC, lock, phantom read, Spring 전파, 2PC, 보상 트랜잭션과 어떻게 이어지나요.
3. 인덱스는 왜 B+Tree이고, 커버링 인덱스와 실제 row fetch는 어떻게 다른가요.
4. blocking, non-blocking, async, event loop, epoll은 각각 무엇을 해결하고 무엇을 해결하지 못하나요.
5. `@Transactional`은 왜 프록시를 거쳐야 하고, self-invocation에서는 왜 깨질 수 있나요.
6. Kafka와 RabbitMQ는 로그와 큐라는 모델 차이 때문에 순서, 재처리, 확장 방식이 어떻게 달라지나요.
7. TLS 핸드셰이크에서 인증서와 ECDHE가 각각 무엇을 해결하나요.
8. 장애가 났을 때 p99 latency, CPU, memory, GC, thread, DB, network, broker lag 중 어디서부터 좁혀 가나요.

이 8개를 먼저 닫은 뒤에는 아래 5개를 다음 우선순위로 봅니다.
이들은 단독 암기 주제라기보다, 실무 경험 질문이나 시스템 설계 질문에서 여러 문서 주제가 한꺼번에 섞여 나오는 경로입니다.

1. 검색과 NoSQL 시스템은 shard, replica, query fan-out, heap, GC, 장애 승격을 어떻게 같이 설명하나요.
2. 조회 API가 느릴 때 애플리케이션 코드, mapper/ORM, DB 실행 계획, payload 크기를 어떻게 나누어 보나요.
3. 100GB 파일 다운로드나 스트리밍은 page cache, TCP window, proxy buffering, backpressure를 어떻게 지나가나요.
4. 로그인과 OAuth 질문에서 인증, 인가, session, token, `state`, PKCE를 어떻게 분리해 말하나요.
5. 컨테이너에서 서비스가 죽거나 느릴 때 namespace, cgroup, PID 1, OOM, health check를 어떻게 확인하나요.

## 모든 답변에 깔리는 기본 프레임

면접 답변은 보통 아래 다섯 문장 뼈대로 정리하면 덜 흔들립니다.

```text
이 질문은 결국 [문제]를 해결하는 방식에 관한 질문입니다.
핵심 약속은 [불변식]입니다.
실제로는 [낮은 계층]에서 [중간 계층]을 거쳐 [애플리케이션 계층]으로 올라옵니다.
대신 [트레이드오프]가 생기므로 [조건]에서는 다른 선택이 낫습니다.
확인은 [관측 지표, 명령, 로그, 실행 계획]으로 합니다.
```

이 뼈대는 답변을 기계적으로 만들기 위한 문장이 아닙니다.
면접관이 꼬리 질문을 던졌을 때 "어느 계층의 질문인지" 빠르게 분류하기 위한 기준입니다.

## 이 문서에 주제를 추가하는 기준

이 guide는 두 번째 커리큘럼 문서가 아니라 실전 답변 경로 지도입니다.
따라서 새 주제는 아래 조건을 통과할 때만 들어옵니다.

1. 실제 면접에서 한 질문으로 들어올 가능성이 높습니다.
2. 적어도 두 개 이상의 대주제 문서가 함께 설명되어야 합니다.
3. 기존 섹션의 꼬리 질문 한 줄로는 답변 경로가 충분히 복원되지 않습니다.
4. `첫 30초 답변`, `이어 말할 순서`, `꼬리 질문 지도`, `실무 답변 포인트`, `확인 경로`로 압축해도 핵심이 살아남습니다.

조건을 통과하지 못한 내용은 이 guide에 계속 붙이지 않고, 각 대주제 문서의 deep rewrite 후보로 둡니다.

## 1. 웹 요청 하나는 OS, 프록시, 톰캣, 스프링, DB를 어떻게 지나가나요

**첫 30초 답변**

웹 요청은 브라우저나 클라이언트 코드에서 HTTP 메시지로 만들어지고, TCP 연결과 필요하면 TLS 핸드셰이크를 거쳐 서버 커널의 소켓 버퍼에 도착합니다.
Nginx 같은 리버스 프록시는 그 요청을 파싱해 upstream인 Tomcat이나 Spring Boot 내장 서버로 다시 전달하고, 서블릿 컨테이너는 요청을 `DispatcherServlet`으로 넘깁니다.
그 뒤 컨트롤러, 서비스, 트랜잭션, 커넥션 풀, DB 실행 계획을 거쳐 결과가 반대 방향으로 돌아갑니다.

**이어 말할 순서**

1. 클라이언트는 DNS, TCP 3-way handshake, TLS 핸드셰이크를 거쳐 서버와 연결을 만듭니다.
2. 서버 커널은 NIC에서 받은 패킷을 TCP/IP 스택으로 처리하고, 포트에 바인딩된 리스닝 소켓과 연결별 소켓으로 나눕니다.
3. Nginx는 `accept()`된 연결에서 HTTP 요청을 읽고, 설정에 따라 upstream으로 새 HTTP 요청을 만듭니다.
4. Tomcat은 connector thread나 executor를 통해 요청을 받고, 서블릿 컨테이너 안에서 `DispatcherServlet`까지 전달합니다.
5. Spring MVC는 handler mapping, argument binding, validation, controller, service, repository 흐름으로 이동합니다.
6. DB 접근이 필요하면 HikariCP 같은 커넥션 풀에서 커넥션을 빌리고, 트랜잭션 경계 안에서 SQL을 실행합니다.
7. 응답은 DB result set, domain object, DTO, HTTP response, proxy response, TCP segment 순서로 돌아갑니다.

**꼬리 질문 지도**

- Nginx는 왜 필요한가요: TLS 종료, static file, connection buffering, rate limit, upstream health check, L7 routing을 애플리케이션 앞에서 처리할 수 있습니다.
- Tomcat은 왜 필요한가요: Java 웹 애플리케이션이 HTTP 요청을 서블릿 API와 스레드 모델 위에서 처리할 수 있게 하는 컨테이너입니다.
- keep-alive는 왜 중요한가요: 요청마다 TCP와 TLS 연결을 새로 만들지 않아 handshake 비용과 `TIME_WAIT` 부담을 줄입니다.
- 병목은 어디서 생기나요: listen backlog, file descriptor limit, Nginx worker, Tomcat thread pool, DB connection pool, slow query, lock wait, 외부 API timeout 중 하나가 먼저 포화됩니다.
- WebFlux나 Netty면 무엇이 달라지나요: thread-per-request 모델보다 적은 이벤트 루프 스레드로 많은 연결을 다루지만, CPU 작업이나 blocking call을 event loop에서 실행하면 오히려 전체 처리가 막힙니다.

**실무 답변 포인트**

대용량 트래픽 질문에서는 "서버를 늘립니다"보다 먼저 어떤 자원이 포화되는지 말합니다.
연결 수가 문제면 backlog, file descriptor, worker, keep-alive, idle timeout을 봅니다.
처리 시간이 문제면 애플리케이션 thread dump, DB slow query, p95/p99 latency, upstream timeout을 봅니다.
DB connection pool이 작아서 기다리는지, 너무 커서 DB를 압박하는지도 함께 봅니다.

**확인 경로**

- 관련 문서: [network-web-protocols.md](network-web-protocols.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [concurrency-async-io.md](concurrency-async-io.md)
- 확인 명령 후보: `ss -antp`, `lsof -i :80`, Nginx access/error log, Tomcat access log, Spring actuator metrics, DB slow query log

## 2. 트랜잭션은 ACID에서 MVCC, 격리 수준, 전파, 2PC, 보상 트랜잭션까지 어떻게 이어지나요

**첫 30초 답변**

트랜잭션은 여러 읽기와 쓰기를 하나의 논리적 작업으로 묶는 장치입니다.
ACID는 그 작업이 전부 성공하거나 실패해야 한다는 원자성, 무결성 제약을 깨지 않아야 한다는 일관성, 동시에 실행되어도 서로 망치지 않아야 한다는 격리성, 커밋 뒤 장애가 나도 남아야 한다는 지속성을 말합니다.
MVCC는 특히 격리성을 구현하는 대표 방식으로, 읽는 쪽에는 일정 시점의 스냅샷을 보여 주고 쓰는 쪽은 새 버전을 만들게 해서 읽기와 쓰기 충돌을 줄입니다.

**이어 말할 순서**

1. ACID는 트랜잭션의 목표이고, MVCC와 lock은 그 목표를 구현하는 수단입니다.
2. 격리 수준은 동시 트랜잭션이 서로의 중간 상태를 얼마나 볼 수 있는지 정하는 정책입니다.
3. `READ COMMITTED`는 보통 statement마다 커밋된 최신 상태를 보고, `REPEATABLE READ`는 한 트랜잭션 안에서 일관된 스냅샷을 봅니다.
4. phantom read는 같은 조건으로 다시 조회했을 때 새 row가 보이는 현상인데, SQL 표준 설명과 실제 DB 엔진 구현이 다를 수 있습니다.
5. PostgreSQL은 `REPEATABLE READ`를 스냅샷 격리에 가깝게 구현하므로 phantom까지 막지만, serialization anomaly는 `SERIALIZABLE`에서 다룹니다.
6. Spring의 `@Transactional`은 DB 격리 수준 자체가 아니라, 애플리케이션 메서드 호출을 어떤 트랜잭션 경계에 넣을지 정합니다.
7. 여러 DB나 여러 서비스가 하나의 business transaction에 참여하면 2PC, saga, outbox, idempotency 같은 분산 일관성 문제가 됩니다.

**꼬리 질문 지도**

- 2PC와 2PL은 무엇이 다른가요: 2PC는 여러 참여자에게 commit/rollback 결정을 합의시키는 분산 원자성 프로토콜이고, 2PL은 lock 획득과 해제 단계를 제한해 직렬 가능성을 얻는 동시성 제어 프로토콜입니다.
- PostgreSQL에서 2PC를 왜 쓰나요: 외부 트랜잭션 매니저가 여러 자원에 대한 commit 결정을 조율할 때, DB가 prepared transaction 상태를 durable하게 보관해야 하기 때문입니다.
- 2PC의 위험은 무엇인가요: coordinator 장애나 prepared transaction 방치가 lock과 자원을 오래 잡을 수 있고, 운영 복구 절차가 필요합니다.
- 보상 트랜잭션은 rollback과 같은가요: 아닙니다. 이미 커밋된 로컬 트랜잭션을 의미상 되돌리는 별도 비즈니스 행위입니다. 외부 결제 취소, 포인트 환불, 재고 복구처럼 정확한 과거 상태 복원이 아니라 반대 효과를 내는 작업일 수 있습니다.
- Spring 전파는 왜 중요한가요: service 단위의 원자성을 잡아야 하는데 repository 메서드마다 트랜잭션을 나누면 주문 저장과 결제 저장이 따로 커밋될 수 있습니다.
- checked exception은 왜 주의해야 하나요: Spring 기본 정책에서는 runtime exception 중심으로 rollback 되므로, checked exception을 rollback하려면 정책을 명시해야 합니다.

**실무 답변 포인트**

결제나 주문 질문에서는 "무조건 강한 트랜잭션"이라고 말하지 않습니다.
한 DB 안의 짧은 작업은 DB 트랜잭션으로 묶습니다.
외부 PG, 메시지 발행, 다른 서비스 DB까지 걸리는 긴 흐름은 2PC가 가능한지, 운영 부담이 감당되는지, 실패 보상과 중복 처리 방지가 가능한지로 나누어 답합니다.
실무에서는 local transaction, outbox table, idempotency key, retry, dead-letter queue, compensation을 함께 설계하는 경우가 많습니다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md), [distributed-systems-architecture.md](distributed-systems-architecture.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 실험 후보: 두 DB 세션에서 같은 row와 range를 읽고 수정하며 격리 수준별 결과를 비교합니다. Spring에서는 service 레벨 트랜잭션과 repository 레벨 트랜잭션의 commit 경계를 테스트로 나눕니다.

## 3. 인덱스는 왜 B+Tree를 쓰고 실제 데이터 읽기와 어떻게 연결되나요

**첫 30초 답변**

DB 인덱스는 테이블 전체를 매번 읽지 않기 위해 만든 정렬된 보조 구조입니다.
B+Tree는 한 페이지 안에 많은 key를 담아 fanout을 크게 만들고, 모든 leaf를 같은 깊이에 둬서 최악의 탐색 비용을 작게 유지합니다.
내부 노드는 길 안내용 key를 담고, leaf는 key와 row 위치 또는 실제 row를 담기 때문에, point lookup과 range scan을 모두 안정적으로 처리할 수 있습니다.

**이어 말할 순서**

1. DB는 보통 디스크와 메모리 page 단위로 읽기 때문에, 자료구조도 page I/O를 줄이는 방향이어야 합니다.
2. B+Tree의 internal page는 key와 child pointer를 많이 담아 한 번 내려갈 때 후보 범위를 크게 줄입니다.
3. 모든 leaf가 같은 깊이에 있으므로 특정 key만 유난히 오래 걸리는 편향을 줄입니다.
4. leaf page들이 key 순서로 이어져 있어 range scan이 자연스럽습니다.
5. secondary index를 타면 index leaf에서 row id나 primary key를 얻고, 다시 table이나 clustered index로 실제 row를 찾아갈 수 있습니다.
6. covering index는 쿼리에 필요한 column이 모두 index 안에 있어 table row fetch를 생략합니다.
7. 인덱스는 read를 빠르게 하지만 write, update, delete 때 tree 유지 비용과 page split 비용을 만듭니다.

**꼬리 질문 지도**

- 왜 balanced여야 하나요: tree가 한쪽으로 치우치면 최악의 탐색 비용이 커지고, 특정 key 범위가 계속 느려집니다.
- 왜 깊이가 작아야 하나요: depth 하나가 늘 때마다 page read 후보가 늘어나며, 디스크 I/O나 buffer pool miss가 latency를 크게 키웁니다.
- 깊이가 왜 보통 몇 단계에 머무나요: page 하나에 key를 수백 개 담을 수 있으면 fanout이 커져 수억 건도 root, internal, leaf 몇 단계로 좁혀집니다.
- B Tree와 B+Tree는 무엇이 다른가요: B Tree는 내부 노드에도 data를 둘 수 있고, B+Tree는 실제 data 또는 row reference를 leaf에 모아 range scan과 높은 fanout에 유리합니다.
- clustered index와 heap table은 무엇이 다른가요: clustered 구조에서는 primary key 순서가 실제 row 저장 위치와 강하게 연결되고, heap 구조에서는 index가 별도 row 위치를 가리킵니다.
- index scan이 항상 빠른가요: 아닙니다. selectivity가 낮아 많은 row를 가져오면 random row fetch 비용 때문에 sequential scan이 더 나을 수 있습니다.
- composite index에서 순서가 왜 중요한가요: `(a, b)` index는 `a`로 먼저 좁히는 쿼리에 강하고, `b`만 조건인 쿼리에는 기대한 만큼 도움이 안 될 수 있습니다.

**실무 답변 포인트**

인덱스 질문에서는 "B+Tree라서 빠릅니다"에서 멈추면 약합니다.
page, fanout, buffer pool, selectivity, covering, row fetch, write amplification까지 이어 말해야 합니다.
성능 질문이면 반드시 `EXPLAIN`을 언급합니다.
좋은 답변은 "이 쿼리가 index를 탔는가"보다 "index를 탄 뒤 실제로 몇 row와 몇 page를 읽었는가"까지 봅니다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md)
- 확인 실험 후보: 같은 조건으로 `SELECT indexed_column`, `SELECT *`, 낮은 selectivity 조건, 높은 selectivity 조건을 나눠 `EXPLAIN`과 실제 읽은 row 수를 비교합니다.

## 4. 수억 건 테이블은 인덱스만으로 해결되지 않을 때 어떻게 운영하나요

**첫 30초 답변**

수억 건 테이블 문제는 인덱스 하나를 추가하는 문제가 아니라 working set, write cost, migration risk, query routing을 함께 다루는 운영 문제입니다.
먼저 실제 조회 패턴과 보존 정책을 확인하고, 인덱스 정리, range scan 최적화, archive, partitioning, replica, batch migration을 조합합니다.
서비스 중단이 불가능하면 한 번에 옮기지 않고 작은 batch와 검증 가능한 cutover 경로로 나눠야 합니다.

**이어 말할 순서**

1. 현재 느린 이유를 먼저 분리합니다. full scan인지, index selectivity 문제인지, lock wait인지, buffer pool miss인지, network 전송량인지 확인합니다.
2. 자주 쓰는 쿼리와 오래된 데이터 접근 빈도를 나눠 hot data와 cold data를 분리합니다.
3. 중복 인덱스와 거의 쓰지 않는 인덱스를 정리해 write 비용을 줄입니다.
4. ID나 날짜처럼 라우팅 가능한 기준이 있으면 partitioning이나 table split 후보가 됩니다.
5. 기존 인터페이스를 보존하려면 view, application routing, compatibility API 중 하나를 선택합니다.
6. 데이터 이동은 chunk 단위로 하고, 각 chunk마다 row count, checksum, 누락 범위, rollback 방법을 확인합니다.
7. 최종적으로 `EXPLAIN`, slow query, replication lag, lock wait, batch duration으로 성공 여부를 봅니다.

**꼬리 질문 지도**

- archive와 partitioning은 무엇이 다른가요: archive는 덜 쓰는 데이터를 별도 저장소나 테이블로 이동해 현재 working set을 줄이는 것이고, partitioning은 조회와 저장 경로를 기준에 따라 여러 물리 단위로 나누는 것입니다.
- view로 통합하면 끝인가요: 아닙니다. view 자체에는 보통 별도 index가 없고, optimizer가 각 underlying table의 index를 잘 쓰는지 확인해야 합니다.
- application routing은 왜 위험한가요: table 선택 규칙이 애플리케이션 계약이 되므로, id range나 date range 경계가 틀리면 누락과 중복이 생깁니다.
- batch migration에서 무엇을 확인하나요: source count, target count, min/max id, checksum, failed batch retry, lock time, replication lag, 새 write 경로를 확인합니다.
- 온라인 DDL을 못 쓰는 구버전이면 어떻게 하나요: shadow table, trigger나 dual write, chunk copy, read routing 전환 같은 보수적인 절차가 필요하지만, 구현 복잡도와 장애 시 복구 비용도 함께 설명해야 합니다.

**실무 답변 포인트**

대용량 테이블 질문의 강한 답변은 "새 테이블로 나누겠습니다"가 아닙니다.
어떤 조회를 빠르게 만들고, 어떤 쓰기 비용을 감수하며, 어떤 기간 동안 어떤 fallback이 가능한지를 말해야 합니다.
정산, 결제, 감사 데이터라면 삭제보다 보존과 재조회 가능성이 먼저입니다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md)
- 확인 실험 후보: 대표 쿼리별 `EXPLAIN`, index cardinality, slow query log, batch copy dry-run count, source/target checksum 비교

## 5. 동시성 문제는 race condition, visibility, ordering, deadlock으로 어떻게 나뉘나요

**첫 30초 답변**

동시성 문제는 단순히 두 스레드가 동시에 실행되는 문제가 아니라, 공유 상태를 여러 실행 흐름이 어떤 순서로 읽고 쓰는지의 문제입니다.
race condition은 실행 순서에 따라 결과가 달라지는 문제이고, visibility는 한 스레드의 변경을 다른 스레드가 언제 보느냐의 문제이며, ordering은 CPU나 컴파일러 최적화로 실행 순서가 관측상 달라지는 문제입니다.
lock, `volatile`, atomic operation, immutable object는 각각 다른 문제를 해결합니다.

**이어 말할 순서**

1. 먼저 공유 mutable state가 있는지 확인합니다.
2. `count++` 같은 read-modify-write는 원자적이지 않으므로 race condition이 생길 수 있습니다.
3. 한 스레드가 쓴 값을 다른 스레드가 못 보면 visibility 문제가 됩니다.
4. Java에서는 `synchronized`, `volatile`, `Lock`, atomic class가 happens-before 관계나 원자성을 제공합니다.
5. `synchronized`는 mutual exclusion과 visibility를 함께 제공하지만, `volatile`은 단일 변수 가시성과 ordering에는 강해도 compound operation 원자성은 보장하지 않습니다.
6. lock 순서가 뒤엉키면 deadlock이 생기고, lock을 너무 크게 잡으면 contention 때문에 throughput이 떨어집니다.
7. immutable object와 thread confinement는 공유 변경 자체를 줄여 문제를 단순하게 만듭니다.

**꼬리 질문 지도**

- 읽기 전용 객체는 왜 안전한가요: 객체가 완전히 초기화된 뒤 변경되지 않고 안전하게 publish 되면 여러 스레드가 같은 값을 읽어도 경쟁이 없습니다.
- `volatile int count`로 `count++`가 안전한가요: 아닙니다. `count++`는 읽기, 증가, 쓰기 세 단계라 visibility만으로 원자성이 생기지 않습니다.
- deadlock을 어떻게 줄이나요: lock ordering, timeout, lock scope 축소, try-lock, 더 높은 수준의 concurrent collection을 사용합니다.
- synchronized와 ReentrantLock은 무엇이 다른가요: 둘 다 mutual exclusion을 제공하지만, ReentrantLock은 tryLock, interruptible lock, fair lock 같은 제어 옵션을 제공합니다.
- optimistic locking은 DB lock과 같은가요: 보통 version이나 CAS 값을 비교해 충돌을 나중에 감지하는 방식이며, 자원을 미리 막는 pessimistic lock과 다릅니다.

**실무 답변 포인트**

동시성 답변에서는 용어를 한꺼번에 섞지 않습니다.
"이 문제는 원자성 문제인지, 가시성 문제인지, 순서 문제인지, 진행 보장 문제인지 먼저 나눕니다"라고 말씀드리면 답변이 단단해집니다.
그 뒤에 코드 레벨 수단과 DB 레벨 수단이 같은 이름의 lock이어도 적용 계층과 비용이 다르다는 점을 설명합니다.

**확인 경로**

- 관련 문서: [concurrency-async-io.md](concurrency-async-io.md), [language-runtime.md](language-runtime.md)
- 확인 실험 후보: 의도적으로 `count++`를 여러 스레드에서 반복해 누락을 보고, `AtomicInteger`, `synchronized`, `LongAdder`로 바꿔 결과와 성능을 비교합니다.

## 6. blocking, non-blocking, async, event loop는 서로 어떻게 다른가요

**첫 30초 답변**

blocking과 non-blocking은 호출한 스레드가 기다리는 방식의 차이고, sync와 async는 결과를 받는 방식의 차이입니다.
event loop는 준비된 I/O 이벤트를 적은 수의 스레드가 반복 처리하는 실행 모델입니다.
Linux의 `epoll`은 많은 file descriptor 중 준비된 소켓만 알려 주는 readiness notification이고, 이것만으로 CPU 작업이나 모든 파일 I/O가 자동으로 빨라지는 것은 아닙니다.

**이어 말할 순서**

1. blocking I/O는 syscall이나 library call이 완료될 때까지 현재 스레드가 기다립니다.
2. non-blocking I/O는 지금 준비되지 않았으면 `EAGAIN` 같은 신호로 즉시 돌아오게 합니다.
3. busy loop로 계속 물어보면 CPU를 낭비하므로, `select`, `poll`, `epoll`, `kqueue`, IOCP 같은 감시 구조가 필요해집니다.
4. `epoll_wait()`은 등록된 fd 중 ready 상태가 된 항목을 돌려줍니다.
5. 네트워크 소켓은 ready 상태 변화가 의미 있지만, 일반 파일은 보통 항상 읽기 가능한 것처럼 보이므로 event loop와 맞지 않는 경우가 많습니다.
6. Node.js의 libuv는 네트워크 소켓은 event loop로, 파일 I/O나 일부 DNS, crypto 작업은 thread pool로 처리할 수 있습니다.
7. Netty와 WebFlux는 event loop thread를 막지 않는 코드를 전제로 높은 동시 연결 수를 다룹니다.

**꼬리 질문 지도**

- non-blocking이면 async인가요: 아닙니다. non-blocking read를 직접 반복 호출하면 호출자는 여전히 결과 확인을 직접 관리합니다. async는 완료 통지를 callback, future, coroutine continuation 같은 방식으로 받습니다.
- event loop가 왜 적은 스레드로 많은 연결을 처리하나요: 대부분의 연결이 CPU를 쓰는 시간이 아니라 I/O 대기 시간이기 때문에, 준비된 소켓만 처리하면 됩니다.
- event loop가 언제 망가지나요: handler 안에서 blocking DB call, file read, sleep, CPU heavy 작업을 하면 같은 event loop에 묶인 다른 연결이 밀립니다.
- coroutine은 thread인가요: 아닙니다. coroutine은 중단과 재개 가능한 실행 단위이고, 실제 실행은 thread 위에서 일어납니다.
- io_uring은 epoll과 무엇이 다른가요: epoll은 준비 여부를 알려 주는 모델에 가깝고, io_uring은 제출 큐와 완료 큐로 실제 I/O 요청과 완료를 더 직접적으로 다루는 모델입니다.

**실무 답변 포인트**

비동기 질문에서 "스레드를 적게 씁니다"만 말하면 부족합니다.
대기 시간이 긴 I/O를 다룰 때 event loop가 유리하지만, blocking library를 섞으면 이 장점이 사라진다고 말해야 합니다.
그래서 실무에서는 non-blocking stack을 쓰더라도 DB driver, HTTP client, file I/O, serialization, logging appender까지 함께 확인해야 합니다.

**확인 경로**

- 관련 문서: [concurrency-async-io.md](concurrency-async-io.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 실험 후보: event loop handler에 `sleep`이나 blocking call을 넣고 p99 latency가 어떻게 튀는지 봅니다. Linux에서는 `strace`나 thread dump로 blocking syscall을 확인합니다.

## 7. `java -jar` 이후 JVM, class loader, Spring Boot, AOP proxy는 어떤 순서로 동작하나요

**첫 30초 답변**

`java -jar`는 먼저 OS가 `java` 실행 파일을 프로세스로 만들고, JVM이 JAR의 manifest와 classpath를 해석해 main class를 찾은 뒤 실행합니다.
Spring Boot에서는 `main()`이 `SpringApplication.run()`을 호출하고, application context가 bean definition을 읽어 bean을 만들고 의존성을 연결합니다.
웹 애플리케이션이면 내장 Tomcat이 뜨고, AOP와 `@Transactional` 같은 기능은 실제 객체 앞에 proxy를 세워 메서드 호출을 감쌉니다.

**이어 말할 순서**

1. shell은 `execve` 계열 호출로 `java` 바이너리를 실행합니다.
2. OS loader는 실행 파일과 shared library를 메모리에 매핑하고 프로세스 주소 공간을 준비합니다.
3. JVM은 옵션, classpath, jar manifest, class loader 계층을 준비합니다.
4. class loader는 bootstrap, platform, application, 그리고 Spring Boot의 launcher class loader 같은 계층을 통해 class를 찾습니다.
5. `main()`이 호출되고 `SpringApplication.run()`이 environment, bean definition, auto-configuration, application context를 초기화합니다.
6. 웹이면 embedded servlet container가 port를 열고 `DispatcherServlet`을 등록합니다.
7. AOP proxy는 bean을 감싸 transaction, security, logging 같은 cross-cutting concern을 메서드 호출 앞뒤에 삽입합니다.

**꼬리 질문 지도**

- 부모 위임 모델은 왜 필요한가요: JDK 핵심 class를 애플리케이션이 임의로 덮어쓰지 못하게 하고, 공통 class 충돌을 줄입니다.
- Spring Boot fat jar는 왜 별도 loader가 필요한가요: nested jar 구조를 일반 classpath처럼 읽으려면 Boot loader가 내부 lib와 application class를 해석해야 합니다.
- IoC와 DI는 어떻게 다른가요: IoC는 객체 생성과 연결의 제어권이 컨테이너로 넘어간 구조이고, DI는 그 제어권을 이용해 필요한 의존성을 주입하는 방식입니다.
- `@Transactional`은 왜 self-invocation에서 안 먹을 수 있나요: 같은 객체 안의 `this.method()` 호출은 proxy를 거치지 않기 때문입니다.
- AOP proxy가 JDK dynamic proxy인지 CGLIB인지 왜 중요한가요: interface 기반인지 class subclassing 기반인지에 따라 final method, visibility, type casting 같은 제약이 달라집니다.

**실무 답변 포인트**

Spring 질문은 annotation 이름을 외우는 질문처럼 보여도 실제로는 "누가 객체를 만들고, 누가 호출을 감싸고, 어떤 호출 경로에서 그 감싸기가 빠지는가"를 보는 경우가 많습니다.
따라서 `@Transactional` 답변은 항상 proxy, transaction manager, connection binding, propagation, rollback policy, self-invocation을 함께 묶어야 합니다.

**확인 경로**

- 관련 문서: [language-runtime.md](language-runtime.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md)
- 확인 실험 후보: self-invocation 메서드와 외부 bean 호출 메서드를 나누고, transaction active 여부와 rollback 여부를 테스트합니다.

## 8. Kafka, RabbitMQ, outbox, idempotent consumer는 왜 함께 물리나요

**첫 30초 답변**

메시징은 서비스 사이의 시간, 장애, 처리 속도를 분리하기 위한 구조입니다.
Kafka는 append-only log와 partition을 중심으로 여러 consumer group이 같은 데이터를 각자 offset으로 읽는 모델에 강하고, RabbitMQ는 exchange, queue, routing, ack를 중심으로 작업 전달과 라우팅에 강합니다.
분산 시스템에서는 메시지를 보냈는데 DB commit이 실패하거나, DB commit은 됐는데 메시지 발행이 실패하는 문제가 생기므로 outbox와 idempotent consumer가 함께 나옵니다.

**이어 말할 순서**

1. producer는 이벤트나 command를 broker로 보냅니다.
2. broker는 durable log나 queue에 메시지를 보관하고 consumer에게 전달합니다.
3. consumer는 처리 뒤 ack나 offset commit으로 "어디까지 처리했는지"를 남깁니다.
4. 장애가 나면 같은 메시지가 다시 전달될 수 있으므로 consumer는 idempotent해야 합니다.
5. DB 변경과 메시지 발행을 원자적으로 묶기 어렵기 때문에 outbox table에 이벤트를 같은 DB transaction으로 저장하고, 별도 publisher가 broker로 내보내는 방식이 자주 쓰입니다.
6. 실패 메시지는 retry, backoff, dead-letter queue, manual replay 경로로 관리합니다.

**꼬리 질문 지도**

- Kafka는 왜 throughput이 높은가요: partitioned append-only log, sequential I/O, batching, page cache 활용이 잘 맞기 때문입니다.
- Kafka의 순서는 어디까지 보장되나요: partition 안에서는 순서가 있지만, topic 전체나 여러 partition 전체 순서는 별도 설계 없이는 보장되지 않습니다.
- RabbitMQ는 왜 라우팅에 강한가요: exchange type과 binding key로 direct, fanout, topic 같은 전달 패턴을 표현할 수 있습니다.
- exactly once는 믿어도 되나요: broker 내부 처리 보장과 외부 DB side effect의 exactly once는 다릅니다. 실무 답변에서는 idempotency key와 transaction boundary를 함께 말해야 합니다.
- outbox의 단점은 무엇인가요: outbox table 증가, publisher lag, 중복 발행, cleanup, ordering 처리를 운영해야 합니다.

**실무 답변 포인트**

메시징 답변에서 가장 중요한 것은 "한 번만 처리됩니다"라고 쉽게 말하지 않는 것입니다.
네트워크와 프로세스 장애가 있으면 at-least-once 재전달이 흔하고, 그래서 consumer 쪽 중복 처리 방지가 사실상 필수입니다.
결제나 주문에서는 event id, business idempotency key, unique constraint, 처리 상태 machine을 같이 둡니다.

**확인 경로**

- 관련 문서: [messaging-event-driven.md](messaging-event-driven.md), [distributed-systems-architecture.md](distributed-systems-architecture.md)
- 확인 실험 후보: consumer가 처리 후 ack 전에 죽는 상황, ack 후 DB write 전에 죽는 상황을 나누어 재전달과 중복 처리를 확인합니다.

## 9. 분산 시스템에서 strong consistency, eventual consistency, replication, quorum은 어떻게 엮이나요

**첫 30초 답변**

분산 시스템은 여러 노드가 같은 상태를 다루기 때문에, 장애와 네트워크 지연 속에서 어떤 읽기와 쓰기를 허용할지 정해야 합니다.
strong consistency는 읽는 쪽이 최신 쓰기와 모순되지 않게 보는 성질을 강하게 요구하고, eventual consistency는 더 이상 업데이트가 없다면 복제본들이 결국 같은 값으로 수렴한다는 모델입니다.
replication, quorum, leader election, consensus는 이 선택을 구현하는 수단입니다.

**이어 말할 순서**

1. 단일 DB 안에서는 lock, MVCC, WAL 같은 메커니즘으로 일관성을 다룹니다.
2. 여러 노드로 복제하면 leader, follower, replication log, lag가 생깁니다.
3. follower read를 허용하면 latency와 read capacity는 좋아질 수 있지만 stale read가 생길 수 있습니다.
4. quorum read/write는 여러 replica 중 일정 수 이상의 응답을 요구해 최신성 확률이나 보장을 높입니다.
5. network partition이 생기면 모든 노드가 항상 응답하는 가용성과 모든 응답이 최신이라는 일관성을 동시에 지키기 어렵습니다.
6. consensus는 여러 노드가 하나의 순서나 leader 결정에 합의하도록 도와주지만, latency와 운영 복잡도가 듭니다.
7. business flow에서는 완전한 strong consistency보다 invariant별로 강한 구간과 eventual 구간을 나누는 경우가 많습니다.

**꼬리 질문 지도**

- CAP를 어떻게 말해야 하나요: 평소에는 latency와 consistency tradeoff가 더 많이 보이고, partition 상황에서는 consistency와 availability 중 무엇을 희생할지 명시해야 합니다.
- replication lag가 실제로 어떤 문제를 만드나요: 방금 쓴 주문이 조회 API에서 안 보이거나, 중복 요청 방지 조회가 stale data를 보고 잘못 통과할 수 있습니다.
- read-your-writes는 무엇인가요: 자신이 방금 쓴 값은 이후 읽기에서 볼 수 있어야 한다는 사용자 관점 보장입니다.
- leader election은 왜 필요한가요: 같은 데이터를 여러 노드가 동시에 leader처럼 쓰면 split brain과 충돌이 생깁니다.
- consensus를 모든 곳에 쓰면 되나요: strong ordering이 필요한 metadata나 critical state에는 좋지만, 모든 business event에 적용하면 latency와 가용성 비용이 커질 수 있습니다.

**실무 답변 포인트**

분산 일관성 질문에서는 "정합성이 중요합니다"라는 말보다 어떤 invariant를 반드시 지켜야 하는지 말해야 합니다.
예를 들어 결제 승인 중복 방지, 재고 음수 방지, 포인트 중복 적립 방지처럼 깨지면 돈이나 신뢰가 깨지는 invariant는 강하게 잡고, 알림이나 검색 색인 반영은 eventual로 둘 수 있습니다.

**확인 경로**

- 관련 문서: [distributed-systems-architecture.md](distributed-systems-architecture.md), [database-storage-search-nosql.md](database-storage-search-nosql.md), [messaging-event-driven.md](messaging-event-driven.md)
- 확인 실험 후보: primary write 후 replica read를 강제로 보내 lag를 관측하고, stale read가 business invariant를 깨는지 시나리오 테스트를 만듭니다.

## 10. TCP, HTTP, keep-alive, TLS, reverse proxy는 한 요청 경로에서 어떻게 만나나요

**첫 30초 답변**

TCP는 신뢰성 있는 byte stream을 제공하고, HTTP는 그 byte stream 위에서 request와 response 의미를 정의합니다.
TLS는 HTTP 메시지가 오가기 전에 인증과 키 교환을 수행해 암호화된 채널을 만들며, HTTPS는 HTTP를 TLS 위에 올린 것입니다.
reverse proxy는 클라이언트와 서버 사이에서 연결을 받아 backend로 다시 연결을 만들고, routing, TLS termination, buffering, compression, timeout 같은 정책을 적용합니다.

**이어 말할 순서**

1. TCP는 SYN, SYN-ACK, ACK로 연결을 수립하고, sequence number와 ack로 순서와 재전송을 관리합니다.
2. TLS를 쓰면 HTTP 요청 전 `ClientHello`, `ServerHello`, certificate, key exchange, finished 흐름으로 session key를 합의합니다.
3. HTTP/1.1 keep-alive는 같은 TCP 연결에서 여러 요청을 재사용해 handshake 비용을 줄입니다.
4. HTTP/2는 하나의 연결 안에서 stream multiplexing을 제공하지만, TCP head-of-line blocking 영향은 남을 수 있습니다.
5. reverse proxy는 client-side connection과 upstream connection을 분리합니다.
6. timeout은 client, proxy, app server, DB, external API마다 따로 있고, 서로 맞지 않으면 중간 계층에서 끊깁니다.

**꼬리 질문 지도**

- keep-alive를 너무 길게 두면 무엇이 문제인가요: idle connection이 file descriptor와 memory를 잡고, load balancer나 proxy의 connection table을 압박할 수 있습니다.
- TCP와 HTTP의 timeout은 어떻게 다른가요: TCP 연결 자체의 상태와 HTTP request 처리 시간은 다른 계층의 정책입니다.
- TLS termination은 어디서 하나요: Nginx나 load balancer에서 끝낼 수도 있고, 애플리케이션까지 end-to-end로 가져갈 수도 있습니다. 보안 경계와 운영 편의가 tradeoff입니다.
- proxy와 reverse proxy는 무엇이 다른가요: forward proxy는 client를 대신해 외부로 나가고, reverse proxy는 server 앞에서 client 요청을 받아 backend로 전달합니다.
- streaming 응답은 무엇이 어려운가요: 전체 파일을 memory에 올리지 않아야 하고, backpressure, client disconnect, proxy buffering, timeout을 함께 봐야 합니다.

**실무 답변 포인트**

네트워크 답변에서는 "HTTP는 TCP 위에서 동작합니다"에서 멈추지 않습니다.
실제 장애는 TIME_WAIT 증가, TLS handshake 비용, keep-alive timeout mismatch, proxy buffering, slow client, upstream timeout처럼 계층 사이 설정 불일치에서 자주 나옵니다.

**확인 경로**

- 관련 문서: [network-web-protocols.md](network-web-protocols.md), [security-cryptography.md](security-cryptography.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 명령 후보: `curl -v`, `openssl s_client`, `tcpdump`, `ss -s`, Nginx upstream timing log

## 11. TLS와 암호학 질문은 인증, 키 교환, 암호화, 해시를 어떻게 분리해야 하나요

**첫 30초 답변**

TLS는 단순히 "공개키로 암호화한다"가 아닙니다.
인증서는 서버가 정말 그 도메인의 주체인지 확인하는 데 쓰이고, ECDHE 같은 키 교환은 클라이언트와 서버가 네트워크에 비밀 키를 직접 보내지 않고 shared secret을 만들기 위해 쓰입니다.
실제 데이터 암호화는 핸드셰이크 뒤 만들어진 대칭키로 처리하고, 무결성은 AEAD나 MAC 계열 메커니즘이 담당합니다.

**이어 말할 순서**

1. 클라이언트는 `ClientHello`로 지원 TLS 버전과 cipher suite 후보를 보냅니다.
2. 서버는 인증서와 key exchange 정보를 보내 자신의 신원을 증명하고 session key 재료를 제공합니다.
3. 클라이언트는 인증서 chain, hostname, validity, trust anchor를 확인합니다.
4. ECDHE에서는 양쪽이 임시 개인키와 공개값으로 같은 shared secret을 계산합니다.
5. 양쪽은 shared secret과 hello random들을 사용해 traffic key를 만듭니다.
6. 이후 HTTP 데이터는 대칭키 기반으로 빠르게 암호화되고 무결성 검증됩니다.

**꼬리 질문 지도**

- 전방 비밀성은 무엇인가요: 서버 장기 private key가 나중에 유출되어도 과거 세션의 트래픽을 해독하기 어렵게 하는 성질입니다. 세션마다 임시 key exchange를 쓰기 때문에 가능합니다.
- 암호화와 해시는 무엇이 다른가요: 암호화는 key로 원문과 암호문을 변환하고 복호화가 가능합니다. 해시는 단방향 요약값입니다.
- 서명은 무엇을 보장하나요: private key 소유자가 특정 데이터에 서명했음을 검증하고, 데이터 변조를 탐지하게 해 줍니다.
- password는 암호화하면 되나요: 보통 복호화가 필요 없으므로 salt와 느린 password hashing을 사용합니다.
- 인증서가 있어도 중간자 공격이 가능한가요: trust store, hostname verification, private key 보호, 잘못된 인증서 수락 정책이 깨지면 가능합니다.

**실무 답변 포인트**

TLS 질문에서 가장 흔한 약한 답변은 "공개키로 데이터를 암호화합니다"입니다.
현대 TLS에서는 인증, 키 교환, traffic key derivation, 대칭 암호화가 역할을 나눠 가진다고 설명해야 합니다.
그리고 인증서 검증을 꺼 버리는 `trust all` 코드가 왜 위험한지 바로 이어 말할 수 있어야 합니다.

**확인 경로**

- 관련 문서: [security-cryptography.md](security-cryptography.md), [network-web-protocols.md](network-web-protocols.md)
- 확인 명령 후보: `openssl s_client -connect host:443 -servername host`, `curl -v https://host`

## 12. JVM GC와 런타임 성능은 heap, allocation, pause, throughput을 어떻게 이어 말하나요

**첫 30초 답변**

GC는 더 이상 접근할 수 없는 객체를 찾아 heap memory를 회수하는 런타임 기능입니다.
JVM은 객체 allocation이 많고 heap이 커질수록 GC 비용과 pause가 성능에 영향을 줄 수 있으므로, young generation, old generation, region, root scanning, marking, compaction 같은 메커니즘을 사용합니다.
튜닝의 목표는 "GC를 없애는 것"이 아니라 latency, throughput, memory footprint 중 서비스 목표에 맞는 균형을 잡는 것입니다.

**이어 말할 순서**

1. 객체는 대부분 heap에 할당되고, thread마다 stack frame에는 지역 변수와 호출 상태가 쌓입니다.
2. 많은 객체는 금방 죽으므로 generational GC는 young 영역을 자주 빠르게 청소합니다.
3. 오래 살아남은 객체는 old 영역이나 region으로 이동하고, 여기서 회수 비용이 더 커질 수 있습니다.
4. GC는 root에서 시작해 reachable object graph를 찾고, unreachable object를 회수합니다.
5. Stop-the-world는 GC가 일관된 object graph를 다루기 위해 application thread를 잠시 멈추는 구간입니다.
6. G1, ZGC 같은 collector는 pause time, heap size, throughput 목표가 다릅니다.
7. GC 문제는 log, allocation rate, promotion rate, pause time, heap occupancy로 확인합니다.

**꼬리 질문 지도**

- VM과 runtime과 GC는 같은가요: 아닙니다. VM은 bytecode 같은 중간 표현을 실행하는 환경이고, runtime은 memory, scheduling, library support 같은 실행 지원 전체를 말하며, GC는 그중 memory 회수 기능입니다.
- heap을 크게 하면 해결되나요: OOM 가능성은 줄 수 있지만, pause와 memory footprint가 늘 수 있습니다. 반대로 너무 작으면 GC가 너무 자주 돕니다.
- memory leak은 GC 언어에서 가능한가요: 가능합니다. 더 이상 business 의미가 없는 객체라도 static map, cache, listener, ThreadLocal 등에서 reachable하면 GC가 회수할 수 없습니다.
- `finalize`나 명시적 GC 호출은 좋은가요: 보통 예측 가능성과 성능을 해치므로 일반 해법으로 삼지 않습니다.
- GC 튜닝보다 먼저 볼 것은 무엇인가요: 불필요한 allocation, 큰 object graph, cache 정책, batch size, serialization buffer, query result size를 먼저 봅니다.

**실무 답변 포인트**

GC 답변은 collector 이름 나열보다 "서비스 목표"로 시작합니다.
p99 latency가 중요하면 pause를 보고, batch throughput이 중요하면 전체 처리량을 보며, memory 비용이 중요하면 heap occupancy와 object lifetime을 봅니다.
GC log 없이 collector만 바꾸는 것은 진단이 아니라 추측에 가깝습니다.

**확인 경로**

- 관련 문서: [language-runtime.md](language-runtime.md), [concurrency-async-io.md](concurrency-async-io.md)
- 확인 명령 후보: GC log, Java Flight Recorder, heap dump, allocation profiler, thread dump

## 13. 장애 분석 질문은 관측 지표에서 병목 계층으로 어떻게 좁혀 가나요

**첫 30초 답변**

장애 분석은 증상에서 바로 원인을 찍는 일이 아니라, 어느 계층의 병목인지 좁혀 가는 일입니다.
먼저 사용자 영향과 시간 범위를 고정하고, p50/p95/p99 latency, error rate, traffic, saturation을 봅니다.
그 다음 CPU, memory, GC, thread, connection pool, DB, network, broker lag, external dependency를 순서대로 확인해 병목 계층을 찾습니다.

**이어 말할 순서**

1. 증상과 영향 범위를 먼저 고정합니다. 몇 시부터, 어떤 endpoint, 어떤 사용자, 어떤 error code인지 봅니다.
2. latency와 error가 같이 올랐는지, latency만 올랐는지, traffic 증가와 함께인지 확인합니다.
3. 애플리케이션에서는 thread pool queue, active thread, blocked thread, GC pause, exception rate를 봅니다.
4. DB에서는 connection pool wait, lock wait, slow query, execution plan, replication lag를 봅니다.
5. 네트워크에서는 connect timeout, read timeout, retransmission, DNS, TLS handshake, upstream timing을 봅니다.
6. 메시징에서는 consumer lag, retry storm, DLQ 증가, partition skew를 봅니다.
7. 원인을 찾은 뒤에는 재발 방지 지표와 alert threshold를 남깁니다.

**꼬리 질문 지도**

- CPU가 낮은데 느린 이유는 무엇인가요: I/O wait, lock wait, connection pool wait, external API wait처럼 CPU를 쓰지 않고 기다리는 병목일 수 있습니다.
- p99만 튀는 이유는 무엇인가요: 평균은 정상이어도 일부 요청이 lock, GC, cold cache, slow query, retry, queueing delay를 만날 수 있습니다.
- thread dump에서 무엇을 보나요: 많은 thread가 같은 lock, socket read, DB driver, pool borrow, file I/O에서 멈춰 있는지 봅니다.
- slow query가 없는데 DB가 느릴 수 있나요: connection pool wait, lock wait, network, result set 과다 전송, buffer pool miss, replica lag가 원인일 수 있습니다.
- 장애 대응과 근본 해결은 어떻게 나누나요: 우선 우회나 rate limit으로 영향도를 줄이고, 이후 원인 메커니즘과 재발 방지 장치를 닫습니다.

**실무 답변 포인트**

면접에서 장애 경험을 말할 때는 "로그를 봤습니다"보다 "어떤 가설을 세우고 어떤 관측으로 배제했는지"가 중요합니다.
좋은 답변은 직접 원인과 근본 원인을 나눕니다.
예를 들어 "DB connection pool exhausted"는 직접 원인이고, 그 앞의 slow query, 외부 API 지연, transaction scope 과다, thread leak이 근본 원인일 수 있습니다.

**확인 경로**

- 관련 문서: [problem-solving-code-quality.md](problem-solving-code-quality.md), [network-web-protocols.md](network-web-protocols.md), [database-storage-search-nosql.md](database-storage-search-nosql.md), [messaging-event-driven.md](messaging-event-driven.md)
- 확인 명령 후보: dashboard, log correlation id, thread dump, heap/GC log, DB slow query log, `EXPLAIN`, broker lag metrics

## 14. 알고리즘과 코드 품질 질문은 복잡도, 자료구조, 불변식, 테스트 가능성으로 어떻게 답하나요

**첫 30초 답변**

알고리즘 질문은 정답 코드를 외우는 문제가 아니라, 입력 크기와 제약에서 어떤 자료구조와 불변식을 선택해야 하는지 설명하는 문제입니다.
먼저 완전 탐색이 왜 비싼지 말하고, 중복 계산, 정렬된 성질, 그래프 구조, greedy choice, dynamic programming state 같은 개선 축을 찾습니다.
코드 품질 질문에서는 그 선택이 테스트 가능하고 변경에 강하며, 실패했을 때 어디서 깨졌는지 드러나는지도 함께 봅니다.

**이어 말할 순서**

1. 입력 크기와 시간/메모리 제약을 먼저 확인합니다.
2. 가장 단순한 baseline을 말하고, 그 복잡도가 왜 부족한지 계산합니다.
3. 중복 계산이면 memoization이나 DP를, 빠른 lookup이면 hash map이나 set을, 순서가 중요하면 heap, tree, sort, two pointer를 생각합니다.
4. greedy는 "지금 최선의 선택이 전체 최선으로 이어지는 불변식"을 설명해야 합니다.
5. DP는 state, transition, base case, iteration order를 설명해야 합니다.
6. graph는 node, edge, weight, direction, visited invariant를 먼저 고정합니다.
7. 구현 뒤에는 boundary case와 failure case로 검증합니다.

**꼬리 질문 지도**

- 시간 복잡도만 보면 충분한가요: 아닙니다. memory, constant factor, cache locality, I/O, DB round trip 같은 실제 비용도 봐야 합니다.
- hash map은 항상 O(1)인가요: 평균적으로 빠르지만 hash collision, resizing, memory overhead가 있고, 순서 조건에는 맞지 않습니다.
- DP와 greedy를 어떻게 구분하나요: greedy는 한 번의 선택을 되돌리지 않아도 되는 증명이 필요하고, DP는 여러 부분 문제의 최적 결과를 조합합니다.
- 테스트 더블은 왜 필요한가요: 외부 시스템이나 느린 의존성을 대체해 단위 테스트가 빠르고 결정적으로 실패하게 만듭니다.
- 좋은 코드 설계는 무엇인가요: 책임 경계가 분명하고, 도메인 불변식이 코드 구조에 드러나며, 테스트가 business rule을 설명해야 합니다.

**실무 답변 포인트**

알고리즘과 코드 품질은 별개가 아닙니다.
실무에서는 빠른 알고리즘을 고르는 것만큼, 그 알고리즘의 전제와 실패 조건을 코드와 테스트에 남기는 것이 중요합니다.
면접에서는 "왜 이 자료구조가 이 입력 제약에서 맞는지"와 "나중에 요구가 바뀌면 어디를 바꾸면 되는지"를 함께 답하면 강합니다.

**확인 경로**

- 관련 문서: [problem-solving-code-quality.md](problem-solving-code-quality.md), [dynamic_programming.md](../algorithms/dynamic_programming.md)
- 확인 실험 후보: baseline과 개선안을 같은 입력에서 실행해 시간 증가율을 비교하고, boundary case를 먼저 테스트합니다.

## 15. 검색과 NoSQL은 shard, replica, query fan-out, heap을 어떻게 함께 다루나요

**첫 30초 답변**

검색 엔진과 NoSQL은 RDB 인덱스를 단순히 더 빠르게 만든 버전이 아닙니다.
Elasticsearch 같은 검색 시스템은 데이터를 여러 shard에 나누고, coordinator가 쿼리를 각 shard로 보내 결과를 다시 병합합니다.
replica는 읽기 부하 분산과 장애 시 승격에 쓰이지만, shard 수와 heap 사용량, GC, refresh 지연, reindex 비용이 함께 따라옵니다.

**이어 말할 순서**

1. RDB는 보통 transaction과 정규화된 원본 데이터에 강하고, 검색 엔진은 text search, ranking, aggregation, 분산 fan-out 조회에 강합니다.
2. 문서는 index에 들어갈 때 routing 규칙에 따라 primary shard에 저장되고 replica shard로 복제됩니다.
3. 검색 요청은 coordinator node에 도착하고, coordinator는 관련 shard에 쿼리를 보내 병렬로 실행시킵니다.
4. 각 shard는 자기 local segment에서 검색한 결과를 돌려주고, coordinator는 score, sort, pagination 조건에 맞춰 결과를 병합합니다.
5. replica가 있으면 읽기를 분산하고 primary 장애 시 승격할 수 있지만, replica도 저장 공간과 복제 비용을 씁니다.
6. shard가 너무 많으면 병렬성이 좋아지는 대신 heap metadata, file handle, cluster coordination, merge 비용이 커집니다.
7. 대량 조회는 `from/size`만 키우지 않고 `search_after`, scroll, point-in-time, batch export 같은 방식을 구분합니다.

**꼬리 질문 지도**

- RDB index를 두면 되는데 왜 검색 엔진을 쓰나요: full-text search, 형태소 분석, relevance scoring, 다중 field ranking, 분산 집계가 필요하면 검색 엔진이 더 자연스럽습니다.
- 검색 엔진을 source of truth로 둬도 되나요: 보통 결제, 주문, 계정 같은 원본 상태는 RDB나 transactional store에 두고, 검색 엔진은 조회용 read model로 둡니다.
- refresh와 commit은 같은가요: 아닙니다. 검색 가능해지는 refresh 주기와 durable하게 저장되는 commit/fsync 경로는 다른 개념입니다.
- shard를 많이 만들면 항상 빠른가요: 아닙니다. 작은 shard가 너무 많으면 heap, metadata, merge, recovery, coordination 비용이 늘어납니다.
- OOM은 왜 생기나요: aggregation, fielddata, cache, 너무 큰 result window, 과도한 shard 수, JVM heap 설정과 GC 문제가 함께 원인이 될 수 있습니다.

**실무 답변 포인트**

검색 API 설계 질문에서는 "Elasticsearch를 붙입니다"가 답이 아닙니다.
원본 데이터 변경을 outbox나 CDC로 search index에 반영할지, 검색 결과가 잠깐 stale해도 되는지, 재색인 중 alias 전환을 어떻게 할지, 장애 시 RDB fallback이 필요한지를 함께 말해야 합니다.
검색은 빠른 조회를 주지만, 색인 지연과 운영 복잡도를 새로 만듭니다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md), [distributed-systems-architecture.md](distributed-systems-architecture.md), [messaging-event-driven.md](messaging-event-driven.md)
- 확인 실험 후보: shard 수가 다른 index에서 같은 aggregation을 실행하고 latency, heap usage, GC log, rejected search thread, result window 비용을 비교합니다.

## 16. 조회 API 성능은 애플리케이션과 DB 경계를 어떻게 나누어 보나요

**첫 30초 답변**

조회 API 성능은 DB 인덱스 하나만의 문제가 아니라 요청 파라미터, pagination 방식, SQL 실행 계획, mapper나 ORM의 row 조립, transaction scope, connection pool 대기, JSON 직렬화, response size가 같이 만드는 결과입니다.
먼저 DB가 느린지 애플리케이션이 많이 가져와 많이 버리는지 나누고, 그 다음 실행 계획과 애플리케이션 경계를 각각 줄입니다.

**이어 말할 순서**

1. endpoint별 p95/p99 latency, response size, rows returned, rows examined, pool wait를 먼저 봅니다.
2. DB에서는 `EXPLAIN`, index selectivity, sort/temp table, lock wait, buffer pool miss를 봅니다.
3. 애플리케이션에서는 N+1 query, 반복 mapper 호출, DTO 변환 비용, JSON serialization, compression 비용을 봅니다.
4. pagination은 offset이 커질수록 앞 row를 많이 건너뛸 수 있으므로 cursor/keyset pagination 후보를 검토합니다.
5. 필요한 column만 projection하면 covering index나 smaller row transfer가 가능해집니다.
6. read-only transaction은 영속성 context나 lock 정책, flush 가능성, connection 사용 시간을 줄이는 관점에서 봅니다.
7. cache를 넣을 때는 hit ratio보다 invalidation, stale data 허용 범위, stampede 방지를 먼저 말합니다.

**꼬리 질문 지도**

- N+1은 왜 생기나요: 한 번에 가져올 수 있는 연관 데이터를 row마다 추가 조회하면 네트워크 round trip과 DB execution이 반복됩니다.
- offset pagination은 왜 느려지나요: 뒤 페이지로 갈수록 DB가 앞쪽 row를 세고 버리는 비용이 커질 수 있습니다.
- cursor pagination은 언제 어렵나요: 정렬 기준이 안정적이어야 하고, 중간 삽입/삭제, 복합 정렬, 임의 페이지 이동 요구와 tradeoff가 생깁니다.
- DTO projection은 왜 도움이 되나요: 필요한 column만 가져와 row fetch, network transfer, object allocation, serialization 비용을 줄입니다.
- cache는 언제 위험한가요: 원본 변경과 cache invalidation이 어긋나면 stale read가 business invariant를 깨뜨릴 수 있습니다.

**실무 답변 포인트**

조회 API 장애 경험을 말할 때는 "인덱스를 추가했습니다"에서 끝내지 않습니다.
요청 하나가 DB에서 몇 row를 읽고, 애플리케이션에서 몇 객체를 만들고, 응답으로 몇 MB를 보냈는지 말하면 강합니다.
특히 `SELECT *`, 큰 page size, N+1, 무제한 export, eager loading, JSON 직렬화 비용은 DB와 애플리케이션 사이에서 같이 봐야 합니다.

**확인 경로**

- 관련 문서: [database-storage-search-nosql.md](database-storage-search-nosql.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [problem-solving-code-quality.md](problem-solving-code-quality.md)
- 확인 실험 후보: 같은 API를 `SELECT *`와 projection, offset과 cursor, N+1과 join/fetch batch로 나눠 `EXPLAIN`, query count, response bytes, allocation profile을 비교합니다.

## 17. 파일 스트리밍은 HTTP, TCP, page cache, backpressure를 어떻게 지나가나요

**첫 30초 답변**

큰 파일 스트리밍은 파일을 전부 메모리에 올려 응답하는 방식이 아닙니다.
서버는 파일을 작은 단위로 읽고, 커널 page cache와 애플리케이션 buffer, socket send buffer, TCP window를 거쳐 클라이언트로 흘려보냅니다.
클라이언트나 네트워크가 느리면 backpressure가 생기므로, proxy buffering, timeout, range request, client disconnect까지 함께 봐야 합니다.

**이어 말할 순서**

1. HTTP 요청은 다운로드 대상과 range, 인증, 압축 여부 같은 정책을 결정합니다.
2. 애플리케이션은 파일 descriptor를 열고 일정 chunk 단위로 읽거나, 서버와 OS가 지원하면 `sendfile` 같은 zero-copy 경로를 사용할 수 있습니다.
3. 디스크에서 읽은 data block은 보통 커널 page cache에 올라오고, `read()`는 그 데이터를 user buffer로 복사합니다.
4. user buffer의 데이터는 `write()`나 response stream을 통해 socket send buffer로 들어갑니다.
5. TCP는 상대방 receive window와 congestion window에 맞춰 전송량을 조절합니다.
6. reverse proxy가 buffering을 켜면 애플리케이션과 클라이언트 사이의 흐름이 달라지고, 메모리와 디스크 임시 파일 비용이 생길 수 있습니다.
7. 클라이언트가 끊기면 server는 write 실패나 connection reset을 감지하고 파일 handle과 작업 상태를 정리해야 합니다.

**꼬리 질문 지도**

- streaming이면 메모리를 안 쓰나요: 전체 파일을 한 번에 올리지 않는다는 뜻이지 buffer, page cache, socket buffer가 없어지는 것은 아닙니다.
- `sendfile`은 무엇을 줄이나요: user space로 복사했다가 다시 kernel space로 쓰는 복사를 줄일 수 있습니다.
- 일반 파일도 `epoll`로 처리하면 되나요: 많은 경우 일반 파일은 항상 ready처럼 보이므로 네트워크 소켓과 같은 event readiness 모델이 잘 맞지 않습니다.
- range request는 왜 필요한가요: 이어받기, 동영상 seek, 부분 다운로드에 필요합니다.
- proxy buffering은 언제 문제가 되나요: 큰 응답을 proxy가 먼저 받아 쌓으면 latency, disk temp file, memory 사용량이 커지고 실시간 streaming 성질이 약해질 수 있습니다.

**실무 답변 포인트**

100GB export 질문에서는 "streaming으로 내려줍니다"가 아니라 누가 속도를 조절하는지 말해야 합니다.
DB cursor나 batch read, 파일 생성 위치, page cache 압박, response chunk, proxy timeout, client 재시도, 다운로드 재개 정책을 함께 잡아야 합니다.
파일 다운로드가 느릴 때 CPU가 낮아도 disk I/O, network window, proxy buffering, client 속도가 병목일 수 있습니다.

**확인 경로**

- 관련 문서: [network-web-protocols.md](network-web-protocols.md), [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md), [concurrency-async-io.md](concurrency-async-io.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 명령 후보: `curl -v -r`, `tcpdump`, `ss -ti`, Nginx upstream timing log, `iostat`, `pidstat -d`, application access log의 bytes sent

## 18. 로그인, 세션, 토큰, OAuth는 인증과 인가를 어떻게 분리해 말하나요

**첫 30초 답변**

로그인은 사용자가 누구인지 확인하는 인증이고, 인가는 그 사용자가 어떤 자원에 접근할 수 있는지 결정하는 과정입니다.
세션과 토큰은 인증 결과를 이후 요청에서 다시 증명하기 위한 수단이며, OAuth는 사용자가 제3자 애플리케이션에 자신의 자원 접근 권한을 위임하는 인가 프레임입니다.
Authorization Code 흐름에서는 code를 서버가 token으로 교환하고, `state`는 CSRF 방지, PKCE는 code 가로채기 방지에 쓰입니다.

**이어 말할 순서**

1. HTTPS/TLS는 통신 채널을 보호하지만, 사용자가 누구인지 판단하는 login 자체를 대신하지 않습니다.
2. ID/password login은 credential을 확인하고 session id나 access token 같은 후속 인증 수단을 발급합니다.
3. cookie session은 server-side session store를 참조하고, JWT는 token 자체에 claims와 signature를 담을 수 있습니다.
4. OAuth Authorization Code 흐름은 browser redirect, authorization code, token exchange, resource API 호출로 이어집니다.
5. `state`는 요청을 시작한 사용자 세션과 callback을 묶어 CSRF를 막습니다.
6. PKCE는 code verifier와 code challenge로 authorization code가 탈취되어도 token 교환을 어렵게 만듭니다.
7. token은 만료, refresh, rotation, revocation, 저장 위치, scope를 함께 설계해야 합니다.

**꼬리 질문 지도**

- OAuth는 로그인인가요: OAuth 자체는 인가 프로토콜이고, 로그인 용도로 쓰려면 OpenID Connect처럼 identity layer를 함께 봐야 합니다.
- JWT는 session보다 항상 좋은가요: 아닙니다. stateless 검증은 편하지만 즉시 폐기, 권한 변경 반영, token 탈취 대응이 어려울 수 있습니다.
- CSRF와 XSS는 어떻게 다른가요: CSRF는 사용자의 인증 상태를 악용해 원치 않는 요청을 보내게 하는 공격이고, XSS는 공격자 script가 브라우저 안에서 실행되는 문제입니다.
- CORS는 보안 인증인가요: 아닙니다. 브라우저가 cross-origin 요청을 어떻게 허용할지 정하는 정책이지 서버 간 호출이나 인증 자체를 보장하지 않습니다.
- 비밀번호는 암호화해서 저장하나요: 복호화가 필요 없으므로 salt와 느린 password hashing을 사용합니다.

**실무 답변 포인트**

OAuth 경험을 말할 때는 "토큰을 받았습니다"보다 redirect URI 고정, `state` 검증, PKCE, token scope, 만료와 갱신, 저장 위치, 로그 마스킹까지 이어 말해야 합니다.
내 서비스 로그인과 외부 API 권한 위임을 섞으면 답변이 흐려집니다.
먼저 인증과 인가를 나누고, 그 위에 session, JWT, OAuth token을 배치하면 꼬리 질문에 강해집니다.

**확인 경로**

- 관련 문서: [security-cryptography.md](security-cryptography.md), [network-web-protocols.md](network-web-protocols.md), [spring-backend-frameworks.md](spring-backend-frameworks.md), [source/_source-context-and-question-bank.md](source/_source-context-and-question-bank.md)
- 확인 실험 후보: OAuth callback에서 `state` mismatch, redirect URI mismatch, expired authorization code, wrong PKCE verifier를 각각 만들어 FAIL 경로를 확인합니다.

## 19. 컨테이너에서 Spring 서비스가 느리거나 죽을 때 무엇을 확인하나요

**첫 30초 답변**

컨테이너는 별도 커널을 띄우는 가상 머신이 아니라, 호스트 커널을 공유하면서 namespace로 보이는 세계를 나누고 cgroup으로 자원 사용을 제한하는 실행 단위입니다.
Spring 서비스가 컨테이너 안에서 느리거나 죽으면 애플리케이션 로그만 보지 않고 CPU quota, memory limit, cgroup OOM, PID 1 signal 처리, health check, restart 정책, container network를 함께 확인해야 합니다.

**이어 말할 순서**

1. image는 애플리케이션 파일과 user-space dependency를 담지만, syscall은 호스트 커널이 처리합니다.
2. PID, network, mount, user namespace는 컨테이너가 자기만의 프로세스와 파일 시스템, 네트워크를 보는 것처럼 만듭니다.
3. cgroup은 CPU, memory, pids, block I/O 같은 자원을 제한하고 관측합니다.
4. JVM은 container memory limit을 기준으로 heap, metaspace, direct buffer, thread stack을 함께 써야 합니다.
5. 컨테이너의 PID 1은 signal forwarding과 zombie reaping을 제대로 처리해야 graceful shutdown이 됩니다.
6. liveness/readiness health check는 죽은 프로세스 재시작과 트래픽 투입 가능 여부를 구분해야 합니다.
7. rollout 중에는 startup time, readiness, DB migration, connection drain, rollback 경로를 함께 봐야 합니다.

**꼬리 질문 지도**

- container와 VM은 무엇이 다른가요: VM은 guest kernel을 따로 실행하고, container는 host kernel을 공유합니다.
- 컨테이너 메모리를 넘으면 JVM OOM인가 container OOM인가요: heap OOM일 수도 있고, cgroup limit 초과로 kernel이 프로세스를 죽이는 OOM kill일 수도 있습니다.
- CPU limit이 낮으면 어떤 증상이 생기나요: thread는 많아도 실제 CPU time을 충분히 못 받아 latency가 튀고 GC나 JIT도 늦어질 수 있습니다.
- PID 1 문제는 왜 중요한가요: signal을 무시하거나 자식 프로세스를 회수하지 못하면 graceful shutdown과 zombie process 처리에 문제가 생깁니다.
- health check는 왜 두 종류가 필요한가요: process 생존과 traffic 처리 가능 상태는 다르기 때문입니다.

**실무 답변 포인트**

컨테이너 장애 답변에서는 "Pod를 재시작했습니다"보다 왜 재시작됐는지부터 말합니다.
exit code, previous logs, OOMKilled 여부, cgroup memory, CPU throttling, readiness 실패, probe timeout, connection drain 실패를 순서대로 보면 운영 경험이 드러납니다.
Spring Boot라면 actuator health와 metrics를 probe와 dashboard에 어떻게 연결했는지도 함께 말합니다.

**확인 경로**

- 관련 문서: [problem-solving-code-quality.md](problem-solving-code-quality.md), [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md), [spring-backend-frameworks.md](spring-backend-frameworks.md)
- 확인 명령 후보: `docker inspect`, `docker stats`, container logs, `/sys/fs/cgroup` 지표, actuator health, orchestrator event log

## 20. 고가용성은 replication, 장애 격리, health check, failover를 어떻게 설계하나요

**첫 30초 답변**

고가용성은 장애가 없다는 뜻이 아니라, 장애가 나도 약속한 핵심 기능을 계속 제공하거나 빠르게 복구하는 능력입니다.
replication은 데이터나 인스턴스를 여러 곳에 두는 수단이고, 장애 격리와 health check, failover, graceful degradation은 장애가 전체 서비스로 번지는 것을 줄이는 운영 설계입니다.
일관성 질문과 겹치지만, 고가용성 답변의 중심은 최신성보다 영향 범위와 복구 시간입니다.

**이어 말할 순서**

1. 먼저 보호할 사용자 기능과 SLO를 정합니다. 모든 기능이 같은 가용성 목표를 갖지는 않습니다.
2. 단일 장애점이 있는지 봅니다. load balancer, DB primary, cache, broker, external API가 후보가 됩니다.
3. replication은 read capacity와 failover 후보를 만들지만, replication lag와 split brain 위험을 만듭니다.
4. health check는 process alive, dependency availability, traffic readiness를 구분해야 합니다.
5. failover는 감지, 승격, routing 전환, stale client connection 처리, rollback 경로가 필요합니다.
6. 장애 격리는 timeout, circuit breaker, bulkhead, queue, rate limit으로 연쇄 장애를 막습니다.
7. graceful degradation은 비핵심 기능을 줄이고 핵심 기능을 살리는 정책입니다.

**꼬리 질문 지도**

- replication만 하면 HA인가요: 아닙니다. 장애 감지, 승격, routing 전환, 데이터 정합성, 운영 runbook이 없으면 복제본이 있어도 장애가 길어집니다.
- active-active는 항상 좋은가요: write conflict, session affinity, cache consistency, idempotency, split brain을 해결해야 합니다.
- health check가 너무 엄격하면 어떻게 되나요: 일시적인 dependency 지연으로 정상 인스턴스가 계속 빠져나가 오히려 장애가 커질 수 있습니다.
- backup과 HA는 같은가요: backup은 복구 자료이고, HA는 서비스 지속과 빠른 복구 구조입니다.
- degraded mode는 왜 필요한가요: 추천, 알림, 검색 같은 부가 기능을 잠시 줄여 주문, 결제, 로그인 같은 핵심 기능을 살릴 수 있습니다.

**실무 답변 포인트**

시스템 설계 질문에서 "서버를 이중화합니다"만 말하면 부족합니다.
어떤 장애를 어떤 시간 안에 감지하고, 누가 traffic을 빼고, 어떤 데이터는 stale해도 되고, 어떤 기능은 중단해도 되는지 말해야 합니다.
가용성은 topology 그림보다 운영 절차와 실패 모드까지 함께 있어야 설득됩니다.

**확인 경로**

- 관련 문서: [distributed-systems-architecture.md](distributed-systems-architecture.md), [messaging-event-driven.md](messaging-event-driven.md), [database-storage-search-nosql.md](database-storage-search-nosql.md), [problem-solving-code-quality.md](problem-solving-code-quality.md)
- 확인 실험 후보: primary DB 중단, broker 중단, external API timeout, cache 장애를 각각 주입하고 failover time, error rate, degraded behavior를 기록합니다.

## 21. Spring의 IoC, DI, AOP proxy, test double은 객체 경계와 테스트 가능성을 어떻게 만드나요

**첫 30초 답변**

IoC는 객체 생성과 연결의 제어권을 애플리케이션 코드 밖의 컨테이너로 옮기는 구조이고, DI는 그 구조를 이용해 필요한 의존성을 주입하는 방식입니다.
AOP proxy는 객체 호출 앞뒤를 감싸 transaction, logging, security 같은 공통 기능을 넣습니다.
test double은 같은 의존성 경계에 fake, stub, mock을 넣어 외부 시스템 없이 business rule을 검증하게 해 줍니다.

**이어 말할 순서**

1. Spring은 component scan, `@Bean`, configuration class를 통해 bean definition을 만듭니다.
2. container는 bean을 생성하고 constructor, field, method 등을 통해 dependency를 연결합니다.
3. interface나 class 경계가 분명하면 실제 구현 대신 test double을 넣기 쉽습니다.
4. AOP proxy는 실제 bean 앞에 대리 객체를 세워 method call을 감쌉니다.
5. 같은 객체 안의 `this.method()` 호출은 proxy를 거치지 않으므로 transaction이나 security advice가 빠질 수 있습니다.
6. mock은 호출 행위를 검증하는 데 강하고, stub은 정해진 응답을 돌려주는 데 강합니다.
7. 과한 mocking은 실제 wiring, SQL, transaction, serialization 문제를 숨길 수 있으므로 slice test와 integration test가 필요합니다.

**꼬리 질문 지도**

- `@Bean`과 `@Component`는 무엇이 다른가요: `@Component`는 class를 scan해 등록하고, `@Bean`은 configuration method가 반환한 객체를 등록합니다.
- DI가 왜 테스트에 좋나요: 객체가 자기 의존성을 직접 만들지 않으므로 테스트에서 대체 구현을 주입할 수 있습니다.
- JDK dynamic proxy와 CGLIB은 왜 중요한가요: interface 기반 proxy인지 class subclass proxy인지에 따라 final class, final method, type casting 제약이 달라집니다.
- mock과 stub은 어떻게 다른가요: stub은 상태나 반환값 중심이고, mock은 호출 여부와 interaction 검증 중심입니다.
- `@SpringBootTest`만 쓰면 충분한가요: 전체 wiring 검증에는 좋지만 느리고 실패 지점이 넓습니다. 단위, slice, integration test를 목적별로 나누는 편이 좋습니다.

**실무 답변 포인트**

Spring 설계 질문에서는 "DI를 쓰면 결합도가 낮아집니다"에서 멈추지 않습니다.
어떤 dependency가 business rule이고 어떤 dependency가 외부 I/O인지 나누고, 외부 I/O 경계만 test double로 바꾸며, transaction이나 proxy가 실제로 필요한 경우에는 integration test로 한 번 닫는다고 말하면 강합니다.
테스트 가능한 구조는 mock을 많이 쓰는 구조가 아니라 책임 경계가 잘 보이는 구조입니다.

**확인 경로**

- 관련 문서: [spring-backend-frameworks.md](spring-backend-frameworks.md), [problem-solving-code-quality.md](problem-solving-code-quality.md), [language-runtime.md](language-runtime.md)
- 확인 실험 후보: 생성자 주입 service를 fake repository로 단위 테스트하고, self-invocation transaction case는 Spring context integration test로 확인합니다.

## 22. 숫자 표현, CPU cache, memory barrier는 정확성과 성능 질문에서 어떻게 등장하나요

**첫 30초 답변**

낮은 계층 지식은 면접에서 갑자기 하드웨어 강의로 나오는 것이 아니라, 정확성과 성능의 경계 조건으로 등장합니다.
금융이나 정산에서는 부동소수점 오차와 integer overflow가 문제가 되고, 멀티스레드에서는 CPU cache와 reordering 때문에 visibility와 ordering 문제가 생깁니다.
성능 분석에서는 cache miss, context switch, NUMA, memory bandwidth가 "코드는 같지만 왜 느린가"를 설명하는 단서가 됩니다.

**이어 말할 순서**

1. 정수는 고정된 bit 폭 안에서 표현되므로 범위를 넘으면 overflow가 생깁니다.
2. 부동소수점은 이진수 근사 표현이므로 `0.1` 같은 십진수가 정확히 표현되지 않을 수 있습니다.
3. CPU는 main memory보다 훨씬 빠른 cache를 계층적으로 두고, cache line 단위로 데이터를 가져옵니다.
4. 멀티코어에서는 각 core의 cache가 같은 memory 값을 서로 다르게 볼 수 있으므로 cache coherence protocol이 필요합니다.
5. 컴파일러와 CPU는 성능을 위해 instruction reordering을 할 수 있고, memory barrier와 happens-before가 관측 순서를 제한합니다.
6. context switch는 register, scheduler queue, cache locality에 비용을 만듭니다.
7. NUMA 환경에서는 어떤 CPU가 어느 memory bank에 접근하느냐에 따라 latency가 달라집니다.

**꼬리 질문 지도**

- 돈 계산에 `double`을 쓰면 왜 위험한가요: 십진 소수를 이진 부동소수점으로 정확히 표현하지 못해 누적 오차가 생길 수 있습니다.
- `volatile`은 무엇을 보장하나요: visibility와 일정한 ordering에는 도움을 주지만 compound operation의 원자성은 보장하지 않습니다.
- false sharing은 무엇인가요: 서로 다른 변수가 같은 cache line에 있어 여러 core가 불필요하게 cache line을 주고받는 현상입니다.
- context switch는 왜 비용인가요: 실행 상태 저장/복원뿐 아니라 cache와 TLB locality가 깨질 수 있습니다.
- NUMA는 언제 중요한가요: 큰 heap, 많은 thread, memory bandwidth가 중요한 서버에서 remote memory 접근 비용이 보일 수 있습니다.

**실무 답변 포인트**

정확성 질문에서는 금액, 시간, ID, bit flag처럼 "틀리면 안 되는 값"의 표현을 먼저 말합니다.
성능 질문에서는 CPU 사용률만 보지 말고 cache miss, lock contention, context switch, allocation, GC, NUMA 같은 낮은 계층 신호를 후보로 둡니다.
다만 모든 문제를 하드웨어로 끌고 가지 말고, 애플리케이션/DB/네트워크 원인이 배제된 뒤 낮은 계층으로 내려가는 편이 안전합니다.

**확인 경로**

- 관련 문서: [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md), [concurrency-async-io.md](concurrency-async-io.md), [language-runtime.md](language-runtime.md), [problem-solving-code-quality.md](problem-solving-code-quality.md)
- 확인 실험 후보: `0.1 + 0.2`, integer overflow, unsafely shared flag, false sharing microbenchmark, `perf stat`의 context switch/cache miss를 각각 관찰합니다.

## 23. VM, runtime, GC, scheduler는 언어 실행 모델 비교에서 어떻게 나뉘나요

**첫 30초 답변**

VM, runtime, GC, scheduler는 같은 말이 아닙니다.
VM은 bytecode 같은 중간 표현을 실행하는 환경이고, runtime은 memory 관리, thread나 coroutine scheduling, standard library support, system call 연결 같은 실행 지원 전체입니다.
JVM은 bytecode와 JIT, GC를 갖고, Go는 VM 없이 native binary로 실행되지만 runtime scheduler와 GC를 갖고, PHP는 Zend Engine과 OPcache, request 처리 모델을 중심으로 설명해야 합니다.

**이어 말할 순서**

1. Java는 `.java`를 `.class` bytecode로 컴파일하고, JVM이 class loading, verification, interpretation/JIT compilation을 담당합니다.
2. JIT은 자주 실행되는 hot path를 native code로 바꿔 성능을 높이지만 warm-up과 code cache, profiling 비용이 있습니다.
3. PHP는 보통 request마다 script를 Zend Engine이 해석하거나 opcode로 실행하고, OPcache가 컴파일된 opcode 재사용을 돕습니다.
4. Go는 native binary로 컴파일되지만 goroutine scheduler, GC, stack growth, network poller 같은 runtime 기능을 포함합니다.
5. Kotlin coroutine은 OS thread가 아니라 중단과 재개 가능한 실행 단위이며, JVM 위에서는 dispatcher가 thread pool에 coroutine을 배치합니다.
6. blocking call을 만났을 때 누가 다른 작업을 계속 실행하게 하는지가 runtime 비교의 핵심입니다.
7. 성능 비교는 "언어가 빠르다"가 아니라 startup, warm-up, GC pause, memory footprint, concurrency model, deployment model로 나눠야 합니다.

**꼬리 질문 지도**

- Go는 runtime이 있는데 VM이 없다는 말이 무슨 뜻인가요: bytecode VM 위에서 실행하지는 않지만, scheduler와 GC 같은 실행 지원 코드는 binary와 함께 들어갑니다.
- coroutine은 thread보다 왜 가볍나요: OS thread stack과 kernel scheduling을 직접 늘리는 대신, 중단 지점의 continuation과 dispatcher로 실행을 나눌 수 있기 때문입니다.
- JIT은 항상 빠른가요: long-running hot path에는 유리하지만, 짧은 실행이나 cold start에서는 warm-up 비용이 보일 수 있습니다.
- PHP-FPM은 왜 process pool을 쓰나요: request 격리와 운영 단순성이 있지만, request마다 상태를 오래 유지하는 서버와는 다른 성능 모델을 가집니다.
- GC가 있는 언어는 memory leak이 없나요: business 의미상 더 이상 필요 없는 객체가 reachable하면 GC가 회수하지 못합니다.

**실무 답변 포인트**

언어 비교 질문에서 "Go는 빠르고 Java는 무겁습니다"처럼 말하면 약합니다.
서버가 long-running인지, cold start가 중요한지, concurrency가 I/O 대기 중심인지 CPU 중심인지, memory footprint와 GC pause가 서비스 목표에 어떤 영향을 주는지 나눠야 합니다.
런타임 선택은 문법 취향보다 배포, 운영, 관측, 팀 숙련도까지 함께 보는 판단입니다.

**확인 경로**

- 관련 문서: [language-runtime.md](language-runtime.md), [concurrency-async-io.md](concurrency-async-io.md), [os-kernel-computer-architecture.md](os-kernel-computer-architecture.md)
- 확인 실험 후보: 같은 HTTP echo나 CPU-bound loop를 JVM, Go, PHP-FPM 모델로 나눠 startup time, steady-state latency, memory, GC log, thread/goroutine count를 비교합니다.

## 복합 질문으로 연습하기

이 절은 실제 면접처럼 한 질문에서 여러 주제가 한꺼번에 따라오는 연습용입니다.
각 질문은 외울 답이 아니라 말하는 순서를 익히기 위한 것입니다.

### 결제 승인 API를 설계해 보세요

**첫 답변**

결제 승인 API는 중복 승인 방지, 외부 PG 장애 처리, DB 원자성, 메시지 발행, 운영 추적성을 함께 만족해야 합니다.
요청이 들어오면 idempotency key로 중복을 먼저 막고, local DB transaction 안에서 결제 시도 상태와 outbox event를 함께 기록합니다.
외부 PG 호출은 timeout과 retry 정책을 명확히 두고, 성공이나 실패 결과를 상태 machine으로 반영하며, 후속 알림이나 정산 이벤트는 outbox publisher가 broker로 내보냅니다.

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

먼저 어느 계층에서 느린지 좁힙니다.
API p99가 튀는지, DB slow query가 늘었는지, connection pool wait가 생겼는지, GC나 thread block이 있는지 확인합니다.
DB 쪽이면 실행 계획이 index scan에서 full scan으로 바뀌었는지, selectivity가 나빠졌는지, covering index가 깨졌는지, lock wait나 buffer pool miss가 있는지 봅니다.

**이어갈 꼬리**

1. 관측: p99, error, QPS, pool wait, slow query
2. DB: `EXPLAIN`, rows examined, index cardinality, filesort, temporary table
3. 인덱스: composite order, covering, range condition, row fetch
4. 테이블 운영: archive, partitioning, hot/cold data
5. 애플리케이션: pagination, N+1 query, serialization cost, response size
6. 네트워크: large payload, proxy buffering, client disconnect

### WebFlux로 바꾸면 서버가 더 빨라지나요

**첫 답변**

항상 빨라지는 것은 아닙니다.
WebFlux나 Netty 기반 non-blocking stack은 많은 I/O 대기 연결을 적은 event loop thread로 다룰 때 유리합니다.
하지만 DB driver, file I/O, 외부 library, JSON 처리, 암호화 같은 blocking 또는 CPU-heavy 작업을 event loop에서 실행하면 오히려 전체 연결이 같이 밀릴 수 있습니다.

**이어갈 꼬리**

1. 개념 구분: blocking vs non-blocking, sync vs async
2. OS: socket fd, `epoll`, readiness event
3. 런타임: event loop, worker pool, backpressure
4. Spring: MVC thread-per-request와 WebFlux event loop 모델
5. DB: R2DBC 같은 non-blocking driver가 필요한 이유
6. 운영: thread dump 대신 event loop latency, queue depth, blocking call detection을 본다

### 로그인과 HTTPS 보안을 설명해 보세요

**첫 답변**

HTTPS는 TLS로 통신 채널을 보호하고, 로그인은 사용자의 인증 정보를 검증해 application session이나 token을 발급하는 별도 문제입니다.
TLS에서는 인증서로 서버 신원을 확인하고, ECDHE 같은 키 교환으로 session key를 만든 뒤, 실제 HTTP 데이터는 대칭키로 암호화합니다.
비밀번호는 복호화할 필요가 없으므로 salt와 느린 password hashing으로 저장하고, session cookie나 token은 탈취와 재사용을 막기 위해 만료, secure flag, same-site, rotation 같은 정책을 둡니다.

**이어갈 꼬리**

1. TCP/TLS: handshake, certificate, ECDHE, traffic key
2. 암호학: encryption, hash, signature, MAC 구분
3. 웹 보안: cookie, SameSite, CSRF, XSS, CORS
4. 운영: certificate renewal, private key 보호, TLS termination 위치
5. 장애: handshake failure, expired certificate, wrong SNI, trust store 문제

### 검색 API를 설계해 보세요

**첫 답변**

상품 검색 API는 RDB query 하나로만 보지 않습니다.
원본 상품과 재고, 가격은 transactional store에서 관리하고, 검색용 문서는 Elasticsearch 같은 search index로 비동기 반영할 수 있습니다.
검색 요청은 coordinator가 shard로 fan-out하고 결과를 병합하므로, shard sizing, refresh 지연, ranking, pagination, stale data 허용 범위를 함께 설계해야 합니다.

**이어갈 꼬리**

1. 원본성: RDB source of truth, search index read model
2. 반영 경로: outbox, CDC, reindex, alias switch
3. 검색 실행: analyzer, shard fan-out, replica read, result merge
4. pagination: `from/size`, `search_after`, scroll, export
5. 운영: heap, GC, shard count, slow query, rejected search thread
6. 일관성: 가격/재고 stale 허용 여부, fallback, 재시도

### 100GB 파일을 다운로드하게 해야 합니다

**첫 답변**

파일 전체를 메모리에 올리지 않고 chunk 단위로 흘려보내야 합니다.
파일은 disk와 page cache를 거쳐 애플리케이션 buffer나 zero-copy 경로로 socket send buffer에 들어가고, TCP window와 클라이언트 속도에 맞춰 전송됩니다.
따라서 range request, client disconnect, proxy buffering, timeout, disk I/O, network throughput을 함께 설계합니다.

**이어갈 꼬리**

1. HTTP: auth, range, content length, chunked transfer
2. OS: file descriptor, page cache, `read`, `sendfile`, zero-copy
3. 네트워크: TCP window, congestion control, slow client
4. Spring: streaming response, worker thread, blocking file I/O
5. Proxy: buffering, temp file, timeout, upstream timing
6. 운영: `iostat`, bytes sent, client abort, retry/resume

### 컨테이너 배포 후 서비스가 OOMKilled로 재시작됩니다

**첫 답변**

먼저 JVM heap OOM인지 cgroup memory limit 초과로 kernel이 프로세스를 죽인 것인지 나눕니다.
컨테이너는 host kernel을 공유하고 cgroup limit 안에서 heap, metaspace, direct buffer, thread stack, native memory를 모두 써야 합니다.
따라서 container event, previous logs, heap dump, GC log, memory limit, CPU throttling, readiness 실패를 함께 봅니다.

**이어갈 꼬리**

1. 컨테이너 모델: namespace, cgroup, shared kernel
2. JVM memory: heap, metaspace, direct buffer, thread stack
3. 장애 신호: exit code, OOMKilled, GC overhead, native memory
4. 운영: health check, graceful shutdown, PID 1, restart policy
5. 배포: rolling update, readiness gate, connection drain, rollback
6. 재발 방지: memory budget, request/limit, alert, load test

### OAuth 기반 외부 API 연동을 설계해 보세요

**첫 답변**

OAuth 연동은 사용자 로그인 문장으로만 답하면 부족합니다.
사용자 동의로 authorization code를 받고, 서버가 code를 access token으로 교환해 외부 API를 호출하며, `state`와 PKCE로 callback 위조와 code 탈취를 막습니다.
토큰은 scope, 만료, refresh, 저장 위치, 암호화 또는 접근 통제, 로그 마스킹까지 운영 정책을 가져야 합니다.

**이어갈 꼬리**

1. 인증/인가 구분: login, OAuth, OpenID Connect
2. 흐름: redirect, authorization code, token exchange, resource API
3. 방어: redirect URI, `state`, PKCE, HTTPS
4. token 운영: access token, refresh token, expiry, rotation, revocation
5. 저장: DB encryption, secret manager, log masking, least privilege
6. 장애: expired token, consent revoked, rate limit, retry/backoff

## 낯선 복합 질문을 받았을 때 복구 루틴

준비하지 못한 질문이 들어오면 바로 정답을 꾸미지 않습니다.
아래 다섯 단계로 질문을 분류하면 모르는 부분을 정직하게 남기면서도 사고 과정을 계속 보여 줄 수 있습니다.

1. 질문이 어느 계층의 문제인지 먼저 나눕니다.

    예를 들어 "서버가 느립니다"는 애플리케이션, 런타임, OS, DB, 네트워크, 외부 의존성 중 어디인지 아직 모릅니다.

2. 확실한 사실과 추론을 분리합니다.

    "p99만 튄다면 일부 요청이 lock, GC, slow query, retry, queueing을 만났을 수 있습니다"처럼 가능성을 말하되, 아직 원인이라고 단정하지 않습니다.

3. 가장 먼저 깨지면 안 되는 불변식을 말합니다.

    결제라면 중복 승인 방지, 검색이라면 원본 데이터와 index의 허용 가능한 지연, 인증이라면 token 탈취 방지가 불변식이 됩니다.

4. 확인 경로를 제안합니다.

    `EXPLAIN`, thread dump, GC log, packet capture, broker lag, container event처럼 어느 증거가 어떤 가설을 통과 또는 탈락시키는지 말합니다.

5. 모르는 부분은 좁혀서 인정합니다.

    "해당 DB 엔진의 정확한 격리 구현은 확인이 필요하지만, 제가 먼저 볼 경로는 MVCC snapshot과 lock wait입니다"처럼 범위를 좁히면 답변이 무너지지 않습니다.

## 마지막 점검 질문

문서를 덮고 아래 질문에 자기 말로 답해 봅니다.
막히는 질문이 있으면 해당 대주제 문서로 돌아가 세부를 다시 읽습니다.

1. `@Transactional` 질문에서 DB 격리 수준, Spring proxy, propagation, rollback policy, distributed transaction을 한 흐름으로 설명할 수 있나요.
2. B+Tree 질문에서 page, fanout, depth, leaf link, row fetch, covering index, write cost를 빠뜨리지 않고 말할 수 있나요.
3. event loop 질문에서 `epoll`이 "완료"가 아니라 "준비됨"을 알려 주는 모델이라는 점을 설명할 수 있나요.
4. 메시징 질문에서 broker의 전달 보장과 외부 DB side effect의 정확히 한 번 처리가 다른 문제임을 말할 수 있나요.
5. 장애 질문에서 직접 원인과 근본 원인을 분리하고, 어떤 관측으로 가설을 배제했는지 설명할 수 있나요.
6. 보안 질문에서 인증서, 키 교환, 세션 키, 대칭 암호, 해시, 서명을 서로 섞지 않고 설명할 수 있나요.
7. 시스템 설계 질문에서 작은 기술 단위의 제약을 큰 아키텍처 선택으로 연결할 수 있나요.
8. 검색/NoSQL 질문에서 shard fan-out, replica, result merge, heap/GC, stale index를 한 흐름으로 설명할 수 있나요.
9. 조회 API 성능 질문에서 N+1, pagination, projection, connection pool, serialization, response size를 DB와 애플리케이션 경계로 나누어 볼 수 있나요.
10. 큰 파일 스트리밍 질문에서 page cache, socket buffer, TCP window, proxy buffering, backpressure를 빠뜨리지 않고 말할 수 있나요.
11. OAuth 질문에서 login, authorization, authorization code, `state`, PKCE, token lifecycle을 섞지 않고 설명할 수 있나요.
12. 컨테이너 장애 질문에서 namespace, cgroup, PID 1, OOMKilled, health check, graceful shutdown을 연결할 수 있나요.
13. 고가용성 질문에서 replication과 failover, 장애 격리, graceful degradation, MTTR을 구분할 수 있나요.
14. Spring 설계 질문에서 IoC/DI, AOP proxy, test double, integration test의 역할을 한 경계 모델로 설명할 수 있나요.
15. 낮은 계층 질문에서 부동소수점, overflow, CPU cache, memory barrier, context switch가 정확성과 성능에 어떻게 나타나는지 말할 수 있나요.
16. 언어 런타임 비교 질문에서 VM, runtime, GC, scheduler, JIT, coroutine/goroutine을 역할별로 나누어 설명할 수 있나요.
