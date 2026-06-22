# 무엇을 박제할 것인가 — 동작을 고정하고 구현은 풀어 둔다

characterization(골든마스터) 테스트로 레거시에 안전망을 씌울 때, "일단 강하게 다 고정해 두면 리팩토링이 안전하겠지"라는 직관이 든다. 목적은 옳다 — 변경의 사이드 이펙트가 실패로 *즉시* 드러나 리팩토러가 안심하게 하는 것. 그런데 **무엇을** 강하게 고정하느냐가 그 목적의 성패를 가른다. 이 노트는 "동작을 고정하고 구현은 풀어 둔다"가 왜 그 목적에 부합하고, "구현까지 고정"이 왜 목적을 *배신*하는지 정리한다.

[`./characterization_testing.md`](./characterization_testing.md)가 "현재 동작을 박제한다"는 절차라면, 이 노트는 그 박제의 *대상 선택*에 대한 것이다.

## 목차

- [1. 직답](#1-직답)
- [2. 리팩토링의 정의를 따라가면 답이 정해진다](#2-리팩토링의-정의를-따라가면-답이-정해진다)
- [3. 안정감의 두 종류](#3-안정감의-두-종류)
- [4. 중립 예시: 같은 시나리오, 두 가지 테스트](#4-중립-예시-같은-시나리오-두-가지-테스트)
- [5. 순서·상호작용이 곧 동작인 예외](#5-순서상호작용이-곧-동작인-예외)
- [6. 리팩토링 vs 수정](#6-리팩토링-vs-수정)
- [7. 조율 서비스: 조립된 출력을 핀하라](#7-조율-서비스-조립된-출력을-핀하라)
- [8. 테스트에 이빨이 있는지는 뮤테이션으로 잰다](#8-테스트에-이빨이-있는지는-뮤테이션으로-잰다)
- [9. 스스로 확인하기](#9-스스로-확인하기)
- [관련 노트](#관련-노트)

## 1. 직답

리팩토링 안전망은 **관찰 가능한 동작(반환값·예외·영속 상태·경계 동작)을 강하게**, **구현 세부(내부 호출 순서·협력자 상호작용 횟수·내부 객체의 필드들)는 고정하지 않게** 박제해야 한다. 좋은 안전망은 **"관찰 가능한 동작이 바뀔 때만, 그리고 바뀌면 반드시"**(if-and-only-if) 빨개진다. 구현을 박제하면 동작이 그대로인 안전한 리팩토링에도 red가 떠서, 안전망이 목적과 *반대로* 작동한다.

## 2. 리팩토링의 정의를 따라가면 답이 정해진다

리팩토링(Fowler)은 **관찰 가능한 동작을 바꾸지 않고** 내부 구조만 바꾸는 것이다. 그러니 테스트가 내부 구조(호출 순서, 어느 협력자를 부르는지, 어느 단계에서 예외가 나는지, 내부 커맨드의 필드들)를 단언하고 있으면 이런 일이 벌어진다. 리팩토러가 동작은 보존한 채 구조만 바꾼다 — 정당한 리팩토링이다. 그런데 구현을 박제한 그 테스트가 깨진다. red를 본 리팩토러는 "동작을 깬 건가, 아니면 그냥 과명세 테스트를 건드린 건가"를 **가려낼 수 없다.**

즉 구현 박제는 리팩토링의 정의와 정면으로 충돌한다. Beck의 테스트 desiderata 중 *structure-insensitive*(구조 비민감)가 이 성질을 가리킨다.

## 3. 안정감의 두 종류

"리팩토러가 안정감을 느끼는 게 중요하다"는 옳다. 그러나 안정감에는 두 종류가 있고, 구현 박제는 *가짜* 쪽을 준다.

| 고정 대상 | green의 의미 | 리팩토링(동작 보존) 시 | 리팩토러 체감 |
|---|---|---|---|
| **동작** (출력·예외·상태·경계) | 동작이 보존됐다 | red면 진짜 회귀 | **진짜 안정감** — red를 신뢰, 변경이 빨라짐 |
| **구현** (순서·상호작용·내부 필드) | 구조까지 그대로다 | 동작 같아도 red(거짓 경보) | **가짜 안정감** — red에 둔감해지거나 리팩토링 마비 |

구현을 박제하면 "사이드 이펙트가 바로 보인다"가 아니라 "**모든 변경이 빨개져서 진짜 사이드 이펙트가 잡음에 묻힌다**". 안전망의 신호 대 잡음비가 나빠진다 — 이건 [`./characterization_testing.md`](./characterization_testing.md) §7의 "과도한 핀" 실패 모드를 리팩토러의 심리 측면에서 본 것이다.

## 4. 중립 예시: 같은 시나리오, 두 가지 테스트

주문 결제를 조율하는 서비스. 내부 협력자 여러 개 + 외부 의존(결제 게이트웨이) 하나.

```java
// 검증 → 재고예약 → 결제(외부) → 주문저장 → 영수증
public Receipt checkout(Cart cart) {
    if (!validator.isValid(cart)) throw new CheckoutException("INVALID_CART");
    var reservation = stockReserver.reserve(cart.items());     // ①
    var payment     = paymentGateway.charge(cart.total());     // ② 외부
    var order       = orderRepository.save(toOrder(cart, payment)); // ③
    return new Receipt(order.id(), payment.approvalNo(), cart.total());
}
```

관찰 가능한 동작: 반환 `Receipt`, 검증 실패 시 예외, 결제 실패 시 주문 미저장.
구현 세부: ①②③의 *순서*, 각 협력자에 넘기는 내부 커맨드의 필드들, 실패가 *몇 번째 단계*에서 나는지.

### ❌ 나쁜 예 — 구현 박제(mockist)

```java
// 호출 순서를 고정 — 협력자를 합치거나 순서를 바꾸면(동작 동일) 깨진다
var inOrder = inOrder(stockReserver, paymentGateway, orderRepository);
inOrder.verify(stockReserver).reserve(any());
inOrder.verify(paymentGateway).charge(any());
inOrder.verify(orderRepository).save(any());

// 내부 커맨드의 12개 필드를 1:1 단언 — 프로덕션 상수의 거울(mirror test)
verify(stockReserver).reserve(captor.capture());
assertThat(captor.getValue().warehouseCode()).isEqualTo("WH01");
// ... 필드 11개 더 (프로덕션 상수가 바뀌면 테스트도 같이 바꿔야 함)

// "어느 단계까지 호출됐나"를 고정 — 내부 흐름에 결합
verify(stockReserver).reserve(any());
verify(orderRepository, never()).save(any());
```

깨지는 경우: ①②③ 순서 변경, 협력자 병합, 내부 커맨드 재구성 — **전부 동작은 그대로인데 red.**

### ✅ 좋은 예 — 동작 박제(상태 검증 + 경계값)

```java
// 가벼운 in-memory fake — 상호작용이 아니라 결과 "상태"를 검증(classicist)
var orders  = new InMemoryOrderRepository();
var gateway = new FakePaymentGateway().approveWith("APPROVAL-1");
var service = new CheckoutService(realValidator, new InMemoryStockReserver(stock), gateway, orders);

// 관찰 가능한 출력 + 부작용(상태)
var receipt = service.checkout(cart(item("A", 2)).total(3000));
assertThat(receipt.approvalNo()).isEqualTo("APPROVAL-1");
assertThat(orders.findById(receipt.orderId())).isPresent();

// "결제 안 됨"을 호출 횟수가 아니라 관찰 가능한 사실로
assertThatThrownBy(() -> service.checkout(emptyCart()))
    .isInstanceOf(CheckoutException.class).hasMessageContaining("INVALID_CART");
assertThat(gateway.charges()).isEmpty();

// "주문 미저장"을 결과 상태로 (never() 호출검증이 아님)
gateway.failWith("PG_DOWN");
assertThatThrownBy(() -> service.checkout(cart(item("A", 1)).total(1000)));
assertThat(orders.all()).isEmpty();
```

①②③ 순서를 바꿔도, 협력자를 합쳐도, 내부 커맨드를 재구성해도 — **동작이 같으면 통과**. 이게 리팩토링에 강한 안전망이다. (mock 상호작용 검증의 코드 냄새는 [`./testing_mock_and_codesmell.md`](./testing_mock_and_codesmell.md), 대역 종류는 [`./testing_double.md`](./testing_double.md).)

## 5. 순서·상호작용이 곧 동작인 예외

규칙은 "상호작용을 절대 검증하지 마라"가 아니다. **그 상호작용 자체가 사용자가 신경 쓰는 계약**이면 검증이 정당하다.

- "외부 청구가 성공한 *뒤에만* 주문을 확정한다", "차변을 먼저, 대변을 나중" → 순서 자체가 계약.
- "검증 실패 시 외부 청구를 *하지 않는다*" → 부작용 부재가 계약(단, 가능하면 호출검증보다 *결과 상태/부작용 없음*으로 확인).

판별 기준: 그 순서/호출이 **명세에 적힐 만한 약속**인가, 아니면 단지 *현재 구현이 그렇게 짜인 것*인가. 후자면 박제하지 않는다.

## 6. 리팩토링 vs 수정

- **리팩토링(동작 보존)**: 구조 비민감 테스트가 필요. 동작 테스트가 green을 유지해 "안전"을 보장.
- **수정(동작 변경)**: 동작 델타를 보여주는 테스트가 필요. 동작 테스트의 red가 "무엇이 바뀌는가"를 알려줌.
- **구현 박제 테스트는 둘 다에 깔끔히 기여하지 못한다** — 구조 변경과 동작 변경을 뒤섞기 때문.

## 7. 조율 서비스: 조립된 출력을 핀하라

검증 → 여러 협력자 호출 → 결과 조립 → 다운스트림 전달, 같은 *조율(orchestration)* 서비스에서는 흔히 **모든 협력자를 mock하고 "각 협력자에 넘긴 중간 커맨드의 필드들"을 일일이 단언**한다. 여기엔 함정이 둘 겹쳐 있다.

하나는 거울 테스트다. 중간 커맨드의 필드 값은 대개 프로덕션 상수를 그대로 복사한 것이라, 프로덕션을 바꾸면 테스트도 따라 바꿔야 한다. 결국 "프로덕션과 테스트가 같은 값을 쓰는지"만 확인할 뿐 버그는 못 잡는다. 다른 하나는 더 고약하다. 협력자(빌더·포매터)까지 mock하면 조립된 출력 자체가 가짜(`"PART-A" + "PART-B"` 같은)라서, 정작 *실제 조립 로직*은 한 번도 돌지 않는다. 그러니 "출력이 올바른가"라는 가장 중요한 질문을 검증하지 못한다.

더 나은 방식은 **sociable unit**이다. 값싸고 결정적인 *순수* 협력자(빌더·매퍼·포매터)는 **실제 객체로** 쓰고, 난수·시계 같은 비결정 소스와 외부 호출·DB 같은 경계만 mock한다. 그런 뒤 다운스트림으로 나간 **실제 출력**(반환값 + 경계로 나간 메시지)을 박제한다. 이러면 가짜가 아니라 *실제로 조립된 출력*에 핀이 걸린다.

"unit"을 한 클래스로 좁히면 협력자를 다 mock하게 되어 중간 커맨드 단언으로 흐른다. 반대로 unit을 *동작 단위*로 보면 순수 협력자까지 포함해 실제로 돌리고 출력만 본다(classicist/Detroit). 동작 단위로 보는 쪽이 리팩토링에 강하고, mock으론 못 잡던 **조립 회귀까지 실제로 잡는다**. 실무에서 이렇게 바꾸면 보통 테스트 수·단언이 줄고(중간 필드 수십 개 단언 제거) 출력 회귀 검출력은 오히려 올라간다.

### 7.1 바이트·고정폭 레코드 출력의 함정

출력이 바이트 레코드/고정폭 메시지(외부 전송 포맷)면, 그걸 **Java String 통짜로 박제하지 마라**.

- 멀티바이트 인코딩(비ASCII 문자가 섞인 레코드)에서 String 표현은 런타임·소스파일 인코딩에 취약하고, 로그로 보면 깨져 보여 골든값을 복사하기도 어렵다.
- 대신 **(a) 바이트 길이**, **(b) 관찰 가능한 ASCII 필드(식별자·금액·코드 등)를 `contains`/오프셋으로**, 필요하면 **(c) 바이트의 해시/hex 지문**으로 박제한다. (a)+(b)는 읽히는 매뉴얼, (c)는 임의 영역의 변경까지 잡는 최대 민감도.
- 트레이드오프: (a)+(b)만 쓰면 단언하지 않은 영역(예: 고정 filler)의 변경을 놓칠 수 있다. 최대 안전망이 필요하면 (c) 지문을 더한다. ASCII 필드 `contains`는 멀티바이트 깨짐과 무관하게 안전하다는 점이 (b)의 핵심 이점.

## 8. 테스트에 이빨이 있는지는 뮤테이션으로 잰다

테스트가 실제로 결함을 잡는지(흔히 "teeth가 있다"고 한다 — false-green이 아닌지)는 육안이 아니라 **뮤테이션 테스트**(예: PIT)로 정량화한다.

- 살아남은 뮤턴트 = 약한 테스트(검출 실패).
- 동작을 안 바꾸는 뮤턴트(동치)에 깨지는 테스트 = 과명세 신호.

대규모 표준화에서 "단언이 의미를 갖는지 육안 점검"은 비현실적이다. 뮤테이션 점수 표본이 측정 가능한 게이트가 된다.

## 9. 스스로 확인하기

- 동작을 보존하는 리팩토링(메서드 추출·협력자 병합·호출 순서 변경)을 했다. 내 테스트는 *살아남는가*? 깨진다면 그 테스트는 동작을 박제했나 구현을 박제했나?
- 조율 서비스에서 협력 빌더를 mock하고 그 입력 커맨드의 필드를 단언하고 있다. 이 테스트는 *조립된 실제 출력*을 검증하나, 아니면 가짜 출력 + 프로덕션 상수의 거울인가? 순수 빌더를 실제로 쓰면 무엇이 달라지나?
- 출력이 멀티바이트 고정폭 레코드다. 이걸 String 통짜로 박제할 때의 위험은? 무엇으로 대신 박제하나?
- "결제 실패 시 주문 미저장"을 `verify(repo, never()).save()`로 쓴 것과 `assertThat(orders.all()).isEmpty()`로 쓴 것의 차이는? 어느 쪽이 리팩토링에 강한가?
- 내가 단언한 호출 순서는 "명세에 적힐 약속"인가 "현재 구현의 배선"인가?
- green인데 사실은 아무 결함도 못 잡는 테스트를, 코드를 안 보고 어떻게 식별하나?
- "강하게 고정하면 안전하다"에서 *무엇을* 강하게 고정해야 그 안전이 진짜가 되나?

## 관련 노트

- 이 규칙들의 이론 출처(Beck 구조둔감·거울 테스트·Dodds 트로피): [`./testing_theory_lineage.md`](./testing_theory_lineage.md)
- 박제의 절차와 SUT 경계(상속 금지): [`./characterization_testing.md`](./characterization_testing.md)
- 대역(mock/stub/fake)의 종류: [`./testing_double.md`](./testing_double.md)
- mock 과사용의 코드 냄새: [`./testing_mock_and_codesmell.md`](./testing_mock_and_codesmell.md)
- 테스트 종류와 계층: [`./testing_types.md`](./testing_types.md)
