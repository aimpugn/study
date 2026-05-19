# Replication, Backup, Recovery

## replication log, binlog/WAL shipping, lag

복제는 primary에서 생긴 변경을 다른 서버가 따라가게 만드는 장치이지만, 읽기 일관성을 자동으로 보장하는 장치는 아니다. 이 절은 PostgreSQL의 WAL shipping/streaming replication과 MySQL의 binary log replication을 같은 시간축 위에 올려 놓고, commit이 성공한 순간과 replica에서 그 값이 보이는 순간 사이에 어떤 간격이 생기는지 설명한다. 로컬 seed인 `database/replication.md`와 `database/replication_lag.md`는 writer가 로그를 만들고 reader가 적용한다는 뼈대를 이미 갖고 있지만, 여기서는 그 뼈대를 운영자가 실제로 관측할 수 있는 LSN, binlog position, relay log, replay lag의 언어로 다시 세운다.

이 절에서 공식 근거로 삼은 자료는 다음과 같다. 링크는 단순 참고가 아니라 본문 판단의 경계다. PostgreSQL과 MySQL은 같은 단어를 쓰더라도 로그 이름, 보존 책임, failover 절차, 관측 뷰가 다르므로, 공통 원리와 제품별 실행 방법을 분리해 읽어야 한다.

- PostgreSQL 18 Log-Shipping Standby Servers: https://www.postgresql.org/docs/current/warm-standby.html
- MySQL 8.4 Replication: https://dev.mysql.com/doc/refman/8.4/en/replication.html
- 로컬 seed: database/replication.md, database/replication_lag.md

복제가 필요한 배경은 단순히 읽기 성능을 늘리는 데서 끝나지 않는다. 하나의 DB 서버가 모든 읽기와 쓰기, 백업, 분석, 장애 대기를 동시에 맡으면 작은 장애가 전체 서비스 장애가 된다. 그래서 primary는 쓰기 원장 역할을 유지하고, standby 또는 replica는 읽기 분산, 백업 분리, 장애 대기, 원격지 복사 같은 역할을 나누어 맡는다. 하지만 역할을 나누는 순간 '같은 데이터'라는 말은 더 이상 한 시각의 완전 동일성을 뜻하지 않는다. 네트워크와 로그 적용이라는 시간차가 들어오기 때문이다.

가장 작은 작동 단위는 row 자체가 아니라 변경 로그다. PostgreSQL은 데이터 파일을 직접 복사해 계속 맞추는 것이 아니라 WAL record를 만들고, standby가 그 record를 받아 replay한다. MySQL은 source가 binary log event를 남기고 replica가 relay log를 거쳐 적용한다. 용어는 다르지만 공통 구조는 `commit -> durable log -> transfer -> receive -> apply -> visible read`다. 이 흐름에서 lag는 하나의 물리 위치 차이이자 사용자가 보는 시간 차이다.

```text
T0 client:  UPDATE account SET balance = balance - 100 WHERE id = 10;
T1 primary: transaction commit, WAL/binlog durable
T2 network: WAL record or binlog event leaves primary/source
T3 replica: log received and flushed as WAL/relay log
T4 replica: replay/apply updates local data files or storage state
T5 reader:  SELECT balance FROM account WHERE id = 10 sees new value

lag window = T5 - T1
data-loss window in async shipping = records committed at primary but not yet safely shipped/applied
```

이 timeline에서 가장 흔한 함정은 T1과 T5를 같은 순간으로 생각하는 것이다. API 서버는 T1에서 commit 성공을 받았으므로 사용자에게 성공을 보여 줄 수 있다. 그러나 바로 이어지는 조회가 replica로 라우팅되면 그 조회는 T4가 끝난 뒤에야 새 값을 본다. 따라서 복제 구조를 읽기 분산에 쓰려면 도메인별 읽기 계약이 필요하다. 게시글 목록처럼 몇 초 늦어도 되는 읽기는 replica로 보내도 되지만, 결제 직후 잔액 확인이나 주문 상세처럼 자기 쓰기 읽기가 필요한 요청은 primary로 보내거나 replica가 특정 LSN/GTID 이상을 따라잡았는지 확인해야 한다.

