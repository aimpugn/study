#!/bin/sh
# 지금 이 순간의 상태 전부를 라벨과 함께 찍습니다. 장애 전후를 비교하기 위한 것.
echo "################ ${1:-snapshot} ################"
echo "# 시각: $(date '+%F %T')"
for ns in lb1 lb2; do
  if ip netns exec "$ns" ip -br addr show up 2>/dev/null | grep -q 10.0.1.100
    then role="VIP 보유"; else role="대기 또는 사망"; fi
  echo
  echo "===== $ns ($role) ====="
  echo "-- 보유 주소 --";     ip netns exec "$ns" ip -br addr | grep -v '^lo'
  echo "-- 서비스 테이블 --"; ip netns exec "$ns" ipvsadm -Ln         | tail -n +2
  echo "-- 누적 통계 --";     ip netns exec "$ns" ipvsadm -Ln --stats | tail -n +2
  echo "-- 연결 테이블 --";   ip netns exec "$ns" ipvsadm -Lnc        | tail -n +2
done
echo
echo "===== 클라이언트 소켓 ====="
ip netns exec cli ss -tin '( dport = :8080 )'
