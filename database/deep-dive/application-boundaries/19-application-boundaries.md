# Application Boundaries

## connection, session, transaction manager, pooling

이 절은 `connection, session, transaction manager, pooling`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 DB 커넥션이 단순 소켓이 아니라 세션 상태와 트랜잭션 소유권을 담는 실행 컨텍스트라는 점을 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 connection pool을 단순 소켓 캐시로 보는 함정이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 작은 요청 하나가 커넥션을 빌리는 장면에서 시작한다. HTTP 요청 R-1001이 서비스에 들어오면 트랜잭션 매니저는 풀에서 물리 커넥션을 빌리고, 그 커넥션의 세션 상태를 이번 요청의 실행 조건으로 만든다. 이 흐름은 단순 함수 호출이 아니라 autoCommit, isolation, schema, readOnly, DB backend pid, lock ownership이 함께 움직이는 장면이다. 그래서 DU42의 첫 질문은 어떤 SQL을 실행했는가보다 누가 세션 상태를 바꾸고 누가 commit과 반납을 책임지는가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | Oracle Java SE 21 JDBC Connection API: https://docs.oracle.com/en/java/javase/21/docs/api/java.sql/java/sql/Connection.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Spring Framework transaction reference: https://docs.spring.io/spring-framework/reference/data-access/transaction.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `jvm/spring/spring_transactional.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `database/database-deep-study-plan.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 커넥션 하나가 들고 있는 상태

JDBC 공식 API는 `Connection`을 특정 데이터베이스와 맺은 세션으로 설명한다. 같은 객체 안에서 SQL이 실행되고 결과가 돌아오며, 자동 커밋 모드, 격리 수준, 읽기 전용 힌트, 카탈로그, 스키마, 클라이언트 정보 같은 상태도 붙어 있다. 그래서 풀에서 커넥션을 빌린다는 말은 열린 TCP 연결 하나를 빌리는 것에 그치지 않는다. 이전 사용자가 남긴 세션 상태를 초기화하고, 이번 요청이 어떤 트랜잭션으로 묶일지 정하고, 끝나면 다음 사용자에게 오염 없이 돌려주는 계약까지 함께 빌리는 것이다.

```text
HTTP request R-1001
  -> service method enters
  -> transaction manager borrows Connection C7
  -> C7 session state: autoCommit=false, isolation=READ_COMMITTED, schema=app, tx=TX-881
  -> repository executes UPDATE orders SET status=PAID WHERE id=42
  -> commit releases DB locks, pool reset returns C7
next HTTP request R-1002 must not inherit TX-881 or temporary session choices
```

이 trace에서 움직이는 대상은 `C7`이라는 커넥션이다. `C7`은 요청 사이를 재사용될 수 있지만, 트랜잭션 `TX-881`은 요청 R-1001의 경계에 속한다. 풀은 물리 자원을 재사용하되 논리 작업 단위는 재사용하지 않도록 초기화해야 한다. 운영 장애는 바로 이 구분이 흐려질 때 생긴다. 예를 들어 이전 요청이 `readOnly=true`나 특정 schema를 남겼는데 반납 과정에서 복구되지 않으면 다음 요청은 전혀 다른 코드인데도 쓰기 실패나 엉뚱한 스키마 접근을 만난다.
#### DU42-1.1 커넥션은 소켓보다 넓은 세션이다

1회차 관점에서 커넥션을 소켓 캐시로만 보면 `close()`를 네트워크 종료로 오해한다. 풀 환경의 `close()`는 대개 물리 연결을 끊는 행위가 아니라 현재 대여를 끝내고 세션을 재사용 가능한 상태로 돌리는 행위다. JDBC의 `commit()`은 현재 트랜잭션 동안 만든 변경을 영구화하고 잡고 있던 DB lock을 놓는 API이며, `rollback()`은 현재 트랜잭션 변경을 되돌리고 lock을 놓는다. 이 둘은 소켓 생존 여부와 별개다. 그래서 커넥션 생명주기는 물리 연결 생존, 세션 설정, 트랜잭션 경계, 풀 대여 경계라는 네 층으로 나눠 읽어야 한다.

실무에서 이 구분이 깨지면 장애가 애매하게 보인다. 풀 size는 충분한데 request latency가 늘거나, 특정 API 뒤에만 다른 API가 실패하거나, 테스트에서는 통과하지만 운영에서만 schema, role, isolation 문제가 재현된다. 함정은 '커넥션이 닫히지 않았다'라는 관측을 곧바로 leak으로 단정하는 것이다. 먼저 pool active/idle, DB session state, transaction open time, borrow stack을 같이 봐야 한다.

#### DU42-1.2 트랜잭션 매니저는 owner를 정한다

트랜잭션 매니저의 핵심 역할은 SQL을 대신 실행하는 것이 아니라 이번 실행 단위에서 누가 커밋과 롤백을 결정하는지 정하는 것이다. Spring의 트랜잭션 추상화는 실제 리소스 트랜잭션을 시작하고, 현재 스레드에 리소스를 묶고, 정상 반환이면 commit, 실패 규칙에 걸리면 rollback을 호출한다. 이때 repository는 보통 자신이 커밋하지 않는다. repository가 중간에 임의로 commit하면 상위 서비스가 여러 변경을 하나의 불변식으로 묶으려는 목적이 깨진다.

운영 함정은 helper나 DAO가 '안전하게 저장하려고' 내부 commit을 호출하는 것이다. 당장은 데이터가 남으니 안정적으로 보이지만, 상위 use case가 실패했을 때 일부 row만 확정되는 찢어진 상태가 된다. 그래서 커밋 권한은 owner-facing boundary에 있어야 하고, 하위 계층은 같은 트랜잭션에 참여하는 쪽이 기본이다.

#### DU42-1.3 풀은 성능 장치이면서 격리 장치다

커넥션 풀은 매 요청마다 로그인, 인증, 네트워크 핸드셰이크를 새로 하지 않게 해 성능을 높인다. 그러나 더 중요한 책임은 재사용되는 세션을 요청 단위로 격리하는 것이다. 풀은 borrow 시 유효성을 확인하고, return 시 autoCommit, readOnly, isolation, catalog, schema, warnings 같은 상태를 정리해야 한다. 드라이버와 풀 구현마다 세부 정책은 다르므로 애플리케이션은 '풀에 넣었으니 알아서 깨끗하다'가 아니라 자신이 바꾼 상태를 가능한 한 명시적으로 복구하는 습관을 가져야 한다.

장애 관측은 pool metric에서 시작한다. active connection이 max에 붙어 있고 DB에서는 idle in transaction 세션이 오래 남아 있다면 단순 QPS 증가가 아니라 트랜잭션을 열어 둔 채 외부 호출이나 긴 계산을 수행했을 가능성이 크다. 반대로 DB 세션은 적고 애플리케이션 thread가 대기 중이면 pool acquisition timeout이 병목일 수 있다.

#### DU42-1.4 세션 상태는 요청 사이에 새어 나갈 수 있다

DB 세션에는 temporary table, session variable, search path, time zone, lock wait timeout, statement timeout처럼 요청 코드에 직접 보이지 않는 상태가 붙을 수 있다. 한 요청에서 장애 대응을 위해 `SET statement_timeout=5000`을 실행하고 복구하지 않았다면, 같은 물리 세션을 다음 요청이 빌렸을 때 갑자기 5초 제한에 걸릴 수 있다. JDBC 공식 설명이 설정 변경에는 가능한 Connection API를 쓰라고 말하는 이유도 이런 세션 상태를 드라이버와 풀의 관리 범위 안에 두기 위해서다.

이 함정은 재현이 어렵다. 특정 요청 순서 뒤에만 실패하고, 재시작하면 사라지며, 같은 SQL을 DB 콘솔에서 실행하면 성공한다. 그래서 관측에는 connection id, backend pid, schema/search_path, transaction start timestamp를 로그에 연결하는 것이 도움이 된다.

#### DU42-1.5 auto-commit은 작은 편의이자 큰 오해의 근원이다

JDBC 커넥션은 기본적으로 auto-commit mode일 수 있고, 이 상태에서는 각 statement 실행 뒤 변경이 자동 확정된다. 학습할 때는 편하지만 서비스 코드에서는 여러 statement를 하나의 도메인 불변식으로 묶기 어렵다. 트랜잭션 매니저가 시작되면 보통 autoCommit=false로 바꾸고, 경계가 끝날 때 commit 또는 rollback을 수행한다. 이 전환 자체도 세션 상태 변화이므로 pool 반납 시 복구가 필요하다.

실무에서 `INSERT` 하나씩은 성공했는데 전체 주문 생성이 반쯤 남는 문제는 auto-commit 경계 오해에서 자주 나온다. 특히 batch import나 보상 처리에서 한 row씩 auto-commit하면 실패 지점 이후 재시작 기준을 따로 설계해야 한다. 반대로 너무 큰 트랜잭션은 lock과 undo/redo 압력을 키우므로 작업 단위의 business invariant와 DB 비용을 함께 정해야 한다.

#### DU42-1.6 검증은 pool 숫자와 DB 세션을 같이 본다