| 관측 위치 | PostgreSQL에서 보는 값 | MySQL에서 보는 값 | 운영 해석 |
| --- | --- | --- | --- |
| primary/source log head | pg_current_wal_lsn() | SHOW BINARY LOG STATUS | 새 변경이 어디까지 생성되었는지 |
| sent position | pg_stat_replication.sent_lsn | source의 binlog dump thread 상태 | primary가 어디까지 보냈는지 |
| received position | pg_last_wal_receive_lsn(), pg_stat_wal_receiver.flushed_lsn | replica relay log 수신 위치 | 네트워크 또는 수신 지연 여부 |
| replayed/applied position | pg_last_wal_replay_lsn() | replica SQL applier position | 적용 병목 여부 |
| client-visible read | standby SELECT 결과 | replica SELECT 결과 | 사용자가 보는 최신성 |

PostgreSQL 공식 문서는 log shipping이 비동기이며 commit 뒤 WAL record가 보내지기 때문에 primary가 치명적으로 실패하면 아직 ship되지 않은 transaction을 잃을 수 있다고 설명한다. streaming replication은 WAL file이 가득 찰 때까지 기다리지 않고 record를 흘려 보내므로 파일 기반 shipping보다 창이 작지만, 기본 동작은 여전히 비동기다. MySQL 공식 문서도 replication은 기본적으로 asynchronous이고 replica가 source에 계속 연결되어 있을 필요는 없다고 말한다. 이 두 문장을 합치면 복제의 핵심 한계가 선명해진다. 복제는 availability와 scale-out을 주지만, 별도 설정과 정책 없이 모든 읽기를 최신으로 만들지는 않는다.

```sql
-- PostgreSQL lag를 위치 차이로 나누어 보는 예시
-- primary
SELECT pg_current_wal_lsn() AS primary_lsn;

SELECT application_name, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
       pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS bytes_not_replayed
FROM pg_stat_replication;

-- standby
SELECT pg_last_wal_receive_lsn() AS received,
       pg_last_wal_replay_lsn() AS replayed,
       now() - pg_last_xact_replay_timestamp() AS replay_time_lag;

-- MySQL replica 관측 예시
SHOW REPLICA STATUS\G
SHOW BINARY LOG STATUS;
```

위 쿼리는 'lag가 있다'를 네 구간으로 쪼갠다. primary_lsn과 sent_lsn 차이가 크면 primary가 WAL을 만들고도 보내는 데 밀리고 있을 수 있다. sent_lsn과 receive_lsn 차이가 크면 네트워크나 receiver 쪽 병목을 의심한다. receive_lsn과 replay_lsn 차이가 크면 standby가 받았지만 적용하지 못한 것이다. MySQL도 source의 binary log 위치, replica의 relay log 수신 위치, SQL applier 적용 위치를 나누어 봐야 같은 방식으로 원인을 좁힐 수 있다.

| 시간 | 사용자 요청 | primary/source | replica | 읽기 결과 |
| --- | --- | --- | --- | --- |
| 10:00:00.000 | 주문 생성 요청 | BEGIN | 아직 이전 상태 | - |
| 10:00:00.030 | 주문 row insert | WAL/binlog event 생성 준비 | 아직 이전 상태 | - |
| 10:00:00.050 | commit 성공 응답 | commit LSN 0/70001A0 durable | event 미수신 | primary는 주문 존재 |
| 10:00:00.060 | 주문 상세 조회가 replica로 감 | 새 event 전송 중 | 아직 replay 전 | 주문 없음 가능 |
| 10:00:00.180 | receiver 수신 | sent 완료 | relay/WAL receive 완료 | 아직 주문 없음 가능 |
| 10:00:00.240 | apply 완료 | 다음 event 생성 중 | replay/apply 완료 | 주문 보임 |

이 표의 목적은 replica read가 틀렸다고 몰아가는 것이 아니다. replica는 자신이 적용한 지점까지는 일관된 DB다. 다만 사용자가 기대하는 최신성과 replica가 보장하는 최신성이 다를 수 있다. 따라서 장애 대응 문서에는 'replica에서 주문이 안 보였다'라는 증상만 남기면 부족하고, 그 시각의 primary commit 위치, replica receive 위치, replay/apply 위치, 애플리케이션 read routing 결정을 함께 남겨야 한다.

