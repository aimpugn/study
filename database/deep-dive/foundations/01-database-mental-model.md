# Database mental model foundations

## DB는 어떤 문제를 해결하는 시스템인가

데이터베이스를 처음 배울 때 가장 위험한 오해는 “결국 파일에 데이터를 저장하는 프로그램 아닌가?”라는 생각입니다. 이 말은 절반만 맞습니다. 데이터베이스도 결국 디스크나 SSD 위의 파일, 메모리 버퍼, 운영체제의 입출력 기능을 사용합니다. 하지만 DB가 해결하는 문제는 “값을 어딘가에 적는다”가 아니라 “여러 사람이 동시에 읽고 쓰는 중요한 상태를, 나중에 다시 믿을 수 있게 남긴다”입니다. 파일 하나에 JSON을 저장하는 방식도 처음에는 충분해 보입니다. 사용자가 한 명이고, 프로세스가 하나이고, 장애가 나도 다시 만들 수 있는 임시 데이터라면 실제로 그렇게 해도 됩니다. 문제는 데이터가 점점 서비스의 약속이 되는 순간부터 시작됩니다. 주문 상태, 결제 금액, 계좌 잔액, 재고 수량, 권한, 계약 이력처럼 틀리면 다시 계산하기 어렵거나 고객에게 바로 영향을 주는 값은 단순한 파일 쓰기만으로는 다루기 어렵습니다.

PostgreSQL 공식 튜토리얼은 관계형 데이터베이스를 “테이블의 모음”으로 소개하고, 각 테이블이 행과 열로 이루어진다고 설명합니다. MySQL 공식 문서는 MySQL을 SQL을 사용하는 관계형 데이터베이스 관리 시스템으로 설명합니다. 이 두 설명은 DB의 겉모습을 잡는 데 좋습니다. 하지만 공부할 때는 여기서 한 단계 더 내려가야 합니다. 테이블은 사람이 이해하는 논리적 모양이고, DBMS는 그 모양을 유지하려고 파일, 페이지, 인덱스, 로그, 락, 트랜잭션, 권한, 복구 절차를 함께 운영합니다. 그래서 DB는 “테이블 저장소”라기보다 “공유 상태를 안전하게 관리하는 운영체제에 가까운 서비스”로 읽는 편이 더 정확합니다.

가장 작은 예부터 보겠습니다. 회원 정보를 파일에 저장하는 프로그램을 생각해 봅니다.

```text
users.json
[
  {"id": 1, "email": "a@example.com", "point": 1000},
  {"id": 2, "email": "b@example.com", "point": 500}
]
```

이 파일에서 1번 사용자의 포인트를 100 차감하려면 프로그램은 보통 파일을 읽고, JSON을 파싱하고, 배열에서 `id = 1`인 원소를 찾고, `point`를 900으로 바꾼 뒤, 파일 전체를 다시 씁니다. 혼자 실행할 때는 자연스럽습니다. 하지만 웹 서비스에서는 같은 사용자의 포인트 차감 요청이 동시에 두 번 들어올 수 있습니다.

```text
처음 파일: id=1 point=1000

요청 A: 파일 읽음 -> point=1000 확인 -> 100 차감 예정
요청 B: 파일 읽음 -> point=1000 확인 -> 100 차감 예정

요청 A: point=900을 파일에 씀
요청 B: point=900을 파일에 씀

최종 파일: point=900
기대 결과: point=800
```

이 현상은 단순히 코드에 `if`가 빠져서 생긴 버그가 아닙니다. “읽은 값이 내가 쓸 때까지 그대로일 것”이라는 가정이 깨졌기 때문에 생긴 문제입니다. 파일 시스템은 바이트를 저장하고 읽게 해 주지만, 이 JSON 안의 `point`가 포인트 잔액이고, 두 요청이 같은 값을 동시에 바꾸면 안 되고, 차감 결과가 음수가 되면 안 된다는 도메인 규칙까지 알아서 지켜 주지 않습니다. 파일 잠금을 직접 걸 수는 있습니다. 하지만 파일 전체를 잠그면 동시에 읽고 쓰는 성능이 떨어지고, 프로세스가 죽었을 때 잠금 상태를 어떻게 복구할지, 일부만 쓴 파일을 어떻게 감지할지, 여러 파일을 함께 바꾸는 작업을 어떻게 원자적으로 만들지 같은 문제가 곧바로 따라옵니다.

DB는 바로 이 지점에서 등장합니다. DB가 제공하는 `UPDATE users SET point = point - 100 WHERE id = 1`은 단지 파일의 일부 문자열을 바꾸는 명령이 아닙니다. DB는 이 명령을 받으면 어떤 행을 찾을지 결정하고, 그 행을 바꾸는 동안 다른 트랜잭션과 충돌하지 않도록 조정하고, 변경 전후의 복구 정보를 로그에 남기고, 장애가 나면 어디까지 반영했고 어디부터 되돌려야 하는지 판단합니다. 이 모든 책임이 합쳐져야 “포인트가 100 줄었다”는 결과를 나중에도 믿을 수 있습니다.

```sql
CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    point integer NOT NULL CHECK (point >= 0)
);

UPDATE users
SET point = point - 100
WHERE id = 1 AND point >= 100;
```

위 SQL에는 파일 저장 코드보다 많은 의미가 들어 있습니다. `PRIMARY KEY`는 행을 식별하는 기준을 고정합니다. `UNIQUE`는 같은 이메일이 두 번 들어오지 않게 합니다. `NOT NULL`은 반드시 있어야 하는 값을 비워 둘 수 없게 합니다. `CHECK`는 포인트가 음수가 되지 않아야 한다는 규칙을 DB 안쪽에 둡니다. `UPDATE`의 `WHERE point >= 100`은 차감 가능한 상태에서만 변경하겠다는 조건입니다. 애플리케이션 코드가 실수해도 DB의 제약이 마지막 방어선이 됩니다. 파일 저장소에서는 이런 규칙이 코드 곳곳의 약속으로 흩어지지만, DB에서는 데이터에 가까운 곳에서 다시 검증됩니다.

여기서 중요한 mental model은 DB가 네 가지 책임을 한 몸으로 묶는다는 점입니다.

| 책임 | 파일만 쓸 때 개발자가 직접 떠안는 문제 | DB가 제공하는 대표 장치 |
|---|---|---|
| 저장 | 어떤 포맷으로 쓰고, 일부만 쓴 파일을 어떻게 감지할지 정해야 한다 | 테이블, row, page, WAL, checkpoint |
| 질의 | 필요한 값을 찾으려고 파일 전체를 읽거나 별도 색인을 직접 관리해야 한다 | SQL, optimizer, index, statistics |
| 동시성 | 동시에 읽고 쓰는 요청 사이의 충돌을 직접 막아야 한다 | transaction, lock, MVCC, isolation |
| 복구 | 프로세스나 서버가 죽었을 때 어디까지 믿을지 직접 판단해야 한다 | redo/undo log, crash recovery, backup |

이 네 가지는 따로따로 붙일 수 있는 편의 기능처럼 보이지만 실제로는 서로 얽혀 있습니다. 인덱스를 만들면 질의는 빨라지지만 쓰기 때 인덱스도 함께 갱신해야 합니다. 트랜잭션이 있으면 중간 상태를 다른 사용자에게 보여 주지 말아야 하므로 동시성 제어가 필요합니다. 복구 로그를 남기면 장애 뒤에는 데이터 파일과 로그를 서로 맞춰야 합니다. 제약 조건이 있으면 INSERT와 UPDATE 시점에 검사가 필요하고, 그 검사가 동시에 실행되는 다른 작업과 충돌하지 않게 해야 합니다. DB는 이 얽힘을 “애플리케이션 개발자가 매번 새로 구현하지 않아도 되는 공통 시스템”으로 만든 것입니다.

역사적으로도 DB는 단순 파일 저장을 더 예쁘게 포장하려고 나온 것이 아닙니다. 초기 업무 시스템은 파일 중심으로 데이터를 처리했습니다. 부서마다 파일이 따로 있고, 프로그램마다 자기 파일 형식을 알았습니다. 이 방식은 단일 배치 작업에는 괜찮았지만, 여러 프로그램이 같은 데이터를 다루기 시작하면 중복, 불일치, 동시성, 복구 문제가 커졌습니다. 관계형 모델은 데이터를 특정 프로그램의 파일 구조에 묶어 두지 않고, 관계와 제약으로 표현하려는 시도였습니다. SQL은 그 관계형 데이터에 “무엇을 원한다”를 선언적으로 말하는 언어가 되었습니다. DBMS는 그 선언을 실제 파일 접근, 인덱스 탐색, 로그 기록, 락 획득, 결과 반환으로 바꾸는 실행 시스템이 되었습니다.

파일 저장소와 DB의 차이를 더 손으로 따라가 보겠습니다. 쇼핑몰에서 주문을 생성하고 재고를 하나 줄이는 작업을 파일 두 개로 구현한다고 가정합니다.

```text
orders.json
[]

stocks.json
[
  {"product_id": 10, "quantity": 1}
]
```

두 요청이 동시에 같은 상품을 주문하면 다음 일이 벌어질 수 있습니다.

```text
T0  stocks.json quantity=1
T1  요청 A가 quantity=1을 읽음
T2  요청 B가 quantity=1을 읽음
T3  요청 A가 orders.json에 주문 A 추가
T4  요청 A가 stocks.json quantity=0 저장
T5  요청 B가 orders.json에 주문 B 추가
T6  요청 B가 stocks.json quantity=0 저장

결과:
- 주문은 2개 생김
- 재고는 0개로 보임
- 실제로는 재고 1개를 2번 팔았음
```

이 문제를 해결하려고 파일 락을 걸면 이번에는 장애 복구가 남습니다.

```text
T0  요청 A가 stocks.json lock 획득
T1  요청 A가 orders.json에 주문 A 추가
T2  프로세스가 죽음
T3  stocks.json은 아직 줄지 않음

결과:
- 주문은 생겼는데 재고는 줄지 않음
- 사용자는 주문 성공처럼 볼 수도 있음
- 재시도하면 중복 주문이 생길 수도 있음
```

DB에서는 이 작업을 하나의 트랜잭션으로 묶습니다.

```sql
BEGIN;

INSERT INTO orders(order_id, product_id, quantity)
VALUES (1001, 10, 1);

UPDATE stocks
SET quantity = quantity - 1
WHERE product_id = 10
  AND quantity >= 1;

COMMIT;
```

