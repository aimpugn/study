# MVCC는 lock-free 구호가 아니라 어떤 버전 가시성 계약인가?

MVCC는 Multi-Version Concurrency Control, 즉 하나의 논리 row에 여러 버전을 남기고 각 transaction이나 statement가 볼 수 있는 버전을 고르는 동시성 제어 방식이다. 중요한 말은 `여러 버전`보다 `볼 수 있는 버전을 고른다`다. MVCC는 읽기와 쓰기가 서로 덜 막히게 해 주지만, lock이 없어지는 기술도 아니고 모든 격리 문제를 자동으로 없애는 마법도 아니다. 쓰기 충돌, unique constraint, cleanup, vacuum, purge, 오래 열린 transaction, serialization failure는 여전히 남는다.

면접에서 MVCC를 `읽기는 lock을 안 잡습니다`로 설명하면 절반만 맞고 절반은 위험하다. Plain SELECT가 writer를 막지 않는 경우가 많다는 말은 맞다. 하지만 update는 새 버전을 만들고 이전 버전을 보존해야 하며, 어떤 transaction이 어떤 버전을 볼 수 있는지 판단하려면 transaction id, snapshot, undo record, tuple metadata 같은 계약이 필요하다. 또 cleanup은 그 old version을 아직 보는 transaction이 없는지 확인해야 한다. 그래서 MVCC의 중심 질문은 `lock을 쓰나 안 쓰나`가 아니라 `이 읽기에서 어느 version이 visible인가`다.

## 2-5분 개요

짧게 답하면 이렇게 말할 수 있다. MVCC는 DB가 row를 덮어쓰기 하나로만 보지 않고, 여러 version과 transaction metadata를 이용해 각 reader에게 일관된 view를 제공하는 방식이다. PostgreSQL은 update 때 heap에 새 tuple version을 만들고 이전 tuple에는 visibility metadata가 남는다. Reader는 snapshot과 tuple의 `xmin`, `xmax` 같은 정보를 비교해 보이는 tuple을 고른다. Vacuum은 더 이상 어떤 snapshot도 필요로 하지 않는 old tuple을 정리한다.

InnoDB는 설명의 방향이 다르다. InnoDB는 clustered index record에 최신 쪽 record를 두고, consistent read가 필요하면 read view와 undo log를 이용해 과거 version을 재구성할 수 있다. Update는 undo record를 남기고, purge thread는 더 이상 필요한 read view가 없을 때 old version을 정리한다. 그래서 PostgreSQL과 InnoDB 모두 MVCC지만 old version의 저장 위치와 cleanup 방식이 다르다.

MVCC가 주는 장점은 읽기와 쓰기의 충돌을 줄이는 것이다. T1이 오래 읽고 있어도 T2는 같은 row를 update하고 commit할 수 있다. T1은 자기 snapshot 기준에서 old version을 보고, T2 이후 시작한 읽기는 new version을 볼 수 있다. 하지만 두 writer가 같은 row를 update하면 여전히 대기하거나 충돌한다. `SELECT ... FOR UPDATE` 같은 locking read는 current row에 lock을 걸 수 있다. Serializable isolation에서는 위험한 read/write dependency를 감지하고 transaction을 abort시킬 수 있다.

따라서 좋은 답변은 네 문장을 포함한다. 첫째, MVCC는 snapshot 기준으로 visible version을 고르는 계약이다. 둘째, PostgreSQL은 tuple/xid/vacuum, InnoDB는 read view/undo/purge를 중심으로 설명한다. 셋째, reader-writer 충돌은 줄지만 writer-writer conflict와 cleanup 비용은 남는다. 넷째, 오래 열린 transaction은 old version 정리를 막아 성능과 저장 공간 문제를 만든다.

## 먼저 잡아야 할 작은 모델

가장 작은 모델은 같은 row가 update되면서 두 version이 생기는 상황이다.

```text
초기 상태
  account(id='A', balance=1000)

T1
  BEGIN;
  SELECT balance FROM account WHERE id='A'; -- 1000을 본다.

T2
  BEGIN;
  UPDATE account SET balance=900 WHERE id='A';
  COMMIT;

T1
  SELECT balance FROM account WHERE id='A'; -- 무엇을 볼까?
```

정답은 DBMS와 isolation level, snapshot 생성 시점에 달려 있다. `READ COMMITTED`라면 두 번째 SELECT가 T2 commit 이후 새 statement snapshot을 만들고 900을 볼 수 있다. `REPEATABLE READ`나 transaction snapshot을 유지하는 격리라면 T1은 계속 1000을 볼 수 있다. 중요한 것은 DB가 table 전체를 복사했다는 뜻이 아니다. Reader가 가진 snapshot 기준과 row version metadata를 비교해 어떤 version을 읽을지 결정한다는 뜻이다.

더 작은 trace는 이렇게 그릴 수 있다.

```text
version chain for account A

v1: balance=1000, created_by=T0, replaced_by=T2
v2: balance=900,  created_by=T2, replaced_by=none

T1 snapshot says:
  T2는 내가 snapshot을 만들 때 아직 완료되지 않았거나, 내 기준에서는 보이지 않는다.

T1 visible version:
  v2는 T2가 만든 버전이라 보이지 않는다.
  v1은 T1 기준에서 아직 살아 있는 버전이다.
```

