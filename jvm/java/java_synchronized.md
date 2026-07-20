# `synchronized` 범위 설계와 오용 회고

## 개요

`synchronized`는 잘못된 기능이 아니다. 공유 상태의 불변식을 짧고 분명한 임계 영역으로 보호할 때는 Java에서
가장 안전하고 읽기 쉬운 동기화 수단 중 하나다. 문제는 보호해야 할 상태보다 넓은 monitor를 선택하거나,
여러 값을 함께 판단해야 하는데 단일 변수만 Atomic으로 바꾸는 데서 생긴다.

이 문서는 다음 질문에 답한다.

- 인스턴스 `synchronized` 메서드는 정확히 무엇을 잠그는가?
- 같은 객체의 다른 메서드와 일반 필드 접근도 모두 멈추는가?
- 소스 코드는 class file과 JVM에서 어떻게 실행되는가?
- `synchronized`는 가시성, 원자성, 실행 순서를 어떻게 보장하는가?
- 넓은 메서드 잠금은 왜 불필요한 직렬화를 만드는가?
- 모든 필드를 Atomic으로 바꾸어도 복합 조건의 경쟁이 남는 이유는 무엇인가?
- `volatile`, Atomic, `synchronized`, `ReentrantLock`, `ReadWriteLock`은 언제 선택하는가?
- 정상 설계상 도달 불가능한 상태를 왜 방어해야 하는가?

핵심 판단은 세 문장으로 압축할 수 있다.

1. 인스턴스 `synchronized` 메서드는 그 메서드를 호출한 객체인 `this`의 monitor를 잠근다.
2. 같은 객체의 monitor를 사용하는 모든 동기화 영역은 서로 배타적이지만, 동기화하지 않은 메서드와 필드
   접근까지 자동으로 막지는 않는다.
3. 올바른 동기화 경계는 클래스나 메서드의 크기가 아니라 함께 지켜야 하는 불변식의 크기로 정한다.

## 가장 작은 예: 같은 객체의 세 동기화 영역

다음 세 영역은 모두 같은 `service` 객체의 monitor를 사용한다.

```java
final class ChannelService {
    synchronized void recordActivity() {
        // synchronized (this)와 같은 monitor
    }

    synchronized void requestStop() {
        // recordActivity와 동시에 실행될 수 없다.
    }

    void status() {
        synchronized (this) {
            // 위 두 메서드와 같은 monitor
        }
    }
}
```

스레드 A가 `recordActivity()`에 들어가면 스레드 B는 같은 객체의 `requestStop()`이나 `status()`의 동기화
블록에 들어갈 수 없다. 스레드 B는 monitor를 얻을 때까지 기다린다.

그러나 다음 호출은 자동으로 막히지 않는다.

```java
final class ChannelService {
    private int state;

    synchronized void requestStop() {
        state = 1;
    }

    int unsafeState() {
        return state; // 같은 monitor를 획득하지 않는다.
    }
}
```

`requestStop()`이 실행 중이라고 해서 `unsafeState()`가 멈추는 것은 아니다. 객체를 잠근다는 표현은 객체의
모든 접근을 봉쇄한다는 뜻이 아니다. 정확한 뜻은 "그 객체의 monitor를 얻으려는 다른 동기화 영역과 상호
배제된다"이다. JLS도 monitor 획득이 동기화되지 않은 필드 접근이나 메서드 호출을 막지 않는다고 명시한다.

이 차이를 놓치면 두 가지 상반된 실수가 생긴다.

- 필요 없는 메서드까지 같은 monitor에 묶어 처리량을 낮춘다.
- 일부 메서드만 `synchronized`로 만들고 객체 전체가 안전해졌다고 오해한다.

## 인스턴스, 클래스, 별도 객체 monitor는 서로 다르다

### 인스턴스 메서드

```java
public synchronized void update() {
    // lock: this
}
```

인스턴스 메서드는 `this`의 monitor를 사용한다. 인스턴스가 두 개라면 monitor도 서로 다르므로 두 호출은 동시에
실행될 수 있다.

