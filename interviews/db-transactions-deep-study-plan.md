# DB 트랜잭션, 중복 실행, 금액 계산, 은행 시나리오 심화 정리 계획

이 문서는 DB 트랜잭션, MVCC, 격리 수준, ACID, 따닥 이슈, 금액 계산, 은행/결제 트랜잭션 시나리오를 장문 학습 문서로 승격하기 전의 계획서입니다.
목표는 긴 글을 많이 만드는 것이 아니라, 학습자가 나중에 트랜잭션 경계, 동시성 실패, 금액 불변식, 외부 금융 시스템 실패를 자기 말로 다시 설명하고 직접 검증할 수 있는 문서 세트를 만드는 것입니다.

이번 계획은 `multi-agent`, `dialectic-kernel`, `review-kernel`, `study-explanation` 기준으로 먼저 세운 것입니다.
따라서 아래 내용은 본문 작성 전의 고정 계약이며, 실제 작성 단계에서는 이 계획을 WORK 문서와 함께 다시 읽고 각 단위의 길이와 품질을 기계적으로 검증해야 합니다.

## 1. 요청 정규화

사용자 요청은 다음처럼 정규화합니다.

- goal: DB 트랜잭션과 금융/결제 실무 시나리오를 깊게 이해할 수 있는 장문 학습 문서 세트를 만든다.
- scope: `interviews/` 루트의 정식 학습 자산, 기존 `database-storage-search-nosql.md`, `spring-backend-frameworks.md`, `distributed-systems-architecture.md`, `jvm/spring/spring_transactional.md`, `knowledge/cards/*`, `domains/firmbanking/*`, `math/*`는 원재료와 연결 자산으로 본다.
- mode: planning first, then explain/execute.
- finish: 실제 본문 작성 tranche에서는 final review, 기계 검증, path-limited commit까지 닫는다. 이 계획 문서는 planning tranche다.
- must_keep: 각 주요 개념/시나리오는 최소 15,000자 이상으로 작성한다. 단, `undo log`, `requestHash`, `next-key lock` 같은 하위 용어는 독립 15,000자 단위가 아니라 상위 주요 단위의 필수 구성요소로 둔다.
- extra_checks: MVCC/격리 수준은 PostgreSQL과 MySQL InnoDB 차이를 분리한다. `따닥 이슈`는 UI debounce, HTTP 재시도, DB 원자 선출, 외부 금융 부작용을 섞지 않는다. 금액 계산은 `BigDecimal 쓰기`로 끝내지 않는다. 은행 시나리오는 단순 mutable balance가 아니라 원장, 상태 전이, 정산, 대사를 포함한다.

## 2. 문서 배치 결정

기존 `database-storage-search-nosql.md`는 원문 배치본에 가깝고, 이미 여러 DB/검색/NoSQL 주제가 섞여 있습니다.
여기에 15,000자 이상 장문 섹션을 대량 삽입하면 원재료와 정식 monograph가 한 파일 안에서 섞입니다.
따라서 새 장문 세트는 `interviews/` 루트에 별도 파일로 둡니다.

권장 파일 구조는 다음입니다.

1. `db-transactions-deep-study.md`

    전체 지도, 읽는 순서, 핵심 용어, 각 상세 문서 링크를 담는 허브 문서입니다.
    이 파일은 각 장문 문서를 대체하지 않고, `어떤 순서로 읽어야 하는가`와 `어떤 시나리오에서 어느 문서로 가야 하는가`를 고정합니다.

2. `db-transactions-core-concepts.md`

    DB 내부 개념을 다룹니다.
    트랜잭션 생명주기, ACID, WAL/undo/redo, MVCC, 격리 수준, 락, 2PL, deadlock, 제약조건과 불변식을 포함합니다.

3. `db-transactions-application-boundaries.md`

    애플리케이션 경계에서 트랜잭션 개념이 어떻게 깨지거나 보강되는지 다룹니다.
    따닥/중복 요청, 멱등성, winner 선출, 금액 계산, Spring `@Transactional` 경계를 포함합니다.

4. `db-transactions-financial-scenarios.md`

    은행/결제 흐름을 다룹니다.
    이체, 결제 승인/매입/취소/환불, 외부 응답 유실, 정산, 대사, outbox/inbox, saga, 2PC/XA/JTA를 포함합니다.

이 구조의 장점은 세 가지입니다.
첫째, DB 엔진 내부 설명과 금융 도메인 상태 전이가 서로를 오염시키지 않습니다.
둘째, 각 파일 단위로 15,000자 검증과 중복 문단 검사를 수행하기 쉽습니다.
셋째, 기존 원문 배치본과 정식 심화 문서의 역할이 분리됩니다.

## 3. 전체 학습 순서

학습자는 아래 순서로 읽어야 합니다.
순서를 바꾸면 `멱등성`, `정산`, `사후 대사` 같은 큰 시나리오를 DB 트랜잭션 하나로 잘못 이해할 위험이 큽니다.

1. row, table, index, page, buffer pool 같은 저장 단위
2. autocommit, begin, commit, rollback, savepoint
3. ACID와 도메인 불변식
4. WAL, redo, undo, fsync, crash recovery
5. MVCC와 snapshot/read view/tuple visibility
6. SQL 표준 격리 수준과 실제 PostgreSQL/MySQL InnoDB 차이
7. row lock, gap lock, next-key lock, predicate lock, advisory lock, deadlock
8. optimistic lock, version column, unique constraint, compare-and-swap
9. idempotency key, request hash, winner election, response replay
10. money amount representation, scale, rounding, allocation, reconciliation
11. Spring transaction proxy, propagation, rollback rules, thread-bound resource
12. ledger, authorization, settlement, cancellation, refund, chargeback, reconciliation
13. outbox, inbox, claim lease, saga, 2PC/XA/JTA, external timeout recovery

## 4. 주요 작성 단위 레지스트리

아래 12개 단위가 실제 15,000자 이상 작성 대상입니다.
검증 단위는 각 파일의 `##` 주요 단위입니다.
각 `##` 본문은 heading line을 제외하고 Unicode 문자 수 15,000자 이상이어야 하며, 표와 코드 블록은 본문 문자 수에 포함합니다.

