# MVCC snapshot visibility

## MVCC 역사와 등장 배경

MVCC는 락을 없앤 기술이 아니라 여러 버전과 가시성 규칙으로 읽기 일관성과 동시성을 함께 얻으려는 동시성 제어 방식이다. 기존 `database/mvcc.md`는 MVCC의 목적, RDBMS 역사, Reed의 논문, MGA와 rollback segment라는 두 흐름을 넓게 모아 두었다. 이 절은 그 seed를 보존하되 공식 PostgreSQL과 MySQL 문서에 맞춰 현재 엔진 관점으로 다시 정리한다.

배경은 공유 데이터의 압력이다. 단일 사용자 장부에서는 순서대로 읽고 쓰면 된다. 다중 사용자 OLTP에서는 긴 읽기가 쓰기를 막고, 긴 쓰기가 읽기를 막고, 아무 제어도 없으면 dirty read와 lost update가 생긴다. MVCC는 현재 값을 단순히 덮어쓰는 대신 버전과 snapshot/read view를 두어 `내가 볼 수 있는 과거 세계`를 판정한다.

이 절에서 하중을 지탱하는 공식 자료는 다음과 같다. 링크 자체가 결론은 아니며, 본문에서는 각 공식 문서가 제공하는 규칙을 계좌 이체, row version, read view, vacuum/purge trace와 연결해 다시 검증 가능한 형태로 풀어 쓴다.

- PostgreSQL MVCC: https://www.postgresql.org/docs/current/mvcc.html
- MySQL InnoDB multi-versioning: https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html
- Local seed: database/mvcc.md

```text
PostgreSQL 계열에 가까운 heap version model
page before: [tuple A xmin=10 xmax=-- value=100]
UPDATE by tx=30
page after : [tuple A xmin=10 xmax=30 value=100]
             [tuple B xmin=30 xmax=-- value=120]
old snapshot -> tuple A
new snapshot -> tuple B

InnoDB 계열에 가까운 undo chain model
record before: [id=1 value=100 trx_id=10 roll_ptr=null]
UPDATE by tx=30
record after : [id=1 value=120 trx_id=30 roll_ptr=undo#7]
undo#7       : [old value=100 old_trx_id=10 prev=null]
old read view -> undo#7을 따라 value=100 재구성
```

| 질문 | PostgreSQL 쪽 모델 | InnoDB 쪽 모델 | 공통 판단 |
|---|---|---|---|
| 과거 버전 위치 | heap tuple | undo log chain | snapshot/read view가 볼 버전 선택 |
| 정리 작업 | vacuum | purge | 오래된 reader가 정리를 막음 |
| writer 충돌 | row lock/update conflict | record/next-key lock/write conflict | MVCC여도 writer는 충돌 |
| 관측 시작점 | pg_stat_activity, n_dead_tup | innodb_trx, history list | long transaction 추적 |

### 1. 역사적 압력은 읽기-쓰기 충돌이었다

2PL 같은 lock 기반 방식은 직렬 가능성을 만들 수 있지만 읽기와 쓰기의 대기 비용이 크다. 읽기 비중이 높은 시스템에서 보고서나 사용자 조회가 쓰기를 막으면 처리량이 떨어지고, 쓰기가 조회를 막으면 응답 지연이 늘어난다.

MVCC는 읽기에게 과거 version을 제공해 이 충돌을 줄인다. 다만 writer끼리 같은 row를 수정하는 충돌은 여전히 남기 때문에 lock-free라고 부르면 안 된다.

### 2. 버전은 복사본이 아니라 가시성 판정 대상이다

여러 버전을 저장한다는 말만으로는 부족하다. 각 version에는 만든 transaction, 끝낸 transaction, commit 여부, active transaction 집합과 비교할 정보가 필요하다.

그래서 MVCC의 핵심은 `몇 개 version이 있나`가 아니라 `이 transaction이 어느 version을 볼 수 있나`다. 이 판정이 snapshot isolation, repeatable read, consistent read의 실제 의미를 만든다.

### 3. MGA와 rollback/undo 계열은 저장 위치가 다르다

기존 메모의 MGA/rollback segment 구분은 중요하다. PostgreSQL은 old/new tuple이 heap에 남는 모델로 이해하기 좋고, InnoDB는 최신 clustered record와 undo chain으로 과거를 재구성하는 모델로 이해해야 한다.

이 차이를 놓치면 PostgreSQL의 vacuum을 MySQL에 그대로 대입하거나 InnoDB undo를 PostgreSQL heap tuple과 같은 저장 방식으로 설명하는 오류가 생긴다.

### 4. snapshot 시점은 격리 수준마다 다르다

MVCC라고 해서 항상 트랜잭션 시작 시점 하나만 보는 것은 아니다. PostgreSQL READ COMMITTED는 문장마다 snapshot을 잡고, REPEATABLE READ는 트랜잭션 snapshot 성격이 강하다. InnoDB도 isolation level과 consistent read 생성 시점에 따라 read view 재사용이 달라진다.

따라서 정확한 설명은 `어느 DBMS의 어느 격리 수준에서 어느 읽기 종류인가`를 포함해야 한다.

### 5. consistent read와 locking/current read를 구분한다

일반 SELECT가 과거 snapshot을 읽을 수 있다고 해서 `SELECT FOR UPDATE`도 같은 경로라고 보면 안 된다. 수정 전 선점이 필요한 읽기는 lock을 잡고 최신/현재 버전을 다르게 다룬다.

이 구분은 특히 InnoDB에서 중요하다. plain SELECT와 locking read를 섞어 쓰는 코드는 같은 트랜잭션 안에서도 관측과 대기 양상이 달라질 수 있다.

### 6. 오래된 version은 언젠가 비용이 된다

버전을 남기는 방식은 읽기 일관성을 주지만 storage와 cleanup 비용을 만든다. PostgreSQL은 dead tuple과 vacuum, InnoDB는 undo history와 purge로 비용이 드러난다.

long transaction은 과거 snapshot/read view를 붙잡아 cleanup horizon을 뒤로 밀 수 있다. 읽기만 하는 작업도 정리 지연의 원인이 될 수 있다.

### 7. MVCC는 rollback과도 연결된다

변경 전 상태를 남기는 구조는 rollback에도 필요하다. InnoDB undo log는 rollback과 consistent read에 모두 쓰이고, Oracle undo도 읽기 일관성과 rollback을 함께 지원한다.

따라서 undo를 단순 취소 로그라고만 보면 MVCC를 절반만 이해한 것이다. 반대로 MVCC version을 recovery log와 같은 것으로 뭉개도 안 된다.

### 8. 동시성 이득은 공짜가 아니다

MVCC는 reader/writer blocking을 줄일 수 있지만, 더 많은 metadata, version cleanup, conflict detection, transaction id 관리가 필요하다. 성능이 좋아 보이는 이유와 새로 생긴 비용을 함께 설명해야 한다.

운영에서 이 비용은 vacuum lag, purge lag, undo tablespace 증가, bloat, serialization failure, lock wait로 관측된다.

### 9. 분산과 replica에서는 snapshot 의미가 더 조심스럽다

read replica, logical replication, distributed transaction에서는 어떤 시점의 commit을 어느 노드가 보았는지가 추가된다. MVCC snapshot이 local engine 안에서 닫혀도 replica lag나 external consistency까지 자동 보장하지 않는다.

이 절의 범위는 단일 engine 내부지만, 실무에서는 read-your-writes와 replica routing 문제로 바로 이어진다.

### 10. 역사는 암기표가 아니라 설계 이유를 보여 주는 도구다

Reed, Bernstein/Goodman, InterBase/Firebird, Oracle, PostgreSQL, InnoDB의 이름을 외우는 것보다 중요한 것은 공유 데이터에서 읽기 일관성과 동시성 요구가 어떻게 버전 관리 구조를 낳았는지 이해하는 것이다.

기존 메모의 역사 흐름은 보존하되, 제품별 현재 동작은 vendor 공식 문서와 실험으로 다시 닫아야 한다.

```sql
-- two-session lab의 관측 출발점
-- PostgreSQL
SELECT pid, state, xact_start, backend_xmin, query
FROM pg_stat_activity
WHERE xact_start IS NOT NULL OR backend_xmin IS NOT NULL;

-- MySQL/InnoDB
SELECT trx_id, trx_started, trx_state, trx_query
FROM information_schema.innodb_trx
ORDER BY trx_started;
SHOW ENGINE INNODB STATUS;
```

검증은 두 세션으로 시작한다. 세션 A가 snapshot/read view를 만든 뒤 세션 B가 같은 row를 update하고 commit한다. 세션 A가 같은 값을 계속 보는지, 격리 수준을 바꾸면 언제 새 값을 보는지, locking read를 쓰면 대기가 생기는지 확인한다. PASS는 결과를 맞히는 것이 아니라 `어떤 version 저장 구조와 어떤 visibility 규칙 때문에 그 결과가 나왔는지`를 설명하는 것이다.

### 반복을 줄인 최종 mental model

MVCC를 머릿속에 넣을 때는 세 단어를 순서대로 붙이면 된다. 첫째는 version이다. 한 row가 바뀌었다고 해서 모든 reader가 즉시 새 row만 보아야 하는 것은 아니다. 이미 시작된 reader에게는 과거 row image가 필요할 수 있다. PostgreSQL은 heap tuple 자체가 여러 version으로 남는 그림에 가깝고, InnoDB는 clustered record의 현재 모습에서 undo record를 따라 과거 image를 복원하는 그림에 가깝다. 이 차이는 저장 위치의 차이이지, 과거 시점을 보여 준다는 목표 자체의 차이만은 아니다.

둘째는 snapshot이다. version이 여러 개 있어도 어느 version을 보여 줄지 정하는 기준이 없으면 DB는 일관된 읽기를 제공할 수 없다. snapshot은 내가 보는 세계의 경계다. 이 경계 안에서 이미 끝난 transaction의 변화는 보이고, 아직 진행 중이거나 내가 보는 시점보다 나중에 끝난 변화는 보이지 않는다. 그래서 snapshot은 최신성의 약속이 아니라 일관성의 약속이다. 고객이 방금 바꾼 값이 화면에 안 보인다고 말하면, 개발자는 먼저 같은 transaction인지, 같은 connection인지, replica를 읽는지, 격리 수준이 무엇인지 확인해야 한다.

셋째는 cleanup이다. version을 남기는 순간 DB는 언젠가 그것을 치워야 한다. 과거 image가 필요한 reader가 남아 있으면 cleanup은 기다린다. 이 때문에 오래 열린 transaction은 자기 요청만 느리게 만드는 것이 아니라, 다른 update가 만든 죽은 version을 오래 붙잡아 저장 공간과 index scan 비용을 키운다. 실무에서 MVCC 장애는 대개 이 세 단어 중 하나를 빼먹을 때 생긴다. version만 보면 읽기와 쓰기가 안 막힌다는 반쪽 설명이 되고, snapshot만 보면 최신이어야 한다는 오해가 생기며, cleanup을 빼면 vacuum이나 purge가 왜 운영 장애의 중심이 되는지 이해하지 못한다.

작은 비유로는 편집 이력 있는 문서를 떠올릴 수 있다. 여러 사람이 같은 문서를 고치지만, 어떤 독자는 자기 검토를 시작한 시점의 문서를 계속 보고 싶다. 그러려면 문서 시스템은 이전 문단을 일정 기간 보관하고, 독자별로 어떤 버전을 볼지 정하며, 더 이상 아무도 보지 않는 이력을 지워야 한다. DB의 MVCC도 이와 비슷하다. 다만 DB에서는 그 이력이 단순 편의 기능이 아니라 commit, rollback, lock, vacuum, purge, replication, backup과 연결된 핵심 저장 계약이다.

```text
MVCC 설명 자가 점검
  1. UPDATE 하나가 old version과 new version을 어떻게 남기는가?
  2. reader는 어떤 snapshot/read view로 둘 중 하나를 고르는가?
  3. writer끼리 충돌할 때 MVCC가 lock을 완전히 없애는가?
  4. 오래 열린 transaction이 cleanup을 왜 막는가?
  5. PostgreSQL과 InnoDB는 과거 image를 어디에 보관하는가?
```

이 다섯 질문에 답할 수 있으면 기존 `database/mvcc.md`의 핵심 질문은 보존된 것이다. 원문에서 좋았던 부분은 버전을 남기면 읽기와 쓰기를 덜 막을 수 있다는 문제의식이다. 새 문서에서는 그 문제의식을 공식 용어와 실험 가능한 trace로 다시 묶었다. 이제 MVCC는 암기할 정의가 아니라, 동시성 이득을 얻는 대신 version 보관과 cleanup 비용을 떠안는 설계로 기억하면 된다.

MVCC를 실제 코드 설계로 옮길 때는 “읽기 일관성”과 “업무 선점”을 분리해야 한다. 보고서 화면은 같은 시점의 세계를 안정적으로 읽는 것이 중요하므로 snapshot이 강한 도움을 준다. 반대로 재고 차감, 좌석 예약, 쿠폰 사용처럼 앞으로의 변경 권리를 잡아야 하는 작업은 snapshot만으로 충분하지 않다. 이런 작업은 조건부 UPDATE, unique constraint, `SELECT ... FOR UPDATE`, idempotency key, retry 정책을 함께 설계해야 한다. MVCC는 과거를 일관되게 보여 주지만, 미래의 경합을 자동으로 예약해 주지는 않는다.

