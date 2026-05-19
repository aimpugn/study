# Schema Design Constraints

## normalization, denormalization, keys, constraints

스키마 설계에서 정규화, 키, 제약 조건은 "테이블을 예쁘게 나누는 규칙"이 아니라 데이터베이스가 어떤 사실을 직접 지키고, 어떤 사실은 애플리케이션이나 운영 절차에 맡길지를 정하는 경계선이다. 같은 주문 금액, 같은 고객 주소, 같은 상품 가격이 여러 행에 흩어져 있을 때 가장 먼저 물어야 할 질문은 "몇 개 테이블로 나눌까"가 아니라 "이 값은 어느 한 곳에서만 참이어야 하는 사실인가, 아니면 특정 시점의 기록으로 복제되어도 되는 사실인가"다. 정규화(normalization)는 중복을 무조건 제거하자는 미학이 아니라 갱신 이상(update anomaly), 삽입 이상(insert anomaly), 삭제 이상(delete anomaly)을 줄이기 위해 사실의 소유자를 좁히는 작업이다. 반정규화(denormalization)는 그 반대말처럼 보이지만, 제대로 쓰이면 정규화를 포기하는 것이 아니라 이미 정규화된 사실을 특정 읽기 경로, 감사 요구, 시점 기록, 장애 격리 요구에 맞게 의도적으로 복제하고 검증 책임을 새로 붙이는 작업이다.

관계형 모델이 등장한 배경을 먼저 잡아야 이 구분이 선명해진다. Codd의 1970년 관계형 모델 논문은 사용자가 데이터가 기계 내부에 어떤 포인터나 파일 배치로 저장되는지 알지 않아도, 관계(relation)와 연산으로 데이터를 다룰 수 있어야 한다는 문제의식에서 출발했다. 이 배경에서 정규형(normal form)은 단순한 테이블 분할 규칙이 아니라 "사용자 모델 안에서 중복과 일관성 문제를 다룰 수 있게 하는 표현 방식"이었다. 오늘의 MySQL이나 PostgreSQL은 내부적으로 B-tree 인덱스, 페이지, MVCC, lock, WAL처럼 복잡한 물리 구조를 쓰지만, 개발자가 스키마를 설계할 때는 여전히 "어떤 사실을 어떤 relation에 둘 것인가"라는 논리 모델을 먼저 결정한다. 그 논리 모델이 흐리면 뒤에서 아무리 인덱스를 잘 붙여도 데이터가 서로 다른 말을 하기 시작한다.

가장 작은 예제로 주문 시스템을 생각해 보자. 사용자가 상품을 주문하면 `orders`, `order_items`, `customers`, `products` 같은 테이블이 자연스럽게 떠오른다. 초보 설계에서는 한 화면에 필요한 값을 모두 `orders` 한 행에 넣고 싶어진다. 예를 들어 주문번호, 고객명, 고객 이메일, 배송지, 상품명, 상품 가격, 수량, 쿠폰명, 결제수단, 주문상태를 한 행에 두면 조회는 쉬워 보인다. 하지만 이 설계는 "주문이라는 사건", "고객이라는 주체", "상품이라는 판매 단위", "결제라는 거래"를 한 테이블에 섞는다. 그 결과 고객 이메일을 수정할 때 과거 주문 행까지 모두 수정해야 하는지, 상품명을 바꿀 때 과거 주문 영수증의 상품명도 바꿔야 하는지, 상품이 아직 주문되지 않았을 때 상품 정보를 어디에 저장할지 같은 질문이 한꺼번에 터진다.

```text
처음 떠올리기 쉬운 넓은 주문 행

orders_flat
+----------+-------------+------------------+-------------+------------+--------------+-----+-----------+
| order_id | customer_id | customer_email   | product_id  | product_nm | unit_price   | qty | status    |
+----------+-------------+------------------+-------------+------------+--------------+-----+-----------+
| 1001     | 7           | a@shop.test      | 501         | Keyboard   | 80000        | 1   | PAID      |
| 1002     | 7           | a@shop.test      | 502         | Mouse      | 30000        | 2   | PAID      |
| 1003     | 7           | old@shop.test    | 503         | Monitor    | 250000       | 1   | CANCELLED |
+----------+-------------+------------------+-------------+------------+--------------+-----+-----------+

값의 움직임
customer.email 변경 -> 1001, 1002, 1003 중 무엇을 바꿀지 결정해야 함
product.name 변경    -> 과거 주문 기록인지 현재 상품명인지 구분해야 함
unit_price 변경      -> 현재 판매가인지 주문 당시 가격인지 구분해야 함
```

이 표에서 `customer_email`은 같은 고객을 가리키는 여러 주문에 반복된다. 이메일이 고객의 현재 연락처라면 고객 테이블 하나가 소유해야 한다. 반대로 `unit_price`가 주문 당시 결제 금액을 계산하기 위한 가격이라면 상품 테이블의 현재 가격과 분리되어 주문 품목에 남아야 한다. 같은 "중복"처럼 보이는 두 값이 서로 다른 결론을 갖는 이유는 값의 소유자가 다르기 때문이다. 정규화는 중복을 보고 반사적으로 테이블을 쪼개는 작업이 아니라, "이 값은 어떤 사실을 대표하며, 그 사실의 변경이 어느 행들을 같이 움직여야 하는가"를 묻는 작업이다.

정규화를 제대로 이해하려면 키(key)를 먼저 이해해야 한다. 키는 단지 인덱스를 잘 타기 위한 컬럼이 아니다. 후보 키(candidate key)는 한 relation 안에서 한 tuple을 유일하게 식별할 수 있는 속성 집합이고, 기본 키(primary key)는 그 후보 중 이 테이블의 대표 식별자로 선택한 것이다. 외래 키(foreign key)는 한 테이블의 값이 다른 테이블의 대표 사실을 참조한다는 선언이다. PostgreSQL 공식 문서는 기본 키가 유일성과 `NOT NULL`을 함께 요구하고, 기본 키를 선언하면 참조하는 외래 키의 기본 대상 컬럼이 된다고 설명한다. MySQL 공식 문서도 외래 키 관계를 부모 테이블과 자식 테이블 사이의 참조 관계로 설명하며, 외래 키 제약이 관련 데이터를 일관되게 유지하는 데 쓰인다고 설명한다. 즉 키와 제약은 성능 힌트 이전에 데이터 의미를 기계가 검사할 수 있는 형태로 올린 것이다.

다음처럼 논리 사실을 나누면 각 테이블은 자기 책임을 갖는다.

```sql
CREATE TABLE customers (
  id BIGINT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  display_name VARCHAR(100) NOT NULL
);

CREATE TABLE products (
  id BIGINT PRIMARY KEY,
  sku VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(200) NOT NULL,
  current_price NUMERIC(12, 2) NOT NULL CHECK (current_price >= 0)
);

CREATE TABLE orders (
  id BIGINT PRIMARY KEY,
  customer_id BIGINT NOT NULL REFERENCES customers (id),
  ordered_at TIMESTAMP NOT NULL,
  status VARCHAR(30) NOT NULL
);

CREATE TABLE order_items (
  order_id BIGINT NOT NULL REFERENCES orders (id),
  line_no INTEGER NOT NULL,
  product_id BIGINT NOT NULL REFERENCES products (id),
  ordered_product_name VARCHAR(200) NOT NULL,
  ordered_unit_price NUMERIC(12, 2) NOT NULL CHECK (ordered_unit_price >= 0),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (order_id, line_no)
);
```

여기서 `products.current_price`와 `order_items.ordered_unit_price`는 일부러 분리했다. 현재 판매가는 상품의 현재 상태이고, 주문 당시 단가는 주문 사건의 기록이다. `ordered_product_name`도 겉으로는 상품명 중복이지만, 영수증이나 환불 감사에서 "그때 고객에게 어떤 이름으로 팔렸는가"가 필요하다면 주문 품목의 사실이다. 반대로 고객의 현재 이메일을 매 주문 품목에 넣는 것은 보통 위험하다. 과거 주문 알림 수신 주소를 보존해야 한다면 `orders.contact_email_at_order`처럼 이름으로 시점 의미를 드러내고, 현재 연락처와 다른 사실임을 분명히 해야 한다. 이런 이름은 문서 장식이 아니라 운영 중 실수를 줄이는 안전장치다.

정규화가 막는 대표 이상은 세 가지다. 갱신 이상은 같은 사실이 여러 곳에 있어 한쪽만 바뀌는 문제다. 삽입 이상은 어떤 사실을 저장하려면 아직 존재하지 않는 다른 사실까지 억지로 만들어야 하는 문제다. 삭제 이상은 한 사실을 지우다가 다른 독립 사실까지 사라지는 문제다. `orders_flat`에 상품 정보를 같이 넣으면 아직 주문되지 않은 상품을 저장할 수 없거나, 마지막 주문을 삭제하면서 상품 정의까지 잃는 이상이 생긴다. 고객 이메일도 여러 주문 행에 반복되면 일부 주문에는 새 이메일, 일부 주문에는 옛 이메일이 남아 "현재 고객 이메일"이라는 질문에 테이블이 하나의 답을 주지 못한다.

```text
정규화 전후의 사실 소유자

Before: orders_flat
  order_id=1001 owns customer_email? product_name? unit_price? status?
  -> 한 행이 너무 많은 사실을 소유하려고 해서 변경 규칙이 충돌한다.

After:
  customers(id=7)        -> 현재 고객 이메일과 표시 이름을 소유
  products(id=501)       -> 현재 상품명과 현재 판매가를 소유
  orders(id=1001)        -> 주문 사건, 주문 고객, 주문 시각, 주문 상태를 소유
  order_items(1001, 1)   -> 주문 안의 1번 품목, 주문 당시 이름/가격/수량을 소유

변경 trace:
  고객 이메일 변경
    UPDATE customers SET email='new@shop.test' WHERE id=7;
    -> 과거 주문 금액과 품목 기록은 움직이지 않음

  상품 현재 가격 변경
    UPDATE products SET current_price=85000 WHERE id=501;
    -> 새 주문 기본 가격만 바뀜
    -> 이미 생성된 order_items.ordered_unit_price는 움직이지 않음
```