이 SQL 조각을 볼 때 `BEGIN`과 `COMMIT`은 단순한 문법 장식이 아닙니다. `BEGIN`은 “지금부터 여러 변경을 하나의 작업 단위로 보겠다”는 경계를 엽니다. `COMMIT`은 “이 작업 단위의 결과를 영구적으로 인정하겠다”는 지점입니다. 중간에 실패하면 DB는 작업 단위 전체를 버리거나, 적어도 인정할 수 있는 상태로 되돌릴 수 있어야 합니다. 그래서 트랜잭션은 저장과 복구, 동시성, 제약 검사를 동시에 끌고 옵니다.

실무에서 중요한 함정은 “DB를 쓰면 자동으로 안전하다”가 아니라 “DB가 제공하는 안전 장치를 어떤 경계에 걸었는지”입니다. 예를 들어 위 `UPDATE`가 실제로 한 행을 바꾸었는지 확인하지 않으면 재고가 부족한 주문도 성공처럼 처리될 수 있습니다.

```text
UPDATE stocks ... WHERE product_id=10 AND quantity >= 1

affected_rows = 1  -> 재고 차감 성공, 주문 확정 가능
affected_rows = 0  -> 재고 부족 또는 상품 없음, 주문 확정하면 안 됨
```

DB는 “한 행도 바뀌지 않았다”는 관측 가능한 신호를 줍니다. 애플리케이션은 그 신호를 읽고 다음 상태를 결정해야 합니다. DB의 제약과 트랜잭션은 개발자의 판단을 대체하지 않습니다. 대신 판단이 의존할 수 있는 강한 신호와 실패 경계를 제공합니다.

이 차이는 운영 장애에서도 드러납니다. 파일 저장 방식에서 주문과 재고 불일치가 생기면 보통 로그, 파일 수정 시각, 백업본, 애플리케이션 로그를 뒤져 사람이 원인을 추정해야 합니다. DB를 쓰면 적어도 트랜잭션 로그, 서버 로그, 쿼리 로그, lock wait, deadlock report, slow query, replication lag, backup recovery point 같은 관측 지점이 생깁니다. 물론 DB도 설정과 운영이 잘못되면 데이터를 잃을 수 있습니다. 하지만 좋은 DB 사용은 “문제가 절대 안 생긴다”가 아니라 “문제가 생겼을 때 어디를 보면 되는지, 어떤 단위로 되돌리거나 재시도할 수 있는지”를 만들어 줍니다.

```sql
-- PostgreSQL에서 현재 실행 중인 세션과 대기 이벤트를 보는 대표 관측 지점
SELECT pid, state, wait_event_type, wait_event, query
FROM pg_stat_activity
WHERE datname = current_database();

-- MySQL에서 현재 세션/프로세스 상태를 보는 대표 관측 지점
SHOW PROCESSLIST;
```

위 명령들은 DB를 “검은 상자”로 쓰지 않기 위한 최소 출발점입니다. PostgreSQL의 `pg_stat_activity`는 세션 상태와 대기 이벤트를 보여 줍니다. MySQL의 `SHOW PROCESSLIST`는 연결과 실행 중인 명령의 상태를 확인하는 데 쓰입니다. 이 명령을 안다고 해서 모든 장애를 해결할 수 있는 것은 아니지만, 파일 저장소와 달리 DB는 동시성 문제를 관찰할 수 있는 표면을 제공합니다. 포인트 차감이 멈췄는지, 락을 기다리는지, 느린 쿼리인지, 연결이 쌓였는지 같은 질문을 데이터로 좁혀 갈 수 있습니다.

DB가 제공하는 질의 능력도 단순 편의 기능이 아닙니다. 파일 저장 방식에서는 “어제 생성된 주문 중 금액이 10만 원 이상이고 아직 정산되지 않은 주문”을 찾으려면 파일 전체를 읽고 애플리케이션 코드로 필터링해야 합니다. 데이터가 커지면 매번 전체 파일을 읽는 비용이 커지고, 필터 조건이 다양해질수록 별도 색인을 직접 만들어야 합니다. DB에서는 같은 요구를 SQL로 표현합니다.

```sql
SELECT order_id, user_id, amount
FROM orders
WHERE created_at >= DATE '2026-05-18'
  AND created_at <  DATE '2026-05-19'
  AND amount >= 100000
  AND settlement_status = 'READY';
```

이 쿼리를 DB가 반드시 한 가지 방식으로 실행하는 것은 아닙니다. 작은 테이블이면 전체를 훑을 수 있고, `created_at` 인덱스가 있으면 날짜 범위를 먼저 좁힐 수 있고, 복합 인덱스가 있으면 상태와 날짜를 함께 좁힐 수 있습니다. 애플리케이션은 “어떤 조건의 행을 원한다”고 말하고, DB는 통계와 인덱스를 보고 접근 경로를 고릅니다. 이 분리가 있기 때문에 데이터가 커져도 코드를 매번 새 파일 탐색 알고리즘으로 바꾸지 않고, 인덱스와 통계와 쿼리 형태를 조정할 수 있습니다.

여기서도 함정이 있습니다. SQL을 쓰면 모두 빠를 것이라고 생각하면 안 됩니다. DB는 선언적 질의를 실행 계획으로 바꾸지만, 잘못된 인덱스, 오래된 통계, 조건식의 형태, 타입 변환, collation, 함수 사용 때문에 예상과 다른 계획을 고를 수 있습니다. 그래서 DB를 제대로 쓴다는 것은 SQL 문법을 아는 데서 끝나지 않습니다. `EXPLAIN`으로 계획을 보고, 실제 실행 시간과 읽은 행 수를 관측하고, 인덱스가 어떤 비교 규칙을 타는지 확인해야 합니다.

```sql
EXPLAIN
SELECT order_id, user_id, amount
FROM orders
WHERE created_at >= DATE '2026-05-18'
  AND created_at <  DATE '2026-05-19'
  AND settlement_status = 'READY';
```

검증 기준은 구체적이어야 합니다. PASS는 “원하는 기간과 상태 조건을 만족하는 행만 반환하고, 계획이 전체 테이블을 불필요하게 크게 훑지 않는다는 근거가 보인다”입니다. FAIL은 “결과는 맞아 보이지만 실행 계획이 대부분의 행을 읽고 있거나, 타입 변환 때문에 인덱스를 타지 못하거나, 시간이 데이터 증가와 함께 급격히 늘어나는 신호가 보인다”입니다. 파일 저장소에서는 이런 질문을 애플리케이션 코드와 파일 구조 수준에서 직접 만들어야 하지만, DB는 계획과 통계라는 관측 가능한 언어를 제공합니다.

DB의 복구 책임은 특히 과소평가되기 쉽습니다. 파일에 `write()`를 호출했다고 해서 데이터가 영구적으로 안전하다는 뜻은 아닙니다. 운영체제는 성능을 위해 쓰기를 메모리에 모아 두었다가 나중에 디스크에 반영할 수 있습니다. 프로세스가 죽는 경우와 머신 전원이 나가는 경우도 다릅니다. 파일 일부만 쓰인 상태, 디렉터리 엔트리만 바뀐 상태, 임시 파일 rename 도중의 상태를 모두 고려해야 합니다. DB는 이런 문제를 피하려고 로그를 먼저 쓰는 방식, 체크포인트, flush 정책, page checksum, recovery 절차를 둡니다. 세부 구현은 PostgreSQL, MySQL/InnoDB, SQLite, Oracle마다 다르지만, 공통 질문은 같습니다. “장애 직전 사용자가 성공으로 본 변경은 장애 뒤에도 남아 있는가?” “성공으로 인정하지 않은 중간 변경은 장애 뒤에 사라지는가?”

```text
트랜잭션 T가 point=1000 -> 900 변경

1. 변경 의도를 로그에 기록한다.
2. 데이터 page는 메모리에서 먼저 바뀔 수 있다.
3. COMMIT이 성공하려면 DB가 정한 내구성 조건을 만족해야 한다.
4. 장애 뒤 recovery가 로그와 page를 비교한다.
5. commit된 변경은 다시 적용하고, commit되지 않은 변경은 인정하지 않는다.
```

이 trace는 실제 DB별 recovery 알고리즘의 모든 세부를 말하지는 않습니다. 하지만 첫 mental model로는 충분히 중요합니다. DB는 단지 현재 파일 내용을 보는 것이 아니라, “어떤 변경을 성공으로 인정했는가”라는 기록을 함께 봅니다. 그래서 WAL이나 redo log 같은 이름이 등장합니다. 로그는 디버깅용 텍스트 로그가 아니라, 장애 뒤 데이터 파일을 믿을 수 있는 상태로 다시 만들기 위한 복구 자료입니다.

애플리케이션 개발자에게 DB mental model이 필요한 이유는 실제 버그가 DB 바깥과 안쪽 경계에서 많이 생기기 때문입니다. 예를 들어 개발자가 다음처럼 구현했다고 가정합니다.

```text
1. 애플리케이션에서 현재 포인트 SELECT
2. Java/Kotlin 코드에서 point >= 100 검사
3. point - 100 값을 계산
4. UPDATE users SET point = 계산값 WHERE id = 1
```

이 코드는 단일 요청에서는 맞아 보입니다. 하지만 두 요청이 동시에 실행되면 둘 다 같은 `point=1000`을 읽고 같은 `point=900`을 쓸 수 있습니다. DB를 쓰고 있어도 파일 저장소에서 봤던 lost update 문제가 다시 나타납니다. 더 안전한 형태는 계산과 조건을 DB 안의 하나의 `UPDATE`로 밀어 넣는 것입니다.

```sql
UPDATE users
SET point = point - 100
WHERE id = 1
  AND point >= 100;
```

이렇게 쓰면 DB는 해당 행의 현재 상태를 기준으로 조건과 변경을 함께 처리합니다. 여전히 isolation level과 DB별 동작을 이해해야 하지만, 적어도 “읽고 계산하고 다시 쓰는 사이”가 애플리케이션 코드 바깥에 길게 열려 있지는 않습니다. 이 차이를 모르면 DB를 쓰면서도 파일 저장소처럼 사고하게 됩니다.

관계형 DB가 항상 정답이라는 뜻은 아닙니다. 로그 수집, 캐시, 검색, 대용량 이벤트 스트림, 분석용 컬럼 저장소, 문서 저장소는 각자 다른 trade-off를 갖습니다. 하지만 그 대안들도 결국 저장, 질의, 동시성, 복구 중 어떤 책임을 어디로 옮기는지로 비교해야 합니다. Redis를 캐시로 쓰면 빠른 읽기를 얻는 대신 영속성과 복구 경계를 다시 설계해야 합니다. Elasticsearch/OpenSearch를 검색에 쓰면 전문 검색과 inverted index를 얻는 대신 원장성 있는 쓰기와 트랜잭션 경계를 RDBMS와 다르게 봐야 합니다. 파일과 DB의 차이를 모르면 이런 도구 비교도 “빠르다/느리다” 정도로만 남습니다.

