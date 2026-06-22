# 규칙 뒤의 이름들 — Beck 구조둔감 · 거울 테스트 · Dodds 트로피

우리 테스트 규칙은 새로 지어낸 게 아니라 널리 받아들여진 이론에서 빌려 왔다. 그런데 "어느 규칙이 누구에게서 왔나"를 잘못 적으면, 그 규칙이 빌려 온 권위까지 같이 샌다. 그래서 이전 맥락을 끊고 원전을 직접 대조하는 *cold 리뷰*를 한 번 돌렸고, 거기서 걸린 세 가지를 **출처·실제 의미·지키는 경우·지키지 않는 경우**로 정리한다.

리뷰가 잡아낸 세 가지는 이렇다.

- 가장 의심했던 **Beck의 "구조 비민감"이 오히려 정확**했다 — Test Desiderata에 실제로 그 이름으로 올라 있는 성질이다.
- 덜 의심했던 **"거울 테스트 = Meszaros"가 오귀속**이었다 — 그 이름은 Jason Rudolph가 2008년에 직접 붙였다.
- **Dodds 트로피를 "통합/E2E에 무게"라 적은 건 부정확**했다 — 무게는 *통합*에 실리고, E2E는 오히려 일부러 *적게* 둔다.

세 건 모두 의심한 방향이 빗나갔다. 직접 읽기 전엔 어디가 틀렸는지조차 거꾸로 짚고 있었던 셈이다.

## 목차

