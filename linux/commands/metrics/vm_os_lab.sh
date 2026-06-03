#!/bin/bash
# vm_os_lab.sh — 모니터(vm_os_watch.sh)를 "교본"으로 만드는 부하 실험 세트
#
# 사용법
#   터미널 1:  bash vm_os_watch.sh        # 모니터를 띄워 두고
#   터미널 2:  bash vm_os_lab.sh          # 실험 목록 보기
#              bash vm_os_lab.sh 2        # 2번 실험 실행
#
# 학습 루프 — 모든 실험이 같은 형식이다
#   ① "예측"을 읽고, 모니터의 어떤 수치가 어느 방향으로 움직일지 스스로 답해 본다
#   ② 실행하면서 모니터를 본다
#   ③ "해설"과 자기 예측을 대조한다 — 틀린 지점이 공부할 지점이다
#
# 실험 목록
#   1  CPU 1코어 점유            — us·load 의 관계, load 를 vCPU 로 나눠 보는 이유
#   2  iowait 의 착시 (2단계)    — wa 가 왜 힌트일 뿐인지를 직접 목격 (핵심 실험)
#   3  buffered 쓰기와 writeback — 쓰기가 "미뤄지는" 것을 dirty → bo 로 관찰
#   4  동기(direct) 쓰기의 비용  — 소요시간 ≈ 건수 × await 라는 산수
#   5  페이지 캐시               — 같은 읽기가 1회차엔 보이고 2회차엔 숨는 이유
#   6  메모리 점유와 avail       — cached 가 커도 안전하지 않은 경우 (tmpfs)
#   7  정리                      — 실험이 만든 파일 삭제
#
# 서버 영향 요약 (기본값 LAB_MB=256·LAB_SEC=20 기준. root 불필요, 시스템 설정 변경 없음, 일반 파일만 사용)
#   1: vCPU 1개를 20초 점유. 디스크·메모리 영향 없음
#   2: (a) 4k direct 읽기 20초 — QD1 이라 대역폭은 작지만 실제 디스크 IOPS 발생
#      (b) 추가로 "전체 vCPU" 20초 점유 ← 가장 공격적. 운영 중이면 앱 지연 유발 가능
#   3: 256MB(LAB_MB) 디스크 쓰기 1회 + writeback + sync(시스템 전체 플러시 1회)
#   4: 4k direct 쓰기 5000건(약 20MB) — 소요시간 = 5000 × await. 느린 스토리지면 길어진다
#   5: 64MB 쓰기 + 64MB 읽기 1회 (가장 가벼움)
#   6: 가용 메모리의 30%(최대 2G)를 20초 점유 후 반환 — OOM 안전 마진 70% 유지
#   잔여물: $LABD/test.dat (LAB_MB 크기, 7번으로 삭제) 뿐.
#   중단(Ctrl-C): 부하 프로세스 즉시 종료 + 임시 파일(/dev/shm 포함) 자동 정리 (트랩)
#
# 주의
#   - 2~5번은 디스크에, 6번은 메모리에 "실제" 부하를 건다 — 그게 목적이다. 운영 서버라면 한가한 시간에,
#     VMware 공유 데이터스토어라면 그 시간만큼 이웃 VM 에게도 경합(noisy neighbor)을 만든다는 점을 인지하고.
#   - 쓰는 곳은 일반 파일뿐이다: $LABD (기본 /var/tmp/vm_os_lab), /dev/shm/vm_os_lab.mem
#   - 환경변수: LAB_MB(파일 크기 MB, 기본 256) LAB_SEC(단계 길이 초, 기본 20) LAB_AUTO=1(Enter 생략)
#     운영 장비에서 줄여서: LAB_SEC=10 LAB_MB=64 bash vm_os_lab.sh 2

set -u
LAB_MB=${LAB_MB:-256}
LAB_SEC=${LAB_SEC:-20}
LAB_AUTO=${LAB_AUTO:-0}
LABD=${LABD:-/var/tmp/vm_os_lab}

# 어떤 식으로 끝나든(Ctrl-C 포함) 일시 파일은 정리하고 부하 프로세스는 즉시 멈춘다.
# test.dat 만 재사용을 위해 남긴다 — 7번 실험으로 삭제.
cleanup_files() { rm -f "$LABD/buf.dat" "$LABD/sync.dat" "$LABD/cache.dat" /dev/shm/vm_os_lab.mem; }
trap cleanup_files EXIT
trap 'pkill -P $$ 2>/dev/null; exit 130' INT TERM

