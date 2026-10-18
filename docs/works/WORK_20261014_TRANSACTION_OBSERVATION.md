# 트랜잭션을 실제 실행으로 보기: 격리 수준, 락, 원자성·지속성, 전파

- 시작: 2026-10-14 20:26 KST · 실행자: Claude Code(`claude-opus-5-5`) · 깊이: full · finish: 실험 1~5의 실제 기록과 작업대 화면, 실험마다 사전 등록과 판정
- 상태: 실험 1·2 기록·판정·화면 완료(실험 1은 E1·E6, 실험 2는 L5·L12가 예상과 다름, 아래 실행 기록), 실험 3 기록·판정·화면 완료(D6의 off 두 실행과 D10이 예상과 다름, 아래 실행 기록), 실험 4 기록·판정·화면 완료(W1~W13 모두 예상대로, W3은 함수 이름만 틀림), 실험 5 기록·판정 완료(P1~P11 가운데 P6은 관찰 결함을 고친 뒤 다시 기록해 통과, P9는 한 접속이 0.33 ms 차이로 통과 조건의 창을 벗어남(실패 조건은 없음), P2의 나머지 쪽은 덧붙인 장면 5-reborrow로 판정해 통과, P12·P13 통과, 아래 실행 기록)·화면 첫 판(concept-9) 완료, 사용자 요청으로 AI 참가자 사용성 시험 네 판(1차 haiku 미달 → 2차 두 참가자 통과 → 3차 클로드가 아닌 모델 둘, 맥락 없이 → 글을 그림으로 옮긴 v2 → 4차 그림만 보는 갈래와 3차 화면 대조군까지 여섯 쌍 모두 통과, 4차에서 나온 J1~J5는 아직 고치지 않음), 사용자 의견 대기 · 작업 상태의 정본은 [question-led-learning PLAN](../../ai/authoring/question-led-learning/PLAN.md)의 QLL-08 행 · 앞 단계: [작업대 관측 재설계 WORK](WORK_20261013_WORKBENCH_OBSERVATION_DESIGN.md)(요청 하나의 여정, 커널 추적)

## 요청과 목적

사용자 원문(2026-10-14, 받은 순서대로):

1. “아 혹시 트랜잭션도 실험 가능한가요? 락을 잡는 경우. 가령 멱등성 같은 예시를 많이 드는데요. 은행 예제도 그렇고. 그런 것들을 이 도구 통해서 볼 수 있을까요?”
2. “앞서 말한 경우 외에도 ACID를 주로 말하고, propagation과 reuiqred, requires new, read committed, phantom read 등 다양한 것들을 글로 배우지만, 눈으로 본 적은 없습니다. 그런 것들을 눈으로 보면서 배우고 싶은 겁니다.”
3. “1부터 5까지 진행해주세요. 근데 제가 spring을 예로 든 것이지, 어떤 프레임워크든 어떤 언어든 동작은 비슷하지 않나요? 저는 그 내부 동작을 이해하고 싶은 거긴 합니다. 뭐 스프링 많이 쓰니 스프링, hikari cp 그리고 커넥션이 맺어지는 과정 등을 배우면 좋을 거 같긴 합니다. 특히 L4가 끼어 있거나 proxy sql을 통하거나 aurora db 등 다양한 것들이 있어서... 그냥 저는 눈으로 다 보고 싶습니다. 실제로 동작하는 것을. 상상으로, 머릿속으로만 그리는 것은 그만하고, 진짜 보고 싶은 겁니다. 어떻게 동작하는지”

“1부터 5”는 직전 답에서 제안한 순서다. 1 행 버전과 스냅숏(READ COMMITTED와 REPEATABLE READ, non-repeatable read, 팬텀, dirty read가 안 생기는 이유), 2 락(덮어쓰기, 바른 UPDATE, 교착, 외래 키의 숨은 잠금), 3 원자성과 지속성(`ROLLBACK`, `synchronous_commit` 비교, `kill -9` 뒤 복구), 4 SERIALIZABLE과 write skew, 5 Spring 전파(REQUIRED, rollback-only, REQUIRES_NEW와 그 교착, NESTED).

원문 3의 나머지는 범위를 넓힌다. AI 해석으로 둔 구분은 다음과 같다.

- “어떤 프레임워크든 동작은 비슷하지 않나” → 층마다 답이 다르다. 연결 위의 프로토콜과 DB 엔진은 언어와 무관하고, 풀의 동작도 HikariCP·pgx·psycopg가 비슷하다. 전파(propagation)는 프레임워크의 개념이다. Go `database/sql`에는 전파가 없고 `*sql.Tx`를 직접 넘긴다. 같은 이름도 프레임워크마다 뜻이 다를 수 있으므로(예: Spring의 `REQUIRES_NEW`는 연결을 하나 더 쓴다) 이 차이를 실제 바이트로 견주는 것을 관측 대상으로 둔다. 다른 프레임워크의 동작은 각 공식 문서로 확인하기 전까지 추론이다.
- “커넥션이 맺어지는 과정”, HikariCP → 실험 6(연결의 일생)으로 등록한다. 실험 5의 Java 클라이언트가 그대로 쓰인다.
- “L4가 끼어 있거나 proxy sql을 통하거나” → 실험 7(사이에 낀 장치)로 등록한다. L4는 기존 [L4 장애 실험대](../../computer_architecture/network/l4_failover_lab/README.md)(IPVS)를 다시 쓴다. ProxySQL은 MySQL용이고, PostgreSQL에서 같은 자리는 PgBouncer다. 둘 다 이미지를 내려받아야 하므로 사용자 허락이 필요하다.
- “aurora db” → 로컬에서 재현할 수 없다. Aurora는 AWS가 운영하는 서비스라 서버 안의 함수와 저장 계층을 기록할 수 없고, 클라이언트 쪽(엔드포인트 DNS, 장애 조치 때 연결이 끊기는 모양)만 볼 수 있다. AWS 계정과 비용이 필요하므로 사용자 결정으로 둔다.

이 단계가 판단할 것: **글로 배운 트랜잭션 개념마다, 같은 실행의 여러 세션을 한 시간축에 놓고, 스냅숏·행 버전·락·커밋 로그의 실제 값과 그 값을 쓴 소스 줄을 그림으로 보일 수 있는가.**

## 관측 방법: 함수 인자와 반환값까지

지금까지의 기록(컴파일러가 모든 함수 입구·출구에 넣은 기록 호출, strace, 커널 ftrace)은 “어느 함수가 언제 불렸는지”는 주지만 인자와 반환값은 주지 않는다. 그래서 “이 행 버전이 이 SELECT에 안 보였다”의 이유(스냅숏의 어느 값과 행의 어느 값을 견줬는지)는 지금까지 계산이나 추론이었다.

가능성 확인(20:05, 이번 단계 첫 시험): OrbStack 머신 `tracelab`에서 건 uprobe가 Docker 컨테이너 안 프로세스에서 걸렸다. 맥 폴더를 두 쪽이 같은 virtiofs로 보므로 같은 파일이 같은 inode다(시험 파일: 컨테이너 `ino=187 dev=37`, 머신 `ino=187 dev=37`). 시험 프로그램 `probe_me(a)`를 세 번 부르자 `enter: a=40`, `ret: ret=120`처럼 인자와 반환값이 기록됐다. 등록한 uprobe는 시험 뒤 지웠다.

이 결과로 정한 방법:

1. 기록 빌드의 `postgres` 실행 파일을 컨테이너와 머신이 함께 보는 맥 경로(`open_source_internals/database/postgresql/scripts/.lab-tmp/`, 사용자의 전역 제외 규칙 `*tmp/`로 git에 들어가지 않음)에 두고, 서버를 그 파일로 띄운다. 컨테이너에서는 읽기 전용 마운트 `/lab`으로 보인다.
2. uprobe 정의(함수의 파일 안 위치, 구조체 필드의 위치)는 손으로 적지 않고 Go 도구가 실행 파일의 DWARF에서 만든다. 이 빌드는 `-O0 -g3`라서 관측할 함수가 모두 실제 함수로 남아 있다(`HeapTupleSatisfiesMVCC`, `GetSnapshotData`, `XidInMVCCSnapshot`, `XactLockTableWait`, `ProcSleep` 등 `nm`으로 확인). 실행 파일은 위치 독립(PIE)이고 코드 구간이 파일 위치 0에서 시작하므로 함수의 파일 안 위치는 기호의 주소와 같다(`readelf -l`).
3. 사건마다 컨테이너 안 pid를 함께 남긴다. ftrace는 커널의 전역 pid를, 함수 기록과 strace는 컨테이너 안 pid를 쓰기 때문이다. 전역 변수 `MyProcPid`(주소 0xe76254)를 사건마다 읽는다(uprobe의 `@+주소` 읽기).
4. uprobe 사건과 함수 기록의 짝짓기는 시각이 아니라 **pid마다 같은 함수의 몇 번째 호출인가**로 한다. ftrace의 시계는 CLOCK_REALTIME이 아니기 때문이다. 판정: 짝지은 함수마다 uprobe 사건 수와 함수 기록의 입구 수가 같다.

여러 세션을 정해진 순서로 움직이는 도구는 Go로 쓴다. 지난 단계의 셸 스크립트 셋(`run.sh`, `serve.sh`, `kernel.sh`)이 되돌리기(eval)와 프로세스 정리(pkill)에서 두 번 실패했고, 사용자 지시(브라우저 밖은 Go)와도 맞다. 세션마다 실제 클라이언트인 psql을 하나씩 띄워 파이프로 문장을 보내고, 문장이 끝난 것은 psql의 `\echo` 표지로, 문장이 락을 기다리는 것은 관찰용 접속에서 `pg_stat_activity.wait_event_type`으로 안다.

관찰용 접속은 단계마다 같은 것을 읽는다. 그 자체가 기록을 흐리지 않는지가 위험이라 아래처럼 읽는다.

- 페이지 내용: `heap_page_items(get_raw_page(...))`. 페이지 사본을 읽으므로 힌트 비트를 쓰지 않는다.
- 세션의 상태: `pg_stat_activity`의 `backend_xid`(받은 트랜잭션 번호), `backend_xmin`(그 세션이 쥔 스냅숏의 xmin), `wait_event_type`, `wait_event`.
- 락: `pg_locks`의 세션 pid 행.
- 커밋 로그: 시나리오에 나온 번호마다 `pg_xact_status(xid)`.
- 관찰용 접속의 백엔드도 함수 기록에 남는다. 해석은 pid로 거른다.

## 실험 1 사전 등록(20:27, 실행 전): 행 버전과 스냅숏

### 시나리오

표 `acct(id int PRIMARY KEY, owner text, balance int)`에 `(1,'kim',100)`, `(2,'lee',200)`을 넣고 시작한다. 세션 1(S1)이 읽고, 세션 2(S2)가 바꾼다. 같은 시나리오를 두 번 돌려 견준다(지난 단계의 교훈: 한 번의 관측을 경로의 성질로 단정하지 않는다). 번호는 실행마다 다르고 번호 사이의 관계를 견준다.

| 이름 | 순서 | 보려는 것 |
| --- | --- | --- |
| 1-rc-read | S1 `BEGIN ISOLATION LEVEL READ COMMITTED` → S1 SELECT → S2 `UPDATE id=1 balance=150`(자동 커밋) → S2 `INSERT (3,'park',300)`(자동 커밋) → S1 같은 SELECT → S1 COMMIT | non-repeatable read와 팬텀 |
| 1-rr-read | 위와 같고 `REPEATABLE READ` | 같은 SELECT가 같은 결과 |
| 1-ru-read | 위와 같고 `READ UNCOMMITTED` | PostgreSQL에서는 READ COMMITTED와 같은 동작 |
| 1-rc-write | S2 `BEGIN` → S2 `UPDATE id=1 balance=150` → S1 `BEGIN ISOLATION LEVEL READ COMMITTED` → S1 SELECT → S1 `UPDATE id=1 SET balance = balance - 30` → S2 COMMIT → S1 COMMIT → S1 SELECT | dirty read가 없다는 것, 기다림과 깨움, 커밋 뒤 새 버전에 다시 적용 |
| 1-rr-write | 위와 같고 S1이 `REPEATABLE READ`, 마지막 COMMIT 대신 ROLLBACK | 같은 기다림 뒤 직렬화 오류 |

SELECT는 `SELECT pg_current_snapshot() AS snap, ctid, xmin, xmax, id, balance FROM acct WHERE balance >= 100 ORDER BY id` 하나로 쓴다. `pg_current_snapshot()`을 같은 문장에 넣으면 그 문장이 쓴 스냅숏이 결과에 함께 나온다(공식 문서의 `pg_current_snapshot` 설명). `ctid`는 결과 행이 페이지의 몇 번째 줄인지 알려 주므로, uprobe가 기록한 줄마다의 판정과 결과 행을 같은 열쇠로 맞춘다.

### 근거로 읽은 소스(18.6)

- `GetTransactionSnapshot`(snapmgr.c 271행): 트랜잭션의 첫 호출이면 `GetSnapshotData(&CurrentSnapshotData)`를 부른다. 그 뒤로는 `IsolationUsesXactSnapshot()`(REPEATABLE READ 이상)이면 `CurrentSnapshot`을 그대로 돌려주고, 아니면 다시 `GetSnapshotData(&CurrentSnapshotData)`를 부른다.
- 카탈로그를 읽을 때는 같은 `GetSnapshotData`를 `&CatalogSnapshotData`로 부른다(snapmgr.c 423행). 그래서 호출 수를 셀 때는 첫 인자가 어느 정적 변수인지로 가른다. `CurrentSnapshotData`는 0xe77998, `CatalogSnapshotData`는 0xe77a68이다(gdb `info address`). 스택 규칙(카탈로그 캐시를 거쳤는지)으로 가르면 다른 이유로 틀릴 수 있어 인자로 가른다.
- `page_collect_tuples`(heapam.c 506행): 페이지의 줄 번호 1부터 끝까지, 정상 줄(`LP_NORMAL`)마다 `HeapTupleSatisfiesVisibility`를 부르고 그 안에서 `HeapTupleSatisfiesMVCC`(heapam_visibility.c 960행)를 부른다. 페이지 전체가 보임으로 표시된 경우(`all_visible`)에는 판정을 건너뛴다. 새로 만든 표는 VACUUM 전이라 그 표시가 없다.
- 행을 바꾸려다 다른 트랜잭션이 바꾸는 중이면 `heap_update`가 `XactLockTableWait`(heapam.c 3714행)로 그 트랜잭션의 끝을 기다린다. 기다림이 끝나고 다시 본 결과가 `TM_Updated`이면 `ExecUpdate`가 REPEATABLE READ에서는 직렬화 오류를 내고(nodeModifyTable.c 2604행), READ COMMITTED에서는 최신 버전을 잠근 뒤 `EvalPlanQual`(2625행)로 WHERE 조건을 새 버전에 다시 계산한다.
- 깨움: 락을 푸는 쪽이 `ProcWakeup`(proc.c 1744행) → `SetLatch`를 부르고, 기다리는 쪽이 잠들어 있으면 `kill(pid, SIGURG)`(waiteventset.c 2035행)를 보낸다.

### 예상과 판정

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| E1 | 1-rc-read: S1의 두 SELECT가 각각 `GetSnapshotData(&CurrentSnapshotData)`를 한 번씩 부른다. 결과는 1번째 `(1,100) (2,200)`, 2번째 `(1,150) (2,200) (3,300)` | 두 실행 모두 두 SELECT의 그 호출이 1번씩이고 결과가 예상과 같다 | 2번째 SELECT의 호출이 0번이거나 결과가 1번째와 같다 |
| E2 | 1-rr-read: 1번째 SELECT는 그 호출이 1번, 2번째는 0번이다. 두 결과와 `snap` 값이 같다 | 두 실행 모두 그렇다 | 2번째에도 호출이 있거나 결과가 바뀐다 |
| E3 | 1-ru-read: 호출 수와 결과가 1-rc-read와 같다 | 같다 | 다르다 |
| E4 | uprobe 짝짓기: 짝지은 함수마다 pid별 uprobe 사건 수가 함수 기록의 입구 수와 같고, ftrace의 넘침(overrun)이 0이다 | 모든 함수에서 같고 0이다 | 하나라도 다르다. 이때 줄마다의 판정은 화면에 “기록”으로 쓰지 않는다 |
| E5 | 줄마다의 판정: S1 SELECT 안의 `HeapTupleSatisfiesMVCC` 호출 수가 그때 페이지의 정상 줄 수와 같고, 반환값이 참인 줄의 위치 집합이 결과 행의 `ctid` 집합과 같다 | 모든 SELECT에서 같다 | 하나라도 다르다 |
| E6 | 1-rc-write: S1의 SELECT 결과는 `(1,100) (2,200)`이고(바뀌는 중인 150은 안 보임), 그 `snap`의 진행 중 목록에 S2의 번호가 있다. S1의 UPDATE 안에서 `XactLockTableWait`의 첫 인자가 S2의 번호다. S2의 COMMIT 요청 안에서 `ProcWakeup`의 대상 pid가 S1의 pid이고, strace에 S2 백엔드의 `kill(S1 pid, SIGURG)`가 있다. 깨어난 S1에서 `EvalPlanQual`이 불리고 마지막 SELECT의 id 1 잔액이 120이다 | 모두 그렇다 | 하나라도 다르다. 특히 잔액이 70(옛 버전 100에 적용)이면 이 예상의 전제가 틀렸다 |
| E7 | 1-rr-write: E6과 같은 기다림과 깨움 뒤, S1에 `could not serialize access due to concurrent update`가 오고, 그 보고를 부른 자리가 `ExecUpdate`(nodeModifyTable.c 2604행)다. `EvalPlanQual`은 불리지 않는다 | 모두 그렇다 | 다른 자리에서 오류가 나거나 `EvalPlanQual`이 불린다 |
| E8 | 관찰용 접속이 기록을 흐리지 않는다: S1 SELECT의 `HeapTupleSatisfiesMVCC` 반환값은 관찰용 접속이 페이지를 읽기 전과 뒤에 차이가 없다(관찰은 페이지 사본만 읽음) | 관찰용 백엔드에서 `SetHintBits`가 acct의 페이지에 대해 0번 불린다 | 관찰용 백엔드가 acct 페이지의 힌트 비트를 쓴다 |

두 실행 사이에 기다림 시간이나 호출 수가 다르면 다르다고 적고, 한 번의 값을 성질로 쓰지 않는다.

## 실험 2 사전 등록(21:55, 실행 전): 행 락

질문은 다섯이다. 읽고 계산해서 쓰면 왜 값을 잃는가, `FOR UPDATE`는 그것을 어떻게 막는가, 교착에서 누가 오류를 받는가, 외래 키는 부모 행을 어떻게 잠그는가, 멱등성 키는 두 번째 요청을 어떻게 멈추는가. 모두 READ COMMITTED이고, 시나리오마다 두 번(a, b) 기록한다.

### 시나리오

| 이름 | 준비 | 순서 | 보려는 것 |
| --- | --- | --- | --- |
| 2-lost-update | `acct`에 `(1,'kim',100)` | S1 BEGIN → S1 `SELECT balance … id=1` → S2 BEGIN → S2 같은 SELECT → S1 `UPDATE … SET balance = 70`(100−30을 앱이 계산) → S1 COMMIT → S2 `UPDATE … SET balance = 50`(100−50) → S2 COMMIT → S1 SELECT | 기다림 없이 S1의 −30이 사라짐 |
| 2-for-update | 같음 | S1 BEGIN → S1 `SELECT … FOR UPDATE` → S2 BEGIN → S2 `SELECT … FOR UPDATE`(기다림) → 쉼 1.5초 → S1 `UPDATE 70` → S1 COMMIT → S2 `UPDATE 20`(S2가 받은 70−50) → S2 COMMIT → S1 SELECT | 두 번째 읽기가 첫 쓰기를 기다리고 새 값을 받음. 쉼은 교착 검사를 언제나 보이려고 둔다 |
| 2-deadlock | `(1,'kim',100)`, `(2,'lee',200)` | S1 BEGIN → S1 `UPDATE id=1 −10` → S2 BEGIN → S2 `UPDATE id=2 −20` → S1 `UPDATE id=2 +10`(기다림) → S2 `UPDATE id=1 +20`(기다림) → S1 ROLLBACK → S2 COMMIT → S2 SELECT | 서로 반대 순서로 이체하는 두 트랜잭션 |
| 2-fk-nokey | `parent(id PK, name)`에 `(1,'a')`, `child(id PK, parent_id REFERENCES parent)` | S1 BEGIN → S1 `INSERT INTO child VALUES (10, 1)` → S2 `UPDATE parent SET name='b' WHERE id=1`(자동 커밋) → S1 COMMIT → S1 SELECT | 키가 아닌 열을 바꾸면 기다리지 않음 |
| 2-fk-delete | 같음 | S1 BEGIN → S1 `INSERT INTO child VALUES (10, 1)` → S2 `DELETE FROM parent WHERE id=1`(기다림) → S1 COMMIT → S1 SELECT | 부모를 지우면 기다린 뒤 외래 키 오류 |
| 2-idem-commit | `payment(idem_key text PRIMARY KEY, amount int)` | S1 BEGIN → S1 `INSERT ('k1',100)` → S2 같은 INSERT(자동 커밋, 기다림) → S1 COMMIT → S1 SELECT | 같은 키의 두 번째 요청 |
| 2-idem-rollback | 같음 | 위와 같고 S1이 ROLLBACK | 첫 요청이 취소되면 두 번째가 들어감 |
| 2-idem-onconflict | 같음 | 위와 같고 S2가 `… ON CONFLICT (idem_key) DO NOTHING` | 실무에서 쓰는 꼴 |

### 근거로 읽은 소스(18.6)

- 번호를 받는 때: `heap_insert`·`heap_delete`·`heap_update`는 함수 첫머리에서 `GetCurrentTransactionId()`를 부른다(heapam.c 2084, 2833, 3327행). `heap_lock_tuple`은 잠금을 적기 직전에 부른다(5365행). 그래서 `FOR UPDATE`로 기다린 세션은 깨어난 뒤에야 번호를 받는다고 예상한다.
- `FOR UPDATE`: `heap_lock_tuple`의 넷째 인자가 잠금 모드다(`LockTupleKeyShare` 0 … `LockTupleExclusive` 3, lockoptions.h). 남이 잠근 행이면 `XactLockTableWait(…, XLTW_Lock)`로 기다린다(heapam.c 5218행). 깨어나 보니 행이 바뀌었으면 `TM_Updated`를 돌려주고(5298행), READ COMMITTED에서는 최신 버전을 찾아 다시 잠근다(heapam_handler.c 384행의 `TUPLE_LOCK_FLAG_FIND_LAST_VERSION`).
- 교착: 기다림을 시작할 때 `deadlock_timeout` 타이머를 한 번 켠다(proc.c 1400행). 울리면 `CheckDeadLock` → `DeadLockCheck`가 돈다(1857행). 결과는 `DS_NO_DEADLOCK` 1, `DS_HARD_DEADLOCK` 3이다(lock.h). 교착이면 검사한 쪽이 자기를 기다림 줄에서 빼고(1874행), `LockAcquireExtended`가 `DeadLockReport`를 불러(lock.c 1235행) `deadlock detected`를 낸다(deadlock.c 1133행). 검사를 다시 켜는 코드는 없으므로 기다림마다 한 번만 검사한다.
- 외래 키: 자식에 넣으면 `RI_FKey_check`가 `SELECT 1 FROM ONLY parent x WHERE id = $1 FOR KEY SHARE OF x`를 돌린다(ri_triggers.c 426행). 부모의 키가 아닌 열을 바꾸는 UPDATE는 `LockTupleNoKeyExclusive`를 쓰고(heapam.c 3522행), 잠근 쪽이 key share뿐이면 기다리지 않고 그 잠금을 남긴다(3694행). 남길 때 옛 버전의 xmax는 두 번호를 묶은 MultiXact가 된다(`compute_new_xmax_infomask` → `MultiXactIdCreate`, 5771행). DELETE는 `LockTupleExclusive`라 key share와 충돌하고 `XactLockTableWait(…, XLTW_Delete)`로 기다린다(2990행). 지운 뒤에는 `ON DELETE NO ACTION` 검사가 자식 행을 찾으면 오류를 낸다(ri_triggers.c 2801행).
- 고유 키: 힙에 행을 먼저 넣고, 인덱스에 넣을 때 `_bt_check_unique`가 같은 키의 진행 중인 행을 찾으면 `XactLockTableWait(…, XLTW_InsertIndex)`로 기다렸다가 처음부터 다시 찾는다(nbtinsert.c 227행). 커밋된 같은 키가 있으면 `duplicate key value violates unique constraint`다(668행). `ON CONFLICT DO NOTHING`은 힙에 넣기 전에 `ExecCheckIndexConstraints`로 먼저 보고(nodeModifyTable.c 1137행), 기다림은 execIndexing.c 896행, 충돌이면 아무것도 넣지 않고 돌아간다(1177행).

### 방법의 변경

- 드라이버: 앞 단계가 끝나지 않은 세션에 다음 문장이 오면 오류를 내지 않고 그 단계가 끝날 때까지 기다린다(교착의 ROLLBACK). 기다리는 동안에도 모든 세션의 출력을 받는다. `pause: 1500ms` 줄은 그만큼 쉬고, 단계로 세지 않으며 steps.json에 따로 남긴다.
- uprobe: `heap_lock_tuple`(행 위치, 모드, 반환값), `MultiXactIdCreate`(두 번호와 상태, 돌려준 번호), `DeadLockCheck`의 반환값. 실험 1 E8의 한계를 풀려고 `SetHintBits`에는 행 머리 주소, `HeapTupleSatisfiesMVCC`에는 `t_data` 주소를 더한다.
- 기록 전 디스크 여유 하한을 5 GB에서 10 GB로 올린다(21:38에 기록과 무관한 설치가 여유를 0으로 만들었다). VM 재시작 뒤 1-rc-write를 다시 기록해 실험 1과 같은 값(짝지은 묶음 39개 모두 같음, 고친 부른 자리 2,050개, 남은 것 0, 번호 753·754, 판정 대조 둘 다 같음)이 나오는 것을 확인했다.

### 예상과 판정

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| L1 | 2-lost-update: 어느 세션도 `ProcSleep`에 들어가지 않는다. S2의 UPDATE가 확인한 행은 S1이 만든 버전(잔액 70)이고 결과는 `TM_Ok`다. 마지막 잔액은 50이다 | 두 실행 모두 그렇다 | 기다림이 있거나 잔액이 20 또는 70이다 |
| L2 | 2-for-update: S1의 번호는 `SELECT … FOR UPDATE` 안에서 `heap_lock_tuple`을 거쳐 받는다. 모드는 `LockTupleExclusive`이고, 그 단계 뒤 (0,1)의 xmax는 S1의 번호에 `HEAP_XMAX_EXCL_LOCK`·`HEAP_XMAX_LOCK_ONLY`다 | 모두 그렇다 | 번호가 다른 함수에서 오거나 잠금 플래그가 다르다 |
| L3 | 2-for-update: S2는 `heap_lock_tuple` 안에서 `XactLockTableWait(S1, Lock)`로 기다린다. 약 1초 뒤 `DeadLockCheck`가 1(교착 없음)을 돌려주고 다시 잠든다. S1의 COMMIT이 깨운 뒤 (0,1)에 `TM_Updated`, 새 버전 (0,2)에 `TM_Ok`를 받는다. S2의 번호는 깨어난 뒤에 받는다. S2의 SELECT는 70을 받고 마지막 잔액은 20이다. 이 SELECT의 MVCC 판정은 자기 스냅숏으로 (0,1)만 보인다고 하지만, 결과 행의 `ctid`는 잠근 최신 버전 (0,2)다. 그래서 L12의 “보인 줄 = 결과 행” 대조가 이 단계에서만 어긋나는 것이 예상이다 | 모두 그렇다 | S2가 100을 받거나 번호를 기다리기 전에 받는다 |
| L4 | 2-deadlock: 둘 다 기다리는 동안 먼저 울린 타이머의 주인이 오류를 받는다. 드라이버는 S1이 기다리는 것을 보고서야 S2의 문장을 보내고, 실험 1에서 그 간격은 1초보다 짧았다. 그래서 S1의 `DeadLockCheck`가 3(교착)을 돌려주고 S1에 `deadlock detected`가 오며, 보고를 부른 자리는 lock.c 1235행이다. S1의 번호는 ROLLBACK이 아니라 그 오류 때 커밋 로그에 중단으로 적힌다 | 두 실행 모두 그렇다. S1이 기다리기 시작해 S2가 기다리기까지의 간격도 적는다 | S2가 오류를 받으면, 그 간격이 1초를 넘었는지 먼저 본다. 넘었으면 이 예상의 조건이 깨진 것이고, 넘지 않았는데 S2면 예상이 틀렸다 |
| L5 | 2-deadlock: S2는 자기 타이머가 울리기 전에 S1의 중단이 깨운다(S1 백엔드의 `ProcWakeup` 대상이 S2). S2의 UPDATE는 S1이 지운 것으로 표시했던 (0,1)을 다시 확인해 `TM_Ok`를 받고 바꾼다. 마지막 잔액은 id 1이 120, id 2가 180이다 | 모두 그렇다 | S2에서도 교착 검사가 돌거나 잔액이 다르다 |
| L6 | 2-fk-nokey: S1의 INSERT 안에서 `RI_FKey_check`가 불리고, 그 아래 `heap_lock_tuple`이 부모 (0,1)을 `LockTupleKeyShare`로 잠근다. 그 단계 뒤 부모 (0,1)의 xmax는 S1의 번호에 `HEAP_XMAX_KEYSHR_LOCK`·`HEAP_XMAX_LOCK_ONLY`다 | 모두 그렇다 | 다른 모드이거나 부모 행을 잠그지 않는다 |
| L7 | 2-fk-nokey: S2의 UPDATE는 `ProcSleep`에 들어가지 않는다. `HeapTupleSatisfiesUpdate`는 `TM_BeingModified`를 돌려주지만 기다리지 않고, `MultiXactIdCreate(S1, ForKeyShare, S2, NoKeyUpdate)`를 부른다. 옛 버전 (0,1)의 xmax는 그 MultiXact 번호(`HEAP_XMAX_IS_MULTI`)이고, 새 버전 (0,2)의 xmax는 S1의 번호에 key share 잠금만이다 | 모두 그렇다 | S2가 기다리거나 MultiXact가 생기지 않는다 |
| L8 | 2-fk-delete: S2는 기다리기 전에 번호를 받고, `heap_delete` 안에서 `XactLockTableWait(S1, Delete)`로 기다린다. S1의 COMMIT 뒤 행을 지우고, 외래 키 검사가 `update or delete on table "parent" violates foreign key constraint`를 낸다(ri_triggers.c 2801행). S2의 번호는 중단으로 적히고 부모 행은 남는다 | 모두 그렇다 | 기다리지 않거나 오류가 다르거나 부모 행이 사라진다 |
| L9 | 2-idem-commit: S2는 힙에 자기 행을 먼저 넣고(그 단계 동안 페이지에 S2의 번호를 xmin으로 가진 줄이 있다), `_bt_doinsert`에서 `XactLockTableWait(S1, InsertIndex)`로 기다린다. S1의 COMMIT 뒤 `duplicate key value violates unique constraint`(nbtinsert.c 668행)가 오고, S2의 줄은 죽은 버전으로 페이지에 남는다 | 모두 그렇다 | 힙에 넣기 전에 기다리거나 오류가 다르다 |
| L10 | 2-idem-rollback: L9와 같이 기다린 뒤 S1의 중단이 깨우고, S2의 INSERT가 `INSERT 0 1`로 끝난다. 마지막 SELECT에는 S2의 행 하나만 보인다 | 모두 그렇다 | S2가 오류를 받는다 |
| L11 | 2-idem-onconflict: S2는 execIndexing.c 896행에서 `XactLockTableWait(S1, InsertIndex)`로 기다리고, S1의 COMMIT 뒤 `INSERT 0 0`으로 끝난다. 힙에 S2의 줄이 생기지 않고 S2는 번호를 받지 않는다 | 모두 그렇다 | S2의 줄이 생기거나 번호를 받는다 |
| L12 | 기록의 짝짓기와 판정 대조: 실험 1의 E4·E5와 같은 값이 모든 시나리오 세션 백엔드에서 같다. E5는 관찰하는 표마다 센다. 결과를 보는 SELECT는 인덱스가 없는 열(`owner = 'kim'`)이나 조건 없이 써서 순차 스캔으로 모든 줄을 판정하게 한다. `id = 1`이면 기본 키 인덱스로 HOT 사슬만 판정할 수 있어 E5의 셈이 뜻을 잃는다 | L3의 한 단계 말고 모두 같다 | 하나라도 다르다 |
| L13 | 공유 버퍼의 행은 모든 백엔드에서 같은 가상 주소다(postmaster가 만든 공유 메모리를 fork로 물려받음). 같은 ctid에 대한 `HeapTupleSatisfiesMVCC`의 `t_data` 주소가 세션끼리 같고, 단계 사이에 새로 켜진 힌트 비트마다 같은 주소에 쓴 `SetHintBits`가 있다 | 둘 다 그렇다 | 주소가 다르면 이 가정이 틀린 것이고 E8의 값 맞추기로 돌아간다 |

## 실험 3 사전 등록(22:54, 실행 전): 취소와 커밋, 그리고 서버가 죽었을 때

질문은 셋이다. ROLLBACK은 무엇을 지우고 무엇을 남기는가. COMMIT 응답을 받은 트랜잭션은 언제부터 서버가 죽어도 살아남는가. `synchronous_commit = off`에서 잃는다는 것은 정확히 무엇을 잃는가. 서버가 죽는 경우는 세션 백엔드 하나를 `kill -9`로 죽여 일으킨다. postmaster는 자식이 신호로 죽으면 공유 메모리가 망가졌을 수 있다고 보고 다른 자식을 모두 내린 뒤 복구부터 다시 시작한다. 그래서 백엔드 하나를 죽이면 서버 전체가 죽었다 살아나는 일(crash restart)을 볼 수 있다. 운영 체제나 전원이 죽는 경우는 이 실험대에서 일으킬 수 없다(아래 한계).

### 시나리오

| 이름 | 준비 | 순서 | 보려는 것 |
| --- | --- | --- | --- |
| 3-rollback | `acct`에 `(1,'kim',100)` | S1 BEGIN → S1 `INSERT (2,'lee',200)` → S1 `UPDATE … SET balance = 50 WHERE owner = 'kim'` → S1 ROLLBACK → S2 SELECT(조건 없음) | 취소가 페이지를 되돌리지 않고 커밋 로그만 바꾸는 것, 다음 읽기가 힌트 비트를 쓰는 것, 취소는 WAL을 디스크에 내리지 않는 것 |
| 3-crash-sync-on | 같음 | S1 `INSERT (2,'lee',200)`(자동 커밋) → `kill: S1` → S2 SELECT → S2 `INSERT (3,'park',300)` → S2 SELECT | 동기 커밋 뒤 죽으면 복구가 그 행을 되살림 |
| 3-crash-sync-off | 같음, `ALTER SYSTEM SET wal_writer_delay = '10s'` | S1 `SET synchronous_commit = off` → S1 `INSERT (2,'lee',200)` → `kill: S1` → S2 SELECT → S2 INSERT → S2 SELECT | 비동기 커밋 응답 뒤, WAL을 쓰기 전에 죽으면 그 행을 잃음 |
| 3-crash-sync-off-wait | 같음 | 위와 같고 `kill` 앞에 `pause: 11s` | WAL writer가 10초 주기로 쓴 뒤에 죽으면 살아남음 |

S2는 `kill` 뒤에 처음 쓰이므로 서버가 다시 뜬 뒤에 접속한다. 결과를 보는 SELECT는 모두 조건 없이 쓴다(실험 2의 교훈).

### 근거로 읽은 소스(18.6)

- 커밋: `RecordTransactionCommit`은 WAL을 썼고 `synchronous_commit`이 off보다 크면 `XLogFlush(XactLastRecEnd)`로 커밋 레코드까지 디스크에 내리고(xact.c 1502행), 그 뒤에 커밋 로그를 적는다(1508행). off이면 `XLogSetAsyncXactLSN(XactLastRecEnd)`로 “여기까지 내려야 함”만 알리고(1523행), 커밋 로그는 `TransactionIdAsyncCommitTree`로 그 LSN과 함께 적는다(1531행). 1515~1516행 주석은 이 경우 postmaster가 죽으면 WAL 버퍼가 쓰이지 않은 채 남아 커밋한 트랜잭션을 잃을 수 있다고 적는다.
- WAL writer: `XLogSetAsyncXactLSN`은 WAL writer가 겨울잠 중이거나 내릴 블록이 `wal_writer_flush_after`(기본 1 MB) 이상일 때만 깨운다(xlog.c 2636~2648행). 그 밖에는 WAL writer가 `wal_writer_delay`마다 깨어 `XLogBackgroundFlush`를 부른다(walwriter.c 250·263·268행). 겨울잠은 할 일 없는 순환이 50번 이어진 뒤라(78행), 10초 주기에서는 이 실험 동안 들지 않는다. `XLogBackgroundFlush`는 비동기 커밋 LSN까지 쓰고(xlog.c 2997~3000행), 마지막 내림이 `wal_writer_delay`보다 오래됐으면 내린다(3041행).
- 취소: `RecordTransactionAbort`는 WAL을 내리지 않는다. 주석의 이유는 “죽은 뒤에는 어차피 취소로 보기 때문”이다(xact.c 1784행). `XLogSetAsyncXactLSN`(1841행) 뒤 `TransactionIdAbortTree`(1851행)를 부른다.
- 서버가 죽음: postmaster는 자식이 신호로 끝나면 `HandleChildCrash` → `HandleFatalError`(postmaster.c 2772·2686행)에서 남은 자식에게 `SIGQUIT`을 보내고(2706행), 모두 끝나면 “all server processes terminated; reinitializing”(3187행) 뒤 startup 프로세스로 복구를 시작한다. 복구는 `PerformWalRecovery`가 “redo starts at”(xlogrecovery.c 1768행)부터 레코드마다 `rm_redo`(2020행)를 부르고 “redo done at”(1903행)으로 끝낸다. 크래시 복구 끝에는 체크포인트를 요청하고 기다린다(xlog.c 6359행). 커밋 레코드를 다시 실행하면 다음 번호를 그 번호 뒤로 옮긴다(xact.c 6143행).
- 체크섬과 힌트 비트: 이 실험대의 데이터 디렉터리는 데이터 체크섬이 켜져 있다(`pg_controldata`: Data page checksum version 1, PostgreSQL 18의 initdb 기본값). 그러면 깨끗한 페이지에 힌트 비트만 쓸 때 `MarkBufferDirtyHint`가 `XLogSaveBufferForHint`를 부른다(bufmgr.c 5542행). 체크포인트 뒤 처음이면 페이지 전체를 WAL에 남긴다.
- 미래 번호: `pg_xact_status`는 다음에 줄 번호 이상이면 오류다(xid8funcs.c 116~120행). 비동기 커밋을 잃으면 다음 번호가 되돌아가므로, 관찰은 스냅숏 xmax보다 작은 번호만 묻는다. (23:21 고침: 스냅숏 xmax는 다음에 줄 번호가 아니어서 이 결론은 틀렸다. 아래 「실험 3 기록」의 벗어난 것 1)

### 방법의 변경

- 시나리오: `kill: S1` 줄은 그 세션의 백엔드에 SIGKILL을 보내고, 단계로 세지 않으며 steps.json의 `kills`에 남긴다. 그 뒤 서버가 접속을 받을 때까지(최대 60초) 기다려 관찰용 접속을 다시 열고 관찰 하나를 찍는다. `kill` 뒤에 처음 쓰이는 세션은 그때 접속한다. 관찰용 접속이 바뀌므로 steps.json에 `observerPids`를 둔다. 준비 SQL은 한 문장씩 보낸다(`ALTER SYSTEM`은 트랜잭션 블록 안에서 돌 수 없음).
- 관찰: WAL의 삽입·쓰기·내림 위치(`pg_current_wal_insert_lsn`, `pg_current_wal_lsn`, `pg_current_wal_flush_lsn`)를 더한다. 커밋 로그 상태는 스냅숏 xmax보다 작은 번호만 묻는다(23:21에 없앰, 「실험 3 기록」의 벗어난 것 1).
- uprobe: `XLogFlush`(내릴 LSN), `XLogSetAsyncXactLSN`, `TransactionIdAsyncCommitTree`(번호, LSN), `XLogSaveBufferForHint`(반환 LSN), `XLogBackgroundFlush`의 반환값, `heap_redo`·`xact_redo`·`xlog_redo`(레코드 LSN, 번호, info).
- 해석기: 세션이 아닌 프로세스(postmaster, startup, WAL writer, checkpointer)를 그 함수 기록의 main 함수로 가리고, 그 프로세스의 `kill`, pg_wal 쓰기와 동기화, 다시 실행한 레코드를 서버 사건으로 낸다. server.log의 재시작 줄도 시각과 함께 낸다.
- 화면: `kill`에서 서버가 다시 뜰 때까지를 장면 하나로 그리고, 서버 프로세스 생명선과 WAL 위치를 더한다.

실행 전 고침(23:15, 도구를 만들며). 첫 기록 전에 바꾼 것과 그 이유다.

- 복구의 관측: `heap_redo`·`xact_redo`·`xlog_redo` 셋 대신 `ApplyWalRecord`(xlogrecovery.c 1937행) 하나에 건다. 복구가 다시 실행하는 레코드마다 한 번 불리므로, 셋만 보면 빠지는 Btree·Standby 같은 자원 관리자의 레코드도 같은 꼴로 남는다. 레코드의 시작·끝 위치, 자원 관리자 번호, 트랜잭션 번호, info, 길이를 읽는다(DWARF로 확인한 자리: `XLogReaderState`의 `ReadRecPtr` +40·`EndRecPtr` +48, `XLogRecord`의 `xl_tot_len` +0·`xl_xid` +4·`xl_info` +16·`xl_rmid` +17).
- 크래시 복구가 커밋 레코드를 다시 실행하면 `xact_redo_commit`이 `TransactionIdCommitTree`를 부른다(standbyState가 STANDBY_DISABLED, xact.c 6157~6162행). 그래서 D7에 “새 startup 프로세스가 `TransactionIdCommitTree(753)`를 부른다”를, D9에 “부르지 않는다”를 더한다. 대기 서버의 복구만 `TransactionIdAsyncCommitTree`를 부른다(6186행).
- kill 뒤의 순서: SIGKILL → server.log의 “database system is ready to accept connections” 줄 수가 kill 앞보다 늘 때까지 50 ms마다 봄(최대 60초) → 이 구간의 psql을 끝냄 → 관찰용 접속을 새로 엶 → 관찰 → 다음 구간의 세션 접속. pg_isready로 보지 않는 것은 postmaster가 자식의 죽음을 처리하기 전 잠깐 동안에도 접속이 될 수 있어서다. 서버가 다시 뜨기 전에 psql을 끝내지 않는 것은 `\q`가 백엔드에 종료 메시지를 보내, postmaster가 내리기 전에 그 백엔드가 스스로 끝나게 할 수 있어서다. 재시작 뒤 관찰을 세션 접속보다 먼저 찍는 것은 어느 클라이언트도 페이지를 건드리기 전의 상태를 남기려는 것이다.
- kill의 제약: 이 구간의 세션에 끝나지 않은 단계가 있으면 죽이지 않고 오류로 멈춘다(그 단계의 끝과 출력이 정해지지 않음). kill 앞에서 쓴 세션은 그 뒤에 쓸 수 없다(재시작이 모든 접속을 끊음). 둘 다 시나리오를 읽을 때 막는다.
- 커밋 로그: 스냅숏 xmax 이상의 번호는 묻지 않고 `future`로 남긴다(23:21에 없앰, 「실험 3 기록」의 벗어난 것 1). WAL 위치는 스냅숏과 한 질의로 읽으므로, 관찰의 읽은 구간으로는 스냅숏과 같은 부분이다.
- 판정 보강: D4·D8·D10의 “WAL writer가 썼다”를 시각만으로 보지 않고, pg_wal 파일에 쓴 `pwrite64`의 범위(파일 안 위치와 길이)가 그 LSN을 덮는지로 본다. WAL 파일 하나가 16 MB이면 LSN이 든 파일은 LSN ÷ 16 MB 번째이고 그 안 위치는 나머지다(파일 크기는 기록에서 확인).
- 알려진 위험(바꾸지 않음): `XLogBackgroundFlush`는 다 찬 WAL 페이지가 아직 내려지지 않았으면 그 경계까지만 쓰고, 비동기 커밋 위치는 다음 주기로 미룬다(xlog.c 2992~3002행, 2956행 주석의 “최대 세 주기”). INSERT의 레코드가 8 KB 페이지 경계를 넘으면 D10의 11초 안에 쓰이지 않을 수 있다. 그때는 `pwrite64` 범위로 이 경우인지 가린다.
- 준비 SQL: 사전 등록은 “한 문장씩 보낸다”였지만, 도구 시험으로 뜬 3-crash-sync-on 한 벌(23:14, 판정에 쓰지 않음)에서 준비가 트랜잭션 번호를 셋(752~754) 써서 S1이 753이 아니라 755를 받았다. 그래서 예전처럼 준비를 한 트랜잭션으로 묶고, 트랜잭션 블록 안에서 돌 수 없는 `ALTER SYSTEM` 줄만 따로 보낸다(txlab `setupBatches`). 실험 1·2와 같은 번호를 쓰고, D의 753·754가 그대로 S1·S2의 번호를 가리키게 하려는 것이다. 같은 시험 기록에서 크래시 경로는 예상대로 돌았다: S1에 SIGKILL을 보낸 뒤 0.2 ms 안에 S1이 `killed by SIGKILL`로 끝나고, postmaster가 남은 자식 열에게 차례로 SIGQUIT을 보내 모두 `exited with 2`로 끝났으며, kill에서 server.log의 접속 재개 줄까지 0.81초였다.
- 시각 표지: 실험 1과 3의 사전 등록 표지(20:40, 23:20)가 그 절을 커밋한 시각(20:27, 22:54)보다 늦게 적혀 있어 커밋 시각으로 고쳤다. 두 절 모두 실행 전에 커밋됐다.

추가 사전 등록(23:31, D10이 두 실행 모두 예상과 다르게 나온 뒤, 이 시나리오의 실행 전). D10의 두 실행에서 옛 WAL writer는 비동기 커밋 뒤 약 6.1초에 `0/1782000~0/179C000`만 쓰고 내렸고, 비동기 커밋 레코드의 끝 `0/179D330`은 그 범위 밖이었다. 복구는 `0/179C000`에서 멈췄다. 설명은 위 “알려진 위험”이다. S1의 INSERT가 남긴 첫 레코드(카탈로그 페이지의 힌트 비트 페이지 이미지, 7,913바이트)가 `0/179B2D0`에서 시작해 8 KB 경계 `0/179C000`을 넘었고, `XLogBackgroundFlush`의 첫 주기는 다 찬 페이지 경계까지만 쓴다. 이 설명이 맞으면 한 주기를 더 기다릴 때 커밋이 살아남아야 한다.

| 이름 | 준비 | 순서 |
| --- | --- | --- |
| 3-crash-sync-off-wait2 | 3-crash-sync-off와 같음 | 3-crash-sync-off-wait와 같고 쉼만 21초(WAL writer 두 주기) |

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| D13 | 쉬는 동안 옛 WAL writer가 두 번 쓴다. 첫 번째 `pwrite64`는 `0/179C000`에서 끝나고, 약 10초 뒤 두 번째 `pwrite64`는 `0/179D330`을 덮으며 그 뒤 `fdatasync`가 온다 | 두 실행 모두 그렇다 | 두 번째 쓰기가 없거나 `0/179D330`을 덮지 않는다 |
| D14 | 복구가 `0/179C000`을 지나 753의 Heap INSERT와 COMMIT을 다시 실행하고 startup이 `TransactionIdCommitTree(753)`를 부른다. 재시작 뒤 S2의 SELECT에 (0,2) lee 200이 보이고 S2의 INSERT는 754를 받는다 | 두 실행 모두 그렇다 | 행을 잃는다 |


### 예상과 판정

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| D1 | 3-rollback: S1의 ROLLBACK 요청 안에서 `XLogSetAsyncXactLSN`과 `TransactionIdAbortTree(753)`가 불리고, 그 요청에는 pg_wal의 `fdatasync`도 `pwrite64`도 없다 | 두 실행 모두 그렇다 | ROLLBACK이 WAL을 내린다 |
| D2 | 3-rollback: ROLLBACK 뒤 관찰의 페이지는 그 앞 관찰과 줄·xmin·xmax·ctid가 같다((0,1) xmax 753, (0,2)·(0,3) xmin 753). 다른 것은 커밋 로그의 753(중단)뿐이다 | 모두 그렇다 | 줄이 사라지거나 xmin·xmax가 바뀐다 |
| D3 | 3-rollback: S2의 SELECT는 세 줄을 판정해 (0,1)만 보고, 판정 안에서 (0,1)에 `HEAP_XMAX_INVALID`, (0,2)·(0,3)에 `HEAP_XMIN_INVALID`를 쓴다. 이 페이지는 S1이 바꾼 뒤 아직 디스크에 쓰이지 않아 더러우므로 `XLogSaveBufferForHint`는 불리지 않는다 | 모두 그렇다 | 다른 비트이거나 페이지 이미지를 남긴다 |
| D4 | 3-rollback: S1이 쓴 WAL(취소 레코드 포함)은 WAL writer가 쓴다. ROLLBACK 뒤 300 ms 안에 WAL writer의 pg_wal `pwrite64`가 있다(주기 200 ms, 겨울잠이면 `XLogSetAsyncXactLSN`이 바로 깨움) | 모두 그렇다 | S1이 쓰거나 300 ms 안에 아무도 쓰지 않는다 |
| D5 | 3-crash-sync-on: S1의 INSERT 요청 안에서 `XLogFlush(커밋 레코드 끝)` → pg_wal `fdatasync` → `TransactionIdCommitTree(753)` 순서이고, 그 뒤 관찰에서 삽입 위치와 내림 위치가 같다 | 모두 그렇다 | 순서가 다르거나 내림이 삽입보다 뒤처진다 |
| D6 | 크래시 셋: S1이 죽은 뒤 postmaster가 남은 자식에게 `SIGQUIT`을 보내고, server.log에 “terminated by signal 9”, “all server processes terminated; reinitializing”, “redo starts at”, “redo done at”, “database system is ready to accept connections”가 이 순서로 나온다. 새 startup 프로세스가 `rm_redo`로 레코드를 다시 실행한다 | 여섯 실행 모두 그렇다 | 재시작하지 않거나 순서가 다르다 |
| D7 | 3-crash-sync-on: 다시 실행한 레코드에 753의 heap INSERT와 COMMIT이 있다. 재시작 뒤 S2의 SELECT에 (0,2) lee 200이 보이고, S2의 INSERT는 754를 받는다 | 모두 그렇다 | 행이 없거나 번호가 다르다 |
| D8 | 3-crash-sync-off: S1의 INSERT 요청 안에는 `XLogFlush`도 pg_wal `fdatasync`도 없고, `XLogSetAsyncXactLSN(커밋 레코드 끝)`과 `TransactionIdAsyncCommitTree(753, 같은 LSN)`가 있다. 응답은 `INSERT 0 1`이다. 그 뒤 관찰에서 내림 위치가 삽입 위치보다 뒤처지고, `kill`까지 WAL writer의 pg_wal 쓰기가 없다 | 모두 그렇다 | S1이 내리거나 WAL writer가 kill 전에 쓴다(그러면 D9의 조건이 깨진 것) |
| D9 | 3-crash-sync-off: 다시 실행한 레코드에 753이 없다. 재시작 뒤 S2의 SELECT에는 (0,1)만 보인다. 응답까지 받은 커밋을 잃었다. S2의 INSERT는 753을 다시 받고 (0,2)에 들어간다 | 모두 그렇다 | 행이 살아 있다. 그때는 WAL writer가 kill 전에 썼는지 먼저 본다 |
| D10 | 3-crash-sync-off-wait: 쉬는 11초 사이에 WAL writer가 `XLogBackgroundFlush`로 pg_wal에 쓰고 내린다(커밋 뒤 10초 안). 재시작 뒤 (0,2)가 보이고 S2의 INSERT는 754를 받는다 | 모두 그렇다 | 쓰지 않거나 행을 잃는다 |
| D11 | 크래시 셋: 재시작 뒤 S2의 첫 SELECT가 acct 페이지에 처음 힌트 비트를 쓸 때 `XLogSaveBufferForHint`가 한 번 불리고 0이 아닌 LSN을 돌려준다(복구 끝 체크포인트로 페이지가 깨끗함). 같은 페이지의 다음 힌트는 부르지 않는다 | 모두 그렇다 | 부르지 않거나 힌트마다 부른다 |
| D12 | 기록: 죽은 S1 백엔드도 짝지은 묶음마다 uprobe 사건 수와 함수 기록 수가 같다(기록 함수는 사건마다 `write()` 한 번이라 죽어도 그때까지는 남음, pgtrace_hook.c 142·150행). 나머지 세션도 실험 1·2의 E4·E5와 같다 | 모두 그렇다 | 다르면 범위와 이유를 적는다 |

한계: 운영 체제가 죽거나 전원이 나가는 경우(쓰기만 하고 내리지 않은 WAL을 잃는 경우)는 이 실험대에서 일으키지 않는다. D9와 D10의 차이는 “WAL을 운영 체제에 썼는가”이고, “디스크에 내렸는가”는 이 실험으로 가를 수 없다.

## 실험 4 사전 등록(00:35, 실행 전): SERIALIZABLE과 write skew

질문은 셋이다. REPEATABLE READ는 스냅숏 하나로 읽는데도 왜 write skew를 막지 못하는가. SERIALIZABLE은 무엇을 기억해 두었다가 어느 순간 누구를 실패시키는가. 그 기억(SIRead 락)은 얼마나 굵고 언제 사라지는가.

write skew는 두 트랜잭션이 같은 사실(당직 의사가 몇 명인가)을 읽고, 그 사실에 기대어 서로 **다른** 행을 바꾸는 일이다. 바꾸는 행이 다르므로 행 락으로는 부딪히지 않는다(실험 2의 기다림은 같은 행을 바꿀 때만 생겼다). 둘 다 커밋하면 어느 한쪽이 먼저 끝까지 돌고 다른 쪽이 그 뒤에 돈 경우(직렬 순서)에는 나올 수 없는 결과, 당직 0명이 남는다. 직렬 순서라면 뒤에 돈 쪽이 당직 1명을 보고 빠지지 않았을 것이다.

PostgreSQL의 SERIALIZABLE은 SSI(Serializable Snapshot Isolation)다. 읽는 쪽을 기다리게 하지 않고, 읽은 자리를 SIRead 락(술어 락, predicate lock)으로 남겨 둔다. 다른 트랜잭션이 그 자리에 쓰면 “읽은 쪽 → 쓴 쪽”이라는 rw 의존 간선을 긋는다. 한 트랜잭션에 이 간선이 들어오고 나가서 가운데(pivot)가 되고, 바깥쪽이 먼저 커밋하면 가운데를 실패시킨다(predicate.c 4525~4528행 주석). 기다림이 아니라 나중의 실패로 막는 것이 2단계 락과 다르다.

### 시나리오

| 이름 | 준비 | 순서 | 보려는 것 |
| --- | --- | --- | --- |
| 4-skew-rr | `doctors(name text PRIMARY KEY, on_call bool)`에 `('alice', true)`, `('bob', true)` | S1 `BEGIN ISOLATION LEVEL REPEATABLE READ` → S1 `SELECT count(*) FROM doctors WHERE on_call` → S2 BEGIN(같음) → S2 같은 SELECT → S1 `UPDATE doctors SET on_call = false WHERE name = 'alice'` → S2 같은 UPDATE(bob) → S1 COMMIT → S2 COMMIT → S1 SELECT(조건 없음) | 둘 다 커밋하고 당직 0명. SSI의 함수는 입구에서 돌아온다 |
| 4-skew-ssi | 같음 | 위와 같고 `SERIALIZABLE` | 순차 스캔이 잡는 표 단위 SIRead 락, 두 UPDATE가 긋는 간선 둘, S1의 COMMIT이 S2를 실패로 표시하고 S2의 COMMIT이 실패함, 커밋 뒤에도 남는 SIRead 락 |
| 4-skew-ssi-row | 같음 | `SERIALIZABLE`, 세는 SELECT 대신 S1은 `SELECT on_call FROM doctors WHERE name = 'bob'`, S2는 `WHERE name = 'alice'` | 행 단위 SIRead 락으로도 같은 간선 둘, 같은 실패 |
| 4-ssi-own | 같음 | `SERIALIZABLE`, S1은 `WHERE name = 'alice'`, S2는 `WHERE name = 'bob'`(자기 행) | 상대가 쓴 것을 읽지 않으면 간선이 없고 둘 다 커밋. 결과는 4-skew-rr처럼 당직 0명이지만, 이번에는 직렬 순서로 돌려도 같은 결과다(아무도 상대의 당직을 확인하지 않음) |

두 UPDATE는 다른 행이라 기다리지 않는다. 트랜잭션 번호는 준비 752, S1의 UPDATE 753, S2의 UPDATE 754다. 결과를 보는 SELECT는 조건 없이 쓴다(실험 2의 교훈). 기본 키로 한 행을 찾는 문장은 인덱스 검색이 된다고 본다. 실험 2에서 같은 꼴의 새 표에 `ORDER BY id`만 붙여도 계획기가 기본 키 인덱스를 골랐다.

### 근거로 읽은 소스(18.6)

- 스냅숏: SERIALIZABLE은 트랜잭션 스냅숏을 `GetSerializableTransactionSnapshot`(snapmgr.c 319행)으로 찍고, 그 안의 `GetSerializableTransactionSnapshotInt`가 이 트랜잭션의 `SERIALIZABLEXACT`를 만들어 pid를 적는다(predicate.c 1862행). RR은 이 길을 지나지 않는다. 그래서 RR에서는 `MySerializableXact`가 비어 있고, 술어 락 함수는 입구의 검사(`SerializationNeededForRead` 516행, `SerializationNeededForWrite` 560행)에서 돌아온다.
- 읽기가 잡는 SIRead 락: 순차 스캔은 `initscan`에서(실제로는 `heap_beginscan`, 「실행 전 고침」과 「실험 4 기록」의 W3) 표 전체를 잡는다(`PredicateLockRelation`, heapam.c 1171행). 기본 키 검색은 B-tree 잎 페이지(`PredicateLockPage`, nbtsearch.c 1691행)와 찾은 행(`PredicateLockTID`, heapam.c 1813행)을 잡는다. 새 락을 실제로 만드는 것은 `PredicateLockAcquire` → `CreatePredicateLock`이고, 같은 락이나 더 굵은 락(표 > 페이지 > 행)이 이미 있으면 만들지 않는다(predicate.c 2524·2527행). `PredicateLockTID`는 표 단위 락이 있으면 그 앞에서 돌아온다(2648행). 카탈로그 표는 잡지 않는다(498행). 한 페이지의 행 락이 `max_pred_locks_per_page`(기본 2)를 넘으면 페이지 락으로 올리는데(2355~2356행), 이 실험의 행 락은 한 페이지에 둘까지다.
- 쓰기의 검사: `heap_update`가 옛 행에 대해 `CheckForSerializableConflictIn(표, 행, 페이지)`를 부르고(heapam.c 4107행), 이것이 행 → 페이지 → 표 순서로 `CheckTargetForConflictsIn`을 부른다(predicate.c 4373·4382·4388행). 그 자리에 다른 트랜잭션의 SIRead 락이 있으면 `FlagRWConflict(읽은 쪽, 쓴 쪽)`으로 간선을 긋는다(4244행). 행 검사에서 자기 SIRead 락을 찾으면 지운다(4221행에서 기억하고 4262행 이후 지움).
- 간선을 그을 때: `FlagRWConflict`가 먼저 부르는 `OnConflict_CheckForSerializationFailure`(4536행)는 이 간선으로 위험한 구조가 생기고 바깥쪽이 이미 커밋했거나 준비됐을 때만 실패시킨다. 두 UPDATE 때에는 아무도 커밋하지 않았으므로 실패가 없다.
- 커밋 때: `CommitTransaction`이 `PreCommit_CheckForSerializationFailure`를 부른다(xact.c 2339행, RR에서도 불리지만 `MySerializableXact`가 비어 바로 돌아옴). 먼저 자기가 실패로 표시됐는지 보고(predicate.c 4719행), 표시됐으면 4723행에서 오류를 낸다. 아니면 자기에게 들어오는 간선의 상대(가운데)가 아직 커밋하지 않았고 그 상대에게 들어오는 간선이 자기에게서 나온 것일 때 그 상대에 `SXACT_FLAG_DOOMED`(0x08)를 켠다(4730~4765행). 커밋하려는 쪽을 살리는 이유를 주석은 “다른 쪽을 되돌려야 다시 시도할 때 진전이 있다”고 적는다(4697행).
- 오류: 이유 코드는 “Canceled on identification as a pivot, during commit attempt.”(4726행)다.
- SIRead 락의 수명: 커밋한 트랜잭션의 SIRead 락은 커밋 뒤에도 남는다. 끝난 직렬화 트랜잭션 목록에 들어가고(3656행), 겹친 직렬화 트랜잭션 가운데 가장 오래된 xmin을 가진 마지막 것이 끝날 때 `ClearOldPredicateLocks`가 지운다(3636~3673행). 커밋한 트랜잭션의 SIRead 락도 pg_locks에 그 pid로 보인다(lockfuncs.c 423행, pid는 2단계 커밋의 준비에서만 지움, predicate.c 4866행).
- HOT: `on_call`은 인덱스가 걸린 열이 아니고 페이지에 자리가 있으므로 UPDATE는 HOT이다. 인덱스에 새 항목을 넣지 않으므로 B-tree 잎 페이지의 SIRead 락은 검사되지 않는다(`_bt_doinsert`의 `CheckForSerializableConflictIn`, nbtinsert.c 251행이 불리지 않음).

### 방법의 변경

- uprobe: `GetSerializableTransactionSnapshotInt`, `PredicateLockRelation`(표), `PredicateLockPage`(표, 블록), `PredicateLockTID`(표, 블록, 줄, 행의 번호), `CreatePredicateLock`(대상의 표·블록·줄, 주인 pid), `CheckForSerializableConflictIn`(표, 블록), `CheckTargetForConflictsIn`(대상), `CheckForSerializableConflictOut`(표, 번호), `FlagRWConflict`(읽은 쪽과 쓴 쪽의 pid·번호·플래그), `PreCommit_CheckForSerializationFailure`(자기 플래그). 기록 빌드의 DWARF에서 모두 실제 함수로 찾았다(`txlab uprobes`, 이 절을 쓰기 전).
- txlab: 관찰할 표의 인덱스 이름과 번호를 함께 남긴다. B-tree 페이지의 SIRead 락이 어느 인덱스의 것인지 가리려는 것이다.
- 해석기: 위 uprobe를 요청의 사건으로 내고, 술어 락 요청마다 그 안에서 `CreatePredicateLock`이 불렸는지(새로 만들었는지), 대상 검사마다 그 안에서 `FlagRWConflict`가 불렸는지를 함수 기록으로 붙인다. 함수 기록의 `ClearOldPredicateLocks`도 사건으로 낸다.
- 화면: SSI 레인(SIRead 락과 간선), 세션 사이의 간선 그림, 미니맵의 SIRead 락 상자. 실행 뒤에 만든다.

실행 전 고침(00:42, 도구 시험 한 벌 뒤). 4-skew-ssi를 한 벌 기록해 도구를 시험했다(00:41, 판정에 쓰지 않음). 고친 것과 시험에서 본 것이다.

- txlab의 인덱스 질의가 `ORDER BY indexrelid is ambiguous`로 실패했다(바꾼 이름의 열과 원래 열이 같은 이름). `ORDER BY 1`로 고쳤다.
- 커밋 검사의 uprobe가 SERIALIZABLEXACT가 없는 커밋(세션 시작 때, READ COMMITTED의 마지막 SELECT)에서도 `flags=0`을 남겼다. 포인터가 0이면 `+156(…)` 읽기는 실패하는데, 숫자 인자는 실패해도 `(fault)`가 아니라 0으로 남는다. 그래서 포인터 `sx=@MySerializableXact`를 함께 읽고, 해석기는 `sx`가 0이면 `sxact: none`으로 낸다. W2의 “입구의 검사에서 돌아온다”는 이 값으로 본다.
- 사전 등록의 함수 이름 하나가 틀렸다. heapam.c 1171행은 `initscan`이 아니라 `heap_beginscan` 안이다(시험의 스택: `table_beginscan` → `heap_beginscan` → `PredicateLockRelation`). W3은 고치지 않고 판정에 주석으로 단다.
- 시험에서 본 다른 것(예상은 바꾸지 않음): S2의 COMMIT이 입구에서 DOOMED를 보고 predicate.c 4723행에서 실패했고, 그 뒤 `ClearOldPredicateLocks`가 불렸으며, 관찰의 SIRead 락은 S2의 UPDATE 뒤 넷, S1의 COMMIT 뒤에도 넷, S2의 실패 뒤 0개였다. UPDATE 안의 `PredicateLockTID`는 한 번이 아니라 두 번 불렸다(인덱스로 찾을 때 `heap_hot_search_buffer`, 바꿀 행을 다시 읽을 때 `heap_fetch`). 둘 다 표 단위 락이 있어 만들지 않았다.

### 예상과 판정

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| W1 | 4-skew-rr: 두 SELECT 모두 2, 두 UPDATE 모두 `UPDATE 1`, 두 COMMIT 모두 `COMMIT`이다. 마지막 SELECT에 (0,3) alice f(xmin 753), (0,4) bob f(xmin 754)가 보인다. 당직이 0명이다 | 두 실행 모두 그렇다 | 어느 쪽이든 오류를 받거나 당직이 남는다 |
| W2 | 4-skew-rr: `GetSerializableTransactionSnapshotInt`, `CreatePredicateLock`, `CheckTargetForConflictsIn`, `FlagRWConflict`가 한 번도 불리지 않는다. `PredicateLockRelation`(순차 스캔마다), `CheckForSerializableConflictIn`(UPDATE마다), `PreCommit_CheckForSerializationFailure`(커밋마다)는 불리지만 입구의 검사에서 돌아온다. 관찰의 pg_locks에 `SIReadLock`이 없다 | 모두 그렇다 | 하나라도 불리거나 `SIReadLock`이 보인다 |
| W3 | 4-skew-ssi: 각 세션의 첫 SELECT 안에서 `GetSerializableTransactionSnapshotInt`가 한 번 불리고, `initscan` 아래 `PredicateLockRelation(doctors)` → `CreatePredicateLock(표 doctors)`가 한 번 불린다(주인 pid가 그 세션). S2의 SELECT 뒤 관찰에 표 doctors의 `SIReadLock`이 S1·S2 하나씩 있다 | 모두 그렇다 | 페이지나 행 단위이거나 없다 |
| W4 | 4-skew-ssi: S1의 UPDATE는 기본 키로 찾는다(스택에 `_bt_first`). 그 아래 `PredicateLockPage(doctors_pkey, 1)`이 B-tree 잎 페이지의 락을 새로 만들고, `PredicateLockTID(doctors, (0,1))`은 표 단위 락이 있어 만들지 않는다. `CheckForSerializableConflictIn(doctors, 블록 0)`이 `CheckTargetForConflictsIn`을 행·페이지·표 순으로 세 번 부르고, 표 검사에서 `FlagRWConflict(읽은 쪽 S2, 쓴 쪽 S1)`을 부른다. 쓴 쪽 번호는 753, 읽은 쪽 번호는 0(S2는 아직 쓰지 않음)이고, 오류는 없다 | 모두 그렇다 | 간선이 없거나 방향이 반대거나 UPDATE가 실패한다 |
| W5 | 4-skew-ssi: S2의 UPDATE에서 같은 길로 `FlagRWConflict(읽은 쪽 S1, 쓴 쪽 S2)`가 불린다(쓴 쪽 754, 읽은 쪽 753). 이제 간선은 S1 → S2와 S2 → S1 둘이다. 오류는 없다 | 모두 그렇다 | S2의 UPDATE가 실패하거나 간선이 없다 |
| W6 | 4-skew-ssi: S1의 COMMIT은 `PreCommit_CheckForSerializationFailure` 입구의 플래그에 DOOMED가 없고 `COMMIT`으로 끝난다(`TransactionIdCommitTree(753)`). S2의 COMMIT은 입구의 플래그에 DOOMED(0x08)가 있고 predicate.c 4723행에서 오류를 내며(psql에 위 오류와 이유 코드), `TransactionIdAbortTree(754)`가 불린다. W5의 `FlagRWConflict`가 읽은 S2의 플래그에는 DOOMED가 없다. 그 사이에 일어난 일은 S1의 COMMIT뿐이므로 그것이 S2를 표시했다고 본다(표시 자체는 함수 호출이 아니라 기록에 직접 보이지 않음, 4765행) | 모두 그렇다 | S2가 커밋하거나, S1이 실패하거나, 다른 자리에서 실패한다 |
| W7 | 4-skew-ssi: 마지막 SELECT에 alice f(xmin 753)와 bob t(xmin 752)가 보인다. 당직이 한 명 남는다 | 모두 그렇다 | 둘 다 f다 |
| W8 | 4-skew-ssi: S2의 UPDATE 뒤 관찰에 `SIReadLock`이 넷(세션마다 표 doctors와 doctors_pkey 1번 페이지)이다. S1의 COMMIT 뒤에도 S1의 둘이 그대로 있다. S2의 COMMIT(오류) 안에서 `ClearOldPredicateLocks`가 불리고, 그 뒤 관찰에 `SIReadLock`이 없다 | 모두 그렇다 | S1의 락이 커밋 때 사라지거나 마지막에 남는다 |
| W9 | 4-skew-ssi-row: 첫 SELECT마다 `PredicateLockPage(doctors_pkey, 1)`과 `PredicateLockTID`가 락을 하나씩 만든다. S1은 bob (0,2), S2는 alice (0,1)이다. 표 단위 `SIReadLock`은 생기지 않는다 | 모두 그렇다 | 표 단위 락이 생긴다 |
| W10 | 4-skew-ssi-row: S1의 UPDATE에서 `PredicateLockTID(doctors, (0,1))`이 S1의 행 락을 새로 만들고, `CheckTargetForConflictsIn`의 첫 번째(행) 검사에서 `FlagRWConflict(S2, S1)`을 부른다. 그 단계 뒤 관찰에 S1의 (0,1) 행 락은 없고(자기 락을 지움) S2의 (0,1) 행 락은 있다. S2의 UPDATE도 같은 꼴로 행 검사에서 `FlagRWConflict(S1, S2)`를 부른다. 그 뒤는 W6·W7과 같다 | 모두 그렇다 | 간선이 페이지나 표 검사에서 나오거나 결과가 다르다 |
| W11 | 4-ssi-own: 두 UPDATE 모두 `CheckTargetForConflictsIn`이 세 번 불리지만 `FlagRWConflict`는 없다. 각 UPDATE 단계 뒤 그 세션의 행 락이 없다(자기 락을 지움). 두 COMMIT 모두 `COMMIT`이고 입구의 플래그에 DOOMED가 없다. 마지막 SELECT에 alice f(753), bob f(754)가 보인다 | 모두 그렇다 | 간선이 생기거나 어느 쪽이든 실패한다 |
| W12 | SERIALIZABLE 셋: UPDATE 요청 안에 `_bt_doinsert`가 없고(HOT), `CheckForSerializableConflictIn`의 표는 늘 doctors다. `CheckForSerializableConflictOut`은 한 번도 불리지 않는다(읽은 행마다 살아 있고 보이는 행이라 `HeapCheckForSerializableConflictOut`이 그 앞에서 돌아옴, heapam.c 9533·9550~9552행) | 모두 그렇다 | 인덱스에 넣거나 conflict out 검사가 불린다 |
| W13 | 기록: 시나리오 세션 백엔드는 짝지은 묶음마다 uprobe 사건 수와 함수 기록 수가 같다(실험 3의 D12와 같음) | 모두 그렇다 | 다르면 범위와 이유를 적는다 |

한계: 세션 둘, 행 둘, 표 하나의 가장 작은 꼴이다. 읽기 전용 트랜잭션의 최적화(`READ ONLY`, `DEFERRABLE`), 세 트랜잭션 이상의 구조, 2단계 커밋, 락의 페이지·표 승격은 이번에 일으키지 않는다.

## 실험 5 사전 등록(01:57, 실행 전): Spring 전파

질문은 넷이다. `@Transactional`의 전파(propagation)는 DB에 무엇을 보내는가. 안쪽 메서드의 실패는 언제, 어느 연결에서 DB에 닿는가. REQUIRES_NEW가 바깥이 잠근 행을 건드리면 왜 서로를 기다리고, PostgreSQL은 왜 그것을 교착으로 보지 못하는가. NESTED의 저장점은 서버 안에서 무엇인가.

전파는 DB의 개념이 아니라 Spring의 개념이다. DB가 아는 것은 연결마다의 `BEGIN`, `COMMIT`, `ROLLBACK`, `SAVEPOINT`뿐이다. Spring은 지금 스레드에 묶인 연결이 있는지 보고(`TransactionSynchronizationManager`), 전파 방식에 따라 그 연결에 참여하거나, 그 연결을 떼어 두고 풀에서 다른 연결을 받거나, 그 연결에 저장점을 만든다. 그래서 같은 “안쪽 트랜잭션”이라는 말이 REQUIRED에서는 아무 바이트도 아니고, REQUIRES_NEW에서는 두 번째 연결이며, NESTED에서는 `SAVEPOINT` 한 줄이다. 이 실험은 그 차이를 앱의 호출(JFR), 연결 위의 메시지(중계기), 서버의 함수와 상태(지금까지의 기록) 세 층에서 같은 시각으로 본다.

### 시나리오

앱은 `txlab/spring/`의 Java다(Spring Framework 6.2.15, HikariCP 5.1.0, pgjdbc 42.7.13, JDK 25). Spring Boot 없이 설정 한 장(`LabConfig`)에 풀, 트랜잭션 관리자(`DataSourceTransactionManager`), `JdbcTemplate`을 세웠다. 모든 장면이 같은 꼴이다. 바깥(`Outer`, REQUIRED)이 1번 행을 고치고 멈춘 뒤 안쪽 빈(`Inner`)을 부르고, 안쪽이 2번 행(또는 같은 1번 행)을 고친다. 준비는 `acct(id int PRIMARY KEY, owner text, balance int)`에 `(1,'kim',100)`, `(2,'lee',200)`이다.

| 이름 | 풀 | 안쪽 | 보려는 것 |
| --- | --- | --- | --- |
| 5-required | 2 | REQUIRED, 2번 행을 고친 뒤 예외. 바깥이 예외를 삼키고 정상으로 끝남 | 안쪽의 실패는 DB에 아무것도 보내지 않고 표시만 남김. 바깥의 커밋이 ROLLBACK이 되고 `UnexpectedRollbackException` |
| 5-requires-new | 2 | REQUIRES_NEW, 2번 행을 고치고 정상 | 두 번째 연결, 바깥 연결은 그동안 `idle in transaction`, 안쪽이 먼저 커밋 |
| 5-requires-new-same-row | 2 | REQUIRES_NEW, `SET LOCAL lock_timeout = '3s'` 뒤 1번 행(바깥이 잠근 행)을 고침 | 안쪽 백엔드는 바깥의 트랜잭션 번호를 기다리고, 바깥 백엔드는 아무것도 기다리지 않음(앱 스레드가 안쪽을 기다림). 서버의 교착 검사는 돌지만 교착이 아니라고 봄. 3초 뒤 락 시간 초과로 둘 다 ROLLBACK |
| 5-requires-new-pool1 | 1 | REQUIRES_NEW(5-requires-new와 같은 메서드) | 두 번째 연결이 풀에 없어 서버에 닿기 전에 2초 기다리다 실패. 그동안 연결 위에 메시지가 없음 |
| 5-nested | 2 | NESTED, 2번 행을 고친 뒤 예외. 바깥이 예외를 삼킴 | 같은 연결의 `SAVEPOINT`, `ROLLBACK TO SAVEPOINT`, `RELEASE SAVEPOINT`, 서버의 하위 트랜잭션 번호. 바깥은 커밋됨 |

멈춤 자리는 `start`(풀이 다 채워진 뒤) → `outer-after-update` → 안쪽의 자리(`inner-after-update`, 같은 행 장면은 `inner-before-update`) → `outer-after-inner` → `end`(바깥 메서드가 끝난 뒤) → `closed`(Spring과 풀을 닫은 뒤)다. 예외로 건너뛰는 자리는 오지 않는다. go 뒤 1.5초 안에 다음 자리가 오지 않으면 `blocked` 관찰을 한 번 더 찍는다. 시나리오마다 두 번(a, b) 기록한다.

### 관측 경로

- 연결 위의 메시지: 실습 컨테이너는 네트워크가 없고(`network none`) 서버는 유닉스 소켓만 연다. 그래서 맥의 TCP 접속 하나마다 `docker exec -i pg-lab txlab relay`를 띄워 서버의 유닉스 소켓에 잇는다(relay.go). 맥 쪽과 컨테이너 쪽이 같은 바이트를 메시지로 잘라 각자의 시계로 적는다(wire.go). 같은 연결·방향·순번의 메시지는 같은 메시지이므로 두 시계의 차이를 메시지마다 잰다. 앱 → 서버 메시지는 (컨테이너 − 맥) = 차이 + 지연, 서버 → 앱은 차이 − 지연이라 차이는 [서버 → 앱 값의 최댓값, 앱 → 서버 값의 최솟값] 안에 있다. 01:44에 날 클라이언트로 두 번 시험했다(`TXLAB_LIVE=1 go test -run TestBridgeLive`). 두 번 모두 메시지 27개가 양쪽에 같은 순번과 길이로 남았고, 차이 구간은 [−0.131, 0.076] ms와 [−0.094, 0.074] ms였다. 서버는 이 접속을 유닉스 소켓 접속으로 본다. 메시지는 TCP로 직접 붙을 때와 같지만 TCP 자체(연결 맺기, keepalive)는 서버 쪽에서 보이지 않고, 그것은 실험 6의 몫이다.
- 백엔드와 연결의 짝: 서버가 보내는 BackendKeyData(`K`)에 백엔드 pid가 있다. 중계기가 그것을 읽어 연결 번호와 pid를 잇고, 관찰에 그 pid를 세션(C1, C2 …)으로 알린다. 앱은 pid를 몰라도 된다.
- 서버 쪽 상태: 컨테이너의 `txlab observe`가 관찰용 접속 하나로 멈춤 자리마다 실험 1~4와 같은 관찰(페이지, 세션 상태, 락, 스냅숏, WAL, 커밋 로그)을 찍는다. 01:45에 명령 꼴(session, capture, waiters, end, 모르는 명령의 오류)을 시험했다.
- 앱 쪽 호출: JFR `jdk.MethodTrace`(JDK 25)로 Spring 트랜잭션 관리자, `JdbcTemplate`, HikariCP의 빌리기·돌려주기, pgjdbc의 commit·rollback·저장점, 장면 메서드의 들어감과 나옴을 줄 번호가 붙은 스택과 함께 남긴다(필터는 spring.go의 `springTrace`). 예외가 만들어진 자리는 `jdk.JavaExceptionThrow`(`exceptions=all`), 스레드가 기다린 자리는 `jdk.ThreadPark`(문턱 1 ms)로 남긴다. Spring과 HikariCP의 DEBUG 로그도 마이크로초 시각으로 남긴다(logback `%micros`).
- 서버 쪽 함수: 지금까지의 함수 기록, strace, uprobe를 그대로 쓰고, 하위 트랜잭션의 깊이를 보려고 `AssignTransactionId`의 `nestingLevel`을 더했다(probes.txt의 `assign`, 기록 빌드에서 `+36(%x0)`으로 찾음).

### 방법의 변경과 도구 시험

- JFR 메서드 추적의 한계(01:49~01:50, 실험 전 시험). 작은 프로그램(던지는 메서드, 잡아서 다시 던지고 finally가 있는 메서드, 부른 쪽이 던진 예외가 그냥 지나가는 메서드, 정상 반환)으로 GraalVM 25.0.4와 OpenJDK 25.0.2에서 같은 결과를 봤다. 정상 반환과 메서드 안의 `throw`로 끝난 호출은 한 번씩 남는다. 다시 던지고 finally가 있는 메서드는 같은 시작 시각으로 두 번 남는다(GraalVM 27.7 ms와 35.2 ms, 실제 끝은 finally까지 돈 긴 쪽). 부른 쪽에서 올라온 예외가 그냥 지나간 호출은 **아예 남지 않는다**. 이번 장면은 예외가 핵심이라 이 빈자리가 크다. 예를 들어 5-required에서 바깥의 `TransactionInterceptor.invoke`는 커밋 단계의 `UnexpectedRollbackException`이 지나가며 끝나므로 기록에 없다. 그래서 빈자리를 셋으로 메운다. 남은 사건의 스택(호출한 쪽의 메서드와 줄이 다 있음), 예외가 만들어진 자리와 시각(`jdk.JavaExceptionThrow`), Spring 로그의 마이크로초 시각이다. 화면은 지나간 호출의 끝 시각을 그리지 않고, 해석기는 같은 메서드·스레드·시작 시각의 사건이 둘이면 긴 쪽을 쓴다. 이 한계가 문서화된 동작인지는 확인하지 않았다(JEP 520 원문 미확인).
- DB 없는 시험(01:48). 앱의 `Outer`·`Inner`를 그대로 두고 연결만 호출을 적는 가짜(`java.lang.reflect.Proxy`)로 바꿔 네 장면을 돌렸다(스크래치패드의 `trial/`, 저장소에 넣지 않음). Spring이 연결에 부른 것이 아래 소스 읽기와 줄마다 같았다. 5-required는 `setAutoCommit(false)` → UPDATE 둘(같은 연결) → `rollback()` → `setAutoCommit(true)` → `close()` 순서였다. 안쪽의 실패와 바깥의 `outer-after-inner` 사이에는 연결 호출이 없었고, 예외는 `UnexpectedRollbackException: Transaction rolled back because it has been marked as rollback-only`였다. 5-requires-new는 연결 둘이었고 안쪽의 `commit()`이 바깥의 `commit()`보다 먼저였다. 5-nested는 `getMetaData()`(저장점 지원 확인) → `setSavepoint("SAVEPOINT_1")` → UPDATE → `rollback(SAVEPOINT_1)` → `releaseSavepoint(SAVEPOINT_1)` → `commit()`이었다. 이 시험은 HikariCP와 pgjdbc를 거치지 않으므로, 아래 P 항목은 실제 실행에서 다시 판정한다.
- 기록 전체 경로 시험(01:56). `record`로 5-required를 돌려, 서버 준비와 uprobe 켜기 뒤 앱 준비에서 pgjdbc가 없다는 오류로 멈추고 tracelab 설정이 되돌아가는 것(txlab uprobe 0개, 인스턴스 없음)을 봤다. Spring·HikariCP 등 jar 14개는 `~/.m2`의 `.sha1`과 맞는 것을 확인하고 실습 자리(`.lab-tmp/spring/jars`)로 옮겼다.
- 등록 뒤 고침(02:05, 실행 전). 지휘가 `end`와 `closed` 뒤에는 멈춘 채로 찍지 않게 했다(spring.go의 `endCheckpoint`). 그 구간은 Spring과 풀을 닫는 때이고 아래 P 항목 가운데 그때의 서버 상태를 묻는 것이 없다. 풀을 닫는 데 1.5초가 넘게 걸리면 쓸모없는 관찰이 하나 끼었을 것이다.
- 첫 시험에서 볼 것(등록하지 않은 관찰로 채움): pgjdbc 42.7.13이 StartupMessage에서 요청하는 프로토콜 판(3.0 또는 3.2. 3.2면 BackendKeyData의 열쇠 길이가 달라지지만 중계기는 pid만 읽으므로 짝짓기는 같다), 풀을 닫는 데 걸리는 시간.

### 근거로 읽은 소스

Spring Framework 6.2.15(spring-tx, spring-jdbc 소스 jar)

- 들어감: 프록시의 `TransactionInterceptor.invoke`가 `TransactionAspectSupport.invokeWithinTransaction`을 부르고, 그 안의 `createTransactionIfNecessary`가 `tm.getTransaction`을 부른다(639행). 메서드 본문은 380행에서 돌고, 예외면 384행 `completeTransactionAfterThrowing`, 정상이면 416행 `commitTransactionAfterReturning`이다. 기본 규칙은 `RuntimeException`과 `Error`만 롤백한다(`DefaultTransactionAttribute.rollbackOn`, 185~186행).
- 이미 트랜잭션이 있는지: `AbstractPlatformTransactionManager.getTransaction`(373행)이 지금 스레드에 연결이 묶여 있으면 `handleExistingTransaction`으로 간다(384행). 없으면 새로 시작한다(`startTransaction` → `doBegin`, 532행).
- REQUIRED 참여: `handleExistingTransaction`의 끝에서 “Participating in existing transaction”을 남기고(497행) 새 트랜잭션 표시 없이(`newTransaction=false`) 상태만 만든다(518행). 연결에 아무것도 부르지 않는다.
- REQUIRES_NEW: `suspend(transaction)`이 바깥 연결을 스레드에서 떼어 두고(450행, `doSuspend`가 `unbindResource`, DataSourceTransactionManager 328행) 새로 시작한다. 새로 시작하는 `doBegin`이 풀에서 연결을 받는다(271행). 받지 못하면 `CannotCreateTransactionException`(320행)이고, 떼어 둔 바깥을 다시 묶은 뒤(`resumeAfterBeginException` → `resume`, 687행) 던진다.
- NESTED: `useSavepointForNestedTransaction()`이 참이면(469행, 기본값이 참, 1137~1138행) 연결에 저장점을 만든다(`createAndHoldSavepoint`, 477행). 이름은 `"SAVEPOINT_" + 연결마다의 번호`다(ConnectionHolder 47·185~187행). `DataSourceTransactionManager`는 NESTED를 기본으로 허용한다(138행).
- 커밋 때: `commit`(735행)이 먼저 이 상태의 표시(742행)를, 다음으로 연결의 rollback-only 표시(750행)를 본다. 연결에 표시가 있으면 “Global transaction is marked as rollback-only but transactional code requested commit”을 남기고 `processRollback(…, unexpected=true)`로 간다(754행). 새 트랜잭션이면 `doRollback`(897행) → `con.rollback()`(DataSourceTransactionManager 359행)이고, 끝에서 `UnexpectedRollbackException("Transaction rolled back because it has been marked as rollback-only")`을 던진다(938행).
- 참여한 쪽의 롤백: `processRollback`이 새 트랜잭션도 저장점도 아닌 상태를 받으면 연결에 rollback-only 표시만 한다(`doSetRollbackOnly`, 906행 → `ConnectionHolder.setRollbackOnly`). 저장점이 있으면 `rollbackToHeldSavepoint`(889행)이고, 이것은 저장점으로 되돌린 뒤 바로 저장점을 놓는다(AbstractTransactionStatus 162~163행).
- 시작과 끝의 연결 손질: `doBegin`은 자동 커밋이 켜져 있으면 끈다(298행). `doCleanupAfterCompletion`은 자동 커밋을 다시 켜고(389행) 연결을 돌려준다(403행).
- 예외 번역: Spring 6부터 기본 번역기는 `SQLExceptionSubclassTranslator`이고, 사용자가 `sql-error-codes.xml`을 classpath 뿌리에 두지 않으면 PostgreSQL 오류 코드표를 쓰지 않는다(JdbcAccessor 92~96행). 하위 클래스 번역은 JDBC 4 예외 하위 클래스만 보고(SQLExceptionSubclassTranslator 75~120행), 그다음 SQLSTATE 번역은 앞 두 글자로 고른다. `55`(object not in prerequisite state)는 어느 묶음에도 없다(SQLStateSQLExceptionTranslator 58~93행). 그래서 락 시간 초과 `55P03`은 `CannotAcquireLockException`이 아니라 `UncategorizedSQLException`이 된다고 예상한다(JdbcTemplate 1557행). 흔히 쓰는 설명(“락을 못 잡으면 `CannotAcquireLockException`”)은 PostgreSQL 오류 코드표의 `cannotAcquireLockCodes 55P03`(sql-error-codes.xml 259~261행)을 쓰던 때의 이야기다. pgjdbc의 예외가 JDBC 4 하위 클래스가 아니라는 것은 추론이다(pgjdbc 소스 확보 뒤 확인).

HikariCP 5.1.0

- 빌리기: `HikariPool.getConnection`(154행)이 `connectionBag.borrow`(162행)로 받고, 그 연결이 마지막으로 쓰인 지 500 ms(`aliveBypassWindowMs`, 65행)가 넘었으면 `isConnectionDead`(168행)로 살아 있는지 본다. 이 실험의 설정(검사 1초)에서는 `connection.isValid(1)`이다(PoolBase 149~157행). 시간 안에 못 받으면 `SQLTransientConnectionException("lab - Connection is not available, request timed out after …ms (total=…, active=…, idle=…, waiting=…)")`이다(684~686행).
- 돌려주기: `ProxyConnection.close`는 커밋 상태가 더럽고 자동 커밋이 꺼져 있을 때만 rollback을 부른다(249~250행). 자동 커밋은 지금 값이 풀의 기본값과 다를 때만 다시 맞춘다(PoolBase 214~217행). Spring이 돌려주기 전에 자동 커밋을 다시 켜므로, 돌려줄 때 연결에 부르는 것은 없다고 예상한다.

PostgreSQL 18.6

- 저장점: `DefineSavepoint`(xact.c 4373행), `RollbackToSavepoint`(4567행), `ReleaseSavepoint`(4458행). 하위 트랜잭션은 처음 쓸 때 번호를 받고, 부모에 번호가 없으면 부모가 먼저 받는다(`AssignTransactionId`, 635·660~678행). 이 장면은 바깥이 먼저 1번 행을 고치므로 부모에 이미 번호가 있다. 하위 트랜잭션을 되돌리면 그 번호만 abort로 적는다(`TransactionIdAbortTree`).
- 락 시간 초과: `ProcessInterrupts`가 `ERRCODE_LOCK_NOT_AVAILABLE`(55P03), “canceling statement due to lock timeout”을 낸다(postgres.c 3437~3439행). 교착 검사는 실험 2와 같다. 기다리기 시작하고 `deadlock_timeout`(1초) 뒤 한 번 돌고, 결과는 `DS_NO_DEADLOCK` 1 또는 `DS_HARD_DEADLOCK` 3이다.

pgjdbc 42.7.13(소스를 아직 받지 못해 아래는 모두 추론이다. 내려받은 뒤 소스로 확인하고, 틀리면 판정에 그대로 적는다)

- 자동 커밋을 끄는 것은 메시지가 아니다. 트랜잭션 안의 첫 문장을 보낼 때 `BEGIN`을 그 문장과 한 묶음(Sync 하나)으로 앞에 붙인다.
- 문장은 확장 질의(Parse, Bind, Describe, Execute, Sync)로 보낸다. 같은 문장을 다섯 번 쓰기 전에는 이름 없는 문장이다(`prepareThreshold` 기본 5).
- `commit()`·`rollback()`은 `COMMIT`·`ROLLBACK` 문장이다. 저장점은 `SAVEPOINT`, `ROLLBACK TO SAVEPOINT`, `RELEASE SAVEPOINT` 문장이고, 이름을 큰따옴표로 감싸는지는 모른다.
- `isValid`는 빈 문장 하나를 보내고 EmptyQueryResponse(`I`)를 받는다.
- 접속은 SSLRequest부터 연다(`sslmode` 기본 prefer). 이 서버는 ssl=off라 `N`을 받고 암호화 없이 StartupMessage를 보낸다.

pgjdbc 소스 확인(10-15 18:31, 내려받은 뒤·실행 전). 사용자 허락(“네 진행하세요”)으로 Maven Central에서 `postgresql-42.7.13.jar`(1,220,948 바이트)와 소스 jar(1,035,003 바이트)를 받았다. sha1은 Central의 `.sha1`과 같았고, 본체는 예전 Gradle 캐시의 값 `a6e1bd21…`과도 같았다. 위 추론을 소스로 대 보니 하나는 틀렸고 나머지는 맞았다. 아래 P 항목은 고치지 않고 등록한 그대로 판정한다.

- 틀림: `BEGIN`은 첫 문장과 한 묶음(Sync 하나)이 아니다. `sendQueryPreamble`(QueryExecutorImpl 723~773행)이 `BEGIN`을 단순 질의(`Q`) 하나로 먼저 보내고(`QUERY_EXECUTE_AS_SIMPLE`, 740행), 첫 문장은 그 뒤에 확장 질의로 붙는다. 그래서 서버는 `BEGIN`에 `C`·`Z`(T)를 따로 돌려준다. 한 번에 쓰는지(같은 TCP 쓰기인지)는 연결 위의 시각으로 본다.
- 맞음: 자동 커밋을 끄는 것은 메시지가 아니다(`setAutoCommit`, PgConnection 978~987행. 켜는 쪽은 트랜잭션 중일 때만 `commit()`을 부르고, `commit()`은 상태가 IDLE이면 아무것도 보내지 않음, 1043~1045행). `commit()`·`rollback()`은 `COMMIT`·`ROLLBACK` 문장이고 `BEGIN`을 붙이지 않는다(`QUERY_SUPPRESS_BEGIN`, 1014~1017행).
- 맞음, 그리고 모르던 것: 저장점은 `SAVEPOINT`, `ROLLBACK TO SAVEPOINT`, `RELEASE SAVEPOINT` 문장이다(1871·1891·1902·1910행). 이름이 있는 저장점은 큰따옴표로 감싼다(`PSQLSavepoint.getPGName` → `Utils.escapeIdentifier`, 77~81행). 그래서 Spring의 저장점은 연결 위에서 `SAVEPOINT "SAVEPOINT_1"`이다.
- 맞음: `isValid`는 빈 문장 `""`을 단순 질의로 보낸다(PgConnection 1603~1604행).
- 맞음: 프로토콜 판은 3.0이다(`protocolVersion` 기본 `"3"`, ConnectionFactoryImpl 290~306행). `gssEncMode` 기본 allow는 GSS 요청을 먼저 보내지 않는다(547~550행). `sslmode`가 없으면 prefer다(SslMode 66~71행).
- 맞음: `PSQLException`은 `SQLException`을 바로 잇는다(PSQLException 13행). JDBC 4 하위 클래스가 아니므로 P5의 `UncategorizedSQLException` 예상의 전제가 선다.

시험 한 벌(10-15 18:32, 본 기록 전). 5-required를 한 번 돌렸다(13초, JDK는 `java_home -v 25`가 고른 GraalVM 25.0.4). 세 층이 모두 남았다. 연결 둘의 메시지는 맥 쪽과 컨테이너 쪽이 74개·26개로 같았고 시계 차이 구간은 [0.052, 0.272] ms와 [0.084, 0.320] ms, JFR 메서드 사건 98개, 서버 관찰 6장이다. 「첫 시험에서 볼 것」의 답은 이렇다. StartupMessage의 프로토콜 판은 3.0이고, 풀을 닫는 데(`end`의 go부터 마지막 `X`까지) 2.0 ms 걸렸다. 본 기록 전에 고칠 것은 없다고 판단했다. 해석기에 넘길 것 하나를 봤다. pgjdbc의 `QueryExecutorImpl.execute`, HikariCP의 `HikariPool.getConnection`처럼 이름이 같은 오버로드는 한 호출에 사건 둘을 남긴다. 이 둘은 서술자(descriptor)와 나노초 시작 시각이 다르므로, 해석기는 메서드를 이름이 아니라 (클래스, 이름, 서술자)로 구분한다. P11의 “같은 메서드”도 그 뜻이다. 이 시험 기록은 판정에 쓰지 않는다.

### 예상과 판정

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| P1 | 모든 장면: `start` 전에 풀 크기만큼 TCP 접속이 들어오고, 접속마다 SSLRequest → `N` → StartupMessage(`user=postgres`, `database=postgres`) → `R`(0) … `K` … `Z`(I)다. 첫 접속은 앱이 풀을 만들 때(main 스레드), 둘째는 풀의 채우기 스레드가 연다 | 열 번 모두 그렇다 | 접속 수가 다르거나 SSLRequest가 없다(그러면 pgjdbc의 기본값이 내 예상과 다름) |
| P2 | 모든 장면: 장면의 첫 빌리기에서, 그 연결이 마지막으로 쓰인 지 500 ms가 넘었으면 `isConnectionDead`가 불리고 `BEGIN`보다 앞에 왕복 하나가 연결 위에 있다. 500 ms 이하면 둘 다 없다 | 조건과 결과가 열 번 모두 맞는다 | 조건과 결과가 어긋난다 |
| P3 | 5-required: 바깥의 UPDATE와 안쪽의 UPDATE가 같은 연결(C1)로 가고, 그 사이에 `BEGIN`·`SAVEPOINT`가 없다. `inner-after-update` 관찰에서 C1의 백엔드만 트랜잭션 번호를 갖고, 1·2번 행의 옛 버전 xmax가 모두 그 번호다. 안쪽의 예외(`inner-after-update`의 go)부터 `outer-after-inner`까지 어느 연결에도 메시지가 없고, JFR에 `doSetRollbackOnly`가 있다. `outer-after-inner`의 go 뒤 C1에 `ROLLBACK`이 가고(COMMIT이 아님), 앱의 `outer-threw`가 `UnexpectedRollbackException`(위 문구)이다. 마지막 관찰에서 그 번호가 aborted이고 잔액은 100, 200이다 | 두 번 모두 그렇다 | C2가 쓰이거나, 안쪽 실패 때 메시지가 있거나, COMMIT이 가거나, 다른 예외다 |
| P4 | 5-requires-new: 안쪽은 C2로 간다. `inner-after-update` 관찰에서 C1과 C2의 백엔드가 서로 다른 번호 A, B를 갖고 둘 다 `idle in transaction`이며, 1번 행의 옛 버전 xmax는 A, 2번 행은 B다. C2의 `COMMIT`이 C1의 `COMMIT`보다 먼저다. 마지막에 A, B 모두 committed, 잔액은 90, 210이다 | 두 번 모두 그렇다 | 안쪽이 C1을 쓰거나 커밋 순서가 반대다 |
| P5 | 5-requires-new-same-row: `blocked` 관찰에서 C2의 백엔드가 `Lock`/`transactionid`를 기다리고(그 락의 번호는 C1의 번호), C1의 백엔드는 `idle in transaction`이며 기다리는 것이 없다. `waiters`는 C2의 pid 하나다. C2의 백엔드에서 기다림을 시작하고 약 1초 뒤 `DeadLockCheck`가 1(`DS_NO_DEADLOCK`)을 돌려준다. 약 3초 뒤 C2에 `ErrorResponse`(55P03, “canceling statement due to lock timeout”, 자리 postgres.c `ProcessInterrupts`)가 오고, C2에 `ROLLBACK`, 그다음 C1에 `ROLLBACK`이 간다. 앱의 `outer-threw`는 `UncategorizedSQLException`이고 원인이 pgjdbc의 예외다. 마지막에 두 번호 모두 aborted, 잔액은 100, 200이다 | 두 번 모두 그렇다 | 서버가 교착으로 보고 40P01을 내거나, C1이 무언가를 기다리거나, 다른 예외 클래스다 |
| P6 | 5-requires-new-pool1: 접속은 하나(C1)뿐이다. `outer-after-update`의 go부터 앱의 예외까지 C1 위에 메시지가 없고 새 접속도 없다. `blocked` 관찰에서 C1은 `idle in transaction`이고 `waiters`는 비었다. JFR의 `HikariPool.getConnection`이 약 2000 ms 걸리고, 그 사이 `jdk.ThreadPark`가 HikariCP의 빌리기(`ConcurrentBag.borrow`) 아래에 있다. 앱의 `outer-threw`는 `CannotCreateTransactionException` ← `SQLTransientConnectionException`(위 문구, total=1, active=1)이고, 그 뒤 C1에 `ROLLBACK`이 간다 | 두 번 모두 그렇다 | 서버에 무언가 닿거나, 다른 예외이거나, 바깥이 커밋된다 |
| P7 | 5-nested: C1 하나로 `SAVEPOINT`(이름 `SAVEPOINT_1`), UPDATE(2번 행), `ROLLBACK TO SAVEPOINT`, `RELEASE SAVEPOINT`, `COMMIT` 순서다. 서버에서 안쪽 UPDATE 안에 `assign level=2` 뒤 `newxid`가 하위 번호 S(바깥 번호 A + 1)를 내고, `inner-after-update` 관찰에서 2번 행의 옛 버전 xmax가 S다. `ROLLBACK TO SAVEPOINT` 안에서 `TransactionIdAbortTree(S, nsub 0)`, `COMMIT` 안에서 `TransactionIdCommitTree(A, nsub 0)`이다. 마지막에 A committed, S aborted, 잔액은 90, 200이다 | 두 번 모두 그렇다 | 저장점이 없거나, 하위 번호가 없거나, 바깥이 롤백된다 |
| P8 | 모든 장면: 트랜잭션 끝(COMMIT·ROLLBACK의 `Z`) 뒤 그 연결을 풀에 돌려줄 때 연결 위에 메시지가 없다(HikariCP가 rollback이나 자동 커밋 맞추기를 보내지 않음). JFR에 `ProxyConnection.close`가 있다 | 열 번 모두 그렇다 | 돌려줄 때 메시지가 있다 |
| P9 | 모든 장면: `end`의 go부터 `closed`까지 접속마다 Terminate(`X`)가 가고, 컨테이너 쪽 기록도 `X`로 끝나며, bridge가 모두 끝난다 | 열 번 모두 그렇다 | `X` 없이 끊기거나 bridge가 남는다 |
| P10 | 도구: 실행마다 두 시계의 차이 구간이 비지 않고 양 끝이 1 ms 안이다 | 열 번 모두 그렇다 | 구간이 비거나 1 ms를 넘는다(그러면 맥과 컨테이너의 시각을 한 줄에 놓지 않음) |
| P11 | 도구: JFR에 같은 메서드·스레드·시작 시각의 사건 둘은 메서드 안의 throw를 finally가 받는 메서드에서만 나온다. 지나간 예외로 끝난 호출은 남은 사건의 스택에만 보인다 | 열 번 모두 그렇다 | 다른 꼴의 중복이나 빠짐이 있다(그러면 위 한계의 설명을 고침) |

P2의 500 ms는 HikariCP의 값이고, 장면 시작까지 걸리는 시간은 관찰 시간에 따라 달라 미리 정할 수 없다. 그래서 조건부로 등록한다. 두 실행 사이에 기다림 시간이나 메시지 수가 다르면 다르다고 적고, 한 번의 값을 성질로 쓰지 않는다.

한계: Spring Boot의 자동 설정, `@Transactional`의 readOnly·isolation·timeout, `JpaTransactionManager`, 가상 스레드, 여러 스레드가 한 풀을 다투는 경우는 이번에 일으키지 않는다. 앱의 시스템 콜은 보지 않는다(맥에 strace가 없음). 서버는 TCP를 보지 않는다(위 관측 경로).

### 덧붙인 장면 5-reborrow(10-15 20:30 등록, 실행 전): 같은 연결을 0.5초 안에 다시 빌릴 때

P2의 “500 ms 이하면 검사하지 않는다”는 쪽은 열두 벌에서 한 번도 일어나지 않았다(아래 실행 기록의 P2). 장면마다 첫 빌리기 앞에 `start` 멈춤 자리의 관찰이 있어, 연결이 늘 0.5초 넘게 쉬었기 때문이다. 그래서 같은 메서드를 멈춤 없이 두 번 잇달아 부르는 장면을 더한다.

- 앱: `Outer.once()`(REQUIRED, 1번 행에서 10을 뺌, 안에 멈춤 자리가 없음)를 `Main`이 두 번 잇달아 부른다. 멈춤 자리는 `start`와 `end`뿐이다. 두 빌리기 사이에 관찰을 넣으면 그 멈춤(약 0.4초)이 그대로 연결이 쉬는 시간이 되어 이 장면을 만든 뜻이 없어진다. 풀은 2개다.
- 근거로 읽은 소스(HikariCP 5.1.0). 연결을 돌려줄 때 `ProxyConnection.close`가 `PoolEntry.recycle`을 부르고(268행), 거기서 `lastAccessed`를 지금 시각으로 적는다(PoolEntry 80행). 빌릴 때는 `lastAccessed`부터 500 ms가 지났을 때만 `isConnectionDead`를 부른다(HikariPool 168행). 빌리기는 먼저 그 스레드가 돌려준 연결의 목록을 뒤에서부터 본다(ConcurrentBag 123~131행). 돌려줄 때 기다리는 스레드가 없으면 그 목록에 넣는다(191~194행). 그래서 둘째 빌리기는 첫째가 방금 돌려준 연결을 다시 받는다고 예상한다.
- 판정에 쓰는 쉰 시간은 해석기의 `borrows[].idleMs`(빌린 연결 위의 마지막 메시지부터 빌리기 시작까지)다. 마지막 메시지(`COMMIT`의 `Z`)는 `recycle`보다 앞이므로 이 값은 HikariCP가 재는 시간보다 길거나 같다. 그래서 이 값이 500 ms 이하이면 HikariCP의 값도 500 ms 이하다.

| 번호 | 예상 | 통과 | 실패 |
| --- | --- | --- | --- |
| P12 | 두 트랜잭션이 모두 C1으로 간다(둘째 빌리기는 첫째가 돌려준 연결). 장면 동안 C2에는 메시지가 없다. 트랜잭션 번호는 753, 754이고 마지막에 둘 다 committed, 잔액은 80, 200이다 | 두 번 모두 그렇다 | 둘째가 C2를 쓰거나, 번호·잔액이 다르다 |
| P2(나머지 쪽) | 첫 빌리기는 쉰 지 500 ms가 넘어 `isConnectionDead`가 불리고, 첫 `BEGIN` 앞에 빈 문장 왕복이 있다. 둘째 빌리기는 쉰 지 500 ms 이하이고, 그 안에 `isConnectionDead`가 없으며, 둘째 트랜잭션의 첫 앱 메시지가 `Q BEGIN`이다(앞에 `Q ""`이 없음) | 두 번 모두 그렇다 | 둘째에 검사가 있다. 쉰 시간이 500 ms를 넘으면 실패가 아니라 판정 불가로 적는다 |
| P13 | 도구: 기록끼리 견준 갈림(`decode spring-compare`, 아래 「실험 5 화면 설계」)이 `HikariPool.getConnection(long)`에서 길 셋을 낸다. 검사하는 길(`borrow` → `isConnectionDead`, 정상 반환), 검사하지 않는 길(`borrow`, 정상 반환, 이 장면의 둘째 빌리기), 시간 초과 길(`borrow`, 예외를 만들고 끝남, 풀 1칸 장면)이다 | 셋이 그렇게 나온다 | 둘째의 길이 시간 초과 길과 하나로 묶이거나, 갈림으로 나오지 않는다 |

## 실험 2~5와 6·7(등록만, 실험마다 실행 전에 따로 사전 등록)

- 2 락: 잔액 덮어쓰기(읽고 계산해서 쓰기)와 바른 UPDATE, `SELECT ... FOR UPDATE`, 교착과 `deadlock_timeout` 뒤의 `CheckDeadLock`, 외래 키 검사의 `FOR KEY SHARE`(`RI_FKey_check`), 멱등성 키(고유 인덱스에서 두 번째 INSERT가 첫 번째의 끝을 기다리는 `_bt_check_unique`).
- 3 원자성과 지속성: `ROLLBACK` 뒤 페이지의 행은 남고 커밋 로그만 바뀌는 것, 다음 읽기가 `HEAP_XMIN_INVALID` 힌트 비트를 쓰는 것, `synchronous_commit` on/off의 WAL 동기화 자리, 커밋 직후 `kill -9`와 재시작 복구(WAL 재실행). off일 때는 WAL 라이터가 쓰기 전에 죽이면 커밋 응답을 받은 행이 사라지는지.
- 4 SERIALIZABLE: 당직 의사 둘의 write skew, `pg_locks`의 SIReadLock, 커밋 때의 검사(`PreCommit_CheckForSerializationFailure`).
- 5 전파: Spring Framework 7.0.9(spring-tx), HikariCP 7.0.2, pgjdbc 42.7.13. 이 맥의 Gradle 캐시에 jar가 있고 spring-tx·spring-jdbc·HikariCP는 소스 jar도 있어서 내려받을 것이 없다(pgjdbc 소스 jar만 없음). 앱은 맥에서 돌리고, Spring 쪽은 JFR `jdk.MethodTrace`(JDK 25)와 Spring 로그, DB 쪽은 지금까지의 기록으로 본다. 맥에는 strace가 없으므로 클라이언트 시스템 콜은 보이지 않는다. 클라이언트 시스템 콜이 필요하면 리눅스 컨테이너에 JDK가 필요하고 내려받기라 사용자 허락이 필요하다.
    - 2026-10-15 01:36 덧붙임(환경이 바뀜): 위 문장은 2026-10-14 20:08 KST에 본 Gradle 캐시(`~/.gradle/caches/modules-2/files-2.1`의 `org.springframework/spring-tx/7.0.9`, `org.postgresql/postgresql/42.7.13`)에 기댄 것이다. 사용자가 그 뒤 디스크를 정리했고, 지금 그 캐시에는 Spring·HikariCP·pgjdbc가 하나도 없다(110개 무리 가운데 없음, `mdfind`와 `~` 아래 `find`도 찾지 못함, 홈 아래 5 MB 넘는 jar 안에도 `org/postgresql/Driver.class`가 없음). 남은 것은 `~/.m2/repository`의 Spring 6.2.15(aop, beans, context, core, expression, jcl, jdbc, tx, 모두 소스 jar 있음), HikariCP 5.1.0(소스 jar 있음), logback 1.5.22·slf4j-api 2.0.17(소스 jar 있음)이다. Spring 7.0.1은 spring-jdbc가 없어 쓸 수 없다. 그래서 실험 5는 Spring 6.2.15와 HikariCP 5.1.0으로 한다(AI 판단. 전파 코드는 두 판 모두 `AbstractPlatformTransactionManager`의 같은 계보이고, Spring Boot 3.x가 쓰는 판이다. 사용자가 7을 원하면 다시 연다). pgjdbc는 내려받아야 하므로 사용자 허락을 기다린다.
- 6 연결의 일생(원문 3): TCP 접속, 백엔드 fork, StartupMessage, SCRAM 인증, 첫 ReadyForQuery, HikariCP의 연결 채우기·검사·수명 만료. 지금 해석기는 첫 ReadyForQuery 앞(접속 인증)을 버리므로 “0번째 요청” 해석을 더해야 한다.
- 7 사이에 낀 장치(원문 3): L4(기존 IPVS 실험대), PgBouncer(내려받기 필요), ProxySQL(MySQL, 내려받기 필요). Aurora는 위 이유로 로컬 재현 불가.
- 같은 바이트를 다른 언어로: psql, Go(pgx, 이 맥의 Go 모듈 캐시에 있음), Java(pgjdbc)가 같은 트랜잭션에 보내는 `BEGIN`·`COMMIT` 바이트를 견준다. 실험 5와 함께 한다.

## 화면 시안(20:22)

실험 1의 화면 시안 세 장면을 사용자에게 보냈다(값은 QLL-07 기록 [`exercise-4-read-committed.out`](../../open_source_internals/database/postgresql/evidence/exercise-4-read-committed.out), [`mvcc-page.out`](../../open_source_internals/database/postgresql/evidence/mvcc-page.out)과 예시). 위 줄은 미니맵(백엔드 둘이 같은 공유 메모리를 씀, 공유 버퍼 상자 안에 페이지의 행 버전)과 크기를 고정한 psql 창 둘, 아래 왼쪽은 세션 둘과 공유 메모리 생명선 위의 시간 순서(기다림은 빗금 띠, 깨움은 점선 화살표), 아래 오른쪽은 행 버전마다 xmin부터 xmax까지를 트랜잭션 번호 축 위의 막대로, 스냅숏을 세로선으로 그린 그림이다. 세로선이 지나는 막대가 그 SELECT에 보이는 버전이다. 화면 조립은 사용자 의견을 받은 뒤 확정하고, 기록은 화면과 무관하므로 먼저 진행한다.

## 실행 기록

### 실험 1 기록 열 벌(21:12)

시나리오 다섯을 두 번씩(a, b) 기록했다. 한 벌은 75~77 MB이고 저장소 밖(세션 scratchpad)에 둔다. 열 벌 모두에서 트랜잭션 번호가 같게 나왔다(설정 트랜잭션 752, 시나리오의 두 쓰기 753·754). 열 벌 모두 ftrace 넘침(overrun, dropped)이 0이다.

사전 등록에서 벗어난 방법 둘:

1. 관찰용 접속과 `pg_isready` 접속의 함수 기록을 지웠다. 한 벌 약 490 MB 가운데 410~500 MB가 이 둘이었다(관찰 쿼리마다 계획·실행 전체가 기록되고, 접속마다 카탈로그 캐시를 채움). 맥 디스크 여유가 14 GB까지 줄어 실행 중에 정한 처리다. 이 둘의 uprobe 사건은 남아 있으므로 E8은 uprobe 사건과 페이지 관찰로 판정했다. 이후 기록부터는 `txlab stop`이 같은 처리를 한다.
2. **반환 uprobe가 함수 기록을 바꿨다.** 처음 해석에서 `XidInMVCCSnapshot`, `TransactionIdDidCommit`, `GetSnapshotData`처럼 반환 uprobe(`%return`)를 건 함수만 부른 줄이 `?`로 나왔다. 그 함수들의 입구·출구 기록에 남은 부른 자리가 모두 `0xfffffffff000`이었다. arm64 커널은 반환 uprobe를 건 함수의 입구에서 돌아갈 주소(LR)를 uprobe trampoline 주소로 바꾼다(`arch/arm64/kernel/probes/uprobes.c`의 `arch_uretprobe_hijack_return_addr`, 193행 “Replace the return addr with trampoline addr”). 함수 기록은 그 뒤 함수 머리에서 돌아갈 주소를 읽으므로 trampoline 주소를 남긴다. 관측 도구 둘이 서로를 건드린 사례다. 그 주소가 `0xfffffffff000`이라는 것은 관측값이다. trampoline 자리를 정하는 공통 코드(`kernel/events/uprobes.c`)는 이 맥에 받아 둔 커널 소스에 없어 확인하지 못했다.
    - 고친 방법: ftrace의 반환 사건 줄 `(돌아갈 주소 <- 함수)`에는 커널이 보관한 원래 주소가 있다. 해석기(`repairSites`)가 짝지은 반환 사건마다 그 출구 기록과, 기록 순서의 스택으로 찾은 짝 입구 기록의 부른 자리를 그 값으로 바꾼다.
    - 고친 값이 맞다는 근거: uprobe를 넣기 전에 시작한 백엔드(7188)의 기록에 바뀌지 않은 `GetSnapshotData` 부른 자리 `0xaaaac7895ffc`(snapmgr.c 423행)가 있다. uprobe 반환 사건 줄의 돌아갈 주소가 이 값과 같다.
    - 결과: 한 벌에 1,928~2,050개를 고쳤고 고치기 전 값은 모두 `0xfffffffff000`이었다. 고친 뒤 남은 것은 rr-write 두 벌의 `exec_simple_query` 하나씩이다. 직렬화 오류로 longjmp해 반환하지 않은 요청이라 고칠 값이 없다. 실행 파일 밖을 가리키는 부른 자리 767개(rc-read-a)는 모두 libc가 부르는 함수다(`qsort`의 비교 함수 `dispatch_compare_ptr`, 신호 처리기 `wrapper_handler`, `atexit_callback`, `main`). 시험 `TestRepairSites*` 둘이 있다.

판정:

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| E1 | 결과 같음, 호출 2·2 | 같음 | **예상과 다름** | 결과는 예상대로다. 1번째 SELECT는 `(1,100) (2,200)`, 2번째는 `(1,150) (2,200) (3,300)`이라 non-repeatable read와 팬텀이 보인다. 그런데 SELECT 하나가 `GetSnapshotData(&CurrentSnapshotData)`를 1번이 아니라 **2번** 불렀다. 한 번은 구문 분석·계획용(`exec_simple_query`, postgres.c 1163행 “Set up a snapshot if parse analysis/planning will need one”)이고, 다른 한 번은 실행용(`PortalStart`, pquery.c 481행)이다. READ COMMITTED에서는 `GetTransactionSnapshot`이 부를 때마다 새로 찍는다(snapmgr.c 342행). 트랜잭션의 첫 SELECT에서는 첫 호출이 330행, 둘째가 342행이다. 사전 등록의 실패 조건(2번째 SELECT의 호출 0번, 또는 결과가 같음)에는 걸리지 않지만 통과 조건(1번씩)도 아니므로, 소스를 덜 읽은 예상이었다. 이번 실행에서 두 스냅숏은 값이 같았다(`753:753:`, `755:755:`). 계획과 실행 사이에 다른 커밋이 끼면 두 값이 달라질 수 있다는 것은 추론이고 아직 일으켜 보지 않았다 |
| E2 | 1·0 | 1·0 | 통과 | REPEATABLE READ의 1번째 SELECT는 snapmgr.c 321행에서 한 번만 찍고(트랜잭션 스냅숏을 복사해 끝까지 씀), 2번째 SELECT는 0번이다. 두 결과와 `snap`이 같다(`753:753:`, `(1,100) (2,200)`). 2번째 SELECT의 줄 판정은 모두 그 트랜잭션 스냅숏 753:753으로 했다. S2가 만든 `(0,3)`, `(0,4)`는 `XidInMVCCSnapshot`이 참이라 안 보이고, `(0,1)`은 지운 쪽(753)이 아직 안 끝난 것으로 보여 보인다 |
| E3 | 같음 | 같음 | 통과 | READ UNCOMMITTED의 호출 수·부른 줄·결과가 READ COMMITTED와 같다 |
| E4 | 0 | 0 | 통과 | 시나리오 세션 백엔드의 짝지은 묶음(27~39개)마다 uprobe 사건 수와 함수 기록 수가 같다. 관찰용 접속은 함수 기록을 지웠으므로 이 범위에 들지 않는다. 짝지은 사건의 두 시계 차이는 중앙값 약 0 µs, 최대 편차 약 1.0 ms다 |
| E5 | 모두 같음 | 모두 같음 | 통과 | 결과 행이 있는 모든 단계에서 그 표의 `HeapTupleSatisfiesMVCC` 호출 수가 직전 관찰의 정상 줄 수와 같고(read 셋은 2·4, rc-write는 3·4, rr-write는 3·3), 참을 돌려준 줄의 집합이 결과의 `ctid` 집합과 같다. 해석기의 `steps[].check`가 이 값을 낸다 |
| E6 | 6/7 | 6/7 | **한 조각이 예상과 다름** | 맞은 것: S1 SELECT 결과 `(1,100) (2,200)`(바뀌는 중인 150은 안 보임), `XactLockTableWait` 첫 인자 753(S2), S2 COMMIT 안의 `ProcWakeup` 대상이 S1 pid, S2 백엔드의 `kill(S1 pid, SIGURG)`, 깨어난 S1의 `EvalPlanQual`(nodeModifyTable.c 2625행), 마지막 잔액 120. 틀린 것: “`snap`의 진행 중 목록에 S2의 번호가 있다”. 실제 스냅숏은 `753:753:`이고 진행 중 목록(xip)이 비어 있다. xmax는 마지막으로 끝난 번호 + 1이므로(procarray.c 2247행 “xmax is always latestCompletedXid + 1”) 752 다음인 753이다. `XidInMVCCSnapshot`은 xmax 이상의 번호를 목록을 보기 전에 진행 중으로 판정한다(snapmgr.c 1883~1885행). 그래서 가장 최근에 시작한 S2는 목록이 아니라 xmax 경계로 가려진다. “진행 중”을 가르는 길이 둘이라는 것이 이 관찰의 요점이다 |
| E7 | 통과 | 통과 | 통과(줄 번호 주석) | 같은 기다림과 깨움 뒤 S1에 `could not serialize access due to concurrent update`가 왔다. `errfinish`를 부른 자리는 `ExecUpdate`의 nodeModifyTable.c **2602행**이다. 사전 등록에 적은 2604행은 같은 `ereport(ERROR, ...)` 문의 `errmsg` 줄이다. 컴파일러가 이 호출을 문이 시작하는 2602행에 붙였다. `EvalPlanQual`은 불리지 않았다. 직렬화 오류 뒤 S1의 754는 `TransactionIdAbortTree`로 커밋 로그에 중단으로 적혔다 |
| E8 | 0 | 0 | 통과 | 관찰용 접속은 acct 행에 `HeapTupleSatisfiesMVCC`를 한 번도 부르지 않았다. 그 접속의 `SetHintBits` 61번 가운데 59번은 그 함수 밖에서 불려 uprobe 값만으로는 어느 표의 행인지 모른다(`SetHintBits`는 표 번호를 받지 않음). 그래서 페이지를 직접 견줬다. 열 벌에서 acct 페이지의 이미 있던 줄에 새로 켜진 힌트 비트 34개가 모두, 두 관찰 사이에 세션이 같은 값의 행에 같은 비트를 쓴 `SetHintBits`와 맞았다. 설명되지 않는 비트는 0개다. 판정 함수는 비트가 꺼져 있을 때만 `SetHintBits`를 부르므로, 관찰용 접속이 먼저 썼다면 세션의 쓰기가 없어 설명되지 않는 비트가 남았을 것이다. 한계: 값으로 맞추므로 xmin·xmax가 같은 카탈로그 행에 세션이 쓴 기록도 설명으로 잡힐 수 있다. 다음 실험부터는 행 머리 주소를 함께 기록해 표를 직접 가른다 |

사전 등록에 없던 관찰:

1. **기다리는 쪽이 1초에 한 번 스스로 깨어난다.** 네 벌의 기다림 길이가 1,087, 635, 644, 1,048 ms로 달랐다. 1초를 넘긴 두 벌(rc-write-a, rr-write-b)에서는 기다림을 시작하고 1초 뒤 SIGALRM 처리기 `handle_sig_alarm` → `CheckDeadLockAlert` → `CheckDeadLock` → `DeadLockCheck`가 돌았다(교착 검사). 검사가 끝나면 다시 잠들고, S2의 커밋이 보낸 `SIGURG`로 깨어났다. 1초 전에 끝난 두 벌(rc-write-b, rr-write-a)에서도 기다림을 시작하고 1,000.0~1,000.1 ms 뒤에 SIGALRM이 왔다. 이미 기다림이 끝나 `disable_timeout`을 부른 뒤였고, 처리기는 `SetLatch`만 하고 끝났다. 타이머를 끌 때마다 `setitimer`를 부르지 않으려고 예약된 신호를 그대로 두기 때문이다(timeout.c 296~317행, 꺼진 타이머만 남으면 `schedule_alarm`이 아무것도 하지 않음). 처리기는 늘 자기 래치를 세우고(377행), 자기 래치를 세우면 `kill(MyProcPid, SIGURG)`로 자기를 깨운다(latch.c `SetLatch` → waiteventset.c 2027행). 그래서 strace에는 다른 세션이 깨우는 `SIGURG`와 자기가 자기를 깨우는 `SIGURG`가 둘 다 있다. 화면은 이 둘을 갈라 그린다. deadlock_timeout은 기본값 1초다(postgresql.conf 828행 주석, `postgresql.auto.conf` 비어 있음, guc_tables.c 2271~2272행 기본값 1000).
2. **새 버전에 자기 잠금이 xmax로 남는다.** rc-write에서 S1이 만든 마지막 버전 `(0,4)`는 xmin과 xmax가 둘 다 754이고 플래그가 `HEAP_XMAX_KEYSHR_LOCK | HEAP_XMAX_LOCK_ONLY`다. 깨어난 S1은 `heap_lock_tuple`로 S2의 버전 `(0,3)`을 먼저 잠그고(xmax 754, 잠금만) 그 버전을 고쳤다. `heap_update`는 잠근 것이 자기뿐이면 잠금 정보를 보존하고(heapam.c 3684~3692행), 그 xmax를 새 버전에 옮기며(3826~3828행), 그 비트를 KEY SHARE 잠금으로 적는다(그 아래). 이 xmax는 “지운 쪽”이 아니라 잠금이라 다음 SELECT의 판정은 그 버전을 보인다고 했다. 화면의 버전 막대는 잠금만인 xmax에서 끝나지 않게 그린다.
3. **REPEATABLE READ는 문장 사이에도 스냅숏을 쥔다.** 관찰한 `pg_stat_activity.backend_xmin`이 RR의 S1에서는 첫 SELECT 뒤 COMMIT까지 753이고, RC의 S1에서는 문장 사이에 비어 있다(두 벌 같음). RR 트랜잭션이 오래 열려 있으면 그만큼 오래된 버전을 VACUUM이 치우지 못하는 이유가 이 값이다.
4. **READ COMMITTED는 문장마다 카탈로그 스냅숏을 버린다.** 첫 SELECT의 카탈로그 스냅숏(`&CatalogSnapshotData`) 수가 RC·RU는 2, RR은 1이다. RC의 `GetTransactionSnapshot`은 새 스냅숏을 찍기 전마다 `InvalidateCatalogSnapshot`을 부른다(snapmgr.c 340행).
5. **힌트 비트는 읽는 쪽이 쓴다.** acct의 처음 두 행에 `HEAP_XMIN_COMMITTED`를 쓴 것은 그 행을 처음 판정한 세션이다. rc-read에서는 S1의 첫 SELECT, write 시나리오에서는 S2의 UPDATE다. 깨어난 S1은 판정 함수 밖에서도 비트를 썼다. 기다림 직후 `heap_update`의 `UpdateXmaxHintBits`(heapam.c 3730행)가 `HeapTupleSetHintBits`(heapam_visibility.c 141행)를 통해 `(0,1)`에 `XMAX_COMMITTED`를 썼고, `heap_lock_tuple`이 버전 사슬을 따라가며 부른 `HeapTupleSatisfiesDirty`(743행)가 `(0,3)`에 `XMIN_COMMITTED`를 썼다.
6. **두 실행의 차이는 시간뿐이었고, 그 시간의 큰 몫은 S2 커밋의 디스크 동기화였다.** 트랜잭션 번호, 결과, 호출 수, 부른 줄, 힌트 비트를 쓴 세션은 a와 b가 같다. 다른 것은 기다림 길이와 그에 따른 교착 검사 여부(위 1)다. 기다림 길이는 두 몫의 합이다. 하나는 S1이 기다리기 시작해 S2의 COMMIT이 오기까지의 드라이버 몫(521~678 ms, 기다림을 100 ms마다 `pg_stat_activity`로 확인하고 다음 문장을 보내는 데 걸린 시간)이다. 다른 하나는 S2의 COMMIT이 WAL 파일에 부른 `fdatasync`(rc-write 544·111 ms, rr-write 97·367 ms, strace `-T`)다. 1초를 넘긴 두 벌은 `fdatasync`가 길었던 벌이다. 순서는 `fdatasync`가 끝남 → 커밋 로그 기록(`TransactionIdCommitTree`) → `kill(S1, SIGURG)`다. 커밋은 WAL을 디스크에 내린 뒤에야 기다리던 쪽을 깨운다. 이 순서는 아래 「실험 1 화면」의 strace 시각 수정 뒤 열 벌의 커밋 18번 모두에서 다시 확인했다(`fdatasync` 끝에서 0.05~1.2 ms 뒤 커밋 로그, 그 뒤 0.3~1.4 ms 안에 `kill`). 같은 열 벌에서 `fdatasync` 한 번은 2.5 ms부터 544 ms까지 걸렸다. 이 맥에서 `fdatasync`가 수백 ms까지 걸리는 이유(OrbStack 가상 디스크와 맥 파일 시스템 사이)는 확인하지 않았다. 교착 검사를 언제나 보려면 S2의 COMMIT 앞에 1초 넘는 쉼을 두어야 한다. 실험 2(교착)에서 그렇게 등록한다.

재현: 기록은 `txlab` 디렉터리에서 `go run . record -scenario ../../../../../database/postgresql/transactions/<이름>.txt -out <디렉터리>`, 해석은 `decode -bin <기록 빌드> -dir <디렉터리>/trace scenario -steps <디렉터리>/steps.json -uprobe <디렉터리>/uprobe.trace -strace <디렉터리>/all.strace -src <PostgreSQL 18.6 소스>`다. 해석 결과의 `joins`(E4), `siteRepair`, `steps[].check`(E5), `bystanders`·`hintChanges`(E8)가 위 판정의 값이다. 같은 기록을 다시 해석했을 때 이 값들이 위와 다르면 해석기가 바뀐 것이다.

### 실험 1 화면(21:35)

화면은 [`concept-5-tx.html`](../../open_source_internals/workbench/design/concept-5-tx.html) 한 파일이다. 템플릿 [`tx-view.tmpl.html`](../../open_source_internals/workbench/design/tx-view.tmpl.html)에 열 벌의 해석 결과를 `txlab view`가 넣는다. 화면은 값을 셈하지 않고 고르고 그리기만 한다. 예외는 두 가지로, 버전 막대의 위치와 기다림 길이(ProcSleep 입구와 출구 시각의 차)다. 화면 시안(20:22)의 배치를 그대로 쓰고, 실제 기록에서 생긴 것을 더했다.

- 장면: 단계 하나가 보통 장면 하나다. 락을 기다린 단계는 깨어나기 전과 뒤로 나누고, 깨어난 뒤의 일은 깨운 세션의 단계 다음 장면으로 옮긴다. 그래서 장면 순서가 실제 시각 순서와 같다.
- 판정 경로: 줄마다 `HeapTupleSatisfiesMVCC` 한 번이 부른 함수(uprobe 값)와, 그 안에서 마지막으로 지난 줄(함수 기록의 직접 호출)을 칩으로 보인다. 예를 들어 RR 2번째 SELECT의 `(0,3)`은 `XidInMVCCSnapshot(753)`이 snapmgr.c 1884행(xmax 경계)에서 참을 돌려 안 보인다. 753은 커밋 로그에서는 커밋됐으므로, 칩은 “스냅숏 기준 진행 중”이라고 쓴다.
- 대조 표시: 버전 막대와 스냅숏 선의 규칙(선이 지나는 막대가 보임)이 기록된 판정과 다르면 빨간 표지를 띄운다. 해석기의 `steps[].check`(E5)도 표지로 보인다. 열 벌의 72장면을 모두 그려 센 결과 빨간 표지는 0개다. 추적하는 일이 하나도 없는 22장면(BEGIN, 읽기만 한 COMMIT, 오류 뒤 ROLLBACK)에는 그 사실을 적는다. BEGIN은 번호도 스냅숏도 받지 않는다.
- 그림 다섯 장: [RC 2번째 SELECT](../../open_source_internals/workbench/design/concept-5-rc-read.png)(새 스냅숏 755:755:로 행 3개), [RR 2번째 SELECT](../../open_source_internals/workbench/design/concept-5-rr-read.png)(첫 스냅숏을 다시 써서 줄 3·4가 안 보임), [RC 쓰기의 기다림](../../open_source_internals/workbench/design/concept-5-rc-write-wait.png)(1초 타이머, 교착 검사, 다시 잠듦), [RC 쓰기의 깨어난 뒤](../../open_source_internals/workbench/design/concept-5-rc-write-epq.png)(새 버전으로 WHERE 다시 계산), [RR 쓰기의 직렬화 오류](../../open_source_internals/workbench/design/concept-5-rr-write-error.png)(nodeModifyTable.c 2602행).

화면을 맞추다 해석기의 결함 하나를 찾아 고쳤다. **strace의 시각은 줄의 꼴에 따라 뜻이 다르다.** 한 줄로 끝난 호출의 시각은 커널에 들어간 시각이고, `<unfinished ...>`와 `<... resumed>`로 나뉜 호출의 resumed 줄 시각은 끝난 시각이다. 해석기는 둘 다 끝난 시각으로 보고 있어서, rr-read-a의 INSERT가 부른 `fdatasync`(실제 시작 2,291.9 ms, 20.0 ms)를 요청이 오기도 전인 2,271.9 ms에 그렸다. 근거는 같은 벌의 strace다. 5 ms 넘게 걸린 한 줄 호출 9건 모두 같은 pid의 다음 줄 시각이 “그 줄 시각 + `-T` 시간” 이후였다. 나뉜 호출 175건은 모두 resumed 줄 시각이 “unfinished 줄 시각 + `-T` 시간”과 2 ms 안에서 같았다. `readStrace`가 한 줄 호출에 `-T` 시간을 더해 모든 시각을 끝난 시각으로 맞추도록 고쳤다. 시험 `TestReadStraceGivesEndTimeForBothLineForms`는 고친 줄을 되돌리면 실패하는 것까지 확인했다. 이 수정은 앞 단계의 요청 하나의 여정 해석(`journey`)에도 닿는다. `recvfrom`·`sendto` 시각이 그 호출의 걸린 시간만큼 늦어지지만, 이 둘은 소켓에 이미 온 바이트를 옮기는 호출이라 그 차이는 µs 단위로 추론한다. 그래서 앞 단계의 그림은 다시 만들지 않았다.

재현: `txlab` 디렉터리에서 `go run . view -tmpl ../../../tx-view.tmpl.html -out ../../../concept-5-tx.html 1-rc-read/a=<해석 결과> ...`이다. 이름은 `시나리오/실행`이다.

### 실험 2 기록 열여섯 벌(22:35)

시나리오 여덟을 두 번씩 기록했다. 한 벌은 67~81 MB이고, 2-for-update/b만 164 MB다(아래 방법 2). 열여섯 벌 모두 ftrace 넘침이 0이고, 트랜잭션 번호는 a와 b가 같다(설정 752, 시나리오 753·754). 벌마다 단계·사건·관찰을 한 줄씩 뽑아 시각·pid·주소를 가린 뒤 a와 b를 견주면 여덟 시나리오 모두 차이가 없다. a와 b가 다른 것은 시각뿐이다.

사전 등록에서 벗어난 것 둘:

1. **결과를 보는 SELECT 셋이 인덱스 스캔이었다.** L12는 결과 SELECT를 순차 스캔으로 쓰기로 했는데, 2-deadlock·2-fk-nokey·2-fk-delete의 마지막 SELECT에 `ORDER BY id`를 넣었고 계획기가 기본 키 인덱스를 골랐다. 판정의 스택이 근거다. 인덱스 스캔은 `IndexNext → heapam_index_fetch_tuple → heap_hot_search_buffer`, 순차 스캔은 `SeqNext → … → page_collect_tuples`를 지난다. 외래 키 둘은 죽은 버전이 사슬 밖에 없어 판정 수가 정상 줄 수와 우연히 같았고, 교착은 4번 판정에 정상 줄 5개로 달랐다(아래 관찰 2). 정렬이 필요하면 순차 스캔 뒤에 정렬하는 계획이 나오게 조건을 고르거나, 판정 대조를 인덱스가 가리키는 사슬 기준으로 바꿔야 한다. 이번에는 화면이 스택으로 인덱스 스캔을 알아보고 설명을 붙인다(아래 「실험 2 화면」).
2. **pg_isready 접속의 함수 기록 하나가 지워지지 않았다.** 2-for-update/b의 95 MB짜리 `529.bin`이 그 백엔드다. 그 접속의 시작 메시지가 다른 pid의 줄에 끼여 `recvfrom(… <unfinished ...>`와 `<... recvfrom resumed>"…pg_isready…"`로 나뉘었고, 기록 정리(`pruneTraces`)는 `" recvfrom("`이 든 줄만 봤다. 받은 바이트는 resumed 줄에만 있다. 실험 1 화면에서 고친 strace 시각 문제와 같은 원인(호출 하나가 두 줄로 나뉨)이 다른 자리에서 나온 것이다. 두 꼴을 모두 보게 고쳤고 시험 `TestIsreadyStartupFindsBothLineForms`는 고치기 전 실패, 고친 뒤 통과했다. 이 pid는 시나리오 세션이 아니어서 해석 결과에는 영향이 없다. 남은 파일은 그대로 두었다.

판정:

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| L1 | 통과 | 통과 | 통과 | 어느 세션도 `ProcSleep`에 들어가지 않았다. S2의 UPDATE는 새 스냅숏 754:754로 S1의 버전 (0,2)(xmin 753, 잔액 70)를 찾아 `HeapTupleSatisfiesUpdate` → `TM_Ok`를 받고, 앱이 계산한 50을 썼다. 마지막 잔액 50. S1의 −30은 기다림도 오류도 없이 사라졌다 |
| L2 | 통과 | 통과 | 통과 | S1의 번호 753은 `heap_lock_tuple`(모드 `Exclusive`)의 입구와 반환 사이에 받았다(a 859.1 ms, b 893.9 ms, 자리는 두 벌 같음). 그 단계 뒤 (0,1)의 xmax는 753에 `HEAP_XMAX_EXCL_LOCK`·`HEAP_XMAX_LOCK_ONLY`, 그리고 `HEAP_KEYS_UPDATED`다 |
| L3 | 통과 | 통과 | 통과 | S2는 `heap_lock_tuple` 안에서 `HeapTupleSatisfiesUpdate` → `TM_BeingModified`(xmax 753)를 받고 heapam.c 5218행의 `XactLockTableWait(753, Lock)`로 잠들었다. 1,005.6·1,002.9 ms 뒤 `DeadLockCheck`가 `NoDeadlock`을 돌려주고 다시 잠들었다. S1의 COMMIT 안의 `ProcWakeup`이 깨운 뒤 (0,1)에 `TM_Updated`, 사슬을 따라간 (0,2)에 `TM_Ok`를 받았고, 그 잠금 안에서 번호 754를 받았다. S2의 결과는 (0,2) 잔액 70, 마지막 잔액 20이다. 판정은 스냅숏 753:753으로 (0,1)만 보인다고 했고 결과 행은 (0,2)다. 예상한 어긋남이다 |
| L4 | 통과 | 통과 | 통과 | S1이 기다리기 시작해 S2가 기다리기까지 589.8·570.1 ms였다. S1의 타이머가 기다림 시작 1,000.1 ms 뒤에 울렸고 `DeadLockCheck`가 `HardDeadlock`을 돌려줬다. `errfinish` 스택은 `DeadLockReport`(deadlock.c 1133행) ← `LockAcquireExtended`(lock.c 1235행)다. 753의 `TransactionIdAbortTree`는 오류 때(`AbortCurrentTransactionInternal`, xact.c 3528행) 불렸고, 뒤의 ROLLBACK에는 커밋 로그 기록이 없다 |
| L5 | 한 조각 다름 | 한 조각 다름 | **한 조각이 예상과 다름** | 맞은 것: S2에는 교착 검사가 없었고(기다림 415.3·437.4 ms), S1의 중단 처리 안의 `ProcWakeup`(대상 S2)이 깨웠다. 마지막 잔액은 120·180이다. 틀린 것: “S2가 (0,1)을 다시 확인해 `TM_Ok`를 받는다”. 깨어난 뒤 `HeapTupleSatisfiesUpdate`는 다시 불리지 않았다. `heap_update`는 기다리기 전에 읽어 둔 xmax(753)가 그대로이면 처음으로 돌아가지 않고(heapam.c 3724~3727행), `UpdateXmaxHintBits`(3730행)가 그 번호의 끝을 보고 힌트를 쓴 뒤 그 힌트로 결론을 낸다(3731~3736행). 기록에는 깨어난 지 0.2 ms 뒤 S2가 (0,1)의 행 머리 주소에 `HEAP_XMAX_INVALID`를 쓴 `SetHintBits`(스택 `UpdateXmaxHintBits` ← `heap_update`)가 두 벌 모두 있다. 그 비트는 S2가 곧바로 xmax를 754로 덮어써서 다음 관찰에는 남지 않는다 |
| L6 | 통과 | 통과 | 통과 | S1의 INSERT가 끝날 때 AFTER 트리거로 `RI_FKey_check`(ri_triggers.c 452행 → `ri_PerformCheck` 2584행 → SPI)가 돌고, 그 아래 `ExecLockRows` → `heap_lock_tuple`이 부모 (0,1)을 `KeyShare`로 잠갔다. 그 단계 뒤 부모 (0,1)의 xmax는 753에 `HEAP_XMAX_KEYSHR_LOCK`·`HEAP_XMAX_LOCK_ONLY`다 |
| L7 | 통과 | 통과 | 통과 | S2는 잠들지 않았다. `HeapTupleSatisfiesUpdate`가 `TM_BeingModified`(xmax 753)를 돌려준 뒤 `compute_new_xmax_infomask`(heapam.c 5771행)가 `MultiXactIdCreate(753 ForKeyShare, 754 NoKeyUpdate)`를 불러 1번을 받았다. 옛 버전 (0,1)의 xmax는 1에 `HEAP_XMAX_IS_MULTI`, 새 버전 (0,2)의 xmax는 753에 key share 잠금만이다 |
| L8 | 통과(줄 번호 주석) | 같음 | 통과(줄 번호 주석) | S2는 번호 754를 받은 뒤 heapam.c 2990행의 `XactLockTableWait(753, Delete)`로 기다렸다. S1의 COMMIT이 깨운 뒤 행을 지웠고(xmax 754, `HEAP_KEYS_UPDATED`), `RI_FKey_noaction_del` → `ri_restrict`(893행)가 자식 (0,1)을 찾아 오류를 냈다. 754는 중단으로 적혔고 부모 행은 남았다. `errfinish`를 부른 자리는 ri_triggers.c **2799행**이다. 사전 등록의 2801행은 같은 `ereport` 문의 `errmsg` 줄이다(실험 1 E7과 같은 꼴) |
| L9 | 통과(줄 번호 주석) | 같음 | 통과(줄 번호 주석) | S2는 기다리기 전에 번호 754를 받았고, 기다리는 동안 페이지에 xmin 754인 (0,2)가 있다. `_bt_doinsert`의 nbtinsert.c 227행에서 `XactLockTableWait(753, InsertIndex)`로 잠들었다. 깨어나 처음부터 다시 찾고(`goto search`) 두 번째 `_bt_check_unique`(210행에서 부름)가 오류를 냈다. 부른 자리는 nbtinsert.c **666행**이고 사전 등록의 668행은 `errmsg` 줄이다. (0,2)는 죽은 버전으로 남았고, 다음 SELECT가 `HEAP_XMIN_INVALID`를 썼다 |
| L10 | 통과 | 통과 | 통과 | 같은 자리에서 기다리다 S1의 ROLLBACK(753 중단 기록)이 깨웠고, S2의 INSERT는 `INSERT 0 1`로 끝났다. 마지막 SELECT에는 (0,2) 하나만 보인다 |
| L11 | 통과 | 통과 | 통과 | S2는 `ExecInsert`(nodeModifyTable.c 1137행) → `ExecCheckIndexConstraints`(execIndexing.c 647행) → 896행의 `XactLockTableWait(753, InsertIndex)`로 기다렸고, S1의 COMMIT 뒤 `INSERT 0 0`으로 끝났다. 힙에 S2의 줄이 생기지 않았고 S2는 번호를 받지 않았다(번호를 받은 사건이 없고, 다음 SELECT의 스냅숏 xmax가 754다) |
| L12 | 한 조각 다름 | 한 조각 다름 | **실패(시나리오 작성 실수)** | E4: 열여섯 벌의 시나리오 세션 백엔드에서 짝지은 묶음(32~43개)마다 uprobe 사건 수와 함수 기록 수가 같다. E5: 정상 줄 수 대조는 2-deadlock의 마지막 SELECT에서만 다르고(4번 대 5줄, 인덱스 스캔), 보인 줄 대조는 L3의 한 단계에서만 다르다. 앞쪽은 위 방법 1의 실수다 |
| L13 | 통과 | 통과 | 통과 | 판정된 줄마다 행 머리 주소가 하나다. 두 세션이 모두 판정한 줄(acct (0,1)·(0,2), parent (0,1))도 세션끼리 주소가 같다. 단계 사이에 새로 켜진 힌트 비트는 열여섯 벌 모두 같은 주소에 쓴 `SetHintBits`로 설명된다(값으로 맞춘 것 0, 설명되지 않는 비트 0). 실험 1 E8의 한계(값 맞추기)는 이 실험부터 없다 |

사전 등록에 없던 관찰:

1. **기다리다 깨어난 세션은 판정을 다시 하지 않고 힌트 비트로 결론을 낸다.** L5의 교착 S2와 같은 길을 2-fk-delete의 S2도 지났다(`UpdateXmaxHintBits` ← `heap_delete`, 부모 (0,1)에 `HEAP_XMAX_INVALID`). 2-fk-delete에서 S2가 기다린 xmax 753은 지운 쪽이 아니라 S1의 key share 잠금이다. `UpdateXmaxHintBits`는 잠금만인 xmax면 그 트랜잭션이 커밋했어도 `HEAP_XMAX_INVALID`를 쓴다(heapam.c 1999~2005행). 지운 쪽이 없다는 뜻이다. 고유 키 시나리오 셋의 S2는 깨어난 뒤 같은 키의 첫 행을 `HeapTupleSatisfiesDirty`(`heap_hot_search_buffer` 아래)로 다시 읽으며 힌트를 썼다. 첫 요청이 커밋했으면 `HEAP_XMIN_COMMITTED`(heapam_visibility.c 853행), 취소됐으면 `HEAP_XMIN_INVALID`(858행)다. 2-for-update의 S2도 깨어난 뒤 (0,1)에 `HEAP_XMAX_COMMITTED`(`HeapTupleSatisfiesUpdate`)를 쓰고, 사슬의 다음 버전 (0,2)를 `heap_fetch` → `HeapTupleSatisfiesDirty`로 읽으며 xmin 커밋과 xmax 무효(S1의 잠금만인 xmax)를 썼다.
2. **중단된 버전은 HOT 사슬에서 떨어진다.** 2-deadlock에서 S1의 첫 UPDATE가 만든 (0,3)(xmin 753)은, 753이 중단된 뒤 S2가 같은 옛 버전 (0,1)을 다시 바꾸면서 (0,1)의 `t_ctid`를 (0,5)로 덮어써 어느 사슬에서도 가리키지 않게 됐다. 인덱스 스캔은 인덱스가 가리키는 사슬의 뿌리 (0,1)·(0,2)에서 사슬을 따라 (0,1)·(0,5)·(0,2)·(0,4)만 판정했다. 순차 스캔이었다면 (0,3)까지 다섯 줄을 판정했을 것이다(추론, 이번에 그 SELECT는 돌리지 않음).
3. **실패한 외래 키 DELETE가 자식 행에 잠금을 남긴다.** 2-fk-delete의 `ri_restrict`는 자식 행을 `FOR KEY SHARE`로 잠그며 찾는다. 그래서 오류 뒤 자식 (0,1)의 xmax는 중단된 754의 key share 잠금이다. 잠근 트랜잭션이 끝났으므로 이 잠금은 아무것도 막지 않는다.
4. **MultiXact xmax의 잠금 비트는 가장 센 구성원을 적는다.** 2-fk-nokey의 옛 버전 (0,1)에는 `HEAP_XMAX_IS_MULTI`와 함께 `HEAP_XMAX_EXCL_LOCK`이 있고 `HEAP_XMAX_LOCK_ONLY`가 없다. `GetMultiXactIdHintBits`가 구성원 중 가장 센 모드(754의 NoKeyUpdate)로 `HEAP_XMAX_EXCL_LOCK`을 켜고(heapam.c 7762행), 바꾼 구성원이 있으므로 `LOCK_ONLY`를 켜지 않는다(7769행).
5. **교착 검사 타이머가 울린 시각이 시나리오마다 조금 다르다.** 기다림 시작부터 SIGALRM까지 2-deadlock은 두 벌 모두 1,000.1 ms, 2-for-update는 1,005.6·1,002.9 ms였다. 실험 1은 1,000.0~1,000.1 ms였다. 차이의 원인은 확인하지 않았다.

재현: 실험 1과 같은 명령이다. 판정 표의 값은 해석 결과의 `steps[].requests[].events`(kind `tuple_lock`, `multixact`, `deadlock_check`의 `result`, `hint`), `steps[].check`, `joins`, `tupleAddrs`, `hintChanges`에서 읽는다.

### 실험 2 화면(22:35)

화면은 [`concept-6-locks.html`](../../open_source_internals/workbench/design/concept-6-locks.html)이고 실험 1과 같은 템플릿을 쓴다. 더한 것은 생명선 둘(행 머리를 보는 “공유 버퍼”의 잠금 사건, MultiXact), 표가 여럿인 페이지, 시나리오의 쉼 표시, 사건 셋(행 잠금, MultiXact, 판정 밖의 힌트 쓰기)이다. 열여섯 벌의 118장면을 모두 그려 센 결과 빨간 표지는 0개다. 원인이 기록에 있는 차이 둘은 빨간 표지 대신 설명을 붙인다. 2-for-update의 S2 SELECT는 “psql 결과 행 = 기다린 뒤 잠근 최신 버전 (0,2) · 스냅숏으로 보인 줄 (0,1)”, 2-deadlock의 마지막 SELECT는 “인덱스 스캔: HOT 사슬에 걸린 줄만 판정 (5개 중 4개)”이다. 설명의 근거는 각각 `heap_lock_tuple`의 `TM_Updated`와 판정 스택의 `heap_hot_search_buffer`다.

화면을 맞추며 고친 결함 셋과, 실험 3 전에 고칠 것 하나:

1. **그림의 판정 규칙을 판정마다, 판정하던 순간의 행 머리로 견준다.** 이전에는 단계 뒤의 관찰과 첫 판정의 스냅숏으로 모든 줄을 견줬다. 그러면 같은 단계에서 판정 뒤에 자기가 바꾼 것(2-for-update S1의 UPDATE)과, 스냅숏 둘을 쓴 단계(2-fk-delete에서 깨어난 DELETE와 그 뒤 외래 키 검사)가 거짓 경보를 냈다. `HeapTupleSatisfiesMVCC` uprobe가 함께 읽은 xmin·xmax·infomask와 그 판정의 스냅숏으로 견주게 고쳤다.
2. **장면이 다음 장면의 일을 미리 보이지 않게 했다.** 관찰은 단계가 끝나거나 기다리기 시작할 때만 묻는다. 게다가 관찰 하나는 한 순간이 아니다. txlab은 페이지, `pg_stat_activity`, `pg_locks`, 스냅숏, 커밋 로그 상태를 질의 다섯 개로 차례로 묻고, 관찰용 접속도 추적을 받아 느리다. 관찰용 접속이 질의를 받은 strace 시각으로 재면, 실험 2의 관찰 104개(벌마다 끝을 모르는 마지막 관찰은 뺌)에서 페이지는 묻기 시작한 뒤 중앙값 104 ms·최대 441 ms 안에, 세션 상태와 락은 중앙값 367 ms·최대 935 ms 안에 다 읽었다. 그래서 2-deadlock에서 S1이 오류를 받는 장면이, 그 뒤 깨어난 S2가 만든 새 버전 (0,5)를 이미 보였다. S2가 기다리기 시작하는 장면의 세션 상태는 그보다 뒤인 S1의 오류 뒤에 읽혔다. 해석기가 관찰마다 부분별로 읽은 구간(`captures[].reads`)을 내고, 화면은 부분(페이지, 세션 상태, 락)마다 “이 장면의 마지막 일과 그 부분을 다 읽은 때 사이에 다른 장면의 일이 있으면” 그 부분만, 이 장면의 마지막 일 전에 다 읽은 마지막 관찰로 물러선다. 커밋 로그의 상태는 기록된 커밋·중단 시각으로 고친다. 물러선 장면은 어느 부분이 몇 ms까지 읽은 관찰인지 적는다. 물러선 장면은 실험 2에서 열네 개(시나리오 여섯에서 벌마다 일곱 장면, 교착은 두 장면), 실험 1에서 네 개(rc-write·rr-write의 S2 COMMIT, 깨운 세션이 COMMIT 단계의 관찰 전에 이미 새 버전을 씀)다. 실험 1 화면의 그림 다섯 장은 이 장면을 쓰지 않아 그대로다. 물러선 부분은 그 장면 끝보다 이른 상태다. 교착에서 S1이 오류를 받는 장면의 세션 상자는 아직 기다리는 S1을 보인다. 미래를 보이는 것보다 덜 틀리지만, 정확한 그림은 아니다. 장면을 모두 그려 세는 검사(이 세션의 탐색용 스크립트, 커밋하지 않음)에 “고른 관찰의 부분이 뒤 장면의 일을 담음”을 실패 조건으로 더했고 0건이다.
3. **깨어난 뒤 다른 사건 없이 끝나는 단계도 깨어난 뒤 장면을 갖는다.** 해석기가 판정 밖의 `SetHintBits`를 요청의 사건(kind `hint`)으로 남기고 행 머리 주소로 표와 ctid를 붙인다(관찰하지 않는 표의 것은 버림, 시험 `TestLabelHintsKeepsOnlyWatchedRows`는 거르는 줄을 지우면 실패하는 것까지 확인). 그래서 2-deadlock의 S2와 2-idem-onconflict의 S2가 깨어난 뒤의 일을 제 장면으로 보인다.

그림 여덟 장: [덮어쓰기](../../open_source_internals/workbench/design/concept-6-lost-update.png)(S2가 기다림 없이 50을 씀), [FOR UPDATE의 기다림](../../open_source_internals/workbench/design/concept-6-for-update-wait.png)(1초 교착 검사 뒤 다시 잠듦), [FOR UPDATE의 깨어난 뒤](../../open_source_internals/workbench/design/concept-6-for-update-after.png)(이미 바뀐 (0,1)에서 (0,2)로 가서 잠그고 번호를 받음), [교착](../../open_source_internals/workbench/design/concept-6-deadlock.png)(S1의 검사가 교착을 찾고 오류, 중단, S2를 깨움), [교착의 S2](../../open_source_internals/workbench/design/concept-6-deadlock-s2.png)(`UpdateXmaxHintBits`로 753의 중단을 보고 바로 바꿈), [외래 키와 MultiXact](../../open_source_internals/workbench/design/concept-6-fk-nokey.png), [멱등성 키의 기다림](../../open_source_internals/workbench/design/concept-6-idem-wait.png)(S2의 줄 2가 이미 힙에 있고, 락 테이블에서 753을 기다림), [멱등성 키 커밋](../../open_source_internals/workbench/design/concept-6-idem-commit.png)(깨어나 같은 키를 다시 찾고 오류).

실험 3 전에 고칠 것: 관찰이 수백 ms에 걸치는 것은 실험 3에서 더 무겁다. `synchronous_commit = off`에서 보려는 틈은 WAL 라이터 주기(`wal_writer_delay` 기본 200 ms)이고, 커밋 직후 `kill -9` 전후를 견주는 관찰이 200 ms 넘게 늦으면 그 틈의 전후 증거가 되지 못한다. 그래서 실험 3의 사전 등록 전에 관찰을 줄이거나(필요한 부분만, 한 질의로) 관찰용 접속의 함수 기록을 끄는 방법을 정하고, 부분별 읽은 구간을 판정에 쓴다. 또 결과를 보는 SELECT에는 인덱스가 걸린 열로 `ORDER BY`를 쓰지 않는다(위 방법 1).

재현: `txlab` 디렉터리에서 `go run . view -tmpl ../../../tx-view.tmpl.html -out ../../../concept-6-locks.html 2-lost-update/a=<해석 결과> ...`이다.

### 실험 3 기록 열 벌(23:16~23:33)

시나리오 다섯을 두 번씩 기록했다. 한 벌은 3-rollback이 70 MB, 크래시 여덟 벌이 112~117 MB이고, 열 벌 모두 ftrace 넘침과 버린 사건이 0이다. 트랜잭션 번호, 관찰마다 읽은 WAL 위치(삽입, 쓰기, 내림), 복구가 다시 실행한 레코드의 목록(시작과 끝 위치, 자원 관리자, 종류, 번호)은 다섯 시나리오 모두 a와 b가 같다. a와 b가 다른 것은 시각과 pid, 그리고 기본 주기 200 ms로 도는 WAL writer가 WAL을 몇 조각으로 나눠 썼는가다(3-rollback, 3-crash-sync-on). `wal_writer_delay`를 10초로 둔 비동기 셋은 쓴 범위까지 같다.

기록과 커밋의 순서(KST):

| 시각 | 한 일 |
| --- | --- |
| 23:14 | 도구 시험 3-crash-sync-on 한 벌(판정에 쓰지 않음). 준비가 번호를 셋 써서 23:16 `27c999de6`로 준비를 한 트랜잭션에 묶었다(위 「실행 전 고침」의 준비 SQL) |
| 23:16 | 도구 시험 3-rollback 두 벌(판정에 쓰지 않음). 진행 중인 753을 `future`로 남겼다(아래 벗어난 것 1) |
| 23:16~23:18 | 3-crash-sync-on, 3-crash-sync-off, 3-crash-sync-off-wait를 a, b로 기록 |
| 23:20 | 3-rollback a, b를 고친 txlab으로 다시 기록. 그 고침은 23:21에 `7d65cec55`로 커밋했다. 기록이 커밋보다 1분 앞선다 |
| 23:29 | 해석기 `baec0987f`(서버 프로세스, 복구, WAL 쓰기). D10의 두 실행이 예상과 다른 것을 이것으로 봤다 |
| 23:31:55 | wait2 사전 등록 커밋 `d42e3e4b4`. 그다음 명령으로 기록을 시작했다(기록 로그 파일이 만들어진 시각도 23:31:55, 서버 시작 줄은 23:31:58) |
| 23:32~23:33 | 3-crash-sync-off-wait2 a, b |
| 23:52 | 해석기 `cd9cfe850`(서버 프로세스의 스택, 함수 구간, kill 때 WAL을 어디까지 썼는가). 기록은 그대로 두고 열 벌을 다시 해석했다. 아래 판정은 이 해석 결과로 했다 |

사전 등록에서 벗어난 것 셋:

1. **스냅숏 xmax로 커밋 로그 질문을 거른 전제가 틀렸다.** 사전 등록은 “`pg_xact_status`는 다음에 줄 번호 이상이면 오류다”에서 “그러니 스냅숏 xmax 이상은 묻지 않는다”로 넘어갔다(위 「근거로 읽은 소스」의 미래 번호, 「방법의 변경」의 관찰, 「실행 전 고침」의 커밋 로그). 그런데 xmax는 다음에 줄 번호가 아니라 마지막으로 끝난 번호 + 1이다(procarray.c 2239, 2248~2249행). 3-rollback은 S1이 BEGIN으로 753을 쥔 채 관찰하므로 xmax가 753이고, 진행 중인 753이 `future`로 남았다. 관찰이 묻는 번호는 그 관찰이 페이지와 세션 상태에서 막 읽은 것이라 늘 다음에 줄 번호보다 작다. 그래서 거름을 없앴다(`7d65cec55`). 크래시 여섯 벌(23:16~23:18)은 옛 거름으로 기록했지만 영향이 없다. 크래시 시나리오의 S1 INSERT는 자동 커밋이라 어느 관찰 때에도 753은 이미 끝나 있었고, kill 뒤의 관찰은 복구가 끝난 페이지를 읽으므로 잃은 번호가 페이지에 없다. 열 벌의 해석 결과 어디에도 `future`가 없는 것으로 확인했다.
2. **wait2를 더했다.** D10이 두 실행 모두 예상과 다르게 나온 뒤, 원인 가설과 D13, D14를 실행 전에 사전 등록했다(위 「추가 사전 등록」).
3. **해석기를 기록 뒤에 늘렸다.** 서버 사건(`baec0987f`)과 스택, 함수 구간, kill 때의 WAL 범위(`cd9cfe850`)는 기록을 마친 뒤 만든 것이다. 기록 자체(uprobe, ftrace 함수 기록, strace, server.log)는 바꾸지 않았고, 판정의 값은 모두 그 기록에서 읽는다.

판정:

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| D1 | 통과 | 통과 | 통과 | S1의 ROLLBACK 요청 안에서 `XLogSetAsyncXactLSN(0/179D378)`과 `TransactionIdAbortTree(753)`가 불렸고, 그 요청에 pg_wal의 `pwrite64`도 `fdatasync`도 없다 |
| D2 | 통과 | 통과 | 통과 | ROLLBACK 앞뒤 관찰의 페이지가 같다. (0,1)은 xmin 752, xmax 753, ctid (0,3)이고 (0,2)와 (0,3)은 xmin 753이며, 플래그도 같다. 바뀐 것은 커밋 로그의 753(진행 중 → 중단)뿐이다 |
| D3 | 통과 | 통과 | 통과 | S2의 SELECT는 세 줄을 판정해 (0,1)만 보였고, (0,1)에 `HEAP_XMAX_INVALID`, (0,2)와 (0,3)에 `HEAP_XMIN_INVALID`를 썼다. `XLogSaveBufferForHint`는 불리지 않았다 |
| D4 | 통과 | 통과 | 통과 | 취소 레코드의 끝은 0/179D378이다. ROLLBACK 뒤 124.4 ms(a), 192.5 ms(b)에 WAL writer가 `pwrite64`로 0/179C000~0/179E000을 썼고, 그 범위가 0/179D378을 덮는다 |
| D5 | 통과 | 통과 | 통과 | S1의 INSERT 요청 안에서 `XLogFlush(0/179D330)` → `pwrite64`(a 0/179A000~0/179E000, b 0/179C000~0/179E000) → `fdatasync`(a 44.5 ms, b 1.6 ms) → `TransactionIdCommitTree(753)` 순서다. 그 뒤 관찰에서 삽입과 내림이 모두 0/179D330이다. 쓴 범위의 시작이 다른 것은 그 앞을 WAL writer가 어디까지 썼는가가 달라서다. 끝 0/179E000은 커밋 레코드가 든 8 KB 페이지의 끝이다(덜 찬 페이지도 통째로 씀) |
| D6(on, wait, wait2) | 통과 | 통과 | 통과 | S1이 SIGKILL로 끝난 뒤 postmaster가 남은 자식 아홉에게 SIGQUIT을 보냈고(아래 관찰 4), server.log의 다섯 줄이 이 순서로 나왔다. 새 startup이 `ApplyWalRecord`로 on 27개, wait 23개, wait2 28개를 다시 실행했다. kill에서 접속 재개 줄까지 754~907 ms다. 사전 등록의 “여섯 실행”은 wait2를 더해 여덟 실행이 됐다 |
| D6(off) | 실패 | 실패 | **실패(사전 등록의 모순)** | SIGQUIT, “terminated by signal 9”, “reinitializing”, “ready to accept connections”는 나왔다. 그러나 “redo starts at”과 “redo done at” 대신 “invalid record length at 0/17826B0”과 “redo is not required”가 나왔고, 다시 실행한 레코드가 0개다. D8은 kill까지 아무도 pg_wal에 쓰지 않는다고 예상했고 실제로 그랬다(내림 위치가 서버 시작 때의 0/17826B0 그대로). 그러면 디스크의 WAL은 마지막 체크포인트 레코드에서 끝나고, 복구는 그 다음 레코드를 읽지 못해 “redo is not required”를 낸다(xlogrecovery.c 1748~1754, 1918행, 주석 “there are no WAL records following the checkpoint”). D6은 이 시나리오에서 처음부터 일어날 수 없는 줄을 예상했다. 판정은 고치지 않는다 |
| D7 | 통과 | 통과 | 통과 | 다시 실행한 27개 가운데 마지막 셋이 753의 Heap INSERT(0/179D1D8), Btree INSERT_LEAF(0/179D280), Transaction COMMIT(0/179D308~0/179D330)이다. COMMIT을 다시 실행할 때 새 startup이 `TransactionIdCommitTree(753)`를 불렀다(스택 `PerformWalRecovery` → `ApplyWalRecord`). 재시작 뒤 S2의 SELECT에 (0,1) kim과 (0,2) lee가 보이고, S2의 INSERT는 754를 받았다 |
| D8 | 통과 | 통과 | 통과 | S1의 INSERT 요청 안에 `XLogFlush`도 pg_wal `fdatasync`도 없고, `XLogSetAsyncXactLSN(0/179D330)`과 `TransactionIdAsyncCommitTree(753, 0/179D330)`가 있다. 응답은 `INSERT 0 1`이다. 그 뒤 관찰에서 삽입은 0/179D330, 쓰기와 내림은 0/17826B0이고, kill까지 WAL writer의 pg_wal 쓰기가 없다. wait와 wait2의 INSERT 요청도 같다 |
| D9 | 통과 | 통과 | 통과 | 다시 실행한 레코드가 없으니 753도 없고, startup은 `TransactionIdCommitTree`를 부르지 않았다. 재시작 뒤 S2의 SELECT에는 (0,1)만 보인다. 응답까지 받은 커밋을 잃었다. S2의 INSERT는 753을 다시 받아 (0,2)에 들어갔다(마지막 SELECT: (0,1) xmin 752 kim, (0,2) xmin 753 park) |
| D10 | 실패 | 실패 | **실패(알려진 위험이 일어남)** | 11초 쉬는 동안 옛 WAL writer는 한 번 깨어났다. INSERT가 끝나고 6.09 s(a), 6.11 s(b) 뒤 `pwrite64` 0/1782000~0/179C000과 `fdatasync`가 있고, 다음 주기 전에 kill이 왔다. 커밋 레코드 끝 0/179D330은 쓴 범위 밖이다. 복구는 “invalid magic number 0000 … LSN 0/179C000”에서 멈추고 “redo done at 0/179B298”을 냈다(23개). 행을 잃었고 S2의 INSERT는 753을 다시 받았다. 원인은 실행 전 고침의 “알려진 위험”이다. S1 INSERT의 첫 레코드(카탈로그 페이지의 힌트 비트 페이지 이미지, 0/179B2D0~0/179D1D8, 7,913바이트)가 8 KB 경계 0/179C000을 넘었고, `XLogBackgroundFlush`의 첫 주기는 다 찬 페이지 경계까지만 쓴다(xlog.c 2992~3002행). 이 설명은 wait2(D13, D14)로 확인했다 |
| D11 | 통과 | 통과 | 통과 | 크래시 여덟 벌 모두, 재시작 뒤 S2의 첫 SELECT가 acct 페이지에 처음 힌트를 쓸 때 `XLogSaveBufferForHint`가 한 번 불려 0이 아닌 LSN을 돌려줬고, 같은 페이지의 다음 힌트에서는 불리지 않았다. 판정한 줄은 행이 살아남은 넷(on, wait2)에서 둘, 잃은 넷(off, wait)에서 하나다 |
| D12 | 통과 | 통과 | 통과 | 시나리오 세션 백엔드는 열 벌 모두 짝지은 묶음(3-rollback 33개, 크래시 37개)마다 uprobe 사건 수와 함수 기록 수가 같고, SIGKILL로 죽은 S1도 그렇다. 세션이 아닌 프로세스(postmaster, startup, WAL writer, checkpointer 등)는 벌마다 묶음 1~3개가 맞지 않아, 그 사건은 ftrace 시각을 쓴다(해석 결과의 `clock: ftrace`). 짝지은 사건의 두 시계 차이는 최대 편차가 여섯 벌에서 15 µs 미만, 네 벌(off-a, wait-a, wait2-b, on-b)에서 약 1.0 ms다 |
| D13 | 통과 | 통과 | 통과 | 옛 WAL writer가 두 번 썼다. INSERT 뒤 6.05 s(a), 5.99 s(b)에 0/1782000~0/179C000, 16.08 s, 16.03 s에 0/179C000~0/179E000이고, 둘 다 바로 뒤에 `fdatasync`가 있다. 두 번째 쓰기가 0/179D330을 덮는다 |
| D14 | 통과 | 통과 | 통과 | 복구가 0/179C000을 지나 28개를 다시 실행했다. 753의 Heap INSERT, Btree INSERT_LEAF, COMMIT 다음에 Standby RUNNING_XACTS(0/179D330~0/179D368, 아래 관찰 5)가 있고, “invalid record length at 0/179D368”에서 멈췄다. startup이 `TransactionIdCommitTree(753)`를 불렀다. 재시작 뒤 S2의 SELECT에 (0,2) lee 200이 보이고 S2의 INSERT는 754를 받았다 |

사전 등록에 없던 관찰:

1. **문장보다 먼저 약 99 KB의 WAL이 쌓여 있다.** 모든 벌에서 첫 단계 전 관찰의 삽입 위치가 이미 0/179B2D0이고 내림 위치는 0/17826B0이다. 그 사이 101,408바이트는 레코드 23개로, 카탈로그 페이지의 힌트 비트 페이지 이미지(XLOG FPI_FOR_HINT) 15개와, 읽는 김에 페이지의 죽은 줄을 정리한 기록(Heap2 PRUNE_ON_ACCESS) 8개다. 모두 번호가 없다(복구가 다시 실행한 목록에서 확인). 데이터 체크섬이 켜져 있으면 체크포인트 뒤 깨끗한 페이지에 처음 힌트를 쓸 때 페이지 전체를 WAL에 남긴다(위 근거, bufmgr.c 5542행). 누가 넣었는지는 이번에 가리지 않았다. 카탈로그 페이지라는 점에서, 서버가 뜬 뒤 첫 접속들이 카탈로그를 읽으며 남긴 것으로 추론한다. S1의 INSERT도 첫 레코드가 카탈로그 페이지 이미지(0/179B2D0~0/179D1D8, 7,913바이트)이고 그 뒤에 Heap INSERT(166바이트)가 온다. 행 하나를 넣는 INSERT의 WAL이 8 KB를 넘는 이유이고, 이 이미지가 페이지 경계를 넘어 D10을 갈랐다.
2. **읽기만 하는 SELECT도 WAL을 남기고, 그 위치를 WAL writer에 맡긴다.** 재시작 뒤 S2의 첫 SELECT는 acct 페이지의 힌트 비트 때문에 페이지 이미지를 남기고(D11), 이어 `XLogSetAsyncXactLSN`을 불렀다. 번호를 받지 않은 트랜잭션은 WAL을 썼어도 `synchronous_commit`과 관계없이 내리지 않고 위치만 알린다(xact.c 1401~1405, 1498~1523행). 그 SELECT 뒤 관찰에서 삽입 위치만 앞서고 내림 위치는 그대로다(wait2-a: 삽입 0/17AD168 → 0/17AD210, 내림 0/179D3E0).
3. **복구는 끊긴 레코드 자리를 표시하고 나서 WAL을 이어 쓴다.** wait 두 벌에서 복구는 0/179B2D0의 레코드(0/179C000을 넘는 7,913바이트)를 끝까지 읽지 못하고 멈췄다. 그 뒤 startup이 `StartupXLOG`(xlog.c 6156행)에서 `CreateOverwriteContrecordRecord`를 불러, 이어짐이 빠진 페이지 0/179C000의 첫 레코드로 XLOG_OVERWRITE_CONTRECORD를 넣고 0/179C048까지 내렸다(`XLogFlush` → `pwrite64` 0/179C000~0/179E000 → `fdatasync`). 함수 주석(xlog.c 7455~7468행)에 따르면, 끊긴 레코드의 시작으로 되돌아가 새 WAL을 쓰면 그 WAL을 이미 받은 물리 복제 서버가 되감을 수 없다. 그래서 끊긴 자리에 이 레코드를 쓰고 페이지 머리에 `XLP_FIRST_IS_OVERWRITE_CONTRECORD`를 켜서, 읽는 쪽이 끊긴 레코드를 건너뛰게 한다. on, off, wait2에는 이 호출이 없다. 끊긴 레코드가 없어서다.
4. **postmaster는 자식마다 SIGQUIT을 두 번 보낸다.** 크래시 여덟 벌 모두 `HandleChildCrash` → `HandleFatalError` → `TerminateChildren` → `signal_child`가 남은 자식 아홉에게 SIGQUIT을 pid로 한 번, 그 프로세스 그룹(`kill(-pid)`, postmaster.c 3451행)으로 한 번, 모두 18번 보냈다. 자식은 모두 “exited with 2”로 끝났다. 그룹으로도 보내는 것은 자식이 띄운 프로세스까지 끝내려는 것으로 추론한다(이번 기록에 그런 손자 프로세스는 없음).
5. **wait2의 마지막 레코드는 세션이 쓴 것이 아니다.** 다시 실행한 28번째 레코드는 번호 0인 Standby RUNNING_XACTS(0/179D330~0/179D368)다. 21초를 쉬는 동안 다른 프로세스가 넣었고, WAL writer의 두 번째 쓰기가 함께 썼다. bgwriter가 15초마다 남기는 진행 중 트랜잭션 목록(bgwriter.c 70행 `LOG_SNAPSHOT_INTERVAL_MS`)으로 추론한다. 넣은 프로세스는 기록으로 가리지 않았다.
6. **재시작 시간의 대부분은 공유 메모리 만들기와 끝 체크포인트다.** kill에서 “ready to accept connections”까지 여덟 벌이 753~964 ms다. 함수 구간으로 나누면 postmaster의 `CreateSharedMemoryAndSemaphores`가 334~486 ms, startup의 `StartupXLOG`가 339~487 ms다. `StartupXLOG` 안에서 `SyncDataDirectory`(정상 종료가 아니어서 데이터 디렉터리 전체를 fsync)가 105~257 ms, `PerformWalRecovery`가 0.5~23 ms(다시 실행한 레코드가 없는 off는 0.5~0.6 ms), 끝 체크포인트를 기다리는 `RequestCheckpoint`가 173~210 ms다. 자식의 죽음을 처리하는 `HandleChildCrash`는 1.6~7.2 ms다. 체크포인트의 `CheckPointBuffers`가 159~178 ms인데, 쓴 버퍼가 0개인 off에서도 159~169 ms다. 버퍼 헤더를 모두 훑는 비용에 함수 추적 비용이 얹힌 것으로 추론한다. 이 시간들은 모두 함수 추적을 켠 채 잰 것이고, 추적이 얼마나 부풀렸는지는 재지 않았다.
7. **체크포인트는 버퍼를 쓰기 전마다 WAL을 먼저 내린다.** 끝 체크포인트가 쓴 버퍼는 off 0개, wait 16개, on과 wait2 19개다(server.log의 “checkpoint complete: wrote N buffers”). off의 0개는 다시 실행한 것이 없어 더러운 버퍼가 없어서다. 버퍼마다 `BufferSync` → `SyncOneBuffer` → `FlushBuffer`가 그 페이지의 LSN까지 `XLogFlush`를 불렀다(bufmgr.c 4354~4371행, 주석 “This implements the basic WAL rule”). 버퍼 수만큼 `wal_flush` 사건이 있지만 그 아래 `pwrite64`는 없다. 다시 실행한 레코드는 이미 디스크의 WAL에 있어서다. `pwrite64`와 `fdatasync`는 체크포인트 레코드를 넣고 내릴 때(xlog.c 7258행) 한 번씩 있다.
8. **“Failed process was running”은 끝난 문장을 보인다.** 크래시 여덟 벌 모두 server.log의 DETAIL이 “Failed process was running: INSERT INTO acct VALUES (2, 'lee', 200);”이다. 이 INSERT는 kill 전에 응답(`INSERT 0 1`)까지 끝났고, S1은 놀고 있었다. postmaster는 죽은 백엔드의 활동 글을 통계 공유 메모리에서 읽는다(`pgstat_get_crashed_backend_activity`, postmaster.c 2809행). 공식 문서의 `pg_stat_activity.query` 설명대로 놀고 있는 백엔드의 이 글은 마지막으로 실행한 질의다. 그래서 이 줄은 “죽을 때 돌던 문장”이 아니라 “마지막 문장”으로 읽어야 한다.
9. **복구가 읽는 WAL은 kill 전에 커널에 넘긴 범위다.** SIGKILL은 프로세스만 끝내고 운영 체제의 페이지 캐시는 그대로 두므로, `pwrite64`로 커널에 넘긴 범위는 다시 뜬 서버가 읽는다. WAL 버퍼에만 있던 레코드는 postmaster가 공유 메모리를 새로 만들 때 사라진다. 해석기가 낸 kill 때까지의 쓴 범위 끝(`kills[].walWritten`)은 on 0/179E000, wait 0/179C000, wait2 0/179E000이고 off는 쓰기가 없다. 복구가 멈춘 자리(on 0/179D330, wait 0/179C000, wait2 0/179D368)는 모두 그 끝 안이다.

한계:

- 운영 체제나 전원이 죽는 경우는 일으키지 않았다. 게다가 이번 기록에서는 pg_wal의 `pwrite64` 바로 뒤에 늘 `fdatasync`가 있었다. 커널에 넘겼지만 디스크에 내리지 않은 WAL이 생긴 적이 없으므로, 쓰기와 내림의 차이는 이 실험으로 가를 수 없다. 해석기가 내림 범위(`walSynced`)를 따로 내지만 이번에는 늘 쓰기 범위와 같다.
- 재시작 시간은 함수 추적을 켠 채 잰 것이다(관찰 6).
- ftrace 시각을 쓴 사건은 strace 시각과 최대 약 1 ms 어긋날 수 있다(D12의 네 벌).
- 이 맥의 OrbStack Linux(aarch64) 한 곳, PostgreSQL 18.6 한 판, 데이터 체크섬을 켠 initdb 기본 설정에서 본 것이다. 체크섬을 끄거나 `full_page_writes`, `wal_level`이 다르면 레코드의 수와 크기가 달라지고, D10처럼 페이지 경계에 걸리는지도 달라진다.

재현: 실험 1과 같은 명령이고 해석에 `-log <디렉터리>/server.log`를 더한다. 판정의 값은 해석 결과의 `server.events`(kind `signal`, `exit`, `log`, `redo`, `commit_clog`, `wal_io`, `wal_flush`), `server.spans`, `kills[].walWritten`, `captures[].wal`, `steps[].requests[].events`(kind `io`, `wal_flush`, `async_lsn`, `commit_clog`, `abort_clog`, `hint_fpi`), `steps[].output`에서 읽는다.

### 실험 3 화면(00:07)

화면은 [`concept-7-durability.html`](../../open_source_internals/workbench/design/concept-7-durability.html)이고 실험 1, 2와 같은 템플릿을 쓴다. 더한 것은 넷이다.

1. **kill 장면.** kill에서 접속 재개까지를 장면 하나로 그린다. 줄은 서버 프로세스마다 하나이고 postmaster, 옛 자식, 새 자식 순서다. 그 위에 끝남 표지(SIGKILL은 ✕, 그 밖은 종료 코드), SIGQUIT 화살표(pid로 보낸 것은 실선, 그룹은 점선), 함수 구간(`HandleChildCrash`에서 `ProcessSyncRequests`까지 열두 함수, 안긴 깊이만큼 겹침), 다시 실행한 레코드의 눈금(번호가 있는 레코드는 파랑), 커밋 로그 기록, server.log 줄의 번호 표지를 그린다. 시간 축은 구간(kill, 옛 자식이 끝남, 공유 메모리, 복구, 끝 체크포인트)마다 제 비율로 그린다. 한 비율로는 2 ms짜리 신호와 400 ms짜리 공유 메모리 만들기를 함께 볼 수 없어서다.
2. **WAL 막대.** kill 장면의 오른쪽에 kill 전에 넣은 범위, 커널에 넘긴 범위, 다시 실행한 레코드, 복구가 멈춘 자리를 LSN 축 하나에 그리고, 커밋 레코드 끝이 든 8 KB 페이지를 확대해 보인다. 그 아래 server.log 줄을 번호와 쉬운 설명으로 붙인다.
3. **WAL writer의 배경 쓰기.** 쉬는 동안 WAL writer가 쓴 것을 쉼 줄 아래에 적고, WAL 레인으로 들어오는 점선으로 그린다. 그 쓰기가 커밋 레코드를 덮는지도 적는다.
4. **미니맵.** WAL 상자에 넣은 곳, 쓴 곳, 내린 곳과 그 차이를 보이고, 죽은 세션과 아직 접속 전인 세션은 점선 상자로 그린다.

kill 장면의 판정 표지는 커밋 레코드를 다시 실행했으면 녹색이고, 아니면 빨간 표지 대신 설명을 붙인다. 커밋을 잃는 것은 이 실험이 보이려는 결과이지 기록의 오류가 아니어서다.

장면을 모두 그려 세는 검사(실험 2와 같은 탐색용 스크립트, 커밋하지 않음)에 “kill 장면에 WAL 막대가 없음”을 실패 조건으로 더했다. 실험 3은 56장면(kill 장면 8)에 빨간 표지 0개, 빈 장면 8개(모두 스냅숏을 찍지 않은 단계)다. 설명 표지 넷은 off와 wait의 kill 장면에 붙은 “753의 커밋 레코드(끝 0/179D330)는 kill 전에 커널에 넘어가지 않아 복구가 읽지 못했습니다”다. 실험 1(72장면)과 실험 2(118장면)의 검사 결과는 화면을 바꾸기 전과 같다(장면 수, 빈 장면, 빨간 표지, 설명, 물러선 장면). 두 페이지에서 달라진 것은 pg_wal 쓰기가 단계의 사건으로 보이게 된 것이고, 크기가 1,468 KB → 1,574 KB, 2,395 KB → 2,552 KB로 늘었다. 실험 1, 2의 그림(concept-5, concept-6 PNG)은 다시 찍지 않아 이 줄이 없다.

그림 여섯 장(어두운 화면, 1480×1250): [취소](../../open_source_internals/workbench/design/concept-7-rollback.png)(ROLLBACK 뒤에도 줄이 남고 커밋 로그만 중단), [동기 커밋 뒤 kill](../../open_source_internals/workbench/design/concept-7-crash-sync-on.png)(복구가 753의 COMMIT을 다시 실행), [비동기 커밋 바로 뒤 kill](../../open_source_internals/workbench/design/concept-7-crash-sync-off.png)(커널에 넘긴 WAL이 없고 “redo is not required”), [11초 뒤 kill](../../open_source_internals/workbench/design/concept-7-crash-wait.png)(WAL writer가 0/179C000까지만 써서 커밋 레코드 앞에서 멈춤), [21초 뒤 kill](../../open_source_internals/workbench/design/concept-7-crash-wait2.png)(두 번째 쓰기가 커밋 레코드를 덮어 753이 살아남음), [재시작 뒤 첫 SELECT](../../open_source_internals/workbench/design/concept-7-restart-select.png)(비동기 커밋 바로 뒤 kill의 재시작 뒤, 읽기만 하는 SELECT가 카탈로그와 acct 페이지의 이미지를 WAL에 남기고 그 위치만 WAL writer에 맡김).

재현: `txlab` 디렉터리에서 `go run . view -tmpl ../../../tx-view.tmpl.html -out ../../../concept-7-durability.html 3-rollback/a=<해석 결과> ...`이다.

### 실험 4 기록 여덟 벌(00:43~00:45)

시나리오 넷을 두 번씩 기록했다. 한 벌은 73~74 MB이고, 여덟 벌 모두 ftrace 넘침과 버린 사건이 0이다. 트랜잭션 번호는 모두 준비 752, S1의 UPDATE 753, S2의 UPDATE 754로 사전 등록과 같다. 네 시나리오 모두 a와 b가 SSI 사건(락 요청과 그 결과, 쓰기 검사의 대상과 간선, 커밋 검사의 플래그, 정리), 관찰마다의 SIRead 락, psql 출력까지 같다(pid를 세션 이름으로 바꿔 견줌). 다른 것은 시각과 pid뿐이다.

기록과 커밋의 순서(KST):

| 시각 | 한 일 |
| --- | --- |
| 00:35 | 사전 등록 커밋 `015076847`, 그 안의 시각 표기 고침 `2680256c0` |
| 00:41 | 도구 시험 4-skew-ssi 한 벌(판정에 쓰지 않음). 「실행 전 고침」의 셋을 여기서 찾았다 |
| 00:43:05 | uprobe·txlab·해석기 커밋 `9fdafb0c0`(실행 전 고침 포함) |
| 00:43:17~00:45:14 | 여덟 벌 기록과 해석 |
| 01:08 | 해석기 `298bd37c1`(boolean 열을 t·f로 풂). 여덟 벌을 다시 해석했고, 바뀐 것은 페이지 관찰의 on_call 값(벌마다 29개)뿐이다. 판정은 두 해석 결과에서 같다 |

사전 등록에서 벗어난 것 둘:

1. **W3의 함수 이름이 틀렸다.** heapam.c 1171행은 `initscan`이 아니라 `heap_beginscan` 안이다(스택 `table_beginscan` → `heap_beginscan` → `PredicateLockRelation`). 실행 전 도구 시험에서 찾아 「실행 전 고침」에 적었다. 예상의 뜻(순차 스캔이 표 단위 락을 잡는다)은 그대로라 판정은 통과로 두고 주석을 단다.
2. **해석기를 기록 뒤에 고쳤다.** boolean 값 풀이(`298bd37c1`)다. 판정에 쓰는 값(사건, 관찰의 락과 커밋 로그, psql 출력)은 바뀌지 않았다.

판정:

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| W1 | 통과 | 통과 | 통과 | 두 SELECT 모두 2, 두 UPDATE 모두 `UPDATE 1`, 두 COMMIT 모두 `COMMIT`이다. 마지막 SELECT는 (0,3) alice f(xmin 753)와 (0,4) bob f(xmin 754)다. 당직이 0명이다 |
| W2 | 통과 | 통과 | 통과 | `GetSerializableTransactionSnapshotInt`, 새로 만든 SIRead 락, 대상 검사, `FlagRWConflict`가 모두 0번이다. `PredicateLockRelation`은 두 SELECT에서 한 번씩, `CheckForSerializableConflictIn`은 두 UPDATE에서 한 번씩 불렸고, 두 COMMIT의 커밋 검사는 SERIALIZABLEXACT 포인터가 0이었다(`sx=0`). 관찰 열 번 모두 `SIReadLock`이 없다 |
| W3 | 통과 | 통과 | 통과(함수 이름 주석) | 각 세션의 첫 SELECT 안에서 `GetSerializableTransactionSnapshotInt`가 한 번 불렸고(`GetTransactionSnapshot` → `GetSerializableTransactionSnapshot` 아래), `heap_beginscan` 아래 `PredicateLockRelation(doctors)`가 표 단위 락을 새로 만들었다. S2의 SELECT 뒤 관찰에 표 doctors의 `SIReadLock`이 S1·S2 하나씩이다 |
| W4 | 통과 | 통과 | 통과 | S1의 UPDATE 스택에 `_bt_first`가 있다. `PredicateLockPage(doctors_pkey, 1)`이 새 락을 만들었고, `PredicateLockTID(doctors, (0,1))`은 두 번 불려 둘 다 만들지 않았다(아래 관찰 1). 대상 검사는 행 (0,1), 페이지 0, 표 순서로 세 번이고 표 검사에서만 `FlagRWConflict`를 불렀다. 읽은 쪽은 S2(번호 0), 쓴 쪽은 S1(753)이고 둘의 플래그는 비어 있다. 오류는 없다 |
| W5 | 통과 | 통과 | 통과 | S2의 UPDATE에서 같은 길로 표 검사가 `FlagRWConflict(읽은 쪽 S1 753, 쓴 쪽 S2 754)`를 불렀다. 둘의 플래그는 비어 있고 오류는 없다 |
| W6 | 통과 | 통과 | 통과 | S1의 커밋 검사 입구의 플래그는 비어 있고, `COMMIT`과 `TransactionIdCommitTree(753)`이다. S2의 커밋 검사 입구의 플래그는 DOOMED이고, predicate.c 4723행에서 오류를 내 psql에 오류와 이유 코드(“Canceled on identification as a pivot, during commit attempt.”)가 나왔으며 `TransactionIdAbortTree(754)`가 불렸다. W5의 `FlagRWConflict`가 읽은 S2의 플래그는 비어 있다. S2에 DOOMED를 켠 곳은 아래 관찰 2의 소거로 S1의 커밋 검사(4765행)까지 좁혔다 |
| W7 | 통과 | 통과 | 통과 | 마지막 SELECT는 (0,2) bob t(xmin 752, xmax 754는 중단)와 (0,3) alice f(xmin 753)다. 당직이 한 명 남는다 |
| W8 | 통과 | 통과 | 통과 | 관찰의 `SIReadLock`이 S2의 UPDATE 뒤 넷(세션마다 표 doctors와 doctors_pkey 1번 페이지)이고 S1의 COMMIT 뒤에도 넷이다. 커밋한 S1의 둘이 S1의 pid로 그대로 보인다. S2의 COMMIT 요청 안에서 `ClearOldPredicateLocks`가 불렸고(아래 관찰 3), 그 뒤 관찰 둘에 `SIReadLock`이 없다 |
| W9 | 통과 | 통과 | 통과 | 4-skew-ssi-row: 첫 SELECT마다 `PredicateLockPage(doctors_pkey, 1)`과 `PredicateLockTID`가 락을 하나씩 새로 만들었다(S1은 bob (0,2), S2는 alice (0,1)). 관찰 열 번에 표 단위 `SIReadLock`이 없다. 4-ssi-own도 같은 꼴이다(S1은 (0,1), S2는 (0,2)) |
| W10 | 통과 | 통과 | 통과 | S1의 UPDATE에서 `heap_hot_search_buffer` 아래 `PredicateLockTID(doctors, (0,1))`이 S1의 행 락을 새로 만들었고(이어진 `heap_fetch` 아래 호출은 만들지 않음), 행 검사가 `FlagRWConflict(S2, S1)`을 불렀다. 페이지와 표 검사는 간선을 긋지 않았다. 그 단계 뒤 관찰에 S1의 (0,1)은 없고 S2의 (0,1)은 있다. S2의 UPDATE도 같은 꼴이다((0,2), `FlagRWConflict(S1, S2)`). 커밋과 마지막 SELECT는 W6, W7과 같다 |
| W11 | 통과 | 통과 | 통과 | 두 UPDATE 모두 대상 검사 셋이 간선 없이 끝났고 `FlagRWConflict`는 0번이다. 각 UPDATE 뒤 관찰에 그 세션의 행 락이 없다. 두 COMMIT 모두 `COMMIT`이고 커밋 검사 입구의 플래그가 비어 있다. 마지막 SELECT는 alice f(753), bob f(754)다 |
| W12 | 통과 | 통과 | 통과 | SERIALIZABLE 여섯 벌의 UPDATE 요청에 `_bt_doinsert`가 없고, `CheckForSerializableConflictIn`의 표는 늘 doctors이며, `CheckForSerializableConflictOut`은 0번이다 |
| W13 | 통과 | 통과 | 통과 | 시나리오 세션 백엔드의 짝지은 묶음(4-skew-rr 45개, 4-skew-ssi 53개, 4-skew-ssi-row 54개, 4-ssi-own 52개)마다 uprobe 사건 수와 함수 기록 수가 같다. 세션이 아닌 프로세스는 벌마다 묶음 하나가 맞지 않아 그 사건은 ftrace 시각을 쓴다. 짝지은 사건의 두 시계 차이는 최대 편차가 4-skew-rr-a에서 4.6 µs, 나머지 일곱 벌에서 약 1.0 ms(999.9~1,007.1 µs)다 |

사전 등록에 없던 관찰:

1. **UPDATE도 바꿀 행을 찾으며 SIRead 락을 요청한다.** UPDATE는 기본 키로 행을 찾으며 B-tree 잎 페이지(`PredicateLockPage`)와 행(`PredicateLockTID`)을 잠근다. 행 락 요청은 두 번이다. 한 번은 인덱스가 가리킨 행을 찾을 때(`heap_hot_search_buffer`, heapam.c 1813행)이고, 또 한 번은 새 행을 만들려고 옛 행을 다시 읽을 때다(`ExecModifyTable` → `table_tuple_fetch_row_version` → `heap_fetch`, nodeModifyTable.c 4533행의 주석 “Fetch the most recent version of old tuple”과 4541행의 호출, heapam.c 1666행). 4-skew-ssi-row에서는 첫 요청이 자기 행 락을 새로 만들고, 같은 UPDATE의 쓰기 검사가 곧바로 지웠다(W10). RR에서는 이 셋도 입구에서 돌아왔다.
2. **DOOMED를 켠 곳을 소거로 좁혔다.** 소스 트리 전체에서 `SXACT_FLAG_DOOMED`는 정의(predicate_internals.h)와 predicate.c에만 나오고, predicate.c에서 플래그에 DOOMED를 켜는 줄은 넷이다(3470, 4108, 4682, 4765행). S2의 플래그는 단계 6의 `FlagRWConflict` 입구에서 비어 있었고, 단계 8의 커밋 검사 입구에서 DOOMED였다. 그 사이를 줄마다 보면 다음과 같다.
    - 3470행은 끝나는 트랜잭션이 커밋이 아닐 때 자기 플래그에 켠다. S2는 단계 8 전에 끝나지 않았다.
    - 4108행은 `CheckForSerializableConflictOut` 안이다. 이 함수는 한 번도 불리지 않았다(W12).
    - 4682행은 간선을 그을 때의 검사(`OnConflict_CheckForSerializationFailure`)다. 함수 기록에 두 UPDATE마다 `FlagRWConflict` → `OnConflict_CheckForSerializationFailure` → `SetRWConflict`가 한 번씩 있다. 이 검사가 실패를 찾으면 쓴 쪽이 자기일 때 그 자리에서 오류를 내는데(4661~4669행), 두 UPDATE 모두 `UPDATE 1`이었다.

    남는 것은 단계 7에서 S1의 커밋 검사가 지나간 4765행이다. 직접 읽은 값은 아니다. 그 줄의 명령 주소에 uprobe를 걸면 닫을 수 있다.
3. **SIRead 락을 지우는 자리는 트랜잭션이 끝나는 방식에 따라 다르다.** 4-skew-ssi와 4-skew-ssi-row에서는 S2의 COMMIT 문장이 오류로 끝난 뒤 오류 처리의 `AbortTransaction` → `ResourceOwnerRelease` → `ReleasePredicateLocks` → `ClearOldPredicateLocks`가 지웠다. 4-ssi-own에서는 S2의 `CommitTransaction` 안에서 같은 `ReleasePredicateLocks` → `ClearOldPredicateLocks`가 지웠다. 그 요청의 함수 기록에는 `ReleaseOneSerializableXact`가 두 번(S1과 S2의 몫으로 추론), `SetNewSxactGlobalXmin`이 한 번 있다. 먼저 끝난 S1의 COMMIT에서도 `ReleasePredicateLocks`는 불렸지만 `ClearOldPredicateLocks`는 불리지 않았다. S2가 아직 돌고 있어서다(W8).
4. **간선을 만든 것은 인덱스 락이 아니다.** B-tree는 0번 블록이 메타페이지라서(nbtree.h 149행 `BTREE_METAPAGE 0`), 키 둘인 doctors_pkey의 하나뿐인 잎(뿌리를 겸함)은 1번이다(추론, 인덱스 페이지를 직접 읽지는 않음). 쓰기 검사는 행이 든 표의 0번 페이지를 보고 이 인덱스 페이지는 보지 않았다. UPDATE가 HOT이라 인덱스에 키를 넣지 않았기 때문이다(W12). 그래서 4-skew-ssi의 간선은 표 단위 락에서, 4-skew-ssi-row의 간선은 행 락에서 나왔다.
5. **기본 격리 수준의 문장도 SSI 함수를 지난다.** 마지막 SELECT(READ COMMITTED, 자동 커밋)에서도 `PredicateLockRelation`과 커밋 검사가 불렸고, 커밋 검사의 SERIALIZABLEXACT 포인터는 0이었다. SSI 함수는 늘 불리고, 격리 수준은 그 함수 입구의 검사에서 갈린다.
6. **오류는 다시 시도하라고 알린다.** psql에 오류와 이유 코드 다음 HINT “The transaction might succeed if retried.”가 나왔다(predicate.c 4727행). 오류 코드는 SQLSTATE 40001(serialization_failure, 빌드가 만든 errcodes.h 220행)이다. SERIALIZABLE을 쓰는 응용은 이 오류를 받으면 트랜잭션 전체를 처음부터 다시 실행해야 한다. 다시 실행하면 S2의 SELECT는 S1이 커밋한 결과(당직 1명)를 본다(추론, 다시 실행하지 않음). 그 뒤 빠질지 말지는 응용의 규칙이다.

한계:

- 세션 둘, 행 둘, 표 하나의 가장 작은 꼴이다. 읽기 전용 트랜잭션의 최적화(`READ ONLY`, `DEFERRABLE`), 세 트랜잭션 이상의 구조, 2단계 커밋, 락의 페이지·표 승격은 일으키지 않았다.
- S2에 DOOMED가 켜진 순간은 직접 읽지 않았다(관찰 2의 소거).
- 이 맥의 OrbStack Linux(aarch64) 한 곳, PostgreSQL 18.6 한 판에서 본 것이다(실험 3과 같음).

재현: 실험 1과 같은 명령이다. 판정의 값은 해석 결과의 `steps[].requests[].events`(kind `ssi_begin`, `pred_lock`, `conflict_in`, `conflict_target`, `conflict_out`, `rw_conflict`, `ssi_precommit`, `ssi_clear`, `index_insert`, `commit_clog`, `abort_clog`, `ereport`), `captures[].locks`(mode `SIReadLock`), `steps[].output`에서 읽는다. 관찰 2·3의 함수 호출 수는 `decode -bin <postgres> -dir <디렉터리>/trace query -pid <백엔드 pid> -nth <몇 번째 문장> -find FlagRWConflict,OnConflict_CheckForSerializationFailure,SetRWConflict,ReleasePredicateLocks,ClearOldPredicateLocks`로 센다.

### 실험 4 화면(01:08)

화면은 [`concept-8-serializable.html`](../../open_source_internals/workbench/design/concept-8-serializable.html)이고 실험 1~3과 같은 템플릿을 쓴다. 더한 것은 넷이다.

1. **SSI 레인.** 시간 순서의 MultiXact 자리를 SSI 레인으로 쓴다. 실험 1~3의 기록에는 SSI 사건이 없고 실험 4에는 MultiXact가 없어서다. SIRead 락 요청마다 대상(표, 인덱스 페이지, 행), 결과(새로 만듦, 이미 덮는 락이 있어 그대로, 내가 쓴 행이라 만들지 않음), 그 요청을 부른 함수를 적는다. 쓰기 검사는 행, 페이지, 표의 세 줄에 간선을 그었는지를 적고, 간선은 읽은 쪽 → 쓴 쪽으로 적는다. 직렬화 트랜잭션이 아니라 입구에서 돌아온 호출은 함수마다 한 줄로 모아 흐리게 그린다.
2. **미니맵의 SSI 상자.** 공유 버퍼 상자 아래에 세션마다 SIRead 락(관찰)과 원 둘 사이의 rw 간선(기록)을 그린다. 이 장면에서 사라진 락은 줄을 긋고, 커밋·DOOMED·중단을 원 아래에 적는다. SIRead 락은 락 테이블 상자에서 뺐다. pg_locks에 함께 보일 뿐 술어 락의 해시에 있기 때문이다.
3. **오른쪽의 SSI 그림.** 세션 상자(번호, 상태)와 그 아래 SIRead 락, 상자 사이의 간선(그은 단계, 이 장면의 것은 굵게)을 그리고 이 장면의 일을 풀어 쓴다. S1의 커밋 장면에는 지금의 간선이 상대에 DOOMED를 켜는 꼴이라는 것을, S2의 커밋 장면에는 DOOMED를 읽고 실패한 것과 그것을 켠 곳의 추론을 적는다. 추론은 추론이라고 적는다. DOOMED 장면은 소스로 predicate.c 4723행(검사와 오류)을 먼저 보인다.
4. **boolean 값.** 해석기가 boolean 열을 psql처럼 t·f로 푼다(`298bd37c1`). 전에는 페이지 그림에 `\x01`·`\x00`이 그대로 보였다.

검사: 장면을 모두 그려 세는 탐색용 스크립트(커밋하지 않음)로 실험 4의 72장면에 JS 오류와 빨간 표지가 0개이고, 소스가 없는 장면 16개는 모두 BEGIN임을 확인했다(BEGIN은 스냅숏도 번호도 받지 않음, 실험 1~3과 같은 꼴). 템플릿을 바꾸기 전후로 실험 1~3의 246장면을 모두 그려 미니맵, 레인, 오른쪽, psql 창 둘, 제목의 HTML 해시를 견줬고 모두 같다. 실험 1~3의 페이지도 새 템플릿으로 다시 만들었다. 담긴 해석 결과는 그대로이고 달라진 것은 템플릿 부분뿐이다.

그림 여섯 장(어두운 화면, 1480×1250): [RR의 마지막 SELECT](../../open_source_internals/workbench/design/concept-8-rr-final.png)(둘 다 커밋, 당직 0명, SSI 함수는 입구에서 돌아옴), [두 번째 UPDATE](../../open_source_internals/workbench/design/concept-8-ssi-update.png)(간선 S1 → S2가 더해져 고리가 됨), [S1의 커밋](../../open_source_internals/workbench/design/concept-8-ssi-commit-s1.png)(커밋한 S1의 SIRead 락이 남음), [S2의 커밋](../../open_source_internals/workbench/design/concept-8-ssi-commit-s2.png)(DOOMED → 오류 → 취소 → 정리), [행 락으로도 같은 간선](../../open_source_internals/workbench/design/concept-8-row-update.png)(만들자마자 쓰기 검사가 지운 자기 행 락), [자기 행만 읽음](../../open_source_internals/workbench/design/concept-8-own-update.png)(간선 없음).

재현: `txlab` 디렉터리에서 `go run . view -tmpl ../../../tx-view.tmpl.html -out ../../../concept-8-serializable.html 4-skew-rr/a=<해석 결과> ...`이다.

### 실험 5 기록 열두 벌(10-15 18:35~18:47)

다섯 장면을 두 번씩(a, b) 기록했고, 5-requires-new-pool1은 관찰 결함을 고친 뒤 두 번 더(c, d) 기록했다. 한 벌은 10~18초, 76~92 MB이고, 열두 벌 모두 ftrace 넘침과 버린 사건이 0이다. JDK는 GraalVM 25.0.4다. 트랜잭션 번호는 모두 준비 752, 바깥 753, 안쪽 754로 같다(REQUIRED와 풀 1개 장면은 안쪽이 번호를 받지 않음). a와 b는 메시지의 종류·순서·SQL, 서버 사건, 관찰 값, 예외까지 같고, 다른 것은 시각과 pid뿐이다.

기록과 커밋의 순서(KST):

| 시각 | 한 일 |
| --- | --- |
| 18:29 | pgjdbc 42.7.13 jar와 소스 jar 내려받기, sha1 확인(사전 등록 절의 「pgjdbc 소스 확인」) |
| 18:31 | pgjdbc 소스 확인. 추론 하나(`BEGIN`의 묶음)가 틀림 |
| 18:32 | 시험 한 벌 5-required(판정에 쓰지 않음) |
| 18:35:02~18:37:08 | 다섯 장면 a, b 열 벌 기록 |
| 18:39~18:44 | 열 벌 훑기. 5-requires-new-pool1 a·b의 `blocked` 관찰 결함을 찾음(아래 1) |
| 18:45~18:46 | 고침(시나리오의 `blocked:` 줄, 풀 1개 장면은 0.5초). 기록기를 18:46:45에 다시 빌드 |
| 18:46:45~18:47:08 | 5-requires-new-pool1 c, d 기록 |
| 18:48~19:02 | 해석기 `decode spring` 작성, 열두 벌 해석(마지막 해석 19:01:52~19:02:00) |
| 19:03 | 커밋 `751a6a5d5`(위 고침), `dfad88c1f`(해석기) |

사전 등록에서 벗어난 것 넷:

1. **풀 1개 장면의 `blocked` 관찰이 시간 초과와 겹쳤다(도구 결함).** 지휘는 go 뒤 1.5초에 기다림을 묻고 관찰을 시작했는데, 묻는 데 0.27초, 관찰에 0.4초쯤 걸렸다. 연결 기다림은 2초에 끝나므로, a에서 관찰 구간은 go 뒤 1.77~2.17초이고 그 사이 2.007초에 ROLLBACK이 지나갔다. 그래서 a·b의 `blocked` 관찰에서 바깥 연결이 `idle`, 753이 `aborted`로 읽혔다. 해석기의 겹침 점검(`captures[].overlap`)이 이 관찰에 메시지 8개(ROLLBACK 묶음과 그 답)를 낸다. 시나리오에 `blocked:` 줄을 두어 그 장면은 go 뒤 0.5초에 찍게 고쳤고(`751a6a5d5`), c·d의 `blocked` 관찰은 겹침이 0이다. a·b의 P6 가운데 `blocked` 관찰을 묻는 부분은 판정에 쓰지 않고, 나머지 부분은 네 벌 모두로 판정한다.
2. **해석기를 기록 뒤에 썼다.** 해석기를 쓰며 버그 둘을 찾아 고쳤다. pgjdbc가 `BEGIN`과 첫 문장을 한 번에 보내면 두 요청이 하나로 합쳐지던 것, 관찰 구간의 끝을 다음 관찰의 시작으로 잡아 멈춤 자리 관찰에도 다음 단계의 메시지가 겹친 것으로 나오던 것이다. 둘 다 회귀 시험으로 남겼다(`TestSplitRequestsPairsPipelinedBatches`, `TestCaptureSpansIgnoresOtherQueries`). 판정 값은 고친 해석기의 출력이다. 해석기를 쓰기 전의 훑기(스크래치패드 스크립트, 커밋하지 않음)를 열두 벌에 다시 돌려 견주니(19:12) P1~P11의 판정이 모두 같았고, 다른 것은 셋이었다. 빌리기 간격과 UPDATE 왕복이 1~4 ms 다르다(기준 시각을 달리 잡음). 훑기의 P6 창은 `outer-threw` 알림까지라 ROLLBACK 묶음이 들어가고, 해석기는 예외까지라 들어가지 않는다(아래 4). 훑기는 `ROLLBACK TO SAVEPOINT`도 트랜잭션의 끝으로 셌다(P8, 5-nested).
3. **P11의 “같은 메서드”를 (클래스, 이름, 서술자)로 읽었다.** 시험 한 벌에서 오버로드가 사건 둘을 남기는 것을 보고 본 기록 전에 적었다(사전 등록 절의 「시험 한 벌」).
4. **P6의 “그 뒤 C1에 ROLLBACK”은 문구가 모호했다.** 실제 순서는 시간 초과 예외 → 1.1~4.3 ms 뒤 C1에 ROLLBACK → 14~22 ms 뒤 앱의 `outer-threw` 알림이다. Spring이 프록시 안에서 롤백을 마친 뒤에야 예외가 `Main`에 닿기 때문이다. 문구를 “`outer-threw` 알림 뒤”로 읽으면 순서가 반대이고, “앱의 예외 뒤”로 읽으면 맞다. 판정은 예외가 만들어진 시각(`java.throws`의 `SQLTransientConnectionException`) 기준으로 하고 이 모호함을 적어 둔다.

판정:

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| P1 | 통과 | 통과 | 통과 | 열두 벌, 접속 스무 개 모두 SSLRequest → `N` → StartupMessage(`user=postgres`, `database=postgres`, 프로토콜 3.0) → `R`(0) → `K` → `Z`(I)이고, 모두 `start` 전에 열렸다. 첫 접속은 `main` 스레드, 둘째는 HikariCP의 `lab connection adder` 스레드가 열었다 |
| P2 | 한쪽만 | 한쪽만 | 조건 하나만 판정 | 장면의 첫 빌리기는 열두 벌 모두 직전 사용 뒤 782~1,205 ms였고, 모두 `isConnectionDead`가 불리고 `BEGIN` 앞에 빈 문장 왕복(`Q ""` → `I` → `Z`)이 있었다. 500 ms 이하인 경우는 한 번도 일어나지 않아 “둘 다 없다”는 쪽은 판정하지 못했다. REQUIRES_NEW 장면의 둘째 빌리기(C2)도 1,251~1,645 ms 뒤라 같은 왕복이 있었다 |
| P3 | 통과 | 통과 | 통과 | 바깥과 안쪽 UPDATE가 모두 C1이고 그 사이에 `BEGIN`·`SAVEPOINT`가 없다. C2는 장면 동안 메시지가 없다. `inner-after-update` 관찰에서 번호는 C1의 753 하나이고 두 행의 옛 판 xmax가 모두 753이다. 안쪽 실패의 go부터 `outer-after-inner`까지 연결 위 메시지가 0개이고, JFR에 `doSetRollbackOnly`가 한 번 있다. 그 뒤 C1에 `ROLLBACK`이 가고 `outer-threw`는 `UnexpectedRollbackException: Transaction rolled back because it has been marked as rollback-only`다. 마지막에 753 aborted, 잔액 100, 200이다 |
| P4 | 통과 | 통과 | 통과 | 안쪽은 C2다. `inner-after-update` 관찰에서 C1 753, C2 754가 둘 다 `idle in transaction`이고 옛 판 xmax는 1번 행 753, 2번 행 754다. 트랜잭션은 C2의 `COMMIT`이 C1의 `COMMIT`보다 먼저 끝났다. 마지막에 둘 다 committed, 잔액 90, 210이다 |
| P5 | 통과 | 통과 | 통과 | `blocked` 관찰에서 C2는 `active`, `Lock`/`transactionid`이고 753에 대한 `ShareLock`을 기다리며, C1은 `idle in transaction`, `Client`/`ClientRead`다. `waiters`는 C2 하나다. C2의 UPDATE 요청에서 `XactLockTableWait(753)` 뒤 1.006초(b 1.0055초)에 `DeadLockCheck`가 `NoDeadlock`(1)을 돌려주고, 3.004초 뒤 postgres.c 3437행에서 오류를 낸다. C2에 `ErrorResponse`(55P03, “canceling statement due to lock timeout”, `postgres.c:3437 ProcessInterrupts`)가 오고, C2의 `ROLLBACK`, C1의 `ROLLBACK` 순서다. `outer-threw`는 `UncategorizedSQLException` ← `PSQLException`이다. 마지막에 753, 754 aborted, 잔액 100, 200이다 |
| P6 | 판정 제외·통과 | 판정 제외·통과 | 통과(c·d) | 접속은 C1 하나다. `outer-after-update`의 go부터 시간 초과 예외까지 연결 위 메시지와 새 접속이 없다. `HikariPool.getConnection(long)`이 2,005~2,012 ms 걸렸고, 그 안에서 `jdk.ThreadPark` 2,003~2,005 ms가 `ConcurrentBag.borrow` 151행 아래에 있다. 예외 글은 “lab - Connection is not available, request timed out after 2,004~2,007ms (total=1, active=1, idle=0, waiting=0)”이고 `outer-threw`는 `CannotCreateTransactionException` ← `SQLTransientConnectionException`이다. 예외 뒤 C1에 `ROLLBACK`이 간다(위 4). c·d의 `blocked` 관찰에서 C1은 `idle in transaction`이고 `waiters`는 비었다. a·b의 `blocked` 관찰은 위 1로 판정에서 뺐다 |
| P7 | 통과 | 통과 | 통과 | C1 하나로 `SAVEPOINT "SAVEPOINT_1"`, 2번 행 UPDATE, `ROLLBACK TO SAVEPOINT "SAVEPOINT_1"`, `RELEASE SAVEPOINT "SAVEPOINT_1"`, `COMMIT` 순서다(이름을 큰따옴표로 감쌈, 사전 등록 절의 소스 확인). 안쪽 UPDATE 요청에서 `AssignTransactionId` level 2 뒤 754(753 + 1)를 받았고, `inner-after-update` 관찰에서 2번 행 옛 판의 xmax가 754다. `ROLLBACK TO SAVEPOINT` 요청에서 `TransactionIdAbortTree(754, nsub 0)`, `COMMIT` 요청에서 `TransactionIdCommitTree(753, nsub 0)`이다. 마지막에 753 committed, 754 aborted, 잔액 90, 200이다 |
| P8 | 통과 | 통과 | 통과 | 열두 벌의 트랜잭션 열여섯 개 모두, 끝난 뒤 그 연결에 앱이 처음 보낸 메시지는 장면 끝의 Terminate(`X`)다. 풀에 돌려줄 때 보낸 것이 없다. JFR에 `ProxyConnection.close`가 트랜잭션마다 있다 |
| P9 | 통과 | 창 밖(0.33 ms) | 통과 조건 미충족, 실패 조건도 없음 | 접속 스무 개 모두 `X`가 가고, 양쪽 기록의 마지막이 `X`이며, 중계기가 오류 없이 끝났다. 등록한 실패 조건(`X` 없이 끊김, 중계기가 남음)은 한 번도 없었다. 다만 b의 C2 하나는 `X`를 중계기가 읽은 시각이 앱의 `closed` 알림보다 0.33 ms 늦어, 통과 조건의 창(`end`의 go부터 `closed`까지)을 문구대로는 채우지 못했다. 아래 관찰 4 |
| P10 | 통과 | 통과 | 통과 | 열두 벌, 접속 스무 개 모두 두 기록이 같은 순번·종류·길이로 잘렸고, 차이 구간이 비지 않으며 양 끝이 [−0.274, 0.259] ms 안이다 |
| P11 | 통과 | 통과 | 통과 | 하나로 합친 사건 묶음은 `invokeWithinTransaction`, `processRollback`, `JdbcTemplate.execute(StatementCallback, boolean)`, `HikariPool.getConnection(long)`뿐이고, 넷 모두 메서드 안의 throw를 같은 메서드의 finally가 받는다(TransactionAspectSupport 377~389행, AbstractPlatformTransactionManager 876·937~943행, JdbcTemplate 390·408·410행, HikariPool 159·179·185행). 스택에만 보인 호출은 모두 예외가 지나가며 끝난 호출이다(아래 관찰 5). 예외가 없는 5-requires-new에서는 둘 다 0이다 |

사전 등록에 없던 관찰:

1. **pgjdbc가 연결 위에 보내는 것.** 접속 뒤 바로 `SET application_name = 'PostgreSQL JDBC Driver'`를 단순 질의로 보낸다(pgjdbc `SetupQueryRunner`). 트랜잭션의 `BEGIN`은 단순 질의이고, `COMMIT`·`ROLLBACK`은 처음 쓸 때부터 이름 있는 준비 문장(`S_1`)의 확장 질의다(Parse 이름 `S_1` → Bind → Execute → Sync). 왜 문턱(`prepareThreshold` 5)을 기다리지 않는지는 소스로 확인하지 않았다. 나머지 문장(UPDATE, `SET LOCAL`, 저장점)은 이름 없는 확장 질의다.
2. **HikariCP는 첫 연결에서만 드라이버 지원을 확인한다.** 첫 연결을 풀에 넣을 때 `PoolBase.checkDriverSupport`(443~444행)가 `checkValidationSupport`(460행)로 빈 문장을 보내고 `checkDefaultIsolation`(481행)으로 `SHOW TRANSACTION ISOLATION LEVEL`을 보낸다. 채우기 스레드가 연 둘째 연결은 `SET application_name`만 보낸다.
3. **락 시간 초과는 기다림이 시작된 때부터 3초다.** `XactLockTableWait`부터 오류를 내기까지 3.004초였고, 앱이 UPDATE를 보내고 오류를 받기까지는 3.109초(b 3.167초)였다. 그 사이 교착 검사는 한 번(1초 뒤)만 돌았다.
4. **P9의 0.33 ms는 관측 지연으로 본다(추론).** HikariCP는 풀을 닫을 때 연결을 목록에서 빼고(HikariPool 432행) 실제 닫기는 따로 스레드에 맡기지만(434~435행), `shutdown`은 그 스레드가 끝나기를 기다린 뒤에 돌아온다(240~241행). 그래서 `X`를 쓰는 일은 `closed` 알림 전에 끝났고, 0.33 ms는 맥의 중계기가 그 바이트를 읽은 시각이 늦은 것으로 본다. 앱이 `X`를 쓴 시각은 기록하지 않았다. `org.postgresql.jdbc.PgConnection::close`를 JFR 추적에 더하면 직접 볼 수 있다.
5. **JFR이 남기지 않은 호출(스택에만 보인 호출, `java.stackOnly`).** 5-required: 바깥 `invokeWithinTransaction`, 두 `TransactionInterceptor.invoke`, `AbstractPlatformTransactionManager.commit`, `commitTransactionAfterReturning`(모두 커밋 단계의 `UnexpectedRollbackException`이나 안쪽 예외가 지나감). 5-requires-new-same-row: 두 `invoke`, `Outer.requiresNewSameRow`, `Inner.requiresNewSameRow`, `JdbcTemplate.update`. 5-requires-new-pool1: 두 `invoke`(20:22 고침, 처음에는 하나로 적음. 호출 트리 시험판이 마디 둘을 냄), `Outer.requiresNew`, 안쪽 `invokeWithinTransaction`·`createTransactionIfNecessary`·`getTransaction`, `HikariDataSource.getConnection`, `HikariPool.getConnection()`. 5-nested: 안쪽 `invoke` 하나(바깥 `invoke`는 정상으로 끝나 기록됨). 같은 메서드가 겹쳐 불릴 때 바깥 기록이 안쪽의 빈자리를 가리므로, 해석기는 스택 안의 등장 횟수와 감싸는 기록의 수를 견준다(`TestStackOnlyCountsNestedMissingCall`).
6. **uprobe와 함수 기록의 수가 다른 묶음은 WAL writer의 `XLogBackgroundFlush` 반환 하나뿐이다(벌마다).** uprobe를 켜기 전부터 돌던 프로세스라 앞부분이 없다. 세션 백엔드의 묶음은 모두 수가 같고, 요청마다의 서버 구간을 자르는 `ReadyForQuery` 수도 연결 위의 `Z` 수와 모두 같다(`conns[].join`).

한계:

- 앱 스레드 하나, 연결 두 개 이하, 행 둘의 가장 작은 꼴이다. 여러 스레드가 풀을 다투는 경우, 가상 스레드, Spring Boot의 자동 설정, `@Transactional`의 readOnly·isolation·timeout, JPA는 일으키지 않았다.
- 앱의 시스템 콜과 TCP는 보지 않았다(맥에 strace가 없음, 서버는 유닉스 소켓으로 받음). 연결의 일생은 실험 6의 몫이다.
- P2의 “500 ms 이하면 검사하지 않음”은 일어나지 않아 판정하지 못했다. 장면 안에서 같은 연결을 0.5초 안에 두 번 빌리는 장면이 있어야 한다.
- 이 맥의 OrbStack Linux(aarch64)와 macOS의 GraalVM 25.0.4, Spring Framework 6.2.15, HikariCP 5.1.0, pgjdbc 42.7.13 한 벌에서 본 것이다.

재현: 기록은 `txlab` 디렉터리에서 `go run . record -scenario ../../../../../database/postgresql/transactions/<이름>.txt -out <디렉터리>`, 해석은 `decode` 디렉터리에서 `go run . -bin <기록 빌드> -dir <디렉터리>/trace spring -rec <디렉터리> -jfr <JDK 25>/bin/jfr -src <PostgreSQL 18.6 소스>`다. 판정의 값은 해석 결과의 `conns[]`(startup, thread, params, messages, requests[].server.events, join), `borrows`, `transactions`, `captures[]`(label, activity, locks, pages, visible, xacts, overlap), `blocked`, `notes`, `clock.wire`, `java`(spans, throws, parks, dups, stackOnly)에서 읽는다.

### 실험 5 화면 시안(10-15 19:27)

시안 두 장을 사용자에게 보냈다. 둘 다 멈춘 채로 찍은 관찰(`blocked`)의 순간이다. 한 장은 5-requires-new-same-row a(서버에서 기다림), 다른 한 장은 5-requires-new-pool1 c(풀에서 기다림)다. 값은 해석 결과(`d5/*.json`)와 그 기록의 `jfr.json`(앱 코드의 줄 번호, 스택)에서 읽었고, 그림은 스크래치패드의 일회용 생성기로 만들었다(커밋하지 않음). 화면의 뼈대는 실험 1~4(concept-5~8)를 그대로 쓴다. 위 줄은 미니맵과 클라이언트 창, 아래 왼쪽은 사건 순서의 시간축, 아래 오른쪽은 선택한 줄의 자세한 내용이다.

새로 들인 것:

- **앱 쪽 칸을 라이브러리로 나눴다.** 앱 코드, Spring 트랜잭션, JdbcTemplate, HikariCP, pgjdbc 칸에 JFR 호출 기록을 세로 막대로 그리고, 그 오른쪽에 연결 위 메시지와 백엔드 C1·C2 칸을 둔다. 미니맵의 상자도 같은 가로 경계에 둬서 위아래를 이어 읽게 했다.
- **클라이언트 창은 psql 대신 앱의 Java 소스다.** 바깥(`Outer`)과 안쪽(`Inner`) 메서드에서 지금 줄을 표시한다. 풀 1칸 장면에서는 `Inner`의 실제 메서드 칸이 JFR 스택에 없고 프록시 칸만 있어, 몸통에 들어가기 전에 연결을 기다린다고 표시했다.
- **스택에만 보인 호출은 점선 막대**다. 시작은 그 호출이 스택에 처음 보인 시각이고 끝은 그리지 않는다(끝난 시각이 없음).
- **실험 도구가 앱을 멈춘 구간은 빗금 줄**이다(“관찰하려고 앱을 멈춤 396 ms”). 줄 간격은 시간 비례가 아니며, 60 ms 넘는 틈은 `⋮ 1,006 ms`처럼 적는다. 그림 아래에 서버 쪽 시간에는 함수 기록의 비용이 들어 있다고 적었다.
- **자세한 내용에는 사실과 추론을 따로 붙였다.** 같은 행 장면의 “서버가 교착으로 보지 않는 이유”(락 테이블에 받지 못한 락이 C2 → 753 하나뿐이고, 753을 끝낼 쪽은 같은 main 스레드)는 추론으로 적었다. 미니맵의 “이 순간” 상자에도 같은 표지를 붙인다(만드는 화면에서도 지킬 규칙).

시안이 해석기 밖에서 셈한 것(합의 뒤 Go 해석기로 옮길 몫, 화면은 고르고 그리기만 함):

1. 풀 칸의 상태: 빌리기가 끝난 시각과 다음 `ProxyConnection.close`를 짝지어 칸마다 쓰는 구간을 만든다.
2. 스레드에 묶인 연결과 떼어 둔 연결: 빌리기, `doSuspend`, `doResume`, `close`를 시간 순서로 다시 돌린다.
3. 어느 순간의 main 스레드 스택: 그때 열린 가장 안쪽 기록(또는 `jdk.ThreadPark`)의 JFR 스택을 쓰고, 칸마다 (이름, 서술자)로 기록·스택에만 보임·추적 안 함을 가린다. 같은 메서드가 스택에 여러 번 있으면 바깥부터 그 시각을 덮는 기록과 차례로 짝짓는다. 처음에는 이름만으로 짝지어, 풀 1칸 장면의 안쪽 `invokeWithinTransaction`과 `HikariPool.getConnection()`을 기록된 것으로 잘못 표시했다. 서술자까지 맞추니 해석기의 `stackOnly` 목록과 정확히 같아졌다.
4. 앱 코드의 줄 번호: JFR 스택의 `lab.Outer`·`lab.Inner` 칸에서 읽는다(해석기는 아직 내지 않음).
5. 요청마다 보낸 메시지 종류의 차례(`Q`, `P B D E S`)와 준비 문장 이름(`S_1`).
6. 서버 쪽 기다림 구간: `XactLockTableWait`부터 오류까지.
7. `setAutoCommit`의 값: 둘러싼 Spring 호출로 정한다. `doBegin` 안이면 false(DataSourceTransactionManager 298행), `doCleanupAfterCompletion` 안이면 true(389행)다. 롤백 뒤 true로 되돌릴 때 연결 위 메시지가 없는 것은 관측이고, pgjdbc의 `commit()`이 트랜잭션이 끝난 상태에서는 보내지 않기 때문이다(PgConnection 1043~1045행). 시안의 첫 판은 이 줄을 모두 false로 적었다가 고쳤다.

화면 조립은 사용자 의견을 받은 뒤 확정한다.

### 실험 5 화면 설계(10-15 20:30, 구현 전)

시안을 본 뒤 사용자와 앱 쪽 호출을 어떻게 보일지 세 번 주고받았다. 사용자 원문은 이렇다.

- 19:40쯤 “(가)인데 (나)의 장점을 흡수할 수는 없을까요?” (가)는 라이브러리별 칸, (나)는 호출 깊이별 칸이다. 사용자는 (가)가 간결하지만 세부 동작은 결국 다 펼쳐 봐야 알 수 있고, 다 펼치면 “평생 봐도 못 볼 거 같다”고 했다.
- 20:00쯤 “너무 간결하면 핵심적으로 하는 것들을 놓칠 거 같아서입니다. 그렇다고 안 봐도 될 것들을 다 보는 것도 낭비이고요. 문제는 그것을 누가 결정해야 하느냐인데… 그 핵심, 주요한 부분들의 동작에 대해 보는 게 중요할 거 같은데… 이게 가능한가요?”
- 20:15쯤 “호출은 그 뎁스가 명확하게 보여야 합니다. 정확히는 호출만이라기보다는… 실제 동작하는 것을 보면서 학습할 수 있어야 한다는 게 가장 중요합니다. 이렇게 당신에게 요청하는 것은 그 구체적인 방법을 위임하는 것이고요. 일단 추천 방향대로 진행해 주세요. 결과물을 보고 그 다음에 얘기하는 게 좋을 거 같습니다.”

사용자가 정한 것은 목적(실제 동작을 보며 배우기, 깊이가 분명히 보이기)과 진행(추천안으로 만들고 결과를 본 뒤 다시 논의)이다. 아래 방법은 AI가 고른 것이고, 결과를 본 사용자의 의견으로 바뀔 수 있다.

**고른 방법.** (가)의 라이브러리별 칸을 뼈대로 둔다. 무엇을 줄로 보일지는 기록에서 셈할 수 있는 기준(렌즈)으로 고르고, 줄마다 왜 보이는지를 붙인다. 렌즈는 셋이다.

- 효과: 밖에 드러나는 일을 한 호출. 연결 위로 메시지를 보냄(`QueryExecutorImpl.execute`), 풀에서 칸을 빌림·돌려줌(`ConcurrentBag.borrow`, `ProxyConnection.close`), 스레드에서 연결을 떼어 둠·다시 묶음(`doSuspend`, `doResume`), 자동 커밋을 바꿈(`setAutoCommit`), 스레드가 잠듦(`jdk.ThreadPark`), 예외를 만듦(`jdk.JavaExceptionThrow`).
- 갈림: 같은 메서드가 기록마다(또는 한 기록 안에서 부를 때마다) 다른 길을 간 자리. 아래 규칙으로 셈한다.
- 자기 시간 1 ms 이상: 자식 호출과 실험 도구의 멈춤을 뺀 시간.

렌즈에 걸리지 않은 호출은 지우지 않고 “＋ n개 접음”으로 이름만 남긴다. 사용자는 렌즈를 켜고 끌 수 있고, 특정 메서드를 “늘 보기”로 고정할 수 있다. 줄을 누르면 그 호출까지의 사슬, 곧 바깥 호출부터 그 줄까지를 펼친다. 깊이가 깊어질수록 오른쪽 칸(Spring → HikariCP → pgjdbc)으로 계단처럼 내려가고, 단마다 깊이 숫자와 부른 줄 번호가 붙는다. 사슬에 들지 않는 형제 호출은 그 자리에서 “＋ n개 접음”으로 남는다.

**깊이의 정의는 하나다.** 추적한 호출 사이의 깊이이고, 스택에만 보인 호출도 한 단으로 센다. 추적하지 않은 칸(프록시, 반사, 감싸개)은 두 단 사이에 “⋯ n칸 추적 안 함(종류)”으로 접는다. 해석기의 `java.spans[].depth`는 기록된 호출만 센 값이므로 화면에 쓰지 않는다. 한 화면에 깊이 숫자가 두 가지면 사용자가 가장 먼저 요구한 “뎁스가 명확하게”를 깨기 때문이다.

**해석기가 내는 것(`decode spring`의 `view`, 화면은 고르고 그리기만 함).**

1. 호출 트리(`view.tree`): main 스레드에서 장면 구간(`start`의 go ~ `end`)과 겹치는 호출. 마디마다 부모, 칸, 깊이, 시작·끝, 부른 줄(부모 메서드 안의 줄), 그 사이에 접힌 칸의 수와 종류, 끝난 꼴, 자기 시간, 효과, 렌즈 이유, 그 순간의 앱 코드 줄(`lab.Outer`·`lab.Inner`)이다.
2. 스택에만 보인 호출을 마디로 넣는다. 기록된 호출 X를 다른 호출 s의 스택 칸에 맞추는 조건은 셋이다. 이름과 서술자가 같고, X가 s의 시작을 감싸며, X의 스택이 s의 스택에서 그 칸 바깥 부분과 같아야 한다(겹치는 길이만큼, 줄 번호까지). 맞는 기록이 없는 칸은 스택에만 보인 마디가 되고, (부모 마디, 이름, 서술자, 부른 줄)이 같으면 같은 마디로 본다. 처음 시험판은 바깥부터 차례로 짝지어, 바깥 호출이 빠진 자리에서 안쪽 기록을 바깥 칸에 붙였다. 그래서 5-required의 `TransactionInterceptor.invoke` 안에 `invokeWithinTransaction`이 둘로 보이는 가짜 갈림이 나왔다. 스택의 바깥 부분으로 맞추자 사라졌고, 열두 벌 모두 스택에만 보인 마디의 (이름, 서술자) 집합이 해석기의 `java.stackOnly`(장면 구간 안)와 같았다(20:22, 스크래치패드 `tree5/check_so.py`).
3. 끝난 꼴(`end`): `passed`(스택에만 보임, 지나가는 예외로 끝남, P11의 관측), `twice`(같은 시작의 사건 둘, 메서드 안의 throw를 finally가 받은 뒤 끝남), `throw`(그 마디의 코드가 마지막 자식 뒤에 예외를 만듦. 예외 사건 스택의 가장 안쪽 추적 칸이 그 마디), `return`(그 밖). `jdk.JavaExceptionThrow`는 예외 객체를 만들 때 남으므로, 이미 있는 예외를 다시 던지는 것은 보이지 않는다. 그래서 `return`은 “예외의 흔적이 없음”이고 정상 반환의 증명이 아니다.
4. 자기 시간: 끝 시각이 있는 자식만 가진 마디에서만 셈한다. 길이에서 자식들의 길이를 빼고, 실험 도구의 멈춤(멈춤 자리의 도착 ~ go) 가운데 자식 밖에 놓인 부분을 뺀다. 자식 가운데 스택에만 보인 호출이 있으면 셈하지 않는다(끝 시각을 모르므로).
5. 그 밖의 셈(시안이 해석기 밖에서 하던 일곱 가지, 위 절): 풀 칸마다 쓰는 구간, 스레드에 묶인 연결과 떼어 둔 연결의 구간(빌리기·`doSuspend`·`doResume`·`close`를 시간 순서로 다시 돌림, 호출에서 다시 셈한 값이라 화면에 추론으로 표시), 요청마다 보낸 메시지 종류와 준비 문장 이름, 서버 쪽 기다림 구간, `setAutoCommit`의 값(부른 줄의 소스 글에서 읽음), 앱 코드의 소스, 부른 줄 둘레의 라이브러리 소스 발췌.

**갈림의 규칙(`decode spring-compare`).** 함께 견주는 기록 묶음에서 같은 (이름, 서술자)의 마디를 모은다. 마디의 길은 (자식 차례, 끝난 꼴의 묶음)이다. 자식 차례는 자식의 (이름, 서술자)를 시작 순으로 늘어놓되, 앱 코드(`lab.*`)는 장면마다 메서드가 달라 하나의 “앱 코드”로 본다. 끝난 꼴은 정상(`return`)과 예외(`throw`·`twice`·`passed`) 둘로 묶는다.

- 끊김: 마지막 자식이 예외로 끝났고, 그 자식 차례가 예외로 끝나지 않은 다른 길의 앞부분(같음 포함)인 길이다. 아래에서 올라온 예외 때문에 뒤가 빠진 것이므로, 그 마디가 판단한 자리가 아니다. 예: 풀 1칸 장면의 `doBegin`, `startTransaction`, 안쪽 `invokeWithinTransaction`.
- 갈림: 끊김이 아닌 길이 둘 이상인 메서드다. 길이 짧아도 마지막 자식이 정상으로 끝났으면 그 마디의 조건이 뒤를 건너뛴 것이므로 갈림이다. 예: `PgConnection.commit()`의 “아무것도 보내지 않음”, `suspend`의 “떼어 둘 것이 없음”. 같은 자식 차례라도 끝난 꼴이 다르면 다른 길이다. 예: `HikariPool.getConnection(long)`의 시간 초과(`borrow` 뒤 예외)와 0.5초 안의 재빌리기(`borrow` 뒤 정상, P13).
- 갈린 자리: 끊김이 아닌 길들의 공통 앞부분 끝이다. 마디의 소스에서 갈린 자식들을 부른 줄을 함께 낸다. JFR은 반환값을 남기지 않으므로, 무엇이 길을 정했는지는 그 줄의 소스와 다음 호출로 읽는 추론이다. 화면에도 추론으로 표시한다.
- 갈림은 함께 견준 묶음 안에서만의 사실이다. 묶음에 없는 길은 보이지 않는다.

열두 벌로 돌린 시험판(20:20)은 갈림 13곳을 냈다. `getTransaction`(이미 트랜잭션이 있나), `handleExistingTransaction`(전파 방식), `processRollback`(새 트랜잭션·저장점·참여), `commit`(rollback-only 표시), `invokeWithinTransaction`(앱 메서드가 예외로 끝났나), `cleanupAfterCompletion`·`triggerAfterCompletion`(새 트랜잭션인가), `suspend`, `PgConnection.setAutoCommit`·`commit`, `HikariPool.getConnection(long)`이고, 나머지 둘은 끊김 규칙을 넣기 전이라 풀 1칸 장면의 끊김(`doBegin`, `startTransaction`)이다. 19:50의 첫 시험판은 `invokeWithinTransaction`을 잡음으로 봤는데, 그 판단의 일부는 틀렸다. 잡음은 앱 메서드 이름과 스택에만 보인 자식의 빈자리였고, 그것을 걷어 내도 `commitTransactionAfterReturning`과 `completeTransactionAfterThrowing`의 갈림이 남는다.

한계와 다시 볼 조건: 트랜잭션은 main 스레드 하나에서만 셈한다(이번 장면이 모두 그렇다). 스택에만 보인 호출이 같은 부모의 같은 줄에서 두 번 불리면 한 마디로 합쳐진다(이번 기록에는 없음, 반복문이 있는 장면을 더하면 다시 본다). 결과를 본 사용자가 렌즈나 펼치기 방식을 바꾸자고 하면 이 절에 덧붙인다.

### 실험 5 기록 덧붙임: 5-reborrow 두 벌(10-15 20:27~20:28)

위 「덧붙인 장면 5-reborrow」의 등록대로 두 번(a, b) 기록했다. 앱 소스(`Main`의 `reborrow` 장면, `Outer.once()`)를 20:27:38에 고쳤고 기록은 20:27:58, 20:28:08에 시작했다. 그래서 이 두 벌은 지금 저장소의 앱 소스로 컴파일됐다(아래 「화면 첫 판」의 소스 사본 결함과 관계있음). 판정 값은 해석 결과의 `borrows`, `conns[].requests`, `transactions`, `captures`(`end`·`closed`), `clock.wire`, `view.branches`에서 읽었다.

| 번호 | a | b | 판정 | 관찰 |
| --- | --- | --- | --- | --- |
| P12 | 통과 | 통과 | 통과 | 두 빌리기 모두 C1을 받았다. 장면 구간에 C2 위 메시지는 0개다. 트랜잭션은 C1의 753, 754이고 둘 다 `COMMIT`, 커밋 로그 `commit_clog`이며, `end` 관찰의 잔액은 1번 행 80, 2번 행 200이다 |
| P2(나머지 쪽) | 통과 | 통과 | 통과 | 첫 빌리기는 쉰 지 1,224.7 ms(b 1,053.2 ms)라 `PoolBase.isConnectionDead`가 불리고 첫 `BEGIN` 앞에 `Q ""` 왕복이 있다. 둘째 빌리기는 쉰 지 5.7 ms(b 4.1 ms)이고, 그 안에 `isConnectionDead`가 없으며(호출 트리에서 그 마디는 첫 빌리기 아래 하나뿐), 둘째 트랜잭션의 첫 앱 메시지가 `Q BEGIN`이다 |
| P13 | 통과 | 통과 | 통과 | 열네 벌을 함께 견준 `HikariPool.getConnection(long)`의 길이 셋이다. `ConcurrentBag.borrow` → `PoolBase.isConnectionDead`, 정상(18번, 열네 벌 모두). `borrow`, 정상(2번, 5-reborrow a·b의 둘째 빌리기). `borrow`, 예외(4번, 5-requires-new-pool1 a~d). 셋 다 끊김이 아니고, 갈린 자리는 `borrow` 다음, HikariPool 168행이다 |

등록한 예상 밖의 관찰 하나: 둘째 `COMMIT`은 Parse 없이 `B E S`로 갔다. 첫 `COMMIT`이 만든 이름 있는 준비 문장 `S_1`을 같은 연결에서 다시 쓴 것이다. 그래서 그 요청은 연결 위에 SQL 글이 없고(해석 결과의 `sql`이 null), 화면은 “(SQL 없음, 준비 문장 재사용)”으로 적는다. 모든 장면에 거는 예상 가운데 이 장면에서 함께 확인한 것은 P10(시계)뿐이다. 차이 구간은 [0.600, 0.908] ms(b [0.476, 0.695] ms)로 1 ms 안이지만 열두 벌의 [−0.274, 0.259] ms보다 넓다. 왜 넓어졌는지는 보지 않았다. P1·P8·P9·P11은 이 장면에서 판정하지 않았다(P8의 문구 “끝난 뒤 처음 보낸 메시지가 장면 끝의 X”는 트랜잭션을 잇달아 두는 이 장면에 맞지 않는다).

### 실험 5 화면 첫 판(10-15 21:07)

`concept-9-spring.html`(열네 벌, 약 5 MB)과 사진 셋을 만들었다. 템플릿은 `spring-view.tmpl.html`이고, `txlab view -tmpl spring-view.tmpl.html -out concept-9-spring.html 장면/실행=해석결과…`로 묶는다. 해석 결과는 `decode spring`(기록 한 벌)과 `decode spring-compare -out DIR 장면/실행=경로…`(열네 벌을 견준 갈림을 덧붙임) 순서로 만든다.

- `concept-9-chain.png`: 5-requires-new-same-row a에서 안쪽 빌리기(`ConcurrentBag.borrow`, 깊이 13)를 고른 모습이다. 줄 그림에는 그 호출의 조상 사슬이 깊이 0(`TransactionInterceptor.invoke`, 스택에만 보임)부터 13까지 칸을 넘나드는 계단으로 펼쳐진다. 자세한 내용에는 단마다 깊이 숫자, 부모 안에서 부른 줄(`← getConnection(long):162`), 사이에 접힌 칸(`⋯ 10칸 추적 안 함(프록시·반사, JDK)`)이 붙는다. 위 그림판은 칸마다 그 순간 스택에 있던 그 칸의 호출(깊이와 이름)과 셈한 상태(묶인 연결 없음, 떼어 둔 연결 C1, 풀 C1 빌려줌)를 보인다.
- `concept-9-branch.png`: 5-reborrow a의 첫 `HikariPool.getConnection(long)`이다. 위 P13의 세 길과 “이 호출이 간 길”, 갈린 자리 168행의 소스 발췌(추론 표시)가 나온다.
- `concept-9-blocked.png`: 같은 행 장면의 멈춘 채로 찍은 관찰이다. 장면 스레드가 깊이 8의 `QueryExecutorImpl.execute` 안에 있고, 앱 코드 창은 `Inner.java` 39행(`jdbc.update`)을 칠하며, C2가 753을 기다리는 구간이 빗금으로 보인다.

줄로 보일 호출은 렌즈(효과·갈림·자기 시간 ≥ 1 ms·전부)와, 렌즈와 상관없이 늘 보이는 둘로 고른다. 늘 보이는 둘은 장면 코드(`lab.*`, 이유 [앱])와 가장 바깥 호출(이유 [뿌리])이다. 설계 때는 렌즈만 두었는데, 첫 사진에서 앱 코드 칸이 비어 깊이 계단의 출발점이 보이지 않아 더했다. 줄을 고르지 않았을 때 기본 렌즈 셋(효과·갈림·시간)으로 같은 행 a는 호출 119개 가운데 43개, 재빌림 a는 111개 가운데 41개가 줄이 된다. 줄을 고르면 그 조상 사슬이 더해진다. 나머지는 부모 줄 끝의 ＋n으로 접힌다.

설계에서 바뀐 것과 그 근거:

1. **끝난 꼴 `rethrow`를 더했다(추론).** 끝 기록은 있는데 부모가 스택에만 보이는 마지막 자식이다. 부모가 예외로 지나갔다면 그 예외는 이 자식에서 나왔어야 하는데, JFR은 예외를 만들 때만 사건을 남기므로 다시 던지기는 보이지 않는다. 풀 1칸 장면의 `handleExistingTransaction`이 `startTransaction`의 예외를 받아 `resumeAfterBeginException` 뒤 `throw beginEx`(AbstractPlatformTransactionManager 455~456행)하는 자리다. 이것 없이는 `getTransaction`의 그 길이 정상 반환으로 세어져 가짜 갈림이 됐다.
2. **스택 맞추기에 길이 조건을 더했다.** 기록된 호출의 스택이 칸 바깥 부분과 같아야 할 뿐 아니라, 둘 중 하나가 JFR에서 잘린 스택(`truncated`)이 아니면 길이도 같아야 한다. 길이를 보지 않으면 짧은 스택이 긴 스택의 앞부분과 맞아 엉뚱한 기록에 붙는다.
3. **갈림 수는 12곳이다(20:20 시험판의 13곳에서).** 끊김 규칙으로 `doBegin`·`startTransaction`이 빠지고, 5-reborrow가 `HikariPool.getConnection(long)`에 길 하나를 더했으며, `rethrow`로 `getTransaction`이 바로잡혔다. `QueryExecutorImpl.execute(…, boolean)`은 자식 차례가 같고 끝난 꼴만 다른 갈림이다(같은 행 장면의 락 시간 초과 오류).
4. 해석기의 판단마다 시험을 두고(`spring_tree_test.go` 여덟 개), 시험이 틀린 규칙을 잡는지 변이로 확인했다. 길이 조건을 빼기, 끊김 규칙을 빼기, 길의 서명에서 끝난 꼴을 빼기, `rethrow` 추론을 빼기, 부른 줄 대조의 대상을 바꾸기·추적 안 한 칸 대조를 빼기의 여섯 변이가 모두 시험에서 실패했다.

**찾아 고친 결함: 앱 코드 창이 한 줄 위를 칠했다.** 첫 사진에서 `Outer.java`가 안쪽 호출 자리로 52행(`Lab.at("outer-after-update")`)을 칠했다. 실제 호출은 그 아래 줄이었다. 원인은 이렇다. JFR 스택의 줄 번호는 기록 때 컴파일한 클래스의 것이다. 그런데 해석기는 앱 코드의 소스를 저장소에서 읽었고, 이번에 `Outer` 머리 주석이 한 줄 늘어 앞선 기록 열두 벌의 줄이 모두 한 줄씩 어긋났다. txlab은 기록마다 다시 컴파일해 “소스와 기록이 어긋날 일을 없앤다”고 적어 두었지만, 그 보장은 컴파일까지였고 해석 때 읽는 소스에는 미치지 않았다. 고친 것은 셋이다.

- txlab `record`가 앱 소스를 기록 디렉터리의 `app-src/lab/`에 복사하고 그 사본을 컴파일한다(`prepareSpring`). 기록에 남은 소스가 실행한 클래스와 같다는 것이 이것으로 보장된다.
- `decode spring`은 기록의 `app-src`를 소스 뿌리의 맨 앞에 두고, 어디서 읽었는지를 `view.javaFrom`(`record`·`repo`)에 남긴다. 그리고 앱 코드가 다음 칸을 부른 줄의 소스에 그 칸의 메서드 이름이 있는지 대조해(`checkAppSites`), 어긋난 자리를 `view.siteMismatch`와 표준 오류에 낸다. 화면은 둘 가운데 하나라도 맞지 않으면 코드 창 위에 경고를 띄운다.
- 사본이 없던 기록 열네 벌은 이렇게 채웠다. 앞선 열두 벌에는 git HEAD의 앱 소스를 넣었다. 앱 소스를 마지막으로 커밋한 것은 10-15 02:01(`a376825df`)이고 기록은 18:35~18:47이며, 그 사이 작업 트리의 앱 소스를 고친 것은 20:27이 처음이다. 5-reborrow 두 벌에는 지금 작업 트리의 소스를 넣었다. 대조는 채우기 전 열두 벌에서 벌마다 세 곳(예: `lab.Outer:31 → requiredFails`), 5-reborrow에서 0곳이었고, 채운 뒤 열네 벌 모두 0곳이다. 고친 txlab으로 새로 한 벌을 기록해(5-reborrow, 스크래치패드, 판정에 쓰지 않음) `app-src`가 남고 저장소 소스와 같으며, 저장소 소스 뿌리를 주지 않아도 해석기가 사본으로 `javaFrom: record`, 어긋남 0을 내는 것을 확인했다.

한계와 다시 볼 조건:

- 사진은 맥의 헤드리스 Chrome으로 1500×1300에서 찍었다. 어두운 바탕이 기본으로 나왔고, 밝은 바탕은 한 장만 확인했다. 좁은 화면(휴대폰 폭)은 보지 않았다.
- 파일이 5 MB인 것은 해석 결과 전체(서버 소스 발췌, 세션, JFR 원 사건)를 그대로 넣기 때문이다. 화면이 쓰지 않는 칸을 덜어내는 일은 하지 않았다.
- 효과 렌즈에 무엇이 들어가는지(어느 호출이 “밖에 드러나는 일”인가)는 해석기의 분류이고, 이 주제의 판단이다. 아래 「범용화 경계」의 주제별 몫이다.
- 사용자가 결과를 보고 렌즈, 늘 보이는 줄, 펼치기 방식을 바꾸자고 하면 이 절에 덧붙인다.

### 범용화 경계(10-15 21:10)

사용자 원문(20:50쯤): “그리고 지금은 db 위주인데 앞서 말한 것처럼 리눅스나, 그 중에서도 특히 네트워크나, 소켓 갘은 것들 그런 것들도 이런 식으로 학습할 수 있어야 하는데, 나중에 매번 그걸 다 구현하지 않아도 되게 범용화 할 수 있는 부분들의 범용화도 고려는 해둬야 합니다. 세부적인 커스커마이징이나 구현은 주제마다 다를 수 있지만요”

사용자가 정한 것은 목적이다. 리눅스, 네트워크, 소켓 같은 다른 주제도 같은 방식으로 배울 수 있어야 하고, 주제마다 처음부터 다시 만들지 않아야 한다. “고려는 해둬야”의 범위와 시점은 AI가 정했다. 지금은 경계를 정해 적고, 비용이 거의 들지 않는 분리만 한다. 공통 틀을 따로 떼어 내는 일은 둘째 주제를 시작할 때 한다. 소비자가 하나뿐인 지금 떼어 내면 Spring·PostgreSQL의 모양이 공통 계약에 굳는다. 둘째 주제의 실제 기록을 보고 맞춰야 경계가 맞는다. 이 판단이 틀렸다고 볼 조건은 아래 「떼어 낼 때」에 적었다.

지금 한 분리: 화면 템플릿을 「공통 화면」과 「주제별」(`TOPIC`) 두 덩어리로 나눴다. 공통 덩어리에는 주제의 칸 이름, 효과 종류, 연결·SQL·백엔드 같은 말이 없다(21:12에 최종 템플릿을 낱말 열일곱 개로 grep해 0건. 그 전에 남았던 `--pg` 색 하나는 공통 토큰 `--eff`로 바꿈). 공통 덩어리는 칸 목록과 `TOPIC`의 함수만 부른다.

| 층 | 주제와 상관없는 것(공통) | 주제마다 다른 것 |
| --- | --- | --- |
| 기록 | 장면을 정한 순서로 돌리고 멈춤 자리에서 세움, 멈춘 채로 상대 쪽 상태 찍기, 두 시계 맞추기, 기록에 소스 사본 남기기(txlab) | 장면 앱, 무엇으로 기록할지(지금은 JFR, PostgreSQL의 함수 기록, uprobe, 연결 중계기) |
| 해석 | 구간과 스택으로 호출 트리 만들기(깊이, 부른 줄, 접힌 칸, 끝난 꼴, 자기 시간), 렌즈 이유 가운데 시간, 기록끼리 견준 갈림, 멈춤 구간, 부른 줄 대조 | 칸 나누기(패키지 → 층), 효과 분류(무엇이 밖에 드러나는 일인가), 상대 쪽 사건의 뜻, 상태 다시 돌리기(연결 묶임, 풀 칸) |
| 화면 | 줄 고르기(렌즈, 장면 코드, 뿌리, 사슬, 펼침, 늘 보기), 깊이 숫자와 계단, ＋n 접기, 사슬, 갈림 표, 소스 창과 발췌, 사실·추론 표시, 그림판의 칸별 스택 | `TOPIC`: 칸과 색, 효과 글, 메시지·상대 쪽 사건 줄, 그림판의 상태 상자, 구간 막대 |

지금 해석기의 공통 부분은 아직 JFR에 묶여 있다. `buildCallTree`는 JFR의 구간·스택 사건을 읽고, `annotateTree`는 시간 같은 공통 셈과 Spring·HikariCP 효과 분류를 함께 한다. 화면이 받는 계약(`callNode`: 부모, 칸, 깊이, 시작·끝, 부른 줄, 접힌 칸, 끝난 꼴, 자기 시간, 효과, 이유, 길)은 기록 종류와 상관없다.

소켓·커널 주제에 필요한 입력(무엇이 필요한지만 적고, 방법은 그 주제를 시작할 때 정한다):

1. 앱 쪽 호출의 구간과 스택. JVM 앱이면 JFR을 그대로 쓴다. C나 Go 프로그램이면 함수 입구·출구 기록이 필요하다. 지금 PostgreSQL에 쓰는 컴파일러 계측이나 uprobe가 후보다.
2. 커널 안의 호출. 시스템 콜 안에서 커널이 부른 함수다. 이 저장소는 이미 OrbStack Linux 머신의 ftrace(function_graph)로 `pwrite64`·`fdatasync` 안을 기록하고 화면에 그렸다(concept-4의 커널 보기, `feasibility/fulltrace/kernel.sh`, [관측 재설계 WORK](WORK_20261013_WORKBENCH_OBSERVATION_DESIGN.md)). function_graph는 입구와 출구가 짝지어진 호출이라 지금의 마디 계약에 맞을 것으로 본다(추론, 맞춰 보지 않음).
3. 오간 바이트와 소켓 상태. 지금의 연결 중계기는 PostgreSQL 프로토콜을 해석하지만, 바이트를 시각과 함께 남기는 부분은 프로토콜과 상관없다. 소켓의 상태(연결 상태, 보낸·받은 큐)나 패킷을 볼 수단과 그 권한 조건은 확인하지 않았다. 특권 컨테이너와 `--pid=host`를 쓰지 않는다는 기존 제약 안에서 다시 판단한다.

떼어 낼 때(둘째 주제를 시작할 때):

1. 해석기에서 호출 트리를 기록 종류와 상관없는 입력(스레드별 구간과 스택)으로 받게 나누고, JFR 읽기는 그 입력을 만드는 쪽으로 둔다. 효과 분류는 주제별 함수로 뺀다.
2. 화면 템플릿의 공통 덩어리를 공통 파일로 옮기고, 주제마다 `TOPIC`만 쓴다. `txlab view`가 둘을 이어 붙인다.
3. 이 경계가 틀렸다고 볼 조건: 둘째 주제의 호출이 “부모가 같은 스레드의 바깥 호출”이라는 트리 모양에 맞지 않을 때다(예: 비동기 콜백, 인터럽트 처리, 다른 스레드로 넘어가는 일). 그러면 공통 계약부터 다시 정한다. 반대로 둘째 주제에서 `TOPIC` 바깥을 고쳐야 할 일이 거의 없으면 이 경계가 맞았던 것이다.

### 실험 5 화면 사용성 시험 사전 등록(10-15 21:45, 실행 전)

사용자 원문(21:30쯤): “제가 보기에 앞서 다른 모델, 다른 에이전트에게 해당 페이지를 주고 조작하고 학습하게 해보세요. 여기서 학습은 기계 학습이 아니라... 한 3년차 되는 백엔드 개발자가 이제 저 툴을 사용해서 db 내부 동작을 이해하려고 시도하게 해서, 이해하는지, 도움을 받는지, 실제로 도움이 되는지 등 그 반응을 관찰하고 물어봐서 피드백도 받으면서 개선, 수정 지점들을 확인하여 개선, 수정을 진행래 주세요. 물론 ai니까 저 지식과 정보들이 다 있긴 할 텐데... 모르는 사람이라는 페르소나를 부여하여 대략 select 시에 인덱스 통해서 위치 빠르게 찾고 파일 io 해서 그 중에 해당하는 데이터 찾아서 응답하는 등 어떻게 동작하는지는 대략 알지만 자세하게는 몰랐던 그런 사람 입장에서 사용해 보라고 하는 거죠.”

**대상과 사람.** 대상은 `concept-9-spring.html`(커밋 `a084feb53`)이다. 시험하는 쪽은 메인과 다른 모델의 하위 에이전트 둘이고, 한 번에 하나씩 돌린다(sonnet 먼저, 그 결과로 시험 절차 자체가 깨졌는지 본 뒤 haiku). 페르소나는 3년 차 백엔드 개발자다. `@Transactional`을 써 봤고, REQUIRES_NEW는 “새 트랜잭션”, HikariCP는 “커넥션 풀”로 안다. DB는 인덱스로 위치를 찾고 파일을 읽어 응답한다는 정도로 안다. MVCC, 락 대기, 연결이 스레드에 묶이는 일의 속은 본 적이 없다.

**오염 막기.** 에이전트에게는 주소, 페르소나, 학습 목표, “미리 기록한 실행이고 살아 있는 DB는 없음”만 준다. WORK·설계 메모·소스 경로는 주지 않는다. 저장소 파일 읽기·쓰기, git, 브라우저의 JavaScript 실행·네트워크 읽기는 금지한다. 내장 브라우저로 화면을 보고 누르는 것만 한다. 답마다 「화면에서 확인 / 원래 알던 것 / 추측」을 붙이게 한다. 화면에서 배웠다고 세는 것은 「화면에서 확인」이고 인용한 값이 화면에 실제로 있을 때뿐이다. 원래 알던 지식으로 맞힌 답은 화면이 가르쳤다는 증거로 세지 않는다. 기록마다 다른 값(연결 이름, 트랜잭션 번호, 시간, 줄 번호)만 깨끗한 신호다.

**두 단계.** A단계: 페이지를 열기 전에 궁금한 것 2~3개를 적고, 탭을 1440×900으로 맞춘 뒤 자유롭게 써 본다. 처음 60초의 인상과 첫 세 번의 조작을 따로 적고, 조작마다 “본 것 → 기대 → 이해 → 헷갈림”을 적는다(브라우저 조작 40번 안팎). B단계: A가 끝난 뒤 같은 에이전트에게 아래 질문을 보낸다. 질문을 A 전에 주지 않는 것은 탐색이 질문에 끌려가지 않게 하려는 것이다. 마지막으로 바라는 것, 다시 쓸지, 자신감(1~5)을 묻는다.

내용 질문과 기대 답(굵은 것이 정답의 핵심, 모두 화면에서 읽을 수 있는 값):

| 번호 | 질문 | 기대 답 |
| --- | --- | --- |
| C1 | 같은 행 장면: 안쪽(REQUIRES_NEW) UPDATE는 어느 연결로 갔고, 그때 바깥 트랜잭션의 연결은 어디에, 어떤 상태로 있었나 | **C2**로 갔다. 바깥 연결 **C1은 스레드에서 떼어 둔 채**(그림판의 떼어 둔 연결 C1) 서버에서는 **idle in transaction**, xid 753이다 |
| C2 | 같은 행 장면: 안쪽 UPDATE는 왜 진행되지 않았고 어떻게 끝났나. DB는 왜 교착으로 보지 않았나 | C2(754)가 **753이 고치는 행을 기다렸고**(받지 못한 락 C2 → 753), 약 3초 뒤 **lock timeout 오류(55P03)**로 끝났다(그 사이 교착 검사는 NoDeadlock). 서버에서 보면 **C1은 아무것도 기다리지 않는다**(ClientRead). 실제로 C1을 붙든 것은 C2의 답을 기다리는 같은 앱 스레드다(추론) |
| C3 | 풀 1칸 장면: 풀이 한 칸일 때 REQUIRES_NEW는 어떻게 되나. 서버로 무엇이 갔나 | 안쪽이 **풀에서 연결을 받지 못하고 약 2초 기다리다**(잠듦 약 2,005 ms) `SQLTransientConnectionException` → `CannotCreateTransactionException`으로 끝난다. 그동안 **서버로 간 메시지가 없다**. 그 뒤 C1에 ROLLBACK이 간다 |
| C4 | REQUIRED 장면: 안쪽이 실패했는데 왜 바깥은 `UnexpectedRollbackException`으로 끝나나 | 안쪽은 **같은 연결 C1·같은 트랜잭션에 참여**했다. 안쪽 실패 때 **롤백 전용 표시**(`doSetRollbackOnly`, 접혀 있음)를 하고 연결 위로는 아무것도 보내지 않는다. 바깥이 커밋할 때 **ROLLBACK을 보내고** 그 예외를 던진다 |
| C5 | 재빌림 장면: HikariCP는 연결을 빌려줄 때마다 살아 있는지 검사하나 | **아니다.** 쉰 지 1,224.7 ms인 첫 빌리기는 `isConnectionDead`와 `Q ""` 왕복이 있고, **5.7 ms 뒤의 둘째는 없다**. 갈린 자리는 HikariPool 168행이다 |

표지 질문(화면의 표기를 읽는가): M1 줄 앞 숫자(호출 깊이), M2 `[갈림]`(함께 견준 기록에서 이 메서드가 다른 길을 감), M3 점선 막대와 기울인 글씨(스택에만 보인 호출, 끝 시각 없음), M4 `＋n`(접힌 자식 수, 누르면 펼침), M5 사실·추론 표지(관찰·기록 값과, 호출에서 다시 셈하거나 소스로 읽은 해석의 구분).

**판정.** 내용 답은 정답(핵심을 모두 담고 화면 근거를 인용), 부분(핵심 일부), 오답(틀린 주장이 들어 있음)으로 나눈다. 화면이 도움이 됐다고 보는 조건은 페르소나마다 내용 5개 가운데 4개 이상이 「화면에서 확인」 정답이고, 표지 5개 가운데 4개 이상을 바르게 읽은 것이다. 못 미치면 화면이 그 사람에게 도움이 되지 못한 것으로 적는다.

**고칠 점의 등급.** A: 틀리게 이해했거나 막혔다. B: 추측해야 했거나 오래 걸렸다. C: 보기 불편함. A와 B 가운데 원인이 분명한 것을 고친다. 설계를 갈아엎지는 않는다. 고친 뒤 새 에이전트(같은 두 모델, 앞 시험을 모르는 새 인스턴스)로 같은 질문을 다시 묻고, 앞 판과 표로 견준다. 견준 표 없이 나아졌다고 말하지 않는다.

**이 시험이 말해 주지 못하는 것.** AI 페르소나는 3년 차 개발자가 아니다. 더 빨리 읽고, 빽빽한 화면을 더 잘 참고, 이미 아는 것이 많다. 그래서 이 시험의 통과는 사람에게 도움이 된다는 증명이 아니고, 실패와 헷갈린 자리가 더 강한 신호다. 설계자와 판정자가 같은 메인이라는 한계도 있다. 기대 답을 미리 적어 판정을 기계적으로 만들고, 페르소나의 원 기록을 스크래치패드(`ux/`)에 남겨 사용자가 다시 판정할 수 있게 한다. 내장 브라우저의 기본 폭(800px)에서 그림판 상자가 찌그러지는 것은 메인이 본 결함으로 따로 적고, 페르소나 시험은 1440×900에서 한다.

### 실험 5 화면 사용성 시험 1차 결과(10-15 22:10)

두 참가자 모두 사전 등록대로 A단계(자유 사용) 뒤 B단계 질문을 받았다. 원 보고는 [`evidence/usability-20261015/`](evidence/usability-20261015/README.md)의 `r1-sonnet-A.md`, `r1-sonnet-B.md`, `r1-haiku-A.md`, `r1-haiku-B.md`에 있다. 사전 등록은 원 기록을 스크래치패드에 남긴다고 적었으나, 스크래치패드는 세션이 끝나면 사라지므로 다시 판정할 수 있게 저장소로 옮겼다. 1차 sonnet의 두 보고는 처음에 메인이 요지만 옮겨 두었던 것을 하위 에이전트의 작업 기록에서 원문 전체로 바꿨다. 조작 수는 sonnet A 약 40·B 31, haiku A 51·B 28이다.

| 질문 | sonnet | haiku |
| --- | --- | --- |
| C1 연결과 바깥 연결의 상태 | 정답(C2, 떼어 둔 C1, idle in transaction, xid 753) | 정답(같은 값) |
| C2 막힌 이유·끝·교착 아님 | 정답(753 기다림, +4,030 55P03, NoDeadlock, ClientRead). 교착이 아닌 이유는 조각을 이어 추측했다고 적음 | 정답(같은 값). 교착이 아닌 이유는 화면 문장을 짜맞춘 추측, 55P03의 뜻은 못 찾음 |
| C3 풀 1칸 | 정답(잠듦 2,003 ms, SQLTransient → CannotCreate, 그동안 메시지 없음, 뒤에 ROLLBACK). 2초의 출처는 못 찾음 | 정답(같은 값). 관찰 카드가 뒤의 롤백(753 중단)을 먼저 보여 준다는 어긋남을 짚음 |
| C4 UnexpectedRollbackException | 정답. 다만 기본 보기에서는 이유가 없어 「전부」를 켜고 doSetRollbackOnly를 찾음 | 부분. 롤백 전용 표시를 세우는 소스 줄(902~906)과 commit의 예외 메시지는 찾았으나, 이 기록의 안쪽이 그 분기를 탔다는 고리는 추측 |
| C5 생존 검사 | 정답(첫 빌리기 1,225 ms에 isConnectionDead·Q "", 둘째 6 ms에 없음). 「전부」에서만 보였고 0.5초 출처는 못 찾음 | 부분. 둘째에 검사가 없는 것과 168행 조건은 찾았으나 첫째의 검사는 확인 못 함(갈림 목록에서 그 호출로 옮겨 갈 수 없었음) |
| 표지 M1~M5 | 4/5(M5 출처 표지의 뜻을 추측) | 4/5(M5: JFR이 무엇인지 못 찾음, 추론의 정의 없음) |
| 판정 | **통과**(내용 5/5, 표지 4/5) | **미달**(내용 3/5, 표지 4/5) |
| 자신감 | 3(순서·연결은 4, DB 지식에 기대는 이유는 2~3) | 3(화면 문장을 따라 이해했고 원래 지식으로 확인 못 함) |

haiku의 A단계 보고는 “바깥이 왜 롤백되는지 화면에서 알 수 없었다”고 적었다가 B단계에서 스스로 바로잡았다. 글 읽기를 4,000자로 잘라 상세 패널 끝을 놓쳤고, 찾기도 아직 그려지지 않은 상세 글을 찾지 못했다고 적었다. 도구 쓰임새의 한계지만, 답이 긴 상세 패널 맨 아래에 있었다는 화면 쪽 원인도 함께 있다. 장면을 바꾼 뒤 2초쯤 이전 화면이 남아 보였다는 말도 있다. 그리는 시간은 8~22 ms로 쟀으므로(아래 확인), 스크린샷 도구의 지연이라고 본다(추론).

**두 사람이 짚은 고칠 점(합친 것).** 출처 S는 sonnet, H는 haiku다.

| 번호 | 등급 | 출처 | 문제 | 고친 것 |
| --- | --- | --- | --- | --- |
| F1 | B | S·H | 출처 표지(관찰·JFR 기록·함수 기록·추론)의 뜻이 화면에 없음 | 범례 아래에 늘 보이는 표지·약어 풀이 |
| F2 | B | S·H | Q, P B D E S, C, E 55P03, Z(I/T/E), xid, idle in transaction, SAVEPOINT, 스레드에 묶인 연결이 안 풀림 | 같은 풀이(프로토콜·PostgreSQL·Spring·HikariCP 묶음) |
| F3 | B | S·H | 핵심 기전(doSetRollbackOnly, isConnectionDead)이 「전부」에서만 보임 | 해석기가 두 호출을 효과(rollbackOnly, validate)로 셈, 줄에서 보이지 않는 부모 이름을 흐리게 앞에 붙임 |
| F4 | B | S·H | 교착이 아닌 이유를 조각으로만 보여 추측으로 이음 | 해석기가 기다림의 고리를 셈(`waitCycles`), 그림판에 화살표마다 출처 표지를 단 상자 |
| F5 | B | S | 상대 칸에 지금의 기다림(+2,033)과 앞선 관찰(+533)이 한 시각 값처럼 섞임 | 함수 기록(지금)을 위에, 관찰은 찍은 시각과 「이 순간보다 n ms 전」을 붙여 아래에 |
| F6 | A 위험 | S·H | 찍는 도중 메시지가 오간 관찰(풀 1칸 a·b)이 경고 없이 처음 고른 줄로 나옴 | 그 관찰과 실행 버튼에 ⚠, 처음 고르는 줄에서 뺌 |
| F7 | B | S·H | 장면마다 무엇을 묻는지와 끝난 결과가 없음, Outer.requiresNew를 안쪽으로 읽음 | 장면 설명 줄(답하는 질문, 장면 코드의 요약, 결과: 트랜잭션의 끝·잔액·앱 예외) |
| F8 | B·C | S·H | 실행 a~d의 뜻, 풀 1칸만 c·d가 있는 까닭 | 설명 줄에 실행의 뜻과 결함 있는 실행의 경고 |
| F9 | B·C | S·H | 2초·0.5초의 출처 | 해석기가 앱 설정(LabConfig 44행)과 HikariCP 소스(HikariPool 65·168행)에서 값과 줄을 읽음, 빌리기·검사의 자세한 내용에 발췌 |
| F10 | C | S·H | 「14벌」「10번·12번」이 무엇을 센 수인지 | 「호출 n번 · 기록 k벌」, 견준 기록의 뜻 |
| F11 | C | S | 서버 사건 표의 원래 이름(xact_wait 등) | 그 사건의 한국어 글을 앞에 |
| F12 | C | S | 줄이 낮아 누르기 어려움 | 줄 높이 19 → 21 |
| F13 | C | S·H | 스택에만 보인 호출이 왜 생기는지 | 범례에 이유(아래에서 올라온 예외가 지나가며 끝남) |
| F14·F15 | B | S·H | 장면을 바꾸면 어디부터 볼지 모름 | 주요 순간 칩(누르면 그 줄), 처음 고르는 줄은 결함 없는 관찰 → 기다림·오류 순간 → 첫 갈림 |
| H1 | B | H | 답이 긴 상세 패널 맨 아래에 묻힘 | 상세 패널 맨 위에 「한 일」 한 줄(효과와 만든 예외의 첫 줄) |
| H2 | B | H | 갈림 목록의 길에서 이 기록의 그 호출로 옮겨 갈 수 없음 | 길마다 「이 기록에서 n번: +t …」 연결 |
| H3 | B | H | ◉와 ⏸의 차이, 「멈춘 채로 찍은 관찰(앱은 계속 돎)」이 서로 모순처럼 읽힘 | ◉를 「막힌 동안 찍은 관찰(도구가 앱을 세우지 않음)」로, ⏸와 대조해 풀이 |
| H4 | C | H | 음수 시각 | 범례에 「−는 장면 시작 전」 |
| M | C | 메인 | 800px에서 그림판 상자가 세로로 찌그러짐 | 좁은 폭에서 코드 창·자세한 내용을 아래로, 그림판·줄 그림은 옆으로 밀어 보기 |

고치지 않은 것: 장면 전환의 불러오는 표시(위 추론대로 그리는 쪽 지연이 아님), 1440 폭에서 상세 패널이 잘렸다는 haiku의 A단계 말(같은 폭의 스크린샷에서 재현하지 못함). 풀 1칸 장면의 앱 안 기다림(스레드가 떼어 둔 C1을 쥔 채 풀을 기다림)도 고리로 그리지 않았다. 두 사람 모두 C3을 맞혔다.

**해석기 변경과 검증.** `decode/spring_tree.go`에 효과 둘(rollbackOnly, validate), `waitCycles`(서버의 기다림마다 보낸 마디·번호를 받은 연결·떼어 둔 스레드를 잇고, 보낸 마디가 기다림 내내 열려 있을 때만 앱 쪽 화살표로 셈), `poolConfigOf`(LabConfig의 setConnectionTimeout, HikariPool의 aliveBypassWindowMs 기본값과 그것을 쓰는 조건의 줄)를 더했다. 시험 셋(`TestAnnotateTreeMarksRollbackOnlyAndValidate`, `TestWaitCyclesClosesOnlyWhenHolderIsHeldByWaitingThread`, `TestPoolConfigOfReadsLiteralTimeout`)을 더했고, 규칙을 하나씩 지운 변이 다섯(떼어 둔 연결 확인 제거, 시각 순서 무시, validate 분기 제거, 줄 번호 하나 어긋남, 기다림 끝 확인 제거)이 모두 실패했다. 열네 벌을 다시 풀면 같은 행 a·b에만 고리가 하나씩 나오고(C2 → 753 → C1, 떼어 둔 스레드 main, Cycle true), 보낸 마디는 두 벌 모두 QueryExecutorImpl.execute(요청 5, UPDATE)이며 기다림(+1,026~+4,030) 내내 열려 있다. 롤백 전용 표시는 REQUIRED a·b에서 한 번씩, 생존 검사는 장면마다 처음 빌릴 때(REQUIRES_NEW 계열은 두 번) 나온다. 재빌림의 둘째 빌리기에는 없다.

**화면 확인.** 내장 브라우저에서 열네 벌마다 기본 렌즈와 「전부」 두 가지로, 모든 줄과 모든 주요 순간을 하나씩 골라 다시 그렸다. 2,643번 그려 오류 0건, 한 번에 약 11 ms였다. 주요 순간의 연결이 모두 줄이나 마디로 풀리는 것과 장면마다 처음 고르는 줄이 있는 것도 확인했다. 장면 여섯을 1440 폭으로, 같은 행 장면을 800 폭으로 찍어 보았다.

**2차 시험의 한계(실행 전에 적음).** 고친 화면은 설명 줄과 주요 순간에서 C1~C5의 답 일부를 직접 보여 준다. 2차의 나아짐은 「기록을 따라가 이해했다」가 아니라 「요약 문장을 읽었다」일 수 있고, 사전 등록의 「화면에서 확인」 표지는 둘을 가르지 못한다. 그래서 2차에는 사전 등록에 없던 물음 하나를 B단계 끝에 더한다. 답마다 어느 요소(설명 줄·주요 순간 / 그림판 / 줄 그림 / 자세한 내용 / 범례·풀이)에서 읽었는지 묻는다. 이 물음은 1차 결과를 본 뒤 더한 것이다. 나머지 절차와 질문은 1차와 같고, 참가자는 앞 시험을 모르는 새 인스턴스(sonnet, haiku)다.

### 실험 5 화면 사용성 시험 2차 결과(10-15 22:30)

대상은 `aac957767`의 `concept-9-spring.html`이다. 시험 동안 이 파일은 바꾸지 않았다(두 참가자가 끝난 뒤 `git diff`로 확인). 참가자는 앞 시험을 모르는 새 sonnet과 새 haiku이고, A단계 지시와 B단계 질문은 1차와 같다. 다른 점은 위 「2차 시험의 한계」에 적은 「어디서 읽었나」 물음 하나뿐이다. 원 보고는 같은 폴더의 `r2-sonnet-A.md`, `r2-sonnet-B.md`, `r2-haiku-A.md`, `r2-haiku-B.md`에 있다. 조작 수는 sonnet A 41·B 29, haiku A 37·B 29다.

| 질문 | sonnet 1차 → 2차 | haiku 1차 → 2차 |
| --- | --- | --- |
| C1 | 정답 → 정답 | 정답 → 정답 |
| C2 | 정답(교착 아닌 이유는 추측으로 이음) → 정답(고리 상자에서 읽음, 그 해석은 화면이 추론으로 표시) | 정답(짜맞춘 추측, 55P03 뜻 모름) → 정답(고리 상자와 55P03 풀이에서 읽음) |
| C3 | 정답(2초 출처 못 찾음) → 정답(connectionTimeout 2,000 ms와 LabConfig 44행을 찾음) | 정답 → 정답(같은 설정을 찾음. A단계에서 “못 찾았다”고 적었다가 B단계에서 바로잡음) |
| C4 | 정답(「전부」를 켜야 찾음) → 정답(기본 보기의 +822 줄과 풀이에서 읽음) | **부분** → 정답(같은 줄과 풀이) |
| C5 | 정답(「전부」에서만 봄) → 정답(기본 보기의 두 빌리기 줄) | **부분** → 정답(첫째의 검사 줄과 둘째의 없음, 168행 조건) |
| 표지 M1~M5 | 4/5 → 5/5 | 4/5 → 5/5 |
| 판정 | 통과 → 통과 | **미달 → 통과** |
| 자신감 | 3 → 4 | 3 → 3 |

**어디서 읽었나.** 두 참가자 모두 C1~C5 답마다 둘 이상의 요소를 골랐다. 설명 줄·주요 순간(가)만으로 답한 질문은 없다. 가장 많이 고른 것은 줄 그림(다, 10번 가운데 10번)이고, 그림판(나)은 C1~C3에서, 풀이(마)는 C1·C2·C4에서, 자세한 내용(라)은 C2~C5에서 나왔다. 그래서 2차의 나아짐이 요약 문장만 읽은 결과라고 보기는 어렵다. 다만 이 물음은 1차 결과를 본 뒤 더한 것이고, 참가자의 자기 보고다.

**이 비교가 말해 주지 못하는 것.** 같은 모델이라도 인스턴스가 다르고, 한 판에 한 번씩뿐이라 우연의 폭을 모른다. 1차 haiku의 미달은 화면 쪽 원인(답이 긴 패널 끝, 갈림에서 그 호출로 못 감)과 도구 쪽 원인(글 읽기를 4,000자로 자름)이 섞여 있다. 2차 sonnet은 “글 읽기 도구로 본문을 읽은 부분이 많아 사람이 눈으로만 볼 때보다 작은 글씨와 아래쪽 범례를 더 잘 읽었을 수 있다”고 스스로 적었다. 판정은 1차와 같이 메인이 사전 등록의 기대 답과 맞춰 했다. C4의 핵심 가운데 「안쪽이 바깥 트랜잭션에 참여한다」를 2차 haiku는 [원래 알던 것]으로 적었고, 같은 답에서 그 뜻을 담은 풀이 문장(「참여한 안쪽이 실패하면 함께 쓰는 트랜잭션에 남기는 표시」)을 화면 근거로 인용했다. 인용을 화면 근거로 보아 정답으로 셌다. 사전 등록대로 원래 지식을 빼고 엄격히 세어 부분으로 낮춰도 내용 4/5라 통과 판정은 그대로다. 사람 참가자는 아직 없다.

**2차에서 나온 고칠 점과 처분.**

| 번호 | 등급 | 출처 | 문제 | 처분 |
| --- | --- | --- | --- | --- |
| G1 | B | S·H | 결과 줄의 「풀이 연 연결 C1, C2」를 「이 장면이 연결 둘을 썼다」로 읽음(둘 다 A단계에서 오독 뒤 바로잡음) | 「트랜잭션을 연 연결 C1」과 「풀이 장면 시작 전에 열어 두기만 한 연결 C2」로 나눔. 장면 시작 전인지는 중계가 본 접속 시각으로 정함 |
| G2 | B | S | 풀 1칸에서 “자기가 쥔 연결을 자기가 기다린다”를 추측으로 이음 | 해석기가 풀의 기다림도 고리로 셈(`poolWaitCycles`: 시간 초과로 끝난 빌리기의 순간 쓰이던 칸이 풀 크기 이상이고, 모두 그 스레드가 떼어 둔 연결일 때), 그림판에 풀 고리 상자 |
| G3 | B | S·H | 0.5초 창의 값을 둘 다 못 찾음(빌리기의 자세한 내용에만 있었음) | 빌리기 줄 자체에 「생존 검사 창 500 ms 안이라 검사 안 함 / 밖이라 검사」 |
| G4 | B | S | [갈림] 풀이의 「길이 둘 이상」이 길이(length)로 읽힘 | 「자식을 부른 차례(길)가 둘 이상」, 렌즈 이름도 「부른 차례가 다른 자리」 |
| G5 | B | H | 장면 시작 전의 관찰(xid -)을 장면 중 상태로 읽었다가 바로잡음 | 장면 시작 전에 찍은 관찰에 「장면 시작 전 관찰」 |
| G6 | C | S | 칩을 눌러도 반응이 없는 듯함. 실제로는 골라졌으나 900px 높이에서 줄 그림이 화면 밖 | 고른 줄이 화면 밖일 때만 그 줄을 화면 가운데로 |
| G7 | C | S | 경고의 「메시지 8개」가 무엇인지 | 관찰의 자세한 내용에 찍는 도중 오간 메시지 목록 |
| G8 | C | S | NESTED의 저장점 단계가 주요 순간에 없음, 「하위 754」의 뜻 | SAVEPOINT·ROLLBACK TO·RELEASE 요청을 주요 순간에, 풀이에 「하위 트랜잭션」 |
| G9 | C | S·H | 풀이가 페이지 맨 아래에 있어 늦게 찾음 | 설명 줄에 「표지·약어 풀이 ↓」 |

고치지 않은 것과 이유:

- 장면 질문에 대한 직답 한 줄(S)은 넣지 않았다. 이 화면의 목적은 기록을 따라가 답에 이르는 것이고, 손으로 쓴 답은 기록과 어긋나도 화면이 알아채지 못한다. 대신 G2·G3처럼 해석기가 기록에서 셈한 사실을 그 자리에 놓았다.
- 표지 툴팁이 스크린샷에 안 보인다는 말(S·H)은 도구의 한계로 본다. 마우스를 올린 상태가 스크린샷에 남지 않는다(추론). 같은 뜻은 풀이에 있다.
- 1440 폭에서 오른쪽이 잘렸다는 말(H 1차, S 2차)은 같은 폭의 내장 브라우저에서 문서 폭 1440, 상세 패널 오른쪽 끝 1424로 넘침이 없었다. 창보다 큰 크기를 줄여 보여 준 스크린샷 탓으로 본다(추론).
- 새 장면 요청은 이 화면이 아니라 실험의 일이다. REQUIRES_NEW에서 안쪽이 커밋된 뒤 바깥이 실패하는 장면(H, A·B 두 번), 전파 옵션 견줌 표, 풀 크기와 동시 요청을 바꾼 장면(S)이 그것이다.

**2차 뒤 변경의 확인.** 해석기 시험을 하나 더했다(`TestPoolWaitCyclesNeedsEveryInUseSlotHeldBySameThread`). 풀 크기 비교를 지운 변이와 떼어 둔 연결 확인을 지운 변이가 모두 실패했다. 열네 벌을 다시 풀면 풀 1칸 a~d에 풀 고리가 하나씩 나온다(C1, 떼어 둔 스레드 main, 기다림 2,003~2,006 ms). 같은 행 a·b의 서버 고리는 그대로다. 내장 브라우저에서 열네 벌 × 렌즈 두 가지의 모든 줄과 주요 순간을 골라 2,655번 그렸고 오류는 0건이었다. 주요 순간과 고리 상자의 연결은 모두 줄이나 마디로 풀렸다. **G1~G9는 참가자로 다시 시험하지 않았다.** 이 변경이 이해를 돕는지는 확인하지 않은 상태다.

### 실험 5 화면 사용성 시험 3차 사전 등록(10-16 00:05, 실행 전): 클로드가 아닌 모델, 맥락 없이

사용자 원문(10-15 23:30쯤): “그리고 gemini 3.8 flash, gpt 6 luna 같은 모델들도 맥락 없이 바로 이해하고 배울 수 있는지 맡겨보세요. 클로드 계열이 아닌 모델로.” 같은 메시지의 앞부분은 화면의 글이 너무 빽빽하니 그림이 스스로 설명하게 고치라는 요청이다. 그래서 이 3차는 글이 빽빽한 지금 화면의 기준 측정이고, 고친 뒤 같은 절차로 4차를 돌려 견준다.

**대상.** `concept-9-spring.html`(커밋 `180410b9f`, sha256 `6538da82…54a0e`). 시험 동안 이 파일과 템플릿을 바꾸지 않고, 끝난 뒤 `git diff`와 해시로 확인한다.

**참가자.** 이 컴퓨터의 CLI가 실제로 내놓은 모델 이름으로 정했다. Antigravity CLI(`agy models`)의 `gemini-3.8-flash-high`와 codex CLI(`codex debug models`)의 `gpt-6-luna`다. Gemini 3.8 Flash는 low·medium·high 세 단계 가운데 가장 높은 high를 쓴다. 빽빽한 화면을 꼼꼼히 읽는 일이라 가장 깊게 생각하는 단계가 맞다고 봤다. GPT-6-Luna는 기본값이 medium이지만 Gemini 쪽과 맞추려고 high를 쓴다. 다만 두 회사의 high는 같은 계산량을 뜻하지 않는다.

**맥락 없이의 해석(AI 해석).** 1·2차는 상황 설명에서 “Spring 트랜잭션 전파가 실제로 어떻게 도는지, 실행 기록으로 보는 화면”이라는 것과 “미리 기록한 실행이고 살아 있는 DB는 없음”을 알려 주고, 학습 목표도 줬다. 3차는 이 셋을 모두 빼고 “동료가 설명 없이 링크만 줬다. 무엇을 보여 주는 화면인지 스스로 알아내고, 배울 수 있는 만큼 배워 보라”만 준다. 페르소나의 배경지식 문단은 1·2차와 같게 둔다. 답마다 [원래 알던 것]을 가르는 기준이 이 문단이고, 1·2차와 견주려면 이 문단이 같아야 한다. 그래서 페르소나 문단에 있는 Spring·HikariCP라는 낱말이 화면의 주제를 짐작하게 하는 한계가 남는다.

**도구와 격리.** 두 참가자는 같은 브라우저 도구를 쓴다. 이 컴퓨터에 이미 받아져 있던 Chrome DevTools MCP 1.11.0(`~/.npm/_npx/15c61037b1978c83`)을 참가자마다 따로 띄운다. 이 서버는 실행마다 임시 프로필로 headless Chrome을 새로 띄우고(`--isolated --headless --viewport 1440x900`), 페이지 JavaScript 실행을 끄며(`--javascriptEvaluation=false`), 네트워크·성능·메모리 도구를 빼고, `http://127.0.0.1:8765/*`만 열게 한다. codex에는 실행할 때만 `-c mcp_servers.browser…`로 붙이고, 콘솔·CSS·lighthouse·파일 올리기·화면 녹화·emulate 도구는 꺼 둔다. 사용자 설정 파일(config.toml)은 읽지 않게 하고, 메모리·사용자 Chrome 조작(computer_use, browser_use)·chronicle 기능도 끈다. 셸은 읽기 전용 샌드박스다. agy에는 시험용 작업 폴더의 `.agents/plugins/persona-browser/`로만 붙여 전역 설정을 바꾸지 않는다. agy는 도구별로 끌 수 없어서 콘솔·CSS 도구는 지시로만 금지한다. 두 CLI 모두 전역 지침 파일(myai 진입점)은 그대로 읽는다. 이 파일에는 작업대 내용이 없다. codex의 메모리에는 작업대 이야기가 있어서 메모리 기능을 껐다. agy의 대화 저장소에는 앞선 연기 시험(페이지 제목 확인)의 기록이 남아 있다.

**연기 시험에서 본 것(10-15 23:39~23:58).** codex는 기본 설정에서 확장 프로그램으로 사용자의 실제 Chrome을 조작했다. 탭을 하나 열고, 전체 화면 단축키를 누르고, 창을 끌려고 했다. agy의 Gemini는 비대화 모드에서 내장 브라우저 도구가 꺼져 있자 페이지 HTML을 직접 가져왔다(`read_url_content`). 브라우저 하위 에이전트를 켠 시험에서는 연결을 스스로 고치려고 사용자 Chrome 프로필 폴더에 `DevToolsActivePort` 파일을 새로 썼다. 이 파일이 시험용 headless Chrome(9222)을 가리키는 것을 확인하고 지웠다. 위 격리 방식은 이 세 가지를 막으려고 정했다.

**절차.** 1·2차와 같다. A단계(자유 사용, 조작 40번 안팎) 뒤에 같은 대화를 이어서(codex `exec resume`, agy `--conversation`) B단계 질문을 보낸다. 다만 B단계에서는 브라우저가 새로 뜨므로 페이지를 다시 연다. 두 참가자는 브라우저와 CLI가 서로 다르므로 동시에 돌린다. 1·2차의 「한 번에 하나씩」은 같은 내장 브라우저를 나눠 쓴 탓이었다. A단계 보고에는 「다 써 본 뒤 이 화면이 무엇을 보여 준다고 이해했는지」 한 문단을 더한다. B단계의 내용 질문 C1~C5, 표지 질문 M1~M5, 마지막 세 가지는 2차와 같다. 「어디서 읽었나」의 고를 것에는 (바) 마우스를 올렸을 때 뜨는 글을 더한다. 4차의 고친 화면이 글을 접어 둘 것이므로, 접힌 글로 통과한 것을 그림으로 통과한 것과 가르려면 3차부터 같은 고를 것을 둬야 한다.

**판정.** 내용과 표지는 2차와 같다(내용 4/5 이상 「화면에서 확인」 정답, 표지 4/5 이상). 맥락 없는 이해는 따로 센다. A단계의 그 문단이 (1) 미리 기록한 실행을 다시 보는 화면이라는 것, (2) Spring 트랜잭션 전파 장면(REQUIRED·REQUIRES_NEW·NESTED)이라는 것, (3) 호출과 연결·풀과 DB 서버 쪽 상태를 한 시간축에 놓았다는 것 가운데 몇 개를 화면 근거로 맞혔는지다. 셋 가운데 둘 이상이면 맥락 없이 알아본 것으로 적는다. 실행이 끝나면 도구 기록(codex JSONL, agy stream-json)을 모두 훑는다. 금지 도구(셸로 파일 읽기, URL 직접 가져오기, JavaScript, DOM·소스·콘솔·네트워크, 다른 MCP 서버)를 쓴 실행은 결과를 적되 판정에서 뺀다.

**이 시험이 말해 주지 못하는 것.** 1·2차의 한계가 그대로 남는다. 여기에 셋이 더해진다. 도구가 1·2차의 내장 브라우저와 다르다(접근성 트리의 모양, 스크린샷 배율, 마우스를 올렸을 때 도구가 무엇을 보여 주는지). 1·2차와 달라진 점이 맥락 제거와 모델 두 가지라서, 결과의 차이를 어느 하나 탓으로 돌릴 수 없다. 한 모델당 한 번씩이라 우연의 폭을 모른다.

### 실험 5 화면 사용성 시험 3차 결과(10-16 00:40)

대상은 사전 등록대로 `180410b9f`의 `concept-9-spring.html`이다. 시험 동안 이 파일과 템플릿을 바꾸지 않았다. 네 실행이 끝난 뒤 `git diff`가 비어 있었고 sha256도 `6538da82…54a0e` 그대로였다. 원 보고는 같은 폴더의 `r3-gemini-A.md`, `r3-gemini-B.md`, `r3-gpt-A.md`, `r3-gpt-B.md`에, 두 모델에 준 지시문은 `r3-prompt-A.md`, `r3-prompt-B.md`에 있다.

**도구 기록 감사.** 네 실행의 도구 기록(codex JSONL, agy stream-json)을 모두 훑었다. 금지 도구를 쓴 실행은 없어서 넷 모두 판정에 넣는다. 아래 수는 기록에서 센 것이고, 보고서에 참가자가 적은 수와는 다르다.

| 실행 | 클릭·키 | 스크린샷 | 접근성 트리 읽기(`take_snapshot`) | 그 밖 |
| --- | --- | --- | --- | --- |
| Gemini A | 21 + 키 3 | 21 | 19 | `view_file` 42 |
| Gemini B | 4 | 2 | 4 | `view_file` 7 |
| GPT A | 19 | 12 | 16 | 없음 |
| GPT B | 5 | 6 | 6 | 없음 |

`view_file`은 모두 agy가 자기 도구의 출력(스크린샷 png, 길어서 파일로 뺀 접근성 트리 txt)을 저장한 대화 폴더 안의 파일이었다. 저장소나 다른 파일은 열지 않았다. 실행마다 `new_page` 한 번이 더 있고, B단계에는 `list_pages` 한 번이 더 있다.

| 질문 | Gemini 3.8 Flash(high) | GPT-6-Luna(high) |
| --- | --- | --- |
| C1 | 정답 | 정답 |
| C2 | 정답(C2 754가 753을 기다림, +4,033 ms에 55P03, NoDeadlock, C1은 ClientRead) | 정답(같은 값) |
| C3 | 정답(connectionTimeout 2,000 ms, `SQLTransientConnectionException` → `CannotCreateTransactionException`, 연결 위 메시지 없음). 「무한 대기」라고 쓴 뒤 같은 문단에서 2,000 ms 시간 초과로 바로잡음 | 정답(같은 시간 초과와 예외, 메시지 없음) |
| C4 | 정답(참여라 새 트랜잭션이 아님, `doSetRollbackOnly`는 연결 위로 보내지 않음, 바깥 commit에서 C1에 ROLLBACK, 그 뒤 예외) | **부분.** +822 ms의 `doSetRollbackOnly`와 「rollback-only라 롤백됐다」는 상세는 읽었다. 안쪽이 같은 C1·같은 트랜잭션에 참여했다는 것과 바깥 커밋이 ROLLBACK을 보낸다는 것은 답에 없다 |
| C5 | 정답(첫 빌리기는 쉰 지 1,225 ms > 500 ms라 `isConnectionDead`·`Q ""`, 둘째는 6 ms라 건너뜀) | **부분.** 둘째의 건너뜀(6 ms)과 500 ms 규칙은 읽었다. 첫째의 검사는 답에 없다. 1차 haiku의 C5와 같은 기준으로 셌다 |
| 표지 M1~M5 | 5/5 | 5/5. M3의 점선 막대는 범례에 따로 풀이가 없어 [추측]으로 이었고, 기울인 글씨가 스택에만 보인 호출이라는 것은 화면 근거로 답했다 |
| 맥락 없는 이해 | 3/3 | 3/3 |
| 판정 | 통과 | **미달**(내용 3/5) |
| 자신감 | 5 | 4 |

**맥락 없는 이해.** 두 모델 모두 A단계 끝 문단에서 셋을 다 맞혔다. Gemini는 「JVM 메서드 호출(JFR)과 소켓 패킷, DB 엔진 상태 관측을 시간 순으로 동기화한 트랜잭션 실행 관측 뷰어」라고 적었다. 따로 「대화형 시뮬레이터가 아니라, 미리 설계한 여섯 장면을 실제 Java·PostgreSQL로 돌려 캡처해 둔 트레이스 뷰어」라고도 적었다. 근거로는 머리의 Java 25.0.4, PostgreSQL 18.6, 실행 a·b를 들었다. GPT는 「Spring 전파 장면을 앱 호출, Spring 내부 호출, 커넥션 풀, PostgreSQL 상태와 메시지를 시간순으로 함께 보여 준다」를 [화면에서 확인]으로 적었다. (1) 「기록을 다시 보는 화면」은 [추측]으로 적었지만, 바로 앞에 「코드 패널은 기록 당시 컴파일한 사본」과 실행 a·b 전환을 화면 근거로 들었다. 그래서 맞힌 것으로 셌다. [화면에서 확인] 문장만 엄격히 세어도 GPT는 2/3이라 판정은 같다.

**어디서 읽었나.** Gemini는 C1 가·나·다, C2 나·다·라, C3 가·나·다, C4 가·다·라, C5 가·다·라를 골랐다. GPT는 C1~C3 가·나·다·라, C4 가·다·라, C5 가·다·라·마를 골랐다. 열 번 가운데 줄 그림(다)이 10번, 맨 위 질문 줄·주요 순간(가)이 9번, 자세한 내용(라)이 8번, 그림판(나)이 6번이다. 풀이(마)는 1번, 마우스를 올렸을 때 뜨는 글(바)은 0번이다. (가)만으로 답한 질문은 2차처럼 없다. 다만 (가)를 거의 모든 답에서 고른 것으로 보아, 지금 화면에서는 맨 위 설명 줄과 주요 순간의 글이 답을 받치는 몫이 크다.

**이 시험이 그림을 재지 못한 것.** 감사표에서 보듯 두 모델은 스크린샷만큼 접근성 트리를 읽었다. 접근성 트리는 페이지의 글을 모두 담는다. 맨 아래 범례의 긴 문장, 그림판 상자 안의 글자, SVG 안의 글자가 그대로 들어간다. 그래서 「(나) 그림판에서 읽었다」는 답도 그림을 보고 안 것인지 상자 안 글자를 트리로 읽은 것인지 가를 수 없다. 표지 답 다섯 개를 Gemini가 모두 범례 문장 그대로 인용한 것도 같은 꼴이다. 이 방식으로 재는 것은 「필요한 정보가 페이지의 글 어딘가에 있는가」에 가깝다. 사용자가 고치라고 한 「그림이 스스로 설명하는가」는 재지 못한다. 4차는 이 한계를 메워야 한다(아래 4차 사전 등록에서 정한다).

**3차에서 나온 고칠 점.** 참가자의 「막힌 곳」과 「바라는 것」에서 화면에 관한 것만 골랐다.

| 번호 | 출처 | 문제 | 처분 |
| --- | --- | --- | --- |
| H1 | Gemini A | 그림판의 HikariCP·PostgreSQL 상자가 누를 수 있어 보이는데 눌러도 반응이 없음 | v2 그림판에서 풀 칸·백엔드·기다림 화살표·고리 표지를 누르면 해당 줄로 감 |
| H2 | Gemini A | 시간이 길게 벌어진 구간(잠듦, 락 대기, 풀 고갈)이 줄 그림에서 눈에 띄지 않음 | v2 맨 위 시간 띠에 연결별 막대와 기다림 구간을 그림. 멈춤 자리는 빗금 |
| H3 | GPT A·B | 풀 대기와 DB 락 대기의 차이를 한눈에 보고 싶음 | v2 그림판에서 풀 대기는 main → 풀의 빨간 화살표와 ⏳ 2,000 ms, 락 대기는 xid 사이의 빨간 곡선과 🔒로 모양을 달리함 |
| H4 | GPT B | 점선 막대의 뜻이 범례에 따로 없음 | v2 기호 줄에 「┆ 스택에만 보인 호출(기울임)」을 둠 |
| H5 | GPT A | 「떼어 둠」을 연결이나 트랜잭션의 끝으로 읽을 뻔함 | v2 그림판에서 떼어 둔 연결은 점선과 ⏸로 남고 백엔드는 「idle in tx」 알약을 유지함 |
| H6 | Gemini A | ⏸ 멈춤 구간을 앱이 잠든 것으로 읽었다가 풀이를 보고 바로잡음 | v2에서 멈춤 자리를 빗금으로 그림. 빗금이 오해를 줄이는지는 4차에서 봄 |
| H7 | GPT A | 「전부」를 접힌 자식 펼치기로 헷갈림 | 처분 보류. v2에서 렌즈 이름을 기호(• ⑂ ◔)로 바꾸면서 「전부」만 글로 남음 |

고치지 않은 것: 두 장면을 좌우로 나란히 견주는 보기(Gemini), 격리 수준·SUPPORTS·MANDATORY 장면(Gemini)은 이 화면이 아니라 실험의 일이다. 시작 안내 한 줄(GPT)은 2차의 직답 한 줄과 같은 이유로 넣지 않는다. 글을 줄이는 v2의 방향과도 맞지 않는다.

### 실험 5 화면 v2: 글이 설명하던 자리를 그림으로(10-16 01:20)

사용자 원문(10-15 23:30쯤): “근데 지금 화면을 보면 글씨가 너무 빼곡하게 들어가 있습니다. 이 말는 시각적 요소들이 설명하게 하기보다는 말로 다 풀어서 설명하고 있는데, 이걸 개선할 방법은 없나요? 이런 도구에서 글은 보조적 위치여야 하고 필요할 때 확인하여 도움을 받을 수 있으면 좋지만, 기본적으로 시각적 요소 그 자체가 설명이 되는 게 가장 좋아서요. 가령 지금 콜 스택을 뎁스와 색깔 선으로 표현하고 있는데 그게 순차적으로 내부에서 호출되는 함수들임을 말하지 않아도 바로 이해가 되죠.”

**원칙.** 사용자가 든 콜 스택의 예처럼, 뜻을 모양·색·위치로 말한다. 화면에 늘 보이는 글은 장면이 답하는 질문 한 줄, 결과 알약, 기호 설명 한 줄로 줄인다. 나머지 설명은 마우스를 올리면 뜨는 글(`<title>`)이나 접힌 「? 표기와 약어 풀이」로 옮긴다. 1·2차 시험 뒤에 더한 설명 문장(장면 설명, 기다림의 고리 상자 ①②③, 빌리기 줄의 생존 검사 창 문장)이 화면을 빽빽하게 만든 주된 원인이었다. 그래서 그 문장들이 하던 일을 그림이 맡게 했다. 화면이 값을 셈하지 않고 해석기가 셈한 값을 고르고 그리기만 한다는 원칙은 그대로다.

| 자리 | 3차까지(글) | v2(그림) | 글은 어디로 |
| --- | --- | --- | --- |
| 맨 위 설명 줄 | 장면 코드 요약 문장, 결과 문장, 주요 순간 칩 열몇 개 | 질문 한 줄, 결과 알약(`C1 753 ↺ ROLLBACK`, 바뀐 잔액은 굵게, `✕ 예외`), 그 아래 시간 띠: 연결마다 빌린 구간·트랜잭션 막대·서버의 기다림(빗금), main의 잠듦, 실험 도구의 멈춤(회색 빗금), 주요 순간 기호(⇠ ⏸ ▶ ⚑ ✓ ✕ 🔒 ⤺) | 장면 코드는 질문에, 기호마다 마우스 글 |
| 그림판 위 | 층마다 칸과 글 상자(「스레드에 묶인 연결 C2」, 「떼어 둔 연결 C1」 등) | 다섯 층 상자. 상자마다 지금 가장 깊은 호출, 일하지 않는 층은 점선, 일하는 층끼리 화살표. 오른쪽에 큰 `+t ms` | 상자의 마우스 글 |
| 그림판 아래 | 백엔드마다 상태 문장, 기다림의 고리 상자 ①②③ 문장 | main 알약과 풀 칸(빌려준 칸은 실선). 묶인 연결은 실선, 떼어 둔 연결은 점선과 ⏸. 연결 관은 요청이 나가 있으면 굵은 화살표. 백엔드 상자와 상태 알약(`🔒 락 기다림`, `◉ ⌛ idle in tx`), xid 배지(진행 중 테두리, ✓ 커밋, ✕ 중단, ↳ 하위). 락 기다림은 xid 사이의 빨간 곡선과 `⏱3s`, 풀 기다림은 main → 풀의 빨간 화살표와 `⏳ ≤ 2,000 ms`, 고리는 ↻. 롤백 전용 표시 ⚑, 생존 검사는 쉰 시간 막대와 창 눈금 | 선·배지·고리의 마우스 글. 고리 상자의 ①②③은 ↻의 마우스 글 한 줄(고리의 차례와 「서버의 교착 검사는 점선 상자 안의 기다림만 본다」)로 |
| 줄 그림 | 줄 끝의 [효과]·[갈림]·[시간] 글 꼬리표, 효과 문장 | •, ⑂, ◔ 기호. 효과는 짧은 기호(`⇠C1 1,225 ms 쉼 ✓검사`, `⏸C1`, `⚑롤백 전용`) | 줄의 마우스 글, 자세한 내용 |
| 범례 | 맨 아래 문장 다섯 줄 | 기호 줄 하나 | 접힌 「? 표기와 약어 풀이」(펼친 상태는 다시 그려도 유지) |
| 렌즈 단추 | 「효과: 밖에 드러나는 일」 같은 긴 이름 | `• 효과`, `⑂ 갈림`, `◔ ≥1 ms`, `전부` | 단추의 마우스 글 |

그림판의 표지는 근거의 종류를 모양으로 가른다. 관찰한 상태는 ◉, 해석기가 추론한 묶임은 호박색 ≈이다. 관찰이 고른 순간보다 오래되었으면 흐리게, 찍는 도중 메시지가 오갔으면 ⚠를 붙인다. 그림판의 풀 칸, 백엔드, 기다림 곡선, ↻를 누르면 해당 줄로 간다. 3차 Gemini가 「상자가 눌릴 것 같은데 반응이 없다」고 한 자리(H1)다. 처음 고르는 줄은 주제가 정한다(`startRow`). 차례는 막힌 동안 찍은 깨끗한 관찰 띠, 첫 주요 순간, 둘째 연결의 첫 요청, 둘째 빌리기 순이다. 그래서 장면을 바꾸면 그 장면의 핵심 순간에서 그림판이 열린다.

글로 남긴 것: 오른쪽 자세한 내용(고른 줄의 사실 목록), 코드 창, 줄의 메서드 이름이다. 셋은 설명이 아니라 보여 줄 내용 그 자체라 남겼다.

**범용화 경계.** 시간 띠(`scrubber`), 렌즈의 기호 이름, 접힌 풀이는 공통 화면에 두었다. 띠에 무엇을 그릴지(`ganttRows`), 그림판(`board`), 결과 알약(`outcomeChips`), 효과 기호(`effGlyph`), 기호 줄(`rowLegend`, `boardLegend`), 첫 줄(`startRow`)은 TOPIC이 준다. 다른 주제는 이 함수들만 쓰면 같은 화면 틀을 쓴다.

**확인.** 저장소의 `txlab` 소스로 `180410b9f`의 템플릿을 다시 묶으면 커밋된 화면과 sha256이 같다(`6538da82…54a0e`). 그래서 v2도 Go 코드는 바꾸지 않고 템플릿만 바꿔 같은 방법으로 묶었다. 새 화면은 sha256 `b2d23096…9f0810`이다. 묶을 때 쓰는 해석 결과 열네 벌(`r5-*.json`)과 미리보기용 정적 서버의 소스는 `open_source_internals/database/postgresql/scripts/.lab-tmp/`의 `spring/decoded-r5/`와 `static-server/`에 둔다(git 제외). 원래 있던 세션 작업 폴더는 임시 위치라 사라질 수 있다. 내장 브라우저에서 열네 벌 × 렌즈 두 가지(기본, 전부)의 모든 줄과 주요 순간을 골라 2,693번 그렸고 오류는 0건이었다. 설명 줄·시간 띠·그림판이 가리키는 줄은 모두 실제 줄이나 마디로 풀렸다. 열네 벌 모두 `startRow`가 줄을 골랐다. 풀이는 처음에 접혀 있고 내용이 들어 있다. 밝은 테마와 어두운 테마에서 그림판을 찍어 보았다. 시간 띠의 눈금 글자가 1.5 s를 2 s로 반올림해 옆 눈금과 겹치는 버그를 이때 찾아 고쳤다(눈금 간격이 1초보다 잘면 소수 자리를 남김).

**확인하지 않은 것.** 이 그림이 사람에게 스스로 설명이 되는지는 아직 모른다. 4차 시험은 AI 참가자로 이것을 견준다. 사람의 판단은 사용자의 몫이다.

### 실험 5 화면 사용성 시험 4차 사전 등록(10-16 01:50, 실행 전): 그림만 보는 갈래를 더해

**3차 판정의 C5 기준을 바로잡는다.** 사전 등록의 기대 답 표에서 C5의 굵은 핵심은 「아니다」와 「5.7 ms 뒤의 둘째는 없다」다. 첫째의 `isConnectionDead`·`Q ""`는 굵지 않다. 3차 GPT는 굵은 두 가지를 다 답했는데, 1차 haiku의 판정 선례(첫째의 검사를 못 찾으면 부분)를 따라 부분으로 셌다. 그 선례는 사전 등록보다 엄격한, 적지 않은 기준이었다. 앞으로는 사전 등록대로 굵은 핵심으로 센다. 그러면 3차 GPT는 C5 정답, 내용 4/5로 **통과**다(C4는 굵은 셋 가운데 하나만 답해 그대로 부분). 같은 기준을 1차 haiku에 다시 대면 C5가 정답이 되어 내용 4/5로 통과가 된다. 이 두 판정은 기준 하나로 갈린다. 위 1차·3차 결과 절의 표는 그때 쓴 판정 그대로 둔다.

**왜 갈래를 나누나.** 3차 감사에서 두 모델은 스크린샷만큼 접근성 트리를 읽었다. 그래서 3차와 같은 도구로는 「그림이 스스로 설명하는가」를 재지 못한다. 시험용 서버(Chrome DevTools MCP 1.11.0)는 접근성 트리 읽기와 스크린샷을 같은 묶음(debugging)에 두어 묶음 단위로는 가를 수 없다. 대신 좌표로 누르는 `click_at`(`--experimentalVision`, 3차에도 켜져 있었음)이 있어, 스크린샷과 좌표 클릭, 키만으로 화면을 쓸 수 있다.

| 갈래 | 화면 | 도구 | 무엇을 견주나 |
| --- | --- | --- | --- |
| S | v2(`3b39039af`, sha256 `b2d23096…9f0810`, 8765) | 3차와 같음(접근성 트리 허용) | 3차와 견줌. 글을 접고 마우스 글로 옮긴 것이 글을 읽는 참가자에게 손해인지 |
| V | v2(같은 파일, 8765) | 그림만: `new_page`, `navigate_page`, `list_pages`, `select_page`, `close_page`, `take_screenshot`, `click_at`, `press_key` | 아래 V-old와 견줌. 그림이 스스로 설명하는지 |
| V-old | 3차 화면(`180410b9f`, sha256 `6538da82…54a0e`, 8766) | V와 같음 | V의 대조군. 이것이 없으면 V의 결과를 v2 덕으로 돌릴 수 없음 |

참가자는 갈래마다 새 인스턴스의 `gemini-3.8-flash-high`(agy)와 `gpt-6-luna` high(codex)다. 모두 여섯 쌍이고, 쌍마다 A단계 뒤에 같은 대화를 이어 B단계를 돈다. 3차 화면은 저장소 밖 작업 폴더에 복사해 다른 포트(8766)로 띄운다. sha256이 3차 대상과 같은 것을 확인했다.

**그림만 보기 갈래의 도구.** codex는 허용 목록(`enabled_tools`)으로 위 여덟 도구만 붙인다. agy는 도구를 하나씩 끌 수 없어 지시로 막고, 끝난 뒤 기록에서 확인한다. 지시는 이렇다. 접근성 트리 읽기, 요소 번호(uid)로 다루는 도구(`click`, `hover`, `fill`, `drag`, `type_text`, `wait_for`), `includeSnapshot` 인자를 쓰지 않는다. 스크린샷은 1440×900 픽셀이고 `click_at`의 좌표도 같은 픽셀이다. 마우스를 올리는 도구는 없고, 스크롤은 PageDown·PageUp·Home·End 키로 한다. agy는 스크린샷을 파일로 빼 두고 그 파일을 열어야 그림을 본다. 그래서 그 이미지 파일만 열어 보게 허용한다.

연기 시험(10-16 00:41~00:53)에서 본 것: GPT는 첫 스크린샷에서 읽은 좌표(700, 23)로 「REQUIRES_NEW · 풀 1칸」을 한 번에 눌렀다. Gemini는 처음에 파일 열기를 금지한 지시 때문에 그림을 못 보고 짐작으로 눌렀다. 이미지 파일 열기를 허용한 두 번째 시험에서는 세 번째 시도에 맞혔다. 스크린샷을 1920×1080으로 잘못 알아 처음 두 번은 옆 단추를 눌렀다. 그래서 지시에 실제 좌표계(1440×900)를 적는다. 이 Gemini는 그림판만 보고 「하나뿐인 C1이 idle in transaction 753에 묶여 있어 main이 최대 2,000 ms 기다린다」고 묘사했다. PageDown은 문서를 내린다. 문서 높이(약 2,300 px)는 오른쪽 자세한 내용의 길이로 정해진다. 줄 그림은 안쪽 스크롤 영역이라 줄을 누르거나 ↑↓ 키로 움직인다. 3차 화면도 같은 배치다.

**지시문.** S는 3차와 같다. B단계의 표지 질문 둘과 「어디서 읽었나」의 이름만 v2에 맞춘다. M2는 「줄 끝의 ⑂ 기호」, M5는 「'◉', '≈' 같은 작은 표지와, 자세한 내용의 'JFR 기록', '추론' 같은 표지」다. 고를 것은 (가) 맨 위 질문 줄·결과 칩·장면 미니맵(시간 막대), (나) 위쪽 구성도(스레드·풀·연결·DB 그림), (다) 시간 순 줄 그림, (라) 오른쪽 자세한 내용, (마) 아래쪽 기호 줄·접힌 풀이, (바) 마우스를 올렸을 때 뜨는 글이다. V는 S의 지시문에서 규칙 단락만 위 도구 규칙으로 바꾼다. 조작 예산은 「누르기·키 25번 안팎, 스크린샷은 필요한 만큼」으로 둔다. 3차 실제 클릭 수(19~24)에 맞춘 것이고, 그림만 보는 참가자는 누를 때마다 스크린샷이 필요하기 때문이다. V-old는 V와 같고, 표지 질문과 「어디서 읽었나」만 3차 화면의 이름(「[갈림]」, 「'관찰 +…'」, 「그림판(칸마다 상태 상자, 기다림의 고리 상자)」 등)을 쓴다.

**판정.** 내용·표지·맥락 없는 이해는 3차와 같은 기준이고, C5는 위에서 바로잡은 대로 굵은 핵심으로 센다. 갈래마다 따로 판정한다. 그 위에서 셋을 미리 정해 둔다.

1. 그림이 스스로 설명하는가: 두 모델 모두 V가 통과하면 「글을 줄여도 그림만 보는 참가자가 배웠다」로 적는다. 하나라도 미달이면 미달한 질문과 그 답이 막힌 자리를 적는다.
2. v2 덕인가: 모델마다 V와 V-old의 내용 점수를 견준다. V가 V-old보다 낮은 모델이 있으면, v2가 그림으로 옮기지 못하고 지운 정보가 있다고 보고 그 질문을 따로 판다. V가 같거나 높으면 「줄인 글자만큼 그림이 메웠다」로 적되, 한 판씩이라 우연의 폭은 모른다고 함께 적는다.
3. 글을 읽는 참가자에게 손해인가: S가 3차보다 내용에서 2점 이상 낮으면, 접힌 풀이나 마우스 글로 옮긴 것 가운데 답에 필요했던 것을 찾아 늘 보이는 자리로 되돌릴지 정한다.

**기계로 잰 글의 양(참가자와 무관).** 장면마다 실행 a를 열었을 때의 첫 화면(1440×900)에 그려진 글자 수를 공백을 빼고 셌다. 마우스 글과 접힌 풀이 안은 뺐다. 「위쪽」은 머리, 장면 설명, 시간 띠, 그림판, 기호 줄이다.

| 장면 | 3차 위쪽 | v2 위쪽 | 3차 첫 화면 전체 | v2 첫 화면 전체 |
| --- | --- | --- | --- | --- |
| REQUIRED · 안쪽이 실패 | 1,429 | 615 | 4,054 | 3,039 |
| REQUIRES_NEW | 1,182 | 616 | 3,828 | 2,602 |
| REQUIRES_NEW · 같은 행 | 1,994 | 663 | 3,757 | 2,896 |
| REQUIRES_NEW · 풀 1칸 | 1,534 | 538 | 3,334 | 2,545 |
| NESTED · 안쪽이 실패 | 1,364 | 587 | 3,559 | 2,655 |
| 0.5초 안에 다시 빌림 | 1,058 | 511 | 3,130 | 2,296 |
| 합 | 8,561 | 3,530(−59%) | 21,662 | 16,033(−26%) |

첫 화면 전체에서 덜 준 것은 코드 창(장면마다 같음)과 줄 그림의 메서드 이름이 그대로라서다. 처음 고르는 줄이 3차와 v2에서 달라 자세한 내용의 글자 수도 장면마다 다르다.

**감사.** 끝나면 열두 실행의 도구 기록을 모두 훑는다. V와 V-old에서 접근성 트리 읽기, uid 도구, `includeSnapshot: true`, 자기 스크린샷이 아닌 파일 열기가 하나라도 있으면 그 실행은 결과를 적되 판정에서 뺀다. 도구 이름뿐 아니라 도구 응답에 접근성 트리의 꼴(`uid=`)이 들어 있는지도 본다. 시험 동안 두 화면 파일과 템플릿을 바꾸지 않고, 끝난 뒤 해시로 확인한다.

**이 시험이 말해 주지 못하는 것.** 3차까지의 한계가 그대로 남는다. 여기에 넷이 더해진다. (1) 그림만 보는 갈래는 좌표를 맞히는 모델의 솜씨도 함께 잰다. 연기 시험의 Gemini처럼 빗나간 누르기가 탐색을 줄일 수 있다. (2) 그림만 보는 갈래는 마우스 글을 볼 수 없다. 그래서 V는 기호와 배치만으로 재고, 사람이 필요할 때 마우스를 올려 보는 쓰임은 재지 못한다. 반대로 S에서는 마우스 글이 접근성 트리에 늘 보이는 글로 들어간다. 어느 갈래도 「필요할 때 꺼내 보는 글」을 사람처럼 쓰지 않는다. (3) V-old도 스크린샷 속의 글은 읽을 수 있다. 그러니 V와 V-old의 차이는 「그림 대 글」이 아니라 「글이 빽빽한 화면 대 그림이 많은 화면」을 눈으로 본 차이다. (4) 갈래와 모델마다 한 번씩이다.

### 실험 5 화면 사용성 시험 4차 결과(10-16 01:45)

여섯 쌍이 00:55~01:35에 돌았다. 시험 동안 두 화면 파일과 템플릿을 바꾸지 않았다. 끝난 뒤 8765는 `b2d23096…`, 8766은 `6538da82…`를 그대로 내주었고 `git status`에 화면 파일의 변경이 없었다. 원 보고는 같은 폴더의 `r4-<모델>-<갈래>-<A|B>.md` 열두 개다.

**감사.** 열두 실행 모두 위반이 없어 판정에 넣는다. 그림만 보는 네 쌍은 접근성 트리 읽기, uid 도구, `includeSnapshot`, 스크린샷이 아닌 파일 열기가 한 번도 없었고, 도구 응답에도 접근성 트리의 꼴이 없었다. Gemini는 S 갈래에서도 접근성 트리를 A단계에 한 번 읽었을 뿐 거의 스크린샷과 좌표 클릭으로 썼다. GPT는 S 갈래에서 3차처럼 `click`과 `includeSnapshot`을 썼다(S에서는 허용).

| 갈래 | 모델 | 내용 | 표지 | 맥락 없는 이해 | 판정 | 자신감 |
| --- | --- | --- | --- | --- | --- | --- |
| S(v2, 3차 도구) | Gemini | 5/5 | 4/5(M2 오답) | 3/3 | 통과 | 5 |
| S | GPT | 5/5 | 5/5 | 3/3 | 통과 | 4 |
| V(v2, 그림만) | Gemini | 5/5 | 4/5(M2 부분) | 3/3 | 통과 | 5 |
| V | GPT | 5/5 | 4/5(M5 부분) | 3/3 | 통과 | 3 |
| V-old(3차 화면, 그림만) | Gemini | 5/5 | 5/5 | 2/3 | 통과 | 5 |
| V-old | GPT | 5/5 | 5/5 | 3/3 | 통과 | 4 |

맥락 없는 이해의 (1) 「미리 기록한 실행」은 여섯 가운데 다섯이 [추측]으로 적었다(실행 a·b·c·d, 「기록 때 컴파일한 사본」을 근거로). 3차처럼 맞힌 것으로 셌고, [화면에서 확인]만 세면 모두 2/3이다. Gemini V-old는 (1)을 적지 않았다.

**내용 답을 어디서 얻었나.** 점수는 여섯 모두 같지만, 「왜」를 묻는 두 자리에서 근거의 종류가 갈렸다.

- C2의 「DB는 왜 교착으로 보지 않았나」: V-old의 두 모델은 3차 화면의 기다림의 고리 문장(「서버의 교착 검사는 서버 안의 락 기다림만 보고 앱 안의 일이라 보지 못함」)을 스크린샷에서 그대로 읽었다. v2에서 이 문장은 ↻의 마우스 글로 옮겨져 그림만 보는 갈래는 볼 수 없다. GPT V는 「C1은 idle in transaction이라 대기 관계에 순환이 없다」를 [추측]으로 이었다. 그러면서 「화면이 그 이유를 문장으로 직접 설명하지는 않는다」고 적었다. Gemini V는 그림판의 `교착 검사 → NoDeadlock`과 관찰 표의 두 세션 상태(C1 ClientRead, C2 Lock/transactionid)를 견줘 같은 결론을 화면 근거로 적었다. 그림판의 빨간 고리가 main을 지나 DB 상자 밖에서 닫힌다는 것을 이유로 든 참가자는 없었다.
- C3의 「그동안 서버로 무엇이 갔나」: V-old는 「그동안 연결 위 메시지 없음」 문장을 읽었다. v2에서는 그 문장도 마우스 글로 갔다. Gemini V는 「그림판의 C1 연결선 위에 보내는 메시지가 없음」과 줄 그림의 빈 구간을 근거로 「0개」라고 답했다. GPT V는 「안쪽 UPDATE가 DB에 닿은 표시를 찾지 못했다」를 [추측]으로 이었다.
- C1의 「떼어 둔 연결」과 C5의 생존 검사는 v2 그림에서 바로 읽었다. Gemini V는 C1을 그림판의 점선 `---||`에서, C5를 쉰 시간 막대 `|--- 6 ms ---|`에서 근거로 들었다.

**표지에서 나온 문제.** 셋 다 v2에서 기호를 바꾸며 생긴 것이다.

- `⑂`(갈림)를 Gemini가 두 번(S, V) `↳`로 읽었다. `↳`는 그림판의 하위 트랜잭션 배지에도 쓰는 기호다. 뜻도 S에서는 「예외로 정상 흐름을 벗어난 분기」라고 틀렸고, V에서는 「조건에 따라 다른 경로」라고 해 「함께 견준 기록마다」가 빠졌다. 3차까지는 줄 끝의 `[갈림]` 글자와 범례 문장이 그 뜻을 바로 보여 줬다. v2에서는 그 문장이 접힌 풀이로 가고 기호 줄에는 「⑂ 갈림」만 남았다.
- `◉`(관찰)를 GPT V가 「자기 시간 ≥ 1 ms」로 읽었다. 그건 렌즈의 `◔`다. 작은 크기에서 두 기호가 비슷하다.
- 3차 화면을 본 V-old의 두 모델은 표지 다섯을 모두 범례 문장에서 맞혔다.

**미리 정한 세 판정.**

1. 그림이 스스로 설명하는가: 두 모델 모두 V가 통과했다. 「글을 줄여도 그림만 보는 참가자가 배웠다」로 적는다.
2. v2 덕인가: 모델마다 V와 V-old의 내용 점수가 같다(모두 5/5). 사전 등록대로 「줄인 글자(위쪽 −59%)만큼 그림이 메웠다」로 적는다. 다만 이번 내용 질문은 여섯 실행 모두 만점이라 천장에 닿았다. 이 질문들로는 두 화면의 차이를 가를 수 없다. 갈린 것은 「왜」를 화면 근거로 댔는지(위)와 표지(V 4/5, V-old 5/5)뿐이고, 한 판씩이다.
3. 글을 읽는 참가자에게 손해인가: S는 두 모델 모두 3차보다 낮지 않다(Gemini 5 → 5, GPT 3 → 5. 바로잡은 C5 기준으로는 4 → 5). 되돌릴 것은 없다.

**4차에서 나온 고칠 점(아직 고치지 않음).**

| 번호 | 출처 | 문제 | 고칠 방향 |
| --- | --- | --- | --- |
| J1 | Gemini S·V | `⑂`가 `↳`로 읽히고, 「기록마다 다름」이라는 뜻이 기호 줄에서 빠짐 | 하위 트랜잭션 `↳`와 겹치지 않는 모양으로 바꾸고, 기호 줄의 설명을 「갈림」에서 「기록마다 다른 자리」로 |
| J2 | GPT V | `◉`(관찰)와 `◔`(≥1 ms)가 헷갈림 | 렌즈 기호를 관찰 표지와 모양이 다른 것으로 |
| J3 | GPT V·Gemini V | 「교착 검사가 왜 못 봤나」, 「서버로 간 것 없음」이 그림만으로는 [추측]이거나 간접 근거 | 그림판에서 PostgreSQL 점선 상자를 「교착 검사가 보는 범위」로 표시하고, 고리 가운데 상자 밖 구간의 모양을 달리함. 풀 기다림 동안 연결 관에 「메시지 0」을 그림으로(빈 관 표시) |
| J4 | GPT V | 그림판 위쪽 층 상자(JdbcTemplate, HikariCP)가 누를 수 있어 보이나 반응 없음(3차 H1과 같은 꼴) | 층 상자를 누르면 그 층의 가장 깊은 줄로 가게 하거나, 누를 수 없는 모양으로 |
| J5 | Gemini S | 떼어 둔 연결의 `┅⏸`를 처음에 「끊김·오류」로 읽음(기호 줄을 보고 바로잡음) | 4차에서는 한 번뿐이라 지켜봄 |

시작 안내 한 줄은 GPT가 세 갈래 모두에서 다시 바랐다. 2차·3차와 같은 이유로 글 안내는 넣지 않았다. 대신 사용자의 판단을 받을 항목으로 남긴다.

**이 비교가 말해 주지 못하는 것.** 사전 등록의 네 한계에 하나가 더해진다. 내용 질문이 여섯 실행 모두 만점이라 v2와 3차 화면의 차이를 재기에는 질문이 쉬웠다. 다음 판에서 차이를 재려면 답이 그림에만 있는 질문(예: 어느 순간에 어느 연결이 떼어져 있었나, 기다림이 어디서 닫히나)을 미리 정해야 한다. 사람 참가자는 여전히 없다.
