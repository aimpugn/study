# 선불 잔액 사용처리(차감)의 정확성: 동시성, 트랜잭션, 격리 수준

선불(prepaid) 서비스는 "먼저 충전하고 나중에 쓴다"는 한 문장으로 요약됩니다. 그런데 실제로 어렵고 사고가 자주 나는 쪽은 충전이 아니라 **사용처리(차감)** 입니다. 이 문서는 "충전은 이미 끝났다"고 가정하고, 처리계(처리를 담당하는 서버)가 그 잔액을 사용처리할 때 무엇을 고려해야 하는지를 다룹니다. 핵심 질문은 하나입니다. **두 요청이 같은 잔액을 동시에 쓰려 할 때, 어떻게 하면 "있는 돈보다 더 쓰는 일"을 구조적으로 막을 수 있는가.**

이 문서는 같은 질문을 세 번, 점점 더 구체적으로 답합니다. 먼저 언어와 프레임워크에 무관한 원리로, 그다음 일반적인 프레임워크와 트랜잭션 관리 관점으로, 마지막으로 Spring Boot + MyBatis + MySQL을 쓰는 strict-io 환경(`kcf-projects`)의 실제 코드로 내려갑니다. 트랜잭션 격리 수준, `READ COMMITTED`와 `REPEATABLE READ`의 차이, 팬텀 읽기(phantom read), 전파(propagation)는 이 질문을 푸는 과정에서 필요할 때마다 깊게 끌어옵니다.

## 목차

