# PostgreSQL Deep Dive

이 문서는 PostgreSQL을 heap tuple, MVCC, vacuum, planner, lock, WAL, replication 관측 흐름으로 읽기 위한 장문 학습 문서다. 각 절은 실제 장애와 실험으로 다시 재생할 수 있게 구성한다.

## PostgreSQL heap, tuple, xmin/xmax, vacuum

PostgreSQL MVCC를 이해할 때 가장 중요한 첫 문장은 “UPDATE는 기존 row를 제자리에서 고치는 것이 아니라, 새 tuple version을 만들고 이전 version을 언젠가 정리할 대상으로 남긴다”이다. 이 문장을 붙잡으면 heap, tuple, `xmin`, `xmax`, vacuum, visibility map, free space map이 하나의 흐름으로 연결된다. 반대로 vacuum을 “디스크 청소”라고만 이해하면, 왜 오래 열린 transaction 하나가 table bloat를 만들고, 왜 `VACUUM`이 planner statistics와 index-only scan 성능에도 영향을 주고, 왜 transaction ID wraparound가 안전 문제인지 설명하지 못한다. PostgreSQL 공식 문서는 concurrency control 장에서 여러 session이 같은 data에 접근할 때 효율적 접근과 엄격한 data integrity를 함께 달성하는 것이 목표라고 설명한다. Routine Vacuuming 문서는 PostgreSQL database가 주기적인 vacuuming을 필요로 하며, 그 이유가 updated/deleted row가 차지한 공간 회수 또는 재사용, planner statistics 갱신, visibility map 갱신, transaction ID wraparound 방지라고 정리한다. 이 절은 그 네 가지 이유가 heap tuple versioning에서 어떻게 나오는지 아래에서부터 쌓는다.

### 첫 번째 벽돌: heap page와 tuple version

PostgreSQL의 기본 table storage는 heap이다. 여기서 heap은 Java heap처럼 객체가 있는 memory 영역이 아니라, table row version이 특정 logical order 없이 page 안에 저장되는 relation storage를 뜻한다. PostgreSQL storage page layout 문서는 table과 index가 page 배열로 저장되고, table에서는 item이 row이며 index에서는 item이 index entry라고 설명한다. Table row는 대부분의 machine에서 23-byte fixed header를 갖고, 그 뒤에 null bitmap과 user data가 이어진다. 이 header 안에는 MVCC 판단에 필요한 transaction metadata가 들어 있다. 학습용으로는 `xmin`, `xmax`, `ctid` 세 가지를 먼저 붙잡으면 된다.

| 항목 | 쉬운 뜻 | 실무에서 묻는 질문 |
|---|---|---|
| `xmin` | 이 tuple version을 만든 transaction ID | 이 row version은 언제부터 보일 수 있는가 |
| `xmax` | 이 tuple version을 delete하거나 update로 대체한 transaction ID | 이 row version은 언제부터 더 이상 보이면 안 되는가 |
| `ctid` | 현재 tuple version의 physical 위치 `(block, offset)` | UPDATE 뒤 같은 logical row가 page 안 어디로 갔는가 |

`ctid`는 안정적인 business key가 아니다. UPDATE가 새 tuple version을 만들면 `ctid`가 바뀔 수 있다. 하지만 학습과 관측에는 매우 좋은 첫 벽돌이다. “같은 logical row인데 physical tuple version이 바뀐다”는 사실을 눈으로 보여 주기 때문이다.

```sql
CREATE TABLE pg_mvcc_lab (
  id bigint PRIMARY KEY,
  status text NOT NULL,
  note text
);

INSERT INTO pg_mvcc_lab VALUES (1, 'READY', 'first version');

SELECT ctid, xmin, xmax, id, status, note
FROM pg_mvcc_lab
WHERE id = 1;
```

처음 관측은 대략 다음 모양이다. 실제 `xmin` 숫자는 환경마다 다르므로 숫자 자체를 외우면 안 된다.

```text
ctid   xmin  xmax  id  status  note
-----  ----  ----  --  ------  -------------
(0,1)  740   0     1   READY   first version
```

이제 같은 row를 update한다.

```sql
UPDATE pg_mvcc_lab
SET status = 'PAID', note = 'second version'
WHERE id = 1;

SELECT ctid, xmin, xmax, id, status, note
FROM pg_mvcc_lab
WHERE id = 1;
```

보이는 결과는 새 version 하나뿐일 수 있다.

```text
ctid   xmin  xmax  id  status  note
-----  ----  ----  --  ------  --------------
(0,2)  741   0     1   PAID    second version
```

하지만 여기서 중요한 것은 `(0,1)`이 곧바로 물리적으로 사라졌다고 단정하지 않는 것이다. PostgreSQL은 transaction visibility 규칙 때문에 이전 tuple version을 일정 시간 남겨야 한다. 다른 transaction이 UPDATE 전에 잡은 snapshot으로 읽고 있다면, 그 transaction에는 `(0,1)`이 여전히 맞는 version일 수 있다. 따라서 UPDATE는 “old tuple을 지우고 new value를 덮어씀”이 아니라 “old tuple을 더 이상 최신이 아니게 표시하고 new tuple을 추가함”에 가깝다.

### tuple version trace: UPDATE가 만든 연결

다음 trace는 실제 내부 field를 모두 표현하지 않은 학습용 단순화다. 그래도 `xmin`, `xmax`, `ctid`가 어떤 판단에 쓰이는지는 충분히 보여 준다.

```text
초기 상태
  heap page 0
  offset 1: tuple(id=1, status='READY', xmin=740, xmax=0, ctid=(0,1))

UPDATE id=1 SET status='PAID' by transaction 741
  old version:
    offset 1: tuple(id=1, status='READY', xmin=740, xmax=741, ctid=(0,2))
  new version:
    offset 2: tuple(id=1, status='PAID', xmin=741, xmax=0, ctid=(0,2))

reader snapshot:
  snapshot sees transaction 740 committed
  snapshot does not see transaction 741 yet
    -> offset 1 is visible
    -> offset 2 is invisible

new reader after 741 commit:
  snapshot sees transaction 741 committed
    -> offset 1 is dead for this snapshot
    -> offset 2 is visible
```

이 trace에서 `xmax=741`은 “old tuple을 transaction 741이 끝낸다”는 표식이다. `xmax`가 있다는 사실만으로 항상 삭제 완료라고 읽으면 안 된다. Transaction 741이 abort되면 다른 판단이 필요하고, row lock과 multixact가 얽히면 더 복잡한 의미가 생긴다. 그러나 첫 이해에서는 `xmin`이 시작점, `xmax`가 끝점 후보, snapshot이 둘 사이에서 visible version을 고른다고 보면 된다.

### PostgreSQL MVCC는 InnoDB undo-chain 모델과 다르다

로컬 seed `database/mvcc.md`는 MVCC 구현 방식의 두 흐름을 구분하려고 한다. 그 방향은 중요하다. InnoDB는 현재 clustered index record와 undo record chain을 통해 과거 version을 재구성하는 쪽에 가깝다. PostgreSQL은 heap에 tuple version을 남기고, 각 tuple header의 transaction metadata로 snapshot visibility를 판단하는 쪽에 가깝다. 둘 다 “읽기와 쓰기가 서로를 덜 막게 한다”는 목표는 공유하지만, 공간 회수와 관측 방식은 달라진다.

| 비교 축 | PostgreSQL heap MVCC | InnoDB undo 기반 MVCC |
|---|---|---|
| 최신 row 위치 | heap tuple version 중 visible한 최신 version | clustered index record |
| 과거 version | heap 안 old tuple version | undo log record chain |
| 정리 작업 | vacuum이 dead tuple과 visibility metadata를 정리 | purge가 undo/history를 정리 |
| 관측 감각 | `xmin`, `xmax`, `ctid`, `n_dead_tup`, vacuum stats | history list, undo tablespace, InnoDB status |