복제 지연의 원인은 크게 생성량, 전송, 적용, 충돌로 나뉜다. 생성량 문제는 primary에서 WAL/binlog가 너무 많이 생기는 상황이다. 대량 update, index 생성, batch job, vacuum이나 purge와 겹친 쓰기 폭증이 여기에 들어간다. 전송 문제는 네트워크 대역폭, TLS 비용, cross-region latency, receiver process 장애로 나타난다. 적용 문제는 replica CPU, disk fsync, SQL applier 병렬성, long running query와 충돌한다. 충돌 문제는 hot standby read query가 replay와 부딪히거나, MySQL row event 적용 순서가 긴 transaction 하나에 막히는 식으로 드러난다.

복제 slot과 WAL 보존도 실무 함정이다. standby가 뒤처져도 primary가 필요한 WAL을 너무 일찍 재활용하지 않게 해 주는 장치는 안전망이다. 그러나 standby가 끊긴 상태로 오래 남으면 primary의 `pg_wal`이 계속 커질 수 있다. 공식 문서는 replication slot이 WAL segment를 많이 붙잡아 공간을 채울 수 있으므로 제한과 관측이 필요하다고 경고한다. 안전망은 관측 없이 켜 두면 다른 종류의 장애가 된다.

복제 설계에서 senior가 먼저 묻는 질문은 'replica 몇 대인가'가 아니라 '어떤 읽기가 얼마나 stale해도 되는가'다. 사용자 세션, 관리자 대시보드, 검색 인덱싱, 통계 집계, 백업, 장애 대기 노드는 서로 다른 최신성 요구를 가진다. 이 요구를 나누지 않고 모든 SELECT를 replica로 보내면, 장애가 아니라 정상 lag만으로도 사용자 신뢰를 잃을 수 있다.

```text
read routing contract example

write class       required read freshness        default read target       fallback
------------      -----------------------        ------------------       --------
order create      own write must be visible      primary                  wait for replica LSN then replica
product list      seconds of staleness allowed   nearest replica          primary only during incident
settlement close  no stale financial read        primary                  no replica fallback
analytics chart   minutes acceptable             analytics replica        show data_as_of timestamp
```

검증은 작은 실험으로도 가능하다. primary에서 row를 쓰고 commit LSN을 기록한 뒤, 즉시 standby에서 같은 row를 읽으며 `pg_last_wal_replay_lsn()`이 commit LSN을 넘기 전과 후를 비교한다. MySQL에서는 source의 binary log status와 replica status를 기록하고, 대량 update나 replica CPU 부하를 걸어 `Seconds_Behind_Source`와 relay/apply 위치 변화를 본다. PASS는 lag가 발생했을 때 위치 차이와 사용자 읽기 결과가 같은 방향으로 설명되는 것이다. FAIL은 lag 숫자는 있는데 어느 구간이 늦는지 설명하지 못하거나, replica read 정책이 도메인 최신성 요구와 연결되지 않는 것이다.

### 운영 리허설 카드: 복제 지연을 설명하고 줄이는 판단

1. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

2. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

3. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

4. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

5. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

6. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

7. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

8. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

9. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

10. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

11. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

12. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

13. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

14. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

15. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

16. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

17. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

18. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

19. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

20. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

21. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

22. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

23. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

24. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

25. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

26. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

27. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

28. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

29. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

30. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

31. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

32. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

33. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

34. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

35. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

36. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

37. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

38. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

39. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

40. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

41. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

42. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

43. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

44. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

45. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

46. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

47. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

48. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

49. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

50. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

51. 쓰기 API가 성공 응답을 돌려준 직후 같은 사용자의 읽기 요청을 replica로 보내면, 애플리케이션은 자기 쓰기를 읽지 못할 수 있다. 이것은 replica가 틀렸다는 뜻이 아니라 primary commit, 로그 전송, relay 또는 WAL 수신, replay 적용이라는 네 단계 중 마지막 단계가 아직 끝나지 않았다는 뜻이다. 그래서 사용자 세션의 read-after-write가 중요하면 일정 시간 primary로 읽거나, commit LSN 또는 GTID를 세션에 들고 다니며 replica가 그 지점 이상을 재생했는지 확인하는 정책이 필요하다.

