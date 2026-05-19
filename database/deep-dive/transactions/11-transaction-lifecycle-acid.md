# Transaction lifecycle and ACID

## transaction boundary, autocommit, savepoint, ACID

트랜잭션 경계의 등장 배경은 여러 사용자가 공유 데이터를 동시에 바꾸는 환경에서 반쪽 변경을 막아야 했다는 데 있다. 트랜잭션 경계는 여러 SQL 문장을 하나의 논리적 변화로 묶는 선이다. 자동 커밋이 켜져 있으면 문장 하나가 끝날 때마다 그 문장이 곧 작은 트랜잭션처럼 확정되고, 명시적 `BEGIN` 또는 `START TRANSACTION`을 쓰면 사용자가 `COMMIT`이나 `ROLLBACK`을 보낼 때까지 여러 문장이 같은 경계 안에 머문다. 이 차이는 문법 문제가 아니라 장애 반경을 정하는 문제다. 계좌 이체에서 출금만 커밋되고 입금과 원장 기록이 빠지면 SQL은 일부 성공했지만 업무는 실패한다.

세션은 서버와 맺은 연결이고 트랜잭션은 그 연결 안에서 시작되고 끝나는 작업 단위다. 커넥션 풀을 쓰는 애플리케이션에서는 같은 물리 세션이 여러 요청에 재사용되므로, 예외 경로에서 트랜잭션을 닫지 않으면 다음 요청이 이전 요청의 열린 경계, 임시 설정, lock wait, snapshot 영향을 물려받을 수 있다. 그래서 운영 리뷰에서는 `쿼리가 맞는가`만 보지 않고 `경계가 모든 exit path에서 닫히는가`를 본다.

이 절에서 하중을 지탱하는 공식 자료는 다음과 같다. 링크 자체가 결론은 아니며, 본문에서는 각 공식 문서가 제공하는 규칙을 계좌 이체, row version, read view, vacuum/purge trace와 연결해 다시 검증 가능한 형태로 풀어 쓴다.

- PostgreSQL transaction tutorial: https://www.postgresql.org/docs/current/tutorial-transactions.html
- MySQL COMMIT/ROLLBACK: https://dev.mysql.com/doc/refman/8.4/en/commit.html
- MySQL SAVEPOINT: https://dev.mysql.com/doc/refman/8.4/en/savepoint.html

```sql
CREATE TABLE accounts (
  account_id bigint PRIMARY KEY,
  balance numeric(12,2) NOT NULL CHECK (balance >= 0)
);
CREATE TABLE ledger (
  ledger_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  from_account bigint NOT NULL,
  to_account bigint NOT NULL,
  amount numeric(12,2) NOT NULL CHECK (amount > 0)
);
INSERT INTO accounts VALUES (100, 1000.00), (200, 300.00);

-- autocommit=true라면 아래 첫 UPDATE가 끝나는 순간 출금만 확정될 수 있다.
UPDATE accounts SET balance = balance - 500 WHERE account_id = 100;
-- 장애 발생: 입금과 원장 기록은 실행되지 않는다.
UPDATE accounts SET balance = balance + 500 WHERE account_id = 200;
INSERT INTO ledger(from_account, to_account, amount) VALUES (100, 200, 500);

-- 명시적 경계는 세 변화의 운명을 하나로 묶는다.
BEGIN;
UPDATE accounts SET balance = balance - 500 WHERE account_id = 100;
UPDATE accounts SET balance = balance + 500 WHERE account_id = 200;
INSERT INTO ledger(from_account, to_account, amount) VALUES (100, 200, 500);
COMMIT;
```

| 시점 | 경계 | 100번 잔액 | 200번 잔액 | ledger | 의미 |
|---|---|---:|---:|---|---|
| t0 | 시작 전 | 1000 | 300 | 없음 | 총액 1300 |
| t1 | autocommit 첫 문장 후 | 500 | 300 | 없음 | 돈이 사라진 중간 상태가 확정됨 |
| t2 | 명시적 트랜잭션 내부 | 세션 내부 500 | 세션 내부 800 | 세션 내부 있음 | 아직 경계 밖에는 확정 아님 |
| t3 | COMMIT | 500 | 800 | 있음 | 총액 1300, 이체 완료 |
| t4 | ROLLBACK | 1000 | 300 | 없음 | 이체 전 상태 복귀 |

### 1. 자동 커밋은 트랜잭션이 없다는 뜻이 아니다

자동 커밋은 DB가 문장마다 트랜잭션 경계를 자동으로 닫는 모드다. 한 줄 `UPDATE`도 내부적으로는 원자적으로 실행되고 commit 또는 rollback된다. 다만 여러 문장을 하나의 업무 단위로 묶지 못한다. 그래서 자동 커밋은 간단한 관리 쿼리에는 편하지만, debit-credit-journal처럼 함께 성공해야 하는 변화에는 위험하다.

운영에서 이 차이는 수동 SQL과 batch migration에서 크게 드러난다. 자동 커밋 상태를 확인하지 않고 큰 update를 실행하면 성공 직후 되돌림은 별도 보정 작업이 된다. 반대로 너무 큰 명시적 트랜잭션은 오래 lock을 잡고 MVCC 정리를 지연시킬 수 있다. 따라서 경계 크기는 업무 원자성과 운영 압력을 함께 보고 정한다.

### 2. BEGIN은 상태 변화를 시작하는 세션 명령이다

`BEGIN`이나 `START TRANSACTION`은 테이블에 데이터를 쓰는 명령이 아니라 세션의 transaction state를 바꾸는 명령이다. 이 상태 안에서 실행된 SQL들은 같은 commit/rollback 운명을 공유한다. 프레임워크가 이 명령을 대신 보내더라도 실제 경계는 JDBC connection이나 DB 세션에 묶인다.

이 때문에 같은 클래스 내부 호출로 Spring proxy를 우회하거나, 예외가 트랜잭션 관리자까지 전파되지 않으면 기대한 경계가 생기지 않는다. annotation은 의도를 표시하지만, 실제 경계는 호출 경로, connection binding, 예외 전파, finally cleanup이 함께 닫을 때만 만들어진다.

### 3. COMMIT은 애플리케이션 성공 응답보다 먼저 닫혀야 한다

커밋은 변경을 영속 상태로 확정하는 경계다. 애플리케이션이 DB commit 전에 사용자에게 성공 응답을 보내면, 사용자는 성공을 보았지만 DB는 rollback될 수 있다. 반대로 commit이 성공했는데 응답 전 네트워크가 끊기면 사용자는 실패처럼 느끼지만 DB에는 성공이 남을 수 있다.

따라서 결제와 정산 코드는 성공 응답, DB commit, 외부 API side effect, message publish 순서를 별도로 설계한다. DB 트랜잭션은 DB 내부 변경을 원자화할 뿐 외부 시스템까지 자동으로 원자화하지 않는다. 이 경계 혼동은 outbox, idempotency, reconciliation 설계로 이어진다.

### 4. ROLLBACK은 경계 안 변화를 취소하지만 외부 부작용은 되감지 못한다

롤백은 현재 트랜잭션 안에서 DB가 관리하는 변경을 취소한다. 하지만 이미 외부 PG사로 승인 요청을 보냈거나 파일을 썼거나 메시지를 발행했다면 DB rollback이 그 외부 부작용을 자동으로 취소하지 않는다. 그래서 외부 호출을 열린 DB 트랜잭션 안에 오래 넣는 설계는 두 방향의 불일치를 만든다.

실무에서는 DB 상태를 먼저 `PENDING`으로 commit하고 outbox를 통해 외부 호출을 처리하거나, 외부 승인 뒤 내부 원장을 idempotent하게 반영하고 reconciliation으로 보정하는 식의 설계를 비교한다. 어떤 방식이든 rollback의 영향 범위를 과장하지 않는 것이 중요하다.

### 5. SAVEPOINT는 부분 커밋이 아니라 부분 후퇴 지점이다

SAVEPOINT는 트랜잭션 전체를 닫지 않고 내부에 이름 있는 되돌림 지점을 만든다. `ROLLBACK TO SAVEPOINT`는 그 지점 이후 변경만 취소하고 바깥 트랜잭션은 계속 살린다. 하지만 바깥 트랜잭션이 rollback되면 savepoint 이전 변경도 모두 사라진다.

업무적으로 실패해도 되는 보조 감사 적재나 debug log insert는 savepoint로 감쌀 수 있다. 그러나 출금은 유지하고 입금만 되돌리는 식으로 핵심 원자성을 깨는 데 쓰면 안 된다. savepoint 이름은 `sp1`보다 `optional_audit`처럼 업무 의미를 드러내야 장애 분석이 가능하다.

### 6. Atomicity는 SQL 성공 개수가 아니라 업무 변화 묶음이다

원자성은 여러 변화가 하나의 성공 또는 하나의 실패로 관측되어야 한다는 뜻이다. 계좌 이체에서는 출금, 입금, 원장 기록이 한 묶음이다. 첫 UPDATE가 성공했는지보다 세 변화가 commit 시점에 모두 남았는지가 중요하다.

테스트도 같은 방식이어야 한다. 두 번째 UPDATE에서 의도적으로 제약조건 위반을 만들고 rollback 뒤 잔액 합계와 ledger count가 원상인지 확인한다. 정상 케이스만 보는 테스트는 원자성 실패를 잡지 못한다.