이 차이를 모르고 “MVCC니까 다 비슷하다”고 생각하면 PostgreSQL vacuum을 InnoDB purge와 같은 이름의 청소 작업 정도로 낮춰 보게 된다. PostgreSQL에서는 dead tuple이 heap과 index scan 비용, table bloat, planner statistics, visibility map까지 영향을 준다. Vacuum은 저장공간 문제만이 아니라 query plan과 index-only scan 가능성, transaction ID wraparound 안전성을 함께 관리한다.

### Vacuum은 왜 네 가지 일을 하는가

공식 Routine Vacuuming 문서는 `VACUUM`이 table을 정기적으로 처리해야 하는 이유를 네 가지로 든다. 첫째, updated/deleted row가 차지한 disk space를 recover하거나 reuse한다. 둘째, query planner가 쓰는 data statistics를 update한다. 셋째, index-only scan을 빠르게 하는 visibility map을 update한다. 넷째, transaction ID wraparound나 multixact ID wraparound 때문에 아주 오래된 data를 잃는 상황을 막는다. 이 네 가지는 따로 붙은 기능 목록이 아니라 heap MVCC의 자연스러운 결과다.

UPDATE가 old tuple을 남기면 어느 순간에는 “이 old tuple을 볼 수 있는 active snapshot이 더 이상 없다”는 판단이 가능해진다. 그때 vacuum은 dead tuple을 제거하거나 그 공간을 재사용 가능하게 만든다. 이 작업이 없으면 table file 안에 더 이상 어떤 transaction에도 보이지 않는 tuple이 쌓인다. 이것이 흔히 말하는 table bloat의 한 원인이다. Bloat는 단순히 disk를 더 쓰는 문제가 아니다. 같은 logical row 수를 읽기 위해 더 많은 page를 훑어야 하고, cache 효율이 떨어지고, sequential scan cost와 random page access가 늘 수 있다.

Statistics update도 같은 흐름에서 나온다. Planner는 table row 수, page 수, value distribution 같은 추정치를 기반으로 plan을 고른다. `pg_class` 공식 문서는 `relpages`가 on-disk representation의 page 수 추정치이고 planner가 쓰며, `VACUUM`, `ANALYZE`, 일부 DDL이 갱신한다고 설명한다. `reltuples`도 live row 수 추정치이고 planner가 쓴다. Table에 update/delete가 많았는데 statistics가 낡으면 planner는 실제보다 table이 작다고 생각하거나, 특정 조건의 선택도를 잘못 보고 seq scan과 index scan 사이에서 틀린 선택을 할 수 있다.

Visibility map은 “이 page의 tuple들이 모든 active transaction에 visible하다” 또는 “모두 frozen이다” 같은 정보를 page 단위 bit로 저장한다. 공식 storage VM 문서는 각 heap relation이 `_vm` suffix의 별도 relation fork에 visibility map을 갖고, 첫 번째 bit가 all-visible page를 나타내며 index-only scan에 쓰일 수 있다고 설명한다. 또한 visibility map bit는 conservative하다. set되어 있으면 조건이 참임을 보장하지만, set되어 있지 않다고 반드시 거짓인 것은 아니다. Vacuum은 이 bit를 set할 수 있고, page의 data-modifying operation은 bit를 clear한다. 그래서 vacuum은 “죽은 row를 치우는 작업”을 넘어 “index-only scan이 heap fetch를 생략할 수 있는 근거를 만드는 작업”이기도 하다.

Transaction ID wraparound는 더 낮은 층의 안전 문제다. PostgreSQL transaction ID는 무한히 큰 정수가 아니라 wraparound 가능성이 있는 식별자다. 아주 오래된 `xmin`이 freeze되지 않으면 미래의 transaction ID와 비교할 때 과거/미래 판단이 망가질 수 있다. 그래서 vacuum은 오래된 tuple의 transaction ID를 freeze해서 “이 tuple은 충분히 오래되어 모든 정상 snapshot에 보인다”고 표시하는 안전 장치도 수행한다. 이 부분을 모르고 autovacuum을 무작정 꺼 버리면, 단순 성능 튜닝이 아니라 data availability를 위협하는 선택이 된다.

### Free space map과 visibility map을 같이 읽기

PostgreSQL physical storage에는 main fork만 있는 것이 아니다. Free Space Map, 즉 FSM은 relation 안 page별 여유 공간을 추적한다. 공식 storage FSM 문서는 heap과 대부분의 index relation이 `_fsm` suffix의 별도 relation fork를 갖고, bottom level FSM page가 각 heap 또는 index page의 free space를 one byte로 표현하며, upper level이 lower level 정보를 aggregate한다고 설명한다. Visibility Map, 즉 VM은 앞에서 본 것처럼 all-visible과 all-frozen 정보를 page당 두 bit로 갖는다.

학습용으로 한 table의 상태를 다음처럼 그릴 수 있다.

```text
relation app.order_event

main fork:
  page 0: live tuple 많음, dead tuple 일부, free space 적음
  page 1: all visible, free space 보통
  page 2: update/delete 후 dead tuple 많음

order_event_fsm:
  page 0 -> free space 12%
  page 1 -> free space 35%
  page 2 -> free space 70%

order_event_vm:
  page 0 -> all-visible=0, all-frozen=0
  page 1 -> all-visible=1, all-frozen=0
  page 2 -> all-visible=0, all-frozen=0

planner / executor:
  index-only scan can skip heap check for page 1
  insert may choose page 2 if tuple fits and FSM says enough free space
  vacuum can remove dead tuples and may set VM bits if page becomes all-visible
```

이 그림에서 FSM과 VM은 서로 다른 질문에 답한다. FSM은 “새 tuple을 넣을 공간이 어디에 있는가”에 가깝고, VM은 “이 page를 heap visibility check 없이 믿어도 되는가”에 가깝다. 둘 다 vacuum과 update/delete workload의 영향을 받지만, 같은 지도는 아니다. 이 차이를 모르면 index-only scan이 왜 heap fetch를 많이 하는지, insert가 왜 table 끝에만 붙지 않는지, vacuum이 왜 statistics와 visibility에 동시에 등장하는지 헷갈린다.

### Autovacuum을 꺼야 할 때보다 조정해야 할 때가 많다

운영에서 autovacuum 때문에 I/O가 튄다는 불만은 실제로 자주 있다. 하지만 해법이 곧바로 autovacuum off는 아니다. Autovacuum은 dead tuple 정리, stats freshness, visibility map, wraparound 방지를 맡는다. 꺼 버리면 지금 보이는 I/O는 줄 수 있지만, 나중에 더 큰 vacuum, table bloat, plan regression, wraparound emergency로 돌아올 수 있다. 시니어 운영자는 autovacuum을 “귀찮은 background job”이 아니라 “MVCC의 비용을 뒤로 미루지 않게 하는 amortization 장치”로 본다. 여기서 amortization은 한꺼번에 큰 비용을 내지 않도록 작업을 조금씩 나눠 치르는 방식이라고 풀어 이해하면 된다.

조정은 table별 workload를 봐야 한다. 작은 lookup table과 hot update table은 같은 threshold가 맞지 않는다. Update/delete가 많은 table에서는 `autovacuum_vacuum_scale_factor`를 낮춰 더 자주 청소하게 할 수 있고, insert-heavy table에서는 insert threshold도 고려한다. 단, 이 절은 구체 parameter guide가 아니라 heap tuple과 vacuum의 원리를 설명하는 DU이므로, 핵심은 “왜 조정해야 하는가”다. Dead tuple이 쌓이는 속도, vacuum이 처리하는 속도, long transaction이 oldest xmin을 붙잡는지, visibility map이 다시 set되는지, planner statistics가 stale한지를 같이 봐야 한다.

### 관측 query: tuple과 vacuum을 숫자로 확인하기

다음 query들은 PostgreSQL 환경에서 heap MVCC와 vacuum 상태를 좁혀 보는 출발점이다.

```sql
-- tuple version을 직접 보는 작은 실험
SELECT ctid, xmin, xmax, id, status, note
FROM pg_mvcc_lab
ORDER BY id;

-- table별 live/dead tuple 추정과 vacuum/analyze 시점
SELECT
  schemaname,
  relname,
  n_live_tup,
  n_dead_tup,
  last_vacuum,
  last_autovacuum,
  last_analyze,
  last_autoanalyze
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY n_dead_tup DESC;

-- planner가 보는 page/row 추정치
SELECT
  relname,
  relpages,
  reltuples,
  relallvisible,
  relallfrozen
FROM pg_class
WHERE relname = 'pg_mvcc_lab';
```

