# tips

- [tips](#tips)
  - [블록을 매개변수로 넘기기](#블록을-매개변수로-넘기기)
  - [reflection](#reflection)
  - [regex VS Contains. Best Performance?](#regex-vs-contains-best-performance)
  - [What is the equivalent of Java static final fields in Kotlin?](#what-is-the-equivalent-of-java-static-final-fields-in-kotlin)

## 블록을 매개변수로 넘기기

```kotlin
object ApiClient {
    suspend fun post(uri: String, request: Request, headers: (HeadersBuilder.() -> Unit)?): PhpResponse {
        return client.post(uri) {
            timeout {
                requestTimeoutMillis = 10000
            }
            setBody(request.toJson())
            if (headers != null) {
                headers(headers)
            }
        }.body()
    }

    suspend fun postWithAccessToken(uri: String, request: Request): PhpResponse {
        contract()
        return post(uri, request) {
            append(HttpHeaders.Authorization, getAccessToken(apiKey, apiSecret))
        }
    }
}
```

`HeadersBuilder.() -> Unit` 사용함으로써 코드 블록을 매개변수로 넘길 수 있다

## reflection

- <https://stackoverflow.com/a/70178693>

```kotlin
fun Any.toMultiValueMap(): LinkedMultiValueMap<String, String> {
    try {
        val params = LinkedMultiValueMap<String, String>()
        this.javaClass.kotlin.memberProperties.forEach {
            when (val value = it.get(this)) {
                is List<*> -> {
                    setAsList(it.name, value, params)
                }
                is Map<*, *> -> {
                    setAsMap(it.name, value, params)
                }
                else -> setAsString(it.name, value, params)
            }
        }

        return params
    } catch (e: Exception) {
        throw IncorrectCastingException("Failed convert object to MultiValueMap, ${e.localizedMessage}")
    }
}
```

## [regex VS Contains. Best Performance?](https://stackoverflow.com/questions/2023792/regex-vs-contains-best-performance)

- [Both are fast enough, but contains is faster. Facts: ~20mil ops vs ~1mil ops](https://stackoverflow.com/a/53333624)

```kotlin
@State(Scope.Benchmark)
public class Main {

  private String uri = "https://google.com/asdfasdf/ptyojty/aeryethtr";

  @Benchmark
  @Warmup(iterations = 5)
  @Measurement(iterations = 5)
  @Fork(value = 1, warmups = 0)
  public void initContains() throws InterruptedException {
    if (uri.contains("/br/fab") || uri.contains("/br/err") || uri.contains("/br/sts")) {}
  }

  @Benchmark
  @Warmup(iterations = 5)
  @Measurement(iterations = 5)
  @Fork(value = 1, warmups = 0)
  public void initMatches() throws InterruptedException {
    if (uri.matches(".*/br/(fab|err|sts).*")) {}
  }

  public static void main(String[] args) throws Exception {
    org.openjdk.jmh.Main.main(args);
  }
}
```

```log
# Run complete. Total time: 00:00:37

Benchmark           Mode  Cnt         Score         Error  Units
Main.initContains  thrpt    5  21004897.968 ± 1987176.746  ops/s
Main.initMatches   thrpt    5   1177562.581 ±  248488.092  ops/s
```

## [What is the equivalent of Java static final fields in Kotlin?](https://stackoverflow.com/questions/40352879/what-is-the-equivalent-of-java-static-final-fields-in-kotlin)

- [static field](https://kotlinlang.org/docs/java-to-kotlin-interop.html#static-fields)

```kotlin
class Hello {
    companion object {
        // 단, `const val`는 원시(primitive) 타입만 사용할 수 있다
        const val MAX_LEN = 20
    }
}
```

- `@JvmStatic` 사용할 수 있다

```kotlin
class Hello {
    companion object {
        // with static getter & setter
        @JvmStatic val MAX_LEN = 20
    }
}
```

- `@JvmField` 사용할 수 있다

```kotlin
class Hello {
    companion object {
        // without static getter & setter
        @JvmField val MAX_LEN = 20
    }
}
```
