# Spring과 백엔드 프레임워크

- [Spring과 백엔드 프레임워크](#spring과-백엔드-프레임워크)
    - [먼저 기억할 정리](#먼저-기억할-정리)
    - [AOP와 어노테이션 처리](#aop와-어노테이션-처리)
        - [jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리](#jvm-계열-언어에서-어노테이션-선언-방법-및-spring-어노테이션-동작-원리)
            - [원문: jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리](#원문-jvm-계열-언어에서-어노테이션-선언-방법-및-spring-어노테이션-동작-원리)
                - [jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리](#jvm-계열-언어에서-어노테이션-선언-방법-및-spring-어노테이션-동작-원리-1)
        - [어노테이션 프로세싱과 Spring의 내부 동작 원리](#어노테이션-프로세싱과-spring의-내부-동작-원리)
            - [원문: 어노테이션 프로세싱과 Spring의 내부 동작 원리](#원문-어노테이션-프로세싱과-spring의-내부-동작-원리)
                - [어노테이션 프로세싱과 Spring의 내부 동작 원리](#어노테이션-프로세싱과-spring의-내부-동작-원리-1)
    - [Bean 생명주기와 선택](#bean-생명주기와-선택)
        - [`@Bean` vs. `@Component`](#bean-vs-component)
            - [원문: `@Bean` vs. `@Component`](#원문-bean-vs-component)
                - [`@Bean` vs. `@Component`](#bean-vs-component-1)
        - [스프링 부트 시작과 빈 라이프 사이클](#스프링-부트-시작과-빈-라이프-사이클)
            - [원문: 스프링 부트 시작과 빈 라이프 사이클](#원문-스프링-부트-시작과-빈-라이프-사이클)
                - [스프링 부트 시작과 빈 라이프 사이클](#스프링-부트-시작과-빈-라이프-사이클-1)
                    - [SpringApplication.run()](#springapplicationrun)
                    - [스프링 빈의 라이프사이클](#스프링-빈의-라이프사이클)
        - [적절한 Bean 선택](#적절한-bean-선택)
            - [원문: 적절한 Bean 선택](#원문-적절한-bean-선택)
                - [적절한 Bean 선택](#적절한-bean-선택-1)
    - [IoC와 DI](#ioc와-di)
        - [Inversion of Control (IoC)](#inversion-of-control-ioc)
            - [원문: Inversion of Control (IoC)](#원문-inversion-of-control-ioc)
                - [Inversion of Control (IoC)](#inversion-of-control-ioc-1)
        - [Spring Framework의 Dependency Injection (DI)](#spring-framework의-dependency-injection-di)
            - [원문: Spring Framework의 Dependency Injection (DI)](#원문-spring-framework의-dependency-injection-di)
                - [Spring Framework의 Dependency Injection (DI)](#spring-framework의-dependency-injection-di-1)
    - [Servlet/Tomcat 요청 처리](#servlettomcat-요청-처리)
        - [NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우](#nginx와-java-spring-web-애플리케이션이-함께-사용되는-경우)
            - [원문: NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우](#원문-nginx와-java-spring-web-애플리케이션이-함께-사용되는-경우)
                - [NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우](#nginx와-java-spring-web-애플리케이션이-함께-사용되는-경우-1)
        - [Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정](#nginx가-tomcat을-업스트림으로-요청을-전달하고-tomcat이-spring-애플리케이션의-restcontroller나-controller로-요청을-전달하는-과정)
            - [원문: Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정](#원문-nginx가-tomcat을-업스트림으로-요청을-전달하고-tomcat이-spring-애플리케이션의-restcontroller나-controller로-요청을-전달하는-과정)
                - [Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정](#nginx가-tomcat을-업스트림으로-요청을-전달하고-tomcat이-spring-애플리케이션의-restcontroller나-controller로-요청을-전달하는-과정-1)
        - [Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정](#tomcat과-spring의-동작-원리-및-war-파일-처리-과정)
            - [원문: Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정](#원문-tomcat과-spring의-동작-원리-및-war-파일-처리-과정)
                - [Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정](#tomcat과-spring의-동작-원리-및-war-파일-처리-과정-1)
        - [nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유](#nginx와-javakotlin-spring-web-app-사이에-tomcat이-필요한-이유)
            - [원문: nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유](#원문-nginx와-javakotlin-spring-web-app-사이에-tomcat이-필요한-이유)
                - [nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유](#nginx와-javakotlin-spring-web-app-사이에-tomcat이-필요한-이유-1)
    - [Spring Boot 실행 모델](#spring-boot-실행-모델)
        - [2. Spring 관련 질문](#2-spring-관련-질문)
            - [원문: 2. Spring 관련 질문](#원문-2-spring-관련-질문)
                - [2. Spring 관련 질문](#2-spring-관련-질문-1)
                    - [질문 4. Spring IoC(Inversion of Control)와 DI(Dependency Injection)은 어떤 원리로 동작하나요?](#질문-4-spring-iocinversion-of-control와-didependency-injection은-어떤-원리로-동작하나요)
                    - [**답변**](#답변)
                    - [질문 5. Spring AOP(Aspect Oriented Programming)에서는 어떤 방식으로 공통 로직을 분리하나요?](#질문-5-spring-aopaspect-oriented-programming에서는-어떤-방식으로-공통-로직을-분리하나요)
                    - [**답변**](#답변-1)
                    - [질문 6. @Transactional 애노테이션이 어떻게 트랜잭션을 관리하나요?](#질문-6-transactional-애노테이션이-어떻게-트랜잭션을-관리하나요)
                    - [**답변**](#답변-2)
        - [Spring boot에서 톰캣 실행 후 일어나는 일](#spring-boot에서-톰캣-실행-후-일어나는-일)
            - [원문: Spring boot에서 톰캣 실행 후 일어나는 일](#원문-spring-boot에서-톰캣-실행-후-일어나는-일)
                - [Spring boot에서 톰캣 실행 후 일어나는 일](#spring-boot에서-톰캣-실행-후-일어나는-일-1)
        - [Spring의 동시성 및 Thread Management 정리](#spring의-동시성-및-thread-management-정리)
            - [원문: Spring의 동시성 및 Thread Management 정리](#원문-spring의-동시성-및-thread-management-정리)
                - [Spring의 동시성 및 Thread Management 정리](#spring의-동시성-및-thread-management-정리-1)
        - [`java -jar SpringBootApp.jar` 명령어의 실행 과정](#java--jar-springbootappjar-명령어의-실행-과정)
            - [원문: `java -jar SpringBootApp.jar` 명령어의 실행 과정](#원문-java--jar-springbootappjar-명령어의-실행-과정)
                - [`java -jar SpringBootApp.jar` 명령어의 실행 과정](#java--jar-springbootappjar-명령어의-실행-과정-1)
                    - [바이너리를 프로세스로 만드는 과정](#바이너리를-프로세스로-만드는-과정)
                    - [Java Virtual Machine (JVM) 시작](#java-virtual-machine-jvm-시작)
                    - [JVM 내부 메모리 초기화](#jvm-내부-메모리-초기화)
                    - [JAR 파일 로드](#jar-파일-로드)
                    - [클래스 로더에 의해 클래스 로드](#클래스-로더에-의해-클래스-로드)
                    - [`main()` 메서드 실행](#main-메서드-실행)
                    - [Spring Boot의 초기화 단계](#spring-boot의-초기화-단계)
                    - [(a) 스프링 애플리케이션 컨텍스트 초기화](#a-스프링-애플리케이션-컨텍스트-초기화)
                    - [(b) 내장 웹 서버(Tomcat, Jetty 등) 시작](#b-내장-웹-서버tomcat-jetty-등-시작)
                    - [(c) DispatcherServlet 등록](#c-dispatcherservlet-등록)
                    - [(d) 애플리케이션 로직 실행 준비 완료](#d-애플리케이션-로직-실행-준비-완료)
                    - [6. 운영 상태로 전환](#6-운영-상태로-전환)
                    - [추가 사항](#추가-사항)
        - [과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교](#과거-톰캣--스프링-실행-방식-vs-spring-boot-비교)
            - [원문: 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교](#원문-과거-톰캣--스프링-실행-방식-vs-spring-boot-비교)
                - [과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교](#과거-톰캣--스프링-실행-방식-vs-spring-boot-비교-1)
    - [Spring HTTP 클라이언트와 Reactive](#spring-http-클라이언트와-reactive)
        - [Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계](#mono와-flux-개념-필요성-이벤트-루프와의-관계)
            - [원문: Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계](#원문-mono와-flux-개념-필요성-이벤트-루프와의-관계)
                - [Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계](#mono와-flux-개념-필요성-이벤트-루프와의-관계-1)
        - [RestTemplate vs WebClient 차이](#resttemplate-vs-webclient-차이)
            - [원문: RestTemplate vs WebClient 차이](#원문-resttemplate-vs-webclient-차이)
                - [RestTemplate vs WebClient 차이](#resttemplate-vs-webclient-차이-1)
        - [Spring WebFlux의 이벤트 루프(Event Loop)](#spring-webflux의-이벤트-루프event-loop)
            - [원문: Spring WebFlux의 이벤트 루프(Event Loop)](#원문-spring-webflux의-이벤트-루프event-loop)
                - [Spring WebFlux의 이벤트 루프(Event Loop)](#spring-webflux의-이벤트-루프event-loop-1)
    - [트랜잭션과 데이터 접근](#트랜잭션과-데이터-접근)
        - [Spring JDBC에서도 데이터를 가져오는 방식](#spring-jdbc에서도-데이터를-가져오는-방식)
            - [원문: Spring JDBC에서도 데이터를 가져오는 방식](#원문-spring-jdbc에서도-데이터를-가져오는-방식)
                - [Spring JDBC에서도 데이터를 가져오는 방식](#spring-jdbc에서도-데이터를-가져오는-방식-1)
        - [hikari cp](#hikari-cp)
            - [원문: hikari cp](#원문-hikari-cp)
                - [hikari cp](#hikari-cp-1)
        - [리파지토리 메서드에 `@Transactional` 사용](#리파지토리-메서드에-transactional-사용)
            - [원문: 리파지토리 메서드에 `@Transactional` 사용](#원문-리파지토리-메서드에-transactional-사용)
                - [리파지토리 메서드에 `@Transactional` 사용](#리파지토리-메서드에-transactional-사용-1)

Spring Boot, IoC, Bean, Transaction, Servlet/Tomcat, WebClient/WebFlux처럼 백엔드 프레임워크가 런타임 위에 올리는 실행 모델을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## 먼저 기억할 정리

Spring 문서는 어노테이션 이름을 외우는 문서가 아니라 "프레임워크가 객체 생성, 호출 경계, 요청 처리, transaction 상태를 어디에서 대신 관리하는가"를 보는 문서입니다.

```text
java -jar
  -> JVM class loading
  -> SpringApplication / ApplicationContext
  -> bean definition and lifecycle
  -> proxy / interceptor / transaction manager
  -> Tomcat worker or Netty event loop
  -> controller / service / repository
  -> connection pool / DB session
```

비교축은 객체 소유권과 호출 경계입니다. `@Component`와 `@Bean`은 객체를 누가 등록하고 조립하는지의 문제이고, AOP proxy는 메서드 호출이 proxy 경계를 지나야 advice나 transaction이 적용된다는 문제입니다. Servlet MVC와 WebFlux는 HTTP 요청을 어떤 thread/event loop 흐름으로 이어 가는지가 다릅니다. `@Transactional`도 annotation 자체가 아니라 transaction manager, connection binding, commit/rollback 경계가 함께 움직여야 의미가 있습니다.

검증 anchor는 startup log, bean definition, proxy class, actuator metrics, thread dump, connection pool metric, transaction log입니다. "Spring이 해 준다"는 말을 만나면 어떤 객체와 상태를 Spring이 소유하는지 바로 풀어야 합니다.

## AOP와 어노테이션 처리

### jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리

#### 원문: jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리

<!-- curriculum-chunk: sha256=e3ca203455c03868b03f52265b53e911636ae758fcf154c017684f2aac2425e5 major=spring-backend-frameworks mid=AOP와 어노테이션 처리 sub=jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리 sources=source/interview_questions.md:3733-3734, source/interviews.md:3733-3734 -->

> Source: `source/interview_questions.md:3733-3734`
> Classification reason: aop/annotation
> Duplicate source aliases: `source/interview_questions.md:3733-3734, source/interviews.md:3733-3734`

##### jvm 계열 언어에서 어노테이션 선언 방법 및 spring 어노테이션 동작 원리

<!-- /curriculum-chunk -->

### 어노테이션 프로세싱과 Spring의 내부 동작 원리

#### 원문: 어노테이션 프로세싱과 Spring의 내부 동작 원리

<!-- curriculum-chunk: sha256=b0686550ff1bce8b58fb5dc9cc0fb277a5ac2b30dba4f490ceffe0b88e35496e major=spring-backend-frameworks mid=AOP와 어노테이션 처리 sub=어노테이션 프로세싱과 Spring의 내부 동작 원리 sources=source/interview_questions.md:3509-3561, source/interviews.md:3509-3561 -->

> Source: `source/interview_questions.md:3509-3561`
> Classification reason: aop/annotation
> Duplicate source aliases: `source/interview_questions.md:3509-3561, source/interviews.md:3509-3561`

##### 어노테이션 프로세싱과 Spring의 내부 동작 원리

1. 어노테이션 기반 핸들러 매핑

   스프링은 `HandlerMapping` 인터페이스를 구현한 여러 클래스를 통해 URL 패턴과 컨트롤러 메서드를 매핑합니다.
   `@RequestMapping`, `@GetMapping` 등의 어노테이션이 컨트롤러 메서드에 붙어있다면, 이를 HandlerMethod로 등록합니다.

    ```java
    public class RequestMappingHandlerMapping extends AbstractHandlerMethodMapping<RequestMappingInfo> {
        @Override
        protected RequestMappingInfo getMappingForMethod(Method method, Class<?> handlerType) {
            RequestMapping requestMapping = AnnotationUtils.findAnnotation(method, RequestMapping.class);
            if (requestMapping != null) {
                // URL 패턴과 메서드를 매핑하는 로직
                return createRequestMappingInfo(requestMapping, method);
            }
            return null;
        }
    }
    ```

   `RequestMappingHandlerMapping` 클래스는 `@RequestMapping`이나 `@GetMapping`이 붙은 메서드를 스캔하고, URL 패턴과 매핑 정보를 `RequestMappingInfo` 객체로 저장합니다.
   이 정보는 나중에 `DispatcherServlet`이 요청을 처리할 때 사용됩니다.

2. `@RestController`가 동작하는 방식

   `@RestController`는 기본적으로 `@Controller` + `@ResponseBody`의 역할을 수행합니다.
   즉, 해당 클래스가 HTTP 요청을 처리하는 컨트롤러임을 나타내고, 반환되는 값이 직렬화되어 클라이언트로 반환됨을 의미합니다. 주로 JSON 형식으로 응답을 반환하는 RESTful 웹 서비스에서 사용됩니다.

    ```java
    @Target({ElementType.TYPE})
    @Retention(RetentionPolicy.RUNTIME)
    @Documented
    @Controller
    @ResponseBody
    public @interface RestController {
        // RestController는 결국 두 개의 어노테이션이 결합된 형태
    }
    ```

자바 어노테이션은 런타임에 유지되며, 바이트 코드 수준에서는 클래스나 메서드 위에 메타데이터로 저장됩니다.
이는 리플렉션을 통해 런타임에 인식할 수 있습니다.
스프링은 이를 사용하여 해당 클래스나 메서드가 어떤 역할을 하는지 동적으로 처리합니다.

리플렉션을 통해 Spring Framework는 클래스와 메서드에 붙어 있는 어노테이션을 분석하고, 이를 통해 적절한 빈 등록, 핸들러 매핑 등을 수행합니다.

```java
Method method = MyController.class.getMethod("hello");
RequestMapping mapping = method.getAnnotation(RequestMapping.class);
```

위 코드에서와 같이, 스프링은 리플렉션을 이용해 메서드나 클래스에 붙은 어노테이션을 탐색하고, 그에 맞는 동작을 수행합니다.

<!-- /curriculum-chunk -->

## Bean 생명주기와 선택

### `@Bean` vs. `@Component`

#### 원문: `@Bean` vs. `@Component`

<!-- curriculum-chunk: sha256=278f51f50c41f300905e44d02a69995ae67d858ce7aac742c63dbcec887bf506 major=spring-backend-frameworks mid=Bean 생명주기와 선택 sub=`@Bean` vs. `@Component` sources=source/interview_questions.md:9752-9936, source/interviews.md:9700-9884 -->

> Source: `source/interview_questions.md:9752-9936`
> Classification reason: bean lifecycle
> Duplicate source aliases: `source/interview_questions.md:9752-9936, source/interviews.md:9700-9884`

##### `@Bean` vs. `@Component`

스프링에서 빈(Bean)을 등록하는 두 가지 주요 방식인 `@Bean`과 `@Component`는 개념적으로 유사하지만 동작 방식, 사용 목적, 내부 처리 과정에서 큰 차이점이 존재합니다.

- `@Component`: 자동 빈 등록

  `@Component`는 클래스 레벨에서 사용되며, 해당 클래스를 스프링 컨테이너가 자동으로 감지하여 빈으로 등록하는 역할을 합니다.
  이는 "Component Scanning" 이라는 기능에 의해 동작하며, 클래스가 특정 패키지 내에 존재하면 자동으로 빈으로 등록됩니다.

    - 핵심 특징
        - 클래스 단위로 선언하여 스캔 범위(`@ComponentScan`) 내에서 자동으로 감지됨
        - 개발자가 직접 빈의 생성 과정에 관여하지 않음
        - 스프링 애플리케이션의 주요 계층을 표현하는 확장 어노테이션 존재 (`@Service`, `@Repository`, `@Controller`)

    ```java
    @Component
    public class MyComponent {
        public void doSomething() {
            System.out.println("Component is working...");
        }
    }
    ```

  자동 감지(Auto-detection)에 의해 빈으로 등록됨

- `@Bean`: 수동 빈 등록

  `@Bean`은 메서드 레벨에서 사용되며, 해당 메서드가 반환하는 객체를 수동으로 스프링 컨테이너에 빈으로 등록하는 방식입니다.
  이 방식은 "Java 기반의 명시적 빈 등록" 방법으로, 주로 `@Configuration` 클래스에서 사용됩니다.

    - 핵심 특징
        - 메서드 단위에서 빈을 정의하고 직접 반환값을 생성함
        - 객체의 생성 로직을 세밀하게 제어할 수 있음
        - 외부 라이브러리 클래스나 객체를 스프링 빈으로 등록할 때 유용

    ```java
    @Configuration
    public class AppConfig {
        @Bean
        public MyService myService() {
            return new MyService();  // 객체 직접 생성
        }
    }
    ```

  개발자가 `new MyService()`를 직접 호출하여 빈을 수동으로 등록

- `@Bean`과 `@Component`의 동작 방식 분석

    - `@Component`의 내부 동작 (`Component Scanning`)

        1. Spring Boot가 실행되면 `@SpringBootApplication` 내부의 `@ComponentScan`이 특정 패키지를 스캔
        2. 패키지 내에서 `@Component`가 선언된 클래스를 찾음
        3. 해당 클래스를 인스턴스화하여 스프링 컨테이너에 빈으로 등록

        ```java
        @SpringBootApplication
        public class Application {
            public static void main(String[] args) {
                ApplicationContext context = SpringApplication.run(Application.class, args);

                MyComponent component = context.getBean(MyComponent.class);
                component.doSomething();
            }
        }
        ```

      실행 시 `MyComponent`가 자동으로 빈으로 등록됨

      자동 빈 등록 과정:
        - `ClassPathBeanDefinitionScanner`가 클래스패스를 검색하여 `@Component`가 붙은 클래스를 찾음
        - 찾은 클래스를 `BeanDefinition`으로 변환 후 `ApplicationContext`에 등록

    - `@Bean`의 내부 동작 (`Java 기반 수동 등록`)

        1. 스프링이 `@Configuration`이 선언된 클래스를 찾아 객체 생성
        2. 해당 클래스 내부의 `@Bean` 메서드를 실행하여 빈을 반환
        3. 반환된 객체를 컨테이너에 빈으로 등록
        4. 추후 같은 빈이 요청될 때 동일한 인스턴스를 반환 (싱글톤 유지)

        ```java
        @Configuration
        public class AppConfig {
            @Bean
            public DataSource dataSource() {
                HikariDataSource ds = new HikariDataSource();
                ds.setJdbcUrl("jdbc:mysql://localhost:3306/mydb");
                ds.setUsername("user");
                ds.setPassword("password");
                return ds;
            }
        }
        ```

      실행 흐름:
        - `AnnotationConfigApplicationContext`가 `@Configuration` 클래스를 감지
        - `@Bean` 메서드를 호출하여 반환값을 빈으로 등록
        - 스프링 컨테이너에서 `dataSource()` 호출 시 이미 생성된 빈을 반환

- `@Component` 확장 어노테이션 분석

  스프링은 `@Component`를 보다 명확한 역할로 구분하기 위해 확장 어노테이션을 제공합니다.

    - `@Service`

        - 비즈니스 로직을 담당하는 클래스에서 사용
        - 일반적으로 `Service Layer`의 빈으로 사용됨

        ```java
        @Service
        public class UserService {
            public String getUser() {
                return "User A";
            }
        }
        ```

    - `@Repository`

        - DAO(데이터 액세스) 계층에서 사용
        - 데이터 예외 처리를 스프링이 자동으로 처리하도록 지원

        ```java
        @Repository
        public class UserRepository {
            public User findById(Long id) {
                return new User(id, "User A");
            }
        }
        ```

    - `@Controller`

        - Spring MVC의 컨트롤러 역할
        - 웹 요청을 받아서 처리하는 프레젠테이션 계층에서 사용

        ```java
        @Controller
        public class UserController {
            @GetMapping("/user")
            public String getUser() {
                return "User A";
            }
        }
        ```

      > `@RestController`는 `@Controller` + `@ResponseBody` 조합으로 JSON을 직접 반환

- 언제 `@Bean`을 쓰고, 언제 `@Component`를 써야 할까?

  ✅ `@Component`를 사용할 때:
    - 애플리케이션 내부에서 직접 만든 클래스를 빈으로 등록할 때
    - 자동 감지 기능을 활용하여 깔끔한 코드 유지가 가능할 때
    - 서비스, 레포지토리, 컨트롤러 등 계층별 역할을 표현하고 싶을 때

  ✅ `@Bean`을 사용할 때:
    - 외부 라이브러리 객체를 빈으로 등록해야 할 때
    - 객체 생성 과정이 복잡하여 커스텀 로직을 포함해야 할 때
    - 동일한 객체를 여러 개의 빈으로 등록할 필요가 있을 때

- 빈 라이프사이클과 연계한 심층 분석

  📌 빈의 생성 과정
    1. 객체 생성 (`@Component`는 `Component Scan`, `@Bean`은 `@Configuration` 메서드 실행)
    2. 초기화 (`@PostConstruct` 및 `InitializingBean`)
    3. 사용
    4. 소멸 (`@PreDestroy` 및 `DisposableBean`)

    ```java
    @Component
    public class LifecycleComponent {
        @PostConstruct
        public void init() {
            System.out.println("빈 초기화");
        }

        @PreDestroy
        public void destroy() {
            System.out.println("빈 소멸");
        }
    }
    ```

  `@Component`든 `@Bean`이든 동일한 라이프사이클을 따름

<!-- /curriculum-chunk -->

### 스프링 부트 시작과 빈 라이프 사이클

#### 원문: 스프링 부트 시작과 빈 라이프 사이클

<!-- curriculum-chunk: sha256=9cc8c8e4d902f89eff1e54ce27f2826147b3a8d88094bb226f77b98b35fbbf22 major=spring-backend-frameworks mid=Bean 생명주기와 선택 sub=스프링 부트 시작과 빈 라이프 사이클 sources=source/interview_questions2.md:1630-1742, source/interviews2.md:1630-1742 -->

> Source: `source/interview_questions2.md:1630-1742`
> Classification reason: bean lifecycle
> Duplicate source aliases: `source/interview_questions2.md:1630-1742, source/interviews2.md:1630-1742`

##### 스프링 부트 시작과 빈 라이프 사이클

스프링 부트 애플리케이션은 일반적으로 `SpringApplication.run()` 메서드를 호출하여 시작됩니다.
이 과정은 크게 SpringApplication 초기화, 환경 설정, 애플리케이션 컨텍스트 생성 및 부트스트랩, 그리고 빈 초기화 및 실행으로 나뉩니다.

###### SpringApplication.run()

- `SpringApplication.run()` 호출

    1. `SpringApplication` 객체 생성:
        - 애플리케이션이 실행되면 `SpringApplication` 클래스의 정적 메서드 `run()`이 호출됩니다.
        - 내부적으로 `SpringApplication` 인스턴스가 생성되며, 애플리케이션의 실행 환경을 설정합니다.

    2. 스프링 부트 설정 초기화:
        - 기본적으로 실행될 애플리케이션 타입을 결정합니다.
            - `REACTIVE`: 웹플럭스 기반 애플리케이션.
            - `SERVLET`: 일반적인 웹 MVC 기반 애플리케이션.
            - `NONE`: 콘솔 기반 애플리케이션.
        - 실행 환경(`SpringApplicationEnvironment`)과 컨텍스트 타입(`ApplicationContext`)을 초기화.

- `prepareEnvironment()`

    1. 프로퍼티 및 설정 로드:
        - `application.properties` 또는 `application.yml` 파일을 읽어 Spring 환경에 추가.
        - 운영 체제 환경 변수, JVM 시스템 속성, 외부 프로퍼티 소스 등을 결합하여 스프링 환경(`Environment`) 객체에 저장.

    2. EnvironmentPostProcessor 호출:
        - 환경 설정을 커스터마이징할 수 있는 확장 포인트를 실행.

- `prepareContext()`

    1. ApplicationContext 생성:
        - 스프링 컨텍스트(ApplicationContext)가 생성됩니다. 기본적으로 다음 중 하나를 선택:
            - `AnnotationConfigServletWebServerApplicationContext`: 서블릿 기반 애플리케이션.
            - `ReactiveWebServerApplicationContext`: 웹플럭스 기반 애플리케이션.
            - `GenericApplicationContext`: 일반 콘솔 애플리케이션.

    2. BeanDefinition 로드:
        - `@ComponentScan` 및 `@EnableAutoConfiguration`에 의해 클래스 경로에서 빈 정의가 로드됩니다.

    3. ApplicationContextInitializer 호출:
        - `SpringApplication`에 등록된 초기화기(ApplicationContextInitializer)가 실행됩니다.
        - 컨텍스트를 추가적으로 초기화할 수 있는 확장 포인트.

- `refreshContext()`

    1. ApplicationContext의 `refresh()` 호출:
        - `ApplicationContext`의 핵심 메서드로, 빈 팩토리 초기화와 빈 라이프사이클의 시작점이 됩니다.
        - 아래에서 자세히 설명할 빈 라이프사이클의 대부분이 이 단계에서 진행됩니다.

    2. CommandLineRunner 및 ApplicationRunner 실행:
        - `refresh()` 이후, `CommandLineRunner`와 `ApplicationRunner` 인터페이스를 구현한 빈들이 실행됩니다.
        - 이 단계는 애플리케이션 로직을 실행할 수 있는 지점을 제공합니다.

- `run()` 종료
    - 스프링 부트 애플리케이션이 완전히 초기화되고 실행 준비가 완료됩니다.
    - 애플리케이션이 HTTP 요청을 수신 대기하거나 백그라운드 작업을 시작합니다.

###### 스프링 빈의 라이프사이클

스프링 컨테이너는 빈을 생성, 초기화, 사용, 소멸시키는 일련의 과정을 관리합니다. 이 과정은 다음과 같은 단계로 구성됩니다.

- 빈 정의 등록 (Bean Definition)
    1. 클래스 경로 스캔:
        - `@Component`, `@Configuration`, `@Service`, `@Repository`, `@Controller` 등이 선언된 클래스를 탐색.
        - 이러한 클래스는 빈 정의(Bean Definition)으로 컨테이너에 등록됩니다.

    2. 빈 팩토리 초기화:
        - 스프링 컨테이너는 빈 정의를 기반으로 빈 팩토리(BeanFactory)를 초기화합니다.
        - 빈 정의에는 빈의 클래스 타입, 생성 방법, 의존성 정보 등이 포함됩니다.

- 빈 생성 및 초기화

  빈 라이프사이클의 주요 단계는 다음과 같습니다:

    1. 빈 인스턴스 생성
        - 스프링 컨테이너가 빈 정의에 따라 빈 인스턴스를 생성합니다.
        - 기본적으로 리플렉션을 사용하여 객체를 생성하며, 생성자 주입 시 필요한 의존성을 해결합니다.

    2. 의존성 주입 (Dependency Injection)
        - 생성된 빈에 `@Autowired`, `@Value` 등을 통해 의존성을 주입합니다.
        - 의존성은 생성자, 세터 메서드, 또는 필드를 통해 주입될 수 있습니다.

    3. 빈 초기화 (Initialization)
        1. `Aware` 인터페이스 처리:
            - 빈이 특정 스프링 컨텍스트 정보에 접근해야 하는 경우 `Aware` 인터페이스를 구현합니다.
            - 예: `BeanNameAware`, `ApplicationContextAware`.
            - 스프링이 빈에 컨텍스트 정보를 주입합니다.

        2. `BeanPostProcessor`의 `postProcessBeforeInitialization` 호출:
            - 빈 초기화 전에 호출되는 사용자 정의 로직 실행.

        3. 초기화 메서드 실행:
            - `@PostConstruct`로 지정된 메서드 실행.
            - `InitializingBean` 인터페이스의 `afterPropertiesSet()` 호출.
            - XML이나 Java Config에서 `init-method`로 설정된 메서드 실행.

        4. `BeanPostProcessor`의 `postProcessAfterInitialization` 호출:
            - 초기화 후 추가 처리 작업을 수행.

- 빈 사용
    - 애플리케이션에서 빈이 사용되는 단계입니다.
    - 의존성 주입이 완료된 빈은 다른 빈이나 컴포넌트에 의해 사용됩니다.

- 빈 소멸 (Destruction)
    1. 소멸 전 작업:
        - `@PreDestroy`로 지정된 메서드 실행.
        - `DisposableBean` 인터페이스의 `destroy()` 호출.
        - XML이나 Java Config에서 `destroy-method`로 설정된 메서드 실행.

    2. 빈 제거:
        - 스프링 컨테이너가 빈에 대한 참조를 제거하여, 해당 객체가 GC(Garbage Collection) 대상이 되도록 만듭니다.

<!-- /curriculum-chunk -->

### 적절한 Bean 선택

#### 원문: 적절한 Bean 선택

<!-- curriculum-chunk: sha256=a1ed97fa860f4444dcb0f512eece507d599f79b495106fd2b7ab2e97964f0dcf major=spring-backend-frameworks mid=Bean 생명주기와 선택 sub=적절한 Bean 선택 sources=source/interview_questions.md:5098-5284, source/interviews.md:5098-5284 -->

> Source: `source/interview_questions.md:5098-5284`
> Classification reason: bean lifecycle
> Duplicate source aliases: `source/interview_questions.md:5098-5284, source/interviews.md:5098-5284`

##### 적절한 Bean 선택

Spring에서 DI(Dependency Injection)를 사용할 때, 한 인터페이스를 구현한 여러 클래스가 존재할 경우, 어떤 빈을 주입할지 결정하는 방법에는 여러 가지가 있습니다.
기본적으로, Spring은 인터페이스 타입으로 주입할 때 동일한 타입의 여러 빈이 있을 경우 주입할 빈을 결정할 수 있도록 다양한 메커니즘을 제공합니다.

1. `@Primary` 어노테이션 사용
2. `@Qualifier` 어노테이션 사용
3. 컬렉션 또는 맵 형태로 주입
4. `@Profile` 어노테이션을 통한 환경에 따른 빈 선택
5. `@Autowired`의 생성자 주입을 통한 명시적 빈 주입
6. Spring Expression Language(SpEL)를 사용하여 주입

각 방법을 구체적으로 살펴보겠습니다.

1. `@Primary` 어노테이션

   `@Primary` 어노테이션은 동일한 타입의 여러 빈이 있을 때, 우선적으로 주입할 빈을 지정하는 방법입니다.
   여러 구현 클래스가 존재하는 경우, 하나의 빈에 `@Primary`를 지정하면 Spring이 이 빈을 기본적으로 주입합니다.

    ```java
    public interface GreetingService {
        void greet();
    }

    @Service
    @Primary  // 기본적으로 주입할 빈
    public class EnglishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hello");
        }
    }

    @Service
    public class SpanishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hola");
        }
    }
    ```

   이 경우, `GreetingService` 타입으로 주입을 받을 때 `EnglishGreetingService`가 기본적으로 주입됩니다.

    ```java
    @Autowired
    private GreetingService greetingService;  // EnglishGreetingService가 주입됨
    ```

2. `@Qualifier` 어노테이션

   `@Qualifier`는 특정한 빈 이름을 지정하여 원하는 빈을 주입할 수 있는 방법입니다.
   인터페이스의 구현체가 여러 개 있을 때, 주입하려는 빈을 명시적으로 지정할 수 있습니다.

    ```java
    @Service("englishGreetingService")
    public class EnglishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hello");
        }
    }

    @Service("spanishGreetingService")
    public class SpanishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hola");
        }
    }
    ```

   이제 `@Qualifier`로 주입할 빈을 지정할 수 있습니다.

    ```java
    @Autowired
    @Qualifier("spanishGreetingService")
    private GreetingService greetingService;  // SpanishGreetingService가 주입됨
    ```

   이 방법은 매우 명시적이고, 다양한 상황에서 유용하게 사용됩니다.

3. 컬렉션 또는 맵 형태로 주입

   Spring은 동일한 타입의 모든 빈을 컬렉션으로 주입할 수 있습니다.
   이를 통해 여러 빈을 한꺼번에 주입받아 사용할 수 있습니다.

    ```java
    @Autowired
    private List<GreetingService> greetingServices;  // 모든 GreetingService 구현체가 주입됨

    public void greetAll() {
        greetingServices.forEach(GreetingService::greet);
    }
    ```

   이 경우, `greetingServices` 리스트에는 `EnglishGreetingService`와 `SpanishGreetingService` 두 빈이 모두 포함되어 순서대로 주입됩니다.

   또는, 맵(Map) 형태로 주입받을 수도 있습니다.

    ```java
    @Autowired
    private Map<String, GreetingService> greetingServiceMap;  // 빈 이름과 함께 주입됨

    public void greetByLanguage(String language) {
        GreetingService service = greetingServiceMap.get(language);
        service.greet();
    }
    ```

   맵으로 주입받으면, 빈 이름을 키로 하여 해당 빈을 사용할 수 있습니다.

4. `@Profile` 어노테이션을 통한 빈 선택

   `@Profile` 어노테이션을 사용하여 특정 환경(프로파일)에 맞는 빈을 주입할 수 있습니다.
   개발, 테스트, 프로덕션 등 프로파일에 따라 주입되는 빈을 달리할 때 유용합니다.

    ```java
    @Profile("english")
    @Service
    public class EnglishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hello");
        }
    }

    @Profile("spanish")
    @Service
    public class SpanishGreetingService implements GreetingService {
        @Override
        public void greet() {
            System.out.println("Hola");
        }
    }
    ```

   프로파일을 설정하면, 해당 환경에 맞는 빈이 주입됩니다.

    ```properties
    # application.properties
    spring.profiles.active=english
    ```

   이 설정을 사용하면, 현재 프로파일에 맞는 빈이 자동으로 주입됩니다.

5. `@Autowired`의 생성자 주입을 통한 명시적 빈 주입

   Spring에서는 생성자 주입을 통해 특정 구현체를 명확하게 주입할 수 있습니다.
   생성자에 `@Qualifier`를 사용하여 명시적으로 주입할 수 있습니다.

    ```java
    public class GreetingController {

        private final GreetingService greetingService;

        @Autowired
        public GreetingController(@Qualifier("spanishGreetingService") GreetingService greetingService) {
            this.greetingService = greetingService;
        }

        public void greet() {
            greetingService.greet();  // SpanishGreetingService 사용
        }
    }
    ```

   생성자를 사용한 주입은 의존성 주입의 명확성과 불변성을 유지하는 장점이 있습니다.

6. Spring Expression Language(SpEL)를 사용하여 빈 주입

   Spring에서는 SpEL(Spring Expression Language)을 사용하여 빈 주입 시 더욱 유연한 제어가 가능합니다.
   예를 들어, 조건에 따라 빈을 동적으로 주입할 수 있습니다.

    ```java
    @Autowired
    @Qualifier("#{systemProperties['language'] == 'es' ? 'spanishGreetingService' : 'englishGreetingService'}")
    private GreetingService greetingService;
    ```

   이 예시에서는 시스템 프로퍼티의 값에 따라 `spanishGreetingService` 또는 `englishGreetingService`를 동적으로 주입할 수 있습니다.

Spring에서 한 인터페이스를 구현한 여러 클래스가 존재할 때, 다양한 방법으로 알맞은 빈을 주입할 수 있습니다.
`@Primary`와 `@Qualifier`를 이용한 명시적 빈 주입 외에도 컬렉션 주입, `@Profile`, SpEL, 생성자 주입 등을 통해 다양한 시나리오에서 유연하게 DI(Dependency Injection)를 사용할 수 있습니다.

개발자가 `@Bean` 메서드로 수동 설정하지 않고도, 이와 같은 메커니즘을 활용하면 좀 더 자동화된 방식으로 상황에 맞는 빈을 적절히 주입할 수 있습니다.

<!-- /curriculum-chunk -->

## IoC와 DI

### Inversion of Control (IoC)

#### 원문: Inversion of Control (IoC)

<!-- curriculum-chunk: sha256=3dd832da926ab8337d0813e717d42d6871c6949796f3ff1fcdc13528673fc669 major=spring-backend-frameworks mid=IoC와 DI sub=Inversion of Control (IoC) sources=source/interview_questions.md:4980-5097, source/interviews.md:4980-5097 -->

> Source: `source/interview_questions.md:4980-5097`
> Classification reason: ioc/di
> Duplicate source aliases: `source/interview_questions.md:4980-5097, source/interviews.md:4980-5097`

##### Inversion of Control (IoC)

Inversion of Control (IoC)는 소프트웨어 엔지니어링에서 사용되는 디자인 원칙으로, *프로그램의 제어 흐름을 역전*시키는 것을 말합니다.
전통적인 프로그래밍에서는 프로그래머가 프로그램의 흐름을 제어하지만, IoC에서는 외부 프레임워크나 컨테이너가 그 역할을 담당합니다.
> 컨테이너: 객체의 생명주기와 의존성을 관리하는 프레임워크의 핵심 컴포넌트

IoC는 "할리우드 원칙(Don't call us, we'll call you)"으로도 알려져 있으며, *프로그램의 부품(컴포넌트)들이 프레임워크에 의해 관리되고 호출되는 구조*를 만듭니다.

주요 원칙:

- 제어의 역전: 프로그램의 제어 흐름이 프레임워크에 의해 관리됩니다.
- 의존성 관리: 컴포넌트 간의 의존성을 외부에서 관리합니다.
- 유연성: 프로그램 구조의 유연성과 확장성을 향상시킵니다.

의존성 주입 (DI)은 IoC를 구현하는 디자인 패턴 중 하나입니다.
서비스 로케이터는 필요한 서비스나 컴포넌트를 찾아주는 IoC 구현 방식이고,
팩토리 패턴은 객체 생성을 캡슐화하는 디자인 패턴으로 IoC 구현에 자주 사용됩니다.

IoC의 이론적 기반은 다음과 같은 소프트웨어 설계 원칙과 밀접하게 연관되어 있습니다:

```java
// 간단한 IoC 컨테이너 구현 예제

import java.util.HashMap;
import java.util.Map;

/*
 * 간단한 IoC 컨테이너 구현
 * 이 클래스는 객체의 생성과 의존성 주입을 관리하고 검색하는 기능을 제공합니다.
 */
public class SimpleIoCContainer {
    private Map<Class<?>, Object> container = new HashMap<>();

    /*
     * 객체를 컨테이너에 등록합니다.
     * @param clazz 등록할 객체의 클래스
     * @param instance 등록할 객체의 인스턴스
     */
    public void register(Class<?> clazz, Object instance) {
        container.put(clazz, instance);
    }

    /*
     * 컨테이너에서 객체를 검색합니다.
     * @param clazz 검색할 객체의 클래스
     * @return 검색된 객체 인스턴스
     * @throws IllegalArgumentException 객체가 등록되어 있지 않은 경우
     */
    public <T> T resolve(Class<T> clazz) {
        T instance = (T) container.get(clazz);
        if (instance == null) {
            throw new IllegalArgumentException("No instance registered for class: " + clazz.getName());
        }
        return instance;
    }
}

/*
 * 메시지 서비스 인터페이스
 * 메시지 전송 기능을 추상화한 인터페이스입니다.
 */
interface MessageService {
    void sendMessage(String message);
}

/*
 * 이메일 서비스 구현체
 * `MessageService`의 구체적인 구현입니다.
 */
class EmailService implements MessageService {
    @Override
    public void sendMessage(String message) {
        System.out.println("Sending email: " + message);
    }
}

/*
 * 사용자 서비스
 * `MessageService`에 의존하며, 이 의존성은 생성자를 통해 주입됩니다.
 * `UserService`가 `MessageService`의 구체적인 구현에 직접 의존하지 않고,
 * 컨테이너를 통해 의존성을 주입받음으로써 IoC 원칙을 실현합니다.
 */
class UserService {
    private MessageService messageService;

    // IoC 컨테이너에 의해 의존성이 주입됩니다.
    public UserService(MessageService messageService) {
        this.messageService = messageService;
    }

    public void notifyUser(String message) {
        messageService.sendMessage(message);
    }
}

/*
 * 메인 애플리케이션 클래스
 * IoC 컨테이너를 설정하고 사용하는 예를 보여줍니다.
 */
public class Application {
    public static void main(String[] args) {
        // IoC 컨테이너 생성
        SimpleIoCContainer container = new SimpleIoCContainer();

        // 객체 등록
        container.register(MessageService.class, new EmailService());
        container.register(UserService.class, new UserService(container.resolve(MessageService.class)));

        // 컨테이너에서 UserService 해결 및 사용
        UserService userService = container.resolve(UserService.class);
        userService.notifyUser("Hello, IoC!");
    }
}
```

1. 단일 책임 원칙 (SRP): IoC를 통해 각 컴포넌트는 자신의 핵심 기능에만 집중할 수 있습니다. 의존성 관리와 같은 부가적인 책임은 컨테이너가 담당합니다.
2. 개방-폐쇄 원칙 (OCP): IoC는 기존 코드를 수정하지 않고도 새로운 기능을 추가할 수 있게 해줍니다. 이는 프로그램을 확장에는 열려있고 수정에는 닫혀있게 만듭니다.
3. 의존성 역전 원칙 (DIP): IoC는 고수준 모듈이 저수준 모듈에 직접 의존하지 않고, 둘 다 추상화에 의존하게 함으로써 DIP를 실현합니다.
4. 인터페이스 분리 원칙 (ISP): IoC는 컴포넌트들이 필요한 인터페이스만을 알도록 돕습니다. 이는 시스템을 더 모듈화하고 유연하게 만듭니다.

<!-- /curriculum-chunk -->

### Spring Framework의 Dependency Injection (DI)

#### 원문: Spring Framework의 Dependency Injection (DI)

<!-- curriculum-chunk: sha256=e41fc19b35166b560cbac0e4f144a5d80fdfd4c3537eb9411d3385609f5bb5d7 major=spring-backend-frameworks mid=IoC와 DI sub=Spring Framework의 Dependency Injection (DI) sources=source/interview_questions.md:4874-4979, source/interviews.md:4874-4979 -->

> Source: `source/interview_questions.md:4874-4979`
> Classification reason: ioc/di
> Duplicate source aliases: `source/interview_questions.md:4874-4979, source/interviews.md:4874-4979`

##### Spring Framework의 Dependency Injection (DI)

Dependency Injection(DI)은 객체 간의 의존성을 외부에서 주입하는 디자인 패턴 및 프로그래밍 기법입니다.
객체가 자신의 의존성을 직접 생성하지 않고, 외부로부터 주입받습니다.

주요 원칙:

- 의존성 분리: 객체는 자신의 의존성을 직접 생성하지 않습니다.
- 설정의 외부화: 의존성 설정을 외부(주로 IoC 컨테이너)에서 관리합니다.
  > IoC 컨테이너: Bean의 생성, 설정, 관리를 담당하는 Spring의 핵심 컴포넌트
- Autowiring: 의존성을 자동으로 연결하는 Spring의 기능
- 인터페이스 기반: 구체적인 구현보다는 인터페이스에 의존합니다.

객체의 생성과 생명주기 관리를 개발자가 아닌 프레임워크가 담당한다는 Inversion of Control (IoC) 원칙에 기반합니다.

DI의 핵심 이론은 객체 지향 설계의 SOLID 원칙, 특히 의존성 역전 원칙(Dependency Inversion Principle)과 밀접하게 연관되어 있습니다.

1. 단일 책임 원칙 (SRP): DI를 통해 각 클래스는 의존성 관리라는 책임에서 벗어나, 본연의 기능에만 집중할 수 있습니다.
2. 개방-폐쇄 원칙 (OCP): DI를 사용하면 기존 코드를 수정하지 않고도 새로운 의존성을 주입할 수 있어, 확장에는 열려있고 수정에는 닫혀있는 설계가 가능합니다.
3. 의존성 역전 원칙 (DIP): DI는 고수준 모듈이 저수준 모듈에 직접 의존하지 않고, 둘 다 추상화(인터페이스)에 의존하게 함으로써 DIP를 실현합니다.

```java
/*
 * 메시지 서비스의 인터페이스를 정의합니다.
 * 이 인터페이스는 메시지 전송 기능을 추상화합니다.
 */
public interface MessageService {
    void sendMessage(String message, String recipient);
}

/*
 * 이메일을 통해 메시지를 전송하는 구체적인 구현 클래스입니다.
 * MessageService 인터페이스를 구현합니다.
 */
@Service
public class EmailService implements MessageService {
    @Override
    public void sendMessage(String message, String recipient) {
        // 이메일 전송 로직 구현
        System.out.println("Sending email to " + recipient + ": " + message);
    }
}

/*
 * 사용자 서비스 클래스입니다.
 * MessageService에 의존하며, 이를 생성자를 통해 주입받습니다.
 */
@Service
public class UserService {
    private final MessageService messageService;

    /*
     * UserService 생성자입니다.
     * @param messageService 주입받을 MessageService 구현체
     */
    @Autowired
    public UserService(MessageService messageService) {
        this.messageService = messageService;
    }

    /*
     * 사용자에게 알림을 보내는 메서드입니다.
     * @param user 알림을 받을 사용자
     * @param message 전송할 메시지
     */
    public void notifyUser(String user, String message) {
        messageService.sendMessage(message, user);
    }
}

/*
 * 애플리케이션의 진입점입니다.
 * Spring 애플리케이션 컨텍스트를 설정하고 UserService를 사용하는 예를 보여줍니다.
 */
@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        ApplicationContext context = SpringApplication.run(Application.class, args);
        UserService userService = context.getBean(UserService.class);
        userService.notifyUser("user@example.com", "Hello, Spring DI!");
    }
}
```

1. `MessageService` 인터페이스는 메시지 전송 기능을 추상화합니다.
2. `EmailService`는 `MessageService`의 구체적인 구현입니다.
3. `UserService`는 `MessageService`에 의존하며, 이를 생성자를 통해 주입받습니다.
4. `@Service` 어노테이션은 Spring에게 이 클래스들을 Bean으로 관리하라고 지시합니다.
5. `@Autowired` 어노테이션은 Spring에게 적절한 `MessageService` 구현체를 주입하라고 지시합니다.

이 구조는 `UserService`가 `MessageService`의 구체적인 구현에 의존하지 않고, 인터페이스에만 의존하게 함으로써 DIP를 실현합니다. 또한, `UserService`는 `MessageService`의 생성에 대해 알 필요가 없어 SRP를 따릅니다.

DI는 시스템 아키텍처에 다음과 같은 영향을 미칩니다:

1. DI는 시스템을 작고 독립적인 모듈로 분리하는 것을 촉진합니다.
2. 새로운 기능을 추가할 때, 기존 코드를 수정하지 않고 새로운 구현을 주입할 수 있습니다.
3. 의존성을 쉽게 모의 객체(mock)로 대체할 수 있어, 단위 테스트가 용이해집니다.
4. 코드 간의 결합도가 낮아져 유지보수가 쉬워집니다.

그러나 다음 사항을 고려해야 합니다:

- DI 프레임워크는 학습 곡선이 있으며, 초기 설정이 복잡할 수 있습니다.
- 런타임 시 의존성 해결로 인한 약간의 오버헤드가 발생할 수 있습니다.

참고 링크:

- [Spring Framework Documentation](https://docs.spring.io/spring-framework/docs/current/reference/html/core.html#beans-introduction): Spring의 공식 문서로, DI와 IoC에 대한 깊이 있는 설명을 제공합니다.
- [Martin Fowler's article on Dependency Injection](https://martinfowler.com/articles/injection.html): DI 개념의 원문 소개와 깊이 있는 분석을 제공합니다.

<!-- /curriculum-chunk -->

## Servlet/Tomcat 요청 처리

### NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우

#### 원문: NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우

<!-- curriculum-chunk: sha256=4a990f5630a43e40db2275eecf7d74c443dd67089293c1019ff8793b06192538 major=spring-backend-frameworks mid=Servlet/Tomcat 요청 처리 sub=NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우 sources=source/interview_questions.md:7178-7293, source/interviews.md:7126-7241 -->

> Source: `source/interview_questions.md:7178-7293`
> Classification reason: servlet request path
> Duplicate source aliases: `source/interview_questions.md:7178-7293, source/interviews.md:7126-7241`

##### NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우

NGINX와 Java Spring Web 애플리케이션이 함께 사용되는 경우, NGINX는 리버스 프록시(reverse proxy) 역할을 합니다.
클라이언트의 웹 요청을 받아서 Java Spring 애플리케이션 서버로 전달하고, 그 응답을 다시 클라이언트에게 전달하는 역할을 합니다.

1. 클라이언트 요청 (웹 브라우저, HTTP 클라이언트 등)

   클라이언트가 웹 브라우저나 HTTP 클라이언트를 사용하여 Java Spring Web 애플리케이션에 접근하려고 할 때,
   클라이언트는 HTTP 또는 HTTPS 요청을 전송합니다.
   이 요청은 인터넷을 통해 서버의 네트워크 인터페이스 카드(NIC)로 도달합니다.

2. 서버의 NIC가 네트워크 패킷을 수신

   NIC의 드라이버가 패킷을 커널의 네트워크 스택으로 전달합니다.

3. 서버의 커널(Network Stack)

   서버의 운영체제 커널은 네트워크 인터페이스 카드(NIC)에서 받은 요청 패킷을 TCP/IP 스택을 통해 처리합니다.

   커널의 네트워크 스택에서 수신된 이더넷 프레임을 IP 패킷으로 해석하고, IP 계층이 패킷들을 재조립하여 완전한 IP 데이터그램을 생성합니다.
   TCP 계층에서 IP 데이터그램의 페이로드를 처리하여 TCP 세그먼트로 디코딩합니다. 이 과정에서 IP 주소 및 포트 번호가 맞는지 확인합니다.
   TCP 계층에서는 데이터 무결성을 보장하기 위해 재전송, 패킷 순서 확인 등의 작업을 수행합니다.

   재구성된 TCP 세그먼트는 해당 포트(443)에 바인딩된 소켓의 수신 버퍼로 전달됩니다.

   443 포트는 계속 열려 있어야 새로운 연결을 받을 수 있습니다.
   이는 "리스닝 소켓"이라고 불립니다.
    - 서버의 443 포트에서 리스닝하는 소켓이 새 연결 요청을 받습니다.
    - 연결이 수립되면, 커널은 새로운 소켓을 생성합니다. 이 소켓은 임시 포트 번호(ephemeral port)를 할당받습니다.
    - 이 새 소켓은 특정 클라이언트와의 통신에 사용됩니다.

   따라서 443 포트는 새 연결을 위해 계속 열려 있고, 각 클라이언트 연결은 고유한 소켓과 임시 포트 번호를 가집니다.

4. NGINX 프로세스 깨우기

   커널은 연결된 소켓 버퍼에 데이터를 저장합니다.
   그리고 `epoll` 시스템 콜은 소켓에 데이터가 도착했음을 감지해 대기 중인 프로세스를 깨웁니다.
   이때, Nginx와 같은 웹 서버는 리스닝 소켓에서 대기하고 있다가, 클라이언트와의 새로운 연결을 위해 새로운 소켓(랜덤 포트)을 생성합니다.

5. NGINX 웹 서버 (리버스 프록시 역할)

   NGINX는 수신된 데이터를 HTTP 요청으로 파싱합니다.
   NGINX는 웹 서버이자 리버스 프록시로 설정되었기 때문에, HTTP 요청을 처리한 후, 특정 조건에 따라 요청을 백엔드 서버로 프록시할 수 있습니다.
   요청이 Spring Web App으로 프록시되어야 한다고 판단되면, NGINX는 upstream 서버(Spring App)로 요청을 전달할 준비를 합니다.

   이 과정은 보통 NGINX의 설정 파일(`nginx.conf`)에서 정의됩니다.

    ```conf
    # NGINX 설정 파일
    server {
        listen 80;
        server_name example.com;

        location / {
            proxy_pass http://localhost:8080;  # Java Spring 애플리케이션이 동작하는 포트로 프록시
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
    ```

   여기서, `proxy_pass` 지시어는 요청을 백엔드의 Java Spring 서버로 전달하는 역할을 합니다.

6. NGINX에서 Spring App으로 요청 전달

   NGINX는 클라이언트의 요청을 기반으로 새로운 HTTP 요청을 생성합니다.
   이 요청은 로컬호스트의 8080 포트로 전송됩니다 (Spring App이 리스닝 중인 포트).
   이 과정은 내부적으로 로컬 소켓 통신을 통해 이루어집니다.

7. 백엔드 서버 (Java Spring 애플리케이션 서버)

   NGINX가 리버스 프록시 역할을 할 때, HTTP/1.1 또는 HTTP/2 등의 프로토콜을 사용하여 요청을 백엔드로 전달합니다.

   이 백엔드 서버는 보통 Tomcat, Jetty, Undertow와 같은 서블릿 컨테이너에서 동작하는 Java 애플리케이션입니다.
   Spring Boot라면 내장된 Tomcat 서버가 8080 포트에서 요청을 수신합니다.
   요청은 Servlet 컨테이너를 통과하여 Spring의 DispatcherServlet에 도달합니다.

8. 서블릿 컨테이너 (Tomcat 등)

   Java Spring 애플리케이션은 서블릿 컨테이너(Tomcat, Jetty, Undertow) 내에서 동작합니다.
   서블릿 컨테이너는 Java 애플리케이션의 HTTP 요청을 처리하는 환경을 제공합니다.

   NGINX가 전달한 HTTP 요청이 서블릿 컨테이너의 포트(예: 8080)로 도달하면, 서블릿 컨테이너는 이 요청을 받아 해당하는 Spring 컨트롤러로 전달합니다.

9. Spring MVC 처리

   Spring 애플리케이션은 요청 URL, HTTP 메서드(GET, POST 등)에 따라 Spring MVC 컨트롤러에서 해당 요청을 처리합니다.

   Spring의 요청 처리 흐름:
    1. 요청이 `DispatcherServlet`에 도달합니다. `DispatcherServlet`은 Spring의 프론트 컨트롤러로, 모든 HTTP 요청을 중앙에서 처리합니다.
    2. `DispatcherServlet`은 요청에 맞는 컨트롤러 메서드를 찾습니다.
    3. 컨트롤러는 요청을 처리하고, 비즈니스 로직을 수행합니다. 필요할 경우, 서비스 계층 및 데이터베이스에 접근하여 데이터를 가져옵니다.
    4. 처리 결과를 모델(Model) 객체에 담고, 이를 뷰(View)에 전달합니다. 뷰는 클라이언트에 반환될 HTML, JSON, XML 등의 응답 데이터로 렌더링됩니다.

10. 응답 생성 및 반환

    Spring 애플리케이션이 요청을 처리한 후, 응답 데이터를 생성합니다.
    이 응답은 HTTP 응답으로 포맷팅되어 서블릿 컨테이너에 의해 클라이언트로 반환될 준비를 합니다.

    서블릿 컨테이너(Tomcat 등)는 이 HTTP 응답을 다시 NGINX로 전달합니다.

11. NGINX를 통한 응답 반환

    NGINX는 백엔드 서버(Java Spring 애플리케이션)로부터 받은 응답을 클라이언트에게 전달합니다.
    이 과정에서 NGINX는 캐싱을 할 수도 있고, 추가적인 HTTP 헤더를 조작할 수도 있습니다.

    최종적으로 NGINX는 클라이언트(웹 브라우저 등)에게 HTTP 응답을 반환합니다.

12. 클라이언트 응답 수신

    클라이언트는 NGINX로부터 반환된 HTTP 응답을 받습니다.
    이 응답은 HTML, JSON, 또는 다른 포맷으로 되어 있으며, 브라우저는 이를 렌더링하거나 처리합니다.
    이로써 전체 요청-응답 사이클이 완료됩니다.

<!-- /curriculum-chunk -->

### Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정

#### 원문: Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정

<!-- curriculum-chunk: sha256=570de8383d142addbf85d9dc4cd4f5e67696bbd9f844afe009ad056d92670844 major=spring-backend-frameworks mid=Servlet/Tomcat 요청 처리 sub=Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정 sources=source/interview_questions.md:3313-3508, source/interviews.md:3313-3508 -->

> Source: `source/interview_questions.md:3313-3508`
> Classification reason: servlet request path
> Duplicate source aliases: `source/interview_questions.md:3313-3508, source/interviews.md:3313-3508`

##### Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정

Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정은 여러 단계로 이루어져 있습니다.
이 과정에서 중요한 요소는 Tomcat의 서블릿 컨테이너와 Spring 프레임워크의 동작 방식입니다.

1. Nginx에서 Tomcat으로의 요청 전달: Nginx는 리버스 프록시로 동작하여 HTTP 요청을 Tomcat으로 전달합니다.
2. Tomcat에서 서블릿 컨테이너로 요청 처리: Tomcat의 서블릿 컨테이너가 HTTP 요청을 받고 서블릿으로 전달합니다.
3. Spring `DispatcherServlet`이 요청을 처리: Spring의 핵심 컴포넌트인 `DispatcherServlet`이 요청을 받아 적절한 컨트롤러로 요청을 라우팅합니다.
4. 컨트롤러 메서드가 실행됨: `@RestController`나 `@Controller`가 요청을 처리하고, 결과를 반환합니다.

이제 각 단계를 더 자세히 살펴보면 다음과 같습니다.

1. Nginx에서 Tomcat으로 요청 전달

    - Nginx는 리버스 프록시로서 클라이언트의 HTTP 요청을 받아 내부적으로 Tomcat 서버로 전달합니다.
    - 일반적으로 Nginx는 `proxy_pass` 지시어를 사용하여 Tomcat이 리스닝 중인 포트(기본적으로 8080)로 요청을 전달합니다.

    ```nginx
    server {
        listen 80;

        location / {
            proxy_pass http://localhost:8080;  # 요청을 Tomcat(8080)으로 전달
            proxy_set_header Host $host;       # 원본 호스트 헤더 유지
            proxy_set_header X-Real-IP $remote_addr;  # 클라이언트 IP 전달
        }
    }
    ```

   여기서 Nginx는 리버스 프록시(Reverse Proxy) 로 동작하면서, 클라이언트의 HTTP 요청을 받아 내부적으로 Tomcat 등의 백엔드 서버로 전달하는 역할을 수행합니다.

    ```sh
    클라이언트 → (소켓 연결) → Nginx → (소켓 연결) → Tomcat
    ```

    1. 클라이언트 → Nginx
        - 클라이언트가 Nginx(80번 포트)로 HTTP 요청을 보냄.
        - Nginx는 listen 80; 설정에 따라 80번 포트에 소켓을 바인딩하여 대기.(listen 상태로 대기)
        - `accept()` 호출을 통해 클라이언트의 연결을 처리.

        ```sh
        # lsof -i :80
        nginx    12345 www-data  6u  IPv4 123456789  0t0  TCP *:80 (LISTEN)
        nginx    12345 www-data  7u  IPv4 123456790  0t0  TCP 192.168.1.100:80->192.168.1.50:53214 (ESTABLISHED)
        ```

    2. Nginx → Tomcat
        - Nginx는 새로운 소켓을 생성하여 Tomcat(8080)으로 요청을 전달.
        - 즉, Nginx가 `proxy_pass`를 수행할 때, Nginx가 클라이언트 역할이 되고, Tomcat이 서버 역할이 됨.
        - 이때, Nginx는 HTTP 클라이언트처럼 Tomcat에 새로운 TCP 연결을 맺음.
            - Nginx는 백엔드 서버(Tomcat)로 새로운 HTTP 요청을 생성.
            - 새로운 TCP 소켓을 생성하여 Tomcat(8080)과 연결.
            - Nginx는 클라이언트 요청을 Tomcat으로 전달 (`write()` 호출).
            - Tomcat이 응답을 보내면, Nginx는 이를 클라이언트에게 다시 전달.

        ```sh
        # lsof -i :8080
        tomcat   23456 tomcat   8u  IPv4 234567890  0t0  TCP *:8080 (LISTEN)
        tomcat   23456 tomcat   9u  IPv4 234567891  0t0  TCP 127.0.0.1:8080->127.0.0.1:55555 (ESTABLISHED)
        ```

   이처럼 HTTP 기반 프록시(proxy_pass)는 HTTP 헤더를 재구성하고, 요청을 수정할 수 있습니다.

   HTTP 요청을 새로 만들지 않고, 소켓을 통해 직접 전달하는 방식이 가능합니다.
    - AJP (Apache JServ Protocol)
    - FastCGI
    - gRPC 또는 WebSocket을 활용한 직접 통신
    - Unix Domain Socket (UDS) 사용

        ```nginx
        server {
            listen 80;

            location / {
                # /var/run/tomcat.sock → Tomcat이 수신하는 Unix 소켓.
                # 네트워크 스택을 거치지 않으므로 TCP보다 빠르게 로컬 프로세스 간 통신 가능.
                proxy_pass http://unix:/var/run/tomcat.sock;
            }
        }
        ```

   다만, TCP 소켓을 직접 사용하면, HTTP 헤더를 재구성하는 과정 없이 바로 요청을 전달할 수 있지만, 요청 정보(예: 헤더, 쿠키, 세션 관리)를 Nginx에서 수정할 방법이 사라질 수도 있습니다.

2. Tomcat 서블릿 컨테이너로 요청 처리

   Tomcat은 서블릿 컨테이너로서 자바 웹 애플리케이션을 실행하는 플랫폼입니다.
   Tomcat은 HTTP 요청을 받아 서블릿(Servlet)으로 전달하고, 서블릿은 클라이언트의 요청을 처리합니다.

    - HTTP 요청 수신: Tomcat은 클라이언트(Nginx)를 통해 들어온 HTTP 요청을 받습니다.
    - 서블릿에 요청 전달: Tomcat은 서블릿 컨테이너를 통해 요청을 관리하며, 서블릿 컨테이너는 요청을 서블릿으로 전달합니다.

   톰캣이 시작될 때, `web.xml` 파일이나 자바 설정 클래스를 통해 `DispatcherServlet`이 초기화됩니다.
   이 서블릿은 Spring MVC의 핵심 역할을 수행하는 프론트 컨트롤러로, 모든 HTTP 요청을 가로채서 처리할 컨트롤러를 찾아 호출하는 역할을 합니다.

    - Spring MVC 애플리케이션에서는 `DispatcherServlet`이 프론트 컨트롤러로 동작하며, 모든 요청을 받아들이는 중앙 허브 역할을 합니다.
    - Tomcat은 `DispatcherServlet`이 모든 경로를 처리하도록 설정되어 있으므로, 모든 요청은 `DispatcherServlet`으로 전달됩니다.

    ```xml
    <servlet>
        <servlet-name>dispatcher</servlet-name>
        <servlet-class>org.springframework.web.servlet.DispatcherServlet</servlet-class>
    </servlet>

    <servlet-mapping>
        <servlet-name>dispatcher</servlet-name>
        <url-pattern>/</url-pattern>
    </servlet-mapping>
    ```

   이 설정에서 `DispatcherServlet`이 모든 요청(모든 URL 경로에 대해 `/*`)을 받는 역할을 하며, 이후의 처리는 Spring MVC에 의해 이루어집니다.

3. Spring의 `DispatcherServlet`이 요청 처리

   `DispatcherServlet`은 프론트 컨트롤러 패턴을 구현한 클래스입니다.
   모든 HTTP 요청은 Tomcat 서블릿 컨테이너에 의해 `DispatcherServlet`으로 전달됩니다.
   `DispatcherServlet`은 다음과 같은 역할을 수행합니다:

    1. 핸들러 매핑(Handler Mapping):

       들어온 요청을 처리할 핸들러(Controller 메서드)를 찾습니다.
       이때, 요청 URL과 매핑된 컨트롤러를 찾기 위해 `HandlerMapping` 인터페이스를 사용합니다.
       Spring은 `@RequestMapping`과 같은 어노테이션을 이용해 URL 경로와 컨트롤러 메서드를 매핑합니다.

    2. 핸들러 어댑터(Handler Adapter):

       매핑된 핸들러를 호출하기 위해 `HandlerAdapter`를 사용하여 적절한 컨트롤러 메서드를 호출합니다.

        ```java
        public void doDispatch(HttpServletRequest request, HttpServletResponse response) throws Exception {
            // DispatcherServlet에서 요청을 처리할 핸들러 찾기
            HandlerExecutionChain handler = getHandler(request);

            // 핸들러 실행
            if (handler != null) {
                handle(request, response, handler);
            }
        }
        ```

    3. 메서드 실행 및 결과 반환:

       컨트롤러 메서드가 실행되어 비즈니스 로직을 처리하고, 그 결과를 반환합니다.
       `@RestController`의 경우, 메서드의 반환값은 JSON 형식으로 자동 직렬화되어 클라이언트로 전송됩니다.

       요청을 처리한 후, 뷰(View)가 필요할 경우 `ViewResolver`를 통해 적절한 뷰를 선택하여 렌더링합니다. (예: JSP, Thymeleaf 등)

   `DispatcherServlet`이 요청을 받아 적절한 컨트롤러 메서드를 호출하는 것이 Spring MVC의 핵심 동작 방식입니다.

   `DispatcherServlet`이 초기화되면 Spring의 `ApplicationContext`도 함께 초기화됩니다.
   이때 Spring은 클래스패스에 존재하는 모든 컴포넌트(즉, `@Controller`, `@RestController`, `@Service`, `@Repository`가 붙은 클래스)를 스캔하고, 이들을 빈(Bean)으로 등록합니다.
   `ClassPathScanningCandidateComponentProvider`와 같은 컴포넌트 스캐닝 메커니즘을 사용하여 클래스패스에서 적절한 빈(Bean)들을 찾고, Reflection을 이용해 `@RestController` 등의 어노테이션이 붙은 클래스를 인식합니다.

    ```java
    // ClassPathScanningCandidateComponentProvider 클래스를 이용한 스캔
    public Set<BeanDefinition> findCandidateComponents(String basePackage) {
        Set<BeanDefinition> candidates = new LinkedHashSet<>();
        // 클래스패스에서 @Component, @Controller 등의 어노테이션이 붙은 클래스를 스캔
        for (String className : this.scanner.findCandidateComponents(basePackage)) {
            candidates.add(loadBeanDefinition(className));
        }
        return candidates;
    }
    ```

   `@GetMapping("/hello")`는 `/hello` 경로에 대한 GET 요청을 처리하도록 설정된 컨트롤러 메서드입니다.
   Spring의 HandlerMapping이 이 매핑 정보를 이용해 `/hello` 경로에 대한 요청을 `hello()` 메서드로 전달하게 됩니다.

    ```java
    @RestController
    public class MyController {

        @GetMapping("/hello")
        public String hello() {
            return "Hello, World!";
        }
    }
    ```

4. 컨트롤러 메서드가 실행되고 응답 반환

   Spring은 어노테이션 기반 설정을 통해 각종 빈(Bean)과 매핑을 자동으로 구성합니다.

   예를 들어, `@RestController`는 스프링이 해당 클래스를 컨트롤러로 인식하도록 하고,
   `@GetMapping`, `@PostMapping` 등의 어노테이션은 HTTP 메서드와 URL 경로를 처리할 메서드를 지정합니다.

    - 핸들러 찾기: `DispatcherServlet`은 `HandlerMapping`을 통해 적절한 컨트롤러를 찾아 해당 메서드를 호출합니다.
    - 메서드 실행: 찾은 핸들러 메서드를 실행하여 클라이언트 요청을 처리합니다.
    - 응답 반환: 메서드가 결과를 반환하면, Spring은 그 결과를 HTTP 응답으로 만들어 클라이언트에게 전달합니다.

   응답 처리 과정은 다음과 같습니다.

    - 뷰 리졸버(View Resolver): 만약 일반적인 `@Controller`라면, `ViewResolver`가 HTML, JSP 등의 템플릿을 렌더링하여 응답을 만듭니다.
    - JSON 응답: `@RestController`라면, 반환된 객체는 JSON으로 직렬화되어 클라이언트에 응답됩니다.

   예를 들어, `hello()` 메서드는 문자열 `"Hello, World!"`를 반환하며, 이 문자열이 JSON 응답으로 변환되어 클라이언트에 전달됩니다.

<!-- /curriculum-chunk -->

### Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정

#### 원문: Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정

<!-- curriculum-chunk: sha256=ef085883176d8f705c06f908b436ecfa80d2117d591fa53a9e27789f971c96c8 major=spring-backend-frameworks mid=Servlet/Tomcat 요청 처리 sub=Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정 sources=source/interview_questions.md:9937-10147, source/interviews.md:9885-10095 -->

> Source: `source/interview_questions.md:9937-10147`
> Classification reason: servlet request path
> Duplicate source aliases: `source/interview_questions.md:9937-10147, source/interviews.md:9885-10095`

##### Tomcat과 Spring의 동작 원리 및 WAR 파일 처리 과정

Apache Tomcat은 Java EE(Java Enterprise Edition) 환경에서 Servlet, JSP(JavaServer Pages), WebSocket 등을 실행하는 서블릿 컨테이너(Servlet Container) 입니다.
Spring Boot와 결합하여 내장형 웹 서버로 활용할 수도 있고, 독립 실행형 WAR(Web Application Archive) 파일을 배포하여 서블릿 기반 애플리케이션을 실행할 수도 있습니다.

- Tomcat의 구조 및 동작 원리

  Tomcat은 다음과 같은 주요 구성 요소로 이루어져 있습니다.

  | 구성 요소   | 설명                                                             |
        | ----------- | ---------------------------------------------------------------- |
  | Catalina    | Tomcat의 핵심 엔진. 모든 요청을 처리                             |
  | Connector   | 클라이언트 요청을 받아들이고, HTTP/HTTPS 프로토콜을 처리         |
  | Service     | 여러 개의 Connector를 포함하는 컨테이너                          |
  | Engine      | 하나의 `Catalina` 내에서 여러 개의 `Host`를 관리                 |
  | Host        | 하나의 Tomcat 인스턴스에서 여러 개의 웹 애플리케이션을 실행 가능 |
  | Context     | 각각의 웹 애플리케이션을 관리하는 컨텍스트 (`/example`, `/app`)  |
  | ClassLoader | `.war` 파일 내부의 클래스와 리소스를 로드                        |

  🔹 Tomcat의 요청 처리 흐름
  1️⃣ 클라이언트가 HTTP 요청 전송
  2️⃣ Connector가 요청을 받아 HTTP 프로토콜을 파싱
  3️⃣ Engine → Host → Context로 요청이 전달됨
  4️⃣ Context가 `web.xml` 또는 `@WebServlet`에 등록된 서블릿을 찾음
  5️⃣ 요청이 `DispatcherServlet`(Spring) 또는 서블릿으로 전달됨
  6️⃣ 응답을 생성하고, 클라이언트에게 반환

  📌 이때 Spring Boot에서는 `DispatcherServlet`이 요청을 받아 `Controller`에 매핑된 메서드를 호출함.

- Tomcat과 Spring의 동작 관계

  Tomcat이 Spring 애플리케이션을 실행하는 방식은 다음과 같습니다.
  Spring 기반 웹 애플리케이션은 Servlet 기반이므로 Tomcat이 제공하는 Servlet API를 활용하여 동작합니다.
  Spring Boot는 Tomcat을 내장 웹 서버(Embedded Tomcat) 로 활용할 수도 있고, 배포 가능한 WAR 파일을 생성하여 Tomcat에 배포할 수도 있음.

  🔹 Spring Boot(내장 Tomcat) 실행 흐름
  1️⃣ `java -jar myapp.jar` 실행
  2️⃣ Spring Boot가 내장형 Tomcat 서버를 시작 (`TomcatWebServer`)
  3️⃣ `SpringApplication.run()`이 실행되면서 `DispatcherServlet`이 등록됨
  4️⃣ HTTP 요청이 DispatcherServlet을 통해 Spring MVC 핸들러로 전달됨
  5️⃣ 요청이 처리되고 HTTP 응답이 반환됨

  🔹 배포형 WAR 실행 흐름 (외부 Tomcat 사용)
  1️⃣ Tomcat이 실행되면서 `webapps/` 디렉토리에 있는 WAR 파일을 자동 배포
  2️⃣ `web.xml` 또는 `@WebServlet` 기반으로 DispatcherServlet이 등록됨
  3️⃣ 요청이 들어오면 Tomcat의 Catalina → Host → Context 로 요청이 전달됨
  4️⃣ `DispatcherServlet`이 요청을 받아 `Controller`로 전달
  5️⃣ 결과를 다시 Tomcat을 통해 클라이언트에 반환

  📌 즉, 내장형 Tomcat을 사용할 때는 Spring Boot가 Tomcat을 직접 관리하지만, 배포형 WAR 파일을 사용할 때는 Tomcat이 Spring 애플리케이션을 관리함.

- WAR 파일의 내부 구조 및 Tomcat의 처리 과정

  WAR(Web Application Archive)는 Java 웹 애플리케이션을 배포하기 위한 압축 파일 포맷입니다.
  이 파일을 Tomcat의 `webapps/` 폴더에 배포하면 자동으로 해석되어 애플리케이션이 실행됩니다.

  📌 WAR 파일 내부 구조

    ```sh
    myapp.war
    │── META-INF/
    │   └── MANIFEST.MF
    │── WEB-INF/
    │   ├── web.xml
    │   ├── classes/
    │   │   ├── com/example/Main.class
    │   │   └── com/example/controller/UserController.class
    │   ├── lib/
    │   │   ├── spring-core.jar
    │   │   ├── spring-web.jar
    │   │   └── hibernate.jar
    │   └── applicationContext.xml
    │── static/
    │   ├── css/
    │   ├── js/
    │   └── images/
    │── index.html
    ```

    - META-INF/ → JAR/WAR의 메타데이터 (`MANIFEST.MF`)
    - WEB-INF/ → 웹 애플리케이션의 내부 설정 (`web.xml`, `classes/`, `lib/`)
    - WEB-INF/classes/ → 컴파일된 `.class` 파일 (Spring 컨트롤러 포함)
    - WEB-INF/lib/ → 외부 라이브러리 JAR 파일 (Spring, Hibernate 등)
    - static/ → 정적 리소스 (HTML, CSS, JS 등)

- Tomcat이 WAR 파일을 배포하는 과정

  Tomcat은 `webapps/` 디렉토리에 배포된 WAR 파일을 자동으로 해석하여 애플리케이션을 실행합니다.

  🔹 WAR 파일 배포 과정
  1️⃣ `webapps/` 폴더에 `.war` 파일이 복사됨
  2️⃣ Tomcat이 WAR 파일을 자동 해제(Extract)하여 `webapps/myapp/` 디렉토리 생성
  3️⃣ `WEB-INF/web.xml` 또는 `@WebServlet`을 기반으로 서블릿 초기화
  4️⃣ `DispatcherServlet`이 컨텍스트를 로드하고 Spring 애플리케이션이 실행
  5️⃣ 요청이 들어오면 Servlet Container가 요청을 `DispatcherServlet`으로 전달
  6️⃣ Spring MVC가 요청을 `Controller`로 라우팅하고 응답을 반환

  📌 결과적으로, Tomcat은 웹 애플리케이션을 실행하는 컨테이너 역할을 하며, Spring은 그 위에서 비즈니스 로직을 처리하는 프레임워크로 동작함.

- Tomcat의 WAR 파일 클래스 로딩 과정

  Tomcat은 WAR(Web Application Archive) 파일을 배포하면 내부적으로 여러 개의 `ClassLoader`를 사용하여 클래스를 로드하고 애플리케이션을 실행합니다.

    - Tomcat의 ClassLoader 계층 구조

      Tomcat은 자체적인 `ClassLoader` 계층 구조를 유지하며, 애플리케이션 클래스와 공유 라이브러리를 분리하여 로드합니다.

        1. Bootstrap ClassLoader: Java 기본 클래스를 로드합니다. (`$JAVA_HOME/lib`)
        2. System ClassLoader: JDK의 확장 클래스를 로드합니다. (`$JAVA_HOME/lib/ext`)
        3. Common ClassLoader: `common/lib`에 위치한 라이브러리를 로드합니다.
        4. Catalina ClassLoader: Tomcat의 핵심 실행 코드 (`catalina.jar` 등)를 로드합니다.
        5. Shared ClassLoader: 모든 웹 애플리케이션에서 공유되는 라이브러리를 로드합니다. (`shared/lib`)
        6. Webapp ClassLoader: 각 웹 애플리케이션의 클래스 및 라이브러리를 로드합니다. (`WEB-INF/classes`, `WEB-INF/lib`)

      이러한 계층 구조를 통해 Tomcat은 각 웹 애플리케이션이 독립적으로 실행되도록 보장합니다.

    - WAR 파일에서 클래스 로딩 프로세스

        1. WAR 파일 배포: Tomcat이 `webapps` 디렉토리 내에서 `.war` 파일을 감지하면 이를 자동으로 추출합니다.
        2. `WEB-INF/classes`에서 클래스 로딩: 애플리케이션의 `.class` 파일을 `WebappClassLoader`를 사용하여 로드합니다.
        3. `WEB-INF/lib`에서 JAR 로딩: 애플리케이션이 의존하는 JAR 파일들을 로드합니다.
        4. Parent Delegation Model: Tomcat의 `ClassLoader`는 부모 `ClassLoader`에서 클래스를 먼저 찾고, 없을 경우 애플리케이션의 클래스를 로드하는 부모 위임 모델을 따릅니다.
        5. Servlet 및 필터 초기화: `web.xml` 설정에 따라 서블릿과 필터를 로드하고 초기화합니다.

    - Parent Delegation Model의 예외
        - `conf/lib` 또는 `shared/lib`에 존재하는 라이브러리는 모든 애플리케이션에서 공유되므로, 특정 버전 충돌이 발생할 수 있음. `WEB-INF/lib`와 `common/lib`에서 중복된 라이브러리가 존재할 경우, `tomcat.util.scan.StandardJarScanner`를 조정하여 충돌을 방지할 수 있음.
        - `WEB-INF/lib`에 포함된 라이브러리가 우선 적용되도록 설정하면 충돌을 방지할 수 있음.

- Tomcat 내부의 요청 처리 및 ThreadPoolExecutor 구조

  Tomcat은 HTTP 요청을 처리하기 위해 내부적으로 스레드 풀(`ThreadPoolExecutor`)을 활용합니다.
  Tomcat의 요청 처리 흐름은 다음과 같습니다.

    1. 클라이언트 요청 수신: Tomcat의 `Connector`가 `ServerSocket`을 통해 요청을 받음.
    2. 요청을 `RequestProcessor`에 전달: Tomcat의 `Endpoint`가 요청을 처리할 스레드를 생성하거나 기존 스레드 풀에서 재사용.
    3. `ThreadPoolExecutor`에서 스레드 할당:
        - 코어 스레드 개수(`minSpareThreads`)만큼 미리 생성하여 대기.
        - 요청이 증가하면 `maxThreads`만큼 동적으로 확장.
        - 요청이 끝나면 스레드가 풀에 반환됨.
    4. 요청을 `HttpServlet`으로 전달: 요청이 적절한 서블릿으로 라우팅되어 실행됨.
    5. 응답 반환 및 스레드 해제: 처리된 응답이 클라이언트로 반환되며, 스레드는 다시 풀로 반환됨.

  Tomcat의 `server.xml`에서 스레드 풀(ThreadPoolExecutor) 설정을 조정할 수 있습니다.

    ```xml
    <Connector port="8080" protocol="HTTP/1.1"
            connectionTimeout="20000"
            redirectPort="8443"
            maxThreads="200"
            minSpareThreads="25"
            acceptCount="100"/>
    ```

    - `maxThreads`: 최대 동시 요청을 처리할 스레드 개수 (기본값: 200)
    - `minSpareThreads`: 미리 생성할 코어 스레드 개수 (기본값: 10~25)
    - `acceptCount`: 큐에서 대기할 수 있는 최대 요청 개수. `acceptCount`를 너무 크게 설정하면 메모리 사용량이 급격히 증가할 수 있으므로, 최적의 값(100~500)을 설정해야 함.

- Spring 기반 WAR 배포 실습

  Spring Boot 프로젝트를 `war` 파일로 패키징하여 Tomcat에 배포 가능하게 만들려면 다음을 수행해야 합니다.

    1. `pom.xml` 설정 변경

        ```xml
        <packaging>war</packaging>

        <dependencies>
            <!-- Spring Boot Starter -->
            <dependency>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-starter-web</artifactId>
            </dependency>

            <!-- Tomcat을 제공하는 대신 제공된 컨테이너에서 실행 -->
            <dependency>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-starter-tomcat</artifactId>
                <scope>provided</scope>
            </dependency>
        </dependencies>
        ```

    2. `SpringBootServletInitializer` 상속

        ```java
        @SpringBootApplication
        public class MyApplication extends SpringBootServletInitializer {
            @Override
            protected SpringApplicationBuilder configure(SpringApplicationBuilder builder) {
                return builder.sources(MyApplication.class);
            }
        }
        ```

    3. WAR 파일 생성

        ```sh
        mvn clean package
        ```

    4. Tomcat `webapps/`에 배포

        ```sh
        cp target/myapp.war /usr/local/tomcat/webapps/
        ```

- Tomcat은 Servlet 기반의 Java 웹 애플리케이션을 실행하는 컨테이너
- Spring Boot는 내장형 Tomcat을 활용하여 쉽게 배포 가능
- WAR 파일을 배포하면 Tomcat이 내부적으로 클래스를 로드하고 애플리케이션을 실행
- Spring의 `DispatcherServlet`이 요청을 처리하고, Tomcat은 서블릿 컨테이너 역할을 수행

<!-- /curriculum-chunk -->

### nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유

#### 원문: nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유

<!-- curriculum-chunk: sha256=45be648af6bcb3d9b08238cb0d2d1c58dcf8b56b113e8a58b03183809eabafd9 major=spring-backend-frameworks mid=Servlet/Tomcat 요청 처리 sub=nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유 sources=source/interview_questions.md:7331-7414, source/interviews.md:7279-7362 -->

> Source: `source/interview_questions.md:7331-7414`
> Classification reason: servlet request path
> Duplicate source aliases: `source/interview_questions.md:7331-7414, source/interviews.md:7279-7362`

##### nginx와 java/kotlin spring web app 사이에 tomcat이 필요한 이유

과거에는 NGINX나 Apache HTTP Server가 리버스 프록시 역할을 하고, 그 뒤에 Tomcat 같은 서블릿 컨테이너가 애플리케이션 서버로 동작하는 구조가 흔했습니다.

- NGINX -> Tomcat -> Spring 애플리케이션 구조

    1. 클라이언트가 웹 브라우저나 HTTP 클라이언트를 통해 HTTP/HTTPS 요청을 보냅니다.
    2. NGINX 리버스 프록시 역할:

       클라이언트의 요청은 NGINX(또는 Apache HTTP 서버)에 도착합니다.
       NGINX는 리버스 프록시로 동작하며,
        - 정적 리소스(CSS, JS, 이미지 등)는 직접 제공합니다.
        - 동적 요청(예: Java 애플리케이션 관련 요청)은 백엔드 서버인 Tomcat으로 전달합니다.

       NGINX는 클라이언트와 오픈한 TCP 연결을 유지하고, 클라이언트로부터 받은 요청을 백엔드 서버(Tomcat)로 전달하는 역할을 합니다.

    3. Tomcat 서블릿 컨테이너:

       Tomcat은 서블릿 컨테이너로서 동작합니다.

       > 서블릿 컨테이너는 Java EE 표준에서 규정한 서블릿 API를 구현한 서버입니다.
       > HTTP 요청과 응답을 관리하며, Spring 같은 프레임워크가 웹 애플리케이션의 요청-응답 라이프사이클을 처리할 수 있도록 합니다.

       서블릿 컨테이너는 Java 웹 애플리케이션의 서블릿(Servlet)을 실행하고, 요청을 처리하는 역할을 담당합니다.

       Tomcat은 HTTP 요청을 받아, 이를 Spring MVC 애플리케이션에서 등록된 컨트롤러로 전달합니다.

    4. Spring 애플리케이션:

       Tomcat은 요청을 Spring MVC 애플리케이션으로 전달하고, 애플리케이션은 요청에 맞는 비즈니스 로직을 수행하여 응답을 생성합니다.
       이때 Tomcat은 요청을 받고 응답을 클라이언트로 반환하기까지 HTTP 요청 처리의 중심이 됩니다.

    5. 응답 반환:

       Tomcat은 Spring 애플리케이션으로부터 받은 응답을 다시 NGINX로 전달하고, NGINX는 이를 클라이언트에게 최종적으로 반환합니다.

  톰캣이 중간에 있어야 했던 이유는 다음과 같습니다:
    - 서블릿 컨테이너 역할:

      과거에는 Spring 같은 웹 애플리케이션 프레임워크가 서블릿 기반으로 설계되었습니다.
      서블릿 컨테이너는 웹 요청을 처리할 수 있는 기반을 제공하며, 이는 Java EE 표준에 맞춰 만들어진 Tomcat과 같은 컨테이너에서 작동해야 했습니다.

    - 동적 웹 콘텐츠 처리:

      NGINX는 주로 정적 파일을 제공하는 데 최적화되어 있습니다.
      반면 Tomcat은 동적 웹 콘텐츠(JSP, 서블릿, Spring 컨트롤러)와 비즈니스 로직을 처리할 수 있는 기능을 제공했습니다.

    - Java 표준을 준수하는 실행 환경:

      Java EE에서 정의한 웹 애플리케이션 표준(서블릿, JSP 등)을 준수하기 위해 서블릿 컨테이너가 필요했습니다.
      톰캣은 이러한 표준을 제공하는 대표적인 컨테이너 중 하나였습니다.

- 현재 Spring Boot 환경에서의 동작 방식 (Java -jar 방식)

  현재는 Spring Boot와 같은 프레임워크가 등장하면서, 톰캣이 애플리케이션 서버로서 내장형(Embedded)으로 동작하는 방식이 일반적입니다.
  이 경우, Tomcat은 독립적인 애플리케이션 서버로 존재하지 않고, Spring Boot 애플리케이션에 내장된 서버로 동작하게 됩니다.

1. Java 애플리케이션 실행:

   Spring Boot 애플리케이션은 `java -jar spring_web_app.jar` 명령어로 실행됩니다.
   이 명령어는 Java 가상 머신(JVM)에서 Spring Boot 애플리케이션을 실행하는 명령어입니다.

   이 애플리케이션은 Spring Boot에 내장된 Tomcat(또는 Jetty, Undertow) 서버를 자동으로 시작하여 HTTP 서버로 동작하게 만듭니다.

2. 내장형 Tomcat 서버:

   내장된 Tomcat 서버는 특정 포트(기본 8080)에서 클라이언트의 HTTP 요청을 리슨합니다.
   Tomcat은 NGINX처럼 자체적인 HTTP 서버 기능을 제공하며, 클라이언트로부터 들어오는 요청을 처리합니다.
   NGINX가 없는 경우에도 Tomcat은 HTTP 서버로서 동작할 수 있습니다.

3. Spring 애플리케이션 요청 처리:

   Tomcat은 요청을 받아, 이를 Spring의 DispatcherServlet으로 전달합니다.
   Spring MVC는 요청 URL과 HTTP 메서드에 맞는 컨트롤러를 찾아서 요청을 처리하고, 응답을 생성합니다.
   비즈니스 로직을 처리한 후, Spring은 응답을 Tomcat으로 반환하고, Tomcat은 이를 클라이언트에게 다시 반환합니다.

4. NGINX와의 통합:

   NGINX가 있는 환경에서는 NGINX가 리버스 프록시 역할을 하여 클라이언트 요청을 받아, Spring Boot 애플리케이션으로 프록시할 수 있습니다.

   예를 들어, `proxy_pass http://localhost:8080;` 설정을 통해 NGINX는 백엔드 Spring Boot 서버로 요청을 전달합니다.

   이 과정에서 NGINX는 여전히 정적 리소스를 제공할 수 있고, Spring Boot는 동적 요청을 처리합니다.

<!-- /curriculum-chunk -->

## Spring Boot 실행 모델

### 2. Spring 관련 질문

#### 원문: 2. Spring 관련 질문

<!-- curriculum-chunk: sha256=9d899daa45c907a19bd94fb75a54841b31152e7c83f6140072ed09e1f26a1a7f major=spring-backend-frameworks mid=Spring Boot 실행 모델 sub=2. Spring 관련 질문 sources=source/interview_questions3.md:130-164 -->

> Source: `source/interview_questions3.md:130-164`
> Classification reason: spring boot

##### 2. Spring 관련 질문

###### 질문 4. Spring IoC(Inversion of Control)와 DI(Dependency Injection)은 어떤 원리로 동작하나요?

###### **답변**

1. **IoC(Inversion of Control)**란, 객체의 생성과 의존성 설정을 개발자가 직접 하지 않고, **Spring 컨테이너**가 대신 맡는 구조를 말합니다. 전통적 방법에서는 각 객체가 스스로 필요한 다른 객체를 생성하거나 찾아야 하지만, IoC를 사용하면 컨테이너가 Bean을 생성하고 연결(주입)해줍니다.
2. **DI(Dependency Injection)**는 "의존성 주입"이란 의미로, A 객체가 B 객체를 필요로 할 때, A 스스로 B를 생성하지 않고, 컨테이너가 A를 만들 때 B를 주입해주는 방식입니다. 예를 들어, `@Autowired`, `@Inject`, `@Resource` 어노테이션 등을 통해 다른 Bean이 자동으로 주입됩니다.
3. 스프링은 **Bean Definition**(XML 설정, 자바 Config, Component Scan 등)을 바탕으로 어떤 클래스를 어떤 빈(Bean)으로 등록할지 파악합니다. 그리고 빈을 생성할 때, 생성자/필드/setter를 통해 필요한 다른 빈들을 연결합니다.
4. 결국 IoC/DI를 통해 **“의존성 관리”**가 단순해지고, 객체 간 결합도가 낮아집니다. 테스트 코드 작성(단위 테스트)나 교체 가능성이 높아지며, 애플리케이션 구조가 유연해집니다.

---

###### 질문 5. Spring AOP(Aspect Oriented Programming)에서는 어떤 방식으로 공통 로직을 분리하나요?

###### **답변**

1. AOP(Aspect Oriented Programming)는 **여러 클래스나 메서드에 공통적으로 필요한 기능**(예: 로깅, 트랜잭션, 보안 검사)을 분리하는 기법입니다. 예를 들어, 모든 서비스 메서드 호출 전후에 로깅을 남기고 싶을 때 AOP가 유용합니다.
2. 스프링 AOP는 주로 **프록시(Proxy) 방식**으로 구현됩니다. 즉, 실제 객체(타겟 객체)를 감싸는 **프록시 객체**를 만들어, 메서드가 호출될 때 프록시가 먼저 관여하여, 부가기능(Advice)을 수행하고, 이후 실제 메서드를 호출하거나 호출을 가로챕니다.
3. 이러한 프록시는 JDK Dynamic Proxy(인터페이스 기반)나 CGLIB(클래스 상속) 등을 통해 생성됩니다. @Aspect로 정의된 클래스 안에 **Pointcut**(적용 대상)과 **Advice**(전/후/주변 로직)가 명시됩니다.
4. 예를 들어 `@Around` 어노테이션을 사용하면 메서드 호출 전, 후 로직을 쉽게 삽입할 수 있고, 예외 발생 시 특정 처리를 할 수도 있습니다.
5. AOP는 구현부와 부가기능을 분리하여 유지보수를 용이하게 하고, 로깅, 트랜잭션, 보안, 모니터링 등 횡단 관심사를 깔끔하게 처리하도록 돕습니다.

---

###### 질문 6. @Transactional 애노테이션이 어떻게 트랜잭션을 관리하나요?

###### **답변**

1. @Transactional은 Spring AOP의 한 사례로 볼 수 있습니다. 스프링은 @Transactional이 붙은 Bean 메서드를 호출할 때, **프록시**가介入하여 트랜잭션 매니저(PlatformTransactionManager)를 통해 DB 커넥션을 시작(또는 기존 트랜잭션에 참여)합니다.
2. 메서드가 정상적으로 종료되면, 프록시는 커밋을 호출하고, 예외가 발생하면 롤백을 호출합니다. 이렇게 DB 트랜잭션 범위를 자동으로 관리하므로, 개발자는 로직에 집중할 수 있습니다.
3. 다만, 자기 자신 메서드끼리 호출(예: this.internalCall()) 시에는 프록시가 관여하지 못하므로 트랜잭션이 적용되지 않을 수 있습니다. 이를 방지하기 위해서는 자기 자신이 아닌 다른 Bean이 호출하도록 구조를 설계하거나, AspectJ를 사용하는 방법 등이 있습니다.

---

<!-- /curriculum-chunk -->

### Spring boot에서 톰캣 실행 후 일어나는 일

#### 원문: Spring boot에서 톰캣 실행 후 일어나는 일

<!-- curriculum-chunk: sha256=7d2e7ceb6eeb2ae4e0444bee6bb64f64164b6295bb0c375fec7cce6c7556c0d8 major=spring-backend-frameworks mid=Spring Boot 실행 모델 sub=Spring boot에서 톰캣 실행 후 일어나는 일 sources=source/interview_questions.md:4754-4873, source/interviews.md:4754-4873 -->

> Source: `source/interview_questions.md:4754-4873`
> Classification reason: spring boot
> Duplicate source aliases: `source/interview_questions.md:4754-4873, source/interviews.md:4754-4873`

##### Spring boot에서 톰캣 실행 후 일어나는 일

Spring Boot는 내장 톰캣을 포함하여 독립 실행형 애플리케이션을 구동할 수 있습니다.
이를 위해 Spring Boot는 다양한 설정을 자동으로 처리하고, 애플리케이션이 정상적으로 시작될 수 있도록 여러 단계를 거칩니다.

1. Spring Boot 애플리케이션 시작

   Spring Boot 애플리케이션이 실행되면, `SpringApplication.run()` 메서드가 호출됩니다.
   이 메서드는 애플리케이션의 부트스트랩 과정에서 모든 설정과 컴포넌트 초기화를 담당하는 핵심 역할을 합니다.

    ```java
    @SpringBootApplication
    public class MyApplication {
        public static void main(String[] args) {
            SpringApplication.run(MyApplication.class, args);  // 애플리케이션 실행
        }
    }
    ```

   이 `SpringApplication.run()` 호출을 기점으로 아래 단계들이 순차적으로 이루어집니다.

2. 애플리케이션 컨텍스트 초기화

   Spring Boot는 애플리케이션 컨텍스트(Application Context)를 초기화합니다.
   여기서 애플리케이션 컨텍스트는 Spring의 IoC(제어의 역전) 컨테이너로, 애플리케이션의 모든 빈(bean)을 관리하고, 애플리케이션의 컴포넌트와 서비스들을 생성 및 초기화하는 역할을 합니다.

    - 빈 정의: `@Component`, `@Service`, `@Repository`, `@Controller` 등으로 선언된 빈들을 찾아 컨텍스트에 등록합니다.
    - 설정 파일 로딩: `application.properties` 또는 `application.yml` 파일에서 설정을 로드하고, 애플리케이션 속성을 초기화합니다.

3. 내장 톰캣 서버 생성 및 초기화

   Spring Boot는 내장된 톰캣 서버를 포함하고 있기 때문에, 별도의 서블릿 컨테이너 설치 없이 자체적으로 톰캣 서버를 시작합니다.
   이 과정에서 Spring Boot는 자동 설정을 통해 톰캣 서버 인스턴스를 생성하고 초기화합니다.

    1. `TomcatServletWebServerFactory` 생성

       Spring Boot는 자동 설정을 통해 `TomcatServletWebServerFactory`를 생성합니다.
       이 팩토리 클래스는 내장 톰캣의 설정과 생성을 담당하며, 서블릿 웹 서버(Servlet Web Server) 환경을 위한 기본 구성을 제공합니다.

        ```java
        @Bean
        public ServletWebServerFactory servletContainer() {
            return new TomcatServletWebServerFactory();
        }
        ```

    2. 톰캣 커넥터 설정
        - 커넥터는 톰캣이 외부 클라이언트와 통신할 수 있도록 네트워크 소켓을 열고, HTTP 요청을 받을 준비를 합니다.
        - 포트 설정: 기본적으로 톰캣은 8080 포트에서 리스닝을 시작합니다. 이는 `server.port` 속성을 통해 변경할 수 있습니다.

            ```properties
            # 기본적으로 8080이지만, 9090으로 변경 가능
            server.port=9090
            ```

    3. 톰캣 초기화 및 서블릿 컨테이너 시작

        - `TomcatServletWebServerFactory`는 톰캣의 인스턴스를 생성한 후, 이를 시작하는 과정에서 `Tomcat` 객체를 초기화합니다.
        - 서블릿 컨테이너 초기화: 톰캣은 서블릿 컨테이너로서 `DispatcherServlet`을 등록하고 초기화합니다. `DispatcherServlet`은 Spring MVC의 핵심 서블릿으로, 모든 HTTP 요청을 처리하고, 컨트롤러로 전달하는 역할을 합니다.

        ```java
        DispatcherServlet dispatcherServlet = new DispatcherServlet();
        ServletRegistration.Dynamic registration = servletContext.addServlet("dispatcher", dispatcherServlet);
        registration.setLoadOnStartup(1);
        registration.addMapping("/");
        ```

4. 서블릿과 필터 등록

   Spring Boot는 웹 애플리케이션에서 필요한 서블릿, 필터, 리스너들을 자동으로 등록합니다.
   특히, Spring MVC의 `DispatcherServlet`은 모든 요청을 처리하는 프론트 컨트롤러로 등록됩니다.

    1. DispatcherServlet 등록
        - `DispatcherServlet`은 클라이언트로부터 들어오는 모든 HTTP 요청을 수신하고 처리하는 프론트 컨트롤러입니다.
        - 핸들러 매핑과 뷰 리졸버를 사용해 클라이언트 요청을 적절한 컨트롤러 메서드로 매핑한 후, 결과를 렌더링합니다.

    2. 필터와 리스너 등록
        - Spring Security와 같은 기능을 사용하는 경우, 필터(Filter)도 자동으로 등록됩니다.
        - 필터는 HTTP 요청을 처리하기 전에 인증, 인가와 같은 작업을 수행할 수 있도록 해줍니다.
        - 필요에 따라 리스너(Listener)도 초기화되어 특정 이벤트(예: 세션 생성, 파괴 등)를 처리할 수 있습니다.

5. 애플리케이션 컨텍스트 빈 초기화

   Spring Boot는 애플리케이션 컨텍스트에 등록된 빈(bean)을 초기화하고, 필요한 의존성을 주입(DI)합니다.
   이는 애플리케이션 전반에서 사용되는 서비스, 리포지토리, 컨트롤러 등이 포함됩니다.

    1. 빈 주입 (DI)

       각 빈에 필요한 의존성이 자동으로 주입됩니다.
       이를 통해 애플리케이션의 모든 레이어가 연결되고, 의존성 주입을 통한 비즈니스 로직이 실행될 준비를 합니다.

    2. 애플리케이션 이벤트 발생

       Spring Boot는 애플리케이션 이벤트(Application Events)를 발생시키며, 특정 이벤트에 반응하는 리스너들을 호출합니다.
       예를 들어, `ApplicationReadyEvent`는 애플리케이션이 준비된 후 발생하여, 특정 작업을 시작할 수 있는 신호로 사용됩니다.

6. 톰캣 서버 시작 및 HTTP 요청 대기

   모든 설정이 완료되면, 톰캣 서버는 지정된 포트(기본 8080)에서 리스닝을 시작합니다.
   이 상태에서 서버는 HTTP 요청을 기다리며, 요청이 들어오면 `DispatcherServlet`을 통해 해당 요청을 처리합니다.

    - 클라이언트 요청이 들어오면, 톰캣은 HTTP 프로토콜을 통해 요청을 수신하고, 이를 Spring의 `DispatcherServlet`으로 전달합니다.
    - `DispatcherServlet`은 요청을 핸들러 매핑을 통해 적절한 컨트롤러로 라우팅하고, 처리된 결과를 클라이언트에 응답합니다.

7. 애플리케이션 상태 확인 및 모니터링 (Actuator)

   Spring Boot는 Actuator를 통해 애플리케이션의 상태를 모니터링하고 관리할 수 있는 기능을 제공합니다.
   Actuator는 다양한 엔드포인트(예: `/actuator/health`, `/actuator/metrics`)를 통해 애플리케이션의 상태와 메트릭스를 제공하며, 애플리케이션이 정상적으로 동작하는지 확인할 수 있습니다.

Spring Boot 내장 톰캣 실행 후 요약:

1. Spring Boot 애플리케이션 실행: `SpringApplication.run()`이 호출되면서 애플리케이션 실행이 시작됩니다.
2. 애플리케이션 컨텍스트 초기화: IoC 컨테이너가 초기화되며, 애플리케이션의 빈을 등록하고 설정 파일을 로드합니다.
3. 내장 톰캣 서버 생성 및 초기화: `TomcatServletWebServerFactory`가 톰캣 인스턴스를 생성하고, 포트와 커넥터 설정을 통해 HTTP 요청을 처리할 준비를 합니다.
4. 서블릿 및 필터 등록: `DispatcherServlet`과 필요한 필터 및 리스너들이 등록되고 초기화됩니다.
5. 빈 초기화 및 의존성 주입: Spring의 빈이 초기화되고, 의존성 주입이 완료됩니다.
6. 톰캣 서버 시작: 톰캣이 포트에서 리스닝을 시작하고, HTTP 요청을 기다립니다.
7. HTTP 요청 처리: `DispatcherServlet`이 요청을 수신하고, 컨트롤러를 호출하여 클라이언트에 응답을 반환합니다.
8. Actuator를 통한 상태 모니터링: 애플리케이션 상태를 실시간으로 모니터링할 수 있습니다.

<!-- /curriculum-chunk -->

### Spring의 동시성 및 Thread Management 정리

#### 원문: Spring의 동시성 및 Thread Management 정리

<!-- curriculum-chunk: sha256=3cbab0a1753af49827699909b90387d7648fff52690a8b6a961be1f9eac46abe major=spring-backend-frameworks mid=Spring Boot 실행 모델 sub=Spring의 동시성 및 Thread Management 정리 sources=source/interview_questions.md:10267-10595, source/interviews.md:10215-10543 -->

> Source: `source/interview_questions.md:10267-10595`
> Classification reason: spring boot
> Duplicate source aliases: `source/interview_questions.md:10267-10595, source/interviews.md:10215-10543`

##### Spring의 동시성 및 Thread Management 정리

- `ThreadLocal`이란?
    - `ThreadLocal<T>`은 각각의 스레드가 독립적인 변수 공간을 가지도록 보장하는 클래스입니다.
    - `ThreadLocal<T>`은 스레드별로 독립적인 변수 저장 공간을 제공하는 클래스.
    - 같은 `ThreadLocal` 인스턴스를 여러 스레드에서 참조하더라도, 각 스레드는 자신의 값만 저장하고 읽음.

    ```java
    public class ThreadLocalExample {
        private static ThreadLocal<Integer> threadLocal = ThreadLocal.withInitial(() -> 0);

        public static void main(String[] args) {
            Runnable task = () -> {
                threadLocal.set((int) (Math.random() * 100));
                System.out.println(Thread.currentThread().getName() + " : " + threadLocal.get());
            };

            new Thread(task).start();
            new Thread(task).start();
        }
    }
    ```

  객체(`new` 키워드로 생성된 인스턴스)는 항상 Heap에 저장됩니다.
  하지만 `ThreadLocal`이 관리하는 값(value)은 `ThreadLocalMap`이라는 특별한 저장소에 보관됩니다.
  `ThreadLocalMap` 자체는 각 `Thread` 객체 내부에서 관리되지만, Heap 메모리에 존재합니다.
  즉, `ThreadLocalMap`은 Heap에 존재하지만, 각 `Thread` 객체 내부에 포함되어 있으며,
  각 `Thread` 객체에 종속됩니다.

    ```java
    public class Thread {
        /* 각 스레드가 개별적으로 가지는 ThreadLocal 저장소 */
        ThreadLocalMap threadLocals;
    }
    ```

  📌 `ThreadLocal`이 값을 저장하는 과정
    1. Thread 객체는 Heap에 저장됨
        - `Thread` 객체는 `new Thread()`를 호출할 때 Heap에 생성됨.
    2. ThreadLocalMap도 Heap에 저장됨
        - `Thread` 객체의 필드(`threadLocals`)는 Heap에 저장된 `ThreadLocalMap`을 가리킴.
    3. ThreadLocal의 값도 Heap에 저장됨
        - `ThreadLocalMap`의 `Entry` 내부의 `value` 필드도 Heap에 저장됨.

  📌 동작 원리:
  `ThreadLocal.get()`을 호출하면, 현재 스레드의 `ThreadLocalMap`에서 값을 찾습니다.
  값이 없으면 `initialValue()`를 호출하여 값을 생성하고 저장합니다.
  이후 `get()`을 호출하면, 처음 저장된 값을 그대로 반환합니다.

  📌 사용 시 주의점은 다음과 같습니다:
  ✅ 메모리 누수 가능성
    - `ThreadLocal` 변수는 스레드가 종료될 때까지 `ThreadLocalMap`에 남아 있음.
    - 특히 WAS(Web Application Server) 환경에서 스레드 풀을 사용하면, 스레드가 재사용되므로 의도치 않은 값이 유지될 수 있음.
    - 해결 방법: 사용 후 반드시 `remove()` 호출

        ```java
        threadLocal.remove();
        ```

  ✅ 스레드 풀(Thread Pool)과 함께 사용 시 주의
    - 스레드가 재사용될 경우, 이전 요청의 `ThreadLocal` 값이 남아 있을 수 있음
    - 해결 방법: AOP 또는 Interceptor에서 `remove()` 자동 호출

  ✅ 싱글톤 빈과 함께 사용 시 주의
    - `ThreadLocal`이 싱글톤 빈 내부에서 사용되면, 여러 스레드가 공유하는 객체 내에서 독립적인 값을 유지하려는 의도와 충돌 가능성 존재.

  📌 `ThreadLocal`의 실무 적용 사례:
    - 사용자 세션 관리: 웹 애플리케이션에서 로그인한 사용자의 정보를 각 요청별로 유지하는 데 사용됨.

        ```java
        public class UserSession {
            private static final ThreadLocal<String> currentUser = new ThreadLocal<>();

            public static void setUser(String user) {
                currentUser.set(user);
            }

            public static String getUser() {
                return currentUser.get();
            }

            public static void clear() {
                currentUser.remove();
            }
        }
        ```

      ➡ 각 요청이 독립적으로 사용자 정보를 유지할 수 있음.

    - 트랜잭션 ID 관리: 마이크로서비스 환경에서 각 요청마다 고유한 트랜잭션 ID를 부여하고 유지하는 데 사용.

        ```java
        public class TransactionContext {
            private static final ThreadLocal<String> transactionId = ThreadLocal.withInitial(UUID::randomUUID);

            public static String getTransactionId() {
                return transactionId.get();
            }

            public static void clear() {
                transactionId.remove();
            }
        }
        ```

      ➡ 마이크로서비스에서 트랜잭션 ID를 로깅 및 추적하는 데 유용.

    - 데이터베이스 커넥션 관리
        - Spring의 `DataSource`는 `ThreadLocal`을 활용하여 각 스레드가 고유한 데이터베이스 커넥션을 유지하도록 보장.

- Spring에서 동시성 문제를 해결하는 방법

  📌 주요 동시성 문제는 다음과 같습니다:

    1. 싱글톤 빈(Singleton Bean)에서 상태 유지
        - Spring 컨테이너는 기본적으로 싱글톤 스코프로 빈을 관리.
        - 공유 필드가 있으면 여러 스레드가 접근하면서 동기화 문제가 발생.

    2. 비동기 처리에서 스레드 안전성 부족
        - Spring의 `@Async`는 기본적으로 새로운 스레드를 사용하지만, 공유 자원 접근 시 충돌 가능성 존재.

  📌 해결 방법

    - ✅ Stateless Design (무상태 설계): 상태를 가지는 공유 객체 대신, 메서드마다 독립적인 데이터를 사용.

        ```java
        @Service
        public class SafeService {
            public int calculate(int input) {
                return input * 2; // 상태를 가지지 않음
            }
        }
        ```

    - ✅ Synchronized 블록 또는 ReentrantLock 사용

        ```java
        public class Counter {
            private int count = 0;
            private final ReentrantLock lock = new ReentrantLock();

            public void increment() {
                lock.lock();
                try {
                    count++;
                } finally {
                    lock.unlock();
                }
            }
        }
        ```

    - ✅ `ThreadLocal`을 활용한 스레드별 데이터 저장

        ```java
        @Component
        public class UserContext {
            private static final ThreadLocal<String> userThreadLocal = new ThreadLocal<>();

            public void setUser(String user) {
                userThreadLocal.set(user);
            }

            public String getUser() {
                return userThreadLocal.get();
            }

            public void clear() {
                userThreadLocal.remove();
            }
        }
        ```

    - ✅ Spring의 Prototype Scope 활용: 멀티스레드 환경에서는 싱글톤 빈이 아니라 프로토타입 빈을 사용하는 것이 안전.

        ```java
        @Component
        @Scope("prototype")
        public class TaskProcessor {
            public void processTask() {
                // 매번 새로운 인스턴스가 생성되므로 동시성 문제 없음
            }
        }
        ```

- 싱글톤 빈과 멀티스레딩 문제

  📌 싱글톤 빈이 동시성 문제를 유발하는 이유:
    - Spring의 기본 빈 스코프는 `@Singleton`
    - 여러 스레드가 같은 인스턴스의 필드에 접근하면 동기화 문제가 발생.

  📌 해결책:

    1. 불변 객체(Immutable Object) 사용

        ```java
        @Service
        public class ImmutableService {
            private final String config;

            public ImmutableService() {
                this.config = "fixed value";
            }

            public String getConfig() {
                return config;
            }
        }
        ```

    2. `@Scope("prototype")`을 사용하여 빈을 스레드별로 생성

        ```java
        @Component
        @Scope("prototype")
        public class RequestScopedService {
            public String process() {
                return "New instance per request";
            }
        }
        ```

    3. 동기화 블록 적용

        ```java
        @Service
        public class SynchronizedService {
            private int counter = 0;

            public synchronized void increment() {
                counter++;
            }
        }
        ```

- Spring에서 비동기 프로그래밍 (`@Async`, `CompletableFuture`)

    - `@Async`를 활용한 비동기 처리: `@Async`는 별도의 스레드 풀에서 작업을 실행하여 비동기 처리를 가능하게 함.

        ```java
        @Service
        public class AsyncService {
            @Async
            public void asyncMethod() {
                System.out.println("Executing in thread: " + Thread.currentThread().getName());
            }
        }
        ```

    - 비동기 결과를 `CompletableFuture`로 반환

        ```java
        @Service
        public class AsyncService {
            @Async
            public CompletableFuture<String> asyncMethod() {
                return CompletableFuture.supplyAsync(() -> "Completed!");
            }
        }
        ```

    - 커스텀 스레드 풀 설정

        ```java
        @Configuration
        @EnableAsync
        public class AsyncConfig {
            @Bean(name = "customExecutor")
            public Executor asyncExecutor() {
                return Executors.newFixedThreadPool(5);
            }
        }
        ```

        ```java
        @Service
        public class CustomAsyncService {
            @Async("customExecutor")
            public void customAsyncMethod() {
                System.out.println("Executing in custom thread pool");
            }
        }
        ```

- `ExecutorService`와 Spring `TaskExecutor` 차이

  | 항목 | ExecutorService | Spring TaskExecutor |
        |-----------|----------------|----------------|
  | API | `java.util.concurrent.ExecutorService` | `org.springframework.core.task.TaskExecutor` |
  | 관리 방식 | 개발자가 직접 관리 | Spring이 관리 |
  | 설정 가능 여부 | `newFixedThreadPool()` 등으로 설정 가능 | `@EnableAsync` 및 `@Async`와 함께 사용 |
  | 주요 사용 예 | 일반적인 멀티스레딩 작업 | Spring의 비동기 작업 (`@Async`) |

  ✅ Spring에서 `ExecutorService` 대신 `TaskExecutor`를 사용해야 하는 이유
    - Spring 컨텍스트에서 관리되는 스레드 풀을 활용할 수 있음.
    - 애플리케이션 컨텍스트 종료 시, 자동으로 스레드 풀 정리 가능.

- Spring에서 Reactor와 Project Loom의 차이

    - Reactor (Spring WebFlux)
        - 리액티브 프로그래밍 기반의 논블로킹 I/O 처리
        - `Mono` 및 `Flux`를 사용하여 비동기 스트림 처리
        - Spring WebFlux와 함께 사용됨

        ```java
        Mono<String> mono = Mono.just("Hello, Reactor!")
                .map(String::toUpperCase);
        mono.subscribe(System.out::println);
        ```

    - Project Loom (Virtual Threads)
        - Java 21에서 추가된 경량 스레드(Virtual Threads) 기반의 동시성 처리
        - 기존의 `Thread` 객체와 유사하지만, OS 스레드가 아니라 JVM에서 관리됨.
        - 블로킹 작업을 효율적으로 수행할 수 있음.

        ```java
        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            executor.submit(() -> System.out.println("Hello from Loom"));
        }
        ```

    - Reactor vs. Project Loom 비교

      | 항목 | Reactor | Project Loom |
                    |------|--------|--------------|
      | 패러다임 | 논블로킹 리액티브 프로그래밍 | 전통적인 스레드 기반 프로그래밍 |
      | 사용 방식 | `Flux`, `Mono` 사용 | `Thread.ofVirtual()` 사용 |
      | 대상 애플리케이션 | 고성능 API 게이트웨이 | 일반적인 동시성 처리 |

<!-- /curriculum-chunk -->

### `java -jar SpringBootApp.jar` 명령어의 실행 과정

#### 원문: `java -jar SpringBootApp.jar` 명령어의 실행 과정

<!-- curriculum-chunk: sha256=f9346e51b36ebd858e4ecf416e24af9c09a37f790ae3ddd1bf687d3744efd49b major=spring-backend-frameworks mid=Spring Boot 실행 모델 sub=`java -jar SpringBootApp.jar` 명령어의 실행 과정 sources=source/interview_questions2.md:702-1119, source/interviews2.md:702-1119 -->

> Source: `source/interview_questions2.md:702-1119`
> Classification reason: spring boot
> Duplicate source aliases: `source/interview_questions2.md:702-1119, source/interviews2.md:702-1119`

##### `java -jar SpringBootApp.jar` 명령어의 실행 과정

###### 바이너리를 프로세스로 만드는 과정

사용자가 쉘에서 명령어(`java -jar`)를 입력하거나 프로그램이 `fork()` 또는 `exec()` 시스템 콜을 호출하면, 운영 체제에 새로운 프로세스를 생성하라는 요청이 전달됩니다.

`fork()` 시스템 콜은 부모 프로세스를 복제하여 자식 프로세스를 생성하는 기능을 합니다.

- PCB 생성

  부모 프로세스의 PCB를 기반으로 자식 프로세스의 PCB가 새롭게 만들어지고, 일부 정보가 복사됩니다.
  자식 프로세스는 고유한 PID를 가지며, 이를 포함한 정보를 새 PCB에 저장합니다

  부모 프로세스의 PCB에서 복사되는 주요 정보:
    - 프로세스 상태: 새 PCB는 Ready 상태로 초기화됩니다.
    - 프로세스 계층 구조: 부모-자식 관계가 기록되며, 부모 PID(PPID)가 저장됩니다.
    - 파일 디스크립터:
        - 부모 프로세스의 열린 파일 디스크립터가 복사됩니다.
        - 파일 디스크립터 테이블은 공유되지 않고 별도로 복사되어, 자식은 독립적으로 파일에 접근합니다.
    - 스케줄링 정보: 스케줄링 우선순위와 같은 정보가 복사됩니다.
    - 환경 변수 포인터: 환경 변수는 복사됩니다. 자식이 이를 수정하더라도 부모에게 영향을 미치지 않습니다.

  PCB에서 새롭게 할당되는 정보:
    - 새로운 PID: 자식 프로세스는 고유의 PID를 부여받습니다.
    - 프로세스 통계: CPU 사용 시간, 페이지 결함(page fault) 등은 초기화됩니다.

- 프로세스 메모리 공간 복제

  Copy-On-Write(CoW) 방식으로 복제가 이뤄집니다.
  부모 프로세스의 메모리 공간은 초기에는 복사되지 않습니다.

  자식 프로세스는 부모와 동일한 메모리 페이지를 공유하지만,
  어느 한쪽이 해당 페이지에 쓰기 작업을 시도하면, 운영 체제는 해당 페이지를 복사합니다.

  이를 통해 메모리 사용량을 줄이고, 프로세스 생성 속도를 크게 향상시킬 수 있습니다.

  복제되는 메모리 영역:
    - 코드 영역 (Code Segment): 부모와 자식이 읽기 전용으로 공유합니다. 일반적으로 변경되지 않으므로, 복사가 필요 없습니다.
    - 데이터 영역 (Data Segment): 초기에는 부모와 공유하지만, 쓰기 작업 시 복사됩니다.
    - 힙 영역 (Heap):
        - 부모와 공유하며, 쓰기 작업 시 복사됩니다.
        - 동적 메모리(malloc 등)가 여기에 포함됩니다.
    - 스택 영역 (Stack): 쓰기 작업이 많으므로, Copy-On-Write 시 가장 먼저 복사되는 경우가 많습니다.

  페이지 테이블 복사:
    - 부모 프로세스의 페이지 테이블은 자식 프로세스에 복사됩니다.
    - 페이지 테이블은 가상 주소에서 물리 주소로의 매핑 정보를 저장하며, 초기에 부모와 동일한 페이지를 참조하도록 설정됩니다.

- 파일 디스크립터 복제

  파일 디스크립터 테이블은 복제되며, 자식은 부모와 독립적으로 파일 작업을 수행할 수 있습니다.

    - 부모 프로세스에서 열려 있는 모든 파일 디스크립터가 자식 프로세스로 복사됩니다.
    - 파일 디스크립터 테이블은 복제되지만, 파일 오프셋은 공유됩니다.
        - 예를 들어, 부모가 파일을 읽거나 쓰면 자식이 동일한 오프셋에서 작업을 이어갑니다.
        - 필요시 `lseek` 등을 통해 독립적으로 설정 가능합니다.

새롭게 생성된 프로세스는 자식 프로세스가 되며, 부모와 거의 동일한 상태로 시작합니다.

그리고 복제된 프로세스는 고유한 프로세스 ID(PID)를 부여받습니다.
운영 체제는 고유한 PID를 관리하기 위한 PID 테이블을 유지합니다.

- 테이블은 사용 가능한 PID와 현재 사용 중인 PID를 추적합니다.
- PID는 일반적으로 1부터 시작하며, 사용 가능한 최대값은 시스템 설정에 따라 다릅니다(예: 32767 또는 4194304).
  새로운 프로세스가 생성되면, PID 테이블에서 사용 가능한 최소값을 선택하여 새 프로세스에 할당합니다.
  프로세스 종료 시 해당 PID는 다시 사용 가능 상태로 돌아갑니다.

이 시점에서 부모와 자식 프로세스는 동일한 프로그램을 실행하고 있습니다.

자식 프로세스는 `exec()` 시스템 콜을 호출하여 특정 프로그램 바이너리를 로드합니다.
이 과정에서 자식 프로세스의 메모리 공간은 새롭게 로드된 프로그램으로 덮어씌워집니다.

- 실행 파일의 코드 세그먼트 (명령어 집합).
- 초기 데이터 세그먼트 (전역 변수, 초기화된 데이터 등).
- 힙과 스택 영역 초기화.

이 과정에서 기존 메모리는 삭제되며, 로드된 프로그램이 새로운 프로세스의 실행 주체가 됩니다.

프로세스 생성 중에는 메모리가 할당됩니다.
운영 체제는 프로세스에 대해 가상 메모리 공간을 설정합니다.
가상 메모리는 페이징(Paging) 또는 세그먼테이션(Segmentation) 기법을 사용하여 실제 물리 메모리를 매핑합니다.
초기 설정은 다음과 같이 이뤄집니다:

- 코드 세그먼트: 실행 바이너리의 명령어가 로드됩니다.
- 데이터 세그먼트: 초기화된 전역 변수 및 정적 변수가 여기에 저장됩니다.
- 힙 영역: 런타임 동적 메모리 할당(`malloc`, `new`)에 사용됩니다.
- 스택 영역: 함수 호출 및 지역 변수 저장에 사용됩니다.

프로세스의 가상 주소를 물리적 메모리로 매핑하기 위한 페이지 테이블이 생성됩니다.
이 테이블은 프로세스 실행 중 메모리 접근 요청을 관리합니다.

부모 프로세스의 파일 디스크립터가 복제되며, 자식 프로세스는 동일한 파일에 접근할 수 있습니다.

프로세스가 생성되면, 운영 체제는 프로세스를 준비 상태(Ready State)로 설정합니다.
준비 상태에서 CPU 스케줄러가 프로세스를 선택하여 실행하게 됩니다.

CPU 스케줄러가 프로세스에 CPU를 할당하면, 프로세스는 실행 상태(Running State)로 전환됩니다.

프로세스가 I/O 작업을 기다리는 동안 대기 상태(Waiting State)로 전환됩니다.

마지막으로 프로세스가 작업을 완료하거나 오류로 인해 종료되면, 운영 체제는 해당 프로세스를 종료 상태로 설정하고 PID를 해제합니다.

###### Java Virtual Machine (JVM) 시작

사용자가 `java -jar` 명령을 실행하면 운영 체제의 쉘은 `java` 바이너리 프로그램을 실행합니다.
`java` 바이너리는 Java Virtual Machine의 실행 환경을 초기화하며, 이 바이너리는 네이티브 코드로 작성된 프로그램으로, JVM 실행의 시작점입니다.

운영 체제는 JVM을 새로운 프로세스로 실행합니다.
새로운 프로세스에는 PID(Process ID)가 부여되며, 운영 체제가 제공하는 기본적인 메모리 공간(코드, 데이터, 힙, 스택 영역)을 할당받습니다.

JVM은 바이트코드로 컴파일된 Java 애플리케이션을 실행하기 위한 런타임 환경을 제공합니다.
`-jar` 옵션은 실행할 JAR 파일을 명시하며, 해당 JAR 파일의 엔트리 포인트(Entry Point)를 자동으로 찾습니다.

운영 체제에서 JVM 프로세스에 할당된 메모리 공간은 다음과 같이 구성됩니다:

- Code Segment: 네이티브 JVM 실행 코드와 관련된 명령어가 저장됩니다.
- Data Segment: 정적 데이터와 초기화된 전역 변수가 저장됩니다.
- Heap: JVM의 Heap 영역에 사용됩니다.
- Stack: 각 스레드의 Java 스택이 포함됩니다.
- OS Kernel 영역: 운영 체제가 관리하는 메모리 영역으로, 파일 I/O 및 네트워크 작업에 사용됩니다.

###### JVM 내부 메모리 초기화

JVM은 실행될 프로그램에 필요한 메모리 공간을 운영 체제로부터 요청합니다.
JVM은 이 메모리 공간을 다음과 같이 주요 영역으로 나눕니다:

- Method Area (클래스 메타데이터 저장 공간)

  모든 스레드에서 공유되는 영역으로, JVM이 클래스 로더에 의해 로드한 클래스의 메타데이터를 저장하는 공간입니다.
  메서드 코드(바이트코드), 상수 풀(Constant Pool), 정적 변수, 클래스 레벨 데이터가 여기에 저장됩니다.

  Java 8부터는 HotSpot JVM에서 이 영역이 메타스페이스(Metaspace)로 대체되었습니다.
  Java 8 이전에는 `Permanent Generation`이라는 영역에서 클래스 메타데이터를 관리했으나, 이는 JVM 힙 공간의 일부로 관리되었습니다.
  Java 8부터는 `Metaspace`가 도입되었으며, 이는 네이티브 메모리를 활용합니다.
    - 동적으로 크기를 확장할 수 있어 메모리 부족 문제를 해결합니다.
    - `-XX:MetaspaceSize`: 초기 크기 설정.
    - `-XX:MaxMetaspaceSize`: 최대 크기 제한.

  저장되는 데이터:
    1. 클래스 정보:
        - 클래스 이름, 접근 제한자(`public`/`private`), 부모 클래스 정보, 구현된 인터페이스 목록.
        - 클래스와 관련된 메서드, 필드, 정적 변수의 구조 정보.

    2. 상수 풀(Constant Pool):
        - 클래스 파일의 심볼릭 참조와 리터럴 데이터(문자열 상수, 숫자 리터럴 등)가 저장됩니다.
        - 상수 풀은 런타임 중에 사용되며, JVM이 바이트코드를 실행하면서 참조를 실제 메모리 주소로 변환하는 데 활용됩니다.

    3. 정적 변수:
        - `static` 키워드로 선언된 클래스 변수는 Method Area에 저장되며, 애플리케이션 종료 시까지 유지됩니다.

        ```java
        class Example {
            static int sharedCounter = 0;
        }
        ```

    4. 메서드 코드:
        - 메서드 바이트코드가 포함됩니다.
        - 각 메서드의 시그니처, 반환 타입, 로컬 변수와 같은 실행에 필요한 정보도 여기에 포함됩니다.

- Heap (객체 저장 공간)

  Heap 메모리는 JVM에서 가장 중요한 영역 중 하나로, 런타임 동안 생성되는 모든 객체와 배열이 저장되는 공간입니다.
  이 영역은 Garbage Collector(GC)에 의해 관리되며, Java의 메모리 누수를 방지하고 JVM에 의해 크기가 동적으로 조정될 수 있습니다.

  JVM 실행 시 초기 Heap 크기(`-Xms`)와 최대 Heap 크기(`-Xmx`)가 설정됩니다.
  기본적으로 JVM은 시스템 메모리의 일부를 사용하며, 개발자가 옵션을 통해 조정할 수 있습니다.
  Heap 메모리는 필요에 따라 확장되며, 최대 크기(`-Xmx`)를 초과할 경우 `OutOfMemoryError`가 발생합니다.

  객체 생성 과정은 다음과 같습니다.
    - `new` 키워드로 객체가 생성되면 JVM은 Heap에서 메모리를 할당합니다.

        ```java
        // `obj` 참조는 스택에, 객체 데이터는 Heap에 저장됩니다.
        MyClass obj = new MyClass();

        // 런타임 중 크기가 가변적인 데이터 구조(예: `ArrayList`, `HashMap`)가 Heap에 저장됩니다.
        List<String> data = new ArrayList<>();
        data.add("item1");
        ```

    - 생성된 객체의 참조는 스택(Frame Data)에 저장됩니다.

  Generation 모델:
    - Young Generation:
        - 새로운 객체가 생성되고, 짧은 생명 주기의 객체를 관리합니다.
        - 대부분의 객체는 단명하며, Young Generation에서 생성되었다가 곧 GC에 의해 제거됩니다.

        1. Eden Space:
            - 객체가 처음 생성되는 영역.
            - 생성된 객체 대부분은 이곳에서 사라지며, 일정 기준을 충족한 객체만 Survivor Space로 이동합니다.
        2. Survivor Spaces (S0, S1):
            - Eden Space에서 살아남은 객체가 이동하는 영역.
            - Young Generation의 Copying GC에서 한 Survivor Space에서 다른 공간으로 객체를 이동합니다.

    - Old Generation:
        - Young Generation에서 오래 생존한 객체가 이동하며, 장수 객체를 관리합니다.
        - 예를 들어, 대규모 데이터 구조 또는 애플리케이션 전반에 걸쳐 유지되는 객체가 여기에 저장됩니다.
        - Garbage Collection:
            - Old Generation의 객체는 Full GC에서 관리됩니다.
            - Minor GC에 비해 더 많은 리소스를 소비합니다.

    - Metaspace:
        - 클래스 메타데이터가 저장되는 공간으로, Java 8부터 Heap 메모리에서 분리되어 네이티브 메모리를 사용합니다.

- Stack (스레드별 스택)

  Java Stack은 JVM에서 각 스레드마다 독립적으로 생성되는 메모리 공간으로, 메서드 호출과 반환, 지역 변수, 연산 중간 값 등을 관리합니다.
  메모리 관리는 LIFO(Last-In-First-Out) 방식으로 이루어지며, 메서드 호출 시 새로운 스택 프레임(Stack Frame)이 추가되고, 메서드가 반환되면 해당 프레임이 제거됩니다.

  JVM은 스레드가 생성될 때 고유한 스택을 할당합니다.
  스택 크기는 JVM 옵션(`-Xss`)으로 설정할 수 있습니다.

  다음과 같은 역할들을 수행합니다:
    - 메서드 호출 관리(Frame Data):
        - 메서드 호출과 관련된 모든 데이터를 저장합니다.
        - 각 스레드가 고유의 스택을 가지므로, 스레드 간 데이터 충돌이 없습니다.
        - 예외 처리를 위한 데이터도 여기에 포함됩니다.

    - 지역 변수 및 매개변수 저장:
        - 메서드의 지역 변수와 매개변수가 저장됩니다.

        ```java
        // `a`와 `b`는 Local Variables에 저장됩니다.
        int sum(int a, int b) {
            return a + b;
        }
        ```

    - 연산 스택(Operand Stack) 관리:
        - 바이트코드 실행 중 발생하는 연산 중간 값이 저장됩니다.
        - 예: `a + b` 연산에서 `a`와 `b`가 Operand Stack에 쌓이고, 덧셈 연산 후 결과가 다시 스택에 저장됩니다.

- Program Counter Register (PC Register)

  PC Register는 JVM 명령어 해석기(Interpreter)가 다음에 실행할 명령어를 추적하는 역할을 합니다.
  모든 스레드는 독립적인 PC Register를 가지며, 스레드 간 충돌을 방지합니다.

  스레드 생성 시 PC Register가 초기화됩니다.
  메서드 호출 시 해당 메서드의 첫 번째 명령어 주소가 PC Register에 저장됩니다.

  명령어 해석기가 PC Register를 읽어 현재 실행 중인 명령을 확인하고, 다음 명령을 실행합니다.
  네이티브 메서드 실행 중일 때는 PC Register 값이 정의되지 않을 수 있습니다.

  다음과 같은 역할들을 수행합니다.
    1. 명령어 위치 추적:
        - 현재 실행 중인 JVM 명령어(바이트코드)의 주소를 저장합니다.
    2. 스레드 안전성 제공:
        - 각 스레드가 독립적인 PC Register를 가지므로, 동시 실행 환경에서도 안전하게 명령어를 추적할 수 있습니다.

  JVM의 명령어 해석기(Interpreter)가 다음에 실행할 명령을 결정하는 데 사용됩니다.

- Native Method Stack (네이티브 메서드 호출용 스택)

  Native Method Stack은 Java 애플리케이션이 네이티브 코드(C/C++)를 실행할 때 사용하는 메모리 영역입니다.
  JVM은 Java Native Interface(JNI)를 통해 외부 라이브러리와 상호작용하며, 이 과정에서 네이티브 메서드 호출 정보를 관리합니다.

  네이티브 메서드 호출 시 Native Method Stack이 초기화됩니다.
  호출이 완료되면 스택은 정리됩니다.

  예를 들어, `System.loadLibrary()`로 외부 네이티브 라이브러리를 로드하거나 파일 I/O, 네트워크 소켓 접근 등의 경우에 사용됩니다.

  다음과 같은 역할들을 수행합니다.
    1. 네이티브 라이브러리 호출:
        - 네이티브 코드에서 정의된 함수 호출을 처리합니다.
    2. 네이티브 데이터 관리:
        - 네이티브 함수 호출과 관련된 데이터(매개변수, 반환 값 등)를 저장합니다.

  네이티브 메서드를 호출하기 위한 스택으로, 네이티브 라이브러리 함수와의 상호작용을 처리합니다.

JVM의 메모리 할당 및 관리는 다음과 같이 이뤄집니다.

- 클래스 로드 단계
    1. Loading: 클래스 로더가 `.class` 파일을 읽어 JVM에 로드.
        - 클래스 로더(ClassLoader)가 `.class` 파일을 읽고 메타데이터를 Method Area에 적재합니다.
        - 이는 애플리케이션이 처음으로 해당 클래스를 참조할 때 발생합니다.

    2. Linking:
        - Verification: JVM은 바이트코드의 유효성을 확인하여, 잘못된 바이트코드로 인해 런타임 에러가 발생하는 것을 방지합니다.

        - Preparation:
            - 클래스의 정적 변수와 기본값이 설정됩니다.
            - 예: `static int x;`는 0으로 초기화됩니다.

        - Resolution:
            - 심볼릭 참조를 실제 메모리 주소로 변환합니다.
            - 클래스 내부의 메서드 참조, 필드 참조 등이 이 단계에서 해결됩니다.

    3. Initialization: `static` 초기화 블록과 정적 변수가 실행됩니다.

        ```java
        static int x = 10;
        static {
            x = 20;
        }
        ```

- 객체 생성과 Heap 관리
    - `new` 키워드를 사용하면 JVM은 Heap에 객체를 생성하고, 해당 객체의 참조를 반환합니다.
    - GC가 주기적으로 Heap을 스캔하여 더 이상 참조되지 않는 객체를 회수합니다.

- 스레드 스택 관리
    - 각 스레드는 고유의 스택을 가지고 있으며, 메서드 호출 시마다 스택 프레임을 추가합니다.
    - 메서드 종료 시 해당 스택 프레임은 제거됩니다.

- Garbage Collection

  Garbage Collection은 더 이상 참조되지 않는 객체를 자동으로 회수하여 메모리 누수를 방지합니다.

  GC 알고리즘은 다음과 같습니다:
    - Mark-and-Sweep:
        - 사용 중인 객체를 표시하고, 사용되지 않는 객체를 삭제.
        - Mark 단계: 모든 객체를 스캔하여 참조 중인 객체를 마킹.
        - Sweep 단계: 참조되지 않는 객체를 제거하고, 빈 메모리를 회수.
    - Copying:
        - Young Generation에서 사용.
        - 사용 중인 객체를 복사하여 새로운 공간에 할당.
        - 살아남은 객체를 Eden Space에서 Survivor Space로 복사하고, Eden Space를 비웁니다.
    - Generational GC:
        - Heap을 Young Generation과 Old Generation을 나눠 효율적으로 관리.
        - Young Generation에서 Minor GC, Old Generation에서 Major GC 또는 Full GC를 수행.

  Young과 Old 영역별로 서로 다른 GC가 수행됩니다.
    1. Minor GC:
        - Young Generation에서 발생하며, Eden Space와 Survivor Space의 객체를 정리합니다.
        - 대부분의 객체는 Eden Space에서 제거되며, 살아남은 객체는 Survivor Space로 이동합니다.

    2. Major GC (Full GC):
        - Old Generation과 Young Generation을 모두 정리합니다.
        - 더 많은 시간이 소요되며, 애플리케이션 성능에 영향을 미칠 수 있습니다.

  JVM은 다양한 GC 알고리즘을 제공하며, 애플리케이션의 요구에 따라 선택할 수 있습니다:

    - Serial GC:
        - 단일 스레드 환경에 적합.
        - `-XX:+UseSerialGC` 옵션으로 활성화.
    - Parallel GC:
        - 다중 스레드 환경에 적합.
        - `-XX:+UseParallelGC` 옵션으로 활성화.
    - G1 GC (Garbage First):
        - 대규모 애플리케이션에 적합하며, Full GC 발생 빈도를 줄임.
        - `-XX:+UseG1GC` 옵션으로 활성화.

###### JAR 파일 로드

JVM은 지정된 JAR 파일을 로드합니다.
이 과정에서 JAR 파일 내부의 `META-INF/MANIFEST.MF` 파일을 읽습니다.
이 파일에는 `Main-Class`라는 속성이 있는데, 이는 JVM이 실행해야 할 클래스의 FQCN(Fully Qualified Class Name)을 지정합니다.

```plaintext
Main-Class: com.example.SpringBootApp
```

메인 클래스를 확인하고, JVM 내부에서 `main()` 메서드의 시작점을 결정합니다.

Spring Boot는 이 속성을 사용하여 메인 애플리케이션 클래스를 지정합니다.

###### 클래스 로더에 의해 클래스 로드

JVM은 클래스 로더(Class Loader)를 사용해 애플리케이션이 실행되는 데 필요한 클래스와 의존성을 로드합니다.

Spring Boot JAR 파일은 Fat JAR로, 모든 종속 라이브러리가 포함되어 있습니다.
이를 통해 실행 시 추가적인 라이브러리 설치 없이 독립적으로 실행 가능합니다.

Spring Boot의 `org.springframework.boot.loader.JarLauncher`가 이를 처리하며, Fat JAR의 구조를 분석하고 적절히 클래스를 로드합니다.

###### `main()` 메서드 실행

Manifest 파일에 명시된 메인 클래스의 `public static void main(String[] args)` 메서드가 호출됩니다.

Spring Boot 애플리케이션의 경우, 이 메서드는 일반적으로 `SpringApplication.run()`을 호출하여 애플리케이션 컨텍스트를 초기화합니다.

```java
public class SpringBootApp {
    public static void main(String[] args) {
        SpringApplication.run(SpringBootApp.class, args);
    }
}
```

###### Spring Boot의 초기화 단계

Spring Boot 애플리케이션의 초기화 단계는 다음과 같습니다.

###### (a) 스프링 애플리케이션 컨텍스트 초기화

- SpringApplication이 실행되면서 애플리케이션 컨텍스트(Application Context)가 생성됩니다.
- 설정된 `@Configuration` 클래스와 `@ComponentScan`을 통해 애플리케이션의 구성 요소가 로드됩니다.
- Bean Factory가 필요한 Bean들을 생성 및 초기화합니다.

###### (b) 내장 웹 서버(Tomcat, Jetty 등) 시작

- Spring Boot는 내장 웹 서버를 포함하고 있으므로, 애플리케이션 컨텍스트를 초기화한 후 내장 서버를 실행합니다.
- 기본적으로 Spring Boot는 Tomcat 서버를 사용하며, 포트 8080에서 수신 대기합니다(설정에 따라 다를 수 있음).

###### (c) DispatcherServlet 등록

- Spring MVC 애플리케이션인 경우, `DispatcherServlet`이 서블릿 컨테이너에 등록됩니다.
- 요청 매핑, 핸들러 설정 등이 완료됩니다.

###### (d) 애플리케이션 로직 실행 준비 완료

- 애플리케이션의 주요 로직(예: REST API, 서비스)이 요청을 처리할 준비가 됩니다.

###### 6. 운영 상태로 전환

- Spring Boot 애플리케이션이 정상적으로 실행되었다면, 콘솔에 다음과 비슷한 로그가 표시됩니다.

     ```plaintext
     Started SpringBootApp in 3.456 seconds (JVM running for 3.890)
     ```

- 내장 서버가 시작된 상태에서 HTTP 요청을 수신하거나, 비동기 작업을 처리할 준비가 완료됩니다.

###### 추가 사항

1. 환경 변수 및 프로파일
    - 실행 시 `application.properties` 또는 `application.yml` 파일을 로드하며, 환경 변수와 프로파일 설정을 반영합니다.
    - `-Dspring.profiles.active=dev`와 같은 옵션을 사용해 프로파일을 지정할 수도 있습니다.

2. 종료 처리
    - 애플리케이션이 종료될 때 JVM은 모든 리소스를 해제하며, Spring Boot는 등록된 `@PreDestroy` 메서드를 호출하여 필요한 종료 작업을 수행합니다.

<!-- /curriculum-chunk -->

### 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교

#### 원문: 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교

<!-- curriculum-chunk: sha256=159a3dc8e8415e1a2e039d76c14ccb03d865551b9d8de132a4ba7bf5ea29b736 major=spring-backend-frameworks mid=Spring Boot 실행 모델 sub=과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교 sources=source/interview_questions.md:4583-4753, source/interviews.md:4583-4753 -->

> Source: `source/interview_questions.md:4583-4753`
> Classification reason: spring boot
> Duplicate source aliases: `source/interview_questions.md:4583-4753, source/interviews.md:4583-4753`

##### 과거 톰캣 & 스프링 실행 방식 vs Spring boot 비교

Spring 애플리케이션을 톰캣(Tomcat)을 사용하여 실행하는 방식은 Spring Boot 등장 이전과 Spring Boot 등장 이후로 크게 나눌 수 있습니다.
이 두 방식의 차이점은 톰캣 서버 관리와 애플리케이션의 배포 방식에서 발생합니다. 각각의 방식을 구체적으로 설명드리겠습니다.

1. 기존 방식: 별도로 톰캣을 실행하고 web.xml을 사용하는 경우

   Spring Boot 이전에는 톰캣과 같은 서블릿 컨테이너를 애플리케이션 서버로 사용하여 Spring 애플리케이션을 실행했습니다.
   이 방식에서는 *톰캣을 별도로 설치하고, war(WAR) 파일 형태로 애플리케이션을 배포*하는 방식이 일반적이었습니다.

    1. 톰캣 서버 설치:
        - 톰캣은 Apache Tomcat 사이트에서 별도로 다운로드하여 설치해야 했습니다. 톰캣은 서블릿 컨테이너로서 웹 애플리케이션의 실행 환경을 제공합니다.
        - 톰캣 서버의 `conf/server.xml` 파일에서 포트(예: 8080)와 호스트 설정을 정의합니다.

            ```xml
            <!-- conf/server.xml -->
            <Connector port="8080" protocol="HTTP/1.1"
                    connectionTimeout="20000"
                    redirectPort="8443" />
            ```

        - webapps 폴더에 배포된 애플리케이션을 자동으로 탐지하여 실행하는 역할을 수행합니다.

    2. Spring 애플리케이션 구성:

       Spring 애플리케이션의 서블릿 설정과 매핑 정보는 일반적으로 `web.xml` 파일을 사용하여 디스패처 서블릿(DispatcherServlet)과 같은 서블릿들을 설정합니다.
       이 파일은 `WEB-INF/web.xml` 경로에 위치하며, 서블릿 컨테이너가 애플리케이션을 시작할 때 이 파일을 읽어 서블릿을 초기화합니다.

       `web.xml` 파일에는 서블릿 매핑, 필터, 리스너 등의 정보가 포함되어 있습니다.
       주로 `DispatcherServlet`과 같은 Spring MVC의 진입점을 설정합니다.

        ```xml
        <!-- WEB-INF/web.xml -->
        <web-app>
            <!-- DispatcherServlet 설정 -->
            <servlet>
                <servlet-name>dispatcher</servlet-name>
                <servlet-class>org.springframework.web.servlet.DispatcherServlet</servlet-class>
                <load-on-startup>1</load-on-startup>
            </servlet>

            <!-- DispatcherServlet URL 매핑 -->
            <servlet-mapping>
                <servlet-name>dispatcher</servlet-name>
                <url-pattern>/</url-pattern>
            </servlet-mapping>

            <!-- 컨텍스트 리스너 설정 -->
            <listener>
                <listener-class>org.springframework.web.context.ContextLoaderListener</listener-class>
            </listener>

            <!-- Spring Context 파일 위치 -->
            <context-param>
                <param-name>contextConfigLocation</param-name>
                <param-value>/WEB-INF/spring-servlet.xml</param-value>
            </context-param>
        </web-app>
        ```

    3. Spring 설정 파일 (spring-servlet.xml):

       web.xml에서 참조하는 Spring 설정 파일은 주로 WEB-INF 아래에 위치하며, 애플리케이션의 Spring 빈(bean) 설정, 데이터베이스 연결, 서비스 레이어 등의 설정을 포함합니다.

        ```xml
        <!-- WEB-INF/spring-servlet.xml -->
        <beans xmlns="http://www.springframework.org/schema/beans"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://www.springframework.org/schema/beans
                http://www.springframework.org/schema/beans/spring-beans.xsd">

            <!-- View Resolver -->
            <bean class="org.springframework.web.servlet.view.InternalResourceViewResolver">
                <property name="prefix" value="/WEB-INF/views/" />
                <property name="suffix" value=".jsp" />
            </bean>

            <!-- Controller, Service, Repository 빈 설정 등 -->
        </beans>
        ```

    4. WAR 파일로 배포:
        - Spring 애플리케이션은 WAR(WAR: Web Application Archive) 파일로 패키징됩니다.
        - WAR 파일은 `WEB-INF/web.xml`, 애플리케이션의 소스 코드와 리소스 파일, 설정 파일, 정적 자원, 서블릿 설정 등을 포함하는 아카이브 파일입니다.

    5. 톰캣에 WAR 배포:
        - 톰캣의 `webapps` 디렉토리에 WAR 파일을 배포하거나, 톰캣 관리 콘솔을 사용하여 WAR 파일을 업로드합니다.
        - 톰캣은 배포된 WAR 파일을 자동으로 해제(압축을 풀고)하여 해당 애플리케이션을 실행합니다.
        - 톰캣은 해당 애플리케이션을 서블릿 컨테이너에서 실행하며, `web.xml`을 통해 애플리케이션의 엔트리 포인트를 설정하고, 서블릿과 필터를 초기화합니다.

    6. 애플리케이션 실행:
        - 톰캣 서버를 수동으로 시작해야 합니다. 이를 위해 톰캣의 bin/startup.sh (Linux/Unix) 또는 bin/startup.bat (Windows) 스크립트를 실행합니다.
        - 톰캣 서버가 시작되면 지정된 포트(기본적으로 8080)에서 리스닝을 시작하고, 클라이언트 요청을 처리합니다.
        - 클라이언트는 브라우저에서 `http://localhost:8080/애플리케이션이름`으로 접근하여 애플리케이션에 접속할 수 있습니다.

   단점:
    - 톰캣 관리의 복잡성: 톰캣 서버는 별도로 관리해야 하며, 여러 애플리케이션이 같은 톰캣 인스턴스에서 동작할 때 충돌이나 자원 관리 문제가 발생할 수 있습니다.
    - WAR 배포의 번거로움: WAR 파일로 패키징하고, 이를 톰캣에 배포하는 과정이 번거롭습니다.
    - 설정의 복잡성: `web.xml`을 통한 서블릿 설정, 매핑, 필터, 리스너 관리가 복잡할 수 있습니다.

2. Spring Boot 등장 후: 내장 톰캣을 사용하는 방식

   Spring Boot는 내장 톰캣(Embedded Tomcat)을 사용하여 자동으로 웹 애플리케이션을 실행합니다.
   Spring Boot의 가장 큰 장점은 개발자가 별도로 서블릿 컨테이너를 설치하거나 관리할 필요 없이, 애플리케이션이 자체적으로 톰캣 서버를 포함하여 실행된다는 점입니다.
   이 방식은 jar(JAR) 파일로 패키징되어 실행됩니다.

    1. Spring Boot 프로젝트 설정:
       Spring Boot에서는 `web.xml`을 사용하지 않고, Java Config를 사용합니다.
       `@SpringBootApplication` 어노테이션을 사용하여 애플리케이션 진입점을 정의하고, 모든 설정을 자바 코드로 구성할 수 있습니다.

        ```java
        @SpringBootApplication
        public class Application {
            public static void main(String[] args) {
                SpringApplication.run(Application.class, args);
            }
        }
        ```

    2. 내장 톰캣 포함:
       Spring Boot는 내장 톰캣을 자동으로 포함합니다.
       Maven이나 Gradle을 사용하여 프로젝트를 빌드할 때, Spring Boot는 톰캣을 의존성에 포함시켜 자동으로 애플리케이션에 포함합니다.

        ```xml
        <!-- Maven 예시 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        ```

    3. JAR 파일로 패키징:
        - Spring Boot 애플리케이션은 JAR(Java Archive) 파일로 패키징됩니다. JAR 파일에는 애플리케이션 코드와 함께 내장 톰캣이 포함되어 있습니다. 이 JAR 파일은 스탠드얼론 애플리케이션처럼 실행 가능합니다.
        - `mvn package` 또는 `gradle build` 명령어를 사용하여 JAR 파일로 패키징할 수 있습니다.

    4. 애플리케이션 실행:
        - JAR 파일로 패키징된 애플리케이션은 다음과 같이 명령어 한 줄로 실행할 수 있습니다:

            ```bash
            java -jar myapp.jar
            ```

        - 이 명령어를 실행하면 내장된 톰캣이 시작되며, 애플리케이션이 자동으로 포트 8080에서 리스닝을 시작합니다.

    5. 자동 설정 및 실행:
        - Spring Boot는 자동 설정(Autoconfiguration) 기능을 사용하여 톰캣을 포함한 모든 웹 관련 설정을 자동으로 처리합니다. 개발자는 복잡한 설정 없이 즉시 애플리케이션을 실행할 수 있습니다.
        - 추가적인 톰캣 설정이 필요할 경우, `application.properties`나 `application.yml` 파일을 통해 쉽게 설정을 변경할 수 있습니다.

            ```properties
            server.port=9090  # 포트 변경 예시
            ```

    6. 애플리케이션 접속:
        - 애플리케이션이 실행되면 클라이언트는 `http://localhost:8080`에서 애플리케이션에 접근할 수 있습니다.

   장점:
    - 내장 톰캣: 별도로 톰캣을 설치하거나 관리할 필요 없이 애플리케이션 내부에 톰캣이 포함됩니다.
    - JAR 파일 배포: WAR 파일 대신 JAR 파일로 애플리케이션을 패키징하고 실행할 수 있어 배포가 간편합니다.
    - 자동 설정: Spring Boot는 웹 서버, 서블릿 설정 등을 자동으로 처리하므로, 설정 복잡성이 크게 줄어듭니다.
    - 독립 실행 가능: Spring Boot 애플리케이션은 JAR 파일만으로 독립적으로 실행될 수 있어, 다른 애플리케이션 서버에 의존하지 않습니다.

| 기능          | 기존 방식 (WAR + 외부 톰캣)   | Spring Boot 방식 (내장 톰캣 + JAR)     |
|-------------|-----------------------|----------------------------------|
| 서버 설치       | 톰캣을 별도로 설치해야 함        | 내장 톰캣을 포함하여 별도 설치 불필요            |
| 배포 형태       | WAR 파일을 배포            | JAR 파일을 실행                       |
| 애플리케이션 설정   | `web.xml` 및 서블릿 설정 필요 | `@SpringBootApplication` 사용      |
| 애플리케이션 실행   | 톰캣에 WAR 파일 배포 후 서버 시작 | `java -jar` 명령어로 실행              |
| 확장성 및 설정 관리 | 서버와 애플리케이션 설정을 따로 관리  | Spring Boot에서 자동 설정 및 관리         |
| 포트 설정       | `server.xml`에서 포트 설정  | `application.properties`에서 간편 설정 |
| 애플리케이션 독립성  | 톰캣 서버에 의존             | 애플리케이션이 독립적으로 실행 가능              |

<!-- /curriculum-chunk -->

## Spring HTTP 클라이언트와 Reactive

### Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계

#### 원문: Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계

<!-- curriculum-chunk: sha256=b71bcb619097feb3a39e67ed99738279adc94714bada10e54d15bd91827abf49 major=spring-backend-frameworks mid=Spring HTTP 클라이언트와 Reactive sub=Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계 sources=source/interview_questions.md:10732-10854, source/interviews.md:10680-10802 -->

> Source: `source/interview_questions.md:10732-10854`
> Classification reason: spring reactive/http client
> Duplicate source aliases: `source/interview_questions.md:10732-10854, source/interviews.md:10680-10802`

##### Mono와 Flux: 개념, 필요성, 이벤트 루프와의 관계

Spring WebFlux에서 비동기(Asynchronous) 및 논블로킹(Non-blocking) 데이터 스트림을 처리하는 두 가지 주요 타입이 `Mono<T>`와 `Flux<T>` 입니다.

- Mono와 Flux 개념

  Spring WebFlux는 Reactor 기반의 리액티브 프로그래밍(reactive programming) 모델을 따릅니다.
  이 모델에서는 데이터를 스트림(Stream) 형태로 비동기 처리하는데, 이를 표현하는 두 가지 핵심 타입이 `Mono<T>`와 `Flux<T>`입니다.

  | 타입   | 설명 |
        |------------|----------|
  | `Mono<T>` | 0개 또는 1개의 데이터를 비동기적으로 반환하는 타입 |
  | `Flux<T>` | 0개 이상의 데이터 스트림을 비동기적으로 반환하는 타입 |

  즉, "단일 값이 반환될지, 다중 값이 반환될지"에 따라 Mono와 Flux가 제공됨.
  Mono와 Flux를 나누는 이유는 데이터 흐름을 효율적으로 표현하기 위해서입니다.

    - Mono 예제 (단일 데이터 반환)

        ```java
        Mono<String> mono = Mono.just("Hello Mono");

        mono.subscribe(value -> System.out.println("Received: " + value));
        // Received: Hello Mono
        ```

      `"Hello Mono"` 하나의 값만 반환하는 단일 비동기 작업

    - Flux 예제 (다중 데이터 스트림)

        ```java
        Flux<String> flux = Flux.just("Apple", "Banana", "Cherry");

        flux.subscribe(value -> System.out.println("Received: " + value));
        // Received: Apple
        // Received: Banana
        // Received: Cherry
        ```

      `"Apple"`, `"Banana"`, `"Cherry"` 여러 개의 데이터를 스트리밍 방식으로 전달

- 비동기 데이터 흐름의 특성 차이
    - 일반적인 REST API 요청은 단일 데이터 반환이므로 `Mono<T>`가 적합.
    - WebSocket, SSE(Server-Sent Events), gRPC 스트리밍과 같은 연속적인 데이터 스트림은 `Flux<T>`가 적합.

- 기존 Future, CompletableFuture와의 차이

  기존의 `Future<T>` 또는 `CompletableFuture<T>`는 단일 비동기 결과만 반환 가능하지만,
  리액티브 스트림(reactive streams) 은 여러 개의 데이터를 이벤트 기반으로 전달할 수 있어야 하므로 `Flux<T>`가 필요.

  | 비교 항목  | `Future<T>` | `CompletableFuture<T>` | `Mono<T>` | `Flux<T>` |
        |---------------|--------------|-----------------|----------|----------|
  | 반환 개수 | 1개 | 1개 | 0 또는 1개 | 0개 이상 |
  | 비동기 지원 | O | O | O | O |
  | 데이터 스트리밍 | X | X | X | O |

  ➡ 즉, 단일 데이터라면 Mono, 다중 스트림이라면 Flux가 필요.

- 이벤트 루프(Event Loop)와의 관계

  Spring WebFlux에서 `Mono`와 `Flux`는 내부적으로 Reactor Netty의 이벤트 루프(Event Loop) 모델을 사용합니다.

    ```java
    WebClient webClient = WebClient.create();

    Mono<String> response = webClient.get()
            .uri("https://example.com/api")
            .retrieve()
            .bodyToMono(String.class);

    response.subscribe(result -> System.out.println("Response: " + result));
    ```

  1️⃣ `WebClient.get()` 요청이 생성됨
  2️⃣ Reactor Netty의 이벤트 루프(Event Loop)에 비동기 작업이 등록됨
  3️⃣ 요청이 완료되면 스레드 블로킹 없이 이벤트 루프에서 결과를 전달
  4️⃣ `Mono.subscribe()`를 통해 데이터가 비동기적으로 처리됨

  ➡ 이벤트 루프는 Non-blocking I/O 방식으로 요청을 관리하며, 기존의 동기 방식과 달리 스레드가 대기하지 않음.

    ```java
    Flux<Long> flux = Flux.interval(Duration.ofSeconds(1))
                        .map(i -> i + 1)
                        .take(5);

    flux.subscribe(System.out::println);
    // 1
    // 2
    // 3
    // 4
    // 5
    ```

    - `Flux.interval()`이 이벤트 루프를 통해 1초마다 새로운 값을 생성.
    - 하나의 스레드에서 연속적으로 데이터를 생성 및 전달.

  ➡ 즉, `Mono`와 `Flux`는 이벤트 루프를 활용하여 논블로킹 방식으로 데이터를 스트리밍함.

- 실무에서 Mono와 Flux를 활용하는 사례

    - REST API에서 `Mono` 사용

        ```java
        @GetMapping("/user/{id}")
        public Mono<User> getUser(@PathVariable String id) {
            return userService.getUserById(id);
        }
        ```

        - 단일 사용자 정보를 반환해야 하므로 `Mono<User>` 사용.

    - SSE(Server-Sent Events)에서 `Flux` 사용

        ```java
        @GetMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
        public Flux<String> streamEvents() {
            return Flux.interval(Duration.ofSeconds(1))
                    .map(i -> "Event #" + i);
        }
        ```

        - 연속적인 데이터 스트림을 반환해야 하므로 `Flux<String>` 사용.

<!-- /curriculum-chunk -->

### RestTemplate vs WebClient 차이

#### 원문: RestTemplate vs WebClient 차이

<!-- curriculum-chunk: sha256=754fcd2098062c3d4bb3e4b16e0d1dc56f10cb8f88663205a16be94cd3c6bc18 major=spring-backend-frameworks mid=Spring HTTP 클라이언트와 Reactive sub=RestTemplate vs WebClient 차이 sources=source/interview_questions.md:10596-10696, source/interviews.md:10544-10644 -->

> Source: `source/interview_questions.md:10596-10696`
> Classification reason: spring reactive/http client
> Duplicate source aliases: `source/interview_questions.md:10596-10696, source/interviews.md:10544-10644`

##### RestTemplate vs WebClient 차이

Spring에서 HTTP 클라이언트를 사용할 때 가장 많이 비교되는 두 가지 옵션이 `RestTemplate` 과 `WebClient` 입니다.

- RestTemplate의 배경

    - `RestTemplate`은 Spring 3에서 등장하여 오래도록 사용된 HTTP 동기(Blocking) 클라이언트입니다.
    - 기반 기술: Java의 `HttpURLConnection`, Apache HttpClient 등을 내부적으로 활용
    - 동기 방식으로 요청을 수행하며, 한 개의 HTTP 요청을 보낼 때 해당 스레드는 응답을 받을 때까지 블로킹됨.

    ```java
    RestTemplate restTemplate = new RestTemplate();
    ResponseEntity<String> response = restTemplate.getForEntity("https://example.com/api", String.class);
    System.out.println(response.getBody()); // 동기적으로 응답을 기다림
    ```

- WebClient의 등장 배경

  Spring 5부터 Spring WebFlux 모듈이 추가되면서 `WebClient`가 등장했습니다.
    - 기존 `RestTemplate`이 블로킹(Blocking) 방식이었던 것과 달리, `WebClient`는 완전한 논블로킹(Non-blocking) 방식을 채택.
    - 기반 기술: Reactor 기반, Netty와 같은 논블로킹 I/O 라이브러리를 사용
    - 이벤트 루프(Event Loop) 기반의 요청 처리로 높은 동시성을 지원

    ```java
    WebClient webClient = WebClient.create();

    Mono<String> response = webClient.get()
            .uri("https://example.com/api")
            .retrieve()
            .bodyToMono(String.class);

    response.subscribe(System.out::println); // 비동기적으로 응답 처리
    ```

- 내부 동작 원리 비교: 스레드 모델 및 요청 처리 방식

  두 라이브러리는 HTTP 요청을 처리하는 방식이 완전히 다릅니다.

    - RestTemplate: 동기 블로킹 모델
        - `RestTemplate`은 각 요청이 스레드를 점유하며 응답을 받을 때까지 블로킹됨.
        - Thread-per-request 모델을 사용하여, 요청이 많아질 경우 스레드 풀이 고갈될 위험이 있음.

      동작 과정:
        1. 호출 스레드가 HTTP 요청을 생성
        2. HTTP 요청을 보내고, 해당 스레드는 응답을 받을 때까지 대기(Blocking)
        3. 응답을 받으면, 데이터를 리턴하고 요청 종료

      Spring 내부에서 `RestTemplate`은 `SimpleClientHttpRequestFactory` 또는 `HttpComponentsClientHttpRequestFactory`를 사용하여 HTTP 요청을 보냅니다.

        ```java
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(5000);
        factory.setReadTimeout(5000);

        RestTemplate restTemplate = new RestTemplate(factory);
        ResponseEntity<String> response = restTemplate.getForEntity("https://example.com/api", String.class);
        ```

        - 내부적으로 Java 기본 `HttpURLConnection`을 활용하여 요청을 수행.
        - `HttpURLConnection`은 블로킹 방식이므로 응답이 도착할 때까지 스레드는 대기 상태.

    - WebClient: 비동기 논블로킹 모델
        - `WebClient`는 Spring WebFlux의 리액티브 스트림(Reactive Streams) 기반으로 동작.
        - Thread-per-request 모델이 아닌, 이벤트 루프(Event Loop) 기반으로 요청을 처리.
        - Netty의 Reactor 패턴을 활용하여 I/O 작업을 비동기적으로 실행.

      WebClient 내부 코드 흐름은 다음과 같습니다:
        1. HTTP 요청을 비동기적으로 생성 (`subscribe()`)
        2. Netty Event Loop에 요청을 등록
        3. I/O 작업이 완료되면 콜백 방식으로 응답을 처리
        4. 응답 데이터를 리액티브 스트림을 통해 전달 (`Mono<T>` 또는 `Flux<T>`)

        ```java
        WebClient webClient = WebClient.create();
        Mono<String> response = webClient.get()
                .uri("https://example.com/api")
                .retrieve()
                .bodyToMono(String.class);

        response.subscribe(result -> System.out.println("Response: " + result)); // 비동기적 실행
        ```

        - Netty 이벤트 루프 기반이므로, 요청을 보낸 후에도 현재 스레드는 블로킹되지 않음.
        - 응답이 도착하면 Reactor의 Scheduler를 통해 백그라운드 스레드에서 처리.

Spring 공식 문서에서는 RestTemplate는 더 이상 새로운 기능이 추가되지 않을 예정이며, `WebClient`를 대체 기술로 권장하고 있습니다.
동기 방식의 요청이 필요한 경우 `WebClient`도 `.block()`을 사용하여 동기적으로 호출할 수 있기 때문에, 동기 방식의 요청이 필요하다고 RestTemplate을 사용해야 하는 것은 아닙니다.

```java
WebClient webClient = WebClient.create();
String response = webClient.get()
    .uri("https://example.com/api")
    .retrieve()
    .bodyToMono(String.class)
    .block();  // 동기 처리
```

🔹 RestTemplate는 Deprecated 예정인가?

- Spring 5부터 `RestTemplate`이 deprecate되지는 않았지만, 더 이상 발전하지 않음.
- `WebClient`를 적극적으로 사용해야 하며, 기존 `RestTemplate`을 `WebClient`로 마이그레이션하는 것이 권장됨.

<!-- /curriculum-chunk -->

### Spring WebFlux의 이벤트 루프(Event Loop)

#### 원문: Spring WebFlux의 이벤트 루프(Event Loop)

<!-- curriculum-chunk: sha256=b580f94b8abedad94fd01a0b8b05ee919c0cf7a4f2f5c3dd14b6a72adfca972b major=spring-backend-frameworks mid=Spring HTTP 클라이언트와 Reactive sub=Spring WebFlux의 이벤트 루프(Event Loop) sources=source/interview_questions.md:10697-10731, source/interviews.md:10645-10679 -->

> Source: `source/interview_questions.md:10697-10731`
> Classification reason: spring reactive/http client
> Duplicate source aliases: `source/interview_questions.md:10697-10731, source/interviews.md:10645-10679`

##### Spring WebFlux의 이벤트 루프(Event Loop)

Spring WebFlux가 사용하는 이벤트 루프는 Reactor Netty 또는 Jetty, Undertow 등의 서버 구현체에서 제공하는 논블로킹 이벤트 루프입니다.

참고로 Spring WebFlux의 이벤트 루프(Event Loop) 는 libuv와 직접적인 관련이 없습니다.
libuv는 Node.js에서 사용되는 이벤트 루프 라이브러리로,

- 주로 JavaScript의 비동기 I/O 이벤트 처리를 담당하며,
- C 기반으로 구현되어 있고, epoll (Linux), kqueue (macOS, BSD), IOCP (Windows) 등의 OS별 I/O 멀티플렉싱 기술을 활용합니다.

WebFlux의 이벤트 루프와 libuv는 같은 개념(이벤트 루프 모델)을 공유하지만, 내부적으로 동작 방식과 구현체가 다릅니다.

- WebClient와 Spring WebFlux의 이벤트 루프는 어떻게 동작하는가?

  Spring WebFlux는 Reactor Netty를 기본 서버로 사용하며, 내부적으로 Netty의 이벤트 루프(Event Loop) 를 활용합니다.
  Netty의 이벤트 루프는 Java NIO 기반으로 동작하며, libuv와 유사한 논블로킹 I/O 모델을 구현하고 있습니다.

  🔹 WebClient 요청 처리 과정에서 이벤트 루프의 역할
    1. `WebClient`가 HTTP 요청을 생성하고, Reactor Netty의 `EventLoopGroup`에 등록
    2. `EventLoopGroup`이 논블로킹 방식으로 요청을 관리 (`Selector`를 사용하여 소켓 이벤트 감지)
    3. 네트워크 I/O 작업이 완료되면, 스레드를 블로킹하지 않고 이벤트 핸들러에서 응답을 처리
    4. `Mono` 또는 `Flux` 형태로 데이터를 스트리밍하여 반환

    ```java
    WebClient webClient = WebClient.create();

    Mono<String> response = webClient.get()
            .uri("https://example.com/api")
            .retrieve()
            .bodyToMono(String.class);

    response.subscribe(System.out::println);
    ```

  ➡ 현재 스레드는 블로킹되지 않으며, Reactor Netty 이벤트 루프가 비동기적으로 I/O를 처리.

<!-- /curriculum-chunk -->

## 트랜잭션과 데이터 접근

### Spring JDBC에서도 데이터를 가져오는 방식

#### 원문: Spring JDBC에서도 데이터를 가져오는 방식

<!-- curriculum-chunk: sha256=0d87faef55b3fa8c3d2fba53599801d5bbff5bce866f7d93e50b1a2334b5af99 major=spring-backend-frameworks mid=트랜잭션과 데이터 접근 sub=Spring JDBC에서도 데이터를 가져오는 방식 sources=source/interview_questions.md:9330-9422, source/interviews.md:9278-9370 -->

> Source: `source/interview_questions.md:9330-9422`
> Classification reason: spring data access
> Duplicate source aliases: `source/interview_questions.md:9330-9422, source/interviews.md:9278-9370`

##### Spring JDBC에서도 데이터를 가져오는 방식

Spring JDBC에서도 데이터를 가져오는 방식은 PDO의 `fetch()`와 매우 유사합니다.
즉, Spring JDBC는 결과셋을 한 번에 모두 메모리로 로드하지 않고, 필요할 때마다 한 행씩 가져와 처리하는 방식으로 동작할 수 있습니다.
이를 통해 메모리 사용량을 줄이고, 대규모 데이터를 효율적으로 처리할 수 있습니다.

Spring JDBC에서 데이터베이스에서 데이터를 가져오는 방식은 JDBC 표준에 따라 이루어지며, 크게 두 가지 방식으로 나눌 수 있습니다:

1. `RowMapper` 인터페이스를 사용하는 방식: 한 번에 한 행씩 매핑하여 처리합니다.
2. `ResultSetExtractor` 인터페이스를 사용하는 방식: 전체 결과셋을 처리할 수 있습니다.

Spring JDBC의 데이터 페칭 방식은 다음과 같습니다.

1. `JdbcTemplate`을 이용한 `RowMapper`

   Spring JDBC의 `JdbcTemplate`을 사용하면, `RowMapper`를 통해 *결과셋을 한 번에 하나씩 가져와 처리*할 수 있습니다.
   이 방식은 스트리밍 방식으로 대량의 데이터를 처리하는 데 적합합니다.

    - `RowMapper`는 데이터베이스의 각 행을 애플리케이션에서 사용할 수 있는 객체로 변환하는 역할을 합니다.
    - 데이터베이스 쿼리 결과에서 한 행씩 읽어 처리하는 방식이므로, 메모리 사용을 최소화할 수 있습니다.

    ```java
    String sql = "SELECT * FROM users";

    List<User> users = jdbcTemplate.query(sql, new RowMapper<User>() {
        @Override
        public User mapRow(ResultSet rs, int rowNum) throws SQLException {
            User user = new User();
            user.setId(rs.getInt("id"));
            user.setName(rs.getString("name"));
            return user;
        }
    });
    ```

   이 방식은 `RowMapper`가 결과셋의 각 행에 대해 호출되므로, 전체 데이터를 메모리에 로드하지 않고 순차적으로 처리할 수 있습니다.

2. `ResultSetExtractor`

   `ResultSetExtractor`는 결과셋 전체를 처리할 때 사용됩니다.
   이는 결과셋 전체를 메모리에 로드하는 방식이므로, 대량의 데이터를 처리할 때는 메모리 사용량에 주의해야 합니다.
   `fetchAll()`과 유사한 방식이라고 볼 수 있습니다.

    ```java
    String sql = "SELECT * FROM users";

    List<User> users = jdbcTemplate.query(sql, new ResultSetExtractor<List<User>>() {
        @Override
        public List<User> extractData(ResultSet rs) throws SQLException {
            List<User> list = new ArrayList<>();
            while (rs.next()) {
                User user = new User();
                user.setId(rs.getInt("id"));
                user.setName(rs.getString("name"));
                list.add(user);
            }
            return list;
        }
    });
    ```

   이 방식은 `while (rs.next())` 루프를 통해 결과셋을 한 행씩 처리하지만,
   전체 결과셋을 메모리에 저장하는 콜렉션(`List`)에 넣기 때문에,
   메모리 사용이 `RowMapper` 방식에 비해 상대적으로 더 많아질 수 있습니다.

Spring JDBC는 JDBC 표준을 따르므로, JDBC의 기본 동작 원리와 유사합니다.
즉, 데이터베이스에서 쿼리 실행 시 결과셋이 네트워크를 통해 청크(chunk) 단위로 전송되며, JDBC 드라이버가 이를 `ResultSet` 객체로 제공합니다.
`ResultSet` 객체는 데이터베이스로부터 한 행씩 데이터를 가져올 수 있는 인터페이스를 제공하므로, Spring JDBC는 이를 통해 데이터를 스트리밍 방식으로 처리할 수 있습니다.

- `RowMapper`나 `ResultSetExtractor` 모두 JDBC의 `ResultSet`을 사용하여 한 행씩 데이터를 가져와 처리합니다. 이 때문에 대량의 데이터를 메모리 부담 없이 처리할 수 있습니다.
- 페이징 또는 배치 처리를 통해 결과셋이 매우 클 경우에도 이를 적절히 나누어 처리하는 것이 가능합니다.

Spring JDBC가 MySQL이나 다른 데이터베이스에서 데이터를 가져올 때, JDBC 드라이버는 네트워크를 통해 데이터베이스 서버와 통신합니다.
이 과정은 다음과 같이 이루어집니다:

1. SQL 쿼리 전송: 애플리케이션이 `JdbcTemplate`을 통해 SQL 쿼리를 작성하고, 이를 데이터베이스에 전송합니다.
2. 결과셋 수신: MySQL 서버는 쿼리를 처리하고, 그 결과를 네트워크 패킷으로 나누어 클라이언트(애플리케이션)로 전송합니다.
3. JDBC 드라이버는 데이터베이스 서버로부터 패킷을 순차적으로 수신하고, 이 데이터를 ResultSet으로 제공합니다.
4. Spring JDBC는 이 ResultSet을 통해 한 행씩 데이터를 가져와 애플리케이션에서 처리합니다.

이 과정은 청크 단위의 데이터 수신과 순차적 처리로 인해, 메모리를 효율적으로 사용할 수 있게 해줍니다.

Spring에서 대량의 데이터를 효율적으로 처리하려면 다음과 같은 방법을 사용할 수 있습니다:

1. 페이징 처리: 쿼리에 `LIMIT`와 `OFFSET`을 사용하여 데이터베이스에서 한 번에 처리할 데이터 양을 제한하는 방식입니다. 이를 통해 쿼리 결과를 여러 번에 나누어 가져오면 메모리 사용을 줄일 수 있습니다.

   ```java
   String sql = "SELECT * FROM users LIMIT ? OFFSET ?";
   List<User> users = jdbcTemplate.query(sql, new Object[] {limit, offset}, new UserRowMapper());
   ```

2. 배치 처리: 대량의 데이터를 한 번에 메모리에 로드하지 않고, 부분적으로 가져와 처리한 후 다음 데이터를 가져오는 방식입니다. 이를 통해 메모리 부담을 줄일 수 있습니다.

<!-- /curriculum-chunk -->

### hikari cp

#### 원문: hikari cp

<!-- curriculum-chunk: sha256=b8026b4047c16f08b387c6d4d400c47d6f9d063c95ad0deb890f5f18c8ca1013 major=spring-backend-frameworks mid=트랜잭션과 데이터 접근 sub=hikari cp sources=source/interview_questions.md:3674-3732, source/interviews.md:3674-3732 -->

> Source: `source/interview_questions.md:3674-3732`
> Classification reason: spring data access
> Duplicate source aliases: `source/interview_questions.md:3674-3732, source/interviews.md:3674-3732`

##### hikari cp

HikariCP는 고성능 JDBC 커넥션 풀로서, Java 애플리케이션에서 데이터베이스와의 연결을 효율적으로 관리하기 위해 사용됩니다.

- 데이터베이스 연결 비용: 데이터베이스에 새로운 연결을 생성하고 종료하는 것은 비용이 많이 드는 작업입니다.
- 효율성 향상: 커넥션 풀은 미리 일정 수의 연결을 생성해 두고 재사용함으로써 애플리케이션의 성능을 향상시킵니다.
- 자원 관리: 동시 연결 수를 제한하여 데이터베이스의 과부하를 방지합니다.

HikariCP의 주요 특징은 다음과 같습니다.

- 경량성: 다른 커넥션 풀에 비해 메모리 사용량이 적습니다.
- 고성능: 낮은 대기 시간과 높은 처리량을 제공합니다.
- 신뢰성: 안정적인 커넥션 관리와 복구 메커니즘을 갖추고 있습니다.
- 구성 용이성: 최소한의 설정으로도 효과적으로 동작하며, 필요한 경우 세부적인 튜닝이 가능합니다.

HikariCP의 기능들은 다음과 같습니다:

- 커넥션 풀링

  애플리케이션이 시작될 때 HikariCP는 설정된 최소 수(minimumIdle)의 커넥션을 생성합니다.

  커넥션 대여 및 반납:
    - 대여: 애플리케이션에서 데이터베이스 작업이 필요할 때 커넥션 풀에서 사용 가능한 커넥션을 대여합니다.
    - 반납: 작업이 끝나면 커넥션을 풀에 반납하여 다른 요청에서 재사용할 수 있게 합니다.

  설정된 최대 수(maximumPoolSize)를 초과하지 않도록 커넥션 생성을 관리합니다.

- 커넥션 유효성 검사

    - 일정한 주기로 커넥션의 유효성을 확인하여 문제가 있는 커넥션을 제거합니다.
    - keepalive: 오래된 커넥션이 방치되지 않도록 주기적으로 테스트 쿼리를 실행합니다.

- 초기화 시 데이터베이스 연결 실패를 빠르게 감지하여 애플리케이션이 즉시 대처할 수 있게 합니다.(fail-fast)
- 특정 시간 동안 사용되지 않은 커넥션을 풀에서 제거하여 자원을 관리합니다.
- 문제가 있는 커넥션을 자동으로 교체하여 안정성을 유지합니다.

요청이 많을 때 HikariCP의 동작은 다음과 같습니다.

1. 최대 풀 크기 도달하는 경우

   HikariCP는 설정된 `maximumPoolSize`까지 커넥션을 생성하여 동시에 처리할 수 있는 최대 연결 수를 제한합니다.

   요청이 많아지면 커넥션 풀이 최대 크기에 도달하고, 추가 요청은 커넥션이 반환될 때까지 대기하게 됩니다.
   이를 통해 데이터베이스에 과도한 부하가 걸리지 않도록 제어합니다.

   커넥션 풀 크기를 너무 작게 설정하면 대기 시간이 증가하고, 너무 크게 설정하면 데이터베이스에 과부하가 발생할 수 있습니다.

2. 커넥션 획득 대기

    - 커넥션 타임아웃:
        - `connectionTimeout` 설정에 따라, 커넥션을 얻기 위해 대기하는 최대 시간을 지정합니다.
        - 이 시간을 초과하면 예외(`SQLTransientConnectionException`)가 발생하여 애플리케이션이 대처할 수 있게 합니다.
    - 대기 큐 관리:
        - HikariCP는 내부적으로 대기 중인 스레드를 효율적으로 관리하여 성능 저하를 최소화합니다.

3. 스루풋과 대기 시간

   HikariCP는 최소한의 동기화와 락을 사용하여 스루풋을 최대화하고 대기 시간을 최소화합니다.

   애플리케이션 레벨에서 요청이 폭주하더라도, 커넥션 풀을 통해 데이터베이스의 안정성을 유지합니다.

<!-- /curriculum-chunk -->

### 리파지토리 메서드에 `@Transactional` 사용

#### 원문: 리파지토리 메서드에 `@Transactional` 사용

<!-- curriculum-chunk: sha256=eb171617889f2e1002729b67300822ac8d568333fcba10d61c743aeb46460ae3 major=spring-backend-frameworks mid=트랜잭션과 데이터 접근 sub=리파지토리 메서드에 `@Transactional` 사용 sources=source/interview_questions.md:3248-3312, source/interviews.md:3248-3312 -->

> Source: `source/interview_questions.md:3248-3312`
> Classification reason: spring data access
> Duplicate source aliases: `source/interview_questions.md:3248-3312, source/interviews.md:3248-3312`

##### 리파지토리 메서드에 `@Transactional` 사용

Spring에서의 트랜잭션 관리와 관련하여 `@Transactional` 어노테이션의 작동 방식, 현재의 트랜잭션 관리 구현 방식의 장단점, 그리고 베스트 프랙티스에 대해 차례대로 설명하겠습니다.

1. `@Transactional` 어노테이션의 작동 방식

   `@Transactional` 어노테이션은 AOP(Aspect-Oriented Programming) 방식을 사용하여 트랜잭션을 관리합니다.
   이 어노테이션이 적용된 메서드가 호출되면, 프록시 객체가 생성되어 트랜잭션의 시작과 종료, 롤백을 처리하게 됩니다.
   기본적인 작동 방식은 다음과 같습니다:

    - 트랜잭션 시작: `@Transactional`이 선언된 메서드가 호출되면, 트랜잭션이 시작됩니다. 이때, 트랜잭션 매니저가 메서드를 트랜잭션 범위 내에서 실행할 준비를 합니다.
    - 트랜잭션 커밋: 메서드가 정상적으로 종료되면, 트랜잭션이 커밋됩니다.
    - 트랜잭션 롤백: 메서드 실행 중에 체크되지 않은 예외(RuntimeException)가 발생하면, 트랜잭션은 롤백됩니다. 기본 설정에서는 `RuntimeException` 및 그 하위 클래스에서만 롤백이 발생하며, 체크 예외는 기본적으로 롤백되지 않습니다. 롤백 정책은 설정을 통해 수정할 수 있습니다.

2. 현재 구현된 방식의 장단점 (`@Transactional`이 repository 메서드에 붙은 경우)

   현재 구현된 방식은 repository 레벨에서 각 메서드마다 `@Transactional`을 적용한 형태입니다.
   이를 통해 각 메서드가 호출될 때마다 트랜잭션이 시작되고, 해당 메서드 실행에만 트랜잭션이 적용됩니다.

    - 장점
        1. 세분화된 트랜잭션 관리: 트랜잭션을 더 세밀하게 관리할 수 있습니다. 각 repository 메서드마다 트랜잭션이 독립적으로 처리되므로, 한 메서드에 문제가 발생해도 다른 메서드의 트랜잭션에 영향을 주지 않습니다.
        2. 특정 작업에만 트랜잭션 적용 가능: 트랜잭션이 반드시 필요한 메서드에서만 트랜잭션을 적용할 수 있으므로, 성능 최적화에 도움이 될 수 있습니다. 특히, 읽기 전용 작업에서 트랜잭션을 피하거나, 특정 작업에서만 트랜잭션을 사용할 수 있습니다.

    - 단점
        1. 트랜잭션 경계 문제:

           여러 repository 메서드를 호출하는 애플리케이션 레벨의 서비스가 있다면, 트랜잭션이 메서드 호출 단위로 분리됩니다.
           즉, 하나의 작업이 여러 repository 메서드를 호출하는 경우, 각 호출마다 별도의 트랜잭션이 적용되므로, 작업 전체를 하나의 트랜잭션으로 처리하지 못합니다. 이는 데이터 일관성 문제가 발생할 수 있음을 의미합니다.
            - 예: `saveOrder()` 메서드와 `savePayment()` 메서드를 각각 호출하는 서비스에서, 트랜잭션이 각각의 메서드에만 적용되어 두 작업이 원자적으로 처리되지 않는 문제.

        2. 트랜잭션 전파 문제:

           repository 메서드에 `@Transactional`을 적용하면, 상위 레벨(서비스나 애플리케이션 레이어)에서 트랜잭션 전파 설정이 정확히 관리되지 않을 수 있습니다.
           상위 레이어에서 트랜잭션 전파 규칙을 정의하지 않으면, 트랜잭션 범위가 예상치 못하게 작동할 수 있습니다.

        3. 복잡한 트랜잭션 제어:

           트랜잭션이 여러 repository 메서드에 분산되어 있는 경우, 트랜잭션 관리가 복잡해지고, 전체 비즈니스 로직을 하나의 트랜잭션 단위로 처리하는 것이 어려울 수 있습니다.

베스트 프랙티스는 트랜잭션 관리를 서비스 레이어나 애플리케이션 레이어에서 중앙 집중적으로 관리하는 것입니다.
일반적으로 비즈니스 로직을 담당하는 서비스 계층에서 `@Transactional`을 사용하는 방식이 더 효율적이고 유지 관리에 용이합니다.
이유는 다음과 같습니다:

1. 비즈니스 로직 기반 트랜잭션:

   트랜잭션을 비즈니스 로직 단위로 묶어서 관리할 수 있습니다.
   즉, 여러 repository 메서드 호출을 포함하는 하나의 비즈니스 로직 전체가 원자적으로 실행되며, 데이터 일관성을 보장할 수 있습니다.
    - 예: `createOrder()` 메서드가 `saveOrder()`와 `savePayment()`를 호출하는 경우, 이 전체 작업을 하나의 트랜잭션으로 처리하여 문제 발생 시 전체가 롤백됩니다.

2. 단순한 트랜잭션 경계 설정:

   트랜잭션의 경계를 서비스 계층에서 설정하면, 트랜잭션 관리가 단순해집니다.
   여러 레이어에 걸쳐 트랜잭션이 흩어져 있지 않으므로, 예기치 못한 트랜잭션 문제를 방지할 수 있습니다.

3. 트랜잭션 전파 제어 용이:

   서비스 레벨에서 트랜잭션을 관리하면, 트랜잭션 전파(propagation) 옵션을 일관되게 제어할 수 있습니다.
   상위 레이어에서 트랜잭션을 전파받아 여러 작업을 하나로 처리할지, 각각 별도로 처리할지를 명확하게 결정할 수 있습니다.
    - 예: 하나의 메서드에서 또 다른 트랜잭션을 호출할 때, Propagation.REQUIRED나 Propagation.REQUIRES_NEW 등의 정책을 쉽게 설정 가능.

4. 단일 트랜잭션 관리 포인트: 서비스 레이어에서 트랜잭션을 관리하면, 트랜잭션의 시작, 커밋, 롤백을 일관성 있게 관리할 수 있습니다. 이를 통해 트랜잭션 관련 버그를 쉽게 찾고 해결할 수 있습니다.

하지만 서비스 레이어에서 트랜잭션을 관리하는 경우, 일부 메서드에 트랜잭션이 필요 없을 때도 트랜잭션을 사용하게 되는 경우가 있을 수 있습니다.
하지만 이런 경우는 `readOnly = true`와 같은 옵션을 통해 최적화할 수 있습니다.

<!-- /curriculum-chunk -->
