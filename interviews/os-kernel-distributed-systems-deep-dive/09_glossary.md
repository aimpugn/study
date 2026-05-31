# 09. Glossary

이 용어장은 영어 단어를 한국어로 치환하기 위한 표가 아닙니다. 목표는 처음 보는 용어를 쉬운 한국어로 붙잡고, 공식 문서를 다시 찾을 수 있게 원어를 남기며, 자주 헷갈리는 쌍을 바로 분리하는 것입니다.

| 용어 | 쉬운 한국어 의미 | 헷갈리는 개념 | 첫 등장 | 면접 한 문장 |
|---|---|---|---|---|
| 사용자 모드(user mode) | 애플리케이션 코드가 제한된 권한으로 실행되는 CPU 상태 | 커널 모드 | 01 | user mode 코드는 장치와 커널 자료구조를 직접 만지지 못하고 syscall로 요청합니다. |
| 커널 모드(kernel mode) | CPU가 권한 있는 명령과 커널 자료구조 접근을 허용하는 실행 상태 | root user | 01 | kernel mode는 사용자 계정 권한이 아니라 CPU와 OS가 정한 privileged execution state입니다. |
| 시스템 콜(system call) | 사용자 코드가 커널 기능을 요청하는 공식 진입 경로 | 일반 함수 호출 | 01 | `write()` wrapper는 함수처럼 보이지만 실제 파일 쓰기는 syscall entry로 커널에 요청합니다. |
| 파일 디스크립터(file descriptor) | process별 table에서 열린 파일 객체를 가리키는 작은 정수 | path, inode | 01 | `fd=3`은 파일 이름이 아니라 현재 process의 descriptor table index입니다. |
| 열린 파일 설명(open file description) | file offset과 status flag를 가진 system-wide 열린 파일 상태 | file descriptor | 01 | 여러 fd가 같은 open file description을 공유하면 offset도 공유될 수 있습니다. |
| VFS(virtual filesystem) | 여러 filesystem을 공통 file API로 보이게 하는 커널 계층 | 실제 filesystem | 01 | VFS는 요청을 dispatch하지, 그 자체가 데이터를 영구 저장하는 filesystem은 아닙니다. |
| page cache | 파일 내용을 memory page로 들고 있는 커널 cache | application cache | 01 | `write()` 성공은 page cache dirty 상태일 수 있고, stable storage 보장과 다릅니다. |
| dirty page | memory에는 최신 내용이 있지만 storage에 아직 반영되지 않았을 수 있는 file page | durable write | 01 | dirty page는 빠른 write를 가능하게 하지만 crash 시점의 내구성 질문을 남깁니다. |
| fsync | file의 in-core data/metadata를 storage 쪽으로 동기화하려는 system call | write, sync | 01 | `fsync()`는 `write()`보다 강한 요청이지만 directory entry와 storage stack caveat를 함께 봐야 합니다. |
| partial write | 요청한 byte 수보다 적게 받아들인 write 성공 | write failure | 01 | `write()` 반환값이 `len`보다 작으면 남은 byte를 다시 써야 합니다. |
| interrupt | 외부 장치가 CPU 흐름을 끊고 커널 처리를 요청하는 사건 | trap, syscall | 01 | network packet 도착은 interrupt로 시작할 수 있지만 user thread가 즉시 실행된다는 뜻은 아닙니다. |
| trap | 현재 instruction 실행 중 생기는 커널 진입 사건 | interrupt | 01 | page fault와 syscall은 원인이 다른 trap 계열 사건으로 설명할 수 있습니다. |
| context switch | CPU가 실행하던 thread 상태를 저장하고 다른 thread 상태로 바꾸는 일 | thread 생성 | 01 | context switch가 많으면 CPU가 실제 업무보다 실행 흐름 교체에 시간을 쓸 수 있습니다. |
| runnable queue | 실행 가능하지만 CPU를 아직 받지 못한 task들의 대기열 | request queue | 01 | CPU 사용률이 낮아도 runnable queue나 blocking wait 때문에 latency가 커질 수 있습니다. |
| virtual memory | process가 독립된 주소 공간을 가진 것처럼 보이게 하는 메모리 모델 | physical memory | 01 | virtual address는 page table을 거쳐 physical frame으로 번역됩니다. |
| page fault | 주소 번역이나 권한 확인이 실패해 커널이 개입하는 사건 | segmentation fault | 01 | page fault는 lazy allocation처럼 정상 흐름일 수도 있고 invalid access일 수도 있습니다. |
| socket buffer | kernel이 network send/receive byte를 임시로 담는 buffer | application queue | 01 | network backpressure는 application queue뿐 아니라 socket buffer에서도 나타납니다. |
| partial failure | 분산 시스템 일부만 실패하거나 응답이 불명확한 상태 | 전체 장애 | 02 | timeout은 상대가 죽었다는 증명이 아니라 적용 여부 불명확 상태일 수 있습니다. |
| happens-before | 어떤 사건이 다른 사건보다 먼저라고 말할 수 있는 논리적 관계 | wall-clock order | 02 | message send는 receive보다 먼저지만, 관계 없는 두 node의 timestamp만으로 전역 순서를 단정할 수 없습니다. |
| log | 상태 변화를 순서대로 남겨 replay나 복구를 가능하게 하는 기록 | 단순 로그 파일 | 02 | Kafka log, Cassandra commit log, Spark lineage는 모두 복구 기준점이지만 역할이 다릅니다. |
| partition | data나 일을 나누는 단위 | replica | 02 | partition은 성능 단위이면서 ordering, ownership, task scheduling의 경계가 됩니다. |
| replica | 같은 data나 log를 복사해 둔 사본 | partition | 02 | replica가 많아도 어떤 write가 몇 replica에 도달했는지 모르면 안전성을 말할 수 없습니다. |
| quorum | replica 중 몇 개의 응답을 성공 조건으로 볼지 정한 기준 | consensus | 02 | quorum은 응답 집합 조건이고, consensus는 값이나 순서에 대한 합의 protocol입니다. |
| consensus | 여러 node가 같은 값이나 log 순서에 안전하게 합의하는 protocol 문제 | quorum | 02 | Raft 같은 consensus는 leader, term, log index, majority commit rule로 안전성을 만듭니다. |
| linearizability | 실제 시간 순서와 맞는 하나의 원자적 순서처럼 보이는 성질 | eventual consistency | 02 | "consistent"라고 말하기 전에 어떤 read가 어떤 write를 반드시 봐야 하는지 물어야 합니다. |
| eventual consistency | update가 멈추고 통신이 회복되면 replica가 결국 수렴하는 성질 | stale read 없음 | 02 | eventual은 아무 값이나 읽어도 된다는 뜻이 아니라 수렴 경로와 repair가 필요하다는 뜻입니다. |
| backpressure | downstream이 느릴 때 upstream 속도를 줄이거나 실패를 드러내는 구조 | retry | 02 | backpressure 없이 retry하면 장애를 회복하지 않고 증폭할 수 있습니다. |
| checkpoint | 실패 후 다시 시작할 기준점을 stable storage에 남기는 것 | cache | 02 | checkpoint는 재계산 비용을 줄이지만 정상 실행 비용을 늘립니다. |
| Kafka offset | partition log 안에서 record 위치를 나타내는 번호 | timestamp | 03 | offset은 partition 안의 순서를 말하며 topic 전체 전역 순서는 아닙니다. |
| ISR(in-sync replicas) | leader log를 충분히 따라온 Kafka replica 집합 | 모든 replica | 03 | `acks=all`은 ISR 조건과 관련되지만 storage fsync와 같은 말은 아닙니다. |
| consumer group | 여러 consumer가 partition을 나눠 읽고 offset을 공유하는 그룹 | topic | 03 | 같은 group 안에서는 partition assignment가 병렬성의 핵심입니다. |
| log compaction | Kafka에서 key별 최신 record 중심으로 log를 정리하는 정책 | Cassandra compaction | 03 | Kafka compaction은 event history 전체 보존과 다른 목적입니다. |
| token ring | Cassandra에서 partition key hash token을 node range에 배치하는 모델 | Kafka partition | 04 | token ring은 data ownership과 replica placement를 정합니다. |
| replication factor(RF) | Cassandra에서 data를 몇 replica에 둘지 정한 값 | consistency level(CL) | 04 | RF는 사본 개수이고 CL은 요청 성공에 필요한 응답 수입니다. |
| memtable | Cassandra가 write를 memory에 정렬해 보관하는 구조 | page cache | 04 | memtable은 Cassandra 내부 write buffer이고, page cache는 OS의 file cache입니다. |
| SSTable | Cassandra가 memtable을 disk에 flush해 만든 immutable sorted file | commit log | 04 | SSTable은 read 대상 data file이고 commit log는 crash recovery용 append log입니다. |
| tombstone | Cassandra delete를 표시하는 marker | 즉시 물리 삭제 | 04 | tombstone은 compaction 전까지 read 비용과 resurrection 위험에 영향을 줄 수 있습니다. |
| compaction | 여러 SSTable을 읽어 merge하고 새 SSTable을 만드는 background 작업 | compression | 04 | compaction은 read amplification을 줄이지만 disk I/O와 write amplification을 만듭니다. |
| hinted handoff | 잠시 unavailable한 replica에 나중에 전달할 hint를 저장하는 장치 | repair 보장 | 04 | hints는 best-effort이며 anti-entropy repair를 대체하지 않습니다. |
| lineage | Spark partition을 다시 계산할 수 있게 transformation graph를 기억한 것 | checkpoint | 05 | lineage는 재계산 계획이고 checkpoint는 중간 결과를 stable storage에 남기는 기준점입니다. |
| driver | Spark application을 조율하고 job/stage/task를 scheduling하는 process | executor | 05 | driver는 단순 client가 아니라 scheduling과 metadata의 중심입니다. |
| executor | Spark task를 실행하고 data를 memory/disk에 보관하는 worker process | worker node | 05 | executor는 node 전체가 아니라 application별로 뜨는 process입니다. |
| stage | shuffle boundary를 기준으로 나뉜 task 묶음 | job | 05 | job은 action에서 생기고, stage는 dependency와 shuffle boundary로 나뉩니다. |
| shuffle | data를 key/partitioner 기준으로 다시 배치하는 작업 | network copy만 | 05 | shuffle은 network뿐 아니라 serialization, disk spill, memory, skew 문제입니다. |
| spill | memory에 다 담지 못한 중간 data를 disk로 밀어내는 일 | checkpoint | 05 | spill은 정상 처리 중 pressure 완화이고, checkpoint는 recovery 기준점입니다. |

