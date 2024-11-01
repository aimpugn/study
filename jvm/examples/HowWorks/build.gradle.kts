/**
 * Kotlin DSL scripts
 * - https://docs.gradle.org/current/userguide/kotlin_dsl.html#sec:scripts
 *
 * 버전 관리는 version catalog를 따릅니다.
 * - https://docs.gradle.org/current/userguide/platforms.html#sec:version-catalogs
 * - https://docs.gradle.org/current/userguide/dependency_management_basics.html#version_catalog
 */
plugins {
    /**
     * Java 프로젝트를 빌드하고 관리하기 위한 표준 작업과 설정을 자동으로 추가합니다.
     * 이 플러그인은 built-in 플러그
     * `java` 플러그인을 적용하면 Gradle은 Java 컴파일 및 테스트, JAR 파일 생성과 관련된 작업들을 추가합니다.
     * - `compileJava` 작업이 추가되어 Java 소스 코드를 컴파일합니다.
     * - `test` 작업이 추가되어 JUnit과 같은 테스트 프레임워크를 실행할 수 있습니다.
     * - `jar` 작업이 추가되어 컴파일된 클래스 파일과 종속성을 포함하는 JAR 파일을 생성합니다.
     * - 기본 소스셋 경로(`src/main/java` 및 `src/test/java`)를 자동으로 인식합니다.
     *
     * 이 외에도 `implementation`, `compileOnly` 등 여러 의존성 구성들을 추가합니다.
     *
     * References:
     * - https://docs.gradle.org/current/userguide/java_plugin.html
     * - https://docs.gradle.org/current/dsl/org.gradle.api.plugins.JavaPluginExtension.html
     * - https://docs.gradle.org/current/javadoc/org/gradle/api/plugins/JavaPlugin.html
     */
    java

    /**
     * kotlin("jvm") 플러그인은 Kotlin 코드를 JVM 바이트코드로 컴파일하기 위해 필요한 Kotlin 컴파일러와 관련 설정을 Gradle에 추가합니다.
     * - Kotlin 컴파일러가 Kotlin 코드를 JVM이 실행할 수 있는 바이트코드로 컴파일하도록 설정하고, 이를 통해 같은 프로젝트 내 Java 코드와 Kotlin 코드를 함께 사용 가능
     * - Kotlin 코드를 JVM 바이트코드로 컴파일할 때 사용할 Kotlin 컴파일러를 설정
     * - `src/main/kotlin`과 `src/test/kotlin`을 각각 애플리케이션과 테스트 소스 디렉터리로 인식
     * - `compileKotlin` 및 `compileTestKotlin` 작업을 자동으로 추가하여 Kotlin 컴파일러를 통해 Kotlin 파일을 JVM 바이트코드로 변환
     * - `org.jetbrains.kotlin:kotlin-stdlib` 라이브러리를 기본적으로 컴파일 및 런타임 클래스패스에 포함시켜, Kotlin 표준 라이브러리 함수와 기능을 사용할 수 있도록 지원
     *
     * 다음 코드와 동일합니다:
     * ```
     * kotlin("jvm") version "2.0.0"
     * ```
     *
     * Gradle은 기본적으로 Kotlin DSL(.gradle.kts 파일)을 통해 Kotlin을 빌드 스크립트 언어로 사용할 수 있도록 지원하지만,
     * Kotlin 코드 컴파일 자체는 기본 제공 기능이 아닌 `org.jetbrains.kotlin.jvm` 플러그인을 통해 가능합니다.
     * 이를 통해 Java 프로젝트와 유사한 방식으로 필요한 경우에만 Kotlin 컴파일러와 관련 라이브러리를 추가하여 사용하는 형태입니다.
     */
    alias(libs.plugins.kotlin.jvm)
}

group = "me.aimpugn"
version = "0.0.1-SNAPSHOT"

/**
 * 멀티 모듈 프로젝트 시 하위 프로젝트들에 대해 설정할 수 있습니다.
 *
 * References:
 * - https://docs.gradle.org/current/userguide/intro_multi_project_builds.html
 * - https://discuss.gradle.org/t/multi-module-project-how-where-to-apply-plugins
 */
