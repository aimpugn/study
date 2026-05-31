# 08. Experiments and Observability

> 실험의 목적은 멋진 benchmark 숫자를 얻는 것이 아니라, 문서에서 설명한 상태 이동을 내 눈으로 확인하는 것입니다.
> Linux와 macOS는 관찰 도구와 권한 모델이 다르므로 같은 개념을 같은 명령으로 확인하려 하면 실패할 수 있습니다.
> 각 실험은 목적, 전제, 명령, 예상 관찰, PASS 신호, FAIL 신호를 분리해 둡니다.

## 1. 도구 지도

> 관찰 도구는 결론이 아니라 렌즈입니다.
> `strace`, `dtrace`, `vmstat`, `iostat`, `ss`, `lsof`, `jstack`, Spark UI, Kafka/Cassandra CLI는 서로 다른 계층을 봅니다.

| 목적 | Linux | macOS | 주의 |
|---|---|---|---|
| syscall 관찰 | `strace` | `dtruss`, `dtrace` | macOS SIP/권한 제한 가능 |
| CPU/memory/I/O 추세 | `vmstat` | `vm_stat`, `top` | 필드 의미가 다름 |
| block I/O | `iostat -xz` | `iostat -w` | 옵션과 출력 다름 |
| socket 상태 | `ss -tin`, `netstat` | `netstat` | `ss`는 Linux 중심 |
| 열린 파일 | `lsof` | `lsof` | 권한 필요할 수 있음 |
| JVM thread | `jstack`, `jcmd` | `jstack`, `jcmd` | JDK 설치 필요 |
| Kernel tracing | `perf`, `ftrace`, eBPF tools | `dtrace`, Instruments | 권한/보안 정책 주의 |
| Kafka | `kafka-topics.sh`, `kafka-consumer-groups.sh` | same if installed | local cluster 필요 |
| Cassandra | `cqlsh`, `nodetool` | same if installed | local cluster 필요 |
| Spark | Spark UI, `spark-submit` | same if installed | Java/Python version 주의 |

## 2. Experiment: syscall과 durable write 구분

> `write()`는 user buffer에서 kernel boundary를 넘는 사건이고, `fsync()`는 파일 data를 더 강하게 storage로 밀어 넣는 사건입니다.
> 이 둘을 구분하지 못하면 Kafka/Cassandra durability 질문에서 틀린 답을 하기 쉽습니다.

### 목적

파일 쓰기가 syscall로 내려가는 것을 보고, `write()`와 `fsync()`가 별도 syscall임을 확인합니다.

### 전제

- Linux: `strace` 설치
- macOS: `sudo dtruss` 가능해야 함

### 명령

Linux:

```bash
strace -e trace=write,fsync,fdatasync -o /tmp/write-trace.log python3 - <<'PY'
import os
fd = os.open("/tmp/durable-demo.txt", os.O_CREAT | os.O_WRONLY | os.O_TRUNC, 0o644)
os.write(fd, b"hello\n")
os.fsync(fd)
os.close(fd)
PY
sed -n '1,80p' /tmp/write-trace.log
```

macOS:

```bash
sudo dtruss -t write -t fsync -o /tmp/write-trace.log python3 - <<'PY'
import os
fd = os.open("/tmp/durable-demo.txt", os.O_CREAT | os.O_WRONLY | os.O_TRUNC, 0o644)
os.write(fd, b"hello\n")
os.fsync(fd)
os.close(fd)
PY
```

### 예상 관찰 결과

`write`와 `fsync`가 별도 syscall로 보입니다.

### PASS 신호

- `write(..., "hello\n", ...)` 또는 유사한 syscall이 보입니다.
- `fsync(fd)`가 별도로 보입니다.

### FAIL 신호

- tracing 권한 문제로 syscall이 보이지 않습니다.
- `write()`가 보였다는 이유만으로 device durability를 증명했다고 해석합니다.

## 3. Experiment: page cache 효과 관찰

> page cache는 disk I/O를 숨기기도 하고, memory pressure에서 다시 드러나기도 합니다.
> 같은 파일을 두 번 읽을 때 두 번째가 빠른 이유는 애플리케이션 code가 아니라 kernel cache일 수 있습니다.

### 목적

같은 파일을 반복 읽을 때 page cache 효과가 보일 수 있음을 확인합니다.

### 전제

- 큰 파일을 만들 disk 공간이 있어야 합니다.
- `/tmp`가 memory filesystem이면 결과가 왜곡될 수 있습니다.

### 명령

```bash
dd if=/dev/zero of=/tmp/page-cache-demo.bin bs=1m count=256
sync
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
```

Linux에서 cache drop은 root 권한과 system-wide 영향이 있어 이 문서에서는 기본 실험으로 권장하지 않습니다.

### 예상 관찰 결과

