#!/bin/bash
# memory_metrics_verify.sh
#
# memory_metrics.md 의 주장을 실행으로 확인한다. 읽기 전용이며 시스템 상태를 바꾸지 않는다.
#
#   check-free      free(1) 각 컬럼이 /proc/meminfo 의 어느 필드에서 나오는지 대조
#   check-avail     커널 si_mem_available() 을 사용자 공간에서 재현해 MemAvailable 과 비교
#   check-commit    CommitLimit = SwapTotal + MemTotal * overcommit_ratio / 100 검산
#   compare-usage   같은 스냅샷에 네 가지 "사용률" 계산식을 대입해 차이를 보여 준다
#   watch-usage     N 초 간격으로 사용률 계산식별 변동폭을 관찰 (기본 12회 x 5초)
#   snapshot        진단용 원시 스냅샷 덤프
#   all             check-* 전부
#
# 사용:  ./memory_metrics_verify.sh [subcommand] [args]
#
# PASS/FAIL 기준은 memory_metrics.md 7장에 있다.

set -u

MI=$(mktemp) ; ZI=$(mktemp) ; FR=$(mktemp)
trap 'rm -f "$MI" "$ZI" "$FR"' EXIT

snap() {
    # 세 소스를 최대한 같은 시점에 뜬다. 그래도 수 ms 의 드리프트는 남는다.
    free -k | sed -n 2p > "$FR"
    cp /proc/meminfo  "$MI"
    cp /proc/zoneinfo "$ZI" 2>/dev/null || : > "$ZI"
}

g() { awk -v k="$1" -F'[: ]+' '$1 == k { print $2; exit }' "$MI"; }

pct() { awk -v a="$1" -v b="$2" 'BEGIN { if (b == 0) print "n/a"; else printf "%.1f", a * 100 / b }'; }

hr() { printf '%s\n' "------------------------------------------------------------"; }

# ---------------------------------------------------------------- check-free
check_free() {
    snap
    read -r _ TOT USED FREE SHARED BC AVAIL < "$FR"
    local MT MF MA BU CA SR SH
    MT=$(g MemTotal); MF=$(g MemFree);  MA=$(g MemAvailable)
    BU=$(g Buffers);  CA=$(g Cached);   SR=$(g SReclaimable); SH=$(g Shmem)

    echo "[check-free] free(1) 컬럼 <- /proc/meminfo  (단위 kB)"
    echo "  free --version: $(free --version 2>&1)"
    hr
    printf '%-12s %14s %14s %10s  %s\n' "컬럼" "free 출력" "meminfo 유도" "차이" "유도식"
    printf '%-12s %14s %14s %10s  %s\n' "total"      "$TOT"    "$MT"               "$((TOT - MT))"              "MemTotal"
    printf '%-12s %14s %14s %10s  %s\n' "free"       "$FREE"   "$MF"               "$((FREE - MF))"             "MemFree"
    printf '%-12s %14s %14s %10s  %s\n' "available"  "$AVAIL"  "$MA"               "$((AVAIL - MA))"            "MemAvailable"
    printf '%-12s %14s %14s %10s  %s\n' "shared"     "$SHARED" "$SH"               "$((SHARED - SH))"           "Shmem"
    printf '%-12s %14s %14s %10s  %s\n' "buff/cache" "$BC"     "$((BU + CA + SR))" "$((BC - BU - CA - SR))"     "Buffers + Cached + SReclaimable"
    hr
    echo "  used 후보 A  total - available                        = $((MT - MA))"
    echo "               free.used 와 차이                        = $((MT - MA - USED))"
    echo "  used 후보 B  total - free - buffers - cached - srecl  = $((MT - MF - BU - CA - SR))"
    echo "               free.used 와 차이                        = $((MT - MF - BU - CA - SR - USED))"
    echo
    echo "  판정: 차이가 0 에 가까운 쪽이 이 procps-ng 버전의 used 정의다."
    echo "        스냅샷 간 드리프트로 수 MB 오차는 정상. 수백 MB 차이는 다른 식이라는 뜻."
}

