# Operations observability and troubleshooting

## DB troubleshooting by symptom

이 절은 데이터베이스 장애를 증상에서 시작해 lock, I/O, CPU, 실행 계획, 복제 지연으로 좁혀 가는 방법을 다룬다. 장애 대응에서 가장 위험한 습관은 느린 쿼리 하나를 보고 바로 인덱스를 추가하거나, CPU가 높다는 이유만으로 스펙을 올리거나, 복제 지연이 보인다는 이유만으로 reader를 늘리는 것이다. PostgreSQL 공식 문서는 데이터베이스 활동을 보려면 누적 통계뿐 아니라 `ps`, `top`, `iostat`, `vmstat` 같은 운영체제 도구와 `EXPLAIN`을 함께 봐야 한다고 설명한다. MySQL도 Performance Schema를 통해 서버 내부 실행 이벤트, wait, statement, file I/O, table lock 같은 관측 표면을 제공한다. 이 말은 DB troubleshooting이 한 화면의 숫자 맞추기가 아니라 여러 층의 사건을 같은 시간축으로 맞추는 작업이라는 뜻이다.

증상 기반 진단은 먼저 사용자가 본 현상으로 시작한다. API가 느린지, 특정 SQL만 느린지, 전체 연결이 고갈되는지, standby reader가 오래된 데이터를 반환하는지, 특정 테넌트만 실패하는지, 배치 시간에만 흔들리는지를 나눈다. 그 다음에야 DB 내부 상태를 본다. `pg_stat_activity`의 `wait_event_type`, MySQL `performance_schema.events_waits_summary_*`, slow query log, `EXPLAIN`, replication lag, OS I/O 지표가 서로 이어지면 원인 후보가 좁아진다. 반대로 이 표면들이 서로 다른 이야기를 하면 아직 원인을 모르는 상태다.

```text
증상: 결제 승인 API p95 8초
  -> 애플리케이션: DB call duration p95 7.6초
  -> DB 세션: active 12개, wait_event_type='Lock'
  -> blocker: UPDATE settlement_batch ... idle in transaction
  -> plan: 피해 쿼리 자체는 index scan, row 수 정상
  -> 조치: blocker transaction 종료/수정, 긴 트랜잭션 경계 제거

증상: 같은 API p95 8초
  -> 애플리케이션: DB call duration p95 7.6초
  -> DB 세션: wait_event_type='IO'
  -> pg_stat_io/read_time 증가 또는 OS iostat await 증가
  -> plan: seq scan으로 3천만 row 읽음
  -> 조치: 통계/조건식/인덱스/파티션 경계 검토
```

위 두 trace는 응답 시간이 같아도 원인이 완전히 다르다는 것을 보여 준다. 첫 번째는 쿼리를 빨리 만들 문제가 아니라 대기 중인 lock을 풀 문제다. 두 번째는 쿼리가 너무 많은 page를 읽는 문제다. 둘을 구분하지 못하면 첫 번째 장애에서 인덱스를 추가하고, 두 번째 장애에서 blocker를 찾느라 시간을 잃는다. 이 구분이 DU49의 핵심이다.

### 증상에서 원인 후보로 내려가는 기본 순서

| 관측 증상 | 먼저 볼 DB 표면 | 함께 볼 외부 표면 | 조심할 오판 |
|---|---|---|---|
| 특정 API만 느림 | 해당 SQL의 `EXPLAIN`, 실행 시간, wait event | 애플리케이션 span, connection pool wait | SQL이 느린지 pool 대기인지 섞는 것 |
| 전체 DB가 느림 | active session, global wait, I/O 통계 | CPU steal, disk await, network retransmit | DB 설정만 바꾸고 host 병목을 놓치는 것 |
| 간헐적 timeout | lock wait, deadlock, long transaction | 배치/스케줄러 시간표 | 평균 지표만 보고 순간 대기를 못 보는 것 |
| reader stale | replication sender/receiver, lag byte/time | 네트워크, reader CPU, apply thread | reader 증설로 apply 병목을 가리는 것 |
| 검색 또는 조회 결과가 흔들림 | plan change, stats freshness, snapshot boundary | deploy, analyze, vacuum, failover | 결과 오류와 관측 시점 차이를 혼동하는 것 |