두 번째 read가 더 빠르거나 device read 증가가 작을 수 있습니다.

### PASS 신호

- 반복 read에서 시간 차이나 I/O 통계 차이를 관찰합니다.

### FAIL 신호

- 차이가 없다고 page cache가 없다고 결론 냅니다. SSD, memory filesystem, OS cache 상태, 파일 크기 때문에 차이가 작을 수 있습니다.

## 4. Experiment: runnable queue와 context switch

> CPU 병목은 CPU 사용률 하나로 끝나지 않습니다.
> runnable queue와 context switch가 늘면 thread가 CPU를 받기 위해 기다릴 수 있습니다.

### 목적

CPU-bound 작업을 여러 개 띄우고 `vmstat`에서 runnable queue와 context switch를 관찰합니다.

### 전제

- Linux `vmstat`
- macOS는 `top`, `powermetrics`, Activity Monitor로 대체 관찰

### 명령

Linux:

```bash
python3 - <<'PY' &
while True:
    pass
PY
python3 - <<'PY' &
while True:
    pass
PY
vmstat 1 5
pkill -f 'while True'
```

### 예상 관찰 결과

CPU core 수와 background loop 수에 따라 `r`과 context switch가 변합니다.

### PASS 신호

- runnable queue가 생기고 CPU user time이 증가합니다.

### FAIL 신호

- shared machine에서 `pkill`이 다른 작업을 종료할 수 있습니다. 개인 실험 환경에서만 사용하고, 더 안전하게는 PID를 기록해 종료합니다.

## 5. Experiment: socket state와 backpressure 감각

> network backpressure는 애플리케이션 queue뿐 아니라 kernel socket send/receive queue로도 나타납니다.
> receiver가 느리면 sender는 결국 buffer와 flow control에 막힙니다.

### 목적

간단한 TCP 연결을 만들고 socket state를 관찰합니다.

### 전제

- `nc` 또는 `python3`
- Linux는 `ss`, macOS는 `netstat`

### 명령

Terminal 1:

```bash
python3 - <<'PY'
import socket, time
s = socket.socket()
s.bind(("127.0.0.1", 18080))
s.listen(1)
conn, _ = s.accept()
time.sleep(30)
PY
```

Terminal 2:

```bash
python3 - <<'PY'
import socket
s = socket.create_connection(("127.0.0.1", 18080))
s.sendall(b"x" * 1024 * 1024)
input("press enter to close")
PY
```

Terminal 3:

Linux:

```bash
ss -tin sport = :18080 or dport = :18080
```

macOS:

```bash
netstat -an | grep 18080
```

### 예상 관찰 결과

연결 상태와 queue 관련 정보를 볼 수 있습니다. OS와 버전에 따라 세부 queue 숫자는 다릅니다.

### PASS 신호

- TCP connection이 ESTABLISHED 상태로 보입니다.
- receiver가 읽지 않으면 sender가 eventually block될 수 있음을 관찰합니다.

### FAIL 신호

- loopback 실험 결과를 실제 datacenter network latency와 동일하게 해석합니다.

## 6. Experiment: JVM thread dump로 lock/wait 보기

> JVM 기반 Kafka, Cassandra, Spark는 모두 OS process이면서 JVM thread 집합입니다.
> thread dump는 CPU가 낮은데 latency가 높은 상황에서 lock wait, blocking I/O, GC 영향을 추적하는 첫 도구가 될 수 있습니다.

### 목적

JVM process의 thread 상태를 확인합니다.

### 전제

- JDK 설치
- Java process 실행 중

### 명령

```bash
jps -l
jcmd <pid> Thread.print | sed -n '1,160p'
```

### 예상 관찰 결과

thread 이름, 상태(`RUNNABLE`, `WAITING`, `BLOCKED`), stack trace가 보입니다.

### PASS 신호

- 어떤 thread가 lock, sleep, I/O, executor work를 수행 중인지 구분합니다.

### FAIL 신호

- 한 번의 dump로 결론을 확정합니다. 여러 번 sampling해야 transient wait와 persistent blockage를 구분할 수 있습니다.

## 7. Kafka Local Experiment: offset과 lag

> Kafka consumer lag는 log end offset과 group committed offset 사이 거리입니다.
> 이 실험은 lag가 "broker에 쌓인 처리 대기 상태"라는 것을 확인합니다.

### 목적

topic에 record를 넣고 consumer group offset과 lag를 관찰합니다.

### 전제

- Local Kafka cluster 실행 중
- Kafka command-line tools 사용 가능

### 명령

```bash
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic lag-demo --partitions 3 --replication-factor 1
seq 1 1000 | kafka-console-producer.sh --bootstrap-server localhost:9092 --topic lag-demo
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic lag-demo --group lag-demo-group --max-messages 100
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group lag-demo-group
```

### 예상 관찰 결과

