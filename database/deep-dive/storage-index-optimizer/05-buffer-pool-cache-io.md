# Buffer pool, cache, I/O

이 파일은 page가 메모리와 디스크 사이를 오가는 경로와 checkpoint/flush 압력이 latency로 나타나는 과정을 설명합니다.

## buffer pool과 OS page cache

DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이 절의 핵심 질문은 “같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/database-deep-study-plan.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/runtime-config-resource.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

DBMS가 buffer pool을 갖게 된 배경은 저장 장치와 메모리 사이의 큰 지연 차이입니다. OS page cache만으로도 파일 block 재사용은 가능하지만, DBMS는 page LSN, dirty 상태, transaction visibility, replacement 우선순위, checkpoint와의 관계를 알아야 합니다. 그래서 DB는 자체 buffer를 두고 OS cache와 협력하거나 일부 중복을 줄이는 선택을 합니다.

### 손으로 따라가는 trace

```text
SELECT * FROM accounts WHERE id = 42

DB buffer miss path
  executor needs table/index page P100
  -> shared buffer / buffer pool lookup: miss
  -> OS read request
  -> OS page cache hit or storage read
  -> DB buffer frame gets P100
  -> visibility and predicate check
  -> row returned

DB buffer hit path
  executor needs P100
  -> DB buffer lookup: hit
  -> pin/latch page frame
  -> still checks tuple visibility, locks, predicates
  -> row returned

application cache hit
  request key found in Redis/local cache
  -> SQL may not run at all
  -> DB visibility and buffer path skipped
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. | InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. | buffer pool, DB buffer hit |
| 물리 모델 | PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. | InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. | shared_buffers, dirty page |
| 관측 모델 | PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. | InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. | OS page cache, buffer frame |
| 장애 모델 | PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. | InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. | DB buffer hit, pin |

### 첫 번째 벽돌

첫 번째 벽돌에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 buffer pool, shared_buffers, OS page cache 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 캐시 hit를 애플리케이션 캐시와 같은 의미로 오해하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

buffer pool을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM accounts WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, dirty page와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 shared_buffers, OS page cache, DB buffer hit 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 shared buffer miss를 곧바로 물리 디스크 read로 해석하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

shared_buffers을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM pg_stat_bgwriter;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, buffer frame와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 OS page cache, DB buffer hit, dirty page 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 buffer pool size만 키우면 I/O 문제가 사라진다고 믿는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

OS page cache을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_buffer_pool%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, pin와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 DB buffer hit, dirty page, buffer frame 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 hit ratio 평균만 보고 latch contention, dirty page, eviction을 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

DB buffer hit을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, latch와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 dirty page, buffer frame, pin 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 캐시 hit를 애플리케이션 캐시와 같은 의미로 오해하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

dirty page을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- cold/warm cache에서 같은 SELECT를 반복하고 DB 지표와 OS iostat을 같은 timeline에 둔다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, LRU와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 buffer frame, pin, latch 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 shared buffer miss를 곧바로 물리 디스크 read로 해석하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

buffer frame을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM accounts WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, read-ahead와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 pin, latch, LRU 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 buffer pool size만 키우면 I/O 문제가 사라진다고 믿는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

pin을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM pg_stat_bgwriter;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, double caching와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 latch, LRU, read-ahead 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 hit ratio 평균만 보고 latch contention, dirty page, eviction을 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

latch을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_buffer_pool%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, working set와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 LRU, read-ahead, double caching 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 캐시 hit를 애플리케이션 캐시와 같은 의미로 오해하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

LRU을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, buffer pool와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 같은 page가 DB buffer pool과 OS page cache를 지나 CPU까지 올라올 때, 어떤 cache hit가 무엇을 생략하고 무엇을 생략하지 않는가라는 질문은 storage block -> OS page cache -> DB buffer frame -> executor -> result row 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. DB buffer pool은 애플리케이션 캐시가 아니라 디스크 page를 DB 엔진이 다시 읽고 쓰기 위해 붙잡아 두는 메모리 영역입니다. 이때 read-ahead, double caching, working set 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL shared_buffers는 DB가 직접 관리하는 page cache이고, OS page cache와 함께 작동하므로 shared buffer miss가 곧 물리 disk read라는 뜻은 아닙니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB buffer pool은 table과 index page를 cache하는 핵심 영역이며, dirty page, LRU, flush list, checkpoint age와 강하게 연결됩니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 shared buffer miss를 곧바로 물리 디스크 read로 해석하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

read-ahead을 볼 때 중요한 것은 이름을 외우는 것이 아니라 storage block -> OS page cache -> DB buffer frame -> executor -> result row 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- cold/warm cache에서 같은 SELECT를 반복하고 DB 지표와 OS iostat을 같은 timeline에 둔다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, shared_buffers와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- 캐시 hit를 애플리케이션 캐시와 같은 의미로 오해하는 함정
- shared buffer miss를 곧바로 물리 디스크 read로 해석하는 함정
- buffer pool size만 키우면 I/O 문제가 사라진다고 믿는 함정
- hit ratio 평균만 보고 latch contention, dirty page, eviction을 놓치는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM accounts WHERE id = 42;
SELECT * FROM pg_stat_bgwriter;
SHOW GLOBAL STATUS LIKE Innodb_buffer_pool%;
SHOW ENGINE INNODB STATUS;
-- cold/warm cache에서 같은 SELECT를 반복하고 DB 지표와 OS iostat을 같은 timeline에 둔다.
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

DB buffer hit는 SQL 결과 cache hit가 아닙니다. executor는 여전히 plan을 실행하고 predicate와 visibility를 확인합니다. 또한 OS cache와 DB buffer pool은 단순 경쟁 관계가 아닙니다. PostgreSQL은 둘을 함께 고려하는 경우가 흔하고, InnoDB는 buffer pool 중심 설계와 flush method 선택이 더 직접적인 성능 축이 됩니다.

### 다시 설명해 보기

같은 SELECT를 두 번 실행했을 때 첫 실행과 두 번째 실행의 BUFFERS 결과가 어떻게 달라질지 말해 보십시오. 그 다음 Redis cache hit와 DB buffer hit에서 각각 어떤 일이 생략되는지 한 문장씩 대조해 보십시오.

### 통합 사례: “캐시 hit인데 느린 SELECT”를 다시 진단하기

`EXPLAIN (ANALYZE, BUFFERS)`에서 shared hit가 대부분인데 SELECT가 느린 경우를 생각해 봅니다. 이때 “디스크 문제가 아니니 DB 문제도 아니다”라고 말하면 반쯤만 맞습니다. disk read는 줄었을 수 있지만 executor는 여전히 tuple visibility를 확인하고, predicate를 평가하고, buffer pin/latch를 잡고, dead tuple을 건너뛰고, 정렬이나 집계를 수행할 수 있습니다. DB buffer hit는 I/O 일부를 줄인 것이지 query execution을 생략한 것이 아닙니다.

진단은 세 갈래로 나눕니다. 첫째, plan이 예상보다 많은 page나 row를 만지는지 봅니다. 둘째, page는 memory에 있지만 CPU나 latch contention이 병목인지 봅니다. 셋째, dead tuple이나 bloat 때문에 memory hit가 많아도 쓸모없는 tuple을 많이 거르는지 봅니다. PostgreSQL이라면 BUFFERS, row estimates, wait events, dead tuple 통계를 같이 보고, InnoDB라면 buffer pool hit, row lock wait, history list, handler read 지표를 같이 봅니다.

애플리케이션 cache와의 비교도 필요합니다. Redis나 local cache hit는 SQL 자체를 건너뛰므로 DB transaction visibility와 lock/latch 경로가 사라질 수 있습니다. DB buffer hit는 SQL이 계속 실행되므로 transaction isolation, MVCC visibility, CPU cost가 남습니다. 따라서 “cache hit”라는 단어를 들으면 먼저 어느 계층의 cache인지 물어야 합니다. 계층을 묻지 않으면 해결책도 애플리케이션 cache 증설, index 조정, vacuum, memory sizing 사이에서 흔들립니다.

PASS는 같은 query를 cold/warm 상태에서 실행해 DB buffer hit와 OS read, CPU time, wait event가 어떻게 달라지는지 설명하는 것입니다. FAIL은 hit ratio 하나로 결론을 내거나, OS cache hit와 DB buffer hit를 구분하지 못하거나, page는 memory에 있는데 visibility와 bloat 때문에 느린 상황을 놓치는 것입니다.

### 마지막 실무 연결: cache hit는 어디의 hit인지 먼저 물어야 한다

운영 지표에서 cache hit가 높다고 해서 모든 I/O 문제가 사라진 것은 아니다. DB buffer pool hit가 높아도 OS page cache miss나 storage flush 대기가 병목일 수 있고, 반대로 OS cache가 데이터를 들고 있어도 DB는 buffer pool에서 page replacement 압박을 겪을 수 있다. 애플리케이션 cache hit와 DB cache hit도 전혀 다른 층의 이야기다.

그래서 장애 분석에서는 cache라는 단어를 단독으로 쓰지 않는다. DB buffer pool인지, shared buffers인지, OS page cache인지, 애플리케이션 local cache인지, CDN인지 이름을 붙인다. 그 다음 요청 하나가 어느 cache를 지나고 어느 지점에서 miss가 발생했는지 trace로 그린다. 이 구분이 있어야 메모리를 늘릴지, 쿼리를 줄일지, flush/checkpoint를 볼지 판단할 수 있다.


## random/sequential I/O, fsync, flush, checkpoint pressure

쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이 절의 핵심 질문은 “평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/database-deep-study-plan.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/runtime-config-wal.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-checkpoints.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

WAL/redo와 checkpoint는 모든 commit마다 data page 전체를 강제로 쓰기에는 너무 비싸지만 crash 뒤 committed 변경은 복구해야 한다는 모순에서 나왔습니다. 변경 설명은 먼저 순차 log에 남기고, data page는 dirty 상태로 모아 두었다가 checkpoint와 background flush로 내려보냅니다. 이 절충은 정상 시 성능을 높이지만, flush를 너무 오래 미루면 나중에 쓰기 압력이 몰립니다.

### 손으로 따라가는 trace

```text
T0 workload updates many rows
  -> WAL/redo records are appended sequentially
  -> data pages become dirty in memory