PostgreSQL의 누적 통계는 즉시 갱신되는 값이 아니라는 점도 중요하다. 공식 문서는 누적 통계가 shared memory에 일정 간격으로 반영되고, 현재 트랜잭션 안에서는 조회한 통계가 캐시될 수 있다고 설명한다. 그래서 장애 중에 같은 세션에서 통계를 반복 조회하면 숫자가 멈춘 것처럼 보일 수 있다. 이때 `pg_stat_activity`처럼 현재 활동을 보여 주는 동적 정보와 누적 counter를 섞어 읽어야 한다. MySQL Performance Schema도 이벤트를 수집하지만, 설정된 instrument와 consumer에 따라 보이는 범위가 달라진다. 관측 표면은 진실 그 자체가 아니라, DB가 노출하도록 설정된 창이다.

### 복제 지연을 증상으로 읽는 방법

로컬 seed `database/replication_lag.md`는 writer에서 reader로 변경 로그가 전달되고 적용되는 데 시간이 걸리며, reader CPU, I/O, 네트워크, 동시성 문제가 lag를 키울 수 있다고 정리한다. 이 절에서는 그 설명을 운영 진단 흐름으로 바꾼다. 복제 지연은 하나의 숫자가 아니라 네 단계 중 어디가 밀리는지 보는 문제다.

```text
writer commit
  -> WAL/binlog 생성
  -> network 전송
  -> replica 수신
  -> replay/apply
  -> reader query가 새 상태 관측

lag 위치 후보:
  A. writer에서 로그 생성은 빠르지만 전송 대기
  B. 전송은 되지만 replica 수신/저장 대기
  C. 수신은 되지만 apply thread가 CPU/I/O/lock 때문에 늦음
  D. apply는 됐지만 애플리케이션이 오래된 reader endpoint를 봄
```

복제 지연이 보일 때 reader를 늘리는 조치는 D나 read traffic 분산에는 도움될 수 있지만, C의 apply 병목에는 도움이 되지 않을 수 있다. reader가 많아질수록 각 reader가 로그를 받아 적용해야 하므로 네트워크와 저장소 비용도 같이 본다. PASS는 lag가 어느 단계에서 증가하는지 관측 지표로 분리하고, 그 단계에 맞는 완화책을 적용하는 것이다. FAIL은 lag 숫자만 보고 reader 추가, writer 스펙업, 쿼리 튜닝 중 하나를 감으로 고르는 것이다.

### 재현 가능한 진단 루틴

```sql
-- PostgreSQL: 지금 무엇을 기다리는지 먼저 본다.
SELECT pid, state, wait_event_type, wait_event, now() - xact_start AS xact_age, query
FROM pg_stat_activity
WHERE datname = current_database()
ORDER BY xact_age DESC NULLS LAST;

-- PostgreSQL: table/index I/O와 tuple 변화를 같이 본다.
SELECT relname, seq_scan, idx_scan, n_tup_ins, n_tup_upd, n_tup_del, n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;

-- MySQL: Performance Schema/sys schema가 켜져 있다면 statement와 wait 요약을 본다.
SELECT event_name, count_star, sum_timer_wait
FROM performance_schema.events_waits_summary_global_by_event_name
ORDER BY sum_timer_wait DESC
LIMIT 10;
```

이 명령들의 PASS는 '한 쿼리가 느리다'라는 말을 적어도 하나의 내부 대기, I/O, plan, 복제 위치, 또는 애플리케이션 pool 대기로 바꾸는 것이다. FAIL은 숫자는 많이 모았지만 어떤 의사결정도 하지 못하는 상태다. 관측은 수집량이 아니라 구분력으로 평가한다.

### 이 절에서 고쳐야 하는 오해

첫째, slow query는 항상 index 문제라는 오해를 버려야 한다. 느린 쿼리는 lock을 기다릴 수도 있고, connection pool에서 시작 전부터 기다렸을 수도 있고, 네트워크 왕복에서 시간이 사라졌을 수도 있다. 둘째, 평균 지표는 장애를 숨길 수 있다. p50이 정상이어도 p99가 lock wait에 묶이면 사용자는 장애를 본다. 셋째, 복제 지연은 reader가 약해서만 생기지 않는다. writer 로그 생성, 전송, replica 수신, apply, reader routing이 모두 후보가 된다.

### 등장 배경 요약