커넥션 경계 검증은 애플리케이션 로그만으로 닫히지 않는다. 풀의 active/idle/pending metric, DB의 현재 세션 목록, 열린 트랜잭션 시작 시각, lock wait, statement timeout, commit/rollback count를 함께 본다. 테스트에서는 service method 진입 전후로 connection identity를 찍고, 같은 트랜잭션 안의 mapper 호출이 같은 connection을 쓰는지 확인할 수 있다.

운영에서 함정은 평균 latency만 보는 것이다. 커넥션 누수나 긴 트랜잭션은 평균보다 꼬리 지연과 timeout으로 먼저 보인다. `pool.pending > 0`, `active == max`, `idle in transaction` 증가, commit 없이 오래 열린 backend pid가 동시에 보이면 소켓 문제가 아니라 transaction/session ownership 문제로 봐야 한다.



### 작은 SQL trace로 다시 보기

```sql
-- request R-1001 begins
SELECT txid_current();        -- DB별 함수는 다르지만 관측 목적은 transaction identity 확인이다.
UPDATE account SET balance = balance - 10000 WHERE id = 10;
UPDATE account SET balance = balance + 10000 WHERE id = 20;
-- service returns normally -> transaction manager calls commit()

-- request R-1002 must see committed result, not the previous uncommitted session state
SELECT id, balance FROM account WHERE id IN (10, 20);
```

이 SQL trace의 핵심은 두 UPDATE가 같은 물리 트랜잭션 안에서 확정되어야 한다는 점이다. 두 UPDATE 사이에서 애플리케이션이 외부 API를 오래 기다리면 계좌 row lock이 길게 잡히고, 다른 요청은 대기하거나 timeout이 난다. 그래서 트랜잭션 경계 설계는 '어디서 commit하느냐'뿐 아니라 '트랜잭션을 열어 둔 채 무엇을 기다리느냐'까지 포함한다.
### 관측과 검증