```java
ChannelService first = new ChannelService();
ChannelService second = new ChannelService();

// first.update()와 second.update()는 서로 다른 monitor를 사용한다.
```

Spring singleton처럼 모든 요청이 한 인스턴스를 공유한다면 여러 요청이 같은 `this` monitor로 모인다. 같은
singleton의 `synchronized` 인스턴스 메서드는 이름과 역할이 달라도 서로 배타적이다.

### 정적 메서드

```java
public static synchronized void updateGlobalState() {
    // lock: ChannelService.class
}
```

정적 `synchronized` 메서드는 인스턴스가 아니라 해당 클래스의 `Class` 객체 monitor를 사용한다.
`service.update()`와 `ChannelService.updateGlobalState()`는 서로 다른 monitor를 사용하므로 자동으로 상호
배제되지 않는다.

### 별도의 비공개 lock 객체

```java
final class ChannelService {
    private final Object lifecycleLock = new Object();

    void requestStop() {
        synchronized (lifecycleLock) {
            // lifecycle 상태만 보호한다.
        }
    }
}
```

`synchronized (this)`는 lock의 정체를 외부에 노출한다. 외부 코드도 `synchronized (service)`를 실행할 수 있어
뜻하지 않은 간섭이나 lock 순서 역전을 만들 수 있다. intrinsic lock이 필요하지만 객체 전체를 공개된 lock으로
삼을 이유가 없다면 `private final` lock 객체가 경계를 더 분명하게 만든다.

별도 lock 객체가 언제나 더 좋은 것은 아니다. 보호 대상이 객체 전체의 단일 불변식이고 모든 접근이 같은 규칙을
따른다면 `synchronized` 메서드가 더 단순하고 실수할 지점도 적다. lock을 나누기 전에 어떤 필드가 반드시 함께
보여야 하는지 먼저 적어야 한다.

## 소스에서 JVM까지: 실제로 무엇이 실행되는가

이 저장소의
[`SynchronizedBytecodeExample.java`](../examples/synchronized/SynchronizedBytecodeExample.java)는 같은 monitor를
사용하는 메서드와 블록을 나란히 보여 준다.

```java
public synchronized void instanceMethod() {
    value++;
}

public void instanceBlock() {
    synchronized (this) {
        value++;
    }
}

public static synchronized void staticMethod() {
}
```

Java 25에서 다음 명령으로 class file을 확인할 수 있다.

```powershell
$out = Join-Path $env:TEMP 'study-synchronized-bytecode'
javac -d $out jvm/examples/synchronized/SynchronizedBytecodeExample.java
javap -classpath $out -c -v SynchronizedBytecodeExample
```

관찰해야 할 차이는 두 가지다.

1. `instanceMethod()`와 `staticMethod()`에는 class file의 `ACC_SYNCHRONIZED` method flag가 붙는다.
2. `instanceBlock()`의 bytecode에는 `monitorenter`와 정상·예외 경로의 `monitorexit`가 나타난다.

동기화 메서드 본문에는 보통 `monitorenter`가 보이지 않는다. JVMS는 `ACC_SYNCHRONIZED`가 붙은 메서드를
호출할 때 JVM이 monitor 진입과 이탈을 암묵적으로 처리하도록 규정한다. 동기화 블록은 컴파일러가
`monitorenter`와 `monitorexit` 명령으로 변환한다.

예외가 발생해도 monitor는 풀려야 한다. 동기화 블록의 bytecode에는 예외 경로에서도 `monitorexit`를 실행하기
위한 exception table이 생성된다. 동기화 메서드는 정상 반환과 비정상 종료 모두에서 JVM이 monitor를 해제한다.

### monitor의 소유권과 재진입

각 Java 객체에는 논리적으로 monitor가 연결된다. 한 시점에 한 스레드만 그 monitor를 소유할 수 있다. 다른
스레드가 이미 소유한 monitor에 진입하려 하면 소유권을 얻을 때까지 기다린다.

같은 스레드는 자신이 이미 가진 monitor를 다시 획득할 수 있다. 이를 재진입(reentrancy)이라고 한다. JVM은
재진입 횟수를 세고, 획득한 횟수만큼 해제된 뒤에야 다른 스레드가 monitor를 얻을 수 있게 한다.

