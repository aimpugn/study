#!/bin/bash
# 탐색용 코드입니다(2026-10-13, 요청의 여정 시험 P21~P25). 작업대 도구를 Go로 통일하므로,
# 이 순서를 Go 하위 명령으로 다시 만든 뒤 이 파일을 지웁니다.
#
# 하는 일: 기록 빌드로 두 번째 서버를 새로 띄우고 켜 둡니다. 모든 프로세스의 함수 기록은 /tmp/pgtrace/<pid>.bin에,
#          파일·소켓 시스템 콜은 /tmp/pgi/all.strace에 쌓입니다(-yy로 디스크립터마다 경로를, -T로 커널 안에서 걸린 시간을 붙임).
#          모든 프로세스를 계속 기록하므로 70분에 약 2.8 GB가 쌓였습니다. 시험이 끝나면 끄고 /tmp/pgtrace를 지웁니다.
# 전제: run.sh와 같음(pg-lab 컨테이너 안, postgres 사용자, /src의 기록 빌드).
# 실행: docker cp serve.sh pg-lab:/tmp/ && docker exec pg-lab bash /tmp/serve.sh
# 쿼리: docker exec -it pg-lab psql -h /tmp/pgi -p 5433 -d postgres   (실제 클라이언트로 아무 쿼리나)
# 끄기: docker exec pg-lab pg_ctl -D /tmp/pgi/data stop -m fast
# 실패 신호: 서버가 60초 안에 준비되지 않으면 0이 아닌 코드로 끝납니다.
set -u
B=/src/postgresql-18.6/src/backend/postgres
R=/tmp/pgi
P="psql -X -h $R -p 5433 -d postgres"
pg_ctl -D $R/data stop -m fast >/dev/null 2>&1
rm -rf $R /tmp/pgtrace && mkdir -p $R/bin /tmp/pgtrace
cp $B $R/bin/postgres
ln -s /usr/local/pgsql/share $R/share
ln -s /usr/local/pgsql/lib $R/lib
initdb -D $R/data --no-locale -E UTF8 >/dev/null || exit 1

# 준비: 기록 없이 띄워 테이블을 만들고 내린다. 본 실행은 공유 버퍼가 빈 상태에서 시작한다.
$R/bin/postgres -D $R/data -c listen_addresses= -c unix_socket_directories=$R -p 5433 >$R/server-setup.log 2>&1 &
for i in $(seq 30); do pg_isready -q -h $R -p 5433 && break; sleep 1; done
$P -q -c "CREATE TABLE t(id int, name text); INSERT INTO t VALUES (1,'a'),(2,'b'),(3,'c');"
pg_ctl -D $R/data stop -m fast >/dev/null

PGTRACE_DIR=/tmp/pgtrace setsid nohup strace -f -yy -s 1024 --seccomp-bpf -ttt -T \
  -e trace=openat,pread64,preadv,pwrite64,pwritev,fdatasync,fsync,sendto,recvfrom -o $R/all.strace \
  $R/bin/postgres -D $R/data -c listen_addresses= -c unix_socket_directories=$R -p 5433 >$R/server.log 2>&1 </dev/null &
for i in $(seq 60); do pg_isready -q -h $R -p 5433 && { echo "ready $(date +%s%N)"; exit 0; }; sleep 1; done
echo "서버가 준비되지 않았습니다" >&2; exit 1
