# 망 분리 AI 개발/배포 아키텍처 발표 설명서

작성일: 2026-04-22
기준 구조도: [`closed_network_ai_arch.html`](closed_network_ai_arch.html)
대상 독자: CTO, 보안 담당자, 인프라 엔지니어, 개발자
목적: HTML 구조도를 보면서 “이 서버들을 왜 이렇게 배치하고, 실제 개발과 배포가 어떤 코드 흐름으로 이어지는지” 설명할 수 있게 만드는 발표 대본 겸 기술 설명서

이 문서는 연구/개발망에서 Claude를 이용해 개발하고, GitLab Runner가 빌드와 테스트를 수행한 뒤, DMZ 완충 구역의 Jenkins를 통해 내부망 서버에 롤링 배포하는 MVP 구조를 설명한다. 여기서 MVP는 “보안 원칙을 포기한 임시 구조”가 아니라, 처음 몇 주 안에 검증할 수 있도록 범위를 줄인 구조라는 뜻이다. 처음부터 모든 자동화와 모든 보안 도구를 붙이지는 않지만, 나중에 서버 수, 프로젝트 수, 배포 대상이 늘어나도 같은 방향으로 확장될 수 있게 핵심 경계는 먼저 고정한다.

이 구조의 핵심은 도구 이름이 아니다. GitLab, Runner, Nexus, Jenkins, L4, MySQL 같은 구성요소보다 먼저 봐야 하는 것은 데이터가 어느 망에서 어느 망으로 흐르는지, 어떤 정보가 밖으로 나가면 안 되는지, 어떤 서버가 어떤 권한을 가져도 되는지다. 구조도는 이 판단을 한 화면에 보여주기 위해 연구/개발망, DMZ, 내부망을 가로로 나누고, 흐름을 `연구/개발망 -> DMZ -> 내부망` 방향으로 고정해 둔다.

## 목차

