# 망 분리 환경 AI Agent 기반 개발/배포 자동화 구조 검토

작성일: 2026-04-21  
대상 독자: 보안 담당자, 인프라 담당자, 개발자  
산출물: 설계 검토 문서 + 회의 공유용 단일 HTML 파일  

이 문서는 특정 도구를 먼저 정답으로 두지 않고, 연결 방향과 데이터 흐름을 기준으로 AI agent 기반 개발 자동화와 배포 자동화 구조를 비교합니다. `확인`은 공식 문서나 제품 문서에서 직접 확인한 사실이고, `추론`은 그 사실과 제시된 망 분리 제약을 결합해 도출한 판단입니다.

## 1. 핵심 결론

현재 제약에서 가장 안전한 기본 구조는 **연구/개발망 GitLab은 source와 build 후보를 관리하고, 내부망은 artifact 검증, 배포 승인, 배포 실행, 배포 결과 로그를 소유하는 구조**입니다. DMZ는 배포 명령권자가 아니라, 먼저 `webhook inbox + artifact quarantine + 최소 검증 지점`으로 제한하는 편이 안전합니다.

`확인`: GitLab Runner는 GitLab 서버를 지속적으로 polling해 pending job을 받고, runner가 서버와 API로 통신합니다. GitLab 문서는 runner가 `POST /api/v4/jobs/request`로 job을 요청한다고 설명합니다. 따라서 runner가 어느 망의 GitLab에 붙는지는 단순 실행 위치가 아니라, job trace와 상태가 돌아가는 데이터 흐름을 결정합니다. [GitLab CI/CD development guidelines](https://docs.gitlab.com/development/cicd/)

`확인`: GitLab webhook은 secret token을 `X-Gitlab-Token` header로 보내고, custom payload template으로 body에 담는 정보를 줄일 수 있습니다. 또한 GitLab의 webhook request history에는 GitLab이 보낸 request와 받은 response detail이 남을 수 있습니다. [GitLab Webhooks](https://docs.gitlab.com/user/project/integrations/webhooks/)

`확인`: Jenkins는 inbound agent용 TCP listener를 열 수 있고, WebSocket transport를 쓰면 별도 TCP agent port 없이 agent가 controller에 붙을 수 있습니다. Jenkins credentials는 controller에 암호화 저장되지만, credentials masking은 build script가 악의적이거나 부주의하면 우회될 수 있습니다. [Jenkins exposed services](https://www.jenkins.io/doc/book/security/services/), [Jenkins credentials](https://www.jenkins.io/doc/book/using/using-credentials/), [Jenkins credentials masking limitations](https://www.jenkins.io/blog/2019/02/21/credentials-masking/)

`확인`: NIST SP 800-207의 zero trust 원칙은 네트워크 위치만으로 신뢰하지 않고, 사용자, 장치, resource, workflow별로 인증과 권한을 판단하라고 설명합니다. 이 요구사항에서는 “내부망에 있으니 안전하다”도, “DMZ라서 괜찮다”도 충분한 근거가 아닙니다. [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final)

`추론`: 연구/개발망에 GitLab을 두는 것 자체는 hard constraint와 즉시 충돌하지 않습니다. 충돌은 **내부망 배포 결과, 로그, 서버 목록, token, runner callback, artifact 소비 상태**가 연구/개발망 GitLab이나 업무망으로 되돌아갈 때 발생합니다. 따라서 연구/개발망 GitLab은 “배포 후보 생성”까지만 맡고, 내부망의 실제 배포 상태는 내부망 전용 감사 저장소에 남기는 구조가 기본값이어야 합니다.

`추론`: 내부망으로 들어오는 source/artifact 흐름은 통제된 inbound로 설계할 수 있지만, 내부망에서 밖으로 나가는 배포 상태와 로그는 현재 요구사항의 명시 금지 흐름입니다. “배포 성공/실패만 알려주는 것”도 운영 상태와 서버군 정보를 담을 수 있으므로 기본 금지로 보아야 합니다.

## 2. 현재 요구사항에서 가장 위험한 지점

가장 위험한 지점은 **배포 자동화 도구가 내부망 실행 결과를 외부 쪽 control plane으로 되돌리는 구조**입니다. CI/CD 도구는 보통 job scheduling, trace upload, artifact upload, environment/deployment status update, retry 상태, 실패 원인 기록을 자연스럽게 중앙 서버에 모읍니다. 그러나 이 중앙 서버가 연구/개발망이나 업무망에 있으면, 내부망에서 외부 방향으로 데이터가 흐릅니다.

특히 아래 항목은 모두 데이터 흐름으로 봐야 합니다.

- 배포 성공/실패 상태
- job trace, console log, stack trace, error message
- artifact download URL, digest, package metadata, SBOM 내용
- runner token, job token, deploy token, SSH key path, 환경 변수 이름
- 내부 서버 목록, 서버 alias, 배포 group, 환경명, port, health check URL
- webhook response body와 header
- AI agent가 읽은 내부 로그, 터미널 출력, config, 장애 분석 요약

`확인`: GitLab CI/CD variables 문서는 masked variable이 secret 유출 방지의 완전한 보장이 아니며, 외부 secrets provider와 file type variables를 고려하라고 설명합니다. [GitLab CI/CD variables](https://docs.gitlab.com/ci/variables/)

`확인`: GitLab Runner 보안 문서는 self-managed runner가 job에 정의된 코드를 실행하기 때문에 runner host와 network에 위험이 있고, shell executor와 privileged Docker executor가 특히 위험하다고 설명합니다. 또한 runner를 별도 network segment에 두는 방안을 권고합니다. [GitLab Runner security](https://docs.gitlab.com/runner/security/)

`추론`: AI agent를 넣으면 이 위험은 커집니다. AI agent는 로그를 요약하거나 실패 원인을 설명하는 과정에서 내부 서버명, 설정값, token 모양, 업무 데이터 샘플을 자연어로 재생산할 수 있습니다. 내부망에서 AI agent가 외부 LLM API나 연구/개발망 agent runtime으로 내부 로그를 보내면, 이는 명백한 내부망 외부 유출 후보입니다.

## 3. 데이터 흐름 표

| 출발 망 | 도착 망 | 이동 데이터 | 허용 여부 | 위험 | 통제 방법 |
|---|---|---|---|---|---|
| 연구/개발망 | 연구/개발망 GitLab | source code, merge request, pipeline config | 허용 후보 | source code에 운영 secret이 섞이면 이후 모든 경로로 확산 | secret scanning, protected branch, MR approval, production secret 저장 금지 |
| 연구/개발망 GitLab | DMZ webhook inbox | deploy trigger, project id, commit SHA, artifact digest | 허용 후보 | payload에 branch/user/internal env 정보가 과다 포함될 수 있음 | custom payload template, 최소 필드, mTLS, allowlist, `X-Gitlab-Token` 검증 |
| 연구/개발망 build runner | DMZ artifact quarantine | build artifact, checksum, SBOM, provenance, signature | 허용 후보 | 악성 artifact 또는 위조 artifact가 내부로 들어갈 수 있음 | immutable 저장, malware/secret scan, signature/checksum 검증, 승인 전 quarantine |
| 연구/개발망 | 내부망 | source repository clone | 원칙적 비권장 | 내부망에 전체 Git history와 개발 secret이 들어옴 | artifact 기반 배포 우선, 꼭 필요하면 read-only mirror와 history/secret 정책 |
| DMZ | 내부망 | 검증된 artifact manifest, artifact blob | 허용 후보 | DMZ compromise 시 변조 artifact 주입 | 내부망에서 digest/signature/provenance 재검증, DMZ를 trust root로 보지 않음 |
| DMZ | 내부망 | deploy command, SSH/API call | 검토 필요 | DMZ가 내부 서버에 명령권을 갖게 됨 | push 대신 pull 우선, push면 명령 allowlist, per-target 최소 권한, 내부 승인 |
| 내부망 agent | DMZ | poll request, artifact id, agent id | 검토 필요 | agent id와 target group이 내부 topology를 드러낼 수 있음 | 서버명 대신 group alias, coarse request, 응답 로그 제한, DMZ에 raw status 저장 금지 |
| 내부망 | 연구/개발망 GitLab | runner callback, job trace, deployment status, environment update | 금지 기본값 | 내부 서버 상태와 로그가 연구/개발망으로 역류 | 내부망 runner를 외부 GitLab에 등록하지 않음, GitLab deployment status 업데이트 금지 |
| 내부망 | 업무망 | 배포 결과, 운영 로그, 에러 메시지 | 금지 기본값 | 내부 운영 정보와 장애 정보가 업무망으로 유출 | 내부 전용 포털/SIEM, 외부 공유는 보안 승인된 요약만 |
| 내부망 | 외부 LLM/API | AI agent prompt, log summary, config, stack trace | 금지 기본값 | 내부 데이터가 외부 서비스로 유출 | 내부망 전용 model/runtime, redaction, outbound block, prompt audit |
| 업무망 | 연구/개발망 GitLab | 요구사항, approval comment, issue text | 정책 결정 필요 | 업무 데이터가 연구/개발망으로 이동 | 데이터 등급 분류, 업무망 전용 ticket mirror 또는 sanitized issue |
| GitLab webhook receiver | GitLab | webhook response body/header | 금지 또는 최소화 | response detail이 GitLab history에 저장될 수 있음 | 2xx + opaque event id만 반환, 내부 상태/에러 메시지 반환 금지 |
| 내부망 | 내부망 audit store | deploy ledger, server result, rollback log | 허용 | 감사 로그 자체가 민감 정보 저장소가 됨 | WORM/append-only, RBAC, 보존 기간, 검색 권한 분리 |

## 4. 가능한 아키텍처 대안 3개 이상

### 대안 A. DMZ push 중계형

GitLab webhook이 DMZ 서버로 들어오고, DMZ 서버가 내부 서버에 SSH, HTTP API, WinRM, 배포 agent API 등으로 배포 명령을 내립니다.

장점은 구조가 단순하고 개발자가 보기에는 merge 후 바로 배포되는 흐름을 만들기 쉽다는 점입니다. 단점은 DMZ가 내부망에 대한 명령권과 credential을 갖게 된다는 점입니다. DMZ가 단순 proxy가 아니라 orchestrator처럼 변하면 서버 목록, secret, rollback 상태, 배포 결과를 DMZ가 저장하게 됩니다.

이 방식은 PoC의 가장 쉬운 출발점처럼 보이지만, 현재 제약에서는 보수적으로 보아야 합니다. DMZ가 내부망으로 command를 push하는 순간, DMZ compromise가 내부 compromise로 이어질 가능성이 커집니다.

### 대안 B. 내부망 pull agent형

GitLab은 DMZ에 배포 후보 manifest와 signed artifact만 남기고, 내부망 deploy agent가 정해진 주기로 DMZ artifact repository 또는 DMZ inbox에서 가져옵니다. 배포 대상, 실제 승인, 배포 결과, rollback 상태는 내부망에서만 관리합니다.

이 방식의 핵심은 “밖에서 내부를 때리는 배포 명령”이 아니라 “내부가 검증된 후보를 가져와 내부 정책으로 배포한다”입니다. 내부망에서 DMZ로 나가는 poll request 자체도 데이터 흐름이지만, server name과 raw status를 담지 않도록 줄일 수 있습니다.

### 대안 C. Jenkins controller/agent 분리형

Jenkins controller를 어디에 두느냐가 핵심입니다.

- controller가 연구/개발망 또는 업무망에 있고 내부망 agent가 붙으면, 내부망 agent의 build/deploy log와 status가 controller로 돌아가므로 현재 금지 흐름과 충돌합니다.
- controller가 DMZ에 있으면 내부망 agent가 DMZ로 상태를 되돌립니다. 이 흐름은 연구/개발망/업무망보다는 낫지만, DMZ가 내부 배포 로그와 credential의 집합점이 될 수 있습니다.
- controller가 내부망에 있으면 배포 로그와 secret은 내부망에 머물 수 있습니다. 대신 GitLab webhook은 내부 controller를 직접 호출하지 말고 DMZ inbox나 artifact promotion 절차를 통해 내부 controller가 pull하도록 만드는 편이 안전합니다.

### 대안 D. GitLab Runner 분리형

GitLab Runner를 연구/개발망에 두면 source build/test에는 자연스럽습니다. 그러나 내부망 배포 runner를 연구/개발망 GitLab에 등록하면 runner가 GitLab과 job request, job token, trace, status를 주고받으므로 내부망에서 연구/개발망으로 데이터가 흐릅니다.

따라서 GitLab Runner는 아래처럼 분리해야 합니다.

- 연구/개발망 runner: build/test/sign/SBOM 생성까지만 수행
- DMZ runner: 가능하면 검증/복사까지만 수행하고 내부 배포 secret 보관 금지
- 내부망 runner: 내부 GitLab mirror 또는 내부 CI controller에만 붙고, 연구/개발망 GitLab로 trace/status를 보내지 않음

### 대안 E. Artifact promotion형

연구/개발망에서 artifact 후보를 만들고, DMZ quarantine에서 metadata와 서명을 검증한 뒤, 내부망 artifact mirror로 승격합니다. 내부 배포는 내부 orchestrator가 내부 mirror에서 가져갑니다.

이 방식은 운영 복잡도는 높지만, 데이터 경계가 가장 명확합니다. source code, artifact, deployment state, credential, audit log의 위치를 서로 분리할 수 있기 때문입니다.

`확인`: SLSA는 artifact와 provenance를 검증할 때 builder identity, provenance signature, artifact digest, canonical source repository, buildType, externalParameters를 확인하라고 설명합니다. [SLSA verifying artifacts](https://slsa.dev/spec/v1.2/verifying-artifacts)

`확인`: Sigstore Cosign은 container image나 blob signature를 검증하고, signature payload가 image digest와 일치하는지 확인할 수 있습니다. [Sigstore Cosign verify](https://docs.sigstore.dev/cosign/verifying/verify/)

### 대안 F. 내부 GitOps/pull controller형

Kubernetes나 유사한 declarative deployment를 쓴다면 내부망 GitOps controller가 내부 mirror의 manifest와 artifact를 pull하고 reconcile합니다. 단, Git repository가 연구/개발망에 있으면 controller가 outbound로 repo를 읽어야 하므로, 내부망에서 연구/개발망으로 request metadata가 나갑니다. 이때도 내부망 전용 config mirror 또는 DMZ/내부 artifact mirror를 두는 편이 안전합니다.

`확인`: GitOps agent는 repo source에 outbound TCP가 필요하다고 문서화된 사례가 있습니다. [Azure Arc GitOps with Argo CD network requirements](https://learn.microsoft.com/en-us/azure/azure-arc/kubernetes/tutorial-use-gitops-argocd)

## 5. 각 대안의 비교표

| 대안 | 보안성 | 운영 복잡도 | 장애 복구성 | 확장성 | 개발자 편의성 | 감사 가능성 | 판단 |
|---|---|---:|---:|---:|---:|---:|---|
| A. DMZ push 중계형 | 낮음~중간 | 낮음 | 중간 | 중간 | 높음 | 중간 | 빠른 PoC는 가능하지만 DMZ 명령권이 커져 비권장 |
| B. 내부망 pull agent형 | 높음 | 중간 | 높음 | 높음 | 중간 | 높음 | 현재 제약에 가장 잘 맞는 기본형 |
| C. Jenkins controller/agent 분리형 | 배치에 따라 다름 | 중간~높음 | 중간 | 중간 | 중간 | 중간~높음 | controller는 내부망에 둘 때만 안전성이 올라감 |
| D. GitLab Runner 분리형 | 배치에 따라 다름 | 중간 | 중간 | 높음 | 높음 | GitLab 중심이면 높음, 내부망 제약에서는 제한 | 내부망 runner가 외부 GitLab에 붙는 구조는 금지 기본값 |
| E. Artifact promotion형 | 높음 | 높음 | 높음 | 높음 | 중간 | 높음 | 권장. 보안/감사/확장성의 균형이 좋음 |
| F. 내부 GitOps/pull controller형 | 높음 | 중간~높음 | 높음 | 높음 | 중간 | 높음 | K8s/declarative 환경이면 강력. repo/status 역류 통제 필요 |

## 6. 권장 아키텍처

권장안은 **Artifact promotion형을 중심으로 내부망 pull agent형을 결합한 구조**입니다.

1. 연구/개발망

    GitLab은 source code, merge request, pipeline definition, build candidate를 관리합니다. AI 개발 agent는 여기서 코드 작성, 테스트 생성, 문서화, 리뷰 보조를 수행할 수 있지만, 내부 운영 로그와 production secret에는 접근하지 않습니다.

2. 연구/개발망 build runner

    build/test를 수행하고, artifact, checksum, SBOM, provenance, signature를 생성합니다. production deploy key, internal SSH key, 내부 서버 목록은 주입하지 않습니다.

3. DMZ

    DMZ는 `webhook inbox`와 `artifact quarantine`으로 시작합니다. GitLab webhook secret과 mTLS로 event source를 확인하고, payload는 project, commit SHA, artifact digest, signature reference 정도로 줄입니다. DMZ는 내부 서버 목록과 rollback state를 갖지 않습니다.

4. 내부망

    내부망 deploy orchestrator가 DMZ quarantine 또는 내부 mirror에서 artifact 후보를 가져와 다시 검증합니다. 실제 대상 서버 선택, 승인, 배포 실행, 실패 처리, rollback, 감사 로그는 내부망에 남깁니다.

5. 감사와 상태

    GitLab에는 merge, pipeline, build artifact 후보, release evidence까지만 남깁니다. 내부망에는 “누가, 언제, 어떤 commit/artifact digest를, 어떤 내부 서버 group에, 어떤 결과로 배포했는지”를 남깁니다. 내부 서버명과 raw log는 외부로 내보내지 않습니다.

### 자산별 권장 위치

| 자산 | 권장 위치 | 이유 |
|---|---|---|
| Source code | 연구/개발망 GitLab | 개발 협업과 AI 개발 자동화의 중심. 단, 운영 secret과 내부 topology 금지 |
| Build artifact 후보 | 연구/개발망 artifact registry 또는 DMZ quarantine | 내부 배포 전 검증 대상 |
| 배포 가능 artifact | 내부망 artifact mirror | 내부망 배포 소비자가 외부 repo를 직접 신뢰하지 않도록 함 |
| 배포 설정 | 일반 설정은 Git, 내부 대상/서버 topology는 내부망 | source와 운영 topology를 분리 |
| Secret | 내부망 vault/KMS/HSM | production secret이 연구/개발망/업무망/DMZ로 번지는 것을 차단 |
| Build log | 연구/개발망 | 내부 서버 정보 없는 build/test 로그만 |
| Deploy log/status | 내부망 audit/log store | hard constraint상 외부 역류 금지 |
| 승인 기록 | 내부망 또는 정책상 승인된 별도 시스템 | 승인 주체와 내부 결과를 분리해 저장 |

### 자격증명 정책

| 자격증명 | 권장 보관 위치 | 최소 권한 | 회전/폐기 |
|---|---|---|---|
| Webhook secret | GitLab 설정 + DMZ webhook receiver | event source 검증 전용 | 정기 회전, dual-secret 전환, 유출 시 즉시 폐기 |
| Deploy token | artifact repo별 read-only token | `read_package_registry` 또는 `read_registry`처럼 필요한 read scope만 | 만료일 필수, project/group 범위 제한 |
| SSH key | 내부망 deploy orchestrator 또는 내부 vault | 대상 서버 group별 forced command 또는 제한 계정 | 주기적 교체, 서버 group 분리, 미사용 key 폐기 |
| Runner token | runner host local config | runner 인증 전용 | runner 단위 rotate/revoke, clone 탐지 |
| Signing key | KMS/HSM 또는 서명 전용 서비스 | build step이 private key material을 직접 읽지 못하게 분리 | key id 추적, dual-sign migration, revoke list |
| mTLS certificate | 각 boundary component local keystore | endpoint identity 확인 | 내부 CA, 짧은 유효기간, 자동 갱신 |

`확인`: GitLab deploy token은 repository/package/container registry 접근에 쓸 수 있고 scope를 선택합니다. GitLab 문서는 deploy token이 long-lived라 공격자에게 매력적이라고 경고합니다. [GitLab deploy tokens](https://docs.gitlab.com/user/project/deploy_tokens/)

`확인`: GitLab runner authentication token은 runner machine의 `config.toml`에 저장되고, runner가 GitLab에서 job을 받을 때 사용합니다. 파일 시스템 compromise가 token 노출로 이어질 수 있습니다. [GitLab token overview](https://docs.gitlab.com/security/tokens/)

### Artifact 검증과 승인 위치

Artifact 검증은 한 번으로 끝내지 않고, **생산 지점, DMZ quarantine, 내부 소비 지점**에서 나누어 봅니다.

1. Build 단계

    test, dependency scan, SBOM 생성, checksum 생성, provenance 생성, signing을 수행합니다. signing key는 build script가 직접 읽는 환경 변수보다 KMS/HSM 또는 서명 전용 서비스에 두는 편이 안전합니다.

2. DMZ quarantine 단계

    webhook payload schema, artifact digest, signature, SBOM 존재 여부, secret scan, malware scan을 확인합니다. 이 단계는 내부 배포 승인 전 “후보 검문소”이지 최종 신뢰 근거가 아닙니다.

3. 내부망 promotion 단계

    내부망 verifier가 artifact digest, signature, provenance, builder identity, allowed source repository, allowed branch/tag, SBOM/vulnerability policy를 다시 확인합니다. 이 단계에서 PASS한 artifact만 internal artifact mirror에 들어갑니다.

4. 배포 승인 단계

    승인자는 내부망 포털에서 artifact digest와 검증 결과를 보고 승인합니다. GitLab protected environment approval을 쓰고 싶다면 GitLab이 어느 망에 있는지 다시 봐야 합니다. 연구/개발망 GitLab에 내부 deployment approval/status가 남으면 현재 제약과 충돌할 수 있습니다.

## 7. 비권장 아키텍처와 이유

### 내부망 GitLab Runner가 연구/개발망 GitLab에 직접 등록되는 구조

비권장입니다. Runner는 job을 받기 위해 GitLab과 통신하고, job trace/status를 되돌리는 것이 일반 동작입니다. 내부망 배포 job이 실패하면 실패 로그, 서버 접근 에러, 환경 변수명, artifact path, 내부 endpoint가 GitLab에 남을 수 있습니다.

### 업무망 Jenkins controller + 내부망 deploy agent 구조

비권장입니다. Jenkins controller가 업무망에 있으면 내부망 agent의 log/status가 업무망으로 돌아갑니다. 현재 hard constraint가 “내부망에서 업무망으로 데이터가 흘러가면 안 됨”이라면, 이 구조는 기본적으로 충돌합니다.

### DMZ에 production deploy secret을 모아 두는 push orchestrator 구조

비권장입니다. DMZ는 외부와 내부가 만나는 경계이므로 공격 노출면이 큽니다. DMZ가 내부 서버 SSH key, 서버 목록, rollback 상태를 모두 갖는 순간, DMZ compromise는 내부 배포 권한 compromise에 가까워집니다.

### GitLab environment/deployment status를 내부 배포 결과의 system of record로 쓰는 구조

연구/개발망 GitLab을 쓴다면 비권장입니다. GitLab deployments API는 deployed commit SHA와 deployment status를 다룰 수 있지만, 내부망 배포 결과를 GitLab에 update하면 내부망에서 연구/개발망으로 상태 데이터가 이동합니다. [GitLab Deployments API](https://docs.gitlab.com/api/deployments/)

### AI agent가 내부 로그를 연구/개발망 또는 외부 LLM으로 보내 분석하는 구조

비권장입니다. AI agent의 prompt와 tool output은 로그보다 더 압축된 유출물이 될 수 있습니다. 내부망 장애 원인 요약에는 서버명, endpoint, token prefix, 계정명, stack trace, 업무 데이터 샘플이 섞일 수 있습니다.

## 8. 최소 PoC 구성안

PoC는 “편하게 전부 자동 배포”가 아니라, **데이터 흐름이 실제로 통제되는지 검증하는 최소 구조**로 잡는 편이 좋습니다.

1. 연구/개발망 GitLab

    - protected branch와 MR approval 설정
    - build/test runner는 연구/개발망에만 배치
    - production secret, 내부 SSH key, 내부 서버 목록은 GitLab variable에 저장하지 않음

2. Build pipeline

    - artifact 생성
    - checksum 생성
    - SBOM 생성
    - artifact signature 생성
    - 배포 후보 manifest 생성

3. DMZ webhook inbox

    - `X-Gitlab-Token` 검증
    - source IP/host allowlist
    - event idempotency 저장
    - payload 최소화 검증
    - response body에는 내부 상태를 절대 넣지 않음

4. DMZ artifact quarantine

    - artifact upload 또는 pull-through 저장
    - immutable path: `<project>/<commit>/<artifact-digest>`
    - secret scan과 signature verification 결과 저장
    - 내부 서버명이나 target list 저장 금지

5. 내부망 deploy verifier/agent

    - DMZ에서 manifest와 artifact를 가져옴
    - 내부 public key/root of trust로 signature/provenance 검증
    - allowed source repository와 branch/tag 확인
    - 내부 승인 후 단일 test server group에만 배포
    - deploy result는 내부 audit log에만 기록

6. 실패/복구 PoC

    - webhook 중복 수신: 같은 event id는 한 번만 처리
    - artifact digest 불일치: 배포 차단
    - signature 실패: 배포 차단
    - 일부 서버 실패: 성공/실패 서버를 분리 기록하고 실패 서버만 재시도
    - DMZ 장애: 내부 mirror에 이미 승격된 artifact는 계속 배포 가능
    - rollback: 이전 artifact digest로 재배포

7. PoC PASS 기준

    - 내부망 deploy log/status가 연구/개발망 GitLab과 업무망에 남지 않는다.
    - GitLab webhook history에 내부 서버명, 내부 error, 내부 endpoint가 남지 않는다.
    - DMZ에는 production SSH key와 내부 target inventory가 없다.
    - 내부망 verifier가 signature/digest 오류를 실제로 차단한다.
    - 같은 webhook event를 두 번 보내도 배포가 중복 실행되지 않는다.

## 9. 실제 도입 전 반드시 확인해야 할 질문 목록

1. Source code의 보안 등급은 무엇인가?

    연구/개발망 GitLab에 source code를 두어도 되는지부터 확정해야 합니다. 코드 안에 업무 데이터 샘플, 내부 endpoint, 운영 설정, secret이 포함된다면 GitLab 위치 판단이 달라집니다.

2. “내부망에서 연구/개발망/업무망으로 데이터가 흘러가면 안 된다”의 예외가 있는가?

    배포 성공/실패 같은 coarse status도 금지인지, 보안팀이 승인한 요약 이벤트는 가능한지 정책을 먼저 정해야 합니다.

3. DMZ는 어떤 권한을 가질 수 있는가?

    단순 relay, artifact quarantine, 검증 지점, orchestrator 중 어디까지 허용할지 결정해야 합니다.

4. 내부망에서 DMZ로 나가는 poll request는 허용되는가?

    허용된다면 어떤 metadata까지 가능한지 정해야 합니다. agent id, server group, artifact id도 정보입니다.

5. 내부 배포 대상 서버 목록은 어디에서 관리하는가?

    Git repo에 둘지, 내부 CMDB에 둘지, deploy orchestrator DB에 둘지 결정해야 합니다. 현재 제약에서는 내부망 자산으로 두는 편이 안전합니다.

6. Artifact repository는 어디에 둘 것인가?

    연구/개발망 registry, DMZ quarantine, 내부 mirror를 어떤 제품으로 구성하고, replication 또는 import/export를 누가 수행할지 정해야 합니다.

7. Signing key와 verification trust root는 누가 운영하는가?

    build signing key, 내부 verification public key, rotation, revoke, dual-sign 기간을 운영 정책으로 잡아야 합니다.

8. 취약점 scan 실패 기준은 무엇인가?

    CVSS 기준, exploitability, allowlist, exception approval, emergency deploy 정책을 정해야 합니다.

9. AI agent는 어느 망에서 어떤 tool을 쓸 수 있는가?

    연구/개발망 agent는 source와 test까지, 내부망 agent는 배포 검증과 내부 로그 분석까지처럼 role을 분리해야 합니다. 외부 LLM API 사용 가능 여부는 별도 보안 결정입니다.

10. 감사 로그의 보존 기간과 조회 권한은 무엇인가?

    내부 deploy ledger는 보안팀, 인프라팀, 개발팀이 모두 필요로 하지만, raw log와 내부 topology는 최소 권한으로 분리해야 합니다.

11. 장애 시 수동 개입 권한은 누가 갖는가?

    DMZ 장애, 내부 agent 장애, 일부 서버 실패, rollback 필요 상황에서 누가 override할 수 있는지와 그 override가 어디에 기록되는지 정해야 합니다.

12. 방화벽과 네트워크 ACL은 연결 방향 기준으로 문서화되어 있는가?

    `GitLab -> DMZ`, `내부 agent -> DMZ`, `내부 orchestrator -> target`, `DMZ -> 내부` 각각의 port, protocol, 인증, 목적, 로그 정책을 분리해야 합니다.

13. DNS와 인증서는 망별로 어떻게 분리되는가?

    내부 DNS 이름이 GitLab webhook response나 DMZ log에 노출되지 않아야 합니다. mTLS certificate의 발급 CA와 폐기 절차도 정해야 합니다.

14. GitLab release evidence, audit events, protected environment를 어디까지 쓸 것인가?

    GitLab이 연구/개발망에 있으면 내부 배포 결과의 감사 원장으로 쓰기는 어렵습니다. 대신 build/release 후보 evidence까지만 쓰는 방안을 검토해야 합니다. [GitLab release evidence](https://docs.gitlab.com/user/project/releases/release_evidence/), [GitLab audit events](https://docs.gitlab.com/user/compliance/audit_events/)

## 10. 최종 권고

최종 권고는 **연구/개발망 GitLab + DMZ artifact quarantine + 내부망 artifact mirror + 내부망 pull deploy agent/orchestrator**입니다. 이 구조는 개발자 경험을 완전히 희생하지 않으면서도, 내부망 배포 결과가 연구/개발망이나 업무망으로 되돌아가는 경로를 기본 차단합니다.

도구 선택은 다음 순서로 결정하는 편이 좋습니다.

1. 먼저 데이터 흐름 정책을 확정합니다.

    내부망에서 연구/개발망/업무망으로 어떤 데이터도 나가지 않는다는 정책이 유지된다면, GitLab Runner나 Jenkins agent가 외부 controller에 붙는 구조는 배포 단계에서 제외해야 합니다.

2. 다음으로 artifact promotion과 검증 모델을 확정합니다.

    배포 단위는 source clone이 아니라 signed artifact + SBOM + checksum + provenance가 되어야 합니다. 내부망은 DMZ나 GitLab을 trust root로 보지 말고, 소비 지점에서 다시 검증해야 합니다.

3. 그 다음 배포 실행 방식을 고릅니다.

    내부망 pull agent가 기본값입니다. Jenkins나 GitLab Runner를 쓰더라도 controller는 내부망에 두거나, 최소한 내부망 결과가 외부 controller로 올라가지 않는 구조로 제한해야 합니다.

4. 마지막으로 개발자 경험을 설계합니다.

    개발자는 merge 후 build 후보 생성까지는 GitLab에서 확인합니다. 내부 배포 승인과 실제 결과는 내부망 포털에서 확인합니다. 외부 가시성이 필요하다면 보안팀이 승인한 수준의 sanitized status만 별도 정책으로 설계해야 하며, 기본값은 외부 공유 금지입니다.

이 권고는 “webhook을 날리면 된다”가 아니라, **어떤 데이터가 어디로 이동하는지**를 먼저 고정하는 설계입니다. 이 순서를 지켜야 나중에 Jenkins, GitLab Runner, artifact repository, AI agent, GitOps 도구 중 무엇을 붙이더라도 같은 보안 경계를 유지할 수 있습니다.

참고한 주요 공식/제품 문서:

- [NIST SP 800-207 Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [GitLab CI/CD development guidelines: runner scheduling and communication](https://docs.gitlab.com/development/cicd/)
- [GitLab Webhooks](https://docs.gitlab.com/user/project/integrations/webhooks/)
- [GitLab CI/CD variables](https://docs.gitlab.com/ci/variables/)
- [GitLab Runner security](https://docs.gitlab.com/runner/security/)
- [GitLab token overview](https://docs.gitlab.com/security/tokens/)
- [GitLab deploy tokens](https://docs.gitlab.com/user/project/deploy_tokens/)
- [GitLab environments](https://docs.gitlab.com/ci/environments/)
- [GitLab protected environments](https://docs.gitlab.com/ci/environments/protected_environments/)
- [GitLab deployments API](https://docs.gitlab.com/api/deployments/)
- [GitLab audit events](https://docs.gitlab.com/user/compliance/audit_events/)
- [GitLab release evidence](https://docs.gitlab.com/user/project/releases/release_evidence/)
- [Jenkins exposed services and ports](https://www.jenkins.io/doc/book/security/services/)
- [Jenkins using credentials](https://www.jenkins.io/doc/book/using/using-credentials/)
- [Jenkins credentials masking limitations](https://www.jenkins.io/blog/2019/02/21/credentials-masking/)
- [SLSA Build Provenance](https://slsa.dev/spec/v1.2-rc2/build-provenance)
- [SLSA Verifying Artifacts](https://slsa.dev/spec/v1.2/verifying-artifacts)
- [Sigstore Cosign verifying signatures](https://docs.sigstore.dev/cosign/verifying/verify/)
- [NTIA SBOM Minimum Elements](https://www.ntia.gov/files/ntia/publications/sbom_minimum_elements_report.pdf)
- [Azure Arc GitOps with Argo CD network requirements](https://learn.microsoft.com/en-us/azure/azure-arc/kubernetes/tutorial-use-gitops-argocd)
- [Red Hat OpenShift disconnected environments](https://docs.redhat.com/en/documentation/openshift_container_platform/4.21/html-single/disconnected_environments/index)

## 11. 회의 공유용 서버 구조도 HTML 파일

아래 HTML은 외부 CDN, 외부 이미지, 외부 CSS/JS 없이 열 수 있는 단일 파일입니다. 같은 내용은 별도 파일 [`ai/network_separated_ai_agent_cicd.html`](network_separated_ai_agent_cicd.html)에도 저장했습니다.

```html
<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>망 분리 환경 AI Agent / CI-CD 배포 구조 검토</title>
  <style>
    :root {
      --bg: #f6f8fb;
      --ink: #182230;
      --muted: #5f6c7b;
      --line: #d8dee8;
      --dev: #e8f3ff;
      --biz: #f3ecff;
      --dmz: #fff4d6;
      --int: #e9f8ef;
      --ok: #177245;
      --ok2: #1b62b6;
      --warn: #b86b00;
      --bad: #c93535;
      --card: #ffffff;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      background: var(--bg);
      color: var(--ink);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans KR", Arial, sans-serif;
      line-height: 1.52;
    }

    header {
      padding: 26px 32px 18px;
      background: #172033;
      color: #fff;
    }

    h1 {
      margin: 0 0 8px;
      font-size: 25px;
      letter-spacing: 0;
    }

    h2 {
      margin: 26px 0 12px;
      font-size: 18px;
    }

    h3 {
      margin: 16px 0 8px;
      font-size: 15px;
    }

    p {
      margin: 7px 0;
    }

    main {
      padding: 22px 28px 36px;
      max-width: 1600px;
      margin: 0 auto;
    }

    .summary {
      display: grid;
      grid-template-columns: repeat(4, minmax(220px, 1fr));
      gap: 10px;
      margin: 16px 0 18px;
    }

    .pill {
      border-radius: 8px;
      border: 1px solid var(--line);
      background: var(--card);
      padding: 12px 14px;
    }

    .pill strong {
      display: block;
      margin-bottom: 4px;
      font-size: 13px;
    }

    .ok {
      border-left: 5px solid var(--ok);
    }

    .warn {
      border-left: 5px solid var(--warn);
    }

    .bad {
      border-left: 5px solid var(--bad);
    }

    .info {
      border-left: 5px solid var(--ok2);
    }

    .legend {
      display: flex;
      gap: 14px;
      flex-wrap: wrap;
      align-items: center;
      margin: 10px 0 16px;
      color: var(--muted);
      font-size: 13px;
    }

    .legend span {
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }

    .swatch {
      width: 34px;
      height: 0;
      border-top: 4px solid var(--ok);
    }

    .swatch.blue {
      border-color: var(--ok2);
    }

    .swatch.warn {
      border-color: var(--warn);
    }

    .swatch.bad {
      border-color: var(--bad);
      border-top-style: dashed;
    }

    .map {
      position: relative;
      min-width: 1180px;
      min-height: 392px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      overflow: hidden;
    }

    .lanes {
      position: absolute;
      inset: 0;
      display: grid;
      grid-template-columns: repeat(4, 1fr);
    }

    .zone {
      padding: 14px 16px;
      border-right: 1px solid var(--line);
    }

    .zone:last-child {
      border-right: 0;
    }

    .zone h3 {
      margin: 0 0 10px;
      font-size: 15px;
    }

    .dev {
      background: var(--dev);
    }

    .biz {
      background: var(--biz);
    }

    .dmz {
      background: var(--dmz);
    }

    .internal {
      background: var(--int);
    }

    .node {
      width: 245px;
      min-height: 46px;
      margin: 10px 0;
      padding: 9px 10px;
      border: 1px solid #aab4c2;
      border-radius: 8px;
      background: rgba(255,255,255,0.86);
      box-shadow: 0 1px 0 rgba(20, 32, 48, 0.05);
      font-size: 13px;
    }

    .node strong {
      display: block;
      font-size: 13px;
    }

    .node small {
      color: var(--muted);
    }

    svg.arrows {
      position: absolute;
      inset: 0;
      pointer-events: none;
    }

    .flow-label {
      font-size: 12px;
      font-weight: 700;
      fill: #172033;
      paint-order: stroke;
      stroke: #fff;
      stroke-width: 4px;
      stroke-linejoin: round;
    }

    .callout-grid {
      display: grid;
      grid-template-columns: repeat(3, minmax(280px, 1fr));
      gap: 12px;
      margin: 14px 0;
    }

    .card {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 14px;
    }

    table {
      width: 100%;
      border-collapse: collapse;
      background: var(--card);
      border: 1px solid var(--line);
      font-size: 13px;
    }

    th, td {
      border: 1px solid var(--line);
      padding: 9px 10px;
      vertical-align: top;
    }

    th {
      background: #edf1f7;
      text-align: left;
    }

    .tag {
      display: inline-block;
      border-radius: 999px;
      padding: 2px 8px;
      font-size: 12px;
      font-weight: 700;
      white-space: nowrap;
    }

    .tag.ok {
      color: #0f5c35;
      background: #dcf4e6;
      border: 0;
    }

    .tag.warn {
      color: #7b4900;
      background: #ffe7b8;
      border: 0;
    }

    .tag.bad {
      color: #8a2020;
      background: #ffe1df;
      border: 0;
    }

    ul, ol {
      padding-left: 22px;
    }

    li {
      margin: 5px 0;
    }

    .wide {
      overflow-x: auto;
      padding-bottom: 6px;
    }

    footer {
      margin-top: 22px;
      color: var(--muted);
      font-size: 12px;
    }

    @media print {
      body {
        background: #fff;
        color: #000;
      }

      header {
        background: #fff;
        color: #000;
        border-bottom: 2px solid #000;
      }

      main {
        padding: 14px;
        max-width: none;
      }

      .summary,
      .callout-grid {
        grid-template-columns: repeat(2, 1fr);
      }

      .map {
        transform: scale(0.78);
        transform-origin: top left;
        width: 1280px;
        margin-bottom: -70px;
        break-inside: avoid;
      }

      h2, h3, .card, table {
        break-inside: avoid;
      }
    }
  </style>
</head>
<body>
  <header>
    <h1>망 분리 환경 AI Agent / CI-CD 배포 자동화 구조</h1>
    <p>목표는 개발 자동화의 속도를 얻되, 내부망에서 연구/개발망 또는 업무망으로 흘러가는 데이터 흐름을 기본 차단하는 것입니다.</p>
  </header>

  <main>
    <section class="summary">
      <div class="pill ok">
        <strong>권장 구조</strong>
        연구/개발망 GitLab은 source/build 후보까지만 담당하고, 내부 배포 판단과 결과 로그는 내부망에 둡니다.
      </div>
      <div class="pill bad">
        <strong>금지 기본값</strong>
        내부망 Runner/Agent가 연구/개발망 GitLab 또는 업무망 Jenkins로 로그, 상태, 메타데이터를 되돌리는 구조입니다.
      </div>
      <div class="pill warn">
        <strong>DMZ 역할</strong>
        DMZ는 명령권자가 아니라 검증된 artifact 후보를 받는 quarantine/inbox로 시작하는 편이 안전합니다.
      </div>
      <div class="pill info">
        <strong>핵심 관점</strong>
        배포 성공/실패, 로그, token, 서버 목록, 에러 메시지, artifact metadata도 모두 데이터 흐름입니다.
      </div>
    </section>

    <section>
      <h2>범례</h2>
      <div class="legend">
        <span><i class="swatch blue"></i>허용 흐름: 검증된 source/artifact/trigger가 안쪽으로 이동</span>
        <span><i class="swatch warn"></i>검토 필요: 내부 메타데이터가 DMZ까지 나갈 수 있음</span>
        <span><i class="swatch bad"></i>금지 흐름: 내부망 정보가 연구/개발망 또는 업무망으로 역류</span>
      </div>
    </section>

    <section>
      <h2>권장 구조: Artifact Promotion + 내부망 Pull Agent</h2>
      <div class="wide">
        <div class="map" aria-label="network architecture map">
          <div class="lanes">
            <div class="zone dev">
              <h3>연구/개발망</h3>
              <div class="node"><strong>GitLab</strong><small>source code, merge request, CI pipeline</small></div>
              <div class="node"><strong>Build Runner</strong><small>test, build, SBOM, checksum, signing</small></div>
              <div class="node"><strong>AI 개발 Agent</strong><small>코드/문서 보조. 내부 로그와 운영 secret 접근 금지</small></div>
            </div>
            <div class="zone biz">
              <h3>업무망</h3>
              <div class="node"><strong>업무 사용자</strong><small>요청, 변경 승인 후보. 내부 배포 로그 수신 금지</small></div>
              <div class="node"><strong>협업 포털</strong><small>정책 승인 화면은 가능하나 내부망 결과 표시 금지</small></div>
            </div>
            <div class="zone dmz">
              <h3>DMZ</h3>
              <div class="node"><strong>Webhook Inbox</strong><small>secret 검증, idempotency key, allowlist</small></div>
              <div class="node"><strong>Quarantine Artifact Repo</strong><small>signed artifact, SBOM, checksum, provenance</small></div>
              <div class="node"><strong>검증 지점</strong><small>서명/무결성/취약점 정책 확인. 배포 secret 없음</small></div>
            </div>
            <div class="zone internal">
              <h3>내부망</h3>
              <div class="node"><strong>Internal Artifact Mirror</strong><small>승격된 artifact와 manifest</small></div>
              <div class="node"><strong>Deploy Orchestrator</strong><small>대상 서버 목록, 승인, 재시도, rollback</small></div>
              <div class="node"><strong>Pull Deploy Agents</strong><small>서버별 배포 수행. 결과는 내부 감사 로그에만 기록</small></div>
              <div class="node"><strong>Target Servers</strong><small>운영/검증/배치 서버</small></div>
              <div class="node"><strong>Internal Audit Log</strong><small>누가, 언제, 어떤 artifact를 어디에 배포했는지</small></div>
            </div>
          </div>
          <svg class="arrows" viewBox="0 0 1200 392" preserveAspectRatio="none">
            <defs>
              <marker id="arrowBlue" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#1b62b6"></path>
              </marker>
              <marker id="arrowGreen" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#177245"></path>
              </marker>
              <marker id="arrowWarn" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#b86b00"></path>
              </marker>
              <marker id="arrowBad" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto">
                <path d="M0,0 L0,6 L8,3 z" fill="#c93535"></path>
              </marker>
            </defs>
            <line x1="245" y1="92" x2="610" y2="92" stroke="#1b62b6" stroke-width="4" marker-end="url(#arrowBlue)"></line>
            <text x="333" y="78" class="flow-label">deploy trigger + minimal metadata</text>
            <line x1="250" y1="178" x2="620" y2="178" stroke="#177245" stroke-width="4" marker-end="url(#arrowGreen)"></line>
            <text x="346" y="164" class="flow-label">signed artifact + SBOM + checksum</text>
            <line x1="765" y1="178" x2="922" y2="178" stroke="#177245" stroke-width="4" marker-end="url(#arrowGreen)"></line>
            <text x="780" y="164" class="flow-label">verified artifact promotion</text>
            <line x1="1014" y1="212" x2="1014" y2="292" stroke="#177245" stroke-width="4" marker-end="url(#arrowGreen)"></line>
            <text x="1025" y="254" class="flow-label">deploy command 내부 전용</text>
            <line x1="927" y1="226" x2="760" y2="226" stroke="#b86b00" stroke-width="4" marker-end="url(#arrowWarn)"></line>
            <text x="760" y="214" class="flow-label">poll request: artifact id / agent id</text>
            <line x1="760" y1="260" x2="930" y2="260" stroke="#b86b00" stroke-width="4" marker-end="url(#arrowWarn)"></line>
            <text x="775" y="248" class="flow-label">DMZ push command는 검토 필요</text>
            <line x1="965" y1="330" x2="252" y2="330" stroke="#c93535" stroke-width="4" stroke-dasharray="9 7" marker-end="url(#arrowBad)"></line>
            <text x="420" y="316" class="flow-label">금지: internal log/status/token/server list to 연구/개발망</text>
            <line x1="940" y1="350" x2="488" y2="350" stroke="#c93535" stroke-width="4" stroke-dasharray="9 7" marker-end="url(#arrowBad)"></line>
            <text x="565" y="378" class="flow-label">금지: internal deploy result/error/log to 업무망</text>
          </svg>
        </div>
      </div>
    </section>

    <section class="callout-grid">
      <div class="card">
        <h3>권장 구조</h3>
        <p>GitLab은 연구/개발망의 source와 build 후보 관리에 둡니다. DMZ는 webhook inbox와 artifact quarantine으로 제한하고, 내부망 deploy orchestrator와 pull agent가 최종 배포와 rollback을 소유합니다.</p>
      </div>
      <div class="card">
        <h3>비권장 구조</h3>
        <p>내부망 GitLab Runner가 연구/개발망 GitLab에 등록되어 job trace, 상태, 로그를 되돌리는 구조는 내부 데이터 역류 경로가 생깁니다. Jenkins controller가 외부/업무망에 있고 내부 agent가 붙는 구조도 같은 문제가 있습니다.</p>
      </div>
      <div class="card">
        <h3>숨은 데이터 흐름</h3>
        <p>실패 메시지, stack trace, 배포 대상 호스트명, agent 식별자, 환경 변수명, artifact download URL, runner token, SBOM의 내부 dependency명도 데이터입니다. 기본값은 내부망 밖으로 내보내지 않는 것입니다.</p>
      </div>
    </section>

    <section>
      <h2>데이터 흐름 요약</h2>
      <table>
        <thead>
          <tr>
            <th>출발</th>
            <th>도착</th>
            <th>데이터</th>
            <th>판정</th>
            <th>통제</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>연구/개발망</td>
            <td>DMZ</td>
            <td>webhook trigger, commit SHA, artifact digest</td>
            <td><span class="tag ok">허용 후보</span></td>
            <td>mTLS, webhook secret, payload 최소화, idempotency key, allowlist</td>
          </tr>
          <tr>
            <td>연구/개발망</td>
            <td>DMZ</td>
            <td>signed build artifact, SBOM, checksum, provenance</td>
            <td><span class="tag ok">허용 후보</span></td>
            <td>quarantine repo, immutable storage, signature/checksum verification</td>
          </tr>
          <tr>
            <td>DMZ</td>
            <td>내부망</td>
            <td>검증된 artifact manifest</td>
            <td><span class="tag ok">허용 후보</span></td>
            <td>내부망에서 재검증 후 internal mirror로 승격</td>
          </tr>
          <tr>
            <td>DMZ</td>
            <td>내부망</td>
            <td>deploy command</td>
            <td><span class="tag warn">검토 필요</span></td>
            <td>가능하면 pull로 대체. push라면 명령 allowlist와 내부 승인 필요</td>
          </tr>
          <tr>
            <td>내부망</td>
            <td>DMZ</td>
            <td>poll request, artifact id, agent id</td>
            <td><span class="tag warn">검토 필요</span></td>
            <td>내부 호스트명 제거, 그룹 alias, coarse status 금지, 최소 질의</td>
          </tr>
          <tr>
            <td>내부망</td>
            <td>연구/개발망</td>
            <td>deploy log, status, stack trace, server list, token</td>
            <td><span class="tag bad">금지</span></td>
            <td>GitLab environment 업데이트 금지, 내부 audit log에만 저장</td>
          </tr>
          <tr>
            <td>내부망</td>
            <td>업무망</td>
            <td>운영 배포 결과, 장애 로그, 환경 metadata</td>
            <td><span class="tag bad">금지</span></td>
            <td>필요 시 보안 승인된 별도 요약 채널만 설계</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section>
      <h2>Credential 위치</h2>
      <table>
        <thead>
          <tr>
            <th>자격증명</th>
            <th>권장 위치</th>
            <th>금지/주의</th>
            <th>회전 정책</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Webhook secret</td>
            <td>GitLab 설정과 DMZ receiver</td>
            <td>내부망 배포 권한과 결합 금지</td>
            <td>이중 secret 전환 기간을 두고 정기 회전</td>
          </tr>
          <tr>
            <td>Deploy token</td>
            <td>필요한 artifact repo에 read-only로 한정</td>
            <td>source write, broad API scope 금지</td>
            <td>만료일, 소유자, 프로젝트 범위 고정</td>
          </tr>
          <tr>
            <td>SSH key / host key</td>
            <td>내부망 orchestrator 또는 내부 vault</td>
            <td>DMZ, 연구/개발망 GitLab 변수 저장 금지</td>
            <td>서버 그룹별 분리, forced command, 정기 교체</td>
          </tr>
          <tr>
            <td>Runner token</td>
            <td>해당 runner host의 local config</td>
            <td>내부 runner가 외부 GitLab에 callback하는 구조 주의</td>
            <td>runner 단위 rotate/revoke, clone 탐지</td>
          </tr>
          <tr>
            <td>Signing key</td>
            <td>KMS/HSM 또는 서명 전용 서비스</td>
            <td>일반 build script와 같은 환경 변수로 노출 금지</td>
            <td>key id, 폐기 목록, public key 배포, dual-sign 전환</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="callout-grid">
      <div class="card">
        <h3>Artifact 이동 경로</h3>
        <ol>
          <li>GitLab merge 후 build runner가 artifact, checksum, SBOM, provenance를 생성합니다.</li>
          <li>DMZ quarantine repo가 artifact와 manifest를 받되 내부 대상 서버 정보는 받지 않습니다.</li>
          <li>내부망 mirror가 artifact를 가져와 signature와 digest를 다시 확인한 뒤 배포 후보로 승격합니다.</li>
        </ol>
      </div>
      <div class="card">
        <h3>감사 로그 위치</h3>
        <p>GitLab에는 merge, pipeline, release evidence까지만 남깁니다. 내부망에는 deployment ledger를 별도로 두고, 서버별 결과와 rollback 기록을 내부망 밖으로 내보내지 않습니다.</p>
      </div>
      <div class="card">
        <h3>장애/복구 흐름</h3>
        <p>Webhook 중복은 idempotency key로 제거하고, 일부 서버 실패는 서버별 state machine으로 재시도합니다. Rollback은 이전 artifact digest를 기준으로 수행하며, raw log는 내부 감사 저장소에만 남깁니다.</p>
      </div>
    </section>

    <section>
      <h2>개발자 사용 흐름</h2>
      <ol>
        <li>개발자는 연구/개발망 GitLab에 merge request를 올리고 AI 개발 agent는 코드/테스트/문서 보조만 수행합니다.</li>
        <li>merge 후 pipeline이 test, build, scan, SBOM, signing을 수행하고 DMZ에 배포 후보 manifest를 남깁니다.</li>
        <li>운영 승인자는 내부망 배포 포털에서 artifact digest와 검증 결과를 보고 승인합니다.</li>
        <li>내부 pull agent가 대상 서버에 배포하고 결과를 내부 audit log에 기록합니다.</li>
        <li>개발자가 외부에서 볼 수 있는 정보는 보안팀이 허용한 후보 생성 상태까지로 제한합니다.</li>
      </ol>
    </section>

    <section>
      <h2>주요 판단 근거</h2>
      <ul>
        <li>NIST SP 800-207은 네트워크 위치만으로 신뢰하지 말고 resource 중심으로 인증과 권한을 판단하라고 설명합니다.</li>
        <li>GitLab Runner는 GitLab 서버를 polling해 job을 받고 trace/status를 되돌리는 구조이므로, 내부망 runner가 연구/개발망 GitLab에 붙으면 내부 데이터 역류 경로가 생깁니다.</li>
        <li>GitLab webhook은 secret header와 payload template을 제공하지만, payload 자체가 어떤 정보를 싣는지 별도로 최소화해야 합니다.</li>
        <li>Jenkins credentials와 GitLab CI/CD variables 모두 masking이 완전한 비밀 유출 방지책은 아니므로, production secret은 내부망 secret store에 두는 편이 안전합니다.</li>
        <li>SLSA와 Sigstore의 핵심은 artifact를 소비하는 지점에서 provenance/signature/digest를 검증하는 것입니다.</li>
      </ul>
    </section>

    <footer>
      Placeholder only. 실제 회사명, 서버명, IP, 계정명, token 값은 포함하지 않았습니다.
    </footer>
  </main>
</body>
</html>
```
