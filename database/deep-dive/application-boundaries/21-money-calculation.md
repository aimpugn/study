# Money Calculation

## decimal, rounding, allocation, minor unit

이 절은 `decimal, rounding, allocation, minor unit`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 금액을 binary floating point가 아니라 십진 정책, scale, rounding, minor unit, 배분 잔여 규칙으로 다뤄야 하는 이유를 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 BigDecimal(double)과 반올림 누적 오차 함정이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 작은 금액 배분 하나에서 시작한다. 10,000원을 세 명에게 나누면 수학적으로는 3,333.333...원이지만 실제 지급 단위가 1원인 시스템은 소수 원을 줄 수 없다. 따라서 DU45의 첫 질문은 어떤 숫자 타입을 썼는가가 아니라 총액, currency, scale, rounding mode, minor unit, 잔여 1원 배분 규칙이 하나의 계산 계약으로 고정되어 있는가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | Oracle Java SE 21 BigDecimal API: https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/math/BigDecimal.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Oracle Java SE 21 RoundingMode API: https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/math/RoundingMode.html | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `math/precision.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `math/bankers_rounding.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/payment/payment_operation.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/firmbanking/_proto.firmbanking_batch_firm.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 10,000원을 세 명에게 나누기

금액 계산의 첫 벽돌은 `0.1 + 0.2` 같은 유명한 부동소수점 예시가 아니라 실제 정산 행이다. 10,000원을 세 명에게 균등 배분하면 수학적으로는 3,333.333...원이지만 원화 minor unit은 1원이라 소수 원을 지급할 수 없다. 그래서 시스템은 '계산 결과'와 '지급 가능한 단위' 사이의 차이를 정책으로 닫아야 한다. 이 정책이 없으면 같은 총액을 계산하는 서비스마다 3,333원 세 건으로 9,999원을 만들거나, 한쪽은 반올림해 10,002원을 만들 수 있다.

```text
total = 10000 KRW, receivers = A/B/C, minor unit = 1 KRW
raw share = 3333.333333...
floor share = 3333 each -> allocated = 9999 -> remainder = 1
policy: give remainder to deterministic order by receiver id
A=3334, B=3333, C=3333 -> sum = 10000
consumer: ledger rows and payout file can reconcile total exactly
```

이 trace는 금액 계산이 숫자 타입 선택만으로 끝나지 않는다는 점을 보여 준다. `BigDecimal`을 써도 배분 정책이 없으면 잔여 1원을 누가 가져갈지 결정되지 않는다. 반대로 minor unit을 정수로 저장해도 수수료율, 세금, 환율, 할인 비율을 곱할 때 rounding mode와 scale이 없으면 결과가 흔들린다.
#### DU45-1.1 binary floating point는 십진 돈의 표현 방식이 아니다

컴퓨터의 `double`은 이진 부동소수점이다. 0.1원, 0.01달러 같은 십진 소수는 이진수로 정확히 끝나지 않는 경우가 많아 근사값으로 저장된다. 로컬 seed `math/precision.md`도 0.1과 0.2의 합이 정확히 0.3으로 비교되지 않는 예를 든다. 금융 시스템에서 문제는 화면에 0.30으로 보이느냐가 아니라, 반복 합산과 비교, 반올림 경계에서 근사 오차가 정책처럼 굳어지는 것이다.

함정은 `double`로 계산한 뒤 마지막에만 `BigDecimal`로 감싸면 안전하다고 믿는 것이다. 이미 근사값이 들어온 뒤에는 `new BigDecimal(double)`이 그 근사값을 십진으로 정밀하게 표현할 뿐 원래 의도한 0.1을 복원하지 않는다. 문자열, 정수 minor unit, 또는 `BigDecimal.valueOf`처럼 의도를 보존하는 입력 경로를 사용해야 한다.

#### DU45-1.2 BigDecimal은 값, scale, rounding policy의 조합이다

Java `BigDecimal` 공식 API는 임의 정밀도의 signed decimal이며 unscaled value와 scale로 값을 표현한다고 설명한다. 예를 들어 12345와 scale 2는 123.45다. 이 구조 덕분에 십진 금액을 정확하게 표현할 수 있지만, scale이 의미 없이 자동 통일되는 것은 아니다. add/subtract/multiply/divide마다 preferred scale이 있고, 나눗셈은 끝나지 않는 십진 결과를 만들 수 있으므로 rounding mode가 필요하다.

실무 함정은 `BigDecimal`이면 반올림 문제가 사라진다고 믿는 것이다. `1.divide(3)`은 정책 없이 정확한 십진수로 표현할 수 없어 예외가 날 수 있고, `setScale(2)`도 객체를 바꾸는 mutator가 아니라 새 값을 반환한다. 반환값을 받지 않으면 scale 조정이 적용되지 않는다.

#### DU45-1.3 rounding mode는 회계 정책이다

