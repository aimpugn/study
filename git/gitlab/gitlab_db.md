# GitLab DB

## 1. PostgreSQL 만 지원

GitLab 12.1 이전에는 MySQL도 지원했지만, **GitLab 12.1 (2019년 7월) 부터 MySQL 지원이 완전히 제거**되어 현재는 **PostgreSQL이 유일한 공식 지원 DB**입니다.

**근거:**
- GitLab 공식 문서: "GitLab requires PostgreSQL. MySQL/MariaDB is no longer supported."
- 이유는 PostgreSQL의 고유 기능(부분 인덱스, CTE, window functions, `LATERAL` join, JSONB 등)을 적극 활용하기 때문. MySQL 호환을 유지하면서는 쿼리 최적화와 기능 추가에 제약이 컸음.
- MariaDB, MySQL, Aurora MySQL 모두 **공식적으로 비지원**. (Aurora **PostgreSQL**도 GitLab 14.4부터 비지원이고, 정식 지원은 일반 PostgreSQL과 Google Cloud SQL for PostgreSQL, Azure Database for PostgreSQL 정도)

지원하는 PostgreSQL 버전은 GitLab 버전마다 달라집니다. 예를 들어 GitLab 17.x 는 PostgreSQL 14/15/16 을 지원하고, 16 이 권장입니다. (정확한 매트릭스는 사용 중인 GitLab 버전 릴리스 노트 확인 필요)

---

## 2. 별도 PostgreSQL 서버가 없을 때 — **Omnibus 패키지가 로컬에 PostgreSQL을 자동 실행**

이게 GitLab의 가장 큰 특징 중 하나인데, 설치 방식에 따라 동작이 다릅니다.

### 2.1 Omnibus 패키지 (가장 일반적)

`gitlab-ee` / `gitlab-ce` deb/rpm 을 설치하면, 패키지 안에 **PostgreSQL 바이너리가 번들로 포함**되어 있습니다. 별도로 PostgreSQL을 설치하지 않아도 즉시 동작합니다.

```text
/opt/gitlab/embedded/
├── bin/
│   ├── postgres          ← 번들된 PostgreSQL 서버
│   ├── psql
│   ├── pg_dump
│   └── ...
└── postgresql/<version>/
```

- 데이터 디렉토리: `/var/opt/gitlab/postgresql/data`
- 소켓: `/var/opt/gitlab/postgresql/`
- 기본 리스닝: **Unix socket only** (TCP는 `gitlab.rb`에서 명시적으로 켜야 함)
- 관리 명령: `gitlab-ctl status postgresql`, `gitlab-psql` (자동 인증된 CLI)
- 사용자: `gitlab-psql` 라는 OS 계정으로 실행

`/etc/gitlab/gitlab.rb` 의 기본값:

```ruby
postgresql['enable'] = true   # 기본 true → 로컬 번들 PG 실행
```

**즉, 아무 설정 없이 `gitlab-ctl reconfigure` 만 해도 단일 호스트에서 GitLab + PostgreSQL이 같이 떠 있는 상태가 됩니다.** 소규모 설치에서는 이게 일반적이고 정상적인 형태예요.

### 2.1.1 소켓 전용의 실체 — port 는 5432 인데 TCP 는 닫혀 있다

위에서 "Unix socket only"라고 적은 상태가 실제로 어떻게 보이는지 실측으로 고정합니다. Omnibus 기본 설정의 GitLab 서버에서 (2026-06 관측):

```text
# gitlab-psql -c 'SHOW listen_addresses;'
 listen_addresses
------------------

(1 row)

# gitlab-psql -c 'SHOW port;'
 port
------
 5432
(1 row)
```

처음 보면 "5432로 구성되어 있으니 TCP 5432를 듣고 있겠구나"라고 읽기 쉽지만, 반대입니다.

- `listen_addresses`가 **빈 값**입니다. PostgreSQL 공식 의미로 빈 값은 "어떤 IP에서도 듣지 않음 = TCP 완전 비활성, Unix 소켓으로만 접속 가능"입니다. 순정 PostgreSQL의 기본값은 `'localhost'`(TCP 켜짐)인데, Omnibus가 이를 빈 값으로 바꿔서 배포하는 것입니다. 모든 GitLab 구성 요소(Puma, Sidekiq 등)가 같은 호스트에 있으니 네트워크 통로가 필요 없고, 노출 면적과 비밀번호 관리를 동시에 없애는 의도된 선택입니다.
- 그런데도 `port`가 5432인 이유는, 이 값이 TCP 포트이자 **소켓 파일 이름의 접미사**(`/var/opt/gitlab/postgresql/.s.PGSQL.5432`)로도 쓰이기 때문입니다. 문패는 5432인데 전화선은 뽑혀 있는 상태입니다.
- `gitlab-psql` 명령이 비밀번호 없이 붙는 것도 같은 그림의 일부입니다. 이 래퍼는 그 소켓으로 접속하고, peer 인증(커널이 접속 프로세스의 OS 사용자를 확인)이 통과시킵니다.

