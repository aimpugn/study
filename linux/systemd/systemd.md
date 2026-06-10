# systemd — PID 1은 어떻게 서비스의 생애 주기 전체를 관리하는가

이 문서의 출발 질문은 네 개가 연쇄된 것입니다.

1. **systemd는 어떤 문제를 풀려고 등장했고, SysV init과 근본적으로 무엇이 다른가?**
2. **유닛 파일 문법(`Type=`, `Restart=`, 리소스·보안 지시어)은 각각 어떤 운영 요구에서 나왔나?**
3. **일반(서비스) 계정으로 데몬을 돌리려면 어떻게 하나 — 그리고 왜 `su`로 들어간 계정에서는 `systemctl`이 안 되나?**
4. **컨테이너에서 systemd를 PID 1로 쓸 때 왜 `SIGTERM`이 아니라 `SIGRTMIN+3`을 보내야 하나?**

네 질문을 관통하는 한 가지 사실이 있습니다. **systemd는 "부팅 스크립트를 순서대로 돌리는 도구"가 아니라, 서비스가 시스템 전체와 맺는 관계(의존·순서·자원·수명·권한)를 선언으로 받아 커널 프리미티브(cgroup, 소켓, 신호, 네임스페이스)에 직결시키는 런타임**이라는 점입니다. su 문제도, SIGRTMIN+3도 이 모델의 귀결입니다.

> 이 문서의 관측값은 WSL2 위 Ubuntu 26.04 LTS(systemd 259.5, 커널 6.6.114.1-microsoft-standard-WSL2, PID 1 = systemd, `systemctl is-system-running` = `running`)에서 2026-06-10에 직접 실행해 얻은 것입니다. 정리 전 원자료는 [systemd_raw.md](systemd_raw.md)에 보존되어 있습니다.

## 목차

