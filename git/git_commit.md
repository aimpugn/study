# Git Commit

## git commit reference

`HEAD` 기준으로 특정 연산자를 사용하여 커밋을 참조할 수 있습니다.

- `HEAD`: 현재 체크아웃된 커밋을 가리키는 포인터
- `~`(틸드): 조상 커밋을 참조하는 연산자입니다.

    가령 `HEAD~2`는 현재 커밋의 2단계 이전 커밋을 가리킵니다.

    그리고 `HEAD~3..HEAD`와 같이 커밋 범위를 지정할 수 있습니다.

- `^`(캐럿): 부모 커밋을 참조하는 연산자입니다.

    `HEAD^`는 현재 커밋의 직전 커밋을 가리킵니다.

- `@{}`: reflog 항목을 참조하는 문법입니다.

### `HEAD~n`의 동작

`HEAD~n`은 현재 커밋에서 n번째 조상 커밋을 가리킵니다.
여기서 주의할 점은 `~`는 첫 번째 부모를 따라 거슬러 올라간다는 것입니다.

- `HEAD~0`: 현재 커밋
- `HEAD~1`: 현재 커밋의 직전 커밋(== `HEAD^`)
- `HEAD~2`: 현재 커밋의 2단계 이전 커밋
- `HEAD~n`: 현재 커밋에서 n번째 이전 커밋 (부모의 부모의 부모...)

```sh
❯ gloga
* eb9e2cb (HEAD -> main) squash
* b108996 squash
* 6d46fb5 (origin/main, origin/HEAD) 공부한 내용 가다듬고 정리
* 298ccca Update README.md
* 86f9d5a Initial commit

❯ git rev-parse HEAD
eb9e2cb

❯ git rev-parse HEAD~0
eb9e2cb

# tilde
❯ git rev-parse HEAD~1
b108996

# caret
❯ git rev-parse HEAD^1
b108996

# tilde
❯ git rev-parse HEAD~2
6d46fb5

# caret
❯ git rev-parse HEAD^2
HEAD^2
fatal: ambiguous argument 'HEAD^2': unknown revision or path not in the working tree.
```

### `HEAD^`의 동작

`HEAD^`는 `HEAD~1`과 동일하게 현재 커밋의 직전 커밋을 가리킵니다.
하지만 `^`는 머지 커밋의 경우 다르게 동작할 수 있습니다:

- `HEAD^1`: 첫 번째 부모 커밋(== `HEAD^`)

    ```sh
    [ "$(git rev-parse HEAD^)" = "$(git rev-parse HEAD^1)" ] && echo "true" || echo "false"
    true
    ```

- `HEAD^2`: 두 번째 부모 커밋 (머지 커밋의 경우)
- `HEAD^n`: 현재 커밋의 n번째 부모 (병합 커밋에서 유용)

### `HEAD@{숫자}`의 동작

`HEAD@{숫자}`는 리파지터리 작업 기록(reflog)을 기반으로 한 참조 방식입니다.
Reflog는 로컬 리파지터리에서 `HEAD`의 변경 이력을 기록합니다.

- `HEAD@{0}`: 현재 HEAD 위치
- `HEAD@{1}`: 직전에 HEAD가 가리키던 위치
- `HEAD@{2}`: 그 이전에 HEAD가 가리키던 위치

이 방식은 `~`나 `^`와는 달리, 브랜치 전환이나 리셋 등의 작업 내역을 포함한 시간 순서에 따른 참조를 제공합니다.

`master@{1}`과 같이 특정 브랜치의 이전 위치를 참조하거나,
`HEAD@{1.day.ago}`와 같이 상대적인 날짜로 커밋을 참조할 수 있습니다.

## Practical Implementation

다음은 각 참조 방식의 실제 사용 예시입니다:

```bash
# 현재 커밋의 2단계 이전 커밋 보기
git show HEAD~2

# 현재 커밋의 직전 커밋 보기
git show HEAD^

# 머지 커밋의 경우 두 번째 부모 커밋 보기
git show HEAD^2

# reflog 기반으로 두 번째 이전 HEAD 위치 보기
git show HEAD@{2}
```

## Comparative Analysis

| 참조 방식 | 의미                             | 특징                          |
| --------- | -------------------------------- | ----------------------------- |
| HEAD~n    | n번째 조상 커밋                  | 항상 첫 번째 부모를 따라 이동 |
| HEAD^     | 직전 커밋                        | 머지 커밋에서 부모 선택 가능  |
| HEAD@{n}  | reflog 기반 n번째 이전 HEAD 위치 | 작업 내역 기반, 시간순 참조   |

## Revert Commit

Revert 커밋은 이전 커밋의 변경사항을 정확히 반대로 적용합니다.
그러나 Git의 관점에서 이는 새로운 변경사항입니다.
따라서 Revert의 Revert는 원래 커밋과 동일한 효과를 가지지만, Git은 이를 다른 커밋으로 취급합니다.

## 기타

1. Pro Git Book: [7.1 Git Tools - Revision Selection](https://git-scm.com/book/en/v2/Git-Tools-Revision-Selection)
2. Git 공식 문서: [gitrevisions Documentation](https://git-scm.com/docs/gitrevisions)
