plugins {
    idea

    /**
     * Java 애플리케이션의 진입점(main 클래스)을 지정하고, 실행 가능 JAR 파일을 빌드하거나 애플리케이션을 CLI에서 바로 실행할 수 있도록 돕는 플러그인입니다.
     * 주로 CLI 애플리케이션을 빌드하고 실행하는 데 사용됩니다.
     * - `run` 작업이 추가되어 gradle run 명령으로 애플리케이션을 실행할 수 있습니다.
     * - `application` 블록의 `mainClass` 속성을 통해 애플리케이션의 진입점 클래스를 지정할 수 있습니다.
     * - `installDist`와 `distZip` 작업이 추가되어 애플리케이션의 실행 가능한 디렉토리 구조를 설정하고, 이를 ZIP 파일로 배포할 수 있습니다.
     *
     * `application` 플러그인은 `java` 플러그인을 암시적으로 적용합니다.
     *
     * References:
     * - https://docs.gradle.org/current/javadoc/org/gradle/api/plugins/JavaApplication.html
     */
    application

    /**
     * kotlin("jvm") 플러그인은 Kotlin 코드를 JVM 바이트코드로 컴파일하기 위해 필요한 Kotlin 컴파일러와 관련 설정을 Gradle에 추가합니다.
     */
    alias(libs.plugins.kotlin.jvm)
    /**
     * 코틀린은 기본적으로 `final class`이고, 필요한 경우에 `open`으로 지정해야 합니다.
     * `plugin.allopen` 플러그인을 기반으로 하여 Spring 프레임워크에서 클래스와 메서드가 `open` 상태로 동작하도록 만들어줍니다.
     * 이를 통해 Spring AOP(Aspect-Oriented Programming) 및 프록시 기반 기능을 사용할 수 있습니다.
     * - 클래스 레벨 어노테이션
     *    - `@Component`
     *    - `@Configuration`
     *    - `@Controller`
     *    - `@RestController`
     *    - `@Service`
     *    - `@Repository`
     * - 메서드 레벨 어노테이션:
     *    - `@Async
     *    - `@Transactional`
     *    - `@Cacheable`
     *
     * References:
     * - https://plugins.gradle.org/plugin/org.jetbrains.kotlin.plugin.spring
     * - https://kotlinlang.org/docs/all-open-plugin.html#spring-support
     */
    alias(libs.plugins.kotlin.spring)
    /**
     * kotlin("plugin.jpa")은 Kotlin에서 JPA(Java Persistence API)를 사용할 때 필요한 설정을 간편하게 처리해주는 컴파일러 플러그인입니다.
     * Kotlin은 기본적으로 이러한 생성자를 제공하지 않는데 JPA는 리플렉션을 통해 엔티티를 생성하기 때문에 기본 생성자가 필요합니다.
     * JPA 엔티티 클래스에 기본 생성자(no-arg constructor)를 자동으로 생성하여 JPA와 Kotlin 간의 호환성을 높여줍니다.
     * - `plugin.spring` 및 `plugin.allopen`과 함께 사용됩니다.
     * - `no-arg` 플러그인을 기반으로 하며, 특정 어노테이션이 붙은 클래스에 대해 컴파일러가 바이트코드에 기본 생성자를 삽입합니다.
     * - `@Entity`, `@Embeddable`, 그리고 `@MappedSuperclass` 어노테이션이 붙은 클래스에
     *   `no-arg` 어노테이션을 지정하여 컴파일 시점에 기본 생성자를 추가됩니다.
     *
     * References:
     * - https://plugins.gradle.org/plugin/org.jetbrains.kotlin.plugin.jpa
     * - https://kotlinlang.org/docs/no-arg-plugin.html#jpa-support
     */
    alias(libs.plugins.kotlin.jpa)

    /**
     * Spring 관련 플러그인
     */
    alias(libs.plugins.spring.boot)
    alias(libs.plugins.spring.dependency.management)

    /**
     * KSP는 직접 코틀린 코드를 분석하며, [KAPT 대비 최대 2배 이상 빠르다](https://android-developers.googleblog.com/2021/09/accelerated-kotlin-build-times-with.html)고 합니다.
     * KAPT는 Java 어노테이션 처리 인프라를 통해 동작하는데, 이 때문에 Java 어노테이션 프로세서에서 관심을 갖는 정보를
     * 제공하기 위해 Kotlin 코드를 Java 스텁(stub)으로 컴파일합니다.
     * 이러한 스텁 생성은 비용이 많이 들고, 컴파일러는 프로그램의 모든 심볼을 여러 번 확인(resolve)해야 합니다.
     *
     * References:
     * - https://kotlinlang.org/docs/ksp-quickstart.html
     * - https://developer.android.com/build/migrate-to-ksp
     */
    // alias(libs.plugins.devtools.ksp)
    // alias(libs.plugins.kotlin.kapt)
}

