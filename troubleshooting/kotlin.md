# Kotlin troubleshooting

- [Kotlin troubleshooting](#kotlin-troubleshooting)
    - [Accessing non-final property \<VARIABLE\_NAME\> in constructor](#accessing-non-final-property-variable_name-in-constructor)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - ['when' expression must be exhaustive, add necessary 'is Nicepay' branch or 'else' branch instead](#when-expression-must-be-exhaustive-add-necessary-is-nicepay-branch-or-else-branch-instead)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [No tests found for given includes: ... (--tests filter)](#no-tests-found-for-given-includes----tests-filter)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [Could not resolve org.jetbrains.kotlin:kotlin-reflect:1.2.71. Required by: subproject:test](#could-not-resolve-orgjetbrainskotlinkotlin-reflect1271-required-by-subprojecttest)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [APPLICATION FAILED TO START, required a bean of type something that could not be found](#application-failed-to-start-required-a-bean-of-type-something-that-could-not-be-found)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [executeMany; SQL \[QUERY\] current transaction is aborted, commands ignored until end of transaction block](#executemany-sql-query-current-transaction-is-aborted-commands-ignored-until-end-of-transaction-block)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [Incremental compilation analysis failed](#incremental-compilation-analysis-failed)
        - [문제](#문제-6)
    - [flyway ERROR: relation ... already exists](#flyway-error-relation--already-exists)
        - [문제](#문제-7)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - [MockKException: no answer found for](#mockkexception-no-answer-found-for)
        - [문제](#문제-8)
        - [원인](#원인-7)
        - [해결](#해결-7)
    - [kover - Unresolved reference: excludes](#kover---unresolved-reference-excludes)
        - [문제](#문제-9)
        - [원인](#원인-8)
        - [해결](#해결-8)

## Accessing non-final property <VARIABLE_NAME> in constructor

- [Why does this fix the “accessing non-final property in constructor” warning?](https://discuss.kotlinlang.org/t/why-does-this-fix-the-accessing-non-final-property-in-constructor-warning/9782/3)
- [Delegated properties](https://kotlinlang.org/docs/delegated-properties.html)

### 문제

### 원인

### 해결

```diff
class Obj {
- val charsetName = "euc-kr"
+ private final val charsetName = "euc-kr"
```

## 'when' expression must be exhaustive, add necessary 'is Nicepay' branch or 'else' branch instead

### 문제

```kotlin
when(channel) {
^^^^ 에러 발생
    is Channel.Ksnet -> do something
    is Channel.TossPayments -> do something
}
```

### 원인

- Channel에 Nicepay라는 타입을 새로 추가
- [`when`과 `exhaustive`](https://kotlinlang.org/spec/expressions.html#exhaustive-when-expressions)

> The type of the resulting when expression is the [least upper bound](https://kotlinlang.org/spec/type-system.html#least-upper-bound) of the types of all its entries. If when expression is not exhaustive, it has type kotlin.Unit and may be used only as a statement.

- [Exhaustiveness of When Statements](https://java-to-kotlin.dev/route_changes/exhaustiveness-of-when-statements.html)

### 해결

```kotlin
when(channel) {
    is Channel.Ksnet -> do something
    is Channel.TossPayments -> do something
    is Channel.Nicepay -> do something // 추가
}
```

## No tests found for given includes: ... (--tests filter)

### 문제

intellij 통해서 gradle run configuration으로 실행 시에 에러 발생

### 원인

- [No tests found for given includes Error, when running Parameterized Unit test in Android Studio](https://stackoverflow.com/questions/30474767/no-tests-found-for-given-includes-error-when-running-parameterized-unit-test-in)
- [How do I use the native JUnit 5 support in Gradle with the Kotlin DSL?](https://stackoverflow.com/questions/50128728/how-do-i-use-the-native-junit-5-support-in-gradle-with-the-kotlin-dsl/50128729#50128729)

JUnit4 버전에서 발생하는 이슈다, `useJunitPlatform`을 `build.gradle`에 추가하지 않아서 발생하는 문제다 등등 여러 얘기가 있지만,
제안된 해결 방안으로는 해결할 수 없었다.

### 해결

- [[IntelliJ] JUnit4 + Gradle - update broke all tests with error: no tests found for given includes](https://linked2ev.github.io/devsub/2019/09/30/Intellij-junit4-gradle-issue/)

intellij 설정에서 테스트 시 intellij idea 사용하도록 하면 정상적으로 실행이 됐다
별개로 run configuration에서 kotest로 테스트 실행하면 정상적으로 실행이 됐다.

intellij 통해서 gradle 실행 시에 발생하는 문제가 아닐까 싶다.

## Could not resolve org.jetbrains.kotlin:kotlin-reflect:1.2.71. Required by: subproject:test

### 문제

`:subproject:test` task 없음에도 계속 여기서 필요하다고 나옴

### 원인

알 수 없음

### 해결

- IntelliJ -> 설정 -> build -> gradle -> Intellij 사용

## APPLICATION FAILED TO START, required a bean of type something that could not be found

### 문제

```log
Exception encountered during context initialization - cancelling refresh attempt: org.springframework.beans.factory.UnsatisfiedDependencyException: Error creating bean with name 'apiTransactionGrpcService' defined in file [/Users/rody/IdeaProjects/some-project/subproject/presentation/build/classes/kotlin/main/service/package/name/transaction/grpc/api/ApiTransactionGrpcService.class]: Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.UnsatisfiedDependencyException: Error creating bean with name 'cancelTransactionService' defined in URL [jar:file:/Users/rody/IdeaProjects/some-project/subproject/application/build/libs/application-0.0.1-SNAPSHOT-plain.jar!/service/package/name/transaction/cancel/CancelTransactionService.class]: Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.NoSuchBeanDefinitionException: No qualifying bean of type 'finance.chai.gateway.transaction.entity.ChannelRepository' available: expected at least 1 bean which qualifies as autowire candidate. Dependency annotations: {}
```

```log
Exception encountered during context initialization - cancelling refresh attempt: org.springframework.beans.factory.UnsatisfiedDependencyException: Error creating bean with name 'apiTransactionGrpcService' defined in file [/Users/rody/IdeaProjects/some-project/subproject/presentation/build/classes/kotlin/main/service/package/name/transaction/grpc/api/ApiTransactionGrpcService.class]: Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.UnsatisfiedDependencyException: Error creating bean with name 'cancelTransactionService' defined in URL [jar:file:/Users/rody/IdeaProjects/some-project/subproject/application/build/libs/application-0.0.1-SNAPSHOT-plain.jar!/service/package/name/transaction/cancel/CancelTransactionService.class]: Unsatisfied dependency expressed through constructor parameter 0; nested exception is org.springframework.beans.factory.NoSuchBeanDefinitionException: No qualifying bean of type 'finance.chai.gateway.transaction.entity.ChannelRepository' available: expected at least 1 bean which qualifies as autowire candidate. Dependency annotations: {}
```

```log
***************************
APPLICATION FAILED TO START
***************************

Description:

Parameter 0 of constructor in finance.chai.gateway.transaction.cancel.CancelTransactionService required a bean of type 'finance.chai.gateway.transaction.entity.ChannelRepository' that could not be found.


Action:

Consider defining a bean of type 'finance.chai.gateway.transaction.entity.ChannelRepository' in your configuration.
```

### 원인

bean으로 주입되어야 하는 implementation이 infrastructure 레이어에 있는데, `build.gradle.kts`의 dependencies에서 누락되어 있었음.
그래서 의존성을 찾을 수가 없어서 익셉션 발생.

### 해결

dependencies에 의존성 추가

```kts
implementation(project("subproject:infrastructure"))
```

## executeMany; SQL [QUERY] current transaction is aborted, commands ignored until end of transaction block

### 문제

```log
Executing SQL statement [SELECT nicepay_transaction.transaction_id FROM nicepay_transaction WHERE nicepay_transaction.transaction_id = $1 AND nicepay_transaction.deleted = $2 LIMIT 1]


Error: SEVERITY_LOCALIZED=ERROR, SEVERITY_NON_LOCALIZED=ERROR, CODE=42P01, MESSAGE=relation \"nicepay_transaction\" does not exist, POSITION=48, FILE=parse_relation.c, LINE=1392, ROUTINE=parserOpenTable


Executing SQL statement [SELECT nicepay_transaction.transaction_id FROM nicepay_transaction WHERE nicepay_transaction.transaction_id = $1 AND nicepay_transaction.deleted = $2 LIMIT 1]


Error: SEVERITY_LOCALIZED=ERROR, SEVERITY_NON_LOCALIZED=ERROR, CODE=25P02, MESSAGE=current transaction is aborted, commands ignored until end of transaction block, FILE=postgres.c, LINE=1470, ROUTINE=exec_parse_message


Initiating transaction rollback


io.r2dbc.postgresql.ExceptionFactory$PostgresqlNonTransientResourceException: current transaction is aborted, commands ignored until end of transaction block
 at io.r2dbc.postgresql.ExceptionFactory.createException(ExceptionFactory.java:99)
 Suppressed: reactor.core.publisher.FluxOnAssembly$OnAssemblyException:
Assembly trace from producer [reactor.core.publisher.FluxLiftFuseable] :
 reactor.core.publisher.Flux.handle
 io.r2dbc.postgresql.ExtendedFlowDelegate.runQuery(ExtendedFlowDelegate.java:127)
Error has been observed at the following site(s):
 *_______Flux.handle ⇢ at io.r2dbc.postgresql.ExtendedFlowDelegate.runQuery(ExtendedFlowDelegate.java:127)
 *__Mono.flatMapMany ⇢ at io.r2dbc.postgresql.ExtendedQueryPostgresqlStatement.lambda$execute$6(ExtendedQueryPostgresqlStatement.java:196)
 |_      Flux.handle ⇢ at io.r2dbc.postgresql.PostgresqlResult.map(PostgresqlResult.java:107)
 *______Flux.flatMap ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.lambda$all$2(DefaultFetchSpec.java:89)
 *____Flux.usingWhen ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:134)
Original Stack Trace: .. 생략 ..


Wrapped by: org.springframework.dao.DataAccessResourceFailureException: executeMany; SQL [SELECT nicepay_transaction.transaction_id FROM nicepay_transaction WHERE nicepay_transaction.transaction_id = $1 AND nicepay_transaction.deleted = $2 LIMIT 1]; current transaction is aborted, commands ignored until end of transaction block; nested exception is io.r2dbc.postgresql.ExceptionFactory$PostgresqlNonTransientResourceException: [25P02] current transaction is aborted, commands ignored until end of transaction block
 at org.springframework.r2dbc.connection.ConnectionFactoryUtils.convertR2dbcException(ConnectionFactoryUtils.java:226)
 Suppressed: reactor.core.publisher.FluxOnAssembly$OnAssemblyException:
Assembly trace from producer [reactor.core.publisher.FluxLift] :
 reactor.core.publisher.Flux.onErrorMap
 org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:146)
Error has been observed at the following site(s):
 *____Flux.onErrorMap ⇢ at org.springframework.r2dbc.core.DefaultDatabaseClient.inConnectionMany(DefaultDatabaseClient.java:146)
 |_                   ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.all(DefaultFetchSpec.java:87)
 |_         Flux.next ⇢ at org.springframework.r2dbc.core.DefaultFetchSpec.first(DefaultFetchSpec.java:82)
 |_   Mono.hasElement ⇢ at org.springframework.data.r2dbc.core.R2dbcEntityTemplate.doExists(R2dbcEntityTemplate.java:384)
 |_                   ⇢ at org.springframework.data.r2dbc.core.R2dbcEntityTemplate.exists(R2dbcEntityTemplate.java:357)
 |_        Mono.retry ⇢ at finance.chai.gateway.transaction.nicepay.SpringDataNicepayTransactionRepository.exists$suspendImpl(SpringDataNicepayTransactionRepository.kt:70)
 |_ Mono.contextWrite ⇢ at kotlinx.coroutines.reactor.ReactorContextInjector.injectCoroutineContext(ReactorContextInjector.kt:21)
Original Stack Trace: .. 생략 ..


Wrapped by: finance.chai.gateway.transaction.exception.UnspecifiedDatabaseException: Cannot access the database.
 at finance.chai.gateway.transaction.nicepay.SpringDataNicepayTransactionRepository.exists$suspendImpl(SpringDataNicepayTransactionRepository.kt:72)
 Suppressed: reactor.core.publisher.FluxOnAssembly$OnAssemblyException:
Assembly trace from producer [reactor.core.publisher.MonoLift] :
 reactor.core.publisher.Mono.create
 kotlinx.coroutines.reactor.MonoKt.monoInternal(Mono.kt:85)
Error has been observed at the following site(s):
 *_________Mono.create ⇢ at kotlinx.coroutines.reactor.MonoKt.monoInternal(Mono.kt:85)
 |_                    ⇢ at kotlinx.coroutines.reactor.MonoKt.mono(Mono.kt:34)
 |_        Mono.filter ⇢ at org.springframework.core.CoroutinesUtils.invokeSuspendingFunction(CoroutinesUtils.java:80)
 |_    Mono.onErrorMap ⇢ at org.springframework.core.CoroutinesUtils.invokeSuspendingFunction(CoroutinesUtils.java:81)
 |_                    ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$null$0(TransactionAspectSupport.java:918)
 *______Mono.usingWhen ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$null$3(TransactionAspectSupport.java:914)
 |_ Mono.onErrorResume ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$null$3(TransactionAspectSupport.java:927)
 *__________Mono.error ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$null$2(TransactionAspectSupport.java:928)
 *___________Mono.then ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$null$2(TransactionAspectSupport.java:928)
 *________Mono.flatMap ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.lambda$invokeWithinTransaction$4(TransactionAspectSupport.java:911)
 *________Mono.flatMap ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.invokeWithinTransaction(TransactionAspectSupport.java:910)
 |_  Mono.contextWrite ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.invokeWithinTransaction(TransactionAspectSupport.java:934)
 |_  Mono.contextWrite ⇢ at org.springframework.transaction.interceptor.TransactionAspectSupport$ReactiveTransactionSupport.invokeWithinTransaction(TransactionAspectSupport.java:935)
Original Stack Trace: .. 생략 ..
```

### 원인

이름이 변경된 테이블이 없어서 발생한 이슈

### 해결

`nicepay_transaction` 테이블 생성

```sql
-- Nicepay start
CREATE TABLE IF NOT EXISTS nicepay_transaction
(
    transaction_id  VARCHAR(127) NOT NULL,
    tid             VARCHAR      NOT NULL,
    moid            VARCHAR      NOT NULL,
    total_amount    BIGINT       NOT NULL,
    tax_free_amount BIGINT       NOT NULL,
    vat_amount      BIGINT       NULL,
    channel_id      VARCHAR(127) NULL,
    store_id        VARCHAR      NULL,
    "version"       VARCHAR      NULL,
    deleted         bool         NOT NULL DEFAULT false,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT nicepay_transaction_pkey PRIMARY KEY (transaction_id)
);
COMMENT ON COLUMN nicepay_transaction.version IS 'V1: Core, V2: TX';
ALTER TABLE nicepay_transaction
    OWNER TO postgres;

CREATE TRIGGER update_nicepay_transaction_modified_at
    BEFORE UPDATE
    ON
        nicepay_transaction
    FOR EACH ROW
EXECUTE PROCEDURE update_modified_at_tables();
-- Nicepay end
```

## Incremental compilation analysis failed

### 문제

```log
[KOTLIN] [IC] Incremental compilation analysis failed: java.lang.IllegalStateException: The following LookupSymbols are not yet converted to ProgramSymbols: LookupSymbol(name=-initializecancelResponse, scope=finance.chai.gateway.transaction.grpc)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.toProgramSymbols(ClasspathChangesComputer.kt:345)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.computeFineGrainedKotlinClassChanges(ClasspathChangesComputer.kt:258)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.computeKotlinClassChanges(ClasspathChangesComputer.kt:180)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.computeClassChanges(ClasspathChangesComputer.kt:153)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.computeChangedAndImpactedSet(ClasspathChangesComputer.kt:104)
    at org.jetbrains.kotlin.incremental.classpathDiff.ClasspathChangesComputer.computeClasspathChanges(ClasspathChangesComputer.kt:72)
    at org.jetbrains.kotlin.incremental.IncrementalJvmCompilerRunner.calculateSourcesToCompileImpl(IncrementalJvmCompilerRunner.kt:249)
    at org.jetbrains.kotlin.incremental.IncrementalJvmCompilerRunner.calculateSourcesToCompileImpl(IncrementalJvmCompilerRunner.kt:192)
    at org.jetbrains.kotlin.incremental.IncrementalJvmCompilerRunner.calculateSourcesToCompileImpl(IncrementalJvmCompilerRunner.kt:125)
    at org.jetbrains.kotlin.incremental.IncrementalCompilerRunner.calculateSourcesToCompile(IncrementalCompilerRunner.kt:252)
    at org.jetbrains.kotlin.incremental.IncrementalCompilerRunner.sourcesToCompile(IncrementalCompilerRunner.kt:242)
    at org.jetbrains.kotlin.incremental.IncrementalCompilerRunner.compileImpl(IncrementalCompilerRunner.kt:160)
    at org.jetbrains.kotlin.incremental.IncrementalCompilerRunner.compile(IncrementalCompilerRunner.kt:79)
    at org.jetbrains.kotlin.daemon.CompileServiceImplBase.execIncrementalCompiler(CompileServiceImpl.kt:625)
    at org.jetbrains.kotlin.daemon.CompileServiceImplBase.access$execIncrementalCompiler(CompileServiceImpl.kt:101)
    at org.jetbrains.kotlin.daemon.CompileServiceImpl.compile(CompileServiceImpl.kt:1746)
    at jdk.internal.reflect.GeneratedMethodAccessor114.invoke(Unknown Source)
    at java.base/jdk.internal.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
    at java.base/java.lang.reflect.Method.invoke(Method.java:568)
    at java.rmi/sun.rmi.server.UnicastServerRef.dispatch(UnicastServerRef.java:360)
    at java.rmi/sun.rmi.transport.Transport$1.run(Transport.java:200)
    at java.rmi/sun.rmi.transport.Transport$1.run(Transport.java:197)
    at java.base/java.security.AccessController.doPrivileged(AccessController.java:712)
    at java.rmi/sun.rmi.transport.Transport.serviceCall(Transport.java:196)
    at java.rmi/sun.rmi.transport.tcp.TCPTransport.handleMessages(TCPTransport.java:587)
    at java.rmi/sun.rmi.transport.tcp.TCPTransport$ConnectionHandler.run0(TCPTransport.java:828)
    at java.rmi/sun.rmi.transport.tcp.TCPTransport$ConnectionHandler.lambda$run$0(TCPTransport.java:705)
    at java.base/java.security.AccessController.doPrivileged(AccessController.java:399)
    at java.rmi/sun.rmi.transport.tcp.TCPTransport$ConnectionHandler.run(TCPTransport.java:704)
    at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1136)
    at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:635)
    at java.base/java.lang.Thread.run(Thread.java:833)
Falling back to non-incremental compilation
```

## flyway ERROR: relation ... already exists

### 문제

```log
Exception encountered during context initialization - cancelling refresh attempt: org.springframework.beans.factory.BeanCreationException: Error creating bean with name 'flyway' defined in class path resource [service/package/name/transaction/configuration/FlywayConfiguration.class]: Bean instantiation via factory method failed; nested exception is org.springframework.beans.BeanInstantiationException: Failed to instantiate [org.flywaydb.core.Flyway]: Factory method 'flyway' threw exception; nested exception is org.flywaydb.core.internal.command.DbMigrate$FlywayMigrateException: Migration V20230220180721__nicepay.sql failed
---------------------------------------------
SQL State  : 42P07
Error Code : 0
Message    : ERROR: relation \"nicepay_transaction\" already exists
Location   : db/migration/V20230220180721__nicepay.sql (/Users/rody/IdeaProjects/some-project/subproject/presentation/file:/Users/rody/IdeaProjects/some-project/subproject/infrastructure/build/libs/infrastructure-0.0.1-SNAPSHOT-plain.jar!/db/migration/V20230220180721__nicepay.sql)
Line       : 8
Statement  : -- Nicepay start
-- public.nicepay_transaction definition

-- Drop table

-- DROP TABLE public.nicepay_transaction;

CREATE TABLE public.nicepay_transaction
(
    transaction_id  varchar(127) NOT NULL,
    tid             varchar      NOT NULL,
    moid            varchar      NOT NULL,
    total_amount    int8         NOT NULL,
    tax_free_amount int8         NOT NULL,
    vat_amount      int8         NULL,
    channel_id      varchar(127) NULL,
    store_id        varchar      NULL,
    \"version\"       varchar      NULL,
    deleted         bool         NOT NULL DEFAULT false,
    created_at      timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at     timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT nicepay_transaction_pkey PRIMARY KEY (transaction_id)
)
```

### 원인

에러 로그가 알려주다시피, 이미 테이블이 있어서 발생

### 해결

1. 해당 테이블을 삭제하고 실행
2. 존재하지 않으면 실행 `CREATE TABLE IF NOT EXISTS`

## MockKException: no answer found for

### 문제

```log
no answer found for: NicepayParser(#8).parseDepositNotification(MallReserved1=4768800525&ServiceCl=0&MallReserved3=&MallReserved2=&RcptAuthCode=I79881493&FnCd=003&BuyerEmail=anony%40some-qwerty-org.io&MallReserved9=&MallReserved8=&MallReserved5=&MallReserved4=&MallReserved7=&MallUserID=&MallReserved6=&MallReserved=imp_uid%3Dimp_123456789012%2Crequest_id%3Dreq_123456789012%2Cuser_code%3Dimp84043725%2Cvbank_due%3D202303082359&StateCd=0&BuyerAuthNum=&VbankName=%B1%E2%BE%F7%C0%BA%C7%E0&MOID=imp_123456789012&VbankNum=05000711997900&MallReserved10=&ReceitType=1&VbankInputName=%C8%AB%B1%E6%B5%BF&AuthCode=&AuthDate=230307231137&MID=someMID&Amt=1000&EtcSvcCl=0&TID=someMID03012303072307055719&GoodsName=%C8%AB%B1%E6%B5%BF%20%6C%6F%76%65%20%6E%61%74%75%72%61%6C%20%67%6F%6C%64%20%72%69%6E%67%20%28%B0%A2%C0%CE%29&MerchantKey=dbW%2Fd3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h%2Bwky%2BLMmP1PbLw%3D%3D&FnName=%B1%E2%BE%F7%C0%BA%C7%E0&RcptTID=someMID04012303072311372047&CancelMOID=&name=%C8%AB%B1%E6%B5%BF&PayMethod=VBANK&ResultMsg=%BD%C2%C0%CE&RcptType=1&ResultCode=4110)
io.mockk.MockKException: no answer found for: NicepayParser(#8).parseDepositNotification(MallReserved1=4768800525&ServiceCl=0&MallReserved3=&MallReserved2=&RcptAuthCode=I79881493&FnCd=003&BuyerEmail=anony%40some-qwerty-org.io&MallReserved9=&MallReserved8=&MallReserved5=&MallReserved4=&MallReserved7=&MallUserID=&MallReserved6=&MallReserved=imp_uid%3Dimp_123456789012%2Crequest_id%3Dreq_123456789012%2Cuser_code%3Dimp84043725%2Cvbank_due%3D202303082359&StateCd=0&BuyerAuthNum=&VbankName=%B1%E2%BE%F7%C0%BA%C7%E0&MOID=imp_123456789012&VbankNum=05000711997900&MallReserved10=&ReceitType=1&VbankInputName=%C8%AB%B1%E6%B5%BF&AuthCode=&AuthDate=230307231137&MID=someMID&Amt=1000&EtcSvcCl=0&TID=someMID03012303072307055719&GoodsName=%C8%AB%B1%E6%B5%BF%20%6C%6F%76%65%20%6E%61%74%75%72%61%6C%20%67%6F%6C%64%20%72%69%6E%67%20%28%B0%A2%C0%CE%29&MerchantKey=dbW%2Fd3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h%2Bwky%2BLMmP1PbLw%3D%3D&FnName=%B1%E2%BE%F7%C0%BA%C7%E0&RcptTID=someMID04012303072311372047&CancelMOID=&name=%C8%AB%B1%E6%B5%BF&PayMethod=VBANK&ResultMsg=%BD%C2%C0%CE&RcptType=1&ResultCode=4110)
    at io.mockk.impl.stub.MockKStub.defaultAnswer(MockKStub.kt:93)
    at io.mockk.impl.stub.MockKStub.answer(MockKStub.kt:42)
    at io.mockk.impl.recording.states.AnsweringState.call(AnsweringState.kt:16)
    at io.mockk.impl.recording.CommonCallRecorder.call(CommonCallRecorder.kt:53)
    at io.mockk.impl.stub.MockKStub.handleInvocation(MockKStub.kt:271)
    at io.mockk.impl.instantiation.JvmMockFactoryHelper$mockHandler$1.invocation(JvmMockFactoryHelper.kt:24)
    at io.mockk.proxy.jvm.advice.Interceptor.call(Interceptor.kt:21)
    at io.mockk.proxy.jvm.advice.BaseAdvice.handle(BaseAdvice.kt:42)
    at io.mockk.proxy.jvm.advice.jvm.JvmMockKProxyInterceptor.interceptNoSuper(JvmMockKProxyInterceptor.java:45)
    at finance.chai.gateway.transaction.nicepay.NicepayParser$Subclass7.parseDepositNotification(Unknown Source)
    at finance.chai.gateway.transaction.notification.NotificationServiceNicepayExtKt.notifyNicepayDeposit(NotificationServiceNicepayExt.kt:12)
    at finance.chai.gateway.transaction.notification.NotificationService.notifyDeposit$suspendImpl(NotificationService.kt:32)
    at finance.chai.gateway.transaction.notification.NotificationService.notifyDeposit(NotificationService.kt)
    at finance.chai.gateway.transaction.notification.NotificationServiceTest$2.invokeSuspend(NotificationServiceTest.kt:116)
    at finance.chai.gateway.transaction.notification.NotificationServiceTest$2.invoke(NotificationServiceTest.kt)
    at finance.chai.gateway.transaction.notification.NotificationServiceTest$2.invoke(NotificationServiceTest.kt)
    at io.kotest.core.spec.style.scopes.DescribeSpecRootScope$describe$1.invokeSuspend(DescribeSpecRootScope.kt:42)
    at io.kotest.core.spec.style.scopes.DescribeSpecRootScope$describe$1.invoke(DescribeSpecRootScope.kt)
    at io.kotest.core.spec.style.scopes.DescribeSpecRootScope$describe$1.invoke(DescribeSpecRootScope.kt)
    at io.kotest.core.spec.style.scopes.RootScopeKt$addTest$1.invokeSuspend(RootScope.kt:36)
    at io.kotest.core.spec.style.scopes.RootScopeKt$addTest$1.invoke(RootScope.kt)
    at io.kotest.core.spec.style.scopes.RootScopeKt$addTest$1.invoke(RootScope.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$innerExecute$1.invokeSuspend(TestCaseExecutor.kt:83)
    at io.kotest.engine.test.TestCaseExecutor$execute$innerExecute$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$innerExecute$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.CoroutineDebugProbeInterceptor.intercept(CoroutineDebugProbeInterceptor.kt:29)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.InvocationTimeoutInterceptor.intercept(InvocationTimeoutInterceptor.kt:28)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2$3.invokeSuspend(TestInvocationInterceptor.kt:36)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2$3.invoke(TestInvocationInterceptor.kt)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2$3.invoke(TestInvocationInterceptor.kt)
    at io.kotest.mpp.ReplayKt.replay(replay.kt:18)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2.invokeSuspend(TestInvocationInterceptor.kt:31)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2.invoke(TestInvocationInterceptor.kt)
    at io.kotest.engine.test.TestInvocationInterceptor$intercept$2.invoke(TestInvocationInterceptor.kt)
    at kotlinx.coroutines.intrinsics.UndispatchedKt.startUndispatchedOrReturn(Undispatched.kt:89)
    at kotlinx.coroutines.CoroutineScopeKt.coroutineScope(CoroutineScope.kt:264)
    at io.kotest.engine.test.TestInvocationInterceptor.intercept(TestInvocationInterceptor.kt:30)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.TimeoutInterceptor.intercept(TimeoutInterceptor.kt:33)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.BlockedThreadTimeoutInterceptor.intercept(BlockedThreadTimeoutInterceptor.kt:74)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.CoroutineLoggingInterceptor.intercept(CoroutineLoggingInterceptor.kt:30)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.SoftAssertInterceptor.intercept(SoftAssertInterceptor.kt:26)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.AssertionModeInterceptor.intercept(AssertionModeInterceptor.kt:24)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.LifecycleInterceptor.intercept(LifecycleInterceptor.kt:51)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.EnabledCheckInterceptor.intercept(EnabledCheckInterceptor.kt:31)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.TestCaseExtensionInterceptor$intercept$2.invokeSuspend(TestCaseExtensionInterceptor.kt:24)
    at io.kotest.engine.test.interceptors.TestCaseExtensionInterceptor$intercept$2.invoke(TestCaseExtensionInterceptor.kt)
    at io.kotest.engine.test.interceptors.TestCaseExtensionInterceptor$intercept$2.invoke(TestCaseExtensionInterceptor.kt)
    at io.kotest.engine.test.TestExtensions.intercept(TestExtensions.kt:154)
    at io.kotest.engine.test.interceptors.TestCaseExtensionInterceptor.intercept(TestCaseExtensionInterceptor.kt:24)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.CoroutineErrorCollectorInterceptor$intercept$3.invokeSuspend(CoroutineErrorCollectorInterceptor.kt:28)
    at io.kotest.engine.test.interceptors.CoroutineErrorCollectorInterceptor$intercept$3.invoke(CoroutineErrorCollectorInterceptor.kt)
    at io.kotest.engine.test.interceptors.CoroutineErrorCollectorInterceptor$intercept$3.invoke(CoroutineErrorCollectorInterceptor.kt)
    at kotlinx.coroutines.intrinsics.UndispatchedKt.startUndispatchedOrReturn(Undispatched.kt:89)
    at kotlinx.coroutines.BuildersKt__Builders_commonKt.withContext(Builders.common.kt:169)
    at kotlinx.coroutines.BuildersKt.withContext(Unknown Source)
    at io.kotest.engine.test.interceptors.CoroutineErrorCollectorInterceptor.intercept(CoroutineErrorCollectorInterceptor.kt:27)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invokeSuspend(TestCaseExecutor.kt:92)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.TestCaseExecutor$execute$3$1.invoke(TestCaseExecutor.kt)
    at io.kotest.engine.test.interceptors.CoroutineDispatcherFactoryInterceptor$intercept$4.invokeSuspend(coroutineDispatcherFactoryInterceptor.kt:57)
    at io.kotest.engine.test.interceptors.CoroutineDispatcherFactoryInterceptor$intercept$4.invoke(coroutineDispatcherFactoryInterceptor.kt)
    at io.kotest.engine.test.interceptors.CoroutineDispatcherFactoryInterceptor$intercept$4.invoke(coroutineDispatcherFactoryInterceptor.kt)
    at io.kotest.engine.concurrency.FixedThreadCoroutineDispatcherFactory$withDispatcher$4.invokeSuspend(FixedThreadCoroutineDispatcherFactory.kt:53)
    at kotlin.coroutines.jvm.internal.BaseContinuationImpl.resumeWith(ContinuationImpl.kt:33)
    at kotlinx.coroutines.DispatchedTask.run(DispatchedTask.kt:106)
    at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1136)
    at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:635)
    at java.base/java.lang.Thread.run(Thread.java:833)
```

### 원인

kotest 실행 시에 메서드 호출은 있는데, 예상되는 return이 미리 설정되어 있지 않다면 위와 같은 에러 발생

### 해결

```kotlin
coEvery {
    // 호출될 메서드와 완벽히 일치하는 메서드 호출
    // 또는 모두 any()호 채우기
} returns answer // 리턴할 결과
```

## kover - Unresolved reference: excludes

### 문제

```kts
tasks {
    kover {
        filters {
            classes {
                excludes +=
//              ^^^^^^^^ Unresolved reference: excludes 에러
                    listOf(
                        "package.name.*"
                    )
            }
        }
    }

    named("generateJooq") { dependsOn(composeUp) }
}
```

### 원인

`classes`와 `projects` 등 kotlin dsl에 같이 사용중인 변수가 있는데, `kover`에서 사용중인 이름과 겹쳐서 발생.

```log
/caches/8.2.1/kotlin-dsl/accessors/598604974bcaa30709029155e7fe8756/sources/org/gradle/kotlin/dsl/Accessorsem8w6wnof1lrw3ubqr6eh9gcj.kt
```

### 해결

`classes`나 `projects`를 잘못 파악하는 것이므로, filter 하위에서 `this.`를 붙여서 `KoverProjectFilters`의 속성임을 명시

```kts
tasks {
    kover {
        filters {
            this.classes {
//          ^^^^^ `classes` 변수의 스코프를 지정
                excludes +=
                    listOf(
                        "package.name.*"
                    )
            }
        }
    }

    named("generateJooq") { dependsOn(composeUp) }
}
```
