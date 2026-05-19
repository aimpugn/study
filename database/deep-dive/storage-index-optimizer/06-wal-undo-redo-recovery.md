# WAL, undo, redo, recovery

이 파일은 log를 하나의 기록 파일로 뭉개지 않고 crash recovery와 PITR에서 각 log가 어떤 소비자에게 읽히는지 설명합니다.

## WAL/redo/undo를 값 변화 trace로 이해하기

WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이 절의 핵심 질문은 “한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/mvcc.md; database/replication.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/wal-intro.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-redo-log.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-undo-logs.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

write-ahead logging은 force/no-force, steal/no-steal recovery trade-off에서 나온 절충입니다. 모든 commit마다 data page를 강제로 쓰면 정상 성능이 너무 나빠지고, dirty page를 마음대로 미루면 crash 뒤 무엇이 반영되었는지 설명할 수 없습니다. log를 먼저 안정화하면 data page flush를 미루면서도 crash 뒤 redo 또는 undo 판단을 할 수 있습니다.

### 손으로 따라가는 trace

```text
initial data page P100
  accounts[42].balance = 10000

T1 update balance = 9000
  redo/WAL: page P100 change is described
  undo: old value 10000 can be reconstructed if needed
  buffer page P100 becomes dirty
  commit log record is flushed according to durability policy

crash before data page P100 reaches disk
  disk P100 may still say 10000
  log says committed update exists

recovery
  if committed log record exists and pageLSN is old -> redo balance=9000
  if uncommitted change exists -> undo/visibility hides or rolls it back

rule
  log durable before dirty data page flush -> recovery can decide safely
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. | InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. | WAL, write-ahead |
| 물리 모델 | PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. | InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. | redo log, pageLSN |
| 관측 모델 | PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. | InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. | undo log, LSN |
| 장애 모델 | PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. | InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. | write-ahead, commit record |

### 첫 번째 벽돌

첫 번째 벽돌에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 WAL, redo log, undo log 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 WAL을 백업 파일이나 replication log와 같은 것으로 뭉개는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

WAL을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_current_wal_lsn();` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, pageLSN와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 redo log, undo log, write-ahead 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 redo와 undo를 단순 반대말로만 외우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

redo log을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance - 1000 WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, LSN와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 undo log, write-ahead, pageLSN 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 PostgreSQL과 InnoDB의 old version 위치 차이를 무시하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

undo log을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, commit record와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 write-ahead, pageLSN, LSN 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 commit 성공과 애플리케이션 응답 성공을 같은 사건으로 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

write-ahead을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_os_log%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, dirty page와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 pageLSN, LSN, commit record 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 WAL을 백업 파일이나 replication log와 같은 것으로 뭉개는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

pageLSN을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- commit 직후 crash/restart rehearsal은 로컬 disposable DB에서만 수행한다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, rollback와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 LSN, commit record, dirty page 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 redo와 undo를 단순 반대말로만 외우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