PASS 신호는 workload 설명과 지표가 맞는 것이다. 예를 들어 update/delete가 많은 table의 `n_dead_tup`이 빠르게 늘고, long transaction이 오래 열려 있으며, autovacuum이 반복적으로 제대로 끝나지 못한다면 bloat와 snapshot retention을 의심할 수 있다. `relallvisible`이 낮고 index-only scan에서 heap fetch가 많다면 visibility map 갱신과 page churn을 함께 봐야 한다. FAIL 신호는 `n_dead_tup` 하나만 보고 vacuum이 문제라고 단정하거나, `VACUUM FULL`을 즉시 실행하는 것이다. `VACUUM FULL`은 table rewrite와 강한 lock을 동반할 수 있으므로 일반 vacuum과 운영 영향이 다르다. 공간 회수 목적이라면 lock, disk free, replication, maintenance window를 따로 닫아야 한다.

### 장기 transaction과 vacuum 방해 trace

Vacuum 문제의 대표적인 실전 함정은 “아무 query도 느리지 않은데 dead tuple이 안 줄어든다”는 상황이다. 원인은 오래 열린 transaction일 수 있다.

```text
T0  Session A starts transaction and reads table
    snapshot xmin boundary = 800

T1  Session B updates 1,000,000 rows
    old tuple versions become dead for new snapshots
    but Session A may still need versions visible at T0

T2  autovacuum scans table
    sees many old tuple versions
    cannot remove versions that may be needed by Session A snapshot

T3  Session A remains idle in transaction for 3 hours
    n_dead_tup remains high
    table/index scans pay extra cost

T4  Session A commits
    next vacuum can remove more dead tuples
```

관측은 다음처럼 시작한다.

```sql
SELECT
  pid,
  usename,
  state,
  xact_start,
  now() - xact_start AS xact_age,
  wait_event_type,
  wait_event,
  query
FROM pg_stat_activity
WHERE xact_start IS NOT NULL
ORDER BY xact_start;
```

이 query의 PASS 신호는 vacuum이 왜 못 치우는지 “oldest transaction” 후보와 연결하는 것이다. FAIL 신호는 dead tuple이 많다는 이유만으로 autovacuum worker 수나 cost limit만 올리는 것이다. Worker와 cost를 올려도 오래된 snapshot이 필요로 하는 tuple은 지울 수 없다. 먼저 idle in transaction, long report transaction, logical replication slot, prepared transaction 같은 retention source를 봐야 한다.

### 흔한 오해를 고치기

첫 번째 오해는 “PostgreSQL row는 UPDATE하면 그 자리에서 값만 바뀐다”는 것이다. 실제로는 새 tuple version이 만들어지고, old tuple은 visibility 규칙에 따라 한동안 남는다. 이 차이를 모르면 `ctid` 변화, bloat, vacuum 필요성을 설명할 수 없다.

두 번째 오해는 “vacuum은 디스크 청소다”라는 것이다. Vacuum은 dead tuple 공간 회수 또는 재사용뿐 아니라 planner statistics, visibility map, transaction ID wraparound 방지에 연결된다. 디스크 사용량만 보고 vacuum의 성공/실패를 판단하면 index-only scan과 plan quality, wraparound safety를 놓친다.

세 번째 오해는 “autovacuum이 느리면 끄면 된다”는 것이다. Autovacuum은 비용을 없애는 기능이 아니라 MVCC 비용을 제때 치르게 하는 기능이다. 끄면 당장은 조용해질 수 있지만, 나중에 더 큰 비용으로 돌아온다. 조정할 때도 table workload와 long transaction, replication slot, freeze age를 함께 봐야 한다.

네 번째 오해는 “`xmin`과 `xmax`만 보면 모든 visibility를 직접 판정할 수 있다”는 것이다. 학습에는 좋지만 실제 PostgreSQL visibility는 commit log 상태, snapshot, hint bit, multixact, row lock, isolation level이 얽힌다. 따라서 `xmin`/`xmax`는 first brick이지 완전한 user-space visibility evaluator가 아니다.

### 직접 replay할 작은 실험

로컬 PostgreSQL이 있다면 두 session으로 다음을 해 볼 수 있다.

```sql
-- session A
BEGIN;
SELECT ctid, xmin, xmax, id, status FROM pg_mvcc_lab WHERE id = 1;

-- session B
UPDATE pg_mvcc_lab SET status = 'CANCELLED' WHERE id = 1;
COMMIT;

-- session A
SELECT ctid, xmin, xmax, id, status FROM pg_mvcc_lab WHERE id = 1;
COMMIT;

-- 새 session
SELECT ctid, xmin, xmax, id, status FROM pg_mvcc_lab WHERE id = 1;
VACUUM (VERBOSE, ANALYZE) pg_mvcc_lab;
```

READ COMMITTED와 REPEATABLE READ에서 session A의 두 번째 SELECT 결과가 어떻게 달라지는지 비교하면 snapshot boundary를 더 잘 이해할 수 있다. PASS 신호는 “같은 logical row의 visible version이 transaction snapshot에 따라 달라질 수 있다”는 점을 설명하는 것이다. FAIL 신호는 `ctid`가 바뀐 것을 보고 row identity가 바뀌었다고 business 의미까지 오해하는 것이다. Business identity는 `id`, physical tuple location은 `ctid`다.

### HOT update와 index bloat 감각

PostgreSQL heap MVCC를 조금 더 실무적으로 이해하려면 HOT update도 알아야 한다. HOT는 Heap-Only Tuple의 약자로, update가 index key column을 바꾸지 않고 같은 page 안에 새 tuple version을 둘 수 있을 때 index entry를 새로 만들지 않고 heap tuple chain으로 해결하는 최적화다. 이 설명은 PostgreSQL 내부 세부에 가까우므로 처음부터 외울 필요는 없지만, “UPDATE는 항상 모든 index를 다시 쓴다”와 “UPDATE는 index에 영향이 없다”라는 양쪽 오해를 동시에 막아 준다.

예를 들어 `orders(id primary key, status indexed, note not indexed)`가 있다고 하자. `note`만 바꾸는 update는 index key를 바꾸지 않으므로 HOT update 후보가 될 수 있다. 반대로 `status`를 바꾸면 `status` index entry가 달라져야 하므로 index 쪽 작업이 필요하다. 실제 HOT 가능 여부는 page 여유 공간과 table/index 상태에 따라 달라진다. 그래서 `fillfactor`를 낮춰 page에 update 여유 공간을 남기는 tuning이 hot update table에서 의미를 가질 수 있다. 다만 fillfactor를 낮추면 같은 row 수를 위해 더 많은 page를 쓰므로 read path와 cache 효율에는 비용이 생긴다.

```text
not indexed column update 후보:
  index entry: id=1 -> heap page 0 offset 1
  heap page:
    (0,1) old tuple note='a' xmax=900 -> points to (0,2)
    (0,2) new tuple note='b' xmin=900
  index does not need a new key entry if HOT conditions hold

indexed column update:
  status index old key ('READY', pointer) must stop being current
  status index new key ('PAID', pointer) must be reachable
```

이 감각이 왜 중요한가. Update-heavy table에서 indexed column을 자주 바꾸면 heap bloat뿐 아니라 index bloat와 WAL volume도 함께 커진다. “row 수는 그대로인데 table과 index가 커진다”는 현상을 단순히 디스크 낭비로만 보면 안 된다. Query는 더 많은 page를 읽고, vacuum은 더 많은 dead tuple과 index entry를 처리하고, planner는 stale하거나 부정확한 추정에 노출될 수 있다. Column을 index에 넣는 결정은 read query만 빠르게 하는 결정이 아니라 update cost를 바꾸는 결정이다.

### Freeze를 이해하지 못하면 wraparound 경고가 공포가 된다

