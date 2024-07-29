# git reset

- [git reset](#git-reset)
    - [`reset`?](#reset)
    - [undo commit squash](#undo-commit-squash)
    - [undo](#undo)
    - [`--hard` 옵션](#--hard-옵션)
        - [git reflog 통해 과거로 돌아가기](#git-reflog-통해-과거로-돌아가기)
        - [git reset --hard origin/HEAD](#git-reset---hard-originhead)
        - [원격 브랜치의 내용으로 덮어쓰기](#원격-브랜치의-내용으로-덮어쓰기)

## `reset`?

`git reset` 명령어는 Git에서 매우 강력한 도구 중 하나로, 로컬 저장소의 현재 브랜치의 HEAD를 다른 상태로 이동시키거나, 인덱스(스테이지 영역)와 워킹 디렉토리의 상태를 변경하는 데 사용됩니다. 이 명령어는 커밋을 취소하거나 변경사항을 되돌리는 데 자주 사용되며, 다양한 옵션을 통해 다른 수준의 리셋을 할 수 있습니다.

우선 Git의 핵심 구성 요소를 이해하려면 다음 세 가지 주요 영역을 알아야 합니다:

- **워킹 디렉토리(Working Directory)**:

   사용자의 작업 영역으로, 파일과 디렉토리가 실제로 수정되는 곳입니다.
   이곳에서 파일을 변경하고 일상적인 작업을 수행합니다.

- **인덱스(Index) 또는 스테이징 영역(Staging Area)**:

   커밋하기 전에 변경사항을 준비하는 영역입니다.
   `git add` 명령어를 사용해 워킹 디렉토리에서 변경된 파일들을 스테이징 영역으로 이동시킵니다.

- **HEAD**:

   현재 체크아웃된 커밋을 가리키는 포인터입니다.
   기본적으로 이는 로컬 브랜치의 최신 커밋을 참조합니다.

`git reset` 명령어는 세 가지 주요 영역 중 하나 또는 복수의 영역에 영향을 미칩니다.
이 명령어는 주로 다음 세 가지 옵션과 함께 사용됩니다:

- `git reset --soft <commit>`:

   HEAD는 지정된 커밋으로 이동하지만 인덱스(스테이지)와 워킹 디렉토리는 변경되지 않습니다.
   이 옵션은 커밋을 되돌리되, 변경된 파일들을 다시 커밋할 준비 상태로 두고 싶을 때 유용합니다.

- `git reset <commit>`, `git reset --mixed <commit>`(기본 옵션):

   HEAD와 인덱스가 지정된 커밋으로 재설정되지만, 워킹 디렉토리는 영향을 받지 않습니다.
   이는 커밋을 취소하고 해당 커밋의 변경사항을 다시 스테이징하지 않은 상태로 두고 싶을 때 사용합니다.

- **`--hard`**:

   HEAD, 인덱스, 워킹 디렉토리가 모두 지정된 커밋으로 재설정됩니다.
   이는 모든 로컬 변경사항을 버리고 특정 상태로 완전히 되돌리고자 할 때 사용됩니다.

리셋 실행 시 내부적으로는 다음과 같은 일이 이뤄집니다.

- **HEAD 이동**

    `git reset`은 먼저 HEAD를 지정된 커밋으로 이동합니다.

- **인덱스 조정**

    `--mixed`와 `--hard` 옵션을 사용하면 인덱스가 HEAD가 가리키는 커밋의 상태로 업데이트됩니다.

- **워킹 디렉토리 변경**

    `--hard` 옵션을 사용할 경우, 워킹 디렉토리의 파일들도 HEAD가 가리키는 커밋의 상태로 되돌려집니다.
    이는 모든 로컬 변경사항을 제거합니다.

## undo commit squash

```bash
git -c core.quotepath=false \
    -c log.showSignature=false \
    reset \
    --keep \
    30e273a289a96684745c3bce85490f0d564724cc
```

## undo

```bash
git reset --hard HEAD@{1}
```

```log
14387ef5 (HEAD -> some-qwerty-api-service/feature/application) HEAD@{0}: reset: moving to 14387ef5703a813952e707a5e3e3be05ec7f1799
c404325d HEAD@{1}: rebase (finish): returning to refs/heads/some-qwerty-api-service/feature/application
c404325d HEAD@{2}: rebase (pick): some-qwerty-api-service/payments/receipt: `kcp`, `uplus` 경우 쿼리 스트링 빌드시 값이 없어도 키는 유지하도록 수정
ae536b48 HEAD@{3}: rebase (pick): some-qwerty-api-service/payments: card issuer/publish 기관 코드를 command 통해서 사용하도록 수정
90aff2bf HEAD@{4}: rebase (pick): .github/workflows: allow github actions to access private repo
a72c8fcc HEAD@{5}: rebase (pick): some-qwerty-api-service: `go generate` 주석 대신 `.mockery.yaml` 적용
0ad8086e HEAD@{6}: rebase (pick): `*_service_interface/generated` -> `go/interface/*`
```

## `--hard` 옵션

`--hard` 옵션은 현재 작업 디렉토리의 모든 파일을 재설정 대상 커밋의 상태로 강제로 되돌립니다. 이는 작업 디렉토리와 인덱스(스테이징 영역) 모두에서 변경사항을 삭제합니다.

### [git reflog](./reflog.md) 통해 과거로 돌아가기

```bash
# `reflog`로 어떤 행동들이 있었는지 확인하기
git reflog
```

```bash
# 리베이스 직전 시점 위치가 `HEAD@{112}`라서, 그때로 되돌아 간다
g reset --hard HEAD@{112}
```

### git reset --hard origin/HEAD

명령어 `git reset --hard origin/HEAD`는 Git에서 사용하는 매우 강력한 명령어로, 로컬 저장소의 현재 브랜치를 원격 저장소의 현재 브랜치 상태(`origin/HEAD`)로 강제로 재설정합니다.

`origin/HEAD`는 원격 저장소 `origin`의 기본 브랜치(대개 `master` 또는 `main`)를 가리킵니다.
즉, `origin/HEAD`는 원격 저장소의 최신 커밋을 참조합니다.

명령어 실행 과정은 다음과 같습니다:

- **HEAD 이동**:

    로컬 저장소의 HEAD(현재 브랜치를 가리키는 포인터)가 `origin/HEAD`가 가리키는 커밋으로 이동합니다.
    이는 로컬 브랜치가 원격 브랜치의 최신 상태를 반영하게 됩니다.

- **인덱스 재설정**

    인덱스(스테이징 영역)가 `origin/HEAD`의 커밋 상태로 재설정됩니다.
    즉, 스테이지된 변경사항이 모두 제거되고 `origin/HEAD` 커밋의 상태가 스테이지됩니다.

- **작업 디렉토리 재설정**

    작업 디렉토리의 파일들이 `origin/HEAD` 커밋의 상태로 되돌려집니다.
    로컬에서 변경된 파일, 새로 생성된 파일, 삭제된 파일 등의 모든 변경사항이 이전 상태로 복구됩니다.

이 명령어는 로컬에서 작업한 내용이 모두 삭제되고 원격 저장소의 상태로 되돌려집니다.
따라서 작업 중인 중요한 변경사항이 있다면 이 명령어 사용 전에 반드시 백업을 해야 합니다.

만약 로컬에서의 중요한 커밋을 잃어버리고 싶지 않다면, 먼저 다른 브랜치에 백업을 하거나 `git stash`를 사용해 임시로 변경사항을 저장할 수 있습니다.

### 원격 브랜치의 내용으로 덮어쓰기

```bash

```