consumer가 100개만 읽고 멈추면 remaining offsets가 lag로 남습니다.

### PASS 신호

- partition별 current offset, log end offset, lag가 보입니다.
- lag는 group별 상태임을 확인합니다.

### FAIL 신호

- lag를 topic 자체의 미처리 record 수로만 해석합니다. group마다 lag가 다릅니다.

## 8. Kafka Local Experiment: partition ordering

> Kafka ordering은 partition 안에서 성립합니다.
> 같은 key가 같은 partition으로 가는지와 partition별 offset이 독립 증가하는지 확인합니다.

### 목적

keyed record의 partition assignment와 per-partition ordering을 관찰합니다.

### 전제

- Local Kafka cluster
- console producer가 key parsing을 지원

### 명령

```bash
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic key-demo --partitions 3 --replication-factor 1
printf 'u1:a\nu1:b\nu2:c\nu1:d\n' | kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic key-demo \
  --property parse.key=true \
  --property key.separator=:
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic key-demo --from-beginning \
  --property print.key=true --property print.partition=true --property print.offset=true --max-messages 4
```

### 예상 관찰 결과

같은 key `u1` record가 같은 partition에 배치되는 것을 기대할 수 있습니다. partitioner 설정에 따라 결과는 달라질 수 있습니다.

### PASS 신호

- offset은 partition별로 증가합니다.

### FAIL 신호

- 전체 출력 순서를 topic-wide total order로 해석합니다.

## 9. Cassandra Local Experiment: write/read trace

> Cassandra read/write는 coordinator와 replica 사이의 메시지, consistency level, storage path를 거칩니다.
> CQL tracing은 작은 query 하나가 내부에서 어떤 replica와 상호작용하는지 보여 줍니다.

### 목적

Consistency level과 tracing으로 coordinator path를 관찰합니다.

### 전제

- Local Cassandra cluster
- `cqlsh`

### 명령

```sql
CREATE KEYSPACE IF NOT EXISTS demo
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS demo.users (
  id text PRIMARY KEY,
  status text
);

CONSISTENCY ONE;
TRACING ON;
INSERT INTO demo.users (id, status) VALUES ('u1', 'ACTIVE');
SELECT * FROM demo.users WHERE id='u1';
```

### 예상 관찰 결과

trace에 parsing, replica contact, read/write activity가 나타납니다.

### PASS 신호

- consistency level과 replica 수가 query path에 영향을 준다는 감각을 얻습니다.

### FAIL 신호

- single-node RF=1 실험으로 quorum/failure behavior까지 검증했다고 해석합니다.

## 10. Cassandra Local Experiment: SSTable과 compaction 관찰

> Cassandra의 write path는 memtable flush로 SSTable을 만들고, compaction은 SSTable을 병합합니다.
> 이 실험은 LSM 구조의 흔적을 `nodetool`로 관찰하는 데 초점을 둡니다.

### 목적

flush와 compaction 관련 지표를 확인합니다.

### 전제

- Local Cassandra
- 운영 data가 아닌 실험 keyspace

### 명령

```bash
nodetool tablestats demo.users
nodetool flush demo users
nodetool tablestats demo.users
nodetool compact demo users
nodetool compactionstats
```

### 예상 관찰 결과

SSTable count, disk space, compaction 상태가 변할 수 있습니다.

### PASS 신호

- memtable flush와 SSTable/compaction 지표를 연결합니다.

### FAIL 신호

- production에서 무작정 `compact`를 실행합니다. 운영 compaction은 강한 I/O side effect가 있습니다.

## 11. Spark Local Experiment: DAG, stage, shuffle

> Spark action 하나는 DAG, stage, task로 쪼개집니다.
> wide dependency가 있는 연산은 shuffle stage를 만들고 Spark UI에서 확인할 수 있습니다.

### 목적

local mode에서 shuffle이 stage 경계를 만드는지 관찰합니다.

### 전제

- Spark 설치
- Java/Python version이 Spark 문서 요구와 맞음

### 명령

```bash
pyspark --master local[2]
```

PySpark shell:

```python
rdd = sc.parallelize([(i % 10, i) for i in range(100000)], 4)
out = rdd.groupByKey().mapValues(lambda xs: sum(xs))
out.collect()
```

브라우저:

```text
http://localhost:4040
```

### 예상 관찰 결과

Spark UI에서 job, stage, task, shuffle read/write가 보입니다.

### PASS 신호

- `groupByKey`가 shuffle을 만들고 stage가 나뉘는 것을 확인합니다.

### FAIL 신호

- local mode 결과를 cluster network shuffle 비용과 동일하게 해석합니다.

## 12. Spark Local Experiment: cache와 recomputation

> Spark cache는 성능 최적화이고, lineage는 recomputation 근거입니다.
> cache하지 않은 RDD를 여러 action에서 쓰면 lineage가 반복 실행될 수 있습니다.

