# Security governance and access control

## roles, users, privileges, grants, ownership

이 절은 role, user, privilege, grant, ownership이 실제 실행 권한을 어떻게 결정하는지 설명한다. 보안 초보자는 계정을 '로그인할 수 있는 이름'으로만 보고, 권한을 'SELECT 가능 여부' 정도로만 생각하기 쉽다. 하지만 데이터베이스 권한 모델은 신원, 소유권, 멤버십, 객체별 권한, 기본 권한, 관리 권한이 겹친 구조다. PostgreSQL 공식 문서는 role이 사용자 또는 사용자 그룹처럼 동작할 수 있고, 객체를 소유하며, 다른 role에게 권한을 부여할 수 있다고 설명한다. MySQL 공식 문서는 account에 부여된 privilege가 어떤 operation을 수행할 수 있는지 결정하고, privilege가 global, database, object 같은 서로 다른 범위에 적용된다고 설명한다.

먼저 권한을 물 흐름으로 본다. 사용자가 SQL을 보내면 DB는 '이 연결은 누구인가', '그 role이 어떤 role의 멤버인가', '대상 객체의 소유자는 누구인가', '객체에 어떤 privilege가 부여되어 있는가', '이 명령이 ownership이나 administrative privilege를 요구하는가'를 차례로 본다. 애플리케이션 계정에 DBA 권한을 주는 함정은 이 흐름을 한 번에 우회한다. 당장은 장애가 줄어 보이지만, 실수한 SQL 하나가 schema 변경, 전체 table scan, destructive DDL, 다른 서비스 데이터 접근으로 번질 수 있다.

```text
SQL 요청: app_order가 SELECT * FROM payment_secret 실행
  -> identity: current_user = app_order
  -> membership: app_order is member of role app_readonly?
  -> object: payment_secret table owner = security_admin
  -> privilege: app_order 또는 상위 role에 SELECT 권한이 있는가?
  -> column/policy: 특정 column만 허용되는가? RLS가 켜져 있는가?
  -> result: 허용, 거부, 또는 row/column 제한
```

이 trace에서 중요한 점은 ownership과 privilege가 다르다는 것이다. 객체 소유자는 보통 그 객체를 변경하거나 권한을 관리할 수 있는 강한 위치에 있다. 반면 application role은 필요한 DML만 가져야 한다. `GRANT ALL`은 개발 중에는 편하지만 운영에서는 사고 반경을 크게 만든다. least privilege, 즉 필요한 최소 권한은 도덕적 구호가 아니라 장애 반경을 줄이는 설계다.

### PostgreSQL과 MySQL을 같은 질문으로 읽기

| 질문 | PostgreSQL 감각 | MySQL 감각 | 운영 함정 |
|---|---|---|---|
| 로그인 주체는 무엇인가 | role에 LOGIN 속성이 있으면 user처럼 연결 가능 | account는 user와 host 조합으로 식별 | host 범위를 `%`로 넓게 둠 |
| 그룹 권한은 어떻게 주는가 | role membership과 inheritance | role과 grant, default role | 멤버십만 주고 활성화 경계를 안 봄 |
| 객체 권한은 어디에 붙는가 | table, schema, sequence, function 등 | global, database, table, column, routine 등 | database 권한을 global로 줌 |
| 소유자는 무엇을 할 수 있는가 | 객체 owner가 권한 관리와 변경의 중심 | DEFINER, object ownership 감각과 privilege 결합 | migration 계정과 runtime 계정 혼합 |
| 확인 명령은 무엇인가 | `\du`, `\dp`, catalog query, `SET ROLE` | `SHOW GRANTS`, information_schema/mysql tables | 선언과 실제 세션 권한 불일치 |