그래서 이 섹션의 핵심 판단은 간단합니다. DB는 파일보다 큰 저장소가 아니라, 공유 상태의 신뢰성을 관리하는 시스템입니다. 테이블, SQL, 인덱스, 트랜잭션, 로그, 락, 복구, 권한은 각각 독립된 잡기술이 아니라 같은 문제를 나눠 맡은 장치입니다. 실무에서 DB를 잘 쓴다는 것은 이 장치들이 어느 책임을 맡는지 알고, 애플리케이션이 어느 신호를 확인해야 하는지 아는 것입니다.

마지막으로 직접 검증할 수 있는 작은 실험을 남깁니다. 실제 DB가 없어도 손으로 먼저 따라갈 수 있습니다.

```text
실험 A: 파일 사고 실험
초기 point=1000
요청 A와 B가 동시에 100 차감
각 요청이 read -> calculate -> write를 따로 수행한다고 놓고 최종값을 적어 본다.
PASS: lost update가 왜 point=900으로 남는지 설명할 수 있다.
FAIL: "두 번 실행했으니 당연히 800"이라고 말하면서 read 시점과 write 시점을 분리하지 못한다.

실험 B: SQL 사고 실험
초기 point=1000
UPDATE users SET point = point - 100 WHERE id=1 AND point >= 100;
두 요청이 같은 행을 바꿀 때 DB가 row 단위 충돌을 조정해야 함을 설명한다.
PASS: affected_rows를 주문 확정 조건으로 써야 함을 말할 수 있다.
FAIL: UPDATE를 보냈다는 사실만으로 비즈니스 성공을 선언한다.

실험 C: 복구 사고 실험
orders insert 뒤 stocks update 전에 프로세스가 죽었다.
BEGIN/COMMIT이 없을 때와 있을 때 사용자가 믿을 수 있는 상태가 어떻게 달라지는지 적어 본다.
PASS: 트랜잭션 경계가 "여러 변경을 하나의 인정 단위로 묶는다"는 점을 설명할 수 있다.
FAIL: 로그 파일이나 백업만 있으면 자동으로 일관성이 맞는다고 생각한다.
```

이 실험의 목적은 DB 내부 구현을 다 외우는 것이 아닙니다. 첫 벽돌을 확실히 붙잡는 것입니다. 파일은 바이트를 저장합니다. DB는 여러 actor가 공유하는 상태에 대해 저장, 질의, 동시성, 복구 책임을 함께 제공합니다. 이후 인덱스, WAL, MVCC, isolation, optimizer를 배울 때도 이 첫 문장을 잃지 않으면 세부 기능을 흩어진 암기 목록이 아니라 하나의 시스템으로 다시 조립할 수 있습니다.

공식 근거로 다시 연결하면, PostgreSQL의 튜토리얼은 테이블·행·열이라는 관계형 데이터의 기본 표면을 보여 주고, MySQL 문서는 MySQL을 SQL 기반 관계형 데이터베이스 관리 시스템으로 설명합니다. 이 문서에서 더해 둔 부분은 그 공식 표면이 실무에서 왜 필요한지에 대한 해석입니다. 여러 사용자가 동시에 중요한 값을 바꾸고, 장애 뒤에도 성공한 변경을 믿어야 하는 상황에서는 테이블이라는 논리 표면 뒤에 동시성 제어와 복구 체계가 반드시 따라와야 합니다. 이 해석은 특정 DBMS 하나에만 묶이지 않는 일반 원리이지만, 실제 세부는 항상 사용하는 DB의 공식 문서와 실험으로 다시 확인해야 합니다.

파일 저장소와 DB의 차이를 더 분명히 보려면 “파일에 기능을 하나씩 붙이면 DB가 되는가”라는 질문을 해 보면 좋습니다. 처음에는 파일 저장에 잠금만 붙이면 충분해 보입니다. 프로세스가 JSON 파일을 수정하기 전에 운영체제 파일 락을 잡고, 수정이 끝나면 락을 푸는 식입니다. 이 방법은 동시에 두 프로세스가 같은 파일을 덮어쓰는 문제를 줄입니다. 하지만 곧바로 세 가지 질문이 남습니다. 첫째, 읽기만 하는 요청도 락을 기다려야 하는가. 둘째, 파일 하나가 아니라 주문 파일과 재고 파일을 함께 바꾸면 락 순서를 어떻게 정할 것인가. 셋째, 락을 잡은 프로세스가 죽었을 때 중간 변경을 누가 복구할 것인가. 이 질문에 답하려고 read lock, write lock, lock timeout, deadlock detection, atomic rename, temp file, journal, checksum, backup을 붙이다 보면 어느 순간 저장소가 아니라 작은 DBMS를 만들고 있습니다.

```text
직접 만든 파일 저장소에 기능을 붙이는 흐름

1. users.json을 안전하게 쓰고 싶다.
   -> temp file + rename을 쓴다.

2. 동시에 두 요청이 덮어쓰면 안 된다.
   -> file lock을 둔다.

3. orders.json과 stocks.json을 같이 바꿔야 한다.
   -> 여러 파일 lock 순서와 작업 단위가 필요하다.

4. 중간에 죽으면 어디까지 반영됐는지 알아야 한다.
   -> journal 또는 log가 필요하다.

5. 특정 사용자만 빨리 찾고 싶다.
   -> 별도 index 파일이 필요하다.

6. index와 data 파일이 서로 어긋나면 안 된다.
   -> data/index/log를 함께 복구하는 규칙이 필요하다.
```

이 trace의 마지막에 도착한 구조가 바로 DB가 이미 제공하는 세계입니다. DB가 “복잡한 도구”라서 어려운 것이 아니라, 원래 문제가 복잡하기 때문에 DB가 복잡해진 것입니다. 개발자가 이 복잡도를 애플리케이션 코드 곳곳에 흩어 놓느냐, 검증된 DBMS의 책임으로 모으느냐가 차이입니다. 물론 DBMS도 모든 문제를 공짜로 해결하지 않습니다. 트랜잭션 범위를 잘못 잡으면 락 경합이 커지고, 인덱스를 잘못 만들면 쓰기 비용이 커지고, 격리 수준을 잘못 이해하면 이상 현상이 생기고, 백업을 검증하지 않으면 복구할 수 없는 백업만 쌓입니다. 하지만 최소한 문제의 이름과 관측 지점과 검증 방법이 있습니다.

관계형 DB의 “관계”도 이 맥락에서 봐야 합니다. 테이블은 엑셀 시트처럼 행과 열을 보여 주지만, DB의 테이블은 단순한 격자 UI가 아닙니다. 테이블은 어떤 종류의 사실을 어떤 속성으로 기록할지 정한 계약입니다. `users` 테이블은 사용자의 식별자와 이메일과 포인트를 기록합니다. `orders` 테이블은 주문이라는 사건을 기록합니다. `stocks` 테이블은 상품별 재고 상태를 기록합니다. 이 세 테이블이 따로 존재할 때 더 중요한 것은 서로의 관계입니다. 주문의 `user_id`가 실제 사용자와 연결되는지, 주문의 `product_id`가 실제 상품과 연결되는지, 재고 수량이 음수가 되지 않는지 같은 규칙은 애플리케이션 메모리 안의 임시 판단이 아니라 데이터의 장기 계약입니다.

```sql
CREATE TABLE products (
    product_id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE orders (
    order_id bigint PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id),
    product_id bigint NOT NULL REFERENCES products(product_id),
    quantity integer NOT NULL CHECK (quantity > 0)
);
```

외래 키와 체크 제약을 쓰면 DB는 “없는 사용자에게 주문이 달리는 상태”나 “수량이 0인 주문” 같은 오류를 데이터 경계에서 막습니다. 파일 저장 방식에서는 이런 검사가 저장 코드마다 반복됩니다. 한 API에서는 검사하고, 배치 작업에서는 빠뜨리고, 운영자가 수동 보정 스크립트를 돌릴 때 또 빠뜨리는 식으로 drift가 생깁니다. DB 제약은 그런 drift를 줄입니다. 이 말은 모든 비즈니스 규칙을 DB 제약으로만 표현하라는 뜻은 아닙니다. 할인 정책, 배송 상태 전이, 외부 결제 승인 같은 규칙은 애플리케이션 서비스와 도메인 코드가 더 잘 표현할 수 있습니다. 하지만 데이터가 절대 가져서는 안 되는 기본 모순은 DB 근처에 두는 편이 안전합니다.

DB가 해결하는 문제에는 “시간”도 포함됩니다. 파일을 열어 현재 내용을 보면 지금 상태는 알 수 있지만, 어떤 요청이 어떤 순서로 성공했는지, 중간 실패를 어떻게 처리했는지, 장애 뒤 어느 시점으로 돌아갈 수 있는지는 별도 장치가 필요합니다. DB는 트랜잭션 로그, commit 순서, snapshot, backup, point-in-time recovery 같은 형태로 시간 축을 다룹니다. 이 시간 축이 없으면 운영 사고가 났을 때 “현재 파일이 이렇게 생겼다”만 남고, 왜 그렇게 되었는지 추적하기 어렵습니다.

```text
시간을 포함한 주문 처리 trace

T1 BEGIN
T2 INSERT orders(order_id=1001)
T3 UPDATE stocks(product_id=10, quantity=0)
T4 COMMIT 성공
T5 서버 장애
T6 recovery

복구 판단:
- T4까지 commit된 변경은 남아야 한다.
- T4 전에 commit되지 않은 다른 트랜잭션의 중간 변경은 남으면 안 된다.
- index는 orders/stocks의 최종 상태와 맞아야 한다.
```

이 시간 모델은 “백업이 있으니 괜찮다”는 말보다 강합니다. 백업은 특정 시점의 사본입니다. 로그와 복구 절차는 특정 시점 이후 성공한 변경을 어떻게 재구성할지 설명합니다. 운영에서 정말 필요한 질문은 “백업 파일이 있는가”가 아니라 “어느 시점까지 복구할 수 있고, 복구된 데이터가 애플리케이션의 불변식과 맞는가”입니다. DB를 공부할 때 backup, WAL/binlog, replication, checkpoint를 나중에 따로 배우더라도, 여기서 먼저 이 질문을 잡아 두면 세부 기능이 왜 필요한지 잊지 않습니다.

또 하나의 중요한 차이는 DB가 실패를 표현하는 방식입니다. 파일 저장 코드에서는 실패가 보통 예외, 반환값, 깨진 파일, 불완전한 로그로 나타납니다. DB에서는 실패가 더 구조화된 신호로 드러납니다. unique constraint violation은 중복된 값을 넣으려 했다는 뜻입니다. foreign key violation은 참조 대상이 없다는 뜻입니다. deadlock detected는 두 트랜잭션이 서로 필요한 잠금을 기다리다가 DB가 하나를 희생시켰다는 뜻입니다. lock timeout은 필요한 리소스를 정해진 시간 안에 얻지 못했다는 뜻입니다. serialization failure는 더 높은 격리 수준에서 동시에 실행된 트랜잭션을 모두 성공으로 인정하면 일관성이 깨질 수 있다는 뜻입니다. 이 신호들은 단순한 에러 문자열이 아니라 재시도, 사용자 안내, 보상 처리, 운영 알림의 기준이 됩니다.

