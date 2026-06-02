#!/bin/bash
# vm_io_health.sh — VMware 게스트(Linux) 스토리지 I/O 1차 점검
# 사용법: sudo bash vm_io_health.sh [점검할 마운트포인트(기본 /)]
# 출력: 화면 + /tmp/io-health-<일시>.log
# 해석 기준은 같은 디렉터리의 storage_io_vmware.md 참고

TARGET="${1:-/}"
exec > >(tee "/tmp/io-health-$(date +%F-%H%M%S).log") 2>&1

sec() { printf '\n========== %s ==========\n' "$1"; }

sec "0. 기본 정보 (커널·디스크·큐 설정)"
uname -r
lsblk -o NAME,SIZE,TYPE,MOUNTPOINTS 2>/dev/null || lsblk
for d in /sys/block/sd* /sys/block/nvme*; do
    [ -e "$d" ] || continue
    printf '%s: scheduler=%s queue_depth=%s nr_requests=%s\n' \
        "${d##*/}" \
        "$(cat "$d/queue/scheduler" 2>/dev/null)" \
        "$(cat "$d/device/queue_depth" 2>/dev/null)" \
        "$(cat "$d/queue/nr_requests" 2>/dev/null)"
done

sec "1. 개요 — vmstat (b·si/so·wa·st)"
vmstat 1 5

sec "2. 디바이스별 지연·큐·사용률 — iostat"
iostat -xz 1 3 2>/dev/null || { echo "sysstat 미설치 (dnf/apt install sysstat). 원시값:"; cat /proc/diskstats; }

sec "3. 프로세스별 I/O — pidstat"
pidstat -d 1 3 2>/dev/null || echo "sysstat 미설치"

sec "4. D state 프로세스"
ps -eo state,pid,wchan:24,cmd | awk 'NR==1 || $1 ~ /^D/'

sec "5. 커널 로그의 디스크 에러"
dmesg -T 2>/dev/null | grep -iE "i/o error|timeout|task abort|reset|blocked for more than" | tail -20

sec "6. 용량·inode"
df -h; echo; df -i

sec "7. VMware 메모리 압박 — balloon·swap"
if command -v vmware-toolbox-cmd >/dev/null 2>&1; then
    echo "balloon: $(vmware-toolbox-cmd stat balloon)"
    echo "swap   : $(vmware-toolbox-cmd stat swap)"
else
    echo "open-vm-tools 미설치"
fi

sec "8. 실측 지연 — ioping ($TARGET)"
if command -v ioping >/dev/null 2>&1; then
    ioping -D -c 5 "$TARGET" || ioping -c 5 "$TARGET"
else
    echo "ioping 미설치 (dnf/apt install ioping)"
fi
