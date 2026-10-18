# Spring 전파 관측 화면 사용성 시험 보고서

**참가자 역할**: 3년 차 백엔드 개발자 (Spring Boot, `@Transactional`, HikariCP, PostgreSQL 환경 개발 경험)

---

## 1. 열기 전 질문 2~3개

1. **질문 1**: 동료가 Spring 관련이라며 링크를 줬는데, 내가 실무에서 자주 쓰는 `@Transactional` 전파 옵션(`REQUIRED`, `REQUIRES_NEW`)이 실제로 내부에서 어떻게 굴러가는지 보여주는 화면일까?
2. **질문 2**: HikariCP 커넥션 풀과 DB 커넥션이 스프링 트랜잭션과 어떻게 엮여서 돌아가는지(스레드에 커넥션이 언제 묶이고 언제 풀리는지) 시각적으로 볼 수 있을까?
3. **질문 3**: 내가 직접 파라미터나 설정을 바꿔서 테스트해 볼 수 있는 대화형 시뮬레이터일까, 아니면 미리 실행해 둔 트레이스를 분석하는 뷰어일까?

---

## 2. 첫인상(60초)과 첫 세 번의 조작

### 첫 60초 동안의 첫인상
페이지가 열리자마자 가장 먼저 상단의 **"Spring 전파"** 타이틀과 시나리오 버튼들(`REQUIRED · 안쪽이 실패`, `REQUIRES_NEW`, `REQUIRES_NEW · 같은 행`, `REQUIRES_NEW · 풀 1칸`, `NESTED · 안쪽이 실패`, `0.5초 안에 다시 빌림`)이 눈에 들어왔습니다. 

화면 중앙에는 **앱 프로세스(main 스레드)**, **Spring 트랜잭션**, **JdbcTemplate**, **HikariCP(최대 2칸, C1 빌려줌)**, **pgjdbc**, **PostgreSQL 18.6 컨테이너(C1, C2)**까지 컴포넌트 구조가 블록 다이어그램으로 한눈에 펼쳐져 있었고, 오른쪽에는 실제 자바 코드(`Inner.java`, `Outer.java`)와 호출 스택이, 아래쪽에는 시간축(+ms)에 따른 호출 흐름과 DB 프로토콜 패킷이 시퀀스 다이어그램처럼 빼곡하게 그려져 있었습니다.

"아, 이건 단순히 개념 설명 글이나 가짜 모식도가 아니라, 실제 자바 프로세스와 도커의 PostgreSQL 사이에서 일어난 메서드 호출(JFR)과 네트워크 패킷을 통째로 캡처해 둔 초정밀 관측 도구구나!"라는 생각이 들었습니다.

가장 먼저 눈길이 간 곳은 상단 요약문의 `결과(실행 a) ... ROLLBACK ... 앱이 받은 예외: outer-caught IllegalStateException · outer-threw UnexpectedRollbackException`이었습니다. 평소 업무에서 "바깥 메서드에서 try-catch로 예외를 잡았는데 왜 롤백되고 UnexpectedRollbackException이 터지지?" 하고 답답했던 적이 있었기 때문에, 이 원인을 보여주는 주요 순간부터 확인해 보고 싶었습니다.

### 첫 세 번의 조작

- **조작 1: 주요 순간의 `+822 롤백 전용 표시` 버튼 클릭**
  - **이유**: 바깥 메서드가 예외를 잡았는데 왜 결국 롤백되었는지 그 시발점을 확인하고 싶었습니다.
  - **일어난 일**: 아래 타임라인의 `8 DataSourceTransactionManager.doSetRollbackOnly` 줄이 파란색으로 하이라이트되었고, 중앙 상단의 Spring 트랜잭션 박스에 스택(`7 processRollback`, `8 doSetRollbackOnly`)이 올라왔습니다. 오른쪽 패널에는 `DataSourceTransactionManager.doSetRollbackOnly`의 상세 정보와 함께 **"한 일: 롤백 전용 표시(연결 위로는 보내지 않음)"**이라는 설명이 떴습니다. 안쪽 트랜잭션이 실패했을 때 DB로 바로 ROLLBACK 명령을 보내는 것이 아니라, 스프링 메모리 상에 롤백 플래그만 세팅한다는 사실을 처음 눈으로 확인했습니다.
