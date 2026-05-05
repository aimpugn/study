# Interview Question And Source Index

이 파일은 원문 chunk가 최종 curriculum의 어느 위치에 들어갔는지 빠르게 찾기 위한 색인입니다.

| 대분류 | 중분류 | 소분류 | 원문 제목 | Source | Aliases |
|---|---|---|---|---|---|
| Source Context And Question Bank | 원재료 메타데이터 | 파일 앞부분과 생성 맥락 | [source front matter] | `interview_questions.md:1-99` | `interviews.md:1-99` |
| 10. 문제 해결, 코드 품질, 운영 실천 | 운영 실천과 배포 기반 | Docker 컨테이너 가상화 | Docker 컨테이너 가상화 | `interview_questions.md:100-232` | `interviews.md:100-232` |
| 03. OS, 커널, 컴퓨터 구조 | CPU와 숫자 표현 | 부동소수점 | 부동소수점 | `interview_questions.md:233-634` | `interviews.md:233-634` |
| 02. 동시성, 비동기, I/O | 공유 상태와 경쟁 조건 | 경쟁 조건 (Race Condition) | 경쟁 조건 (Race Condition) | `interview_questions.md:635-1098` | `interviews.md:635-1098` |
| 02. 동시성, 비동기, I/O | 교착과 진행 보장 | 교착 상태 (Deadlock) | 교착 상태 (Deadlock) | `interview_questions.md:1099-1288` | `interviews.md:1099-1288` |
| 04. 네트워크와 웹 프로토콜 | HTTP/gRPC와 스트리밍 | 네트워크를 통한 스트리밍 방식 | 네트워크를 통한 스트리밍 방식 | `interview_questions.md:1289-1514` | `interviews.md:1289-1514` |
| 04. 네트워크와 웹 프로토콜 | HTTP/gRPC와 스트리밍 | 서버에서 100GB 파일을 스트리밍 방식으로 읽는 방법 (CPU 4코어, RAM 16GB, SWAP 16GB 조건에서) | 서버에서 100GB 파일을 스트리밍 방식으로 읽는 방법 (CPU 4코어, RAM 16GB, SWAP 16GB 조건에서) | `interview_questions.md:1515-1661` | `interviews.md:1515-1661` |
| 08. 분산 시스템과 아키텍처 | 일관성과 분산 트랜잭션 | eventual consistency | eventual consistency | `interview_questions.md:1662-1674` | `interviews.md:1662-1674` |
| 08. 분산 시스템과 아키텍처 | 가용성, 토폴로지, 확장성 | 토폴로지 | 토폴로지 | `interview_questions.md:1675-1734` | `interviews.md:1675-1734` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | DB 접근과 저장소 운영 | DB와 레플리케이션과 고가용성 | DB와 레플리케이션과 고가용성 | `interview_questions.md:1735-2084` | `interviews.md:1735-2084` |
| 08. 분산 시스템과 아키텍처 | 가용성, 토폴로지, 확장성 | Distributed Systems에서 고가용성이란? | Distributed Systems에서 고가용성이란? | `interview_questions.md:2085-2322` | `interviews.md:2085-2322` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 트랜잭션, 락, 격리 | [DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html) | [DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html) | `interview_questions.md:2323-2521` | `interviews.md:2323-2521` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 트랜잭션, 락, 격리 | ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어) | ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어) | `interview_questions.md:2522-2576` | `interviews.md:2522-2576` |
| 04. 네트워크와 웹 프로토콜 | TCP/IP와 소켓 통신 | NAT(Network Address Translation) | NAT(Network Address Translation) | `interview_questions.md:2577-2637` | `interviews.md:2577-2637` |
| 04. 네트워크와 웹 프로토콜 | TCP/IP와 소켓 통신 | IP 구조 | IP 구조 | `interview_questions.md:2638-2753` | `interviews.md:2638-2753` |
| 03. OS, 커널, 컴퓨터 구조 | 서버 하드웨어와 운영 환경 | 왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요? | 왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요? | `interview_questions.md:2754-2923` | `interviews.md:2754-2923` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 인덱스와 조회 성능 | B Tree & B+ Tree | B Tree & B+ Tree | `interview_questions.md:2924-3024` | `interviews.md:2924-3024` |
| 03. OS, 커널, 컴퓨터 구조 | 부팅과 init 시스템 | 서버 부팅 과정 | 서버 부팅 과정 | `interview_questions.md:3025-3148` | `interviews.md:3025-3148` |
| 03. OS, 커널, 컴퓨터 구조 | 부팅과 init 시스템 | systemd가 다른 시스템 데몬들을 실행하는 과정 | systemd가 다른 시스템 데몬들을 실행하는 과정 | `interview_questions.md:3149-3225` | `interviews.md:3149-3225` |
| 08. 분산 시스템과 아키텍처 | MSA와 결제 시스템 설계 | 결제 MSA에서 결제 PG사 기능 구현 | 결제 MSA에서 결제 PG사 기능 구현 | `interview_questions.md:3226-3247` | `interviews.md:3226-3247` |
| 09. Spring과 백엔드 프레임워크 | 트랜잭션과 데이터 접근 | 리파지토리 메서드에 `@Transactional` 사용 | 리파지토리 메서드에 `@Transactional` 사용 | `interview_questions.md:3248-3312` | `interviews.md:3248-3312` |
| 09. Spring과 백엔드 프레임워크 | Servlet/Tomcat 요청 처리 | Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정 | Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정 | `interview_questions.md:3313-3508` | `interviews.md:3313-3508` |
| 09. Spring과 백엔드 프레임워크 | AOP와 어노테이션 처리 | 어노테이션 프로세싱과 Spring의 내부 동작 원리 | 어노테이션 프로세싱과 Spring의 내부 동작 원리 | `interview_questions.md:3509-3561` | `interviews.md:3509-3561` |
| 01. 언어와 런타임 | Kotlin 언어 기능 | kotlin의 `when` 표현식에서의 exhaustiveness 체크 | kotlin의 `when` 표현식에서의 exhaustiveness 체크 | `interview_questions.md:3562-3673` | `interviews.md:3562-3673` |
| 09. Spring과 백엔드 프레임워크 | 트랜잭션과 데이터 접근 | hikari cp | hikari cp | `interview_questions.md:3674-3732` | `interviews.md:3674-3732` |
| 09. Spring과 백엔드 프레임워크 | AOP와 어노테이션 처리 | jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리 | jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리 | `interview_questions.md:3733-3734` | `interviews.md:3733-3734` |
| 02. 동시성, 비동기, I/O | Goroutine과 런타임 스케줄링 | Go의 Goroutine은 M 대 N 모델 멀티스레딩 | Go의 Goroutine은 M 대 N 모델 멀티스레딩 | `interview_questions.md:3735-3746` | `interviews.md:3735-3746` |
| 02. 동시성, 비동기, I/O | Coroutine과 협력형 실행 | Kotlin Coroutine | Kotlin Coroutine | `interview_questions.md:3747-3770` | `interviews.md:3747-3770` |
| 02. 동시성, 비동기, I/O | 스레드 생성과 스케줄링 | `pthread_create` 이후 커널에서 발생하는 작업들 | `pthread_create` 이후 커널에서 발생하는 작업들 | `interview_questions.md:3771-3811` | `interviews.md:3771-3811` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | 비동기 프로그래밍 | 비동기 프로그래밍 | `interview_questions.md:3812-3873` | `interviews.md:3812-3873` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | libuv, aio_*, io_uring | libuv, aio_*, io_uring | `interview_questions.md:3874-4103` | `interviews.md:3874-4103` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | 싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱 | 싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱 | `interview_questions.md:4104-4154` | `interviews.md:4104-4154` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | 블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리 | 블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리 | `interview_questions.md:4155-4469` | `interviews.md:4155-4469` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | 논블로킹 vs 비동기 차이 | 논블로킹 vs 비동기 차이 | `interview_questions.md:4470-4523` | `interviews.md:4470-4523` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | I/O 서브시스템 | I/O 서브시스템 | `interview_questions.md:4524-4582` | `interviews.md:4524-4582` |
| 09. Spring과 백엔드 프레임워크 | Spring Boot 실행 모델 | 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교 | 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교 | `interview_questions.md:4583-4753` | `interviews.md:4583-4753` |
| 09. Spring과 백엔드 프레임워크 | Spring Boot 실행 모델 | Spring boot에서 톰캣 실행 후 일어나는 일 | Spring boot에서 톰캣 실행 후 일어나는 일 | `interview_questions.md:4754-4873` | `interviews.md:4754-4873` |
| 09. Spring과 백엔드 프레임워크 | IoC와 DI | Spring Framework의 Dependency Injection (DI) | Spring Framework의 Dependency Injection (DI) | `interview_questions.md:4874-4979` | `interviews.md:4874-4979` |
| 09. Spring과 백엔드 프레임워크 | IoC와 DI | Inversion of Control (IoC) | Inversion of Control (IoC) | `interview_questions.md:4980-5097` | `interviews.md:4980-5097` |
| 09. Spring과 백엔드 프레임워크 | Bean 생명주기와 선택 | 적절한 Bean 선택 | 적절한 Bean 선택 | `interview_questions.md:5098-5284` | `interviews.md:5098-5284` |
| 03. OS, 커널, 컴퓨터 구조 | 스케줄링과 선점 | 선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이 | 선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이 | `interview_questions.md:5285-5286` | `interviews.md:5285-5286` |
| 01. 언어와 런타임 | 언어 런타임 비교 | VM, GC, 그리고 런타임 | VM, GC, 그리고 런타임 | `interview_questions.md:5287-5319` | `interviews.md:5287-5319` |
| 01. 언어와 런타임 | 언어 런타임 비교 | Golang의 Concurrent Mark and Sweep(CMS) | Golang의 Concurrent Mark and Sweep(CMS) | `interview_questions.md:5320-5419` | `interviews.md:5320-5419` |
| 01. 언어와 런타임 | 언어 런타임 비교 | Apache + mod_php와 Nginx + php-fpm의 차이 | Apache + mod_php와 Nginx + php-fpm의 차이 | `interview_questions.md:5420-5544` | `interviews.md:5420-5544` |
| 01. 언어와 런타임 | JVM 실행 모델 | JVM, PHP runtime, Go runtime | JVM, PHP runtime, Go runtime | `interview_questions.md:5545-5614` | `interviews.md:5545-5614` |
| 03. OS, 커널, 컴퓨터 구조 | 프로세스 생성과 실행 | fork | fork | `interview_questions.md:5615-5634` | `interviews.md:5615-5634` |
| 03. OS, 커널, 컴퓨터 구조 | 프로세스 생성과 실행 | exec | exec | `interview_questions.md:5635-5646` | `interviews.md:5635-5646` |
| 03. OS, 커널, 컴퓨터 구조 | 프로세스 생성과 실행 | java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지 | java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지 | `interview_questions.md:5647-6498` | `interviews.md:5647-6498` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | 네트워크 통신 시 epoll 동작 방식 | 네트워크 통신 시 epoll 동작 방식 | `interview_questions.md:6499-6611` | `interviews.md:6499-6611` |
| 04. 네트워크와 웹 프로토콜 | TCP/IP와 소켓 통신 | 클라이언트와 서버 간의 소켓 통신 과정 | 클라이언트와 서버 간의 소켓 통신 과정 | `interview_questions.md:6612-6819` | `interviews.md:6612-6819` |
| 05. 보안과 암호학 | TLS/HTTPS 핸드셰이크 | HTTPS 통신에서 TLS 핸드셰이크 | HTTPS 통신에서 TLS 핸드셰이크 | `interview_questions.md:6820-7075` |  |
| 05. 보안과 암호학 | 키 교환과 전방 비밀성 | DHE(Diffie-Hellman)/ECDHE와 TLS | DHE(Diffie-Hellman)/ECDHE와 TLS | `interview_questions.md:7076-7177` | `interviews.md:7024-7125` |
| 09. Spring과 백엔드 프레임워크 | Servlet/Tomcat 요청 처리 | NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우 | NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우 | `interview_questions.md:7178-7293` | `interviews.md:7126-7241` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | 소켓, epoll, nginx, 과다한 요청 | 소켓, epoll, nginx, 과다한 요청 | `interview_questions.md:7294-7330` | `interviews.md:7242-7278` |
| 09. Spring과 백엔드 프레임워크 | Servlet/Tomcat 요청 처리 | nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유 | nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유 | `interview_questions.md:7331-7414` | `interviews.md:7279-7362` |
| 05. 보안과 암호학 | TLS/HTTPS 핸드셰이크 | HTTPS 통신 과정 | HTTPS 통신 과정 | `interview_questions.md:7415-7418` | `interviews.md:7363-7366` |
| 03. OS, 커널, 컴퓨터 구조 | 서버 하드웨어와 운영 환경 | 노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요? | 노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요? | `interview_questions.md:7419-7420` | `interviews.md:7367-7368` |
| Source Context And Question Bank | 질문 은행과 면접 시나리오 | 프로젝트에서 사용한 오픈소스들에 대해서 내부적으로 어떻게 돌아가고 있는지 알고 있어야 | 프로젝트에서 사용한 오픈소스들에 대해서 내부적으로 어떻게 돌아가고 있는지 알고 있어야 | `interview_questions.md:7421-7424` | `interviews.md:7369-7372` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색 엔진과 샤딩 | Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명 | Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명 | `interview_questions.md:7425-7435` | `interviews.md:7373-7383` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색 엔진과 샤딩 | ElasticSearch에서 OOM(Out of Memory) 문제 | ElasticSearch에서 OOM(Out of Memory) 문제 | `interview_questions.md:7436-7500` | `interviews.md:7384-7448` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색 엔진과 샤딩 | [ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers) | [ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers) | `interview_questions.md:7501-7537` | `interviews.md:7449-7485` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색 엔진과 샤딩 | Elasticsearch의 샤드와 레플리카 샤드 | Elasticsearch의 샤드와 레플리카 샤드 | `interview_questions.md:7538-7605` | `interviews.md:7486-7553` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색 엔진과 샤딩 | ElasticSearch 쿼리 과정 예시 | ElasticSearch 쿼리 과정 예시 | `interview_questions.md:7606-7755` | `interviews.md:7554-7703` |
| 07. 메시징과 이벤트 기반 구조 | Broker 비교와 선택 | rabbit mq, kafka 등 차이 | rabbit mq, kafka 등 차이 | `interview_questions.md:7756-7757` | `interviews.md:7704-7705` |
| 07. 메시징과 이벤트 기반 구조 | 메시징 프로토콜 | [표준 메세징 프로토콜 정리 (AMQP, STOMP, MQTT)](https://velog.io/@holicme7/%ED%91%9C%EC%A4%80-%EB%A9%94%EC%84%B8%EC%A7%95-%ED%94%84%EB%A1%9C%ED%86%A0%EC%BD%9C-%EC%A0%95%EB%A6%AC-AMQP-STOMP-MQTT) | [표준 메세징 프로토콜 정리 (AMQP, STOMP, MQTT)](https://velog.io/@holicme7/%ED%91%9C%EC%A4%80-%EB%A9%94%EC%84%B8%EC%A7%95-%ED%94%84%EB%A1%9C%ED%86%A0%EC%BD%9C-%EC%A0%95%EB%A6%AC-AMQP-STOMP-MQTT) | `interview_questions.md:7758-7759` | `interviews.md:7706-7707` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 트랜잭션, 락, 격리 | 데이터베이스 락(Lock)과 격리 수준(Isolation Level) | 데이터베이스 락(Lock)과 격리 수준(Isolation Level) | `interview_questions.md:7760-7805` | `interviews.md:7708-7753` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 트랜잭션, 락, 격리 | MVCC와 스냅숏 격리 | MVCC와 스냅숏 격리 | `interview_questions.md:7806-7811` | `interviews.md:7754-7759` |
| 04. 네트워크와 웹 프로토콜 | TCP/IP와 소켓 통신 | TCP 컨제스천 컨트롤 | TCP 컨제스천 컨트롤 | `interview_questions.md:7812-7815` | `interviews.md:7760-7763` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 검색/NoSQL 저장소 | 카우치베이스 | 카우치베이스 | `interview_questions.md:7816-7975` | `interviews.md:7764-7923` |
| 05. 보안과 암호학 | TLS/HTTPS 핸드셰이크 | pkcs 11 멀티 프로세스 curl 에러 | pkcs 11 멀티 프로세스 curl 에러 | `interview_questions.md:7976-7977` | `interviews.md:7924-7925` |
| 10. 문제 해결, 코드 품질, 운영 실천 | 알고리즘과 문제 해결 | 정규표현식 | 정규표현식 | `interview_questions.md:7978-7979` | `interviews.md:7926-7927` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 트랜잭션, 락, 격리 | 2단계커밋과 2단계 잠금 | 2단계커밋과 2단계 잠금 | `interview_questions.md:7980-8056` | `interviews.md:7928-8004` |
| 04. 네트워크와 웹 프로토콜 | Proxy와 L7 진입 경로 | proxy와 reverse proxy | proxy와 reverse proxy | `interview_questions.md:8057-8109` | `interviews.md:8005-8057` |
| 10. 문제 해결, 코드 품질, 운영 실천 | 설계 원칙과 패턴 | 프록시 패턴 | 프록시 패턴 | `interview_questions.md:8110-8549` | `interviews.md:8058-8497` |
| 05. 보안과 암호학 | 키 교환과 전방 비밀성 | 전방 비밀성 | 전방 비밀성 | `interview_questions.md:8550-8568` | `interviews.md:8498-8516` |
| 07. 메시징과 이벤트 기반 구조 | Broker 비교와 선택 | RabbitMQ와 Kafka | RabbitMQ와 Kafka | `interview_questions.md:8569-8817` | `interviews.md:8517-8765` |
| 07. 메시징과 이벤트 기반 구조 | Consumer와 클라이언트 구현 | RabbitMQ가 Node.js 앱으로 메시지를 전달하는 과정 | RabbitMQ가 Node.js 앱으로 메시지를 전달하는 과정 | `interview_questions.md:8818-8868` | `interviews.md:8766-8816` |
| 07. 메시징과 이벤트 기반 구조 | 메시징 프로토콜 | `amqplib`의 동작 원리 | `amqplib`의 동작 원리 | `interview_questions.md:8869-8896` | `interviews.md:8817-8844` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | Node.js의 싱글 스레드와 libuv의 멀티 스레딩 | Node.js의 싱글 스레드와 libuv의 멀티 스레딩 | `interview_questions.md:8897-9127` | `interviews.md:8845-9075` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | epoll 상세 | epoll 상세 | `interview_questions.md:9128-9213` | `interviews.md:9076-9161` |
| 03. OS, 커널, 컴퓨터 구조 | 프로세스 생성과 실행 | `&`와 background process | `&`와 background process | `interview_questions.md:9214-9215` | `interviews.md:9162-9163` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | DB 접근과 저장소 운영 | PDO가 데이터베이스에서 데이터 가져오는 원리 | PDO가 데이터베이스에서 데이터 가져오는 원리 | `interview_questions.md:9216-9329` | `interviews.md:9164-9277` |
| 09. Spring과 백엔드 프레임워크 | 트랜잭션과 데이터 접근 | Spring JDBC에서도 데이터를 가져오는 방식 | Spring JDBC에서도 데이터를 가져오는 방식 | `interview_questions.md:9330-9422` | `interviews.md:9278-9370` |
| 10. 문제 해결, 코드 품질, 운영 실천 | 테스트와 대역 객체 | 테스트 더블 | 테스트 더블 | `interview_questions.md:9423-9460` | `interviews.md:9371-9408` |
| 10. 문제 해결, 코드 품질, 운영 실천 | 테스트와 대역 객체 | java, php에서 mock 생성 원리 | java, php에서 mock 생성 원리 | `interview_questions.md:9461-9751` | `interviews.md:9409-9699` |
| 09. Spring과 백엔드 프레임워크 | Bean 생명주기와 선택 | `@Bean` vs. `@Component` | `@Bean` vs. `@Component` | `interview_questions.md:9752-9936` | `interviews.md:9700-9884` |
| 09. Spring과 백엔드 프레임워크 | Servlet/Tomcat 요청 처리 | Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정 | Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정 | `interview_questions.md:9937-10147` | `interviews.md:9885-10095` |
| 01. 언어와 런타임 | Class loading과 실행 준비 | 부모 위임 모델 | 부모 위임 모델 | `interview_questions.md:10148-10266` | `interviews.md:10096-10214` |
| 09. Spring과 백엔드 프레임워크 | Spring Boot 실행 모델 | Spring의 동시성 및 Thread Management 정리 | Spring의 동시성 및 Thread Management 정리 | `interview_questions.md:10267-10595` | `interviews.md:10215-10543` |
| 09. Spring과 백엔드 프레임워크 | Spring HTTP 클라이언트와 Reactive | RestTemplate vs WebClient 차이 | RestTemplate vs WebClient 차이 | `interview_questions.md:10596-10696` | `interviews.md:10544-10644` |
| 09. Spring과 백엔드 프레임워크 | Spring HTTP 클라이언트와 Reactive | Spring WebFlux의 이벤트 루프(Event Loop) | Spring WebFlux의 이벤트 루프(Event Loop) | `interview_questions.md:10697-10731` | `interviews.md:10645-10679` |
| 09. Spring과 백엔드 프레임워크 | Spring HTTP 클라이언트와 Reactive | Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계 | Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계 | `interview_questions.md:10732-10854` | `interviews.md:10680-10802` |
| 02. 동시성, 비동기, I/O | Event loop와 네트워크 런타임 | Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현 | Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현 | `interview_questions.md:10855-10992` | `interviews.md:10803-10940` |
| Source Context And Question Bank | 질문 은행과 면접 시나리오 | 시나리오1 | 시나리오1 | `interview_questions.md:10993-11231` | `interviews.md:10941-11179` |
| Source Context And Question Bank | 질문 은행과 면접 시나리오 | 시나리오 2 | 시나리오 2 | `interview_questions.md:11232-11371` | `interviews.md:11180-11319` |
| Source Context And Question Bank | 원재료 메타데이터 | 파일 앞부분과 생성 맥락 | [source front matter] | `interview_questions2.md:1-101` | `interviews2.md:1-101` |
| 02. 동시성, 비동기, I/O | 공유 상태와 경쟁 조건 | 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지 | 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지 | `interview_questions2.md:102-249` | `interviews2.md:102-249` |
| 02. 동시성, 비동기, I/O | Coroutine과 협력형 실행 | 코루틴을 Cooperative Multitasking 이라고 하는 이유? | 코루틴을 Cooperative Multitasking 이라고 하는 이유? | `interview_questions2.md:250-356` | `interviews2.md:250-356` |
| 02. 동시성, 비동기, I/O | Coroutine과 협력형 실행 | 코루틴 제어의 양보 | 코루틴 제어의 양보 | `interview_questions2.md:357-408` | `interviews2.md:357-408` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | Kotlin의 async HTTP 요청과 코루틴 비교 | Kotlin의 async HTTP 요청과 코루틴 비교 | `interview_questions2.md:409-438` | `interviews2.md:409-438` |
| 02. 동시성, 비동기, I/O | Blocking, non-blocking, async 구분 | HTTP 요청을 Async로 처리했을 때의 이점 | HTTP 요청을 Async로 처리했을 때의 이점 | `interview_questions2.md:439-541` | `interviews2.md:439-541` |
| 02. 동시성, 비동기, I/O | OS I/O multiplexing | 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우 | 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우 | `interview_questions2.md:542-701` | `interviews2.md:542-701` |
| 09. Spring과 백엔드 프레임워크 | Spring Boot 실행 모델 | `java -jar SpringBootApp.jar` 명령어의 실행 과정 | `java -jar SpringBootApp.jar` 명령어의 실행 과정 | `interview_questions2.md:702-1119` | `interviews2.md:702-1119` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 대용량 테이블 운영 | 수억건의 데이터가 존재하는 테이블 처리 방법 | 수억건의 데이터가 존재하는 테이블 처리 방법 | `interview_questions2.md:1120-1408` | `interviews2.md:1120-1408` |
| 06. 데이터베이스, 저장소, 검색/NoSQL | 인덱스와 조회 성능 | 3억건의 데이터의 B+Tree 깊이 계산 | 3억건의 데이터의 B+Tree 깊이 계산 | `interview_questions2.md:1409-1498` | `interviews2.md:1409-1498` |
| 01. 언어와 런타임 | Kotlin 언어 기능 | infix, inline, noinline, crossinline | infix, inline, noinline, crossinline | `interview_questions2.md:1499-1629` | `interviews2.md:1499-1629` |
| 09. Spring과 백엔드 프레임워크 | Bean 생명주기와 선택 | 스프링 부트 시작과 빈 라이프 사이클 | 스프링 부트 시작과 빈 라이프 사이클 | `interview_questions2.md:1630-1742` | `interviews2.md:1630-1742` |
| 01. 언어와 런타임 | 함수형 프로그래밍 모델 | 함수형 | 함수형 | `interview_questions2.md:1743-2120` | `interviews2.md:1743-2120` |
| 04. 네트워크와 웹 프로토콜 | HTTP/gRPC와 스트리밍 | HTTP 요청이 일반적인 TCP 연결 방식과 Keep-Alive TCP 연결 방식 이루어지는 경우 | HTTP 요청이 일반적인 TCP 연결 방식과 Keep-Alive TCP 연결 방식 이루어지는 경우 | `interview_questions2.md:2121-2296` | `interviews2.md:2121-2296` |
| Source Context And Question Bank | 원재료 메타데이터 | 파일 앞부분과 생성 맥락 | [source front matter] | `interview_questions3.md:1-91` |  |
| 01. 언어와 런타임 | JVM 실행 모델 | 1. Java 관련 질문 | 1. Java 관련 질문 | `interview_questions3.md:92-129` |  |
| 09. Spring과 백엔드 프레임워크 | Spring Boot 실행 모델 | 2. Spring 관련 질문 | 2. Spring 관련 질문 | `interview_questions3.md:130-164` |  |
| 06. 데이터베이스, 저장소, 검색/NoSQL | DB 접근과 저장소 운영 | 3. MySQL 관련 질문 | 3. MySQL 관련 질문 | `interview_questions3.md:165-188` |  |
| 02. 동시성, 비동기, I/O | Event loop와 네트워크 런타임 | 4. Netty 관련 질문 | 4. Netty 관련 질문 | `interview_questions3.md:189-212` |  |
| 08. 분산 시스템과 아키텍처 | 가용성, 토폴로지, 확장성 | 5. 종합 아키텍처/적용 질문 | 5. 종합 아키텍처/적용 질문 | `interview_questions3.md:213-236` |  |
| 02. 동시성, 비동기, I/O | 공유 상태와 경쟁 조건 | 자바 쓰레드 동기화 | 자바 쓰레드 동기화 | `interview_questions3.md:237-393` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 자바 쓰레드, 상태, 그리고 모니터 등 | 자바 쓰레드, 상태, 그리고 모니터 등 | `interview_questions3.md:394-395` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 1. 자바에서의 쓰레드와 모니터 개념 | 1. 자바에서의 쓰레드와 모니터 개념 | `interview_questions3.md:396-405` |  |
| 02. 동시성, 비동기, I/O | 스레드 상태와 스케줄링 | 2. 쓰레드 상태(States) 개념 | 2. 쓰레드 상태(States) 개념 | `interview_questions3.md:406-420` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 3. synchronized 블록과 모니터 잠금 | 3. synchronized 블록과 모니터 잠금 | `interview_questions3.md:421-440` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 4. wait()와 notify() 메서드의 메커니즘 | 4. wait()와 notify() 메서드의 메커니즘 | `interview_questions3.md:441-462` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 5. wait set의 작동 원리 | 5. wait set의 작동 원리 | `interview_questions3.md:463-472` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 6. 쓰레드 상태: BLOCKED vs WAITING | 6. 쓰레드 상태: BLOCKED vs WAITING | `interview_questions3.md:473-490` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 7. notify가 없으면 wait한 스레드는 어떻게 되는가? | 7. notify가 없으면 wait한 스레드는 어떻게 되는가? | `interview_questions3.md:491-498` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 8. JVM 레벨에서의 synchronized 작동 방식 | 8. JVM 레벨에서의 synchronized 작동 방식 | `interview_questions3.md:499-515` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 9. 실제 동작 시나리오 예시 | 9. 실제 동작 시나리오 예시 | `interview_questions3.md:516-534` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 10. 핵심 요약 및 결론 | 10. 핵심 요약 및 결론 | `interview_questions3.md:535-550` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시 | 자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시 | `interview_questions3.md:551-730` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 쓰레드 wait와 notify 예제 | 쓰레드 wait와 notify 예제 | `interview_questions3.md:731-932` |  |
| 02. 동시성, 비동기, I/O | Java monitor와 wait/notify | 7. 결론 정리 | 7. 결론 정리 | `interview_questions3.md:933-947` |  |
| Source Context And Question Bank | 질문 은행과 면접 시나리오 | Interview | Interview | `interview_questions4.md:1-222` |  |
| Source Context And Question Bank | 질문 은행과 면접 시나리오 | 심층 기술 면접 시나리오 | 심층 기술 면접 시나리오 | `interview_s4.md:1-241` |  |
| 05. 보안과 암호학 | TLS/HTTPS 핸드셰이크 | HTTPS 통신에서 TLS 핸드셰이크 | HTTPS 통신에서 TLS 핸드셰이크 | `interviews.md:6820-7023` |  |
