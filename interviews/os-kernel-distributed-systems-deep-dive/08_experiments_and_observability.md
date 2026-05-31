# 08. Experiments and Observability

이 문서의 실험은 benchmark 경쟁이 아니라 개념 검증입니다. 숫자가 예쁘게 나오지 않아도 괜찮습니다. 중요한 것은 "문서에서 말한 상태 이동이 실제 관찰 표면에서 어떻게 보이는가"입니다.

모든 실험은 개인 local 환경에서 실행하는 것을 전제로 합니다. production, shared server, 회사 개발 cluster에서 그대로 실행하지 않습니다. 특히 CPU 부하, disk I/O, Cassandra compaction, Spark shuffle은 주변 작업에 피해를 줄 수 있습니다.

## 도구를 고르는 기준

| 보고 싶은 상태 | Linux 예 | macOS 예 | 주의 |
|---|---|---|---|
| system call | `strace` | `dtruss`, `dtrace` | macOS SIP/권한 제한 가능 |
| CPU runnable/context switch | `vmstat`, `pidstat` | `top`, Activity Monitor | 필드 의미가 다름 |
| disk I/O | `iostat -xz` | `iostat -w` | 옵션 차이 큼 |
| socket queue | `ss -tin`, `netstat` | `netstat` | `ss`는 Linux 중심 |
| 열린 파일 | `lsof` | `lsof` | 권한 필요 |
| JVM thread | `jcmd`, `jstack` | `jcmd`, `jstack` | JDK 필요 |
| Kafka | `kafka-consumer-groups.sh`, broker metrics | same | local cluster 필요 |
| Cassandra | `nodetool`, `cqlsh` | same | 명령 부작용 주의 |
| Spark | Spark UI, event log, `spark-submit` | same | local mode와 cluster mode 차이 |

## 1. `write()`와 `fsync()`가 다른 system call임을 본다

목적은 파일 쓰기가 커널 system call로 내려가고, `write()`와 `fsync()`가 별도 요청임을 확인하는 것입니다.

전제는 Linux에 `strace`가 설치되어 있거나, macOS에서 `dtruss` 권한이 있는 것입니다.

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

예상 관찰은 `write(...)`와 `fsync(...)`가 따로 보이는 것입니다. PASS는 두 syscall이 분리되어 보이고, `write()`만으로 fd sync를 증명하지 않는 것입니다. FAIL은 tracing 권한 문제를 개념 실패로 해석하거나, `write()`가 보였다는 이유로 storage durability를 증명했다고 말하는 것입니다.

macOS에서는 다음과 같이 시도할 수 있지만, SIP와 권한 제한 때문에 실패할 수 있습니다.

```bash
sudo dtruss -t write -t fsync -o /tmp/write-trace.log python3 - <<'PY'
import os

fd = os.open("/tmp/durable-demo.txt", os.O_CREAT | os.O_WRONLY | os.O_TRUNC, 0o644)
os.write(fd, b"hello\n")
os.fsync(fd)
os.close(fd)
PY
```

## 2. page cache 효과를 조심스럽게 본다

목적은 같은 파일을 반복 읽을 때 두 번째 read가 page cache 덕분에 빠를 수 있음을 관찰하는 것입니다.

전제는 local disk 공간이 충분하고 `/tmp`가 memory filesystem이 아닌 환경입니다.

```bash
dd if=/dev/zero of=/tmp/page-cache-demo.bin bs=1m count=256
sync
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
rm -f /tmp/page-cache-demo.bin
```

PASS는 두 번째 read 시간이 줄거나 device read 증가가 작게 보이는 것입니다. FAIL은 차이가 없다고 page cache가 없다고 결론 내리는 것입니다. SSD, memory 크기, 이미 warm cache였는지, OS 정책 때문에 차이는 작을 수 있습니다. Linux의 `/proc/sys/vm/drop_caches`는 system-wide 영향이 있으므로 이 문서에서는 기본 실험으로 사용하지 않습니다.