```text
DB 실패 신호를 애플리케이션 결정으로 바꾸는 예

unique violation on users.email
  -> 이미 가입된 이메일 안내 또는 idempotent success 검토

affected_rows = 0 on stock decrement
  -> 재고 부족으로 주문 실패 처리

deadlock detected
  -> 같은 트랜잭션을 짧은 backoff 뒤 재시도

connection timeout
  -> DB 작업 성공 여부가 불명확할 수 있으므로 idempotency key로 재조회
```

이 표면을 이해하지 못하면 DB 에러를 모두 “500”으로 뭉개거나, 반대로 모든 에러를 무조건 재시도해서 더 큰 장애를 만듭니다. 시니어 개발자는 DB가 돌려준 실패 신호를 보고 “이 실패는 입력이 잘못된 것인가, 동시에 실행된 작업과 충돌한 것인가, DB가 바빠서 관측이 불확실한 것인가, 이미 성공했지만 응답만 잃은 것인가”를 구분하려고 합니다. DB mental model은 이런 운영 판단까지 포함합니다.

파일 저장소와 DB를 비교할 때 성능도 자주 오해합니다. 작은 파일 하나를 읽는 것은 DB 쿼리보다 빠를 수 있습니다. 그래서 “파일이 더 빠르다”는 결론을 내리기 쉽습니다. 하지만 성능 질문은 항상 조건을 포함해야 합니다. 데이터가 100행인지 1억 행인지, 동시에 몇 명이 쓰는지, 어떤 조건으로 찾는지, 장애 뒤 재처리가 필요한지, 인덱스를 유지해야 하는지, 결과의 최신성이 얼마나 중요한지에 따라 답이 바뀝니다. DB는 단일 파일 읽기보다 항상 빠른 도구가 아니라, 다양한 질의와 동시 쓰기와 복구를 함께 고려할 때 총 실패 비용을 낮추는 시스템입니다.

```text
성능 질문을 바꾸는 조건

질문이 약한 형태:
"JSON 파일이 빠른가, DB가 빠른가?"

질문이 강한 형태:
"사용자 100만 명의 포인트를 여러 API와 배치가 동시에 바꾸고,
장애 뒤 성공한 차감은 보존해야 하며,
사용자 id/email/최근 변경일로 조회해야 할 때,
어떤 저장 구조와 동시성/복구 경계가 필요한가?"
```

두 번째 질문으로 바꾸면 DB가 왜 필요한지 더 잘 보입니다. 파일은 여전히 일부 상황에서 좋은 선택입니다. 설정 파일, 빌드 산출물, append-only 로컬 로그, 임시 캐시, 작은 개인 도구의 상태 저장에는 파일이 더 단순하고 충분할 수 있습니다. 반대로 여러 actor가 같은 상태를 바꾸고, 상태의 모순을 막아야 하고, 장애 뒤에도 성공 여부를 설명해야 한다면 DB가 제공하는 공통 장치를 검토해야 합니다. 선택 기준은 도구 이름이 아니라 책임의 위치입니다.

이 섹션을 마무리하기 전에 작은 기준 문장을 남깁니다. 어떤 데이터를 DB에 넣을지 파일에 둘지 고민할 때 아래 질문에 답해 봅니다.

```text
1. 이 값이 틀리면 고객, 돈, 권한, 정산, 재고, 감사에 영향을 주는가?
2. 둘 이상의 요청, 배치, 운영자가 같은 값을 바꾸는가?
3. 같은 값에 대해 "없는 값", "중복 값", "음수", "참조 대상 없음" 같은 모순을 막아야 하는가?
4. 장애 뒤 어느 변경이 성공했고 어느 변경이 실패했는지 설명해야 하는가?
5. 여러 조건으로 조회하거나 정렬하거나 집계해야 하는가?
6. 문제 발생 시 대기, 락, 느린 쿼리, 실패 코드를 관측해야 하는가?
```

대부분의 답이 “예”라면 DB가 필요한 가능성이 큽니다. 대부분의 답이 “아니오”라면 파일이나 더 단순한 저장소가 맞을 수 있습니다. 이 질문들은 특정 제품을 고르는 체크리스트가 아니라, DB가 해결하는 문제의 윤곽을 잡는 도구입니다. 이후 어떤 DBMS를 고르든 이 책임들이 어디에 있는지 계속 확인해야 합니다.

한 가지 더 붙이면, DB는 “데이터를 한곳에 모으는 권위”도 제공합니다. 파일이 여러 서버에 흩어져 있으면 어떤 파일이 최신인지, 어느 복사본이 성공한 쓰기를 포함하는지, 누가 어떤 버전을 읽었는지 설명하기 어렵습니다. DB를 쓰면 적어도 클라이언트는 정해진 endpoint와 transaction boundary를 통해 상태를 읽고 씁니다. 운영자는 그 endpoint의 연결 수, 트랜잭션 수, replication 상태, 백업 시점, 권한 부여 이력을 관찰합니다. 이 권위가 없으면 서비스는 점점 “어딘가의 파일을 누군가 수정했을 것”이라는 추정으로 운영됩니다. DB의 권위는 중앙 집중만 뜻하지 않습니다. 분산 DB나 복제 구조에서도 어떤 노드가 쓰기 권한을 갖는지, 어떤 로그 순서가 진실인지, 읽기 replica가 얼마나 늦을 수 있는지 같은 규칙이 필요합니다.

```text
권위 없는 파일 복사본

server-a/users.json: point=900, updated_at=10:01
server-b/users.json: point=1000, updated_at=10:00
server-c/users.json: point=800, updated_at=10:02

질문:
- 어떤 값이 진짜인가?
- 10:02의 point=800은 어떤 요청 두 개를 포함하는가?
- server-b가 뒤늦게 1000을 다시 배포하면 어떻게 막는가?
```

DB를 쓴다고 이 질문이 사라지지는 않습니다. primary/replica, replication lag, failover, split-brain 같은 더 어려운 문제가 생길 수 있습니다. 하지만 DB 세계에서는 이 문제들이 이름을 갖습니다. primary가 쓰기 권위를 갖는지, consensus가 commit 순서를 정하는지, replica read가 stale해도 되는지, backup restore 뒤 application idempotency를 어떻게 확인할지 같은 식으로 논의할 수 있습니다. 파일 복사본이 흩어진 상태보다 훨씬 관측 가능하고 검증 가능한 질문이 됩니다.

DB를 단순 저장소로 보면 권한과 감사도 놓칩니다. 파일은 운영체제 권한으로 읽고 쓰기를 제한할 수 있지만, “이 사용자는 `orders`의 금액은 볼 수 있지만 개인정보 컬럼은 볼 수 없다”, “이 배치 계정은 INSERT만 하고 DELETE는 못 한다”, “누가 언제 권한을 바꾸었는지 남긴다” 같은 세밀한 데이터 경계는 직접 구현해야 합니다. DB는 사용자, 역할, 권한, view, row-level security, audit extension 같은 장치를 통해 데이터 접근 자체를 관리 대상으로 만듭니다. 이 기능들은 보안 단원에서 다시 깊게 보겠지만, mental model 단계에서도 DB가 “값 저장”뿐 아니라 “값에 접근하는 행위의 경계”까지 다룬다는 점을 기억해야 합니다.

마지막 실무 함정은 “DB에 넣었으니 애플리케이션은 상태를 생각하지 않아도 된다”입니다. DB는 강한 경계를 제공하지만, 애플리케이션이 transaction을 너무 넓게 잡으면 사용자가 외부 API 응답을 기다리는 동안 DB 락을 오래 붙잡을 수 있습니다. 반대로 transaction을 너무 잘게 쪼개면 주문 생성은 commit됐는데 결제 요청 실패를 처리하지 못하는 중간 상태가 생길 수 있습니다. DB는 상태를 안전하게 바꿀 수 있는 도구를 제공하고, 애플리케이션은 어떤 상태 전이가 하나의 transaction이어야 하는지, 어떤 외부 부작용은 outbox나 재시도 큐로 분리해야 하는지 결정합니다. DB mental model의 최종 목표는 DB 안쪽을 숭배하는 것이 아니라, 애플리케이션과 DB가 나눠 가져야 할 책임을 정확히 보는 것입니다.

이 책임 분리는 테스트에서도 드러납니다. DB를 mocking해서 repository 메서드가 호출됐는지만 확인하면 “SQL을 보냈다”는 사실은 검증할 수 있지만, unique 제약, 외래 키, affected row, transaction rollback, NULL 비교, collation 비교, lock conflict 같은 DB 고유의 의미는 검증하지 못합니다. 반대로 모든 테스트를 실제 DB 통합 테스트로만 만들면 느리고 깨지기 쉬울 수 있습니다. 좋은 기준은 단순합니다. 애플리케이션의 분기와 계산은 단위 테스트로 빠르게 확인하고, DB가 책임지는 계약은 실제 DB 또는 그에 준하는 환경에서 작은 통합 테스트로 확인합니다. 예를 들어 포인트 차감은 서비스 단위 테스트에서 “affected_rows=0이면 재고 부족으로 본다”를 확인하고, DB 통합 테스트에서 `CHECK(point >= 0)`과 동시 UPDATE가 실제로 어떤 결과를 내는지 확인합니다.

```text
테스트 책임 분리

애플리케이션 단위 테스트:
  given affected_rows = 0
  when 주문 확정 로직 실행
  then 재고 부족 결과를 반환한다.

DB 통합 테스트:
  given users.point = 50
  when UPDATE users SET point = point - 100 WHERE id=1 AND point >= 100
  then affected_rows = 0이고 point는 50으로 남는다.
```

이렇게 보면 DB는 테스트하기 어려운 외부 의존성이 아니라, 반드시 검증해야 하는 계약의 일부입니다. 파일 저장소를 쓸 때도 같은 질문은 남지만, DB는 제약과 트랜잭션과 관측 명령을 통해 그 계약을 더 명확한 형태로 드러냅니다.