### 7. Isolation은 경계 안 중간 상태를 누가 보느냐의 문제다

트랜잭션 안에서 세션 자신은 중간 상태를 볼 수 있다. 다른 세션이 그 중간 상태를 볼 수 있는지는 격리 수준과 엔진 구현이 정한다. READ COMMITTED에서는 다른 세션의 commit 전 변경을 보지 않지만 문장마다 새 커밋 결과를 볼 수 있고, REPEATABLE READ는 더 안정된 snapshot을 유지할 수 있다.

따라서 `트랜잭션으로 묶었다`는 말만으로 동시성 의미가 끝나지 않는다. 경계, 격리 수준, MVCC/lock 구현을 함께 말해야 한다. 이 구분은 DU27 이후 MVCC와 isolation anomaly 절로 이어진다.

### 8. Durability는 commit 이후 crash recovery와 연결된다

지속성은 commit 응답 이후 장애가 나도 결과가 복구되어야 한다는 뜻이다. 이 속성은 WAL, redo log, fsync, checkpoint 같은 더 아래층 구조와 연결된다. 트랜잭션 lifecycle에서는 commit이 단순한 문자열 응답이 아니라 복구 가능한 경계라는 점만 먼저 잡는다.

운영에서 commit latency가 튀거나 disk flush가 밀리면 애플리케이션은 DB가 느리다고만 느끼지만 실제로는 durability 비용이 드러난 것일 수 있다. 그래서 성능 분석에서는 SQL 실행 시간과 commit 대기 시간을 분리해서 보는 것이 좋다.

### 9. 트랜잭션 경계는 커넥션 풀 위생과 연결된다

요청이 끝났는데 열린 트랜잭션이 남아 있으면 다음 요청이 같은 connection을 빌릴 때 오염된 상태를 받을 수 있다. 격리 수준 변경, read only flag, lock wait timeout, temporary table, session variable도 함께 문제가 된다.

좋은 pool 설정은 반환 시 rollback과 상태 reset을 보장하고, 애플리케이션 로그는 request id와 transaction begin/end를 남긴다. DB의 `pg_stat_activity`나 `information_schema.innodb_trx`에서 오래된 transaction이 보이면 코드의 finally path를 의심해야 한다.

### 10. 큰 batch는 원자성과 정리 비용 사이에서 경계를 나눠야 한다

백만 건을 한 트랜잭션으로 바꾸면 실패 시 되돌리기는 명확하지만 lock, undo/WAL, vacuum/purge 지연, replication lag가 커진다. 반대로 한 건마다 commit하면 중간 실패 후 재개는 쉬울 수 있지만 업무적으로 전체 all-or-nothing이 필요하면 맞지 않는다.

따라서 batch는 chunk id, 재시도 가능성, progress table, 보정 SQL, 관측 지표를 함께 설계한다. 트랜잭션 경계는 개발 편의가 아니라 장애 복구 전략의 일부다.

```sql
SELECT pid, state, xact_start, now() - xact_start AS xact_age, query
FROM pg_stat_activity
WHERE xact_start IS NOT NULL
ORDER BY xact_start;

SELECT trx_id, trx_started, trx_state, trx_query
FROM information_schema.innodb_trx
ORDER BY trx_started;

SHOW ENGINE INNODB STATUS;
```

재현 검증은 실패 지점을 바꿔 가며 한다. 첫 update 뒤 예외를 던진다. ledger insert에서 unique violation을 만든다. savepoint 안의 보조 로그 insert만 실패시킨다. 각 경우에 잔액 합계, ledger count, 열린 transaction 관측, 애플리케이션 성공 응답 순서를 확인한다. PASS는 업무적으로 함께 성공해야 하는 변화가 함께 남거나 함께 사라지는 것이고, FAIL은 SQL 일부 성공을 업무 성공으로 응답하거나 요청 종료 뒤 열린 transaction이 남는 것이다.
### 장애 지점으로 다시 읽는 트랜잭션 생애주기

트랜잭션을 제대로 이해하려면 `BEGIN`부터 `COMMIT`까지를 문법 순서로 외우는 데서 멈추면 안 된다. 강한 학습 방법은 업무 변화가 아직 메모리 안 의도인지, 현재 세션 안에서만 보이는 중간 상태인지, 다른 세션에도 보이는 커밋 상태인지, crash 뒤에도 복구되어야 하는 durable state인지 시간축으로 나누는 것이다. 계좌 이체에서는 출금, 입금, 원장 기록이 같은 운명을 가져야 한다. 이 세 변화가 같은 SQL 파일에 붙어 있다는 사실은 아무 보장도 하지 않는다. 보장은 같은 트랜잭션 경계 안에 들어가고, 실패 경로가 rollback으로 닫히고, 성공 응답이 commit 이후에만 나가며, commit 후 알 수 없는 상태를 idempotency와 reconciliation으로 다룰 때 생긴다.

공식 문서가 말하는 autocommit의 의미도 이 시간축에서 읽어야 한다. PostgreSQL은 명시적 `START TRANSACTION`이나 `BEGIN` 뒤에 오지 않는 명령이 끝나면 사실상 각 명령 뒤에 commit이 있는 것처럼 볼 수 있다고 설명한다. MySQL도 기본 autocommit 모드에서는 각 statement가 실패하지 않으면 atomic하게 commit된 것처럼 동작한다고 설명한다. 이 두 문장은 `트랜잭션이 없다`가 아니라 `문장 하나짜리 경계가 자동으로 닫힌다`로 읽어야 한다. 그래서 단일 `UPDATE account SET last_login_at = now()` 같은 작업은 autocommit이 자연스럽지만, 이체처럼 여러 row와 ledger가 함께 움직이는 작업은 명시적 경계가 필요하다.

```text
업무 요청 R-9001: 100번 계좌에서 200번 계좌로 500 이체

시간        DB 세션 내부                         외부에서 보이는 상태                 장애 판단
----------  -----------------------------------  ----------------------------------  -----------------------------
t0          아직 BEGIN 전                         100=1000, 200=300, ledger 없음       재시도해도 부작용 없음
t1          BEGIN                                 여전히 t0 상태                       경계는 열렸지만 외부 변화 없음
t2          100번 -500 실행                       다른 세션은 보통 기존 값만 봄          여기서 예외가 나면 rollback 필요
t3          200번 +500 실행                       다른 세션은 보통 기존 값만 봄          출금/입금 균형은 세션 안에서만 성립
t4          ledger insert                         다른 세션은 보통 기존 값만 봄          원장까지 들어가야 업무 의미 완성
t5          COMMIT 전                             아직 성공 응답 금지                   crash 시 rollback 또는 recovery 대상
t6          COMMIT 성공                           100=500, 200=800, ledger 있음        이제 사용자 성공 응답 가능
t7          응답 전 네트워크 단절                 DB에는 성공이 남음                    request_id로 기존 결과 replay
```

이 표에서 가장 위험한 지점은 t2와 t6이다. t2에서 출금만 성공한 뒤 애플리케이션이 예외를 삼키고 다음 요청으로 넘어가면 돈이 사라진다. t6에서 commit은 성공했지만 응답이 실패하면 사용자는 실패로 생각하고 다시 요청할 수 있다. 첫 번째 문제는 rollback과 경계 정리로 막고, 두 번째 문제는 request id unique key와 결과 replay로 막는다. 트랜잭션을 배우는 목적은 `COMMIT은 성공, ROLLBACK은 실패`라는 단어를 외우는 것이 아니라, 이런 애매한 지점에서 어느 상태를 진실로 삼을지 결정하는 것이다.

### savepoint를 쓰기 전에 물어야 하는 질문

Savepoint는 부분 커밋이 아니다. 같은 트랜잭션 안에서 뒤로 물러설 수 있는 내부 표식이다. 이 차이를 놓치면 `출금은 살리고 입금만 되돌린다`처럼 핵심 원자성을 깨는 설계를 만들 수 있다. MySQL 공식 문서도 `ROLLBACK TO SAVEPOINT`가 transaction 전체를 끝내지 않고 savepoint 이후의 변경을 되돌린다고 설명하며, commit이나 이름 없는 rollback이 모든 savepoint를 없앤다고 설명한다. 즉 savepoint는 최종 성공 경계를 만들지 않는다. 최종 성공 경계는 여전히 바깥 `COMMIT` 하나다.

Savepoint가 어울리는 경우는 실패해도 업무 성공을 깨지 않는 보조 작업이다. 예를 들어 이체 본체는 반드시 성공해야 하지만, 선택적인 debug payload 적재나 비핵심 audit enrichment가 실패할 수 있다. 이때 savepoint 이름을 `sp1`로 두면 장애 로그에서 의미가 사라진다. `optional_audit_payload`처럼 왜 되돌릴 수 있는지 드러나는 이름을 쓰면, 나중에 운영자가 `이 savepoint 이후의 실패는 이체 본체를 깨지 않는다`고 판단할 수 있다.

```sql
BEGIN;

UPDATE accounts
SET balance = balance - 500
WHERE account_id = 100 AND balance >= 500;

UPDATE accounts
SET balance = balance + 500
WHERE account_id = 200;

INSERT INTO ledger(from_account, to_account, amount)
VALUES (100, 200, 500);

SAVEPOINT optional_audit_payload;
INSERT INTO transfer_audit_payload(request_id, raw_json)
VALUES ('R-9001', '{...large debug payload...}');
-- audit payload가 너무 커서 실패하면 여기까지만 되돌린다.
ROLLBACK TO SAVEPOINT optional_audit_payload;

COMMIT;
```

