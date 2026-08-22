#!/bin/sh
# 장애를 냅니다. 종류를 어휘로 고릅니다.
#
#   takeover        활성 lb1 사망, 대기 lb2 가 두 VIP 인계 (keepalived 가 하는 일)
#   vanish          활성 lb1 사망, 아무도 인계하지 않음
#   backend  <N>    백엔드 N 을 통째로 끊음 (정전·네트워크 단절)
#   backend-proc <N> 백엔드 N 의 프로세스만 죽임 (DB 프로세스 crash)
#   evict    <N>    백엔드 N 을 lb1 의 풀에서 제외 (헬스체크 탈락)
set -e
kind=${1:-takeover}; n=$2

kill_lb1() {
  ip netns exec lb1 ipvsadm --stop-daemon master 2>/dev/null || true
  ip netns exec lb1 ip link set f1 down
  ip netns exec lb1 ip link set k1 down
}

case "$kind" in
  takeover)
    kill_lb1
    ip netns exec lb2 ipvsadm --stop-daemon backup 2>/dev/null || true
    ip netns exec lb2 ip addr add 10.0.1.100/32 dev f2
    ip netns exec lb2 ip addr add 10.0.2.254/32 dev k2
    # "이제 이 VIP 의 MAC 은 나다" 를 앞단과 뒷단에 알립니다. 이걸 빠뜨리면
    # 클라이언트가 죽은 lb1 의 MAC 으로 계속 보내 전혀 다른 이유로 연결이 끊깁니다.
    ip netns exec lb2 arping -q -U -c 2 -I f2 10.0.1.100
    ip netns exec lb2 arping -q -U -c 2 -I k2 10.0.2.254
    ;;
  vanish)       kill_lb1 ;;
  backend)      ip netns exec "b$n" ip link set "kb$n" down ;;
  backend-proc)
    # 해당 백엔드 프로세스 하나만 죽입니다. busybox pkill 은 믿을 수 없어 PID 로 직접 찾습니다.
    pid=$(ps -o pid,args | awk -v pat="serve --name B$n " 'index($0, pat) { print $1; exit }')
    [ -n "$pid" ] && kill "$pid"
    ;;
  evict)        ip netns exec lb1 ipvsadm -d -t 10.0.1.100:8080 -r "10.0.2.1$n:8080" ;;
  *) echo "알 수 없는 장애 종류: $kind" >&2; exit 2 ;;
esac