이 trace가 중요한 이유는 정규화가 "조회할 때 join을 늘리는 선택"이 아니라 "변경의 반경을 좁히는 선택"이라는 점을 보여 주기 때문이다. 운영 장애에서 더 무서운 것은 join이 하나 늘어나는 비용보다, 어느 행이 진짜인지 알 수 없는 상태다. 읽기 성능은 인덱스, 캐시, 물리 설계, 집계 테이블, read model로 다룰 수 있지만, 잘못된 사실 소유권은 나중에 쿼리 튜닝으로 복구하기 어렵다.

정규형을 교과서식으로 나열하면 1정규형, 2정규형, 3정규형, BCNF 같은 이름이 나온다. 이 문서의 목적은 정규형 증명 문제를 푸는 것이 아니라 실무 설계를 다시 설명할 수 있게 만드는 것이므로, 각 이름을 "어떤 종류의 의존을 제거하려는가"로 읽으면 충분하다. 1정규형은 한 칸에 반복 목록을 넣지 말고 relation의 값 단위를 분명히 하라는 출발점이다. `orders.product_ids = '501,502,503'`처럼 쉼표로 묶은 값은 SQL이 각 상품을 독립 행으로 검사하고 참조하고 집계하기 어렵게 만든다. 2정규형은 복합 키의 일부에만 의존하는 값을 분리하라는 신호다. `(order_id, line_no)`가 품목을 식별하는데 `customer_email`이 `order_id`에만 의존한다면 품목 테이블에 있으면 안 된다. 3정규형은 키가 아닌 값이 다른 키 아닌 값을 통해 결정되는 전이 의존을 줄이라는 신호다. `orders.customer_id -> customers.grade -> discount_rate`라면 주문 행에 현재 등급 할인율을 그대로 두는 순간 등급 정책 변경과 주문 기록의 의미가 충돌한다.

실무에서는 이 원칙을 너무 기계적으로 적용해도 문제가 생긴다. 모든 코드 테이블을 극단적으로 쪼개고, 작은 enum 값 하나마다 외래 키를 걸고, 읽기 경로마다 열 개 이상의 join을 강제하면 시스템은 논리적으로는 아름다워도 장애 대응이 어려워질 수 있다. 하지만 이것은 "정규화가 성능의 반대말"이라는 뜻이 아니다. 더 정확한 표현은 "정규화는 쓰기 일관성과 사실 소유권을 안정시키고, 읽기 경로의 비용은 별도 설계로 해결한다"다. 반정규화는 이 별도 설계 중 하나다. 좋은 반정규화는 다음 네 가지를 문서와 코드에 남긴다. 어떤 정규화된 원본에서 나온 값인지, 언제 복사되는지, 원본이 바뀌면 어떻게 갱신되거나 일부러 갱신되지 않는지, 불일치를 어떻게 관측하고 복구하는지다.

예를 들어 주문 목록 화면이 고객명, 결제금액, 최근 배송상태를 한 번에 보여 줘야 한다면 매 요청마다 `orders`, `customers`, `payments`, `shipments`, `order_items`를 모두 join하는 대신 `order_summary` 같은 읽기 모델을 둘 수 있다. 이 읽기 모델은 원본 사실이 아니라 조회 편의를 위한 projection, 즉 원본에서 투영한 요약이다. 한국어로는 "읽기용 요약 테이블" 정도로 먼저 이해하면 된다. 이 테이블은 정규화 모델을 대체하지 않는다. 원본 테이블이 쓰기와 진실의 기준이고, 요약 테이블은 화면과 검색을 위해 원본 일부를 복사한다. 따라서 요약 테이블에는 생성 규칙, 지연 허용 범위, 재생성 방법, 검증 쿼리가 같이 있어야 한다.

```sql
CREATE TABLE order_summary (
  order_id BIGINT PRIMARY KEY,
  customer_id BIGINT NOT NULL,
  customer_display_name VARCHAR(100) NOT NULL,
  total_amount NUMERIC(12, 2) NOT NULL,
  item_count INTEGER NOT NULL,
  latest_delivery_status VARCHAR(30) NOT NULL,
  source_updated_at TIMESTAMP NOT NULL,
  summary_refreshed_at TIMESTAMP NOT NULL
);

-- 검증용 관측 쿼리 예시
SELECT s.order_id,
       s.total_amount AS summary_total,
       SUM(i.ordered_unit_price * i.quantity) AS source_total
FROM order_summary s
JOIN order_items i ON i.order_id = s.order_id
GROUP BY s.order_id, s.total_amount
HAVING s.total_amount <> SUM(i.ordered_unit_price * i.quantity);
```

이 쿼리는 반정규화를 안전하게 쓰는 감각을 보여 준다. 반정규화된 `summary_total`이 원본 품목 합계와 다르면 이 행은 성능 최적화 산출물이 아니라 잘못된 데이터다. 운영에서는 이 검증을 배치 점검, 모니터링 쿼리, 데이터 품질 대시보드, 재생성 job의 사전 점검으로 둘 수 있다. 반정규화가 위험해지는 지점은 요약 테이블을 만들었기 때문이 아니라, 요약 테이블이 원본인지 복사본인지, 어떤 지연을 허용하는지, 깨졌을 때 어떤 쿼리로 찾는지 없는 상태로 운영하기 때문이다.

제약 조건(constraint)은 이 설계 의도를 데이터베이스에 맡기는 방법이다. PostgreSQL 공식 문서는 제약 조건을 데이터 타입만으로는 표현하기 어려운 더 세밀한 제한을 테이블에 선언하는 방법으로 설명하고, 위반하는 값을 저장하려 하면 오류가 난다고 설명한다. `CHECK`, `NOT NULL`, `UNIQUE`, `PRIMARY KEY`, `FOREIGN KEY`, `EXCLUDE` 같은 제약은 각각 다른 종류의 불변식을 담당한다. 불변식(invariant)은 상태가 바뀌어도 계속 참이어야 하는 조건이다. 쉬운 말로는 "이 테이블은 어떤 순간에도 이 규칙을 어기면 안 된다"는 약속이다. 애플리케이션 코드에서도 검사를 할 수 있지만, 코드 검사는 모든 쓰기 경로가 같은 코드를 통과한다는 전제가 필요하다. 데이터베이스 제약은 배치, 운영자 SQL, 다른 서비스, 마이그레이션 스크립트처럼 여러 경로가 들어와도 마지막 문에서 같은 규칙을 검사한다.

`CHECK`는 한 행 안의 값 관계를 표현하는 데 강하다. 상품 가격이 0 이상이어야 한다거나, 할인 가격이 정가보다 낮아야 한다거나, 수량이 양수여야 한다는 규칙은 `CHECK`로 잘 드러난다. 다만 PostgreSQL은 `CHECK`가 다른 행이나 다른 테이블의 데이터를 참조하는 식을 지속적인 일관성 보장으로 쓰면 안 된다고 경고한다. `CHECK`는 새로 삽입되거나 수정되는 행을 검사하는 성격이기 때문에, 다른 행의 나중 변경까지 계속 추적하는 장치가 아니다. 이 제약을 어기면 단순 테스트에서는 동작해 보이더라도 덤프와 복원 같은 운영 상황에서 깨질 수 있다. 따라서 "한 고객의 활성 쿠폰은 하나만 있어야 한다"처럼 여러 행 사이의 규칙은 보통 `UNIQUE` 제약, 부분 unique index, 외래 키, trigger, 직렬화된 트랜잭션, 애플리케이션 락 같은 다른 장치와 함께 설계해야 한다.

`NOT NULL`은 컬럼이 `NULL`을 가질 수 없음을 선언한다. 이 제약은 단순해 보이지만 도메인 모델을 매우 강하게 만든다. `orders.customer_id`가 nullable이면 익명 주문, 임시 주문, 삭제된 고객 주문, 아직 고객 연결 전 주문 같은 여러 상태가 한 컬럼 안에 섞인다. 이 상태들이 실제 비즈니스에 필요하다면 별도의 상태 컬럼이나 별도 테이블로 의미를 드러내야 한다. 필요하지 않다면 `NOT NULL`로 막는 편이 낫다. PostgreSQL 문서는 대부분의 데이터베이스 설계에서 대다수 컬럼은 `NOT NULL`이어야 한다는 팁을 제공한다. 이 말은 모든 컬럼을 무조건 막으라는 뜻이 아니라, `NULL`이 의미 있는 도메인 상태인지 증명하지 못하면 숨은 분기와 쿼리 실수를 늘린다는 뜻으로 읽는 편이 실무에 맞다.

`UNIQUE`는 한 컬럼이나 컬럼 집합의 값 조합이 중복되지 않게 한다. 하지만 unique에서 `NULL`을 어떻게 다루는지는 DBMS마다 차이가 있다. PostgreSQL 공식 문서는 기본적으로 unique 제약에서 두 `NULL` 값이 같다고 간주되지 않으며, `NULLS NOT DISTINCT`로 동작을 바꿀 수 있다고 설명한다. 이 지점은 실무 함정이다. `users.email UNIQUE`만 보고 "이메일은 하나만 들어간다"고 생각했는데 이메일이 nullable이면 `NULL` 이메일 사용자는 여러 명 생길 수 있다. 이게 의도라면 괜찮다. 가입 전 임시 사용자에게 이메일이 없을 수 있기 때문이다. 하지만 "이메일이 없으면 가입 완료가 아니다"라는 도메인이라면 `email NOT NULL UNIQUE` 또는 상태별 부분 unique 전략이 필요하다.

