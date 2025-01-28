package spring.service

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import org.springframework.core.task.SimpleAsyncTaskExecutor
import org.springframework.core.task.TaskExecutor
import org.springframework.scheduling.annotation.Async
import org.springframework.scheduling.concurrent.ConcurrentTaskExecutor
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor
import org.springframework.stereotype.Service
import spring.TrackingExecutorTracker
import java.time.Instant
import java.util.concurrent.*
import java.util.stream.LongStream
import kotlin.random.Random
import kotlin.system.measureTimeMillis
import kotlin.time.DurationUnit
import kotlin.time.toDuration
import kotlin.time.toJavaDuration

/**
 * 동기적인 작업은 A, B, C 세 작업이 있다면 A, B, C 순서대로 처리됩니다.
 * 비동기란 결국, 작업 A가 완료되기를 기다리지 않고, 작업 B, C 작업들을 병렬로 진행할 수 있는 것을 의미합니다.
 *
 * 만약 A 작업이 I/O로 인해서 대기 시간이 길다면, 그 시간만큼 CPU 리소스를 낭비하게 됩니다.
 * 비동기를 통해 CPU가 쉬지 않고 대기 상태의 스레드 대신 다른 작업을 실행하도록 함으로써 CPU 리소스를를 최정화할 수 있습니다.
 * 또한 블로킹 없이 다음 작업을 처리하므로 응답성이 높아집니다.
 *
 * [TaskExecutor], [CompletableFuture], [Async] 어노테이션, 코루틴 등을 테스트 해봅니다.
 */
