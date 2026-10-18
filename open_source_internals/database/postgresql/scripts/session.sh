#!/bin/sh
# run-trace.sh와 run-strace.sh가 함께 쓰는 세션 준비 함수입니다. 직접 실행하지 않고 source로 읽습니다.
#
# start_session "<SQL>"
#   pg-lab 안에서 psql 연결 하나를 열고, 같은 SQL로 캐시를 데운 뒤 /tmp/ready가 생길 때까지 기다리게 합니다.
#   반환: 전역 변수 PID에 그 연결을 처리하는 백엔드 프로세스 번호를 넣습니다.
#   실패: 60초 안에 캐시 데우기가 끝나지 않으면 종료 코드 1로 끝납니다.
# finish_session
#   psql이 끝날 때까지 최대 30초 기다린 뒤 psql의 출력을 보여 줍니다.
#
# 추적기는 PID에 붙은 다음 /tmp/ready를 만들어야 합니다. 그래야 캐시 데우기 실행이 기록에 섞이지 않습니다.
# 호스트의 PG_OPTIONS는 psql의 PGOPTIONS로 넘어가 연결 시작 때 설정을 바꿉니다(예: -c synchronous_commit=off).
# SQL 안에 SET을 함께 쓰면 psql이 문장마다 따로 보내므로, 추적은 첫 문장에서 끝나 버립니다.

start_session() {
	sql="$1"

	# 지난 실행의 표시 파일을 지웁니다. /tmp/ready는 root로 실행한 추적기가 만들므로 root로 지워야 합니다.
	docker exec -u root pg-lab rm -f /tmp/pid /tmp/warm /tmp/ready /tmp/psql.out

	# 한 연결 안에서 PID를 남기고, 캐시를 데우고, 추적기가 준비될 때까지 기다렸다가 SQL을 보냅니다.
	printf '%s\n' \
		'\o /tmp/pid' 'SELECT pg_backend_pid();' '\o' \
		"$sql" '\! touch /tmp/warm' \
		'\! while [ ! -e /tmp/ready ]; do sleep 0.2; done' \
		"$sql" |
		docker exec -i pg-lab sh -c 'cat > /tmp/session.sql'
	docker exec -d -e PGOPTIONS="${PG_OPTIONS:-}" pg-lab \
		sh -c 'psql -X -At -f /tmp/session.sql > /tmp/psql.out 2>&1'

	# 캐시 데우기가 끝난 뒤에만 붙어야 그 실행이 기록에 섞이지 않습니다.
	i=0
	until docker exec pg-lab test -e /tmp/warm; do
		i=$((i + 1))
		[ "$i" -ge 300 ] && { echo "세션이 준비되지 않았습니다" >&2; exit 1; }
		sleep 0.2
	done
	PID=$(docker exec pg-lab head -1 /tmp/pid)
}

finish_session() {
	i=0
	until ! docker exec pg-lab pgrep -f 'psql -X -At -f /tmp/session.sql' >/dev/null; do
		i=$((i + 1))
		[ "$i" -ge 150 ] && break
		sleep 0.2
	done
	echo "--- psql"
	docker exec pg-lab cat /tmp/psql.out
}
