# Storage files, pages, rows

이 파일은 storage-index-optimizer 구간에서 row가 물리 저장 구조 안에 놓이고, MVCC cleanup이 그 구조를 어떻게 다시 다루는지 설명합니다.

## file, page, extent, row, tuple layout

데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이 절의 핵심 질문은 “INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/database-deep-study-plan.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/storage-page-layout.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-physical-structure.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

초기 파일 저장은 record를 순서대로 붙이는 모델로도 설명할 수 있었지만, 다중 사용자 transaction, crash recovery, index lookup, concurrent scan이 들어오면서 row만으로는 충분하지 않았습니다. 디스크와 메모리는 block/page 단위로 움직이고, DBMS는 그 page 안에서 어떤 tuple이 보이는지, 어떤 변경이 먼저 log에 남았는지, 어떤 index entry가 어떤 위치를 가리키는지 판단해야 했습니다.

### 손으로 따라가는 trace

```text
SQL row: {id:42, owner:'kim', balance:10000}
  -> engine tuple/record
  -> page slot or clustered leaf position
  -> buffer frame receives page
  -> WAL/redo records page change
  -> disk file eventually stores page bytes

PostgreSQL heap page
  + PageHeaderData(lsn, lower, upper)
  + ItemId[1] -> tuple offset 8120
  + free space
  + HeapTupleHeader(xmin,xmax,ctid,null bitmap) + values

InnoDB clustered leaf page
  + file/page header
  + infimum/supremum pseudo records
  + user record(key=42,trx_id,roll_ptr,values)
  + page directory
  + file trailer
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. | InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. | file, tuple |
| 물리 모델 | PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. | InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. | page, record |
| 관측 모델 | PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. | InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. | extent, line pointer |
| 장애 모델 | PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. | InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. | tuple, ctid |

### 첫 번째 벽돌

첫 번째 벽돌에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 file, page, extent 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 row 하나 수정이 파일 한 줄 수정이라고 생각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

file을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_column_size(row(42, 'kim', 10000));` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, record와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 page, extent, tuple 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 index entry가 모든 DBMS에서 같은 방식으로 row를 직접 가리킨다고 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

page을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_relation_size('accounts'), pg_total_relation_size('accounts');` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, line pointer와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 extent, tuple, record 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 테이블 크기 증가를 row 수 증가만으로 해석하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

extent을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- pageinspect 사용 가능 환경에서 heap page item을 확인한다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, ctid와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 tuple, record, line pointer 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 pageinspect나 size 통계를 운영 맥락 없이 절대값으로 믿는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

tuple을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW TABLE STATUS LIKE 'accounts';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, clustered index와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 record, line pointer, ctid 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 row 하나 수정이 파일 한 줄 수정이라고 생각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

record을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM information_schema.innodb_tablespaces WHERE name LIKE '%accounts%';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, tablespace와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 line pointer, ctid, clustered index 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 index entry가 모든 DBMS에서 같은 방식으로 row를 직접 가리킨다고 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

line pointer을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_column_size(row(42, 'kim', 10000));` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, free space map와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 ctid, clustered index, tablespace 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 테이블 크기 증가를 row 수 증가만으로 해석하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