## 3. runnable queue와 context switch를 안전하게 만든다

이 실험은 CPU-bound process를 몇 개 띄운 뒤 `vmstat`에서 runnable queue와 context switch 변화를 보는 것입니다. 코드 문자열로 광범위하게 process를 찾아 종료하는 방식은 다른 Python 작업까지 죽일 수 있으므로 사용하지 않습니다. 여기서는 PID를 기록하고 trap으로 정리합니다.

Linux:

```bash
tmpdir="$(mktemp -d)"
cleanup() {
  if [ -f "$tmpdir/pids" ]; then
    while read -r pid; do
      kill "$pid" 2>/dev/null || true
    done < "$tmpdir/pids"
  fi
  rm -rf "$tmpdir"
}
trap cleanup EXIT

python3 - <<'PY' &
while True:
    pass
PY
echo "$!" >> "$tmpdir/pids"

python3 - <<'PY' &
while True:
    pass
PY
echo "$!" >> "$tmpdir/pids"

vmstat 1 5
cleanup
trap - EXIT
```

PASS는 runnable queue(`r`)나 context switch(`cs`)가 변하는 것을 보는 것입니다. FAIL은 shared machine에서 CPU-bound loop를 오래 방치하거나 PID 없이 광범위한 kill pattern을 쓰는 것입니다.

macOS에서는 Activity Monitor나 `top -o cpu`로 CPU-bound process를 확인하고, 같은 PID 정리 원칙을 적용합니다.

## 4. socket queue와 backpressure 감각을 본다

목적은 receiver가 읽지 않을 때 sender가 결국 kernel socket buffer와 TCP flow control의 영향을 받는다는 점을 관찰하는 것입니다.

Terminal 1:

```bash
python3 - <<'PY'
import socket, time

s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("127.0.0.1", 18080))
s.listen(1)
conn, _ = s.accept()
time.sleep(30)
conn.close()
s.close()
PY
```

Terminal 2:

```bash
python3 - <<'PY'
import socket

s = socket.create_connection(("127.0.0.1", 18080))
payload = b"x" * 1024 * 1024
for _ in range(64):
    s.sendall(payload)
input("press enter to close")
s.close()
PY
```

Terminal 3:

```bash
# Linux
ss -tin sport = :18080 or dport = :18080

# macOS fallback
netstat -an | grep 18080
```

PASS는 TCP connection과 send/receive queue 관련 정보가 보이는 것입니다. FAIL은 loopback 결과를 실제 datacenter latency와 동일하게 해석하는 것입니다.

## 5. JVM thread dump로 waiting과 blocked를 본다

목적은 JVM 기반 Kafka/Cassandra/Spark에서 CPU가 낮아도 thread가 lock이나 sleep, I/O에서 기다릴 수 있음을 보는 것입니다.

전제는 JDK가 있고 Java process가 실행 중인 것입니다.

```bash
jcmd
jcmd <pid> Thread.print | sed -n '1,120p'
```

PASS는 `RUNNABLE`, `WAITING`, `TIMED_WAITING`, `BLOCKED` 상태를 구분해 보는 것입니다. FAIL은 모든 `RUNNABLE`을 CPU 실행 중으로 해석하는 것입니다. JVM의 `RUNNABLE`은 native I/O wait를 포함해 OS CPU running과 정확히 같지 않을 수 있습니다.

## 6. epoll readiness와 non-blocking read를 본다

목적은 epoll이 "작업 완료"가 아니라 "fd에서 읽을 수 있을 가능성"을 알려 주며, application이 non-blocking read를 반복해 `EAGAIN`까지 drain해야 한다는 점을 확인하는 것입니다.

전제는 Linux와 Python 3입니다. macOS의 `selectors`는 kqueue로 매핑될 수 있으므로 결과 표현은 다르지만 readiness 모델을 관찰할 수 있습니다.

