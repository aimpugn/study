package async

import java.util.concurrent.*
import java.util.function.Supplier


fun main() {
    val executor = Executors.newFixedThreadPool(2);

    val future = executor.submit(Callable { ->
        // 비동기 작업
        Thread.sleep(1000);
        42
    })

    // 나중에 결과를 받아옴
    val result = future.get(); // 블로킹 호출
    println("result: $result")
    executor.shutdown()

    // ForkJoinPool.commonPool()에서 비동기적으로 이 Supplier를 실행하여 CompletableFuture를 반환합니다.
    // - CompletableFuture: 비동기 작업의 결과를 비동기적으로 반환할 수 있도록 설계된 클래스
    val cfuture = CompletableFuture.supplyAsync { 42 };

    cfuture.thenApply {
        // 비동기 작업 결과로 받습니다.
        it * 2;
    }.thenAccept {
        println("최종 결과: $it");
    };
}