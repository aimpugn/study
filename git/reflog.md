# reflog

- [reflog](#reflog)
    - [reflog?](#reflog-1)
    - [하위 명령](#하위-명령)
        - [`show`](#show)
        - [`expire`](#expire)
        - [`delete`](#delete)
        - [`exists`](#exists)
    - [reflog 포맷](#reflog-포맷)
    - [Examples](#examples)
        - [리베이스 완료 후 ORIG\_HEAD로 돌아가기](#리베이스-완료-후-orig_head로-돌아가기)

## reflog?

- `reflog`는 특정 레퍼런스(주로 `HEAD`)의 이력을 기록
- 참조 로그(Reference log) 또는 `reflogs`는 분기 팁 및 기타 참조가 로컬 저장소에서 업데이트된 시기를 기록한다.
- 참조 로그는 다양한 Git 명령에서 참조의 이전 값을 지정하는 데 유용하다. 예를 들어,
    - `HEAD@{2}`: HEAD가 두 번 이동하기 전에 있었던 위치
    - `master@{one.week.ago}`: 이 로컬 저장소에서 1주일 전에 마스터가 가리켰던 위치
- 자세한 내용은 gitrevisions(7)을 참조

## 하위 명령

### `show`

하위 명령이 없는 경우 기본값인 `show`는 명령줄(또는 기본적으로 `HEAD`)에 제공된 참조 로그를 표시한다.
`reflog`는 모든 최근 작업을 다루고 `HEAD` reflog는 분기 전환(switching)을 기록한다.
`git reflog show`는 `git log -g --abbrev-commit --pretty=oneline`의 별칭이다.(자세한 내용은 git-log(1)를 참조)

### `expire`

`expire` 하위 명령은 이전 `reflog` 항목을 정리한다.
1. 만료 시간보다 오래된 항목들
2. 현재 팁(tip)에서 도달할 수 없고, '만료-도달할 수 없는(expire-unreachable)' 시간보다 오래된 항목들

> **'만료-도달할 수 없는(expire-unreachable)' 시간?**
>
> Git에서 설정할 수 있는 특정한 시간 값으로, `reflog` 항목이 얼마나 오래 저장될 것인지를 결정한다.
> 특히 "도달할 수 없는(unreachable)" 상태의 커밋들에 적용된다.
>
> **도달할 수 없는 커밋(unreachable commit)?**
>
> 현재 브랜치나 태그 등의 팁(tip)에서 어떤 경로를 통해서도 도달할 수 없는 커밋을 의미

이는 일반적으로 최종 사용자가 직접 사용하지 않는다.(git-gc(1)를 참조)

### `delete`

`reflog`에서 단일 항목을 삭제한다.
해당 인수는 정확한 항목이어야 한다(예: `git reflog delete master@{2}`)
또한 이 명령은 일반적으로 최종 사용자가 직접 사용하지 않는다

### `exists`

`ref`에 `reflog`가 있는지 확인한다.
`reflog`가 존재하면 0 상태로 종료되고, 존재하지 않으면 0이 아닌 상태로 종료된다.

## reflog 포맷

```log
`<커밋 해시>` `HEAD@{숫자}` `<명령어>` `<작업 로그>`
```

- `<커밋 해시>`: 해당 로그 항목에서 `HEAD`가 가리키고 있던 커밋의 해시
- `HEAD@{숫자}`: 해당 로그 항목이 기록된 시간 순서를 나타내는 인덱스
- `<명령어>`: 해당 로그 항목이 기록된 Git 명령어 또는 작업
- `<작업 로그>`: 명령어 또는 작업에 대한 추가적인 정보나 설명

```log
cb24b913a (HEAD -> some-qwerty-api-service/feature/application-arranged, tag: some-qwerty-api-service/v1.0.5, origin/some-qwerty-api-service/feature/application-arranged) HEAD@{0}: reset: moving to HEAD@{112}
a5db85155 HEAD@{1}: reset: moving to HEAD@{1}
a5db85155 HEAD@{2}: rebase (finish): returning to refs/heads/some-qwerty-api-service/feature/application-arranged
a5db85155 HEAD@{3}: rebase (continue): some-qwerty-api-service/payments: Add `SearchAfter` 적용
```

## Examples

### 리베이스 완료 후 ORIG_HEAD로 돌아가기

`git reflog`는 Git에서 현재 저장소의 헤드(HEAD)가 가리키던 참조들의 변화 기록을 보여준다.
이 명령어는 여러분이 수행한 Git 작업의 이력을 확인할 수 있게 해주며, 실수로 커밋을 잃어버렸을 때 이를 복구하는 데 유용하다.

```bash
# 가장 최근의 작업이다. `git reset` 명령어를 사용하여 `HEAD`를 `ORIG_HEAD`로 이동시켰다. 
# 이는 보통 병합이나 리베이스 후에 변경사항을 되돌리고자 할 때 사용된다.
61e01e865 (HEAD -> feature/awesome) HEAD@{0}: reset: moving to ORIG_HEAD

# 리베이스 작업을 마쳤으며, `refs/heads/feature/awesome` 브랜치로 돌아왔다.
21f424ec5 HEAD@{1}: rebase (finish): returning to refs/heads/feature/awesome 

# 리베이스 과정 중에 `coreapi/utils/testing.go` 파일을 삭제하는 커밋을 리베이스 했다.
21f424ec5 HEAD@{2}: rebase (continue): coreapi/utils/testing.go: Delete

# `coreapi/users`에 다양한 개선사항을 추가하는 커밋을 리베이스 했다.
d4060af4d HEAD@{3}: rebase (continue): coreapi/users: Various enhancements

# 리베이스를 시작하기 위해 `5520ae19803c247344455b2825bf4df6b4e94db8` 커밋으로 체크아웃 했다.
# 이것은 리베이스를 시작할 기점을 설정한다.
5520ae198 (origin/feature/awesome) HEAD@{4}: rebase (start): checkout 5520ae19803c247344455b2825bf4df6b4e94db8

# 리베이스를 중단하고(`abort`), `feature/awesome` 브랜치로 돌아왔다. 
# 이는 리베이스 도중 충돌이 발생했거나, 다른 이유로 리베이스를 취소하기로 결정했을 때 발생할 수 있다.
61e01e865 (HEAD -> feature/awesome) HEAD@{5}: rebase (abort): returning to refs/heads/feature/awesome
```
