package networks

import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import util.RunExample
import java.lang.reflect.Field
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.time.Duration
import java.util.concurrent.CompletableFuture
import kotlin.concurrent.thread

object HttpClientProvider {
    // Singleton HttpClient
    val client: HttpClient = HttpClient.newBuilder()
        .connectTimeout(Duration.ofSeconds(10)) // 연결 시간 제한
        .version(HttpClient.Version.HTTP_2) // HTTP/2 지원
        .build()
}

fun fetchUrl(url: String): HttpResponse<String>? {
    // -- 동기적 실행 부분 시작 --
    val client = HttpClientProvider.client
    val request = HttpRequest.newBuilder()
        .uri(URI.create(url))
        .GET()
        .build()
    println("[${Thread.currentThread().name}] HttpClient instance hashCode: ${client.hashCode()}")
    printConnectionPool(client)
    // -- 동기적 실행 부분 끝 --

    // -- 비동기 작업 시작 --
    val response: CompletableFuture<HttpResponse<String>> =
        client.sendAsync(request, HttpResponse.BodyHandlers.ofString())

    // -- 비동기 작업 완료 대기 --
    return response.join() // blocking
}

fun printConnectionPool(client: HttpClient) {
    try {
        // HttpClientImpl의 실제 구현체를 가져옴
        val implField: Field = client::class.java.getDeclaredField("impl")
        implField.isAccessible = true
        val clientImpl = implField.get(client)

        // ConnectionPool 필드 접근
        val poolField: Field = clientImpl::class.java.getDeclaredField("connections")
        poolField.isAccessible = true
        val connectionPool = poolField.get(clientImpl)

        println("[${Thread.currentThread().name}] ConnectionPool instance: $connectionPool, hashCode: ${connectionPool.hashCode()}")
    } catch (e: Exception) {
        println(
            "[${Thread.currentThread().name}] Error accessing connection pool: ${e.message}, ${e.cause}, ${e.localizedMessage}\n${
                e.stackTrace.toList().joinToString("\n")
            }"
        )
    }
}

@RunExample
fun multiThreadRequestsExample() {
    val threads = mutableListOf<Thread>()
    listOf(
        "https://www.google.com",
        "https://www.naver.com",
        "https://www.facebook.com"
    ).forEach { url ->
        threads += thread {
            val result = fetchUrl(url)

            println("[${Thread.currentThread().name}] status: ${result?.statusCode()}, contents-length: ${result?.body()?.length}")
        }
    }

    threads.forEach { it.join() }
    // 클라이언트는 같은 클라이언트를 사용하지만, 결국 그 내부에서는 ConnectionPool을 통해 각 스레드별로 별도의 Connection들을 관리합니다.
    //
    // - hashCode가 같은 것으로 보아 같은 클라이언트를 사용함을 알 수 있습니다.
    // - 그리고 ConnectionPool도 같은 풀을 사용하되, 서로 다른 커넥션을 사용함을 알 수 있습니다.
    // - 모든 스레드가 동시에 시작되어 즉시 실행됩니다.
    // - OS 스케줄러의 선점적(preemptive) 스레드 스케줄링으로 인해 실행 순서는 무작위입니다.
    // - 시스템 자원(CPU, 메모리)을 놓고 경쟁합니다.
    //
    // Output:
    //  [Thread-2] HttpClient instance hashCode: 355161145
    //  [Thread-1] HttpClient instance hashCode: 355161145
    //  [Thread-0] HttpClient instance hashCode: 355161145
    //  [Thread-0] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@7e48a851, hashCode: 2118690897
    //  [Thread-2] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@7e48a851, hashCode: 2118690897
    //  [Thread-1] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@7e48a851, hashCode: 2118690897
    //  [Thread-1] status: 200, contents-length: 194759
    //  [Thread-0] status: 200, contents-length: 20103
    //  [Thread-2] status: 200, contents-length: 77815
}

@RunExample
suspend fun coroutineRequestsExample() {
    coroutineScope {
        listOf(
            "https://www.google.com",
            "https://www.naver.com",
            "https://www.facebook.com"
        ).map { url ->
            // async로 각 URL을 병렬로 요청
            // 하나의 워커 스레드가 코루틴을 실행하면서 `fetchUrl()` 함수 내부의 blocking 연산들(println, 객체 생성 등)을 순차적으로 실행
            async {
                val result = fetchUrl(url) // 여기까지는 순차 실행

                // 이후 결과 출력은 나중에
                println(
                    "[${Thread.currentThread().name}] " +
                            "status: ${result?.statusCode()}, " +
                            "contents-length: ${result?.body()?.length}"
                )
            }
        }.awaitAll() // 모든 결과 대기

        // - hashCode가 같은 것으로 보아 같은 클라이언트를 사용함을 알 수 있습니다.
        // - 그리고 ConnectionPool도 같은 풀을 사용하되, 서로 다른 커넥션을 사용함을 알 수 있습니다.
        // - 코루틴은 DefaultDispatcher(Dispatchers.Default)의 스레드 풀을 사용하며, 각 작업이 순차적으로 스레드 풀의 워커에 할당됩니다.
        // - 협력적(cooperative) 스케줄링으로 인해 실행 패턴이 더 예측 가능합니다.
        // - 함수 내부의 blocking 연산들(println, 객체 생성 등)을 순차적으로 실행됨을 알 수 있습니다.
        //
        // Output:
        //  [DefaultDispatcher-worker-1] HttpClient instance hashCode: 1442919923
        //  [DefaultDispatcher-worker-1] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@4d04e7bd, hashCode: 1292167101
        //  [DefaultDispatcher-worker-2] HttpClient instance hashCode: 1442919923
        //  [DefaultDispatcher-worker-2] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@4d04e7bd, hashCode: 1292167101
        //  [DefaultDispatcher-worker-3] HttpClient instance hashCode: 1442919923
        //  [DefaultDispatcher-worker-3] ConnectionPool instance: jdk.internal.net.http.ConnectionPool@4d04e7bd, hashCode: 1292167101
        //  [DefaultDispatcher-worker-2] status: 200, contents-length: 199128
        //  [DefaultDispatcher-worker-1] status: 200, contents-length: 21161
        //  [DefaultDispatcher-worker-3] status: 200, contents-length: 78321
    }
}