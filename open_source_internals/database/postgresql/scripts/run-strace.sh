#!/bin/sh
# SQL 한 문장을 실행하는 동안 백엔드와 WAL writer가 부르는 시스템 콜을 strace로 기록합니다.
#
# 사용: run-strace.sh <시스템 콜 목록> "<SQL>" [SQL 뒤에 더 기록할 초]
#   시스템 콜 목록은 strace -e trace= 형식입니다(예: pwrite64,fdatasync).
#   SQL이 끝난 뒤에도 WAL writer가 늦게 하는 일을 보려고 기본 1초를 더 기록합니다.
# 전제: start-lab.sh로 띄운 컨테이너 pg-lab
# 출력: "역할 시각 시스템 콜" 줄. 역할은 backend 또는 walwriter입니다. 이어서 psql의 결과.
# 실패: 세션이나 WAL writer를 찾지 못하면 0이 아닌 종료 코드로 끝납니다.
#
# 백엔드만 보면 synchronous_commit=off에서 flush가 "사라진" 것처럼 보입니다.
# 실제로는 WAL writer가 넘겨받으므로 두 프로세스를 함께 기록합니다.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
. "$here/session.sh"

syscalls="$1"
sql="$2"
linger="${3:-1}"

start_session "$sql"
walwriter=$(docker exec pg-lab sh -c "ps -eo pid,args | awk '/postgres: walwriter/ && !/awk/ {print \$1}'")
[ -n "$walwriter" ] || { echo "WAL writer를 찾지 못했습니다" >&2; exit 1; }

# -y는 파일 디스크립터 옆에 실제 파일 경로를, -tt는 마이크로초 단위 시각을 붙입니다.
docker exec -u root pg-lab rm -f /tmp/strace.out
docker exec -u root -d pg-lab \
	strace -tt -y -e trace="$syscalls" -p "$PID" -p "$walwriter" -o /tmp/strace.out
until docker exec pg-lab test -e /tmp/strace.out; do sleep 0.1; done
sleep 0.5
docker exec -u root pg-lab touch /tmp/ready

finish_session

# WAL writer가 비동기 커밋을 flush할 시간을 준 뒤 기록을 멈춥니다.
sleep "$linger"
docker exec -u root pg-lab pkill -INT strace || true
sleep 0.3
echo "--- strace (backend=$PID walwriter=$walwriter)"
docker exec pg-lab cat /tmp/strace.out |
	sed -e "s/^$PID /backend /" -e "s/^$walwriter /walwriter /"