이 모델에서 `snapshot`은 사진처럼 table을 물리적으로 복사한 결과가 아니다. Snapshot은 보이는 transaction과 보이지 않는 transaction을 판정하기 위한 기준이다. PostgreSQL에서는 snapshot이 active transaction id 범위와 목록을 담고, tuple의 `xmin`과 `xmax`를 비교한다. InnoDB에서는 read view가 transaction id 경계를 담고, 필요하면 undo chain을 따라 이전 값을 재구성한다.

이 모델을 모르면 MVCC 설명은 두 방향으로 틀어진다. 하나는 `lock이 없으니 빠르다`라는 성능 구호로 끝나는 것이다. 다른 하나는 PostgreSQL의 heap tuple 방식을 모든 MVCC 구현에 그대로 적용하는 것이다. 면접에서는 둘 다 위험하다. 공통 원리는 version visibility이고, 제품별 구현은 그 visibility를 어떤 metadata와 저장 구조로 실현하는지다. 이 visibility가 실제 업무 경계와 만나는 지점은 [트랜잭션 경계와 ACID](06-transaction-acid-boundary.md)에서 다시 이어진다.

## 깊은 메커니즘

### Snapshot은 복사본이 아니라 판정표다

Snapshot을 처음 배울 때 `그 시점의 DB 사진`이라고 비유할 수는 있다. 하지만 내부 동작을 설명할 때 이 비유를 그대로 믿으면 틀린다. DBMS가 큰 table 전체를 복사해서 transaction마다 들고 있으면 OLTP workload를 감당하기 어렵다. 실제로는 transaction id와 active transaction 목록, commit 상태, undo pointer, tuple header 같은 정보를 이용해 각 row version을 읽을 때마다 visible 여부를 판정한다.

PostgreSQL의 단순화된 visibility 질문은 다음과 같다.

```text
이 tuple을 만든 xmin transaction은 내 snapshot 기준에서 committed인가?
이 tuple을 삭제하거나 대체한 xmax transaction은 내 snapshot 기준에서 committed인가?
내가 볼 수 없는 transaction이 만든 tuple이면 invisible이다.
내가 볼 수 없는 transaction이 삭제한 tuple이면 아직 visible일 수 있다.
```

실제 규칙은 hint bit, frozen xid, command id, subtransaction 같은 세부를 더 가진다. 하지만 면접 답변의 중심은 이 정도면 충분하다. Tuple header의 transaction metadata와 reader snapshot이 만나 visible version을 고른다는 것이다.

InnoDB의 consistent read는 read view를 만든다. Read view는 대략 `이 read가 시작될 때 아직 active였던 transaction들`, `내 read view보다 확실히 과거에 완료된 transaction`, `내 기준에서는 아직 보이면 안 되는 transaction`을 가르는 경계다. 단순히 transaction id 숫자가 작으면 항상 보인다고 외우면 안 되고, active 목록과 commit 상태를 함께 봐야 한다. Reader가 현재 clustered record를 보았는데 그 record를 만든 transaction이 자기 read view에서 보이지 않으면, undo log를 따라 과거 모습을 재구성한다. 그래서 InnoDB MVCC를 말할 때는 undo log가 단순 rollback 자료가 아니라 consistent read의 과거 version 재구성에도 쓰인다는 점을 말해야 한다.

### PostgreSQL은 새 tuple을 만들고 old tuple을 나중에 치운다

PostgreSQL에서 update는 물리적으로 기존 row를 제자리에서 덮어쓰기보다 새 tuple version을 만든다고 이해하는 편이 좋다. Index entry와 HOT update 여부 같은 세부는 상황에 따라 달라지지만, 논리적으로는 old version과 new version이 잠시 공존한다. Old tuple은 `xmax` 등을 통해 더 이상 현재 버전이 아님을 나타내고, new tuple은 자기 `xmin`으로 누가 만들었는지 나타낸다.

```text
UPDATE account SET balance=900 WHERE id='A'

old tuple
  balance=1000
  xmin=T0
  xmax=T2

new tuple
  balance=900
  xmin=T2
  xmax=none
```

Reader는 자기 snapshot 기준으로 T2가 보이는지 확인한다. T2 commit 이전 snapshot이면 old tuple이 보일 수 있고, T2 commit 이후 statement snapshot이면 new tuple이 보일 수 있다. 이 구조 덕분에 reader는 writer가 만든 new version 때문에 무조건 막히지 않는다.

하지만 old tuple은 영원히 남을 수 없다. 더 이상 어떤 snapshot도 old tuple을 필요로 하지 않을 때 vacuum이 dead tuple을 정리한다. 오래 열린 transaction이 있으면 그 transaction의 snapshot이 old version을 볼 가능성이 있으므로 vacuum이 안전하게 지우지 못한다. 이때 table bloat, index bloat, autovacuum 지연, transaction id wraparound 위험 같은 운영 문제가 생긴다.