# --------------------------------------------------------------- check-avail
check_avail() {
    snap
    if [ ! -s "$ZI" ]; then
        echo "[check-avail] SKIP — /proc/zoneinfo 를 읽을 수 없다 (일부 컨테이너/가상 환경)"
        return 0
    fi

    local PAGE_KB MF MA AF IF KR
    PAGE_KB=$(( $(getconf PAGE_SIZE) / 1024 ))
    MF=$(g MemFree); MA=$(g MemAvailable)
    AF=$(g 'Active(file)'); IF=$(g 'Inactive(file)'); KR=$(g KReclaimable)

    # zone 별 low 워터마크 합계와 totalreserve( max(lowmem_reserve[]) + high 워터마크 ) 합계
    local WLOW_PAGES TRESERVE_PAGES
    read -r WLOW_PAGES TRESERVE_PAGES < <(awk '
        /^Node .*zone/                { inzone = 1; maxprot = 0; low = 0; high = 0 }
        inzone && $1 == "low"         { low = $2 }
        inzone && $1 == "high"        { if (high == 0) high = $2 }   # zone 워터마크는 첫 high 만
        inzone && $1 == "protection:" {
            for (i = 2; i <= NF; i++) { gsub(/[(),]/, "", $i); if ($i + 0 > maxprot) maxprot = $i + 0 }
            wlow += low; treserve += maxprot + high
            inzone = 0
        }
        END { print wlow + 0, treserve + 0 }
    ' "$ZI")

    local WLOW TRESERVE PAGECACHE AVAIL cut1 cut2
    WLOW=$(( WLOW_PAGES * PAGE_KB ))
    TRESERVE=$(( TRESERVE_PAGES * PAGE_KB ))
    PAGECACHE=$(( AF + IF ))

    cut1=$(( PAGECACHE / 2 )); [ "$cut1" -gt "$WLOW" ] && cut1=$WLOW
    cut2=$(( KR / 2 ));        [ "$cut2" -gt "$WLOW" ] && cut2=$WLOW

    AVAIL=$(( (MF - TRESERVE) + (PAGECACHE - cut1) + (KR - cut2) ))
    [ "$AVAIL" -lt 0 ] && AVAIL=0

    echo "[check-avail] si_mem_available() 재현  (단위 kB)"
    hr
    echo "  PAGE_SIZE          = $(getconf PAGE_SIZE) B"
    echo "  wmark_low 합계     = ${WLOW_PAGES} pages = ${WLOW}"
    echo "  totalreserve 합계  = ${TRESERVE_PAGES} pages = ${TRESERVE}"
    echo "  MemFree            = ${MF}"
    echo "  pagecache          = Active(file) ${AF} + Inactive(file) ${IF} = ${PAGECACHE}"
    echo "  KReclaimable       = ${KR}"
    hr
    echo "  (MemFree - totalreserve)          = $(( MF - TRESERVE ))"
    echo "  + (pagecache   - min(half,wlow))  = $(( PAGECACHE - cut1 ))   [차감 ${cut1}]"
    echo "  + (reclaimable - min(half,wlow))  = $(( KR - cut2 ))   [차감 ${cut2}]"
    echo "  = 재현값                          = ${AVAIL}"
    echo "    커널 MemAvailable               = ${MA}"
    echo "    차이                            = $(( AVAIL - MA ))  ($(awk -v a="$AVAIL" -v m="$MA" 'BEGIN { if (m == 0) print "n/a"; else printf "%.4f", (a - m) * 100 / m }')%)"
    echo
    echo "  판정: 오차 1% 이내면 PASS. 스냅샷 드리프트가 대부분이다."
}

# -------------------------------------------------------------- check-commit
check_commit() {
    snap
    local MT ST CL CA_S OM OR calc
    MT=$(g MemTotal); ST=$(g SwapTotal); CL=$(g CommitLimit); CA_S=$(g Committed_AS)
    OM=$(sysctl -n vm.overcommit_memory 2>/dev/null || echo '?')
    OR=$(sysctl -n vm.overcommit_ratio  2>/dev/null || echo '?')

    echo "[check-commit] CommitLimit 검산  (단위 kB)"
    hr
    echo "  vm.overcommit_memory = ${OM}   (0=휴리스틱, 1=항상 허용, 2=엄격)"
    echo "  vm.overcommit_ratio  = ${OR}"
    echo "  MemTotal   = ${MT}"
    echo "  SwapTotal  = ${ST}"
    if [ "$OR" != '?' ]; then
        calc=$(( ST + MT * OR / 100 ))
        echo "  계산식     = SwapTotal + MemTotal * ratio/100 = ${calc}"
    fi
    echo "  CommitLimit(커널) = ${CL}"
    echo "  Committed_AS      = ${CA_S}   ($(pct "$CA_S" "$CL")% of CommitLimit)"
    echo
    echo "  주의: Committed_AS 는 '약속한 양'이지 '쓰고 있는 양'이 아니다."
    echo "        overcommit_memory=0 이면 CommitLimit 은 강제되지 않는다."
}

# ------------------------------------------------------------- compare-usage
compare_usage() {
    snap
    local MT MF MA BU CA SR
    MT=$(g MemTotal); MF=$(g MemFree); MA=$(g MemAvailable)
    BU=$(g Buffers);  CA=$(g Cached);  SR=$(g SReclaimable)
    local BC=$(( BU + CA + SR ))

    echo "[compare-usage] 같은 순간, 네 가지 '메모리 사용률'  (단위 kB)"
    hr
    printf '%-46s %10s  %s\n' "계산식" "결과" "무엇을 재는가"
    printf '%-46s %9s%%  %s\n' "(MemTotal - MemFree) / MemTotal"          "$(pct $((MT - MF)) "$MT")" "커널이 손대지 않은 페이지 외 전부. 캐시 포함"
    printf '%-46s %9s%%  %s\n' "(MemTotal - MemAvailable) / MemTotal"     "$(pct $((MT - MA)) "$MT")" "새 할당에 내줄 수 없는 몫. 커널 추정"
    printf '%-46s %9s%%  %s\n' "(MemTotal - MemFree - buffcache) / Total" "$(pct $((MT - MF - BC)) "$MT")" "익명 페이지 + 회수 불가 커널 메모리"
    printf '%-46s %9s%%  %s\n' "buffcache / MemTotal"                     "$(pct "$BC" "$MT")" "캐시가 차지한 몫 (참고)"
    hr
    echo "  MemTotal ${MT} / MemFree ${MF} / MemAvailable ${MA} / buff+cache ${BC}"
    echo
    echo "  같은 서버, 같은 순간이다. 어느 식을 쓰느냐로 수십 %p 가 갈린다."
}

# --------------------------------------------------------------- watch-usage
watch_usage() {
    local n=${1:-12} iv=${2:-5}
    echo "[watch-usage] ${iv}초 간격 ${n}회 — 계산식별 변동폭 관찰"
    hr
    printf '%-10s %12s %12s %12s %12s\n' "시각" "free기준%" "avail기준%" "MemFree(MB)" "buffcache(MB)"
    local i
    for (( i = 0; i < n; i++ )); do
        snap
        local MT MF MA BC
        MT=$(g MemTotal); MF=$(g MemFree); MA=$(g MemAvailable)
        BC=$(( $(g Buffers) + $(g Cached) + $(g SReclaimable) ))
        printf '%-10s %12s %12s %12s %12s\n' "$(date +%H:%M:%S)" \
            "$(pct $((MT - MF)) "$MT")" "$(pct $((MT - MA)) "$MT")" \
            "$(( MF / 1024 ))" "$(( BC / 1024 ))"
        [ $(( i + 1 )) -lt "$n" ] && sleep "$iv"
    done
    hr
    echo "  파일 I/O 가 있는 시간대에 돌리면 두 열의 변동폭 차이가 뚜렷해진다."
    echo "  MemFree 가 움직이는 동안 avail기준 이 거의 고정이면 3.4절의 불변식이 성립한 것이다."
}

# ------------------------------------------------------------------ snapshot
snapshot() {
    snap
    echo "===== uname ====="            ; uname -r
    echo "===== free -w -m ====="       ; free -w -m
    echo "===== /proc/meminfo ====="    ; cat "$MI"
    echo "===== 워터마크 sysctl ====="  ; sysctl vm.min_free_kbytes vm.watermark_scale_factor vm.swappiness \
                                                  vm.dirty_ratio vm.dirty_background_ratio \
                                                  vm.overcommit_memory vm.overcommit_ratio 2>&1
    echo "===== /proc/zoneinfo 워터마크 ====="
    awk '/^Node .*zone/ { z = $0 } /^ +(pages free|min|low|high|protection:)/ { print z " | " $0 }' "$ZI" 2>/dev/null | head -40
    echo "===== /proc/vmstat 발췌 ====="
    grep -E '^(pgpgin|pgpgout|pswpin|pswpout|pgfault|pgmajfault|pgscan_kswapd|pgsteal_kswapd|pgscan_direct|pgsteal_direct|allocstall_|oom_kill|nr_free_pages|nr_dirty|nr_writeback) ' /proc/vmstat
    echo "===== PSI ====="              ; cat /proc/pressure/memory 2>&1
    echo "===== cgroup v2 ====="
    for f in memory.max memory.high memory.current memory.events; do
        [ -r "/sys/fs/cgroup/$f" ] && echo "-- $f: $(cat "/sys/fs/cgroup/$f")"
    done
    head -12 /sys/fs/cgroup/memory.stat 2>/dev/null
}

case "${1:-all}" in
    check-free)    check_free ;;
    check-avail)   check_avail ;;
    check-commit)  check_commit ;;
    compare-usage) compare_usage ;;
    watch-usage)   shift; watch_usage "$@" ;;
    snapshot)      snapshot ;;
    all)           check_free; echo; check_avail; echo; check_commit; echo; compare_usage ;;
    *)             sed -n '2,20p' "$0"; exit 2 ;;
esac