dependencies {
    implementation(libs.spring.boot.starter)
    implementation(libs.spring.boot.starter.web)
    implementation(libs.spring.boot.starter.aop)
    implementation(libs.aspectj.aspectjrt)
    implementation(libs.kotlinx.coroutines.core)
    // 단순 MVC 경우라 하더라도 coroutines-reactor 패키지가 필요합니다.
    // `CoroutinesUtils.invokeSuspendingFunction`를 사용하는데, 이때 `org.reactivestreams.Publisher`를 사용합니다.
    // ```
    // java.lang.ClassNotFoundException: org.reactivestreams.Publisher
    // ... 생략...
    // 	at org.springframework.web.method.support.InvocableHandlerMethod.invokeSuspendingFunction(InvocableHandlerMethod.java:293) ~[spring-web-6.2.1.jar:6.2.1]
    // ```
    // - https://docs.spring.io/spring-framework/reference/languages/kotlin/coroutines.html#dependencies
    implementation(libs.kotlinx.coroutines.reactor)
    /* DB */
    // Spring Data JPA는 Spring 프레임워크 위에서 동작하는 모듈로, JPA를 더 쉽게 사용할 수 있도록 도와줍니다.
    // - 기본적인 CRUD 메서드(예: `save`, `findById`, `delete`)를 자동으로 생성.
    // - 메서드 이름 기반 쿼리 생성(예: `findByNameAndAge`).
    // - JPQL이나 Native Query를 간단히 정의할 수 있는 기능 제공.
    implementation(libs.spring.boot.starter.data.jpa)
    runtimeOnly(libs.db.h2)
    // [Hibernate](https://docs.jboss.org/hibernate/orm/current/introduction/html_single/Hibernate_Introduction.html#hello-hibernate)
    // - JPA의 구현체 중 하나로, JPA 스펙을 따르면서 캐싱, 지연 로딩(Lazy Loading), 배치 처리 등 추가적인 기능과 최적화를 제공합니다.
    implementation(libs.db.hibernate)
    implementation(libs.db.hibernate.agroal) // connection pool
    implementation(libs.db.agroal.pool) // connection pool
    // JPA Metamodel Generator: 정적 메타모델(static metamodel)을 생성해주는 어노테이션 프로세서입니다.
    // - https://docs.jboss.org/hibernate/orm/6.6/introduction/html_single/Hibernate_Introduction.html#generator
    // 단, 다음과 같은 이유로 사용하지 않습니다:
    // `kapt`는 코틀린 코드를 스캔하여 어노테이션 프로세서를 돌려야 할 부분을 찾고,
    // 어노테이션 프로세서가 처리하는 데 필요한 정보는 갖는 자바 스텁(stub) 파일을 생성합니다. 이 스텁은 아직 완전한 Java 클래스가 아닙니다.
    // 그 후 `jpamodelgen`가 메타 모델 클래스(ex: ItemKey_.java)를 생성되고, 그 후 Java 컴파일러에 의해 '.class'로 컴파일 됩니다.
    // 마지막으로 Kotlin 컴파일러는 Kotlin 코드와 앞서 생성된 '.class' 파일을 함께 사용하여 최종 바이트코드를 생성합니다.
    // 이 과정에서 Kotlin 코드가 메타모델 클래스의 상수(예: `ItemKey_.CREATED`)를 참조하려고 합니다.
    // 그러나 메타모델 클래스(`ItemKey_`)가 생성되고 컴파일되는 시점과 Kotlin 코드가 이를 참조하는 시점이 앞서 설명한 것처럼
    // 완전히 순차적으로 이뤄지는 게 아니라, "동일한 빌드 사이클 안에서 처리"되어서 이를 컴파일 타임 상수로 인식하지 못하는 것으로 보입니다.
    // ```
    // @Column(name = ItemValue_.MODIFIED, nullable = false)
    // var modified: Instant? = null,
    // ```
    //
    // 위와 같이 사용하는 상태에서 다시 `./gradlew app:build`를 싱행하면 아래와 같은 에러가 발생합니다.
    //
    // ```
    // error: element value must be a constant expression
    // @jakarta.persistence.Column(name = null, nullable = false)
    //                                    ^
    // ```
    // 이미 완전히 빌드된 라이브러리(JAR 등)의 `public static final` 필드는 문제없이 컴파일 타임 상수로 간주됩니다.
    // 그런데 이 경우 `null`이 되는 것을 보면, 코틀린 컴파일러가 메타 모델의 상수를 컴파일 타임 상수로 처리하지 못하는 것으로 보입니다.
    // 따라서 메타 모델 생성하여 사용하지 않도록 주석 처리합니다.
    // kapt(libs.db.hibernate.jpamodelgen)
    implementation(libs.db.hibernate.validator) // validator
    implementation(libs.jakarta.el.api) // validator

    testImplementation(libs.spring.boot.starter.test)

    testRuntimeOnly(libs.junit.platform.launcher)

    // KProperty 객체 정보를 더 자세하게 출력하기 위해 추가(없으면 'Kotlin reflection is not available' 출력)
    implementation(libs.kotlin.reflect)
    // Use the Kotlin JUnit 5 integration.
    testImplementation(libs.kotlin.test.junit5)

    // Use the JUnit 5 integration.
    testImplementation(libs.junit.jupiter.engine)

    // This dependency is used by the application.
    implementation(libs.guava)
}

