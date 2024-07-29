# thread leak

## http client 이슈

### 문제

`Parent Job is Completed` 문제 해결 위해 HttpClient 생성후 DI → Class에서 생성하는 로직으로 변경

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

근데 thread leak 발생. HttpClient 생성로직 변경이 , `Parent Job is Completed`  에러를 해결하기 위한 근본적인 로직이 아니다. HttpClient 호출 방식이 문제였다.

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

### 해결

```kotlin
val ocrResponse: HttpResponse = httpClient.post(conf.url) {
    ....
}
```