Transaction ID wraparound는 PostgreSQL을 처음 운영할 때 특히 무섭게 보인다. 핵심은 transaction ID 비교가 원형 공간에서 이루어지므로 너무 오래된 tuple의 `xmin`을 그대로 두면 미래 transaction에서 visibility 판단이 뒤집힐 위험이 있다는 것이다. Vacuum freeze는 충분히 오래되어 모든 정상 transaction에 visible하다고 볼 수 있는 tuple을 frozen 상태로 표시해 이 위험을 줄인다. 따라서 freeze는 “오래된 데이터를 건드리는 이상한 청소”가 아니라 MVCC 시간표를 안전하게 접는 작업이다.

```text
단순 시간표:
  xid 100 creates tuple A
  xid 200, 300, 400 ... many transactions pass
  tuple A is still live and visible to everyone

without freezing:
  xid counter eventually wraps
  old/new comparison can become ambiguous

with freezing:
  vacuum marks tuple A as frozen
  future snapshots do not need to compare old xid 100 in the same way
```

Wraparound 관련 vacuum은 일반 성능 튜닝보다 안전 우선이다. Autovacuum이 aggressive하게 보이고 I/O를 쓰더라도, 그것이 wraparound 방지 목적이라면 임의로 중단하기 전에 database age, table age, replication slot, long transaction을 봐야 한다. 이 상황에서 “서비스 부하가 있으니 autovacuum을 끄자”는 결정은 매우 위험하다. 더 나은 방향은 오래된 transaction을 제거하고, table별 vacuum cost와 threshold를 조정하고, 필요하면 maintenance window에서 manual vacuum/freeze를 계획하는 것이다.

### Vacuum과 ANALYZE를 분리해서 생각하기

`VACUUM`, `ANALYZE`, `VACUUM ANALYZE`라는 명령 이름 때문에 둘이 하나로 뭉개지기 쉽다. Vacuum은 dead tuple 정리, visibility map, freeze 같은 storage/MVCC 쪽 일이 중심이고, Analyze는 planner statistics를 수집하는 일이 중심이다. 물론 autovacuum daemon은 autoanalyze도 수행하고, routine vacuuming 문서도 planner statistics 갱신을 vacuuming 이유 중 하나로 다룬다. 하지만 장애 분석에서는 둘을 분리해야 한다.

예를 들어 query plan이 갑자기 나빠졌는데 `n_dead_tup`은 낮고 `last_autoanalyze`가 오래되었다면 dead tuple 청소보다 statistics freshness가 더 직접 원인일 수 있다. 반대로 statistics는 최신인데 table size가 계속 커지고 index-only scan이 heap fetch를 많이 한다면 visibility map과 bloat를 봐야 한다. `VACUUM (ANALYZE)`를 실행하면 둘을 함께 건드리므로 증상이 좋아져도 원인을 둘 중 어디로 봐야 하는지 흐려질 수 있다. 학습 단계에서는 일부러 `ANALYZE`만 실행한 전후와 `VACUUM`만 실행한 전후를 나눠 보면 좋다.

```sql
-- planner statistics만 새로 수집
ANALYZE pg_mvcc_lab;

-- dead tuple 정리와 visibility/freeze 관련 작업
VACUUM (VERBOSE) pg_mvcc_lab;

-- 둘을 함께
VACUUM (VERBOSE, ANALYZE) pg_mvcc_lab;
```

PASS 신호는 어떤 명령이 어떤 증상을 바꿨는지 분리해서 설명하는 것이다. FAIL 신호는 `VACUUM ANALYZE` 한 번으로 좋아졌다는 사실만 보고 “청소가 됐다”고 말하는 것이다. 좋아진 이유가 statistics 갱신인지, dead tuple 정리인지, visibility map 갱신인지, cache warm-up인지 따로 확인해야 한다.

### Interview 관점의 압축 답변

PostgreSQL MVCC를 짧게 설명해야 한다면 이렇게 말할 수 있어야 한다. PostgreSQL은 heap에 tuple version을 남기고, 각 version의 `xmin`과 `xmax` 같은 transaction metadata와 snapshot을 비교해 어떤 version이 보이는지 결정한다. UPDATE는 기존 tuple을 제자리 수정하는 것이 아니라 새 tuple version을 만들기 때문에 dead tuple이 쌓일 수 있다. Vacuum은 이 dead tuple을 제거하거나 공간을 재사용 가능하게 만들 뿐 아니라 planner statistics, visibility map, transaction ID wraparound 방지도 담당한다. 그래서 vacuum은 단순 디스크 청소가 아니라 PostgreSQL MVCC가 장기적으로 건강하게 유지되기 위한 핵심 maintenance path다.

이 답변에서 꼭 따라와야 할 trap은 `ctid`와 business identity 구분이다. `ctid`는 physical tuple location이므로 update 후 바뀔 수 있다. Application key로 쓰면 안 된다. 학습과 디버깅에서는 유용하지만, durable identity는 primary key나 business key가 맡아야 한다.

### Index-only scan이 heap과 vacuum을 다시 만나게 하는 지점

PostgreSQL에서 index-only scan이라는 이름은 오해를 부른다. 이름만 보면 index만 읽고 끝나는 것 같지만, 실제로는 “index tuple만으로 필요한 column을 얻을 수 있고, visibility map이 해당 heap page의 tuple들이 모두 visible하다고 말해 줄 때 heap 방문을 생략할 수 있다”에 가깝다. Index에는 key와 포함 column이 있을 수 있지만, MVCC visibility의 최종 근거는 heap tuple의 transaction metadata다. Visibility map의 all-visible bit가 없다면 executor는 index entry가 가리키는 heap tuple을 확인해야 한다. 그래서 vacuum이 visibility map을 갱신하지 못하면 covering index를 만들어도 heap fetch가 많이 남을 수 있다.

```text
index-only scan 후보:
  query needs columns: (tenant_id, created_at, status)
  index contains:      (tenant_id, created_at) INCLUDE (status)

case A: visibility map says heap page 42 is all-visible
  index entry -> answer from index tuple
  heap page 42 does not need per-tuple visibility check

case B: visibility map bit is clear for heap page 42
  index entry -> must visit heap page 42
  tuple xmin/xmax/snapshot check still needed
```

이 trace는 vacuum을 planner와 executor 성능으로 연결한다. `VACUUM`이 dead tuple을 정리하고 visibility map bit를 set할 수 있으면 index-only scan은 heap fetch를 줄일 수 있다. 반대로 update/delete가 잦은 table은 data-modifying operation이 visibility map bit를 clear하므로, index-only scan이 기대만큼 좋아지지 않을 수 있다. 따라서 “index-only scan을 위해 INCLUDE index를 만들었다”는 설계는 vacuum 주기, update 빈도, visibility map 상태와 함께 검증해야 한다. `EXPLAIN (ANALYZE, BUFFERS)`에서 `Heap Fetches`가 많이 보이면 index 설계만 볼 것이 아니라 visibility map과 vacuum 상태도 확인해야 한다.

### Bloat 진단은 크기 비교보다 원인 분리가 먼저다

Table bloat라는 말도 조심해서 써야 한다. 모든 큰 table이 bloat된 것은 아니다. 많은 live row를 담기 위해 큰 것은 정상이고, dead tuple과 비효율적 page 사용 때문에 필요 이상으로 큰 것이 bloat다. 운영에서 bloat 의심이 생기면 먼저 workload를 분리한다. 최근 대량 update/delete가 있었는지, autovacuum이 제때 끝났는지, long transaction이나 replication slot이 old tuple 제거를 막는지, fillfactor와 HOT update 가능성이 workload에 맞는지, index key update가 많은지 확인한다.

```sql
SELECT
  relname,
  n_live_tup,
  n_dead_tup,
  vacuum_count,
  autovacuum_count,
  analyze_count,
  autoanalyze_count
FROM pg_stat_user_tables
WHERE relname = 'pg_mvcc_lab';
```