idea {
    module {
        // https://discuss.gradle.org/t/how-do-i-force-gradle-to-download-dependency-sources/34726/2
        isDownloadSources = true
        isDownloadJavadoc = true
    }
}

// 어노테이션 프로세서가 필요한 경우 설정합니다.
// kapt {
//     correctErrorTypes = true
//     generateStubs = true
// }

tasks.withType<Test> {
    useJUnitPlatform()
}

tasks.withType<JavaCompile> {
    options.compilerArgs.add("-verbose")
    // 어노테이션 프로세서로 생성된 소스 파일들이 위치할 디렉토리를 설정합니다.
    // 생성된 소스 파일을 저장하고 이를 컴파일 경로에 포함시킵니다.
    // - https://docs.gradle.org/current/dsl/org.gradle.api.tasks.compile.CompileOptions.html#org.gradle.api.tasks.compile.CompileOptions:generatedSourceOutputDirectory
    // options.generatedSourceOutputDirectory.set(file("${layout.buildDirectory}/generated"))
}

tasks.named<Test>("test") {
    // Use JUnit Platform for unit tests.
    useJUnitPlatform()
}

tasks.register("sourceSets") {
    val mainSourceSet: SourceSet = sourceSets["main"]!!

    println("App source set names: ${sourceSets.names}") // [main, test]
    mainSourceSet.java {
        println(this.displayName)
        this.asFileTree.forEach { println("- ${it.absolutePath}") }
    }
    mainSourceSet.kotlin {
        println(this.displayName)
        this.asFileTree.forEach { println("- ${it.absolutePath}") }
    }
    println("annotationProcessorPath: ${mainSourceSet.annotationProcessorPath.asPath}")
    mainSourceSet.annotationProcessorPath.asFileTree.forEach { println("- ${it.absolutePath}") }

    println("runtimeClasspath:")
    mainSourceSet.runtimeClasspath.asFileTree.forEach { println("- ${it.absolutePath}") }
}
/**
 * Commands:
 *
 *    # packaging jar file
 *    ./gradlew -Ppkg=tmp pkgJar
 *    # At project root
 *    (mkdir -p tmp/TmpKt && cd tmp/TmpKt && jar xvf ../../app/build/libs/tmp-example.jar)
 */