이 예시는 savepoint를 권장하기 위한 예시가 아니라 경계의 의미를 분리하기 위한 예시다. audit payload가 규제상 필수 증빙이라면 savepoint로 삼키면 안 된다. 그때는 audit까지 본체 원자성에 포함하거나, 본체를 실패시키거나, 별도 보정 가능한 outbox로 나눠야 한다. 핵심 질문은 `이 하위 작업이 실패해도 업무 성공이라고 말할 수 있는가`다. 이 질문에 답하지 않고 savepoint를 쓰면, savepoint는 유연성이 아니라 데이터 손상을 숨기는 장치가 된다.

### autocommit과 운영 SQL의 가장 흔한 사고

운영자가 수동으로 데이터를 보정할 때 autocommit은 특히 무섭다. 개발 코드에서는 프레임워크 트랜잭션 관리자가 경계를 만들어 주지만, 콘솔에서는 현재 세션 설정이 곧 안전선이다. autocommit이 켜진 상태에서 `UPDATE transfers SET status='CANCELLED' WHERE created_at < ...`를 실행하면 실행이 끝난 행은 바로 확정된다. 뒤늦게 조건이 넓었다는 사실을 발견해도 같은 transaction의 rollback으로 되돌릴 수 없다. 그때의 복구는 새 트랜잭션에서 반대 보정 SQL을 쓰는 일이 되고, 원래 값의 스냅샷을 남겨 두지 않았다면 복구 정확도는 떨어진다.

그래서 운영 SQL은 실행 전후가 아니라 실행 중간까지 설계해야 한다. 먼저 대상 row를 select해서 임시 테이블이나 파일로 보존한다. 그다음 같은 조건으로 count를 확인하고, 변경 SQL에는 가능하면 request id, primary key range, 상태 조건을 함께 넣는다. 자동 커밋을 끄고 명시적 transaction을 열었다면, 너무 오래 열어 vacuum이나 purge, lock wait을 만들지 않도록 작은 batch로 나눈다. 안전한 운영 SQL은 `한 번에 되돌릴 수 있는가`, `대상이 정확히 보존되었는가`, `실패하면 어디서 재시작하는가`를 답할 수 있어야 한다.

```sql
-- 보정 전 대상 보존
CREATE TEMP TABLE fix_transfer_target AS
SELECT transfer_id, status, updated_at
FROM transfers
WHERE status = 'REQUESTED'
  AND created_at < now() - interval '1 day'
  AND external_state = 'EXPIRED';

SELECT count(*) FROM fix_transfer_target;

BEGIN;
UPDATE transfers t
SET status = 'CANCELLED', updated_at = now()
FROM fix_transfer_target f
WHERE t.transfer_id = f.transfer_id
  AND t.status = 'REQUESTED';
-- affected row count가 fix_transfer_target count와 맞는지 확인한 뒤 commit한다.
COMMIT;
```

이 검증에서 PASS는 `대상 보존 count`, `UPDATE affected row count`, `변경 후 status 분포`, `업무 로그의 request id`가 서로 맞는 것이다. FAIL은 보정 SQL이 성공했다는 메시지만 보고 끝내는 것이다. SQL 성공 메시지는 업무 성공의 충분조건이 아니다. 특히 autocommit 환경에서는 성공 메시지가 곧 복구 난이도 증가를 뜻할 수 있다.

### 커넥션 풀은 트랜잭션 경계의 증폭기다

서버 애플리케이션에서 트랜잭션은 보통 connection에 붙는다. 커넥션 풀은 물리 connection을 재사용하므로, 요청 A가 transaction을 제대로 닫지 못하면 요청 B가 같은 connection을 빌릴 때 오염된 상태를 받을 수 있다. 많은 pool과 framework는 반환 시 rollback/reset을 수행하지만, 이것을 막연히 믿으면 안 된다. 장애 분석에서는 pool 설정, transaction manager, 예외 전파, finally cleanup, timeout 설정을 함께 본다.

오염은 단순히 열린 transaction만 뜻하지 않는다. isolation level을 `SERIALIZABLE`로 올리고 원복하지 않은 connection, `read only` 상태가 남은 connection, session variable이 바뀐 connection, lock wait timeout이 바뀐 connection도 다음 요청에 영향을 준다. 그래서 transaction 경계 문서는 DB 내부 문서이면서 애플리케이션 boundary 문서이기도 하다. 경계는 SQL 파일이 아니라 request lifecycle 안에서 닫힌다.

```text
request A
  borrow connection C17
  BEGIN
  UPDATE accounts ...
  unexpected exception
  rollback 누락 또는 예외가 transaction manager 밖에서 삼켜짐
  return C17 to pool

request B
  borrow connection C17
  첫 SELECT/UPDATE가 이상하게 대기하거나 이전 상태의 영향을 받음

확인 순서
  application log: request id, connection id, transaction begin/end
  DB view: 오래 열린 transaction, state, query, wait event
  pool metric: active/idle/borrow timeout, reset failure count
```

이 흐름은 `트랜잭션은 DB가 알아서 닫는다`는 오해를 고친다. DB는 commit이나 rollback 명령을 받으면 경계를 닫는다. 애플리케이션이 그 명령을 보내지 못하면 세션은 열린 상태로 남을 수 있다. 따라서 좋은 코드는 성공 경로보다 예외 경로에서 더 엄격하다. 모든 exit path가 commit, rollback, close 중 하나로 끝나는지 확인해야 한다.

### commit 이후의 모르는 상태를 처리하는 법

분산 애플리케이션에서는 commit 성공 여부를 애플리케이션이 모르는 경우가 자주 생긴다. DB가 commit을 완료했지만 응답 패킷이 끊길 수 있고, 애플리케이션 프로세스가 commit 직후 죽을 수도 있다. 이때 naive retry는 같은 이체를 두 번 만들 수 있다. 그래서 트랜잭션 경계와 idempotency는 같이 배워야 한다. request id를 unique key로 두고, 성공 결과를 response table이나 ledger에서 재구성할 수 있어야 한다.

```sql
CREATE TABLE transfer_requests (
  request_id text PRIMARY KEY,
  transfer_id bigint UNIQUE,
  status text NOT NULL CHECK (status IN ('PROCESSING', 'POSTED', 'FAILED')),
  response_json text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

처음 요청은 `request_id='R-9001'` row를 만들고 이체를 진행한다. commit 후 응답이 끊기면 재시도 요청은 같은 request id insert에서 unique 충돌을 만난다. 이 충돌은 무조건 실패가 아니라 기존 처리 결과를 조회하라는 신호다. 기존 row가 `POSTED`라면 같은 성공 응답을 재생하고, `PROCESSING`이 오래 남아 있으면 owner timeout과 reconciliation으로 간다. commit 후 모르는 상태를 설계하지 않으면, 트랜잭션 자체는 정확해도 API는 중복 부작용을 만들 수 있다.

### 트랜잭션 길이는 업무 원자성과 정리 비용 사이의 선택이다

큰 트랜잭션은 all-or-nothing을 이해하기 쉽게 만든다. 그러나 update가 많은 큰 transaction은 lock 보유 시간, WAL/redo/undo 증가, replication lag, vacuum/purge 지연, rollback 비용을 키운다. 작은 transaction은 장애 반경을 줄이고 재시작을 쉽게 만들 수 있지만, 업무가 정말 전체 원자성을 요구하면 중간 성공 상태가 문제가 된다. 따라서 `무조건 크게`나 `무조건 작게`는 둘 다 약하다.

batch 설계에서는 먼저 업무 원자성의 단위를 정한다. 정산 파일 전체가 한 번에 성공해야 하는지, 가맹점 단위로 성공해도 되는지, 거래 row 단위로 재시작 가능한지 확인한다. 그다음 각 단위마다 progress table을 두고, commit 뒤 관측 가능한 checkpoint를 남긴다. 실패한 단위는 재시도하고, 성공한 단위는 idempotent하게 건너뛴다. 이 구조가 있으면 작은 transaction을 쓰면서도 업무 일관성을 잃지 않는다.

```text
batch transfer posting
  unit = merchant_id + settlement_date
  state table = REQUESTED -> POSTING -> POSTED -> RECONCILED
  each unit transaction:
    1. claim unit with conditional update
    2. post ledger rows
    3. write posting summary
    4. commit
  retry:
    POSTING이 timeout이면 ledger/request_id로 실제 반영 여부 확인 후 재개
