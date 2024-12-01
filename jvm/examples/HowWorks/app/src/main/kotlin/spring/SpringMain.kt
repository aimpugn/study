package spring

import main.spring.MyTestProperty
import org.springframework.boot.*
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.context.properties.bind.Bindable
import org.springframework.boot.context.properties.bind.Binder
import org.springframework.context.annotation.Description

/**
 * `open` 이어야 합니다.
 * ```
 * org.springframework.beans.factory.parsing.BeanDefinitionParsingException: Configuration problem: @Configuration class 'SpringMain' may not be final. Remove the final modifier to continue.
 * ```
 */
@SpringBootApplication
@Description(value = "Spring Main Class to Learn Spring")
open class SpringMain

/**
 * 스프링 부트의 실행 과정([SpringApplication.run])은 다음 과정을 거칩니다:
 * 1. 부트스트랩 컨텍스트 생성: [SpringApplication.createBootstrapContext]
 * 2. 주요 이벤트를 감지하고 처리하기 위한 `RunListener` 얻기: [SpringApplication.getRunListeners]
 * 3. 환경 준비하기: [SpringApplication.prepareEnvironment]
 * 4. 애플리케이션 컨텍스트 생성: [SpringApplication.createApplicationContext]
 * 5. 애플리케이션 컨텍스트 준비: [SpringApplication.prepareContext]
 * 6. 애플리케이션 컨텍스트 리프레시: [SpringApplication.refreshContext]
 *
 * ### [SpringApplication] 생성
 *
 * 최초 생성시 컨텍스트 이니셜라이저와 리스너들을 추가합니다.
 * - [org.springframework.context.ApplicationContextInitializer]를 구현한 인스턴스들을 [SpringApplication.initializers]에 추가
 * - [org.springframework.context.ApplicationListener]를 구현한 인스턴스들을 [SpringApplication.listeners]에 추가
 *
 * 컨텍스트 이니셜라이저나 애플리케이션 리스너는 spring boot 시에 필요한 각 패키지의 `spring.factories` 파일에 정의되어 있습니다. 예를 들어:
 * - [spring-boot/src/main/resources/META-INF/spring.factories](https://github.com/spring-projects/spring-boot/blob/32433e84f3048c397aef3c3a8d63a0ae1ca7d5fc/spring-boot-project/spring-boot/src/main/resources/META-INF/spring.factories)
 * - [spring-boot-autoconfigure/src/main/resources/META-INF/spring.factories](https://github.com/spring-projects/spring-boot/blob/32433e84f3048c397aef3c3a8d63a0ae1ca7d5fc/spring-boot-project/spring-boot-autoconfigure/src/main/resources/META-INF/spring.factories)
 *
 * 웹 애플리케이션의 타입은 [WebApplicationType.deduceFromClasspath]을 통해 추론하여 [SpringApplication.webApplicationType]을 설정합니다.
 *
 * ### 부트스트랩 컨텍스트 생성
 *
 * > [SpringApplication.createBootstrapContext]
 *
 * [SpringApplication.createBootstrapContext]은 [DefaultBootstrapContext] 인스턴스를 생성합니다.
 * [DefaultBootstrapContext]는 애플리케이션 컨텍스트가 생성되기 전에 필요한 리소스나 설정을 등록하고 관리하는 컨테이너입니다.
 * 초기화 단계에서 필요한 컴포넌트들을 미리 준비하여 애플리케이션의 원활한 실행을 지원합니다.
 *
 * 부트스트랩 컨텍스를 생성할 때 `META-INF/spring.factories`([org.springframework.core.io.support.SpringFactoriesLoader.FACTORIES_RESOURCE_LOCATION]) 파일을 통해
 * [BootstrapRegistryInitializer] 타입의 인스턴스들을 동적으로 로드하여 초기화를 수행합니다.
 *
 * [BootstrapRegistryInitializer]를 통해 환경 정보 준비 이전에 필요한 리소스를 설정할 수 있습니다.
 * 예를 들어,
 * - Spring Cloud Config와 같은 외부 구성 서버를 사용하는 경우, 애플리케이션이 시작될 때 해당 서버로부터 설정을 가져와야 합니다.
 *   이때 부트스트랩 단계에서 구성 서버와의 연결을 설정할 수 있습니다.
 * - 애플리케이션이 시작될 때 AWS Secrets Manager나 HashiCorp Vault와 같은 비밀 관리 서비스로부터
 *   데이터베이스 자격 증명이나 API 키와 같은 민감하지만 필요한 정보를 미리 로드할 수 있습니다.
 * - 애플리케이션의 로깅 설정을 외부 시스템이나 파일에서 동적으로 가져와야 하는 경우, 부트스트랩 단계에서 로깅 설정을 초기화할 수 있습니다.
 *
 * ### 주요 이벤트를 감지하고 처리하기 위한 `RunListener` 얻기
 *
 * > [SpringApplication.getRunListeners]
 *
 * [SpringApplicationRunListener] 인터페이스를 구현한 클래스들을 로드합니다.
 * 별다른 설정이 없다면 [org.springframework.boot.context.event.EventPublishingRunListener]가 로드됩니다.
 *
 * `EventPublishingRunListener`는 내부적으로 [org.springframework.context.event.SimpleApplicationEventMulticaster.multicastEvent]를 사용하여
 * 이벤트를 전파합니다. 가령 애플리케이션의 실행 과정에서 특정 이벤트들(예: 초기화, 환경 설정, 컨텍스트 준비 등)을 전파하고 특정 이벤트마다 커스텀 로직을 실행할 수 있게 해줍니다.
 * 전파되는 이벤트들의 이름은 [SpringApplicationRunListeners]에서, 타입은 [org.springframework.boot.context.event.EventPublishingRunListener]에서 확인할 수 있습니다.
 *
 * 그리고 [SpringApplicationRunListeners] 인스턴스를 생성하여 반환합니다.
 *
 * ### 환경 준비하기
 *
 * > [SpringApplication.prepareEnvironment]
 *
 * [org.springframework.boot.ApplicationContextFactory.DEFAULT] 즉, [DefaultApplicationContextFactory.createEnvironment]를 사용해서
 * [org.springframework.core.env.ConfigurableEnvironment] 를 생성합니다.
 * 이때 `META-INF/spring.factories` 등록된 [ApplicationContextFactory] 구현체중, 주어진 [WebApplicationType]를 지원하는 팩토리를 사용합니다.
 * - [org.springframework.boot.web.servlet.context.ServletWebServerApplicationContextFactory.createEnvironment]
 * - [org.springframework.boot.web.reactive.context.ReactiveWebServerApplicationContextFactory.createEnvironment]
 *
 * 환경 [SpringApplication.bindToSpringApplication]에서 `application.yaml`이나 `application.properties` 파일에 정의된 속성들을
 * `SpringApplication` 인스턴스의 속성에 바인딩합니다. 가령:
 * - `spring.main.web-application-type`은 [SpringApplication.webApplicationType]에 설정
 * - `spring.main.banner-mode`는 [SpringApplication.bannerMode]에 설정
 * - `spring.main.lazy-initialization`는 [SpringApplication.lazyInitialization]에 설정
 * - [그 외 공통 애플리케이션 속성 목록](https://docs.spring.io/spring-boot/appendix/application-properties/index.html) 참고
 *
 * 속성의 소스가 로드되는 순서는 [링크](https://docs.spring.io/spring-boot/reference/features/external-config.html)를 참고합니다.
 *
 * 환경 준비가 끝나면 "spring.boot.application.environment-prepared"([org.springframework.boot.context.event.ApplicationEnvironmentPreparedEvent]) 이벤트를 전파합니다.
 *
 * ### 애플리케이션 컨텍스트 생성
 *
 * > [SpringApplication.createApplicationContext]
 *
 * [WebApplicationType]에 따라 애플리케이션 컨텍스트를 생성합니다:
 * - [org.springframework.boot.web.servlet.context.ServletWebServerApplicationContextFactory.create]
 *   => [org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext] 생성
 * - [org.springframework.boot.web.reactive.context.ReactiveWebServerApplicationContextFactory.create]
 *   => [org.springframework.boot.web.reactive.context.AnnotationConfigReactiveWebServerApplicationContext] 생성
 *
 * ### 애플리케이션 컨텍스트 준비
 *
 * > [SpringApplication.prepareContext]
 *
 * 앞서 [SpringApplication.createApplicationContext] 통해 생성된 컨텍스트를 사용할 수 있도록 준비합니다.
 * 이를 위해 [SpringApplication] 인스턴스 생성 시 `spring.factories` 통해서 로드된 컨텍스트 이니셜라이저들의 `initialize`를 실행합니다.
 * - [org.springframework.boot.autoconfigure.SharedMetadataReaderFactoryContextInitializer.initialize]
 * - [org.springframework.boot.context.config.DelegatingApplicationContextInitializer.initialize]
 * - [org.springframework.boot.context.ContextIdApplicationContextInitializer.initialize]
 * - [org.springframework.boot.autoconfigure.logging.ConditionEvaluationReportLoggingListener.initialize]
 * - [org.springframework.boot.context.ConfigurationWarningsApplicationContextInitializer.initialize]
 * - [org.springframework.boot.rsocket.context.RSocketPortInfoApplicationContextInitializer.initialize]
 * - [org.springframework.boot.web.context.ServerPortInfoApplicationContextInitializer.initialize]
 *
 * 초기화 후
 * 1. 애플리케이션 컨테스트가 준비되었다는 "spring.boot.application.context-prepared"([org.springframework.boot.context.event.ApplicationContextInitializedEvent]) 이벤트를 전파합니다.
 * 2. 부트스트랩 컨테스트를 종료합니다.
 *
 * 이후 컨텍스트가 리프레시되기 전에 애플리케이션 컨텍스트의 빈 팩토리에 필요한 싱글톤 빈 등록, 빈 팩토리의 동작 방식 설정, 빈 팩토리 후 처리기(PostProcessor) 추가 등의 작업을 수행합니다.
 * 참고로 Servlet 경우 [org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext.beanFactory]를 사용하게 됩니다.
 * - "springApplicationArguments" 빈을 싱글톤 빈으로 등록합니다.
 *   `@Autowired` 또는 `@Inject`를 통해 [ApplicationArguments]를 주입받을 수 있게 됩니다.
 *    ```
 *    @Component
 *    public class MyComponent {
 *        private final ApplicationArguments args;
 *
 *        public MyComponent(ApplicationArguments args) {
 *            this.args = args;
 *        }
 *    }
 *    ```
 * - 빈 팩토리가 [org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory]이면 순환 참조 허용 여부(`spring.main.allow-circular-references`) 설정
 * - 빈 팩토리가 [org.springframework.beans.factory.support.DefaultListableBeanFactory]이면 빈 정의 재정의 허용 여부(`spring.main.allow-bean-definition-overriding`)를 설정
 * - 지연 초기화 설정(`spring.main.lazy-initialization`)이면 [LazyInitializationBeanFactoryPostProcessor] 추가
 *
 * 그리고 AOT(Ahead-of-Time) 컴파일이 아닌 경우, 애플리케이션의 소스들을 컨텍스트에 로드([SpringApplication.load])합니다.
 * 이 로드 과정을 통해 [SpringBootApplication]이 붙은 클래스(예: [SpringMain])를 기반으로 빈 정의([BeanDefinitionLoader])를 생성하고 빈을 로드합니다.
 * 그 이유는:
 * - [SpringBootApplication]이 붙은 클래스는 애플리케이션의 빈 정의를 시작하는 역할을 합니다.
 * - [SpringBootApplication]이 붙은 [SpringMain]는 실제로는 구성([org.springframework.context.annotation.Configuration]) 클래스로 사용되므로,
 *   해당 클래스에 설정된 [org.springframework.context.annotation.Primary], [org.springframework.context.annotation.Description] 등을 설정합니다.
 * - [org.springframework.context.annotation.ComponentScan] 어노테이션을 포함하고 있어, 지정된 패키지 및 그 하위 패키지에서 컴포넌트들을 스캔하고 빈으로 등록합니다.
 * - [org.springframework.boot.autoconfigure.EnableAutoConfiguration] 어노테이션을 포함하고 있어, 스프링 부트의 자동 구성을 활성화합니다.
 *
 * 구체적인 과정은 다음과 같습니다.
 * 1. [org.springframework.boot.BeanDefinitionLoader.load] 통해서 로드하고,
 * 2. [org.springframework.context.annotation.AnnotatedBeanDefinitionReader.register]가 실행되고,
 * 3. [org.springframework.context.annotation.AnnotatedBeanDefinitionReader.doRegisterBean]가 호출됩니다.
 *   여기서 주어진 bean class는 [SpringBootApplication] 어노테이션이 붙은 "class spring.SpringMain" 클래스입니다.
 *   [org.springframework.context.annotation.AnnotationConfigUtils.processCommonDefinitionAnnotations]를 통해 공통 어노테이션들을 처리합니다.
 * 4. [SpringMain] 클래스에 선언된 애노테이션에서 메타데이터를 추출하여 빈 정의를 생성하고 등록합니다.
 *   좀 더 구체적으로는 [org.springframework.beans.factory.support.DefaultListableBeanFactory.registerBeanDefinition] 통해 빈 정의 맵에 등록됩니다.
 *
 * 컨텍스트 준비가 끝나면 "spring.boot.application.context-loaded"([org.springframework.boot.context.event.ApplicationPreparedEvent]) 이벤트를 전파합니다.
 *
 * ### 컨텍스트 리프레시
 *
 * > [SpringApplication.refreshContext]
 *
 * [SpringApplication.refresh]는 [org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext.refresh]를 실행하고,
 * 서블릿 웹 앱인 경우 [org.springframework.context.support.AbstractApplicationContext.refresh]가 실행됩니다.
 *
 * - 리프레시 위해 컨텍스트 준비: [org.springframework.context.support.AbstractApplicationContext.prepareRefresh]
 *    - 시작 시간 설정
 *    - `active=true` 설정
 *    - `ApplicationServletEnvironment`에서 서블릿 기반 `StubPropertySource`를 주어진 서블릿 컨텍스트와 서블릿 구성으로 생성된 실제 인스턴스로 대체 등
 *
 * - 빈 팩토리 준비: [org.springframework.context.support.AbstractApplicationContext.prepareBeanFactory].
 *    - 빈 팩토리에 스프링 표현식 언어(`SpEL`)를 처리할 수 있는 빈 표현 리졸버(`StandardBeanExpressionResolver`) 설정
 *    - 프로퍼티 값을 적절한 타입으로 변환하기 위해 `Property` 에디터(`ResourceEditorRegistrar`) 추가
 *    - `*Aware` 인터페이스 구현한 빈에 애플리케이션 컨텍스트를 제공하는 `ApplicationContextAwareProcessor`를 이 빈 팩토리 통해 생성되는 빈에 적용될 수 있도록
 *      후처리기로 등록하고, `ApplicationContextAwareProcessor`가 처리해주는 인터페이스들을 일반적인 의존성 주입 대상에서 제외합니다.
 *    - `ResourceLoader`, `ApplicationEventPublisher`, `ApplicationContext` 타입의 빈을 요청할 때, 현재 컨텍스트를 제공하도록 설정합니다.
 *    - `ApplicationListener` 인터페이스를 구현한 빈을 감지하고, 애플리케이션 이벤트 멀티캐스터에 등록하거나 제거하는 `ApplicationListenerDetector`를
 *      Bean Post Processor로 추가
 *
 * - 빈 팩토리 후처리([org.springframework.context.support.AbstractApplicationContext.postProcessBeanFactory])
 *
 *   [org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext.postProcessBeanFactory]가 호출되고,
 *   빈 팩토리에 [org.springframework.boot.web.servlet.context.WebApplicationContextServletContextAwareProcessor]가 등록됩니다.
 *
 * - 빈 팩토리 후처리기 호출([org.springframework.context.support.AbstractApplicationContext.invokeBeanFactoryPostProcessors])
 *
 *   [org.springframework.context.support.PostProcessorRegistrationDelegate.invokeBeanFactoryPostProcessors]를 거치게 되고,
 *   [org.springframework.context.annotation.ConfigurationClassPostProcessor.processConfigBeanDefinitions]가 호출됩니다.
 *   결국 [org.springframework.context.annotation.ConfigurationClassParser.doProcessConfigurationClass]가 호출되는데,
 *   여기서 [org.springframework.context.annotation.PropertySources], [org.springframework.context.annotation.ComponentScan] 등을 처리합니다.
 *
 * - 빈 후처리기 등록([org.springframework.context.support.AbstractApplicationContext.registerBeanPostProcessors])
 *
 *   [org.springframework.context.support.PostProcessorRegistrationDelegate.registerBeanPostProcessors]를 호출하는데,
 *   여기서 `PriorityOrdered`, `Ordered`, `BeanPostProcessor` 타입의 후처리기를 등록합니다.
 *
 * - 메시지 소스 초기화([org.springframework.context.support.AbstractApplicationContext.initMessageSource])
 *
 * - 애플리케이션 이벤트 멀티캐스터 초기화([org.springframework.context.support.AbstractApplicationContext.initApplicationEventMulticaster])
 *
 * 컨텍스트별로 리프레시를 수행할 수 있도록 [org.springframework.context.support.AbstractApplicationContext.onRefresh] 메서드 시그니처를 정의하는데,
 * 실제 구현체는 [WebApplicationType]에 따라 상이합니다:
 * - 서블릿: [org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext.onRefresh]
 * - 리액티브: [org.springframework.boot.web.reactive.context.ReactiveWebServerApplicationContext.onRefresh]
 *
 * 이때 실제로 톰캣 웹 서버를 생성([org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext.createWebServer])합니다.
 *
 * 웹 서버 생성 역시 [WebApplicationType]에 따라 적절한 팩토리에 위임([[org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext.getWebServerFactory]])합니다.
 * 이때 빈 팩토리를 통해서 팩토리를 가져오는데, [org.springframework.boot.autoconfigure.web.embedded.TomcatWebServerFactoryCustomizer.customize]를
 * 거치게 됩니다. 이 커스터마이저에서 최소 스레드 수, 최대 스레드 수, [org.apache.tomcat.util.net.NioEndpoint.maxQueueSize], URI 인코딩 등
 * 톰캣 웹 서버를 생성하는 데 필요한 구성 값들을 설정합니다.
 *
 * 그리고 현재 환경에 해당하는 팩토리의 `getWebServer` 메서드를 실행합니다.
 * - 서블릿: [org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory.getWebServer]
 * - 리액티브: [org.springframework.boot.web.embedded.tomcat.TomcatReactiveWebServerFactory.getWebServer]
 *
 * 서블릿 웹 서버 팩토리에서 웹 서버를 생성(`getWebServer`)할 때 [org.apache.catalina.connector.Connector]라는 것을 생성하는데,
 * 이 커넥터는 톰캣에서 네트워크 계층과 애플리케이션 계층의 연결을 담당합니다. 주요 역할은 다음과 같습니다:
 * - 네트워크 소켓을 열어 클라이언트로부터 HTTP 요청을 수신합니다.
 * - `HTTP`(1.1, 2), [`AJP`](https://stackoverflow.com/a/27372321) 등의 프로토콜 지원하며 요청 데이터를 적절히 파싱하고 처리합니다.
 * - [org.apache.catalina.connector.CoyoteAdapter]를 통해 [org.apache.catalina.startup.Tomcat]의 내부 서블릿 컨테이너로 요청을 전달합니다.
 *   어댑터는 요청이 톰캣의 서블릿 컨테이너로 라우팅되도록 돕습니다.
 * - 컨테이너에서 생성된 응답을 클라이언트로 전달합니다.
 *
 * [org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory.getTomcatWebServer]가 실행되고,
 * 톰캣 웹 서버 인스턴스가 생성되면서 [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.initialize]를 수행합니다.
 *
 * `initialize`중에 [org.apache.catalina.startup.Tomcat.server]에 설정된 [org.apache.catalina.core.StandardServer.start]가 실행됩니다.
 * 1. [org.apache.catalina.core.StandardServer.state]는 "NEW"이므로 [org.apache.catalina.core.StandardServer.init]과
 * [org.apache.catalina.core.StandardServer.initInternal]이 실행됩니다.
 *
 * 2. 그리고 [org.apache.catalina.core.StandardServer]에 설정되어 있는 [org.apache.catalina.core.StandardService.init]과
 *   [org.apache.catalina.core.StandardService.initInternal]이 실행됩니다.
 *
 * 3. [org.apache.catalina.core.StandardService]에 설정된 [org.apache.catalina.core.StandardEngine.init]과
 *   [org.apache.catalina.core.StandardEngine.initInternal]이 실행됩니다.
 *
 * 4. [org.apache.catalina.core.StandardService]에 설정된 [org.apache.catalina.mapper.MapperListener.init]과
 *   [org.apache.catalina.mapper.MapperListener.initInternal]이 실행됩니다.
 *
 * 5. 마지막으로 [org.apache.catalina.core.StandardService]에 설정된 [org.apache.catalina.connector.Connector.init]과
 *   [org.apache.catalina.connector.Connector.initInternal]이 실행됩니다.
 *   이 커넥터에서 실질적으로 요청을 라우팅하는 [org.apache.catalina.connector.CoyoteAdapter] 생성하여 서비스에 설정하고,
 *   [org.apache.coyote.http11.Http11NioProtocol.init]를 실행합니다.
 *
 * 톰캣 경우 HTTP 요청을 받을 것이므로 [org.apache.coyote.http11.Http11NioProtocol]를 생성([org.apache.coyote.ProtocolHandler.create])합니다.
 * 그리고 다시 [org.apache.tomcat.util.net.NioEndpoint]를 [org.apache.coyote.AbstractProtocol.endpoint]에 설정하는데,
 * 추후 이 엔드포인트 구현체를 통해 각각 다음의 역할을 수행하는 스레드들을 실행하게 됩니다.
 * - [org.apache.tomcat.util.net.Acceptor]
 * - [org.apache.catalina.Executor]
 * - [org.apache.tomcat.util.net.NioEndpoint.Poller]
 *
 * 하지만 여기서는 커넥터와 [org.apache.catalina.core.StandardService], [org.apache.catalina.startup.Tomcat]가 서로 상호 바인딩하거나,
 * 커넥터에 포트나 호스트 등을 설정([org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory.customizeConnector]) 등을 수행합니다.
 *
 * [org.springframework.context.support.AbstractApplicationContext.finishRefresh]로 리프레시 과정을 끝나면,
 * [org.springframework.context.support.DefaultLifecycleProcessor.onRefresh]로 리프레시 과정중 라이프사이클 프로세서가 수행해야 하는 작업을 시작합니다.
 * - [org.springframework.context.support.DefaultLifecycleProcessor.startBeans] 실행
 * - [org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext.createWebServer]에서
 *   `"webServerStartStop"`라는 싱글톤으로 등록한 [org.springframework.boot.web.servlet.context.WebServerStartStopLifecycle.start] 실행
 * - [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.start]로 웹 서버 시작
 *
 * [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.start] 과정은 다음과 같습니다.
 * 1. [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.addPreviouslyRemovedConnectors]가 실행됩니다.
 * 2. [org.apache.catalina.core.StandardService.addConnector]에서 라이프사이클이 현재는 "STARTED" 상태이므로
 *   앞서 생성했던 [org.apache.catalina.connector.Connector.start] 및 [org.apache.catalina.connector.Connector.startInternal]가 실행됩니다.
 * 3. 커넥터의 `startInternal`에서 [org.apache.coyote.http11.Http11NioProtocol.start], [org.apache.tomcat.util.net.NioEndpoint.start],
 *   그리고 [org.apache.tomcat.util.net.NioEndpoint.startInternal]이 실행됩니다.
 *   [org.apache.tomcat.util.net.NioEndpoint.executor]가 없으므로 [org.apache.tomcat.util.net.NioEndpoint.createExecutor] 실행되면서
 *   [org.apache.tomcat.util.threads.ThreadPoolExecutor]가 생성됩니다. 추후 이 스레드 풀에 의해 요청을 처리하는 `http-nio-8080-exec-N`이 생성됩니다.
 * 4. [org.apache.tomcat.util.threads.ThreadPoolExecutor] 생성, 연결 수 제한 등 관리를 위한 [org.apache.tomcat.util.threads.LimitLatch] 설정
 * 5. [org.apache.tomcat.util.net.NioEndpoint.Poller], [org.apache.tomcat.util.net.Acceptor] 스레드를 실행
 *
 * 그리고 [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.initialize] 마지막에
 * [org.springframework.boot.web.embedded.tomcat.TomcatWebServer.startNonDaemonAwaitThread]가 실행되고,
 * 별도의 새로운 스레드에서 요청을 수신하는 [org.apache.catalina.core.StandardServer.await]가 실행됩니다.
 *
 * [org.apache.catalina.core.StandardServer.awaitSocket]에 "localhost:8080"으로 요청을 수신하는 서버 소켓이 생성되고,
 * 반복문 안에서 [java.net.ServerSocket.accept]로 요청이 들어오길 기다립니다.
 *
 * 톰캣의 요청 처리 과정을 아래에서 별도로 정리합니다.
 *
 * ### 톰캣 웹 서버 실행 및 요청 처리 과정
 *
 * 다음과 같이 요청을 하면:
 * ```
 * curl localhost:8080
 * ```
 *
 * `http-nio-8080-Acceptor` 스레드의 [org.apache.tomcat.util.net.Acceptor]에서 이를 `accept` 합니다.
 * ```
 * socket = endpoint.serverSocketAccept();
 * ```
 * - [org.apache.tomcat.util.net.NioEndpoint.serverSocketAccept]에서 [sun.nio.ch.ServerSocketChannelImpl.accept]가 실행되고,
 *
 * - [sun.nio.ch.ServerSocketChannelImpl.implAccept], 실제로는 [sun.nio.ch.Net.accept]에서 `accept` 하고,
 *   요청을 처리하기 위해 새로운 포트의 새로운 소켓을 생성합니다. 해당 소켓에 대한 [java.io.FileDescriptor]도 함께 리턴받습니다.
 *    ```
 *    ❯ lsof -p 44358 | rg 61201
 *    java    44358 rody   59u    IPv6 0xae1aa036c93b94e8       0t0                 TCP localhost:http-alt->localhost:61201 (ESTABLISHED)
 *    ```
 *    - local: /[0:0:0:0:0:0:0:1]:8080
 *    - remote: /[0:0:0:0:0:0:0:1]:61201
 *    - `java.io.FileDescriptor.fd`: 59
 *
 * - [org.apache.tomcat.util.net.NioEndpoint.Poller.addEvent]로 [org.apache.tomcat.util.collections.SynchronizedQueue]에 요청을 이벤트로 추가하고
 *   Acceptor는 다시 요청을 수신하러 돌아갑니다. 요청의 처리는 이벤트 큐를 기반으로 동작합니다.
 *
 * 동시에 별도 스레드에서 실행중인 [org.apache.tomcat.util.net.NioEndpoint.Poller]에서는 다음과 같은 작업들이 수행됩니다.
 * [org.apache.tomcat.util.net.NioEndpoint.Poller.selector]인 [sun.nio.ch.SelectorImpl.select]
 * -> [sun.nio.ch.SelectorImpl.lockAndDoSelect]
 * -> [sun.nio.ch.KQueueSelectorImpl.doSelect]
 * -> [sun.nio.ch.KQueue.poll]
 * 이렇게 호출되는데, 이 `KQueue.poll`은 8080 포트로 리스닝중인 소켓에 요청이 들어오는지 감지하는 static native 메서드입니다.
 *
 * 이런 소켓 이벤트 감지하는 멀티플렉싱 기능들은 플랫폼별로 다릅니다:
 * - [macosx 경우](https://github.com/openjdk/jdk/blob/50b4cbd8a4159a8657f4525e4023f3a498020493/src/java.base/macosx/native/libnio/ch/KQueue.c#L86-L113)
 * - [linux 경우](https://github.com/openjdk/jdk/blob/50b4cbd8a4159a8657f4525e4023f3a498020493/src/java.base/linux/native/libnio/ch/EPoll.c#L81-L96)
 *
 * 현재 테스트중인 MacBook 경우 `kqueue`를 사용중이며,
 * 만약 리눅스라면 [java.nio.channels.Selector.open]시 [`EPollSelectorImpl`](https://github.com/openjdk/jdk/blob/50b4cbd8a4159a8657f4525e4023f3a498020493/src/java.base/linux/classes/sun/nio/ch/EPollSelectorImpl.java#L93-L136)의
 * `doSelect`가 실행될 겁니다.
 *
 * 요청이 들어오면 이벤트를 감지하고, [org.apache.tomcat.util.net.NioEndpoint.Poller.timeout]으로 타임아웃 여부를 체크합니다.
 * [sun.nio.ch.KQueueSelectorImpl.processUpdateQueue]에서 큐에 이벤트를 등록하고,
 *
 * [org.apache.tomcat.util.net.NioEndpoint.processSocket]에서 [org.apache.tomcat.util.threads.ThreadPoolExecutor.execute]로
 * `http-nio-8080-exec-N` 스레드에서 처리가 됩니다.
 *
 * `accept`한 요청이 큐에 담겼으면, 그 큐를 처리하는 것은 [org.apache.tomcat.util.threads.ThreadPoolExecutor]입니다.
 * 8080 포트는 계속 요청을 들어야 하므로 빠르게 다시 [org.apache.tomcat.util.net.Acceptor]가 `accept`하게 되고,
 * 그때 새로 생성된 소켓은 [org.apache.tomcat.util.net.NioEndpoint.SocketProcessor.doRun]을 통해
 * [org.apache.coyote.AbstractProtocol.ConnectionHandler.process]로 전달됩니다.
 *
 * 그리고 다시 [org.apache.coyote.http11.Http11Processor.process]에서 해당 소켓이 `"OPEN_READ"` 상태인 경우
 * [org.apache.coyote.http11.Http11Processor.service]가 호출됩니다.
 *
 * [org.apache.coyote.http11.Http11InputBuffer.parseRequestLine]에서 이제 실제로 요청을 라인 바이 라인으로 읽습니다.
 * ```
 * GET / HTTP/1.1
 * Host: localhost:8080
 * User-Agent: curl/8.7.1
 * Accept: 모두
 * ```
 *
 * 그리고 실제로 헤더를 파싱합니다([org.apache.coyote.http11.Http11InputBuffer.parseHeaders]).
 *
 * [org.apache.coyote.http11.Http11Processor.adapter]인 [org.apache.catalina.connector.CoyoteAdapter.service]로 요청과 응답이 전달되고,
 * 내부에서 [org.apache.catalina.connector.Request], [org.apache.catalina.connector.Response]로 형변환 됩니다.
 *
 * [org.apache.catalina.connector.CoyoteAdapter]가 요청을 서블릿 컨테이너로 라우팅하는 순서는 다음과 같습니다.
 * 참고로 이때 컨텍스트는 [org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory.prepareContext]에서 생성된
 * [org.springframework.boot.web.embedded.tomcat.TomcatEmbeddedContext]입니다.
 *
 * ```
 * connector.getService().getContainer().getPipeline().getFirst().invoke(request, response);
 *
 * Tomcat
 * └─ StandardServer                                  <-- Tomcat 서버의 최상위 컴포넌트로서, 서버의 전체적인 설정과 라이프사이클을 관리합니다.
 *    └─ StandardService                              <-- 하나의 서버에 여러 서비스를 정의할 수 있으며, `Connector`, `Engine` 등을 관리합니다.
 *       ├── Connector                                <-- `ProtocolHandler`와 `Endpoint`를 통해 요청을 받아들이고, `CoyoteAdapter`를 통해 `Engine`으로 전달합니다.
 *       │   ├── ProtocolHandler
 *       │   ├── Endpoint
 *       │   └── CoyoteAdapter
 *       ├── MapperListener                           <-- `Mapper`와 함께 동작하여 요청 URL을 적절한 `Context`나 `Wrapper`로 매핑합니다.
 *       └── StandardEngine                           <-- 실제 요청 처리 로직을 담당하며, 가상 호스트를 관리합니다.
 *           ├── StandardPipeline                     <-- 엔진 레벨에서의 `StandardPipeline`입니다. 여러 `Valve`를 연결하여 요청 처리 흐름을 정의합니다.
 *           │   └── StandardEngineValve              <-- 엔진 레벨에서의 기본 `Valve`입니다. 요청을 적절한 `Host`로 전달하는 역할을 합니다.
 *           └── StandardHost(added as child)         <-- 도메인 이름을 기반으로 컨텍스트(Context)를 선택
 *               ├── StandardPipeline                 <-- 호스트 레벨에서의 `StandardPipeline`입니다.
 *               │   └── StandardHostValve            <-- 호스트 레벨에서의 기본 `Valve`입니다. 요청을 적절한 `Context`로 전달하는 역할을 합니다.
 *               └── TomcatEmbeddedContext            <-- 하나의 웹 애플리케이션을 나타내며, 특정 경로에 매핑된 애플리케이션을 관리합니다.
 *                   ├── StandardPipeline             <-- 컨텍스트 레벨에서의 `StandardPipeline`입니다.
 *                   │   └── StandardContextValve     <-- 컨텍스트 레벨에서의 기본 `Valve`입니다. 요청을 적절한 `Wrapper`로 전달하는 역할을 합니다.
 *                   └── StandardWrapper(s)           <-- 특정 서블릿을 감싸고 서블릿 인스턴스화 및 호출 관리
 *                       └── StandardPipeline         <-- Wrapper 레벨에서의 `StandardPipeline`입니다.
 *                           └── StandardWrapperValve <-- Wrapper 레벨에서의 기본 `Valve`입니다. 실제 서블릿의 `service` 메서드를 호출하여 요청을 처리합니다.
 *
 * ```
 *
 * - [org.apache.catalina.core.ContainerBase]를 상속하는 클래스들은 모두 각각 [org.apache.catalina.core.StandardPipeline]을 갖고 있으며,
 *   해당 파이프라인과 관련된 [org.apache.catalina.Valve]를 갖고 있습니다.
 *   이 각각의 파이프라인과 밸브를 통해 요청을 적절한 서블릿으로 전달할 수 있습니다.
 *
 * - [org.apache.catalina.Valve] 구현체 이름은 `Standard{Engine|Host|Context|Wrapper}Valve` 패턴을 따릅니다.
 *
 * - 가장 깊은 단계인 [org.apache.catalina.core.StandardWrapperValve.invoke]에서 요청에 대해 필터 체인을 호출하는데,
 *   이때 서블릿의 `service` 메서드도 같이 호출됩니다.
 *    ```
 *    // Call the filter chain for this request
 *    // NOTE: This also calls the servlet's service() method
 *    Container container = this.container;
 *    ```
 */
