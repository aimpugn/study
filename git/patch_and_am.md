# patch and am

- [patch and am](#patch-and-am)
    - [git patch?](#git-patch)
    - [`git format-patch`](#git-format-patch)
        - [특정 커밋의 패치 파일 생성](#특정-커밋의-패치-파일-생성)
        - [커밋 범위의 패치 파일 생성](#커밋-범위의-패치-파일-생성)
        - [패치 파일 적용](#패치-파일-적용)
    - [`git am`](#git-am)
        - [패치 파일 적용](#패치-파일-적용-1)
        - [여러 패치 파일 적용](#여러-패치-파일-적용)
        - [3-way 머지 사용](#3-way-머지-사용)
        - [충돌 처리](#충돌-처리)
    - [`format-patch`와 `am` 사용례](#format-patch와-am-사용례)
        - [IntelliJ Git Tool에서 Create Patch 기능에 사용되는 Git CLI 명령어](#intellij-git-tool에서-create-patch-기능에-사용되는-git-cli-명령어)
            - ["Base path doesn't contain all selected changes" 오류](#base-path-doesnt-contain-all-selected-changes-오류)

## git patch?

`git patch`는 Git 명령어가 아니지만, 일반적으로 패치 파일(patch file)을 다루는 데 사용됩니다.
Git에서 패치를 생성하거나 적용할 때 사용하는 방법을 설명합니다.

패치 파일은 코드 변경 사항을 기록한 파일로, 이를 통해 변경 사항을 다른 사람과 공유하거나 적용할 수 있습니다. 패치 파일은 주로 `diff` 형식으로 저장됩니다.

## `git format-patch`

`git format-patch`는 특정 커밋 또는 커밋 범위의 변경 사항을 패치 파일로 생성하는 명령어입니다.
이 패치 파일은 나중에 `git apply` 명령어로 적용할 수 있습니다.

### 특정 커밋의 패치 파일 생성

```bash
# git format-patch -1 <commit-hash>
git format-patch -1 abc123
```

- `-1`:  하나의 커밋에 대해 패치 파일을 생성하라는 의미입니다.

`abc123` 커밋의 변경 사항을 포함하는 패치 파일을 생성합니다.
생성된 파일은 `0001-<commit-message>.patch` 형태로 저장됩니다.

### 커밋 범위의 패치 파일 생성

```bash
# git format-patch <start-commit>..<end-commit>
git format-patch abc123..def456
```

- `<start-commit>`에서 `<end-commit>`까지의 모든 커밋에 대해 패치 파일을 생성합니다.

`abc123` 이후부터 `def456`까지의 모든 커밋에 대한 패치 파일을 생성합니다.
각 커밋마다 별도의 패치 파일이 생성됩니다.

### 패치 파일 적용

- `git apply`

    생성된 패치 파일을 적용하려면 `git apply` 명령어를 사용합니다.

    ```bash
    git apply 0001-some-change.patch
    ```

    이 명령어는 `0001-some-change.patch` 파일에 기록된 변경 사항을 현재 작업 디렉터리에 적용합니다.

- `git am`

    `git am` 명령어는 `git apply`와 달리 패치 파일을 적용하고, 해당 패치의 커밋 메시지를 사용하여 새로운 커밋을 생성합니다.

    ```bash
    git am < 0001-some-change.patch
    ```

    이 명령어는 `0001-some-change.patch` 파일에 기록된 변경 사항을 적용하고, 새로운 커밋을 생성합니다.

## `git am`

> git-am - Apply a series of patches from a mailbox

`git am`은 Git에서 패치 파일을 적용하고, 해당 패치의 커밋 메시지와 메타데이터를 사용하여 새로운 커밋을 생성하는 명령어입니다.
이 명령어는 주로 이메일로 전송된 패치를 적용하거나, 다른 개발자가 제공한 패치 파일을 프로젝트에 반영할 때 사용됩니다.

`git am`을 사용하는 이유는 다음과 같습니다:

- **커밋 메타데이터 보존**

    `git am`은 패치 파일의 내용을 현재 브랜치에 적용합니다.
    패치 파일에 포함된 원본 커밋의 작성자, 날짜, 메시지를 그대로 보존하여 새로운 커밋을 생성합니다.

- **편리한 패치 관리**

    이메일로 전달받은 패치를 쉽게 적용하고 관리할 수 있습니다.

- **자동화된 패치 적용**

    디렉토리 내 여러 패치 파일을 순차적으로 적용할 수 있습니다.
    여러 패치 파일을 일괄 적용하여 수작업을 줄일 수 있습니다.

- **충돌 처리**

    패치 적용 중 충돌이 발생하면 이를 해결할 수 있는 도구와 명령을 제공합니다.

### 패치 파일 적용

이 명령어는 패치 파일을 현재 브랜치에 적용하고, 새로운 커밋을 생성합니다.

```bash
git am < patch-file.patch
```

### 여러 패치 파일 적용

이 명령어는 현재 작업 중인 디렉토리에서 모든 `.patch` 파일을 찾아서 순차적으로 적용합니다.
여기서 `<`는 필요 없습니다.

예를 들어, 현재 디렉토리에 `0001-fix-bug.patch`, `0002-add-feature.patch` 등의 파일이 있다면, 이 명령어를 실행하면 이 파일들이 순차적으로 적용됩니다.

```bash
git am *.patch
```

특정 경로에 있는 모든 패치 파일을 적용한는 것도 가능합니다.
`/path/to/patches/` 디렉토리에 있는 모든 `.patch` 파일을 찾아서 순차적으로 적용합니다.
여기서도 `<`는 필요 없습니다.

예를 들어, `/path/to/patches/` 디렉토리에 `0001-fix-bug.patch`, `0002-add-feature.patch` 등의 파일이 있다면, 이 명령어를 실행하면 이 파일들이 순차적으로 적용됩니다.

```bash
git am /path/to/patches/*.patch
```

### 3-way 머지 사용

3-way 머지 전략을 사용하여 패치를 적용할 때 충돌을 보다 쉽게 해결할 수 있습니다.

```bash
git am -3 < patch-file
```

### 충돌 처리

패치를 적용하는 중에 충돌이 발생하면 Git은 충돌을 해결할 때까지 패치 적용을 중단합니다.
다음 단계는 충돌을 해결하는 방법입니다.

- 충돌 해결 후 계속 진행

    충돌을 해결하고 `git add` 명령어로 변경 사항을 스테이징한 후, `git am --continue` 명령어로 패치 적용을 계속 진행합니다.

    ```bash
    git add <conflicted-file>
    git am --continue
    ```

- 충돌 해결 없이 패치 적용을 중단하고, 이전 상태로 되돌립니다.

    ```bash
    git am --abort
    ```

## `format-patch`와 `am` 사용례

- 최신 커밋을 패치 파일로 생성합니다.

    생성된 파일은 `0001-<commit-message>.patch`와 같은 형식으로 저장됩니다.

    ```bash
    # 패치 파일 생성 (`git format-patch`)
    git format-patch -1 HEAD
    ```

- 패치 파일 적용 (`git am`)

    이 명령어는 패치 파일을 현재 브랜치에 적용하고, 패치 파일의 커밋 메시지와 작성자 정보를 사용하여 새로운 커밋을 생성합니다.

    ```bash
    git am < 0001-<commit-message>.patch
    ```

## 기타

### IntelliJ Git Tool에서 Create Patch 기능에 사용되는 Git CLI 명령어

IntelliJ의 Create Patch 기능은 기본적으로 `git diff` 명령어를 사용하여 패치를 생성합니다.
이 과정에서 Git CLI 명령어로 변환하면 다음과 같은 명령어가 사용됩니다.

```bash
git diff <commit> > <patch-file>
```

여기서 `<commit>`은 비교할 기준이 되는 커밋입니다 (예: HEAD, 특정 커밋 해시). `<patch-file>`은 생성될 패치 파일의 경로입니다.

Git CLI 명령어 자체에는 "base path" 설정 옵션이 없습니다.
하지만, 패치 파일의 경로를 상대 경로로 만들기 위해서는 다음과 같은 방법을 사용할 수 있습니다:

```bash
# 프로젝트 루트 디렉토리에서 실행
cd /home/user/project
git diff HEAD > my-changes.patch
```

이렇게 하면 패치 파일에 기록된 경로는 프로젝트 루트 디렉토리를 기준으로 상대 경로로 저장됩니다.

#### "Base path doesn't contain all selected changes" 오류

이 오류 메시지는 선택한 변경 사항이 지정된 "base path"에 포함되지 않는 경우 발생합니다.
기본적으로, 패치 파일을 생성할 때 IntelliJ는 선택한 파일의 변경 사항을 포함시키고자 합니다.
"base path"는 이 변경 사항이 기준으로 삼을 디렉토리를 지정합니다.

이 오류는 다음과 같은 이유로 발생할 수 있습니다:

1. **잘못된 Base Path**:
   - 선택한 파일의 변경 사항이 지정한 "base path" 디렉토리 하위에 있지 않을 때 발생합니다.
   - 예를 들어, 프로젝트 루트 디렉토리에서 선택한 파일이 `src/main/java/App.java`인 경우, "base path"를 `src/main`으로 설정하면 이 오류가 발생합니다. `src/main/java`는 `src/main` 디렉토리 하위에 없기 때문입니다.

2. **불일치하는 디렉토리 구조**:
   - 선택한 파일의 경로가 설정된 "base path"와 일치하지 않는 경우입니다.
   - 패치를 생성할 때 모든 파일이 설정된 "base path" 하위에 있어야 합니다.

- 프로젝트 구조:

  ```bash
  /home/user/project
  |-- src
  |   |-- main
  |       |-- java
  |           |-- App.java
  |-- test
      |-- main
          |-- java
              |-- AppTest.java
  ```

- 선택한 변경 사항: `src/main/java/App.java`
- **올바른 Base Path**: `/home/user/project`
- **잘못된 Base Path**: `/home/user/project/src`

위의 예시에서 "Base path"를 `/home/user/project/src`로 설정하면 오류가 발생합니다.
이는 `App.java` 파일이 `/home/user/project/src/main/java` 경로에 있기 때문입니다.
