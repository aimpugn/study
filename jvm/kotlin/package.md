# Package

- [Package](#package)
    - [`/subproject/` 디렉토리 추가](#subproject-디렉토리-추가)

## `/subproject/` 디렉토리 추가

서브프로젝트 디렉터리들을 `subproject/` 하위로 이동하여 작업 시 네비게이션이 편하도록 변경

```ASIS
project/
    .git/
    .github/
    application/
    buildSrc/
    docs/
    domain/
    gradle/
    infrastructure/
    interface/
    presenstation/
    .gitignore
    .gitmodule
    build.gradle.kts
```

```TOBE
project/
    .git/
    .github/
    buildSrc/
    docs/
    gradle/
    subproject/
        application/
        domain/
        infrastructure/
        interface/
        presenstation/
    .gitignore
    .gitmodule
    build.gradle.kts
```
