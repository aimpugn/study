#!/bin/sh

echo "Run kafka"
/__cacert_entrypoint.sh /etc/kafka/docker/run &

KAFKA_PID=$!

# Kafka가 시작될 때까지 대기합니다.
echo "Waiting Kafka to be ready..."
sleep 5

# 3개의 파티션 생성
# 3개의 브로커에 복제 설정
echo "Create app-logs topic"
/opt/kafka/bin/kafka-topics.sh --create \
  --if-not-exists \
  --topic app-logs \
  --bootstrap-server kafka1:9092 \
  --partitions 3 \
  --replication-factor 3

# 현재 실행 중인 백그라운드 프로세스가 종료될 때까지 컨테이너를 유지
# - https://www.man7.org/linux/man-pages/man1/wait.1p.html
echo "Wait kafka(pid: $KAFKA_PID) in working"
wait $KAFKA_PID
echo "$KAFKA_PID was terminated by a SIG$(kill -l $?) signal."
