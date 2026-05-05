# 08. 분산 시스템과 아키텍처

MSA, consistency, availability, topology, saga, idempotency처럼 작은 기술 단위를 큰 시스템 설계로 조립하는 판단 축을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## MSA와 결제 시스템 설계

### 결제 MSA에서 결제 PG사 기능 구현

#### 원문: 결제 MSA에서 결제 PG사 기능 구현

<!-- curriculum-chunk: sha256=99df37e18a868fa1a65954dfeb8e07402d5bb84b7bf65cbea0dea80d43fb0dc6 major=distributed-systems-architecture mid=MSA와 결제 시스템 설계 sub=결제 MSA에서 결제 PG사 기능 구현 sources=source/interview_questions.md:3226-3247, source/interviews.md:3226-3247 -->

> Source: `source/interview_questions.md:3226-3247`
> Classification reason: service architecture
> Duplicate source aliases: `source/interview_questions.md:3226-3247, source/interviews.md:3226-3247`

##### 결제 MSA에서 결제 PG사 기능 구현

kotlin & spring & hikari pool & arrow-kt(함수형 라이브러리) 등을 사용하는 ddd(4 layered architecture)로 presentation, application, domain, infrastructure로 구성된 결제 담당 마이크로 서비스가 있습니다.
쿼리는 orm이 아닌 직접 쿼리를 코드에 작성한 방식으로 구현되어 있습니다.

결제 요청이 gRPC 요청이 들어오면 presentation 레이어에서 결제 요청 Command로 바뀌고, 애플리케이션 레이어에서 요청이 들어온 PG사의 결제 서비스를 선택하여 해당 command를 전달합니다.
이때 애플리케이션 레이어에서는 요청으로 들어온 PG사에 따라 구현된 애플리케이션 레이어의 인스턴스를 직접 맵핑하도록 구현됐습니다.
구현된 PG사들 enum이 정의되어 있고 when을 통해 exhaustiveness를 체크해서 누락 여부를 컴파일 타임에 알 수 있게 구현됐습니다.
애플리케이션 레이어에서는 도메인 레이어의 인터페이스를 구현한 인프라 레이어의 인스턴스를 생성자로 주입 받아서 갖고 있습니다.
따라서 이 인터페이스로 드러나 있는 구현체에 해당 커맨드를 다시 도메인 레이어에 정의된 결제 요청 DTO로 변환합니다.
그리고 인프라스트럭처의 구현체는 해당 PG사와 통신하고 그 결과를 리턴합니다.
애플리케이션 레이어에서는 결과를 저장하는 등 후처리를 하고, 프리젠테이션 레이어에서 필요한 형태로 데이터를 마련하여 리턴합니다.
프리젠테이션 레이어는 다시 gRPC 응답 인터페이스에 맞는 DTO로 변환해서 최종 응답합니다.
PG사와의 통신부에는 어쩔 수 없이 외부 client 라이브러리를 사용하므로 try ~ catch가 있을 수밖에 없지만, 클라이언트에서 발생하는 Exception을 제외하고는 모두 도메인에 에러들을 일일이 정의했습니다.
그래서 웹 통신이 정상적으로 이뤄졌더라도 결제가 성공 또는 실패할 수 있는데, `Either<Error, Success>` 로 실패 또는 성공 둘중 하나를 반드시 리턴하도록 구현됐습니다.
에러 핸들링은 애플리케이션 레이어에서 마찬가지로 했는데, 에러들은 sealed class로 선언되어서 반드시 when을 통해 exhaustiveness 체크를 하도록 했습니다. 즉, 결제 요청한 클라이언트로 어떻게 응답할지 애플리케이션 레이어에서 반드시 한번 개발자가 결정하도록 구현했습니다.

