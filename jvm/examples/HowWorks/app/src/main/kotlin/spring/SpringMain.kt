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
 * [부트스트랩 컨텍스트(Bootstrap Context)](https://docs.spring.io/spring-cloud-commons/reference/spring-cloud-commons/application-context-services.html#the-bootstrap-application-context)는 스프링 부트 2.4 버전부터 도입된 기능입니다.
 * 다양한 소스에서 설정을 가져와야 하는 복잡한 구성 환경에서 구성 데이터를 일관성 있게 관리하기 위해 도입되었다고 합니다.
 * 애플리케이션 컨텍스트가 생성되기 전에 외부 소스(예: 구성 서버, 클라우드 서비스)에서 필요한 설정이나 리소스를 가져와서 초기화하고 로드하는 데 사용됩니다.
 * 예를 들어,
 * - [Spring Cloud Config - Client Side Usage](https://docs.spring.io/spring-cloud-config/docs/current/reference/html/#_client_side_usage)처럼 같은 외부 서버로부터 설정을 가져올 수 있습니다.
 * - 애플리케이션이 시작될 때 AWS Secrets Manager나 HashiCorp Vault와 같은 secret 관리 서비스로부터
 *   데이터베이스 자격 증명이나 API 키와 같은 민감하지만 필요한 정보를 미리 로드할 수 있습니다.
 *    - [SpringBoot & AWS Secrets Manager](https://repost.aws/articles/ARrbXsydIkSAqKLrWhos7GnQ)
 *    - [spring-cloud-vault](https://github.com/spring-cloud/spring-cloud-vault?tab=readme-ov-file#client-side-usage)
 *
 * [SpringApplication.createBootstrapContext] 메서드는:
 * 1. 애플리케이션 컨텍스트가 생성되기 전에 필요한 리소스나 설정을 등록하고 관리하는 컨테이너인 [DefaultBootstrapContext] 인스턴스를 생성합니다.
 * 2. 부트스트랩 컨텍스를 생성할 때 `META-INF/spring.factories`([org.springframework.core.io.support.SpringFactoriesLoader.FACTORIES_RESOURCE_LOCATION]) 파일을 통해
 *   [BootstrapRegistryInitializer] 타입의 인스턴스들을 동적으로 로드하여 초기화를 수행합니다.
 *
 * [DefaultBootstrapContext]는 [BootstrapRegistry]를 상속하는데, 이 레지스트리는 애플리케이션 컨텍스트 생성 전에 필요한 리소스나 설정을 등록할 수 있는 인터페이스입니다.
 * 그리고 [BootstrapRegistryInitializer]를 구현하면 커스텀 초기화 로직을 작성할 수 있습니다.
 *
 * ### 주요 이벤트를 감지하고 처리하기 위한 `RunListener` 얻기
 *
 * > [SpringApplication.getRunListeners]
 *
 * [SpringApplicationRunListener] 인터페이스를 구현한 클래스들을 로드합니다.
 * 이 [SpringApplicationRunListener] 구현체들은 [스프링 부트 애플리케이션의 실행 단계에서 발생하는 이벤트를 처리하기 위해 사용](https://docs.spring.io/spring-boot/reference/features/spring-application.html#features.spring-application.application-events-and-listeners)됩니다.
 * 별다른 설정이 없다면 [org.springframework.boot.context.event.EventPublishingRunListener]가 로드됩니다.
 *
 * [org.springframework.boot.context.event.EventPublishingRunListener.initialMulticaster] 인스턴스([org.springframework.context.event.SimpleApplicationEventMulticaster.multicastEvent])를 사용하여
 * 이벤트를 전파합니다.
 * 가령 애플리케이션의 실행 과정에서 특정 이벤트들(예: 초기화, 환경 설정, 컨텍스트 준비 등)을 전파하고 특정 이벤트마다 커스텀 로직을 실행할 수 있게 해줍니다.
 * 전파되는 이벤트들의 이름은 [SpringApplicationRunListeners]에서, 타입은 [org.springframework.boot.context.event.EventPublishingRunListener]에서 확인할 수 있습니다.
 *
 * 그리고 [SpringApplicationRunListeners] 인스턴스를 생성하여 반환합니다.
 *
 * ### 환경 준비하기
 *
 * > [SpringApplication.prepareEnvironment]
 *
 * 스프링 부트는 다양한 환경에서 유연하게 동작할 수 있도록 [애플리케이션의 설정을 외부화](https://docs.spring.io/spring-boot/reference/features/external-config.html)합니다.
 *
 * 우선 [org.springframework.boot.ApplicationContextFactory.DEFAULT] 즉, [DefaultApplicationContextFactory.createEnvironment]를 사용해서
 * [org.springframework.core.env.ConfigurableEnvironment] 를 생성합니다.
 * 이때 `META-INF/spring.factories` 등록된 [ApplicationContextFactory] 구현체중, 주어진 [WebApplicationType]를 지원하는 팩토리를 사용합니다.
 * - [org.springframework.boot.web.servlet.context.ServletWebServerApplicationContextFactory.createEnvironment]
 * - [org.springframework.boot.web.reactive.context.ReactiveWebServerApplicationContextFactory.createEnvironment]
 *
 * 환경 준비가 끝나면 "spring.boot.application.environment-prepared"([org.springframework.boot.context.event.ApplicationEnvironmentPreparedEvent]) 이벤트를 전파합니다.
 * 정확히는 모든 환경이 준비가 된 건 아니지만, 이 이벤트를 전파함으로서 환경 후 처리기 리스너([org.springframework.boot.env.EnvironmentPostProcessorApplicationListener])에서
 * 나머지 구성을 마무리하도록 합니다.
 * 이 리스너는 [`spring-boot/src/main/resources/META-INF/spring.factories`](https://github.com/spring-projects/spring-boot/blob/91778e9f96fa0ab561b43705c67a7236c4cbafe9/spring-boot-project/spring-boot/src/main/resources/META-INF/spring.factories#L50)에 정의되어 있습니다.
 *
 * [org.springframework.boot.env.EnvironmentPostProcessorApplicationListener.onApplicationEnvironmentPreparedEvent]에서 환경이 준비됐다는 이벤트를 수신하면,
 * 실질적인 후처리는 [org.springframework.boot.context.config.ConfigDataEnvironmentPostProcessor.postProcessEnvironment]에서 담당합니다.
 *
 * [org.springframework.boot.context.config.ConfigDataEnvironment.processAndApply]에서 `application.yaml` 등 구성 정보를 로드하여 환경 인스턴스에 적용합니다.
 * 여기서 다음과 같은 단계들을 거치게 됩니다.
 * - [org.springframework.boot.context.config.ConfigDataEnvironment.processInitial]:
 *   Spring 프로파일이나 특정 컨텍스트(Activation Context)가 결정되기 전, 기본적인 구성 데이터를 로드합니다.
 *
 * - [org.springframework.boot.context.config.ConfigDataEnvironment.processWithoutProfiles]:
 *   Spring 프로파일이 지정되지 않은 설정값을 적용합니다.
 *
 * - [org.springframework.boot.context.config.ConfigDataEnvironment.withProfiles]:
 *   활성화된 프로파일에 따라 데이터를 필터링하거나 추가 로드합니다.
 *
 * - [org.springframework.boot.context.config.ConfigDataEnvironment.processWithProfiles]:
 *   활성화된 프로파일에 따른 구성 데이터를 로드 및 병합합니다.
 *
 * - [org.springframework.boot.context.config.ConfigDataEnvironment.applyToEnvironment]:
 *   병합된 데이터를 [org.springframework.boot.web.servlet.context.ApplicationServletEnvironment]에 반영합니다.
 *
 * 이 과정에서 [org.springframework.boot.context.config.ConfigDataEnvironmentContributors.withProcessedImports],
 * [org.springframework.boot.context.config.ConfigDataImporter.resolve],
 * [org.springframework.boot.context.config.ConfigDataLocationResolvers.resolve] 등이 참여합니다.
 *
 * 이때 구성 데이터 경로 리졸버(`ConfigDataLocationResolver`) 목록은 [`spring-boot/src/main/resources/META-INF/spring.factories`](https://github.com/spring-projects/spring-boot/blob/91778e9f96fa0ab561b43705c67a7236c4cbafe9/spring-boot-project/spring-boot/src/main/resources/META-INF/spring.factories#L12-L15)에
 * 설정되어 있습니다.
 * `spring.factories` 통해 로드되는 경로 리졸버는 다음과 같습니다:
 * - [org.springframework.boot.context.config.ConfigTreeConfigDataLocationResolver]: [디렉토리 트리에 파일명은 키, 내용은 값](https://docs.spring.io/spring-boot/reference/features/external-config.html#features.external-config.files.configtree)이 되는 경우
 *    ```
 *    etc/
 *       config/
 *         myapp/
 *           username
 *           password
 *    ```
 * - [org.springframework.boot.context.config.StandardConfigDataLocationResolver]: `application.{yaml,yml,xml,properties}` 파일
 *
 * 구성 데이터를 로드하는 방법은 비슷하므로, [org.springframework.boot.context.config.ConfigDataEnvironment.processInitial] 위주로
 * 정리를 해보면 다음과 같습니다.
 *
 * 우선 구체적으로 어떤 경로에 어떤 구성 데이터가 위치하는지 알아야 합니다.
 * [org.springframework.boot.context.config.ConfigDataImporter.resolve] 호출 시에는 `List<ConfigDataLocation> locations` 경로 목록이 전달되는데,
 * 이 경로들은 [org.springframework.boot.context.config.ConfigDataEnvironment]에서 관리합니다.
 *
 * [org.springframework.boot.context.config.ConfigDataEnvironment.LOCATION_PROPERTY] 속성으로
 * [임의 경로를 지정](https://docs.spring.io/spring-boot/reference/features/external-config.html#features.external-config.files)할 수도 있습니다.
 * ```
 * $ java -jar myproject.jar --spring.config.location=\
 * 	optional:classpath:/default.properties,\
 * 	optional:classpath:/override.properties
 * ```
 *
 * 하지만 임의 경로를 지정하지 않는다면, [org.springframework.boot.context.config.ConfigDataLocation] 배열로 이뤄진
 * [org.springframework.boot.context.config.ConfigDataEnvironment.DEFAULT_SEARCH_LOCATIONS]을 기본값으로 사용합니다.
 * 이 [org.springframework.boot.context.config.ConfigDataLocation]는 설정 파일 위치를 탐색하기 위한 경로들입니다.
 * [org.springframework.boot.context.config.ConfigDataLocation.value]에 다음과 같이 `;`로 구분되는 문자열이 설정됩니다.
 * - 클래스패스 경우
 *    ```
 *    optional:classpath:/;optional:classpath:/config/
 *    ```
 *    - `classpath:/`: 클래스패스 루트에서 설정 파일을 찾습니다.
 *
 *       만약 빌드된 애플리케이션 경우 `resources` 디렉토리가 클래스패스에 포함됩니다.
 *       - `resources/application.yaml`은 `classpath:/application.yaml`에 있다고 판단하고,
 *       - `resources/config/application.yaml`은 `classpath:/config/application.yaml`에 있다고 판단하게 됩니다.
 *
 *    - `classpath:/config`: 클래스패스 내 `config` 폴더에서 설정 파일을 찾습니다.
 *
 * - 파일 시스템 경로 경우
 *    ```
 *    # `{glob}` 부분은 원래 `*`인데 주석으로 잘못 인식돼서 `{glob}`으로 대체
 *    optional:file:./;optional:file:./config/;optional:file:./config/{glob}/
 *    ```
 *    - `file:./`: 현재 작업 디렉토리(`./`)에서 설정 파일을 찾습니다.
 *
 *        여기서 `file`이라 함은 외부 파일 시스템 경로로, 현재 테스트중인 로컬 머신의 작업 디렉토리(working directory)를 의미합니다.
 *        이는 실행 시점의 [`user.dir` 시스템 속성](https://stackoverflow.com/a/16239152)으로 결정됩니다.
 *
 *        - IDE 경우 루트 프로젝트입니다.
 *           ```
 *           # IDE가 실행하는 명령어
 *           /path/to/bin/java \
 *             -javaagent:/path/to/intellij.app/Contents/lib/idea_rt.jar=61795:/path/to/intellij.app/bin \
 *             -Dfile.encoding=UTF-8 \
 *             -Dsun.stdout.encoding=UTF-8 \
 *             -Dsun.stderr.encoding=UTF-8 \
 *             -classpath <무수히 많은 jar 파일> \
 *             spring.SpringMainKt
 *
 *           ... 생략 ...
 *
 *           /path/to/project/app
 *           ```
 *
 *        - gradle 명령어로 실행 경우 명령어를 실행한 디렉토리입니다.
 *           ```
 *           cd /path/to/project
 *           ./gradlew app:bootRun
 *
 *           ... 생략 ...
 *
 *           /path/to/project/app
 *           ```
 *
 *        - jar 경우 jar 파일을 실행한 디렉토리입니다.
 *           ```
 *           ./gradlew app:bootJar
 *           cd app/build/libs
 *           java -jar app.jar
 *
 *           ... 생략 ...
 *
 *           /path/to/project/app/build/lib
 *           ```
 *
 *    - `optional:file:./config/`: 현재 작업 디렉터리의 `config` 디렉토리(`./config/`)에서 설정 파일을 찾습니다.
 *    - `optional:file:./config/{glob}/`: 현재 작업 디렉터리의 `config` 디렉토리 하위 모든 디렉토리(`./config/{glob}/`)에서 설정 파일을 찾습니다.
 *
 * 보통은 `application.yaml` 등을 사용할 텐데, 이 경우 [org.springframework.boot.context.config.StandardConfigDataLocationResolver]를
 * 사용하여 환경 설정 정보를 로드하게 됩니다.
 *
 * [org.springframework.boot.context.config.StandardConfigDataLocationResolver.resolve] 과정에서 구성이 위치하는 경우를 찾기 위해
 * [org.springframework.boot.context.config.StandardConfigDataLocationResolver.getReferencesForConfigName]를 호출합니다.
 * 이때 `spring.factories` 통해 이미 로드된 [org.springframework.boot.context.config.StandardConfigDataLocationResolver.propertySourceLoaders]를
 * 사용하여 구성 파일 확장자별로 로드할 준비를 합니다.
 * - [org.springframework.boot.env.YamlPropertySourceLoader]: `yaml`, `yml` 파일
 * - [org.springframework.boot.env.PropertiesPropertySourceLoader]: `properties`, `xml` 파일
 *
 * 각 프로퍼티 소스 로더가 구현한 [org.springframework.boot.env.PropertySourceLoader.getFileExtensions]을
 * 사용하여 다음과 같이 가능한 경로들을 만듭니다.
 * ```
 * // 파일 시스템 경우
 * "file:./application.yaml"
 * "file:./application.yml"
 * "file:./application.xml"
 * "file:./application.properties"
 * "file:./config/application.yaml"
 * "file:./config/application.yml"
 * "file:./config/application.xml"
 * "file:./config/application.properties"
 * "file:./config/{glob}/application.yaml"
 * "file:./config/{glob}/application.yml"
 * "file:./config/{glob}/application.xml"
 * "file:./config/{glob}/application.properties"
 *
 * // classpath 경우
 * "classpath:/application.yaml"
 * "classpath:/application.yml"
 * "classpath:/application.xml"
 * "classpath:/application.properties"
 * "classpath:/config/application.yaml"
 * "classpath:/config/application.yml"
 * "classpath:/config/application.xml"
 * "classpath:/config/application.properties"
 * ```
 *
 * 근데 IDE 통해 실행할 경우 보통 프로젝트 루트나 `config` 디렉토리에 `application.{yaml,yml,properties,xml}` 파일을 두는 경우는 없습니다.
 * 다만 IntelliJ 경우 [SpringMain] 실행시 아래와 같은 경로들이 클래스패스에 자동으로 추가됩니다.
 * ```
 * -classpath\
 *   /path/to/project/app/build/classes/java/main:\
 *   /path/to/project/app/build/classes/kotlin/main:\
 *   /path/to/project/app/build/resources/main
 * ```
 *
 * 그래서 로컬 머신에서 테스트 시 다음 경로를 클래스패스로 인식하고 `yaml` 파일을 로드합니다.
 * ```
 * file:/path/to/project/app/build/resources/main/application.yaml
 * ```
 *
 * 리졸브된 후보 리소스들이 준비되었다면, [org.springframework.boot.context.config.ConfigDataImporter.resolveAndLoad]에서
 * [org.springframework.boot.context.config.ConfigDataImporter.load]가 실행됩니다.
 *
 * 그리고 리졸브된 구성에 따라 데이터를 로드하는 로더들이 상이한데, [org.springframework.boot.context.config.ConfigDataLoaders]는 이런 로더들을 관리합니다.
 * - [org.springframework.boot.context.config.StandardConfigDataLoader]
 * - [org.springframework.boot.context.config.ConfigTreeConfigDataLoader]
 *
 * 그리고 [org.springframework.boot.context.config.ConfigDataLoaders.load] 메서드가 실행되면서 적절한 로더를 판단하고 로드를 시도합니다.
 *
 * 현재 테스트중 `yaml` 파일을 로드하므로 [org.springframework.boot.context.config.StandardConfigDataLoader.load]에서
 * [org.springframework.boot.env.YamlPropertySourceLoader.load]가 실행됩니다.
 * 더 구체적으로 본다면 다음과 같습니다.
 * 1. [org.springframework.beans.factory.config.YamlProcessor.createYaml]에서 `yaml` 구성을 나타낼 [org.yaml.snakeyaml.Yaml] 인스턴스를 생성하고,
 * 2. [org.springframework.beans.factory.config.YamlProcessor.process]에서
 *   [org.springframework.core.io.ClassPathResource.getInputStream]로 해당 파일에 대한 스트림을 가져온 후
 *   [org.yaml.snakeyaml.Yaml.loadAll]이 실행되면,
 * 3. [org.yaml.snakeyaml.constructor.BaseConstructor.getData]에서 `yaml` 파일의 데이터를 가져옵니다.
 *   이때 순회하면서 하나하나 가져오는 게 아니라, `yaml` 파일을 통째로 로드합니다.
 *   그래서 사실 왜 for 문으로 스트림을 여러번 반복하는지 모르겠지만, 어쨌든 이 단계에서 `yaml` 파일의 내용은 [org.yaml.snakeyaml.Yaml] 인스턴스를 채우게 됩니다.
 * 4. [org.springframework.boot.env.OriginTrackedMapPropertySource]를 생성하고,
 *   [org.springframework.core.env.PropertySource] 목록에 추가하여 리턴합니다.
 * 5. [org.springframework.core.env.PropertySource] 목록을 [org.springframework.boot.context.config.ConfigData] 인스턴스로 감싸서 리턴합니다.
 *
 * 그럼 다시 [org.springframework.boot.context.config.ConfigDataEnvironment.processAndApply] 단계로 돌아가보면,
 * [org.springframework.boot.context.config.ConfigDataEnvironment.processInitial] 메서드를 통해
 * Spring 프로파일이나 특정 컨텍스트(Activation Context)가 결정되기 전, 기본적인 구성 데이터를 로드가 완료됩니다.
 *
 * 이후에는 프로파일 없는 데이터 처리, 프로파일 활성화 처리, 프로파일에 따른 구성 데이터 로드 및 병합 등을 처리하고
 * 최종적으로 [org.springframework.boot.context.config.ConfigDataEnvironment.applyToEnvironment] 메서드를 통해
 * 현재 [org.springframework.boot.context.config.ConfigDataEnvironment.environment]에 반영합니다.
 *
 * [SpringApplication.prepareEnvironment]에서 구성 데이터가 모두 로드되었다면,
 * [SpringApplication.bindToSpringApplication] 메서드가 실행되어 로드된 구성 데이터가 [SpringApplication] 인스턴스의 속성에 바인딩합니다.
 * 예를 들어:
 * - `spring.main.web-application-type`은 [SpringApplication.webApplicationType]에 설정
 * - `spring.main.banner-mode`는 [SpringApplication.bannerMode]에 설정
 * - `spring.main.lazy-initialization`는 [SpringApplication.lazyInitialization]에 설정
 * - [그 외 공통 애플리케이션 속성 목록](https://docs.spring.io/spring-boot/appendix/application-properties/index.html) 참고
 *
 * 속성의 소스가 로드되는 순서는 [링크](https://docs.spring.io/spring-boot/reference/features/external-config.html)를 참고합니다.
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
 * 앞서 [SpringApplication.createApplicationContext] 통해 생성된 애플리케이션 컨텍스트를 실제로 사용하기 전에 필요한 준비 작업을 수행합니다.
 * - 환경 설정
 * - 초기화기(initializer) 적용
 * - 빈 팩토리 설정 등
 *
 * 애플리케이션이 실행될 때 환경 설정이 컨텍스트에 반영되도록 컨텍스트에 환경(ex:[org.springframework.boot.web.servlet.context.ApplicationServletEnvironment])
 * 인스턴스를 설정합니다.
 *
 * 초기화기(initializer)는 컨텍스트가 완전히 생성되기 전에 필요한 추가 설정이나 빈을 등록하기 위해 사용됩니다.
 * 초기화기(initializer) 적용은 [SpringApplication] 인스턴스 생성 시 `spring.factories` 통해서 로드된 컨텍스트 이니셜라이저들의 `initialize`를
 * 실행하는 것으로 이뤄집니다.
 * - [org.springframework.boot.autoconfigure.SharedMetadataReaderFactoryContextInitializer.initialize]
 * - [org.springframework.boot.context.config.DelegatingApplicationContextInitializer.initialize]
 * - [org.springframework.boot.context.ContextIdApplicationContextInitializer.initialize]
 * - [org.springframework.boot.autoconfigure.logging.ConditionEvaluationReportLoggingListener.initialize]
 * - [org.springframework.boot.context.ConfigurationWarningsApplicationContextInitializer.initialize]
 * - [org.springframework.boot.rsocket.context.RSocketPortInfoApplicationContextInitializer.initialize]
 * - [org.springframework.boot.web.context.ServerPortInfoApplicationContextInitializer.initialize]
 *
 * 초기화 후 애플리케이션 컨테스트가 준비되었다는 "spring.boot.application.context-prepared"([org.springframework.boot.context.event.ApplicationContextInitializedEvent])
 * 이벤트를 전파하고, 부트스트랩 컨테스트를 종료합니다.
 *
 * 그 다음 애플리케이션 컨텍스트의 빈 팩토리에 필요한 싱글톤 빈 등록, 빈 팩토리의 동작 방식 설정, 빈 팩토리 후 처리기(PostProcessor) 추가 등의 작업을 수행합니다.
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
 *
 * - 빈 팩토리가 [org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory]이면
 *   순환 참조 허용 여부(`spring.main.allow-circular-references`) 설정
 *
 * - 빈 팩토리가 [org.springframework.beans.factory.support.DefaultListableBeanFactory]이면
 *   빈 정의 재정의 허용 여부(`spring.main.allow-bean-definition-overriding`)를 설정
 *
 * 그 다음 컨텍스트에 대해:
 * - 지연 초기화 설정(`spring.main.lazy-initialization`)이면 [LazyInitializationBeanFactoryPostProcessor] 추가합니다.
 * - 애플리케이션 종료를 방지하기 위한 [org.springframework.boot.SpringApplication.KeepAlive] 리스너를 추가합니다.
 * - 프로퍼티 소스의 우선순위를 올바르게 적용하여 설정 충돌을 방지하기 위해 [org.springframework.boot.SpringApplication.PropertySourceOrderingBeanFactoryPostProcessor] 추가합니다.
 * - AOT(Ahead-of-Time) 컴파일된 아티팩트를 사용하지 않는 경우, 런타임에 소스 코드를 기반으로 빈 정의를 생성하고 애플리케이션 컨텍스트에 등록하기 위해
 *   애플리케이션의 소스들을 컨텍스트에 로드([SpringApplication.load])합니다.
 *   일반적으로 `@SpringBootApplication` 어노테이션이 붙은 메인 클래스(ex: [SpringMain])입니다.
 *
 * 애플리케이션의 소스들을 컨텍스트에 로드([SpringApplication.load])하는 과정을 통해 `@SpringBootApplication` 어노테이션이 붙은
 * 메인 클래스(ex: [SpringMain])를 기반으로 빈 정의 로더([BeanDefinitionLoader])를 생성하고 빈을 로드합니다:
 * - [SpringBootApplication] 어노테이션이 붙은 클래스는 스프링 부트 애플리케이션의 메인 구성([org.springframework.context.annotation.Configuration]) 클래스로서,
 *   자동 구성과 컴포넌트 스캔을 통해 애플리케이션의 빈 정의를 시작하는 출발점이 됩니다.
 *   그 자신도 빈으로 등록되며, [org.springframework.context.annotation.Primary], [org.springframework.context.annotation.Description] 등을 설정합니다.
 *
 * - [org.springframework.context.annotation.ComponentScan] 어노테이션을 포함하고 있어,
 *   지정된 패키지 및 그 하위 패키지에서 컴포넌트들을 스캔하고 빈으로 등록합니다.
 *
 * - [org.springframework.boot.autoconfigure.EnableAutoConfiguration] 어노테이션을 포함하고 있어,
 *   스프링 부트의 자동 구성을 활성화합니다.
 *
 * 구체적인 과정은 다음과 같습니다.
 * 1. [org.springframework.boot.BeanDefinitionLoader.load] 통해서 로드하고,
 * 2. [org.springframework.context.annotation.AnnotatedBeanDefinitionReader.register]가 실행되고,
 * 3. [org.springframework.context.annotation.AnnotatedBeanDefinitionReader.doRegisterBean]가 호출됩니다.
 *   여기서 주어진 빈 클래스는 [SpringBootApplication] 어노테이션이 붙은 "class spring.SpringMain" 클래스입니다.
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
 * 애플리케이션 컨텍스트를 리프레시함으로써 빈들을 초기화하고 설정을 적용합니다.
 * 이를 통해 애플리케이션이 실행되기 전에 모든 빈들이 올바르게 설정되고 준비되도록 합니다.
 *
 * [SpringApplication.refresh]는 [org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext.refresh]를 실행하고,
 * 서블릿 웹 앱인 경우 [org.springframework.context.support.AbstractApplicationContext.refresh]가 실행됩니다.
 *
 * - 리프레시 위해 컨텍스트 준비: [org.springframework.context.support.AbstractApplicationContext.prepareRefresh]
 *    - 시작 시간 설정
 *    - `active` 플래그에 `true` 값 설정하여 컨텍스트가 활성화되었음 기록
 *    - `ApplicationServletEnvironment`에서 서블릿 관련 프로퍼티 소스를 적용
 *    - `GenericWebApplicationContext.initPropertySources`
 *
 * - 빈 팩토리 준비: [org.springframework.context.support.AbstractApplicationContext.prepareBeanFactory].
 *    - 빈 팩토리에 스프링 표현식 언어(`SpEL`)를 처리할 수 있는 빈 표현 리졸버(`StandardBeanExpressionResolver`) 설정합니다.
 *
 *    - 프로퍼티 값을 적절한 타입으로 변환하기 위해 `Property` 에디터(`ResourceEditorRegistrar`) 추가합니다.
 *
 *    - `*Aware` 인터페이스 구현한 빈(`ApplicationContextAware`, `ResourceLoaderAware` 등)들이 애플리케이션 컨텍스트를 주입받을 수 있도록
 *      `ApplicationContextAwareProcessor`를 후처리기로 등록합니다.
 *      그리고 `ApplicationContextAwareProcessor`가 처리해주는 인터페이스들을 일반적인 의존성 주입 대상에서 제외합니다.
 *
 *    - 컨텍스트 자체를 의미하는 `Environment`, `ResourceLoader`, `ApplicationEventPublisher`, `ApplicationContext` 등의 빈은
 *      자동 주입 대상에서 제외하여 순환 참조를 예방합니다.
 *
 *    - `Environment`, `ResourceLoader`, `ApplicationEventPublisher`, `ApplicationContext` 타입의 빈을 요청할 때,
 *      현재 컨텍스트(ex: `AnnotationConfigServletWebServerApplicationContext`)를 제공하도록 빈 팩토리를 설정합니다.
 *
 *    - `ApplicationListener` 인터페이스를 구현한 빈을 감지하고, 애플리케이션 이벤트 멀티캐스터에 등록하거나 제거하는 `ApplicationListenerDetector`를
 *      Bean Post Processor로 추가합니다.
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
 *
 * 1. [org.apache.catalina.core.StandardService]
 * 2. [org.apache.catalina.core.StandardEngine]
 * 3. [org.apache.catalina.core.StandardPipeline]
 * 4. [org.apache.catalina.core.StandardEngineValve]
 * 5. [org.apache.catalina.core.StandardEngineValve.invoke]
 * 6. [org.apache.catalina.core.StandardHost]
 * 6. [org.apache.catalina.core.StandardHostValve]
 * 7. [org.apache.catalina.authenticator.NonLoginAuthenticator.invoke]
 * 7. [org.apache.catalina.core.StandardContextValve.invoke]
 * 8. [org.apache.coyote.http11.Http11Processor.action]
 * 9. [org.apache.catalina.core.StandardWrapper]
 * 10. [org.apache.catalina.core.StandardPipeline]
 * 11. [org.apache.catalina.core.StandardWrapperValve.invoke]
 * 12. [org.apache.catalina.core.StandardWrapper.initServlet]
 * 13. [org.springframework.web.servlet.DispatcherServlet.init]
 * 13. [org.springframework.web.servlet.FrameworkServlet.initServletBean]
 * 13. [org.springframework.web.servlet.FrameworkServlet.initWebApplicationContext]
 * 13. [org.springframework.web.servlet.DispatcherServlet.onRefresh]
 * 13. [org.springframework.web.servlet.DispatcherServlet.initStrategies]
 *
 *
 *
 * [org.springframework.web.context.support.WebApplicationContextUtils.registerWebApplicationScopes]
 * - Register web-specific scopes ("request", "session", "globalSession", "application") with the given BeanFactory, as used by the [org.springframework.web.context.WebApplicationContext].
 * - Spring에서 **스코프(scope)**는 빈(Bean)이 생성되고 관리되는 생명 주기를 정의하는 개념입니다. registerWebApplicationScopes 메서드는 **웹 애플리케이션의 주요 스코프(요청, 세션, 애플리케이션 등)**를 BeanFactory에 등록하여, 웹 환경에서 스코프별로 빈을 적절히 관리할 수 있도록 설정합니다.
 *   Spring에서 제공하는 기본 스코프:
 *   - 싱글톤(`singleton`): (기본값) 애플리케이션 전체에서 단일 인스턴스가 생성됩니다.
 *   - 프로토타입(`prototype`): 요청할 때마다 새로운 인스턴스가 생성됩니다.
 *   - 요청(`request`): HTTP 요청 당 하나의 인스턴스가 생성됩니다. (웹 환경)
 *   - 세션(`session`): HTTP 세션 당 하나의 인스턴스가 생성됩니다. (웹 환경)
 *   - 애플리케이션(`application`): 서블릿 컨텍스트 당 하나의 인스턴스가 생성됩니다. (웹 환경)
 *   - 웹 소켓(`websocket`): 웹소켓 세션 당 하나의 인스턴스가 생성됩니다. (웹 환경)
 *
 * /var/folders/p_/t39sykwx5kx03wh6sj5tf3840000gn/T/tomcat.8080.18326210342099403924
 * /private/var/folders/p_/t39sykwx5kx03wh6sj5tf3840000gn/T/tomcat.8080.18326210342099403924
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

    println(System.getProperty("user.dir"))
    // IDE 경우: /path/to/project
    // gradle 경우: /path/to/project/app
    // jar(app/build/libs 이동): /path/to/project/app/build/libs
}