권한 설계는 보통 세 계층으로 나누면 단순해진다. 첫째, 사람 운영 계정은 감사와 긴급 대응을 위해 명확한 role을 가진다. 둘째, migration 계정은 schema 변경 권한을 가지되 runtime에서는 쓰지 않는다. 셋째, application 계정은 서비스가 실제로 수행하는 DML과 필요한 sequence/function 실행 권한만 가진다. 이 세 계층을 섞으면 장애 대응 때 '누가 무엇을 바꿨는가'를 추적하기 어렵고, 애플리케이션 버그가 schema 손상으로 확장된다.

### 권한 matrix 예시

```text
role                         login  schema owner  DDL  DML          grant option
security_admin               no     yes           yes  emergency    yes
migration_order_service      yes    no            yes  seed only    no
app_order_writer             yes    no            no   select/insert/update owned tables no
app_order_reader             yes    no            no   select only  no
analyst_readonly_masked      yes    no            no   select masked views no
```

위 matrix의 목적은 권한을 줄이는 것이 아니라 책임 경계를 보이게 하는 것이다. migration 계정이 운영 중 애플리케이션 pool에 들어가 있으면 FAIL이다. app writer가 `DROP TABLE`이나 `ALTER TABLE`을 할 수 있으면 FAIL이다. analyst가 원본 PII table을 직접 볼 수 있고 masking view를 우회할 수 있으면 FAIL이다. PASS는 각 계정이 자기 업무를 수행할 만큼만 권한을 갖고, 거부되어야 할 명령이 실제로 거부되는 것이다.

### 검증 예시

```sql
-- PostgreSQL 예시: runtime role이 DDL을 못 하는지 확인한다.
SET ROLE app_order_writer;
SELECT order_id FROM orders LIMIT 1;
ALTER TABLE orders ADD COLUMN debug_text text;
-- 기대: SELECT는 성공, ALTER는 permission denied

-- MySQL 예시: 실제 계정에 부여된 권한을 확인한다.
SHOW GRANTS FOR 'app_order_writer'@'10.%';
-- 기대: 필요한 schema/table DML만 있고 global administrative privilege가 없다.
```

권한 설계의 PASS는 허용되어야 할 업무가 성공하고, 실패해야 할 업무가 실패하는 것이다. 실패 케이스를 테스트하지 않는 권한 문서는 보안 문서가 아니라 희망 사항이다.

### 등장 배경 요약

DB 권한 모델이 복잡하게 등장한 배경은 데이터가 여러 프로그램과 여러 사람이 함께 쓰는 공유 자산이 되었기 때문이다. 파일 하나를 혼자 읽고 쓸 때는 운영자 계정 하나로 충분하지만, 운영 DB에서는 애플리케이션, 배치, migration, analyst, emergency operator가 서로 다른 일을 한다. 그래서 role, user, grant, ownership은 번거로운 장식이 아니라 사고 반경을 줄이고 감사 가능성을 남기는 실행 경계다.

### 공식 근거와 로컬 seed

- PostgreSQL Database Roles: https://www.postgresql.org/docs/current/user-manag.html
- MySQL Privileges Provided by MySQL: https://dev.mysql.com/doc/refman/8.4/en/privileges-provided.html
- local seed: `database/database-deep-study-plan.md`

공식 자료는 PostgreSQL의 role 통합 모델과 MySQL의 privilege 범위를 고정한다. 이 절은 두 엔진의 차이를 외우기보다, identity, membership, ownership, object privilege, verification이라는 공통 질문으로 읽도록 구성했다.

### DB 권한 모델 replay drill 1

roles, users, privileges, grants, ownership의 replay drill 1은 login role 분리 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 login role 분리 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 2

roles, users, privileges, grants, ownership의 replay drill 2은 schema ownership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema ownership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 3

roles, users, privileges, grants, ownership의 replay drill 3은 role membership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 role membership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 4

roles, users, privileges, grants, ownership의 replay drill 4은 object privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 object privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 5

roles, users, privileges, grants, ownership의 replay drill 5은 column-level access 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 column-level access 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 6