사실 처음에는 interface를 정의하고, 인터페이스 통해서 DI를 하도록 구현했었습니다.
하지만 나중에는 Spring 통해서 DI하는 게 아닌, 직접 맵핑을 하는 방식으로 리팩토링 됐습니다.
당시 CTO가 이 리팩토링을 진행했는데, 근거는 프레임워크가 많은 것을 해주지만, 그것들이 디버깅하기 어렵기 때문에, 손이 가더라도 최대한 가능한 한 많이 드러내서 개발자가 직접 제어할 수 있도록 하는 게 목표였습니다.
스프링의 어노테이션이나 DI 등 프레임워크가 제공하는 기능은 필요 최소한으로 사용했습니다.

<!-- /curriculum-chunk -->

## 가용성, 토폴로지, 확장성

### 5. 종합 아키텍처/적용 질문

#### 원문: 5. 종합 아키텍처/적용 질문

<!-- curriculum-chunk: sha256=28838cb61519cf42d3af9ce4ad80c7f37a375288db18e89e1bf7c983b1a921d5 major=distributed-systems-architecture mid=가용성, 토폴로지, 확장성 sub=5. 종합 아키텍처/적용 질문 sources=source/interview_questions3.md:213-236 -->

> Source: `source/interview_questions3.md:213-236`
> Classification reason: availability/topology

##### 5. 종합 아키텍처/적용 질문

###### 질문 11. Spring + Netty + MySQL로 고부하(High Traffic) 시스템을 만든다면 어떤 아키텍처 구성을 권장하시나요?

###### **답변**

1. **Netty**로 외부 연결을 처리할 경우, 대규모 클라이언트 연결을 효율적으로 관리할 수 있습니다. 예를 들어, 실시간 채팅 서버나 대규모 TCP 연결이 필요한 서비스에 적합합니다.
2. Netty에서 받은 데이터를 **비즈니스 로직** 처리로 넘길 때는, 스프링의 @Service나 @Component Bean으로 구성된 로직을 호출할 수 있습니다. 이때 Netty의 이벤트 루프는 **논블로킹** 방식이므로, 블로킹 연산(예: DB쿼리, 파일 I/O)은 별도의 Worker ThreadPool로 분리하는 것이 좋습니다.
3. DB로는 **MySQL**을 사용하되, 트랜잭션이 필요한 로직은 스프링의 @Transactional을 통해 관리할 수 있습니다. 동시 트랜잭션이 많다면, 커넥션 풀(HikariCP 등) 크기를 최적화해야 하며, MySQL 자체도 연결 수와 쿼리 부하를 감당할 수 있게 튜닝해야 합니다.
4. 더 나아가, 대규모 트래픽을 처리하려면 **로드밸런서**(예: L4 또는 L7)로 여러 서버 인스턴스로 트래픽을 분산하고, **MySQL Replication**(마스터-슬레이브 구조)나 **Sharding**을 통해 DB 부하를 분산하기도 합니다.
5. 애플리케이션 서버는 Spring Boot나 Spring Cloud 등의 생태계를 활용할 수 있으며, Netty 기반 마이크로서비스 간 통신을 gRPC 등으로 구현하면, 대규모 시스템에서 성능과 확장성을 동시에 잡을 수 있습니다.

---

###### 질문 12. Java 스레드풀과 Netty 이벤트 루프를 함께 사용할 때 주의점은 무엇인가요?

###### **답변**

1. Netty 이벤트 루프(이하 EL)는 **논블로킹 I/O**에 특화된 구조입니다. EL 스레드에서는 **Blocking 연산**(예: DB 쿼리, 동기식 HTTP 호출)을 수행하면 전체 성능이 급격히 떨어집니다.
2. 따라서, DB 처럼 블로킹 API를 호출해야 한다면, “별도의 쓰레드풀(Worker Pool)”을 구성하여 그곳에서 실행되도록 합니다. Netty 핸들러는 작업 요청을 EL에서 받아 큐에 넣고, 이후 Worker Pool에서 결과가 나오면 EL로 다시 결과를 전달하는 식의 구조를 가지면 됩니다.
3. Spring에서 `@Async`나 `TaskExecutor`를 설정하거나, 직접 `ThreadPoolExecutor`를 만들 수도 있습니다. 이렇게 하면 Netty EL은 I/O 이벤트에만 집중하고, 무거운 연산은 다른 곳에서 처리하므로 시스템 전반이 균형 있게 동작합니다.