외래 키는 참조 무결성(referential integrity)을 지킨다. 쉬운 말로는 "자식 행이 가리키는 부모 행이 실제로 존재해야 한다"는 규칙이다. MySQL 공식 문서의 문장처럼 부모 테이블은 초기 컬럼 값을 가진 테이블이고, 자식 테이블은 그 부모 컬럼 값을 참조하는 값을 가진 테이블이다. `order_items.product_id`가 `products.id`를 참조하면 존재하지 않는 상품을 주문 품목에 넣는 것을 막을 수 있다. 그러나 외래 키는 도메인 의미를 자동으로 완성하지 않는다. 부모 삭제 시 `RESTRICT`, `NO ACTION`, `CASCADE`, `SET NULL`, `SET DEFAULT` 중 어떤 행동을 택할지는 관계의 의미에 따라 달라진다. 주문이 삭제되면 주문 품목도 같이 없어져야 할 수 있지만, 상품이 삭제된다고 과거 주문 품목이 같이 지워지면 감사와 정산 기록이 망가질 수 있다. 이 때문에 삭제 정책은 "편하게 cascade"가 아니라 "자식이 부모 없이는 존재할 수 없는 구성 요소인가, 아니면 과거 사건 기록인가"로 판단해야 한다.

```text
ON DELETE 선택 trace

orders(id=1001)
  |
  +-- order_items(order_id=1001, line_no=1)

case A: 주문 취소가 아니라 테스트 주문 데이터 삭제
  DELETE FROM orders WHERE id=1001;
  -> order_items는 orders의 구성 요소이므로 CASCADE가 자연스러울 수 있음

case B: 상품 판매 중지
  DELETE FROM products WHERE id=501;
  -> 과거 order_items까지 삭제되면 정산/감사 기록이 사라짐
  -> RESTRICT로 막고 product.status='DISCONTINUED' 같은 상태 전환이 더 안전할 수 있음

case C: 선택 담당자 삭제
  DELETE FROM product_managers WHERE id=9;
  -> 상품은 계속 존재하고 담당자만 선택 정보라면 SET NULL이 의미 있을 수 있음
```

제약 조건은 이름도 중요하다. 데이터베이스가 자동으로 만든 이름은 오류 메시지와 마이그레이션에서 읽기 어렵다. `CONSTRAINT valid_discount CHECK (price > discounted_price)`처럼 이름을 주면 실패 로그가 바로 도메인 규칙을 말해 준다. `fk_order_items_orders`, `fk_order_items_products`, `uq_users_email`, `ck_order_items_quantity_positive` 같은 이름은 길지만 운영 로그와 migration diff에서 힘을 발휘한다. 좋은 제약 이름은 테이블, 제약 종류, 도메인 규칙을 드러낸다. 나쁜 이름은 `constraint_1`, `chk_value`, `fk_id`처럼 다음 사람이 어떤 규칙이 깨졌는지 다시 `SHOW CREATE TABLE`부터 찾아야 하게 만든다.

여기서 senior practical trap, 즉 실무에서 숙련자도 밟는 함정은 제약을 "나중에 데이터가 깨지면 붙이면 되는 것"으로 보는 태도다. 이미 깨진 데이터가 들어간 뒤 `NOT NULL`, `UNIQUE`, `FOREIGN KEY`를 추가하면 migration은 실패하거나 긴 테이블 검사를 수행한다. 더 위험한 경우는 운영자가 실패를 피하려고 제약 검사를 꺼 두거나, 애플리케이션 코드에만 검사를 남긴 채 데이터베이스 규칙을 포기하는 것이다. 그러면 당장은 배포가 되지만, 나중에 어떤 배치나 운영 SQL이 우회했는지 알 수 없는 상태가 된다. 제약은 새 기능의 마지막 장식이 아니라 데이터 모델을 만들 때 함께 설계해야 하는 실패 방지 장치다.

제약을 추가할 때는 먼저 현재 데이터가 규칙을 만족하는지 관측해야 한다. 예를 들어 `order_items.quantity NOT NULL CHECK (quantity > 0)`를 붙이고 싶다면, migration 전에 다음 쿼리로 위반 후보를 찾아야 한다.

```sql
-- 수량 제약 추가 전 관측
SELECT order_id, line_no, quantity
FROM order_items
WHERE quantity IS NULL OR quantity <= 0
ORDER BY order_id, line_no
LIMIT 50;

-- 주문 품목의 부모 주문 누락 관측
SELECT i.order_id, i.line_no
FROM order_items i
LEFT JOIN orders o ON o.id = i.order_id
WHERE o.id IS NULL
LIMIT 50;

-- 이메일 unique 전 중복 관측
SELECT email, COUNT(*) AS cnt
FROM customers
WHERE email IS NOT NULL
GROUP BY email
HAVING COUNT(*) > 1
ORDER BY cnt DESC, email
LIMIT 50;
```

PASS 신호는 결과가 0행이거나, 발견된 행에 대해 정리 migration과 보정 정책이 먼저 준비되는 것이다. FAIL 신호는 "대부분 괜찮아 보인다"는 감각만으로 제약을 추가하는 것이다. 운영에서는 이 관측 쿼리를 migration 전 검증 단계, 배포 체크리스트, 또는 CI의 fixture 검증으로 옮길 수 있다. 중요한 점은 제약 추가가 단일 DDL이 아니라 "현재 데이터 관측 -> 위반 데이터 정리 -> 제약 추가 -> 이후 쓰기 경로 검증"이라는 흐름이라는 것이다.

정규화와 제약 조건은 인덱스와도 연결된다. PostgreSQL 문서는 unique 제약과 primary key가 자동으로 unique B-tree index를 만든다고 설명한다. MySQL InnoDB도 외래 키를 위해 참조 컬럼의 인덱스 조건과 제한을 갖는다. 따라서 제약을 붙이면 논리 규칙뿐 아니라 물리 구조가 생기고, 쓰기 비용과 lock 경합도 달라진다. `UNIQUE (email)`은 중복을 막는 동시에 이메일 조회를 빠르게 할 수 있지만, 이메일 변경이 많은 시스템에서는 쓰기 경합과 인덱스 유지 비용이 생긴다. `FOREIGN KEY`는 잘못된 참조를 막지만 부모/자식 변경 시 추가 검사를 수행하고, 삭제 정책에 따라 더 많은 행을 잠글 수 있다. 이 비용 때문에 제약을 버리는 것이 아니라, 어떤 규칙을 DB에 맡기고 어떤 규칙을 별도 프로세스에 맡길지 근거를 남겨야 한다.

반정규화와 제약은 서로 배척하지 않는다. 오히려 반정규화가 들어갈수록 제약과 관측이 더 필요하다. 예를 들어 `order_summary.total_amount`는 원본 합계와 일치해야 한다. 이 규칙은 단순 `CHECK`로 표현하기 어렵다. 한 행 안의 값이 아니라 `order_items` 여러 행의 합계와 비교해야 하기 때문이다. 그러면 선택지는 몇 가지다. 쓰기 트랜잭션 안에서 원본과 summary를 함께 갱신하고 실패 시 롤백한다. 이벤트 기반으로 비동기 갱신하되 지연 허용 시간을 정하고 불일치 탐지 쿼리를 둔다. 주기적으로 summary를 재생성하고 배포 전후 비교를 한다. 어떤 선택이든 "정규화된 원본이 무엇이고, 복제본이 언제 얼마나 늦어져도 되는가"가 명시되어야 한다.

```text
반정규화된 summary 갱신 방식 비교

방식 1: 같은 트랜잭션에서 원본과 summary 동시 갱신
  장점 -> 읽기 모델이 거의 즉시 일치
  비용 -> 주문 쓰기 경로가 무거워지고 실패 반경이 커짐
  관측 -> 트랜잭션 실패율, row lock wait, summary 불일치 0행 확인

방식 2: 이벤트로 summary 비동기 갱신
  장점 -> 주문 쓰기 경로가 가벼움
  비용 -> 잠깐 낡은 summary가 보일 수 있음
  관측 -> 이벤트 lag, dead-letter queue, 불일치 행 수, 재처리 성공률

방식 3: 배치로 summary 재계산
  장점 -> 규칙이 단순하고 복구가 쉬움
  비용 -> 실시간 화면에는 부적합할 수 있음
  관측 -> 마지막 재계산 시각, 변경량, 원본 대비 차이
```

정규화를 설계할 때 자주 나오는 오해는 "join이 있으면 느리니 처음부터 한 테이블로 둔다"는 것이다. 이 오해는 논리 모델과 물리 실행 모델을 섞는다. join이 비용을 만들 수 있는 것은 맞지만, join 비용은 인덱스, 선택도, 통계, join 알고리즘, 캐시 상태에 따라 달라진다. 반면 잘못 중복된 사실은 쿼리 플랜과 무관하게 데이터 의미를 깨뜨린다. 먼저 사실 소유권을 올바르게 잡고, 실제 조회가 느리면 `EXPLAIN`, 실행 시간, row count, buffer read 같은 관측으로 병목을 확인한 뒤 읽기 모델이나 인덱스를 설계하는 순서가 안전하다. "느릴 것 같다"는 추측만으로 정규화를 포기하면, 나중에 성능 문제도 해결하지 못하고 데이터 정합성 문제까지 떠안는다.

또 다른 오해는 "외래 키는 마이크로서비스에서 무조건 쓰면 안 된다"는 식의 단정이다. 서비스별 데이터베이스를 엄격히 나누고 다른 서비스의 테이블을 직접 참조하지 않는 구조에서는 DB 외래 키를 걸 수 없거나 걸지 않는 것이 맞을 수 있다. 하지만 같은 서비스의 같은 데이터베이스 안에서 주문과 주문 품목처럼 강한 생명주기 관계가 있는데도 막연히 "운영에서 불편하다"는 이유로 외래 키를 없애면, 고아 행(orphan row)을 찾고 지우는 책임이 모두 애플리케이션과 운영 배치로 이동한다. 외래 키를 쓰지 않는 선택도 가능하지만, 그 선택은 "어디서 참조 무결성을 관측하고 복구하는가"까지 포함해야 한다.

스키마 설계의 검증은 ERD를 보는 것만으로 끝나지 않는다. 최소한 다음 질문에 SQL로 답할 수 있어야 한다. 각 테이블의 기본 키는 무엇인가. 후보 키 중 대표 키를 왜 이것으로 정했는가. nullable 컬럼은 실제 도메인 상태를 나타내는가. unique에서 `NULL` 중복 허용이 의도인가. 외래 키 삭제 정책은 부모와 자식의 생명주기를 반영하는가. 반정규화된 값은 어떤 원본에서 왔고 어떻게 불일치를 찾는가. 이 질문에 답하지 못하면 설계는 아직 문서로는 그럴듯하지만 운영에서 replay 가능한 지식이 아니다.