roles, users, privileges, grants, ownership의 replay drill 6은 grant option 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 grant option 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 7

roles, users, privileges, grants, ownership의 replay drill 7은 host-qualified account 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 host-qualified account 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 8

roles, users, privileges, grants, ownership의 replay drill 8은 migration vs runtime 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 migration vs runtime 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 9

roles, users, privileges, grants, ownership의 replay drill 9은 emergency privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 emergency privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 10

roles, users, privileges, grants, ownership의 replay drill 10은 auditability 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 auditability 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 11

roles, users, privileges, grants, ownership의 replay drill 11은 least privilege test 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 least privilege test 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 12

roles, users, privileges, grants, ownership의 replay drill 12은 ownership transfer 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 ownership transfer 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 13

roles, users, privileges, grants, ownership의 replay drill 13은 login role 분리 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 login role 분리 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 14

roles, users, privileges, grants, ownership의 replay drill 14은 schema ownership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema ownership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 15

roles, users, privileges, grants, ownership의 replay drill 15은 role membership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 role membership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 16

roles, users, privileges, grants, ownership의 replay drill 16은 object privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 object privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 17

roles, users, privileges, grants, ownership의 replay drill 17은 column-level access 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 column-level access 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 18

roles, users, privileges, grants, ownership의 replay drill 18은 grant option 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 grant option 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 19

roles, users, privileges, grants, ownership의 replay drill 19은 host-qualified account 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 host-qualified account 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 20

roles, users, privileges, grants, ownership의 replay drill 20은 migration vs runtime 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 migration vs runtime 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 21

roles, users, privileges, grants, ownership의 replay drill 21은 emergency privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 emergency privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 22

roles, users, privileges, grants, ownership의 replay drill 22은 auditability 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 auditability 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 23

roles, users, privileges, grants, ownership의 replay drill 23은 least privilege test 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 least privilege test 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 24

roles, users, privileges, grants, ownership의 replay drill 24은 ownership transfer 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 ownership transfer 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 25

roles, users, privileges, grants, ownership의 replay drill 25은 login role 분리 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 login role 분리 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 26

roles, users, privileges, grants, ownership의 replay drill 26은 schema ownership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema ownership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 27

roles, users, privileges, grants, ownership의 replay drill 27은 role membership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 role membership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 28

roles, users, privileges, grants, ownership의 replay drill 28은 object privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 object privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 29

roles, users, privileges, grants, ownership의 replay drill 29은 column-level access 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 column-level access 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 30

roles, users, privileges, grants, ownership의 replay drill 30은 grant option 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 grant option 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 31

roles, users, privileges, grants, ownership의 replay drill 31은 host-qualified account 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 host-qualified account 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 32

roles, users, privileges, grants, ownership의 replay drill 32은 migration vs runtime 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 migration vs runtime 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 33

roles, users, privileges, grants, ownership의 replay drill 33은 emergency privilege 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 emergency privilege 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 34

roles, users, privileges, grants, ownership의 replay drill 34은 auditability 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 auditability 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 35

roles, users, privileges, grants, ownership의 replay drill 35은 least privilege test 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 least privilege test 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 36

roles, users, privileges, grants, ownership의 replay drill 36은 ownership transfer 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 ownership transfer 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 36의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 36에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 37

roles, users, privileges, grants, ownership의 replay drill 37은 login role 분리 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 login role 분리 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 37의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 37에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 권한 모델 replay drill 38

roles, users, privileges, grants, ownership의 replay drill 38은 schema ownership 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema ownership 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 38의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

roles, users, privileges, grants, ownership의 replay drill 38에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

## row-level security, encryption, auditing, secret hygiene

