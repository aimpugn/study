package spring.aop

import org.aspectj.lang.JoinPoint
import org.aspectj.lang.annotation.*
import org.springframework.aop.framework.Advised
import org.springframework.aop.framework.AopContext
import org.springframework.stereotype.Component

/**
 * Spring AOP는 프록시 기반 AOP를 사용하여 동작합니다.
 * 주로 비즈니스 로직의 메서드 진입점에서 로깅, 보안, 트랜잭션 관리 등을 처리하기 위한 경량 AOP를 목표로 하기 때문입니다.
 * 반면 AspectJ는 AOP를 구현하기 위해 위빙(Weaving) 과정을 수행합니다.
 * 위빙은 Aspect의 정의를 타겟 클래스와 결합하여 최종 바이트코드를 생성하는 과정입니다.
 *
 * ### 포인트컷
 *
 * 어떤 조건에서 [org.aopalliance.aop.Advice]를 실행할 것인지를 결정합니다.
 * 어떤 Join Point(메서드, 클래스, 어노테이션 등)가 AOP 로직에 의해 가로채질 것인지를 정의합니다.
 *
 * 포인트컷은 어떤 조건에서 실행될 것인지를 의미하므로, 어떤 타이밍에 실행할지 결정하는 어드바이스에 독립적입니다.
 * 즉, `execution(${조건})`을 설정하면, [After], [Before], 또는 [Around] 등
 * 어느 타이밍에 실행할지는 별도로 설정할 수 있습니다.
 *
 * ### 어드바이스
 *
 * 언제(At What Timing) 실행할지를 결정합니다.
 * Join Point(메서드, 클래스, 어노테이션 등) 전/후 어떤 작업을 실행할 것인지 정의합니다.
 *
 * 스프링에서 각 [org.aopalliance.aop.Advice]는 스프링 빈입니다.
 * 어드바이스 인스턴스는 모든 [Advised] 객체에서 공유되거나, 각 [Advised] 오브젝트에 대해 고유할 수 있습니다.
 * 이것들은 프록시된 객체의 상태에 의존하지 않고 새로운 상태를 추가하지도 않습니다. 이들은 단지 메서드와 인수에 따라 행동합니다.
 *
 * References:
 * - [Declaring a Pointcut](https://docs.spring.io/spring-framework/reference/core/aop/ataspectj/pointcuts.html)
 * - [Pointcut API in Spring](https://docs.spring.io/spring-framework/reference/core/aop-api/pointcuts.html)
 * - [Advice API in Spring](https://docs.spring.io/spring-framework/reference/core/aop-api/advice.html)
 * - [The AspectJ Language](https://github.com/eclipse-aspectj/aspectj/blob/master/docs/progguide/language.adoc)
 * - [Intro to AspectJ](https://www.baeldung.com/aspectj)
 */
@Aspect
@Component
open class AopExample {

    /**
     * [spring.controller.HomeController.home] 메서드 실행에 매칭하는 [Pointcut].
     * `execution` 포인트컷은 메서드 실행 조인 포인트를 매칭합니다.
     *
     * ```
     * execution(modifiers-pattern?
     *          ret-type-pattern
     *          declaring-type-pattern?name-pattern(param-pattern)
     *          throws-pattern?)
     * ```
     *
     * `aspectjrt`는 aspectj runtime 줄임말로, 런타임에 프록시 통해서 실행합니다.
     * CGLIB 경우 상속을 통해서 프록시를 생성하는데, kotlin 경우 기본적으로 `public final` 클래스 또는 메서드입니다.
     *
     * 프록시 통해서 사용하기 위해서는 `open`으로 `final`을 제거해줘야 합니다. (ex: [spring.controller.HomeController.home])
     */
    @Pointcut("execution(* spring.controller.HomeController.home())")
    fun atHomeMethodExecution() {
        // `@Pointcut`은 AOP가 적용될 지점(조인포인트 조건)을 정의하기 위한 애노테이션입니다.
        // 따라서 본문에는 로직이 없어야 합니다.
    }