DU42의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU42 verification checklist
1. log connection borrow/return with request id and transaction id
2. observe pool active/idle/pending while running a slow transactional endpoint
3. query DB session table for long transaction start time and lock wait
4. run a test that mutates readOnly/isolation and verifies reset before next borrow
```

DU42의 검증은 테스트 한 번보다 관측 연결이 중요하다. 배포 전에는 같은 요청 안의 repository 호출이 같은 connection과 transaction id를 쓰는지 확인하고, 배포 후에는 pool active/idle/pending, DB session age, lock wait, backend pid를 request id와 묶어 본다. 특히 커넥션 경계에서는 성공 응답보다 반납 시점과 세션 복구 여부가 더 중요하다.
### 기억할 압축 문장

커넥션 풀은 열린 소켓을 아끼는 장치이지만, 서비스 관점에서는 세션 상태와 트랜잭션 소유권을 안전하게 대여하고 반납하는 장치다. 커넥션을 빌린 순간부터 반납할 때까지 누가 커밋을 결정하는지, 어떤 세션 상태가 바뀌었는지, 다음 요청에게 무엇이 새지 않아야 하는지를 함께 추적해야 한다.

#### DU42-S1 등장 배경: 매 요청마다 연결을 새로 열던 비용

초기 DB 애플리케이션은 요청마다 연결을 만들고 인증하고 SQL을 실행한 뒤 닫는 흐름으로도 충분했다. 그러나 웹 트래픽이 늘어나면 연결 생성 비용과 DB 서버의 세션 생성 비용이 요청 처리 시간의 큰 부분을 차지한다. 그래서 풀링은 물리 연결을 재사용해 이 비용을 줄이는 배경에서 등장했다. 다만 성능 이유로 재사용을 시작한 순간, 세션 상태가 요청 사이에 새는 새로운 문제가 생겼다.

이 배경을 알면 풀링을 단순 최적화로만 보지 않게 된다. 풀은 빠르게 빌려 주는 장치이면서, 이전 요청의 schema, isolation, readOnly, temporary state, 열린 트랜잭션 흔적을 다음 요청에게 넘기지 않는 격리 장치다. 성능 개선이 만든 위험을 격리 규칙이 다시 잡아 주는 구조라고 보면 장애 원인을 더 빨리 좁힐 수 있다.

#### DU42-S2 driver, pool, transaction manager의 세 책임

JDBC driver는 DB 프로토콜과 세션 API를 제공하고, pool은 그 세션 객체의 대여와 반납을 관리하며, transaction manager는 이번 use case에서 commit과 rollback을 결정한다. 세 계층이 모두 커넥션을 말하기 때문에 헷갈리지만 질문이 다르다. driver의 질문은 SQL을 어떻게 보내고 받는가이고, pool의 질문은 누가 언제 빌리고 반납하는가이며, transaction manager의 질문은 어떤 작업 묶음이 확정되는가이다.

장애 대응에서도 이 세 책임을 분리해야 한다. SQL 문법 오류는 driver나 DB가 알려 주지만, 반납되지 않는 커넥션은 pool metric에서 보이고, commit되지 않은 긴 트랜잭션은 transaction owner와 DB session table에서 보인다. 한 화면에서 모두 connection으로 보인다고 한 원인으로 몰면 진단이 늦어진다.

#### DU42-S3 세션 초기화는 보이지 않는 안전 작업이다

커넥션이 풀로 돌아갈 때 애플리케이션은 보통 그 장면을 보지 않는다. 그러나 그 순간 autoCommit 복구, readOnly 해제, isolation 복구, warnings clear, statement close, transaction cleanup 같은 작업이 일어난다. 이 작업은 문서나 로그에 잘 드러나지 않지만, 실패하면 다음 요청이 전혀 다른 조건에서 실행된다.

실무에서는 이 보이지 않는 초기화 실패가 특정 순서에서만 재현되는 장애로 나타난다. 예를 들어 관리자 API가 높은 isolation을 설정한 뒤 복구하지 않았고, 다음 일반 조회가 같은 세션을 빌리면 갑자기 lock wait가 길어질 수 있다. 그래서 connection customizer나 raw SQL SET을 쓸 때는 반납 전 복구 경로를 테스트해야 한다.

#### DU42-S4 트랜잭션을 열고 외부를 기다리는 비용

트랜잭션이 열리면 커넥션 하나가 pool에서 점유되고, DB에서는 lock이나 snapshot, undo 정보가 유지될 수 있다. 이 상태에서 PG 호출, 파일 다운로드, 다른 서비스 HTTP 요청을 기다리면 데이터베이스 자원이 외부 지연 시간에 묶인다. 작은 트래픽에서는 괜찮아 보이지만 장애 때 외부 timeout이 길어지면 pool 고갈과 DB lock 대기가 동시에 커진다.

이 함정은 대개 서비스 코드가 보기 좋게 한 메서드 안에 있을 때 생긴다. 주문 생성, 결제 승인, 재고 차감, 알림 발송을 모두 하나의 `@Transactional`에 넣으면 읽기는 쉽지만 장애 반경은 커진다. 외부 부작용은 상태 머신과 outbox, 멱등키, 후속 복구로 분리하는 편이 더 안전한 경우가 많다.

#### DU42-S5 관측값을 한 줄로 묶는 법

실제 장애 로그에는 request id만 있으면 부족하다. request id, thread name, pool connection id, DB backend pid, transaction id, borrow elapsed, commit elapsed, rollback 여부를 한 줄에 가까운 형태로 묶어야 한다. 그래야 특정 요청이 어떤 세션을 빌렸고, 그 세션이 DB에서 어떤 lock을 잡았으며, 언제 반납됐는지 이어 볼 수 있다.

이 관측 설계가 없으면 운영자는 애플리케이션 로그와 DB 로그를 손으로 맞춰야 한다. 초 단위 timestamp가 맞지 않거나 timezone이 다르면 몇 시간이 사라진다. 시니어가 먼저 보는 지점은 코드보다 관측 가능한 상관키다. 상관키가 없으면 좋은 설계도 장애 순간에는 증명하기 어렵다.

#### DU42-S6 pool size 증설이 항상 해답은 아니다

커넥션 부족 알람이 뜨면 pool size를 늘리는 반응이 흔하다. 그러나 원인이 긴 트랜잭션, 커넥션 누수, 외부 호출 대기, 느린 쿼리라면 pool을 늘려도 DB에는 더 많은 동시 세션과 lock 대기가 몰린다. 잠깐 timeout은 줄 수 있지만 DB 압력이 커져 전체 장애가 길어질 수 있다.

따라서 증설 전에는 active가 왜 오래 머무는지 봐야 한다. borrow stack sampling, slow query, transaction age, thread dump, external latency를 같이 확인한다. pool size는 처리량 조절 장치이지 근본 원인 은폐 장치가 아니다.

#### DU42-S7 읽기 전용과 격리 수준은 요청 정책이다

readOnly와 isolation은 SQL 한 문장의 속성이 아니라 요청 동안 적용되는 세션 정책일 수 있다. 조회 API라고 모두 같은 readOnly가 아니며, 재고 확인처럼 최신성이 중요한 읽기와 리포트 조회처럼 replica 지연을 허용하는 읽기는 다르다. 격리 수준도 도메인 불변식과 이상 현상 허용 범위에 따라 정해야 한다.

함정은 annotation 옵션을 성능 튜닝 표식처럼 붙이는 것이다. `readOnly=true`가 replica routing과 연결된 환경에서 쓰기 직후 읽기를 하면 지연된 값을 볼 수 있다. 격리 수준을 올리면 anomaly는 줄지만 lock 비용이 커질 수 있다. 옵션은 항상 관측 가능한 목적과 함께 둬야 한다.

#### DU42-S8 반납 누락과 긴 대여를 구분한다

커넥션 누수는 반납이 영원히 일어나지 않는 경우이고, 긴 대여는 반납은 되지만 너무 늦는 경우다. 둘 다 pool active 증가로 보이지만 처방이 다르다. 누수는 close 경로, try-with-resources, transaction synchronization cleanup을 봐야 하고, 긴 대여는 그 안에서 무엇을 기다렸는지 봐야 한다.

운영 함정은 누수 탐지 로그가 없다고 문제가 없다고 보는 것이다. 모든 장애가 누수는 아니다. 외부 API 지연 20초가 트랜잭션 안에 들어가면 커넥션은 정상 반납되지만 사용자에게는 timeout과 pool 대기처럼 보인다. 그래서 active duration histogram이 중요하다.

#### DU42-S9 DB session table을 읽는 습관

애플리케이션 메트릭은 pool 관점이고 DB session table은 DB 관점이다. PostgreSQL이면 `pg_stat_activity`, MySQL이면 `performance_schema`나 processlist 계열에서 session state, query, transaction age, wait event를 확인한다. 이 둘을 같이 보면 애플리케이션이 빌렸다고 믿는 커넥션이 DB에서 실제로 무엇을 하는지 드러난다.

검증은 간단한 실험으로도 가능하다. 의도적으로 트랜잭션을 열고 sleep하는 endpoint를 만든 뒤 pool active와 DB session age가 같이 증가하는지 본다. 이 실험은 운영 장애를 축소한 모형이고, 문서의 설명이 실제 관측으로 닫히는 지점이다.

#### DU42-S10 최종 설계 질문

커넥션 경계 설계의 마지막 질문은 pool 종류나 설정값이 아니라 owner다. 누가 대여하고, 누가 상태를 바꾸며, 누가 commit하고, 누가 반납하는가. 이 owner가 한 흐름으로 설명되지 않으면 코드가 아무리 짧아도 장애 때 복구하기 어렵다.

따라서 리뷰에서는 `Connection`이 인자로 깊이 전달되는지, repository가 commit을 호출하는지, raw SQL SET이 복구되는지, 외부 호출이 트랜잭션 안에 있는지, pool metric이 있는지부터 본다. 이 질문들이 닫히면 connection, session, transaction manager, pooling은 하나의 경계 모델로 묶인다.

### DU42 운영 케이스 매트릭스

DU42의 본문을 실제 장애 대응에 쓰려면 `connection, session, transaction manager, pooling`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| 풀 고갈과 긴 트랜잭션 분리 | active=max, pending 증가, DB에는 idle in transaction | pool 증설 전 트랜잭션 age와 외부 대기 구간을 분리한다 | pool metric + DB session table |
| 세션 설정 오염 | 특정 관리자 API 뒤 일반 API가 다른 schema나 timeout으로 실패 | raw SET이나 connection property 변경의 복구 경로를 확인한다 | connection id별 schema/search_path/readOnly 로그 |
| autoCommit 경계 오해 | batch import 중 일부 row만 저장되고 나머지는 실패 | row 단위 자동 확정인지 chunk transaction인지 명확히 한다 | commit count, 실패 row 이후 재시작 결과 |
| 외부 호출이 트랜잭션 안에 있음 | PG timeout 동안 DB row lock이 오래 유지됨 | 외부 부작용 전후 상태를 나눠 transaction 시간을 줄인다 | lock wait graph, external latency, tx duration |
| 커넥션 누수와 정상 장기 대여 | active가 줄지 않거나 너무 늦게 줄어듦 | close 누락인지 긴 작업인지 duration histogram으로 나눈다 | leak detection log, borrow stack, active duration |
| readOnly routing 부작용 | 쓰기 직후 읽기가 replica 지연 값을 봄 | readOnly가 replica routing에 연결되는지 확인한다 | routing datasource log, replica lag metric |
| isolation 상향의 비용 | 일부 API에서 lock wait와 deadlock 증가 | 격리 수준이 막으려는 anomaly와 비용을 비교한다 | DB wait event, deadlock log |
| statement timeout 상속 | 특정 요청 뒤 긴 리포트 쿼리가 조기 실패 | 세션 timeout 변경 후 복구 여부를 테스트한다 | connection return hook, statement_timeout 조회 |
| DB session과 application log 불일치 | 앱은 idle인데 DB는 active query로 보임 | backend pid와 request id를 연결한다 | backend pid, request id MDC, slow query log |
| transaction owner 불명확 | DAO helper가 내부 commit 호출 | commit authority를 service/use case owner로 되돌린다 | code search commit/rollback, integration rollback test |
| pool validation 과신 | isValid는 통과하지만 업무 SQL은 실패 | 연결 생존과 업무 권한/스키마 검증을 분리한다 | validation query, first business query error |
| 운영 runbook 부재 | 알람은 있는데 누가 무엇을 확인할지 모름 | metric -> DB query -> code owner 순서를 runbook화한다 | runbook drill, incident note |

#### DU42-M1 풀 고갈과 긴 트랜잭션 분리

이 상황에서는 커넥션이 부족하다는 표면 증상만 보고 풀 크기를 키우면 DB 동시 세션과 lock wait가 함께 늘 수 있다. 먼저 어떤 요청이 커넥션을 오래 빌렸는지 보고, 그 요청이 SQL을 실행 중인지 외부 API를 기다리는지 나눠야 한다.

#### DU42-M2 세션 설정 오염

세션 오염은 재시작하면 사라지고 요청 순서가 맞아야 나타난다. 그래서 장애 당시 같은 connection id를 어떤 요청들이 차례로 썼는지 연결해야 한다. 이 연결이 없으면 SQL 자체가 틀렸는지 세션이 오염됐는지 구분할 수 없다.

#### DU42-M3 autoCommit 경계 오해

autoCommit은 간단한 도구지만 batch에서는 재시작 기준을 바꾼다. 한 row씩 확정되면 실패 후 앞 row를 건너뛰어야 하고, chunk 단위면 chunk 전체를 다시 실행해야 한다. 이 차이가 문서와 코드에 드러나야 한다.

#### DU42-M4 외부 호출이 트랜잭션 안에 있음

긴 외부 호출은 DB가 해결할 수 없는 지연을 DB 자원에 묶는다. 결제 승인, 파일 전송, 원격 조회는 상태 머신과 멱등 복구로 분리할 수 있는지 먼저 본다.

#### DU42-M5 커넥션 누수와 정상 장기 대여

누수는 반납이 없는 상태이고 장기 대여는 반납이 늦은 상태다. 둘 다 active 증가로 보이지만 수정 위치가 다르다. leak detection과 duration histogram을 함께 둬야 한다.

#### DU42-M6 readOnly routing 부작용

readOnly는 성능 힌트일 수 있지만 프로젝트에서는 라우팅 정책으로도 쓰일 수 있다. 쓰기 후 읽기 일관성이 필요한 요청은 primary read가 필요한지 분명히 해야 한다.

#### DU42-M7 isolation 상향의 비용

격리 수준은 높을수록 안전하다는 단순 축이 아니다. 막으려는 이상 현상과 실제 lock 비용을 함께 봐야 한다. 필요하면 특정 critical section만 잠그는 설계가 더 낫다.

#### DU42-M8 statement timeout 상속

statement timeout은 장애 확산을 줄이는 좋은 도구지만 세션에 남으면 다음 요청의 정책이 된다. 풀 반납 전 복구가 실제로 되는지 통합 테스트가 필요하다.

#### DU42-M9 DB session과 application log 불일치

DB가 보는 세션과 앱이 보는 요청을 연결하지 못하면 원인을 추측하게 된다. backend pid를 로그에 남기면 느린 쿼리와 사용자 요청을 바로 이어 볼 수 있다.

#### DU42-M10 transaction owner 불명확

하위 helper가 commit하면 상위 불변식이 찢어진다. commit 호출은 경계 owner에게 집중시키고, 하위 계층은 같은 트랜잭션에 참여하는 형태로 유지한다.

#### DU42-M11 pool validation 과신

pool validation은 연결이 살아 있는지 보는 신호일 뿐 업무 schema, 권한, search path가 맞다는 증거가 아니다. 첫 업무 쿼리 실패를 validation 성공으로 덮으면 안 된다.

#### DU42-M12 운영 runbook 부재

커넥션 장애는 여러 계층을 건드리므로 runbook이 없으면 각 팀이 자기 계층만 본다. pool, DB, application trace를 어떤 순서로 확인할지 미리 정해야 복구가 빨라진다.

### DU42 면접식 꼬리 질문과 실전 답변

DU42를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. 커넥션 풀 size를 늘리면 pool timeout은 해결되는가?

일시적으로 대기 시간은 줄 수 있지만 원인이 긴 트랜잭션이나 외부 호출 대기라면 해결이 아니다. pool size를 키우면 DB 세션 수와 동시 lock 경쟁도 늘어난다. 먼저 active connection의 체류 시간, DB session state, lock wait, 외부 호출 latency를 같이 봐야 한다.

함정은 active=max라는 숫자를 capacity 부족으로만 보는 것이다. 같은 숫자는 누수, 느린 쿼리, 외부 API 대기, 트랜잭션 안 sleep에서도 나온다.

#### Q2. Connection.close는 실제 TCP 연결 종료인가?

풀을 쓰는 환경에서는 대개 물리 연결 종료가 아니라 대여 반납이다. 애플리케이션은 close를 호출하지만 풀은 상태를 정리한 뒤 idle로 되돌릴 수 있다. 따라서 close 호출 여부와 DB 세션 종료 여부는 같은 질문이 아니다.

함정은 DB에서 세션이 계속 보인다고 close가 안 됐다고 단정하는 것이다. pool idle 세션은 정상일 수 있다.

#### Q3. 세션 상태 오염은 어떻게 재현하는가?

한 요청에서 schema, statement_timeout, isolation 같은 세션 값을 바꾼 뒤 복구하지 않고 반납하게 만든다. 다음 요청이 같은 물리 세션을 빌렸을 때 해당 값이 남아 있으면 재현된다. 테스트는 connection id를 고정하거나 pool size를 1로 줄이면 쉬워진다.

함정은 SQL만 단독 실행해 보고 문제가 없다고 결론 내리는 것이다. 오염은 요청 순서와 세션 재사용이 있어야 보인다.

#### Q4. 트랜잭션 경계 안에서 파일 IO를 하면 왜 위험한가?

파일 IO나 원격 IO는 DB가 제어할 수 없는 지연이다. 그동안 커넥션과 lock, snapshot이 유지되면 다른 요청이 기다린다. 실패하면 DB rollback은 되지만 외부 IO 부작용은 되돌리지 못할 수 있다.

함정은 하나의 메서드 안에 있으면 원자적이라고 믿는 것이다. DB 트랜잭션은 외부 세계를 자동으로 rollback하지 않는다.

#### Q5. idle in transaction은 왜 나쁜 신호인가?

SQL 실행은 끝났지만 transaction이 열린 채 세션이 놀고 있다는 뜻일 수 있다. 이 상태는 lock, snapshot, vacuum/purge, undo retention에 영향을 줄 수 있다. 애플리케이션에서는 응답 생성이나 외부 대기 중일 수 있다.

함정은 idle이라는 단어만 보고 안전하다고 보는 것이다. transaction이 붙은 idle은 자원을 잡고 있을 수 있다.

### DU42 최종 보강: connection state는 요청이 끝나도 흔적을 남길 수 있다

커넥션 풀을 안전하게 쓰려면 반납하면 끝이라는 감각을 버려야 한다. 커넥션은 socket만이 아니라 transaction isolation, autocommit, schema/search path, session variable, temporary table, prepared statement cache 같은 상태를 품을 수 있다. 좋은 풀은 빌려 주기 전후로 이 상태를 정리하지만, 드라이버와 프레임워크 설정이 어긋나면 한 요청의 상태가 다음 요청에 영향을 줄 수 있다. 예를 들어 테스트에서만 `SET search_path`를 바꿨는데 반납 시 초기화가 빠지면, 다음 요청이 다른 schema를 읽는 장애가 생긴다.

검증은 간단하다. 같은 풀에서 커넥션을 두 번 빌려 첫 번째 요청이 session setting을 바꾸고, 두 번째 요청이 기본값으로 돌아왔는지 확인한다. 이 실험을 통해 connection pool이 단순 소켓 캐시가 아니라 session state boundary라는 점을 눈으로 확인할 수 있다. 그래서 커넥션 장애를 볼 때는 pool metric, DB session, 요청 trace, session reset 정책을 한 줄로 이어서 읽어야 한다.

### DU42 운영 복구 훈련: pool timeout 알람을 받은 순간

pool timeout 알람이 울리면 가장 먼저 해야 할 일은 풀 크기를 늘리는 것이 아니라, active connection이 무엇을 기다리는지 분리하는 것이다. 같은 active=max라도 원인은 전혀 다를 수 있다. SQL이 실제로 오래 실행 중일 수도 있고, 이미 SQL은 끝났지만 트랜잭션이 열린 채 외부 API 응답을 기다릴 수도 있으며, 예외 경로에서 connection이 반납되지 않았을 수도 있다. 이 세 경우는 모두 사용자에게 같은 timeout으로 보일 수 있지만 수정 위치는 각각 SQL/인덱스, 트랜잭션 경계, finally cleanup이다.

```text
pool timeout triage
  application metric:
    active=max, pending>0, borrow timeout 증가
  DB session view:
    active query          -> query plan, I/O, lock wait 확인
    idle in transaction   -> 애플리케이션이 트랜잭션을 연 채 다른 일을 기다리는지 확인
    no matching session   -> pool leak, validation, 네트워크 단절, 로그 상관키 확인
  application trace:
    request id -> connection id -> backend pid -> tx begin/end -> external call span