PostgreSQL 운영에서 MVCC를 의심할 때는 `pg_stat_activity`에서 오래 열린 transaction, `pg_stat_user_tables`의 dead tuple 수, autovacuum 동작, vacuum freeze 관련 지표를 본다. 면접 답변에서도 `MVCC라서 읽기가 빠릅니다`에서 끝내지 말고, 오래 열린 transaction이 cleanup을 막는 비용까지 말하면 훨씬 실전적인 답이 된다.

### InnoDB는 read view와 undo chain으로 과거를 재구성한다

InnoDB는 PostgreSQL처럼 old row version이 heap에 그대로 쌓이는 방식으로만 설명하면 안 된다. InnoDB record에는 transaction id와 roll pointer 같은 정보가 있고, update 전 값을 담은 undo record가 연결된다. Consistent read는 현재 record가 자기 read view에서 보이지 않으면 undo record를 따라가며 visible한 과거 모습을 찾는다.

```text
clustered index record now
  balance=900
  trx_id=T2
  roll_ptr -> undo record for balance=1000

T1 read view
  T2 is not visible

T1 reads
  current record invisible
  follow roll_ptr
  reconstructed old version balance=1000
```

Undo는 rollback에도 필요하지만, consistent read에도 필요하다. 그래서 오래 열린 read view가 있으면 purge가 undo history를 지우지 못하고 history list가 길어질 수 있다. 이 상태가 길어지면 undo tablespace 사용량, purge lag, read 성능 문제로 이어질 수 있다.

InnoDB에서 또 중요한 구분은 consistent nonlocking read와 locking read다. Plain SELECT는 read view를 통해 snapshot을 읽을 수 있지만, `SELECT ... FOR UPDATE`나 `LOCK IN SHARE MODE` 계열은 current row를 대상으로 lock을 건다. Update나 delete는 결국 current version과 충돌을 조정해야 한다. 따라서 `InnoDB MVCC는 lock을 안 씁니다`라는 답은 틀린다. 정확한 답은 `일관된 읽기에는 read view와 undo를 사용해 reader-writer blocking을 줄이지만, locking read와 write conflict에서는 lock을 쓴다`다.

### Isolation level은 snapshot 생성 시점과 conflict 처리 방식을 바꾼다

MVCC와 isolation level은 분리되어 있지만 연결되어 있다. Isolation level은 reader가 어떤 snapshot을 언제 만들고 얼마나 오래 유지하는지, 어떤 anomaly를 막기 위해 conflict를 어떻게 처리하는지를 정한다. PostgreSQL `READ COMMITTED`에서는 각 statement가 새 snapshot을 본다. 같은 transaction 안에서도 첫 SELECT는 old value, 두 번째 SELECT는 다른 transaction commit 이후 new value를 볼 수 있다. `REPEATABLE READ`에서는 transaction 시작 이후 같은 snapshot을 유지해 같은 query가 같은 committed view를 보게 한다.

MySQL/InnoDB의 기본 `REPEATABLE READ`는 consistent read에 transaction read view를 쓰지만, phantom 방어와 locking read에서는 next-key lock 같은 InnoDB 고유 전략이 등장한다. 같은 이름의 isolation level이라도 PostgreSQL과 MySQL/InnoDB에서 anomaly와 lock behavior가 완전히 같다고 말하면 안 된다.

Serializable도 제품별로 다르다. PostgreSQL은 Serializable Snapshot Isolation 계열로 위험한 dependency를 감지하고 serialization failure를 낼 수 있다. 애플리케이션은 이 오류를 정상적인 concurrency control 결과로 보고 transaction을 재시도해야 한다. InnoDB에서 serializable은 plain SELECT를 shared locking read처럼 다루는 등 다른 전략을 쓴다. 면접에서는 격리 수준 표를 외우기보다 `snapshot 생성 시점`, `read/write dependency`, `lock 범위`, `retry 필요성`을 함께 말하는 편이 안전하다.

### Writer-writer conflict는 사라지지 않는다

MVCC의 대표 장점은 reader와 writer가 덜 막힌다는 것이다. 하지만 두 writer가 같은 row를 동시에 update하려고 하면 하나는 기다리거나 실패한다. Unique index에 같은 key를 넣으려는 transaction들도 충돌한다. Predicate invariant를 동시에 깨는 write skew는 더 높은 isolation이나 별도 lock/constraint/retry가 필요하다.

```text
T1: UPDATE account SET balance = balance - 100 WHERE id='A';
T2: UPDATE account SET balance = balance - 200 WHERE id='A';

T2는 T1의 row lock 또는 version conflict 결과를 기다린다.
T1 commit/rollback 이후 T2는 현재 row를 기준으로 다시 판단한다.
```

이 흐름에서 MVCC는 old committed value를 reader에게 보여 줄 수는 있지만, 두 writer가 같은 current value를 동시에 덮어쓰게 두지는 않는다. DBMS는 write lock, current read, conflict detection을 통해 일관성을 맞춘다. 그래서 high-contention counter나 hot account row는 MVCC가 있어도 병목이 된다. 해결은 shard counter, batch aggregation, optimistic version check, queue serialization, business redesign처럼 workload에 따라 달라진다.

### Cleanup은 correctness의 일부다

