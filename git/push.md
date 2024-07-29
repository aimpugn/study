# push

- [push](#push)
    - [로컬 브랜치를 원격 저장소의 특정 브랜치로 푸시](#로컬-브랜치를-원격-저장소의-특정-브랜치로-푸시)

## 로컬 브랜치를 원격 저장소의 특정 브랜치로 푸시

```bash
git push -u origin feature/aaa
```

이 명령을 사용함으로써, 로컬의 `feature/aaa` 브랜치가 원격 저장소의 `feature/aaa` 브랜치와 연결되며, 이후에는 간단히 `git push` 또는 `git pull` 명령만으로 동기화를 진행할 수 있습니다.

- **git push**

    이 명령은 로컬 저장소의 변경사항을 원격 저장소로 전송합니다.
    이를 통해 다른 사람들이 최신 코드에 접근할 수 있게 됩니다.

- **-u 옵션**

    이 옵션은 `--set-upstream`의 축약형입니다.
    이 옵션을 사용하면, 현재 브랜치의 원격 추적 브랜치를 설정합니다.
    즉, 이후에 `git push`나 `git pull` 명령을 입력할 때 브랜치 이름을 명시하지 않아도 자동으로 이 브랜치와 동기화를 시도합니다.

- **origin**:

    원격 저장소의 이름입니다.
    `origin`은 클론(clone)을 생성할 때 기본적으로 설정되는 원격 저장소의 기본 이름입니다.

- **feature/aaa**

    이는 원격 저장소에 푸시하려는 브랜치의 이름입니다.
    일반적으로 `feature/` 접두사는 새로운 기능 개발을 위한 브랜치임을 나타냅니다.
    이 예에서는 `feature/aaa` 브랜치에 코드를 푸시하고 있습니다.

Intellij에서 로컬에서 새로 생성된 브랜치 push할 때 사용하는 명령어

```bash
git -c log.showSignature=false \
    push --progress --porcelain origin \
    refs/heads/coreapi-rm-unnecessary-types:refs/heads/coreapi-rm-unnecessary-types \
    --set-upstream \
    --follow-tags
```
