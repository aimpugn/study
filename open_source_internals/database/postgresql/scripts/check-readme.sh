#!/bin/sh
# README.md 본문이 인용한 C 이름이 빌드한 18.6 소스에 실제로 정의되어 있는지 확인합니다.
#
# 사용: check-readme.sh
# 전제: start-lab.sh로 띄운 컨테이너 pg-lab
# 출력: 정의를 찾지 못한 이름. 모두 찾으면 "README 인용 확인: 이름 N개, 모두 정의 있음"
# 실패: 찾지 못한 이름이 하나라도 있으면 종료 코드 1로 끝납니다.
#
# README는 함수를 `이름()`, 구조체를 `CamelCase`, 매크로와 열거 상수를 `대문자_이름`으로 씁니다.
# 이 규칙을 지킨 이름만 검사하므로, 줄 번호 인용과 설정 이름은 이 스크립트가 확인하지 않습니다.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
readme="$here/../README.md"
src=/src/postgresql-18.6/src

# PostgreSQL 소스 밖의 이름입니다. Docker 권한, 실습 스크립트의 환경 변수, InnoDB의 숨은 열입니다.
outside="SYS_PTRACE PG_OPTIONS DB_TRX_ID DB_ROLL_PTR"

# 코드 블록 밖의 본문에서 규칙에 맞는 이름을 뽑습니다.
prose=$(awk '/^```/ { inblock = !inblock; next } !inblock' "$readme")
funcs=$(printf '%s\n' "$prose" | grep -oE '`[A-Za-z_][A-Za-z0-9_]*\(\)`' | tr -d '`()' | sort -u)
types=$(printf '%s\n' "$prose" | grep -oE '`[A-Z][a-z]+[A-Z][A-Za-z]*`' | tr -d '`' | sort -u)
macros=$(printf '%s\n' "$prose" | grep -oE '`[A-Z][A-Z0-9]*_[A-Z0-9_]*`' | tr -d '`' | sort -u)

missing=""
count=0

# 함수와 구조체는 find-def.sh로 정의 줄을 찾습니다.
for name in $(docker exec pg-lab sh /lab/find-def.sh $funcs $types | awk -F'\t' '$2 == "-" { print $1 }'); do
	missing="$missing $name"
done
count=$(printf '%s\n' $funcs $types | grep -c . || true)

# 매크로는 #define, 열거 상수, 또는 출력용 문자열("이름")로 소스에 있어야 합니다.
for name in $macros; do
	count=$((count + 1))
	case " $outside " in *" $name "*) continue ;; esac
	docker exec pg-lab grep -rqE "^#define[[:space:]]+$name\b|^[[:space:]]+$name[[:space:]]*(=|,|$)|\"$name\"" \
		"$src/include" "$src/backend" || missing="$missing $name"
done

if [ -n "$missing" ]; then
	echo "정의를 찾지 못한 이름:$missing"
	exit 1
fi
echo "README 인용 확인: 이름 ${count}개, 모두 정의 있음"
