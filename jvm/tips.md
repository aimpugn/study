# Tips

- [Tips](#tips)
    - [\*-sources.jar 파일 내부 검색](#-sourcesjar-파일-내부-검색)
    - [`jar` 명령어로 특정 파일만 해제하여 보기](#jar-명령어로-특정-파일만-해제하여-보기)

## *-sources.jar 파일 내부 검색

```sh
#!/bin/bash

# ```
# search_jars "<PACKAGE_NAME_PATTERN>" "<PATTERN_TO_SEARCH_IN_JAR>"
# ```
#
# Example:
#
# ```
# ❯ search_jars "org.springframework" "ServletWebServerApplicationContextFactory"
# $HOME/.gradle/caches/modules-2/files-2.1/org.springframework.boot/spring-boot/3.3.4/1ecb958e8a4fa486e02d039d6324dd1c1b7aa77c/spring-boot-3.3.4-sources.jar:
# org.springframework.boot.web.servlet.context.ServletWebServerApplicationContextFactory
#
# $HOME/.gradle/caches/modules-2/files-2.1/org.springframework.boot/spring-boot/3.3.4/1ecb958e8a4fa486e02d039d6324dd1c1b7aa77c/spring-boot-3.3.4-sources.jar:
# class ServletWebServerApplicationContextFactory implements ApplicationContextFactory {
# ```
#
search_jars() {
    local packageName="$1"
    local pattern="$2"

    if [ -z "$packageName" ]; then
        echo "search_jars: packageName is empty"
        return 1
    fi

    if [ -z "$pattern" ]; then
        echo "search_jars: pattern is empty"
        return 1
    fi

    local gradleFilesPath="$HOME/.gradle/caches/modules-2/files-2.1"

    for file in $(cd "$gradleFilesPath" && (rg --files | rg "$packageName" | rg "sources.jar")); do
        search_jar "$gradleFilesPath/$file" "$pattern"
    done
}


# ```
# search_jar "<PATH_TO_SOURCE_JAR>" "<PATTERN_TO_SEARCH_IN_JAR>"
# ```
#
# Example:
#
# ```
# ❯ search_jar spring-core-6.1.13-sources.jar "Servlet"
# spring-core-6.1.13-sources.jar:
#  * @see org.springframework.web.context.support.StandardServletEnvironment
#
# spring-core-6.1.13-sources.jar:
#  * objects, {@code ServletContext} and {@code ServletConfig} objects (for access to init
#
# spring-core-6.1.13-sources.jar:
#      * creation time.  For example, a {@code ServletContext}-based property source
# ```
#
search_jar() {
    local jarFile="$1"
    local pattern="$2"

    if [ -z "$jarFile" ]; then
        echo "search_jar: jarFile is empty"
        return 1
    fi

    if [ -z "$pattern" ]; then
        echo "search_jar: pattern is empty"
        return 1
    fi

    # jar 파일 내용을 임시로 스트림으로 읽어서 검색
    unzip -p "$jarFile" \
        | rg --color always "$pattern" \
        | awk -v jar="$jarFile" '{print jar ":\n" $0 "\n"}'
}
```

## `jar` 명령어로 특정 파일만 해제하여 보기

jar 파일 내부 구조를 확인합니다.

```sh
jar tf my-app.jar
```

- `-t`, `--list`: List the table of contents for the archive
- `-f`, `--file=FILE`: The archive file name. When omitted, either stdin or stdout is used based on the operation

그리고 원하는 경로의 파일을 해제합니다.

```sh
# my-app.jar에서 META-INF/MANIFEST.MF 파일만 추출
jar xf my-app.jar META-INF/MANIFEST.MF
```
