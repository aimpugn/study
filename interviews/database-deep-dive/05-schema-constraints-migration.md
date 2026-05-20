# 스키마 변경은 컬럼을 바꾸는 일이 아니라 데이터의 의미와 운영 시간을 함께 바꾸는 일이다

스키마 설계와 migration을 면접에서 설명할 때 흔한 약한 답변은 "정규화해서 중복을 줄이고, 필요하면 반정규화하고, Flyway로 버전 관리합니다"에서 멈추는 것입니다. 이 답변은 키워드는 갖췄지만, 실제로 어떤 불변식을 보호하는지, constraint가 어떤 버그를 구조적으로 막는지, schema diff가 왜 위험 신호인지, online DDL이 왜 완전한 무중단을 의미하지 않는지 설명하지 못합니다.

이 문서는 normalization/denormalization, key/constraint, schema diff, online DDL, Flyway, migration verification, large-table change risk를 하나의 흐름으로 묶습니다. 핵심은 schema가 단순한 저장 모양이 아니라 데이터 의미의 계약이라는 점입니다. Migration은 그 계약을 이미 살아 있는 데이터와 실행 중인 애플리케이션 위에서 바꾸는 작업입니다. 따라서 좋은 설계는 DDL 문장만 맞는 것이 아니라, 기존 데이터가 새 규칙을 만족하는지, 배포 순서가 안전한지, rollback 또는 forward fix가 가능한지, 대용량 table에서 lock과 rewrite 위험이 어느 정도인지까지 닫아야 합니다. Large table 변경이 page, WAL/redo, checkpoint를 어떻게 흔드는지는 [storage page와 buffer I/O](02-storage-pages-buffer-io.md)와 [WAL/redo 복구](03-wal-redo-undo-crash-recovery-pitr.md), 새 index가 plan을 어떻게 바꾸는지는 [인덱스와 옵티마이저](04-index-query-optimizer.md), transaction 경계에서 constraint가 왜 마지막 방어선이 되는지는 [transaction과 ACID boundary](06-transaction-acid-boundary.md)로 이어집니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [정규화는 중복 제거가 아니라 사실의 소유권을 정하는 일이다](#정규화는-중복-제거가-아니라-사실의-소유권을-정하는-일이다)
    - [반정규화는 성능 최적화가 아니라 동기화 계약이다](#반정규화는-성능-최적화가-아니라-동기화-계약이다)
    - [key는 row identity와 business identity를 나눈다](#key는-row-identity와-business-identity를-나눈다)
    - [foreign key는 참조 무결성과 운영 결합을 함께 만든다](#foreign-key는-참조-무결성과-운영-결합을-함께-만든다)
    - [check, not null, default는 application invariant를 DB에 내리는 방법이다](#check-not-null-default는-application-invariant를-db에-내리는-방법이다)
    - [schema diff는 차이를 보여 주지만 위험도를 대신 판단하지 않는다](#schema-diff는-차이를-보여-주지만-위험도를-대신-판단하지-않는다)
    - [online DDL은 알고리즘과 lock 수준을 확인해야 한다](#online-ddl은-알고리즘과-lock-수준을-확인해야-한다)
    - [대용량 DDL은 DB 엔진 밖의 OS I/O 경로까지 흔든다](#대용량-ddl은-db-엔진-밖의-os-io-경로까지-흔든다)
    - [Flyway는 migration 순서와 이력을 관리하지만 안전성 판단을 대신하지 않는다](#flyway는-migration-순서와-이력을-관리하지만-안전성-판단을-대신하지-않는다)
    - [large-table change는 expand, backfill, constrain, cleanup으로 나누면 안전해진다](#large-table-change는-expand-backfill-constrain-cleanup으로-나누면-안전해진다)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
    - [constraint가 race condition을 막는 모습](#constraint가-race-condition을-막는-모습)
    - [nullable column 추가와 backfill 흐름 재생](#nullable-column-추가와-backfill-흐름-재생)
    - [PostgreSQL에서 NOT VALID foreign key 감각 재생](#postgresql에서-not-valid-foreign-key-감각-재생)
    - [Flyway 적용 기록 확인 감각](#flyway-적용-기록-확인-감각)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
    - [전이 질문: 데이터 의미와 배포 시간축을 함께 바꾸는 문제](#전이-질문-데이터-의미와-배포-시간축을-함께-바꾸는-문제)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

스키마는 데이터가 어떤 사실을 말하는지 정하는 계약입니다. Normalization, 즉 정규화는 한 사실을 한 곳에서 말하게 만들어 update anomaly, delete anomaly, insert anomaly를 줄입니다. 예를 들어 user email을 order row마다 복사해 두면 email 변경 때 과거 주문 row를 모두 고쳐야 하고, 일부만 바뀌면 서로 다른 email이 같은 user를 가리키는 모순이 생깁니다. 반대로 denormalization, 즉 반정규화는 조회 성능이나 읽기 모델을 위해 중복을 의도적으로 도입하는 선택입니다. 이 선택이 안전하려면 원본 사실, 복사된 값, 갱신 경로, 검증 루틴이 함께 있어야 합니다.

Key와 constraint는 애플리케이션 코딩 규칙을 데이터베이스가 직접 지키게 하는 장치입니다. Primary key는 row identity를 만듭니다. Unique constraint는 같은 business key가 중복되는 것을 막습니다. Foreign key는 참조 대상이 실제로 존재하는지 보장합니다. Check constraint는 row 안의 값 조합이 규칙을 만족하는지 확인합니다. Not null은 "값이 반드시 있어야 한다"는 계약을 만듭니다. PostgreSQL 문서는 check, not-null, unique, primary key, foreign key를 table constraint의 핵심 형태로 설명하고, MySQL 문서도 foreign key가 table 간 reference를 만들고 related data consistency를 유지한다고 설명합니다.

Schema diff는 "운영 DB와 기대 schema가 같은가"를 확인하는 도구입니다. 하지만 diff는 판단의 시작이지 끝이 아닙니다. column 하나가 다르다는 사실만으로 위험도를 알 수 없습니다. 그 column이 nullable인지, default가 있는지, backfill이 필요한지, index rebuild가 필요한지, foreign key 검증이 필요한지, application version과 호환되는지 봐야 합니다. Migration verification은 적용 전 diff, 적용 기록, 적용 후 schema, row count와 constraint violation, slow query/lock wait를 함께 봅니다.

Online DDL은 운영 중 schema change를 더 안전하게 만드는 기능이지만, "락이 전혀 없다"는 뜻은 아닙니다. MySQL InnoDB online DDL은 operation마다 `INSTANT`, `INPLACE`, `COPY` 가능 여부와 concurrent DML 허용 여부가 다릅니다. PostgreSQL도 `CREATE INDEX CONCURRENTLY`, `ALTER TABLE ... VALIDATE CONSTRAINT`, `NOT VALID` 같은 도구를 제공하지만, 각각 제한과 실패 모드가 있습니다. 대용량 table에서 column type 변경, not null 추가, index 생성, foreign key 추가는 table rewrite, validation scan, metadata lock, replication lag, disk amplification을 만들 수 있습니다.

Flyway 같은 migration tool은 schema 변경을 versioned script로 관리하고, 적용 이력을 기록하고, checksum으로 이미 적용된 migration file이 바뀌었는지 감지합니다. Flyway가 migration 질서를 제공해도 DDL 자체가 안전해지는 것은 아닙니다. migration script의 idempotency, 배포 순서, backfill 분리, rollback 전략, DBMS별 DDL 알고리즘, verification query는 여전히 설계해야 합니다.

좋은 면접 답변은 "정규화 vs 반정규화"와 "Flyway로 관리"를 분리하지 않습니다. 먼저 어떤 사실과 불변식을 schema가 보호해야 하는지 말하고, 그 다음 constraint로 실수 여지를 줄이며, 변경이 필요할 때는 expand -> backfill -> verify -> contract switch -> cleanup 같은 단계로 이미 운영 중인 데이터와 application을 안전하게 이동시킨다고 설명해야 합니다.

## 먼저 잡아야 할 작은 모델

가장 작은 모델은 주문 주소를 어떻게 저장할지 결정하는 장면입니다. 처음 설계는 단순할 수 있습니다.

```text
users
  id
  email
  default_shipping_address

orders
  id
  user_id
  total_amount
```

이 구조에서 주문 당시 배송 주소를 나중에 확인해야 한다면 문제가 생깁니다. 사용자가 기본 배송지를 바꾸면 과거 주문의 배송지까지 바뀐 것처럼 보일 수 있습니다. 그래서 주문 시점의 주소는 order의 사실일 수 있습니다.

```text
normalized-ish split

users
  id
  email

addresses
  id
  user_id
  recipient
  line1
  postal_code

orders
  id
  user_id
  shipping_address_id
  total_amount
```

하지만 이 구조도 "주문 당시 주소"를 완전히 보호하지 못할 수 있습니다. `addresses` row가 나중에 수정되면 과거 주문도 바뀐 주소를 가리킵니다. 그래서 주문에는 snapshot을 따로 둘 수 있습니다.

```text
orders
  id
  user_id
  shipping_recipient
  shipping_line1
  shipping_postal_code
  total_amount
```

여기서 핵심은 "정규화가 항상 좋고 반정규화가 항상 나쁘다"가 아닙니다. 어떤 값이 현재 user profile의 사실인지, 주문 당시 계약의 사실인지 구분해야 합니다. 주문 당시 배송지는 user profile의 중복이 아니라 order의 독립적인 이력 사실일 수 있습니다. 반대로 단순히 목록 조회를 빠르게 하려고 user email을 orders에 복사했다면, 그 값은 원본과 동기화 규칙이 필요한 반정규화입니다.

같은 문자열 값도 어떤 사실을 대표하느냐에 따라 schema 위치가 달라집니다.

| 값 | 사실의 소유자 | 값이 바뀔 때 따라와야 하는 질문 | 자연스러운 저장 방식 |
| --- | --- | --- | --- |
| 사용자의 현재 기본 배송지 | `users` 또는 `addresses` | 사용자가 프로필을 고치면 즉시 바뀌어도 되는가 | 현재 상태 table에 두고 주문은 참조만 합니다. |
| 주문 당시 배송지 | `orders` | 프로필 변경 뒤에도 과거 주문 증빙이 유지되어야 하는가 | 주문 row에 snapshot으로 저장합니다. |
| 목록 조회용 사용자 이름 | 읽기 모델 또는 projection | 원본 변경 뒤 얼마나 늦게 따라가도 되는가 | 반정규화하되 갱신 경로와 검증 query를 둡니다. |

이 표가 중요한 이유는 column 이름이 곧 업무 의미를 설명하지 못하기 때문입니다. `address_id`라는 이름만 보면 현재 주소를 참조하는지, 주문 당시 주소를 가리키는지 알기 어렵습니다. 그래서 schema 설계에서는 column을 어디에 둘지보다 먼저 "이 값은 나중에 바뀌어도 되는 현재 상태인가, 아니면 그 시점의 사실로 남아야 하는 이력인가"를 정해야 합니다. 이 질문이 닫히면 정규화와 반정규화는 취향 문제가 아니라 데이터 의미의 배치 문제가 됩니다.

다음 migration trace는 이미 운영 중인 `orders` table에 `status` column을 추가하는 상황을 기준으로 합니다.

```text
현재
  orders(id, user_id, total_amount)
  application reads/writes without status

목표
  orders(id, user_id, total_amount, status NOT NULL DEFAULT 'CREATED')
  application requires status
```

한 번에 `ADD COLUMN status NOT NULL` 또는 `ADD COLUMN status NOT NULL DEFAULT 'CREATED'`로 목표 모양까지 가려 하면 DBMS와 version, default 표현식에 따라 결과가 크게 갈립니다. 기존 row 때문에 즉시 실패할 수도 있고, metadata-only로 끝날 수도 있고, table rewrite나 긴 metadata lock, validation scan이 생길 수도 있습니다. 더 안전한 흐름은 보통 contract를 넓힌 뒤 좁히는 방식입니다.

```text
1. expand
   ADD COLUMN status nullable or with safe default
   old app still works
   new app can start writing status

2. backfill
   existing rows get status='CREATED'
   batch size controls lock, WAL/redo, replication lag

3. verify
   SELECT count(*) FROM orders WHERE status IS NULL;
   expected 0

4. constrain
   add NOT NULL / CHECK / foreign key after data satisfies the rule

5. switch contract
   app now treats status as required

6. cleanup
   remove compatibility path only after all deployed versions no longer need it
```

이 작은 모델에서 움직이는 대상은 column definition, existing row values, application contract입니다. Migration은 DDL 문장 하나가 아니라 이 세 대상이 같은 방향으로 이동하는 과정입니다.

## 깊은 메커니즘

### 정규화는 중복 제거가 아니라 사실의 소유권을 정하는 일이다

정규화는 흔히 "중복을 줄이는 것"으로 설명되지만, 더 정확히는 한 사실을 누가 소유하는지 정하는 일입니다. 같은 사실이 여러 row에 흩어져 있으면 update할 때 일부만 바뀌는 모순이 생깁니다. 예를 들어 `users.email`과 `orders.user_email`이 둘 다 현재 email을 의미한다면, email 변경 때 orders 전체를 업데이트해야 합니다. 일부 row가 남으면 같은 user가 여러 email을 가진 것처럼 보입니다.

```text
bad duplicate for current fact

users
  id=42, email='new@example.com'

orders
  id=1, user_id=42, user_email='old@example.com'
  id=2, user_id=42, user_email='new@example.com'

question
  user 42의 현재 email은 무엇인가?
```

정규화된 구조에서는 현재 email은 `users`가 소유합니다. 주문 목록에서 email을 보여 줘야 하면 join으로 가져오거나, 읽기 모델에 복사하되 그것이 cache/denormalized projection이라는 사실과 갱신 경로를 명시해야 합니다. 반대로 주문 당시 email이 법적 증빙이나 알림 이력의 일부라면 `orders.contact_email_at_order_time`처럼 이름부터 의미를 분리하는 편이 낫습니다. 이것은 중복이 아니라 다른 사실입니다.

정규화는 무조건 높은 normal form을 목표로 하는 의식이 아닙니다. 면접 답변에서는 "어떤 update anomaly를 막는가"와 "어떤 query와 운영 비용을 새로 만드는가"를 함께 말해야 합니다. 너무 잘게 쪼개면 join 비용과 transaction 경계가 늘고, 너무 뭉치면 중복과 inconsistency 위험이 커집니다.

### 반정규화는 성능 최적화가 아니라 동기화 계약이다

반정규화는 read path를 빠르게 하거나 특정 조회 모델을 단순하게 만들기 위해 같은 정보를 의도적으로 저장하는 선택입니다. 예를 들어 `orders`에 `user_display_name`을 복사해 두면 주문 목록에서 join을 줄일 수 있습니다. 하지만 그 값이 현재 display name인지, 주문 당시 display name인지, 검색용 snapshot인지 정해야 합니다. 이 결정을 숨기면 장애 때 원인을 찾기 어렵습니다.

```text
denormalized read model

source of truth
  users(id=42, display_name='Rody')

read model
  order_summaries(order_id=100, user_id=42, user_display_name='Rody')

must define
  when user display_name changes:
    update summaries immediately?
    async event projection?
    keep historical name unchanged?
```

반정규화가 안전하려면 최소 네 가지가 필요합니다. 첫째, 원본 사실이 어디 있는지 알아야 합니다. 둘째, 복사된 값이 어떤 목적의 snapshot인지 알아야 합니다. 셋째, 갱신 경로가 있어야 합니다. 넷째, 드리프트를 확인하는 query나 job이 있어야 합니다. 이 네 가지 없이 "조회가 느리니 column을 하나 더 둔다"는 선택은 나중에 데이터 불일치를 만듭니다.

### key는 row identity와 business identity를 나눈다

Primary key는 row를 식별하는 값입니다. 보통 surrogate key인 `id`를 쓰지만, business identity와 같지는 않을 수 있습니다. `users.id`는 row identity이고, `users.email`은 login identity일 수 있습니다. email 변경이 가능하다면 primary key로 email을 쓰는 것은 위험할 수 있습니다. 반대로 `countries.iso_code`처럼 business key가 안정적이면 natural key로 쓸 수도 있습니다.

Unique constraint는 business identity를 DBMS가 직접 지키게 만듭니다.

```sql
CREATE TABLE users (
  id bigint PRIMARY KEY,
  email text NOT NULL UNIQUE
);
```

이 constraint가 없으면 애플리케이션에서 "가입 전 email 중복 확인"을 해도 race condition이 생깁니다.

```text
T1: SELECT email='kim@example.com' -> none
T2: SELECT email='kim@example.com' -> none
T1: INSERT kim@example.com
T2: INSERT kim@example.com

without unique constraint:
  duplicate business identity can be committed

with unique constraint:
  one insert succeeds, the other fails
```

이 trace는 constraint의 중요한 역할을 보여 줍니다. Constraint는 validation message를 예쁘게 만들기 위한 장식이 아니라, 동시성 환경에서도 깨지면 안 되는 불변식을 DBMS가 serialization point에서 지키게 하는 장치입니다.

### foreign key는 참조 무결성과 운영 결합을 함께 만든다

Foreign key는 한 table의 값이 다른 table의 key를 참조한다는 계약입니다. MySQL 문서는 foreign key가 table 간 cross-reference를 만들고 related data를 consistent하게 유지한다고 설명합니다. PostgreSQL도 foreign key constraint가 참조 table의 값과 맞아야 함을 보장한다고 설명합니다.

```sql
CREATE TABLE users (
  id bigint PRIMARY KEY
);

CREATE TABLE orders (
  id bigint PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id)
);
```

이 구조는 존재하지 않는 user의 order를 막습니다. 하지만 운영 결합도 만듭니다. parent row delete/update는 child row와 연결되고, cascade policy를 정해야 합니다. `ON DELETE CASCADE`는 편리하지만, user 삭제 하나가 대량 order 삭제로 이어질 수 있습니다. `RESTRICT`나 `NO ACTION`은 무결성을 지키지만 삭제 작업이 실패할 수 있습니다. soft delete를 쓰는 시스템에서는 foreign key와 logical deletion 정책이 어긋날 수도 있습니다.

대용량 table에 foreign key를 나중에 추가할 때는 기존 데이터 검증 scan과 lock을 고려해야 합니다. PostgreSQL에는 `NOT VALID`로 constraint를 추가한 뒤 나중에 `VALIDATE CONSTRAINT`를 수행하는 흐름이 있습니다. 이 방식은 기존 row 검증을 분리해 주지만 lock이 전혀 없다는 뜻은 아니므로, 정확한 lock mode와 validation 시점을 확인해야 합니다. MySQL InnoDB는 foreign key 추가가 table rebuild나 metadata lock과 연결될 수 있습니다. 그러므로 "참조 무결성을 위해 FK를 추가합니다"라는 말 뒤에는 기존 데이터가 이미 깨져 있지 않은지, 추가 중 write가 어떻게 처리되는지, rollback path가 무엇인지가 따라와야 합니다.

### check, not null, default는 application invariant를 DB에 내리는 방법이다

`NOT NULL`은 값이 반드시 존재해야 한다는 계약입니다. `DEFAULT`는 새 row에서 값이 생략됐을 때 어떤 값을 넣을지 정합니다. `CHECK`는 row 내부 값이 어떤 predicate를 만족해야 하는지 정합니다. PostgreSQL 문서는 check constraint가 각 row가 insert/update될 때 boolean expression을 만족해야 한다고 설명합니다. 다만 check expression이 null을 내면 violation으로 보지 않는다는 SQL식 NULL 경계가 있습니다. 그래서 `CHECK (price > 0)`은 `price` null을 막지 못하고, `NOT NULL`을 함께 써야 합니다.

```sql
CREATE TABLE products (
  id bigint PRIMARY KEY,
  price numeric(12, 2) NOT NULL,
  CHECK (price > 0)
);
```

이 constraint는 애플리케이션 bug가 있어도 음수 가격이 들어가는 것을 막습니다. 하지만 모든 business rule을 check로 내릴 수는 없습니다. 다른 table을 조회해야 하는 rule, 시간에 따라 달라지는 rule, 외부 시스템 상태와 연결된 rule은 DB constraint만으로 표현하기 어렵습니다. 이때도 DB constraint로 표현 가능한 핵심 invariant는 DB에 두고, 나머지는 transaction boundary와 application service에서 지키는 식으로 책임을 나눠야 합니다.

`NOT NULL`과 `CHECK`가 서로 다른 일을 한다는 점은 작은 입력으로 보면 더 분명합니다.

```text
INSERT price = NULL
  -> NOT NULL이 먼저 "값이 없음"을 거절합니다.
  -> CHECK (price > 0)만 있으면 SQL의 NULL 판정 때문에 통과할 수 있습니다.

INSERT price = -10
  -> NOT NULL은 통과합니다. 값은 존재하기 때문입니다.
  -> CHECK (price > 0)이 "가격은 0보다 커야 한다"는 의미를 거절합니다.

INSERT price = 1000
  -> NOT NULL 통과
  -> CHECK 통과
  -> row가 table에 들어갑니다.
```

이 trace의 핵심은 constraint를 "입력 검증을 DB에 한 번 더 둔다" 정도로 작게 보지 않는 것입니다. `NOT NULL`은 존재 여부를, `CHECK`는 row 내부 값의 의미를, `UNIQUE`는 여러 row 사이의 business identity를, `FOREIGN KEY`는 table 사이의 존재 관계를 지킵니다. 서로 보호하는 실패 모드가 다르므로, 하나를 넣었다고 다른 규칙까지 자동으로 닫히지는 않습니다.

### schema diff는 차이를 보여 주지만 위험도를 대신 판단하지 않는다

Schema diff는 기대 schema와 실제 schema의 차이를 찾습니다. 개발 DB, staging DB, production DB가 다르면 migration 적용 누락, hotfix, 수동 변경, tool 설정 차이가 드러납니다. 하지만 diff는 "무엇이 다르다"를 보여 줄 뿐, "이 차이가 안전한가"는 별도 판단입니다.

```text
diff result
  production.orders.status is nullable
  expected.orders.status is NOT NULL

must ask
  existing NULL rows exist?
  application old version can write NULL?
  backfill completed?
  adding NOT NULL scans or rewrites table?
  replica lag or lock wait expected?
```

`db_diff.md` 같은 로컬 자료는 diff 도구와 직접 비교 script의 방향을 제공합니다. 그러나 운영 판단에서는 diff 결과를 migration plan과 연결해야 합니다. 특히 대용량 table에서는 "column 하나 차이"가 table rewrite, index rebuild, long metadata lock으로 이어질 수 있습니다. Diff를 보고 곧바로 DDL을 실행하는 것이 아니라, 차이의 의미와 적용 경로를 분리해야 합니다.

Diff 한 줄은 보통 여러 검증 질문으로 풀어야 합니다.

| diff에 보이는 변화 | 바로 묻는 질문 | 확인 예시 |
| --- | --- | --- |
| `status`가 nullable입니다. | 기존 row와 old application이 아직 null을 만들 수 있나요? | `SELECT count(*) FROM orders WHERE status IS NULL;` |
| expected index가 없습니다. | 실제 plan이 느린 scan으로 가나요, 아니면 다른 index로 충분한가요? | 대표 query `EXPLAIN`과 production cardinality 확인 |
| foreign key가 없습니다. | 이미 orphan row가 있나요, 추가 중 write는 어떻게 막거나 검증하나요? | child에서 parent를 못 찾는 anti-join 확인 |
| column type이 다릅니다. | 변환이 lossless인가요, table rewrite나 index rebuild가 필요한가요? | sample cast, max length, lock/rewrite 알고리즘 확인 |

이렇게 풀어 보면 diff는 "운영 DB가 틀렸다"는 판결문이 아니라 "어떤 계약이 아직 목표와 다르다"는 탐지 결과입니다. 적용 판단은 데이터 값, application 배포 상태, DBMS별 DDL 알고리즘, lock과 I/O 비용을 합쳐야 닫힙니다.

### online DDL은 알고리즘과 lock 수준을 확인해야 한다

Online DDL은 운영 중 schema 변경을 가능하게 하는 기능군입니다. 하지만 이름 때문에 "무조건 무중단"으로 오해하기 쉽습니다. MySQL InnoDB online DDL 문서는 각 operation마다 algorithm과 lock, concurrent DML 허용 여부가 다릅니다. 어떤 변경은 `INSTANT`로 metadata만 바뀔 수 있고, 어떤 변경은 `INPLACE`라 해도 background 작업과 metadata lock이 생기며, 어떤 변경은 `COPY`로 table을 새로 만들어야 할 수 있습니다.

PostgreSQL도 `CREATE INDEX CONCURRENTLY`처럼 write blocking을 줄이는 방법이 있지만, 일반 `CREATE INDEX`보다 더 오래 걸리고, transaction block 안에서 실행할 수 없으며, 실패하면 invalid index가 남을 수 있습니다. `ALTER TABLE`의 일부 operation은 table rewrite를 일으키고, 일부는 metadata change로 끝납니다. Version에 따라 최적화가 달라지므로 "PostgreSQL은 ADD COLUMN DEFAULT가 항상 table rewrite다" 같은 오래된 지식을 현재 version에 그대로 적용하면 안 됩니다.

```text
large table DDL checklist

operation
  ADD COLUMN?
  ALTER TYPE?
  CREATE INDEX?
  ADD FOREIGN KEY?
  SET NOT NULL?

risk
  table rewrite?
  validation scan?
  metadata lock?
  concurrent DML blocked?
  replication lag?
  disk temporary space?
  rollback cost?
```

이 체크를 통해 DDL은 개발 환경에서 빨리 끝났더라도 운영에서 위험할 수 있음을 설명할 수 있습니다. 개발 DB의 row 수가 1000개이고 production이 3억 row라면 같은 DDL 문장도 완전히 다른 사건입니다. 운영 영향은 schema 문법만이 아니라 page rewrite, index rebuild, WAL/redo 증가, checkpoint와 replication lag까지 이어지므로 앞선 [storage I/O 모델](02-storage-pages-buffer-io.md)을 같이 가져와야 합니다.

### 대용량 DDL은 DB 엔진 밖의 OS I/O 경로까지 흔든다

대용량 DDL을 설명할 때는 DBMS 내부 lock만 보시면 부족합니다. Table rewrite, index build, backfill update는 DB buffer pool이나 shared buffer를 더럽히고, WAL/redo를 늘리며, 결국 운영체제의 page cache, filesystem, block layer, device driver, storage queue까지 I/O를 밀어 넣습니다. Linux 문서 기준으로 file write는 먼저 page cache에 더러운 page를 만들 수 있고, `fsync`는 그 변경을 storage device까지 밀어내 완료를 기다립니다. Linux block layer는 파일시스템과 block device driver 사이에서 요청을 software queue와 hardware dispatch queue로 보내며, device가 바로 받을 수 없으면 나중에 다시 보냅니다.

```text
large table backfill / index build
  DB worker
    -> buffer pool/shared buffers dirty page 증가
    -> WAL/redo record 생성
    -> kernel page cache에 write 전달
    -> filesystem이 block I/O 요청 생성
    -> block layer가 merge/reorder/dispatch
    -> device driver/NVMe queue
    -> disk/SSD가 write 완료
    -> fsync 또는 checkpoint가 완료를 기다림
```

이 경로 때문에 migration은 SQL 문장 하나보다 넓은 운영 사건이 됩니다. DB lock wait가 짧아도 WAL flush가 밀리면 commit latency가 흔들릴 수 있고, checkpoint가 더러운 page를 한꺼번에 밀어내면 평소 query까지 I/O wait를 겪을 수 있습니다. Replica는 primary가 만든 WAL/binlog를 받아 다시 apply해야 하므로, primary에서 안전해 보인 backfill도 replica lag로 늦게 나타날 수 있습니다. 면접에서 `online DDL을 썼습니다`라고만 말하지 말고, `lock`, `rewrite`, `WAL/redo`, `checkpoint`, `replication lag`, `OS page cache와 fsync 대기`를 함께 관측했다고 설명하면 운영 감각이 훨씬 선명해집니다.

### Flyway는 migration 순서와 이력을 관리하지만 안전성 판단을 대신하지 않는다

Flyway는 versioned migration을 파일로 두고, 적용 이력을 schema history table에 기록합니다. 이미 적용된 migration의 checksum이 바뀌면 drift를 감지할 수 있습니다. Redgate Flyway 문서도 migration을 database를 한 version에서 다음 version으로 옮기는 script로 설명하고, versioned migration과 repeatable migration을 구분합니다.

```text
typical Flyway flow

V1__create_users.sql
V2__create_orders.sql
V3__add_order_status.sql

flyway migrate
  reads schema history
  finds pending versions
  applies in order
  records success and checksum

flyway validate
  compares applied migrations with local files
  detects changed checksum, missing files, ordering problems
```

이 구조는 팀이 어떤 migration이 적용되었는지 공유하는 데 매우 유용합니다. 하지만 `V3__add_order_status.sql` 안에 위험한 DDL이 들어 있으면 Flyway가 자동으로 안전하게 바꿔 주지는 않습니다. 대용량 backfill, lock timeout, retry, data verification, online DDL algorithm hint, rollback/forward fix 전략은 script 작성자가 설계해야 합니다.

이미 적용된 versioned migration file을 수정하는 것은 위험합니다. 로컬 파일은 바뀌었지만 production에는 예전 내용이 적용되어 있을 수 있고, checksum mismatch가 생깁니다. 일반적으로 새 migration을 추가해 forward fix를 만듭니다. `repair`는 이력을 고치는 강한 도구이므로, 실제 DB 상태와 repository migration file의 관계가 검증된 상황에서만 써야 합니다. 이 판단은 "이력 표를 맞춘다"가 아니라 "운영 DB가 어떤 변경을 실제로 겪었는지 감사 가능하게 남긴다"는 문제입니다.

### large-table change는 expand, backfill, constrain, cleanup으로 나누면 안전해진다

대용량 table에서 가장 안전한 기본 사고는 "한 번에 완성된 schema로 점프"가 아니라, application과 DB가 동시에 견딜 수 있는 중간 계약을 거쳐 목표 계약으로 좁혀 가는 것입니다. 이것은 무조건 느리게 하자는 말이 아니라, 기존 데이터와 running application을 깨지 않기 위한 계약 이동입니다.

```text
target: orders.status must be NOT NULL

phase 1 expand
  ADD COLUMN status nullable
  old app ignores it
  new app starts writing it

phase 2 backfill
  UPDATE orders SET status='CREATED'
  WHERE status IS NULL
  in small batches

phase 3 verify
  count NULL rows
  verify application writes status
  monitor locks, dead tuples, redo/WAL, replica lag

phase 4 constrain
  SET NOT NULL or add CHECK after no NULL remains

phase 5 cleanup
  remove compatibility code and temporary indexes
```

이 흐름은 모든 변경에 기계적으로 맞는 답은 아니지만, 실패 모드를 줄이는 강한 기본값입니다. 특히 column rename, type change, large index creation, foreign key validation, table split, backfill은 application version과 DB schema version이 잠시 공존하는 시간을 고려해야 합니다. 이 공존을 생각하지 않으면 배포 중간에 old app이 새 schema를 못 읽거나, new app이 old schema에 쓰려다 실패합니다.

단계별로 "무엇이 아직 허용되는가"를 적어 두면 rollback과 forward fix 판단이 쉬워집니다.

| 단계 | DB schema | old app | new app | 실패했을 때 안전한 방향 |
| --- | --- | --- | --- | --- |
| expand 직후 | 새 column은 있지만 필수는 아닙니다. | 기존처럼 읽고 씁니다. | 새 column을 채워 쓰기 시작할 수 있습니다. | old app과 new app을 모두 견디므로 되돌리거나 다시 배포할 수 있습니다. |
| backfill 중 | 일부 row만 새 값을 가집니다. | 여전히 호환되어야 합니다. | null 가능성을 방어해야 합니다. | batch를 멈추고 원인 row를 고칩니다. |
| constrain 직전 | 모든 row가 새 규칙을 만족해야 합니다. | 더 이상 null을 만들면 안 됩니다. | 새 규칙을 전제로 동작할 수 있습니다. | 검증 query가 0이 아니면 constrain을 보류합니다. |
| cleanup 이후 | 새 규칙이 유일한 계약입니다. | 남아 있으면 실패할 수 있습니다. | 정상 경로입니다. | rollback보다 forward fix가 더 안전한 경우가 많습니다. |

이 표의 목적은 migration을 느리게 만들자는 것이 아닙니다. 배포 중간의 어느 순간에도 DB와 애플리케이션이 서로를 이해할 수 있는지 확인하는 것입니다. 이 확인 없이 한 번에 최종 schema로 이동하면, SQL은 성공했는데 old application이 장애를 내거나, application은 배포됐는데 replica lag 때문에 read model이 늦게 따라오는 식의 문제가 생깁니다.

## DBMS별 경계

PostgreSQL은 constraint와 online-ish schema change에서 강력한 도구를 제공합니다. `CHECK`, `NOT NULL`, `UNIQUE`, `PRIMARY KEY`, `FOREIGN KEY`를 table constraint로 다룹니다. `CREATE INDEX CONCURRENTLY`는 write blocking을 줄이지만 일반 index build보다 복잡하고 transaction block 안에서 실행할 수 없습니다. `ALTER TABLE ... ADD CONSTRAINT ... NOT VALID`와 `VALIDATE CONSTRAINT`는 기존 row 검증을 나중으로 미루는 데 유용합니다. 단, validation scan과 lock mode는 operation마다 확인해야 합니다.

MySQL 8.4 InnoDB는 online DDL operation별로 algorithm과 lock behavior를 확인해야 합니다. `ALGORITHM=INSTANT`는 특정 metadata-only 변경에서 매우 빠를 수 있지만 모든 변경에 적용되지 않습니다. `INPLACE`도 table copy가 없다는 의미에 가깝지, DML 영향이 전혀 없다는 뜻은 아닙니다. `COPY`는 table copy를 만들 수 있어 대용량 table에서 매우 위험합니다. Foreign key와 character string column은 character set과 collation compatibility도 확인해야 합니다.

Constraint semantics도 제품별 차이가 있습니다. PostgreSQL은 오래전부터 check constraint를 enforcement 대상으로 다뤘고, MySQL은 8.0.16부터 check constraint enforcement가 본격적으로 의미를 가집니다. MySQL 5.7이나 5.5 같은 오래된 운영 환경을 다룰 때는 문법이 받아들여져도 실제 enforcement가 없는 경우를 조심해야 합니다. 이 저장소의 MySQL 5.5 대용량 인덱싱 메모도 같은 이유로 version-specific operation risk를 분리해야 함을 보여 줍니다.

Flyway는 DBMS 독립적인 migration 관리 도구지만, SQL script의 실제 의미는 DBMS별입니다. `CREATE INDEX CONCURRENTLY`는 PostgreSQL 문법이고, MySQL의 online DDL hint와 다릅니다. Repeatable migration은 view, procedure, function처럼 현재 정의를 다시 맞추는 데 유용할 수 있지만, 대용량 data migration을 repeatable로 넣으면 위험할 수 있습니다. Migration 도구와 DBMS DDL 기능을 같은 층으로 섞지 않아야 합니다.

## 직접 재생해 보기

아래 replay는 disposable DB에서만 실행합니다. 운영 DB에서 lock이나 rewrite를 유발할 수 있는 DDL을 실험하지 않습니다.

### constraint가 race condition을 막는 모습

두 세션을 열고 unique constraint 유무를 비교합니다.

```sql
CREATE TABLE users_constraint_lab (
  id bigint PRIMARY KEY,
  email text NOT NULL UNIQUE
);
```

세션 A와 세션 B에서 동시에 같은 email insert를 시도합니다.

```sql
INSERT INTO users_constraint_lab VALUES (1, 'kim@example.com');
INSERT INTO users_constraint_lab VALUES (2, 'kim@example.com');
```

PASS 신호는 하나가 성공하고 하나가 unique violation으로 실패하는 것입니다. 애플리케이션의 사전 중복 조회만으로는 이 race를 막을 수 없습니다. FAIL 신호는 "가입 전에 SELECT로 확인하면 충분하다"고 설명하는 것입니다.

### nullable column 추가와 backfill 흐름 재생

```sql
CREATE TABLE orders_migration_lab (
  id bigint PRIMARY KEY,
  total_amount numeric(12, 2) NOT NULL
);

INSERT INTO orders_migration_lab VALUES (1, 1000), (2, 2000);

ALTER TABLE orders_migration_lab
ADD COLUMN status text;

UPDATE orders_migration_lab
SET status = 'CREATED'
WHERE status IS NULL;

SELECT count(*) AS null_status_count
FROM orders_migration_lab
WHERE status IS NULL;
```

PASS 신호는 backfill 뒤 `null_status_count = 0`이 되는 것입니다. 그 다음에야 `NOT NULL`이나 check를 붙일 후보가 됩니다. 실제 production에서는 batch update, lock timeout, retry, replica lag 관측이 필요합니다. FAIL 신호는 기존 row가 많은 table에 바로 `NOT NULL` contract를 요구하면서 데이터 검증을 생략하는 것입니다.

### PostgreSQL에서 NOT VALID foreign key 감각 재생

PostgreSQL에서만 해당하는 흐름입니다.

```sql
CREATE TABLE parent_lab (id bigint PRIMARY KEY);
CREATE TABLE child_lab (id bigint PRIMARY KEY, parent_id bigint);

ALTER TABLE child_lab
ADD CONSTRAINT child_parent_fk
FOREIGN KEY (parent_id) REFERENCES parent_lab(id)
NOT VALID;

ALTER TABLE child_lab
VALIDATE CONSTRAINT child_parent_fk;
```

PASS 신호는 constraint 추가와 기존 row 검증 단계를 분리할 수 있다는 점을 이해하는 것입니다. 이 흐름도 lock이 전혀 없다는 뜻은 아니므로 공식 문서와 운영 관측으로 확인해야 합니다. FAIL 신호는 `NOT VALID`를 "검증하지 않는 FK"로 영구 사용해도 된다고 설명하는 것입니다. 목표는 보통 기존 데이터 검증을 나중에 안전하게 수행하는 것입니다.

### Flyway 적용 기록 확인 감각

로컬 Flyway 프로젝트에서는 다음 흐름을 확인합니다.

```bash
flyway info
flyway validate
flyway migrate
flyway info
```

PASS 신호는 pending migration이 적용되고 schema history에 version, description, checksum, success 상태가 남는 것입니다. FAIL 신호는 이미 적용한 `V2__...sql`을 수정하고 `repair`로 덮으면 된다고 가볍게 말하는 것입니다. 실제 DB 상태와 파일 내용이 맞는지 확인하지 않은 repair는 audit trail을 망가뜨릴 수 있습니다.

## 면접 꼬리 질문

1. 정규화와 반정규화의 기준은 무엇인가?

    먼저 어떤 사실이 어디에서 한 번만 말해져야 하는지 봅니다. 정규화는 update anomaly를 줄이고 사실 소유권을 분명히 합니다. 반정규화는 읽기 성능이나 조회 모델을 위해 중복을 도입하지만, 원본, 갱신 경로, 드리프트 검증이 함께 있어야 합니다.

2. unique constraint가 있는데도 애플리케이션에서 중복 확인을 해야 하나?

    사용자 경험을 위해 사전 확인을 할 수는 있지만, 최종 보장은 unique constraint가 해야 합니다. 동시 insert race에서는 두 요청이 모두 사전 확인을 통과할 수 있습니다. DB constraint가 마지막 방어선입니다.

3. foreign key를 항상 걸어야 하나?

    참조 무결성을 DB에서 보장해야 하면 강력한 선택입니다. 하지만 cascade 정책, 대량 삭제, online migration, cross-service ownership, logical deletion과 충돌할 수 있습니다. FK를 걸지 않는다면 application이나 batch verification이 어떤 방식으로 무결성을 지키는지 설명해야 합니다.

4. online DDL이면 무중단인가?

    아닙니다. operation, DBMS version, algorithm, lock mode, table size, concurrent workload에 따라 다릅니다. metadata lock, validation scan, table rebuild, replication lag, disk usage를 확인해야 합니다.

5. Flyway가 있으면 schema drift가 사라지나?

    Flyway는 versioned migration과 checksum으로 drift를 발견하고 관리하는 데 도움을 줍니다. 그러나 수동 hotfix, 이미 적용된 파일 수정, DBMS별 DDL 차이, 실패한 migration 후 상태 판단은 여전히 운영 책임입니다.

6. 대용량 table에 NOT NULL column을 추가할 때 무엇을 확인하나?

    default와 nullable 전략, 기존 row backfill, batch size, lock duration, table rewrite 여부, WAL/redo 증가, replica lag, application old/new version compatibility, constraint validation 순서를 확인합니다.

## 함정 질문

1. "정규화하면 항상 좋은 설계인가요?"

    아닙니다. 정규화는 중복과 anomaly를 줄이지만 join 비용과 transaction 경계를 늘릴 수 있습니다. 주문 당시 주소처럼 원본 user profile과 다른 이력 사실은 중복이 아니라 별도 사실로 저장해야 할 수 있습니다.

2. "반정규화는 나쁜 냄새인가요?"

    무조건 나쁘지 않습니다. 읽기 모델, 분석 집계, 검색, 이벤트 projection에서는 필요할 수 있습니다. 문제는 원본과 복사본의 의미, 갱신 경로, 검증 루틴 없이 중복을 만드는 것입니다.

3. "schema diff가 없으면 migration은 안전한가요?"

    diff가 없다는 것은 기대 schema와 현재 schema가 같다는 구조적 신호일 뿐입니다. 데이터 값이 constraint를 만족하는지, index가 사용되는지, long transaction이나 lock wait가 없는지, application version이 호환되는지는 별도 확인해야 합니다.

4. "ADD COLUMN은 항상 가벼운 작업인가요?"

    DBMS version과 column definition에 따라 다릅니다. nullable metadata-only일 수 있고, default나 type change 때문에 table rewrite가 필요할 수 있습니다. MySQL InnoDB에서는 INSTANT/INPLACE/COPY 가능 여부를 확인해야 하고, PostgreSQL도 version별 최적화가 다릅니다.

5. "이미 적용된 Flyway migration을 수정하고 checksum repair하면 되나요?"

    원칙적으로 새 migration으로 forward fix하는 편이 안전합니다. `repair`는 실제 DB 상태와 local migration 파일의 관계를 확인한 뒤 이력 metadata를 고치는 도구이지, 적용된 schema를 자동으로 바꾸는 도구가 아닙니다.

6. "FK가 있으면 데이터 정합성 문제가 모두 사라지나요?"

    FK는 참조 존재를 보장하지만 business state machine, 금액 합계, cross-database reference, soft delete visibility, 외부 시스템 상태는 보장하지 못합니다. 가능한 invariant는 DB constraint로 내리고, 나머지는 transaction/application verification으로 닫아야 합니다.

### 전이 질문: 데이터 의미와 배포 시간축을 함께 바꾸는 문제

다른 플랫폼으로 옮겨도 schema 변경은 `데이터 의미 변경`과 `배포 순서 변경`이 함께 움직입니다. document store의 field 추가, search index의 mapping 변경, API response shape 변경도 같은 질문을 받습니다. old reader와 new writer가 동시에 존재하는 동안 어떤 형태가 안전한지 먼저 정해야 합니다.

## 더 깊게 볼 자료

공식 문서는 constraint와 DDL 동작의 1차 근거입니다. 이 저장소의 기존 `database/` 문서는 schema diff, Flyway, 대용량 indexing 운영 감각을 제공하지만, DBMS별 DDL 가능 여부와 lock 경계는 공식 문서로 다시 확인해야 합니다.

- PostgreSQL current docs
    - [Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
    - [ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html)
    - [CREATE INDEX](https://www.postgresql.org/docs/current/sql-createindex.html)
    - [Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
    - [Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html)
    - [Write-Ahead Logging](https://www.postgresql.org/docs/current/wal-intro.html)
    - [WAL Configuration](https://www.postgresql.org/docs/current/wal-configuration.html)
- MySQL 8.4 Reference Manual
    - [CREATE TABLE and Generated Columns / Constraints](https://dev.mysql.com/doc/refman/8.4/en/create-table.html)
    - [Using FOREIGN KEY Constraints](https://dev.mysql.com/doc/refman/8.4/en/create-table-foreign-keys.html)
    - [Online DDL Operations](https://dev.mysql.com/doc/refman/8.4/en/innodb-online-ddl-operations.html)
    - [ALTER TABLE Statement](https://dev.mysql.com/doc/refman/8.4/en/alter-table.html)
- Linux / POSIX I/O references
    - [Linux kernel docs: Memory Management Concepts](https://docs.kernel.org/admin-guide/mm/concepts.html)
    - [Linux kernel docs: Multi-Queue Block IO Queueing Mechanism](https://www.kernel.org/doc/html/latest/block/blk-mq.html)
    - [Linux man-pages: fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)
- Flyway official docs
    - [Flyway Migrations](https://documentation.red-gate.com/flyway/flyway-concepts/migrations)
    - [Flyway Validate](https://documentation.red-gate.com/flyway/reference/commands/validate)
    - [Flyway Repair](https://documentation.red-gate.com/flyway/reference/commands/repair)
- Repo study material
    - `database/deep-dive/schema-migration-ops/10-schema-design-constraints.md`
    - `database/db_diff.md`
    - `database/migration/flyway/flyway.md`
    - `database/mysql/problems/mysql5_online_large_indexing.md`
    - `interviews/database-storage-search-nosql.md`
