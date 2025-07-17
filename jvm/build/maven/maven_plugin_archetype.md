# Maven Plugin archetype

- [Maven Plugin archetype](#maven-plugin-archetype)
    - [Maven Plugin archetype](#maven-plugin-archetype-1)
    - [`generate`](#generate)
    - [기타](#기타)

## [Maven Plugin archetype](https://maven.apache.org/archetype/maven-archetype-plugin/index.html)

archetype이라고 불리는 템플릿으로부터 Maven 프로젝트를 생성하는 것을 돕습니다.

## `generate`

```sh
mvn archetype:generate \
    -DgroupId=io.aimpugn.jv.file \
    -DartifactId=jv-file \
    -DarchetypeArtifactId=maven-archetype-quickstart \
    -DarchetypeVersion=1.5 \
    -DinteractiveMode=false
```

## 기타

- [archetype:generate](https://maven.apache.org/archetype/maven-archetype-plugin/generate-mojo.html)