```

여기서 senior failure trap은 progress state와 DB transaction을 같은 것으로 착각하는 것이다. progress state는 여러 transaction에 걸친 업무 진행률이고, DB transaction은 한 번의 commit 경계다. 둘을 분리하면 batch를 작게 나누면서도 전체 업무를 추적할 수 있다.

### 검증을 면접 답변이 아니라 실험으로 닫기

이 DU를 제대로 통과했는지 확인하려면 두 세션과 장애 주입을 함께 써야 한다. 첫째, autocommit 상태에서 출금 update만 실행한 뒤 장애를 가정하고 잔액 합계가 깨지는지 본다. 둘째, 명시적 transaction에서 같은 장애를 만들고 rollback 뒤 원상 복귀되는지 본다. 셋째, savepoint 안의 optional 작업만 실패시키고 본체 commit 여부를 확인한다. 넷째, commit 직후 응답 실패를 흉내 내고 request id 재시도가 기존 결과를 재생하는지 본다.

```text
PASS로 볼 수 있는 관측
  - 실패 전후 accounts 총액이 기대와 맞다.
  - ledger row는 본체 성공 때만 남는다.
  - optional savepoint 실패가 본체 원자성을 흐리지 않는다.
  - 요청 종료 뒤 열린 transaction이 남지 않는다.
  - commit 후 응답 실패 재시도는 새 이체가 아니라 기존 결과 조회로 닫힌다.

FAIL로 볼 수 있는 관측
  - SQL 일부 성공을 API 성공으로 응답한다.
  - rollback 뒤 ledger나 balance 일부가 남는다.
  - pool 반환 뒤 다음 요청이 이전 transaction 상태를 물려받는다.
  - 운영 보정 SQL의 대상 보존 없이 autocommit update를 실행한다.
```

이 검증은 단순한 복습이 아니다. 트랜잭션 경계는 눈에 보이지 않기 때문에, 실패 지점과 관측 지표를 만들어야만 진짜로 이해된다. 문장 하나가 자동으로 commit되는 세계와 여러 문장이 하나의 운명을 공유하는 세계를 구분할 수 있으면, ACID의 A와 D는 추상 속성이 아니라 장애 처리 규칙으로 바뀐다.

### 로그와 지표로 경계를 증명하기

트랜잭션 경계는 코드 리뷰에서 보이는 annotation만으로 증명되지 않는다. 실제 운영에서는 request id, connection id, transaction begin/end, affected row count, commit latency, rollback count가 함께 남아야 한다. 특히 금융성 workflow에서는 `성공 응답을 보냈다`와 `DB commit이 성공했다`와 `외부 시스템도 같은 결과를 알고 있다`를 분리해서 추적해야 한다. 세 신호가 같은 request id로 이어지지 않으면 장애 시점에 어느 상태를 진실로 볼지 알 수 없다.

```text
권장 로그 흐름
  request_id=R-9001 tx=begin conn=C17
  request_id=R-9001 debit affected=1 account=100 amount=500
  request_id=R-9001 credit affected=1 account=200 amount=500
  request_id=R-9001 ledger inserted ledger_id=7781
  request_id=R-9001 tx=commit latency_ms=18
  request_id=R-9001 response=success transfer_id=7781
```

이 로그에서 빠지면 가장 위험한 값은 affected row count다. 조건부 UPDATE를 썼는데 affected=0을 무시하면 잔액 부족이나 상태 불일치를 성공으로 처리할 수 있다. commit latency도 중요하다. SQL 실행은 빠른데 commit에서 오래 걸리면 disk flush, replication, lock release, group commit 같은 아래층 문제를 봐야 한다. rollback count가 늘면 단순 실패율이 아니라 어떤 invariant가 자주 충돌하는지 확인해야 한다.

운영 대시보드도 transaction lifecycle을 반영해야 한다. 평균 SQL 시간만 보면 열린 transaction이 오래 남는 문제를 놓친다. `oldest transaction age`, `idle in transaction count`, `commit/rollback ratio`, `transaction timeout`, `pool borrow timeout`, `deadlock/serialization retry count`를 함께 봐야 한다. PostgreSQL에서는 `pg_stat_activity.xact_start`와 wait event를 보고, MySQL에서는 `information_schema.innodb_trx`와 InnoDB status를 본다. 이 관측이 있어야 transaction boundary가 문서 속 개념이 아니라 운영 가능한 계약이 된다.

### 외부 API를 트랜잭션 안에 둘 때의 실제 비용

DB transaction 안에서 외부 HTTP API를 기다리는 코드는 처음에는 자연스러워 보인다. `출금하고 PG 승인 받고 원장 쓰고 commit`처럼 한 함수 안에 있으면 이해하기 쉽다. 하지만 열린 DB transaction은 외부 API 지연 시간 동안 lock, snapshot, undo/WAL 자원을 붙잡을 수 있다. 외부 API가 3초 늦어지면 DB 내부 lock 보유 시간도 3초 늘어난다. 트래픽이 몰리면 이 구조는 DB 문제처럼 보이지만 원인은 외부 대기일 수 있다.

더 나은 설계 후보는 업무에 따라 다르다. 외부 승인 전에 내부 상태를 `PENDING`으로 commit하고 outbox worker가 외부 호출을 수행할 수 있다. 반대로 외부 승인 결과가 먼저 있어야 내부 원장을 만들 수 있다면, 외부 승인 id를 idempotent하게 저장하고 내부 반영을 retry 가능한 transaction으로 분리할 수 있다. 어느 쪽이든 하나의 DB transaction이 외부 세계 전체를 원자화한다고 믿으면 안 된다.

```text
위험한 단일 경계
  BEGIN
  UPDATE account lock 보유
  call external PG for 3s
  INSERT ledger
  COMMIT

분리한 경계 예
  tx1: request_id claim + PENDING 저장 + outbox insert -> COMMIT
  worker: external PG call
  tx2: approval_id 반영 + ledger POSTED + outbox done -> COMMIT
  reconciliation: PENDING/approval 불일치 추적
```

분리한 설계는 더 복잡하지만 실패 상태를 이름 붙일 수 있다. `PENDING인데 외부 호출 전`, `외부 승인 성공인데 내부 반영 전`, `내부 반영 성공인데 응답 실패`가 각각 관측 가능한 상태가 된다. 트랜잭션 경계를 잘 잡는다는 것은 모든 것을 한 transaction에 넣는 것이 아니라, 실패를 복구 가능한 상태로 나누는 것이다.

### rollback을 성공처럼 착각하지 않기

Rollback은 실패한 업무를 깨끗하게 지워 주는 도구처럼 보이지만, 실제로는 DB가 관리하는 변경에만 적용된다. 또한 큰 transaction의 rollback은 비용이 크다. 많은 row를 update한 뒤 실패하면 DB는 undo/WAL/redo 정보를 이용해 되돌려야 하고, 그동안 lock과 I/O가 계속 영향을 줄 수 있다. 그래서 대량 작업은 `실패하면 rollback하면 되지`라는 말로 충분하지 않다.

대량 작업에서는 사전 검증이 rollback보다 중요하다. 대상 row count를 먼저 확인하고, 작은 chunk로 나누고, 각 chunk가 idempotent하게 재실행 가능한지 확인한다. 중간 실패 후 재개할 수 있는 progress table이 있으면 rollback 범위를 줄일 수 있다. 업무적으로 전체 all-or-nothing이 반드시 필요하다면 큰 transaction을 감수할 수 있지만, 그때도 lock timeout, replication lag, rollback 시간, 운영 중단 가능성을 미리 계산해야 한다.

```text
대량 보정 전 질문
  - 이 작업의 원자성 단위는 전체 파일인가, 고객인가, row인가?
  - 실패한 chunk를 다시 실행해도 같은 결과가 되는가?
  - rollback이 시작되면 몇 분 동안 lock을 잡을 수 있는가?
  - commit 뒤 응답 실패나 작업자 crash를 어떻게 구분하는가?
  - 작업 전 대상 row의 원본을 어디에 보존했는가?
```

이 질문이 닫히면 rollback은 막연한 안전망이 아니라 비용이 있는 복구 도구로 자리 잡는다. 트랜잭션 lifecycle의 실무 감각은 commit만큼 rollback의 비용을 아는 데서 완성된다.

### 마지막 경계 질문: 누구의 성공인가

트랜잭션 경계를 설계할 때 마지막으로 물어야 할 질문은 `누구의 성공을 말하는가`이다. DB 입장에서는 commit이 성공이면 현재 transaction은 성공이다. API 입장에서는 commit 뒤 응답이 사용자에게 도달해야 성공처럼 보인다. 외부 PG사나 메시지 브로커까지 얽히면 성공의 주체가 더 늘어난다. 이 셋을 한 단어로 뭉개면 장애가 났을 때 서로 다른 진실이 충돌한다.

예를 들어 DB commit은 성공했고 HTTP 응답은 실패했다면, 서버 로그에는 exception이 남고 DB에는 ledger가 남는다. 사용자는 실패로 보고 다시 버튼을 누른다. 이때 시스템이 `이전 request_id는 이미 POSTED이고 transfer_id=7781이다`라고 답할 수 있으면 성공이다. 반대로 새 transfer를 하나 더 만들면 트랜잭션 자체는 매번 원자적으로 성공했지만 업무는 중복 처리된다. 그래서 트랜잭션 경계는 idempotency와 함께 완성된다.

```text
성공 주체 분리
  DB 성공       : COMMIT이 durable state를 만들었다.
  API 성공      : client가 같은 결과를 확인했다.
  외부 성공     : PG/메시지/파일 등 DB 밖 부작용이 기대 상태가 되었다.
  업무 성공     : 위 상태들이 request_id나 external_id로 대사 가능하다.