```sql
-- MySQL / PostgreSQL에서 개념적으로 확인할 수 있는 관측 축
-- 실제 catalog 이름은 DBMS별로 다르므로 운영 DB에 맞게 조정한다.

-- 1. 테이블별 primary key 확인
SELECT table_name, constraint_name
FROM information_schema.table_constraints
WHERE constraint_type = 'PRIMARY KEY'
ORDER BY table_name;

-- 2. nullable인데 상태 의미가 불분명한 후보 찾기
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE is_nullable = 'YES'
  AND table_schema NOT IN ('information_schema', 'pg_catalog', 'mysql', 'performance_schema')
ORDER BY table_name, ordinal_position;

-- 3. foreign key 목록 확인
SELECT table_name, constraint_name
FROM information_schema.table_constraints
WHERE constraint_type = 'FOREIGN KEY'
ORDER BY table_name, constraint_name;
```

이 관측 쿼리의 목적은 "제약이 많으면 좋다"를 증명하는 것이 아니다. 목적은 데이터베이스가 실제로 어떤 규칙을 알고 있는지 확인하는 것이다. 문서에는 `orders.customer_id`가 필수라고 적혀 있는데 catalog에서는 nullable이면 문서와 실행 경계가 다르다. 코드에는 `email`이 유일하다고 가정하는데 DB에는 unique가 없으면 동시 가입 요청에서 중복이 생길 수 있다. 외래 키가 없으면 삭제 정책이 코드나 운영 절차에 있어야 한다. 관측 결과는 설계와 운영 책임의 차이를 드러내는 증거다.

실무에서는 정규화 수준을 결정할 때 읽기/쓰기 비율, 변경 빈도, 감사 요구, 데이터 보존 정책, 장애 복구 방식을 함께 본다. 현재 값만 의미 있는 프로필 정보와 과거 시점이 중요한 결제 정보는 다르게 설계한다. 자주 바뀌는 상품 설명과 주문 당시 계약 문구도 다르게 설계한다. 이벤트 소싱을 쓰는 시스템에서는 원본 event log와 projection table의 관계가 정규화/반정규화 논의를 다른 형태로 반복한다. 데이터 웨어하우스에서는 star schema처럼 분석을 위해 의도적으로 중복을 둔다. 하지만 어느 경우에도 "무엇이 원본이고 무엇이 파생인가"라는 질문은 사라지지 않는다.

키 선택에서도 같은 사고가 필요하다. 실무에서는 자연 키(natural key)와 대리 키(surrogate key)를 자주 비교한다. 자연 키는 주민등록번호, 사업자등록번호, ISO 코드, 상품 SKU처럼 도메인 바깥에서도 의미가 있는 식별자다. 대리 키는 `BIGINT AUTO_INCREMENT`, sequence, UUID처럼 시스템이 부여한 식별자다. 자연 키는 사람이 이해하기 쉽고 중복을 막는 규칙을 직접 드러내지만, 도메인 정책이 바뀌거나 개인정보/보안 요구가 생기면 변경 비용이 크다. 대리 키는 내부 참조가 안정적이고 join이 단순하지만, 도메인상 중복을 막으려면 별도 unique 제약이 필요하다. 따라서 `customers.id`를 대리 키로 쓰더라도 `customers.email`이나 `customers.external_customer_no`가 도메인상 유일해야 한다면 unique 제약을 따로 둬야 한다. 대리 키를 만들었다고 후보 키 검토가 사라지는 것은 아니다.

```text
대리 키만 있고 도메인 unique가 없는 경우

customers
+----+------------------+--------------+
| id | email            | display_name |
+----+------------------+--------------+
|  1 | a@shop.test      | A            |
|  2 | a@shop.test      | A duplicate  |
+----+------------------+--------------+

애플리케이션 코드의 숨은 가정:
  SELECT * FROM customers WHERE email='a@shop.test';
  -> 한 명만 나올 것이라고 가정했지만 DB는 두 명을 허용했다.

보강:
  ALTER TABLE customers ADD CONSTRAINT uq_customers_email UNIQUE (email);
  단, email nullable 정책과 기존 중복 정리 결과를 먼저 검증해야 한다.
```

복합 키(composite key)도 무조건 피할 대상이 아니다. `order_items`의 `(order_id, line_no)`는 주문 안에서 몇 번째 품목인지를 자연스럽게 드러낸다. 여기에 별도 `id`를 붙일 수도 있지만, 그러면 `order_id, line_no`의 유일성은 여전히 별도 제약으로 남겨야 한다. 반대로 외부 API나 다른 테이블이 품목을 자주 참조하고, line 번호가 재정렬될 수 있다면 대리 키가 더 안전할 수 있다. 기준은 "짧은 키가 예쁜가"가 아니라 "이 식별자가 도메인 변화 속에서 안정적인가, 참조하는 소비자가 무엇을 필요로 하는가"다. 특히 결제, 정산, 회계처럼 감사가 중요한 테이블에서는 한 번 발급된 식별자가 나중에 재사용되거나 의미가 바뀌면 안 된다. 키는 단지 row locator가 아니라 감사 추적의 anchor가 된다.

제약 조건을 어디에 둘지도 설계 판단이다. 예를 들어 주문 상태가 `READY`, `PAID`, `CANCELLED`, `REFUNDED` 중 하나여야 한다면 `CHECK (status IN (...))`로 둘 수 있고, 별도 `order_status_codes` 테이블을 만들고 외래 키로 참조할 수도 있고, 애플리케이션 enum만 둘 수도 있다. 상태 종류가 작고 배포와 함께만 바뀌며 DBMS가 check를 잘 지원한다면 `CHECK`가 단순하다. 운영자가 상태 코드를 데이터로 추가해야 하거나, 상태별 표시명/정렬/전이 정책을 함께 관리해야 한다면 reference table이 낫다. 애플리케이션 enum만 두는 방식은 가장 가볍지만 DB 밖 쓰기 경로가 생기면 깨질 수 있다. 같은 상태 코드라도 변경 주체와 관측 책임에 따라 다른 설계가 이길 수 있다.

```sql
-- 작은 고정 vocabulary는 CHECK로 충분할 수 있다.
ALTER TABLE orders
  ADD CONSTRAINT ck_orders_status
  CHECK (status IN ('READY', 'PAID', 'CANCELLED', 'REFUNDED'));

-- 운영 데이터로 관리되는 vocabulary는 reference table이 더 낫다.
CREATE TABLE order_status_codes (
  code VARCHAR(30) PRIMARY KEY,
  display_name VARCHAR(100) NOT NULL,
  terminal BOOLEAN NOT NULL
);

ALTER TABLE orders
  ADD CONSTRAINT fk_orders_status
  FOREIGN KEY (status) REFERENCES order_status_codes (code);
```

이 선택의 함정은 "정답 패턴"을 외우는 것이다. `CHECK`가 항상 낫지도 않고 reference table이 항상 과한 것도 아니다. 상태 코드가 release마다 코드와 함께 바뀐다면 reference table은 운영자가 잘못 수정할 수 있는 새로운 표면을 만든다. 반대로 상태 코드가 정산, 고객센터, 통계 화면과 함께 관리되는 운영 데이터라면 `CHECK`를 바꿀 때마다 DDL migration을 해야 하는 것이 부담이 된다. 제약은 데이터를 막는 문법이 아니라 변경 권한을 어디에 둘지 정하는 방식이다.

시간에 따라 의미가 바뀌는 제약은 특히 조심해야 한다. 예를 들어 "한 사용자는 활성 구독을 하나만 가질 수 있다"는 규칙은 `(user_id, status)` unique만으로는 부족할 수 있다. `status='ACTIVE'`인 행만 하나여야 하기 때문이다. PostgreSQL에서는 partial unique index로 표현할 수 있지만, 모든 DBMS가 같은 방식으로 지원하지 않는다. MySQL에서는 generated column, 별도 active table, transaction 안의 locking read, 애플리케이션 단의 winner election 같은 다른 설계를 검토해야 할 수 있다. 이때도 먼저 보호할 불변식을 말로 고정해야 한다. "unique index를 붙인다"가 목표가 아니라 "같은 사용자에게 동시에 두 개의 활성 구독이 관측되면 안 된다"가 목표다.

```text
불변식 먼저 쓰기

나쁜 목표 표현:
  subscriptions에 unique index를 붙인다.

좋은 목표 표현:
  같은 user_id에 대해 ACTIVE 상태 subscription은 어느 시점에도 최대 1개여야 한다.

설계 후보:
  A. partial unique index where status='ACTIVE'
  B. active_subscriptions(user_id primary key, subscription_id unique) 별도 테이블
  C. user_id 단위 advisory lock 또는 SELECT ... FOR UPDATE 후 상태 전환

검증:
  SELECT user_id, COUNT(*)
  FROM subscriptions
  WHERE status='ACTIVE'
  GROUP BY user_id
  HAVING COUNT(*) > 1;
```

이렇게 쓰면 DBMS별 기능 차이가 보여도 목표가 흔들리지 않는다. PostgreSQL partial unique index가 가능하면 A가 단순할 수 있고, 여러 DBMS를 지원해야 하거나 활성 구독만 빠르게 조회해야 하면 B가 더 명시적일 수 있다. C는 코드 책임이 커지고 모든 쓰기 경로가 같은 locking 규칙을 따라야 하므로 더 조심해야 한다. 설계 비교의 기준은 "내가 아는 기능인가"가 아니라 불변식을 누가 가장 안정적으로 지키고, 실패했을 때 어떤 관측으로 찾을 수 있는가다.