실제 점유 여부는 설정값이 아니라 소켓 관측으로 판정합니다.

```sh
ss -ltnp | grep 5432    # 아무것도 안 나옴 → TCP 미청취
ss -xlp | grep PGSQL    # .s.PGSQL.5432 소켓 파일이 보임
```

같은 호스트에 다른 PostgreSQL(예: Nexus용)을 추가할 때의 함의: TCP 5432가 비어 있으므로 지금은 5432로 띄워도 충돌하지 않지만, `gitlab.rb` 변경(예: `postgresql['listen_address']` 설정, 복제·모니터링 연동)으로 번들 PG의 TCP가 켜지는 순간 `Address already in use`가 미래에 터집니다. 그래서 추가 인스턴스는 5433 이상을 권장합니다. TCP/소켓 두 접속 통로와 `listen_addresses`의 일반 원리는 [database/postgresql/connections.md](../../database/postgresql/connections.md), GitLab 호스트에 Nexus용 PostgreSQL을 배치하는 적용 사례는 [linux/privilege-isolation.md](../../linux/privilege-isolation.md)에 정리되어 있습니다.

### 2.2 외부 PostgreSQL 사용 시 (HA, RDS, 분리 서버 등)

`gitlab.rb` 에서 번들 PG를 명시적으로 끄고 외부 접속 정보를 줍니다:

```ruby
postgresql['enable'] = false        # 번들 PG 끄기

gitlab_rails['db_adapter']  = 'postgresql'
gitlab_rails['db_host']     = '10.0.3.10'
gitlab_rails['db_port']     = 5432
gitlab_rails['db_database'] = 'gitlabhq_production'
gitlab_rails['db_username'] = 'gitlab'
gitlab_rails['db_password'] = '...'
```

이 경우 외부 PG는 **사전에 DB·유저·확장(extension)을 생성**해 둬야 합니다 (`pg_trgm`, `btree_gist`, `plpgsql` 등 필수 extension).

### 2.3 Helm chart (Kubernetes)

`gitlab/gitlab` Helm chart 도 in-cluster PostgreSQL(Bitnami chart 기반)을 기본으로 띄웁니다. 다만 이는 **개발/평가용**이고, **운영에서는 외부 관리형 DB 사용을 공식 권장**합니다.

```yaml
# values.yaml
postgresql:
  install: false         # in-cluster PG 비활성화
global:
  psql:
    host: my-rds.amazonaws.com
    port: 5432
    database: gitlabhq_production
    username: gitlab
    password:
      secret: gitlab-postgres-password
      key: password
```

---

## 3. 정리 표

| 설치 방식 | 외부 PG 미지정 시 동작 | 권장 운영 형태 |
| --- | --- | --- |
| Omnibus (deb/rpm) | `/opt/gitlab/embedded` 번들 PG가 로컬에서 자동 실행 | 소규모: 그대로 OK / 중대규모·HA: 외부로 분리 |
| Helm (K8s) | chart의 in-cluster Postgres pod이 뜸 | 운영은 관리형 PG (RDS, Cloud SQL 등) |
| Source 설치 | 자동 설치 안 됨 — OS 패키지로 직접 설치 필요 | 직접 운영 |
| Docker (`gitlab/gitlab-ee` 이미지) | 컨테이너 내부에 번들 PG 실행 (Omnibus 기반) | 운영은 외부 PG로 분리 |

---

## 4. HA 맥락에서 중요한 포인트

이전 답변에서 다룬 Gitaly Cluster 와 연결지어 보면:

- GitLab HA 구성에서는 **PostgreSQL을 별도 노드(또는 관리형 서비스)로 분리**하는 것이 거의 필수입니다. 로컬 번들 PG 그대로 두면 그 노드가 SPOF가 되니까요.
- 이때 **GitLab 메인 DB**(`gitlabhq_production`)와 **Praefect용 DB**(`praefect_production`)는 **물리적으로 분리하거나 최소한 인스턴스를 분리**하는 것이 권장됩니다. Praefect의 replication 큐 부하가 메인 DB 성능에 영향을 주는 걸 막기 위함.
- PostgreSQL 자체의 HA는 GitLab이 직접 해주지 않습니다. **Patroni + Consul** 조합(Omnibus가 패키징해 둠) 또는 **클라우드 관리형 PG(RDS Multi-AZ, Cloud SQL HA 등)** 를 사용합니다.

---

요약하면 — **GitLab은 PostgreSQL만 지원하며, 외부 DB를 지정하지 않으면 Omnibus 패키지가 들고 온 번들 PostgreSQL이 로컬에서 자동으로 실행**됩니다. 단일 노드 설치에서는 이 상태가 정상이고, HA로 가는 순간 외부로 빼는 흐름이에요.

특정 환경(예: RHEL + 외부 PG + HA)이나 버전에 맞춘 더 구체적인 설정이 필요하시면 알려주세요.