```bash
python3 - <<'PY'
import os, select

r, w = os.pipe()
os.set_blocking(r, False)

ep = select.epoll()
ep.register(r, select.EPOLLIN)

os.write(w, b"hello")
print("events:", ep.poll(0))
print("read:", os.read(r, 2))
print("events after partial read:", ep.poll(0))
print("read rest:", os.read(r, 1024))
try:
    os.read(r, 1)
except BlockingIOError as e:
    print("EAGAIN after drain:", e.errno)
PY
```

PASS는 event가 readiness를 알려 주고, 일부만 읽으면 아직 읽을 byte가 남을 수 있으며, 완전히 drain한 뒤 non-blocking read가 `EAGAIN`으로 돌아오는 것을 보는 것입니다. FAIL은 epoll event를 "read 작업이 끝났다"로 해석하는 것입니다. 이 실험은 pipe로 readiness 감각을 보는 것이며, TCP packet path 전체를 증명하지는 않습니다.

## 7. mmap과 page fault 감각을 본다

목적은 `mmap()`이 파일 byte를 process address space에 보이게 만들지만, 실제 page는 첫 접근 시점의 page fault와 page cache를 통해 올라올 수 있음을 관찰하는 것입니다.

Linux:

```bash
python3 - <<'PY' &
import mmap, os, time

path = "/tmp/mmap-demo.bin"
with open(path, "wb") as f:
    f.truncate(64 * 1024 * 1024)

with open(path, "r+b") as f:
    m = mmap.mmap(f.fileno(), 0)
    print("pid", os.getpid(), flush=True)
    time.sleep(5)
    total = 0
    for i in range(0, len(m), 4096):
        total += m[i]
    print("touched", total)
    time.sleep(5)
PY
pid=$!
sleep 1
awk '{print "before minor_faults="$10, "major_faults="$12}' /proc/$pid/stat 2>/dev/null || true
sleep 7
awk '{print "after minor_faults="$10, "major_faults="$12}' /proc/$pid/stat 2>/dev/null || true
wait "$pid"
rm -f /tmp/mmap-demo.bin
```

PASS는 mmap 후 실제 접근 시점에 page fault 관련 counters가 움직일 수 있음을 이해하는 것입니다. FAIL은 counter 해석이 어렵다고 mmap과 page fault 관계가 없다고 결론 내리는 것입니다. `/proc/<pid>/stat` field 해석은 Linux 문서를 확인해야 하며, 이 명령은 학습용 힌트입니다.

## 8. container/cgroup memory limit 관측 위치를 확인한다

목적은 container나 cgroup 환경에서 process가 보는 memory 문제가 host 전체 memory와 다를 수 있음을 확인하는 것입니다.

Linux cgroup v2가 있는 환경:

```bash
test -f /sys/fs/cgroup/memory.current && cat /sys/fs/cgroup/memory.current
test -f /sys/fs/cgroup/memory.max && cat /sys/fs/cgroup/memory.max
cat /proc/self/cgroup
```

PASS는 현재 process가 속한 cgroup과 memory limit/current 사용량을 확인하는 것입니다. FAIL은 host의 `free` 결과만 보고 container OOM 가능성을 배제하는 것입니다. macOS Docker Desktop에서는 Linux VM 내부에서 확인해야 하므로 host 터미널 결과와 다를 수 있습니다.

## 9. perf/off-CPU 관측의 입구를 확인한다

목적은 CPU flame graph만으로는 sleep, lock, I/O wait를 설명하지 못한다는 점을 확인하는 것입니다.

Linux:

```bash
perf stat -e context-switches,cpu-migrations,page-faults sleep 2
```

가능하면 별도 test process에 대해 `perf record`나 eBPF/bpftrace를 사용할 수 있지만, 권한과 kernel 설정이 필요합니다.

PASS는 context switch와 page fault 같은 event가 CPU 사용률과 별개로 관찰될 수 있음을 보는 것입니다. FAIL은 production에서 권한/overhead 검토 없이 profiling을 켜는 것입니다.

## 10. Kafka local 관찰