```java
synchronized (lock) {
    synchronized (lock) {
        // 같은 스레드이므로 deadlock 없이 재진입한다.
    }
}
```

### monitor 구현 세부는 규격과 분리한다

JLS와 JVMS가 보장하는 것은 상호 배제, 재진입, 정상·비정상 종료 시 해제, 메모리 순서다. 객체 헤더의 구체적인
비트 배치, spinning 횟수, OS mutex나 park로 넘어가는 조건은 JVM 구현과 버전에 따라 달라질 수 있다.

HotSpot 같은 JVM은 경쟁이 없는 monitor의 비용을 줄이기 위해 여러 최적화를 적용할 수 있다. 그렇더라도 두
스레드가 같은 monitor를 요구하면 의미상 한 스레드만 임계 영역을 실행한다. "JVM이 lock을 최적화한다"는 말은
불필요한 직렬화가 사라진다는 뜻이 아니다.

## `synchronized`가 제공하는 세 가지와 제공하지 않는 것

### 상호 배제

같은 monitor를 사용하는 임계 영역에는 한 스레드만 들어간다. `count++`처럼 읽기, 계산, 쓰기로 나뉘는 연산도
같은 monitor 안에서 수행하면 다른 동일 monitor 사용자와 섞이지 않는다.

### 메모리 가시성과 순서

JMM에서 한 monitor의 unlock은 이후 같은 monitor의 lock보다 happens-before 관계에 있다. 먼저 임계 영역을
끝낸 스레드의 쓰기는 나중에 같은 monitor를 획득한 스레드에게 보이도록 연결된다.

이 보장은 단순히 "CPU cache를 비운다"라는 한 문장으로 설명할 수 없다. 컴파일러와 CPU는 단일 스레드 의미를
깨지 않는 범위에서 명령을 재배치할 수 있다. JMM의 happens-before 규칙은 다른 스레드가 어떤 쓰기를 관찰해야
하는지 언어 차원에서 정한다. JVM은 그 규칙을 만족하도록 machine instruction과 compiler barrier를 선택한다.

### 재진입

같은 스레드는 같은 monitor를 여러 번 획득할 수 있다. 동기화 메서드가 같은 객체의 다른 동기화 메서드를
호출해도 그것만으로 self-deadlock이 발생하지 않는다.

### 제공하지 않는 것

`synchronized`는 다음을 자동으로 제공하지 않는다.

- 다른 monitor를 사용하는 코드와의 상호 배제
- 동기화하지 않은 접근의 안전성
- 여러 객체에 걸친 transaction
- 공정한 lock 획득 순서
- timeout이나 interrupt 가능한 lock 획득
- deadlock 방지
- 임계 영역 내부 I/O의 빠른 완료
- 보호 대상보다 넓은 lock 범위에서의 좋은 처리량

## 회고 사례: 메서드 전체를 잠그면 왜 문제가 되는가

상시 요청을 받는 singleton 서비스에 다음 메서드가 있다고 가정한다.

```java
final class ChannelManager {
    synchronized void recordActivity() {
        lastActivityNanos = System.nanoTime();
    }

    synchronized void beginDrain() {
        state = DRAINING;
    }

    synchronized String currentStatus() {
        return state.name();
    }

    synchronized void closeSomeConnections() {
        // 연결 목록을 읽고 일부를 닫는다.
    }
}
```

작성 의도는 이해할 수 있다. 상태 전이를 한 번만 수행하고, 최근 활동 시각을 안전하게 공개하며, 종료 명령끼리
충돌하지 않게 하려는 것이다. 그러나 실제 lock 범위는 의도보다 넓다.

### 일반 요청과 관리 작업이 같은 monitor에서 경쟁한다

`recordActivity()`는 모든 요청이 지나는 hot path다. 같은 singleton의 다른 동기화 메서드와 monitor를 공유하면
서로 관계없는 상태 조회나 연결 정리가 활동 시각 기록을 막을 수 있다. 반대로 요청이 몰리면 관리 작업도 같은
monitor를 기다린다.