tasks.register<Jar>("pkgJar") {
    /**
     * gradle 빌드 라이프사이클은 세 단계를 거칩니다.
     * 1. 초기화:
     *   `settings.gradle(.kts)` 파일을 찾아서 [`Settings`](https://docs.gradle.org/current/dsl/org.gradle.api.initialization.Settings.html) 인스턴스를 생성하고,
     *   설정을 평가하여 빌드를 구성하는 프로젝트(및 포함된 빌드)를 결정합니다.
     *   모든 프로젝트에 대해 [`Project`](https://docs.gradle.org/current/dsl/org.gradle.api.Project.html) 인스턴스를 생성합니다.
     * 2. 구성:
     *   빌드에 참여하는 모든 프로젝트의 빌드 스크립트 `build.gradle(.kts)`를 평가하고, 요청된 작업들에 대한 작업 그래프를 생성합니다.
     *   빌드 스크립트의 코드가 실행되면서 작업의 속성을 설정하고 액션을 등록하는 등 작업을 구성하지만, 액션 자체는 실행되지 않습니다.
     * 3. 실행:
     *   선택된 작업들을 스케쥴하고 실행합니다. 작업 간의 의존성이 실행 순서를 결정합니다.
     *   작업들의 실행은 병렬로 이뤄질 수 있습니다.
     *
     * 따라서 액션 실행이 아니더라도 구성 단계에서 스크립트의 코드 실행이 이뤄지고, 이에 따라 아직 전달되지 않은 속성으로 인해 익셉션이 발생할 수 있습니다.
     *
     * References:
     * - https://docs.gradle.org/current/userguide/build_lifecycle.html#sec:build_phases
     * - https://docs.gradle.org/current/userguide/implementing_custom_tasks.html#task_actions
     * - https://docs.gradle.org/current/userguide/organizing_tasks.html
     */
    doFirst {
        val targetPackage = findProperty("pkg") as String?
        check(!targetPackage.isNullOrBlank()) {
            "Error: 'pkg' property is required but was not provided.\n" +
                    "Use `-Ppkg=<TARGET_PACKAGE_PATH>`\n" +
                    "Example:\n\n" +
                    "\t./gradlew -Ppkg=safety pkgJar"
        }
        archiveBaseName.set("$targetPackage-example")
        val mainSourceSet = sourceSets.main.get()
        println("mainSourceSet.output: ${mainSourceSet.output}")
        from(mainSourceSet.output) {
            include("$targetPackage/**") // targetPackage 패키지의 클래스 파일만 포함
        }
        duplicatesStrategy = DuplicatesStrategy.EXCLUDE
    }
}

/**
 * `application` 블록은 프로젝트가 로드될 때 즉시 평가됩니다.
 * 즉, 특정 작업 실행 여부와 상관없이 항상 초기화 시점에서 실행됩니다.
 * - JVM 애플리케이션 실행과 관련된 기능을 제공하므로 Java 코드 컴파일에 필수적인 `java` 플러그인을 암시적으로 적용
 * - 이떄 `main` 소스 집합(source set)이 실질적으로 소스 코드와 실행할 애플리케이션(응용 프로그램)의 실제 내용
 *
 * References:
 * - https://docs.gradle.org/current/userguide/application_plugin.html
 */
application {
    gradle.taskGraph.whenReady {
        if (this.hasTask(":app:run")) {
            // mainClass 속성을 사용하도록 합니다.
            val targetMainClass = findProperty("mainClass") as String?
            check(!targetMainClass.isNullOrBlank()) {
                "Error: 'mainClass' property is required but was not provided." +
                        "Use `-PmainClass=<QUALIFIED_CLASS_NAME>`" +
                        "Example:\n\n" +
                        "\t./gradlew -PmainClass=safety.SafetyExampleKt run"
            }
            mainClass.set(targetMainClass)
        }

        if (listOf(":app:bootJar", ":app:bootRun").any { hasTask(it) }) {
            mainClass = "spring.SpringMainKt"
        }
    }
}

