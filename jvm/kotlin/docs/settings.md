# Settings

- [Settings](#settings)
  - [gradle](#gradle)
    - [`tasks.getByName<Jar>("jar")`](#tasksgetbynamejarjar)
  - [compiler options](#compiler-options)
    - [`--Xjsr305=strict`](#--xjsr305strict)

## gradle

### `tasks.getByName<Jar>("jar")`

`bootJar` 또는 `bootWar` tasks 설정되면, `jar` 또는 `war` 작업은 비활성화된다.
두 executable 아카이브 및 normal 아카이브를 빌드하도록 구성할 수 있다.

[Differences Between JAR and WAR Packaging](https://www.baeldung.com/java-jar-war-packaging)

- `jar`: Java Archive
  - 컴파일된 자바 라이브러리와 애플리케이션의 `.class` 파일과 자원(resouces)의 압축한 버전을 갖는 zip file
  - files
    - libraries
    - resources
    - metadata

```tree
META-INF/
    MANIFEST.MF
com/
    baeldung/
        MyApplication.class
```

- `war`: Web Application Archive | Web Application Resource
  - `Servlet/JSP container`에 배포할 수 있는 웹 애플리케이션을 패키징하기 위해 사용

```tree
META-INF/
    MANIFEST.MF
WEB-INF/
    web.xml
    jsp/
        helloWorld.jsp
    classes/
        static/
        templates/
        application.properties
    lib/
        // *.jar files as libs
```

## compiler options

### `--Xjsr305=strict`

Java는 type-system에서 null-safety를 허용하지 않지만 Spring Framework는 org.springframework.lang 패키지에 선언된 도구 친화적인(Tooling-friendly) Annotation을 통해 전체 Spring Framework API의 null-safety를 제공

기본적으로 Kotlin에서 사용되는 Java API의 타입은 null 체크가 완화된 [플랫폼](https://kotlinlang.org/docs/java-interop.html#null-safety-and-platform-types) 타입으로 인식

[JSR 305 annotations](https://kotlinlang.org/docs/java-interop.html#jsr-305-support) + Spring Nullability Annotation에 대한 Kotlin 지원은 컴파일 타임에 null 관련 문제를 처리 할 수 있다는 이점을 가지고 Kotlin 개발자에게 전체 Spring Framework API에 대한 null-safety를 제공

[JSR-305 custom nullability qualifiers](https://github.com/Kotlin/KEEP/blob/master/proposals/jsr-305-custom-nullability-qualifiers.md#jsr-305-custom-nullability-qualifiers)
[JSR 305: Annotations for Software Defect Detection](https://jcp.org/en/jsr/detail?id=305)
