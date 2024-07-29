# Troubleshootings

- [Troubleshootings](#troubleshootings)
    - [.UnsupportedMediaTypeException: Content type 'application/x-www-form-urlencoded' not supported for bodyType](#unsupportedmediatypeexception-content-type-applicationx-www-form-urlencoded-not-supported-for-bodytype)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Not enough information to infer type variable T](#not-enough-information-to-infer-type-variable-t)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - ['inline' modifier is not allowed on virtual members. Only private or final members can be inline](#inline-modifier-is-not-allowed-on-virtual-members-only-private-or-final-members-can-be-inline)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [AES SecretKeyFactory not available](#aes-secretkeyfactory-not-available)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [An overriding function is not allowed to specify default values for its parameters](#an-overriding-function-is-not-allowed-to-specify-default-values-for-its-parameters)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [Incorrect casting object \[error cause:null\]](#incorrect-casting-object-error-causenull)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [](#)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - [Could not read property private final ... from column](#could-not-read-property-private-final--from-column)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-7)
    - [Non exhaustive 'when' statements on enum will be prohibited in](#non-exhaustive-when-statements-on-enum-will-be-prohibited-in)
        - [문제](#문제-8)
        - [원인](#원인-8)
        - [해결](#해결-8)

## .UnsupportedMediaTypeException: Content type 'application/x-www-form-urlencoded' not supported for bodyType

### 문제

```text
Failed to communication with pg!: Content type 'application/x-www-form-urlencoded' not supported for bodyType=org.springframework.web.reactive.function.BodyInserters$$Lambda$1598/0x0000000801a33bf0; nested exception is org.springframework.web.reactive.function.UnsupportedMediaTypeException: Content type 'application/x-www-form-urlencoded' not supported for bodyType=org.springframework.web.reactive.function.BodyInserters$$Lambda$1598/0x0000000801a33bf0)
```

### 원인

### 해결

## Not enough information to infer type variable T

### 문제

```kotlin
    private suspend inline fun <reified T : Any> requestFormData(
        crossinline block: WebClient.() -> RequestSpec
    ): Client.Companion.WrappedResponse<T> {
        val response = this.request(niceTotalTimeoutMillis) {
                         // ^-- 에러 발생
```

### 원인

- 응답해야 하는 타입을 추론할 수 없었기 때문에 발생.

### 해결

- 응답되는 타입을 명시

```kotlin
    private suspend inline fun <reified T : Any> requestFormData(
        crossinline block: WebClient.() -> RequestSpec
    ): Client.Companion.WrappedResponse<T> {
        val response: Client.Companion.WrappedResponse<T> = this.request(niceTotalTimeoutMillis) {
```

## 'inline' modifier is not allowed on virtual members. Only private or final members can be inline

### 문제

```kotlin
class SomeClass {
    inline fun <reified T : Any> validate(response: Client.Companion.WrappedResponse<T>) {
    // ^-- 에러 발생    
        val response = wrappedResponse.body
```

### 원인

- 함수가 override 될 수 있는 virtual이기 때문
- `private` 사용하면 non-virtual, closed 함수가 된다

### 해결

- `private`으로 해도 괜찮다면, `private`으로 만든다

```kotlin
class SomeClass {
    private inline fun <reified T : Any> validate(response: Client.Companion.WrappedResponse<T>) {
    // ^-- 에러 발생    
        val response = wrappedResponse.body
```

## AES SecretKeyFactory not available

### 문제

- `SecretKeyFactory` 사용해서 암호화 하려니 에러 발생

### 원인

- [이미 알려진 버그](https://stackoverflow.com/a/8397072)라고 한다
- [SecretKeyFactory doesn't support algorithm "AES" on Windows and Linux](https://bugs.openjdk.org/browse/JDK-7022467) 코멘트에 따르면 `SecretKeySpec` 사용해서 AES 키 생성할 수 있으니, 사실 Factory가 굳이 필요 없다고 한다

```text
This is failing because there is no Java implementation of AES SecretKeyFactory in the JDK. On Solaris it works because there is a PKCS11 implementation. We should not include this as a required algorithm - this was a mistake. An AES SecretKeyFactory implementation does not provide much value, since developers can use a generic SecretKeySpec object to create an AES key and don't really need a SecretKeyFactory.

I will file a CCC and remove this from the required algorithms.
```

### 해결

- `SecretKeySpec` 사용

```kotlin
val secretKeySpec = SecretKeySpec(password.toByteArray(), "AES")
```

## An overriding function is not allowed to specify default values for its parameters

### 문제

```kotlin
interface A {
    fun test(val default: String? = null)
}

class B : A {
    override test(val default: String? = null) {
        //                            ^^^^^^^^ 에러 발생
        // IMPLEMENT
    }
}
```

### 원인

- 오버라이드 하는 함수는 기본 값 지정할 수가 없음

### 해결

- 오버라이드 한 함수에서 기본값 제거

## Incorrect casting object [error cause:null]

### 문제

```log
finance.chai.gateway.transaction.exception.IncorrectCastingException: Incorrect casting object [error cause:null]
 at finance.chai.gateway.transaction.util.JsonExtKt.toJson(JsonExt.kt:29)
 at finance.chai.gateway.transaction.service.CancelTransactionServiceNiceLogic.cancel$suspendImpl(CancelTransactionServiceNiceLogic.kt:86)
 at finance.chai.gateway.transaction.service.CancelTransactionServiceNiceLogic$cancel$1.invokeSuspend(CancelTransactionServiceNiceLogic.kt)
 at kotlin.coroutines.jvm.internal.BaseContinuationImpl.resumeWith(ContinuationImpl.kt:33)
 at kotlinx.coroutines.DispatchedTask.run(DispatchedTask.kt:106)
 at kotlinx.coroutines.scheduling.CoroutineScheduler.runSafely(CoroutineScheduler.kt:571)
 at kotlinx.coroutines.scheduling.CoroutineScheduler$Worker.executeTask(CoroutineScheduler.kt:750)
 at kotlinx.coroutines.scheduling.CoroutineScheduler$Worker.runWorker(CoroutineScheduler.kt:678)
 at kotlinx.coroutines.scheduling.CoroutineScheduler$Worker.run(CoroutineScheduler.kt:665)

```

### 원인

- JSON 변환 시 중간에 `LocalDateTime` 직렬화 에러 발생. 그런데 이 전에는 정상적으로 작동했었음
- 문제는 `LocalDateTime` 타입의 문제가 아니라, 임의의 getter가 추가돼서 발생했던 문제. `cancelledAt` 필드가 없지만 갖고 있는 속성을 사용해서 언제 취소 되는지 연산하는 `getCancelledAt` 메서드를 추가했더니 `cancelledAt` 속성을 직렬화 시도

```log
 0 = {JsonMappingException$Reference@20037} "finance.chai.gateway.transaction.client.nice.NiceClient$Dto$CancelResponse["cancelledAt"]"
_processor = {WriterBasedJsonGenerator@18544} 
_location = null
backtrace = {Object[6]@20028} 
detailMessage = "Java 8 date/time type `java.time.LocalDateTime` not supported by default: add Module "com.fasterxml.jackson.datatype:jackson-datatype-jsr310" to enable handling"
cause = {InvalidDefinitionException@19864} "com.fasterxml.jackson.databind.exc.InvalidDefinitionException: Java 8 date/time type `java.time.LocalDateTime` not supported by default: add Module "com.fasterxml.jackson.datatype:jackson-datatype-jsr310" to enable handling (through reference chain: finance.chai.gateway.transaction.client.nice.NiceClient$Dto$CancelResponse["cancelledAt"])"
 _type = {SimpleType@20026} "[simple type, class java.time.LocalDateTime]"
 _beanDesc = null
 _property = null
 _path = {LinkedList@20027}  size = 1
  0 = {JsonMappingException$Reference@20037} "finance.chai.gateway.transaction.client.nice.NiceClient$Dto$CancelResponse["cancelledAt"]"
 _processor = {WriterBasedJsonGenerator@18544} 
  _writer = {SegmentedStringWriter@18199} 
   _buffer = {TextBuffer@18278} "{"ResultCode":"2001","ResultMsg":"취소 성공","TID":"someMID01012212020158314867","CancelAmt":100,"MID":"someMID","Moid":"imp_131894456221","Signature":"some signature","PayMethod":"CARD","CancelDate":"20221205","CancelTime":"183714","CancelNum":"취소번호","RemainAmt":"취소 후 잔액","MallReserved":"","cancelledAt""
   writeBuffer = null
   lock = {SegmentedStringWriter@18199} 
```

### 해결

- `getCanclledAt` 메서드를 지우고, `get~` 아닌 `at()` 메서드 추가하여 `cancelled.at()`처럼 사용하도록 수정

##

### 문제

```log
Error creating bean with name 'niceHttpClient': Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.NoSuchBeanDefinitionException: No qualifying bean of type 'org.springframework.web.reactive.function.client.WebClient' available: expected at least 1 bean which qualifies as autowire candidate. Dependency annotations: {@org.springframework.beans.factory.annotation.Qualifier("niceWebClient")}
org.springframework.beans.factory.UnsatisfiedDependencyException: Error creating bean with name 'niceHttpClient': Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.NoSuchBeanDefinitionException: No qualifying bean of type 'org.springframework.web.reactive.function.client.WebClient' available: expected at least 1 bean which qualifies as autowire candidate. Dependency annotations: {@org.springframework.beans.factory.annotation.Qualifier("niceWebClient")}
 at org.springframework.beans.factory.support.ConstructorResolver.createArgumentArray(ConstructorResolver.java:800)
 at org.springframework.beans.factory.support.ConstructorResolver.autowireConstructor(ConstructorResolver.java:229)
 at org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory.autowireConstructor(AbstractAutowireCapableBeanFactory.java:1372)
 at org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory.createBeanInstance(AbstractAutowireCapableBeanFactory.java:1222)
 at org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory.doCreateBean(AbstractAutowireCapableBeanFactory.java:582)
 at org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory.createBean(AbstractAutowireCapableBeanFactory.java:542)
 at org.springframework.beans.factory.support.AbstractBeanFactory.lambda$doGetBean$0(AbstractBeanFactory.java:335)
 at org.springframework.beans.factory.support.DefaultSingletonBeanRegistry.getSingleton(DefaultSingletonBeanRegistry.java:234)
 at org.springframework.beans.factory.support.AbstractBeanFactory.doGetBean(AbstractBeanFactory.java:333)
 at org.springframework.beans.factory.support.AbstractBeanFactory.getBean(AbstractBeanFactory.java:208)
 at org.springframework.beans.factory.support.DefaultListableBeanFactory.preInstantiateSingletons(DefaultListableBeanFactory.java:953)
 at org.springframework.context.support.AbstractApplicationContext.finishBeanFactoryInitialization(AbstractApplicationContext.java:918)
 at org.springframework.context.support.AbstractApplicationContext.refresh(AbstractApplicationContext.java:583)
 at org.springframework.boot.SpringApplication.refresh(SpringApplication.java:732)
 at org.springframework.boot.SpringApplication.refreshContext(SpringApplication.java:414)
 at org.springframework.boot.SpringApplication.run(SpringApplication.java:302)
 at org.springframework.boot.test.context.SpringBootContextLoader.loadContext(SpringBootContextLoader.java:136)
 at org.springframework.test.context.cache.DefaultCacheAwareContextLoaderDelegate.loadContextInternal(DefaultCacheAwareContextLoaderDelegate.java:99)
 at org.springframework.test.context.cache.DefaultCacheAwareContextLoaderDelegate.loadContext(DefaultCacheAwareContextLoaderDelegate.java:124)
 at org.springframework.test.context.support.DefaultTestContext.getApplicationContext(DefaultTestContext.java:124)
 at io.kotest.extensions.spring.SpringAutowireConstructorExtension.instantiate(SpringAutowireConstructorExtension.kt:27)
 at io.kotest.engine.spec.InstantiateSpecKt.createAndInitializeSpec(instantiateSpec.kt:30)
 at io.kotest.engine.spec.InstantiateSpecKt.instantiate(instantiateSpec.kt:11)
 at io.kotest.engine.spec.SpecRefKt.instance(SpecRef.kt:14)
 at io.kotest.engine.spec.SpecExecutor.createInstance-gIAlu-s(SpecExecutor.kt:127)
 at io.kotest.engine.spec.SpecExecutor.access$createInstance-gIAlu-s(SpecExecutor.kt:48)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$innerExecute$1.invokeSuspend(SpecExecutor.kt:85)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$innerExecute$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$innerExecute$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.FinalizeSpecInterceptor.intercept-0E7RQCE(FinalizeSpecInterceptor.kt:19)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.PrepareSpecInterceptor.intercept-0E7RQCE(PrepareSpecInterceptor.kt:20)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.ApplyExtensionsInterceptor.intercept-0E7RQCE(ApplyExtensionsInterceptor.kt:35)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.SpecFinishedInterceptor.intercept-0E7RQCE(SpecFinishedInterceptor.kt:19)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.SpecStartedInterceptor.intercept-0E7RQCE(SpecStartedInterceptor.kt:19)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.SpecRefExtensionInterceptor$intercept$inner$1.invokeSuspend(SpecRefExtensionInterceptor.kt:24)
 at io.kotest.engine.spec.interceptor.SpecRefExtensionInterceptor$intercept$inner$1.invoke(SpecRefExtensionInterceptor.kt)
 at io.kotest.engine.spec.interceptor.SpecRefExtensionInterceptor$intercept$inner$1.invoke(SpecRefExtensionInterceptor.kt)
 at io.kotest.engine.spec.interceptor.SpecRefExtensionInterceptor.intercept-0E7RQCE(SpecRefExtensionInterceptor.kt:27)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.RequiresTagSpecInterceptor.intercept-0E7RQCE(RequiresTagSpecInterceptor.kt:35)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.TagsExcludedSpecInterceptor.intercept-0E7RQCE(TagsExcludedDiscoveryExtension.kt:32)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.SystemPropertySpecFilterInterceptor.intercept-0E7RQCE(SystemPropertySpecFilterInterceptor.kt:49)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.SpecFilterInterceptor.intercept-0E7RQCE(SpecFilterInterceptor.kt:37)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.IgnoredSpecInterceptor.intercept-0E7RQCE(IgnoredSpecInterceptor.kt:42)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.interceptor.EnabledIfSpecInterceptor.intercept-0E7RQCE(EnabledIfSpecInterceptor.kt:40)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invokeSuspend(SpecExecutor.kt:90)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor$referenceInterceptors$3$1.invoke(SpecExecutor.kt)
 at io.kotest.engine.spec.SpecExecutor.referenceInterceptors(SpecExecutor.kt:91)
 at io.kotest.engine.spec.SpecExecutor.execute(SpecExecutor.kt:60)
 at io.kotest.engine.ConcurrentTestSuiteScheduler$schedule$8$1$2.invokeSuspend(ConcurrentTestSuiteScheduler.kt:71)
 at kotlin.coroutines.jvm.internal.BaseContinuationImpl.resumeWith(ContinuationImpl.kt:33)
 at kotlinx.coroutines.DispatchedTask.run(DispatchedTask.kt:106)
 at kotlinx.coroutines.EventLoopImplBase.processNextEvent(EventLoop.common.kt:279)
 at kotlinx.coroutines.BlockingCoroutine.joinBlocking(Builders.kt:85)
 at kotlinx.coroutines.BuildersKt__BuildersKt.runBlocking(Builders.kt:59)
 at kotlinx.coroutines.BuildersKt.runBlocking(Unknown Source)
 at kotlinx.coroutines.BuildersKt__BuildersKt.runBlocking$default(Builders.kt:38)
 at kotlinx.coroutines.BuildersKt.runBlocking$default(Unknown Source)
 at io.kotest.common.RunBlockingKt.runBlocking(runBlocking.kt:3)
 at io.kotest.engine.launcher.MainKt.main(main.kt:34)
Caused by: org.springframework.beans.factory.NoSuchBeanDefinitionException: No qualifying bean of type 'org.springframework.web.reactive.function.client.WebClient' available: expected at least 1 bean which qualifies as autowire candidate. Dependency annotations: {@org.springframework.beans.factory.annotation.Qualifier("niceWebClient")}
 at org.springframework.beans.factory.support.DefaultListableBeanFactory.raiseNoMatchingBeanFound(DefaultListableBeanFactory.java:1799)
 at org.springframework.beans.factory.support.DefaultListableBeanFactory.doResolveDependency(DefaultListableBeanFactory.java:1355)
 at org.springframework.beans.factory.support.DefaultListableBeanFactory.resolveDependency(DefaultListableBeanFactory.java:1309)
 at org.springframework.beans.factory.support.ConstructorResolver.resolveAutowiredArgument(ConstructorResolver.java:887)
 at org.springframework.beans.factory.support.ConstructorResolver.createArgumentArray(ConstructorResolver.java:791)
 ... 92 more
```

### 원인

- 의존성 주입이 제대로 이뤄지지 않았기 때문으로 보인다

```kotlin
@SpringBootTest(
    classes = [
        NiceHttpClient::class
    ]
)
class NiceHttpClientTest(
    @Qualifier("niceHttpClient")
    val niceClient: NiceHttpClient
) : FreeSpec({
```

### 해결

- `@SpringBootTest` 어노테이션에 필요한 클래스 추가하여 주입하도록 수정

```kotlin
@SpringBootTest(
    classes = [
        NiceHttpClient::class,
        NiceConfiguration::class, // `NiceHttpClient` 클래스에 `WebClient` 주입
        NiceResponseValidator::class // `NiceHttpClient` 클래스에 `NiceResponseValidator` 주입
    ]
)
class NiceHttpClientTest(
    @Qualifier("niceHttpClient")
    val niceClient: NiceClient // 굳이 구현 클래스일 필요 없다
) : FreeSpec({
```

## Could not read property private final ... from column

### 문제

```log
Exception Stack Trace:
org.springframework.data.mapping.MappingException: Could not read property private final finance.chai.gateway.transaction.model.nice.NicePaymentMethod finance.chai.gateway.transaction.repository.SpringDataNiceTransactionRepository$Companion$NiceTransactionRow.paymentMethod from column payment_method!
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:187)
 Suppressed: The stacktrace has been enhanced by Reactor, refer to additional information below: 
Assembly trace from producer [reactor.core.publisher.FluxLiftFuseable] :
 reactor.core.publisher.Flux.handle
 io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
Error has been observed at the following site(s):
 *______Flux.handle ⇢ at io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
 *_____Flux.flatMap ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.lambda$all$2(DefaultFetchSpec.java:89)
 *___Flux.usingWhen ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:134)
 |_ Flux.onErrorMap ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:146)
 |_                 ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.all(DefaultFetchSpec.java:87)
 |_     Flux.buffer ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:66)
 |_    Flux.flatMap ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:67)
 |_       Flux.next ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:77)
 |_    Mono.flatMap ⇢ at org.springframework.data.r2dbc.core.R2dbcEntityTemplate.doSelect(R2dbcEntityTemplate.java:403)
 |_                 ⇢ at org.springframework.data.r2dbc.core.ReactiveSelectOperationSupport$ReactiveSelectSupport.one(ReactiveSelectOperationSupport.java:140)
 |_      Mono.retry ⇢ at finance.chai.gateway.transaction.repository.SpringDataNiceTransactionRepository.find$suspendImpl(SpringDataNiceTransactionRepository.kt:101)
Original Stack Trace:
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:187)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.access$000(MappingR2dbcConverter.java:66)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter$RowParameterValueProvider.getParameterValue(MappingR2dbcConverter.java:718)
  at org.springframework.data.mapping.model.SpELExpressionParameterValueProvider.getParameterValue(SpELExpressionParameterValueProvider.java:53)
  at org.springframework.data.relational.core.conversion.BasicRelationalConverter$ConvertingParameterValueProvider.getParameterValue(BasicRelationalConverter.java:291)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.extractInvocationArguments(ClassGeneratingEntityInstantiator.java:276)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator$EntityInstantiatorAdapter.createInstance(ClassGeneratingEntityInstantiator.java:248)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.createInstance(ClassGeneratingEntityInstantiator.java:89)
  at org.springframework.data.relational.core.conversion.BasicRelationalConverter.createInstance(BasicRelationalConverter.java:148)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.createInstance(MappingR2dbcConverter.java:329)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:126)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:121)
  at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:46)
  at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:29)
  at io.r2dbc.postgresql.PostgresqlResult.lambda$map$2(PostgresqlResult.java:124)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:169)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.MonoFlatMapMany$FlatMapManyInner.onNext(MonoFlatMapMany.java:250)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxFilterFuseable$FilterFuseableSubscriber.onNext(FluxFilterFuseable.java:118)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxContextWrite$ContextWriteSubscriber.onNext(FluxContextWrite.java:107)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at io.r2dbc.postgresql.util.FluxDiscardOnCancel$FluxDiscardOnCancelSubscriber.onNext(FluxDiscardOnCancel.java:86)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxDoFinally$DoFinallySubscriber.onNext(FluxDoFinally.java:130)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxCreate$BufferAsyncSink.drain(FluxCreate.java:793)
  at reactor.core.publisher.FluxCreate$BufferAsyncSink.next(FluxCreate.java:718)
  at reactor.core.publisher.FluxCreate$SerializedFluxSink.next(FluxCreate.java:154)
  at io.r2dbc.postgresql.client.ReactorNettyClient$Conversation.emit(ReactorNettyClient.java:654)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.emit(ReactorNettyClient.java:906)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:780)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:686)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
  at reactor.netty.channel.FluxReceive.drainReceiver(FluxReceive.java:279)
  at reactor.netty.channel.FluxReceive.onInboundNext(FluxReceive.java:388)
  at reactor.netty.channel.ChannelOperations.onInboundNext(ChannelOperations.java:404)
  at reactor.netty.channel.ChannelOperationsHandler.channelRead(ChannelOperationsHandler.java:93)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
  at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:327)
  at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:314)
  at io.netty.handler.codec.ByteToMessageDecoder.callDecode(ByteToMessageDecoder.java:435)
  at io.netty.handler.codec.ByteToMessageDecoder.channelRead(ByteToMessageDecoder.java:279)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
  at io.netty.channel.DefaultChannelPipeline$HeadContext.channelRead(DefaultChannelPipeline.java:1410)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.DefaultChannelPipeline.fireChannelRead(DefaultChannelPipeline.java:919)
  at io.netty.channel.nio.AbstractNioByteChannel$NioByteUnsafe.read(AbstractNioByteChannel.java:166)
  at io.netty.channel.nio.NioEventLoop.processSelectedKey(NioEventLoop.java:722)
  at io.netty.channel.nio.NioEventLoop.processSelectedKeysOptimized(NioEventLoop.java:658)
  at io.netty.channel.nio.NioEventLoop.processSelectedKeys(NioEventLoop.java:584)
  at io.netty.channel.nio.NioEventLoop.run(NioEventLoop.java:496)
  at io.netty.util.concurrent.SingleThreadEventExecutor$4.run(SingleThreadEventExecutor.java:986)
  at io.netty.util.internal.ThreadExecutorMap$2.run(ThreadExecutorMap.java:74)
  at io.netty.util.concurrent.FastThreadLocalRunnable.run(FastThreadLocalRunnable.java:30)
  at java.base/java.lang.Thread.run(Thread.java:833)
Caused by: java.lang.IllegalArgumentException: No enum constant finance.chai.gateway.transaction.model.nice.NicePaymentMethod.CARD
 at java.base/java.lang.Enum.valueOf(Enum.java:273)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.getPotentiallyConvertedSimpleRead(MappingR2dbcConverter.java:277)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readValue(MappingR2dbcConverter.java:204)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:184)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.access$000(MappingR2dbcConverter.java:66)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter$RowParameterValueProvider.getParameterValue(MappingR2dbcConverter.java:718)
 at org.springframework.data.mapping.model.SpELExpressionParameterValueProvider.getParameterValue(SpELExpressionParameterValueProvider.java:53)
 at org.springframework.data.relational.core.conversion.BasicRelationalConverter$ConvertingParameterValueProvider.getParameterValue(BasicRelationalConverter.java:291)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.extractInvocationArguments(ClassGeneratingEntityInstantiator.java:276)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator$EntityInstantiatorAdapter.createInstance(ClassGeneratingEntityInstantiator.java:248)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.createInstance(ClassGeneratingEntityInstantiator.java:89)
 at org.springframework.data.relational.core.conversion.BasicRelationalConverter.createInstance(BasicRelationalConverter.java:148)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.createInstance(MappingR2dbcConverter.java:329)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:126)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:121)
 at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:46)
 at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:29)
 at io.r2dbc.postgresql.PostgresqlResult.lambda$map$2(PostgresqlResult.java:124)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:169)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.MonoFlatMapMany$FlatMapManyInner.onNext(MonoFlatMapMany.java:250)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxFilterFuseable$FilterFuseableSubscriber.onNext(FluxFilterFuseable.java:118)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxContextWrite$ContextWriteSubscriber.onNext(FluxContextWrite.java:107)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at io.r2dbc.postgresql.util.FluxDiscardOnCancel$FluxDiscardOnCancelSubscriber.onNext(FluxDiscardOnCancel.java:86)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxDoFinally$DoFinallySubscriber.onNext(FluxDoFinally.java:130)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxCreate$BufferAsyncSink.drain(FluxCreate.java:793)
 at reactor.core.publisher.FluxCreate$BufferAsyncSink.next(FluxCreate.java:718)
 at reactor.core.publisher.FluxCreate$SerializedFluxSink.next(FluxCreate.java:154)
 at io.r2dbc.postgresql.client.ReactorNettyClient$Conversation.emit(ReactorNettyClient.java:654)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.emit(ReactorNettyClient.java:906)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:780)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:686)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.netty.channel.FluxReceive.drainReceiver(FluxReceive.java:279)
 at reactor.netty.channel.FluxReceive.onInboundNext(FluxReceive.java:388)
 at reactor.netty.channel.ChannelOperations.onInboundNext(ChannelOperations.java:404)
 at reactor.netty.channel.ChannelOperationsHandler.channelRead(ChannelOperationsHandler.java:93)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
 at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:327)
 at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:314)
 at io.netty.handler.codec.ByteToMessageDecoder.callDecode(ByteToMessageDecoder.java:435)
 at io.netty.handler.codec.ByteToMessageDecoder.channelRead(ByteToMessageDecoder.java:279)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
 at io.netty.channel.DefaultChannelPipeline$HeadContext.channelRead(DefaultChannelPipeline.java:1410)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.DefaultChannelPipeline.fireChannelRead(DefaultChannelPipeline.java:919)
 at io.netty.channel.nio.AbstractNioByteChannel$NioByteUnsafe.read(AbstractNioByteChannel.java:166)
 at io.netty.channel.nio.NioEventLoop.processSelectedKey(NioEventLoop.java:722)
 at io.netty.channel.nio.NioEventLoop.processSelectedKeysOptimized(NioEventLoop.java:658)
 at io.netty.channel.nio.NioEventLoop.processSelectedKeys(NioEventLoop.java:584)
 at io.netty.channel.nio.NioEventLoop.run(NioEventLoop.java:496)
 at io.netty.util.concurrent.SingleThreadEventExecutor$4.run(SingleThreadEventExecutor.java:986)
 at io.netty.util.internal.ThreadExecutorMap$2.run(ThreadExecutorMap.java:74)
 at io.netty.util.concurrent.FastThreadLocalRunnable.run(FastThreadLocalRunnable.java:30)
 at java.base/java.lang.Thread.run(Thread.java:833)
",
  "context": {},
  "etc": {
    "stack_trace": "java.lang.IllegalArgumentException: No enum constant finance.chai.gateway.transaction.model.nice.NicePaymentMethod.CARD
 at java.lang.Enum.valueOf(Enum.java:273)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.getPotentiallyConvertedSimpleRead(MappingR2dbcConverter.java:277)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readValue(MappingR2dbcConverter.java:204)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:184)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.access$000(MappingR2dbcConverter.java:66)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter$RowParameterValueProvider.getParameterValue(MappingR2dbcConverter.java:718)
 at org.springframework.data.mapping.model.SpELExpressionParameterValueProvider.getParameterValue(SpELExpressionParameterValueProvider.java:53)
 at org.springframework.data.relational.core.conversion.BasicRelationalConverter$ConvertingParameterValueProvider.getParameterValue(BasicRelationalConverter.java:291)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.extractInvocationArguments(ClassGeneratingEntityInstantiator.java:276)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator$EntityInstantiatorAdapter.createInstance(ClassGeneratingEntityInstantiator.java:248)
 at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.createInstance(ClassGeneratingEntityInstantiator.java:89)
 at org.springframework.data.relational.core.conversion.BasicRelationalConverter.createInstance(BasicRelationalConverter.java:148)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.createInstance(MappingR2dbcConverter.java:329)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:126)
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:121)
 at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:46)
 at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:29)
 at io.r2dbc.postgresql.PostgresqlResult.lambda$map$2(PostgresqlResult.java:124)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:169)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.MonoFlatMapMany$FlatMapManyInner.onNext(MonoFlatMapMany.java:250)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxFilterFuseable$FilterFuseableSubscriber.onNext(FluxFilterFuseable.java:118)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxContextWrite$ContextWriteSubscriber.onNext(FluxContextWrite.java:107)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at io.r2dbc.postgresql.util.FluxDiscardOnCancel$FluxDiscardOnCancelSubscriber.onNext(FluxDiscardOnCancel.java:86)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxDoFinally$DoFinallySubscriber.onNext(FluxDoFinally.java:130)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxCreate$BufferAsyncSink.drain(FluxCreate.java:793)
 at reactor.core.publisher.FluxCreate$BufferAsyncSink.next(FluxCreate.java:718)
 at reactor.core.publisher.FluxCreate$SerializedFluxSink.next(FluxCreate.java:154)
 at io.r2dbc.postgresql.client.ReactorNettyClient$Conversation.emit(ReactorNettyClient.java:654)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.emit(ReactorNettyClient.java:906)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:780)
 at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:686)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
 at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
 at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.core.publisher.FluxOnAssembly$OnAssemblySubscriber.onNext(FluxOnAssembly.java:539)
 at reactor.netty.channel.FluxReceive.drainReceiver(FluxReceive.java:279)
 at reactor.netty.channel.FluxReceive.onInboundNext(FluxReceive.java:388)
 at reactor.netty.channel.ChannelOperations.onInboundNext(ChannelOperations.java:404)
 at reactor.netty.channel.ChannelOperationsHandler.channelRead(ChannelOperationsHandler.java:93)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
 at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:327)
 at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:314)
 at io.netty.handler.codec.ByteToMessageDecoder.callDecode(ByteToMessageDecoder.java:435)
 at io.netty.handler.codec.ByteToMessageDecoder.channelRead(ByteToMessageDecoder.java:279)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
 at io.netty.channel.DefaultChannelPipeline$HeadContext.channelRead(DefaultChannelPipeline.java:1410)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
 at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
 at io.netty.channel.DefaultChannelPipeline.fireChannelRead(DefaultChannelPipeline.java:919)
 at io.netty.channel.nio.AbstractNioByteChannel$NioByteUnsafe.read(AbstractNioByteChannel.java:166)
 at io.netty.channel.nio.NioEventLoop.processSelectedKey(NioEventLoop.java:722)
 at io.netty.channel.nio.NioEventLoop.processSelectedKeysOptimized(NioEventLoop.java:658)
 at io.netty.channel.nio.NioEventLoop.processSelectedKeys(NioEventLoop.java:584)
 at io.netty.channel.nio.NioEventLoop.run(NioEventLoop.java:496)
 at io.netty.util.concurrent.SingleThreadEventExecutor$4.run(SingleThreadEventExecutor.java:986)
 at io.netty.util.internal.ThreadExecutorMap$2.run(ThreadExecutorMap.java:74)
 at io.netty.util.concurrent.FastThreadLocalRunnable.run(FastThreadLocalRunnable.java:30)
 at java.lang.Thread.run(Thread.java:833)
Wrapped by: org.springframework.data.mapping.MappingException: Could not read property private final finance.chai.gateway.transaction.model.nice.NicePaymentMethod finance.chai.gateway.transaction.repository.SpringDataNiceTransactionRepository$Companion$NiceTransactionRow.paymentMethod from column payment_method!
 at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:187)
 Suppressed: reactor.core.publisher.FluxOnAssembly$OnAssemblyException: 
Assembly trace from producer [reactor.core.publisher.FluxLiftFuseable] :
 reactor.core.publisher.Flux.handle
 io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
Error has been observed at the following site(s):
 *______Flux.handle ⇢ at io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
 *_____Flux.flatMap ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.lambda$all$2(DefaultFetchSpec.java:89)
 *___Flux.usingWhen ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:134)
 |_ Flux.onErrorMap ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:146)
 |_                 ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.all(DefaultFetchSpec.java:87)
 |_     Flux.buffer ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:66)
 |_    Flux.flatMap ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:67)
 |_       Flux.next ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.one(DefaultFetchSpec.java:77)
 |_    Mono.flatMap ⇢ at org.springframework.data.r2dbc.core.R2dbcEntityTemplate.doSelect(R2dbcEntityTemplate.java:403)
 |_                 ⇢ at org.springframework.data.r2dbc.core.ReactiveSelectOperationSupport$ReactiveSelectSupport.one(ReactiveSelectOperationSupport.java:140)
 |_      Mono.retry ⇢ at finance.chai.gateway.transaction.repository.SpringDataNiceTransactionRepository.find$suspendImpl(SpringDataNiceTransactionRepository.kt:101)
Original Stack Trace:
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.readFrom(MappingR2dbcConverter.java:187)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.access$000(MappingR2dbcConverter.java:66)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter$RowParameterValueProvider.getParameterValue(MappingR2dbcConverter.java:718)
  at org.springframework.data.mapping.model.SpELExpressionParameterValueProvider.getParameterValue(SpELExpressionParameterValueProvider.java:53)
  at org.springframework.data.relational.core.conversion.BasicRelationalConverter$ConvertingParameterValueProvider.getParameterValue(BasicRelationalConverter.java:291)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.extractInvocationArguments(ClassGeneratingEntityInstantiator.java:276)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator$EntityInstantiatorAdapter.createInstance(ClassGeneratingEntityInstantiator.java:248)
  at org.springframework.data.mapping.model.ClassGeneratingEntityInstantiator.createInstance(ClassGeneratingEntityInstantiator.java:89)
  at org.springframework.data.relational.core.conversion.BasicRelationalConverter.createInstance(BasicRelationalConverter.java:148)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.createInstance(MappingR2dbcConverter.java:329)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:126)
  at org.springframework.data.r2dbc.convert.MappingR2dbcConverter.read(MappingR2dbcConverter.java:121)
  at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:46)
  at org.springframework.data.r2dbc.convert.EntityRowMapper.apply(EntityRowMapper.java:29)
  at io.r2dbc.postgresql.PostgresqlResult.lambda$map$2(PostgresqlResult.java:124)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:169)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.MonoFlatMapMany$FlatMapManyInner.onNext(MonoFlatMapMany.java:250)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxFilterFuseable$FilterFuseableSubscriber.onNext(FluxFilterFuseable.java:118)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxContextWrite$ContextWriteSubscriber.onNext(FluxContextWrite.java:107)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at io.r2dbc.postgresql.util.FluxDiscardOnCancel$FluxDiscardOnCancelSubscriber.onNext(FluxDiscardOnCancel.java:86)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxDoFinally$DoFinallySubscriber.onNext(FluxDoFinally.java:130)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxCreate$BufferAsyncSink.drain(FluxCreate.java:793)
  at reactor.core.publisher.FluxCreate$BufferAsyncSink.next(FluxCreate.java:718)
  at reactor.core.publisher.FluxCreate$SerializedFluxSink.next(FluxCreate.java:154)
  at io.r2dbc.postgresql.client.ReactorNettyClient$Conversation.emit(ReactorNettyClient.java:654)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.emit(ReactorNettyClient.java:906)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:780)
  at io.r2dbc.postgresql.client.ReactorNettyClient$BackendMessageSubscriber.onNext(ReactorNettyClient.java:686)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxHandleFuseable$HandleFuseableSubscriber.onNext(FluxHandleFuseable.java:184)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxPeekFuseable$PeekFuseableSubscriber.onNext(FluxPeekFuseable.java:210)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
  at finance.chai.gateway.transaction.configuration.MdcContextLifter.onNext(MdcContextLifterConfiguration.kt:40)
  at reactor.core.publisher.FluxHide$SuppressFuseableSubscriber.onNext(FluxHide.java:137)
  at reactor.core.publisher.FluxMapFuseable$MapFuseableSubscriber.onNext(FluxMapFuseable.java:127)
  at reactor.netty.channel.FluxReceive.drainReceiver(FluxReceive.java:279)
  at reactor.netty.channel.FluxReceive.onInboundNext(FluxReceive.java:388)
  at reactor.netty.channel.ChannelOperations.onInboundNext(ChannelOperations.java:404)
  at reactor.netty.channel.ChannelOperationsHandler.channelRead(ChannelOperationsHandler.java:93)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
  at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:327)
  at io.netty.handler.codec.ByteToMessageDecoder.fireChannelRead(ByteToMessageDecoder.java:314)
  at io.netty.handler.codec.ByteToMessageDecoder.callDecode(ByteToMessageDecoder.java:435)
  at io.netty.handler.codec.ByteToMessageDecoder.channelRead(ByteToMessageDecoder.java:279)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.AbstractChannelHandlerContext.fireChannelRead(AbstractChannelHandlerContext.java:357)
  at io.netty.channel.DefaultChannelPipeline$HeadContext.channelRead(DefaultChannelPipeline.java:1410)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:379)
  at io.netty.channel.AbstractChannelHandlerContext.invokeChannelRead(AbstractChannelHandlerContext.java:365)
  at io.netty.channel.DefaultChannelPipeline.fireChannelRead(DefaultChannelPipeline.java:919)
  at io.netty.channel.nio.AbstractNioByteChannel$NioByteUnsafe.read(AbstractNioByteChannel.java:166)
  at io.netty.channel.nio.NioEventLoop.processSelectedKey(NioEventLoop.java:722)
  at io.netty.channel.nio.NioEventLoop.processSelectedKeysOptimized(NioEventLoop.java:658)
  at io.netty.channel.nio.NioEventLoop.processSelectedKeys(NioEventLoop.java:584)
  at io.netty.channel.nio.NioEventLoop.run(NioEventLoop.java:496)
  at io.netty.util.concurrent.SingleThreadEventExecutor$4.run(SingleThreadEventExecutor.java:986)
  at io.netty.util.internal.ThreadExecutorMap$2.run(ThreadExecutorMap.java:74)
  at io.netty.util.concurrent.FastThreadLocalRunnable.run(FastThreadLocalRunnable.java:30)
  at java.lang.Thread.run(Thread.java:833)
```

### 원인

- 칼럼에 저장된 `"CARD"`와 enum 이름인 `"PAYMETHOD_CARD"`가 상이해서 매핑 실패

### 해결

- DB 데이터 보정

## Non exhaustive 'when' statements on enum will be prohibited in

### 문제

```kotlin
command.virtualAccountForm.cashReceipt?.let {
    when (it.type) {
    ^^^^ 
    // Non exhaustive 'when' statements on enum will be prohibited in 1.7, 
    // add 'RECEIPT_TYPE_UNSPECIFIED', 'RECEIPT_TYPE_ANONYMOUS' branches 
    // or 'else' branch instead
        ReceiptType.RECEIPT_TYPE_PERSONAL -> {
            cashReceiptType = "1"
            receiptTypeNo = it.registrationNumber
        }
        ReceiptType.RECEIPT_TYPE_CORPORATE -> {
            cashReceiptType = "2"
            receiptTypeNo = it.businessNumber
        }
    }
}
```

### 원인

- [`Exhaustive When Statement`](https://okky.kr/articles/1200915)
    - 파일 단계에서 내가 사용한 변수에 대한 모든 경우의 수를 핸들링 했는지 검사해주는 것

### 해결

- [`else` 브랜치 추가하거나, 모든 경우의 수를 추가](https://stackoverflow.com/a/52813225)
