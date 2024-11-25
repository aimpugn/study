# Troubleshootings

## Could not autowire. No beans of 'NicePaymentsTransactionRepository' type found

### 문제

```text
Could not autowire. No beans of 'SOME_TYPE' type found
```

### 원인

- [어노테이션 없이도 작동해야 하지만, IntelliJ에서 계속 에러로 보이므로 `@Repository` 어노테이션을 명시](https://stackoverflow.com/a/34823996)

### 해결

- JAP data가 아닌, Spring repository 구현 클래스가 별도로 필요했었다. 실제 구현체를 만드니 이상이 없어졌음

## No enum constant finance.chai.gateway.transaction.model.nice.NicePaymentMethod.CARD

### 문제

- 데이터 읽어올 때 맵핑 에러 발생

```log
java.lang.IllegalArgumentException: No enum constant finance.chai.gateway.transaction.model.nice.NicePaymentMethod.CARD
    at java.lang.Enum.valueOf(Enum.java:273)
Wrapped by: org.springframework.data.mapping.MappingException: Could not read property private final finance.chai.gateway.transaction.model.nice.NicePaymentMethod finance.chai.gateway.transaction.repository.SpringDataNiceTransactionRepository$Companion$NiceTransactionRow.paymentMethod from column payment_method!
    at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:187)
    Suppressed: reactor.core.publisher.FluxOnAssembly$OnAssemblyException: 
Assembly trace from producer [reactor.core.publisher.FluxLiftFuseable] :
    reactor.core.publisher.Flux.handle
    io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
```

### 원인

- `enum class NicePaymentMethod`에 `CARD`로 되어 있는 로우가 없기 때문

### 해결

- 스프링에서 `enum`을 저장할 떄는 `enum`의 이름을 그대로 저장한다
- 가령 아래 `SOME_DETAIL_ENUM`을 어떤 칼럼에 저장하면, `SOME_DETAIL_ENUM` 이대로 저장된다.

```kotlin
enum class SomeEum(
    val key: String?
){
    SOME_DETAIL_ENUM("what is this")        
}
```

```sql
-- SELECT *
-- FROM some_table;
-- some_column1 | some_column2
-- SOME_DETAIL_ENUM | ...
```

## euc-kr의 `%EF%BF%BD` 문제

### 문제

`%C8%AB%B1%E6%B5%BF`(홍길동)을 원하는데, Spring @RequestBody에서 `%C8%AB%EF%BF%BD%E6%B5%BF`를 받음

### 원인

### 해결

```kotlin
@PostMapping
suspend fun noticeDeposit(@RequestHeader headers: HttpHeaders, request: HttpServletRequest): String {
    // Because of `EF BF BD` issue between utf-8 and euc-kr conversion,
    // read bytes and then make it as original EUC-KR string
    val requestBody = withContext(Dispatchers.IO) {
        request.inputStream.readAllBytes()
    }.toString(Charset.forName("EUC-KR"))

    getLogger().info(requestBody)
    val headerMap = headers.toSingleValueMap()
    MDC.put("request_id", headerMap[Constants.HEADER_KEY_X_REQUEST_ID.lowercase()])
    MDC.put("trace_id", headerMap[Constants.HEADER_KEY_X_B3_TRACE_ID.lowercase()])

    notificationService.notifyDeposit(PgProvider.NICE, pgNotificationBody = requestBody)

    return "OK"
}
```

```log
// euc-kr
MallReserved1=4768800525&ServiceCl=0&MallReserved3=&MallReserved2=&RcptAuthCode=I79881493&FnCd=003&BuyerEmail=anony%40some-qwerty-org.io&MallReserved9=&MallReserved8=&MallReserved5=&MallReserved4=&MallReserved7=&MallUserID=&MallReserved6=&MallReserved=imp_uid%3Dimp_123456789012%2Crequest_id%3Dreq_123456789012%2Cuser_code%3Dimp84043725%2Cvbank_due%3D202303082359&StateCd=0&BuyerAuthNum=&VbankName=%B1%E2%BE%F7%C0%BA%C7%E0&MOID=imp_123456789012&VbankNum=05000711997900&MallReserved10=&ReceitType=1&VbankInputName=%C8%AB%B1%E6%B5%BF&AuthCode=&AuthDate=230307231137&MID=someMID&Amt=1000&EtcSvcCl=0&TID=someMID03012303072307055719&GoodsName=%C8%AB%B1%E6%B5%BF%20%6C%6F%76%65%20%6E%61%74%75%72%61%6C%20%67%6F%6C%64%20%72%69%6E%67%20%28%B0%A2%C0%CE%29&MerchantKey=dbW%2Fd3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h%2Bwky%2BLMmP1PbLw%3D%3D&FnName=%B1%E2%BE%F7%C0%BA%C7%E0&RcptTID=someMID04012303072311372047&CancelMOID=&name=%C8%AB%B1%E6%B5%BF&PayMethod=VBANK&ResultMsg=%BD%C2%C0%CE&RcptType=1&ResultCode=4110

// utf-8
MallReserved1=4768800525&ServiceCl=0&MallReserved3=&MallReserved2=&RcptAuthCode=I79881493&FnCd=003&BuyerEmail=anony%40some-qwerty-org.io&MallReserved9=&MallReserved8=&MallReserved5=&MallReserved4=&MallReserved7=&MallUserID=&MallReserved6=&MallReserved=imp_uid%3Dimp_123456789012%2Crequest_id%3Dreq_123456789012%2Cuser_code%3Dimp84043725%2Cvbank_due%3D202303082359&StateCd=0&BuyerAuthNum=&VbankName=%EA%B8%B0%EC%97%85%EC%9D%80%ED%96%89&MOID=imp_123456789012&VbankNum=05000711997900&MallReserved10=&ReceitType=1&VbankInputName=%ED%99%8D%EA%B8%B8%EB%8F%99&AuthCode=&AuthDate=230307231137&MID=someMID&Amt=1000&EtcSvcCl=0&TID=someMID03012303072307055719&GoodsName=%ED%99%8D%EA%B8%B8%EB%8F%99+love+natural+gold+ring+%28%EA%B0%81%EC%9D%B8%29&MerchantKey=dbW%2Fd3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h%2Bwky%2BLMmP1PbLw%3D%3D&FnName=%EA%B8%B0%EC%97%85%EC%9D%80%ED%96%89&RcptTID=someMID04012303072311372047&CancelMOID=&name=%ED%99%8D%EA%B8%B8%EB%8F%99&PayMethod=VBANK&ResultMsg=%EC%8A%B9%EC%9D%B8&RcptType=1&ResultCode=4110
```
