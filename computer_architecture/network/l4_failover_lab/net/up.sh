#!/bin/sh
# 망을 세우고 백엔드를 띄웁니다. 장애는 내지 않습니다.
#   $1 스케줄러  $2 플래그  $3 sync(0|1)  $4 sloppy_tcp(0|1)  $5 lb2 의 백엔드 목록
set -e
sched=${1:-rr}; flags=${2:--}; sync=${3:-0}; sloppy=${4:-0}; reals2=${5:-11 12 13}
LAB=$(dirname "$0")/..

# busybox 의 pkill -x 는 0 을 돌려주면서도 죽이지 않습니다. killall 은 comm 을 정확히 봅니다.
killall l4lab 2>/dev/null || true
sleep 0.3
rm -rf /work/.gates
sh "$LAB/net/build.sh"
sh "$LAB/net/ipvs.sh" lb1 "$sched" "$flags"
sh "$LAB/net/ipvs.sh" lb2 "$sched" "$flags" "$reals2"
for ns in lb1 lb2; do ip netns exec "$ns" sysctl -q -w net.ipv4.vs.sloppy_tcp="$sloppy"; done

if [ "$sync" = "1" ]; then
  # 연결이 생기자마자 동기화하도록 문턱값을 낮춥니다. 그래도 전파는 비동기입니다.
  ip netns exec lb1 sh -c 'echo "0 1" > /proc/sys/net/ipv4/vs/sync_threshold'
  ip netns exec lb1 ipvsadm --start-daemon master --mcast-interface=f1 --syncid 7
  ip netns exec lb2 ipvsadm --start-daemon backup --mcast-interface=f2 --syncid 7
fi

for i in 1 2 3; do
  ip netns exec "b$i" /work/bin/l4lab serve --name "B$i" --port 8080 >/dev/null 2>&1 &
done
sleep 1