@Service
class AsyncService(
    private val taskExecutor: TaskExecutor,
) {
    /**
     * [SimpleAsyncTaskExecutor]는 [TaskExecutor] 구현체입니다.
     *
     * [Executors]는 스레드 풀 개념에 대한 JDK 이름입니다.
     * 단, 실제로 스레드 풀 구현을 보장하지 않으므로 "executor"라고 합니다.
     *
     * [TaskExecutor]는 필요한 경우 다른 Spring 구성 요소에 스레드 풀링에 대한 추상화를 제공하기 위해 만들어졌습니다.
     * [org.springframework.context.event.SimpleApplicationEventMulticaster],
     * `AbstractMessageListenerContainer`, `Quartz` 등 모두 [TaskExecutor]를 사용합니다.
     * 이 인터페이스는 [Executor]와 동일합니다.
     *
     * [SimpleAsyncTaskExecutor]는 요청마다 새로운 스레드를 생성하여 비동기로 요청을 처리하지만,
     * 스레드를 재사용하지 않습니다.
     * 다만 슬롯이 비워질 때까지 한도를 넘는 모든 호출을 차단하는 동시성 한도를 지원합니다.
     *
     * 진짜 스레드 풀이 필요하다면 [ThreadPoolTaskExecutor]를 사용합니다.
     *
     * References:
     * - [TaskExecutor Types](https://docs.spring.io/spring-framework/reference/integration/scheduling.html#scheduling-task-executor-types)
     */
    fun simpleAsyncTaskExecutorExample(virtualTread: Boolean): Long {
        val sate = SimpleAsyncTaskExecutor("Example-SimpleAsyncTaskExecutor-").apply {
            concurrencyLimit = 3
            setVirtualThreads(virtualTread)
        }

        // 모든 비동기 작업이 완료될 때까지 기다리려면, `CountDownLatch` 또는 `CompletableFuture`를 사용합니다.
        return measureTimeMillis {
            val completableFutures = List(5) {
                CompletableFuture.runAsync({
                    val currThread = Thread.currentThread()
                    val sleepDuration = Random.nextLong(1000, 3000)
                        .toDuration(DurationUnit.MILLISECONDS)

                    println("${Instant.now()} [$currThread] 비동기 작업 실행 중... while $sleepDuration")
                    Thread.sleep(sleepDuration.toJavaDuration()) // 작업을 시뮬레이션
                    println("${Instant.now()} [$currThread] 비동기 작업 완료!")
                }, sate)
            }

            // 모든 비동기 작업이 완료될 때까지 메인 스레드를 블로킹하여 결과를 기다립니다.
            CompletableFuture.allOf(*completableFutures.toTypedArray()).join()
            // Output: OS 스레드 경우
            //  2025-01-28T08:21:52.113246Z [Thread[#60,Example-SimpleAsyncTaskExecutor-1,5,main]] 비동기 작업 실행 중... while 1.266s
            //  2025-01-28T08:21:52.113431Z [Thread[#62,Example-SimpleAsyncTaskExecutor-3,5,main]] 비동기 작업 실행 중... while 1.397s
            //  2025-01-28T08:21:52.113309Z [Thread[#61,Example-SimpleAsyncTaskExecutor-2,5,main]] 비동기 작업 실행 중... while 2.842s
            //  2025-01-28T08:21:53.384621Z [Thread[#60,Example-SimpleAsyncTaskExecutor-1,5,main]] 비동기 작업 완료!
            //  2025-01-28T08:21:53.385856Z [Thread[#63,Example-SimpleAsyncTaskExecutor-4,5,main]] 비동기 작업 실행 중... while 1.169s
            //  2025-01-28T08:21:53.515265Z [Thread[#62,Example-SimpleAsyncTaskExecutor-3,5,main]] 비동기 작업 완료!
            //  2025-01-28T08:21:53.516293Z [Thread[#64,Example-SimpleAsyncTaskExecutor-5,5,main]] 비동기 작업 실행 중... while 2.764s
            //  2025-01-28T08:21:54.558474Z [Thread[#63,Example-SimpleAsyncTaskExecutor-4,5,main]] 비동기 작업 완료!
            //  2025-01-28T08:21:54.960779Z [Thread[#61,Example-SimpleAsyncTaskExecutor-2,5,main]] 비동기 작업 완료!
            //  2025-01-28T08:21:56.282928Z [Thread[#64,Example-SimpleAsyncTaskExecutor-5,5,main]] 비동기 작업 완료!
            //
            // Output: Virtual 스레드 경우
            //  2025-01-28T08:21:36.984134Z [VirtualThread[#47,Example-SimpleAsyncTaskExecutor-1]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 실행 중... while 2.133s
            //  2025-01-28T08:21:36.984135Z [VirtualThread[#50,Example-SimpleAsyncTaskExecutor-3]/runnable@ForkJoinPool-1-worker-3] 비동기 작업 실행 중... while 1.762s
            //  2025-01-28T08:21:36.984134Z [VirtualThread[#49,Example-SimpleAsyncTaskExecutor-2]/runnable@ForkJoinPool-1-worker-2] 비동기 작업 실행 중... while 1.376s
            //  2025-01-28T08:21:38.366205Z [VirtualThread[#49,Example-SimpleAsyncTaskExecutor-2]/runnable@ForkJoinPool-1-worker-2] 비동기 작업 완료!
            //  2025-01-28T08:21:38.367809Z [VirtualThread[#58,Example-SimpleAsyncTaskExecutor-4]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 실행 중... while 2.342s
            //  2025-01-28T08:21:38.747579Z [VirtualThread[#50,Example-SimpleAsyncTaskExecutor-3]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 완료!
            //  2025-01-28T08:21:38.747989Z [VirtualThread[#59,Example-SimpleAsyncTaskExecutor-5]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 실행 중... while 2.061s
            //  2025-01-28T08:21:39.119604Z [VirtualThread[#47,Example-SimpleAsyncTaskExecutor-1]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 완료!
            //  2025-01-28T08:21:40.715607Z [VirtualThread[#58,Example-SimpleAsyncTaskExecutor-4]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 완료!
            //  2025-01-28T08:21:40.812779Z [VirtualThread[#59,Example-SimpleAsyncTaskExecutor-5]/runnable@ForkJoinPool-1-worker-1] 비동기 작업 완료!
        }
    }

    /**
     * [ConcurrentTaskExecutor]는 [java.util.concurrent.Executor]를
     * 스프링 [TaskExecutor]로 변환하는 어댑터입니다.
     *
     * References:
     * - [TaskExecutor Types](https://docs.spring.io/spring-framework/reference/integration/scheduling.html#scheduling-task-executor-types)
     */
    fun concurrentTaskExecutorExample(type: ExecutorType): Long {
        val executor: Executor = selectExecutor(type)
        println("Executor type: $type, Selected executor: $executor")
        val cte = ConcurrentTaskExecutor(executor)

        return measureTimeMillis {
            val completableFutures = List(5) {
                CompletableFuture.runAsync({
                    val currThreadName = Thread.currentThread().name
                    val sleepDuration = Random.nextLong(1000, 3000)
                        .toDuration(DurationUnit.MILLISECONDS)

                    println("${Instant.now()} [$currThreadName] 비동기 작업 실행 중... while $sleepDuration")
                    Thread.sleep(sleepDuration.toJavaDuration()) // 작업을 시뮬레이션
                    println("${Instant.now()} [$currThreadName] 비동기 작업 완료!")
                }, cte)
            }

            CompletableFuture.allOf(*completableFutures.toTypedArray()).join()
        }
    }

    private fun selectExecutor(type: ExecutorType): ExecutorService = when (type) {
        /**
         * 고정된 개수의 스레드를 재사용하는 스레드 풀을 생성합니다.
         * - 공유되는 무제한 크기의 큐(shared unbounded queue, [LinkedBlockingQueue])를 기반으로 동작합니다.
         * - 모든 스레드가 작업중일 때 새로운 작업이 추가되면, 스레드를 사용할 수 있을 때까지 큐에서 대기합니다.
         */
        ExecutorType.FIXED_THREAD_POOL -> Executors.newFixedThreadPool(5)
        /**
         * 작업 도난(work-stealing) 알고리즘을 사용하는 풀입니다.
         * - 각 스레드 별 Double-Ended Queue(Deque)([ForkJoinPool.WorkQueue])를 사용하여 작업 간의 충돌(contention)을 줄입니다.
         * - 실행 순서가 보장되지 않습니다.
         * - [ForkJoinPool]을 사용합니다.
         * - 병렬성 수준(parallelism level)은 작업 처리에 참여중이거나 참여할 수 있는 최대 스레드 개수를 의미합니다.
         * - 실제 스레드 개수는 동적으로 증가 또는 감소할 수 있습니다.
         *
         * > [작업 도난(Work Stealing)](https://en.wikipedia.org/wiki/Work_stealing)은 작업 처리 중 한 스레드가 자신에게 할당된 작업을 완료했지만,
         * 다른 스레드가 아직 작업을 처리 중인 경우, 여유 스레드가 다른 스레드의 작업을 가져와 처리하는 방식을 의미합니다.
         */
        ExecutorType.WORKER_STEALING_POOL -> Executors.newWorkStealingPool(5)
        /**
         * 단일 스레드가 작업을 순차적으로 처리하는 [ExecutorService]입니다.
         * - 작업은 FIFO 순서로 실행됩니다.
         * - 단일 스레드가 실패하면 새 스레드가 생성되어 작업을 계속 처리합니다.
         */
        ExecutorType.SINGLE_THREAD -> Executors.newSingleThreadExecutor()
        /**
         * 동적으로 스레드를 생성하거나 캐싱하는 스레드 풀입니다.
         * - 필요한 경우 새 스레드를 생성하며, 기존 스레드를 재사용합니다.
         * - 스레드의 keep alive 설정은 60초입니다.
         * - 사용하지 않을 때는 리소스를 소모하지 않습니다.
         * - [SynchronousQueue]를 사용하여 대기열을 최소화합니다.
         */
        ExecutorType.CACHED_THREAD_POOL -> Executors.newCachedThreadPool()
        /**
         * [java.lang.VirtualThread]를 사용하는 [ExecutorService]입니다.
         * - 작업마다 새로운 가상 스레드를 생성합니다.
         * - 가상 스레드는 경량화되어 자원 사용을 최소화하고, 많은 동시 작업을 처리할 수 있습니다.
         *
         * 이는 다음과 같습니다:
         * ```
         * Executors.newThreadPerTaskExecutor(가상_스레드_팩토리)
         * // 또는
         * ThreadPerTaskExecutor.create(가상_스레드_팩토리)
         * ```
         * - [ThreadPerTaskExecutor]: 각 작업마다 새로운 스레드를 생성합니다. 스레드의 개수는 무제한입니다.
         */
        ExecutorType.VIRTUAL_THREAD -> Executors.newVirtualThreadPerTaskExecutor()
        /**
         * 주기적으로 실행되거나 지연된 실행이 필요할 때 사용하는 [ExecutorService]입니다.
         * - [ScheduledThreadPoolExecutor]를 기반으로 동작하며, 주어진 간격 또는 일정에 따라 작업을 실행합니다.
         * - 코어 스레드 풀 크기([ScheduledThreadPoolExecutor.corePoolSize])는 일정한 스레드 수를 유지합니다.
         */
        ExecutorType.SCHEDULED_THREAD_POOL -> Executors.newScheduledThreadPool(5)
        /**
         *
         */
        ExecutorType.CUSTOM_THREAD_POOL -> {
            /**
             * 스레드 풀에 유지할 스레드 개수입니다.
             *
             * 단, [ThreadPoolExecutor.allowCoreThreadTimeOut]이 설정된다면,
             * 코어 스레드도 소멸될 수 있습니다.
             */
            val threadPoolSize = 5

            /**
             * 스레드 풀이 관리할 수 있는 최대 스레드 개수입니다.
             */
            val threadPoolMaxSize = 10

            /**
             * 스레드 풀에서 유휴 상태인 스레드가 작업을 기다릴 수 있는 시간입니다.
             */
            val keepAliveTime = 60L

            /**
             * `keepAliveTime`의 단위를 나타냅니다. 기본값은 나노초입니다.
             */
            val keepAliveTimeUnit = TimeUnit.SECONDS


            ThreadPoolExecutor(
                threadPoolSize,
                threadPoolMaxSize,
                keepAliveTime,
                keepAliveTimeUnit,
                LinkedBlockingQueue() // FIFO 방식으로 `Integer.MAX_VALUE` 크기의 대기열을 제공합니다.
            )
        }
    }

    /**
     * [Async] 어노테이션은 메서드 레벨에서 선언적으로 비동기 실행을 정의하는 것을 지원합니다.
     * Spring 컨텍스트에서 스레드 풀을 통합 관리하고 싶을 때 사용합니다.
     *
     * 값을 반환하는 메서드도 비동기로 호출될 수 있지만, `void` 또는 [Future] 타입의 값을 반환해야 합니다.
     *
     * [Async] 어노테이션은 [jakarta.annotation.PostConstruct] 같은 라이프사이클 콜백과 함께 사용할 수 없습니다.
     * 비동기로 스프링 빈을 초기화하려면, [Async] 어노테이션이 붙은 메서드를 호출하는 별도의 스프링 빈을 사용해야 합니다.
     * ```
     *  public class SampleBeanImpl implements SampleBean {
     *      @Async
     *      void doSomething() {
     *          // ...
     *      }
     *  }
     *
     *  public class SampleBeanInitializer {
     *
     *      private final SampleBean bean;
     *
     *      public SampleBeanInitializer(SampleBean bean) {
     *          this.bean = bean;
     *      }
     *
     *      @PostConstruct
     *      public void initialize() {
     *          bean.doSomething();
     *      }
     *  }
     * ```
     *
     * 또한 [Async.value]로 어떤 [Executor] 또는 [TaskExecutor]를 사용할지 지정할 수 있습니다.
     * [spring.AppConfig.customTaskExecutor]를 참고합니다.
     */
    @Async
    fun asyncComputeWithAsyncAnnotation(): CompletableFuture<Long> {
        // @Async는 Spring의 `TaskExecutor` 통해 새 스레드를 할당합니다.
        println(
            """
            [asyncComputeWithAsyncAnnotation]
            - Executor injected by Spring: $taskExecutor
            - ExecutorTracker currentExecutor: ${TrackingExecutorTracker.currentExecutor}
            - Thread inside @Async: ${Thread.currentThread()}
        """.trimIndent()
        )
        // Output:
        //  [asyncComputeWithAsyncAnnotation]
        //  - Executor injected by Spring: spring.TrackingCustomTaskExecutor@42e32feb
        //  - ExecutorTracker currentExecutor: spring.TrackingCustomTaskExecutor@42e32feb
        //  - Thread inside @Async: Thread[#47,Thread-by-CustomExecutor-47,5,main]
        return CompletableFuture.supplyAsync {
            measureTimeMillis {
                simulateCompute(1, 0)
                // Output:
                //  [simulateCompute]
                //  - Current thread: Thread[#48,ForkJoinPool.commonPool-worker-1,5,main]
                //  - start: 0
                //  - last: 10000000
                //  - result: 99999990000000
            }
        }
    }

    /**
     * 코루틴은 JVM 상에서 스레드를 관리하지 않고, [kotlinx.coroutines.CoroutineDispatcher]를 통해 스레드 풀을 사용합니다.
     */
    suspend fun computeWithCoroutine(
        division: Int,
        dispatcherType: DispatcherType,
    ): Long {
        val coroutineContext = when (dispatcherType) {
            DispatcherType.DEFAULT -> Dispatchers.Default
            DispatcherType.IO -> Dispatchers.IO
            DispatcherType.EXECUTOR -> taskExecutor.asCoroutineDispatcher()
            DispatcherType.SINGLE -> Executors.newSingleThreadExecutor().asCoroutineDispatcher()
        }

        println("coroutineContext: $coroutineContext, division: $division")

        return measureTimeMillis {
            val result = coroutineScope {
                (0 until division).map { i ->
                    async(coroutineContext) {
                        return@async simulateCompute(
                            division = division,
                            idx = i,
                        )
                    }
                }.awaitAll().sum()
            }
            println("total result: $result")
            // Output:
            //  coroutineContext: Dispatchers.IO, division: 4
            //  [simulateCompute]
            //  - Current thread: Thread[#49,DefaultDispatcher-worker-3,5,main]
            //  - ExecutorTracker currentExecutor: null
            //  - start: 2500000
            //  - last: 5000000
            //  - result: 18749997500000
            //  [simulateCompute]
            //  - Current thread: Thread[#48,DefaultDispatcher-worker-2,5,main]
            //  - ExecutorTracker currentExecutor: null
            //  - start: 5000000
            //  - last: 7500000
            //  - result: 31249997500000
            //  [simulateCompute]
            //  - Current thread: Thread[#47,DefaultDispatcher-worker-1,5,main]
            //  - ExecutorTracker currentExecutor: null
            //  - start: 0
            //  - last: 2500000
            //  - result: 6249997500000
            //  [simulateCompute]
            //  - Current thread: Thread[#50,DefaultDispatcher-worker-4,5,main]
            //  - ExecutorTracker currentExecutor: null
            //  - start: 7500000
            //  - last: 10000000
            //  - result: 43749997500000
            //  total result: 99999990000000

            result
        }
    }

    private fun simulateCompute(division: Int, idx: Int): Long {
        val end = 10_000_000
        val chunkSize = end / division
        val start = chunkSize * idx
        val last = if (idx == division - 1) {
            end
        } else {
            start + chunkSize
        }

        val result = LongStream.range(start.toLong(), last.toLong())
            .map { it * 2 }
            .map { it / 2 }
            .map { it * 2 }
            .sum()

        println(
            """
            [simulateCompute]
            - Current thread: ${Thread.currentThread()}
            - ExecutorTracker currentExecutor: ${TrackingExecutorTracker.currentExecutor}
            - start: $start
            - last: $last
            - result: $result
            """.trimIndent()
        )

        return result
    }
}