장애 대응에서도 이 분리가 중요하다. 사용자가 “조회가 오래된 값을 보여 준다”고 말하면 먼저 snapshot, replica lag, cache를 구분한다. 운영자가 “update가 막힌다”고 말하면 writer lock, gap/next-key lock, deadlock, long transaction을 구분한다. DBA가 “테이블이 커진다”고 말하면 version cleanup, vacuum/purge, undo history, autovacuum/purge worker를 본다. 모두 MVCC 주변에서 생기는 증상이지만 같은 조치로 해결되지 않는다.

```text
증상별 첫 분기
  오래된 값 조회:
    transaction snapshot인가, read replica인가, application cache인가?
  writer 대기:
    같은 row writer 충돌인가, range lock인가, deadlock인가?
  저장 공간 증가:
    dead tuple/undo history인가, index bloat인가, long transaction인가?
  재시도 오류:
    serialization failure인가, deadlock victim인가, lock timeout인가?
```

이 분기표가 있으면 MVCC 설명은 면접 답변을 넘어 운영 runbook이 된다. 좋은 문서는 이처럼 개념을 장애 분류와 연결해야 한다. 그래야 학습자가 나중에 PostgreSQL이나 InnoDB에서 다른 용어를 만나도, 먼저 “이 증상은 version, snapshot, cleanup, writer coordination 중 어디에서 왔는가”라고 물을 수 있다.

MVCC를 학습한 뒤 마지막으로 해야 할 연습은 같은 UPDATE를 세 번 그리는 것이다. 첫 번째 그림은 논리 row가 어떻게 바뀌는지 보여 준다. 두 번째 그림은 PostgreSQL heap tuple과 vacuum horizon으로 그린다. 세 번째 그림은 InnoDB clustered record와 undo chain으로 그린다. 세 그림이 같은 업무 사건을 설명하지만 서로 다른 물리 구조를 쓴다는 사실을 손으로 확인하면, 특정 엔진 문서를 읽을 때도 길을 잃지 않는다.

이 연습의 PASS 조건은 엔진 이름을 지우고도 문제 압력을 말할 수 있는 것이다. 긴 읽기가 쓰기를 막지 않게 하고 싶다. 하지만 과거 값을 보여 주려면 version이 필요하다. version을 남기면 cleanup이 필요하다. cleanup은 오래된 snapshot 앞에서 멈춘다. 이 네 문장이 자연스럽게 이어지면 MVCC의 역사와 등장 배경은 단순 암기가 아니라 설계 이유로 자리 잡는다.

### 추가 판정 질문: MVCC가 없으면 어떤 운영 결정을 해야 하는가

MVCC가 왜 필요한지 더 깊게 이해하려면 반대로 MVCC가 없는 시스템을 상상해 보면 된다. 긴 보고서 SELECT가 시작된 동안 모든 writer를 막으면 읽기는 일관되지만 서비스 쓰기 처리량이 크게 떨어진다. 반대로 writer를 자유롭게 두고 reader가 매번 최신 page를 읽게 하면 보고서는 서로 다른 시점의 값을 섞을 수 있다. MVCC는 이 둘 사이에서 과거 version을 비용으로 지불하고 읽기 일관성과 쓰기 동시성을 함께 얻으려는 선택이다.

이 선택을 이해하면 MVCC의 한계도 자연스럽다. 과거 version을 보여 주는 것은 현재 자원을 예약하는 것과 다르다. 재고 차감이나 좌석 예약처럼 미래의 삽입과 변경을 막아야 하는 문제에는 lock, constraint, retry가 여전히 필요하다. MVCC를 잘 안다는 것은 장점만 말하는 것이 아니라, 어느 순간부터 다른 동시성 도구를 꺼내야 하는지 아는 것이다.
### 왜 버전이라는 생각이 필요해졌는가

MVCC의 출발점은 멋진 내부 구조가 아니라 답답한 운영 압력이다. 여러 사용자가 같은 데이터베이스를 읽고 쓰는 순간, 시스템은 셋 중 하나를 골라야 한다. 첫째, 읽는 동안 쓰기를 막아 일관성을 쉽게 얻는다. 둘째, 쓰는 동안 읽기를 막아 중간 상태 노출을 피한다. 셋째, 과거 값을 일정 기간 보관하고 reader마다 볼 수 있는 세계를 정한다. 첫 번째와 두 번째 선택은 단순하지만 동시성이 떨어진다. 세 번째 선택이 MVCC의 기본 방향이다. 버전을 남기고, snapshot 또는 read view로 볼 수 있는 버전을 고르고, 더 이상 필요 없는 버전을 정리하는 비용을 받아들인다.

PostgreSQL 공식 문서는 MVCC를 각 SQL statement가 어느 시점의 snapshot을 보는 방식으로 설명하고, 이 방식이 전통적인 locking 방식보다 reader와 writer의 충돌을 줄인다고 설명한다. MySQL InnoDB 공식 문서는 multi-versioning에서 undo log와 read view가 consistent read에 쓰이고, update undo log가 오래된 snapshot 때문에 버려지지 못할 수 있다고 설명한다. 두 엔진의 저장 위치는 다르지만 문제의 압력은 같다. 읽기와 쓰기를 덜 막으려면 과거를 보관해야 하고, 과거를 보관하면 언젠가 안전하게 지워야 한다.

```text
락 중심 단순 모델
  reader A가 orders 전체 합계를 읽는 동안 writer B가 환불 row를 수정하려 한다.
  선택 1: reader가 끝날 때까지 writer 대기 -> 쓰기 지연 증가
  선택 2: writer가 끝날 때까지 reader 대기 -> 조회 지연 증가
  선택 3: reader는 시작 시점의 과거 version을 읽고 writer는 새 version을 만든다 -> cleanup 비용 발생

MVCC 모델의 새 책임
  version storage  : 과거 row image를 어디에 둘 것인가?
  visibility rule  : reader가 어떤 version을 볼 수 있는가?
  cleanup horizon  : 이제 어떤 과거 version을 지워도 되는가?
  writer ordering  : 같은 row를 동시에 쓰려는 writer는 어떻게 조정하는가?
```

이 네 책임을 함께 말해야 MVCC가 lock-free라는 오해를 피할 수 있다. MVCC는 읽기 경로 일부에서 writer와의 직접 충돌을 줄인다. 그러나 writer끼리 같은 row를 바꾸면 여전히 lock, conflict, deadlock, serialization failure가 생긴다. 과거 version을 보여 주는 능력과 미래 변경 권리를 예약하는 능력은 다른 문제다.

### 기존 mvcc.md seed를 어떻게 승격해 읽을 것인가

기존 `database/mvcc.md`의 좋은 점은 MVCC를 단일 제품 기능으로 보지 않고 역사, MGA, rollback segment, transaction id, read view를 넓게 모았다는 데 있다. 다만 deep-dive 문서에서는 그 원재료를 그대로 반복하지 않고 현재 학습자가 손으로 재구성할 수 있는 형태로 바꿔야 한다. 역사 이름은 배경을 설명하는 데 쓰고, 제품별 현재 동작은 공식 PostgreSQL/MySQL 문서와 실험으로 닫는다.

MGA에 가까운 모델은 PostgreSQL을 이해하는 데 좋다. UPDATE가 old tuple을 즉시 덮어쓰지 않고, heap에 새 tuple을 추가하며, old tuple에는 끝난 transaction 정보가 붙는다. rollback/undo 계열 모델은 InnoDB를 이해하는 데 좋다. clustered record는 최신 값 쪽으로 움직이고, 과거 값은 undo chain으로 복원된다. 둘을 같은 그림으로 합치면 오해가 생긴다. 같은 업무 사건을 두 엔진이 다른 물리 구조로 구현한다는 점을 초반부터 고정해야 한다.

```text
같은 UPDATE를 두 번 그리기

업무 사건:
  account_id=7의 balance를 100에서 120으로 바꾼다.

PostgreSQL식 사고:
  heap tuple old: xmin=10, xmax=30, balance=100
  heap tuple new: xmin=30, xmax=--, balance=120
  snapshot이 tx 30을 인정하지 않으면 old를 본다.

InnoDB식 사고:
  clustered record current: DB_TRX_ID=30, balance=120, roll_ptr=undo#7
  undo#7 old image   : old_trx_id=10, balance=100
  read view가 tx 30을 인정하지 않으면 undo#7로 내려간다.
```

이 비교는 저장 위치를 외우기 위한 표가 아니다. 나중에 PostgreSQL vacuum과 InnoDB purge를 다르게 이해하기 위한 준비다. PostgreSQL은 heap에 남은 dead tuple을 vacuum이 정리한다. InnoDB는 더 이상 필요 없는 undo history와 delete-marked record를 purge가 정리한다. 이름을 섞으면 장애 대응 검색어도 틀어진다.

### snapshot은 최신성이 아니라 일관성의 약속이다

MVCC를 처음 배운 개발자가 자주 묻는 질문은 `왜 방금 commit된 값이 안 보이지?`다. 답은 대개 snapshot 또는 read view에 있다. snapshot은 최신 값을 보겠다는 약속이 아니라, 특정 시점의 일관된 세계를 보겠다는 약속이다. PostgreSQL READ COMMITTED에서는 statement마다 snapshot이 달라질 수 있고, PostgreSQL REPEATABLE READ에서는 transaction 안에서 더 안정된 snapshot을 유지한다. InnoDB도 REPEATABLE READ와 READ COMMITTED에서 consistent read의 read view 생성과 재사용이 달라진다.

이 차이를 보고서 예제로 보면 쉽다. 주문 합계를 읽고, 결제 합계를 읽고, 재고 합계를 읽는 긴 보고서가 있다고 하자. 중간에 다른 transaction이 결제를 환불하면, 보고서는 주문은 환불 전, 결제는 환불 후, 재고는 또 다른 시점으로 섞어 볼 수 있다. 안정된 snapshot은 보고서가 같은 시점의 세계를 보게 만든다. 하지만 그 대가로 보고서는 최신 환불을 보지 못할 수 있다. 이때 `stale`이라는 단어를 쓰기 전에, 요구사항이 최신성인지 일관성인지 구분해야 한다.

```text
보고서 읽기 trace
  t0: snapshot/read view 생성
  t1: SELECT 주문 합계 -> 1,000,000
  t2: 다른 transaction이 환불 100,000 commit
  t3: SELECT 결제 합계 -> snapshot 기준이면 여전히 환불 전 값
  t4: report는 t0의 세계를 일관되게 보여 줌

업무 질문
  - 이 화면은 최신 환불을 반드시 보여야 하는가?
  - 아니면 한 시점의 합계가 서로 맞는 것이 더 중요한가?
  - read replica를 읽는다면 MVCC snapshot과 replica lag를 따로 구분했는가?
```

이 질문을 하지 않으면 MVCC를 알면서도 사용자 화면을 잘못 설계한다. 관리자 대시보드는 최신성이 더 중요할 수 있고, 회계 보고서는 한 시점 일관성이 더 중요할 수 있다. 같은 SELECT라도 업무 목적이 다르면 transaction 경계와 isolation 선택이 달라진다.

### writer 충돌은 사라지지 않는다

MVCC는 reader와 writer의 직접 대기를 줄일 수 있지만 writer끼리의 순서를 없애지 않는다. 두 transaction이 같은 재고 row를 동시에 차감하면 둘 중 하나는 기다리거나 실패하거나 재시도해야 한다. 두 transaction이 서로 다른 row를 반대 순서로 잡으면 deadlock도 가능하다. 이 사실을 빼고 MVCC를 설명하면 재고 차감, 좌석 예약, 쿠폰 사용 같은 코드를 위험하게 만든다.

```text
writer 충돌 예
  A: UPDATE seats SET reserved=true WHERE seat_id=1;
  B: UPDATE seats SET reserved=true WHERE seat_id=1;

MVCC가 해 주는 일:
  reader는 과거 version을 볼 수 있다.

MVCC만으로 안 되는 일:
  두 writer가 같은 seat_id의 미래 소유권을 동시에 얻는 것.

필요한 장치:
  row lock, unique constraint, 조건부 update, retry, idempotency key
```

이 경계가 중요하다. `SELECT로 봤을 때 비어 있었다`는 것은 예약 권리를 얻었다는 뜻이 아니다. 업무 자원을 선점하려면 쓰기 조건이나 lock이 필요하다. MVCC는 과거를 읽는 기술이지 미래를 예약하는 기술이 아니다.

### cleanup은 MVCC의 뒤쪽 절반이다

버전을 남기는 기술은 반드시 cleanup 질문을 만든다. 누가 과거 version을 필요로 하는가, 그 필요가 언제 끝나는가, 끝난 뒤 누가 어떤 비용으로 정리하는가. PostgreSQL에서는 이 질문이 vacuum, visibility map, freeze, transaction id wraparound로 이어진다. MySQL InnoDB에서는 purge, history list length, undo tablespace, long running transaction으로 이어진다.

오래 열린 read-only transaction도 비용을 만든다. 겉으로는 아무 row도 수정하지 않지만, 그 transaction의 snapshot/read view가 과거 image를 필요로 할 수 있으면 cleanup은 기다린다. 그래서 `읽기 전용이면 운영에 무해하다`는 말은 틀릴 수 있다. 보고서, export, 백업, 디버깅 세션, idle in transaction이 모두 cleanup horizon을 붙잡을 수 있다.

```text
MVCC cleanup 사고 흐름
  1. UPDATE/DELETE가 old version을 만든다.
  2. 오래된 reader가 그 old version을 볼 수 있다.
  3. cleanup은 old reader가 끝날 때까지 기다린다.
  4. update/delete가 계속 들어오면 old version 또는 undo history가 쌓인다.
  5. storage, buffer, index scan, purge/vacuum workload 비용으로 드러난다.
```

