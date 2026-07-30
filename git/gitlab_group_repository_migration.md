# GitLab 그룹의 저장소를 새 GitLab로 일괄 이전하기

- [무엇을 자동화하는가](#무엇을-자동화하는가)
- [실행 전에 준비할 것](#실행-전에-준비할-것)
- [Windows PowerShell에서 실행하기](#windows-powershell에서-실행하기)
- [Linux Bash에서 실행하기](#linux-bash에서-실행하기)
- [출력과 재실행을 판단하는 법](#출력과-재실행을-판단하는-법)
- [안전 경계와 제외 범위](#안전-경계와-제외-범위)
- [실패 상황별 조치](#실패-상황별-조치)
- [직접 검증한 범위](#직접-검증한-범위)

## 무엇을 자동화하는가

원본 GitLab 그룹 아래에 프로젝트가 많다면
[`migrate_gitlab_group.py`](tools/migrate_gitlab_group.py)로 다음 작업을
반복할 수 있습니다.

1. 원본 그룹의 모든 하위 그룹과 프로젝트를 GitLab API로 조회합니다.
2. 대상의 같은 경로에 이미 있는 그룹은 재사용합니다.
3. 대상에 이미 있는 프로젝트는 아무것도 변경하지 않고 건너뜁니다.
4. 없는 하위 그룹과 프로젝트만 생성합니다.
5. 선택한 경우, 이번 도구가 생성한 빈 프로젝트에만 원본의 모든 브랜치와
   태그를 전송합니다.
6. 원본과 대상의 브랜치·태그별 객체 ID(SHA)를 비교합니다.

예를 들어 다음 경로는 상대 경로를 유지한 채 매핑됩니다.

| 원본 | 대상 |
| --- | --- |
| `kcf/payment-api` | `axlab/division-dev/kcf/payment-api` |
| `kcf/payments/card-api` | `axlab/division-dev/kcf/payments/card-api` |

기본 실행은 계획만 출력합니다. `--apply`가 없으면 그룹이나 프로젝트를
생성하지 않고 Git 데이터도 전송하지 않습니다.

`--apply --migrate-created`를 사용해도 기존 대상 프로젝트는 변경하지
않습니다. 생성 직후 통신이 끊긴 프로젝트처럼 이 도구의 상태 파일에서
소유권을 확인할 수 없는 대상 프로젝트도 안전을 위해 건너뜁니다.

## 실행 전에 준비할 것

### 1. 실행 환경

다음 프로그램이 필요합니다.

- Python 3.10 이상
- Git
- LFS 저장소를 옮기는 경우 Git LFS
- 원본과 대상 GitLab의 최종 HTTPS 주소

HTTP 주소를 HTTPS로 바꾸는 리다이렉트에 의존하지 않습니다. 이 도구는
토큰을 다른 주소로 전달하지 않기 위해 API와 Git의 리다이렉트를
거부합니다. 브라우저에서 처음 입력하는 주소가 아니라, 리다이렉트가 끝난
GitLab의 정식 HTTPS 주소를 `--source-url`과 `--target-url`에 넣습니다.

### 2. 토큰 권한

원본 토큰은 그룹·프로젝트를 조회하고 각 저장소를 clone할 수 있어야
합니다. 개인 액세스 토큰을 쓴다면 보통 다음 범위가 필요합니다.

- `read_api`
- `read_repository`

대상 토큰은 하위 그룹과 프로젝트를 만들고, 저장소에 push하고, 기본
브랜치를 설정할 수 있어야 합니다. 개인 액세스 토큰을 쓴다면 `api` 범위가
필요합니다. 그룹 액세스 토큰을 쓴다면 API 호출용 `api`와 Git push용
`write_repository`가 모두 필요합니다. GitLab 공식 범위표에서 개인
액세스 토큰의 `api`는 Git-over-HTTP 쓰기까지 포함하지만, 그룹 액세스
토큰의 `api`에는 그 권한이 포함되지 않기 때문입니다.

토큰 범위만 넓다고 권한이 생기지는 않습니다. 토큰 소유자는 대상 루트
그룹에서 하위 그룹과 프로젝트를 만들 수 있는 역할도 가져야 합니다.

GitLab 버전과 그룹 정책에 따라 생성 권한이 달라질 수 있습니다. 실제
적용 전에 [액세스 토큰 범위](https://docs.gitlab.com/security/tokens/access_token_scopes/)와
대상 그룹의 프로젝트·하위 그룹 생성 정책을 함께 확인합니다.

### 3. 대상 루트 그룹

대상 루트 그룹은 미리 만들어 둡니다. 예를 들어 대상 경로가
`axlab/division-dev/kcf`라면 이 그룹까지는 존재해야 합니다. 그 아래의
하위 그룹과 프로젝트는 도구가 필요할 때 생성합니다.

대상 프로젝트를 만들 때 README, 라이선스, `.gitignore`는 초기화하지
않습니다. 이렇게 해야 첫 Git push가 원본 이력과 충돌하지 않습니다.

### 4. 프록시의 push 제한

대상 GitLab 앞에 NGINX 같은 리버스 프록시가 있고 큰 저장소 push가
`HTTP 413`으로 실패했다면 먼저 프록시의 요청 본문 제한을 해결해야
합니다. 이 도구는 프록시 설정을 우회하거나 변경하지 않습니다.

160 MB인 저장소도 전송되는 pack 크기는 160 MB와 정확히 같지 않습니다.
NGINX의 `client_max_body_size`, GitLab Workhorse와 상위 프록시의 제한,
timeout을 실제 전송 크기와 시간에 맞춰 확인합니다.

## Windows PowerShell에서 실행하기

아래 예제는 저장소 루트에서 PowerShell 7로 실행합니다.

### 1. 현재 PowerShell 프로세스에 토큰을 넣는다

`Read-Host -MaskInput`을 사용하면 토큰 문자열이 화면과 명령 기록에
그대로 남지 않습니다.

```powershell
$env:SOURCE_GITLAB_TOKEN = Read-Host '원본 GitLab 토큰' -MaskInput
$env:TARGET_GITLAB_TOKEN = Read-Host '대상 GitLab 토큰' -MaskInput
```

환경 변수는 현재 PowerShell과 여기서 시작한 자식 프로세스가 읽을 수
있습니다. 공용 PC나 신뢰할 수 없는 프로세스가 함께 실행되는 환경에서는
사용하지 않습니다.

### 2. 계획을 확인한다

```powershell
python .\git\tools\migrate_gitlab_group.py `
    --source-url 'https://source.gitlab.example' `
    --target-url 'https://target.gitlab.example' `
    --source-group 'kcf' `
    --target-group 'axlab/division-dev/kcf' `
    --work-dir 'C:\gitlab-migration\kcf'
```

`PLAN_CREATE_GROUP`와 `PLAN_CREATE_PROJECT`만 출력되고 대상 GitLab은
변경되지 않아야 합니다. `SKIP_PROJECT_EXISTS`는 같은 대상 경로가 이미
있어서 건너뛴다는 뜻입니다.

### 3. 프로젝트 하나로 생성과 이전을 검증한다

정확한 원본 프로젝트 전체 경로를 정규식으로 지정합니다. 필터를 사용하면
그 프로젝트에 필요한 하위 그룹만 생성 대상이 됩니다.

```powershell
python .\git\tools\migrate_gitlab_group.py `
    --source-url 'https://source.gitlab.example' `
    --target-url 'https://target.gitlab.example' `
    --source-group 'kcf' `
    --target-group 'axlab/division-dev/kcf' `
    --work-dir 'C:\gitlab-migration\kcf' `
    --project-regex '^kcf/payments/payment-api$' `
    --apply `
    --migrate-created

if ($LASTEXITCODE -ne 0) {
    throw '시험 이전에 실패했습니다.'
}
```

`CREATE_PROJECT`, `MIGRATE_REFS`, `VERIFY_PASS` 순서로 끝나야 합니다.
대상 GitLab에서 브랜치, 태그, 기본 브랜치와 일반 clone도 확인합니다.
이 시험 프로젝트는 이전이 끝날 때까지 원본 push가 없는 저장소를
고릅니다. 한 번 `VERIFY_PASS`가 난 프로젝트는 같은 상태 파일에서
`SKIP_VERIFIED`로 처리되므로 이후 원본 변경을 자동으로 다시 반영하지
않습니다.

### 4. 전체 대상 프로젝트를 먼저 만든다

원본 서비스가 운영 중인 동안에는 Git 전송 없이 그룹과 빈 프로젝트만
만들 수 있습니다. 시험 이전이 통과하면 같은 작업 디렉터리로 필터를
제거하고 `--migrate-created` 없이 실행합니다.

```powershell
python .\git\tools\migrate_gitlab_group.py `
    --source-url 'https://source.gitlab.example' `
    --target-url 'https://target.gitlab.example' `
    --source-group 'kcf' `
    --target-group 'axlab/division-dev/kcf' `
    --work-dir 'C:\gitlab-migration\kcf' `
    --apply

if ($LASTEXITCODE -ne 0) {
    throw '대상 프로젝트 준비에 실패했습니다.'
}
```

### 5. 원본 쓰기를 중지하고 전체 Git 데이터를 이전한다

최종 이전 직전에 원본 그룹의 push를 중지합니다. push를 계속 허용하면
프로젝트마다 fetch 시점이 달라지고, 검증 뒤에 원본에 추가된 커밋은
대상에 없습니다. 원본 쓰기를 멈춘 상태에서 같은 작업 디렉터리에
`--migrate-created`를 추가합니다.

```powershell
python .\git\tools\migrate_gitlab_group.py `
    --source-url 'https://source.gitlab.example' `
    --target-url 'https://target.gitlab.example' `
    --source-group 'kcf' `
    --target-group 'axlab/division-dev/kcf' `
    --work-dir 'C:\gitlab-migration\kcf' `
    --apply `
    --migrate-created

if ($LASTEXITCODE -ne 0) {
    throw '하나 이상의 프로젝트 이전에 실패했습니다.'
}
```

Git LFS 객체도 이전해야 한다면 Git LFS 설치를 확인한 뒤
`--migrate-lfs`를 추가합니다.

```powershell
git lfs version

python .\git\tools\migrate_gitlab_group.py `
    --source-url 'https://source.gitlab.example' `
    --target-url 'https://target.gitlab.example' `
    --source-group 'kcf' `
    --target-group 'axlab/division-dev/kcf' `
    --work-dir 'C:\gitlab-migration\kcf' `
    --apply `
    --migrate-created `
    --migrate-lfs
```

이전에 일반 Git 이전만 검증한 도구 소유 프로젝트가 있더라도 같은 상태
파일에 `--migrate-lfs`를 추가하면 `RESUME_LFS`로 다시 처리합니다.

모든 프로젝트가 `VERIFY_PASS` 또는 이전에 완료된 `SKIP_VERIFIED`로
끝나고 대상의 일반 clone까지 확인한 뒤 사용자 push 경로를 대상 GitLab로
전환합니다.

### 6. 토큰을 현재 프로세스에서 지운다

```powershell
Remove-Item Env:SOURCE_GITLAB_TOKEN
Remove-Item Env:TARGET_GITLAB_TOKEN
```

## Linux Bash에서 실행하기

아래 예제는 저장소 루트에서 Bash로 실행합니다.

### 1. 현재 셸에 토큰을 넣는다

```bash
read -rsp '원본 GitLab 토큰: ' SOURCE_GITLAB_TOKEN
printf '\n'
read -rsp '대상 GitLab 토큰: ' TARGET_GITLAB_TOKEN
printf '\n'
export SOURCE_GITLAB_TOKEN TARGET_GITLAB_TOKEN
```

### 2. 계획을 확인한다

```bash
python3 ./git/tools/migrate_gitlab_group.py \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir "$PWD/.gitlab-migration/kcf"
```

### 3. 프로젝트 하나로 검증한다

```bash
python3 ./git/tools/migrate_gitlab_group.py \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir "$PWD/.gitlab-migration/kcf" \
    --project-regex '^kcf/payments/payment-api$' \
    --apply \
    --migrate-created
```

종료 코드가 `0`이고 `VERIFY_PASS`가 출력되어야 합니다.

### 4. 전체 대상 프로젝트를 먼저 만든다

```bash
python3 ./git/tools/migrate_gitlab_group.py \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir "$PWD/.gitlab-migration/kcf" \
    --apply
```

### 5. 원본 쓰기를 중지하고 전체 Git 데이터를 이전한다

원본 push를 중지한 뒤 같은 작업 디렉터리로 실행합니다.

```bash
python3 ./git/tools/migrate_gitlab_group.py \
    --source-url 'https://source.gitlab.example' \
    --target-url 'https://target.gitlab.example' \
    --source-group 'kcf' \
    --target-group 'axlab/division-dev/kcf' \
    --work-dir "$PWD/.gitlab-migration/kcf" \
    --apply \
    --migrate-created
```

모든 프로젝트와 대상의 일반 clone을 확인한 뒤 현재 셸에서 토큰을
지웁니다.

```bash
unset SOURCE_GITLAB_TOKEN TARGET_GITLAB_TOKEN
```

## 출력과 재실행을 판단하는 법

| 출력 | 의미 | 다음 행동 |
| --- | --- | --- |
| `PLAN_CREATE_GROUP` | 적용 시 만들 하위 그룹 | 경로가 맞는지 확인 |
| `PLAN_CREATE_PROJECT` | 적용 시 만들 프로젝트 | 경로가 맞는지 확인 |
| `SKIP_GROUP_EXISTS` | 대상 그룹을 재사용 | 정상 |
| `SKIP_PROJECT_EXISTS` | 기존 프로젝트를 변경하지 않음 | 내용이 달라도 자동 병합하지 않음 |
| `CREATE_GROUP` | 없는 하위 그룹 생성 완료 | 계속 진행 |
| `CREATE_PROJECT` | README 없는 빈 프로젝트 생성 완료 | 계속 진행 |
| `BLOCK_CREATE_AMBIGUOUS` | 생성 응답 실패 뒤 같은 경로가 발견됨 | 자동 이전 중단, 소유권 수동 확인 |
| `MIGRATE_REFS` | 브랜치·태그 전송 시작 | Git 출력 확인 |
| `VERIFY_PASS` | 원본과 대상의 브랜치·태그 SHA 일치 | 해당 프로젝트 통과 |
| `MIGRATION_FAIL` | 전송 또는 SHA 검증 실패 | 원인 조치 후 같은 명령 재실행 |
| `RESUME_CREATED` | 이전 실행에서 만든 프로젝트의 작업 재개 | 정상 |
| `RESUME_LFS` | Git 검증 뒤 LFS 전송을 추가로 수행 | 정상 |
| `SKIP_VERIFIED` | 이전에 검증까지 끝난 프로젝트 | 정상 |

작업 디렉터리에는 다음 자산이 남습니다.

- `state.json`: 도구가 생성한 프로젝트 ID, 대상 경로, 진행 상태
- `repositories/`: 프로젝트별 로컬 bare 저장소와 Git 객체

토큰은 상태 파일에 기록하지 않습니다. 실패 후에는 작업 디렉터리를
삭제하지 말고 같은 인자로 재실행합니다. 그러면 `migration_failed` 상태인
도구 소유 프로젝트만 이어서 처리할 수 있습니다.

생성 요청의 응답을 받지 못한 뒤 대상 경로가 생기면 도구는 그 프로젝트를
자신이 만들었다고 단정할 수 없습니다. 이 경우
`BLOCK_CREATE_AMBIGUOUS`로 계속 실패합니다. 대상 프로젝트의 생성자와
내용을 확인한 뒤, 빈 프로젝트가 이 작업에서 생성된 것이 확실할 때만
수동으로 삭제하고 같은 명령을 재실행합니다.

`state.json`을 잃어버린 뒤 대상 프로젝트만 남았다면 도구는 그 프로젝트를
기존 프로젝트로 보고 건너뜁니다. 이 동작은 자동 덮어쓰기보다 복구가
번거롭더라도 기존 데이터를 보호하는 쪽을 택한 결과입니다.

## 안전 경계와 제외 범위

### 브랜치와 태그만 전송한다

이 도구는 다음 두 refspec만 가져오고 보냅니다.

```text
refs/heads/*:refs/heads/*
refs/tags/*:refs/tags/*
```

`git push --mirror`를 사용하지 않으므로 GitLab의
`refs/merge-requests/*` 같은 서버 관리 ref를 대상에 쓰지 않습니다. 대상
push에는 force와 prune을 사용하지 않습니다. 여러 ref는 `--atomic`으로
함께 전송하여 서버가 지원하는 범위에서 전부 성공하거나 전부 실패하게
합니다.

### 기존 프로젝트는 이름만 같아도 건너뛴다

“이미 존재”는 대상의 전체 프로젝트 경로가 같다는 뜻입니다. 도구는 기존
프로젝트의 브랜치나 SHA가 원본과 같은지 판단한 뒤 덮어쓰지 않습니다.
경로가 있으면 즉시 `SKIP_PROJECT_EXISTS`로 처리합니다.

기존 프로젝트 B에만 새 브랜치가 있고 원본 A의 과거 브랜치·태그를
추가해야 하는 상황은 이 자동화의 범위가 아닙니다. 그 경우에는 양쪽 ref의
충돌 여부와 보존 정책을 따로 정한 뒤 선택적으로 push해야 합니다.

### GitLab 프로젝트 메타데이터는 별도다

다음 항목은 이전하지 않습니다.

- Merge Request와 이슈
- 멤버와 권한
- protected branch와 approval rule
- CI/CD 변수와 runner
- webhook, deploy key, deploy token
- release 첨부 파일과 package/container registry
- 위키와 서브모듈 저장소

`--copy-visibility`를 지정하면 원본 하위 그룹과 프로젝트의 공개 범위를
복사하려고 시도합니다. 지정하지 않으면 대상 namespace의 기본값을
따릅니다. 민감한 저장소가 포함되어 있다면 기본값을 사용하고 대상 정책을
먼저 확인하는 편이 안전합니다.

GitLab API가 제공하는 그룹 프로젝트 조회, 하위 그룹 생성과 프로젝트
생성 규칙은 [Groups API](https://docs.gitlab.com/api/groups/)와
[Projects API](https://docs.gitlab.com/api/projects/)에서 확인할 수
있습니다. 프로젝트 수가 100개를 넘는 경우도
[REST API 페이지네이션](https://docs.gitlab.com/api/rest/#pagination)을
따라 모든 페이지를 조회합니다.

## 실패 상황별 조치

### `HTTP 401` 또는 `HTTP 403`

토큰 범위와 토큰 소유자의 실제 그룹 역할을 함께 확인합니다. 원본 조회는
되지만 clone이 실패한다면 `read_repository`와 프로젝트 접근 권한을
확인합니다. 대상 프로젝트 조회는 되지만 생성이 실패한다면 대상 루트
그룹의 하위 그룹·프로젝트 생성 권한을 확인합니다.

### `HTTP 301`, `302`, `307`, `308`

`--source-url` 또는 `--target-url`이 정식 주소가 아닙니다. 응답의
`Location`을 무조건 신뢰해 토큰을 전달하지 말고, 운영자가 확인한 최종
HTTPS GitLab 주소로 입력을 고칩니다.

### `HTTP 413`

Git 클라이언트나 refspec 문제가 아니라 중간 프록시가 요청 본문을 거부한
것입니다. NGINX와 그 앞단 프록시의 본문 크기 제한을 올린 뒤 같은 작업
디렉터리로 재실행합니다. `http.postBuffer`를 크게 하는 것은 서버가 낸
413의 근본 해결이 아닙니다.

### Git 명령이 오래 멈춘 것처럼 보인다

이 도구는 fetch와 push에 `--progress`를 사용합니다. 그래도 출력이 전혀
없다면 먼저 DNS, 프록시, 방화벽, 대상 HTTPS 포트 연결을 확인합니다.
160 MB 저장소의 압축과 pack 생성에는 시간이 걸릴 수 있지만, 연결 자체가
되지 않는 상태와 구분해야 합니다.

### `--atomic` push가 거부된다

대상 서버가 atomic push를 지원하지 않거나 protected branch, hook,
저장소 정책이 ref 하나를 거부한 것입니다. atomic을 제거해 부분 성공을
허용하지 말고 대상 정책과 거부된 ref를 확인합니다.

### 일부 프로젝트만 시험하고 싶다

`--project-regex`는 원본 전체 경로에 적용됩니다.

```text
--project-regex '^kcf/payments/payment-api$'
```

`--max-projects 1`도 사용할 수 있지만 프로젝트 ID 정렬에서 처음 선택된
항목이 이미 대상에 존재할 수 있습니다. 이름을 아는 시험 프로젝트에는
정확한 정규식을 쓰는 편이 명확합니다.

## 직접 검증한 범위

이 스크립트에는 다음 자동 테스트가 함께 있습니다.

```powershell
python -m unittest .\git\tools\test_migrate_gitlab_group.py
```

테스트는 다음 동작을 확인합니다.

- 대상에 이미 있는 프로젝트를 생성하거나 이전하지 않음
- 없는 중첩 그룹과 프로젝트만 생성
- 프로젝트 필터에서 불필요한 하위 그룹을 만들지 않음
- 이전 실패 후 상태 파일을 이용해 재실행
- API 페이지네이션과 `PRIVATE-TOKEN` 헤더
- Git 인증 헤더를 해당 저장소 URL에만 한정
- 로컬 bare 저장소 사이에서 여러 브랜치와 태그 전송
- 전송 후 원본과 대상의 ref별 SHA 일치

로컬 테스트는 실제 GitLab의 그룹 정책, 토큰 역할, protected branch,
hook, 저장소 quota, 프록시 제한까지 검증하지는 않습니다. 따라서 실제
이전은 계획 출력, 프로젝트 하나의 시험 이전, 전체 이전 순서로 진행합니다.

같은 상태 파일을 여러 프로세스에서 동시에 사용하지 않습니다. 이 도구는
한 작업 디렉터리에서 순차 실행하는 방식만 지원합니다.