이 절은 보안을 SQL injection 방지로 축소하지 않고, 행 단위 보안, 암호화, 감사, secret hygiene이 데이터 보호 흐름에서 어떤 위치를 맡는지 설명한다. SQL injection을 막는 것은 중요하지만, 그것은 입력 문자열이 SQL 명령으로 해석되지 않게 하는 한 경계일 뿐이다. 운영 데이터 보호는 누가 어떤 행을 볼 수 있는가, 민감 값이 전송·저장·로그·백업·덤프에서 어떻게 노출되는가, 누가 언제 권한을 바꿨는가, secret이 repo나 runbook에 남았는가까지 포함한다.

PostgreSQL Row-Level Security는 표준 SQL privilege로 table 접근이 허용된 뒤에도 사용자별로 어떤 row가 보이거나 수정될 수 있는지 제한하는 장치다. 공식 문서는 RLS를 켠 table에 policy가 없으면 default-deny가 적용되고, policy 표현식이 row마다 평가되며, superuser나 `BYPASSRLS` 속성 role은 우회할 수 있다고 설명한다. 이 말은 RLS가 table 권한의 대체물이 아니라 table 권한 위에 놓이는 추가 필터라는 뜻이다. 또한 backup 같은 작업에서 RLS가 조용히 row를 빠뜨리면 위험하므로, RLS가 적용되는지 확인하는 운영 경계도 필요하다.

```text
일반 privilege:
  app_support can SELECT ON tickets

RLS policy:
  support user는 tenant_id = current_setting('app.tenant_id')인 row만 볼 수 있음

요청:
  SELECT * FROM tickets WHERE status = 'OPEN'

실행 의미:
  table privilege 통과
  -> 각 row마다 tenant_id policy 평가
  -> 사용자의 WHERE 조건과 policy 조건이 함께 적용
  -> 허용 row만 반환
```

이 trace에서 중요한 오해는 RLS가 '쿼리 뒤에 붙는 UI filter'가 아니라 DB 실행 안쪽에서 적용되는 보안 조건이라는 점이다. 하지만 RLS도 만능이 아니다. policy가 다른 table을 subquery로 읽으면 동시성 race나 정보 누출 가능성이 생길 수 있고, owner나 BYPASSRLS 권한은 정책을 우회할 수 있다. 따라서 RLS 설계는 `GRANT`, owner, connection role, application session variable, backup 계정, migration 계정을 함께 봐야 한다.

### secret hygiene: 로컬 seed를 본문으로 승격하지 않는 이유

`database/elasticsearch/tools/esdump/auth.ini`처럼 인증 파일 형태를 가진 자료는 학습 예제로 다룰 때도 원문을 그대로 본문에 올리지 않는다. 실제 값이 비어 있거나 오래된 파일처럼 보여도, 파일 이름과 구조 자체가 운영자가 인증 정보를 어디에 두는지 암시할 수 있다. 따라서 본문에는 원문 값을 복사하지 않고, 아래처럼 synthetic redaction 예시로만 다룬다.

```text
나쁜 운영 문서:
  elasticdump --httpAuthFile=./auth.ini
  auth.ini 원문 내용을 runbook에 붙여 넣음

좋은 운영 문서:
  elasticdump --httpAuthFile=./auth.redacted.ini
  auth.redacted.ini:
    user=REDACTED
    password=REDACTED
  실제 secret 위치: password manager 또는 CI secret store
  검증: repo grep에서 password 원문 없음
```

secret hygiene은 개발자 양심에 맡길 일이 아니다. repo에 남는 예제는 synthetic 값이어야 하고, 로그와 dump에는 token, password, connection string, cookie, private endpoint가 들어가지 않아야 한다. dump/restore나 reindex 문서가 운영 편의를 위해 인증 파일을 보여 줄 때도, 본문은 형식과 위험만 설명하고 실제 값은 별도 secret manager 경계에 둔다.

### 암호화와 감사의 위치