반올림은 수학 함수가 아니라 계약이다. Java `RoundingMode.HALF_EVEN`은 동률일 때 짝수 쪽으로 가며 반복 계산에서 누적 오류를 통계적으로 줄이는 방식으로 설명된다. 그러나 어떤 도메인은 HALF_UP, 어떤 도메인은 DOWN, 어떤 도메인은 고객 유리 방향, 어떤 도메인은 세법상 절사 규칙을 쓴다. 따라서 rounding mode는 util 내부 기본값이 아니라 상품, 세금, 수수료, 정산 계약 문서에서 내려와야 한다.

함정은 '은행가 반올림이 가장 좋다'를 모든 결제에 적용하는 것이다. 통계적 편향 감소와 법적/계약적 정산 규칙은 다른 질문이다. HALF_EVEN은 반복 계산의 평균 오차를 줄일 수 있지만, 특정 세금 계산이 원 단위 절사를 요구한다면 HALF_EVEN은 오답이다.

#### DU45-1.4 minor unit은 저장 단위와 표시 단위를 분리한다

KRW는 보통 1원이 최소 지급 단위이고, USD는 1 cent가 흔한 최소 단위다. 결제 API와 원장에서는 금액을 minor unit 정수로 저장하는 방식을 많이 쓴다. 정수 저장은 비교와 합계를 단순하게 만들지만, 입력/표시/수수료율 계산에서는 여전히 decimal policy가 필요하다. `1000`이 1000원인지 10.00달러인지 알려면 currency와 scale이 함께 있어야 한다.

운영 함정은 amount 컬럼만 보고 화폐 단위를 추론하는 것이다. 다중 통화, 세금 포함/별도, 포인트, 쿠폰, 정산 통화가 섞이면 같은 숫자 1000이 서로 다른 의미를 가진다. 금액 값은 currency, minor unit exponent, 계산 정책, 표시 locale과 함께 다뤄야 한다.

#### DU45-1.5 배분은 합계 보존 불변식으로 검증한다

할인, 수수료, 세금, 포인트를 여러 line item에 배분할 때 가장 중요한 불변식은 배분 후 합계가 원래 총액과 정확히 같아야 한다는 점이다. 보통 각 항목의 raw share를 계산하고, minor unit으로 내린 뒤, 남은 잔여 단위를 deterministic order로 배분한다. order는 금액 큰 순, line id 순, 계약 우선순위처럼 재현 가능해야 한다.

함정은 각 항목을 독립 반올림한 뒤 합계를 맞추지 않는 것이다. 화면에는 항목별로 자연스러워 보여도 합계가 결제 승인 금액과 1원 다르면 정산 파일, 영수증, 회계 전표가 서로 맞지 않는다. 작은 1원 차이가 대량 거래에서는 매일 수천 건의 수동 대사 비용이 된다.

#### DU45-1.6 검증은 속성 기반과 회계 trace를 같이 둔다

금액 계산 테스트는 예시 몇 개보다 불변식이 중요하다. 모든 배분 결과의 합은 원래 총액과 같아야 하고, 각 항목은 minor unit 정수여야 하며, 음수가 허용되지 않는 도메인에서는 음수가 나오면 안 된다. rounding boundary인 x.5, 세율 곱셈, 할인액이 항목 수보다 작은 경우, 큰 금액과 작은 금액이 섞인 경우를 따로 테스트한다.

운영에서는 계산 input, policy version, raw decimal, rounded amount, remainder allocation order를 감사 로그로 남겨야 한다. 장애가 나면 최종 금액만으로는 어떤 정책이 적용됐는지 복원할 수 없다. 특히 정책 변경 전후에는 같은 주문이 어떤 버전으로 계산됐는지 추적 가능해야 한다.



### Java 계산 trace

```java
BigDecimal total = new BigDecimal("10000");
BigDecimal raw = total.divide(new BigDecimal("3"), 10, RoundingMode.DOWN);
BigDecimal floor = raw.setScale(0, RoundingMode.DOWN);
long allocated = floor.longValueExact() * 3;
long remainder = total.longValueExact() - allocated;
// A gets +1 by deterministic order: A=3334, B=3333, C=3333
```

| 계산 단계 | 값 | 정책 | 다음 소비자 |
| --- | ---: | --- | --- |
| total | 10000 | KRW minor unit | 배분 함수 |
| raw share | 3333.3333333333 | 내부 계산 scale 10 | 절사 단계 |
| floor share | 3333 | DOWN to 0 scale | 잔여 계산 |
| remainder | 1 | deterministic receiver order | 원장 row 생성 |
| final sum | 10000 | sum preservation | 정산/대사 |

이 예시는 `BigDecimal` 사용법보다 계산 계약을 보여 주기 위한 것이다. 실무 코드는 currency별 exponent, 음수 처리, 환불 배분, 세금 포함/별도, 최대/최소 수수료를 별도 정책 객체로 가져야 한다. 그래도 first brick은 같다. 원래 총액, raw 계산, minor unit 변환, 잔여 배분, 합계 검증이 순서대로 보여야 한다.
### 관측과 검증