## 헷갈리는 쌍 빠른 점검

- `fd`와 file path: path는 open 시점의 이름이고, fd는 열린 뒤 process가 사용하는 작은 handle입니다.
- `write()`와 `fsync()`: `write()`는 byte를 kernel write path에 넘기고, `fsync()`는 file state를 storage 쪽으로 동기화하려는 요청입니다.
- partition과 replica: partition은 나눈 조각이고, replica는 같은 조각의 사본입니다.
- quorum과 consensus: quorum은 응답 수 조건이고, consensus는 값/순서 합의 protocol입니다.
- Kafka compaction과 Cassandra compaction: Kafka는 key별 log 정리, Cassandra는 SSTable merge입니다.
- Spark cache와 checkpoint: cache는 성능 최적화, checkpoint는 lineage를 끊는 recovery 기준점입니다.

## 문서를 덮고 확인할 것

용어 하나를 고를 때마다 다음 문장으로 말해 봅니다.

```text
X는 Y가 아니라 Z다. 왜냐하면 실제 상태는 A에 있고, 실패하면 B를 확인해야 하기 때문이다.
```

예를 들어 "quorum은 consensus가 아니라 replica 응답 수 조건이다. 왜냐하면 실제 상태는 각 replica의 data/version에 있고, 실패하면 어떤 read/write 집합이 교차했는지 확인해야 하기 때문이다"처럼 말할 수 있으면 암기에서 한 걸음 벗어난 것입니다.