제약 조건은 migration과도 연결되므로 설계 문서에 "처음부터 붙일 제약"과 "데이터 정리 후 붙일 제약"을 나누는 편이 좋다. 새 테이블을 만들 때는 `NOT NULL`, `CHECK`, `UNIQUE`, `FOREIGN KEY`를 처음부터 넣는 것이 대체로 쉽다. 기존 큰 테이블에 붙일 때는 현재 데이터 위반 여부, DDL 알고리즘, lock, index 생성 비용을 봐야 한다. 특히 외래 키를 나중에 추가하면 자식 테이블의 모든 기존 행이 부모를 참조하는지 검사해야 한다. 이 검사는 데이터 품질을 올리는 좋은 기회이지만, 운영 시간에 갑자기 실행하면 장애 요인이 된다. 따라서 "제약은 좋은 것"이라는 결론만으로 부족하고, 언제 어떻게 붙일지까지 migration 계획으로 이어져야 한다.

정리하면 DU23의 핵심은 다음 하나로 압축된다. 정규화는 중복 제거 규칙이 아니라 사실 소유권과 변경 반경을 설계하는 방법이고, 반정규화는 그 소유권을 이해한 뒤 읽기/감사/운영 요구 때문에 복제본을 만들고 관측 책임을 붙이는 방법이다. 키와 제약 조건은 이 설계를 데이터베이스가 직접 검사하게 만드는 장치다. PostgreSQL과 MySQL 공식 문서가 제약을 데이터 일관성과 참조 무결성의 실행 규칙으로 설명하는 이유도 여기에 있다. 좋은 스키마는 ERD 그림이 예쁜 스키마가 아니라, 어떤 값이 어디서 참이고, 어떤 변경이 어디까지 퍼지고, 어떤 쿼리로 깨짐을 확인할 수 있는지까지 설명되는 스키마다.

공식 및 1차 근거로는 IBM Research의 Codd 관계형 모델 논문 소개(`https://research.ibm.com/publications/a-relational-model-of-data-for-large-shared-data-banks`), PostgreSQL 18 `DDL Constraints`(`https://www.postgresql.org/docs/current/ddl-constraints.html`), MySQL 8.4 `FOREIGN KEY Constraints`(`https://dev.mysql.com/doc/refman/8.4/en/create-table-foreign-keys.html`)를 사용했다. 특히 PostgreSQL 문서의 `CHECK`, `NOT NULL`, `UNIQUE`, `PRIMARY KEY`, `FOREIGN KEY` 설명과 MySQL 문서의 부모/자식 외래 키 설명은 제약 조건을 단순 문법이 아니라 불변식과 참조 무결성으로 읽는 근거가 된다. 이 문서의 예제 SQL은 학습용 최소 모델이며, 실제 운영 DB에서는 DBMS 버전, isolation, DDL 알고리즘, catalog 차이를 별도로 확인해야 한다.

## migration, schema diff, online DDL, Flyway

스키마 마이그레이션은 `ALTER TABLE` 몇 줄을 배포하는 일이 아니라 이미 데이터가 들어 있는 물리 구조, 그 구조를 읽고 쓰는 애플리케이션, replication, lock, rollback, 관측 절차를 같은 시간축 위에 올리는 일이다. 개발 환경에서 컬럼 하나를 추가하는 명령은 즉시 끝나 보이지만, 운영 환경에서는 그 컬럼을 추가하기 위해 테이블 메타데이터만 바꾸는지, 테이블을 다시 쓰는지, 기존 행을 검증하는지, DML을 허용하는지, replica lag를 만드는지, 실패하면 어느 지점까지 되돌릴 수 있는지가 모두 달라진다. 그래서 migration을 이해할 때는 "DDL을 실행했다"가 아니라 "DDL이 어떤 단계에서 어떤 lock과 데이터 이동을 만들고, 애플리케이션 버전은 그 사이 어떤 스키마를 기대하는가"를 추적해야 한다.

먼저 schema diff부터 보자. 이 저장소의 `database/db_diff.md`와 `database/examples/db_diff/*`는 개발 DB와 운영 기준 DB를 비교하는 작은 seed를 남겨 둔다. 예제는 Liquibase diff를 사용해 `dev`와 `prod` MySQL 컨테이너의 스키마를 비교한다. `dev`에는 `users.email`, `posts.updated_at`, `comments.updated_at`, `only_at_dev`가 있고, `prod`에는 `posts.only_at_prod`가 있다. diff 결과는 `Missing Column(s)`, `Unexpected Column(s)`, `Changed Column(s)`, `Unexpected Table(s)` 같은 목록으로 나온다. 이 결과는 migration script가 아니다. 이 결과는 "두 환경의 구조가 어디서 갈라졌는지"를 보여 주는 관측 자료다. 관측 자료를 곧바로 운영 DDL로 바꾸면 위험하다. 왜냐하면 어떤 차이는 의도된 다음 버전이고, 어떤 차이는 운영 hotfix이고, 어떤 차이는 개발자의 실험일 수 있기 때문이다.

```text
로컬 seed의 schema diff 흐름

prod 기준(reference)                 dev 비교 대상(comparison)
---------------------                -------------------------
users(id, username, password_hash,   users(id, username, password_hash,
      created_at)                         email, created_at)

posts(id, user_id, title, body,      posts(id, user_id, title, body,
      only_at_prod, created_at)           created_at, updated_at)

comments(id, post_id, comment,       comments(id, post_id, comment,
         created_at)                          created_at, updated_at)

                                      only_at_dev(id, ref_id, extra, created_at)

diff output
  Missing Column(s)    -> prod에는 있는데 dev에는 없음: posts.only_at_prod
  Unexpected Column(s) -> dev에는 있는데 prod에는 없음: users.email, posts.updated_at ...
  Changed Column(s)    -> 컬럼 순서 차이: posts.created_at, users.created_at
  Unexpected Table(s)  -> dev에만 있음: only_at_dev
```

이 trace에서 `Missing`과 `Unexpected`은 절대적 좋고 나쁨이 아니다. Liquibase 명령에서 `--referenceUrl`이 기준이고 `--url`이 비교 대상이므로, 기준에만 있으면 `Missing`, 비교 대상에만 있으면 `Unexpected`로 보인다. 같은 결과라도 기준을 바꾸면 표현이 바뀐다. 운영에서 이 점을 헷갈리면 "prod에만 있는 컬럼을 지워야 한다"거나 "dev에만 있는 실험 테이블을 운영에 추가해야 한다"는 잘못된 결론으로 갈 수 있다. 따라서 diff를 읽을 때 첫 줄의 `Reference Database`와 `Comparison Database`를 먼저 확인해야 한다. 이 저장소 예제의 `diff_db.sh`는 `referenceUrl=jdbc:mysql://localhost:3308/some_service`, `url=jdbc:mysql://localhost:3307/some_service`로 prod를 기준, dev를 비교 대상으로 둔다.

schema diff의 좋은 사용법은 세 단계다. 첫째, 두 환경의 차이를 관측한다. 둘째, 각 차이를 의도된 product change, 이미 반영된 운영 hotfix, 실험 부산물, tool noise로 분류한다. 셋째, 실제 migration script는 이 분류 결과를 바탕으로 사람이 작성하거나 검토한다. Flyway의 migration 기반 접근은 이 세 번째 지점에 가깝다. Redgate Flyway 공식 문서는 migration이 개발 DB의 증분 변경을 포착한 SQL script이고, version control에 추적되어 다른 환경에 같은 순서로 배포된다고 설명한다. Flyway가 핵심으로 삼는 것은 "현재 DB를 비교해서 즉석에서 마음대로 바꾼다"가 아니라, 버전이 붙은 script를 순서대로 적용하고 schema history로 무엇이 적용되었는지 추적하는 흐름이다.

Flyway의 versioned migration은 한 번 적용되는 순서 있는 변경이다. 공식 문서에 따르면 versioned migration은 대상 DB에 버전 순서대로 정확히 한 번 적용되고, `flyway_schema_history` 테이블이 어떤 migration이 적용되었는지와 checksum을 저장한다. checksum은 이미 적용된 파일이 나중에 바뀌었는지 감지하는 장치다. 이 구조가 중요한 이유는 운영 DB가 "현재 코드의 SQL 파일을 다시 읽으면 같은 상태가 된다"는 식으로 움직이지 않기 때문이다. 운영 DB는 시간이 흐르며 V1, V2, V3를 적용한 누적 상태다. 이미 prod에 적용된 V2 파일을 고치면, 새로 만든 local DB와 이미 V2를 지나온 prod DB가 서로 다른 역사를 갖게 된다. Flyway validate는 이 차이를 실패로 드러낸다.

```text
Flyway versioned migration timeline

git repository
  V1__create_users.sql
  V2__add_users_email.sql
  V3__create_posts.sql

empty database
  flyway migrate
    -> apply V1
    -> record V1 checksum in flyway_schema_history
    -> apply V2
    -> record V2 checksum
    -> apply V3
    -> record V3 checksum

permanent downstream database(prod)
  already has V1, V2, V3 checksums

bad move:
  edit V2__add_users_email.sql after prod already applied it
  flyway validate
    -> local V2 checksum != prod recorded V2 checksum
    -> validation fails

safe move:
  create V4__alter_users_email_length.sql
  flyway migrate
    -> apply only V4
    -> history remains replayable
```

이 흐름은 "migration 파일은 배포 후 수정하지 않는다"는 규칙을 단순한 팀 취향이 아니라 재현성의 문제로 만든다. 개발자가 V2를 고쳐서 local DB에서는 깨끗해 보여도, prod는 이미 옛 V2의 결과를 품고 있다. schema diff는 현재 구조 차이를 보여 주지만, Flyway history는 그 구조가 어떤 script 순서로 만들어졌는지 보여 준다. 운영에서 중요한 것은 둘 다다. 구조가 같아도 history가 다르면 다음 migration이나 rollback 판단이 달라질 수 있고, history가 같아도 수동 hotfix 때문에 구조가 달라질 수 있다.