say()   { printf '%s\n' "$*"; }
hr()    { printf -- '──────────────────────────────────────────────────────────────\n'; }
pause() { [ "$LAB_AUTO" = 1 ] || read -r -p "   [Enter] 시작 (Ctrl-C 중단) "; }

need_space() { # $1 = 필요한 MB (여유는 2배 요구)
    local free
    free=$(df -Pm "$LABD" | awk 'NR==2{print $4}')
    if [ "$free" -lt $(( $1 * 2 )) ]; then
        say "여유 공간 부족: ${free}MB < $(( $1 * 2 ))MB — LAB_MB 를 줄이거나 LABD 를 바꿔라"
        exit 1
    fi
}

cpu_spin() { # $1=코어 수  $2=초 — 백그라운드로 띄운다 (timeout 이 알아서 끝냄)
    local i
    for (( i=0; i<$1; i++ )); do
        timeout "$2" bash -c 'while :; do :; done' &
    done
}

mk_testfile() { # direct 쓰기로 만들어 페이지 캐시에 남지 않게 한다 (2·5번 실험의 전제)
    mkdir -p "$LABD"
    [ -f "$LABD/test.dat" ] && return 0
    need_space "$LAB_MB"
    say "   테스트 파일 생성 중 (${LAB_MB}MB, direct — 이것도 모니터에 w/s 로 보인다)..."
    dd if=/dev/zero of="$LABD/test.dat" oflag=direct bs=1M count="$LAB_MB" 2>/dev/null
}

direct_read() { # $1=초 — 캐시를 우회해 디바이스에서 4k 단위로 계속 읽는다
    local end=$(( SECONDS + $1 ))
    while (( SECONDS < end )); do
        dd if="$LABD/test.dat" of=/dev/null iflag=direct bs=4k count=2048 2>/dev/null
    done
}

exp1() {
    hr; say "■ 실험 1 — CPU 1코어 점유 (${LAB_SEC}초)"
    say "예측: us 가 코어 1개만큼 상승(vCPU 4 면 +25%p), 부하 1분값은 즉시가 아니라 서서히 1.0 을 향해 감."
    say "      디스크 await 는 변화 없음 — CPU 와 디스크는 별개의 축이다."
    pause
    cpu_spin 1 "$LAB_SEC"; wait
    say "해설: us = 사용자 코드 실행 비율. load 는 1분 지수평균이라 천천히 따라오고, 끝나도 천천히 빠진다."
    say "      load 를 vCPU 수로 나눠 보는 이유: 4.0 은 vCPU 4 에선 만석, 8 에선 절반이다."
}

exp2() {
    mk_testfile
    hr; say "■ 실험 2 — iowait 의 착시 (2단계 × ${LAB_SEC}초) ★핵심"
    say "(a) direct 읽기 단독:    wa 상승, 디스크 r/s 수백, await 수 ms."
    say "(b) 같은 읽기 + 전 코어 CPU 점유: 예측해 보라 — wa 는 어떻게 되나? await 는?"
    pause
    say "--- (a) 읽기 단독 ---"
    direct_read "$LAB_SEC"
    say "--- (b) 같은 읽기 + CPU 스핀 ---"
    cpu_spin "$(nproc)" "$LAB_SEC"
    direct_read "$LAB_SEC"
    wait
    say "해설: (b)에서 r/s·await 는 (a)와 같은데 wa 만 0 근처로 떨어졌을 것이다."
    say "      iowait 의 정의가 'CPU 가 놀면서 I/O 를 기다린 시간'이라서다 — CPU 가 바쁘면 잡히지 않는다."
    say "      같은 디스크 부하가 wa 에선 보였다 안 보였다 한다 → wa 는 힌트, 판정은 await. 이 규칙의 실체다."
}

exp3() {
    mkdir -p "$LABD"; need_space "$LAB_MB"
    hr; say "■ 실험 3 — buffered 쓰기와 writeback"
    say "예측: dd 자체는 금방 끝난다(디스크보다 빠르게). 그동안 dirty 가 쌓이고,"
    say "      dd 가 끝난 '뒤에도' wMB/s 가 한동안 이어진다 — 커널이 뒤에서 내려쓰는 것."
    pause
    time dd if=/dev/zero of="$LABD/buf.dat" bs=1M count="$LAB_MB" 2>/dev/null
    say "   dd 종료. 지금부터 ${LAB_SEC}초 — 모니터에서 dirty 가 줄며 디스크 쓰기가 이어지는 것을 보라."
    sleep "$LAB_SEC"
    sync; rm -f "$LABD/buf.dat"
    say "해설: write() 는 페이지 캐시를 더럽히고 즉시 반환한다. 실제 쓰기는 writeback 스레드가 나중에."
    say "      '쓰기가 빠르다'는 앱의 체감과 '디스크가 바쁘다'는 지표가 시간차를 갖는 이유."
    say "      dirty 가 한도(vm.dirty_ratio)를 넘으면 그때는 쓰는 프로세스가 직접 막힌다."
}

