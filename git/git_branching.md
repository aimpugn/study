# git branch

- [git branch](#git-branch)
    - [basic](#basic)
    - [현재 브랜치명 보기](#현재-브랜치명-보기)
    - [원격 브랜치 목록 보기](#원격-브랜치-목록-보기)
    - [특정 브랜치 클론](#특정-브랜치-클론)
    - [원격 브랜치 이름 변경 후 업데이트](#원격-브랜치-이름-변경-후-업데이트)
    - [로컬 브랜치 생성 후 remote 브랜치 생성](#로컬-브랜치-생성-후-remote-브랜치-생성)
    - [upstream](#upstream)
        - [show current upstream](#show-current-upstream)
        - [unset upstream](#unset-upstream)
        - [set upstream](#set-upstream)
        - [git --track vs --set-upstream vs --set-upstream-to](#git---track-vs---set-upstream-vs---set-upstream-to)
    - [diverge 상태](#diverge-상태)
        - [diverge란?](#diverge란)
        - [원인](#원인)
        - [해결 방법](#해결-방법)
        - [level rampup1 설명](#level-rampup1-설명)
        - [level rampup1 문제](#level-rampup1-문제)
    - [level rampup2](#level-rampup2)
        - [level rampup3 설명](#level-rampup3-설명)
        - [level rampup3 문제](#level-rampup3-문제)
    - [level rampup3](#level-rampup3)
        - [level rampup3: 설명](#level-rampup3-설명-1)
        - [level rampup3: 문제](#level-rampup3-문제-1)
    - [level rampup4](#level-rampup4)
        - [level rampup4: 설명](#level-rampup4-설명)
        - [level rampup4: 문제](#level-rampup4-문제)
    - [level move1](#level-move1)
        - [level move1: 설명](#level-move1-설명)
        - [level move1: 문제](#level-move1-문제)
    - [level move2](#level-move2)
        - [level move2: 설명](#level-move2-설명)
        - [level move2: 문제](#level-move2-문제)
    - [level mixed1](#level-mixed1)
        - [level mixed1: 설명](#level-mixed1-설명)
        - [level mixed1: 문제](#level-mixed1-문제)

## basic

```shell
#!/bin/bash
# 현재 브랜치명 보기
# 1.
git branch --show-current
# 2. https://stackoverflow.com/a/47095729
git rev-parse --abbrev-ref HEAD

# 특정 브랜치 클론
git clone -b "branch/name_to_clone" --single-branch "${URL_OF_REPOSITORY}" /path/to/project/directory

# 특정 브랜치 클론 후 다른 remote 브랜치 가져오기
git fetch origin main
# remote 가져온 후 checkout
git checkout FETCH_HEAD -b main
```

## 현재 브랜치명 보기

```shell
git branch --show-current
```

```shell
# https://stackoverflow.com/a/47095729 
git rev-parse --abbrev-ref HEAD
```

## 원격 브랜치 목록 보기

```bash
# `--remotes`
git branch -r
```

```bash
git ls-remote --heads
```

## 특정 브랜치 클론

```shell
git clone -b "branch/name_to_clone" --single-branch "${URL_OF_REPOSITORY}" /path/to/project/directory
```

```shell
# 특정 브랜치 클론 후 다른 remote 브랜치 가져오기
git fetch origin main
```

```shell
# remote 가져온 후 checkout
git checkout FETCH_HEAD -b main
```

## 원격 브랜치 이름 변경 후 업데이트

```shell
git branch -m <OLD_NAME> <NEW_NAME>
git fetch origin
# -u <upstream>, --set-upstream-to=<upstream>
git branch -u origin/<NEW_NAME> <NEW_NAME>
git remote set-head origin -a
```

## 로컬 브랜치 생성 후 remote 브랜치 생성

```bash
# 새로운 브랜치 생성
git checkout -b new-branch

# 리모트 저장소에 새 브랜치 푸시
# `-u`: upstream을 의미. 로컬 브랜치와 리모트 브랜치 간의 연결을 설정
git push -u origin new-branch
```

## [upstream](../terminology/upstream.md)

### show current upstream

```bash
# https://stackoverflow.com/a/9753364
git rev-parse --abbrev-ref --symbolic-full-name @{u}
# origin/coreapi/feature/application-payments-layered-architecture

git rev-parse --symbolic-full-name @{u}
# refs/remotes/origin/coreapi/feature/application-payments-layered-architecture
```

- `git rev-parse`

    Git에서 다양한 유형의 객체(예: 커밋, 태그, 브랜치 등)를 식별하는 데 사용됩니다.
    이 명령은 주어진 이름을 Git이 이해할 수 있는 형식으로 변환합니다.

- `--abbrev-ref`

    이 옵션은 주어진 레퍼런스(참조)의 가장 짧은 비-앰비규어스(ambiguous, 모호하지 않은) 형태를 출력합니다.

    예를 들어, 전체 브랜치 이름이 `refs/heads/master`일 경우, `--abbrev-ref` 옵션은 단순히 `master`로 출력합니다.

- `--symbolic-full-name`

    이 옵션은 주어진 레퍼런스의 전체 심볼릭 이름을 출력합니다.

    예를 들어, `master` 브랜치에 대해 이 옵션을 사용하면 `refs/heads/master`와 같이 전체 경로를 포함한 이름을 반환합니다. 이는 레퍼런스의 정확한 위치를 파악하고자 할 때 유용합니다.

- `@{u}` 또는 `@{upstream}`

    이 특수한 참조는 현재 체크아웃된 브랜치의 "upstream" 브랜치를 가리킵니다.
    "upstream" 브랜치는 보통 원격 저장소의 브랜치를 의미하며, 로컬 브랜치와 연관된 브랜치입니다.

    예를 들어, 로컬의 `master` 브랜치가 `origin/master`를 추적하고 있다면, `@{u}`는 `origin/master`를 가리킵니다.

```bash
git for-each-ref --format='%(upstream:short)' "$(git symbolic-ref -q HEAD)"
```

또는 [`--vv` 옵션](https://git-scm.com/docs/git-branch#Documentation/git-branch.txt--vv) 사용

```shell
git branch -vv
```

### [unset upstream](https://git-scm.com/docs/git-branch#Documentation/git-branch.txt---unset-upstream)

```shell
git branch --unset-upstream {<branch-name>|current-branch}
```

### set upstream

- ~~`--set-upstream`~~ 사용하지 말 것
- 대신 `--set-upstream-to` 또는 `--track` 사용

```shell
git branch --set-upstream-to=upstream/foo

git branch --set-upstream-to=upstream/foo foo

git branch --set-upstream-to=origin/some-qwerty-api-service/feature/application \
    some-qwerty-api-service/feature/application

git branch --set-upstream-to=origin/coreapi-rm-unnecessary-types coreapi-rm-unnecessary-types
```

### [git --track vs --set-upstream vs --set-upstream-to](https://gist.github.com/miku/613ccf7a7030a6f32df1)

## diverge 상태

```bash
hint: You have divergent branches and need to specify how to reconcile them.
hint: You can do so by running one of the following commands sometime before
hint: your next pull:
hint: 
hint:   git config pull.rebase false  # merge
hint:   git config pull.rebase true   # rebase
hint:   git config pull.ff only       # fast-forward only
hint: 
hint: You can replace "git config" with "git config --global" to set a default
hint: preference for all repositories. You can also pass --rebase, --no-rebase,
hint: or --ff-only on the command line to override the configured default per
hint: invocation.
fatal: Need to specify how to reconcile divergent branches.

--- 번역:

분기된 브랜치가 있으며 이를 어떻게 조정할지 명시해야 합니다.
다음 명령 중 하나를 다음 풀(pull) 실행 전에 실행하여 지정할 수 있습니다:

  git config pull.rebase false  # 병합(merge)
  git config pull.rebase true   # 리베이스(rebase)
  git config pull.ff only       # 단순 전진 병합(fast-forward only)

"git config" 대신 "git config --global"을 사용하면 모든 저장소에 대한 기본 설정을 지정할 수 있습니다.
또한 명령어 실행 시 --rebase, --no-rebase, --ff-only 옵션을 추가하여 설정된 기본값을 재정의할 수 있습니다.
오류: 분기된 브랜치를 어떻게 조정할지 명시해야 합니다.
```

### diverge란?

"Diverge"라는 단어는 라틴어 "divergere"에서 유래되었습니다. "di-"는 "두 개의," "apart"를 의미하고, "vergere"는 "향하다"를 의미합니다. 따라서 "divergere"는 글자 그대로 "다른 방향으로 향하다"라는 의미를 가집니다.

다른 맥락에서도 "diverge"는 비슷한 의미로 사용됩니다. 예를 들어, 수학에서는 두 선이 서로 멀어지는 경향을 설명할 때 사용하며, 일반적인 대화에서는 의견, 관심사, 경로 등이 서로 다른 방향으로 발전할 때 사용됩니다. 이러한 사용은 모두 기본적인 의미인 "서로 다른 방향으로 벗어나다"를 공유합니다.

Git에서 "diverge"(분기, 갈라짐 등)는 두 브랜치가 *마지막 공통 커밋 이후 서로 다른 커밋을 가지고 각각 발전*했을 때 사용하는 용어입니다. 이 상태에서는 간단한 fast-forward 병합이 불가능하며, 병합(merge) 또는 리베이스(rebase)를 통해 두 브랜치의 변경사항을 통합해야 합니다.

이 상태에서 `git pull`을 실행하면 Git은 자동으로 이 divergent 상태를 어떻게 처리할지 결정할 수 없기 때문에 오류 메시지를 표시합니다.

### 원인

`git pull`은 기본적으로 `git fetch`와 `git merge FETCH_HEAD`를 연속으로 실행하는 명령어입니다.

여기서 발생하는 문제는 다음과 같습니다:

1. **git fetch**

    원격 저장소의 최신 변경사항을 로컬 저장소로 가져오지만 로컬 브랜치의 파일은 변경하지 않습니다.

2. **git merge FETCH_HEAD**

    가져온 변경사항을 현재 로컬 브랜치와 병합하려 시도합니다.
    이때 로컬과 원격 브랜치가 diverge한 상태라면 Git은 자동으로 병합하지 못하고 사용자에게 입력을 요구합니다.

### 해결 방법

Git은 diverge된 브랜치를 처리하기 위해 세 가지 옵션을 제공합니다:

1. **병합 (Merge)**:
   - `git config pull.rebase false` 또는 단순히 `git pull` 명령어 실행 후 병합을 선택합니다.
   - 로컬과 원격 변경사항을 새로운 'merge commit'을 생성하여 통합합니다.
   - 이 방법은 브랜치의 히스토리를 보존하지만, 브랜치 히스토리가 복잡해질 수 있습니다.

2. **리베이스 (Rebase)**:
   - `git config pull.rebase true` 설정 후 `git pull`을 실행합니다.
   - 로컬 브랜치에서 작업한 커밋들을 임시로 제거하고, 원격 브랜치의 최신 상태 위에 다시 적용합니다.
   - 히스토리가 선형적으로 유지되지만, 리베이스 과정에서 충돌이 발생할 수 있으며, 충돌을 수동으로 해결해야 합니다.

3. **Fast-forward only**:
   - `git config pull.ff only`를 설정하면 Git은 fast-forward 병합만 수행하도록 설정됩니다.
   - 이 설정은 diverge 상태에서 `git pull`을 허용하지 않으므로, 브랜치가 이미 fast-forward 가능한 상태여야 합니다.

4. [origin remote의 특정 파일로 덮어쓰거나 로컬 전체 파일 덮어쓰기](https://stackoverflow.com/questions/3949804/force-overwrite-of-local-file-with-whats-in-origin-repo)

    ```bash
    # 특정 파일로 덮어쓰기
    git fetch
    git checkout origin/main <filepath>
    ```

    ```bash
    # 모든 변경된 파일 덮어쓰기
    git fetch origin
    git reset --hard origin/your-branch-name
    ```

### level rampup1 설명

- `HEAD`:
    - 현재 체크아웃한 커밋에 대한 symbolic name
    - 항상 작업 트리에 반영된 가장 최근 커밋을 가리킨다
    - 일반적으로 `bugFix` 같은 브랜치 이름을 가리킨다
    - 커밋하면, `bugFix`의 상태가 변경되고, 이 변경 사항은 `HEAD` 통해서 볼 수 있다
- `Detaching HEAD`
    - 브랜치 대신 커밋에 `HEAD`를 붙인다는 의미
        1. `HEAD -> main -> C1`와 같은 모습이라고 할 때,
        2. `git checkout C1` 명령어 통해서 C1 커밋으로 체크아웃 하면,
        3. `HEAD -> C1`가 된다

### level rampup1 문제

1. To complete this level, let's detach `HEAD` from `bugFix` and attach it to the commit instead.
    - Specify this commit by its hash.
    - The hash for each commit is displayed on the circle that represents the commit.

```shell
git checkout C4
```

## level rampup2

### level rampup3 설명

- 현실에서는 나이스 한 커밋 트리 시각화가 없을 것이므로, 해시를 보기 위해 `git log`를 사용해야 할 것
- 실제 커밋 해시는 훨씬 길지만, 커밋을 유니크하게 식별할 정도로 충분한 해시 문자를 지정하는 것으로 충분.
    - `fed2da64c0efc5293610bdd892f82a58e8cbc5d8` 대신, `fed2`을 지정하는 것도 가능
- 해시로 커밋을 지정한느 것은 편한 방법이 아니므로, 깃은 relative refs 제공
- `Relative Refs`
    - 이를 통해 어딘가 기억할 수 있는 곳(`bugFix` 또는 `main` 브랜치)에서 시작하고 작업할 수 있다.
    - `^`: 한 번에 하나의 커밋 위로 이동
        - ref 이름에 붙일 때마다, 해당 커밋의 부모를 찾으라는 것
        - `main^`: `main`의 첫번째 부모
        - `main^^`: `main`의 조부모
    - `~<num>`: 숫자만큼 커밋 위로 이동

### level rampup3 문제

1. To complete this level, check out the parent commit of `bugFix`. This will detach `HEAD`. You can specify the hash if you want, but try using relative refs instead!

```shell
git checkout bugFix^
```

## level rampup3

### level rampup3: 설명

- 커밋 트리에서 많은 수준을 이동하고 싶을 경우, `~`(tilde) 연산자에 숫자를 붙여서 그만큼의 부모로 타고 올라갈 수 있다
- `relative refs`를 사용하는 가장 일반적인 방법중 하나는, 브랜치를 이동하는 것(move branches around.)
    - `git branch -f main HEAD~3`: 강제로(`-f`) `main` 브랜치를 `HEAD`의 세 수준 위의 부모로 이동시킨다

### level rampup3: 문제

> Now that you have seen `relative refs` and branch forcing in combination, let's use them to solve the next level.
> To complete this level, move `HEAD`, `main`, and `bugFix` to their goal destinations shown.

```shell
git checkout C1

git branch -f bugFix HEAD^
git branch -f main C6
```

## level rampup4

깃에는 변경 사항을 되돌리는 여러 방법이 있고, 커밋과 마찬가지로, 변경 사항을 되돌리는 것은 두 방법이 있다

1. low-level 컴포넌트(개별 파일 stating 또는 chunks): `git reset`
2. high-level 컴포넌트(변경 사항이 실제로 어떻게 되돌려지는지): `git revert`

### level rampup4: 설명

1. `git reset`
    - 브랜치 참조를 시간상 이전 커밋으로 이동시켜서 변경 사항을 되돌린다
    - 그런 의미에서 "기록을 다시 쓴다(rewriting history)"라고 생각할 수 있다
    - 애초에 커밋이 발생하지 않은 것처럼 브랜치를 되돌린다
    - `git reset HEAD~1`:
        - 현재 체크아웃한 커밋의 부모로 브랜치 참조 이동하고
        - 로컬 리파지토리는 현재 커밋이 없었던 것과 같은 상태가 된다
2. `git revert`
    - "기록을 다시 쓰는 것"은 다른 사람이 사용하는 remote branch에는 작동하지 않는다
    - 변경 사항을 되돌리고 다른 사람과 공유하려면, `git revert` 사용해야 한다
        - 이러면 되돌리고 싶은 커밋 다음에 새로운 커밋이 생긴다.
        - **커밋을 되돌린다**는 변경사항을 가져오기 때문

### level rampup4: 문제

> To complete this level, reverse the most recent commit on both `local` and `pushed`. You will revert two commits total (one per branch).  
> Keep in mind that pushed is a remote branch and local is a local branch -- that should help you choose your methods.

1. `local` 브랜치 되돌리기

```shell
git reset local^
```

2. remote의 `pushed` 브랜치 되돌리기

```shell
git checkout pushed
git revert pushed
```

## level move1

"moving work around"(이 작업을 여기로, 저 작업을 저기로)에 대해 배워보자

### level move1: 설명

1. `git cherry-pick <commit1> <commit2> <...>`
    - 원하는 커밋과 해시를 아는 경우 유용
    - 현재 위치(`HEAD`) 아래의 일련의 커밋을 복사
    - 시나리오
        1. `side` 브랜치에서 작업하고 이를 `main` 브랜치로 복사하려고 할 때,
        2. rebase도 가능하지만, cherry-pick도 가능:
           `git checkout main`
           `git cherry-pick C2 C4`: 체크아웃한 main 다음에 C2, C4 복사한 신규 커밋 생성

### level move1: 문제

> To complete this level, simply copy some work from the three branches shown into main. You can see which commits we want by looking at the goal visualization.

1. `bugFix`(`C3` 커밋) 브랜치

```shell
git cherry-pick C3
```

2. `side`(`C5` 커밋)의 부모 커밋 `C4`

```shell
git cherry-pick C4
```

3. `another`(`C7` 커밋)

```shell
git cherry-pick C7
```

결국 한 명령어로 보자면,

```shell
git cherry-pick C3 C4 C7
```

rebase로 문제를 푼다면?

```shell
git rebase C3
```

## level move2

### level move2: 설명

- 어떤 커밋을 원하는지 모를 경우? interactive rebase 사용해서 리베이스하려는 일련의 커밋 리뷰 가능
- `git rebase -i`
    - 리베이스 대상 아래로 어떤 커밋이 복사되는지 보여주는 UI가 나타난다
    - 커밋 해시, 메시지 등 또한 보여준다.
    - 할 수 있는 것
        1. 커밋의 순서를 바꿀 수 있다
        2. 모든 커밋을 유지하거나, 특정 커밋을 드랍할 수 있다.
        3. squashing(combining) commits
        4. amending commit messages
        5. commit 자체를 수정
- `git rebase -i HEAD~4`
    1. 커밋의 순서를 바꾼다
    2. 그러면 동일한 커밋을 복사한 새로운 커밋이 새로운 순서로 생성

### level move2: 문제

> To finish this level, do an interactive rebase and achieve the order shown in the goal visualization. Remember you can always `undo` or `reset` to fix mistakes :D

- AS-IS: `C0 <- C1 <- C2 <- C3 <- C4 <- C5(main*)`(화살표는 부모를 가리킨다)
- TO-BE: `C0 <- C1 <- C3' <- C5' <- C4'(main*)`

```shell
git rebase -i HEAD~4
```

## level mixed1

### level mixed1: 설명

### level mixed1: 문제
