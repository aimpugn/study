# Psalm

- [Psalm](#psalm)
    - [psalm?](#psalm-1)
    - [기능](#기능)
        - [Mixed type warnings](#mixed-type-warnings)
        - [Intelligent logic checks](#intelligent-logic-checks)
        - [Property initialisation checks](#property-initialisation-checks)
        - [Taint analysis](#taint-analysis)
        - [Language Server](#language-server)
        - [Automatic fixes](#automatic-fixes)
        - [Automatic refactoring](#automatic-refactoring)
    - [IDE 지원](#ide-지원)
        - [PhpStorm](#phpstorm)
            - [전제 조건](#전제-조건)
            - [`composer.json` 통한 가이드](#composerjson-통한-가이드)
                - [Install and configure Psalm](#install-and-configure-psalm)
                - [Configure Psalm in PhpStorm](#configure-psalm-in-phpstorm)
                - [`Psalm`을 PhpStorm 검사로 활성화](#psalm을-phpstorm-검사로-활성화)
            - [`brew install psalm` 통해 설치한 경우](#brew-install-psalm-통해-설치한-경우)
            - [검사 설정에서 Psalm 유효성 검사 활성화](#검사-설정에서-psalm-유효성-검사-활성화)
            - [Use extended `@psalm` annotations](#use-extended-psalm-annotations)
    - [Psalm annotations](#psalm-annotations)
    - [stubs](#stubs)
        - [템플릿 리파지토리 사용하기](#템플릿-리파지토리-사용하기)
        - [Stub files?](#stub-files)
        - [generating stubs](#generating-stubs)
        - [Advanced topics](#advanced-topics)
            - [Starting from scratch](#starting-from-scratch)
    - [기타](#기타)

## psalm?

- psalm (n.):
    - 시편(詩篇, 영어: Book of Psalms)은 유대교에서 쓰는 타나크(기독교에서 쓰는 구약성경)의 일부
    - psallein "현악기 연주하다, 당기다, 끌다" (feel 참조)에서 유래
- 프로그램을 분석하고 타입 관련 버그를 가능한 한 많이 탐색하는 정적 분석기

## 기능

### Mixed type warnings

- If Psalm cannot infer a type for an expression then it uses a `mixed` placeholder type.
- `mixed` types can sometimes mask bugs, so keeping track of them helps you avoid a number of common pitfalls.

### Intelligent logic checks

- Psalm keeps track of logical assertions made about your code, both treated as issues.

    ```php
    if ($a && $a) {}
    if ($a && !$a) {}
    ```

- Psalm also keeps track of logical assertions made in prior code paths, preventing issues like:

    ```php
    if ($a) {

    } elseif ($a) {
        
    }
    ```

### Property initialisation checks

Psalm checks that all properties of a given object have values after the constructor is called.

### Taint analysis

Psalm can detect security vulnerabilities in your code.

### Language Server

Psalm has a Language Server that’s compatible with a range of different IDEs.

### Automatic fixes

Psalm can fix many of the issues it finds automatically.

### Automatic refactoring

Psalm can also perform simple refactors from the command line.

## [IDE 지원](https://psalm.dev/docs/running_psalm/language_server/)

### PhpStorm

- `2020.3`부터 [네이티브로 지원](https://www.jetbrains.com/help/phpstorm/using-psalm.html#prerequisites)하고 있다

#### 전제 조건

- 실행 가능한 PHP 엔진을 포함하는 디렉토리는 반드시 시스템 `path`에 추가돼야 한다
- 이를 통해 code quality tool 스크립트가 시스템 전체 PHP 엔진에 대한 호출을 실행할 수 있다

#### `composer.json` 통한 가이드

##### [Install and configure Psalm](https://www.jetbrains.com/help/phpstorm/using-psalm.html#installing-configuring-code-sniffer)

- `composer.json`에 `vimeo/psalm` 의존성 추가하고 설치하도록 가이드
- `Psalm`이 처음 구성된 후, `composer.json`의 추가적인 수정은 검사(inspection) 구성에 영향을 미치지 않는다. 최신 변경사항을 적용하려면, `Psalm` 구성을 재설정 해야 한다
    - `설정 | PHP | 품질 도구 | Psalm`으로 이동하여 `Configuration` 옆의 `...` 버튼 클릭
    - `Psalm path` 필드를 비운다
    - `composer.json` 에디터 패널 상단의 `Update`를 글릭하여 프로젝트의 컴포저 의존성을 업데이트 한다
    - PhpStorm은 `Psalm` 구성을 새로 수행하여 `composer.json` 변경 사항을 적용한다

##### [Configure Psalm in PhpStorm](https://www.jetbrains.com/help/phpstorm/using-psalm.html#installing-configuring-code-sniffer)

- Composer와 함께 `Psalm`을 설치하면 PhpStorm이 `vendor/bin` 폴더에서 `Psalm`의 실행 파일을 자동으로 감지하고 시스템 경로에 구성된 PHP 인터프리터가 이를 실행하도록 설정한다
- `설정 | PHP | 품질 도구 | Psalm`에서 다음과 같은 설정을 할 수 있다
    - 기본 PHP 인터프리터를 변경하거나,
    - 수동으로 다운로드하여 설치한 Psalm 실행 파일의 경로를 설정하거나,
    - PhpStorm에서 실행할 때 Psalm에 전달할 몇 가지 옵션을 추가
- `설정 | PHP | 품질 도구 | Psalm`
    - Configuration: 기본 PHP 인터프리터와 `Psalm` 실행 파일 경로를 수정할 수 있다
    - Show ignored files: `Psalm` 유효성 검사에서 파일 제외 가능
    - Options:
        - `Psalm`을 PhpStorm 검사로 실행할 때 옵션 추가 가능
        - [`Psalm` 레퍼런스 페이지](https://www.jetbrains.com/help/phpstorm/php-quality-tools.html#psalm-options)에 설명된 필드들 수정
            - `Show info`: 선택하면 Psalm에서 구성 파일에 지정된 errorLevel보다 낮은 수준의 오류를 보고하도록 한다. 이 확인란을 선택하지 않으면 해당 오류는 무시된다.

##### `Psalm`을 PhpStorm 검사로 활성화

- `composer.json`의 `scripts` 항목에 `Psalm` 구성에 대한 정보를 포함시킬 수 있다.

    ```json
    "scripts": {
        "psalm": "vendor/bin/psalm --config=psalm.xml"
    }
    ```

- 만약 `composer.json`의 `scripts` 항목에 구성 파일이 지정되지 않는다면, PhpStorm은 프로젝트 루트를 추가적으로 체크해서 `psalm.xml` 또는 `psalm.xml.dist`라는 기본 이름을 가진 규칙 집합(ruleset)을 찾는다
- Psalm을 처음 구성한 후에는 `composer.json`을 추가로 수정해도 검사 구성에 영향을 미치지 않는다. 최신 변경 사항을 적용하려면 `설정 | PHP | 품질 도구 | Psalm` 페이지에서 Psalm 구성을 재설정하고 프로젝트 종속성을 업데이트한다

#### `brew install psalm` 통해 설치한 경우

```shell
brew install psalm
```

- `composer.json` 변경사항 적용 불필요: 시스템 수준에서 설치된 경우, `composer.json`의 변경사항이 `Psalm` 설정에 직접적인 영향을 미치지 않는다. 따라서 `composer.json`을 수정하고 이를 Psalm 설정에 적용하는 과정이 필요하지 않는다.

#### 검사 설정에서 Psalm 유효성 검사 활성화

1. `설정 | Editor | Inspections`로 이동
2. `PHP|Quality Tools`를 펼쳐서 `Psalm validation` 체크박스 선택
3. 페이지의 우측 패널에서, PhpStorm이 Psalm 검사 출력을 어떻게 다룰 것인지 구성
    1. Scope
    2. Severity
    3. Highlighting in editor

#### Use extended `@psalm` annotations

- 일반 [PHPDoc 주석](https://www.jetbrains.com/help/phpstorm/phpdoc-comments.html) 외에도 PhpStorm은 코드 분석을 수행하기 위해 Psalm에서 사용하는 [Psalm 전용 주석](https://psalm.dev/docs/annotating_code/supported_annotations/)을 지원
- 주석을 지정할 때 대부분의 경우 `@psalm-` 부분을 생략할 수 있다
- `^Ctrl + Space` 로 자동 완성도 지원

## [Psalm annotations](https://www.jetbrains.com/help/phpstorm/php-type-checking.html#psalm-annotations)

## stubs

### 템플릿 리파지토리 사용하기

Github의 [플러그인 템플릿 리포지토리](https://github.com/weirdan/psalm-plugin-skeleton)로 이동하여 로그인한 후 `Use this template` 버튼을 클릭

### Stub files?

- Stub 파일은 Psalm의 확장된 문서 블록을 upstream 소스 파일에 직접 추가할 수 없는 경우, third-party type 정보를 재정의할 수 있는 방법을 제공
- 관례에 따라 스텁 파일의 확장자는 .phpstub으로 되어 있어 IDE가 실제 PHP 코드로 인식하지 못하도록 한다

### [generating stubs](https://psalm.dev/docs/running_psalm/plugins/authoring_plugins/#generating-stubs)

```shell
# if add dev require
composer require --dev cakephp/chronos

# and then generate stubs
vendor/bin/psalm --generate-stubs=stubs/chronos.phpstub
```

### Advanced topics

#### Starting from scratch

## 기타

- [MixedReturnTypeCoercion with functions returning arrays #6862](https://arc.net/l/quote/vpoeundp)