목적은 topic partition과 consumer group offset/lag가 실제로 분리된 상태임을 보는 것입니다. 전제는 local Kafka cluster가 있고 실험용 topic을 사용한다는 것입니다.

```bash
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic demo-topic
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group demo-group
```

PASS는 partition별 current offset, log end offset, lag를 구분해 보는 것입니다. FAIL은 lag를 곧바로 consumer thread 부족으로 단정하는 것입니다. broker disk, partition skew, downstream processing이 모두 후보입니다.

## 11. Cassandra local 관찰

목적은 Cassandra node의 상태와 compaction/repair 관련 신호를 읽는 것입니다. 전제는 개인 local Cassandra cluster입니다.

안전한 read-only 관찰:

```bash
nodetool status
nodetool tpstats
nodetool compactionstats
```

`nodetool compact`는 정상 검증 명령이 아닙니다. local disposable cluster에서 compaction 효과를 일부러 관찰할 때만 opt-in으로 실행합니다. shared cluster나 production에서 실행하면 disk I/O를 크게 만들 수 있습니다.

PASS는 pending compaction, thread pool pending/blocked, node status를 읽고 어떤 queue가 자라는지 말하는 것입니다. FAIL은 `compact`를 일반 점검 명령처럼 실행하는 것입니다.

## 12. Spark local 관찰

목적은 action이 job/stage/task로 나뉘고, shuffle이 stage boundary와 spill metric을 만든다는 점을 보는 것입니다.

local mode 예시:

```bash
spark-submit --master local[2] examples/src/main/python/wordcount.py /tmp/input.txt
```

실제 path는 설치 방식마다 다릅니다. Spark UI가 떠 있다면 `Jobs`, `Stages`, `Executors`, `SQL` 탭에서 task duration, shuffle read/write, spill, GC time을 봅니다.

PASS는 job -> stage -> task 계층과 shuffle metric을 구분하는 것입니다. FAIL은 local mode 결과를 cluster scheduling, remote shuffle, executor loss와 동일하게 해석하는 것입니다.

## 13. packet/request path는 한 명령으로 증명되지 않는다

목적은 NIC, TCP, socket, epoll, application parser가 여러 관측 표면에 걸쳐 있다는 점을 확인하는 것입니다. local loopback으로 `tcpdump`나 `ss`를 볼 수 있지만, loopback은 실제 NIC DMA/NAPI path를 그대로 지나지 않을 수 있습니다.

안전한 local 관찰:

```bash
python3 -m http.server 18081 >/tmp/http-demo.log 2>&1 &
pid=$!
sleep 1
curl -s http://127.0.0.1:18081/ >/dev/null
ss -tin sport = :18081 or dport = :18081 2>/dev/null || netstat -an | grep 18081 || true
kill "$pid"
rm -f /tmp/http-demo.log
```

PASS는 socket connection 관측이 application request 처리와 별도 표면임을 이해하는 것입니다. FAIL은 loopback 실험으로 NIC DMA ring, hardware interrupt, real network latency를 모두 증명했다고 말하는 것입니다.

## 실험 후 기록할 것

각 실험 뒤에는 다음 네 줄을 적습니다.

```text
observed state:
where it sits:
what changed:
what I must not overclaim:
```

예를 들어 `write()` 실험 뒤에는 `observed state: write and fsync syscalls`, `where it sits: process -> kernel syscall boundary`, `what changed: file bytes accepted then fd sync requested`, `what I must not overclaim: write alone proves durable storage`처럼 씁니다.

## 14. `wc`, `du`, `df`는 서로 다른 질문에 답한다

문서 corpus 검증에서도 쓰지만, 운영에서도 크기 숫자는 질문이 다릅니다. `wc -c`는 file byte 수를 봅니다. `du`는 filesystem block 사용량을 봅니다. `df`는 filesystem 전체 free space를 봅니다. Sparse file, compression, block size, hard link, reflink, snapshot, deleted-but-open file 때문에 숫자가 다를 수 있습니다. Kafka log segment, Cassandra SSTable, Spark spill file을 볼 때도 같은 차이가 나타납니다.

