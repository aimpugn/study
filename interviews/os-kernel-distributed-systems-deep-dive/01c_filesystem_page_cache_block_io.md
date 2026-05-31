# 01c. Filesystem, Page Cache, Block I/O

Kafka가 log segment에 append하고, Cassandra가 commit log와 SSTable을 쓰고, Spark가 shuffle spill file을 만드는 순간 세 시스템은 모두 파일 시스템과 block I/O 위에 올라갑니다. 파일을 쓴다는 말은 단순히 "디스크에 byte를 보낸다"가 아닙니다. file descriptor table, open file description, inode, dentry, VFS, page cache, dirty page, writeback, journaling, block layer, device queue가 순서대로 끼어듭니다.

이 문서의 목표는 `write()` 성공, page cache 반영, `fsync()` 완료, replication ack, storage device flush를 구분해 말하는 것입니다. 이 구분이 없으면 "Kafka는 ack했는데 왜 데이터가 사라질 수 있나", "Cassandra commit log sync가 왜 latency를 키우나", "Spark spill은 왜 memory 문제가 아니라 disk 문제로도 보이나" 같은 질문에서 흔들립니다.

## path, file descriptor, inode는 서로 다른 이름표다

파일을 열기 전에는 path가 중요합니다. `/var/lib/kafka/topic-0/000000000000.log` 같은 문자열은 directory entry를 따라 어떤 inode를 찾을지 알려 줍니다. 파일을 연 뒤 application이 주로 쓰는 것은 path가 아니라 file descriptor입니다. file descriptor는 process별 table의 작은 정수이고, 그 정수는 kernel의 open file description을 가리킵니다. open file description은 file offset과 status flag를 포함합니다.

```
path string
  -> directory lookup
  -> dentry cache
  -> inode
  -> open file description
  -> process fd table entry
  -> fd number returned to user process
```

inode는 파일의 metadata와 data block 위치를 나타내는 filesystem object입니다. dentry는 directory name과 inode 연결을 cache합니다. VFS는 ext4, XFS, tmpfs, NFS 같은 filesystem이 공통 file API로 보이게 하는 커널 계층입니다. 이 구조를 모르면 `rename()`이 왜 atomic하게 보일 수 있는지, 새 파일 생성 후 directory `fsync()`를 왜 따로 이야기하는지, fd를 이미 얻은 뒤 path를 바꿔도 fd가 계속 파일을 가리킬 수 있는 이유를 설명하기 어렵습니다.

## `write()`에서 block device까지 내려가는 길

일반적인 buffered regular-file write를 따라가면 다음과 같습니다. 예외는 있습니다. `O_DIRECT`, device file, network filesystem, tmpfs, pipe/socket은 경로가 다릅니다. 하지만 Kafka/Cassandra/Spark의 일반 disk-backed file을 이해하는 기본 모델로는 이 trace가 중요합니다.

```
user write(fd, buf, len)
  -> syscall entry
  -> fd table lookup
  -> open file description and file offset
  -> VFS dispatch
  -> filesystem write operation
  -> copy bytes from user buffer
  -> page cache page updated
  -> page marked dirty
  -> write() returns accepted byte count

later
  -> background writeback or fsync request
  -> filesystem maps file offset to blocks/extents
  -> block layer queues bios
  -> I/O scheduler / device queue
  -> driver submits to SSD/HDD/NVMe
  -> completion interrupt or polling path
```

이 trace에서 `write()` 반환은 "kernel/filesystem write path가 byte를 받아들였다"에 가깝습니다. dirty page가 stable storage에 내려갔는지는 별도 문제입니다. background writeback이 나중에 처리할 수도 있고, application이 `fsync()`로 더 강하게 밀어붙일 수도 있습니다.

## page cache는 성능과 착각을 동시에 만든다

page cache는 파일 내용을 page 단위로 들고 있는 kernel memory입니다. read path에서는 disk에서 읽은 file page를 재사용하게 해 줍니다. write path에서는 application이 준 byte를 먼저 memory의 file page에 반영하고 dirty로 표시하게 해 줍니다.

```
first read:
  disk -> page cache -> user buffer

second read:
  page cache -> user buffer

buffered write:
  user buffer -> page cache dirty page
  later writeback -> disk
```

Kafka가 최근 record를 빠르게 consumer에게 줄 수 있는 이유 중 하나는 log segment byte가 page cache에 남아 있을 수 있기 때문입니다. Cassandra SSTable read도 OS page cache의 도움을 받을 수 있습니다. Spark shuffle spill file도 local disk에 내려간 뒤 다시 읽힐 때 page cache 영향을 받을 수 있습니다.

그러나 page cache 때문에 "빠르게 반환됐다"와 "저장 장치에 안전하게 기록됐다"가 분리됩니다. page cache는 성능을 높이지만 crash consistency 질문을 남깁니다.

## `fsync()`는 만능 주문이 아니라 더 강한 동기화 요청이다