- **조작 2: 주요 순간의 `+1,236 C1 ROLLBACK` 버튼 클릭**
  - **이유**: 롤백 전용 마킹만 해두고 DB로 보내지 않았다면, 실제 DB 커넥션(C1)에는 언제 ROLLBACK SQL/패킷이 나가는지 확인하기 위해서였습니다.
  - **일어난 일**: 타임라인 +1,236 ms 지점의 `pgjdbc` `QueryExecutorImpl.execute`가 선택되었고, 우측 패널에 `C1 · 요청 8: 준비 문장 S_1 ROLLBACK (P B D E S)` 메시지와 서버 사건 `+1,238 커밋 로그: 753 중단 abort_clog`가 나타났습니다. 바깥 메서드(`Outer.required`)가 끝날 때 스프링이 정상 커밋(`AbstractPlatformTransactionManager.commit`)을 시도하다가, 앞서 찍힌 롤백 전용 표시를 보고 `processRollback` -> `doRollback`을 호출하여 비로소 DB에 ROLLBACK을 보낸 것임을 확인했습니다.
- **조작 3: 주요 순간의 `+1,241 ✕ UnexpectedRollbackException` 버튼 클릭**
  - **이유**: DB 롤백 이후 호출자에게 예외가 어떻게 최종 전달되는지 보기 위해 눌렀습니다.
  - **일어난 일**: `AbstractPlatformTransactionManager.processRollback`이 선택되면서 우측에 `한 일: ✕ UnexpectedRollbackException: Transaction rolled back because it has been marked as rollback-only`라는 예외 발생 지점과 호출 사슬(`TransactionInterceptor.invoke` -> `invokeWithinTransaction` -> `commitTransactionAfterReturning` -> `commit` -> `processRollback`)이 명확하게 표시되었습니다.

---

## 3. 조작 기록 표

