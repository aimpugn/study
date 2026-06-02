#!/bin/bash
# vm_os_watch.sh — VMware 게스트(Oracle Linux 8) 1초 주기 OS 상태 상주 모니터
#
# 사용법
#   bash vm_os_watch.sh [간격초]          # 기본 1초. Ctrl-C 로 종료. -h 도움말
#   DISKS="sda dm-2" bash vm_os_watch.sh  # 감시 디스크 지정 (기본: sd*/vd*/nvme* 자동)
#   W_AWAIT=10 C_AWAIT=30 bash vm_os_watch.sh   # 임계값 덮어쓰기 (아래 변수 전부 가능)
#
# 설계
#   - 의존성: bash + gawk + coreutils 만 사용 (/proc 직접 계산, sysstat 불필요)
#   - 표시: 값을 왼쪽 고정폭에, 한글 라벨(축약어, 원어)을 오른쪽에 두는 세로 배치.
#     권장 터미널 크기 100x45 이상.
#   - 임계값 초과: 화면에서 노랑(주의)/빨강(위험)으로 칠하고,
#     동시에 경보 이력 파일($LOGF)에 시각과 함께 기록한다 (같은 항목은 30초에 1번만 기록).
#     → 화면을 안 보고 있던 사이의 스파이크도 이력으로 남는다.
#   - D state 경보 시에는 어떤 프로세스가 D 인지도 이력에 같이 남긴다.
#
# 각 수치의 의미 (출처와 해석 — 자세한 원리는 storage_io_vmware.md)
#   [CPU]  /proc/stat 의 1초 차분
#     us (user)    사용자 코드 실행 비율
#     sy (system)  커널 코드 실행 비율 (irq/softirq 포함)
#     wa (iowait)  CPU 가 놀면서 I/O 완료를 기다린 비율 — 힌트일 뿐, 판정은 디스크 await 로
#     st (steal)   하이퍼바이저가 vCPU 를 안 준 시간. ※ VMware 는 커널·ESXi 조합에 따라
#                  항상 0 으로 나올 수 있다 — 0 이어도 호스트 CPU 경합을 배제하지 못함(esxtop %RDY 필요)
#   [부하] /proc/loadavg — 실행 대기 + D state 태스크의 평균. vCPU 수 대비 %로 환산해 판단
#   [프로세스] /proc/stat 의 procs_running / procs_blocked
#     D (uninterruptible)  디스크/NFS 응답을 기다리며 시그널도 못 받는 프로세스 수.
#                          순간적인 1~2 는 정상, 지속·증가는 스토리지 지연의 직접 증거
#   [메모리] /proc/meminfo
#     avail (MemAvailable)  실제로 쓸 수 있는 메모리(캐시 회수분 포함). 낮을수록 위험
#     dirty                 아직 디스크로 안 내려간 쓰기. 급증 후 일괄 쓰기는 writeback 의 정상 동작
#   [스왑] /proc/vmstat 의 pswpin/pswpout 차분 (KiB/s 환산)
#     si/so (swap in/out)   0 이 아니면 메모리 부족이 디스크 I/O 로 전이되는 중
#   [VMware] vmware-toolbox-cmd (있을 때만, 5초마다 갱신)
#     balloon  호스트가 풍선 드라이버로 회수해 간 메모리(MB). >0 = 호스트 메모리 압박
#     hvswap   하이퍼바이저가 게스트 몰래 스왑한 양(MB). >0 = 심각
#   [디스크] /proc/diskstats 의 1초 차분 (iostat -x 와 같은 원천)
#     r/s w/s (reads/writes per second)  초당 완료 요청 수
#     rMB/s wMB/s                        처리량
#     await (average wait)  요청 1건의 평균 완료 시간 ms (큐 대기 포함) — ★ 판단의 중심.
#                           게스트에서는 ESXi+데이터스토어+어레이 왕복이 전부 합산된 값
#     aqu (avg queue size)  평균 동시 진행 요청 수. aqu ≈ IOPS × await(초) 가 정상(리틀의 법칙)
#     util (utilization)    디바이스가 바빴던 시간 %. 가상 디스크는 병렬 처리라 과대평가될 수 있음
#   [네트워크] /proc/net/dev 차분. 오류·유실(err+drop)이 늘면 그 자체로 경보
#
# 임계값 기본 (통상 관례 기준 — 환경에 맞게 조정. SAN 이 빠르면 await 를 10/30 으로 내려도 됨)
#   wa 20/40%  st 5/15%  load1 vCPU의 100/200%  avail 15/7% 이하
#   swap I/O >0 주의, 2MiB/s 위험  await 20/50ms  util 90/99%  aqu 8/32  D-state 1/5
#
# 참고: 물리 호스트(예: ProLiant DL380 Gen10)의 하드웨어 상태는 게스트에서 보이지 않는다.
#       그쪽은 iLO / ESXi(esxtop·vCenter) 영역이다.