MVCC는 old version을 남겨 reader에게 일관된 view를 제공한다. 하지만 old version을 언제 지워도 되는지 판단하지 못하면 storage가 계속 커진다. Cleanup은 단순 청소가 아니라 correctness와 연결되어 있다. 아직 어떤 transaction이 old version을 볼 수 있는데 지우면 그 transaction의 snapshot 계약이 깨진다. 반대로 아무도 필요로 하지 않는 old version을 오래 남기면 성능과 공간이 나빠진다.

PostgreSQL에서 vacuum은 dead tuple을 정리하고 visibility map, freeze 같은 작업과 연결된다. InnoDB에서 purge는 undo history를 정리한다. 이름은 다르지만 공통 질문은 같다.

```text
old version을 아직 볼 수 있는 snapshot이 있는가?
없다면 정리해도 되는가?
정리가 늦어지면 어떤 지표가 나빠지는가?
```

운영 장애에서 cleanup 지연은 종종 느린 query보다 오래 열린 transaction에서 시작된다. Report job이 transaction을 열어 둔 채 cursor로 천천히 읽거나, connection pool이 `idle in transaction` 상태를 방치하거나, migration script가 긴 transaction을 유지하면 old version이 지워지지 않는다. 좋은 MVCC 답변은 이런 운영 신호까지 이어져야 한다.


### Index와 visibility는 함께 움직인다

MVCC는 heap이나 clustered record만의 문제가 아니다. Query는 보통 index를 통해 후보 row를 찾고, 그 다음 visibility를 확인한다. PostgreSQL에서 index entry가 가리키는 heap tuple이 현재 snapshot에서 visible하지 않을 수 있다. 이때 DB는 다른 tuple version을 확인하거나, index-only scan이 가능하더라도 visibility map을 확인해야 한다. Visibility map은 page 안의 tuple들이 모두 visible하다고 판단할 수 있는지 알려 주어 index-only scan의 heap 접근을 줄이는 데 도움을 준다. 따라서 `index-only scan이면 heap을 절대 보지 않는다`는 답도 과장이다. Visibility 확인 조건이 맞을 때 heap 방문이 줄어드는 것이다.

InnoDB에서는 secondary index entry가 clustered primary key를 통해 clustered record로 이어진다. Consistent read가 필요한 경우 현재 record가 read view에서 invisible하면 undo를 따라 과거 version을 재구성한다. Secondary index만 보고 끝낼 수 있는지, clustered record를 봐야 하는지, undo를 따라가야 하는지는 query와 version 상태에 따라 달라진다. 그래서 MVCC는 optimizer와 storage engine의 경계에도 영향을 준다. 오래된 snapshot이 많고 update가 잦으면 단순 SELECT도 더 많은 version 판단을 하게 될 수 있다.

### Old version은 삭제 대상이 아니라 아직 누군가의 현재일 수 있다

MVCC를 이해하는 좋은 문장은 `old version은 과거 쓰레기가 아니라 어떤 snapshot에게는 현재 row다`이다. T1이 transaction snapshot을 잡은 뒤 T2가 update하고 commit하면, T2 이후 시작한 transaction에게는 new version이 현재다. 그러나 T1에게는 old version이 여전히 현재다. Cleanup은 이 old version을 필요로 하는 transaction이 하나도 없다는 판단이 닫힌 뒤에야 가능하다.

이 관점은 장애 대응에서 중요하다. Disk가 찬다고 old version을 강제로 지우면 snapshot 계약이 깨진다. Vacuum이나 purge를 빠르게 만들고 싶다면 먼저 오래 열린 transaction을 닫고, batch 크기를 줄이고, autovacuum/purge가 따라갈 수 있게 workload를 조정해야 한다. 원인 transaction을 보지 않고 cleanup 설정만 키우면, 잠시 나아지는 것처럼 보이다가 같은 문제가 반복된다.

### Read-only transaction도 비용을 만든다

읽기 전용 transaction은 데이터를 바꾸지 않으니 안전하다고 생각하기 쉽다. 하지만 snapshot을 오래 붙잡는 읽기 전용 transaction은 old version cleanup을 막을 수 있다. 보고서가 cursor로 몇 시간 동안 천천히 읽거나, application이 transaction을 연 채 사용자 입력을 기다리거나, connection pool이 `idle in transaction`을 방치하면 writer가 만든 old version이 정리되지 않는다.

PostgreSQL에서는 이런 상태가 dead tuple 증가와 vacuum 지연으로 보이고, InnoDB에서는 undo history 증가와 purge 지연으로 보일 수 있다. 따라서 read workload도 transaction boundary를 짧게 유지해야 한다. 긴 분석 query가 필요하면 replica, snapshot export, logical dump, analytical store처럼 OLTP와 분리된 경로를 검토한다. 이 선택은 성능 최적화가 아니라 OLTP cleanup 계약을 보호하는 선택이다.

### MVCC anomaly는 snapshot 일관성과 직렬 가능성이 다르다는 데서 나온다

Snapshot이 일관적이라는 말은 그 snapshot 안에서 각 row version 선택이 서로 모순되지 않는다는 뜻이다. 하지만 여러 transaction이 동시에 읽은 조건을 근거로 서로 다른 row를 update하면 전체 결과가 어떤 serial order와도 맞지 않을 수 있다. Write skew가 대표적이다.