DU45의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU45 verification checklist
1. assert allocation sum equals original total for many totals and receiver counts
2. test BigDecimal string/valueOf inputs and reject BigDecimal(double) in money constructors
3. test rounding modes at .5 boundaries for product-specific policy
4. log policy version, raw amount, rounded amount, remainder recipient, and final sum
```

DU45의 검증은 최종 숫자만 비교하지 않는다. API 입력 문자열, 내부 BigDecimal, DB 저장값, 원장 minor unit, 정산 파일 출력값을 같은 fixture에 두고 어느 경계에서 scale과 rounding이 적용됐는지 재생한다. PASS는 합계 보존과 정책 version 추적이 동시에 되는 것이고, FAIL은 1원 차이를 오차 허용으로 숨기는 것이다.
### 기억할 압축 문장

돈은 숫자 하나가 아니라 currency, scale, rounding mode, minor unit, 배분 잔여 정책, 합계 보존 불변식의 묶음이다. `BigDecimal`은 좋은 표현 도구지만 정책을 대신 정해 주지 않으므로, 계산 경계와 검증 trace를 함께 설계해야 한다.

#### DU45-S1 등장 배경: 표시 가능한 돈과 계산 가능한 돈의 간격

사람은 10,000원이나 12.34달러처럼 십진 표기를 돈으로 이해한다. 컴퓨터의 빠른 숫자 타입은 보통 이진 표현을 쓰고, 이진 표현은 많은 십진 소수를 정확히 담지 못한다. 이 간격 때문에 금융 시스템은 일반 숫자 계산과 다른 표현 정책을 발전시켰다.

이 배경을 알면 `double이 빠르다`와 `BigDecimal이 정확하다`의 단순 비교에서 벗어난다. 질문은 속도보다 어떤 표현이 계약된 금액 단위와 반올림 규칙을 잃지 않는가다.

#### DU45-S2 입력 경계에서 이미 정책이 시작된다

금액은 DB나 계산 함수 안에서만 정책이 필요한 것이 아니다. HTTP JSON, CSV 정산 파일, 관리자 화면, 외부 PG 응답을 파싱하는 순간부터 scale과 currency가 붙어야 한다. 문자열 `10.0`과 `10.00`은 같은 수치일 수 있지만 표시 정밀도나 계약 의미가 다를 수 있다.

함정은 controller에서 `double amount`로 받은 뒤 service에서 BigDecimal로 바꾸는 것이다. 이미 입력 경계에서 의도를 잃었다. 금액 DTO는 문자열, minor unit 정수, 또는 명시적 decimal parser로 들어와야 한다.

#### DU45-S3 정수 minor unit의 장점과 한계

minor unit 정수 저장은 합계, 비교, 원장 대사에 강하다. 9700원은 long 9700으로 더하고 빼면 오차가 없다. 그러나 세율 10%, 수수료 2.9%, 환율 1342.15처럼 비율 계산을 하는 순간 decimal intermediate가 다시 등장한다.

그래서 정수 저장을 택해도 decimal 계산 정책은 사라지지 않는다. 저장 단위는 정수, 계산 중간값은 BigDecimal, 최종 지급 가능값은 minor unit으로 변환하는 계층을 분리해야 한다.

#### DU45-S4 scale은 소수점 자리수가 아니라 의미다

BigDecimal의 scale은 단지 보기 좋은 자릿수가 아니다. 10.0과 10.00은 compareTo로는 같을 수 있지만 equals에서는 scale까지 볼 수 있다. DB decimal column, JSON 문자열, 회계 파일에서 scale은 통화 precision이나 계약 표현을 나타낼 수 있다.

함정은 map key나 중복 판단에 BigDecimal equals를 무심코 쓰는 것이다. 같은 금액인데 scale이 달라 다른 key로 들어갈 수 있다. 금액 타입은 비교 정책을 명시해야 한다.

#### DU45-S5 나눗셈은 반드시 정책을 요구한다

1원을 3명에게 나누거나 수수료율을 역산하면 끝나지 않는 십진수가 나온다. BigDecimal의 exact divide는 이런 경우 예외를 낼 수 있다. 이 예외는 귀찮은 런타임 문제가 아니라 정책 부재를 알려 주는 좋은 신호다.

실무에서는 이 예외를 catch해서 대충 반올림하지 말고, 어떤 scale과 rounding mode가 계약인지 정해야 한다. 세금, 수수료, 포인트, 환율은 서로 다른 정책을 가질 수 있다.

#### DU45-S6 잔여 배분은 공정성과 재현성의 문제다

할인 100원을 3개 line item에 배분하면 34,33,33처럼 잔여 1원을 누군가 받아야 한다. 이때 순서를 매번 랜덤으로 두면 재현성이 깨지고, 항상 첫 항목이면 특정 파트너가 누적 이익이나 손실을 볼 수 있다.

정책은 도메인에 맞게 정해야 한다. 금액 큰 항목 우선, 생성 순서, 판매자 id 순, 고객 유리 방향 등 여러 선택지가 있고, 선택 이유와 검증 방법을 남겨야 한다.

#### DU45-S7 환불은 원 결제 배분의 역함수가 아니다

부분 환불은 단순히 원 결제 계산을 다시 비율로 돌리면 안 될 수 있다. 원 결제 때 잔여 1원을 A가 받았다면 환불 때도 그 이력을 고려해야 합계가 맞는다. 환불은 원 거래의 allocation ledger를 참조해야 한다.

함정은 현재 line item 가격으로 재계산하는 것이다. 결제 후 가격, 세율, 쿠폰 정책이 바뀌었을 수 있다. 환불은 당시 계산 정책 version과 당시 배분 결과를 기준으로 해야 한다.

#### DU45-S8 세금 포함과 세금 별도는 다른 계산 그래프다

부가세 포함 금액에서 세액을 역산하는 것과 공급가액에 세율을 곱하는 것은 같은 그래프가 아니다. 중간 rounding 위치가 다르면 1원 차이가 난다. 로컬 math/divisor 메모처럼 divisor 방식도 정책으로 읽어야 한다.

문제는 두 방식이 모두 그럴듯하다는 점이다. 계약서, 세법, PG/ERP 요구가 어느 방식을 요구하는지 확인하고, 문서와 테스트에 같은 예시를 넣어야 한다.

#### DU45-S9 관측 가능한 계산 로그

금액 장애에서 최종 amount만 있으면 부족하다. input amount, currency, policy version, tax included flag, raw decimal, rounded decimal, remainder, allocation recipient, output minor amount가 보여야 한다.

이 로그는 개인정보가 아니어도 민감할 수 있으므로 보존 범위와 masking을 정해야 한다. 그러나 아예 없으면 1원 차이를 설명할 수 없어 정산 팀이 수동으로 엑셀을 재현하게 된다.

#### DU45-S10 최종 설계 질문

금액 타입을 설계할 때 마지막 질문은 숫자를 어떻게 저장하느냐가 아니라 누가 어떤 단위로 어떤 정책을 적용했는지 다시 설명할 수 있느냐이다. 이 질문이 닫히면 BigDecimal, long minor unit, DB decimal, JSON string을 역할에 맞게 배치할 수 있다.

그래서 DU45의 결론은 `double을 쓰지 말자`보다 넓다. 돈은 표현, 계산, 배분, 반올림, 보존, 대사의 연결된 계약이고, 각 경계가 관측 가능해야 한다.

### DU45 추가 실전 사례: 수수료, 세금, 정산 금액이 한 줄에 섞일 때

가맹점 결제 10,000원에 플랫폼 수수료 3%, 부가세 10%가 섞인다고 가정하자. `10000 * 0.03`은 300원으로 보이지만, 수수료가 부가세 포함인지 별도인지에 따라 원장 row가 달라진다. 수수료 300원이 부가세 포함이면 공급가액과 세액을 다시 나눠야 하고, 별도면 300원에 세액 30원이 추가될 수 있다. 그래서 금액 계산은 amount 하나를 반환하는 함수가 아니라 계산 그래프와 정책 version을 반환하는 함수에 가깝다.

```text
payment gross: 10000 KRW
fee policy A: fee includes VAT
  fee_total=300 -> supply=273, vat=27, merchant_net=9700
