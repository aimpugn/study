# git show-ref

## [show-ref](https://git-scm.com/docs/git-show-ref)?

- 특정 커밋 해시에 대한 참조(refs)를 보는 명령어
- Git 저장소의 모든 참조(refs), 예를 들어 브랜치, 태그, 리모트 트래킹 브랜치 등과 그들의 커밋 해시를 나열한다
- List references in a local repository

## options

- `--head`: Show the HEAD reference, even if it would normally be filtered out.
- `--heads`: Limit to "refs/heads". Not mutually exclusive with `--tags`.
- `--tags`: Limit to "refs/tags". Not mutually exclusive with `--heads`.

## 사용례

```sh
git show-ref --heads --tags
```

이 명령어는 모든 로컬 브랜치와 태그에 대한 참조를 보여줍니다. 만약 특정 커밋 해시에 대한 참조만 보고 싶다면, 해당 커밋 해시를 명령어에 추가합니다:

```sh
git show-ref | grep [커밋해시]
```

여기서 `[커밋해시]`는 찾고자 하는 커밋의 해시입니다. 이 명령어는 해당 커밋 해시와 관련된 모든 참조를 출력합니다.
