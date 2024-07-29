# Commit

- [Commit](#commit)
    - [Git Commit?](#git-commit)
    - [commit 단위 나누기](#commit-단위-나누기)
    - [commit 관리 전략](#commit-관리-전략)
    - [`git commit --verbose --amend`, `gc!`](#git-commit---verbose---amend-gc)
    - [`git commit --allow-empty`](#git-commit---allow-empty)
    - [`git commit --amend`](#git-commit---amend)

## [Git Commit](https://git-scm.com/docs/git-commit)?

## commit 단위 나누기

1. 파일 또는 모듈 단위로 작게 나눈다
2. 실제 구현체를 먼저 개발하고 커밋한다
3. 구현체를 사용하도록 적용하고 커밋한다

## commit 관리 전략

## `git commit --verbose --amend`, `gc!`

- 명령어 `git commit --verbose --amend`는 Git에서 현재 브랜치의 마지막 커밋을 수정하는 데 사용됩니다.

    1. `--amend`: 이 옵션은 가장 최근에 완료된 커밋을 수정합니다. 만약 파일을 스테이지에 추가하거나, 커밋 메시지를 변경하고 싶을 때 사용할 수 있습니다. 기존의 커밋을 대체하는 새로운 커밋이 생성됩니다.

    2. `--verbose` (또는 `-v`): 이 옵션은 커밋 메시지를 작성할 때 `diff`를 포함하여 보여줍니다. 이는 커밋에 포함되는 변경 사항을 자세히 확인하고 싶을 때 유용합니다. 커밋 메시지 편집기가 열리면 변경된 내용의 `diff`가 보여지며, 사용자는 이를 참고하여 보다 상세한 커밋 메시지를 작성할 수 있습니다.

- 사용 예시

    - 파일을 수정하거나 새 파일을 추가한 후 `git add`로 변경사항을 스테이지에 올린다면, `git commit --amend`를 사용하여 이 변경사항을 마지막 커밋에 포함시킬 수 있습니다.

    - 마지막 커밋의 메시지만 수정하고 싶다면, `git commit --amend`를 실행한 후 커밋 메시지 편집기에서 메시지를 변경할 수 있습니다.

`--amend` 옵션을 사용하면 기존의 커밋을 새로운 커밋으로 대체합니다.
이는 공개 레포지토리에 이미 푸시된 커밋을 수정할 때 주의해야 합니다.
다른 사람이 해당 커밋을 기반으로 작업을 진행했다면, 커밋 히스토리가 변경되어 혼란을 야기할 수 있습니다.
이 경우, `git push --force` 옵션을 사용하여 변경된 커밋을 강제로 푸시해야 할 수도 있습니다.

## `git commit --allow-empty`

```log
some-qwerty-api-service/payments/code/command.go

#
# It looks like you may be committing a cherry-pick.
# If this is not correct, please run
#    git update-ref -d CHERRY_PICK_HEAD
# and try again.


# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
#
# Author:    rody <rody@RODY-EX-MAC.(none)>
# Date:      Thu Dec 21 12:08:55 2023 +0900
#
# interactive rebase in progress; onto c2045b11
# Last commands done (2 commands done):
#    edit a29f3d9d some-qwerty-api-service/payments/receipt/command.go: 불필요 속성 삭제
#    pick c9fa7546 some-qwerty-api-service/payments/code/command.go
# Next commands to do (71 remaining commands):
#    pick 7340cd04 some-qwerty-api-service/payments/sort.go: Refactor `PaymentV1Response` -> `PaymentResponseElement`
#    pick 7d4db4f6 some-qwerty-api-service/payments/sqls.go: Add `V2PgList` and refactor `ImpUIDs` from string to []string
# You are currently rebasing branch 'some-qwerty-api-service/feature/application' on 'c2045b11'.
#
# Changes not staged for commit:
#    modified:   ../interfaces/channel-service/channel-service-interface (new commits)
#    modified:   ../interfaces/merchant-service/merchant-service-interface (new commits)
#
# Untracked files:
#    .env.local
#    .env.prod
#    .env.prod.bak
#    .env.stg
#    payments/receipt/mock_ReceiptCommand.go
#    ../some-qwerty-pay-service/.env.local
#    ../some-qwerty-pay-service/tmp/
#    ../result.json
#
```

## `git commit --amend`

마지막 커밋에서 파일을 잊으셨나요? 커밋 후 바로 오타를 발견하셨나요? 먼저 업데이트된 버전의 파일을 스테이지에 추가합니다.
그런 다음 `--amend` 플래그를 사용하여 스테이징 된 변경 사항을 커밋합니다.
그러면 이전 변경 사항과 새 변경 사항이 모두 포함된 새 커밋으로 이전 커밋을 덮어쓰게 됩니다.

```bash
git add <file>

git commit --amend
```