- [1. Beck — Test Desiderata, 특히 #8 구조 비민감](#1-beck--test-desiderata-특히-8-구조-비민감)
- [2. 거울 테스트 — Meszaros가 아니라 Jason Rudolph(2008)](#2-거울-테스트--meszaros가-아니라-jason-rudolph2008)
- [3. Dodds — Testing Trophy](#3-dodds--testing-trophy)
- [4. 우리 규칙과의 매핑 (정정 반영)](#4-우리-규칙과의-매핑-정정-반영)
- [5. 스스로 확인하기](#5-스스로-확인하기)
- [출처 (1차)](#출처-1차)
- [관련 노트](#관련-노트)

---

## 1. Beck — Test Desiderata, 특히 #8 구조 비민감

### 무엇에 대한 내용인가

Kent Beck이 2019년에 정리한, **좋은 테스트가 가지면 하는 12가지 성질**의 목록이다("desiderata" = 바람직한 것들). 어느 하나가 절대 규칙이 아니라, 서로 맞바꿀 수 있는 *축*들이라는 게 핵심이다.

| # | 성질 | 한 줄 뜻 |
|---|---|---|
| 1 | Isolated | 실행 순서가 달라도 같은 결과 |
| 2 | Composable | 1개를 돌리든 100만 개를 돌리든 같은 결과 |
| 3 | Fast | 빨리 돈다 |
| 4 | Inspiring | 통과하면 자신감이 생긴다 |
| 5 | Writable | 대상 코드 비용 대비 싸게 쓸 수 있다 |
| 6 | Readable | 왜 이 테스트를 쓰는지 읽힌다 |
| **7** | **Behavioral** | **대상의 *동작*이 바뀌면 반응한다** |
| **8** | **Structure-insensitive** | **대상의 *구조*만 바뀌면 결과가 변하지 않는다** |
| 9 | Automated | 사람 개입 없이 돈다 |
| 10 | Specific | 실패하면 원인이 분명하다 |
| 11 | Deterministic | 아무것도 안 바뀌면 결과도 안 바뀐다 |
| 12 | Predictive | 다 통과하면 운영에 내보내도 된다 |

### 실제 의미

#7과 #8은 한 쌍이고, Beck이 한 문장으로 묶었다:

> "Tests should be **coupled to the behavior** of code and **decoupled from the structure** of code."
> (테스트는 코드의 *동작에는* 결합하고 *구조와는* 분리돼야 한다.)

이 한 문장이 "동작을 박제하고 구현은 풀어 둔다"(→ [`./pin_behavior_not_implementation.md`](./pin_behavior_not_implementation.md))의 **이론적 뼈대**다. #7이 없으면 안전망이 회귀를 놓치고(둔감), #8이 없으면 정당한 리팩토링에 거짓 경보가 뜬다(과민).

### 지키는 경우 / 지키지 않는 경우

| | 지킴 (구조 비민감 + 동작 민감) | 어김 (구조 민감) |
|---|---|---|
| 무엇을 단언 | 반환·예외·저장 상태·나간 메시지 | 호출 순서·횟수·중간 커맨드 필드 |
| 리팩토링(동작 보존) 시 | green 유지 | red (거짓 경보) |
| 동작이 진짜 바뀌면 | red | red (구분 불가) |
| KCF 예 | `Kfsfco04u0Test` — 조립된 전문 출력을 박제, 빌더 내부 순서는 안 봄 | `verify(inOrder)`, 빌더에 넘긴 25필드 1:1 단언 |

### 놓치기 쉬운 단서 — 성질들은 서로 맞바꾼다

Beck의 단서:

> "No property should be given up without receiving a property of greater value in return."
> (어떤 성질도 더 큰 가치를 돌려받지 않고는 포기하지 마라.)

구조 비민감(#8)도 공짜가 아니다. 완전히 구조 비민감한 테스트는 보통 더 위 계층(통합·E2E)인데, 그건 #3 Fast와 #10 Specific(실패 위치가 또렷한가)을 깎는다 — E2E가 빨개지면 "어디가" 깨졌는지 단번에 안 보인다. 그래서 한 계층에 몰지 않고 섞는다. 이 맞교환이 곧 [3절 Dodds 트로피](#3-dodds--testing-trophy)의 동기다.

---

## 2. 거울 테스트 — Meszaros가 아니라 Jason Rudolph(2008)

### 출처 정정

"거울 테스트(ugly mirror)"라는 이름은 **Jason Rudolph가 2008년 블로그 글에서 직접 붙였다**(본인 명명, 다른 사람 귀속 없음). Gerard Meszaros의 《xUnit Test Patterns》가 명명한 smell이 *아니다*. 우리 정본 표(§13)가 이걸 Meszaros로 적었던 걸 정정했다. (Meszaros와의 관계는 아래.)

### 무엇에 대한 내용인가 / 실제 의미

단언이 *기대값*을 적는 게 아니라 **구현의 로직을 그대로 복제**하는 안티패턴이다. Rudolph의 예(Ruby):

```ruby
User = Struct.new(:first_name, :last_name, :email) do
  def to_s
    "#{last_name}, #{first_name} <#{email}>"
  end
end

# ❌ 거울: 단언이 구현 식을 그대로 비춘다 — 절대 어긋날 수 없다
assert_equal "#{user.last_name}, #{user.first_name} <#{user.email}>", user.to_s

# ✅ 리터럴: 요구사항을 그대로 말한다
assert_equal "Smith, John <jsmith@example.com>", user.to_s
```

왜 나쁜가 — Rudolph의 말:

> "we're forced to **mentally reverse engineer the code** in order to see through to the underlying requirement."
> (요구사항을 보려고 코드를 머릿속에서 거꾸로 풀어야 한다.)

거울 단언은 프로덕션과 *같은 식*을 쓰므로 **둘이 동시에 틀려도 통과**한다. 버그 검출력이 0이고, 구현이 바뀌면 테스트도 같이 바꿔야 하는 유지비만 남는다. 처방은 한 줄이다: **리터럴로 단언하라.**

> "Any time we can assert on a literal value, our test will be better off because of it."

### 지키는 경우 / 지키지 않는 경우 (KCF)

| | 지킴 | 어김(거울) |
|---|---|---|
| 기대값의 출처 | **실행해서 관측한 리터럴** (골든마스터) 또는 스펙 | 프로덕션 코드에서 베낀 식·상수 |
| 예 | `assertThat(sent).contains("000000654321")` — 코드를 돌려서 본 채번 결과를 리터럴로 박제 | 프로덕션이 `setCocnRcvgInqTlgSno("900000000008")`인데 테스트가 같은 `"900000000008"`을 프로덕션에서 복사해 단언 |
| 결과 | 동작이 바뀌면 잡힌다 | 프로덕션·테스트가 같은 출처라 영영 안 어긋남 |

핵심 판별: **기대값이 "독립된 진실"(관측값/스펙)에서 왔나, 아니면 대상 코드 자신에서 왔나.** 후자면 거울이다.

> 미묘한 차이 하나: Rudolph 원전은 *식(로직) 복제*를, 우리 정본 §12는 *상수 값 복제*를 "거울"이라 부른다. 둘은 같은 병의 두 증상이다 — 테스트가 자기 기대값을 *대상 코드에서* 끌어와 독립적 오라클이 없어진다는 점에서. 우리는 Rudolph의 이름을 빌려 둘 다를 가리킨다.

### Meszaros와의 관계 (왜 헷갈렸나)

Meszaros가 명명한 인접 개념은 **"Fragile Test"**(취약 테스트, behavior smell)와 그 원인 **"Overspecified Software"**(테스트에 구현 세부를 인코딩해 SUT를 과명세). 거울 테스트는 이 가족의 *한 구체 형태*라서 자연스레 Meszaros로 착각하기 쉽다 — 하지만 "mirror"라는 이름표는 Rudolph 것이다(xunitpatterns의 Fragile Test 페이지에 "mirror"라는 말은 없다). 즉 **개념의 뿌리는 Meszaros(과명세/취약), 이름은 Rudolph(거울)**.

---

## 3. Dodds — Testing Trophy

### 무엇에 대한 내용인가

Kent C. Dodds(2018)가 제시한 **테스트 분포 모델**. "단위 테스트를 가장 많이"라는 전통적 *피라미드*에 대한 대안으로, 트로피 모양으로 비중을 다시 그린다.

| 층 (아래→위) | 크기 | 무엇 |
|---|---|---|
| **Static** | 바닥(작음) | 린터·타입체커 — 컴파일 전에 잡는 것 |
| **Unit** | 작음 | 의존 없거나 협력자 mock한 개별 함수·클래스 |
| **Integration** | **가장 큼** | 여러 unit이 함께 도는 것 (무게 중심) |
| **End-to-End** | 꼭대기(중간) | mock 거의 없이 실제 사용처럼 관통 |

슬로건: **"Write tests. Not too many. Mostly integration."**
지도 원리: **"The more your tests resemble the way your software is used, the more confidence they can give you."** (사용 방식에 가까울수록 신뢰가 커진다.)

### 실제 의미

축은 **신뢰도(사용 방식에 얼마나 가까운가) 대비 비용(속도·취약성)**이다. ROI가 *통합*에서 최대다:

- **E2E**: 신뢰도 최고, 그러나 느리고 플래키하고 비싸다 → 의도적으로 *최소*.
- **Integration**: 실제에 충분히 가까우면서 E2E보다 빠르고 안정적 → *주력*.
- **Unit**: 빠르지만, 협력자를 다 mock한 *솔리터리* 유닛은 가짜 출력을 검증해 신뢰도가 낮다.

> 정정 포인트: 우리 §13은 "통합/E2e에 무게"였지만, 트로피의 무게는 **통합**에 있고 E2E는 *적게* 둔다. 또 트로피는 원래 **프런트엔드(JS)** 맥락 개념이고, 우리는 **I/O 중심 서버(펌뱅킹)** 에 번안해 적용한다("실제 DB+내부 실제, 외부기관만 경계 대역").

### 지키는 경우 / 지키지 않는 경우

| | 지킴 (통합 중심) | 어김 |
|---|---|---|
| 분포 | Integration 주력 + sociable Unit + E2E 소수 | (a) 솔리터리 유닛 피라미드: 다 mock → 가짜 출력, 신뢰도↓ / (b) 아이스크림콘: 대부분 E2E → 느리고 플래키 |
| KCF 예 | `bo.service()` 직접호출 + 실제 DB(Integration), `so`는 mockMvc로 HTTP 관통(E2E), 순수 빌더는 실제로(sociable Unit) | 협력자를 전부 mock하고 중간 커맨드만 단언하는 "고립 유닛" 더미 |

KCF의 계층 규칙(→ [`./testing_types.md`](./testing_types.md))이 트로피와 맞는 이유: 우리는 솔리터리 유닛을 양산하지 않고, **실제 DB를 관통하는 통합/E2E**에 무게를 둔다.

---

## 4. 우리 규칙과의 매핑 (정정 반영)

| 우리 규칙 | 이름(이론) | 출처 |
|---|---|---|
| 현재 동작을 박제, 버그도 일단 박제 | 특성화(characterization) 테스트 | Feathers, 《WELC》 → [`./characterization_testing.md`](./characterization_testing.md) |
| 동작만 박제·구현은 풀어 둠 / 리팩토링에 안 깨짐 | 리팩토링 정의 + **구조 비민감 #8** | Fowler 《Refactoring》 + **Beck, Test Desiderata** |
| 협력자 다 mock 말고 실제로 (sociable) | classicist / Detroit | Fowler 「Mocks Aren't Stubs」 → [`./pin_behavior_not_implementation.md`](./pin_behavior_not_implementation.md) |
| 중간 커맨드 필드 단언 금지 | Overspecified / Fragile Test | Meszaros, 《xUnit Test Patterns》 |
| **거울 단언 금지** | **거울 테스트(ugly mirror)** | **Jason Rudolph(2008) — Meszaros 아님** |
| 통합/E2E(실제 DB)에 무게 | **통합 무게·E2E 최소** Testing Trophy | Dodds(프런트엔드發 → 서버 번안) |
| 순서·횟수는 계약일 때만 | mockist 상호작용 과검증 비판 | Fowler 「Mocks Aren't Stubs」 |

굵게 표시한 두 줄이 cold 리뷰로 정정된 곳이다(거울=Rudolph, 트로피=통합 무게).

---

## 5. 스스로 확인하기

- #7 Behavioral과 #8 Structure-insensitive를 한 문장으로 묶으면? 그 문장이 "무엇을 박제하나"에 어떻게 답하나?
- 내 테스트의 기대값은 *관측/스펙*에서 왔나, 아니면 *대상 코드*에서 베껴 왔나? 후자면 무슨 안티패턴이고 왜 버그를 못 잡나?
- 골든마스터가 리터럴을 박제하는 건 거울 테스트와 같은가 다른가? (둘 다 리터럴인데 왜 하나는 좋고 하나는 나쁜가?)
- 트로피에서 무게가 가장 큰 층은? E2E를 가장 많이 쓰면 어떤 모양이 되고 뭐가 문제인가?
- "구조 비민감"을 극단으로 밀면 어떤 성질(들)을 잃나? Beck의 맞교환 원칙은 그걸 뭐라 하나?

## 출처 (1차)

- Kent Beck — [Test Desiderata](https://medium.com/@kentbeck_7670/test-desiderata-94150638a4b3) (12개 성질, #7/#8)
- Martin Fowler — [Mocks Aren't Stubs](https://martinfowler.com/articles/mocksArentStubs.html) (classicist vs mockist), 그리고 《Refactoring》(리팩토링 정의)
- Jason Rudolph — [Testing anti-patterns: The ugly mirror](https://jasonrudolph.com/blog/2008/07/30/testing-anti-patterns-the-ugly-mirror/) (2008)
- Kent C. Dodds — [The Testing Trophy](https://kentcdodds.com/blog/the-testing-trophy-and-testing-classifications)
- Gerard Meszaros — [Fragile Test](http://xunitpatterns.com/Fragile%20Test.html) / Overspecified Software (《xUnit Test Patterns》)
- Michael Feathers — 《Working Effectively with Legacy Code》 (특성화 테스트)

## 관련 노트

- 동작 vs 구현, sociable unit: [`./pin_behavior_not_implementation.md`](./pin_behavior_not_implementation.md)
- 특성화(골든마스터) 절차: [`./characterization_testing.md`](./characterization_testing.md)
- 테스트 계층(Unit/Integration/E2E): [`./testing_types.md`](./testing_types.md)
- mock 과사용 code smell: [`./testing_mock_and_codesmell.md`](./testing_mock_and_codesmell.md)
- 대역(mock/stub/fake)의 종류: [`./testing_double.md`](./testing_double.md)
