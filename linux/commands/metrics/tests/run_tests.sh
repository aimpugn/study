#!/bin/bash
# run_tests.sh — vm_os_watch.sh 회귀 테스트
#
# 무엇을 검증하는가
#   1) 계산 정확성: 가짜 스냅샷(정답을 손으로 계산해 둔 것)을 AWKPROG 에 넣어
#      수치·색 단계·경보 발행·상태 집계가 기대값과 일치하는지
#   2) 레이아웃 불변식: "들여쓰기2 + 값10 + 공백3 + 라벨", 디스크 표 정렬(표시폭)
#   3) 실행 동작(Linux 에서만): --once 의 제어코드 0건, 로그 포맷/회전/이관, 상주 기동
#
# 사용법:  bash tests/run_tests.sh   (스크립트 수정 후 1분 회귀)
# 종료코드: 실패 1건이라도 있으면 1
#
# 픽스처 정답표 (prev.snap 기준 Δ, dt=1.000s)
#   cur_warn:   us 12.0 / sy 5.5 / wa 25.0 / id 57.5,  load 4.20=vCPU4 의 105%, D=3
#               avail 2.1G(13%), sda: r20 w380, await (162+16074)/400=40.6ms,
#               aqu 16236/1000=16.2, util 984/1000→98%, net 12.30/3.10 Mb/s
#               경보 6건(전부 주의): cpu_wa load proc_d mem_avail vmw_balloon disk_sda_await
#   cur_swapin: si Δ4페이지×4KiB=16K/s, so 0 → swap_in 정보성 주의 1건만
#   cur_swapout:so Δ600×4=2400K/s(≥2048 위험), avail 5%(≤7 위험) → 위험 2건
#   cur_clean:  전부 정상 → 경보 0건, 상태 0/0

set -u
cd "$(dirname "$0")"
SRC=../vm_os_watch.sh

FAILS=0; PASSES=0
ok()   { PASSES=$((PASSES+1)); printf 'PASS  %s\n' "$1"; }
bad()  { FAILS=$((FAILS+1));   printf 'FAIL  %s\n' "$1"; }
check(){ # $1=설명, 이후 = 평가할 명령 (성공=PASS)
    local desc="$1"; shift
    if "$@" >/dev/null 2>&1; then ok "$desc"; else bad "$desc"; fi
}

# ── 0. 문법 ───────────────────────────────────────────────────────────────────
check "문법: vm_os_watch.sh"  bash -n "$SRC"
check "문법: vm_io_health.sh" bash -n ../vm_io_health.sh
check "문법: vm_os_lab.sh"    bash -n ../vm_os_lab.sh

# ── 1. AWKPROG 추출 ──────────────────────────────────────────────────────────
# 스크립트의 작은따옴표 블록(AWKPROG='...')만 떼어 gawk 로 단독 실행한다.
# 추출 가능성 자체가 계약: 여는 줄 "AWKPROG='" / 닫는 줄 "}'" 은 유지돼야 한다.
sed -n "/^AWKPROG='/,/^}'/p" "$SRC" | sed -e "1s/^AWKPROG='//" -e '$ s/.$//' > prog.awk
check "추출: prog.awk 마지막 줄이 '}'" test "$(tail -1 prog.awk)" = "}"

run_case() { # $1=cur 픽스처  $2=BALLOON  $3=HVSWAP   — 색은 [Y]/[R]/[/] 마커로
    gawk -v NCPU=4 -v DISKS="" -v BALLOON="$2" -v HVSWAP="$3" \
         -v W_CPU_WA=20 -v C_CPU_WA=40 -v W_CPU_ST=5 -v C_CPU_ST=15 \
         -v W_LOAD=100 -v C_LOAD=200 -v W_MEMA=15 -v C_MEMA=7 \
         -v C_SWAP=2048 -v W_AWAIT=20 -v C_AWAIT=50 \
         -v W_UTIL=90 -v C_UTIL=99 -v W_AQU=8 -v C_AQU=32 \
         -v W_D=1 -v C_D=5 \
         -v YEL="[Y]" -v RED="[R]" -v DIM="" -v RST="[/]" \
         -f prog.awk prev.snap "$1"
}