이 흐름은 PostgreSQL과 InnoDB 모두에서 같은 질문으로 시작하지만 관측 명령은 다르다. PostgreSQL은 `pg_stat_activity`, `backend_xmin`, `pg_stat_all_tables`, vacuum log를 본다. InnoDB는 `information_schema.innodb_trx`, `SHOW ENGINE INNODB STATUS`, history list length, undo tablespace를 본다. 같은 원리를 같은 명령으로 보려 하면 실패한다.

### 역사 절에서 남겨야 할 실무 감각

MVCC 역사를 배운 뒤 남아야 하는 것은 연표가 아니라 trade-off 감각이다. 과거 version을 보관하면 읽기 일관성과 writer 진행성을 얻을 수 있다. 하지만 저장 공간, cleanup, transaction id 관리, writer conflict 처리, 오래 열린 transaction 감시가 필요해진다. 이것이 MVCC의 장점과 비용이다. 성능 기술이라고만 부르면 cleanup 비용이 빠지고, 복잡한 내부 구현이라고만 부르면 reader/writer 충돌 감소라는 이점이 빠진다.

실무에서는 증상을 네 갈래로 나눠 본다. 오래된 값 조회는 snapshot/read view, replica lag, application cache를 의심한다. writer 대기는 row lock, gap/next-key lock, deadlock, serialization failure를 의심한다. 저장 공간 증가는 dead tuple, undo history, bloat, long transaction을 의심한다. 재시도 오류는 isolation conflict와 deadlock victim을 구분한다. 이 분기표가 있어야 MVCC 지식이 운영 판단으로 전환된다.

```text
증상에서 MVCC 질문으로 이동하기

오래된 값을 본다
  -> 현재 transaction snapshot인가?
  -> read replica 지연인가?
  -> application cache인가?

쓰기 요청이 대기한다
  -> 같은 row writer 충돌인가?
  -> range lock 또는 next-key lock인가?
  -> deadlock 감지 후 victim이 되었는가?

테이블 또는 undo가 커진다
  -> 오래된 snapshot/read view가 cleanup을 막는가?
  -> autovacuum/purge가 부하를 못 따라가는가?
  -> batch transaction이 너무 큰가?
```

이런 질문을 자연스럽게 던질 수 있으면 MVCC는 더 이상 면접용 정의가 아니다. 특정 엔진을 만나도 먼저 version, visibility, cleanup, writer coordination을 찾게 된다. 이 네 축은 앞으로 DU28, DU29, isolation anomaly, lock/deadlock 절을 읽는 공통 지도다.

### 손으로 끝내는 복습 trace

마지막으로 같은 UPDATE를 세 번 그려 보면 된다. 첫 번째는 논리 row만 그린다. `balance=100`이 `balance=120`이 된다. 두 번째는 PostgreSQL식 heap tuple로 그린다. old tuple과 new tuple이 함께 남고 snapshot이 둘 중 하나를 고른다. 세 번째는 InnoDB식 clustered record와 undo chain으로 그린다. 현재 record가 120을 담고 undo가 100을 담으며 read view가 어느 쪽을 볼지 정한다.

```text
자가 판정
  - 엔진 이름을 가려도 왜 version이 필요한지 말할 수 있는가?
  - PostgreSQL과 InnoDB의 과거 image 저장 위치를 구분할 수 있는가?
  - snapshot/read view가 최신성 약속이 아니라 일관성 약속임을 설명할 수 있는가?
  - MVCC가 writer 충돌을 없애지 않는다는 반례를 들 수 있는가?
  - cleanup이 오래된 reader 앞에서 멈추는 이유를 운영 지표와 연결할 수 있는가?
```

이 질문에 답하면 MVCC의 역사와 등장 배경은 단순 배경지식이 아니라 설계 압력의 설명이 된다. 긴 읽기와 활발한 쓰기를 동시에 다루려는 요구가 version을 만들었고, version은 snapshot과 cleanup이라는 새 책임을 만들었다. 그 책임을 엔진별 공식 용어로 읽는 것이 이후 절의 기반이다.

### MVCC를 선택하지 않았을 때의 비용을 상상하기

MVCC가 왜 나왔는지 더 분명히 보려면 MVCC가 없는 읽기 많은 시스템을 상상하면 된다. 쇼핑몰 관리자가 하루 주문 리포트를 길게 읽는 동안 주문 상태 update가 전부 기다린다면 쓰기 처리량이 떨어진다. 반대로 update를 자유롭게 허용하면서 리포트가 매번 최신 page를 읽으면 주문 합계와 결제 합계가 서로 다른 시점을 섞을 수 있다. MVCC는 이 둘 사이에서 과거 version 보관 비용을 내고, reader에게 일관된 시점을 제공하며, writer가 계속 진행할 여지를 만든다.

이 선택은 모든 문제를 없애지 않는다. 과거 version을 보관하면 cleanup이 필요하고, cleanup은 오래된 snapshot 앞에서 기다린다. writer끼리의 충돌은 여전히 조정해야 한다. transaction id나 read view 같은 메타데이터도 관리해야 한다. 따라서 MVCC의 역사적 의미는 `락보다 좋다`가 아니라 `동시성 병목을 version과 cleanup 비용으로 바꾼다`에 가깝다.

```text
문제 변환
  before MVCC 고민:
    reader와 writer가 서로 기다림
    일관성을 얻으려면 긴 lock이 필요함

MVCC 이후 고민:
    version을 어디에 저장할 것인가
    snapshot/read view를 어떻게 만들 것인가
    writer 충돌을 어떻게 감지할 것인가
    오래된 version을 언제 정리할 것인가
```

이 관점은 이후 엔진 비교에도 도움이 된다. PostgreSQL과 InnoDB는 같은 문제를 다르게 변환한다. PostgreSQL은 heap tuple version과 vacuum의 언어로, InnoDB는 undo chain과 purge의 언어로 답한다. 같은 질문을 다른 공식 용어로 읽는 능력이 MVCC 학습의 실전 목표다.

### 면접 답변으로 압축할 때 버리면 안 되는 것

짧게 말해야 하는 자리에서도 MVCC 설명에는 세 문장이 남아야 한다. 첫째, 여러 transaction이 같은 데이터를 다룰 때 reader가 일관된 과거 시점을 볼 수 있게 version을 둔다. 둘째, 이 덕분에 reader와 writer의 직접 충돌을 줄이지만 writer끼리의 lock과 conflict는 사라지지 않는다. 셋째, 남겨 둔 version은 vacuum이나 purge 같은 cleanup으로 회수해야 하며 오래 열린 transaction이 이를 막을 수 있다.

이 세 문장을 빼고 `읽기와 쓰기가 서로 막지 않습니다`만 말하면 위험하다. 그 문장은 일반적인 장점을 말하지만 제한 조건을 숨긴다. `lock-free`라고 말하면 더 위험하다. 실제 DBMS는 MVCC와 lock을 함께 쓴다. 특히 업무 자원을 선점하는 코드는 조건부 write나 locking read가 필요하다.

```text
짧은 답변 골격
  MVCC는 row의 과거 version과 snapshot/read view를 이용해 reader가 일관된 시점을 보게 하는 방식입니다.
  그래서 일반 읽기와 쓰기의 직접 충돌을 줄일 수 있지만, 같은 row를 쓰는 writer 충돌이나 예약 문제까지 없애지는 않습니다.
  남은 version은 PostgreSQL의 vacuum, InnoDB의 purge 같은 정리 작업으로 회수되며, 오래 열린 transaction은 이 정리를 늦출 수 있습니다.
```

이 정도로 압축해도 version, visibility, cleanup, writer conflict가 모두 남는다. 이것이 filler 없는 압축이다.

### MVCC를 다른 주제로 넘기는 연결 문장

MVCC 역사 절은 여기서 끝나지만, 실제 학습은 다음 절들과 연결되어야 한다. PostgreSQL tuple visibility 절에서는 `version을 어디에 저장하는가`라는 질문이 heap tuple, xmin/xmax, vacuum으로 내려간다. InnoDB 절에서는 같은 질문이 clustered record, undo log, read view, purge로 내려간다. isolation anomaly 절에서는 `snapshot이 어떤 이상 현상을 막고 어떤 이상 현상은 남기는가`를 본다. lock/deadlock 절에서는 `writer coordination은 왜 여전히 필요한가`를 본다.

이 연결이 있어야 MVCC가 독립된 암기 항목으로 남지 않는다. MVCC는 transaction, isolation, storage, cleanup, operations를 이어 주는 중간 축이다. transaction lifecycle은 commit/rollback 경계를 만들고, MVCC는 그 경계들이 동시에 있을 때 각 reader가 어느 세계를 볼지 정한다. Vacuum과 purge는 그 세계들이 사라진 뒤 남은 과거를 회수한다. 이 흐름을 한 문장으로 말하면, DB는 현재 값 하나만 저장하는 것이 아니라 `누가 어느 시점의 값을 볼 권리가 있는지`까지 저장하고 관리한다.

```text
다음 절로 넘기는 질문
  DU28: PostgreSQL은 과거 version을 heap tuple로 어떻게 남기고 언제 vacuum하는가?
  DU29: InnoDB는 현재 clustered record와 undo chain으로 과거 image를 어떻게 재구성하는가?
  DU30: snapshot이 있어도 어떤 anomaly가 남고 어떤 isolation에서 거부되는가?
  DU32: MVCC가 있어도 lock, gap, predicate, deadlock은 왜 필요한가?
```

이 질문들이 자연스럽게 이어지면 MVCC의 역사와 등장 배경은 `옛날에 이런 기술이 나왔다`가 아니라 다음 학습 단원을 여는 reasoning spine이 된다.

### 학습자가 끝에 가져가야 할 한 문장

MVCC의 한 문장 요약은 `DB가 현재 값 하나만 관리하지 않고, 각 transaction이 볼 수 있는 과거의 세계를 version과 visibility 규칙으로 관리한다`이다. 이 문장에는 중요한 제한도 숨어 있다. 과거의 세계를 보여 주려면 과거 image가 필요하고, 과거 image가 필요 없어지는 시점을 알아야 하며, 같은 미래를 두 writer가 동시에 차지하지 못하게 조정해야 한다. 그래서 MVCC는 version storage, snapshot/read view, cleanup, lock을 함께 이해해야 한다.

이 한 문장을 PostgreSQL로 말하면 heap tuple, xmin/xmax, snapshot, vacuum이 된다. InnoDB로 말하면 clustered record, undo log, read view, purge가 된다. 면접이나 설계 리뷰에서 엔진 이름이 바뀌어도 이 변환을 할 수 있으면, MVCC를 정의로 외운 것이 아니라 구조로 이해한 것이다.

### 마지막 확인: 최신성과 일관성을 분리해 말하기

MVCC를 설명할 때 최신성과 일관성을 분리해 말할 수 있으면 오해가 크게 줄어든다. 최신 값을 보는 화면은 짧은 transaction, 새 statement, primary read, cache 무효화가 중요하다. 한 시점의 합계를 맞춰야 하는 보고서는 안정된 snapshot이 중요하다. 둘 다 `정확한 조회`처럼 보이지만 요구하는 시간 계약이 다르다. 이 차이를 말할 수 있어야 MVCC를 실제 요구사항에 연결할 수 있다.
## PostgreSQL tuple visibility와 vacuum

PostgreSQL의 MVCC는 heap tuple의 `xmin`과 `xmax`, transaction snapshot, transaction status, vacuum이 함께 만드는 시스템이다. 기존 `database/mvcc.md`는 PostgreSQL이 `t_xmin`, `t_xmax`, snapshot structure를 사용하고 old tuple과 new tuple을 heap에 함께 둔다고 정리했다. 이 절은 그 seed를 보존하면서 공식 PostgreSQL MVCC와 routine vacuuming 문서에 맞춰 visibility와 vacuum을 한 흐름으로 설명한다.

첫 모델은 `xmin은 이 tuple을 만든 transaction`, `xmax는 이 tuple을 끝낸 transaction`, `snapshot은 내가 볼 수 있는 transaction의 경계`다. UPDATE는 old tuple을 지우는 대신 old tuple의 종료 정보를 표시하고 new tuple을 만든다. SELECT는 후보 tuple들의 `xmin/xmax`와 snapshot을 비교해 현재 statement 또는 transaction이 볼 수 있는 version을 고른다.

이 절에서 하중을 지탱하는 공식 자료는 다음과 같다. 링크 자체가 결론은 아니며, 본문에서는 각 공식 문서가 제공하는 규칙을 계좌 이체, row version, read view, vacuum/purge trace와 연결해 다시 검증 가능한 형태로 풀어 쓴다.

- PostgreSQL MVCC: https://www.postgresql.org/docs/current/mvcc.html
- PostgreSQL routine vacuuming: https://www.postgresql.org/docs/current/routine-vacuuming.html
- Local seed: database/mvcc.md

```text
초기 INSERT by tx=10 commit
T1: ctid=(0,1), id=7, balance=100, xmin=10, xmax=--

UPDATE by tx=30, commit 전
old T1: ctid=(0,1), id=7, balance=100, xmin=10, xmax=30
new T2: ctid=(0,2), id=7, balance=120, xmin=30, xmax=--

snapshot S_old에서 tx=30이 active
- T1 xmin=10 visible
- T1 xmax=30 not visible as deletion
=> balance=100

snapshot S_new after tx=30 commit
- T1 xmax=30 visible as deletion
- T2 xmin=30 visible as creation
=> balance=120
```

