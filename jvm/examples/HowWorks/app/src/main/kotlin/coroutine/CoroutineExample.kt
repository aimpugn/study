package coroutine

import kotlinx.coroutines.*
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.filter
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.map
import util.RunExample
import kotlin.system.measureTimeMillis

class CoroutineExamples {
    /**
     * 기본적인 코루틴 예제
     */
    suspend fun whereAreSuspensions(param: String) {
        println("First suspension")
        delay(1000)  // 첫 번째 중단점
        println("Second suspension")
        delay(2000)  // 두 번째 중단점
        println(param)
    }

    /**
     * 기본적인 코루틴 예제
     * launch와 async의 차이를 보여줍니다.
     */
    suspend fun basicCoroutines() = coroutineScope {
        // launch는 Job을 반환하며, 완료를 기다리지 않고 다음 코드를 실행
        val job = launch {
            delay(1000) // 1초 대기
            println("launch 완료")
        }

        // async는 Deferred<T>를 반환하며, await()로 결과를 기다릴 수 있음
        val deferred = async {
            delay(2000) // 2초 대기
            "async 완료" // 반환값
        }

        job.join() // launch가 완료될 때까지 대기
        println(deferred.await()) // async의 결과를 기다리고 출력
    }

    /**
     * 코루틴 컨텍스트와 디스패처의 사용 예제
     */
    suspend fun coroutineContexts() {
        // 메인 스레드에서 실행
        // withContext(Dispatchers.Main) {
        //     println("메인 스레드에서 실행: ${Thread.currentThread().name}")
        // }

        // IO 작업을 위한 스레드 풀에서 실행
        withContext(Dispatchers.IO) {
            println("IO 스레드에서 실행: ${Thread.currentThread().name}")
            // 네트워크 요청이나 파일 작업 등을 시뮬레이션
            delay(100)
        }

        // CPU 집약적 작업을 위한 스레드 풀에서 실행
        withContext(Dispatchers.Default) {
            println("Default 스레드에서 실행: ${Thread.currentThread().name}")
            // 복잡한 계산을 시뮬레이션
            var result = 0
            for (i in 1..1000000) result += i
        }
    }

    /**
     * 코루틴의 취소와 타임아웃 처리 예제
     */
    suspend fun cancellationAndTimeout() = coroutineScope {
        val job = launch {
            try {
                repeat(1000) { i ->
                    println("작업 $i 실행중...")
                    delay(500)
                }
            } catch (e: CancellationException) {
                println("코루틴이 취소됨")
                throw e // 취소를 전파하기 위해 예외를 다시 던짐
            } finally {
                println("정리 작업 수행")
            }
        }

        delay(1500) // 1.5초 대기
        job.cancel() // 코루틴 취소

        // withTimeout을 사용한 타임아웃 처리
        try {
            withTimeout(1000) {
                repeat(10) {
                    delay(200)
                    println("타임아웃 테스트 $it")
                }
            }
        } catch (e: TimeoutCancellationException) {
            println("타임아웃 발생")
        }
    }

    /**
     * Flow를 사용한 비동기 스트림 처리 예제
     */
    suspend fun flowExample() {
        // Flow 생성
        val flow = flow {
            for (i in 1..5) {
                delay(100) // 각 값 사이에 지연
                emit(i) // 값 방출
            }
        }

        // Flow 처리
        flow.map { it * 2 } // 각 값을 2배로
            .filter { it > 5 } // 5보다 큰 값만 필터
            .collect { value -> // 결과 수집
                println("Collected value: $value")
            }
    }

    /**
     * 채널을 사용한 코루틴 간 통신 예제
     */
    suspend fun channelExample() = coroutineScope {
        val channel = Channel<Int>()

        // 생산자 코루틴
        val producer = launch {
            repeat(5) {
                channel.send(it) // 채널에 값 전송
                delay(100)
            }
            channel.close() // 채널 닫기
        }

        // 소비자 코루틴
        val consumer = launch {
            for (value in channel) {
                println("Received: $value")
            }
        }

        // 모든 작업이 완료될 때까지 대기
        producer.join()
        consumer.join()
    }

    /**
     * 예외 처리와 수퍼바이저 잡 예제
     */
    suspend fun exceptionHandling() = coroutineScope {
        val supervisor = SupervisorJob()

        val scope = CoroutineScope(coroutineContext + supervisor)

        val job1 = scope.launch {
            try {
                delay(100)
                throw RuntimeException("Job 1 실패")
            } catch (e: Exception) {
                println("Job 1 예외 처리: ${e.message}")
            }
        }

        val job2 = scope.launch {
            delay(200)
            println("Job 2 완료")
        }

        // 모든 작업이 완료될 때까지 대기
        joinAll(job1, job2)
    }

