# git submodules

- [git submodules](#git-submodules)
    - [git submodule](#git-submodule)
    - [init submodule](#init-submodule)
    - [`git submodule update`](#git-submodule-update)
        - [`--init --recursive`](#--init---recursive)

## git submodule

Git 서브모듈은 독립적인 Git 저장소를 다른 저장소의 서브디렉토리로 포함시킬 수 있는 기능입니다.

## init submodule

```shell
git submodule update --init --recursive
```

## `git submodule update`

서브모듈의 내용을 최신 상태로 업데이트합니다.
이 명령어는 상위 저장소에 기록된 서브모듈 커밋으로 체크아웃합니다.

```man
update [--init] [--remote] [-N|--no-fetch] [--[no-]recommend-shallow] [-f|--force] [--checkout|--rebase|--merge] [--reference <repository>] [--depth <depth>] [--recursive]
[--jobs <n>] [--[no-]single-branch] [--filter <filter spec>] [--] [<path>...]
    Update the registered submodules to match what the superproject expects by cloning missing submodules, fetching missing commits in submodules and updating the working tree
    of the submodules. The "updating" can be done in several ways depending on command line options and the value of submodule.<name>.update configuration variable. The command
    line option takes precedence over the configuration variable. If neither is given, a checkout is performed. The update procedures supported both from the command line as
    well as through the submodule.<name>.update configuration are:

    checkout
        the commit recorded in the superproject will be checked out in the submodule on a detached HEAD.

        If --force is specified, the submodule will be checked out (using git checkout --force), even if the commit specified in the index of the containing repository already
        matches the commit checked out in the submodule.

    rebase
        the current branch of the submodule will be rebased onto the commit recorded in the superproject.

    merge
        the commit recorded in the superproject will be merged into the current branch in the submodule.

    The following update procedures are only available via the submodule.<name>.update configuration variable:

    custom command
        arbitrary shell command that takes a single argument (the sha1 of the commit recorded in the superproject) is executed. When submodule.<name>.update is set to !command,
        the remainder after the exclamation mark is the custom command.

    none
        the submodule is not updated.

    If the submodule is not yet initialized, and you just want to use the setting as stored in .gitmodules, you can automatically initialize the submodule with the --init
    option.

    If --recursive is specified, this command will recurse into the registered submodules, and update any nested submodules within.

    If --filter <filter spec> is specified, the given partial clone filter will be applied to the submodule. See git-rev-list(1) for details on filter specifications.
```

```shell
# This works for any of the supported update procedures (--checkout, --rebase, etc.)
# The only change is the source of the target SHA-1
# For example
# `submodule update --remote --merge` will merge upstream submodule changes into the submodules
# while `submodule update --merge` will merge superproject gitlink changes into the submodules
# 서브모듈 업데이트

git submodule update --remote --merge
```

- `--init`:

    서브모듈을 처음 클론한 후 초기화하지 않으면 서브모듈의 내용을 가져오지 않습니다.
    `--init`을 사용하여 `.gitmodules` 파일에 정의된 서브모듈의 URL을 읽어 서브모듈 디렉토리에 Git 저장소를 초기화합니다.

    ```bash
    git submodule update --init
    ```

- `--recursive`:

    서브모듈 내의 서브모듈도 포함하여 업데이트합니다.
    서브모듈의 서브모듈이 있는 경우에도 이를 모두 초기화하고 업데이트합니다

### `--init --recursive`

이 명령어는 서브모듈을 초기화하고 업데이트하며, 재귀적으로 모든 하위 서브모듈까지 포함하여 처리합니다.
`.gitmodules` 파일에 정의된 URL을 읽어 서브모듈을 초기화하고, 상위 저장소에 기록된 서브모듈 커밋으로 서브모듈의 상태를 동기화합니다.
서브모듈의 서브모듈까지 모두 처리하여 전체 저장소의 일관성을 유지합니다.

즉, 서브모듈과 관련된 모든 저장소를 한 번에 설정하고 최신 상태로 유지할 수 있습니다.

1. 초기에는 서브모듈 디렉토리가 비어 있습니다.

    ```bash
    main-repo/           (상위 저장소)
    ├── .git/            (Git 메타데이터)
    ├── .gitmodules      (서브모듈 정보 파일)
    ├── submodule1/      (서브모듈 디렉토리, 비어 있음)
    └── submodule2/      (서브모듈 디렉토리, 비어 있음)
    ```

2. `.gitmodules` 파일에는 서브모듈들에 다른 저장소에 대한 정보가 저장되어 있습니다.

    ```plaintext
    [submodule "submodule1"]
        path = submodule1
        url = https://github.com/user/submodule1.git
    [submodule "submodule2"]
        path = submodule2
        url = https://github.com/user/submodule2.git
    ```

3. `git submodule update --init --recursive` 실행하면 서브모듈을 초기화하고 업데이트 합니다.

    1. 서브모듈 초기화

        서브모듈 디렉토리 (`submodule1` 및 `submodule2`)에 Git 저장소를 초기화합니다.
        `.gitmodules` 파일에서 서브모듈의 URL을 읽어 해당 리포지토리를 클론합니다.

    2. 서브모듈 업데이트

        서브모듈 디렉토리에 해당 커밋을 체크아웃합니다.
        상위 저장소에 기록된 서브모듈 커밋으로 서브모듈의 상태를 동기화합니다.

    3. 재귀적으로 업데이트

        서브모듈 내의 서브모듈까지 포함하여 초기화하고 업데이트합니다.

4. 실행 후 디렉토리 구조는 다음과 같습니다.

    ```bash
    main-repo/
    ├── .git/
    ├── .gitmodules
    ├── submodule1/
    │   ├── .git/
    │   ├── (submodule1의 파일들)
    │   └── submoduleA/   (submodule1 내의 서브모듈)
    │       ├── .git/
    │       └── (submoduleA의 파일들)
    └── submodule2/
        ├── .git/
        └── (submodule2의 파일들)
    ```
