# GitLab 그룹 저장소 생성 및 Git 데이터 이전 자동화

## 0. Meta

- 작업 제목: GitLab 그룹 저장소 생성 및 Git 데이터 이전 자동화
- WORK 파일 경로: `docs/works/WORK_20260730_GITLAB_GROUP_REPOSITORY_PROVISIONER.md`
- 저장소: `study`
- 작업 유형: `design | execute`
- 작업 깊이: `full`
- 관련 요청: 원본 `kcf` 그룹의 프로젝트를 대상 그룹 아래에 생성하고 Git 이력을 이전하는 반복 작업 자동화
- 대상 경로 / 자산: `git/tools`, `git/gitlab_group_repository_migration.md`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 두 GitLab의 액세스 토큰과 그룹 경로를 입력받아 원본 하위 그룹과 프로젝트를 조회하고, 대상에 없는 항목만 생성한다.
- mode: 기본 `plan`, 명시적 `--apply`에서만 외부 상태 변경
- must_keep:
  - 대상에 이미 존재하는 프로젝트는 자동 변경하지 않는다.
  - 토큰을 명령행, URL, 로그, 상태 파일에 기록하지 않는다.
  - GitLab 내부 ref는 push하지 않는다.
  - 부분 실패 후 재실행할 수 있다.
- extra_checks:
  - API 페이지네이션
  - 중첩 그룹의 부모 순서
  - 생성 프로젝트의 README 미초기화
  - branches와 tags의 SHA 사후 비교

## 2. Root-First Framing

- 근본 문제: 단순 반복문은 대상 경로 충돌, hidden ref, 토큰 노출, 부분 실패 이후의 재실행을 안전하게 처리하지 못한다.
- 성공 정의:
  - 계획 실행은 외부 상태를 바꾸지 않는다.
  - 적용 실행은 없는 그룹과 프로젝트만 생성한다.
  - 기존 대상 프로젝트는 `SKIP_EXISTING`으로 남는다.
  - 도구가 생성한 프로젝트만 선택적으로 Git 데이터 이전과 검증을 수행한다.
  - 모의 API 테스트와 로컬 Git 테스트가 통과한다.
- BLOCKED 조건: 대상 루트 그룹 부재, API 권한 부족, 기존 경로 충돌, Git 전송 또는 ref 검증 실패

## 3. Scope

- 포함:
  - 그룹과 하위 그룹 조회 및 생성
  - 프로젝트 조회 및 생성
  - 선택적인 branches/tags 이전
  - 기본 브랜치 설정
  - 상태 파일과 프로젝트별 작업 디렉터리
- 제외:
  - 기존 대상 프로젝트 자동 덮어쓰기
  - Merge Request, 이슈, CI/CD 변수, webhook, 권한 이전
  - protected branch 자동 변경
  - 대상 루트 그룹 자동 생성

## 4. Evidence

- GitLab Groups API는 그룹 프로젝트 조회에서 `include_subgroups`, `with_shared`, 페이지네이션을 제공한다.
- GitLab Groups API는 `parent_id`로 하위 그룹을 생성한다.
- GitLab Projects API는 `namespace_id`로 그룹 아래에 빈 프로젝트를 생성하며 `initialize_with_readme=false`를 지원한다.
- Git push는 명시적 `refs/heads/*`와 `refs/tags/*` refspec, `--atomic`, `--porcelain`을 제공한다.
- GitLab 서버 관리 ref는 전체 `push --mirror`에서 거부될 수 있으므로 portable Git 데이터 범위를 branches와 tags로 제한한다.

## 5. Design

- Linux의 Bash 4 이상, curl, jq, Git과 일반 GNU 도구만 사용한다.
- Python은 실행 의존성에서 제외한다.
- jq를 우선 사용하고, jq가 없는 환경에서는 Perl의 표준 JSON 모듈을
  보조 경로로 사용할 수 있다.
- 토큰은 기본적으로 `SOURCE_GITLAB_TOKEN`, `TARGET_GITLAB_TOKEN` 환경 변수에서 읽는다.
- 실행 모드:
  - 기본: 계획 출력만
  - `--apply`: 없는 그룹과 프로젝트 생성
  - `--apply --migrate-created`: 상태 파일에서 이 도구가 생성했다고 확인된 프로젝트만 이전
- 대상에 이미 존재하지만 상태 파일 소유권 근거가 없는 프로젝트는 항상 건너뛴다.
- Git 인증은 토큰이 URL이나 명령행에 노출되지 않도록 자식 프로세스의 일회성 Git 설정 환경 변수로 전달한다.
- source fetch는 branches와 tags만 로컬 bare 저장소에 배치한다.
- target push는 force와 prune 없이 atomic push하고 SHA를 비교한다.

## 6. Verification

- PASS: WSL2 Ubuntu에서 두 스크립트의 `bash -n`
- PASS: WSL2 Ubuntu에서 CLI `--help`
- PASS: WSL2 Ubuntu에서 jq 1.8.1 경로로 6개 Bash 테스트
  - JSON 응답과 프로젝트 생성 payload
  - 상태 파일과 중첩 그룹 경로 매핑
  - 모의 GitLab API 계획 실행, POST 차단, curl 인자 토큰 비노출
  - 적용 모드에서 기존 프로젝트 보존, 없는 그룹과 프로젝트만 생성
  - 잘못된 프로젝트 정규식의 API 접근 전 거부
  - 로컬 bare source/target의 여러 branches/tags 실제 이전
  - ref SHA 일치와 대상 전용 ref 거부
- PASS: jq가 없는 WSL2 Ubuntu에서 Perl fallback으로 같은 6개 테스트
- PASS: ShellCheck 0.11.0
- PASS: 최종 변경의 `git diff --check`
- 미실행: 실제 사내 GitLab API와 프록시를 통과하는 end-to-end 실행

## 7. Open Risks

- 실제 GitLab 버전과 그룹 정책에 따라 프로젝트 생성 또는 초기 push 권한이 거부될 수 있다.
- 외부 NGINX 요청 크기 제한은 스크립트가 바꾸지 않는다. HTTP 413은 운영 설정을 수정한 뒤 재실행해야 한다.
- Git LFS는 명시적 `--migrate-lfs`에서만 처리한다.

## 8. Final Audit

- 대상 기존 프로젝트는 상태 파일 소유권 근거가 없으면 항상 건너뛴다.
- 토큰은 환경 변수로만 받고 URL, 명령 인자, 상태 파일에 저장하지 않는다.
- HTTP와 리다이렉트는 기본적으로 실패-폐쇄 처리한다.
- target push는 branches와 tags만 대상으로 하며 force와 prune을 사용하지 않는다.
- 대상 ref에는 누락, SHA 불일치, 예상하지 않은 branch/tag가 없어야 검증을 통과한다.
- 실제 사내 GitLab 권한, 보호 규칙, hook, quota, NGINX 413 제한은 배포 환경에서 시험 프로젝트로 별도 확인해야 한다.