ctid을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pg_relation_size('accounts'), pg_total_relation_size('accounts');` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, visibility map와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 clustered index, tablespace, free space map 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 pageinspect나 size 통계를 운영 맥락 없이 절대값으로 믿는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

clustered index을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- pageinspect 사용 가능 환경에서 heap page item을 확인한다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, TOAST와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 tablespace, free space map, visibility map 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 row 하나 수정이 파일 한 줄 수정이라고 생각하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

tablespace을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW TABLE STATUS LIKE 'accounts';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, fillfactor와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 INSERT한 row 하나가 파일 안에서 어떤 단위로 자리를 얻고, SELECT와 UPDATE가 그 자리를 어떻게 다시 찾는가라는 질문은 logical row -> tuple/record -> page slot -> buffer frame -> disk block 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. 데이터베이스의 row는 파일 한 줄에 붙어 있는 문자열이 아니라 page 안에서 header, pointer, payload, visibility metadata로 나뉘어 놓입니다. 이때 free space map, visibility map, TOAST 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL heap page는 PageHeaderData, ItemId 배열, free space, tuple payload, special space로 나뉘며 index entry는 heap TID를 통해 tuple 위치를 다시 찾습니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 primary key clustered index leaf page가 row의 본체가 되고, tablespace, segment, extent, page 계층이 leaf와 non-leaf page를 관리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 index entry가 모든 DBMS에서 같은 방식으로 row를 직접 가리킨다고 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

free space map을 볼 때 중요한 것은 이름을 외우는 것이 아니라 logical row -> tuple/record -> page slot -> buffer frame -> disk block 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM information_schema.innodb_tablespaces WHERE name LIKE '%accounts%';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, file와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- row 하나 수정이 파일 한 줄 수정이라고 생각하는 함정
- index entry가 모든 DBMS에서 같은 방식으로 row를 직접 가리킨다고 보는 함정
- 테이블 크기 증가를 row 수 증가만으로 해석하는 함정
- pageinspect나 size 통계를 운영 맥락 없이 절대값으로 믿는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
SELECT pg_column_size(row(42, 'kim', 10000));
SELECT pg_relation_size('accounts'), pg_total_relation_size('accounts');
-- pageinspect 사용 가능 환경에서 heap page item을 확인한다.
SHOW TABLE STATUS LIKE 'accounts';
SELECT * FROM information_schema.innodb_tablespaces WHERE name LIKE '%accounts%';
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

가장 큰 오해는 “row가 물리 저장의 최소 단위”라는 생각입니다. 실제 최소 운반 단위는 대개 page이고, row는 그 page 안의 해석 가능한 payload입니다. 또 PostgreSQL heap TID와 InnoDB clustered primary key를 같은 포인터처럼 설명하면 secondary index 비용, HOT update 가능성, page split 위험을 모두 잘못 읽게 됩니다.

### 다시 설명해 보기

accounts row 하나를 PostgreSQL과 InnoDB에 각각 넣었다고 가정하고, index lookup이 어떤 page를 거쳐 payload까지 도달하는지 손으로 그려 보십시오. 그림에서 line pointer, tuple header, clustered leaf record, roll pointer 중 하나를 지웠을 때 어떤 설명이 끊기는지도 말해야 합니다.

### 통합 사례: row 설계를 page 비용으로 다시 읽기

`accounts` 테이블에 `memo text`, `metadata jsonb`, `updated_at` 같은 column이 추가되는 상황을 생각해 봅니다. 논리 모델에서는 column 세 개가 늘어난 것뿐이지만, 물리 모델에서는 한 page에 들어가는 tuple 수가 줄고 variable-length payload가 TOAST나 overflow 구조와 만날 가능성이 생깁니다. PostgreSQL에서는 tuple header와 null bitmap, alignment, TOAST 여부가 heap page 밀도를 바꾸고, InnoDB에서는 clustered leaf record의 크기와 page split 가능성이 달라집니다. 이 차이는 곧 buffer hit율, index lookup 뒤 heap fetch 비용, checkpoint 때 써야 할 dirty page 수로 이어집니다.

그래서 시니어는 schema review에서 “이 column이 필요하냐”만 묻지 않습니다. 자주 읽는 hot row에 큰 payload가 같이 붙어도 되는지, update가 잦은 column과 거의 읽지 않는 큰 column을 같은 table에 둘지, PostgreSQL이라면 fillfactor와 HOT update 가능성이 어떤지, InnoDB라면 primary key와 secondary index leaf가 얼마나 커지는지 묻습니다. 같은 데이터를 저장하더라도 row 폭과 page 배치가 달라지면 query plan이 같아도 실제 I/O가 달라집니다.

작은 검증은 세 table을 만들어 비교하면 됩니다. 첫째는 fixed-width column만 둡니다. 둘째는 nullable text를 추가하되 대부분 null로 둡니다. 셋째는 긴 text/json payload를 실제로 채웁니다. 같은 row 수를 넣은 뒤 `pg_relation_size`, `pg_total_relation_size`, `EXPLAIN (ANALYZE, BUFFERS)`를 비교하면 row 수가 같아도 page 수와 buffer 접근이 달라지는 것을 볼 수 있습니다. InnoDB에서는 data_length, index_length, buffer pool read request 변화를 함께 봅니다.

이 사례의 결론은 “큰 column을 쓰지 말라”가 아닙니다. 중요한 것은 논리 row 설계를 page 비용으로 번역하는 습관입니다. API 응답에 필요한 값, 검색 조건에 자주 쓰이는 값, 업데이트가 잦은 값, 감사 목적으로만 남기는 값을 같은 row에 둘 때 엔진이 어떤 page를 몇 번 만지는지 말할 수 있어야 합니다. 그 설명이 가능하면 index, cache, vacuum, WAL 비용이 같은 설계 결정의 다른 면으로 보입니다.

## heap table, clustered table, row movement, vacuum/purge

UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이 절의 핵심 질문은 “UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가”입니다. 이 질문을 닫아야 다음 DU에서 index, optimizer, transaction, recovery를 따로 외우지 않고 하나의 흐름으로 다시 조립할 수 있습니다.

### 공식 근거와 로컬 seed

로컬 seed는 `database/mvcc.md`입니다. 이 seed는 질문과 기존 학습 흔적을 보존하는 재료이고, 구현 세부의 하중은 아래 공식 1차 자료로 다시 닫습니다.

- https://www.postgresql.org/docs/current/routine-vacuuming.html
- https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html

공식 문서는 벤더가 보장하거나 설명하는 구조를 확인하는 근거입니다. 본문 trace는 그 구조를 손으로 따라가기 위해 단순화한 모델이므로, 실제 page의 모든 bit와 예외를 대체하지 않습니다.

### 등장 배경

MVCC는 읽기와 쓰기의 충돌을 줄이기 위해 version을 남기는 방향으로 발전했습니다. 하지만 version을 남기면 반드시 “언제 지울 수 있는가”라는 후속 문제가 생깁니다. PostgreSQL의 vacuum과 InnoDB의 purge는 이 역사적 타협의 청소 절차입니다. cleanup은 미화 작업이 아니라 MVCC가 만든 version 부채를 시스템이 갚는 과정입니다.

### 손으로 따라가는 trace

```text
T0 committed row: accounts[42]=10000
T1 long reader starts and keeps old snapshot
T2 writer updates balance to 9000