| tuple 상태 | reader 관점 | vacuum 관점 | 관측 단서 |
|---|---|---|---|
| live | snapshot에서 보임 | 제거 불가 | 정상 조회 |
| recently dead | 새 snapshot에는 불필요하지만 old snapshot은 볼 수 있음 | horizon 전까지 제한 | backend_xmin, old xact |
| dead | active snapshot 누구도 필요 없음 | vacuum 후보 | n_dead_tup 감소 가능 |
| frozen | 오래된 xmin을 freeze | wraparound 방지 | autovacuum freeze 로그 |

### 1. xmin/xmax는 단순 숫자 대소 비교가 아니다

`xmax`가 있다고 무조건 invisible인 것은 아니다. 그 transaction이 내 snapshot에서 아직 active라면 old tuple은 여전히 보일 수 있다. `xmin`도 마찬가지로 transaction이 active이거나 미래라면 new tuple은 보이면 안 된다.

따라서 visibility는 tuple header, snapshot active set, transaction commit/abort 상태가 함께 만든다. 학습용으로는 생성자/종료자/관측시점이라는 세 단어를 먼저 잡는다.

### 2. READ COMMITTED와 REPEATABLE READ는 snapshot 재사용이 다르다

PostgreSQL READ COMMITTED에서는 statement마다 새 snapshot을 얻을 수 있어 같은 transaction 안의 두 SELECT가 다른 commit 결과를 볼 수 있다. REPEATABLE READ에서는 transaction snapshot이 더 안정적으로 유지된다.

두 세션 실험에서 이 차이를 관찰하면 isolation level이 추상 용어가 아니라 SELECT 결과 변화로 보인다.

### 3. system column은 학습용 관측이지 업무 API가 아니다

`xmin`, `xmax`, `ctid`를 SELECT하면 내부 움직임을 배우는 데 도움이 된다. 하지만 transaction id wraparound, freeze, HOT update, tuple movement 같은 이유로 업무 로직이 이 값에 의존하면 위험하다.

디버깅 손전등과 도메인 계약을 구분하는 것이 중요하다. 내부 column은 문제를 설명하게 돕지만, 비즈니스 규칙은 명시적 column과 constraint로 표현해야 한다.

### 4. vacuum은 MVCC 비용 회수 장치다

UPDATE/DELETE가 old tuple을 즉시 지우지 않기 때문에 vacuum이 필요하다. vacuum은 더 이상 어떤 active snapshot도 필요로 하지 않는 dead tuple을 정리하거나 재사용 가능하게 만든다.

autovacuum은 방해꾼이 아니라 PostgreSQL MVCC가 정상적으로 숨 쉬기 위한 배경 작업이다. 성능 문제가 났다고 끄는 것은 비용 회수를 뒤로 미루는 결정이다.

### 5. 오래 열린 transaction은 vacuum horizon을 붙잡는다

`BEGIN` 후 idle in transaction인 세션도 오래된 snapshot이나 xmin horizon을 남길 수 있다. 그러면 다른 세션들이 만든 dead tuple을 vacuum이 제거하지 못한다.

관측은 `pg_stat_activity`의 `xact_start`, `backend_xmin`, query, client 정보를 같이 본다. 오래된 세션을 종료하기 전에는 업무 소유자와 rollback 영향도 확인해야 한다.

### 6. dead tuple 증가는 storage 문제이자 query 문제다

dead tuple이 쌓이면 table과 index가 커지고 cache 효율이 떨어지며 scan 비용과 vacuum 비용이 늘 수 있다. 단순 disk 사용량 문제가 아니라 query plan과 응답 시간 문제로 이어진다.

그래서 `n_dead_tup` 같은 통계, autovacuum 시각, table size, index bloat 추정, slow query 변화를 함께 본다.

### 7. writer conflict는 MVCC에서도 남는다

reader가 old version을 볼 수 있다고 해서 writer끼리 같은 row를 동시에 마음대로 update하는 것은 아니다. row lock, update conflict, serialization failure가 발생할 수 있다.

애플리케이션은 serialization failure나 deadlock을 transient error로 보고 재시도할 수 있어야 한다. MVCC를 읽기 최적화로만 이해하면 write path 실패 처리가 빠진다.

### 8. HOT update는 tuple chain 감각을 넓힌다

index key가 바뀌지 않는 update에서는 Heap-Only Tuple 최적화가 index 갱신 비용을 줄일 수 있다. 자세한 내부는 별도 주제지만, logical row가 physical tuple chain이 될 수 있다는 감각을 강화한다.

이 감각은 pageinspect 실험, fillfactor, HOT update ratio, index bloat 분석으로 이어진다.

### 9. hint bit와 transaction status는 반복 판정 비용을 줄인다

기존 메모의 `t_infomask` 표는 PostgreSQL이 tuple 상태를 header와 transaction status로 관리한다는 중요한 단서다. 처음 학습에서는 bit 암기보다 commit/abort 판정이 반복될 수 있고 이를 줄이는 metadata가 있다는 역할을 잡는다.

세부 bit는 PostgreSQL source나 pageinspect와 함께 볼 때 의미가 생긴다. 본문 핵심은 visibility 경로다.

### 10. wraparound와 freeze는 오래된 transaction id 관리 문제다

transaction id는 무한히 커지는 추상 시간이 아니다. wraparound 위험 때문에 PostgreSQL은 오래된 tuple의 xmin을 freeze해 과거 transaction을 안전하게 처리한다.

운영에서 freeze vacuum 경고를 무시하면 심각한 장애로 이어질 수 있다. MVCC는 버전뿐 아니라 transaction id 수명 관리까지 포함한다.

```sql
SELECT pid, usename, state, xact_start, backend_xmin,
       now() - xact_start AS xact_age,
       wait_event_type, wait_event, query
FROM pg_stat_activity
WHERE xact_start IS NOT NULL OR backend_xmin IS NOT NULL
ORDER BY xact_start NULLS LAST;

SELECT schemaname, relname, n_live_tup, n_dead_tup,
       last_vacuum, last_autovacuum, vacuum_count, autovacuum_count
FROM pg_stat_all_tables
WHERE relname = 'accounts';

VACUUM (VERBOSE, ANALYZE) accounts;
```

검증 lab은 오래 열린 transaction 하나와 update/delete를 반복하는 다른 세션으로 만든다. 오래 열린 세션이 있을 때 dead tuple 정리가 제한되는지 보고, 세션 종료 뒤 vacuum이 더 진행되는지 확인한다. PASS는 `왜 이전에는 지울 수 없었고 왜 이후에는 지울 수 있는지`를 snapshot horizon으로 설명하는 것이다. FAIL은 dead tuple 숫자만 보고 vacuum 횟수를 늘리는 결론으로 바로 뛰는 것이다.

### PostgreSQL visibility를 운영 언어로 다시 말하기

PostgreSQL의 `xmin`과 `xmax`는 초보자에게 숫자 두 개처럼 보이지만, 실제로는 tuple이 어느 transaction에 의해 태어났고 어느 transaction에 의해 끝났는지를 알려 주는 단서다. reader가 tuple을 볼 때 DB는 그 숫자를 단순히 크기 비교하지 않는다. transaction status, snapshot의 active transaction 집합, commit 여부, 현재 command의 시점까지 함께 판단한다. 그래서 `xmin`이 작으면 항상 보인다는 식으로 외우면 곧 틀린다. 더 안전한 설명은 snapshot이 그 tuple의 탄생을 인정하고, 그 tuple의 종료를 아직 인정하지 않으면 보인다는 것이다.

운영에서 이 설명이 중요한 이유는 vacuum 때문이다. vacuum은 죽은 tuple을 지우는 청소 작업이지만, 무조건 많이 돌린다고 좋은 것도 아니고, 무조건 늦춘다고 안전한 것도 아니다. 오래 열린 transaction이 과거 snapshot을 붙잡으면, vacuum은 그 snapshot이 필요로 할 수 있는 tuple을 지울 수 없다. 이때 `n_dead_tup`이 늘고 table/index bloat가 커지며, 같은 SELECT가 더 많은 page를 읽어야 할 수 있다. 개발자는 느린 쿼리를 보고 인덱스만 추가하기 전에, update가 많은 테이블인지, 오래 열린 transaction이 있는지, autovacuum이 충분히 따라오는지 확인해야 한다.

```text
PostgreSQL MVCC 장애 trace
  증상:
    orders 조회 p95 증가, CPU보다 buffer read 증가
  데이터 변화:
    상태 update가 많은 테이블, dead tuple 증가
  붙잡힌 경계:
    오래 열린 transaction의 backend_xmin이 vacuum horizon을 뒤로 잡음
  관측:
    pg_stat_activity.xact_start
    pg_stat_activity.backend_xmin
    pg_stat_all_tables.n_dead_tup
    EXPLAIN (ANALYZE, BUFFERS)
  조치:
    오래 열린 transaction owner 확인
    autovacuum 설정과 update 패턴 재검토
    필요하면 batch 크기와 transaction 경계 축소
```

이 trace에서 중요한 점은 vacuum을 실행했다는 사실이 곧 해결을 뜻하지 않는다는 것이다. vacuum이 지울 수 없는 이유가 남아 있으면 수동 vacuum도 기대만큼 효과가 없을 수 있다. 반대로 오래 열린 transaction을 무작정 종료하면 사용자 작업이나 batch rollback이 발생할 수 있다. 따라서 실무 조치는 항상 owner, 업무 영향, 재시작 가능성, rollback 비용을 함께 확인한다.

PostgreSQL tuple visibility를 학습할 때 자주 생기는 또 다른 오해는 system column을 업무 로직에 쓰고 싶어지는 것이다. `xmin`, `ctid`는 학습과 디버깅에 유용하지만, 업무 식별자나 낙관적 lock version으로 쓰기에는 조심해야 한다. vacuum, update, tuple 이동, wraparound 관리 같은 엔진 내부 정책과 연결되어 있기 때문이다. 업무 version이 필요하면 명시적인 version column이나 updated_at 정책을 두는 편이 더 안전하다. system column은 엔진의 속마음을 보여 주는 관측 창이지, 애플리케이션 계약의 공개 API가 아니다.

마지막으로 PostgreSQL의 장점과 비용을 같은 문장에 넣어야 한다. PostgreSQL은 MVCC 덕분에 reader가 writer를 덜 막는 강한 장점을 가진다. 동시에 그 장점은 dead tuple, vacuum, freeze, transaction id 관리라는 운영 비용을 만든다. 이 두 면을 함께 말할 수 있어야 PostgreSQL은 heap tuple version과 vacuum을 통해 동시성과 복구 가능성을 유지하고, 그 대가로 오래 열린 transaction과 bloat를 관리해야 한다는 실무 문장이 된다.

PostgreSQL 장애를 회고할 때는 `vacuum이 늦었다`에서 멈추지 말고 왜 늦었는지 분해해야 한다. autovacuum worker가 부족했는지, table별 scale factor가 부하에 맞지 않았는지, 오래 열린 transaction이 horizon을 막았는지, update pattern이 HOT update를 충분히 활용하지 못했는지, index가 많아 update 비용과 cleanup 비용이 함께 늘었는지 확인한다. 같은 vacuum 지연이라도 조치는 다르다. 설정을 바꿔야 할 때도 있고, transaction 경계를 줄여야 할 때도 있으며, batch update를 쪼개야 할 때도 있다.

검증은 항상 실제 숫자로 닫는다. `pg_stat_all_tables`의 live/dead tuple 추세, `pg_stat_activity`의 오래 열린 transaction, `EXPLAIN (ANALYZE, BUFFERS)`의 읽은 page 수, autovacuum log를 같은 시간축에 놓는다. 이 네 가지가 같은 이야기를 하면 원인 신뢰도가 올라간다. 하나만 맞고 나머지가 맞지 않으면 아직 추론 단계다. 이런 태도가 있어야 PostgreSQL MVCC를 단순 내부 구조가 아니라 관측 가능한 운영 모델로 다룰 수 있다.

### 추가 판정 질문: vacuum 문제인가, plan 문제인가

PostgreSQL 운영에서 느린 조회가 나타나면 vacuum 문제와 plan 문제를 분리해야 한다. dead tuple과 bloat 때문에 더 많은 page를 읽는 경우도 있고, 통계가 낡아 optimizer가 잘못된 plan을 고른 경우도 있다. 두 문제는 함께 나타날 수 있지만 조치가 다르다. vacuum만 보면 통계 문제를 놓치고, index만 보면 죽은 tuple 비용을 방치한다.

그래서 `EXPLAIN (ANALYZE, BUFFERS)`, `pg_stat_all_tables`, autovacuum log, 오래 열린 transaction을 같이 본다. 실제 읽은 buffer가 늘고 dead tuple이 증가했으며 vacuum이 오래된 transaction 때문에 멈춰 있다면 MVCC cleanup 문제가 강하다. row estimate가 크게 틀리고 통계 갱신 뒤 plan이 바뀐다면 optimizer 쪽 근거가 강하다. 이 구분이 있어야 PostgreSQL을 감으로 튜닝하지 않는다.
### tuple visibility를 한 줄씩 판정하는 법

PostgreSQL visibility를 배울 때 가장 먼저 버려야 할 shortcut은 `xmin이 작으면 보이고 xmax가 있으면 안 보인다`이다. PostgreSQL은 tuple header만 보고 기계적으로 결정하지 않는다. tuple을 만든 transaction이 commit했는지, tuple을 끝낸 transaction이 내 snapshot에서 보이는지, 그 transaction이 아직 active인지, abort되었는지, 현재 command가 자기 transaction 안에서 만든 변화인지까지 본다. 그래서 학습할 때는 숫자의 크기보다 `탄생을 인정하는가`, `종료를 인정하는가`라는 질문으로 읽는 편이 안전하다.

