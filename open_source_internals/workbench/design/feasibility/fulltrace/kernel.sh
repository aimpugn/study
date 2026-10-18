#!/bin/bash
# 탐색용 코드입니다(2026-10-13, 커널 추적 G1~G5·A1~A3). 작업대 도구를 Go로 통일하므로, 이 순서를 Go 하위 명령으로 다시 만든 뒤 이 파일을 지웁니다.
#
# 하는 일: 기록 서버(serve.sh)를 띄우고, psql 접속 하나의 백엔드가 부른 pwrite64·pwritev·fdatasync 안에서 커널이 부른 함수를
#          OrbStack Linux 머신의 ftrace(function_graph)로 기록합니다. 컨테이너 쪽은 strace(-T)와 PostgreSQL 함수 기록이 함께 남습니다.
#          접속에서 친 문장은 SELECT pg_backend_pid(), INSERT 한 줄, CREATE INDEX이고, 문장마다 앞뒤로 백엔드의 /proc/<pid>/io와
#          블록 장치 통계(/sys/block/vd*/stat)를 남깁니다.
# 전제: Docker 컨테이너 pg-lab(serve.sh와 같음)과, root로 tracefs를 쓸 수 있는 OrbStack 머신 tracelab(orb create alpine tracelab).
#       ftrace는 VM의 커널 하나를 기록하므로 머신에서 컨테이너의 프로세스가 보입니다. pid는 컨테이너 안과 다릅니다.
# 실행: bash kernel.sh <출력 디렉터리>
# 결과: <출력>/kernel.trace(function_graph), all.strace, trace/<백엔드 pid>.bin, postgres(기록 빌드), forks.txt, psql_io.txt
# 해독: decode -bin <출력>/postgres -dir <출력>/trace journey -pid <백엔드 pid> -nth 2 -strace <출력>/all.strace \
#         -src <PostgreSQL 소스> -kernel <출력>/kernel.trace -ksrc <Linux v7.2.9 소스>
# 실패 신호: 접속 직전의 fork를 보지 못하면 1로 끝납니다. 끝에 출력되는 per_cpu 통계의 overrun이 0이 아니면 버퍼가 넘쳐 기록이 빠진 것입니다.
#
# 1차 실행의 교훈. set_graph_function은 최상위 추적기에만 있고 인스턴스(instances/*)에는 없습니다. 인스턴스로 시작하자 거르지 못한
# function_graph가 VM의 모든 함수를 기록해 32 MB 버퍼를 50 ms 만에 덮어썼습니다. 그래서 최상위를 쓰고, 세 시스템 콜 안만 기록하고,
# 버퍼가 차면 앞을 덮지 않고 멈추게(overwrite 0) 합니다. 바꾼 설정은 끝에서 원래 값으로 되돌립니다.
# syscall_user_buf_size를 0으로 두는 것은, 시스템 콜 이벤트가 사용자 버퍼의 앞부분(기본 63 B)을 기록에 복사하지 않게 하기 위해서입니다.
set -euo pipefail
OUT=${1:?출력 디렉터리를 주세요}
rm -rf "$OUT"; mkdir -p "$OUT"
SRV=$(cd "$(dirname "$0")" && pwd)/serve.sh
M() { orb -m tracelab -u root sh -c "$1"; }
C() { docker exec pg-lab bash -c "$1"; }
T=/sys/kernel/tracing

# 되돌릴 값을 이름 있는 변수로 저장합니다. 1차 판의 eval 반복은 값을 엉뚱한 파일에 써서 "Invalid argument"로 실패했습니다.
read -r BUF PROC ABST TAIL IRQS OVW UBUF <<< "$(M "cut -d' ' -f1 $T/buffer_size_kb; cat $T/options/funcgraph-proc $T/options/funcgraph-abstime \
  $T/options/funcgraph-tail $T/options/funcgraph-irqs $T/options/overwrite $T/syscall_user_buf_size" | tr '\n' ' ')"
echo "saved: buffer=$BUF proc=$PROC abstime=$ABST tail=$TAIL irqs=$IRQS overwrite=$OVW user_buf=$UBUF"
restore() {
  M "echo 0 > $T/tracing_on; echo nop > $T/current_tracer; echo > $T/set_graph_function; echo > $T/set_ftrace_pid; echo > $T/set_event_pid
     echo 0 > $T/events/enable; echo 0 > $T/events/sched/sched_process_fork/filter; echo $BUF > $T/buffer_size_kb
     echo $PROC > $T/options/funcgraph-proc; echo $ABST > $T/options/funcgraph-abstime; echo $TAIL > $T/options/funcgraph-tail
     echo $IRQS > $T/options/funcgraph-irqs; echo $OVW > $T/options/overwrite; echo $UBUF > $T/syscall_user_buf_size
     echo > $T/trace; echo 1 > $T/tracing_on" || echo "추적 설정을 되돌리지 못했습니다" >&2
}
trap restore EXIT

# 1. postgres가 fork하는 순간만 봅니다. 접속 직전에 비워, 그 사이의 fork만 후보로 남깁니다.
M "echo 0 > $T/tracing_on; echo nop > $T/current_tracer; echo > $T/trace; echo 0 > $T/syscall_user_buf_size
   echo 'parent_comm == \"postgres\"' > $T/events/sched/sched_process_fork/filter; echo 1 > $T/events/sched/sched_process_fork/enable; echo 1 > $T/tracing_on"