enum class ExecutorType(val value: String) {
    FIXED_THREAD_POOL("fixed"),
    WORKER_STEALING_POOL("fork-join"),
    SINGLE_THREAD("single"),
    CACHED_THREAD_POOL("cached"),
    VIRTUAL_THREAD("virtual"),
    SCHEDULED_THREAD_POOL("scheduled"),
    CUSTOM_THREAD_POOL("custom");
}

enum class DispatcherType(val value: String) {
    /**
     * JVM과 Kotlin/Native의 공유 스레드 풀을 사용합니다.
     *
     * 이 디스패처가 사용하는 스레드의 최대 개수는 CPU 코어의 개수이고, 최소 개수는 2입니다.
     *
     * 로그를 출력해보면 실제로 각 코루틴마다 별도의 스레드가 생성되는 것을 확인할 수 있습니다.
     * ```
     * chunkSize: 2500000, division: 4
     * [asyncComputeWithAsyncAnnotation]
     * - Thread inside supplyAsync: Thread[#50,DefaultDispatcher-worker-4,5,main]
     * - start: 5000000
     * - last: 7500000
     * - result: 31249997500000
     * [asyncComputeWithAsyncAnnotation]
     * - Thread inside supplyAsync: Thread[#48,DefaultDispatcher-worker-2,5,main]
     * - start: 7500000
     * - last: 10000000
     * - result: 43749997500000
     * [asyncComputeWithAsyncAnnotation]
     * - Thread inside supplyAsync: Thread[#49,DefaultDispatcher-worker-3,5,main]
     * - start: 0
     * - last: 2500000
     * - result: 6249997500000
     * [asyncComputeWithAsyncAnnotation]
     * - Thread inside supplyAsync: Thread[#47,DefaultDispatcher-worker-1,5,main]
     * - start: 2500000
     * - last: 5000000
     * - result: 18749997500000
     * total result: 99999990000000
     * ```
     *
     * References:
     * - [Default](https://kotlinlang.org/api/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-dispatchers/-default.html)
     */
    DEFAULT("default"),