```

이 흐름에서 중요한 것은 숫자 하나를 원인으로 보지 않는 태도다. pool active는 애플리케이션이 커넥션을 빌렸다는 신호이고, DB session view는 그 커넥션이 DB에서 어떤 상태인지 보여 주며, request trace는 왜 그 상태가 오래 지속됐는지 알려 준다. 세 값을 함께 묶지 못하면 운영자는 DB팀에는 느린 쿼리를, 애플리케이션팀에는 커넥션 누수를, 인프라팀에는 네트워크 문제를 동시에 의심하게 된다.

커넥션 경계의 좋은 설계는 복구 때 사람이 덜 추측하게 만든다. 요청 로그에는 적어도 borrow 시작/반납 시각, transaction begin/commit/rollback 시각, DB backend pid 또는 connection id, 외부 호출 span이 같이 있어야 한다. 이 정보가 있으면 "외부 PG timeout 30초 동안 row lock을 잡고 있었다" 또는 "SQL은 20ms였지만 응답 직렬화 중 트랜잭션이 열린 채 남았다"처럼 원인을 좁힐 수 있다. 반대로 이 정보가 없으면 pool size를 키우는 조치가 우연히 증상을 늦출 뿐, 다음 장애에서 더 큰 동시 lock 경쟁을 만들 수 있다.

세션 상태 복구도 같은 방식으로 검증한다. 통합 테스트에서 pool size를 1로 낮추고 첫 요청이 `SET statement_timeout`, `SET search_path`, isolation 변경, readOnly 변경을 수행하게 만든다. 두 번째 요청은 아무 설정도 하지 않고 같은 connection을 빌린 뒤 기본값을 확인한다. 이 테스트가 실패하면 SQL 자체가 맞아도 요청 사이에 실행 조건이 새어 나간다. 실무에서 가장 무서운 커넥션 장애는 항상 재현되는 장애가 아니라 특정 요청 순서에서만 드러나는 장애다.

마지막으로 connection owner를 코드 리뷰 질문으로 고정한다. Controller가 connection을 알 필요는 없고, DAO helper가 commit을 결정해서도 안 되며, service/use case 경계가 업무 원자성을 설명해야 한다. 하위 계층은 같은 transaction에 참여해 SQL을 실행할 수 있지만, 언제 확정하고 언제 되돌릴지는 상위 업무 경계가 정한다. 이 원칙이 지켜지면 pool, session, transaction manager는 서로 다른 도구처럼 보여도 하나의 요청 생명주기로 읽힌다.

대여 시간도 한 덩어리 숫자로 보지 않는다. 커넥션을 빌린 뒤 실제 첫 SQL까지 걸린 시간, SQL 실행 시간, commit/rollback 시간, 응답 생성이나 외부 호출 때문에 반납이 늦어진 시간을 나누어야 한다. 같은 8초 대여라도 7.8초가 SQL이면 실행 계획과 lock을 보고, 7.8초가 외부 API 대기라면 트랜잭션 경계와 outbox/idempotency를 본다. 이 분해가 없으면 `connection borrow time이 길다`는 말은 너무 넓어서 아무도 바로 고칠 수 없는 알람이 된다.

따라서 좋은 관측은 span 이름부터 다르게 둔다. `db.connection.borrow`, `db.transaction.open`, `db.sql.execute`, `db.transaction.commit`, `db.connection.return`이 분리되어 있으면 요청 하나가 DB 자원을 언제 붙잡았는지 보인다. 이 작은 이름 구분은 장애 때 큰 차이를 만든다. 운영자는 풀을 늘릴지, SQL을 고칠지, 외부 호출을 트랜잭션 밖으로 뺄지 훨씬 빨리 판단할 수 있다.

또 하나의 실전 기준은 반납 실패보다 반납 지연을 더 자주 본다는 점이다. 완전한 누수는 leak detection이 비교적 잘 잡지만, 트랜잭션 안에서 응답 body를 만들거나 원격 API를 기다리는 구조는 정상적으로 반납되기 때문에 누수처럼 보이지 않는다. 그러나 피크 시간에는 이 지연이 풀 전체를 점유한다. 그래서 커넥션 분석에서는 "반납했는가"와 "언제 반납했는가"를 반드시 나눠야 한다.

리뷰 체크도 구체적으로 한다. 트랜잭션 안에서 HTTP client, file upload, 대량 JSON 직렬화, 메시지 publish, 사용자 알림 발송을 하는 코드가 있으면 먼저 의심한다. 그 작업이 DB row lock이나 snapshot을 붙잡은 채 실행되어야 하는지 묻고, 그렇지 않다면 DB 상태를 짧게 commit한 뒤 outbox나 후속 worker로 넘기는 구조를 검토한다. connection pool 문제는 설정 파일의 숫자만이 아니라 서비스 메서드 안의 대기 지점에서 만들어진다.

## Spring/JPA/MyBatis transaction boundary

이 절은 `Spring/JPA/MyBatis transaction boundary`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 Spring의 선언적 트랜잭션이 프록시, 전파 속성, rollback rule, JPA EntityManager, MyBatis SqlSession 경계로 어떻게 내려가는지 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 self-invocation과 checked exception rollback 오해이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 작은 서비스 호출 하나에서 시작한다. Controller가 `OrderService` bean을 호출할 때 실제로는 proxy가 먼저 받고, proxy 안의 transaction interceptor가 transaction attribute를 읽은 뒤 실제 target method를 호출한다. JPA repository와 MyBatis mapper는 이 경계 안에서 같은 물리 트랜잭션에 참여할 수 있지만, flush와 SQL 실행을 관측하는 시점은 다르다. 그래서 DU43의 첫 질문은 annotation이 어디 붙었는가가 아니라 호출이 proxy를 통과했는가와 commit owner가 누구인가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | Spring declarative transaction docs: https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Spring rollback rules: https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/rolling-back.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `jvm/spring/spring_transactional.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `interviews/spring-backend-frameworks.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 프록시를 통과한 호출만 경계가 된다

Spring의 `@Transactional`은 메서드 바디 안에 마법처럼 박히는 문장이 아니다. 일반적인 프록시 모드에서는 외부 호출이 프록시를 통과할 때 `TransactionInterceptor`가 먼저 실행되고, 이 인터셉터가 트랜잭션 속성을 읽고 트랜잭션 매니저에게 begin/participate/commit/rollback을 위임한다. 따라서 같은 객체 안에서 `this.inner()`처럼 자기 메서드를 직접 호출하면 프록시를 통과하지 않아 새 트랜잭션 속성이 적용되지 않을 수 있다.

```text
Controller -> OrderService proxy -> TransactionInterceptor
    -> PlatformTransactionManager.getTransaction(REQUIRED)
    -> real OrderService.placeOrder()
        -> JpaRepository / MyBatis mapper participates in bound resource
    -> commit or rollback by rule