---

<!-- /curriculum-chunk -->

### Distributed Systems에서 고가용성이란?

#### 원문: Distributed Systems에서 고가용성이란?

<!-- curriculum-chunk: sha256=acbb78758cd2d0774237bf20ddaa22cadb50216e8c08fc39b1ba1ff006956f11 major=distributed-systems-architecture mid=가용성, 토폴로지, 확장성 sub=Distributed Systems에서 고가용성이란? sources=source/interview_questions.md:2085-2322, source/interviews.md:2085-2322 -->

> Source: `source/interview_questions.md:2085-2322`
> Classification reason: availability/topology
> Duplicate source aliases: `source/interview_questions.md:2085-2322, source/interviews.md:2085-2322`

##### Distributed Systems에서 고가용성이란?

고가용성이란 시스템이 "약속된 서비스를 중단 없이 지속적으로 제공할 수 있는 능력"을 의미합니다.
이는 장애가 "없는" 것이 아니라, 장애가 발생하더라도 서비스가 "중단되지 않는" 상태를 말합니다.

$Availability = \frac{MTTF}{MTTF + MTTR}$

여기서:
- MTTF (Mean Time To Failure): 평균 고장 시간
- MTTR (Mean Time To Recovery): 평균 복구 시간

시스템의 가용성은 보통 "9의 개수"로 표현됩니다:
- 99% ("Two Nines"): 연간 87.6시간 다운타임
- 99.9% ("Three Nines"): 연간 8.76시간 다운타임
- 99.99% ("Four Nines"): 연간 52.56분 다운타임
- 99.999% ("Five Nines"): 연간 5.26분 다운타임

분산 시스템에서는 다음 세 가지 특성 중 동시에 두 가지만 만족할 수 있습니다:
1. 일관성(Consistency): 모든 노드가 동일한 시점에 동일한 데이터를 가짐
2. 가용성(Availability): 모든 요청이 성공 또는 실패 응답을 받음
3. 분할 허용성(Partition Tolerance): 네트워크 분할이 발생해도 시스템이 계속 동작

시스템에서 발생할 수 있는 다양한 유형의 장애:

1. Fail-Stop:
    - 노드가 완전히 중지
    - 다른 노드들이 장애를 감지 가능

2. Crash:
    - 노드가 갑자기 중지
    - 다른 노드들이 장애를 즉시 감지하지 못할 수 있음

3. Byzantine:
    - 노드가 임의의 잘못된 동작을 수행
    - 가장 복잡한 장애 유형

고가용성의 기본 원칙은 다음과 같습니다.
1. 단일 장애점 제거 (No Single Point of Failure)
    - 모든 핵심 컴포넌트는 이중화
    - 모든 중요 연결은 대체 경로 확보

2. 장애 격리 (Failure Isolation)
    - 장애의 전파 방지
    - 부분 장애가 전체 장애로 확대되는 것을 방지

3. 신속한 장애 감지와 복구
    - 자동화된 모니터링
    - 자동화된 복구 절차

4. 우아한 성능 저하 (Graceful Degradation)
    - 완전한 장애 대신 부분적 기능 저하 허용
    - 핵심 기능 유지를 위한 비핵심 기능 포기

현실 세계에서 다음과 같은 시나리오들이 가능합니다

1. 상점 시나리오: "편의점의 고가용성"

    ```mermaid
    graph TB
        A[정전] --> B{UPS 작동}
        B -->|Yes| C[기본 전력 공급]
        B -->|No| D{발전기 작동}
        D -->|Yes| E[백업 전력 공급]
        D -->|No| F[비상 조명만 가동]
    ```

    문제 상황: 전력 공급 중단
    - 냉장/냉동 시설 가동 중지 위험
    - POS 시스템 작동 중지 위험

    고가용성 솔루션

    1. 우아한 성능 저하
        - 1단계: UPS로 핵심 시스템만 유지
        - 2단계: 수동 계산기와 영수증으로 전환
        - 3단계: 현금 거래만 허용

    2. 비상 대응 절차

        ```python
        def handle_power_failure():
            if check_ups_status():
                run_critical_systems()
            elif check_generator_status():
                run_all_systems()
            else:
                run_emergency_mode()
        ```