| id | 대상 파일 | `##` 제목 후보 | 최소 본문 길이 | 상태 |
|---|---|---|---:|---|
| T01 | `db-transactions-core-concepts.md` | 트랜잭션 생명주기와 ACID | 15,000자 | 계획 완료 |
| T02 | `db-transactions-core-concepts.md` | WAL, undo, redo, 지속성 | 15,000자 | 계획 완료 |
| T03 | `db-transactions-core-concepts.md` | MVCC와 snapshot | 15,000자 | 계획 완료 |
| T04 | `db-transactions-core-concepts.md` | 격리 수준과 anomaly | 15,000자 | 계획 완료 |
| T05 | `db-transactions-core-concepts.md` | 락, 2PL, deadlock | 15,000자 | 계획 완료 |
| T06 | `db-transactions-core-concepts.md` | consistency, 제약조건, 도메인 불변식 | 15,000자 | 계획 완료 |
| T07 | `db-transactions-application-boundaries.md` | 따닥, 중복 요청, 멱등성 | 15,000자 | 계획 완료 |
| T08 | `db-transactions-application-boundaries.md` | 금액 계산과 반올림 정책 | 15,000자 | 계획 완료 |
| T09 | `db-transactions-application-boundaries.md` | Spring 트랜잭션 경계 | 15,000자 | 계획 완료 |
| T10 | `db-transactions-financial-scenarios.md` | 은행 이체와 결제 상태 머신 | 15,000자 | 계획 완료 |
| T11 | `db-transactions-financial-scenarios.md` | 외부 응답 유실과 정산/대사 | 15,000자 | 계획 완료 |
| T12 | `db-transactions-financial-scenarios.md` | 분산 트랜잭션, saga, outbox | 15,000자 | 계획 완료 |

`db-transactions-deep-study.md` 허브 문서는 15,000자 이상 필수 단위가 아닙니다.
허브는 길이보다 정확한 지도, 링크, 용어 경계, 읽기 순서가 중요합니다.

## 5. 단위별 teaching spine

### T01. 트랜잭션 생명주기와 ACID

읽은 뒤 학습자는 `트랜잭션은 여러 SQL을 감싸는 문법이 아니라, 실패 시 어느 상태까지 보존할지를 정하는 작업 단위`라고 설명할 수 있어야 합니다.
가장 큰 learner gap은 ACID의 네 글자를 외우면서도 `autocommit=true` 단일문, 명시적 `BEGIN`, savepoint, commit 이후 crash 사이의 차이를 흐리게 이해하는 것입니다.

첫 벽돌은 계좌 A에서 B로 1,000원을 옮기는 작은 SQL trace입니다.

```sql
BEGIN;
UPDATE accounts SET balance = balance - 1000 WHERE account_id = 'A';
UPDATE accounts SET balance = balance + 1000 WHERE account_id = 'B';
COMMIT;
```

이 trace는 곧바로 `A만 차감되고 B는 증가하지 않은 상태`가 언제 관측될 수 있는지, 언제 rollback으로 사라지는지, commit 뒤에는 어떤 복구 책임이 생기는지로 이어져야 합니다.
본문은 `입력 -> SQL 실행 -> undo 가능 상태 -> commit 결정 -> durable 상태 -> 소비자` 흐름으로 씁니다.

반드시 닫을 지식은 다음입니다.

- autocommit과 명시적 트랜잭션의 차이
- transaction boundary와 connection boundary
- savepoint와 전체 rollback의 차이
- ACID의 `Consistency`는 DB가 모든 비즈니스 의미를 자동으로 맞춘다는 뜻이 아니라, 제약조건과 애플리케이션 불변식이 성공 전후에 보존되어야 한다는 뜻
- isolation은 `혼자 실행되는 척`이 아니라, 실제로는 특정 anomaly를 허용하거나 막는 동시성 계약
- durability는 commit 응답과 디스크 flush, WAL 정책, replication 정책이 어떻게 연결되는지까지 봐야 한다는 점

15,000자 이상을 채우는 논리는 다음 순서입니다.
정의 목록으로 시작하지 않고, 작은 이체 trace에서 시작해 `중간 실패`, `rollback`, `commit`, `crash`, `constraint violation`, `다른 세션의 읽기`를 하나씩 추가합니다.
그 뒤 ACID 네 속성을 각각 같은 trace 위에서 다시 읽고, 마지막에는 `이 설명이 은행 이체 전체를 해결하지 못하는 이유`를 밝힙니다.
은행/결제에서는 외부 은행 호출, 응답 유실, 정산 파일, 대사가 별도 문제로 남기 때문입니다.

검증 anchor는 PostgreSQL/MySQL에서 `BEGIN`, `ROLLBACK`, `COMMIT`, `SAVEPOINT`를 직접 실행하는 2-session SQL입니다.
PASS는 rollback 전 변경이 commit 뒤 남지 않는 것, savepoint rollback이 부분 변경만 되돌리는 것, constraint violation이 트랜잭션 상태를 어떻게 바꾸는지 관측되는 것입니다.
FAIL은 commit 전 변경을 durable 결과처럼 설명하거나, ACID의 C를 DB 내부 제약조건만으로 축소하는 것입니다.

### T02. WAL, undo, redo, 지속성

읽은 뒤 학습자는 `커밋은 데이터 페이지가 이미 완전히 쓰였다는 뜻이 아니라, 복구할 수 있는 로그 순서가 닫혔다는 뜻`이라고 설명할 수 있어야 합니다.
learner gap은 `DB가 변경된 row를 바로 파일에 쓴다`는 단순 모델입니다.

첫 벽돌은 `balance=1000 -> 700` 변경이 memory page, undo record, redo/WAL record, commit record, page flush를 지나는 시간표입니다.

```text
t0 row: accounts(A).balance = 1000
t1 transaction begins
t2 undo record: old balance = 1000
t3 buffer page: balance becomes 700 in memory
t4 redo/WAL record: page change can be replayed
t5 commit record is flushed according to durability policy
t6 data page is flushed later
```

본문은 undo와 redo를 같은 말로 뭉개지 않아야 합니다.
undo는 읽기 일관성과 rollback에, redo/WAL은 crash recovery와 durability에 더 직접 연결됩니다.
엔진마다 구현은 다르므로 PostgreSQL WAL과 MySQL InnoDB redo/undo를 같은 abstraction으로만 설명한 뒤, 실제 차이는 별도 비교로 내려갑니다.

반드시 닫을 지식은 다음입니다.

- buffer pool/page cache와 디스크 파일의 차이
- write-ahead logging이 필요한 이유
- commit record와 fsync 정책
- crash가 commit 전, commit 후, data page flush 전, checkpoint 전 어디서 나는지에 따른 복구 차이
- group commit이 성능과 latency를 어떻게 바꾸는지
- undo가 MVCC snapshot과 rollback에 어떻게 참여하는지
- durable write와 replicated durability는 다른 문제라는 점

