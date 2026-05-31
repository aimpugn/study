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
