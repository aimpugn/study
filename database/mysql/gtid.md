# MySQL GTID (Global Transaction Identifier) 심화 정리

이 문서는 MySQL의 GTID가 **낮은 레벨에서 정확히 어떻게 동작하는지**, GTID를 쓰지 않는 좌표 기반 복제에서 **애플리케이션과 운영이 무엇을 떠안게 되는지**, 그리고 **GTID로 바꾸면 무엇이 좋아지고 그 대가는 무엇인지**를 실무 관점에서 닫는 것을 목표로 한다. 단순히 "트랜잭션마다 전역 ID를 붙인다"는 한 줄 요약을 넘어, binary log 안의 이벤트 한 개 수준까지 내려가서 다시 위로 올라온다.

근거는 MySQL 8.4 / 9.x 공식 매뉴얼을 1차 자료로 삼았고, 문서 끝 [근거](#근거)에 링크를 모았다. 버전에 따라 동작이 다른 부분은 본문에서 시점을 함께 적는다.

## 목차

- [MySQL GTID (Global Transaction Identifier) 심화 정리](#mysql-gtid-global-transaction-identifier-심화-정리)
    - [목차](#목차)
    - [0. 한 문단 직답](#0-한-문단-직답)
    - [1. 왜 GTID가 등장했는가 — 좌표 기반 복제의 근본 문제](#1-왜-gtid가-등장했는가--좌표-기반-복제의-근본-문제)
    - [2. 가장 작은 단위 — GTID의 형식과 표기](#2-가장-작은-단위--gtid의-형식과-표기)
        - [태그드 GTID (tagged GTID) — 2026년 6월 기준 최신 형식](#태그드-gtid-tagged-gtid--2026년-6월-기준-최신-형식)
    - [3. 낮은 레벨 ①: 소스에서 GTID가 만들어지는 순간](#3-낮은-레벨--소스에서-gtid가-만들어지는-순간)
    - [4. 낮은 레벨 ②: `gtid_executed` / `gtid_purged` / `mysql.gtid_executed`](#4-낮은-레벨--gtid_executed--gtid_purged--mysqlgtid_executed)
    - [5. 낮은 레벨 ③: 레플리카의 적용과 auto-skip](#5-낮은-레벨--레플리카의-적용과-auto-skip)
    - [6. 낮은 레벨 ④: auto-positioning 핸드셰이크](#6-낮은-레벨--auto-positioning-핸드셰이크)
        - [purge된 GTID를 요구하면 — 1236 에러](#purge된-gtid를-요구하면--1236-에러)
    - [7. 비-GTID(좌표) 모드에서 애플리케이션·운영이 떠안는 것](#7-비-gtid좌표-모드에서-애플리케이션운영이-떠안는-것)
    - [8. GTID로 바꾸면 좋아지는 것](#8-gtid로-바꾸면-좋아지는-것)
    - [9. 대가 — `enforce_gtid_consistency`가 거는 제약](#9-대가--enforce_gtid_consistency가-거는-제약)
    - [10. 무중단 전환 절차 (OFF → ON)](#10-무중단-전환-절차-off--on)
    - [11. 애플리케이션 관점 실무 체크리스트](#11-애플리케이션-관점-실무-체크리스트)
    - [12. 직접 확인 — 검증 경로](#12-직접-확인--검증-경로)
    - [13. 흔한 오해와 반례](#13-흔한-오해와-반례)
    - [근거](#근거)

---

## 0. 한 문단 직답

GTID는 **하나의 트랜잭션에 복제 토폴로지 전체에서 유일한 이름을 붙이는 장치**다. 이름의 형태는 `서버UUID:순번`(예: `3E11FA47-71CA-11E1-9E33-C80AA9429562:23`)이고, 이 이름은 트랜잭션이 어느 서버를 거쳐 복제되더라도 바뀌지 않는다. GTID가 없던 시절에는 "이 트랜잭션"을 가리키는 좌표가 `binlog 파일 이름 + 파일 안의 바이트 위치`였는데, 이 좌표는 **서버마다 다르고** 장애 복구(failover) 때 사람이 손으로 다시 계산해야 했다. GTID는 그 좌표를 서버 독립적인 전역 이름으로 바꿔서, 레플리카가 "나는 이 이름들까지 실행했다"만 말하면 소스가 "그럼 나머지를 보내주마"라고 답하는 자동 위치 결정(auto-positioning)을 가능하게 한다. 그 결과 failover·토폴로지 재구성·중복 적용 방지가 자동화되지만, 그 대가로 트랜잭션과 GTID가 1:1로 대응해야 한다는 제약(`enforce_gtid_consistency`)이 따라온다. **주의: GTID는 복제의 _식별·정합성_ 문제를 푸는 기능이지, 복제 _지연(lag)_ 이나 _데이터 유실(RPO)_ 을 없애는 기능이 아니다.** 복제는 GTID를 켜도 기본은 여전히 비동기다.

---

## 1. 왜 GTID가 등장했는가 — 좌표 기반 복제의 근본 문제

GTID를 이해하려면 먼저 "GTID가 없으면 무엇이 불편한가"를 봐야 한다. MySQL 복제의 뼈대는 버전과 무관하게 같다. 소스(source, 옛 master)는 모든 변경을 **binary log(binlog)** 에 이벤트로 남기고, 레플리카(replica, 옛 slave)는 그 이벤트를 받아 자신의 **relay log**에 적은 뒤 적용한다. 흐름은 `commit → binlog 기록 → 전송 → relay log 수신 → 적용 → 조회 가능`이다. 이 큰 그림은 [복제·백업·복구 정리](../../interviews/database-deep-dive/09-replication-lag-backup-failover.md)에서 더 넓게 다룬다.

문제는 "레플리카가 소스의 어디까지 따라왔는가"를 무엇으로 표시하느냐에 있다. GTID 이전의 답은 **좌표(coordinates)**, 즉 `(binlog 파일 이름, 파일 안 위치)` 쌍이었다.

```sql
-- 좌표 기반 복제 설정 (GTID 없이): 사람이 파일명과 위치를 직접 지정
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST='10.16.130.42',
  SOURCE_LOG_FILE='mysql-bin.019927',   -- 어느 파일부터
  SOURCE_LOG_POS=1232584;               -- 그 파일의 몇 번째 바이트부터
```

이 모델의 결정적 약점은 **좌표가 서버 로컬(server-local)이라는 점**이다. binlog 파일 이름과 위치는 각 서버가 독립적으로 매기므로, 소스 A의 `mysql-bin.019927/1232584`와 레플리카 B가 같은 트랜잭션을 자기 binlog에 적은 위치는 전혀 다르다. 평소에는 레플리카가 소스의 좌표를 그대로 추적하면 되니 문제가 드러나지 않는다. 문제는 **소스가 죽어서 다른 노드를 새 소스로 승격할 때** 터진다.

기존 소스 A가 죽고, 레플리카 B를 새 소스로 올린 뒤, 또 다른 레플리카 C를 B에 붙인다고 하자. C는 "나는 A 기준 `mysql-bin.019927/1232584`까지 받았다"는 것만 안다. 그런데 B에서 그 트랜잭션이 어느 좌표에 있는지는 A의 좌표와 아무 관계가 없다. 그래서 운영자는 B의 binlog를 뒤져서 "C가 마지막으로 받은 트랜잭션이 B에서는 어디인가"를 사람이 찾아 `SOURCE_LOG_POS`를 지정해야 했다. 한 칸이라도 틀리면 **트랜잭션을 건너뛰어 데이터가 유실되거나, 이미 적용한 트랜잭션을 다시 적용해 중복·깨짐**이 생긴다. 이 좌표 재계산은 자동화하기 어렵고, 장애 한복판에서 사람이 하기에 위험했다.

GTID(MySQL 5.6에서 도입)는 이 문제를 정면으로 푼다. 트랜잭션에 **서버와 무관한 전역 이름**을 붙이면, "C가 어디까지 받았나"를 좌표가 아니라 GTID 집합으로 말할 수 있고, 그 집합은 A·B·C 어디서 봐도 같은 의미를 가진다. 그러면 새 소스 B는 "C가 이미 가진 GTID를 빼고 나머지를 보내면 된다"를 스스로 계산할 수 있다. 좌표라는 서버 로컬 언어를, GTID라는 토폴로지 공용어로 바꾼 것이다.

---

## 2. 가장 작은 단위 — GTID의 형식과 표기

GTID 한 개는 콜론으로 구분된 두 부분이다.

```
source_id : transaction_id
   │              │
   │              └── 그 서버에서 commit된 순서 번호. 1부터 시작하고 0은 없다. 빈틈 없이 단조 증가.
   └── 보통 그 서버의 server_uuid (auto.cnf에 저장된 128비트 UUID)
```

예를 들어 UUID가 `3E11FA47-71CA-11E1-9E33-C80AA9429562`인 서버에서 23번째로 commit된 트랜잭션의 GTID는 다음과 같다.

```
3E11FA47-71CA-11E1-9E33-C80AA9429562:23
```

`source_id`가 `server_uuid`라는 점이 중요하다. `server_id`(정수, 복제용 식별자)가 아니다. UUID는 서버를 처음 기동할 때 데이터 디렉터리의 `auto.cnf`에 생성·저장되며, 이 값이 GTID의 "출신 서버"를 영구히 박아 넣는다. 그래서 트랜잭션이 어느 서버에서 처음 생겼는지가 이름 안에 들어 있다.

여러 GTID를 한꺼번에 나타낼 때는 **GTID set** 표기를 쓴다. 연속 구간은 `1-5`처럼 묶고, 여러 서버는 콤마로 잇는다.

```
# 한 서버에서 1~5번
3E11FA47-71CA-11E1-9E33-C80AA9429562:1-5

# 한 서버에서 1~3번, 11번, 47~49번 (구간과 단일이 섞임)
3E11FA47-71CA-11E1-9E33-C80AA9429562:1-3:11:47-49

# 두 서버에서 온 트랜잭션 (멀티 소스)
2174B383-5441-11E8-B90A-C80AA9429562:1-3,
24DA1671-0C0C-11E8-8442-00059A3C7B00:1-19
```

이 구간 표기는 단순한 출력 형식이 아니라 **저장 효율**과 직결된다. 레플리카가 수백만 개의 트랜잭션을 실행해도 그것이 연속이라면 `uuid:1-9000000` 한 줄로 표현된다. 뒤에서 볼 `mysql.gtid_executed` 테이블도 이 구간을 행으로 저장하고 주기적으로 압축한다.

### 태그드 GTID (tagged GTID) — 2026년 6월 기준 최신 형식

MySQL 8.3(혁신 릴리스)에서 도입되어 8.4 LTS에 들어온 **태그드 GTID**는 세 부분이다.

```
source_id : tag : transaction_id

ed102faf-eb00-11eb-8f20-0c5415bfaa1d:Domain_1:117
```

`tag`는 `[a-zA-Z_]`로 시작하는 최대 32자 문자열로, **같은 서버에서 생긴 트랜잭션을 용도별로 묶는** 사용자 정의 라벨이다. 예컨대 관리용 트랜잭션과 일반 트랜잭션을 다른 태그로 분리하면, GTID set 안에서 그룹별로 구간을 따로 관리할 수 있다.

```
# 같은 서버, 태그 두 개가 각각 독립적인 구간을 가진다
3E11FA47-71CA-11E1-9E33-C80AA9429562:Domain_1:1-3:15-21:Domain_2:8-52
```

> 시점 주의: 태그 기능은 8.3 이상에서만 존재한다. 5.7/8.0 운영 환경이라면 `source_id:transaction_id` 두 부분 형식만 생각하면 된다. 대부분의 실무 환경(2026년 6월 현재 여전히 8.0이 다수)에서는 태그 없는 형식이 기본이다.

---

## 3. 낮은 레벨 ①: 소스에서 GTID가 만들어지는 순간

이제 "트랜잭션마다 ID가 붙는다"를 binlog 이벤트 수준으로 내려가 본다. 핵심 사실은 **GTID가 binlog 안에 별도의 이벤트(`Gtid_log_event`)로, 해당 트랜잭션 바로 앞에 기록된다**는 것이다.

소스에서 클라이언트 트랜잭션이 commit되는 순간 일어나는 일을 순서대로 보면 다음과 같다.

1. **GTID 할당**: 그 서버에서 아직 쓰지 않은 가장 작은 순번을 골라 `server_uuid:순번` 형태의 GTID를 만든다. 단, **binlog에 기록되는 트랜잭션에만** 할당한다. 읽기 전용이거나, 필터로 걸러졌거나, 롤백된 트랜잭션은 GTID를 받지 못한다.
2. **binlog에 원자적 기록**: GTID를 `Gtid_log_event`로 트랜잭션 본문 **바로 앞에** 적는다. 즉 binlog에서 한 트랜잭션은 `[Gtid_log_event] → [BEGIN] → [본문] → [COMMIT/Xid]` 묶음으로 나타난다. GTID와 트랜잭션이 같은 binlog 기록 안에서 원자적으로 함께 남으므로, "GTID는 있는데 트랜잭션은 없다"는 어긋남이 생기지 않는다.
3. **`gtid_executed` 갱신**: commit 직후(원자적이지는 않게) 그 GTID가 `@@GLOBAL.gtid_executed` 집합에 더해진다. 이 변수는 "이 서버가 지금까지 실행을 끝낸 모든 GTID"를 담는다.

이 동작을 만드는 세션 변수가 `gtid_next`다. 기본값은 `AUTOMATIC`이며, 이때 서버가 위처럼 새 GTID를 자동 생성한다. (8.3+에서는 `AUTOMATIC:태그`로 태그를 지정할 수 있다.)

실제로 `mysqlbinlog`로 GTID 모드 binlog를 풀어 보면, 트랜잭션 앞에 `SET @@SESSION.GTID_NEXT=...`가 박혀 있는 것을 볼 수 있다. 아래는 형식을 보여주기 위한 대표적인 출력 모양이다(값은 환경마다 다름).

```text
# at 234
#260617 10:00:00 server id 1  end_log_pos 299  GTID  last_committed=0  sequence_number=1  rbr_only=yes
SET @@SESSION.GTID_NEXT= '3E11FA47-71CA-11E1-9E33-C80AA9429562:23'/*!*/;
# at 299
#260617 10:00:00 server id 1  end_log_pos 372  Query   thread_id=10  exec_time=0  error_code=0
BEGIN
/*!*/;
# at 372
### UPDATE `app`.`account`
###   WHERE @1=10
###   SET  @2=900
# at 520
#260617 10:00:00 server id 1  end_log_pos 551  Xid = 88
COMMIT /*!*/;
```

여기서 읽어야 할 것은 두 가지다. 첫째, `GTID` 이벤트가 트랜잭션의 머리에 붙어 있고, 그 안에서 `GTID_NEXT`를 명시적으로 `3E11FA47-...:23`으로 설정한다. 둘째, 이 트랜잭션을 다른 서버가 재생할 때는 새 GTID를 만드는 게 아니라 **이 이름을 그대로 이어받는다**(5장). 비교를 위해, GTID를 끄면 같은 트랜잭션의 binlog에는 `GTID` 이벤트가 아예 없고 `Query`/`Xid` 이벤트만 남아, 트랜잭션을 가리키는 유일한 좌표가 `at 299` 같은 **파일 내 위치뿐**이 된다. 이 차이가 1장에서 말한 "좌표 대 전역 이름"의 물리적 실체다.

`last_committed`/`sequence_number`는 GTID 자체는 아니고, 레플리카가 트랜잭션을 **병렬로 적용**해도 되는지(서로 충돌하지 않는지)를 판단하기 위한 그룹 커밋 메타데이터다. GTID 모드와 함께 자주 보이므로 같이 언급해 둔다.

---

## 4. 낮은 레벨 ②: `gtid_executed` / `gtid_purged` / `mysql.gtid_executed`

GTID 상태를 추적하는 데는 두 개의 변수와 한 개의 시스템 테이블, 그리고 binlog 파일 헤더 한 종류가 함께 쓰인다. 이 네 가지의 관계를 정확히 잡는 것이 GTID 운영의 절반이다.

**`@@GLOBAL.gtid_executed`** — 이 서버가 실행을 끝낸 모든 GTID의 집합. binlog 파일에 아직 남아 있든 이미 지워졌든 상관없이, "한 번이라도 실행했다"면 여기에 들어 있다. failover나 정합성 비교의 기준이 되는 핵심 값이다.

**`@@GLOBAL.gtid_purged`** — `gtid_executed`의 **부분집합**으로, 실행은 했지만 **현재 어느 binlog 파일에도 본문이 남아 있지 않은** GTID들이다. 다음 세 경로로 채워진다.

- binlog를 끈 채(`log_bin=OFF`) 적용한 복제 트랜잭션
- 오래되어 purge(삭제)된 binlog 파일에 있던 트랜잭션
- `SET @@GLOBAL.gtid_purged`로 운영자가 명시적으로 "이건 이미 적용된 것으로 쳐라"라고 기록한 경우

`gtid_purged`가 중요한 이유는, 새로 붙는 레플리카가 요구하는 GTID가 이미 purge되어 있으면 소스가 그 본문을 보내줄 수 없기 때문이다(6장의 1236 에러).

**`mysql.gtid_executed` 테이블** — `gtid_executed`의 디스크 영속 사본이다. 구조는 다음과 같다(8.4 기준, `gtid_tag` 칼럼은 태그드 GTID 지원으로 추가됨).

```sql
CREATE TABLE gtid_executed (
  source_uuid    CHAR(36)  NOT NULL,
  interval_start BIGINT    NOT NULL,
  interval_end   BIGINT    NOT NULL,
  gtid_tag       CHAR(32)  NOT NULL,
  PRIMARY KEY (source_uuid, gtid_tag, interval_start)
);
```

이 테이블의 존재 이유는 **binlog가 없어도 GTID 상태를 잃지 않기 위해서**다. 레플리카에서 binlog를 꺼 두었거나, binlog가 모두 purge되어도, 이 테이블이 "어디까지 실행했는가"를 보존한다. 시간이 지나면 개별 GTID가 행 하나씩 쌓이므로, MySQL은 이를 구간으로 합쳐 압축한다(`interval_start=31, interval_end=31`인 행들이 `31~35` 한 행으로). 압축 주기는 `gtid_executed_compression_period`로 조절하는데, binlog가 켜져 있으면 binlog 회전 시점에 갱신·압축되므로 별도 압축 스레드가 불필요하고, 공식 문서는 이 경우 값을 `0`으로 두기를 권한다.

**`Previous_gtids_log_event`** — 각 binlog 파일의 헤더에 들어가는 이벤트로, **그 파일이 시작되기 직전까지 서버가 실행한 GTID 전체**를 담는다. 즉 binlog 파일 하나만 열어도 "이 파일이 다루는 GTID 범위의 시작점"을 O(1)로 알 수 있다. 6장의 auto-positioning에서 소스가 "어느 파일부터 보내야 하는가"를 빠르게 찾는 데 바로 이 헤더를 쓴다.

```text
# mysqlbinlog로 binlog 파일 머리를 보면:
# at 4
#260617 09:00:00 server id 1  Previous-GTIDs
# 3E11FA47-71CA-11E1-9E33-C80AA9429562:1-22
```

정리하면, `gtid_executed`(메모리·논리) ⊇ `gtid_purged`(본문 없는 부분), `mysql.gtid_executed`(디스크 영속), `Previous_gtids_log_event`(파일별 시작점)가 같은 GTID 우주를 서로 다른 각도에서 본 그림이다.

---

## 5. 낮은 레벨 ③: 레플리카의 적용과 auto-skip

소스에서 만들어진 GTID가 레플리카로 건너가면, 레플리카는 **그 GTID를 새로 만들지 않고 그대로 물려받는다**. 이것이 "전역 이름"이 토폴로지 전체에서 유지되는 메커니즘이다. 레플리카에서 한 트랜잭션을 적용하는 과정은 다음과 같다.

1. relay log에서 트랜잭션 앞의 `Gtid_log_event`를 읽어, 세션의 `@@SESSION.gtid_next`를 그 GTID(`uuid:N`)로 설정한다. "다음 트랜잭션은 반드시 이 이름을 쓰라"는 지시다.
2. **소유권 확인**: 다른 스레드가 같은 GTID를 점유 중인지 본다. `@@GLOBAL.gtid_owned`가 현재 사용 중인 GTID와 그 소유 스레드를 보여준다.
3. **auto-skip 판단**: 이 GTID가 이미 `gtid_executed`에 있으면 — 즉 이미 실행했다면 — **에러 없이 트랜잭션 전체를 건너뛴다.** 본문 SQL은 한 줄도 실행되지 않는다.
4. 아직 실행 안 했으면 적용하고, `gtid_next`가 이미 정해져 있으니 새 GTID를 만들지 않고 소스의 이름으로 commit한다.
5. binlog가 켜져 있으면 자기 binlog에 `Gtid_log_event`로 영속화하고, 꺼져 있으면 `mysql.gtid_executed`에 직접 적는다. 그 후 `gtid_executed`에 더해진다.

이 **auto-skip(자동 건너뛰기)** 이 GTID의 핵심 안전장치다. 공식 문서의 표현을 빌리면, "어떤 GTID로 commit된 트랜잭션은 한 서버에서 한 번을 넘겨 적용될 수 없다." 이미 실행한 GTID를 다시 받으면 조용히 무시하므로, **중복 적용이 구조적으로 막힌다.** 다이아몬드 토폴로지(한 레플리카가 두 경로로 같은 트랜잭션을 받는 구조)에서 두 번째로 도착한 사본이 자동으로 버려지는 것도 이 덕분이다.

동시성도 안전하다. 같은 GTID로 두 세션이 동시에 트랜잭션을 시작하려 하면, 한쪽이 commit 또는 rollback할 때까지 다른 쪽은 **블록**된다. 먼저 시작한 쪽이 commit하면 나머지는 블록이 풀리며 전부 auto-skip되고, rollback하면 한 세션만 진행한다.

> 비교 포인트: 좌표 기반 복제에는 이 "이미 했는지"를 트랜잭션 단위로 판단하는 장치가 없다. 그래서 좌표를 잘못 지정하면 같은 트랜잭션이 두 번 적용되거나(중복) 건너뛰어진다(유실). auto-skip은 "정확히 한 번 적용"을 좌표 정확도가 아니라 트랜잭션 이름으로 보장한다.

---

## 6. 낮은 레벨 ④: auto-positioning 핸드셰이크

GTID가 주는 가장 큰 실무 가치는 **auto-positioning**이다. 레플리카를 `SOURCE_AUTO_POSITION=1`(옛 `MASTER_AUTO_POSITION=1`)로 설정하면, 좌표를 한 번도 지정하지 않고도 복제가 자동으로 올바른 지점에서 시작한다.

```sql
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST   = '10.16.130.42',
  SOURCE_USER   = 'repl',
  SOURCE_AUTO_POSITION = 1;   -- SOURCE_LOG_FILE / SOURCE_LOG_POS 를 주지 않는다
```

레플리카가 소스에 접속할 때의 핸드셰이크는 정확히 이렇게 동작한다.

1. **레플리카가 자신의 GTID 집합을 보낸다.** 이 집합은 `@@GLOBAL.gtid_executed`(이미 실행 완료)와 Performance Schema의 `replication_connection_status.RECEIVED_TRANSACTION_SET`(받기는 했고 아직 적용 중일 수 있는 것)의 **합집합**이다. 말로 풀면 "내가 이미 받았거나 실행한 GTID는 이것들이다."
2. **소스가 차집합을 계산한다.** 소스는 가장 최근 binlog 파일부터 거꾸로 각 파일 헤더의 `Previous_gtids_log_event`를 확인해, 레플리카가 빠뜨린 GTID를 포함하는 첫 파일을 찾는다(4장에서 본 파일 헤더가 여기서 쓰인다). 그 파일부터 읽으며 **레플리카가 안 가진 GTID의 트랜잭션만 보내고, 이미 가진 것은 건너뛴다.**
3. 레플리카는 받은 트랜잭션을 5장의 절차로 적용하고, auto-skip이 혹시 모를 중복을 또 한 번 막는다.

이때 운영자가 보는 두 값이 `SHOW REPLICA STATUS`(옛 `SHOW SLAVE STATUS`)의 **`Retrieved_Gtid_Set`**(이 레플리카가 소스에서 받아온 GTID)과 **`Executed_Gtid_Set`**(실제로 실행을 끝낸 GTID, 곧 `@@GLOBAL.gtid_executed`)이다. 둘의 차이가 "받았지만 아직 적용 못 한" 백로그이며, 이것이 GTID 언어로 표현된 복제 지연이다.

### purge된 GTID를 요구하면 — 1236 에러

auto-positioning이 항상 자동으로 복구되는 것은 아니다. 레플리카가 요구하는 GTID의 본문이 소스에서 이미 purge되었거나 `gtid_purged`에 들어가 있으면, 소스는 다음 에러를 보내고 복제가 시작되지 않는다.

- 소스 측 서버 에러: **`ER_SOURCE_HAS_PURGED_REQUIRED_GTIDS`** (MASTER→SOURCE 용어 정리 이전 구버전 명칭은 `ER_MASTER_HAS_PURGED_REQUIRED_GTIDS`)
- 레플리카 I/O 스레드가 보는 와이어 에러: **`Last_IO_Errno: 1236`** 과 "The replica is connecting ... but the source has purged binary logs containing GTIDs that the replica requires" 류 메시지
- 소스 에러 로그에는 어떤 GTID가 비었는지 `ER_FOUND_MISSING_GTIDS` 경고로 남는다

이 상황은 **자동 복구 불가**다. 따라잡는 데 필요한 트랜잭션 본문이 사라졌기 때문이다. `SOURCE_AUTO_POSITION`을 끄고 재접속하면 빠진 트랜잭션을 그냥 잃을 뿐이므로 해선 안 된다. 올바른 대응은 (a) 그 빠진 GTID들을 가진 다른 소스에서 받아오거나, (b) 더 최신 백업으로 레플리카를 새로 만드는 것이다. 실무에서 이 에러는 보통 "레플리카가 오래 멈춰 있는 동안 소스가 binlog를 너무 빨리 지웠다"는 신호이므로, `binlog_expire_logs_seconds`(보존 기간)와 레플리카 다운타임의 관계를 같이 봐야 한다.

---

## 7. 비-GTID(좌표) 모드에서 애플리케이션·운영이 떠안는 것

여기부터가 "GTID가 아닐 때 MySQL을 쓰는 쪽이 무엇을 고려해야 하는가"라는 질문에 대한 답이다. 먼저 경계를 분명히 하자. **복제를 전혀 쓰지 않는 단일 서버라면 GTID 여부는 사실상 무의미하다.** GTID가 푸는 문제는 전부 "여러 노드 사이에서 같은 트랜잭션을 어떻게 식별·정합시키는가"이기 때문이다. 그러나 현실의 운영 DB는 거의 항상 읽기 분산용 레플리카, HA용 standby, 백업용 노드, CDC 파이프라인 중 하나 이상을 달고 있고, 그 순간부터 좌표 모드의 부담이 애플리케이션과 운영에 스며든다.

**① failover 때 좌표를 사람이 재계산해야 한다.** 1장에서 설명한 그대로다. 소스가 죽으면 새 소스의 binlog에서 각 레플리카가 멈춘 지점을 찾아 `SOURCE_LOG_FILE`/`SOURCE_LOG_POS`를 손으로 맞춰야 하고, 틀리면 유실·중복이 난다. 이 위험 때문에 비-GTID 환경의 failover는 자동화 난도가 높고, 사람이 개입하는 절차서에 의존하게 된다.

**② read-after-write를 좌표로 직접 다뤄야 한다.** 애플리케이션이 쓰기 직후 같은 데이터를 레플리카에서 읽어야 한다면, "레플리카가 내 쓰기를 따라잡았는가"를 확인해야 한다. 좌표 모드의 도구는 `SOURCE_POS_WAIT('파일', 위치)`(옛 `MASTER_POS_WAIT`)인데, 이는 애플리케이션이 **방금 한 쓰기의 binlog 파일명·위치를 알아내서 들고 다녀야** 동작한다. 좌표는 서버 로컬이라 라우팅이 끼면 더 까다롭다. (GTID 모드에서는 이것이 GTID 집합 기반의 `WAIT_FOR_EXECUTED_GTID_SET`으로 훨씬 단순해진다 — 8장.)

**③ 토폴로지를 바꾸면 매번 좌표를 다시 잡아야 한다.** 레플리카를 다른 소스 아래로 옮기거나, 중간 단계 복제(cascading)를 끼우면 좌표가 전부 새로 계산된다.

**④ CDC·복제 도구가 좌표를 추적하며, 그 좌표의 연속성에 묶인다.** 이 점은 실제 환경 로그로 보는 게 가장 명확하다. 아래는 이 저장소의 [Debezium 트러블슈팅 노트](../../troubleshooting/debezium.md)에 있는, **GTID가 꺼진** MySQL(`10.16.130.42`)에 붙은 Debezium MySQL 커넥터의 실제 오프셋 로그다.

```text
sourceInfo=SourceInfo [currentGtid=null,
                       currentBinlogFilename=mysql-bin.019927,
                       currentBinlogPosition=1232584, ...],
restartGtidSet=null, currentGtidSet=null,
restartBinlogFilename=mysql-bin.019927, restartBinlogPosition=1232584
```

`currentGtid=null`, `currentGtidSet=null`이라는 것은 이 서버가 GTID 모드가 아니어서, Debezium이 변경 위치를 **GTID가 아니라 `mysql-bin.019927`이라는 파일명과 `1232584`라는 바이트 위치로** 추적하고 있다는 뜻이다. 이 모드에서 커넥터의 재시작 지점(`restartBinlogFilename`/`Position`)은 그 특정 소스의 binlog 좌표에 묶인다. 그래서 소스가 교체되거나 그 binlog가 purge되면 커넥터의 오프셋은 더 이상 유효하지 않고, 스냅샷을 다시 뜨거나 오프셋을 손보는 운영 작업이 필요해진다. Debezium은 소스가 GTID 모드이면 GTID 집합으로 재시작 지점을 관리할 수 있어 소스 전환에 훨씬 강하다 — 이것이 CDC를 운영하는 팀이 GTID를 선호하는 직접적 이유다.

**⑤ 애플리케이션 코드 자체는 대개 좌표를 직접 보지 않는다.** 정직하게 말하면, 평범한 JDBC/ORM 애플리케이션 코드는 binlog 좌표를 다루지 않는다. 부담은 주로 (a) 인프라/HA 자동화, (b) read-after-write가 필요한 일부 경로, (c) CDC·복제·백업 도구 계층에 몰린다. 그래서 "비-GTID라 코드가 복잡해진다"기보다 "비-GTID라 **운영과 도구 계층의 자동화가 어렵고 사람 절차에 의존하게 된다**"가 더 정확한 진술이다.

---

## 8. GTID로 바꾸면 좋아지는 것

7장의 부담을 뒤집으면 그대로 GTID의 이득이 된다. 다만 "무엇이 좋아지는가"를 과장 없이 적는 것이 중요하다.

**① failover·switchover의 단순화 (가장 큰 이득).** `SOURCE_AUTO_POSITION=1`만 설정하면 좌표 계산이 사라진다. 새 소스로 레플리카를 붙일 때 "너 어디까지 했어? 나머지 줄게"가 자동으로 일어난다. 이 덕분에 GTID는 사실상 모든 현대 MySQL HA 스택의 전제 조건이다 — 그룹 복제(Group Replication)와 InnoDB Cluster는 **GTID를 필수로 요구**하고, Orchestrator·ProxySQL 같은 외부 HA·라우팅 도구도 GTID를 기반으로 토폴로지를 안전하게 재배치한다.

**② 중복·유실의 구조적 방지.** auto-skip(5장)이 "정확히 한 번 적용"을 트랜잭션 이름으로 보장한다. 좌표 한 칸 실수로 인한 유실/중복이라는 실패 모드 자체가 사라진다.

**③ read-after-write가 좌표 대신 GTID로.** 애플리케이션이 쓰기 후 받은 GTID(예: `SELECT @@SESSION.gtid_next` 또는 `session_track_gtids` 활용)를 들고, 레플리카에서 다음을 호출하면 그 트랜잭션이 적용될 때까지 기다릴 수 있다.

```sql
-- 레플리카에서: 이 GTID 집합이 실행될 때까지 (최대 10초) 대기
SELECT WAIT_FOR_EXECUTED_GTID_SET('3E11FA47-71CA-11E1-9E33-C80AA9429562:23', 10);
-- 0 반환: 따라잡음 / 1 반환: 타임아웃
```

좌표를 들고 다닐 필요 없이, 서버 독립적인 GTID 하나로 read-after-write 일관성을 맞출 수 있다.

**④ 정합성 진단·비교가 쉬워진다.** 두 노드의 `gtid_executed`를 집합 연산으로 비교하면 "누가 무엇을 빠뜨렸는가"가 바로 나온다.

```sql
-- 소스에는 있는데 이 레플리카에 없는 GTID = 아직 안 따라온 트랜잭션
SELECT GTID_SUBTRACT('<소스의 gtid_executed>', @@GLOBAL.gtid_executed);

-- A의 집합이 B의 집합에 포함되는가? (1이면 B가 A를 모두 가짐)
SELECT GTID_SUBSET('<A>', '<B>');
```

좌표 모드에서는 "두 서버가 같은 지점인가"를 이렇게 단순한 집합 연산으로 답할 수 없다.

**⑤ 백업·PITR과의 결합.** GTID 모드에서 백업(예: mysqldump의 `--set-gtid-purged`, Percona XtraBackup, MySQL Enterprise Backup)은 그 백업이 담은 `gtid_executed`를 함께 기록한다. 복원한 노드를 복제에 붙일 때 `gtid_purged`만 맞춰 주면 auto-positioning이 알아서 나머지를 따라잡는다. PITR의 시간 좌표 위에 GTID라는 또 하나의 정합성 축이 생기는 셈이다. (백업·PITR의 큰 그림은 [복제·백업·복구 정리](../../interviews/database-deep-dive/09-replication-lag-backup-failover.md) 참고.)

**무엇이 좋아지지 _않는지_ 도 분명히.** GTID는 복제를 동기로 만들지 않는다. 기본 복제는 GTID를 켜도 비동기이고, 따라서 복제 지연과 장애 시 RPO(유실 가능 데이터 폭)는 그대로 남는다. 유실 창을 줄이려면 그것은 GTID가 아니라 **semisynchronous 복제**나 **Group Replication**의 영역이다. GTID는 "어느 트랜잭션을 어디까지 적용했나"를 정확히 만들 뿐, "얼마나 빨리/안 잃고 전달되나"를 바꾸지 않는다.

---

## 9. 대가 — `enforce_gtid_consistency`가 거는 제약

GTID의 자동화는 공짜가 아니다. auto-skip과 auto-positioning이 성립하려면 **트랜잭션 하나에 GTID 하나가 정확히 1:1로 대응**해야 한다. 한 트랜잭션이 GTID 두 개를 필요로 하거나, 부분만 복제되면 이름이 깨진다. 그래서 GTID 모드는 `enforce_gtid_consistency` 변수로 "1:1을 깨뜨릴 수 있는 구문"을 금지한다. 값은 `OFF`/`WARN`/`ON`이며, GTID 모드를 켜려면 `ON`이어야 한다.

금지·제약되는 대표 항목은 다음과 같다.

**① 트랜잭션·비트랜잭션 엔진 혼합.** 한 구문이나 한 트랜잭션 안에서 InnoDB(트랜잭션) 테이블과 MyISAM(비트랜잭션) 테이블을 함께 갱신할 수 없다. 비트랜잭션 부분은 롤백되지 않으므로 GTID 하나로 원자성을 표현할 수 없기 때문이다. (실무에서는 모든 테이블을 InnoDB로 쓰면 자연히 해결된다.)

**② `CREATE TABLE ... SELECT`.** 이 구문은 DDL(테이블 생성)과 DML(행 삽입)을 한 덩어리로 섞는다. 원자적 DDL을 지원하는 엔진에서는 단일 트랜잭션으로 기록되지만, GTID 일관성 관점에서 전통적으로 문제가 되어 `ON`에서 막힌다. **배치·마이그레이션 스크립트에서 자주 쓰는 패턴이라 전환 시 가장 흔히 걸리는 지점이다.** `CREATE TABLE ...` 후 `INSERT ... SELECT`로 분리하면 된다.

**③ 트랜잭션/프로시저/함수/트리거 안의 `CREATE TEMPORARY TABLE` / `DROP TEMPORARY TABLE`.** `binlog_format=STATEMENT`일 때 트랜잭션 컨텍스트 안에서 임시 테이블 생성/삭제가 금지된다. (`ROW`/`MIXED`에서는 임시 테이블이 binlog에 기록되지 않으므로 트랜잭션 안에서도 허용되지만 레플리카로 복제되지 않는다.)

또한 복제 운영 명령에도 제약이 생긴다.

**④ `sql_replica_skip_counter`(옛 `sql_slave_skip_counter`)를 쓸 수 없다.** 좌표 모드에서 "다음 N개 이벤트를 건너뛰어라"로 복제를 강제 진행시키던 방법이 GTID 모드에선 막힌다. 대신 건너뛰려는 트랜잭션의 GTID를 빈 트랜잭션으로 직접 commit해서 `gtid_executed`에 넣는 방식을 쓴다.

```sql
-- GTID 모드에서 특정 트랜잭션을 의도적으로 건너뛰는 정석
SET GTID_NEXT='3E11FA47-71CA-11E1-9E33-C80AA9429562:24';
BEGIN; COMMIT;               -- 빈 트랜잭션으로 그 GTID를 "실행됨"으로 표시
SET GTID_NEXT='AUTOMATIC';
START REPLICA;
```

**⑤ `IGNORE_SERVER_IDS`를 쓸 수 없다.** 이미 적용된 트랜잭션은 auto-skip이 자동으로 무시하므로 불필요하고, GTID와 함께 쓰면 충돌한다.

마지막으로 중요한 단서: `enforce_gtid_consistency`는 **binlog에 기록되는 구문에만** 적용된다. binlog가 꺼져 있거나 필터로 제외된 구문은 검사·강제되지 않는다.

---

## 10. 무중단 전환 절차 (OFF → ON)

GTID는 MySQL 5.7.6부터 **서버 재시작 없이** 운영 중에 켤 수 있다(그 이전 5.6에서는 모든 서버를 멈추고 동시에 바꿔야 했다). 핵심 아이디어는 두 개의 중간 상태 `OFF_PERMISSIVE`/`ON_PERMISSIVE`를 두어, 토폴로지 안의 서버들이 한꺼번에 점프하지 않고 단계적으로 넘어가게 하는 것이다. 각 상태의 의미는 다음과 같다.

- `OFF`: 새 트랜잭션은 익명(GTID 없음). GTID 달린 복제 트랜잭션은 거부.
- `OFF_PERMISSIVE`: 새 트랜잭션은 익명. 복제로 들어오는 것은 **GTID·익명 둘 다 허용**.
- `ON_PERMISSIVE`: 새 트랜잭션은 GTID 생성. 복제로 들어오는 것은 둘 다 허용.
- `ON`: 새 트랜잭션은 GTID 생성. **GTID 달린 것만 허용**, 익명 거부.

`OFF`인 서버는 GTID를 거부하므로, 소스가 갑자기 GTID를 만들기 시작하면 레플리카가 깨진다. 그래서 "모두가 둘 다 받아들이는" permissive 구간을 거쳐 안전하게 넘어간다. 전체 절차는 다음과 같다(토폴로지의 **모든 서버**에 적용).

1. 모든 서버에서 일관성 위반을 경고로만 표시: `SET @@GLOBAL.enforce_gtid_consistency = WARN;` — 정상 워크로드를 돌리며 에러 로그에 9장의 위반 구문이 찍히는지 관찰한다. 찍히면 애플리케이션·배치를 먼저 고친다.
2. 위반이 없으면 강제로 전환: `SET @@GLOBAL.enforce_gtid_consistency = ON;`
3. 모든 서버: `SET @@GLOBAL.gtid_mode = OFF_PERMISSIVE;` (서버 간 순서는 무관, 단 다음 단계 전에 전부 완료)
4. 모든 서버: `SET @@GLOBAL.gtid_mode = ON_PERMISSIVE;`
5. 모든 서버에서 익명 트랜잭션이 빠질 때까지 대기:
   ```sql
   SHOW STATUS LIKE 'Ongoing_anonymous_transaction_count';  -- 0 이 될 때까지
   ```
   레플리카에서는 잠깐 다시 0이 아닐 수 있는데, **한 번이라도 0에 도달**했으면 충분하다.
6. 5단계까지 만들어진 익명 트랜잭션이 모든 서버에 복제되어 사라질 때까지 기다린다(이 동안 쓰기는 계속 가능).
7. (PITR용 binlog를 쓴다면) 익명 트랜잭션이 든 옛 binlog가 더 이상 필요 없어질 때까지 기다리고 `FLUSH LOGS` 후 정리.
8. 최종 전환: `SET @@GLOBAL.gtid_mode = ON;`
9. 설정 파일에 영속화하고(`gtid_mode=ON`, `enforce_gtid_consistency=ON`) 레플리카를 auto-positioning으로 전환:
   ```sql
   STOP REPLICA;
   CHANGE REPLICATION SOURCE TO SOURCE_AUTO_POSITION = 1;
   START REPLICA;
   ```

이 절차는 **언제든 멈추고, 재개하고, 역방향으로 되돌릴 수 있다**는 점이 설계상 보장된다(반대로 끄려면 `ON → ON_PERMISSIVE → OFF_PERMISSIVE → OFF` 역순). 8.0.23+에서는 이미 GTID인 소스에서 익명 트랜잭션을 받는 레플리카를 위해 `ASSIGN_GTIDS_TO_ANONYMOUS_TRANSACTIONS` 옵션도 있어, 마이그레이션 조합이 더 유연해졌다.

---

## 11. 애플리케이션 관점 실무 체크리스트

GTID를 켜거나 GTID 환경에서 애플리케이션을 운영할 때 코드·배치·도구 계층에서 실제로 점검할 것들이다.

- **평범한 CRUD 코드는 손댈 게 거의 없다.** JDBC/ORM 레벨에서 GTID는 투명하다. 부담은 아래 특수 경로에 몰린다.
- **read-after-write가 필요한 경로를 식별한다.** 결제 직후 잔액 확인처럼 자기 쓰기를 즉시 읽어야 하는 요청은 소스로 보내거나, GTID 기반 `WAIT_FOR_EXECUTED_GTID_SET`으로 따라잡은 레플리카로만 보낸다. 커넥터에 따라 `session_track_gtids=OWN_GTID`로 방금 쓴 GTID를 응답에서 받아올 수 있다.
- **배치·마이그레이션 SQL에서 9장의 금지 구문을 제거한다.** 특히 `CREATE TABLE ... SELECT`와 트랜잭션 안 임시 테이블이 흔한 함정이다. 스키마 마이그레이션 도구([Flyway 노트](../migration/flyway/flyway.md) 등)로 DDL을 관리한다면, 마이그레이션 스크립트가 이 제약을 지키는지 전환 전에 `enforce_gtid_consistency=WARN`으로 미리 확인한다.
- **CDC 파이프라인은 GTID를 선호하도록 설정한다.** Debezium MySQL 커넥터는 소스가 GTID 모드이면 GTID 집합으로 오프셋을 관리해 소스 전환에 강하다(7장의 실제 로그 비교 참고). 비-GTID이면 binlog 파일+위치에 묶인다.
- **HA·라우팅 스택의 전제를 확인한다.** InnoDB Cluster/Group Replication은 GTID가 필수다. ProxySQL·Orchestrator도 GTID를 전제로 동작이 단순해진다.
- **커넥션 단위로 `gtid_next`를 만지지 않는다.** auto-skip 디버깅 등 특수 상황 외에 애플리케이션이 `GTID_NEXT`를 직접 설정하면 후속 트랜잭션 동작이 꼬인다.

---

## 12. 직접 확인 — 검증 경로

작은 실험으로 위 내용을 직접 재생할 수 있다. 각 항목은 무엇을 보면 PASS이고 무엇이면 FAIL인지 함께 적는다.

**현재 모드와 상태 확인.**
```sql
SELECT @@GLOBAL.gtid_mode, @@GLOBAL.enforce_gtid_consistency;
SELECT @@GLOBAL.gtid_executed, @@GLOBAL.gtid_purged;
```
PASS: `gtid_mode=ON`이면 `gtid_executed`가 `uuid:1-N` 형태로 채워진다. FAIL: `gtid_mode=OFF`인데 `gtid_executed`가 비어 있지 않다면 과거 이력이 섞인 것이므로 출처를 확인.

**binlog 안의 GTID 이벤트 눈으로 보기.**
```bash
mysqlbinlog --base64-output=DECODE-ROWS -v mysql-bin.000001 | grep -A1 'GTID_NEXT'
```
PASS: 트랜잭션마다 `SET @@SESSION.GTID_NEXT='uuid:N'`이 본문 앞에 보인다(3장). FAIL: GTID 이벤트가 없고 `Query`/`Xid`만 있으면 그 binlog는 비-GTID(익명) 구간이다.

**auto-skip 재생.** 소스에서 트랜잭션 하나를 commit해 GTID를 만들고, 레플리카에 정상 적용시킨 뒤, 레플리카에서 같은 GTID로 빈 트랜잭션을 commit해 본다.
```sql
-- 레플리카에서, 이미 적용된 GTID를 다시 시도
SET GTID_NEXT='<그 GTID>'; BEGIN; INSERT ...; COMMIT; SET GTID_NEXT='AUTOMATIC';
```
PASS: 에러 없이 INSERT가 무시된다(이미 실행한 GTID라 auto-skip). FAIL: 같은 행이 두 번 들어가면 GTID 모드가 아니거나 다른 GTID를 쓴 것.

**auto-positioning 핸드셰이크 관측.**
```sql
SHOW REPLICA STATUS\G   -- Auto_Position: 1, Retrieved_Gtid_Set, Executed_Gtid_Set 확인
```
PASS: `Auto_Position=1`이고 `Executed_Gtid_Set`이 소스를 따라 증가한다. `Retrieved`−`Executed` 차이가 곧 적용 백로그. FAIL: `Last_IO_Errno=1236`이면 6장의 purge 에러.

**정합성 비교.**
```sql
-- 소스에 있고 레플리카에 없는 트랜잭션
SELECT GTID_SUBTRACT('<source gtid_executed>', '<replica gtid_executed>');
```
PASS: 빈 문자열이면 레플리카가 소스를 완전히 따라잡음. 비어 있지 않으면 그 GTID들이 미적용분.

---

## 13. 흔한 오해와 반례

- **"GTID를 켜면 복제가 동기가 된다."** ✗ 기본 복제는 GTID와 무관하게 비동기다. GTID는 식별·정합성 기능이지 전달 보장 기능이 아니다. 유실 창을 줄이려면 semisync나 Group Replication이 필요하다(8장 마지막).
- **"GTID면 데이터 유실이 없다."** ✗ 위와 같은 오해. 소스가 commit하고 레플리카로 보내기 전에 죽으면 비동기 복제에서는 그 트랜잭션이 유실될 수 있다. RPO는 복제 전달 방식의 문제다.
- **"GTID면 read-after-write가 자동으로 보장된다."** ✗ 여전히 복제 지연이 있으므로 방금 쓴 값이 레플리카에 즉시 보이지 않는다. 다만 그것을 다루는 도구가 좌표(`SOURCE_POS_WAIT`)에서 GTID(`WAIT_FOR_EXECUTED_GTID_SET`)로 단순해졌을 뿐이다.
- **"`SOURCE_AUTO_POSITION=1`이면 어떤 장애든 자동 복구된다."** ✗ 레플리카가 요구하는 GTID 본문이 소스에서 purge되면 `1236` 에러로 멈추고, 다른 소스나 새 백업이 필요하다(6장).
- **`gtid_mode`와 `enforce_gtid_consistency`를 같은 것으로 착각.** ✗ 둘은 별개 변수다. `enforce_gtid_consistency`는 "GTID를 깨는 구문을 막는 문지기"이고, `gtid_mode`는 "실제로 GTID를 생성·요구하는가"이다. 전환 시 항상 `enforce_gtid_consistency`를 먼저 닫고 `gtid_mode`를 올린다(10장 순서).
- **"`server_id`가 GTID의 출신 식별자다."** ✗ GTID의 `source_id`는 `server_uuid`(auto.cnf의 128비트 UUID)이지 `server_id`(정수)가 아니다.
- **"좌표 모드보다 GTID가 항상 빠르다."** ✗ GTID는 성능 최적화 기능이 아니다. `gtid_executed` 갱신·`mysql.gtid_executed` 영속화의 약간의 오버헤드가 있고, 이득은 속도가 아니라 운영 안전성·자동화다.

---

## 근거

MySQL 8.4 / 9.x 공식 매뉴얼(Replication with Global Transaction Identifiers)을 1차 자료로 삼았다.

- GTID 개념·형식·저장 (태그드 GTID, `mysql.gtid_executed`, 구간 표기, auto-skip): https://dev.mysql.com/doc/refman/8.4/en/replication-gtids-concepts.html
- GTID 생애주기 (생성·binlog 기록·`gtid_executed`/`gtid_purged`): https://dev.mysql.com/doc/refman/8.4/en/replication-gtids-lifecycle.html
- GTID auto-positioning 핸드셰이크 (`Retrieved`/`Executed_Gtid_Set`, `ER_SOURCE_HAS_PURGED_REQUIRED_GTIDS`, `ER_FOUND_MISSING_GTIDS`): https://dev.mysql.com/doc/refman/8.4/en/replication-gtids-auto-positioning.html
- GTID 모드 무중단 활성화 절차 (`WARN`/`ON`, `OFF_PERMISSIVE`/`ON_PERMISSIVE`, `Ongoing_anonymous_transaction_count`): https://dev.mysql.com/doc/refman/8.4/en/replication-mode-change-online-enable-gtids.html
- GTID 기반 복제 제약 (`enforce_gtid_consistency`, `CREATE TABLE ... SELECT`, 임시 테이블, `sql_replica_skip_counter`, `IGNORE_SERVER_IDS`): https://dev.mysql.com/doc/refman/8.4/en/replication-gtids-restrictions.html
- GTID 관련 시스템 변수: https://dev.mysql.com/doc/refman/8.4/en/replication-options-gtids.html

저장소 내 연관 노트:

- [복제·백업·PITR·failover 정리](../../interviews/database-deep-dive/09-replication-lag-backup-failover.md) — 복제 일반과 lag, RPO/RTO의 큰 그림
- [Debezium 트러블슈팅](../../troubleshooting/debezium.md) — 비-GTID 환경에서 CDC가 binlog 좌표로 동작하는 실제 로그
- [Flyway 마이그레이션](../migration/flyway/flyway.md) — DDL 마이그레이션 시 GTID 제약 점검 맥락
