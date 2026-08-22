#!/bin/sh
# 망을 세웁니다.
#
#   cli ─┐                        ┌─ lb1 ─┐                  ┌─ b1
#        ├─ br-front (L2 스위치) ─┤       ├─ br-back (L2) ───┼─ b2
#        ┘                        └─ lb2 ─┘                  └─ b3
#
#   앞단 VIP  10.0.1.100  클라이언트가 접속하는 주소     ← 장애 시 lb1 → lb2 로 이사
#   뒷단 GW   10.0.2.254  백엔드의 기본 게이트웨이       ← 함께 이사
#
# 뒷단 게이트웨이도 같이 이사해야 합니다. IPVS NAT 모드에서는 백엔드의 응답이 반드시
# L4 를 거쳐 돌아와야 하므로, 기본 경로가 죽은 lb1 을 가리키면 연결 테이블을 아무리
# 잘 넘겨받아도 돌아오는 길이 끊깁니다.
set -e

# --- 정리. netns 삭제는 비동기이므로 실제로 사라질 때까지 기다립니다. ---
for ns in $(ip netns list 2>/dev/null | awk '{print $1}'); do ip netns del "$ns" 2>/dev/null || true; done
tries=0
while [ -n "$(ip netns list 2>/dev/null | awk '{print $1}')" ] && [ $tries -lt 50 ]; do
  sleep 0.2; tries=$((tries + 1))
done
ip link del br-front 2>/dev/null || true
ip link del br-back  2>/dev/null || true
# 브리지를 지워도 그 포트였던 veth 는 남습니다. 다음 실행과 이름이 충돌하므로 함께 지웁니다.
for link in $(ip -br link 2>/dev/null | awk -F'@' '/^p-/{print $1}'); do ip link del "$link" 2>/dev/null || true; done
sleep 0.2

for ns in cli lb1 lb2 b1 b2 b3; do
  ip netns add "$ns"
  ip netns exec "$ns" ip link set lo up
done
for br in br-front br-back; do ip link add "$br" type bridge; ip link set "$br" up; done

# attach <netns> <브리지> <인터페이스 이름> <주소/prefix>
attach() {
  ip link add "$3" type veth peer name "p-$3"
  ip link set "$3" netns "$1"
  ip link set "p-$3" master "$2"
  ip link set "p-$3" up
  ip netns exec "$1" ip link set "$3" up
  ip netns exec "$1" ip addr add "$4" dev "$3"
}

attach cli br-front fc  10.0.1.10/24
attach lb1 br-front f1  10.0.1.1/24
attach lb2 br-front f2  10.0.1.2/24
attach lb1 br-back  k1  10.0.2.1/24
attach lb2 br-back  k2  10.0.2.2/24
attach b1  br-back  kb1 10.0.2.11/24
attach b2  br-back  kb2 10.0.2.12/24
attach b3  br-back  kb3 10.0.2.13/24

for i in 1 2 3; do ip netns exec "b$i" ip route add default via 10.0.2.254; done
for i in 1 2; do
  ip netns exec "lb$i" sysctl -q -w net.ipv4.ip_forward=1
  # 이걸 켜지 않으면 IPVS 가 per-CPU 카운터를 집계하지 않아,
  # 트래픽이 흘러도 `ipvsadm -Ln --stats` 가 계속 0 으로 보입니다.
  ip netns exec "lb$i" sysctl -q -w net.ipv4.vs.run_estimation=1
done

# 처음에는 lb1 이 활성입니다. 두 VIP 를 모두 소유합니다.
ip netns exec lb1 ip addr add 10.0.1.100/32 dev f1
ip netns exec lb1 ip addr add 10.0.2.254/32 dev k1