15,000자 이상을 채우는 논리는 `변경된 값 하나`를 따라가는 방식입니다.
처음에는 단일 row update, 다음에는 두 row 이체, 다음에는 group commit, 다음에는 장애 시점별 recovery matrix, 마지막에는 PostgreSQL/MySQL 차이와 운영 설정을 붙입니다.
이 장은 장애 복구를 실제로 상상하게 만드는 것이 핵심이므로 timeline과 before/after 표가 반드시 필요합니다.

검증 anchor는 공식 문서와 작은 실험입니다.
실제 crash 실험은 환경 비용이 크므로 문서 작성 시에는 PostgreSQL/MySQL 공식 복구 설명, 설정값, 로그 용어를 1차 자료로 삼고, 로컬에서는 `transaction commit/rollback` 관측과 가능하면 Docker 기반 crash lab을 별도 선택 실험으로 둡니다.

### T03. MVCC와 snapshot

읽은 뒤 학습자는 `MVCC는 row를 복사해 읽기만 빠르게 만드는 마법이 아니라, 각 트랜잭션이 볼 수 있는 버전 집합을 정하는 규칙`이라고 설명할 수 있어야 합니다.
learner gap은 `읽기는 락을 안 잡는다`를 `쓰기 충돌도 없다`로 오해하는 것입니다.

첫 벽돌은 같은 row의 여러 버전입니다.

```text
accounts(A)
  v1: balance=1000, created_by=T10, deleted_by=T20
  v2: balance=700,  created_by=T20, deleted_by=null

T15 snapshot sees v1
T25 snapshot sees v2
```

이 단순 그림으로 PostgreSQL tuple visibility와 InnoDB undo 기반 consistent read를 구분합니다.
본문은 `snapshot이 언제 만들어지는가`, `그 snapshot이 어떤 version을 보게 하는가`, `내 트랜잭션의 쓰기는 왜 보이는가`, `오래된 version은 언제 청소되는가`를 시간 순서로 설명해야 합니다.

반드시 닫을 지식은 다음입니다.

- 다중 버전 동시성 제어(MVCC)의 한국어 풀이와 원어 병기
- statement snapshot과 transaction snapshot
- PostgreSQL tuple visibility, xmin/xmax 계열 개념은 공식 문서 범위에서 필요한 만큼만 사용
- MySQL InnoDB read view와 undo log 기반 consistent read
- read-only query가 writer를 막지 않는 대표 장점
- writer-writer conflict, lost update, write skew는 별도 문제로 남는다는 점
- vacuum/purge가 늦어질 때 생기는 비용

15,000자 이상을 채우는 논리는 `두 세션이 같은 row를 읽고 쓰는 trace`를 PostgreSQL 관점과 MySQL 관점으로 나눠 반복하는 것입니다.
먼저 엔진 중립 모델을 세우고, 곧바로 엔진별 차이를 붙입니다.
이 장은 특히 `MVCC는 격리 수준의 구현 수단이지 ACID 전체의 동의어가 아니다`를 반복해서 바로잡아야 합니다.

검증 anchor는 2-session SQL입니다.
세션 A가 transaction을 열고 같은 row를 읽는 동안 세션 B가 update/commit한 뒤, 세션 A의 isolation level별 재조회 결과를 비교합니다.
PASS는 statement 단위 snapshot과 transaction 단위 snapshot의 차이가 실제 출력으로 보이는 것입니다.

### T04. 격리 수준과 anomaly

읽은 뒤 학습자는 `격리 수준은 높은/낮은 등급표가 아니라, 어떤 동시 실행 결과를 허용할지 정하는 계약`이라고 설명할 수 있어야 합니다.
learner gap은 SQL 표준 anomaly 표를 외운 뒤 PostgreSQL/MySQL 실제 동작까지 같은 표로 단정하는 것입니다.

첫 벽돌은 두 세션 schedule입니다.

```text
Session A                         Session B
BEGIN;                            BEGIN;
SELECT balance FROM A; -- 1000
                                   UPDATE A SET balance = 700;
                                   COMMIT;
SELECT balance FROM A; -- ?
COMMIT;
```

이 schedule 하나를 `READ COMMITTED`, `REPEATABLE READ`, `SERIALIZABLE`에서 반복합니다.
그 뒤 dirty read, non-repeatable read, phantom read, lost update, write skew를 각각 별도 schedule로 보여 줍니다.

반드시 닫을 지식은 다음입니다.

- SQL 표준의 dirty read, non-repeatable read, phantom read, serialization anomaly
- PostgreSQL은 `READ UNCOMMITTED`를 `READ COMMITTED`처럼 취급한다는 점
- PostgreSQL `REPEATABLE READ`가 표준 최소 요구보다 강하게 phantom을 막는 경우가 있다는 점
- PostgreSQL `SERIALIZABLE`은 실패 가능한 거래로 보고 retry가 필요하다는 점
- MySQL InnoDB 기본 `REPEATABLE READ`, consistent read, locking read, next-key lock 차이
- `SELECT`와 `SELECT ... FOR UPDATE`는 같은 읽기가 아니라는 점
- 격리 수준을 올리는 것과 애플리케이션 불변식이 자동으로 안전해지는 것은 다르다는 점

15,000자 이상을 채우는 논리는 anomaly별로 `현상 이름 -> 실제 schedule -> 값 변화표 -> 엔진별 결과 -> 해결책 -> 비용`을 반복하는 것입니다.
표 하나로 끝내지 않고, 각 anomaly가 어떤 업무 사고로 보이는지 붙입니다.
예를 들어 lost update는 포인트 차감, 재고 감소, 잔액 갱신에서 어떻게 나타나는지 다룹니다.

검증 anchor는 PostgreSQL/MySQL 각각의 2-session script입니다.
PASS는 같은 schedule이 엔진과 격리 수준에 따라 다르게 보이는 것을 기계 출력으로 확인하는 것입니다.
FAIL은 `REPEATABLE READ에서는 phantom이 발생한다` 같은 면접용 압축문을 엔진 구분 없이 본문 결론으로 쓰는 것입니다.

### T05. 락, 2PL, deadlock

읽은 뒤 학습자는 `락은 동시성을 막는 장치가 아니라, 어떤 충돌을 기다리게 할지 명시하는 장치`라고 설명할 수 있어야 합니다.
learner gap은 `트랜잭션을 걸면 자동으로 안전하다`, `SELECT FOR UPDATE를 붙이면 다 해결된다`, `strict 2PL은 deadlock을 방지한다` 같은 압축 오해입니다.

첫 벽돌은 S/X lock 호환표와 두 계좌 update deadlock입니다.

