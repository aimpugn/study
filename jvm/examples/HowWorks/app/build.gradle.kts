plugins {
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
     * Spring 관련 플러그인
     */
    alias(libs.plugins.spring.boot)
    alias(libs.plugins.spring.dependency.management)
}

dependencies {
    implementation(libs.spring.boot.starter)
    implementation(libs.spring.boot.starter.web)
    implementation(libs.kotlinx.coroutines.core)
    testImplementation(libs.spring.boot.starter.test)

    testRuntimeOnly(libs.junit.platform.launcher)

    // KProperty 객체 정보를 더 자세하게 출력하기 위해 추가(없으면 'Kotlin reflection is not available' 출력)
    runtimeOnly(libs.kotlin.reflect)
    // Use the Kotlin JUnit 5 integration.
    testImplementation(libs.kotlin.test.junit5)

    // Use the JUnit 5 integration.
    testImplementation(libs.junit.jupiter.engine)

    // This dependency is used by the application.
    implementation(libs.guava)
}

tasks.withType<Test> {
    useJUnitPlatform()
}

tasks.withType<JavaCompile> {
    options.compilerArgs.add("-verbose")
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
            // 속성으로 받은 mainClass를 사용하거나, 기본값을 지정
            val targetMainClass = findProperty("mainClass") as String?
            check(!targetMainClass.isNullOrBlank()) {
                "Error: 'mainClass' property is required but was not provided."
            }
            mainClass.set(targetMainClass)
        }

    }
}

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
tasks.withType<JavaExec> {
    standardInput = System.`in`
}