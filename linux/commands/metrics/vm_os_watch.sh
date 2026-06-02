#!/bin/bash
# ══════════════════════════════════════════════════════════════════════════════
# vm_os_watch.sh — VMware 게스트(Oracle Linux 8) 1초 주기 OS 상태 상주 모니터
#
# 무엇을 하는가
#   /proc 의 누적 카운터를 1초 간격으로 두 번 떠서 그 차이(Δ)로 CPU·부하·메모리·
#   스왑·디스크·네트워크의 "그 1초"를 계산해 보여 준다. 임계값을 넘은 수치는
#   색(노랑=주의, 빨강=위험)으로 강조하고, 동시에 경보 이력 파일에 남긴다.
#   의존성은 bash + gawk + coreutils 뿐이다 (sysstat 불필요).
#
# 사용법
#   bash vm_os_watch.sh [옵션] [간격초]
#     (없음)        1초 주기 상주 모드. Ctrl-C 로 종료
#     2             2초 주기
#     -1, --once    한 프레임만 일반 텍스트로 출력하고 종료
#                   — 화면 제어코드 0건이라 티켓 첨부·증거 캡처·cron 에 적합
#     -l, --log P   경보 이력 파일 경로 지정 (예: --log /tmp/vm_os_watch_alerts.log)
#     -h, --help    이 도움말
#
# 경보 이력 (로그)
#   위치 결정 순서: ① --log 인자 ② 환경변수 LOGF ③ 자동 선택
#     자동 선택: /var/log 쓰기 가능 → /var/log/vm_os_watch_alerts.log
#               아니면 → ~/.local/state/vm_os_watch/alerts.log
#     ※ 과거 기본값 /tmp 는 재부팅·정리 정책으로 파일이 사라질 수 있어 기본에서
#       제외했다. /tmp 에 남기고 싶으면 --log /tmp/... 로 명시하면 된다.
#   과거 위치(/tmp)에 이력이 남아 있으면 새 파일을 처음 만들 때 1회 이어붙인다.
#   포맷: "YYYY-MM-DD HH:MM:SS [레벨] 메시지" (레벨 = INFO|WARN|CRIT)
#     시작·회전·이관 같은 사건은 [INFO] 로 같은 파일에 남아, 이력만 읽어도
#     "언제 모니터가 떠 있었는지"까지 재구성할 수 있다.
#   회전: ROTATE_MB(기본 5)MB 초과 시 .1 로 한 개 보관 후 같은 경로에 새로 시작.
#         ROTATE_MB=0 이면 회전 없음.
#
# 환경 변수 (전부 선택)
#   INTERVAL=1            갱신 주기(초) — 첫 위치 인자와 동일
#   DISKS="sda dm-2"      감시 디스크 지정 (기본: sd*/vd*/xvd*/nvme* 자동 감지)
#   LOGF=경로             경보 이력 파일 (--log 과 동일, 인자가 우선)
#   ROTATE_MB=5           로그 회전 기준 MB (0=끔)
#   LEGACY_LOG=/tmp/...   이어붙일 과거 로그 위치 (기본: /tmp/vm_os_watch_alerts.log)
#   W_*/C_*               임계값(주의/위험) — 아래 "임계값" 참고. 0 이면 해당 경보 끔
#
# 임계값 기본 (통상 관례 — 환경에 맞게 조정. all-flash 어레이면 await 10/30 권장)
#   W_CPU_WA/C_CPU_WA=20/40(%)  W_CPU_ST/C_CPU_ST=5/15(%)
#   W_LOAD/C_LOAD=100/200(load1 ÷ vCPU %)  W_MEMA/C_MEMA=15/7(가용 % 이하)
#   C_SWAP=2048(so KiB/s 위험; so>0 주의, si 단독은 정보성 주의)
#   W_AWAIT/C_AWAIT=20/50(ms)  W_UTIL/C_UTIL=90/99(%)  W_AQU/C_AQU=8/32  W_D/C_D=1/5
#
# 지표가 "왜 그렇게 측정되는가"는 아래 본문(특히 AWKPROG 안)의 주석에 계산
# 지점마다 적어 두었다. 배경 원리·판정법은 storage_io_vmware.md (특히 3·8·9장).
#
# 검증: tests/run_tests.sh — 수치·경보·정렬 회귀를 1분 안에 확인한다.
# 참고: 물리 호스트(예: ProLiant DL380 Gen10)의 하드웨어 상태는 게스트에서
#       보이지 않는다. 그쪽은 iLO / ESXi(esxtop·vCenter) 영역이다.
# ══════════════════════════════════════════════════════════════════════════════

# ── 인자 파싱 ─────────────────────────────────────────────────────────────────
ONCE=0; LOGF_ARG=""; INTERVAL_ARG=""
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            # 도움말 = 이 파일 상단의 주석 블록 (셔뱅 다음부터 첫 비주석 줄 전까지)
            awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"; exit 0 ;;
        -1|--once) ONCE=1 ;;
        -l|--log)
            [ $# -ge 2 ] || { echo "오류: $1 다음에 경로가 필요합니다 (-h 참고)" >&2; exit 2; }
            LOGF_ARG="$2"; shift ;;
        --log=*)   LOGF_ARG="${1#--log=}" ;;
        ''|*[!0-9]*) echo "오류: 알 수 없는 인자 '$1' (-h 참고)" >&2; exit 2 ;;
        *)         INTERVAL_ARG="$1" ;;          # 숫자만 남음 = 갱신 주기
    esac
    shift