안전한 local 관찰:

```bash
tmp=$(mktemp -d)
printf 'hello\n' > "$tmp/a.txt"
ln "$tmp/a.txt" "$tmp/b.txt"
wc -c "$tmp/a.txt" "$tmp/b.txt"
du -sk "$tmp"
df -k "$tmp"
rm -rf "$tmp"
```

PASS는 "파일 내용 크기", "디스크 block 사용량", "filesystem 여유 공간"을 구분하는 것입니다. FAIL은 `df`가 충분하니 특정 process write가 안전하다고 단정하거나, `du`만 보고 open-deleted file이나 quota를 놓치는 것입니다.

## 15. cgroup 관측은 host와 process 사이의 중간 계층이다

컨테이너 환경에서는 host 전체 metric과 process metric 사이에 cgroup이 있습니다. CPU throttling, memory OOM, I/O pressure는 cgroup 단위로 나타날 수 있습니다. Host CPU가 남아도 container quota가 다 되면 throttling되고, host memory가 남아도 cgroup limit을 넘으면 process가 kill될 수 있습니다.

관찰 경로는 배포 환경에 따라 다르지만 Linux에서는 보통 `/sys/fs/cgroup` 아래에서 cpu, memory event를 볼 수 있습니다. Kubernetes에서는 `kubectl top`, container runtime metric, kubelet event, pod status reason을 함께 봅니다. PASS는 "host는 여유인데 container가 제한에 걸릴 수 있다"를 이해하는 것입니다. FAIL은 container OOM을 JVM heap OOM과 같은 것으로 기록하는 것입니다.

## 16. source ledger reachability는 사실성의 전부가 아니다

Source URL이 200을 반환한다고 본문 claim이 자동으로 맞는 것은 아닙니다. Reachability는 자료에 접근 가능하다는 최소 조건입니다. 본문 claim이 source의 어느 부분에 기대는지, version-sensitive한지, 공식 문서인지, 논문인지, man page인지, 추론인지 따로 적어야 합니다. 이 corpus의 `10_source_ledger.md`가 link list가 아니라 claim-level ledger인 이유입니다.

간단한 reachability check:

```bash
while read -r url; do
  curl -I -L --max-time 10 "$url" >/dev/null && echo "OK $url" || echo "FAIL $url"
done <<'URLS'
https://man7.org/linux/man-pages/man2/write.2.html
https://docs.kernel.org/mm/page_cache.html
https://kafka.apache.org/documentation/
https://cassandra.apache.org/doc/latest/
https://spark.apache.org/docs/latest/
URLS
```

PASS는 URL 접근성과 claim support tier를 분리하는 것입니다. FAIL은 link가 살아 있다는 이유로 제품 version별 동작이나 kernel implementation detail을 확정하는 것입니다.

## 17. 관측 장부를 쓰는 이유

실험은 기억보다 기록이 중요합니다. 같은 명령을 실행해도 kernel version, filesystem, container permission, hardware, load, product version에 따라 결과가 달라집니다. 그래서 결과를 "정답"으로 적기보다 관측 조건과 한계를 적어야 합니다.

```text
experiment:
environment:
command:
observed state:
where it sits:
claim supported:
claim not supported:
next check:
```

이 형식은 OS와 분산 시스템 모두에 유용합니다. Kafka broker 하나를 멈춘 실험은 leader election과 producer behavior를 일부 보여 주지만 모든 outage를 증명하지 않습니다. Cassandra node 하나를 끊은 실험은 CL별 behavior를 보여 주지만 long partition, repair, clock skew를 모두 증명하지 않습니다. Spark executor kill 실험은 task retry를 보여 주지만 external sink 중복을 자동으로 검증하지 않습니다. 관측 장부는 배운 것을 과장하지 않게 해 줍니다.

## 18. 작은 실험을 제품 장애로 이어 붙이기