LSN을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_current_wal_lsn();` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, consistent read와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 commit record, dirty page, rollback 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 PostgreSQL과 InnoDB의 old version 위치 차이를 무시하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

commit record을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance - 1000 WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, binlog와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 dirty page, rollback, consistent read 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 commit 성공과 애플리케이션 응답 성공을 같은 사건으로 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

dirty page을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, replication log와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 rollback, consistent read, binlog 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 WAL을 백업 파일이나 replication log와 같은 것으로 뭉개는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

rollback을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_os_log%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, WAL와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 한 row update가 data page보다 log에 먼저 남아야 crash 뒤 committed 변경과 uncommitted 변경을 구분할 수 있는 이유는 무엇인가라는 질문은 old value -> change log -> dirty page -> commit record -> crash replay or rollback 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. WAL, redo, undo는 모두 로그라는 이름으로 뭉개기 쉽지만, 서로 맡는 질문이 다릅니다. 이때 consistent read, binlog, replication log 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL WAL은 data file 변경보다 먼저 log를 남겨 crash recovery와 replication/PITR 재료가 되며, old version은 heap tuple과 transaction visibility 규칙으로 설명됩니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB redo log는 dirty page crash recovery를 위한 재적용 재료이고, undo log는 rollback과 consistent read의 old version 재구성에 관여합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 redo와 undo를 단순 반대말로만 외우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

consistent read을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old value -> change log -> dirty page -> commit record -> crash replay or rollback 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- commit 직후 crash/restart rehearsal은 로컬 disposable DB에서만 수행한다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, redo log와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- WAL을 백업 파일이나 replication log와 같은 것으로 뭉개는 함정
- redo와 undo를 단순 반대말로만 외우는 함정
- PostgreSQL과 InnoDB의 old version 위치 차이를 무시하는 함정
- commit 성공과 애플리케이션 응답 성공을 같은 사건으로 보는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
SELECT pg_current_wal_lsn();
EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance - 1000 WHERE id = 42;
SHOW ENGINE INNODB STATUS;
SHOW GLOBAL STATUS LIKE Innodb_os_log%;
-- commit 직후 crash/restart rehearsal은 로컬 disposable DB에서만 수행한다.
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

WAL이 있으니 SQL을 다시 실행한다고 생각하면 틀립니다. recovery는 애플리케이션 SQL 텍스트를 재실행하는 것이 아니라 page와 record 수준 변경을 일관된 상태로 맞춥니다. WAL archive와 backup도 같지 않습니다. WAL archive는 base backup 이후 변경을 이어 붙이는 재료이고, redo log는 replication용 binlog와도 목적이 다릅니다.

### 다시 설명해 보기

balance 10000에서 9000으로 update한 뒤 commit 전 crash, commit 후 data page flush 전 crash, checkpoint 후 crash 세 경우를 나눠 recovery 뒤 SELECT 결과를 써 보십시오. 각 경우 WAL/redo/undo 중 무엇이 판단 근거인지 표시해야 합니다.

### 통합 사례: commit 성공 응답 직전 process가 죽었을 때

애플리케이션이 결제 승인 row를 `APPROVED`로 update하고 DB commit은 성공했지만, HTTP 응답을 보내기 직전에 process가 죽었다고 가정합니다. DB 관점에서는 commit record가 durable하면 crash 뒤 row는 승인 상태로 복구되어야 합니다. 하지만 클라이언트는 응답을 못 받았으므로 실패로 인식하고 재시도할 수 있습니다. 이때 WAL/redo/undo의 문제와 idempotency의 문제가 만납니다. DB recovery는 committed row를 살리는 데 성공해도, 외부 actor에게는 unknown state가 남습니다.

이 사례에서 WAL/redo는 DB 내부 일관성을 설명합니다. data page가 disk에 내려가기 전 crash가 났더라도 log가 durable하면 recovery가 승인 상태를 다시 적용합니다. undo 또는 visibility 규칙은 commit되지 않은 변경이 남지 않게 합니다. 그러나 HTTP 응답, message publish, PG 승인 callback은 DB log의 소비자가 아닙니다. 그래서 애플리케이션은 idempotency key, outbox, 상태 조회 API, reconciliation으로 외부 세계와 다시 맞춰야 합니다.

PostgreSQL과 InnoDB 차이도 여기에 남습니다. PostgreSQL은 heap tuple visibility와 WAL replay를 통해 committed 상태를 보존하고, InnoDB는 redo와 undo를 통해 data page와 transaction 상태를 맞춥니다. MySQL binlog는 replication/PITR 쪽 소비자가 읽는 log이지 InnoDB dirty page crash recovery 자체를 대체하지 않습니다. WAL도 replication에 쓰일 수 있지만, WAL의 1차 질문은 crash 뒤 data directory를 일관되게 여는 것입니다.

PASS는 “DB commit은 됐지만 클라이언트 응답은 모른다”는 상태를 분리해 설명하고, 재시도 시 같은 idempotency key로 현재 DB 상태를 확인하는 흐름을 제시하는 것입니다. FAIL은 WAL이 있으니 외부 중복 호출도 안전하다고 말하거나, 애플리케이션 로그 성공/실패를 DB commit durable 여부와 같은 사건으로 취급하는 것입니다.

## crash recovery, checkpoint, PITR

복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이 절의 핵심 질문은 “crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/replication.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/continuous-archiving.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-recovery.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

복구 기술은 DB가 항상 깨끗하게 종료되지 않는다는 현실에서 발전했습니다. 전원 장애, process crash, storage 지연, operator 실수는 정상 commit 흐름을 중간에서 끊습니다. checkpoint와 WAL/redo는 crash 뒤 최신 committed 상태로 돌아가기 위한 내부 장치이고, continuous archiving과 PITR은 사람이 선택한 과거 시점으로 돌아가기 위한 운영 장치입니다.

### 손으로 따라가는 trace

```text
T0 base backup starts
  data files are copied while database is running
  WAL archive must keep changes during and after backup