# 2. 기록 서버를 띄우고, psql을 FIFO에 물려 접속 하나를 계속 쓰게 합니다. 문장은 FIFO로 하나씩 보냅니다.
docker cp "$SRV" pg-lab:/tmp/serve.sh >/dev/null
C "bash /tmp/serve.sh"
M "echo > $T/trace"
C "rm -f /tmp/pgi/q; mkfifo /tmp/pgi/q; (setsid sleep 600 > /tmp/pgi/q &) ; (setsid psql -X -h /tmp/pgi -p 5433 -d postgres -e < /tmp/pgi/q > /tmp/pgi/psql.out 2>&1 &)"
# 문장 앞뒤의 \! 줄은 psql이 셸에서 실행합니다. BPID는 첫 문장이 \setenv로 정한 백엔드 pid입니다.
SNAP='sh -c '\''echo == %s; cat /proc/$BPID/io; for d in vda vdb vdc; do echo $d $(cat /sys/block/$d/stat); done'\'' >> /tmp/pgi/io.log'
cat > "$OUT/s0.sql" <<'SQL'
SELECT pg_backend_pid() AS bpid \gset
\setenv BPID :bpid
\echo bpid :bpid
SQL
{ printf "\\\\! $SNAP\n" before-insert; echo "INSERT INTO t VALUES (4,'d');"; printf "\\\\! $SNAP\n" after-insert; } > "$OUT/s1.sql"
{ echo "CREATE INDEX t_name_idx ON t(name);"; printf "\\\\! $SNAP\n" after-index; } > "$OUT/s2.sql"
for f in s0 s1 s2; do docker cp "$OUT/$f.sql" pg-lab:/tmp/pgi/$f.sql >/dev/null; done
C "cat /tmp/pgi/s0.sql > /tmp/pgi/q"
sleep 2
M "grep -v '^#' $T/trace | grep sched_process_fork" > "$OUT/forks.txt" || true
cat "$OUT/forks.txt"
G=$(sed -E 's/.*child_pid=([0-9]+).*/\1/' "$OUT/forks.txt" | tr '\n' ' ')
echo "candidate kernel pids: $G"
[ -n "$G" ] || { echo "fork를 보지 못했습니다" >&2; exit 1; }

# 3. 후보 pid만, 세 시스템 콜 안의 함수 경로만 기록합니다. 인터럽트 처리는 빼고, 줄마다 시각과 프로세스를 붙입니다.
#    시스템 콜 이벤트(sys_enter_*)의 인자 fd·count·pos가 strace의 같은 호출과 짝짓는 열쇠입니다(decode/kernel.go matchKernel).
M "echo 0 > $T/tracing_on; echo > $T/trace; echo 0 > $T/events/sched/sched_process_fork/enable; echo 65536 > $T/buffer_size_kb
   echo $G > $T/set_ftrace_pid; echo $G > $T/set_event_pid
   echo '__arm64_sys_pwrite64 __arm64_sys_pwritev __arm64_sys_fdatasync' > $T/set_graph_function
   for e in sys_enter_pwrite64 sys_exit_pwrite64 sys_enter_pwritev sys_exit_pwritev sys_enter_fdatasync sys_exit_fdatasync; do echo 1 > $T/events/syscalls/\$e/enable; done
   echo function_graph > $T/current_tracer
   for o in funcgraph-proc funcgraph-abstime funcgraph-tail; do echo 1 > $T/options/\$o; done; echo 0 > $T/options/funcgraph-irqs; echo 0 > $T/options/overwrite
   cat $T/current_tracer; cat $T/set_graph_function | tr '\n' ' '; echo"

# 4. 문장 둘을 보냅니다.
M "echo 1 > $T/tracing_on"
C "cat /tmp/pgi/s1.sql > /tmp/pgi/q"; sleep 2
C "cat /tmp/pgi/s2.sql > /tmp/pgi/q"; sleep 3
M "echo 0 > $T/tracing_on; grep -E 'entries|overrun|dropped' $T/per_cpu/cpu0/stats"

# 5. 모읍니다. 서버를 내리고 함수 기록을 지웁니다. pkill은 패턴을 줄 머리와 끝에 묶어, 이 명령을 실행하는 셸 자신을 죽이지 않게 합니다.
M "cat $T/trace" > "$OUT/kernel.trace"
C "cat /tmp/pgi/psql.out; echo; cat /tmp/pgi/io.log" > "$OUT/psql_io.txt"
BP=$(C "sed -n 's/^bpid \([0-9]*\)$/\1/p' /tmp/pgi/psql.out")
echo "backend container pid: $BP"
C "pkill -f '^sleep 600\$' || true; pg_ctl -D /tmp/pgi/data stop -m fast >/dev/null; sleep 1"
docker cp pg-lab:/tmp/pgi/all.strace "$OUT/all.strace" >/dev/null
mkdir -p "$OUT/trace"; docker cp pg-lab:/tmp/pgtrace/$BP.bin "$OUT/trace/$BP.bin" >/dev/null
docker cp pg-lab:/tmp/pgi/bin/postgres "$OUT/postgres" >/dev/null
C "du -sh /tmp/pgtrace; rm -rf /tmp/pgtrace; ps -eo args | grep -cE '^(strace -f|/tmp/pgi/bin)' || true"
wc -l "$OUT/kernel.trace"; ls -la "$OUT"