```text
invariant: 최소 한 명의 의사가 on_call이어야 한다.
T1 snapshot: A=true, B=true -> A=false로 update
T2 snapshot: A=true, B=true -> B=false로 update
둘 다 commit하면 A=false, B=false
```

각 transaction은 자기 snapshot에서 합리적인 결정을 했다. 그러나 전체 결과는 invariant를 깬다. MVCC는 consistent snapshot을 제공하지만, 항상 serializable outcome을 보장하는 것은 아니다. 이 문제를 막으려면 serializable isolation과 retry, predicate를 보호하는 lock, guard row, constraint 재설계가 필요하다. 따라서 MVCC 문서는 [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)과 연결되어야 한다.

### Monitoring 문장을 만들 수 있어야 한다

면접에서 운영 경험을 묻는다면 지표 이름만 나열하지 말고 문장으로 바꿔야 한다. PostgreSQL에서 `dead tuples가 늘고 oldest transaction age가 크며 autovacuum이 반복적으로 지연된다`면, 오래 열린 transaction이 old tuple 정리를 막는다는 가설을 세운다. InnoDB에서 `history list length가 계속 늘고 purge thread가 따라가지 못한다`면, undo history 생성 속도와 purge 가능 조건을 본다. Serialization failure가 증가하면 hot predicate, transaction 크기, retry backoff를 본다.

좋은 관측 문장은 `무엇이 생성되고`, `무엇이 정리를 막고`, `어떤 사용자 증상으로 보이는지`를 포함한다. 예를 들어 `장시간 read transaction이 snapshot을 잡고 있어 update/delete가 만든 old version을 purge하지 못하고, 그 결과 table bloat와 cache 효율 저하로 query latency가 오른다`라고 말할 수 있어야 한다.

## DBMS별 경계

### PostgreSQL

PostgreSQL은 MVCC를 heap tuple과 transaction id 중심으로 설명한다. Tuple header에는 생성 transaction과 삭제 또는 대체 transaction을 판단할 수 있는 정보가 있고, snapshot은 어떤 transaction id가 visible한지 판단하는 기준을 제공한다. `READ COMMITTED`와 `REPEATABLE READ`, `SERIALIZABLE`의 snapshot 사용 방식이 다르며, Serializable에서는 serialization failure와 retry가 설계의 일부다.

Vacuum은 PostgreSQL MVCC 설명에서 빠지면 안 된다. Update/delete가 많고 오래 열린 transaction이 있으면 dead tuple이 쌓인다. Autovacuum이 제때 돌지 못하면 table과 index가 커지고, visibility map과 freeze 문제까지 연결될 수 있다. `VACUUM`은 단순한 성능 작업이 아니라 MVCC가 남긴 과거 version을 안전하게 정리하는 작업이다.

### MySQL/InnoDB

InnoDB는 read view, undo log, clustered index record, purge를 중심으로 설명한다. Consistent read는 snapshot 기준으로 undo chain을 따라 과거 version을 재구성할 수 있고, purge는 오래된 undo를 정리한다. 기본 isolation level과 locking read 동작은 PostgreSQL과 다르다.

InnoDB row-level lock은 index record lock과 밀접하게 연결된다. `SELECT ... FOR UPDATE`가 어떤 index range를 훑는지에 따라 record lock, gap lock, next-key lock 범위가 달라진다. 이는 MVCC 문서만으로 닫히지 않고 [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)과 함께 봐야 한다. 하지만 MVCC 관점에서 기억할 말은 plain consistent read와 locking read를 절대 섞지 않는 것이다.

### 다른 DBMS와 용어 경계

Oracle, SQL Server, distributed SQL 제품들도 여러 version 또는 snapshot 계열의 동시성 제어를 제공할 수 있다. 그러나 undo segment, tempdb version store, timestamp oracle, hybrid logical clock, consensus log 등 구현 경계는 다르다. 따라서 면접에서 특정 제품을 받으면 공통 원리를 먼저 말한 뒤, 그 제품의 version 저장 위치와 cleanup 방식을 분리해 말해야 한다.

`snapshot isolation`이라는 용어도 주의해야 한다. SQL 표준 isolation level 이름과 각 제품의 snapshot isolation 구현은 일대일로 맞지 않는다. PostgreSQL의 Repeatable Read가 표준에서 말하는 phantom 허용 여부와 어떻게 다른지, InnoDB Repeatable Read가 next-key lock으로 phantom을 어떻게 다루는지 같은 차이가 꼬리 질문으로 자주 나온다.

## 직접 재생해 보기

1. 두 session에서 snapshot 생성 시점을 비교한다.

    Session A:

    ```sql
    BEGIN;
    SELECT balance FROM account WHERE id = 'A';
    -- 여기서 Session B가 update/commit할 때까지 기다린다.
    SELECT balance FROM account WHERE id = 'A';
    COMMIT;
    ```

    Session B:

    ```sql
    BEGIN;
    UPDATE account SET balance = 900 WHERE id = 'A';
    COMMIT;
    ```

    `READ COMMITTED`에서는 Session A의 두 번째 SELECT가 새 값을 볼 수 있다. Transaction snapshot을 유지하는 격리에서는 같은 값을 계속 볼 수 있다. PASS 신호는 isolation level을 바꿨을 때 결과가 왜 달라지는지 snapshot 생성 시점으로 설명하는 것이다.