이 문서의 실험들은 작습니다. 작은 `write()` 실험은 Kafka flush policy 전체를 증명하지 않습니다. 작은 socket 실험은 real NIC DMA/NAPI를 증명하지 않습니다. 작은 Spark local mode 실험은 cluster shuffle을 증명하지 않습니다. 그래도 가치가 있습니다. 각 실험은 큰 시스템의 한 층을 손으로 만져 보게 해 줍니다.

```
small experiment
  -> reveals one lower-layer mechanism
  -> records what it does not prove
  -> becomes a question to ask in production
```

예를 들어 page cache 실험 뒤에는 Kafka segment read에서 page cache hit/miss를 어떻게 추론할지 묻습니다. Futex/lock 실험 뒤에는 JVM thread dump와 off-CPU profile을 어떻게 볼지 묻습니다. Socket queue 실험 뒤에는 Kafka fetch timeout이나 Cassandra read timeout에서 `ss`와 packet capture를 어떻게 함께 볼지 묻습니다. Spark local shuffle 실험 뒤에는 cluster에서 shuffle read blocked time과 disk spill을 어떻게 볼지 묻습니다.

## 19. 마지막 점검표

실험을 마친 뒤에는 세 가지를 꼭 확인합니다. 첫째, 내가 직접 관측한 것은 무엇인가. 둘째, 그 관측이 어느 계층의 claim만 지지하는가. 셋째, 제품이나 분산 환경으로 확대하려면 어떤 추가 증거가 필요한가. 이 세 가지가 없으면 실험은 배움을 주는 대신 과신을 만들 수 있습니다.

```
direct observation
  -> local command output

supported claim
  -> narrow mechanism

not yet supported
  -> production behavior, version-specific guarantee, distributed recovery
```

이 구분을 지키면 실험 문서는 단순 명령 모음이 아니라 reasoning 도구가 됩니다. 면접에서도 "제가 직접 증명한 범위는 여기까지이고, 실제 운영에서는 이 지표를 추가로 보겠습니다"라고 말할 수 있습니다. 그것이 OS와 분산 시스템을 공부하는 가장 안전한 태도입니다.

## 20. Kafka 관측 rehearsal

실제 Kafka cluster가 없는 환경에서도 관측 질문은 미리 연습할 수 있습니다. Produce latency를 받으면 `request queue`, `local append`, `remote replication`, `response send`로 나눕니다. Consumer lag를 받으면 `broker fetch`, `consumer processing`, `downstream sink`, `partition skew`, `rebalance`로 나눕니다. Cluster가 있다면 broker metrics, client metrics, disk/network/GC/cgroup metric을 같은 timeline에 놓습니다.

기록 예시는 다음과 같습니다.

```text
observed state:
  consumer lag for partition 7 grows faster than others
where it sits:
  Kafka partition-level progress, not yet root cause
what changed:
  partition 7 input rate or processing time may differ
what I must not overclaim:
  adding consumers helps only if partition assignment and downstream allow it
next check:
  per-partition rate, consumer processing time, rebalance log, downstream latency
```

이 rehearsal은 실제 장애에서 질문 순서를 줄여 줍니다. Lag 숫자 하나를 보고 consumer 수부터 늘리지 않고, partition과 queue owner를 먼저 찾게 됩니다.

## 21. Cassandra 관측 rehearsal

Cassandra cluster가 있다면 `nodetool tpstats`, table metrics, client latency, GC log, disk I/O, network metric을 같이 봅니다. Cluster가 없다면 trace를 종이에 그리는 것만으로도 효과가 있습니다. Write timeout은 coordinator가 CL을 만족할 enough response를 받지 못했다는 뜻입니다. Read timeout은 replica read path, coordinator reconciliation, network, GC, disk, compaction과 모두 양립합니다.

```text
observed state:
  write timeout at CL=QUORUM
where it sits:
  coordinator did not receive enough replica responses before deadline
what changed:
  one or more replicas may be slow or unavailable
what I must not overclaim:
  write definitely failed, or definitely succeeded everywhere
next check:
  coordinator logs, replica logs, pending tasks, commitlog sync, GC, iostat, network
```

