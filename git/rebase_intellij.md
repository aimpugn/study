# Rebase via IntelliJ

- [Rebase via IntelliJ](#rebase-via-intellij)
    - [Example 1](#example-1)
        - [리베이스 전](#리베이스-전)
        - [inellij 인터페이스1](#inellij-인터페이스1)
        - [좌측에 Rebasing `61e01e86` from `some/feature/AAA` 항목을 Accept Left](#좌측에-rebasing-61e01e86-from-somefeatureaaa-항목을-accept-left)
        - [inellij 인터페이스2](#inellij-인터페이스2)
    - [Example 2](#example-2)
        - [리베이스 전](#리베이스-전-1)
        - [Start Rebase: Stop at `cbb6e7eb` commit](#start-rebase-stop-at-cbb6e7eb-commit)
        - [`fa15cefb9` 커밋으로 soft reset](#fa15cefb9-커밋으로-soft-reset)
        - [`cbb6e7eb2`에서 추가됐지만, 이제는 불필요해진 파일 두 개를 삭제 후 커밋](#cbb6e7eb2에서-추가됐지만-이제는-불필요해진-파일-두-개를-삭제-후-커밋)
        - [Continue rebase 시에 충돌 발생](#continue-rebase-시에-충돌-발생)
            - [\<-\> Commit: command.go](#--commit-commandgo)
            - ["Merge Revisions for /path/to/file/some\_api/pgutil/command.go"](#merge-revisions-for-pathtofilesome_apipgutilcommandgo)
            - [파일을 직접 더블 클릭하는 경우와 Resolve Conflict로 머지하는 경우](#파일을-직접-더블-클릭하는-경우와-resolve-conflict로-머지하는-경우)
            - [충돌 원인](#충돌-원인)
        - [충돌 해결(Accept Left) 후 Continue rebase 후 다시 충돌](#충돌-해결accept-left-후-continue-rebase-후-다시-충돌)
    - [CLI로 충돌 해결](#cli로-충돌-해결)
    - [용어](#용어)
        - [대상 브랜치(기준이 되는 브랜치)](#대상-브랜치기준이-되는-브랜치)
            - [리베이스의 컨텍스트](#리베이스의-컨텍스트)
            - [대상 브랜치의 역할](#대상-브랜치의-역할)
            - [리베이스 과정에서의 충돌 해결](#리베이스-과정에서의-충돌-해결)

## Example 1

### 리베이스 전

```bash
* 61e01e865 (HEAD -> some/feature/AAA) some_api/users: Various enhancements
* d31b2a74a some_api/utils/testing.go: Delete
* 5520ae198 (origin/some/feature/AAA) some_api: Add entry point and application
```

1. `d31b2a74a`: repository_test.go 파일의 수정이 있었고
2. `61e01e865`: repository_test.go -> user_test.go란 파일로 변경되어 있음

그리고 리베이스 시작.

```bash
* 61e01e865 (some/feature/AAA) some_api/users: Various enhancements
* d31b2a74a some_api/utils/testing.go: Delete
* 5520ae198 (HEAD, origin/some/feature/AAA) some_api: Add entry point and application
```

### inellij 인터페이스1

- 좌측: Rebasing `61e01e86` from s`ome/feature/AAA`
- 중간: Result
- 우측: Already rebased commits

### 좌측에 Rebasing `61e01e86` from `some/feature/AAA` 항목을 Accept Left

Continue rebasing

```bash
* e838b6067 (HEAD) some_api/users: Various enhancements
| * 61e01e865 (some/feature/AAA) some_api/users: Various enhancements
| * d31b2a74a some_api/utils/testing.go: Delete
|/
* 5520ae198 (origin/some/feature/AAA) some_api: Add entry point and application
```

### inellij 인터페이스2

- 좌측: Rebasing `d31b2a74` from `some/feature/AAA`
- 중간: Result
- 우측: Already rebased commits(아마도 앞서 리베이스 성공한 `61e01e86` 커밋으로 보임)

앞서 성공한 61e01e86 커밋이 당시 최신 커밋이었어서, 이번에는 우측의 Accept Right
그리고 실제로도 우측이 최종적으로 원하는 코드

## Example 2

### 리베이스 전

```bash
* e05b253b1 (HEAD -> some/feature/BBB) go mod tidy
* 81ffd9a80 some_api/main.go: Add main entry point
* 3af727da6 some_api/usecase/payments/legacy.go
* 62d2371c6 some_api/usecase/payments/command.go
* 7cf551cbd mocks: Modify `.mockery.yaml` and re-generate mocks
* d57916a35 some_api/api: Add api layer
* 102abbf3b some_api/bootstrap: Add bootstrap layer
* 3f21e1a14 some_api/usecase: Add usecase layer
* df20c30cb some_api/repository: Add repository layer
* 46ee1ac82 some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg
* cbb6e7eb2 some_api/commands: Add commands and legacy
* fa15cefb9 some_api/domain: Add domain layer
```

### Start Rebase: Stop at `cbb6e7eb` commit

```bash
* e05b253b1 (some/feature/BBB) go mod tidy
* 81ffd9a80 some_api/main.go: Add main entry point
* 3af727da6 some_api/usecase/payments/legacy.go
* 62d2371c6 some_api/usecase/payments/command.go
* 7cf551cbd mocks: Modify `.mockery.yaml` and re-generate mocks
* d57916a35 some_api/api: Add api layer
* 102abbf3b some_api/bootstrap: Add bootstrap layer
* 3f21e1a14 some_api/usecase: Add usecase layer
* df20c30cb some_api/repository: Add repository layer
* 46ee1ac82 some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg
* cbb6e7eb2 (HEAD) some_api/commands: Add commands and legacy
* fa15cefb9 some_api/domain: Add domain layer
```

### `fa15cefb9` 커밋으로 soft reset

```bash
git -c core.quotepath=false -c log.showSignature=false reset --soft fa15cefb977ca1dabdfc20e554c72b32e89f14ed
```

```bash
* e05b253b1 (some/feature/BBB) go mod tidy
* 81ffd9a80 some_api/main.go: Add main entry point
* 3af727da6 some_api/usecase/payments/legacy.go
* 62d2371c6 some_api/usecase/payments/command.go
* 7cf551cbd mocks: Modify `.mockery.yaml` and re-generate mocks
* d57916a35 some_api/api: Add api layer
* 102abbf3b some_api/bootstrap: Add bootstrap layer
* 3f21e1a14 some_api/usecase: Add usecase layer
* df20c30cb some_api/repository: Add repository layer
* 46ee1ac82 some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg
* cbb6e7eb2 some_api/commands: Add commands and legacy
* fa15cefb9 (HEAD) some_api/domain: Add domain layer
```

### `cbb6e7eb2`에서 추가됐지만, 이제는 불필요해진 파일 두 개를 삭제 후 커밋

1. interactive rebase로 `cbb6e7eb2`로 이동하고 stop
2. soft reset으로 `fa15cefb9`로 HEAD 이동
3. 원래 `cbb6e7eb2`에서 "추가"됐던 파일 삭제하고 수정된 내역 commit해서 `dc95036e4` 추가
    - some_api/command/pgutil.go 추가했지만 삭제
    - some_api/command/modified.go 수정된 파일 커밋
4. continue rebase
5. 충돌 발생해서 보니 쌩뚱맞게 some_api/pgutil/command.go에 대한 충돌 발생. 근데 원래 some_api/command/pgutil.go 이 파일을 some_api/pgutil/command.go 이 파일로 이동시킨 거긴 함. 근데 이거는 앞서 다른 squash commits 통해서 커밋들을 스쿼시 해둔 상태.

```bash
* dc95036e4 (HEAD) (Rebase) some_api/newly_committed_after_rebase
| * e05b253b1 (some/feature/BBB) go mod tidy
| * 81ffd9a80 some_api/main.go: Add main entry point
| * 3af727da6 some_api/usecase/payments/legacy.go
| * 62d2371c6 some_api/usecase/payments/command.go
| * 7cf551cbd mocks: Modify `.mockery.yaml` and re-generate mocks
| * d57916a35 some_api/api: Add api layer
| * 102abbf3b some_api/bootstrap: Add bootstrap layer
| * 3f21e1a14 some_api/usecase: Add usecase layer
| * df20c30cb some_api/repository: Add repository layer
| * 46ee1ac82 some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg
| * cbb6e7eb2 some_api/commands: Add commands and legacy
|/
* fa15cefb9 some_api/domain: Add domain layer
```

방금 커밋한 지점으로 `HEAD`가 이동한다.

### Continue rebase 시에 충돌 발생

충돌 해결 창: Rebasing branch `some/feature/BBB` onto `fa15cefb9`
- 파일명: some_api/pgutil/command.go
- Yours(`some/feature/BBB`): Modified
- Theirs(`fa15cefb9`): Deleted

#### <-> Commit: command.go

Commit 탭을 보면 Merge Conflicts가 나타났고, 여기서 `some_api/pgutil/command.go` 파일을 더블 클릭해보면 여기서도 삼분할로 나온다.

- 좌측: `293660b7bb1351d579f9baae939cb682b8fa673b`

    현재 작업중인 로컬 브랜치에서의 파일 상태를 보여준다.
    즉, 현재 `HEAD`, 즉 리베이스 과정 중에 stop하고 새로 커밋된 상태의 파일이다.
    작업자가 최신으로 커밋한 내용.

- 가운데: Base version

    충돌이 발생하기 전의 원본 파일 상태를 나타낸다.
    즉, 두 변경 사항(로컬과 서버)이 적용되기 전의 기본 파일 상태를 보여준다.

- 우측: Change from server

    리베이스하려는 대상 브랜치(서버 상의 상태)에서의 파일 상태를 보여준다.
    이 경우 `fa15cefb9` 커밋에서의 파일 상태일 것이다.
    여기서는 파일이 삭제된 상태로 표시될 수 있다.

- 좌측 창과 우측 창의 차이점 비교:

    로컬 변경사항과 서버 변경사항을 비교하여 어떤 부분에서 충돌이 발생했는지 명확히 파악할 수 있다.

- 가운데 창에서 수동 병합

    사용자는 가운데 창에서 직접 수정을 가하여 최종 파일을 만들 수 있다.
    여기서는 삭제된 파일에 대해 로컬에서 수정한 부분을 유지할지, 삭제를 수용할지 결정해야 한다.

IntelliJ 없이 CLI에서 이러한 충돌을 해결하려면 `git status`를 통해 충돌 상태를 확인한다.

```bash
... 중략 ...

Unmerged paths:
  (use "git restore --staged <file>..." to unstage)
  (use "git add/rm <file>..." as appropriate to mark resolution)
    deleted by us:   pgutil/command.go

... 중략 ...
```

그리고 직접 충돌이 발생한 파일을 열어 수정한다. 충돌된 부분은 보통 다음과 같이 표시된다:

```plaintext
<<<<<<< HEAD
(현재 브랜치의 내용)
=======
(리베이스하려는 대상 브랜치의 내용)
>>>>>>> some_commit_id
```

수정 후, 수정된 파일을 `git add`로 스테이징하고 `git rebase --continue`를 통해 리베이스 과정을 계속 진행할 수 있다. 만약 리베이스를 중단하고 싶다면 `git rebase --abort`를 사용하여 모든 변경사항을 원래대로 되돌릴 수 있다.

#### "Merge Revisions for /path/to/file/some_api/pgutil/command.go"

이러고 충돌 해소를 위해 merge를 눌러 보면 "Merge Revisions for /path/to/file/some_api/pgutil/command.go"라는 타이틀의 창이 뜬다.

- 좌측: Rebasing `46ee1ac8` from `some/feature/BBB` (Show details)

    현재 작업 중인 브랜치(`some/feature/BBB`)에서의 파일 상태를 보여준다.
    여기서는 파일이 수정되어 있습니다.

- 가운데: Result

    우선 처음에는 현재 `HEAD`에서의 파일 내용을 보여준다.
    이곳은 충돌 해결 후의 결과를 보여주며, 사용자가 직접 편집할 수 있다.

- 우측: Already rebased commits (Show details)

    리베이스 대상 브랜치(`fa15cefb9`)에서의 파일 상태를 보여준다.
    파일이 삭제된 상태이므로 비어 있습니다.

- "Rebasing 46ee1ac8 from some/feature/BBB" 우측에 "Show details"

    "some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg"라는 커밋 상세 정보가 보인다.
    이것은 사용자가 현재 충돌을 해결하고 있는 특정 커밋의 세부사항을 제공한다.
    이 커밋이 어떤 변경사항을 포함하고 있는지에 대한 힌트를 줍니다.

    리베이스 과정 중 충돌을 해결할 때, 각 커밋의 세부사항을 이해하는 것이 중요합니다. 왜냐하면:

    1. 코드 이해: 커밋 메시지를 통해 해당 커밋에서 무엇이 변경되었는지 이해할 수 있으므로, 충돌이 발생한 코드를 보다 정확하게 해석할 수 있습니다.
    2. 적절한 결정: 커밋에서 수행된 변경사항을 기반으로 충돌 해결 시 어떤 코드를 유지할지 또는 수정할지 결정할 수 있습니다. 예를 들어, 두 패키지의 기능이 하나로 통합된 경우, 두 부분의 코드 모두 중요할 수 있으므로 충돌 해결 시 이를 반영해야 할 수 있습니다.
    3. 충돌 해결 전략: 커밋의 세부사항을 알면 충돌 해결 방법을 더 잘 선택할 수 있습니다. 예를 들어, 파일이 삭제되었는데 로컬 변경에서 중요한 추가가 이루어진 경우, 이 파일을 삭제하지 않고 로컬 변경을 유지할지 결정할 수 있습니다.

#### 파일을 직접 더블 클릭하는 경우와 Resolve Conflict로 머지하는 경우

- 파일을 직접 더블 클릭

    리베이스 과정 중 발생한 전반적인 충돌 상황을 보여줍니다.
    사용자는 여기서 충돌의 원인과 세부적인 변화를 파악할 수 있습니다.

    - 좌측: `293660b7bb1351d579f9baae939cb682b8fa673b`

        리베이스 과정 중에 생성된 임시 커밋의 상태를 나타냅니다.
        여기서는 사용자가 리베이스를 중단하고 직접 작업한 결과를 반영한 커밋입니다.

    - 가운데: Base version

        이는 충돌 발생 전의 기본 버전, 즉 공통 조상을 나타냅니다.
        양쪽의 변경 사항과 비교하기 위한 기준점으로 사용됩니다.

    - 우측: Change from server

        리베이스 대상 브랜치에서의 변경 사항을 보여줍니다.
        즉, `some/feature/BBB` 브랜치를 `fa15cefb9` 커밋 위로 리베이스하려 할 때 해당 커밋(`fa15cefb9`)에서의 파일 상태입니다.

- Resolve Conflict로 머지하는 경우

    실제로 사용자가 해결해야 할 특정 충돌에 초점을 맞춥니다.
    여기서는 특정 커밋에서의 변경 사항과 리베이스하려는 대상의 상태 사이의 충돌을 직접 해결할 수 있도록 돕습니다.

    - 좌측: Rebasing `46ee1ac8` from `some/feature/BBB`

        충돌을 일으킨 특정 커밋에서의 파일 상태를 보여줍니다.
        이 커밋은 파일에 변화를 가한 커밋으로, 여기서는 `some_api/pgutil` 디렉토리 내의 변경을 나타냅니다.

        좌측에서 파일이 여전히 존재하고 변경사항이 포함된 것은, 작업자의 브랜치에서 이 파일이 여전히 중요하고 변경되었음을 의미합니다. 이것은 리베이스를 시도할 때 충돌의 원인이 됩니다.

    - 가운데: Result

        충돌 해결 후의 최종 결과를 나타냅니다.
        사용자가 이 창에서 변경 사항을 직접 편집하여 충돌을 해결할 수 있습니다.

    - 우측: Already rebased commits(`Theirs`)

        이미 리베이스가 완료된 커밋들의 상태를 나타냅니다.
        리베이스 대상 커밋(`fa15cefb9` 등)에서 삭제되었기 때문에 빈 파일로 나옵니다.
        리베이스 과정에서 여러분이 대상으로 설정한 커밋 또는 그 이후의 커밋에서 해당 파일이 삭제된 것일 수 있습니다.

        이는 여러분의 작업 브랜치(`some/feature/BBB`)와 대상 브랜치 사이에서 파일에 대한 상반된 변경사항이 존재함을 의미합니다.

이러한 충돌은 리베이스를 시도하는 동안 한 브랜치에서는 파일을 계속 사용하고 수정한 반면, 다른 브랜치에서는 같은 파일을 삭제했기 때문에 발생합니다. Git은 이런 상황에서 어떤 변경사항을 우선해야 할지 자동으로 결정할 수 없으므로 사용자에게 해결을 요청합니다.

#### 충돌 원인

이 충돌은 `46ee1ac8` 커밋과 리베이스 과정 중에 생성한 `293660b7b` 커밋 사이에서 발생한 것입니다.
이 두 커밋 간의 충돌은 다음과 같은 상황에서 발생했습니다:

1. **커밋 `46ee1ac8`:**

    이 커밋에서는 `some_api/pgutil` 디렉토리에 대한 중요한 변경이 이루어졌습니다.
    이 변경은 패키지 합병(`Merge X and Y pkg into Z pkg`)을 포함하고 있으며, 이 과정에서 `command.go` 파일 등에 대한 변경이 있었을 수 있습니다.

2. **커밋 `293660b7b`:**

    이는 리베이스 과정 중 여러분이 수행한 임시 커밋으로, 이 커밋에서 `command.go` 파일에 대한 변경 또는 작업이 있었습니다.
    이 커밋은 리베이스를 시도하면서 생성된 것으로, 여러분의 최신 작업이 반영된 상태입니다.

3. **충돌의 원인:**

    리베이스 *대상 브랜치(기준이 되는 브랜치)*에서 `command.go` 파일이 삭제되거나 변경되었을 때, 여러분의 브랜치에서 해당 파일에 대한 중요한 변경이 이루어진 경우입니다.

    *대상 브랜치(기준이 되는 브랜치)*는 `some/feature/BBB` 브랜치를 리베이스하려는 브랜치, 즉 리베이스 작업의 기준점으로 사용되는 브랜치를 말합니다. 이 경우, 커밋 `fa15cefb9` 이후의 브랜치 상태가 대상 브랜치에 해당됩니다.

    리베이스 과정은 여러분의 커밋(`293660b7b`)을 대상 브랜치의 최신 상태(`fa15cefb9` 이후)에 적용하려 할 때, 파일의 상태가 서로 충돌하여 이 문제가 발생합니다.

4. **`Already rebased commits` / `Theirs`:** 이 섹션에서 파일이 "Deleted"로 표시되는 것은, 리베이스하려는 대상 브랜치에서 해당 파일이 이미 삭제되었거나 변경된 상태를 반영하는 것입니다. 이는 `command.go` 파일에 대한 여러분의 변경사항과 충돌하게 됩니다.

- 충돌 해결 방법
    - **파일 유지:** 여러분이 변경한 `command.go`의 내용을 계속 유지하려면, "Accept Yours"를 선택하여 여러분의 변경사항을 적용합니다.
    - **파일 삭제 수락:** 대상 브랜치에서의 파일 삭제를 수락하려면, "Accept Theirs"를 선택하여 파일 삭제를 확정합니다.
- 해결 후 절차
    - 충돌을 해결한 후에는 `git add`를 통해 변경사항을 스테이징하고, `git rebase --continue`로 리베이스를 계속 진행하면 됩니다. 만약 리베이스를 중단하고 싶다면, `git rebase --abort`를 사용하여 모든 변경사항을 원래 상태로 되돌립니다.

### 충돌 해결(Accept Left) 후 Continue rebase 후 다시 충돌

```bash
* 7dd60c27a (HEAD) mocks: Modify `.mockery.yaml` and re-generate mocks
* 0384ced32 some_api/api: Add api layer
* 7456b4d34 some_api/bootstrap: Add bootstrap layer
* 4245d1cf1 some_api/usecase: Add usecase layer
* a4d658b13 some_api/repository: Add repository layer
* 7f28ed3a7 some_api/pgutil: Merge `code` and `receipt` pkg into `pgutil` pkg
* dc95036e4 (Rebase) some_api/newly_committed_after_rebase
| * e05b253b1 (some_api/feature/application-payments-layered-architecture) go mod tidy
| * 81ffd9a80 some_api/main.go: Add main entry point
| * 3af727da6 some_api/usecase/payments/legacy.go
| * 62d2371c6 some_api/usecase/payments/command.go
| * 7cf551cbd mocks: Modify `.mockery.yaml` and re-generate mocks
| * d57916a35 some_api/api: Add api layer
| * 102abbf3b some_api/bootstrap: Add bootstrap layer
| * 3f21e1a14 some_api/usecase: Add usecase layer
| * df20c30cb some_api/repository: Add repository layer
| * 46ee1ac82 some_api/pgutil: Merge `X` and `Y` pkg into `Z` pkg
| * cbb6e7eb2 some_api/commands: Add commands and legacy
|/
* fa15cefb9 some_api/domain: Add domain layer
```

- 좌측: Rebasing `62d2371c` from `some_api/feature/BBB`
- 가운데: Result
- 우측: Already rebased commits

```bash
git show 62d2371c | cat
commit 62d2371c6824981c88d864134e4ec4591e112b1c
Author: rody <rody@some.domain>
Date:   Mon Apr 15 15:47:35 2024 +0900

    some_api/usecase/payments/command.go

diff --git a/some_api/commands/payment_response_builder.go b/some_api/usecase/payments/command.go
similarity index 74%
rename from some_api/commands/payment_response_builder.go
rename to some_api/usecase/payments/command.go
index f76241d7b..29740606a 100644
--- a/some_api/commands/payment_response_builder.go
+++ b/some_api/usecase/payments/command.go
@@ -1,9 +1,10 @@
-package commands
+package payments

 import (
        "time"

        "github.com/private-org/go/some_api/domain"
+       "github.com/private-org/go/some_api/pgutil"
 )

 type PaymentResponseBuilder interface {
@@ -13,11 +14,11 @@ type PaymentResponseBuilder interface {
 }

 type PaymentResponseBuilderCommand interface {
-       PGUtil
-       GetBankInfo() *domain.Info
+       pgutil.Command
+       GetBankInfo() *pgutil.FinancialInfo
        GetCancelReceiptURL() (*string, error)
-       GetCardInfo(other ...string) *domain.Info
-       GetCardIssuerInfo() *domain.Info
+       GetCardInfo(other ...string) *pgutil.FinancialInfo
+       GetCardIssuerInfo() *pgutil.FinancialInfo
        GetCardPublishCode() string
        GetCardType() *int32
        GetCurrency() string
```

좌측

## CLI로 충돌 해결

IntelliJ 없이 이 충돌을 해결하려면 CLI를 사용해야 합니다. 기본적인 Git 명령어로 충돌을 해결할 수 있습니다:

1. **리베이스 시작:**

   ```bash
   git checkout some/feature/BBB
   git rebase fa15cefb9
   ```

2. **충돌 확인:**

   ```bash
   git status
   ```

   이 명령은 충돌 상태의 파일들을 보여줍니다.

3. **파일 직접 수정:**
   충돌이 발생한 파일을 열고 직접 수정합니다. 이 경우 `some_api/pgutil/command.go`가 삭제되었으므로, 해당 파일을 삭제하거나 필요한 변경을 적용해야 합니다.

4. **충돌 해결 마킹:**

   ```bash
   git add some_api/pgutil/command.go
   ```

   파일을 수정 후에는 `git add`를 사용하여 충돌 해결을 마킹합니다.

5. **리베이스 계속:**

   ```bash
   git rebase --continue
   ```

   나머지 리베이스 과정을 계속 진행합니다.

6. **문제가 계속되면 리베이스 중단:**

   ```bash
   git rebase --abort
   ```

   리베이스를 중단하고 원래 상태로 돌아갑니다.

위의 절차를 따라 CLI 환경에서도 IntelliJ의 편리한 그래픽 인터페이스 없이 충돌을 해결할 수 있습니다. 필요에 따라 적절한 텍스트 편집기를 사용하여 파일을 직접 수정하면 됩니다.

## 용어

### 대상 브랜치(기준이 되는 브랜치)

여기서 언급한 대상 브랜치(기준이 되는 브랜치)는 여러분이 `some/feature/BBB` 브랜치를 리베이스하려는 브랜치, 즉 리베이스 작업의 기준점으로 사용되는 브랜치를 말합니다. 이 경우, 커밋 `fa15cefb9` 이후의 브랜치 상태가 대상 브랜치에 해당됩니다.

#### 리베이스의 컨텍스트

리베이스는 하나의 브랜치를 다른 브랜치의 최신 커밋 위로 "재배치"하는 과정입니다. 이 과정에서 첫 번째 브랜치의 커밋들이 대상 브랜치의 커밋들 이후에 다시 적용되므로, 마치 이 커밋들이 처음부터 대상 브랜치에서 이루어진 것처럼 보이게 됩니다.

#### 대상 브랜치의 역할

- **업데이트된 기준점:** 리베이스는 대상 브랜치의 최신 상태를 기준으로 합니다. 이는 리베이스하는 브랜치의 커밋들이 대상 브랜치의 최신 커밋 후에 적용되어야 함을 의미합니다.
- **변경사항 반영:** 대상 브랜치에서 발생한 모든 변경사항(새로운 커밋, 수정, 삭제 등)은 리베이스 과정에서 고려되어야 합니다. 이 변경사항들이 리베이스 중 충돌의 원인이 될 수 있습니다.

#### 리베이스 과정에서의 충돌 해결

- **삭제된 파일의 처리:** 만약 대상 브랜치에서 어떤 파일이 삭제되었고, 리베이스하려는 브랜치에서 해당 파일이 여전히 존재하거나 수정되었다면, 이는 충돌을 일으킬 수 있습니다.
- **수정된 파일의 충돌:** 두 브랜치에서 동일한 파일에 대해 서로 다른 수정이 이루어졌다면, 이 또한 충돌의 원인이 됩니다.

결국 `fa15cefb9` 커밋 이후의 브랜치 상태는 리베이스의 기준점으로서, `some/feature/BBB` 브랜치에서 발생한 변경사항과 대조되어 충돌을 일으킬 수 있는 모든 요소를 포함하고 있습니다. 이를 이해하고 적절히 대처하는 것이 중요합니다.
