# 스프링이 로드한 빈을 Mockito로 spy/mock 해 갈아끼울 수 있는가 — 프록시·컨텍스트 캐싱의 함정

테스트에서 "대부분은 실제 빈으로 돌리되, 한 협력자만 실패(throw)하게 바꿔 그 분기를 검증"하고 싶을 때가 있다(예: DB insert가 실패해도 응답은 정상이어야 하는 swallow 분기). 직관적으로는 "그 케이스에서만 로드된 빈을 `spy`해서 그 메서드만 throw 시키면 되지 않나" 싶다. 이 노트는 **왜 그게 케이스(테스트 메서드) 단위로는 깨끗하게 안 되는지**를, 실제 MyBatis 매퍼 빈으로 실측한 PoC와 함께 정리한다.

결론을 한 줄로: **`Mockito.spy(로드된_빈)`은 예외 없이 "되긴" 하지만 *별개 인스턴스*를 만들 뿐이라, 이미 주입이 끝난 실행 그래프에는 영향을 주지 못한다.** 빈 교체는 본질적으로 컨텍스트(=테스트 클래스) 레벨이다.

## 목차

- [1. 직답](#1-직답)
- [2. 무엇을 하려 했나](#2-무엇을-하려-했나)
- [3. 재현 PoC](#3-재현-poc)
- [4. 왜 안 되는가 — 메커니즘](#4-왜-안-되는가--메커니즘)
- [5. 그래서 선택지](#5-그래서-선택지)
- [6. 일반 원칙](#6-일반-원칙)
- [7. 스스로 확인하기](#7-스스로-확인하기)
- [관련 노트](#관련-노트)

## 1. 직답

- 스프링이 DI로 이미 주입해 둔 빈을 **테스트 메서드 단위로 test double로 바꾸는** 깨끗한 방법은 없다. 빈 그래프는 컨텍스트 생성 시점에 고정된다.
- `Mockito.spy(빈)`을 호출하면 예외는 안 나지만(인터페이스 기반 프록시를 또 만든다) **새 객체**가 나온다(`sameInstance == false`). SUT가 잡고 있는 참조는 여전히 원본이라 spy를 stub해도 **SUT 동작은 안 바뀐다.**
- 빈을 갈아끼우는 정당한 지점은 **컨텍스트 구성 시점**(`@MockBean`/`@SpyBean`, 또는 `@ContextConfiguration` + `@Bean @Primary`)뿐이고, 이는 **클래스 단위**다(스프링 TestContext가 설정을 키로 컨텍스트를 캐싱하므로). 그래서 "일부 real, 일부 mock"은 보통 **별도 nested 클래스**로 갈린다.
- `@SpyBean`/`@MockBean`은 `spring-boot-test` 의존이 있어야 한다. 그게 없는 모듈(순수 `@ContextConfiguration` + `SpringExtension`)에서는 아예 못 쓴다.

## 2. 무엇을 하려 했나

`@Nested E2e`(mockMvc로 전부 실제 빈 + 실제 DB)와 `@Nested PartialE2e`(매퍼 빈 하나만 throw로 바꿔 "insert 실패 → swallow → 정상 응답"을 검증)가 별도 클래스로 있었다. 둘을 한 E2e 클래스로 합치고 "그 케이스에서만 빈을 mock"하면 PartialE2e라는 별도 클래스가 필요 없지 않을까 — 가 출발점이었다.

핵심 제약: 해피 경로 E2e는 **실제 매퍼로 실제 insert가 돌아야** 의미가 있다(그게 E2e의 충실도다). 그러니 "해피=real, 실패 케이스=throw"를 **한 클래스 안에서 케이스별로** 전환할 수 있어야 흡수가 성립한다.

## 3. 재현 PoC

대상 빈은 MyBatis 매퍼다(`@Mapper public interface FRM_DPCRQS_LST { int pi001(Map); }`). 순수 `@ContextConfiguration` 기반 테스트 베이스를 상속해 로드된 빈을 찍어 봤다.

```java
public class SpyLoadedBeanPocTest extends SfmServiceTestBase { // @ExtendWith(SpringExtension) + @ContextConfiguration

    @Autowired
    FRM_DPCRQS_LST frmDpcrqsLst; // 스프링이 로드한 매퍼 빈

    @Test
    void probe() {
        System.out.println("class       = " + frmDpcrqsLst.getClass().getName());
        System.out.println("isJdkProxy  = " + java.lang.reflect.Proxy.isProxyClass(frmDpcrqsLst.getClass()));
        var spy = Mockito.spy(frmDpcrqsLst);
        System.out.println("spy class   = " + spy.getClass().getName()
            + " sameInstance=" + (spy == frmDpcrqsLst));
        var mock = Mockito.mock(FRM_DPCRQS_LST.class);
        System.out.println("mock class  = " + mock.getClass().getName());
    }
}
```

실행 결과:

```
class       = jdk.proxy2.$Proxy73
isJdkProxy  = true
spy class   = jdk.proxy2.$Proxy73   sameInstance=false
mock class  = ...FRM_DPCRQS_LST$MockitoMock$WeTG9P3d
```

추가로 `mvn dependency:tree -Dincludes=org.springframework.boot:spring-boot-test`는 빈 결과(BUILD SUCCESS이나 트리에 없음) → 이 모듈엔 `spring-boot-test`가 없어 `@SpyBean`/`@MockBean`은 임포트조차 안 된다.

## 4. 왜 안 되는가 — 메커니즘

### 4.1 로드된 매퍼는 JDK 동적 프록시다

MyBatis `@Mapper`는 구현 클래스가 없는 인터페이스다. 스프링이 등록하는 빈은 `MapperFactoryBean`이 만든 **JDK 동적 프록시**(`MapperProxy`를 `InvocationHandler`로 하는 `$ProxyN`)다. 그래서 `getClass()`가 `jdk.proxy2.$Proxy73`, `Proxy.isProxyClass == true`로 나온다.

### 4.2 `Mockito.spy(obj)`는 "원본을 바꾸는" 게 아니라 "새 객체를 만든다"

`Mockito.spy(obj)`는 obj의 상태를 복사한 **새 spy 인스턴스**를 돌려준다. JDK 프록시처럼 인터페이스 기반이면 Mockito가 그 인터페이스로 또 다른 프록시 spy를 만들어 원본에 위임한다 — 그래서 예외 없이 `$Proxy73`이 다시 나오지만 `sameInstance == false`다.

문제의 핵심은 여기다. SUT 그래프(`서비스 → BO → 매퍼`)는 **이미 원본 빈 참조로 주입이 끝났다.** `spy()`가 만든 새 객체는 그 그래프 어디에도 안 꽂혀 있으므로, spy를 `doThrow`로 stub해도 SUT가 부르는 건 여전히 원본이다. spy의 stub은 아무 효과가 없다.

영향을 주려면 그 spy를 **살아 있는 BO 빈의 매퍼 필드에 reflection으로 다시 꽂아야** 하는데, 그 BO는 스프링이 캐싱하는 **공유 싱글톤**이라 — 다음 절 참고 — 다른 테스트로 누수되고, 끝나면 원복해야 하며, 깨지기 쉽다. (즉 가능은 해도 안티패턴이다.)

### 4.3 스프링 TestContext는 설정을 키로 컨텍스트를 캐싱한다

스프링 TestContext 프레임워크는 `@ContextConfiguration`(+ `@MockBean`/`@ActiveProfiles` 등) 조합을 **키**로 `ApplicationContext`를 만들어 **캐싱·재사용**한다. 한 키(=한 테스트 클래스 설정)에 대해 빈 그래프는 한 번 만들어지고 고정된다. 따라서 "이 메서드는 real 매퍼, 저 메서드는 throw 매퍼"를 같은 클래스 안에서 바꾸려면 메서드마다 컨텍스트를 새로 만들어야 하고(`@DirtiesContext` — 느리고 지저분), 그래서 보통은 **설정이 다른 별도 클래스**로 가른다. 빈 교체가 "클래스 레벨"인 근본 이유다.

### 4.4 깨끗한 교체 지점은 컨텍스트 구성 시점뿐

- `@MockBean`(전부 mock) / `@SpyBean`(기본 real + 케이스별 stub): 스프링 부트 테스트가 컨텍스트 만들 때 빈 정의를 갈아끼운다. **단 `spring-boot-test` 의존 필요.**
- `@ContextConfiguration(classes = MockCfg.class)` + `@Bean @Primary` mock: 부트 없이도 되는 수동 방식. 이것도 **클래스 레벨**이다.
- 둘 다 "해피=real, 실패=throw"를 한 클래스에서 케이스별로 못 바꾼다. `@SpyBean`만 "기본 real + 그 케이스만 stub"이 되는데, 그건 부트 의존이 있을 때 얘기다.

## 5. 그래서 선택지

| 안 | 방법 | 한계 |
|---|---|---|
| 별도 nested 유지 | `@ContextConfiguration` + `@Bean @Primary` mock 으로 그 클래스만 매퍼 교체 | 클래스 하나 더. 대신 실제 트랜잭션 스택까지 swallow 검증 |
| Unit으로 커버 | sociable Unit에서 그 BO를 concrete mock 으로 두고 throw → 응답 정상 단언 | 실제 트랜잭션/프록시 통합은 안 봄 |
| `@SpyBean` 도입 | `spring-boot-test` 의존 추가 후 기본 real + 케이스별 stub → 한 클래스로 흡수 | 의존 추가 + 컨텍스트 캐시 키 분리(성능) |

순수 `@ContextConfiguration` 모듈에서 의존 추가 없이 "케이스별 빈 제어"로 흡수하는 길은 사실상 없다.

## 6. 일반 원칙

- **test double로 스프링 빈을 갈아끼우는 건 "객체 연산"이 아니라 "컨텍스트 연산"이다.** 이미 주입된 객체를 사후에 `spy`/`mock`해도 그래프는 안 바뀐다. 교체는 컨텍스트가 빈을 만들기 *전에* 일어나야 한다.
- 그래서 "일부 real + 일부 double"의 경계는 자연스럽게 **테스트 클래스(=컨텍스트)** 단위로 떨어진다. 같은 클래스에서 케이스별로 바꾸려 하면 `@DirtiesContext`(비용)나 reflection 재주입(누수)으로 빠진다.
- 인터페이스 빈(JDK 프록시)은 `Mockito.mock(인터페이스)`은 깔끔하나(새 객체), `spy(로드된프록시)`는 "되는 것처럼 보이지만" 별개 객체라 함정이다. 무엇을 검증하려는지(throw만 vs 일부 real 위임)에 따라 mock/spy/`@SpyBean`을 가른다.
- 빠른 판별: **double을 케이스마다 다르게 두고 싶은가?** 그렇다면 컨텍스트를 나눠야 한다(별도 클래스). 한 종류 double로 클래스 전체가 족한가? 그러면 클래스 레벨 교체로 충분하다.

## 7. 스스로 확인하기

- 로드된 빈을 `spy`한 결과가 원본과 `==`인가? (아니라면 그 spy는 SUT에 영향을 못 준다.)
- 내가 바꾸려는 빈은 컨텍스트 생성 *전*에 교체되나, *후*에 객체를 손대고 있나?
- "일부 real, 일부 mock"을 한 클래스에서 케이스별로 하려다 `@DirtiesContext`나 reflection 재주입에 의존하고 있지 않은가? 그렇다면 클래스를 나누는 게 맞다.
- `@SpyBean`/`@MockBean`을 쓰려는데 그 모듈에 `spring-boot-test`가 실제로 있는가(`dependency:tree`로 확인)?

## 관련 노트

- [`./testing_double.md`](./testing_double.md) — test double(dummy/stub/spy/mock/fake) 분류
- [`./testing_mock.md`](./testing_mock.md) — Mockito mock/spy 사용
- [`./testing_mock_and_codesmell.md`](./testing_mock_and_codesmell.md) — 과도한 mock의 냄새
- [`./pin_behavior_not_implementation.md`](./pin_behavior_not_implementation.md) — 경계만 double로 두고 관찰 동작을 핀