메서드가 짧으면 한 번의 대기 시간은 작을 수 있다. 하지만 문제는 요청 수만큼 lock 획득 경쟁이 반복된다는
점이다. 처리량과 tail latency는 평균 임계 영역 길이뿐 아니라 경쟁 빈도와 기다리는 스레드 수에도 영향을 받는다.

### 잠금 범위가 코드 구조에 숨는다

`synchronized`가 method modifier에 붙으면 그 메서드 전체가 임계 영역이다. 나중에 로그, collection 순회,
callback, 네트워크 I/O가 추가돼도 lock 범위가 자동으로 넓어진다. 리뷰어는 어떤 문장이 실제 불변식 때문에
보호되는지 구별하기 어려워진다.

### 한 monitor가 안전하다는 착시를 만든다

모든 메서드에 `synchronized`를 붙여도 framework가 내부에서 변경하는 queue, connection registry, executor
작업까지 같은 transaction이 되지는 않는다. 그 객체들은 자신의 동시성 규칙을 가진다. 애플리케이션 monitor는
외부 객체의 상태 변화와 원자적으로 묶이지 않는다.

따라서 "manager를 잠갔으니 channel 전체가 잠겼다"는 결론은 틀리다. 실제로 함께 보호되는 것은 같은 monitor를
획득하는 코드뿐이다.

## 반대쪽 실패: 전부 Atomic으로 바꾸면 충분한가

넓은 lock을 제거한 뒤 각 필드를 Atomic으로 바꾸면 개별 읽기와 쓰기는 안전해진다.

```java
AtomicReference<State> state = new AtomicReference<>(SERVING);
AtomicLong lastActivity = new AtomicLong();
```

다음 전이는 CAS로 중복 실행을 막을 수 있다.

```java
state.compareAndSet(DRAINING, STOPPING);
```

그러나 Atomic은 기본적으로 한 변수의 원자적 연산을 제공한다. "최근 활동이 충분히 오래 없었고, queue가 비어
있고, 상태가 DRAINING이면 STOPPING으로 바꾼다"는 조건은 세 상태를 묶은 불변식이다.

다음 실행 순서를 보자.

```text
종료 스레드: lastActivity가 오래됐다고 읽음
종료 스레드: queue size가 0이라고 읽음
요청 스레드: 새 요청의 lastActivity를 기록함
종료 스레드: DRAINING -> STOPPING CAS 성공
요청 스레드: queue에 요청을 넣음
```

각 Atomic 연산과 queue 연산은 자체적으로 thread-safe일 수 있다. 그래도 여러 연산 사이의 의미적 원자성은
생기지 않는다. 상태 CAS는 `state`가 예상 값이었다는 사실만 검증한다. `lastActivity`와 queue가 그사이에
변하지 않았다는 사실은 검증하지 않는다.

이것이 메모리 안전과 프로토콜 안전의 차이다.

- 메모리 안전: 값이 찢어지거나 손실되지 않고 각 연산이 규정된 순서로 보인다.
- 프로토콜 안전: 여러 값으로 표현한 업무 규칙이 하나의 유효한 상태 전이로 지켜진다.

단일 변수 Atomic을 여러 개 둔다고 복합 불변식이 자동으로 원자화되지는 않는다.

## 더 나은 경계: 일반 요청은 병렬, 최종 전환만 배타

종료 중에도 요청을 계속 처리해야 하고, 최종 전환 시점에만 신규 인계를 막아야 하는 서비스라면 read-write
admission 경계를 사용할 수 있다.

```java
final class AdmissionGate {
    private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();
    private final Lock admission = lock.readLock();
    private final Lock finalization = lock.writeLock();

    boolean tryBeginRequest() {
        admission.lock();
        if (state == STOPPING) {
            admission.unlock();
            return false;
        }
        lastActivityNanos.accumulateAndGet(System.nanoTime(), Math::max);
        return true;
    }

    void endRequestHandoff() {
        admission.unlock();
    }

    boolean tryFinalizeDrain() {
        if (!finalization.tryLock()) {
            return false;
        }
        try {
            if (state != DRAINING || !isQuietNow() || !queue.isEmpty()) {
                return false;
            }
            state = STOPPING;
            return true;
        } finally {
            finalization.unlock();
        }
    }
}
```