PostgreSQL
  old tuple: xmin=10 xmax=30 value=10000  -- T1 may still see it
  new tuple: xmin=30 xmax=-- value=9000   -- newer snapshots see it
  VACUUM waits until no active snapshot needs xid < 30

InnoDB
  clustered record: value=9000 trx_id=30 roll_ptr -> undo(old value=10000)
  read view from T1 follows undo
  purge waits until no read view needs that undo history

if T1 stays open
  PostgreSQL: dead tuples remain, bloat grows
  InnoDB: history list grows, purge lag grows
```

### 같은 질문으로 PostgreSQL과 InnoDB 대조하기

| 질문 | PostgreSQL | MySQL/InnoDB | 관측 신호 |
|---|---|---|---|
| 논리 모델 | PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. | InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. | heap table, HOT update |
| 물리 모델 | PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. | InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. | clustered table, VACUUM |
| 관측 모델 | PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. | InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. | dead tuple, autovacuum |
| 장애 모델 | PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. | InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. | HOT update, purge |

### 첫 번째 벽돌

첫 번째 벽돌에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 독자가 빈손으로 추상 개념을 따라가지 않도록 가장 작은 상태를 먼저 고정합니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 heap table, clustered table, dead tuple 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 long transaction이 vacuum/purge를 막아 bloat/history list를 키우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

heap table을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT relname, n_dead_tup, last_autovacuum FROM pg_stat_all_tables WHERE relname = 'accounts';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, VACUUM와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 구조가 생긴 이유

구조가 생긴 이유에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 이 구조는 이름이 멋있어서가 아니라 비용과 실패 모드를 줄이기 위해 생겼습니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 clustered table, dead tuple, HOT update 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 VACUUM FULL이나 OPTIMIZE TABLE을 원인 분석 전에 실행하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

clustered table을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pid, state, backend_xmin, xact_start FROM pg_stat_activity WHERE backend_xmin IS NOT NULL;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, autovacuum와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### PostgreSQL 쪽 읽기

PostgreSQL 쪽 읽기에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. PostgreSQL을 볼 때는 heap, tuple header, WAL, vacuum, shared buffer가 어떤 순서로 만나는지 분리합니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 dead tuple, HOT update, VACUUM 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 PostgreSQL dead tuple과 InnoDB undo history를 같은 지표처럼 읽는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

dead tuple을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, purge와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### InnoDB 쪽 읽기

InnoDB 쪽 읽기에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. InnoDB를 볼 때는 clustered index, undo, redo, buffer pool, purge가 같은 row 생애 주기에서 어떻게 만나는지 분리합니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 HOT update, VACUUM, autovacuum 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 자동 vacuum/purge가 있으니 애플리케이션 transaction 경계는 중요하지 않다고 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

HOT update을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM information_schema.innodb_trx;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, undo history와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 값 변화 trace

값 변화 trace에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 값이나 page 상태가 언제 바뀌고, 다음 단계에서 누가 그 상태를 소비하는지 따라갑니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 VACUUM, autovacuum, purge 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 long transaction이 vacuum/purge를 막아 bloat/history list를 키우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

VACUUM을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- 긴 트랜잭션을 열고 다른 세션에서 update/delete를 반복한 뒤 cleanup 지표 변화를 본다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, read view와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 운영 관측

운영 관측에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 운영에서는 한 지표를 단독 결론으로 쓰지 않고 시간축과 소비자를 맞춥니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 autovacuum, purge, undo history 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 VACUUM FULL이나 OPTIMIZE TABLE을 원인 분석 전에 실행하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

autovacuum을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT relname, n_dead_tup, last_autovacuum FROM pg_stat_all_tables WHERE relname = 'accounts';` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, xmin horizon와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 성능 함정