```text
candidate tuple 판정

old tuple: id=7, balance=100, xmin=10, xmax=30
new tuple: id=7, balance=120, xmin=30, xmax=--

snapshot S1:
  tx 10은 이미 commit됨
  tx 30은 아직 active 또는 snapshot 이후 commit

S1에서 old tuple:
  탄생 xmin=10은 인정된다.
  종료 xmax=30은 아직 인정되지 않는다.
  => old tuple은 보인다.

S1에서 new tuple:
  탄생 xmin=30은 아직 인정되지 않는다.
  => new tuple은 보이지 않는다.
```

이 작은 판정이 DU28의 중심이다. `xmax가 있으니 old tuple은 죽었다`고 말하면 동시 update 중인 세션의 reader가 왜 old 값을 볼 수 있는지 설명하지 못한다. 반대로 `new tuple이 있으니 최신 값을 봐야 한다`고 말하면 snapshot isolation을 설명하지 못한다. visibility는 물리적으로 존재하는 tuple 중에서 내 snapshot이 인정하는 버전을 고르는 규칙이다.

### READ COMMITTED와 REPEATABLE READ를 결과 변화로 구분하기

PostgreSQL 공식 문서는 READ COMMITTED에서 일반 SELECT가 query 시작 시점의 snapshot을 본다고 설명한다. 같은 transaction 안에서도 두 SELECT 사이에 다른 transaction이 commit하면 두 번째 SELECT는 새 값을 볼 수 있다. 반면 REPEATABLE READ에서는 transaction 수준의 안정된 snapshot으로 이해해야 한다. 이 차이는 isolation 이름보다 두 세션 실험으로 익혀야 한다.

```sql
-- Session A
BEGIN ISOLATION LEVEL READ COMMITTED;
SELECT balance FROM accounts WHERE account_id = 7; -- 100
-- Session B가 balance=120으로 UPDATE 후 COMMIT
SELECT balance FROM accounts WHERE account_id = 7; -- READ COMMITTED라면 120 가능
COMMIT;

-- Session A 다시
BEGIN ISOLATION LEVEL REPEATABLE READ;
SELECT balance FROM accounts WHERE account_id = 7; -- 120
-- Session B가 balance=140으로 UPDATE 후 COMMIT
SELECT balance FROM accounts WHERE account_id = 7; -- 같은 transaction snapshot이면 120
COMMIT;
```

여기서 중요한 함정은 `트랜잭션 안이면 항상 같은 값을 본다`가 아니라는 점이다. PostgreSQL READ COMMITTED에서는 statement마다 snapshot이 달라질 수 있다. 따라서 애플리케이션이 한 transaction 안에서 여러 번 조회하며 같은 값을 기대한다면 isolation level과 lock 전략을 명시해야 한다. 반대로 최신 commit을 statement마다 보고 싶은 업무라면 READ COMMITTED가 자연스러울 수 있다. isolation은 좋고 나쁨이 아니라 어떤 관측 계약을 원하는지의 선택이다.

### vacuum은 청소가 아니라 안전한 회수다

PostgreSQL에서 UPDATE/DELETE가 old tuple을 즉시 제거하지 않는 이유는 오래된 snapshot이 그 tuple을 볼 수 있기 때문이다. Vacuum은 더 이상 어떤 active transaction도 필요로 하지 않는 tuple을 회수한다. 그래서 vacuum이 못 지웠다는 사실은 `vacuum이 부족하다`일 수도 있지만, `아직 지우면 안 된다`일 수도 있다. 공식 routine vacuuming 문서가 visibility map, wraparound 방지, freeze를 함께 다루는 이유도 여기에 있다. Vacuum은 공간 회수뿐 아니라 MVCC의 안전한 시간 관리를 맡는다.

```text
vacuum decision trace
  tuple T_old: xmin=10, xmax=30
  active snapshot 중 하나가 tx 30 이전 세계를 볼 수 있음
  -> T_old는 recently dead처럼 취급되어 제거하면 안 됨

오래된 snapshot 종료
  이제 active snapshot 누구도 T_old를 볼 수 없음
  -> vacuum이 T_old 공간을 재사용 가능하게 만들 수 있음
```

이 구조 때문에 `VACUUM을 더 자주 돌리면 항상 해결`이라는 말은 반쪽이다. 오래된 transaction이 horizon을 붙잡으면 vacuum을 실행해도 제거할 수 있는 것이 제한된다. 먼저 오래 열린 transaction, replication slot, prepared transaction, autovacuum 설정, update 부하를 함께 봐야 한다. Vacuum은 원인 해결 도구이기도 하지만, 원인이 남아 있으면 증상 완화에 그칠 수 있다.

### 오래 열린 transaction을 찾는 순서

PostgreSQL 운영에서 dead tuple이 늘고 vacuum이 따라오지 못하면 첫 관측은 `pg_stat_activity`다. `xact_start`가 오래되었는지, `backend_xmin`이 있는지, state가 `idle in transaction`인지, query와 application_name이 무엇인지 본다. 하지만 이 값을 보자마자 세션을 kill하면 안 된다. 읽기 전용 report인지, 배치 write인지, 사용자가 아직 기다리는 요청인지, rollback 비용이 큰지 확인해야 한다.

```sql
SELECT pid, application_name, client_addr, state,
       xact_start, now() - xact_start AS xact_age,
       backend_xmin, wait_event_type, wait_event,
       left(query, 120) AS query_sample
FROM pg_stat_activity
WHERE xact_start IS NOT NULL OR backend_xmin IS NOT NULL
ORDER BY xact_start NULLS LAST;
```

이 쿼리의 PASS는 오래된 세션 id를 찾는 것이 아니다. PASS는 그 세션이 어떤 업무 요청에서 생겼고, 어떤 테이블의 cleanup을 막을 가능성이 있으며, 종료하면 어떤 rollback이나 사용자 영향을 만들지 설명하는 것이다. FAIL은 `pid가 오래됐으니 종료` 또는 `n_dead_tup이 많으니 vacuum`으로 바로 뛰는 것이다. senior한 대응은 항상 증상, owner, 영향, 회복 경로를 함께 본다.

### dead tuple은 plan 문제와 섞여 보인다

Dead tuple이 많아지면 table size와 index size가 커지고 cache 효율이 떨어질 수 있다. 그러면 같은 query가 더 많은 page를 읽고, `EXPLAIN (ANALYZE, BUFFERS)`에서 shared hit/read가 늘 수 있다. 하지만 느린 query가 모두 vacuum 문제는 아니다. 통계가 낡아 cardinality estimate가 틀렸거나, 데이터 분포가 바뀌었거나, parameter 때문에 plan이 바뀌었을 수도 있다. PostgreSQL 진단은 vacuum과 optimizer를 분리해서 보되, 둘이 만날 수 있다는 점도 인정해야 한다.

```text
느린 조회 분기
  dead tuple/bloat 쪽 근거:
    pg_stat_all_tables.n_dead_tup 증가
    table/index size 증가
    오래 열린 transaction 또는 autovacuum 지연
    BUFFERS에서 읽는 page 증가

plan/statistics 쪽 근거:
    estimated rows와 actual rows 큰 차이
    ANALYZE 후 plan 변화
    특정 parameter에서만 나쁜 plan
    index 선택이 데이터 분포와 맞지 않음
```

이 분기를 닫지 않고 index를 추가하면, 죽은 tuple 때문에 커진 테이블에 index를 하나 더 얹어 update 비용과 vacuum 비용을 더 키울 수 있다. 반대로 vacuum만 보면 통계 문제를 놓칠 수 있다. 좋은 운영 문서는 두 원인을 경쟁 가설로 세우고 관측으로 좁힌다.

### system column은 관측 창이지 업무 계약이 아니다

`xmin`, `xmax`, `ctid`는 학습 실험에서 매우 유용하다. row version이 실제로 바뀌는 모습을 볼 수 있고, UPDATE 후 physical tuple이 달라지는 감각을 얻을 수 있다. 그러나 이 값들을 업무 로직의 공개 계약으로 쓰면 위험하다. PostgreSQL 내부 정책, vacuum, freeze, HOT update, tuple movement, transaction id wraparound 관리가 얽혀 있기 때문이다.

낙관적 locking이 필요하면 명시적 `version` column을 두는 편이 낫다. 업무 row 식별에는 primary key를 쓰고, 변경 감지에는 application-owned version이나 updated_at을 쓴다. system column은 장애 분석과 학습에서만 손전등처럼 사용한다. 손전등으로 길을 비출 수는 있지만, 손전등 자체를 도로 표지판으로 삼으면 안 된다.

```sql
-- 학습/디버깅용 관측
SELECT ctid, xmin, xmax, account_id, balance
FROM accounts
WHERE account_id = 7;

-- 업무 계약용 optimistic lock 예
UPDATE accounts
SET balance = balance - :amount,
    version = version + 1
WHERE account_id = :id
  AND version = :expected_version
  AND balance >= :amount;
```

이 구분은 문서 품질에서도 중요하다. 내부 column을 보여 주되, 독자가 그것을 실무 API로 오해하지 않게 바로 경계를 적어야 한다.

### freeze와 wraparound는 vacuum을 생존 장치로 만든다

PostgreSQL transaction id는 무한한 직선 시간이 아니다. 공식 문서가 wraparound를 강조하는 이유는 오래된 transaction id가 잘못 해석되면 과거 tuple이 미래 tuple처럼 보일 수 있기 때문이다. Vacuum은 오래된 tuple을 freeze하여 모든 현재와 미래 transaction에 안전하게 보이도록 만든다. 따라서 autovacuum freeze 경고는 단순 성능 알림이 아니라 write 중단으로 이어질 수 있는 생존 신호다.

이 내용은 초보자에게 어렵지만, 최소한 `vacuum은 디스크 청소만 하는 작업이 아니다`는 문장을 남겨야 한다. PostgreSQL은 MVCC를 위해 transaction id와 tuple visibility를 계속 관리한다. 이 관리가 밀리면 성능 문제가 아니라 가용성 문제가 된다. 그래서 운영자는 table별 age, autovacuum freeze, long transaction을 함께 본다.

### page와 index-only scan까지 이어지는 visibility map

Routine vacuuming 문서에서 visibility map을 다루는 이유도 visibility가 query 실행과 연결되기 때문이다. PostgreSQL index entry에는 tuple visibility 정보가 충분히 들어 있지 않기 때문에 일반 index scan은 heap tuple을 확인해야 할 수 있다. visibility map이 어떤 page의 tuple들이 모두 모든 transaction에 visible하다고 알려 주면 index-only scan이 heap fetch를 줄일 수 있다. 즉 vacuum은 단지 죽은 tuple을 치우는 것이 아니라, 어떤 page를 더 싸게 읽을 수 있는지에 대한 메타데이터도 갱신한다.

이 연결을 알면 vacuum 지연이 왜 SELECT 성능으로 돌아오는지 이해하기 쉽다. update가 많은 table에서 visibility map bit가 자주 내려가고 dead tuple이 늘면, index-only scan 기대가 깨지고 heap fetch가 늘 수 있다. `인덱스가 있는데 왜 heap을 읽지?`라는 질문은 MVCC visibility와 만난다.

### PostgreSQL DU의 운영형 마무리

이 절을 제대로 이해했다면 장애 앞에서 질문 순서가 달라진다. `vacuum을 돌릴까요?`보다 먼저 `누가 과거 tuple을 아직 필요로 하는가?`를 묻는다. `인덱스를 추가할까요?`보다 먼저 `실제 page read와 dead tuple, row estimate가 어떤 이야기를 하는가?`를 묻는다. `xmin을 써도 될까요?`보다 먼저 `이 값은 관측용인가 업무 계약인가?`를 묻는다.

```text
PostgreSQL MVCC 현장 점검표
  1. tuple visibility를 xmin/xmax 숫자 하나가 아니라 snapshot과 transaction status로 설명할 수 있는가?
  2. READ COMMITTED와 REPEATABLE READ의 결과 차이를 두 세션으로 재현할 수 있는가?
  3. vacuum이 못 지우는 이유를 오래된 snapshot, replication slot, autovacuum 설정, update 부하로 나눠 볼 수 있는가?
  4. 느린 query를 dead tuple/bloat 문제와 plan/statistics 문제로 분리해 볼 수 있는가?
  5. system column을 학습용 관측과 업무용 계약으로 구분할 수 있는가?
```

이 다섯 질문이 닫히면 PostgreSQL tuple visibility와 vacuum은 내부 구조 암기가 아니라 운영 판단 모델이 된다. UPDATE 하나가 old tuple과 new tuple을 만들고, snapshot이 볼 tuple을 고르며, vacuum이 더 이상 필요 없는 tuple을 안전하게 회수한다. 그 대가로 오래 열린 transaction과 bloat, freeze, visibility map, plan 변화까지 함께 봐야 한다.

### 작은 lab으로 visibility와 vacuum을 함께 보기

PostgreSQL lab은 단순히 `xmin`을 보는 데서 끝내면 약하다. 같은 UPDATE가 어떤 tuple version을 만들고, 오래 열린 transaction이 있을 때 vacuum이 왜 제한되는지 함께 봐야 한다. 실험은 세 세션이면 충분하다. 세션 A는 REPEATABLE READ transaction을 열고 row를 읽는다. 세션 B는 같은 row를 여러 번 update하고 commit한다. 세션 C는 `pg_stat_activity`, `pg_stat_all_tables`, `VACUUM VERBOSE`를 본다.

