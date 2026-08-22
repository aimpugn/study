#!/usr/bin/env bash
# L4 장애 전환 실험실 진입점.
#
#   준비/정리   ./lab.sh up                      ./lab.sh down
#               ./lab.sh build                   측정 도구(Rust) 다시 컴파일
#   자동 실행   ./lab.sh all                     ./lab.sh 1 ~ ./lab.sh 7   (logs/ 에 기록)
#   수동 실습   ./lab.sh manual [스케줄러] [플래그] [sync] [sloppy]
#               ./lab.sh talk                    클라이언트로 직접 타이핑
#               ./lab.sh watch                   두 L4 의 연결 테이블 실시간 관찰
#               ./lab.sh fail [종류] [번호]      장애 주입 (기본 takeover)
#               ./lab.sh snap <라벨>             지금 상태 전체를 라벨 붙여 기록
#   기록        ./lab.sh logs                    남은 로그 목록
#   기타        ./lab.sh shell                   컨테이너 셸
#
# 장애 종류: takeover | vanish | backend N | backend-proc N | evict N
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
C=l4lab
RUNTIME=alpine:3.20
BUILDER=l4lab-builder
LOGDIR="$HERE/logs"
SESSION="$LOGDIR/session.log"
ENVS=()

need_container() {
  docker inspect "$C" >/dev/null 2>&1 || { echo "실험대가 없습니다. 먼저 ./lab.sh up 을 실행하세요."; exit 1; }
}

build() {
  mkdir -p "$HERE/bin"
  docker build -q -t "$BUILDER" -f "$HERE/Dockerfile.builder" "$HERE" >/dev/null
  # 문서 없는 함수가 하나라도 있으면 여기서 멈춥니다. 비공개 헬퍼까지 포함합니다.
  # 주석을 쓰는 것을 규율에 맡기지 않고 기계가 막게 하려는 것입니다.
  docker run --rm -v "$HERE":/w -w /w -e CARGO_TARGET_DIR=/w/target "$BUILDER" sh -c '
    cargo clippy --release -- -D missing_docs -D clippy::missing_docs_in_private_items &&
    cargo build --release &&
    cp target/release/l4lab bin/l4lab'
  echo "측정 도구 빌드 완료: bin/l4lab"
}

load_modules() {
  # ip_vs 는 호스트 커널 모듈입니다. 컨테이너에는 /lib/modules 가 없으므로
  # PID 1 의 마운트 네임스페이스로 들어가서 적재합니다.
  docker run --rm --privileged --pid=host "$RUNTIME" \
    nsenter -t 1 -m -u -i -n -p -- sh -c \
    'for m in ip_vs ip_vs_rr ip_vs_sh ip_vs_mh; do modprobe $m || exit 1; done' \
    || { echo "ip_vs 적재 실패: 이 커널은 IPVS 를 제공하지 않습니다"; exit 1; }
  echo "ip_vs 모듈 적재 완료"
}

up() {
  command -v docker >/dev/null || { echo "docker 가 필요합니다"; exit 1; }
  [ -x "$HERE/bin/l4lab" ] || build
  load_modules
  docker rm -f "$C" >/dev/null 2>&1 || true
  docker run -d --privileged --name "$C" -v "$HERE":/work "$RUNTIME" sleep infinity >/dev/null
  docker exec "$C" apk add -q --no-cache iproute2 ipvsadm iputils-arping tcpdump
  mkdir -p "$LOGDIR"
  echo "실험대 준비 완료 (컨테이너: $C, 로그: logs/)"
}

env_for() { ENVS=(); for e in "$@"; do ENVS+=(-e "$e"); done; }

# 실험 한 건. 화면에 보여 주면서 동시에 logs/ 에 남깁니다.
record() { # record <로그이름> <컨테이너에서 실행할 명령...>
  need_container; mkdir -p "$LOGDIR"
  local tag=$1; shift
  local file="$LOGDIR/$(date '+%Y%m%d-%H%M%S')_${tag}.log"
  docker exec ${ENVS[@]+"${ENVS[@]}"} "$C" timeout 150 "$@" 2>&1 | tee "$file"
  echo "  → 기록: ${file#"$HERE"/}"
}

