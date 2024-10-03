# kotest

## 개별 테스트 실행

```shell
./gradlew :subproject:domain:test --tests '*TimeExtTest'
```

## [캐시 없이 실행](https://stackoverflow.com/a/29428063)

`--info` 옵션으로 상세 내역을 출력해보면 test 태스크가 최신이면 스킵을 하는 거 같다.(`Skipping task ':domain:test' as it is up-to-date.`)

```log
> Task :domain:test UP-TO-DATE
Excluding []
Caching disabled for task ':domain:test' because:
  Build cache is disabled
Skipping task ':domain:test' as it is up-to-date.
:domain:test (Thread[included builds,5,main]) completed. Took 0.017 secs.
producer locations for task group 0 (Thread[included builds,5,main]) started.
producer locations for task group 0 (Thread[included builds,5,main]) completed. Took 0.0 secs.
```

그래서 테스트 실행 전 테스트 결과를 깨끗하게 만들고 싶다면, [`cleanTest` 태스크](https://docs.gradle.org/current/userguide/java_plugin.html#sec:java_tasks)사용.

> cleanTaskName - Deletes files created by specified task. cleanJar will delete the JAR file created by the jar task, and cleanTest will delete the test results created by the test task.

그래서 이전 테스트 결과 지우고 새롭게 테스트 실행하려면, 아래 명령어 사용

- `cleanTest`: 이전 테스트 결과 제거
- `--no-build-cache`: 빌드 캐시 없도록 함

```shell
./gradlew cleanTest :domain:test --no-build-cache --tests 'finance.chai.gateway.transaction.util.TimeExtTest' 
```

## [테스트 결과를 터미널에서 바로 보기](https://technology.lastminute.com/junit5-kotlin-and-gradle-dsl/)

```kts
tasks {
    test {
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

        // See https://github.com/gradle/kotlin-dsl/issues/836
        addTestListener(object : TestListener {
            override fun beforeSuite(suite: TestDescriptor) {}
            override fun beforeTest(testDescriptor: TestDescriptor) {}
            override fun afterTest(testDescriptor: TestDescriptor, result: TestResult) {}

            override fun afterSuite(suite: TestDescriptor, result: TestResult) {
                if (suite.parent == null) { // root suite
                    logger.lifecycle("----")
                    logger.lifecycle("Test result: ${result.resultType}")
                    logger.lifecycle("Test summary: ${result.testCount} tests, " +
                        "${result.successfulTestCount} succeeded, " +
                        "${result.failedTestCount} failed, " +
                        "${result.skippedTestCount} skipped")
                }
            }
        })
    }
}
```