```text
T1: lock A -> waits for B
T2: lock B -> waits for A

wait-for graph:
T1 -> T2
T2 -> T1
```

본문은 락의 종류를 많이 나열하기보다, `누가 무엇을 기다리는가`를 먼저 보여 줍니다.
그 뒤 row lock, table lock, gap lock, next-key lock, predicate lock, advisory lock을 실제 쓰임과 관측 지점으로 확장합니다.

반드시 닫을 지식은 다음입니다.

- shared/exclusive lock과 compatibility
- row lock은 읽기 자체를 막기보다 같은 row의 writer/locker를 막는 경우가 많다는 점
- `SELECT FOR UPDATE`, `FOR SHARE`, `SKIP LOCKED`의 쓰임
- MySQL InnoDB gap/next-key lock은 범위 조건과 phantom 방지에 연결된다는 점
- PostgreSQL predicate/SSI 설명은 공식 문서 범위에서 격리 장과 연결
- 2PL과 strict 2PL의 목적 차이
- 2PC와 2PL은 이름만 비슷한 다른 문제라는 점
- deadlock detection, timeout, lock ordering, retry

15,000자 이상을 채우는 논리는 `대기 그래프`를 중심으로 합니다.
처음에는 한 row update, 다음에는 두 row transfer deadlock, 다음에는 범위 조회와 phantom 방지, 다음에는 queue worker의 `SKIP LOCKED`, 마지막에는 운영에서 lock wait를 찾는 쿼리까지 이어갑니다.

검증 anchor는 `pg_locks`, `SHOW ENGINE INNODB STATUS`, `information_schema` 계열 lock wait 관측입니다.
PASS는 lock wait와 deadlock이 실제로 관측되고, retry/ordering으로 해결되는 것을 보여 주는 것입니다.

### T06. consistency, 제약조건, 도메인 불변식

읽은 뒤 학습자는 `일관성은 DB 제약조건만의 일이 아니라, 트랜잭션 전후에 도메인 불변식이 깨지지 않도록 DB와 애플리케이션이 나누어 맡는 일`이라고 설명할 수 있어야 합니다.
learner gap은 ACID의 C를 `DB가 알아서 정상 상태를 만든다`로 이해하는 것입니다.

첫 벽돌은 잔액 총합 보존입니다.

```text
before: A=1000, B=500, total=1500
transfer 300
after:  A=700,  B=800, total=1500
```

이 trace는 단순하지만, 바로 다음에 `수수료`, `보류 금액`, `취소`, `부분 실패`, `중복 실행`을 추가하면 총합 불변식이 어떤 모습으로 바뀌는지 드러납니다.

반드시 닫을 지식은 다음입니다.

- primary key, unique, foreign key, check constraint가 막을 수 있는 실패
- constraint로 막기 어려운 cross-row/cross-table/cross-system invariant
- optimistic lock/version column의 역할
- compare-and-swap와 conditional update
- idempotency table의 unique constraint는 row 중복이 아니라 실행권 중복을 막기 위해 쓰인다는 점
- 도메인 정책은 트랜잭션 안에서 계산하고, DB는 가능한 제약을 구조화해 주는 협력 관계라는 점

15,000자 이상을 채우는 논리는 같은 불변식을 여러 구현 수단으로 보호해 보는 방식입니다.
`CHECK(balance >= 0)`로 막히는 것, `UNIQUE(idempotency_key)`로 막히는 것, `UPDATE ... WHERE version=?`로 막히는 것, `SERIALIZABLE + retry`가 필요한 것을 나눕니다.

검증 anchor는 intentionally failing SQL입니다.
PASS는 제약조건 위반, version mismatch, unique conflict가 각각 다른 실패 신호로 드러나는 것입니다.

### T07. 따닥, 중복 요청, 멱등성

읽은 뒤 학습자는 `따닥 이슈는 버튼을 두 번 누르는 UI 문제가 아니라, 같은 의미의 요청이 여러 번 도착하고 외부 부작용이 중복될 수 있는 전체 경로 문제`라고 설명할 수 있어야 합니다.
learner gap은 debounce, HTTP method idempotency, idempotency key, DB transaction, 외부 PG/은행 실행을 한 단어로 뭉개는 것입니다.

첫 벽돌은 결제 버튼을 두 번 누른 장면입니다.

```text
click#1 -> POST /payments key=K1 -> server starts processing
click#2 -> POST /payments key=K1 -> server receives duplicate
network timeout hides response of click#1
client retries key=K1
```

본문은 이 흐름을 UI, HTTP, application, DB, external system으로 나눕니다.
UI disable은 사용자 경험을 줄일 뿐이고, 서버 멱등성은 반드시 별도로 필요합니다.
HTTP의 `PUT`/`DELETE` idempotent 의미와 `POST + Idempotency-Key`는 다른 층위입니다.

반드시 닫을 지식은 다음입니다.

- 멱등성(idempotency)의 HTTP 의미와 업무 실행 의미의 차이
- idempotency key, request hash, response replay
- `PROCESSING`, `SUCCESS`, `FAILED`, `UNKNOWN` 상태 머신
- DB unique constraint 기반 winner election
- `find -> 없으면 execute`가 깨지는 interleaving
- 외부 부작용은 winner 선출 이후에만 실행해야 한다는 점
- TTL, request hash mismatch, concurrent in-flight conflict, replay response 정책

15,000자 이상을 채우는 논리는 실패 interleaving을 많이 보여 주는 것입니다.
첫 번째 실패는 UI disable만 있는 경우, 두 번째 실패는 key는 있지만 request hash가 없는 경우, 세 번째 실패는 DB 저장이 외부 호출 뒤에 있는 경우, 네 번째 실패는 성공 응답이 유실된 경우입니다.
각 실패마다 어떤 관측 로그가 남고, 어떤 schema/status가 필요해지는지 설명합니다.

검증 anchor는 기존 `knowledge/cards/K-IDEMPOTENT-EXECUTION-WINNER-ELECTION-REPLAY.md`의 실험을 확장한 Java/SQL race lab입니다.
PASS는 action execution count가 1로 고정되는 것입니다.

### T08. 금액 계산과 반올림 정책

읽은 뒤 학습자는 `금액 계산은 타입 하나를 고르는 문제가 아니라, 표현 단위, 소수 자릿수, 반올림 시점, 잔여금 배분, 합계 대조 정책을 함께 정하는 문제`라고 설명할 수 있어야 합니다.
learner gap은 `BigDecimal 쓰면 안전하다` 또는 `정수 minor unit이면 항상 충분하다`는 단순화입니다.