- [0. 이 문서가 답하는 질문과 범위](#0-이-문서가-답하는-질문과-범위)
- [1. 첫 번째 벽돌: 잔액 100원에 80원을 동시에 두 번 쓰면](#1-첫-번째-벽돌-잔액-100원에-80원을-동시에-두-번-쓰면)
- [2. 사전 검증만으로는 왜 부족한가 (check-then-act 함정)](#2-사전-검증만으로는-왜-부족한가-check-then-act-함정)
- [3. 언어·프레임워크 무관 원리](#3-언어프레임워크-무관-원리)
    - [3.1 불변식을 읽는 곳이 아니라 쓰는 곳에서 지킨다](#31-불변식을-읽는-곳이-아니라-쓰는-곳에서-지킨다)
    - [3.2 세 가지 정확성 전략: 원자적 조건부 갱신·비관적 락·낙관적 락](#32-세-가지-정확성-전략-원자적-조건부-갱신비관적-락낙관적-락)
    - [3.3 금액 관리: 원장 우선과 잔액 투영](#33-금액-관리-원장-우선과-잔액-투영)
    - [3.4 같은 요청과 다른 요청: 멱등성과 동시성은 다른 문제다](#34-같은-요청과-다른-요청-멱등성과-동시성은-다른-문제다)
- [4. 트랜잭션과 격리 수준 심층](#4-트랜잭션과-격리-수준-심층)
    - [4.1 차감 작업을 위협하는 이상 현상은 둘이다](#41-차감-작업을-위협하는-이상-현상은-둘이다)
    - [4.2 READ COMMITTED와 REPEATABLE READ: 스냅샷의 함정](#42-read-committed와-repeatable-read-스냅샷의-함정)
    - [4.3 팬텀 읽기와 write skew: 한도 검증이 무너지는 자리](#43-팬텀-읽기와-write-skew-한도-검증이-무너지는-자리)
    - [4.4 트랜잭션 범위와 전파(propagation)](#44-트랜잭션-범위와-전파propagation)
- [5. 일반 프레임워크에서의 처리](#5-일반-프레임워크에서의-처리)
- [6. kcf-projects strict-io 구체화](#6-kcf-projects-strict-io-구체화)
    - [6.1 strict-io 계층과 트랜잭션 경계](#61-strict-io-계층과-트랜잭션-경계)
    - [6.2 현재 펌뱅킹 입금 흐름이 실제로 하는 일](#62-현재-펌뱅킹-입금-흐름이-실제로-하는-일)
    - [6.3 차감 가드를 strict-io의 어디에 두는가](#63-차감-가드를-strict-io의-어디에-두는가)
    - [6.4 MySQL·InnoDB·HikariCP에서 확인할 점](#64-mysqlinnodbhikaricp에서-확인할-점)
- [7. 실패 모드와 직접 검증](#7-실패-모드와-직접-검증)
- [8. 한 장 정리](#8-한-장-정리)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 0. 이 문서가 답하는 질문과 범위

선불 서비스의 생애주기는 크게 충전, 사용처리(차감), 환불·취소, 정산으로 나뉩니다. 이 문서의 범위는 그중 **사용처리** 한 단계입니다. 충전으로 잔액 100,000원이 이미 안전하게 적립됐다고 보고, 결제·송금·포인트 차감처럼 "그 잔액에서 일부를 빼는" 작업을 처리계가 어떻게 정확하게 수행하는지를 봅니다.

읽고 나면 다음을 스스로 설명할 수 있어야 합니다.

- 잔액 차감에서 사고를 만드는 가장 흔한 동시성 현상이 **갱신 손실(lost update)** 이고, 그 결과가 **초과 차감/이중 사용(oversell)** 이라는 점
- 사전 검증(pre-validation)이 왜 필요하면서도 **정확성의 보증 장치는 될 수 없는지**
- 같은 목표를 푸는 세 가지 수단(원자적 조건부 갱신, 비관적 락, 낙관적 락)이 각각 언제 더 나은지
- `READ COMMITTED`와 `REPEATABLE READ`에서 차감 로직이 어떻게 다르게 동작하고, 어디서 스냅샷(snapshot)이 함정이 되는지
- 단건 잔액의 갱신 손실과, 일별 한도 같은 집계 검증의 **write skew(쓰기 왜곡)** 가 왜 다른 방어를 요구하는지

이 문서는 격리 수준의 일반 이론, MVCC 내부 구조, 멱등성·아웃박스(outbox)의 전체 카탈로그를 처음부터 다시 쓰지 않습니다. 그 부분은 이미 저장소에 깊게 정리돼 있으므로 필요한 자리에서 연결합니다. 대신 이 문서는 **선불 차감이라는 한 작업에 그 이론들을 적용**하는 데 집중합니다. 일반 이론은 [트랜잭션 고립 수준과 이상 현상](../../database/deep-dive/transactions/13-isolation-anomalies.md), [MVCC](../../database/mvcc.md), [애플리케이션 경계·멱등성·금액·아웃박스](../../interviews/database-deep-dive/12-application-boundaries-idempotency-money-outbox.md)에서 확인할 수 있습니다.

## 1. 첫 번째 벽돌: 잔액 100원에 80원을 동시에 두 번 쓰면

추상 설명보다 가장 작은 사고 장면을 먼저 봅니다. 계좌 한 줄과 요청 두 개만 있으면 이 문서의 모든 위험이 드러납니다.

```sql
-- 가장 단순한 선불 잔액 한 줄
CREATE TABLE prepaid_account (
    account_id  BIGINT      PRIMARY KEY,
    balance     BIGINT      NOT NULL,   -- 최소 단위(원) 정수로 저장
    CHECK (balance >= 0)
);

INSERT INTO prepaid_account(account_id, balance) VALUES (1, 100);
```

이제 처리계가 흔히 짜는 차감 코드를 의사코드로 보면, 대부분 "읽고 → 검사하고 → 쓴다"의 세 박자입니다.

```text
deduct(account_id, useAmount):
    balance = SELECT balance FROM prepaid_account WHERE account_id = ?   # 읽기
    if balance < useAmount:                                             # 검사(사전 검증)
        reject("잔액 부족")
    newBalance = balance - useAmount
    UPDATE prepaid_account SET balance = newBalance WHERE account_id = ? # 쓰기
```

한 요청만 보면 완벽합니다. 문제는 같은 계좌에 80원짜리 사용처리 두 건이 거의 동시에 들어올 때입니다. 두 트랜잭션 `A`, `B`가 시간 위에서 어떻게 엇갈리는지 그대로 따라가 봅니다.

```text
잔액 시작값: balance = 100,  두 요청 모두 useAmount = 80

time | 트랜잭션 A                          | 트랜잭션 B                          | DB의 balance
-----+-------------------------------------+-------------------------------------+-------------
t1   | SELECT balance -> 100               |                                     | 100
t2   |                                     | SELECT balance -> 100               | 100
t3   | check: 100 >= 80  -> 통과           |                                     | 100
t4   |                                     | check: 100 >= 80  -> 통과           | 100
t5   | UPDATE balance = 100 - 80 = 20      |                                     | 20
t6   |                                     | UPDATE balance = 100 - 80 = 20      | 20   (A의 차감이 사라짐)
-----+-------------------------------------+-------------------------------------+-------------
결과: 160원어치를 썼는데 잔액은 20원.  실제로 빠진 돈은 80원뿐.
```

두 요청이 모두 성공했고, 화면에는 "결제 완료"가 두 번 떴습니다. 그런데 잔액은 100에서 20으로 80만 줄었습니다. 80 + 80 = 160원어치 서비스를 제공하고 잔액은 80원만 깎은 것입니다. `t6`에서 `B`의 `UPDATE`가 `A`가 쓴 20을 덮어쓰면서, **`A`의 차감이 통째로 사라졌습니다.**

이 현상의 이름이 **갱신 손실(lost update)** 입니다. 같은 행을 두 트랜잭션이 "읽은 값 기준으로" 각자 계산해서 쓰면, 늦게 쓴 쪽이 먼저 쓴 쪽의 변경을 덮어씁니다. 선불 도메인에서 갱신 손실의 결과는 **초과 차감/이중 사용(oversell)**, 즉 있는 잔액보다 더 많은 서비스가 나가는 사고입니다. 결제·포인트·쿠폰·재고처럼 "한정된 수량을 깎는" 모든 도메인이 같은 모양의 사고를 공유합니다.

이 한 장면이 첫 번째 벽돌입니다. 앞으로 나오는 모든 해법은 결국 "이 `t1`~`t6` 인터리빙에서 `B`가 100을 다시 읽지 못하게 하거나, `B`의 쓰기가 거부되게 만드는" 방법입니다.

## 2. 사전 검증만으로는 왜 부족한가 (check-then-act 함정)

위 코드에는 `if balance < useAmount: reject`라는 사전 검증이 분명히 있었습니다. 그런데도 사고가 났습니다. 사전 검증을 더 꼼꼼히, 더 앞단에, 더 여러 번 한다고 이 사고가 막히지는 않습니다. 이유를 정확히 짚어야 합니다.

문제는 **검사(check)와 행위(act) 사이에 시간 간격이 있고, 그 사이에 세상이 바뀐다**는 데 있습니다. `A`가 `t3`에서 "100 >= 80"을 확인한 순간의 사실은, `t5`에서 `UPDATE`를 실행하는 순간에는 더 이상 보장되지 않습니다. `B`가 그 사이에 끼어들 수 있기 때문입니다. 이렇게 "읽어서 확인한 조건"을 "나중에 쓸 때까지 유효할 것"이라고 믿는 패턴을 **check-then-act 경쟁 조건(race condition)** 이라고 부릅니다.

그래서 사전 검증의 역할을 정확히 자리매김해야 합니다.

- 사전 검증이 **하는 일**: 명백히 틀린 요청을 싸고 빠르게 거른다. 잔액이 한참 모자라거나, 한도를 명백히 넘거나, 계좌 상태가 정지면 비싼 처리에 들어가기 전에 바로 거절해 사용자 경험과 자원을 아낀다.
- 사전 검증이 **하지 못하는 일**: 동시에 들어온 다른 요청이 같은 잔액을 건드리는 것을 막지 못한다. 검사 시점의 잔액이 쓰기 시점까지 그대로라는 보장을 주지 못한다.

정리하면 이렇습니다.

> **사전 검증은 "빠른 거절(fast reject)"을 위한 최적화이고, 정확성의 보증 장치가 아닙니다. 정확성은 반드시 값을 실제로 바꾸는 그 지점에서 한 번 더 지켜져야 합니다.**

이 한 문장이 이 문서의 중심축입니다. 다음 장의 세 가지 전략은 전부 "보증을 쓰는 지점으로 옮기는" 방법입니다. 사전 검증은 그대로 두되, 거기에 기대지 않습니다.

## 3. 언어·프레임워크 무관 원리

### 3.1 불변식을 읽는 곳이 아니라 쓰는 곳에서 지킨다

선불 차감이 지켜야 할 업무 불변식(invariant, 항상 참이어야 하는 조건)은 단순합니다. **`차감 후 잔액 >= 0`**. 1장의 사고는 이 불변식을 "읽은 값으로 미리 계산해서" 판단했기 때문에 깨졌습니다. 같은 불변식을 **쓰는 순간에 강제**하면 사고가 사라집니다.

쓰는 곳에서 불변식을 지키는 방법은 두 갈래입니다.

1. 한 문장(single statement) 안에서 검사와 갱신을 동시에 한다.

    DB가 행을 갱신할 때는 그 행에 배타적 잠금(exclusive lock)을 잡습니다. 검사 조건을 `UPDATE`의 `WHERE`에 붙이면, DB가 잠금을 잡은 상태에서 "지금 이 순간의 확정된 값"으로 조건을 평가합니다. 검사와 행위 사이의 틈이 사라집니다.

2. 검사부터 갱신까지를 하나의 잠금 구간으로 묶는다.

    먼저 행을 잠그고(`SELECT ... FOR UPDATE`), 그 잠금을 쥔 채로 읽고 계산하고 쓰고 커밋합니다. 다른 트랜잭션은 잠금이 풀릴 때까지 그 행을 건드리지 못하므로, 그사이 세상이 바뀌지 않습니다.

어느 쪽이든 핵심은 **"검사한 사실이 쓰기까지 유효하다"를 DB의 잠금으로 보장**한다는 점입니다. 애플리케이션 메모리 안에서 비교한 결과는 이 보장을 주지 못합니다.

### 3.2 세 가지 정확성 전략: 원자적 조건부 갱신·비관적 락·낙관적 락

같은 목표(초과 차감 금지)를 세 가지 수단으로 풀 수 있습니다. 셋은 우열이 아니라 **경쟁이 잦은 정도와 작업의 복잡도**에 따라 갈립니다.

1. 원자적 조건부 갱신 (atomic conditional update)

    검사와 차감을 한 `UPDATE` 문에 합칩니다.

    ```sql
    UPDATE prepaid_account
       SET balance = balance - :useAmount
     WHERE account_id = :id
       AND balance  >= :useAmount;      -- 잔액이 충분할 때만 깎인다
    ```

    이 한 문장이 영향 준 행 수(affected rows)를 봅니다. **1이면 성공, 0이면 잔액 부족 또는 경쟁에서 밀린 것**이므로 거절합니다. 1장의 사고를 이 문장으로 다시 돌려 보면, `B`의 `UPDATE`는 `A`가 이미 잔액을 20으로 만든 뒤 실행되므로 `WHERE balance >= 80`이 거짓이 되어 0행이 갱신되고 거절됩니다.

    ```text
    time | 트랜잭션 A                              | 트랜잭션 B                              | balance
    -----+-----------------------------------------+-----------------------------------------+--------
    t1   | UPDATE ... SET balance=balance-80       |                                         | 100->20
         |  WHERE id=1 AND balance>=80  (1행)      |                                         |
    t2   | COMMIT                                  |                                         | 20
    t3   |                                         | UPDATE ... SET balance=balance-80       | 20
         |                                         |  WHERE id=1 AND balance>=80  (0행!)     |
    t4   |                                         | affected=0 -> "잔액 부족"으로 거절       | 20
    ```

    애플리케이션이 잔액을 읽어서 빼지 않고, "빼라"는 의도와 "충분할 때만"이라는 조건을 DB에 통째로 넘깁니다. **단건 잔액 차감의 기본값으로 가장 권장**됩니다. 별도 `SELECT`가 없어 왕복이 한 번이고, 잠금 구간이 짧습니다.

2. 비관적 락 (pessimistic lock)

    차감 결정이 한 문장으로 끝나지 않을 때 씁니다. 잔액을 깎으면서 원장(ledger)도 남기고, 여러 행이나 외부 조건을 함께 봐야 한다면, 먼저 잠그고 그 안에서 계산합니다.

    ```sql
    -- 트랜잭션 안에서
    SELECT balance FROM prepaid_account
     WHERE account_id = :id
     FOR UPDATE;                 -- 이 행을 배타적으로 잠근다

    -- 이제 읽은 balance는 커밋 전까지 다른 트랜잭션이 못 바꾼다
    -- 잔액 검사, 원장 insert, 잔액 update를 안전하게 수행
    ```

    `FOR UPDATE`가 행 잠금을 잡는 동안 다른 차감 트랜잭션은 같은 행의 `FOR UPDATE`에서 대기합니다. 경쟁이 직렬화되므로 가장 직관적이지만, 잠금 대기와 교착(deadlock)을 설계해야 하고, 잠금을 쥔 채 외부 호출을 하면 대기가 길어집니다.

3. 낙관적 락 (optimistic lock)

    충돌이 드물다고 보고, 잠그지 않고 진행하되 쓸 때 들킵니다. 버전 컬럼(version)을 두고 갱신 조건에 넣습니다.

    ```sql
    UPDATE prepaid_account
       SET balance = :newBalance,
           version = version + 1
     WHERE account_id = :id
       AND version   = :readVersion;   -- 내가 읽었을 때의 버전과 같을 때만
    ```

    0행이 갱신되면 그사이 누가 먼저 바꾼 것이므로, 다시 읽어서 재시도합니다. 충돌이 적으면 잠금 대기가 없어 처리량이 좋지만, 충돌이 잦으면 재시도가 폭증합니다.

세 전략을 한눈에 비교하면 이렇습니다.

| 전략 | 막는 방식 | 잘 맞는 상황 | 새로 떠안는 비용 |
| --- | --- | --- | --- |
| 원자적 조건부 갱신 | 검사+갱신을 한 문장으로, 행 잠금 안에서 조건 재평가 | 단건 잔액 차감 | 복잡한 다단계 로직은 한 문장에 못 담음 |
| 비관적 락(`FOR UPDATE`) | 먼저 잠그고 그 안에서 읽기·계산·쓰기 | 잔액+원장+여러 행을 함께 결정 | 잠금 대기, 교착, 외부 호출 시 점유 시간 |
| 낙관적 락(version) | 잠그지 않고 쓸 때 버전으로 충돌 감지 | 충돌이 드문 경우 | 충돌 잦으면 재시도 폭증 |

세 전략 모두 1장의 사고를 막습니다. 차이는 "막는 비용을 어디서 치르는가"입니다. 원자적 조건부 갱신은 비용이 가장 작아서, 단건 잔액이라면 이것을 기본값으로 두고 나머지는 필요할 때만 선택하는 편이 좋습니다.

### 3.3 금액 관리: 원장 우선과 잔액 투영

차감의 동시성을 막는 것과 별개로, 금액 자체를 어떻게 들고 있느냐가 사고의 또 다른 축입니다. 핵심 원칙은 **잔액 숫자 하나를 진실로 삼지 말고, "왜 이 잔액이 되었는가"를 먼저 기록**하는 것입니다.

- 원장 우선(ledger-first): 차감이 일어날 때 잔액만 90,000 → 80,000으로 바꾸지 말고, "주문 O100 때문에 10,000 차변" 같은 **원장 항목을 먼저 남깁니다.** 잔액은 원장의 합을 빠르게 보기 위한 **투영(projection)** 으로 둡니다. 그래야 나중에 불일치가 나도 원장으로 재계산하고 원인을 추적할 수 있습니다.
- 최소 단위 정수: 원화는 1원, 달러는 센트처럼 가장 작은 결제 단위를 정수로 저장하면 반올림 오차가 줄고 비교가 단순해집니다. 세율·수수료처럼 소수가 필요하면 십진 고정 정밀 타입(`DECIMAL`/`NUMERIC`, Java `BigDecimal`)을 쓰되 **반올림 자리(scale)와 반올림 방식(rounding mode)을 한 흐름에서 하나로 고정**합니다.
- 음수 금지 불변식: `balance >= 0`을 DB `CHECK` 제약으로도 둡니다. 애플리케이션 버그로 음수 차감이 들어와도 원장이 깨지지 않게 하는 마지막 방어선입니다.

금액 모델의 세부(이진 부동소수점 회피, 나눗셈 후 1원 배분 규칙, 원장과 잔액의 분리, 결제 상태 기계)는 [애플리케이션 경계·멱등성·금액·아웃박스](../../interviews/database-deep-dive/12-application-boundaries-idempotency-money-outbox.md)에 깊게 정리돼 있습니다. 이 문서에서는 차감의 정확성과 직접 맞닿는 "원장 우선 + 쓰는 곳에서 불변식 강제"만 들고 갑니다.

### 3.4 같은 요청과 다른 요청: 멱등성과 동시성은 다른 문제다

여기서 자주 섞이는 두 문제를 분리해야 합니다. 둘 다 "중복 차감"처럼 보이지만 원인과 해법이 다릅니다.

1. 같은 요청이 두 번 들어옴 (멱등성 문제)

    사용자가 결제 버튼을 두 번 누르거나, 게이트웨이 타임아웃 뒤 같은 요청이 재시도됩니다. **같은 의도**가 반복된 것이므로, 한 번만 처리하고 두 번째는 첫 결과를 그대로 돌려줘야 합니다. 해법은 **멱등성 키(idempotency key)** 에 유니크 제약을 걸어 두 번째 삽입을 막고, 그것을 "이미 처리됨"으로 해석하는 것입니다.

2. 서로 다른 요청이 같은 잔액을 경쟁 (동시성 문제)

    1장의 사고처럼, **다른 두 결제**가 우연히 같은 잔액을 동시에 깎으려는 것입니다. 둘 다 정당한 요청이고, 합쳐서 잔액을 넘기는 게 문제입니다. 해법은 3.2의 잠금/조건부 갱신입니다.

```text
같은 요청 두 번                         서로 다른 요청 두 개
key=pay:order-100 (재시도)              key=pay:order-100,  key=pay:order-200
  -> 한 번만 처리, 결과 재생             -> 둘 다 처리하되, 합이 잔액을 넘으면 한쪽 거절
  방어: 멱등성 키 + 유니크 제약          방어: 행 잠금 / 원자적 조건부 갱신
```

둘은 함께 필요합니다. 멱등성 키만 있으면 동시 경쟁을 못 막고, 잠금만 있으면 재시도로 들어온 같은 요청을 한 번 더 처리해 버립니다. 선불 차감은 보통 **멱등성 키(같은 요청 방어) + 원자적 조건부 갱신 또는 행 잠금(다른 요청 경쟁 방어)** 을 같이 씁니다.

## 4. 트랜잭션과 격리 수준 심층

지금까지의 해법은 "행 잠금"과 "조건부 갱신"이라는 메커니즘에 기대고 있었습니다. 이 메커니즘이 실제로 어떻게 동작하는지는 트랜잭션 격리 수준(isolation level)에 달려 있습니다. 사용자가 특히 깊게 보고 싶어 하는 `READ COMMITTED`, 팬텀 읽기, 전파를 차감 작업에 붙여서 봅니다.

### 4.1 차감 작업을 위협하는 이상 현상은 둘이다

격리 수준이 막거나 허용하는 이상 현상(anomaly)의 일반 카탈로그(더러운 읽기/dirty read, 반복 불가능 읽기/non-repeatable read, 팬텀 읽기/phantom read, 직렬화 이상/serialization anomaly)는 [트랜잭션 고립 수준과 이상 현상](../../database/deep-dive/transactions/13-isolation-anomalies.md)에 schedule 단위로 정리돼 있습니다. 선불 차감에서 실제로 사고를 내는 현상은 그중 둘입니다.

- 단건 잔액 차감 → **갱신 손실(lost update)**. 한 행을 두 트랜잭션이 읽은 값 기준으로 각자 갱신하면 한쪽이 사라진다. 1장의 사고가 이것이다.
- 집계 한도 검증 → **write skew(쓰기 왜곡)**. 여러 행을 합산해서 "한도 미만"을 확인한 두 트랜잭션이 각자 새 행을 추가하면, 각자는 정당했지만 합치면 한도를 넘는다. 4.3에서 다룬다.

이 둘을 구분하는 게 중요합니다. **갱신 손실은 "같은 행"의 문제**라 그 행을 잠그면 막히고, **write skew는 "조건에 맞는 행들의 집합"의 문제**라 단순히 기존 행을 잠그는 것만으로는 막히지 않습니다. 방어 수단이 다릅니다.

### 4.2 READ COMMITTED와 REPEATABLE READ: 스냅샷의 함정

먼저 두 격리 수준이 "읽기"를 어떻게 다루는지 정확히 잡습니다. MySQL InnoDB 공식 문서 기준입니다.

> **InnoDB의 기본 격리 수준은 `REPEATABLE READ`입니다.** ("This is the default isolation level for InnoDB." — MySQL Reference Manual)

이 사실은 자주 오해됩니다. 많은 애플리케이션이 `READ COMMITTED`로 운영되는 이유는 InnoDB의 기본값이 그래서가 아니라, **프레임워크나 설정에서 명시적으로 바꾸기 때문**입니다(6장의 kcf 사례가 정확히 그렇습니다). 그래서 "우리 DB는 MySQL이니 READ COMMITTED겠지"라고 가정하면 안 되고, 실제 설정을 확인해야 합니다.

두 수준의 읽기 규칙은 이렇게 다릅니다.

- `REPEATABLE READ`에서 일반 `SELECT`(잠그지 않는 일관된 읽기, consistent read)는 **트랜잭션의 첫 읽기 시점에 만든 스냅샷**을 트랜잭션 끝까지 봅니다. 그래서 같은 행을 두 번 읽으면 같은 값이 나옵니다.
- `READ COMMITTED`에서 일반 `SELECT`는 **매 문장마다 새 스냅샷**을 잡습니다. 그래서 다른 트랜잭션이 그사이 커밋하면 두 번째 `SELECT`는 새 값을 봅니다(반복 불가능 읽기).

여기서 차감 로직의 **스냅샷 함정**이 나옵니다. 직관적으로는 "`REPEATABLE READ`가 더 엄격하니 안전하겠지"라고 생각하기 쉽지만, 잠그지 않는 `SELECT`로 잔액을 읽어서 차감을 결정하면 `REPEATABLE READ`가 오히려 **오래된 값**을 보게 만듭니다.

```text
REPEATABLE READ에서 "일반 SELECT로 읽고 UPDATE"가 위험한 이유

t1  A: BEGIN
t2  A: SELECT balance -> 100        (스냅샷 고정: A는 끝까지 100을 본다)
t3                                   B: BEGIN; UPDATE balance=20; COMMIT
t4  A: 다시 SELECT balance -> 100    (여전히 스냅샷 100, 실제 DB는 20)
t5  A: UPDATE balance = 100 - 80     (오래된 100 기준으로 덮어씀)  <- 사고
```

그래서 격리 수준을 올리는 것은 갱신 손실의 해법이 아닙니다. 해법은 3.2처럼 **잠그는 읽기(locking read)** 나 **조건부 갱신**을 쓰는 것입니다. 잠그는 갱신은 대상 행에 배타적 잠금을 잡은 뒤, 스냅샷이 아니라 그 행의 **최신 커밋 버전을 다시 읽어(current read)** `WHERE` 조건을 재평가합니다. 특히 `READ COMMITTED`에서는 대상 행이 이미 다른 트랜잭션에 잠겨 있으면 **준일관 읽기(semi-consistent read)** 로 최신 커밋 버전을 먼저 보고, 조건에 맞을 때만 잠금을 기다립니다(MySQL 문서). 어느 경우든 원자적 조건부 갱신의 `AND balance >= :useAmount`는 스냅샷이 아니라 최신 확정값으로 검사되므로, `READ COMMITTED`든 `REPEATABLE READ`든 안전하게 동작합니다.

정리하면, 차감 정확성에 대해 격리 수준이 주는 메시지는 이렇습니다.

> **격리 수준을 올린다고 갱신 손실이 막히지 않습니다. 갱신 손실은 "잠그는 읽기 또는 조건부 갱신"으로 막고, 격리 수준은 그 위에서 잠금 범위와 팬텀 방어를 결정합니다.**

### 4.3 팬텀 읽기와 write skew: 한도 검증이 무너지는 자리

단건 잔액이 아니라 **합산 한도**를 검증할 때는 다른 종류의 사고가 납니다. 예를 들어 "이 계좌의 오늘 입금 총액이 일별 한도를 넘으면 거절"이라는 규칙을 생각해 봅니다. 한도는 1,000원이고 오늘까지 누적이 600원이라고 합시다. 처리계는 보통 이렇게 합니다.

```text
1) SELECT SUM(amount) FROM tx WHERE account=? AND date=today   -> 오늘까지 600
2) if 600 + thisAmount > 1000:  reject     (한도 1000)
3) else: INSERT INTO tx (..., amount=thisAmount)
```

이것을 두 트랜잭션이 동시에 하면, 1장과 비슷하지만 대상이 "행 하나"가 아니라 "조건에 맞는 행들의 집합"이라는 점이 다릅니다. 각자 300원씩 입금하는 두 요청을 시간 위에 펼쳐 봅니다. 한 건씩 보면 600 + 300 = 900으로 한도 안이지만, 둘이 겹치면 합계가 1,200이 됩니다.

```text
한도 1000, 현재 합계 600, 두 요청 모두 300원

time | 트랜잭션 A                          | 트랜잭션 B                          | 실제 합계
-----+-------------------------------------+-------------------------------------+----------
t1   | SUM -> 600                          |                                     | 600
t2   |                                     | SUM -> 600                          | 600
t3   | 600 + 300 = 900 <= 1000  -> 통과    |                                     | 600
t4   |                                     | 600 + 300 = 900 <= 1000  -> 통과    | 600
t5   | INSERT 300                          |                                     | 900
t6   |                                     | INSERT 300                          | 1200  <- 한도 1000 초과!
```

각 트랜잭션은 자기가 본 스냅샷 안에서 정당한 결정을 했습니다. 그런데 합쳐 놓으면 어떤 직렬 순서로도 설명되지 않는 결과(한도 초과)가 나옵니다. 이것이 **write skew(쓰기 왜곡)** 이고, 그 바탕에는 **팬텀 읽기(phantom read)**, 즉 "조건에 맞는 행 집합이 그사이 바뀌는" 현상이 있습니다. `A`가 본 "오늘 입금 집합"에 `B`가 새 행을 끼워 넣었고, 둘은 서로의 새 행을 보지 못했습니다.

기존 행을 잠그는 것만으로는 이걸 못 막습니다. **아직 존재하지 않는 행(B가 넣을 INSERT)** 이 문제이기 때문입니다. 방어 수단은 다음과 같습니다.

1. 집계를 잠글 수 있는 한 행으로 물질화(materialize)한다.

    "계좌별·일자별 누적 합계"를 별도 카운터 행 하나로 유지하고, 그 행을 `FOR UPDATE`로 잠그거나 원자적 조건부 갱신으로 올립니다. 그러면 집합 문제가 다시 단건 행 문제로 환원되어 3.2의 해법이 그대로 통합니다.

    ```sql
    UPDATE daily_limit_counter
       SET used_amount = used_amount + :amount
     WHERE account_id = :id AND day = :today
       AND used_amount + :amount <= :limit;   -- 한도 안일 때만
    -- affected=0 이면 한도 초과로 거절
    ```

2. 범위에 갭 잠금(gap lock)을 건다.

    `REPEATABLE READ`의 잠그는 읽기는 **next-key lock**(행 잠금 + 그 앞 갭 잠금)으로 범위를 잠가, 다른 트랜잭션이 그 범위에 새 행을 끼워 넣지 못하게 막습니다. MySQL 문서는 이를 "팬텀을 막기 위해 InnoDB가 쓰는 알고리즘"이라고 설명합니다. 반대로 `READ COMMITTED`는 **갭 잠금을 거의 끄기** 때문에(외래 키·중복 키 검사 외) 범위에 새 행이 끼어들 수 있어 팬텀이 가능합니다.

3. `SERIALIZABLE`로 올린다.

    가장 강하게 막지만 잠금과 직렬화 비용이 커서, 보통은 1번(집계 행 물질화)이 더 실용적입니다.

여기서 `READ COMMITTED`의 성격이 드러납니다. `READ COMMITTED`는 갭 잠금을 거의 끄므로 잠금 대기와 교착이 적어 동시성이 좋지만, **범위 기반 한도 검증을 DB 격리만으로는 지켜 주지 않습니다.** 그래서 `READ COMMITTED`에서 한도를 정확히 막으려면 1번처럼 집계를 단건 행으로 만들어 명시적으로 잠가야 합니다. 이 점이 6장의 kcf 사례에서 그대로 살아납니다.

### 4.4 트랜잭션 범위와 전파(propagation)

마지막으로 "트랜잭션을 어디서 시작해서 어디서 끝낼 것인가", 그리고 "이미 트랜잭션이 있을 때 어떻게 합칠 것인가"를 봅니다.

전파(propagation)는 메서드가 호출됐을 때 기존 트랜잭션과의 관계를 정합니다. 차감과 직접 관련된 둘만 봅니다.

1. `REQUIRED`

    기존 트랜잭션이 있으면 참여하고, 없으면 새로 만듭니다. **잔액 차감 + 원장 기록 + 상태 변경처럼 "모두 함께 커밋되거나 함께 롤백되어야 하는" 작업은 보통 하나의 `REQUIRED` 트랜잭션으로 묶습니다.** 차감은 됐는데 원장이 안 남으면 금액이 새기 때문입니다.

2. `REQUIRES_NEW`

    기존 트랜잭션을 잠시 멈추고(suspend) **독립된 새 트랜잭션**을 만들어, 바깥이 롤백돼도 살아남게 합니다. 감사 로그(audit log), 아웃박스(outbox), 시도 기록처럼 "본 작업이 실패해도 남아야 하는 기록"에 씁니다. 다만 `REQUIRES_NEW`는 커넥션을 하나 더 빌리므로, 남발하면 커넥션 풀과 교착 위험이 함께 커집니다.

트랜잭션 범위에서 차감과 관련해 가장 중요한 규칙은 **"잠금을 쥔 채 외부 시스템을 호출하지 말라"** 입니다. 잔액 행을 `FOR UPDATE`로 잠근 트랜잭션 안에서 외부 결제망·은행 전문 호출을 하면, 그 호출이 느리거나 타임아웃 나는 동안 행 잠금과 DB 커넥션을 계속 점유합니다. 같은 계좌를 기다리는 다른 요청이 줄줄이 막히고, 풀이 마르면 무관한 요청까지 영향을 받습니다. 그래서 외부 호출이 끼는 흐름은 "로컬 의도를 먼저 확정(커밋)하고, 외부 호출 결과를 상태로 닫는" 형태로 분리하는 편이 안전합니다. 이 경계의 전체 그림은 [애플리케이션 경계 문서](../../interviews/database-deep-dive/12-application-boundaries-idempotency-money-outbox.md)와 연결됩니다.

전파의 물리/논리 트랜잭션, 커밋 시점, 프록시 동작, self-invocation, MyBatis 연동의 세부는 [Spring Transactional](../../jvm/spring/spring_transactional.md)에 깊게 정리돼 있습니다. 격리·락·교착의 일반 메커니즘은 [전파·격리·락·데드락](../../interviews/database-deep-dive/08-isolation-lock-deadlock.md)을 함께 보면 됩니다.

## 5. 일반 프레임워크에서의 처리

원리를 일반적인 ORM/트랜잭션 프레임워크가 어떻게 표현하는지 봅니다. 이름은 달라도 결국 4장까지의 메커니즘 중 하나로 내려갑니다.

- 선언형 트랜잭션: 대부분의 프레임워크는 메서드 경계에 트랜잭션을 선언하는 방식(Spring `@Transactional`, JTA, 각 언어의 데코레이터/애너테이션)을 제공합니다. 선언형은 트랜잭션 시작·커밋·롤백 코드를 비즈니스 로직에서 걷어내 주지만, **경계가 사라진 게 아니라 프레임워크의 가로채기 지점으로 옮겨졌을 뿐**입니다. 그래서 프록시를 지나지 않는 호출(self-invocation), 잘못된 트랜잭션 매니저 선택, 예외를 삼켜 롤백이 안 되는 경우는 "애너테이션을 붙였는데 안 묶이는" 사고로 나타납니다.
- 원자적 조건부 갱신: ORM에서도 결국 `UPDATE ... WHERE balance >= ?` 같은 쿼리를 직접 쓰거나(쿼리 매퍼), 커스텀 업데이트 쿼리로 내려보냅니다. ORM의 더티 체킹(dirty checking)으로 "엔티티를 읽어서 필드를 바꾸고 저장"하는 흐름은 1장의 read-modify-write와 같으므로, 차감에는 적합하지 않습니다.
- 낙관적 락: 많은 ORM이 버전 컬럼을 1급으로 지원합니다(JPA `@Version` 등). 갱신 시 자동으로 `WHERE version = ?`를 붙이고, 충돌이면 예외를 던집니다. 애플리케이션은 그 예외를 잡아 재시도합니다.
- 비관적 락: 리포지토리·쿼리 수준에서 `FOR UPDATE`를 거는 잠금 모드를 제공합니다(JPA `LockModeType.PESSIMISTIC_WRITE` 등). 프레임워크가 방언(dialect)에 맞는 잠금 절을 생성합니다.
- 실행 시점 차이: 영속성 컨텍스트가 있는 ORM(JPA류)은 `save`를 호출해도 실제 SQL이 플러시(flush) 시점까지 미뤄질 수 있습니다. 반면 쿼리 매퍼(MyBatis류)는 매퍼 호출 시점에 SQL이 바로 나갑니다. 차감처럼 "언제 잠금이 잡히고 언제 갱신이 반영되는가"가 정확성에 직결되는 작업에서는 이 실행 시점 차이를 반드시 알고 있어야 합니다.

즉 프레임워크는 새로운 정확성 보장을 만들어 주지 않습니다. **3장의 세 전략을 더 편하게 표현해 줄 뿐**이고, 어떤 전략을 고르고 어디에 둘지는 여전히 설계자의 몫입니다.

## 6. kcf-projects strict-io 구체화

이제 가장 구체적인 층으로 내려갑니다. Spring Boot + MyBatis + MySQL을 쓰는 strict-io 환경의 실제 코드(`kcf-projects`의 펌뱅킹 서비스)에서 위 원리가 어디에 어떻게 앉는지 봅니다. 아래 코드·구조 인용은 모두 저장소에서 직접 확인한 사실입니다.

### 6.1 strict-io 계층과 트랜잭션 경계

strict-io는 모든 서비스 작업을 입력·출력 객체 한 쌍으로 명시하고, 계층 책임을 `Channel -> SO -> BO -> Mapper -> Util`로 고정하는 구조입니다(`docs/STRICT_IO_PROJECT_SHAPE_TEMPLATE.md`). 각 계층의 책임은 이렇게 갈립니다.

- Channel: HTTP/소켓/배치 입력을 받아 요청 봉투를 풀고, 서비스 코드 검증과 공통 응답 봉투를 만든다.
- SO(ServiceObject): 외부 유스케이스를 조율(orchestration)하고, 공개 입출력을 해석하며, **트랜잭션 경계 진입을 판단**한다.
- BO(BusinessObject): 내부 업무 규칙과 상태 전이, 매퍼·외부 어댑터 조합을 담당한다.
- Mapper: DB I/O, SQL, 행 매핑만 담당한다.
- Util: 순수·무상태 공통 함수.

SO와 BO는 자기 이름의 입출력 한 쌍을 가지며, 최상위 공개 메서드는 `service(ctx, in)` 하나로 둡니다. 인터페이스 자체가 이를 강제합니다.

```java
// local-sfm-shims/sfm5-core/.../ServiceObject.java (직접 확인)
public interface ServiceObject<IN, OUT> {
    public OUT service( ServiceContext ctx, IN in ) throws Exception ;
}
```

트랜잭션 경계는 **SO에 선언**됩니다. 가상계좌 수취조회중계처리 SO의 실제 선언은 이렇습니다.

```java
// kcf-firmbanking-service/.../service/fir/so/Kfsfir02u0.java (직접 확인)
@Service("kfsfir02u0")
@Transactional(value = "kcffirmTransactionManager",
               isolation = Isolation.READ_COMMITTED,
               propagation = Propagation.REQUIRES_NEW,
               rollbackFor = Exception.class)
public class Kfsfir02u0 implements ServiceObject<Kfsfir02u0In, Kfsfir02u0Out> {
    @Autowired BusinessObject<BfirVacnTrLstPreChkIn, BfirVacnTrLstPreChkOut> bfirVacnTrLstPreChk; // 사전검증
    @Autowired BusinessObject<BfirVacnTrLstRegIn,    BfirVacnTrLstRegOut>    bfirVacnTrLstReg;    // 등록
    @Autowired BusinessObject<BfirVacnTrLstUpdIn,    BfirVacnTrLstUpdOut>    bfirVacnTrLstUpd;    // 변경
    // ...
}
```

여기서 4장의 두 개념이 동시에 보입니다.

1. `isolation = READ_COMMITTED`

    앞서 InnoDB 기본값은 `REPEATABLE READ`라고 했는데, 이 코드는 **명시적으로 `READ COMMITTED`로 내립니다.** 따라서 이 서비스는 갭 잠금이 거의 꺼진 환경에서 동작합니다. 일반 `SELECT`는 문장마다 최신 커밋을 보고, 범위 삽입(팬텀)은 격리만으로는 막히지 않습니다. 4.3에서 예고한 "한도 검증은 명시적 잠금이 필요하다"가 이 설정 때문에 그대로 적용됩니다.

2. `propagation = REQUIRES_NEW`

    각 외부 유스케이스(SO)를 독립된 트랜잭션으로 처리합니다. 호출자 트랜잭션과 분리되어, 이 처리 단위가 그 자체로 원자적인 한 건이 됩니다. 같은 SO가 여러 BO(사전검증 → 등록 → 변경)를 부르더라도 하나의 트랜잭션 안에서 함께 커밋·롤백됩니다.

strict-io 템플릿은 "트랜잭션 래퍼, 전파 경계, datasource/session template"과 "DB 행 생성·조회·갱신 순서, lock/readback, 상태 전이"를 **하드 컨트랙트(hard contract)**, 즉 함부로 바꾸면 외부 호환성·트랜잭션·DB 상태가 깨지는 축으로 못 박습니다. 즉 이 프로젝트의 규범 자체가 "어떤 read가 lock을 잡는가"를 명시적으로 검토하라고 요구합니다(템플릿 9절 DB-first state machine).

### 6.2 현재 펌뱅킹 입금 흐름이 실제로 하는 일

펌뱅킹 가상계좌 입금/수취조회 흐름은 일별 한도(4.3의 집계 한도)와 예치금 잔액 검증을 함께 수행합니다. 한도 상태는 enum으로 명시돼 있습니다.

```java
// kcf-firmbanking-service/.../service/rsi/VacnDdyLmtSts.java (직접 확인)
public enum VacnDdyLmtSts {
    NO_LIMIT("한도 없음"),
    UNDER_LIMIT("한도 미만"),
    VIOLATE_VACN_DDY_TOT_ROM_LMT_CNT("가상계좌 일별 총입금한도 건수(vacnDdyTotRomLmtCnt) 위반"),
    VIOLATE_VACN_DDY_TOT_ROM_LMT_AMT("가상계좌 일별 총입금한도 금액(vacnDdyTotRomLmtAmt) 위반"),
    VIOLATE_VACN_BYCCS_ROM_LMT_AMT("가상계좌 건별 입금한도 금액(vacnBycsRomLmtAmt) 위반");

    public boolean canDeposit() {                      // 한도 없음 또는 한도 미만일 때만 입금 허용
        return NO_LIMIT.equals(this) || UNDER_LIMIT.equals(this);
    }
}
```

이 검증을 담당하는 사전검증 BO(`BrsiRomRcvgPreChk`)가 핵심입니다. 이 BO는 검사를 시작할 때 **하위이용기관 원장(마스터) 행을 `SELECT ... FOR UPDATE`로 잠그고 읽습니다.** 그 행 하나에 예치금 잔액(`blnc_rmd`)과 한도 설정값(`vacn_ddy_tot_rom_lmt_amt` 등)이 함께 들어 있습니다.

```sql
-- FRM_USNSG_MST.xml ps002 (직접 확인): 마스터 행을 잠그고 읽는다
SELECT ..., blnc_rmd,                                          -- 예치금 잔액
       vacn_bycs_rom_lmt_amt, vacn_ddy_tot_rom_lmt_amt,
       vacn_ddy_tot_rom_lmt_cnt                                -- 한도 설정값
  FROM FRM_USNSG_MST
 WHERE ...
   FOR UPDATE;                                                 -- 이 행을 배타적으로 잠근다
```

잠그고 읽은 뒤, 그 잠금 구간 안에서 일별 누적을 별도 합산(`SUM`, 잠금 없음)으로 읽어 한도를 검사합니다.

```java
// BrsiRomRcvgPreChk.java (직접 확인)
frm_usnsg_mst_ps002_out540 = frm_usnsg_mst.ps002(...);   // line 104: 위 FOR UPDATE 읽기
// ... 마스터 행 잠금을 쥔 채로 일별 누적을 합산으로 읽는다
var ddyVacnTotTrAmt = toBigDecimal(frm_rsldrt_lst_ps003_out.get("ddy_vacn_tot_tr_amt")); // line 301
// 한도 검사 (line 332)
if (vacnDdyTotRomLmtAmt.compareTo(BigDecimal.ZERO) > 0
        && ddyVacnTotTrAmt.add(trAmt).compareTo(vacnDdyTotRomLmtAmt) > 0)
    return VacnDdyLmtSts.VIOLATE_VACN_DDY_TOT_ROM_LMT_AMT;
```

그리고 잔액은 같은 마스터 행에 **원자적으로 누적**됩니다. 애플리케이션이 읽어서 더하지 않고 DB가 더합니다.

```sql
-- FRM_USNSG_MST.xml pu001 (직접 확인)
UPDATE FRM_USNSG_MST SET ..., blnc_rmd = blnc_rmd + #{blnc_rmd} WHERE ...
```

이 코드에는 1장의 read-modify-write 사고를 막는 두 메커니즘(3.2의 비관적 락 + 원자적 누적)이 모두 들어 있습니다.

- `ps002`의 `FOR UPDATE`가 하위이용기관 마스터 행에 배타적 잠금을 잡습니다. `BrsiRomRcvgPreChk`는 이 잠금을 쥔 채로 일별 누적(`ps003`)·한도·예치금을 검사하므로, 같은 하위이용기관의 다른 검사 트랜잭션은 이 잠금을 기다리며 **직렬화**됩니다. 4.3의 write skew가 "잠글 수 있는 한 행"으로 환원됩니다.
- 잔액 자체는 `blnc_rmd = blnc_rmd + #{}`로 DB가 더합니다. 애플리케이션이 읽어서 더하지 않으므로, 이 쓰기 한 줄만 놓고 보면 1장의 read-modify-write 갱신 손실이 생기지 않습니다.

`BrsiRomRcvgPreChk` 자신은 `@Transactional(... isolation = READ_COMMITTED, propagation = REQUIRED ...)`라(직접 확인), 이를 호출하는 SO(예: `Kfsfco02u0`도 `REQUIRED`)의 트랜잭션에 합류해 그 SO가 커밋할 때까지 `FOR UPDATE` 잠금을 유지합니다.

같은 요청의 중복(멱등성)은 별도 장치로 다룹니다. 저장소는 거래일련번호 같은 업무 키에 유니크 제약을 두고, 중복 삽입 시 `DuplicateKeyException`을 잡아 **별도 결과코드(`RC_DUP`, `KFS005`)로 거절**하는 패턴을 씁니다(예: 거래 등록 BO `Bf88InbAdcReg`, 직접 확인). 이는 같은 업무 키의 두 번째 삽입을 DB가 막는다는 뜻이지, 첫 결과를 그대로 재생(replay)한다는 뜻은 아닙니다. 이를 멱등으로 닫으려면 호출자가 `RC_DUP`를 "이미 처리됨"으로 해석하는 계약이 필요합니다(3.4). 즉 **다른 요청의 동시 경쟁은 마스터 행 `FOR UPDATE`로, 같은 요청의 재시도는 업무 키 유니크 제약으로** 막습니다.

마지막으로 이 문서가 끝까지 추적하지 못해 `추가 확인 필요`로 남기는 두 경계가 있습니다.

1. 잠금 단위와 한도 범위의 일치

    `FOR UPDATE`는 하위이용기관 마스터 행 단위로 걸리고, 일별 한도 합계(`ps003`)는 하위이용기관 + 가상계좌번호 + 일자 단위로 집계됩니다. 한 가상계좌의 일별 합계에 영향을 주는 입금이 모두 같은 마스터 행 잠금을 공유해야 write skew가 완전히 닫힙니다.

2. 검사 잠금과 잔액 누적의 트랜잭션 관계

    검사의 `FOR UPDATE`(`BrsiRomRcvgPreChk`, `REQUIRED`)와 잔액 누적 `pu001`이 항상 한 트랜잭션에 같이 있지는 않습니다. `pu001`은 `BrsiRomPcsRsltRfl`, `BrsiRomDpcBlncRmdRfl` 같은 결과반영 BO에서 호출되고, 이들은 대부분 `propagation = REQUIRES_NEW`(독립 트랜잭션)입니다(직접 확인). 즉 "검사 때 거는 잠금"과 "나중에 반영하는 누적"이 입금 생애주기에서 어떻게 맞물리는지는 더 추적해야 합니다.

두 경계 모두 strict-io 템플릿의 DB-first state machine 검토(템플릿 9절, "어떤 read가 lock을 잡는지")가 코드에서 보이게 만들라고 요구하는 지점입니다. 다만 확실한 것은, 무방비 read-modify-write가 아니라 `FOR UPDATE`(검사 직렬화)와 원자적 누적(쓰기)이라는 두 방어가 모두 코드에 존재한다는 점입니다.

### 6.3 두 가지 가드 모양과 strict-io 적합성

6.2에서 본 것처럼 kcf 입금 흐름은 검사 구간을 **마스터 행 `FOR UPDATE`(비관적 락)** 로 직렬화하고, 잔액을 **원자적으로 누적**합니다. 3.2의 2번과 4.3의 1번 패턴이 코드에 실제로 들어 있는 셈입니다. 같은 strict-io 계층 책임 안에서 차감 경로를 새로 설계할 때도 작업 성격에 따라 두 모양 중 하나를 고릅니다. 둘 다 SO는 트랜잭션 경계와 조율만, BO는 업무 규칙, Mapper는 잠금/조건을 가진 SQL을 맡는 구조에 맞습니다.

1. 단건 잔액만 차감하면 원자적 조건부 갱신 매퍼

    잔액과 한도·원장을 함께 볼 필요가 없는 순수 차감이라면, `FOR UPDATE` 없이 조건부 `UPDATE` 한 문장이 더 가볍습니다. 영향 행 수를 검사해 `ServiceException`으로 처리합니다. strict-io 관례인 "정확히 한 행 계약은 공통 ensure helper로"와 "`ServiceException(rc, mcd, msg)`로 실패 신호"에 그대로 맞습니다.

    ```xml
    <!-- 매퍼 XML: 조건부 차감 -->
    <update id="pu002" parameterType="hashmap">
        UPDATE PREPAID_BALANCE
           SET balance = balance - #{useAmount}
             , last_chg_dt = #{last_chg_dt}
         WHERE account_id = #{account_id}
           AND balance   >= #{useAmount}
    </update>
    ```

    ```java
    // BO 안에서 (개념 예시; rc/mcd/ServiceException는 저장소 실제 관례)
    int affected = prepaidBalanceMapper.pu002(params);
    if (affected != 1) {
        // 잔액 부족 또는 동시 경쟁에서 밀림 -> 업무 실패로 처리
        throw new ServiceException(ServiceGlobal.RC_ERR, "KFSxxx", "지급가능잔액이 부족합니다.");
    }
    ```

    `ServiceException`은 저장소에서 `rc`(결과코드), `mcd`(메시지코드), 메시지를 들고 다니는 타입입니다(직접 확인). 잔액 부족을 이 타입으로 처리하면 채널까지 기계가 읽을 수 있는 실패 신호가 전달됩니다.

2. 잔액·한도·원장을 함께 결정하면 비관적 락 (kcf가 쓰는 방식)

    여러 값을 함께 보고 결정해야 하면 6.2의 kcf 흐름처럼 마스터 행을 `SELECT ... FOR UPDATE`로 잠근 뒤, 그 안에서 한도 합계·잔액을 검사하고 갱신합니다. 집계 한도(4.3)가 잠글 수 있는 한 행으로 환원되어 `READ COMMITTED`에서도 정확해집니다.

어느 쪽이든 핵심은 strict-io 규범과 일치합니다. **정확성 가드는 사전검증의 Java 비교 자체가 아니라, "잠그는 읽기를 하는 매퍼" 또는 "조건을 가진 갱신 매퍼"에 있습니다.** 사전검증의 Java 비교는 그것이 잠금 구간 안에 있을 때 유효하고(6.2의 kcf가 그렇습니다), 잠금 밖이라면 2장의 빠른 거절 역할로만 봐야 합니다.

### 6.4 MySQL·InnoDB·HikariCP에서 확인할 점

마지막으로 이 스택에서 실무적으로 점검할 항목입니다.

- 격리 수준: 코드가 `READ COMMITTED`로 명시했으므로 갭 잠금에 기댈 수 없다. kcf는 그래서 한도·잔액을 마스터 행 `FOR UPDATE`(6.2)로 명시적으로 잠가 직렬화한다. 순수 단건 잔액 차감만이라면 조건부 갱신으로도 `READ COMMITTED`에서 안전하다.
- 갭 잠금: `READ COMMITTED`는 갭 잠금이 거의 꺼져 교착이 적은 대신, 범위 삽입을 막지 못한다. 그래서 범위 한도는 갭 잠금 대신 마스터/카운터 행 잠금으로 환원하는 편이 `READ COMMITTED`에 더 잘 맞는다. 굳이 범위 잠금이 필요해 `REPEATABLE READ`로 올린다면 next-key lock이 늘어 교착 가능성도 함께 본다.
- 잠금 점유와 전파: kcf 입금 흐름은 마스터 행 `FOR UPDATE`를 트랜잭션 끝까지 쥔다(`REQUIRED`). 그 사이 은행 전문 같은 외부 호출이 끼면 같은 하위이용기관의 다른 입금이 그만큼 대기하므로, 잠금 구간 안의 외부 호출 시간과 HikariCP 풀 크기를 함께 본다. SO를 `REQUIRES_NEW`로 두면 커넥션을 추가로 빌리는 점도 같이 본다.
- 금액 타입: 거래금액이 DTO 경계에서 `String`으로 오가고 쓰기 직전 `BigDecimal`로 변환되는 모양이라면(저장소 관례), 저장 컬럼의 정밀도/스케일과 반올림 방식을 한 곳에서 고정하고, `balance >= 0` 같은 불변식을 가능하면 DB 제약으로도 둔다.
- 교착 순서: 잔액 행과 원장 행을 함께 잠근다면 항상 같은 순서로 잠가 교착을 피하고, 교착 발생 시 재시도 경로를 둔다.

## 7. 실패 모드와 직접 검증

설명을 덮고도 직접 재현할 수 있어야 이해가 닫힙니다. 로컬 MySQL(InnoDB) 세션 두 개로 1장의 사고와 3.2의 방어를 눈으로 확인합니다.

먼저 위험한 버전(read-modify-write)을 두 세션에서 교차 실행합니다.

```sql
-- 준비
CREATE TABLE prepaid_account (account_id BIGINT PRIMARY KEY, balance BIGINT NOT NULL);
INSERT INTO prepaid_account VALUES (1, 100);

-- 세션 A                              -- 세션 B
SET autocommit=0;                       SET autocommit=0;
BEGIN;                                  BEGIN;
SELECT balance FROM prepaid_account     SELECT balance FROM prepaid_account
 WHERE account_id=1;     -- 100          WHERE account_id=1;     -- 100
-- 둘 다 100을 봤다
UPDATE prepaid_account SET balance=20    UPDATE prepaid_account SET balance=20
 WHERE account_id=1;                     WHERE account_id=1;   -- A 커밋 후 대기/덮어쓰기
COMMIT;                                 COMMIT;
SELECT balance;  -- 20                  SELECT balance;  -- 20
```

- FAIL 신호: 두 세션 모두 성공했고 최종 잔액이 20이다. 160원어치를 처리하고 80만 깎였다(초과 차감).

이제 안전한 버전(원자적 조건부 갱신)으로 바꿔 같은 경쟁을 만듭니다.

```sql
INSERT INTO prepaid_account VALUES (1, 100) 
  ON DUPLICATE KEY UPDATE balance=100;   -- 잔액 100으로 초기화

-- 세션 A                                          -- 세션 B
BEGIN;                                              BEGIN;
UPDATE prepaid_account SET balance=balance-80
 WHERE account_id=1 AND balance>=80;
SELECT ROW_COUNT();  -- 1   (A가 1행 갱신)
                                                   UPDATE prepaid_account SET balance=balance-80
                                                    WHERE account_id=1 AND balance>=80;
                                                   -- ^ A가 쥔 행 잠금을 기다리며 블로킹
COMMIT;            -- 잠금 해제, balance=20 확정
                                                   -- 잠금 풀린 뒤 B의 UPDATE 진행: balance=20이라 WHERE 거짓
                                                   SELECT ROW_COUNT();  -- 0   (B는 거절)
                                                   COMMIT;
SELECT balance;  -- 20                              SELECT balance;  -- 20
```

`ROW_COUNT()`는 바로 앞 문장의 영향 행 수를 돌려주므로, `COMMIT` 전에 `UPDATE` 직후 호출해야 위 값이 보입니다.

- PASS 신호: 한 세션만 1행을 갱신해 성공하고, 다른 세션은 0행이라 "잔액 부족"으로 거절된다. 80원어치만 처리되고 잔액은 20이다.
- 추가 관찰: 세션 B의 `UPDATE`는 세션 A가 커밋할 때까지 행 잠금을 기다린다(잠그는 쓰기). A 커밋 후 B는 스냅샷이 아니라 최신 확정값 20을 보고 `WHERE balance>=80`이 거짓이 되어 0행이 된다. 4.2에서 말한 "잠금을 기다린 뒤 최신 확정값으로 조건을 재평가"하는 모습이다.

한도(집계) 버전은 4.3의 카운터 행으로 같은 실험을 만들 수 있습니다. 두 세션이 동시에 `used_amount + :amt <= :limit` 조건부 갱신을 시도하면, 합이 한도를 넘기는 쪽은 0행으로 거절됩니다.

손으로 다시 떠올려 보는 질문입니다.

- 1장의 `t1`~`t6`에서, `B`가 100을 다시 읽지 못하게 하려면 `t2`의 `SELECT`를 무엇으로 바꿔야 하는가?
- `READ COMMITTED`로 올리면 1장의 갱신 손실이 막히는가? 왜 막히지 않는가?
- 단건 잔액 차감과 일별 한도 검증은 같은 `FOR UPDATE` 한 줄로 똑같이 막을 수 있는가? 아니라면 무엇이 다른가?
- `@Transactional`을 `REQUIRES_NEW`로 둔 SO 안에서 외부 은행 전문 호출이 5초 걸린다면, 어떤 자원이 5초 동안 점유되는가?

## 8. 한 장 정리

선불 잔액 사용처리의 정확성은 결국 "검사한 사실을 쓰기까지 유효하게 만드는가"로 수렴합니다. 핵심 축을 먼저 평서문으로 정리합니다.

- 사고의 정체: 단건 잔액은 **갱신 손실(lost update)**, 집계 한도는 **write skew**. 둘 다 결과는 초과 차감/한도 초과다.
- 사전 검증의 위치: 빠른 거절을 위한 최적화일 뿐, 정확성의 보증 장치가 아니다. 보증은 값을 바꾸는 지점에서 한 번 더 지킨다.
- 격리 수준의 한계: 격리를 올린다고 갱신 손실이 막히지 않는다. 잠그는 읽기나 조건부 갱신이 막고, 격리는 그 위에서 잠금 범위와 팬텀 방어를 정한다.
- 멱등성과 동시성: 같은 요청 두 번은 멱등성 키로, 다른 요청의 경쟁은 잠금/조건부 갱신으로. 둘은 함께 필요하다.

같은 내용을 사고 → 메커니즘 → strict-io에서의 위치로 묶으면 이렇습니다.

| 위협 | 막는 메커니즘(기본값) | 격리/락 관점 | strict-io에서 두는 곳 |
| --- | --- | --- | --- |
| 단건 잔액 갱신 손실 | 원자적 조건부 갱신 `... WHERE balance >= :amt` | 행 잠금 후 최신 확정값으로 조건 재검사, RC/RR 모두 안전 | 차감 Mapper + 영향행수 검사(BO) |
| 잔액+원장 다단계 결정 | 비관적 락 `SELECT ... FOR UPDATE` | 행 잠금으로 경쟁 직렬화 | BO에서 잠금 읽기 후 갱신, SO가 트랜잭션 경계 |
| 집계 한도 write skew | 잠글 수 있는 마스터/카운터 행 `FOR UPDATE`/조건부 증가 또는 RR 갭 잠금 | RC는 갭 잠금 꺼짐 → 명시적 행 잠금 필요 | kcf 방식: 마스터 행 `FOR UPDATE`로 직렬화(비관적 락) |
| 같은 요청 재시도 | 멱등성 키 유니크 제약 | 중복 키 → 삽입 거절 | 업무 키 유니크 + `DuplicateKeyException` → `RC_DUP` 거절(호출자가 멱등 해석) |
| 다단계 원자성 | 한 트랜잭션(`REQUIRED`)으로 잔액+원장+상태 | 함께 커밋/롤백 | SO의 `@Transactional` 경계 |

선불 차감을 새로 설계한다면 출발점은 단순합니다. **단건 잔액은 원자적 조건부 갱신 + 영향행수 검사**를 기본값으로 두고, **집계 한도는 잠글 수 있는 카운터 행**으로 환원하며, **사전 검증은 빠른 거절로만** 쓰고, **멱등성 키로 재시도를 흡수**합니다. 격리 수준은 이 선택들이 정해진 뒤 잠금 범위와 교착을 조정하는 다이얼로 봅니다.

## 더 깊게 볼 자료

이 문서는 선불 차감이라는 한 작업에 집중했습니다. 아래 문서들은 각 축의 일반 이론과 인접 주제를 더 깊게 다룹니다.

- [트랜잭션 고립 수준과 이상 현상](../../database/deep-dive/transactions/13-isolation-anomalies.md): 더러운 읽기·반복 불가능 읽기·팬텀 읽기·직렬화 이상을 schedule 단위로 정리합니다.
- [MVCC](../../database/mvcc.md): 스냅샷과 다중 버전이 격리 수준을 어떻게 구현하는지 내부 구조를 봅니다.
- [전파·격리·락·데드락](../../interviews/database-deep-dive/08-isolation-lock-deadlock.md): 행 락, 갭 락, next-key lock, 교착의 일반 메커니즘을 봅니다.
- [애플리케이션 경계·멱등성·금액·아웃박스](../../interviews/database-deep-dive/12-application-boundaries-idempotency-money-outbox.md): 멱등성 상태 기계, 원장/잔액 모델, 외부 부작용과 아웃박스의 전체 그림을 봅니다.
- [Spring Transactional](../../jvm/spring/spring_transactional.md): 전파의 물리/논리 트랜잭션, 커밋 시점, 프록시·self-invocation, MyBatis 연동을 봅니다.
- [MySQL Reference Manual - InnoDB Transaction Isolation Levels](https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-isolation-levels.html): 기본값 `REPEATABLE READ`, consistent read와 locking read, semi-consistent read를 원문으로 확인합니다.
- [MySQL Reference Manual - InnoDB Next-Key Locking](https://dev.mysql.com/doc/refman/8.0/en/innodb-next-key-locking.html): next-key lock이 팬텀을 막는 방식을 원문으로 확인합니다.