T1 commit returns
  -> log fsync policy decides commit wait
  -> many data pages may still be dirty

T2 dirty pages accumulate
  -> checkpoint distance/checkpoint age grows
  -> background flush tries to catch up

T3 checkpoint pressure
  -> many scattered data pages must be written
  -> foreground queries may wait or help write
  -> API p99 spikes although average TPS looked fine

T4 pressure falls
  -> average latency looks normal again
  -> root cause disappears unless timeline metrics were kept
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. | InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. | random I/O, WAL flush |
| 물리 모델 | PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. | InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. | sequential I/O, redo log |
| 관측 모델 | PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. | InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. | fsync, dirty page |
| 장애 모델 | PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. | InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. | WAL flush, checkpoint |

### 첫 번째 벽돌

첫 번째 벽돌에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 random I/O, sequential I/O, fsync 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 평소 빠른 쓰기만 보고 checkpoint 순간 지연을 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

random I/O을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM pg_stat_bgwriter;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, redo log와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 sequential I/O, fsync, WAL flush 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 fsync 설정을 성능 옵션으로만 보고 crash 손실 범위를 문서화하지 않는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

sequential I/O을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, dirty page와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 fsync, WAL flush, redo log 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 checkpoint를 자주 하면 무조건 좋다고 생각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