첫 벽돌은 세 가지 표현의 비교입니다.

```text
double:       0.1 + 0.2 = 0.30000000000000004
BigDecimal:  "0.10" + "0.20" = 0.30
minor unit:  10 cents + 20 cents = 30 cents
```

본문은 표현 방식에서 끝나지 않고, `수수료 2.9% + 30원`, `세금`, `N명 분할`, `환불`, `부분 취소`, `환율`, `합계 대조`로 확장해야 합니다.

반드시 닫을 지식은 다음입니다.

- 이진 부동소수점이 십진 금액에 부적합한 이유
- Java `BigDecimal`의 scale, precision, rounding mode
- `new BigDecimal(double)`과 `BigDecimal(String)`의 차이
- ISO 4217 minor unit과 통화별 scale
- 정수 minor unit의 장점과 overflow/다통화 한계
- round half up, half even, floor/ceiling 등 정책의 도메인 소유자
- allocation/remainder: 100원을 3명에게 나눌 때 1원이 어디로 가는가
- DB `DECIMAL`과 애플리케이션 decimal의 scale 일치
- 표시 금액과 저장 금액, 계산 중간값을 분리하는 이유

15,000자 이상을 채우는 논리는 `값이 조금씩 어긋나는 trace`입니다.
작은 `0.1 + 0.2`에서 시작하고, 반올림 시점이 다른 두 계산이 최종 합계를 다르게 만드는 표를 보여 줍니다.
그 다음 수수료/세금/분할/환불을 추가해 실제 금융 장애로 이어지는 경로를 설명합니다.

검증 anchor는 BigDecimal 단위 테스트와 DB DECIMAL round-trip입니다.
PASS는 같은 입력에서 double 실패, BigDecimal/정수 방식 성공, 반올림 정책 차이를 명확히 관측하는 것입니다.

### T09. Spring 트랜잭션 경계

읽은 뒤 학습자는 `@Transactional은 메서드에 붙은 주석이 아니라, 프록시가 메서드 호출을 감싸고 현재 스레드 또는 reactive context에 리소스를 묶는 경계`라고 설명할 수 있어야 합니다.
learner gap은 `@Transactional을 붙이면 내부 호출과 비동기 작업까지 모두 자동으로 같은 트랜잭션이다`라는 오해입니다.

첫 벽돌은 호출 경로입니다.

```text
Controller
  -> Service proxy
  -> TransactionInterceptor
  -> PlatformTransactionManager.getTransaction()
  -> Connection bound to current thread
  -> Repository / MyBatis / JPA
  -> commit or rollback
```

본문은 Spring 공식 문서의 declarative transaction 구현 설명을 기반으로 해야 합니다.
특히 imperative transaction은 thread-bound resource를 쓰고, reactive transaction은 Reactor context를 쓴다는 차이를 분리합니다.

반드시 닫을 지식은 다음입니다.

- proxy 기반 트랜잭션과 self-invocation 문제
- `REQUIRED`, `REQUIRES_NEW`, `NESTED`의 물리/논리 트랜잭션 차이
- rollback rule: unchecked exception 기본 rollback, checked exception 정책
- `UnexpectedRollbackException`
- `TransactionSynchronizationManager`와 connection/session binding
- MyBatis/JPA flush와 commit의 차이
- async thread, event listener, transactional outbox 경계
- reactive transaction의 context 기반 경계

15,000자 이상을 채우는 논리는 `한 요청이 Spring service로 들어와 DB commit까지 가는 trace`입니다.
그 뒤 같은 trace에 내부 메서드 호출, 다른 빈 호출, 예외, `REQUIRES_NEW`, async, reactive pipeline을 하나씩 추가합니다.

검증 anchor는 작은 Spring test 또는 기존 `jvm/spring/spring_transactional.md`의 흐름을 기반으로 한 로그 관측입니다.
PASS는 트랜잭션 시작/참여/커밋 시점이 로그로 구분되는 것입니다.

### T10. 은행 이체와 결제 상태 머신

읽은 뒤 학습자는 `은행/결제 트랜잭션은 내부 DB transaction 하나가 아니라, 원장 상태 전이와 외부 시스템 결과를 대사 가능한 형태로 보존하는 장기 흐름`이라고 설명할 수 있어야 합니다.
learner gap은 `A 계좌 차감 + B 계좌 증가` 예제를 은행 시스템 전체 모델로 오해하는 것입니다.

첫 벽돌은 원장 이벤트입니다.

```text
ledger_entries
  1. debit  A 1000  transfer_id=T1
  2. credit B 1000  transfer_id=T1

invariant: transfer_id=T1의 debit 합계와 credit 합계가 같다.
```

이 장은 mutable balance를 바로 중심에 두지 않습니다.
먼저 원장 append와 합계 보존을 고정하고, 그 위에 available balance, hold, authorization, settlement, cancellation, refund, chargeback을 쌓습니다.

반드시 닫을 지식은 다음입니다.

- single-row balance update와 double-entry ledger의 차이
- ledger balance와 available balance
- authorization, capture/settlement, cancel, refund, chargeback
- 상태 전이의 허용/금지 transition
- idempotency key와 business transaction id
- 내부 원장 commit과 외부 은행/PG 응답의 경계
- 감사 로그와 수정 거래/reversal

15,000자 이상을 채우는 논리는 단일 이체를 여러 상태로 확장하는 방식입니다.
처음에는 내부 계좌 간 이체, 다음에는 외부 은행 송금 의뢰, 다음에는 결제 승인/매입, 다음에는 취소/환불, 마지막에는 대사와 운영자 보정입니다.

검증 anchor는 fake ledger table과 상태 전이 테스트입니다.
PASS는 허용되지 않은 transition이 실패하고, ledger 합계 불변식이 깨지지 않는 것입니다.

### T11. 외부 응답 유실과 정산/대사

읽은 뒤 학습자는 `은행은 성공했지만 우리 서버가 timeout을 본 상황은 rollback으로 해결하는 문제가 아니라, 상태를 UNKNOWN으로 보존하고 재조회/결과 파일/대사로 닫아야 하는 문제`라고 설명할 수 있어야 합니다.
learner gap은 외부 시스템 호출을 내부 DB transaction처럼 rollback할 수 있다고 착각하는 것입니다.

첫 벽돌은 timeout timeline입니다.

```text
t0 우리 서버: 송금 요청 저장 PENDING
t1 우리 서버 -> 은행: 송금 전문 전송
t2 은행: 실제 처리 SUCCESS
t3 네트워크: 응답 유실
t4 우리 서버: timeout, 결과 UNKNOWN
t5 재조회/결과 파일/대사로 SUCCESS 확인
```