DB troubleshooting의 등장 배경은 데이터베이스가 더 이상 단일 SQL 실행기가 아니라 애플리케이션, 운영체제, 저장소, 복제 경로와 동시에 맞물리는 공유 상태 시스템이 되었기 때문이다. 예전에는 느린 SQL 하나를 사람이 직접 보고 고치면 충분한 경우가 많았지만, 지금의 서비스에서는 connection pool, background job, reader endpoint, 배치, lock wait, query planner, storage flush가 같은 증상으로 나타난다. 그래서 이 절은 장애를 하나의 지표로 판정하지 않고, 증상에서 내부 대기와 실행 경로로 내려가는 방식으로 설명한다.

### 공식 근거와 로컬 seed

- PostgreSQL Monitoring Database Activity: https://www.postgresql.org/docs/current/monitoring.html
- PostgreSQL Cumulative Statistics System: https://www.postgresql.org/docs/current/monitoring-stats.html
- MySQL Performance Schema: https://dev.mysql.com/doc/refman/8.4/en/performance-schema.html
- local seed: `database/replication_lag.md`

위 공식 자료는 관측 표면의 존재와 한계를 고정하고, 로컬 seed는 replication lag를 CPU, I/O, network, concurrency 관점으로 풀어 둔 출발점이다. 이 절은 그 둘을 합쳐 '증상 -> 관측 표면 -> 원인 후보 -> 검증 조치'로 재구성했다.

### 증상 기반 DB 진단 replay drill 1

DB troubleshooting by symptom의 replay drill 1은 lock wait 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock wait 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 2

DB troubleshooting by symptom의 replay drill 2은 I/O saturation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 I/O saturation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 3

DB troubleshooting by symptom의 replay drill 3은 CPU-bound execution 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 CPU-bound execution 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 4

DB troubleshooting by symptom의 replay drill 4은 plan regression 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 plan regression 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 5

DB troubleshooting by symptom의 replay drill 5은 connection pool exhaustion 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 connection pool exhaustion 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 6

DB troubleshooting by symptom의 replay drill 6은 replication apply lag 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication apply lag 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 7

DB troubleshooting by symptom의 replay drill 7은 vacuum or purge pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 vacuum or purge pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 8

DB troubleshooting by symptom의 replay drill 8은 metadata or DDL blocking 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 metadata or DDL blocking 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 9

DB troubleshooting by symptom의 replay drill 9은 cache cliff 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 cache cliff 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 10

DB troubleshooting by symptom의 replay drill 10은 log or WAL pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log or WAL pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 11

DB troubleshooting by symptom의 replay drill 11은 statistics staleness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 statistics staleness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 12

DB troubleshooting by symptom의 replay drill 12은 network or TLS boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 network or TLS boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 13

DB troubleshooting by symptom의 replay drill 13은 lock wait 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock wait 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 14

DB troubleshooting by symptom의 replay drill 14은 I/O saturation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 I/O saturation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 15

DB troubleshooting by symptom의 replay drill 15은 CPU-bound execution 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 CPU-bound execution 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 16

DB troubleshooting by symptom의 replay drill 16은 plan regression 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 plan regression 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 17

DB troubleshooting by symptom의 replay drill 17은 connection pool exhaustion 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 connection pool exhaustion 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 18

DB troubleshooting by symptom의 replay drill 18은 replication apply lag 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication apply lag 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 19

DB troubleshooting by symptom의 replay drill 19은 vacuum or purge pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 vacuum or purge pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 20

DB troubleshooting by symptom의 replay drill 20은 metadata or DDL blocking 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 metadata or DDL blocking 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 21

DB troubleshooting by symptom의 replay drill 21은 cache cliff 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 cache cliff 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 22

DB troubleshooting by symptom의 replay drill 22은 log or WAL pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log or WAL pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 23

DB troubleshooting by symptom의 replay drill 23은 statistics staleness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 statistics staleness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 24

DB troubleshooting by symptom의 replay drill 24은 network or TLS boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 network or TLS boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 25

DB troubleshooting by symptom의 replay drill 25은 lock wait 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock wait 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 26

DB troubleshooting by symptom의 replay drill 26은 I/O saturation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 I/O saturation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 27

DB troubleshooting by symptom의 replay drill 27은 CPU-bound execution 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 CPU-bound execution 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 28

DB troubleshooting by symptom의 replay drill 28은 plan regression 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 plan regression 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 29