```

이 네 줄을 구분하면 장애 대응 언어도 선명해진다. `DB는 성공했지만 API 응답이 실패했다`, `외부 승인은 성공했지만 내부 원장이 없다`, `내부 원장은 성공했지만 메시지 publish가 지연 중이다`처럼 말할 수 있다. 이런 문장을 만들 수 있어야 트랜잭션 lifecycle을 실제 시스템 설계에 적용한 것이다.
## consistency와 도메인 불변식

Consistency가 별도 속성으로 강조된 배경은 데이터베이스가 단순 저장소가 아니라 업무 상태를 오래 보존하는 공유 장부가 되었기 때문이다. ACID의 Consistency는 데이터베이스가 모든 비즈니스 규칙을 자동으로 알아서 맞춘다는 뜻이 아니다. 트랜잭션 전후의 상태가 선언된 제약조건과 도메인 불변식을 만족해야 한다는 뜻이다. 어떤 불변식은 `NOT NULL`, `CHECK`, `UNIQUE`, `FOREIGN KEY`로 DB에 선언할 수 있고, 어떤 불변식은 애플리케이션 서비스, 상태 머신, 동시성 제어, 외부 대사 작업까지 함께 있어야 지킬 수 있다.

불변식은 상태가 바뀌어도 계속 참이어야 하는 문장이다. `잔액은 음수가 아니다`, `request_id는 한 번만 성공한다`, `정산 마감 후에는 취소할 수 없다`, `외부 결제 승인과 내부 원장 기록은 대사 가능해야 한다`가 모두 불변식이다. Consistency를 제대로 다루려면 이 문장을 먼저 쓰고, 각 문장을 DB가 강제할 수 있는 것과 애플리케이션이 책임져야 하는 것으로 나눠야 한다.

이 절에서 하중을 지탱하는 공식 자료는 다음과 같다. 링크 자체가 결론은 아니며, 본문에서는 각 공식 문서가 제공하는 규칙을 계좌 이체, row version, read view, vacuum/purge trace와 연결해 다시 검증 가능한 형태로 풀어 쓴다.

- PostgreSQL constraints: https://www.postgresql.org/docs/current/ddl-constraints.html
- MySQL constraint invalid data handling: https://dev.mysql.com/doc/refman/8.4/en/constraint-invalid-data.html

```sql
CREATE TABLE accounts (
  account_id bigint PRIMARY KEY,
  owner_id bigint NOT NULL,
  balance numeric(12,2) NOT NULL CHECK (balance >= 0),
  status text NOT NULL CHECK (status IN ('OPEN', 'FROZEN', 'CLOSED'))
);
CREATE TABLE transfers (
  transfer_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  request_id text NOT NULL UNIQUE,
  from_account bigint NOT NULL REFERENCES accounts(account_id),
  to_account bigint NOT NULL REFERENCES accounts(account_id),
  amount numeric(12,2) NOT NULL CHECK (amount > 0),
  status text NOT NULL CHECK (status IN ('REQUESTED', 'POSTED', 'CANCELLED')),
  CHECK (from_account <> to_account)
);
```

| 불변식 | DB constraint | 애플리케이션 책임 | 실패 증상 |
|---|---|---|---|
| 잔액은 음수가 아니다 | CHECK 가능 | 조건부 UPDATE와 동시성 처리 | 음수 잔액 |
| 요청 id는 한 번만 성공한다 | UNIQUE 가능 | idempotent response 저장 | 중복 결제 |
| 동결 계좌 출금 금지 | 상태 값 CHECK만으로 부족 | status 조건 write | 동결 후 출금 |
| 월 한도 초과 금지 | 단일 row CHECK로 부족 | 집계/락/serializable retry | 동시에 한도 초과 |
| 외부 승인과 내부 원장 대응 | DB 단독 불가 | outbox/reconciliation | 승인 성공, 내부 미반영 |

### 1. DB가 잘 지키는 불변식부터 내려야 한다

값이 비어 있으면 안 된다, 금액은 양수다, 부모 row가 있어야 한다, 같은 request id는 한 번만 들어간다는 규칙은 DB가 안정적으로 강제할 수 있다. 이런 규칙을 애플리케이션 if문에만 두면 batch, 관리자 SQL, 다른 서비스, 복구 스크립트가 쉽게 우회한다.

constraint는 장애 로그와 운영자에게도 계약을 드러낸다. 이름이 `chk_transfer_amount_positive`처럼 원인을 말하면 실패 로그만 보고도 어떤 불변식이 깨졌는지 알 수 있다.

### 2. DB가 모르는 도메인 규칙을 과장하지 않는다

DB는 `FROZEN 계좌에서 출금 금지`라는 업무 의미를 column 이름만 보고 알지 못한다. `status` 값 집합은 CHECK로 제한할 수 있지만, 어떤 상태에서 어떤 전이가 가능한지는 조건부 UPDATE, trigger, service method, state machine이 함께 책임져야 한다.

이 구분이 없으면 Consistency를 DB 마법으로 오해한다. 좋은 설계는 DB에 structural guard를 두고, 도메인 서비스에는 전이 의도와 실패 처리를 둔다.

### 3. 읽고 판단한 뒤 쓰는 코드는 race에 약하다

애플리케이션이 `SELECT balance`로 값을 읽고 메모리에서 판단한 뒤 별도 `UPDATE`를 보내면 그 사이에 다른 transaction이 같은 row를 바꿀 수 있다. 동시 출금에서 두 요청이 같은 잔액을 읽고 모두 통과하는 문제가 대표적이다.

더 안전한 방식은 판단과 쓰기를 한 SQL에 묶는 것이다. `UPDATE accounts SET balance = balance - :amount WHERE account_id = :id AND balance >= :amount AND status = OPEN`처럼 쓰고 affected row count를 성공/실패 신호로 삼는다.

### 4. 집계 불변식은 단일 row constraint로 닫히지 않는다

월 한도, 일일 누적, 좌석 잔여 수, 쿠폰 전체 사용량처럼 여러 row의 합이나 시간 범위를 보는 규칙은 단일 row CHECK로 표현하기 어렵다. 두 transaction이 같은 집계값을 읽고 각자 통과하면 commit 뒤 전체 합계가 깨진다.

해결 후보는 counter row lock, 조건부 atomic update, SERIALIZABLE과 retry, 별도 ledger와 보정 등이다. 어떤 후보가 맞는지는 즉시 거부가 필요한지, 실패 비용이 큰지, retry가 가능한지에 따라 달라진다.

### 5. NULL과 3값 논리는 constraint를 흐리게 만든다

`CHECK (amount > 0)`만 두고 `NOT NULL`을 빼면 NULL이 의도와 다르게 통과하는 형태가 될 수 있다. SQL의 UNKNOWN은 애플리케이션 boolean과 다르다. 그래서 필수 값은 NOT NULL과 CHECK를 함께 두고 negative test로 확인한다.

UNIQUE도 NULL 처리와 복합 key 설계에 따라 의미가 달라진다. constraint는 선언했다고 끝이 아니라 허용 값 공간을 실제 insert/update로 확인해야 한다.

### 6. 커밋 시점 consistency와 문장 시점 consistency를 구분한다

어떤 제약은 각 문장 끝에 검사되고, 어떤 제약은 deferrable하게 커밋 시점까지 미룰 수 있다. 중간 상태가 잠시 깨져도 커밋 전 복구되면 허용되는 설계가 있고, 한 문장마다 반드시 지켜야 하는 설계가 있다.

이 구분 없이 `항상 일관적이어야 한다`라고만 말하면 실제 트랜잭션 내부 상태를 설명할 수 없다. 업무가 요구하는 것은 중간 상태 금지인지, 외부 관측 금지인지, 커밋 상태 보존인지 분리해야 한다.

### 7. 외부 시스템과의 consistency는 DB constraint 밖에 있다

PG 승인, 메시지 발행, 파일 생성, 다른 서비스 호출은 DB constraint가 직접 강제하지 못한다. DB transaction rollback도 이미 나간 외부 호출을 되감지 못한다.

그래서 외부 대응 관계는 outbox, saga, idempotency key, reconciliation report로 닫는다. 이 영역을 DB ACID 안에 있다고 말하면 장애 복구 설계가 비게 된다.

### 8. 마이그레이션은 기존 dirty data를 먼저 만난다

새 constraint를 추가할 때 현재 table에 이미 위반 row가 있으면 migration은 실패하거나 데이터를 잘못 보정한다. 운영에서는 `ALTER TABLE ADD CONSTRAINT` 전에 위반 후보를 SELECT하고, 보정 정책과 rollback script를 준비해야 한다.

Consistency는 미래 입력만 막는 것이 아니라 과거 데이터의 실제 상태를 인정하고 정리하는 일이다. 이 단계를 생략하면 배포 중단 또는 조용한 데이터 손상이 생긴다.

### 9. Consistency와 Isolation은 서로 기대지만 같은 말이 아니다

Consistency는 허용된 상태 집합 안에 남는 문제고, Isolation은 동시 transaction들이 서로의 중간 상태를 어떻게 보느냐의 문제다. 한도 race처럼 consistency 위반은 isolation 선택이 약할 때 드러나지만, 두 단어를 합치면 원인을 놓친다.

분석 순서는 불변식을 먼저 쓰고, 그 불변식을 깨는 동시 실행 순서를 그린 뒤, 어떤 isolation/lock/write pattern이 그 순서를 막는지 확인하는 것이다.

### 10. 검증은 성공 입력보다 실패 입력이 더 중요하다

constraint와 domain invariant는 실패해야 할 입력을 넣어 봐야 강도를 알 수 있다. 음수 금액, 없는 계좌, 중복 request id, 동결 계좌 출금, 동시에 한도 초과하는 요청을 각각 테스트해야 한다.

PASS는 실패가 DB나 서비스의 정해진 error contract로 닫히고 상태가 남지 않는 것이다. FAIL은 exception은 났지만 일부 row가 남거나, 실패가 warning으로 바뀌거나, 재시도 때 다른 결과를 만드는 것이다.

```sql
-- 실패해야 하는 입력을 직접 넣어 보는 검증
INSERT INTO transfers(request_id, from_account, to_account, amount, status)
VALUES ('r-neg', 100, 200, -1, 'REQUESTED');