요청 경로는 read lock을 공유하므로 여러 요청이 동시에 진행할 수 있다. 종료 스레드는 write lock을 얻은 뒤
quiet와 queue를 다시 확인한다. write lock을 가진 동안에는 새 요청이 admission과 queue handoff 사이로
끼어들 수 없다.

실제 구현에서는 `tryBeginRequest()`와 `endRequestHandoff()`를 반드시 `try-finally`로 짝지어야 한다.

```java
if (!gate.tryBeginRequest()) {
    return reject();
}
try {
    return enqueueAndAcknowledge(request);
} finally {
    gate.endRequestHandoff();
}
```

이 설계가 항상 정답은 아니다. 요청 임계 영역이 길거나 read/write 비율이 다르다면 Atomic 상태 객체,
`Semaphore`, actor/event loop, framework lifecycle 기능이 더 적합할 수 있다. 중요한 것은 lock 종류부터 고르는
것이 아니라 종료 판단과 신규 요청 인계 사이의 불변식을 먼저 적는 것이다.

## STOPPING 이후 요청은 도달 불가능해야 하는데 왜 방어하는가

정상 배포 계약이 다음과 같다고 하자.

1. 서버가 readiness를 503으로 바꾼다.
2. L4가 health check 결과를 반영해 서버를 대상에서 제외한다.
3. 서버는 충분히 긴 quiet 기간 동안 신규 연결과 요청이 없음을 확인한다.
4. 서버가 STOPPING으로 전환하고 종료한다.

이 계약이 모든 계층에서 정확히 지켜진다면 STOPPING 이후 애플리케이션 handler에 도착하는 요청은 없어야 한다.
코드에서도 이를 정상 분기로 취급하면 안 된다. STOPPING 이후 거부 횟수가 반복적으로 관측된다면 방어 로직이
잘 작동했다는 뜻보다 L4 제외 계약이나 직접 접속 경로가 깨졌다는 신호에 가깝다.

그럼에도 방어는 필요하다. "도달 불가능"은 애플리케이션 상태 머신의 전제이지 네트워크 전체가 제공하는 원자적
보장이 아니기 때문이다. 다음 조건에서는 경계 요청이 도착할 수 있다.

- L4 health check에는 주기와 연속 실패 임계치가 있어 503과 제외 사이에 시간이 있다.
- 여러 L4 노드가 같은 health 상태를 서로 다른 시점에 관찰할 수 있다.
- L4가 이미 backend를 선택했거나 전달을 시작한 요청이 health 상태 변경 뒤에 도착할 수 있다.
- health check가 신규 연결 선택만 막고 기존 keep-alive 또는 장기 TCP 연결은 그대로 둘 수 있다.
- 클라이언트, 운영 도구, 같은 네트워크의 다른 서비스가 L4를 우회해 backend에 직접 접속할 수 있다.
- OS accept queue, socket receive buffer, framework executor queue에 이미 들어온 작업이 스레드 scheduling이나
  stop-the-world pause 뒤에 늦게 handler를 실행할 수 있다.
- readiness 포트의 상태 변경과 데이터 포트의 요청 수신은 하나의 분산 transaction이 아니다.
- 잘못된 L4 설정, health endpoint 오연결, 배포 자동화 순서 오류가 정상 전제를 깨뜨릴 수 있다.

따라서 방어 경로의 역할은 정상 요청을 처리하는 것이 아니다. 이미 STOPPING으로 확정된 상태를 되돌리거나 새
작업을 queue에 넣지 않고, 프로토콜이 허용한다면 실패 응답을 시도한 뒤 종료를 계속해야 한다. 응답 전송 자체도
프로세스 종료와 경쟁할 수 있으므로 best-effort일 수 있다. 최종 안전장치는 새 작업을 수락하지 않는 것이다.

운영 관측도 구분해야 한다.

