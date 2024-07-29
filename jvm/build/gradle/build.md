# build

- [build](#build)
    - [dependencies](#dependencies)
        - [api vs implementation](#api-vs-implementation)
            - [`api`](#api)
            - [`implementaion`](#implementaion)
    - [`out` directory](#out-directory)
        - [`NicepayHttpClient.kt`](#nicepayhttpclientkt)

## dependencies

### api vs implementation

- [(Gradle dependency) api와 implementation 차이](https://jongmin92.github.io/2019/05/09/Gradle/gradle-api-vs-implementation/)
- [Using dependency configurations](https://docs.gradle.org/current/userguide/dependency_management_for_java_projects.html#sec:configurations_java_tutorial)
- [Android studio - Dependency configurations](https://developer.android.com/studio/build/dependencies?utm_source=android-studio#dependency_configurations)

#### `api`

> 의존 라이브러리 수정시 해당 모듈을 의존하고 있는 모듈들 또한 재빌드

- `A`(api) <- `B` <- `C` 일 때, `C` 에서 `A` 를 접근할 수 있음
- `A` 수정시 `B` 와 `C` 모두 재빌드

gradle 문서 설명

> The dependencies required to compile the production source of the project which **are part of the API** exposed by the project. For example the project uses Guava and exposes public interfaces with Guava classes in their method signatures.
>
> 프로젝트가 노출하는 **API의 일부인** 프로젝트 프로덕션 소스를 컴파일하는 데 필요한 의존성.

andoid studio 문서 설명

> Gradle
>
> 1. adds the dependency to the `compile classpath` and `build output`.
>
> When a module includes an api dependency, it's letting Gradle know that the module wants to transitively export that dependency to other modules, so that it's available to them at both runtime and compile time.
>
> This configuration behaves just like compile (which is now deprecated), but you should use it with caution and only with dependencies that you need to transitively export to other upstream consumers. That's because, if an api dependency changes its external API, Gradle recompiles all modules that have access to that dependency at compile time. So, having a large number of api dependencies can significantly increase build time. Unless you want to expose a dependency's API to a separate module, library modules should instead use implementation dependencies.

#### `implementaion`

> 의존 라이브러리 수정시 본 모듈까지만 재빌드

- `A`(implementation) <- `B` <- `C` 일 때, `C` 에서 `A` 를 접근할 수 없음
- `A` 수정시 `B` 까지 재빌드

gradle 문서 설명

> The dependencies required to compile the production source of the project which **are not part of the API** exposed by the project. For example the project uses Hibernate for its internal persistence layer implementation.
>
> 프로젝트가 노출하는 **API의 일부가 아닌** 프로젝트의 프로덕션 소스를 컴파일하는 데 필요한 의존성들.

andoid studio 문서 설명

> Gradle
>
> 1. adds the dependency to the `compile classpath`
> 2. and packages the dependency to the `build output`.
>
> However, when your module configures an implementation dependency, it's letting Gradle know that you do not want the module to leak the dependency to other modules at compile time. That is, the dependency is available to other modules only at runtime.
>
> Using this dependency configuration instead of api or compile (deprecated) can result in significant build time improvements because it reduces the number of modules that the build system needs to recompile. For example, if an implementation dependency changes its API, Gradle recompiles only that dependency and the modules that directly depend on it. Most app and test modules should use this configuration.

## `out` directory

### `NicepayHttpClient.kt`

하나의 클래스라고 해도, 컴파일을 하면 내부 각 클래스 및 메서드들별로 `.class` 파일이 분리가 된다

디렉토리 구조가 `build`와 `out`이 있는데, `build`는 gradle 통해 빌드할 경우, `out`은 intellij 통해 빌드할 경우의 아웃풋 디렉토리다. `out` 경로는 `Intellij > Project Structure > Modules > 각 프로젝트 > Paths 탭`에 설정되어 있다

```tree
rootProject/
  ㄴ subproject/
    ㄴ application/
      ㄴ build/ (by gradle)
      ㄴ out/ (by intellij setting)
      ㄴ src/
        ㄴ main/
          ㄴ kotlin
```

`out` 디렉토리인 경우.

```shell
 rody ~/IdeaProjects/some-project/subproject/infrastructure/out/production/classes/service/package/name/transaction/nicepay  ls -a | grep NicepayHttpClient
NicepayHttpClient$CancelRequestBody.class
NicepayHttpClient$CancelResponseBody.class
NicepayHttpClient$ConfirmRequestBody.class
NicepayHttpClient$EdiType.class
NicepayHttpClient$IssueVirtualAccountRequest.class
NicepayHttpClient$IssueVirtualAccountResponse.class
NicepayHttpClient$NetCancelRequestBody.class
NicepayHttpClient$NicepayRawAuthResult.class
NicepayHttpClient$PaymentResponse.class
NicepayHttpClient$confirm$1.class
NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1$1.class
NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1.class
NicepayHttpClient$requestFormData$$inlined$request$1.class
NicepayHttpClient$requestFormData$response$1$1.class
NicepayHttpClient.class
```

`build` 디렉토리인 경우

```shell
 rody  ~/IdeaProjects/some-project/subproject/infrastructure/build/classes/kotlin/main/service/package/name/transaction/nicepay ls -a | grep NicepayHttpClient
NicepayHttpClient$CancelRequestBody.class
NicepayHttpClient$CancelResponseBody.class
NicepayHttpClient$ConfirmRequestBody.class
NicepayHttpClient$EdiType.class
NicepayHttpClient$IssueVirtualAccountRequest.class
NicepayHttpClient$IssueVirtualAccountResponse.class
NicepayHttpClient$NetCancelRequestBody.class
NicepayHttpClient$NicepayRawAuthResult.class
NicepayHttpClient$PaymentResponse$WhenMappings.class
NicepayHttpClient$PaymentResponse.class
NicepayHttpClient$confirm$1.class
NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1$1.class
NicepayHttpClient$confirm$suspendImpl$$inlined$requestFormData$1.class
NicepayHttpClient$requestFormData$$inlined$request$1.class
NicepayHttpClient$requestFormData$response$1$1.class
NicepayHttpClient.class
NicepayHttpClientKt.class
```

빌드 툴에 따라 메서드가 위치하는 경로가 다르다. 가령 공통으로 쓰일 함수를 class 밖에 선언한 경우

```kt
@Repository
class NicepayHttpClient(
    private val niceWebClient: WebClient,
    private val nicepayResponseValidator: NicepayResponseValidator,
    private val niceTotalTimeoutMillis: Long
) : NicepayClient {
  ... 생략 ...
}

// 클래스 밖에 함수를 선언
private fun toNicepayAmt(amt: Long) = amt.toString().padStart(12, '0')
```

`out` 디렉토리에는 `NicepayHttpClient.class` 하나의 클래스 파일에 같이 존재하지만,

```shell
ls -a | grep NicepayHttpClient | xargs grep "toNicepayAmt"
Binary file NicepayHttpClient.class matches
```

`build`인 경우 `NicepayHttpClientKt.class` 파일에 별도로 따로 빠져 있다. 그리고 이 파일을 연결해서 사용하는 거 같다.

```shell
ls -a | grep NicepayHttpClient | xargs grep "toNicepayAmt"
Binary file NicepayHttpClientKt.class matches
```
