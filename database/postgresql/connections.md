# PostgreSQL 접속 경로와 배치 패턴

"PostgreSQL은 항상 TCP 5432로 듣는가?"라는 질문에 답합니다. 직답은 "아니다"입니다. 순정 설치의 기본값은 TCP(localhost)와 Unix 소켓을 함께 여는 것이지만, TCP는 설정 한 줄로 완전히 끌 수 있고, 실제 제품들은 소켓 전용으로 배포하기도 합니다. 이 구분을 모르면 "`SHOW port`가 5432인데 왜 TCP 5432에 아무것도 안 보이지?" 같은 혼동이 생깁니다. 실제로 GitLab Omnibus가 번들한 PostgreSQL이 정확히 그런 상태로 운영됩니다([gitlab_db.md](../../git/gitlab/gitlab_db.md) 참조).

## 목차

- [두 개의 접속 통로](#두-개의-접속-통로)
- [무엇이 어느 통로를 여는가](#무엇이-어느-통로를-여는가)
- [지금 어떤 통로가 열려 있는지 관측하기](#지금-어떤-통로가-열려-있는지-관측하기)
- [통로와 인증의 결합 pg_hba 와 peer](#통로와-인증의-결합-pg_hba-와-peer)
- [네 가지 배치 패턴](#네-가지-배치-패턴)
- [같은 호스트에 인스턴스 여러 개를 둘 때](#같은-호스트에-인스턴스-여러-개를-둘-때)
- [teach back 과 검증 경로](#teach-back-과-검증-경로)
- [근거](#근거)

## 두 개의 접속 통로

PostgreSQL 서버 프로세스는 클라이언트를 받는 통로를 두 종류 가집니다.

- TCP 소켓

    `IP:포트` 형태의 네트워크 통로입니다. 다른 호스트에서 접속할 수 있는 유일한 통로이고, `127.0.0.1`(loopback)로 한정하면 같은 호스트 전용이 되지만 여전히 "TCP를 듣는" 상태입니다.

- Unix 도메인 소켓

    파일시스템 위의 특수 파일을 통한 같은-호스트 전용 통로입니다. 네트워크 스택을 타지 않습니다. 파일 이름은 `.s.PGSQL.<port>` 형식이고, 디렉터리는 배포판마다 다릅니다. 예를 들어 RHEL 계열 시스템 패키지는 `/var/run/postgresql/.s.PGSQL.5432`, GitLab Omnibus 번들은 `/var/opt/gitlab/postgresql/.s.PGSQL.5432`에 만듭니다.

클라이언트 쪽에서 이 선택은 `-h` 옵션으로 갈립니다. `psql`을 `-h` 없이 실행하면 소켓으로 붙고, `-h 127.0.0.1`을 주면 TCP로 붙습니다. 같은 DB라도 들어가는 문이 다르고, 뒤에서 보듯 문마다 적용되는 인증 규칙도 다릅니다.

## 무엇이 어느 통로를 여는가

`postgresql.conf`의 세 파라미터가 통로를 결정합니다.

- `listen_addresses` — TCP를 제어합니다. 기본값은 `'localhost'`라서 순정 설치는 `127.0.0.1:5432`를 듣습니다. `'*'`는 모든 인터페이스, IP 목록은 그 주소들만. 그리고 공식 문서 기준으로 **빈 값(`''`)이면 TCP를 전혀 듣지 않고 Unix 소켓으로만 접속**할 수 있습니다.
- `unix_socket_directories` — 소켓 파일을 만들 디렉터리(들)입니다. 빈 값으로 두면 소켓도 안 만들 수 있지만 그런 구성은 드뭅니다.
- `port` — 이중 역할입니다. TCP 포트 번호이면서, 동시에 **소켓 파일 이름의 접미사**(`.s.PGSQL.<port>`)로도 쓰입니다.

여기서 흔한 오해 하나를 바로잡습니다. `SHOW port`가 5432를 보여준다고 해서 TCP 5432를 점유 중이라는 뜻이 아닙니다. `listen_addresses`가 비어 있으면 그 5432는 소켓 파일 이름에만 쓰입니다. 문패에 5432라고 적혀 있지만 전화선(TCP)은 뽑혀 있는 상태입니다. 점유 여부의 판정은 설정값이 아니라 아래 관측 명령으로 합니다.

## 지금 어떤 통로가 열려 있는지 관측하기

```sh
# TCP: 이 줄에 아무것도 안 나오면 TCP 는 듣지 않는 것
ss -ltnp | grep 5432

# Unix 소켓: .s.PGSQL.<port> 가 보이면 소켓은 열려 있는 것
ss -xlp | grep PGSQL
```

서버 안에서 설정값으로 확인하려면 psql에서 다음을 봅니다.

```sql
SHOW listen_addresses;        -- 빈 값이면 TCP 미청취
SHOW unix_socket_directories; -- 소켓 파일 위치
SHOW port;                    -- TCP 포트이자 소켓 파일명 접미사
```

두 관측을 함께 보는 이유가 있습니다. `ss`는 "지금 실제로 무엇이 열려 있는가"(사실)를, `SHOW`는 "왜 그렇게 열려 있는가"(설정)를 보여줍니다. 진단할 때는 항상 사실을 먼저 보고 설정으로 원인을 좁힙니다.

## 통로와 인증의 결합 pg_hba 와 peer

접속 통로는 단순히 경로 차이가 아니라 인증 모델 차이로 이어집니다. `pg_hba.conf`의 규칙이 통로별로 나뉘기 때문입니다.

- `local`로 시작하는 줄은 Unix 소켓 접속에만 적용됩니다.
- `host`로 시작하는 줄은 TCP 접속에만 적용됩니다.

소켓에서만 가능한 인증이 `peer`입니다. 커널이 소켓 반대편 프로세스의 OS 사용자를 알려주면(같은 호스트라서 가능), PostgreSQL이 그 OS 사용자 이름과 DB role 이름을 대조해서 일치하면 통과시킵니다. 비밀번호가 아예 없습니다. `sudo -u postgres psql`이 비밀번호 없이 바로 붙는 이유가 이것입니다. OS 사용자 `postgres`로 소켓에 접속했고, 같은 이름의 DB role이 있으니 peer가 통과시킨 겁니다.

TCP에는 peer가 없으므로 `scram-sha-256` 같은 비밀번호 기반(또는 인증서 기반) 인증을 씁니다. 그래서 "소켓 전용 + peer" 조합은 네트워크 노출이 0이면서 비밀번호 관리도 없는 로컬 운영 모델이 되고, 제품들이 번들 DB에 이 조합을 즐겨 쓰는 이유가 됩니다.

## 네 가지 배치 패턴

같은 PostgreSQL이라도 "누가 소유하고 누가 접속하느냐"에 따라 배치가 갈립니다. 실무에서 만나는 형태는 크게 네 가지입니다.

1. 시스템 공유 인스턴스

    배포판 패키지(`dnf install postgresql-server` 등)로 설치한 인스턴스 하나에 여러 앱이 각자의 DB와 role을 만들어 같이 씁니다. 같은 호스트의 앱들은 소켓 또는 localhost TCP로 붙습니다. 전통적인 형태이고, DB·role 단위 격리로 충분한 다수의 사내 앱에 적합합니다. 약점은 라이프사이클 공유입니다. 한 앱 때문에 PostgreSQL을 올리거나 재시작하면 전부 영향을 받습니다.

2. 제품 번들 전용 인스턴스

    제품이 자기 PostgreSQL을 통째로 들고 와서, 전용 OS 계정과 전용 경로에서, 제품 자신의 감독자(supervisor)로 돌립니다. GitLab Omnibus가 `gitlab-psql` 계정과 `gitlab-ctl`로 돌리는 게 이 패턴이고, Chef Infra Server도 `opscode-pgsql` 계정으로 똑같이 합니다. 우연한 유사성이 아니라 GitLab의 Omnibus 패키징 자체가 Chef의 Omnibus 도구에서 온 같은 계보입니다. 이 패턴의 특징은 (1) TCP를 꺼 둔 소켓 전용이 기본이기 쉽고, (2) 설정 파일을 제품의 reconfigure가 소유해서 수동 변경이 덮어써지며, (3) 버전이 제품 릴리스에 묶인다는 점입니다. 그래서 **번들 인스턴스를 다른 앱이 재사용하는 것은 금물**입니다.

3. 컨테이너 전용 인스턴스

    앱 스택마다 PostgreSQL 컨테이너를 하나씩 둡니다. 데이터는 볼륨, 접속은 포트 매핑으로 명시되고, 버전 선택과 폐기가 자유롭습니다. 호스트에 이미 다른 PostgreSQL이 있어도 충돌 축이 포트 매핑 하나로 줄어듭니다. 루트리스 컨테이너면 무권한 계정으로도 운영할 수 있습니다.

4. 외부 또는 관리형 인스턴스

    별도 호스트의 PostgreSQL 또는 RDS 같은 관리형 서비스입니다. 같은 호스트가 아니므로 **Unix 소켓이 없고 TCP만 가능**하며, 따라서 peer 인증도 불가능합니다. 비밀번호(scram) 또는 클라우드 IAM 인증을 씁니다. production 권장 형태인 경우가 많습니다.

| 패턴 | 접속 통로 | 인스턴스 소유자 | 다른 앱과 공유 |
| :--- | :--- | :--- | :--- |
| 시스템 공유 | 소켓 + localhost TCP | OS 관리자 | DB/role 단위로 가능 |
| 제품 번들 | 소켓 전용이 기본 | 제품 (reconfigure) | 금물 |
| 컨테이너 전용 | 매핑된 TCP | 앱 스택 | 스택 밖과는 안 함 |
| 외부/관리형 | TCP만 (소켓 없음) | DBA/클라우드 | 정책에 따라 |

어느 패턴인지 판정하는 질문은 단순합니다. "이 PostgreSQL의 설정 파일을 누가 소유하는가, 그리고 버전 업그레이드를 누가 결정하는가." 답이 "제품의 reconfigure"라면 번들 패턴이고, 그 인스턴스는 그 제품 전용으로 두어야 합니다.

## 같은 호스트에 인스턴스 여러 개를 둘 때

한 호스트에 PostgreSQL 인스턴스가 여러 개 공존하는 것 자체는 정상이며(번들 + 시스템, 번들 + 컨테이너 등), 충돌 축은 네 가지뿐입니다. TCP 포트, 소켓 디렉터리, 데이터 디렉터리, OS 계정. 뒤의 셋은 패턴별로 이미 떨어져 있는 경우가 대부분이라, 실질 충돌 축은 TCP 포트 하나로 줄어듭니다.

여기서 기본 포트 5432는 피하는 편이 좋습니다. 이유는 두 가지입니다. 첫째, 지금 비어 있어도 소켓 전용이던 다른 인스턴스가 설정 변경으로 TCP를 켜는 순간 충돌이 미래에 터집니다. 둘째, 5432는 모든 클라이언트 도구의 기본값이라 "어느 인스턴스에 붙었는지"의 신원 혼동을 만듭니다. 두 번째 인스턴스부터 5433 이상을 주면 포트 번호 자체가 신원 라벨이 됩니다. 이 원칙을 GitLab 호스트에 Nexus용 PostgreSQL을 추가하는 상황에 적용한 사례가 [linux/privilege-isolation.md](../../linux/privilege-isolation.md)의 토폴로지 절에 있습니다.

## teach back 과 검증 경로

이 문서를 덮고 다음을 자기 말로 설명할 수 있으면 핵심을 잡은 것입니다.

- `SHOW port`가 5432인데 `ss -ltnp`에 5432가 없는 상태는 어떻게 가능한가.
- `sudo -u postgres psql`이 비밀번호 없이 붙는 메커니즘은 무엇이고, 같은 일이 TCP 접속에서는 왜 불가능한가.
- 번들 PostgreSQL을 다른 앱이 재사용하면 안 되는 이유 세 가지는 무엇인가.

직접 확인하려면 손에 있는 아무 PostgreSQL에서 위 관측 명령 두 개(`ss -ltnp`, `ss -xlp`)와 `SHOW listen_addresses`를 비교해 보면 됩니다. 컨테이너로 실험한다면 `listen_addresses`를 `''`로 바꿔 재시작한 전후의 `ss` 출력을 비교하는 것이 가장 빠른 체감 경로입니다.

## 근거

- `listen_addresses`, `port`, `unix_socket_directories`의 의미와 기본값: PostgreSQL 공식 문서 [Connections and Authentication](https://www.postgresql.org/docs/current/runtime-config-connection.html). 빈 `listen_addresses`가 소켓 전용을 뜻한다는 것도 이 문서가 명시합니다.
- `local`/`host` 규칙과 인증 방식: [pg_hba.conf](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html), [peer 인증](https://www.postgresql.org/docs/current/auth-peer.html).
- GitLab Omnibus 번들 PG의 소켓 전용 실측: [git/gitlab/gitlab_db.md](../../git/gitlab/gitlab_db.md).
- Chef Infra Server의 `opscode-pgsql` 번들: [Chef Infra Server 문서](https://docs.chef.io/server/).
- 같은 호스트 공존과 포트 분리의 적용 사례: [linux/privilege-isolation.md](../../linux/privilege-isolation.md).