    @Before("atHomeMethodExecution()")
    fun loggingBeforeAtHomeMethodExecution() {
        println("Logging before spring.controller.HomeController.home() execution")
        printCurrentProxy()
    }

    fun printCurrentProxy() {
        val currentProxy = AopContext.currentProxy()
        println("\tcurrentProxy: $currentProxy")
        println("\tcurrentProxy::class.java.name: ${currentProxy::class.java.name}")
        println("\tcurrentProxy is Advised: ${currentProxy is Advised}")

        val advised = currentProxy as Advised
        println("\tcurrentProxy's target: ${advised.targetSource.target.javaClass}")
        println("\tcurrentProxy's advisorCount: ${advised.advisorCount}")
        advised.advisors.forEachIndexed { idx, advisor ->
            println("\t\t${idx + 1}. advisor")
            println("\t\t  - AspectJPointcutAdvisor: ${advisor.javaClass}") // 대상 메서드 하나마다 Advisor 인스턴스 하나 생성
            println("\t\t  - Advisor detail: $advisor")
        }
        // Output:
        //  currentProxy: spring.controller.HomeController@21c75084
        //  currentProxy::class.java.name: spring.controller.HomeController$$SpringCGLIB$$0
        //  currentProxy is Advised: true
        //  currentProxy's target: class spring.controller.HomeController
        //  currentProxy's advisorCount: 4
        //      1. advisor
        //        - AspectJPointcutAdvisorImpl: class org.springframework.aop.interceptor.ExposeInvocationInterceptor$1
        //        - Advisor detail: org.springframework.aop.interceptor.ExposeInvocationInterceptor.ADVISOR
        //      2. advisor
        //        - AspectJPointcutAdvisorImpl: class org.springframework.aop.aspectj.annotation.InstantiationModelAwarePointcutAdvisorImpl
        //        - Advisor detail: InstantiationModelAwarePointcutAdvisor: expression [atHomeMethodExecution()]; advice method [public final void spring.aop.LoggingAspect.logBefore()]; perClauseKind=SINGLETON
        //                                  ... 생략 ...
        // ```
        // spring.controller.HomeController$$SpringCGLIB$$0
        // - HomeController: 원래 실행 대상인 클래스
        // - $$SpringCGLIB$$0: Spring이 `CGLIB`을 사용하여 만든 하위 클래스
        // ```
        //
        // Spring이 `CGLIB`을 사용하여 만든 프록시 객체임을 의미합니다.
        // Spring 컨테이너가 AOP 적용을 위해 `HomeController`를 감싸기 위해 `CGLIB`을 사용했다는 의미이며,
        // `HomeController` 자체는 원래 `CGLIB` 프록시가 아니라 단순한 스프링 빈입니다.
    }

    /**
     * ```
     * @annotation(SomeAnnotation)
     * ```
     */
    @Before("@annotation(spring.aop.LoggingBefore)")
    fun loggingBeforeAtLoggingBeforeAnnotation(joinPoint: JoinPoint) {
        val methodName = joinPoint.signature.name
        val className = joinPoint.signature.declaringTypeName
        val arguments = joinPoint.args

        println("Logging before for LoggingBefore annotation")
        printCurrentProxy()
        println("\t$className.$methodName(${arguments.joinToString(", ")})")
        // Output:
        // Logging before for LoggingBefore annotation
        //    spring.controller.HomeController.printName(user)
        // Logging before for LoggingBefore annotation
        //    spring.controller.HomeController.printName(favicon.ico)
    }

    /**
     * ```
     * @annotation(SomeAnnotation)
     * ```
     */
    @Before("@annotation(spring.aop.LoggingBefore) && args(name)")
    fun loggingBeforeAtLoggingBeforeAnnotationWithNameArg(name: String) {
        println("Logging before for LoggingBefore annotation with `name` arg")
        printCurrentProxy()
        println("\tname: $name")
        // Output:
        // Logging before for LoggingBefore annotation with `name` arg
        //    name: user
        // Logging before for LoggingBefore annotation with `name` arg
        //    name: favicon.ico
    }
}