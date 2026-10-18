#!/bin/bash
# 탐색용 코드입니다(2026-10-13, 작업대 관측 재설계의 P15~P20 시험). 작업대 도구를 Go로 통일하므로,
# 이 순서를 Go 하위 명령으로 다시 만든 뒤 이 파일을 지웁니다.
#
# 하는 일: 기록 빌드(-finstrument-functions + pgtrace_hook.c)로 만든 실행 파일로 두 번째 서버를 띄우고,
#          strace로 모든 프로세스의 파일·소켓 시스템 콜을 함께 남기며 세션 다섯 개(S1~S5)를 실행합니다.
# 전제: pg-lab 컨테이너 안, postgres 사용자, /src/postgresql-18.6/src/backend/postgres가 기록 빌드
#       (빌드 순서는 WORK_20261013_WORKBENCH_OBSERVATION_DESIGN.md의 「결과: 모든 함수 호출 기록」)
# 실행: docker cp run.sh pg-lab:/tmp/ && docker exec pg-lab bash /tmp/run.sh
# 결과: /tmp/pgtrace/<pid>.bin(프로세스별 기록), /tmp/pgi/all.strace, /tmp/pgi/ps.txt, 표준 출력에 psql 결과
# 실패 신호: 서버가 60초 안에 준비되지 않거나, 끝의 server.log 검사에 PANIC·FATAL·signal 줄이 나옵니다.
# pg-lab의 원래 서버(/usr/local/pgsql, 소켓 /tmp)와 데이터는 건드리지 않습니다. /tmp/pgi와 /tmp/pgtrace는 지우고 새로 만듭니다.
set -u
B=/src/postgresql-18.6/src/backend/postgres
R=/tmp/pgi
P="psql -X -h $R -p 5433 -d postgres"
rm -rf $R /tmp/pgtrace && mkdir -p $R/bin /tmp/pgtrace
cp $B $R/bin/postgres
ln -s /usr/local/pgsql/share $R/share
ln -s /usr/local/pgsql/lib $R/lib
initdb -D $R/data --no-locale -E UTF8 >/dev/null || exit 1

# 준비: 기록 없이 띄워 테이블을 만들고 내린다.
$R/bin/postgres -D $R/data -c listen_addresses= -c unix_socket_directories=$R -p 5433 >$R/server-setup.log 2>&1 &
for i in $(seq 30); do pg_isready -q -h $R -p 5433 && break; sleep 1; done
$P -q -c "CREATE TABLE t(id int, name text); INSERT INTO t VALUES (1,'a'),(2,'b'),(3,'c');"
pg_ctl -D $R/data stop -m fast >/dev/null

# 본 실행: 기록을 켜고, 모든 프로세스의 파일·네트워크 시스템 콜을 strace로 함께 남긴다.
echo "start $(date +%s%N)"
PGTRACE_DIR=/tmp/pgtrace strace -f --seccomp-bpf -ttt -e trace=openat,pread64,preadv,pwrite64,pwritev,fdatasync,fsync,sendto,recvfrom -o $R/all.strace \
  $R/bin/postgres -D $R/data -c listen_addresses= -c unix_socket_directories=$R -p 5433 >$R/server.log 2>&1 &
for i in $(seq 60); do pg_isready -q -h $R -p 5433 && break; sleep 1; done
echo "ready $(date +%s%N)"
ps -u postgres -o pid,ppid,cmd | grep -v grep | grep -E "postgres|strace" > $R/ps.txt

echo "== S1 SELECT"; $P -c '\timing on' -c 'SELECT pg_backend_pid()' -c 'SELECT * FROM t' -c 'SELECT * FROM t'
echo "== S2 CREATE TABLE"; $P -c '\timing on' -c 'SELECT pg_backend_pid()' -c 'CREATE TABLE u(id int, name text)'
echo "== S3 CREATE INDEX"; $P -c '\timing on' -c 'SELECT pg_backend_pid()' -c 'CREATE INDEX ON t(id)'
echo "== S4 INSERT"; $P -c '\timing on' -c 'SELECT pg_backend_pid()' -c "INSERT INTO t VALUES (4,'d')"
echo "== S5 ERROR"; $P -c '\timing on' -c 'SELECT pg_backend_pid()' -c 'SELECT 1/0' -c 'SELECT 1'
sleep 1
pg_ctl -D $R/data stop -m fast >/dev/null
sleep 1
echo "stop $(date +%s%N)"
grep -c . $R/all.strace
grep -E "PANIC|FATAL|terminated by signal|server process" $R/server.log | head
ls /tmp/pgtrace | wc -l; du -sh /tmp/pgtrace