52. 복제 지연을 한 숫자로만 보면 원인을 놓친다. primary가 로그를 아직 보내지 못한 지연, 네트워크가 relay를 늦게 전달한 지연, replica가 로그를 받았지만 디스크에 flush하지 못한 지연, flush는 했지만 SQL apply나 WAL replay가 뒤처진 지연은 서로 다른 조치가 필요하다. PostgreSQL의 pg_stat_replication 필드 차이와 standby의 receive/replay LSN 차이를 나누어 보는 이유가 여기에 있다.

53. MySQL 복제에서 source와 replica라는 이름은 책임을 분명히 한다. source는 binary log에 변경 이벤트를 남기고, replica는 그 이벤트를 relay log로 받아 적용한다. 따라서 replica에서 읽을 수 있다는 사실은 'source의 모든 commit을 이미 적용했다'는 보장이 아니다. 특히 장기 분석 쿼리, 느린 디스크, 단일 SQL applier 병목은 relay는 쌓이는데 apply가 늦어지는 전형적인 장애 모양을 만든다.

54. PostgreSQL streaming replication은 파일 하나가 꽉 찰 때까지 기다리지 않고 WAL record를 보낼 수 있어 파일 기반 log shipping보다 지연 창이 작다. 하지만 기본은 비동기이므로 primary commit과 standby visible 사이에는 여전히 작은 창이 있다. 이 창을 0으로 착각하면 장애 조치 직후 사라진 거래, 방금 쓴 주문이 조회되지 않는 사용자 경험, 캐시 무효화 순서 오류가 생긴다.

55. 복제 slot은 standby가 받지 못한 WAL을 primary가 너무 일찍 지우지 않게 막아 주지만, 끊긴 standby를 오래 방치하면 pg_wal 공간을 계속 붙잡는 위험을 만든다. slot은 데이터 유실 방지 장치이면서 동시에 디스크 고갈을 만드는 장치다. 그래서 slot 사용 여부는 '안전하다'가 아니라 'retained WAL을 관측하고 제한할 운영 책임이 생긴다'로 읽어야 한다.

56. replica를 백업 용도로 쓰는 것은 source 부하를 줄이는 좋은 방법이지만, replica가 멈춰 있거나 뒤처져 있으면 백업 기준 시점도 뒤처진다. 백업 파일이 일관적이라는 말과 업무가 원하는 최신 RPO를 만족한다는 말은 다르다. 백업 직전 replica의 replay position, 지연 시간, source와의 연결 상태를 함께 기록해야 나중에 복구 결과를 설명할 수 있다.

57. 읽기 부하 분산은 복제의 흔한 장점이지만, 모든 읽기를 무조건 replica로 보내는 정책은 도메인 일관성을 애플리케이션 밖으로 밀어낸다. 상품 목록처럼 몇 초 늦어도 되는 읽기와 결제 승인 직후 잔액 확인처럼 늦으면 안 되는 읽기는 라우팅 정책이 달라야 한다. 복제 구조는 DB 설정이 아니라 애플리케이션 읽기 계약과 함께 설계해야 한다.

58. 장애 대응에서 'lag가 커졌다'는 알림만 있으면 이미 늦다. WAL 생성량, 네트워크 전송량, replica apply 처리량, long running query, vacuum conflict, disk fsync latency가 같이 보여야 원인을 좁힐 수 있다. 특히 lag가 증가한 시각과 배포, 배치, 인덱스 생성, 통계 수집 시각을 겹쳐 보는 습관이 실무에서 시간을 줄인다.

59. 동기 복제는 commit 응답 전에 일부 standby 확인을 기다리도록 해 데이터 손실 창을 줄이지만, 원격 노드나 네트워크가 느려지면 primary commit latency를 직접 끌어올린다. 고가용성과 쓰기 지연은 독립 변수가 아니다. 금융 원장처럼 손실 창을 좁혀야 하는 쓰기와 로그성 이벤트처럼 지연을 견딜 수 있는 쓰기를 같은 동기 정책으로 묶으면 비용이 과해진다.