    /**
     * 실제 사용 사례: 병렬 API 호출 시뮬레이션
     */
    suspend fun parallelApiCalls() = coroutineScope {
        val time = measureTimeMillis {
            val deferreds = List(3) { index ->
                async(Dispatchers.IO) {
                    simulateApiCall(index)
                }
            }

            // 모든 결과 수집
            val results = deferreds.awaitAll()
            println("모든 API 호출 결과: $results")
        }

        println("총 소요 시간: ${time}ms")
    }

    /**
     * API 호출을 시뮬레이션하는 보조 함수
     */
    private suspend fun simulateApiCall(index: Int): String {
        delay(1000) // 네트워크 지연 시뮬레이션
        return "API $index 응답"
    }
}

@RunExample
suspend fun coroutineExamples() {
    coroutineScope {
        withContext(Dispatchers.Default) {
            val examples = CoroutineExamples()

            println("=== 기본 코루틴 예제 ===")
            examples.whereAreSuspensions("whereAreSuspensions")
            examples.basicCoroutines()

            println("\n=== 코루틴 컨텍스트 예제 ===")
            examples.coroutineContexts()

            println("\n=== 취소와 타임아웃 예제 ===")
            examples.cancellationAndTimeout()

            println("\n=== Flow 예제 ===")
            examples.flowExample()

            println("\n=== 채널 예제 ===")
            examples.channelExample()

            println("\n=== 예외 처리 예제 ===")
            examples.exceptionHandling()

            println("\n=== 병렬 API 호출 예제 ===")
            examples.parallelApiCalls()
        }
    }
}

/**
 * ### [Dispatchers.Default]
 *
 * CPU 집약적 작업(CPU-intensive tasks)에 최적화된 스레드 풀을 사용합니다.
 * [DefaultExecutor]라는 공유 스레드 풀에서 코루틴이 실행됩니다.
 * 작업을 균등하게 분배하여 병렬 처리를 최대화합니다.
 *
 * 다음과 같은 작업들에 적합합니다.
 * - 수학 계산, 데이터 처리, 알고리즘 실행 등 CPU를 많이 사용하는 작업.
 * - 비동기 작업이 아닌 계속해서 CPU를 사용하는 연산.
 *
 * ### [Dispatchers.IO]
 *
 * I/O 작업(Input/Output-intensive tasks)에 최적화된 스레드 풀을 사용합니다.
 * 스레드 개수는 제한적이지 않으며, [Dispatchers.Default]의 스레드 풀보다 더 많은 스레드를 생성할 수 있습니다.
 * 기본적으로 JVM에게 허용된 CPU 코어 수의 64배([Runtime.availableProcessors] * 64)까지 스레드를 생성할 수 있습니다.
 *
 * [DefaultExecutor]와 동일한 스레드 풀을 기반으로 하지만, 추가적으로 더 많은 스레드를 생성할 수 있습니다.
 *
 * 다음과 같은 작업들에 적합합니다.
 * - 네트워크 요청, 파일 읽기/쓰기, 데이터베이스 작업 등 I/O 중심의 비동기 작업.
 * - 블로킹 시간이 긴 작업.
 *
 */
@RunExample
suspend fun coroutineWithDispatcher() {
    val jobs = List(100) { // 100,000 코루틴 생성
        /**
         * [CoroutineScope]는 코루틴의 생명주기를 관리하고, 코루틴을 구조적으로 실행할 수 있도록 도와줍니다.
         * 다음을 포함하는 코루틴 컨텍스트(CoroutineContext)를 제공합니다.
         * - [CoroutineDispatcher]: 코루틴이 실행될 스레드를 결정 ([Dispatchers.IO], [Dispatchers.Default] 등).
         * - [Job]: 코루틴의 생명주기를 추적.
         * - 그 외 에러 핸들링, 로그 등의 커스텀 컨텍스트 요소.
         *
         * [CoroutineScope]는 생성된 코루틴을 그룹화하고, 상위 [Job]을 통해 모든 하위 코루틴의 생명주기를 일괄적으로 관리합니다.
         */
        val coroutineScope = CoroutineScope(Dispatchers.Default)
        coroutineScope.launch {// Default Dispatcher 사용 (스레드 풀)
            delay(1000) // Suspend execution
            println("Coroutine $it is done")
        }
    }

    jobs.forEach { it.join() } // 모든 코루틴이 완료되길 기다림
}