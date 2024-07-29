# 회고

## data class 정의와 레이어간 전달

### 문제

```kotlin
// application/.../service/CheckoutTransactionServiceNiceLogic.kt
val credential = channel.secret.toObject<NiceCredential>()
val niceRawAuthResult = command.pgRawRequest.toObject<NiceClient.Dto.NiceRawAuthResult>()
val confirmRequestAmt = niceRawAuthResult.amt.toLong()
val confirmRequest = NiceClient.Dto.ConfirmRequest(
    tid = niceRawAuthResult.txTid,
    authToken = niceRawAuthResult.authToken,
    mid = credential.mid,
    amt = confirmRequestAmt,
    charSet = Charsets.UTF_8,
).makeSignData(credential.merchantKey)

val response = niceClient.confirm(
    confirmRequest
)

...


if (!niceClient.verify(confirmRequest, approved, credential.merchantKey)) {
    throw PgpResponseException("Failed to verify signature of approval")
}
```

```kotlin
//domain/.../client/nice/NiceClient.kt
/**
 * 인증 결과[NiceRawAuthResult]로 승인 요청
 */
suspend fun confirm(
    request: ConfirmRequest
): Client.Companion.WrappedResponse<PaymentResponse>
}
```

```kotlin
// infrastructure/.../client/nice/NiceHttpClient.kt
override suspend fun confirm(
    request: NiceClient.Dto.ConfirmRequest
) = requestFormData<NiceClient.Dto.PaymentResponse> {
    post().uri("/webapi/pay_process.jsp").bodyValue(
        request.toMultiValueMap(charset(charsetName), naming)
    )
}
```

DDD에서 요청은 presentation -> application -> domain -> infrastructure 순서로 전달이 되는데,
이미 `application` 레이어에서 `infrastructure` 레이어에서 PG사로 보낼 데이터 클래스를 미리 인스턴스화 하며, 이는 안티 패턴에 해당했다.

`domain` 레이어에 정의된 `*Client`의 Dto는 도메인에서 다음 레이어로 가장 보내기 편한 방법을 정의해야 하지, PG사로 보낼 사양에 종속되면 안됐다.

### 개선

1. 나이스의 경우 인증 결과가 JSON 문자열(`command.pgRawRequest`)로 넘어오며, 인증 결과 데이터를 transaction 로우 저장 시 사용하기 위해 `application` 레이어에서 데이터 클래스로 변환한다. 그 외에는 필요 없으므로 credential과 인증 원천 데이터는 바로 도메인 레이어로 내린다

```kotlin
// application/.../service/CheckoutTransactionServiceNiceLogic.kt
val niceRawAuthResult = command.pgRawRequest.toObject<NiceClient.Dto.NiceRawAuthResult>()
val response = niceClient.confirm(
    channel.secret.toObject(),
    niceRawAuthResult
)
```

2. 도메인 레이어는 인프라에서 구현할 메서드 시그니처를 정의해 둔다

```kotlin
// domain/.../client/nice/NiceClient.kt
/**
 * 인증 결과[NiceRawAuthResult]로 승인 요청
 */
suspend fun confirm(
    niceCredential: NiceCredential,
    niceRawAuthResult: NiceRawAuthResult
): Client.Companion.WrappedResponse<PaymentResponse>
```

3. 클라이언트 구현체는 실제 데이터 클래스 변환을 담당하고, 비정상 케이스 경우 등에 대해 익셉션 발생 시킨다

```kotlin
// infrastructure/.../client/nice/NiceHttpClient.kt
override suspend fun confirm(
    niceCredential: NiceCredential,
    niceRawAuthResult: NiceClient.Dto.NiceRawAuthResult
): Client.Companion.WrappedResponse<NiceClient.Dto.PaymentResponse> {
    val confirmRequest = ConfirmRequest(
        tid = niceRawAuthResult.txTid,
        authToken = niceRawAuthResult.authToken,
        mid = niceCredential.mid,
        amt = niceRawAuthResult.amt.toLong(),
        charSet = Charsets.UTF_8,
        signData = makeSignData(niceRawAuthResult, niceCredential)
    )
    val response = requestFormData<NiceClient.Dto.PaymentResponse> {
        post().uri("/webapi/pay_process.jsp").bodyValue(
            confirmRequest.toMultiValueMap(charset(charsetName), naming)
        )
    }

    if (!verify(confirmRequest, response.body, niceCredential.merchantKey)) {
        throw PgpResponseException("Failed to verify signature of approval")
    }

    return response
}
```

## 이넘 코드 600개 이상을 정하는 것

### 문제

나이스페이먼츠의 경우 응답 코드가 600개를 넘는데, 이를 모두 이넘으로 선언해야 할까?

### 생각

결과 코드에 따라 응답을 명확하게 내려주려면 결국 정규화가 필요하다.
이를 호출하는 곳에서는 응답하는 결과(상태 코드, 메타 데이터, 응답 데이터 등)를 믿고 처리하게 되는데,
시작부터 분류가 명확하지 않으면 나중에 클라이언트의 코드가 자주 수정될 수 있으며, 사용성이 떨어지게 된다.
너무 많다고 생각되더라도 결국 다 맵핑하는 게 맞지 않을까 싶다.
