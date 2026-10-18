#!/bin/sh
# 빌드한 PostgreSQL 소스 트리(src와 contrib)에서 함수·구조체가 정의된 파일과 줄을 찾습니다.
#
# 사용: find-def.sh 이름...
#   소스 트리 위치는 PG_SRC 환경 변수로 바꿀 수 있습니다(기본 /src/postgresql-18.6).
# 출력: 정의마다 "이름<TAB>상대경로:줄" 한 줄. 정의가 없으면 "이름<TAB>-".
# 실패: 소스 트리가 없으면 종료 코드 2로 끝납니다.
#
# 학습 문서에 인용한 이름이 실제로 빌드한 소스에 있는지 확인하는 데 씁니다.
# 호출 위치나 extern 선언은 정의가 아니므로 세지 않습니다.
# contrib도 찾는 것은 실습이 쓰는 pageinspect 확장의 SQL 함수(heap_page_items 등)가 거기 있기 때문입니다.
set -eu

src="${PG_SRC:-/src/postgresql-18.6}"
if [ ! -d "$src/src" ]; then
	echo "소스 트리가 없습니다: $src" >&2
	exit 2
fi
cd "$src"

for name in "$@"; do
	# PostgreSQL은 pgindent 규칙에 따라 함수 정의의 이름을 줄 맨 앞에 둡니다.
	# 그래서 "이름(" 으로 시작하는 줄이 함수 정의이고, 구조체는 "struct 이름" 줄이 정의입니다.
	hits=$(grep -rnE --include='*.c' --include='*.h' \
		"^${name}\(|^(typedef )?struct ${name}\$" src contrib || true)

	# 찾지 못한 이름도 결과에 남겨야 예측과 실제의 차이를 셀 수 있습니다.
	if [ -z "$hits" ]; then
		printf '%s\t-\n' "$name"
	else
		echo "$hits" | cut -d: -f1,2 | sed "s|^|${name}	|"
	fi
done