```text
실험 흐름
  A: BEGIN ISOLATION LEVEL REPEATABLE READ;
  A: SELECT ctid, xmin, xmax, balance FROM accounts WHERE account_id=7;

  B: UPDATE accounts SET balance = balance + 1 WHERE account_id=7;
  B: COMMIT;
  B: 이 작업을 여러 번 반복

  C: pg_stat_activity에서 A의 xact_start/backend_xmin 확인
  C: pg_stat_all_tables에서 dead tuple 증가 확인
  C: VACUUM VERBOSE 결과에서 제거 가능한 tuple과 남는 tuple 확인

  A: COMMIT;
  C: 다시 VACUUM 후 변화 확인
```

이 lab의 핵심은 `vacuum 전에는 안 됐고 나중에는 됐다`를 보는 것이 아니라, 왜 그 차이가 생겼는지 설명하는 것이다. A가 snapshot을 들고 있는 동안 B가 만든 old tuple 중 일부는 A에게 필요할 수 있다. A가 끝나면 cleanup horizon이 앞으로 이동하고 vacuum이 더 많은 tuple을 회수할 수 있다. 이렇게 설명할 수 있으면 dead tuple을 숫자가 아니라 visibility 규칙의 결과로 읽게 된다.

### 운영에서 바로 쓰는 판단 문장

PostgreSQL MVCC 장애를 회고할 때 좋은 문장은 원인과 조치를 분리한다. `autovacuum이 늦었다`는 증상에 가까울 수 있다. 원인은 오래 열린 transaction일 수도 있고, table별 autovacuum threshold가 부하에 맞지 않을 수도 있고, update pattern이 너무 크거나 index가 너무 많을 수도 있다. 조치는 각각 다르다. transaction owner를 끊는 것, batch를 줄이는 것, autovacuum 설정을 조정하는 것, index를 정리하는 것은 같은 해결책이 아니다.

```text
판단 문장 예
  관측: n_dead_tup 증가와 table size 증가가 동시에 보인다.
  추가 근거: 오래된 backend_xmin을 가진 idle in transaction 세션이 있다.
  해석: vacuum이 게으른 것이 아니라 제거 안전선이 뒤에 묶였을 가능성이 높다.
  조치: 세션 owner와 rollback 영향 확인 후 종료/수정, 이후 vacuum 진행과 query buffer read 재확인.
```

이렇게 쓰면 단정과 추론이 분리된다. PostgreSQL 운영에서 중요한 태도는 한 지표로 결론을 내리지 않는 것이다. `pg_stat_activity`, table stats, vacuum log, query plan, application log가 같은 이야기를 할 때 비로소 원인 신뢰도가 올라간다.

### tuple visibility를 코드 리뷰에 적용하기

PostgreSQL MVCC 지식은 DBA만의 지식이 아니다. 애플리케이션 코드 리뷰에서도 중요하다. 긴 streaming response 안에서 transaction을 열어 두는 코드, `@Transactional(readOnly=true)` 메서드 안에서 외부 API를 기다리는 코드, batch 전체를 한 transaction으로 묶는 코드, cursor를 닫지 않는 report 코드가 모두 vacuum horizon에 영향을 줄 수 있다. 읽기 전용이라는 이름은 안전 보증이 아니다. 오래 열린 snapshot은 cleanup을 늦출 수 있다.

코드 리뷰 질문은 구체적이어야 한다. 이 메서드는 transaction 안에서 네트워크 I/O를 하는가. result set을 모두 읽기 전에 사용자 응답이나 파일 쓰기를 오래 하는가. pagination export가 한 transaction에서 수십 분 유지되는가. 예외가 발생하면 cursor와 transaction이 닫히는가. 이런 질문은 성능 취향이 아니라 PostgreSQL MVCC의 비용 회수 경로와 연결된다.

```text
리뷰에서 잡아야 할 냄새
  - read-only transaction 안에서 외부 HTTP 호출
  - 대량 export가 하나의 transaction을 오래 유지
  - batch update가 수십만 row를 한 commit으로 처리
  - 예외 경로에서 connection/cursor 정리 누락
  - idle in transaction timeout이 없는 운영 세션
```

이 냄새를 발견하면 바로 금지하기보다 업무 요구를 본다. 정말 한 시점의 snapshot이 필요하면 transaction을 유지해야 할 수 있다. 다만 그 비용을 알고 timeout, batch size, replica 사용, export 방식, 관측 지표를 함께 설계해야 한다. PostgreSQL tuple visibility를 이해한다는 것은 이런 trade-off를 코드 리뷰 언어로 바꾸는 것이다.

### 장애 회고에 남겨야 하는 문장

PostgreSQL MVCC 장애 회고에는 `vacuum을 수행했다`보다 더 구체적인 문장이 필요하다. 예를 들어 `orders 테이블의 dead tuple 증가는 40분 동안 열린 report transaction의 backend_xmin 때문에 회수 경계가 뒤로 묶였고, report transaction 종료 후 vacuum이 dead tuple을 회수하면서 buffer read가 감소했다`처럼 원인과 관측, 조치가 이어져야 한다. 이런 문장은 다음 사람이 같은 증상을 만났을 때 바로 검증할 수 있다.

반대로 `autovacuum 문제였다`는 문장은 약하다. autovacuum worker 수 문제인지, scale factor 문제인지, 오래 열린 transaction 문제인지, update burst 문제인지 알 수 없기 때문이다. PostgreSQL tuple visibility를 제대로 배웠다면 회고 문장은 항상 snapshot horizon, dead tuple, vacuum 가능성, query 관측을 함께 담아야 한다.

### 마지막 확인: vacuum을 끄는 선택의 의미

성능 문제를 피하려고 autovacuum을 끄거나 과도하게 늦추는 선택은 PostgreSQL MVCC의 비용 회수 장치를 미루는 일이다. 잠깐의 I/O를 줄일 수는 있지만 dead tuple, visibility map 갱신 지연, freeze 위험, table bloat가 뒤에서 쌓일 수 있다. 정말 조정이 필요하다면 끄는 것이 아니라 table별 부하, update 비율, 업무 시간대, long transaction 패턴에 맞춰 threshold와 scale factor, cost limit, timeout을 조정해야 한다.

이 판단도 관측으로 닫는다. autovacuum 로그가 언제 시작하고 왜 멈추는지, table별 dead tuple이 얼마나 쌓이는지, 오래 열린 transaction이 반복되는지, vacuum 뒤 query buffer read가 줄어드는지를 본다. Vacuum은 귀찮은 background job이 아니라 PostgreSQL이 version을 남긴 대가를 회수하는 정규 경로다. 이 감각이 있으면 vacuum 설정 변경도 단순 튜닝이 아니라 MVCC 안정성 변경으로 다룰 수 있다.
## InnoDB read view, undo, purge

InnoDB의 MVCC는 read view, clustered index record의 transaction id, undo log chain, purge가 함께 만드는 시스템이다. 기존 `database/mvcc.md`는 InnoDB가 `TRX_ID`, `TRX_NO`, `ReadView`, `Roll Pointer`, undo block을 사용한다고 정리했다. 이 절은 그 seed를 보존하되 공식 MySQL 문서에 맞춰 consistent read, undo, purge를 하나의 row movement로 설명한다.

가장 중요한 repair는 PostgreSQL heap tuple 그림을 InnoDB에 그대로 옮기지 않는 것이다. InnoDB는 clustered index record가 최신 값 쪽으로 바뀌고, 이전 값은 undo log record에 남는다. consistent read는 read view 기준으로 현재 record가 너무 새롭다고 판단하면 rollback pointer를 따라 undo chain을 거슬러 올라가 과거 row image를 재구성한다.

이 절에서 하중을 지탱하는 공식 자료는 다음과 같다. 링크 자체가 결론은 아니며, 본문에서는 각 공식 문서가 제공하는 규칙을 계좌 이체, row version, read view, vacuum/purge trace와 연결해 다시 검증 가능한 형태로 풀어 쓴다.

- MySQL InnoDB consistent read: https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html
- MySQL InnoDB undo logs: https://dev.mysql.com/doc/refman/8.4/en/innodb-undo-logs.html
- MySQL InnoDB multi-versioning: https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html
- Local seed: database/mvcc.md

```text
초기 commit된 row
clustered record R0: id=7, balance=100, DB_TRX_ID=10, DB_ROLL_PTR=null

Session A creates read view RV_A while tx=30 is active
RV_A says tx=30 is not visible

Session B tx=30 updates and commits
clustered record R1: id=7, balance=120, DB_TRX_ID=30, DB_ROLL_PTR=undo#7
undo#7: old balance=100, old DB_TRX_ID=10, prev=null

Session A consistent read
1. R1의 DB_TRX_ID=30은 RV_A에서 보이면 안 된다.
2. DB_ROLL_PTR를 따라 undo#7로 간다.
3. undo#7의 old trx id=10은 RV_A에서 보인다.
4. A에게 balance=100을 반환한다.
```

| 요소 | 역할 | PostgreSQL과 다른 점 | 운영 관측 |
|---|---|---|---|
| read view | 어떤 trx id를 볼 수 있는지 정함 | snapshot structure와 목적은 유사하지만 구현 다름 | isolation별 반복 조회 |
| undo log | rollback과 consistent read 과거 image 제공 | heap old tuple이 아니라 undo chain | history list, undo tablespace |
| purge | 불필요한 undo history 제거 | vacuum과 이름/구조 다름 | SHOW ENGINE INNODB STATUS |
| locking read | current read/lock path | plain consistent read와 다름 | data_locks, lock wait |

### 1. read view는 보이는 transaction의 경계표다

read view는 현재 transaction이 어떤 다른 transaction의 변경을 볼 수 있는지 판정하는 기준이다. active transaction list와 경계 id를 바탕으로 record의 transaction id가 보이는지 판단한다.

용어 표기는 자료마다 다를 수 있으므로 `Up_Limit_ID` 같은 이름 암기보다 의미를 먼저 잡는다. 의미는 `내가 볼 수 없는 아직 진행 중이거나 미래인 transaction 집합`이다.

### 2. consistent nonlocking read는 과거 snapshot을 읽을 수 있다

일반 consistent read는 row lock을 잡지 않고 read view에 맞는 row image를 반환할 수 있다. 현재 clustered record가 read view에 맞지 않으면 undo chain으로 과거 값을 만든다.

이 때문에 writer가 commit한 뒤에도 오래된 transaction의 plain SELECT는 이전 값을 볼 수 있다. 이것은 stale bug가 아니라 선택한 isolation/read view의 결과다.

### 3. locking read는 plain SELECT와 다르다

`SELECT ... FOR UPDATE` 같은 locking read는 consistent read와 같은 경로로만 이해하면 안 된다. 수정 전 선점이 목적이므로 lock을 잡고 current version과 충돌한다.

재고 차감이나 좌석 예약에서 plain SELECT 후 애플리케이션 if문으로 판단하면 race가 생긴다. 조건부 UPDATE나 locking read, unique constraint, retry를 함께 설계해야 한다.

### 4. undo log는 rollback과 consistent read를 동시에 지원한다

undo log는 실패한 transaction을 되돌리는 데 쓰이고, read view가 과거 값을 요구할 때 row image를 재구성하는 데도 쓰인다. 그래서 단순 취소 로그라고만 부르면 MVCC 설명이 약해진다.

반대로 undo가 있으니 과거를 무한히 읽을 수 있다고 생각해도 틀리다. purge가 더 이상 필요 없는 undo history를 정리한다.

### 5. purge는 오래된 read view 앞에서 멈출 수 있다

어떤 transaction이 오래된 read view를 붙잡으면 purge는 그 read view가 필요로 할 수 있는 undo record를 제거할 수 없다. history list가 길어지고 undo tablespace 사용량이 늘 수 있다.

관측은 `information_schema.innodb_trx`, `SHOW ENGINE INNODB STATUS`, performance_schema lock 뷰, 애플리케이션 request log를 함께 본다.

### 6. history list length는 추세로 읽는다

history list length가 순간적으로 크다고 바로 장애는 아니다. 쓰기 부하가 높으면 늘었다가 purge가 따라잡을 수 있다. 문제는 오래된 transaction과 함께 지속적으로 증가하고 성능/공간 압력으로 이어지는 경우다.

따라서 한 번의 숫자보다 시간 흐름, purge 속도, undo tablespace, 오래된 trx_started를 함께 본다.

### 7. REPEATABLE READ와 READ COMMITTED의 read view 생성 차이를 실험한다

InnoDB 기본 REPEATABLE READ에서는 같은 transaction 안의 consistent read가 안정된 결과를 볼 수 있다. READ COMMITTED에서는 statement마다 read view를 새로 만들어 두 번째 SELECT가 새 commit을 볼 수 있다.

실험은 isolation level을 바꿔 같은 두 세션 update를 반복하면 된다. 결과 차이가 read view 재사용 차이를 보여 준다.

### 8. secondary index만 보고 MVCC 비용을 없다고 말하지 않는다

보조 인덱스로 후보를 찾더라도 version visibility 판단에는 clustered record와 transaction metadata, undo가 관여할 수 있다. delete-marked record와 purge 전 index entry도 range query와 성능에 영향을 줄 수 있다.

covering index가 항상 MVCC 비용을 0으로 만든다는 식의 단정은 위험하다. 실제 plan과 engine status를 함께 본다.

### 9. gap/next-key lock은 MVCC 밖의 보조 장치가 아니다

InnoDB isolation은 MVCC consistent read만으로 완성되지 않는다. range update, phantom 방지, locking read에서는 gap lock과 next-key lock이 등장한다.

이 절은 visibility 중심이지만, 실전에서는 locks/latches/deadlocks 절과 함께 읽어야 한다. `MVCC니까 phantom이 없다` 같은 짧은 답은 엔진 동작을 가린다.