OrderService.this.innerTransactional() -> real method directly
    -> no proxy boundary -> no new transactional advice for inner method
```

이 작은 그림이 DU43의 중심이다. annotation 위치보다 호출 경로가 먼저다. annotation은 후보 metadata이고, 프록시를 통과한 호출이 실제 경계가 된다. 이 사실을 잊으면 `REQUIRES_NEW`를 붙였는데도 새 트랜잭션이 열리지 않거나, checked exception이 던져졌는데도 rollback되지 않는 상황을 Spring 버그로 오해한다.
#### DU43-1.1 REQUIRED는 논리 범위와 물리 트랜잭션을 분리한다

`PROPAGATION_REQUIRED`는 현재 트랜잭션이 없으면 새 물리 트랜잭션을 열고, 이미 있으면 그 트랜잭션에 참여한다. 그래서 A 서비스가 트랜잭션을 열고 B, C 서비스를 차례로 호출하면 B와 C의 `@Transactional`은 각각 논리 범위를 만들 수 있지만 DB 입장에서는 같은 커넥션과 같은 물리 트랜잭션으로 실행된다. 로컬 seed `jvm/spring/spring_transactional.md`도 A->B->C 호출에서 최종 commit은 바깥 A 경계에서 한 번만 일어난다고 정리한다.

함정은 내부 C가 성공했다고 곧바로 DB에 확정됐다고 믿는 것이다. 내부 범위는 참여했을 뿐 commit owner가 아니다. C에서 rollback-only가 표시되면 바깥 A가 정상 반환처럼 보여도 최종 commit 시점에 실패가 드러날 수 있다. 운영에서는 이 상황이 '마지막에 UnexpectedRollbackException이 났다'로 보이고, 실제 원인은 더 안쪽 예외 처리에 숨어 있다.

#### DU43-1.2 rollback rule은 예외 종류와 전파 방식에 의존한다

Spring 공식 rollback 문서는 기본 설정에서 runtime unchecked exception과 Error가 rollback-only를 만들고, checked exception은 기본적으로 rollback하지 않는다고 설명한다. 이 규칙은 자바 예외 계층을 모르면 위험하다. 결제 승인 후 checked business exception을 던졌는데 transaction이 commit되어 버리거나, 반대로 잡아서 삼킨 runtime exception 때문에 rollback-only만 남고 바깥에서 commit을 시도하다 실패할 수 있다.

실무 함정은 `catch (Exception e) { log; return failure; }` 형태다. 예외를 잡아 정상 응답으로 바꾸면 트랜잭션 인터셉터는 정상 반환으로 볼 수 있다. rollback이 필요하면 rollbackFor를 명시하거나, 예외를 다시 던지거나, TransactionStatus에 rollback-only를 명시하는 정책을 세워야 한다. 이 결정은 API 응답 모양이 아니라 데이터 불변식 기준으로 해야 한다.

#### DU43-1.3 JPA와 MyBatis는 같은 트랜잭션에 참여하지만 flush 감각이 다르다

JPA는 영속성 컨텍스트가 변경을 모았다가 flush 시점에 SQL로 내보낼 수 있다. MyBatis는 mapper 호출이 SQL 실행에 더 직접적으로 대응한다. 둘 다 Spring 트랜잭션 경계 안에 묶일 수 있지만, '메서드 호출 시점에 DB row가 이미 바뀌었는가'라는 감각은 다르다. JPA에서는 flush 전까지 메모리 상태와 DB 상태가 달라 보일 수 있고, MyBatis에서는 호출 직후 SQL이 날아갔지만 commit 전이라 다른 트랜잭션에는 보이지 않을 수 있다.

이 차이를 섞어 쓰는 서비스에서 장애가 난다. 예를 들어 JPA entity를 수정하고 flush 전에 MyBatis mapper로 같은 row를 조회하면 기대한 변경이 안 보일 수 있다. 반대로 MyBatis update가 실행됐다고 외부 시스템에 성공 이벤트를 보냈는데 commit이 나중에 실패하면 이벤트와 DB가 갈라진다. 그래서 혼합 경계에서는 flush, commit, event publish 시점을 명시해야 한다.

#### DU43-1.4 읽기 전용 트랜잭션은 보안 장벽이 아니라 힌트와 정책의 조합이다

`readOnly=true`는 프레임워크와 드라이버, DB가 최적화하거나 쓰기를 막는 데 사용할 수 있는 힌트로 이해해야 한다. 모든 DB와 드라이버가 같은 방식으로 강제하지 않는다. 따라서 읽기 전용 서비스에서 쓰기 mapper를 호출하면 반드시 실패한다고 가정하면 안 된다. 쓰기를 구조적으로 막고 싶다면 repository 분리, 테스트, DB 권한, review rule을 함께 둬야 한다.

운영 함정은 readOnly를 성능 옵션처럼 무차별 적용하는 것이다. JPA flush mode, connection read-only flag, replica routing 같은 정책이 엮이면 읽기 경로가 replica로 가거나 flush가 억제될 수 있다. 이때 같은 요청 안에서 쓰기를 기대하면 증상이 환경별로 달라진다.

#### DU43-1.5 트랜잭션 경계 안의 외부 호출은 lock 시간을 늘린다

Spring 경계가 편하다고 모든 작업을 하나의 `@Transactional` 안에 넣으면 DB lock과 connection 점유 시간이 외부 API 지연에 묶인다. 결제 승인 API, 파일 업로드, 메시지 브로커 publish, 원격 조회를 트랜잭션 안에서 기다리면 실패 처리 자체는 쉬워 보여도 pool 고갈과 lock wait가 늘어난다. 외부 호출과 DB commit을 원자적으로 묶을 수 없는 경우에는 outbox, 상태 머신, 멱등키, 재조회 기반 복구를 함께 설계해야 한다.

실무에서는 '트랜잭션으로 감싸면 안전하다'가 가장 비싼 오해가 된다. DB와 외부 PG가 같은 2PC에 참여하지 않는 이상, 트랜잭션은 내부 DB 변경만 보호한다. 외부 성공 후 내부 rollback, 내부 commit 후 외부 timeout 같은 회색 상태는 트랜잭션 annotation이 아니라 상태 모델과 대사 절차로 해결한다.

#### DU43-1.6 테스트는 프록시 경유와 rollback 규칙을 직접 때려야 한다

트랜잭션 테스트는 성공 저장만 보면 부족하다. 같은 클래스 self-invocation, public/protected 접근, checked exception, runtime exception, 내부 catch, REQUIRES_NEW, REQUIRED 참여, JPA flush, MyBatis mapper 호출을 분리해 관측해야 한다. 테스트에서 `@Transactional`을 붙이면 테스트 자체가 rollback될 수 있으므로, 실제 HTTP 경계나 별도 트랜잭션에서 commit되는 경로와 다를 수 있다.

함정은 테스트가 데이터 정리를 위해 자동 rollback되는 것을 운영 코드의 rollback 검증으로 착각하는 것이다. 테스트 프레임워크가 만든 트랜잭션과 서비스 프록시가 만든 트랜잭션은 겹칠 수 있지만 같은 목적이 아니다. 운영 경계 검증에는 commit 후 재조회, 예외별 상태 확인, 로그의 transaction name/isolation, connection identity가 필요하다.



### 예외별 상태 trace

```text
case A: RuntimeException escapes service
  proxy sees unchecked exception -> mark rollback -> DB changes undone

