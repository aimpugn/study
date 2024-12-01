package spring

import org.springframework.boot.autoconfigure.EnableAutoConfiguration
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

/**
 * `open`이어야 합니다.
 * ```
 * Configuration problem: @Configuration class 'MySpringConfig' may not be final. Remove the final modifier to continue.
 * ```
 *
 * [EnableAutoConfiguration]는 스프링 부트가 클래스패스에서 발견된 의존성을 기반으로 자동으로 설정을 구성하도록 합니다.
 * 이 어노테이션은 [SpringBootApplication] 어노테이션에 이미 포함되어 있으므로 추가할 필요 없습니다.
 * 단, 테스트시(`@SpringBootTest`)에 필요하다면 명시적으로 선언해야 할 수도 있습니다.
 */
@Configuration
open class MySpringConfig {

    /**
     * `open`이어야 합니다.
     * ```
     * Configuration problem: @Bean method 'myConfigurationProperty' must not be private or final; change the method's modifiers to continue.
     * ```
     */
    @Bean
    open fun myConfigurationProperty(): MyCustomProperty = MyCustomProperty()
}