# 01c. Filesystem, Page Cache, Block I/O

## 목차

- [path, file descriptor, inode는 서로 다른 이름표다](#path-file-descriptor-inode는-서로-다른-이름표다)
- [`write()`에서 block device까지 내려가는 길](#write에서-block-device까지-내려가는-길)
- [page cache는 성능과 착각을 동시에 만든다](#page-cache는-성능과-착각을-동시에-만든다)
- [`fsync()`는 만능 주문이 아니라 더 강한 동기화 요청이다](#fsync는-만능-주문이-아니라-더-강한-동기화-요청이다)
- [filesystem journal과 database log는 서로 다른 계층의 log다](#filesystem-journal과-database-log는-서로-다른-계층의-log다)
- [block layer는 I/O를 device queue로 바꾸는 병목 지점이다](#block-layer는-io를-device-queue로-바꾸는-병목-지점이다)
- [현실 시나리오: `write()`는 성공했는데 장애 후 일부 record가 없다](#현실-시나리오-write는-성공했는데-장애-후-일부-record가-없다)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [VFS는 다양한 파일시스템을 하나의 파일처럼 보이게 한다](#vfs는-다양한-파일시스템을-하나의-파일처럼-보이게-한다)
- [Page cache writeback은 application 밖에서 계속 움직인다](#page-cache-writeback은-application-밖에서-계속-움직인다)
- [`fsync()`와 directory fsync를 분리해야 한다](#fsync와-directory-fsync를-분리해야-한다)
- [Crash consistency는 순서를 잃었을 때도 구조를 복구하는 문제다](#crash-consistency는-순서를-잃었을-때도-구조를-복구하는-문제다)
- [Block layer와 장치 queue는 병렬성을 주지만 큐도 만든다](#block-layer와-장치-queue는-병렬성을-주지만-큐도-만든다)
- [Device driver, DMA, interrupt completion은 storage에도 있다](#device-driver-dma-interrupt-completion은-storage에도-있다)
- [Remote filesystem과 distributed storage는 local file API의 경계를 흐린다](#remote-filesystem과-distributed-storage는-local-file-api의-경계를-흐린다)
- [Interview replay: 파일 쓰기에서 제품 내구성까지 한 번에 말하기](#interview-replay-파일-쓰기에서-제품-내구성까지-한-번에-말하기)
- [File descriptor limit과 inode/dentry cache도 운영 지표다](#file-descriptor-limit과-inodedentry-cache도-운영-지표다)
- [Permission, quota, mount option은 같은 `EACCES` 뒤에 숨어 있다](#permission-quota-mount-option은-같은-eacces-뒤에-숨어-있다)
- [작은 실험으로 page cache와 sync 감각 만들기](#작은-실험으로-page-cache와-sync-감각-만들기)

Kafka가 log segment에 append하고, Cassandra가 commit log와 SSTable을 쓰고, Spark가 shuffle spill file을 만드는 순간 세 시스템은 모두 파일 시스템과 block I/O 위에 올라갑니다. 파일을 쓴다는 말은 단순히 "디스크에 byte를 보낸다"가 아닙니다. file descriptor table, open file description, inode, dentry, VFS, page cache, dirty page, writeback, journaling, block layer, device queue가 순서대로 끼어듭니다.

이 문서의 목표는 `write()` 성공, page cache 반영, `fsync()` 완료, replication ack, storage device flush를 구분해 말하는 것입니다. 이 구분이 없으면 "Kafka는 ack했는데 왜 데이터가 사라질 수 있나", "Cassandra commit log sync가 왜 latency를 키우나", "Spark spill은 왜 memory 문제가 아니라 disk 문제로도 보이나" 같은 질문에서 흔들립니다.

처음에는 파일 쓰기를 네 이름으로 나누면 됩니다.

- **path는 파일을 찾기 위한 이름**입니다. 디렉터리 탐색을 통해 inode에 도달합니다.
- **file descriptor는 process 안에서 열린 파일 상태를 가리키는 작은 정수 handle**입니다. 파일 자체가 아니라 커널 테이블의 index입니다.
- **page cache는 파일 내용을 page 단위로 들고 있는 커널 메모리 캐시**입니다. 빠른 read/write를 돕지만 storage 내구성과는 다른 시점을 만듭니다.
- **block device queue는 filesystem이 만든 I/O 요청이 실제 저장 장치 앞에서 기다리는 큐(queue)**입니다.

```text
path lookup
  -> inode
  -> open file state
  -> fd number
  -> write(fd, bytes)
  -> page cache dirty page
  -> later writeback / fsync
  -> block device queue
```

이 trace를 잡으면 "파일에 썼다"라는 말을 `fd` lookup, page cache 변경, dirty page writeback, device completion으로 나누어 설명할 수 있습니다.

## path, file descriptor, inode는 서로 다른 이름표다

파일을 열기 전에는 path가 중요합니다. `/var/lib/kafka/topic-0/000000000000.log` 같은 문자열은 directory entry를 따라 어떤 inode를 찾을지 알려 줍니다. 파일을 연 뒤 application이 주로 쓰는 것은 path가 아니라 file descriptor입니다. file descriptor는 process별 table의 작은 정수이고, 그 정수는 kernel의 open file description을 가리킵니다. open file description은 file offset과 status flag를 포함합니다.

```text
path string
  -> directory lookup
  -> dentry cache
  -> inode
  -> open file description
  -> process fd table entry
  -> fd number returned to user process
```

inode는 파일의 metadata와 data block 위치를 나타내는 filesystem object입니다. dentry는 directory name과 inode 연결을 cache합니다. VFS는 ext4, XFS, tmpfs, NFS 같은 filesystem이 공통 file API로 보이게 하는 커널 계층입니다. 이 구조를 모르면 `rename()`이 왜 atomic하게 보일 수 있는지, 새 파일 생성 후 directory `fsync()`를 왜 따로 이야기하는지, fd를 이미 얻은 뒤 path를 바꿔도 fd가 계속 파일을 가리킬 수 있는 이유를 설명하기 어렵습니다.

초보자는 각 이름표의 소유자와 수명을 같이 보면 덜 흔들립니다.

| 이름 | 주로 누가 소유하는가 | 언제 바뀌는가 | 헷갈리면 생기는 오해 |
| --- | --- | --- | --- |
| path | directory tree와 namespace | rename, mount, directory update | path가 바뀌면 열린 fd도 사라진다고 착각한다 |
| dentry | kernel cache | directory lookup, cache eviction | 이름과 inode 연결을 매번 disk에서만 찾는다고 생각한다 |
| inode | filesystem | file metadata update, unlink/link | 서로 다른 path가 같은 파일 identity를 가질 수 있음을 놓친다 |
| open file description | kernel의 열린 파일 상태 | `open()`, `dup()`, `fork()` 공유 | 두 fd가 file offset을 공유할 수 있음을 놓친다 |
| fd number | process fd table | `open()`, `close()`, `dup2()` | fd 숫자를 파일 자체로 착각한다 |

## `write()`에서 block device까지 내려가는 길

일반적인 buffered regular-file write를 따라가면 다음과 같습니다. 예외는 있습니다. `O_DIRECT`, device file, network filesystem, tmpfs, pipe/socket은 경로가 다릅니다. 하지만 Kafka/Cassandra/Spark의 일반 disk-backed file을 이해하는 기본 모델로는 이 trace가 중요합니다.

```text
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

이 trace에서 `write()` 반환은 "kernel/filesystem write path가 byte를 받아들였다"에 가깝습니다. dirty page가 stable storage에 내려갔는지는 별도 문제입니다. background writeback이 뒤이어 처리할 수도 있고, application이 `fsync()`로 더 강하게 밀어붙일 수도 있습니다.

## page cache는 성능과 착각을 동시에 만든다

page cache는 파일 내용을 page 단위로 들고 있는 kernel memory입니다. read path에서는 disk에서 읽은 file page를 재사용하게 해 줍니다. write path에서는 application이 준 byte를 먼저 memory의 file page에 반영하고 dirty로 표시하게 해 줍니다.

```text
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

`fsync(fd)`는 file의 in-core data와 필요한 metadata를 storage 쪽으로 동기화하려는 system call입니다. `fdatasync()`는 보통 data와 data를 읽는 데 필요한 metadata에 더 초점을 둡니다. 하지만 directory entry와 rename pattern, filesystem journal, drive write cache, network filesystem은 추가 주의점을 만듭니다.

예를 들어 새 파일을 쓰고 atomic하게 교체하려는 흔한 패턴은 단순히 파일 `fsync()` 하나로 끝나지 않습니다.

```text
write temp file
  -> fsync(temp fd)
  -> rename(temp, final)
  -> fsync(parent directory)
```

왜 directory를 sync하느냐가 중요합니다. 파일 내용은 sync되었어도 "이 이름이 이 inode를 가리킨다"는 directory entry update가 crash 후 보존되는지는 별도 metadata 문제일 수 있기 때문입니다. filesystem마다 세부 보장은 다르므로, 운영 설계에서는 target filesystem 문서를 확인해야 합니다.

Cassandra commit log sync 정책은 이 tradeoff를 그대로 노출합니다. 더 자주 sync하면 crash 후 손실 창은 줄지만 write latency가 올라갑니다. Kafka의 `acks=all`은 replica protocol의 성공 조건이지, 각 broker의 local storage flush와 같은 말이 아닙니다. 이 차이를 섞으면 내구성 설명이 틀어집니다.

## filesystem journal과 database log는 서로 다른 계층의 log다

filesystem journal은 filesystem metadata나 data consistency를 지키기 위한 filesystem 내부 장치입니다. database나 Kafka/Cassandra의 commit log는 application-level recovery를 위한 장치입니다. 이름이 모두 log라고 해서 같은 약속을 하지 않습니다.

```text
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

```text
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

| 성공처럼 보이는 이벤트 | 실제 뜻 | 아직 남는 질문 |
| --- | --- | --- |
| `write()` returned N | kernel/file path가 N byte를 받아들임 | dirty page가 storage에 갔는가? |
| `fsync()` returned 0 | file sync 요청이 성공함 | directory entry, device cache, filesystem 주의점은? |
| Kafka `acks=all` | leader가 in-sync replica 조건을 기준으로 protocol ack를 반환함 | `min.insync.replicas`, 현재 ISR size, high watermark visibility, unclean leader election, local flush 정책은? |
| Cassandra CL success | coordinator가 consistency level에 필요한 replica 응답 수를 받음 | 각 replica의 commitlog/memtable ack, `commitlog_sync`, read repair, regular repair, LWT/Accord 경로는? |

장애 후 복구를 설명하려면 이 중 어느 지점까지 증거가 남았는지 찾아야 합니다. "파일을 썼다"라는 말 하나로는 부족합니다.

Kafka와 Cassandra의 성공 조건은 OS storage 성공과 같은 층이 아닙니다.

- Kafka `acks=all`은 replica protocol의 조건입니다. `min.insync.replicas`보다 적은 in-sync replica만 남으면 성공할 수 없고, high watermark가 어디까지 올라갔는지에 따라 consumer에게 보이는 범위가 달라집니다. 하지만 이 말만으로 각 broker가 local disk flush까지 끝냈다고 말할 수는 없습니다.
- Cassandra consistency level은 coordinator가 몇 replica 응답을 성공으로 볼지 정하는 read/write 계약입니다. Replica 안에서는 commitlog append와 memtable 반영, commitlog sync mode가 따로 있고, read repair와 regular repair는 replica 간 차이를 나중에 줄이는 경로입니다. LWT나 Accord 계열의 합의 경로는 일반 write CL 답변과 분리해 말해야 합니다.

## 문서를 덮고 확인할 것

이 절까지의 기억 지도는 짧습니다. path는 이름이고, inode는 filesystem의 파일 identity에 가깝고, fd는 process 안에서 열린 파일 상태를 가리키는 handle입니다. `write()`는 page cache의 dirty page를 만들 수 있고, block layer와 device queue는 그 dirty page가 storage로 내려갈 때의 대기 지점입니다.

- path, fd, open file description, inode를 한 문장씩 구분해 보세요.
- buffered `write()`가 page cache dirty page를 만들고, `fsync()`가 더 강한 storage 동기화 요청이라는 흐름을 그려 보세요.
- filesystem journal과 Kafka/Cassandra log가 왜 다른 계층의 log인지 설명해 보세요.
- Kafka segment append, Cassandra compaction, Spark spill이 모두 disk I/O지만 서로 다른 access pattern을 만든다는 점을 비교해 보세요.

## VFS는 다양한 파일시스템을 하나의 파일처럼 보이게 한다

사용자 프로그램은 `open`, `read`, `write`, `fsync`, `close` 같은 system call을 사용합니다. 이 API만 보면 ext4, XFS, tmpfs, NFS, procfs, device file이 모두 비슷한 파일처럼 보입니다. 이 착시를 만드는 공통 계층이 VFS입니다. VFS(virtual filesystem switch)는 커널 안에서 pathname, file descriptor, inode, dentry, mount, filesystem operation을 연결해 주는 인터페이스입니다. 덕분에 application은 같은 system call을 호출하지만, 뒤에서는 각 filesystem이 자기 방식으로 block mapping, metadata update, permission, cache, journaling을 처리합니다.

```text
user process
  -> open("/var/lib/kafka/log/topic-0/000.log")
  -> VFS path lookup
  -> dentry cache and mount namespace
  -> inode representing the file
  -> filesystem-specific operations
  -> page cache and block mapping
```

Pathname은 문자열입니다. Dentry는 directory entry lookup 결과를 cache하는 커널 객체입니다. Inode는 파일의 metadata와 실제 파일 identity에 가까운 객체입니다. File descriptor는 process-local integer이고, open file description은 file offset과 status flag를 담는 kernel-side 열린 파일 상태입니다. 이 네 개를 섞으면 눈에 잘 안 띄는 버그를 설명하지 못합니다. 두 fd가 같은 open file description을 공유하면 file offset도 공유할 수 있습니다. Hard link는 서로 다른 pathname이 같은 inode를 가리키는 구조입니다. Rename은 pathname과 directory entry를 바꾸지만 이미 열린 fd가 가리키는 file object를 곧바로 무효화하지 않습니다.

권한도 VFS 경로에서 해석됩니다. 프로세스 권한 정보(process credential), 파일 시스템 권한 비트, ACL, 마운트 옵션, capability, namespace가 모두 영향을 줄 수 있습니다. 컨테이너 안에서 `/data`가 보인다는 사실과 host에서 같은 path가 무엇을 가리키는지는 mount namespace 때문에 다를 수 있습니다. Kafka log directory permission 문제, Cassandra data directory ownership 문제, Spark local directory mount 문제는 모두 "파일을 열 수 없다"라는 단순 오류로 보이지만, 실제로는 path lookup, namespace, permission, quota, mount option 중 하나가 깨진 것입니다.

## Page cache writeback은 application 밖에서 계속 움직인다

Buffered write에서 application이 넘긴 bytes는 보통 page cache의 file page에 반영되고 dirty로 표시됩니다. `write()`가 반환된 뒤에도 dirty page는 메모리에 남아 있을 수 있습니다. 커널은 background writeback thread, memory pressure, dirty ratio, filesystem 정책, explicit `fsync()` 같은 조건에 따라 dirty page를 storage로 보냅니다. 그래서 "write가 빠르다"는 것이 "storage가 빠르다"와 같은 말이 아닙니다.

```text
application write()
  -> copy bytes into page cache page
  -> mark page dirty
  -> return to user process

later
  -> writeback selects dirty pages
  -> filesystem maps file offsets to blocks/extents
  -> block layer submits I/O
  -> completion clears dirty state
```

Dirty page가 너무 많아지면 application write가 갑자기 느려질 수 있습니다. 커널이 writeback을 따라가지 못하면 쓰는 thread가 balance dirty pages 경로에서 throttle될 수 있습니다. Kafka producer latency가 어느 순간 튀거나 Cassandra commitlog append가 평소보다 느려지는 경우, user-space lock이나 GC뿐 아니라 dirty writeback과 block device queue를 같이 봐야 합니다. Spark shuffle spill도 같은 local disk와 page cache를 사용하므로, 많은 executor task가 동시에 spill하면 dirty page와 writeback 압력이 커질 수 있습니다.

Page cache는 read에도 영향을 줍니다. 첫 read는 disk에서 가져오지만, 다음 read는 memory에서 빠르게 올 수 있습니다. Readahead는 순차 접근을 예상해 뒤쪽 page를 미리 가져옵니다. Kafka consumer가 sequential fetch를 하면 page cache와 readahead가 유리하게 작동할 수 있습니다. Cassandra는 random read가 많아 Bloom filter와 index로 SSTable 접근을 줄이려 하고, compaction은 큰 sequential I/O를 만들 수 있습니다. Spark shuffle read는 많은 작은 block을 읽으면서 metadata lookup과 random access 비용을 만들 수 있습니다.

## `fsync()`와 directory fsync를 분리해야 한다

`fsync(fd)`는 열린 파일에 대한 데이터와 필요한 metadata를 storage 쪽으로 동기화하려는 요청입니다. 하지만 "파일이 안전하다"라는 말은 상황에 따라 더 좁게 나눠야 합니다. 기존 파일의 data overwrite와 새 파일 생성, rename을 통한 atomic replace, directory entry update는 서로 다른 내구성 질문을 만듭니다. 새 파일을 쓰고 `rename()`으로 교체하는 패턴에서는 파일 fd에 대한 `fsync()`뿐 아니라 부모 directory의 entry가 durable한지도 고려해야 할 수 있습니다.

```text
write temp file
  -> fsync(temp fd)
  -> rename temp to final name
  -> fsync(parent directory) may be needed for directory entry durability
```

파일시스템마다 보장과 주의점이 다르고, storage device의 volatile cache, write barrier, mount option, network filesystem도 영향을 줍니다. 그래서 `fsync()`는 강한 요청이지만 만능 보증 문장은 아닙니다. 정확한 보장은 운영체제, filesystem, device, mount option, application pattern에 묶어 말해야 합니다.

Kafka의 durability 설정도 같은 방식으로 읽어야 합니다. Producer가 ack를 받는다는 것은 broker protocol 관점의 성공입니다. 그 record가 leader broker page cache에만 있는지, replica ISR에 복제되었는지, flush policy에 의해 storage sync까지 갔는지는 별도 축입니다. Cassandra write success도 commitlog sync mode, replica response, consistency level, hinted handoff, repair 상태와 함께 봐야 합니다. Spark checkpoint도 checkpoint file을 썼다는 사실과 외부 storage의 rename/fsync/visibility 보장은 구분해야 합니다. 특히 object storage나 Hadoop/Spark commit protocol을 쓰는 경우에는 POSIX local filesystem의 `rename()`/directory `fsync()` 감각이 그대로 옮겨가지 않을 수 있으므로 storage backend별 보장을 확인해야 합니다.

## Crash consistency는 순서를 잃었을 때도 구조를 복구하는 문제다

Crash consistency는 전원이 나가거나 kernel panic이 발생했을 때 filesystem 구조와 application data가 어떤 상태로 남는지를 다룹니다. Storage는 application이 생각한 순서대로 모든 write를 반영하지 않을 수 있고, 일부 metadata만 반영되거나 data만 반영될 수 있습니다. Filesystem journal은 metadata update를 log처럼 기록해 crash 뒤 filesystem 구조를 일관된 상태로 회복하려는 장치입니다. 하지만 application이 "record A 다음 record B"라는 의미를 갖는다면, 그 의미는 application log나 transaction protocol이 따로 관리해야 합니다.

```text
application-level intent
  append record R
  update index I
  acknowledge request

filesystem-level concern
  allocate blocks
  update inode size
  update directory metadata
  replay journal after crash

application recovery
  inspect log/checksum/offset
  discard partial record
  rebuild index or replay committed entries
```

이 구분 때문에 database와 message broker는 write-ahead log, checksum, length prefix, segment recovery, snapshot, manifest 같은 구조를 갖습니다. Kafka log segment는 record batch와 offset, checksum, index를 통해 장애 후 어디까지 유효한지 찾습니다. Cassandra는 commitlog를 replay해 memtable에 반영되지 못한 mutation을 복원하고, SSTable은 immutable file과 compaction 결과의 atomic visibility를 관리합니다. Spark streaming checkpoint는 driver 재시작 시 offset, state, progress를 복구하기 위한 application-level 기록입니다.

Log-structured thinking은 storage를 in-place update가 아니라 append와 merge로 다루는 사고입니다. 순차 append는 disk와 SSD에 유리하고 crash recovery가 쉬워질 수 있지만, 오래된 version과 tombstone, compaction, space amplification이 따라옵니다. Cassandra의 LSM storage는 이 tradeoff를 정면으로 선택합니다. Kafka의 partition log도 append와 segment retention을 중심으로 생각합니다. Spark shuffle file은 장기 저장 log는 아니지만, stage 사이의 materialized intermediate data라는 점에서 failure와 cleanup 질문을 만듭니다.

## Block layer와 장치 queue는 병렬성을 주지만 큐도 만든다

Block layer는 filesystem이 만든 I/O 요청을 device가 처리할 수 있는 request로 정리합니다. HDD 시절에는 seek를 줄이는 scheduling이 중요했고, SSD/NVMe에서는 여러 queue와 parallel command가 중요합니다. Queue depth는 동시에 outstanding 상태로 둘 수 있는 I/O 수입니다. 너무 낮으면 device parallelism을 쓰지 못하고, 너무 높으면 latency가 늘고 tail이 길어질 수 있습니다. I/O scheduler는 throughput과 latency 사이에서 요청을 정렬하거나 merge하거나 fairness를 조절합니다.

이 절에서 새로 나오는 이름은 아래처럼 이어집니다.

- bio: file offset에서 block 위치로 바뀐 뒤 만들어지는 낮은 수준의 block I/O 단위입니다.
- block request: device queue가 처리할 수 있도록 bio를 묶거나 정리한 요청입니다.
- queue depth: 장치 앞에 동시에 outstanding 상태로 둘 수 있는 요청 수입니다.
- completion: 장치가 I/O를 끝냈다고 kernel에 알려 waiting task나 writeback path가 다음으로 갈 수 있게 하는 이벤트입니다.

```text
many application writes
  -> dirty pages
  -> writeback submits bio objects
  -> block layer builds device requests
  -> device queue waits with finite queue depth
  -> driver submits commands
  -> completion event clears wait or dirty state
```

NVMe는 빠르지만 무한히 빠르지 않습니다. Random write, fsync-heavy workload, write amplification, device garbage collection, thermal throttling, filesystem metadata update가 latency를 만들 수 있습니다. Network-attached storage나 cloud block device는 더 많은 층이 있습니다. Application은 local file처럼 보지만 실제로는 network, remote storage controller, replication, throttling policy가 뒤에 있을 수 있습니다. 그래서 cloud 환경에서 `fsync()` latency가 튈 때는 local kernel trace와 provider metric을 함께 봐야 합니다.

Kafka와 Cassandra에서 disk는 단순 저장소가 아니라 protocol latency에 연결됩니다. Kafka leader가 record를 append하고 replica fetcher가 따라오며, consumer fetch가 page cache나 disk에서 읽습니다. Cassandra coordinator는 replica에 mutation을 보내고, replica는 commitlog와 memtable을 거쳐 ack합니다. Compaction은 foreground read/write와 같은 device bandwidth를 씁니다. Spark는 shuffle spill과 read가 stage boundary를 지배할 수 있습니다. 세 제품 모두 "디스크가 느리다"가 아니라 "어느 queue와 어느 write pattern이 어떤 protocol 대기를 만들었는가"를 봐야 합니다.

## Device driver, DMA, interrupt completion은 storage에도 있다

네트워크에서 DMA와 interrupt를 배우지만 storage도 같은 원리를 갖습니다. Driver는 block request를 device command로 바꾸고, device는 DMA를 통해 memory와 data를 주고받으며, completion interrupt나 polling으로 커널에 완료를 알립니다. Application thread는 synchronous read/write나 `fsync()`에서 직접 기다릴 수도 있고, background writeback이 기다릴 수도 있습니다. 완료가 오면 waiting task가 wakeup되고 scheduler가 다시 실행 기회를 줍니다.

```text
read miss in page cache
  -> filesystem maps file offset
  -> block request submitted
  -> driver programs device DMA
  -> task sleeps waiting for page
  -> device completes
  -> interrupt/polling completion
  -> page becomes uptodate
  -> task wakes and copies or maps data
```

이 trace를 알면 "read system call이 왜 오래 걸렸는가"를 단순히 disk latency로만 보지 않습니다. Page cache miss, filesystem metadata lookup, block queue congestion, device completion, scheduler wakeup이 모두 포함됩니다. `iostat`의 await와 utilization, `pidstat -d`, `blktrace`/eBPF block trace, filesystem-specific stats, application latency histogram을 함께 놓고 봐야 합니다.

## Remote filesystem과 distributed storage는 local file API의 경계를 흐린다

운영체제 교과서의 filesystem 장은 보통 local disk를 기준으로 시작하지만, 백엔드 시스템은 NFS, EBS 같은 cloud block storage, HDFS, object storage, distributed filesystem 위에서 자주 실행됩니다. Application은 path와 fd를 쓰지만 뒤쪽에는 network, remote server, replication, cache consistency, lease, rename semantics가 있을 수 있습니다. Local filesystem에서 익숙한 atomic rename이나 fsync 비용 감각이 그대로 적용되지 않을 수 있습니다.

Spark checkpoint를 object storage에 쓰는 경우와 local POSIX filesystem에 쓰는 경우는 rename, listing consistency, commit protocol이 달라질 수 있습니다. Kafka log directory는 일반적으로 local disk를 기대하는 설계가 강하고, remote/network filesystem은 latency와 semantics 문제가 큽니다. Cassandra data directory도 low-latency local disk와 predictable fsync를 전제로 tuning되는 경우가 많습니다. Storage가 멀어지면 파일 API는 같아 보여도 failure domain과 latency distribution이 달라집니다.

따라서 storage를 설명할 때 마지막 질문은 "이 fd 뒤에 실제로 무엇이 있는가"입니다. Local NVMe인지, cloud block volume인지, network filesystem인지, container overlay filesystem인지, tmpfs인지에 따라 page cache, writeback, fsync, crash recovery, throughput, tail latency가 달라집니다. 같은 `write()`라도 뒤쪽 장치와 filesystem이 다르면 시스템의 의미가 달라질 수 있습니다.

## Interview replay: 파일 쓰기에서 제품 내구성까지 한 번에 말하기

이 장의 답변은 이렇게 압축할 수 있습니다. "파일을 쓴다는 것은 path 문자열이 VFS lookup을 거쳐 inode와 open file description으로 연결되고, buffered write가 page cache page를 dirty로 만들며, writeback이나 fsync가 그 dirty page를 filesystem과 block layer, driver, device queue로 내려보내는 일입니다. `write()` 성공은 보통 kernel path가 byte를 받아들였다는 뜻이지 항상 durable storage를 뜻하지 않습니다."

꼬리 질문이 오면 네 층을 나눕니다. 첫째, file descriptor와 inode/open file description/pathname은 다릅니다. 둘째, page cache는 성능을 높이지만 dirty data와 crash timing을 만듭니다. 셋째, filesystem journal은 filesystem metadata consistency를 돕지만 application-level log semantics를 대신하지 않습니다. 넷째, block layer와 device queue, cloud storage, replication이 실제 latency와 durability를 바꿉니다.

Kafka로 연결하면 partition log append, page cache, flush, replica ack, high watermark를 나눕니다. Cassandra로 연결하면 commitlog append, memtable, SSTable flush, compaction, consistency level을 나눕니다. Spark로 연결하면 shuffle spill, checkpoint file, object storage commit semantics를 나눕니다. 이렇게 설명하면 파일 시스템 장이 단순한 OS 암기가 아니라 distributed product의 내구성과 성능을 읽는 기반이 됩니다.

## File descriptor limit과 inode/dentry cache도 운영 지표다

서버가 많은 connection과 file을 다루면 file descriptor limit이 병목이 됩니다. Socket도 fd이고, regular file도 fd이며, epoll instance도 fd입니다. Kafka broker는 많은 log segment와 client/replica socket을 열 수 있고, Cassandra는 SSTable file과 commitlog, streaming socket을 열며, Spark executor는 shuffle file과 network connection을 사용합니다. `Too many open files`는 단순 설정 오류처럼 보이지만, 실제로는 process-local fd table과 system-wide file object, inode/dentry cache가 함께 압박받는 신호입니다.

```text
many partitions or connections
  -> more sockets and files
  -> process fd table grows
  -> kernel file objects, dentries, inodes grow
  -> limit or memory pressure appears
```

열린 fd가 많다고 모두 문제는 아닙니다. 문제는 왜 많은지, lifetime이 의도된 것인지, close가 누락되었는지, segment/file rotation이 과한지, connection churn이 심한지입니다. `lsof`, `/proc/<pid>/fd`, `ss`, product metrics를 같이 보면 file leak, connection leak, 정상 cache, burst traffic을 구분할 수 있습니다. Dentry와 inode cache는 path lookup을 빠르게 하지만 memory를 쓰므로, memory pressure 아래에서는 reclaim 대상이 되기도 합니다.

## Permission, quota, mount option은 같은 `EACCES` 뒤에 숨어 있다

파일 I/O 오류는 error code 하나로 올라오지만 원인은 여러 층입니다. `EACCES`나 `EPERM`은 Unix permission bit, ACL, ownership, capability, readonly mount, SELinux/AppArmor 같은 LSM policy, container mount, seccomp와 연결될 수 있습니다. `ENOSPC`는 실제 disk block 부족일 수도 있고 inode 부족, project quota, filesystem reservation일 수도 있습니다. `EROFS`는 readonly filesystem이고, `EDQUOT`는 quota 초과입니다.

```text
application open/write fails
  -> check path namespace
  -> check permission/ownership/ACL
  -> check mount options
  -> check quota and free inodes
  -> check filesystem/device health
```

Kafka log directory가 readonly로 mount되었거나 disk full이 나면 broker는 partition log append를 계속할 수 없습니다. Cassandra data directory의 permission이 틀리면 node startup이나 flush가 실패합니다. Spark local directory가 작은 ephemeral storage 위에 있으면 shuffle spill이 `No space left on device`로 stage를 깨뜨릴 수 있습니다. 이 오류들은 application log에 한 줄로 나오지만, OS layer에서는 path, namespace, permission, quota, block availability를 차례로 봐야 합니다.

## 작은 실험으로 page cache와 sync 감각 만들기

Linux 환경이라면 같은 파일을 두 번 읽는 간단한 실험으로 page cache를 느낄 수 있습니다. 첫 read는 storage에서 page를 가져오고, 두 번째 read는 page cache에서 더 빠르게 올 가능성이 큽니다. 물론 실제 cache drop은 권한이 필요하고 다른 system activity가 영향을 주므로, 실험은 격리된 환경에서 해야 합니다. 목적은 숫자를 외우는 것이 아니라 "파일 read가 disk와 page cache 중 어디서 왔는가"를 묻는 습관입니다.

```text
first sequential read
  -> page cache miss
  -> block I/O
  -> pages become cached

second sequential read
  -> page cache hit
  -> much less device I/O
```

Sync 실험도 마찬가지입니다. 작은 write를 여러 번 하고 매번 `fsync()`하는 경우와, 여러 record를 모아 한 번에 sync하는 경우 latency와 throughput이 달라집니다. Kafka와 Cassandra가 batching과 flush/sync 정책을 제공하는 이유는 이 tradeoff 때문입니다. 단, 실험 결과를 제품에 그대로 대입하면 안 됩니다. 제품은 replication, checksum, compression, segment/index, group commit, background thread를 추가로 갖습니다. 실험은 lower-layer 감각을 만드는 출발점입니다.

마지막으로 파일 I/O를 설명할 때는 항상 두 질문을 같이 둡니다. 첫째, application이 어떤 성공을 받았는가. 둘째, 그 성공 뒤에 아직 어떤 queue와 cache와 recovery boundary가 남아 있는가. 이 두 질문을 분리하면 `write`, `fsync`, filesystem journal, Kafka ack, Cassandra consistency level, Spark checkpoint를 같은 단어로 뭉개지 않고 각 계층의 약속으로 읽을 수 있습니다.