case B: CheckedBusinessException escapes service without rollbackFor
  proxy sees checked exception -> default no rollback -> commit can happen

case C: RuntimeException caught and converted to FailureResponse
  proxy sees normal return -> commit unless rollback-only was set explicitly

case D: inner REQUIRED marks rollback-only, outer catches and returns OK
  outer commit attempt -> UnexpectedRollbackException -> caller learns commit did not happen
```

이 trace는 Spring rollback rule을 외우기 위한 표가 아니라 장애 재현 순서다. 운영에서 결제 주문이 `FAILED` 응답을 반환했는데 DB에는 `PAID` row가 남았다면 case B나 C를 의심해야 한다. 반대로 API는 정상 응답을 만들려 했지만 마지막에 rollback-only 때문에 예외가 터졌다면 case D처럼 내부 실패를 잡아 삼킨 코드가 있는지 봐야 한다.
### 관측과 검증

DU43의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU43 verification checklist
1. write tests for runtime exception, checked exception, rollbackFor, and caught exception
2. verify self-invocation does not create a new proxy boundary in proxy mode
3. log TransactionSynchronizationManager resource binding around JPA and MyBatis calls
4. force JPA flush before MyBatis read when mixed access must observe the same change
```

DU43의 검증은 예외 종류와 호출 경로를 직접 때려야 닫힌다. runtime exception, checked exception, rollbackFor, catch 후 정상 반환, self-invocation, JPA flush 전 MyBatis 조회를 각각 fixture로 만든다. PASS는 DB 최종 상태와 transaction log가 설명한 규칙과 일치하는 것이고, FAIL은 annotation은 보이지만 proxy 경유나 rollback rule이 재현되지 않는 것이다.
### 기억할 압축 문장

Spring 트랜잭션 경계는 annotation이 붙은 줄이 아니라 프록시를 통과한 호출과 트랜잭션 매니저가 소유한 commit/rollback 결정으로 생긴다. JPA와 MyBatis는 그 경계에 참여할 수 있지만 flush, SQL 실행, commit 관측 시점은 다르므로 예외 규칙과 호출 경로를 테스트로 고정해야 한다.

#### DU43-S1 등장 배경: 트랜잭션 코드가 비즈니스 코드를 덮던 시절

선언적 트랜잭션은 서비스 메서드마다 try, commit, catch, rollback, finally close를 반복하던 비용에서 등장했다. 반복 코드는 실수하기 쉽고, 예외 종류별 정책이 흩어지며, repository마다 commit을 다르게 호출하게 만든다. Spring은 이 반복을 프록시와 transaction manager로 끌어올려 업무 코드는 도메인 흐름에 집중하게 한다.

하지만 이 배경은 선언적 트랜잭션이 모든 위험을 없앤다는 뜻이 아니다. 반복 코드는 줄었지만 호출 경로, rollback rule, resource binding은 더 보이지 않게 되었다. 그래서 `@Transactional`을 읽을 때는 annotation 글자보다 프록시 진입 여부와 transaction manager 선택을 먼저 확인해야 한다.

#### DU43-S2 프록시 모드는 호출 그래프를 요구한다

프록시 기반 AOP는 외부 객체가 proxy reference를 호출할 때 동작한다. 같은 클래스 내부의 자기 호출은 target object 안에서 바로 이동하므로 advice가 적용되지 않는다. 이 구조는 성능이나 구현 세부가 아니라 트랜잭션 경계의 실제 성립 조건이다.

실무 함정은 `private` 메서드에 붙인 annotation이나 `this.recalculate()` 호출을 보고 새 트랜잭션이 열린다고 믿는 것이다. 코드 리뷰에서는 annotation 위치보다 bean 사이 호출인지, proxy를 통해 들어오는지, 테스트가 그 경로를 실제로 밟는지 확인해야 한다.

#### DU43-S3 JPA flush와 MyBatis 실행 시점의 차이

JPA는 변경 감지를 통해 entity 상태를 모아 두다가 flush 시점에 SQL을 낼 수 있다. MyBatis는 mapper 호출과 SQL 실행의 대응이 더 직접적이다. 둘이 같은 Spring transaction 안에 있어도 관측 가능한 DB 상태는 flush 여부에 따라 달라질 수 있다.

이 차이를 모르면 JPA로 저장한 뒤 MyBatis로 집계하는 코드가 예상과 다르게 동작한다. 해결은 무조건 flush를 남발하는 것이 아니라 같은 aggregate 안에서는 접근 방식을 섞지 않거나, 반드시 섞어야 하면 flush 시점을 명시하고 테스트로 고정하는 것이다.

#### DU43-S4 rollback rule은 업무 실패와 기술 실패를 분리한다

Spring 기본 rollback rule은 unchecked exception 중심이다. 이는 모든 업무 실패가 rollback된다는 뜻이 아니다. 품절, 한도 초과, 외부 승인 거절 같은 business exception이 checked exception으로 설계되어 있다면 기본 규칙에서는 commit될 수 있다.

시니어 실무자는 예외 클래스 이름보다 그 예외가 데이터 불변식을 깨는지 본다. 실패 응답을 저장해야 하는 예외와 전체 변경을 되돌려야 하는 예외는 다르다. rollbackFor는 기술 옵션이 아니라 업무 상태 모델과 함께 결정되어야 한다.

#### DU43-S5 REQUIRES_NEW는 만능 감사 로그 도구가 아니다

감사 로그를 남기려고 내부 메서드에 `REQUIRES_NEW`를 붙이는 패턴이 있다. 이 방식은 바깥 트랜잭션이 rollback되어도 로그는 남길 수 있지만, 자기 호출이면 적용되지 않고, 과도하게 쓰면 커넥션을 추가로 점유하며, 내부 commit이 외부 실패와 다른 시간축을 만든다.

따라서 별도 트랜잭션이 필요한 이유가 명확해야 한다. 실패 감사, outbox 저장, 보상 예약처럼 바깥 rollback과 독립되어야 하는 사실이면 후보가 될 수 있다. 그러나 단순히 중간 저장을 보장하려는 목적이면 도메인 일관성을 찢는지 먼저 봐야 한다.

#### DU43-S6 transaction manager 선택은 resource 선택이다

DataSourceTransactionManager, JpaTransactionManager, JtaTransactionManager는 같은 이름의 트랜잭션 매니저지만 묶는 resource와 동작 경계가 다르다. 단일 JDBC datasource, JPA EntityManager, 여러 XA resource는 같은 문제가 아니다.

멀티 데이터베이스 환경에서 아무 manager나 골라 annotation에 넣으면 해결되는 것이 아니다. 두 DB commit을 원자적으로 묶어야 하는지, 한쪽은 best-effort로 충분한지, outbox나 saga로 풀어야 하는지 먼저 결정해야 한다.

#### DU43-S7 테스트 트랜잭션은 운영 트랜잭션과 다를 수 있다

Spring test에서 `@Transactional`을 붙이면 테스트가 끝날 때 rollback되는 편리한 환경이 생긴다. 그러나 이 트랜잭션은 테스트 격리를 위한 장치이고, 실제 controller-service 경계의 commit 동작과 같다고 볼 수 없다.

운영 경계 검증에는 실제 service proxy를 호출하고 commit 후 별도 transaction에서 재조회하는 테스트가 필요하다. MockMvc 테스트 안의 트랜잭션, 테스트 메서드 트랜잭션, service 트랜잭션이 어떻게 겹치는지 모르면 검증이 거짓 안정감을 준다.

#### DU43-S8 예외를 잡는 위치가 commit을 바꾼다

프록시는 메서드 밖으로 어떤 결과가 나오는지 본다. 내부에서 runtime exception을 잡아 정상 객체로 반환하면 프록시는 성공으로 볼 수 있다. 반대로 내부에서 rollback-only를 표시하고 바깥이 정상 반환하려 해도 commit 시점에 예외가 날 수 있다.