fee policy B: fee excludes VAT
  fee_supply=300, fee_vat=30, merchant_net=9670
same rate label '3%' -> different ledger and settlement result
```

이 trace는 같은 `3%`라는 말이 서로 다른 회계 결과를 만들 수 있음을 보여 준다. 운영에서 수수료 정책이 바뀌면 과거 거래를 현재 정책으로 다시 계산해서는 안 된다. 주문에는 당시 policy id, tax mode, rounding mode, allocation result를 남겨야 한다. 그래야 환불, 취소, dispute가 몇 주 뒤에 들어와도 원 거래의 계산 방식을 복원할 수 있다.

금액 검증은 예시 기반 테스트와 속성 기반 테스트를 함께 둔다. 예시 테스트는 10,000원, 9,999원, 1원, 0원, 음수 환불, 다중 통화처럼 사람이 읽을 수 있는 케이스를 고정한다. 속성 기반 테스트는 임의 총액과 항목 수에 대해 배분 합계가 원래 총액과 같고, 모든 결과가 minor unit 정수이며, 허용되지 않은 음수가 나오지 않는다는 불변식을 반복 확인한다.

시니어 실무자가 특히 경계하는 부분은 엑셀과 서버 계산의 불일치다. 정산팀은 엑셀에서 반올림 위치를 다르게 잡을 수 있고, PG 리포트는 major unit 문자열로 금액을 내려줄 수 있으며, 내부 원장은 minor unit 정수일 수 있다. 세 표면의 단위를 맞추지 않고 차이만 보면 매일 1원짜리 장애가 쌓인다. 그래서 문서와 테스트에는 서버 계산식뿐 아니라 정산 파일의 입력/출력 예시도 있어야 한다.

### DU45 마지막 검산: 1원 차이를 장애로 취급하는 이유

금액 문서에서 1원은 작은 값이 아니라 불변식이 깨졌다는 신호다. 주문 상세, 영수증, PG 승인 금액, 내부 원장, 정산 파일이 모두 같은 총액을 설명해야 하는데 한 곳이라도 1원이 다르면 어느 계산 경계가 다른 정책을 썼다는 뜻이다. 그래서 검증은 `오차 허용`이 아니라 `차이 분류`로 닫는다. 반올림 위치가 다른지, 세금 포함 여부가 다른지, minor unit 변환이 두 번 일어났는지, 환불이 원 배분 결과를 참조하지 않았는지 순서대로 확인한다.

### DU45 운영 케이스 매트릭스

DU45의 본문을 실제 장애 대응에 쓰려면 `decimal, rounding, allocation, minor unit`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| scale equality | 10.0과 10.00 비교 | 비교 정책 명시 | compareTo/equals test |
| division | 1/3 수수료 | rounding mode 필요 | ArithmeticException test |
| refund allocation | 부분 환불 | 원 배분 결과 참조 | refund sum check |
| tax inclusive | 부가세 포함 역산 | 계산 그래프 분리 | tax trace |
| currency exponent | JPY와 USD 혼합 | currency별 minor unit | currency table |
| policy version | 정책 변경 후 취소 | 원 거래 version 사용 | policy id log |
| excel mismatch | 정산팀 엑셀 차이 | 단위와 rounding 위치 대조 | csv fixture |
| negative adjustment | 환불/보정 음수 | 허용 도메인 제한 | negative case test |

### DU45 면접식 꼬리 질문과 실전 답변

DU45를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. BigDecimal을 쓰면 금액 문제가 끝나는가?

아니다. BigDecimal은 십진 표현 도구지만 scale, rounding mode, allocation remainder, currency minor unit 정책을 대신 정하지 않는다. 정책 없는 divide나 배분은 여전히 실패하거나 불일치를 만든다.

함정은 타입을 바꿨다는 이유로 회계 정책 검증을 생략하는 것이다.

#### Q2. 왜 BigDecimal(double)이 위험한가?

double에 들어간 순간 십진 의도는 이진 근사값이 된다. BigDecimal(double)은 그 근사값을 정확히 담을 뿐 원래 사람이 입력한 0.1을 복원하지 않는다. 문자열이나 minor unit 정수 입력이 안전하다.

함정은 마지막 변환만 BigDecimal이면 정확해진다고 믿는 것이다.

#### Q3. 반올림 모드는 어떻게 고르는가?

상품 계약, 세법, PG/ERP 요구, 고객 유리 정책, 통계적 편향 감소 중 어떤 목적이 우선인지에 따라 정한다. HALF_EVEN이 좋은 경우가 있지만 모든 금융 계산의 기본 답은 아니다.

함정은 은행가 반올림을 만능 규칙처럼 적용하는 것이다.

#### Q4. 배분에서 잔여 1원은 왜 중요한가?

잔여 1원을 누가 가져가는지 정하지 않으면 합계 보존이 깨지거나 실행마다 결과가 달라진다. deterministic order와 원 거래 배분 기록이 있어야 환불과 대사에서도 같은 결정을 재현한다.

함정은 1원을 오차 허용으로 넘기는 것이다. 금융에서는 작은 차이가 불변식 실패 신호다.

#### Q5. 정수 minor unit만 쓰면 충분한가?

저장과 합산에는 강하지만 비율, 세금, 환율, 수수료 계산에는 decimal intermediate가 필요하다. 따라서 저장 단위와 계산 정책을 분리해야 한다.

함정은 long amount만 있으면 currency와 scale 의미를 따로 저장하지 않아도 된다고 보는 것이다.

### DU45 마지막 close scenario: 환불이 원 결제 계산을 다시 열 때

부분 환불은 원 결제의 역계산처럼 보이지만 실제로는 원 결제 당시의 배분 결과를 참조해야 한다. 10,000원 주문에서 A 품목이 잔여 1원을 받아 3,334원으로 배분되었다면, A만 환불할 때 현재 정책으로 다시 3,333원을 만들면 원장과 영수증이 어긋난다. 결제 후 수수료율이나 세금 정책이 바뀌었다면 재계산은 더 위험하다.

```text
original allocation
  A=3334, B=3333, C=3333, policy=ALLOC-V1
