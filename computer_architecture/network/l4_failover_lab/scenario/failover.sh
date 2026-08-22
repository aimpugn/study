#!/bin/sh
# L4 장애 전환 실험 한 건.
#   $1 제목  $2 스케줄러  $3 플래그  $4 sync  $5 sloppy  $6 장애 종류 (takeover | vanish)
# 환경변수: FLOWS(1) SEED(1) TIMEOUT(8) SETTLE(3) REALS2(11 12 13)
set -e
title=$1; sched=${2:-rr}; flags=${3:--}; sync=${4:-0}; sloppy=${5:-0}; kind=${6:-takeover}
LAB=$(dirname "$0")/..
G=/work/.gates

sh "$LAB/net/up.sh" "$sched" "$flags" "$sync" "$sloppy" "${REALS2:-11 12 13}"
mkdir -p "$G"

echo "===================================================================="
echo "  $title"
echo "  스케줄러=$sched 플래그=${flags} sync=$sync sloppy_tcp=$sloppy 장애=$kind 흐름=${FLOWS:-1}"
echo "===================================================================="

ip netns exec cli /work/bin/l4lab probe --to 10.0.1.100:8080 \
  --flows "${FLOWS:-1}" --stages 2 --seed "${SEED:-1}" --timeout "${TIMEOUT:-8}" --gates "$G" \
  > "$G/probe.out" 2>&1 &
probe=$!
tries=0
while [ ! -f "$G/ready.1" ] && [ $tries -lt 200 ]; do sleep 0.05; tries=$((tries + 1)); done
[ -f "$G/ready.1" ] || { echo "측정 도구가 연결하지 못했습니다:"; cat "$G/probe.out"; exit 1; }

echo "(상태 전파 대기 ${SETTLE:-3}s — IPVS sync 는 비동기다)"
sleep "${SETTLE:-3}"
for ns in lb1 lb2; do
  echo "[장애 전] $ns 연결 테이블:"
  ip netns exec "$ns" ipvsadm -Lnc | tail -n +2 | grep -v TIME_WAIT | sed 's/^/    /'
done

echo ">>> 장애 주입: $kind <<<"
sh "$LAB/net/fail.sh" "$kind"

echo "[장애 후] lb2 연결 테이블:"
ip netns exec lb2 ipvsadm -Lnc | tail -n +2 | grep -v TIME_WAIT | sed 's/^/    /'

# 아무도 VIP 를 인계하지 않은 경우에는 클라이언트 소켓이 어떻게 죽어 가는지가 핵심입니다.
# ESTABLISHED 로 보이지만 재전송만 쌓입니다.
if [ "$kind" = "vanish" ]; then
  ( sleep 4
    echo "[4초 뒤] 클라이언트 소켓 상태:"
    ip netns exec cli ss -tin '( dport = :8080 )' | head -3 | sed 's/^/    /' ) &
fi

touch "$G/go.1"
wait $probe || true
echo "--- 기존 연결 판정 ---"
cat "$G/probe.out"