이 때문에 서비스의 catch block은 단순 로깅 코드가 아니다. catch가 데이터 확정 여부를 바꾼다. 리뷰에서는 catch 후 반환값, rollback-only 설정, 재throw 여부, 실패 상태 저장 트랜잭션이 서로 맞는지 확인해야 한다.

#### DU43-S9 관측 가능한 transaction name을 남긴다

긴 장애에서 `transaction active`만 보이면 부족하다. 어떤 service method가 열었는지, propagation은 무엇인지, isolation/readOnly는 무엇인지, connection id는 무엇인지가 보여야 한다. Spring은 transaction name과 synchronization 정보를 관리하므로 적절한 로깅 지점에서 이를 남길 수 있다.

이 정보가 있으면 A->B->C 호출에서 어디가 owner인지, 내부 참여자인지, 어떤 mapper 호출이 같은 resource를 썼는지 추적할 수 있다. observability 없는 transaction은 정상일 때만 편하고 장애 때는 검은 상자가 된다.

#### DU43-S10 최종 설계 질문

Spring/JPA/MyBatis 경계 설계는 `@Transactional`을 어디에 붙일까가 아니라 어떤 use case가 하나의 commit decision을 가져야 하는가에서 시작한다. 그 다음 프록시 경유, repository 접근 방식, flush, rollback rule, 외부 호출 분리, event 발행 시점을 배치한다.

이 질문 순서가 바뀌면 annotation이 목적을 이끈다. 목적은 도메인 불변식이고, annotation은 그 불변식을 DB transaction으로 보호하는 수단이다. 그래서 DU43의 결론은 프레임워크 사용법이 아니라 owner와 관측 가능한 경계다.

### DU43 운영 케이스 매트릭스

DU43의 본문을 실제 장애 대응에 쓰려면 `Spring/JPA/MyBatis transaction boundary`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| self-invocation | 같은 bean 내부에서 this.method 호출 | 프록시 advice가 적용되지 않는 경로로 판정한다 | proxy class log, transaction active check |
| checked exception commit | 업무 checked exception이 밖으로 나감 | rollbackFor가 없으면 commit될 수 있음을 테스트한다 | exception type별 DB 상태 |
| catch 후 정상 반환 | runtime exception을 잡고 실패 응답 반환 | 프록시는 정상 반환으로 볼 수 있음을 확인한다 | catch block review, rollback-only 여부 |
| JPA 수정 후 MyBatis 조회 | flush 전 mapper가 같은 row를 조회 | flush 시점이나 접근 방식 통일을 결정한다 | SQL log, flush count |
| REQUIRES_NEW 남용 | 감사 로그를 항상 새 트랜잭션으로 저장 | 독립 확정이 업무적으로 필요한지 검토한다 | connection usage, nested tx test |
| 테스트 rollback 착시 | 테스트는 통과하지만 운영 commit이 다름 | 테스트 트랜잭션과 서비스 트랜잭션을 분리한다 | commit 후 별도 transaction 재조회 |
| readOnly 쓰기 기대 | readOnly 안에서 mapper update 호출 | 힌트/강제/라우팅 정책을 프로젝트 기준으로 확인한다 | DB error, routing log, flush mode |
| 동적 transaction manager | 멀티 DB에서 manager를 런타임 선택 | resource 경계와 원자성 요구를 먼저 정한다 | manager bean, datasource route log |
| rollback-only 숨김 | 내부 참여 트랜잭션이 rollback-only 표시 | 외부 정상 반환 시 UnexpectedRollbackException 가능성을 확인한다 | transaction debug log |
| event 발행 시점 | DB commit 전에 메시지 발행 | commit 성공 후 발행 또는 outbox를 검토한다 | message log vs DB commit log |
| lazy loading 경계 | transaction 밖에서 entity graph 접근 | 조회 모델을 transaction 안에서 완성한다 | LazyInitializationException, SQL log |
| 최종 runbook | 트랜잭션 장애 원인 불명 | 호출 경로, 예외 규칙, resource binding 순서로 좁힌다 | trace id + tx name + SQL log |

#### DU43-M1 self-invocation

같은 클래스 내부 호출은 annotation이 있어도 프록시를 통과하지 않을 수 있다. 해결은 메서드 분리, self proxy 주입, AspectJ mode 검토처럼 호출 경로를 바꾸는 쪽이지 annotation을 더 붙이는 것이 아니다.

#### DU43-M2 checked exception commit

checked exception은 업무 오류 표현에 자주 쓰이지만 Spring 기본 rollback과 어긋날 수 있다. 어떤 업무 실패가 데이터를 되돌려야 하는지 명시해야 한다.

#### DU43-M3 catch 후 정상 반환

catch는 로깅 이상의 의미를 가진다. 예외를 잡는 순간 transaction interceptor가 보는 결과가 바뀐다. 실패 상태 저장과 rollback 중 무엇이 목적인지 분리해야 한다.

#### DU43-M4 JPA 수정 후 MyBatis 조회

JPA 영속성 컨텍스트와 MyBatis SQL 실행은 관측 시점이 다르다. 혼합할 때는 flush를 명시하거나 같은 repository 기술 안에서 닫는 편이 안전하다.

#### DU43-M5 REQUIRES_NEW 남용

새 트랜잭션은 바깥 rollback과 다른 시간축을 만든다. 감사 로그에는 유용할 수 있지만 중간 업무 데이터를 확정하는 데 쓰면 일관성을 찢을 수 있다.

#### DU43-M6 테스트 rollback 착시

테스트의 자동 rollback은 데이터 정리 장치다. 운영 코드의 rollback 정책을 검증하려면 실제 service proxy 호출과 commit 후 재조회가 필요하다.

#### DU43-M7 readOnly 쓰기 기대

readOnly는 환경에 따라 최적화 힌트, flush 억제, replica routing과 연결된다. 쓰기를 막고 싶다면 테스트와 권한까지 같이 둬야 한다.

#### DU43-M8 동적 transaction manager

manager 선택은 기술 옵션이 아니라 어떤 resource를 하나의 commit decision에 묶을지의 결정이다. 두 DB 원자성이 필요하면 다른 설계가 필요할 수 있다.

#### DU43-M9 rollback-only 숨김

내부 실패를 잡아도 rollback-only가 남으면 바깥 commit에서 실패한다. 이 예외는 호출자가 commit 성공으로 오해하지 않도록 하는 안전장치다.

#### DU43-M10 event 발행 시점

트랜잭션 안에서 외부 메시지를 바로 보내면 DB rollback과 메시지 발행이 갈라질 수 있다. afterCommit hook이나 outbox가 필요한지 본다.

#### DU43-M11 lazy loading 경계

JPA의 lazy loading은 transaction/session 생존과 연결된다. controller에서 늦게 entity를 걷는 구조는 경계를 흐린다.

#### DU43-M12 최종 runbook

Spring 트랜잭션 장애는 annotation만 보면 안 된다. 실제 proxy 경유, exception escape, resource binding, commit owner를 순서대로 확인해야 한다.

### DU43 면접식 꼬리 질문과 실전 답변

DU43를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. 왜 self-invocation은 @Transactional이 안 먹을 수 있는가?

프록시 모드에서는 외부 호출이 proxy를 통과할 때 advice가 붙는다. 같은 객체 안에서 this로 이동하면 proxy 밖에서 target method를 직접 부르는 셈이므로 새 transaction attribute가 적용되지 않을 수 있다.

함정은 annotation이 메서드에 보이면 런타임 경계도 반드시 생긴다고 믿는 것이다.

#### Q2. checked exception은 항상 rollback되는가?

Spring 기본 규칙에서는 checked exception이 기본 rollback 대상이 아니다. rollback이 필요하면 rollbackFor를 지정하거나 예외 계층을 바꿔야 한다. 업무 실패를 checked로 표현할 때는 DB 확정 여부를 따로 결정해야 한다.

함정은 실패 응답을 던졌으니 당연히 DB도 되돌아간다고 믿는 것이다.

#### Q3. JPA와 MyBatis를 같은 트랜잭션에서 섞어도 되는가?

가능은 하지만 flush와 SQL 실행 관측 시점이 다르다. JPA 변경이 아직 flush되지 않았는데 MyBatis가 같은 row를 읽으면 기대와 다를 수 있다. 혼합이 필요하면 flush 시점과 접근 책임을 명시한다.

함정은 같은 DB 트랜잭션이면 모든 계층이 같은 메모리 상태를 본다고 생각하는 것이다.

#### Q4. 테스트의 @Transactional은 운영 검증인가?

테스트 rollback은 격리 장치일 뿐 운영 commit 경계를 그대로 검증하지 않는다. 운영과 같은 service proxy를 호출하고 commit 후 별도 transaction에서 재조회해야 실제 확정 여부를 알 수 있다.

함정은 테스트 종료 rollback을 서비스 rollback 규칙의 증거로 쓰는 것이다.

#### Q5. 내부 예외를 catch하면 어떤 일이 생기는가?