60. 복제 토폴로리를 여러 단계로 만들면 primary의 네트워크 부담은 줄지만, downstream replica는 upstream의 지연을 물려받는다. 사용자는 자신이 읽는 replica가 primary와 얼마나 떨어져 있는지 모른다. 운영자는 topology hop 수, upstream 상태, cascading failover 후 timeline 추적을 문서화해야 한다. otherwise read path가 왜 늦었는지 사고 후 설명할 수 없다.

## backup, restore, PITR, failover, consistency

백업과 복구는 파일 보관 기능이 아니라 시간 복원 기능이다. 이 절은 backup file, WAL 또는 binary log, 목표 복구 시점, failover 판단을 하나의 경로로 묶어 설명한다. 핵심 오해는 '백업 파일만 있으면 원하는 시점으로 복구된다'는 생각이다. 실제로 원하는 시점으로 돌아가려면 기준 백업이 있어야 하고, 그 백업 이후 목표 시점까지의 로그가 빠짐없이 있어야 하며, 복원 절차가 격리된 환경에서 검증되어야 한다.

이 절에서 공식 근거로 삼은 자료는 다음과 같다. 링크는 단순 참고가 아니라 본문 판단의 경계다. PostgreSQL과 MySQL은 같은 단어를 쓰더라도 로그 이름, 보존 책임, failover 절차, 관측 뷰가 다르므로, 공통 원리와 제품별 실행 방법을 분리해 읽어야 한다.

- PostgreSQL 18 Backup and Restore: https://www.postgresql.org/docs/current/backup.html
- PostgreSQL 18 Continuous Archiving and PITR: https://www.postgresql.org/docs/current/continuous-archiving.html
- MySQL 8.4 Backup and Recovery: https://dev.mysql.com/doc/refman/8.4/en/backup-and-recovery.html
- MySQL 8.4 Point-in-Time Recovery Using Binary Log: https://dev.mysql.com/doc/refman/8.4/en/point-in-time-recovery-binlog.html

PostgreSQL 공식 문서는 backup approach를 SQL dump, file system level backup, continuous archiving으로 나눈다. PITR을 이해하려면 세 번째 접근이 중요하다. PostgreSQL은 데이터 파일 변경을 WAL에 남기므로, file-system-level base backup과 WAL archive를 결합하면 backup 이후의 변경을 다시 replay할 수 있다. 그리고 replay를 끝까지 하지 않고 특정 시점에서 멈추면 그 시점의 일관된 snapshot을 얻는다. 이 한 문장이 PITR의 뼈대다.

MySQL도 같은 원리를 다른 이름으로 제공한다. full backup이 기준 상태를 제공하고, binary log가 그 이후의 변경 event를 제공한다. `mysqlbinlog`는 binary log event를 사람이 읽거나 다시 적용할 수 있는 형태로 바꾼다. 공식 문서는 binary log file 목록과 현재 status를 확인하고, event time이나 position으로 필요한 구간을 골라 적용할 수 있다고 설명한다. 즉 MySQL PITR도 '백업 파일 + 로그 구간 선택 + 재적용'이다.

```text
restore path mental model

base backup at 01:00
      |
      v
data files represent a possibly fuzzy but recoverable starting state
      |
      v
WAL/binlog sequence: 01:00 -> 01:10 -> 01:20 -> 01:30 -> 01:40
      |
      v
recovery target: 01:27:15 just before wrong DELETE
      |
      v
replay every required record, then stop before the destructive event
```

이 흐름에서 backup file은 출발점이고 WAL/binlog는 시간축이다. 둘 중 하나만 있으면 시간 복원은 닫히지 않는다. backup file만 있으면 backup 시각으로만 돌아간다. 로그만 있으면 적용할 기준 데이터 파일이 없다. 일부 로그가 빠지면 그 뒤 시점은 신뢰할 수 없다. 그래서 복구 가능성을 말할 때는 '마지막 백업 시각'만 말하지 않고, '마지막 검증된 full/base backup, 필요한 로그 chain의 시작과 끝, 목표 시점, 복구 리허설 결과'를 함께 말해야 한다.

