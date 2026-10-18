#!/bin/sh
# SQL 한 문장을 실행하는 동안 그 연결의 백엔드 프로세스가 지나는 함수를 gdb로 기록합니다.
#
# 사용: run-trace.sh <감시 목록 파일 이름> "<SQL>" [호출 스택을 남길 함수]
#   감시 목록은 이 scripts 폴더 안의 파일 이름입니다(컨테이너 안에서는 /lab/<이름>).
# 전제: start-lab.sh로 띄운 컨테이너 pg-lab
# 출력: 표준 출력에 trace.py의 STOP/BT/END/NOBREAK 기록과 psql의 결과
# 실패: 60초 안에 세션이 준비되지 않으면 종료 코드 1로 끝납니다.
#
# 같은 SQL을 먼저 한 번 실행해 카탈로그 캐시를 데운 뒤에 붙습니다(session.sh).
# 그래야 기록에 테이블 정의 조회가 섞이지 않고, 이 SQL 자체가 하는 일만 남습니다.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
. "$here/session.sh"

watch="$1"
sql="$2"
bt_at="${3:-}"

start_session "$sql"

# 디버거는 SYS_PTRACE를 가진 root로 실행하고, 중단점을 다 걸면 /tmp/ready를 만듭니다.
docker exec -u root \
	-e TRACE_WATCH="/lab/$watch" -e TRACE_BT_AT="$bt_at" -e TRACE_READY=/tmp/ready \
	pg-lab gdb -q -p "$PID" -batch -x /lab/trace.py 2>&1 |
	grep -E '^(STOP|BT|END|NOBREAK)'

finish_session
