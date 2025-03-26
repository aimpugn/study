# Spring Transactional

- [Spring Transactional](#spring-transactional)
    - [테스트에서의 Transactional](#테스트에서의-transactional)
        - [테스트 클래스 초기화](#테스트-클래스-초기화)
        - [트랜잭션 적용 판단](#트랜잭션-적용-판단)
        - [`MockMvc` 준비](#mockmvc-준비)
        - [`MockMvc` 통한 요청 수행: `DispatcherServlet` 내부 흐름](#mockmvc-통한-요청-수행-dispatcherservlet-내부-흐름)
        - [응답 처리](#응답-처리)
        - [테스트 종료](#테스트-종료)

## 테스트에서의 Transactional

### 테스트 클래스 초기화

`@ExtendWith(SpringExtension.class)` 또는 `@RunWith(SpringRunner.class)`에 의해 JUnit이 Spring TestContext Framework를 로드합니다.

`@RunWith(SpringRunner.class)`는 JUnit 4 기반의 테스트에서 Spring Context를 통합하기 위해 등장했다고 합니다.
- `@RunWith`는 JUnit 4의 런너(Runner)를 확장할 수 있습니다.
- [`@RunWith(SpringRunner.class)`](https://docs.spring.io/spring-framework/reference/testing/testcontext-framework/support-classes.html#testcontext-junit4-runner)는 사용자 정의 스프링 `Runner`를 실행하도록 구성합니다.

하지만 JUnit 5에서 확장 모델을 사용하여 Spring 기능을 통합하기 위해[`@ExtendWith(SpringExtension.class)`](https://docs.spring.io/spring-framework/reference/testing/testcontext-framework/support-classes.html#testcontext-junit-jupiter-extension)가 등장했습니다. 스프링 외에 Mockito(`MockitoExtension.java`), 문서화(`RestDocumentationExtension.java`) 등의 확장도 사용할 수 있습니다.
- `SpringExtension`은 `BeforeEachCallback`, `AfterEachCallback`, `TestInstancePostProcessor`, `ParameterResolver` 등 여러 인터페이스를 구현합니다. 이 확장 포인트를 통해 Spring의 `TestContextManager`를 연동하여 `ApplicationContext` 생성, 의존성 주입, 트랜잭션 설정 등을 처리합니다.
- `@ParameterizedTest`와 함께 사용할 수 있습니다.

`@ContextConfiguration`, `@WebAppConfiguration`에 따라 `ApplicationContext`이 생성됩니다.

### 트랜잭션 적용 판단

`@Transactional`이 테스트 클래스나 메서드에 붙어 있으면, `TransactionalTestExecutionListener`가 동작합니다.
이 경우 테스트 시작 시 `PlatformTransactionManager.getTransaction()`으로 트랜잭션을 시작합니다.
그리고 테스트 종료 시 `rollback()` 수행합니다. 다만, `MockMvc` 내부의 트랜잭션과는 별개일 수 있습니다.

### `MockMvc` 준비

`MockMvcBuilders.webAppContextSetup(context)`로 `DispatcherServlet`와 비슷한 구조를 구성합니다.

실제 요청은 서블릿 컨테이너가 아닌 Mock 환경에서 `MockHttpServletRequest`, `MockHttpServletResponse` 사용합니다.

### `MockMvc` 통한 요청 수행: `DispatcherServlet` 내부 흐름

Spring MVC는 다음과 같은 단계로 요청을 처리합니다:

1. `DispatcherServlet.doDispatch()`

    요청을 수신하면 `HandlerMapping`으로 요청을 처리할 `Controller` Bean을 결정합니다.
    `HandlerAdapter`로 실행 방법(보통 `RequestMappingHandlerAdapter`)을 결정합니다.

2. `HandlerInterceptor.preHandle()`

    `LocaleChangeInterceptor`, `SecurityContextPersistenceFilter`, custom 로깅 등을 처리하기 위해 `HandlerInterceptor`의 `preHandle()` 호출합니다.

3. 실제 `Controller` 호출

    `@RestController` 또는 `@Controller` 의 메서드가 실행됩니다.
    메서드 내부에서 보통 `@Service` -> `@Repository`까지 흐름이 이어집니다.

4. `@Transactional` 프록시

    ```java
    @Service
    public class OrderService {
        @Transactional
        public void placeOrder() {
            // DB insert
        }
    }
    ```

    ```sh
    Client (e.g. Controller)
        │
        ▼
    Spring AOP Proxy (CGLIB-generated OrderService subclass)
        │
        ▼
    TransactionInterceptor.invoke()
        ├─ getTransactionAttribute()
        ├─ txManager.getTransaction() // 트랜잭션 시작
        ├─ methodInvocation.proceed() // 실제 OrderService.placeOrder()
        └─ txManager.commit() 또는 txManager.rollback() // 커밋 또는 롤백
    ```

    `@Transactional` 어노테이션을 AOP로 처리하기 위해 다음과 같은 클래스들이 참여합니다:
    - `@EnableTransactionManagement`:

        트랜잭션 AOP 시스템의 활성화 시작점입니다.
        어떤 설정 클래스를 가져와야 할지를 결정합니다.
        `@Import(TransactionManagementConfigurationSelector.class)`를 내부적으로 포함합니다.

    - `TransactionManagementConfigurationSelector`:

        `@EnableTransactionManagement`를 처리하기 위한 설정 클래스를 동적으로 고르는 역할을 합니다.
        AOP 인프라 등록하는 `AutoProxyRegistrar`, 트랜잭션 관련 Bean을 구성하는 `ProxyTransactionManagementConfiguration` 두 클래스를 리턴합니다.

    - `ProxyTransactionManagementConfiguration`:

        `@Transactional` 어노테이션을 가진 Bean에 프록시를 씌우기 위해, 다음과 같이 필요한 모든 Bean을 등록합니다.
        - 어떤 트랜잭션 속성이 붙었는지 분석하는 `TransactionAttributeSource`
        - 트랜잭션을 실행/종료하는 실제 `Advice`인 `TransactionInterceptor`
        - `@Transactional`이 붙은 메서드 감지하고 `Advice`를 부착하는 `TransactionAdvisor`

    - `InfrastructureAdvisorAutoProxyCreator`:

        `BeanPostProcessor`의 일종으로, Spring Bean 초기화 단계에서 `@Transactional` 어노테이션이 붙은 Bean을 자동으로 프록시로 감싸는 역할을 수행합니다.

        `Advisor` (`Pointcut` + `Advice`)를 검사하고 해당 Bean에 프록시를 생성합니다.

    - `TransactionAttributeSource`:

        `@Transactional(propagation = REQUIRES_NEW, rollbackFor = ...)` 같은 트랜잭션 속성 정보를 코드에서 가져올 수 있도록 지원합니다.
        가령 `REQUIRES_NEW`는 현재 트랜잭션을 suspend 하고 새로운 트랜잭션 시작합니다.

        메서드 및 클래스를 기반으로 트랜잭션 속성을 분석합니다. (e.g. `AnnotationTransactionAttributeSource`)

    - `TransactionInterceptor`:
        - 메서드 호출 가로채어 트랜잭션 경계 처리

        프록시가 메서드 호출을 가로챘을 때, 실제로 트랜잭션을 시작/커밋/롤백을 처리합니다.
        `invoke(MethodInvocation)` 안에서 다음과 같은 작업들을 수행합니다.
        - 트랜잭션 속성 조회
        - `PlatformTransactionManager.getTransaction(...)` 호출
        - 메서드 실행 (`invocation.proceed()`)
        - 결과에 따라 `commit()` or `rollback()`

    - `PlatformTransactionManager`: 추
        - 상화된 트랜잭션 관리 인터페이스

        JDBC, JPA, Hibernate, JMS 등 다양한 트랜잭션 구현체에 대한 동일한 인터페이스를 정의합니다.
        - JDBC -> `DataSourceTransactionManager`
        - JPA EntityManager -> `JpaTransactionManager`
        - Hibernate SessionFactory -> `HibernateTransactionManager`
        - R2DBC -> `ReactiveTransactionManager`

    ```sh
    @EnableTransactionManagement
       │
       ▼
    TransactionManagementConfigurationSelector
       ├─ AutoProxyRegistrar ──> InfrastructureAdvisorAutoProxyCreator
       └─ ProxyTransactionManagementConfiguration
               ├─ AnnotationTransactionAttributeSource
               ├─ TransactionInterceptor
               ├─ Advisor (Pointcut + Interceptor)
               └─ PlatformTransactionManager (예: DataSourceTransactionManager)
    ```

    서비스가 AOP 프록시(`@Transactional`, `@Aspect`, `@Async` 등) 경우 프록시가 intercept 합니다.

    그러면 `TransactionInterceptor.invoke()` 호출되고, `TransactionAttributeSource`에서 해당 메서드에 대한 트랜잭션 속성 읽습니다.

    `PlatformTransactionManager.getTransaction()` 호출하여 트랜잭션 시작/중첩 여부를 판단하고, 이후 실제 메서드를 호출합니다.
    실제로는 `DataSourceTransactionManager`(JDBC) 혹은 `JpaTransactionManager`(JPA Entity Manager) 등 `PlatformTransactionManager` 구현체가 사용됩니다.
    이 구현체들은 JDBC나 JPA의 트랜잭션 시작/커밋/롤백 API를 호출합니다.

5. 로직 실행 및 DB 작업

    `@Repository`는 JDBC 또는 JPA를 호출합니다.
    `TransactionSynchronizationManager`에 트랜잭션 바인딩된 상태입니다.

6. 트랜잭션 종료

    서비스 메서드 정상 종료하면 커밋하고, 예외가 발생하면 롤백합니다.
    `TransactionStatus`에 따라 `commit()` or `rollback()`이 이뤄집니다.

7. `Interceptor.postHandle()` 또는 `Interceptor.afterCompletion()`: 응답 가공, cleanup 로직 등을 실행합니다.

### 응답 처리

`ViewResolver`는 JSON 또는 HTML로 변환합니다.
`HttpMessageConverter`를 통해 객체를 응답으로 직렬화하고, `MockHttpServletResponse`에 담아서 반환합니다.

### 테스트 종료

테스트 메서드에 `@Transactional`이 붙은 경우 `TestTransaction.endTransaction()` 호출되고, 무조건 rollback 처리합니다.

하지만, `MockMvc` 내부에서 `REQUIRES_NEW`로 실행된 트랜잭션은 별도의 트랜잭션이므로 rollback 되지 않습니다.
`TransactionSynchronizationManager`는 thread-local 기반으로 mockMvc와 별개

참고로 `@DirtiesContext`가 있는 경우 해당 context 폐기하지만, 그렇지 않으면 재사용합니다.