exp4() {
    mkdir -p "$LABD"; need_space 64
    hr; say "■ 실험 4 — 동기(direct) 쓰기의 비용"
    say "예측: 실험 3과 반대로 dd 자체가 느리다. 진행 중에 w/s 와 await 가 실시간으로 보인다."
    say "      소요시간 ≈ 5000건 × await. 끝난 뒤 time 출력과 await 로 검산해 보라."
    pause
    time dd if=/dev/zero of="$LABD/sync.dat" oflag=direct bs=4k count=5000 2>/dev/null
    rm -f "$LABD/sync.dat"
    say "해설: 캐시를 우회하면 쓰기 1건 = 스토리지 스택 전체 왕복 1번이라 지연이 그대로 노출된다."
    say "      DB 커밋(fsync)이 디스크 지연에 민감한 이유가 이 산수다 — await 20ms 면 단일 스레드 초당 50커밋이 상한."
}

exp5() {
    mkdir -p "$LABD"; need_space 64
    hr; say "■ 실험 5 — 페이지 캐시: 읽기는 숨는다"
    say "예측: 같은 파일을 두 번 읽는다. 1회차엔 r/s·rMB/s 가 보이고, 2회차엔 디스크가 조용한데 더 빨리 끝난다."
    pause
    dd if=/dev/zero of="$LABD/cache.dat" oflag=direct bs=1M count=64 2>/dev/null   # direct 생성 → 아직 캐시에 없음
    say "--- 1회차 (cold — 디스크에서) ---"
    time cat "$LABD/cache.dat" >/dev/null
    say "--- 2회차 (warm — 캐시에서) ---"
    time cat "$LABD/cache.dat" >/dev/null
    rm -f "$LABD/cache.dat"
    say "해설: 2회차는 페이지 캐시에서 끝나 디스크 지표에 잡히지 않는다."
    say "      '앱은 느린데 디스크는 조용'이 가능한 이유이자, 메모리 압박으로 캐시가 줄면 갑자기 느려지는 이유."
}

exp6() {
    hr; say "■ 실험 6 — 메모리 점유와 avail 의 반응 (tmpfs)"
    local avail take
    avail=$(awk '/^MemAvailable:/{print int($2/1024)}' /proc/meminfo)
    take=$(( avail * 30 / 100 )); [ "$take" -gt 2048 ] && take=2048
    if [ "$take" -lt 64 ]; then say "가용 메모리가 너무 적다(${avail}MB) — 실험 생략"; return 0; fi
    say "현재 avail ${avail}MB — 그중 ${take}MB 를 tmpfs(/dev/shm)에 점유한다."
    say "예측: avail 하락. 함정 하나 — 페이지 캐시(cached)는 오히려 '증가'한다. 왜일까?"
    pause
    dd if=/dev/zero of=/dev/shm/vm_os_lab.mem bs=1M count="$take" 2>/dev/null
    say "   ${LAB_SEC}초 유지 — 모니터에서 avail 과 cached 의 방향을 보라."
    sleep "$LAB_SEC"
    rm -f /dev/shm/vm_os_lab.mem
    say "해설: tmpfs(Shmem)는 통계상 cached 에 포함되지만 파일처럼 버릴 수 없어(스왑으로만 밀려남)"
    say "      avail 은 같이 줄어든다. 'cached 크니까 안전'이 항상 참이 아닌 이유 — 판정 기준은 항상 avail."
}

exp7() {
    rm -rf "$LABD" /dev/shm/vm_os_lab.mem
    say "정리 완료: $LABD, /dev/shm/vm_os_lab.mem"
}

case "${1:-}" in
    1) exp1 ;;
    2) exp2 ;;
    3) exp3 ;;
    4) exp4 ;;
    5) exp5 ;;
    6) exp6 ;;
    7) exp7 ;;
    *)
        awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"
        ;;
esac