| 복구 질문 | 필요한 증거 | PostgreSQL 예 | MySQL 예 |
| --- | --- | --- | --- |
| 어느 상태에서 시작하는가 | base/full backup 식별자 | pg_basebackup label, backup history file | full backup manifest 또는 dump/physical backup 기록 |
| 어느 로그부터 필요한가 | backup start/end 위치 | backup history file이 가리키는 WAL segment | SHOW BINARY LOG STATUS 또는 backup 시점 binlog file/position |
| 어디서 멈출 것인가 | time, LSN, transaction, position | recovery_target_time 또는 관련 recovery target | mysqlbinlog --stop-datetime 또는 position 선택 |
| 복구가 맞는가 | 업무 검증 쿼리 | row count, checksum, 핵심 거래 샘플 | 동일한 검증 쿼리와 application smoke test |
| 운영 전환 가능한가 | fencing과 client routing | old primary 차단, timeline 확인 | source/replica 전환과 write 차단 |

백업 종류도 목적이 다르다. SQL dump는 논리 구조와 데이터를 SQL로 내보내므로 마이그레이션, 일부 객체 복원, 버전 간 이동에 유용하다. 그러나 PostgreSQL 공식 문서가 말하듯 pg_dump와 pg_dumpall은 continuous archiving의 file-system-level backup이 아니며 WAL replay에 필요한 충분한 정보를 담지 않는다. 반대로 물리 backup은 cluster 전체를 빠르게 되살리는 데 유리하지만 일부 table만 골라 사람이 읽기 좋게 복원하는 데는 불편할 수 있다. 좋은 전략은 하나를 만능으로 고르지 않고 RTO/RPO와 복구 단위를 기준으로 조합한다.

RPO는 잃을 수 있는 데이터의 최대 시간 폭이고, RTO는 서비스를 되살리는 데 걸리는 시간이다. 매일 01:00 dump만 있으면 15:00 장애에서 최대 14시간 데이터를 잃을 수 있고, 대용량 dump를 다시 load하는 동안 서비스가 오래 멈춘다. continuous archiving은 RPO를 줄이지만 WAL replay가 길어질 수 있다. 자주 base backup을 뜨면 replay 길이는 줄지만 저장 비용과 백업 부하가 늘어난다. 이 trade-off를 수치로 쓰지 않으면 백업 정책은 안심 문서가 될 뿐이다.

```bash
# PostgreSQL PITR drill sketch: 실제 운영 명령이 아니라 절차를 고정하기 위한 관측 흐름
# 1. 기준 백업 식별
ls -lh /backup/pg/base/2026-05-19T010000Z

# 2. WAL archive 연속성 확인
ls /backup/pg/wal | sort | head
ls /backup/pg/wal | sort | tail

# 3. 복구 target 설정 예시
cat > recovery.conf.sample <<'EOF'
restore_command = 'cp /backup/pg/wal/%f %p'
recovery_target_time = '2026-05-19 14:27:15+09'
recovery_target_action = 'pause'
EOF

# 4. 복구 후 검증 쿼리
psql -c "SELECT count(*) FROM orders WHERE created_at < '2026-05-19 14:27:15+09';"

# MySQL PITR drill sketch
mysql -e "SHOW BINARY LOGS;"
mysql -e "SHOW BINARY LOG STATUS;"
mysqlbinlog --stop-datetime='2026-05-19 14:27:15' binlog.000123 | mysql restore_db
```

failover는 복구와 닮았지만 목적이 다르다. 복구는 손상이나 삭제 이전 시점으로 돌아가는 작업이고, failover는 primary가 더 이상 쓰기를 받을 수 없을 때 standby 또는 replica를 새 primary로 승격해 서비스를 계속하는 작업이다. 두 작업 모두 로그 위치와 일관성이 중요하지만, failover에는 old primary fencing이 추가된다. 기존 primary가 네트워크 partition 때문에 잠깐 고립된 것뿐인데 새 primary를 승격하고 old primary도 계속 쓰기를 받으면 split-brain이 된다.