done

INTERVAL="${INTERVAL_ARG:-${INTERVAL:-1}}"
COOLDOWN="${COOLDOWN:-30}"          # 같은 경보를 이력에 다시 적기까지의 간격(초).
                                    # 지속 장애 때 초당 1줄씩 쌓이는 것을 막되,
                                    # 화면 색은 매 프레임 유지되므로 인지에는 영향 없다.
DISKS="${DISKS:-}"
ROTATE_MB="${ROTATE_MB:-5}"; case "$ROTATE_MB" in ''|*[!0-9]*) ROTATE_MB=5 ;; esac
LEGACY_LOG="${LEGACY_LOG:-/tmp/vm_os_watch_alerts.log}"

# 임계값 — 모두 환경변수로 덮어쓸 수 있고, 0 이면 그 경보를 끈다(lvl_* 함수가 0 을 무시).
W_CPU_WA=${W_CPU_WA:-20}  C_CPU_WA=${C_CPU_WA:-40}
W_CPU_ST=${W_CPU_ST:-5}   C_CPU_ST=${C_CPU_ST:-15}
W_LOAD=${W_LOAD:-100}     C_LOAD=${C_LOAD:-200}
W_MEMA=${W_MEMA:-15}      C_MEMA=${C_MEMA:-7}
C_SWAP=${C_SWAP:-2048}
W_AWAIT=${W_AWAIT:-20}    C_AWAIT=${C_AWAIT:-50}
W_UTIL=${W_UTIL:-90}      C_UTIL=${C_UTIL:-99}
W_AQU=${W_AQU:-8}         C_AQU=${C_AQU:-32}
W_D=${W_D:-1}             C_D=${C_D:-5}

# ── 터미널 준비 ───────────────────────────────────────────────────────────────
# 색은 "터미널일 때만" 켠다. 파일로 리다이렉트하거나 --once 캡처할 때
# 제어코드가 섞이면 캡처물을 읽을 수 없게 되기 때문이다.
if [ -t 1 ]; then
    YEL=$'\033[33m'; RED=$'\033[31;1m'; DIM=$'\033[2m'; RST=$'\033[0m'
else
    YEL=''; RED=''; DIM=''; RST=''
fi

TMPD=$(mktemp -d /tmp/vm_os_watch.XXXXXX) || exit 1
PREV="$TMPD/prev"; CUR="$TMPD/cur"
# 종료 시: 임시 디렉터리 정리 + (상주·터미널 모드였다면) 숨겼던 커서 복원
trap 'rm -rf "$TMPD"; [ "$ONCE" = 0 ] && [ -t 1 ] && printf "\033[?25h\033[0m\n"' EXIT
trap 'exit 0' INT TERM
# 터미널 크기가 바뀌면 잔상이 남으므로 전체를 한 번 지운다
trap '[ -t 1 ] && printf "\033[2J"' WINCH

NCPU=$(nproc)
VIRT=$(systemd-detect-virt 2>/dev/null || echo '?')
HAVE_VMT=0; command -v vmware-toolbox-cmd >/dev/null 2>&1 && HAVE_VMT=1
BALLOON_MB=-1; HVSWAP_MB=-1

