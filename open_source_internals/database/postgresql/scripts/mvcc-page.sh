#!/bin/sh
# 한 행을 INSERT·UPDATE·VACUUM하면서 페이지 안의 행 버전(xmin/xmax)과 줄 포인터가 어떻게 바뀌는지 보여 줍니다.
#
# 사용: mvcc-page.sh
#   ISOLATION 환경 변수로 3단계 읽기 트랜잭션의 격리 수준을 바꿉니다(기본 REPEATABLE READ).
# 전제: start-lab.sh로 띄운 컨테이너 pg-lab(pageinspect 확장이 설치된 이미지)
# 출력: 단계마다 트랜잭션 번호와 테이블 u의 0번 페이지 내용(heap_page_items)
# 실패: SQL이 실패하면 그 자리에서 0이 아닌 종료 코드로 끝납니다.
#       ISOLATION 값이 격리 수준 이름이 아니면 종료 코드 2, 읽기 트랜잭션이 30초 안에 준비되지 않으면 종료 코드 1로 끝납니다.
#
# 3단계에서는 읽기 트랜잭션을 먼저 열어 둔 채 다른 세션이 같은 행을 갱신합니다.
# 먼저 시작한 읽기가 옛 버전을 계속 보는지 확인해야, 옛 버전을 페이지에 남기는 이유가 보입니다.
set -eu

isolation="${ISOLATION:-REPEATABLE READ}"

# 값에 오타가 있으면 BEGIN이 실패하고 두 읽기가 각자 자동 커밋으로 돌아, READ COMMITTED와 같은 결과가 조용히 나옵니다.
# 그래서 실행 전에 이름을 확인합니다.
case "$isolation" in
"READ COMMITTED" | "REPEATABLE READ" | "SERIALIZABLE") ;;
*)
	echo "ISOLATION은 READ COMMITTED, REPEATABLE READ, SERIALIZABLE 가운데 하나여야 합니다: $isolation" >&2
	exit 2
	;;
esac

# wait_for <컨테이너 안에서 검사할 셸 조건> <실패 메시지>
#   조건이 참이 될 때까지 0.1초마다 최대 30초 기다립니다.
#   넘으면 메시지와 읽기 세션의 출력을 보이고 종료 코드 1로 끝납니다.
# 읽기 세션이 SQL 오류로 멈추면 표시 파일이 생기지 않으므로, 기다리는 시간에 한도를 둡니다.
wait_for() {
	i=0
	until docker exec pg-lab sh -c "$1"; do
		i=$((i + 1))
		if [ "$i" -ge 300 ]; then
			echo "$2" >&2
			docker exec pg-lab cat /tmp/reader.out >&2 || true
			exit 1
		fi
		sleep 0.1
	done
}

sql() { docker exec pg-lab psql -X -q -v ON_ERROR_STOP=1 -P pager=off -c "$1"; }

page() {
	echo "-- 테이블 u, 페이지 0"
	# lp_off는 보통 행 데이터의 위치지만, REDIRECT 상태에서는 이어지는 줄 번호입니다.
	sql "SELECT lp, lp_flags, lp_off, t_xmin, t_xmax, t_ctid,
	            (heap_tuple_infomask_flags(t_infomask, t_infomask2)).raw_flags AS flags
	     FROM heap_page_items(get_raw_page('u', 0))"
}

# 0. 빈 테이블 하나로 시작합니다.
docker exec -u root pg-lab rm -f /tmp/r-ready /tmp/w-done /tmp/reader.out
sql "CREATE EXTENSION IF NOT EXISTS pageinspect; DROP TABLE IF EXISTS u; CREATE TABLE u(id int, v text);"

# 1. INSERT: 새 버전 하나가 생기고 xmin에 이 트랜잭션 번호가 들어갑니다.
echo "## 1. INSERT"
sql "BEGIN; SELECT pg_current_xact_id() AS insert_xid; INSERT INTO u VALUES (1, 'a'); COMMIT;"
page

# 2. UPDATE: 행을 제자리에서 고치지 않고 새 버전을 추가하는지 봅니다.
echo "## 2. UPDATE v='b'"
sql "BEGIN; SELECT pg_current_xact_id() AS update_xid; UPDATE u SET v = 'b' WHERE id = 1; COMMIT;"
page

# 3. 읽기 트랜잭션을 먼저 열고, 그동안 다른 세션이 갱신합니다.
#    REPEATABLE READ는 첫 쿼리 때 스냅숏을 고정하고, READ COMMITTED는 문장마다 새로 잡습니다.
echo "## 3. 먼저 시작한 읽기 트랜잭션($isolation)이 있는 동안 UPDATE v='c'"
printf '%s\n' \
	"BEGIN ISOLATION LEVEL $isolation;" \
	"SELECT 'reader 1st read' AS step, v, pg_current_snapshot() AS snapshot FROM u;" \
	'\! touch /tmp/r-ready' \
	'\! while [ ! -e /tmp/w-done ]; do sleep 0.1; done' \
	"SELECT 'reader 2nd read' AS step, v FROM u;" \
	'COMMIT;' |
	docker exec -i pg-lab sh -c 'cat > /tmp/reader.sql'
docker exec -d pg-lab sh -c 'psql -X -q -v ON_ERROR_STOP=1 -P pager=off -f /tmp/reader.sql > /tmp/reader.out 2>&1'
wait_for 'test -e /tmp/r-ready' "읽기 트랜잭션이 준비되지 않았습니다"
sql "BEGIN; SELECT pg_current_xact_id() AS update_xid; UPDATE u SET v = 'c' WHERE id = 1; COMMIT;"
sql "SELECT 'new session read' AS step, v FROM u;"
page
docker exec pg-lab touch /tmp/w-done
# 조건을 sh -c로 실행하므로 그 sh의 명령줄에도 같은 문자열이 들어갑니다. [p]sql로 써야 pgrep이 sh 자신과 맞지 않습니다.
wait_for '! pgrep -f "[p]sql -X -q -v ON_ERROR_STOP=1 -P pager=off -f /tmp/reader.sql" >/dev/null' "읽기 트랜잭션이 끝나지 않았습니다"
echo "-- 먼저 시작한 읽기 트랜잭션의 출력"
docker exec pg-lab cat /tmp/reader.out

# 4. VACUUM: 어떤 스냅숏에도 더는 보이지 않는 옛 버전을 정리합니다.
echo "## 4. VACUUM"
sql "VACUUM u;"
page