DB troubleshooting by symptom의 replay drill 29은 connection pool exhaustion 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 connection pool exhaustion 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 30

DB troubleshooting by symptom의 replay drill 30은 replication apply lag 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication apply lag 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 31

DB troubleshooting by symptom의 replay drill 31은 vacuum or purge pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 vacuum or purge pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 32

DB troubleshooting by symptom의 replay drill 32은 metadata or DDL blocking 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 metadata or DDL blocking 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 33

DB troubleshooting by symptom의 replay drill 33은 cache cliff 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 cache cliff 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 34

DB troubleshooting by symptom의 replay drill 34은 log or WAL pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log or WAL pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 증상 기반 DB 진단 replay drill 35

DB troubleshooting by symptom의 replay drill 35은 statistics staleness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 statistics staleness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

DB troubleshooting by symptom의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

## metrics, logs, locks, slow query, capacity

이 절은 metric, log, lock, slow query, capacity를 따로 외우지 않고, 각각이 어떤 내부 상태를 비추는지 연결해서 읽는 방법을 다룬다. 운영에서 흔한 실패는 CPU 그래프, slow log, lock table, replication lag, disk usage를 각각 다른 사람이 따로 보고 서로 다른 조치를 내리는 것이다. PostgreSQL의 통계 view는 현재 활동과 누적 counter를 나누어 제공하고, MySQL Performance Schema는 wait, stage, statement, transaction, file I/O, table lock 같은 server event를 표로 노출한다. 이 구조를 이해하면 지표를 '높다/낮다'로 보는 대신 '어느 내부 단계가 시간을 소비했는가'로 읽을 수 있다.

metric은 결과 숫자이고, log는 사건의 문장이고, lock은 진행을 막은 소유권이며, slow query는 특정 실행의 비용 기록이고, capacity는 이런 사건이 반복될 때 시스템이 버틸 수 있는 범위다. 다섯 가지를 한 줄로 묶으면 다음과 같다.

```text
요청 1개
  -> connection 획득
  -> SQL parse/plan/execute
  -> lock 획득 또는 대기
  -> buffer/cache hit 또는 disk read
  -> WAL/binlog flush
  -> row/result 반환
  -> metric counter 증가, log line 생성, wait sample 기록
```

같은 요청도 어디서 멈추느냐에 따라 관측값이 달라진다. lock에서 멈추면 CPU가 낮아도 latency가 높다. sort나 hash aggregate에서 멈추면 CPU와 temp file이 보인다. disk read에서 멈추면 I/O wait와 buffer miss가 보인다. commit path에서 멈추면 WAL flush나 redo log wait가 보인다. 그래서 metric은 서로를 보정해야 한다.

### 지표를 내부 상태에 연결하는 표

| 관측 표면 | 내부 질문 | PostgreSQL 예 | MySQL 예 | 오판 위험 |
|---|---|---|---|---|
| active session | 지금 실행 중인가, 기다리는가 | `pg_stat_activity.state`, `wait_event_type` | `performance_schema.threads`, processlist | active 수를 workload로만 보는 것 |
| wait event | 무엇을 기다리는가 | Lock, IO, LWLock, Client | wait/io, wait/lock, wait/synch | wait 이름만 보고 원인 확정 |
| slow query | 어떤 SQL 실행이 오래 걸렸는가 | log_min_duration_statement, EXPLAIN | slow query log, statement events | 느린 실행과 느린 시작을 혼동 |
| table/index stats | 어떤 객체가 많이 읽히고 바뀌는가 | `pg_stat_user_tables`, `pg_stat_all_indexes` | table_io_waits summary | 누적 counter를 순간 지표처럼 읽기 |
| capacity | 반복될 때 한계가 어디인가 | cache hit, bgwriter/checkpointer, WAL | buffer pool, redo, temp, connections | 평균만 보고 burst 여유 무시 |

PostgreSQL 공식 문서는 누적 통계가 실시간으로 즉시 반영되지 않고, 트랜잭션 안에서는 조회 결과가 캐시될 수 있다고 설명한다. 이 사실은 dashboard를 볼 때 매우 중요하다. 1초 전의 counter와 지금의 active session을 같은 정확도로 비교하면 안 된다. 반대로 MySQL Performance Schema는 현재/과거/요약 event를 제공하지만, instrument와 consumer 설정에 따라 수집되는 것이 달라진다. 그래서 '보이지 않는다'는 말은 '없다'가 아니라 '수집하지 않았거나 다른 표면에서 봐야 한다'일 수 있다.