| 번호 | 한 일 | 기대 | 이해 | 헷갈림 |
|:---:|:---|:---|:---|:---|
| 1 | `new_page`로 페이지 접속 | 스프링 전파 관련 학습 페이지가 뜰 것 | 실제 자바 앱과 DB 간의 정밀 관측 도구임을 확인 | 화면에 정보가 너무 많아 어디부터 봐야 할지 살짝 막막함 |
| 2 | `+822 롤백 전용 표시` 클릭 | 안쪽 예외 시 롤백 마킹 동작 확인 | `doSetRollbackOnly`가 실행되며 "연결 위로는 보내지 않음(메모리 플래그)"을 이해 | "연결 위로는 보내지 않음"이 구체적으로 DB 통신을 안 한다는 뜻인지 확인 필요했음 |
| 3 | `+1,236 C1 ROLLBACK` 클릭 | 실제 DB 롤백 시점 확인 | 바깥 메서드 정상 리턴 후 스프링이 `commit()` 중에 롤백 플래그를 보고 실제 `ROLLBACK` 패킷을 전송함을 이해 | 없음 |
| 4 | `+1,241 ✕ UnexpectedRollbackException` 클릭 | 예외 생성 위치 확인 | `processRollback` 안에서 롤백 완료 후 `UnexpectedRollbackException`을 던짐을 확인 | 바깥에서 try-catch를 했는데도 왜 프록시 단계에서 다시 던져지는지 호출 사슬로 이해 |
| 5 | `표지·약어 풀이 ↓` 클릭 | 낯선 약어들(P B D E S, Z, xid 등)의 정의 확인 | Q(단순질의), P B D E S(확장질의), Z(준비완료), xid(트랜잭션 번호), doSuspend 등 용어 풀이 습득 | 풀이가 화면 아래로 펼쳐져서 스크롤을 확인해야 했음 |
| 6 | 시나리오 `REQUIRES_NEW` 클릭 | 새 트랜잭션 열릴 때 커넥션 동작 확인 | 바깥 트랜잭션 C1 외에 C2 커넥션이 추가로 사용되는 결과 확인 | 트랜잭션 두 개가 스레드 하나에서 어떻게 공존하는지 궁금증 생김 |
| 7 | 주요 순간 `+454 C1 떼어 둠` 클릭 | 기존 트랜잭션 커넥션의 행방 확인 | `DataSourceTransactionManager.doSuspend`가 실행되며 C1을 스레드에서 내려두고 풀에서 C2를 빌림을 이해 | C1이 풀에 반납된 것인지 아니면 스레드가 쥐고만 있는 것인지 다이어그램을 자세히 살펴봄 |
| 8 | 시나리오 `REQUIRES_NEW · 풀 1칸` 클릭 | 커넥션 풀 1개일 때의 문제 확인 | "기다림의 고리(풀)" 경고 박스와 함께 `ConcurrentBag.borrow`에서 2초간 멈추는 현상 목격 | 데드락인지 단순 타임아웃인지 헷갈렸으나 "자기 자신과의 데드락"임을 설명문으로 이해 |
| 9 | `+2,592 ✕ SQLTransientConnectionException` 클릭 | 풀 타임아웃 예외 지점 확인 | 2,000ms 타임아웃 후 `SQLTransientConnectionException` 발생, 스프링이 `doResume`으로 C1을 복구 후 롤백함을 이해 | 풀 에러가 났는데 C1이 정상적으로 롤백되고 닫히는지 추적 |
| 10 | 시나리오 `REQUIRES_NEW · 같은 행` 클릭 | 바깥과 안쪽이 동일 행 수정 시 동작 확인 | DB 행 락 대기 발생 확인. "기다림의 고리"에 DB 데드락 감지기(`NoDeadlock`)가 감지 못한다는 내용 확인 | 왜 DB 데드락 감지가 안 되는지 의아했으나, C1이 DB 안에서 락을 기다리는 게 아니라 앱 입력을 기다리는 상태(`idle in transaction`)여서 서버 관점에서는 단순 대기라는 점을 깨달음 |
| 11 | 시나리오 `NESTED · 안쪽이 실패` 클릭 | 한 번도 안 써본 NESTED 동작 확인 | DB 커넥션을 C1 하나만 유지한 채 `SAVEPOINT "SAVEPOINT_1"`을 만들고, 안쪽 실패 시 `ROLLBACK TO SAVEPOINT`만 수행함을 확인 | 하위 트랜잭션(754)만 취소되고 바깥 트랜잭션(753)은 살아남아 정상 커밋(kim 90)되는 차이를 완벽히 이해 |
| 12 | 시나리오 `0.5초 안에 다시 빌림` 클릭 | 0.5초 조건의 의미 확인 | 첫 번째 빌릴 때는 `+8.93 생존 검사`가 있었는데, 두 번째 빌릴 때는 생존 검사가 없음 | 왜 검사를 안 하는지 헷갈려서 두 번째 빌림 순간을 클릭해 봄 |
| 13 | 주요 순간 `+132 C1 빌림` 클릭 | 생존 검사 생략 이유 확인 | 우측 상세에 "풀에서 C1 빌림 · 쉰 지 6 ms < 창 500 ms -> 검사 건너뜀"이라는 HikariCP 최적화 로직 확인 | HikariCP에 500ms 유휴 검사 면제 규칙이 있다는 것을 처음 알게 됨 |
| 14 | 상단 `실행 b` 클릭 | 실행 a와 실행 b의 차이 확인 | 같은 시나리오를 다시 돌렸을 때의 실제 관측치(시간이 132ms -> 126ms 등으로 미세하게 다름)를 보여줌을 확인 | 모의 데이터가 아니라 실제 다회차 런 기록임을 확신 |
| 15 | 시나리오 `REQUIRED · 안쪽이 실패` 복귀 | 필터 기능 및 내부 호출 탐색 준비 | 초기 시나리오로 복귀 | 없음 |
| 16 | 보이는 줄 필터 `전부` 클릭 | 숨겨진 스프링 내부 호출 확인 | 호출 수가 33개에서 82개로 늘어나며 프록시, `obtainDataSource`, `isExistingTransaction` 등 모든 내부 메서드 노출 | 너무 많은 줄이 펼쳐져서 스크롤 압박이 생김 |
| 17 | 중앙 PostgreSQL 박스(`C1 · pid 4239`) 클릭 시도 | DB 세션 상세 정보 확인 기대 | 클릭 이벤트가 동작하지 않음 | 이 박스가 클릭 가능한 버튼인지 단순 정보 표시인지 혼란 |
| 18 | 우측 호출 사슬에서 `2 Outer.required` 클릭 | 호출 트리 탐색 기대 | 타임라인과 좌측 상자 상태가 `Outer.required` 실행 시점(+8.59ms)으로 정확히 점프함 | 완벽하게 연동되어 동작함 |
| 19 | 키보드 `ArrowDown` 입력 (1회) | 아래 줄 이동 기대 | 바로 아래 줄인 `JdbcTemplate.update`(+8.60ms)로 선택이 이동하고 좌측·우측 패널이 동기화됨 | 키보드로 한 줄씩 따라갈 수 있다는 점이 매우 편리했음 |
| 20 | 키보드 `ArrowDown` 입력 (2회) | 추가 이동 기대 | `JdbcTemplate.execute`(+8.69ms)로 이동 | 없음 |
| 21 | 키보드 `ArrowDown` 입력 (3회) | 추가 이동 기대 | `QueryExecutorImpl.execute`(+9.07ms)로 이동하며 실제 DB로 `BEGIN + UPDATE` 패킷이 날아가는 순간 확인 | 자바 메서드 호출이 pgjdbc를 거쳐 소켓 패킷으로 바뀌는 경계를 직접 확인 |