partial refund A
  use original A allocation 3334, not recomputed 3333
  write refund ledger with original allocation reference
```

이 사례의 검증은 원 결제 row와 환불 row를 함께 보는 것이다. 환불 금액 합계가 원 결제의 해당 line allocation을 초과하지 않고, 환불 후 남은 금액과 환불 원장이 원래 총액을 설명해야 한다. 이 검증이 없으면 1원 차이는 고객 문의 때가 아니라 월말 정산에서 발견된다.

### DU45 최종 보강: DB DECIMAL 컬럼도 정책을 대신하지 않는다

DB에 `DECIMAL(19, 4)`를 쓰면 이진 부동소수점 오차는 피할 수 있지만, 그것만으로 금액 정책이 완성되지는 않는다. 컬럼 scale 4는 저장 가능한 자리수일 뿐, KRW 지급 단위가 1원인지, USD 표시 단위가 cent인지, 세금 계산을 어느 자리에서 반올림해야 하는지, 배분 잔여를 누가 가져가는지 알려 주지 않는다. 따라서 DB schema는 금액 정책의 일부 근거일 수 있지만 정책 전체가 아니다.

실무에서는 DB scale과 애플리케이션 scale이 어긋나는 경우가 많다. 애플리케이션은 10자리 중간 계산을 한 뒤 DB에는 4자리로 저장하고, 화면에는 2자리로 표시할 수 있다. 이 세 경계가 문서화되어 있지 않으면 어느 단계에서 값이 잘렸는지 추적할 수 없다. 검증 fixture에는 API 입력 문자열, 내부 BigDecimal, DB 저장값, 정산 파일 출력값을 함께 넣어야 한다. 그래야 `10.005`가 어느 경계에서 `10.00` 또는 `10.01`이 되었는지 다시 설명할 수 있다.

또 하나의 실전 함정은 금액과 수량을 같은 방식으로 다루는 것이다. 수량은 소수 셋째 자리까지 허용하지만 금액은 원 단위인 상품, 무게 기반 상품, 포인트와 현금이 섞인 주문에서는 각 값의 단위가 다르다. 이름이 모두 amount처럼 보여도 currency amount, quantity, point, tax base는 서로 다른 타입이어야 한다. 타입이 분리되면 잘못된 덧셈이 컴파일 단계나 테스트 단계에서 드러난다.

### DU45 아주 작은 값의 정책

0원, 1원, 최소 수수료, 무료 쿠폰 같은 작은 값은 예외가 아니라 금액 모델의 경계 테스트다. 1원 미만 수수료를 절사하면 플랫폼 수익이 사라질 수 있고, 최소 수수료를 적용하면 고객에게 예상보다 큰 비용이 보일 수 있다. 무료 쿠폰과 포인트를 같이 쓰면 현금 결제액은 0원이지만 세금 기준 금액은 남을 수 있다. 그래서 금액 정책은 큰 주문 평균값이 아니라 가장 작은 지급 단위에서도 검증되어야 한다.

이 경계 테스트의 PASS 조건은 작은 값에서도 합계 보존, 음수 방지, 세금/수수료 정책, 원장 row 의미가 모두 설명되는 것이다. FAIL 조건은 `금액이 작으니 무시`하는 분기가 생겨 정산 파일과 영수증이 서로 다른 이야기를 하는 경우다.

마지막으로 DU45의 핵심 검증 문장은 짧다. 금액 계산은 최종 숫자가 그럴듯한지가 아니라, 입력 단위와 정책 version과 반올림 위치와 잔여 배분이 다시 재생되는지로 통과한다.

### DU45 source boundary note

Java BigDecimal과 RoundingMode 문서는 십진 표현과 반올림 API의 사실 근거다. 하지만 어떤 rounding mode가 세금, 수수료, 포인트, 환불에 맞는지는 공식 Java 문서가 아니라 도메인 계약이 정한다. 따라서 이 절의 결론은 API 사용법과 회계 정책을 분리해 읽어야 한다.

### DU45 최종 보강: 금액 테스트는 예쁜 숫자가 아니라 깨지는 숫자로 시작한다

금액 계산을 검증할 때 `10000`, `20000`처럼 나누어떨어지는 값만 쓰면 가장 중요한 결함이 숨어 버린다. 테스트 fixture에는 `1`, `2`, `3`, `10.005`, `0.1`, `999999999999.99`, 서로 다른 currency, 무료 쿠폰, 부분 환불, 수수료 최소값처럼 정책 경계를 건드리는 값이 들어가야 한다. 이런 값은 보기에는 사소하지만, 실제 장애는 대부분 이 경계에서 시작한다.

예를 들어 100원을 세 명에게 나누면 34,33,33처럼 잔여 1원의 주인이 필요하다. 1원을 세 명에게 나누면 두 명은 0원을 받을 수 있는데, 이때 0원 row를 남길지, 아예 생성하지 않을지, 회계 보고서에는 어떻게 표시할지 결정해야 한다. 이 정책이 없으면 합계는 맞는데 라인별 금액이 이상하다는 문의가 생긴다. 금액 테스트는 합계 보존만이 아니라 line item별 설명 가능성까지 보아야 한다.

운영 로그도 같은 방식으로 설계한다. `amount=1000` 하나로는 부족하다. `input_amount`, `currency`, `policy_version`, `rounding_mode`, `scale`, `raw_decimal`, `rounded_minor`, `remainder_rule`, `allocation_target`을 남기면 1원 차이가 났을 때 어느 경계에서 결정됐는지 되짚을 수 있다. 이 로그가 없다면 장애 회고는 코드 추측과 엑셀 재계산으로 흐른다.

금액 정책은 시간이 지나며 바뀔 수 있다는 점도 중요하다. 오늘의 수수료 반올림 정책과 작년 주문의 환불 정책이 다르면, 환불은 현재 정책이 아니라 원 거래 당시 정책을 따라야 할 수 있다. 그래서 원장이나 결제 상세에는 policy version이 남아야 한다. 코드가 최신 정책만 알고 있으면 과거 거래를 재계산할 때 1원 차이가 반복된다.

검증 fixture도 version을 포함해야 한다. 같은 입력 금액을 `POLICY-2025`와 `POLICY-2026`으로 각각 계산해 결과가 왜 다른지 설명할 수 있어야 한다. 이렇게 해야 금액 계산은 산술 함수가 아니라 시간에 따라 versioned되는 도메인 계약이라는 사실이 문서와 테스트에 함께 남는다.

### 추가 판정 질문: 금액 타입은 계산을 금지하기도 해야 한다

좋은 금액 타입은 덧셈과 반올림을 편하게 해 주는 데서 끝나지 않는다. 서로 다른 currency를 더하지 못하게 하고, 수량과 금액을 섞지 못하게 하며, tax base와 payable amount를 같은 값처럼 다루지 못하게 해야 한다. 이런 금지는 귀찮은 제약이 아니라 장애를 앞당겨 발견하는 장치다.

테스트도 실패해야 할 조합을 포함해야 한다. KRW와 USD를 더하려 하면 실패해야 하고, tax-included amount와 tax-exclusive amount를 같은 함수에 넣으면 정책을 요구해야 하며, rounding mode 없이 divide를 호출하면 명시적으로 막혀야 한다. 금액 계산에서 좋은 API는 가능한 일을 늘리는 API가 아니라, 해서는 안 되는 일을 줄이는 API다.

이 관점은 문서 작성에도 중요하다. 금액 문서를 읽은 사람이 “BigDecimal을 쓰면 된다”에서 멈추면 아직 부족하다. 어떤 값은 더하면 안 되고, 어떤 값은 반올림 정책 없이 나누면 안 되며, 어떤 값은 원 거래의 정책 version을 따라야 한다고 말할 수 있어야 한다.

금액 계산의 마지막 실무 감각은 "언제 계산하지 않을 것인가"다. 이미 원장에 확정된 금액은 화면을 다시 그리거나 환불을 처리할 때 새 정책으로 재계산하지 않는다. 새 정책은 앞으로 들어올 거래에 적용하고, 과거 거래는 당시 입력값과 당시 정책 version과 당시 배분 결과를 읽어야 한다. 이 구분이 없으면 코드가 더 똑똑해질수록 과거 거래가 조용히 바뀐다.

따라서 금액 테이블에는 계산 결과만이 아니라 계산 근거가 필요하다. 최소한 currency, minor unit amount, policy id, rounding mode, allocation group, source transaction id가 남아야 하고, 필요하면 raw decimal과 rounded decimal을 audit table에 남긴다. 이 정보가 있으면 1원 차이가 났을 때 "현재 코드로 다시 계산하면 맞다"가 아니라 "당시에는 어떤 정책으로 이렇게 계산됐다"를 설명할 수 있다. 금융 시스템에서 이 설명 가능성은 성능 최적화보다 더 오래 남는 품질이다.

### DU45 운영 복구 훈련: 정산 파일과 서버 금액이 1원 다를 때

정산 파일과 내부 원장이 1원 다를 때 가장 먼저 할 일은 어느 쪽이 틀렸는지 단정하는 것이 아니다. 먼저 같은 금액을 같은 단위로 보고 있는지 확인한다. PG 리포트는 major unit 문자열을 줄 수 있고, 내부 원장은 minor unit 정수를 저장할 수 있으며, 회계 파일은 세금 포함 금액과 수수료를 분리해 줄 수 있다. 단위와 기준 시점이 다르면 1원 차이는 계산 오류가 아니라 비교 방식 오류일 수 있다.

```text
1원 차이 triage
  input amount      : customer paid 10000 KRW
  fee policy        : fee includes VAT or excludes VAT?
  rounding point    : line item, order total, payout batch 중 어디인가?
  stored ledger     : gross, fee, net이 분리되어 있는가?
  settlement report : PG가 어떤 currency unit과 tax mode로 내려줬는가?
  repair            : 원 계산 근거를 보존한 adjustment인가, 조용한 overwrite인가?