case "${1:-all}" in
  build) build ;;
  up)    up ;;
  down)  docker rm -f "$C" >/dev/null 2>&1 || true; echo "정리 완료 (logs/ 와 bin/ 은 남습니다)" ;;

  1) env_for; record exp1 sh /work/scenario/failover.sh \
       "실험1 · 상태 유지형(rr) + 상태 인계 없음" rr - 0 0 takeover ;;
  2) env_for; record exp2 sh /work/scenario/failover.sh \
       "실험2 · 상태 유지형(rr) + sync daemon 상태 인계" rr - 1 0 takeover ;;
  3) env_for; record exp3 sh /work/scenario/failover.sh \
       "실험3 · 해시형(mh/Maglev) + 상태 인계 없음" mh mh-port 0 1 takeover ;;
  4) env_for TIMEOUT=15; record exp4 sh /work/scenario/failover.sh \
       "실험4 · 아무도 VIP 를 인계하지 않음" rr - 0 0 vanish ;;
  5) env_for FLOWS=16 SEED=1
     record exp5a sh /work/scenario/failover.sh "5a · 다중 흐름 · rr" rr - 0 1 takeover
     record exp5b sh /work/scenario/failover.sh "5b · 다중 흐름 · mh" mh mh-port 0 1 takeover
     env_for FLOWS=18 SEED=1 REALS2="11 12"
     record exp5c sh /work/scenario/failover.sh \
       "5c · 다중 흐름 · mh, 장애와 동시에 B3 이탈" mh mh-port 0 1 takeover ;;
  6) env_for
     record exp6 sh -c 'echo "실험6 · 백엔드가 죽으면 L4 가 기존 연결을 살릴 수 있는가"
       echo; echo "########## 사례 A · 프로세스만 죽음 ##########"
       sh /work/scenario/backend-death.sh backend-proc
       echo; echo "########## 사례 B · 서버가 통째로 사라짐 ##########"
       sh /work/scenario/backend-death.sh backend' ;;
  7) env_for; record exp7 sh -c 'echo "실험7 · TCP 종단은 누구인가"; echo
       sh /work/scenario/endpoints.sh' ;;
  all) docker inspect "$C" >/dev/null 2>&1 || up
       for n in 1 2 3 4 5 6 7; do "$0" "$n"; echo; done ;;

  manual)
    need_container
    docker exec "$C" sh /work/net/up.sh "${2:-rr}" "${3:--}" "${4:-0}" "${5:-0}"
    docker exec "$C" ip netns exec lb1 ipvsadm -Ln | tail -n +2
    cat <<'MSG'

터미널 세 개를 띄우고 순서대로 진행하세요.

  터미널 A   ./lab.sh talk       연결을 맺고 직접 타이핑합니다
  터미널 B   ./lab.sh watch      두 L4 의 연결 테이블을 지켜봅니다
  터미널 C   ./lab.sh snap 장애전  →  ./lab.sh fail  →  ./lab.sh snap 장애후

터미널 A 에서 아무 글자나 치면 [B1]/[B2]/[B3] 중 하나가 답합니다.
그 상태로 터미널 C 에서 fail 한 뒤, 다시 터미널 A 에서 같은 연결에 타이핑해 보세요.
MSG
    ;;
  talk)  need_container
         docker exec -it "$C" ip netns exec cli /work/bin/l4lab talk \
           --to 10.0.1.100:8080 --timeout "${2:-10}" ;;
  watch) need_container; docker exec -it "$C" sh /work/observe/watch.sh ;;
  fail)  need_container; docker exec "$C" sh /work/net/fail.sh "${2:-takeover}" "${3:-}"
         echo "장애 주입 완료: ${2:-takeover} ${3:-}" ;;
  snap)  need_container; mkdir -p "$LOGDIR"
         docker exec "$C" sh /work/observe/snapshot.sh "${2:-snapshot}" | tee -a "$SESSION"
         echo "  → 누적 기록: ${SESSION#"$HERE"/}" ;;

  logs)  [ -d "$LOGDIR" ] || { echo "아직 로그가 없습니다"; exit 0; }
         ls -lhtr "$LOGDIR" | tail -n +2 | awk '{printf "  %-28s %6s  %s %s %s\n", $9, $5, $6, $7, $8}' ;;
  shell) need_container; docker exec -it "$C" sh ;;

  *) sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'; exit 1 ;;
esac