# ── 옵션 ──────────────────────────────────────────────────────────────────────
case "${1:-}" in
    -h|--help) awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"; exit 0 ;;
esac

INTERVAL="${1:-${INTERVAL:-1}}"
LOGF="${LOGF:-/tmp/vm_os_watch_alerts.log}"
COOLDOWN="${COOLDOWN:-30}"          # 같은 경보의 이력 기록 간격(초)
DISKS="${DISKS:-}"                  # 빈 값이면 sd*/vd*/xvd*/nvme* 자동 감지

W_CPU_WA=${W_CPU_WA:-20}  C_CPU_WA=${C_CPU_WA:-40}
W_CPU_ST=${W_CPU_ST:-5}   C_CPU_ST=${C_CPU_ST:-15}
W_LOAD=${W_LOAD:-100}     C_LOAD=${C_LOAD:-200}      # load1 / vCPU %
W_MEMA=${W_MEMA:-15}      C_MEMA=${C_MEMA:-7}        # MemAvailable % (이하일 때)
C_SWAP=${C_SWAP:-2048}                               # si+so KiB/s (초과 시 위험)
W_AWAIT=${W_AWAIT:-20}    C_AWAIT=${C_AWAIT:-50}     # ms
W_UTIL=${W_UTIL:-90}      C_UTIL=${C_UTIL:-99}       # %
W_AQU=${W_AQU:-8}         C_AQU=${C_AQU:-32}
W_D=${W_D:-1}             C_D=${C_D:-5}              # D state 프로세스 수

# ── 준비 ──────────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
    YEL=$'\033[33m'; RED=$'\033[31;1m'; DIM=$'\033[2m'; RST=$'\033[0m'
else
    YEL=''; RED=''; DIM=''; RST=''
fi

TMPD=$(mktemp -d /tmp/vm_os_watch.XXXXXX) || exit 1
PREV="$TMPD/prev"; CUR="$TMPD/cur"
trap 'rm -rf "$TMPD"; printf "\033[?25h\033[0m\n"' EXIT
trap 'exit 0' INT TERM

NCPU=$(nproc)
VIRT=$(systemd-detect-virt 2>/dev/null || echo '?')
HAVE_VMT=0; command -v vmware-toolbox-cmd >/dev/null 2>&1 && HAVE_VMT=1
BALLOON_MB=-1; HVSWAP_MB=-1

refresh_vmt() {
    [ "$HAVE_VMT" = 1 ] || return 0
    BALLOON_MB=$(vmware-toolbox-cmd stat balloon 2>/dev/null | awk '{print $1+0}')
    HVSWAP_MB=$(vmware-toolbox-cmd stat swap 2>/dev/null | awk '{print $1+0}')
    [ -n "$BALLOON_MB" ] || BALLOON_MB=-1
    [ -n "$HVSWAP_MB" ]  || HVSWAP_MB=-1
}