프록시는 메서드 밖으로 나온 결과를 본다. runtime exception을 잡아 정상 응답을 반환하면 commit될 수 있고, 내부에서 rollback-only가 표시되면 바깥 commit 시 UnexpectedRollbackException이 날 수 있다.

함정은 catch block을 단순 로깅으로 보는 것이다. catch는 transaction outcome을 바꾼다.

### DU43 마지막 close scenario: 성공 응답과 rollback-only의 충돌

서비스 A가 B를 호출하고, B 안에서 C가 재고 차감에 실패했지만 예외를 잡아 `OUT_OF_STOCK` 결과 객체로 바꿨다고 하자. C가 참여 트랜잭션에 rollback-only를 남겼다면 A는 정상 응답을 만들려 해도 최종 commit에서 실패할 수 있다. 이 장면은 Spring이 이상하게 동작한 것이 아니라, 내부 논리 트랜잭션이 전체 물리 트랜잭션을 더 이상 commit할 수 없다고 표시했기 때문이다.

```text
A proxy enters REQUIRED -> physical tx opened
B participates REQUIRED -> no new commit owner
C participates REQUIRED -> marks rollback-only after stock failure
B catches exception -> returns business failure object
A tries normal return -> transaction manager attempts commit
commit sees rollback-only -> rollback + UnexpectedRollbackException
```

이 시나리오의 교훈은 실패 응답을 만들고 싶은 코드와 DB를 되돌리고 싶은 코드가 같은 경계 안에서 충돌할 수 있다는 점이다. 실패 응답을 정상적으로 저장해야 한다면 실패 상태 저장 트랜잭션과 실패를 유발한 업무 트랜잭션을 분리해야 할 수 있다. 반대로 전체를 되돌리는 것이 맞다면 예외를 잡아 정상 반환하지 말고 호출자에게 rollback 사실이 보이게 해야 한다.

### DU43 source boundary note

이 절에서 Spring 공식 문서는 프록시 기반 선언적 트랜잭션과 rollback rule의 사실 근거로만 사용한다. JPA와 MyBatis를 함께 쓰는 운영 판단은 로컬 seed와 일반적인 실행 모델에 근거한 설계 추론이며, 실제 프로젝트에서는 SQL log와 transaction synchronization 관측으로 다시 닫아야 한다.

### DU43 최종 보강: 트랜잭션 경계는 코드 위치가 아니라 호출 경로로 결정된다

`@Transactional`이 붙은 메서드를 봤다고 해서 트랜잭션이 항상 열린다고 생각하면 안 된다. Spring의 일반적인 선언적 트랜잭션은 proxy를 통과한 호출에 advice를 적용한다. 같은 객체 내부에서 자기 메서드를 직접 호출하면 proxy를 지나지 않아 기대한 경계가 생기지 않을 수 있다. 이 문제는 초급자 실수처럼 보이지만, 서비스가 커지고 helper 메서드가 많아지면 시니어도 놓치기 쉽다.

운영 장애에서는 이 차이가 왜 rollback이 안 됐는가라는 질문으로 나타난다. 로그에는 예외가 찍혔고 메서드에는 annotation이 있지만, 실제 호출은 proxy 밖에서 내부 메서드를 직접 탔다면 transaction manager는 그 경계를 열지 않았다. 검증은 테스트에서 proxy를 통과한 호출과 self-invocation을 나란히 실행하고, 실제 connection autocommit 상태와 commit/rollback 결과를 확인하는 방식으로 닫는다. 이 습관이 있어야 Spring/JPA/MyBatis 경계를 문법이 아니라 실행 경로로 읽을 수 있다.

### 추가 판정 질문: commit owner는 누구인가

Spring/JPA/MyBatis가 섞인 서비스에서는 어느 코드가 commit owner인지 분명해야 한다. Repository나 mapper가 SQL을 실행하더라도 commit을 결정하는 주체는 보통 transaction manager다. JPA flush는 SQL을 DB에 보낼 수 있지만 그것이 commit을 뜻하지는 않는다. MyBatis mapper가 update count를 반환해도 바깥 transaction이 rollback되면 결과는 사라진다.

이 구분이 없으면 개발자는 SQL 로그에 UPDATE가 찍힌 것만 보고 성공했다고 생각한다. 검증은 의도적으로 예외를 발생시킨 뒤 SQL 로그, transaction synchronization, 최종 DB row를 함께 확인하는 방식으로 닫는다. commit owner를 말할 수 있어야 프레임워크 경계를 안전하게 다룰 수 있다.

### DU43 운영 복구 훈련: rollback이 안 된 것처럼 보이는 장애

Spring 트랜잭션 장애는 대개 "분명히 예외가 났는데 왜 DB가 남았지" 또는 "성공 응답을 만들었는데 왜 commit에서 터졌지"라는 말로 시작한다. 이때 바로 annotation 위치를 고치기보다 호출 경로를 먼저 그려야 한다. 외부 bean에서 proxy를 통과했는지, 같은 bean 내부 `this` 호출인지, 예외가 proxy 밖으로 나갔는지, 내부에서 잡혀 정상 반환으로 바뀌었는지, rollback-only가 남았는지 순서대로 본다.

```text
transaction incident trace
  request R-77
    -> OrderService proxy entered
      -> createOrder target method
        -> reserveStock() self-invocation?        check proxy boundary
        -> paymentMapper.insert()                 check resource binding
        -> jpaEntity changed, not flushed yet      check flush timing
        -> checked BusinessException thrown        check rollback rule
        -> catch block returns failure response    check rollback-only and commit outcome
```

이 trace에서 SQL 로그만 보면 일부 update와 insert가 보일 수 있다. 그러나 SQL 실행은 commit이 아니다. JPA flush는 변경을 DB에 보낼 수 있지만 transaction이 끝나기 전까지 확정은 아니다. MyBatis update count도 같은 물리 트랜잭션이 rollback되면 사라진다. 그래서 장애 분석에서는 SQL이 찍혔는가보다 transaction manager가 어떤 결정을 내렸는가를 봐야 한다.

테스트도 운영 호출 경로를 닮아야 한다. 테스트 클래스에 붙은 `@Transactional`은 테스트 종료 후 rollback으로 DB를 청소하는 데 유용하지만, 서비스 proxy가 운영에서 commit하는 방식과 다를 수 있다. 운영 검증을 하려면 service bean을 실제 proxy로 호출하고, 호출이 끝난 뒤 별도 transaction 또는 새 connection에서 row를 다시 읽어야 한다. 같은 테스트 안에서 persistence context가 가진 entity를 보는 것은 commit 결과를 보는 것이 아니다.

JPA와 MyBatis를 섞는 경우에는 flush 경계가 특히 중요하다. JPA가 entity를 변경했지만 아직 flush하지 않은 상태에서 MyBatis mapper가 같은 row를 조회하면 DB에는 이전 값이 있을 수 있다. 반대로 MyBatis가 직접 update한 row를 JPA persistence context가 이미 들고 있으면 entity는 오래된 값을 계속 보여 줄 수 있다. 같은 DB transaction 안이라는 사실은 commit 운명을 공유한다는 뜻이지, 모든 접근 계층의 메모리 view가 자동으로 동기화된다는 뜻은 아니다.

외부 메시지 발행 시점도 commit owner 질문으로 돌아온다. DB commit 전에 Kafka 메시지를 보내면 DB rollback 후에도 메시지가 남을 수 있고, commit 후 메시지 전송이 실패하면 DB에는 성공이 있지만 소비자는 모를 수 있다. 그래서 afterCommit hook, outbox table, message relay, idempotent consumer를 비교해야 한다. 트랜잭션 경계를 잘 안다는 것은 annotation 옵션을 많이 아는 것이 아니라, DB 확정과 외부 부작용 사이의 틈을 어디서 보완할지 설명할 수 있다는 뜻이다.

### DU43 마지막 운영 연습: 장애 로그 한 줄에서 경계를 복원하기

운영 로그에 `UnexpectedRollbackException`이 한 줄 찍혔다고 가정하자. 이 예외를 보는 순간 필요한 질문은 "왜 Spring이 예외를 던졌나"가 아니라 "어느 내부 참여자가 rollback-only를 표시했나"다. 먼저 trace id로 service call graph를 복원하고, 각 method의 propagation, 예외 종류, catch 위치, DB write 여부를 시간순으로 적는다. 그다음 transaction debug log에서 physical transaction이 하나였는지, `REQUIRES_NEW`가 실제로 새 connection을 잡았는지, commit 시점에 rollback-only가 있었는지 본다.

이 훈련은 코드 리뷰에도 그대로 적용된다. `@Transactional`이 붙은 public method만 세는 리뷰는 부족하다. 실제로 proxy를 통과하는 진입점은 어디인지, 내부 private/helper 호출이 새 경계를 기대하고 있지는 않은지, checked exception이 업무 실패인데 commit을 허용하는지, catch 후 정상 반환이 commit을 만들지 확인해야 한다. 이런 질문을 매번 던지면 트랜잭션은 프레임워크 마법이 아니라 호출 경로와 예외 정책으로 결정되는 실행 계약이 된다.
