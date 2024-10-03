# Coroutine

- [Coroutine](#coroutine)
    - [Thread와 Coroutine](#thread와-coroutine)
        - [스레드(Thread)?](#스레드thread)
        - [코루틴(Coroutine)?](#코루틴coroutine)
    - [contexts](#contexts)
        - [`Dispatchers.IO`](#dispatchersio)

## Thread와 Coroutine

스레드와 코루틴은 모두 동시성 처리를 위한 도구입니다.
동시성이란 시스템에서 여러 작업이 동시에 실행되거나, 효율적으로 상호작용할 수 있도록 관리하는 것을 의미합니다.
주로 CPU나 네트워크, 파일 입출력 등의 자원을 공유하면서 발생하는 문제를 해결하는 방식입니다.

둘 다 동시성(Concurrency)을 처리하는 메커니즘이지만, 코루틴은 스레드보다 더 가볍고 효율적인 동시성 처리를 제공합니다.

| 항목             | 스레드(Thread)                                                 | 코루틴(Coroutine)                              |
| ---------------- | -------------------------------------------------------------- | ---------------------------------------------- |
| 관리 주체        | 운영체제(OS), JVM이 OS 스레드를 관리                           | Kotlin의 코루틴 라이브러리가 관리              |
| 생성 비용        | 높음 (OS 리소스 및 메모리 할당)                                | 낮음 (스레드 대비 메모리와 리소스 사용이 적음) |
| 동시성 처리 방식 | 운영체제의 **선점형 멀티태스킹** 방식으로 스케줄링             | **협력적 멀티태스킹** (suspend/resume)         |
| 메모리 사용량    | 스레드 하나당 고정된 스택 메모리 할당 (수백 KB~수 MB)          | 동적으로 할당되는 적은 메모리 사용             |
| 비동기 작업      | I/O 작업에서 블로킹 가능, 스레드가 리소스를 점유한 상태로 대기 | 비동기 작업을 **비블로킹** 방식으로 처리 가능  |
| 컨텍스트 스위칭  | 운영체제가 스레드 상태를 저장/복원, 비용이 높음                | **가벼운 컨텍스트 스위칭**, 효율적 상태 전환   |
| 스레드와의 관계  | 각 스레드는 자체 실행 흐름을 가짐                              | 하나의 스레드에서 여러 코루틴 실행 가능        |

### 스레드(Thread)?

스레드는 운영체제(OS)가 관리하는 단위로, 각 스레드는 독립적인 실행 흐름을 가집니다.
- 스레드는 운영체제가 관리하기 때문에 생성 및 전환(Context Switching) 비용이 높습니다.
- 스레드 풀에 제한이 있으며, 많은 스레드를 생성하면 메모리 부족 등의 문제를 일으킬 수 있습니다.
- 스레드에서 I/O 작업이나 네트워크 요청 같은 블로킹 작업이 있으면, 해당 스레드는 다른 작업을 처리할 수 없고, 다른 스레드가 그 일을 대신 처리해야 합니다.

```kotlin
fun main() {
    Thread {
        println("Thread 작업 실행 중...")
    }.start()
}
```

### 코루틴(Coroutine)?

코루틴은 Kotlin에서 제공하는 비동기 프로그래밍의 경량화된 메커니즘입니다.
코루틴은 스레드처럼 여러 작업을 동시에 처리하는 것처럼 보이지만, 운영체제가 아닌 Kotlin 런타임에서 관리되므로 매우 가볍습니다.
- 수천 개의 코루틴을 실행하더라도 한두 개의 스레드를 사용하는 경우가 많습니다. 스레드보다 메모리와 생성 비용이 훨씬 적습니다.
- 특정 작업(I/O 등)이 완료될 때까지 스레드를 차단하지 않고, 그 동안 다른 코루틴이 실행될 수 있도록 관리됩니다.
- `suspend` 키워드를 통해 언제 실행을 일시 중지하고, 다시 재개(resume)할지 직접 관리할 수 있습니다.

```kotlin
import kotlinx.coroutines.*

fun main() = runBlocking {
    launch {
        println("코루틴 작업 실행 중...")
    }
}
```

## contexts

- [Coroutine context and dispatchers](https://kotlinlang.org/docs/coroutine-context-and-dispatchers.html)

### `Dispatchers.IO`

```kotlin
@PostMapping
suspend fun noticeDeposit(@RequestHeader headers: HttpHeaders, request: HttpServletRequest): String {
    val bytes = request.inputStream.readAllBytes()
                                    ^^^^^^^^^^^^^^
                                    Possibly blocking call in non-blocking context could lead to thread starvation
```

```kotlin
@PostMapping
suspend fun noticeDeposit(@RequestHeader headers: HttpHeaders, request: HttpServletRequest): String {
    val bytes = withContext(Dispatchers.IO) {
        request.inputStream.readAllBytes()
    }
```
