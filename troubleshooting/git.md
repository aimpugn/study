# Git Troubleshooting

- [Git Troubleshooting](#git-troubleshooting)
    - [`develop -> develop (non-fast-forward)`](#develop---develop-non-fast-forward)
        - [문제](#문제)
        - [원인](#원인)
            - [`fast-forward`와 `non-fast-forward`](#fast-forward와-non-fast-forward)
            - [`non-fast-forward` 에러 다루기](#non-fast-forward-에러-다루기)
            - [Your branch and 'origin/refactoring\_base' have diverged, and have 13 and 30 different commits each, respectively](#your-branch-and-originrefactoring_base-have-diverged-and-have-13-and-30-different-commits-each-respectively)
        - [해결](#해결)
    - [Need to specify how to reconcile divergent branches](#need-to-specify-how-to-reconcile-divergent-branches)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [Your token has not been granted the required scopes to execute this query](#your-token-has-not-been-granted-the-required-scopes-to-execute-this-query)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [You are in 'detached HEAD' state](#you-are-in-detached-head-state)
        - [문제](#문제-3)
        - [원인](#원인-3)
            - [`HEAD`?](#head)
        - [해결](#해결-3)
    - [GIT Error "hint: Waiting for your editor to close the file](#git-error-hint-waiting-for-your-editor-to-close-the-file)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [fatal: No url found for submodule path 'grpc-merchant-gateway-service/merchant-gateway-service-interface' in .gitmodules](#fatal-no-url-found-for-submodule-path-grpc-merchant-gateway-servicemerchant-gateway-service-interface-in-gitmodules)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [rebase is not possible in the detached head state](#rebase-is-not-possible-in-the-detached-head-state)
        - [문제](#문제-6)
        - [원인](#원인-6)
    - [sh: /usr/local/bin/nvim: No such file or directory](#sh-usrlocalbinnvim-no-such-file-or-directory)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-6)
    - [Repository not found](#repository-not-found)
        - [문제](#문제-8)
        - [원인](#원인-8)
        - [해결](#해결-7)
    - [브랜치 생성 및 커밋 후 push 시에 에러 발생](#브랜치-생성-및-커밋-후-push-시에-에러-발생)
        - [문제](#문제-9)
        - [원인](#원인-9)
            - [에러 메시지의 이해](#에러-메시지의-이해)
        - [해결](#해결-8)
    - [You asked to amend the most recent commit, but doing so would make it empty](#you-asked-to-amend-the-most-recent-commit-but-doing-so-would-make-it-empty)
        - [문제](#문제-10)
        - [원인](#원인-10)
        - [해결](#해결-9)
    - [refusing to allow a Personal Access Token to create or update workflow `${file}` without `workflow` scope](#refusing-to-allow-a-personal-access-token-to-create-or-update-workflow-file-without-workflow-scope)
        - [문제](#문제-11)
        - [원인](#원인-11)
        - [해결](#해결-10)
            - [Personal Access Token(PAT) 갱신 방법](#personal-access-tokenpat-갱신-방법)
    - [You have divergent branches and need to specify how to reconcile them](#you-have-divergent-branches-and-need-to-specify-how-to-reconcile-them)
        - [문제](#문제-12)
        - [원인](#원인-12)
        - [해결](#해결-11)
    - [main, develop 각각에서 release 브랜치 생성 후, main에 머지하려고 할 때 diff가 다름](#main-develop-각각에서-release-브랜치-생성-후-main에-머지하려고-할-때-diff가-다름)
        - [문제](#문제-13)
        - [원인](#원인-13)
        - [해결](#해결-12)

## `develop -> develop (non-fast-forward)`

### 문제

intellij에서 develop 브랜치 업데이트하는데 에러 발생하며 업데이트 실패

```log
10:00:01.398: [transaction-gateway-service] git -c credential.helper= -c core.quotepath=false -c log.showSignature=false fetch origin develop:develop --recurse-submodules=no --progress --prune
From https://github.com/some-org/transaction-gateway-service
 ! [rejected]          develop    -> develop  (non-fast-forward)
```

TGS remote로부터 git pull을 받으려는데 거부됨

```log
15:27:18.631: [transaction-gateway-service-for-review] git -c credential.helper= -c core.quotepath=false -c log.showSignature=false fetch origin refactoring_base:refactoring_base --recurse-submodules=no --progress --prune
From https://github.com/some-org/transaction-gateway-service
 ! [rejected]          refactoring_base -> refactoring_base  (non-fast-forward)
```

그리고 다른 브랜치에서 `refactoring_base`로 체크아웃 하면 아래 같은 정보가 출력

```shell
 rody  ~/IdeaProjects/transaction-gateway-service-for-review   refactor/tosspayments_client ±  gco refactoring_base
M interface/channel-service/channel-service-interface
M interface/transaction-service/transaction-service-interface
Switched to branch 'refactoring_base'
Your branch and 'origin/refactoring_base' have diverged,
and have 13 and 30 different commits each, respectively.
  (use "git pull" to merge the remote branch into yours)
```

### 원인

intellij에서 사용하는 명령어에 문제가 있었던 거로 보인다.

```shell
git -c credential.helper= -c core.quotepath=false -c log.showSignature=false fetch origin refactoring_base:refactoring_base --recurse-submodules=no --progress --prune
```

- [Git에서 커밋을 잃지 않고 원격 리포지토리를 변경할 수 없는 경우가 있습니다. 이 경우 푸시가 거부됩니다.](https://docs.github.com/ko/get-started/using-git/dealing-with-non-fast-forward-errors)

#### `fast-forward`와 `non-fast-forward`

- `fast-forward`
    - 머지 커밋 없이 수행되는 일종의 머지를 설명하기 위한 용어.
    - *다른 브랜치를 가져와 합치려는 브랜치*가 *그 가져오려는 브랜치*의 조상일 때 발생한다.
    - `fast-forward`이라는 용어는 Git이 Git 그래프에 새로운 머지 노드 생성하게 되는 새로운 머지 커밋 없이 빠르고 효율적으로 머지를 수행할 수 있다는 점에 기인한다.
- `non-fast-forward`
    - *다른 브랜치를 가져와 합치려는 브랜치*로부터 *그 가져오려는 브랜치*가 분기(diverged)됐고, 두 브랜치에서 충돌하는 변경 사항이 있을 때 발생한다.
    - 이 경우 Git은 브랜치 포인터를 앞으로 이동시킬 수 없다(can not move the branch pointer **forward**)
    - 대신 Git은 새로운 머지 커밋을 만들어서 두 브랜치의 변경 사항을 가져오고, 이 충돌을 직접 해결해야 한다.

즉, `forward`라는 용어는 현재 브랜치에 머지된 새로운 변경사항들을 반영하기 위해 **Git 그래프에서 브랜치 포인터가 앞으로 이동**하는 것에 기인한다.

#### [`non-fast-forward` 에러 다루기](https://docs.github.com/en/get-started/using-git/dealing-with-non-fast-forward-errors)

종종 git은 커밋을 잃지 않고는 원격 리파지토리를 변경할 수 없게 한다.
다른 사람이 같은 브랜치에 푸시했다면, 나의 변경 사항을 push할 수 없다.

이 경우

1. 원격 브랜치의 변화를 [가져와서 머지](https://docs.github.com/en/get-started/using-git/getting-changes-from-a-remote-repository)하여 수정하거나
2. 간단하게 `git pull`을 사용하여 두 명령어를 한번에 수행할 수 있다.

#### Your branch and 'origin/refactoring_base' have diverged, and have 13 and 30 different commits each, respectively

- [master branch and 'origin/master' have diverged, how to 'undiverge' branches'?](https://stackoverflow.com/questions/2452226/master-branch-and-origin-master-have-diverged-how-to-undiverge-branches)

### 해결

`git pull`을 하면 해결이 됐다

```shell
 rody  ~/IdeaProjects/transaction-gateway-service-for-review  ⇅ refactoring_base ±  git pull
Successfully rebased and updated refs/heads/refactoring_base.
```

아마 `git config --global pull.rebase true`해서 pull 시에 기본적으로 rebase하도록 한 설정 후부터 이런 문제가 발생한 거 같다.

## Need to specify how to reconcile divergent branches

### 문제

```shell
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

### 원인

- [git pull 시 발생하는 warning 해결하기(Need to specify how to reconcile divergent branches)](https://velog.io/@eunddodi/git-pull-%EC%8B%9C-%EB%B0%9C%EC%83%9D%ED%95%98%EB%8A%94-warning-%ED%95%B4%EA%B2%B0%ED%95%98%EA%B8%B0Need-to-specify-how-to-reconcile-divergent-branches)
- [Solved "fatal: Need to specify how to reconcile divergent branches"](https://www.cyberithub.com/solved-fatal-need-to-specify-how-to-reconcile-divergent-branches/)

### 해결

- git config로 해결할 수 있겠지만, 그냥 기존 develop 브랜치를 제거하고 origin/develop을 새로 체크아웃

## Your token has not been granted the required scopes to execute this query

### 문제

private package를 다운받으려는데 받지 못해서 링크로 직접 접근을 해보니, 아래 같은 에러 응답

```log
Your token has not been granted the required scopes to execute this query. The 'id' field requires one of the following scopes: ['read:packages'], but your token has only been granted the: ['repo', 'workflow'] scopes. Please modify your token's scopes at: https://github.com/settings/tokens
```

### 원인

- 새로운 fine grained token을 발급 받아서 사용했는데, 해당 토큰에 적절한 권한이 부여되지 않은 것 같음
- 새로 발급 시에 다른 권한도 추가해봤는데... package에 대한 권한 부여에 실패

### 해결

- classic 토큰을 새로 발급해서 사용
    - `repo`, `workflow`, `write:packages`

## You are in 'detached HEAD' state

### 문제

```log
Note: switching to 'origin/refactoring_base'.

You are in 'detached HEAD' state. You can look around, make experimental
changes and commit them, and you can discard any commits you make in this
state without impacting any branches by switching back to a branch.

If you want to create a new branch to retain commits you create, you may
do so (now or later) by using -c with the switch command. Example:

  git switch -c <new-branch-name>

Or undo this operation with:

  git switch -

Turn off this advice by setting config variable advice.detachedHead to false

HEAD is now at a1be145a refactor: CardInformationClient (#451)
```

### 원인

- [Why did my Git repo enter a detached HEAD state?](https://stackoverflow.com/a/3965714)
    - remote 브랜치를 track하기 전에 체크아웃하면, detached HEAD 상태가 된다
- git HEAD가 브랜치 대신 커밋을 직접 가리키는 경우

```gitlogq
* 226520fb Rename Channel.channelId to Channel.id
*   a1be145a (HEAD, origin/refactoring_base) refactor: CardInformationClient (#451)
|\
| * 991d9d09 Fix wrong package include
```

`git branch`로 현재 로컬의 브랜치 확인하면 아래와 같다. 즉, `origin/refactoring_base` 브랜치가 로컬에 없음

```log
* (HEAD detached at origin/refactoring_base)
  develop
  feature/delete-either
  feature/ksnet
  fix/sync-ksnet-transaction
  main
  refactor/card_info_client
```

#### `HEAD`?

- [git의 내부 - git refs - HEAD](https://git-scm.com/book/ko/v2/Git%EC%9D%98-%EB%82%B4%EB%B6%80-Git-Refs#ref_the_ref)
- ["Git detached head" is a weird error message to receive...](https://www.cloudbees.com/blog/git-detached-head)

> Most of the time, HEAD points to a branch name.
> When you add a new commit, your branch reference is updated to point to it, but HEAD remains the same.
> When you change branches, HEAD is updated to point to the branch you’ve switched to.
> All of that means that, in these scenarios, HEAD is synonymous with “the last commit in the current branch.”
> This is the normal state, in which HEAD is attached to a branch.

- 대부분의 경우 `HEAD`는 브랜치 이름을 가리킨다.
    - 새로운 커밋을 추가하면, 브랜치 참조(reference)는 이 커밋을 가리키도록 업데이트 되지만, `HEAD`는 그대로 유지된다
    - 브랜치를 변경하면, `HEAD`는 변경한 브랜치를 가리키도록 업데이트 된다.
- 그렇다면 `HEAD` === 현재 브랜치의 마지막 커밋

### 해결

원격 브랜치를 이름으로 체크아웃

```shell
 git checkout -b refactoring_base origin/refactoring_base --
```

또는 [`git switch` 사용](https://stackoverflow.com/a/1783426)

## GIT Error "hint: Waiting for your editor to close the file

### 문제

`grbi` 테스트 해보는데, 아래 같은 에러 발생

```log
grbi
hint: Waiting for your editor to close the file... fatal: cannot run /usr/local/bin/nvim: No such file or directory
error: unable to start editor '/usr/local/bin/nvim
```

### 원인

- [GIT Error "hint: Waiting for your editor to close the file](https://stackoverflow.com/questions/69209854/git-error-hint-waiting-for-your-editor-to-close-the-file)

### 해결

> You have set up your editor to be something like notepad or any other external program for your editor.
> Once you try to commit it open, you have to type your message and close it.

에디터 설정을 해야 한다.

editor 설정되어 있는지 확인은 아래 명령어로 가능

```shell
git config -l --show-origin | grep editor
```

[`core.editor`로 설정](https://git-scm.com/book/en/v2/Customizing-Git-Git-Configuration#_core_editor) 가능

```shell
git config --global core.editor emacs
```

## fatal: No url found for submodule path 'grpc-merchant-gateway-service/merchant-gateway-service-interface' in .gitmodules

### 문제

### 원인

### 해결

```shell
git rm -r --cached --ignore-unmatch presentation/transaction-gateway-service-interface
git rm -r --cached --ignore-unmatch grpc-merchant-gateway-service/merchant-gateway-service-interface
git rm -r --cached --ignore-unmatch interface/merchant-gateway/merchant-gateway-interface
git submodule init
git submodule update --remote
```

## rebase is not possible in the detached head state

### 문제

Intellij에서 A 브랜치를 B 브랜치 위로 리베이스 하려는데, rebase 버튼이 비활성화 되어 있고, 버튼 위에 마우스를 올리니 아래 같은 메시지로 이유를 알려줌

```log
rebase is not possible in the detached head state
```

### 원인

기존에 rebase가 진행중이던 게 남아있었던 거 같다. iterm 터미널로 보면 아래 같이 나오고 있었다.

```log
rody  ~/IdeaProjects/paymentadmin   refactor/PORT-519/channel-migration-feature >R>
```

## sh: /usr/local/bin/nvim: No such file or directory

### 문제

```log
sh: /usr/local/bin/nvim: No such file or directory
```

### 원인

설정된 editor 경로가 맞지 않음

> `--get`
>
> Get the value for a given key (optionally filtered by a regex matching the value). Returns error code 1 if the key was not found and the last value if multiple key values were found.

```shell
git config --get core.editor
```

### 해결

editor 설정 수정

```shell
git config --global core.editor emacs
```

## Repository not found

### 문제

```log
17:47:54.758: [some-qwerty-org.io-go] git -c core.quotepath=false -c log.showSignature=false fetch origin feature/corewebhook:feature/corewebhook --recurse-submodules=no --progress --prune
remote: Repository not found.
fatal: repository 'https://github.com/some-org/some-qwerty-org.io-go.git/' not found
```

### 원인

```shell
# 현재 사용중인 credential helper 확인
❯ git config --global credential.helper
osxkeychain


# credential 정보 확인
echo "protocol=https
host=github.com" | git credential fill
protocol=https
host=github.com
username=aimpugn
password=gho_TOKEN # expired


# 권한 있는지 확인. 근데 없음.
❯ g ls-remote
remote: Repository not found.
fatal: repository 'https://github.com/some-org/some-qwerty-org.io-go.git/' not found


# 어떤 크리덴셜에 권한이 있는지 여부 확인
❯ g ls-remote https://aimpugn:<VALID_CREDENTIAL>@github.com/some-org/some-qwerty-org.io-go.git
f326223925bde5f9498415ff2e7ce6f8cb48d810    HEAD
d5a7b6416189e140b1de27442f276b75fcf5d1f6    refs/heads/some-qwerty-api-service/feature/application
f789294769853904cdf0dc31ae13c3a13f63461f    refs/heads/corepay/chore/delete-unnecessary-logic-on-danal-tpay-ctp
0c9649a68ced2fd404f94911aa01ca6e8b1f1ee1    refs/heads/corepay/chore/depr-ctp-gen


# 하지만 만료된 토큰으로 실행하면 에러 발생
❯ g ls-remote https://aimpugn:gho_TOKEN@github.com/some-org/some-qwerty-org.io-go.git
remote: Repository not found.
fatal: repository 'https://github.com/some-org/some-qwerty-org.io-go.git/' not found
```

### 해결

- `Keychain Access`에서 `github.com`의 credential 업데이트

## 브랜치 생성 및 커밋 후 push 시에 에러 발생

### 문제

```shell
fatal: The current branch feature/verify-phone-number-when-web has no upstream branch.
To push the current branch and set the remote as upstream, use

    git push --set-upstream origin feature/verify-phone-number-when-web

To have this happen automatically for branches without a tracking
upstream, see 'push.autoSetupRemote' in 'git help config'.
```

### 원인

Git에서 새로 생성된 로컬 브랜치를 원격 저장소에 아직 연결하지 않았을 때 발생

이와 관련된 Git의 원리는?

- 로컬 및 원격 저장소:
    - Git은 분산 버전 관리 시스템
    - 개발자는 로컬 저장소에서 작업을 수행하고, 원격 저장소와의 동기화를 통해 협업을 진행한다
- 브랜치
    - Git에서 브랜치는 코드 변경사항을 분리하여 관리하기 위한 독립적인 경로
    - 개발자는 기능 추가, 버그 수정 등을 위해 새 브랜치를 만들고,
    - 작업 후에는 해당 작업 브랜치를 기본 브랜치(보통 `master` 또는 `main`)에 병합한다
- 로컬 브랜치와 원격 추적 브랜치
    - `로컬 브랜치`: 개발자의 컴퓨터에만 존재하는 독립된 라인의 작업을 나타낸다
    - `원격 추적 브랜치`: 원격 저장소에 있는 브랜치의 상태를 추적한다

#### 에러 메시지의 이해

- "The current branch `feature/verify-phone-number-when-web` has no upstream branch."
    - 현재 작업 중인 로컬 브랜치가 원격 저장소의 어떤 브랜치와도 연결되어 있지 않다는 것을 의미
- 이는 `git checkout -b [branch-name]`으로 새 브랜치를 생성했지만, 이 브랜치를 원격 저장소에 아직 등록(push)하지 않았기 때문에 발생

### 해결

- Git은 로컬 브랜치를 처음 원격 저장소에 푸시할 때, `로컬 브랜치`가 어떤 `원격 추적 브랜치`와 연결될지를 설정

    ```shell
    git push --set-upstream origin [branch-name]
    ```

    - `--set-upstream`: 로컬 브랜치와 원격 브랜치 간의 연결(업스트림)을 설정
    - `origin`: 원격 저장소의 기본 이름
    - `[branch-name]`: 푸시하려는 로컬 브랜치의 이름

- 위 명령을 실행하면, 로컬 브랜치가 원격 저장소의 새 브랜치로 푸시되고, 두 브랜치 간의 연결이 설정된다. 이후로는 `git push` 명령만으로도 해당 브랜치에 변경사항을 푸시할 수 있다

## You asked to amend the most recent commit, but doing so would make it empty

### 문제

```log
Rebasing (1/10)
Rebasing (2/10)

You asked to amend the most recent commit, but doing so would make it empty.
You can repeat your command with `--allow-empty`, or you can remove the commit entirely with "git reset HEAD^". interactive rebase in progress; onto `ea8ec79d4`

Last commands done (2 commands done):
reword b2e228299 some-qwerty-api-service/payments: NO CHANGE
fixup 793b49982 some-qwerty-api-service/payment.go: Refactor remove errors

Next commands to do (8 remaining commands):
pick c4a42a563 some-qwerty-api-service/README.md: Add
pick edda2ebdb refactor: `env`를 payments 패키지에서 변환해서 사용하도록 수정
(use "git rebase --edit-todo" to view and edit) You are currently rebasing branch 'some-qwerty-api-service/feature/application-arranged' on 'ea8ec79d4'. (all conflicts fixed: run "git rebase --continue")

Untracked files: (use "git add <file>..." to include in what will be committed)
some-qwerty-api-service/.env.local
some-qwerty-api-service/.env.prod
some-qwerty-api-service/.env.prod.bak
some-qwerty-api-service/.env.stg
some-qwerty-api-service/payments/receipt/mock_ReceiptCommand.go
corepay/.env.local
corepay/tmp/
portdoc/some-qwerty-org.io-go.tar.gz
some-qwerty-org.io-go.iml
result.json

No changes

Could not apply `793b49982`... some-qwerty-api-service/payment.go: Refactor remove errors

Local changes were shelved before rebase.
```

### 원인

제공된 로그를 분석하면, Git rebase 과정 중에 몇 가지 문제가 발생한 것으로 보입니다. 주요 문제는 다음과 같습니다:

1. "Empty Commit" 문제: "You asked to amend the most recent commit, but doing so would make it empty." 이 메시지는 가장 최근의 커밋을 수정(amend)하려 했으나, 해당 커밋에 변경 사항이 없어서 비어 있는 커밋이 되었다는 것을 의미합니다. 이럴 때는 커밋을 유지하고 싶다면 `--allow-empty` 옵션을 사용하거나, 아니면 해당 커밋을 제거할 수 있습니다.

2. 충돌 해결 후 커밋 적용 실패: "Could not apply `793b49982`... some-qwerty-api-service/payment.go: Refactor remove errors" 이 메시지는 특정 커밋을 적용하는 데 실패했다는 것을 의미합니다. 이는 보통 충돌(conflict) 때문에 발생합니다.

3. Untracked Files: rebase 과정 중에 추적되지 않는 파일들이 존재합니다. 이 파일들은 rebase에 직접적으로 영향을 주지는 않지만, 작업 디렉토리의 상태를 나타냅니다.

### 해결

1. 빈 커밋 처리:
   - 빈 커밋을 유지하고 싶지 않다면, 해당 커밋을 제거할 수 있다

     ```bash
     git rebase --skip
     ```

   - 빈 커밋을 유지하고 싶다면, `--allow-empty` 옵션을 사용하여 rebase를 계속 진행

     ```bash
     git rebase --continue --allow-empty
     ```

2. 커밋 적용 실패 해결:
   - 충돌이 발생한 경우, 충돌을 해결한 후에 `git add <파일명>`으로 변경 사항을 스테이징하고, `git rebase --continue`로 rebase를 계속 진행
   - 만약 충돌이 없고 커밋 적용에 실패한 경우, `git rebase --skip`을 사용하여 해당 커밋을 건너뛰고 다음 커밋으로 넘어간다

3. Untracked Files 관리:
   - rebase와 관련 없는 파일들은 현재 작업과 상관없다면 무시하거나, 필요에 따라 `.gitignore` 파일에 추가할 수 있다
   - 중요한 파일이면 별도로 백업을 해두는 것이 좋다

4. IntelliJ에서의 Rebase:
   - Retry 로 다시 시도하면 이어서 계속 진행된다

## refusing to allow a Personal Access Token to create or update workflow `${file}` without `workflow` scope

### 문제

```log
!    refs/heads/some-qwerty-api-service/fix/broken-unsued-lint-check:refs/heads/some-qwerty-api-service/fix/broken-unsued-lint-check    [remote rejected] (refusing to allow a Personal Access Token to create or update workflow `.github/workflows/golangci-lint.yml` without `workflow` scope)
```

### 원인

이 에러 메시지는 GitHub에서 발생하는 것으로, GitHub Actions 워크플로우 파일(`.github/workflows/golangci-lint.yml`)을 생성하거나 업데이트하려고 할 때, 사용 중인 Personal Access Token(PAT)에 필요한 권한(`workflow` 스코프)이 없음을 나타냅니다. GitHub Actions는 GitHub의 CI/CD 기능을 제공하며, 워크플로우 파일을 통해 빌드, 테스트, 배포 등의 자동화된 작업을 정의합니다.

### 해결

1. **Personal Access Token(PAT) 확인**: 먼저, 현재 사용 중인 Personal Access Token의 권한을 확인합니다. GitHub에서는 보안상의 이유로 워크플로우 파일을 생성하거나 수정할 때 `workflow` 스코프가 있는 토큰을 요구합니다.

2. **PAT에 `workflow` 스코프 추가**: PAT를 생성하거나 수정하여 `workflow` 스코프를 포함시키는 것이 필요합니다. 다음 단계를 따라 `workflow` 스코프를 추가할 수 있습니다:
   - GitHub에서 우측 상단의 프로필 아이콘을 클릭한 후, "Settings"를 선택합니다.
   - 왼쪽 사이드바에서 "Developer settings"를 클릭합니다.
   - "Personal access tokens"을 선택하고, "Generate new token" 또는 기존 토큰을 선택하여 "Edit"을 클릭합니다.
   - "Select scopes" 섹션에서 `workflow` 체크박스를 선택합니다.
   - 페이지 하단에서 "Generate token"을 클릭하여 변경사항을 저장합니다.

3. **변경된 PAT 사용**: 새로운 권한이 포함된 PAT를 사용하여 워크플로우 파일을 생성하거나 업데이트합니다. 이를 위해 로컬 Git 설정에서 사용 중인 토큰을 업데이트하거나, GitHub Actions에서 사용 중인 토큰을 업데이트해야 할 수 있습니다.

4. **GitHub Actions 재실행**: 필요한 권한이 포함된 PAT를 사용하도록 설정한 후, GitHub Actions 워크플로우를 다시 실행하여 문제가 해결되었는지 확인합니다.

로컬 작업 사본에서 Git 토큰을 갱신하는 방법은 사용하는 인증 방식과 Git 저장소 호스팅 서비스(예: GitHub, GitLab, Bitbucket 등)에 따라 다를 수 있습니다. 여기서는 GitHub을 예로 들어 설명하겠습니다. GitHub에서는 Personal Access Token(PAT)을 사용하여 Git 명령어를 인증하는 방법을 권장합니다.

#### Personal Access Token(PAT) 갱신 방법

1. **새로운 PAT 생성**:
   - GitHub에서 우측 상단의 프로필 아이콘을 클릭한 후, "Settings"를 선택합니다.
   - 왼쪽 사이드바에서 "Developer settings"를 클릭합니다.
   - "Personal access tokens"을 선택하고, "Generate new token"을 클릭합니다.
   - 필요한 권한을 선택하고, "Generate token"을 클릭하여 새로운 토큰을 생성합니다.
   - 생성된 토큰을 안전한 곳에 복사해 두세요. 페이지를 벗어나면 다시 볼 수 없습니다.

2. **로컬 Git 설정에서 PAT 갱신**:
   - Git은 HTTPS를 통해 원격 저장소에 접근할 때 인증 정보를 캐싱할 수 있습니다. 이 캐시된 인증 정보를 새로운 PAT으로 갱신해야 합니다.
   - 인증 정보를 캐싱하는 방법은 운영 체제에 따라 다릅니다. 일반적으로 사용되는 몇 가지 방법은 다음과 같습니다.

     **Windows에서**:
     - Git Credential Manager를 사용하고 있다면, 다음 명령어를 실행하여 인증 정보를 삭제합니다.

       ```bash
       git credential-manager-core erase
       ```

     - 다음에 GitHub에 접근할 때 새로운 PAT을 입력하라는 메시지가 나타납니다.

     **macOS에서**:
     - 키체인 접근(Keychain Access) 애플리케이션을 열고, GitHub에 관련된 인증 정보를 찾아 삭제합니다.
     - 다음에 GitHub에 접근할 때 새로운 PAT을 입력하라는 메시지가 나타납니다.

     **Linux에서**:
     - Git Credential Storage를 사용하고 있다면, 다음 명령어를 실행하여 인증 정보를 삭제합니다.

       ```bash
       git credential-cache exit
       ```

     - 또는 `~/.git-credentials` 파일을 직접 편집하여 오래된 토큰을 새로운 토큰으로 교체할 수 있습니다.

3. **새로운 PAT 사용**:
   - 다음 Git 명령어를 실행할 때, 사용자 이름은 GitHub 사용자 이름을, 비밀번호는 새로운 PAT을 입력합니다.

이러한 단계를 통해 로컬 작업 사본에서 Git 토큰을 갱신할 수 있습니다. 각 단계에서 발생할 수 있는 구체적인 상황이나 오류 메시지에 따라 추가 조치가 필요할 수 있습니다.

## You have divergent branches and need to specify how to reconcile them

### 문제

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

### 원인

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

### 해결

`pull.rebase`를 `true`로 설정하고 pull을 하면 문제가 해결됩니다.

```bash
❯ git config --global pull.rebase true
❯ g pull origin develop
From https://github.com/private-org/api-service
 * branch                develop    -> FETCH_HEAD
Successfully rebased and updated refs/heads/develop.
```

## main, develop 각각에서 release 브랜치 생성 후, main에 머지하려고 할 때 diff가 다름

### 문제

우선 전체적인 흐름은 다음과 같습니다.

1. develop에서 파생되었고 main으로 rebase & merge 될 release 브랜치에 A, B, C 기능이 있음
2. C 기능이 먼저 나가야 하고, 혹시 오작동 일으킬까봐 release에서 A, B를 Revert
3. C 기능 배포해서 main에 머지. 이때 A, B를 Revert한 커밋이 main에 포함
4. C 기능 관련해서 배포를 나중에 다시 해달라는 요청으로 배포는 기존 버전으로 배포하고, main 및 develop에서 C 기능 커밋 제거
5. 근데 A, B를 Revert한 커밋은 여전히 main에 남아 있음
6. 이제 다시 A, B Revert한 것을 다시 Revert 하려고 하는데 아래와 같은 문제 발생

```mermaid
gitGraph
    commit id: "Initial state"
    branch develop
    checkout develop
    commit id: "Feat A"
    commit id: "Feat B"
    commit id: "Feat C"
    branch release
    checkout release
    commit id: "Create release with A, B, C"
    commit id: "PR: Revert A" type: REVERSE
    commit id: "PR: Revert B" type: REVERSE
    checkout main
    merge release id: "Release with C, Revert A&B"
    checkout develop
    merge main id: "Sync develop with main"
    commit id: "Remove C from develop"
    checkout main
    commit id: "Remove C from main"
    checkout develop
    commit id: "Try to restore A&B"
```

현재 상황:
- main 브랜치에는 A와 B의 Revert 커밋이 남아있습니다.
- develop 브랜치에는 A와 B의 원래 커밋, Revert 커밋, 그리고 복원 시도 커밋이 모두 존재합니다.
- release 브랜치는 A와 B가 Revert된 상태로 main에 병합되었습니다.

이 그래프는 release 브랜치에서 A와 B를 Revert하는 PR 과정을 명확히 보여주며, 이로 인해 발생한 복잡한 커밋 히스토리를 나타냅니다. 이런 상황에서 develop에서 main으로 병합하려 할 때 예상치 못한 커밋들이 포함되거나 충돌이 발생하는 이유를 더 잘 이해할 수 있습니다.

develop 브랜치 기준으로 release 브랜치를 생성하여
main으로 rebase & merge 하는 PR 생성시에 아래처럼 포함돼선 안되는 커밋들이 포함되는데요

```bash
Feat: A feature (8bb7437ab57a163f3074d6b983b5cd7361719638)

Revert "Feat: A feature" (ef5cf575dbab32f22129f8a08c0e12341d78d7e8)

Revert "Feat: B feature" (67ae79d7ce7ab5c9a9f2b5a80c9278a65a5792fc)
```

근데 main 브랜치 기준으로 새로운 release 브랜치 생성하여 main에 rebase & merge 하는 PR을 생성하면,
제가 원하는 두 개의 커밋만 머지 대상으로 나오고, 파일 idff도 정상적으로 나옵니다.

```bash
Revert "Revert "Feat: A feature""
Revert "Revert "Feat: B feature""
```

그래서 develop과 main의 커밋 해시를 확인해보니 아래와 같은데요.

- develop의 커밋

    ```bash
    Revert "Feat: A feature" (ef5cf575)

    This reverts commit 779d9e2c1e68fc6b2c562d1ce97c06d2bc82378d.


    Revert "Feat: B feature" (67ae79d7)

    This reverts commit 735413b8
    ```

- main의 커밋

    ```bash
    Revert "Feat: A feature" (4cf90f1f)

    This reverts commit 779d9e2c1e68fc6b2c562d1ce97c06d2bc82378d.


    Revert "Feat: B feature" (c84c96d3)

    This reverts commit 735413b8
    ```

main에 있는 커밋과 develop에 있는 커밋이 메시지는 같은데 커밋 해시가 다릅니다.

제가 원하는 것은 develop에서 release 생성하여 main에 머지하는 것인데,
develop에서 release 생성하여 main에 머지하려고 하면
- 문제가 되는 Revert 커밋을 drop 커밋할 때:
    PR에서 "This branch has conflicts that must be resolved"라고 하면서
    Resolve conflicts 하라고 나오고

    PR의 Commits 탭에는
    Feat: A feature (8bb7437ab57a163f3074d6b983b5cd7361719638)

    이 커밋만 잡힙니다.

- 문제가 되는 Revert 커밋을 다시 Revert 커밋할 때:
    잡히면 안되는 커밋까지 잡히면서

    ```bash
    Feat: A feature (8bb7437ab57a163f3074d6b983b5cd7361719638)

    Revert "Feat: A feature" (ef5cf575dbab32f22129f8a08c0e12341d78d7e8)

    Revert "Feat: B feature" (67ae79d7ce7ab5c9a9f2b5a80c9278a65a5792fc)

    # 아래가 실제로 반영 원하는 커밋
    Revert "Revert "Feat: A feature"" (61eabaa3acd20b4f50cc24df56b9512731cd7342)

    Revert "Revert "Feat: B feature"" (5d3c815417479f55da1816f485d14df475a98139)
    ```

이렇게 나옵니다.

문제의 원인과 해결 방법을 git과 github 작동 원리에 기반하여 이해하고 해결하고 싶습니다.

### 원인

이 문제는 Git의 커밋 히스토리 관리 방식과 브랜치 간 동기화 과정에서 발생하는 복잡성에서 비롯됩니다.

주요 원인은 다음과 같습니다:

1. 커밋 해시의 불일치
2. Revert 커밋의 특성
3. 브랜치 간 동기화 문제
4. Git의 merge 및 rebase 전략

단계별 상세 설명:

1. 초기 상태:
    `develop`과 `main` 브랜치가 동기화된 상태였습니다.

2. 기능 개발 (develop 브랜치):
    A, B, C 기능이 develop 브랜치에서 개발되었습니다.
    각 기능은 별도의 커밋으로 추가되었을 것입니다.

3. Release 브랜치 생성 및 기능 Revert:
    `develop`에서 `release` 브랜치를 생성했습니다.
    C 기능만 배포하기 위해 `release` 브랜치에서 A와 B 기능을 Revert했습니다.
    이 과정에서 다음과 같은 Revert 커밋이 생성되었습니다:

    ```plaintext
    Revert "Feat: A feature" (ef5cf575)
    Revert "Feat: B feature" (67ae79d7)
    ```

4. Main 브랜치로의 Rebase & Merge:
    `release` 브랜치를 `main` 브랜치로 rebase & merge 했습니다.
    이 과정에서 C 기능과 함께 A, B 기능의 Revert 커밋이 `main`에 포함되었습니다.
    GitHub의 rebase & merge 옵션은 fast-forward merge를 수행하므로, 새로운 merge 커밋을 생성하지 않습니다.

    `main` 브랜치의 Revert 커밋 해시는 `release` 브랜치의 해시와 동일합니다:

    ```plaintext
    Revert "Feat: A feature" (ef5cf575)
    Revert "Feat: B feature" (67ae79d7)
    ```

    Rebase & merge 과정에서 `release` 브랜치의 커밋들이 `main` 브랜치의 최신 커밋 위에 재배치됩니다.
    이 과정에서 커밋 내용은 동일하게 유지되며, 커밋 해시도 변경되지 않습니다.
    결과적으로 `main` 브랜치는 `release` 브랜치의 상태를 그대로 가져오게 됩니다.

5. Develop 브랜치 동기화:
    `develop` 브랜치를 `main` 브랜치와 동기화할 때,
    이미 `main`에 존재하는 Revert 커밋들이 `develop`에도 포함됩니다.

    이로 인해 `develop` 브랜치에는 다음과 같은 상황이 발생합니다:
    - A, B 기능의 원래 커밋
    - C 기능 커밋
    - A, B 기능을 Revert하는 커밋 (main에서 가져온 것)

    이는 `develop` 브랜치에 모순된 상태를 만들어, 이후 A와 B 기능을 복원하려 할 때 문제가 발생할 수 있습니다.

6. C 기능 제거:
    C 기능 배포 연기로 인해 main과 develop에서 C 기능 관련 커밋을 제거했습니다.
    그러나 A와 B 기능의 Revert 커밋은 여전히 main에 남아있게 되었습니다.

7. A, B 기능 복원 시도:
    이제 A와 B 기능을 다시 복원하려고 시도하고 있습니다.
    이 과정에서 다음과 같은 문제가 발생합니다:

    - Revert 커밋을 drop하려 할 때:
        Git은 이미 main에 존재하는 Revert 커밋과 develop의 Revert 커밋을 다른 것으로 인식합니다.
        이로 인해 충돌이 발생하고, Git은 사용자에게 충돌 해결을 요구합니다.

    - Revert 커밋을 다시 Revert하려 할 때:
        이 경우, Git은 원래의 기능 커밋, Revert 커밋, 그리고 Revert의 Revert 커밋을 모두 별개의 커밋으로 인식합니다.
        결과적으로 원하지 않는 커밋들까지 PR에 포함되게 됩니다.

기술적 설명:

1. Git의 커밋 식별:
    - Git은 SHA-1 해시를 사용하여 각 커밋을 고유하게 식별합니다.
    - 이 해시는 커밋의 내용, 작성자, 시간, 부모 커밋 등을 모두 고려하여 생성됩니다.
    - 따라서 같은 변경사항이라도 다른 시점이나 다른 브랜치에서 커밋되면 다른 해시를 갖게 됩니다.

2. Git의 merge 전략:
    - Git은 merge 시 3-way merge 알고리즘을 사용합니다.
    - 이 알고리즘은 두 브랜치의 공통 조상을 찾아 변경사항을 비교합니다.
    - 복잡한 merge 상황에서는 예상치 못한 결과가 발생할 수 있습니다.

3. Revert 커밋의 특성:
    - Revert 커밋은 이전 커밋의 변경사항을 정확히 반대로 적용합니다.
    - 그러나 Git의 관점에서 이는 새로운 변경사항입니다.
    - 따라서 Revert의 Revert는 원래 커밋과 동일한 효과를 가지지만, Git은 이를 다른 커밋으로 취급합니다.

4. GitHub의 PR 처리:
    - GitHub는 PR을 생성할 때 base 브랜치(이 경우 main)와 compare 브랜치(이 경우 release) 사이의 차이를 분석합니다.
    - 이 과정에서 GitHub는 Git의 merge-base 알고리즘을 사용하여 두 브랜치의 공통 조상을 찾습니다.
    - PR에 표시되는 커밋은 이 공통 조상 이후에 추가된 커밋들입니다.

---

문제의 원인은 다음으로 예상됩니다.

1. **커밋 해시의 불일치**:
   - `rebase`는 기존 커밋을 기반으로 새로운 커밋을 생성하므로 커밋 해시가 변경됩니다
   - `rebase` 과정에서 동일한 커밋 메시지를 가진 새로운 커밋이 생성되어, `develop` 브랜치와 `main` 브랜치에 동일한 내용의 커밋이 다른 해시로 존재하게 됩니다.

2. **Revert 커밋의 복잡성**:
   - `Revert "Feat: A feature"` 같은 커밋이 `develop`과 `main` 브랜치에 서로 다른 해시로 존재합니다.
   - 이는 `rebase` 중 커밋 메타데이터(커미터, 시간 등)가 변경되었기 때문입니다【22†source】.

3. **충돌 해결의 어려움**:
   - `develop` 브랜치에서 `release` 브랜치를 생성하고 `main` 브랜치로 `rebase` & `merge`하려고 할 때, 예상치 못한 커밋들이 포함되어 충돌이 발생합니다.
   - `Revert` 커밋을 삭제하거나 다시 `Revert`하는 과정에서 불필요한 커밋이 추가되며 히스토리가 복잡해집니다.

Git과 GitHub의 동작 원리 기반하여 분석하면 다음과 같습니다.

1. **커밋 해시의 생성**:
   - Git은 커밋의 내용뿐 아니라 커밋 시점, 부모 커밋 해시, 저자 정보 등 여러 요소를 기반으로 해시를 생성합니다. 따라서 같은 내용의 커밋이라도 해시가 달라질 수 있습니다.
   - 이는 `rebase` 과정에서 특히 문제가 됩니다. `rebase`는 커밋을 새로 생성하여 히스토리를 재작성하므로, 새로운 커밋 해시가 생성됩니다【23†source】.

2. **Rebase와 Merge의 차이**:
   - `merge`는 두 브랜치를 통합하며, 기존 커밋 히스토리를 보존합니다. 반면, `rebase`는 한 브랜치의 커밋을 다른 브랜치의 커밋 뒤로 재배치하여 히스토리를 재작성합니다.
   - `rebase` 과정에서 충돌이 발생할 수 있으며, 이는 수동으로 해결해야 합니다. 또한, 기존 커밋이 새로운 커밋으로 재작성되므로 커밋 해시가 변경됩니다【23†source】【24†source】.

3. **Revert 커밋 처리**:
   - `Revert` 커밋은 원래 커밋을 되돌리는 역할을 합니다. 다시 `Revert`하는 경우, 원래 커밋과 동일한 내용을 가지지만 새로운 해시를 가지는 커밋이 생성됩니다.
   - 이러한 Revert 커밋들이 중복되거나 예상치 못한 위치에 포함되어 히스토리가 복잡해질 수 있습니다.

### 해결

1. **Merge 사용**:
    `rebase` 대신 `merge` 전략을 사용하여 브랜치를 통합합니다.
    `merge`는 히스토리를 그대로 유지하므로 커밋 해시가 변경되지 않습니다.

    ```bash
    # develop 브랜치를 최신 상태로 업데이트
    git checkout develop
    git pull origin develop

    # develop 브랜치에서 새로운 release 브랜치를 생성
    git checkout -b release

    # main 브랜치의 변경사항을 release 브랜치로 병합
    git merge main

    # 충돌 발생 시 수동으로 해결
    git commit
    git push origin release
    ```

    git merge는 두 브랜치의 히스토리를 그대로 유지하면서 통합합니다.
    커밋 해시가 변경되지 않으므로, 기존 커밋들의 식별성이 유지됩니다.
    충돌이 발생하면 수동으로 해결할 수 있어, 세밀한 제어가 가능합니다.

    이 방법은 현재 상황을 해결할 수 있지만, 몇 가지 주의점이 있습니다:
    1. merge 커밋이 생성되어 히스토리가 복잡해질 수 있습니다.
    2. 불필요한 Revert 커밋들이 그대로 유지될 수 있습니다.
    3. 충돌 해결 과정에서 실수할 경우, 원하지 않는 변경사항이 포함될 수 있습니다.

2. **불필요한 Revert 커밋 정리**:
    불필요한 Revert 커밋을 정리하고, 필요한 경우 새롭게 Revert 커밋을 생성하여 히스토리를 명확하게 유지합니다.

    ```bash
    # develop 브랜치로 전환합니다.
    git checkout develop

    # 최근 n개의 커밋에 대해 interactive rebase를 시작합니다.
    git rebase -i HEAD~n  # n은 제거하려는 커밋 수

    # Revert 커밋을 삭제하고 저장
    git push -f origin develop
    ```

    불필요한 Revert 커밋을 제거하여 히스토리를 깔끔하게 정리합니다.
    원하는 기능(A와 B)을 다시 포함시키면서, 불필요한 Revert 기록은 제거합니다.

    이 방법은 효과적으로 문제를 해결할 수 있습니다:
    1. 히스토리가 깔끔하게 정리됩니다.
    2. 원하는 기능만 남기고 불필요한 Revert 커밋을 제거할 수 있습니다.
    3. 그러나 *강제 푸시(`-f`)를 사용하므로, 팀원들과의 협업 시 주의*가 필요합니다.

3. **충돌 해결 및 PR 생성**:

    PR 생성 시 발생하는 충돌을 수동으로 해결하고, 불필요한 커밋을 제거한 후 PR을 진행합니다.
    이 과정에서 GitHub의 커밋 스쿼싱(squash) 기능을 사용하여 커밋을 합치는 것도 고려할 수 있습니다.

    이 방법도 문제를 해결할 수 있지만, 몇 가지 고려사항이 있습니다:
    1. 충돌 해결 과정이 복잡할 수 있으며, 실수의 여지가 있습니다.
    2. 스쿼시를 사용하면 세부적인 커밋 히스토리가 손실될 수 있습니다.
    3. PR의 변경사항을 정확하게 제어할 수 있지만, 시간이 많이 소요될 수 있습니다.

---

해결 방법:

1. 인터랙티브 리베이스(Interactive Rebase) 사용

    ```bash
    git checkout -b new-release main
    git rebase -i HEAD~n  # n은 수정하고자 하는 커밋 수
    ```

    이 과정에서 A와 B 기능의 Revert 커밋을 drop하고, 필요한 커밋만 남깁니다.

    인터랙티브 리베이스는 각 커밋을 개별적으로 처리할 수 있습니다.
    따라서 커밋 히스토리를 재작성하여 불필요한 커밋을 제거하거나 수정할 수 있습니다.

    이 과정에서 `pick` 대신 `drop`을 선택하여 불필요한 커밋을 제거할 수 있습니다.

2. 원하는 변경사항 적용

    A와 B 기능을 다시 추가하는 새로운 커밋을 생성합니다.
    새로운 커밋을 생성하면 Git은 이를 완전히 새로운 변경사항으로 인식합니다.

    Revert 커밋을 제거한 후 A와 B 기능을 다시 추가하면, 이는 새로운 커밋으로 기록되므로 이전 Revert 히스토리와 무관하게 깨끗한 상태를 유지할 수 있습니다.

3. 강제 푸시

    강제 푸시는 로컬 브랜치의 히스토리를 원격 브랜치로 덮어쓰는 방법입니다.

    ```bash
    git push origin new-release --force
    ```

    `--force` 옵션을 사용하면 기존의 복잡한 히스토리를 새롭게 정리된 히스토리로 완전히 대체할 수 있습니다.
    이는 원격 브랜치의 히스토리를 깨끗하게 유지할 수 있는 강력한 방법입니다.

4. 새로운 Pull Request 생성

    new-release 브랜치에서 main 브랜치로 새로운 PR을 생성합니다.
    새로운 PR을 생성하면 GitHub는 새롭게 정리된 커밋 히스토리만을 비교 대상으로 삼게 됩니다.

    GitHub는 PR을 생성할 때 base 브랜치와 compare 브랜치 사이의 차이를 분석합니다.
    새로운 PR을 생성하면, GitHub는 새롭게 정리된 커밋 히스토리만을 비교하게 되어 원하는 변경사항만 표시됩니다.

왜 이 방법이 효과적인가:

1. 인터랙티브 리베이스의 동작 원리:
   - Git의 rebase 명령어는 커밋 히스토리를 재작성합니다.
   - 인터랙티브 모드(`-i` 옵션)를 사용하면 각 커밋을 개별적으로 수정, 삭제, 병합할 수 있습니다.
   - 이 과정에서 문제가 되는 Revert 커밋을 완전히 제거할 수 있습니다.
   - Git은 리베이스 과정에서 각 커밋을 순차적으로 적용하므로, 제거된 커밋의 영향이 완전히 사라집니다.

2. 새로운 커밋 생성의 의미:
   - 기존의 A와 B 기능을 다시 추가하는 새 커밋을 만들면, Git은 이를 *완전히 새로운 변경사항*으로 인식합니다.
   - 이렇게 하면 이전의 복잡한 Revert 히스토리와 관계없이 깔끔한 변경사항을 얻을 수 있습니다.

3. 강제 푸시의 효과:
   - `--force` 옵션은 원격 브랜치의 히스토리를 로컬의 새로운 히스토리로 완전히 대체합니다.
   - 이는 기존의 복잡한 커밋 히스토리를 새롭게 정리된 히스토리로 덮어쓰는 효과가 있습니다.

4. 새로운 PR의 의미:
   - 새로운 PR을 생성함으로써, GitHub는 새롭게 정리된 커밋 히스토리만을 비교 대상으로 삼게 됩니다.
   - 이전의 복잡한 Revert 히스토리는 이 새로운 PR에서 완전히 배제됩니다.