T1 transaction A commits
T2 checkpoint completes
T3 transaction B commits
T4 mistaken DELETE commits

crash recovery after T3.5
  restore current data directory
  replay log from checkpoint to latest durable commit
  goal: latest consistent state

PITR after mistake at T4
  restore base backup from T0
  replay WAL/binlog only until target before T4
  stop before mistaken DELETE
  open restored DB on a new timeline or isolated instance
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. | InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. | crash recovery, base backup |
| 물리 모델 | PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. | InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. | checkpoint, WAL archive |
| 관측 모델 | PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. | InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. | PITR, binary log |
| 장애 모델 | PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. | InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. | base backup, recovery target time |

### 첫 번째 벽돌

첫 번째 벽돌에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 crash recovery, checkpoint, PITR 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 백업 성공과 복구 가능성을 같은 것으로 착각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

crash recovery을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- PostgreSQL restore rehearsal: base backup + restore_command + recovery_target_time 확인` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, WAL archive와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 checkpoint, PITR, base backup 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 crash recovery와 PITR을 섞어 최신 복구와 과거 시점 복구를 혼동하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

checkpoint을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_is_in_recovery();` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, binary log와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 PITR, base backup, WAL archive 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 timezone과 commit 시각 없이 target time을 고르는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

PITR을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, recovery target time와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 base backup, WAL archive, binary log 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 DB만 과거로 돌리고 외부 결제/메시지/cache/search 상태를 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

base backup을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW BINARY LOGS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, timeline와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 WAL archive, binary log, recovery target time 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 백업 성공과 복구 가능성을 같은 것으로 착각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

WAL archive을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW MASTER STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, GTID와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 binary log, recovery target time, timeline 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 crash recovery와 PITR을 섞어 최신 복구와 과거 시점 복구를 혼동하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

binary log을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- 복구 인스턴스에서 핵심 business invariant query를 실행한다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, RPO와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 recovery target time, timeline, GTID 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 timezone과 commit 시각 없이 target time을 고르는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

recovery target time을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- PostgreSQL restore rehearsal: base backup + restore_command + recovery_target_time 확인` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, RTO와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 timeline, GTID, RPO 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 DB만 과거로 돌리고 외부 결제/메시지/cache/search 상태를 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

timeline을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_is_in_recovery();` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, restore rehearsal와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 GTID, RPO, RTO 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 백업 성공과 복구 가능성을 같은 것으로 착각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

GTID을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, crash recovery와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 crash recovery와 PITR은 각각 어디까지 log를 replay해야 하며, 백업 성공만으로 복구 가능성을 말할 수 없는 이유는 무엇인가라는 질문은 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. 복구 가능성은 백업이 있다는 말로 닫히지 않습니다. 이때 RPO, RTO, restore rehearsal 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL continuous archiving은 base backup과 WAL archive를 함께 보존하고 recovery target time/name/LSN 같은 목표까지 replay하여 PITR을 수행합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB crash recovery는 redo/undo로 engine consistency를 맞추고, MySQL point-in-time restore는 backup image와 binary log position/GTID 보존 전략을 함께 요구합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 crash recovery와 PITR을 섞어 최신 복구와 과거 시점 복구를 혼동하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