INSERT INTO transfers(request_id, from_account, to_account, amount, status)
VALUES ('r-missing', 999, 200, 10, 'REQUESTED');

-- 동시성 race를 줄이는 조건부 write
UPDATE accounts
SET balance = balance - 500
WHERE account_id = 100
  AND status = 'OPEN'
  AND balance >= 500;
```

검증의 PASS 기준은 실패해야 하는 입력이 실패하고, 실패 뒤에도 table에 반쪽 상태가 남지 않으며, 동시 요청에서도 한도와 잔액 불변식이 깨지지 않는 것이다. 관측은 constraint violation 로그, duplicate key 발생률, affected row count, reconciliation diff, dead letter queue, 보정 ledger로 한다. 이 관측이 없으면 Consistency는 설계 문장일 뿐 운영 계약이 아니다.
### Consistency는 불변식 목록에서 시작한다

Consistency를 가장 많이 오해하는 방식은 `트랜잭션을 쓰면 DB가 알아서 정상 상태를 만든다`고 생각하는 것이다. DB는 선언된 제약조건을 강제할 수 있고, 특정 격리 수준과 lock으로 동시 실행을 조정할 수 있으며, commit/rollback으로 경계를 닫을 수 있다. 하지만 무엇이 정상 상태인지는 schema와 애플리케이션이 먼저 말해야 한다. `잔액은 음수가 아니어야 한다`는 DB가 CHECK로 지킬 수 있다. `정산 마감 후에는 취소할 수 없다`는 상태 전이 규칙과 업무 시간, 권한, 외부 정산 파일까지 알아야 한다. 즉 consistency는 DB 마법이 아니라 불변식의 배치 문제다.

좋은 설계는 불변식을 세 층으로 나눈다. 첫째는 DB가 직접 강제해야 하는 구조 규칙이다. primary key, foreign key, unique, not null, check가 여기에 들어간다. 둘째는 한 SQL 안에 넣어야 안전한 조건부 쓰기 규칙이다. `balance >= amount`, `status = 'OPEN'`, `remaining_count > 0` 같은 조건이 여기에 들어간다. 셋째는 여러 row, 외부 시스템, 시간 흐름을 묶는 업무 규칙이다. 월 한도, 정산 마감, 외부 승인과 내부 원장 대응, 보정 가능성은 별도 상태 머신과 대사 작업이 필요하다.

```text
불변식 배치표

불변식 문장                           DB constraint       조건부 쓰기       서비스/운영 책임
------------------------------------  ------------------  --------------  -------------------------------
계좌 id는 유일하다                     PRIMARY KEY         -               -
이체 금액은 0보다 커야 한다             CHECK + NOT NULL    -               입력 오류 응답
같은 request_id는 한 번만 처리된다       UNIQUE             UPSERT/claim     기존 응답 replay
출금 후 잔액은 음수가 되면 안 된다        CHECK              WHERE balance>=  실패 응답/재시도 정책
동결 계좌에서는 출금할 수 없다            status 값 CHECK    WHERE status     상태 전이 owner
월 한도는 누적 합계를 넘으면 안 된다       -                  counter update   집계 lock/retry/reconciliation
외부 승인과 내부 원장은 대사 가능해야 한다  -                  outbox insert    reconciliation/report/보정
```

이 표를 쓰는 이유는 책임을 회피하지 않기 위해서다. DB constraint로 내려야 할 것을 서비스 if문에만 두면 우회 경로가 생긴다. 반대로 DB가 모르는 외부 상태까지 constraint가 해결한다고 믿으면 장애 복구가 비어 버린다. Consistency의 실력은 어떤 규칙을 어디에 둘지 말할 수 있는 능력에서 드러난다.

### 제약조건은 문서가 아니라 실행되는 계약이다

PostgreSQL과 MySQL 공식 문서는 primary key, foreign key, check 같은 제약조건을 schema 수준의 규칙으로 다룬다. 이 규칙들은 단순 문서화가 아니라 insert/update/delete 실행을 거부하는 장치다. 그래서 제약조건 이름은 운영 품질에 영향을 준다. `transfers_chk_1`보다 `chk_transfers_amount_positive`가 낫다. 장애 로그에 constraint 이름이 남을 때 어떤 불변식이 깨졌는지 바로 보이기 때문이다.

제약조건이 강할수록 애플리케이션 코드는 단순해질 수 있다. 하지만 단순해진다는 말은 검증을 안 해도 된다는 뜻이 아니다. 애플리케이션은 사용자에게 자연스러운 오류를 돌려주고, 중복 요청이면 기존 결과를 재생하고, constraint violation이 예상 가능한 업무 실패인지 개발 버그인지 분류해야 한다. DB가 거부한 순간은 끝이 아니라 error contract가 시작되는 지점이다.

```sql
ALTER TABLE transfers
  ADD CONSTRAINT chk_transfers_amount_positive CHECK (amount > 0),
  ADD CONSTRAINT chk_transfers_distinct_accounts CHECK (from_account <> to_account),
  ADD CONSTRAINT uq_transfers_request_id UNIQUE (request_id);

-- 검증용 실패 입력: 모두 실패해야 한다.
INSERT INTO transfers(request_id, from_account, to_account, amount, status)
VALUES ('T-BAD-AMOUNT', 100, 200, 0, 'REQUESTED');

INSERT INTO transfers(request_id, from_account, to_account, amount, status)
VALUES ('T-SAME-ACCOUNT', 100, 100, 100, 'REQUESTED');
```

여기서 PASS는 insert가 실패했다는 사실만이 아니다. 실패 뒤 row가 남지 않아야 하고, 애플리케이션은 어떤 불변식이 깨졌는지 사용자가 이해할 수 있는 오류로 바꿔야 하며, 운영 로그에는 request id와 constraint name이 남아야 한다. FAIL은 DB exception을 500으로만 던지고, 재시도 때 또 같은 exception을 만들며, 어느 규칙이 깨졌는지 아무도 알 수 없는 상태다.

### 읽고 판단한 뒤 쓰는 구조를 조건부 쓰기로 바꾸기

동시성에서 consistency가 깨지는 대표 경로는 `SELECT로 읽고 애플리케이션에서 판단한 뒤 UPDATE`하는 구조다. 두 요청이 동시에 balance 1000을 읽고 각각 700 출금을 허용하면, 각 요청은 자기 기준으로는 정상이다. 하지만 commit 뒤 잔액은 음수가 될 수 있다. 이 문제는 불변식을 읽기와 쓰기 사이에 흩어 놓았기 때문에 생긴다.

더 강한 구조는 불변식을 쓰기 조건 안에 넣는 것이다. `UPDATE accounts SET balance = balance - :amount WHERE account_id = :id AND balance >= :amount AND status = 'OPEN'`은 판단과 상태 변경을 한 statement로 묶는다. affected row count가 1이면 성공이고 0이면 잔액 부족 또는 상태 불일치다. 이 방식은 DB가 row lock과 조건 재검사를 통해 중간 race를 줄이게 만든다.

```text
나쁜 흐름
  A: SELECT balance=1000
  B: SELECT balance=1000
  A: if 1000 >= 700 then UPDATE balance=300
  B: if 1000 >= 700 then UPDATE balance=-400

더 안전한 흐름
  A: UPDATE ... WHERE balance >= 700  -- affected=1
  B: UPDATE ... WHERE balance >= 700  -- A 이후 조건 재평가, affected=0
```

이 흐름에서 중요한 관측값은 affected row count다. affected row count를 버리면 DB가 준 성공/실패 신호를 버리는 것이다. 서비스는 affected=0을 단순 장애가 아니라 업무 거부로 해석해야 한다. 그리고 이 거부가 너무 자주 발생한다면 잔액 부족인지, 동시 요청 경쟁인지, 잘못된 상태 전이인지 metric으로 나눠야 한다.

### 집계 불변식은 별도 설계가 필요하다

`월 한도는 100만 원을 넘으면 안 된다` 같은 규칙은 단일 row CHECK로 표현하기 어렵다. 거래 ledger 여러 row의 합계를 봐야 하고, 시간 구간도 들어간다. 두 transaction이 같은 월 누적 90만 원을 읽고 각각 8만 원을 추가하면, 각자 읽은 시점에는 허용되지만 commit 뒤 누적은 106만 원이 된다. 이것은 constraint가 없는 것이 아니라 불변식의 모양이 여러 row와 시간 구간을 가로지르기 때문에 생기는 문제다.

해결 후보는 여러 가지다. counter row를 두고 그 row를 조건부 update할 수 있다. SERIALIZABLE isolation과 retry로 동시 schedule을 DB가 거부하게 만들 수 있다. ledger append만 허용하고 사후 대사로 보정하는 업무도 있을 수 있다. 선택은 업무 실패 비용과 사용자 경험에 따라 달라진다. 즉시 거부가 필요한 선불 잔액 차감은 조건부 update가 어울리고, 월말 리포트의 통계 보정은 reconciliation이 더 현실적일 수 있다.

```sql
-- 월 한도 counter를 한 row에 모아 조건부로 갱신하는 예
UPDATE monthly_transfer_limits
SET used_amount = used_amount + :amount
WHERE account_id = :account_id
  AND yyyy_mm = :yyyy_mm
  AND used_amount + :amount <= limit_amount;