### 목적

cache 전후 action 시간이 달라질 수 있음을 확인합니다.

### 전제

- Spark local mode

### 명령

```python
import time
rdd = sc.parallelize(range(20_000_000), 4).map(lambda x: x * x)
t = time.time(); rdd.count(); print("first", time.time() - t)
t = time.time(); rdd.count(); print("second without cache", time.time() - t)
cached = rdd.cache()
t = time.time(); cached.count(); print("cache materialize", time.time() - t)
t = time.time(); cached.count(); print("second with cache", time.time() - t)
```

### 예상 관찰 결과

cache materialization 이후 반복 action이 빨라질 수 있습니다.

### PASS 신호

- cache는 action에서 실제 materialize되며, memory에 들어가는 범위에서 재사용 이점이 보입니다.

### FAIL 신호

- cache가 항상 빠르다고 결론 냅니다. memory 부족, serialization, eviction, GC가 있으면 다를 수 있습니다.

## 13. Reality Scenario: Linux page cache 때문에 durable write가 헷갈린다

> 빠른 write latency는 durable write 증거가 아닙니다.
> Kafka/Cassandra 같은 system은 이 경계를 replication, commit log, fsync policy와 함께 다룹니다.

1. 관측된 증상

    write latency는 낮지만 crash 후 일부 data가 기대와 다르게 복구됩니다.

2. 가능한 원인 후보

    page cache writeback 전 crash, fsync policy, storage cache, app-level ack 위치, replica ack 오해.

3. OS/kernel 관점에서 볼 지점

    syscall trace, fsync 호출 여부, writeback, disk flush behavior.

4. distributed-system 관점에서 볼 지점

    quorum/replication ack가 어떤 failure model을 막는지 봅니다.

5. 시스템 내부 구조와 연결되는 지점

    Kafka broker log append and replication, Cassandra commit log sync, Spark checkpoint storage.

6. 확인 명령 또는 로그

    `strace -e trace=write,fsync`, broker/database config, application ack log.

7. 잘못된 결론의 예

    "`write()`가 성공했으니 storage는 안전하다."

8. 더 나은 추론 과정

    ack가 user/kernel/device/replica 중 어느 boundary 뒤에 찍혔는지 확인합니다.

## 14. Reality Scenario: JVM GC, memory pressure, disk saturation이 분산 증상으로 보인다

> JVM pause와 OS resource pressure는 distributed failure detector 입장에서는 느린 node처럼 보일 수 있습니다.
> 제품 metric과 OS/JVM metric을 시간순으로 맞춰 봐야 합니다.

1. 관측된 증상

    Kafka consumer rebalance, Cassandra timeout, Spark executor lost가 같은 시간대에 발생합니다.

2. 가능한 원인 후보

    GC pause, disk saturation, memory pressure, network retransmission, node overload.

3. OS/kernel 관점에서 볼 지점

    `vmstat`, `iostat`, process RSS, swap, socket queue.

4. distributed-system 관점에서 볼 지점

    heartbeat timeout, group rebalance, failure detector conviction, task retry.

5. 시스템 내부 구조와 연결되는 지점

    Kafka group coordinator, Cassandra gossip/failure detector, Spark executor heartbeat.

6. 확인 명령 또는 로그

    GC logs, `jcmd Thread.print`, product coordinator logs, OS metrics.

7. 잘못된 결론의 예

    "세 시스템이 각각 독립적으로 장애가 났다."

8. 더 나은 추론 과정

    같은 host/time window에서 resource pressure가 먼저 시작됐는지 확인하고, distributed symptom이 그 결과인지 분리합니다.

## Active Recall

1. `write()`와 `fsync()`를 syscall trace에서 각각 어떻게 찾을 수 있나요?
2. page cache 실험에서 두 번째 read가 빠르지 않아도 page cache 설명이 틀렸다고 단정하면 안 되는 이유는 무엇인가요?
3. `vmstat`, `iostat`, `ss`, `jcmd`가 각각 어느 계층의 queue나 state를 보여 주나요?
4. Kafka consumer lag 실험에서 lag가 topic 전체 상태가 아니라 group별 상태인 이유는 무엇인가요?
5. Cassandra `nodetool compact`를 운영 cluster에서 함부로 실행하면 왜 위험한가요?
6. Spark UI에서 shuffle bottleneck을 볼 때 stage 평균만 보면 왜 부족한가요?

## 근거와 더 읽을 자료

- [01_os_kernel_foundations.md](01_os_kernel_foundations.md)
- [02_distributed_system_foundations.md](02_distributed_system_foundations.md)
- [03_kafka_deep_dive.md](03_kafka_deep_dive.md)
- [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md)
- [05_spark_deep_dive.md](05_spark_deep_dive.md)
- [10_source_ledger.md](10_source_ledger.md)
