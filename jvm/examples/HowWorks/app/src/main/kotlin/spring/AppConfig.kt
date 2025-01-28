package spring

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.EnableAspectJAutoProxy
import org.springframework.core.task.TaskExecutor
import org.springframework.data.jpa.repository.config.EnableJpaAuditing
import org.springframework.scheduling.annotation.EnableAsync
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor
import java.util.concurrent.ThreadFactory


@Configuration
// https://docs.spring.io/spring-boot/reference/features/aop.html#page-title
// `AspectJ`가 클래스 패스에 있으면, 스프링 부트의 자동 구성은 AspectJ 자동 프록시를 활성화하므로,
// `@EnableAspectJAutoProxy`를 추가하지 않아도 됩니다.
// 단, 여기서는 구성 테스트 및 `exposeProxy` 설정을 추가하여 오버라이드하기 위해 직접 명시합니다.
@EnableAspectJAutoProxy(exposeProxy = true)
@EnableJpaAuditing
@EnableAsync
class AppConfig {
    /**
     * References:
     * - [Using a TaskExecutor](https://docs.spring.io/spring-framework/reference/integration/scheduling.html#scheduling-task-executor-usage)
     */
    @Bean(name = ["customTaskExecutor"])
    fun customTaskExecutor(): TaskExecutor {
        val executor = ThreadPoolTaskExecutor().apply {
            corePoolSize = 5
            maxPoolSize = 10
            queueCapacity = 25
            setThreadNamePrefix("CustomExecutor-")
            setThreadFactory(ThreadFactory { r: Runnable? ->
                val thread = Thread(r)
                thread.setName("Thread-by-CustomExecutor-" + thread.threadId())
                thread.setDaemon(false)
                thread
            })
        }

        executor.initialize()

        return TrackingCustomTaskExecutor(executor)
    }
}