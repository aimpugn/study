# 복제 지연, 백업, PITR, failover는 각각 어떤 장애 경계를 다루는가?

복제, 백업, PITR, failover는 모두 데이터 신뢰성과 관련 있지만 같은 기능이 아닙니다. 복제는 primary에서 생긴 변경을 다른 노드가 따라가게 하는 흐름입니다. 백업은 특정 시점의 데이터를 별도 매체나 별도 위치에 보존하는 흐름입니다. PITR, point-in-time recovery는 기준 백업과 연속 로그를 조합해 특정 과거 시점으로 복원하는 절차입니다. Failover는 장애가 난 쓰기 주체를 내리고 다른 노드를 새 primary로 승격하며 client routing과 old primary fencing을 바꾸는 운영 절차입니다.

이 네 가지를 한 문장으로 뭉개면 장애 대응이 위험해집니다. Replica가 있어도 실수로 지운 row는 그대로 복제될 수 있습니다. Backup이 있어도 복구 훈련을 해 보지 않으면 RTO를 모릅니다. PITR을 하려면 base backup뿐 아니라 목표 시점까지 이어지는 WAL archive나 binary log chain이 필요합니다. Failover를 하려면 promote 명령뿐 아니라 이전 primary가 다시 쓰기를 받지 못하게 막고, 새 primary 기준으로 replica와 application routing을 재구성해야 합니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [Replication은 변경 기록을 소비하는 경로다](#replication은-변경-기록을-소비하는-경로다)
    - [복제와 백업 I/O는 OS page cache, flush, block queue를 함께 소비한다](#복제와-백업-io는-os-page-cache-flush-block-queue를-함께-소비한다)
    - [Lag는 초 단위 지연이 아니라 읽기 계약의 틈이다](#lag는-초-단위-지연이-아니라-읽기-계약의-틈이다)
    - [Async replication은 RPO를 만든다](#async-replication은-rpo를-만든다)
    - [Backup은 현재 장애보다 과거 회귀를 위한 재료다](#backup은-현재-장애보다-과거-회귀를-위한-재료다)
    - [PITR은 기준 백업과 연속 로그가 모두 있어야 한다](#pitr은-기준-백업과-연속-로그가-모두-있어야-한다)
    - [Failover는 promotion, routing, fencing, divergence audit의 합이다](#failover는-promotion-routing-fencing-divergence-audit의-합이다)
    - [Replication과 WAL/PITR 문서는 연결하되 중복하지 않는다](#replication과-walpitr-문서는-연결하되-중복하지-않는다)
    - [Read routing은 데이터 신선도 요구를 먼저 분류해야 한다](#read-routing은-데이터-신선도-요구를-먼저-분류해야-한다)
    - [Replication conflict는 apply를 멈추게 할 수 있다](#replication-conflict는-apply를-멈추게-할-수-있다)
    - [Backup retention은 business recovery question으로 정해야 한다](#backup-retention은-business-recovery-question으로-정해야-한다)
    - [Failover 판단은 자동화할수록 split-brain 방어가 중요하다](#failover-판단은-자동화할수록-split-brain-방어가-중요하다)
    - [Failback은 failover보다 쉬운 후속 작업이 아니다](#failback은-failover보다-쉬운-후속-작업이-아니다)
    - [복구 훈련은 작은 샘플이 아니라 실제 병목을 드러내야 한다](#복구-훈련은-작은-샘플이-아니라-실제-병목을-드러내야-한다)
- [DBMS별 경계](#dbms별-경계)
    - [PostgreSQL](#postgresql)
    - [MySQL/InnoDB](#mysqlinnodb)
    - [Distributed SQL과 managed database](#distributed-sql과-managed-database)
- [직접 재생해 보기](#직접-재생해-보기)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
    - [장애 시나리오로 구분하기](#장애-시나리오로-구분하기)
    - [복구 검증을 문장으로 바꾸는 법](#복구-검증을-문장으로-바꾸는-법)
    - [면접에서 좋은 답과 약한 답의 차이](#면접에서-좋은-답과-약한-답의-차이)
    - [데이터 손실 범위를 설명하는 연습](#데이터-손실-범위를-설명하는-연습)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

짧게 답하면 이렇게 말할 수 있습니다. Replication은 현재의 변경을 다른 노드가 따라가게 하는 장치이고, backup/PITR은 과거의 특정 상태로 돌아가기 위한 장치이며, failover는 쓰기 주체와 client 경로를 바꾸는 운영 절차입니다. 복제 지연은 단순 성능 지표가 아니라 read-after-write와 장애 시 데이터 손실 가능성을 바꾸는 일관성 지표입니다.

가장 작은 예시는 주문 생성 뒤 replica에서 바로 읽는 상황입니다. Primary에서 `INSERT order(id=10)`이 commit되었지만, replica가 아직 그 log를 받거나 apply하지 못했다면 사용자는 방금 만든 주문을 replica read에서 보지 못합니다. 이것이 read-after-write 깨짐입니다. 통계 조회라면 허용할 수 있지만 결제 완료 화면이라면 writer로 읽거나, session token/LSN/binlog position을 이용해 충분히 따라잡은 replica로만 읽어야 합니다.

Backup은 replica와 목적이 다릅니다. Replica는 primary의 현재 변경을 따라갑니다. 누군가 실수로 `DELETE FROM orders`를 실행하고 commit하면 replica도 그 delete를 따라갈 수 있습니다. Backup은 그 이전 상태로 돌아가기 위한 재료입니다. Delayed replica는 일정 시간 늦게 따라가므로 실수 대응에 도움이 될 수 있지만, 별도의 backup 보존과 restore 검증을 완전히 대체하지 않습니다.

PITR은 `어제 밤 백업을 복원한다`보다 더 세밀합니다. 예를 들어 14:03에 잘못된 delete가 발생했다면 14:02:59까지 복원하고 싶습니다. 이를 위해서는 기준 full/base backup과 그 이후의 연속 WAL archive 또는 binary log가 필요합니다. 로그가 하나라도 빠지면 목표 시점까지 재생할 수 없습니다. WAL과 redo/binlog 역할을 섞기 쉽다면 먼저 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)을 같이 봐야 합니다.

Failover는 promote보다 넓습니다. 새 primary를 승격한 뒤 application이 어디로 쓰기를 보낼지, old primary가 살아 돌아왔을 때 split-brain을 막을지, 새 primary에 없는 transaction이 있었는지, replica들을 새 topology에 어떻게 붙일지 정해야 합니다. 그래서 고가용성 답변에는 RPO, RTO, lag, fencing, divergence check가 함께 들어가야 합니다.

## 먼저 잡아야 할 작은 모델

작은 모델은 주문 row 하나가 primary에서 commit되고 replica로 전달되는 흐름입니다.

```text
t0 primary
  INSERT INTO orders(id, user_id, status) VALUES (10, 7, 'PAID');
  COMMIT;

t1 replication stream
  변경 기록이 network를 통해 replica로 이동합니다.

t2 replica
  아직 apply 전입니다.
  SELECT * FROM orders WHERE id=10; -- no rows

t3 replica apply
  변경 기록을 재생합니다.

t4 replica
  SELECT * FROM orders WHERE id=10; -- row visible
```

이 모델에서 복제 지연은 `t1`부터 `t3`까지의 간격입니다. 물리적으로는 network transfer, receiver buffer, relay log, WAL receiver, apply worker, disk flush, conflict, long query 등 여러 원인이 있을 수 있습니다. 사용자에게는 `방금 쓴 데이터가 보이지 않는다`로 나타납니다.

복제 지연과 사용자에게 stale read처럼 보이는 문제는 한 덩어리 숫자로 보이지만, 실제로는 여러 구간으로 나뉩니다.

| 구간 | 멈추면 보이는 현상 | 먼저 확인할 것 |
| --- | --- | --- |
| primary가 log를 만들고 flush하는 구간 | commit 자체가 느려지거나 log 전송이 늦게 시작됩니다. | WAL/binlog write와 fsync 지연, checkpoint, storage latency |
| network 전송 구간 | replica가 아직 log를 받지 못합니다. | sender/receiver 위치, network throughput, replication connection |
| replica가 log를 저장하는 구간 | 받은 위치는 앞서지만 flush 위치가 뒤처질 수 있습니다. | relay/WAL receiver flush, replica disk I/O |
| replica apply 구간 | 받았지만 replay/executed 위치가 뒤처집니다. | applier thread, long query conflict, lock conflict, schema mismatch |
| read routing 구간 | replica는 따라잡았지만 application이 stale replica로 보낼 수 있습니다. 이 경우는 복제 지연 자체가 아니라 routing 오류에 가깝습니다. | router policy, session token, lag threshold |

이 표를 사용하면 `lag가 10초입니다`라는 숫자를 원인으로 착각하지 않게 됩니다. 같은 stale read라도 network 병목, replica disk 병목, apply conflict, 잘못된 read routing은 서로 다른 장애입니다.

Backup 모델은 다릅니다.

```text
day 1 00:00 base backup
  orders table snapshot preserved

day 1 00:00-14:02 archived logs
  정상 변경 기록 보존

day 1 14:03 accidental delete
  DELETE FROM orders; COMMIT;

desired restore
  base backup + logs until 14:02:59
  delete 직전 상태로 복원
```

이 흐름에서 replica는 delete를 따라갈 수 있지만 PITR은 delete 직전으로 돌아갑니다. 따라서 복제와 백업은 서로 보완 관계입니다. 복제는 장애 시 빠른 전환과 read scale에 좋고, backup/PITR은 논리적 실수와 과거 시점 복구에 필요합니다.

Failover 모델은 또 다릅니다.

```text
before
  primary P accepts writes
  replica R follows P
  clients write to P

disaster
  P unreachable or unsafe

action
  promote R to new primary
  route clients to R
  fence P so it cannot accept writes
  rebuild replicas from R

audit
  compare last acknowledged commit / LSN / GTID
  disclose possible data loss window
```

여기서 `promote R`은 한 줄이지만 운영 절차 전체는 여러 단계입니다. Old primary fencing이 없으면 P가 network partition 뒤 살아 돌아와 별도 writes를 받을 수 있습니다. Application routing이 늦으면 일부 client가 여전히 P로 쓸 수 있습니다. Replica lag가 컸으면 R에는 P의 마지막 commit 일부가 없을 수 있습니다.

## 깊은 메커니즘

### Replication은 변경 기록을 소비하는 경로다

DBMS는 보통 변경 자체를 logical 또는 physical log 형태로 남깁니다. PostgreSQL은 WAL을 기반으로 streaming replication과 archive recovery를 구성합니다. MySQL은 binary log를 source에서 남기고 replica가 relay log로 받아 apply합니다. 용어는 다르지만 공통 흐름은 같습니다.

```text
primary transaction commit
  -> durable change log record
  -> sender transmits log
  -> replica receives log
  -> replica replays/applies log
  -> replica exposes changed state to reads
```

이 흐름에서 commit과 replica apply는 같은 순간이 아닐 수 있습니다. 비동기 복제에서는 primary가 replica apply를 기다리지 않고 client에게 commit success를 반환할 수 있습니다. 동기 또는 semi-sync 계열은 더 강한 확인을 요구하지만 latency와 availability tradeoff가 생깁니다. Consensus 기반 distributed SQL은 또 다른 모델로, quorum commit을 통해 replica agreement를 commit 경로에 넣을 수 있습니다. 이 문서에서는 전통적인 primary-replica 복제와 backup/failover 경계를 중심으로 다루고, consensus와 shard/partition 경계는 [파티션, 샤딩, 분산 SQL](10-partition-sharding-distributed-sql.md)로 넘깁니다.

### 복제와 백업 I/O는 OS page cache, flush, block queue를 함께 소비한다

복제 지연과 백업 시간을 설명할 때는 DB 로그 위치만 보시면 부족합니다. Primary는 commit을 위해 WAL이나 binary log를 쓰고 flush해야 합니다. Replica는 그 로그를 받고, 파일에 기록하고, 다시 apply하면서 data page를 읽고 씁니다. Backup은 data file을 대량으로 읽고, object storage나 다른 disk로 쓰며, PITR을 위해 WAL/binlog archive도 계속 보존합니다. 이 모든 작업은 DBMS 내부 thread만의 일이 아니라 운영체제의 memory manager, filesystem, block layer, device driver, storage queue를 함께 사용합니다.

```text
primary commit
  WAL/binlog write + fsync
        |
        v
network transfer to replica
        |
        v
replica receive/flush/apply
        |
        v
data page read/write, checkpoint/restartpoint

backup at same time
  reads many data files
  writes backup stream/archive
  competes for page cache, block queues, disk bandwidth
```

이 그림은 `lag가 늘었습니다`라는 증상을 더 잘 나눠 줍니다. Network가 느려서 log를 못 받는 지연, replica가 받은 log를 disk에 flush하지 못하는 지연, apply worker가 lock conflict나 long query 때문에 못 따라가는 지연, backup이 같은 storage bandwidth를 써서 생기는 지연은 처방이 다릅니다. PostgreSQL의 sent/write/flush/replay LSN이나 MySQL의 retrieved/executed GTID 차이를 볼 때도 이 단계 구분을 붙여야 합니다. OS 관점에서는 page cache가 memory pressure를 받고 dirty page writeback이 밀리는지, block queue가 포화되었는지, storage가 fsync 완료를 늦게 돌려주는지를 같이 봅니다.

백업이 복제와 같은 시간대에 돌 때는 I/O의 방향도 함께 봅니다.

```text
primary
  foreground commit: WAL/binlog fsync
  checkpoint: dirty data page flush
  backup: data file read + archive write

replica
  receive: log write/flush
  apply: data page read/write
  restartpoint/checkpoint: dirty page flush
  read traffic: user query page read
```

이 흐름에서 backup은 단순히 "읽기 작업"이 아닐 수 있습니다. 압축, 암호화, object storage 업로드, WAL/binlog archive 보존, checksum 검증이 붙으면 CPU, network, disk write도 함께 씁니다. 그래서 복제 지연이 백업 시간대에만 증가한다면 DBMS 지표와 함께 OS page cache, dirty writeback, block queue, network egress를 같이 확인해야 합니다.

### Lag는 초 단위 지연이 아니라 읽기 계약의 틈이다

Replication lag를 `몇 초 늦습니다`라고만 말하면 부족합니다. Lag는 어떤 사용자가 어떤 read를 했을 때 어떤 history를 보는가를 바꿉니다. 방금 주문한 사용자는 주문 상세 페이지에서 자기 주문을 기대합니다. Replica가 늦으면 `주문이 없습니다`가 나오고, 사용자는 결제가 실패했다고 생각할 수 있습니다.

Read-after-write를 지키는 방법은 여러 가지입니다. 가장 단순한 방법은 쓰기 직후 일정 시간 또는 같은 session의 중요한 read를 primary로 보내는 것입니다. 더 정교한 방법은 commit LSN, GTID, binlog position, session consistency token을 기록하고, replica가 그 위치까지 따라잡았는지 확인한 뒤 읽는 것입니다. Lag threshold가 너무 크면 replica를 read pool에서 빼는 routing도 필요합니다.

```text
write response includes commit_position = LSN 0/500ABCD
next read asks router:
  find replica whose replay_lsn >= 0/500ABCD
  if none exists within timeout, read from primary
```

이 방식은 제품별로 구현 세부가 다릅니다. PostgreSQL에서는 LSN과 replay location, MySQL에서는 GTID set이나 binlog position을 볼 수 있습니다. 중요한 것은 `어떤 위치까지 반영되면 내 write가 보인다고 말할 수 있는가`입니다.

Application 입장에서는 이 위치 값을 사용자 경험으로 번역해야 합니다.

| read 상황 | 허용 가능한 신선도 | routing 예시 | 사용자에게 보이는 위험 |
| --- | --- | --- | --- |
| 결제 직후 주문 상세 | 방금 commit한 위치 이상 | primary 또는 caught-up replica | 결제했는데 주문이 없다고 보일 수 있습니다. |
| 본인 게시글 작성 직후 | session write 위치 이상 | session stickiness 또는 LSN/GTID token | 사용자가 재등록을 시도할 수 있습니다. |
| 관리자 통계 | 몇 분 지연 허용 가능 | lag threshold 이하 replica | 숫자가 늦을 수 있으므로 freshness 표시가 필요합니다. |
| 대량 리포트 | eventual consistency 허용 | replica 또는 analytical store | 최신 거래 누락을 업무가 허용하는지 명시해야 합니다. |

이 분류가 있으면 replica routing은 단순 부하 분산이 아니라 read 계약 선택이 됩니다. 모든 read를 primary로 보내면 일관성은 단순하지만 확장성이 줄고, 모든 read를 replica로 보내면 사용자 직후 읽기에서 신뢰가 깨질 수 있습니다.

### Async replication은 RPO를 만든다

RPO, recovery point objective는 장애 때 잃을 수 있는 데이터 범위에 대한 목표입니다. 비동기 복제에서 primary가 commit success를 반환했지만 replica가 아직 log를 받지 못한 상태로 primary가 영구 장애를 만나면, replica를 promote해도 그 transaction은 없을 수 있습니다. 이때 RPO는 0이 아닙니다.

```text
t0 client commit success on primary
t1 primary has not sent log to replica
t2 primary disk lost
t3 replica promoted

new primary history does not include t0 transaction
```

Semi-sync나 synchronous replication은 이 창을 줄일 수 있습니다. 하지만 더 많은 replica 확인을 commit path에 넣을수록 write latency가 늘고, replica 장애나 network partition 때 primary availability가 떨어질 수 있습니다. 면접에서 `동기로 하면 안전합니다`라고만 답하면 tradeoff를 놓칩니다. 안전성, latency, availability, 운영 복잡도를 같이 말해야 합니다.

### Backup은 현재 장애보다 과거 회귀를 위한 재료다

Backup은 `복사본 하나`가 아니라 복원 가능한 재료와 절차의 묶음입니다. Full backup, incremental backup, WAL archive, binary log, snapshot, object storage lifecycle, encryption key, restore credentials, schema version, application compatibility가 함께 닫혀야 합니다. Backup 파일이 존재한다는 사실과 서비스가 복구된다는 사실은 다릅니다.

복제와 비교하면 목적이 선명해집니다.

```text
replica
  primary의 현재 변경을 따라갑니다.
  장애 전환과 read scale에 유용합니다.
  실수 delete도 빠르게 따라갈 수 있습니다.

backup/PITR
  과거 특정 상태로 돌아가기 위한 재료입니다.
  실수 delete, corruption 발견, ransomware, 잘못된 migration 대응에 필요합니다.
  restore test 없이는 실제 RTO를 알 수 없습니다.
```

운영에서 backup 검증은 restore를 포함해야 합니다. Backup file checksum만 맞아도, 복원 대상 DB version이 다르거나 extension이 없거나, encryption key가 없거나, log chain이 끊겼거나, application migration과 시간이 맞지 않으면 서비스 복구는 실패합니다. 복구 절차가 실제 운영 runbook으로 내려가면 [운영, 보안, 트러블슈팅](13-operations-security-troubleshooting.md)의 관측과 권한 경계까지 함께 확인해야 합니다.

### PITR은 기준 백업과 연속 로그가 모두 있어야 한다

PostgreSQL의 PITR은 base backup과 archived WAL을 이용해 목표 시점까지 recovery합니다. MySQL은 full backup과 binary log replay로 유사한 목표를 달성할 수 있습니다. 제품별 명령은 다르지만 구조는 같습니다.

```text
base backup at T0
log segment 1: T0 -> T1
log segment 2: T1 -> T2
log segment 3: T2 -> T3
recovery target: T2.5

필요한 것:
  base backup
  log segment 1, 2, 3 중 T2.5까지의 연속 구간
  target time 또는 target LSN/transaction
```

로그가 하나 빠지면 그 뒤 시점으로 갈 수 없습니다. Backup 보존 정책도 이 사실을 기준으로 세워야 합니다. Base backup은 있는데 WAL archive retention이 짧으면 오래된 시점으로 복구할 수 없습니다. 반대로 WAL archive는 있는데 기준 backup이 없으면 replay 시작점이 없습니다.

PITR 뒤에는 timeline/history 개념도 중요합니다. 특정 과거 시점으로 돌아간 뒤 새 write를 받으면 원래 history와 다른 가지가 생깁니다. PostgreSQL은 timeline으로 이런 history를 구분합니다. 이 경계를 모르면 복구 후 어떤 WAL을 이어 적용할지, 기존 replica를 그대로 붙일 수 있는지 혼동합니다.

로그 체인이 끊기는 실패는 작은 timeline으로 보면 바로 드러납니다.

```text
base backup B0 at 00:00

WAL/binlog:
  S1 00:00 -> 01:00  exists
  S2 01:00 -> 02:00  missing
  S3 02:00 -> 03:00  exists

target recovery time: 02:30

result:
  B0 + S1까지는 재생할 수 있습니다.
  S2가 없으므로 S3로 건너뛰어 02:30 상태를 만들 수 없습니다.
```

이 예시는 backup retention을 "며칠 보관"으로만 정하면 부족한 이유를 보여 줍니다. 복구 가능한 시간 범위는 기준 백업과 그 이후 연속 로그가 함께 결정합니다. 둘 중 하나만 보존되면 원하는 목표 시점으로 돌아갈 수 없습니다.

### Failover는 promotion, routing, fencing, divergence audit의 합이다

Failover를 한 줄 명령으로 이해하면 운영에서 크게 다칩니다. 새 primary 승격은 시작일 뿐입니다. Client write endpoint를 바꾸고, connection pool과 DNS/load balancer를 갱신하고, old primary가 살아 돌아와도 write를 받지 못하게 fencing해야 합니다. Replica들은 새 primary를 따라가도록 재구성해야 합니다.

Split-brain은 가장 위험한 실패입니다.

```text
network partition
  app group A sees old primary P
  app group B sees promoted primary R
both accept writes

result
  two histories diverge
  automatic merge may be impossible
```

이를 막으려면 health check만으로 primary를 판단하지 않고, quorum, fencing token, STONITH, cloud volume detach, orchestrator lease 같은 운영 장치가 필요할 수 있습니다. 어떤 장치를 쓰는지는 환경별로 다르지만, 답변에는 `old primary를 어떻게 쓰기 불능으로 만들었는가`가 들어가야 합니다.

Divergence audit도 필요합니다. Failover 시점에 old primary의 last committed position과 new primary의 replay/applied position을 비교합니다. Async replication이면 데이터 손실 가능성을 disclose해야 합니다. MySQL GTID, PostgreSQL LSN/timeline 같은 위치 식별자가 이때 쓰입니다.

Failover runbook은 보통 다음 상태 전이를 모두 닫아야 합니다.

```text
detect
  primary가 죽었는지, 느린 것인지, network partition인지 판정합니다.

choose candidate
  가장 최신이고 일관된 replica를 고릅니다.

fence old primary
  old primary가 다시 write를 받지 못하게 만듭니다.

promote
  candidate를 new primary로 올립니다.

route
  application write endpoint와 connection pool을 new primary로 돌립니다.

audit
  old/new log position을 비교하고 손실 가능 commit 범위를 공개합니다.

rebuild
  나머지 replica를 new primary 기준으로 다시 붙입니다.
```

이 순서에서 `promote`만 성공해도 failover가 끝난 것은 아닙니다. 특히 detect와 fence가 약하면 split-brain 위험이 남고, route가 약하면 일부 client가 여전히 old endpoint로 write를 보낼 수 있습니다.

### Replication과 WAL/PITR 문서는 연결하되 중복하지 않는다

WAL 자체의 record 구조, redo/undo, checkpoint, crash recovery는 별도 WAL deep-dive에서 다루는 것이 좋습니다. 이 문서에서는 WAL을 `여러 소비자가 읽는 변경 기록`으로만 봅니다. 같은 WAL이 crash recovery, streaming replication, archive recovery/PITR에서 쓰일 수 있지만, 소비 목적과 보장 경계는 다릅니다.

```text
crash recovery
  같은 primary가 crash 전 committed state를 복구합니다.

streaming replication
  다른 node가 primary 변경을 따라갑니다.

archive recovery / PITR
  backup에서 시작해 과거 목표 시점까지 log를 재생합니다.
```

이 세 경로를 구분하면 `WAL이 있으니 backup이 필요 없나요?`, `replica가 있으니 PITR이 되나요?`, `failover하면 crash recovery도 끝난 건가요?` 같은 함정 질문에 흔들리지 않습니다. WAL record와 checkpoint, crash recovery 자체는 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)의 주제이고, 이 문서는 그 로그를 복제·복구·전환 절차가 어떻게 소비하는지에 집중합니다.

### Read routing은 데이터 신선도 요구를 먼저 분류해야 한다

모든 read가 같은 freshness를 요구하지 않습니다. 결제 직후 주문 상세, 비밀번호 변경 직후 로그인, 방금 쓴 게시글 미리보기는 read-after-write가 중요합니다. 반면 하루 단위 통계, 관리자 대시보드, 검색 색인 갱신 상태는 몇 초 또는 몇 분 지연을 허용할 수 있습니다. Replica routing은 이 차이를 코드로 표현해야 합니다.

```text
read class A: must see own write
  route to primary or caught-up replica
  carry commit position in session

read class B: must be within 5 seconds
  route to replica if lag <= 5s
  otherwise primary or degraded response

read class C: eventually consistent report
  route to replica or analytical store
  show data freshness timestamp
```

이 분류가 없으면 replica를 붙인 뒤 장애가 UX로 나타납니다. 사용자는 결제 성공 직후 주문이 없다는 화면을 보고 재결제를 시도할 수 있습니다. 운영자는 replica lag 알람을 성능 경고로만 보고 넘기다가 failover 때 데이터 손실 창을 과소평가할 수 있습니다.

### Replication conflict는 apply를 멈추게 할 수 있다

Replica는 primary log를 단순히 받기만 하는 것이 아니라 적용해야 합니다. 적용 중 오류가 나면 replication이 멈추거나 지연될 수 있습니다. MySQL에서는 duplicate key, missing row, schema mismatch, non-deterministic statement, 권한/DDL 문제 등이 applier error로 나타날 수 있습니다. PostgreSQL physical streaming은 같은 cluster binary state를 replay하는 성격이 강하지만, hot standby query conflict나 recovery conflict가 read-only query 취소로 나타날 수 있습니다.

따라서 lag를 볼 때는 network만 의심하지 않습니다. Receiver가 log를 받는지, flush했는지, replay했는지 단계별로 나눕니다. PostgreSQL의 sent/write/flush/replay LSN 차이, MySQL의 retrieved/executed GTID 차이는 이 구분을 돕습니다. `받았지만 적용하지 못하는 지연`과 `아예 받지 못한 지연`은 원인과 대응이 다릅니다.

### Backup retention은 business recovery question으로 정해야 한다

Backup 보존 기간은 storage 비용만으로 정하면 안 됩니다. 사용자가 한 달 뒤 잘못된 데이터 import를 발견할 수 있는가, 법적 감사가 몇 년치 데이터를 요구하는가, ransomware 감염을 언제 알 수 있는가, schema migration 뒤 이전 version으로 복원해야 할 수 있는가를 묻습니다. Recovery question이 보존 정책을 정합니다.

```text
question: 7일 전 잘못된 batch가 고객 등급을 바꿨습니다.
need: 7일 전 이전 상태와 이후 정상 transaction을 비교할 수 있는 backup/log

question: 오늘 14:03 delete를 되돌리고 싶습니다.
need: delete 직전까지 PITR 가능한 base backup + continuous log

question: primary region 전체가 사라졌습니다.
need: cross-region copy, credentials, network, application endpoint plan
```

Backup encryption key도 복구 재료입니다. Backup file과 WAL archive가 있어도 key가 없으면 복구할 수 없습니다. 권한과 secret rotation, object storage lifecycle, immutability policy까지 포함해야 합니다.

### Failover 판단은 자동화할수록 split-brain 방어가 중요하다

자동 failover는 사람이 느리게 판단하는 시간을 줄여 RTO를 낮출 수 있습니다. 하지만 잘못된 자동 failover는 split-brain을 빠르게 만듭니다. Network partition 때문에 monitor는 primary가 죽었다고 판단했지만 primary는 일부 client에게 여전히 reachable할 수 있습니다. 이때 새 primary를 올리면 두 writer가 생깁니다.

안전한 자동화는 보통 quorum 판단, lease, fencing, cloud control plane action, storage detach, old primary kill 같은 장치를 함께 씁니다. 단순 TCP health check 하나로 primary 생사를 판단하지 않습니다. 자동화가 어려운 환경이라면 failover를 수동으로 하되 runbook과 drill로 시간을 줄이는 편이 나을 수 있습니다. 중요한 것은 자동/수동 자체가 아니라 `누가 쓰기 권한을 유일하게 갖는가`가 증명되는 것입니다.

### Failback은 failover보다 쉬운 후속 작업이 아니다

장애가 끝난 뒤 old primary를 다시 붙이는 작업을 failback이라고 부를 수 있습니다. 이 작업은 단순히 old primary를 켜는 일이 아닙니다. Old primary가 장애 직전 어떤 log position까지 commit했는지, new primary history와 divergence가 있는지, old primary 데이터를 버리고 re-clone해야 하는지, timeline/GTID 관계가 이어지는지 확인해야 합니다.

Async failover에서 old primary가 new primary보다 앞선 transaction을 일부 갖고 있을 수도 있습니다. 그 transaction을 수동으로 복구할지, 손실로 인정할지, application-level reconciliation으로 맞출지 결정해야 합니다. 이 판단은 데이터 손실 가능성을 숨기지 않고 RPO 관점으로 공개해야 합니다. 좋은 runbook은 failover보다 failback과 postmortem까지 포함합니다.

### 복구 훈련은 작은 샘플이 아니라 실제 병목을 드러내야 한다

Restore drill을 너무 작은 샘플 DB로만 하면 실제 병목을 놓칩니다. 수 TB 백업을 object storage에서 내려받는 시간, WAL/binlog replay 속도, index rebuild, checksum 검증, application warm-up, DNS TTL, connection pool 재연결 시간이 RTO를 결정합니다. 작은 샘플은 절차 검증에는 좋지만, 시간 목표 검증에는 실제 크기에 가까운 rehearsal이 필요합니다.

복구 훈련 결과는 `성공` 한 단어가 아니라 숫자와 실패 지점이어야 합니다. Base backup 다운로드 40분, WAL replay 25분, schema migration compatibility check 5분, application smoke test 10분처럼 나누어야 다음 개선 지점이 보입니다. 한 번도 복원하지 않은 backup은 아직 신뢰 가능한 backup이 아니라 backup 후보에 가깝습니다.

## DBMS별 경계

### PostgreSQL

PostgreSQL은 WAL을 중심으로 streaming replication, log shipping, continuous archiving, PITR을 구성합니다. Streaming replication은 primary의 WAL record를 standby가 받아 replay하는 흐름입니다. Standby는 read-only query를 받을 수 있지만 replay lag와 query conflict가 있을 수 있습니다. Continuous archiving은 WAL segment를 안전한 위치에 보관해 base backup과 함께 PITR을 가능하게 합니다.

PostgreSQL failover에서는 timeline이 중요합니다. Standby를 promote하면 새 timeline이 생길 수 있고, 기존 primary나 다른 standby를 새 timeline에 어떻게 합류시킬지 결정해야 합니다. Replication slot은 WAL 보존을 도와주지만, consumer가 멈추면 WAL이 쌓여 disk를 압박할 수 있습니다. 따라서 slot lag도 운영 지표입니다.

관측값으로는 `pg_stat_replication`, sent/write/flush/replay LSN, `pg_last_wal_replay_lsn()`, replay delay, replication slot retained WAL, archive command success, restore command success, backup label과 recovery target을 봅니다.

### MySQL/InnoDB

MySQL replication은 binary log와 relay log, replica SQL applier를 중심으로 설명합니다. GTID를 쓰면 transaction identity와 failover/rejoin을 관리하기 쉬워집니다. Statement-based, row-based, mixed logging 같은 binlog format 차이도 복제 정확성과 디버깅에 영향을 줍니다. 현대 OLTP에서는 row-based가 예측 가능할 때가 많지만, 실제 설정을 확인해야 합니다.

MySQL에서 InnoDB redo log와 binary log는 역할이 다릅니다. Redo log는 InnoDB crash recovery와 durability에 중요하고, binary log는 replication과 point-in-time recovery에 중요합니다. Commit path에서는 둘 사이의 정합성을 맞추기 위해 two-phase commit 성격의 조정이 필요합니다. 면접에서 redo와 binlog를 섞으면 replication/PITR 답변이 흔들립니다. 이 차이는 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)의 로그 역할 구분과 함께 보면 더 선명합니다.

관측값으로는 replica lag, relay log position, retrieved/executed GTID set, replication applier errors, delayed replication 설정, `SHOW REPLICA STATUS` 계열 출력, backup tool의 binlog coordinates를 봅니다.

### Distributed SQL과 managed database

Distributed SQL은 단순 async primary-replica와 다를 수 있습니다. Raft/Paxos 계열 consensus를 commit path에 넣으면 leader와 quorum replica가 합의한 뒤 commit을 인정합니다. 이 경우 failover와 replication lag의 의미가 전통적인 async replica와 다르게 나타납니다. 하지만 consensus는 network round trip과 quorum availability 비용을 만듭니다.

Managed database는 자동 backup, PITR, multi-AZ failover를 제공할 수 있습니다. 그러나 제공 기능의 이름만 믿으면 안 됩니다. Retention 기간, 복구 가능한 최소 단위, cross-region replica의 lag, failover 때 DNS 전파 시간, read endpoint behavior, backup encryption key, parameter group과 extension 복구 범위를 확인해야 합니다. Vendor 기능은 편리하지만 운영 계약을 대신 이해해 주지는 않습니다.

## 직접 재생해 보기

1. Read-after-write 깨짐을 trace로 재생합니다.

    실제 replica가 없다면 종이에라도 commit position과 apply position을 나누어 그립니다. 로컬에서 replica를 구성할 수 있다면 primary insert 직후 replica read를 반복하고, replica를 의도적으로 지연시키거나 무거운 query로 apply를 늦춥니다.

    PASS 신호는 `primary commit은 성공했지만 replica replay position이 아직 commit position보다 작아서 row가 보이지 않는다`고 설명하는 것입니다. FAIL 신호는 replica에서 안 보이는 현상을 insert 실패로 오해하는 것입니다.

2. Lag-aware routing 규칙을 만듭니다.

    ```text
    read type: user just wrote order
      required position: commit LSN/GTID from write response
      route: replica whose apply position >= required position, else primary

    read type: daily statistics
      required freshness: 5 minutes
      route: any replica with lag <= 5 minutes
    ```

    PASS 신호는 read마다 freshness 요구가 다르다는 점을 routing policy로 표현하는 것입니다. 모든 read를 primary로 보내면 단순하지만 scale 이점이 줄고, 모든 read를 replica로 보내면 일관성 사고가 납니다.

3. Backup restore drill을 수행합니다.

    작은 DB에서 base backup을 만들고, 몇 개 transaction을 실행한 뒤, 특정 transaction 직전으로 복구하는 절차를 문서화합니다. 실제 명령은 PostgreSQL/MySQL 도구와 버전에 맞춥니다. PASS 신호는 복구 후 row count와 대표 query가 목표 시점과 일치하고, 복구에 걸린 시간이 기록되는 것입니다. FAIL 신호는 backup file 생성만 확인하고 restore를 해 보지 않는 것입니다.

4. Log chain 누락을 일부러 상상합니다.

    PITR에 필요한 WAL segment 또는 binlog file 중 하나가 없을 때 어느 시점까지 복구 가능한지 적습니다. PASS 신호는 `base backup 이후 목표 시점까지 연속 로그가 필요하다`는 문장으로 설명하는 것입니다. 이 실험은 실제 파일을 삭제하라는 뜻이 아니라, restore plan에서 누락이 어떤 실패를 만드는지 확인하는 것입니다.

5. Failover runbook을 검증합니다.

    Runbook에는 장애 판정 기준, promote 명령, old primary fencing, client routing 변경, new primary health check, replica 재구성, divergence audit, rollback 또는 rejoin 전략이 있어야 합니다. PASS 신호는 누가 어떤 순서로 무엇을 확인하는지 한 줄씩 실행 가능하게 적혀 있는 것입니다. FAIL 신호는 `replica promote`만 있고 이전 primary 처리와 client endpoint가 없는 것입니다.

## 면접 꼬리 질문

- Replica가 있으면 backup이 필요 없나요?

    아닙니다. Replica는 현재 변경을 따라가므로 실수 삭제나 잘못된 migration도 따라갈 수 있습니다. Backup과 PITR은 과거 시점으로 돌아가기 위한 재료입니다. Replica는 availability와 read scale에 도움을 주지만 backup을 대체하지 않습니다.

- 비동기 복제에서 commit 성공은 replica 반영을 뜻하나요?

    보통 아닙니다. Primary가 commit success를 반환한 뒤 replica가 나중에 log를 받고 apply할 수 있습니다. 이 창이 replication lag이며, 장애 시 RPO와 read-after-write를 바꿉니다.

- Read-after-write를 어떻게 보장하나요?

    중요한 read는 primary로 보내거나, write commit position을 session에 저장하고 그 위치까지 따라잡은 replica로만 보냅니다. Lag threshold routing, sticky session, causal token 같은 전략을 쓸 수 있습니다. 비용은 primary load와 latency입니다.

- Delayed replica는 PITR을 대체하나요?

    완전한 대체는 아닙니다. Delayed replica는 실수 직후 빠른 대응에 도움을 줄 수 있지만 retention window 밖의 복구, corruption, storage failure, backup compliance, restore drill을 대신하지 않습니다. Backup과 log archive가 여전히 필요합니다.

- Failover 후 old primary는 왜 위험한가요?

    Old primary가 살아 돌아와 write를 받으면 split-brain이 됩니다. 두 primary가 서로 다른 history를 만들면 자동 merge가 어렵거나 불가능합니다. Fencing과 routing 통제가 failover의 핵심입니다.

- RPO와 RTO는 어떻게 설명하나요?

    RPO는 장애 때 잃을 수 있는 데이터 범위에 대한 목표이고, RTO는 서비스를 다시 사용할 수 있게 만드는 시간 목표입니다. Async replication lag는 RPO에, restore time과 failover runbook은 RTO에 직접 영향을 줍니다.

## 함정 질문

- "복제가 있으니 고가용성과 데이터 보호가 모두 해결됩니다"라고 말하는 것

    복제는 current change를 따라가는 장치입니다. 논리적 실수, corruption, delayed discovery, ransomware 같은 문제에는 backup/PITR이 필요합니다. Availability와 recoverability를 나누어야 합니다.

- "Replication lag는 단지 성능 지연입니다"라고 말하는 것

    Lag는 사용자가 어떤 history를 읽는지 바꾸는 일관성 문제입니다. Read-after-write, failover data loss, report freshness, delayed alert가 모두 lag와 연결됩니다.

- "PITR은 full backup을 복원하는 것입니다"라고 말하는 것

    Full/base backup은 시작점일 뿐입니다. 목표 시점까지의 연속 WAL/binlog가 필요합니다. Log chain이 끊기면 목표 시점으로 갈 수 없습니다.

- "Promote하면 failover가 끝납니다"라고 말하는 것

    Promote는 새 primary 후보를 쓰기 가능하게 하는 단계입니다. Client routing, old primary fencing, topology rebuild, divergence audit, application health check가 닫혀야 failover가 끝납니다.

- "동기 복제를 켜면 단점 없이 RPO 0이 됩니다"라고 말하는 것

    동기 확인 범위를 commit path에 넣으면 latency와 availability tradeoff가 생깁니다. Replica 장애나 network partition 때 쓰기를 계속 받을지 멈출지 정책을 정해야 합니다.

### 장애 시나리오로 구분하기

이 주제는 작은 장애 시나리오로 나누어 말하면 훨씬 선명해집니다. 첫 번째는 `primary는 살아 있지만 replica가 늦는 상황`입니다. 이때 write는 성공하고 서비스는 대부분 정상처럼 보입니다. 하지만 replica로 간 read는 과거를 읽습니다. 해결은 backup이나 failover가 아니라 read routing과 lag-aware decision입니다. 사용자가 방금 쓴 데이터는 primary로 읽거나, replica가 특정 LSN/GTID 이상을 적용할 때까지 기다려야 합니다.

두 번째는 `primary가 죽었고 replica는 조금 늦은 상황`입니다. 이때 질문은 failover와 RPO입니다. 새 primary로 올릴 replica가 마지막 commit까지 받았는지 확인해야 하고, 받지 못했다면 어떤 transaction이 사라질 수 있는지 사용자와 시스템이 감당할 수 있어야 합니다. 비동기 복제에서는 이 위험이 구조적으로 남습니다. 동기 복제는 위험을 줄이지만 primary commit latency와 장애 시 availability 비용을 만듭니다.

세 번째는 `운영자가 잘못된 DELETE를 실행한 상황`입니다. 이때 replica는 보통 보호막이 아닙니다. DELETE가 정상 commit으로 복제되기 때문입니다. 필요한 것은 backup, PITR, delayed replica, audit log, 권한 통제입니다. 목표 시점은 DELETE 직전이어야 하고, 그 시점까지 log chain이 이어져 있어야 합니다. 복구 뒤에는 삭제된 row만 되돌릴지 전체 DB를 특정 시점으로 되돌릴지 결정해야 합니다. 전체 DB rollback은 다른 정상 transaction까지 되돌릴 수 있으므로 업무 영향 분석이 필요합니다.

네 번째는 `failover 후 old primary가 다시 살아나는 상황`입니다. 이때 가장 위험한 것은 두 노드가 모두 write를 받는 것입니다. Split-brain이 발생하면 나중에 두 history를 자동으로 합치기 어렵습니다. 따라서 old primary fencing은 선택 절차가 아니라 failover의 핵심입니다. Cloud managed service가 이 과정을 대신해 주더라도, 애플리케이션 connection이 어디로 붙는지, DNS나 proxy cache가 얼마나 남는지, connection pool이 old endpoint를 붙잡고 있지 않은지 확인해야 합니다.

```text
lag only           -> read routing / session consistency / lag threshold
primary lost       -> failover / RPO judgement / latest safe replica
bad data committed -> backup or PITR / audit / selective restore
old primary alive  -> fencing / split-brain prevention / routing verification
```

이 네 줄을 외우라는 뜻은 아닙니다. 같은 고가용성이라는 단어 아래 실제로는 서로 다른 질문이 숨어 있음을 보여 주는 지도입니다. 면접에서 상황이 주어지면 먼저 어떤 줄에 해당하는지 분류하고, 그 다음 필요한 도구를 말하면 됩니다.

### 복구 검증을 문장으로 바꾸는 법

복구를 검증했다는 말은 명령이 성공했다는 뜻만으로는 부족합니다. `restore completed` 로그가 있어도 application이 schema version mismatch로 뜨지 않거나, secret 접근이 안 되거나, 권한이 빠졌거나, 특정 index가 없어 핵심 query가 느려질 수 있습니다. 따라서 복구 검증은 DBMS 레벨과 애플리케이션 레벨을 나누어 닫습니다.

DBMS 레벨에서는 목표 시점, timeline 또는 binlog position, row count, checksum, 핵심 table의 referential integrity를 봅니다. 애플리케이션 레벨에서는 로그인, 주문 조회, 결제 상태 조회, write smoke test, background worker 재시작, read/write routing을 확인합니다. 보안 레벨에서는 backup decrypt key, least privilege 계정, audit trail, restored environment의 외부 발송 차단을 봅니다. 복구 환경에서 실수로 실제 고객에게 메시지를 보내면 복구 훈련 자체가 사고가 될 수 있기 때문입니다.

```text
restore DB -> verify data shape -> boot application -> verify critical read/write -> verify routing and side effects are safe
```

RTO/RPO도 이 검증을 통해 숫자로 바꿔야 합니다. `백업은 매일 있습니다`는 RPO를 정확히 말하지 않습니다. 마지막 full backup 이후 WAL/binlog archive가 있다면 RPO는 더 작을 수 있고, archive가 끊기면 더 커질 수 있습니다. `자동 failover가 됩니다`도 RTO를 정확히 말하지 않습니다. 감지 시간, promotion 시간, DNS/proxy 전환, application reconnect, warmup, 검증 시간이 모두 들어갑니다.

### 면접에서 좋은 답과 약한 답의 차이

약한 답은 보통 도구 이름으로 끝납니다. `replica를 둡니다`, `backup을 합니다`, `failover를 구성합니다`가 그렇습니다. 좋은 답은 도구가 답하는 질문을 먼저 말합니다. `replica는 읽기 부하와 장애 대응에 도움을 주지만 lag 때문에 read-after-write를 깨뜨릴 수 있습니다`, `backup은 실수 삭제와 과거 복구를 위한 것이며 restore drill로만 검증됩니다`, `failover는 새 primary 승격뿐 아니라 old primary fencing과 client routing까지 포함합니다`처럼 말합니다.

또 약한 답은 정상 시나리오만 말합니다. 좋은 답은 실패한 뒤 무엇을 확인할지 말합니다. Replica lag가 threshold를 넘으면 어떤 요청을 primary로 보낼지, backup log chain이 끊기면 어느 시점까지만 복구 가능한지, failover 중 old primary가 살아 있으면 누가 쓰기 권한을 막는지, 복구 뒤 정상 transaction과 잘못된 transaction을 어떻게 구분할지까지 이어집니다. 이 지점이 실제 운영 경험과 단순 개념 암기를 가르는 부분입니다.

### 데이터 손실 범위를 설명하는 연습

복제 장애 답변에서 가장 어려운 부분은 데이터 손실을 추상적으로 말하지 않는 것입니다. `조금 잃을 수 있습니다`는 설명이 아닙니다. 어떤 commit이 client에게 성공으로 응답됐고, 그 commit의 log가 어느 replica까지 도착했으며, failover 후보가 그 log를 적용했는지를 말해야 합니다. PostgreSQL이라면 LSN과 replay 위치, MySQL이라면 GTID나 binary log position이 이런 판단의 언어가 됩니다. 제품 이름이 달라도 질문은 같습니다. 새 primary가 되려는 노드는 사용자가 성공으로 믿고 있는 변경을 어디까지 알고 있는가?

```text
client received commit success -> primary log position advanced -> replica received/applied position -> failover candidate chosen -> possible lost commits identified
```

이 trace를 그리면 비동기 복제의 RPO가 감각적으로 보입니다. Commit 성공 응답이 client에게 돌아갔지만 replica가 아직 그 log를 받지 못했다면, primary가 완전히 사라진 뒤 그 commit은 새 history에 없을 수 있습니다. 반대로 동기 복제는 commit 응답 전에 다른 노드의 확인을 기다려 이 틈을 줄입니다. 하지만 네트워크가 느리거나 replica가 멈추면 primary write latency가 증가하거나 쓰기를 멈출 수 있습니다. 그러므로 `동기 복제면 안전합니다`도 절반만 맞습니다. 어떤 장애에서 availability를 포기할지, 어떤 장애에서 데이터 손실을 감수할지 선택한 것입니다.

복구 뒤에는 손실 가능성이 있는 구간을 사용자 언어로 바꿔야 합니다. `LSN 0/16B6C50 이후 commit이 없습니다`는 운영자에게 필요한 정보지만, 서비스 의사결정에는 `13:01:12부터 13:01:18 사이 성공 응답을 받은 주문 7건이 새 primary에 없을 수 있습니다`처럼 번역되어야 합니다. 이 번역이 있어야 재처리, 고객 안내, 정산 보정, audit 기록이 가능합니다. 면접에서 이 수준까지 말하면 단순 인프라 지식이 아니라 데이터 책임을 이해하고 있다는 신호가 됩니다.

## 더 깊게 볼 자료

공식 자료:

- [PostgreSQL current: High Availability, Load Balancing, and Replication](https://www.postgresql.org/docs/current/high-availability.html)
- [PostgreSQL current: Log-Shipping Standby Servers](https://www.postgresql.org/docs/current/warm-standby.html)
- [PostgreSQL current: Continuous Archiving and Point-in-Time Recovery](https://www.postgresql.org/docs/current/continuous-archiving.html)
- [PostgreSQL current: WAL Configuration](https://www.postgresql.org/docs/current/wal-configuration.html)
- [MySQL 8.4 Reference Manual: Replication](https://dev.mysql.com/doc/refman/8.4/en/replication.html)
- [MySQL 8.4 Reference Manual: Backup and Recovery](https://dev.mysql.com/doc/refman/8.4/en/backup-and-recovery.html)
- [Linux kernel docs: Memory Management Concepts](https://docs.kernel.org/admin-guide/mm/concepts.html)
- [Linux kernel docs: Multi-Queue Block IO Queueing Mechanism](https://www.kernel.org/doc/html/latest/block/blk-mq.html)
- [Linux man-pages: fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)

저장소 안에서 이어 볼 자료:

- [database/replication.md](../../database/replication.md)
- [database/replication_lag.md](../../database/replication_lag.md)
- [database/deep-dive/reliability-distribution/15-replication-backup-recovery.md](../../database/deep-dive/reliability-distribution/15-replication-backup-recovery.md)
- [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)
- [distributed-systems-architecture.md](../distributed-systems-architecture.md)

WAL record 구조와 crash recovery 세부는 WAL 문서에서 깊게 보고, 이 문서에서는 WAL이나 binlog가 어떤 소비 경로에서 replication, PITR, failover 판단으로 이어지는지에 집중합니다. 같은 log라도 crash recovery, standby replay, PITR의 목적과 실패 경계는 다릅니다.