---

## 4. 다 써 본 뒤 화면에 대한 이해 (한 문단)

이 화면은 Spring Framework의 `@Transactional` 전파 옵션(`REQUIRED`, `REQUIRES_NEW`, `NESTED`)과 HikariCP 커넥션 풀, PostgreSQL 데이터베이스 사이에서 일어나는 실제 내부 상호작용을 JVM 메서드 호출(JFR)과 소켓 패킷, DB 엔진 상태 관측을 통해 시간 순으로 정밀하게 동기화하여 보여주는 **트랜잭션 실행 관측 뷰어**입니다. [화면에서 확인] 실무에서 프레임워크가 추상화해 숨겨두었던 프록시 호출 사슬, 커넥션의 `suspend`/`resume`, HikariCP의 생존 검사 생략, 그리고 DB 레벨의 SAVEPOINT 및 락 대기 상태가 단일 타임라인 상에서 유기적으로 연결되어 시각화됩니다. [화면에서 확인] 백엔드 개발자가 흔히 겪는 "내부 롤백 전용 마킹으로 인한 UnexpectedRollbackException"이나 "풀 크기 1일 때 REQUIRES_NEW의 커넥션 고갈 데드락", "동일 행 수정 시 앱-DB 분산 데드락"과 같은 까다로운 장애 상황을 재현하고 원인을 규명해 주는 훌륭한 교육 및 디버깅 도구라고 생각합니다. [추측]

---

## 5. 1번 질문마다, 그리고 그 밖에 배운 것

### 질문 1: 전파 옵션(REQUIRED, REQUIRES_NEW, NESTED)이 실제로 어떻게 동작하며 왜 롤백이 번지는가?
- **REQUIRED에서 예외 전파**: 안쪽 메서드(`Inner.requiredFails`)에서 예외가 발생하면 즉시 DB에 ROLLBACK을 날리지 않고, 스프링의 `DataSourceTransactionManager.doSetRollbackOnly`가 실행되어 트랜잭션에 "롤백 전용 표시" 플래그만 세팅합니다. [화면에서 확인: 타임라인 +822 ms, 우측 상세 `DataSourceTransactionManager.doSetRollbackOnly`, "한 일: 롤백 전용 표시(연결 위로는 보내지 않음)"]
- 이후 바깥 메서드가 `try-catch`로 예외를 잡아 정상 종료하더라도, 스프링의 `AbstractPlatformTransactionManager.commit` 단계에서 이 롤백 플래그를 검사하여 실제 DB로 `C1 ROLLBACK`을 날리고 바깥 호출자에게 `UnexpectedRollbackException`을 던집니다. [화면에서 확인: 타임라인 +1,235 ~ +1,241 ms, 우측 패널 `AbstractPlatformTransactionManager.processRollback`]
- **REQUIRES_NEW의 동작**: 바깥 트랜잭션의 커넥션(C1)을 닫거나 반납하지 않고 `DataSourceTransactionManager.doSuspend`를 통해 스레드에서 잠시 "떼어 둡니다". 그리고 HikariCP 풀에서 새 커넥션(C2)을 빌려와 완전히 독립된 트랜잭션(754)을 열고 커밋한 뒤, `doResume`으로 다시 C1을 스레드에 묶어서 바깥 트랜잭션(753)을 이어갑니다. [화면에서 확인: `REQUIRES_NEW` 시나리오 주요 순간 `+454 C1 떼어 둠`, `+454 C2 빌림`, `+932 C2 COMMIT`, `+946 C1 다시 묶음`]
- **NESTED의 동작**: 새 커넥션을 빌리지 않고 기존 C1 커넥션 위에서 DB의 `SAVEPOINT "SAVEPOINT_1"` 명령을 실행합니다. 안쪽 메서드가 실패하면 전체 롤백이 아니라 `C1 ROLLBACK TO SAVEPOINT`를 날려 하위 트랜잭션(754)만 취소하고, 바깥 트랜잭션(753)은 그대로 유지되어 정상 커밋됩니다. [화면에서 확인: `NESTED` 시나리오 타임라인 +472 ms `C1 SAVEPOINT`, +890 ms `C1 ROLLBACK TO SAVEPOINT`, +1,248 ms `C1 COMMIT`]

