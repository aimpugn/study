# Database Operations, Security, Troubleshooting Deep Dive

운영에서 데이터베이스 문제는 대개 한 단어로 오지 않습니다. "느립니다", "가끔 timeout이 납니다", "replica에서 데이터가 안 보입니다", "권한이 갑자기 안 됩니다", "누가 데이터를 봤는지 모르겠습니다"처럼 증상은 흐리고 원인은 여러 층에 걸칩니다. 면접에서 좋은 답은 바로 인덱스를 추가하거나 권한을 열어 주는 것이 아니라, 어떤 관측값으로 어느 가능성을 줄이고, 어떤 조치가 되돌릴 수 있으며, 어떤 조치가 더 큰 장애를 만들 수 있는지 설명하는 것입니다.

이 문서는 slow query, lock, replication lag, capacity, observability, roles/grants/ownership/RLS/encryption/auditing/secrets를 하나의 운영 흐름으로 묶습니다. 핵심은 "DB가 느리다"라는 말에서 CPU, I/O, lock wait, plan regression, pool starvation, replication apply delay, checkpoint pressure, permission boundary, secret rotation까지 가능한 원인들을 차례대로 좁히는 능력입니다.

민감한 로컬 인증 파일이나 개인 환경 파일은 본문 근거로 직접 인용하지 않습니다. 그런 파일은 "민감정보가 있을 수 있으므로 원문을 열어 옮기지 않는 source boundary"로만 다룹니다. 운영·보안 문서에서 필요한 것은 실제 secret 값이 아니라, secret이 어디에 저장되면 위험한지, 어떻게 회전해야 하는지, 감사 로그와 권한 경계를 어떻게 남겨야 하는지입니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [Slow query: 느린 SQL 하나가 아니라 실행 경로를 본다](#slow-query-느린-sql-하나가-아니라-실행-경로를-본다)
    - [Lock troubleshooting: blocker를 찾고 transaction boundary를 고친다](#lock-troubleshooting-blocker를-찾고-transaction-boundary를-고친다)
    - [Replication lag: 전송 지연과 적용 지연을 분리한다](#replication-lag-전송-지연과-적용-지연을-분리한다)
    - [Capacity: 병목은 자원 사용률보다 queue에서 드러난다](#capacity-병목은-자원-사용률보다-queue에서-드러난다)
    - [Observability: query, wait, resource, business signal을 연결한다](#observability-query-wait-resource-business-signal을-연결한다)
    - [OS 지표는 DB wait event의 바깥쪽 원인을 보여 준다](#os-지표는-db-wait-event의-바깥쪽-원인을-보여-준다)
    - [Roles, grants, ownership: 권한은 계층으로 설계한다](#roles-grants-ownership-권한은-계층으로-설계한다)
    - [Row-Level Security와 tenant boundary](#row-level-security와-tenant-boundary)
    - [Encryption, auditing, secrets: 암호화는 저장 위치와 키 관리까지 포함한다](#encryption-auditing-secrets-암호화는-저장-위치와-키-관리까지-포함한다)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
    - [PostgreSQL blocker 찾기](#postgresql-blocker-찾기)
    - [PostgreSQL plan estimate와 actual 비교](#postgresql-plan-estimate와-actual-비교)
    - [MySQL slow query와 InnoDB 상태 보기](#mysql-slow-query와-innodb-상태-보기)
    - [권한 최소화 실험](#권한-최소화-실험)
    - [RLS tenant leak 확인](#rls-tenant-leak-확인)
    - [replication lag 정책 점검](#replication-lag-정책-점검)
- [면접 꼬리 질문](#면접-꼬리-질문)
    - [느린 쿼리가 들어오면 제일 먼저 무엇을 보나요?](#느린-쿼리가-들어오면-제일-먼저-무엇을-보나요)
    - [lock wait timeout이 많으면 timeout 값을 늘리면 되나요?](#lock-wait-timeout이-많으면-timeout-값을-늘리면-되나요)
    - [replication lag가 있을 때 애플리케이션은 무엇을 해야 하나요?](#replication-lag가-있을-때-애플리케이션은-무엇을-해야-하나요)
    - [권한 문제를 빨리 해결하려고 superuser를 주면 왜 위험한가요?](#권한-문제를-빨리-해결하려고-superuser를-주면-왜-위험한가요)
    - [RLS를 켜면 tenant isolation이 완전히 보장되나요?](#rls를-켜면-tenant-isolation이-완전히-보장되나요)
    - [감사 로그에는 무엇을 남겨야 하나요?](#감사-로그에는-무엇을-남겨야-하나요)
- [함정 질문](#함정-질문)
    - ["쿼리가 느리면 인덱스를 추가하면 되죠?"](#쿼리가-느리면-인덱스를-추가하면-되죠)
    - ["connection pool이 부족하니 max pool size를 올리면 되죠?"](#connection-pool이-부족하니-max-pool-size를-올리면-되죠)
    - ["replica lag는 DBA가 해결할 문제라 application은 신경 안 써도 되죠?"](#replica-lag는-dba가-해결할-문제라-application은-신경-안-써도-되죠)
    - ["암호화 at rest를 켰으니 데이터 유출 걱정은 끝인가요?"](#암호화-at-rest를-켰으니-데이터-유출-걱정은-끝인가요)
    - ["권한 오류가 나면 GRANT ALL로 열고 나중에 줄이면 되죠?"](#권한-오류가-나면-grant-all로-열고-나중에-줄이면-되죠)
    - ["민감 파일을 한 번만 보고 문서에 참고하면 괜찮지 않나요?"](#민감-파일을-한-번만-보고-문서에-참고하면-괜찮지-않나요)
    - [마지막 연결: 증상에서 원인으로 좁혀 가는 관측 경로](#마지막-연결-증상에서-원인으로-좁혀-가는-관측-경로)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

느린 쿼리를 받으면 첫 반응은 "인덱스가 없나?"일 수 있습니다. 하지만 운영에서는 그보다 먼저 증상을 분류해야 합니다. 한 query만 느린가, 전체 DB가 느린가, 특정 시간대만 느린가, write가 몰릴 때 느린가, replica에서만 느린가, application pool에서 기다리는가, DB 안에서 lock을 기다리는가, disk read가 늘었는가, planner가 plan을 바꿨는가를 나누어야 합니다. 같은 5초 latency라도 원인이 index miss, lock wait, connection pool exhaustion, checkpoint write burst, replication replay lag이면 조치가 완전히 다릅니다.

Slow query troubleshooting의 기본 순서는 query text를 보고 바로 고치는 것이 아니라, 실제 실행 계획과 관측 지표를 맞추는 것입니다. PostgreSQL에서는 `EXPLAIN (ANALYZE, BUFFERS)`, `pg_stat_activity`, `pg_stat_statements`, wait_event, locks, table/index statistics를 봅니다. MySQL에서는 slow query log, `EXPLAIN`/`EXPLAIN ANALYZE`, Performance Schema, `SHOW PROCESSLIST`, InnoDB status를 봅니다. query plan은 추정이고, 실행 시간은 결과입니다. 추정 rows와 실제 rows가 크게 다르면 통계나 selectivity 문제를 의심하고, buffers/read가 많으면 I/O와 cache를 봅니다. 실행 계획을 읽는 기본기는 [인덱스와 optimizer](04-index-query-optimizer.md)와 이어집니다.

Lock 문제는 "누가 막고 누가 기다리는가"를 그리는 일이 먼저입니다. blocker가 하나인지, chain인지, idle in transaction인지, DDL인지, application transaction이 외부 API 호출을 끼고 오래 열렸는지 봐야 합니다. lock wait timeout을 늘리는 것은 증상을 늦출 뿐일 수 있습니다. 올바른 조치는 transaction 길이 줄이기, access order 통일, index로 lock 범위 줄이기, DDL scheduling, retry 정책, 빠른 blocker kill 여부 판단처럼 원인별로 달라집니다. lock 자체의 원리는 [격리 수준, lock, deadlock](08-isolation-lock-deadlock.md), 애플리케이션이 lock을 오래 잡는 경계는 [애플리케이션 경계, 멱등성, 돈, outbox](12-application-boundaries-idempotency-money-outbox.md)와 함께 봐야 합니다.

Replication lag는 "복제 지연 10초"라는 숫자 하나로 끝나지 않습니다. primary에서 WAL/binlog가 생성되고, replica로 전송되고, 쓰이고, flush되고, replay/apply되는 단계가 있습니다. lag가 network 전송 지연인지, replica I/O인지, apply thread가 single-thread bottleneck인지, long query가 recovery conflict를 만드는지 분리해야 합니다. 애플리케이션은 replica lag가 있을 때 read-your-writes를 보장할지, primary read로 fallback할지, stale read를 허용할지 정책을 가져야 합니다. 복제 단계와 장애 전환 판단은 [복제, 지연, 백업, failover](09-replication-lag-backup-failover.md)를 같이 보면 더 정확합니다.

보안은 GRANT 문 몇 개가 아니라 ownership, role hierarchy, row-level security, encryption, auditing, secret handling의 합입니다. application runtime role과 migration role, read-only analytics role, DBA role을 분리하고, 객체 owner가 runtime role이 되지 않게 해야 합니다. RLS는 테이블 안 row 단위 접근을 제한하지만 우회 권한과 정책 누락을 조심해야 합니다. encryption은 저장 시 암호화, 전송 구간 TLS, application-level encryption, key management를 나눠 봐야 합니다. audit log는 누가 언제 무엇을 봤고 바꿨는지 재구성할 수 있어야 하지만, 민감값 자체를 로그로 남기면 안 됩니다.

## 먼저 잡아야 할 작은 모델

운영 사고를 작은 모델로 줄이면 다음과 같습니다.

```text
user request
  -> application thread
  -> connection pool 대기 또는 borrow
  -> DB session
  -> parse/plan
  -> lock wait
  -> CPU execution
  -> buffer cache / disk I/O
  -> WAL/binlog write
  -> commit response
  -> replica send/write/flush/replay
```

latency는 이 경로의 어느 지점에서도 생길 수 있습니다. 사용자가 "DB가 느리다"고 말해도 실제로는 application이 pool에서 connection을 기다렸을 수 있습니다. DB query가 실행되기도 전에 queueing이 생긴 것입니다. 반대로 DB 안에서는 query가 CPU를 쓰는 중일 수도 있고, lock을 기다리는 중일 수도 있고, disk page read를 기다리는 중일 수도 있습니다. 운영자는 먼저 대기 위치를 찾아야 합니다.

작은 진단 표를 머리에 둡니다.

| 증상 | 먼저 볼 관측값 | 흔한 원인 후보 | 바로 하면 위험한 조치 |
|---|---|---|---|
| 특정 query만 느림 | 실행 계획, actual rows, buffers, slow log | 인덱스 누락, 통계 오류, plan regression, parameter skew | 인덱스만 무작정 추가 |
| 전체 DB가 느림 | CPU, I/O, active session, wait event, pool metrics | 부하 증가, lock storm, checkpoint, connection 폭증 | max connection만 증가 |
| write가 느림 | WAL/binlog fsync, dirty page, lock, replication sync | disk latency, checkpoint, hot row, synchronous replication | durability 설정 무작정 완화 |
| replica가 늦음 | send/write/flush/replay lag, apply queue | network, replica disk, long query, single apply bottleneck | replica read 계속 허용 |
| 권한 오류 | current role, object owner, grants, default privileges | migration role/owner 혼선, schema usage 누락 | broad superuser grant |
| 데이터 노출 우려 | audit log, role mapping, RLS policy, query path | application role 과권한, policy 누락, secret 로그 | 로그를 더 많이 찍으며 민감값 노출 |

Slow query의 작은 모델은 "쿼리 문장 -> 계획 -> 실행 -> 대기"입니다. SQL text만 보면 조건과 join이 보입니다. 계획을 보면 DBMS가 어떤 access path를 선택했는지 보입니다. 실행 통계를 보면 그 선택이 실제 데이터 분포와 맞았는지 보입니다. wait 지표를 보면 DB가 계산 중인지 기다리는 중인지 보입니다.

Lock의 작은 모델은 "blocker, waiter, resource"입니다. 누가 어떤 row/table/index/gap/transaction id를 잡고 있고, 누가 그것을 기다리는지, blocker는 왜 아직 commit/rollback하지 않았는지 찾습니다. blocker가 idle in transaction이면 application이 transaction을 열어 둔 채 멈춘 것입니다. blocker가 오래 실행 중인 DDL이면 배포 절차 문제일 수 있습니다. blocker가 hot row update라면 데이터 모델과 access pattern 문제일 수 있습니다.

Capacity의 작은 모델은 "평균이 아니라 tail과 saturation"입니다. CPU 50% 평균이어도 single hot query나 I/O queue가 tail latency를 만들 수 있습니다. connection 수가 넉넉해 보여도 active transaction이 늘면 lock graph가 복잡해질 수 있습니다. disk 사용률이 70%라도 WAL burst나 temp file spill이 생기면 순간적으로 장애가 납니다. 운영에서는 평균값 하나보다 queue length, p95/p99 latency, active session, wait breakdown, growth rate를 봅니다.

Security의 작은 모델은 "identity -> permission -> object -> row -> audit"입니다. 누가 접속했는가, 그 role이 어떤 schema/table/function에 어떤 권한을 갖는가, 객체 owner는 누구인가, row-level policy가 어떤 조건을 적용하는가, 접근 기록이 남는가를 순서대로 봅니다. secret은 이 흐름의 입구를 여는 열쇠입니다. secret 값은 코드, 문서, 로그, ticket, screenshot에 남기지 않고, secret manager와 rotation 절차로 관리해야 합니다.

## 깊은 메커니즘

### Slow query: 느린 SQL 하나가 아니라 실행 경로를 본다

Slow query log는 출발점이지 결론이 아닙니다. MySQL slow query log는 오래 걸린 query를 기록하고, PostgreSQL은 `log_min_duration_statement`나 `pg_stat_statements`로 느린 statement를 찾을 수 있습니다. 하지만 느린 query text만 보면 왜 느린지 알 수 없습니다. 같은 SQL이라도 parameter 값, 데이터 분포, cache 상태, lock wait, plan 선택, concurrent workload에 따라 시간이 달라집니다.

실행 계획은 estimated cost와 row estimate를 보여 줍니다. PostgreSQL의 `EXPLAIN (ANALYZE, BUFFERS)`는 실제 실행 시간, 실제 row 수, loop count, buffer hit/read/dirtied를 보여 줍니다. MySQL 8.0 이후에도 `EXPLAIN ANALYZE`로 runtime 정보를 볼 수 있습니다. 중요한 질문은 "인덱스를 탔나?"보다 "DBMS가 몇 row를 예상했고 실제로 몇 row를 읽었는가"입니다. 예상이 크게 틀리면 통계가 오래됐거나, column correlation을 planner가 몰랐거나, parameter distribution이 skewed할 수 있습니다.

인덱스 추가는 비용이 있는 조치입니다. read query는 빨라질 수 있지만 write마다 index maintenance 비용이 늘고, storage와 buffer cache를 더 씁니다. composite index는 column order와 predicate 형태가 맞아야 합니다. PostgreSQL에서는 partial index, expression index, extended statistics가 유용할 수 있고, MySQL에서는 covering index와 clustered lookup 비용, prefix index, collation, generated column index를 고려할 수 있습니다. 운영에서 인덱스는 "쿼리 하나를 빠르게 하는 도구"이면서 "write path와 cache footprint를 바꾸는 schema change"입니다.

Query rewrite도 단순 문법 취향이 아닙니다. `OR` 조건이 index 사용을 막는지, leading wildcard LIKE가 search index를 못 타게 하는지, implicit cast가 index를 무력화하는지, function wrapping이 predicate pushdown을 막는지, pagination offset이 뒤 페이지에서 점점 느려지는지 봐야 합니다. 특히 `ORDER BY ... LIMIT`는 index order와 잘 맞으면 매우 빠르지만, 정렬 대상이 크면 temp file이나 filesort가 생깁니다.

### Lock troubleshooting: blocker를 찾고 transaction boundary를 고친다

Lock wait가 보이면 먼저 blocker/waiter graph를 만듭니다. PostgreSQL에서는 `pg_stat_activity`와 `pg_locks`, `pg_blocking_pids()`를 조합합니다. MySQL에서는 Performance Schema data lock tables, metadata lock tables, processlist, InnoDB status를 봅니다. "어떤 query가 느리다"가 아니라 "어떤 session이 어떤 resource를 기다리고 있고, 그 resource를 누가 잡고 있는가"로 바꿔야 합니다.

긴 transaction은 두 가지 장애를 만듭니다. 하나는 lock을 오래 잡는 장애입니다. UPDATE 후 외부 API를 호출하고, 응답을 기다리는 동안 row lock이 유지되면 다른 transaction이 줄줄이 대기합니다. 다른 하나는 MVCC 정리 지연입니다. PostgreSQL에서는 오래 열린 snapshot이 vacuum을 막고, InnoDB에서는 undo purge를 늦출 수 있습니다. transaction이 idle in transaction으로 오래 남아 있으면 "실행 중인 query가 없는데도 장애를 만드는" 상태가 됩니다.

Deadlock은 서로 다른 transaction이 상대가 가진 lock을 기다리는 순환입니다. DBMS는 보통 deadlock을 감지해 하나를 abort합니다. 애플리케이션은 deadlock error를 재시도 가능한 오류로 분류해야 합니다. 하지만 재시도만 넣으면 충분하지 않습니다. update 순서를 통일하고, hot row를 줄이고, 범위 조건에 적절한 index를 만들고, transaction 안에서 처리하는 일을 줄여야 합니다.

DDL lock도 중요합니다. 운영 중 `ALTER TABLE`은 DBMS와 작업 종류에 따라 강한 lock을 요구하거나 metadata lock을 잡을 수 있습니다. MySQL에서는 metadata lock 때문에 짧은 select가 DDL을 막고, DDL 대기가 뒤의 query를 다시 막는 chain이 생길 수 있습니다. PostgreSQL에서도 table rewrite가 필요한 DDL이나 강한 lock mode가 필요한 작업은 배포 시점을 조정해야 합니다. "DDL은 schema 변경이니 애플리케이션 트래픽과 별개"가 아닙니다.

### Replication lag: 전송 지연과 적용 지연을 분리한다

Replication lag는 primary와 replica 사이의 시간 차이지만, 내부 단계는 여러 개입니다. PostgreSQL physical replication에서는 WAL이 primary에서 생성되고, standby로 전송되고, standby에서 write/flush/replay됩니다. MySQL replication에서는 binary log가 source에서 생성되고, replica I/O thread가 받아 relay log에 쓰고, SQL/applier thread가 적용합니다. multi-threaded replication을 쓰더라도 dependency와 commit order 때문에 병목이 남을 수 있습니다.

Lag 원인은 다양합니다. network가 느려 WAL/binlog 전송이 늦을 수 있습니다. replica disk가 느려 write/flush가 늦을 수 있습니다. replica에서 long query가 recovery conflict를 만들 수 있습니다. primary에서 대량 update나 index build가 발생해 log volume이 폭증할 수 있습니다. replica CPU가 부족해 apply가 늦을 수 있습니다. 따라서 lag 숫자 하나를 보고 replica를 재시작하는 식의 조치는 위험합니다.

애플리케이션 정책도 필요합니다. 사용자가 글을 쓴 직후 read replica로 조회하면 아직 안 보일 수 있습니다. 이때 primary read로 일정 시간 fallback할지, session consistency token을 둘지, stale read를 허용하는 화면과 허용하지 않는 화면을 나눌지 정해야 합니다. 금융 잔액, 주문 결제 상태, 권한 변경 직후 확인 같은 경로는 stale read를 허용하기 어렵습니다. 검색 화면이나 통계 화면은 일부 지연을 허용할 수 있습니다.

### Capacity: 병목은 자원 사용률보다 queue에서 드러난다

Capacity planning은 CPU, memory, disk, network 숫자를 크게 잡는 일이 아닙니다. 어떤 workload가 어느 자원을 얼마나 오래 점유하고, queue가 어디에 생기는지 보는 일입니다. DB에서는 active connection 수, transaction duration, lock wait, disk I/O latency, WAL/binlog write rate, temp file usage, buffer cache hit, checkpoint activity, table/index growth rate가 중요합니다.

Connection 수는 특히 조심해야 합니다. DB server는 수천 connection을 받을 수 있어도 active query를 동시에 효율적으로 처리할 수 있는 수는 훨씬 작을 수 있습니다. 애플리케이션 instance가 늘면서 각 instance의 pool max를 그대로 두면 전체 DB connection limit을 빠르게 소진합니다. 더 많은 connection은 더 많은 throughput이 아니라 더 긴 queue와 더 복잡한 lock contention을 만들 수 있습니다.

Disk capacity는 사용률만 보면 늦습니다. table growth, index growth, WAL/binlog retention, replication slot, backup retention, temp file, audit log를 함께 봅니다. PostgreSQL replication slot이 오래된 WAL 제거를 막거나, MySQL binary log retention이 예상보다 커지면 disk full이 될 수 있습니다. Disk full은 단순 저장 공간 문제가 아니라 commit failure와 crash recovery risk로 이어집니다.

Checkpoint와 background write도 tail latency에 영향을 줍니다. Dirty page가 많이 쌓였다가 한꺼번에 flush되면 query latency가 튈 수 있습니다. PostgreSQL에서는 checkpoint settings와 WAL volume, bgwriter/checkpointer 통계를 보고, MySQL InnoDB에서는 dirty page, redo log capacity, flush activity를 봅니다. Write-heavy workload에서는 p95/p99 latency가 평균보다 먼저 나빠집니다.

### Observability: query, wait, resource, business signal을 연결한다

DB observability는 CPU dashboard 하나로 충분하지 않습니다. 최소한 네 층이 필요합니다.

1. Query 층
느린 query fingerprint, 호출 횟수, total time, mean/p95 time, rows examined/returned, plan 변화가 필요합니다. PostgreSQL의 `pg_stat_statements`, MySQL Performance Schema와 slow log가 여기에 해당합니다.

2. Wait 층
DB session이 CPU를 쓰는지, lock을 기다리는지, I/O를 기다리는지, client read/write를 기다리는지 봅니다. PostgreSQL wait_event와 MySQL Performance Schema wait instrumentation이 여기에 해당합니다.

3. Resource 층
CPU, memory, disk latency, IOPS, network, WAL/binlog volume, connection count, buffer cache, temp file 사용량을 봅니다.

4. Business 층
결제 성공률, 주문 생성 latency, 로그인 실패율, reconciliation backlog, outbox lag 같은 업무 지표를 봅니다. DB 지표가 정상이더라도 business 지표가 나쁘면 사용자 장애입니다.

### OS 지표는 DB wait event의 바깥쪽 원인을 보여 준다

DB wait event는 데이터베이스 내부에서 세션이 무엇을 기다리는지 알려 주지만, 그 기다림의 바깥쪽 원인은 운영체제와 장치 계층에 있을 수 있습니다. 예를 들어 PostgreSQL에서 WAL sync 시간이 늘거나 MySQL에서 redo flush가 느려지면 DBMS 내부에서는 commit 대기처럼 보입니다. 하지만 실제 병목은 커널 page cache의 dirty writeback, filesystem journal, block layer queue, device driver, storage controller cache flush, 또는 같은 서버에서 실행 중인 다른 process의 I/O 압박일 수 있습니다.

```text
DB symptom
  commit latency / checkpoint spike / replica flush lag
        |
        v
DB internal signal
  WAL sync time, fsync count, dirty buffers, wait_event
        |
        v
OS signal
  iostat await, vmstat dirty/writeback, pidstat -d,
  CPU run queue, context switch, page reclaim, disk queue depth
        |
        v
hardware / virtualization signal
  controller cache flush, cloud block volume latency, NVMe queue, RAID battery state
```

이 구분을 놓치면 위험한 조치를 하게 됩니다. `fsync` 지연이 보인다고 durability 설정을 바로 낮추면 장애는 줄어들 수 있지만 crash safety를 잃을 수 있습니다. Connection pool 대기가 보인다고 pool size를 키우면 DB process는 더 많은 runnable task와 lock contention을 떠안을 수 있습니다. Replica lag가 보인다고 replica만 재시작하면 실제 원인인 primary WAL 폭증이나 replica disk flush 병목은 그대로 남습니다. 면접 답변에서는 DB 지표를 먼저 잡고, 그 지표가 OS의 CPU scheduling, memory reclaim, block I/O queue, storage flush 중 어디와 만나는지까지 한 단계 더 내려가야 합니다.

로그는 재구성 가능해야 하지만 민감값을 담으면 안 됩니다. Query parameter 전체를 그대로 남기면 주민번호, 카드번호, access token 같은 값이 노출될 수 있습니다. 좋은 로그는 request id, transaction id, query fingerprint, duration, rows, error code, sanitized parameters, application command id를 남깁니다. 민감값은 masking 또는 redaction하고, 필요할 때도 별도 보안 절차를 거쳐 제한적으로 조회합니다.

### Roles, grants, ownership: 권한은 계층으로 설계한다

DB 권한 설계의 첫 단계는 identity 분리입니다. application runtime role, migration role, read-only reporting role, batch role, DBA role을 나눕니다. Runtime role이 schema owner이거나 superuser면 애플리케이션 취약점 하나가 전체 schema 변경과 데이터 파괴로 이어질 수 있습니다. Migration role은 DDL 권한을 갖지만 application request path에서는 쓰지 않는 편이 안전합니다.

Ownership과 privileges는 다릅니다. 객체 owner는 그 객체를 변경할 수 있는 강한 권한을 갖고, privileges는 다른 role에게 부여된 사용 권한입니다. PostgreSQL에서는 schema usage와 table privileges, sequence privileges가 따로 필요할 수 있습니다. MySQL에서는 global, database, table, column, routine 권한이 계층적으로 있습니다. 권한 오류를 broad grant로 덮으면 당장은 해결돼도 최소 권한 원칙이 깨집니다.

Default privileges도 놓치기 쉽습니다. 새 table을 만들 때 runtime role에 권한이 자동으로 부여되지 않으면 배포 후 특정 table만 접근 실패가 납니다. 반대로 default로 과도한 권한을 주면 새 object가 자동으로 노출됩니다. 권한은 migration과 함께 versioned artifact로 관리하고, production에서 수동 grant를 반복하지 않는 편이 좋습니다.

### Row-Level Security와 tenant boundary

RLS는 row 단위 접근 정책을 DBMS가 적용하게 하는 기능입니다. PostgreSQL RLS가 대표적입니다. multi-tenant table에서 `tenant_id = current_setting('app.tenant_id', true)`처럼 요청 또는 transaction에 설정한 tenant context를 정책에 사용하면 application bug가 WHERE tenant condition을 빠뜨려도 DB가 row를 걸러 줄 수 있습니다. `true` 인자는 설정값이 없을 때 오류 대신 null을 돌려주므로, 정책을 "tenant context가 없으면 보이지 않음" 쪽으로 설계하기 쉽습니다. 하지만 RLS는 만능이 아닙니다. table owner나 bypass 권한을 가진 role, security definer function, connection session setting 오염, policy가 적용되지 않는 maintenance path를 조심해야 합니다.

Tenant boundary를 connection session state에 의존하면 pool reset이 더 중요해집니다. 요청 A가 tenant id를 session variable로 설정하고, 반납 시 clear하지 않으면 요청 B가 잘못된 tenant context를 물려받을 수 있습니다. 그래서 RLS를 쓸 때는 tenant setting을 transaction-local로 두거나, request마다 확실히 set/reset하고, 테스트에서 cross-tenant leak을 검증해야 합니다.

Firestore 같은 document DB의 security rules도 RLS와 비슷한 역할을 하지만 DBMS SQL 권한과는 다릅니다. Client SDK가 직접 DB에 접근하는 모델에서는 security rules가 application authorization의 중요한 일부가 됩니다. Admin SDK는 rules를 우회할 수 있으므로 server-side path의 권한 검사는 별도로 필요합니다. "DB 보안"은 제품별 권한 모델을 구체적으로 알아야 합니다.

### Encryption, auditing, secrets: 암호화는 저장 위치와 키 관리까지 포함한다

Encryption at rest는 disk나 tablespace에 저장된 데이터가 물리적으로 유출됐을 때 보호를 제공합니다. TLS는 client-server 또는 replication network 구간에서 데이터를 보호합니다. Column/application-level encryption은 DB administrator도 평문을 보지 못하게 하는 목표에 가까울 수 있지만, query 기능과 key management가 어려워집니다. 어떤 암호화를 쓰든 key가 같은 서버 파일에 평문으로 있으면 보호 효과가 크게 줄어듭니다.

Secrets는 코드와 문서에 넣지 않습니다. DB password, API key, private key, connection string은 secret manager나 deployment platform secret store에 두고, 최소 권한과 rotation을 적용합니다. 운영 문서에는 secret의 값이 아니라 "어디에서 주입되고, 누가 읽을 수 있고, 어떻게 회전하고, 회전 중 구버전과 신버전이 어떻게 공존하는가"를 적습니다. 장애 분석 중에도 실제 값은 ticket, chat, commit, screenshot에 남기지 않습니다.

Auditing은 두 가지 질문에 답해야 합니다. 누가 어떤 데이터에 접근했는가, 누가 권한이나 schema를 바꿨는가. 감사 로그는 tamper-resistant storage로 보내고, application user id와 DB session을 연결할 수 있어야 합니다. 단, audit log에 민감한 payload를 그대로 넣으면 로그가 새로운 유출 지점이 됩니다. query text와 bind parameter를 어떻게 redaction할지, privileged access를 어떻게 별도 표시할지 정해야 합니다.

## DBMS별 경계

| 운영 주제 | MySQL 8.4 / InnoDB | PostgreSQL current | 판단 경계 |
|---|---|---|---|
| slow query 수집 | slow query log, Performance Schema statement tables, `EXPLAIN ANALYZE` | `log_min_duration_statement`, `pg_stat_statements`, `EXPLAIN (ANALYZE, BUFFERS)` | query text보다 runtime plan과 wait/resource를 함께 본다 |
| lock 관측 | `SHOW PROCESSLIST`, Performance Schema data locks/metadata locks, `SHOW ENGINE INNODB STATUS` | `pg_stat_activity`, `pg_locks`, `pg_blocking_pids`, wait_event | blocker/waiter/resource graph를 만든다 |
| MVCC 정리 | undo history와 purge 지연 | dead tuple, autovacuum, freeze, visibility map | long transaction의 증상이 다르게 나타난다 |
| replication lag | binlog, relay log, I/O thread, SQL/applier thread, GTID | WAL send/write/flush/replay, replication slot, standby conflict | lag를 전송과 적용 단계로 쪼갠다 |
| capacity 압박 | buffer pool, redo log, dirty page flush, temp table, connection/thread | shared buffers, OS cache, WAL, checkpoint, temp file, autovacuum | 평균 사용률보다 queue와 tail latency를 본다 |
| role/privilege | account와 host, global/db/table/column/routine privileges | role, membership, schema/table/sequence privileges, owner | runtime/migration/owner를 분리한다 |
| row-level policy | SQL 표준적 RLS보다 application predicate나 view/definer pattern을 주로 검토 | Row-Level Security policy 제공 | 제품 기능과 우회 경로를 구분한다 |
| encryption | tablespace encryption, TLS, keyring component 등 | TLS, storage-level encryption은 환경 의존, column/application encryption 설계 | DB 기능과 인프라 기능, key 관리 위치를 분리한다 |
| auditing | Enterprise Audit 또는 plugin/로그/프록시 기반 선택지 | `pgaudit` 확장, logging, proxy/audit pipeline | 감사 목적과 민감값 redaction을 먼저 정한다 |

MySQL에서 운영자가 자주 놓치는 경계는 metadata lock입니다. 긴 SELECT나 transaction이 table metadata lock을 잡고 있으면 DDL이 대기하고, 그 DDL 대기가 뒤따라오는 query들을 막는 chain이 생길 수 있습니다. InnoDB row lock만 보고 있으면 이 상황을 놓칩니다. Performance Schema metadata lock 관측이 필요한 이유입니다.

PostgreSQL에서 자주 놓치는 경계는 autovacuum과 transaction age입니다. query가 느려졌을 때 index만 추가하면 dead tuple과 stale statistics 문제를 덮을 수 있습니다. `pg_stat_user_tables`, `n_dead_tup`, `last_autovacuum`, transaction age, replication slot 상태를 함께 봐야 합니다. 특히 idle in transaction은 lock과 vacuum 양쪽에 영향을 줍니다.

권한 경계도 다릅니다. PostgreSQL에서는 schema usage 권한이 없으면 table select 권한이 있어도 접근이 안 될 수 있고, sequence 권한이 없어 insert가 실패할 수 있습니다. MySQL에서는 account가 user와 host 조합으로 식별되므로 같은 username도 접속 host에 따라 다른 권한을 가질 수 있습니다. 면접에서 권한 문제를 받으면 "GRANT SELECT ON table" 한 줄보다 role hierarchy, schema/object 권한, default privileges, owner를 말해야 합니다.

## 직접 재생해 보기

아래 실험은 민감 데이터가 없는 로컬 disposable DB에서 실행합니다.

### PostgreSQL blocker 찾기

세 session을 엽니다.

```sql
-- session A
BEGIN;
UPDATE account_pg SET balance = balance + 1 WHERE id = 1;

-- session B
BEGIN;
UPDATE account_pg SET balance = balance - 1 WHERE id = 1;

-- session C
SELECT
a.pid,
a.state,
a.wait_event_type,
a.wait_event,
pg_blocking_pids(a.pid) AS blockers,
a.query
FROM pg_stat_activity a
WHERE a.datname = current_database();
```

PASS 신호는 session B가 lock 대기 중이고, `pg_blocking_pids`로 session A를 찾는 것입니다. FAIL 신호는 session A를 commit/rollback하지 않아 실험 후에도 lock이 남는 것입니다. 실험이 끝나면 모든 session을 정리합니다.

### PostgreSQL plan estimate와 actual 비교

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM account_pg
WHERE status = 'ACTIVE'
ORDER BY updated_at DESC
LIMIT 20;
```

PASS 신호는 plan node마다 estimated rows와 actual rows, buffer hit/read, sort 여부를 읽는 것입니다. 만약 actual rows가 estimate보다 크게 다르면 `ANALYZE account_pg;` 후 다시 비교합니다. 그래도 차이가 크면 column correlation, skew, composite index 필요성을 검토합니다.

### MySQL slow query와 InnoDB 상태 보기

MySQL 로컬 환경에서 slow query log를 켜고 작은 실험 query를 실행합니다. 운영 DB에서 전역 설정을 바꾸지 않습니다.

```sql
SET SESSION long_query_time = 0.1;
SET SESSION min_examined_row_limit = 0;

EXPLAIN ANALYZE
SELECT *
FROM account_innodb
WHERE status = 'ACTIVE'
ORDER BY updated_at DESC
LIMIT 20;

SHOW ENGINE INNODB STATUS;
```

PASS 신호는 slow query 기록, explain runtime, InnoDB lock/deadlock section을 서로 연결해 보는 것입니다. FAIL 신호는 slow log에 찍힌 query text만 보고 원인을 확정하는 것입니다.

### 권한 최소화 실험

PostgreSQL 예시입니다.

```sql
CREATE ROLE app_runtime LOGIN;
CREATE ROLE app_migration LOGIN;

CREATE SCHEMA app AUTHORIZATION app_migration;
CREATE TABLE app.orders (
id BIGSERIAL PRIMARY KEY,
tenant_id TEXT NOT NULL,
amount NUMERIC(19, 2) NOT NULL
);

GRANT USAGE ON SCHEMA app TO app_runtime;
GRANT SELECT, INSERT, UPDATE ON app.orders TO app_runtime;
```

실제 secret을 문서나 commit에 남기지 않습니다. 접속 비밀값이 필요하면 로컬 실험 환경의 안전한 주입 경로에서 별도로 설정합니다. PASS 신호는 runtime role이 필요한 DML은 수행하지만 schema owner 권한이나 DDL 권한은 갖지 않는 것입니다. FAIL 신호는 편의를 위해 runtime role을 owner나 superuser로 만드는 것입니다.

### RLS tenant leak 확인

```sql
ALTER TABLE app.orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON app.orders
USING (tenant_id = current_setting('app.tenant_id', true));

BEGIN;
SET LOCAL app.tenant_id = 'tenant-a';
SELECT * FROM app.orders;
COMMIT;
```

PASS 신호는 tenant setting이 바뀌면 보이는 row가 달라지고, setting이 없거나 잘못되면 접근이 실패하거나 빈 결과가 되는 것입니다. FAIL 신호는 connection pool에서 tenant setting이 다음 요청으로 새어 나가는 것입니다. 이 실험은 반드시 transaction-local setting과 pool reset을 함께 검증해야 합니다.

### replication lag 정책 점검

로컬에서 실제 replication을 구성하지 않아도 정책 실험은 가능합니다.

```text
1. 사용자가 주문을 생성합니다.
2. API는 primary DB에 commit합니다.
3. 바로 이어지는 "주문 상세 조회"가 replica를 읽는다고 가정합니다.
4. replica가 5초 늦으면 사용자는 방금 만든 주문을 못 봅니다.
5. 이 화면은 primary read fallback이 필요한가, stale read 안내가 가능한가를 정합니다.
```

PASS 신호는 화면/업무별 stale read 허용 여부가 분리되는 것입니다. FAIL 신호는 모든 read를 replica로 보내고 lag를 운영팀 문제로만 보는 것입니다.

## 면접 꼬리 질문

### 느린 쿼리가 들어오면 제일 먼저 무엇을 보나요?

먼저 범위를 나눕니다. 특정 query인지 전체 DB인지, 특정 시간대인지, primary인지 replica인지, application pool 대기인지 DB 내부 대기인지 봅니다. 그 다음 query fingerprint, 실행 계획, actual rows, buffers/I/O, wait event, lock 상태를 확인합니다. 인덱스 추가는 후보 중 하나이지 첫 결론이 아닙니다.

### lock wait timeout이 많으면 timeout 값을 늘리면 되나요?

대부분은 근본 해결이 아닙니다. timeout을 늘리면 사용자는 더 오래 기다리고, blocker는 더 오래 자원을 잡을 수 있습니다. 먼저 blocker/waiter graph를 만들고, transaction duration, access order, hot row, DDL, missing index를 확인해야 합니다. 일시적 충돌에는 retry가 필요하지만, lock 범위를 줄이는 설계도 같이 봐야 합니다.

### replication lag가 있을 때 애플리케이션은 무엇을 해야 하나요?

업무별 stale read 허용 여부를 정해야 합니다. 쓰기 직후 반드시 최신값이 필요한 경로는 primary read로 보내거나 consistency token을 사용합니다. 지연 허용 가능한 검색/통계 화면은 replica read를 유지할 수 있습니다. lag가 커지면 circuit breaker처럼 replica read를 일시 중단하는 정책도 필요합니다.

### 권한 문제를 빨리 해결하려고 superuser를 주면 왜 위험한가요?

superuser나 owner 권한은 application bug나 SQL injection이 schema 변경, 권한 변경, 데이터 유출, 감사 우회로 이어질 수 있게 합니다. runtime role은 필요한 DML만 가져야 하고, migration role과 owner role은 분리해야 합니다. 권한 오류는 필요한 object privilege와 schema privilege, default privilege를 좁혀 해결하는 편이 안전합니다.

### RLS를 켜면 tenant isolation이 완전히 보장되나요?

RLS는 강한 보호 장치가 될 수 있지만 완전 자동 보장은 아닙니다. policy 누락, owner/bypass 권한, security definer function, session setting 오염, admin path 우회가 있을 수 있습니다. RLS는 application predicate를 대체한다기보다 DB 단의 마지막 방어선으로 보고, cross-tenant test를 자동화해야 합니다.

### 감사 로그에는 무엇을 남겨야 하나요?

누가, 언제, 어떤 권한으로, 어떤 object나 business resource에 접근했는지 재구성할 수 있어야 합니다. request id, application user id, DB role, query fingerprint, row count, status, error code가 유용합니다. 하지만 민감한 parameter 값이나 secret 자체를 남기면 audit log가 유출 지점이 됩니다. redaction과 접근 통제가 필요합니다.

## 함정 질문

### "쿼리가 느리면 인덱스를 추가하면 되죠?"

인덱스는 후보일 뿐입니다. lock wait, stale statistics, wrong join order, parameter skew, disk I/O, checkpoint, pool exhaustion, network, replica lag가 원인일 수 있습니다. 인덱스는 write cost와 storage cost도 만듭니다. 먼저 실행 계획과 wait/resource를 확인해야 합니다.

### "connection pool이 부족하니 max pool size를 올리면 되죠?"

pool 부족의 원인이 connection leak, long transaction, 외부 API inside transaction, DB saturation이면 max만 올려도 해결되지 않습니다. 오히려 DB active session이 늘어 lock contention과 latency가 커질 수 있습니다. pool wait time, active/idle count, transaction duration, DB wait를 함께 봐야 합니다.

### "replica lag는 DBA가 해결할 문제라 application은 신경 안 써도 되죠?"

아닙니다. application이 replica read를 선택한 순간 stale read 가능성을 받아들인 것입니다. 사용자가 방금 쓴 데이터를 바로 읽어야 하는 화면은 application routing 정책이 필요합니다. DBA는 lag 원인을 줄일 수 있지만, consistency 의미는 애플리케이션이 결정해야 합니다.

### "암호화 at rest를 켰으니 데이터 유출 걱정은 끝인가요?"

아닙니다. at-rest encryption은 disk나 snapshot 유출에 대한 보호입니다. DB에 정상 접속한 과권한 role, application log에 찍힌 민감값, 유출된 secret, 관리자 계정 오남용은 별도 문제입니다. TLS, least privilege, audit, secret rotation, application-level encryption이 필요한지 분리해서 판단해야 합니다.

### "권한 오류가 나면 GRANT ALL로 열고 나중에 줄이면 되죠?"

운영에서는 "나중에 줄인다"가 잘 지켜지지 않습니다. broad grant는 사고 표면을 넓히고, 누가 어떤 권한을 왜 갖는지 추적하기 어렵게 만듭니다. 필요한 schema/table/sequence/function 권한을 확인하고, migration과 함께 versioned grant로 남기는 편이 낫습니다.

### "민감 파일을 한 번만 보고 문서에 참고하면 괜찮지 않나요?"

괜찮지 않습니다. 민감 파일은 값이 한 번 문서, commit, chat, ticket에 들어가는 순간 회수하기 어렵습니다. 운영 문서에는 실제 값이 아니라 secret boundary, 주입 경로, rotation 절차, 접근 권한, redaction 정책을 적습니다. 장애 분석에서도 값 자체를 옮기지 않고 필요한 소유자와 안전한 조회 경로를 사용합니다.

### 마지막 연결: 증상에서 원인으로 좁혀 가는 관측 경로

이 주제의 답변은 `사용자 증상 -> request id -> query fingerprint -> wait state -> blocker 또는 resource 병목 -> 안전한 완화 -> 재발 방지`로 이어져야 합니다. `slow query, blocker, wait event, replication lag, privilege, secret boundary`는 따로 외우는 운영 단어가 아니라, 장애를 더 좁은 증거로 바꾸는 관측 경로입니다. 면접에서는 이 경로를 말해야 "인덱스 추가", "권한 부여", "replica 재시작" 같은 빠른 처방이 왜 위험할 수 있는지 설명할 수 있습니다.

작은 반례는 `느린 API를 보고 바로 index를 추가했지만 실제 원인은 long transaction lock wait인 상황`입니다. 이 반례를 처리하지 못하면 앞의 설명이 부분적으로 맞더라도 큰 설명은 틀립니다. 그래서 답변은 항상 공통 원리에서 시작하되 곧바로 범위를 좁혀야 합니다. 어느 DBMS인지, 어느 version인지, 어떤 isolation이나 일관성 요구인지, 이 작업이 운영 중 변경인지 장애 복구인지에 따라 같은 단어가 다른 위험을 만듭니다.

DBMS별로는 PostgreSQL은 pg_stat_activity, pg_locks, EXPLAIN BUFFERS를 보고, MySQL/InnoDB는 processlist, performance schema, InnoDB status, slow log를 봅니다. 이 차이는 세부 취향이 아니라 실제 장애 대응과 설계 판단을 바꿉니다. 면접에서 제품 이름을 들었다면, 그 제품의 문서와 운영 지표로 돌아가야 합니다. 제품 이름이 없다면 일반 원리를 말하되, 일반 원리가 제품별 구현을 덮어쓴다고 말하면 안 됩니다.

운영에서 다시 확인할 신호는 `p95/p99 latency, lock wait seconds, active connection count, replica seconds behind source, failed login/audit events`입니다. 이 값들은 답변을 검증 가능한 문장으로 바꿔 줍니다. 좋은 답변은 `느립니다`, `안전합니다`, `일관됩니다`에서 멈추지 않고, 어떤 값이 어느 범위에 있으면 그렇게 판단하는지 말합니다. 반대로 그 값이 예상과 다르면 처음 세운 mental model을 다시 열어야 합니다.

```text
alert -> request id -> slow query -> wait state -> blocker/root cause -> safe mitigation -> permanent schema/query/transaction fix
```

이 trace를 손으로 다시 그리면, 운영 대응을 용어 목록으로 외웠는지, 아니면 관측값으로 원인을 좁히고 안전한 조치를 고를 수 있는지가 드러납니다. 각 화살표마다 `이 단계에서 읽기 전용으로 확인할 수 있는 값은 무엇인가`, `지금 조치가 되돌릴 수 있는가`, `이 조치가 lock, lag, 권한, secret 노출 중 무엇을 더 악화할 수 있는가`를 붙여 봅니다. 답이 막히는 화살표가 있으면 바로 그 지점에서 장애 대응이 추측으로 바뀝니다.

면접에서는 최종적으로 이렇게 압축할 수 있습니다. 운영 대응은 먼저 read-only 관측으로 증상, 직접 원인, 근본 원인을 나눈 뒤 안전한 완화와 재발 방지를 분리한다고 말합니다. 그 다음 꼬리 질문이 오면, 이 문장의 한 단어를 골라 작은 trace로 내려가면 됩니다. 이 방식은 외운 답을 길게 늘이는 것이 아니라, 짧은 답을 근거 있는 구조로 확장하는 방식입니다.

## 더 깊게 볼 자료

- [MySQL 8.4 Reference Manual - The Slow Query Log](https://dev.mysql.com/doc/refman/8.4/en/slow-query-log.html): MySQL slow query 수집의 공식 기준입니다.
- [MySQL 8.4 Reference Manual - Performance Schema](https://dev.mysql.com/doc/refman/8.4/en/performance-schema.html): statement, wait, lock 관측을 위한 출발점입니다.
- [MySQL 8.4 Reference Manual - InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html): record/gap/next-key lock과 운영 lock 분석을 확인합니다.
- [MySQL 8.4 Reference Manual - Privileges Provided by MySQL](https://dev.mysql.com/doc/refman/8.4/en/privileges-provided.html): MySQL privilege 계층을 확인합니다.
- [PostgreSQL Current Documentation - Monitoring Database Activity](https://www.postgresql.org/docs/current/monitoring.html): `pg_stat_activity`, wait event, 통계 view의 기준 자료입니다.
- [PostgreSQL Current Documentation - The Statistics Collector](https://www.postgresql.org/docs/current/monitoring-stats.html): table/index/query statistics를 해석할 때 봅니다.
- [PostgreSQL Current Documentation - Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html): plan과 actual execution을 함께 읽는 방법을 확인합니다.
- [PostgreSQL Current Documentation - Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html): PostgreSQL lock mode와 row/table lock 동작을 확인합니다.
- [PostgreSQL Current Documentation - Database Roles](https://www.postgresql.org/docs/current/user-manag.html): role과 membership 설계의 공식 기준입니다.
- [PostgreSQL Current Documentation - Privileges](https://www.postgresql.org/docs/current/ddl-priv.html): schema/table/sequence 권한을 확인합니다.
- [PostgreSQL Current Documentation - Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html): RLS 정책과 우회 경계를 확인합니다.
- [PostgreSQL Current Documentation - Encryption Options](https://www.postgresql.org/docs/current/encryption-options.html): PostgreSQL에서 암호화를 어떤 층에서 다룰 수 있는지 확인합니다.