### 느린 쿼리 하나를 capacity 언어로 바꾸기

```text
slow log:
  query_time=4.8s rows_examined=15,000,000 rows_sent=20

해석 1: rows_examined가 rows_sent보다 매우 크다.
  -> 후보: 조건식이 index를 못 탐, 통계가 틀림, leading column 불일치, 함수/형 변환

해석 2: query_time은 큰데 rows_examined는 작다.
  -> 후보: lock wait, network send, client fetch, disk flush, metadata lock

해석 3: 특정 시간대에만 느리다.
  -> 후보: 배치, checkpoint, backup, autovacuum/purge, noisy neighbor, cold cache
```

이 trace에서 중요한 것은 slow query log만으로 결론을 내리지 않는 것이다. slow query는 사건 표지판이다. 원인은 plan, wait, object stats, host metrics, 배치 시간표와 맞춰야 한다. capacity 판단도 마찬가지다. 'CPU 70%'는 여유가 있다는 뜻일 수도 있고, single-thread 병목 때문에 한 core가 꽉 찼다는 뜻일 수도 있다. 'cache hit 99%'는 좋아 보일 수 있지만, 1% miss가 초당 수천 번의 random read라면 tail latency를 만들 수 있다.

### 관측 루틴을 문장으로 고정하기

운영 runbook은 명령 목록보다 판단 문장으로 시작해야 한다. `SELECT * FROM pg_stat_activity`를 실행한다는 문장은 충분하지 않다. 더 좋은 문장은 '현재 느린 요청이 실행 중인지, lock이나 I/O를 기다리는지, 또는 client fetch에서 멈췄는지 확인한다'이다. MySQL에서도 `performance_schema`를 본다는 말보다 'statement digest, wait event, table I/O, lock wait 중 어떤 event family가 시간을 먹는지 본다'가 더 강하다.

```text
metric -> question -> command -> interpretation -> action
CPU 95% -> 계산이 병목인가? -> plan actual time + host run queue -> CPU-bound면 쿼리/연산량 축소
Lock wait -> 누가 소유자인가? -> blocking session graph -> blocker transaction 처리
I/O wait -> 어떤 object를 읽는가? -> table/index I/O + plan buffers -> scan/index/cache 조정
Lag -> 어느 복제 단계인가? -> send/receive/apply 위치 -> network/apply/query 분리
```

이 흐름을 따르면 dashboard가 많아져도 판단이 흐려지지 않는다. 각 panel은 어떤 내부 질문에 답하는지 이름이 있어야 하고, 그 질문에 답하지 못하면 panel 수를 줄이거나 query를 바꿔야 한다.

### 검증과 replay

이 절의 검증은 실제 운영 DB를 건드리지 않고도 할 수 있다. synthetic schema에서 느린 scan, lock wait, temp spill을 일부러 만들고, 각 상황에서 어떤 지표가 바뀌는지 기록한다. PASS는 같은 SQL 지연을 lock-bound, I/O-bound, CPU-bound로 구분할 수 있는 것이다. FAIL은 모든 지연을 slow query 하나로 묶거나, metric이 올랐다는 사실만 남기고 내부 상태를 설명하지 못하는 것이다.

### 등장 배경 요약

관측 체계가 중요해진 배경은 DB 장애가 단일 로그 한 줄로 설명되지 않는 경우가 늘었기 때문이다. metric은 시간에 따른 숫자, log는 확정된 사건, lock은 진행을 막은 소유권, slow query는 특정 실행의 비용, capacity는 반복 부하를 견디는 여유를 말한다. 이 다섯 표면은 따로 등장한 도구가 아니라 같은 실행 경로를 다른 각도에서 보여 주는 장치이므로 함께 읽어야 한다.

### 공식 근거와 로컬 seed

- PostgreSQL Cumulative Statistics System: https://www.postgresql.org/docs/current/monitoring-stats.html
- MySQL Performance Schema: https://dev.mysql.com/doc/refman/8.4/en/performance-schema.html
- local seed: `database/database-deep-study-plan.md`, `database/replication_lag.md`

공식 자료는 관측값이 어떻게 수집되고 어떤 제약을 갖는지 알려 준다. 이 절은 그 제약을 운영 판단으로 번역해, metric과 log가 내부 상태의 어느 조각인지 확인하도록 구성했다.