성능 함정에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 평균 성능은 좋아 보여도 tail latency, cleanup lag, recovery gap은 뒤늦게 나타날 수 있습니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 purge, undo history, read view 같은 용어는 장식이 아니라 판단 단위입니다. 관측 모델은 통계 view, status output, log line, OS metric을 어떤 순서로 해석할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 PostgreSQL dead tuple과 InnoDB undo history를 같은 지표처럼 읽는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

purge을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT pid, state, backend_xmin, xact_start FROM pg_stat_activity WHERE backend_xmin IS NOT NULL;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, history list length와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 복구 가능성

복구 가능성에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 장애 뒤 설명 가능한 시스템인지 보려면 정상 동작이 아니라 실패 뒤 재구성 경로를 확인합니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 undo history, read view, xmin horizon 같은 용어는 장식이 아니라 판단 단위입니다. 장애 모델은 정상 path가 끊겼을 때 어떤 증거로 안전하게 복구하거나 중단할지 정하는 모델입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 자동 vacuum/purge가 있으니 애플리케이션 transaction 경계는 중요하지 않다고 보는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

undo history을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SHOW ENGINE INNODB STATUS;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, bloat와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 애플리케이션 경계

애플리케이션 경계에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. DB 내부 현상은 connection pool, transaction boundary, batch job, 외부 side effect와 이어집니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 read view, xmin horizon, history list length 같은 용어는 장식이 아니라 판단 단위입니다. 논리 모델은 SQL이 표현하는 row, predicate, transaction boundary입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 long transaction이 vacuum/purge를 막아 bloat/history list를 키우는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

read view을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `SELECT * FROM information_schema.innodb_trx;` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, heap table와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 다른 주제로 전이