이 기준을 기억하면 “DB를 왜 배워야 하는가”라는 질문도 좁아집니다. DB를 배우는 일은 제품별 명령어를 많이 외우는 일이 아니라, 데이터가 틀리면 안 되는 순간에 어떤 책임을 어느 층에 둘지 판단하는 훈련입니다. 파일, 캐시, 메시지 큐, 검색 엔진, RDBMS를 비교할 때도 같은 질문을 씁니다. 누가 쓰기 권위를 갖는가, 어떤 단위가 원자적으로 성공하는가, 실패 신호는 무엇인가, 복구 뒤 어떤 상태를 믿을 수 있는가, 성능 문제를 어디서 관측하는가. 이 질문에 답할 수 있으면 DB는 더 이상 “백엔드가 붙이는 저장소”가 아니라 서비스의 상태 신뢰성을 세우는 핵심 시스템으로 보입니다.

이 첫 기준이 잡히면 이후 단원에서 만나는 page, index, transaction, isolation, replication 같은 단어는 모두 같은 질문의 세부 답으로 배치됩니다. page는 값을 어떤 물리 단위로 읽고 쓰는지, index는 어떤 조건을 빨리 좁히는지, transaction은 어떤 변경 묶음을 성공으로 인정하는지, isolation은 동시에 실행된 작업을 어떤 순서처럼 보이게 할지, replication은 성공한 변경을 다른 노드로 어떻게 전달할지에 대한 답입니다. 이 연결을 놓치지 않는 것이 중요합니다.

### 마지막 실무 연결: DB를 선택한다는 것은 책임을 위임한다는 뜻이다

파일에 JSON을 쓰는 방식도 작은 서비스에서는 충분할 수 있다. 하지만 여러 요청이 동시에 같은 값을 바꾸고, 장애 뒤에도 어떤 변경이 성공했는지 설명해야 하며, 조건 검색과 권한과 백업과 복구가 필요해지는 순간 이야기가 달라진다. DB를 선택한다는 것은 단순 저장소를 고르는 일이 아니라 동시성, 복구, 질의, 무결성 책임을 검증 가능한 시스템에 위임하는 일이다.

그래서 설계 리뷰에서 첫 질문은 “DB를 쓰는가”가 아니라 “어떤 책임을 DB에 맡기고 어떤 책임을 애플리케이션에 남기는가”여야 한다. 중복 요청 방지는 unique key와 idempotency table이 맡을 수 있고, 금액 합계 보존은 constraint와 원장 검증이 도울 수 있으며, 외부 API 부작용은 DB 트랜잭션만으로 닫히지 않으므로 outbox나 reconciliation이 필요하다. 이 분담을 말할 수 있어야 DB mental model이 실제 시스템 설계로 이어진다.


## 논리 모델과 물리 실행 모델을 분리해 읽기

SQL을 공부할 때 두 번째로 큰 함정은 “SQL에 적힌 순서가 DB가 실제로 일하는 순서”라고 믿는 것입니다. `SELECT ... FROM ... WHERE ... ORDER BY ...`라고 쓰니 DB가 먼저 `SELECT` 목록을 계산하고, 그다음 `FROM`을 보고, 그다음 `WHERE`를 적용한다고 생각하기 쉽습니다. 하지만 SQL은 기본적으로 결과를 선언하는 언어입니다. 사용자는 “어떤 결과가 필요하다”를 논리적으로 말하고, DB는 그 결과를 만들기 위해 실제 파일과 인덱스와 메모리를 어떤 순서로 접근할지 결정합니다. 이 둘을 분리하지 못하면 실행 계획을 읽을 때 계속 헷갈립니다.

PostgreSQL의 `SELECT` 공식 문서는 `WITH`, `FROM`, `WHERE`, `GROUP BY`, `HAVING`, `SELECT` 출력식, `DISTINCT`, 집합 연산, `ORDER BY`, `LIMIT/OFFSET`, 잠금 절 같은 논리적 처리 단계를 설명합니다. 이 순서는 SQL 결과가 어떤 의미로 만들어지는지 이해하는 데 중요합니다. 반면 PostgreSQL의 `EXPLAIN` 문서는 쿼리를 실행하기 위해 생성된 실행 계획을 보여 준다고 설명합니다. 실행 계획에는 sequential scan, index scan, join, sort, aggregate 같은 물리 연산이 나옵니다. 같은 SQL이라도 테이블 크기, 인덱스, 통계, 파라미터 값, 설정에 따라 실행 계획이 달라질 수 있습니다.

작은 테이블부터 보겠습니다.

```sql
CREATE TABLE orders (
    order_id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    status text NOT NULL,
    amount integer NOT NULL,
    created_at date NOT NULL
);

INSERT INTO orders(order_id, user_id, status, amount, created_at) VALUES
  (1, 10, 'READY', 120000, DATE '2026-05-18'),
  (2, 10, 'PAID',   50000, DATE '2026-05-18'),
  (3, 20, 'READY',  30000, DATE '2026-05-19'),
  (4, 30, 'READY', 200000, DATE '2026-05-18');
```

이제 다음 질의를 봅니다.

```sql
SELECT user_id, count(*) AS ready_count, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 100000
ORDER BY total_amount DESC;
```

논리 모델에서 이 쿼리는 다음 흐름으로 읽습니다.

```text
원본 orders
  -> FROM orders로 후보 행 집합을 만든다.
  -> WHERE status='READY' AND created_at='2026-05-18'로 행을 줄인다.
  -> GROUP BY user_id로 같은 user_id끼리 묶는다.
  -> count(*), sum(amount)를 그룹마다 계산한다.
  -> HAVING sum(amount) >= 100000으로 그룹을 줄인다.
  -> SELECT user_id, ready_count, total_amount 모양으로 결과를 만든다.
  -> ORDER BY total_amount DESC로 결과 행을 정렬한다.
```

손으로 따라가면 이렇게 됩니다.

| 단계 | 남는 행 또는 그룹 | 설명 |
|---|---|---|
| FROM | 1, 2, 3, 4 | 테이블의 논리적 후보 전체 |
| WHERE | 1, 4 | READY이면서 2026-05-18인 행만 남음 |
| GROUP BY | user_id=10: row 1 / user_id=30: row 4 | 사용자별 그룹 |
| aggregate | user_id=10 total=120000 count=1 / user_id=30 total=200000 count=1 | 그룹별 계산 |
| HAVING | 두 그룹 모두 남음 | 각 total이 100000 이상 |
| SELECT | `(10,1,120000)`, `(30,1,200000)` | 출력 컬럼 이름과 값 결정 |
| ORDER BY | `(30,1,200000)`, `(10,1,120000)` | total_amount 내림차순 |

이 표는 “SQL의 의미”를 설명합니다. DB가 실제로 디스크에서 행 1, 2, 3, 4를 꼭 이 순서로 읽었다는 뜻은 아닙니다. 테이블에 인덱스가 없다면 DB는 테이블 전체를 훑으면서 조건을 평가할 수 있습니다. `status, created_at` 복합 인덱스가 있다면 인덱스로 `READY`와 날짜 조건에 맞는 row 위치를 먼저 좁힐 수 있습니다. 통계상 `created_at` 조건이 더 선택적이면 날짜 인덱스를 먼저 볼 수도 있습니다. 결과는 같아야 하지만 물리 경로는 달라질 수 있습니다.

```sql
CREATE INDEX orders_status_created_at_idx
ON orders(status, created_at);

EXPLAIN
SELECT user_id, count(*) AS ready_count, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 100000
ORDER BY total_amount DESC;
```

실제 출력은 DB 버전, 통계, 데이터 크기에 따라 달라지지만, 물리 실행 계획은 대략 이런 종류의 노드로 읽습니다.

```text
Sort
  Sort Key: (sum(amount)) DESC
  -> HashAggregate
       Group Key: user_id
       Filter: (sum(amount) >= 100000)
       -> Index Scan using orders_status_created_at_idx on orders
            Index Cond: (status = 'READY' AND created_at = '2026-05-18')
```

이 계획을 논리 순서와 나란히 놓으면 분리가 선명해집니다.

| 논리 단계 | 물리 계획에서 보일 수 있는 연산 | 같은 것인가 |
|---|---|---|
| FROM/WHERE | Seq Scan, Index Scan, Bitmap Heap Scan, Filter, Index Cond | 논리 조건을 만족하는 행을 찾는 물리 방법 |
| GROUP BY | HashAggregate, GroupAggregate | 그룹을 만드는 물리 알고리즘 |
| HAVING | Aggregate 노드의 Filter | 그룹 계산 뒤 버리는 조건 |
| SELECT | 출력 projection | 필요한 컬럼과 식을 만드는 단계 |
| ORDER BY | Sort, Incremental Sort, index order 이용 | 결과 순서를 맞추는 방법 |

“같은 것인가”라는 질문에 대한 답은 “대응하지만 같은 것은 아니다”입니다. `WHERE`는 논리적으로 어떤 행이 남아야 하는지를 말합니다. `Index Cond`는 그 조건을 인덱스 탐색으로 밀어 넣은 물리적 실행 조각입니다. `GROUP BY`는 논리적으로 같은 `user_id`를 하나의 그룹으로 묶으라는 의미입니다. `HashAggregate`는 해시 테이블을 써서 그룹을 만드는 물리 알고리즘입니다. `ORDER BY`는 결과의 순서 의미입니다. `Sort`는 그 순서를 만들기 위해 메모리나 디스크를 쓰는 물리 연산입니다.

이 분리는 alias 함정에서도 드러납니다. SQL을 처음 배울 때 다음 쿼리가 될 것 같아 보입니다.

```sql
SELECT amount * 1.1 AS gross_amount
FROM orders
WHERE gross_amount >= 100000;
```

하지만 많은 DB에서 이 형태는 실패합니다. 논리적으로 `WHERE`는 `SELECT` 출력 alias가 만들어지기 전에 행을 필터링하기 때문입니다. PostgreSQL의 `SELECT` 처리 순서를 기준으로 보면 `WHERE`는 `SELECT` 목록보다 앞에서 평가됩니다. 따라서 `gross_amount`는 아직 `WHERE`가 참조할 수 있는 이름이 아닙니다. 대신 식을 반복하거나, 서브쿼리나 CTE로 한 단계 이름을 만든 뒤 바깥에서 필터링합니다.

```sql
-- 식을 반복하는 방법
SELECT amount * 1.1 AS gross_amount
FROM orders
WHERE amount * 1.1 >= 100000;

-- 논리 단계를 하나 더 만드는 방법
SELECT *
FROM (
    SELECT order_id, amount * 1.1 AS gross_amount
    FROM orders
) AS priced
WHERE gross_amount >= 100000;
```

반대로 `ORDER BY gross_amount`는 허용되는 경우가 많습니다. 논리적으로 정렬은 출력식이 만들어진 뒤에 적용되기 때문입니다. 이 차이를 “DB마다 alias 규칙이 좀 이상하다”로 외우면 오래가지 않습니다. `WHERE`는 행을 남길지 버릴지 결정하는 단계이고, `SELECT` alias는 결과 행의 출력 모양을 만드는 단계이며, `ORDER BY`는 그 결과를 정렬하는 단계라고 읽으면 자연스럽게 복원됩니다.