- 0건: 정상적인 L4 제외와 quiet 판정이 작동했다.
- 드문 1건: in-flight 전달이나 scheduling 경계일 수 있으므로 요청 시각과 연결 생성 시각을 조사한다.
- 반복 발생: 직접 접속, 기존 연결 정책, health check 설정, L4 노드 간 반영 지연을 점검한다.

도달 불가능 상태를 방어한다는 이유로 정상 설계를 느슨하게 만들면 안 된다. 방어 코드는 조용히 요청을 계속
처리하는 fallback이 아니라, 전제 위반을 드러내고 더 큰 상태 훼손을 막는 마지막 경계다.

## 도구 선택 기준

### `final`과 불변 객체

생성 뒤 바뀌지 않는 값은 먼저 불변으로 만든다. constructor에서 설정한 `final` 필드는 정상적으로 생성된 객체가
공개될 때 특별한 JMM 초기화 보장을 받는다. 변경이 없으므로 lock이나 CAS가 필요하지 않다.

적합한 경우는 설정 snapshot, stateless service의 협력 객체 참조, 요청별 immutable value object다.

### `volatile`

한 필드의 최신 값 공개가 필요하고, 갱신이 단순 대입이며, 기존 값을 읽어 계산한 뒤 다시 쓰는 복합 연산이 없을
때 적합하다.

```java
private volatile boolean stopping;
```

`volatile`은 가시성과 volatile 순서를 제공하지만 다음 연산을 원자화하지 않는다.

```java
volatile int count;
count++; // read -> add -> write이므로 경쟁 시 증가를 잃을 수 있다.
```

### Atomic 클래스와 CAS

하나의 변수에 대한 read-modify-write나 조건부 상태 전환이 필요할 때 적합하다.

```java
if (state.compareAndSet(SERVING, DRAINING)) {
    publishReadiness503();
}
```

첫 번째 스레드만 전환에 성공하므로 중복 drain 시작을 막을 수 있다. Java SE 25의 Atomic 클래스는 해당 단일
변수에 대해 VarHandle 기반의 원자적 접근과 명시된 memory ordering을 제공한다.

다음 경우에는 단일 Atomic만으로 충분하지 않다.

- 서로 다른 Atomic 변수와 collection 상태를 한 snapshot으로 판단해야 한다.
- CAS retry 안에서 I/O나 부수 효과가 발생한다.
- 상태 전이가 너무 복잡해 허용 전이를 읽기 어렵다.

여러 필드가 항상 함께 바뀌어야 한다면 immutable state record 하나를 `AtomicReference<StateSnapshot>`으로
관리하는 방법도 있다. 단, state 바깥의 queue나 socket 상태까지 자동으로 묶이지는 않는다.

### 좁은 `synchronized` 블록

하나의 명확한 불변식을 짧은 임계 영역으로 보호하고 timeout, interruptible acquisition, 여러 condition이 필요
없다면 좋은 기본 선택이다. Java가 정상·예외 경로의 monitor 해제를 보장하므로 명시적 unlock 누락 위험도 없다.

```java
synchronized (lifecycleLock) {
    if (state == SERVING) {
        state = DRAINING;
        drainStartedAt = clock.nanoTime();
    }
}
```

### `ReentrantLock`

`synchronized`와 같은 배타 lock이 필요하지만 다음 기능이 필요할 때 선택한다.

- `tryLock()` 또는 timeout
- lock 대기 중 interrupt 처리
- 여러 `Condition`
- method 경계를 넘어선 명시적 lock protocol
- 필요성이 검증된 fairness 정책

항상 `try-finally`로 해제해야 한다.

```java
lock.lock();
try {
    updateState();
} finally {
    lock.unlock();
}
```

단지 `synchronized`보다 고급으로 보인다는 이유로 바꾸면 코드만 복잡해진다. JDK 24부터는 일반적인
`synchronized` 대기 때문에 virtual thread가 carrier thread에 고정되던 문제도 JEP 491로 거의 제거됐다.
Java 24 이상에서의 선택은 pinning 회피보다 필요한 기능과 lock 범위에 근거해야 한다.

### `ReadWriteLock`