`fsync(fd)`는 file의 in-core data와 필요한 metadata를 storage 쪽으로 동기화하려는 system call입니다. `fdatasync()`는 보통 data와 data를 읽는 데 필요한 metadata에 더 초점을 둡니다. 하지만 directory entry와 rename pattern, filesystem journal, drive write cache, network filesystem은 추가 caveat를 만듭니다.

예를 들어 새 파일을 쓰고 atomic하게 교체하려는 흔한 패턴은 단순히 파일 `fsync()` 하나로 끝나지 않습니다.

```
write temp file
  -> fsync(temp fd)
  -> rename(temp, final)
  -> fsync(parent directory)
```

왜 directory를 sync하느냐가 중요합니다. 파일 내용은 sync되었어도 "이 이름이 이 inode를 가리킨다"는 directory entry update가 crash 후 보존되는지는 별도 metadata 문제일 수 있기 때문입니다. filesystem마다 세부 보장은 다르므로, 운영 설계에서는 target filesystem 문서를 확인해야 합니다.

Cassandra commit log sync 정책은 이 tradeoff를 그대로 노출합니다. 더 자주 sync하면 crash 후 손실 창은 줄지만 write latency가 올라갑니다. Kafka의 `acks=all`은 replica protocol의 성공 조건이지, 각 broker의 local storage flush와 같은 말이 아닙니다. 이 차이를 섞으면 내구성 설명이 틀어집니다.

## filesystem journal과 database log는 서로 다른 계층의 log다

filesystem journal은 filesystem metadata나 data consistency를 지키기 위한 filesystem 내부 장치입니다. database나 Kafka/Cassandra의 commit log는 application-level recovery를 위한 장치입니다. 이름이 모두 log라고 해서 같은 약속을 하지 않습니다.

```
filesystem journal
  -> filesystem structure consistency
  -> directory/inode/block allocation metadata recovery

Kafka / Cassandra log
  -> application record or mutation recovery
  -> offset, ack, replay, replica protocol
```

파일시스템이 journal을 쓴다고 Cassandra commit log가 필요 없어지지 않습니다. filesystem은 "파일시스템 구조가 깨지지 않도록" 도와주지만, "어떤 mutation을 ack했고 memtable에 반영했는가"는 Cassandra가 알아야 합니다. Kafka도 OS filesystem에 파일을 쓰지만, record offset, leader epoch, replica high watermark 같은 application-level 의미는 Kafka protocol이 관리합니다.

## block layer는 I/O를 device queue로 바꾸는 병목 지점이다

filesystem은 file offset을 block 또는 extent로 바꿉니다. block layer는 bio/request를 만들고 device queue로 보냅니다. SSD/NVMe는 parallel queue와 낮은 seek cost를 갖지만, 무한히 빠르지 않습니다. HDD는 seek와 rotational latency가 크므로 sequential I/O와 random I/O 차이가 더 큽니다. SSD도 write amplification, garbage collection, queue depth, fsync latency 영향을 받습니다.

```
dirty pages selected
  -> filesystem maps offsets
  -> bios created
  -> block queue
  -> device driver
  -> device internal cache/controller
  -> media
  -> completion
```

Kafka는 append-only segment와 sequential read/write로 이 경로에 유리한 모양을 만들려 합니다. Cassandra는 write path를 append와 memtable로 빠르게 만들지만, compaction은 많은 SSTable을 읽고 새 SSTable을 쓰기 때문에 background I/O가 커집니다. Spark shuffle은 많은 작은 block과 spill file을 만들 수 있어 local disk와 metadata overhead를 압박합니다.

## 현실 시나리오: `write()`는 성공했는데 장애 후 일부 record가 없다

이 상황에서 먼저 나눠야 할 성공은 네 가지입니다.

| 성공처럼 보이는 사건 | 실제 뜻 | 아직 남는 질문 |
|---|---|---|
| `write()` returned N | kernel/file path가 N byte를 받아들임 | dirty page가 storage에 갔는가? |
| `fsync()` returned 0 | file sync 요청이 성공함 | directory entry, device cache, filesystem caveat는? |
| Kafka `acks=all` | Kafka ISR 조건을 만족함 | local flush 정책, unclean election, replica state는? |
| Cassandra CL success | 필요한 replica 응답 수를 받음 | commit log sync mode, failed replica, repair state는? |

장애 후 복구를 설명하려면 이 중 어느 지점까지 증거가 남았는지 찾아야 합니다. "파일을 썼다"라는 말 하나로는 부족합니다.

## 문서를 덮고 확인할 것

- path, fd, open file description, inode를 한 문장씩 구분해 보세요.
- buffered `write()`가 page cache dirty page를 만들고, `fsync()`가 더 강한 storage 동기화 요청이라는 흐름을 그려 보세요.
- filesystem journal과 Kafka/Cassandra log가 왜 다른 계층의 log인지 설명해 보세요.
- Kafka segment append, Cassandra compaction, Spark spill이 모두 disk I/O지만 서로 다른 access pattern을 만든다는 점을 비교해 보세요.
