# Testing And Code Design

> 원문 보존형 이동본입니다. 이 파일의 source chunk 본문은 원본 `intervie*.md`에서 그대로 복사되었고, 기술적 보강과 딥 리라이트는 다음 단계에서 수행합니다.

## Source Chunks

<!-- source-chunk: sha256=57802fcd658431753e41b9209b86469613946e86eb5e9b8c69234461489dc072 topic=testing-and-code-design sources=interview_questions.md:8110-8549, interviews.md:8058-8497 -->

> Source: `interview_questions.md:8110-8549`
> Duplicate source aliases: `interview_questions.md:8110-8549, interviews.md:8058-8497`

## 프록시 패턴

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

<!-- /source-chunk -->

<!-- source-chunk: sha256=cf39d1f0e810910e502e8a906af69f65c30a91c69d75fb041eee6c5869bfecdf topic=testing-and-code-design sources=interview_questions.md:9423-9460, interviews.md:9371-9408 -->

> Source: `interview_questions.md:9423-9460`
> Duplicate source aliases: `interview_questions.md:9423-9460, interviews.md:9371-9408`

## 테스트 더블

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

<!-- /source-chunk -->