| 시나리오 | 나쁜 판단 | 필요한 확인 | 운영 기록 |
| --- | --- | --- | --- |
| 실수로 DELETE 실행 | 가장 최근 백업만 복원 | DELETE 직전 target time/position과 로그 연속성 | 삭제 tx id, target, 검증 쿼리 |
| primary 디스크 장애 | 가장 빠른 replica 승격 | replica replay 위치와 데이터 손실 허용치 | promote 시각, last replay LSN/binlog pos |
| ransomware 의심 | 온라인 백업 즉시 연결 | 오염 전 백업과 격리된 복구 환경 | backup integrity, isolated network |
| schema migration 실패 | dump 일부만 덮어쓰기 | DDL 전후 로그와 migration transaction 경계 | migration id, rollback 가능성 |
| replica backup 사용 | backup success만 기록 | backup 당시 replica lag와 source position | source/replica positions, lag |

일관성은 'DB가 켜졌다'보다 강한 조건이다. 복구된 DB가 crash recovery를 통과해 open되더라도 업무적으로 필요한 invariant가 깨졌을 수 있다. 예를 들어 주문 header는 있는데 order_item 일부가 목표 시점 뒤 event라서 없거나, outbox event는 replay되었는데 외부 publish 여부와 맞지 않을 수 있다. 그래서 복구 검증은 DB 엔진 레벨 확인과 업무 샘플 확인을 모두 가져야 한다. 엔진 레벨은 로그 replay 완료, recovery target 도달, checksum, table 접근 가능성이다. 업무 레벨은 핵심 aggregate 수, 금액 합계, 최근 거래 샘플, idempotency/outbox 상태다.

백업 운영에서 자주 빠지는 것은 configuration과 secret이다. PostgreSQL WAL archiving은 데이터 변경은 복원할 수 있지만 수동으로 편집한 `postgresql.conf`, `pg_hba.conf`, `pg_ident.conf` 같은 설정 변경은 별도 백업이 필요하다고 공식 문서가 말한다. MySQL도 데이터 파일과 binary log만으로 application connection secret, parameter group, user privilege drift, router 설정까지 되살릴 수 없다. DB 백업은 서비스 복구의 일부일 뿐이며, 연결 설정과 권한, 배포 환경까지 함께 복구되어야 실제 RTO가 닫힌다.

검증 경로는 최소 세 단계로 나눈다. 첫째, 백업 생성 직후 manifest나 file list와 로그 시작점을 기록한다. 둘째, 주기적으로 격리된 복구 환경에서 임의 target time을 골라 replay한다. 셋째, 복구된 DB에 application smoke test와 업무 invariant query를 실행한다. PASS는 목표 시점 이전 데이터가 있고 목표 시점 이후 잘못된 event가 없으며, 필요한 로그 chain과 backup metadata로 그 결과를 설명할 수 있는 상태다. FAIL은 backup job은 성공했지만 replay를 해 본 적이 없거나, 특정 target time을 재현할 로그가 없거나, 복구 후 어떤 데이터가 맞는지 업무적으로 판정하지 못하는 상태다.

### 복구 리허설 카드: 백업을 실제 복원 능력으로 바꾸기

1. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

2. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

3. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

4. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

5. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

6. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

7. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

8. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

9. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

10. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

11. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

12. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

13. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

14. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

15. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

16. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

17. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

18. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

19. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

20. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

21. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

22. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

23. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

24. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

25. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

26. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

27. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

28. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

29. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

30. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

31. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

32. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

33. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

34. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

35. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

36. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

37. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

38. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

39. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

40. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

41. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

42. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

43. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

44. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

45. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

46. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

47. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

48. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

49. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

50. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

51. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

52. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

53. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

54. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

55. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

56. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

57. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

58. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

59. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

60. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

61. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

62. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

63. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

64. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.

65. 논리 백업은 사람이 읽고 일부 테이블을 옮기기 좋지만, PostgreSQL PITR의 출발점이 되는 파일 시스템 수준 base backup과 같은 것이 아니다. 공식 문서도 pg_dump 계열은 WAL replay에 필요한 파일 시스템 백업이 아니라고 구분한다. 이 차이를 모르고 dump 파일만 보관하면 장애 후 원하는 시점의 전체 cluster를 복원할 수 없다.