DDL의 위험은 history뿐 아니라 실행 방식에도 있다. MySQL InnoDB 공식 문서는 online DDL operation별로 instant, in-place, table rebuild, concurrent DML 허용 여부, metadata-only 여부가 다르다고 표로 제시한다. 예를 들어 secondary index 추가는 in-place이며 테이블을 rebuild하지 않고 concurrent DML을 허용한다고 설명되지만, operation이 끝나기 전에는 해당 테이블에 접근 중인 transaction들이 끝나야 한다. 반대로 컬럼 타입 변경은 `ALGORITHM=COPY`만 지원되는 경우가 있고, `VARCHAR(255)`에서 `VARCHAR(256)`처럼 길이 byte 수가 바뀌는 변경은 in-place가 안 되어 table copy가 필요할 수 있다. `NOT NULL`로 바꾸는 작업은 in-place 예시가 있더라도 테이블을 다시 구성하고 기존 `NULL` 값이 있으면 실패한다.

여기서 "online DDL"이라는 말이 가장 위험한 오해를 만든다. online은 "운영 중 실행할 수 있다"에 가깝지, "비용이 없다"나 "lock이 없다"가 아니다. 어떤 operation은 metadata만 바꾸고 끝난다. 어떤 operation은 테이블 데이터를 새 구조로 다시 쓴다. 어떤 operation은 DML을 허용하지만 시작과 끝에서 metadata lock을 잡는다. 어떤 operation은 긴 transaction 때문에 완료가 대기된다. 어떤 operation은 replica에 적용되는 동안 지연을 만들 수 있다. 그러므로 migration 계획에는 SQL뿐 아니라 예상 algorithm, lock 수준, 데이터 검증, 실행 시간, 중단 조건, 관측 쿼리가 있어야 한다.

```text
ALTER TABLE의 운영 timeline 예시

T-7d  설계
      - 컬럼 추가 목적, nullable 여부, default, backfill 필요 여부 결정
      - 애플리케이션 코드가 구 스키마/신 스키마를 동시에 견딜 수 있는지 확인

T-3d  리허설
      - staging에 prod와 비슷한 row count로 실행
      - 실행 시간, lock wait, replica lag, disk 사용량 관측
      - EXPLAIN이 아니라 DDL progress와 performance_schema를 본다

T-1d  배포 준비
      - flyway validate
      - schema diff로 예상 차이 확인
      - backup/restore 또는 rollback-forward 절차 확인

T0    DDL 실행
      - ALTER TABLE ... ALGORITHM=INPLACE/INSTANT/COPY 여부 명시 또는 확인
      - lock wait, running transaction, replication lag 관측

T+1   애플리케이션 전환
      - 새 컬럼 dual read/write 또는 backfill 완료 후 읽기 경로 전환

T+N   정리
      - 옛 컬럼/옛 코드 제거는 별도 migration으로 실행
      - diff와 history가 의도된 상태인지 다시 확인
```

이 timeline은 expand-and-contract 패턴과 연결된다. 한국어로는 "넓혀 놓고 전환한 뒤 줄이는 배포"라고 이해하면 된다. 예를 들어 `users.email` 컬럼을 추가하고 필수값으로 만들고 싶다고 하자. 운영 DB에 이미 수백만 사용자가 있는데 한 번에 `ADD COLUMN email VARCHAR(255) NOT NULL UNIQUE`를 실행하면 기존 행의 값이 없어서 실패하거나, default 채우기와 unique 검증 때문에 큰 비용이 생길 수 있다. 안전한 흐름은 보통 먼저 nullable 컬럼을 추가하고, 새 코드가 새 컬럼을 쓰게 하고, 기존 데이터를 backfill하고, 중복과 null을 검증하고, 마지막에 `NOT NULL`과 `UNIQUE`를 붙이는 식이다. 이 방식은 단계가 늘어나지만 각 단계의 실패 반경을 줄이고 애플리케이션 rollback 가능성을 보존한다.

```sql
-- 1단계: 스키마를 먼저 넓힌다. 기존 코드가 몰라도 안전해야 한다.
ALTER TABLE users
  ADD COLUMN email VARCHAR(255) NULL;

-- 2단계: 새 코드가 email을 쓰기 시작한다. 기존 행은 아직 NULL일 수 있다.
-- 애플리케이션 배포 후 backfill을 별도로 실행한다.
UPDATE users
SET email = CONCAT('user-', id, '@placeholder.local')
WHERE email IS NULL;

-- 3단계: 제약 추가 전 위반 후보를 관측한다.
SELECT email, COUNT(*) AS cnt
FROM users
WHERE email IS NOT NULL
GROUP BY email
HAVING COUNT(*) > 1;

SELECT COUNT(*) AS null_email_count
FROM users
WHERE email IS NULL;

-- 4단계: 관측이 PASS일 때만 제약을 조인다.
ALTER TABLE users
  MODIFY COLUMN email VARCHAR(255) NOT NULL;

ALTER TABLE users
  ADD CONSTRAINT uq_users_email UNIQUE (email);
```

이 SQL은 학습용으로 단순화되어 있다. 실제로는 placeholder email이 도메인상 허용되는지, 고객에게 노출되지 않는지, unique index 생성 방식과 lock 비용이 얼마인지, 배치가 한 번에 너무 많은 undo/redo와 replication lag를 만들지 않는지 확인해야 한다. 하지만 흐름 자체는 중요하다. migration은 "원하는 최종 스키마"를 한 번에 선언하는 것이 아니라, 현재 데이터와 현재 코드가 버틸 수 있는 중간 호환 상태를 지나 최종 스키마로 가는 시간축이다. 코드 rollback이 가능한 상태와 DB rollback이 가능한 상태는 다르다. DB migration이 이미 prod에 적용된 뒤 코드만 이전 버전으로 되돌리면, 이전 코드가 새 컬럼을 무시할 수는 있어도 삭제된 컬럼을 다시 만들지는 못한다. 그래서 destructive change, 즉 컬럼 삭제나 타입 축소는 더 늦은 단계로 미루는 것이 일반적으로 안전하다.

schema diff는 이 흐름에서 두 번 쓰인다. 실행 전에는 "내가 만들 migration이 의도한 차이만 만드는가"를 확인한다. 실행 후에는 "실제 DB가 기대한 구조가 되었는가"를 확인한다. 이 저장소의 diff 예제는 `Missing Column(s): some_service.posts.only_at_prod`, `Unexpected Column(s): some_service.users.email`, `Unexpected Table(s): only_at_dev` 같은 출력을 보여 준다. 이 결과를 운영 절차에 적용하면 다음처럼 판정해야 한다.

| diff 항목 | 가능한 의미 | 바로 DDL로 바꾸면 생기는 위험 | 필요한 판정 |
|---|---|---|---|
| `users.email` unexpected | dev에서 준비 중인 새 기능 | prod에 곧바로 추가하면 앱 코드, backfill, unique 정책이 준비되지 않았을 수 있음 | 요구사항과 migration 단계 확인 |
| `posts.only_at_prod` missing | prod hotfix 또는 dev 누락 | dev 기준으로 prod 컬럼을 삭제하면 운영 데이터 손실 가능 | 소유 팀과 사용 쿼리 확인 |
| `only_at_dev` unexpected table | 개발 실험 테이블 | 운영에 불필요한 테이블을 만들 수 있음 | 실험 부산물인지 product change인지 분류 |
| column order changed | tool noise일 수도 있음 | 의미 없는 migration을 만들 수 있음 | DBMS에서 컬럼 순서가 실제 의미인지 확인 |

컬럼 순서 차이는 좋은 함정이다. diff 도구는 `posts.created_at`의 order가 바뀌었다고 보여 줄 수 있다. 하지만 대부분의 애플리케이션 SQL이 컬럼명을 명시한다면 컬럼 순서는 기능 의미가 거의 없다. 반대로 `SELECT *`에 의존하는 오래된 export나 batch가 있으면 순서도 문제가 될 수 있다. 결론은 "무시"가 아니라 "이 차이가 실제 consumer에게 의미가 있는가"를 확인하는 것이다. diff 도구는 판단을 대신하지 않는다. diff 도구는 질문 목록을 만들어 준다.

Flyway validate는 diff와 다른 관측을 제공한다. Redgate Flyway 공식 문서에 따르면 validate는 적용된 migration과 사용 가능한 migration을 비교하고, 이름, 타입, checksum 차이, 이미 적용되었지만 로컬에 없는 version, 로컬에는 있지만 아직 적용되지 않은 version 같은 조건에서 실패한다. 이 검증은 schema structure의 모든 의미 차이를 찾는 도구가 아니다. 대신 migration history의 재현성을 지킨다. 그래서 배포 전에는 `flyway validate`와 schema diff를 서로 보완적으로 봐야 한다. validate가 PASS여도 누군가 운영 DB에 수동 DDL을 실행했다면 schema diff가 차이를 보여 줄 수 있다. diff가 같아 보여도 적용된 migration 파일이 바뀌었다면 validate가 실패할 수 있다.

```bash
# 배포 전 최소 관측 루틴 예시

flyway validate
# PASS: 적용된 migration의 이름/타입/checksum/history가 로컬 migration과 일치
# FAIL: 이미 적용된 migration 파일 수정, 누락된 migration, 순서/타입 불일치 의심

liquibase diff \
  --diff-types=tables,columns,indexes,views \
  --referenceUrl=jdbc:mysql://prod/some_service \
  --url=jdbc:mysql://staging_after_migration/some_service
# PASS: 이번 release에서 기대한 차이만 보임
# FAIL: 실험 테이블, hotfix 누락, 예상 밖 index/column 차이 존재

SELECT *
FROM performance_schema.events_stages_current
WHERE EVENT_NAME LIKE 'stage/innodb/alter table%';
# PASS: DDL 진행 상황이 예상 범위 안에서 움직임
# FAIL: 진행이 멈추거나 lock wait/long transaction과 함께 증가하지 않음
```

관측 쿼리는 DBMS와 버전에 맞게 조정해야 한다. MySQL 문서는 InnoDB ALTER TABLE progress를 Performance Schema로 모니터링하는 경로를 따로 제공한다. PostgreSQL은 같은 MySQL online DDL 개념을 그대로 쓰지 않으므로 `pg_stat_activity`, `pg_locks`, `pg_stat_progress_create_index` 같은 다른 catalog와 view를 본다. 여기서 중요한 것은 특정 view 이름을 외우는 것이 아니라, DDL이 실행되는 동안 "누가 기다리는지, 어떤 lock이 잡혔는지, 진행률이 있는지, replica가 밀리는지, 실패하면 어디까지 적용되었는지"를 볼 수 있어야 한다는 점이다.