tasks.withType<JavaExec> {
    /**
     * Gradle은 주로 자동화된 빌드 도구로 사용되며, 빌드 작업은 대부분 사용자와의 상호작용 없이 실행됩니다.
     * 따라서, 기본적으로 사용자 입력을 필요로 하지 않는 빈 입력 스트림이 기본값입니다.
     * 이를 통해 빌드 프로세스가 중단되지 않고 자동으로 실행될 수 있도록 설계되었다고 합니다.
     *
     * 예를 들어, CI/CD 시스템이나 터미널 환경이 아닌 다른 방식으로 Gradle이 호출될 때 표준 입력이 필요 없는 경우가 있습니다.
     *
     * [readlnOrNull] 등을 사용하려면, 표준 입력이 명시적으로 설정해야 합니다.
     *
     * References:
     * - https://docs.gradle.org/current/dsl/org.gradle.api.tasks.JavaExec.html#org.gradle.api.tasks.JavaExec:standardInput
     */
    standardInput = System.`in`
    /**
     * `pkgJar` 작업을 수행하면 `/path/to/project/app/build/libs` 경로에 `<NAME>-example.jar` 아카이브 파일이 생성됩니다.
     * 해당 `jar` 파일들을 런타임 시에 스캔할 수 있도록 클래스패스에 추가합니다.
     *
     * Gradle에서 `compileClasspath`와 `runtimeClasspath`는 서로 다른 목적을 가지고 있습니다.
     * - `compileClasspath`:
     *      소스 코드 컴파일 시에만 필요한 의존성을 포함하고, 실행 단계에서 중복될 수 있습니다.
     * - `runtimeClasspath`:
     *      애플리케이션이 실행될 때 필요한 필요한 모든 런타임 의존성(클래스와 라이브러리)를 포함합니다.
     *
     * 따라서 여기서 [SourceSet.getCompileClasspath] 아닌 [SourceSet.getRuntimeClasspath]만 고려합니다.
     *
     * References:
     * - [Some other simple source set examples](https://docs.gradle.org/current/userguide/java_plugin.html#sec:some_source_set_examples)
     * - [Understanding dependency configurations](https://docs.gradle.org/current/userguide/dependency_configurations.html#sec:what-are-dependency-configurations)
     * - [Caching Java projects - Java compilation](https://docs.gradle.org/current/userguide/caching_java_projects.html#java_compilation)
     * - [Gradle User Home Directory](https://docs.gradle.org/current/userguide/directory_layout.html#dir:gradle_user_home)
     * - [The Dependency Cache](https://docs.gradle.org/current/userguide/dependency_resolution.html#sec:dependency_cache)
     * - [Working with Files](https://docs.gradle.org/current/userguide/working_with_files.html)
     */
    val newClasspath = sourceSets.main.get().runtimeClasspath + fileTree("build/libs") { include("**/*.jar") }
    classpath = newClasspath

    /**
     * JPMS는 기본적으로 모듈 간의 내부 구현을 노출하지 않습니다.
     * 즉, `jdk.internal.net.http` 패키지는 외부 모듈이나 익명 모듈에서 접근할 수 없습니다.
     * 따라서 리플렉션으로 접근하려고 하면 아래와 같은 에러가 발생합니다.
     * ```
     * Unable to make field final jdk.internal.net.http.HttpClientImpl jdk.internal.net.http.HttpClientFacade.impl accessible: module java.net.http does not "opens jdk.internal.net.http" to unnamed module @dbf57b3
     * ```
     *
     * 하지만 `--add-opens`라는 JVM 옵션을 통해 특정 패키지를 open할 수 있습니다.
     * ```
     * --add-opens <source-module>/<package>=<target-module>(,<target-module>)*
     * ```
     *
     * `ALL-UNNAMED`를 사용하면 해당 소스 패키지가 모든 익명 모듈에 export 됩니다.
     * ```
     * --add-opens java.net.http/jdk.internal.net.http=ALL-UNNAMED
     * ```
     *
     * References:
     * - [JEP 403: Strongly Encapsulate JDK Internals](https://openjdk.org/jeps/403)
     * - [JEP 261: Module System - Breaking encapsulation](https://openjdk.org/jeps/261#Breaking-encapsulation)
     * - [7 Migrating From JDK 8 to Later JDK Releases](https://docs.oracle.com/en/java/javase/17/migrate/migrating-jdk-8-later-jdk-releases.html)
     */
    jvmArgs("--add-opens", "java.net.http/jdk.internal.net.http=networks")
}