Compaction rehearsal도 해 봅니다. Pending compaction이 늘면 foreground read/write와 같은 disk와 CPU를 공유하는지 봅니다. Tombstone warning이 있으면 delete/TTL pattern과 repair/compaction window를 봅니다. Cassandra는 product metric과 OS metric이 강하게 붙어 있으므로, 한쪽만 보면 원인을 과하게 단순화하기 쉽습니다.

## 22. Spark 관측 rehearsal

Spark가 있는 환경에서는 Spark UI가 가장 좋은 시작점입니다. Jobs, Stages, Tasks, Executors, SQL 탭을 보며 stage별 duration, task skew, shuffle read/write, spill, GC time, scheduler delay, executor lost reason을 확인합니다. Cluster가 없다면 local mode로 DAG와 shuffle 개념만 익히고, local mode가 cluster network와 executor loss를 증명하지 않는다는 한계를 적습니다.

```text
observed state:
  one task in a stage is much slower than the rest
where it sits:
  Spark stage/task distribution
what changed:
  partition skew, bad executor, slow disk/network, GC, cgroup throttle could fit
what I must not overclaim:
  more executors will fix it
next check:
  task input size, shuffle read time, spill, GC, executor host metrics, skewed key
```

Structured Streaming rehearsal에서는 batch duration, input rate, processed rows per second, state operator size, checkpoint latency, sink commit latency를 봅니다. Batch가 trigger interval보다 오래 걸리면 backlog가 쌓입니다. Source rate를 낮추는 것이 답일 수도 있고, state/shuffle/sink를 고치는 것이 답일 수도 있습니다.

## 23. 관측 timeline을 맞추는 법

여러 도구를 함께 볼 때 가장 흔한 실수는 시간이 어긋난 지표를 한 원인처럼 붙이는 것입니다. Application latency spike가 10:00:05에 있었는데 disk metric은 10:02 평균이고 GC log는 timezone이 다르면 잘못된 결론이 나옵니다. 관측은 같은 시간축에 놓아야 합니다.

```text
timeline
  10:00:00 producer latency starts rising
  10:00:03 broker request queue time rises
  10:00:04 disk await rises
  10:00:06 consumer lag rises
  10:00:10 retries increase
```

이 timeline은 causality의 증명은 아니지만, 가설을 정렬합니다. Disk await가 먼저인지 request queue가 먼저인지, retry가 원인인지 결과인지, GC pause가 spike와 겹치는지 볼 수 있습니다. 분산 시스템에서는 clock skew가 있으므로 node별 timestamp와 timezone, NTP 상태도 조심합니다. 관측 timeline은 "동시에 보였다"와 "원인이다"를 구분하게 해 줍니다.

## 24. 안전한 실험과 위험한 실험

Local file write, loopback socket, small Spark local job 같은 실험은 비교적 안전합니다. Production broker kill, Cassandra node network partition, disk fill, clock skew injection은 위험할 수 있습니다. 장애 주입은 sandbox나 staging, 명시적 승인, rollback, blast radius, monitoring, stop condition이 있어야 합니다. 이 문서는 학습 실험을 제공하지만, production chaos experiment를 무단으로 실행하라는 뜻이 아닙니다.

안전한 실험도 side effect를 남길 수 있습니다. 임시 파일을 지우고, background process를 종료하고, port 충돌을 피하고, 권한이 필요한 명령을 dry-run으로 확인해야 합니다. `tcpdump`, `perf`, eBPF는 권한과 privacy 문제가 있습니다. Packet capture는 민감한 payload를 포함할 수 있으므로 목적과 저장 위치, 삭제 기준을 정해야 합니다.

## 25. 마지막 replay: 관측을 주장으로 바꾸는 단계