### DB 관측 표면 연결 replay drill 1

metrics, logs, locks, slow query, capacity의 replay drill 1은 counter와 gauge 구분 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 counter와 gauge 구분 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 2

metrics, logs, locks, slow query, capacity의 replay drill 2은 wait와 CPU 결합 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 wait와 CPU 결합 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 3

metrics, logs, locks, slow query, capacity의 replay drill 3은 slow log 샘플링 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 slow log 샘플링 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 4

metrics, logs, locks, slow query, capacity의 replay drill 4은 lock graph 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock graph 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 5

metrics, logs, locks, slow query, capacity의 replay drill 5은 log correlation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log correlation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 6

metrics, logs, locks, slow query, capacity의 replay drill 6은 capacity headroom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 capacity headroom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 7

metrics, logs, locks, slow query, capacity의 replay drill 7은 checkpoint pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 checkpoint pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 8

metrics, logs, locks, slow query, capacity의 replay drill 8은 temp file spill 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 temp file spill 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 9

metrics, logs, locks, slow query, capacity의 replay drill 9은 index usefulness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index usefulness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 10

metrics, logs, locks, slow query, capacity의 replay drill 10은 replication metric 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication metric 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 11

metrics, logs, locks, slow query, capacity의 replay drill 11은 error log as symptom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 error log as symptom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 12

metrics, logs, locks, slow query, capacity의 replay drill 12은 observability cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 13

metrics, logs, locks, slow query, capacity의 replay drill 13은 counter와 gauge 구분 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 counter와 gauge 구분 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 14

metrics, logs, locks, slow query, capacity의 replay drill 14은 wait와 CPU 결합 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 wait와 CPU 결합 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 15

metrics, logs, locks, slow query, capacity의 replay drill 15은 slow log 샘플링 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 slow log 샘플링 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 16

metrics, logs, locks, slow query, capacity의 replay drill 16은 lock graph 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock graph 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 17

metrics, logs, locks, slow query, capacity의 replay drill 17은 log correlation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log correlation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 18

metrics, logs, locks, slow query, capacity의 replay drill 18은 capacity headroom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 capacity headroom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 19

metrics, logs, locks, slow query, capacity의 replay drill 19은 checkpoint pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 checkpoint pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 20

metrics, logs, locks, slow query, capacity의 replay drill 20은 temp file spill 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 temp file spill 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 21

metrics, logs, locks, slow query, capacity의 replay drill 21은 index usefulness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index usefulness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 22

metrics, logs, locks, slow query, capacity의 replay drill 22은 replication metric 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication metric 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 23

metrics, logs, locks, slow query, capacity의 replay drill 23은 error log as symptom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 error log as symptom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 24

metrics, logs, locks, slow query, capacity의 replay drill 24은 observability cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 25

metrics, logs, locks, slow query, capacity의 replay drill 25은 counter와 gauge 구분 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 counter와 gauge 구분 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 26

metrics, logs, locks, slow query, capacity의 replay drill 26은 wait와 CPU 결합 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 wait와 CPU 결합 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 27

metrics, logs, locks, slow query, capacity의 replay drill 27은 slow log 샘플링 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 slow log 샘플링 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 28

metrics, logs, locks, slow query, capacity의 replay drill 28은 lock graph 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 lock graph 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 29

metrics, logs, locks, slow query, capacity의 replay drill 29은 log correlation 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 log correlation 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 30

metrics, logs, locks, slow query, capacity의 replay drill 30은 capacity headroom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 capacity headroom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 31

metrics, logs, locks, slow query, capacity의 replay drill 31은 checkpoint pressure 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 checkpoint pressure 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 32

metrics, logs, locks, slow query, capacity의 replay drill 32은 temp file spill 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 temp file spill 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 33

metrics, logs, locks, slow query, capacity의 replay drill 33은 index usefulness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index usefulness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 34

metrics, logs, locks, slow query, capacity의 replay drill 34은 replication metric 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 replication metric 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 35

metrics, logs, locks, slow query, capacity의 replay drill 35은 error log as symptom 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 error log as symptom 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 관측 표면 연결 replay drill 36

metrics, logs, locks, slow query, capacity의 replay drill 36은 observability cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 36의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

metrics, logs, locks, slow query, capacity의 replay drill 36에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.