RPO을 볼 때 중요한 것은 이름을 외우는 것이 아니라 base backup -> archived log stream -> target time -> replay boundary -> verified restored database 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW BINARY LOGS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, checkpoint와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- 백업 성공과 복구 가능성을 같은 것으로 착각하는 함정
- crash recovery와 PITR을 섞어 최신 복구와 과거 시점 복구를 혼동하는 함정
- timezone과 commit 시각 없이 target time을 고르는 함정
- DB만 과거로 돌리고 외부 결제/메시지/cache/search 상태를 놓치는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
-- PostgreSQL restore rehearsal: base backup + restore_command + recovery_target_time 확인
SELECT pg_is_in_recovery();
SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;
SHOW BINARY LOGS;
SHOW MASTER STATUS;
-- 복구 인스턴스에서 핵심 business invariant query를 실행한다.
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

replica를 backup으로 생각하면 위험합니다. replica는 실수 DELETE도 빠르게 따라갈 수 있으므로 과거 시점 복구를 보장하지 않습니다. base backup 파일만 있어도 부족합니다. target까지 필요한 WAL/binlog가 빠짐없이 있어야 하고, restore rehearsal로 target 도달과 business invariant가 확인되어야 합니다.

### 다시 설명해 보기

10:03에 실수 DELETE가 commit되었다고 가정하고 base backup 시각, WAL/binlog 보존 범위, target time, restore 완료 시각, business invariant 확인 query를 timeline에 표시해 보십시오. crash recovery라면 어디까지 replay하고, PITR이라면 어디서 멈추는지 구분해야 합니다.

### 통합 사례: 실수 DELETE 뒤 복구 회의에서 물어야 할 질문

운영자가 10:03:12에 잘못된 DELETE를 실행했고 10:03:15에 commit되었다면, 복구 회의의 첫 질문은 “백업 있나요?”가 아니라 “어느 시점으로 돌아가야 하며, 그 시점까지 필요한 backup과 log가 모두 있나요?”입니다. crash recovery라면 최신 durable 상태로 여는 것이 목표지만, 이 사고에서는 최신 상태가 잘못된 상태입니다. PITR은 base backup에서 시작해 실수 commit 직전까지만 WAL/binlog를 replay해야 합니다.

질문은 순서가 있습니다. base backup이 target보다 이전인지 확인합니다. target까지 필요한 WAL segment 또는 binlog가 누락 없이 있는지 확인합니다. DB server timezone, application log timezone, operator shell history, commit timestamp를 맞춰 target을 고릅니다. 별도 restore 인스턴스에서 target까지 replay하고, 삭제된 row가 살아 있으며 target 이후 정상 transaction을 얼마나 잃는지 확인합니다. 그 다음에야 cutover나 data copy-back을 논의할 수 있습니다.

복구 검증은 engine consistency와 business invariant를 나눠야 합니다. PostgreSQL recovery log가 target에 도달했고 database가 열렸다는 사실은 엔진 층 PASS입니다. 하지만 주문, 결제, 정산, 메시지 발행 상태가 가능한 조합인지 확인하는 것은 도메인 층 PASS입니다. MySQL도 backup restore와 binlog apply가 끝났다는 사실만으로 충분하지 않고, GTID/position과 핵심 table invariant를 확인해야 합니다.

PASS는 restore rehearsal 결과, target 선택 근거, 누락 log 없음, 핵심 invariant query, RTO/RPO 측정값이 남는 것입니다. FAIL은 backup job 성공 화면만 믿거나, replica를 backup처럼 취급하거나, DB만 과거로 돌린 뒤 외부 결제와 메시지 상태를 재조정하지 않는 것입니다.