암호화는 전송 중 보호, 저장 중 보호, 애플리케이션 수준 필드 암호화로 나눠야 한다. TLS는 client와 DB 사이의 네트워크 경로에서 credential과 query/result를 보호한다. 저장소 암호화는 disk나 backup 매체가 유출됐을 때 raw file을 읽기 어렵게 한다. 필드 암호화는 DB 관리자나 dump 파일을 보는 사람에게도 특정 값을 숨기고 싶을 때 검토한다. 하지만 필드 암호화는 검색, 정렬, 부분 업데이트, 키 회전, 장애 대응을 어렵게 만들 수 있으므로 보안 요구와 사용성의 trade-off를 명시해야 한다.

감사는 '나중에 누가 무엇을 했는지 알 수 있게 로그를 많이 남긴다'가 아니다. 감사 로그는 행위자, 대상, 시간, 성공/실패, 권한 변경, 민감 데이터 접근, 긴급 권한 사용을 재구성할 수 있어야 한다. 그리고 감사 로그 자체도 민감 정보가 될 수 있다. SQL text가 password reset token이나 주민번호 같은 값을 포함하면 로그가 새로운 유출 표면이 된다. 그래서 audit과 redaction은 함께 설계해야 한다.

### 검증 예시

```sql
-- PostgreSQL RLS 최소 재현
CREATE TABLE tenant_docs (tenant_id text, doc_id bigint, body text);
ALTER TABLE tenant_docs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_docs
USING (tenant_id = current_setting('app.tenant_id'))
WITH CHECK (tenant_id = current_setting('app.tenant_id'));

-- 기대 trace
-- SET app.tenant_id = 't1' -> t1 row만 SELECT 가능
-- t2 row INSERT/UPDATE 시 WITH CHECK 위반 또는 UPDATE 0
-- table owner/BYPASSRLS 계정은 별도 테스트 필요
```

secret hygiene 검증은 `rg -n "password|secret|token|AKIA|BEGIN PRIVATE" database/deep-dive database/elasticsearch/tools`처럼 원문 노출 후보를 찾는 것에서 시작한다. PASS는 실제 secret 값이 문서나 예제에 없고, redaction 예시와 secret manager 경계가 분리되어 있는 것이다. FAIL은 편의를 위해 인증 파일 원문을 문서에 붙여 넣는 것이다.

### 등장 배경 요약

행 단위 보안과 secret hygiene이 필요해진 배경은 table 권한 하나만으로는 데이터 노출 경계를 충분히 표현할 수 없기 때문이다. 같은 table 안에서도 tenant, 사용자, 업무 역할, 감사 목적에 따라 볼 수 있는 row와 column이 달라진다. 또한 데이터는 DB 안에만 머물지 않고 로그, dump, backup, search export, runbook, CI secret으로 이동하므로, 접근 제어와 암호화와 redaction을 하나의 보호 흐름으로 다뤄야 한다.

### 공식 근거와 로컬 seed

- PostgreSQL Row Security Policies: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- MySQL Security: https://dev.mysql.com/doc/refman/8.4/en/security.html
- local seed: `database/elasticsearch/tools/esdump/auth.ini`는 sensitive-source-do-not-promote로 값 승격 금지

이 절은 보안 기능 목록이 아니라 데이터 접근 흐름을 보호하는 경계들을 연결한다. 권한, RLS, 암호화, 감사, secret hygiene 중 하나라도 비면 다른 경계가 과도한 책임을 떠안는다.

### DB 데이터 보호 replay drill 1

row-level security, encryption, auditing, secret hygiene의 replay drill 1은 RLS default deny 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 RLS default deny 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 2

row-level security, encryption, auditing, secret hygiene의 replay drill 2은 USING vs WITH CHECK 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 USING vs WITH CHECK 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 3

row-level security, encryption, auditing, secret hygiene의 replay drill 3은 policy race 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 policy race 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 4

row-level security, encryption, auditing, secret hygiene의 replay drill 4은 backup boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 backup boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 5

row-level security, encryption, auditing, secret hygiene의 replay drill 5은 transport encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transport encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 6