fun main() {
    val app = SpringApplication(SpringMain::class.java)
    val ctx = app.run()

    // ClassLoader
    println("ctx.classLoader: ${ctx.classLoader.javaClass.canonicalName}") // ctx.classLoader: jdk.internal.loader.ClassLoaders.AppClassLoader
    println("ctx.beanFactory.beanClassLoader: ${ctx.beanFactory.beanClassLoader.javaClass.canonicalName}") // ctx.beanFactory.beanClassLoader: jdk.internal.loader.ClassLoaders.AppClassLoader

    // Property
    val binder = Binder.get(ctx.environment)
    // 직접 바인딩
    val myProp = MyTestProperty()
    binder.bind(MyTestProperty.KEY, Bindable.ofInstance(myProp))
    // @ConfigurationProperties 어노테이션 통한 바인딩
    val myConfigurationProperty = ctx.getBean(MyCustomProperty::class.java)
    println("myProp: $myProp") // myProp: string: My Test String Property, bool: true
    println("myConfigurationProperty: $myConfigurationProperty") // myConfigurationProperty: clientId: client-id-1234-5678, isAdmin: true, applicationType: WEB

    val springMainBean = ctx.beanFactory.getBeanDefinition("springMain")
    println(
        "beanClassName: ${springMainBean.beanClassName}\n" + // beanClassName: spring.SpringMain$$SpringCGLIB$$0
                "description: ${springMainBean.description}\n" + // description: Spring Main Class to Learn Spring
                "scope: ${springMainBean.scope}" // scope: singleton
    )
}