2. PostgreSQL에서 long transaction과 vacuum 영향을 관찰한다.

    한 session에서 `BEGIN` 후 SELECT를 실행하고 transaction을 닫지 않는다. 다른 session에서 같은 table을 대량 update/delete한다. `pg_stat_activity`의 transaction age와 table의 dead tuple 지표를 확인한다. PASS 신호는 오래 열린 transaction이 old tuple 정리를 늦출 수 있음을 설명하는 것이다. FAIL 신호는 vacuum 지연을 단순히 vacuum 설정 문제로만 보고 snapshot 보존 경계를 놓치는 것이다.

3. InnoDB에서 consistent read와 locking read를 비교한다.

    Session A가 transaction을 열고 plain SELECT를 반복한다. Session B가 같은 row를 update/commit한다. 그 다음 Session A에서 plain SELECT와 `SELECT ... FOR UPDATE`를 각각 실행해 결과와 wait behavior를 비교한다. PASS 신호는 consistent read는 read view를 기준으로 old version을 볼 수 있지만 locking read는 current row와 lock conflict를 다룬다는 점을 분리하는 것이다.

4. Writer-writer conflict를 만든다.

    두 session이 같은 row를 update하도록 한다. 하나가 commit하거나 rollback하기 전까지 다른 쪽이 대기하는지, timeout이나 deadlock 상황이 어떻게 나타나는지 본다. PASS 신호는 MVCC가 reader-writer blocking을 줄여도 writer-writer conflict는 남는다고 설명하는 것이다.

5. Cleanup 지표를 문장으로 바꾼다.

    PostgreSQL에서는 dead tuple, autovacuum, transaction age를 본다. InnoDB에서는 history list length, purge lag, undo tablespace 사용량을 본다. 숫자를 본 뒤에는 `old version을 아직 필요로 하는 snapshot이 있어서 정리가 늦다` 또는 `정리 속도보다 update/delete 생성 속도가 빠르다`처럼 원인 문장으로 바꿔야 한다.

## 면접 꼬리 질문

- MVCC는 lock을 쓰지 않는다는 뜻인가요?

    아니다. MVCC는 reader가 snapshot 기준으로 old/new version을 고를 수 있게 해 reader-writer blocking을 줄인다. 하지만 writer-writer conflict, unique constraint, locking read, serializable conflict에서는 lock이나 abort/retry가 필요하다.

- Snapshot은 실제 데이터 복사본인가요?

    보통 아니다. Snapshot은 보이는 transaction을 판정하는 기준이다. PostgreSQL은 snapshot과 tuple transaction metadata를 비교하고, InnoDB는 read view와 undo chain으로 과거 값을 재구성할 수 있다.

- PostgreSQL MVCC와 InnoDB MVCC의 가장 큰 차이는 무엇인가요?

    PostgreSQL은 heap에 tuple version이 남고 vacuum이 dead tuple을 정리한다는 설명이 중심이다. InnoDB는 clustered record와 undo log, read view, purge가 중심이다. 공통 원리는 visible version 선택이지만 저장 위치와 cleanup 방식이 다르다.

- 오래 열린 transaction이 왜 문제인가요?

    오래 열린 snapshot은 old version이 아직 필요할 수 있음을 의미한다. PostgreSQL에서는 vacuum이 dead tuple을 지우지 못하고, InnoDB에서는 purge가 undo history를 정리하지 못할 수 있다. 결과는 table bloat, undo 증가, 성능 저하다.

- READ COMMITTED와 REPEATABLE READ 차이는 MVCC에서 어떻게 보이나요?

    핵심은 snapshot 생성 시점이다. READ COMMITTED는 statement마다 새 snapshot을 보는 경우가 많고, REPEATABLE READ는 transaction 동안 같은 snapshot을 유지하는 경우가 많다. 다만 DBMS별 구현 차이가 있으므로 제품 이름 없이 일반화하면 안 된다. 제품별 anomaly와 lock 차이는 [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)에서 이어서 확인해야 한다.

- Serializable에서 실패가 나는 것은 DB 오류인가요?

    아니다. Serializable 계열에서는 충돌을 감지해 transaction을 abort시키는 것이 correctness mechanism일 수 있다. 애플리케이션은 serialization failure를 재시도 가능한 결과로 다뤄야 한다.

## 함정 질문

- "MVCC라서 SELECT는 절대 lock과 관련 없습니다"라고 말하는 것

    Plain consistent read와 locking read를 섞은 답이다. `SELECT ... FOR UPDATE`는 lock을 건다. Serializable이나 predicate conflict도 읽기와 쓰기의 관계를 추적할 수 있다.

- "Snapshot은 테이블 복사본입니다"라고 말하는 것

    비유로는 시작점이 될 수 있지만 내부 설명으로는 틀리기 쉽다. 실제 핵심은 transaction visibility metadata와 old version 접근 경로다.

- "PostgreSQL과 InnoDB는 둘 다 MVCC니까 old version 저장 방식도 같습니다"라고 말하는 것

    공통 원리와 구현을 섞은 답이다. PostgreSQL은 tuple/vacuum, InnoDB는 undo/read view/purge로 분리해야 한다.