다른 주제로 전이에서 UPDATE와 DELETE 뒤 오래된 row/version은 어디에 남고, vacuum 또는 purge는 언제 그 부채를 치울 수 있는가라는 질문은 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 흐름으로 풀어야 합니다. 이 절의 모델은 index, optimizer, transaction, backup 문서를 읽을 때 다시 쓰이는 공통 발판입니다. UPDATE와 DELETE는 기존 row가 즉시 사라지는 사건이 아니라, 오래된 version을 남기고 나중에 cleanup이 따라오는 시간차 과정입니다. 이때 xmin horizon, history list length, bloat 같은 용어는 장식이 아니라 판단 단위입니다. 물리 모델은 page, file, log, buffer, checkpoint처럼 엔진이 실제로 움직이는 단위입니다. 하지만 이 모델만으로는 충분하지 않으므로, PostgreSQL은 heap 안에 old tuple과 new tuple을 두고 xmin/xmax/ctid, transaction status, vacuum horizon으로 visibility와 cleanup을 판단합니다. 반대로 MySQL/InnoDB 쪽에서는 InnoDB는 clustered record의 최신 상태와 undo log의 old version, read view, purge thread를 함께 사용하여 consistent read와 cleanup을 처리합니다. 그래서 같은 SQL을 보더라도 엔진별로 무엇이 실제 소비자인지 먼저 나눠야 합니다. 실무에서 특히 위험한 지점은 VACUUM FULL이나 OPTIMIZE TABLE을 원인 분석 전에 실행하는 함정입니다. 이 함정은 대개 지표 하나를 전체 원인으로 과잉 해석할 때 생깁니다. 검증은 '빠르다/느리다' 같은 감상이 아니라, page나 log나 version이 어느 단계에 멈춰 있는지 관측해서 PASS와 FAIL을 나누는 방식이어야 합니다.

xmin horizon을 볼 때 중요한 것은 이름을 외우는 것이 아니라 old version -> snapshot/read view boundary -> vacuum or purge -> reusable space 중 어느 지점을 설명하는지 아는 것입니다. 예를 들어 `-- 긴 트랜잭션을 열고 다른 세션에서 update/delete를 반복한 뒤 cleanup 지표 변화를 본다.` 같은 관측은 한 번 실행했다고 결론을 주지 않습니다. 실행 전 workload, transaction 상태, checkpoint나 cleanup의 시각, OS 또는 storage 지표를 함께 붙여야 합니다. 그 결과가 본문 모델과 맞으면 PASS이고, clustered table와 연결되는 다음 소비자를 설명하지 못하면 FAIL입니다. 이 구분은 senior practical trap을 막습니다. 운영에서 장애는 대개 한 계층의 단어로 접수되지만, 실제 원인은 논리 row, 물리 page, 메모리 buffer, log flush, 복구 경계가 시간차를 두고 만나는 지점에 있기 때문입니다.

### 시니어가 자주 밟는 함정

- long transaction이 vacuum/purge를 막아 bloat/history list를 키우는 함정
- VACUUM FULL이나 OPTIMIZE TABLE을 원인 분석 전에 실행하는 함정
- PostgreSQL dead tuple과 InnoDB undo history를 같은 지표처럼 읽는 함정
- 자동 vacuum/purge가 있으니 애플리케이션 transaction 경계는 중요하지 않다고 보는 함정

이 함정들은 대부분 지식이 전혀 없어서가 아니라 관측 단위를 잘못 잡아서 생깁니다. SQL 문장, DB 내부 page/log/version, OS storage metric, 애플리케이션 transaction boundary를 같은 시간축에 놓아야 합니다.

### 관측과 검증 경로

```sql
SELECT relname, n_dead_tup, last_autovacuum FROM pg_stat_all_tables WHERE relname = 'accounts';
SELECT pid, state, backend_xmin, xact_start FROM pg_stat_activity WHERE backend_xmin IS NOT NULL;
SHOW ENGINE INNODB STATUS;
SELECT * FROM information_schema.innodb_trx;
-- 긴 트랜잭션을 열고 다른 세션에서 update/delete를 반복한 뒤 cleanup 지표 변화를 본다.
```

