# Pull

- [Pull](#pull)
    - [`git fetch`와 `git merge FETCH_HEAD` 연속 실행](#git-fetch와-git-merge-fetch_head-연속-실행)
    - [원격이 옳을 때 강제로 로컬 업데이트](#원격이-옳을-때-강제로-로컬-업데이트)
    - [fetch와 pull](#fetch와-pull)
    - [troubleshooting](#troubleshooting)
        - [You have divergent branches and need to specify how to reconcile them](#you-have-divergent-branches-and-need-to-specify-how-to-reconcile-them)

## `git fetch`와 `git merge FETCH_HEAD` 연속 실행

`git pull` 명령어는 원격 저장소의 변경사항을 현재 로컬 브랜치에 통합하는 과정을 자동화한 명령어입니다.
이 명령어는 내부적으로 두 가지 주요 작업을 순차적으로 수행합니다:
1. `git fetch`

    `git fetch` 명령은 원격 저장소에서 최신 변경사항을 가져오지만, 로컬 브랜치의 파일은 변경하지 않습니다.
    이 명령은 원격 저장소의 최신 커밋, 브랜치, 태그 등의 정보를 로컬 저장소의 원격 추적 브랜치에 업데이트합니다.

    예를 들어, `origin/master` 같은 원격 추적 브랜치가 이 단계에서 업데이트됩니다.

2. `git merge FETCH_HEAD`.

    `git fetch`가 완료된 후, `git pull`은 자동으로 `git merge FETCH_HEAD`를 실행합니다.

    `FETCH_HEAD`는 `git fetch`에 의해 마지막으로 가져온 원격 브랜치의 최신 커밋을 가리키는 참조입니다.
    이 명령은 `FETCH_HEAD`가 가리키는 커밋을 현재 체크아웃된 브랜치에 병합합니다.

## 원격이 옳을 때 강제로 로컬 업데이트

원격 저장소의 상태를 강제로 로컬에 적용하려면, 로컬 변경사항을 덮어쓰고 원격 저장소의 상태를 로컬에 반영할 수 있습니다. 이는 로컬의 모든 변경사항과 커밋을 포기한다는 것을 의미하므로 조심해서 수행해야 합니다.

```bash
git fetch origin
git reset --hard origin/your-branch-name
```

여기서 `origin/your-branch-name`은 원격 브랜치의 이름입니다. 이 명령은 로컬 브랜치의 포인터를 원격 브랜치의 최신 커

밋으로 이동시키고, 작업 디렉토리의 모든 변경사항을 원격 브랜치의 상태로 재설정합니다.

## fetch와 pull

`git fetch`와 `git pull`은 모두 원격 저장소의 변경사항을 로컬 저장소로 가져오는 데 사용되는 Git 명령어입니다.

- `git fetch`

    원격 저장소의 최신 커밋, 브랜치, 태그 등의 정보를 로컬 저장소로 가져오지만,
    현재 작업 중인 로컬 브랜치의 상태는 변경하지 않습니다.

    `fetch`는 *원격 저장소의 상태를 확인하고 싶을 때 사용하며, 실제로 로컬의 파일들을 변경하지는 않습니다*.
    정보만 가져오고 병합하지 않습니다.

    ```bash
    git fetch origin
    ```

    이 명령은 원격 저장소인 `origin`의 모든 브랜치 정보를 로컬 저장소에 업데이트하지만,
    현재 체크아웃된 브랜치의 파일은 그대로 유지됩니다.

- `git pull`

    `git fetch` 수행 후 현재 로컬 브랜치와 연결된 원격 브랜치의 변경사항을
    자동으로 현재 브랜치에 병합(`merge`)합니다.

    이는 원격 저장소의 변경사항을 바로 현재 작업 중인 코드에 반영하고 싶을 때 사용합니다.
    정보를 가져온 후 자동으로 병합합니다.

    ```bash
    git pull origin master
    ```

    이 명령은 `origin` 원격 저장소의 `master` 브랜치에서 최신 변경사항을 가져와
    현재 체크아웃된 브랜치에 자동으로 병합합니다.

## troubleshooting

### You have divergent branches and need to specify how to reconcile them

```bash
❯ g pull origin develop    
From https://github.com/private-org/api-service
 * branch                develop    -> FETCH_HEAD
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
```

`git pull origin develop` 명령을 실행했을 때 발생한 메시지는 로컬 브랜치와 원격 브랜치 간에 발생한 분기(divergence)를 어떻게 해결할지 Git이 명확한 지시를 요구하고 있음을 나타냅니다.

여기서 "divergent branches"란 *로컬과 원격 브랜치가 서로 다른 커밋을 가리키고 있어 자동으로 병합이 불가능한 상태*를 말합니다.

이런 상황은 일반적으로 다음과 같은 경우에 발생합니다:

1. **로컬에서 커밋이 발생**: 로컬 브랜치에서 작업을 진행하고 커밋을 추가한 경우.
2. **원격에서 커밋이 발생**: 동시에 다른 개발자가 `origin`의 `develop` 브랜치에 커밋을 추가하고 이를 원격 저장소에 푸시한 경우.

- **FETCH_HEAD 업데이트**

    `git pull` 명령은 내부적으로 `git fetch`를 먼저 실행하여 원격 브랜치의 최신 정보를 로컬로 가져옵니다.
    이 과정에서 `FETCH_HEAD`가 업데이트됩니다.

- **병합 시도 및 충돌 감지**

    이어서 `git merge`가 실행되어 로컬 브랜치와 `FETCH_HEAD` (최근 가져온 원격 브랜치 상태) 사이의 병합을 시도합니다.
    이때 로컬과 원격 브랜치가 서로 다른 경로로 진행된 변경사항(divergence)이 감지되면 자동 병합이 실패합니다.

Git은 이 상황에서 자동으로 병합을 진행할 수 없으므로, 사용자에게 어떻게 병합할지 지시를 요청합니다. 사용자는 다음 중 하나의 방법을 선택할 수 있습니다:

1. **병합(Merge)**:

   `git config pull.rebase false`를 설정하면, 기본적으로 `git pull` 시 병합을 사용하여 로컬 변경사항과 원격 변경사항을 통합합니다.
   이는 두 브랜치의 변경사항을 '병합 커밋'을 통해 통합합니다.

2. **리베이스(Rebase)**:

   `git config pull.rebase true`를 설정하면, 리베이스를 사용하여 로컬 커밋들을 원격 브랜치의 최신 커밋 뒤로 이동시킵니다.
   이는 과거의 커밋 이력을 깔끔하게 유지하는 데 도움이 됩니다.

3. **패스트-포워드(Fast-forward)**:

   `git config pull.ff only`를 설정하면, 패스트-포워드 병합만 허용합니다.
   이는 로컬 브랜치가 원격 브랜치의 직접적인 이력 내에 있을 때만 가능합니다.

`pull.rebase`를 `true`로 설정하고 pull을 하면 문제가 해결됩니다.

```bash
❯ git config --global pull.rebase true 
❯ g pull origin develop                
From https://github.com/private-org/api-service
 * branch                develop    -> FETCH_HEAD
Successfully rebased and updated refs/heads/develop.
```