- "MVCC를 쓰면 deadlock이 없습니다"라고 말하는 것

    Writer-writer conflict, locking read, gap/next-key lock, foreign key check, index update에서 deadlock은 여전히 가능하다. Deadlock과 retry는 isolation-lock 문서의 핵심이다.

- "Vacuum이나 purge는 성능 튜닝 옵션일 뿐입니다"라고 말하는 것

    Cleanup은 snapshot 계약을 지키면서 old version을 제거하는 작업이다. 너무 빠르게 지우면 correctness가 깨지고, 너무 늦게 지우면 성능과 공간이 무너진다.

### 다른 시스템으로 옮겨 보기: row version 중 어느 것이 현재 읽기에 보이는가

다른 엔진을 볼 때도 먼저 old version이 어디에 저장되는지, snapshot 기준이 언제 만들어지는지, cleanup이 무엇을 기다리는지 찾습니다. 이 세 가지를 찾으면 제품 이름이 달라도 MVCC 설명을 새로 세울 수 있습니다. 반대로 셋 중 하나를 모르면 `lock-free` 같은 얕은 문장으로 돌아가기 쉽습니다.

## 답변을 더 단단하게 만드는 판단 흐름

MVCC 답변을 더 깊게 만들려면 `읽기가 막히지 않는다`는 결과보다 그 결과를 가능하게 하는 판정 순서를 말해야 한다. 읽기 요청이 들어오면 DBMS는 먼저 그 요청이 어떤 snapshot을 사용할지 정한다. 그 다음 row의 현재 version을 보고, 그 version을 만든 transaction이 snapshot 기준에서 보이는지 판단한다. 보이지 않으면 PostgreSQL은 다른 tuple version을 찾고, InnoDB는 undo chain을 따라 과거 값을 재구성할 수 있다. 마지막으로 그 version이 아직 삭제되거나 대체되지 않은 것으로 보여야 reader에게 반환된다.

```text
read starts -> snapshot/read view 결정 -> current version 확인 -> visible 아니면 old version 탐색 -> visible version 반환
```

이 흐름에서 `snapshot은 언제 만들어지는가`가 핵심 분기다. READ COMMITTED에서는 statement마다 새 기준을 만들 수 있으므로 같은 transaction 안에서도 두 SELECT 결과가 달라질 수 있다. REPEATABLE READ 계열에서는 transaction 동안 같은 기준을 유지하므로 반복 조회가 같은 값을 볼 수 있다. 하지만 이 말을 표준 isolation 이름만으로 끝내면 위험하다. PostgreSQL의 Repeatable Read와 InnoDB의 Repeatable Read는 phantom 처리와 lock 전략이 다를 수 있다. 면접에서는 `이 DBMS에서는 snapshot 생성 시점과 range conflict 처리가 어떻게 되는가`라고 좁혀야 한다.

두 번째 분기는 `읽기인가, 앞으로 쓸 읽기인가`다. 단순 조회는 consistent read로 old version을 볼 수 있지만, `SELECT ... FOR UPDATE`는 곧 갱신할 row를 보호해야 하므로 current version과 lock manager에 더 가까워진다. 결제 승인 전 잔액 조회처럼 읽은 뒤 바로 갱신하는 흐름에서는 이 차이가 중요하다. 단순 SELECT에서 보인 값이 곧바로 안전한 update 전제가 되지는 않는다. 안전한 갱신은 조건부 update, locking read, optimistic version check, retry 정책 중 하나로 닫아야 한다.

세 번째 분기는 cleanup이다. Old version은 reader를 위해 남지만, 영원히 남으면 저장 공간과 조회 비용을 키운다. PostgreSQL에서 오래 열린 transaction은 vacuum이 dead tuple을 치우지 못하게 할 수 있고, InnoDB에서는 오래된 read view가 purge를 늦출 수 있다. 그래서 MVCC 운영 문제는 단순히 `쿼리가 느리다`가 아니라 `누가 과거를 붙잡고 있는가`라는 질문으로 바뀐다.

```text
long transaction open -> old versions still possibly visible -> cleanup cannot remove them -> bloat or undo history grows
```

이 모델은 실무 사고에도 바로 이어진다. 배치 리포트가 큰 table을 읽으면서 transaction을 오래 열어 두면 writer가 직접 막히지 않더라도 old version 정리 비용이 쌓인다. Connection pool에서 요청이 끝났는데 transaction이 닫히지 않은 `idle in transaction` 상태가 남으면 더 위험하다. 이 상태는 사용자가 보는 기능 오류보다 늦게 드러나지만, 시간이 지나면 vacuum 지연, undo 증가, disk 사용량 증가, plan 악화로 번진다.

면접에서 PostgreSQL과 InnoDB를 비교할 때는 다음 표처럼 같은 질문을 두 제품에 던지면 안전하다.

```text
질문                         PostgreSQL                         InnoDB
old version은 어디에 있는가    heap tuple version                 undo log로 재구성
reader 기준은 무엇인가         snapshot + tuple xmin/xmax          read view + trx id
정리는 누가 하는가             vacuum/autovacuum                  purge
운영 신호는 무엇인가           dead tuples, vacuum delay           history list, purge lag
locking read는 어디서 갈라지나  row/table lock, SSI retry           record/gap/next-key lock
```