- [1. 발표에서 먼저 말할 핵심 결론](#1-발표에서-먼저-말할-핵심-결론)
- [2. 이 구조를 지탱하는 세 가지 원칙](#2-이-구조를-지탱하는-세-가지-원칙)
- [3. HTML 구조도 읽는 법](#3-html-구조도-읽는-법)
- [4. 망별 서버 구성 설명](#4-망별-서버-구성-설명)
- [5. 전체 동작 흐름 15단계 발표 대본](#5-전체-동작-흐름-15단계-발표-대본)
- [6. 코드로 보는 실제 연결 방식](#6-코드로-보는-실제-연결-방식)
- [7. Jenkins가 배포 후 무엇을 모니터링해야 하는가](#7-jenkins가-배포-후-무엇을-모니터링해야-하는가)
- [8. 장애와 복구는 어떻게 설계하는가](#8-장애와-복구는-어떻게-설계하는가)
- [9. 운영팀과 합의해야 할 경계](#9-운영팀과-합의해야-할-경계)
- [10. CTO와 인프라팀에게 설명할 때의 요약 문장](#10-cto와-인프라팀에게-설명할-때의-요약-문장)
- [11. 확인한 사실과 이 문서의 추론](#11-확인한-사실과-이-문서의-추론)
- [12. 참고 문서](#12-참고-문서)

## 1. 발표에서 먼저 말할 핵심 결론

이 아키텍처는 연구/개발망에서 코드를 만들고, DMZ에서 배포 후보를 검증하고, 내부망에서 실제 서비스를 교체하는 구조다. 개발자는 평소처럼 Claude로 개발을 보조받고 GitLab에 commit, push, merge request를 올린다. merge 이후 GitLab Runner 풀은 여러 대의 Runner로 빌드와 테스트 부하를 나누고, 테스트가 통과한 결과물만 `artifact bundle`로 묶어 DMZ에 전달한다. DMZ의 Receiver는 이 bundle이 예상한 파일인지, 해시가 맞는지, manifest가 올바른지 확인한 뒤 Shared Drop Zone에 저장한다. 배포 담당자는 DMZ에 있는 Jenkins에서 배포할 artifact를 선택하고, Jenkins Agent는 내부망 서버에 제한된 서비스 계정 권한으로만 배포 스크립트를 실행한다.

발표 첫 문장은 다음처럼 시작하면 된다.

> 이 구조는 “개발망에서 빌드된 검증 가능한 산출물만 DMZ 완충 구역을 거쳐 내부망으로 들어간다”는 흐름입니다. 내부망의 상세 로그, 운영 secret, 서버 목록은 연구/개발망이나 Claude로 되돌아가지 않습니다. GitLab Runner는 빌드와 테스트를 확장하고, Jenkins는 DMZ에서 배포 시점 선택과 최소 상태 모니터링을 담당합니다.

여기서 중요한 표현은 “검증 가능한 산출물”이다. 내부망에 들어가는 것은 개발자의 전체 작업 폴더나 Git repository 전체가 아니라, 이미 빌드된 JAR, 그 JAR의 checksum, 배포 manifest, 테스트 요약이다. 이 네 가지를 하나의 배포 후보로 묶어야 나중에 “어떤 commit에서 만들어진 어떤 파일이 어떤 서버에 배포되었는가”를 추적할 수 있다.

## 2. 이 구조를 지탱하는 세 가지 원칙

### 2.1 제어 영역과 데이터 영역을 분리한다

구조도 상단의 핵심 메시지는 `Control Plane`과 `Data Plane`의 분리다. 제어 영역은 시스템을 조작하고 기록하는 영역이다. GitLab의 사용자, 권한, merge request, pipeline 상태, issue 같은 정보는 GitLab의 제어 정보다. 데이터 영역은 애플리케이션이 실제로 읽고 쓰는 테스트 데이터와 서비스 데이터를 다루는 영역이다.

이 구조에서 GitLab은 PostgreSQL과 Redis, Gitaly를 사용해 소스 관리와 파이프라인 제어를 담당한다. GitLab 공식 아키텍처 문서는 GitLab 애플리케이션이 사용자, 권한, issue 같은 영속 메타데이터를 PostgreSQL에 저장하고, Git object 접근은 Gitaly를 통해 수행한다고 설명한다. 그래서 구조도에서 `GitLab DB HA (PostgreSQL / Redis)`와 `Gitaly Cluster`는 GitLab 플랫폼 자체를 위한 저장소다.

반대로 `R&D MySQL`은 애플리케이션 개발과 통합 테스트를 위한 데이터베이스다. Runner가 통합 테스트 중에 테이블을 만들고 지우거나 테스트 데이터를 대량으로 넣어도, GitLab의 PostgreSQL과 Gitaly에는 영향을 주지 않아야 한다. 이 분리는 단순히 DB 제품이 다르다는 뜻이 아니라, 테스트 부하와 장애가 GitLab 플랫폼으로 전파되지 않게 만드는 장애 격리선이다.

발표에서는 이렇게 말하면 된다.

> GitLab이 쓰는 PostgreSQL은 소스 관리 플랫폼의 상태를 담는 DB이고, R&D MySQL은 우리가 만드는 서비스의 테스트 데이터를 담는 DB입니다. 둘을 섞으면 무거운 통합 테스트가 GitLab 자체를 느리게 만들 수 있고, 테스트 DB 침해가 소스 관리 제어권 침해로 번질 수 있습니다. 그래서 그림에서 GitLab Core HA와 R&D Application DB를 일부러 나눴습니다.

### 2.2 상태를 외부화한다

Receiver, Runner, Jenkins Agent 같은 실행 노드는 가능하면 상태를 오래 보관하지 않는다. 상태가 없는 서버는 죽어도 새 서버로 바꾸기 쉽다. 반대로 Git repository, GitLab DB, Nexus blob, Drop Zone, Jenkins home, 내부 서비스 로그처럼 나중에 반드시 다시 찾아야 하는 데이터는 별도 저장소에 둔다.

이 원칙을 구조도에서는 다음처럼 표현한다.

- GitLab Web/API 노드는 L4 뒤에 여러 대 둘 수 있는 stateless 성격의 노드다.
- Gitaly Cluster는 Git repository 데이터를 저장하는 stateful 영역이다.
- Nexus Node는 여러 대를 둘 수 있지만, package blob과 DB는 공유 저장소나 외부 DB에 남아야 한다.
- DMZ Receiver는 수신과 검증만 수행하고, 검증된 artifact는 Shared Drop Zone에 저장한다.
- Jenkins Agent는 배포 스크립트를 실행하는 노드이고, 배포 이력과 credential은 Jenkins Controller와 승인된 저장소에서 관리한다.

발표에서는 “서버 두 대를 두는 것”과 “고가용성”을 구분해야 한다. 두 대가 모두 자기 로컬 디스크에만 파일을 저장하면 어느 노드로 들어갔는지에 따라 배포 후보가 사라질 수 있다. 고가용성은 실행 노드를 여러 대 두는 것뿐 아니라, 상태가 남는 위치를 명확히 정하는 것이다.

### 2.3 권한은 서비스 계정과 allowlist로 좁힌다

DMZ에 Jenkins가 있다고 해서 Jenkins가 내부망 서버의 root가 되어도 된다는 뜻은 아니다. Jenkins는 배포를 시작할 수 있어야 하지만, 내부 서버에서 임의 명령을 실행할 수 있으면 안 된다. 그래서 구조도에는 `서비스 계정 권한으로 실행`이라는 표현이 들어간다.

여기서 말하는 서비스 계정은 특정 애플리케이션 배포에만 쓰는 OS 계정이다. 예를 들어 `svc-myapp` 계정은 `/opt/myapp` 아래 release 파일을 교체하고 `myapp.service`를 재시작할 권한만 가진다. Jenkins가 SSH로 접속하더라도, 실제 실행 가능한 명령은 `/opt/myapp/bin/deploy.sh`와 `/opt/myapp/bin/rollback.sh`처럼 미리 허용된 스크립트로 제한한다.

발표에서는 다음처럼 설명하면 된다.

> Jenkins는 내부망에 “아무 명령이나 실행하는 원격 관리자”로 들어가지 않습니다. Jenkins가 할 수 있는 일은 특정 artifact를 특정 배포 스크립트에 넘기는 것뿐입니다. 내부 서버의 상세 로그를 긁어오거나 DB dump를 만들거나 서버 목록을 조회하는 권한은 주지 않습니다.

이 원칙은 Jenkins credential masking만으로는 충분하지 않다는 점과 연결된다. Jenkins 공식 블로그는 credential masking이 로그에 secret이 그대로 찍히는 사고를 줄여줄 수는 있지만, build script가 의도적으로 다른 방식으로 secret을 노출하는 것까지 막을 수는 없다고 설명한다. 그래서 secret을 숨기는 기능보다 더 중요한 것은 credential을 사용할 수 있는 job, folder, agent, command 자체를 좁히는 것이다.

## 3. HTML 구조도 읽는 법

HTML은 세 개의 큰 세로 영역으로 구성되어 있다.

1. 연구/개발망

    개발자가 Claude로 개발을 보조받고 GitLab에 코드를 올리는 영역이다. GitLab Core HA, R&D MySQL, GitLab Runner Pool, Nexus Repository HA가 이 영역에 있다. 여기서 빌드와 테스트가 끝나면 artifact bundle을 DMZ로 보낸다.

2. DMZ 완충 구역

    외부성 있는 연구/개발망과 보호해야 할 내부망 사이에 있는 완충 지대다. 여기에는 Artifact Receiver, Shared Drop Zone, Jenkins Controller, Jenkins Agent가 있다. DMZ는 내부망으로 들어가는 파일을 잠깐 멈춰 세우고 검증하는 장소이며, 배포 시점 선택과 최소 상태 모니터링을 제공한다.

3. 내부망

    실제 서비스가 실행되는 서버들이 있는 영역이다. Internal L4 Switch 뒤에 Target Server A/B가 있고, 배포는 Drain, Deploy, Health Check, Join 순서로 한 대씩 진행된다. 상세 서비스 로그는 내부망 전용 저장소에만 남긴다.

화살표 색은 다음 의미로 읽으면 된다.

| 표시 | 의미 | 발표에서의 설명 |
| --- | --- | --- |
| 파란 실선 | 연구/개발망 안의 build, test, artifact 생성 흐름 | 개발과 검증이 끝난 파일만 다음 단계로 넘어간다. |
| 초록 실선 | DMZ에서 내부망으로 승인된 배포 흐름 | Jenkins가 승인된 스크립트를 호출한다. |
| 주황 표시 | 최소 상태 응답 | 내부망 상세 로그가 아니라 health check 결과 수준만 DMZ Jenkins에 보인다. |
| 빨간 점선/차단 표시 | 금지된 역방향 흐름 | 내부망 로그, secret, 서버 목록이 연구/개발망이나 Claude로 가지 않는다. |

DMZ는 “무조건 안전한 내부망”도 아니고 “그냥 외부망”도 아니다. 이 문서에서는 DMZ를 외부 쪽 시스템과 내부망 사이에서 제한된 프로토콜만 허용하는 완충 구역으로 사용한다. 연구/개발망에서 DMZ로 들어오는 것은 artifact bundle이고, DMZ에서 내부망으로 들어가는 것은 승인된 배포 명령과 artifact다. 내부망에서 DMZ로 돌아오는 것은 원칙적으로 최소 상태값뿐이다.

## 4. 망별 서버 구성 설명

### 4.1 연구/개발망: 개발, 빌드, 테스트, 패키지 생성

연구/개발망은 개발자가 가장 많이 접하는 영역이다. 여기서 Claude는 개발 보조 도구로 사용된다. Claude가 접근할 수 있는 범위는 source code, 테스트 코드, 빌드 오류, 일반적인 개발 문서 수준으로 제한한다. 운영 DB 접속 정보, 내부망 서버명, 운영 로그, 고객 데이터 샘플은 Claude 프롬프트에 들어가면 안 된다.

GitLab Core HA는 다음 네 가지로 나누어 이해하면 쉽다.

| 구성요소 | 역할 | 주의점 |
| --- | --- | --- |
| L4 GitLab VIP | 개발자가 접속하는 GitLab 대표 주소 | Web/API 노드 장애 시 다른 노드로 라우팅한다. |
| GitLab Web/API 노드 | MR, review, pipeline 화면과 API 처리 | 가급적 stateless로 두고 상태는 DB/Gitaly/Redis에 둔다. |
| GitLab DB HA | 사용자, 권한, MR, pipeline 같은 플랫폼 메타데이터 저장 | 애플리케이션 통합 테스트 DB로 쓰면 안 된다. |
| Gitaly Cluster | Git repository object를 저장하고 Git 작업을 처리 | Git 저장소 고가용성 영역이며, 일반 파일 공유 폴더처럼 직접 건드리면 안 된다. |

Gitaly Cluster가 낯설 수 있다. 간단히 말하면 Gitaly는 Git repository에 대한 읽기/쓰기 작업을 GitLab 애플리케이션 대신 처리하는 Git 저장소 서비스다. Praefect를 사용하는 Gitaly Cluster는 여러 Gitaly node를 하나의 virtual storage처럼 보이게 하고, Git 작업을 가능한 Gitaly node로 라우팅하며, repository write를 복제해 장애 허용성을 높인다. GitLab 공식 문서는 Gitaly Cluster가 Git 저장소를 여러 Gitaly node에 저장할 수 있고, Praefect가 라우터와 transaction manager 역할을 한다고 설명한다.

R&D Application DB는 GitLab DB와 별도다. 예를 들어 Spring Boot 서비스가 MySQL을 쓴다면, GitLab Runner는 통합 테스트 때 `jdbc:mysql://rnd-mysql-ha-vip:3306/myapp_test` 같은 R&D 전용 MySQL에 붙는다. 이 DB에는 production data를 넣지 않는다. 테스트 fixture와 개발용 seed data만 둔다.

Nexus Repository는 회사 자체 라이브러리와 외부 Maven dependency를 제공한다. Maven Central에 올릴 수 없는 사내 공통 라이브러리는 Nexus의 hosted repository에 올리고, 외부 dependency는 proxy repository로 받아 캐시한다. 개발자와 Runner는 `maven-public` 같은 group repository 하나만 바라보게 하면, 사내 library와 외부 library를 같은 Maven 설정으로 받을 수 있다.

GitLab Runner Pool은 부하 분산의 핵심이다. Runner가 한 대뿐이면 통합 테스트가 몰릴 때 pipeline queue가 길어진다. Runner를 여러 대 두고 tag를 나누면, 테스트 job과 packaging job을 여러 서버로 분산할 수 있다. GitLab Runner 공식 문서는 runner fleet 운영에서 `concurrent`, `limit`, `request_concurrency` 같은 설정으로 job 흐름을 제어한다고 설명한다.

### 4.2 DMZ: 검증, 보관, 배포 승인

DMZ에는 두 가지 종류의 서버가 있다.

첫 번째는 Artifact Receiver다. Receiver는 GitLab Runner가 보낸 bundle을 받는다. 하지만 받았다고 곧바로 배포 후보로 인정하지 않는다. 먼저 manifest가 있는지, manifest 안의 project와 commit이 허용된 값인지, checksum이 실제 JAR과 일치하는지, tar 압축 안에 이상한 경로가 없는지 확인한다. 검증 전 파일은 `incoming` 또는 `quarantine` 영역에 두고, 검증 통과 후에만 `verified` 영역으로 이동한다.

두 번째는 Jenkins다. 이 구조에서 Jenkins는 build tool이 아니라 CD tool이다. 이미 GitLab Runner가 build와 test를 끝냈기 때문에, Jenkins는 JAR을 다시 빌드하지 않는다. Jenkins가 하는 일은 검증된 artifact 후보를 보여주고, 배포 담당자가 선택한 시점에 내부망 배포 스크립트를 실행하며, 각 서버의 최소 health check 결과를 보여주는 것이다.

Jenkins Controller와 Agent가 모두 DMZ에 있다는 점도 중요하다. Controller는 UI, job 설정, credential store, job history를 가진다. Agent는 실제 shell 명령을 실행한다. Jenkins 공식 문서는 Controller가 직접 build까지 실행하는 standalone 방식은 규모가 커질수록 리소스와 보안 문제가 생기므로 agent로 workload를 위임하는 구조를 설명한다. 이 설계도 같은 이유로 Controller와 Agent 역할을 나눈다.

다만 Jenkins HA는 단순한 active-active가 아니다. Jenkins Controller는 `$JENKINS_HOME`, plugin, job config, build history, credential store 같은 상태를 가진다. 따라서 Active/Standby 구조를 쓸 때는 공유 스토리지, 백업, 복구 절차, plugin 버전 관리, failover 정책을 별도로 설계해야 한다. 구조도는 “Jenkins Master Active/Standby”라고 간략히 표현하지만, 실제 도입 때는 Jenkins 운영 방식 자체가 별도 검토 항목이다.

### 4.3 내부망: 실제 서비스 배포와 내부 로그 보관

내부망은 실제 서비스가 실행되는 곳이다. 구조도는 단순화를 위해 Target Server A/B 두 대만 보여준다. 실제로는 서비스별 서버 그룹이 여러 개 있을 수 있고, 각 그룹 앞에는 L4나 reverse proxy가 있을 수 있다.

롤링 배포는 한 번에 모든 서버를 바꾸지 않는다. 먼저 Server A를 L4에서 제외한다. 이것을 Drain이라고 한다. Drain이 끝나면 Server A에는 신규 요청이 들어가지 않으므로, Jenkins Agent가 해당 서버에 artifact를 복사하고 배포 스크립트를 실행한다. 서비스가 재시작되고 `/actuator/health`가 200을 반환하면 Server A를 다시 L4에 붙인다. 그 다음 Server B에 같은 절차를 반복한다.

내부망 서비스 로그는 내부망에 남는다. Jenkins에는 전체 stack trace를 가져오지 않는다. Jenkins에 남겨도 되는 것은 `Server A 배포 성공`, `Server B health check 실패`, `전체 성공`, `rollback 실행` 같은 요약 상태다. 실패 원인을 자세히 보려면 내부망 로그 저장소나 내부 운영 포털에서 확인해야 한다.

## 5. 전체 동작 흐름 15단계 발표 대본

아래 표는 HTML의 `Main Sequence`를 발표 대본 형태로 풀어 쓴 것이다. `발표 멘트`는 CTO나 인프라팀 앞에서 그대로 말해도 되는 문장이고, `기술 설명`은 질문이 들어왔을 때 덧붙일 내용이다.

| 단계 | 발표 멘트 | 기술 설명 |
| ---: | --- | --- |
| 01 | 개발자는 연구/개발망에서 Claude를 개발 보조로 사용합니다. | Claude는 코드 작성, 테스트 아이디어, 리뷰 보조까지만 담당한다. 내부망 접속, 운영 로그 분석, production secret 주입은 금지한다. |
| 02 | 개발자는 평소처럼 commit하고 GitLab에 push합니다. | source code는 연구/개발망 GitLab에 저장된다. 운영 secret이 repository에 들어가지 않도록 secret scanning과 review 정책을 둔다. |
| 03 | GitLab에서 MR을 만들고 리뷰 후 merge합니다. | 누가 승인했는지, 어떤 commit이 merge되었는지는 GitLab 제어 영역에 기록된다. Git object는 Gitaly에 저장된다. |
| 04 | merge 이후 GitLab pipeline이 시작됩니다. | pipeline trigger는 build/test 시작 신호다. 아직 내부망 배포가 시작된 것은 아니다. |
| 05 | Runner 풀에서 여러 Runner가 빌드와 통합 테스트를 나누어 수행합니다. | Runner는 R&D MySQL을 사용해 테스트한다. GitLab PostgreSQL을 테스트 DB로 쓰지 않는다. |
| 06 | Runner는 Nexus를 통해 사내 라이브러리와 외부 dependency를 받습니다. | 사내 라이브러리는 Nexus hosted repository, 외부 dependency는 proxy/group repository로 관리한다. |
| 07 | 테스트가 통과하면 JAR, checksum, manifest, 테스트 요약을 묶습니다. | 이 bundle이 DMZ로 넘어가는 배포 후보 단위다. manifest는 commit, pipeline, artifact digest를 연결한다. |
| 08 | Runner가 bundle을 DMZ Receiver로 POST 전송합니다. | 이 단계는 연구/개발망에서 DMZ 방향이다. 내부망에서 연구/개발망으로 돌아가는 흐름이 아니다. |
| 09 | DMZ L4가 살아있는 Receiver 중 하나로 upload를 전달합니다. | Receiver는 stateless로 두고, 어느 Receiver가 받아도 같은 검증 결과가 나오게 한다. |
| 10 | Receiver가 manifest와 checksum을 검증한 뒤 Shared Drop Zone에 저장합니다. | 검증 전 파일은 배포 후보가 아니다. 검증 통과 파일만 `verified` 위치로 atomic move한다. |
| 11 | Jenkins Controller가 verified 후보를 배포 가능 목록으로 보여줍니다. | Jenkins는 여기서 build하지 않는다. 이미 빌드된 후보를 선택하는 CD 도구로 동작한다. |
| 12 | 승인자는 Jenkins에서 배포 타이밍을 선택합니다. | MVP에서는 자동 배포보다 수동 승인으로 시작하는 편이 설명과 통제가 쉽다. |
| 13 | Jenkins Agent가 내부망 서버에 제한된 서비스 계정 권한으로 배포 스크립트를 호출합니다. | 임의 shell을 주지 않고 allowlist된 command만 실행한다. SSH key와 sudoers 정책이 핵심이다. |
| 14 | 내부 L4에서 서버를 하나씩 빼고, 배포하고, health check 후 다시 붙입니다. | Drain, deploy, health check, join 순서로 rolling deploy한다. 실패하면 다음 서버로 넘어가지 않는다. |
| 15 | Jenkins는 최소 배포 상태만 보여주고, 상세 로그는 내부망에 남깁니다. | 200 OK, DEPLOY_OK, DEPLOY_FAIL 같은 요약만 DMZ에 남긴다. stack trace와 운영 secret은 반출하지 않는다. |

이 흐름은 다음 한 줄로 압축할 수 있다.

```text
Claude 개발 보조
-> GitLab MR/merge
-> Runner pool build/test
-> Nexus dependency
-> artifact bundle 생성
-> DMZ Receiver 검증
-> Shared Drop Zone 저장
-> Jenkins 수동 승인
-> Jenkins Agent가 내부망 배포 스크립트 호출
-> L4 rolling deploy
-> Jenkins 최소 상태 모니터링
```

## 6. 코드로 보는 실제 연결 방식

이 절의 코드는 실제 운영 코드가 아니라 PoC에서 흐름을 설명하기 위한 골격이다. 운영 반영 전에는 사내 네트워크, 인증서, 계정 정책, sudoers 정책, L4 API, Nexus repository 이름에 맞게 수정해야 한다.

### 6.1 Nexus를 Maven dependency 경계로 쓰는 설정

사내 라이브러리를 Maven Central에 올릴 수 없다면, 연구/개발망의 Nexus에 hosted repository를 둔다. 외부 dependency도 Runner가 직접 인터넷에서 받지 않고 Nexus proxy/group repository를 통하게 하면 의존성 경로가 통제된다.

Maven `settings.xml` 예시는 다음과 같다.

```xml
<!-- .ci/settings-nexus.xml -->
<settings xmlns="http://maven.apache.org/SETTINGS/1.2.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.2.0 https://maven.apache.org/xsd/settings-1.2.0.xsd">
    <mirrors>
        <mirror>
            <id>company-maven-public</id>
            <name>Company Nexus Maven Group</name>
            <url>https://nexus-rnd.example.local/repository/maven-public/</url>
            <mirrorOf>*</mirrorOf>
        </mirror>
    </mirrors>

    <servers>
        <server>
            <id>company-releases</id>
            <username>${env.NEXUS_DEPLOY_USER}</username>
            <password>${env.NEXUS_DEPLOY_PASSWORD}</password>
        </server>
    </servers>
</settings>
```

사내 공통 라이브러리를 배포할 때는 release와 snapshot을 분리한다.

```bash
mvn deploy \
  -s .ci/settings-nexus.xml \
  -DaltDeploymentRepository=company-releases::default::https://nexus-rnd.example.local/repository/maven-releases/
```

이 단계에서 주의할 점은 Nexus가 연구/개발망 빌드 의존성을 제공한다는 것이다. 내부망 운영 서버가 배포 시점마다 연구/개발망 Nexus에 직접 붙어 dependency를 받는 구조는 만들지 않는다. 내부망에는 이미 빌드가 끝난 실행 artifact만 들어간다.

### 6.2 Runner 풀 구성 예시

Runner를 여러 대 두면 빌드와 테스트 부하를 나눌 수 있다. GitLab Runner의 `config.toml`에서 `concurrent`는 한 Runner manager가 동시에 처리할 수 있는 job 수를 제한한다. 아래 예시는 개념을 보여주기 위한 값이다.

```toml
# /etc/gitlab-runner/config.toml
concurrent = 6
check_interval = 3

[[runners]]
  name = "rnd-runner-01"
  url = "https://gitlab-rnd.example.local"
  token = "__RUNNER_AUTH_TOKEN__"
  executor = "shell"
  limit = 2
  [runners.custom_build_dir]
  [runners.cache]

[[runners]]
  name = "rnd-runner-02"
  url = "https://gitlab-rnd.example.local"
  token = "__RUNNER_AUTH_TOKEN__"
  executor = "shell"
  limit = 2

[[runners]]
  name = "rnd-runner-03"
  url = "https://gitlab-rnd.example.local"
  token = "__RUNNER_AUTH_TOKEN__"
  executor = "shell"
  limit = 2
```

보안상 더 엄격하게 가려면 shell executor 대신 Docker/Kubernetes executor를 검토한다. GitLab Runner 보안 문서는 shell executor가 Runner host와 network에 높은 위험을 만들 수 있고, self-managed runner는 job에 정의된 코드를 실행하므로 별도 network segment에서 운영하는 방안을 권고한다. MVP에서도 Runner host에 내부망 SSH key나 production secret을 두지 않는 것이 중요하다.

### 6.3 GitLab CI 파이프라인 예시

GitLab CI는 build/test/package/relay 네 단계로 구성한다. 핵심은 build 결과물을 단순히 JAR 하나로 보내지 않고, checksum과 manifest를 함께 만드는 것이다.

```yaml
# .gitlab-ci.yml
stages:
  - test
  - package
  - relay

variables:
  MAVEN_SETTINGS: ".ci/settings-nexus.xml"
  APP_NAME: "myapp"
  DMZ_UPLOAD_URL: "https://dmz-artifact-receiver.example.local/api/v1/artifacts"

default:
  tags:
    - rnd-runner

test:
  stage: test
  script:
    - export SPRING_DATASOURCE_URL="jdbc:mysql://rnd-mysql-ha-vip.example.local:3306/myapp_test"
    - export SPRING_DATASOURCE_USERNAME="$RND_DB_USER"
    - export SPRING_DATASOURCE_PASSWORD="$RND_DB_PASSWORD"
    - mvn -B clean verify -s "$MAVEN_SETTINGS"
  artifacts:
    when: always
    expire_in: 7 days
    paths:
      - target/surefire-reports/
      - target/failsafe-reports/

package:
  stage: package
  needs:
    - job: test
      artifacts: true
  script:
    - mvn -B -DskipTests package -s "$MAVEN_SETTINGS"
    - export VERSION="$(mvn -q -DforceStdout help:evaluate -Dexpression=project.version)"
    - export ARTIFACT="${APP_NAME}-${VERSION}-${CI_COMMIT_SHORT_SHA}.jar"
    - mkdir -p dist/bundle
    - cp target/*.jar "dist/bundle/${ARTIFACT}"
    - cd dist/bundle
    - sha256sum "${ARTIFACT}" | tee "${ARTIFACT}.sha256"
    - |
      jq -n \
        --arg project "$CI_PROJECT_PATH" \
        --arg commit "$CI_COMMIT_SHA" \
        --arg branch "$CI_COMMIT_REF_NAME" \
        --arg pipeline "$CI_PIPELINE_ID" \
        --arg job "$CI_JOB_ID" \
        --arg artifact "$ARTIFACT" \
        --arg sha256 "$(cut -d ' ' -f 1 "${ARTIFACT}.sha256")" \
        --arg createdAt "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        '{
          schema_version: "deploy-manifest/v1",
          project: $project,
          commit_sha: $commit,
          source_ref: $branch,
          pipeline_id: $pipeline,
          package_job_id: $job,
          artifact: {
            file: $artifact,
            sha256: $sha256,
            type: "spring-boot-jar"
          },
          created_at: $createdAt,
          producer: "gitlab-runner-rnd"
        }' > deploy-manifest.json
    - |
      jq -n \
        --arg result "passed" \
        --arg report "target/failsafe-reports" \
        '{
          result: $result,
          detail_policy: "summary-only",
          raw_log_exported: false,
          report_path_hint: $report
        }' > test-result-summary.json
    - tar -czf "../${APP_NAME}-${CI_COMMIT_SHORT_SHA}.bundle.tgz" .
  artifacts:
    expire_in: 14 days
    access: maintainer
    paths:
      - dist/*.bundle.tgz

relay_to_dmz:
  stage: relay
  needs:
    - job: package
      artifacts: true
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: on_success
  script:
    - export BUNDLE="dist/${APP_NAME}-${CI_COMMIT_SHORT_SHA}.bundle.tgz"
    - |
      curl --fail --show-error --silent \
        --request POST \
        --header "X-Upload-Token: ${DMZ_UPLOAD_TOKEN}" \
        --form "project=${CI_PROJECT_PATH}" \
        --form "commit_sha=${CI_COMMIT_SHA}" \
        --form "bundle=@${BUNDLE}" \
        "${DMZ_UPLOAD_URL}"
```

이 pipeline에서 production DB password, 내부망 SSH key, 내부 서버 목록은 등장하지 않는다. Runner가 알아야 하는 것은 R&D MySQL, Nexus, DMZ Receiver 주소뿐이다. DMZ Receiver 주소도 내부 서버 주소가 아니라 완충 구역의 upload endpoint다.

`artifacts: access: maintainer`는 GitLab job artifact 접근을 줄이기 위한 예시다. GitLab 공식 문서는 job artifact가 build output이나 report file을 보관하고, `artifacts:access`로 누가 다운로드할 수 있는지 제한할 수 있다고 설명한다.

### 6.4 배포 manifest 예시

manifest는 사람이 보기 위한 설명서이면서, 자동화가 검증할 계약서다. 최소한 project, commit, pipeline, artifact file, sha256, 생성 시각을 담아야 한다.

```json
{
  "schema_version": "deploy-manifest/v1",
  "project": "group/myapp",
  "commit_sha": "1f2e3d4c5b6a7890abcdef1234567890abcdef12",
  "source_ref": "main",
  "pipeline_id": "123456",
  "package_job_id": "456789",
  "artifact": {
    "file": "myapp-1.4.2-1f2e3d4c.jar",
    "sha256": "9f3d2c0a6b4e2a1c8f0e0a9b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c8d",
    "type": "spring-boot-jar"
  },
  "created_at": "2026-04-22T09:30:00Z",
  "producer": "gitlab-runner-rnd"
}
```

이 manifest가 없으면 DMZ Receiver와 Jenkins는 파일 이름만 보고 배포해야 한다. 파일 이름은 쉽게 속일 수 있다. manifest와 checksum을 함께 보면 “이 파일이 어떤 commit과 pipeline에서 만들어졌는지”를 비교할 수 있다. 더 높은 보안 수준에서는 여기서 끝내지 않고 SLSA provenance나 signature도 추가한다. SLSA 문서는 artifact와 provenance가 진짜인지 확인하려면 builder identity, provenance envelope signature, artifact digest와 subject 일치 여부, buildType과 externalParameters 기대값을 확인하라고 설명한다.

### 6.5 DMZ Receiver 검증 스크립트 예시

DMZ Receiver는 upload endpoint와 검증기를 나누어 생각하면 쉽다. HTTP server는 파일을 받아 임시 위치에 저장한다. 검증기는 그 파일이 배포 후보가 될 수 있는지 검사한다.

아래는 검증기 골격이다.

```bash
#!/usr/bin/env bash
# /opt/dmz-receiver/bin/verify-and-promote.sh
set -euo pipefail

UPLOAD_ID="$1"
INCOMING_DIR="/var/dmz-receiver/incoming/${UPLOAD_ID}"
WORK_DIR="/var/dmz-receiver/work/${UPLOAD_ID}"
DROP_ZONE="/mnt/idc-shared-drop-zone"
PROJECT_ALLOWLIST="/etc/dmz-receiver/project-allowlist.txt"

BUNDLE="${INCOMING_DIR}/bundle.tgz"

mkdir -p "$WORK_DIR"

if [ ! -f "$BUNDLE" ]; then
  echo "REJECT missing_bundle"
  exit 10
fi

if tar -tzf "$BUNDLE" | grep -E '(^/|(^|/)\.\.(/|$))' >/dev/null; then
  echo "REJECT unsafe_tar_path"
  mkdir -p "${DROP_ZONE}/rejected/${UPLOAD_ID}"
  cp "$BUNDLE" "${DROP_ZONE}/rejected/${UPLOAD_ID}/"
  exit 11
fi

tar -xzf "$BUNDLE" -C "$WORK_DIR"

MANIFEST="${WORK_DIR}/deploy-manifest.json"
if [ ! -f "$MANIFEST" ]; then
  echo "REJECT missing_manifest"
  exit 12
fi

PROJECT="$(jq -r '.project // empty' "$MANIFEST")"
COMMIT_SHA="$(jq -r '.commit_sha // empty' "$MANIFEST")"
ARTIFACT_FILE="$(jq -r '.artifact.file // empty' "$MANIFEST")"
EXPECTED_SHA="$(jq -r '.artifact.sha256 // empty' "$MANIFEST")"

if ! grep -Fx "$PROJECT" "$PROJECT_ALLOWLIST" >/dev/null; then
  echo "REJECT project_not_allowed"
  exit 13
fi

if ! [[ "$COMMIT_SHA" =~ ^[0-9a-f]{40}$ ]]; then
  echo "REJECT invalid_commit_sha"
  exit 14
fi

if ! [[ "$ARTIFACT_FILE" =~ ^[A-Za-z0-9._-]+\.jar$ ]]; then
  echo "REJECT invalid_artifact_name"
  exit 15
fi

ACTUAL_SHA="$(sha256sum "${WORK_DIR}/${ARTIFACT_FILE}" | awk '{print $1}')"
if [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
  echo "REJECT checksum_mismatch"
  mkdir -p "${DROP_ZONE}/rejected/${PROJECT}/${COMMIT_SHA}"
  cp -r "$WORK_DIR" "${DROP_ZONE}/rejected/${PROJECT}/${COMMIT_SHA}/${UPLOAD_ID}"
  exit 16
fi

PROMOTE_DIR="${DROP_ZONE}/verified/${PROJECT}/${COMMIT_SHA}"
STAGING_DIR="${DROP_ZONE}/.staging/${PROJECT}/${COMMIT_SHA}.${UPLOAD_ID}"

mkdir -p "$(dirname "$STAGING_DIR")" "$(dirname "$PROMOTE_DIR")"
rm -rf "$STAGING_DIR"
cp -R "$WORK_DIR" "$STAGING_DIR"
mv "$STAGING_DIR" "$PROMOTE_DIR"

echo "VERIFY_OK project=${PROJECT} commit=${COMMIT_SHA}"
```

여기서 중요한 부분은 `mv "$STAGING_DIR" "$PROMOTE_DIR"`다. 검증 중인 파일을 Jenkins가 읽으면 안 된다. staging 위치에 완전히 복사한 뒤 마지막에 디렉터리를 한 번에 바꾸면, Jenkins는 검증 완료 전 상태를 보지 않는다.

HTTP upload server는 어떤 언어로 만들어도 된다. PoC에서는 아래처럼 단순한 Python 골격으로 시작할 수 있다. 운영에서는 mTLS, WAF, request size limit, audit log, rate limit, virus scan 연동을 추가해야 한다.

```python
# /opt/dmz-receiver/app/receiver.py
from pathlib import Path
from uuid import uuid4
import os
import subprocess

from flask import Flask, abort, request

app = Flask(__name__)

UPLOAD_ROOT = Path("/var/dmz-receiver/incoming")
EXPECTED_TOKEN = os.environ["DMZ_UPLOAD_TOKEN"]


@app.post("/api/v1/artifacts")
def upload_artifact():
    token = request.headers.get("X-Upload-Token", "")
    if token != EXPECTED_TOKEN:
        abort(401)

    bundle = request.files.get("bundle")
    project = request.form.get("project", "")
    commit_sha = request.form.get("commit_sha", "")

    if not bundle or "/" not in project or len(commit_sha) != 40:
        abort(400)

    upload_id = str(uuid4())
    target_dir = UPLOAD_ROOT / upload_id
    target_dir.mkdir(parents=True, exist_ok=False)
    bundle.save(target_dir / "bundle.tgz")

    subprocess.run(
        ["/opt/dmz-receiver/bin/verify-and-promote.sh", upload_id],
        check=True,
        text=True,
    )

    return {"result": "accepted", "upload_id": upload_id}, 202
```

이 endpoint는 내부망 정보를 반환하지 않는다. 실패하더라도 `checksum_mismatch` 같은 거친 코드만 남기고, 내부 서버명이나 배포 상태는 응답하지 않는다.

### 6.6 Jenkins Pipeline 예시

Jenkins는 Drop Zone의 verified 후보를 읽고, 사용자가 배포할 후보와 대상 그룹을 고르게 한다. MVP에서는 `input` step으로 수동 승인을 둔다.

```groovy
// Jenkinsfile
pipeline {
    agent { label 'dmz-deploy-agent' }

    options {
        timestamps()
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '30'))
    }

    parameters {
        string(name: 'PROJECT', defaultValue: 'group/myapp', description: 'GitLab project path')
        string(name: 'COMMIT_SHA', defaultValue: '', description: '40-char commit SHA to deploy')
        choice(name: 'TARGET_GROUP', choices: ['myapp-test', 'myapp-stage'], description: 'Internal target group alias')
    }

    environment {
        DROP_ZONE = '/mnt/idc-shared-drop-zone/verified'
    }

    stages {
        stage('Load Candidate') {
            steps {
                sh '''
                  set -euo pipefail
                  case "$COMMIT_SHA" in
                    (*[!0-9a-f]*|'') echo "Invalid COMMIT_SHA"; exit 2 ;;
                  esac

                  CANDIDATE="${DROP_ZONE}/${PROJECT}/${COMMIT_SHA}"
                  test -f "${CANDIDATE}/deploy-manifest.json"
                  jq -e '.schema_version == "deploy-manifest/v1"' "${CANDIDATE}/deploy-manifest.json" >/dev/null
                  jq -r '.artifact.file' "${CANDIDATE}/deploy-manifest.json" > artifact-name.txt
                  echo "$CANDIDATE" > candidate-path.txt
                '''
            }
        }

        stage('Manual Approval') {
            steps {
                input message: "Deploy ${params.PROJECT}@${params.COMMIT_SHA} to ${params.TARGET_GROUP}?",
                      ok: 'Deploy'
            }
        }

        stage('Rolling Deploy') {
            steps {
                withCredentials([
                    sshUserPrivateKey(
                        credentialsId: 'dmz-jenkins-to-internal-deploy',
                        keyFileVariable: 'SSH_KEY',
                        usernameVariable: 'SSH_USER'
                    )
                ]) {
                    sh '''
                      set -euo pipefail
                      ./scripts/orchestrate-rolling-deploy.sh \
                        "$(cat candidate-path.txt)" \
                        "$TARGET_GROUP" \
                        "$SSH_USER" \
                        "$SSH_KEY"
                    '''
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'deployment-summary.json', allowEmptyArchive: true
        }
    }
}
```

이 Jenkinsfile에서 주목할 부분은 `PROJECT`, `COMMIT_SHA`, `TARGET_GROUP`이다. Jenkins가 내부 서버명을 직접 파라미터로 받지 않는다. 대상은 `myapp-test` 같은 alias로 받는다. 실제 서버 목록은 DMZ Jenkins script 또는 내부망의 제한된 inventory에서 해석한다. 이렇게 해야 사용자가 Jenkins 파라미터에 임의 host를 넣어 secret을 빼내는 위험을 줄일 수 있다.

### 6.7 DMZ Jenkins Agent의 롤링 배포 스크립트 예시

이 스크립트는 하나의 target group에 대해 서버를 순서대로 drain, deploy, health check, join한다. L4 API는 회사 장비마다 다르므로 placeholder로 둔다.

```bash
#!/usr/bin/env bash
# scripts/orchestrate-rolling-deploy.sh
set -euo pipefail

CANDIDATE_PATH="$1"
TARGET_GROUP="$2"
SSH_USER="$3"
SSH_KEY="$4"

case "$TARGET_GROUP" in
  myapp-test)
    TARGET_NODES=("myapp-test-a.internal.example.local" "myapp-test-b.internal.example.local")
    HEALTH_PORT="8080"
    ;;
  myapp-stage)
    TARGET_NODES=("myapp-stage-a.internal.example.local" "myapp-stage-b.internal.example.local")
    HEALTH_PORT="8080"
    ;;
  *)
    echo "DEPLOY_FAIL unknown_target_group"
    exit 20
    ;;
esac

MANIFEST="${CANDIDATE_PATH}/deploy-manifest.json"
ARTIFACT_FILE="$(jq -r '.artifact.file' "$MANIFEST")"
COMMIT_SHA="$(jq -r '.commit_sha' "$MANIFEST")"

SUMMARY="$(pwd)/deployment-summary.json"
jq -n \
  --arg target "$TARGET_GROUP" \
  --arg commit "$COMMIT_SHA" \
  '{target_group: $target, commit_sha: $commit, nodes: []}' > "$SUMMARY"

for NODE in "${TARGET_NODES[@]}"; do
  echo "DEPLOY_NODE_START node_alias=${NODE%%.*}"

  ./scripts/l4-drain.sh "$TARGET_GROUP" "$NODE"

  scp -i "$SSH_KEY" -o StrictHostKeyChecking=yes \
    "$MANIFEST" "${CANDIDATE_PATH}/${ARTIFACT_FILE}" "${CANDIDATE_PATH}/${ARTIFACT_FILE}.sha256" \
    "${SSH_USER}@${NODE}:/var/tmp/myapp-incoming/"

  ssh -i "$SSH_KEY" -o StrictHostKeyChecking=yes "${SSH_USER}@${NODE}" \
    "sudo -n -u svc-myapp -- /opt/myapp/bin/deploy.sh '${COMMIT_SHA}'"

  STATUS="$(curl -sS -o /tmp/health.out -w '%{http_code}' "http://${NODE}:${HEALTH_PORT}/actuator/health" || true)"
  if [ "$STATUS" != "200" ]; then
    ./scripts/l4-join.sh "$TARGET_GROUP" "$NODE" "skip" || true
    jq --arg node "${NODE%%.*}" \
       --arg status "failed" \
       '.nodes += [{node_alias: $node, status: $status}]' "$SUMMARY" > "${SUMMARY}.tmp"
    mv "${SUMMARY}.tmp" "$SUMMARY"
    echo "DEPLOY_FAIL node_alias=${NODE%%.*} health_status=${STATUS}"
    exit 21
  fi

  ./scripts/l4-join.sh "$TARGET_GROUP" "$NODE"

  jq --arg node "${NODE%%.*}" \
     --arg status "ok" \
     '.nodes += [{node_alias: $node, status: $status}]' "$SUMMARY" > "${SUMMARY}.tmp"
  mv "${SUMMARY}.tmp" "$SUMMARY"

  echo "DEPLOY_NODE_OK node_alias=${NODE%%.*}"
done

jq '.result = "success"' "$SUMMARY" > "${SUMMARY}.tmp"
mv "${SUMMARY}.tmp" "$SUMMARY"
echo "DEPLOY_OK target_group=${TARGET_GROUP} commit=${COMMIT_SHA}"
```

여기서 Jenkins가 가져오는 health check는 HTTP status 수준이다. `/actuator/health` 응답 body가 DB 이름, 내부 dependency 이름, exception message를 과하게 담는다면 Jenkins에는 status code만 보이게 하고 상세 body는 내부망 로그로만 남긴다.

### 6.8 내부망 서버의 배포 스크립트 예시

내부 서버에 있는 `deploy.sh`는 서비스 계정으로 실행된다. 이 스크립트는 artifact checksum을 다시 확인하고, release 디렉터리를 만들고, `current` 심볼릭 링크를 바꾼 뒤 서비스를 재시작한다.

```bash
#!/usr/bin/env bash
# /opt/myapp/bin/deploy.sh
set -euo pipefail

COMMIT_SHA="$1"
APP_ROOT="/opt/myapp"
INCOMING="/var/tmp/myapp-incoming"
RELEASE_ROOT="${APP_ROOT}/releases"
CURRENT_LINK="${APP_ROOT}/current"

if ! [[ "$COMMIT_SHA" =~ ^[0-9a-f]{40}$ ]]; then
  echo "DEPLOY_FAIL invalid_commit_sha"
  exit 30
fi

MANIFEST="${INCOMING}/deploy-manifest.json"
ARTIFACT_FILE="$(jq -r '.artifact.file' "$MANIFEST")"
EXPECTED_SHA="$(jq -r '.artifact.sha256' "$MANIFEST")"

if ! [[ "$ARTIFACT_FILE" =~ ^[A-Za-z0-9._-]+\.jar$ ]]; then
  echo "DEPLOY_FAIL invalid_artifact_name"
  exit 31
fi

ACTUAL_SHA="$(sha256sum "${INCOMING}/${ARTIFACT_FILE}" | awk '{print $1}')"
if [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
  echo "DEPLOY_FAIL checksum_mismatch"
  exit 32
fi

RELEASE_DIR="${RELEASE_ROOT}/${COMMIT_SHA}"
mkdir -p "$RELEASE_DIR"
cp "${INCOMING}/${ARTIFACT_FILE}" "${RELEASE_DIR}/app.jar"
cp "$MANIFEST" "${RELEASE_DIR}/deploy-manifest.json"

ln -sfn "$RELEASE_DIR" "$CURRENT_LINK"

sudo -n /bin/systemctl restart myapp.service

echo "DEPLOY_OK commit=${COMMIT_SHA}"
```

심볼릭 링크를 쓰는 이유는 rollback을 단순하게 만들기 위해서다. `current`가 현재 실행 중인 release를 가리키고 있으면, rollback은 이전 release 디렉터리로 링크를 되돌리는 작업이 된다. 대용량 JAR을 매번 덮어쓰기보다 release directory를 보존하는 방식이 감사와 복구에도 유리하다.

systemd service는 `current` 링크를 바라보게 구성한다.

```ini
# /etc/systemd/system/myapp.service
[Unit]
Description=MyApp Spring Boot Service
After=network.target

[Service]
User=svc-myapp
Group=svc-myapp
WorkingDirectory=/opt/myapp/current
ExecStart=/usr/bin/java -jar /opt/myapp/current/app.jar
Restart=on-failure
RestartSec=5
Environment=SPRING_PROFILES_ACTIVE=internal

[Install]
WantedBy=multi-user.target
```

### 6.9 sudoers 예시

Jenkins가 내부 서버에서 범용 sudo를 얻으면 안 된다. 아래 예시는 `deploy-gw`라는 SSH 접속 계정이 `svc-myapp` 권한으로 정해진 배포 스크립트만 실행하도록 제한한다. 그리고 `svc-myapp`은 `myapp.service` restart만 할 수 있게 한다.

```sudoers
# /etc/sudoers.d/myapp-deploy
Defaults:deploy-gw !requiretty
Defaults:svc-myapp !requiretty

deploy-gw ALL=(svc-myapp) NOPASSWD: /opt/myapp/bin/deploy.sh *
deploy-gw ALL=(svc-myapp) NOPASSWD: /opt/myapp/bin/rollback.sh *

svc-myapp ALL=(root) NOPASSWD: /bin/systemctl restart myapp.service
svc-myapp ALL=(root) NOPASSWD: /bin/systemctl status myapp.service
```

이 설정은 PoC용으로 이해하기 쉽게 쓴 예시다. 운영에서는 wildcard 인자 제한, wrapper script, command digest 고정, SSH forced command, auditd 기록까지 추가하는 편이 안전하다. 중요한 점은 Jenkins가 `sudo su -`로 열린 shell을 얻는 것이 아니라, 서비스 계정 권한으로 특정 script만 실행해야 한다는 것이다.

### 6.10 rollback 예시

rollback은 Jenkins가 내부망 상세 로그를 읽는 방식이 아니라, 내부망 서버에 보존된 이전 release로 링크를 되돌리는 방식으로 처리한다.

```bash
#!/usr/bin/env bash
# /opt/myapp/bin/rollback.sh
set -euo pipefail

TARGET_COMMIT_SHA="$1"
APP_ROOT="/opt/myapp"
TARGET_RELEASE="${APP_ROOT}/releases/${TARGET_COMMIT_SHA}"
CURRENT_LINK="${APP_ROOT}/current"

if ! [[ "$TARGET_COMMIT_SHA" =~ ^[0-9a-f]{40}$ ]]; then
  echo "ROLLBACK_FAIL invalid_commit_sha"
  exit 40
fi

if [ ! -f "${TARGET_RELEASE}/app.jar" ]; then
  echo "ROLLBACK_FAIL release_not_found"
  exit 41
fi

ln -sfn "$TARGET_RELEASE" "$CURRENT_LINK"
sudo -n /bin/systemctl restart myapp.service

echo "ROLLBACK_OK commit=${TARGET_COMMIT_SHA}"
```

Jenkins 화면에는 `ROLLBACK_OK`와 대상 commit 정도만 남긴다. 실제 exception log는 내부망 로그 저장소에서 확인한다.

## 7. Jenkins가 배포 후 무엇을 모니터링해야 하는가

Jenkins는 DMZ에 있으므로 내부망의 모든 것을 보는 관측 시스템이 되면 안 된다. 이 구조에서 Jenkins가 봐도 되는 것은 배포 제어에 필요한 최소 상태다.

허용 가능한 예시는 다음과 같다.

```json
{
  "deployment_id": "deploy-20260422-001",
  "project": "group/myapp",
  "commit_sha": "1f2e3d4c5b6a7890abcdef1234567890abcdef12",
  "target_group": "myapp-test",
  "result": "success",
  "nodes": [
    {"node_alias": "node-a", "status": "ok"},
    {"node_alias": "node-b", "status": "ok"}
  ],
  "started_at": "2026-04-22T09:40:00Z",
  "finished_at": "2026-04-22T09:44:10Z"
}
```

주의할 점은 `node_alias`다. 실제 hostname, IP, private DNS, DB endpoint를 Jenkins에 그대로 노출할지 여부는 보안팀과 합의해야 한다. 처음 PoC에서는 `node-a`, `node-b`, `myapp-test`처럼 coarse alias만 보여주는 편이 안전하다.

Jenkins가 보면 안 되는 예시는 다음과 같다.

```text
java.sql.SQLException: Access denied for user 'prod_user'@'10.10.12.34'
jdbc:mysql://prod-db-01.internal.company.local:3306/customer
Authorization: Bearer eyJ...
server inventory: app-prod-a=10.10.1.11, app-prod-b=10.10.1.12
```

개발자가 장애 원인을 봐야 한다면, 내부망 로그 시스템에서 권한을 받아 확인하는 흐름을 따로 둔다. Jenkins는 배포 orchestration 화면이지 내부망 로그 포털이 아니다.

## 8. 장애와 복구는 어떻게 설계하는가

### 8.1 Runner 부하 증가

증상은 pipeline 대기 시간이 길어지는 것이다. 해결은 Runner 수를 늘리거나, job tag를 나누거나, 통합 테스트를 병렬화하는 것이다. 중요한 점은 Runner를 늘릴 때도 내부망 secret을 Runner host에 올리지 않는 것이다. Runner는 연구/개발망의 build/test 역할까지만 맡는다.

### 8.2 Nexus 장애

Nexus가 죽으면 Maven dependency를 받을 수 없어 build가 실패한다. 그래서 구조도에는 Nexus Repository HA가 있다. Sonatype 공식 문서는 Nexus Repository HA가 여러 node를 load balancer 뒤에 두고, 외부 PostgreSQL과 shared blob storage를 사용한다고 설명한다. 단, HA 기능과 요구사항은 제품 edition과 배포 방식에 따라 다르므로 PoC 전에 라이선스와 지원 범위를 확인해야 한다.

### 8.3 DMZ Receiver 장애

Receiver 1이 죽어도 L4가 Receiver 2로 upload를 보내면 된다. 단, Receiver local disk에만 파일을 저장하면 장애 시 파일이 사라진다. 그래서 검증 통과 파일은 Shared Drop Zone에 저장해야 한다.

### 8.4 webhook 또는 upload 중복 수신

같은 commit과 pipeline에서 같은 bundle이 두 번 올라올 수 있다. Receiver는 `project + commit_sha + artifact.sha256`을 idempotency key로 보고, 이미 verified에 같은 후보가 있으면 중복 배포 후보를 만들지 않는다.

예시는 다음과 같다.

```bash
IDEMPOTENCY_KEY="$(jq -r '[.project, .commit_sha, .artifact.sha256] | join(":")' deploy-manifest.json)"
if grep -Fx "$IDEMPOTENCY_KEY" /mnt/idc-shared-drop-zone/index/accepted.keys >/dev/null; then
  echo "ACCEPT already_verified"
  exit 0
fi
```

### 8.5 일부 서버만 배포 실패

Rolling deploy 중 Server A는 성공하고 Server B가 실패할 수 있다. 이때 자동으로 모든 서버를 다시 건드리기보다 다음 규칙을 둔다.

1. 실패한 서버는 L4에 join하지 않는다.
2. 이미 성공한 서버는 그대로 서비스한다.
3. Jenkins는 `partial_failed` 상태를 표시한다.
4. 상세 원인은 내부망 로그에서 확인한다.
5. 재시도는 실패한 서버만 대상으로 수행한다.
6. 전체 rollback이 필요하면 이전 commit으로 `rollback.sh`를 호출한다.

### 8.6 Jenkins Controller 장애

Jenkins Controller가 죽으면 진행 중인 배포 상태가 끊길 수 있다. 그래서 운영에서는 Jenkins job history, credentials, plugin, Jenkins home, backup, failover 절차를 따로 설계해야 한다. Jenkins를 Active/Standby로 그렸다고 해서 자동으로 무중단 Controller HA가 완성되는 것은 아니다. MVP에서는 “Controller 장애 시 배포 중지, Drop Zone 후보 보존, 복구 후 재실행”을 기본 정책으로 두는 편이 안전하다.

### 8.7 내부망 상세 장애 분석

배포 후 애플리케이션이 500을 반환하거나 DB 연결에 실패할 수 있다. 이때 Jenkins는 `health_status=500` 정도만 기록하고, 내부망 로그 저장소에서 상세 원인을 본다. 내부망의 stack trace를 GitLab issue, Claude prompt, Jenkins console에 그대로 붙여 넣는 것은 금지한다.

## 9. 운영팀과 합의해야 할 경계

PoC 전에 아래 항목을 인프라팀, 보안팀, 개발팀이 함께 결정해야 한다.

| 항목 | 합의해야 할 내용 |
| --- | --- |
| 방화벽 | 연구/개발망 Runner -> DMZ Receiver HTTPS, DMZ Jenkins Agent -> 내부 target SSH/HTTPS, Jenkins -> L4 API 허용 여부 |
| DNS | 내부 hostname이 DMZ와 Jenkins 로그에 노출되어도 되는지, alias를 쓸지 |
| TLS/mTLS | GitLab/Runner -> DMZ Receiver 구간에 서버 인증서만 쓸지, client certificate까지 쓸지 |
| GitLab Runner | Runner host 수, tag 정책, executor, 동시 실행 수, R&D MySQL 접근 범위 |
| Nexus | hosted/proxy/group repository 이름, HA edition, 외부 DB, blob storage, backup |
| Drop Zone | incoming, rejected, verified 디렉터리 권한, 보존 기간, immutable 정책 |
| Jenkins | Controller/Agent 배치, credential scope, folder 권한, job 승인자, backup/failover |
| 내부 서버 계정 | SSH 접속 계정, 서비스 계정, sudoers allowlist, forced command, auditd |
| Health Check | Jenkins가 읽어도 되는 endpoint, status code 외 body 노출 여부 |
| 로그 | 내부망 상세 로그 저장 위치, 개발자 조회 절차, 외부 반출 금지 기준 |
| 감사 | 누가, 언제, 어떤 commit/artifact를, 어떤 target group에 배포했는지 보존하는 위치 |

여기서 가장 먼저 닫아야 하는 질문은 “내부망에서 DMZ Jenkins로 돌아오는 최소 상태 응답을 허용할 것인가”다. HTML은 Jenkins가 최소 상태를 본다고 가정한다. 보안 정책상 이것도 외부 유출로 본다면 Jenkins 화면에는 더 적은 정보만 표시하고, 실제 상태 확인은 내부망 포털로 분리해야 한다.

## 10. CTO와 인프라팀에게 설명할 때의 요약 문장

CTO에게는 비용과 위험 감소 중심으로 설명한다.

> 이 구조는 처음부터 완전한 플랫폼을 만들자는 제안이 아닙니다. 몇 주 안에 PoC 가능한 범위로 시작하되, 나중에 프로젝트와 서버가 늘어도 보안 경계를 다시 갈아엎지 않게 만드는 MVP입니다. GitLab Runner는 빌드 부하를 분산하고, Nexus는 사내 라이브러리와 외부 dependency 경로를 통제하고, DMZ Jenkins는 배포 타이밍과 상태를 관리합니다. 내부망의 민감 로그와 secret은 개발망과 Claude로 되돌아가지 않습니다.

인프라팀에게는 서버 배치와 권한 경계 중심으로 설명한다.

> 연구/개발망에는 GitLab HA, Runner Pool, R&D MySQL, Nexus HA를 둡니다. DMZ에는 L4 뒤 Receiver 두 대, Shared Drop Zone, Jenkins Controller/Agent를 둡니다. 내부망에는 L4 뒤 target server group과 내부 로그 저장소를 둡니다. Runner는 내부망에 직접 배포하지 않고, Jenkins는 내부망에서 allowlist된 서비스 계정 스크립트만 호출합니다.

개발자에게는 사용 흐름 중심으로 설명한다.

> 개발자는 평소처럼 Claude로 개발을 보조받고 GitLab에 MR을 올립니다. merge되면 Runner가 테스트와 빌드를 수행합니다. 성공하면 배포 후보가 DMZ에 생기고, 배포 담당자가 Jenkins에서 원하는 시점에 배포합니다. 배포 결과는 Jenkins에서 요약 상태로 보고, 상세 장애 원인은 내부망 로그 시스템에서 확인합니다.

보안팀에게는 금지 흐름 중심으로 설명한다.

> 내부망 상세 로그, 운영 secret, 서버 inventory, DB endpoint, stack trace는 연구/개발망 GitLab이나 Claude로 반출하지 않습니다. Jenkins는 DMZ에 있지만 내부망 전체 관리자 권한을 갖지 않습니다. artifact는 manifest와 checksum으로 검증하고, credential은 최소 scope와 회전 정책을 둡니다.

## 11. 확인한 사실과 이 문서의 추론

이 절은 발표 중 “이건 공식 문서에 근거한 이야기인가, 우리 설계 판단인가”를 구분하기 위해 둔다.

### 11.1 공식 문서로 확인한 사실

- GitLab 애플리케이션은 사용자, 권한, issue 같은 영속 메타데이터를 PostgreSQL에 저장하고, Git object 접근은 Gitaly를 통해 처리한다. 따라서 GitLab DB와 R&D MySQL은 역할이 다르다.
- Gitaly Cluster(Praefect)는 Git repository storage의 장애 허용성과 확장을 위해 여러 Gitaly node와 Praefect를 사용하는 구조다. Praefect는 Gitaly 요청을 라우팅하고 cluster metadata를 관리한다.
- GitLab Runner fleet은 동시 실행 수와 runner worker 구성을 통해 확장할 수 있다. self-managed runner는 CI job에 정의된 코드를 실행하므로 runner host와 network 보안이 중요하다.
- GitLab job artifacts는 build output과 report file을 보관할 수 있고, artifact access 제한을 설정할 수 있다.
- Jenkins는 Controller가 모든 실행을 직접 맡는 방식보다 Agent로 workload를 위임하는 distributed build 구조를 제공한다. Controller는 상태를 가지므로 백업과 복구 정책이 필요하다.
- Jenkins credential은 scope를 좁혀야 하며, credential masking은 script가 의도적으로 secret을 다른 방식으로 노출하는 것까지 막아주지 못한다.
- Nexus Repository HA는 여러 Nexus node를 load balancer 뒤에 두고 외부 DB와 shared blob storage를 쓰는 방식으로 설명된다. HA 기능은 제품 edition과 배포 방식에 따라 확인이 필요하다.
- NIST Zero Trust 문서는 네트워크 위치만으로 암묵적 신뢰를 주지 말고 resource, identity, workflow 중심으로 인증과 권한을 판단해야 한다고 설명한다.
- SLSA는 artifact와 provenance를 검증할 때 builder identity, signature, artifact digest와 provenance subject 일치, buildType과 externalParameters 기대값 확인을 요구한다.

### 11.2 이 문서의 설계 추론

- 연구/개발망에 GitLab을 두는 것 자체보다 위험한 것은 내부망 배포 로그와 상태가 연구/개발망 GitLab 또는 Claude로 되돌아가는 것이다.
- GitLab Runner는 build/test에 집중시키고, 내부망 배포 권한은 Jenkins와 내부 서비스 계정 allowlist로 분리하는 편이 MVP에서 설명과 운영이 쉽다.
- DMZ Receiver는 단순 파일 수신 서버가 아니라 검증 전 파일과 검증 완료 파일을 나누는 quarantine 지점이어야 한다.
- Jenkins는 DMZ에 둘 수 있지만, 내부망 상세 로그 수집기나 내부망 root 명령 실행기가 되면 안 된다.
- 초기 PoC에서는 자동 배포보다 Jenkins 수동 승인 배포가 더 안전하다. merge와 production 반영 사이에 사람이 배포 시점을 잡을 수 있기 때문이다.
- 내부망에서 DMZ Jenkins로 돌아오는 상태 정보도 데이터 흐름이다. 그래서 health check status와 요약 결과만 허용하고 raw log는 내부망에 남기는 정책이 필요하다.

## 12. 참고 문서

- [GitLab architecture overview](https://docs.gitlab.com/development/architecture/)
- [GitLab Gitaly Cluster (Praefect)](https://docs.gitlab.com/administration/gitaly/praefect/)
- [GitLab reference architectures](https://docs.gitlab.com/administration/reference_architectures/)
- [GitLab Runner fleet scaling](https://docs.gitlab.com/runner/fleet_scaling/)
- [GitLab Runner security](https://docs.gitlab.com/runner/security/)
- [GitLab job artifacts](https://docs.gitlab.com/ci/jobs/job_artifacts/)
- [GitLab token overview](https://docs.gitlab.com/security/tokens/)
- [GitLab webhooks](https://docs.gitlab.com/user/project/integrations/webhooks/)
- [Jenkins Architecting for Scale](https://www.jenkins.io/doc/book/scaling/architecting-for-scale/)
- [Jenkins Credentials](https://www.jenkins.io/doc/book/security/credentials/)
- [Jenkins Credentials Masking Limitations](https://www.jenkins.io/blog/2019/02/21/credentials-masking/)
- [Sonatype Nexus Repository High Availability Deployment](https://help.sonatype.com/en/high-availability-deployment.html)
- [Sonatype Nexus Maven Repositories](https://help.sonatype.com/en/maven-repositories.html)
- [NIST SP 800-207 Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [SLSA Verifying Artifacts](https://slsa.dev/spec/v1.2/verifying-artifacts)