fsync을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, checkpoint와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 WAL flush, redo log, dirty page 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 slow query text만 보고 공유 flush pressure를 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

WAL flush을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_log%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, checkpoint age와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 redo log, dirty page, checkpoint 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 평소 빠른 쓰기만 보고 checkpoint 순간 지연을 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

redo log을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- checkpoint 시각, storage write latency, API p99를 같은 timeline에 맞춘다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, flush pressure와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 dirty page, checkpoint, checkpoint age 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 fsync 설정을 성능 옵션으로만 보고 crash 손실 범위를 문서화하지 않는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

dirty page을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM pg_stat_bgwriter;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, group commit와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 checkpoint, checkpoint age, flush pressure 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 checkpoint를 자주 하면 무조건 좋다고 생각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

checkpoint을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance WHERE id = 42;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, tail latency와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 checkpoint age, flush pressure, group commit 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 slow query text만 보고 공유 flush pressure를 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

checkpoint age을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, RPO/RTO와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 flush pressure, group commit, tail latency 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 평소 빠른 쓰기만 보고 checkpoint 순간 지연을 놓치는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

flush pressure을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW GLOBAL STATUS LIKE Innodb_log%;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, random I/O와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 평소 빠른 쓰기가 checkpoint나 flush 순간 p99 latency spike로 바뀌는 경로는 무엇인가라는 질문은 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. 쓰기 성능은 평균적으로 빠르다가 checkpoint나 flush 순간 갑자기 흔들릴 수 있습니다. 이때 group commit, tail latency, RPO/RTO 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 WAL flush, checkpoint_timeout, max_wal_size, checkpoint_completion_target, background writer/checkpointer 동작이 commit 대기와 checkpoint spike에 영향을 줍니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 redo log, dirty page, checkpoint age, page cleaner, adaptive flushing이 함께 작동하며 redo capacity에 가까워질수록 flush pressure가 커집니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 fsync 설정을 성능 옵션으로만 보고 crash 손실 범위를 문서화하지 않는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