2. 의료 시나리오: "중환자실의 고가용성"

    ```mermaid
    graph TD
        A[환자 상태 모니터링] --> B{주 시스템 장애}
        B -->|Yes| C[백업 시스템 활성화]
        B -->|No| D[정상 모니터링]
        C --> E{백업 시스템 장애}
        E -->|Yes| F[수동 모니터링]
        E -->|No| G[백업 모니터링]
    ```

    문제 상황: 생명 유지 장치의 연속성
    - 전원 공급 중단 위험
    - 모니터링 시스템 장애 위험

    자동 전환 시스템:

    ```java
    public class PatientMonitoring {
        private Monitor primaryMonitor;
        private Monitor backupMonitor;
        private ManualMonitoring manualSystem;

        public void handleMonitorFailure() {
            if (!primaryMonitor.isWorking()) {
                switchToBackup();
                alertMedicalStaff();
            }
        }

        private void switchToBackup() {
            if (backupMonitor.isWorking()) {
                activateBackupSystem();
            } else {
                activateManualMonitoring();
            }
        }
    }
    ```

3. 금융 시나리오: "ATM 네트워크의 고가용성"

    ```mermaid
    graph LR
        A[ATM] --> B[주 네트워크]
        A --> C[백업 네트워크]
        B --> D[주 서버]
        C --> E[백업 서버]
        D --> F[데이터베이스 1]
        E --> G[데이터베이스 2]
    ```

    문제 상황: 네트워크 연결 중단
    - 거래 처리 불가
    - 잔액 조회 불가

    오프라인 운영 모드

    ```java
    public class ATMOperation {
        private static final BigDecimal OFFLINE_LIMIT =
            new BigDecimal("100000");

        public Transaction processWithdrawal(Account account,
            BigDecimal amount) {
            if (isOnline()) {
                return processOnlineWithdrawal(account, amount);
            } else {
                return processOfflineWithdrawal(account, amount);
            }
        }

        private Transaction processOfflineWithdrawal(
            Account account, BigDecimal amount) {
            if (amount.compareTo(OFFLINE_LIMIT) > 0) {
                throw new LimitExceededException();
            }
            // 오프라인 거래 기록 저장
            return new Transaction(TransactionType.OFFLINE);
        }
    }
    ```

4. 교통 시나리오: "신호 체계의 고가용성"

    ```mermaid
    graph TD
        A[교통 신호] --> B{주 제어 시스템}
        B -->|정상| C[정상 운영]
        B -->|장애| D{백업 제어 시스템}
        D -->|정상| E[백업 운영]
        D -->|장애| F[독립 운영 모드]
    ```

    문제 상황: 신호 제어 시스템 장애
    - 교통 혼잡 발생
    - 사고 위험 증가

    다단계 백업 시스템:

    ```python
    def traffic_signal_control():
        if main_system.is_operational():
            return main_system.get_signal_pattern()
        elif backup_system.is_operational():
            return backup_system.get_signal_pattern()
        else:
            return default_timing_pattern()
    ```

    독립 운영 모드:
    - 미리 프로그래밍된 기본 신호 패턴 실행
    - 교통 경찰 수동 통제 전환

고가용성을 위해 다음과 같은 구현 원칙들이 존재합니다.

1. 중복성 설계 (Redundancy Design)

    ```python
    class HighAvailabilitySystem:
        def __init__(self):
            self.primary = Component()
            self.secondary = Component()
            self.status = "NORMAL"

        def process_request(self, request):
            try:
                return self.primary.process(request)
            except ComponentFailure:
                self.status = "DEGRADED"
                return self.secondary.process(request)
    ```

