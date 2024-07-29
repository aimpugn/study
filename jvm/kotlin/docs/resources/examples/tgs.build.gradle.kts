import com.google.protobuf.gradle.*
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    /* Core plugins */
    java
    idea

    /* Kotlin plugins */
    kotlin("jvm") version "1.6.10"
    kotlin("plugin.spring") version "1.6.10" apply false
    kotlin("kapt") version "1.6.10" apply false

    /* Spring plugins */
    id("org.springframework.boot") version "2.6.3" apply false
    id("io.spring.dependency-management") version "1.0.11.RELEASE" apply false

    /* Community plugins */
    id("com.google.protobuf") version "0.8.18" apply false
    id("org.jlleitschuh.gradle.ktlint") version "10.2.1" apply false
    id("org.jlleitschuh.gradle.ktlint-idea") version "10.2.1" apply false
    id("org.jetbrains.kotlinx.kover") version "0.6.1" // Kotlin code coverage
    id("org.barfuin.gradle.taskinfo") version "2.1.0" // Show gradle task info
}

allprojects {
    group = "finance.chai.gateway.transaction"
    version = "0.0.1-SNAPSHOT"

    apply {
        plugin("idea")
        plugin("java")
        plugin("org.jetbrains.kotlinx.kover")
    }

    configurations.all {
        resolutionStrategy.cacheChangingModulesFor(0, "minutes")
    }

    repositories {
        mavenCentral()
        maven {
            url = uri("https://maven.pkg.github.com/some-org/port-logger")
            credentials {
                username = System.getenv("PACKAGE_GITHUB_TOKEN")?.let {
                    if (it.isNotBlank()) "some_name-github"
                    else null
                } ?: System.getenv("CHAI_GPR_USERNAME")

                password = System.getenv("PACKAGE_GITHUB_TOKEN") ?: System.getenv("CHAI_GPR_TOKEN")
            }
        }
    }

    // For strict null-safety
    tasks.withType<KotlinCompile> {
        kotlinOptions {
            freeCompilerArgs = listOf("-Xjsr305=strict")
            jvmTarget = "17"
        }
    }

    extra["kotlin-coroutines.version"] = "1.6.0"
}

// The projects containing gRPC / Protobuf generated codes.
val grpcInterfaces = subprojects.filter { it.path.startsWith(":subproject:interface") }.toSet()

configure(subprojects - grpcInterfaces) {
    apply {
        plugin("org.jlleitschuh.gradle.ktlint")
        plugin("org.jlleitschuh.gradle.ktlint-idea")
        plugin("io.spring.dependency-management")
    }

    dependencies {
        // project/buildSrc/src/main/kotlin/Dependencies.kt
        // fun DependencyHandler.loggingImplementation() 정의
        loggingImplementation()

        implementation(kotlin("stdlib-jdk8"))
        implementation(kotlin("reflect"))
        implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:${Version.kotlinCoroutines}")
        implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor:${Version.kotlinCoroutines}")

        implementation("commons-io:commons-io:2.11.0") // Utils about file IO
        implementation("io.arrow-kt:arrow-core:${Version.arrowCore}")
        implementation("io.kotest.extensions:kotest-assertions-arrow:${Version.kotestAssertionsArrow}")
        implementation("io.konform:konform-jvm:${Version.konform}")
        implementation("org.springframework:spring-context")
        implementation("com.fasterxml.jackson.module:jackson-module-kotlin")

        testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:${Version.kotlinCoroutinesTest}")

        testImplementation("org.springframework.boot:spring-boot-starter-test")
        testImplementation("io.mockk:mockk:${Version.mockK}")
        testImplementation("org.assertj:assertj-core:${Version.assertj}")
        testImplementation("org.junit.jupiter:junit-jupiter-api")
        testRuntimeOnly("org.junit.jupiter:junit-jupiter-engine")

        testImplementation("io.kotest:kotest-runner-junit5:${Version.kotest}")
        testImplementation("io.kotest.extensions:kotest-extensions-spring:${Version.kotestSpring}")
        testImplementation("io.kotest:kotest-assertions-konform:${Version.kotestKonform}")
    }

    tasks.withType<Test> {
        useJUnitPlatform()
    }

    kover {
        // Enables report generation for internal subprojects
        htmlReport {
            onCheck.set(true)
        }

        xmlReport {
            onCheck.set(true)
        }
    }
}

configure(grpcInterfaces) {
    apply {
        java
        plugin("org.jetbrains.kotlin.jvm")
        plugin("com.google.protobuf")
    }

    tasks.getByName<Jar>("jar") {
        enabled = true
    }

    dependencies {
        api("com.google.protobuf:protobuf-java:${Version.protobuf}")
        api("com.google.protobuf:protobuf-kotlin:${Version.protobuf}")
        api("io.grpc:grpc-kotlin-stub:${Version.grpcKotlin}")
        implementation("io.grpc:grpc-stub:${Version.grpc}")
        implementation("io.grpc:grpc-protobuf:${Version.grpc}")

        if (JavaVersion.current().isJava9Compatible) {
            // Workaround for @javax.annotation.Generated
            // see: https://github.com/grpc/grpc-java/issues/3633
            compileOnly("javax.annotation:javax.annotation-api:1.3.1")
        }
    }

    sourceSets {
        main {
            proto {
                srcDirs("${project.name}-interface/protobuf")
            }
        }
    }

    // Kotlin wrapper 생성 방법
    // https://github.com/google/protobuf-gradle-plugin/issues/504
    // https://github.com/google/protobuf-gradle-plugin/issues/511
    protobuf {
        protoc {
            artifact = "com.google.protobuf:protoc:${Version.protobuf}"
        }

        plugins {
            id("kotlin")
            id("grpc") {
                // Apple Silicon 용 실행파일이 아직 없기 때문에 Intel 용 grpc-java 실행파일을 사용하여야 한다.
                artifact = if (isSiliconMac()) {
                    "io.grpc:protoc-gen-grpc-java:${Version.grpc}:${Version.macOsPlatform}"
                } else {
                    "io.grpc:protoc-gen-grpc-java:${Version.grpc}"
                }
            }
            id("grpckt") {
                artifact = "io.grpc:protoc-gen-grpc-kotlin:${Version.grpcKotlin}:jdk7@jar"
            }
        }

        generateProtoTasks {
            ofSourceSet("main").forEach {
                it.plugins {
                    id("grpc")
                    id("grpckt")
                    id("kotlin")
                }
            }
        }
    }
}

// Kover Configuration for multi-project build
// reference: https://github.com/Kotlin/kotlinx-kover
koverMerged {
    enable()
}

kotlin {
    jvmToolchain {
        (this as JavaToolchainSpec).languageVersion.set(JavaLanguageVersion.of("17"))
    }
}
