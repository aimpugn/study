# GitLab HA

## GitLab 이중화 — Gitaly Cluster 설정 Playbook 및 동작 원리

GitLab을 이중화(HA)할 때 Git 저장소 계층의 가용성/일관성을 담당하는 것이 **Gitaly Cluster (Praefect)** 입니다. 근거 기반으로 동작 원리부터 Ansible Playbook 예시까지 정리해드릴게요.

> ⚠️ **2025년 기준 주의사항**: GitLab 17.x 부터 Gitaly Cluster의 후속 아키텍처인 **Gitaly Cluster with Raft**(베타)가 도입되었고, 기존 Praefect 기반 Cluster는 **GitLab 19.0(2026년 5월 예정)에서 deprecated** 될 예정입니다. 다만 현재 production-ready인 것은 여전히 Praefect 기반이므로, 본 문서는 그 기준으로 정리합니다. ([Gitaly Cluster docs](https://docs.gitlab.com/administration/gitaly/))

---

### 1. 왜 Gitaly Cluster인가 — 근거

#### 1.1 Git 저장소는 "공유 스토리지"로 이중화하면 안 된다

GitLab 공식 문서가 명시적으로 권고하는 내용:

> Gitaly Cluster의 대안으로 NFS, Gitaly의 sharded 구성 등이 있지만, 이들은 추천되지 않습니다.
>
> - NFS는 GitLab 14.0부터 deprecated 되었고,
> - sharded Gitaly는 노드 단위 고가용성을 제공하지 않습니다.

**근거:**
- Git은 파일 락(lock), `packed-refs`, `index.lock` 등 파일시스템 세부 동작에 민감함 → NFS의 캐시/잠금 의미가 Git과 충돌
- 공유 스토리지 자체가 SPOF가 됨
- 여러 Gitaly 노드가 동일 저장소에 동시 쓰기 → race condition

→ 결론: **각 노드가 자체 디스크를 갖되 애플리케이션 레벨에서 복제**하는 모델이 필요. 이게 Gitaly Cluster.

#### 1.2 권장 토폴로지

GitLab 공식 reference architecture (3K users 이상)에서:

| 구성요소 | 노드 수 | 역할 |
| --- | --- | --- |
| Praefect | 3 | 라우터 + consensus (Postgres 기반) |
| Gitaly | 3 (홀수) | 실제 저장소 보관 (replica) |
| Praefect용 PostgreSQL | 1+ (HA 권장) | replication 메타데이터 |
| Consul | 3 | 서비스 디스커버리 (선택) |

**홀수**가 중요한 이유: Praefect의 primary election은 **strong consistency 모드**에서 quorum(과반)을 요구하기 때문. 짝수면 split-brain 위험.

---

### 2. Gitaly Cluster 동작 원리

#### 2.1 컴포넌트 구조

```
                    [GitLab Rails / Workhorse / Shell]
                                  │
                                  │ gRPC (gitaly-address: tcp://praefect)
                                  ▼
                  ┌───────────────────────────────────┐
                  │         Praefect (router)         │
                  │  - 가상 저장소(virtual storage)    │
                  │  - 라우팅 / 복제 큐 / health check│
                  └───────────────┬───────────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              ▼                   ▼                   ▼
        [Gitaly node-1]     [Gitaly node-2]     [Gitaly node-3]
        (primary for          (secondary)         (secondary)
         repo A)
            │                     │                   │
            └── 자체 디스크에 .git 저장소 보관 ──────┘
                                  │
                                  ▼
                        [PostgreSQL (Praefect용)]
                        - replication state
                        - replication queue
                        - primary election lease
```

#### 2.2 핵심 개념

**가상 저장소 (Virtual Storage)**
GitLab Rails에게는 `default`라는 단일 저장소처럼 보이지만, 실제로는 Praefect가 뒤에 있는 N개 Gitaly 노드로 분산/복제합니다. 클라이언트는 **Praefect만 알면 됨**.

**Primary 선출**
- 저장소(repository) 단위로 primary가 정해짐 — 노드 단위가 아님
- 즉, repo A의 primary는 gitaly-1, repo B의 primary는 gitaly-2 일 수 있음 → 부하 분산
- primary가 unhealthy하면 Praefect가 다른 in-sync replica를 자동 promote

**Replication (두 가지 모드)**

1. **Strong consistency (기본, 권장)** — Reference Transaction Hook
   - `git push` 같은 쓰기 작업이 들어오면 Praefect가 voting transaction 생성
   - 모든 replica가 동일한 결과를 만들어야(quorum) 커밋이 확정됨
   - quorum 실패 시 push 자체가 거부됨 → **데이터 불일치 원천 차단**

2. **Eventual consistency** — Replication queue
   - primary가 먼저 쓰고, replication job을 큐에 넣어 secondary로 비동기 복제
   - 빠르지만 짧은 시간 동안 replica 간 차이 존재

> 실제로 GitLab 14.0부터 strong consistency가 default. ([Gitaly Cluster behind the scenes](https://docs.gitlab.com/administration/gitaly/praefect/))

**읽기 분산 (Distribution of reads)**
읽기 요청은 in-sync replica 중 하나로 분산됨 → 읽기 부하 분산. `up-to-date` replica가 없는 경우만 primary로 fallback.

#### 2.3 쓰기 흐름 (실제로 무슨 일이 일어나는가)

```
1. 사용자가 git push
2. GitLab Workhorse → Praefect (gRPC)
3. Praefect가 해당 repo의 primary Gitaly 식별 (Postgres lookup)
4. Praefect가 모든 healthy replica에 동시에 RPC 전달
5. 각 Gitaly가 reference-transaction hook으로 변경 사항 hash를 Praefect에 vote
6. Praefect가 quorum 확인:
   - quorum 달성 → 모든 노드에 commit 신호 → push 성공
   - quorum 실패 → 모든 노드에 abort 신호 → push 거부
7. 비동기 replication queue로 뒤처진 replica 따라잡기
```

---

### 3. Ansible Playbook 예시

GitLab 공식 [gitlab-environment-toolkit](https://gitlab.com/gitlab-org/gitlab-environment-toolkit) 이 Ansible 기반인데, 핵심만 추려서 정리합니다. (Omnibus 패키지 기준)

#### 3.1 Inventory

```ini
## inventory/hosts.ini
[praefect]
praefect-1 ansible_host=10.0.1.11
praefect-2 ansible_host=10.0.1.12
praefect-3 ansible_host=10.0.1.13

[gitaly]
gitaly-1 ansible_host=10.0.2.11 gitaly_storage=gitaly-1
gitaly-2 ansible_host=10.0.2.12 gitaly_storage=gitaly-2
gitaly-3 ansible_host=10.0.2.13 gitaly_storage=gitaly-3

[praefect_db]
praefect-db-1 ansible_host=10.0.3.11

[gitlab_rails]
rails-1 ansible_host=10.0.4.11
rails-2 ansible_host=10.0.4.12
```

#### 3.2 공통 변수 (group_vars/all.yml)

```yaml
## 모든 노드 간 통신용 공유 시크릿 — 반드시 vault로 암호화
praefect_external_token: "{{ vault_praefect_external_token }}"   # Rails ↔ Praefect
praefect_internal_token: "{{ vault_praefect_internal_token }}"   # Praefect ↔ Gitaly
gitaly_token: "{{ vault_gitaly_token }}"                          # 동일하게 사용

praefect_db_password: "{{ vault_praefect_db_password }}"
praefect_db_host: "10.0.3.11"
praefect_db_name: "praefect_production"
praefect_db_user: "praefect"

virtual_storage_name: "default"
```

#### 3.3 Praefect용 PostgreSQL 준비 (playbook-praefect-db.yml)

```yaml
- name: Provision Praefect PostgreSQL
  hosts: praefect_db
  become: true
  tasks:
    - name: Install GitLab Omnibus (DB role only)
      apt:
        deb: "https://packages.gitlab.com/gitlab/gitlab-ee/packages/ubuntu/jammy/gitlab-ee_17.5.0-ee.0_amd64.deb"

    - name: Configure as standalone Postgres for Praefect
      copy:
        dest: /etc/gitlab/gitlab.rb
        content: |
          roles ['postgres_role']
          postgresql['enable'] = true
          postgresql['listen_address'] = '0.0.0.0'
          postgresql['md5_auth_cidr_addresses'] = ['10.0.1.0/24']
          postgresql['sql_user_password'] = '{{ praefect_db_password | password_hash("md5", "praefect") }}'
          praefect::postgresql['enable'] = false
          gitlab_rails['auto_migrate'] = false
      notify: reconfigure gitlab

    - name: Create praefect database
      command: >
        gitlab-psql -d template1 -c
        "CREATE ROLE praefect WITH LOGIN PASSWORD '{{ praefect_db_password }}';
         CREATE DATABASE praefect_production WITH OWNER praefect;"
      changed_when: false
      failed_when: false

  handlers:
    - name: reconfigure gitlab
      command: gitlab-ctl reconfigure
```

#### 3.4 Praefect 노드 (playbook-praefect.yml)

```yaml
- name: Configure Praefect nodes
  hosts: praefect
  become: true
  tasks:
    - name: Render gitlab.rb for Praefect
      copy:
        dest: /etc/gitlab/gitlab.rb
        content: |
          # 다른 모든 컴포넌트 비활성화
          postgresql['enable'] = false
          redis['enable'] = false
          nginx['enable'] = false
          puma['enable'] = false
          sidekiq['enable'] = false
          gitlab_workhorse['enable'] = false
          prometheus['enable'] = false
          alertmanager['enable'] = false
          grafana['enable'] = false
          gitlab_exporter['enable'] = false
          gitaly['enable'] = false

          # Praefect 활성화
          praefect['enable'] = true
          praefect['configuration'] = {
            listen_addr: '0.0.0.0:2305',
            auth: {
              token: '{{ praefect_external_token }}',
            },
            virtual_storage: [
              {
                name: '{{ virtual_storage_name }}',
                node: [
                  {
                    storage: 'gitaly-1',
                    address: 'tcp://10.0.2.11:8075',
                    token: '{{ praefect_internal_token }}',
                  },
                  {
                    storage: 'gitaly-2',
                    address: 'tcp://10.0.2.12:8075',
                    token: '{{ praefect_internal_token }}',
                  },
                  {
                    storage: 'gitaly-3',
                    address: 'tcp://10.0.2.13:8075',
                    token: '{{ praefect_internal_token }}',
                  },
                ],
              },
            ],
            database: {
              host: '{{ praefect_db_host }}',
              port: 5432,
              user: '{{ praefect_db_user }}',
              password: '{{ praefect_db_password }}',
              dbname: '{{ praefect_db_name }}',
              sslmode: 'disable',
            },
            # Strong consistency: failover에서 in-sync replica만 promote
            failover: {
              enabled: true,
              election_strategy: 'per_repository',
            },
            # Reconciliation: out-of-sync replica를 주기적으로 따라잡게
            reconciliation: {
              scheduling_interval: '5m',
            },
          }

      notify: reconfigure gitlab

    - name: Run Praefect DB migration (한 노드에서만)
      command: gitlab-ctl praefect sql-migrate
      run_once: true
      delegate_to: "{{ groups['praefect'][0] }}"

  handlers:
    - name: reconfigure gitlab
      command: gitlab-ctl reconfigure
```

#### 3.5 Gitaly 노드 (playbook-gitaly.yml)

```yaml
- name: Configure Gitaly nodes
  hosts: gitaly
  become: true
  tasks:
    - name: Render gitlab.rb for Gitaly
      copy:
        dest: /etc/gitlab/gitlab.rb
        content: |
          postgresql['enable'] = false
          redis['enable'] = false
          nginx['enable'] = false
          puma['enable'] = false
          sidekiq['enable'] = false
          gitlab_workhorse['enable'] = false
          prometheus['enable'] = false
          grafana['enable'] = false
          gitlab_exporter['enable'] = false
          praefect['enable'] = false

          # Gitaly 활성화
          gitaly['enable'] = true
          gitaly['configuration'] = {
            listen_addr: '0.0.0.0:8075',
            auth: {
              # Praefect가 이 토큰으로 호출함
              token: '{{ praefect_internal_token }}',
            },
            # storage 이름은 Praefect 설정의 'storage'와 정확히 일치해야 함
            storage: [
              {
                name: '{{ gitaly_storage }}',
                path: '/var/opt/gitlab/git-data/repositories',
              },
            ],
          }

          # Praefect 환경에서는 hashed storage가 강제됨
          gitlab_rails['gitaly_token'] = '{{ praefect_internal_token }}'
      notify: reconfigure gitlab

  handlers:
    - name: reconfigure gitlab
      command: gitlab-ctl reconfigure
```

#### 3.6 GitLab Rails 노드 측 설정

Rails는 **Praefect만 바라봐야** 합니다. 직접 Gitaly를 가리키면 안 됨.

```yaml
- name: Point GitLab Rails at Praefect
  hosts: gitlab_rails
  become: true
  tasks:
    - name: Configure git_data_dirs to use Praefect
      blockinfile:
        path: /etc/gitlab/gitlab.rb
        block: |
          git_data_dirs({
            "default" => {
              "gitaly_address" => "tcp://10.0.1.11:2305",
              "gitaly_token"   => "{{ praefect_external_token }}"
            }
          })
          # 로컬 Gitaly 비활성화
          gitaly['enable'] = false
      notify: reconfigure gitlab

  handlers:
    - name: reconfigure gitlab
      command: gitlab-ctl reconfigure
```

> 실제로는 `10.0.1.11` 단일 IP 대신 **Praefect 앞에 LB(예: HAProxy, Consul + internal LB)** 를 두는 게 정석입니다. Praefect 하나가 죽어도 Rails가 다른 Praefect로 넘어갈 수 있도록.

#### 3.7 검증 작업

```yaml
- name: Verify Gitaly Cluster
  hosts: praefect[0]
  become: true
  tasks:
    - name: Check Praefect → Gitaly 연결 상태
      command: gitlab-ctl praefect dial-nodes
      register: dial
      changed_when: false

    - debug: var=dial.stdout_lines

    - name: Check replication 상태
      command: gitlab-ctl praefect dataloss
      register: dataloss
      changed_when: false

    - debug: var=dataloss.stdout_lines
```

`dial-nodes`는 Praefect가 모든 Gitaly 노드와 통신 가능한지, `dataloss`는 어느 repo가 out-of-sync인지 보여줍니다.

---

### 4. 운영 시 핵심 체크리스트

| 항목 | 이유 |
|---|---|
| Gitaly 노드 수는 **3 이상 홀수** | Quorum (과반) 기반 정합성 |
| Praefect도 **3대 이상**, 앞에 LB | Praefect 자체도 SPOF 방지 |
| Praefect용 Postgres는 **GitLab 메인 DB와 분리** | 메타데이터 격리, replication 큐 부하 분리 |
| `failover.election_strategy: per_repository` | 저장소 단위 primary election (구버전 `sql` 방식보다 권장) |
| 토큰 3종(`praefect_external`, `praefect_internal`, `gitlab-shell`)을 Vault로 관리 | 평문 노출 시 저장소 직접 접근 가능 |
| `gitlab-ctl praefect dataloss`를 모니터링 | Out-of-sync replica 조기 감지 |
| 기존 sharded Gitaly에서 마이그레이션 시 `gitlab-ctl praefect-migrate` 도구 사용 | 수동 rsync는 hashed storage 매핑 깨짐 |
| 백업은 **Praefect 경유**로 받기 (`gitlab-backup`) | 직접 Gitaly 디스크 tar로 묶으면 복제 메타와 어긋남 |

---

### 5. 실패 시나리오별 동작 정리

| 시나리오 | Gitaly Cluster의 처리 |
|---|---|
| Gitaly 1대 죽음 | 해당 노드가 primary였던 repo는 in-sync replica로 자동 failover. 사용자는 인지 못 함 (수 초 지연만) |
| Gitaly 2대 죽음 (3대 중) | Quorum 깨짐 → 쓰기 차단(read-only). 데이터 불일치보다 가용성 희생 — 의도된 설계 |
| Praefect 1대 죽음 | LB가 다른 Praefect로 라우팅. 무중단 |
| Praefect Postgres 죽음 | Praefect 신규 라우팅 결정 불가 → 사실상 cluster 마비. **그래서 Postgres도 HA로 가야 함** |
| 네트워크 파티션 | 소수 측 Praefect는 quorum 못 잡고 쓰기 거부 → split-brain 방지 |

---

### 참고 문서

- [GitLab Docs — Gitaly Cluster](https://docs.gitlab.com/administration/gitaly/)
- [GitLab Docs — Praefect configuration](https://docs.gitlab.com/administration/gitaly/praefect/)
- [Reference architectures (3K~50K users)](https://docs.gitlab.com/administration/reference_architectures/)
- [gitlab-environment-toolkit (공식 Ansible)](https://gitlab.com/gitlab-org/gitlab-environment-toolkit)

위 Playbook은 **개념 정리용 골격**이라 실제 적용하실 때는 사용 중인 GitLab 버전(특히 17.x 이상이면 `praefect['configuration']` 구조 변경 적용됨)과 네트워크/방화벽 정책에 맞춰 보정하셔야 합니다. 특정 버전이나 규모(예: "1000명 사용자" "K8s 환경")에 맞춘 변형이 필요하시면 알려주세요.