# ── 2. 수치·경보 — cur_warn (정답표는 머리 주석) ─────────────────────────────
OUT=$(run_case cur_warn.snap 512 0)
has(){ grep -qF -- "$1" <<<"$OUT"; }
not(){ ! grep -qF -- "$1" <<<"$OUT"; }
check "warn: CPU us 12.0%"            has "12.0%"
check "warn: CPU sy 5.5%"             has "5.5%"
check "warn: CPU wa 25.0% 노랑"       has "[Y]     25.0%[/]"
check "warn: load 105% 노랑"          has "[Y]105%[/]"
check "warn: D state 3 노랑"          has "[Y]         3[/]"
check "warn: avail 2.1G/13% 노랑"     has "[Y]      2.1G[/]"
check "warn: balloon 512M 노랑"       has "[Y]      512M[/]"
check "warn: 디스크 await 40.6ms"     has "40.6ms"
check "warn: 디스크 aqu 16.2"         has "16.2"
check "warn: 디스크 util 98%"         has "98%"
check "warn: 파티션 sda1 미표시"      not "sda1"
check "warn: 루프백 lo 미표시"        not "  lo "
check "warn: 네트워크 12.30 Mb/s"     has "12.30"
for k in cpu_wa load proc_d mem_avail vmw_balloon disk_sda_await; do
    check "warn: 경보 $k 발행"        has "@A@|1|$k"
done
check "warn: 상태 위험0·주의6"        has "@S@|0|6"

# ── 3. 스왑 분기 ─────────────────────────────────────────────────────────────
OUT=$(run_case cur_swapin.snap -1 -1)
check "swapin: si 16K/s 노랑"         has "[Y]     16K/s[/]"
check "swapin: swap_in 정보성 경보"   has "@A@|1|swap_in"
check "swapin: swap_out 경보 없음"    not "swap_out"
check "swapin: tools 미설치 안내"     has "open-vm-tools 미설치"
check "swapin: 상태 위험0·주의1"      has "@S@|0|1"

OUT=$(run_case cur_swapout.snap -1 -1)
check "swapout: so 2400K/s 빨강"      has "[R]   2400K/s[/]"
check "swapout: swap_out 위험 경보"   has "@A@|2|swap_out"
check "swapout: swap_in 경보 없음"    not "swap_in"
check "swapout: avail 5% 위험 경보"   has "@A@|2|mem_avail"
check "swapout: 상태 위험2·주의0"     has "@S@|2|0"

# ── 4. 정상 케이스 ───────────────────────────────────────────────────────────
OUT=$(run_case cur_clean.snap 0 0)
check "clean: 경보 0건"               not "@A@|"
check "clean: 상태 위험0·주의0"       has "@S@|0|0"

# ── 5. 레이아웃 불변식 ───────────────────────────────────────────────────────
# 값 열 규칙: 1-2열 공백, 3-12열 값(우측 정렬), 13-15열 공백, 16열부터 라벨.
# 한글 라벨이 있어도 값이 "앞"이라 이 규칙은 바이트 검사로 충분하다.
# @A@/@S@ 행은 화면 행이 아니므로 검사에서 제외한다 (경보 문구에도 같은 키워드가 있음).
OUT=$(run_case cur_warn.snap 512 0 | grep -v '^@' | sed -e 's/\[[YR]\]//g' -e 's/\[\/\]//g')
BADROWS=$(grep -E "사용자 코드|커널 코드|I/O 완료|빼앗긴|유휴|평균 부하|실행 대기|고착|가용 메모리|페이지 캐시|미기록|스왑 사용량|읽어들임|내려씀|풍선 회수|하이퍼바이저 스왑" <<<"$OUT" \
    | gawk '!(substr($0,1,2)=="  " && substr($0,13,3)=="   " && substr($0,16,1)!=" ")' )
check "레이아웃: 값 열 규칙 (위반 0행)" test -z "$BADROWS"

# 디스크 표: 한글 헤더("장치")가 섞이므로 표시폭(동아시아 폭) 기준으로 검사한다.
if command -v python3 >/dev/null 2>&1; then
    printf '%s\n' "$OUT" > .render.txt
    python3 - <<'PY' && ok "레이아웃: 디스크 표 7컬럼 정렬" || bad "레이아웃: 디스크 표 7컬럼 정렬"
import sys, unicodedata
def disp(s): return sum(2 if unicodedata.east_asian_width(c) in ('W','F') else 1 for c in s)
lines = open('.render.txt', encoding='utf-8').read().splitlines()
hdr = next(l for l in lines if '장치' in l)
dat = next(l for l in lines if l.lstrip().startswith('sda'))
def edge(line, tok):
    i = line.index(tok); return disp(line[:i+len(tok)])
pairs = [('r/s','20'),('w/s','380'),('rMB/s','0.3'),('wMB/s','24.2'),('await','ms'),('aqu','16.2'),('util','%')]
sys.exit(0 if all(edge(hdr,h)==edge(dat,d) for h,d in pairs) else 1)
PY
    rm -f .render.txt
else
    printf 'SKIP  레이아웃: 디스크 표 (python3 없음)\n'
fi

