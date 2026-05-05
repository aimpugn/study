# 인터뷰 질문과 답변 3

- [인터뷰 질문과 답변 3](#인터뷰-질문과-답변-3)
    - [1. Java 관련 질문](#1-java-관련-질문)
        - [질문 1. 자바에서 메모리는 어떻게 구성되어 있으며, 각 영역은 어떤 목적을 가지나요?](#질문-1-자바에서-메모리는-어떻게-구성되어-있으며-각-영역은-어떤-목적을-가지나요)
            - [**답변**](#답변)
        - [질문 2. 자바 가비지 컬렉션 알고리즘에는 어떤 종류가 있으며, 어떤 특징을 갖나요?](#질문-2-자바-가비지-컬렉션-알고리즘에는-어떤-종류가-있으며-어떤-특징을-갖나요)
            - [**답변**](#답변-1)
        - [질문 3. Java 동시성(멀티쓰레드) 프로그래밍에서 주의해야 할 점은 무엇인가요?](#질문-3-java-동시성멀티쓰레드-프로그래밍에서-주의해야-할-점은-무엇인가요)
            - [**답변**](#답변-2)
    - [2. Spring 관련 질문](#2-spring-관련-질문)
        - [질문 4. Spring IoC(Inversion of Control)와 DI(Dependency Injection)은 어떤 원리로 동작하나요?](#질문-4-spring-iocinversion-of-control와-didependency-injection은-어떤-원리로-동작하나요)
            - [**답변**](#답변-3)
        - [질문 5. Spring AOP(Aspect Oriented Programming)에서는 어떤 방식으로 공통 로직을 분리하나요?](#질문-5-spring-aopaspect-oriented-programming에서는-어떤-방식으로-공통-로직을-분리하나요)
            - [**답변**](#답변-4)
        - [질문 6. @Transactional 애노테이션이 어떻게 트랜잭션을 관리하나요?](#질문-6-transactional-애노테이션이-어떻게-트랜잭션을-관리하나요)
            - [**답변**](#답변-5)
    - [3. MySQL 관련 질문](#3-mysql-관련-질문)
        - [질문 7. MySQL에서 인덱스는 어떤 구조이며, 쿼리 성능을 어떻게 높일 수 있나요?](#질문-7-mysql에서-인덱스는-어떤-구조이며-쿼리-성능을-어떻게-높일-수-있나요)
            - [**답변**](#답변-6)
        - [질문 8. MySQL 트랜잭션 격리 수준과 InnoDB의 MVCC 방식은 어떻게 동작하나요?](#질문-8-mysql-트랜잭션-격리-수준과-innodb의-mvcc-방식은-어떻게-동작하나요)
            - [**답변**](#답변-7)
    - [4. Netty 관련 질문](#4-netty-관련-질문)
        - [질문 9. Netty는 어떤 원리로 고성능 네트워크를 구현하나요?](#질문-9-netty는-어떤-원리로-고성능-네트워크를-구현하나요)
            - [**답변**](#답변-8)
        - [질문 10. Netty 파이프라인에서 데이터가 인바운드/아웃바운드로 흐를 때, 예외가 발생하면 어떻게 처리되나요?](#질문-10-netty-파이프라인에서-데이터가-인바운드아웃바운드로-흐를-때-예외가-발생하면-어떻게-처리되나요)
            - [**답변**](#답변-9)
    - [5. 종합 아키텍처/적용 질문](#5-종합-아키텍처적용-질문)
        - [질문 11. Spring + Netty + MySQL로 고부하(High Traffic) 시스템을 만든다면 어떤 아키텍처 구성을 권장하시나요?](#질문-11-spring--netty--mysql로-고부하high-traffic-시스템을-만든다면-어떤-아키텍처-구성을-권장하시나요)
            - [**답변**](#답변-10)
        - [질문 12. Java 스레드풀과 Netty 이벤트 루프를 함께 사용할 때 주의점은 무엇인가요?](#질문-12-java-스레드풀과-netty-이벤트-루프를-함께-사용할-때-주의점은-무엇인가요)
            - [**답변**](#답변-11)
    - [자바 쓰레드 동기화](#자바-쓰레드-동기화)
        - [1. 자바에서의 쓰레드와 모니터 개념](#1-자바에서의-쓰레드와-모니터-개념)
        - [2. 쓰레드 상태(States) 개념](#2-쓰레드-상태states-개념)
        - [3. synchronized 블록과 모니터 잠금](#3-synchronized-블록과-모니터-잠금)
            - [3.1 진입 과정](#31-진입-과정)
            - [3.2 탈출 과정](#32-탈출-과정)
        - [4. wait()와 notify() 메서드의 메커니즘](#4-wait와-notify-메서드의-메커니즘)
            - [4.1 wait()가 하는 일](#41-wait가-하는-일)
            - [4.2 notify()와 notifyAll()](#42-notify와-notifyall)
        - [5. wait set의 작동 원리](#5-wait-set의-작동-원리)
        - [6. 쓰레드 상태: BLOCKED vs WAITING](#6-쓰레드-상태-blocked-vs-waiting)
            - [6.1 BLOCKED 상태](#61-blocked-상태)
            - [6.2 WAITING 상태](#62-waiting-상태)
        - [7. notify가 없으면 wait한 스레드는 어떻게 되는가?](#7-notify가-없으면-wait한-스레드는-어떻게-되는가)
        - [8. JVM 레벨에서의 synchronized 작동 방식](#8-jvm-레벨에서의-synchronized-작동-방식)
            - [8.1 모니터 엔터/익시트(enter/exit)](#81-모니터-엔터익시트enterexit)
            - [8.2 경량화 기법(Biased Locking, Lightweight Locking)](#82-경량화-기법biased-locking-lightweight-locking)
        - [9. 실제 동작 시나리오 예시](#9-실제-동작-시나리오-예시)
        - [10. 핵심 요약 및 결론](#10-핵심-요약-및-결론)
- [자바 쓰레드, 상태, 그리고 모니터 등](#자바-쓰레드-상태-그리고-모니터-등)
    - [1. 자바에서의 쓰레드와 모니터 개념](#1-자바에서의-쓰레드와-모니터-개념-1)
    - [2. 쓰레드 상태(States) 개념](#2-쓰레드-상태states-개념-1)
    - [3. synchronized 블록과 모니터 잠금](#3-synchronized-블록과-모니터-잠금-1)
        - [3.1 진입 과정](#31-진입-과정-1)
        - [3.2 탈출 과정](#32-탈출-과정-1)
    - [4. wait()와 notify() 메서드의 메커니즘](#4-wait와-notify-메서드의-메커니즘-1)
        - [4.1 wait()가 하는 일](#41-wait가-하는-일-1)
        - [4.2 notify()와 notifyAll()](#42-notify와-notifyall-1)
    - [5. wait set의 작동 원리](#5-wait-set의-작동-원리-1)
    - [6. 쓰레드 상태: BLOCKED vs WAITING](#6-쓰레드-상태-blocked-vs-waiting-1)
        - [6.1 BLOCKED 상태](#61-blocked-상태-1)
        - [6.2 WAITING 상태](#62-waiting-상태-1)
    - [7. notify가 없으면 wait한 스레드는 어떻게 되는가?](#7-notify가-없으면-wait한-스레드는-어떻게-되는가-1)
    - [8. JVM 레벨에서의 synchronized 작동 방식](#8-jvm-레벨에서의-synchronized-작동-방식-1)
        - [8.1 모니터 엔터/익시트(enter/exit)](#81-모니터-엔터익시트enterexit-1)
        - [8.2 경량화 기법(Biased Locking, Lightweight Locking)](#82-경량화-기법biased-locking-lightweight-locking-1)
    - [9. 실제 동작 시나리오 예시](#9-실제-동작-시나리오-예시-1)
    - [10. 핵심 요약 및 결론](#10-핵심-요약-및-결론-1)
    - [**자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시**](#자바에서-waitnotify-패턴을-조금-더-현실감-있게-적용한-프로덕션-코드-예시)
        - [1. 핵심 개념 요약](#1-핵심-개념-요약)
        - [2. 예시 코드: 작업 준비(Producer)와 작업 처리(Consumer)](#2-예시-코드-작업-준비producer와-작업-처리consumer)
            - [2.1 코드 흐름 설명](#21-코드-흐름-설명)
        - [3. notify는 누가 언제 실행하고, 그 결과는 무엇인가?](#3-notify는-누가-언제-실행하고-그-결과는-무엇인가)
        - [4. "모든 스레드가 wait만 하고 notify를 호출하는 스레드가 없다면?" 발생하는 문제](#4-모든-스레드가-wait만-하고-notify를-호출하는-스레드가-없다면-발생하는-문제)
        - [5. synchronized(락)을 사용하면 메모리 베리어가 적용되는가?](#5-synchronized락을-사용하면-메모리-베리어가-적용되는가)
        - [6. 마지막 정리](#6-마지막-정리)
    - [쓰레드 wait와 notify 예제](#쓰레드-wait와-notify-예제)
        - [1. 왜 `notify()`가 필요하고, 누가 언제 호출해야 하는가?](#1-왜-notify가-필요하고-누가-언제-호출해야-하는가)
            - [요약](#요약)
            - [예시 흐름](#예시-흐름)
        - [2. 현실 세계 프로덕션 코드 예시: “TaskQueue” 시나리오](#2-현실-세계-프로덕션-코드-예시-taskqueue-시나리오)
            - [2.1 TaskQueue 구현 (프로덕션 코드)](#21-taskqueue-구현-프로덕션-코드)
            - [2.2 실제 doSomething() 메서드 예시](#22-실제-dosomething-메서드-예시)
        - [3. 테스트 코드 (TDD 접근)](#3-테스트-코드-tdd-접근)
        - [4. notify는 명시적으로 호출해야 하며, 누가 언제 실행할까?](#4-notify는-명시적으로-호출해야-하며-누가-언제-실행할까)
        - [5. notify()를 호출한 스레드는 어떻게 되나요?](#5-notify를-호출한-스레드는-어떻게-되나요)
        - [6. `synchronized(lock)` 사용 시 메모리 장벽(Memory Barrier) 적용 여부](#6-synchronizedlock-사용-시-메모리-장벽memory-barrier-적용-여부)
- [7. 결론 정리](#7-결론-정리)

## 1. Java 관련 질문

### 질문 1. 자바에서 메모리는 어떻게 구성되어 있으며, 각 영역은 어떤 목적을 가지나요?

#### **답변**

1. 자바 애플리케이션이 실행될 때, **JVM(Java Virtual Machine)**은 크게 여러 메모리 영역을 관리합니다. 그중 대표적으로 **힙(Heap) 영역**, **스택(Stack) 영역**, 그리고 **메타스페이스(Metaspace)** 영역이 존재합니다.
2. **힙(Heap) 영역**은 객체와 배열이 동적으로 할당되는 곳입니다. 자바의 모든 객체(예: new 키워드로 생성된 인스턴스)가 대부분 힙에 위치합니다. 가비지 컬렉터(GC)가 이 영역을 주기적으로 검사하여, 더 이상 참조되지 않는 객체를 해제합니다.
3. **스택(Stack) 영역**은 각 쓰레드마다 분리되어 있습니다. 한 쓰레드 안에서 메서드가 호출될 때마다 스택 프레임이 쌓이고, 지역 변수와 메서드 실행 컨텍스트가 이 스택 프레임에 보관됩니다. 메서드가 종료되면 해당 스택 프레임이 사라지면서 그 안의 지역 변수가 해제됩니다.
4. **메타스페이스(Metaspace)**는 자바 8부터 도입된 영역으로, 이전에 PermGen으로 불렸던 부분을 대체합니다. 클래스 메타데이터, 리플렉션 정보, 동적 프록시 정보 등을 저장합니다. 이 영역은 OS 메모리를 동적으로 사용할 수 있으므로, PermGen과 달리 OutOfMemoryError 위험이 어느 정도 줄었습니다.
5. 이처럼 자바 메모리가 여러 영역으로 나뉘어 있는 이유는, **가비지 컬렉션 효율성**과 **프로그램 구조화**에 있습니다. 힙은 객체를 통합적으로 관리하고, 스택은 메서드 호출 구조와 밀접하게 연관되며, 메타스페이스는 클래스 정보를 별도로 관리함으로써 JVM이 메모리를 더욱 체계적으로 운용할 수 있게 해줍니다.

---

### 질문 2. 자바 가비지 컬렉션 알고리즘에는 어떤 종류가 있으며, 어떤 특징을 갖나요?

#### **답변**

1. 자바는 객체 생명주기를 자동으로 관리하기 위해 가비지 컬렉션(Garbage Collection)을 사용합니다. 가비지 컬렉션 알고리즘이란, **더 이상 사용되지 않는(참조되지 않는) 객체를 탐색**하여 힙에서 제거하는 로직을 의미합니다.
2. Java 8까지 자주 사용되던 알고리즘으로는 **Serial GC**, **Parallel GC**, **CMS(Concurrent Mark-Sweep)** 등이 있습니다. Serial GC는 단일 쓰레드로 동작하여 간단하지만, 멀티코어 환경에서는 확장성이 떨어집니다. Parallel GC는 여러 쓰레드를 통해 동시에 객체를 수집하기 때문에 처리량(Throughput)이 높아집니다. CMS는 애플리케이션 스레드와 GC 스레드를 부분적으로 병행하여(Concurrent) 멈추는 시간(Stop-the-world pause)을 줄이려 합니다.
3. Java 9 이상에서는 **G1 GC**가 기본이 되었습니다. G1(Garbage First) GC는 힙을 여러 영역(Region)으로 나누고, 객체 참조도를 추적하여 우선적으로 수거해야 하는 영역부터 처리(garbage first)합니다. 이 방식은 대규모 힙에서 멈춤 시간을 줄이는 데 효과적입니다.
4. 최근에는 **ZGC**나 **Shenandoah** 같은 더 최신의 저지연 GC도 등장했습니다. 이들은 매우 큰 힙에서도 짧은 멈춤 시간(Stop-the-world)을 추구합니다. 예를 들어 ZGC는 큰 힙(수십 GB~TB급)에서도 최대 몇 ms 정도의 멈춤 시간을 목표로 합니다.
5. 이러한 다양한 GC 알고리즘이 존재하는 이유는, **애플리케이션 특성**(예: 대규모 서비스, 짧은 응답 시간 요구, CPU 자원 크기)에 따라 GC가 성능에 큰 영향을 미치기 때문입니다. 실제 서비스에서는 GC 로그를 모니터링하여, 적절한 GC 모드를 선택하고 필요하면 JVM 옵션을 조정해 튜닝합니다.

---

### 질문 3. Java 동시성(멀티쓰레드) 프로그래밍에서 주의해야 할 점은 무엇인가요?

#### **답변**

1. 멀티쓰레드 환경에서는 **공유 자원(shared resource)**에 여러 쓰레드가 동시에 접근할 수 있기 때문에, 동기화(synchronization) 문제가 핵심입니다. 즉, 서로 다른 쓰레드가 같은 데이터를 변경하는 순간, 올바르지 않은 결과가 발생할 수 있습니다(예: Race Condition).
2. 이를 방지하기 위해, 자바는 `synchronized` 키워드나 `ReentrantLock` 클래스를 통해 한 번에 하나의 쓰레드만 임계영역(critical section)에 들어가도록 제한합니다.
3. 또한 **volatile** 키워드는 변수의 가시성(visibility)을 보장하기 위한 것입니다. volatile로 선언된 변수에 쓰기가 일어나면, 모든 쓰레드가 즉시 최신 값을 확인하도록 보장합니다.
4. 자바 메모리 모델(JMM)에 따르면, 쓰레드 간에 값이 즉시 전파되지 않을 수도 있고(캐시 혹은 레지스터 최적화 등), 리오더(reordering)가 발생할 수도 있습니다. synchronized나 volatile을 사용하면 happens-before 관계를 형성하여 이러한 문제를 해결할 수 있습니다.
5. **동시성 프로그래밍**은 CPU 코어 수가 늘어난 현대 환경에서 필수적이지만, 설계가 잘못되면 교착상태(Deadlock)나 리소스 경합(Contention)으로 인해 성능이 급격히 하락할 수 있습니다. 따라서 멀티쓰레드를 사용할 때는 임계구역 최소화, 락 분리, 락 프리 알고리즘(Atomic 클래스) 고려 등 세밀한 접근이 필요합니다.

---

## 2. Spring 관련 질문

### 질문 4. Spring IoC(Inversion of Control)와 DI(Dependency Injection)은 어떤 원리로 동작하나요?

#### **답변**

1. **IoC(Inversion of Control)**란, 객체의 생성과 의존성 설정을 개발자가 직접 하지 않고, **Spring 컨테이너**가 대신 맡는 구조를 말합니다. 전통적 방법에서는 각 객체가 스스로 필요한 다른 객체를 생성하거나 찾아야 하지만, IoC를 사용하면 컨테이너가 Bean을 생성하고 연결(주입)해줍니다.
2. **DI(Dependency Injection)**는 "의존성 주입"이란 의미로, A 객체가 B 객체를 필요로 할 때, A 스스로 B를 생성하지 않고, 컨테이너가 A를 만들 때 B를 주입해주는 방식입니다. 예를 들어, `@Autowired`, `@Inject`, `@Resource` 어노테이션 등을 통해 다른 Bean이 자동으로 주입됩니다.
3. 스프링은 **Bean Definition**(XML 설정, 자바 Config, Component Scan 등)을 바탕으로 어떤 클래스를 어떤 빈(Bean)으로 등록할지 파악합니다. 그리고 빈을 생성할 때, 생성자/필드/setter를 통해 필요한 다른 빈들을 연결합니다.
4. 결국 IoC/DI를 통해 **“의존성 관리”**가 단순해지고, 객체 간 결합도가 낮아집니다. 테스트 코드 작성(단위 테스트)나 교체 가능성이 높아지며, 애플리케이션 구조가 유연해집니다.

---

### 질문 5. Spring AOP(Aspect Oriented Programming)에서는 어떤 방식으로 공통 로직을 분리하나요?

#### **답변**

1. AOP(Aspect Oriented Programming)는 **여러 클래스나 메서드에 공통적으로 필요한 기능**(예: 로깅, 트랜잭션, 보안 검사)을 분리하는 기법입니다. 예를 들어, 모든 서비스 메서드 호출 전후에 로깅을 남기고 싶을 때 AOP가 유용합니다.
2. 스프링 AOP는 주로 **프록시(Proxy) 방식**으로 구현됩니다. 즉, 실제 객체(타겟 객체)를 감싸는 **프록시 객체**를 만들어, 메서드가 호출될 때 프록시가 먼저 관여하여, 부가기능(Advice)을 수행하고, 이후 실제 메서드를 호출하거나 호출을 가로챕니다.
3. 이러한 프록시는 JDK Dynamic Proxy(인터페이스 기반)나 CGLIB(클래스 상속) 등을 통해 생성됩니다. @Aspect로 정의된 클래스 안에 **Pointcut**(적용 대상)과 **Advice**(전/후/주변 로직)가 명시됩니다.
4. 예를 들어 `@Around` 어노테이션을 사용하면 메서드 호출 전, 후 로직을 쉽게 삽입할 수 있고, 예외 발생 시 특정 처리를 할 수도 있습니다.
5. AOP는 구현부와 부가기능을 분리하여 유지보수를 용이하게 하고, 로깅, 트랜잭션, 보안, 모니터링 등 횡단 관심사를 깔끔하게 처리하도록 돕습니다.

---

### 질문 6. @Transactional 애노테이션이 어떻게 트랜잭션을 관리하나요?

#### **답변**

1. @Transactional은 Spring AOP의 한 사례로 볼 수 있습니다. 스프링은 @Transactional이 붙은 Bean 메서드를 호출할 때, **프록시**가介入하여 트랜잭션 매니저(PlatformTransactionManager)를 통해 DB 커넥션을 시작(또는 기존 트랜잭션에 참여)합니다.
2. 메서드가 정상적으로 종료되면, 프록시는 커밋을 호출하고, 예외가 발생하면 롤백을 호출합니다. 이렇게 DB 트랜잭션 범위를 자동으로 관리하므로, 개발자는 로직에 집중할 수 있습니다.
3. 다만, 자기 자신 메서드끼리 호출(예: this.internalCall()) 시에는 프록시가 관여하지 못하므로 트랜잭션이 적용되지 않을 수 있습니다. 이를 방지하기 위해서는 자기 자신이 아닌 다른 Bean이 호출하도록 구조를 설계하거나, AspectJ를 사용하는 방법 등이 있습니다.

---

## 3. MySQL 관련 질문

### 질문 7. MySQL에서 인덱스는 어떤 구조이며, 쿼리 성능을 어떻게 높일 수 있나요?

#### **답변**

1. MySQL(InnoDB 엔진 기준)에서 대부분의 인덱스는 **B+Tree** 구조로 되어 있습니다. B+Tree는 트리 높이를 낮게 유지하도록 설계되어, 많은 데이터를 저장하면서도 빠른 검색을 가능케 합니다.
2. InnoDB에서 **PRIMARY KEY**(기본키) 인덱스는 “클러스터링 인덱스”로서, 실제 테이블 레코드 자체가 B+Tree leaf에 저장됩니다. 반면, 보조 인덱스(secondary index)는 leaf 노드에 PK 값을 별도로 저장하고, 그 PK를 통해 테이블 데이터를 찾아가는 방식입니다.
3. 쿼리가 느릴 때는 **쿼리 실행 계획**(`EXPLAIN`)을 통해 인덱스 사용 여부, 풀 스캔 여부, 조인 방식 등을 점검해야 합니다. 필요한 열(WHERE절, JOIN에 사용되는 열)에 적절한 인덱스를 추가하면 성능이 개선될 수 있습니다.
4. 하지만 지나친 인덱스 생성은 INSERT/UPDATE 시 성능을 떨어뜨릴 수 있으므로, 빈도와 활용도를 고려해 인덱스를 설계해야 합니다.

---

### 질문 8. MySQL 트랜잭션 격리 수준과 InnoDB의 MVCC 방식은 어떻게 동작하나요?

#### **답변**

1. 트랜잭션 격리 수준은 한 트랜잭션이 다른 트랜잭션의 중간 변경 내용을 얼마나 볼 수 있는지를 정의합니다. MySQL의 기본값은 **REPEATABLE READ**이며, 다른 선택지로 READ COMMITTED, SERIALIZABLE 등이 있습니다.
2. REPEATABLE READ는 같은 트랜잭션 내에서 같은 SELECT 쿼리를 여러 번 실행해도 결과가 동일하게 보이도록 보장합니다.
3. InnoDB는 **MVCC(Multi-Version Concurrency Control)**를 통해 이 격리 수준을 구현합니다. 구체적으로, 각 행에는 트랜잭션 ID 정보가 기록되며, Undo Log에 이전 버전이 보관됩니다. SELECT 시점의 트랜잭션 ID보다 작은 버전 중에서 최신 것을 읽어 **일관된 스냅샷**을 제공합니다.
4. 이렇게 MVCC를 사용하면 대부분의 읽기 작업이 락 없이도 동작 가능해지고, 동시에 쓰기가 일어날 때도 Undo Log를 참고해 충돌을 최소화합니다.

---

## 4. Netty 관련 질문

### 질문 9. Netty는 어떤 원리로 고성능 네트워크를 구현하나요?

#### **답변**

1. Netty는 **NIO(Non-blocking I/O)** 기반의 **이벤트 루프(EventLoop)** 모델을 채택하고 있습니다. 이는 기존의 Thread-per-Connection 방식을 지양하고, 소수의 쓰레드가 다수의 소켓 채널을 관리하게 만듭니다.
2. 예를 들어, 이벤트 루프는 Selector를 사용해 I/O 이벤트(읽기, 쓰기, 연결 요청 등)를 감지하고, 그 이벤트가 발생한 채널에 대해 처리 로직을 호출합니다.
3. Netty에서는 **ChannelPipeline** 개념을 통해, 인바운드 및 아웃바운드 데이터를 단계별 핸들러로 넘깁니다. 인코더, 디코더, 비즈니스 로직 핸들러를 각각 체인으로 구성할 수 있어, 유지보수와 확장성이 뛰어납니다.
4. 또한 Netty는 OS별로 지원되는 **Native Transport**(예: Epoll on Linux, KQueue on macOS)와 **Zero-Copy** 기법 등을 통해 커널/유저 간 데이터 복사를 최소화하여 성능을 끌어올립니다.

---

### 질문 10. Netty 파이프라인에서 데이터가 인바운드/아웃바운드로 흐를 때, 예외가 발생하면 어떻게 처리되나요?

#### **답변**

1. Netty는 파이프라인을 **ChannelHandler**들의 체인으로 표현합니다. 데이터가 들어오는 방향(인바운드)과 나가는 방향(아웃바운드) 이벤트가 이 체인 위를 흐릅니다.
2. 특정 핸들러에서 예외가 발생하면, Netty는 그 예외를 **exceptionCaught** 메서드로 전파하여 처리하도록 합니다.
3. exceptionCaught가 오버라이딩되지 않았거나, 처리되지 않고 위로 전파되면 파이프라인 상위 레벨(또는 마지막 단계)까지 예외가 전달됩니다. 최종적으로도 처리하지 않으면 채널을 닫을 수도 있습니다.
4. 개발자는 예외 상황에 따라 로그를 남기고 연결을 정상 종료할지, 재시도할지, 특별한 예외 처리를 할지 결정해야 합니다.

---

## 5. 종합 아키텍처/적용 질문

### 질문 11. Spring + Netty + MySQL로 고부하(High Traffic) 시스템을 만든다면 어떤 아키텍처 구성을 권장하시나요?

#### **답변**

1. **Netty**로 외부 연결을 처리할 경우, 대규모 클라이언트 연결을 효율적으로 관리할 수 있습니다. 예를 들어, 실시간 채팅 서버나 대규모 TCP 연결이 필요한 서비스에 적합합니다.
2. Netty에서 받은 데이터를 **비즈니스 로직** 처리로 넘길 때는, 스프링의 @Service나 @Component Bean으로 구성된 로직을 호출할 수 있습니다. 이때 Netty의 이벤트 루프는 **논블로킹** 방식이므로, 블로킹 연산(예: DB쿼리, 파일 I/O)은 별도의 Worker ThreadPool로 분리하는 것이 좋습니다.
3. DB로는 **MySQL**을 사용하되, 트랜잭션이 필요한 로직은 스프링의 @Transactional을 통해 관리할 수 있습니다. 동시 트랜잭션이 많다면, 커넥션 풀(HikariCP 등) 크기를 최적화해야 하며, MySQL 자체도 연결 수와 쿼리 부하를 감당할 수 있게 튜닝해야 합니다.
4. 더 나아가, 대규모 트래픽을 처리하려면 **로드밸런서**(예: L4 또는 L7)로 여러 서버 인스턴스로 트래픽을 분산하고, **MySQL Replication**(마스터-슬레이브 구조)나 **Sharding**을 통해 DB 부하를 분산하기도 합니다.
5. 애플리케이션 서버는 Spring Boot나 Spring Cloud 등의 생태계를 활용할 수 있으며, Netty 기반 마이크로서비스 간 통신을 gRPC 등으로 구현하면, 대규모 시스템에서 성능과 확장성을 동시에 잡을 수 있습니다.

---

### 질문 12. Java 스레드풀과 Netty 이벤트 루프를 함께 사용할 때 주의점은 무엇인가요?

#### **답변**

1. Netty 이벤트 루프(이하 EL)는 **논블로킹 I/O**에 특화된 구조입니다. EL 스레드에서는 **Blocking 연산**(예: DB 쿼리, 동기식 HTTP 호출)을 수행하면 전체 성능이 급격히 떨어집니다.
2. 따라서, DB 처럼 블로킹 API를 호출해야 한다면, “별도의 쓰레드풀(Worker Pool)”을 구성하여 그곳에서 실행되도록 합니다. Netty 핸들러는 작업 요청을 EL에서 받아 큐에 넣고, 이후 Worker Pool에서 결과가 나오면 EL로 다시 결과를 전달하는 식의 구조를 가지면 됩니다.
3. Spring에서 `@Async`나 `TaskExecutor`를 설정하거나, 직접 `ThreadPoolExecutor`를 만들 수도 있습니다. 이렇게 하면 Netty EL은 I/O 이벤트에만 집중하고, 무거운 연산은 다른 곳에서 처리하므로 시스템 전반이 균형 있게 동작합니다.

---

## 자바 쓰레드 동기화

### 1. 자바에서의 쓰레드와 모니터 개념

1. 모든 동물은 생을 마친다(일반 원칙).
2. 자바의 `synchronized` 블록 또는 메서드는 “모니터(Monitor)”를 기반으로 동작한다(특정 원칙).
3. 따라서 `synchronized(obj)` 구문을 만나면, 쓰레드는 `obj`라는 객체의 모니터를 **획득**해야만 블록 내부 코드를 실행할 수 있다(결론).

먼저, “**모니터**”란, 자바에서 **동기화(synchronization)**를 지원하기 위해 객체마다 존재하는 구조체 혹은 **잠금(락) 기법**을 의미합니다. `synchronized (obj)`라고 하면, 해당 obj에 대한 **모니터 잠금**을 쓰레드가 얻어야만(획득해야만) 코드 실행이 가능합니다.

---

### 2. 쓰레드 상태(States) 개념

1. 자바 쓰레드는 **Thread.State** 열거형에 의해 여러 상태로 분류된다(일반 원칙).
2. 그 상태는 크게 **RUNNABLE, BLOCKED, WAITING, TIMED_WAITING, TERMINATED** 등이 있다(특정 원칙).
3. 따라서 어떤 특정 상황에서 쓰레드가 동작 중이거나(waiting, blocked), 종료되었는지 등을 구분할 수 있다(결론).

구체적으로:
- **RUNNABLE**: JVM 내부적으로 “실행 가능 상태” 또는 OS 스케줄러가 CPU에 올릴 수 있는 상태.
- **BLOCKED**: `synchronized`를 위해 모니터 잠금을 **획득하려고 시도**하지만, 잠금이 이미 다른 쓰레드에 의해 점유되어 있어서 **대기 중**인 상태.
- **WAITING**: 특정 조건(예: `wait()`, `join()`, `LockSupport.park()`)을 만족할 때까지 **“(notify 같은) 신호”**를 기다리는 상태.
- **TIMED_WAITING**: `wait(timeout)`, `sleep(timeout)` 처럼 유한 시간 동안만 대기하는 상태.
- **TERMINATED**: 쓰레드 실행이 끝난 상태.

---

### 3. synchronized 블록과 모니터 잠금

#### 3.1 진입 과정

1. 만약 A 쓰레드가 `synchronized(LockObject)` 구문을 실행하려고 시도한다(원인).
2. JVM은 **LockObject**라는 객체의 모니터를 확인한다(작동).
3. 객체 모니터가 사용 중(이미 다른 쓰레드가 갖고 있음)이면, A 쓰레드는 **BLOCKED** 상태로 전환되어 대기한다(결론 1).
4. 만약 모니터가 비어있다면, A 쓰레드는 **LockObject 모니터를 획득**하고 **RUNNABLE** 상태로 블록 내부 코드를 실행한다(결론 2).

즉, `BLOCKED` 상태는 “**모니터 잠금을 얻지 못해서** 대기”인 상태를 의미합니다. `synchronized(obj)`가 끝난 뒤(모니터가 release된 뒤)에야 다시 재시도할 수 있습니다.

#### 3.2 탈출 과정

1. A 쓰레드가 `synchronized(LockObject)` 블록의 끝에 도달하거나, 예외가 발생해 블록을 빠져나가면,
2. 잠금이 풀리고(= 모니터 반납), 대기 중인 다른 쓰레드(만약 있다면)가 모니터를 획득할 기회를 얻는다.

이로써 단일 객체 모니터를 통한 직렬화(One-at-a-time)가 달성됩니다.

---

### 4. wait()와 notify() 메서드의 메커니즘

#### 4.1 wait()가 하는 일

1. 자바에서 `obj.wait()`는 “**이 객체(obj)에 대한 모니터**”를 사용 중인 쓰레드가 임시로 모니터 **점유권을 포기**하고, 해당 객체의 **wait set**에 들어가 기다리게 하는 메서드이다(일반 원칙).
2. 즉, 쓰레드가 synchronized(obj) 블록 내부에서 `obj.wait()`를 호출하면, 자발적으로 모니터를 놓고(wait/release), JVM 내부의 **wait set**에 들어가 **WAITING** 상태로 전환된다(작동).
3. 그 결과, 모니터가 비워지므로 다른 쓰레드가 이 모니터를 획득할 수 있게 된다(결론).

`wait()`가 호출되어 WAITING 상태가 된 쓰레드는, **notify() / notifyAll()** 등이 호출되기 전까지는 깨어나지 않는다. 깨어난 뒤에는 모니터를 다시 획득해야만(`BLOCKED` 과정을 거칠 수 있음) 원래 코드의 이어지는 부분(`doSomething()`)을 실행한다.

#### 4.2 notify()와 notifyAll()

1. `obj.notify()`는 “**obj의 wait set**에서 대기 중인 쓰레드들 중 한 쓰레드**”를 깨우는(awakening) 역할이다(원칙).
2. 깨워진 쓰레드는 **모니터를 다시 획득하려고 시도**하지만, 그 순간 모니터가 이미 다른 쓰레드에 의해 사용 중이면, `BLOCKED` 상태로 잠시 대기해야 한다(작동).
3. 만약 모니터가 비어있다면, 바로 획득한 뒤 **RUNNABLE** 상태가 되어 `wait()` 호출 직후 코드를 이어서 수행한다(결론).

차이점:
- `notify()`는 한 개 쓰레드만 깨운다.
- `notifyAll()`는 wait set의 **모든 쓰레드**를 깨우지만, 결국 모니터는 한 쓰레드만 먼저 획득할 수 있으므로, 나머지 쓰레드는 `BLOCKED` → `RUNNABLE` 순으로 경합한다.

---

### 5. wait set의 작동 원리

1. 어떤 객체 LockObject에 대해 `LockObject.wait()`가 호출되면, 해당 쓰레드는 “LockObject의 **wait set**”에 추가된다(원칙).
2. 그동안 이 쓰레드는 모니터를 포기했으므로, **WAITING** 상태가 된다(작동).
3. 다른 쓰레드가 `LockObject.notify()`나 `LockObject.notifyAll()`을 호출하면, wait set에 있던 쓰레드(들)이 깨운 신호를 받고, 모니터 재획득을 시도한다(결론).

만약 **notify**나 **notifyAll**이 절대 호출되지 않는다면(즉, 락 오브젝트에 대한 신호가 없으면), **해당 쓰레드는 영원히 wait set에서 WAITING 상태**로 남을 수 있습니다. 따라서 wait/notify 구문은 반드시 한 쌍으로 설계해야 합니다.

---

### 6. 쓰레드 상태: BLOCKED vs WAITING

#### 6.1 BLOCKED 상태

1. A 쓰레드가 “**아직 모니터를 획득하지 못했음**”에도 불구하고, `synchronized(obj)` 구문에 진입하려 시도하면, 모니터가 다른 쓰레드에 점유된 경우 BLOCKED 상태가 된다(원칙).
2. 즉, **mutex 락**을 얻지 못해서 대기 중인 상태라고도 할 수 있다(작동).
3. 다른 쓰레드가 모니터를 해제할 때까지 스케줄링 자격이 없다가, 해제된 순간 JVM이 내부적으로 다시 락 획득을 시도해, 성공하면 RUNNABLE이 된다(결론).

#### 6.2 WAITING 상태

1. 쓰레드가 이미 모니터를 확보한 상황에서 `obj.wait()`를 호출하면, 모니터를 자발적으로 반납하고 wait set에 들어가게 된다(원칙).
2. 이는 “**명시적 신호(notify)**”가 오기 전까지는 깨어날 의사가 없음”을 표현하는 것이다(작동).
3. notify() 또는 notifyAll()을 통해 신호가 오면, 쓰레드는 모니터 재획득을 위해 경쟁하다가(그 사이 `BLOCKED`가 될 수도 있음) 성공시 RUNNABLE이 되고, 결국 wait() 호출 직후 코드를 계속 실행한다(결론).

**정리**: BLOCKED는 “모니터를 얻지 못해 문 앞에서 기다림”이고, WAITING은 “모니터를 얻었다가, wait() 호출로 임시 반납 후 notify를 기다리는 상태”입니다.

---

### 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

1. 만약 어떤 스레드가 `obj.wait()`로 WAITING 상태에 진입했는데, 이후 어떤 스레드도 `obj.notify()`나 `obj.notifyAll()`을 호출하지 않는다면(가정),
2. 그 쓰레드는 영원히 wait set에서 깨어나지 못하고, **영구적으로 WAITING 상태**로 남습니다(결론).
3. 실제 시스템에서도, wait/notify를 잘못 구현하면 데드락(deadlock)이나 **무한 대기**가 발생할 수 있으므로, 반드시 논리적으로 짝을 맞춰야 합니다(교훈).

---

### 8. JVM 레벨에서의 synchronized 작동 방식

#### 8.1 모니터 엔터/익시트(enter/exit)

1. 자바 바이트코드 상에서 `synchronized`는 `monitorenter`, `monitorexit` 명령어로 번역된다(원칙).
2. 실제 실행 시, JVM은 객체 헤더(mark word)와 모니터 구조를 이용해 락 상태, 소유 스레드, 대기열 등을 관리한다(작동).
3. `monitorenter`는 락 획득에 성공할 때까지 BLOCKED로 대기할 수도 있고, 성공하면 RUNNABLE로 진행한다(결론).
4. 블록 끝, 예외 발생 시 등에서 `monitorexit`가 호출되어 모니터를 해제한다.

#### 8.2 경량화 기법(Biased Locking, Lightweight Locking)

1. HotSpot JVM은 락 오버헤드를 줄이기 위해 **Biased Locking**이나 **Lightweight Locking** 같은 최적화를 한다(원칙).
2. 단일 스레드가 계속해서 같은 락을 잡는다면 편향 모드(biased)로 전환해 락 재획득 비용을 제거(작동).
3. 스레드 간 실질적 경합이 일어나기 전까지, 락 연산이 매우 가볍게 처리된다(결론).

---

### 9. 실제 동작 시나리오 예시

이제 위의 개념을 토대로, 질문에서 주어진 예시 코드를 분석해봅니다:

```java
synchronized(LockObject) {
    LockObject.wait();
    doSomething();
}
```

1. 이 구문이 실행될 때, 우선 `synchronized(LockObject)`에 진입하려면, 현재 쓰레드는 **LockObject 모니터**를 획득해야 합니다. 이미 다른 쓰레드가 이 모니터를 잡고 있으면, 이 쓰레드는 **BLOCKED 상태**가 되어 모니터 획득을 대기합니다.
2. 만약 성공적으로 모니터를 얻었다면, `LockObject.wait()`를 호출합니다. 이 순간, 쓰레드는 **(a) 모니터를 release**하고, **(b) LockObject의 wait set**에 들어가며, 쓰레드 상태는 **WAITING**으로 바뀝니다.
3. 이제 다른 쓰레드가 `LockObject.notify()` 또는 `notifyAll()`을 호출해주기 전까지는, 이 쓰레드는 깨어나지 못합니다.
4. notify가 도착하면, 대기 중인 쓰레드(혹은 여러 쓰레드) 중 하나(또는 전부)가 **“모니터 재획득”**을 시도합니다. 락을 획득하지 못하면 **BLOCKED**, 획득에 성공하면 **RUNNABLE** 상태로 돌아가 `doSomething()`을 실행합니다.
5. `doSomething()` 실행 후, `synchronized(LockObject)` 블록 끝에 도달해 **monitorexit**가 일어나고, 쓰레드는 모니터를 해제합니다.

---

### 10. 핵심 요약 및 결론

1. **Java 스레드 관리**에서 `synchronized`는 객체 모니터를 기반으로 임계영역 접근을 직렬화한다. 이때 모니터가 이미 점유 중이면 새로 진입하려는 쓰레드는 BLOCKED가 된다.
2. **wait()**는 모니터를 임의로 반납하고 객체의 wait set에 들어가게 하여 쓰레드를 WAITING 상태로 만든다. 이때 **notify()** 혹은 **notifyAll()**이 호출되기 전까지는 깨어나지 못한다.
3. notify가 없으면 해당 쓰레드는 영구 대기할 수 있으므로, 반드시 올바른 로직으로 wait/notify 쌍을 구성해야 한다.
4. **BLOCKED** 상태는 모니터를 아직 얻지 못해 문 앞에서 기다리는 상태이고, **WAITING** 상태는 이미 모니터를 얻은 뒤 wait()를 호출해 놓은 상태다. 깨어날 때도 모니터를 다시 얻어야 하므로, 일시적으로 BLOCKED로 돌아갈 수 있다.
5. JVM 레벨에서는 `monitorenter`/`monitorexit`, wait set, 락 최적화 등 다양한 내부 메커니즘으로 이 과정을 관리한다.

이처럼 자바의 **동기화 모델**은 모니터 잠금과 wait/notify를 결합하여, 스레드 간 통신(Condition Wait)과 임계영역 보호(Mutex Lock)를 동시에 지원합니다. 이것이 자바가 멀티쓰레드 프로그래밍을 강력하고 안전하게 제공하는 기저가 됩니다.

> **결론**: synchronized, wait, notify를 올바르게 사용하기 위해서는
> 1) 모니터 잠금의 획득/해제 시점
> 2) BLOCKED vs WAITING 상태 구분
> 3) notify/notifyAll을 통한 대기 해제 원리
> 등을 숙지해야 하며, 이는 JVM의 모니터 구조와 자바 스레드 상태 모델을 이해해야 명확히 파악할 수 있습니다.

# 자바 쓰레드, 상태, 그리고 모니터 등

## 1. 자바에서의 쓰레드와 모니터 개념

1. 모든 동물은 생을 마친다(일반 원칙).
2. 자바의 `synchronized` 블록 또는 메서드는 “모니터(Monitor)”를 기반으로 동작한다(특정 원칙).
3. 따라서 `synchronized(obj)` 구문을 만나면, 쓰레드는 `obj`라는 객체의 모니터를 **획득**해야만 블록 내부 코드를 실행할 수 있다(결론).

먼저, “**모니터**”란, 자바에서 **동기화(synchronization)**를 지원하기 위해 객체마다 존재하는 구조체 혹은 **잠금(락) 기법**을 의미합니다. `synchronized (obj)`라고 하면, 해당 obj에 대한 **모니터 잠금**을 쓰레드가 얻어야만(획득해야만) 코드 실행이 가능합니다.

---

## 2. 쓰레드 상태(States) 개념

1. 자바 쓰레드는 **Thread.State** 열거형에 의해 여러 상태로 분류된다(일반 원칙).
2. 그 상태는 크게 **RUNNABLE, BLOCKED, WAITING, TIMED_WAITING, TERMINATED** 등이 있다(특정 원칙).
3. 따라서 어떤 특정 상황에서 쓰레드가 동작 중이거나(waiting, blocked), 종료되었는지 등을 구분할 수 있다(결론).

구체적으로:
- **RUNNABLE**: JVM 내부적으로 “실행 가능 상태” 또는 OS 스케줄러가 CPU에 올릴 수 있는 상태.
- **BLOCKED**: `synchronized`를 위해 모니터 잠금을 **획득하려고 시도**하지만, 잠금이 이미 다른 쓰레드에 의해 점유되어 있어서 **대기 중**인 상태.
- **WAITING**: 특정 조건(예: `wait()`, `join()`, `LockSupport.park()`)을 만족할 때까지 **“(notify 같은) 신호”**를 기다리는 상태.
- **TIMED_WAITING**: `wait(timeout)`, `sleep(timeout)` 처럼 유한 시간 동안만 대기하는 상태.
- **TERMINATED**: 쓰레드 실행이 끝난 상태.

---

## 3. synchronized 블록과 모니터 잠금

### 3.1 진입 과정

1. 만약 A 쓰레드가 `synchronized(LockObject)` 구문을 실행하려고 시도한다(원인).
2. JVM은 **LockObject**라는 객체의 모니터를 확인한다(작동).
3. 객체 모니터가 사용 중(이미 다른 쓰레드가 갖고 있음)이면, A 쓰레드는 **BLOCKED** 상태로 전환되어 대기한다(결론 1).
4. 만약 모니터가 비어있다면, A 쓰레드는 **LockObject 모니터를 획득**하고 **RUNNABLE** 상태로 블록 내부 코드를 실행한다(결론 2).

즉, `BLOCKED` 상태는 “**모니터 잠금을 얻지 못해서** 대기”인 상태를 의미합니다. `synchronized(obj)`가 끝난 뒤(모니터가 release된 뒤)에야 다시 재시도할 수 있습니다.

### 3.2 탈출 과정

1. A 쓰레드가 `synchronized(LockObject)` 블록의 끝에 도달하거나, 예외가 발생해 블록을 빠져나가면,
2. 잠금이 풀리고(= 모니터 반납), 대기 중인 다른 쓰레드(만약 있다면)가 모니터를 획득할 기회를 얻는다.

이로써 단일 객체 모니터를 통한 직렬화(One-at-a-time)가 달성됩니다.

---

## 4. wait()와 notify() 메서드의 메커니즘

### 4.1 wait()가 하는 일

1. 자바에서 `obj.wait()`는 “**이 객체(obj)에 대한 모니터**”를 사용 중인 쓰레드가 임시로 모니터 **점유권을 포기**하고, 해당 객체의 **wait set**에 들어가 기다리게 하는 메서드이다(일반 원칙).
2. 즉, 쓰레드가 synchronized(obj) 블록 내부에서 `obj.wait()`를 호출하면, 자발적으로 모니터를 놓고(wait/release), JVM 내부의 **wait set**에 들어가 **WAITING** 상태로 전환된다(작동).
3. 그 결과, 모니터가 비워지므로 다른 쓰레드가 이 모니터를 획득할 수 있게 된다(결론).

`wait()`가 호출되어 WAITING 상태가 된 쓰레드는, **notify() / notifyAll()** 등이 호출되기 전까지는 깨어나지 않는다. 깨어난 뒤에는 모니터를 다시 획득해야만(`BLOCKED` 과정을 거칠 수 있음) 원래 코드의 이어지는 부분(`doSomething()`)을 실행한다.

### 4.2 notify()와 notifyAll()

1. `obj.notify()`는 “**obj의 wait set**에서 대기 중인 쓰레드들 중 한 쓰레드**”를 깨우는(awakening) 역할이다(원칙).
2. 깨워진 쓰레드는 **모니터를 다시 획득하려고 시도**하지만, 그 순간 모니터가 이미 다른 쓰레드에 의해 사용 중이면, `BLOCKED` 상태로 잠시 대기해야 한다(작동).
3. 만약 모니터가 비어있다면, 바로 획득한 뒤 **RUNNABLE** 상태가 되어 `wait()` 호출 직후 코드를 이어서 수행한다(결론).

차이점:
- `notify()`는 한 개 쓰레드만 깨운다.
- `notifyAll()`는 wait set의 **모든 쓰레드**를 깨우지만, 결국 모니터는 한 쓰레드만 먼저 획득할 수 있으므로, 나머지 쓰레드는 `BLOCKED` → `RUNNABLE` 순으로 경합한다.

---

## 5. wait set의 작동 원리

1. 어떤 객체 LockObject에 대해 `LockObject.wait()`가 호출되면, 해당 쓰레드는 “LockObject의 **wait set**”에 추가된다(원칙).
2. 그동안 이 쓰레드는 모니터를 포기했으므로, **WAITING** 상태가 된다(작동).
3. 다른 쓰레드가 `LockObject.notify()`나 `LockObject.notifyAll()`을 호출하면, wait set에 있던 쓰레드(들)이 깨운 신호를 받고, 모니터 재획득을 시도한다(결론).

만약 **notify**나 **notifyAll**이 절대 호출되지 않는다면(즉, 락 오브젝트에 대한 신호가 없으면), **해당 쓰레드는 영원히 wait set에서 WAITING 상태**로 남을 수 있습니다. 따라서 wait/notify 구문은 반드시 한 쌍으로 설계해야 합니다.

---

## 6. 쓰레드 상태: BLOCKED vs WAITING

### 6.1 BLOCKED 상태

1. A 쓰레드가 “**아직 모니터를 획득하지 못했음**”에도 불구하고, `synchronized(obj)` 구문에 진입하려 시도하면, 모니터가 다른 쓰레드에 점유된 경우 BLOCKED 상태가 된다(원칙).
2. 즉, **mutex 락**을 얻지 못해서 대기 중인 상태라고도 할 수 있다(작동).
3. 다른 쓰레드가 모니터를 해제할 때까지 스케줄링 자격이 없다가, 해제된 순간 JVM이 내부적으로 다시 락 획득을 시도해, 성공하면 RUNNABLE이 된다(결론).

### 6.2 WAITING 상태

1. 쓰레드가 이미 모니터를 확보한 상황에서 `obj.wait()`를 호출하면, 모니터를 자발적으로 반납하고 wait set에 들어가게 된다(원칙).
2. 이는 “**명시적 신호(notify)**”가 오기 전까지는 깨어날 의사가 없음”을 표현하는 것이다(작동).
3. notify() 또는 notifyAll()을 통해 신호가 오면, 쓰레드는 모니터 재획득을 위해 경쟁하다가(그 사이 `BLOCKED`가 될 수도 있음) 성공시 RUNNABLE이 되고, 결국 wait() 호출 직후 코드를 계속 실행한다(결론).

**정리**: BLOCKED는 “모니터를 얻지 못해 문 앞에서 기다림”이고, WAITING은 “모니터를 얻었다가, wait() 호출로 임시 반납 후 notify를 기다리는 상태”입니다.

---

## 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

1. 만약 어떤 스레드가 `obj.wait()`로 WAITING 상태에 진입했는데, 이후 어떤 스레드도 `obj.notify()`나 `obj.notifyAll()`을 호출하지 않는다면(가정),
2. 그 쓰레드는 영원히 wait set에서 깨어나지 못하고, **영구적으로 WAITING 상태**로 남습니다(결론).
3. 실제 시스템에서도, wait/notify를 잘못 구현하면 데드락(deadlock)이나 **무한 대기**가 발생할 수 있으므로, 반드시 논리적으로 짝을 맞춰야 합니다(교훈).

---

## 8. JVM 레벨에서의 synchronized 작동 방식

### 8.1 모니터 엔터/익시트(enter/exit)

1. 자바 바이트코드 상에서 `synchronized`는 `monitorenter`, `monitorexit` 명령어로 번역된다(원칙).
2. 실제 실행 시, JVM은 객체 헤더(mark word)와 모니터 구조를 이용해 락 상태, 소유 스레드, 대기열 등을 관리한다(작동).
3. `monitorenter`는 락 획득에 성공할 때까지 BLOCKED로 대기할 수도 있고, 성공하면 RUNNABLE로 진행한다(결론).
4. 블록 끝, 예외 발생 시 등에서 `monitorexit`가 호출되어 모니터를 해제한다.

### 8.2 경량화 기법(Biased Locking, Lightweight Locking)

1. HotSpot JVM은 락 오버헤드를 줄이기 위해 **Biased Locking**이나 **Lightweight Locking** 같은 최적화를 한다(원칙).
2. 단일 스레드가 계속해서 같은 락을 잡는다면 편향 모드(biased)로 전환해 락 재획득 비용을 제거(작동).
3. 스레드 간 실질적 경합이 일어나기 전까지, 락 연산이 매우 가볍게 처리된다(결론).

---

## 9. 실제 동작 시나리오 예시

이제 위의 개념을 토대로, 질문에서 주어진 예시 코드를 분석해봅니다:

```java
synchronized(LockObject) {
    LockObject.wait();
    doSomething();
}
```

1. 이 구문이 실행될 때, 우선 `synchronized(LockObject)`에 진입하려면, 현재 쓰레드는 **LockObject 모니터**를 획득해야 합니다. 이미 다른 쓰레드가 이 모니터를 잡고 있으면, 이 쓰레드는 **BLOCKED 상태**가 되어 모니터 획득을 대기합니다.
2. 만약 성공적으로 모니터를 얻었다면, `LockObject.wait()`를 호출합니다. 이 순간, 쓰레드는 **(a) 모니터를 release**하고, **(b) LockObject의 wait set**에 들어가며, 쓰레드 상태는 **WAITING**으로 바뀝니다.
3. 이제 다른 쓰레드가 `LockObject.notify()` 또는 `notifyAll()`을 호출해주기 전까지는, 이 쓰레드는 깨어나지 못합니다.
4. notify가 도착하면, 대기 중인 쓰레드(혹은 여러 쓰레드) 중 하나(또는 전부)가 **“모니터 재획득”**을 시도합니다. 락을 획득하지 못하면 **BLOCKED**, 획득에 성공하면 **RUNNABLE** 상태로 돌아가 `doSomething()`을 실행합니다.
5. `doSomething()` 실행 후, `synchronized(LockObject)` 블록 끝에 도달해 **monitorexit**가 일어나고, 쓰레드는 모니터를 해제합니다.

---

## 10. 핵심 요약 및 결론

1. **Java 스레드 관리**에서 `synchronized`는 객체 모니터를 기반으로 임계영역 접근을 직렬화한다. 이때 모니터가 이미 점유 중이면 새로 진입하려는 쓰레드는 BLOCKED가 된다.
2. **wait()**는 모니터를 임의로 반납하고 객체의 wait set에 들어가게 하여 쓰레드를 WAITING 상태로 만든다. 이때 **notify()** 혹은 **notifyAll()**이 호출되기 전까지는 깨어나지 못한다.
3. notify가 없으면 해당 쓰레드는 영구 대기할 수 있으므로, 반드시 올바른 로직으로 wait/notify 쌍을 구성해야 한다.
4. **BLOCKED** 상태는 모니터를 아직 얻지 못해 문 앞에서 기다리는 상태이고, **WAITING** 상태는 이미 모니터를 얻은 뒤 wait()를 호출해 놓은 상태다. 깨어날 때도 모니터를 다시 얻어야 하므로, 일시적으로 BLOCKED로 돌아갈 수 있다.
5. JVM 레벨에서는 `monitorenter`/`monitorexit`, wait set, 락 최적화 등 다양한 내부 메커니즘으로 이 과정을 관리한다.

이처럼 자바의 **동기화 모델**은 모니터 잠금과 wait/notify를 결합하여, 스레드 간 통신(Condition Wait)과 임계영역 보호(Mutex Lock)를 동시에 지원합니다. 이것이 자바가 멀티쓰레드 프로그래밍을 강력하고 안전하게 제공하는 기저가 됩니다.

> **결론**: synchronized, wait, notify를 올바르게 사용하기 위해서는
> 1) 모니터 잠금의 획득/해제 시점
> 2) BLOCKED vs WAITING 상태 구분
> 3) notify/notifyAll을 통한 대기 해제 원리
> 등을 숙지해야 하며, 이는 JVM의 모니터 구조와 자바 스레드 상태 모델을 이해해야 명확히 파악할 수 있습니다.

## **자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시**

여러 스레드가 동시에 `wait()`를 호출할 수 있고, 어떤 스레드는 `notify()` 또는 `notifyAll()`을 호출해서 대기 중인 스레드를 깨우는 구조입니다.

### 1. 핵심 개념 요약

1. **wait()**:
   - 현재 **모니터(동기화 락)**를 가진 스레드가 스스로 락을 해제하고, 해당 객체의 **wait set**에서 대기(WAITING 상태)로 들어감.
   - 추후 **notify** 또는 **notifyAll**이 호출되어야만 다시 모니터 획득을 재시도할 수 있음.

2. **notify() / notifyAll()**:
   - 같은 모니터를 사용하는 객체에서, **wait set**에 대기 중인 스레드들을 깨움.
   - 깨운 스레드는 모니터를 다시 획득해야만(RUNNABLE로 복귀 전) `wait()` 다음 코드를 실행할 수 있음.

3. **synchronized**(LockObject) 블록:
   - 해당 블록에 진입하기 위해서는 `LockObject`의 모니터를 획득해야 함.
   - 모니터는 “한 순간에 오직 한 스레드만 접근 가능”을 보장하여 상호배타(Mutual Exclusion)를 실현.
   - synchronized 블록 안에서는 **메모리 베리어**(happens-before)가 보장되어, 해당 블록 내부에서 이루어진 변경은 블록을 빠져나갈 때 외부에 반영되고, 반대로 외부 변경사항을 읽을 때 재주문(reordering)이 억제됨.

4. **"모든 스레드가 wait만 하고 notify를 안 하면?"**
   - 영원히 깨어나지 못함(Deadlock 또는 영구 대기).
   - 따라서 반드시 **notify()** 또는 **notifyAll()**을 호출해야 대기 중인 스레드들이 다시 동작을 이어갈 수 있음.

---

### 2. 예시 코드: 작업 준비(Producer)와 작업 처리(Consumer)

아래 시나리오에서:
- 다수의 Worker 스레드가 `wait()`를 호출해 대기 상태에 들어갑니다.
- Main 스레드(또는 다른 관리 스레드)가 특정 타이밍에 `notifyAll()`을 통해 Worker 스레드들을 깨웁니다.
- 깨어난 Worker 스레드는 모니터를 다시 획득한 후 `doSomething()` 로직을 실행합니다.
- 실제 프로덕션 환경에서는 “job 큐”를 사용하거나, “조건변수(Condition)”를 사용하는 것이 더 직관적이지만, 여기서는 `wait/notify`의 기본 구조를 보여주기 위해 단순화했습니다.

```java
public class WaitNotifyDemo {

    private final Object lockObject = new Object();
    private boolean canProcess = false;
    // 이 변수는 'notify' 시점에 true로 바뀌며, 다른 스레드가 doSomething 해도 좋다는 신호라고 가정.

    /**
     * Worker 스레드가 호출하는 메서드.
     * 이 메서드는 'canProcess'가 false인 상태에서는 wait()를 통해 대기한다.
     * 이후 notifyAll()로 깨어나면 doSomething()을 수행한다.
     */
    public void workerMethod() {
        synchronized (lockObject) {
            while (!canProcess) {
                try {
                    // [1] 모니터 획득 중
                    // [2] wait()를 호출 → 모니터 반납 + wait set 대기 (WAITING 상태)
                    lockObject.wait();
                    // 깨어난 시점에 모니터 재획득 필요 → 만약 다른 스레드가 락을 점유 중이면 BLOCKED로 대기
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    System.out.println("Worker interrupted during wait");
                    return;
                }
            }
            // 이 시점: canProcess == true 이며, 모니터를 다시 획득한 상태
            doSomething();
            // synchronized 블록 끝 → monitorexit
        }
    }

    /**
     * 다른 스레드나 컨트롤러가 호출해서 'canProcess'를 true로 만들고,
     * 대기 중인 Worker 스레드들을 깨우는 역할을 한다.
     */
    public void enableProcessingAndNotifyAll() {
        synchronized (lockObject) {
            // [1] 모니터 획득
            // [2] canProcess 값을 true로 설정
            canProcess = true;
            // [3] wait set에 있는 모든 스레드 깨우기
            lockObject.notifyAll();
            // notifyAll()은 wait set에 있는 스레드들을 깨운다.
            // 하지만 깨어난 스레드들이 모니터 재획득을 위해 경합 → 한 번에 한 스레드씩 들어와서 doSomething() 실행
        }
    }

    /**
     * 실제로 처리해야 할 로직 (예: DB 저장, API 호출 등)
     */
    private void doSomething() {
        System.out.println(Thread.currentThread().getName() + " is doing something...");
        // 좀 더 복잡한 로직이라고 가정해도 됨
        try {
            Thread.sleep(500); // 시뮬레이션
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * 테스트를 위한 메인: 여러 Worker 스레드를 띄우고, 일정 시간 뒤 enableProcessingAndNotifyAll()을 호출.
     */
    public static void main(String[] args) throws InterruptedException {
        WaitNotifyDemo demo = new WaitNotifyDemo();

        // 스레드 5개 생성 → 모두 workerMethod()를 수행
        for (int i = 1; i <= 5; i++) {
            Thread worker = new Thread(demo::workerMethod, "Worker-" + i);
            worker.start();
        }

        // 이 시점에 canProcess = false 이므로, workerMethod 내에서 모두 wait() 상태로 진입할 것임
        System.out.println("All workers are likely waiting now...");

        // 3초 후에 notifyAll()을 호출 (canProcess = true)
        Thread.sleep(3000);
        System.out.println("Now let's enable processing and notify all!");
        demo.enableProcessingAndNotifyAll();

        // 이후 main 스레드가 종료될 때까지 대기
        // 실제로 Worker 스레드들이 깨어나서 doSomething()을 실행할 것임
    }
}
```

#### 2.1 코드 흐름 설명

1. **Worker 스레드** (`workerMethod()`):
   - `synchronized (lockObject)`로 진입
   - `while(!canProcess)` 조건 체크 → `canProcess`가 `false`면 `wait()` 호출
   - `wait()`를 호출하면 모니터를 반납하고 `WAITING` 상태로 들어감.
   - (나중에) `notifyAll()`로 깨워진 뒤 → 모니터 재획득(이때 BLOCKED 일 수 있음) → `doSomething()` 실행.

2. **Notifier 스레드** 또는 **Main 스레드** (`enableProcessingAndNotifyAll()`):
   - `synchronized (lockObject)`로 진입
   - `canProcess = true;` 설정 → 이제 Worker가 doSomething을 수행해도 된다는 조건 충족
   - `lockObject.notifyAll()` 호출 → wait set에서 WAITING 중인 스레드들을 깨움
   - 깨어난 스레드가 모니터를 얻으면 `while(!canProcess)`를 탈출하고 `doSomething()` 호출

3. **테스트(main)**:
   - 5개의 Worker 스레드가 `workerMethod()`를 시작 → 모두 `wait()`로 대기
   - 3초 뒤 `enableProcessingAndNotifyAll()`이 호출 → 모든 worker가 깨어나 `doSomething()` 진행

이 시나리오에서는 “모든 스레드가 락을 얻자마자 `wait()`해버리면 `notify()`는 어떻게 호출되냐?”라는 의문에 대해, 실제로는 **다른 스레드(혹은 동일 스레드의 다른 시점)**가 같은 락을 잡고 `notify()`를 실행해야만 된다는 사실을 보여줍니다.

---

### 3. notify는 누가 언제 실행하고, 그 결과는 무엇인가?

1. notify/notifyAll은 **동일한 모니터 객체**에 대해 `synchronized(obj)` 블록 내부에서만 유효합니다. 즉, 락을 획득한 스레드만 `obj.notify()`를 호출할 수 있습니다.
2. 위 예시에서 `enableProcessingAndNotifyAll()` 메서드가 바로 그 "notify를 하는 로직"입니다. 이 메서드를 실행하는 스레드(메인 스레드)는 **모니터 획득** → `canProcess = true`로 변경 → `notifyAll()`을 호출하고 → 블록을 빠져나갑니다.
3. notifyAll()이 호출되면, **모니터의 wait set**에 있던 스레드가 깨어납니다. 그러나 모니터는 여전히 메인 스레드가 잡고 있으므로, Worker 스레드가 깨어나도 **우선은 BLOCKED**가 될 수 있습니다.
4. 메인 스레드가 모니터를 release(블록 끝)하면, 깨어난 Worker 스레드 중 하나가 모니터를 획득 → `while(!canProcess)` 조건을 재확인 → 이제 true이므로 탈출 → `doSomething()` 실행.
5. 그 다음 모니터를 해제하면 또 다른 Worker 스레드가 모니터를 획득하고 로직을 수행합니다.

---

### 4. "모든 스레드가 wait만 하고 notify를 호출하는 스레드가 없다면?" 발생하는 문제

1. 모든 스레드가 `wait()`로 들어가면, **WAITING 상태**에서 멈춰있습니다.
2. **notify**나 **notifyAll**을 호출하는 스레드가 없다면, 깨어날 방법이 없습니다.
3. 따라서 “**무한 대기**” 혹은 “데드락”이 발생합니다.
4. 이를 방지하려면, **로직상 반드시** 특정 시점에(조건을 만족하면) `notify()`나 `notifyAll()`을 호출해야 합니다.

---

### 5. synchronized(락)을 사용하면 메모리 베리어가 적용되는가?

1. 자바 `synchronized` 구문은 **락을 획득할 때**와 **락을 해제할 때** 메모리 베리어(정확히는 happens-before 관계)를 부여합니다.
2. 즉, 한 스레드가 synchronized 블록 안에서 변경한 데이터는 블록을 빠져나오기 전에 **메모리 가시화**가 이루어집니다. 그리고 다음에 모니터를 획득한 스레드는 그 변경된 상태를 볼 수 있게 됩니다.
3. 따라서 `synchronized`를 적절히 사용하면, 다른 스레드가 임의로 바꾼 값을 못 읽거나, reorder로 인해 예기치 않은 동작이 벌어지는 상황을 완화할 수 있습니다.
4. wait/notify 또한 synchronized 블록 내부에서만 허용되므로, 이로 인한 상태 변경도 자연스럽게 메모리 베리어 효과가 반영됩니다.

---

### 6. 마지막 정리

1. **모니터**(= 객체 락, synchronized)는 “한 순간에 오직 한 스레드만 해당 구간을 실행”하도록 보장합니다.
2. 모든 스레드가 `synchronized(LockObject)`에 들어오자마자 `wait()`만 호출한다면, 실제로는 누군가가 그 모니터를 잡고 `notify()`를 호출하지 않는 이상, 깨어날 스레드가 없게 됩니다.
3. 따라서 “누군가가 `notify()`나 `notifyAll()`을 명시적으로 호출”해야 합니다. 자동 호출되는 일이 절대 없으므로, 프로그래머가 필요한 시점에서 로직을 설계해야 합니다.
4. 위 예시 코드처럼 **조건변수** 성격으로 `wait/notify`를 쓰고자 한다면, 보통 `while(!condition) wait()` 패턴을 씁니다. 이것이 자바에서 표준적인 접근(“guarded block”)입니다.
5. 메모리 베리어 관점에서 `synchronized`는 쓰레드 간 메모리 일관성을 높여주며, wait/notify로 멀티쓰레드 간 통신(이벤트/신호)을 구현합니다.

이상으로, 질문에서 제기된 **“A스레드, B스레드 모두 `wait()`만 반복하면 `doSomething()`은 언제 실행되는가?”, “notify는 명시적으로 호출해야 하는가?”, “메모리 장벽은 어떻게 보장되는가?”** 등 내용을 모두 포함하여, **프로덕션 코드 예시**를 통해 논리적으로 설명했습니다.

## 쓰레드 wait와 notify 예제

1. `synchronized` 블록 내부에서 `wait()`가 반복해서 호출되면 도대체 언제 `doSomething()`이 실행되는가?
2. `notify()`는 누가, 어떻게 호출해야 하는가? 자동 호출되는가?
3. 현실 세계의 **프로덕션 코드**에서 이런 로직을 어떻게 작성하며, **TDD(Test-Driven Development)** 원칙에 따라 테스트 코드를 어떻게 설계할 수 있는가?
4. `synchronized`가 메모리 장벽(메모리 배리어)을 제공하므로, 쓰레드 간에 공유 객체에 대한 접근을 어떻게 제어하는지?

아래 코드는 “**생산자-소비자(Producer-Consumer)**” 시나리오로 구현해 보았습니다. “생산자 스레드”가 작업(Tasks)을 큐에 추가하면 `notify()`를 호출하여, 대기 중인 “소비자 스레드”들을 깨워서 `doSomething()`을 수행하게 하는 구조입니다.

---

### 1. 왜 `notify()`가 필요하고, 누가 언제 호출해야 하는가?

#### 요약

- `wait()`로 **WAITING 상태**가 된 스레드는, **반드시** 다른 쓰레드가 `notify()`(또는 `notifyAll()`)를 호출해야만 깨어날 수 있습니다.
- `notify()`는 자바에서 **자동으로** 호출되는 것이 아니라, **코드 상에서 명시적으로** 호출해야 합니다.
- `notify()`를 호출하는 스레드도 `synchronized(lockObj)` 블록 안에서 모니터를 잡은 상태여야 합니다(왜냐하면 wait/notify 모두 모니터를 필요로 하기 때문입니다).
- `notify()`를 호출하고 블록을 빠져나가야(=모니터 해제) 다른 쓰레드가 모니터를 재획득할 수 있으므로, 그때 비로소 깨어난 스레드가 `doSomething()`을 실행할 기회를 얻게 됩니다.

#### 예시 흐름

1. 스레드 A가 `synchronized(lockObj)`로 모니터 획득
2. `lockObj.wait()` 호출 → **모니터 해제** & **WAITING 상태** & **lockObj의 wait set**에 들어감
3. 스레드 B가 `synchronized(lockObj)`로 진입 (A가 해제했으므로 B가 락 획득 가능)
4. 스레드 B가 어떤 조건을 달성하면 `lockObj.notify()` 호출 후, `synchronized` 블록을 빠져나옴
5. 대기중이던 A 스레드가 깨어나서(=notify 신호) 모니터 재획득 시도 → 획득 성공 → `doSomething()` 실행

**중요**: 모든 스레드가 들어가자마자 `wait()`만 호출한다면, **아무도 notify()를 호출하지 않는 상황**이 되어 영원히 깨어나지 못할 수 있습니다. 실제 프로덕션 코드에서는 “이벤트가 발생했을 때 notify()를 호출하는” 쪽 로직이 반드시 필요합니다.

---

### 2. 현실 세계 프로덕션 코드 예시: “TaskQueue” 시나리오

**시나리오**:
- `TaskQueue` 클래스에 작업(Task)을 추가(생산자)하고, 작업이 없으면 대기(소비자)가 발생.
- 소비자는 `wait()`를 통해 **대기**하고, 작업이 들어오면 `notify()`를 통해 **깨운다**.

#### 2.1 TaskQueue 구현 (프로덕션 코드)

```java
public class TaskQueue {
    private final Object lock = new Object();
    private final LinkedList<String> tasks = new LinkedList<>();

    /**
     * 작업을 큐에 추가하고, 대기 중인 스레드를 깨운다.
     */
    public void addTask(String task) {
        synchronized(lock) {
            tasks.add(task);
            // 작업이 추가되었으니, wait()로 대기 중인 스레드를 깨운다.
            lock.notify();
        }
    }

    /**
     * 큐에서 작업을 꺼낸다.
     * 만약 큐가 비어 있으면 wait()로 대기한다.
     */
    public String getTask() throws InterruptedException {
        synchronized(lock) {
            while (tasks.isEmpty()) {
                // 여기서 락을 포기하고 wait set에 들어가 WAITING 상태
                lock.wait();
            }
            // 깨어난 뒤 모니터를 다시 획득해야 함 (획득 시 BLOCKED → RUNNABLE)
            return tasks.removeFirst();
        }
    }

    /**
     * 큐에 남아있는 작업 수를 반환.
     */
    public int size() {
        synchronized(lock) {
            return tasks.size();
        }
    }
}
```

1. `addTask()` 메서드는 새 작업을 추가한 후, `lock.notify()`를 호출합니다. 이렇게 함으로써 **대기(wait) 중인 스레드**가 있다면 하나를 깨울 수 있습니다.
2. `getTask()`는 큐가 비어있으면 `lock.wait()`를 호출하여, 모니터를 즉시 릴리즈하고 **WAITING 상태**로 들어갑니다. 여기서 `while`문을 사용하는 이유는, 깨어난 뒤(=notify)에도 여전히 큐가 비어있을 수 있기 때문입니다(“spurious wakeup” 처리).
3. `size()` 등 다른 메서드도 `synchronized(lock)`를 통해 동시성을 제어합니다.

---

#### 2.2 실제 doSomething() 메서드 예시

이제 `doSomething()` 로직을 “**작업을 실제 처리**”하는 것으로 가정해봅시다. 예를 들어, “큐에서 가져온 task를 처리”하는 스레드 코드를 작성할 수 있습니다.

```java
public class TaskProcessor {
    private final TaskQueue queue;

    public TaskProcessor(TaskQueue queue) {
        this.queue = queue;
    }

    /**
     * 무한 루프 형태의 소비자 스레드 역할
     * - 큐에서 작업을 꺼내 처리(doSomething()).
     * - 큐가 비어있으면 wait() 상태로 대기.
     */
    public void processLoop() {
        while (true) {
            try {
                String task = queue.getTask();
                // 여기서 getTask()는 synchronized(lock) + wait()를 통한 블록
                doSomething(task);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break; // 스레드를 종료할 수도 있음
            }
        }
    }

    private void doSomething(String task) {
        // 실제 로직: 작업 처리
        System.out.println("Processing: " + task);
        // 여기서 DB 저장, 파일 쓰기 등등 수행할 수 있음
    }
}
```

- `TaskProcessor`는 “소비자(Consumer) 스레드”로, `queue.getTask()`를 통해 작업이 들어올 때까지 대기(`wait()`)하다가, 작업이 있을 때만 깨어나서 `doSomething()`을 수행합니다.
- `getTask()` 메서드는 내부적으로 `synchronized`+`wait()`+`notify()`를 통해 동작하므로, 안전하게 큐를 공유합니다.

---

### 3. 테스트 코드 (TDD 접근)

JUnit을 사용하여, **멀티쓰레드 환경**에서 정상 동작하는지 간단한 테스트를 보여주겠습니다.

```java
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertTrue;

public class TaskQueueTest {

    @Test
    public void testProducerConsumer() throws InterruptedException {
        TaskQueue queue = new TaskQueue();
        TaskProcessor processor = new TaskProcessor(queue);

        // 1. 소비자 스레드(Worker) 생성
        Thread worker = new Thread(processor::processLoop);
        worker.start();

        // 2. 작업을 여러 개 추가 (생산자 역할)
        queue.addTask("Task1");
        queue.addTask("Task2");

        // 3. 소비자가 처리할 시간을 조금 주고...
        Thread.sleep(500);

        // 4. 아직 queue가 비어 있기를 기대
        int size = queue.size();
        assertTrue(size == 0, "All tasks should have been consumed by now.");

        // 5. 스레드 종료를 위해 인터럽트
        worker.interrupt();
        worker.join();
    }
}
```

1. **`TaskQueue queue`**를 만들고, **`TaskProcessor processor`**로 소비자 스레드를 구성합니다.
2. `processor::processLoop`는 무한 루프에서 `queue.getTask()` → `doSomething(task)`를 반복합니다. 큐가 비어있을 경우 자동으로 `wait()`를 통해 대기한다.
3. 생산자 역할(`main` 스레드)이 `queue.addTask("TaskX")`를 호출하면, 내부에서 `notify()`가 일어나고 소비자가 깨어나 `doSomething()`를 수행합니다.
4. 좀 기다린 후(queue가 비어있을 시간 간격) `assertTrue(size == 0)`로 테스트합니다.

---

### 4. notify는 명시적으로 호출해야 하며, 누가 언제 실행할까?

1. **명시적 호출**: 자바 언어에서는 **notify/notifyAll**을 자동으로 호출해주지 않습니다. 프로그래머가 적절한 시점에 `lockObj.notify()`를 해야만 **WAITING** 스레드를 깨울 수 있습니다.
2. **누가 언제**: 일반적으로 “**상태 변화**”가 일어날 때(예: 큐가 비어 있다가 새 작업이 들어온 경우, 어떤 조건이 충족되었을 때) `notify()`를 호출합니다. 위 예시의 `addTask()`처럼, “중요한 상태 변화를 일으킨 쪽”이 다른 스레드(대기 중인 소비자)에게 알려주는 개념입니다.
3. **notify() 호출 스레드**는 `synchronized(obj)` 블록 안에서 모니터를 점유 중이어야 `obj.notify()`를 실행할 수 있습니다. 이 스레드가 블록을 빠져나와야 모니터가 해제되고, 깨어난 스레드가 락을 획득할 수 있습니다.

---

### 5. notify()를 호출한 스레드는 어떻게 되나요?

1. notify()를 호출한 스레드는, 일반적으로 **`notify()` 코드가 속한 synchronized 블록**이 끝날 때까지 모니터를 계속 점유합니다.
2. 블록이 끝나는 시점(`}`)에 `monitorexit`가 실행되며, 모니터가 해제됩니다.
3. 그제서야 WAITING 중이던 스레드가 모니터 획득을 시도(BLOCKED → RUNNABLE)하여, 성공하면 `wait()` 다음 코드를 진행(`doSomething()`)합니다.

즉, **notify()**를 호출하는 순간 곧바로 모니터를 뺏기는 것이 아니라, “**synchronized 블록이 끝날 때까지**”는 계속 모니터를 유지합니다. 이는 **“모니터를 여러 스레드가 동시에 점유할 수 없다”**는 점과 일관됩니다.

---

### 6. `synchronized(lock)` 사용 시 메모리 장벽(Memory Barrier) 적용 여부

1. 자바에서 **synchronized** 블록은 **모니터 락**을 통해, 쓰레드가 블록에 진입할 때와 빠져나올 때 **메모리 장벽**(happens-before 관계)을 설정합니다.
2. 구체적으로, 블록에 진입하기 전까지 쓰레드가 캐시에 가지고 있던 값들이 모두 비워지고, 블록을 빠져나오기 전까지 수행된 writes가 메인 메모리에 반영됩니다.
3. 그 결과, `synchronized`가 보장하는 것은 “**락 해제 후 다른 스레드가 락을 잡았을 때, 전 스레드가 한 변경 사항을 볼 수 있다**”는 것입니다.
4. 따라서 `synchronized(lock)`로 감싼 로직은, 락을 통해 “원자적(atomic) + 가시성(visibility) 보장”이 됩니다. 이는 흔히 “**mutex + 메모리 배리어**”라고 말할 수 있습니다.

---

# 7. 결론 정리

1. `synchronized(LockObject)` 구문을 만나면, 스레드는 LockObject의 **모니터**를 획득해야 합니다. 만약 누군가 이미 락을 갖고 있으면 해당 스레드는 **BLOCKED** 상태가 됩니다.
2. `LockObject.wait()`는 “현재 쓰레드가 모니터를 자발적으로 놓고, wait set에서 WAITING 상태가 된다”는 의미입니다. 이때, doSomething() 같은 다음 로직은 잠시 멈추게 되고, **notify**나 **notifyAll**이 호출되어야만 깨어날 수 있습니다.
3. 깨어난 쓰레드는 다시 모니터를 획득하기 위해 **BLOCKED** 상태를 거칠 수 있고, 획득에 성공하면 RUNNABLE 상태로 돌아가 **doSomething()**을 실행합니다.
4. `notify()`와 `notifyAll()`은 자바가 자동으로 호출해주지 않으므로, **명시적으로** 호출해야 합니다. 보통 “상태 변화를 일으키는 쪽(예: 작업 추가, 특정 플래그 변경)”에서 알맞은 시점에 `notify()`를 호출합니다.
5. `synchronized`는 **메모리 장벽**을 제공하여, 락을 잡고 있는 동안 이뤄진 변경사항이 락 해제 후 다른 스레드에게 확실히 보이도록 합니다.
6. **실무/프로덕션 코드**에서는 “**Producer-Consumer** 패턴” 같은 큐 기반 구조에서 `wait()`/`notify()`를 자주 활용하며, TDD로 이 로직을 검증할 때는 **Thread + Sleep + assert** 같은 방식을 사용하거나, **CountDownLatch**, **CyclicBarrier**, **Awaitility** 등의 테스트 유틸을 함께 쓰기도 합니다.

**최종 결론**:
- 모든 쓰레드가 wait만 하고 notify를 아무도 안 해주면, 영원히 깨어나지 못하므로 **교착(deadlock) 비슷한 상태**가 생깁니다.
- `doSomething()`은 “대기(wait) → notify → 모니터 재획득” 순서가 성립해야 실행됩니다.
- `synchronized(락)`을 사용하면, 해당 락(모니터)에 대해 “상호 배타적 접근(mutex) + 쓰기/읽기의 가시성(메모리 장벽)”이 보장됩니다.

이상으로, **“모든 스레드가 wait를 반복할 것 같은 상황에서 언제 doSomething이 실행되는지, notify는 누가 어떻게 호출해야 하고, 그 호출 후의 흐름이 어떻게 이어지는지, 실전 프로덕션 코드 예시와 TDD 방식을 곁들여”** 자세히 살펴보았습니다.
