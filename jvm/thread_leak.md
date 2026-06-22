# thread leak

## http client 이슈

### 문제

`Parent Job is Completed` 에러를 잡으려고, HttpClient를 DI로 주입받던 방식을 클래스 안에서 직접 생성하는 방식으로 바꿨다.

```kotlin
private val httpClient: HttpClient get() = HttpClient(CIO) {
    expectSuccess = false
    install(JsonFeature) {
        serializer = KotlinxSerializer(
            Json {
                isLenient = true
                ignoreUnknownKeys = true
            }
        )
    }
    install(HttpTimeout) {
        requestTimeoutMillis = 10000
    }
}
```

그랬더니 이번엔 thread leak이 터졌다. 돌이켜보면 생성 위치를 바꾼 것 자체는 `Parent Job is Completed`의 근본 해결책이 아니었다. 진짜 문제는 HttpClient를 *호출하는 방식*에 있었다.

```kotlin
httpClient.use { client ->
    val ocrResponse: HttpResponse = client.post(conf.url) {
        contentType(ContentType.Application.Json)
        headers {
            append("X-OCR-SECRET", conf.ocrSecretKey)
        }
        body = RequestBody(
            listOf(Image(fileFormat, "some-service-ocr", base64String)),
            "business_license",
            "V2",
            0
        )
    }
    val responseBody: String = ocrResponse.receive()

    parseResponseBody(responseBody).mapLeft {
        return Either.Left(NcloudClient.Failure.InternalServerError(it.message, it))
    }
}
```

### 원인

Ktor `HttpClient`는 가벼운 객체가 아니다. 엔진(여기선 `CIO`)이 자체 **코루틴 디스패처와 I/O 스레드**(셀렉터·워커)를 끼고 도는 무거운 자원이다. 그래서 Ktor 공식 문서도 클라이언트 생성을 "값싼 연산이 아니니 여러 요청에 인스턴스를 재사용하라"고 못박고, `close()`는 더 쓰지 않을 때 한 번만 부르라고 안내한다.

위 코드는 이 권고를 두 군데서 한꺼번에 어겼다.

1. `val httpClient get() = HttpClient(CIO) { ... }` 는 **커스텀 게터**다. `httpClient`를 참조할 때마다 게터 본문이 다시 실행되니, 그때마다 *새 클라이언트*(= 새 스레드 셋)가 태어난다.
2. `httpClient.use { ... }` 는 블록 끝에서 `close()`를 부르긴 한다. 하지만 CIO 엔진의 `close()`는 **곧장 반환하고 종료 작업은 백그라운드에서 진행**한다 — 코루틴 기반이라 즉시·동기로 정리가 끝나지 않는다. 요청이 잦으면 만드는 속도가 정리되는 속도를 앞질러, 미처 못 죽은 엔진 스레드가 점점 쌓인다.

화근은 "요청마다 만들고 닫기"를 되풀이한 것이다. 이는 [java/resource_management_and_leaks.md](./java/resource_management_and_leaks.md)가 말하는 **외부 자원 누수**(스레드·핸들을 OS/풀의 하드 한도까지 쌓는 것)의 한 형태다. `close()`를 부르더라도 생성이 잦고 정리가 뒤늦으면 결국 누수처럼 쌓인다.

### 해결

핵심은 매 요청을 `use`로 닫는 게 아니라 **클라이언트를 한 번 만들어 재사용**하는 것이다. 인스턴스를 살려 둔 채 요청만 보내고, `close()`는 정말 그만 쓸 때(앱 종료 등) 한 번만 부른다.

```kotlin
// use { } 로 감싸지 않고, 살아 있는 클라이언트로 요청만 보낸다
val ocrResponse: HttpResponse = httpClient.post(conf.url) {
    contentType(ContentType.Application.Json)
    headers { append("X-OCR-SECRET", conf.ocrSecretKey) }
    body = RequestBody(...)
}
```

다만 `use`를 떼는 건 1차 교정일 뿐이다. `httpClient`가 여전히 *게터*라면 참조할 때마다 새 클라이언트가 생기는 1번 문제가 그대로 남는다. 게터를 한 번만 초기화되는 인스턴스로 바꿔야 "재사용"이 비로소 완성된다.

```kotlin
// get() 게터(X) → 한 번만 초기화되는 프로퍼티(O)
private val httpClient = HttpClient(CIO) { /* 위와 동일한 설정 */ }
// 또는 DI 컨테이너에 싱글턴으로 등록
```

## 관련 노트

- [java/resource_management_and_leaks.md](./java/resource_management_and_leaks.md) — 힙 vs 외부 자원, `close()`와 누수의 일반 원리. 이 사례의 토대.
- [grafana_jvm_dashboard.md](./grafana_jvm_dashboard.md) — live threads 단조 증가로 스레드 누수를 *발견*하는 법(이 문서는 그 *원인 규명*).