### 10. 큰 transaction rollback 자체가 비용이다

오래된 transaction을 kill하면 purge는 풀릴 수 있지만 rollback이 오래 걸릴 수 있고 업무 작업이 실패한다. 특히 큰 update transaction은 되돌림도 많은 undo를 소비한다.

운영 대응은 소유자 확인, 업무 영향, 재시작 가능성, chunk 설계 여부를 보고 결정한다. kill은 관측 뒤 선택하는 조치이지 첫 반응이 아니다.

```sql
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
START TRANSACTION;
SELECT id, balance FROM accounts WHERE id = 7;
-- 다른 세션이 UPDATE 후 COMMIT
SELECT id, balance FROM accounts WHERE id = 7;
SELECT id, balance FROM accounts WHERE id = 7 FOR UPDATE;
COMMIT;

SELECT trx_id, trx_started, trx_state, trx_isolation_level, trx_rows_locked, trx_query
FROM information_schema.innodb_trx
ORDER BY trx_started;

SHOW ENGINE INNODB STATUS\G
```

검증 lab은 세 가지를 나란히 본다. REPEATABLE READ plain SELECT가 같은 값을 유지하는지, READ COMMITTED plain SELECT가 새 commit을 보는지, `FOR UPDATE`가 plain consistent read와 다른 대기/관측을 보이는지 확인한다. PASS는 각 결과를 read view, undo chain, locking/current read로 설명하는 것이다. FAIL은 `InnoDB는 MVCC라서 읽기는 항상 락이 없다` 또는 `REPEATABLE READ면 어떤 SELECT든 같은 값`이라고 말하는 것이다.

### InnoDB read view를 운영 장애로 연결하기

InnoDB의 read view는 어떤 transaction의 변경을 볼 수 있고 어떤 변경은 아직 볼 수 없는가를 정하는 경계표다. PostgreSQL tuple visibility를 heap tuple 관점에서 그렸다면, InnoDB는 clustered record와 undo chain을 함께 그려야 한다. 현재 clustered record의 transaction id가 내 read view에서 보이면 그 값을 읽을 수 있고, 보이지 않으면 rollback pointer를 따라 undo record로 내려가 과거 image를 찾는다. 이 설명은 단순 내부 구현 지식이 아니라, 왜 오래된 read transaction이 undo history와 purge를 붙잡는지 이해하는 열쇠다.

실무에서 가장 위험한 shortcut은 plain SELECT는 lock을 안 잡으니 운영에 부담이 없다는 생각이다. 일반적인 consistent read는 writer를 직접 막지 않을 수 있지만, 오래 유지되는 read view는 purge가 과거 undo를 지우지 못하게 만든다. 즉 읽기는 lock wait 지표에 보이지 않는 방식으로 쓰기 많은 테이블의 뒤처리를 어렵게 만들 수 있다. 보고서 쿼리, 관리자 화면 export, 배치 검증 transaction이 오래 열려 있으면 history list가 길어지고, undo tablespace 압박과 consistent read 비용이 커진다.

```text
InnoDB purge 지연 trace
  1. SHOW ENGINE INNODB STATUS에서 history list length 추세를 본다.
  2. information_schema.innodb_trx에서 오래 열린 transaction을 찾는다.
  3. 오래 열린 transaction의 query, connection, application request id를 연결한다.
  4. update/delete가 많은 테이블과 시간대를 맞춘다.
  5. kill 여부는 rollback 비용과 업무 owner를 확인한 뒤 결정한다.
```

이 trace는 MySQL 운영에서 매우 현실적이다. history list length만 보고 purge thread가 느리다고 결론 내리면 반쪽이다. purge가 느린 원인은 thread 수나 I/O일 수도 있지만, purge가 지워도 되는 경계가 아직 앞으로 오지 않았기 때문일 수도 있다. 따라서 수치는 항상 transaction age와 함께 읽는다.

또 하나의 핵심은 consistent read와 locking read를 분리하는 것이다. 재고를 차감하려고 현재 수량을 확인하는 코드가 plain SELECT를 쓰면 과거 snapshot을 보고 판단할 수 있다. 그 뒤 UPDATE가 조건 없이 들어가면 race가 생긴다. 이때는 `SELECT ... FOR UPDATE`, 조건부 UPDATE, unique constraint, idempotency key 같은 현재 상태를 붙잡는 장치가 필요하다. MVCC는 읽을 수 있게 해 주는 기술이지, 업무 자원을 예약해 주는 기술이 아니다.

PostgreSQL과 InnoDB를 비교할 때도 언어를 섞지 않는 습관이 중요하다. PostgreSQL에서는 vacuum horizon, dead tuple, `pg_stat_activity.backend_xmin`을 말하고, InnoDB에서는 read view, undo history, purge, `information_schema.innodb_trx`, `SHOW ENGINE INNODB STATUS`를 말한다. 같은 원리를 다른 공식 언어로 표현할 수 있어야 장애 대응 검색어와 팀 내 커뮤니케이션이 정확해진다.

마지막으로 InnoDB MVCC의 기억 anchor는 다음 한 문장이다. InnoDB는 현재 clustered record만 보는 것이 아니라, 내 read view가 허락하지 않는 최신 record를 만나면 undo chain을 따라 과거 image를 찾고, 더 이상 필요 없는 undo는 purge가 회수한다. 이 문장에 read view, undo, purge가 모두 들어 있어야 InnoDB MVCC를 실제 운영 문제로 연결할 수 있다.

InnoDB에서 특히 조심할 장애는 “조회는 빠른데 뒤처리가 쌓이는” 형태다. 일반 SELECT가 빠르게 끝나더라도, 어떤 reporting transaction이 오래 열려 있으면 purge가 기다릴 수 있다. update/delete가 많은 서비스에서는 이 지연이 곧 undo tablespace 증가, history list 증가, buffer pool 압박으로 이어진다. 그래서 MySQL 운영자는 slow query log만 볼 것이 아니라 오래 열린 transaction과 purge 추세를 같이 보아야 한다.

또 하나의 실무 포인트는 rollback 비용이다. 큰 write transaction이 실패하면 undo를 이용해 되돌릴 수 있지만, 되돌리는 일 자체가 오래 걸릴 수 있다. 장애 상황에서 문제 transaction을 종료하는 것이 항상 즉시 복구를 뜻하지 않는다. rollback이 진행되는 동안 lock과 I/O가 계속 영향을 줄 수 있다. 따라서 대량 변경은 작은 batch, 명확한 checkpoint, 재시작 가능한 작업 단위로 설계해야 한다. MVCC와 undo를 이해하면 왜 큰 transaction을 조심해야 하는지 감으로가 아니라 구조로 설명할 수 있다.

마지막 검증은 용어 바꾸기다. PostgreSQL 문서에서 “dead tuple과 vacuum”이라고 부른 문제를 MySQL 문서에서는 “undo history와 purge”로 읽어야 한다. 같은 원리를 억지로 같은 단어로 통일하면 검색과 장애 대응이 늦어진다. 팀 문서에는 엔진별 공식 용어를 그대로 쓰고, 그 아래에 공통 질문을 붙이는 편이 좋다. 공통 질문은 단순하다. 누가 과거 version을 필요로 하는가, 그 필요가 언제 끝나는가, 끝난 뒤 누가 어떤 비용으로 정리하는가. 이 질문이 닫히면 InnoDB read view, undo, purge를 운영 언어로 설명할 수 있다.

### 추가 판정 질문: 현재 읽기인가, 일관 읽기인가

InnoDB에서 같은 SELECT처럼 보여도 plain consistent read와 locking read는 운영 의미가 다르다. plain SELECT는 read view에 맞는 과거 image를 반환할 수 있고, `SELECT ... FOR UPDATE`는 현재 row나 index range를 잠그며 이후 변경 권리를 잡으려 한다. 재고 차감에서 이 차이를 놓치면 화면 조회는 안정적인데 실제 차감은 race에 노출되는 구조가 된다.

검증은 두 세션으로 한다. 세션 A가 plain SELECT로 수량을 읽고 오래 머무는 동안 세션 B가 update 후 commit한다. 세션 A의 두 번째 plain SELECT가 어떤 값을 보는지 확인한 뒤, 같은 실험을 `FOR UPDATE`로 바꿔 lock wait와 결과를 비교한다. 이 실험을 통과하면 read view와 undo chain 설명이 업무 자원 선점 문제와 어떻게 다른지 몸으로 이해된다.
### read view를 현재 record와 undo chain 사이에서 읽기

InnoDB를 PostgreSQL처럼 heap에 old tuple과 new tuple이 나란히 남는 그림으로 설명하면 곧 헷갈린다. InnoDB의 기본 그림은 최신 clustered record와 과거 image를 담은 undo chain이다. consistent read가 현재 clustered record를 만났을 때, 그 record의 transaction id가 내 read view에서 보이면 현재 record를 읽는다. 보이지 않으면 rollback pointer를 따라 undo record로 이동하고, 그 과거 image가 read view에서 보일 때까지 거슬러 올라간다. 이것이 InnoDB MVCC 설명의 첫 고리다.

```text
read view RV-A가 tx 30을 볼 수 없다고 하자.

clustered record current
  id=7, balance=120, DB_TRX_ID=30, DB_ROLL_PTR=undo#7

undo#7
  old balance=100, old trx id=10, prev=null

판정
  1. current record의 DB_TRX_ID=30은 RV-A에서 보이지 않는다.
  2. DB_ROLL_PTR를 따라 undo#7로 간다.
  3. undo#7의 old trx id=10은 RV-A에서 보인다.
  4. SELECT 결과는 balance=100이다.
```

이 trace에서 중요한 것은 undo가 단순 rollback 메모가 아니라 consistent read의 과거 image 공급원이기도 하다는 점이다. MySQL 공식 undo log 문서도 undo record가 최신 변경을 되돌리는 정보이면서, 다른 transaction의 consistent read가 원래 데이터를 필요로 할 때 사용된다고 설명한다. 그러므로 undo를 `취소 로그`라고만 부르면 InnoDB MVCC의 절반이 빠진다.

### consistent read와 locking read를 분리해야 한다

InnoDB 공식 문서는 consistent nonlocking read가 특정 시점의 snapshot을 query에 제공한다고 설명한다. 또한 REPEATABLE READ에서는 같은 transaction 안의 consistent read가 첫 read가 만든 snapshot을 읽을 수 있고, READ COMMITTED에서는 각 consistent read가 새 snapshot을 만든다고 설명한다. 하지만 `SELECT ... FOR UPDATE` 같은 locking read는 업무 목적이 다르다. 과거 값을 안정적으로 읽는 것이 아니라, 앞으로 수정할 현재 row나 range를 잠그려는 것이다.

재고 차감 코드에서 이 차이는 치명적이다. 화면 조회는 plain SELECT로 충분할 수 있다. 하지만 실제 차감은 plain SELECT 후 애플리케이션 if문으로 처리하면 race에 노출된다. 현재 수량을 선점하려면 조건부 UPDATE, `SELECT ... FOR UPDATE`, unique constraint, idempotency key, retry 정책을 함께 써야 한다.

```text
재고 차감에서 두 읽기의 차이

plain SELECT
  - read view에 맞는 과거 image를 줄 수 있다.
  - writer를 직접 막지 않을 수 있다.
  - 조회 화면에는 좋지만 미래 차감 권리를 예약하지 않는다.

SELECT ... FOR UPDATE
  - current row 또는 range에 lock을 건다.
  - 다른 writer와 대기/충돌할 수 있다.
  - 차감, 예약, 선점 같은 업무에 필요할 수 있다.
```

이 구분을 모르면 `MVCC니까 SELECT는 안전하다`라는 말로 위험한 코드를 만든다. SELECT가 어떤 읽기인지, 그 결과로 무엇을 하려는지 먼저 말해야 한다.

### purge lag는 undo history와 오래된 read view의 이야기다

InnoDB purge는 더 이상 필요 없는 undo history와 delete-marked record를 회수한다. MySQL 공식 purge configuration 문서는 history list length가 purge lag를 나타내며, 오래 실행 중인 transaction이 consistent read 결과를 유지해야 하기 때문에 history list length 증가를 만들 수 있다고 설명한다. 특히 autocommit을 끄고 SELECT 후 COMMIT/ROLLBACK을 잊은 경우도 예로 든다. 이 문장은 운영에서 매우 중요하다. 읽기만 하는 transaction도 undo cleanup을 붙잡을 수 있기 때문이다.

```text
purge 지연 흐름
  t0: reporting transaction이 REPEATABLE READ에서 read view 생성
  t1: write workload가 같은 테이블을 계속 UPDATE/DELETE
  t2: update undo log와 delete-marked record가 쌓임
  t3: purge는 reporting read view가 필요로 할 수 있는 undo를 지우지 못함
  t4: history list length 증가, undo tablespace 압박, consistent read 비용 증가
  t5: reporting transaction 종료 후 purge가 따라잡기 시작
```

history list length를 볼 때는 순간값보다 추세가 중요하다. write-heavy 시간대에는 늘었다가 줄 수 있다. 문제는 오래된 transaction과 함께 계속 증가하고, undo tablespace나 query latency, purge lag delay로 이어지는 경우다. 따라서 `SHOW ENGINE INNODB STATUS`의 TRANSACTIONS 섹션과 `information_schema.innodb_trx`를 같은 시간축에서 본다.

### 관측 명령은 숫자가 아니라 이야기를 만든다

