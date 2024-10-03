import kotlinx.coroutines.*
import java.util.concurrent.atomic.AtomicInteger
import kotlin.concurrent.thread

// 동시성 문제를 시연하기 위한 공유 자원
private var sharedResource = 0

// 원자성 연산을 위한 AtomicInteger
// AtomicInteger는 스레드 안전한 정수 연산을 제공합니다.
// 참조: https://kotlinlang.org/api/latest/jvm/stdlib/kotlin.concurrent/java.util.concurrent.atomic.-atomic-integer/
private val atomicCounter = AtomicInteger(0)

fun main() {
    // 1. 기본 스레드 생성 및 실행
    // thread() 함수는 새 스레드를 생성하고 즉시 시작합니다.
    // 참조: https://kotlinlang.org/api/latest/jvm/stdlib/kotlin.concurrent/thread.html
    thread {
        println("별도의 스레드에서 실행: ${Thread.currentThread().name}")
    }

    // 2. Runnable 인터페이스를 사용한 스레드 생성
    // Runnable은 실행 가능한 작업을 나타내는 인터페이스입니다.
    // 참조: https://docs.oracle.com/javase/8/docs/api/java/lang/Runnable.html
    val runnable = Runnable {
        println("Runnable에서 실행: ${Thread.currentThread().name}")
    }
    Thread(runnable).start()

    // 3. 람다를 사용한 스레드 생성
    // Kotlin의 람다 표현식을 사용하여 간결하게 스레드를 생성할 수 있습니다.
    Thread {
        println("람다로 생성된 스레드에서 실행: ${Thread.currentThread().name}")
    }.start()

    // 4. 동시성 문제 시연
    // 여러 스레드가 동시에 공유 자원에 접근하면 경쟁 조건이 발생할 수 있습니다.
    val threads = List(1000) {
        thread {
            // 동시성 문제가 발생할 수 있는 연산
            // 여러 스레드가 동시에 이 변수를 증가시키려 하면, 일부 증가 연산이 손실될 수 있습니다.
            sharedResource++
            // 원자적 연산 (동시성 문제 없음)
            // incrementAndGet()은 원자적으로 값을 증가시키고 반환합니다.
            // 참조: https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/atomic/AtomicInteger.html#incrementAndGet--
            atomicCounter.incrementAndGet()
        }
    }
    // join()은 모든 스레드가 완료될 때까지 대기합니다.
    // 참조: https://kotlinlang.org/api/latest/jvm/stdlib/kotlin.concurrent/thread.html
    threads.forEach { it.join() }

    println("일반 카운터 값: $sharedResource") // 예상보다 작은 값이 출력될 수 있음
    println("원자적 카운터 값: ${atomicCounter.get()}") // 항상 1000이 출력됨

    // 5. 코루틴 사용 예제
    // runBlocking은 코루틴을 생성하고 완료될 때까지 현재 스레드를 차단합니다.
    // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/run-blocking.html
    runBlocking {
        // launch는 새 코루틴을 시작하고 호출자에게 즉시 반환합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/launch.html
        launch {
            // delay는 지정된 시간 동안 코루틴을 일시 중단합니다.
            // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/delay.html
            delay(1000L)
            println("코루틴에서 실행: ${Thread.currentThread().name}")
        }
        println("runBlocking: ${Thread.currentThread().name}")
    }

    // 6. 코루틴 스코프 및 Job 사용
    // GlobalScope는 애플리케이션 수명 주기 동안 실행되는 최상위 코루틴을 시작합니다.
    // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-global-scope/
    val job = GlobalScope.launch {
        delay(1000L)
        println("GlobalScope에서 실행: ${Thread.currentThread().name}")
    }
    runBlocking {
        // join()은 Job이 완료될 때까지 대기합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-job/join.html
        job.join()
    }

    // 7. async를 사용한 비동기 작업
    runBlocking {
//            this.coroutineContext
        // async는 결과를 반환하는 비동기 작업을 시작합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/async.html
        val deferredResult = async {
            delay(1000L)
            "비동기 작업 결과"
        }
        // await()는 비동기 작업의 결과를 기다리고 반환합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-deferred/await.html
        println("async 결과: ${deferredResult.await()}")
    }

    // 8. 여러 코루틴 동시 실행
    runBlocking {
        val jobs = List(5) {
            launch {
                delay((it * 100).toLong())
                println("코루틴 $it 실행")
            }
        }
        // joinAll()은 모든 Job이 완료될 때까지 대기합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/join-all.html
        jobs.joinAll()
    }

    // 9. withContext를 사용한 컨텍스트 전환
    runBlocking {
        // withContext는 지정된 코루틴 컨텍스트에서 주어진 일시 중단 블록을 호출하고 완료될 때까지 일시 중단합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/with-context.html
        val result = withContext(Dispatchers.Default) {
            // CPU 집약적인 작업 시뮬레이션
            var sum = 0
            for (i in 1..1_000_000) {
                sum += i
            }
            sum
        }
        println("withContext 결과: $result")
    }

    // 10. 코루틴 취소 예제
    runBlocking {
        val job = launch {
            try {
                repeat(1000) { i ->
                    println("작업 $i 수행 중...")
                    delay(500L)
                }
            } catch (e: CancellationException) {
                // 코루틴이 취소되면 CancellationException이 발생합니다.
                // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-cancellation-exception/
                println("코루틴이 취소됨")
            } finally {
                println("최종 정리 작업 수행")
            }
        }

        delay(1300L) // 잠시 대기
        println("코루틴 취소")
        // cancel()은 Job을 취소합니다.
        // 참조: https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-job/cancel.html
        job.cancel()
        job.join() // 취소 완료 대기
    }
}