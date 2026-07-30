# Python 없는 Linux에서 GitLab 그룹 저장소를 일괄 이전하기

- [무엇을 자동화하는가](#무엇을-자동화하는가)
- [필요한 프로그램과 권한](#필요한-프로그램과-권한)
- [실행 전 점검](#실행-전-점검)
- [안전한 실행 순서](#안전한-실행-순서)
- [출력과 재실행](#출력과-재실행)
- [복사 범위와 안전 경계](#복사-범위와-안전-경계)
- [실패 상황별 조치](#실패-상황별-조치)
- [Linux에서 검증하기](#linux에서-검증하기)

## 무엇을 자동화하는가

[`migrate_gitlab_group.sh`](tools/migrate_gitlab_group.sh)는 원본 GitLab
그룹의 하위 그룹과 프로젝트를 조회하여 대상 GitLab에 없는 항목만
만듭니다. Python은 사용하지 않습니다.

예를 들어 다음 경로는 원본 루트 아래의 상대 경로를 유지합니다.

| 원본 | 대상 |
| --- | --- |
| `kcf/payment-api` | `axlab/division-dev/kcf/payment-api` |
| `kcf/payments/card-api` | `axlab/division-dev/kcf/payments/card-api` |

동작은 세 단계로 나뉩니다.

1. 옵션 없음

    GitLab을 조회하고 계획만 출력합니다. 대상 GitLab은 변경하지 않습니다.

2. `--apply`

    없는 하위 그룹과 빈 프로젝트만 만듭니다. README, 라이선스,
    `.gitignore`는 초기화하지 않습니다.

3. `--apply --migrate-created`

    이 도구의 상태 파일에서 직접 생성했다고 확인되는 프로젝트에만
    원본 브랜치와 태그를 전송합니다.

대상에 이미 존재하는 프로젝트는 내용이나 이름이 같아 보여도 변경하지
않습니다. 대상 프로젝트 B에만 새 브랜치가 있는 상황에서 원본 A의
브랜치를 자동 병합하는 도구가 아닙니다.

## 필요한 프로그램과 권한

### Linux 프로그램

다음 프로그램이 필요합니다.

```text
bash 4 이상
curl
jq
git
base64, awk, sort, sed, grep 등 일반 GNU 도구
```

Git LFS 저장소를 옮길 때만 `git-lfs`가 추가로 필요합니다.

```bash
bash --version
curl --version
jq --version
git --version
git lfs version  # LFS를 사용할 때만
```

스크립트는 `jq`가 있으면 이를 우선 사용합니다. `jq`가 없는 환경에서는
Perl의 `JSON::PP`가 설치되어 있을 때 fallback으로 사용할 수 있지만,
현재 실행 환경처럼 `jq`가 있다면 Perl도 Python도 필요하지 않습니다.

`sh migrate_gitlab_group.sh`로 실행하지 않습니다. 이 스크립트는 Bash
배열과 정규식을 사용하므로 반드시 실행 파일로 직접 호출하거나
`bash migrate_gitlab_group.sh`로 실행합니다.

### GitLab 토큰

원본 토큰은 그룹·프로젝트 API를 읽고 모든 원본 저장소를 clone할 수
있어야 합니다.

- 개인 액세스 토큰: `read_api`, `read_repository`
- 원본 그룹 액세스 토큰: `read_api`, `read_repository`

대상 토큰은 하위 그룹과 프로젝트를 만들고, Git push와 기본 브랜치
설정을 수행할 수 있어야 합니다.

- 개인 액세스 토큰: `api`
- 대상 그룹 액세스 토큰: `api`, `write_repository`

개인 액세스 토큰의 `api`는 Git-over-HTTP 쓰기까지 포함합니다. 그룹
액세스 토큰의 `api`에는 Git-over-HTTP 쓰기가 포함되지 않으므로
`write_repository`를 같이 부여해야 합니다. 자세한 범위는
[GitLab 액세스 토큰 범위](https://docs.gitlab.com/security/tokens/access_token_scopes/)에서
확인합니다.

토큰 범위만으로 실제 권한이 생기지는 않습니다. 대상 토큰의 사용자는
대상 루트 그룹 아래에 하위 그룹과 프로젝트를 만들 수 있는 역할도 가져야
합니다.

### 대상 루트 그룹

`--target-group`으로 지정할 루트 그룹은 미리 만들어 둡니다. 대상 경로가
`axlab/division-dev/kcf`라면 이 그룹까지는 존재해야 합니다. 그 아래의
하위 그룹과 프로젝트는 스크립트가 필요할 때 만듭니다.

## 실행 전 점검

저장소 루트에서 실행 파일과 문법을 확인합니다.

```bash
chmod +x ./git/tools/migrate_gitlab_group.sh
bash -n ./git/tools/migrate_gitlab_group.sh
./git/tools/migrate_gitlab_group.sh --help
```

원본과 대상에는 브라우저에서 처음 입력하는 주소가 아니라 리다이렉트가
끝난 정식 HTTPS 주소를 사용합니다.

```text
https://source.gitlab.example
https://target.gitlab.example
```

스크립트는 토큰을 다른 주소로 전달하지 않기 위해 API와 Git
리다이렉트를 거부합니다. `http://` 주소도 기본적으로 거부합니다. 정말
평문 HTTP가 필요한 격리 환경에서만 위험을 이해한 뒤
`--allow-insecure-http`를 사용합니다.

### 토큰을 현재 Bash 프로세스에 넣는다

다음 방식은 토큰을 화면이나 셸 명령 기록에 그대로 남기지 않습니다.

```bash
read -rsp '원본 GitLab 토큰: ' SOURCE_GITLAB_TOKEN
printf '\n'
read -rsp '대상 GitLab 토큰: ' TARGET_GITLAB_TOKEN
printf '\n'
export SOURCE_GITLAB_TOKEN TARGET_GITLAB_TOKEN
```

스크립트는 API 토큰을 curl 명령 인자로 넘기지 않습니다. curl의 표준
입력 설정으로 전달합니다. Git 인증도 URL이나 Git 명령 인자에 토큰을
넣지 않고 자식 프로세스의 일회성 Git 설정 환경으로 전달합니다.

같은 사용자 권한으로 실행되는 다른 프로세스에서 환경 변수를 읽을 수
있는 운영 체제 설정도 있으므로, 이전 전용 계정과 제한된 실행 호스트를
사용하는 편이 안전합니다.

## 안전한 실행 순서

아래 예제에서 URL과 그룹 경로만 실제 값으로 바꿉니다. 작업 디렉터리는
중간 상태와 bare 저장소를 보존하므로 모든 실행에서 같은 경로를
사용합니다.

### 1. 계획을 확인한다

```bash
./git/tools/migrate_gitlab_group.sh \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir '/data/gitlab-migration/kcf'
```

`PLAN_CREATE_GROUP`, `PLAN_CREATE_PROJECT`, `SKIP_PROJECT_EXISTS`의 경로가
예상과 맞아야 합니다. 이 단계도 상태 파일과 임시 작업 디렉터리는
로컬에 만들지만 GitLab에는 쓰기 요청을 보내지 않습니다.

### 2. 쓰기가 없는 프로젝트 하나로 시험한다

정규식은 원본 프로젝트 전체 경로에 적용됩니다.

```bash
./git/tools/migrate_gitlab_group.sh \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir '/data/gitlab-migration/kcf' \
    --project-regex '^kcf/payments/payment-api$' \
    --apply \
    --migrate-created
```

다음 순서로 끝나야 합니다.

```text
CREATE_PROJECT
MIGRATE_REFS
VERIFY_PASS
```

대상 GitLab에서도 브랜치, 태그, 기본 브랜치와 일반 clone을 확인합니다.
시험 프로젝트는 최종 전환까지 원본 push가 없는 저장소를 고릅니다.
한 번 `VERIFY_PASS`가 난 프로젝트는 같은 상태 파일에서
`SKIP_VERIFIED`로 처리됩니다.

### 3. 전체 대상 프로젝트를 먼저 만든다

원본 서비스를 계속 사용하는 동안 Git 전송 없이 빈 프로젝트를 준비할
수 있습니다.

```bash
./git/tools/migrate_gitlab_group.sh \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir '/data/gitlab-migration/kcf' \
    --apply
```

### 4. 원본 쓰기를 중지하고 전체 이력을 이전한다

최종 실행 전에 원본 그룹의 push를 중지합니다. 원본 쓰기를 계속
허용하면 프로젝트별 fetch 시점이 달라지고, 검증 뒤에 추가된 커밋은
대상에 없습니다.

```bash
./git/tools/migrate_gitlab_group.sh \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir '/data/gitlab-migration/kcf' \
    --apply \
    --migrate-created
```

LFS 객체도 옮겨야 한다면 마지막 옵션을 추가합니다.

```bash
./git/tools/migrate_gitlab_group.sh \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir '/data/gitlab-migration/kcf' \
    --apply \
    --migrate-created \
    --migrate-lfs
```

일반 Git 이전만 검증했던 도구 소유 프로젝트도 같은 상태 파일에
`--migrate-lfs`를 추가하면 `RESUME_LFS`로 다시 처리합니다.

모든 프로젝트가 `VERIFY_PASS` 또는 `SKIP_VERIFIED`로 끝나고 대상의
일반 clone까지 확인한 뒤 사용자 push 경로를 새 GitLab로 전환합니다.

### 5. 토큰을 지운다

```bash
unset SOURCE_GITLAB_TOKEN TARGET_GITLAB_TOKEN
```

## 출력과 재실행

| 출력 | 의미 | 조치 |
| --- | --- | --- |
| `PLAN_CREATE_GROUP` | 적용 시 생성할 하위 그룹 | 경로 확인 |
| `PLAN_CREATE_PROJECT` | 적용 시 생성할 프로젝트 | 경로 확인 |
| `SKIP_GROUP_EXISTS` | 기존 대상 그룹 재사용 | 정상 |
| `SKIP_PROJECT_EXISTS` | 기존 대상 프로젝트 무변경 | 정상 |
| `CREATE_GROUP` | 하위 그룹 생성 완료 | 계속 |
| `CREATE_PROJECT` | 빈 프로젝트 생성 완료 | 계속 |
| `BLOCK_CREATE_AMBIGUOUS` | 생성 응답 실패 뒤 같은 경로 발견 | 소유권 수동 확인 |
| `MIGRATE_REFS` | 브랜치·태그 전송 시작 | Git 진행 출력 확인 |
| `VERIFY_PASS` | 원본과 대상 ref 일치 | 해당 프로젝트 통과 |
| `MIGRATION_FAIL` | 전송·기본 브랜치·검증 실패 | 원인 조치 후 재실행 |
| `RESUME_CREATED` | 도구가 만든 미완료 프로젝트 재개 | 정상 |
| `RESUME_LFS` | LFS 전송을 추가로 수행 | 정상 |
| `SKIP_VERIFIED` | 이전에 검증 완료 | 정상 |
| `SUMMARY` | 생성·skip·검증·실패 집계 | failures가 0인지 확인 |

작업 디렉터리에는 다음 파일이 남습니다.

```text
state.tsv
repositories/
```

`state.tsv`에는 토큰을 기록하지 않습니다. 작업 디렉터리를 삭제하지 않고
같은 인자로 재실행하면 `created`, `migration_failed` 상태를 이어서
처리합니다.

생성 API의 응답을 받지 못한 뒤 같은 대상 경로가 발견되면 이 도구가
만든 프로젝트인지 동시 실행이나 다른 사용자가 만든 것인지 단정할 수
없습니다. 이 경우 `BLOCK_CREATE_AMBIGUOUS`로 계속 실패합니다. 생성자와
내용을 확인하고, 이번 작업에서 생성된 빈 프로젝트가 확실할 때만
수동으로 삭제한 뒤 재실행합니다.

같은 `--work-dir`을 여러 프로세스에서 동시에 사용하지 않습니다. 상태
파일은 한 프로세스의 순차 실행만 지원합니다.

## 복사 범위와 안전 경계

### 브랜치와 태그

다음 refspec만 fetch하고 push합니다.

```text
refs/heads/*:refs/heads/*
refs/tags/*:refs/tags/*
```

`git push --mirror`를 사용하지 않으므로 GitLab의
`refs/merge-requests/*` 같은 서버 관리 ref를 대상에 쓰지 않습니다.
대상 push에는 force와 prune을 사용하지 않습니다. 여러 ref는
`--atomic`으로 전송합니다.

전송 뒤 원본과 대상의 모든 브랜치·태그 SHA를 비교합니다. 누락, SHA
불일치, 대상에만 있는 예상하지 않은 브랜치나 태그가 하나라도 있으면
`VERIFY_PASS`가 되지 않습니다.

### GitLab 메타데이터

다음 항목은 복사하지 않습니다.

- Merge Request와 이슈
- 멤버와 권한
- protected branch와 approval rule
- CI/CD 변수, runner, webhook
- deploy key와 deploy token
- release 첨부 파일과 package/container registry
- 위키와 서브모듈 저장소
- archive 상태와 fork 관계

`--copy-visibility`를 지정하면 원본 하위 그룹과 프로젝트의 visibility를
복사하려고 시도합니다. 지정하지 않으면 대상 namespace의 기본값을
따릅니다.

GitLab API가 제공하는 하위 그룹과 그룹 프로젝트 조회, 프로젝트 생성
규칙은 [Groups API](https://docs.gitlab.com/api/groups/)와
[Projects API](https://docs.gitlab.com/api/projects/)에서 확인할 수
있습니다. 프로젝트 수가 100개를 넘으면
[REST API 페이지네이션](https://docs.gitlab.com/api/rest/#pagination)에
맞춰 다음 페이지를 계속 조회합니다.

## 실패 상황별 조치

### `HTTP 401` 또는 `HTTP 403`

토큰 scope와 토큰 사용자의 실제 그룹 역할을 함께 확인합니다.

- 원본 API는 되지만 fetch 실패: `read_repository`와 프로젝트 접근 권한
- 대상 생성 실패: 대상 그룹의 하위 그룹·프로젝트 생성 권한
- 프로젝트 생성은 되지만 push 실패: `write_repository`, protected
  branch, initial push 정책

### `HTTP 301`, `302`, `307`, `308`

정식 GitLab 주소가 아닙니다. 응답의 `Location`을 자동으로 신뢰해 토큰을
전달하지 않습니다. 운영자가 확인한 최종 HTTPS 주소로 입력을 고칩니다.

### `HTTP 413`

Git refspec 문제가 아니라 GitLab 앞의 NGINX나 다른 프록시가 요청 본문을
거부한 것입니다. `client_max_body_size`, 상위 프록시 제한과 timeout을
조정한 뒤 같은 작업 디렉터리로 재실행합니다.

저장소 크기가 160 MB여도 실제 HTTP 요청의 pack 크기는 정확히 160 MB가
아닙니다. `http.postBuffer`를 크게 하는 것만으로 서버가 반환한 413의
근본 원인이 해결되지는 않습니다.

### Git 명령이 오래 멈춘 것처럼 보인다

fetch와 push에는 `--progress`를 사용합니다. 그래도 출력이 없다면 DNS,
프록시, 방화벽과 대상 HTTPS 포트 연결을 확인합니다. pack 압축 시간과
연결 자체가 되지 않는 상태를 구분해야 합니다.

### `--atomic` push 거부

서버가 atomic push를 지원하지 않거나 protected branch, hook, 저장소
정책이 ref 하나를 거부한 것입니다. atomic을 제거해 부분 성공을
허용하지 말고 거부된 ref와 대상 정책을 확인합니다.

## Linux에서 검증하기

Python 없이 다음 검증을 실행할 수 있습니다.

```bash
bash -n ./git/tools/migrate_gitlab_group.sh
bash -n ./git/tools/test_migrate_gitlab_group.sh
bash ./git/tools/test_migrate_gitlab_group.sh
```

테스트는 다음 동작을 확인합니다.

- GitLab JSON 응답과 빈 프로젝트 생성 payload
- 상태 파일 갱신과 그룹 상대 경로 매핑
- 모의 GitLab API에서 plan이 POST를 보내지 않음
- 적용 모드에서 기존 프로젝트는 보존하고 없는 그룹·프로젝트만 생성
- 토큰이 curl 명령 인자에 포함되지 않음
- 로컬 bare 저장소 사이의 여러 브랜치와 태그 실제 전송
- 원본과 대상의 SHA 일치
- 대상에만 있는 branch 발견 시 검증 실패

로컬 테스트는 실제 GitLab의 역할, 보호 규칙, hook, quota와 NGINX
제한까지 검증하지 않습니다. 실제 환경에서는 계획, 쓰기 없는 프로젝트
하나의 시험 이전, 전체 프로젝트 생성, 원본 쓰기 중지, 전체 이전 순서로
진행합니다.
