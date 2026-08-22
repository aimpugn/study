#!/bin/sh
# TCP 종단이 누구인지를 두 모드에서 같은 방법으로 관측합니다.
# 백엔드가 찍는 "내가 본 상대" 한 줄이 패킷 모드와 프록시 모드를 가릅니다.
set -e
LAB=$(dirname "$0")/..
killall l4lab 2>/dev/null || true
sleep 0.3
rm -f /work/.gates/who.out
sh "$LAB/net/build.sh"
mkdir -p /work/.gates

ip netns exec b1 /work/bin/l4lab serve --name B1 --port 9099 > /work/.gates/who.out 2>&1 &
sleep 0.5
# 패킷 모드: IPVS 가 목적지 주소만 고쳐 쓰고 지나갑니다
ip netns exec lb1 ipvsadm -A -t 10.0.1.100:9090 -s rr
ip netns exec lb1 ipvsadm -a -t 10.0.1.100:9090 -r 10.0.2.11:9099 -m
# 프록시 모드: 프로그램이 연결 두 개의 종단이 됩니다
ip netns exec lb1 /work/bin/l4lab proxy --listen 10.0.1.100:9091 --to 10.0.2.11:9099 >/dev/null 2>&1 &
sleep 1

echo "  주소 배치:  클라이언트=10.0.1.10   VIP=10.0.1.100   LB 뒷단=10.0.2.1   백엔드=10.0.2.11"
echo
for spec in "패킷 모드 (IPVS NAT)|9090" "프록시 모드|9091"; do
  name=${spec%|*}; port=${spec#*|}
  echo "  ── $name ──"
  ip netns exec cli /work/bin/l4lab probe --to "10.0.1.100:$port" --stages 1 --gates /work/.gates \
    | sed -n 's/^연결: local=\(.*\)  peer(소켓이 기억하는 상대)=\(.*\)$/     클라이언트가 본 상대: \2/p'
  sleep 0.5
  echo "     $(tail -1 /work/.gates/who.out)"
  echo
done