여러 reader가 동시에 진행해도 되고 writer만 배타적이어야 할 때 사용한다. 여기서 `read`와 `write`는 반드시
데이터 조회와 수정을 뜻하지 않는다. admission gate처럼 `공유 진입`과 `배타 진입`의 의미로도 쓸 수 있다.
공유 작업이 실제로 충분히 많고 임계 영역이 lock 관리 비용을 상쇄할 만큼 의미가 있을 때 이점이 있다.

admission과 finalization 사례에서는 "일반 요청은 함께 진행해도 되지만 최종 상태 전환은 요청 인계와 겹치면 안
된다"는 불변식이 read/write 구조와 직접 맞는다.

read-write lock도 만능은 아니다. writer starvation, fairness에 따른 처리량 저하, read에서 write로 upgrade할 때의
deadlock 가능성을 검토해야 한다. `ReentrantReadWriteLock`에서는 read lock을 가진 상태로 write lock을 얻는
upgrade가 지원되지 않는다.

### 동시성 collection과 framework 상태

`ConcurrentHashMap`, blocking queue, framework connection registry는 자체 동시성 계약을 가진다. 바깥에서
임의의 lock을 하나 더 둔다고 여러 메서드 호출이 하나의 transaction이 되지는 않는다.

```java
if (queue.isEmpty()) {
    stop();
}
```

`isEmpty()`가 thread-safe여도 검사 직후 다른 스레드가 `offer()`할 수 있다. 검사와 상태 전환 사이를 함께
보호하거나, 신규 admission을 닫은 뒤 queue를 비우는 protocol이 필요하다.

## 성능을 JVM 최적화만으로 설명하면 안 되는 이유

lock 성능에는 적어도 다음 요소가 함께 작용한다.

- 경쟁 없는 monitor 획득 비용
- 같은 cache line을 여러 core가 갱신할 때의 coherence traffic
- 경쟁 시 스레드 대기와 깨우기
- 임계 영역 길이와 lock 획득 빈도
- lock convoy와 tail latency
- virtual thread와 platform thread의 scheduling

JDK 24의 JEP 491 이후 virtual thread는 `synchronized`에서 막힐 때도 대부분 carrier에서 unmount할 수 있다.
하지만 같은 monitor의 임계 영역이 한 번에 한 스레드만 실행된다는 의미는 그대로다. carrier pinning 제거는
불필요한 직렬화 제거가 아니다.

반대로 Atomic도 무조건 싸지 않다. 경쟁이 심한 CAS loop는 반복 실패와 cache line 이동을 만들 수 있다. 복합
상태를 억지로 lock-free로 만들면 코드 검증 비용이 lock 비용보다 커질 수 있다. 성능 판단은 실제 workload에서
Java Flight Recorder, profiler, 부하 테스트로 확인해야 한다.

## 실무 설계 순서

동기화 도구부터 고르지 말고 다음 순서로 결정한다.

1. 공유되는 변경 가능 상태를 찾는다.

    `final`과 지역 변수는 제외한다. 어떤 스레드가 읽고 쓰는지 호출 경로로 확인한다.

2. 불변식을 문장으로 쓴다.

    `state는 중복 전환하지 않는다`는 단일 변수 조건이다. `quiet이고 queue가 비어 있을 때만 stop한다`는 여러
    상태를 묶은 조건이다.

3. 원자적이어야 하는 최소 경계를 정한다.

    전체 메서드가 아니라 검사와 갱신 중 다른 스레드가 끼어들면 안 되는 구간을 찾는다.

4. 임계 영역 안의 blocking 작업과 callback을 제거한다.

    네트워크 I/O, 파일 I/O, 외부 callback, 오래 걸리는 collection 순회는 가능하면 lock 밖에서 수행한다.

5. 가장 단순하게 불변식을 지키는 도구를 고른다.

    불변이면 `final`, 단순 공개면 `volatile`, 단일 조건부 전환이면 Atomic, 짧은 복합 불변식이면 좁은
    `synchronized`, 추가 lock 기능이 필요하면 `Lock`을 고려한다.