본문은 `성공인지 실패인지 모르는 상태`를 부정확한 예외가 아니라 정상 상태로 다뤄야 합니다.
이 상태를 잘못 `FAILED`로 닫으면 중복 송금이나 누락 보정으로 이어질 수 있습니다.

반드시 닫을 지식은 다음입니다.

- timeout은 실패가 아니라 미확정일 수 있다는 점
- UNKNOWN/PROCESSING/SUCCESS/FAILED 상태 분리
- 재조회 API, 결과 파일, settlement report, reconciliation
- 보상 거래와 원거래 취소의 차이
- 운영자 개입이 필요한 상태와 자동 재시도 가능한 상태
- audit trail, correlation id, external reference id
- 내부 DB transaction으로 외부 세계를 rollback할 수 없다는 점

15,000자 이상을 채우는 논리는 실패 시나리오 matrix입니다.
응답 전송 전 실패, 은행 처리 후 응답 유실, 우리 DB commit 전 crash, commit 후 worker crash, 결과 파일 불일치, 중복 callback을 각각 시간표로 보여 줍니다.

검증 anchor는 fake bank와 state machine test입니다.
PASS는 timeout 뒤 즉시 재송금하지 않고, idempotency key/external reference로 조회하거나 대사하는 흐름이 보이는 것입니다.

### T12. 분산 트랜잭션, saga, outbox

읽은 뒤 학습자는 `분산 트랜잭션 대안은 2PC 하나와 saga 하나 중 고르는 문제가 아니라, 어떤 경계를 원자적으로 묶고 어떤 경계는 재시도/보상/대사로 닫을지 정하는 문제`라고 설명할 수 있어야 합니다.
learner gap은 2PC, 2PL, saga, outbox, exactly-once를 같은 층위의 대체재로 보는 것입니다.

첫 벽돌은 local transaction + outbox 흐름입니다.

```text
BEGIN;
INSERT INTO payment_order(... status='REQUESTED');
INSERT INTO outbox(... event_type='SEND_TO_BANK', status='PENDING');
COMMIT;

worker: claim outbox -> send to bank -> mark sent/retry/unknown
```

본문은 먼저 2PC/XA/JTA가 어떤 문제를 해결하는지, 왜 blocking cost와 운영 복잡도가 있는지 설명합니다.
그 뒤 saga는 원자 커밋을 흉내 내는 것이 아니라 보상 가능한 long-running transaction을 설계하는 방식임을 설명합니다.
outbox는 외부 전달 의도를 로컬 transaction 안에 남기는 pattern이고, inbox/consumer idempotency와 함께 써야 합니다.

반드시 닫을 지식은 다음입니다.

- 2PC prepare/commit, coordinator, participant, blocking risk
- 2PC와 2PL 차이
- XA/JTA와 Spring transaction boundary
- saga choreography/orchestration과 compensation
- transactional outbox, inbox, claim lease
- broker exactly-once와 business exactly-once 차이
- CDC와 polling worker의 차이
- replay, dedup, reconciliation과의 연결

15,000자 이상을 채우는 논리는 같은 결제 요청을 세 가지 설계로 비교하는 것입니다.
첫째는 naive direct call, 둘째는 2PC/XA, 셋째는 local transaction + outbox + idempotent consumer + reconciliation입니다.
각 설계의 failure timeline과 운영 관측 지점을 보여 줍니다.

검증 anchor는 기존 `K-ASYNC-EVENTS-OUTBOX-INBOX-CLAIM-LEASE.md`와 작은 worker race lab입니다.
PASS는 여러 worker가 떠도 event가 한 번만 claim되고, 실패 시 retry 상태가 복원되는 것입니다.

## 6. 근거와 소스 전략

기존 원문 배치본만으로는 domain truth를 닫지 않습니다.
아래 소스 계층을 함께 사용합니다.

### 6.1 로컬 원재료와 연결 자산

- `interviews/database-storage-search-nosql.md`: 기존 DB 원문 배치본. ACID, MVCC, lock, 2PC/2PL의 출발점이지만 최종 근거는 아닙니다.
- `interviews/spring-backend-frameworks.md`: Spring 데이터 접근과 트랜잭션 질문 연결.
- `jvm/spring/spring_transactional.md`: Spring transaction propagation과 proxy 흐름 원재료.
- `knowledge/cards/K-IDEMPOTENT-EXECUTION-WINNER-ELECTION-REPLAY.md`: winner election과 replay 정책.
- `knowledge/cards/K-ASYNC-EVENTS-OUTBOX-INBOX-CLAIM-LEASE.md`: outbox/inbox와 claim lease.
- `domains/firmbanking/*`: 은행/펌뱅킹 batch, 결과 수신, 대사 시나리오의 로컬 도메인 원재료.
- `domains/payment/*`: PG, VAN, 결제 운영, 정산, 대사, 영수증, 카드/할부 흐름의 로컬 원재료. 실제 로그나 샘플 JSON을 인용할 때는 secret, token, 개인정보, 내부 식별자를 최소화하거나 익명화합니다.
- `math/bankers_rounding.md`, `math/precision.md`, `domains/scale_factor.md`: 금액 계산 원재료. 단, 현재 수준은 최종 심화 품질이 아니므로 공식 자료와 실험으로 보강합니다.

### 6.2 공식/1차 자료 후보

- PostgreSQL 공식 문서: transaction isolation, concurrency control, explicit locking, WAL, serialization failure handling.
- MySQL InnoDB 공식 문서: transaction model, isolation levels, consistent nonlocking reads, locking reads, next-key/gap locks, autocommit/commit/rollback.
- Spring Framework 공식 문서: declarative transaction, transaction implementation, rollback rules, propagation, transaction synchronization, reactive transaction.
- Java 공식 문서: `BigDecimal`, `RoundingMode`, JDBC transaction API.
- RFC 9110: HTTP method idempotency.
- IETF `Idempotency-Key` draft와 결제 API 공식 문서: POST retry와 idempotency key 실무 정책.
- ISO 4217: 통화 minor unit.

### 6.3 실험/검증 후보

- PostgreSQL 2-session SQL: read committed/repeatable read/serializable 비교, lock wait, deadlock, `SELECT FOR UPDATE`, `SKIP LOCKED`.
- MySQL 2-session SQL: InnoDB repeatable read, read committed, next-key lock, gap lock, deadlock.
- Java race lab: read-then-act 실패와 winner election 성공.
- BigDecimal unit test: `double` 실패, `BigDecimal(String)` 성공, rounding mode 차이, allocation remainder.
- Spring mini app 또는 test: proxy 진입, self-invocation, propagation, rollback rule, async thread 차이.
- fake bank state machine: timeout, duplicate callback, result file reconciliation.

