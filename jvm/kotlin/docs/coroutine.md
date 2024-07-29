# Coroutine

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
