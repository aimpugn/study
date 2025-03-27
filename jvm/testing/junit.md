# JUnit

- [JUnit](#junit)
    - [JUnit](#junit-1)
    - [모듈 및 역할](#모듈-및-역할)
        - [JUnit Platform](#junit-platform)
        - [JUnit Jupiter](#junit-jupiter)
        - [JUnit Vintage](#junit-vintage)
    - [Gradle 설정 예시](#gradle-설정-예시)
    - [JUnit의 Extension](#junit의-extension)
        - [`MockitoExtension`](#mockitoextension)
        - [`SpringExtension`](#springextension)
        - [`SpringExtension`와 `MockitoExtension`를 같이 쓰는 경우?](#springextension와-mockitoextension를-같이-쓰는-경우)
    - [언제 무엇을 써야 할지?](#언제-무엇을-써야-할지)

## JUnit

JUnit은 자바 애플리케이션의 단위 테스트(unit test)를 작성하고 실행할 수 있는 환경을 제공합니다.

JUnit 5는 'JUnit Platform', 'JUnit Jupiter', 'JUnit Vintage'라는 세 가지 모듈로 나누어져 있습니다.
- Platform: 테스트 실행 환경과 테스트 관리 엔진을 제공하는 기반 플랫폼
- Jupiter: 실제 테스트를 작성할 수 있는 API와 확장 메커니즘 제공
- Vintage: 기존 JUnit 3, 4 테스트 코드 실행을 지원 (하위호환성 유지)

## 모듈 및 역할

```sh
Test Code (@Test annotations)
      │
      ▼
JUnit Platform Launcher ▶ (테스트 발견 및 실행 관리)
       └─> TestEngine (JUnit Jupiter TestEngine)
             │
             └─> 테스트 코드 실행
                   │
                   └─ 결과 수집
JUnit 플랫폼 → 결과 보고 (콘솔, IDE, HTML 리포트 등)
```

### JUnit Platform

JUnit Platform은 JUnit 생태계의 기반으로, 테스트를 실행하기 위한 기반이 되는 런처(Launcher)와 테스트 결과를 처리하는 테스트 엔진(TestEngine)으로 구성됩니다.
- 테스트를 발견(discover)하고 실행(execute) 하는 역할을 합니다.
- 테스트 프레임워크를 위한 플랫폼 API를 제공합니다.
- 다른 IDE(IntelliJ, Eclipse 등) 또는 빌드 도구(Gradle, Maven)가 JUnit과 통신할 수 있게 해줍니다.
- 다양한 테스트 엔진을 플러그인 형태로 지원하여, JUnit뿐 아니라 다른 프레임워크도 구동할 수 있습니다.

### JUnit Jupiter

JUnit Jupiter는 실제로 개발자가 사용하여 테스트를 작성하는 API와 확장 기능을 제공하는 모듈입니다.

테스트 작성 API(`@Test`, `@BeforeEach`, `@AfterEach`)와 같은 테스트 메서드와 생명주기 애노테이션을 제공합니다.
그리고 확장 API(`@ExtendWith`)와 확장 포인트를 통해 JUnit의 기능을 손쉽게 확장 가능합니다.

```java
import org.junit.jupiter.api.Test;

public class SimpleTest {
    @Test
    void simpleTest() {
        assertEquals(4, 2 + 2);
    }
}
```

### JUnit Vintage

JUnit 5의 플랫폼에서 JUnit Jupiter 이전 버전(JUnit 4 이하)의 테스트를 지원하는 모듈입니다.

## Gradle 설정 예시

JUnit 5를 사용하기 위한 의존성 설정은 보통 다음과 같이 이루어집니다.

```gradle
dependencies {
    // JUnit Jupiter와 Platform 관련 모듈들의 버전을 통합 관리 (권장 방식)
    testImplementation platform("org.junit:junit-bom:5.12.0")
    // 테스트 코드를 작성할 때 필요한 API와 확장 기능 제공
    testImplementation("org.junit.jupiter:junit-jupiter")
    // DE 및 툴에서 JUnit 테스트를 발견하고 실행할 때 필요한 런처 모듈
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}
```

## JUnit의 Extension

JUnit Extension은 'JUnit 테스트 라이프사이클의 각 단계를 확장'할 수 있도록 하는 메커니즘입니다.
JUnit 5의 라이프사이클은 다음과 같습니다.
1. 테스트 클래스 생성
2. 의존성 주입
3. 테스트 메서드 실행
4. 테스트 결과 평가

JUnit Extension은 이 라이프사이클 각 단계에서 '사용자 정의된 행동을 추가할 수 있는 인터페이스'를 제공하며, 주로 다음과 같은 상황에서 사용됩니다.
- 테스트 클래스나 테스트 메서드가 실행되기 전에 어떤 특정한 작업을 수행하고자 할 때
- 테스트에 필요한 외부 자원을 자동으로 초기화하거나 종료할 때
- 의존성을 자동으로 주입하거나, 테스트 환경을 설정할 때

JUnit 5의 확장은 다음과 같은 인터페이스들을 구현합니다.
- `TestInstancePostProcessor` (테스트 클래스 인스턴스 생성 후 초기화 작업 수행)
- `BeforeAllCallback` / `AfterAllCallback` (클래스 단위 초기화)
- `BeforeEachCallback` / `AfterEachCallback` (각 테스트 메서드 전후 작업)
- `ParameterResolver` (메서드의 매개변수 자동 주입)

### `MockitoExtension`

`MockitoExtension`은 Mockito가 JUnit 5와 통합하기 위해 제공하는 확장입니다.
Mockito의 Mock 객체를 테스트 클래스에 주입해 줍니다.

외부 시스템, DB, Spring 없이 Mock으로만 테스트할 때 사용하면 좋습니다.
테스트의 속도가 매우 중요할 때, 그리고 독립적인 로직의 정확성만을 테스트할 때 유용합니다.

```java
// MockitoExtension이 테스트 클래스 내부를 리플렉션으로 분석합니다.
@ExtendWith(MockitoExtension.class)
class MyServiceTest {

    @Mock
    private SomeDependency dependency;

    // `@InjectMocks`가 붙은 객체(MyService)는 직접 생성자를 호출하거나 필드에 mock(이 경우 dependency)을 삽입합니다.
    // 이 과정에서 Spring은 전혀 관여하지 않습니다.
    @InjectMocks
    private MyService service;

    @Test
    void someTest() {
        when(dependency.method()).thenReturn("someValue");
        assertEquals("someValue", service.methodUnderTest());
    }
}
```

내부적으로 다음과 같이 동작합니다.

`MockitoExtension`은 `TestInstancePostProcessor`를 통해 테스트 클래스가 생성된 직후 필드에 존재하는 `@Mock` 어노테이션을 처리하여 *Mock 객체를 자동으로 생성*합니다.
그 다음 `@InjectMocks`가 붙은 필드를 확인하고, 생성된 Mock 객체들을 해당 필드의 생성자나 필드에 주입합니다.

그리고 `ParameterResolver`를 통해 테스트 메서드의 매개변수로 Mock 객체를 직접 주입할 수 있습니다.

```java
@Test
void testMethod(@Mock SomeDependency dependency) {
    when(dependency.method()).thenReturn("someValue");
    // ...
}
```

테스트 메서드끼리의 격리성을 보장하기 위해, `BeforeEachCallback`와 `AfterEachCallback`을 사용하여 테스트 메서드가 끝난 이후 Mock 객체의 상태를 초기화하는 등의 작업을 수행할 수 있습니다.

### `SpringExtension`

`SpringExtension`은 Spring 프레임워크가 JUnit 5와 통합하기 위해 제공하는 확장입니다.
Spring의 *테스트 애플리케이션 컨텍스트*를 로딩하고 유지합니다.
그리고 테스트 클래스 내에서 자동으로 Spring 빈을 주입할 수 있게 해줍니다.

따라서 Spring 컨텍스트나 빈의 실제 연동을 포함한 통합 테스트를 할 때 사용합니다.
실제 빈과 설정값이 올바르게 로딩되는지, 컨트롤러, 서비스, 데이터 액세스 계층의 통합적인 검증을 수행할 때 유용합니다.

```java
@ExtendWith(SpringExtension.class)
@ContextConfiguration(classes = {MyConfig.class})
class MySpringTest {

    @Autowired
    private MyService myService;

    @Test
    void test() {
        assertNotNull(myService);
    }
}
```

`SpringExtension`은 내부적으로 다음과 같이 동작합니다.

`BeforeAllCallback`을 통해 테스트 클래스 전체에 한 번, 'Spring ApplicationContext를 생성하고 로딩'합니다.
이 `ApplicationContext`는 한번 생성되면 *캐싱*되어, 여러 테스트 클래스나 메서드에서 공유되므로 빠르게 동작합니다.

그리고 `TestInstancePostProcessor`를 통해 Spring이 관리하는 빈을 테스트 클래스의 필드에 주입합니다. (`@Autowired`, `@MockBean` 등 처리)

`MockitoExtension`과 마찬가지로 `ParameterResolver`를 통해 테스트 메서드 매개변수에 Spring 빈을 직접 주입할 수 있습니다.

### `SpringExtension`와 `MockitoExtension`를 같이 쓰는 경우?

일반적으로 `SpringExtension`이 `MockitoExtension`의 역할(Mock 객체 생성과 주입)을 상당부분 흡수한다고 합니다.
그래서 `SpringExtension`과 `@MockBean`을 쓰면 Mockito를 Spring 내에서 자연스럽게 사용할 수 있다고 합니다.

```java
// SpringExtension이 `@SpringBootTest`에 의해 `ApplicationContext`를 구성합니다.
@ExtendWith(SpringExtension.class)
@SpringBootTest
class IntegrationTestExample {

    // `MyService`는 원래 `ExternalService`라는 빈을 의존합니다.
    // 그런데 `@MockBean` 어노테이션이 붙으면, 스프링 컨텍스트 레벨에서 빈 정의 자체를 가짜로 바꿔버립니다.
    // 컨텍스트 초기화 시 다음 작업을 수행됩니다.
    // 1. 지정한 타입의 기존 `ExternalService` 빈을 `ApplicationContext`에서 찾아서 제거합니다.
    // 2. `Mockito.mock(ExternalService.class)`로 mock 객체를 생성한 뒤, 같은 타입의 새로운 빈으로 등록합니다.
    //
    // 따라서 `MyService`의 `@Autowired` 필드는 실제 빈이 아닌 `mockService`를 주입받게 됩니다.
    @MockBean
    private ExternalService mockService;

    // `@InjectMocks`는 생성자, 필드, 세터 등을 사용하여 Java 객체 내부에 mock을 주입합니다.
    // 반면 `@MockBean`은 "Spring ApplicationContext에 등록된 실제 빈을 mock으로 대체"합니다.
    // 그래서 `MyService` 서비스 컴포넌트 생성 시 `Mockito.mock(ExternalService.class)`로 만든 mock이 주입됩니다.
    @Autowired
    private MyService myService;

    @Test
    void integrationTestWithMock() {
        // 실제 테스트 시 `MyService`가 내부적으로 사용하는 `mockService`의 결과가 바뀝니다.
        when(mockService.call()).thenReturn("mock");
        assertEquals("mock", myService.callExternal());
    }
}
```

## 언제 무엇을 써야 할지?

- 유닛 테스트 (POJO): `@InjectMocks` + `@Mock`

    => 빠르고 간단, Spring 컨텍스트 불필요합니다.

- Spring 빈 연동 테스트:  `@MockBean` + `@Autowired`

   => 실제 컨텍스트에서 동작, 의존성 전체를 검증합니다.

- Controller, Service 계층 통합 검증: `@SpringBootTest`, `@MockBean`

    => 전체 구성 속에서 특정 의존성만 모의화합니다.

- 전체 시스템 테스트 (E2E): `@SpringBootTest`, 실제 DB

    => 외부 시스템 제외한 진짜 애플리케이션 흐름 검증합니다.
