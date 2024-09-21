# spring init

- [spring init](#spring-init)
    - [spring initializer](#spring-initializer)
        - [Project Metadata](#project-metadata)
            - [Artifact](#artifact)
    - [Dependency Management](#dependency-management)
    - [구성 파일](#구성-파일)
        - [`rootpath/settings.gradle.kts`](#rootpathsettingsgradlekts)
        - [`rootpath/build.gradle.kts`](#rootpathbuildgradlekts)
    - [실행 가능한 jar 파일 만들기](#실행-가능한-jar-파일-만들기)
    - [코드 구조화하기](#코드-구조화하기)
    - [Deploying Spring Boot Applications](#deploying-spring-boot-applications)
        - [systemd 서비스로 관리](#systemd-서비스로-관리)

## [spring initializer](https://start.spring.io/)

- Project

    - Gradle - Groovy: Groovy DSL을 사용하는 Gradle 프로젝트
    - Gradle - Kotlin: Kotlin DSL을 사용하는 Gradle 프로젝트
    - Maven: XML 기반의 프로젝트 객체 모델(POM)을 사용하는 Maven 프로젝트

- Language

    Java, Kotlin, Groovy

- Spring Boot

    ```sh
    major.minor.patch(SNAPSHOT | M2)

    3.4.0(SNAPSHOT): 개발 중인 버전
    3.4.0(M2): 마일스톤 릴리스

    3.3.4(SNAPSHOT)

    3.3.3: 안정화된 릴리스 버전
    ```

- Project Metadata

    - Group: 조직 또는 그룹의 고유 식별자(ex: `com.example`)

    - Artifact: 프로젝트의 고유 식별자(ex: `customer-service`)

    - Name: 프로젝트의 표시 이름(ex: `CustomerService`)

    - Description: 프로젝트에 대한 간단한 설명

    - Package name: 기본 패키지 이름으로 보통 Group ID와 Artifact ID를 조합하여 생성(ex: `com.example.customer`)

- Packaging

    - Jar: **J**ava **Ar**chive, 실행 가능한 Java 애플리케이션
    - War: **W**eb Application **Ar**chive, 웹 서버에 배포 가능한 웹 애플리케이션

- Java

### Project Metadata

#### Artifact

Artifact 이름은 프로젝트의 빌드 산출물(예: JAR 파일, WAR 파일)의 식별자로 사용됩니다.

Artifact 이름의 역할은 다음과 같습니다:
1. Artifact 이름은 빌드된 산출물을 고유하게 식별합니다.

    예를 들어, 여러 프로젝트에서 동일한 그룹 ID를 공유하더라도 각 프로젝트의 Artifact 이름이 다르면 산출물을 구분할 수 있습니다.

2. Maven이나 Gradle 같은 빌드 도구에서 의존성을 관리할 때, Artifact 이름을 사용해 특정 프로젝트의 빌드 산출물을 참조합니다.

    예를 들어, 다른 프로젝트에서 이 산출물을 의존성으로 추가할 때, `groupId:artifactId:version` 형식으로 참조됩니다.

3. 동일한 Artifact 이름을 가진 프로젝트의 여러 버전이 존재할 수 있습니다. Artifact 이름과 버전을 함께 사용하면 프로젝트의 특정 버전을 명확히 식별하고 관리할 수 있습니다.

4. CI/CD 파이프라인에서 Artifact 이름을 사용해 빌드된 산출물을 저장소에 배포하거나 배포된 산출물을 다운로드해 사용하는 등의 작업이 가능합니다. 이름이 명확하지 않다면 이러한 작업이 복잡해질 수 있습니다.

Artifact 이름은 보통 프로젝트 이름을 기반으로 합니다.
일반적으로 소문자로 작성하고, 단어 사이를 하이픈(`-`)으로 구분되고, 다른 프로젝트와 중복되지 않도록 유니크해야 합니다.

예를 들어, 프로젝트 이름이 `CustomerService`라면, Artifact 이름도 보통 `customer-service` 또는 `customer-service-api` 등으로 설정합니다.

또한 Artifact 이름은 버전 정보와 함께 사용되므로, 이름이 해당 버전의 내용을 잘 나타낼 수 있도록 신중히 선택하는 것이 중요합니다.

예를 들어, "customer-service-v2"와 같이 버전 이름을 포함하거나, 버전은 따로 관리하고 Artifact 이름은 고정된 형태로 유지할 수도 있습니다.

Maven 프로젝트 경우 다음과 같은 설정이 pom.xml 파일로 구성됩니다.

```xml
<!--
    `artifactId`는 `customer-service`이며,
    버전 `1.0.0`의 `customer-service`라는 프로젝트의 산출물을 나타냅니다.
-->
<groupId>com.example</groupId>
<artifactId>customer-service</artifactId>
<version>1.0.0</version>
```

gradle 경우에는 다음과 같이 간소화 됩니다.

```gradle
group = 'com.example'
artifact = 'customer-service'
version = '1.0.0'
```

## [Dependency Management](https://docs.spring.io/spring-boot/reference/using/build-systems.html#using.build-systems.dependency-management)

Spring Boot의 각 릴리스에서는 지원하는 종속성의 선별된 목록을 제공합니다.
실제로는 Spring Boot에서 이러한 종속성을 관리하므로 빌드 구성에서 이러한 종속성에 대한 버전을 제공할 필요가 없습니다.
Spring Boot 자체를 업그레이드하면 이러한 종속성도 일관된 방식으로 업그레이드됩니다.

## 구성 파일

### `rootpath/settings.gradle.kts`

### `rootpath/build.gradle.kts`

## 실행 가능한 jar 파일 만들기

```sh
gradle bootJar
```

생성된 jar 파일 내부를 보고 싶다면 `jar tvf` 를 사용합니다.

```sh
jar tvf build/libs/backend-0.0.1-SNAPSHOT.jar
```

- `t`: 아카이브에 대한 목차를 나열합니다.
- `v`: verbose
- `f`: 아카이브 파일 이름입니다. 생략할 경우 작업에 따라 stdin 또는 stdout이 사용됩니다.

```log
     0 Sat Aug 31 11:07:44 KST 2024 META-INF/
   432 Sat Aug 31 11:07:44 KST 2024 META-INF/MANIFEST.MF
     0 Fri Feb 01 00:00:00 KST 1980 META-INF/services/
    66 Fri Feb 01 00:00:00 KST 1980 META-INF/services/java.nio.file.spi.FileSystemProvider
     0 Fri Feb 01 00:00:00 KST 1980 org/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/
   461 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/JarEntriesStream$InputStreamSupplier.class
  3815 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/JarEntriesStream.class
  1617 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/ManifestInfo.class
  5125 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/MetaInfVersionsInfo.class
  2025 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$JarEntriesEnumeration.class
  2584 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$JarEntryInflaterInputStream.class
  3611 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$JarEntryInputStream.class
  7108 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$NestedJarEntry.class
  1197 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$RawZipDataInputStream.class
  2138 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile$ZipContentEntriesSpliterator.class
 15710 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFile.class
  6568 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/NestedJarFileResources.class
  4916 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/SecurityInfo.class
  1659 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jar/ZipInflaterInputStream.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jarmode/
   293 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/jarmode/JarMode.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/
   316 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/Archive$Entry.class
  4561 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/Archive.class
  4781 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/ClassPathIndexFile.class
  4411 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/ExecutableArchiveLauncher.class
  2134 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/ExplodedArchive$FileArchiveEntry.class
  5794 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/ExplodedArchive.class
  2088 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/JarFileArchive$JarArchiveEntry.class
 10046 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/JarFileArchive.class
   855 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/JarLauncher.class
  2440 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/JarModeRunner.class
  1586 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/LaunchedClassLoader$DefinePackageCallType.class
  8388 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/LaunchedClassLoader.class
  6693 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/Launcher.class
  3688 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/PropertiesLauncher$Instantiator$Using.class
  3650 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/PropertiesLauncher$Instantiator.class
 21958 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/PropertiesLauncher.class
  5277 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/SystemPropertyUtils.class
  1707 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/launch/WarLauncher.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/log/
  1408 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/log/DebugLogger$DisabledDebugLogger.class
  2610 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/log/DebugLogger$SystemErrDebugLogger.class
  1823 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/log/DebugLogger.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/
  1676 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/Handlers.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/
  2604 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/Canonicalizer.class
  6492 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/Handler.class
  3049 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarFileUrlKey.class
  2774 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrl.class
  1542 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrlClassLoader$OptimizedEnumeration.class
  8455 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrlClassLoader.class
  1469 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrlConnection$ConnectionInputStream.class
   750 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrlConnection$EmptyUrlStreamHandler.class
 12244 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/JarUrlConnection.class
  2075 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/LazyDelegatingInputStream.class
  1168 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/Optimizations.class
  1325 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarEntry.class
  3261 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarFile.class
  5756 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarFileFactory.class
  2467 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarFiles$Cache.class
  3857 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarFiles.class
   486 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarManifest$ManifestSupplier.class
  3388 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlJarManifest.class
  3181 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/jar/UrlNestedJarFile.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/
  1468 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/Handler.class
  5775 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/NestedLocation.class
  1590 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/NestedUrlConnection$ConnectionInputStream.class
  7567 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/NestedUrlConnection.class
  3892 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/protocol/nested/NestedUrlConnectionResources.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/util/
  3506 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/net/util/UrlDecoder.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/
  1918 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedByteChannel$Resources.class
  3487 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedByteChannel.class
  2797 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedFileStore.class
  8446 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedFileSystem.class
  9693 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedFileSystemProvider.class
  7441 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/NestedPath.class
  2016 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/nio/file/UriPathEncoder.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/ref/
   667 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/ref/Cleaner.class
  1375 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/ref/DefaultCleaner.class
     0 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/
  1459 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ByteArrayDataBlock.class
   231 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/CloseableDataBlock.class
  1042 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/DataBlock.class
  2444 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/DataBlockInputStream.class
  5932 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/FileDataBlock$FileAccess.class
   843 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/FileDataBlock$Tracker$1.class
   666 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/FileDataBlock$Tracker.class
  4398 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/FileDataBlock.class
  1678 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/NameOffsetLookups.class
  3114 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/VirtualDataBlock.class
  1631 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/VirtualZipDataBlock$DataPart.class
  6546 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/VirtualZipDataBlock.class
  4033 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/Zip64EndOfCentralDirectoryLocator.class
  5724 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/Zip64EndOfCentralDirectoryRecord.class
  9832 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipCentralDirectoryFileHeaderRecord.class
  6349 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipContent$Entry.class
  1432 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipContent$Kind.class
 12953 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipContent$Loader.class
  2295 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipContent$Source.class
 12611 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipContent.class
  4841 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipDataDescriptorRecord.class
  2105 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipEndOfCentralDirectoryRecord$Located.class
  6496 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipEndOfCentralDirectoryRecord.class
  5545 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipLocalFileHeaderRecord.class
  1474 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipString$CompareType.class
  7973 Fri Feb 01 00:00:00 KST 1980 org/springframework/boot/loader/zip/ZipString.class
     0 Sat Aug 31 11:07:44 KST 2024 BOOT-INF/
     0 Sat Aug 31 11:07:44 KST 2024 BOOT-INF/classes/
     0 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/
     0 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/aimpugn/
     0 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/aimpugn/backend/
     0 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/aimpugn/backend/Controller/
   624 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/aimpugn/backend/Controller/HomeController.class
   746 Sat Aug 31 10:58:38 KST 2024 BOOT-INF/classes/me/aimpugn/backend/BackendApplication.class
    32 Sat Aug 31 11:07:44 KST 2024 BOOT-INF/classes/application.properties
     0 Sat Aug 31 11:07:44 KST 2024 BOOT-INF/lib/
1963470 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-boot-autoconfigure-3.3.3.jar
1626131 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-boot-3.3.3.jar
 26141 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/jakarta.annotation-api-2.1.1.jar
1046956 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/spring-webmvc-6.1.12.jar
1901655 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/spring-web-6.1.12.jar
1305612 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-context-6.1.12.jar
417167 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-aop-6.1.12.jar
862392 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-beans-6.1.12.jar
304339 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/spring-expression-6.1.12.jar
1882923 Sat Aug 31 10:27:02 KST 2024 BOOT-INF/lib/spring-core-6.1.12.jar
334352 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/snakeyaml-2.2.jar
296266 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/logback-classic-1.5.7.jar
 23081 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/log4j-to-slf4j-2.23.1.jar
  6351 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/jul-to-slf4j-2.0.16.jar
 24539 Sat Aug 31 10:27:00 KST 2024 BOOT-INF/lib/spring-jcl-6.1.12.jar
132264 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-datatype-jsr310-2.17.2.jar
 10338 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-module-parameter-names-2.17.2.jar
 78492 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-annotations-2.17.2.jar
581927 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-core-2.17.2.jar
 36161 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-datatype-jdk8-2.17.2.jar
1649454 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/jackson-databind-2.17.2.jar
281744 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/tomcat-embed-websocket-10.1.28.jar
3553468 Sat Aug 31 10:54:10 KST 2024 BOOT-INF/lib/tomcat-embed-core-10.1.28.jar
261433 Sat Aug 31 10:54:08 KST 2024 BOOT-INF/lib/tomcat-embed-el-10.1.28.jar
 71767 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/micrometer-observation-1.13.3.jar
615004 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/logback-core-1.5.7.jar
 69435 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/slf4j-api-2.0.16.jar
342535 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/log4j-api-2.23.1.jar
 47508 Sat Aug 31 10:28:40 KST 2024 BOOT-INF/lib/micrometer-commons-1.13.3.jar
 53159 Sat Aug 31 20:07:44 KST 2024 BOOT-INF/lib/spring-boot-jarmode-tools-3.3.3.jar
  1286 Sat Aug 31 20:07:44 KST 2024 BOOT-INF/classpath.idx
   212 Sat Aug 31 20:07:44 KST 2024 BOOT-INF/layers.idx
```

`java -jar` 사용하여 애플리케이션을 실행합니다.

```sh
java -jar build/libs/myproject-0.0.1-SNAPSHOT.jar
```

## 코드 구조화하기

## [Deploying Spring Boot Applications](https://docs.spring.io/spring-boot/how-to/deployment/index.html)

### [systemd 서비스로 관리](https://docs.spring.io/spring-boot/how-to/deployment/installing.html)

`/etc/systemd/system/myapp.service` 파일 생성해서 관리합니다.
`myapp.service` 기본 예제는 다음과 같습니다.

```ini
[Unit]
Description=myapp
After=syslog.target network.target

[Service]
User=myapp
Group=myapp

Environment="JAVA_HOME=/path/to/java/home"

ExecStart=${JAVA_HOME}/bin/java -jar /var/myapp/myapp.jar
ExecStop=/bin/kill -15 $MAINPID
SuccessExitStatus=143

[Install]
WantedBy=multi-user.target
```

또는 다음과 같이 `Restart`, `RestartSec`를 설정할 수도 있습니다.

```ini
[Unit]
Description=My Spring Boot Application
After=syslog.target

[Service]
User=ubuntu
ExecStart=/usr/bin/java -jar /path/to/backend.jar
SuccessExitStatus=143
TimeoutStopSec=10
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

`sudo systemctl start myapp` 등을 통해 서비스를 실행할 수 있습니다.
아래는 graceful shutdown 후 health 체크를 통해 정상적으로 작동하는지 확인하는 스크립트입니다.

```bash
#!/bin/bash

# 1. 현재 실행 중인 애플리케이션 종료
if sudo systemctl is-active --quiet myapp.service; then
    echo "Stopping the current application..."
    sudo systemctl stop myapp.service
    sleep 10  # graceful shutdown을 위한 대기
fi

# 2. 새 버전 JAR 파일 복사
sudo cp /path/to/new/backend.jar /path/to/production/backend.jar

# 3. 새 버전 시작
echo "Starting the new application version..."
sudo systemctl start myapp.service

# 4. 헬스 체크
max_retries=10
count=0
while [ $count -lt $max_retries ]; do
    if curl -s http://localhost:8080/actuator/health | grep -q "UP"; then
        echo "Application is healthy"
        exit 0
    fi
    echo "Waiting for application to become healthy..."
    sleep 5
    count=$((count+1))
done

echo "Application failed to start properly"
exit 1
```