    /**
     * 블로킹 IO 작업(네트워크 요청, 파일 읽기/쓰기 등)들을 '현재 메인 스레드 또는 현재 실행 컨텍스트'에서 '다른 스레드 또는 스레드 풀'로 넘겨 처리하는
     * [kotlinx.coroutines.CoroutineDispatcher]입니다.
     *
     * 필요에 따라 스레드 풀의 스레드는 생성되고 종료됩니다.
     * 이 디스패처에서 작업이 사용하는 스레드의 개수의 상한은 "`kotlinx.coroutines.io.parallelism`" 값으로 정해집니다.
     * 기본적으로 64개의 스레드 또는 코어의 수 중 더 큰 수로 제한됩니다.
     *
     * 이 디스패처는 [Dispatchers.Default]와 스레드를 공유합니다.
     * 따라서, 이미 [Dispatchers.Default]를 사용중일 때,
     * ```
     * withContext(Dispatchers.IO) { }
     * ```
     * 이렇게 [Dispatchers.IO]를 사용한다고 해도 일반적으로 다른 스레드로의 스위칭이 일어나지 않습니다.
     * 이러한 경우 동일한 스레드에서 실행을 유지하려고 합니다.
     *
     * 스레드 공유의 결과, 64개(default parallelism) 이상의 스레드가 생성될 수 있지만 사용되지는 않습니다.
     *
     * References:
     * - [IO](https://kotlinlang.org/api/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-dispatchers/-i-o.html)
     */
    IO("io"),
    EXECUTOR("executor"),
    SINGLE("single");
}