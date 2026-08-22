#!/bin/sh
# L4 를 설정합니다.
#   $1 대상 netns (lb1 | lb2)
#   $2 스케줄러  (rr | sh | mh | wrr)
#   $3 스케줄러 플래그 ("-" 면 없음, 예: mh-port)
#   $4 백엔드 목록 (10.0.2.x 의 마지막 옥텟, 기본 "11 12 13")
set -e
ns=$1; sched=${2:-rr}; flags=$3; reals=${4:-11 12 13}
[ "$flags" = "-" ] && flags=""

ip netns exec "$ns" ipvsadm -C
if [ -n "$flags" ]; then
  ip netns exec "$ns" ipvsadm -A -t 10.0.1.100:8080 -s "$sched" -b "$flags"
else
  ip netns exec "$ns" ipvsadm -A -t 10.0.1.100:8080 -s "$sched"
fi
# -m 은 masquerading, 즉 NAT 모드입니다. 목적지 주소만 고쳐 쓰고 지나갑니다.
for r in $reals; do
  ip netns exec "$ns" ipvsadm -a -t 10.0.1.100:8080 -r "10.0.2.$r:8080" -m
done
