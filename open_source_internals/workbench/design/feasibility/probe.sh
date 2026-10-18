#!/bin/sh
# 탐색용 코드입니다(2026-10-13, 작업대 관측 재설계의 가능성 시험 P1~P11의 재현 경로). 사용자 지시에 따라 작업대 도구는 Go로 통일하므로, 같은 기록을 Go(GDB/MI)로 다시 만든 뒤 이 파일을 지웁니다.
# 가능성 시험 구동기: 작업대와 같은 세션(데우기 뒤 추적)에서 probe.py로 백엔드를 보고, I/O 워커는 strace로 따로 봅니다.
# 사용: probe.sh cold|warm   (cold는 추적 직전에 테이블 t를 공유 버퍼에서 뺍니다)
# 전제: start-lab.sh로 띄운 pg-lab, 그 안에 설치한 pg_buffercache
#       (docker exec -u root pg-lab make -C /src/postgresql-18.6/contrib/pg_buffercache install 뒤 CREATE EXTENSION pg_buffercache)
# 출력: 백엔드의 gdb 기록(STOP, SYSCALL, SYSRET, BUF, PAGE, LP, MCX, CUR)과 I/O 워커의 strace 기록
set -eu
mode="$1"
here=$(cd "$(dirname "$0")" && pwd)
lab="$here/../../../database/postgresql/scripts"
. "$lab/session.sh"

docker cp "$here/probe.py" pg-lab:/tmp/probe.py
docker exec -u root pg-lab rm -f /tmp/gdb-ready /tmp/gdb.out /tmp/io.out
start_session "SELECT * FROM t WHERE id = 2;"
echo "backend=$PID mode=$mode"
docker exec -u root -d -e PROBE_READY=/tmp/gdb-ready pg-lab sh -c "gdb -q -p $PID -batch -x /tmp/probe.py > /tmp/gdb.out 2>&1"
until docker exec pg-lab test -e /tmp/gdb-ready; do sleep 0.1; done

echo "--- P1: gdb가 붙은 백엔드에 strace 붙이기"
docker exec -u root pg-lab timeout 2 strace -p "$PID" -e trace=none 2>&1 | head -3 || true

if [ "$mode" = cold ]; then
	echo "--- 공유 버퍼에서 t 빼기"
	docker exec pg-lab psql -X -At -c "SELECT * FROM pg_buffercache_evict_relation('t')"
fi
echo "--- 추적 직전 공유 버퍼 안의 t 페이지 수"
docker exec pg-lab psql -X -At -c "SELECT count(*) FROM pg_buffercache WHERE relfilenode = 16384"

workers=$(docker exec pg-lab sh -c "pgrep -f 'postgres: io worker' | sed 's/^/-p /' | tr '\n' ' '")
# raw는 읽어 넣는 메모리 주소를 보려고 pread64의 인자를 해석하지 않고 16진수로 남깁니다.
docker exec -u root -d pg-lab sh -c "strace -tt -y -e trace=pread64,preadv,preadv2,openat -e raw=pread64,preadv $workers -o /tmp/io.out"
until docker exec pg-lab test -e /tmp/io.out; do sleep 0.1; done
sleep 0.5
docker exec -u root pg-lab touch /tmp/ready

finish_session
i=0
while docker exec pg-lab pgrep -x gdb >/dev/null && [ "$i" -lt 100 ]; do sleep 0.2; i=$((i + 1)); done
docker exec -u root pg-lab pkill -INT strace || true
sleep 0.3
echo "--- gdb (backend)"
docker exec pg-lab cat /tmp/gdb.out
echo "--- strace (io workers)"
docker exec pg-lab cat /tmp/io.out

# 아래 둘은 추적한 실행이 끝난 직후 다른 연결에서 읽습니다. 같은 순간의 값이 아니라 대조용입니다.
echo "--- pg_buffercache (실행 직후, 쓰인 칸만)"
docker exec pg-lab psql -X -At -F ' ' -c "SELECT b.bufferid, b.reldatabase, b.relfilenode, b.relforknumber, b.relblocknumber, b.usagecount, b.isdirty, coalesce(c.relkind::text, '-'), coalesce(c.relname, '-') FROM pg_buffercache b LEFT JOIN pg_class c ON b.reldatabase = (SELECT oid FROM pg_database WHERE datname = current_database()) AND pg_relation_filenode(c.oid) = b.relfilenode WHERE b.relfilenode IS NOT NULL ORDER BY b.bufferid"
echo "--- heap_page_items (실행 직후)"
docker exec pg-lab psql -X -q -c "CREATE EXTENSION IF NOT EXISTS pageinspect"
docker exec pg-lab psql -X -At -F ' ' -c "SELECT lp, lp_off, lp_len, t_hoff, encode(t_data, 'hex') FROM heap_page_items(get_raw_page('t', 0))"