subprojects {
    /**
     * Note: project 스코프에서는 libs 대신 rootProject.libs 를 사용해야 합니다
     * - https://github.com/gradle/gradle/issues/16708
     * - https://github.com/gradle/gradle/issues/16634#issuecomment-806537122
     *
     * ```
     * println("kotlin.jvm pluginId: ${rootProject.libs.plugins.kotlin.jvm.get().pluginId}") // kotlin.jvm pluginId: org.jetbrains.kotlin.jvm
     * ```
     *
     * 하위 프로젝트에 플러그인들을 실제로 적용할 수 있습니다.
     *
     * ```
     * listOf(
     *     rootProject.libs.plugins.core.java,
     *     rootProject.libs.plugins.core.application,
     *     rootProject.libs.plugins.kotlin.jvm,
     *     rootProject.libs.plugins.spring.boot,
     *     rootProject.libs.plugins.spring.dependency.management,
     * ).forEach { pluginDependencyProvider ->
     *     apply(plugin = pluginDependencyProvider.get().pluginId)
     * }
     * ```
     *
     * 하지만 `allprojects`, `subprojects` 사용은 bad practice 라는 의견이 있어서 각 프로젝트별로 plugin을 적용합니다.
     * - `allprojects`나 `subprojects` 블록을 통해 루트에서 모든 서브 프로젝트를 설정하게 되면,
     *   각 프로젝트 간 설정이 얽히게 되어 프로젝트의 독립성을 잃게 되고, 재사용성과 유연성에 좋지 않은 영향을 끼친다고 합니다.
     * - `allprojects`와 `subprojects`를 통해 모든 설정을 한 곳에서 관리하려고 하면 빌드 스크립트가 복잡해지고 가독성이 떨어질 수 있고,
     *   이 때문에 오류 수정이 어려워지고 유지보수 비용이 증가하는 원인이 될 수 있다고 합니다.
     *
     * References:
     * - https://docs.gradle.org/current/userguide/sharing_build_logic_between_subprojects.html
     *
     * Opinions:
     * 대부분 같은 사람 의견이지만...
     * - https://discuss.gradle.org/t/is-it-possible-to-register-third-party-platforms-in-settings/46352/2
     * - https://discuss.gradle.org/t/java-toolchain-present-in-allprojects-block-is-no-applied-to-subproject-unless-apply-plugin-java-is-in-the-block/46938/2
     * - https://discuss.gradle.org/t/how-to-understand-the-subproject-dependence/47751/2
     */


    /**
     * Java 플러그인이 적용된 하위 프로젝트에만 `java` 설정을 적용하도록 합니다.
     */
    plugins.withType<JavaPlugin> {
        java {
            toolchain {
                languageVersion.set(JavaLanguageVersion.of(21))
            }
        }
    }
}

tasks.register("testActionType") {
    exec {
        commandLine("/bin/bash", "-c", "echo \"By exec coomandLine: $(pwd)\"")

        /**
         * Kotlin DSL에서 [org.gradle.api.Project.java], [subprojects], [exec] 등은 모두
         * `[Action]<in T>`을 매개변수로 받는 메서드입니다.
         *
         * Kotlin에서 단일 추상 멤버 함수만 있는 인터페이스를 함수형 인터페이스 또는 '단일 추상 메서드(SAM, Single Abstract Method) 인터페이스'라고 합니다.
         * [Action] 타입은
         * - 단일 메서드 `execute`를 갖는 SAM(Single Abstract Method) 인터페이스이며,
         * - SAM 인터페이스를 람다 표현식(Kotlin) 또는 클로저(Groovy)의 대상으로 지정하여, 단일 파라미터가 호출의 암시적 리시버가 되도록 하는 [HasImplicitReceiver] 어노테이션이 붙어 있습니다.
         *
         * 이때 일반 람다와 다르게 '리시버가 있는 람다'로 동작합니다.
         * '리시버 있는 람다'는 '특정 객체를 컨텍스트로 사용하는 람다'로, `this`를 통해 리시버 객체에 접근할 수 있습니다.
         *
         * 그렇다면 [Action]은 인터페이스인데 어떻게 구현 없이 바로 실행될까?
         * SAM 변환 때문에 가능합니다.
         * 즉, `{}` 블록이 바로 [Action] 인터페이스를 구현하는 코드입니다.
         *
         * References:
         * - ['리시버가 있는 람다'](https://kotlinlang.org/docs/lambdas.html#function-literals-with-receiver)
         * - [단일 추상 메서드(SAM, Single Abstract Method) 인터페이스](https://kotlinlang.org/docs/fun-interfaces.html)
         * - [SAM 변환](https://kotlinlang.org/docs/fun-interfaces.html#sam-conversions)
         * - https://kotlinlang.org/docs/sam-with-receiver-plugin.html
         */
        Action<String> {
            println("Action executed: ${this.uppercase()}")
        }.execute("test action")

        // isEvenAction1 와 isEvenAction2 는 같은 동작을 하는 코드입니다.
        object : Action<Int> {
            override fun execute(t: Int) {
                println("Number: ${t}, Result: ${t % 2 == 0}")
            }
        }.execute(2)

        Action<Int> {
            println("Number: ${this}, Result: ${this % 2 == 0}")
        }.execute(3)
    }
    println("After Exec done")
    // Output:
    //  Action executed: TEST ACTION
    //  Number: 2, Result: true
    //  Number: 3, Result: false
    //  By exec coomandLine: /Users/rody/VscodeProjects/study/jvm/examples/HowWorks
    //  After Exec done
}