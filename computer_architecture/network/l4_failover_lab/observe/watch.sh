#!/bin/sh
# 두 L4 의 상태를 1초마다 갱신해 보여 줍니다. 장애 전환 순간을 눈으로 보기 위한 창.
while true; do
  printf '\033[H\033[2J'
  echo "=== $(date '+%H:%M:%S')  (Ctrl-C 로 종료) ==="
  for ns in lb1 lb2; do
    # down 된 인터페이스도 주소는 유지합니다. UP 인 것만 봐야 실제 보유자가 가려집니다.
    if ip netns exec "$ns" ip -br addr show up 2>/dev/null | grep -q 10.0.1.100
      then role="◆ VIP 보유"; else role="  대기/사망"; fi
    echo
    echo "───── $ns  $role ─────"
    table=$(ip netns exec "$ns" ipvsadm -Lnc 2>/dev/null | tail -n +2 | grep -v TIME_WAIT)
    if [ -z "$table" ]; then echo "  (연결 테이블 비어 있음)"; else echo "$table" | head -12; fi
  done
  sleep 1
done