### 질문 2: HikariCP 풀과 DB 커넥션이 스프링 트랜잭션과 어떻게 엮여서 돌아가는가?
- **풀 크기 1일 때의 REQUIRES_NEW 데드락**: 풀 최대 크기가 1칸일 때 바깥 메서드가 C1을 쥐고 있는 상태에서 안쪽 REQUIRES_NEW가 실행되면, `ConcurrentBag.borrow`가 풀의 빈자리를 무한정 기다리며 멈춥니다. C1을 풀에 돌려줄 유일한 주체인 main 스레드 자신이 대기 상태에 갇혀 버리기 때문에 앱 내부에서 "기다림의 고리(Deadlock)"가 닫히게 됩니다. [화면에서 확인: `REQUIRES_NEW · 풀 1칸` 시나리오, 붉은색 "기다림의 고리(풀) +587 ~ +2,590 ms" 박스]
- 결국 `connectionTimeout 2,000 ms`가 지나 `SQLTransientConnectionException`이 터진 뒤에야 대기가 풀리고, 스프링이 `doResume`으로 C1을 다시 스레드에 묶은 후 롤백 처리합니다. [화면에서 확인: 타임라인 +2,592 ms `HikariPool.getConnection(long)`, 우측 패널 "Connection is not available, request timed out after 2004ms"]

### 질문 3: 대화형 시뮬레이터인가, 트레이스 뷰어인가?
- 사용자가 임의의 코드를 직접 타이핑하거나 파라미터를 변경하는 샌드박스는 아니며, 사전에 정교하게 설계된 6개의 대표 시나리오에 대해 실제 자바(Java 25)와 PostgreSQL(18.6) 컨테이너를 가동하여 JFR 및 네트워크를 캡처해 둔 **인터랙티브 트레이스 뷰어**였습니다. [화면에서 확인: 상단 헤더 "Java 25.0.4", "PostgreSQL 18.6 · 컨테이너", "실행 a·b", `시각은 장면 시작부터 ms`]

### 그 밖에 화면에서 새롭게 배운 것
1. **HikariCP의 유휴 시간(500ms) 생존 검사 생략 최적화**: 커넥션을 빌릴 때 매번 DB에 검사 쿼리(`isValid` / `isConnectionAlive`)를 날리는 줄 알았는데, 반납된 지 500ms 이내(창 500ms)인 커넥션은 생존 검사를 건너뛰고 바로 반환합니다. [화면에서 확인: `0.5초 안에 다시 빌림` 시나리오 +132 ms `ConcurrentBag.borrow`, "풀에서 C1 빌림 · 쉰 지 6 ms < 창 500 ms -> 검사 건너뜀"]
2. **동일 행 수정 시 앱-DB 분산 데드락과 DB 감지 불가**: 바깥 트랜잭션이 수정한 행을 안쪽 REQUIRES_NEW가 다시 수정하려고 하면 DB 행 락(`Lock/transactionid`) 대기가 발생합니다. 이때 DB의 데드락 감지기는 C1이 락을 기다리는 게 아니라 단순히 클라이언트 명령을 대기(`idle in transaction`)하는 것으로 인식하기 때문에 `NoDeadlock`으로 판정하여 데드락을 감지하지 못합니다. 결국 오직 `lock_timeout`(여기서는 3초)에 의해서만 이 교착 상태가 풀립니다. [화면에서 확인: `REQUIRES_NEW · 같은 행` 시나리오, "기다림의 고리" 설명 및 우측 막힌 동안 관찰 "C2 -> transactionid 753 ShareLock", 커밋 로그 `753 진행 중`]
3. **PostgreSQL 확장 질의 프로토콜의 실체**: 평소에는 추상화되어 몰랐던 드라이버 레벨의 패킷 흐름인 `P B D E S`(Parse, Bind, Describe, Execute, Sync)와 응답 `C`(CommandComplete), `Z(T/I/E)`(ReadyForQuery: Transaction/Idle/Error)가 실제로 어떻게 오가는지 확인했습니다. [화면에서 확인: 표지·약어 풀이, 타임라인 `연결 위 메시지` 열]