# ── 6. 실행 동작 (Linux /proc 필요 — 없으면 건너뜀) ──────────────────────────
if [ -r /proc/stat ] && [ -r /proc/diskstats ]; then
    TD=$(mktemp -d); trap 'rm -rf "$TD"' EXIT

    # 6a. --once: 위치 제어코드(ESC) 0건의 일반 텍스트여야 한다 (캡처 용도 계약)
    OUT=$(LOGF="$TD/once.log" LEGACY_LOG="$TD/none" bash "$SRC" --once 2>/dev/null)
    check "once: ESC 코드 0건"          test "$(printf '%s' "$OUT" | tr -dc '\033' | wc -c)" -eq 0
    check "once: 프레임 출력"           grep -q "VM OS WATCH" <<<"$OUT"
    check "once: 상태 줄 존재"          grep -q "^상태" <<<"$OUT"
    # dt 회귀 가드: date 의 %3N 미지원으로 dt 가 백만 배 부풀던 버그의 재발 방지.
    # 1초 간격이면 헤더가 반드시 "1초 차분"이어야 한다.
    check "once: dt 정상 (1초 차분)"    grep -q "1초 차분" <<<"$OUT"

    # 6b. 도움말·오류 인자·옵션 우선순위
    check "help: 사용법 출력"           bash -c "bash '$SRC' -h | grep -q 사용법"
    check "args: 잘못된 인자 거부(exit2)" bash -c "bash '$SRC' --bogus >/dev/null 2>&1; test \$? -eq 2"
    # --log 인자가 환경변수 LOGF 보다 우선한다는 계약 (헤더 문서에 명시된 순서)
    LOGF="$TD/env.log" LEGACY_LOG="$TD/none" W_MEMA=99 C_MEMA=0 bash "$SRC" --once --log "$TD/arg.log" >/dev/null 2>&1
    check "args: --log 이 env LOGF 보다 우선" bash -c "test -s '$TD/arg.log' && test ! -e '$TD/env.log'"

    # 6c. 로그 포맷: "YYYY-MM-DD HH:MM:SS [레벨] ..." — 경보를 강제로 1건 발생시켜 확인
    LOGF="$TD/fmt.log" LEGACY_LOG="$TD/none" W_MEMA=99 C_MEMA=0 bash "$SRC" --once >/dev/null 2>&1
    check "log: 연도 포함 포맷"         grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} \[(INFO|WARN|CRIT)\] ' "$TD/fmt.log"

    # 6d. 회전: 1MiB 초과 상태에서 기록이 발생하면 .1 보관 + 회전 INFO 가 남아야 한다
    dd if=/dev/zero of="$TD/rot.log" bs=1024 count=1100 2>/dev/null
    LOGF="$TD/rot.log" LEGACY_LOG="$TD/none" ROTATE_MB=1 W_MEMA=99 C_MEMA=0 bash "$SRC" --once >/dev/null 2>&1
    check "rotate: .1 보관"             test -s "$TD/rot.log.1"
    check "rotate: 회전 INFO 기록"      grep -q "로그 회전" "$TD/rot.log"

    # 6e. 이관: 과거 위치에 이력이 있고 새 파일이 없으면 1회 이어붙인다 (원본 보존)
    printf '2026-01-01 00:00:00 [WARN] 과거 경보 표본\n' > "$TD/legacy.log"
    LOGF="$TD/new/alerts.log" LEGACY_LOG="$TD/legacy.log" bash "$SRC" --once >/dev/null 2>&1
    check "migrate: 이어붙임 마커"      grep -q "이어붙임" "$TD/new/alerts.log"
    check "migrate: 과거 내용 보존"     grep -q "과거 경보 표본" "$TD/new/alerts.log"
    check "migrate: 원본 미삭제"        test -s "$TD/legacy.log"

    # 6f. 상주 기동: 4초 돌려 프레임 갱신과 세션 시작 INFO 를 확인
    LOGF="$TD/run.log" LEGACY_LOG="$TD/none" timeout 4 bash "$SRC" 1 > "$TD/run.out" 2>&1 || true
    check "resident: 프레임에 상태 줄"  grep -q "상태" "$TD/run.out"
    check "resident: 시작 INFO 기록"    grep -q "모니터 시작" "$TD/run.log"
else
    printf 'SKIP  실행 동작 테스트 (/proc 없음 — Linux 에서 실행)\n'
fi

# ── 결과 ─────────────────────────────────────────────────────────────────────
printf -- '────────────────────────────────\n'
printf 'PASS %d / FAIL %d\n' "$PASSES" "$FAILS"
rm -f prog.awk
exit $(( FAILS > 0 ? 1 : 0 ))