```

이 순서가 중요한 이유는 금액 차이가 대개 산술 능력 부족에서 나오지 않기 때문이다. 같은 3% 수수료도 부가세 포함인지 별도인지에 따라 원장 row가 달라지고, line item마다 반올림한 뒤 합산하는 방식과 주문 총액에서 한 번 반올림하는 방식은 다른 결과를 낸다. 따라서 금액 장애 분석에서는 `계산식`보다 `계산 그래프`를 먼저 펼쳐야 한다.

테스트 fixture는 운영 대사를 닮아야 한다. 서버 함수의 반환값 하나만 assert하면 부족하다. 입력 JSON, 내부 금액 타입, DB 원장 row, 환불 row, 정산 CSV 한 줄을 같은 케이스에 넣어야 한다. 그래야 1원 차이가 parser에서 생겼는지, rounding mode에서 생겼는지, allocation remainder에서 생겼는지, 파일 export에서 생겼는지 추적할 수 있다.

마지막으로 금액 보정은 원 거래를 덮어쓰는 방식보다 보정 사건을 추가하는 방식이 안전하다. 이미 고객에게 발행된 영수증과 정산 파일이 있다면 과거 row를 조용히 바꾸는 순간 시스템은 두 개의 과거를 갖게 된다. 보정 row는 왜 차이가 났고 누가 승인했으며 어느 정책을 적용했는지를 남긴다. 이 기록이 있으면 1원 차이도 장애가 아니라 설명 가능한 회계 사건으로 정리된다.

### DU45 마지막 운영 연습: 계산 정책을 코드 밖으로 꺼내기

금액 계산에서 정책은 코드 안의 작은 상수가 아니다. 반올림 위치, 최소 수수료, 세금 포함 여부, currency exponent, 잔여 배분 순서는 상품 계약과 회계 요구가 바뀌면 함께 바뀐다. 이 값을 코드에 흩어 놓으면 새 정책을 배포한 순간 과거 거래도 새 코드로 재계산되는 사고가 생긴다. 그래서 계산 정책은 이름과 version을 가진 도메인 객체로 다루는 편이 안전하다.

```text
money policy record
  policy_id: FEE-KR-2026-01
  currency: KRW
  minor_unit: 1
  fee_rate: 0.029
  fee_vat_mode: included
  rounding_scale: 0
  rounding_mode: HALF_UP
  remainder_rule: largest_line_then_order_id