PASS 신호는 “dead tuple이 많다 -> vacuum을 더 세게 돌린다”로 바로 가지 않고, 왜 dead tuple이 쌓였는지 설명하는 것이다. Long transaction이 원인이면 vacuum을 세게 돌려도 제거 가능한 tuple이 제한된다. Statistics가 stale한 것이 원인이면 analyze가 더 직접적인 조치일 수 있다. Update-heavy indexed column이 원인이면 index 설계나 update pattern을 바꿔야 할 수 있다. Disk space를 OS에 즉시 돌려줘야 하는 상황이면 일반 vacuum만으로는 부족하고 rewrite 계열 작업을 검토해야 하지만, 그 작업은 lock과 disk 여유 공간을 따로 요구한다. 그래서 bloat 진단의 첫 목표는 “크다”를 증명하는 것이 아니라 “왜 커졌고, 어떤 조치가 어떤 비용으로 줄일 수 있는가”를 분리하는 것이다.

### 출처와 검증 경계

이 절의 공식 근거는 PostgreSQL 18 문서의 [Concurrency Control](https://www.postgresql.org/docs/current/mvcc.html), [Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html), [Database Page Layout](https://www.postgresql.org/docs/current/storage-page-layout.html), [Visibility Map](https://www.postgresql.org/docs/current/storage-vm.html), [Free Space Map](https://www.postgresql.org/docs/current/storage-fsm.html), [pg_class](https://www.postgresql.org/docs/current/catalog-pg-class.html)다. 로컬 seed는 `database/mvcc.md`와 `database/postgresql/*` 아래의 PostgreSQL 메모와 실험 기록이다. 이 절의 tuple trace는 teaching model이며, PostgreSQL 내부 visibility 함수와 tuple header bit를 완전히 재현한 포맷 설명은 아니다. 그러나 UPDATE가 새 tuple version을 만들고, snapshot이 `xmin`/`xmax`와 transaction 상태를 통해 visible tuple을 고르며, vacuum이 dead tuple 정리와 statistics, visibility map, wraparound safety를 함께 맡는다는 핵심 구조는 공식 문서와 local SQL 관측으로 replay할 수 있다.

## PostgreSQL planner, locks, WAL, replication

PostgreSQL 운영 진단에서 가장 위험한 습관은 관측 view 하나를 보고 원인을 단정하는 것이다. `pg_stat_activity`에 `Lock` wait가 보인다고 lock 자체가 root cause인 것은 아니고, `EXPLAIN`에서 seq scan이 보인다고 무조건 index 부족인 것도 아니며, replication lag가 보인다고 네트워크만 탓할 수도 없다. PostgreSQL의 query 실행과 운영 상태는 planner, lock manager, WAL, replication apply path가 서로 영향을 주며 나타난다. 공식 planner/optimizer 문서는 optimizer의 일이 가능한 실행 계획 중 예상상 가장 빠른 plan을 만드는 것이라고 설명한다. 공식 monitoring 문서는 database activity와 performance를 보기 위한 여러 tool이 있으며, statistics system뿐 아니라 `ps`, `top`, `iostat`, `vmstat`, 그리고 느린 query가 확인되면 `EXPLAIN`을 함께 보아야 한다고 말한다. 이 절은 `database/postgresql/introspection.log` 같은 실제 introspection output을 seed로 삼아, PostgreSQL 운영 진단을 하나의 view가 아니라 “계획 -> 실행 -> 대기 -> WAL -> 복제” 흐름으로 읽는 방법을 설명한다.

### Local seed: introspection log가 알려 주는 관측 습관

`database/postgresql/introspection.log`는 IntelliJ IDEA가 PostgreSQL 15.0 서버에 붙어 schema와 catalog를 읽는 과정을 보여 준다. 로그에는 `select current_database()`, `current_schemas(false)`, `pg_postmaster_start_time()`, `pg_locks`, `txid_current()`, `pg_database`, role 조회, timezone 조회 같은 query가 순서대로 나온다. 이 로그는 애플리케이션 query는 아니지만 운영 진단 학습에 좋은 seed다. 도구가 database를 “마법처럼 introspect”하는 것이 아니라, catalog와 monitoring view를 실제 SQL로 읽고, 각 query의 execution time과 fetched rows를 기록한다는 사실을 보여 주기 때문이다.

로그 일부를 학습용으로 줄이면 다음 흐름이다.

```text
IDE attaches to PostgreSQL
  -> select current_database(), current_schemas(false)
  -> select pg_postmaster_start_time()
  -> select from pg_locks where transactionid is not null
  -> select txid_current()
  -> select current_database(), current_schema(), current_timestamp
  -> select from pg_database
  -> show DateStyle
  -> select from pg_timezone_names / pg_timezone_abbrevs
  -> select from pg_roles and pg_auth_members
```

이 흐름은 세 가지를 가르친다. 첫째, 관측은 SQL이다. PostgreSQL은 `pg_catalog`와 `pg_stat_*` view를 통해 내부 상태의 많은 부분을 관계형 형태로 노출한다. 둘째, 관측 query도 workload다. Catalog query가 많거나 느리면 metadata lock, catalog bloat, permission, network, client tool behavior까지 원인이 될 수 있다. 셋째, 한 view는 한 질문에만 답한다. `pg_locks`는 lockable object와 process의 lock 상태를 보여 주지만, 왜 그 query가 그런 plan을 골랐는지, WAL flush가 막혔는지, replica가 따라오지 못하는지는 별도로 봐야 한다.

### Planner는 정답을 아는 것이 아니라 비용을 추정한다

PostgreSQL 공식 EXPLAIN 문서는 PostgreSQL이 각 query마다 plan을 만들고, query 구조와 data properties에 맞는 plan 선택이 성능에 결정적이며, `EXPLAIN`으로 planner가 만든 plan을 볼 수 있다고 설명한다. Planner/optimizer 문서는 scan, join, sort, aggregate 같은 후보 plan을 만들고, 계산 가능하다면 여러 후보 중 가장 빠를 것으로 예상되는 plan을 고른다고 설명한다. 여기서 “예상”이 중요하다. Planner는 실제 미래를 아는 것이 아니라 statistics와 cost model로 추정한다.

다음 query를 보자.

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM orders
WHERE tenant_id = 10
  AND status = 'READY'
ORDER BY created_at
LIMIT 50;
```

좋은 plan이라면 `(tenant_id, status, created_at)` index를 이용해 필요한 50건을 빨리 찾을 수 있다. 하지만 statistics가 낡았거나 column correlation이 다르거나 LIMIT와 ORDER BY를 만족하는 index가 없으면 planner는 seq scan, bitmap heap scan, sort를 고를 수 있다. Plan은 tree다. 아래 node가 raw row를 만들고, 위 node가 filter, join, sort, aggregate를 수행한다. `actual rows`와 `estimated rows`가 크게 어긋나면 planner가 세상을 잘못 보고 있다는 신호다.

```text
Limit
  -> Index Scan using idx_orders_tenant_status_created
       Index Cond: (tenant_id = 10 AND status = 'READY')
       Buffers: shared hit=120 read=4

대안으로 나쁜 상황:
Limit
  -> Sort
       Sort Key: created_at
       -> Seq Scan on orders
            Filter: tenant_id=10 AND status='READY'
            Rows Removed by Filter: 9,800,000
            Buffers: shared hit=10000 read=80000
```

이 trace에서 진단 질문은 “왜 seq scan이 나왔지?” 하나가 아니다. Index가 없어서인지, index는 있지만 선택도가 낮다고 봤는지, `ORDER BY`와 index order가 맞지 않는지, table statistics가 낡았는지, parameterized query에서 generic plan이 쓰였는지, visibility map 상태 때문에 index-only scan이 깨졌는지 봐야 한다. PostgreSQL deep-dive에서 planner는 고립된 optimizer가 아니라 heap/vacuum/statistics와 연결된 판단자다.

### Lock view는 wait graph의 일부만 보여 준다

공식 `pg_locks` 문서는 active process가 held한 lock 정보를 제공하고, lockable object, requested lock mode, process마다 한 row가 나올 수 있다고 설명한다. 따라서 같은 object가 여러 process에 의해 holding 또는 waiting 상태라면 여러 row가 보인다. 하지만 object에 현재 lock이 없으면 아예 나타나지 않는다. 이 말은 `pg_locks`가 “전체 object 목록”이 아니라 “현재 lock 관계 snapshot”이라는 뜻이다.

Lock 진단의 기본 query는 `pg_stat_activity`와 `pg_locks`를 붙이는 것이다.

```sql
SELECT
  a.pid,
  a.state,
  a.wait_event_type,
  a.wait_event,
  now() - a.xact_start AS xact_age,
  l.locktype,
  l.mode,
  l.granted,
  l.relation::regclass AS relation_name,
  a.query
FROM pg_stat_activity a
LEFT JOIN pg_locks l
  ON a.pid = l.pid
WHERE a.datname = current_database()
ORDER BY a.xact_start NULLS LAST, a.pid;
```

이 query가 보여 주는 것은 “누가 어떤 lock을 들고 있거나 기다리는가”다. Root cause는 아직 아니다. 어떤 backend가 relation lock을 기다리는 이유는 앞 transaction이 schema migration 중이기 때문일 수도 있고, long transaction이 row lock을 들고 있기 때문일 수도 있고, autovacuum과 DDL이 충돌했기 때문일 수도 있다. `wait_event_type='Lock'`은 출발점일 뿐이다. 그 다음에는 blocking PID를 찾아 query text, transaction age, application name, client address, plan, migration 작업 여부를 확인해야 한다.

```sql
SELECT
  blocked.pid AS blocked_pid,
  blocked.query AS blocked_query,
  blocker.pid AS blocker_pid,
  blocker.state AS blocker_state,
  now() - blocker.xact_start AS blocker_xact_age,
  blocker.query AS blocker_query
FROM pg_stat_activity blocked
JOIN LATERAL unnest(pg_blocking_pids(blocked.pid)) AS bpid(pid) ON true
JOIN pg_stat_activity blocker ON blocker.pid = bpid.pid;
```

PASS 신호는 blocked와 blocker를 분리하고, blocker가 왜 lock을 들고 오래 남았는지 transaction boundary까지 보는 것이다. FAIL 신호는 blocked query만 kill하거나, 가장 오래된 query만 보고 장애 원인을 단정하는 것이다. Blocking query는 짧은 update였지만 transaction이 commit되지 않아 lock을 계속 들고 있을 수 있다. 반대로 오래 실행 중인 analytical query는 lock waiter가 아니라 I/O pressure를 만드는 원인일 수 있다.

### WAL은 durability와 replication의 공통 통로다

PostgreSQL WAL 문서는 write-ahead logging이 data integrity를 보장하는 표준 방법이며, table과 index가 있는 data file의 변경은 그 변경을 설명하는 WAL record가 permanent storage에 flush된 뒤에만 data file에 쓰여야 한다고 설명한다. 이 원리는 crash recovery에 필요하다. Dirty page가 data file에 아직 flush되지 않았더라도 WAL이 durable하면 restart 후 replay로 일관성을 되찾을 수 있다. 그러나 PostgreSQL 운영에서 WAL은 recovery만의 문제가 아니다. Streaming replication은 primary가 생성한 WAL record를 standby가 받아 적용하는 방식이다. 공식 warm standby 문서는 standby가 primary에서 WAL을 TCP connection으로 직접 받을 수 있고, streaming replication은 WAL file이 다 찰 때까지 기다리지 않고 생성되는 WAL record를 standby로 stream한다고 설명한다. 기본은 asynchronous이므로 primary commit과 standby visible 사이에는 delay가 있을 수 있다.

WAL 흐름을 query 실행과 연결하면 다음과 같다.

```text
client COMMIT
  -> backend generated WAL records for heap/index changes
  -> WAL flushed according to synchronous_commit / durability path
  -> transaction commit visible on primary
  -> WAL sender streams records to standby
  -> standby WAL receiver writes them
  -> startup/apply process replays records
  -> change becomes visible on standby
```

Replication lag를 볼 때 이 흐름을 쪼개야 한다. Primary에서 WAL 생성이 폭증했는가, WAL sender가 standby로 보내지 못하는가, standby가 receive는 했지만 flush가 느린가, flush는 했지만 replay가 느린가, standby query conflict 때문에 replay가 지연되는가를 나눠야 한다. Lag 숫자 하나만 보고 network 문제라고 단정하면 틀릴 수 있다. Write-heavy workload, long-running standby read query, disk I/O, checkpoint, vacuum, DDL, replication slot retention이 모두 후보가 된다.

### Planner, lock, WAL, replication을 한 장애 trace로 연결하기

실무 장애는 보통 한 층에서만 끝나지 않는다. 다음은 흔한 결합형 trace다.

```text
T0  application deploy adds new query:
    SELECT * FROM order_event
    WHERE tenant_id = ? AND status = ?
    ORDER BY created_at
    LIMIT 100;

T1  missing composite index:
    planner chooses seq scan + sort
    many shared buffers read, high I/O

T2  batch worker updates selected rows:
    UPDATE order_event SET status='PROCESSING' ...
    row locks held longer because each batch scans too much

T3  autovacuum cannot keep up:
    dead tuples increase
    visibility map bits cleared often
    planner statistics become stale between analyze cycles

T4  WAL volume increases:
    heap updates + index updates + vacuum records create more WAL
    standby receives but replay lags

T5  dashboard reads from replica:
    stale status observed
    team suspects replication/network first

real diagnosis:
    plan shape and lock footprint caused write amplification,
    WAL/replication lag is downstream symptom.
```

이 trace의 목적은 PostgreSQL을 겁주는 것이 아니다. 운영 진단에서 view 하나로 결론을 내리면 downstream symptom을 root cause로 착각한다는 점을 보여 주려는 것이다. `pg_stat_replication` lag가 보이면 replication을 봐야 하지만, 동시에 primary의 query plan, row update volume, lock wait, vacuum stats, WAL generation도 봐야 한다. `EXPLAIN`이 seq scan을 보여 주면 index를 검토해야 하지만, 동시에 table bloat와 stale stats가 planner 선택을 왜곡했는지도 봐야 한다.

### 관측 query 묶음: 한 화면이 아니라 decision tree

PostgreSQL 운영 진단은 다음 묶음으로 시작할 수 있다.

```sql
-- 1. 현재 activity와 wait event
SELECT
  pid, application_name, state,
  wait_event_type, wait_event,
  now() - xact_start AS xact_age,
  now() - query_start AS query_age,
  query
FROM pg_stat_activity
WHERE datname = current_database()
ORDER BY query_start NULLS LAST;

-- 2. blocking 관계
SELECT
  blocked.pid AS blocked_pid,
  blocker.pid AS blocker_pid,
  now() - blocker.xact_start AS blocker_xact_age,
  blocker.query AS blocker_query,
  blocked.query AS blocked_query
FROM pg_stat_activity blocked
JOIN LATERAL unnest(pg_blocking_pids(blocked.pid)) AS b(pid) ON true
JOIN pg_stat_activity blocker ON blocker.pid = b.pid;

-- 3. relation statistics와 dead tuple
SELECT
  relname, n_live_tup, n_dead_tup,
  last_autovacuum, last_autoanalyze
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC
LIMIT 20;

-- 4. replication apply 상태
SELECT
  application_name,
  state,
  sent_lsn,
  write_lsn,
  flush_lsn,
  replay_lsn,
  write_lag,
  flush_lag,
  replay_lag
FROM pg_stat_replication;
```

각 query의 PASS/FAIL은 다르다. `pg_stat_activity`는 wait가 있는 backend와 오래 열린 transaction을 찾는 데 좋다. 하지만 plan을 알려 주지는 않는다. Blocking query는 lock owner를 찾는 데 좋다. 하지만 owner가 왜 오래 걸리는지는 `EXPLAIN`, application transaction boundary, I/O 지표를 봐야 한다. `pg_stat_user_tables`는 vacuum/analyze 후보를 보여 준다. 하지만 dead tuple 추정은 추정치이고, table별 workload와 autovacuum 설정을 함께 봐야 한다. `pg_stat_replication`은 sender 관점의 lag를 보여 준다. 하지만 standby 내부 replay conflict는 standby 쪽 log와 activity도 봐야 한다.

### EXPLAIN은 실행 전/후를 구분해서 써야 한다

`EXPLAIN`만 실행하면 실제 query를 실행하지 않고 plan estimate를 보여 준다. `EXPLAIN ANALYZE`는 query를 실제로 실행하고 actual time, rows, loops를 보여 준다. 이 차이를 모르고 production에서 write query에 `EXPLAIN ANALYZE`를 붙이면 실제 변경을 실행할 수 있다. 안전하게 보려면 transaction 안에서 rollback하거나, read-only replica 또는 staging에서 재현하거나, write query는 먼저 SELECT equivalent와 index/statistics를 본다. 이 점은 설명보다 운영 안전에 가깝지만, deep-dive 문서에서는 반드시 남겨야 하는 trap이다.

```sql
BEGIN;
EXPLAIN (ANALYZE, BUFFERS)
UPDATE order_event
SET status = 'PROCESSING'
WHERE tenant_id = 10
  AND status = 'READY';
ROLLBACK;
```

이 패턴도 만능은 아니다. 실제 row lock을 잡고 trigger를 실행하고 WAL을 만들 수 있으며, rollback 자체도 비용이 있다. Production에서 실행하기 전에 statement가 어떤 side effect를 갖는지, timeout과 lock timeout을 걸었는지, 대상 row 수가 제한되는지 확인해야 한다. Safer path는 `EXPLAIN`으로 plan을 먼저 보고, representative SELECT로 row estimate와 buffers를 확인한 뒤, staging이나 maintenance window에서 `EXPLAIN ANALYZE`를 쓰는 것이다.

### Replication lag를 숫자 하나로 보지 않기

Streaming replication은 asynchronous가 기본이므로 작은 지연은 정상일 수 있다. 문제는 지연이 어떤 단계에서 생기는지다.

```text
primary WAL position:
  sent_lsn   -> primary가 standby로 보낸 위치

standby receive/write:
  write_lsn  -> standby가 받은 WAL을 OS/write path에 쓴 위치
  flush_lsn  -> durable하게 flush한 위치

standby replay:
  replay_lsn -> data file에 적용해 query가 볼 수 있게 된 위치

lag type:
  sent - write   : network / receiver / standby write pressure
  write - flush  : standby fsync / disk pressure
  flush - replay : replay apply / conflict / long standby query
```

이 모델은 실제 PostgreSQL view column 이름과 버전별 차이를 확인해야 하지만, 진단 구조로 유용하다. `write_lag`, `flush_lag`, `replay_lag`가 모두 같은 방향으로 커지는지, replay만 커지는지, primary의 WAL generation이 갑자기 늘었는지, standby에 long query가 있는지 분리해야 한다. Replica에서 stale read가 문제라면 애플리케이션도 “방금 쓴 값을 replica에서 바로 읽어야 하는가”라는 consistency 계약을 점검해야 한다. DB만의 문제가 아니라 read-after-write routing, session consistency, retry UX의 문제이기도 하다.

### Senior practical traps

첫 번째 함정은 `pg_stat_activity`의 현재 query만 보고 transaction을 보지 않는 것이다. `state='idle in transaction'` backend는 현재 아무 일을 안 하는 것처럼 보여도 lock과 snapshot을 계속 들고 있을 수 있다. Vacuum 방해, DDL blocking, row lock hold의 원인이 될 수 있으므로 `xact_start`와 `state`를 함께 봐야 한다.

두 번째 함정은 seq scan을 무조건 나쁘게 보는 것이다. 작은 table, 낮은 selectivity, cold index, 병렬 scan, 통계상 대부분 row를 읽어야 하는 query에서는 seq scan이 맞을 수 있다. 문제는 seq scan 자체가 아니라 estimated/actual rows 차이, buffers read, sort spill, query frequency, lock hold time과 결합된 비용이다.

세 번째 함정은 WAL lag를 replication 팀 문제로만 넘기는 것이다. Primary의 bad plan과 hot update가 WAL volume을 늘리고, vacuum과 checkpoint가 I/O를 만들며, standby query conflict가 replay를 늦출 수 있다. Lag는 복제 계층에서 보이는 증상일 수 있지만, 원인은 query design과 write pattern일 수 있다.

네 번째 함정은 catalog/introspection query를 무해한 배경 소음으로 보는 것이다. IDE나 migration tool이 많은 catalog query를 실행하면 connection, lock, catalog cache, statistics view에 부하를 줄 수 있다. `introspection.log`처럼 어떤 query가 실제로 나가는지 남아 있으면, tool behavior도 운영 workload로 읽어야 한다.

### 직접 replay할 작은 실험

다음 실험은 planner와 statistics, lock, WAL 관측을 작게 연결한다.

```sql
CREATE TABLE pg_ops_lab (
  id bigserial PRIMARY KEY,
  tenant_id int NOT NULL,
  status text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  payload text NOT NULL DEFAULT repeat('x', 100)
);

INSERT INTO pg_ops_lab (tenant_id, status, created_at)
SELECT
  (g % 10),
  CASE WHEN g % 20 = 0 THEN 'READY' ELSE 'DONE' END,
  now() - (g || ' seconds')::interval
FROM generate_series(1, 100000) AS g;

ANALYZE pg_ops_lab;

EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM pg_ops_lab
WHERE tenant_id = 3 AND status = 'READY'
ORDER BY created_at
LIMIT 20;

CREATE INDEX idx_pg_ops_lab_claim
ON pg_ops_lab (tenant_id, status, created_at);

ANALYZE pg_ops_lab;

EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM pg_ops_lab
WHERE tenant_id = 3 AND status = 'READY'
ORDER BY created_at
LIMIT 20;
```

PASS 신호는 index 생성 전후 plan node, estimated/actual rows, buffers가 어떻게 바뀌는지 설명하는 것이다. FAIL 신호는 실행 시간이 줄었다는 한 줄만 보고 결론내리는 것이다. Row distribution, cache warmness, table size, selectivity, visibility map 상태가 달라지면 결과도 달라질 수 있다. 이 실험은 “planner는 statistics와 available path를 바탕으로 선택한다”는 구조를 재현하기 위한 것이지, 모든 query에 composite index를 추가하라는 처방이 아니다.

### Checkpoint와 WAL flush를 I/O 증상과 연결하기

WAL을 replication의 입력으로만 보면 primary의 I/O 증상을 놓친다. WAL은 commit durability와 crash recovery의 핵심 경로이기도 하다. 많은 transaction이 짧은 시간에 commit하면 WAL write와 flush가 증가한다. Data page는 나중에 checkpoint와 background writer 경로로 flush될 수 있지만, WAL flush가 지연되면 commit latency가 직접 늘 수 있다. 반대로 checkpoint가 몰리면 data file write가 한꺼번에 늘어 query read I/O와 경쟁할 수 있다. PostgreSQL 운영에서 `iostat`와 `pg_stat_activity` wait event를 함께 보는 이유가 여기에 있다.

```text
write-heavy burst:
  many UPDATE/INSERT
      -> heap/index page dirty
      -> WAL records generated
      -> commit waits for WAL flush depending on settings
      -> checkpoint later writes dirty buffers
      -> standby receives and replays WAL

observability:
  commit slow     : WALWrite/WALSync wait candidates
  read query slow : shared buffer miss + data file read candidates
  checkpoint I/O  : background write pressure
  replica lag     : WAL send/write/flush/replay gap
```

이 trace는 정확한 wait event 이름을 외우라는 뜻이 아니다. Version별 wait event는 달라질 수 있고, extension이나 cloud provider도 관측면을 바꿀 수 있다. 중요한 것은 write workload가 WAL, data file flush, standby replay를 모두 건드린다는 점이다. Query plan이 나빠져 update 대상 row가 100배 늘면 lock hold time만 늘어나는 것이 아니라 WAL volume도 늘고, checkpoint와 replication lag까지 영향을 받을 수 있다.

### Catalog와 statistics view를 볼 때의 권한과 비용

PostgreSQL은 관측 view를 SQL로 제공하기 때문에 편리하지만, 모든 관측이 공짜는 아니다. `pg_stat_activity`는 backend activity를 보지만 query text visibility는 권한에 따라 제한될 수 있다. `pg_locks`는 현재 lock snapshot을 만들고, catalog join을 많이 하면 그 자체가 CPU와 memory를 쓴다. `introspection.log`의 IDE query처럼 schema metadata를 넓게 읽는 도구는 개발 중에는 편리하지만, production에서 너무 자주 실행되면 catalog와 connection pool에 부담을 줄 수 있다. 따라서 관측 query도 목적과 범위를 좁혀야 한다.

좋은 운영 query는 세 가지 특징을 가진다. 첫째, schema나 relation을 제한한다. 둘째, 오래 실행되는 backend나 wait가 있는 backend처럼 우선순위를 둔다. 셋째, 결과를 바로 행동으로 연결하지 않고 다른 view로 반증한다. 예를 들어 `pg_stat_activity`에서 `wait_event_type='Lock'`을 발견하면 바로 kill하지 않고 `pg_blocking_pids`, `pg_locks`, blocker query, transaction age를 확인한다. `pg_stat_replication`에서 replay lag를 발견하면 primary write spike, standby wait event, standby long query, replication slot retention을 나눠 본다.

### 장애 대응 runbook으로 합치기

Planner, lock, WAL, replication을 한 runbook으로 묶으면 다음 순서가 된다.

```text
1. 증상 고정
   - latency increase, lock timeout, stale replica read, CPU/I/O saturation 중 무엇인가

2. 현재 activity 확인
   - pg_stat_activity: active/idle in transaction/wait event/query age

3. blocking 여부 확인
   - pg_blocking_pids + pg_locks
   - blocker transaction age와 query boundary

4. bad plan 후보 확인
   - EXPLAIN, estimated vs actual rows, buffers
   - stale stats, missing index, bloat, visibility map 영향

5. write/WAL pressure 확인
   - write query volume, checkpoint pressure, WAL generation
   - commit latency와 disk wait

6. replication 확인
   - sent/write/flush/replay lag 분리
   - standby long query와 replay conflict 확인

7. 조치 선택
   - cancel/terminate blocker, add/index, analyze/vacuum, traffic shed,
     route reads to primary, pause migration, tune autovacuum 등
```

이 순서의 핵심은 먼저 증상을 고정하고, 그 다음 관측 view를 서로 연결한다는 점이다. Lock timeout이 증상인데 replication부터 보면 늦고, stale replica read가 증상인데 primary `EXPLAIN`만 보면 부족하다. 그러나 어느 경우에도 한 view로 단정하지 않는다. `pg_stat_activity`는 현재 backend state, `EXPLAIN`은 plan shape, `pg_locks`는 wait graph, WAL/replication view는 변경 전파 경로를 담당한다. 서로 다른 질문에 답하는 view를 한 질문에 억지로 쓰면 오진이 생긴다.

### Application consistency와 replica read

Replication deep-dive에서 자주 빠지는 것은 애플리케이션의 읽기 일관성 계약이다. Streaming replication이 asynchronous인 경우 primary commit 직후 standby에서 같은 값을 읽는 것이 보장되지 않는다. 따라서 “복제 지연이 500ms밖에 안 된다”와 “사용자가 결제 완료 직후 주문 상태를 본다”는 서로 다른 문장이다. 사용자 flow가 read-after-write consistency를 요구하면 해당 요청은 primary로 보내거나, session stickiness, LSN wait, retry/backoff, UI pending state 같은 설계가 필요하다. DB 지표가 정상 범위여도 application contract가 더 강하면 장애로 느껴질 수 있다.

```text
request flow:
  POST /payments/complete
    -> primary commit at LSN 0/5000
    -> response OK

  GET /orders/1001 immediately after
    -> routed to replica
    -> replica replay_lsn still 0/4F00
    -> old status observed

DB view:
  replay lag = small but non-zero

user view:
  "방금 결제했는데 아직 READY로 보여요"
```

이 trace는 replication lag를 기술 지표로만 보지 않게 해 준다. 어떤 화면은 약간의 lag를 허용할 수 있고, 어떤 화면은 허용할 수 없다. 따라서 PostgreSQL replication 관측은 application routing과 함께 설계해야 한다. Senior trap은 “lag가 작으니 문제 없음”이다. 문제 여부는 lag 숫자와 사용자/업무 일관성 요구를 함께 봐야 결정된다.

### Interview 관점의 압축 답변

PostgreSQL 운영 진단을 짧게 설명해야 한다면 이렇게 말할 수 있다. Planner는 statistics와 available access path를 바탕으로 예상 비용이 가장 낮은 plan을 고른다. `EXPLAIN`은 그 plan tree와 estimate를 보여 주고, `EXPLAIN ANALYZE`는 실제 실행 결과와 차이를 보여 주지만 side effect에 주의해야 한다. Lock 문제는 `pg_stat_activity`, `pg_locks`, `pg_blocking_pids`로 blocked와 blocker를 분리해 transaction boundary까지 봐야 한다. WAL은 crash recovery와 replication의 공통 log이며, write-heavy workload는 primary commit latency, checkpoint I/O, standby replay lag로 동시에 나타날 수 있다. 그래서 PostgreSQL 장애 분석은 하나의 view가 아니라 plan, wait, storage maintenance, WAL propagation을 연결하는 일이다.

이 답변의 required trap은 관측 view 하나만 보고 원인을 단정하지 않는 것이다. `pg_locks`는 lock 관계를 보여 주지만 plan을 설명하지 않고, `EXPLAIN`은 plan을 보여 주지만 wait owner를 알려 주지 않고, replication lag view는 전파 지연을 보여 주지만 primary에서 왜 WAL이 폭증했는지 설명하지 않는다. 각 view의 질문을 정확히 맞춰야 한다.

### 출처와 검증 경계

이 절의 공식 근거는 PostgreSQL 18 문서의 [Planner/Optimizer](https://www.postgresql.org/docs/current/planner-optimizer.html), [Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html), [Monitoring Database Activity](https://www.postgresql.org/docs/current/monitoring.html), [The Cumulative Statistics System](https://www.postgresql.org/docs/current/monitoring-stats.html), [pg_locks](https://www.postgresql.org/docs/current/view-pg-locks.html), [Write-Ahead Logging](https://www.postgresql.org/docs/current/wal-intro.html), [Log-Shipping Standby Servers](https://www.postgresql.org/docs/current/warm-standby.html)다. 로컬 seed는 `database/postgresql/introspection.log`와 PostgreSQL 관측 메모다. 이 절의 plan, lock, WAL, replication trace는 운영 판단을 위한 모델이며, 실제 column 이름과 wait event, replication lag 표기는 PostgreSQL version과 설정에 따라 다를 수 있다. 검증은 local/staging에서 `EXPLAIN`, `pg_stat_activity`, `pg_locks`, `pg_stat_user_tables`, `pg_stat_replication`을 함께 찍고, 하나의 view에서 나온 가설을 다른 view와 log로 반증해 보는 방식으로 닫아야 한다.

### 마지막 실무 연결: PostgreSQL 진단은 view 하나로 끝나지 않는다

PostgreSQL은 관측 view가 풍부하지만, view 하나만 보고 결론 내리면 위험하다. `pg_stat_activity`는 누가 기다리는지 보여 주지만 왜 그런 plan이 선택되었는지는 말하지 않는다. `EXPLAIN`은 실행 경로를 보여 주지만 replication lag나 WAL flush 압력을 직접 설명하지 않는다. `pg_locks`는 잠금 상태를 보여 주지만 application request id와 연결하지 않으면 업무 영향이 흐려진다.

따라서 PostgreSQL 운영 진단은 planner, lock, WAL, replication을 같은 시간축에 놓는 일이다. 느린 쿼리 하나를 볼 때도 row estimate, lock wait, WAL write, replica replay, autovacuum 활동을 함께 본다. 이 습관이 있어야 관측 view를 많이 아는 것에서 실제 원인을 좁히는 능력으로 넘어갈 수 있다.
