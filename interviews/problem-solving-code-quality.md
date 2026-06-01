# 문제 해결, 코드 품질, 운영 실천

- [문제 해결, 코드 품질, 운영 실천](#문제-해결-코드-품질-운영-실천)
    - [먼저 기억할 정리](#먼저-기억할-정리)
    - [설계 원칙과 패턴](#설계-원칙과-패턴)
        - [프록시 패턴](#프록시-패턴)
            - [원문: 프록시 패턴](#원문-프록시-패턴)
                - [프록시 패턴](#프록시-패턴-1)
    - [알고리즘과 문제 해결](#알고리즘과-문제-해결)
        - [정규표현식](#정규표현식)
            - [원문: 정규표현식](#원문-정규표현식)
                - [정규표현식](#정규표현식-1)
    - [운영 실천과 배포 기반](#운영-실천과-배포-기반)
        - [Docker 컨테이너 가상화](#docker-컨테이너-가상화)
            - [원문: Docker 컨테이너 가상화](#원문-docker-컨테이너-가상화)
                - [Docker 컨테이너 가상화](#docker-컨테이너-가상화-1)
                    - [결론](#결론)
    - [테스트와 대역 객체](#테스트와-대역-객체)
        - [java, php에서 mock 생성 원리](#java-php에서-mock-생성-원리)
            - [원문: java, php에서 mock 생성 원리](#원문-java-php에서-mock-생성-원리)
                - [java, php에서 mock 생성 원리](#java-php에서-mock-생성-원리-1)
        - [테스트 더블](#테스트-더블)
            - [원문: 테스트 더블](#원문-테스트-더블)
                - [테스트 더블](#테스트-더블-1)

알고리즘, 복잡도, 테스트, mock, 패턴, Docker/배포처럼 문제를 풀고 코드를 신뢰 가능하게 만드는 실천 축을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## 먼저 기억할 정리

문제 해결과 코드 품질 문서는 "좋은 습관 목록"이 아니라 불확실한 동작을 검증 가능한 경계로 바꾸는 문서입니다. 알고리즘은 입력이 어떤 상태로 바뀌는지, 테스트는 기대 동작을 어떤 관측값으로 고정하는지, mock은 어떤 외부 의존의 호출과 결과를 대체하는지, Docker는 프로세스와 파일시스템과 네트워크 경계를 어떻게 격리하는지로 읽어야 합니다.

```text
problem input
  -> model / invariant
  -> implementation boundary
  -> test double or real dependency
  -> observable output / log / metric
  -> regression signal
```

비교축은 실제 객체와 대역 객체, 설계 패턴과 런타임 proxy, 개발 환경과 운영 환경입니다. Mock은 "가짜 객체"라는 말보다 어떤 호출을 기록하고 어떤 side effect를 막는지가 중요하고, proxy pattern은 중간 객체가 접근, lazy loading, logging, transaction 같은 정책을 어디서 가로채는지로 설명해야 합니다.

검증 anchor는 unit/integration test, mock interaction verification, benchmark, container inspect/log, deployment health check입니다. 코드 품질 답변은 추상 원칙만 말하지 말고 어떤 실패를 더 빨리 발견하거나 어떤 변경을 더 안전하게 만들었는지로 닫아야 합니다.

## 설계 원칙과 패턴

### 프록시 패턴

#### 원문: 프록시 패턴

<!-- curriculum-chunk: sha256=57802fcd658431753e41b9209b86469613946e86eb5e9b8c69234461489dc072 major=problem-solving-code-quality mid=설계 원칙과 패턴 sub=프록시 패턴 sources=source/interview_questions.md:8110-8549, source/interviews.md:8058-8497 -->

> Source: `source/interview_questions.md:8110-8549`
> Classification reason: code design
> Duplicate source aliases: `source/interview_questions.md:8110-8549, source/interviews.md:8058-8497`

##### 프록시 패턴

일반적으로 프록시는 다른 무언가와 이어지는 인터페이스의 역할을 하는 클래스입니다.
프록시는 어떠한 것(이를테면 네트워크 연결, 메모리 안의 커다란 객체, 파일, 또 복제할 수 없거나 수요가 많은 리소스)과도 인터페이스의 역할을 수행할 수 있습니다.

프록시 패턴은 실제 객체의 동작을 제어하거나 부가적인 기능을 추가하기 위해 활용되는 디자인 패턴입니다.
이 패턴은 실제 객체의 동작을 제어하거나 추가적인 기능을 제공하기 위해 사용됩니다.

- 실제 객체의 생성 비용이 높을 때: 가상 프록시(Virtual Proxy)
- 접근 제어가 필요할 때: 보호 프록시(Protection Proxy)
- 원격 객체에 대한 접근 추상화: 원격 프록시(Remote Proxy)
- 부가적인 기능을 투명하게 추가하고 싶을 때: 스마트 프록시(Smart Reference Proxy)

Spring Framework에서는 프록시 패턴을 기반으로 AOP, 트랜잭션 관리, 지연 로딩 등의 기능을 제공합니다.

- 실제 객체의 생성 비용이 높을 때 (Virtual Proxy)

  예시: 고해상도의 이미지나 대용량 데이터를 로딩해야 하는 경우
    - 애플리케이션에서 고해상도 이미지를 표시해야 하는데, 이미지 파일의 크기가 매우 커서 로딩 시간이 오래 걸립니다.
    - 사용자가 해당 이미지를 반드시 볼 필요가 없을 수도 있으므로, 필요할 때만 이미지를 로딩하고 싶습니다.

  프록시 패턴 적용:
    - 가상 프록시(Virtual Proxy)를 사용하여 실제 이미지 객체의 로딩을 지연시킵니다.
    - 프록시 객체는 이미지가 필요한 시점까지 실제 이미지를 로딩하지 않고 대기합니다.
    - 이미지가 필요할 때 프록시 객체는 실제 이미지 객체를 생성하고 로딩합니다.

    ```kotlin
    // 이미지 인터페이스
    interface Image {
        fun display()
    }

    // 실제 이미지 클래스
    class RealImage(private val filename: String) : Image {
        init {
            loadFromDisk(filename)
        }

        private fun loadFromDisk(filename: String) {
            println("Loading $filename")
            // 실제 이미지 로딩 코드 (시간이 많이 걸림)
        }

        override fun display() {
            println("Displaying $filename")
        }
    }

    // 프록시 이미지 클래스
    class ProxyImage(private val filename: String) : Image {
        private var realImage: RealImage? = null

        override fun display() {
            if (realImage == null) {
                realImage = RealImage(filename) // 실제 객체 생성 지연
            }
            realImage?.display()
        }
    }

    // 사용 예시
    fun main() {
        val image: Image = ProxyImage("test_image.jpg")
        // 실제 이미지는 아직 로딩되지 않음

        // 이미지가 필요할 때
        image.display()
        // 여기서 실제 이미지가 로딩되고 표시됨
    }
    ```

  불필요한 자원 낭비를 방지하고 애플리케이션의 초기 로딩 시간을 단축합니다.

- 접근 제어가 필요할 때 (Protection Proxy)

  예시: 데이터베이스에 민감한 데이터가 있고, 접근 권한이 있는 사용자만 데이터에 접근해야 하는 경우
    - 시스템에서 특정 작업이나 데이터에 대한 접근을 사용자 권한에 따라 제한해야 합니다.
    - 모든 클래스에 권한 체크 로직을 구현하면 코드가 복잡해집니다.

  프록시 패턴 적용:
    - 보호 프록시(Protection Proxy)를 사용하여 접근 제어를 담당하는 프록시 객체를 만듭니다.
    - 프록시 객체는 실제 객체에 대한 접근 전에 권한을 확인하고, 권한이 없으면 예외를 발생시킵니다.

    ```kotlin
    // 업무 인터페이스
    interface WorkService {
        fun performTask()
    }

    // 실제 업무 클래스
    class RealWorkService : WorkService {
        override fun performTask() {
            println("작업을 수행합니다.")
        }
    }

    // 보호 프록시 클래스
    class WorkServiceProxy(private val userRole: String) : WorkService {
        private val realWorkService = RealWorkService()

        override fun performTask() {
            if (userRole == "ADMIN") {
                realWorkService.performTask()
            } else {
                throw IllegalAccessException("접근 권한이 없습니다.")
            }
        }
    }

    // 사용 예시
    fun main() {
        val adminService: WorkService = WorkServiceProxy("ADMIN")
        adminService.performTask() // 작업 수행 성공

        val userService: WorkService = WorkServiceProxy("USER")
        userService.performTask() // 예외 발생: 접근 권한이 없습니다.
    }
    ```

  접근 제어 로직을 중앙화하여 코드의 유지보수성을 높이고, 보안을 강화합니다.

- 원격 객체에 대한 접근 추상화 (Remote Proxy)

  예시: 클라이언트가 원격 서버의 객체를 사용하는 경우
    - 클라이언트 애플리케이션이 원격 서버에서 제공하는 서비스를 사용해야 합니다.
    - 네트워크 통신, 데이터 직렬화 등 복잡한 로직을 클라이언트가 알 필요 없이 서비스를 사용하고 싶습니다.

  프록시 패턴 적용:
    - 원격 프록시(Remote Proxy)를 사용하여 원격 객체에 대한 대리자를 제공합니다.
    - 프록시 객체는 네트워크 통신 및 데이터 처리를 담당하고, 클라이언트는 로컬 객체를 사용하듯이 서비스를 이용할 수 있습니다.

    ```kotlin
    // 서비스 인터페이스
    interface RemoteService {
        fun fetchData(): String
    }

    // 실제 원격 서비스 (서버 측에 존재)
    class RealRemoteService : RemoteService {
        override fun fetchData(): String {
            return "원격 데이터"
        }
    }

    // 원격 프록시 (클라이언트 측에 존재)
    class RemoteServiceProxy : RemoteService {
        override fun fetchData(): String {
            // 네트워크 통신 코드 (예: REST API 호출)
            println("원격 서비스에 요청을 보냅니다.")
            // 실제로는 HTTP 요청 등을 통해 데이터를 가져옴
            return "원격 데이터 (프록시를 통해 수신)"
        }
    }

    // 사용 예시 (클라이언트 측)
    fun main() {
        val service: RemoteService = RemoteServiceProxy()
        val data = service.fetchData()
        println("데이터: $data")
    }
    ```

  네트워크 통신의 복잡성을 숨기고, 클라이언트가 쉽게 원격 서비스를 사용할 수 있게 합니다.

- 부가적인 기능을 투명하게 추가하고 싶을 때 (Smart Reference Proxy)

  예시: 객체의 메서드 호출 전후로 로깅이나 트랜잭션 처리 등을 수행하고 싶은 경우
    - 기존 클래스의 코드를 변경하지 않고 부가적인 기능(로깅, 트랜잭션 등)을 추가하고 싶습니다.
    - 여러 클래스에서 동일한 부가 기능을 적용해야 합니다.

  프록시 패턴 적용:
    - 스마트 프록시(Smart Reference Proxy)를 사용하여 실제 객체의 메서드 호출 전후로 부가적인 처리를 수행합니다.
    - 프록시 객체는 실제 객체와 동일한 인터페이스를 구현하고, 부가 기능을 추가합니다.

    ```kotlin
    // 서비스 인터페이스
    interface Service {
        fun execute()
    }

    // 실제 서비스 클래스
    class RealService : Service {
        override fun execute() {
            println("실제 서비스 실행")
        }
    }

    // 로깅 프록시 클래스
    class LoggingProxy(private val realService: Service) : Service {
        override fun execute() {
            println("메서드 실행 전 로깅")
            realService.execute()
            println("메서드 실행 후 로깅")
        }
    }

    // 사용 예시
    fun main() {
        val service: Service = LoggingProxy(RealService())
        service.execute()
        // 출력:
        // 메서드 실행 전 로깅
        // 실제 서비스 실행
        // 메서드 실행 후 로깅
    }
    ```

  코드 수정 없이 부가적인 기능을 투명하게 추가할 수 있습니다.

Spring Framework는 프록시 패턴을 광범위하게 활용하여 다양한 기능을 제공합니다.

- AOP (Aspect-Oriented Programming, 관점 지향 프로그래밍)

  Spring AOP는 프록시 패턴을 사용하여 메서드 호출 전후로 부가적인 기능(어드바이스)을 적용합니다.
  예를 들어, 메서드 실행 전에 권한 체크, 실행 후 로깅 등을 추가할 수 있습니다.

  동작 방식:
    - JDK 동적 프록시: 인터페이스가 있는 경우, `java.lang.reflect.Proxy`를 사용하여 런타임에 프록시 객체를 생성합니다.
    - CGLIB 프록시: 클래스 기반의 프록시로, 상속을 통해 프록시 객체를 생성합니다.

      > [CGLIB(Code Generator Library): 코드 생성 라이브러리로서 런타임에 동적으로 자바 클래스의 프록시를 생성해주는 기능을 제공한다. 인터페이스가 아닌 클래스에 대해서 동적 프록시를 생성할 수 있다.](https://memodayoungee.tistory.com/151)
      > CGLIB는 타겟에 대한 정보를 직접적으로 제공 받아 바이트 코드를 조작하여 프록시를 생성한다.
      > 때문에 리플렉션을 사용하는 JDK Dynamic Proxy에 비해 성능이 좋다.
      >
      > 또한 CGLIB는 메소드가 처음 호출 되었을 때 동적으로 타겟 클래스의 바이트 코드를 조작하고,
      > 이후 호출 시엔 조작된 바이트 코드를 재사용한다.

    ```java
    @Component
    public class MyService {
        public void perform() {
            System.out.println("서비스 실행");
        }
    }

    @Aspect
    @Component
    public class LoggingAspect {
        @Before("execution(* com.example.MyService.perform(..))")
        public void logBefore(JoinPoint joinPoint) {
            System.out.println("메서드 실행 전 로깅");
        }

        @After("execution(* com.example.MyService.perform(..))")
        public void logAfter(JoinPoint joinPoint) {
            System.out.println("메서드 실행 후 로깅");
        }
    }

    // 사용 예시
    @SpringBootApplication
    public class Application {
        public static void main(String[] args) {
            ApplicationContext context = SpringApplication.run(Application.class, args);
            MyService service = context.getBean(MyService.class);
            service.perform();
            // 출력:
            // 메서드 실행 전 로깅
            // 서비스 실행
            // 메서드 실행 후 로깅
        }
    }
    ```

- `@Transactional` 어노테이션

  `@Transactional`을 사용하여 메서드나 클래스에 트랜잭션을 적용하면,
  Spring은 해당 객체의 프록시를 생성하여 트랜잭션 경계를 관리합니다.
  메서드 호출 시 프록시가 트랜잭션 시작과 종료를 처리합니다.

  프록시 객체가 실제 객체의 메서드를 호출하기 전에 트랜잭션을 시작하고, 메서드 실행 후 트랜잭션을 커밋하거나 롤백합니다.

    ```java
    @Service
    public class OrderService {
        @Transactional
        public void placeOrder(Order order) {
            // 주문 처리 로직
        }
    }

    // 사용 예시
    @Component
    public class OrderController {
        @Autowired
        private OrderService orderService;

        public void processOrder() {
            orderService.placeOrder(new Order());
            // 프록시를 통해 트랜잭션이 관리됨
        }
    }
    ```

  단, 스프링의 `@Transactional`은 프록시 패턴을 사용하므로, 같은 클래스 내의 메소드 호출에서는 트랜잭션이 적용되지 않을 수 있습니다

  Spring의 `@Transactional` 어노테이션은 기본적으로 프록시 패턴을 사용하여 트랜잭션을 관리합니다.
  이로 인해 같은 클래스 내에서 메서드를 호출할 때, 예상대로 트랜잭션이 적용되지 않는 경우가 발생할 수 있습니다.

  이 프록시는 원래 객체를 감싸고 있으며, 클라이언트가 `@Transactional`이 붙은 메서드를 호출할 때 프록시가 중간에 개입하여 트랜잭션을 시작하고, 종료 시 커밋 또는 롤백을 수행합니다.
  프록시의 역할:
    - 클라이언트가 트랜잭션을 관리해야 하는 메서드를 호출할 때, 프록시가 대신 호출을 가로채고 트랜잭션을 처리합니다.
    - 트랜잭션을 시작하거나 롤백할지 여부는 프록시가 결정합니다.

  스프링의 `@Transactional`은 프록시 패턴을 사용하기 때문에, *외부에서 프록시 객체를 통해 메서드가 호출되어야 트랜잭션이 적용*됩니다.
  하지만, 같은 클래스 내에서 `@Transactional`이 붙은 다른 메서드를 호출할 때는 프록시가 개입하지 못합니다.
  즉, 트랜잭션이 설정된 메서드를 호출할 때 프록시 객체를 거치지 않으면, 트랜잭션이 적용되지 않습니다.

    1. 내부 호출 시 프록시가 작동하지 않음:
       같은 클래스 내에서 메서드가 호출될 경우, 이 호출은 직접 호출로 간주되며, 프록시를 거치지 않고 원래 객체의 메서드를 호출하게 됩니다.
       즉, 프록시가 개입할 기회를 얻지 못하므로 트랜잭션이 시작되지 않습니다.

    2. 프록시는 외부 호출을 가로채도록 설계됨:
       프록시 패턴의 핵심은 클라이언트가 외부에서 해당 메서드를 호출할 때만 동작하도록 설계되어 있습니다.
       같은 클래스 내에서 직접 메서드를 호출하면 이 프록시 패턴이 우회되며, 트랜잭션 관리 기능이 작동하지 않습니다.

    ```java
    // 트랜잭션이 적용되지 않는 상황
    @Service
    public class OrderService {

        @Transactional
        public void createOrder() {
            // Some business logic
            updateOrder();  // 같은 클래스 내에서 호출
        }

        @Transactional
        public void updateOrder() {
            // Some update logic
        }
    }
    ```

  위 코드에서 `createOrder()`가 호출될 때, 그 내부에서 `updateOrder()`가 호출되더라도 `updateOrder()`에 선언된 `@Transactional`이 적용되지 않습니다.
  이유는 `createOrder()` 메서드가 `updateOrder()`를 *직접 호출*하고 있으며, 이 호출은 *프록시를 경유하지 않기 때문*입니다.

  이 문제를 해결하기 위한 몇 가지 방법이 있습니다.

    - 자기 자신을 프록시 객체로 호출하기

      스프링에서 자기 자신을 프록시로 호출하게 하는 방법은 `AopContext.currentProxy()`를 사용하는 것입니다.
      이를 통해 자기 자신의 프록시 객체를 참조하여 프록시를 경유해 트랜잭션을 적용할 수 있습니다.

        ```java
        import org.springframework.aop.framework.AopContext;

        @Service
        public class OrderService {

            @Transactional
            public void createOrder() {
                // Some business logic
                ((OrderService) AopContext.currentProxy()).updateOrder();  // 프록시를 통해 호출
            }

            @Transactional
            public void updateOrder() {
                // Some update logic
            }
        }
        ```

      위 코드에서는 `AopContext.currentProxy()`를 사용해 자기 자신의 프록시 객체를 가져와 `updateOrder()`를 호출합니다.
      이렇게 하면 트랜잭션이 정상적으로 적용됩니다.

    - 별도의 서비스 클래스 사용

      한 클래스 내에서 메서드 간 호출로 트랜잭션이 적용되지 않는 문제를 해결하는 또 다른 방법은 트랜잭션이 필요한 메서드를 별도의 서비스 클래스로 분리하는 것입니다.
      이렇게 하면 호출할 때마다 프록시 객체를 통해 메서드가 호출되므로 트랜잭션이 적용됩니다.

        ```java
        @Service
        public class OrderService {

            private final UpdateOrderService updateOrderService;

            public OrderService(UpdateOrderService updateOrderService) {
                this.updateOrderService = updateOrderService;
            }

            @Transactional
            public void createOrder() {
                // Some business logic
                updateOrderService.updateOrder();  // 외부 서비스로 호출
            }
        }

        @Service
        public class UpdateOrderService {

            @Transactional
            public void updateOrder() {
                // Some update logic
            }
        }
        ```

      이 방식에서는 `updateOrder()`가 `UpdateOrderService`라는 별도의 서비스로 분리되므로, 항상 프록시 객체를 통해 호출됩니다.
      따라서 `@Transactional`이 올바르게 작동하게 됩니다.

    - Spring의 `@Transactional`은 프록시 패턴을 사용하여 트랜잭션을 관리합니다.
    - 같은 클래스 내에서 메서드 간 호출이 발생하면, 프록시 객체를 경유하지 않고 직접 호출되므로, `@Transactional`이 작동하지 않습니다.
    - 이를 해결하기 위해서는 프록시 객체를 통해 메서드를 호출하거나, 별도의 서비스 클래스로 트랜잭션이 필요한 메서드를 분리하는 방법이 있습니다.

- Lazy Initialization

  `@Lazy` 어노테이션을 사용하여 빈의 초기화를 지연시킬 수 있습니다.
  실제 빈이 필요할 때까지 초기화를 늦추기 위해 프록시를 사용합니다.

  동작 방식:
    - 프록시 객체는 실제 빈에 대한 참조를 갖지 않고, 메서드 호출 시점에 실제 빈을 초기화하고 호출을 위임합니다.

    ```java
    @Component
    @Lazy
    public class HeavyService {
        public HeavyService() {
            System.out.println("HeavyService 초기화");
            // 초기화 비용이 큰 작업
        }

        public void perform() {
            System.out.println("HeavyService 실행");
        }
    }

    @Component
    public class ApplicationRunner implements CommandLineRunner {
        @Autowired
        private HeavyService heavyService;

        @Override
        public void run(String... args) throws Exception {
            // heavyService가 아직 초기화되지 않음
            heavyService.perform();
            // 여기서 초기화되고 메서드 실행
        }
    }
    ```

<!-- /curriculum-chunk -->

## 알고리즘과 문제 해결

### 정규표현식

#### 원문: 정규표현식

<!-- curriculum-chunk: sha256=f0e0308db7a595715caff3fd24fb9a7d7db7dd828697fed05d89026ea5b04ecc major=problem-solving-code-quality mid=알고리즘과 문제 해결 sub=정규표현식 sources=source/interview_questions.md:7978-7979, source/interviews.md:7926-7927 -->

> Source: `source/interview_questions.md:7978-7979`
> Classification reason: algorithm/problem-solving
> Duplicate source aliases: `source/interview_questions.md:7978-7979, source/interviews.md:7926-7927`

##### 정규표현식

<!-- /curriculum-chunk -->

## 운영 실천과 배포 기반

### Docker 컨테이너 가상화

#### 원문: Docker 컨테이너 가상화

<!-- curriculum-chunk: sha256=538143bd0138734271dc13faa2a4ae0d392583c572cbfb770fc521e95e379e88 major=problem-solving-code-quality mid=운영 실천과 배포 기반 sub=Docker 컨테이너 가상화 sources=source/interview_questions.md:100-232, source/interviews.md:100-232 -->

> Source: `source/interview_questions.md:100-232`
> Classification reason: devops/deployment practice
> Duplicate source aliases: `source/interview_questions.md:100-232, source/interviews.md:100-232`

##### Docker 컨테이너 가상화

Docker는 전통적인 하드웨어 가상화와는 다르게, 운영체제 수준의 가상화를 통해 컨테이너를 실행합니다.
Docker 컨테이너는 커널을 공유하지만, 각 컨테이너는 독립된 환경에서 실행되는 것처럼 동작합니다.
Docker는 이러한 가상화와 격리를 리눅스 커널의 기능을 활용해 구현합니다. 여기에는 네임스페이스(`namespaces`)와 `cgroups`(control groups)라는 핵심 기술이 사용됩니다.

컨테이너는 호스트 OS의 커널을 공유하며, 이는 컨테이너 내부에서 발생하는 시스템 호출이 호스트 커널로 직접 전달된다는 의미입니다.
따라서 Docker는 하이퍼바이저처럼 호스트 OS 위에 새로운 커널을 실행하지 않고, 하나의 커널을 여러 격리된 컨텍스트로 분리하는 방식으로 가상화 효과를 제공합니다.

1. Docker와 운영체제 수준의 가상화

   Docker는 하드웨어 가상화와는 다르게 OS 수준의 가상화를 통해 애플리케이션을 격리합니다.
   Docker 컨테이너는 호스트 시스템의 커널을 공유하면서도 각 컨테이너는 독립된 프로세스, 파일 시스템, 네트워크 환경을 갖는 것처럼 보이게 만듭니다.

   컨테이너는 각자의 애플리케이션과 필요한 라이브러리를 포함하지만, 호스트 커널과 직접 상호작용합니다.
   즉, Docker 컨테이너 내부에서 발생하는 시스템 호출은 호스트의 커널에 의해 처리됩니다.
   이를 가능하게 하는 두 가지 주요 기술은 네임스페이스(`namespaces`)와 `cgroups`(control groups)입니다.

2. 네임스페이스(namespaces): 자원의 격리

   네임스페이스는 커널 레벨에서 자원(리소스)을 격리하는 메커니즘입니다.
   각 네임스페이스는 프로세스 그룹이 독립된 환경에서 실행되는 것처럼 보이도록 만들어줍니다.
   Docker는 이러한 네임스페이스를 활용해 프로세스, 파일 시스템, 네트워크 인터페이스, IPC 등을 각각의 컨테이너에 대해 격리합니다.

   리눅스에는 다음과 같은 여러 네임스페이스가 있습니다:

    1. PID 네임스페이스:
        - 컨테이너는 독립된 프로세스 ID (PID) 공간을 가집니다. 이를 통해 각 컨테이너는 마치 자신만의 프로세스 트리가 있는 것처럼 보입니다.
        - 컨테이너 내에서 `pid=1` 프로세스는 컨테이너의 주 프로세스로 동작하며, 호스트의 PID와는 무관한 프로세스 트리를 형성합니다.

    2. NET 네임스페이스:
        - 컨테이너는 독립적인 네트워크 인터페이스와 IP 주소를 가집니다. 각 컨테이너는 고유한 네트워크 네임스페이스 내에서 동작하며, 다른 컨테이너 및 호스트와 격리된 네트워크 스택을 사용합니다.
        - NAT(Network Address Translation) 등을 통해 호스트의 네트워크와 통신할 수 있지만, 기본적으로는 각 컨테이너가 별도의 네트워크 인터페이스를 가진 것처럼 보입니다.

    3. MNT 네임스페이스 (Mount):
        - 컨테이너는 독립된 파일 시스템을 갖습니다. 호스트의 파일 시스템 일부를 컨테이너 내부에서 마운트하거나, 호스트 파일 시스템과는 별도로 컨테이너에 고유한 파일 시스템을 사용할 수 있습니다.
        - 이로 인해 각 컨테이너는 격리된 루트 파일 시스템을 가지고 실행됩니다.

    4. IPC 네임스페이스 (Inter-process communication):
        - 컨테이너는 독립된 IPC 메커니즘(공유 메모리나 세마포어 등)을 사용하여 다른 컨테이너와 격리된 상태로 통신합니다.

    5. UTS 네임스페이스 (UNIX Timesharing System):
        - 컨테이너는 각자 호스트 이름(hostname)과 도메인 이름(domain name)을 가질 수 있습니다. 이는 네임스페이스 내에서만 유효한 독립된 이름을 설정하여, 컨테이너가 자신만의 네트워크 식별자를 갖는 것처럼 보이게 합니다.

    6. USER 네임스페이스:
        - 컨테이너는 사용자 권한을 격리할 수 있습니다. 컨테이너 내에서는 루트 사용자로 동작하더라도, 실제로는 호스트에서는 권한이 제한된 사용자로 매핑됩니다. 이는 보안적으로 중요한 역할을 합니다.

   네임스페이스 동작 방식 예시:
   컨테이너 내부에서 `ps` 명령어로 프로세스 목록을 보면, 컨테이너 내부 프로세스만 보이게 됩니다.
   이때 실제로는 호스트 OS의 프로세스 ID와 매핑되지만, PID 네임스페이스를 통해 컨테이너 내부에서는 격리된 프로세스 트리처럼 동작합니다.

3. Cgroups (Control Groups): 자원 사용량 제어

   cgroups(Control Groups)는 리눅스 커널 기능으로, 시스템 자원의 사용량을 제어하고 모니터링하는 역할을 합니다.
   이를 통해 CPU 사용률, 메모리, 디스크 I/O, 네트워크 대역폭 등을 특정 프로세스 그룹에 대해 제한하거나 관리할 수 있습니다.

   Docker는 cgroups를 사용해 각 컨테이너가 사용할 수 있는 자원을 제한하거나 할당하며, 이로 인해 리소스의 과도한 사용을 방지하고 안정성을 확보합니다.
   Docker는 컨테이너를 생성할 때, 해당 컨테이너가 사용할 수 있는 리소스를 제어하기 위해 각 컨테이너에 대해 별도의 cgroup을 생성합니다.
   각 cgroup은 독립적으로 CPU, 메모리, I/O 등의 자원을 제어하며, 컨테이너마다 cgroup이 고유하게 설정됩니다.
   이는 하나의 컨테이너가 과도하게 자원을 사용하는 것을 방지하고, 호스트 전체 자원 사용을 안정적으로 관리할 수 있게 해줍니다.

    ```sh
    # 이 명령을 실행하면, Docker는 컨테이너에 대해 독립된 cgroup을 생성하고 해당 설정을 반영합니다.
    # 컨테이너는 자신에게 할당된 메모리와 CPU만 사용할 수 있으며, 이를 초과하려고 하면 cgroup에 의해 제한됩니다.
    # 메모리를 512MB로 제한하고, CPU의 50%만 사용할 수 있도록 설정합니다.
    docker run -d --memory="512m" --cpus="0.5" myapp
    ```

    - CPU 제한:
        - 특정 컨테이너가 사용할 수 있는 *CPU 시간을 제어*할 수 있습니다.
        - 이를 통해 여러 컨테이너가 동시에 실행될 때 공정한 CPU 자원 배분이 이루어집니다.
        - 예: `--cpu-quota`, `--cpu-shares` 옵션을 사용하여 컨테이너의 CPU 사용을 제한할 수 있습니다.
    - 메모리 제한:
        - 각 컨테이너가 사용할 수 있는 최대 메모리 용량을 설정할 수 있습니다.
        - 설정된 메모리를 초과하면 컨테이너는 종료되거나 메모리 부족 오류(OOM)가 발생할 수 있습니다.
    - 디스크 I/O 제어:
        - 특정 컨테이너가 디스크 읽기 및 쓰기 작업에서 사용할 수 있는 I/O 대역폭을 제한할 수 있습니다
        - `--blkio-weight` 옵션을 사용하여 I/O 우선순위를 설정할 수 있습니다.
    - 네트워크 대역폭 제한: 특정 컨테이너가 사용할 수 있는 네트워크 대역폭을 제한할 수 있습니다.

   Docker는 각 컨테이너를 하나의 프로세스 트리로 관리합니다.
   각 컨테이너의 프로세스는 cgroups에 의해 자원 제약을 받으며, 해당 프로세스 트리의 모든 자식 프로세스가 동일한 cgroup 제한을 상속받습니다.
   컨테이너가 시작되면 Docker는 해당 컨테이너의 프로세스들을 특정 cgroup에 할당하여 자원 사용을 제한하거나 관리합니다.

   예를 들어, 컨테이너 A와 컨테이너 B가 있을 때:
    - 컨테이너 A는 cgroup A에 할당되어, CPU 사용률 50%, 메모리 1GB로 제한될 수 있습니다.
    - 컨테이너 B는 cgroup B에 할당되어, CPU 사용률 25%, 메모리 512MB로 제한될 수 있습니다.

   cgroups 동작 방식:
   Docker는 각 컨테이너에 대해 별도의 cgroup을 생성하여 자원 사용을 제어합니다.
   이를 통해 호스트 시스템의 자원이 특정 컨테이너에 의해 고갈되지 않도록 보장합니다.
   cgroups는 컨테이너 간 자원 경쟁을 관리하는 중요한 역할을 합니다.

4. 시스템 콜과 호스트 커널 상호작용

   Docker 컨테이너 내에서 발생하는 모든 시스템 콜은 호스트 커널에서 처리됩니다.
   즉, 컨테이너는 자체적으로 커널을 갖지 않고, 호스트 커널과 직접 상호작용합니다.
   컨테이너 내부에서 동작하는 애플리케이션이 시스템 호출을 하면, 그 호출은 호스트 커널에 의해 실행되며, 이때 네임스페이스와 cgroups에 의해 격리된 상태에서 처리됩니다.

   시스템 콜의 흐름:
    1. 컨테이너 내부 애플리케이션에서 시스템 호출(예: 파일 읽기)을 발생시킵니다.
    2. 해당 시스템 호출은 호스트 커널로 전달됩니다.
    3. 호스트 커널은 네임스페이스와 cgroups 설정에 따라 해당 컨테이너의 자원 제한과 격리를 반영하여 시스템 호출을 처리합니다.
    4. 결과는 컨테이너로 다시 전달됩니다.

   호스트 커널은 네임스페이스를 기반으로 격리된 환경을 유지하면서도 동일한 커널 자원을 공유할 수 있게 하며,
   이러한 동작 방식 덕분에 컨테이너는 마치 별도의 운영체제에서 동작하는 것처럼 느껴지지만 실제로는 호스트 OS의 커널을 공유합니다.

5. 컨테이너의 "가상화된" 환경

   컨테이너는 마치 독립된 OS에서 실행되는 것처럼 보이지만,
   사실상 단일 커널을 공유하면서 격리된 리소스를 사용하는 것입니다.

   이로 인해 다음과 같은 특징이 발생합니다:
    - 가볍고 빠름: 하이퍼바이저를 사용하는 전통적인 가상화와 달리, 컨테이너는 추가적인 커널을 실행하지 않기 때문에 메모리와 CPU 자원을 적게 사용하며, 빠르게 시작할 수 있습니다.
    - 효율적인 자원 사용: 여러 컨테이너가 동일한 커널을 공유하면서도, 각자 자원을 격리하여 사용하는 방식이므로, 리소스를 효율적으로 사용할 수 있습니다.
    - 보안 격리: 네임스페이스와 cgroups을 통해 각 컨테이너는 보안적으로 격리된 상태로 실행되며, 호스트 시스템에 영향을 주지 않고 실행될 수 있습니다.

6. Docker와 서로 다른 OS 사용에 대한 오해

   컨테이너 내부에서 다른 운영체제를 실행할 수 있다는 오해가 종종 있습니다.
   Docker는 호스트 OS의 커널을 그대로 사용하며, 실제로는 동일한 커널 버전의 Linux 기반 운영체제만을 지원합니다.
   Docker 컨테이너에서 제공하는 환경은 호스트 OS 커널과 직접 상호작용하며, 호스트 OS와 동일한 커널을 사용하는 Linux 기반의 파일 시스템과 사용자 공간을 제공합니다.

   다만, 컨테이너 이미지에 포함된 라이브러리, 실행 환경, 도구들이 다르기 때문에, 각 컨테이너가 서로 다른 배포판처럼 보일 수 있습니다.
   예를 들어, Ubuntu 기반 컨테이너와 CentOS 기반 컨테이너는 서로 다른 패키지 매니저와 파일 시스템 구조를 가질 수 있지만, 커널은 동일합니다.

###### 결론

Docker는 리눅스 커널의 네임스페이스와 cgroups 기능을 활용하여, 각 컨테이너가 독립된 운영체제에서 실행되는 것처럼 자원 격리와 가상화를 제공합니다. 이 방식은 하이퍼바이저 기반의 가상화와 달리 단일 커널을 공유하면서도 컨테이너 간 격리된 환경을 유지하여, 경량의 가상화 솔루션을 제공합니다.

Docker 내부의 애플리케이션에서 발생하는 시스템 콜은 호스트 커널에서 처리되며, 이를 통해 각 컨테이너는 독립적인 파일 시스템과 네트워크 공간을 사용하면서도, 호스트 커널과 자원을 효율적으로 공유하게 됩니다.

<!-- /curriculum-chunk -->

## 테스트와 대역 객체

### java, php에서 mock 생성 원리

#### 원문: java, php에서 mock 생성 원리

<!-- curriculum-chunk: sha256=60694b1d89a84a70b64b4693327d5525eba0b2ee177bd81d88a27d0ca242cc08 major=problem-solving-code-quality mid=테스트와 대역 객체 sub=java, php에서 mock 생성 원리 sources=source/interview_questions.md:9461-9751, source/interviews.md:9409-9699 -->

> Source: `source/interview_questions.md:9461-9751`
> Classification reason: testing
> Duplicate source aliases: `source/interview_questions.md:9461-9751, source/interviews.md:9409-9699`

##### java, php에서 mock 생성 원리

Java나 PHP에서 사용되는 Mock 객체는 주로 리플렉션(reflection)과 프록시(proxy) 패턴을 사용하여 동적으로 생성됩니다.
Mocking 프레임워크들은 이러한 기술을 이용해 런타임 시점에 가짜 객체를 생성하고, 특정 메서드 호출 시 지정된 동작을 수행하도록 만들어줍니다.

- 리플렉션 (Reflection)

  리플렉션은 프로그램이 실행 중에 자신의 구조에 대해 알 수 있게 해주고, 런타임에 객체의 메서드나 필드에 접근하고 수정할 수 있는 기능을 제공합니다.
  Java나 PHP의 Mocking 프레임워크는 리플렉션을 통해 객체의 메서드와 필드를 동적으로 조작하고, 호출된 메서드의 정보를 추적할 수 있습니다.

    ```java
    /*
     * 리플렉션을 사용하여 메서드를 동적으로 호출하는 예제
     *
     * 이 메서드는 주어진 객체의 메서드를 이름으로 찾아 호출합니다.
     * 리플렉션을 사용하여 런타임에 메서드의 존재를 확인하고 호출할 수 있습니다.
     *
     * @param obj 메서드를 호출할 객체
     * @param methodName 호출할 메서드의 이름
     * @param args 메서드에 전달할 인자들
     * @return 메서드 호출의 결과
     * @throws Exception 메서드를 찾지 못하거나 호출 중 예외가 발생한 경우
     */
    public static Object invokeMethod(Object obj, String methodName, Object... args) throws Exception {
        // 객체의 클래스 정보를 가져옵니다.
        Class<?> clazz = obj.getClass();

        // 인자들의 클래스 타입을 저장할 배열을 생성합니다.
        Class<?>[] parameterTypes = new Class<?>[args.length];
        for (int i = 0; i < args.length; i++) {
            parameterTypes[i] = args[i].getClass();
        }

        // 지정된 이름과 파라미터 타입을 가진 메서드를 찾습니다.
        Method method = clazz.getMethod(methodName, parameterTypes);

        // 메서드를 호출하고 결과를 반환합니다.
        return method.invoke(obj, args);
    }

    public static void main(String[] args) {
        String str = "Hello, World!";
        try {
            // length() 메서드를 동적으로 호출합니다.
            int length = (int) invokeMethod(str, "length");
            System.out.println("String length: " + length);

            // charAt(int) 메서드를 동적으로 호출합니다.
            char ch = (char) invokeMethod(str, "charAt", 7);
            System.out.println("Character at index 7: " + ch);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
    ```

  이 예제는 리플렉션을 사용하여 메서드를 동적으로 호출하는 방법을 보여줍니다.
  Mock 객체 생성 시 이와 유사한 기술이 사용되어 메서드 호출을 가로채고 사용자 정의 동작을 주입할 수 있습니다.

- 프록시 (Proxy) 패턴

  프록시 패턴은 대리 객체를 생성하여 실제 객체와 상호작용할 때 그 사이에서 중간 역할을 수행하는 방식입니다.
  Mock 객체는 프록시 패턴을 이용하여 실제 객체처럼 보이지만, 실제로는 모킹된 동작을 수행합니다.
  프록시 객체는 실제 객체의 메서드를 호출하기 전에 추가적인 로직을 실행하거나, 아예 다른 동작을 수행하도록 설정될 수 있습니다.

  Java의 `java.lang.reflect.Proxy` 클래스는 인터페이스 기반의 프록시 객체를 동적으로 생성할 수 있도록 해줍니다.

    ```java
    import java.lang.reflect.InvocationHandler;
    import java.lang.reflect.Method;
    import java.lang.reflect.Proxy;

    /*
     * 동적 프록시를 사용하여 인터페이스의 메서드 호출을 가로채는 예제
     *
     * 이 예제는 MyInterface의 모든 메서드 호출을 가로채고 로깅하는 프록시를 생성합니다.
     * 동적 프록시는 런타임에 인터페이스를 구현하는 클래스를 생성하여 메서드 호출을 중개합니다.
     */

    // 프록시할 인터페이스
    interface MyInterface {
        void doSomething();
        String getSomething();
    }

    // 실제 구현 클래스
    class MyImplementation implements MyInterface {
        @Override
        public void doSomething() {
            System.out.println("Doing something");
        }

        @Override
        public String getSomething() {
            return "Something";
        }
    }

    // InvocationHandler 구현
    class LoggingInvocationHandler implements InvocationHandler {
        private final Object target;

        public LoggingInvocationHandler(Object target) {
            this.target = target;
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            System.out.println("Before method: " + method.getName());
            Object result = method.invoke(target, args);
            System.out.println("After method: " + method.getName() + ", result: " + result);
            return result;
        }
    }

    public class DynamicProxyExample {
        public static void main(String[] args) {
            MyInterface realObject = new MyImplementation();

            MyInterface proxyObject = (MyInterface) Proxy.newProxyInstance(
                MyInterface.class.getClassLoader(),
                new Class<?>[] { MyInterface.class },
                new LoggingInvocationHandler(realObject)
            );

            // 프록시 객체를 통한 메서드 호출
            proxyObject.doSomething();
            String result = proxyObject.getSomething();
            System.out.println("Main got result: " + result);
        }
    }
    ```

  이 예제에서 `LoggingInvocationHandler`는 모든 메서드 호출을 가로채고 로깅합니다.
  이는 Mock 객체가 메서드 호출을 가로채고 사용자 정의 동작을 수행하는 방식과 유사합니다.

- 바이트코드 조작

  바이트코드 조작은 컴파일된 클래스의 바이트코드를 직접 수정하는 기술입니다.
  이 방법은 더 강력하고 유연하지만, 복잡성도 높습니다.
  Java에서는 ASM, javassist 같은 라이브러리가 이 기능을 제공합니다.

    ```java
    import javassist.*;

    /*
     * Javassist를 사용하여 런타임에 클래스를 수정하는 예제
     *
     * MyClass의 doSomething 메서드를 런타임에 수정하여
     * 메서드 실행 전후에 로깅을 추가합니다.
     */
    public class BytecodeManipulationExample {
        public static void main(String[] args) throws Exception {
            // 클래스 풀 생성
            ClassPool pool = ClassPool.getDefault();

            // 수정할 클래스 가져오기
            CtClass cc = pool.get("MyClass");

            // 수정할 메서드 가져오기
            CtMethod m = cc.getDeclaredMethod("doSomething");

            // 메서드 내용 수정
            m.insertBefore("System.out.println(\"Before doSomething\");");
            m.insertAfter("System.out.println(\"After doSomething\");");

            // 수정된 클래스를 새 이름으로 저장
            cc.setName("MyClass_Modified");
            cc.writeFile();

            // 수정된 클래스 로드 및 인스턴스 생성
            Class<?> modifiedClass = cc.toClass();
            Object obj = modifiedClass.newInstance();

            // 수정된 메서드 호출
            modifiedClass.getMethod("doSomething").invoke(obj);
        }
    }

    // 원본 클래스
    class MyClass {
        public void doSomething() {
            System.out.println("Doing something");
        }
    }
    ```

- Mocking 프레임워크의 내부 동작 원리

  Mocking 프레임워크들은 리플렉션과 프록시를 사용하여 가짜 객체를 동적으로 생성하고, 다음과 같은 방식으로 동작합니다:

    1. 클래스 분석:
       Mocking 프레임워크는 모킹할 클래스나 인터페이스의 메서드와 필드를 리플렉션을 사용해 분석합니다.

    2. 프록시 생성:

       Mocking 프레임워크는 프록시 객체를 생성하여 실제 클래스처럼 동작하도록 만듭니다.
       Java에서는 `Proxy.newProxyInstance()` 메서드를 통해 동적으로 프록시를 생성하고,
       PHP에서는 `__call()` 메서드를 오버라이드하는 방식으로 비슷한 기능을 제공합니다.

    3. 메서드 호출 가로채기:

       프록시 객체의 메서드가 호출되면 `InvocationHandler`나 `MethodInterceptor`와 같은 핸들러가 호출을 가로채고, 미리 정의된 동작(예: 지정된 값 반환, 예외 발생 등)을 수행합니다.
       이 과정에서 호출 횟수나 인자 등을 기록하여 나중에 검증할 수 있습니다.

    4. 결과 반환 또는 예외 발생:

       모킹된 메서드는 사전에 설정된 동작을 수행하고, 그 결과를 반환하거나 특정 예외를 발생시킵니다.
       이를 통해 실제 객체의 복잡한 로직이나 외부 의존성 없이도 테스트를 진행할 수 있습니다.

Java와 PHP에서의 Mocking 프레임워크

- Java - Mockito

  Java에서 널리 사용되는 Mockito는 프록시와 리플렉션을 결합하여 Mock 객체를 생성합니다.
  Mockito는 다음과 같은 방식으로 작동합니다:

    - `Mockito.mock()`: 특정 클래스나 인터페이스의 Mock 객체를 생성합니다.
    - `when().thenReturn()`: 특정 메서드 호출 시 반환될 값을 지정합니다.
    - `verify()`: 메서드가 호출되었는지, 특정 인자와 함께 호출되었는지 등을 검증할 수 있습니다.

    ```java
    // Mockito 사용 예시
    import static org.mockito.Mockito.*;

    public class MockitoExample {
        public static void main(String[] args) {
            // List 인터페이스의 mock 객체 생성
            List<String> mockedList = mock(List.class);

            // mock 객체의 동작 정의
            when(mockedList.get(0)).thenReturn("first");
            when(mockedList.get(1)).thenThrow(new RuntimeException());

            // mock 객체 사용
            System.out.println(mockedList.get(0)); // 출력: first
            try {
                mockedList.get(1); // RuntimeException 발생
            } catch (RuntimeException e) {
                System.out.println("Exception caught as expected");
            }

            // 메서드 호출 검증
            verify(mockedList).get(0);
            verify(mockedList).get(1);
            verify(mockedList, never()).clear();

            // Mock 객체 생성
            UserRepository mockRepository = mock(UserRepository.class);

            // 특정 메서드의 반환값 설정
            when(mockRepository.findUserById(1)).thenReturn(new User(1, "Test User"));

            // Mock 객체 사용
            User user = mockRepository.findUserById(1);
            System.out.println(user.getName()); // "Test User" 출력

            // 메서드 호출 검증
            verify(mockRepository).findUserById(1);
        }
    }
    ```

- PHP - PHPUnit

  PHP에서는 PHPUnit의 Mock 기능을 통해 가짜 객체를 생성할 수 있습니다.
  PHPUnit은 프록시 객체를 만들어 인터페이스나 클래스의 메서드를 모킹할 수 있도록 도와줍니다.
  PHP는 동적으로 메서드를 호출할 수 있는 `__call()` 메서드와 리플렉션을 활용하여 비슷한 원리로 작동합니다.

    ```php
    // PHPUnit 사용 예시
    use PHPUnit\Framework\TestCase;

    class UserServiceTest extends TestCase {
        public function testGetUserName() {
            // Mock 객체 생성
            $mockRepository = $this->createMock(UserRepository::class);

            // 특정 메서드의 반환값 설정
            $mockRepository->method('findUserById')
                        ->willReturn(new User(1, 'Test User'));

            // Mock 객체 사용
            $userService = new UserService($mockRepository);
            $this->assertEquals('Test User', $userService->getUserName(1));
        }
    }
    ```

  이 코드에서는 `createMock()` 메서드를 사용해 `UserRepository` 인터페이스의 Mock 객체를 생성하고, `findUserById` 메서드가 호출될 때 `User` 객체를 반환하도록 설정합니다.

<!-- /curriculum-chunk -->

### 테스트 더블

#### 원문: 테스트 더블

<!-- curriculum-chunk: sha256=cf39d1f0e810910e502e8a906af69f65c30a91c69d75fb041eee6c5869bfecdf major=problem-solving-code-quality mid=테스트와 대역 객체 sub=테스트 더블 sources=source/interview_questions.md:9423-9460, source/interviews.md:9371-9408 -->

> Source: `source/interview_questions.md:9423-9460`
> Classification reason: testing
> Duplicate source aliases: `source/interview_questions.md:9423-9460, source/interviews.md:9371-9408`

##### 테스트 더블

테스트 더블(Test Double)은 소프트웨어 테스트에서 사용되는 가짜 객체 또는 대체 객체입니다.

테스트 더블이라는 용어는 원래 영화 산업에서 배우를 대신해 위험한 스턴트나 장면을 연기하는 대역 배우(스턴트 더블)에서 유래했습니다.
영화에서 배우가 위험한 장면을 직접 수행하지 않고 대역을 쓰는 것처럼, 소프트웨어 테스트에서도 실제 객체를 사용하지 않고 대체 객체를 사용하는 개념에서 비롯된 것입니다.

실제 객체를 대신하여 테스트를 진행할 때 사용되며, 테스트 환경에서 실제 객체가 가져오는 복잡성을 줄이거나 특정 상황을 재현하기 위해 사용됩니다.

1. Dummy (더미)
    - 실제로 사용되지 않는 객체로, 단순히 메서드의 매개변수를 채우기 위한 목적으로 사용됩니다.
    - 테스트에서 사용되지는 않지만, 파라미터가 필요해서 만들어지는 객체입니다.
    - 예: 메서드에서 파라미터로 전달되지만, 테스트 동안 전혀 사용되지 않는 객체.

2. Fake (페이크)
    - 간단한 구현을 제공하는 객체로, 실제 제품 코드에서는 사용되지 않지만 테스트 환경에서는 충분한 기능을 수행할 수 있습니다.
    - 예: 인메모리 데이터베이스, 실제 데이터베이스 대신 사용하는 간단한 데이터 저장소.

3. Stub (스텁)
    - 미리 정의된 응답을 반환하는 객체로, 외부 시스템에 의존하지 않고 테스트를 수행할 수 있도록 합니다.
    - 예: 특정 메서드 호출 시, 항상 같은 값을 반환하도록 하는 객체.
    - 주로 고정된 반환 값을 제공하여 테스트의 예측 가능성을 높여줍니다.

4. Mock (목)
    - 행동을 사전에 정의하고, 그 행동이 실제로 일어났는지 확인하는 데 중점을 둡니다.
    - 테스트 중 메서드가 호출되었는지, 특정 파라미터로 호출되었는지 등 행위 검증을 위한 객체입니다.
    - Mock 객체는 호출된 메서드나 사용된 인자들이 예상된 대로 작동했는지 검증할 수 있는 기능을 제공합니다.

5. Spy (스파이)
    - 실제 객체처럼 동작하지만, 메서드 호출 등을 기록하여 나중에 검증할 수 있습니다.
    - Mock과 비슷하지만, Spy는 실제로 일부 로직을 수행하면서도 호출된 메서드나 전달된 파라미터를 확인할 수 있게 해줍니다.
    - 행동을 감시하는 목적이지만, 실제 코드의 동작도 수행합니다.

테스트 더블 사용의 장점은 다음과 같습니다:

- 독립성 보장: 외부 시스템(데이터베이스, 네트워크, 파일 시스템 등)과의 의존성을 제거하여, 테스트가 독립적으로 실행될 수 있습니다.
- 테스트 속도 향상: 실제 객체나 외부 시스템 대신 가짜 객체를 사용함으로써 테스트 실행 속도를 높일 수 있습니다.
- 특정 상황 재현: 테스트에서 다양한 경계 조건이나 예외 상황을 쉽게 재현할 수 있습니다.

<!-- /curriculum-chunk -->