관측은 claim 후보입니다. `iostat`에서 await가 높았다고 바로 "디스크가 원인"이라고 확정하지 않습니다. 그 시간에 application request가 disk I/O를 기다렸는지, queue가 어떤 path에서 커졌는지, 다른 지표와 맞는지 봅니다. `ss`에서 send queue가 컸다고 바로 "네트워크가 원인"이라고 하지 않습니다. 상대가 읽지 못하는지, local event loop가 쓰지 못하는지, downstream이 막혔는지 봅니다.

```
observation
  -> candidate mechanism
  -> alternative explanations
  -> next measurement
  -> supported claim with boundary
```

이 단계가 닫히면 실험 문서는 면접 답변으로도 바뀝니다. "제가 본 것은 X이고, 이는 Y 경로를 지지하지만 Z까지 증명하지는 않습니다. 그래서 다음에는 A를 보겠습니다." 이 문장은 OS, Kafka, Cassandra, Spark 어디서나 통합니다. 실험의 목표는 더 많은 명령을 외우는 것이 아니라, 관측에서 과장 없는 claim을 만드는 습관입니다.

## 26. 실험을 다시 읽는 순서

이 문서를 한 번에 다 실행하려 하지 않아도 됩니다. 먼저 OS 단일 머신 실험으로 `write`, `fsync`, socket queue, process/thread, memory 지표를 익힙니다. 그 다음 Kafka/Cassandra/Spark rehearsal로 product metric을 같은 시간축에 올려 봅니다. 마지막으로 source ledger reachability와 claim boundary를 점검합니다. 이 순서가 좋은 이유는 작은 관측에서 시작해 분산 시스템 claim으로 올라가기 때문입니다.

```
local mechanism
  -> product symptom
  -> distributed uncertainty
  -> claim boundary
```

각 실험을 마친 뒤에는 "이 실험이 실제 production에서 무엇을 바로 증명하지 못하는가"를 꼭 적습니다. Local loopback은 real NIC를 증명하지 않습니다. Local filesystem은 cloud block storage를 증명하지 않습니다. Spark local mode는 cluster shuffle을 증명하지 않습니다. 그러나 이 한계를 적는 순간, 작은 실험은 오히려 더 좋은 학습 도구가 됩니다.

## 27. 면접에서 실험을 말하는 법

면접에서 명령어를 줄줄 말하기보다, 실험이 어떤 claim을 가르는지 말하는 편이 좋습니다. "`strace`로 `write`와 `fsync` 호출을 나눠 보고, `iostat`으로 device I/O를 보며, Kafka metric으로 produce ack latency를 함께 보겠습니다"처럼 계층을 연결합니다. Cassandra라면 `nodetool`과 GC log, disk I/O, network를 함께 보고, Spark라면 Spark UI와 executor OS metric을 함께 봅니다.

이렇게 말하면 도구 이름이 목적이 아니라는 점이 드러납니다. 목적은 request가 어느 queue에서 기다렸는지, 어떤 boundary까지 성공했는지, 어떤 retry가 중복을 만들 수 있는지 확인하는 것입니다. 도구는 그 질문에 답하기 위한 관측 표면입니다.

## 28. 이 문서의 완료 기준

이 실험 문서는 모든 환경에서 같은 숫자를 만들기 위한 문서가 아닙니다. 완료 기준은 독자가 각 실험의 위치와 한계를 설명할 수 있는 것입니다. `write()` 실험 뒤에는 durable storage를 과장하지 않아야 합니다. Socket 실험 뒤에는 loopback과 NIC path를 구분해야 합니다. Spark local 실험 뒤에는 cluster scheduling과 remote shuffle을 구분해야 합니다. Source check 뒤에는 URL reachability와 claim support를 구분해야 합니다.

이 기준을 통과하면 실험은 단순한 command collection이 아니라 OS와 분산 시스템을 검증하는 사고 훈련이 됩니다. 관측을 좁게 말하고, 필요한 다음 관측을 정직하게 남기는 습관이 이 corpus 전체의 verification language입니다.