InnoDB 운영 진단은 `SHOW ENGINE INNODB STATUS` 한 번 복사해서 끝나지 않는다. TRANSACTIONS 섹션의 history list length, purge 진행 위치, 오래된 transaction, lock wait를 시간 순서로 읽어야 한다. 여기에 `information_schema.innodb_trx`의 `trx_started`, `trx_state`, `trx_query`, `trx_rows_locked`를 붙이면 어떤 세션이 read view나 lock을 붙잡는지 좁힐 수 있다. 가능하면 애플리케이션 log의 request id, connection id, endpoint까지 연결한다.

```sql
SELECT trx_id, trx_started, trx_state, trx_isolation_level,
       trx_rows_locked, trx_rows_modified,
       trx_mysql_thread_id, trx_query
FROM information_schema.innodb_trx
ORDER BY trx_started;

SHOW ENGINE INNODB STATUS\G
```

PASS는 `history list length가 200000이다`라는 숫자를 말하는 것이 아니다. PASS는 그 숫자가 증가 중인지 감소 중인지, 오래된 read transaction이 있는지, update/delete 부하가 어느 테이블에 집중되는지, purge가 I/O 때문에 느린지 horizon 때문에 못 지우는지 설명하는 것이다. FAIL은 undo tablespace만 늘리거나 purge thread만 늘리고, 오래된 transaction 원인을 남기는 것이다.

### 큰 transaction을 죽이는 결정은 신중해야 한다

오래된 transaction이 purge를 막는다면 종료하고 싶어진다. 그러나 write transaction을 kill하면 rollback이 오래 걸릴 수 있고, 그 rollback도 undo를 읽고 적용하며 I/O와 lock 영향을 남길 수 있다. read-only transaction이라도 사용자가 보는 보고서나 backup 작업일 수 있다. 따라서 kill은 첫 조치가 아니라 검증된 선택이어야 한다.

운영자는 먼저 transaction owner를 찾는다. 어떤 서비스, 어떤 endpoint, 어떤 batch, 어떤 request id인지 본다. 그다음 rows modified, rows locked, 실행 시간, 업무 재시작 가능성을 본다. 죽여도 되는 idle read인지, 죽이면 큰 rollback이 시작되는 write인지 구분한다. kill 이후에는 history list가 줄어드는지, purge가 따라잡는지, lock wait가 풀리는지까지 확인한다. 조치가 끝났다는 판단도 관측으로 닫아야 한다.

### secondary index와 delete-marked record도 비용을 만든다

InnoDB에서 보조 인덱스만으로 후보를 찾는 query도 MVCC와 무관하지 않다. 최종 row image와 visibility 판단에는 clustered record의 transaction metadata와 undo chain이 관여할 수 있다. delete-marked record가 purge 전까지 남아 있으면 range scan과 index maintenance 비용에도 영향을 줄 수 있다. 그래서 `covering index니까 MVCC 비용은 없다`라는 단정은 위험하다. 실제 plan, handler read, InnoDB status, purge 상태를 같이 봐야 한다.

이 내용은 DU15의 covering index와도 연결된다. covering index는 heap/clustered access를 줄일 수 있는 강력한 도구지만, transaction visibility와 cleanup 비용을 마술처럼 제거하지 않는다. 특히 update/delete가 많은 테이블에서는 index 수가 많을수록 write와 purge가 처리해야 할 구조도 늘어난다. 인덱스 설계는 읽기 최적화와 MVCC cleanup 비용을 함께 봐야 한다.

### gap lock과 next-key lock은 visibility와 다른 축이다

InnoDB REPEATABLE READ를 설명할 때 consistent read만 말하면 range write와 phantom 방지의 절반이 빠진다. `SELECT ... FOR UPDATE`나 range update는 record뿐 아니라 gap 또는 next-key lock을 잡을 수 있다. 이것은 새로운 row가 range 안으로 끼어들어 phantom을 만들지 못하게 하는 데 필요하지만, 개발자에게는 예상 못한 insert 대기로 보일 수 있다. 즉 InnoDB 동시성은 read view와 undo만이 아니라 lock 절과 함께 완성된다.

```text
range 선점 예
  A: SELECT * FROM coupons WHERE code BETWEEN 'A' AND 'M' FOR UPDATE;
  B: INSERT INTO coupons(code) VALUES ('K-100');

A가 읽은 row만 막는다고 생각하면 B가 왜 기다리는지 이해하지 못한다.
range 조건에서는 사이 공간까지 lock 범위가 될 수 있다.
```

이 절의 중심은 read view, undo, purge지만, 독자는 여기서 lock 절로 넘어갈 다리를 가져야 한다. consistent read와 current/locking read를 구분할 수 있어야 gap lock을 만나도 `MVCC인데 왜 대기하지?`라고 길을 잃지 않는다.

### MySQL 용어로 말하는 습관

PostgreSQL에서 dead tuple과 vacuum이라고 부르는 문제를 MySQL에서 그대로 `vacuum이 안 돈다`고 말하면 검색과 협업이 늦어진다. InnoDB에서는 read view, undo history, purge, history list length, delete-marked record, `information_schema.innodb_trx`, `SHOW ENGINE INNODB STATUS`를 말해야 한다. 공통 원리는 `누가 과거 version을 필요로 하는가`이지만, 엔진별 공식 언어는 다르다.

```text
공통 질문과 엔진별 언어

공통 질문: 누가 과거 version을 필요로 하는가?
  PostgreSQL: old snapshot, backend_xmin, dead tuple
  InnoDB    : read view, innodb_trx, undo history

공통 질문: 누가 정리하는가?
  PostgreSQL: vacuum/autovacuum
  InnoDB    : purge thread

공통 질문: 무엇을 관측하는가?
  PostgreSQL: pg_stat_activity, pg_stat_all_tables, vacuum log
  InnoDB    : SHOW ENGINE INNODB STATUS, history list length, innodb_trx
```

이 표는 용어 암기표가 아니라 장애 대응 번역기다. 한 엔진에서 배운 원리를 다른 엔진으로 옮길 때는 먼저 공식 용어로 다시 말해야 한다.

### InnoDB DU의 검증 루프

이 절을 닫는 실험은 세 가지다. 첫째, REPEATABLE READ에서 plain SELECT가 같은 read view를 유지하는지 본다. 둘째, READ COMMITTED에서 statement마다 더 최신 commit을 볼 수 있는지 본다. 셋째, 같은 조건을 `FOR UPDATE`로 바꿔 대기와 lock을 관찰한다. 여기에 오래 열린 transaction을 하나 두고 update/delete를 반복하여 history list length가 어떻게 움직이는지 보면 purge까지 연결된다.

```text
PASS 관측
  - plain consistent read 결과를 read view 생성 시점으로 설명한다.
  - current/locking read가 왜 대기하거나 lock을 잡는지 설명한다.
  - undo log가 rollback과 consistent read 양쪽에 쓰임을 말한다.
  - history list length 증가를 오래된 read view와 write workload로 연결한다.
  - kill 또는 commit/rollback 뒤 purge가 따라잡는지 확인한다.

FAIL 관측
  - 일반 SELECT가 항상 최신이라고 말한다.
  - SELECT FOR UPDATE를 consistent read와 같은 것으로 설명한다.
  - undo를 rollback 전용으로만 설명한다.
  - history list length 숫자만 보고 purge thread 설정부터 바꾼다.
```

마지막 기억 문장은 간단하다. InnoDB는 최신 clustered record에서 시작해 read view가 허락하지 않는 변화라면 undo chain을 따라 과거 image를 만들고, 더 이상 어떤 read view도 필요로 하지 않는 undo history는 purge로 회수한다. 이 문장 안에 read view, undo, purge가 모두 들어 있어야 InnoDB MVCC를 운영 문제와 연결할 수 있다.

### undo tablespace 증가를 공간 문제가 아니라 시간 문제로 보기

InnoDB에서 undo tablespace가 커지는 것을 보면 디스크부터 늘리고 싶어진다. 하지만 공간 증설은 시간을 벌 뿐 원인을 없애지 못할 수 있다. undo가 쌓인다는 것은 write workload가 과거 image를 만들고 있고, purge가 그 과거 image를 충분히 빨리 버리지 못한다는 뜻이다. 그 이유는 I/O나 purge thread 부족일 수도 있지만, 오래된 read view가 아직 그 image를 필요로 하기 때문일 수도 있다.

```text
undo 증가 분석 순서
  1. history list length가 일시 증가인지 지속 증가인지 본다.
  2. innodb_trx에서 오래 열린 transaction과 isolation level을 본다.
  3. update/delete가 많은 테이블과 시간대를 맞춘다.
  4. purge가 따라잡는 구간이 있는지 본다.
  5. 오래된 transaction 종료 뒤 history list가 감소하는지 확인한다.
```

이 순서가 중요한 이유는 조치가 달라지기 때문이다. 오래된 reporting transaction이 원인이면 transaction 경계와 export 방식을 바꿔야 한다. write burst가 purge 처리량을 압도한다면 batch 크기, purge 설정, I/O 여유를 본다. 큰 write transaction rollback이 진행 중이라면 기다려야 할 수도 있다. `undo tablespace를 늘렸다`는 조치만으로는 같은 장애가 다시 온다.

### read view와 replication lag를 섞지 않기

애플리케이션이 방금 쓴 값을 못 봤을 때 원인은 여러 가지다. 같은 primary transaction 안의 read view 때문일 수도 있고, READ COMMITTED/REPEATABLE READ 차이일 수도 있고, read replica lag일 수도 있으며, application cache일 수도 있다. InnoDB MVCC는 primary 내부 consistent read의 규칙을 설명한다. replica에서 늦게 보이는 문제는 replication 절의 consistency 계약이다. 둘을 섞으면 잘못된 조치를 한다.

```text
오래된 값 조회 분기
  primary 같은 connection인가?
    yes -> isolation level과 read view 생성 시점 확인
    no  -> read routing과 replica lag 확인

  같은 transaction 안인가?
    yes -> REPEATABLE READ인지 READ COMMITTED인지 확인
    no  -> autocommit과 새 transaction 여부 확인

  DB 밖 cache가 있는가?
    yes -> cache invalidation/ttl 확인
```

이 분기표가 있으면 `MySQL이 이상하다`는 막연한 결론 대신, 어느 층의 시간 차이인지 좁힐 수 있다. InnoDB read view는 강력한 개념이지만 모든 stale read의 원인은 아니다. 좋은 문서는 개념의 적용 범위를 같이 가르친다.

### InnoDB 문서를 읽을 때 놓치기 쉬운 경계

MySQL 문서의 consistent read 설명에는 예외와 경계가 많다. ordinary SELECT와 locking read가 다르고, isolation level에 따라 read view 재사용이 다르며, DDL과 섞이면 consistent read가 기대처럼 동작하지 않는 경우도 있다. 따라서 공식 문장을 인용할 때는 항상 `어떤 SELECT인가`, `어떤 isolation인가`, `같은 transaction 안인가`, `DDL이나 locking clause가 있는가`를 붙여야 한다.

이 경계는 실무 장애에서 바로 드러난다. 개발자는 REPEATABLE READ라고 생각하고 같은 값을 기대했지만 실제 코드는 transaction을 매번 새로 열 수 있다. 반대로 최신 값을 기대했는데 같은 transaction의 consistent read가 이전 read view를 유지할 수 있다. `FOR UPDATE`를 붙인 순간 plain SELECT와 대기 양상이 달라진다. ALTER TABLE 같은 DDL이 끼면 snapshot 설명만으로는 부족하다.

```text
InnoDB read 확인 질문
  - plain SELECT인가, FOR UPDATE/FOR SHARE인가?
  - isolation level은 REPEATABLE READ인가 READ COMMITTED인가?
  - 첫 consistent read가 이미 read view를 만들었는가?
  - 같은 physical connection과 transaction 안인가?
  - DDL, replication lag, application cache와 섞이지 않았는가?
```

이 질문을 매번 붙이면 InnoDB MVCC 설명이 과장되지 않는다. read view, undo, purge는 강력한 모델이지만 모든 읽기와 모든 최신성 문제를 설명하는 만능 단어가 아니다. 경계를 정확히 말하는 것이 고급 설명의 핵심이다.

### 장애 회고에 남겨야 하는 문장

InnoDB MVCC 장애 회고에서도 숫자 하나로 끝내면 안 된다. `History list length가 높았다`는 출발점일 뿐이다. 더 좋은 문장은 `REPEATABLE READ에서 열린 export transaction이 read view를 오래 유지했고, 동시에 주문 상태 UPDATE가 많아 update undo가 history list에 쌓였으며, export 종료 뒤 purge가 따라잡으면서 history list와 undo tablespace 증가가 멈췄다`처럼 시간 흐름을 담는다.

이 문장은 조치도 선명하게 만든다. purge thread부터 늘릴지, export transaction을 짧게 나눌지, autocommit off 세션을 감시할지, batch update를 줄일지 판단할 수 있다. InnoDB read view, undo, purge를 제대로 이해했다면 숫자를 외우는 것이 아니라, 누가 과거 image를 필요로 했고 그 필요가 언제 끝났는지 말할 수 있어야 한다.

### 마지막 확인: commit을 자주 하라는 말의 의미

MySQL 문서가 consistent read만 하는 transaction도 정기적으로 commit하라고 말하는 이유는 읽기가 아무 비용도 없어서가 아니다. 오래 열린 read view는 update undo log를 버리지 못하게 만들 수 있다. 따라서 export나 관리자 조회가 길어질 때는 한 transaction으로 오래 붙잡을 필요가 있는지, chunk마다 새 read view를 써도 되는지, 업무가 요구하는 일관 시점이 무엇인지 먼저 정해야 한다. Commit은 단순 종료 명령이 아니라 purge가 앞으로 나아갈 수 있게 하는 경계이기도 하다.