snap() {
    {
        printf '#T %s\n' "$(date +%s%3N)"
        printf '#STAT\n'; cat /proc/stat
        printf '#DISK\n'; cat /proc/diskstats
        printf '#NET\n';  cat /proc/net/dev
        printf '#VM\n';   grep -E '^pswp(in|out) ' /proc/vmstat
        printf '#MEM\n';  cat /proc/meminfo
        printf '#LOAD\n'; cat /proc/loadavg
    } > "$1"
}

# ── 계산·표시 본체 (prev/cur 스냅샷의 차분 → 화면 행 + @A@ 경보 행) ────────────
# 표시 규칙: 값은 왼쪽 10칸 고정폭(한글 라벨은 폭이 2칸이라 뒤에 둬야 정렬이 깨지지 않음)
AWKPROG='
function pc(v,w,c){ if(c>0 && v>=c) return 2; if(w>0 && v>=w) return 1; return 0 }
function pl(v,w,c){ if(v<=c) return 2; if(v<=w) return 1; return 0 }
function P(s,l){ if(l==2) return RED s RST; if(l==1) return YEL s RST; return s }
function A(sev,key,msg){ printf("@A@|%d|%s|%s\n", sev, key, msg) }
function R(val,label){ printf("  %s   %s\n", val, label) }
{ ix = (FILENAME==ARGV[1]) ? 0 : 1 }
$1=="#T"    { T[ix]=$2; next }
$1=="#STAT" { sec=1; next }
$1=="#DISK" { sec=2; next }
$1=="#NET"  { sec=3; next }
$1=="#VM"   { sec=4; next }
$1=="#MEM"  { sec=5; next }
$1=="#LOAD" { sec=6; next }
sec==1 && $1=="cpu"           { for(i=2;i<=NF&&i<=11;i++) C[ix,i]=$i; next }
sec==1 && $1=="procs_running" { PR[ix]=$2; next }
sec==1 && $1=="procs_blocked" { PB[ix]=$2; next }
sec==2 {
    name=$3
    if (DISKS!="") ok = index(" " DISKS " ", " " name " ")>0
    else           ok = (name ~ /^(sd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme[0-9]+n[0-9]+)$/)
    if (ok) {
        if (ix==1 && !(name in SEEN)) { ORD[++ND]=name; SEEN[name]=1 }
        D[ix,name,1]=$4;  D[ix,name,2]=$6;  D[ix,name,3]=$7
        D[ix,name,4]=$8;  D[ix,name,5]=$10; D[ix,name,6]=$11
        D[ix,name,7]=$13; D[ix,name,8]=$14
    }
    next
}
sec==3 && /:/ && $0 !~ /\|/ {
    line=$0; sub(/:/," ",line); split(line,f," ")
    if (f[1]!="lo") {
        if (ix==1 && !(f[1] in NSEEN)) { NORD[++NN]=f[1]; NSEEN[f[1]]=1 }
        NV[ix,f[1],1]=f[2]; NV[ix,f[1],2]=f[4]+f[5]; NV[ix,f[1],3]=f[10]; NV[ix,f[1],4]=f[12]+f[13]
    }
    next
}
sec==4 { VM[ix,$1]=$2; next }
sec==5 && ix==1 { MEM[$1]=$2; next }
sec==6 && ix==1 { L1=$1; L5=$2; L15=$3; next }
END{
    dt=(T[1]-T[0])/1000.0; if(dt<=0) dt=1
    # CPU: /proc/stat cpu 행 = user nice system idle iowait irq softirq steal
    tt=0; for(i=2;i<=9;i++){ dc[i]=C[1,i]-C[0,i]; tt+=dc[i] }
    if(tt<=0) tt=1
    us=100*(dc[2]+dc[3])/tt; sy=100*(dc[4]+dc[7]+dc[8])/tt
    wa=100*dc[6]/tt; st=100*dc[9]/tt; id=100*dc[5]/tt
    lwa=pc(wa,W_CPU_WA,C_CPU_WA); lst=pc(st,W_CPU_ST,C_CPU_ST)

    printf("[ CPU ]  /proc/stat 1초 차분\n")
    R(sprintf("%9.1f%%",us),         "사용자 코드 실행 (us, user)")
    R(sprintf("%9.1f%%",sy),         "커널 코드 실행 (sy, system+irq+softirq)")
    R(P(sprintf("%9.1f%%",wa),lwa),  "I/O 완료 대기 (wa, iowait) — 힌트일 뿐, 판정은 디스크 평균응답으로")
    R(P(sprintf("%9.1f%%",st),lst),  "하이퍼바이저에 빼앗긴 시간 (st, steal) — VMware 는 0 고정일 수 있음")
    R(sprintf("%9.1f%%",id),         "유휴 (id, idle)")
    if(lwa) A(lwa,"cpu_wa", sprintf("CPU I/O 대기(wa, iowait) %.1f%% — 디스크 평균응답 확인 필요", wa))
    if(lst) A(lst,"cpu_st", sprintf("CPU 하이퍼바이저 경합(st, steal) %.1f%%", st))

    lp=int(100*L1/NCPU+0.5); ll=pc(lp,W_LOAD,C_LOAD)
    b=PB[1]+0; lb=pc(b,W_D,C_D)
    printf("\n[ 부하 · 프로세스 ]  /proc/loadavg · /proc/stat\n")
    R(sprintf("%10.2f",L1), sprintf("평균 부하 1분 (load average 1m) — vCPU %d 개의 %s", NCPU, P(sprintf("%d%%",lp),ll)))
    R(sprintf("%10s",sprintf("%.2f/%.2f",L5,L15)), "평균 부하 5분/15분 (load average 5m/15m)")
    R(sprintf("%10d",PR[1]+0),       "실행 대기 프로세스 (r, procs_running)")
    R(P(sprintf("%10d",b),lb),       "디스크 응답 대기 고착 프로세스 (D, uninterruptible sleep) — 지속 시 스토리지 의심")
    if(ll) A(ll,"load",   sprintf("1분 부하(load1) %.2f = vCPU %d 개의 %d%%", L1, NCPU, lp))
    if(lb) A(lb,"proc_d", sprintf("디스크 응답 대기(D state) 프로세스 %d개", b))

    mt=MEM["MemTotal:"]; ma=MEM["MemAvailable:"]
    ca=MEM["Cached:"]+MEM["Buffers:"]; drt=MEM["Dirty:"]
    ap=int(100*ma/mt+0.5); lm=pl(ap,W_MEMA,C_MEMA)
    su=MEM["SwapTotal:"]-MEM["SwapFree:"]
    si=int((VM[1,"pswpin"]-VM[0,"pswpin"])*4/dt+0.5)
    so=int((VM[1,"pswpout"]-VM[0,"pswpout"])*4/dt+0.5)
    ls=0; if(si+so>0) ls=(si+so>=C_SWAP)?2:1

    printf("\n[ 메모리 · 스왑 ]  /proc/meminfo · /proc/vmstat\n")
    R(P(sprintf("%9.1fG",ma/1048576),lm), sprintf("실제 가용 메모리 (avail, MemAvailable) — 전체 %.1fG 의 %s", mt/1048576, P(sprintf("%d%%",ap),lm)))
    R(sprintf("%9.1fG",ca/1048576),  "페이지 캐시 (cached+buffers) — 줄어들면 읽기가 디스크로 떨어짐")
    R(sprintf("%9dM",int(drt/1024)), "디스크 미기록 쓰기 (dirty) — writeback 대기 중인 양")
    R(sprintf("%9dM",int(su/1024)),  "스왑 사용량 (swap used)")
    R(P(sprintf("%7dK/s",si),ls),    "스왑 읽어들임 (si, swap-in) — 디스크에서 메모리로")
    R(P(sprintf("%7dK/s",so),ls),    "스왑 내려씀 (so, swap-out) — 0 이 아니면 메모리 부족이 디스크 I/O 로 전이")
    if(BALLOON<0 && HVSWAP<0){
        printf("  %s\n", DIM "(open-vm-tools 미설치 — balloon·hvswap 확인 불가)" RST)
    } else {
        R(P(sprintf("%9dM",BALLOON), BALLOON>0?1:0), "VMware 풍선 회수 (balloon, vmmemctl) — >0 이면 호스트가 메모리 회수 중")
        R(P(sprintf("%9dM",HVSWAP),  HVSWAP>0?2:0),  "VMware 하이퍼바이저 스왑 (hvswap, host-level swap) — >0 이면 심각")
    }
    if(lm)        A(lm,"mem_avail",  sprintf("가용 메모리(avail) %d%% (%.1fG)", ap, ma/1048576))
    if(ls)        A(ls,"swap_io",    sprintf("스왑 I/O 발생 si %dK/s · so %dK/s — 메모리 부족이 디스크로 전이", si, so))
    if(BALLOON>0) A(1,"vmw_balloon", sprintf("VMware 풍선 회수(balloon) %dMB — 호스트가 메모리 회수 중", BALLOON))
    if(HVSWAP>0)  A(2,"vmw_hvswap",  sprintf("VMware 하이퍼바이저 스왑(hvswap) %dMB — 호스트 메모리 심각", HVSWAP))

    printf("\n[ 디스크 ]  /proc/diskstats 1초 차분\n")
    printf("  %s\n", DIM "초당요청(r/s·w/s, requests/sec) · 처리량(rMB/s·wMB/s) · 평균응답(await, average wait) · 큐길이(aqu, avg queue size) · 사용률(util, utilization)" RST)
    printf("  장치        r/s     w/s    rMB/s    wMB/s      await     aqu    util\n")
    for(i=1;i<=ND;i++){
        name=ORD[i]
        if(!((0,name,1) in D)) continue
        rr=D[1,name,1]-D[0,name,1]; ww=D[1,name,4]-D[0,name,4]
        rmb=(D[1,name,2]-D[0,name,2])*512/1048576/dt
        wmb=(D[1,name,5]-D[0,name,5])*512/1048576/dt
        ios=rr+ww
        awt=(ios>0)?((D[1,name,3]-D[0,name,3])+(D[1,name,6]-D[0,name,6]))/ios:0
        ut=100*(D[1,name,7]-D[0,name,7])/(dt*1000); if(ut>100)ut=100; if(ut<0)ut=0
        aq=(D[1,name,8]-D[0,name,8])/(dt*1000); if(aq<0)aq=0
        la=pc(awt,W_AWAIT,C_AWAIT); lu=pc(ut,W_UTIL,C_UTIL); lq=pc(aq,W_AQU,C_AQU)
        printf("  %-9s %5d   %5d   %6.1f   %6.1f  %s  %s   %s\n",
            name, int(rr/dt+0.5), int(ww/dt+0.5), rmb, wmb,
            P(sprintf("%7.1fms",awt),la), P(sprintf("%6.1f",aq),lq), P(sprintf("%4d%%",int(ut+0.5)),lu))
        if(la)          A(la,"disk_" name "_await", sprintf("디스크 %s 평균응답(await) %.1fms (주의 %s / 위험 %s)", name, awt, W_AWAIT, C_AWAIT))
        if(lq>=2)       A(lq,"disk_" name "_aqu",   sprintf("디스크 %s 큐길이(aqu) %.1f", name, aq))
        if(lu>=2&&la==0)A(1, "disk_" name "_util",  sprintf("디스크 %s 사용률(util) %d%% 포화 근접", name, int(ut)))
    }

    printf("\n[ 네트워크 ]  /proc/net/dev 1초 차분\n")
    for(i=1;i<=NN;i++){
        nif=NORD[i]
        if(!((0,nif,1) in NV)) continue
        rb=(NV[1,nif,1]-NV[0,nif,1])*8/1000000/dt
        tb=(NV[1,nif,3]-NV[0,nif,3])*8/1000000/dt
        er=(NV[1,nif,2]-NV[0,nif,2])+(NV[1,nif,4]-NV[0,nif,4])
        printf("  %-10s  수신 (rx) %8.2f Mb/s   송신 (tx) %8.2f Mb/s   오류·유실 (err+drop) %s\n",
            nif, rb, tb, P(sprintf("%d",er), er>0?1:0))
        if(er>0) A(1,"net_" nif, sprintf("네트워크 %s 오류·유실(err+drop) +%d", nif, er))
    }
}'