```

이 SQL도 완전한 답은 아니다. counter row가 없을 때 insert 경쟁을 어떻게 처리할지, 월 경계 timezone을 무엇으로 볼지, 취소가 used_amount를 되돌리는지, ledger와 counter가 불일치할 때 어떤 쪽을 source of truth로 볼지 정해야 한다. Consistency를 깊게 다룬다는 것은 이런 꼬리 질문을 `나중에`로 미루지 않고 설계 표면에 올리는 것이다.

### NULL과 UNKNOWN은 불변식의 틈을 만든다

SQL의 CHECK는 애플리케이션의 boolean if와 다르게 동작한다. 특히 NULL이 들어가면 비교 결과가 UNKNOWN이 될 수 있다. `CHECK (amount > 0)`만 두고 `amount`를 nullable로 두면 의도와 다르게 NULL이 통과할 수 있다. 그래서 필수 값에는 `NOT NULL`이 먼저 필요하고, CHECK는 값의 범위를 제한하는 데 써야 한다. 이 작은 차이를 놓치면 `제약조건을 걸었는데 빈 값이 들어왔다`는 이상한 장애가 생긴다.

복합 unique도 마찬가지다. nullable column이 섞인 unique constraint는 DBMS별 세부 동작과 요구 의미를 확인해야 한다. 사용자가 기대하는 `한 사용자당 활성 카드 하나` 같은 규칙은 partial unique index나 상태 조건을 써야 할 수 있고, MySQL에서는 generated column이나 별도 설계를 고려해야 할 수 있다. 제약조건은 선언문 하나로 끝나는 것이 아니라, 허용 값 공간을 실패 입력으로 직접 확인해야 닫힌다.

```text
negative test 목록
  amount = NULL       -> 실패해야 하는가, 허용해야 하는가?
  amount = 0          -> 실패해야 한다면 어떤 constraint가 막는가?
  status = 'DONE'     -> 허용 상태 집합 밖이면 어디서 막는가?
  request_id = NULL   -> 중복 방지 의미가 유지되는가?
  same from/to account -> DB constraint인지 service validation인지?
```

이 목록은 QA 체크리스트가 아니라 설계 질문이다. 각 실패 입력이 어느 계층에서 막히는지 답할 수 없으면, consistency 책임이 흩어져 있다는 뜻이다.

### 외부 시스템과의 일관성은 대사 가능한 상태로 만든다

DB constraint는 외부 PG 승인, 메시지 브로커 publish, 파일 업로드, 다른 서비스의 상태를 직접 보지 못한다. 그래서 `DB transaction 안에서 외부 API를 호출하면 같이 원자적일 것`이라는 생각은 위험하다. DB rollback은 이미 전송된 HTTP 요청을 되돌리지 못하고, 외부 성공 뒤 DB commit이 실패하면 외부와 내부가 갈라진다. 이 갈라짐을 0으로 만들 수 없다면, 적어도 찾고 고칠 수 있게 만들어야 한다.

실무에서는 상태를 `REQUESTED`, `EXTERNAL_APPROVED`, `LEDGER_POSTED`, `RECONCILED`처럼 나누거나, outbox table에 발행해야 할 이벤트를 같은 DB transaction으로 남긴다. commit이 성공하면 별도 publisher가 outbox를 읽어 외부로 보낸다. 외부 응답이 먼저 필요한 업무라면 request id와 외부 승인 id를 저장하고, 내부 반영 실패를 reconciliation job이 찾는다. 어떤 방식이든 consistency는 `항상 한 번에 맞는다`가 아니라 `불일치가 생기면 어느 row와 어느 id로 찾고 복구할 수 있다`까지 포함한다.

```text
외부 승인과 내부 원장 불일치 관측
  payment_request_id
  external_approval_id
  transfer_id
  ledger_posted_at
  outbox_event_id
  reconciliation_status

정상:
  external_approval_id 있음 -> ledger row 있음 -> reconciliation_status='MATCHED'

위험:
  external_approval_id 있음 -> ledger row 없음 -> 보정 또는 취소 필요
  ledger row 있음 -> external approval 없음 -> 내부 오기입 또는 중복 처리 의심
```

이런 대사 표가 없으면 장애 대응은 로그 검색과 기억에 의존한다. Consistency를 DB 안쪽 속성으로만 가르치면 이 실무 실패가 빠진다.

### 마이그레이션은 과거 데이터와 먼저 충돌한다

운영 테이블에 새 CHECK나 foreign key를 추가하는 순간, DB는 과거 데이터까지 검사한다. 이미 위반 row가 있으면 migration은 실패할 수 있다. 더 위험한 경우는 급하게 데이터를 보정하면서 원래 의미를 잃는 것이다. 예를 들어 NULL amount를 0으로 채워 constraint를 통과시켰지만, 실제로는 외부 결제 금액을 다시 조회해야 하는 값이었다면 데이터 손상이다.

따라서 constraint 추가 전에는 위반 후보를 조회하고, 각 후보의 업무 처리를 정해야 한다. 일부는 삭제, 일부는 외부 시스템 재조회, 일부는 상태 변경, 일부는 수동 보류가 될 수 있다. migration은 schema diff가 아니라 데이터 상태를 바꾸는 운영 작업이다.

```sql
-- constraint 추가 전 위반 후보 확인
SELECT transfer_id, request_id, amount, status
FROM transfers
WHERE amount IS NULL OR amount <= 0;

SELECT t.transfer_id, t.from_account
FROM transfers t
LEFT JOIN accounts a ON a.account_id = t.from_account
WHERE a.account_id IS NULL;
```

PASS는 migration이 성공했다는 것만이 아니다. 위반 후보 처리 근거가 남아 있고, 변경 전후 count가 맞고, rollback 또는 보정 경로가 있으며, 애플리케이션의 실패 응답이 새 constraint와 맞아야 한다. FAIL은 schema만 바꾸고 기존 dirty data를 운에 맡기는 것이다.

### 관측 가능한 consistency 대시보드

Consistency를 운영에서 지키려면 위반 후보를 숫자로 볼 수 있어야 한다. 성공 TPS와 평균 응답시간은 중요하지만, 불변식 위반 후보가 늘어나는지 알려 주지 않는다. 금융성 데이터에서는 orphan ledger, negative balance, duplicate request, unreconciled external approval, invalid state transition count가 더 직접적인 건강 신호다.

```sql
-- 예시: 정기적으로 0이어야 하는 후보들
SELECT count(*) AS negative_balance_count
FROM accounts
WHERE balance < 0;

SELECT request_id, count(*)
FROM transfers
GROUP BY request_id
HAVING count(*) > 1;

SELECT count(*) AS unreconciled_approval_count
FROM external_approvals ea
LEFT JOIN ledger l ON l.external_approval_id = ea.external_approval_id
WHERE ea.status = 'APPROVED'
  AND l.ledger_id IS NULL;
```

이 쿼리들은 평소에는 0이어야 한다. 0이 아닌 값이 보이면 이미 업무 불변식이 깨졌거나, 아직 대사되지 않은 중간 상태가 오래 남았다는 뜻이다. 어떤 값은 즉시 장애이고 어떤 값은 SLA 안의 지연일 수 있으므로, 각 count에는 허용 시간과 owner가 붙어야 한다. 좋은 consistency 문서는 constraint 선언뿐 아니라 이런 관측 쿼리까지 남긴다.

### Consistency 설명의 마지막 기준

이 DU의 PASS 조건은 ACID의 C를 한 문장으로 정의하는 것이 아니다. 특정 업무에서 불변식을 문장으로 쓰고, DB가 강제할 수 있는 부분과 애플리케이션이 책임질 부분을 나누고, 동시 실행에서 깨지는 순서를 그리고, 실패 입력으로 직접 확인하고, 운영에서 위반 후보를 관측할 수 있어야 한다. 이 다섯 단계가 닫히면 Consistency는 추상 단어가 아니라 설계와 검증의 연결고리가 된다.

```text
Consistency 자가 점검
  1. 이 업무에서 항상 참이어야 하는 문장을 썼는가?
  2. 각 문장이 DB constraint, 조건부 write, service state machine, reconciliation 중 어디에 있는가?
  3. 동시 요청 두 개가 그 문장을 깨는 순서를 그릴 수 있는가?
  4. 실패해야 하는 입력과 race를 실제로 재현했는가?
  5. 운영에서 위반 후보를 count와 request id로 찾을 수 있는가?
