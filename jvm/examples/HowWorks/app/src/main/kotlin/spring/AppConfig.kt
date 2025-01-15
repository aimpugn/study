package spring

import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.EnableAspectJAutoProxy

@Configuration
// https://docs.spring.io/spring-boot/reference/features/aop.html#page-title
// `AspectJ`가 클래스 패스에 있으면, 스프링 부트의 자동 구성은 AspectJ 자동 프록시를 활성화하므로,
// `@EnableAspectJAutoProxy`를 추가하지 않아도 됩니다.
// 단, 여기서는 구성 테스트 및 `exposeProxy` 설정을 추가하여 오버라이드하기 위해 직접 명시합니다.
@EnableAspectJAutoProxy(exposeProxy = true)
class AppConfig