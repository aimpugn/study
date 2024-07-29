# git fetch

- [git fetch](#git-fetch)
    - [git fetch?](#git-fetch-1)
    - [`--all`: 모든 remote](#--all-모든-remote)
    - [에러 통해 배우기](#에러-통해-배우기)
        - [! \[rejected\] develop -\> develop (non-fast-forward)](#-rejected-develop---develop-non-fast-forward)
            - [명령어 분석](#명령어-분석)
            - [결과 분석](#결과-분석)
            - [해결 방법](#해결-방법)

## git fetch?

- 리파지토리에서 object와 ref를 다운로드

## `--all`: 모든 remote

```shell
git fetch --all --tags

git -c credential.helper= -c core.quotepath=false -c log.showSignature=false fetch origin --recurse-submodules=no --progress --prune
```

## 에러 통해 배우기

### ! [rejected] develop -> develop (non-fast-forward)

```shell
git -c core.quotepath=false -c log.showSignature=false fetch origin develop:develop --recurse-submodules=no --progress --prune

remote: Total 24 (delta 14), reused 24 (delta 14), pack-reused 0 From https://github.com/some-org/some-A-service ! [rejected] develop -> develop (non-fast-forward) + 6e16e8b1a...5cf47e351 develop -> origin/develop (forced update)
```

```shell
12:35:14.931: [some-A-service] git -c core.quotepath=false -c log.showSignature=false fetch origin develop:develop --recurse-submodules=no --progress --prune
From https://github.com/some-org/some-A-service
 ! [rejected]            develop    -> develop  (non-fast-forward)
```

#### 명령어 분석

- `git -c core.quotepath=false -c log.showSignature=false fetch origin develop:develop --recurse-submodules=no --progress --prune`
    - `-c core.quotepath=false`: git 설정 옵션을 임시로 변경. 파일 이름에 특수 문자가 있는 경우 이스케이프 처리를 비활성화
    - `-c log.showSignature=false`: git 설정 옵션을 임시로 변경. 커밋 서명을 표시하지 않도록 한다
    - `fetch`: 원격 저장소에서 로컬 저장소로 브랜치나 태그를 가져옵니다.
    - `origin`: 원격 저장소의 이름입니다.
    - `develop:develop`: 원격 저장소의 develop 브랜치를 로컬 저장소의 develop 브랜치로 가져옵니다.
    - `--recurse-submodules=no`: 서브모듈을 재귀적으로 가져오지 않습니다.
    - `--progress`: 진행 상황을 표시합니다.
    - `--prune`: 원격 저장소에서 삭제된 브랜치를 로컬 저장소에서도 삭제합니다.
- `git fetch origin develop:develop`
    - 이 명령어는 단순히 원격 변경 사항을 추적하고 로컬 브랜치를 업데이트하지 않습니다.
    - `develop:develop` 구문이 사용된 것은 원격 브랜치의 변경 사항을 로컬 브랜치에 직접 반영하려는 의도를 보여줍니다.
    - 하지만, 이 명령어는 보통 `git fetch origin develop`와 같은 방식으로 사용됩니다.

#### 결과 분석

```log
remote: Total 24 (delta 14), reused 24 (delta 14), pack-reused 0 From https://github.com/some-org/some-A-service ! [rejected] develop -> develop (non-fast-forward) + 6e16e8b1a...5cf47e351 develop -> origin/develop (forced update)
```

- `Total 24 (delta 14)`
    - `Total 24`: 원격 저장소에서 가져온 객체의 총 수를 나타냅니다. 객체는 커밋, 트리, 블롭(blob), 태그 등을 포함
    - `(delta 14)`:
        - Git은 델타 압축(delta compression)이라는 기술을 사용하여 데이터를 효율적으로 저장하고 전송합니다.
        - 델타는 이전 객체와의 차이를 나타냅니다.
        - 여기서 delta 14는 14개의 델타 객체가 있음을 나타냅니다.
- `reused 24 (delta 14)`
    - 이는 이미 로컬에 존재하며 재사용된 객체의 수를 나타냅니다.
    - 여기서도 델타 객체가 포함되어 있습니다.
- `pack-reused 0`
    - Git은 객체를 팩(pack)이라는 형식으로 그룹화하여 저장하고 전송합니다.
    - `pack-reused`는 이미 존재하는 팩에서 재사용된 객체의 수를 나타냅니다.
    - 0은 재사용된 팩이 없음을 나타냅니다.
- `develop -> develop`
    - `->` 화살표는 소스 브랜치에서 대상 브랜치로의 맵핑을 나타냅니다.
    - 여기서는 원격의 develop 브랜치를 로컬의 develop 브랜치로 맵핑하려는 시도를 나타냅니다.
    - Git 출력에서 화살표의 방향은 **소스에서 대상으로의 맵핑**을 표시합니다. 예를 들어, `origin/master -> master`는 `origin/master` 브랜치를 로컬 `master` 브랜치로 맵핑하려는 시도를 나타냅니다.
- `! [rejected] develop -> develop (non-fast-forward)`
    - `로컬 develop` 브랜치가 `원격 develop` 브랜치보다 **앞서 있어**, 빠르게 앞으로 이동(fast-forward)할 수 없음을 나타냅니다
    - 일반적으로 로컬 브랜치와 원격 브랜치 사이에 충돌이 있음을 나타냅니다.
    - 로컬 브랜치에는 원격 브랜치에 없는 커밋이 있고, 원격 브랜치에는 로컬 브랜치에 없는 커밋이 있을 때 이런 현상이 발생합니다
- `+ 6e16e8b1a...5cf47e351 develop -> origin/develop (forced update)`
    - `+`: 원격 develop 브랜치가 강제로 업데이트되었음을 나타냅니다
    - `...`: 이 기호는 범위를 나타냅니다.
    - `develop -> origin/develop`: 로컬 `develop` 브랜치에서 원격 `origin/develop` 브랜치로의 맵핑을 나타냅니다. 그러나 이 경우에는 `로컬 develop` 브랜치가 아닌 `원격 develop` 브랜치에 대한 정보를 제공합니다.

#### 해결 방법

- 로컬 브랜치를 재설정하거나 삭제하고 원격 브랜치를 다시 가져오는 것이 한 가지 해결 방법일 수 있습니다. 예를 들어, 로컬 develop 브랜치를 원격 develop 브랜치로 재설정하려면 다음 명령어를 사용할 수 있습니다

```shell
# 이 명령어는 로컬 develop 브랜치의 HEAD를 원격 develop 브랜치의 최신 커밋으로 이동시킵니다.
git reset --hard origin/develop
```

- 또는 로컬 develop 브랜치를 삭제하고 원격 develop 브랜치를 다시 체크아웃

```shell
# `-D`: force delete
git branch -D develop && git checkout develop
```