- [한 줄 답](#한-줄-답)
- [왜 생겼나 — SysV init의 한계와 2010년의 재설계](#왜-생겼나--sysv-init의-한계와-2010년의-재설계)
- [핵심 모델 — 유닛 · 의존성 그래프 · 잡 · cgroup](#핵심-모델--유닛--의존성-그래프--잡--cgroup)
- [유닛 파일 문법 — 지시어가 답하는 운영 질문](#유닛-파일-문법--지시어가-답하는-운영-질문)
- [실전 패턴 — 최소에서 변형까지](#실전-패턴--최소에서-변형까지)
- [사용자 인스턴스 — su로 들어간 계정은 왜 systemctl이 안 되나](#사용자-인스턴스--su로-들어간-계정은-왜-systemctl이-안-되나)
- [컨테이너의 PID 1 — 왜 SIGRTMIN+3인가](#컨테이너의-pid-1--왜-sigrtmin3인가)
- [직접 확인해 보기](#직접-확인해-보기)
- [출처](#출처)
- [저장소 안 관련 문서](#저장소-안-관련-문서)

## 한 줄 답

1. systemd는 SysV init의 순차 부팅·상태 추적 부재를 풀기 위해, **모든 관리 대상을 유닛(unit)이라는 단일 추상으로 통일하고 의존성 그래프를 위상 정렬해 병렬 실행하며, cgroup으로 각 유닛의 모든 프로세스를 추적하는 PID 1**로 2010년에 등장했습니다.
2. 유닛 문법의 각 지시어는 운영 질문 하나씩에 대응합니다 — `Type=`은 "준비 완료를 누가 알리는가", `Restart=`는 "죽으면 어떻게 하는가", `MemoryMax=`는 "자원을 얼마나 쓰게 둘 것인가".
3. 일반 계정의 데몬은 **user 유닛 + `loginctl enable-linger`** 가 정석입니다. `su`로 들어간 계정에서 `systemctl`이 안 되는 이유는 권한이 아니라 **logind 세션 부재**입니다 — su는 호출자가 이미 세션 안이면 새 세션을 만들지 않고, 세션이 없으면 user 매니저도 `/run/user/<uid>`도 없습니다.
4. PID 1인 systemd에게 `SIGTERM`은 종료가 아니라 **자기 재실행(reexec)** 신호입니다. 질서 있는 전체 종료를 트리거하는 신호는 `SIGRTMIN+3`(= `halt.target` 시작)으로 따로 예약되어 있습니다.

## 왜 생겼나 — SysV init의 한계와 2010년의 재설계

### SysV init이 안고 있던 네 가지 문제

1980년대 AT&T System V에서 온 init은 "런레벨 디렉토리의 스크립트를 파일명 순서(`S20network` → `S80nginx`)로 실행"하는 모델이었습니다. 1990년대 중반부터 리눅스 배포판의 표준이었지만, 서버 환경이 커지면서 네 가지 구조적 한계가 드러났습니다.

1. **순차 부팅** — 스크립트가 하나씩 실행되어, SSD·멀티코어 시대에도 부팅이 수십 초씩 걸렸습니다.
2. **상태 추적 부재** — `/etc/init.d/foo start`가 성공했는지, 프로세스가 언제 죽었는지 init 자신이 알 방법이 없었습니다. 데몬이 double-fork로 스스로 백그라운드화하면 부모-자식 관계가 끊겨 추적이 불가능해집니다.
3. **동적 하드웨어 대응 불가** — USB 랜카드처럼 "나중에 등장"하는 디바이스에 맞춰 서비스 기동 시점을 바꿀 수단이 스크립트 체계에는 없었습니다.
4. **커널 신기능과의 단절** — cgroup·네임스페이스 같은 자원 제어 프리미티브를 init 수준에서 통합하지 못했습니다.

### 설계자가 버린 가정들

2010년 4월 30일 Lennart Poettering이 "Rethinking PID 1"을 발표하며 (Kay Sievers와 함께 개발한) systemd를 공개했습니다. 이 설계는 기존 init의 가정 세 개를 명시적으로 버렸습니다.

- **"부팅은 스크립트의 순차 실행이다"** → 부팅은 의존성 그래프(DAG)의 위상 정렬이며, 의존이 풀린 노드는 동시에 실행합니다. 직접 영감은 2005년 Apple launchd의 소켓 활성화입니다 — 소켓을 먼저 열어 두면 "A가 B보다 먼저 떠야 한다"는 순서 제약의 상당수가 사라지고, 부팅이 병렬화됩니다.
- **"데몬은 double-fork로 스스로를 관리한다"** → 데몬화는 매니저의 일입니다. 프로세스는 포그라운드로 실행되고, PID 1이 cgroup으로 그 프로세스와 모든 자손을 단일 객체로 추적합니다. 이 결정 덕분에 `Restart=`, `MemoryMax=` 같은 선언형 정책이 가능해졌습니다.
- **"started를 출력하면 준비된 것이다"** → 준비 완료는 프로토콜(`sd_notify(READY=1)`, 소켓 활성화)로 알립니다.

당시 대안으로 이미 Canonical의 Upstart(2006, Scott James Remnant)가 있었습니다. Upstart는 "이벤트 기반"(X가 일어나면 Y를 시작)이었는데, systemd는 이를 뒤집어 "목표 상태 기반"(Y가 필요하니 의존부터 거꾸로 해결)을 택했습니다. 예측된 부작용 — "PID 1이 너무 많은 일을 한다"는 비대화 비판 — 은 실제로 이후 10년간 init 전쟁의 핵심 논쟁이 됐습니다.

### 타임라인

| 연도 | 사건 | 의미 |
|---|---|---|
| 1983 | AT&T UNIX System V 출시 — 런레벨 기반 SysV init | 이후 20여 년 리눅스 배포판의 표준 |
| 2005 | Apple launchd (Mac OS X 10.4 Tiger) | 소켓 활성화로 부팅 병렬화 — systemd의 직접 영감 |
| 2006 | Canonical Upstart 공개, Ubuntu 6.10 채택 | 이벤트 기반 init 시도 |
| 2010.04 | Poettering "Rethinking PID 1" 발표 | 의존성 그래프 + cgroup 상태 추적 모델 제시 |
| 2011.05 | Fedora 15 — 첫 메이저 배포판 기본 채택 | |
| 2012.10 | Arch Linux 기본 전환 | |
| 2014.02 | Debian 기술위원회 표결(캐스팅보트)로 채택 결정, Ubuntu도 Upstart 포기 선언 | "init 전쟁"의 정점 |
| 2014.06 | RHEL 7 채택 | 엔터프라이즈 표준화 |
| 2014.11 | Devuan 포크 선언 | 반발 진영의 분리 |
| 2015.04 | Debian 8, Ubuntu 15.04 기본 탑재 | 주요 배포판 전환 완료 |
| 2015 | machinectl shell 추가(v225) | su의 세션 한계에 대한 공식 대안 |
| 2016.05 | systemd 230 — `KillUserProcesses=yes` 기본값 전환 시도 | nohup·tmux가 죽는다는 반발로 대부분의 배포판이 no 유지 |
| 2019 | Fedora 31 — cgroup v2 기본 전환 | 단일 계층 자원 모델로 이행 시작 |
| 2022.09 | WSL2 systemd 지원 | 이 문서의 관측 환경 |

## 핵심 모델 — 유닛 · 의존성 그래프 · 잡 · cgroup

### 모든 것이 유닛이다

systemd가 관리하는 대상은 전부 유닛이라는 같은 추상으로 표현되고, 같은 의존성 문법과 같은 도구(`systemctl`, `journalctl`)로 다뤄집니다.

| 유닛 타입 | 관리 대상 | 대표 용도 |
|---|---|---|
| `.service` | 프로세스(데몬·일회성 작업) | 거의 모든 서비스 |
| `.socket` | 소켓(TCP/UDP/UNIX/FIFO) | 소켓 활성화, lazy 기동 |
| `.timer` | 시간 이벤트 | cron 대체 |
| `.target` | 유닛 묶음(동기화 지점) | `multi-user.target` 같은 부팅 단계 |
| `.mount` / `.automount` | 마운트 포인트 | fstab 통합, 지연 마운트 |
| `.slice` | cgroup 계층의 가지 | 워크로드 그룹별 자원 한도 |
| `.scope` | 외부에서 시작된 기존 프로세스 묶음 | 로그인 세션, `systemd-run --scope` |
| `.path` / `.device` | 파일 경로 / udev 디바이스 | 경로·장치 이벤트로 서비스 기동 |

`.slice`와 `.scope`의 구분이 자주 헷갈립니다 — slice는 "자원 계약의 계층 구조"(영구 정의), scope는 "systemd가 fork하지 않은, 이미 떠 있는 프로세스들을 사후에 편입하는 그릇"(런타임 생성)입니다. 로그인 세션(`session-N.scope`)이 scope인 이유가 이것입니다. 세션의 프로세스들은 systemd가 아니라 sshd 등이 fork했기 때문입니다.

### 유닛 파일은 어디서 로드되나

시스템 매니저 기준으로 세 경로를 읽으며, 위가 아래를 덮어씁니다.

| 우선순위 | 경로 | 주체 |
|---|---|---|
| 1 (높음) | `/etc/systemd/system/` | 관리자 |
| 2 | `/run/systemd/system/` | 런타임(generator 등) |
| 3 (낮음) | `/usr/lib/systemd/system/` | 패키지 |

사용자 매니저는 `~/.config/systemd/user/` > `/etc/systemd/user/` > `/usr/lib/systemd/user/` 순으로 같은 구조를 반복합니다.

패키지가 제공한 유닛을 고치고 싶을 때 파일을 직접 수정하면 다음 업그레이드 때 충돌합니다. 대신 `foo.service.d/override.conf` 드롭인에 변경분만 넣습니다(상세는 [실전 패턴](#실전-패턴--최소에서-변형까지)의 변형 3). fstab처럼 유닛이 아닌 설정을 부팅 시점에 유닛으로 변환해 주는 generator 메커니즘은 [commands/systemd.md](../commands/systemd.md)에 따로 정리되어 있습니다.

수정 후에는 `systemctl daemon-reload`로 매니저가 유닛 파일을 다시 읽고 의존성 그래프를 재구성하게 합니다. 서비스 자체 설정 파일을 다시 읽게 하는 `systemctl reload`와는 다른 명령입니다 — 이 혼동쌍은 [commands/systemctl.md](../commands/systemctl.md)에 정리했습니다.

### 의존성과 순서는 다른 축이다 — Requires/Wants vs After/Before

유닛 사이의 관계 지시어에서 가장 흔한 오해는 "Requires를 걸면 순서도 보장된다"입니다. 두 축은 독립입니다.

| 지시어 | 축 | 의미 | 상대가 실패하면 |
|---|---|---|---|
| `Requires=` | 의존 | 함께 활성화되어야 함 (강한 의존) | 이 유닛의 잡도 취소됨 (dependency failed) |
| `Wants=` | 의존 | 함께 활성화를 시도함 (약한 의존) | 이 유닛은 그대로 진행 |
| `After=` | 순서 | 상대가 먼저 "준비 완료"된 뒤에 시작 | (의존 아님 — 상대를 띄우지 않음) |
| `Before=` | 순서 | 상대보다 먼저 시작 | 〃 |
| `Conflicts=` | 배제 | 동시에 활성화될 수 없음 | — |

`Requires=postgresql.service`만 적고 `After=`를 빼면, 두 유닛은 **동시에** 시작됩니다 — postgres가 떠야 한다는 보장은 있지만 먼저 떠 있다는 보장은 없습니다. 그래서 실무 유닛에는 거의 항상 `Requires=`(또는 `Wants=`)와 `After=`가 쌍으로 들어갑니다.

실패가 어떻게 전파되는지는 잡(job) 모델로 설명됩니다. 부팅 시 systemd는 `default.target`을 루트로 그래프를 만들고 위상 정렬 결과를 잡 큐에 넣습니다. 각 잡은 `queued → starting → running → done/failed`로 전이합니다. 예를 들어 `postgresql.service`가 `TimeoutStartSec=30`을 초과해 failed가 되면, 이를 `Requires=`로 의존하는 `myapp.service`의 잡은 실행되지 못한 채 "dependency failed"로 취소되고, `systemctl list-jobs`와 `systemctl status` 양쪽에서 그 인과가 그대로 읽힙니다. 부팅이 느릴 때 가장 늦게 준비된 경로를 역추적하는 도구가 `systemd-analyze critical-chain`입니다. 관측 환경에서의 실제 출력입니다.

```text
graphical.target @2.953s
└─multi-user.target @2.953s
  └─snapd.seeded.service @1.306s +1.190s   # @ = 활성화된 시각, + = 그 유닛이 걸린 시간
    └─basic.target @1.243s                 # 이 체인에서는 snapd.seeded가 1.19s로 최대 지연원
      └─sockets.target @1.242s
        └─snapd.socket @1.212s +27ms
          └─sysinit.target @1.167s
```

전형적인 진단 사례: NVMe로 바꿨는데도 부팅이 40초 걸린다면, 이 체인에 `network-online.target`(DHCP 대기)이 끼어 있고 많은 서비스가 그 뒤에 `After=`로 묶여 있는 경우가 많습니다.

### 상태 추적의 단위는 PID가 아니라 cgroup이다

systemd는 유닛마다 cgroup 디렉토리 하나를 만들고, 그 유닛이 낳는 모든 프로세스·스레드를 그 안에 가둡니다. 관측 환경에서 user 유닛으로 띄운 `sleep` 프로세스의 cgroup 실값입니다.

```text
$ cat /proc/<sleep의 PID>/cgroup
0::/user.slice/user-1001.slice/user@1001.service/app.slice/sleeper.service
#   └ uid 1001의 자원 가지  └ user 매니저      └ 유닛 이름이 곧 cgroup 경로
```

이 한 가지 설계에서 세 가지 운영 기능이 한꺼번에 나옵니다.

1. **죽음의 감지와 재시작** — 메인 PID가 사라지면 cgroup 상태와 유닛 상태의 불일치를 즉시 인지하고 `Restart=` 정책(on-failure는 "0이 아닌 exit, 신호 사망, core dump"를 모두 트리거로 해석)으로 재시작 잡을 만듭니다. `StartLimitIntervalSec=` / `StartLimitBurst=`가 무한 재시작 루프를 막습니다.
2. **신호의 전파** — `systemctl kill --kill-who=all foo.service`는 cgroup 안 모든 프로세스에게 신호를 보냅니다. double-fork로 숨은 손자 프로세스도 빠져나갈 수 없습니다.
3. **자원 한도** — `MemoryMax=500M`을 선언하면 커널이 cgroup 한도에서 메모리를 강제합니다. 한도 초과 시 커널의 cgroup OOM kill이 일어나고, systemd는 그 결과를 유닛 상태(`Result=oom-kill`)로 기록한 뒤 `Restart=` 정책을 적용합니다. (원자료에는 "커널 OOM killer 대신 systemd가 먼저 동작"이라고 적혀 있었는데, 죽이는 주체는 커널이고 systemd는 그 결과를 유닛 모델에 반영하는 쪽이 정확합니다.)

## 유닛 파일 문법 — 지시어가 답하는 운영 질문

### 세 섹션의 역할

| 섹션 | 답하는 질문 | 대표 지시어 |
|---|---|---|
| `[Unit]` | 그래프에서 이 노드는 누구와 어떻게 연결되나 | `Description=`, `Documentation=`, `After=`, `Requires=`, `Wants=`, `Conflicts=`, `ConditionPathExists=` |
| `[Service]` | 프로세스를 어떻게 띄우고, 죽으면 어떻게 하고, 무엇까지 허용하나 | `Type=`, `ExecStart=`, `ExecStartPre=`, `ExecReload=`, `Restart=`, `User=`, `Environment=`, `MemoryMax=`, `PrivateTmp=` |
| `[Install]` | `systemctl enable` 때 어느 타겟에 심볼릭 링크를 걸 것인가 | `WantedBy=`, `RequiredBy=`, `Alias=`, `Also=` |

`enable`의 실체는 심볼릭 링크입니다. `WantedBy=multi-user.target`인 유닛을 enable하면 `/etc/systemd/system/multi-user.target.wants/foo.service` 링크가 생기고, 부팅 시 그 타겟이 활성화될 때 함께 끌려 올라옵니다. 관측 환경에서 user 유닛을 enable했을 때의 실제 출력이 이 구조를 그대로 보여 줍니다.

```text
Created symlink '/home/sdtest/.config/systemd/user/default.target.wants/sleeper.service'
  → '/home/sdtest/.config/systemd/user/sleeper.service'
# user 매니저의 부팅 타겟은 multi-user.target이 아니라 default.target
```

### Type= — "준비 완료"를 누가 어떻게 알리는가

`Type=`은 단순한 분류가 아니라 **"After=로 이 유닛을 기다리는 다른 유닛을 언제 풀어 줄 것인가"의 판정 방법**입니다.

| Type | 준비 판정 시점 | 필요한 협력 | 어울리는 경우 |
|---|---|---|---|
| `simple` | `fork()` 직후 즉시 | 없음 | 준비 시점이 중요하지 않은 포그라운드 데몬 |
| `exec` (v240+) | `execve()` 성공 시 | 없음 | simple과 같되 바이너리 실행 실패를 잡고 싶을 때 |
| `forking` | 부모 프로세스가 종료할 때 | `PIDFile=` | double-fork하는 전통 데몬 |
| `oneshot` | 프로세스가 종료할 때 | (`RemainAfterExit=`) | 스크립트·일회성 작업, 타이머와 결합 |
| `notify` | `sd_notify(READY=1)` 수신 시 | 데몬의 코드 협력 | 준비 시점이 중요한 현대 데몬 |
| `dbus` | D-Bus 이름 획득 시 | `BusName=` | D-Bus 서비스 |
| `idle` | 다른 잡이 끝난 뒤 | 없음 | 콘솔 출력 섞임 방지 정도 |

`forking`과 `PIDFile=`이 필연적 쌍인 이유: 부모가 즉시 종료하는 모델에서는 PID 1이 "진짜 메인 프로세스"를 알 방법이 없으므로, 데몬이 써 주는 PID 파일을 읽어 cgroup 안에서 그 PID를 메인으로 지정해야 재시작 정책이 올바른 대상을 기준으로 동작합니다.

### 안전장치 지시어 — 재시작 루프 · 자원 · 샌드박스

```ini
[Service]
Restart=on-failure            # 비정상 종료(exit!=0, 신호, core dump)에만 재시작
RestartSec=2s                 # 재시작 전 대기 — crash 직후 즉시 루프 방지
StartLimitIntervalSec=30      # 이 창(30초) 안에서
StartLimitBurst=3             # 3회를 넘게 죽으면 더 이상 기동하지 않음 (hard fail)
MemoryMax=500M                # cgroup v2 메모리 상한 — 초과 시 커널 OOM kill + Result=oom-kill
TasksMax=512                  # fork 폭탄 방지 (프로세스+스레드 수 상한)
LimitNOFILE=65536             # FD 한도 (ulimit -n 대응)
PrivateTmp=true               # 이 유닛 전용 /tmp — 마운트 네임스페이스로 격리
ProtectSystem=full            # /usr, /boot, /etc 읽기 전용 마운트
NoNewPrivileges=true          # setuid 바이너리로도 권한 상승 불가
AmbientCapabilities=CAP_NET_BIND_SERVICE  # 루트 없이 1024 미만 포트 바인딩
                              # 주의: 시스템 유닛 전용 — user 매니저는 비특권이라 capability를 줄 수 없음
```

마지막 줄의 주의가 중요합니다. 원자료에는 `AmbientCapabilities=`가 user 유닛 예제에 들어 있었는데, 이는 동작하지 않습니다 — capability를 부여하는 주체가 특권을 가져야 하는데 user 매니저는 해당 사용자 권한으로 돕니다. user 유닛으로 80 포트를 직접 열 수 없는 이유이며, [사용자 인스턴스](#사용자-인스턴스--su로-들어간-계정은-왜-systemctl이-안-되나) 절의 선택 기준으로 이어집니다.

## 실전 패턴 — 최소에서 변형까지

같은 문법이 운영 시나리오마다 어떻게 조합되는지를, 가장 작은 예제에서 시작해 점층적으로 봅니다.

### 최소 — Spring Boot 서비스 한 장

"jar를 nohup으로 띄우고 로그아웃하면 죽는" 문제의 표준 해법입니다(Spring 공식 문서의 배포 방식).

```ini
# /etc/systemd/system/myapp.service
[Unit]
Description=myapp
After=network.target            # 네트워크 설정 이후 시작 (단, "온라인 보장"은 network-online.target)

[Service]
User=myapp                      # 전용 계정으로 — 최소 권한
Group=myapp
Environment="JAVA_HOME=/path/to/java/home"
ExecStart=${JAVA_HOME}/bin/java -jar /var/myapp/myapp.jar
                                # Environment=로 정의한 변수는 ExecStart에서 ${} 확장 가능
SuccessExitStatus=143           # JVM은 SIGTERM(15)을 받으면 128+15=143으로 종료
                                # 이 줄이 없으면 정상 stop이 매번 failed로 기록됨

[Install]
WantedBy=multi-user.target
```

`sudo systemctl enable --now myapp.service`로 등록과 기동을 한 번에 합니다. 이 한 장으로 부팅 자동 기동, crash 시 상태 기록, `journalctl -u myapp` 로그 수집이 전부 따라옵니다.

### 표준 — 소켓 활성화 + Type=notify

드물게 호출되는 도구를 평소에 메모리에 올려 두지 않거나, 재배포 중에도 클라이언트 연결을 유실하지 않으려는 요구의 해법입니다. systemd가 소켓을 먼저 열어 두고, 첫 연결이 오면 그제야 서비스를 fork해 **이미 열린 소켓 FD를 3번부터 넘겨줍니다**(0/1/2 = stdin/stdout/stderr 다음).

```go
// main.go — systemd가 넘겨준 FD 3을 리스너로 쓰는 Go 서버
l, err := net.FileListener(os.NewFile(3, "listener")) // FD 3 = SD_LISTEN_FDS_START
if err != nil { log.Fatal(err) }

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

daemon.SdNotify(false, "READY=1") // 이 시점에야 After=이 풀림 — 포트 바인딩·초기화 완료 보장
log.Fatal(http.Serve(l, mux))
```

```ini
# webapp.socket — 커널이 먼저 리슨 상태로 들고 있는 쪽
[Socket]
ListenStream=127.0.0.1:9000
Accept=no                       # 연결마다 새 프로세스가 아니라, 리스너 자체를 단일 데몬에 인계

[Install]
WantedBy=sockets.target
```

```ini
# webapp.service — 첫 SYN이 올 때만 fork되는 쪽
[Unit]
Requires=webapp.socket          # 소켓과 생사를 같이함
After=webapp.socket

[Service]
Type=notify                     # READY=1 수신 전에는 "준비"로 치지 않음
ExecStart=/usr/local/bin/webapp
Restart=always                  # panic으로 죽어도 소켓 FD는 systemd가 들고 있으므로
                                # 재기동 사이에 도착한 SYN은 커널 큐에 보존됨 — 무중단처럼 보임
```

서비스가 죽었다 살아나는 동안에도 소켓은 systemd 소유로 계속 열려 있다는 점이 고가용성의 핵심입니다. 두 번째 접속자는 실패를 느끼지 못합니다.

### 변형 1 — oneshot + timer: cron을 대체할 때

cron 대비 차별점은 두 가지입니다. 전원이 꺼져 정각을 놓쳐도 부팅 직후 보정 실행되는 `Persistent=true`(마지막 트리거 시각을 `/var/lib/systemd/timers/`의 stamp 파일에 기록해 두고 비교 — 관측 환경에서 `stamp-apt-daily.timer` 등 실재 확인. 원자료의 "journald 데이터베이스에 저장" 서술은 부정확), 그리고 여러 서버가 동시에 I/O를 몰아치는 것을 막는 `RandomizedDelaySec=`입니다.

```ini
# nightly-backup.timer
[Timer]
OnCalendar=*-*-* 02:00:00       # ISO 8601 유사 문법 — "매월 셋째 수요일 22:30"도 선언 가능
Persistent=true                 # 꺼져 있던 시간대의 미실행분을 부팅 직후 보정
RandomizedDelaySec=15m          # 서버 군집의 동시 실행 분산

[Install]
WantedBy=timers.target
```

```ini
# nightly-backup.service — 타이머가 당기는 쪽
[Service]
Type=oneshot
User=backup
IOSchedulingClass=idle          # 본 서비스 I/O에 양보
Nice=19
ExecStart=/usr/local/bin/backup.sh
```

스케줄 확인은 `systemctl list-timers`(다음 실행·마지막 실행이 한 화면에), 실행 이력은 `journalctl -u nightly-backup.service`입니다.

### 변형 2 — 템플릿 유닛: 같은 데몬을 N개 띄울 때

```ini
# sshd@.service — 파일명의 @가 템플릿 표식
[Service]
Type=notify
ExecStart=/usr/sbin/sshd -D -f /etc/ssh/sshd_config.d/%i.conf
# %i = 인스턴스 이름. enable sshd@2222.service 하면 %i가 2222로 치환
```

`systemctl enable --now sshd@2222.service sshd@2223.service`처럼 단일 파일로 다중 포트 SSH, DB 샤드, 프록시 워커 군을 관리합니다.

### 변형 3 — drop-in: 패키지 유닛을 안전하게 덮어쓸 때

```ini
# /etc/systemd/system/nginx.service.d/override.conf — 달라지는 행만 적는다
[Service]
MemoryMax=1G
ProtectSystem=strict
```

`systemctl edit nginx`가 이 구조를 만들어 주며, `daemon-reload` 시 모든 드롭인을 합성해 최종 유닛을 재계산합니다. 패키지 업그레이드가 원본 유닛을 갈아치워도 로컬 정책이 살아남습니다. 합성 결과 확인은 `systemctl cat nginx`.

### 변형 4 — slice·scope: 이미 떠 있는 프로세스까지 자원 계약에 넣기

```bash
# 웹 계층 전체에 자원 상한 — slice는 영구 계층 정의
systemctl set-property web.slice CPUQuota=60% MemoryMax=4G

# 한두 시간 돌릴 분석 스크립트를 scope로 격리 — 일회성 작업의 사후 편입
systemd-run --scope -p CPUQuota=30% -p IOWeight=10 bash analyze.sh
# 반환된 run-rXXXX.scope 이름으로 journalctl -u 추적 가능
```

서비스 유닛과 임시 작업이 같은 cgroup 프레임워크에 들어가므로, "분석 작업이 웹 서비스와 CPU를 어떻게 다퉜는가"가 같은 도구(`systemd-cgtop`, `systemd-cgls`)로 관측됩니다.

## 사용자 인스턴스 — su로 들어간 계정은 왜 systemctl이 안 되나

### 매니저는 둘이다

systemd 매니저는 시스템에 하나(PID 1)만 있는 것이 아닙니다. 사용자마다 전용 매니저 `user@<uid>.service`(프로세스로는 `systemd --user`)가 뜰 수 있고, `systemctl --user`는 PID 1이 아니라 **이 user 매니저에게** D-Bus(`$XDG_RUNTIME_DIR/bus`)로 말을 겁니다. user 유닛은 `~/.config/systemd/user/`에 두고, 루트 권한도 polkit 인가도 없이 그 계정 스스로 관리합니다 — 애초에 "권한 부여"가 필요 없는 영역입니다.

```text
PID 1 (시스템 매니저)
└─ user.slice
   └─ user-1001.slice                  # uid 1001의 자원 가지
      ├─ session-c5.scope             # 로그인 세션 (sshd 등이 fork한 프로세스들)
      └─ user@1001.service            # user 매니저 — systemctl --user의 실제 상대방
         └─ app.slice/sleeper.service # user 유닛은 세션이 아니라 여기 소속 (관측값)
```

이 그림에서 핵심은 user 유닛이 **세션 scope 바깥**, user 매니저 아래에 산다는 점입니다. 로그아웃으로 세션이 사라져도 user 유닛은 영향받지 않습니다 — 단, user 매니저 자체가 살아 있어야 하고, 그 수명이 바로 문제의 전부입니다.

### 로그인 한 번이 만드는 세 가지

SSH·콘솔 로그인의 PAM 스택에는 `pam_systemd`가 들어 있고, pam_systemd(8) man page가 명시하는 대로 세션을 systemd-logind에 등록합니다 ("pam_systemd registers user sessions with the systemd login manager"). 그 등록의 부수 효과로 세 가지가 생깁니다.

1. **`session-N.scope`** — 이 세션의 모든 프로세스를 담는 cgroup.
2. **`/run/user/<uid>`** — `XDG_RUNTIME_DIR`. man page 원문: "created or mounted as new 'tmpfs' file system with quota applied". user bus 소켓이 이 안에 있습니다. 사용자의 마지막 세션이 닫히면 삭제됩니다.
3. **`user@<uid>.service` 기동** — man page 원문: "An instance of the system service user@.service, which runs the systemd user manager instance, is started."

### su의 정확한 규칙 — 두 관측의 대비

흔한 요약은 "su는 세션을 안 만든다"인데, 관측 결과 더 정확한 규칙은 **"호출자가 이미 logind 세션 안이면 새 세션을 만들지 않는다"**입니다. su의 PAM 스택에도 pam_systemd는 들어 있습니다(Debian 계열은 common-session 경유). 같은 머신에서 호출자의 상태만 바꿔 두 번 관측했습니다.

**케이스 A — 호출자가 세션 밖일 때 (wsl.exe가 띄운 셸, cgroup `/init.scope`):**

```text
# su -s /bin/bash - sdtest 직후
XDG_RUNTIME_DIR=[/run/user/1001]                          # 생겼다
cgroup: 0::/user.slice/user-1001.slice/session-c5.scope   # 새 세션 scope가 생겼다
systemctl --user status                                   # user 매니저에 접속됨
```

**케이스 B — 호출자가 세션 안일 때 (root가 세션 `session-c3.scope` 안에서 su):**

```text
# su -s /bin/bash - sdtest2 직후
XDG_RUNTIME_DIR=[]                                        # 비어 있다
cgroup: 0::/user.slice/user-0.slice/session-c3.scope      # 여전히 "root의" 세션 안!
systemctl --user status
Failed to connect to user scope bus via local transport:
$DBUS_SESSION_BUS_ADDRESS and $XDG_RUNTIME_DIR not defined
(consider using --machine=<user>@.host --user to connect to bus of other user)
```

케이스 B가 서버 현장의 전형입니다 — SSH로 로그인한 관리자가 `sudo su - 서비스계정`을 하는 순간이 정확히 "세션 안에서의 su"이기 때문입니다. uid는 바뀌었지만 cgroup상으로는 **남(호출자)의 세션 scope 안**에 있고, 자기 세션·runtime dir·user 매니저는 어디에도 없습니다. root의 세션에 "붙는" 것도 아닙니다 — root는 보통 자기 세션이 없고, 케이스 B에서 보이는 `session-c3.scope`는 su를 호출한 쪽의 세션입니다. (구버전 systemd에서는 같은 상황의 에러가 `Failed to connect to bus: No such file or directory`로 표시됩니다.)

### 증상 세 가지

세션 부재 하나에서 서로 다른 증상 세 개가 갈라져 나옵니다.

1. **`systemctl --user` 실패** — user bus(`$XDG_RUNTIME_DIR/bus`)도 user 매니저도 없으므로 위 관측처럼 접속 자체가 안 됩니다.
2. **비root 계정의 시스템 `systemctl`·`service`도 거부** — 시스템 유닛 제어는 root가 아니면 polkit 인가를 거치는데, polkit은 활성 세션 여부를 판단 기준으로 씁니다. su로 갈아탄 계정은 자기 세션이 없으니 인가받을 길이 없습니다. 관측값(su된 비root가 `systemctl stop systemd-timesyncd` 시도): `Failed to stop systemd-timesyncd.service: Access denied as the requested operation requires interactive authentication. However, interactive authentication has not been enabled by the calling program.` (polkitd가 떠 있는 일반 서버에서는 `Interactive authentication required.`로 표시되는 경우가 많습니다.) `service` 명령은 systemd 시스템에서 systemctl로 넘기는 래퍼라 함께 막힙니다. root로 su 했다면 시스템 systemctl은 정상 동작합니다 — root는 polkit을 타지 않습니다.
3. **"백그라운드 상시 실행"이 보장되지 않음** — su 셸에서 `nohup foo &`로 띄운 프로세스는 cgroup상 호출자의 세션 scope 소속입니다. 그 세션이 끝나면 `logind.conf`의 `KillUserProcesses=`에 따라 함께 종료될 수 있고(관측 환경 기본값: `false` — systemd 230이 yes로 바꾸려다 반발로 대부분의 배포판이 no 유지), 죽지 않더라도 남의 세션에 귀속된 채 떠도는 프로세스가 됩니다. user 매니저 역시 마지막 세션이 닫히면 내려갑니다 — 관측 환경에서도 su 세션 종료 후 별도 호출로 확인하니 `user@1001.service`는 `inactive`, `/run/user/1001`은 삭제되어 있었습니다(다만 이 확인은 호출 사이 WSL VM 재시작 가능성을 배제하지 못하므로, "linger 없이는 상시 유지되지 않는다"의 관측으로 한정합니다). 이 갈래 — "`&`로 띄우면 왜 죽고, 살아남으면 어떤 상태가 되나" — 는 바로 아래 절에서 층별로 해부합니다.

### `&`·nohup으로 띄운 프로세스의 수명 — 세 층이 차례로 결정한다

"백그라운드로 띄워도 세션이 끊기면 죽는다"는 기억이 들쭉날쭉한 이유는, 실제로는 **죽이는 주체가 서로 다른 세 층**이 겹쳐 있기 때문입니다. 층마다 살해 조건도 방어 수단도 다릅니다.

**층 1 — 셸·터미널 층 (`SIGHUP`, systemd 이전부터 있던 규칙).** 터미널이 끊기면(네트워크 단절, 창 닫기) 커널은 세션 리더(셸)와 포그라운드 프로세스 그룹에 SIGHUP을 보냅니다. 백그라운드 잡은 커널의 직접 대상이 아니지만, **대화형 bash가 자기가 받은 SIGHUP을 모든 잡(실행 중·정지 중 포함)에 재전파**하므로 `&` 잡이 여기서 죽습니다. 반면 깨끗하게 `exit`로 나가는 경우 bash는 기본값에서 잡에 SIGHUP을 보내지 않습니다(`huponexit` 옵션이 꺼져 있을 때 — 관측: `shopt -p huponexit` → `shopt -u huponexit`). "어떨 땐 죽고 어떨 땐 살아남던" 경험은 이 차이입니다. `nohup`은 SIGHUP 무시 + I/O 재배선으로 이 층만 방어하고, `setsid`는 커널 세션에서 아예 이탈합니다 — 세션·CTTY·PGID 수준의 상세는 [commands/nohup.md](../commands/nohup.md)에 있습니다.

**층 2 — logind 층 (cgroup 단위 살해).** `KillUserProcesses=yes`면 logind 세션이 끝날 때 그 세션 scope cgroup 안의 **모든** 프로세스에 SIGTERM→SIGKILL이 갑니다. 이 층에서는 nohup도 setsid도 tmux도 소용없습니다 — `setsid`가 바꾸는 것은 **커널 세션(SID)** 이고 logind가 살해 기준으로 삼는 것은 **cgroup(logind 세션)** 인데, setsid는 cgroup을 옮기지 않기 때문입니다. 같은 "세션"이라는 단어가 두 층에서 다른 것을 가리키는 혼동쌍입니다. systemd 230(2016)이 이 값을 yes 기본으로 바꾸려다 "tmux가 죽는다"는 반발을 부른 것이 정확히 이 메커니즘입니다. 관측 환경 기본값은 `false`이므로, 아래 표의 yes 칸은 직접 관측이 아니라 logind.conf(5)의 동작 정의에 따른 것입니다.

**층 3 — systemd 관리 층 (살아남아도 "관리 밖").** 층 1·2를 통과해 살아남은 프로세스가 어떤 상태가 되는지를 관측했습니다. su 세션 안에서 `nohup sleep 300 &`을 띄우고 세션을 닫은 뒤의 실측값입니다.

```text
# 세션이 살아 있을 때의 cgroup
0::/user.slice/user-0.slice/session-c3.scope

# 세션 종료 후 — 프로세스는 생존 (KillUserProcesses=false)
  441 sleep
# cgroup은 그대로 — 이미 닫힌 세션의 scope에 잔류
0::/user.slice/user-0.slice/session-c3.scope

$ systemctl status session-c3.scope
● session-c3.scope - Session c3 of User root
     Active: active (abandoned) since Wed 2026-06-10 16:46:54 KST
#                   ^^^^^^^^^^^ systemd의 공식 어휘 — "버려진" scope
```

`abandoned`가 "systemd에 의해 관리가 안 된다"의 실체입니다. 이 프로세스는 유닛이 아니므로 `Restart=`도 `MemoryMax=`도 없고, `journalctl -u`로 로그를 모으지 못하며, `systemctl list-units`에 보이지 않고, 죽어도 아무도 다시 띄우지 않습니다. 회계상으로만 죽은 세션의 scope에 매달려 있을 뿐입니다.

세 층을 모두 통과하는 정공법은 하나 — **유닛으로 등록하는 것**입니다(시스템 유닛, 또는 user 유닛 + linger).

| 방식 | 터미널 hangup 생존 | 셸 clean exit 생존 | KillUserProcesses=yes 생존 | 재부팅 후 자동 기동 | crash 재시작·자원 한도·유닛 로그 |
|---|---|---|---|---|---|
| `cmd &` | ✗ (bash가 SIGHUP 재전파) | △ (huponexit off면 생존) | ✗ | ✗ | ✗ |
| `nohup cmd &` | ○ (SIGHUP 무시) | ○ | ✗ | ✗ | ✗ |
| `setsid cmd` | ○ (커널 세션 이탈) | ○ | ✗ (cgroup은 그대로) | ✗ | ✗ |
| `tmux` / `screen` | ○ | ○ | ✗ (230 논쟁의 그 사례) | ✗ | ✗ |
| user 유닛 + linger | ○ | ○ | ○ (세션 scope 밖 — 관측값 참조) | ○ | ○ |
| 시스템 유닛 | ○ | ○ | ○ | ○ | ○ |

### 신호 권한과 세션 권한은 다른 층이다

su 문제와 자주 섞이는 오해가 "su로 전환한 유저는 SIGRTMIN+3 같은 것도 못 보낸다"입니다. 두 메커니즘은 층이 다릅니다.

- **세션 권한 (logind/polkit/user bus)** — `systemctl --user`와 비root의 시스템 유닛 제어를 막는 것. 위 증상 1·2의 원인.
- **신호 권한 (커널 kill(2))** — 같은 uid이거나 `CAP_KILL`(즉 root)이면 보낼 수 있음. 세션과 무관.

따라서 `sudo su -`로 root가 됐다면 세션이 없어도 `kill -s SIGRTMIN+3 1`은 **가능**하고, 비root 서비스 계정이라면 su 여부와 무관하게 PID 1에 신호 자체를 못 보냅니다(EPERM). "su 유저는 못 한다"는 기억은 비root 기준으로 결과만 맞고, 막는 메커니즘은 systemctl 계열 = 세션 부재, SIGRTMIN+3 = 신호 권한으로 서로 다릅니다.

### 서비스 계정 셋업 레시피

서비스 계정(예: `prxsvc`)이 자기 데몬을 스스로 배포·관리하는 구성의 전 과정입니다. 모든 단계의 출력은 관측 환경에서 실제로 실행해 확인한 값입니다.

```bash
# 1) root에서 한 번 — 세션 없이도 user 매니저를 상시 기동
loginctl enable-linger prxsvc
ls /var/lib/systemd/linger          # 관측: prxsvc 파일 생성됨 (linger의 영속 표식)
systemctl is-active user@$(id -u prxsvc).service
                                    # 관측: active — enable-linger가 즉시 기동까지 함
ls /run/user/                       # 관측: <uid> 디렉토리 생성됨
```

```ini
# 2) ~prxsvc/.config/systemd/user/my-app.service
[Unit]
Description=my app

[Service]
ExecStart=/opt/my-app/run.sh
Restart=on-failure

[Install]
WantedBy=default.target             # user 매니저의 부팅 타겟 (multi-user.target의 user판)
```

```bash
# 3) prxsvc로 작업 — su 경유라면 세션이 없으므로 환경 변수만 손으로 보충
sudo su -s /bin/bash - prxsvc       # 서비스 계정이 nologin 셸이면 -s 필요
export XDG_RUNTIME_DIR=/run/user/$(id -u)
                                    # linger 덕분에 디렉토리·bus·매니저가 이미 존재
systemctl --user daemon-reload
systemctl --user enable --now my-app.service
systemctl --user is-active my-app.service   # 관측: active

# 4) root에서 su 없이 직접 조작할 수도 있음 (관측 환경 systemd 259에서 동작 확인)
systemctl --user -M prxsvc@ status my-app.service
# 구버전(예: RHEL 8의 239)에서는 이 -M user@ 형태가 지원되지 않을 수 있음 — 버전 확인 필요
```

이렇게 띄운 프로세스의 cgroup은 `/user.slice/user-<uid>.slice/user@<uid>.service/app.slice/my-app.service`(관측값) — 어떤 세션 scope에도 속하지 않으므로 로그아웃·세션 종료와 무관하게 살아남고, 부팅 시에도 linger가 user 매니저를 올리면서 함께 기동됩니다.

### user 유닛이냐, system 유닛 + User= 냐

| 축 | user 유닛 + linger | system 유닛 + `User=` |
|---|---|---|
| 유닛 파일 위치 | `~/.config/systemd/user/` (계정 소유) | `/etc/systemd/system/` (root 소유) |
| 일상 관리 주체 | 계정 스스로 (`systemctl --user`) | root, 또는 좁게 위임받은 계정 |
| root 개입 | 최초 `enable-linger` 한 번 | 유닛 변경 때마다 root 경유 |
| 부팅 기동 | linger → user 매니저 → `default.target` | PID 1 → `multi-user.target` |
| 시스템 유닛과 의존성 | 불가 — user 매니저는 분리된 그래프라 `After=network-online.target` 같은 시스템 타겟 참조 불가 | 자유롭게 가능 |
| 1024 미만 포트 / capability | 불가 (`AmbientCapabilities=` 무효) | 가능 |
| 운영 가시성 | 시스템 `systemctl list-units`에 안 보임 (user@ 안에 숨음) | 전사 표준 도구에서 보임 |
| 로그 접근 | `journalctl --user -u` — 계정이 uid 1000 미만 시스템 계정이면 `systemd-journal` 그룹 필요할 수 있음 | root/`systemd-journal` 그룹 |
| 권한 위임 메커니즘 | 불필요 (자기 영역) | sudoers 또는 polkit rule로 유닛 단위 허용 |
| 어울리는 경우 | 계정 단위 self-service 배포(CI가 ssh로 배포), 비특권 포트 | 부팅 순서·특권 포트·전사 운영 표준이 걸릴 때 |

system 유닛 쪽을 택하면서 계정에게 제어만 위임하는 두 가지 방식:

```bash
# sudoers — 명령 단위로 좁게
prxsvc ALL=(root) NOPASSWD: /usr/bin/systemctl start my-app.service, \
    /usr/bin/systemctl stop my-app.service, /usr/bin/systemctl restart my-app.service
```

```js
// polkit rule (/etc/polkit-1/rules.d/55-prxsvc-myapp.rules, JS rules 지원 배포판)
polkit.addRule(function(action, subject) {
    if (action.id == "org.freedesktop.systemd1.manage-units" &&
        action.lookup("unit") == "my-app.service" &&     // 유닛 단위로 한정
        subject.user == "prxsvc") {
        return polkit.Result.YES;
    }
});
```

### su의 대안들

su 자체가 "로그인"이 아니라 "uid 변경"이라는 한계 인식에서 나온 공식 대안들입니다.

- `machinectl shell prxsvc@` — 제대로 된 logind 세션을 만들어 전환 (systemd-container 패키지 필요; 관측 환경에는 미설치였음)
- `ssh prxsvc@localhost` — 진짜 로그인 세션
- `systemctl --user -M prxsvc@ ...` — root가 세션 생성 없이 대상 user 매니저에 직접 접속
- `systemd-run --uid=prxsvc ...` — 세션이 아니라 유닛으로 실행

## 컨테이너의 PID 1 — 왜 SIGRTMIN+3인가

### PID 1에는 init이 필요하다

컨테이너 안에서 일반 애플리케이션을 PID 1로 두면 두 가지 문제가 생깁니다(Red Hat 컨테이너 블로그의 정리).

1. **좀비 누적** — 고아가 된 프로세스는 PID 1의 자식으로 재입양되는데, PID 1이 `SIGCHLD`를 받고 `wait()`해 주지 않으면 좀비가 무한히 쌓입니다.
2. **syslog 유실** — `/dev/log`를 리스닝하는 프로세스가 없으면 syslog(3)로 보낸 로그가 사라집니다.

systemd를 PID 1로 넣으면 재입양·journald·서비스 그래프가 컨테이너 안에서도 호스트와 동일한 API로 동작합니다.

### systemd PID 1의 신호 의미론

PID 1은 커널이 특별 취급합니다 — 핸들러를 설치하지 않은 신호는 전달되지 않고, 같은 네임스페이스 안에서는 SIGKILL로도 죽일 수 없습니다. 그 위에서 systemd는 신호별 의미를 자체 정의합니다. systemd(1) man page의 SIGNALS 절 원문 기준입니다.

| 신호 | 시스템 매니저의 동작 (man 원문) | 비고 |
|---|---|---|
| `SIGTERM` | "serializes its state, reexecutes itself and deserializes the saved state again" | **종료가 아니라 reexec** — `systemctl daemon-reexec` 상당 |
| `SIGINT` | "start the ctrl-alt-del.target unit" | 콘솔 Ctrl-Alt-Del |
| `SIGRTMIN+3` | "Halts the machine, starts the halt.target unit" | **질서 있는 전체 종료** |
| `SIGRTMIN+4` | "starts the poweroff.target unit" | 전원 차단 |
| `SIGRTMIN+5` | "starts the reboot.target unit" | 재부팅 |
| (참고) user 매니저의 `SIGTERM` | "start the exit.target unit" | user 매니저는 SIGTERM으로 정상 종료함 |

여기서 `docker stop`의 함정이 그대로 설명됩니다. docker stop은 기본적으로 SIGTERM을 보내고 타임아웃 후 SIGKILL을 보냅니다. PID 1이 systemd라면 SIGTERM은 reexec(사실상 무시처럼 보임)이고, 이어지는 SIGKILL은 하위 서비스를 정리할 기회 없이 전체를 즉사시킵니다. `SIGRTMIN+3`을 받아야 systemd가 `halt.target`으로 모든 유닛을 의존성 역순으로 — `systemctl stop`과 동일한 경로로 — 내리며 데이터 손실을 피합니다.

신호 번호로 보내야 한다면 주의가 필요합니다. glibc가 실시간 신호 앞부분(32·33)을 스레딩에 예약하므로 `SIGRTMIN`은 보통 34이고, 따라서 `SIGRTMIN+3` = 37입니다(관측: `kill -l RTMIN+3` → `37`). 번호 37을 하드코딩하기보다 이름으로 지정하는 편이 안전합니다.

### Dockerfile 패턴

```dockerfile
FROM ubuntu:24.04

ENV container=docker
STOPSIGNAL SIGRTMIN+3             # docker stop이 SIGTERM 대신 이 신호를 보내게 함
RUN apt-get update && apt-get install -y systemd systemd-sysv

VOLUME [ "/sys/fs/cgroup" ]       # systemd가 cgroup 트리를 관리할 수 있어야 함
CMD [ "/lib/systemd/systemd" ]    # PID 1로 systemd 실행
```

호스트에서 `docker run --privileged -v /sys/fs/cgroup:/sys/fs/cgroup:rw ...`로 기동하면 컨테이너 내부에서 `systemctl`·`journalctl`·`loginctl`이 호스트와 동일하게 동작합니다.

## 직접 확인해 보기

각 주장에 대해 무엇을 실행하면 되는지와, 이 문서 작성 시의 실제 관측값입니다. 관측 환경: WSL2 Ubuntu 26.04, systemd 259.

| 확인할 주장 | 명령 | 관측값 (PASS 기준) |
|---|---|---|
| systemd가 PID 1로 동작 중 | `ps -p 1 -o comm=` | `systemd` |
| 부팅 임계 경로 | `systemd-analyze critical-chain` | `graphical.target @2.953s └─...` 트리 출력 |
| 로그인 세션·runtime dir | `loginctl list-sessions; echo $XDG_RUNTIME_DIR` | 세션 목록 + `/run/user/1000` |
| su(세션 안)는 세션을 안 만듦 | 로그인 셸에서 `sudo su - <계정>` 후 `echo [$XDG_RUNTIME_DIR]; cat /proc/self/cgroup` | `[]` + 호출자의 `session-cN.scope` 그대로 |
| 그 상태의 systemctl --user 실패 | 위 셸에서 `systemctl --user status` | `Failed to connect to user scope bus ... $XDG_RUNTIME_DIR not defined` (구버전: `No such file or directory`) |
| 비root의 시스템 유닛 제어 거부 | su된 비root에서 `systemctl stop <유닛>` | `Access denied as the requested operation requires interactive authentication...` |
| linger가 user 매니저를 상시 기동 | root: `loginctl enable-linger <계정>` 후 `systemctl is-active user@<uid>.service; ls /run/user/` | `active` + uid 디렉토리 존재 |
| user 유닛은 세션 밖에 산다 | `cat /proc/<유닛 프로세스 PID>/cgroup` | `/user.slice/user-<uid>.slice/user@<uid>.service/app.slice/<유닛>` |
| root의 타 사용자 user 매니저 접속 | `systemctl --user -M <계정>@ is-active <유닛>` | `active` (systemd 259에서 동작; 구버전은 미지원 가능) |
| KillUserProcesses 실값 | `busctl get-property org.freedesktop.login1 /org/freedesktop/login1 org.freedesktop.login1.Manager KillUserProcesses` | `b false` (배포판마다 다름) |
| SIGRTMIN+3의 번호 | `kill -l RTMIN+3` | `37` (glibc 기준 — 번호 대신 이름 사용 권장) |
| enable의 실체는 심볼릭 링크 | `systemctl --user enable <유닛>` 출력 | `Created symlink '...default.target.wants/...'` |
| 살아남은 nohup 프로세스는 "버려진" scope에 잔류 | 세션 안에서 `nohup sleep 300 &` 후 세션 종료, `systemctl status session-cN.scope` | `Active: active (abandoned)` + 프로세스 cgroup이 닫힌 세션 scope 그대로 |
| bash는 clean exit에 잡을 죽이지 않음 (기본값) | `shopt -p huponexit` | `shopt -u huponexit` (off) |

컨테이너의 SIGRTMIN+3 동작은 이 환경에서 검증하지 않았습니다(PID 1에 신호를 보내는 실험은 동작 중인 시스템을 내리므로 제외). systemd(1) man page 인용과 Red Hat·raby.sh 문서를 근거로 두며, 검증하려면 일회용 컨테이너에서 `docker kill -s SIGRTMIN+3 <컨테이너>` 후 내부 서비스가 의존성 역순으로 stop되는 journald 로그를 확인하면 됩니다.

## 출처

- systemd(1) — SIGNALS 절. 신호별 동작 원문 인용의 출처. <https://man7.org/linux/man-pages/man1/systemd.1.html> (freedesktop.org 원본은 봇 차단으로 man7 미러 사용)
- pam_systemd(8) — 세션 등록·`/run/user/$UID`·`user@.service` 기동 원문 인용의 출처. <https://man7.org/linux/man-pages/man8/pam_systemd.8.html>
- Lennart Poettering, "Rethinking PID 1" (2010-04-30) — 설계 배경과 launchd 영감. <http://0pointer.de/blog/projects/systemd.html>
- Red Hat Developers, "Running systemd in a non-privileged container" (2016) — 컨테이너 PID 1 문제와 SIGRTMIN+3. <https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container>
- raby.sh, "SIGTERM and PID 1: why does a container linger after receiving a SIGTERM?" — docker stop과 PID 1 신호 의미론. <https://raby.sh/sigterm-and-pid-1-why-does-a-container-linger-after-receiving-a-sigterm.html>
- Spring Boot 공식 문서, "Installing Spring Boot Applications" — systemd 서비스 배포 예제. <https://docs.spring.io/spring-boot/how-to/deployment/installing.html>
- Red Hat 문서, "cgroup의 systemd 계층 구조 개요" (RHEL 9). <https://docs.redhat.com/ko/documentation/red_hat_enterprise_linux/9/html/managing_monitoring_and_updating_the_kernel/con_overview-of-systemd-hierarchy-for-cgroups_assembly_using-systemd-to-manage-resources-used-by-applications>
- SUSE, "Managing systemd Services". <https://documentation.suse.com/smart/systems-management/html/systemd-management/index.html>
- linuxhandbook, "How to create a systemd service in Linux". <https://linuxhandbook.com/create-systemd-services/>
- 직접 관측 — 이 문서의 모든 "관측:" 표기 값은 WSL2 Ubuntu 26.04 / systemd 259.5에서 2026-06-10 실행한 결과.

## 저장소 안 관련 문서

- [systemd_raw.md](systemd_raw.md) — 이 문서로 정리되기 전의 원자료 모음 (영문 튜토리얼, 사례 서사, AI 답변 원문 포함)
- [commands/systemd.md](../commands/systemd.md) — systemd.generator와 `.d/` 드롭인 상세
- [commands/systemctl.md](../commands/systemctl.md) — `reload` vs `daemon-reload` 혼동쌍
- [commands/su.md](../commands/su.md) — `sudo su -`와 `-`(로그인 셸 시뮬레이션)의 의미
- [commands/nohup.md](../commands/nohup.md) — SIGHUP·세션·CTTY·PGID 메커니즘 상세 (이 문서 "`&`·nohup 수명" 절의 층 1에 해당)
- [commands/crontab.md](../commands/crontab.md) — timer 유닛이 대체하는 대상