```

이 record는 계산을 느리게 만들기 위한 장식이 아니다. 환불, 취소, chargeback, 정산 재생성에서 과거 거래가 어떤 정책으로 계산됐는지 복원하기 위한 증거다. 원 주문이 `FEE-KR-2026-01`로 계산됐다면, 2027년에 환불이 들어와도 당시 정책을 읽어야 한다. 현재 정책으로 다시 계산하면 고객에게 발행한 영수증과 내부 원장이 달라질 수 있다.

테스트는 policy version을 바꾸는 순간 더 중요해진다. 같은 10,000원 결제를 이전 정책과 새 정책으로 각각 계산해 차이를 의도된 변화로 설명할 수 있어야 한다. 차이가 설명되지 않으면 마이그레이션 전에 멈춰야 한다. 돈 계산은 "정확한 숫자 타입"의 문제가 아니라, 시간이 지나도 같은 거래를 같은 이유로 다시 설명할 수 있는 기록의 문제다.

운영에서 또 자주 놓치는 축은 표시 금액과 결제 금액, 정산 금액의 차이다. 고객 화면에는 10,000원이 보였고 PG 승인도 10,000원일 수 있지만, 가맹점 정산에는 수수료와 부가세, 보류 금액, 포인트 사용분이 반영된다. 이 값들을 모두 `amount`라는 이름으로 다루면 어느 금액이 고객에게 청구된 돈인지, 어느 금액이 가맹점에게 지급될 돈인지, 어느 금액이 세금 계산의 기준인지 흐려진다. 그래서 이름부터 `gross_amount`, `fee_amount`, `tax_amount`, `net_settlement_amount`, `point_amount`처럼 책임을 나누어야 한다.

이 이름 구분은 단순 가독성이 아니다. 정산 장애에서 "금액이 맞지 않는다"는 말은 대개 어떤 금액끼리 비교했는지가 빠진 말이다. 고객 청구액과 가맹점 정산액을 비교하면 당연히 수수료만큼 다르고, 세금 포함 수수료와 세금 별도 수수료를 비교하면 반올림 위치가 다르다. 좋은 금액 모델은 비교 가능한 값끼리만 비교하게 만들고, 비교할 때 currency, scale, policy version, 기간, source event가 함께 따라오게 한다.

마지막 실험은 작은 CSV 하나로도 가능하다. 주문 3건, 환불 1건, 포인트 사용 1건, 수수료 정책 2개를 넣고 서버 계산 결과와 정산 파일 결과를 나란히 만든다. 이 파일에서 각 row가 어떤 policy id와 rounding mode를 썼는지 보이면 운영자가 엑셀로 다시 계산해도 같은 결론에 도달한다. 반대로 최종 amount만 있으면 1원 차이를 설명하기 위해 코드를 뒤져야 한다. 금액 계산 문서의 목표는 공식을 외우게 하는 것이 아니라, 차이가 났을 때 어디서 어떤 정책이 적용됐는지 추적하는 눈을 만드는 것이다.