online DDL에서 자주 밟는 senior trap은 `ALGORITHM=INPLACE, LOCK=NONE`을 붙였으니 안전하다고 믿는 것이다. MySQL 문서는 operation별 지원 여부와 예외를 매우 구체적으로 나눈다. secondary index 추가는 concurrent DML을 허용하지만, 새 index는 statement가 끝날 때 커밋된 최신 상태를 반영해야 하므로 관련 transaction 완료를 기다린다. `VARCHAR` 길이를 늘리는 작업도 255 byte 경계를 넘으면 length byte 수가 바뀌어 table copy가 필요할 수 있다. `NOT NULL` 변경은 기존 NULL 값이 있으면 실패한다. 즉 같은 `ALTER TABLE`이라도 데이터 모양과 타입 경계에 따라 비용이 달라진다. 운영 계획은 문법이 아니라 해당 operation의 실제 알고리즘과 현재 데이터 분포를 기준으로 세워야 한다.

또 다른 trap은 rollback을 `flyway undo`나 반대 SQL 하나로 단순화하는 것이다. 많은 운영 migration은 완전한 되돌리기보다 roll-forward가 안전하다. 이미 `email` 컬럼을 추가하고 새 코드가 값을 쓰기 시작했다면 컬럼을 삭제하는 rollback은 새 데이터를 잃는다. 타입을 넓힌 뒤 값을 저장했다면 타입을 다시 줄이는 rollback은 잘릴 수 있다. foreign key를 추가한 뒤 애플리케이션이 고아 행을 만들지 않는다고 가정하기 시작했다면 제약을 제거하는 rollback은 데이터 품질 책임을 갑자기 코드로 되돌린다. 따라서 rollback 계획은 "DDL 반대 명령"이 아니라 "코드와 데이터가 어느 시점까지 어떤 스키마를 함께 견딜 수 있는가"로 작성해야 한다.

```text
컬럼 rename의 안전한 migration 예시

목표: users.username -> users.login_name

위험한 한 방:
  ALTER TABLE users RENAME COLUMN username TO login_name;
  deploy new app
  -> old app rollback 시 username을 찾지 못함

넓혀 놓고 전환:
  V10 add login_name nullable
    users(username, login_name)
  app v1.1 dual write
    INSERT/UPDATE username and login_name together
  V11 backfill login_name from username
    verify no mismatch
  app v1.2 read login_name first, fallback username
  app v1.3 read login_name only after lag window
  V12 enforce NOT NULL/UNIQUE on login_name
  V13 drop username after old app no longer exists

관측:
  SELECT COUNT(*) FROM users WHERE login_name IS NULL;
  SELECT COUNT(*) FROM users WHERE login_name <> username;
  flyway validate;
  schema diff expected: username disappears only at V13
```

이 예시는 단계가 많다. 하지만 각 단계는 실패를 줄이는 구체적 이유가 있다. `V10`은 기존 코드를 깨지 않는다. dual write는 두 컬럼의 동기화를 만든다. backfill은 기존 데이터를 새 모델로 옮긴다. fallback read는 부분 배포와 rollback을 견딘다. `NOT NULL`과 `UNIQUE`는 데이터가 준비된 뒤에만 붙인다. 옛 컬럼 삭제는 모든 old app이 사라진 뒤 실행한다. 이 흐름은 모든 변경에 늘 필요한 것은 아니지만, 이름 변경, 타입 변경, 필수 컬럼 추가, 큰 테이블 제약 추가처럼 실패 비용이 큰 변경에서는 기본 사고방식으로 삼을 만하다.

schema migration은 애플리케이션 배포와 독립적으로 보이지만 실제로는 강하게 결합된다. 애플리케이션은 특정 스키마를 기대한다. SQL mapper, ORM entity, JSON serializer, batch export, BI query, 운영 script 모두 컬럼 이름과 타입을 기대한다. DB는 특정 시점에 하나의 구조만 갖지만, rolling deployment에서는 old app과 new app이 동시에 떠 있을 수 있다. 따라서 안전한 migration은 "old app도 통과, new app도 통과"하는 호환 구간을 만든다. 이를 backward-compatible migration이라고 부를 수 있다. 쉬운 말로는 "이전 코드도 새 DB에서 죽지 않고, 새 코드도 이전 DB 또는 넓혀진 DB에서 죽지 않는 구간"이다.

호환 구간을 만들 때 특히 주의할 것은 default와 backfill이다. MySQL online DDL 문서에서 default 값 변경은 metadata-only일 수 있다고 설명되지만, default는 새 행에만 영향을 준다. 기존 행이 자동으로 도메인상 올바른 값이 되는 것은 아니다. `status DEFAULT 'ACTIVE'`를 추가해도 기존 `NULL` status가 모두 active라는 의미를 갖는지 확인해야 한다. `created_at DEFAULT CURRENT_TIMESTAMP`를 추가해도 과거 행의 생성 시각을 알 수 없다면 placeholder를 넣는 것이 감사상 거짓일 수 있다. migration은 빈 칸을 채우는 기술 작업이 아니라 과거 데이터에 어떤 의미를 부여할지 결정하는 작업이다.

```sql
-- default가 기존 행 의미를 자동으로 해결하지 못하는 예
ALTER TABLE subscriptions
  ADD COLUMN status VARCHAR(20) NULL DEFAULT 'ACTIVE';

-- 새 INSERT는 status가 생길 수 있지만, 기존 행의 실제 상태는 결제/해지 이력으로 판단해야 한다.
SELECT s.id,
       s.status,
       MAX(p.paid_at) AS last_paid_at,
       MAX(c.cancelled_at) AS cancelled_at
FROM subscriptions s
LEFT JOIN payments p ON p.subscription_id = s.id
LEFT JOIN cancellations c ON c.subscription_id = s.id
WHERE s.status IS NULL
GROUP BY s.id, s.status;

-- backfill은 도메인 규칙을 SQL로 드러내야 한다.
UPDATE subscriptions s
SET status = CASE
  WHEN EXISTS (SELECT 1 FROM cancellations c WHERE c.subscription_id = s.id) THEN 'CANCELLED'
  WHEN EXISTS (SELECT 1 FROM payments p WHERE p.subscription_id = s.id) THEN 'ACTIVE'
  ELSE 'PENDING'
END
WHERE status IS NULL;
```

이 worked example에서 핵심은 `ALTER TABLE`보다 backfill 규칙이다. 어떤 행을 `ACTIVE`, `CANCELLED`, `PENDING`으로 볼지 도메인 근거가 있어야 한다. 잘못된 backfill은 스키마 migration이 아니라 데이터 오염이다. 운영에서는 backfill을 작은 batch로 나누고, 각 batch의 affected row count, 실행 시간, lock wait, replication lag를 기록한다. affected row count가 예상보다 크거나 작으면 중단해야 한다. "SQL이 성공했다"는 PASS가 아니다. 예상한 데이터가 예상한 규칙으로 움직였는지가 PASS다.

Flyway를 쓸 때 migration 파일 이름과 분할도 운영 품질에 영향을 준다. `V202605190101__add_users_email.sql`처럼 버전과 설명이 보이면 history에서 의도를 추적하기 쉽다. 하나의 파일에 무관한 변경을 너무 많이 넣으면 실패했을 때 어디서 깨졌는지 찾기 어렵고, 일부만 roll-forward하기도 어렵다. 반대로 너무 잘게 쪼개서 항상 같이 배포되어야 하는 제약과 backfill을 분리하면 중간 상태가 오래 남을 수 있다. 적절한 단위는 "한 번 실패했을 때 원인을 이해하고, 다시 실행하거나 다음 migration으로 보정할 수 있는 의미 단위"다. Flyway는 pending versioned migration을 순서대로 적용하므로, 파일 순서는 배포 순서이자 복구 사고의 순서다.

migration 파일 안에서도 idempotency, 즉 여러 번 실행해도 같은 결과가 되는 성질을 오해하면 안 된다. Flyway versioned migration은 기본적으로 한 번 적용된다. 그래서 파일 안에 무조건 `IF NOT EXISTS`를 넣어 조용히 통과시키는 것이 항상 좋은 것은 아니다. `CREATE TABLE IF NOT EXISTS`가 통과하면 실제 테이블 정의가 기대와 달라도 migration은 성공처럼 보일 수 있다. 반대로 운영 재시도 가능성을 위해 일부 DML backfill은 범위를 제한하고 이미 처리된 행을 건너뛰도록 작성하는 편이 안전하다. DDL은 history로 한 번 실행을 보장하고, 긴 DML은 재시도 가능한 작은 batch로 나누는 식으로 성질을 구분해야 한다.

```text
versioned migration과 재시도 가능 batch의 역할 분리

V20__add_subscription_status.sql
  - ADD COLUMN status NULL
  - schema history에 한 번 적용

backfill job
  - WHERE status IS NULL AND id BETWEEN ? AND ?
  - 5,000 rows 단위로 commit
  - affected row count 기록
  - 실패 시 마지막 성공 범위 이후 재시작

V21__enforce_subscription_status.sql
  - NULL 위반 0건 확인 후 NOT NULL 추가
  - CHECK 또는 reference table 제약 추가

검증
  - flyway validate -> migration history 무결성
  - SELECT COUNT(*) WHERE status IS NULL -> 데이터 준비 상태
  - schema diff -> 예상 구조만 남았는지
```

이 분리는 undo/redo 로그와 replication에도 영향을 준다. 큰 `UPDATE` 하나가 수백만 행을 바꾸면 undo/redo가 커지고, replica SQL thread가 오래 밀리고, lock wait이 늘고, 장애 시 복구 시간이 길어질 수 있다. 작은 batch는 총 시간은 늘 수 있지만 관측과 중단이 쉬워진다. "빠른 한 방"은 local에서는 좋아 보이지만 운영에서는 가장 느린 장애 복구 경로가 될 수 있다. MySQL의 online DDL도 operation 자체는 DML을 허용할 수 있지만, 그 옆에서 대규모 backfill을 동시에 돌리면 버퍼 풀, redo, replica lag, lock 경합이 같이 올라간다.