## 7. Dialectic claim cards

### C1. 문서 구조 claim

- claim: 하나의 대형 파일이 아니라 허브 + 3개 상세 파일로 나누는 것이 더 안전하다.
- premises: 기존 DB 문서는 원문 배치본이고, 이번 주제는 DB 내부, 애플리케이션 경계, 금융 시나리오가 서로 다른 실패 모델을 가진다.
- attack: 파일을 나누면 연결성이 떨어질 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: 허브 문서를 별도로 두어 읽는 순서와 상호 링크를 제공하고, 상세 파일은 실패 모델별로 분리한다.
- support tier: T2 strong inference from repo structure and prior Linux/network monograph.
- admission lane: APPLY for planning, later writing에서 재검증.

### C2. 15,000자 단위 claim

- claim: 각 `##` 주요 개념/시나리오를 최소 15,000자 검증 단위로 둔다.
- premises: 이전 긴 문서 작업에서 `###` subsection length가 검증 단위였고, 이번에는 파일을 나누므로 `##` 주요 단위가 더 자연스럽다.
- attack: 하위 용어까지 모두 15,000자로 만들면 filler가 된다.
- response lane: ACCEPT_REPAIR.
- repaired claim: 12개 주요 단위만 15,000자 이상으로 검증하고, 하위 용어는 해당 단위 안의 필수 구성요소로 둔다.
- support tier: T2 strong inference from prior repo practice.
- admission lane: APPLY for planning.

### C3. MVCC/격리 수준 claim

- claim: MVCC와 격리 수준은 SQL 표준 표, PostgreSQL, MySQL InnoDB를 분리해 설명해야 한다.
- premises: 공식 문서 기준으로 PostgreSQL과 InnoDB의 구현과 허용/차단 현상이 다르다.
- attack: 면접 준비에서는 표준 표가 더 간단하다.
- response lane: REBUT with downgrade of simple table to quick reference only.
- repaired claim: 표준 anomaly 표는 초반 지도에만 쓰고, 실제 본문 결론은 engine-specific schedule과 공식 문서 근거로 닫는다.
- support tier: T1 required during writing, current plan support is T2 because exact source-to-claim mappings are not yet embedded in this artifact.
- admission lane: APPLY for planning, source-gated APPLY for final prose.

### C4. 따닥/멱등성 claim

- claim: 따닥 이슈는 UI debounce, HTTP idempotency, DB winner election, external side effect replay를 분리해야 한다.
- premises: 각 층위가 막는 실패가 다르고, 외부 호출 중복은 UI debounce나 단순 transaction만으로 막을 수 없다.
- attack: 사용자는 실무적으로 `따닥`이라고만 부르므로 너무 쪼개면 읽기 어렵다.
- response lane: ACCEPT_REPAIR.
- repaired claim: 문서 opening은 `따닥`이라는 실제 증상으로 시작하되, 본문은 층위별 실패와 방어로 분해한다.
- support tier: T2 strong inference backed by local idempotency card and RFC/Stripe docs.
- admission lane: APPLY.

### C5. 은행 시나리오 claim

- claim: 은행/결제 시나리오는 mutable balance update가 아니라 ledger/state machine/reconciliation을 중심으로 작성해야 한다.
- premises: 외부 금융 시스템은 내부 DB transaction으로 rollback할 수 없고, timeout/응답 유실/결과 파일/대사가 핵심 실패 경로다.
- attack: 너무 금융 도메인으로 들어가면 DB 트랜잭션 공부에서 벗어날 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: 은행 시나리오는 DB 트랜잭션의 한계를 보여 주는 적용 장으로 두고, 법/업무 세부는 공개 자료와 로컬 firmbanking 원재료 범위에서만 다룬다.
- support tier: T2 strong inference, T1 sources required during writing.
- admission lane: CANDIDATE/APPLY after source verification.

## 8. Multi-agent review record

### R1. 문제 정의와 보호 의도

- proposer claim: 긴 DB 트랜잭션 문서 세트는 먼저 유한한 개념/시나리오 레지스트리를 고정해야 한다.
- critic challenge: `등등`을 열린 범위로 두면 전수 완료를 판정할 수 없다.
- response lane: ACCEPT_REPAIR.
- synthesis: 12개 주요 작성 단위를 frozen registry로 둔다. 하위 개념은 각 단위의 구성요소로 흡수한다.
- completion marker: registry table 작성.

### R2. 근거와 가정

- proposer claim: 기존 원문 배치본은 출발점으로 충분하다.
- critic challenge: 기존 문서에는 PostgreSQL phantom, strict 2PL, ACID C 같은 위험한 압축/오류 후보가 있다.
- response lane: DOWNGRADE.
- synthesis: 기존 원문은 source reservoir로만 사용하고, DB truth는 공식 PostgreSQL/MySQL/Spring/Java/HTTP 자료와 실험으로 닫는다.
- completion marker: source strategy section 작성.

### R3. 체크리스트 누락과 왜곡

- proposer claim: ACID, MVCC, isolation, idempotency, money, banking만 있으면 충분하다.
- critic challenge: WAL/undo/redo, lock/deadlock, Spring boundary, outbox/saga가 빠지면 실제 시나리오를 설명할 수 없다.
- response lane: ACCEPT_REPAIR.
- synthesis: 12개 단위에 WAL, lock, consistency, Spring, outbox를 포함한다.
- completion marker: T01-T12 registry.

### R4. 계획 검증

- proposer claim: `15,000자 이상`은 각 주요 단위에 적용한다.
- protocol challenge: 어떤 heading 단위인지, 어떻게 측정할지 없으면 검증 불가능하다.
- response lane: ACCEPT_REPAIR.
- synthesis: 각 상세 파일의 `##` 주요 단위 본문을 heading 제외 15,000자 이상으로 측정하고, section count TSV를 남긴다.
- completion marker: validation section 작성.

### R5. closure와 다음 단계

- proposer claim: 이 계획이 있으면 바로 전체 완료다.
- sentinel challenge: 계획 tranche는 본문 작성 전 단계이며 whole request complete가 아니다.
- response lane: ACCEPT_REPAIR.
- synthesis: 이 문서는 planning tranche 결과다. 실제 WHOLE_COMPLETE는 T01-T12 작성, 기계 검증, 최종 감사, path-limited commit까지 끝난 뒤에만 선언한다.
- completion marker: closure section 작성.

## 9. Review-kernel result

