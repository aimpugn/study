# Merge

- [Merge](#merge)
    - [Merge?](#merge-1)
    - [`git merge` 과정 예시](#git-merge-과정-예시)
    - [merge commit](#merge-commit)
        - [`--merges` 옵션](#--merges-옵션)
        - [머지 커밋의 특징](#머지-커밋의-특징)
        - [`git log` 명령어로 머지 커밋 확인하기](#git-log-명령어로-머지-커밋-확인하기)
    - [git merge-base](#git-merge-base)
    - [Options](#options)
        - [`--ff`, `--no-ff`, `--ff-only`](#--ff---no-ff---ff-only)
        - [`--min-parents=<number>`, `--max-parents=<number>`, `--no-min-parents`, `--no-max-parents`](#--min-parentsnumber---max-parentsnumber---no-min-parents---no-max-parents)

## Merge?

병합은 두 개의 브랜치를 합치는 작업입니다.
병합을 수행하면, 두 브랜치의 변경 사항을 모두 포함하는 새로운 커밋이 생성됩니다.

일반적인 병합에서 Git은 두 브랜치의 공통 조상을 찾아 3-way merge를 수행합니다.
그러나 리베이스와 강제 푸시를 통해 히스토리를 재작성하면, 이전의 복잡한 병합 기록이 제거되고 선형적인 히스토리가 만들어집니다.

예제:

1. `main` 브랜치와 `feature` 브랜치가 있다
2. `feature` 브랜치에서 새로운 기능을 개발하고, `main` 브랜치는 다른 사람들의 작업으로 인해 업데이트되었다
3. `feature` 브랜치에서 `main` 브랜치를 병합하면, 두 브랜치의 변경 사항을 모두 포함하는 새로운 커밋이 생성된다

## `git merge` 과정 예시

1. Make a new branch called `bugFix`
2. Checkout the `bugFix` branch with git checkout bugFix

    ```shell
    git checkout -b bugFix
    ```

3. Commit once

    ```shell
    git commit
    ```

4. Go back to `main` with `git checkout`

    ```shell
    git checkout main
    ```

5. Commit another time

    ```shell
    git commit
    ```

6. Merge the branch `bugFix` into `main` with `git merge`

    ```shell
    git merge bugFix
    ```

## merge commit

머지 커밋은 두 개 이상의 부모 커밋을 가지는 커밋을 말한다.
일반적으로, 두 개의 브랜치를 합칠 때 생성되며, 이 과정에서 발생하는 변경 사항을 포함한다.
머지 커밋은 브랜치의 히스토리를 유지하면서, 두 브랜치의 변경 사항을 하나의 커밋으로 결합한다.

### `--merges` 옵션

> `--merges`
>
> Print only merge commits.
> This is exactly the same as `--min-parents=2`

### 머지 커밋의 특징

- 머지 커밋은 두 개의 부모 커밋을 가리키는데, 이는 머지되는 두 브랜치의 마지막 커밋을 의미한다.
- 머지 커밋은 두 브랜치의 변경 사항을 하나의 커밋으로 통합하여, 프로젝트의 히스토리에 추가합니다.
- **충돌 해결**: 두 브랜치를 머지할 때 충돌이 발생하면, 이를 해결하고 머지 커밋을 생성하여 변경 사항을 반영합니다.

### `git log` 명령어로 머지 커밋 확인하기

`git log` 명령어는 커밋 히스토리를 보여주며, `--merges` 옵션을 사용하면 머지 커밋만을 필터링하여 보여줍니다.

```bash
git log --merges
```

이 명령어는 현재 브랜치의 머지 커밋 히스토리를 보여줍니다. 특정 브랜치의 머지 커밋을 확인하고 싶다면, 브랜치 이름을 명시할 수 있습니다.

```bash
git log --merges <branch-name>
```

또한, 두 브랜치 사이의 머지 커밋을 확인하고 싶다면, 다음과 같이 두 브랜치 사이의 범위를 지정할 수 있습니다.

```bash
git log --merges <branch-name-1>..<branch-name-2>
```

예를 들어, `origin/main`과 `origin/release/240404` 사이의 머지 커밋을 확인하려면 다음 명령어를 사용합니다.

```bash
git log --merges origin/main..origin/release/240404
```

이 명령어는 `origin/main`과 `origin/release/240404` 사이에 있는 머지 커밋만을 보여줍니다. 결과가 없다면, 지정된 두 브랜치 사이에 머지 커밋이 없다는 것을 의미합니다.

머지 커밋을 확인하는 이 과정은 브랜치 간의 변경 사항을 이해하고, 프로젝트의 히스토리를 추적하는 데 유용합니다.

## git merge-base

- The merge base is the most recent common ancestor of two branches in the commit history.

```shell
git merge-base <branch1> <branch2>
```

## Options

### `--ff`, `--no-ff`, `--ff-only`

Specifies how a merge is handled when the merged-in history is already a descendant
of the current history.  `--ff` is the default unless merging an annotated (and
possibly signed) tag that is not stored in its natural place in the refs/tags/
hierarchy, in which case -`-no-ff` is assumed.

With `--ff`, when possible resolve the merge as a fast-forward (only update the branch
pointer to match the merged branch; do not create a merge commit). When not possible
(when the merged-in history is not a descendant of the current history), create a
merge commit.

With `--no-ff`, create a merge commit in all cases, even when the merge could instead
be resolved as a fast-forward.

With `--ff-only`, resolve the merge as a fast-forward when possible. When not
possible, refuse to merge and exit with a non-zero status.

### `--min-parents=<number>`, `--max-parents=<number>`, `--no-min-parents`, `--no-max-parents`

Show only commits which have at least (or at most) that many parent commits.

In particular,

- `--max-parents=1` is the same as `--no-merges`
- `--min-parents=2` is the same as `--merges`
- `--max-parents=0` gives all root commits
- `--min-parents=3` all octopus merges.

`--no-min-parents` and `--no-max-parents` reset these limits (to no limit) again.

Equivalent forms are `--min-parents=0` (any commit has 0 or more parents) and `--max-parents=-1` (negative numbers denote no upper limit).