# ── 메인 루프 ─────────────────────────────────────────────────────────────────
declare -A LAST
printf '\033[2J\033[?25l'
refresh_vmt
snap "$PREV"
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

    BODY=$(grep -v '^@A@' <<<"$OUT_RAW")

    # 경보 → 이력 기록 (같은 key 는 COOLDOWN 초에 1번만, D state 는 대상 프로세스도 기록)
    while IFS='|' read -r _tag sev key msg; do
        [ -n "$key" ] || continue
        last=${LAST[$key]:--999}
        if (( SECONDS - last >= COOLDOWN )); then
            LAST[$key]=$SECONDS
            lab=WARN; [ "$sev" = 2 ] && lab=CRIT
            echo "$(date '+%m-%d %T') [$lab] $msg" >> "$LOGF"
            if [ "$key" = proc_d ]; then
                ps -eo state,pid,wchan:20,comm --no-headers 2>/dev/null \
                    | awk '$1~/^D/{print "                 └ " $0}' >> "$LOGF"
            fi
        fi
    done < <(grep '^@A@' <<<"$OUT_RAW" || true)

    read -r _up _idle < /proc/uptime
    UPS=${_up%.*}; UPSTR="$((UPS/86400))d $((UPS%86400/3600))h $((UPS%3600/60))m"
    TITLE="── VM OS WATCH ── $(hostname) [${VIRT} guest, vCPU ${NCPU}] ── $(date '+%F %T') ── 가동 ${UPSTR} ── 갱신 ${INTERVAL}s ──"
    TAIL=$(tail -n 5 "$LOGF" 2>/dev/null)

    FRAME="${TITLE}