마이그레이션의 history/context도 중요하다. 예전에는 운영 DB에 직접 접속해 DDL을 실행하거나, 배포 스크립트 한쪽에 SQL을 끼워 넣는 방식이 흔했다. 이런 방식은 작은 팀에서는 빠르지만, 시간이 지나면 "이 DB가 왜 이 구조가 되었는가"를 복원하기 어렵다. Flyway나 Liquibase 같은 도구가 등장한 배경은 DB 구조 변경을 애플리케이션 코드처럼 version control과 배포 파이프라인에 올리기 위해서다. 다만 도구가 있다고 해서 설계 판단이 자동화되지는 않는다. 도구는 순서, checksum, 비교, 적용 기록을 제공한다. 어떤 변경이 안전한지, 어떤 순서로 쪼갤지, 어떤 데이터 의미로 backfill할지는 여전히 개발자와 DBA가 근거를 갖고 결정해야 한다.

운영 절차로 닫으려면 각 migration에 작은 runbook을 붙이는 편이 좋다. runbook은 긴 문서가 아니라 실행 전 확인, 실행 중 관측, 실패 시 중단 조건, 실행 후 검증을 담은 재현 가능한 절차다. 예를 들어 필수 컬럼 추가 migration의 runbook은 다음처럼 쓸 수 있다.

```text
Runbook: users.email 필수화

사전 조건
  - app v1.1 이상이 email nullable을 견딤
  - 신규 가입 write path가 email을 채움
  - backfill SQL이 staging에서 예상 row count로 검증됨

실행 전 확인
  - flyway validate PASS
  - SELECT COUNT(*) FROM users WHERE email IS NULL;  -- 예상 값 확인
  - SELECT email, COUNT(*) ... HAVING COUNT(*) > 1; -- 0행이어야 제약 가능
  - replica lag baseline 기록

실행 중 관측
  - lock wait 증가 여부
  - DDL progress 또는 processlist
  - batch affected rows
  - replica lag

중단 조건
  - lock wait가 서비스 SLO를 넘김
  - replica lag가 읽기 서비스 허용치를 넘김
  - duplicate email 발견
  - affected row count가 예상 범위를 크게 벗어남

실행 후 확인
  - flyway validate PASS
  - schema diff expected only
  - null email count = 0
  - duplicate email count = 0
  - old app rollback 가능 여부 재확인
```

이런 runbook은 형식 문서가 아니라 장애 시간을 줄이는 도구다. migration은 한 번 성공하면 잊히지만, 실패할 때는 모두가 동시에 "지금 멈춰도 되나", "어디까지 적용됐나", "코드 rollback부터 할까 DB부터 볼까"를 묻는다. 사전에 PASS/FAIL 신호와 중단 조건이 없으면 운영자는 감으로 결정하게 된다. 감으로 결정한 migration은 다음 회고에서 재현하기 어렵다.

schema diff와 migration history를 함께 볼 때 한 가지 더 조심할 점은 환경별 drift다. 개발 DB는 여러 실험이 들어가고, staging은 release 후보가 들어가고, prod는 hotfix와 운영 데이터가 있다. `database/examples/db_diff/whendiff.result`처럼 `only_at_dev`가 보이는 경우, 이것은 release 대상이 아니라 실험 부산물일 수 있다. 반대로 prod에만 있는 `only_at_prod`는 운영에서 급히 추가한 컬럼일 수 있다. drift를 무조건 없애는 것이 답은 아니다. 먼저 drift의 출처를 찾아야 한다. 의도된 hotfix라면 migration repository에 정식 script로 흡수해야 한다. 실험 부산물이라면 dev에서 제거하거나 별도 sandbox로 옮겨야 한다. release 후보라면 Flyway migration과 애플리케이션 코드가 같은 PR 또는 같은 배포 묶음에서 추적되어야 한다.

마지막으로, online DDL과 migration 도구가 있어도 "큰 테이블 변경"은 여전히 별도 설계 대상이다. 수억 행 테이블에 `NOT NULL` 제약을 추가하거나 타입을 바꾸거나 primary key를 재구성하는 작업은 단순 DDL보다 데이터 복사 프로젝트에 가깝다. 이때는 shadow table, dual write, trigger 기반 동기화, gh-ost/pt-online-schema-change 같은 online schema change 도구, 또는 애플리케이션 레벨 분할 migration을 검토할 수 있다. MySQL online DDL과 Flyway 문서를 함께 읽어야 하는 이유도 여기에 있다. 실무 판단에서는 DBMS 내장 online DDL이 충분한지, 외부 online schema change 도구가 필요한지, 아니면 도메인 모델을 나눠 단계적으로 옮겨야 하는지 비교해야 한다. 내장 DDL이 충분하다는 결론도 staging 리허설과 관측 지표가 있어야 한다.

실패 상태를 나눠 보는 습관도 필요하다. migration 실패는 하나가 아니다. DDL이 시작 전에 syntax나 권한 문제로 실패할 수 있고, 시작 후 기존 데이터 위반 때문에 실패할 수 있고, table copy 중 disk 부족으로 실패할 수 있고, lock wait timeout으로 실패할 수 있고, Flyway history 기록 전후의 어느 지점에서 실패할 수도 있다. 각각 복구 판단이 다르다. 시작 전에 실패했다면 script를 고치고 다시 실행하면 될 수 있다. 기존 데이터 위반으로 실패했다면 데이터를 정리하지 않고 재시도해도 같은 실패가 반복된다. table copy나 긴 DML 중 실패했다면 임시 객체, undo/redo pressure, replica 상태를 확인해야 한다. Flyway가 failed migration을 history에 기록한 상태라면 `repair` 같은 도구 명령을 검토하기 전에 실제 DB 구조가 어디까지 바뀌었는지 먼저 봐야 한다.

```text
migration 실패 분류표

실패 지점                     먼저 볼 것                         재시도 조건
----------------------------  --------------------------------  ------------------------------
syntax/권한 실패              SQL 문법, 계정 권한                DB 구조 변화 없음 확인 후 수정 재시도
기존 데이터 제약 위반          위반 행 SELECT, 정리 정책          정리 migration 또는 backfill 후 재시도
lock wait timeout             long transaction, metadata lock    차단 transaction 종료/창구 변경 후 재시도
disk/임시공간 부족             tablespace, temp, redo/undo        용량 확보와 중간 객체 확인 후 재시도
replica lag 과다              replica status, relay 적용 속도     읽기 트래픽 영향 해소 후 재개
Flyway checksum/history 실패   flyway_schema_history, 파일 checksum 실제 DB 구조와 history 일치성 확인 후 repair/roll-forward
```

이 표에서 `repair`는 마법의 복구 버튼이 아니다. history 기록이 잘못되었을 때 metadata를 고치는 도구일 수 있지만, 실제 DB 구조가 기대와 다르면 history만 고쳐서는 데이터베이스가 안전해지지 않는다. 반대로 DB 구조는 의도대로 바뀌었는데 migration tool의 기록만 실패했다면 기록을 정리해야 다음 배포가 막히지 않는다. 그래서 실패 후 첫 질문은 "도구가 무엇을 기록했나"와 "DB가 실제로 어떤 구조가 되었나"를 분리하는 것이다.

정상 완료 후에도 검증은 남는다. migration 직후에는 새 코드가 사용하는 query path, old code rollback path, batch/export path, BI/report path를 최소한 smoke test로 확인해야 한다. schema diff가 예상대로여도 권한, view, trigger, generated column, index 선택, 통계 갱신 문제로 애플리케이션이 느려질 수 있다. 특히 큰 index를 새로 만들거나 컬럼 분포가 바뀐 뒤에는 optimizer 통계가 충분히 갱신되었는지, 주요 query의 plan이 바뀌었는지 확인해야 한다. migration은 DDL 성공 메시지에서 끝나는 것이 아니라, 데이터베이스가 새 구조로 실제 workload를 감당하는지 관측할 때 닫힌다.

정리하면 DU24의 핵심은 다음과 같다. schema diff는 현재 구조 차이를 보여 주는 관측 도구이고, Flyway는 변경 script의 순서와 history를 지키는 실행 도구이며, online DDL은 operation별로 다른 lock과 데이터 이동 비용을 갖는 DBMS 기능이다. 셋은 서로 대체제가 아니다. 안전한 migration은 diff로 drift를 보고, Flyway validate로 history를 보고, DBMS 공식 문서로 DDL 알고리즘을 확인하고, staging 리허설과 운영 관측으로 실제 비용을 확인한 뒤, 애플리케이션 호환 구간을 지나 최종 제약을 조인다. `ALTER TABLE`이 즉시 끝난다는 오해를 버리고 migration timeline을 그릴 수 있어야, 스키마 변경을 기능 배포가 아니라 운영 가능한 데이터 이동으로 다룰 수 있다.

공식 근거로는 MySQL 8.4 InnoDB `Online DDL Operations`(`https://dev.mysql.com/doc/refman/8.4/en/innodb-online-ddl-operations.html`), Redgate Flyway `Migrations`(`https://documentation.red-gate.com/fd/migrations-271585107.html`), `Versioned migrations`(`https://documentation.red-gate.com/flyway/flyway-concepts/migrations/versioned-migrations`), `Validate`(`https://documentation.red-gate.com/flyway/reference/commands/validate`) 문서를 사용했다. 로컬 seed로는 `database/db_diff.md`, `database/examples/db_diff/README.md`, `diff_db.sh`, `whendiff/dev/init.sql`, `whendiff/prod/init.sql`, `whendiff.result`, `whensame.result`, 그리고 `database/migration/flyway/flyway.md`를 사용했다. 이 section의 예시는 운영 절차 감각을 세우기 위한 학습 모델이며, 실제 실행 전에는 대상 DBMS 버전, 테이블 크기, replication 구조, lock timeout, backup/restore 조건, 애플리케이션 rolling deployment 방식을 반드시 현재 환경에서 재검증해야 한다.
