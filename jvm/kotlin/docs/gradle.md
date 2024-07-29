# gradle

- [gradle](#gradle)
    - [LibrariesForLibs](#librariesforlibs)
    - [Type-safe project accessors](#type-safe-project-accessors)
    - [특정 테스트 케이스만 실행](#특정-테스트-케이스만-실행)

## LibrariesForLibs

## Type-safe project accessors

```log
Type-safe project accessors is an incubating feature
```

## 특정 테스트 케이스만 실행

```shell
# https://kotest.io/docs/framework/conditional/conditional-tests-with-gradle.html
# https://github.com/kotest/kotest-gradle-plugin/issues/3
./gradlew :subproject:some-B-serivce:test --tests 'kr.some_name.api.core.users.*'
```

```shell
./gradlew :domain:test --tests "finance.chai.gateway.transaction.util.TimeExtTest"
```

`--info`로 로그 확인해 보니 test 태스크가 up-to-date 라고 판단되면 skip

```log
> Task :domain:test UP-TO-DATE
Excluding []
Caching disabled for task ':domain:test' because:
  Build cache is disabled
Skipping task ':domain:test' as it is up-to-date.
:domain:test (Thread[included builds,5,main]) completed. Took 0.017 secs.
producer locations for task group 0 (Thread[included builds,5,main]) started.
producer locations for task group 0 (Thread[included builds,5,main]) completed.
Took 0.0 secs.
```

테스트를 새로 실행하려면 [cleanTest](https://stackoverflow.com/a/29428063)(테스트 결과 제거) && `--no-build-cache` 옵션 사용 필요

```shell
./gradlew cleanTest \
    :domain:test \
    --no-build-cache \
    --tests 'finance.chai.gateway.transaction.util.TimeExtTest'
```

콘솔로 결과 자세히 보거나, 콘솔로 출력하는 걸 보고 싶은 경우 build.gradle.kts에 조금 수정이 필요

```kts
    tasks.withType<Test> {
        useJUnitPlatform()
        testLogging {
            lifecycle {
                events = mutableSetOf(TestLogEvent.FAILED, TestLogEvent.PASSED, TestLogEvent.SKIPPED)
                exceptionFormat = TestExceptionFormat.FULL
                showExceptions = true
                showCauses = true
                showStackTraces = true
                showStandardStreams = true
            }
            info.events = lifecycle.events
            info.exceptionFormat = lifecycle.exceptionFormat
        }
    }
```

```shell

./gradle payment-hub-service:sub
```