---

## 6. 막힌 곳, 헷갈린 곳, 잘못 읽었다가 나중에 바로잡은 곳

1. **아키텍처 박스 인터랙션 혼란**:
   - 중앙의 `C1 · pid 4239`나 `HikariCP` 박스가 시각적으로 강조되어 있어 버튼처럼 누르면 해당 컴포넌트의 상세 설정이나 내부 상태가 뜰 줄 알고 클릭해 보았으나 아무런 반응이 없었습니다. 화면의 안내 문구를 읽고 나서야 실제 인터랙션의 중심은 상단의 **"주요 순간 버튼"**과 하단 타임라인의 **"호출 줄"**, 그리고 **"키보드 방향키"**라는 것을 깨달았습니다.
2. **타임라인 메시지 직접 클릭 시도**:
   - 타임라인에 표시된 SQL 화살표(`-> C1 Q BEGIN...`)를 마우스로 직접 누르려 했으나, 텍스트/이미지 레이어로 묶여 있어 개별 패킷을 마우스로 집어내기 어려웠습니다. 화면 설명에 적힌 `↑ ↓ 키로 줄 이동`을 보고 키보드 방향키를 누르니 순차적으로 정확히 해당 패킷 실행 지점으로 이동할 수 있었습니다.
3. **시간 축의 ⏸(일시정지) 구간에 대한 오해**:
   - 타임라인 중간에 `+105 ~ +452 ms` 구간이 `⏸ 멈춤 자리 ... 실험 도구가 앱을 세우고 상대 쪽 상태를 찍은 구간`이라고 나와 있는 것을 처음엔 "스레드가 300ms 동안 sleep이나 대기에 빠졌나?" 하고 잘못 읽었습니다. 약어 풀이를 읽고 나서야 실험 도구가 DB 상태 뷰(pg_stat_activity 등)를 일관되게 캡처하기 위해 앱을 잠깐 freeze시킨 구간이라 앱 소요 시간에서는 제외된다는 점을 이해했습니다.

---

## 7. 이 화면에 바라는 것

1. **컴포넌트 박스(HikariCP, PostgreSQL 등) 클릭 시 필터링/포커스 연동**:
   - 중앙 아키텍처 다이어그램의 `HikariCP` 상자를 누르면 타임라인에서 HikariCP 관련 호출(`getConnection`, `borrow`, `isConnectionAlive` 등)만 강조되거나, PostgreSQL 박스를 누르면 소켓 패킷과 서버 상태 관측만 모아서 볼 수 있는 뷰 모드가 지원되면 훨씬 직관적일 것 같습니다.
2. **비교 뷰(Split View) 제공**:
   - 예컨대 `REQUIRED · 안쪽이 실패`와 `NESTED · 안쪽이 실패`의 차이점(전체 롤백 vs SAVEPOINT 롤백)이나, `REQUIRES_NEW` 정상 vs `풀 1칸` 데드락을 한 화면에서 좌우로 나란히 놓고 비교할 수 있다면 개념의 차이가 훨씬 더 강렬하게 와닿을 것 같습니다.
3. **타임라인 시간 스케일의 시각적 구분**:
   - `+8.60 ms`에서 갑자기 `+452 ms`나 `+2,592 ms`로 넘어갈 때, 시간 간격이 길게 벌어진 구간(잠듦, 락 대기, 풀 고갈 등)은 타임라인 막대의 색상이나 물결선(생략선) 기호로 "여기서 오랜 시간 멈춰 있었다"는 점이 시각적으로 더 확 띄게 강조되면 좋겠습니다.

---

## 8. 쓴 브라우저 조작 수

- **페이지 접속 (`new_page`)**: 1회
- **마우스 클릭 (`click`)**: 18회 (주요 순간 버튼, 시나리오 전환 탭, 필터 버튼, 호출 사슬 링크 등)
- **키보드 조작 (`press_key`)**: 3회 (`ArrowDown`을 이용한 타임라인 줄 순회)
- **화면 관찰 및 구조 확인 (`take_screenshot`, `take_snapshot`)**: 38회
- **총 브라우저 조작 횟수**: **60회** (순수 사용자 상호작용 조작 22회)