물리 실행 모델은 이 논리 순서를 그대로 줄 세우지 않습니다. DB는 결과 의미를 보존하는 범위에서 연산을 재배치할 수 있습니다. 예를 들어 `WHERE status='READY'`는 가능한 한 빨리 적용하는 편이 유리합니다. 행을 줄인 뒤 그룹화하면 해시 테이블도 작아지고 정렬할 데이터도 줄어듭니다. 그래서 실행 계획에서는 필터가 scan 노드 아래에 붙거나 인덱스 조건으로 들어갑니다. 논리 순서로는 `FROM` 다음 `WHERE`라는 표현이지만, 물리적으로는 인덱스를 타는 순간 이미 조건 일부가 접근 경로가 됩니다.

```text
논리 모델:
orders 전체 후보 -> WHERE로 줄임 -> GROUP -> HAVING -> SELECT -> ORDER

물리 모델 후보 A:
orders heap 전체 읽기 -> status/date 필터 -> hash aggregate -> sort

물리 모델 후보 B:
status/date 인덱스 탐색 -> 해당 heap row 읽기 -> hash aggregate -> sort

물리 모델 후보 C:
created_at 인덱스 탐색 -> status 필터 -> group aggregate -> sort
```

세 후보는 모두 같은 결과를 만들 수 있습니다. 하지만 비용은 다릅니다. 데이터가 아주 작으면 전체 읽기가 더 단순하고 빠를 수 있습니다. `READY`가 대부분인 테이블이면 status 인덱스가 별 도움이 안 될 수 있습니다. 날짜 조건이 매우 좁으면 날짜 인덱스가 좋을 수 있습니다. 복합 인덱스의 컬럼 순서도 영향을 줍니다. DB는 통계와 비용 모델을 보고 고릅니다. 이때 논리 모델은 결과의 정답지를 제공하고, 물리 모델은 정답을 만드는 경로의 비용과 위험을 보여 줍니다.

실무에서 이 구분이 없으면 성능 장애를 잘못 고칩니다. 예를 들어 사용자가 “`WHERE`가 있으니 먼저 필터링되겠지”라고 믿고, 함수로 감싼 조건을 씁니다.

```sql
SELECT order_id
FROM orders
WHERE to_char(created_at, 'YYYY-MM-DD') = '2026-05-18';
```

논리적으로는 날짜가 2026-05-18인 주문을 고르는 말입니다. 하지만 물리적으로는 `created_at` 값을 문자열로 변환한 뒤 비교해야 하므로 일반적인 `created_at` 인덱스를 바로 쓰기 어려울 수 있습니다. 더 나은 형태는 컬럼을 그대로 두고 범위 조건을 겁니다.

```sql
SELECT order_id
FROM orders
WHERE created_at >= DATE '2026-05-18'
  AND created_at <  DATE '2026-05-19';
```

여기서 핵심은 “함수를 쓰면 무조건 나쁘다”가 아닙니다. 함수 기반 인덱스나 expression index가 있는 DB에서는 다른 선택지가 생길 수 있습니다. 핵심은 논리 조건과 물리 접근 가능성을 분리해 읽어야 한다는 점입니다. SQL 결과가 같아 보여도 DB가 인덱스의 정렬된 키를 바로 좁힐 수 있는지, 아니면 모든 행의 값을 계산해야 하는지가 달라집니다.

논리와 물리의 분리는 `JOIN`에서도 중요합니다. 다음 쿼리를 봅니다.

```sql
SELECT u.user_id, u.email, o.order_id
FROM users u
JOIN orders o ON o.user_id = u.user_id
WHERE o.status = 'READY';
```

논리적으로는 `users`와 `orders`를 `user_id`로 결합한 뒤, 주문 상태가 `READY`인 행을 남기는 질의입니다. 하지만 물리적으로 DB는 `orders`에서 `READY`만 먼저 찾고 그 주문의 `user_id`로 `users`를 찾을 수 있습니다. 반대로 `users`를 먼저 훑고 각 사용자별 주문을 찾을 수도 있습니다. 해시 조인을 만들 수도 있고, 중첩 루프 조인을 쓸 수도 있고, 정렬 병합 조인을 쓸 수도 있습니다. SQL에 `users JOIN orders`라고 적힌 순서가 반드시 실제 접근 순서를 고정하지 않습니다.

```text
논리:
users ⋈ orders where orders.status='READY'

물리 후보:
1. orders(status index) -> READY 주문 찾기 -> users PK lookup
2. users scan -> orders(user_id index) lookup -> status filter
3. orders READY rows로 hash table 만들기 -> users scan과 hash join
```

이 차이를 모르면 `FROM`에 테이블을 어떤 순서로 쓰면 성능이 좋아진다고 과하게 믿거나, 실행 계획의 join 순서가 SQL과 다르다고 버그로 오해합니다. 물론 일부 DB나 힌트, outer join, volatile function, limit, lateral reference 같은 요소는 재배치 가능성을 제한합니다. 그래서 “옵티마이저가 아무렇게나 바꾼다”도 틀린 말입니다. 더 정확한 문장은 “DB는 SQL의 논리 의미를 보존하는 범위에서 물리 경로를 고른다”입니다.

논리 모델은 사용자가 지켜야 할 계약도 제공합니다. 예를 들어 `GROUP BY`가 있는 쿼리에서 그룹화하지 않은 컬럼을 그대로 `SELECT`에 쓰면 왜 문제가 되는지 생각해 봅니다.

```sql
SELECT user_id, order_id, sum(amount)
FROM orders
GROUP BY user_id;
```

`user_id=10` 그룹에 주문이 여러 개 있으면 `order_id`는 어떤 값을 보여 주어야 할까요? 논리 모델에서는 그룹 하나가 결과 행 하나가 되므로, 그룹 안에 여러 값이 있는 `order_id`를 그냥 하나 고를 수 없습니다. 그래서 표준적인 SQL 규칙은 그룹화 컬럼이나 집계식만 허용합니다. 일부 DB의 완화된 모드가 임의 값을 보여 줄 수 있지만, 그것은 논리 모델을 흐리게 만듭니다. 실무에서는 이런 완화 규칙에 기대면 DB 설정이나 버전이 바뀔 때 결과가 흔들릴 수 있습니다.

물리 모델을 볼 때는 page, index, heap, buffer, sort memory 같은 단어가 등장합니다. 이 단어들은 논리 세계의 테이블과 컬럼을 부정하지 않습니다. 같은 데이터를 다른 층위에서 보는 말입니다.

```text
논리 세계:
relation/table -> row/tuple -> column -> predicate -> result set

물리 세계:
file -> page/block -> tuple bytes -> index entry -> buffer -> executor node
```

`orders` 테이블이라는 말은 사용자가 이해하는 이름입니다. 물리적으로는 데이터 파일의 여러 page에 row가 흩어져 있고, 인덱스 파일에는 `status, created_at` 같은 키와 row 위치를 찾는 정보가 있습니다. 실행기는 scan 노드로 row 후보를 만들고, filter 노드나 index condition으로 조건을 적용하고, aggregate 노드로 그룹 상태를 만들고, sort 노드로 결과 순서를 맞춥니다. SQL 결과를 이해할 때는 논리 세계가 필요하고, 성능과 장애를 이해할 때는 물리 세계가 필요합니다.

이 구분은 “정답 결과”와 “좋은 실행”을 따로 검증하게 해 줍니다.

```text
검증 1: 논리 결과
입력 rows 4개를 손으로 따라갔을 때 결과가 `(30,1,200000)`, `(10,1,120000)` 순서인가?
PASS: WHERE, GROUP, HAVING, SELECT, ORDER 의미가 맞다.
FAIL: PAID 주문이나 다른 날짜 주문이 섞인다.

검증 2: 물리 계획
EXPLAIN에서 조건이 scan 아래 filter나 index condition으로 내려갔는가?
PASS: 데이터 크기에 맞는 접근 경로가 보이고, 읽는 행 수가 예상 범위다.
FAIL: 작은 실험은 맞지만 큰 데이터에서 불필요한 full scan, sort spill, 잘못된 join 순서가 보인다.
```

운영에서 성능 장애를 만났을 때는 두 검증을 섞으면 안 됩니다. 결과가 맞다고 해서 계획이 좋은 것은 아닙니다. 계획이 좋아 보여도 결과 의미가 틀리면 더 큰 문제입니다. `EXPLAIN ANALYZE`처럼 실제 실행을 동반하는 명령은 더 많은 정보를 주지만, 운영 DB에서 부하와 부작용을 고려해야 합니다. read-only SELECT라도 오래 실행되면 락, 버퍼 압박, I/O 압박을 만들 수 있습니다. 그래서 먼저 개발/스테이징에서 재현하고, 운영에서는 쿼리 텍스트, 파라미터, 통계, row estimate와 actual 차이, slow query 로그를 조심스럽게 봅니다.

논리와 물리를 섞는 또 다른 함정은 NULL과 collation입니다. `WHERE name = 'kim'`이라는 논리 조건은 단순해 보이지만, 물리적으로는 컬럼의 타입, 문자셋, collation, 인덱스 정의에 따라 비교 방식이 달라집니다. `NULL`은 값이 없거나 모르는 상태라서 `=` 비교가 TRUE/FALSE만으로 끝나지 않습니다. 대소문자를 무시하는 collation에서는 `'Kim'`과 `'kim'`이 같게 비교될 수 있고, binary collation에서는 다를 수 있습니다. 따라서 논리 모델을 읽을 때도 타입과 비교 규칙은 논리 의미의 일부로 봐야 하고, 물리 모델을 읽을 때는 그 규칙이 인덱스 탐색과 정렬에 어떤 영향을 주는지 봐야 합니다.

이 섹션의 senior practical trap은 `EXPLAIN`을 결과 정답처럼 읽는 것입니다. 실행 계획은 DB가 선택한 경로를 보여 줍니다. 계획이 예상과 다르면 SQL 의미가 틀렸다는 뜻이 아니라, 통계·인덱스·비용 모델·조건식 형태가 DB에게 그렇게 보였다는 뜻일 수 있습니다. 반대로 계획이 멋있어 보여도 SQL의 논리 의미가 잘못되면 장애입니다. 예를 들어 `LEFT JOIN` 뒤에 오른쪽 테이블 조건을 `WHERE`에 두면 outer join이 inner join처럼 동작할 수 있습니다.

```sql
-- 의도: 주문이 없어도 모든 사용자를 보고 싶다.
SELECT u.user_id, o.order_id
FROM users u
LEFT JOIN orders o ON o.user_id = u.user_id
WHERE o.status = 'READY';
```

