#!/bin/sh
# 백엔드(=DB 한 대)가 죽었을 때, 이미 맺어진 연결을 L4 가 살릴 수 있는가.
#   $1 장애 종류 (backend | backend-proc)
# 커넥션 풀이 들고 있던 연결에 무슨 일이 생기는지를 그대로 재현합니다.
set -e
kind=${1:-backend}
LAB=$(dirname "$0")/..
G=/work/.gates

sh "$LAB/net/up.sh" rr - 0 0
mkdir -p "$G"

ip netns exec cli /work/bin/l4lab probe --to 10.0.1.100:8080 --stages 4 --new-at 4 --timeout 6 --gates "$G" \
  > "$G/probe.out" 2>&1 &
probe=$!
wait_gate() {
  tries=0
  while [ ! -f "$G/ready.$1" ] && [ $tries -lt 400 ]; do sleep 0.05; tries=$((tries + 1)); done
}

wait_gate 1
n=$(cut -d' ' -f1 "$G/backends" | tr -d 'B')
case "$kind" in
  backend)      note="B$n 서버를 통째로 끊는다 (정전·네트워크 단절)" ;;
  backend-proc) note="B$n 의 프로세스만 죽인다 (DB 프로세스 crash)" ;;
esac
echo "  (이 연결이 붙은 백엔드: B$n = 10.0.2.1$n)"
echo "  (장애: $note)"
sh "$LAB/net/fail.sh" "$kind" "$n"
sleep 0.5; touch "$G/go.1"

wait_gate 2
echo "  (헬스체크가 죽은 백엔드를 풀에서 빼는 것과 동등한 조작)"
sh "$LAB/net/fail.sh" evict "$n"
touch "$G/go.2"

wait_gate 3
echo "  (같은 시점에 새 연결을 맺어 본다)"
touch "$G/go.3"

wait $probe || true
cat "$G/probe.out"