row-level security, encryption, auditing, secret hygiene의 replay drill 6은 at-rest encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 at-rest encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 7

row-level security, encryption, auditing, secret hygiene의 replay drill 7은 field encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 field encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 8

row-level security, encryption, auditing, secret hygiene의 replay drill 8은 audit log minimization 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 audit log minimization 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 9

row-level security, encryption, auditing, secret hygiene의 replay drill 9은 secret file redaction 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 secret file redaction 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 10

row-level security, encryption, auditing, secret hygiene의 replay drill 10은 privileged maintenance 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 privileged maintenance 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 11

row-level security, encryption, auditing, secret hygiene의 replay drill 11은 security rules are not filters 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules are not filters 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 12

row-level security, encryption, auditing, secret hygiene의 replay drill 12은 data egress path 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 data egress path 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 13

row-level security, encryption, auditing, secret hygiene의 replay drill 13은 RLS default deny 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 RLS default deny 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 14

row-level security, encryption, auditing, secret hygiene의 replay drill 14은 USING vs WITH CHECK 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 USING vs WITH CHECK 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 15

row-level security, encryption, auditing, secret hygiene의 replay drill 15은 policy race 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 policy race 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 16

row-level security, encryption, auditing, secret hygiene의 replay drill 16은 backup boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 backup boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 17

row-level security, encryption, auditing, secret hygiene의 replay drill 17은 transport encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transport encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 18

row-level security, encryption, auditing, secret hygiene의 replay drill 18은 at-rest encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 at-rest encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 19

row-level security, encryption, auditing, secret hygiene의 replay drill 19은 field encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 field encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 20

row-level security, encryption, auditing, secret hygiene의 replay drill 20은 audit log minimization 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 audit log minimization 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 21

row-level security, encryption, auditing, secret hygiene의 replay drill 21은 secret file redaction 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 secret file redaction 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 22

row-level security, encryption, auditing, secret hygiene의 replay drill 22은 privileged maintenance 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 privileged maintenance 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 23

row-level security, encryption, auditing, secret hygiene의 replay drill 23은 security rules are not filters 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules are not filters 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 24

row-level security, encryption, auditing, secret hygiene의 replay drill 24은 data egress path 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 data egress path 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 25

row-level security, encryption, auditing, secret hygiene의 replay drill 25은 RLS default deny 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 RLS default deny 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 26

row-level security, encryption, auditing, secret hygiene의 replay drill 26은 USING vs WITH CHECK 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 USING vs WITH CHECK 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 27

row-level security, encryption, auditing, secret hygiene의 replay drill 27은 policy race 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 policy race 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 28

row-level security, encryption, auditing, secret hygiene의 replay drill 28은 backup boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 backup boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 29

row-level security, encryption, auditing, secret hygiene의 replay drill 29은 transport encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transport encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 30

row-level security, encryption, auditing, secret hygiene의 replay drill 30은 at-rest encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 at-rest encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 31

row-level security, encryption, auditing, secret hygiene의 replay drill 31은 field encryption 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 field encryption 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 32

row-level security, encryption, auditing, secret hygiene의 replay drill 32은 audit log minimization 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 audit log minimization 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 33

row-level security, encryption, auditing, secret hygiene의 replay drill 33은 secret file redaction 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 secret file redaction 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 34

row-level security, encryption, auditing, secret hygiene의 replay drill 34은 privileged maintenance 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 privileged maintenance 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 35

row-level security, encryption, auditing, secret hygiene의 replay drill 35은 security rules are not filters 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules are not filters 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 36

row-level security, encryption, auditing, secret hygiene의 replay drill 36은 data egress path 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 data egress path 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 36의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 36에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### DB 데이터 보호 replay drill 37

row-level security, encryption, auditing, secret hygiene의 replay drill 37은 RLS default deny 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 RLS default deny 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 37의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

row-level security, encryption, auditing, secret hygiene의 replay drill 37에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.