이 표의 목적은 제품을 외우는 것이 아니다. 공통 원리를 같은 질문으로 나누어 보면 어느 부분은 같은 추상화이고 어느 부분은 제품별 구현인지 보인다. `MVCC는 여러 version을 쓴다`는 작은 문장은 맞지만, old version의 위치와 cleanup 방식을 합쳐 버리면 전체 설명은 틀린다. 그래서 작은 참을 큰 참으로 만들려면 항상 `저장 위치`, `판정 기준`, `정리 주체`, `관측 지표`를 함께 묶어야 한다.

마지막으로 MVCC와 성능의 관계도 신중히 말해야 한다. MVCC는 read/write blocking을 줄여 처리량을 높일 수 있지만, update-heavy workload에서는 old version 생성과 cleanup 비용이 커진다. Hot row에 writer가 몰리면 MVCC가 있어도 lock wait와 retry가 생긴다. Serializable에서 abort가 늘면 application retry가 병목이 될 수 있다. 따라서 MVCC는 `항상 빠르게 해 주는 기능`이 아니라 `일관된 읽기를 제공하면서 동시성을 높이는 대신 version 관리 비용을 지불하는 방식`이라고 말하는 편이 정확하다.

### 작은 장애 시나리오로 검산하기

MVCC를 제대로 이해했는지는 장애 시나리오로 검산하면 금방 드러난다. 첫 번째 시나리오는 오래 열린 리포트 transaction이다. 리포트가 아침 9시에 snapshot을 만들고 30분 동안 큰 table을 읽는 동안, OLTP 요청은 같은 table을 계속 update/delete한다. 사용자는 writer가 막히지 않으니 문제가 없다고 생각할 수 있지만, old version은 그 리포트 snapshot 때문에 아직 필요할 수 있다. 9시 snapshot이 닫히기 전까지 cleanup은 보수적으로 움직이고, 그 사이 dead tuple이나 undo history가 쌓인다.

두 번째 시나리오는 hot row update다. 같은 계정 잔액 row에 수백 요청이 몰리면 MVCC가 있어도 모든 writer가 자유롭게 진행하지 못한다. 각 writer는 current version을 기준으로 충돌을 조정해야 하고, lock wait나 retry가 생긴다. 이때 해결책은 MVCC 설정을 찾는 것이 아니라 workload 모양을 바꾸는 것이다. 예를 들어 ledger append 후 비동기 balance 집계, key별 queue 직렬화, optimistic version check, 한 계정 단위의 command serialization 같은 설계가 후보가 된다.

세 번째 시나리오는 unique key 충돌이다. 두 transaction이 같은 idempotency key로 payment row를 넣으면 둘 다 과거 snapshot에서는 key가 없어 보였을 수 있다. 하지만 unique index는 최종적으로 같은 key 두 개를 허용하지 않는다. 이 지점에서 MVCC visibility와 constraint enforcement가 만난다. 좋은 답변은 `둘 다 못 봤는데 왜 하나가 실패하나요?`라는 질문에 `snapshot read와 unique constraint의 현재성 판단은 같은 문제가 아니다`라고 설명할 수 있어야 한다.

이 세 시나리오는 MVCC를 성능 기능으로만 설명하는 답을 막아 준다. MVCC는 읽기 일관성을 제공하지만, cleanup과 write conflict, constraint enforcement는 계속 시스템의 비용과 실패 모드로 남는다. 따라서 면접에서 MVCC를 말한 뒤에는 반드시 `어떤 지표로 cleanup 지연을 볼 것인가`, `어떤 작업에서 writer conflict가 남는가`, `어떤 에러를 retry 가능한 concurrency 결과로 볼 것인가`를 붙여야 한다.

## 더 깊게 볼 자료

공식 자료:

- [PostgreSQL current: Multiversion Concurrency Control](https://www.postgresql.org/docs/current/mvcc.html)
- [PostgreSQL current: Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL current: Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html)
- [MySQL 8.4 Reference Manual: InnoDB Multi-Versioning](https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html)
- [MySQL 8.4 Reference Manual: Consistent Nonlocking Reads](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html)
- [MySQL 8.4 Reference Manual: InnoDB Transaction Isolation Levels](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html)

저장소 안에서 이어 볼 자료:

- [database/mvcc.md](../../database/mvcc.md)
- [database/deep-dive/transactions/12-mvcc-snapshot-visibility.md](../../database/deep-dive/transactions/12-mvcc-snapshot-visibility.md)
- [database/lock.md](../../database/lock.md)
- [database/postgresql/lock.md](../../database/postgresql/lock.md)
- [트랜잭션 경계와 ACID](06-transaction-acid-boundary.md)
- [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)

이 자료를 읽을 때는 `어디에 old version이 있는가`, `snapshot은 어떤 transaction을 visible로 보는가`, `cleanup은 무엇을 기다리는가`를 계속 묻는다. 이 세 질문을 잡으면 MVCC 답변은 lock-free 구호가 아니라 버전 가시성 계약으로 닫힌다.