# ── 경보 이력 파일 ────────────────────────────────────────────────────────────
resolve_log_path() {
    # 위치 결정 우선순위: ① --log 인자 ② 환경변수 LOGF ③ 자동 선택.
    # ③이 /tmp 를 피하는 이유: /tmp 는 재부팅에 비워질 수 있고 systemd-tmpfiles 가
    # 오래된 파일을 정리(OL8 관례 10일)하므로, "화면을 안 보던 사이의 스파이크를
    # 보존한다"는 이력의 존재 이유와 양립하지 않는다 (실제 소실 사례가 있었다 —
    # storage_io_vmware.md 10장의 배경). /tmp 를 원하면 --log 로 명시하면 된다.
    if   [ -n "$LOGF_ARG" ]; then LOGF="$LOGF_ARG"
    elif [ -n "${LOGF:-}" ]; then :                       # 기존 env 사용자 호환
    elif [ -w /var/log ];    then LOGF="/var/log/vm_os_watch_alerts.log"
    else LOGF="${XDG_STATE_HOME:-$HOME/.local/state}/vm_os_watch/alerts.log"
    fi
    if ! mkdir -p "$(dirname "$LOGF")" 2>/dev/null; then
        LOGF="/tmp/vm_os_watch_alerts.log"                # 최후 수단 — 화면에 휘발 경고
    fi
    LOG_VOLATILE=0; case "$LOGF" in /tmp/*) LOG_VOLATILE=1 ;; esac
}

migrate_legacy_log() {
    # 과거 기본 위치에 쌓인 이력이 있으면 "새 파일을 처음 만들 때 1회" 이어붙인다.
    # 원본은 지우지 않는다 — 사용자의 파일을 파괴하지 않기 위해.
    [ "$LOGF" = "$LEGACY_LOG" ] && return 0
    [ -s "$LEGACY_LOG" ] || return 0
    [ -e "$LOGF" ] && return 0
    {
        printf '%s [INFO] 이전 위치(%s)의 이력을 이어붙임\n' "$(date '+%F %T')" "$LEGACY_LOG"
        cat "$LEGACY_LOG"
    } >> "$LOGF"
}

log_line() { # $1=레벨(INFO|WARN|CRIT)  $2=메시지
    # 회전: ROTATE_MB 초과 시 .1 로 한 개만 보관하고 같은 경로에 새로 시작한다.
    # mv 방식이 안전한 이유: 이 함수는 매번 >> 로 열고 닫으므로(파일 핸들을 들고
    # 있지 않으므로) mv 직후의 다음 기록이 같은 경로에 새 파일을 만든다.
    if [ "$ROTATE_MB" -gt 0 ] && [ -f "$LOGF" ]; then
        local size; size=$(stat -c %s "$LOGF" 2>/dev/null || echo 0)
        if [ "$size" -ge $(( ROTATE_MB * 1024 * 1024 )) ]; then
            mv -f "$LOGF" "$LOGF.1"
            printf '%s [INFO] 로그 회전 — 이전 분량: %s\n' "$(date '+%F %T')" "$LOGF.1" >> "$LOGF"
        fi
    fi
    printf '%s [%s] %s\n' "$(date '+%F %T')" "$1" "$2" >> "$LOGF"
}

resolve_log_path
migrate_legacy_log

# ── VMware Tools (있을 때만) ─────────────────────────────────────────────────
refresh_vmt() {
    # balloon/hvswap 은 ESXi 가 게스트 메모리를 회수·스왑한 양으로, 게스트 안에서
    # 호스트 메모리 압박을 볼 수 있는 거의 유일한 창이다 (원리: 문서 4.5절).
    # 초 단위로 변하는 값이 아니므로 5주기마다 갱신한다 (메인 루프 참조).
    [ "$HAVE_VMT" = 1 ] || return 0
    BALLOON_MB=$(vmware-toolbox-cmd stat balloon 2>/dev/null | awk '{print $1+0}')
    HVSWAP_MB=$(vmware-toolbox-cmd stat swap 2>/dev/null | awk '{print $1+0}')
    [ -n "$BALLOON_MB" ] || BALLOON_MB=-1
    [ -n "$HVSWAP_MB" ]  || HVSWAP_MB=-1
}

# ── 스냅샷 ────────────────────────────────────────────────────────────────────
snap() {
    # /proc 의 카운터들은 전부 "부팅 이후 누적"이라 한 번 읽어서는 속도를 알 수
    # 없다. 그래서 필요한 파일을 통째로 한 파일에 떠 두고(섹션 구분자 #XXXX),
    # 이전/현재 두 스냅샷의 차이를 AWKPROG 가 한 번에 계산한다.
    # 이렇게 하면 주기당 외부 프로세스 호출이 약 6회로 끝난다 (모니터 자체의
    # 부하가 측정 대상을 오염시키지 않게 하기 위한 설계).
    {
        # Δ시간(dt)의 원천. "초.나노초" 형식을 쓰는 이유: %3N(밀리초 절단)은
        # date 구현에 따라 무시되어 나노초가 통째로 붙는 경우가 있고(실측 사례),
        # 그러면 dt 가 백만 배로 부풀어 모든 "초당" 값이 0 으로 보인다.
        # %s.%N 은 awk 가 부동소수 초로 바로 빼면 되고, %N 미지원 환경에서는
        # 소수부가 버려져 초 단위로 자연 강등될 뿐 폭주하지 않는다.
        printf '#T %s\n' "$(date +%s.%N)"
        printf '#STAT\n'; cat /proc/stat
        printf '#DISK\n'; cat /proc/diskstats
        printf '#NET\n';  cat /proc/net/dev
        printf '#VM\n';   grep -E '^pswp(in|out) ' /proc/vmstat
        printf '#MEM\n';  cat /proc/meminfo
        printf '#LOAD\n'; cat /proc/loadavg
    } > "$1"
}

# ══════════════════════════════════════════════════════════════════════════════
# 계산·표시 본체 (gawk)
#   입력: 스냅샷 2개 (ARGV[1]=이전, ARGV[2]=현재)
#   출력: 화면 본문 행 + 경보 행("@A@|레벨|키|메시지") + 상태 행("@S@|위험수|주의수")
#   표시 규칙: 값은 왼쪽 10칸 고정폭, 라벨은 오른쪽 자유폭.
#     이유 — 한글은 터미널에서 2칸 폭이라 라벨을 왼쪽에 두면 printf 패딩으로는
#     열이 반드시 어긋난다. 값을 앞에 두면 정렬 문제가 원천적으로 사라진다.
# ══════════════════════════════════════════════════════════════════════════════
AWKPROG='
# ── 공용 함수 ──────────────────────────────────────────────────────────────
# lvl_high: 값이 클수록 나쁜 지표의 단계 판정 (0=정상 1=주의 2=위험).
#           임계가 0 이면 "경보 끔"으로 취급한다 (w>0, c>0 조건).
function lvl_high(v, w, c) { if (c>0 && v>=c) return 2; if (w>0 && v>=w) return 1; return 0 }
# lvl_low: 값이 작을수록 나쁜 지표(가용 메모리 %)의 단계 판정.
function lvl_low(v, w, c)  { if (v<=c) return 2; if (v<=w) return 1; return 0 }
# paint: 단계에 맞는 색을 입힌다. "패딩을 먼저, 색은 그 위에" — 색 코드는
#        보이지 않는 글자라서 색을 먼저 입히면 printf 폭 계산이 깨지기 때문.
function paint(s, l) { if (l==2) return RED s RST; if (l==1) return YEL s RST; return s }
# alert: 경보 행 발행 + 상태 집계. 실제 이력 기록(쿨다운 포함)은 bash 쪽 책임.
function alert(sev, key, msg) {
    if (sev>=2) NCRIT++; else NWARN++
    printf("@A@|%d|%s|%s\n", sev, key, msg)
}
# row: "값(10칸 우측정렬) + 공백3 + 라벨" 행. 레이아웃 불변식의 구현 지점.
function row(val, label) { printf("  %s   %s\n", val, label) }
# dpos: 누적 카운터의 Δ. 음수면 0 으로 — 디바이스 재생성·카운터 리셋 직후의
#       한 프레임이 거대한 음수/허위 경보로 나타나는 것을 막는다.
function dpos(cur, prv) { cur -= prv; return cur < 0 ? 0 : cur }

# ── 입력 파싱 ──────────────────────────────────────────────────────────────
# ix: 지금 줄이 이전(0)/현재(1) 어느 스냅샷 소속인지. 매 줄 갱신된다.
{ ix = (FILENAME == ARGV[1]) ? 0 : 1 }
$1=="#T"    { SNAPT[ix]=$2; next }
$1=="#STAT" { sec=1; next }
$1=="#DISK" { sec=2; next }
$1=="#NET"  { sec=3; next }
$1=="#VM"   { sec=4; next }
$1=="#MEM"  { sec=5; next }
$1=="#LOAD" { sec=6; next }

# /proc/stat — "cpu " 행은 부팅 이후 누적 틱(USER_HZ). 필드(proc(5)):
#   2 user / 3 nice / 4 system / 5 idle / 6 iowait / 7 irq / 8 softirq / 9 steal
#   (guest 틱은 user 에 포함되므로 2..9 합 = 전체 시간)
# procs_running/blocked 는 누적이 아니라 "지금 순간"의 개수다.
sec==1 && $1=="cpu"           { for (i=2; i<=NF && i<=11; i++) CPU[ix,i]=$i; next }
sec==1 && $1=="procs_running" { RUNNING[ix]=$2; next }
sec==1 && $1=="procs_blocked" { BLOCKED[ix]=$2; next }

# /proc/diskstats — 디바이스별 누적 카운터. 쓰는 필드(커널 admin-guide/iostats):
#   $4 읽기 완료 수      $6 읽은 섹터(512B)   $7 읽기 누적 소요 ms
#   $8 쓰기 완료 수      $10 쓴 섹터          $11 쓰기 누적 소요 ms
#   $13 io_ticks: I/O 가 1개라도 진행 중이던 누적 ms  → util 의 원천
#   $14 weighted: 진행 중 개수×시간의 적분(ms·개)     → 평균 큐 길이의 원천
# 파티션(sda1)이 아닌 디스크 전체(sda)만 잡는다 — 합산이 중복되지 않게.
sec==2 {
    name=$3
    if (DISKS != "") ok = index(" " DISKS " ", " " name " ") > 0
    else             ok = (name ~ /^(sd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme[0-9]+n[0-9]+)$/)
    if (ok) {
        if (ix==1 && !(name in SEEN)) { DISK_ORDER[++DISK_N]=name; SEEN[name]=1 }
        DSK[ix,name,"rd"]=$4;    DSK[ix,name,"rdsec"]=$6;  DSK[ix,name,"rdms"]=$7
        DSK[ix,name,"wr"]=$8;    DSK[ix,name,"wrsec"]=$10; DSK[ix,name,"wrms"]=$11
        DSK[ix,name,"ticks"]=$13; DSK[ix,name,"wticks"]=$14
    }
    next
}

# /proc/net/dev — "iface: rx_bytes ... " 형식. 콜론을 공백으로 바꿔 분리한다.
#   f2 rx_bytes / f4 rx_errs / f5 rx_drop / f10 tx_bytes / f12 tx_errs / f13 tx_drop
# 오류와 드롭은 합쳐서 본다 — 어느 쪽이든 "늘고 있다" 자체가 이상 신호라서다.
sec==3 && /:/ && $0 !~ /\|/ {
    line=$0; sub(/:/," ",line); split(line, f, " ")
    if (f[1] != "lo") {                       # 루프백은 진단 가치가 없어 제외
        if (ix==1 && !(f[1] in NSEEN)) { NET_ORDER[++NET_N]=f[1]; NSEEN[f[1]]=1 }
        NETC[ix,f[1],"rxb"]=f[2];  NETC[ix,f[1],"rxe"]=f[4]+f[5]
        NETC[ix,f[1],"txb"]=f[10]; NETC[ix,f[1],"txe"]=f[12]+f[13]
    }
    next
}

# /proc/vmstat — pswpin/pswpout 은 스왑된 "페이지 수" 누적 (1페이지 = 4KiB).
sec==4 { PSWP[ix,$1]=$2; next }

# /proc/meminfo — 누적이 아닌 현재값이므로 현재 스냅샷만 쓴다.
sec==5 && ix==1 { MEM[$1]=$2; next }

# /proc/loadavg — 1/5/15분 지수평균. 현재 스냅샷만 쓴다.
sec==6 && ix==1 { L1=$1; L5=$2; L15=$3; next }

END {
    # Δ시간(초). 1초를 요청해도 실측은 1.0x초이므로 모든 "초당" 값은 실측 dt 로
    # 나눈다. #T 가 "초.나노초"라 빼기만 하면 초가 된다 (double 정밀도로 µs 충분).
    dt = SNAPT[1] - SNAPT[0]; if (dt <= 0) dt = 1

    # ── CPU ─────────────────────────────────────────────────────────────────
    # 누적 틱의 Δ를 전체 Δ로 나누면 "그 구간의 시간 비율"이 된다.
    # us 에 nice 를 합치는 이유: 둘 다 사용자 코드 실행이고 우선순위만 다르다.
    # sy 에 irq/softirq 를 합치는 이유: 모두 커널이 소비한 시간이라 운영 판단
    # 단위에서는 한 묶음이 읽기 쉽다 (세분이 필요하면 mpstat 영역).
    total = 0
    for (i=2; i<=9; i++) { dcpu[i] = dpos(CPU[1,i], CPU[0,i]); total += dcpu[i] }
    if (total <= 0) total = 1
    us = 100*(dcpu[2]+dcpu[3])/total
    sy = 100*(dcpu[4]+dcpu[7]+dcpu[8])/total
    wa = 100*dcpu[6]/total                  # iowait: "놀면서" I/O 를 기다린 시간만
    st = 100*dcpu[9]/total                  # steal: 하이퍼바이저가 vCPU 를 안 준 시간
    id = 100*dcpu[5]/total
    lwa = lvl_high(wa, W_CPU_WA, C_CPU_WA)
    lst = lvl_high(st, W_CPU_ST, C_CPU_ST)

    printf("[ CPU ]  /proc/stat %s초 차분\n", sprintf("%.0f", dt))
    row(sprintf("%9.1f%%", us),            "사용자 코드 실행 (us, user)")
    row(sprintf("%9.1f%%", sy),            "커널 코드 실행 (sy, system+irq+softirq)")
    row(paint(sprintf("%9.1f%%", wa), lwa), "I/O 완료 대기 (wa, iowait) — 힌트일 뿐, 판정은 디스크 평균응답으로")
    row(paint(sprintf("%9.1f%%", st), lst), "하이퍼바이저에 빼앗긴 시간 (st, steal) — VMware 는 0 고정일 수 있음")
    row(sprintf("%9.1f%%", id),            "유휴 (id, idle)")
    if (lwa) alert(lwa, "cpu_wa", sprintf("CPU I/O 대기(wa, iowait) %.1f%% — 디스크 평균응답 확인 필요", wa))
    if (lst) alert(lst, "cpu_st", sprintf("CPU 하이퍼바이저 경합(st, steal) %.1f%%", st))

    # ── 부하 · 프로세스 ─────────────────────────────────────────────────────
    # load 는 "실행 대기 + D state" 태스크의 지수평균이라, 디스크가 멈춰도
    # 올라간다. vCPU 수 대비 %로 환산해야 "만석"인지 판단할 수 있다.
    lp = int(100*L1/NCPU + 0.5)
    ll = lvl_high(lp, W_LOAD, C_LOAD)
    b  = BLOCKED[1] + 0                     # D state: I/O 응답을 기다리며 시그널도
    lb = lvl_high(b, W_D, C_D)              # 못 받는 상태 — 지속되면 스토리지 증거
    printf("\n[ 부하 · 프로세스 ]  /proc/loadavg · /proc/stat\n")
    row(sprintf("%10.2f", L1), sprintf("평균 부하 1분 (load average 1m) — vCPU %d 개의 %s", NCPU, paint(sprintf("%d%%", lp), ll)))
    row(sprintf("%10s", sprintf("%.2f/%.2f", L5, L15)), "평균 부하 5분/15분 (load average 5m/15m)")
    row(sprintf("%10d", RUNNING[1]+0),     "실행 대기 프로세스 (r, procs_running)")
    row(paint(sprintf("%10d", b), lb),     "디스크 응답 대기 고착 프로세스 (D, uninterruptible sleep) — 지속 시 스토리지 의심")
    if (ll) alert(ll, "load",   sprintf("1분 부하(load1) %.2f = vCPU %d 개의 %d%%", L1, NCPU, lp))
    if (lb) alert(lb, "proc_d", sprintf("디스크 응답 대기(D state) 프로세스 %d개", b))

    # ── 메모리 · 스왑 ───────────────────────────────────────────────────────
    # 부족 판정은 free 가 아니라 MemAvailable 로 한다 — 리눅스는 남는 메모리를
    # 캐시로 쓰므로 free 는 항상 작아 보이고, "캐시 중 회수 가능한 몫"까지 더한
    # MemAvailable 이 실제로 쓸 수 있는 양이기 때문이다.
    mt  = MEM["MemTotal:"];  ma = MEM["MemAvailable:"]
    ca  = MEM["Cached:"] + MEM["Buffers:"]
    drt = MEM["Dirty:"]                     # 아직 디스크로 안 내려간 쓰기 (writeback 대기)
    ap  = int(100*ma/mt + 0.5)
    lm  = lvl_low(ap, W_MEMA, C_MEMA)
    su  = MEM["SwapTotal:"] - MEM["SwapFree:"]
    # si/so: 페이지 수 Δ × 4(KiB/페이지) ÷ dt = KiB/s.
    # 신호가 비대칭이다 — so(내려씀)>0 은 "지금" 메모리 압박이 진행 중이라는 뜻,
    # si(읽어들임)만 >0 이면 과거에 스왑아웃된 페이지를 늦게 되읽는 잔재 회수일
    # 수 있다(리눅스는 메모리가 남아도 스왑 페이지를 선제 복귀시키지 않는다).
    # 실측 사례 해체: storage_io_vmware.md 10장.
    si  = int(dpos(PSWP[1,"pswpin"],  PSWP[0,"pswpin"])  * 4 / dt + 0.5)
    so  = int(dpos(PSWP[1,"pswpout"], PSWP[0,"pswpout"]) * 4 / dt + 0.5)
    lso = 0; if (so > 0) lso = (so >= C_SWAP) ? 2 : 1
    lsi = 0; if (si > 0) lsi = (lso > 0) ? lso : 1

    printf("\n[ 메모리 · 스왑 ]  /proc/meminfo · /proc/vmstat\n")
    row(paint(sprintf("%9.1fG", ma/1048576), lm), sprintf("실제 가용 메모리 (avail, MemAvailable) — 전체 %.1fG 의 %s", mt/1048576, paint(sprintf("%d%%", ap), lm)))
    row(sprintf("%9.1fG", ca/1048576),      "페이지 캐시 (cached+buffers) — 줄어들면 읽기가 디스크로 떨어짐")
    row(sprintf("%9dM", int(drt/1024)),     "디스크 미기록 쓰기 (dirty) — writeback 대기 중인 양")
    row(sprintf("%9dM", int(su/1024)),      "스왑 사용량 (swap used)")
    row(paint(sprintf("%7dK/s", si), lsi),  "스왑 읽어들임 (si, swap-in) — so=0 이고 가용 충분하면 과거 잔재 회수일 수 있음")
    row(paint(sprintf("%7dK/s", so), lso),  "스왑 내려씀 (so, swap-out) — 0 이 아니면 지금 메모리 압박이 진행 중")
    if (BALLOON < 0 && HVSWAP < 0) {
        printf("  %s\n", DIM "(open-vm-tools 미설치 — balloon·hvswap 확인 불가)" RST)
    } else {
        row(paint(sprintf("%9dM", BALLOON), BALLOON>0 ? 1 : 0), "VMware 풍선 회수 (balloon, vmmemctl) — >0 이면 호스트가 메모리 회수 중")
        row(paint(sprintf("%9dM", HVSWAP),  HVSWAP>0  ? 2 : 0), "VMware 하이퍼바이저 스왑 (hvswap, host-level swap) — >0 이면 심각")
    }
    if (lm)          alert(lm, "mem_avail",   sprintf("가용 메모리(avail) %d%% (%.1fG)", ap, ma/1048576))
    if (lso)         alert(lso, "swap_out",   sprintf("스왑 내려씀(so) %dK/s — 지금 메모리 압박 진행 중", so))
    else if (lsi)    alert(1, "swap_in",      sprintf("스왑 읽어들임(si) %dK/s (so=0) — 과거 스왑 잔재 회수 가능성, 가용·스왑사용량 확인", si))
    if (BALLOON > 0) alert(1, "vmw_balloon",  sprintf("VMware 풍선 회수(balloon) %dMB — 호스트가 메모리 회수 중", BALLOON))
    if (HVSWAP > 0)  alert(2, "vmw_hvswap",   sprintf("VMware 하이퍼바이저 스왑(hvswap) %dMB — 호스트 메모리 심각", HVSWAP))

    # ── 디스크 ──────────────────────────────────────────────────────────────
    # iostat -x 와 같은 공식을 쓴다 (원천이 같으므로):
    #   평균응답 await = Δ(소요 ms 합) ÷ Δ(완료 수)   ← 큐 대기 + 처리 시간 전부 포함
    #   사용률  util  = Δio_ticks ÷ 경과 ms           ← "1개라도 진행 중"이던 시간 비율
    #   큐길이  aqu   = Δweighted ÷ 경과 ms           ← 평균 동시 진행 개수
    # 세 값은 리틀의 법칙으로 묶인다: aqu ≈ IOPS × await(초). 어긋나면 측정 의심.
    # 게스트의 await 에는 ESXi·데이터스토어·어레이 왕복이 전부 합산된다 — 그래서
    # await 가 높을 때 "어디가" 느린지는 esxtop(GAVG=KAVG+DAVG)으로만 가른다.
    printf("\n[ 디스크 ]  /proc/diskstats %s초 차분\n", sprintf("%.0f", dt))
    printf("  %s\n", DIM "초당요청(r/s·w/s, requests/sec) · 처리량(rMB/s·wMB/s) · 평균응답(await, average wait) · 큐길이(aqu, avg queue size) · 사용률(util, utilization)" RST)
    printf("  장치        r/s     w/s    rMB/s    wMB/s      await     aqu    util\n")
    for (i=1; i<=DISK_N; i++) {
        name = DISK_ORDER[i]
        if (!((0,name,"rd") in DSK)) continue   # 새로 나타난 디스크: 이전 값이 없으면 다음 프레임부터
        rd  = dpos(DSK[1,name,"rd"], DSK[0,name,"rd"])
        wr  = dpos(DSK[1,name,"wr"], DSK[0,name,"wr"])
        rmb = dpos(DSK[1,name,"rdsec"], DSK[0,name,"rdsec"]) * 512 / 1048576 / dt   # 섹터=512B
        wmb = dpos(DSK[1,name,"wrsec"], DSK[0,name,"wrsec"]) * 512 / 1048576 / dt
        ios = rd + wr
        awt = (ios > 0) ? (dpos(DSK[1,name,"rdms"], DSK[0,name,"rdms"]) + dpos(DSK[1,name,"wrms"], DSK[0,name,"wrms"])) / ios : 0
        ut  = 100 * dpos(DSK[1,name,"ticks"], DSK[0,name,"ticks"]) / (dt*1000); if (ut > 100) ut = 100
        aq  = dpos(DSK[1,name,"wticks"], DSK[0,name,"wticks"]) / (dt*1000)
        la = lvl_high(awt, W_AWAIT, C_AWAIT)
        lu = lvl_high(ut,  W_UTIL,  C_UTIL)
        lq = lvl_high(aq,  W_AQU,   C_AQU)
        printf("  %-9s %5d   %5d   %6.1f   %6.1f  %s  %s   %s\n",
            name, int(rd/dt+0.5), int(wr/dt+0.5), rmb, wmb,
            paint(sprintf("%7.1fms", awt), la), paint(sprintf("%6.1f", aq), lq), paint(sprintf("%4d%%", int(ut+0.5)), lu))
        if (la)               alert(la, "disk_" name "_await", sprintf("디스크 %s 평균응답(await) %.1fms (주의 %s / 위험 %s)", name, awt, W_AWAIT, C_AWAIT))
        if (lq >= 2)          alert(lq, "disk_" name "_aqu",   sprintf("디스크 %s 큐길이(aqu) %.1f", name, aq))
        if (lu >= 2 && la==0) alert(1,  "disk_" name "_util",  sprintf("디스크 %s 사용률(util) %d%% 포화 근접", name, int(ut)))
        # util 경보를 await 정상일 때만 내는 이유: 가상 디스크의 util 은 병렬성
        # 때문에 과대평가될 수 있어, 지연 동반 없는 100% 는 정보성으로만 알린다.
    }

    # ── 네트워크 ────────────────────────────────────────────────────────────
    # bytes Δ × 8 ÷ 10^6 = Mbit/s. 오류·드롭은 절대량보다 "증가" 자체가 신호다.
    printf("\n[ 네트워크 ]  /proc/net/dev %s초 차분\n", sprintf("%.0f", dt))
    for (i=1; i<=NET_N; i++) {
        nif = NET_ORDER[i]
        if (!((0,nif,"rxb") in NETC)) continue
        rxm  = dpos(NETC[1,nif,"rxb"], NETC[0,nif,"rxb"]) * 8 / 1000000 / dt
        txm  = dpos(NETC[1,nif,"txb"], NETC[0,nif,"txb"]) * 8 / 1000000 / dt
        errs = dpos(NETC[1,nif,"rxe"], NETC[0,nif,"rxe"]) + dpos(NETC[1,nif,"txe"], NETC[0,nif,"txe"])
        printf("  %-10s  수신 (rx) %8.2f Mb/s   송신 (tx) %8.2f Mb/s   오류·유실 (err+drop) %s\n",
            nif, rxm, txm, paint(sprintf("%d", errs), errs>0 ? 1 : 0))
        if (errs > 0) alert(1, "net_" nif, sprintf("네트워크 %s 오류·유실(err+drop) +%d", nif, errs))
    }

    # 상태 요약 — 이 프레임에서 발행된 경보의 수 (화면 머리에 표시된다)
    printf("@S@|%d|%d\n", NCRIT+0, NWARN+0)
}'

# ── 메인 루프 ─────────────────────────────────────────────────────────────────
declare -A LAST_LOGGED        # 경보 키 → 마지막 기록 시각(SECONDS). 쿨다운용

# 화면 전체 지우기·커서 숨김은 "상주 + 터미널"일 때만 — --once 캡처나 파일
# 리다이렉트에 제어코드가 섞이면 결과물을 읽을 수 없게 되기 때문이다.
[ "$ONCE" = 0 ] && [ -t 1 ] && printf '\033[2J\033[?25l'

refresh_vmt
snap "$PREV"
[ "$ONCE" = 0 ] && log_line INFO "─── 모니터 시작 (호스트 $(hostname), vCPU ${NCPU}, 간격 ${INTERVAL}s) ───"
LOOP=0

while :; do
    sleep "$INTERVAL"
    snap "$CUR"
    (( LOOP % 5 == 0 )) && refresh_vmt

    OUT_RAW=$(awk \
        -v NCPU="$NCPU" -v DISKS="$DISKS" -v BALLOON="$BALLOON_MB" -v HVSWAP="$HVSWAP_MB" \
        -v W_CPU_WA="$W_CPU_WA" -v C_CPU_WA="$C_CPU_WA" -v W_CPU_ST="$W_CPU_ST" -v C_CPU_ST="$C_CPU_ST" \
        -v W_LOAD="$W_LOAD" -v C_LOAD="$C_LOAD" -v W_MEMA="$W_MEMA" -v C_MEMA="$C_MEMA" \
        -v C_SWAP="$C_SWAP" -v W_AWAIT="$W_AWAIT" -v C_AWAIT="$C_AWAIT" \
        -v W_UTIL="$W_UTIL" -v C_UTIL="$C_UTIL" -v W_AQU="$W_AQU" -v C_AQU="$C_AQU" \
        -v W_D="$W_D" -v C_D="$C_D" \
        -v YEL="$YEL" -v RED="$RED" -v DIM="$DIM" -v RST="$RST" \
        "$AWKPROG" "$PREV" "$CUR")

    BODY=$(grep -v '^@[AS]@' <<<"$OUT_RAW")

    # 상태 요약 — awk 가 센 위험/주의 개수
    NCRIT=0; NWARN=0
    IFS='|' read -r _ NCRIT NWARN <<<"$(grep '^@S@' <<<"$OUT_RAW")"
    NCRIT=${NCRIT:-0}; NWARN=${NWARN:-0}
    if [ "$NCRIT" -gt 0 ] || [ "$NWARN" -gt 0 ]; then
        STATUS="상태  ${RED}위험 ${NCRIT}건${RST} · ${YEL}주의 ${NWARN}건${RST} — 아래 색칠된 항목"
    else
        STATUS="상태  정상 — 임계 초과 없음"
    fi
    [ "$LOG_VOLATILE" = 1 ] && STATUS="$STATUS   ${DIM}※ 이력이 휘발 위치(/tmp)에 기록 중 — --log 로 영속 경로 지정 가능${RST}"

    # 경보 → 이력 기록. 같은 키는 COOLDOWN 초에 1번만 적는다.
    # D state 경보는 "누가 어느 커널 함수에서 굳었는지"(wchan)를 함께 남긴다 —
    # 사후 분석에서 가장 먼저 필요한 정보라서다.
    while IFS='|' read -r _tag sev key msg; do
        [ -n "$key" ] || continue
        last=${LAST_LOGGED[$key]:--999}
        if (( SECONDS - last >= COOLDOWN )); then
            LAST_LOGGED[$key]=$SECONDS
            lab=WARN; [ "$sev" = 2 ] && lab=CRIT
            log_line "$lab" "$msg"
            if [ "$key" = proc_d ]; then
                ps -eo state,pid,wchan:20,comm --no-headers 2>/dev/null \
                    | awk '$1~/^D/{print "                      └ " $0}' >> "$LOGF"
            fi
        fi
    done < <(grep '^@A@' <<<"$OUT_RAW" || true)

    read -r _up _idle < /proc/uptime
    UPS=${_up%.*}; UPSTR="$((UPS/86400))d $((UPS%86400/3600))h $((UPS%3600/60))m"
    TITLE="── VM OS WATCH ── $(hostname) [${VIRT} guest, vCPU ${NCPU}] ── $(date '+%F %T') ── 가동 ${UPSTR} ── 갱신 ${INTERVAL}s ──"
    TAIL=$(tail -n 5 "$LOGF" 2>/dev/null)

    FRAME="${TITLE}
${STATUS}

${BODY}

[ 최근 경보 ]  임계 초과 시각 기록 — 전체 이력: ${LOGF}
${TAIL:-  (경보 없음)}

${DIM}[ 읽는 법 ]
  · 색: 노랑 = 주의, 빨강 = 위험 / 같은 경보는 ${COOLDOWN}초에 1번만 이력에 기록
  · 평균응답 (await, average wait) — 요청 1건 완료까지의 ms. 디스크 판정의 중심. 게스트 값엔 ESXi·어레이 왕복까지 합산
  · 큐길이 (aqu, average queue size) — 동시 진행 요청 수 ≈ IOPS x 응답(초). 계속 늘면 디스크가 못 따라가는 중
  · 사용률 (util, utilization) — 바빴던 시간 비율. 가상 디스크는 100% 여도 여유가 있을 수 있는 보조 지표
  · I/O 대기 (wa, iowait) 와 부하 (load) 는 힌트 — 디스크 문제 판정은 평균응답(await) 으로
  · 임계 주의/위험: wa ${W_CPU_WA}/${C_CPU_WA}% · await ${W_AWAIT}/${C_AWAIT}ms · 가용 ${W_MEMA}/${C_MEMA}% · D ${W_D}/${C_D} — W_*·C_* 환경변수로 조정 (0=끔)${RST}"

    if [ "$ONCE" = 1 ]; then
        # 캡처 모드: 위치 제어코드 없이 프레임 텍스트만. (색은 터미널일 때만 이미 적용)
        printf '%s\n' "$FRAME"
        break
    fi
    # 상주 모드: 커서를 홈으로 보내고(\033[H) 각 줄 끝을 지우며(\033[K) 덮어쓴 뒤
    # 남은 화면을 지운다(\033[0J) — 전체 클리어 방식보다 깜빡임이 없다.
    printf '\033[H%s\033[0J' "$(printf '%s' "$FRAME" | sed $'s/$/\033[K/')"

    mv "$CUR" "$PREV"
    (( LOOP++ ))
done
