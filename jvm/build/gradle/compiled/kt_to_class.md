# kt to class file

kotlin 코드가 다음과 같다면,

```kt
/**
* yyyyMMddHHmmss: 20221212164400
*/
private fun ediDate() = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"))
private final val charset = charset("euc-kr")
private final val naming = NicepayPropertyNamingStrategy()
val formUrlencodedEucKr =
    MediaType.valueOf(MediaType.APPLICATION_FORM_URLENCODED.toString() + ";charset=${charset.name()}")

private suspend inline fun <reified T : NiceResponseWrapper<T>> requestFormData(
    crossinline block: WebClient.() -> RequestSpec
): Client.Companion.WrappedResponse<T> {
    val response: Client.Companion.WrappedResponse<T> = this.request(niceTotalTimeoutMillis) {
        niceWebClient
            .block()
            .headers {
                val formUrlEncodedEucKr = formUrlencodedEucKr
                it.contentType = formUrlEncodedEucKr
                it.accept = listOf(
                    MediaType.APPLICATION_FORM_URLENCODED,
                    MediaType.APPLICATION_JSON,
                    formUrlEncodedEucKr
                )
            }
    }

    nicepayResponseValidator.validate(response)

    return response
}
```

컴파일된 바이트 코드는 아래와 같다

`cat NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1$1.class`

```bytecode
����=akservice/package/name/transaction/nicepay/NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1$1E<T:Ljava/lang/Object;>Ljava/lang/Object;Ljava/util/function/Consumer;java/lang/Objectava/util/function/Consumer:service/package/name/transaction/nicepay/NicepayHttpClienrequestFormDataT(Lkotlin/jvm/functions/Function1;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;


<init>?(Lservice/package/name/transaction/nicepay/NicepayHttpClient;)Vthis$0<Lservice/package/name/transaction/nicepay/NicepayHttpClient;
                                                                                                                                                ()V

thismLservice/package/name/transaction/nicepay/NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1$1;      $receiveraccept)(Lorg/springframework/http/HttpHeaders;)VitgetFormUrlencodedEucKr&()Lorg/springframework/http/MediaType;

        $org/springframework/http/HttpHeaders setContentType'(Lorg/springframework/http/MediaType;)V
                                                                                                    "#
!$"org/springframework/http/MediaType&PPLICATION_FORM_URLENCODED$Lorg/springframework/http/MediaType;
                                                                                                     () '*APPLICATION_JSON
                                                                                                                          ,)    '- kotlin/collections/CollectionsKt/listOf%([Ljava/lang/Object;)Ljava/util/List;
                                                                                                                                                                                                                12
03      setAccept(Ljava/util/List;)V
                                    56
!7formUrlEncodedEucKr&Lorg/springframework/http/HttpHeaders;(Ljava/lang/Object;)V

<p0Ljava/lang/Object;Lkotlin/Metadata;mvkxi0d1R��
��

��

��
                                                                                                                                                                                                                                       ��0��H02
 *00H
d2
  <anonymous>T>Lservice/package/name/transaction/nicepay/NiceResponseWrapper;kotlin.jvm.PlatformTypeWservice/package/name/transaction/nicepay/NicepayHttpClient$requestFormData$response$1$1iservice/package/name/transaction/nicepay/NiinvokeSuspend&(Ljava/lang/Object;)Ljava/lang/Object;FormData$1Q
                                                    STNicepayHttpClient.ktCodeLocalVariableTableLineNumberTableMethodParameters
                                                                                                                               InnerClassesEnclosingMethod      Signature
W2urceFileSourceDebugExtensionRuntimeVisibleAnnotations1
*+�*��X

W�+*��M+,�%+�'N-�+S-�.S-,S-�4�8�Y*
?@A@!B#@$?*DX#9)++:ZA;W=        *+�!�=�Y<X              >?[
\RU]^V_�SMAP
NicepayHttpClient.kt
Kotlin
*S Kotlin
*F
+ 1 NicepayHttpClient.kt
service/package/name/transaction/nicepay/NicepayHttpClient$requestFormData$response$1$1
*L
1#1,799:1
*E
`F@A[IBICIBDIEFIGH[sIJ[ sKsLsMsNs:sOssP%
```

`it.contentType = formUrlEncodedEucKr` 코드는 `$org/springframework/http/HttpHeaders setContentType'(Lorg/springframework/http/MediaType;)V` 처럼 변환된다