이 쿼리는 `WHERE o.status = 'READY'`가 NULL-extended row를 버리기 때문에 주문이 없는 사용자를 제거합니다. 논리 의미가 의도와 다릅니다. 조건을 join 조건으로 옮겨야 “READY 주문이 있으면 붙이고, 없으면 사용자만 남긴다”는 의미가 됩니다.

```sql
SELECT u.user_id, o.order_id
FROM users u
LEFT JOIN orders o
  ON o.user_id = u.user_id
 AND o.status = 'READY';
```

실행 계획보다 먼저 논리 의미를 확인해야 하는 이유가 여기에 있습니다. 물리 최적화는 논리 의미를 보존하는 범위에서만 의미가 있습니다. 논리 의미가 틀린 쿼리를 빠르게 만드는 것은 더 빠르게 틀리는 일입니다.

이 구분을 몸에 붙이려면 작은 `SELECT`를 볼 때마다 두 줄로 읽는 습관을 들이면 됩니다.

```text
1. 논리 읽기:
   이 SQL은 어떤 행 후보를 만들고, 어떤 조건으로 줄이고, 어떤 그룹/출력/정렬을 요구하는가?

2. 물리 읽기:
   DB는 그 의미를 만들기 위해 어떤 scan, index, join, aggregate, sort, memory, temp file을 쓰는가?
```

직접 해 볼 수 있는 replay path는 다음과 같습니다.

```sql
-- 논리 결과를 작게 고정한다.
CREATE TEMP TABLE orders (
    order_id int,
    user_id int,
    status text,
    amount int,
    created_at date
);

INSERT INTO orders VALUES
  (1, 10, 'READY', 120000, DATE '2026-05-18'),
  (2, 10, 'PAID',   50000, DATE '2026-05-18'),
  (3, 20, 'READY',  30000, DATE '2026-05-19'),
  (4, 30, 'READY', 200000, DATE '2026-05-18');

SELECT user_id, count(*) AS ready_count, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 100000
ORDER BY total_amount DESC;

EXPLAIN
SELECT user_id, count(*) AS ready_count, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 100000
ORDER BY total_amount DESC;
```

PASS 기준은 두 가지입니다. 첫째, 손으로 계산한 논리 결과와 실제 결과가 맞아야 합니다. 둘째, `EXPLAIN` 출력에서 scan, aggregate, sort 같은 물리 노드를 찾아 “이 노드는 논리 단계 중 어느 의미를 만들고 있는가”를 설명할 수 있어야 합니다. FAIL 기준도 두 가지입니다. 첫째, `WHERE`와 `HAVING`, `SELECT alias`, `ORDER BY`의 논리 순서를 섞어 결과를 잘못 예측하는 경우입니다. 둘째, 실행 계획의 노드 이름만 읽고 실제로 어떤 행 수를 줄였는지, 어떤 조건이 인덱스에 들어갔는지, 어떤 단계에서 정렬하거나 그룹화했는지 설명하지 못하는 경우입니다.

PostgreSQL 공식 문서가 `SELECT`의 일반 처리 단계를 따로 설명하고 `EXPLAIN`을 따로 설명하는 이유도 여기에 있습니다. 하나는 SQL 결과의 의미를 설명하는 문서이고, 다른 하나는 DB가 그 의미를 실행하기 위해 고른 계획을 관찰하는 문서입니다. 두 문서를 한 화면에 나란히 놓고 읽으면 DB 공부의 기준축이 생깁니다. `SELECT` 문법을 배울 때는 논리 세계를 놓치지 않고, `EXPLAIN`을 볼 때는 물리 세계를 상상합니다. 이후 인덱스, 옵티마이저, MVCC, 락을 배울 때도 같은 습관이 이어집니다. “내가 원하는 결과는 무엇인가”와 “DB가 실제로 어떤 경로로 그 결과를 만드는가”를 분리해 묻는 것이 DB deep-dive의 두 번째 기초입니다.

이 분리는 `SELECT` 문 하나 안에서도 여러 번 필요합니다. 예를 들어 `DISTINCT`를 봅니다.

```sql
SELECT DISTINCT status
FROM orders
WHERE created_at = DATE '2026-05-18'
ORDER BY status;
```

논리적으로는 먼저 날짜 조건에 맞는 주문을 남기고, 그 행들의 `status` 값을 출력한 뒤, 중복된 status를 제거하고, status 순서로 정렬합니다. 물리적으로는 DB가 해시 집합으로 중복을 제거할 수도 있고, 정렬한 뒤 인접한 같은 값을 제거할 수도 있고, 적절한 인덱스가 있으면 이미 정렬된 키를 이용할 수도 있습니다.

```text
논리 trace
orders rows:
  (1, READY, 2026-05-18)
  (2, PAID,  2026-05-18)
  (3, READY, 2026-05-18)
  (4, READY, 2026-05-19)

WHERE created_at = 2026-05-18
  -> READY, PAID, READY

SELECT status
  -> [READY, PAID, READY]

DISTINCT
  -> [READY, PAID]

ORDER BY status
  -> [PAID, READY]
```

물리 후보는 다르게 생깁니다.

```text
후보 A: Seq Scan -> Filter(date) -> HashAggregate(status) -> Sort(status)
후보 B: Index Scan(created_at,status) -> Unique(status) -> 이미 정렬된 결과 이용
후보 C: Bitmap Index Scan(date) -> Bitmap Heap Scan -> Sort(status) -> Unique
```

세 후보 모두 논리 결과는 같을 수 있습니다. 그러나 메모리 사용, 임시 파일 사용, 읽은 page 수는 달라집니다. 그래서 `DISTINCT`를 “중복 제거”라고만 외우면 부족합니다. 논리적으로는 결과 집합의 중복을 제거하는 의미이고, 물리적으로는 DB가 중복을 발견하기 위해 해시나 정렬이나 인덱스 순서를 사용할 수 있다는 점까지 봐야 합니다.

`LIMIT`도 같은 방식으로 읽어야 합니다.

```sql
SELECT order_id, amount
FROM orders
WHERE status = 'READY'
ORDER BY amount DESC
LIMIT 10;
```

논리적으로는 `READY` 주문을 모두 찾고, 금액 내림차순으로 정렬한 뒤, 앞의 10개를 가져오는 뜻입니다. 물리적으로는 `status, amount DESC` 인덱스가 있으면 DB가 처음부터 상위 10개에 가까운 순서로 접근할 수 있습니다. 인덱스가 없으면 많은 행을 읽고 정렬한 뒤 10개만 버릴 수 있습니다. “LIMIT이 있으니 10개만 읽겠지”라는 말은 논리와 물리를 섞은 오해입니다. DB가 10개만 읽을 수 있으려면 그 10개가 정렬상 앞쪽이라는 사실을 물리 접근 경로가 보장해야 합니다.

```text
LIMIT 10의 두 세계

논리:
READY 전체 -> amount DESC 정렬 -> 앞 10개

물리 A:
status_amount_idx를 amount DESC 순서로 읽음 -> 조건에 맞는 10개에서 멈춤

물리 B:
orders 전체 scan -> READY filter -> 모든 READY row sort -> 10개 반환
```

이 차이는 pagination 장애로 바로 이어집니다. `ORDER BY created_at DESC LIMIT 20 OFFSET 100000` 같은 쿼리는 논리적으로 100001번째부터 20개를 보여 달라는 뜻입니다. 물리적으로는 앞의 100000개를 건너뛰기 위해 많은 행을 읽고 버릴 수 있습니다. 그래서 큰 OFFSET 페이지가 느려집니다. keyset pagination이 필요한 이유도 이 논리/물리 분리에서 나옵니다. 사용자는 “다음 페이지”를 원하지만, DB가 매번 앞 페이지를 다 세고 버리는 물리 경로를 타면 비용이 커집니다.

```sql
-- OFFSET 방식: 뒤 페이지로 갈수록 앞 행을 계속 버릴 수 있다.
SELECT order_id, created_at
FROM orders
WHERE status = 'READY'
ORDER BY created_at DESC, order_id DESC
LIMIT 20 OFFSET 100000;

-- keyset 방식: 마지막으로 본 키 뒤를 범위 조건으로 좁힌다.
SELECT order_id, created_at
FROM orders
WHERE status = 'READY'
  AND (created_at, order_id) < (DATE '2026-05-18', 900000)
ORDER BY created_at DESC, order_id DESC
LIMIT 20;
```

여기서도 keyset이 항상 정답이라는 뜻은 아닙니다. 임의 페이지 점프가 필요하면 OFFSET이 단순할 수 있고, 전체 건수와 페이지 번호 UI가 중요하면 다른 설계가 필요합니다. 하지만 물리 실행 모델을 생각하면 “왜 뒤 페이지가 느려지는가”를 설명할 수 있고, 대안을 선택할 때 trade-off를 말할 수 있습니다.

서브쿼리와 CTE도 논리와 물리를 분리해야 합니다. 다음 쿼리를 봅니다.

```sql
WITH ready_orders AS (
    SELECT order_id, user_id, amount
    FROM orders
    WHERE status = 'READY'
)
SELECT user_id, sum(amount)
FROM ready_orders
GROUP BY user_id;
```

논리적으로는 `ready_orders`라는 이름을 가진 중간 결과를 만든 뒤 바깥에서 그룹화하는 것처럼 읽습니다. 하지만 물리적으로 DB는 CTE를 실제 임시 테이블처럼 materialize할 수도 있고, 바깥 쿼리 안으로 inline해서 하나의 계획으로 최적화할 수도 있습니다. PostgreSQL도 버전과 CTE의 성질, `MATERIALIZED`/`NOT MATERIALIZED` 지정 여부에 따라 동작이 달라질 수 있습니다. 그래서 CTE를 “항상 임시 테이블”로 외우거나 “항상 단순 문법 설탕”으로 외우면 둘 다 위험합니다. 논리적으로 이름 붙인 중간 관계이고, 물리적으로는 최적화 경계가 될 수도 있고 아닐 수도 있다고 읽는 편이 안전합니다.

상관 서브쿼리도 기존 seed인 `database/query.md`가 짚은 좋은 예입니다.

```sql
SELECT e.employee_number, e.name
FROM employees e
WHERE e.salary > (
    SELECT avg(e2.salary)
    FROM employees e2
    WHERE e2.department = e.department
);
```

논리적으로 내부 쿼리는 바깥 행의 `department`에 의존합니다. 그래서 학습자는 “바깥 행마다 내부 쿼리를 매번 실행한다”고 이해하기 쉽습니다. 이 설명은 첫 mental model로는 도움이 되지만, 물리 실행의 확정 사실은 아닙니다. 옵티마이저가 decorrelation을 할 수 있으면 부서별 평균을 미리 구한 뒤 join처럼 바꿀 수 있습니다. 반대로 조건이 복잡하거나 DB가 변환하지 못하면 실제로 반복 실행 비용이 커질 수 있습니다. 따라서 실무 검증은 `EXPLAIN`으로 해야 합니다. 논리 의존성을 이해하되, 물리 반복 실행 여부는 계획으로 확인합니다.