6. 정상 경로와 방어 경로를 분리한다.

    도달 불가능해야 하는 상태가 반복되면 조용히 정상 처리하지 않는다. 실패를 관측 가능하게 만들고 상위 계약을
    조사한다.

7. 결정적 경합 테스트를 만든다.

    우연히 race가 발생하기를 기다리지 않는다. latch나 barrier로 두 스레드의 실행 순서를 고정해 "요청 인계 중
    종료 전환이 진행되지 않는다" 같은 불변식을 검증한다.

## 리뷰 체크리스트

- 어떤 객체의 monitor를 잠그는지 한 문장으로 말할 수 있는가?
- 같은 monitor를 사용하는 모든 코드 위치를 찾았는가?
- non-`synchronized` 접근이 남아 있지 않은가?
- lock이 보호하는 필드와 불변식이 명시돼 있는가?
- method 전체가 아니라 더 작은 임계 영역으로 줄일 수 있는가?
- 임계 영역 안에서 I/O, sleep, callback, context close를 수행하지 않는가?
- 여러 Atomic을 사용하면서 복합 snapshot이 안전하다고 오해하지 않았는가?
- queue의 `isEmpty()`나 `size()`와 다음 상태 전환 사이에 producer가 끼어들 수 있는가?
- `tryLock`, timeout, interrupt, condition, read/write 병렬성이 실제로 필요한가?
- lock 순서가 둘 이상이면 전역 순서가 정해져 있는가?
- 정상적으로 불가능한 상태의 방어가 상위 계약 오류를 숨기지 않는가?
- 경합 테스트가 sleep의 우연이 아니라 latch나 barrier로 순서를 통제하는가?
- Java 버전별 virtual thread pinning 설명을 현재 버전의 사실과 섞지 않았는가?

## 스스로 확인할 질문

1. 두 `synchronized` 인스턴스 메서드가 서로 막히는 기준은 클래스가 같은 것인가, 객체가 같은 것인가?
2. `static synchronized` 메서드는 어떤 객체의 monitor를 사용하는가?
3. `synchronized (this)` 중에도 다른 스레드가 일반 getter를 호출할 수 있는 이유는 무엇인가?
4. `volatile int count`의 `count++`가 안전하지 않은 이유는 무엇인가?
5. `AtomicReference<State>`의 CAS가 queue size의 불변성까지 보장하지 않는 이유는 무엇인가?
6. read lock을 여러 요청이 공유하고 write lock이 종료 전환을 담당하면 어떤 race를 막을 수 있는가?
7. STOPPING 이후 요청이 한 건 관측됐다면 코드보다 먼저 어떤 네트워크·운영 전제를 확인해야 하는가?
8. JDK 25에서 `synchronized`가 virtual thread를 대부분 pin하지 않는다는 사실이 넓은 lock 범위를 정당화하지
   못하는 이유는 무엇인가?

## 근거와 추가 읽기

- [JLS 8.4.3.6 - `synchronized` Methods](https://docs.oracle.com/javase/specs/jls/se25/html/jls-8.html#jls-8.4.3.6)
- [JLS 14.19 - The `synchronized` Statement](https://docs.oracle.com/javase/specs/jls/se25/html/jls-14.html#jls-14.19)
- [JLS 17.1 - Synchronization](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.1)
- [JLS 17.4.5 - Happens-before Order](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.4.5)
- [JLS 17.5 - `final` Field Semantics](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.5)
- [JVMS 2.11.10 - Synchronization](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-2.html#jvms-2.11.10)
- [JVMS 6.5 - `monitorenter`](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-6.html#jvms-6.5.monitorenter)
- [JVMS 6.5 - `monitorexit`](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-6.html#jvms-6.5.monitorexit)
- [Java SE 25 `AtomicReference`](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/atomic/AtomicReference.html)
- [Java SE 25 `ReadWriteLock`](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/locks/ReadWriteLock.html)
- [Java SE 25 `ReentrantReadWriteLock`](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/locks/ReentrantReadWriteLock.html)
- [OpenJDK JEP 491 - Synchronize Virtual Threads without Pinning](https://openjdk.org/jeps/491)
