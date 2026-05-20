# 저장소 성능은 row가 아니라 page와 buffer가 움직이는 방식에서 시작된다

데이터베이스 저장 구조를 설명할 때 "디스크에 row를 저장합니다"라고 말하면 거의 모든 중요한 판단을 놓칩니다. DBMS는 보통 row 하나를 독립 파일처럼 다루지 않습니다. row는 page 안에 들어가고, page는 파일 또는 tablespace 안에 있으며, query와 update는 page를 buffer에 올리고 수정하고 다시 내려보내는 흐름으로 실행됩니다. 그래서 저장소 면접 질문의 중심은 "row가 어디 있나요?"가 아니라 "어떤 page를 읽고, 그 page가 어느 cache에 있으며, dirty page가 언제 어떤 방식으로 안정화되는가?"입니다.

이 문서는 file, page, row, heap, clustered index, buffer pool, OS page cache, dirty page, checkpoint, fsync, random/sequential I/O를 하나의 흐름으로 연결합니다. PostgreSQL과 MySQL/InnoDB를 중심으로 설명하되, 제품별 용어가 같은 구조를 뜻한다고 가정하지 않습니다. PostgreSQL의 table heap과 InnoDB의 clustered index는 모두 row를 저장하지만, row 위치와 index lookup의 비용 구조가 다릅니다. 이 문서가 page와 flush 시간축을 잡는다면, log와 복구 시간축은 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md), access path 선택은 [인덱스와 옵티마이저](04-index-query-optimizer.md), 대용량 DDL의 page/write 증폭은 [스키마와 migration](05-schema-constraints-migration.md)에서 이어서 다룹니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [file, page, row는 세 층이다](#file-page-row는-세-층이다)
    - [heap table과 clustered table은 row의 기준 위치가 다르다](#heap-table과-clustered-table은-row의-기준-위치가-다르다)
    - [buffer pool과 OS page cache는 역할이 겹치지만 같지 않다](#buffer-pool과-os-page-cache는-역할이-겹치지만-같지-않다)
    - [OS는 page cache, filesystem, block layer를 거쳐 I/O를 보낸다](#os는-page-cache-filesystem-block-layer를-거쳐-io를-보낸다)
    - [dirty page는 commit과 다른 시간축에 산다](#dirty-page는-commit과-다른-시간축에-산다)
    - [checkpoint는 로그를 버리는 작업이 아니라 recovery 시작점을 줄이는 작업이다](#checkpoint는-로그를-버리는-작업이-아니라-recovery-시작점을-줄이는-작업이다)
    - [fsync는 write system call이 끝났다와 전원이 나가도 남는다 사이의 경계다](#fsync는-write-system-call이-끝났다와-전원이-나가도-남는다-사이의-경계다)
    - [random I/O와 sequential I/O는 plan 선택과 연결된다](#random-io와-sequential-io는-plan-선택과-연결된다)
    - [row 크기와 page density는 query와 write를 동시에 바꾼다](#row-크기와-page-density는-query와-write를-동시에-바꾼다)
    - [삭제와 정리는 다른 시간축이다](#삭제와-정리는-다른-시간축이다)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
    - [PostgreSQL에서 page 단위 사고 재생](#postgresql에서-page-단위-사고-재생)
    - [PostgreSQL에서 checkpoint와 dirty page 감각 재생](#postgresql에서-checkpoint와-dirty-page-감각-재생)
    - [MySQL/InnoDB에서 buffer pool과 index 경로 재생](#mysqlinnodb에서-buffer-pool과-index-경로-재생)
    - [대용량 삭제 뒤 공간 회수 감각 재생](#대용량-삭제-뒤-공간-회수-감각-재생)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

데이터베이스는 row를 한 줄씩 디스크에서 꺼내는 방식으로 동작하지 않습니다. 대부분의 저장 엔진은 고정 크기 page를 기본 I/O 단위로 삼습니다. PostgreSQL 문서는 table과 index가 보통 8KB page의 배열로 저장된다고 설명하고, InnoDB 문서는 table data와 secondary index가 B-tree 구조의 page들로 나뉜다고 설명합니다. row 하나를 읽어도 실제로는 그 row가 들어 있는 page를 읽습니다. row 하나를 수정해도 그 page가 memory 안에서 dirty 상태가 됩니다.

buffer pool 또는 shared buffer는 DBMS가 page를 main memory에 보관하는 영역입니다. MySQL InnoDB의 buffer pool은 table과 index data를 접근할 때 cache하는 main memory 영역이고, PostgreSQL의 `shared_buffers`도 server가 쓰는 shared memory buffer입니다. 하지만 OS도 page cache를 갖고 있습니다. 그래서 "cache hit"라는 말은 어느 cache의 hit인지 물어야 합니다. DBMS buffer hit이어도 OS writeback이나 fsync에서 밀릴 수 있고, DBMS buffer miss여도 OS cache에 있으면 물리 디스크까지 가지 않을 수 있습니다.

dirty page는 buffer 안에서 바뀌었지만 data file에 아직 완전히 내려가지 않은 page입니다. commit이 되었다고 해서 모든 dirty page가 즉시 data file에 쓰이는 것은 아닙니다. 많은 DBMS는 write-ahead log나 redo log를 먼저 안정화하고, dirty page flush는 checkpoint나 background writer가 나중에 처리하게 합니다. PostgreSQL checkpoint는 dirty data page를 disk로 flush하고 checkpoint record를 WAL에 남깁니다. InnoDB도 fuzzy checkpointing으로 modified page를 작은 batch로 flush하며, crash recovery 때 checkpoint 이후 log를 scan해 필요한 변경을 적용합니다.

I/O 성능은 random과 sequential의 차이에서도 나옵니다. 순차적으로 큰 범위를 읽으면 storage와 OS가 read-ahead를 활용하기 쉽습니다. 반대로 index lookup으로 row를 하나씩 많이 찾으면 index leaf는 가까워도 heap이나 clustered page 접근이 흩어질 수 있습니다. SSD에서는 random I/O 비용이 HDD보다 낮지만, 여전히 page miss, write amplification, fsync, checkpoint burst, buffer eviction은 성능에 큰 영향을 줍니다.

좋은 면접 답변은 "인덱스를 타면 빠르다"에서 멈추지 않습니다. 인덱스는 후보 row 위치를 줄여 줄 수 있지만, 실제 비용은 index page, data page, buffer hit, visibility check, dirty page flush, checkpoint 압력까지 이어집니다. 저장소를 page 단위로 보면 느린 query와 느린 write, checkpoint spike, large table DDL 위험을 같은 언어로 설명할 수 있습니다.

## 먼저 잡아야 할 작은 모델

가장 작은 모델은 계좌 row 하나를 update하는 장면입니다.

```sql
UPDATE accounts
SET balance = balance - 1000
WHERE id = 42;
```

처음에는 row 하나만 바뀐다고 생각하기 쉽습니다. 하지만 저장 엔진 관점에서는 page가 움직입니다.

```text
1. index lookup
   primary key 또는 secondary index에서 id=42 위치를 찾습니다.

2. data page read
   row가 들어 있는 page P10이 buffer에 없으면 disk/OS cache에서 읽어 옵니다.

3. row update
   P10 안의 accounts[42] record 또는 tuple version이 바뀝니다.
   P10 is dirty.

4. log write
   변경을 복구할 수 있는 WAL/redo record가 log buffer에 생깁니다.
   commit 규칙에 따라 log가 durable해집니다.

5. page flush later
   P10은 commit 즉시 data file에 내려갈 수도 있지만,
   보통 checkpoint/background flush 시점까지 memory에 남을 수 있습니다.
```

이 trace에서 움직이는 대상은 row가 아니라 page입니다.

```text
before
  disk data file:
P10: balance=10000, page_lsn=100
  DB buffer:
empty

read
  DB buffer:
P10: balance=10000, clean

update
  DB buffer:
P10: balance=9000, dirty, page_lsn=120
  WAL/redo:
LSN 120 describes the change

commit
  WAL/redo durable up to LSN 130
  client sees success
  disk data file may still have balance=10000

checkpoint/flush
  disk data file:
P10: balance=9000, page_lsn=120
  DB buffer:
P10: clean
```

이 상태 전이를 더 정확히 말하려면 "P10이 어디에 있고, 누가 그 상태를 믿는가"를 나눠야 합니다.

| 순간 | DB buffer의 P10 | WAL/redo | data file의 P10 | client가 믿어도 되는 것 |
| --- | --- | --- | --- | --- |
| read 직후 | clean, balance=10000 | 아직 새 record 없음 | balance=10000 | 아직 변경 없음 |
| update 직후 | dirty, balance=9000 | 변경 record가 buffer에 생김 | balance=10000일 수 있음 | transaction 내부에서만 새 값 |
| commit 직후 | dirty일 수 있음 | commit에 필요한 log가 durable boundary 통과 | balance=10000일 수 있음 | commit 성공이면 새 값은 복구되어야 함 |
| checkpoint/flush 뒤 | clean 또는 evict 가능 | 더 오래된 log는 재사용 후보 | balance=9000 | data file도 새 값을 반영 |

이 표는 "commit이 성공했다"와 "data file page가 새 값을 담고 있다"를 분리합니다. DBMS는 commit 성공 시점에 data file을 반드시 최신으로 만들 필요가 없습니다. 대신 crash 뒤 새 값을 다시 만들 수 있는 log를 먼저 안전하게 남깁니다.

이 모델은 왜 DBMS가 log를 필요로 하는지도 보여 줍니다. commit 뒤 data page가 아직 disk에 내려가지 않았는데 서버가 죽어도, durable log가 있으면 recovery가 P10을 다시 고칠 수 있습니다. 반대로 data page를 매 commit마다 강제로 모두 내려보내면 recovery는 단순해질 수 있지만 정상 처리 성능이 크게 떨어집니다. 그래서 저장 엔진은 log를 먼저 안전하게 남기고 page flush를 지연시키는 설계를 많이 씁니다.

또 하나의 작은 모델은 read path입니다.

```text
SELECT * FROM accounts WHERE id = 42

buffer hit path
  DB buffer has page P10
  executor reads row from memory

DB buffer miss but OS cache hit
  DBMS asks OS to read file block
  OS page cache has it
  copy into DB buffer is still needed

physical read
  OS cache misses too
  storage device reads block/page
  DBMS receives page and places it in buffer
```

이 read path에서는 "읽었다"라는 말도 세 단계로 나뉩니다.

| 관측 문장 | 실제 의미 | 아직 모르는 것 |
| --- | --- | --- |
| DB buffer hit | DBMS가 자기 buffer frame에서 page를 찾았습니다. | OS cache나 device 상태는 이번 요청에서 보지 않았습니다. |
| DB buffer miss, OS cache hit | DBMS buffer에는 없었지만 커널 page cache에는 파일 block이 있었습니다. | 물리 장치까지 내려가지 않았더라도 DBMS buffer로 복사하는 비용은 있습니다. |
| physical read | OS page cache에도 없어 storage에서 block을 읽었습니다. | 장치 queue, filesystem layout, read-ahead 정책에 따라 지연이 달라집니다. |

따라서 "cache hit ratio가 높다"는 지표를 볼 때는 어느 cache의 hit인지 먼저 물어야 합니다. DBMS buffer hit가 높으면 executor가 page를 빠르게 얻었을 가능성이 크지만, 동시에 checkpoint나 fsync가 밀리는 write 병목이 없다는 뜻은 아닙니다. OS cache hit가 높으면 물리 read는 줄었을 수 있지만, DBMS는 그 page를 자기 buffer에 올려 page header, visibility, latch 같은 내부 의미를 다시 관리해야 합니다.

따라서 "디스크를 안 읽었다"는 말도 층위를 나눠야 합니다. DBMS buffer에서는 miss였지만 OS cache hit였을 수 있습니다. OS cache에서도 miss라면 storage device I/O가 일어납니다. direct I/O를 쓰는 설정에서는 OS cache를 덜 쓰는 경로도 있습니다. 면접에서는 특정 제품과 설정을 모른다면 "어느 cache 기준인지 확인해야 한다"고 말하는 편이 정확합니다.

## 깊은 메커니즘

### file, page, row는 세 층이다

DBMS의 data file은 보통 page들의 배열로 볼 수 있습니다. PostgreSQL 문서는 table과 index가 fixed-size page 배열로 저장되고, table page에서는 row가 page 어느 곳에든 들어갈 수 있다고 설명합니다. page에는 header, item identifier 배열, free space, item data, index access method가 쓰는 special space 같은 영역이 있습니다. PostgreSQL heap page에서 item identifier는 item의 offset과 length를 가리키고, CTID는 page number와 item identifier index로 row 위치를 표현합니다.

이 구조가 필요한 이유는 row가 항상 같은 byte 위치에 머물지 않기 때문입니다. page 안에서 row data를 compaction하거나 새로운 tuple version을 추가할 수 있고, item identifier가 안정적으로 남으면 외부 참조는 page 번호와 item slot을 통해 row를 찾을 수 있습니다. row를 직접 file offset으로만 기억하면 page 내부 이동이 매우 위험해집니다.

```text
PostgreSQL heap page simplified

+-------------------------+
| PageHeaderData          |
|  pd_lsn                 |
|  pd_lower / pd_upper    |
+-------------------------+
| ItemIdData[1] -> item A |
| ItemIdData[2] -> item B |
+-------------------------+
| free space              |
+-------------------------+
| tuple B                 |
| tuple A                 |
+-------------------------+
```

InnoDB는 table data를 clustered index라는 B-tree 구조에 둡니다. MySQL 문서는 InnoDB table data와 secondary indexes가 B-tree index structure를 사용하고, 전체 table을 나타내는 B-tree index를 clustered index라고 부른다고 설명합니다. clustered index의 leaf record는 row의 모든 column 값을 담고, secondary index의 leaf record는 index column 값과 primary key column 값을 담습니다. 그래서 secondary index에서 row를 찾을 때는 secondary index leaf에서 primary key를 얻고, 다시 clustered index를 찾아 row data에 도달할 수 있습니다.

이 차이는 query 비용을 바꿉니다. PostgreSQL의 일반 index scan은 index와 heap을 모두 볼 수 있고, heap row가 어디든 있을 수 있어 random heap access가 비용이 됩니다. InnoDB secondary index scan은 secondary index leaf와 clustered index lookup이 연결됩니다. primary key가 길면 secondary index entry가 커지는 이유도 여기에 있습니다.

### heap table과 clustered table은 row의 기준 위치가 다르다

heap table은 row가 특정 key 순서로 물리 정렬되어 있다는 약속을 하지 않습니다. PostgreSQL heap에서는 row version이 heap page에 있고, index는 heap tuple을 가리킵니다. update가 일어나면 새 tuple version이 생기고, 오래된 tuple은 transaction visibility 규칙과 vacuum을 거쳐 정리됩니다. 그래서 "table data가 primary key B-tree leaf에 있다"고 PostgreSQL 일반 table을 설명하면 틀립니다.

clustered table은 row data가 cluster key의 leaf 쪽에 놓이는 모델입니다. InnoDB에서는 clustered index가 row data를 저장하고, primary key가 보통 clustered index입니다. 이 구조는 primary key lookup을 빠르게 만들지만, secondary index가 primary key를 포함해야 하므로 primary key 길이와 update 패턴이 전체 index 비용에 영향을 줍니다. 또한 primary key 순서가 insertion locality와 page split에 영향을 줍니다.

```text
PostgreSQL-like heap + secondary index
  email index leaf:
kim@example.com -> (heap page 10, item 3)
  heap page 10:
item 3 -> full tuple

InnoDB clustered + secondary index
  secondary email index leaf:
kim@example.com -> primary key 42
  clustered primary key leaf:
id 42 -> full row
```

두 구조 모두 "index를 타면 row를 더 빨리 찾는다"는 큰 목표를 공유하지만, "row data가 어디에 있는가"와 "secondary lookup이 무엇을 다시 찾아야 하는가"가 다릅니다. 이 경계를 알고 있어야 covering index, index-only scan, primary key design, row movement 비용을 정확히 설명할 수 있습니다.

### buffer pool과 OS page cache는 역할이 겹치지만 같지 않다

DBMS buffer는 DBMS가 page 단위로 관리하는 memory입니다. 어떤 page가 dirty인지, 어떤 page가 어떤 LSN까지 반영했는지, 어떤 page를 eviction할 수 있는지, 어떤 latch가 필요한지는 DBMS가 알아야 합니다. OS page cache는 filesystem block을 memory에 cache합니다. OS는 DBMS page의 transaction 의미를 모릅니다. OS 입장에서는 파일 byte range가 자주 읽히거나 쓰인다는 사실을 볼 뿐입니다.

PostgreSQL 문서는 `shared_buffers`를 database server가 shared memory buffer로 쓰는 양이라고 설명하면서도, PostgreSQL이 operating system cache에도 의존하므로 전용 서버에서 무작정 매우 크게 잡는 것이 항상 더 좋지는 않다고 설명합니다. MySQL InnoDB 문서는 buffer pool이 table과 index data를 cache하는 main memory 영역이고, high-volume read를 위해 page 단위로 나뉜다고 설명합니다.

```text
read path with two caches

executor asks storage manager for page P10
  |
  v
DB buffer lookup
  hit  -> return page P10
  miss -> ask OS to read file block
         |
         v
      OS page cache
        hit  -> copy/read into DB buffer
        miss -> storage device read -> OS cache -> DB buffer
```

여기서 DBMS buffer와 OS page cache는 같은 데이터를 두 번 들고 있을 수도 있습니다. 이를 double buffering이라고 부르기도 합니다. 이것이 항상 낭비라는 뜻은 아닙니다. DBMS buffer는 page의 dirty 여부, pageLSN, replacement 정책, latch와 transaction visibility 같은 DBMS 의미를 붙잡습니다. OS page cache는 file offset과 block 단위 재사용, read-ahead, writeback을 관리합니다. 두 계층의 목적이 다르기 때문에, 어떤 DBMS와 storage 설정에서는 buffered I/O를 쓰고, 어떤 구성에서는 direct I/O나 유사 설정으로 OS cache 경로를 줄이기도 합니다. 제품과 설정을 모르면 "DBMS가 OS cache를 완전히 우회한다"고 단정하지 않는 편이 안전합니다.

운영에서 "cache hit ratio가 높다"는 지표를 볼 때는 어떤 층의 hit인지 확인해야 합니다. InnoDB buffer pool hit가 높으면 storage read pressure가 낮다는 좋은 신호일 수 있습니다. PostgreSQL shared buffer hit가 높아도, checkpoint fsync나 OS writeback에서 latency가 생길 수 있습니다. 반대로 DBMS buffer hit가 낮아도 OS cache가 흡수하고 있다면 물리 device read는 적을 수 있습니다. 하지만 DBMS는 OS cache 내부 상태를 transaction 의미와 직접 연결하지 못하므로, 안정성은 WAL/fsync 같은 명시적 flush 경계로 닫아야 합니다.

### OS는 page cache, filesystem, block layer를 거쳐 I/O를 보낸다

DBMS가 page를 읽고 쓴다고 말할 때, 그 page가 곧바로 SSD 셀이나 HDD platter로 이동한다고 생각하면 안 됩니다. 일반적인 buffered I/O 경로에서는 DBMS가 file descriptor에 대해 `read`나 `write`를 호출하고, 커널은 page cache를 통해 파일 내용을 메모리에 보관합니다. Linux kernel 문서는 page cache가 사용자와 커널 나머지 부분이 filesystem과 상호작용하는 주된 방법이며, `O_DIRECT` 같은 경로로 우회할 수 있다고 설명합니다. 즉 DBMS buffer pool과 OS page cache는 서로 다른 층에서 같은 물리 I/O를 줄이려는 cache입니다.

```text
DBMS buffer miss
  |
  v
read(file, block range)
  |
  v
kernel page cache
  hit  -> DBMS buffer로 page 내용 복사
  miss -> filesystem maps file offset to disk blocks
            |
            v
          block layer queues bio/request
            |
            v
          device driver submits command
            |
            v
          storage returns bytes
```

읽기 경로는 "요청이 아래로 내려가고 bytes가 위로 올라오는" 왕복입니다. DBMS가 원하는 것은 page `P10`이지만, 커널은 파일 offset과 block mapping으로 봅니다.

```text
DBMS term
  relation file + page number P10

kernel/filesystem term
  file descriptor + byte offset range

block layer term
  block device + sector/block request

device term
  controller command + media read
```

이 변환 때문에 DBMS 관측과 OS 관측을 맞춰 읽어야 합니다. DBMS에서는 "heap block 1234 read"처럼 보인 일이 OS에서는 특정 파일의 offset read이고, block layer에서는 여러 bio/request로 쪼개질 수 있습니다. 반대로 `iostat`에서 await가 높다고 해서 곧바로 특정 SQL 하나가 문제라고 말할 수도 없습니다. 같은 device queue를 checkpoint, temp file spill, 다른 process가 함께 쓰고 있을 수 있습니다.

쓰기 경로도 같은 층을 지나지만, 읽기보다 더 조심해야 합니다. `write`가 성공했다는 말은 커널이 데이터를 받았다는 뜻일 수 있고, 장치의 비휘발성 저장소에 반드시 도착했다는 뜻은 아닙니다. Linux `fsync(2)` manual은 `fsync()`가 파일의 수정된 in-core data와 metadata를 disk device 또는 영구 저장 장치로 flush하고, disk cache가 있으면 그것까지 write-through 또는 flush한다고 설명합니다. Linux block layer 문서도 volatile write-back cache가 있는 장치는 운영체제에 I/O 완료를 먼저 알릴 수 있으므로, `fsync`, `sync`, unmount 같은 data integrity operation에서 비휘발성 저장소로 강제 flush해야 한다고 설명합니다.

```text
DBMS dirty page or WAL buffer
  |
  v
write()
  kernel page cache or direct I/O path
  |
  v
filesystem journaling / allocation / ordering rules
  |
  v
block layer request queue
  |
  v
device driver
  |
  v
storage controller cache
  |
  v
non-volatile media

fsync/fdatasync
  waits for the relevant dirty data and required metadata
  may issue cache flush/FUA so completion means a stronger durability boundary
```

쓰기에서 자주 헷갈리는 acknowledgement도 분리해야 합니다.

| 완료처럼 보이는 사건 | 강한 정도 | DB 관점의 의미 |
| --- | --- | --- |
| DBMS log buffer에 record 생성 | 가장 약함 | crash가 나면 메모리와 함께 사라질 수 있습니다. |
| `write()`가 성공 | 중간 | 커널 또는 direct I/O 경로로 bytes를 넘겼지만, 장치의 비휘발성 저장까지 보장한다고 단정하면 안 됩니다. |
| `fdatasync`/`fsync`가 성공 | 더 강함 | 해당 파일 data와 필요한 metadata가 durable boundary를 통과했다고 보는 지점입니다. 실제 강도는 OS, filesystem, device cache 신뢰성에 의존합니다. |
| storage가 forced flush/FUA를 완료 | 장치 경계 | volatile write-back cache가 비워졌다고 기대하는 지점입니다. 장치와 controller가 거짓 완료를 알리면 DBMS도 속을 수 있습니다. |

따라서 durability는 DBMS 코드만의 성질이 아니라, DBMS 설정, 커널 sync 동작, filesystem, storage cache 정책이 함께 만드는 계약입니다. 면접 답변에서는 `fsync`를 "느린 함수"로만 말하지 말고, commit 성공 응답이 어느 durable boundary 뒤에 나가는지와 연결해야 합니다.

이 경로는 왜 DB 운영에서 CPU scheduler와 memory pressure도 중요해지는지 보여 줍니다. DBMS background writer나 checkpointer가 runnable 상태여도 CPU를 받지 못하면 flush가 늦어질 수 있습니다. 메모리가 부족하면 커널 writeback이나 reclaim이 DBMS foreground query와 겹쳐 지연을 만들 수 있습니다. block queue가 길어지면 DBMS 내부에서는 단순한 `fsync` 지연처럼 보이지만, 실제 원인은 filesystem journal, block scheduler, driver, 장치 cache flush 중 하나일 수 있습니다. 그래서 storage 장애를 볼 때는 DB 지표와 함께 `iostat`, `pidstat`, `vmstat`, `perf`, filesystem mount option, device cache 정책을 함께 봐야 합니다.

### dirty page는 commit과 다른 시간축에 산다

dirty page는 memory 안에서 바뀐 page입니다. commit은 transaction 결과를 사용자에게 성공으로 알리는 논리적 사건이고, dirty page flush는 data file에 page image를 쓰는 물리적 사건입니다. 두 사건을 항상 같은 시각으로 묶으면 DBMS 성능을 설명할 수 없습니다.

```text
time ---->

T1 update page P10
  P10 dirty in buffer

T1 commit
  WAL/redo durable
  client receives success

background writer/checkpoint
  P10 flushed to data file

crash can happen between commit and flush
  recovery uses WAL/redo to bring data file forward
```

이 시간축을 네 개의 시계로 나누면 더 안전합니다.

| 시계 | 진행을 보여 주는 대표 단서 | 늦어지면 생기는 현상 |
| --- | --- | --- |
| transaction clock | statement 시작, commit, rollback | lock wait, snapshot 보존, undo/vacuum 지연 |
| log clock | WAL/redo write LSN, flushed LSN | commit latency, recovery에 필요한 log 범위 증가 |
| dirty page clock | dirty buffer count, checkpoint age | checkpoint spike, eviction 지연, writeback 압력 |
| OS/device clock | writeback, block queue, cache flush | `fsync` 지연, read/write latency 증가 |

한 transaction이 commit되었다고 해서 네 시계가 모두 같은 지점에 도착한 것은 아닙니다. commit 시계는 client 응답까지 갔지만 dirty page 시계는 아직 뒤에 있을 수 있습니다. 이 간격을 허용하려고 WAL/redo가 있고, 이 간격이 너무 커지면 checkpoint와 recovery 시간이 문제가 됩니다.

이 구조가 가능한 이유는 write-ahead rule입니다. data page 변경이 disk data file에 늦게 내려가더라도, 그 변경을 다시 만들 수 있는 log가 먼저 안정화되면 crash recovery가 빠진 변경을 redo할 수 있습니다. PostgreSQL WAL reliability 문서는 committed transaction의 data가 power loss 등에도 안전한 nonvolatile area에 저장되어야 한다고 설명하고, OS buffer cache와 disk/controller cache 같은 여러 cache layer 때문에 강제 flush가 단순하지 않다고 설명합니다. 그래서 `fsync`나 sync method, storage write cache 신뢰성이 실제 durability와 연결됩니다.

dirty page를 빨리 flush하면 crash recovery 부담은 줄 수 있지만 정상 I/O가 증가합니다. 너무 늦게 flush하면 checkpoint 때 많은 dirty page가 몰리고 recovery가 길어질 수 있습니다. DBMS는 이 균형을 checkpoint, background writer, flush policy, log file size, dirty page percentage 같은 설정으로 다룹니다.

### checkpoint는 로그를 버리는 작업이 아니라 recovery 시작점을 줄이는 작업이다

PostgreSQL checkpoint는 transaction sequence의 한 지점에서 heap과 index data files가 그 checkpoint 이전 WAL 정보를 반영하도록 보장하는 지점입니다. checkpoint 때 dirty data page가 disk로 flush되고 checkpoint record가 WAL에 쓰입니다. crash recovery는 최신 checkpoint record를 보고 REDO를 시작할 log 위치를 정합니다. checkpoint 이전의 변경은 data file에 있다고 볼 수 있으므로, 그 이전 WAL segment는 archiving 요구가 없으면 재사용하거나 제거할 수 있습니다.

InnoDB도 checkpoint mechanism을 갖습니다. MySQL 문서는 InnoDB가 fuzzy checkpointing을 구현하고, modified database pages를 buffer pool에서 작은 batch로 flush한다고 설명합니다. crash recovery 때 checkpoint label을 찾고, 그 label 이전의 modification은 disk image에 있다고 보고, log file을 checkpoint부터 forward scan해 logged modification을 database에 적용합니다.

```text
checkpoint mental model

LSN 100 update P1
LSN 120 update P2
LSN 150 CHECKPOINT
LSN 170 update P3
LSN 190 commit
crash

recovery:
  start near checkpoint
  redo changes after checkpoint if data page is older
```

checkpoint는 한 번에 "모든 것을 깨끗하게 만든다"라기보다, 복구가 다시 읽어야 할 출발선을 앞으로 당기는 작업입니다.

```text
before checkpoint
  dirty pages:
    P1 pageLSN=100
    P2 pageLSN=120
  WAL durable:
    up to LSN 150

during checkpoint
  write P1/P2 data pages
  fsync data files as required by engine policy
  write checkpoint record

after checkpoint
  recovery can usually start near checkpoint record
  new updates after LSN 150 still need WAL/redo
```

여기서 `usually`라고 조심스럽게 말하는 이유는 제품별로 checkpoint record, restart point, fuzzy checkpoint, full-page image, doublewrite 같은 세부가 다르기 때문입니다. 공통 원리는 "복구가 처음부터 모든 log를 읽지 않도록, data file이 어느 지점까지 따라왔는지 표시한다"입니다. 세부 구현은 PostgreSQL과 InnoDB 문서로 나눠 확인해야 합니다.

checkpoint는 정상 처리에도 비용을 만듭니다. PostgreSQL 문서는 checkpoint가 dirty buffer를 쓰기 때문에 상당한 I/O load를 일으킬 수 있고, checkpoint를 더 자주 하면 crash recovery는 빨라질 수 있지만 dirty page flush 비용이 늘어난다고 설명합니다. InnoDB에서도 log file size와 checkpointing은 disk I/O와 recovery time 사이의 trade-off를 만듭니다. 따라서 checkpoint spike를 "갑자기 디스크가 느려졌다"로만 보면 부족하고, dirty page accumulation과 log checkpoint age를 함께 봐야 합니다.

### fsync는 write system call이 끝났다와 전원이 나가도 남는다 사이의 경계다

애플리케이션이 파일에 write를 호출했다고 해서 data가 durable storage에 도착했다는 뜻은 아닙니다. OS buffer cache, controller cache, disk cache가 사이에 있습니다. PostgreSQL reliability 문서도 disk와 main memory 사이에 여러 cache layer가 있고, OS는 application이 buffer cache에서 disk로 강제 write할 방법을 제공하며 PostgreSQL이 그것을 사용한다고 설명합니다. 하지만 drive/controller write-back cache가 power loss를 견디지 못하면 reliability hazard가 될 수 있습니다.

DBMS의 durability 설정은 이 경계를 조정합니다. PostgreSQL의 `synchronous_commit`, `wal_sync_method`, storage cache 설정, MySQL의 `innodb_flush_log_at_trx_commit` 같은 설정은 commit latency와 crash loss window 사이의 trade-off를 만듭니다. 면접에서는 숫자를 외우기보다 원리를 말해야 합니다. "commit 성공을 언제 client에게 말할 것인가"와 "그 시점에 어떤 durable write가 끝났다고 믿는가"가 핵심입니다.

이 절의 핵심을 손으로 그리면 다음과 같습니다.

```text
T1 commit wants to say "success"

weak path
  WAL record in DBMS memory
    -> client success
    -> power loss
    -> record may be gone

stronger path
  WAL record in DBMS memory
    -> write WAL file
    -> fsync/fdatasync according to DBMS policy
    -> device acknowledges required flush
    -> client success
    -> power loss
    -> recovery can read WAL record
```

실제 DBMS는 group commit처럼 여러 transaction의 log flush를 묶어 commit latency를 줄일 수 있습니다. 이 최적화도 원리를 바꾸지는 않습니다. 여러 transaction이 한 번의 flush boundary를 공유할 수 있을 뿐, 성공이라고 답한 transaction의 log가 DBMS가 요구한 durable boundary를 통과해야 한다는 질문은 남습니다.

### random I/O와 sequential I/O는 plan 선택과 연결된다

row 하나만 찾는다면 index lookup이 좋아 보입니다. 하지만 많은 row를 찾는다면 index lookup이 오히려 비쌀 수 있습니다. index leaf를 따라 후보를 찾은 뒤 data page를 여기저기 random하게 읽어야 할 수 있기 때문입니다. PostgreSQL index-only scan 문서는 일반 index scan에서 index와 heap 양쪽을 fetch해야 하고, index entries는 가까워도 heap rows는 어디든 있을 수 있어 random heap access가 느릴 수 있다고 설명합니다. bitmap scan은 heap access를 물리 순서로 모아 이 비용을 줄이려 합니다.

```text
wide range query

Index path:
  index leaf scan is ordered
  heap pages: P3, P901, P14, P502, P3, P77 ...
  many random heap visits

Sequential scan path:
  heap pages: P1, P2, P3, P4, P5 ...
  filter every row
  predictable sequential read
```

이 차이는 OS와 storage 층에서도 다르게 보입니다.

| 접근 패턴 | DBMS 내부 증상 | OS/storage 쪽 기대 |
| --- | --- | --- |
| 많은 random lookup | buffer miss가 여러 relation page로 흩어집니다. | read-ahead가 덜 맞고 queue가 작은 요청으로 채워질 수 있습니다. |
| sequential scan | 많은 page를 읽지만 순서가 예측 가능합니다. | readahead와 더 큰 연속 I/O가 맞을 수 있습니다. |
| bitmap heap scan | 먼저 row 위치를 모은 뒤 heap page 순서로 방문합니다. | random 방문을 줄이는 대신 memory bitmap 비용이 생깁니다. |
| sort/hash spill | temp file page가 추가로 생깁니다. | query plan 문제가 OS write/read 병목처럼 보일 수 있습니다. |

따라서 느린 query를 볼 때 index 유무만 묻지 말고, `EXPLAIN (ANALYZE, BUFFERS)`의 buffer read/hit, temp read/write, OS의 read latency와 queue도 같이 봅니다. 같은 SQL도 cache가 따뜻할 때와 checkpoint가 몰리는 때의 실제 시간은 달라질 수 있습니다.

SSD에서는 random read가 HDD보다 훨씬 싸지만, random access가 완전히 무료인 것은 아닙니다. page miss가 많아지면 buffer churn이 생기고, CPU cache locality도 나빠질 수 있습니다. write 쪽에서는 random update가 dirty page를 넓게 퍼뜨려 checkpoint flush와 write amplification을 키울 수 있습니다. 따라서 optimizer의 cost model이 `seq_page_cost`, `random_page_cost` 같은 추상 비용을 쓰는 이유도 여기에 있습니다. 실제 hardware, cache 상태, table size, selectivity에 따라 선택이 달라집니다.

### row 크기와 page density는 query와 write를 동시에 바꾼다

row가 작으면 한 page에 더 많은 row가 들어가고, 같은 query가 읽어야 하는 page 수가 줄어듭니다. MySQL InnoDB row format 문서는 더 많은 row가 한 disk page에 들어가면 query와 index lookup이 더 빠를 수 있고, buffer pool memory와 updated value write I/O가 줄어든다고 설명합니다. 반대로 wide row, 큰 variable-length column, off-page storage는 page 수와 I/O를 늘립니다.

```text
8KB page example, simplified

small row 100 bytes:
  about dozens of rows per page
  range scan touches fewer pages

wide row 2000 bytes:
  only a few rows per page
  same row count touches many pages
  secondary lookup may need more data page reads
```

정규화와 반정규화도 page 관점에서 다시 볼 수 있습니다. 너무 많이 join해야 하면 여러 index/page를 오가야 하고, 너무 크게 denormalize하면 한 row가 넓어져 buffer 효율이 떨어질 수 있습니다. schema 설계는 논리 모델만의 문제가 아니라 page density, index width, update fan-out과도 연결됩니다.

### 삭제와 정리는 다른 시간축이다

row를 삭제했다고 disk file이 즉시 줄어드는 것은 아닙니다. PostgreSQL에서는 MVCC 때문에 오래된 tuple version이 남고, vacuum이 더 이상 보이지 않는 dead tuple을 정리해 공간을 재사용 가능하게 만듭니다. InnoDB에서도 delete-marked record와 undo history, purge가 얽힙니다. 이 지연은 단순한 낭비가 아니라 동시성 제어와 복구 가능성을 지키기 위한 대가입니다.

```text
DELETE FROM orders WHERE created_at < '2025-01-01'

logical effect
  새 snapshot에서는 해당 row가 보이지 않습니다.

physical effect
  page 안에는 정리 대기 흔적이 남을 수 있습니다.
  long transaction이 오래된 version을 필요로 하면 정리가 밀립니다.

maintenance
  PostgreSQL vacuum reclaims reusable space inside relation.
  InnoDB purge removes old versions when no read view needs them.
  file size 반환은 별도 table rewrite, optimize, truncate 같은 작업이 필요할 수 있습니다.
```

이 시간축을 모르면 대용량 삭제 뒤 "왜 디스크가 그대로인가요?", "왜 다음 insert는 빨라졌는데 파일 크기는 줄지 않나요?", "왜 오래 열린 transaction이 storage bloat를 만들죠?" 같은 질문에 답하기 어렵습니다.

## DBMS별 경계

PostgreSQL에서 table과 index는 보통 8KB page 배열로 설명됩니다. table heap page는 row item을 저장하고, B-tree index page는 index entry와 sibling link 같은 access method-specific 정보를 갖습니다. 일반 index는 heap과 분리되어 있으므로 index scan 뒤 heap visibility 확인이 필요합니다. index-only scan은 index에 필요한 column이 있고 visibility map이 all-visible인 heap page가 많을 때 heap access를 줄입니다. 왜 index entry만 보고도 row를 반환하지 못하는지는 [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)의 visibility 판단과 함께 보면 더 분명합니다.

PostgreSQL은 OS cache도 활용합니다. `shared_buffers`를 너무 크게 잡으면 항상 이득이라는 식으로 설명할 수 없습니다. 공식 문서는 PostgreSQL이 OS cache에도 의존하므로 dedicated server에서 `shared_buffers`를 25% 정도로 시작하는 예를 들고, 40%를 넘기는 것이 더 낫지 않을 가능성을 언급합니다. 이 숫자는 절대 규칙이 아니라 "DBMS buffer와 OS cache가 함께 있다"는 경계를 보여 주는 자료입니다.

MySQL InnoDB는 buffer pool을 매우 중심적인 tuning 대상으로 둡니다. InnoDB buffer pool은 table과 index data를 cache하고, LRU 변형으로 page를 관리합니다. InnoDB page 기본 크기는 `innodb_page_size` 초기화 설정에 의해 정해지며, index page default는 16KB로 설명됩니다. InnoDB table data는 clustered index에 있고 secondary index에는 primary key columns가 포함됩니다. 그래서 clustered primary key 설계가 secondary index size와 lookup cost에 영향을 줍니다.

checkpoint도 제품별 용어가 다릅니다. PostgreSQL은 WAL checkpoint record와 dirty buffer flush를 기준으로 recovery start point를 줄입니다. InnoDB는 fuzzy checkpointing으로 modified pages를 작은 batch로 flush하고, crash recovery에서 checkpoint label 이후 log를 적용합니다. 두 제품 모두 "checkpoint는 log를 없애는 버튼"이 아니라 "data file과 log의 관계를 안전한 기준점으로 맞추는 작업"입니다. checkpoint가 복구 시작점을 어떻게 줄이는지는 [WAL과 PITR 문서](03-wal-redo-undo-crash-recovery-pitr.md)의 checkpoint 절에서 같은 trace로 다시 확인할 수 있습니다.

fsync와 durability 설정도 제품마다 이름이 다릅니다. PostgreSQL은 WAL sync method와 synchronous commit, checkpoint flush behavior를 봐야 하고, InnoDB는 redo log flush policy, doublewrite buffer, checkpoint, storage cache를 함께 봐야 합니다. 면접 답변에서는 특정 설정값을 외우는 것보다 "commit 응답 시점, durable log, data page flush, crash recovery의 역할"을 분리하는 것이 더 강합니다.

## 직접 재생해 보기

아래 실험은 운영 DB가 아니라 disposable 환경에서만 실행합니다. 목표는 실제 disk crash를 재현하는 것이 아니라, page와 buffer, plan, checkpoint 관측 포인트를 익히는 것입니다.

### PostgreSQL에서 page 단위 사고 재생

```sql
CREATE TABLE storage_lab (
  id integer PRIMARY KEY,
  payload text
);

INSERT INTO storage_lab
SELECT g, repeat('x', 100)
FROM generate_series(1, 10000) AS g;

ANALYZE storage_lab;

EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM storage_lab
WHERE id BETWEEN 100 AND 200;
```

PASS 신호는 `BUFFERS` output에서 shared hit/read 같은 page 접근 단서를 보는 것입니다. index scan이 나올 수도 있고, table 크기와 설정에 따라 다른 plan이 나올 수도 있습니다. 중요한 것은 row count만 보지 않고 buffer page 접근을 함께 읽는 것입니다. FAIL 신호는 "101 rows니까 101번 디스크를 읽었다"처럼 row와 I/O를 1:1로 연결하는 설명입니다.

### PostgreSQL에서 checkpoint와 dirty page 감각 재생

```sql
-- PostgreSQL 17+ current view
SELECT num_timed, num_requested, buffers_written
FROM pg_stat_checkpointer;

CHECKPOINT;

SELECT num_timed, num_requested, buffers_written
FROM pg_stat_checkpointer;
```

PASS 신호는 checkpoint 관련 counter가 변하고, dirty buffer flush가 checkpointer activity로 관측된다는 점입니다. PostgreSQL 16 이하에서는 checkpoint 관련 column이 `pg_stat_bgwriter`에 있었고, PostgreSQL 17부터는 `pg_stat_checkpointer`로 분리되었으므로 실험 환경의 version을 먼저 확인합니다. FAIL 신호는 `CHECKPOINT`를 "데이터를 백업했다"라고 설명하는 것입니다. checkpoint는 backup이 아니라 data file과 WAL recovery start point를 맞추는 기준점입니다.

### MySQL/InnoDB에서 buffer pool과 index 경로 재생

```sql
CREATE TABLE storage_lab (
  id int PRIMARY KEY,
  email varchar(100) NOT NULL,
  payload varchar(200),
  KEY idx_email(email)
) ENGINE=InnoDB;

INSERT INTO storage_lab VALUES
  (1, 'a@example.com', repeat('x', 100)),
  (2, 'b@example.com', repeat('x', 100));

EXPLAIN SELECT * FROM storage_lab WHERE email = 'a@example.com';
SHOW ENGINE INNODB STATUS\G
```

PASS 신호는 secondary index가 선택될 수 있고, InnoDB status에서 buffer pool, I/O, modified pages 같은 운영 단서를 볼 수 있다는 점입니다. 작은 table에서는 optimizer가 다른 선택을 할 수 있습니다. FAIL 신호는 secondary index entry가 항상 full row를 담는다고 설명하는 것입니다. InnoDB secondary index는 primary key columns를 담고, row data는 clustered index에서 찾습니다.

### 대용량 삭제 뒤 공간 회수 감각 재생

PostgreSQL에서 다음 흐름을 실험하면 논리 삭제와 물리 정리의 차이를 볼 수 있습니다.

```sql
CREATE TABLE delete_lab AS
SELECT g AS id, repeat('x', 200) AS payload
FROM generate_series(1, 50000) AS g;

SELECT pg_size_pretty(pg_total_relation_size('delete_lab'));

DELETE FROM delete_lab WHERE id <= 40000;

SELECT pg_size_pretty(pg_total_relation_size('delete_lab'));

VACUUM delete_lab;

SELECT pg_size_pretty(pg_total_relation_size('delete_lab'));
```

PASS 신호는 DELETE 직후 relation size가 기대만큼 바로 줄지 않을 수 있다는 점을 관찰하는 것입니다. VACUUM은 재사용 가능한 공간을 만들지만, 일반적으로 파일을 운영체제에 즉시 크게 반환하는 작업으로 이해하면 안 됩니다. FAIL 신호는 DELETE와 disk free space 반환을 같은 사건으로 설명하는 것입니다.

## 면접 꼬리 질문

1. row 하나를 읽는데 왜 page 이야기를 해야 하나?

    storage device와 DBMS buffer는 row 하나가 아니라 page 단위로 움직이는 경우가 많습니다. row 하나가 필요해도 그 row가 포함된 page를 읽고, 같은 page에 있는 여러 row가 함께 buffer에 올라옵니다. 그래서 row width, page density, locality가 query cost에 영향을 줍니다.

2. buffer pool과 OS page cache는 무엇이 다른가?

    buffer pool은 DBMS가 page와 transaction 의미를 알고 관리하는 cache입니다. dirty 여부, LSN, eviction 가능성, latch 같은 정보를 DBMS가 다룹니다. OS page cache는 file block cache입니다. OS는 DBMS row visibility나 commit 의미를 모릅니다. 두 cache는 I/O를 줄인다는 점은 비슷하지만 책임이 다릅니다.

3. commit되었는데 data page가 아직 disk에 없을 수 있나?

    있을 수 있습니다. 많은 DBMS는 commit 때 변경 page를 모두 data file에 쓰지 않고, durable WAL/redo를 먼저 보장한 뒤 dirty page flush를 나중으로 미룹니다. crash가 나면 log를 재생해 data file을 일관된 상태로 만듭니다.

4. checkpoint가 잦으면 항상 좋은가?

    아닙니다. checkpoint가 잦으면 crash recovery에서 redo해야 할 범위는 줄 수 있지만, 정상 실행 중 dirty page flush가 늘고 latency spike가 생길 수 있습니다. PostgreSQL처럼 checkpoint 뒤 첫 page 변경에서 full-page image가 다시 기록되는 구조에서는 WAL 양도 늘 수 있습니다. checkpoint는 복구 시간과 정상 I/O 사이의 균형점입니다.

5. heap과 clustered index의 핵심 차이는 무엇인가?

    heap에서는 row data가 key 순서와 독립된 heap page에 있고 index가 row 위치를 가리킵니다. clustered index에서는 cluster key의 leaf 쪽에 row data가 함께 있습니다. InnoDB는 primary key 기반 clustered index를 중심으로 설명하고, PostgreSQL 일반 table은 heap + secondary index로 설명하는 편이 정확합니다.

6. fsync가 왜 성능에 큰 영향을 주는가?

    write system call은 보통 OS cache까지의 복사만 의미할 수 있습니다. fsync 계열 동기화는 더 아래 cache와 storage에 안전하게 내리는 경계입니다. commit마다 durable log flush를 강하게 요구하면 latency가 커지고, flush를 늦추면 crash loss window가 생길 수 있습니다.

## 함정 질문

1. "SSD면 random I/O 걱정은 사라지나요?"

    SSD는 HDD보다 random I/O가 훨씬 강하지만, page miss, queue depth, write amplification, fsync latency, buffer churn은 여전히 비용입니다. index lookup이 매우 많은 query가 sequential scan보다 느릴 수 있는 반례도 남아 있습니다.

2. "buffer hit ratio가 높으면 query가 항상 빠른가요?"

    아닙니다. 어느 buffer 기준인지 봐야 하고, CPU filter cost, lock wait, sort/hash spill, network transfer, fsync, checkpoint pressure가 병목일 수 있습니다. cache hit는 중요한 단서지만 전체 성능의 충분조건은 아닙니다.

3. "checkpoint가 끝났으면 WAL이나 redo는 필요 없나요?"

    checkpoint 이전 복구 필요성은 줄지만, checkpoint 이후 변경에는 여전히 log가 필요합니다. WAL archiving, replication, PITR 같은 소비자가 있으면 더 오래 보존해야 할 수도 있습니다.

4. "InnoDB table은 모두 B-tree니까 heap이라는 말은 안 쓰나요?"

    InnoDB table data는 clustered index B-tree에 저장된다고 설명하는 편이 정확합니다. 그러나 일반적인 DB 이론이나 PostgreSQL에서는 heap table이라는 말이 중요합니다. 제품별 저장 구조를 섞지 않는 것이 핵심입니다.

5. "row를 삭제하면 disk 공간이 바로 줄어드나요?"

    보통 그렇지 않습니다. 삭제는 row나 tuple version을 더 이상 보이지 않게 만들고, 나중에 vacuum, purge, page merge, table rebuild, truncate 같은 정리 과정을 거쳐 재사용 또는 반환됩니다. 즉 논리 삭제와 물리 공간 반환은 다른 시간축입니다.

6. "OS cache가 있으니 DB buffer pool을 작게 잡아도 되나요?"

    workload와 DBMS에 따라 다르지만, 두 cache가 같은 역할을 완전히 대체한다고 보면 안 됩니다. DBMS buffer는 page dirty 상태, transaction 의미, replacement policy, read-ahead, latch와 연결됩니다. OS cache는 file block 관점입니다. 둘의 균형은 제품 권장값과 workload 관측으로 정해야 합니다.

7. "page, buffer, log, checkpoint를 키워드로만 나열하면 충분한가요?"

    충분하지 않습니다. `page, buffer frame, dirty page, checkpoint, fsync, random I/O, sequential log`는 암기용 키워드가 아니라 상태 변화의 경로입니다. 먼저 요청이나 row가 들어오고, 그 요청이 어떤 기준으로 분류되며, 그 기준이 어떤 내부 구조를 움직이고, 마지막에 어떤 관측값으로 결과를 확인하는지 말해야 합니다. 이렇게 말하면 단어는 짧아도 설명의 깊이가 생깁니다.

    작은 반례는 `commit이 성공했지만 data file page는 아직 flush되지 않은 상황`입니다. 이 반례를 처리하지 못하면 앞의 설명이 부분적으로 맞더라도 큰 설명은 틀립니다. 그래서 답변은 항상 공통 원리에서 시작하되 곧바로 범위를 좁혀야 합니다. 어느 DBMS인지, 어느 version인지, 어떤 isolation이나 일관성 요구인지, 이 작업이 운영 중 변경인지 장애 복구인지에 따라 같은 단어가 다른 위험을 만듭니다.

    DBMS별로는 PostgreSQL은 heap page와 WAL/checkpoint, InnoDB는 page/buffer pool/redo/doublewrite 같은 경계를 확인해야 합니다. 이 차이는 세부 취향이 아니라 실제 장애 대응과 설계 판단을 바꿉니다. 면접에서 제품 이름을 들었다면, 그 제품의 문서와 운영 지표로 돌아가야 합니다. 제품 이름이 없다면 일반 원리를 말하되, 일반 원리가 제품별 구현을 덮어쓴다고 말하면 안 됩니다.

    운영에서 다시 확인할 신호는 `buffer hit ratio, dirty page count, checkpoint write volume, fsync latency, random read count`입니다. 이 값들은 답변을 검증 가능한 문장으로 바꿔 줍니다. 좋은 답변은 `느립니다`, `안전합니다`, `일관됩니다`에서 멈추지 않고, 어떤 값이 어느 범위에 있으면 그렇게 판단하는지 말합니다. 반대로 그 값이 예상과 다르면 처음 세운 mental model을 다시 열어야 합니다.

    ```text
    read page -> modify buffer frame -> mark dirty -> flush log for commit -> checkpoint writes dirty page -> recovery replays if needed
    ```

    이 trace를 손으로 다시 그릴 때는 page가 어느 순간 "사용자에게 성공으로 보인 상태"가 되고, 어느 순간 "data file에 실제로 내려간 상태"가 되는지 분리해서 봅니다. Dirty page, log flush, checkpoint 사이의 간격을 그릴 수 있으면 buffer hit ratio나 checkpoint 통계도 단순 숫자가 아니라 복구 비용과 지연의 신호로 읽을 수 있습니다. 답이 막히는 화살표가 있으면 그 부분이 storage engine을 다시 확인해야 하는 지점입니다.

    면접에서는 최종적으로 이렇게 압축할 수 있습니다. DB는 row를 바로 디스크에 쓰는 것이 아니라 page 단위로 메모리에 올리고 durable log와 checkpoint로 data file을 맞춘다고 말합니다. 그 다음 꼬리 질문이 오면, 이 문장의 한 단어를 골라 작은 trace로 내려가면 됩니다. 이 방식은 외운 답을 길게 늘이는 것이 아니라, 짧은 답을 근거 있는 구조로 확장하는 방식입니다.

## 더 깊게 볼 자료

공식 문서는 page, buffer, checkpoint, durability 경계를 확인하는 1차 근거입니다. 이 저장소의 기존 `database/deep-dive` 문서는 coverage와 설명 seed로 참고하되, 제품별 사실은 공식 문서로 다시 확인합니다.

- PostgreSQL current docs
    - [Database Page Layout](https://www.postgresql.org/docs/current/storage-page-layout.html)
    - [Resource Consumption: shared_buffers](https://www.postgresql.org/docs/current/runtime-config-resource.html)
    - [Monitoring Statistics: pg_stat_checkpointer](https://www.postgresql.org/docs/current/monitoring-stats.html)
    - [WAL Configuration](https://www.postgresql.org/docs/current/wal-configuration.html)
    - [Reliability and WAL](https://www.postgresql.org/docs/current/wal-reliability.html)
    - [Index-Only Scans and Covering Indexes](https://www.postgresql.org/docs/current/indexes-index-only-scans.html)
    - [Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html)
- MySQL 8.4 Reference Manual
    - [InnoDB Architecture](https://dev.mysql.com/doc/refman/8.4/en/innodb-architecture.html)
    - [InnoDB Buffer Pool](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html)
    - [InnoDB Row Formats](https://dev.mysql.com/doc/refman/8.4/en/innodb-row-format.html)
    - [Clustered and Secondary Indexes](https://dev.mysql.com/doc/refman/8.4/en/innodb-index-types.html)
    - [The Physical Structure of an InnoDB Index](https://dev.mysql.com/doc/refman/8.4/en/innodb-physical-structure.html)
    - [InnoDB Checkpoints](https://dev.mysql.com/doc/refman/8.4/en/innodb-checkpoints.html)
    - [InnoDB Startup Options and System Variables](https://dev.mysql.com/doc/refman/8.4/en/innodb-parameters.html)
- Linux kernel and man-pages
    - [Linux Page Cache](https://docs.kernel.org/next/mm/page_cache.html)
    - [Memory Management APIs: File Mapping and Page Cache](https://docs.kernel.org/core-api/mm-api.html)
    - [Explicit volatile write back cache control](https://docs.kernel.org/block/writeback_cache_control.html)
    - [fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)
- Repo study material
    - `database/deep-dive/storage-index-optimizer/04-storage-files-pages-rows.md`
    - `database/deep-dive/storage-index-optimizer/05-buffer-pool-cache-io.md`
    - `database/deep-dive/storage-index-optimizer/06-wal-undo-redo-recovery.md`
    - `interviews/database-deep-dive/03-wal-redo-undo-crash-recovery-pitr.md`
    - `interviews/database-storage-search-nosql.md`
