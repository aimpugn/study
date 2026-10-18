#!/bin/sh
# 실습용 PostgreSQL 컨테이너 pg-lab을 띄우고 작은 테이블 t를 만듭니다.
#
# 사용: start-lab.sh            (저장소 어디에서 실행해도 됩니다)
# 전제: docker build로 만든 pg-internals:18.6 이미지
# 결과: 실행 중인 컨테이너 pg-lab, 행 3개짜리 테이블 t(id int, name text)
# 실패: 같은 이름의 컨테이너가 이미 있거나 서버가 30초 안에 준비되지 않으면 0이 아닌 코드로 끝납니다.
#
# 안전 경계: 디버거에 필요한 SYS_PTRACE 권한만 더하고, 네트워크를 끊고, 포트를 열지 않습니다.
# 호스트에서는 이 scripts 폴더만 읽기 전용으로 /lab에 연결합니다.
set -eu

here=$(cd "$(dirname "$0")" && pwd)

docker run -d --name pg-lab \
	--cap-add=SYS_PTRACE \
	--network none \
	-v "$here:/lab:ro" \
	pg-internals:18.6 >/dev/null

# 서버가 Unix socket에서 연결을 받을 때까지 기다립니다.
i=0
until docker exec pg-lab pg_isready -q; do
	i=$((i + 1))
	[ "$i" -ge 30 ] && { echo "서버가 준비되지 않았습니다" >&2; exit 1; }
	sleep 1
done

# 추적 결과를 손으로 따라갈 수 있도록 한 페이지에 들어가는 작은 테이블을 씁니다.
docker exec pg-lab psql -X -q -c \
	"CREATE TABLE t(id int, name text); INSERT INTO t VALUES (1,'a'),(2,'b'),(3,'c');"
echo "pg-lab 준비 완료"
