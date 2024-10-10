# Spring Annotations

- [Spring Annotations](#spring-annotations)
    - [Spring Framework 주요 어노테이션 목록](#spring-framework-주요-어노테이션-목록)
    - [빈 구성 관련 어노테이션](#빈-구성-관련-어노테이션)
        - [`@Component`](#component)
        - [`@Service`](#service)
        - [`@Controller`](#controller)
        - [`@RestController`](#restcontroller)
        - [`@Repository`](#repository)
        - [`@Configuration`](#configuration)
        - [`@Bean`](#bean)
    - [의존성 주입 관련 어노테이션](#의존성-주입-관련-어노테이션)
        - [`@Autowired`](#autowired)
        - [`@Value`](#value)
    - [트랜잭션 관련 어노테이션](#트랜잭션-관련-어노테이션)
        - [`@Transactional`](#transactional)
    - [웹 요청 처리 관련 어노테이션](#웹-요청-처리-관련-어노테이션)
        - [`@RequestMapping`](#requestmapping)
        - [`@GetMapping`, `@PostMapping`, `@PutMapping`, `@DeleteMapping`](#getmapping-postmapping-putmapping-deletemapping)
        - [`@PathVariable`](#pathvariable)
        - [`@RequestParam`](#requestparam)
        - [`@RequestBody`](#requestbody)
        - [`@ResponseBody`](#responsebody)

## Spring Framework 주요 어노테이션 목록

- 빈 구성 관련 어노테이션:
    1. `@Component`: 일반적인 Spring 관리 컴포넌트를 정의합니다.
    2. `@Service`: 비즈니스 로직을 담당하는 서비스 계층 컴포넌트를 정의합니다.
    3. `@Controller`/`@RestController`: 웹 요청을 처리하는 컨트롤러를 정의합니다.
    4. `@Repository`: 데이터 접근 계층 컴포넌트를 정의합니다.
    5. `@Configuration`: 설정 클래스를 정의합니다.
    6. `@Bean`: Spring 컨테이너가 관리할 빈을 정의합니다.
    7. `@PostConstruct`: 빈이 초기화될 때 호출되는 메서드를 지정합니다.
    8. `@PreDestroy`: 빈이 소멸되기 직전에 호출되는 메서드를 지정합니다.
- 의존성 주입 관련 어노테이션:
    1. `@Autowired`: 의존성 자동 주입을 위해 사용됩니다.
        > 생성자나 세터 메서드에도 적용 가능합니다.
        > 스프링 4.3 이후부터는 생성자 주입을 더 권장합니다.
    2. `@Value`: 프로퍼티 값을 필드에 주입합니다.
    3. `@Qualifier`
- 유효성 검사와 바인딩 관련 어노테이션:
    1. `@Valid`: 자바에서 지원하는 유효성 검증 어노테이션입니다. 자바 표준 어노테이션(JSR-303)을 따릅니다.
    2. `@Validated`: 스프링의 Validator 인터페이스를 사용하여 유효성 검사를 합니다.
- AOP 관련 어노테이션:
    1. `@Aspect`
    2. `@Before`
    3. `@After`
    4. `@Around`
- 트랜잭션 관련 어노테이션:
    1. `@Transactional`: 트랜잭션 경계를 선언적으로 정의합니다.
- 웹 요청 처리 관련 어노테이션:
    1. `@RequestMapping`: 요청 URL을 특정 메서드와 매핑합니다.
    2. `@GetMapping`/`@PostMapping`/`@PutMapping`/`@DeleteMapping`: 특정 HTTP 메서드에 대한 요청 매핑을 정의합니다.
    3. `@PathVariable`: URL 경로의 변수를 메서드 파라미터로 바인딩합니다.
    4. `@RequestParam`: 요청 파라미터를 메서드 파라미터로 바인딩합니다.
    5. `@RequestBody`: HTTP 요청 본문을 자바 객체로 변환합니다.
    6. `@ResponseBody`: 자바 객체를 HTTP 응답 본문으로 변환합니다.
- 스프링 캐싱 관련 어노테이션:
    1. `@Cacheable`
    2. `@CachePut`
    3. `@CacheEvict`
- 보안 관련 어노테이션:
    1. `@PreAuthorize`
    2. `@Secured`
    3. `@RolesAllowed`
- 테스트 관련 어노테이션
    1. `@Test`
    2. `@RunWith`(SpringRunner.class)
    3. `@SpringBootTest`: 통합 테스트를 위한 어노테이션으로, 스프링 애플리케이션 컨텍스트를 로드합니다.
    4. `@MockBean`: 테스트에서 의존성을 모의하는 데 사용됩니다.
    5. `@DataJpaTest`: JPA 관련 테스트를 할 때 사용합니다.

## 빈 구성 관련 어노테이션

### [`@Component`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/stereotype/Component.html)

`@Component`는 Spring의 컴포넌트 스캔 메커니즘에 의해 자동으로 감지되고 Spring 컨테이너에 빈으로 등록되는 클래스를 표시하는 일반적인 스테레오타입 어노테이션입니다. 이 어노테이션은 클래스를 Spring의 관리 대상으로 만들어, 의존성 주입과 같은 Spring의 기능을 활용할 수 있게 합니다.

`@Component`는 다음과 같은 경우에 주로 사용됩니다:
- 특정 계층(서비스, 레포지토리 등)에 속하지 않는 일반적인 Spring 관리 컴포넌트를 정의할 때
- 사용자 정의 빈을 생성하고 싶을 때
- 다른 스테레오타입 어노테이션(@Service, @Repository, @Controller 등)의 메타 어노테이션으로 사용될 때

```java
// 기본 사용 방식
@Component
public class MyComponent {
    // 컴포넌트 로직
}

// 이름을 지정한 컴포넌트
@Component("customComponentName")
public class NamedComponent {
    // 컴포넌트 로직
}

// 다른 스테레오타입 어노테이션의 기반
@Target({ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Component
public @interface CustomComponent {
    String value() default "";
}
```

`@Component`의 내부 동작 방식:
1. 컴포넌트 스캔: `ApplicationContext`가 초기화될 때, `@ComponentScan`이 설정된 패키지와 그 하위 패키지를 스캔합니다.
2. 클래스 감지: `@Component`가 붙은 클래스를 찾습니다.
3. 빈 정의 생성: 발견된 각 `@Component` 클래스에 대해 BeanDefinition을 생성합니다.
4. 빈 등록: 생성된 `BeanDefinition`을 Spring 컨테이너에 등록합니다.
5. 빈 생성 및 초기화: 컨테이너는 필요에 따라 빈을 생성하고 초기화합니다.
6. 의존성 주입: 필요한 경우 다른 빈을 주입합니다.

```java
[ApplicationContext 초기화]
         |
         v
[ComponentScan 설정 확인]
         |
         v
[패키지 스캔 시작]
         |
         v
[`@Component` 클래스 감지]
         |
         v
[BeanDefinition 생성]
    |    |    |
    v    v    v
[빈1] [빈2] [빈3] ... [빈N] (Spring 컨테이너에 등록)
    |    |    |
    v    v    v
[빈 생성 및 초기화]
    |    |    |
    v    v    v
[의존성 주입 (필요시)]
    |    |    |
    v    v    v
[빈 사용 준비 완료]
```

### [`@Service`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/stereotype/Service.html)

`@Service`는 비즈니스 로직을 수행하는 서비스 계층의 컴포넌트를 나타내는 특별한 `@Component`입니다. 이 어노테이션은 클래스가 비즈니스 서비스를 제공한다는 것을 명시적으로 나타내며, 컴포넌트 스캔 메커니즘에 의해 자동으로 감지되고 Spring 컨테이너에 빈으로 등록됩니다.

`@Service`는 다음과 같은 경우에 주로 사용됩니다:
- 비즈니스 로직을 포함하는 클래스에 사용
- 트랜잭션 관리가 필요한 서비스 레이어에 사용
- 다른 서비스나 리포지토리를 조합하여 복잡한 비즈니스 프로세스를 구현할 때 사용

```java
// 기본 서비스 클래스
@Service
public class UserService {
    private final UserRepository userRepository;

    @Autowired
    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public User createUser(String username, String email) {
        // 비즈니스 로직 구현
    }
}

// 트랜잭션이 적용된 서비스
@Service
@Transactional
public class OrderService {
    // 주문 관련 비즈니스 로직
}

// 다른 서비스를 조합한 복잡한 서비스
@Service
public class ReportService {
    private final UserService userService;
    private final OrderService orderService;

    @Autowired
    public ReportService(UserService userService, OrderService orderService) {
        this.userService = userService;
        this.orderService = orderService;
    }

    public Report generateComplexReport() {
        // 여러 서비스를 조합한 복잡한 비즈니스 로직
    }
}
```

`@Service`의 내부 동작 방식:
1. 컴포넌트 스캔: `@ComponentScan`에 의해 `@Service` 어노테이션이 붙은 클래스를 감지합니다.
2. 빈 정의 생성: 해당 클래스에 대한 `BeanDefinition`을 생성합니다.
3. 빈 등록: Spring 컨테이너에 서비스 빈으로 등록합니다.
4. AOP 프록시 생성: 필요한 경우 (예: `@Transactional` 사용 시) AOP 프록시를 생성합니다.
5. 의존성 주입: 서비스에 필요한 다른 빈들을 주입합니다.

```java
[ApplicationContext 초기화]
         |
         v
[ComponentScan 실행]
         |
         v
[@Service 클래스 감지]
         |
         v
[BeanDefinition 생성]
         |
         v
[Spring 컨테이너에 빈 등록]
         |
         v
[AOP 관련 어노테이션 확인]
         |
    +----+----+
    |         |
    v         v
[AOP 필요] [AOP 불필요]
    |         |
    v         |
[프록시 생성]  |
    |         |
    +----+----+
         |
         v
[의존성 주입]
         |
         v
[초기화 메서드 실행 (있는 경우)]
         |
         v
[서비스 빈 사용 준비 완료]
```

Spring의 핵심 기능인 의존성 주입(DI)과 제어의 역전(IoC)을 구현하는 데 중요한 역할을 합니다.
`@Component`와 `@Service`는 빈의 생성과 관리를 Spring에 위임하며, `@Autowired`는 이러한 빈들 간의 의존성을 자동으로 연결해줍니다.

### [`@Controller`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/stereotype/Controller.html)

`@Controller`는 Spring MVC 프레임워크에서 웹 요청을 처리하는 클래스를 정의하는 데 사용되는 어노테이션입니다.
`@Component`의 특화된 형태로, 컴포넌트 스캔의 대상이 되어 자동으로 Spring 컨테이너에 빈으로 등록됩니다.

`@Controller`는 주로 다음과 같은 경우에 사용됩니다:
1. 웹 페이지를 반환하는 전통적인 Spring MVC 애플리케이션
2. RESTful API와 웹 페이지를 동시에 제공하는 애플리케이션
3. 뷰 기술(예: Thymeleaf, JSP)과 함께 사용되는 경우

```java
// 기본적인 @Controller 사용 예
@Controller
public class HomeController {

    @GetMapping("/")
    public String home(Model model) {
        model.addAttribute("message", "Welcome to our website!");
        return "home";  // 뷰 이름 반환
    }
}

// 모델 데이터와 함께 사용하는 예
@Controller
@RequestMapping("/users")
public class UserController {

    @Autowired
    private UserService userService;

    @GetMapping
    public String listUsers(Model model) {
        model.addAttribute("users", userService.getAllUsers());
        return "userList";
    }

    @PostMapping
    public String createUser(@ModelAttribute User user, BindingResult result) {
        if (result.hasErrors()) {
            return "userForm";
        }
        userService.createUser(user);
        return "redirect:/users";
    }
}

// RESTful API 엔드포인트를 포함한 컨트롤러
@Controller
@RequestMapping("/api")
public class ApiController {

    @GetMapping("/data")
    @ResponseBody
    public Map<String, Object> getData() {
        Map<String, Object> data = new HashMap<>();
        data.put("key", "value");
        return data;  // 자동으로 JSON으로 변환됨
    }
}

// 예외 처리를 포함한 컨트롤러
@Controller
public class ProductController {

    @GetMapping("/products/{id}")
    public String getProduct(@PathVariable Long id, Model model) {
        try {
            model.addAttribute("product", productService.getProduct(id));
            return "productDetails";
        } catch (ProductNotFoundException e) {
            model.addAttribute("error", "Product not found");
            return "error";
        }
    }

    @ExceptionHandler(ProductNotFoundException.class)
    public String handleProductNotFound(ProductNotFoundException ex, Model model) {
        model.addAttribute("error", ex.getMessage());
        return "error";
    }
}
```

`@Controller`의 내부 동작 방식:
1. 컴포넌트 스캔: `ApplicationContext` 초기화 시, `@ComponentScan`이 `@Controller` 어노테이션이 붙은 클래스를 감지합니다.
2. 빈 등록: 감지된 `@Controller` 클래스에 대한 빈 정의를 생성하고 Spring 컨테이너에 등록합니다.
3. 요청 매핑 초기화: `RequestMappingHandlerMapping`이 컨트롤러의 메서드에 정의된 `@RequestMapping` (또는 관련 어노테이션)을 처리하여 요청 경로와 메서드를 매핑합니다.
4. AOP 프록시 생성: 필요한 경우 (예: `@Transactional` 사용 시) AOP 프록시를 생성합니다.
5. 의존성 주입: 컨트롤러에 필요한 다른 빈들(예: `Service`, `Repository`)을 주입합니다.
6. 초기화: `@PostConstruct` 메서드 등을 실행하여 컨트롤러를 초기화합니다.

```java
[ApplicationContext 초기화]
         |
         v
[ComponentScan 실행]
         |
         v
[@Controller 클래스 감지]
         |
         v
[빈 정의 생성 및 등록]
         |
         v
[RequestMappingHandlerMapping 초기화]
         |
         v
[AOP 관련 어노테이션 확인]
         |
    +----+----+
    |         |
    v         v
[AOP 필요] [AOP 불필요]
    |         |
    v         |
[프록시 생성]    |
    |         |
    +----+----+
         |
         v
[의존성 주입 (Service, Repository 등)]
         |
         v
[초기화 메서드 실행 (@PostConstruct)]
         |
         v
[컨트롤러 빈 사용 준비 완료]
```

`@Controller`를 사용한 요청 처리 흐름:

```java
[클라이언트 요청]
         |
         v
[DispatcherServlet]
         |
         v
[HandlerMapping]  --- 요청 URL에 맞는 컨트롤러 메서드 찾기
         |
         v
[HandlerAdapter]  --- 컨트롤러 메서드 실행
         |
         v
[@Controller 메서드]
         |
         v
[뷰 이름 반환]
         |
         v
[ViewResolver]  --- 뷰 이름을 실제 View 객체로 변환
         |
         v
[View 렌더링]  --- 모델 데이터를 사용하여 뷰 생성
         |
         v
[클라이언트에 응답]
```

`@Controller`의 주요 특징:
1. 뷰 이름 반환: 일반적으로 String 타입의 뷰 이름을 반환하여 ViewResolver가 실제 뷰를 찾아 렌더링하도록 합니다.
2. 모델 데이터 처리: Model 객체를 통해 뷰에 전달할 데이터를 설정할 수 있습니다.
3. 폼 처리: `@ModelAttribute`를 사용하여 폼 데이터를 객체에 바인딩하고, `BindingResult`로 유효성 검사 결과를 처리할 수 있습니다.
4. 예외 처리: `@ExceptionHandler`를 사용하여 컨트롤러 내에서 발생하는 특정 예외를 처리할 수 있습니다.
5. `@ResponseBody`: 메서드에 이 어노테이션을 사용하면 뷰를 거치지 않고 직접 HTTP 응답 본문을 작성할 수 있습니다.
6. `@InitBinder`: 데이터 바인딩과 유효성 검사를 커스터마이즈할 수 있습니다.

    ```java
    @Controller
    public class UserController {
        @InitBinder
        public void initBinder(WebDataBinder binder) {
            SimpleDateFormat dateFormat = new SimpleDateFormat("yyyy-MM-dd");
            dateFormat.setLenient(false);
            binder.registerCustomEditor(Date.class, new CustomDateEditor(dateFormat, false));
        }
    }
    ```

7. `@ModelAttribute` 메서드: 컨트롤러의 모든 요청 처리 메서드에서 공통으로 사용할 모델 속성을 정의할 수 있습니다.

    ```java
    @Controller
    public class ProductController {
        @ModelAttribute("categories")
        public List<String> getCategories() {
            return Arrays.asList("Electronics", "Books", "Clothing");
        }
    }
    ```

### [`@RestController`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/RestController.html)

`@RestController`는 Spring Framework 4.0에서 도입된 어노테이션으로, `@Controller`와 `@ResponseBody`를 결합한 특화된 컨트롤러입니다.
모든 요청 처리 메서드에 `@ResponseBody`가 기본적으로 적용됩니다.

`@RestController`는 주로 다음과 같은 경우에 사용됩니다:
1. RESTful API를 제공하는 애플리케이션 개발
2. 마이크로서비스 아키텍처에서 서비스 엔드포인트 구현
3. 프론트엔드와 백엔드가 분리된 Single Page Application(SPA) 백엔드 구현
4. 모바일 애플리케이션을 위한 백엔드 API 개발

```java
// 기본적인 @RestController 사용 예
@RestController
@RequestMapping("/api/users")
public class UserRestController {

    @Autowired
    private UserService userService;

    @GetMapping
    public List<User> getAllUsers() {
        return userService.getAllUsers();
    }

    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable Long id) {
        return userService.getUser(id)
                          .map(ResponseEntity::ok)
                          .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody @Valid User user) {
        User createdUser = userService.createUser(user);
        URI location = ServletUriComponentsBuilder
            .fromCurrentRequest()
            .path("/{id}")
            .buildAndExpand(createdUser.getId())
            .toUri();
        return ResponseEntity.created(location).body(createdUser);
    }

    @PutMapping("/{id}")
    public ResponseEntity<User> updateUser(@PathVariable Long id, @RequestBody @Valid User user) {
        return userService.updateUser(id, user)
                          .map(ResponseEntity::ok)
                          .orElse(ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        if (userService.deleteUser(id)) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}

// 예외 처리를 포함한 @RestController
@RestController
@RequestMapping("/api/products")
public class ProductRestController {

    @Autowired
    private ProductService productService;

    @GetMapping("/{id}")
    public Product getProduct(@PathVariable Long id) {
        return productService.getProduct(id)
                             .orElseThrow(() -> new ProductNotFoundException(id));
    }

    @ExceptionHandler(ProductNotFoundException.class)
    @ResponseStatus(HttpStatus.NOT_FOUND)
    public ErrorResponse handleProductNotFound(ProductNotFoundException ex) {
        return new ErrorResponse("Product not found", ex.getMessage());
    }
}

// 페이징과 정렬을 지원하는 @RestController
@RestController
@RequestMapping("/api/books")
public class BookRestController {

    @Autowired
    private BookRepository bookRepository;

    @GetMapping
    public Page<Book> getAllBooks(Pageable pageable) {
        return bookRepository.findAll(pageable);
    }
}
```

`@RestController`의 내부 동작 방식:
1. 컴포넌트 스캔: `ApplicationContext` 초기화 시, `@ComponentScan`이 `@RestController` 어노테이션이 붙은 클래스를 감지합니다.
2. 빈 등록: 감지된 `@RestController` 클래스에 대한 빈 정의를 생성하고 Spring 컨테이너에 등록합니다.
3. 요청 매핑 초기화: `RequestMappingHandlerMapping`이 컨트롤러의 메서드에 정의된 `@RequestMapping`(또는 관련 어노테이션)을 처리하여 요청 경로와 메서드를 매핑합니다.
4. `ResponseBody` 처리기 등록: 모든 메서드에 `@ResponseBody`가 적용된 것처럼 처리합니다.
5. 메시지 변환기 초기화: `HttpMessageConverter` 구현체들을 초기화하여 Java 객체와 HTTP 메시지 간의 변환을 준비합니다.
6. AOP 프록시 생성: 필요한 경우 (예: `@Transactional` 사용 시) AOP 프록시를 생성합니다.
7. 의존성 주입: 컨트롤러에 필요한 다른 빈들(예: `Service`, `Repository`)을 주입합니다.
8. 초기화: `@PostConstruct` 메서드 등을 실행하여 컨트롤러를 초기화합니다.

```java
[ApplicationContext 초기화]
         |
         v
[ComponentScan 실행]
         |
         v
[@RestController 클래스 감지]
         |
         v
[빈 정의 생성 및 등록]
         |
         v
[RequestMappingHandlerMapping 초기화]
         |
         v
[ResponseBody 처리기 등록]
         |
         v
[메시지 변환기(HttpMessageConverter) 초기화]
         |
         v
[AOP 관련 어노테이션 확인]
         |
    +----+----+
    |         |
    v         v
[AOP 필요] [AOP 불필요]
    |         |
    v         |
[프록시 생성]  |
    |         |
    +----+----+
         |
         v
[의존성 주입 (Service, Repository 등)]
         |
         v
[초기화 메서드 실행 (@PostConstruct)]
         |
         v
[RestController 빈 사용 준비 완료]
```

`@RestController`를 사용한 요청 처리 흐름:

```java
[클라이언트 요청]
         |
         v
[DispatcherServlet]
         |
         v
[HandlerMapping]  --- 요청 URL에 맞는 컨트롤러 메서드 찾기
         |
         v
[HandlerAdapter]  --- 컨트롤러 메서드 실행
         |
         v
[@RestController 메서드]
         |
         v
[반환 객체]
         |
         v
[HttpMessageConverter]  --- Java 객체를 JSON/XML 등으로 변환
         |
         v
[클라이언트에 응답]
```

`@RestController`의 주요 특징:

1. 자동 직렬화: 반환된 객체는 자동으로 JSON 또는 XML로 직렬화됩니다.
2. `ContentNegotiationViewResolver`: Accept 헤더에 따라 적절한 `HttpMessageConverter`를 선택하여 응답을 생성합니다.
3. `@RequestBody`: 요청 본문을 자동으로 Java 객체로 역직렬화합니다.

    ```java
    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody User user) {
        // user 객체는 JSON 요청 본문에서 자동으로 생성됨
        User createdUser = userService.createUser(user);
        return ResponseEntity.ok(createdUser);
    }
    ```

4. `ResponseEntity`: HTTP 상태 코드, 헤더, 본문을 세밀하게 제어할 수 있습니다.

    ```java
    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable Long id) {
        return userService.getUser(id)
                        .map(user -> ResponseEntity.ok().header("Custom-Header", "Value").body(user))
                        .orElse(ResponseEntity.notFound().build());
    }
    ```

5. `@ControllerAdvice`와 결합: 전역 예외 처리와 모델 속성 추가 등을 구현할 수 있습니다.

    ```java
    @RestControllerAdvice
    public class GlobalExceptionHandler {

        @ExceptionHandler(ResourceNotFoundException.class)
        @ResponseStatus(HttpStatus.NOT_FOUND)
        public ErrorResponse handleResourceNotFound(ResourceNotFoundException ex) {
            return new ErrorResponse("Resource not found", ex.getMessage());
        }
    }
    ```

6. 비동기 처리: `Callable`, `DeferredResult`, `CompletableFuture` 등을 사용하여 비동기 요청을 처리할 수 있습니다.

    ```java
    @GetMapping("/async")
    public Callable<String> handleAsync() {
        return () -> {
            Thread.sleep(5000); // 긴 작업 시뮬레이션
            return "Async result";
        };
    }
    ```

### [`@Repository`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/stereotype/Repository.html)

`@Repository`는 Spring Framework에서 데이터 접근 계층(Data Access Layer)의 컴포넌트를 나타내는 `@Component` 어노테이션입니다.
데이터베이스 작업을 수행하는 DAO(Data Access Object) 또는 리포지토리 클래스에 사용됩니다.

`@Repository`는 다음과 같은 주요 기능과 특징을 제공합니다:
1. 컴포넌트 스캔: `@Component`의 특화된 형태로, 컴포넌트 스캔의 대상이 됩니다.
2. 예외 변환: 특정 기술(예: JDBC, JPA)의 예외를 Spring의 `DataAccessException`으로 자동 변환합니다.
3. 트랜잭션 참여: `@Transactional`과 함께 사용될 때 트랜잭션 관리에 참여합니다.

`@Repository`는 주로 다음과 같은 경우에 사용됩니다:
- 데이터베이스와 직접 상호작용하는 클래스에 사용
- ORM 프레임워크(예: JPA, Hibernate)를 사용하는 리포지토리 클래스에 적용
- 캐시나 외부 저장소와 상호작용하는 데이터 접근 계층에 사용

```java
// 기본적인 @Repository 사용 예
@Repository
public class UserRepository {
    @PersistenceContext
    private EntityManager entityManager;

    public User findById(Long id) {
        return entityManager.find(User.class, id);
    }

    public void save(User user) {
        entityManager.persist(user);
    }
}

// Spring Data JPA와 함께 사용하는 예
@Repository
public interface ProductRepository extends JpaRepository<Product, Long> {
    List<Product> findByCategory(String category);
}

// JDBC를 직접 사용하는 예
@Repository
public class OrderRepository {
    @Autowired
    private JdbcTemplate jdbcTemplate;

    public Order findById(Long id) {
        return jdbcTemplate.queryForObject(
            "SELECT * FROM orders WHERE id = ?",
            new Object[]{id},
            (rs, rowNum) -> new Order(rs.getLong("id"), rs.getString("status"))
        );
    }
}

// 캐시와 함께 사용하는 예
@Repository
public class CachedUserRepository {
    @Autowired
    private UserRepository userRepository;

    @Autowired
    private CacheManager cacheManager;

    public User findById(Long id) {
        Cache cache = cacheManager.getCache("users");
        User user = cache.get(id, User.class);
        if (user == null) {
            user = userRepository.findById(id);
            cache.put(id, user);
        }
        return user;
    }
}
```

`@Repository`의 내부 동작 방식:

1. 컴포넌트 스캔: ApplicationContext 초기화 시, `@Component`Scan이 `@Repository` 어노테이션이 붙은 클래스를 감지합니다.
2. 빈 등록: 감지된 `@Repository` 클래스에 대한 빈 정의를 생성하고 Spring 컨테이너에 등록합니다.
3. 예외 변환 설정: PersistenceExceptionTranslationPostProcessor가 `@Repository` 빈에 대한 예외 변환 기능을 설정합니다.
4. AOP 프록시 생성: 필요한 경우 (예: @Transactional 사용 시) AOP 프록시를 생성합니다.
5. 의존성 주입: 리포지토리에 필요한 다른 빈들(예: EntityManager, JdbcTemplate)을 주입합니다.
6. 초기화: @PostConstruct 메서드 등을 실행하여 리포지토리를 초기화합니다.

```java
[ApplicationContext 초기화]
         |
         v
[ComponentScan 실행]
         |
         v
[@Repository 클래스 감지]
         |
         v
[빈 정의 생성 및 등록]
         |
         v
[PersistenceExceptionTranslationPostProcessor 적용]
         |
         v
[AOP 관련 어노테이션 확인]
         |
    +----+----+
    |         |
    v         v
[AOP 필요] [AOP 불필요]
    |         |
    v         |
[프록시 생성]  |
    |         |
    +----+----+
         |
         v
[의존성 주입 (EntityManager, JdbcTemplate 등)]
         |
         v
[초기화 메서드 실행 (@PostConstruct)]
         |
         v
[리포지토리 빈 사용 준비 완료]
```

`@Repository`의 예외 변환 메커니즘은 특히 중요합니다:
1. 예외 발생: 리포지토리 메서드 실행 중 데이터 접근 관련 예외가 발생합니다.
2. AOP 인터셉션: `@Repository`로 생성된 프록시가 예외를 인터셉트합니다.
3. 예외 분석: `PersistenceExceptionTranslationPostProcessor`가 발생한 예외를 분석합니다.
4. 예외 변환: 특정 기술의 예외를 Spring의 `DataAccessException` 계층구조의 적절한 예외로 변환합니다.
5. 예외 전파: 변환된 예외가 호출자에게 전파됩니다.

이런 예외 변환 메커니즘의 주요 이점은 다음과 같습니다:
- 데이터 접근 기술에 독립적인 예외 처리 가능
- 일관된 예외 처리 로직 구현 가능
- 특정 데이터 접근 기술 변경 시 비즈니스 로직 수정 불필요

### [`@Configuration`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/context/annotation/Configuration.html)

`@Configuration`은 Spring의 Java 기반 구성을 정의하는 클래스에 사용되는 어노테이션입니다.
이 어노테이션이 적용된 클래스는 하나 이상의 `@Bean` 메서드를 선언하여 Spring 컨테이너에 의해 관리될 빈 정의를 생성합니다.

이를 통해 XML 구성 없이도 유연하고 타입 안전한 방식으로 애플리케이션의 빈을 정의하고 관리할 수 있습니다.

`@Configuration`은 주로 다음과 같은 경우에 사용됩니다:
1. Java 코드로 Spring 빈 정의
2. 외부 라이브러리나 서드파티 클래스의 빈 등록
3. 프로파일 기반 조건부 빈 구성
4. 복잡한 빈 초기화 로직 구현
5. 빈 간의 의존성 설정

```java
@Configuration
public class AppConfig {

    @Autowired
    private Environment env;

    // 기본 빈 정의
    @Bean
    public UserService userService() {
        return new UserServiceImpl();
    }

    // 의존성 주입을 포함한 빈 정의
    @Bean
    public OrderService orderService(UserService userService) {
        return new OrderServiceImpl(userService);
    }

    // 조건부 빈 정의
    @Bean
    @ConditionalOnProperty(name = "cache.enabled", havingValue = "true")
    public CacheManager cacheManager() {
        return new ConcurrentMapCacheManager();
    }

    // 프로파일 기반 빈 정의
    @Bean
    @Profile("development")
    public DataSource developmentDataSource() {
        return new EmbeddedDatabaseBuilder()
                .setType(EmbeddedDatabaseType.H2)
                .build();
    }

    @Bean
    @Profile("production")
    public DataSource productionDataSource() {
        DriverManagerDataSource dataSource = new DriverManagerDataSource();
        dataSource.setDriverClassName(env.getProperty("db.driver"));
        dataSource.setUrl(env.getProperty("db.url"));
        dataSource.setUsername(env.getProperty("db.username"));
        dataSource.setPassword(env.getProperty("db.password"));
        return dataSource;
    }

    // 외부 라이브러리 빈 등록
    @Bean
    public RestTemplate restTemplate() {
        return new RestTemplate();
    }

    // 복잡한 초기화 로직을 가진 빈
    @Bean
    public ComplexService complexService() {
        ComplexService service = new ComplexService();
        service.setProperty1(env.getProperty("complex.prop1"));
        service.setProperty2(Long.parseLong(env.getProperty("complex.prop2")));
        service.initialize();
        return service;
    }
}
```

`@Configuration`의 내부 동작 방식:
1. 구성 클래스 감지: Spring 애플리케이션 컨텍스트가 시작될 때, `@Configuration` 어노테이션이 붙은 클래스를 스캔합니다.
2. CGLIB 기반 프록시 생성: 감지된 `@Configuration` 클래스에 대해 CGLIB 기반의 서브클래스 프록시를 생성합니다.
3. Bean 메서드 처리: `@Bean` 어노테이션이 붙은 메서드를 찾아 처리합니다.
4. 빈 정의 등록: 각 `@Bean` 메서드에 대해 `BeanDefinition`을 생성하고 Spring 컨테이너에 등록합니다.
5. 의존성 해결: 빈 간의 의존성을 해결하고 필요한 경우 주입합니다.
6. 빈 생성 및 초기화: 등록된 빈 정의에 따라 실제 빈 인스턴스를 생성하고 초기화합니다.

```java
[Spring 애플리케이션 컨텍스트 시작]
         |
         v
[@Configuration 클래스 스캔]
         |
         v
[CGLIB 기반 프록시 생성]
         |
         v
[@Bean 메서드 탐색]
         |
         v
[빈 정의 생성 및 등록]
    +----+----+
    |         |
    v         v
[의존성 해결] [조건 평가 (@Conditional)]
    |         |
    +----+----+
         |
         v
[빈 인스턴스 생성]
         |
         v
[빈 초기화 (@PostConstruct 등)]
         |
         v
[애플리케이션 컨텍스트 준비 완료]
```

1. 싱글톤 보장:

    `@Configuration` 클래스 내의 `@Bean` 메서드는 기본적으로 싱글톤을 보장합니다.
    같은 빈을 여러 번 참조해도 항상 동일한 인스턴스가 반환됩니다.

    ```java
    @Configuration
    public class AppConfig {
        @Bean
        public MyService myService() {
            return new MyServiceImpl();
        }

        @Bean
        public MyController myController() {
            return new MyController(myService());  // 같은 MyService 인스턴스 사용
        }
    }
    ```

2. `@Autowired` 사용:

    `@Configuration` 클래스에서 다른 빈을 주입받을 때 `@Autowired`를 사용할 수 있습니다.

    ```java
    @Configuration
    public class AppConfig {
        @Autowired
        private Environment env;

        @Bean
        public DataSource dataSource() {
            DriverManagerDataSource ds = new DriverManagerDataSource();
            ds.setUrl(env.getProperty("db.url"));
            // ... 기타 설정
            return ds;
        }
    }
    ```

3. `@Import` 사용:

    다른 `@Configuration` 클래스를 가져와 사용할 수 있습니다.

    ```java
    @Configuration
    @Import({SecurityConfig.class, PersistenceConfig.class})
    public class AppConfig {
        // ...
    }
    ```

4. 프로파일과 조건부 빈:

    `@Profile`과 `@Conditional` 어노테이션을 사용하여 환경에 따른 조건부 빈 구성이 가능합니다.

    ```java
    @Configuration
    public class AppConfig {
        @Bean
        @Profile("development")
        public DataSource embeddedDataSource() {
            // 개발 환경용 내장 데이터소스
        }

        @Bean
        @Profile("production")
        @ConditionalOnProperty(name = "use.jndi", havingValue = "true")
        public DataSource jndiDataSource() {
            // JNDI 데이터소스
        }
    }
    ```

5. 메서드 호출을 통한 의존성 주입:

    `@Bean` 메서드 내에서 다른 `@Bean` 메서드를 직접 호출하여 의존성을 주입할 수 있습니다.

    ```java
    @Configuration
    public class AppConfig {
        @Bean
        public Engine engine() {
            return new Engine();
        }

        @Bean
        public Car car() {
            return new Car(engine());  // engine() 메서드 직접 호출
        }
    }
    ```

6. `@Order` 사용:

    여러 `@Configuration` 클래스의 처리 순서를 지정할 수 있습니다.

    ```java
    @Configuration
    @Order(1)
    public class SecurityConfig {
        // ...
    }

    @Configuration
    @Order(2)
    public class PersistenceConfig {
        // ...
    }
    ```

7. 생성자 주입 사용:

    Spring 4.3부터는 `@Configuration` 클래스에서도 생성자 주입을 사용할 수 있습니다.

    ```java
    @Configuration
    public class AppConfig {
        private final Environment env;

        public AppConfig(Environment env) {
            this.env = env;
        }

        @Bean
        public DataSource dataSource() {
            // env 사용
        }
    }
    ```

8. `@PropertySource` 사용:

    외부 프로퍼티 파일을 로드하여 사용할 수 있습니다.

    ```java
    @Configuration
    @PropertySource("classpath:app.properties")
    public class AppConfig {
        @Autowired
        private Environment env;

        @Bean
        public MyBean myBean() {
            MyBean bean = new MyBean();
            bean.setName(env.getProperty("bean.name"));
            return bean;
        }
    }
    ```

### [`@Bean`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/context/annotation/Bean.html)

`@Bean`은 Spring의 Java 기반 구성에서 메서드 레벨에 사용되는 어노테이션으로,
해당 메서드가 Spring 컨테이너에 의해 관리될 빈을 생성한다는 것을 나타냅니다.
주로 `@Configuration` 클래스 내에서 사용되며, 메서드가 반환하는 객체를 Spring 빈으로 등록합니다.

이를 통해 **XML 구성 없이** 유연하고 타입 안전한 방식으로 빈을 정의하고 관리할 수 있습니다.

`@Bean`은 주로 다음과 같은 경우에 사용됩니다:
1. 서드파티 라이브러리의 클래스를 빈으로 등록
2. 조건부 빈 생성
3. 팩토리 메서드를 통한 복잡한 빈 초기화
4. 메서드 이름을 통한 빈 이름 지정
5. 빈의 스코프 및 생명주기 관리

```java
@Configuration
public class AppConfig {

    @Autowired
    private Environment env;

    // 기본 빈 정의
    @Bean
    public DataSource dataSource() {
        DriverManagerDataSource dataSource = new DriverManagerDataSource();
        dataSource.setDriverClassName(env.getProperty("db.driver"));
        dataSource.setUrl(env.getProperty("db.url"));
        dataSource.setUsername(env.getProperty("db.username"));
        dataSource.setPassword(env.getProperty("db.password"));
        return dataSource;
    }

    // 이름이 지정된 빈
    @Bean(name = "myCustomBean")
    public MyService myServiceBean() {
        return new MyServiceImpl();
    }

    // 초기화 및 파괴 메서드 지정
    @Bean(initMethod = "init", destroyMethod = "cleanup")
    public AnotherService anotherService() {
        return new AnotherServiceImpl();
    }

    // 조건부 빈 생성
    @Bean
    @Conditional(DataSourceCondition.class)
    public DataSourceHealthIndicator dataSourceHealthIndicator(DataSource dataSource) {
        return new DataSourceHealthIndicator(dataSource);
    }

    // 빈 스코프 지정
    @Bean
    @Scope("prototype")
    public PrototypeBean prototypeBean() {
        return new PrototypeBean();
    }

    // 팩토리 메서드를 통한 빈 생성
    @Bean
    public static PropertySourcesPlaceholderConfigurer propertySourcesPlaceholderConfigurer() {
        return new PropertySourcesPlaceholderConfigurer();
    }

    // 다른 빈에 의존하는 빈
    @Bean
    public UserService userService(DataSource dataSource) {
        UserServiceImpl service = new UserServiceImpl();
        service.setDataSource(dataSource);
        return service;
    }

    // 컬렉션 빈 생성
    @Bean
    public List<String> myStringList() {
        return Arrays.asList("value1", "value2", "value3");
    }
}
```

`@Bean`의 주요 속성:
1. `name` 또는 `value`: 빈의 이름을 지정합니다. 기본값은 메서드 이름입니다.
2. `initMethod`: 빈 초기화 시 호출될 메서드 이름을 지정합니다.
3. `destroyMethod`: 빈 소멸 시 호출될 메서드 이름을 지정합니다.
4. `autowireCandidate`: 이 빈을 자동 와이어링 후보로 고려할지 여부를 지정합니다.

`@Bean`의 내부 동작 방식:
1. 구성 클래스 처리: Spring 컨테이너가 `@Configuration` 클래스를 처리합니다.
2. 빈 메서드 식별: `@Bean` 어노테이션이 붙은 메서드를 식별합니다.
3. 빈 정의 생성: 각 `@Bean` 메서드에 대해 `BeanDefinition`을 생성합니다.
4. 빈 등록: 생성된 `BeanDefinition`을 Spring 컨테이너에 등록합니다.
5. 빈 초기화: 필요한 시점에 `@Bean` 메서드를 호출하여 빈 인스턴스를 생성하고 초기화합니다.
6. 의존성 주입: 필요한 경우 다른 빈을 주입합니다.
7. 생명주기 관리: 지정된 초기화 메서드와 소멸 메서드를 적절한 시점에 호출합니다.

```java
[@Configuration 클래스 처리]
         |
         v
[@Bean 메서드 식별]
         |
         v
[BeanDefinition 생성]
         |
         v
[빈 등록 (Spring 컨테이너)]
         |
         v
[빈 인스턴스 생성 요청]
         |
         v
[@Bean 메서드 호출]
         |
         v
[빈 인스턴스 생성]
         |
         v
[의존성 주입]
         |
         v
[초기화 메서드 호출 (지정된 경우)]
         |
         v
[빈 사용 준비 완료]
    ... (애플리케이션 실행) ...
         |
         v
[컨테이너 종료 시 소멸 메서드 호출 (지정된 경우)]
```

1. 싱글톤 보장:

    `@Configuration` 클래스 내의 `@Bean` 메서드는 기본적으로 싱글톤 인스턴스를 보장합니다.

    ```java
    @Configuration
    public class AppConfig {
        @Bean
        public MyService myService() {
            return new MyServiceImpl();
        }

        @Bean
        public MyController myController() {
            return new MyController(myService()); // 같은 MyService 인스턴스 반환
        }
    }
    ```

2. 메서드 파라미터를 통한 의존성 주입:

    `@Bean` 메서드의 파라미터로 다른 빈을 받아 의존성을 주입할 수 있습니다.

    ```java
    @Bean
    public UserService userService(
        DataSource dataSource,
        EmailService emailService
    ) {
        UserServiceImpl service = new UserServiceImpl();
        service.setDataSource(dataSource);
        service.setEmailService(emailService);
        return service;
    }
    ```

3. 조건부 빈 생성:

    `@Conditional` 어노테이션과 함께 사용하여 조건에 따라 빈을 생성할 수 있습니다.

    ```java
    @Bean
    @Conditional(WindowsCondition.class)
    public WindowsService windowsService() {
        return new WindowsServiceImpl();
    }
    ```

4. 빈 스코프 지정:

    `@Scope` 어노테이션을 사용하여 빈의 스코프를 지정할 수 있습니다.

    ```java
    @Bean
    @Scope("prototype")
    public PrototypeBean prototypeBean() {
        return new PrototypeBean();
    }
    ```

5. 팩토리 메서드:

    `static` `@Bean` 메서드를 사용하여 팩토리 메서드를 정의할 수 있습니다.

    ```java
    @Bean
    public static PropertySourcesPlaceholderConfigurer propertySourcesPlaceholderConfigurer() {
        return new PropertySourcesPlaceholderConfigurer();
    }
    ```

6. 초기화 및 소멸 메서드:

    빈의 초기화와 소멸 시 호출될 메서드를 지정할 수 있습니다.

    ```java
    @Bean(initMethod = "init", destroyMethod = "cleanup")
    public DatabaseService databaseService() {
        return new DatabaseServiceImpl();
    }
    ```

7. 빈 별칭:

    하나의 빈에 여러 이름을 부여할 수 있습니다.

    ```java
    @Bean(name = { "dataSource", "subsystemA-dataSource", "subsystemB-dataSource" })
    public DataSource dataSource() {
        // ...
    }
    ```

8. 프로파일 특정 빈:

    `@Profile` 어노테이션과 함께 사용하여 특정 프로파일에서만 빈을 생성할 수 있습니다.

    ```java
    @Bean
    @Profile("development")
    public DataSource developmentDataSource() {
        // ...
    }
    ```

## 의존성 주입 관련 어노테이션

### [`@Autowired`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/beans/factory/annotation/Autowired.html)

`@Autowired`는 Spring의 의존성 주입 메커니즘을 통해 빈을 자동으로 주입하는 데 사용되는 어노테이션입니다. 이 어노테이션은 생성자, 메서드, 또는 필드에 적용될 수 있으며, Spring이 해당 타입의 빈을 찾아 자동으로 주입하도록 지시합니다.

`@Autowired`는 다음과 같은 경우에 주로 사용됩니다:
- 생성자 주입: 클래스의 생성자에 사용하여 필수 의존성을 주입할 때 (권장 방식)
- 필드 주입: 클래스의 필드에 직접 사용하여 의존성을 주입할 때
- 세터 메서드 주입: 선택적 의존성을 주입하거나 런타임에 의존성을 변경해야 할 때

```java
// 생성자 주입 (권장 방식)
public class UserService {
    private final UserRepository userRepository;

    @Autowired
    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }
}

// 필드 주입
public class ProductService {
    @Autowired
    private ProductRepository productRepository;
}

// 세터 메서드 주입
public class OrderService {
    private PaymentService paymentService;

    @Autowired
    public void setPaymentService(PaymentService paymentService) {
        this.paymentService = paymentService;
    }
}
```

`@Autowired`의 내부 동작 방식:
1. 빈 생성: Spring 컨테이너가 빈을 생성합니다.
2. 의존성 분석: `@Autowired`가 붙은 필드, 생성자, 또는 메서드를 찾습니다.
3. 타입 매칭: 주입할 의존성의 타입을 분석합니다.
4. 후보 빈 검색: 해당 타입의 빈을 컨테이너에서 찾습니다.
5. 의존성 주입: 찾은 빈을 해당 위치에 주입합니다.
6. 순환 참조 확인: 순환 참조가 있는지 확인하고, 있다면 예외를 발생시킵니다.

```java
[빈 생성]
    |
    v
[@Autowired 위치 스캔]
    |
    v
[의존성 타입 분석]
    |
    v
[컨테이너에서 빈 검색]
    |
    +---------------------+
    |                     |
    v                     v
[단일 빈 발견]    [여러 후보 빈 발견]
    |                     |
    |              [한정자(@Qualifier) 확인]
    |                     |
    v                     v
[빈 주입]          [적절한 빈 선택 및 주입]
    |                     |
    +---------------------+
    |
    v
[순환 참조 확인]
    |
    v
[빈 초기화 완료]
```

### [`@Value`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/beans/factory/annotation/Value.html)

`@Value`는 Spring 프레임워크에서 필드, 메서드 파라미터, 또는 생성자 파라미터에 외부 속성 값을 주입하는 데 사용되는 어노테이션입니다.
이 어노테이션을 사용하면 다음과 같은 외부 소스에서 값을 가져와 빈의 속성에 주입할 수 있습니다.
- 프로퍼티 파일
- 환경 변수
- 시스템 프로퍼티 등

SpEL의 강력한 기능을 활용하여 복잡한 값 계산이나 조건부 로직을 구현할 수 있습니다.
그러나 너무 많은 `@Value` 사용은 코드를 복잡하게 만들 수 있으므로, `@ConfigurationProperties`를 고려하여 구조화된 구성을 사용할 수도 있습니다.

`@Value`는 주로 다음과 같은 경우에 사용됩니다:
1. 프로퍼티 파일의 값 주입
2. 환경 변수 값 주입
3. 시스템 프로퍼티 값 주입
4. SpEL(Spring Expression Language) 표현식 평가
5. 기본값 설정
6. 런타임 시 동적으로 값 결정

```java
@Component
public class MyComponent {

    // 기본 사용법
    @Value("${app.name}")
    private String appName;

    // 기본값 설정
    @Value("${app.description:Default Description}")
    private String appDescription;

    // SpEL 사용
    @Value("#{systemProperties['user.home']}")
    private String userHome;

    // 환경 변수 사용
    @Value("${JAVA_HOME}")
    private String javaHome;

    // 다른 빈의 프로퍼티 참조
    @Value("#{anotherBean.someProperty}")
    private String someProperty;

    // 컬렉션 주입
    @Value("${app.servers}")
    private List<String> servers;

    // 생성자 주입
    private final int maxConnections;

    public MyComponent(@Value("${app.max.connections:100}") int maxConnections) {
        this.maxConnections = maxConnections;
    }

    // 메서드 파라미터 주입
    @Autowired
    public void setDatabaseUrl(@Value("${db.url}") String url) {
        // ...
    }

    // SpEL을 사용한 조건부 값
    @Value("#{systemProperties['os.name'].toLowerCase().contains('windows') ? 'C:/temp' : '/tmp'}")
    private String tempDir;

    // 정규 표현식 사용
    @Value("#{'${csv.values}'.split(',')}")
    private List<String> csvValues;

    // 메서드 호출 결과 주입
    @Value("#{T(java.lang.Math).random()}")
    private double randomValue;
}
```

`@Value`의 주요 특징:
1. 프로퍼티 플레이스홀더: `${property.name}`
2. SpEL 표현식: `#{expression}`
3. 기본값 설정: `${property.name:defaultValue}`
4. 복합 표현식: `${property.name}${another.property}`

`@Value`의 내부 동작 방식:
1. 빈 생성: Spring 컨테이너가 빈을 생성합니다.
2. `@Value` 탐지: `BeanPostProcessor`가 `@Value` 어노테이션을 탐지합니다.
3. 표현식 파싱: 어노테이션의 값을 파싱하여 프로퍼티 플레이스홀더나 SpEL 표현식을 식별합니다.
4. 값 해석: `PropertySourcesPlaceholderConfigurer`나 SpEL 평가기를 사용하여 실제 값을 해석합니다.
5. 타입 변환: 필요한 경우 해석된 값을 대상 필드나 파라미터의 타입으로 변환합니다.
6. 값 주입: 해석 및 변환된 값을 대상 필드나 파라미터에 주입합니다.

```java
                [Spring 컨테이너 빈 생성]
                         |
                         v
                [@Value 어노테이션 탐지]
                         |
                         v
                    [표현식 파싱]
                         |
                         v
      +------------------+--------------------+
      |                                       |
      v                                       v
[프로퍼티 플레이스홀더]                        [SpEL 표현식]
      |                                       |
      v                                       v
[PropertySourcesPlaceholderConfigurer]   [SpEL 평가기]
      |                                       |
      +------------------+--------------------+
                         |
                         v
                      [값 해석]
                         |
                         v
                     [타입 변환]
                         |
                         v
                      [값 주입]
                         |
                         v
                    [빈 초기화 완료]
```

1. 프로퍼티 소스 설정:

    `@PropertySource`를 사용하여 프로퍼티 파일을 명시적으로 로드할 수 있습니다.

    ```java
    @Configuration
    @PropertySource("classpath:application.properties")
    public class AppConfig {
        // ...
    }
    ```

2. 복잡한 타입 변환:

    `CustomEditorConfigurer`나 `ConversionService`를 사용하여 복잡한 타입 변환을 처리할 수 있습니다.

    ```java
    @Configuration
    public class ConversionConfig {
        @Bean
        public ConversionService conversionService() {
            DefaultFormattingConversionService conversionService = new DefaultFormattingConversionService();
            conversionService.addConverter(new StringToMyCustomTypeConverter());
            return conversionService;
        }
    }
    ```

3. 프로파일별 값:

    Spring 프로파일과 함께 사용하여 환경별로 다른 값을 주입할 수 있습니다.

    ```properties
    # application-dev.properties
    app.url=http://dev.example.com

    # application-prod.properties
    app.url=http://prod.example.com
    ```

    ```java
    @Value("${app.url}")
    private String appUrl;
    ```

4. (Spring Cloud Config 사용 시) 런타임 재로딩:

    `@RefreshScope`와 함께 사용하여 런타임에 프로퍼티 값을 갱신할 수 있습니다

    ```java
    @RefreshScope
    @Component
    public class MyComponent {
        @Value("${dynamic.property}")
        private String dynamicProperty;
    }
    ```

5. 리스트나 맵 주입:

    쉼표로 구분된 값을 리스트나 맵으로 주입할 수 있습니다.

    ```properties
    app.list=value1,value2,value3
    app.map={key1:'value1',key2:'value2'}
    ```

    ```java
    @Value("#{'${app.list}'.split(',')}")
    private List<String> appList;

    @Value("#{${app.map}}")
    private Map<String, String> appMap;
    ```

6. 유효성 검사:

    `@Validated`와 함께 사용하여 주입된 값의 유효성을 검사할 수 있습니다.

    ```java
    @Validated
    @Component
    public class MyComponent {
        @Value("${app.port:8080}")
        @Min(1024)
        @Max(65535)
        private int port;
    }
    ```

7. 조건부 빈 생성:

    `@Value`를 `@Conditional`과 함께 사용하여 조건부로 빈을 생성할 수 있습니다.

    ```java
    @Bean
    @Conditional(PropertyExistsCondition.class)
    public MyService myService(@Value("${my.property}") String propertyValue) {
        return new MyServiceImpl(propertyValue);
    }
    ```

## 트랜잭션 관련 어노테이션

### [`@Transactional`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/transaction/annotation/Transactional.html)

`@Transactional`은 Spring 프레임워크에서 선언적 트랜잭션 관리를 위해 사용되는 어노테이션입니다.
이 어노테이션을 메서드나 클래스에 적용하면, 해당 메서드 실행 시 트랜잭션 경계를 자동으로 설정하고 관리합니다.

`@Transactional`은 주로 다음과 같은 경우에 사용됩니다:
1. 데이터베이스 작업의 원자성 보장
2. 여러 데이터베이스 작업을 하나의 트랜잭션으로 묶기
3. 트랜잭션 전파 행위 설정
4. 트랜잭션 격리 수준 설정
5. 읽기 전용 트랜잭션 설정
6. 트랜잭션 타임아웃 설정
7. 특정 예외 발생 시 롤백 설정

```java
@Service
public class UserService {

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private EmailService emailService;

    // 기본 사용법
    @Transactional
    public void createUser(User user) {
        userRepository.save(user);
        emailService.sendWelcomeEmail(user.getEmail());
    }

    // 읽기 전용 트랜잭션
    @Transactional(readOnly = true)
    public User getUserById(Long id) {
        return userRepository.findById(id).orElse(null);
    }

    // 트랜잭션 전파 행위 설정
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void updateUserWithNewTransaction(User user) {
        userRepository.save(user);
    }

    // 특정 예외에 대해 롤백하지 않음
    @Transactional(noRollbackFor = NotFoundException.class)
    public void deleteUser(Long id) {
        User user = userRepository.findById(id)
            .orElseThrow(() -> new NotFoundException("User not found"));
        userRepository.delete(user);
    }

    // 트랜잭션 격리 수준 설정
    @Transactional(isolation = Isolation.SERIALIZABLE)
    public void transferMoney(Long fromId, Long toId, BigDecimal amount) {
        // 송금 로직
    }

    // 트랜잭션 타임아웃 설정
    @Transactional(timeout = 30)
    public void longRunningOperation() {
        // 장시간 실행되는 작업
    }
}
```

`@Transactional`의 주요 속성:
1. `propagation`: 트랜잭션 전파 행위를 설정합니다.
2. `isolation`: 트랜잭션 격리 수준을 설정합니다.
3. `readOnly`: 읽기 전용 트랜잭션 여부를 설정합니다.
4. `timeout`: 트랜잭션 타임아웃 시간을 설정합니다.
5. `rollbackFor`: 특정 예외 발생 시 롤백할 예외 클래스를 지정합니다.
6. `noRollbackFor`: 특정 예외 발생 시 롤백하지 않을 예외 클래스를 지정합니다.

`@Transactional`의 내부 동작 방식:
1. 프록시 생성: Spring은 `@Transactional`이 적용된 클래스나 인터페이스에 대해 프록시를 생성합니다.
2. 메서드 호출 인터셉트: 프록시는 `@Transactional`이 적용된 메서드 호출을 인터셉트합니다.
3. 트랜잭션 시작: `TransactionManager`를 사용하여 트랜잭션을 시작합니다.
4. 비즈니스 로직 실행: 실제 메서드의 비즈니스 로직을 실행합니다.
5. 트랜잭션 완료: 비즈니스 로직 실행 결과에 따라 트랜잭션을 커밋하거나 롤백합니다.
6. 예외 처리: 예외 발생 시 롤백 여부를 결정하고 처리합니다.

```java
    [메서드 호출]
         |
         v
[트랜잭션 프록시 인터셉트]
         |
         v
  [트랜잭션 속성 확인]
         |
         v
   [트랜잭션 시작]
         |
         v
  [비즈니스 로직 실행]
    +----+------+
    |           |
    v           v
[정상 완료]   [예외 발생]
    |           |
    v           v
[트랜잭션 커밋] [롤백 여부 결정]
  |               |
  |          +----+----+
  |          |         |
  |          v         v
  |         [롤백]    [커밋]
  |          |         |
  +----+----+----+-----+
       |
       v
[메서드 반환 또는 예외 전파]
```

1. 내부 메서드 호출:

    같은 클래스 내에서 `@Transactional` 메서드가 같은 클래스의 내부 메서드를 호출할 때, 트랜잭션이 적용되지 않습니다.
    이를 해결하려면 self-invocation을 피하거나 `AopContext.currentProxy()`를 사용해야 합니다.

    ```java
    @Service
    public class UserService {
        @Transactional
        public void createUser(User user) {
            // 트랜잭션이 적용됨
        }

        public void someMethod() {
            createUser(new User()); // 트랜잭션이 적용되지 않음!
        }
    }
    ```

2. 인터페이스 vs 클래스 레벨 어노테이션:

    인터페이스에 `@Transactional`을 적용하면 JDK 동적 프록시가 사용되고,
    클래스에 적용하면 CGLIB 프록시가 사용됩니다.
    클래스 레벨 적용을 권장한다고 합니다.

3. `public` 메서드:

    `@Transactional`은 기본적으로 `public` 메서드에만 적용됩니다.
    `protected`, `private` 메서드에는 적용되지 않습니다.

4. 트랜잭션 전파:

    중첩된 `@Transactional` 메서드 호출 시 트랜잭션 전파 행위를 이해하고 적절히 설정해야 합니다.

    ```java
    @Service
    public class OrderService {
        @Autowired
        private PaymentService paymentService;

        @Transactional
        public void placeOrder(Order order) {
            // 주문 처리 로직
            paymentService.processPayment(order); // 새로운 트랜잭션에서 실행하고 싶다면?
        }
    }

    @Service
    public class PaymentService {
        @Transactional(propagation = Propagation.REQUIRES_NEW)
        public void processPayment(Order order) {
            // 결제 처리 로직
        }
    }
    ```

5. 롤백 규칙:

    기본적으로 런타임 예외(unchecked exceptions)에 대해서만 롤백합니다.
    체크 예외에 대해 롤백하려면 명시적으로 설정해야 합니다.

    ```java
    @Transactional(rollbackFor = Exception.class)
    public void methodThatMightThrowCheckedException() throws SomeCheckedException {
        // ...
    }
    ```

6. 읽기 전용 트랜잭션:

    읽기 전용 작업에는 `readOnly = true`를 사용하여 성능을 최적화할 수 있습니다.

    ```java
    @Transactional(readOnly = true)
    public List<User> getAllUsers() {
        return userRepository.findAll();
    }
    ```

7. 트랜잭션 격리 수준:

    데이터베이스의 기본 격리 수준을 변경하려면 `isolation` 속성을 사용합니다.
    그러나 이는 성능에 영향을 줄 수 있습니다.

8. 테스트에서의 사용:

    테스트 클래스나 메서드에 `@Transactional`을 적용하면, 각 테스트 후에 자동으로 롤백됩니다.

    ```java
    @RunWith(SpringRunner.class)
    @SpringBootTest
    @Transactional
    public class UserServiceTest {
        // 각 테스트 메서드 실행 후 롤백됨
    }
    ```

9. 비동기 메서드:

    `@Async`와 함께 사용할 때 주의가 필요합니다.
    비동기 메서드에서는 트랜잭션이 예상대로 동작하지 않을 수 있습니다.

## 웹 요청 처리 관련 어노테이션

### [`@RequestMapping`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/RequestMapping.html)

`@RequestMapping`은 Spring MVC 및 Spring WebFlux에서 HTTP 요청을 특정 핸들러 메서드에 매핑하는 데 사용됩니다.
클래스 및 메서드 레벨 모두에서 사용할 수 있습니다.
URL 패턴, HTTP 메서드, 요청/응답 콘텐츠 타입 등 다양한 조건을 지정할 수 있습니다.

`@RequestMapping`은 주로 다음과 같은 경우에 사용됩니다:
1. URL 경로를 특정 컨트롤러 또는 핸들러 메서드에 매핑
2. HTTP 메서드(GET, POST, PUT, DELETE 등)에 따른 요청 처리
3. 요청 파라미터, 헤더, 미디어 타입 등에 따른 세부적인 매핑 설정
4. RESTful API 설계 및 구현
5. 정적 리소스 핸들링 설정

```java
// 클래스 레벨에서의 기본 사용
@Controller
@RequestMapping("/users")
public class UserController {

    // GET /users
    @RequestMapping(method = RequestMethod.GET)
    public String listUsers(Model model) {
        // 사용자 목록을 모델에 추가
        return "userList";
    }

    // POST /users
    @RequestMapping(method = RequestMethod.POST)
    public String createUser(@ModelAttribute User user) {
        // 사용자 생성 로직
        return "redirect:/users";
    }

    // GET /users/{id}
    @RequestMapping(value = "/{id}", method = RequestMethod.GET)
    public String getUser(@PathVariable Long id, Model model) {
        // 특정 사용자 정보를 모델에 추가
        return "userDetails";
    }
}

// 메서드 레벨에서의 다양한 속성 사용
@RestController
@RequestMapping("/api")
public class ApiController {

    // GET /api/data?category=books
    @RequestMapping(value = "/data", method = RequestMethod.GET, params = "category=books")
    public List<Book> getBookData() {
        // 책 데이터 반환
    }

    // POST /api/users, Content-Type: application/json
    @RequestMapping(value = "/users", method = RequestMethod.POST, consumes = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<User> createUser(@RequestBody User user) {
        // JSON 형식의 사용자 데이터로 사용자 생성
    }

    // GET /api/reports, Accept: application/pdf
    @RequestMapping(value = "/reports", method = RequestMethod.GET, produces = MediaType.APPLICATION_PDF_VALUE)
    public ResponseEntity<byte[]> getReportAsPdf() {
        // PDF 형식의 보고서 생성 및 반환
    }
}

// 정규 표현식을 사용한 URL 패턴 매핑
@Controller
public class RegexController {

    // /files/2023/01/example.jpg, /files/2023/12/sample.png 등과 매칭
    @RequestMapping("/files/{year:\\d{4}}/{month:\\d{2}}/{file:[a-z]+\\.(?:jpg|png)}")
    public String handleFileRequest(@PathVariable int year, @PathVariable int month, @PathVariable String file) {
        // 파일 처리 로직
    }
}
```

`@RequestMapping`의 주요 속성:
1. `value` 또는 `path`: URL 패턴을 지정합니다.
2. `method`: 허용되는 HTTP 메서드를 지정합니다.
3. `params`: 특정 요청 파라미터의 존재 여부나 값을 확인합니다.
4. `headers`: 특정 요청 헤더의 존재 여부나 값을 확인합니다.
5. `consumes`: 요청의 Content-Type을 제한합니다.
6. `produces`: 응답의 Accept 타입을 제한합니다.

`@RequestMapping`의 내부 동작 방식:
1. 애플리케이션 시작 시 `RequestMappingHandlerMapping` 빈이 생성됩니다.
2. 이 빈은 `@RequestMapping`이 적용된 모든 핸들러 메서드를 스캔합니다.
3. URL 패턴, HTTP 메서드, 파라미터 등의 정보를 추출하여 내부 레지스트리에 저장합니다.
4. 요청이 들어오면 `DispatcherServlet`이 `RequestMappingHandlerMapping`을 통해 적절한 핸들러를 찾습니다.
5. 매칭되는 핸들러가 있으면 `RequestMappingHandlerAdapter`를 통해 해당 메서드를 실행합니다.

```java
[애플리케이션 시작]
         |
         v
[RequestMappingHandlerMapping 빈 생성]
         |
         v
[@RequestMapping 어노테이션 스캔]
         |
         v
[핸들러 메서드 정보 추출 및 저장]
         |
         v
[클라이언트 요청 수신]
         |
         v
[DispatcherServlet이 요청 처리]
         |
         v
[RequestMappingHandlerMapping에 핸들러 조회]
         |
    +----+----+
    |         |
    v         v
[매칭됨]   [매칭 안됨]
    |         |
    v         v
[핸들러 반환] [404 에러]
    |
    v
[RequestMappingHandlerAdapter가 핸들러 실행]
    |
    v
[요청 처리 및 응답 반환]
```

`@RequestMapping`의 고급 기능 및 팁:

1. HTTP 메서드별 단축 어노테이션:

    `@GetMapping`, `@PostMapping`, `@PutMapping`, `@DeleteMapping`, `@PatchMapping` 등을 사용하여 코드를 더 간결하게 만들 수 있습니다.

    ```java
    @GetMapping("/{id}")
    public User getUser(@PathVariable Long id) {
        // 사용자 조회 로직
    }
    ```

2. 다중 URL 패턴: 하나의 핸들러 메서드에 여러 URL 패턴을 매핑할 수 있습니다.

    ```java
    @RequestMapping({"/hello", "/greeting"})
    public String hello() {
        return "Hello, World!";
    }
    ```

3. 와일드카드 및 정규 표현식: URL 패턴에 와일드카드(`*`)나 정규 표현식을 사용하여 유연한 매핑이 가능합니다.

    ```java
    @GetMapping("/users/**")
    public String handleUserRequests() {
        // /users로 시작하는 모든 요청 처리
    }
    ```

4. 매트릭스 변수 지원: 세미콜론으로 구분된 키-값 쌍을 처리할 수 있습니다.

    ```java
    @GetMapping("/owners/{ownerId}/pets/{petId}")
    public Pet findPet(
        @PathVariable String ownerId,
        @PathVariable String petId,
        @MatrixVariable int q
    ) {
        // /owners/42;q=11/pets/21 형태의 URL 처리
    }
    ```

5. 커스텀 어노테이션:`@RequestMapping`을 메타 어노테이션으로 사용하여 커스텀 어노테이션을 만들 수 있습니다.

    ```java
    @Target({ElementType.METHOD})
    @Retention(RetentionPolicy.RUNTIME)
    @RequestMapping(method = RequestMethod.GET, produces = MediaType.APPLICATION_JSON_VALUE)
    public @interface JsonGetMapping {
        @AliasFor(annotation = RequestMapping.class, attribute = "value")
        String[] value() default {};
    }
    ```

### [`@GetMapping`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/GetMapping.html), [`@PostMapping`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/PostMapping.html), [`@PutMapping`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/PutMapping.html), [`@DeleteMapping`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/DeleteMapping.html)

이 어노테이션들은 Spring Framework 4.3에서 도입된 `@RequestMapping`의 특화된 버전입니다.
각각 HTTP `GET`, `POST`, `PUT`, `DELETE` 메서드에 대한 요청을 처리하는 핸들러 메서드를 지정하는 데 사용됩니다.
이들은 `@RequestMapping(method = RequestMethod.XXX)`의 축약형으로, 코드를 더 간결하고 명확하게 만듭니다.

1. `@GetMapping`: 데이터 조회
2. `@PostMapping`: 새 리소스 생성
3. `@PutMapping`: 기존 리소스 전체 수정
4. `@DeleteMapping`: 리소스 삭제

```java
@RestController
@RequestMapping("/api/users")
public class UserController {

    @Autowired
    private UserService userService;

    // GET /api/users
    @GetMapping
    public List<User> getAllUsers() {
        return userService.getAllUsers();
    }

    // GET /api/users/{id}
    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable Long id) {
        return userService.getUser(id)
                          .map(ResponseEntity::ok)
                          .orElse(ResponseEntity.notFound().build());
    }

    // POST /api/users
    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody @Valid User user) {
        User createdUser = userService.createUser(user);
        URI location = ServletUriComponentsBuilder
            .fromCurrentRequest()
            .path("/{id}")
            .buildAndExpand(createdUser.getId())
            .toUri();
        return ResponseEntity.created(location).body(createdUser);
    }

    // PUT /api/users/{id}
    @PutMapping("/{id}")
    public ResponseEntity<User> updateUser(@PathVariable Long id, @RequestBody @Valid User user) {
        return userService.updateUser(id, user)
                          .map(ResponseEntity::ok)
                          .orElse(ResponseEntity.notFound().build());
    }

    // DELETE /api/users/{id}
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        if (userService.deleteUser(id)) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
```

주요 속성:
1. `value` 또는 `path`: URL 패턴을 지정합니다.
2. `params`: 특정 요청 파라미터의 존재 여부나 값을 확인합니다.
3. `headers`: 특정 요청 헤더의 존재 여부나 값을 확인합니다.
4. `consumes`: 요청의 Content-Type을 제한합니다.
5. `produces`: 응답의 Accept 타입을 제한합니다.

```java
// params 속성 사용 예
@GetMapping(value = "/search", params = "query")
public List<Product> searchProducts(@RequestParam String query) {
    // query 파라미터가 있는 요청만 처리
}

// headers 속성 사용 예
@PostMapping(value = "/upload", headers = "content-type=multipart/*")
public String handleFileUpload(@RequestParam("file") MultipartFile file) {
    // 멀티파트 요청만 처리
}

// consumes 속성 사용 예
@PostMapping(value = "/users", consumes = MediaType.APPLICATION_JSON_VALUE)
public ResponseEntity<User> createUser(@RequestBody User user) {
    // JSON 형식의 요청 본문만 처리
}

// produces 속성 사용 예
@GetMapping(value = "/report", produces = MediaType.APPLICATION_PDF_VALUE)
public ResponseEntity<byte[]> getReportAsPdf() {
    // PDF 형식의 응답 생성
}
```

내부 동작 방식:
1. 애플리케이션 시작 시 `RequestMappingHandlerMapping` 빈이 생성됩니다.
2. 이 빈은 `@GetMapping`, `@PostMapping` 등이 적용된 모든 핸들러 메서드를 스캔합니다.
3. 각 어노테이션의 정보를 `@RequestMapping`으로 변환하여 내부 레지스트리에 저장합니다.
4. 요청이 들어오면 `DispatcherServlet`이 `RequestMappingHandlerMapping`을 통해 적절한 핸들러를 찾습니다.
5. 매칭되는 핸들러가 있으면 `RequestMappingHandlerAdapter`를 통해 해당 메서드를 실행합니다.

```java
[애플리케이션 시작]
         |
         v
[RequestMappingHandlerMapping 빈 생성]
         |
         v
[@GetMapping, @PostMapping 등 스캔]
         |
         v
[@RequestMapping으로 변환 및 정보 저장]
         |
         v
[클라이언트 HTTP 요청 수신]
         |
         v
[DispatcherServlet이 요청 처리]
         |
         v
[RequestMappingHandlerMapping에 핸들러 조회]
         |
    +----+----+
    |         |
    v         v
[매칭됨]   [매칭 안됨]
    |         |
    v         v
[핸들러 반환] [404 에러]
    |
    v
[RequestMappingHandlerAdapter가 핸들러 실행]
    |
    v
[요청 처리 및 응답 반환]
```

1. 정규 표현식을 이용한 URL 패턴 매칭:

    ```java
    @GetMapping("/users/{userId:\\d+}")
    public User getUser(@PathVariable Long userId) {
        // 숫자로만 이루어진 userId에 대해서만 매칭
    }
    ```

2. 다중 URL 패턴:

    ```java
    @GetMapping({"/hello", "/greeting"})
    public String hello() {
        return "Hello, World!";
    }
    ```

3. 요청 파라미터 필수 여부 지정:

    ```java
    @GetMapping(value = "/products", params = "category")
    public List<Product> getProductsByCategory(@RequestParam String category) {
        // category 파라미터가 반드시 있어야 함
    }
    ```

4. HTTP 헤더를 이용한 버전 관리:

    ```java
    @GetMapping(value = "/api/users", headers = "X-API-VERSION=1")
    public List<UserV1> getUsersV1() {
        // API 버전 1
    }

    @GetMapping(value = "/api/users", headers = "X-API-VERSION=2")
    public List<UserV2> getUsersV2() {
        // API 버전 2
    }
    ```

5. 응답 미디어 타입 지정:

    ```java
    @GetMapping(value = "/report", produces = { MediaType.APPLICATION_PDF_VALUE, MediaType.APPLICATION_XML_VALUE })
    public ResponseEntity<?> getReport(@RequestHeader("Accept") String acceptHeader) {
        // Accept 헤더에 따라 PDF 또는 XML 형식으로 응답
    }
    ```

### [`@PathVariable`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/PathVariable.html)

`@PathVariable`은 Spring MVC에서 URL 경로의 일부를 메서드 파라미터로 바인딩하는 데 사용되는 어노테이션입니다.
이 어노테이션을 사용하면 동적 URL을 처리하고 RESTful API를 쉽게 구현할 수 있습니다.

`@PathVariable`은 주로 다음과 같은 경우에 사용됩니다:
1. RESTful API에서 리소스 식별자 추출
2. 동적 URL 처리
3. 계층적 리소스 구조 표현
4. URL을 통한 데이터 전달

```java
@RestController
@RequestMapping("/api")
public class UserController {

    // 기본 사용법
    @GetMapping("/users/{id}")
    public User getUser(@PathVariable Long id) {
        // id를 이용해 사용자 정보 조회
        return userService.findById(id);
    }

    // 여러 개의 Path Variable 사용
    @GetMapping("/users/{userId}/posts/{postId}")
    public Post getUserPost(
        @PathVariable Long userId,
        @PathVariable Long postId
    ) {
        // userId와 postId를 이용해 특정 사용자의 특정 게시글 조회
        return postService.findUserPost(userId, postId);
    }

    // Path Variable 이름 명시적 지정
    @GetMapping("/products/{productId}")
    public Product getProduct(@PathVariable("productId") Long id) {
        // productId라는 이름의 Path Variable을 id 파라미터에 바인딩
        return productService.findById(id);
    }

    // 옵셔널 Path Variable
    @GetMapping({"/categories", "/categories/{categoryId}"})
    public List<Category> getCategories(@PathVariable(required = false) Long categoryId) {
        if (categoryId != null) {
            // 특정 카테고리 조회
            return categoryService.findById(categoryId);
        } else {
            // 모든 카테고리 조회
            return categoryService.findAll();
        }
    }

    // 정규표현식을 이용한 Path Variable 제한
    @GetMapping("/files/{fileName:.+}")
    public ResponseEntity<Resource> getFile(@PathVariable String fileName) {
        // .을 포함한 파일명 처리 (예: "document.pdf")
        Resource file = storageService.loadAsResource(fileName);
        return ResponseEntity.ok().body(file);
    }

    // Map을 이용한 다중 Path Variable 처리
    @GetMapping("/search/{category}/{subcategory}")
    public List<Product> searchProducts(@PathVariable Map<String, String> pathVars) {
        String category = pathVars.get("category");
        String subcategory = pathVars.get("subcategory");
        // category와 subcategory를 이용한 검색 로직
        return productService.search(category, subcategory);
    }
}
```

`@PathVariable`의 주요 속성:
1. `name` 또는 `value`: Path Variable의 이름을 지정합니다. 생략 시 파라미터 이름이 사용됩니다.
2. `required`: Path Variable이 필수인지 여부를 지정합니다. 기본값은 true입니다.

`@PathVariable`의 내부 동작 방식:
1. 요청 수신: DispatcherServlet이 클라이언트로부터 HTTP 요청을 받습니다.
2. 핸들러 매핑: RequestMappingHandlerMapping이 요청 URL과 매칭되는 핸들러 메서드를 찾습니다.
3. Path Variable 추출: URL 템플릿에서 정의된 위치의 값을 추출합니다.
4. 데이터 변환: 필요한 경우 문자열 값을 적절한 타입(예: Long, Integer)으로 변환합니다.
5. 파라미터 바인딩: 변환된 값을 `@PathVariable`이 지정된 메서드 파라미터에 바인딩합니다.
6. 메서드 실행: 바인딩된 값을 이용해 핸들러 메서드를 실행합니다.

```java
[클라이언트 HTTP 요청]
         |
         v
[DispatcherServlet]
         |
         v
[RequestMappingHandlerMapping]
         |
         v
[URL 패턴 매칭]
         |
         v
[Path Variable 값 추출]
         |
         v
[데이터 타입 변환]
    +----+----+
    |         |
    v         v
[변환 성공] [변환 실패]
    |         |
    v         v
[파라미터 바인딩] [400 Bad Request]
    |
    v
[핸들러 메서드 실행]
    |
    v
[응답 생성 및 반환]
```

1. 타입 불일치 처리:

    Path Variable의 값이 지정된 타입으로 변환될 수 없는 경우, Spring은 기본적으로 400 Bad Request 오류를 반환합니다.
    이를 커스터마이즈하려면 `@ControllerAdvice`를 사용한 전역 예외 처리를 구현할 수 있습니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {

        @ExceptionHandler(MethodArgumentTypeMismatchException.class)
        public ResponseEntity<String> handleTypeMismatch(
            MethodArgumentTypeMismatchException ex
        ) {
            String name = ex.getName();
            String type = ex.getRequiredType().getSimpleName();
            Object value = ex.getValue();
            String message = String.format("'%s' should be a valid %s and '%s' isn't",
                                            name, type, value);
            return new ResponseEntity<>(message, HttpStatus.BAD_REQUEST);
        }
    }
    ```

2. 정규표현식 활용:

    URL 템플릿에서 정규표현식을 사용하여 Path Variable의 형식을 제한할 수 있습니다.

    ```java
    @GetMapping("/users/{id:[0-9]+}")
    public User getUser(@PathVariable Long id) {
        // 숫자로만 이루어진 id만 처리
    }
    ```

3. 인코딩 처리:

    URL에 특수 문자나 공백이 포함된 경우, 적절한 인코딩/디코딩이 필요할 수 있습니다.

    ```java
    @GetMapping("/products/{name}")
    public Product getProduct(@PathVariable String name) {
        String decodedName = URLDecoder.decode(name, StandardCharsets.UTF_8);
        return productService.findByName(decodedName);
    }
    ```

4. 보안 고려사항:

    Path Variable로 받은 값을 데이터베이스 쿼리나 파일 시스템 접근에 직접 사용할 때는 주의가 필요합니다.
    입력 값 검증과 적절한 이스케이핑을 통해 보안 취약점을 방지해야 합니다.

    ```java
    @GetMapping("/documents/{filename}")
    public ResponseEntity<Resource> getDocument(@PathVariable String filename) {
        // 파일 이름 검증
        if (!isValidFilename(filename)) {
            return ResponseEntity.badRequest().build();
        }
        // 안전한 방식으로 파일 접근
        Resource resource = safelyLoadResource(filename);
        return ResponseEntity.ok().body(resource);
    }
    ```

5. 복잡한 Path Variable 처리:

    여러 개의 Path Variable을 객체로 바인딩해야 할 경우,
    Spring MVC에서 컨트롤러 메서드의 파라미터를 해석하고 바인딩하는 인터페이스인 `HandlerMethodArgumentResolver`를 구현할 수 있습니다.
    커스텀 `HandlerMethodArgumentResolver`를 구현하면 복잡한 객체를 URL 경로 변수나 요청 파라미터로부터 직접 생성할 수 있습니다.

    1. 여러 경로 변수나 요청 파라미터를 하나의 복잡한 객체로 바인딩
    2. 특정 형식의 문자열을 파싱하여 사용자 정의 객체로 변환
    3. 요청의 여러 부분(헤더, 쿠키, 세션 등)에서 데이터를 조합하여 객체 생성

    ```java
    // 커스텀 객체
    public class CustomObject {
        private String path;
        private String param;
        // getters, setters
    }

    // 커스텀 리졸버
    public class CustomObjectResolver implements HandlerMethodArgumentResolver {

        // 이 리졸버가 처리할 수 있는 파라미터 타입 지정
        @Override
        public boolean supportsParameter(MethodParameter parameter) {
            return parameter.getParameterType().equals(CustomObject.class);
        }

        // 실제 객체 생성 로직 구현
        @Override
        public Object resolveArgument(
            MethodParameter parameter,
            ModelAndViewContainer mavContainer,
            NativeWebRequest webRequest,
            WebDataBinderFactory binderFactory
        ) throws Exception {
            String path = webRequest.getAttribute("org.springframework.web.servlet.HandlerMapping.uriTemplateVariables", RequestAttributes.SCOPE_REQUEST).get("path").toString();
            String param = webRequest.getParameter("param");

            CustomObject obj = new CustomObject();
            obj.setPath(path);
            obj.setParam(param);
            return obj;
        }
    }

    // `WebMvcConfigurer`를 구현한 설정 클래스에서 커스텀 리졸버 등록
    @Configuration
    public class WebConfig implements WebMvcConfigurer {
        @Override
        public void addArgumentResolvers(List<HandlerMethodArgumentResolver> resolvers) {
            resolvers.add(new CustomObjectResolver());
        }
    }

    // 컨트롤러에서 사용
    @GetMapping("/custom/{path}")
    public ResponseEntity<String> customMethod(@CustomObject CustomObject obj) {
        // CustomObject 사용
    }
    ```

### [`@RequestParam`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/RequestParam.html)

`@RequestParam`은 Spring MVC에서 HTTP 요청 파라미터를 컨트롤러 메서드의 파라미터에 바인딩하는 데 사용되는 어노테이션입니다.
이 어노테이션을 사용하면 쿼리 파라미터, 폼 데이터, 멀티파트 요청의 파라미터 등을 쉽게 처리할 수 있습니다.

`@RequestParam`은 주로 다음과 같은 경우에 사용됩니다:
1. URL 쿼리 파라미터 추출
2. HTML 폼 데이터 처리
3. 파일 업로드 처리
4. 선택적 파라미터 처리
5. 기본값 설정
6. 파라미터 타입 변환

```java
@RestController
@RequestMapping("/api")
public class ProductController {

    // 기본 사용법
    @GetMapping("/products")
    public List<Product> getProducts(@RequestParam String category) {
        return productService.getProductsByCategory(category);
    }

    // 필수가 아닌 파라미터
    @GetMapping("/search")
    public List<Product> searchProducts(@RequestParam(required = false) String keyword) {
        if (keyword != null) {
            return productService.searchProducts(keyword);
        }
        return productService.getAllProducts();
    }

    // 기본값 설정
    @GetMapping("/popular")
    public List<Product> getPopularProducts(@RequestParam(defaultValue = "10") int limit) {
        return productService.getPopularProducts(limit);
    }

    // 이름 명시적 지정
    @GetMapping("/filter")
    public List<Product> filterProducts(@RequestParam("min_price") double minPrice,
                                        @RequestParam("max_price") double maxPrice) {
        return productService.filterByPrice(minPrice, maxPrice);
    }

    // 멀티값 파라미터
    @GetMapping("/by-ids")
    public List<Product> getProductsByIds(@RequestParam List<Long> ids) {
        return productService.getProductsByIds(ids);
    }

    // Map으로 모든 파라미터 받기
    @PostMapping("/create")
    public Product createProduct(@RequestParam Map<String, String> params) {
        return productService.createProduct(params);
    }

    // 파일 업로드
    @PostMapping("/upload")
    public String handleFileUpload(@RequestParam("file") MultipartFile file) {
        // 파일 처리 로직
        return "File uploaded successfully: " + file.getOriginalFilename();
    }
}
```

`@RequestParam`의 주요 속성:
1. `name` 또는 `value`: 파라미터의 이름을 지정합니다. 생략 시 메서드 파라미터 이름이 사용됩니다.
2. `required`: 파라미터가 필수인지 여부를 지정합니다. 기본값은 true입니다.
3. `defaultValue`: 파라미터가 제공되지 않았을 때 사용할 기본값을 지정합니다.

`@RequestParam`의 내부 동작 방식:
1. 요청 수신: `DispatcherServlet`이 클라이언트로부터 HTTP 요청을 받습니다.
2. 핸들러 매핑: `RequestMappingHandlerMapping`이 요청 URL과 매칭되는 핸들러 메서드를 찾습니다.
3. 파라미터 추출: `HttpServletRequest`에서 지정된 이름의 파라미터 값을 추출합니다.
4. 데이터 변환: 필요한 경우 문자열 값을 적절한 타입(예: int, double, List)으로 변환합니다.
5. 파라미터 바인딩: 변환된 값을 `@RequestParam`이 지정된 메서드 파라미터에 바인딩합니다.
6. 메서드 실행: 바인딩된 값을 이용해 핸들러 메서드를 실행합니다.

```java
[클라이언트 HTTP 요청]
         |
         v
[DispatcherServlet]
         |
         v
[RequestMappingHandlerMapping]
         |
         v
[핸들러 메서드 식별]
         |
         v
[파라미터 값 추출]
         |
         v
[데이터 타입 변환]
    +----+----+
    |         |
    v         v
[변환 성공] [변환 실패]
    |         |
    v         v
[파라미터 바인딩] [400 Bad Request]
    |
    v
[핸들러 메서드 실행]
    |
    v
[응답 생성 및 반환]
```

1. 필수 파라미터 처리:

    `required=true`(기본값)인 파라미터가 요청에 없으면 Spring은 `MissingServletRequestParameterException`을 발생시킵니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {
        @ExceptionHandler(MissingServletRequestParameterException.class)
        public ResponseEntity<String> handleMissingParams(MissingServletRequestParameterException ex) {
            String name = ex.getParameterName();
            return new ResponseEntity<>(name + " parameter is missing", HttpStatus.BAD_REQUEST);
        }
    }
    ```

2. 타입 변환 오류 처리:

    파라미터 값을 지정된 타입으로 변환할 수 없는 경우, `MethodArgumentTypeMismatchException`이 발생합니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {
        @ExceptionHandler(MethodArgumentTypeMismatchException.class)
        public ResponseEntity<String> handleTypeMismatch(MethodArgumentTypeMismatchException ex) {
            String name = ex.getName();
            String type = ex.getRequiredType().getSimpleName();
            Object value = ex.getValue();
            String message = String.format("'%s' should be a valid %s and '%s' isn't",
                                            name, type, value);
            return new ResponseEntity<>(message, HttpStatus.BAD_REQUEST);
        }
    }
    ```

3. 멀티값 파라미터 처리:

    같은 이름의 파라미터가 여러 번 전달되는 경우, List나 배열로 받을 수 있습니다.

    ```java
    @GetMapping("/items")
    public List<Item> getItems(@RequestParam List<String> categories) {
        return itemService.getItemsByCategories(categories);
    }
    ```

4. 파라미터 이름 일치:

    메서드 파라미터 이름과 요청 파라미터 이름이 다를 경우, `@RequestParam`의 `name` 속성을 사용합니다.

    ```java
    @GetMapping("/products")
    public List<Product> getProducts(
        @RequestParam(name = "category_id") Long categoryId
    ) {
        return productService.getProductsByCategoryId(categoryId);
    }
    ```

5. 선택적 파라미터와 기본값:

    `required=false`와 `defaultValue`를 함께 사용하여 선택적 파라미터를 효과적으로 처리할 수 있습니다.

    ```java
    @GetMapping("/search")
    public List<Product> searchProducts(
        @RequestParam(required = false, defaultValue = "") String keyword,
        @RequestParam(required = false, defaultValue = "0") int page,
        @RequestParam(required = false, defaultValue = "10") int size
    ) {
        return productService.searchProducts(keyword, page, size);
    }
    ```

6. 보안 고려사항:

    `@RequestParam`으로 받은 값을 직접 SQL 쿼리나 명령어 실행에 사용하지 않도록 주의해야 합니다.
    항상 적절한 검증과 이스케이핑을 수행해야 합니다.

    ```java
    @GetMapping("/users")
    public List<User> searchUsers(@RequestParam String username) {
        // 잘못된 방법 (SQL Injection 취약점)
        // return jdbcTemplate.query("SELECT * FROM users WHERE username = '" + username + "'", ...);

        // 올바른 방법
        return userRepository.findByUsername(username);
    }
    ```

### [`@RequestBody`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/RequestBody.html)

`@RequestBody`는 Spring MVC에서 HTTP 요청의 본문(body)을 자바 객체로 변환하는 데 사용되는 어노테이션입니다.
주로 JSON이나 XML 형식의 데이터를 객체로 역직렬화할 때 사용됩니다.

`@RequestBody`는 주로 다음과 같은 경우에 사용됩니다:
1. RESTful API에서 JSON/XML 요청 본문 처리
2. 복잡한 객체 구조를 가진 데이터 수신
3. `POST`, `PUT`, `PATCH` 요청의 본문 데이터 처리
4. 커스텀 데이터 형식의 요청 본문 처리

```java
@RestController
@RequestMapping("/api")
public class UserController {

    @Autowired
    private UserService userService;

    // 기본 사용법
    @PostMapping("/users")
    public ResponseEntity<User> createUser(@RequestBody User user) {
        User createdUser = userService.createUser(user);
        return ResponseEntity.status(HttpStatus.CREATED).body(createdUser);
    }

    // 유효성 검사와 함께 사용
    @PostMapping("/register")
    public ResponseEntity<User> registerUser(
        @RequestBody @Valid UserRegistrationDto registrationDto
    ) {
        User user = userService.registerUser(registrationDto);
        return ResponseEntity.ok(user);
    }

    // 제네릭 타입과 함께 사용
    @PostMapping("/bulk-create")
    public ResponseEntity<List<User>> bulkCreateUsers(@RequestBody List<User> users) {
        List<User> createdUsers = userService.bulkCreateUsers(users);
        return ResponseEntity.ok(createdUsers);
    }

    // 중첩된 객체 구조 처리
    @PostMapping("/orders")
    public ResponseEntity<Order> createOrder(@RequestBody OrderRequest orderRequest) {
        Order createdOrder = orderService.createOrder(orderRequest);
        return ResponseEntity.ok(createdOrder);
    }

    // 커스텀 역직렬화와 함께 사용
    @PostMapping("/custom-data")
    public ResponseEntity<String> handleCustomData(
        @RequestBody @JsonDeserialize(using = CustomDataDeserializer.class) CustomData customData
    ) {
        String result = customDataService.processCustomData(customData);
        return ResponseEntity.ok(result);
    }
}

// 요청 DTO 예시
public class UserRegistrationDto {
    @NotBlank
    private String username;

    @Email
    private String email;

    @Size(min = 8, max = 20)
    private String password;

    // getters, setters
}

// 중첩된 객체 구조 예시
public class OrderRequest {
    private List<OrderItem> items;
    private ShippingAddress shippingAddress;
    private PaymentInfo paymentInfo;

    // getters, setters
}
```

`@RequestBody`의 내부 동작 방식:
1. 요청 수신: `DispatcherServlet`이 클라이언트로부터 HTTP 요청을 받습니다.
2. 핸들러 매핑: `RequestMappingHandlerMapping`이 요청 URL과 매칭되는 핸들러 메서드를 찾습니다.
3. 메시지 변환기 선택: `ContentNegotiationManager`가 요청의 `Content-Type`을 확인하고 적절한 `HttpMessageConverter`를 선택합니다.
4. 요청 본문 읽기 및 객체 변환: 선택된 `HttpMessageConverter`가 요청 본문을 읽고 자바 객체로 변환합니다.
5. 유효성 검사: `@Valid` 어노테이션이 사용된 경우, 변환된 객체에 대해 유효성 검사를 수행합니다.
6. 파라미터 바인딩: 변환된 객체를 `@RequestBody`가 지정된 메서드 파라미터에 바인딩합니다.
7. 메서드 실행: 바인딩된 객체를 이용해 핸들러 메서드를 실행합니다.

```java
[클라이언트 HTTP 요청]
         |
         v
[DispatcherServlet]
         |
         v
[RequestMappingHandlerMapping]
         |
         v
[핸들러 메서드 식별]
         |
         v
[ContentNegotiationManager]
         |
         v
[HttpMessageConverter 선택]
         |
         v
[요청 본문 읽기]
         |
         v
[객체 변환 (역직렬화)]
    +----+----+
    |         |
    v         v
[변환 성공] [변환 실패]
    |         |
    v         v
[유효성 검사] [400 Bad Request]
    |
    v
[핸들러 메서드 실행]
    |
    v
[응답 생성 및 반환]
```

1. `Content-Type` 헤더:

    클라이언트는 반드시 적절한 `Content-Type` 헤더(예: `application/json`)를 포함해야 합니다.
    그렇지 않으면 `415 Unsupported Media Type` 오류가 발생할 수 있습니다.

2. 역직렬화 오류 처리:

    요청 본문을 객체로 변환할 수 없는 경우 `HttpMessageNotReadableException`이 발생할 수 있습니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {
        @ExceptionHandler(HttpMessageNotReadableException.class)
        public ResponseEntity<String> handleHttpMessageNotReadable(
            HttpMessageNotReadableException ex
        ) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                                .body("Malformed JSON request: " + ex.getMessage());
        }
    }
    ```

3. 유효성 검사:

    `@Valid` 또는 `@Validated`와 함께 사용하여 입력 데이터의 유효성을 검사할 수 있습니다.
    `MethodArgumentNotValidException`을 처리해야 합니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {
        @ExceptionHandler(MethodArgumentNotValidException.class)
        public ResponseEntity<Map<String, String>> handleValidationException(
            MethodArgumentNotValidException ex
        ) {
            Map<String, String> errors = new HashMap<>();
            ex.getBindingResult().getAllErrors().forEach((error) -> {
                String fieldName = ((FieldError) error).getField();
                String errorMessage = error.getDefaultMessage();
                errors.put(fieldName, errorMessage);
            });
            return ResponseEntity.badRequest().body(errors);
        }
    }
    ```

4. 큰 요청 본문 처리:

    대용량 데이터를 처리할 때는 메모리 사용에 주의해야 합니다.
    스트리밍 방식의 처리를 고려하거나, 최대 요청 크기를 제한할 수 있습니다.

    ```properties
    # application.properties
    spring.servlet.multipart.max-file-size=10MB
    spring.servlet.multipart.max-request-size=10MB
    ```

5. 보안 고려사항:

    `@RequestBody`로 받은 데이터는 신뢰할 수 없는 입력으로 취급해야 합니다.
    적절한 입력 검증과 이스케이핑을 수행해야 합니다.

6. 커스텀 역직렬화:

    복잡한 데이터 구조나 특별한 형식의 데이터를 처리해야 할 경우,
    `Jackson`의 `@JsonDeserialize`를 사용하여 커스텀 역직렬화기를 구현할 수 있습니다.

    ```java
    public class CustomDataDeserializer extends JsonDeserializer<CustomData> {
        @Override
        public CustomData deserialize(
            JsonParser p,
            DeserializationContext ctxt
        ) throws IOException {
            // 커스텀 역직렬화 로직
        }
    }
    ```

7. Optional과 함께 사용:

    요청 본문이 선택적인 경우, `Optional<T>`와 함께 `@RequestBody`를 사용할 수 있습니다.

    ```java
    @PostMapping("/optional-data")
    public ResponseEntity<String> handleOptionalData(
        @RequestBody Optional<SomeData> data
    ) {
        return data.map(d -> ResponseEntity.ok("Data received: " + d))
                .orElse(ResponseEntity.ok("No data provided"));
    }
    ```

### [`@ResponseBody`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/bind/annotation/ResponseBody.html)

`@ResponseBody`는 Spring MVC에서 컨트롤러 메서드의 반환값을 HTTP 응답 본문으로 직접 전송하도록 지시하는 어노테이션입니다.
이 어노테이션을 사용하면 *View Resolver를 통한 뷰 처리 과정을 거치지 않고, 반환된 객체를 HTTP 응답으로 직접 변환*합니다.

`@ResponseBody`는 주로 다음과 같은 경우에 사용됩니다:
1. RESTful API 응답 데이터 생성
2. JSON, XML 등의 형식으로 데이터 반환
3. 비동기 요청에 대한 응답 처리
4. 커스텀 데이터 형식의 응답 생성

```java
@Controller
@RequestMapping("/api")
public class UserController {

    @Autowired
    private UserService userService;

    // 기본 사용법
    @GetMapping("/users/{id}")
    @ResponseBody
    public User getUser(@PathVariable Long id) {
        return userService.getUserById(id);
    }

    // 컬렉션 반환
    @GetMapping("/users")
    @ResponseBody
    public List<User> getAllUsers() {
        return userService.getAllUsers();
    }

    // 복잡한 객체 구조 반환
    @GetMapping("/dashboard")
    @ResponseBody
    public DashboardData getDashboardData() {
        return dashboardService.getDashboardData();
    }

    // 커스텀 직렬화와 함께 사용
    @GetMapping("/custom-data")
    @ResponseBody
    @JsonSerialize(using = CustomDataSerializer.class)
    public CustomData getCustomData() {
        return customDataService.getCustomData();
    }

    // ResponseEntity와 함께 사용
    @GetMapping("/users/{id}")
    public ResponseEntity<User> getUserWithResponseEntity(@PathVariable Long id) {
        User user = userService.getUserById(id);
        if (user != null) {
            return ResponseEntity.ok(user);
        } else {
            return ResponseEntity.notFound().build();
        }
    }
}

// 복잡한 객체 구조 예시
public class DashboardData {
    private List<ChartData> charts;
    private SummaryStats summaryStats;
    private List<RecentActivity> recentActivities;

    // getters, setters
}

// RestController 사용 시 @ResponseBody 생략 가능
@RestController
@RequestMapping("/api/v2")
public class UserRestController {

    @GetMapping("/users/{id}")
    public User getUser(@PathVariable Long id) {
        return userService.getUserById(id);
    }
}
```

`@ResponseBody`의 내부 동작 방식:
1. 메서드 실행: 컨트롤러의 핸들러 메서드가 실행되어 결과 객체를 반환합니다.
2. 반환 타입 확인: `RequestResponseBodyMethodProcessor`가 `@ResponseBody` 어노테이션을 감지합니다.
3. 메시지 변환기 선택: `ContentNegotiationManager`가 클라이언트의 `Accept` 헤더를 확인하고 적절한 `HttpMessageConverter`를 선택합니다.
4. 객체 변환: 선택된 `HttpMessageConverter`가 자바 객체를 지정된 미디어 타입(예: JSON, XML)으로 변환합니다.
5. 응답 생성: 변환된 데이터가 HTTP 응답 본문에 쓰여집니다.
6. 응답 전송: 생성된 HTTP 응답이 클라이언트에게 전송됩니다.

```java
[핸들러 메서드 실행]
         |
         v
[반환 객체 생성]
         |
         v
[@ResponseBody 감지]
         |
         v
[ContentNegotiationManager]
         |
         v
[HttpMessageConverter 선택]
         |
         v
[객체 변환 (직렬화)]
    +----+----+
    |         |
    v         v
[변환 성공] [변환 실패]
    |         |
    v         v
[HTTP 응답 생성] [500 Internal Server Error]
    |
    v
[클라이언트에 응답 전송]
```

1. Content-Type 설정:

    `@ResponseBody`는 기본적으로 클라이언트의 `Accept` 헤더를 기반으로 응답의 `Content-Type`을 결정합니다.
    필요한 경우 `produces` 속성을 사용하여 명시적으로 지정할 수 있습니다.

    ```java
    @GetMapping(value = "/users", produces = MediaType.APPLICATION_JSON_VALUE)
    @ResponseBody
    public List<User> getUsers() {
        return userService.getAllUsers();
    }
    ```

2. 직렬화 설정:

    객체의 특정 필드를 제외하거나 이름을 변경하려면 `Jackson` 어노테이션을 사용할 수 있습니다.

    ```java
    public class User {
        private Long id;
        private String username;

        @JsonIgnore
        private String password;

        @JsonProperty("fullName")
        private String name;

        // getters, setters
    }
    ```

3. 예외 처리:

    `@ResponseBody`를 사용할 때 발생하는 예외(예: 직렬화 오류)를 적절히 처리해야 합니다.

    ```java
    @ControllerAdvice
    public class GlobalExceptionHandler {
        @ExceptionHandler(Exception.class)
        @ResponseBody
        public ErrorResponse handleException(Exception ex) {
            return new ErrorResponse("An error occurred: " + ex.getMessage());
        }
    }
    ```

4. 대용량 데이터 처리:

    대량의 데이터를 반환할 때는 메모리 사용에 주의해야 합니다.
    스트리밍 방식의 응답이나 페이지네이션을 고려해볼 수 있습니다.

    ```java
    @GetMapping("/large-data")
    @ResponseBody
    public ResponseEntity<StreamingResponseBody> getLargeData() {
        StreamingResponseBody responseBody = outputStream -> {
            Writer writer = new BufferedWriter(new OutputStreamWriter(outputStream));
            for (int i = 0; i < 1000000; i++) {
                writer.write("Data line " + i + "\n");
                writer.flush();
            }
        };
        return ResponseEntity.ok()
                .contentType(MediaType.TEXT_PLAIN)
                .body(responseBody);
    }
    ```

5. 국제화(`i18n`) 지원:

    다국어 지원이 필요한 경우, 메시지 소스를 활용하여 동적으로 텍스트를 생성할 수 있습니다.

    ```java
    @Autowired
    private MessageSource messageSource;

    @GetMapping("/greeting")
    @ResponseBody
    public String getGreeting(Locale locale) {
        return messageSource.getMessage("greeting.message", null, locale);
    }
    ```

6. `HATEOAS` 지원:

    RESTful API에서 `HATEOAS`를 구현하려면 *Spring HATEOAS*를 사용할 수 있습니다.

    ```java
    @GetMapping("/users/{id}")
    @ResponseBody
    public EntityModel<User> getUser(@PathVariable Long id) {
        User user = userService.getUserById(id);
        return EntityModel.of(user,
            linkTo(methodOn(UserController.class).getUser(id)).withSelfRel(),
            linkTo(methodOn(UserController.class).getUserOrders(id)).withRel("orders")
        );
    }
    ```

7. 응답 압축:

    대용량 응답의 경우 GZIP 압축을 활성화하여 네트워크 대역폭을 절약할 수 있습니다.

    ```properties
    # application.properties
    server.compression.enabled=true
    server.compression.min-response-size=1024
    server.compression.mime-types=application/json,application/xml,text/html,text/plain
    ```