```

이 질문에 답할 수 없으면 아직 Consistency를 배운 것이 아니라 용어를 본 것이다. 반대로 답할 수 있으면 DB의 제약조건과 애플리케이션의 도메인 규칙을 서로 밀어내지 않고, 각자 잘하는 자리에서 같은 불변식을 지키게 만들 수 있다.

### 불변식은 테스트 이름으로도 드러나야 한다

Consistency 품질은 테스트 이름에서부터 드러난다. `transferTest`보다 `동결_계좌에서는_출금이_거부되고_잔액과_원장이_남지_않는다`가 강하다. 테스트 이름이 불변식 문장이 되면, 어떤 규칙을 지키려는지 코드 밖에서도 보인다. 성공 케이스도 필요하지만 consistency에서는 실패 케이스가 더 중요하다. 실패해야 하는 입력이 실패하고, 실패 뒤 상태가 깨끗하며, 재시도 결과가 안정적인지 확인해야 한다.

```text
테스트 matrix
  정상 이체:
    accounts 합계 보존, ledger 1건, transfer POSTED
  잔액 부족:
    affected row 0, ledger 없음, transfer FAILED 또는 거부 응답
  중복 request_id:
    새 ledger 없음, 기존 response replay
  동결 계좌:
    balance 변화 없음, 상태 전이 오류
  외부 승인 성공 + 내부 commit 실패:
    reconciliation 후보 생성
  내부 commit 성공 + 응답 실패:
    재시도 시 기존 결과 조회
```

이 matrix는 단순 테스트 수를 늘리기 위한 것이 아니다. 각 행은 불변식이 어느 계층에서 깨질 수 있는지 보여 준다. 잔액 부족은 조건부 update가 막아야 하고, 중복 요청은 unique key와 response replay가 막아야 하며, 외부 승인 불일치는 reconciliation이 찾아야 한다. 테스트가 이 책임 분리를 드러내지 못하면 실제 장애에서도 원인을 찾기 어렵다.

### constraint 이름과 오류 번역도 도메인 설계다

DB constraint violation을 사용자에게 그대로 보여 주면 기술적으로는 정확해도 업무적으로는 약하다. `chk_transfers_amount_positive`가 깨졌다면 API는 `이체 금액은 0보다 커야 합니다`처럼 도메인 언어로 바꿔야 한다. `uq_transfers_request_id`가 깨졌다면 무조건 오류로 보낼 것이 아니라 기존 요청 결과를 조회해야 한다. 같은 DB 오류라도 업무 의미가 다르기 때문이다.

이 번역은 catch 블록에서 문자열을 대충 파싱하라는 뜻이 아니다. 가능하면 constraint name, SQL state, vendor error code를 안정적으로 분류하고, 각 분류가 어떤 업무 응답으로 이어지는지 표로 둔다. 이 표는 운영자에게도 도움이 된다. constraint 이름을 보면 어떤 불변식이 자주 깨지는지 지표화할 수 있기 때문이다.

```text
constraint -> 업무 응답
  chk_transfers_amount_positive -> 400 INVALID_AMOUNT
  fk_transfers_from_account     -> 404 FROM_ACCOUNT_NOT_FOUND 또는 내부 데이터 오류
  uq_transfers_request_id       -> 기존 transfer/result 조회 후 replay
  chk_accounts_balance_nonneg   -> 409 INSUFFICIENT_BALANCE 또는 코드 버그 조사
```

여기서 senior failure trap은 모든 constraint violation을 같은 500으로 보내는 것이다. 그러면 사용자는 재시도하고, 재시도는 같은 constraint를 다시 깨며, 운영자는 어떤 업무 규칙이 흔들리는지 알 수 없다. Consistency는 DB 안에서 끝나지 않고 오류 언어와 재시도 정책까지 이어진다.

### consistency를 isolation으로만 해결하려는 함정

동시성 때문에 불변식이 깨졌다고 해서 항상 isolation level을 올리는 것이 답은 아니다. SERIALIZABLE은 강력하지만 재시도 비용과 실패 가능성을 만든다. 조건부 UPDATE 하나로 충분한 문제도 있고, unique key로 winner election을 만들 수 있는 문제도 있으며, counter row lock이 더 직관적인 문제도 있다. 먼저 불변식의 형태를 보고 가장 단순한 강제 지점을 고르는 것이 좋다.

```text
문제 모양별 후보
  한 row의 숫자 하한:
    조건부 UPDATE + affected row count
  중복 요청 winner 결정:
    UNIQUE request_id + insert/claim
  여러 row 집계 한도:
    counter row lock 또는 SERIALIZABLE retry
  외부 시스템 대응:
    outbox/inbox + reconciliation
  상태 전이 순서:
    conditional update WHERE current_status=...
```

이 표의 핵심은 수단을 목표로 착각하지 않는 것이다. 목표는 `불변식이 깨지지 않는 것`이고, isolation level은 후보 중 하나다. 같은 불변식이라도 트래픽, 실패 비용, 응답 요구, DBMS 동작에 따라 다른 수단이 더 나을 수 있다.

### 마지막 손계산: 두 요청이 같은 불변식을 밀 때

마지막으로 두 요청이 같은 계좌에서 700씩 출금하는 장면을 손으로 계산해 보자. 초기 잔액은 1000이다. 나쁜 구현은 둘 다 SELECT로 1000을 읽고 둘 다 통과한다. 좋은 구현은 조건부 UPDATE를 사용해 첫 요청만 affected=1을 받고 두 번째 요청은 affected=0을 받는다. 여기서 consistency는 `잔액이 음수가 되지 않는다`는 불변식이고, 조건부 UPDATE는 그 불변식을 DB write boundary에 놓는 수단이다.

```text
초기: balance=1000

나쁜 구현
  A read 1000 -> pass
  B read 1000 -> pass
  A write 300
  B write -400
  결과: 불변식 깨짐

조건부 write
  A UPDATE ... WHERE balance >= 700 -> affected=1, balance=300
  B UPDATE ... WHERE balance >= 700 -> affected=0, balance=300 유지
  결과: 불변식 유지, B는 업무 거부
```

이 손계산을 할 수 있으면 ACID의 C는 추상 단어가 아니다. 선언 가능한 규칙은 DB constraint로 내리고, 동시성에 민감한 판단은 write boundary로 가져오며, DB 밖의 대응 관계는 대사 가능하게 남긴다는 설계 원칙으로 바뀐다.

### 불변식의 owner를 정하지 않으면 결국 아무도 지키지 않는다

불변식은 문서에 적는 순간 자동으로 지켜지지 않는다. 각 불변식에는 owner가 필요하다. DB가 owner인 규칙은 schema migration과 constraint test가 지킨다. 애플리케이션이 owner인 규칙은 service method, 상태 전이 테스트, 오류 응답 계약이 지킨다. 운영 job이 owner인 규칙은 reconciliation query, alert, 수동 보정 runbook이 지킨다. owner가 없는 불변식은 장애 때마다 `원래 누가 봐야 했지?`라는 질문으로 돌아온다.

예를 들어 `외부 승인과 내부 원장은 대사 가능해야 한다`는 규칙의 owner가 DB constraint일 수는 없다. DB는 외부 승인 시스템의 최종 상태를 모른다. 이 규칙의 owner는 보통 payment workflow와 reconciliation job이다. 반대로 `request_id는 중복되면 안 된다`는 규칙은 DB unique constraint가 owner가 되는 편이 강하다. 서비스 코드만 owner로 두면 batch나 관리자 SQL이 우회할 수 있다.

```text
owner matrix
  DB schema owner:
    primary key, foreign key, not null, positive amount, unique request_id
  service owner:
    status transition, conditional write, duplicate replay, retry classification
  operations owner:
    external/internal mismatch, orphan ledger, stuck PROCESSING, dirty data cleanup
```

Consistency를 guru-level로 설명한다는 것은 결국 이 owner matrix를 말할 수 있다는 뜻이다. 모든 규칙을 DB로 밀어 넣지도 않고, 모든 규칙을 서비스 if문으로 빼지도 않는다. 규칙의 모양과 실패 비용을 보고 가장 강하고 관측 가능한 자리에 둔다.

### 마지막 판정: 상태를 되돌릴 수 있는가, 설명할 수 있는가

Consistency 장애 대응의 마지막 기준은 두 가지다. 첫째, 깨진 상태를 찾을 수 있어야 한다. 둘째, 찾은 상태를 어떤 근거로 보정할지 설명할 수 있어야 한다. negative balance 한 건을 발견했는데 어떤 request가 만들었는지, 어떤 ledger와 연결되는지, 외부 승인 상태가 무엇인지 모르면 보정 SQL은 추측이 된다. 그래서 불변식마다 관측 key가 필요하다. account_id, request_id, transfer_id, external_approval_id, ledger_id가 서로 이어져야 한다.

```text
보정 가능한 consistency 기록
  request_id -> transfer_id -> ledger_id -> account_id 변화
  request_id -> external_approval_id -> reconciliation_status
  transfer_id -> 이전 status -> 새 status -> 변경 actor
```

이 연결이 있으면 장애 뒤에도 사람은 상태를 설명할 수 있다. 연결이 없으면 DB constraint가 아무리 많아도 복구는 기억과 감에 의존한다. 따라서 Consistency의 최종 목표는 오류가 절대 안 나는 환상이 아니라, 허용되지 않은 상태를 최대한 앞에서 막고, 남은 불일치는 빠르게 발견하고, 근거 있는 방식으로 보정할 수 있게 만드는 것이다.
