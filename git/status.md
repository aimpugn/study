# git status

- [git status](#git-status)
    - [git status?](#git-status-1)
    - [출력 섹션](#출력-섹션)
        - [4. Unmerged paths](#4-unmerged-paths)
        - [5. Ignored files](#5-ignored-files)
    - [`-uall`](#-uall)

## git status?

`git status` 명령은 작업 디렉토리와 인덱스(스테이징 영역)의 상태를 보여주는 중요한 도구입니다.

## 출력 섹션

- 이미 스테이징 영역에 추가된 변경사항을 나타냅니다. 즉, 다음 커밋에 포함될 준비가 된 파일들의 목록입니다. 이 파일들은 `git commit` 명령을 실행하면 커밋됩니다.

    ```bash
    Changes to be committed:
        (use "git restore --staged <file>..." to unstage)
        modified:   example.txt
    ```

- 수정되었지만 아직 스테이징 영역에 추가되지 않은 파일들을 나타냅니다. 이 파일들은 `git add <file>`을 사용하여 스테이징 영역에 추가할 수 있습니다.

    ```bash
    Changes not staged for commit:
        (use "git add <file>..." to update what will be committed)
        (use "git restore <file>..." to discard changes in working directory)
            modified:   example.txt
    ```

- 추적이 시작되지 않은 파일들을 나타냅니다. 이 파일들은 저장소에 새로 추가된 파일들로, `git add <file>`을 통해 추적을 시작할 수 있습니다.

    ```bash
    Untracked files:
        (use "git add <file>..." to include in what will be committed)
            newfile.txt
    ```

### 4. Unmerged paths

이 섹션은 충돌이 발생한 파일들을 나타냅니다.
이는 병합(merge) 또는 리베이스(rebase) 중에 발생할 수 있으며, 사용자가 수동으로 충돌을 해결해야 합니다.

**예시:**

```bash
Unmerged paths:
  (use "git add/rm <file>..." as appropriate to mark resolution)
    deleted by us:   conflictfile.txt
```

### 5. Ignored files

이 섹션은 `.gitignore` 파일에 의해 무시된 파일들을 나타낼 수 있습니다.
`git status` 명령에서 기본적으로는 무시된 파일들은 나타나지 않지만, `git status --ignored` 옵션을 사용하면 볼 수 있습니다.

**예시:**

```bash
Ignored files:
  (use "git add -f <file>..." to include in what will be committed)
    ignoredfile.txt
```

## `-uall`

```bash
-u[<mode>], --untracked-files[=<mode>]
```

`mode` 파라미터는 추적되지 않은 파일을 처리하는 방법을 지정하는 데 사용됩니다.
이 파라미터는 선택 사항이며, 기본값은 `all`입니다.
지정된 경우 옵션에 붙여서 사용해야 합니다(예: `-uno`, ~~`-u no`~~).

가능한 옵션은 다음과 같습니다:

- `no`: 추적되지 않은 파일을 표시하지 않습니다.
- `normal`: 추적되지 않은 파일과 디렉토리를 표시합니다.
- `all`: 추적되지 않은 디렉토리 내의 개별 파일도 표시합니다.

`-u` 옵션이 사용되지 않으면, 추적되지 않은 파일과 디렉토리가 표시됩니다(즉, `normal`을 지정한 것과 동일).
이는 새로 생성된 파일을 추가하는 것을 잊지 않도록 돕기 위함입니다.
파일 시스템에서 추적되지 않은 파일을 찾는 데 추가 작업이 필요하기 때문에, 이 모드는 큰 작업 트리에서 시간이 걸릴 수 있습니다.

지원되는 경우 추적되지 않은 캐시와 분할 인덱스를 활성화하는 것을 고려하십시오
- `git update-index --untracked-cache`
- `git update-index --split-index` 참조

그렇지 않으면, 추적되지 않은 파일을 표시하지 않고 `git status`가 더 빠르게 반환되도록 `no`를 사용할 수 있습니다.

기본값은 `git-config(1)`에 문서화된 `status.showUntrackedFiles` 구성 변수로 변경할 수 있습니다.