```text
논리:
각 employee e에 대해 e.department의 평균 급여와 비교한다.

물리 후보 A:
employees scan
  -> 각 row마다 department 평균 subplan 실행

물리 후보 B:
department별 avg를 hash aggregate로 먼저 계산
  -> employees와 department avg를 join
```

이런 예는 “SQL은 선언형이라 성능은 DB가 알아서 한다”라는 오해도 고쳐 줍니다. DB가 많은 최적화를 해 주지만, 쿼리 형태와 통계와 인덱스와 DB 버전에 따라 가능한 변환의 범위가 달라집니다. 개발자는 논리 의미를 명확히 쓰고, 물리 계획을 관측하고, 필요한 경우 쿼리 구조나 인덱스나 통계를 조정해야 합니다.

물리 모델을 읽을 때 자주 나오는 `cost`, `rows`, `width` 같은 값도 조심해야 합니다. PostgreSQL `EXPLAIN`의 기본 출력은 실제 실행 결과가 아니라 planner의 추정입니다. `rows=1000`은 실제로 1000행을 읽었다는 뜻이 아니라, 통계와 조건을 바탕으로 그 정도 행이 나올 것이라고 예상했다는 뜻입니다. `EXPLAIN ANALYZE`를 쓰면 실제 실행 결과가 함께 나오지만, 그 쿼리를 실제로 실행합니다. 운영에서 UPDATE나 DELETE에 `EXPLAIN ANALYZE`를 무심코 쓰면 실제 변경이 일어날 수 있으니 트랜잭션 롤백, 스테이징 재현, 읽기 쿼리부터 확인 같은 안전 경계를 둬야 합니다.

```text
EXPLAIN 읽기 기준

추정 계획만 볼 때:
  EXPLAIN SELECT ...
  -> 실행하지 않고 planner가 고른 계획과 추정치를 본다.

실제 실행까지 볼 때:
  EXPLAIN ANALYZE SELECT ...
  -> 실제 실행 시간과 실제 row 수를 본다.
  -> 부하와 부작용을 고려해야 한다.

PASS:
  추정 rows와 실제 rows의 차이가 큰 지점을 찾고,
  그 차이가 join 순서, index 선택, sort/aggregate 비용에 어떤 영향을 줬는지 설명한다.

FAIL:
  cost 숫자를 절대 시간처럼 읽거나,
  rows 추정을 실제 처리 건수로 오해한다.
```

논리와 물리의 분리는 학습 순서에도 영향을 줍니다. 처음부터 파일 page나 B+tree만 보면 SQL 결과 의미를 놓치고, SQL 문법만 보면 성능과 장애를 놓칩니다. 그래서 DB deep-dive에서는 두 세계를 계속 왕복해야 합니다. `SELECT`를 배우면 `EXPLAIN`을 옆에 둡니다. 인덱스를 배우면 어떤 논리 조건을 어떤 물리 순서로 좁히는지 봅니다. 트랜잭션을 배우면 논리적으로 하나의 작업 단위라는 말이 물리적으로 lock, MVCC snapshot, WAL record, commit record와 어떻게 연결되는지 봅니다. NULL과 collation을 배우면 논리 비교 결과가 인덱스와 unique 판단과 정렬에 어떻게 반영되는지 봅니다.

실무에서 senior가 자주 보는 사고 패턴은 “결과가 맞다 -> 그러니 DB는 정상”으로 멈추지 않는 것입니다. 결과가 맞는지 확인한 뒤에는 데이터가 늘어도 같은 방식으로 버틸지, 인덱스가 실제로 선택되는지, 통계가 틀어졌을 때 계획이 바뀌는지, 정렬이 메모리를 넘쳐 임시 파일을 쓰는지, join 순서가 특정 파라미터에서 뒤집히는지 확인합니다. 반대로 성능만 보고 SQL 의미를 바꾸지도 않습니다. `LEFT JOIN`을 `JOIN`으로 바꾸면 빨라질 수 있지만, 주문이 없는 사용자까지 보여야 하는 화면이라면 요구사항을 깨뜨립니다. 논리 의미가 먼저이고, 물리 최적화는 그 의미를 보존해야 합니다.

이 섹션의 마지막 replay 질문은 다음과 같습니다.

```text
작은 쿼리 하나를 고른다.

1. 먼저 손으로 논리 trace를 쓴다.
   FROM 후보, WHERE 뒤 행, GROUP 뒤 그룹, SELECT 출력, ORDER/LIMIT 결과를 적는다.

2. 그다음 EXPLAIN을 본다.
   scan, filter, join, aggregate, sort, limit 노드를 찾는다.

3. 각 물리 노드 옆에 논리 의미를 붙인다.
   "이 Index Scan은 WHERE 조건 중 status/date를 물리 접근으로 바꾼 것이다."
   "이 HashAggregate는 GROUP BY user_id를 구현하는 방법이다."
   "이 Sort는 ORDER BY total_amount DESC를 맞추는 비용이다."

4. 마지막으로 차이를 말한다.
   "논리적으로는 WHERE 뒤 GROUP이지만, 물리적으로는 인덱스 탐색 순간 조건 일부가 이미 적용된다."
```

이 질문에 답할 수 있으면 SQL을 더 이상 문법 순서로만 읽지 않게 됩니다. DB는 논리적 관계를 물리적 실행으로 번역하는 시스템입니다. 사용자는 논리 결과를 정확히 말해야 하고, DB는 그 결과를 만들 물리 경로를 선택합니다. 좋은 개발자는 두 세계 중 하나만 믿지 않습니다. 결과 의미가 맞는지 손으로 확인하고, 실행 경로가 현실적인지 계획과 관측으로 확인합니다.

끝으로, 논리와 물리 사이에는 “동일한 결과를 보장하는 변환”이라는 조건이 항상 붙어야 합니다. 옵티마이저가 조건을 아래로 밀어 넣거나 join 순서를 바꾸거나 subquery를 join으로 바꾸는 것은 결과가 같다고 판단할 수 있을 때만 안전합니다. `INNER JOIN`에서는 필터를 이동하기 쉬운 경우가 많지만, `LEFT JOIN`에서는 필터 위치 하나가 결과 행의 존재 여부를 바꿉니다. 집계 전 필터인 `WHERE`와 집계 후 필터인 `HAVING`도 서로 바꿀 수 없습니다. `LIMIT`과 `ORDER BY`가 함께 있으면 먼저 어떤 순서를 보장해야 하는지도 중요합니다. 그래서 실행 계획을 읽을 때는 “DB가 왜 이렇게 바꿨지?”만 묻지 말고 “이 변환이 어떤 논리 의미를 보존해야 하지?”를 같이 물어야 합니다.

```text
안전한 변환과 위험한 변환을 구분하는 손검사

원래 쿼리:
LEFT JOIN orders o ON o.user_id = u.user_id
WHERE o.status = 'READY'

질문:
주문이 없는 사용자 row는 어디서 사라지는가?

관찰:
WHERE 단계에서 o.status는 NULL이고,
NULL = 'READY'는 TRUE가 아니므로 row가 사라진다.

결론:
이 조건을 ON으로 옮기는 것은 단순 성능 조정이 아니라 결과 의미 변경이다.
```

이 손검사는 optimizer를 의심하라는 뜻이 아닙니다. 개발자가 직접 쿼리를 “최적화”한다고 바꿀 때 더 필요합니다. 빠른 쿼리를 만들겠다고 조건 위치, join 종류, group 기준, NULL 처리, collation을 바꾸면 논리 결과가 달라질 수 있습니다. 반대로 논리 의미가 같다는 것을 손으로 보일 수 있으면, 그다음에 물리 계획을 보고 더 나은 접근 경로를 찾을 수 있습니다. DB 공부에서 논리 모델과 물리 실행 모델을 분리한다는 말은 결국 이 순서를 지키는 일입니다. 먼저 의미를 고정하고, 그 의미를 보존하는 범위에서 실행을 관찰하고 개선합니다.

이 습관은 장애 회고 문장도 바꿉니다. “인덱스를 안 타서 느렸다”라고 쓰기 전에, 어떤 논리 조건을 어떤 인덱스가 지원해야 했는지, 실제 계획은 어떤 물리 경로를 탔는지, 그 차이가 왜 느린 I/O나 정렬 비용으로 이어졌는지를 적습니다. “쿼리를 수정했다”라고 쓰기 전에, 수정 전후의 논리 결과가 같은지 작은 fixture로 확인하고, 실행 계획에서 읽은 행 수와 정렬 방식이 어떻게 달라졌는지 남깁니다. 그러면 회고는 감상문이 아니라 다음 사람이 같은 문제를 재현하고 검증할 수 있는 학습 자산이 됩니다.

따라서 이 장의 기억 문장은 하나입니다. SQL은 결과의 의미를 말하고, 실행 계획은 그 의미를 만들기 위한 물리 경로를 말합니다. 두 세계는 서로 대응하지만 서로를 대체하지 않습니다. DB를 읽을 때 이 한 문장을 계속 적용하면, 문법 암기와 성능 감각이 따로 놀지 않고 같은 사고 흐름으로 이어집니다.

이후 `NULL`, collation, index, optimizer를 배울 때도 같은 방식으로 돌아오면 됩니다. 먼저 논리적으로 어떤 값이 같고 다르고 남아야 하는지 적고, 그다음 물리적으로 어떤 비교 규칙과 접근 경로가 그 결과를 만들었는지 확인합니다.

### 마지막 실무 연결: 결과표가 같아도 실행 경로는 다를 수 있다

논리 모델과 물리 실행 모델을 분리하는 습관은 성능 장애에서 특히 중요하다. 같은 결과표를 반환하는 두 SQL도 하나는 index range scan으로 몇 page만 읽고, 다른 하나는 전체 테이블을 읽은 뒤 정렬할 수 있다. 애플리케이션 입장에서는 둘 다 같은 JSON 응답이지만, DB 입장에서는 buffer pool, lock, temporary file, CPU, I/O 사용량이 완전히 다르다.

따라서 쿼리 리뷰에서는 “결과가 맞는가” 다음에 “어떤 물리 경로로 맞는 결과를 만들었는가”를 확인한다. `EXPLAIN`, 실제 row count, buffer read, sort spill, lock wait를 보지 않고 논리 SQL만 읽으면, 느린 이유를 코드 스타일 문제나 네트워크 문제로 오해하기 쉽다. DB를 잘 다루는 개발자는 결과표와 실행 경로를 동시에 읽는다.
