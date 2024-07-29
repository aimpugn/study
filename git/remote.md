# git remote

- [git remote](#git-remote)
    - [`원격 추적 브랜치`?](#원격-추적-브랜치)
    - [원격 저장소를 복제(clone)할 때 일어나는 일](#원격-저장소를-복제clone할-때-일어나는-일)
    - [`원격 추적 브랜치`의 주요 특징 및 기능](#원격-추적-브랜치의-주요-특징-및-기능)
    - [git remote?](#git-remote-1)
    - [원격 저장소 확인](#원격-저장소-확인)
    - [원격 저장소 추가](#원격-저장소-추가)
    - [원격 저장소로 push](#원격-저장소로-push)
    - [로컬의 원격 저장소 별칭 삭제](#로컬의-원격-저장소-별칭-삭제)
    - [브랜치 생성 후 upstream 설정](#브랜치-생성-후-upstream-설정)
        - [by set-url](#by-set-url)
        - [by push](#by-push)
    - [remote url 변경하기](#remote-url-변경하기)
    - [remote 조회 권한 있는지 확인](#remote-조회-권한-있는지-확인)
    - [조직이 변경됐을 때](#조직이-변경됐을-때)

## `원격 추적 브랜치`?

`원격 추적 브랜치`는 로컬 저장소에서 원격 저장소의 브랜치를 추적하기 위해 사용된다.
이 브랜치는 로컬에서 원격 저장소의 특정 브랜치의 상태를 반영하며, 원격 저장소와의 동기화를 용이하게 해준다.

## 원격 저장소를 복제(clone)할 때 일어나는 일

- 로컬 저장소의 브랜치들은 복제된 저장소에 `원격 추적 브랜치`로 저장된다.
- `원격 추적 브랜치`는 `.git` 디렉토리 밑에 `refs/remotes` 디렉토리에 저장된다.
- 저장소를 복제하게 되면 원본 저장소에 대한 정보가 `origin`이라는 이름의 리모트로 저장되게 된다
- `원격 추적 브랜치`의 이름은 이 리모트 이름을 이용해 표기된다.
    - 예를 들어, `origin` 리모트에 있는 `master` 브랜치는 `origin/master`라는 이름으로 복제된 저장소에 저장된다.

## `원격 추적 브랜치`의 주요 특징 및 기능

- 원격 브랜치의 로컬 반영
    - `원격 추적 브랜치`는 원격 저장소의 브랜치 상태를 로컬에 반영한다
    - 예를 들어, 원격 저장소에 있는 `origin/main` 브랜치의 최신 상태는 로컬의 `origin/main` `원격 추적 브랜치`에 반영된다
- 동기화 및 비교
    - `원격 추적 브랜치`는 로컬 브랜치와 원격 브랜치 간의 차이를 확인하는 데 사용된다
    - 이를 통해 로컬에서 작업한 내용이 원격 저장소와 어떻게 다른지 비교하고, 필요에 따라 동기화할 수 있다
- 자동 업스트림 설정
    - 새로운 로컬 브랜치를 원격 저장소에 처음 푸시할 때, Git은 이 로컬 브랜치와 연결될 `원격 추적 브랜치`를 자동으로 설정한다
    - 예를 들어, 로컬에서 `feature-branch`를 원격 저장소에 푸시하면, 로컬의 `feature-branch`는 자동으로 원격의 `origin/feature-branch`를 추적하게 된다
- 편리한 푸시 및 풀
    - `원격 추적 브랜치`가 설정되면, `git push`나 `git pull` 명령을 사용할 때 명시적으로 원격 브랜치를 지정하지 않아도 된다
    - Git은 자동으로 연결된 `원격 추적 브랜치`와의 동기화를 수행한다

## git remote?

- 추적되는 리파지토리 집합을 관리
  
```shell
git remote --verbose
```

## 원격 저장소 확인

```bash
git remote -v
```

```bash
❯ git remote -v

# 이 원격 저장소는 `changedorg` 조직의 `someservice` 리포지토리를 가리킵니다. 
# 이름 `otheruser`는 이 원격 저장소를 참조할 때 사용되는 별칭입니다.
otheruser    https://github.com/changedorg/someservice (fetch)
otheruser    https://github.com/changedorg/someservice (push)

# 기본적으로 많은 Git 프로젝트에서 원본 저장소를 가리키는 데 사용되는 이름입니다. 
# 이 경우, 조직 이름이 변경되기 전의 URL을 가리키고 있습니다.
origin    https://github.com/beforeorg/someservice.git (fetch)
origin    https://github.com/beforeorg/someservice.git (push)

# 조직 이름이 변경된 후의 원격 저장소를 가리킵니다. 
# 이는 `otheruser`와 같은 리포지토리를 가리키나, 서로 다른 용도 또는 별칭으로 사용될 수 있습니다.
changedorg    https://github.com/changedorg/someservice (fetch)
changedorg    https://github.com/changedorg/someservice (push)
```

로컬 Git 저장소에는 세 개의 서로 다른 원격 저장소가 설정되어 있음을 나타냅니다.
각각 `otheruser`, `origin`, `changedorg`라는 이름으로 구분되며, 이들 각각에 대해 `fetch`와 `push` 용도로 사용할 URL이 설정되어 있습니다.
여기서 `fetch`는 해당 원격 저장소에서 데이터를 가져오는 데 사용되며, `push`는 로컬의 변경사항을 원격 저장소에 업로드하는 데 사용됩니다.

`origin`, `changedorg`, `otheruser` 등은 원격 저장소에 대한 로컬 별칭입니다.
이 별칭들은 로컬 Git 환경에서만 존재하며, 원격 서버에는 영향을 미치지 않습니다. 각각 다른 원격 리포지토리를 가리키는 데 사용됩니다.
원격 저장소의 실제 이름은 URL에 포함되어 있으며, 이 별칭들은 단지 로컬에서 사용자가 원격 리포지토리를 참조할 때 편의를 제공하기 위한 것입니다.
예를 들어, `git push origin master`는 `origin`이라는 별칭을 사용하여 특정 원격 저장소로 코드를 푸시합니다.

Git은 협업 도구로서 다양한 소스의 리포지토리를 동시에 관리할 수 있도록 설계되었습니다.
예를 들어, 개발자가 여러 기여자의 리포지토리를 포크하거나 여러 개의 업스트림 리포지토리를 동시에 관리할 수 있습니다.
`otheruser`, `changedorg`와 같은 별칭은 여러 출처의 코드를 통합하거나 참조할 때 유용합니다.

## 원격 저장소 추가

```bash
# git remote add origin https://github.com/username/repository.git
git remote add origin <원격 저장소 URL>
```

## 원격 저장소로 push

```bash
git push -u origin master
```

## 로컬의 원격 저장소 별칭 삭제

별칭을 제거하면 해당 원격 저장소에 대한 참조가 로컬 저장소에서 완전히 삭제됩니다.

```bash
git remote remove otheruser
git remote remove changedorg
```

이 작업은 별칭이 참조하는 실제 원격 저장소에는 영향을 주지 않으며, 단지 로컬 Git 설정에서 이러한 원격 저장소에 대한 참조를 제거할 뿐입니다.

## 브랜치 생성 후 upstream 설정

### by set-url

### by push

```shell
# 로컬 브랜치 생성 및 체크아웃
git checkout -b new-branch

# 원격 저장소에 브랜치 푸시 및 `원격 추적 브랜치` 설정
git push --set-upstream origin new-branch
```

## remote url 변경하기

```shell
git remote set-url origin https://github.com/some-qwerty-org.io/$(basename $(pwd)).git
```

## remote 조회 권한 있는지 확인

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
password=gho_aNB9VhF.... # expired


# 권한 있는지 확인. 근데 없음.
❯ g ls-remote
remote: Repository not found.
fatal: repository 'https://github.com/some-org/some-qwerty-org.io-go.git/' not found


# 어떤 크리덴셜에 권한이 있는지 여부 확인
❯ g ls-remote https://aimpugn:<VALID_CREDENTIAL>@github.com/some-org/some-qwerty-org.io-go.git
f326223925bde5f9498415ff2e7ce6f8cb48d810    HEAD
d5a7b6416189e140b1de27442f276b75fcf5d1f6    refs/heads/some-qwerty-api-service/feature/application
f789294769853904cdf0dc31ae13c3a13f63461f    refs/heads/some-qwerty-pay-service/chore/delete-unnecessary-logic-on-danal-tpay-ctp
0c9649a68ced2fd404f94911aa01ca6e8b1f1ee1    refs/heads/some-qwerty-pay-service/chore/depr-ctp-gen


# 하지만 만료된 토큰으로 실행하면 에러 발생
❯ g ls-remote https://aimpugn:gho_aNB9VhF....@github.com/some-org/some-qwerty-org.io-go.git
remote: Repository not found.
fatal: repos
```

## 조직이 변경됐을 때

GitHub에서 조직 이름이 변경되면 원격 리포지토리의 URL도 변경됩니다.
로컬 Git 설정에서 이를 반영하려면 기존의 원격 리포지토리 URL을 새로운 URL로 업데이트해야 합니다.

1. 현재 원격 저장소 URL 확인

    우선 현재 설정된 원격 저장소의 URL을 확인합니다.
    이렇게 하면 어떤 원격 저장소가 설정되어 있는지, 그리고 그것들이 어떤 URL을 사용하고 있는지 알 수 있습니다.

    ```bash
    git remote -v
    ```

    이 명령은 모든 원격 저장소와 그들의 URL을 나열합니다.
     보통 `origin` 이라는 이름으로 기본 원격 저장소가 설정되어 있습니다.

2. 원격 저장소 URL 변경

    변경된 조직의 새 URL로 원격 저장소의 주소를 업데이트해야 합니다. 이를 위해 다음 명령을 사용할 수 있습니다:

    ```bash
    git remote set-url <원격 저장소 이름> <새 URL>
    ```

    예를 들어, 원격 저장소의 이름이 `origin`이고, 조직 이름이 `aaa`에서 `bbb`로 변경되었다면, 아래와 같이 명령을 실행합니다:

    ```bash
    git remote set-url origin https://github.com/bbb/your-repository.git
    ```

    참고로 URL 끝에 `.git`을 붙이는 것은 선택 사항입니다.
    일반적으로 `.git` 확장자는 Git 리포지토리를 명시적으로 나타내는 데 사용되며, 대부분의 Git 호스팅 서비스는 .git이 붙어 있든 없든 URL을 올바르게 처리합니다. 따라서, `.git` 확장자의 유무는 기능상 차이를 만들지 않으며 스타일이나 개인적인 선호에 따라 선택할 수 있습니다.

3. 변경된 URL 확인

    원격 저장소 URL을 변경한 후, 변경 사항이 제대로 적용되었는지 확인해 봅니다.

    ```bash
    git remote -v
    ```

4. 추가 사항

    변경 후에는 로컬 리포지토리와 원격 리포지토리 간의 연결을 테스트하는 것이 좋습니다.

    ```bash
    git fetch origin
    ```

    이 명령은 새로운 URL에서 데이터를 성공적으로 가져올 수 있는지 확인합니다.
    만약 문제가 없다면, 모든 설정이 정상적으로 작동하는 것입니다.
