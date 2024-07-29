# Settings

- [Settings](#settings)
    - [`build.gradle.kts`](#buildgradlekts)
        - [`build.gradle.kts` script itself](#buildgradlekts-script-itself)
        - [`PluginDependencySpecScop` of `plugins` block of `build.gradle.kts`](#plugindependencyspecscop-of-plugins-block-of-buildgradlekts)
    - [plugins](#plugins)
        - [`plugins {}`](#plugins-)
            - [TL;DL](#tldl)
            - [`PluginAware.apply(java.util.Map)` 메서드와의 관계](#pluginawareapplyjavautilmap-메서드와의-관계)
        - [configutations](#configutations)
        - [allprojects](#allprojects)
        - [subprojects](#subprojects)
        - [tasks](#tasks)
        - [kotlin](#kotlin)
        - [최종](#최종)
    - [ETC](#etc)
        - [references](#references)

## `build.gradle.kts`

`build.gradel.kts`는 기본적으로 그 자신 `Project`만 관리한다.

따라서 만약 root project의 `build.gradle.kts`에서 하위 프로젝트들의 의존성을 관리하고 싶다면 `subprojects`, `allprojects` 같은 스크립트 블록으로 관리해야 한다.
아니면 각 하위 프로젝트에서 별도의 `build.gradle.kts`에서 각각 관리를 해야 한다.

### `build.gradle.kts` script itself

```kts
println("==============================================")
println("\tclass: ${this::class}")
println("\tqualifiedName: ${this::class.qualifiedName}")
println("\tjava class: ${this::class.java}")
println("\tjava canonicalName: ${this::class.java.canonicalName}")
println("\tjava packageName: ${this::class.java.packageName}")
println("==============================================")
```

```log
==============================================
 class: class Build_gradle
 qualifiedName: Build_gradle
 java class: class Build_gradle
 java canonicalName: Build_gradle
 java packageName: 
==============================================
```

### `PluginDependencySpecScop` of `plugins` block of `build.gradle.kts`

```log
==============================================
 class: class org.gradle.plugin.use.internal.PluginRequestCollector$PluginDependenciesSpecImpl
 qualifiedName: org.gradle.plugin.use.internal.PluginRequestCollector.PluginDependenciesSpecImpl
 java class: class org.gradle.plugin.use.internal.PluginRequestCollector$PluginDependenciesSpecImpl
 java canonicalName: org.gradle.plugin.use.internal.PluginRequestCollector.PluginDependenciesSpecImpl
 java packageName: org.gradle.plugin.use.internal
==============================================
```

## plugins

### `plugins {}`

`plugins {}` 블록을 사용해서 현재 스크립트에서 사용할 플러그인을 선언할 수 있다. 다만 그렇다고 해서 이 플러그인이 모두 하위프로젝트에 적용되는 것은 아니다.

#### TL;DL

`plugins {}` 블록은 `PluginDependenciesSpecScope`을 수신자(receiver)로 받는다.
`PluginDependenciesSpecScope` 클래스는 오직 `plugins {}` 블록을 `kotlin-dsl/src/main/kotlin/org/gradle/kotlin/dsl/GradleDsl.kt`로 표시하기 위해 존재하며,
따라서 외부 `kotlin-dsl/src/main/kotlin/org/gradle/kotlin/dsl/KotlinBuildScript.kt` 범위에서 제공하는 모든 멤버를 감춘다.

`kotlin-dsl/src/main/kotlin/org/gradle/kotlin/dsl/GradleDsl.kt`는 Gradle DSL임을 구분하기 위한 어노테이션 클래스이며, `@DslMarker` 어노테이션이 붙어 있다.

`PluginDependenciesSpecScope`은 `PluginDependenciesSpec`을 상속하고 생서자 인자로 받는다.

```kt
@GradleDsl
class PluginDependenciesSpecScope internal constructor(
    private val plugins: PluginDependenciesSpec
) : PluginDependenciesSpec {
```

`PluginDependenciesSpec`은 스크립트에서 사용한 플러그인을 선언하기 위한 DSL이다.
즉, 빌드 스크립트에서 `PluginDependenciesSpecScope`을 수신자로 받는 `plugins {}`블록이 `PluginDependenciesSpec`타입이다.
스크립트에서 사용할 플러그인을 선언하기 위해`PluginDependenciesSpec`API를`plugins {}` 블록 바디에서 사용할 수 있다.

```kts
plugins {
    // `PluginDependenciesSpec`을 상속하는 `PluginDependenciesSpecScope`
    /* Core plugins */
    java // == id("org.gradle.java")
    idea

    /* Kotlin plugins */
    kotlin("jvm") version Version.kotlin apply true // == id("org.jetbrains.kotlin.jvm")
    kotlin("plugin.spring") version Version.kotlin
    kotlin("kapt") version Version.kotlin

    /* Spring plugins */
    id("org.springframework.boot") version Version.springBootVersion
    id("io.spring.dependency-management") version Version.springDependencyManagement
}
```

- portal의 community plugin 경우 fully qualified id 사용해야 한다.
    - `java` == `id("org.gradle.java")`
    - `kotlin("jvm")` == `id("org.jetbrains.kotlin.jvm")`
- [`PluginDependenciesSpec`](https://docs.gradle.org/current/javadoc/org/gradle/plugin/use/PluginDependenciesSpec.html) 참고

#### `PluginAware.apply(java.util.Map)` 메서드와의 관계

- `PluginAware.apply(java.util.Map)`은 `Project` 또는 비슷한 오브젝트에 플러그인을 직접 적용할 때 사용된다.
- `plugins {}` 블록은 `PluginAware.apply(java.util.Map)` 메서드와 비슷한 목적을 갖는다
- 주요 차이점은 `plugins {}` 블록을 통해 적용되는 플러그인들은 개념적으로 스크립트에 적용되고, 결과적으로 스크립트 대상에 적용이 된다는 것
- 현재로서는 결과적으로 두 방식에 주목할 만한 차이점은 없다.

### [configutations](https://docs.gradle.org/current/dsl/org.gradle.api.Project.html#org.gradle.api.Project:configurations(groovy.lang.Closure))

```kts
configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}
```

### [allprojects](https://docs.gradle.org/current/dsl/org.gradle.api.Project.html#org.gradle.api.Project:allprojects(groovy.lang.Closure))

### [subprojects](https://docs.gradle.org/current/dsl/org.gradle.api.Project.html#org.gradle.api.Project:subprojects(groovy.lang.Closure))

### tasks

### kotlin

### 최종

```kts
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    /* Core plugins */
    java
    idea

    /* Kotlin plugins */
    kotlin("jvm") version Version.kotlin apply true
    kotlin("plugin.spring") version Version.kotlin
    kotlin("kapt") version Version.kotlin

    /* Spring plugins */
    id("org.springframework.boot") version Version.springBootVersion
    id("io.spring.dependency-management") version Version.springDependencyManagement
}

configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}


allprojects {
    group = "aimpugn.spring.template"

    apply {
        plugin("idea")
        plugin("java")
    }

    repositories {
        mavenCentral()
    }
}


subprojects {
    apply {
        // boot plugin을 적용해야 `developmentOnly`가 컴파일 된다
        // - https://github.com/spring-gradle-plugins/dependency-management-plugin/issues/316#issuecomment-1004915809
        // - https://docs.spring.io/spring-boot/docs/current/gradle-plugin/reference/htmlsingle/#getting-started
        plugin("org.springframework.boot")
        // for spring dependency like `spring-boot-starter-webflux`
        plugin("io.spring.dependency-management")
    }

    dependencies {
        implementation("org.springframework.boot:spring-boot-starter-webflux")
        implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
        implementation("io.projectreactor.kotlin:reactor-kotlin-extensions")
        implementation("org.jetbrains.kotlin:kotlin-reflect")
        implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
        implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")
        developmentOnly("org.springframework.boot:spring-boot-devtools")
        annotationProcessor("org.springframework.boot:spring-boot-configuration-processor")
        testImplementation("org.springframework.boot:spring-boot-starter-test")
        testImplementation("io.projectreactor:reactor-test")
    }
}


tasks {
    withType<KotlinCompile> {
        kotlinOptions {
            freeCompilerArgs = listOf("-Xjsr305=strict")
            jvmTarget = Version.java
        }
    }
    withType<Test> {
        useJUnitPlatform()
    }
}

kotlin {
    jvmToolchain {
        // `java.sourceCompatibility = JavaVersion.VERSION_17`와 함께 사용 불가
        languageVersion.set(JavaLanguageVersion.of(Version.java))
    }
}
```

## ETC

### references

- [Gradle Build Language Reference](https://docs.gradle.org/current/dsl/index.html)
- [Kotlin DSL: from Theory to Practice](https://www.jmix.io/cuba-blog/kotlin-dsl-from-theory-to-practice/)