66. 백업 암호화와 압축은 저장 비용과 보안을 개선하지만, 복구 시에는 키 관리와 압축 해제 시간이 새로운 실패 지점이 된다. 백업 파일이 있어도 키가 없거나, 키 권한이 특정 사람 계정에만 있거나, 압축 포맷을 읽을 도구가 복구 환경에 없으면 실질적으로 백업은 없는 것과 같다.

67. replica에서 백업을 뜨면 source 부하는 줄지만, backup consistency와 freshness를 별도로 기록해야 한다. replica가 source보다 8분 늦은 상태에서 찍은 백업은 그 자체로 일관적일 수 있어도 업무가 기대한 복구 시점을 만족하지 못한다. 백업 메타데이터에 source position, replica replay position, 측정 시각을 남기는 이유다.

68. 복구 리허설은 production data를 직접 건드리지 않는 격리 환경에서 해야 한다. 같은 binary log나 WAL을 잘못된 cluster에 적용하면 복구 검증이 아니라 장애를 만든다. restore drill에는 대상 cluster 식별, 읽기 전용 검증, 애플리케이션 연결 차단, DNS 또는 secret 분리까지 포함되어야 한다.

69. timeline은 PostgreSQL failover에서 특히 중요하다. standby를 승격하면 새 timeline이 생기고, 이후 다른 standby가 어느 timeline을 따라갈지 정해야 한다. '가장 최신 WAL만 있으면 된다'는 말은 timeline history를 모르면 위험하다. 복구 문서는 파일 이름뿐 아니라 timeline 변화도 함께 설명해야 한다.

70. 백업 보존 정책은 오래 보관할수록 좋은 문제가 아니다. 법적 보존, 개인정보 삭제 요구, 저장 비용, 복구 가능한 chain의 완전성, ransomware 방어를 같이 본다. 가장 오래된 full backup을 지울 때 그 뒤 incremental backup이 의존하는지, WAL/binlog가 어느 backup부터 필요한지 확인하지 않으면 나중에 chain 전체가 무너진다.

71. 백업 성공 메시지는 복구 성공의 증거가 아니다. 백업 파일이 있고, 그 백업이 어느 LSN 또는 binlog position에서 시작하고 끝났는지 알고, 그 이후 로그가 끊기지 않고 보존되어 있으며, 목표 시점까지 실제로 replay해 본 기록이 있어야 복구 가능성을 말할 수 있다. 백업은 쓰는 순간보다 복원 리허설을 통과할 때 운영 자산이 된다.

72. PITR은 '파일을 되돌린다'가 아니라 '기준 백업을 놓고 그 뒤 로그를 원하는 지점까지만 다시 실행한다'는 절차다. PostgreSQL에서는 base backup과 archived WAL의 연속성이 중요하고, MySQL에서는 full backup 뒤 binary log event를 시간 또는 position 기준으로 골라 적용하는 감각이 중요하다. 둘 다 로그가 빠지면 목표 시점은 닫히지 않는다.

73. failover는 promote 버튼 하나가 아니다. primary가 정말 죽었는지, 마지막으로 보낸 로그가 어디까지인지, 승격할 standby가 그 로그를 어디까지 replay했는지, 기존 primary가 돌아왔을 때 split-brain을 막을 fencing이 있는지까지 한 묶음이다. 이 확인 없이 급하게 승격하면 잠깐 살아난 old primary가 서로 다른 쓰기 역사를 만들 수 있다.

74. RTO와 RPO를 구분하지 않으면 백업 전략을 잘못 고른다. RTO는 복구하는 데 걸리는 시간이고, RPO는 잃을 수 있는 데이터의 시간 폭이다. 매일 새벽 논리 dump만 있으면 RTO도 길고 RPO도 하루에 가깝다. base backup과 WAL/binlog 보관을 조합하면 RPO를 줄일 수 있지만, replay 시간이 길어져 RTO가 늘 수 있다.