2. 장애 감지와 복구

    ```python
    def monitor_and_recover():
        while True:
            for component in system_components:
                if not component.is_healthy():
                    try:
                        component.restart()
                        log.info("Component restarted successfully")
                    except RecoveryFailure:
                        activate_backup(component)
            time.sleep(CHECK_INTERVAL)
    ```

<!-- /curriculum-chunk -->

### 토폴로지

#### 원문: 토폴로지

<!-- curriculum-chunk: sha256=66d745271071221c437ac922a4b648d4ca47e1912a178db31e49f730ad1478c8 major=distributed-systems-architecture mid=가용성, 토폴로지, 확장성 sub=토폴로지 sources=source/interview_questions.md:1675-1734, source/interviews.md:1675-1734 -->

> Source: `source/interview_questions.md:1675-1734`
> Classification reason: availability/topology
> Duplicate source aliases: `source/interview_questions.md:1675-1734, source/interviews.md:1675-1734`

##### 토폴로지

"토폴로지(Topology)"는 *구성 요소 간의 관계나 연결 방식*을 의미하는 용어로, 다양한 분야에서 사용됩니다.
특히, 수학, 네트워크, 시스템 설계 등에서 각기 다른 맥락에서 쓰이지만, 공통적으로 *요소들 간의 연결 구조*를 설명하는 데 초점이 맞춰져 있습니다.

- 수학적 토폴로지

    수학에서는 "토폴로지"가 공간 내의 점들이 어떻게 연결되고 배열되어 있는지에 대한 추상적 구조를 설명합니다.
    예를 들어, 공간에서의 점들의 "연결성"을 다루며, 기하학적 형태보다는 연결 구조 자체에 초점을 맞춥니다.

- 컴퓨터 네트워크의 토폴로지

    네트워크 토폴로지는 컴퓨터 네트워크의 물리적 또는 논리적 연결 구조를 나타냅니다.
    예를 들어, 버스형, 스타형, 링형 네트워크 등이 네트워크 토폴로지의 예입니다. 각 장치들이 어떻게 연결되고 데이터가 어떻게 흐르는지를 정의합니다.

- 시스템 설계에서의 토폴로지

    시스템 설계에서는 여러 구성 요소(하드웨어, 소프트웨어, 데이터 흐름 등)의 상호작용과 배치 방식을 나타냅니다.
    이는 시스템에서 정보나 신호가 어떻게 전달되고 처리되는지를 설명합니다.

Apache Heron에서 "토폴로지"는 실시간 데이터 스트리밍 애플리케이션의 처리 흐름을 나타내는 구조로, 데이터가 시스템 내에서 어떻게 처리되고 전달되는지를 정의합니다.
Heron의 토폴로지는 컴퓨터 네트워크나 시스템 설계에서 사용되는 개념과 유사하게 논리적 처리 단위 간의 연결 및 데이터 흐름을 의미합니다.

- 토폴로지 구성 요소:
    Heron에서 토폴로지는 스파우트(Spout)와 볼트(Bolt)라는 처리 단위로 구성됩니다.

    - 스파우트(Spout): 외부에서 데이터를 받아오는 역할을 합니다. 이는 데이터 스트림의 시작점입니다.
    - 볼트(Bolt): 데이터를 가공하고 처리하는 역할을 합니다. 예를 들어, 중복 제거, 이벤트 유효성 판단, 필터링, 변환 등의 작업을 수행할 수 있습니다.

- Directed Acyclic Graph (DAG) 구조:

    Heron의 토폴로지는 방향성 비순환 그래프(DAG)로 구성됩니다.
    즉, 데이터는 토폴로지 내에서 한 방향으로만 흐르며, 처리 과정 중 다시 이전 단계로 돌아가거나 무한 루프에 빠지지 않도록 설계됩니다.
    이 구조는 실시간 데이터 처리에서 매우 중요합니다.