- review target: DB 트랜잭션/중복 실행/금액 계산/은행 시나리오 심화 문서 작성 전 계획.
- protected purpose: 긴 글의 양이 아니라, 개념 경계, 근거, 반례, 검증 경로가 닫힌 학습 자산을 만들기 위한 작성 계약을 세우는 것.
- activated gates: purpose fit, claim fidelity, evidence boundary, counterexample resistance, omission/distortion, verification, downstream impact.
- domain packs/evidence: study-explanation, repo-local interviews rules, DB 공식 문서 후보, Spring/Java/HTTP 공식 문서 후보, 로컬 idempotency/outbox/payment/firmbanking/math 원재료.
- evidence boundary: 이 계획은 source strategy와 registry를 닫은 planning artifact다. DB 엔진별 사실과 금융 도메인 세부 결론은 다음 writing tranche에서 공식 자료와 실험으로 다시 닫아야 한다.
- role topology: 실제 delegated sub-agents로 Domain Planner, Critic, Protocol Sentinel을 실행했고, Orchestrator가 findings를 repair/synthesis했다.
- verdict: `ALIGN for planning tranche only`.
- repair disposition: critic의 scope split, idempotency/banking/MVCC/money warnings를 ACCEPT_REPAIR 또는 DOWNGRADE로 반영했다. 남은 risk는 writing tranche source verification으로 넘긴다.
- residual risk: 계획은 충분하지만 실제 본문 품질은 아직 검증되지 않았다. T01-T12 작성 전 이 계획과 WORK를 다시 읽고 source-to-claim mapping을 새로 채워야 한다.

## 10. 작성 체크리스트

실제 본문 작성 단계는 아래 체크리스트를 freeze한 뒤 시작합니다.

- [ ] 허브 파일과 3개 상세 파일 생성
- [ ] T01-T12 모든 `##` 단위 작성
- [ ] 각 T 단위 body 15,000자 이상
- [ ] 각 T 단위에 first brick 또는 concrete artifact 존재
- [ ] 각 T 단위에 값/상태 변화 trace 존재
- [ ] 각 T 단위에 실패 모드와 반례 존재
- [ ] 각 T 단위에 검증 또는 replay 경로 존재
- [ ] MVCC/isolation은 PostgreSQL과 MySQL InnoDB 차이를 분리
- [ ] 2PC와 2PL은 별도 문제로 대조
- [ ] `따닥`은 UI debounce, HTTP retry, DB winner election, 외부 부작용 replay를 분리
- [ ] 금액 계산은 scale, rounding, allocation, DB/app representation, currency minor unit 포함
- [ ] 은행 시나리오는 ledger, available balance, settlement, unknown state, reconciliation 포함
- [ ] Spring 장은 proxy, propagation, rollback rule, resource binding, async/reactive boundary 포함
- [ ] 각 파일의 local link sanity 확인
- [ ] fenced code block balance 확인
- [ ] 중복 장문 문단 탐지
- [ ] 금지 scaffolding 탐지: assistant 진행 멘트, 짧은 면접 답변 양식, 질문 의도 라벨 잔재
- [ ] `git diff --check`
- [ ] dirty tree에서 touched paths만 path-limited staging
- [ ] `git diff --cached --name-status` 확인 후 commit

## 11. 검증 방식

본문 작성 완료 뒤에는 추정이 아니라 스크립트로 검증합니다.
기본 검증은 다음입니다.

```bash
python3 - <<'PY'
from pathlib import Path

targets = [
    Path("interviews/db-transactions-core-concepts.md"),
    Path("interviews/db-transactions-application-boundaries.md"),
    Path("interviews/db-transactions-financial-scenarios.md"),
]

for path in targets:
    text = path.read_text(encoding="utf-8")
    lines = text.splitlines()
    starts = []
    for i, line in enumerate(lines):
        if line.startswith("## ") and not line.startswith("### "):
            starts.append((i, line[3:].strip()))
    for idx, (start, title) in enumerate(starts):
        end = starts[idx + 1][0] if idx + 1 < len(starts) else len(lines)
        body = "\n".join(lines[start + 1:end]).strip()
        print(f"{path}\t{title}\t{len(body)}")
PY
```

PASS 기준은 T01-T12 각 body length가 15,000 이상인 것입니다.
FAIL 기준은 하나라도 15,000 미만이거나, heading 단위가 누락되거나, 중복 장문 문단/깨진 link/fence 불균형이 발견되는 것입니다.

중복 장문 문단 검사는 fenced code block을 제외하고 수행합니다.
문자 수만 맞추기 위해 같은 설명을 반복하는 실패를 막기 위해서입니다.

## 12. downstream impact gate

이 계획은 후속 작성자를 행동하게 만드는 action surface입니다.
예상 downstream actor는 이 저장소에서 다음 본문 작성 tranche를 수행할 AI 또는 사용자입니다.
예상 행동은 4개 파일 생성, 12개 장문 단위 작성, 검증, commit입니다.

되돌림 가능성은 높습니다.
새 파일 중심으로 작성하고 기존 원문 배치본은 링크 동기화만 하므로, 잘못된 방향이면 새 파일을 수정하거나 제거할 수 있습니다.
다만 잘못된 DB/금융 사실을 장문으로 쓰면 학습 자산에 장기 오류가 남으므로, DB 엔진 차이와 금융 시나리오는 공식 자료와 실험으로 닫아야 합니다.

safer path는 다음입니다.

1. 이 계획 파일을 먼저 review한다.
2. T01-T06 core concepts를 먼저 작성하고 검증한다.
3. T07-T09 application boundaries를 작성하고 검증한다.
4. T10-T12 financial scenarios를 작성하고 검증한다.
5. 허브 문서를 마지막에 연결한다.

## 13. closure control

- requested whole objective: DB 트랜잭션, 따닥 이슈, 금액 계산, 은행 트랜잭션 시나리오 등을 15,000자 이상 심화 문서로 정리한다.
- current tranche: 작성 전 multi-agent/dialectic/review 계획 수립.
- achieved closure scope: finite concept/scenario registry, document topology, teaching spine, evidence strategy, validation strategy, critic/sentinel repairs.
- whole-request completion verdict: REQUEST_PARTIAL.
- remaining executable count: 12 primary writing units plus hub creation.
- next immediate target: `db-transactions-core-concepts.md`의 T01-T06 작성.
- current-vs-historical evidence: 이전 Linux/network 작업의 검증 방식은 memory-derived precedent이고, 이번 DB/금융 domain truth는 작성 시 공식 자료와 실험으로 새로 닫아야 한다.