PASS는 본문 trace에서 설명한 상태 변화가 관측 지표와 같은 방향으로 나타나는 것입니다. FAIL은 명령이 실패하는 경우만이 아니라, 지표는 움직였지만 그 지표의 소비자를 설명하지 못하거나, 평균값만 좋아지고 tail latency 또는 복구 가능성이 나빠지는 경우도 포함합니다.

### 오해 바로잡기

DELETE가 commit되면 공간이 바로 줄어야 한다는 생각은 MVCC에서 틀립니다. 오래된 snapshot이나 read view가 아직 이전 version을 볼 수 있으면 엔진은 그 흔적을 보존해야 합니다. cleanup 지표가 나쁘다고 해서 항상 vacuum 설정만 탓해서도 안 됩니다. 오래 열린 transaction, replication slot, batch cursor, backup job이 cleanup horizon을 붙잡을 수 있습니다.

### 다시 설명해 보기

두 세션을 그려 보십시오. 세션 A가 오래된 snapshot을 유지하고, 세션 B가 같은 row를 100번 update합니다. PostgreSQL 그림에서는 dead tuple과 xmin horizon을, InnoDB 그림에서는 undo history와 read view를 표시하십시오. A가 commit되는 순간 cleanup 가능 경계가 어떻게 움직이는지도 설명해야 합니다.

### 통합 사례: 삭제 배치 뒤 디스크가 줄지 않는 이유

운영에서 “어제 오래된 주문 3천만 건을 삭제했는데 디스크가 그대로다”라는 질문은 흔합니다. 이 질문에 바로 `VACUUM FULL`이나 `OPTIMIZE TABLE`을 답하면 위험합니다. 먼저 삭제가 논리적으로 commit되었는지, 오래된 snapshot이 남아 있는지, autovacuum 또는 purge가 실제로 처리할 수 있는 경계까지 왔는지, table과 index 중 어느 쪽에 공간 부채가 남았는지 분리해야 합니다. 삭제는 cleanup의 시작 조건이지, 물리 파일 축소 완료 신호가 아닙니다.

PostgreSQL에서는 대량 DELETE 뒤 dead tuple이 늘고, 일반 VACUUM은 그 공간을 같은 relation 안에서 재사용 가능하게 만들 수 있지만 OS 파일 크기를 즉시 줄이지는 않을 수 있습니다. 파일 크기 축소가 필요하면 VACUUM FULL, CLUSTER, pg_repack 같은 더 강한 재작성 계열 선택지가 나오지만, lock과 I/O, replication 영향이 커집니다. 따라서 먼저 `pg_stat_all_tables`, `pg_stat_activity`, `pg_stat_progress_vacuum`, relation/index size를 보고 원인과 목표를 나눠야 합니다.

InnoDB에서는 delete-marked record와 undo history가 purge를 기다릴 수 있습니다. history list가 계속 늘거나 줄지 않으면 오래된 read view, purge 속도, flush pressure를 같이 봐야 합니다. purge가 끝나도 tablespace가 OS에 즉시 반환되지 않을 수 있고, file-per-table, truncate/optimize, table rebuild 여부에 따라 결과가 달라집니다. “삭제했는데 왜 파일이 안 줄지?”라는 질문은 MVCC cleanup과 tablespace allocation을 함께 읽어야 닫힙니다.

이 사례에서 PASS는 디스크가 당장 줄었는지가 아니라, 삭제 후 version 부채가 어떤 지표로 확인되고 어떤 절차로 재사용 또는 반환되는지 설명되는 것입니다. FAIL은 오래된 transaction이 아직 있는데 강한 재작성 작업을 걸거나, 원인 분석 없이 autovacuum 설정만 키우거나, InnoDB history list와 PostgreSQL dead tuple을 같은 처방으로 다루는 경우입니다.