group commit을 볼 때 중요한 것은 이름을 외우는 것이 아니라 WAL/redo append -> dirty page accumulation -> checkpoint age -> flush pressure -> tail latency 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- checkpoint 시각, storage write latency, API p99를 같은 timeline에 맞춘다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, sequential I/O와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- 평소 빠른 쓰기만 보고 checkpoint 순간 지연을 놓치는 함정
- fsync 설정을 성능 옵션으로만 보고 crash 손실 범위를 문서화하지 않는 함정
- checkpoint를 자주 하면 무조건 좋다고 생각하는 함정
- slow query text만 보고 공유 flush pressure를 놓치는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
SELECT * FROM pg_stat_bgwriter;
EXPLAIN (ANALYZE, BUFFERS, WAL) UPDATE accounts SET balance = balance WHERE id = 42;
SHOW ENGINE INNODB STATUS;
SHOW GLOBAL STATUS LIKE Innodb_log%;
-- checkpoint 시각, storage write latency, API p99를 같은 timeline에 맞춘다.
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

write가 빠르니 storage 문제가 아니라는 판단은 위험합니다. commit 순간에는 log fsync만 기다렸고 data page 쓰기는 뒤로 밀렸을 수 있습니다. checkpoint는 저장 버튼이 아니라 recovery 시작점을 앞당기는 대신 평상시 write pressure를 만드는 장치입니다. 평균 TPS가 좋아져도 tail latency와 crash recovery 목표가 나쁘면 성공이 아닙니다.

### 다시 설명해 보기

T0 update burst, T1 commit success, T2 dirty page growth, T3 checkpoint start, T4 API p99 spike를 timeline으로 그리고 각 시점의 WAL/redo, data page, fsync, storage queue 상태를 표시해 보십시오.

### 통합 사례: 결제 마감 배치 뒤 API p99가 튀는 이유

결제 마감 배치가 밤 1시에 대량 UPDATE를 수행하고, 1시 5분부터 unrelated API의 p99가 튀는 상황을 생각해 봅니다. slow query log에는 평범한 point SELECT가 잡힐 수 있습니다. 하지만 원인은 그 SELECT의 plan이 아니라 배치가 만든 dirty page와 WAL/redo generation, checkpoint pressure일 수 있습니다. 배치 commit은 빠르게 끝났지만 dirty data page flush가 뒤로 밀려 다른 foreground query가 storage write pressure를 함께 맞는 것입니다.

이때 timeline이 핵심입니다. T0 배치 시작, T1 WAL/redo 생성량 증가, T2 dirty page 비율 증가, T3 checkpoint start, T4 storage write latency 상승, T5 API p99 spike, T6 checkpoint complete를 한 그래프에 둡니다. PostgreSQL에서는 checkpoint write/sync time, buffers_backend, WAL generation을 보고, InnoDB에서는 checkpoint age, dirty page percentage, pending writes, log waits를 봅니다. cloud volume이면 burst credit이나 throttling 지표도 같이 둡니다.

대응은 단일 설정으로 끝나지 않습니다. 배치 transaction을 더 작은 단위로 나누고, 불필요한 index를 줄여 dirty page 수를 낮추고, checkpoint가 더 부드럽게 진행되도록 설정을 조정하고, storage throughput을 확보하고, durability 완화가 필요한 경우 업무 손실 범위를 명시합니다. fsync나 synchronous_commit을 약하게 바꾸는 선택은 성능 개선이 아니라 durability 계약 변경입니다. 이 계약을 제품/운영 책임자가 이해하지 못하면 장애 뒤 설명할 수 없습니다.

PASS는 배치 전후 p99와 checkpoint/flush 지표가 같은 시간축에서 개선되는 것입니다. FAIL은 평균 TPS만 좋아졌는데 checkpoint 구간 tail latency가 그대로이거나, fsync를 약하게 해서 빨라졌지만 crash 시 잃을 수 있는 transaction 범위를 누구도 설명하지 못하는 경우입니다.
