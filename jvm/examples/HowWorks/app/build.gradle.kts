/**
 * 버전 관리는 version catalog를 따릅니다.
 * - https://docs.gradle.org/current/userguide/platforms.html#sec:version-catalogs
 * - https://docs.gradle.org/current/userguide/dependency_management_basics.html#version_catalog
 */
plugins {
    // Apply the org.jetbrains.kotlin.jvm Plugin to add support for Kotlin.
    alias(libs.plugins.kotlin.jvm)
    alias(libs.plugins.spring.boot)
    alias(libs.plugins.spring.dependency.management)

    java
    // Apply the application plugin to add support for building a CLI application in Java.
    application
}

group = "me.aimpugn"
version = "0.0.1-SNAPSHOT"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

repositories {
    mavenCentral()
}

dependencies {
    implementation(libs.spring.boot.starter)
    implementation(libs.spring.boot.starter.web)
    implementation(libs.kotlinx.coroutines.core)
    testImplementation(libs.spring.boot.starter.test)

    testRuntimeOnly(libs.junit.platform.launcher)

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

// tasks.withType<JavaCompile> {
//     options.fork = true
//     options.forkOptions.jvmArgs << '-verbose:class'
// }

// Apply a specific Java toolchain to ease working on different environments.
java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

application {
    // Define the main class for the application.
    mainClass = "me.aimpugn.backend.BackendApplicationKt"
}

tasks.named<Test>("test") {
    // Use JUnit Platform for unit tests.
    useJUnitPlatform()
}