${BODY}

[ 최근 경보 ]  임계 초과 시각 기록 — 전체 이력: ${LOGF}
${TAIL:-  (경보 없음)}

${DIM}[ 읽는 법 ]
  · 색: 노랑 = 주의, 빨강 = 위험 / 같은 경보는 ${COOLDOWN}초에 1번만 이력에 기록
  · 평균응답 (await, average wait) — 요청 1건 완료까지의 ms. 디스크 판정의 중심. 게스트 값엔 ESXi·어레이 왕복까지 합산
  · 큐길이 (aqu, average queue size) — 동시 진행 요청 수 ≈ IOPS x 응답(초). 계속 늘면 디스크가 못 따라가는 중
  · 사용률 (util, utilization) — 바빴던 시간 비율. 가상 디스크는 100% 여도 여유가 있을 수 있는 보조 지표
  · I/O 대기 (wa, iowait) 와 부하 (load) 는 힌트 — 디스크 문제 판정은 평균응답(await) 으로
  · 임계 주의/위험: wa ${W_CPU_WA}/${C_CPU_WA}% · await ${W_AWAIT}/${C_AWAIT}ms · 가용 ${W_MEMA}/${C_MEMA}% · D ${W_D}/${C_D} — W_*·C_* 환경변수로 조정${RST}"

    printf '\033[H%s\033[0J' "$(printf '%s' "$FRAME" | sed $'s/$/\033[K/')"

    mv "$CUR" "$PREV"
    (( LOOP++ ))
done
