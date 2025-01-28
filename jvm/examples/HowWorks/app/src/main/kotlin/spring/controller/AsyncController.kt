package spring.controller

import kotlinx.coroutines.coroutineScope
import org.apache.coyote.BadRequestException
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import spring.service.AsyncService
import spring.service.DispatcherType
import spring.service.ExecutorType


@RestController
@RequestMapping(path = ["async"])
class AsyncController(
    private val asyncService: AsyncService,
) {
    fun Long.toElapsedResponse() = mapOf(
        "elapsed" to "$this/ms"
    )

    @GetMapping("/sate")
    fun sate(
        @RequestParam(
            name = "virtual",
            required = false,
            defaultValue = "false"
        ) virtualThread: Boolean,
    ) = asyncService.simpleAsyncTaskExecutorExample(virtualThread).toElapsedResponse()

    @GetMapping("/cte")
    fun cte(
        @RequestParam(
            /**
             * [RequestParam.required]가 `false`인 경우 요청 파라미터가 생략되거나 값이 없을 경우에 null로 처리됩니다.
             * 하지만 Kotlin 함수의 매개변수 `type: ExecutorType`는 [org.jetbrains.annotations.NotNull]로 처리됩니다.
             * 이로 인해 다음과 같은 에러가 발생합니다.
             * ```
             * java.lang.NullPointerException: Parameter specified as non-null is null: method spring.controller.AsyncController.cte, parameter type
             *     at spring.controller.AsyncController.cte(AsyncController.kt) ~[main/:na]
             * ```
             * 이를 해결하기 위해서는:
             * - `type: ExecutorType?`로 nullable 하게 수정하거나
             * - [RequestParam.defaultValue] 값을 설정합니다.
             */
            required = false,
            defaultValue = "",
        ) type: ExecutorType,
    ) = asyncService.concurrentTaskExecutorExample(type).toElapsedResponse()

    @GetMapping("compute/async-annotation")
    fun computeWithCompletableFuture() =
        asyncService.asyncComputeWithAsyncAnnotation().thenApply { it.toElapsedResponse() }

    /**
     * References:
     * - [Controllers](https://docs.spring.io/spring-framework/reference/languages/kotlin/coroutines.html#controllers)
     */
    @GetMapping("compute/coroutine")
    suspend fun computeWithCoroutine(
        @RequestParam(
            required = false,
            defaultValue = "1",
        ) division: Int,
        @RequestParam(
            required = false,
            defaultValue = "default"
        ) dispatcher: DispatcherType,
    ) = coroutineScope {
        if (division == 0) throw BadRequestException("Can not division by zero")

        asyncService.computeWithCoroutine(division, dispatcher).toElapsedResponse()
    }
}