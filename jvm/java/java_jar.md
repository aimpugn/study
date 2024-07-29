# java jar

## Synopsis

jar는 클래스 및 리소스에 대한 아카이브를 생성합니다.
아카이브에서 개별 클래스나 리소스를 조작하거나 복원할 수 있습니다.

```sh
jar [OPTION...] [ [--release VERSION] [-C dir] 파일] ...
```

## 예제

- 두 클래스 파일을 사용하여 classes.jar라는 아카이브 생성

    ```sh
    jar --create --file classes.jar Foo.class Bar.class
    ```

- 기존 Manifest를 사용하여 모든 파일이 foo/에 포함된 아카이브 생성

    ```sh
    jar --create --file classes.jar \
        --manifest mymanifest \
        -C foo/ .
    ```

- 모듈 기술자가 classes/module-info.class에 위치한 모듈형 jar 아카이브 생성

    ```sh
    jar --create --file foo.jar \
        --main-class com.foo.Main \
        --module-version 1.0 \
        -C foo/ classes resources
    ```

- 기존 비모듈 jar를 모듈형 jar로 업데이트

    ```sh
    jar --update --file foo.jar \
        --main-class com.foo.Main \
        --module-version 1.0 \
        -C foo/ module-info.class
    ```

- 일부 파일이 `META-INF/versions/9` 디렉토리에 위치한 다중 릴리스 jar 생성

    ```sh
    jar --create --file mr.jar \
        -C foo classes \
        --release 9 \
        -C foo9 classes
    ```

jar 명령을 짧게 만들거나 단순화하려면 별도의 텍스트 파일에 인수를 지정하고,
at 기호(`@`)를 접두어로 사용하여 jar 명령에 전달할 수 있습니다.

```sh
# classes.list 파일에서 추가 옵션 및 클래스 파일 목록 읽기
jar --create --file my.jar @classes.list
```

## 옵션

> 아카이브는 모듈 기술자 'module-info.class'가 주어진 디렉토리의 루트 또는 jar 아카이브 자체의 루트에 위치한 경우 모듈형 jar입니다.

JAR 파일이 '모듈형'으로 간주되려면 '`module-info.class`' 파일이 특정 위치에 있어야 합니다.
이 '`module-info.class`' 파일은 다음 두 위치 중 하나에 있어야 합니다:
- JAR 파일을 만들 때 사용하는 디렉토리의 최상위 레벨
- 이미 만들어진 JAR 파일 내부의 최상위 레벨

### 기본 작업 모드

- `-c`,`--create`: 아카이브를 생성합니다.

- `-i`, `--generate-index=FILE`: 지정된 jar 아카이브에 대한 인덱스 정보를 생성합니다.

- `-t`, `--list`: 아카이브에 대한 목차를 나열합니다.

- `-u`, `--update`: 기존 jar 아카이브를 업데이트합니다.

- `-x`, `--extract`: 명명된(또는 모든) 파일을 아카이브에서 추출합니다.

- `-d`, `--describe-module`: 모듈 기술자 또는 자동 모듈 이름을 인쇄합니다.

- `--validate`:

    jar 아카이브의 콘텐츠 유효성을 검사합니다.
    이 옵션은 다중 릴리스 jar 아카이브에서 내보낸 API가 모든 다른 릴리스 버전에서 일관성이 있는지 확인합니다.

### 모든 모드에서 적합한 작업 수정자

- `-C DIR`: 지정된 디렉토리로 변경하고 다음 파일을 포함합니다.

- `-f`, `--file=FILE`: 아카이브 파일 이름입니다. 생략할 경우 작업에 따라 stdin 또는 stdout이 사용됩니다.

- `--release VERSION`: 다음 모든 파일을 버전 지정된 jar 디렉토리 (예: `META-INF/versions/VERSION/`)에 배치합니다.

- `-v`, `--verbose`: 표준 출력에 상세 정보 출력을 생성합니다.

### 생성 및 업데이트 모드에서만 적합한 작업 수정자

- `-e`, `--main-class=CLASSNAME`:

    모듈형 또는 실행형 jar 아카이브에 번들로 제공된 독립형 애플리케이션의 애플리케이션 시작 지점입니다.

- `-m`, `--manifest=FILE`: 지정된 Manifest 파일의 Manifest 정보를 포함합니다.

- `-M`, `--no-manifest`: 항목에 대해 Manifest 파일을 생성하지 않습니다.

- `--module-version=VERSION`; 모듈형 jar를 생성하거나 비모듈 jar를 업데이트할 때 모듈 버전입니다.

- `--hash-modules=PATTERN`:

    생성 중인 모듈형 jar 또는 업데이트 중인 비모듈 jar에 직접 또는 간접적으로 의존하고 주어진 패턴과 일치하는
    모듈의 해시를 컴퓨트하고 기록합니다.

- `-p`, `--module-path`: 해시를 생성하기 위한 모듈 종속성의 위치입니다.

### 생성, 업데이트 및 generate-index 모드에서만 적합한 작업 수정자

- `-0`, `--no-compress`: 저장 전용이며 ZIP 압축을 사용하지 않습니다.

- `--date=TIMESTAMP`:

    ISO-8601의 타임스탬프는 항목의 타임스탬프에 사용할 수 있는 선택적 시간대 형식의 확장 오프셋 날짜-시간(예: "2022-02-12T12:30:00-05:00")입니다.

아카이브는 모듈 기술자 'module-info.class'가 주어진 디렉토리의 루트 또는 jar 아카이브 자체의 루트에 위치한 경우 모듈형 jar입니다.
다음 작업은 모듈형 jar를 생성하거나 기존 비모듈 jar를 업데이트할 때만 적합합니다. '--module-version', '--hash-modules' 및 '--module-path'.

long 옵션의 필수 또는 선택적 인수는 해당하는 short 옵션에 대해서도 필수 또는 선택적입니다.