Heron에서 토폴로지 형태로 구성한다는 것은 데이터 처리의 논리적 흐름을 단계별로 구성하고 이를 그래프 형태로 표현하는 것을 의미합니다.
다음과 같은 단계로 설명할 수 있습니다:

1. 스파우트(Spout)에서 데이터 입력:
    Heron 토폴로지는 스파우트에서 시작됩니다.
    스파우트는 외부 시스템(예: Kafka, 메시지 큐 등)으로부터 데이터를 가져와서 처리 흐름을 시작합니다.

2. 볼트(Bolt)에서 데이터 처리:
    데이터는 스파우트에서 볼트로 전달됩니다.
    각 볼트는 특정 작업을 수행하며, 여러 볼트들이 서로 연결되어 있어 데이터가 차례대로 처리됩니다.
    예를 들어, 첫 번째 볼트는 중복 제거 작업을 하고, 두 번째 볼트는 이벤트 유효성 판단 작업을 수행하는 식으로 여러 단계로 처리됩니다.

3. 데이터 흐름의 병렬 및 분기:
    Heron의 토폴로지는 각 단계에서 데이터를 병렬로 처리하거나 분기할 수 있도록 설계됩니다.
    즉, 하나의 데이터를 여러 볼트에서 병렬로 처리하거나, 특정 조건에 따라 다른 경로로 데이터를 분기할 수 있습니다.
    이는 대규모 데이터를 실시간으로 빠르게 처리하는 데 필수적인 구조입니다.

Heron을 이용해 토폴로지를 구성하는 방식은 다음과 같습니다:
- 스파우트가 외부에서 데이터를 받아오고(예: Kafka에서 메시지 수신),
- 첫 번째 볼트는 데이터의 중복을 확인하며, 중복된 데이터는 버리고 유효한 데이터만을 다음 단계로 넘깁니다.
- 두 번째 볼트는 데이터의 유효성을 판단합니다(예: 부정적 경로 여부를 확인).
- 세 번째 볼트는 과금형 이벤트를 판별하여 필요 시 특정 처리 로직을 적용합니다.

이 모든 단계가 DAG 형태의 토폴로지로 구성되며, 각 단계는 독립적이지만 논리적 흐름에 따라 연결되어 있습니다.

<!-- /curriculum-chunk -->

## 일관성과 분산 트랜잭션

### eventual consistency

#### 원문: eventual consistency

<!-- curriculum-chunk: sha256=dcbaffe8c46003c5588c7c6b02445e091dcc87a53447e63f194fb0be8aa6aa68 major=distributed-systems-architecture mid=일관성과 분산 트랜잭션 sub=eventual consistency sources=source/interview_questions.md:1662-1674, source/interviews.md:1662-1674 -->

> Source: `source/interview_questions.md:1662-1674`
> Classification reason: distributed consistency
> Duplicate source aliases: `source/interview_questions.md:1662-1674, source/interviews.md:1662-1674`

##### eventual consistency

Eventual consistency(also called optimistic replication)는 항목이 새롭게 업데이트되지 않는다는 전제하에 항목의 모든 읽기 작업이 최종적으로는 마지막으로 업데이트된 값을 반환한다는 것을 이론적으로 보장합니다.
인터넷 DNS(도메인 이름 시스템)는 eventual consistency 모델이 사용된 시스템의 예로 잘 알려져 있습니다.

반대되는 개념으로 Strong Consistency 가 있습니다. 관계형 데이터베이스가 대표적인 모델입니다.
관계형 데이터베이스에서 트랜잭션의 속성 ACID를 떠올려보면,
- Consistency - 트랜잭션이 완료되면 일관된 데이터를 유지해야합니다.
- 클라이언트가 서로 다른 데이터를 조회할 수 없습니다.
- 그렇기 때문에 서버간 데이터를 복사할 때 락킹 메커니즘이 동작하고, 가용성에 문제가 생기게 됩니다.

그러나 Eventaul Consistency는 동기화 전이라도 이를 노드에 접근을 허용하기 때문에 가용한 상태를 유지합니다.

<!-- /curriculum-chunk -->
