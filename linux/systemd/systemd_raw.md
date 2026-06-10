# systemd

- [systemd](#systemd)
    - [예제](#예제)
        - [Spring application](#spring-application)
- [Systemd Unit Syntax](#systemd-unit-syntax)
    - [Basic Structure of a Systemd Unit File](#basic-structure-of-a-systemd-unit-file)
    - [Example 1: Basic Service Unit](#example-1-basic-service-unit)
        - [Explanation](#explanation)
    - [Example 2: Web Server with Dependencies](#example-2-web-server-with-dependencies)
- [Security settings](#security-settings)
        - [Explanation](#explanation-1)
    - [Example 3: Timer Unit (Scheduled Task)](#example-3-timer-unit-scheduled-task)
- [my-backup.timer](#my-backuptimer)
- [my-backup.service](#my-backupservice)
        - [Explanation](#explanation-2)
    - [Example 4: Mount Unit](#example-4-mount-unit)
        - [Explanation](#explanation-3)
    - [Example 5: Socket-Activated Service](#example-5-socket-activated-service)
- [myapp.socket](#myappsocket)
- [myapp.service](#myappservice)
        - [Explanation](#explanation-4)
    - [Common Directives](#common-directives)
        - [\[Unit\] Section](#unit-section)
        - [\[Service\] Section](#service-section)
        - [\[Install\] Section](#install-section)
    - [Managing Systemd Units](#managing-systemd-units)
- [As the 'web' user](#as-the-web-user)
- [As another user (e.g., during deployment)](#as-another-user-eg-during-deployment)
- [As root](#as-root)
- [`systemd`가 컨테이너에서 PID 1 프로세스로 동작 중이라면, 컨테이너 종료시 올바른 종료 신호가 필요합니다. 그런데 `systemd`는 `SIGTERM`이나 `SIGKILL`로 제대로 종료되지 않을 수 있습니다](#systemd가-컨테이너에서-pid-1-프로세스로-동작-중이라면-컨테이너-종료시-올바른-종료-신호가-필요합니다-그런데-systemd는-sigterm이나-sigkill로-제대로-종료되지-않을-수-있습니다)
- [- Systemd does not exit on sigterm.  Systemd defines that shutdown signal as `SIGRTMIN+3`, docker upstream should send this signal when user does a docker stop](#--systemd-does-not-exit-on-sigterm--systemd-defines-that-shutdown-signal-as-sigrtmin3-docker-upstream-should-send-this-signal-when-user-does-a-docker-stop)
- [- `SIGKILL`(`kill -9`): `systemd`는 즉시 종료되지만, 모든 하위 서비스는 제대로 정리되지 않고 그대로 남아 있을 수 있습니다](#--sigkillkill--9-systemd는-즉시-종료되지만-모든-하위-서비스는-제대로-정리되지-않고-그대로-남아-있을-수-있습니다)
- [](#)
- [`SIGRTMIN+3`신호를 받으면 `systemd`는 내부적으로 `systemctl stop`과 유사한 과정을 거쳐 모든 데몬 및 서비스를 순차적으로 종료함으로써 데이터 손실이나 불안정성을 최소화할 수 있습니다](#sigrtmin3신호를-받으면-systemd는-내부적으로-systemctl-stop과-유사한-과정을-거쳐-모든-데몬-및-서비스를-순차적으로-종료함으로써-데이터-손실이나-불안정성을-최소화할-수-있습니다)
- [](#-1)
- [References](#references)
- [- https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container#](#--httpsdevelopersredhatcomblog20160913running-systemd-in-a-non-privileged-container)
- [- https://raby.sh/sigterm-and-pid-1-why-does-a-container-linger-after-receiving-a-sigterm.html](#--httpsrabyshsigterm-and-pid-1-why-does-a-container-linger-after-receiving-a-sigtermhtml)
- [컨테이너가 시작될 때 항상 실행되어야 하는 명령을 정의합니다](#컨테이너가-시작될-때-항상-실행되어야-하는-명령을-정의합니다)
- [systemd를 PID 1로 실행합니다](#systemd를-pid-1로-실행합니다)
- [init 시스템이 PID 1로 실행되지 않는 경우 다음과 같은 문제들이 있을 수 있습니다](#init-시스템이-pid-1로-실행되지-않는-경우-다음과-같은-문제들이-있을-수-있습니다)
- [- 1번 프로세스가 `SIGCHLD`를 몰라서 컨테이너 내에 무수히 많은 좀비 프로세스가 생성될 수 있고](#--1번-프로세스가-sigchld를-몰라서-컨테이너-내에-무수히-많은-좀비-프로세스가-생성될-수-있고)
- [- 보통 `/dev/syslog`를 리스닝하는 프로세스가 존재하지 않아서 어떤 프로세스가 `syslog`에 로그를 남길 경우 해당 로그들은 사라지고](#--보통-devsyslog를-리스닝하는-프로세스가-존재하지-않아서-어떤-프로세스가-syslog에-로그를-남길-경우-해당-로그들은-사라지고)
- [- 자체적인 Init 커맨드를 빌드해야 합니다](#--자체적인-init-커맨드를-빌드해야-합니다)
- [`systemd`를 사용하면 서비스 실행 위한 표준 리눅스 API를 제공합니다](#systemd를-사용하면-서비스-실행-위한-표준-리눅스-api를-제공합니다)
        - [1부 ― systemd가 탄생한 배경과 근본 개념 (Why \& How)](#1부--systemd가-탄생한-배경과-근본-개념-why--how)
        - [2부 ― 동작 메커니즘의 큰 그림 (전체 흐름)](#2부--동작-메커니즘의-큰-그림-전체-흐름)
        - [3부 ― 실무에서 꼭 알아야 할 Unit 문법](#3부--실무에서-꼭-알아야-할-unit-문법)
        - [4부 ― 추가로 마스터하면 좋은 고급 주제](#4부--추가로-마스터하면-좋은-고급-주제)
        - [5부 ― 학습 로드맵과 권장 자료](#5부--학습-로드맵과-권장-자료)
    - [정리](#정리)
    - [2 부 ― 잡 스케줄러·상태 추적자의 실제 흐름](#2-부--잡-스케줄러상태-추적자의-실제-흐름)
        - [1. 부팅 임계 경로와 잡(job) 오케스트레이션](#1-부팅-임계-경로와-잡job-오케스트레이션)
        - [2. 런타임 중 재시작·리소스 제어·신호 전파](#2-런타임-중-재시작리소스-제어신호-전파)
        - [3. 타겟 전환, 고립(isolate), 종료 시퀀스](#3-타겟-전환-고립isolate-종료-시퀀스)
    - [3 부 ― 유닛 파일 문법의 실전 설계 패턴](#3-부--유닛-파일-문법의-실전-설계-패턴)
        - [A. 포크형 데몬: PostgreSQL을 systemd 친화적으로](#a-포크형-데몬-postgresql을-systemd-친화적으로)
        - [B. notify 서비스: 고가용성 웹 서버와 소켓 활성화](#b-notify-서비스-고가용성-웹-서버와-소켓-활성화)
        - [C. oneshot+timer: 시스템 백업과 전원 오프 구간 대비](#c-oneshottimer-시스템-백업과-전원-오프-구간-대비)
        - [D. 템플릿 유닛과 인스턴스: 다중 SSH 포트](#d-템플릿-유닛과-인스턴스-다중-ssh-포트)
        - [E. 드롭인(override) 전략: 패키지 업그레이드와 로컬 정책 공존](#e-드롭인override-전략-패키지-업그레이드와-로컬-정책-공존)
        - [F. slice·scope: 다계층 리소스 계약](#f-slicescope-다계층-리소스-계약)
        - [맺음말](#맺음말)
    - [기타](#기타)

## 예제

### Spring application

- [systemd 서비스로 관리](https://docs.spring.io/spring-boot/how-to/deployment/installing.html)

```ini
[Unit]
Description=myapp
After=syslog.target network.target

[Service]
User=myapp
Group=myapp

Environment="JAVA_HOME=/path/to/java/home"

ExecStart=${JAVA_HOME}/bin/java -jar /var/myapp/myapp.jar
ExecStop=/bin/kill -15 $MAINPID
SuccessExitStatus=143

[Install]
WantedBy=multi-user.target
```

다음과 같이 실행합니다.

```sh
#!/bin/bash
sudo systemctl start myapp.service
```

---

# Systemd Unit Syntax

Systemd is a system and service manager for Linux operating systems. It uses "units" to manage services, mounts, timers, and other
system components. Let me explain the systemd unit syntax with detailed examples.

## Basic Structure of a Systemd Unit File

Systemd unit files are typically stored in /etc/systemd/system/ (for system-wide units) or /usr/lib/systemd/system/ (for units
provided by installed packages). They use an INI-style syntax with sections and key-value pairs.

## Example 1: Basic Service Unit

ini
[Unit]
Description=My Example Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=myuser
ExecStart=/usr/bin/myapp --option=value
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target

### Explanation

• **[Unit]** section: Contains metadata and dependencies
  • Description: Human-readable description of the service
  • After: Ensures this service starts after the network is up
  • Wants: Expresses a weaker dependency than "Requires"

• **[Service]** section: Defines how the service behaves
  • Type=simple: Process started with ExecStart is the main process
  • User: The user account to run the service
  • ExecStart: The command to start the service
  • Restart: When to restart the service (on-failure, always, on-abnormal, etc.)
  • RestartSec: How long to wait before restarting

• **[Install]** section: Defines how the unit is enabled
  • WantedBy: Which target should include this service when enabled

## Example 2: Web Server with Dependencies

ini
[Unit]
Description=Web Application Server
Documentation=<https://example.com/docs>
After=network.target postgresql.service redis.service
Requires=postgresql.service
Wants=redis.service

[Service]
Type=notify
User=webuser
Group=webgroup
WorkingDirectory=/var/www/myapp
ExecStartPre=/usr/bin/mkdir -p /var/run/myapp
ExecStart=/usr/bin/myapp serve --port=8080
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
TimeoutStartSec=30
EnvironmentFile=/etc/myapp/env.conf
LimitNOFILE=65536

# Security settings

PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target

### Explanation

• Documentation: URL to documentation
• Requires: Hard dependency (if PostgreSQL fails, this service stops)
• Wants: Soft dependency (Redis is wanted but not required)
• Type=notify: Service will notify systemd when it's ready
• WorkingDirectory: Sets the working directory for the process
• ExecStartPre: Command to run before starting the service
• ExecReload: Command to reload the service
• TimeoutStartSec: Maximum time to wait for service to start
• EnvironmentFile: File containing environment variables
• LimitNOFILE: Sets file descriptor limits
• Security settings like PrivateTmp and ProtectSystem enhance security

## Example 3: Timer Unit (Scheduled Task)

ini

# my-backup.timer

[Unit]
Description=Run backup daily at 2am
Documentation=man:systemd.timer(5)

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true
AccuracySec=1h
RandomizedDelaySec=30min

[Install]
WantedBy=timers.target

ini

# my-backup.service

[Unit]
Description=Daily Backup Service
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/backup-script.sh
User=backup
IOSchedulingClass=idle
CPUSchedulingPolicy=idle
Nice=19

[Install]
WantedBy=multi-user.target

### Explanation

• Timer unit triggers the service at specified times
• OnCalendar: Cron-like time specification (daily at 2am)
• Persistent=true: Run immediately if the system was off at scheduled time
• AccuracySec: Accuracy window for timer (can help reduce system load)
• RandomizedDelaySec: Random delay to avoid many timers starting at once
• Type=oneshot: Service runs once and exits
• IOSchedulingClass=idle: Low I/O priority
• Nice=19: Lowest CPU priority

## Example 4: Mount Unit

ini
[Unit]
Description=Mount NFS Share
After=network-online.target
Wants=network-online.target

[Mount]
What=192.168.1.100:/shared
Where=/mnt/nfs-share
Type=nfs
Options=rw,hard,intr,noatime

[Install]
WantedBy=multi-user.target

### Explanation

• [Mount] section: Specific to mount units
• What: The device/share to mount
• Where: The mount point
• Type: Filesystem type
• Options: Mount options

## Example 5: Socket-Activated Service

ini

# myapp.socket

[Unit]
Description=Socket for MyApp Service
Documentation=<https://example.com/docs>

[Socket]
ListenStream=127.0.0.1:8080
Accept=no

[Install]
WantedBy=sockets.target

ini

# myapp.service

[Unit]
Description=MyApp Service
Requires=myapp.socket
After=myapp.socket

[Service]
Type=notify
ExecStart=/usr/bin/myapp --socket-mode
User=myapp
Restart=on-failure

[Install]
WantedBy=multi-user.target

### Explanation

• Socket activation allows systemd to create the socket and pass it to the service
• ListenStream: Defines TCP socket address and port
• Accept=no: Service handles all connections on the socket
• Service is started only when a connection is made to the socket

## Common Directives

### [Unit] Section

• Description: Human-readable description
• Documentation: URLs or man pages
• After/Before: Ordering dependencies
• Requires: Hard dependencies
• Wants: Soft dependencies
• Conflicts: Units that cannot run simultaneously
• ConditionPathExists, ConditionPathIsDirectory: Conditional activation

### [Service] Section

• Type: simple, forking, oneshot, notify, dbus, idle
• ExecStart, ExecStartPre, ExecStartPost: Commands to run
• ExecStop, ExecReload: Commands for stopping/reloading
• Restart: on-success, on-failure, on-abnormal, on-watchdog, on-abort, always
• RestartSec: Time to wait before restart
• TimeoutStartSec, TimeoutStopSec: Timeouts
• User, Group: User/group to run as
• WorkingDirectory: Working directory
• Environment, EnvironmentFile: Environment variables
• KillMode: control, process, mixed, none
• PrivateTmp, ProtectSystem, ReadOnlyDirectories: Security settings

### [Install] Section

• WantedBy, RequiredBy: Target units that want/require this unit
• Alias: Alternative names for the unit
• Also: Other units to enable/disable together with this one

## Managing Systemd Units

After creating a unit file, you need to:

1. Place it in /etc/systemd/system/
2. Reload systemd: sudo systemctl daemon-reload
3. Enable the service: sudo systemctl enable myservice.service
4. Start the service: sudo systemctl start myservice.service
5. Check status: sudo systemctl status myservice.service

Would you like me to provide more specific examples for any particular type of service?

> The service I want to create is run by 'web' user. And I want to make my app service managed by the 'web'. In this case, I think the service unit file is not placed at /etc/systemd/system. Isn't it? If so, where should I place and manage the file? The server is on-premise linux server. So currently I think do version control the service file at gitlab, and modify the file and review, and deploy by git and some shell script, and reload by `systemctl reload my-app.service`. How about this kind of usage?
You're right about user-specific systemd units. When you want a service to be managed by a non-root user like 'web', you should use
the user-specific systemd instance rather than the system-wide one.

For user-specific services, the unit files are placed in different locations:

1. ~/.config/systemd/user/ - For user-created units
2. /usr/lib/systemd/user/ - For distribution-provided user units

In your case, for the 'web' user, the path would be:
/home/web/.config/systemd/user/my-app.service

Your approach for managing this with version control is excellent. Here's how the workflow would look:

1. Store the service unit file in GitLab
2. Review and modify through merge requests
3. Deploy using git and shell scripts to the correct location
4. Reload the service using the user instance of systemd

For user services, you'd use commands like:
bash

# As the 'web' user

systemctl --user daemon-reload
systemctl --user enable my-app.service
systemctl --user start my-app.service
systemctl --user status my-app.service

If you need to run these commands from another user (like during deployment):
bash

# As another user (e.g., during deployment)

sudo -u web XDG_RUNTIME_DIR=/run/user/$(id -u web) systemctl --user daemon-reload
sudo -u web XDG_RUNTIME_DIR=/run/user/$(id -u web) systemctl --user restart my-app.service

For the service to persist after the user logs out, you'll need to enable lingering:
bash

# As root

loginctl enable-linger web

This approach gives you:
1. Version control for your service definitions
2. Proper review process through GitLab
3. Automated deployment
4. Service management without requiring root privileges
5. Separation of concerns (the 'web' user manages its own services)

# `systemd`가 컨테이너에서 PID 1 프로세스로 동작 중이라면, 컨테이너 종료시 올바른 종료 신호가 필요합니다. 그런데 `systemd`는 `SIGTERM`이나 `SIGKILL`로 제대로 종료되지 않을 수 있습니다

# - Systemd does not exit on sigterm.  Systemd defines that shutdown signal as `SIGRTMIN+3`, docker upstream should send this signal when user does a docker stop

# - `SIGKILL`(`kill -9`): `systemd`는 즉시 종료되지만, 모든 하위 서비스는 제대로 정리되지 않고 그대로 남아 있을 수 있습니다

#

# `SIGRTMIN+3`신호를 받으면 `systemd`는 내부적으로 `systemctl stop`과 유사한 과정을 거쳐 모든 데몬 및 서비스를 순차적으로 종료함으로써 데이터 손실이나 불안정성을 최소화할 수 있습니다

#

# References

# - <https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container#>

# - <https://raby.sh/sigterm-and-pid-1-why-does-a-container-linger-after-receiving-a-sigterm.html>

STOPSIGNAL SIGRTMIN+3

COPY docker-entrypoint.sh /usr/local/bin/

# 컨테이너가 시작될 때 항상 실행되어야 하는 명령을 정의합니다

ENTRYPOINT [ "docker-entrypoint.sh" ]

# systemd를 PID 1로 실행합니다

# init 시스템이 PID 1로 실행되지 않는 경우 다음과 같은 문제들이 있을 수 있습니다

# - 1번 프로세스가 `SIGCHLD`를 몰라서 컨테이너 내에 무수히 많은 좀비 프로세스가 생성될 수 있고

# - 보통 `/dev/syslog`를 리스닝하는 프로세스가 존재하지 않아서 어떤 프로세스가 `syslog`에 로그를 남길 경우 해당 로그들은 사라지고

# - 자체적인 Init 커맨드를 빌드해야 합니다

# `systemd`를 사용하면 서비스 실행 위한 표준 리눅스 API를 제공합니다

CMD ["/lib/systemd/systemd"]

---

### 1부 ― systemd가 탄생한 배경과 근본 개념 (Why & How)

리눅스에서 **init** 시스템은 커널이 PID 1로 넘겨주는 첫 프로세스로서 부팅-종료 전 과정을 지휘합니다. 1990 년대 중반부터 대부분의 배포판은 **SysV init** 스크립트 체계를 사용했습니다. 그러나 대규모 서비스가 늘어나면서 “스크립트-기반 부트 시퀀스”는 아래와 같은 구조적 한계가 점점 두드러졌습니다.

- **병렬 부팅 부재** ― Run-level 스크립트가 순차적으로 실행돼 SSD 시대에도 수십 초 이상이 소요됐습니다.
- **의존성/상태 추적의 부재** ― `/etc/init.d/foo start` 가 성공했는지, 프로세스가 언제 죽었는지 init 자체가 모니터링하지 못했습니다.
- **동적 하드웨어·버스 인식 부족** ― USB 랜카드·Bluetooth 등 “나중에 등장”하는 디바이스가 서비스 기동 시점을 바꿔야 하는데, 스크립트 체계는 ‘언제’라는 시간을 표현하기 힘들었습니다.
- **cgroups/네임스페이스 등 커널 신기능과의 결합 부족** ― 자원 제한·경량 컨테이너 운용을 init 수준에서 통합할 수 없었습니다.

2010 년 **Lennart Poettering** 가 주도해 발표한 **systemd**는 “서비스의 *수명 주기*와 *자원 제어*를 커널 프리미티브(cgroups, inotify, netlink, timerfd…)로 통합 관리하는 표준 PID 1”을 목표로 등장했습니다. 크게 두 가지 철학을 제시합니다.

1. **Everything-is-a-dependency graph.**
   서비스·소켓·타이머·마운트·스왑·디바이스·스냅샷까지 모든 객체를 **Unit** 이라는 추상으로 통일하고, 명시적 의존성을 그래프(Directed Acyclic Graph)로 모델링합니다. 그래프가 준비되는 즉시 병렬 실행하므로 *부팅이 짧아지고* 동적 장치 추가에도 재정렬이 자동입니다.

2. **State tracking by the manager itself.**
   스크립트가 “started”라고만 말하던 과거와 달리, systemd는 `ExecStart=` 로 실행한 메인 PID를 직접 추적하고, cgroup 전체를 *단일 컨트롤러 객체*로 묶어 OOM Kill·리소스 사용량·비정상 종료를 실시간 반영합니다. 그러므로 *재시작 정책* `Restart=`·*리소스 제한* `MemoryMax=` 같은 선언적 옵션이 가능해졌습니다.

---

### 2부 ― 동작 메커니즘의 큰 그림 (전체 흐름)

1. **PID 1 초기화** : 커널이 루트 fs를 마운트한 직후 `/usr/lib/systemd/systemd` 를 실행합니다. PID 1은

   - `system.slice` 라는 최상위 cgroup 을 생성해 *모든* 유닛을 여기에 귀속시키고
   - `/etc/systemd/system/*.wants/` 심볼릭 링크를 따라 기본 타겟(`default.target` → 보통 `graphical.target`)을 활성화합니다.

2. **Unit 로드** :
   A. `/etc/systemd/system` (관리자 override) → B. `/run/systemd/system` (런타임) → C. `/usr/lib/systemd/system` (패키지 제공) 순으로 `.service`, `.socket`, `.timer`… 파일을 읽어 **유효성 검사 후 in-memory graph**에 삽입합니다.
   이 단계에서 `Wants=`·`Requires=`·`After=`·`Before=` 관계가 해석됩니다.

3. **잡(Job) 스케줄링** : 그래프 위상 정렬 결과를 토대로 준비된 유닛을 동시에 `fork()` 하여 실행합니다. 각 잡은 `JOB_STARTING→RUNNING→DONE/FAILED` 로 전이하며, 실패 전이는 `Restart=`, `StartLimitInterval=` 정책이 가로챕니다.

4. **런타임 관리** :

   - cgroup 통계 + inotify + signalfd를 통해 메인 PID나 하위 스레드의 죽음/메모리 쓸림/FD 누수를 감시합니다.
   - `systemctl kill` 은 `SIGTERM→SIGKILL` 을 *cgroup 트리 하위 전파* 방식으로 보냅니다.
   - `journalctl` 이 `sd_journal_stream_fd` 를 통해 유닛별 stdout/stderr 및 커널 메시지를 캡처합니다.

5. **타겟(Target) 전환과 종료** : 사용자 `systemctl isolate rescue.target` 같이 타겟을 바꾸면 graph 전체를 *새 루트 노드*로 재계산 후 필요 없는 유닛을 정리(`stop job`)합니다. `systemd-shutdown` 프로세스가 `/proc/1/exe` 재실행으로 전환돼 clean shutdown을 수행합니다.

---

### 3부 ― 실무에서 꼭 알아야 할 Unit 문법

| 섹션             | 대표 Directive                                                                                              | *왜*/ *어떻게*                                                                |
| -------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| **\[Unit]**    | `Description=`, `Documentation=`, `After=`, `Requires=`                                                   | 의존성 그래프의 *노드 메타데이터* 를 서술합니다.                                              |
| **\[Service]** | `Type=simple/forking/notify/oneshot`                                                                      | 프로세스 모델 제어. `notify` 는 `sd_notify()`로 readiness 시점을 알려 *소켓 활성화* 지연을 줄입니다. |
|                | `ExecStartPre=/bin/mkdir -p /run/foo`<br>`ExecStart=/usr/bin/foo`<br>`ExecReload=/bin/kill -HUP $MAINPID` | *선/후행 단계*를 명확히 분리해 idempotent 한 배포를 도와줍니다.                                |
|                | `Restart=on-failure`, `RestartSec=3s`                                                                     | crash loop 방지 + **self-healing** 사용자 경험을 제공합니다.                           |
|                | `User=web`, `Group=web`, `AmbientCapabilities=CAP_NET_BIND_SERVICE`                                       | 루트 권한 없이 80 포트 바인딩 등 최소 권한 원칙.                                            |
|                | `MemoryMax=500M`, `TasksMax=512`                                                                          | cgroup v2 자원 제한을 한 줄로 기술.                                                 |
|                | `ProtectSystem=full`, `PrivateTmp=yes`, `NoNewPrivileges`                                                 | **seccomp, mount namespaces**를 조합해 경량 샌드박스 생성.                            |
| **\[Install]** | `WantedBy=multi-user.target`, `Alias=httpd.service`                                                       | `systemctl enable` 시 `*.wants/` 심볼릭 링크를 생성하여 *부팅 영속성*을 부여합니다.             |

특히 컨테이너·CI 환경에서는 **drop-in override** 기법이 중요합니다. `/etc/systemd/system/foo.service.d/override.conf` 에 *변경분만* 넣으면 패키지 업그레이드 시 충돌 없이 커스텀 설정을 유지할 수 있습니다.

---

### 4부 ― 추가로 마스터하면 좋은 고급 주제

1. **Socket & Bus Activation** : 데몬을 *lazy-loading* 패턴으로 바꾸어 메모리 푸트를 줄이고, 실패 시 graceful degradation 을 구현합니다. `ListenStream=`, `Accept=` 조합, 그리고 DBus 유닛(.service + .busname) 설계 방식을 실험해 보세요.

2. **Timer Unit** : `OnCalendar=` 문법은 system cron 을 대체하며, `Persistent=true` 가 재부팅 이후 누락된 작업을 보전합니다. DevOps 파이프라인의 백업·로그 로테이션에 특히 유용합니다.

3. **Portable Services / systemd-nspawn** : OCI 컨테이너보다 가벼운 “단일 cgroup + namespaces” 샌드박스를 통해 테스트 환경을 구성할 수 있습니다.

4. **journalctl + Structured Logging** : `systemd-cat --level=warning` 이나 `sd_journal_print()` API로 key-value 형식 로그를 써 두면, `journalctl -u foo.service -o json-pretty` 쿼리를 통해 ELK 스택 없이도 구조화된 분석이 가능합니다.

5. **Resource-Control Interfaces** : `systemctl set-property foo.service CPUQuota=40%` 같이 *런타임 조정*이 가능합니다. 오토스케일링 스크립트에서 유닛별 리소스를 실시간으로 리밸런싱하는 패턴이 자주 쓰입니다.

6. **User Instance & Lingering** : 질문하신 “`systemctl --user` 배포” 흐름은 개발자 머신·VSCode 원격용 백그라운드 데몬, 그리고 멀티테넌트 CGI 서비스에서 유용합니다. `loginctl enable-linger web` 을 통해 로그인 세션 종료 후에도 `PID 1/user@UID.service` 가 유지됩니다.

---

### 5부 ― 학습 로드맵과 권장 자료

1. **man pages** : `man systemd.unit`, `man systemd.exec`, `man systemd.service` 를 ‘소설 읽듯’ 정독하세요. 모든 문법은 결국 이 3개 페이지를 중심으로 확장됩니다.
2. **Freedesktop.org 사양서** : [https://systemd.io](https://systemd.io) 는 설계 배경과 커널 인터페이스 논의를 담은 *화이트페이퍼 모음*입니다. 특히 `CGROUPS` , `BOOTCHART` 문서를 권장합니다.
3. **Red Hat Enterprise Linux Systemd Guide** : 기업 운영 환경에서 추천하는 보안·모니터링 인프라 통합 패턴을 상세히 다룹니다.
4. **“The systemd for Administrators” 시리즈** (Lennart Poettering, 2010–2013) 는 각 서브시스템의 철학적 디자인 결정을 직접 설명합니다.
5. **서킷 브레이크 실습** : 의도적으로 `ExecStart=/bin/false` 유닛을 작성해 `Restart=on-failure` 로 루프를 만든 뒤 `StartLimitBurst`, `StartLimitIntervalSec` 작동을 관찰해 보십시오.

---

## 정리

systemd는 “부팅 스크립트를 돌리는 도구”가 아니라 **리눅스 사용자 공간 전체를 그래프·cgroup·네임스페이스·커널 fd 프리미티브로 모델링하는 런타임**입니다. 따라서 *정확한 의존성 선언* + *상태 추적* + *자원 통제* + *보안 샌드박스* 를 **선언형**으로 표현할 수 있고, 이는 유지보수·관측 가능성·보안 측면에서 기존 SysV 시대를 압도합니다.

이번 1차 답변에서는 역사·동작 원리·필수 문법을 ‘왜/어떻게’ 관점으로 묶어 보았습니다.
다음 단계로 **유닛 디버깅(journalctl, systemd-analyze, coredumpctl)** 과 **컨테이너 내부의 PID 1로 systemd를 쓰는 베스트 프랙티스** 등을 더 깊이 보고 싶으시면 알려 주세요.

systemd가 ‘모든 유닛이 하나의 의존성 그래프를 이루는 런타임’이라는 접근법이 실제로 어떻게 움직이는지, 그리고 그 그래프를 기술할 때 어떤 구문이 어떤 맥락에서 필요해지는지를 세밀한 시나리오를 통해 살펴보면 전체 동작이 훨씬 또렷해진다. 이번 2부에서는 PID 1 내부의 잡 스케줄러와 상태 추적자가 부팅·런타임·종료 국면에서 수행하는 과정을 단계별 사례로 해부하고, 3부에서는 그 과정을 기술하는 단위 언어(unit file 문법)를 실전 예제로 연결해 본다.

---

## 2 부 ― 잡 스케줄러·상태 추적자의 실제 흐름

### 1. 부팅 임계 경로와 잡(job) 오케스트레이션

커널이 `/usr/lib/systemd/systemd`를 올리면 PID 1은 먼저 `/usr/lib/systemd/system/default.target`과 그것이 가리키는 `graphical.target`을 루트 노드로 그래프를 구축한다. 각 유닛 노드는 “트리”가 아니라 “유향 비순환 그래프”다. 서로 다른 서비스가 같은 소켓 유닛을 공유하거나, 하나의 타이머가 둘 이상의 oneshot 유닛을 기동할 수 있기 때문이다. 그래프가 완성되면 systemd는 위상 정렬 결과를 잡 큐에 넣는다. 잡은 `queued → starting → running → done/failed` 상태를 거치며, 어떤 노드가 실패하면 그 노드를 `Requires=`로 의존하던 모든 잡이 연쇄적으로 취소된다. DB 서버가 부팅 중 30 초 안에 리스너를 열지 못했다는 시나리오를 생각해 보자. `postgresql.service`가 `TimeoutStartSec=30`을 초과해 실패하면, 이를 `Requires=`로 선언한 `myapp.service`의 잡 객체는 즉시 “dependency failed”로 표기되고 실행 자체가 되지 않는다. `systemctl list-jobs`로 보면 두 유닛이 나란히 “failed”와 “dead”로 남아 있는 것을 확인할 수 있다. 이 메커니즘은 job 객체가 단 하나만 unit에 매핑된다는 D-Bus 인터페이스 설계로 표준화되어 있어, 관리자 GUI나 ansible 모듈이 동일한 상태 전이를 읽을 수 있다. ([freedesktop.org][1])

이때 부팅 시간이 길어지는 지점을 잡으려면 `systemd-analyze critical-chain`이 의존 그래프를 역추적해 “현재 부팅에서 가장 늦게 준비된 경로”를 출력한다. 실제 현장에서 SSD를 NVMe로 교체했는데도 부팅이 40 초 이상 걸리는 사례를 분석해 보면, 종종 네트워크가 DHCP를 기다리는 동안 `network-online.target`을 `After=`로 둔 모든 서비스가 큐에 묶여 있음을 알 수 있다. 그래프의 임계 경로를 한눈에 표시해 주는 이 도구는 잡 스케줄러의 히스토리를 그대로 재생산하여 보여 준다.

### 2. 런타임 중 재시작·리소스 제어·신호 전파

부팅이 끝나면 PID 1은 각 유닛을 별도 cgroup(v2) 폴더에 격리해 놓고, 메인 PID뿐 아니라 그가 낳은 모든 쓰레드·grand-child까지 포함한 자식 집합을 inotify 없이 커널 내부 계층에서 관찰한다. 프로세스가 크래시로 사라지면 systemd는 ‘unit 상태’와 ‘cgroup 상태’가 달라진 것을 즉시 인지해 `Restart=` 정책을 적용한다. `Restart=on-failure`를 둔 웹 애플리케이션을 예로 들면, segfault가 발생했을 때 잡 스케줄러가 새로운 “restart job”을 만들고, `StartLimitIntervalSec`와 `StartLimitBurst` 창구에서 루프 보호기를 돌린다. Red Hat 가이드가 “자체 헬스 체크 스크립트를 실행해 로그와 메모리를 덤프한 뒤 다시 올려라”는 패턴을 소개하는 이유도 여기에 있다.

신호 처리는 *cgroup 전파*라는 특성을 가진다. 관리자가 `systemctl kill --kill-who=all myapp.service`를 실행하면, PID 1은 `SIGTERM`을 잔여 모든 쓰레드·하위 프로세스에게 동시에 전달한다. docker 컨테이너처럼 ‘PID 1이 시그널을 무시하는 환경’에서는 이 전파가 일어나지 않으므로, systemd를 PID 1로 두면 “좀비가 쌓이지 않는다”라는 이점이 바로 여기서 나온다.

### 3. 타겟 전환, 고립(isolate), 종료 시퀀스

서비스 장애로 긴급 모드로 전환해야 할 때 관리자는 `systemctl isolate rescue.target`을 호출한다. isolate는 “새 루트 노드”를 재정의해 현재 그래프와 목표 그래프의 차집합을 계산한 뒤 필요 없는 모든 유닛에 stop 잡을 발급한다. 이 과정은 부팅 때와 동일한 잡 스케줄러를 역방향으로 이용하며, 따라서 부트와 같은 통일된 정책으로 의존성을 지킨다. 종료 시점에는 PID 1이 `/proc/1/exe`를 재실행해 `systemd-shutdown` 모드로 들어가고, 각 파일 시스템을 `umount -r` 하다가 busy 상태를 감지하면 해당 cgroup을 다시 신호해 가며 순서를 보장한다. 컨테이너에서 `STOPSIGNAL SIGRTMIN+3`이 필요한 것도, systemd가 “정식” 종료 경로를 허용하는 유일한 시그널을 실시간 영역에 배정했기 때문이다.

---

## 3 부 ― 유닛 파일 문법의 실전 설계 패턴

### A. 포크형 데몬: PostgreSQL을 systemd 친화적으로

```ini
# /etc/systemd/system/postgresql.service
[Unit]
Description=PostgreSQL database server
After=network.target
Documentation=https://www.postgresql.org/docs/

[Service]
Type=forking                # postmaster가 부모를 종료하고 새 PID로 전환
User=postgres
Group=postgres
# PIDFile이 없으면 Type=forking은 ready 시점을 알 수 없다
PIDFile=/run/postgresql/14-main.pid
ExecStart=/usr/pgsql-14/bin/pg_ctl start -D /var/lib/pgsql/14/data -s -o "-c config_file=/etc/postgresql/14/main/postgresql.conf"
ExecReload=/usr/pgsql-14/bin/pg_ctl reload -D /var/lib/pgsql/14/data -s
# LSN flush가 끝나지 않았을 때 90 초 이상 기다려야 할 수 있다
TimeoutStopSec=90
Restart=on-failure
# crash loop 방지
StartLimitIntervalSec=300
StartLimitBurst=3

[Install]
WantedBy=multi-user.target
```

위 예제에서 `Type=forking`과 `PIDFile=`은 필연적으로 쌍을 이룬다. 포크형 데몬 모델은 부모 프로세스가 즉시 종료되므로 PID 1이 “준비 완료”를 판단할 근거가 없기 때문이다. `pg_ctl`이 PID 파일을 쓰면, systemd는 그 파일을 열어 메인 PID를 읽고 cgroup에 편입시킨다. 재시작 정책이 메인 PID를 기준으로 동작하기 때문에, 데이터베이스가 core dump 로 죽어도 “유닛 상태”는 정확히 fail 로 표시된다. 문법적으로는 `RestartSec=`만 적으면 시스템 전체의 부트·런타임 경험을 동일한 로직으로 통제할 수 있다.

### B. notify 서비스: 고가용성 웹 서버와 소켓 활성화

웹 애플리케이션을 고속 롤링 배포하려면 “리스너 포트가 열리기 전에 *ready* 신호를 보내지 않는” 보장성이 필수다. 이때 `Type=notify`와 `sd_notify(READY=1)`의 조합이 자연스럽다.

```ini
# webapp.service (user 단 또는 system 단 모두 가능)
[Unit]
Description=My Go Web Server
After=network.target
Requires=webapp.socket   # 소켓 활성화와 짝을 이룸

[Service]
Type=notify
User=web
# Podman 컨테이너를 직접 실행해도 됩니다
ExecStart=/usr/local/bin/webapp --listen-fd=3
# conmon이 --sdnotify=conmon 방식으로 READY=1을 보내도록 설정
NotifyAccess=all
Restart=always
Environment=GODEBUG=madvdontneed=1      # 예: 메모리 단편화 완화

[Install]
WantedBy=multi-user.target
```

```ini
# webapp.socket
[Unit]
Description=WebApp socket

[Socket]
ListenStream=0.0.0.0:8080
Accept=no          # 단일 워커 모델, 리버스프록시 앞단

[Install]
WantedBy=sockets.target
```

이 구성에서는 커널이 TCP 8080을 이미 리슨 상태로 열어 둔 채로 첫 번째 SYN 패킷을 systemd가 잡아 두고, webapp 프로세스가 `sd_notify()`로 READY를 보낸 뒤에야 연결을 전달한다. 빠르게 재배포하더라도 커넥션 큐에 쌓인 SYN 패킷은 유실되지 않는다. webapp이 panic으로 종료되면 `Restart=` 정책이 새 프로세스를 띄우고, 이미 열려 있는 소켓 FD를 그대로 상속하므로 클라이언트 관점에서는 오류가 거의 보이지 않는다.

### C. oneshot+timer: 시스템 백업과 전원 오프 구간 대비

`cron`보다 신뢰도가 높은 타이머 예시는 아래와 같다. `Persistent=true`가 핵심인데, 전원이 꺼진 채 정각을 지나도 부팅 직후 backlog가 실행된다.

```ini
# backup.timer
[Unit]
Description=Run nightly backup

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true
RandomizedDelaySec=15m

[Install]
WantedBy=timers.target
```

```ini
# backup.service
[Unit]
Description=Nightly backup to object storage
After=network-online.target

[Service]
Type=oneshot
User=backup
EnvironmentFile=/etc/backup/vars
# 느린 외부 저장소를 감안해 I/O 우선순위 조정
IOSchedulingClass=idle
Nice=19
ExecStart=/usr/local/bin/backup.sh  # 스크립트 내부에 rsync, rclone 등
```

타이머는 `systemctl list-timers`로 스케줄을 점검할 수 있고, 관리자 입장에서는 `OnCalendar=`에 ISO8601 규칙을 그대로 넣기 때문에 “매월 셋째 주 수요일 22:30” 같은 복잡한 일정도 선언형으로 남긴다.

### D. 템플릿 유닛과 인스턴스: 다중 SSH 포트

SSH를 2222, 2223 두 포트에서 별도 키·설정을 사용해 돌려야 한다면, 템플릿 유닛을 하나만 배포하면 된다.

```ini
# sshd@.service
[Unit]
Description=OpenSSH server per instance

[Service]
Type=notify
ExecStart=/usr/sbin/sshd -D -f /etc/ssh/sshd_config.d/%i.conf
```

`systemctl enable sshd@2222.service`와 같이 인스턴스를 만들면 `%i`가 2222로 치환된다. 템플릿 메커니즘은 실질적으로 “변수 기반 매크로”를 제공하기 때문에, 데이터베이스 샤딩, 다중 VPN 포트, 수십 개의 리버스 프록시 워커처럼 반복적인 설정을 단일 파일로 유지할 수 있다.

### E. 드롭인(override) 전략: 패키지 업그레이드와 로컬 정책 공존

패키지 관리자가 제공한 `.service`를 그대로 수정하면 다음 업데이트 때 충돌이 난다. 이를 피하려면 `/etc/systemd/system/nginx.service.d/override.conf` 같이 별도 파일을 두고 달라지는 행만 적는다.

```ini
# override.conf
[Service]
# 배포 환경별 메모리 한도
MemoryMax=1G
ProtectSystem=strict
```

`systemctl daemon-reload`는 각 유닛의 모든 드롭인을 묶어 최종 effective unit을 다시 계산한다. 큰 규모 시스템에서는 reload 시간이 길어질 수 있으므로(유닛 수가 수천 개일 때 몇 초 이상) Red Hat 문서에서도 변경 후 즉시 reload를 강조한다. ([docs.redhat.com][6])

### F. slice·scope: 다계층 리소스 계약

web 애플리케이션과 백업 작업이 한 서버에서 경쟁할 때, 백업은 CPU 40 % 이내로, 웹은 메모리 4 GB 이내로 묶어야 한다는 식의 정책은 slice와 scope로 집행된다.

```bash
# 웹 계층을 별도 slice로
systemctl set-property --runtime web.slice CPUQuota=60% MemoryMax=4G

# ad-hoc 백업 스크립트를 scope로 묶어 제어
systemd-run --scope -p CPUQuota=30% -p IOWeight=10 bash backup.sh
```

slice는 영구 유닛 파일로 남아 구조를 정의하고, scope는 ‘이미 실행 중인 PID 집합’을 임시로 systemd 관리 하에 편입할 때 쓰인다. 동일한 커널 cgroup 인터페이스를 이용하므로, 향후 eBPF 계측이나 머신 러닝 기반 오토스케일링과도 쉽게 연동된다.

---

### 맺음말

잡 스케줄러는 “의존성 그래프를 구성해 목표 상태를 향해 작업 큐를 자동 정렬”하는 엔진이며, 유닛 파일 문법은 그 그래프의 노드·엣지를 **선언형 언어**로 기술하는 수단이다. 포크·notify·oneshot·template·override·slice 같은 문법 기능은 각각 특정 운영 시나리오에서 필요한 보장 조건(준비 시점의 명확성, 반복 인스턴스 관리, 패키지-로컬 정책 분리, 리소스 계약)을 충족시키기 위해 등장했다. 이러한 맥락을 이해하고 나면 systemd 설정은 더 이상 ‘암기 과목’이 아니라, 리눅스 커널의 cgroup·네임스페이스·fd 프리미티브에 대한 *설계도*를 작성하는 작업이 된다. 실무에서는 journal 구조화 로깅, systemd-analyze critical-chain, 그리고 `systemctl show -p`를 함께 사용해 지금 설명한 메커니즘이 실제로 어떻게 작동하는지 확인해 보는 것이 학습에 큰 도움이 된다.

---

부팅이 끝난 뒤에도 서버가 한동안 멀쩡해 보이다가, 월요일 아침 첫 트래픽이 몰리는 순간에만 웹 애플리케이션이 쿼리 오류를 뿜고 중단되는 상황을 자주 겪던 팀이 있었다. 로그를 자세히 뒤져 보니 root 계정이 작성해 둔 두 개의 셸 스크립트가 문제의 핵심이었다. 하나는 데이터베이스를 시작하고 포그라운드를 떠난 뒤 즉시 0 을 반환했고, 다른 하나는 곧바로 그 다음 줄에서 웹 서버를 띄웠다. 데이터베이스 내부 캐시를 채우는 데 평소보다 오래 걸리면 웹 서버는 아직 열리지 않은 TCP 포트에 접속하려고 하다가 예외를 내고 죽었다. 이때 systemd 로 마이그레이션하면서 처음 한 일은 두 파일을 각각 postgresql.service 와 web.service 라는 두 노드로 옮기고, graph 의 간선을 “ web → postgres ”로 묘사하는 `Requires=` 와 `After=` 키로 표현한 것이다. 이제 PID 1 은 그래프를 따라 topological sort 를 돌린 뒤 데이터베이스 프로세스가 실제로 `READY=1` 신호를 보낼 때까지 큐에서 웹 서버 잡을 꺼내지 않는다. 불시에 부팅 시간이 늘어나도 순서는 절대 뒤집히지 않는다. systemd 가 “서비스를 등록하라”는 요구는 사실상 *의존성 규칙을 커널 타이밍과 독립적인 형태로 선언* 하라는 뜻이다. 수년에 걸친 운영 경험은 이 규칙 선언이 없을 때 매번 사람이 장애를 해석해 가며 재현 불가능한 스크립트를 수정해야 한다는 교훈을 남겼다.

며칠 뒤, 같은 서버에서 새 로깅 기능이 추가된 웹 애플리케이션이 매우 불안정해졌다. 오전 내내 segfault 로 쓰러졌다가 cron 이 도는 새벽 두 시에야 겨우 재부팅되곤 했다. systemd 에서는 `Restart=on-failure`를 한 줄 넣어 두면 메인 PID 가 커널에서 비정상 종료로 보고된 즉시 새로운 start job 이 발행된다. 잡 스케줄러는 루프 폭주를 막기 위해 최근 n 초 안에 몇 번 재시도했는지도 기억한다. Red Hat 운영 가이드는 “재시도 간격과 횟수를 정책적으로 제어하라”며 `StartLimitIntervalSec` 과 `StartLimitBurst` 를 권장한다. 이런 정책이 서비스 파일에 들어 있으면 장애가 일어난 다음 날 회의에서 “어제 08:42 에 네 번 재시작했고 오점검은 08:45 에 끝났다”처럼 상태 이력을 그대로 재현할 수 있다. 등록이 곧 *장애 대응 시나리오를 코드화* 하는 과정이라는 사실이 여기서 분명해진다.

운영자는 또 다른 난제를 안고 있었다. 개발 팀이 남긴 임시 파서 프로세스가 메모리를 계속 차지해 서버 전체가 느려지는 일이 반복됐기 때문이다. systemd 는 모든 서비스를 cgroup v2 트리의 한 가지점 아래로 배치해, 프로세스가 자식이나 스레드를 몇 개를 낳더라도 전체를 하나의 리소스 엔티티로 본다. `MemoryMax=500M` 를 넣은 뒤에는 RSS 가 그 한계를 넘자마자 kernel OOM killer 대신 systemd 자체가 `service state: failed (memory)` 로 바꾸고 즉시 재시작을 시도한다. cgroup 을 직접 쓰면 가능하긴 하지만, Red Hat 문서는 “특별히 구버전 컨트롤러가 필요한 경우가 아니라면 systemd 에게 맡겨라”라고 단언한다. 왜냐하면 리소스 한도와 재시작 정책, 의존성 그래프가 동일한 객체 모델 안에 들어 있어야 장애를 설명할 때 “A 서비스가 메모리를 초과해 재시작하면서 B → C 순서로 영향” 같은 인과를 한눈에 추적할 수 있기 때문이다.

백업 팀은 또 다른 고민이 있었다. 전원 점검 때문에 밤새 서버가 꺼져 있으면 cron 의 02 시 작업이 사라지는 바람에 증분 백업 간격이 깨졌다. systemd 의 타이머는 `OnCalendar=*-*-* 02:00:00` 이라는 문법에 `Persistent=true`를 붙이면 “지난번 스케줄이 부팅 중에 놓쳤다면 부팅 직후 즉시 실행”이라는 규칙을 내장한다. 또한 `RandomizedDelaySec=15m`를 더해 열 대의 서버가 동시에 I/O 를 몰아가는 ‘백업 쓰로틀 역설’을 피할 수 있다. 타이머 역시 유닛이므로 실패 로그는 `journalctl -u backup.service`에서 찾으면 되고, 다음 예약 시각은 `systemctl list-timers` 가 알려 준다. 이렇게 스케줄러 교체 없이도 “작업 언제 돌았나” 기록과 “다음 실행은 언제” 예측이 같은 프레임워크에 담긴다.

팀에선 또 한 명의 개발자가 MinIO 같은 개별 객체 저장소를 자기 계정 아래 돌려야 했다. 기존 방식대로 `nohup … &` 를 쓰면 로그아웃 순간 부모 세션이 사라지면서 프로세스가 시그널을 받아 죽는다. systemd-user 인스턴스에 `~/.config/systemd/user/minio.service` 를 놓고 `loginctl enable-linger $USER` 하면, root PID 1 이 그 사용자의 매니저를 부팅 시점부터 살려 둔다 ([freedesktop.org][4], [serverfault.com][5]). 이제 더는 `nohup` 파일을 찾을 필요도, 재부팅 후 수동으로 스크립트를 돌릴 필요도 없다. *서비스로 등록* 했기 때문에 “사용자가 로그인 중인가”라는 조건과 무관하게 정책이 지켜진다. 남아 있는 질문 하나, “왜 일반 사용자에게도 똑같은 문법을 쓰느냐”는 앞에서 본 연속성 때문이다. syslog, 재시작, 의존성 그래프 해석을 root PID 1 과 동일한 로직으로 재활용할 수 있게 된다.

稀하게 호출되는 관리용 REST 툴을 위해서는 *소켓 활성화*가 빛을 발한다. `mytool.socket` 유닛이 포트 9090 을 먼저 열어 두고 대기하면, 애플리케이션 바이너리는 디스크에서 로드되지 않은 채 메모리도 쓰지 않는다. 첫 TCP SYN 패킷이 들어올 때 systemd 가 바이너리를 fork 해 주고, 서비스가 패닉으로 죽더라도 이미 열린 소켓은 그대로 남아 있어 두 번째 접속자는 실패를 못 느낀다. 소켓 FD 가 3 번부터 넘어온다는 규칙은 `LISTEN_FDS` 환경 변수를 통해 표준화됐는데, 이처럼 *언제 기동할지*라는 문제가 문법 조각 하나(소켓 유닛)와 시스템 호출 몇 줄로 해결되는 것이 ‘서비스로 등록’이 주는 또 다른 큰 이유다.

마지막으로 리소스를 많이 잡아먹는 데이터 분석 스크립트를 한두 시간 돌릴 때, 운영자는 `systemd-run --scope -p CPUQuota=30% bash analyze.sh` 처럼 *scope* 단위를 만들어 준다. 그러면 analyze.sh 와 그 자식 PID 들은 모두 별도 cgroup 하에 묶여, 웹 서비스와 싸우지 않고 사용할 CPU 를 스스로 양보한다. slice-scope 모델은 “서버 하나에 여러 워크로드가 얽힐 때 운영 정책을 계층 구조로 기술”하도록 의도됐다. 스크립트 한 줄로 일회성 작업도 동일한 프레임워크에 편입되기 때문에, 장애 보고서에는 서비스 유닛과 분석 작업이 어떤 순서로 자원을 다퉜는지 같은 문장으로 기록된다.

이렇게 사례를 따라가 보면 “왜 systemd 유닛으로 등록해야 하나”라는 질문은 곧 “서비스의 생애 주기를 *정책 객체*로 올려 놓아야만 실제 운영 시나리오를 재현 및 설명하고, 커널 자원을 믿을 만한 방식으로 배분할 수 있다”는 답으로 귀착된다. 의존성 지시어·재시작 규칙·리소스 한도·ready 시그널·타이머·소켓·linger 같은 문법 조각은 각각이 어떤 현실적 장애나 운영 요구에서 나왔는지를 본다면 암기 대상이 아니라 문제-해결 매핑으로 읽힌다. 결국 systemd 를 통해 서비스를 관리한다는 것은 “프로그램을 실행한다”기보다 “프로그램이 시스템 전체와 맺는 관계를 선언하고, 그 선언을 커널 기능에 직결시킨다”는 의미다.

---

개발자가 직접 유닛 파일을 만들어 배포할 때 가장 먼저 체감하는 지점은 “ini 한 장”이 곧바로 커널 자원과 서비스 생애 주기를 묶어 주는 관문이라는 사실이다. 예 - 들어 당신이 **web** 계정으로 Go 애플리케이션을 운영한다고 가정해 보자. 소스 코드 안에는 아래와 같이 `systemd`가 넘겨 줄 파일 디스크립터 3번을 받아 리스너를 꾸리는 한 줄이 들어 있다.

```go
// main.go
package main

import (
    "log"
    "net"
    "net/http"
    "os"

    "github.com/coreos/go-systemd/daemon" // sd_notify 래퍼
)

func main() {
    // systemd-socket-activate로 전달 받은 3번 FD를 net.Listener로 변환
    l, err := net.FileListener(os.NewFile(3, "listener"))
    if err != nil { log.Fatal(err) }

    // 웹 핸들러 준비
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

    // 준비가 끝났음을 PID 1에게 알림
    daemon.SdNotify(false, "READY=1")

    // 서비스 루프 진입
    if err := http.Serve(l, mux); err != nil { log.Fatal(err) }
}
```

이 코드를 빌드한 뒤, **web** 사용자의 홈 디렉터리에 다음과 같은 두 유닛을 배치한다.

```ini
# ~/.config/systemd/user/web.socket
[Unit]
Description=Example Go HTTP socket, user instance

[Socket]
ListenStream=127.0.0.1:9000
Accept=no

[Install]
WantedBy=sockets.target
```

```ini
# ~/.config/systemd/user/web.service
[Unit]
Description=Example Go HTTP server, user instance
After=network.target
Requires=web.socket

[Service]
Type=notify                      # READY=1 신호를 기다림
ExecStart=/home/web/bin/web-demo # 앞서 빌드한 바이너리
Restart=on-failure               # 예외·SIGSEGV 등 비정상 종료 시 재기동
RestartSec=2s                    # 2초 뒤 재시작
StartLimitIntervalSec=30         # 30초 안에
StartLimitBurst=3                # 3회 넘게 죽으면 hard fail
MemoryMax=500M                   # cgroup v2 메모리 상한
AmbientCapabilities=CAP_NET_BIND_SERVICE
User=web
Group=web

[Install]
WantedBy=default.target
```

`systemctl --user daemon-reload` → `systemctl --user enable --now web.socket` 만 실행해 두면, 부팅 시점에는 **소켓만** 열려 있고 애플리케이션 프로세스는 존재하지 않는다. 첫 HTTP SYN 패킷이 도착하는 순간 PID 1이 `fork()` 한 뒤 FD 3번을 건네준다. 준비 완료 시점은 `daemon.SdNotify`가 내보낸 `READY=1` 패킷으로 결정된다. 이 과정에서 서비스는 **소켓 열기·포트 바인딩·re-fork** 등을 직접 구현할 필요 없이 오로지 비즈니스 로직에 집중한다.

`Type=notify`라는 한 줄이 중요한 까닭은 PID 1이 “메인 루프 진입”이라는 시점을 정확히 잡아야 **After=web.service** 같은 의존 노드가 안전하게 뒤이어 올라올 수 있기 때문이다. 이 메커니즘은 `sd_notify(3)` 시스템 호출 표준에 정의돼 있어 C·Go·Rust 어느 언어에서든 동일하게 사용할 수 있다 ([freedesktop.org][1]).

서비스가 예기치 않게 패닉으로 죽으면 cgroup 안의 **main PID**가 사라졌다는 사실을 systemd가 즉시 감지한다. `Restart=on-failure`는 “0이 아닌 exit, 시그널, core dump” 모두를 재시작 트리거로 해석하는 정책이며 ([freedesktop.org][2]), 동시에 두 개의 레이트-리미트 옵션이 “30초 동안 세 번 이상 실패하면 더 이상 기동하지 말라”는 상위 정책을 말해 준다 ([redhat.com][3], [serverfault.com][4]). 운영자는 `systemctl --user status web.service`로 실패 카운트를 읽거나 `journalctl --user -u web.service`로 패닉 스택트레이스를 확인할 수 있다.

리소스 상한선 역시 유닛 내부에서 선언된다. `MemoryMax=500M`이 있으면 커널 OOM killer 대신 systemd가 먼저 동작해 전체 프로세스 트리를 종료하고 “memory-max” 사유를 로그에 남긴다. `/proc/$PID/cgroup`을 보면 `/user.slice/user-$(id -u).slice/user@$(id -u).service/web.service` 같은 경로가 보이는데, 바로 이 cgroup 디렉터리가 메모리·CPU·IO·PIDs 제한값을 실제로 적용하는 장소다.

동일한 시스템, 동일한 문법으로 백업 스크립트도 제어할 수 있다. 다음은 전원이 꺼졌던 날에도 반드시 한 번은 실행되도록 보장하는 타이머-서비스 쌍이다.

```ini
# /etc/systemd/system/nightly-backup.service
[Unit]
Description=Nightly incremental backup to S3
After=network-online.target

[Service]
Type=oneshot
User=backup
IOSchedulingClass=idle
Nice=19
ExecStart=/usr/local/bin/backup.sh

# /etc/systemd/system/nightly-backup.timer
[Unit]
Description=Run nightly backup at 03:00, even if powered off

[Timer]
OnCalendar=*-*-* 03:00:00
Persistent=true                # 전원 꺼져 있던 시간대도 보정
RandomizedDelaySec=10m         # 여러 대 분산

[Install]
WantedBy=timers.target
```

`systemctl enable --now nightly-backup.timer` 이후 `systemctl list-timers`를 확인하면 마지막 실행 시각과 다음 예약 시각이 함께 표시된다. 만약 서버가 정전으로 새벽 세 시를 놓쳤다면 부팅 직후 backlog가 곧바로 처리되는데, 이 신뢰성은 `Persistent=true` 옵션에 근거해 systemd 가 *마지막 성공 시각*과 *예정 시각*을 journald 데이터베이스에 저장해 두기 때문에 가능하다.

별도의 데몬 없이 잠깐 돌리고 끝낼 데이터 분석도 스코프 유닛으로 격리할 수 있다.

```bash
systemd-run --scope -p CPUQuota=20% -p IOWeight=50 bash analyze.sh
```

명령어가 반환하는 스코프 이름(예: “run-r13f8b43cdbd14.scope”)은 `journalctl -u`로 그대로 추적할 수 있으며, 분석 작업이 CPU 20 % 이상을 점유하려 하면 커널 스케줄러가 알아서 가중치를 조정한다. 동일한 압력 정책을 적용하므로 장기 서비스가 일시적 배치 작업 때문에 OOM 되는 상황을 자연스럽게 막는다.

마지막으로, **컨테이너** 안에서 PID 1을 systemd로 교체하고 싶은 경우 다음과 같은 Dockerfile 조각을 실험할 수 있다.

```Dockerfile
FROM ubuntu:24.04

ENV container=docker
STOPSIGNAL SIGRTMIN+3            # systemd 표준 종료 시그널
RUN apt-get update && apt-get install -y systemd systemd-sysv

VOLUME [ "/sys/fs/cgroup" ]       # cgroup v2 전파
CMD [ "/lib/systemd/systemd" ]
```

호스트가 `docker run --privileged -v /sys/fs/cgroup:/sys/fs/cgroup:rw …` 형태로 기동해 주면, 컨테이너 내부에서도 `systemctl`·`journalctl`·`loginctl`이 완전한 기능을 갖춘다. 이때 `SIGTERM` 대신 `SIGRTMIN+3`을 보내야 정상 셧다운 시퀀스가 실행되는데, freedesktop 사양서가 “PID 1로 동작하는 systemd는 실시간 시그널 3번을 ‘종료’로 예약해 두었다”고 못 박고 있기 때문이다.

지금까지 살펴본 예제는 모두 한 가지 사실을 강조한다. **ini 한 장**이 의존성, 재시작 조건, 리소스 상한, 예약 스케줄, 소켓 전달 같은 엔터프라이즈급 요구 사항을 통합하는 “정책 인터페이스”가 된다는 점이다. 컨테이너이든 베어메탈이든, 루트이든 일반 사용자이든, systemd가 해석하는 문법은 동일하고 그 결과물은 커널 cgroup·네임스페이스·file-descriptor 로 바로 환원된다. 따라서 유닛 파일을 작성한다는 행위는 “스크립트 몇 줄”을 넘어서 **리눅스 자원 모델 위에 추상적 계약서를 올려 두는 작업**이며, 그 계약은 `systemctl show -p` `MainPID`, `systemd-cgls`, `journalctl` 같은 도구로 언제든 증빙할 수 있다. 이런 구조 덕분에 개발자는 코드와 정책을 함께 버전 관리하고, 운영자는 장애 분석과 리소스 회계를 한 곳에서 끝낼 수 있다.

## by gemini

네, 제공해주신 방대한 양의 학습 자료를 바탕으로 systemd에 대한 심도 깊고 체계적인 설명을 제공해 드리겠습니다. 제공된 마크다운 파일은 systemd의 철학부터 실전 문법, 고급 주제까지 매우 훌륭하게 정리되어 있습니다.

이 내용을 바탕으로, 요청하신 대로 각 주제를 명확히 나누어 상세하게 설명하겠습니다.

***

### Part 1: Systemd의 탄생 배경과 핵심 철학 (The "Why")

Systemd는 단순히 기존의 **SysVinit**을 개선한 것이 아니라, 현대적인 리눅스 서버 환경이 마주한 근본적인 문제들을 해결하기 위해 완전히 새로운 패러다임을 제시한 시스템입니다.

#### SysVinit의 구조적 한계

과거의 SysVinit은 다음과 같은 명확한 한계를 가지고 있었습니다:

* **순차적 실행으로 인한 부팅 속도 저하**: 런레벨(Run-level) 스크립트들이 정해진 순서대로 하나씩 실행되어야 했습니다. 이로 인해 SSD와 멀티코어 CPU의 성능을 제대로 활용하지 못하고 부팅 시간이 길어졌습니다.
* **불명확한 의존성 및 상태 관리**: 서비스 간의 의존 관계가 스크립트 파일명의 숫자(예: `S20network`, `S80nginx`)와 같은 암묵적인 규칙에 의존했습니다. 또한, init 프로세스 자체가 특정 서비스가 성공적으로 시작되었는지, 혹은 실행 중에 중단되었는지를 추적할 표준적인 방법이 없었습니다.
* **동적 환경 변화 대응의 어려움**: USB 장치 연결이나 네트워크 상태 변경과 같이 시스템이 실행되는 동안 발생하는 이벤트에 유연하게 반응하여 서비스를 시작하거나 중지하기 어려웠습니다.
* **커널 신기능과의 통합 부재**: Cgroups, 네임스페이스와 같은 커널의 강력한 기능들을 init 시스템 차원에서 통합하여 리소스 제어나 격리 환경을 제공하지 못했습니다.

#### Systemd의 두 가지 핵심 철학

Systemd는 이러한 문제들을 해결하기 위해 두 가지 강력한 철학을 기반으로 설계되었습니다.

1.  **"모든 것은 의존성 그래프다 (Everything-is-a-dependency graph)"**
    Systemd는 서비스, 소켓, 마운트, 타이머 등 시스템의 모든 구성 요소를 **유닛(Unit)**이라는 표준화된 객체로 추상화합니다. 그리고 이 유닛들 사이의 관계(`Requires=`, `After=`, `Wants=`)를 명시적으로 선언하여, 시스템 전체를 하나의 거대한 **유향 비순환 그래프(Directed Acyclic Graph, DAG)**로 모델링합니다.
    * **결과**: 서로 의존성이 없는 유닛들은 최대한 **병렬로 실행**하여 부팅 속도를 획기적으로 단축합니다. 또한, 동적으로 장치가 추가되면 의존성 그래프가 실시간으로 재계산되어 필요한 유닛만 활성화할 수 있습니다.

2.  **"상태 추적은 관리자 자신이 한다 (State tracking by the manager itself)"**
    SysVinit에서는 서비스 스크립트가 "시작했다"고 알려주면 그만이었지만, systemd(PID 1)는 `ExecStart=`로 실행한 프로세스의 주 PID(`$MAINPID`)를 직접 추적합니다. 더 나아가 **Cgroups(Control Groups)**를 이용해 해당 서비스가 생성한 모든 자식 프로세스와 스레드까지 하나의 그룹으로 묶어 감시하고 제어합니다.
    * **결과**: 서비스가 예기치 않게 종료되면 systemd가 즉시 인지하고 `Restart=` 정책에 따라 자동으로 재시작할 수 있습니다 (`Self-healing`). 또한, `MemoryMax=`, `CPUQuota=` 같은 지시어를 통해 유닛별로 할당된 리소스를 정확하게 제한하고 추적하는 것이 가능해졌습니다.

이처럼 systemd는 "스크립트 실행기"를 넘어, **리눅스 사용자 공간 전체를 모델링하고 관리하는 통합 런타임**으로 자리매김했습니다.

***

### Part 2: Systemd 동작 메커니즘 (The "How")

Systemd의 동작은 부팅부터 런타임 관리, 종료까지 일관된 잡(Job) 스케줄링과 상태 추적 메커니즘을 통해 이루어집니다.

1.  **부팅 및 초기화**:
    * 커널은 부팅 마지막 단계에서 `/usr/lib/systemd/systemd`를 PID 1로 실행합니다.
    * PID 1은 가장 먼저 모든 유닛을 관리할 최상위 Cgroup인 `system.slice`를 생성합니다.
    * 기본 목표인 `default.target`(대부분 `multi-user.target`이나 `graphical.target`의 심볼릭 링크)을 활성화하고, 이 타겟에 연결된 `.wants/` 디렉터리의 유닛들을 기반으로 초기 의존성 그래프를 구성합니다.

2.  **유닛 로드 및 그래프 구성**:
    * Systemd는 정해진 우선순위(A. `/etc/systemd/system` → B. `/run/systemd/system` → C. `/usr/lib/systemd/system`)에 따라 유닛 파일들을 읽어 메모리 내에 의존성 그래프를 완성합니다. 관리자가 생성한 파일(`/etc`)이 패키지 기본 파일(`/usr/lib`)을 덮어쓰는 구조입니다.

3.  **잡(Job) 스케줄링 및 실행**:
    * 구성된 그래프를 위상 정렬(topological sort)하여, 의존성이 해결된 유닛들을 잡(Job) 큐에 넣고 병렬로 실행합니다.
    * 각 잡은 `queued → starting → running → done/failed` 상태를 거치며, `systemctl list-jobs` 명령어로 현재 진행 중인 잡들을 확인할 수 있습니다.
    * 만약 어떤 유닛이 `Requires=`로 의존하는 다른 유닛의 시작이 실패하면, 해당 잡은 "dependency failed"로 처리되어 실행되지 않습니다. 이는 서비스 장애가 다른 서비스로 전파되는 것을 막는 중요한 메커니즘입니다.

4.  **런타임 관리 및 프로세스 감시**:
    * 부팅이 완료된 후에도 systemd는 cgroup을 통해 각 서비스의 모든 프로세스(자식, 스레드 포함)를 지속적으로 감시합니다.
    * 프로세스가 비정상 종료(crash)되면, systemd는 이를 즉시 감지하고 `Restart=` 정책에 따라 재시작 잡을 생성합니다. 이때 `StartLimitBurst`와 `StartLimitIntervalSec` 옵션은 무한 재시작 루프를 방지하는 역할을 합니다.
    * `systemctl kill` 명령은 지정된 cgroup 내의 모든 프로세스에게 시그널을 전파하므로, 좀비 프로세스를 남기지 않고 서비스를 확실하게 종료시킬 수 있습니다.

5.  **종료 및 타겟 전환**:
    * `systemctl isolate rescue.target`과 같은 명령은 목표 타겟을 시스템의 새로운 루트 노드로 설정하고, 현재 그래프와의 차집합을 계산하여 불필요한 유닛들을 정해진 의존성 순서에 따라 안전하게 중지시킵니다.
    * 시스템 종료 시에는 `systemd-shutdown` 프로세스가 파일 시스템을 안전하게 unmount하고 모든 서비스를 순차적으로 종료하는 과정을 책임집니다.

***

### Part 3: Systemd Unit 파일 핵심 문법 (The "What")

Systemd 유닛 파일은 INI 스타일의 단순한 텍스트 파일이며, 크게 세 가지 섹션으로 구성됩니다.

| 섹션 | 주요 지시어 | 설명 및 목적 |
| :--- | :--- | :--- |
| **`[Unit]`** | `Description`, `Documentation`, `After`, `Before`, `Requires`, `Wants` | **의존성 그래프의 노드 정보**를 정의합니다. `Description`은 사람이 읽기 위한 설명이며, `After`/`Requires` 등은 그래프의 엣지(연결 관계)를 명시합니다. |
| **`[Service]`** | `Type`, `ExecStart`, `ExecStop`, `ExecReload`, `Restart`, `User`, `Group`, `EnvironmentFile`, `WorkingDirectory`, 보안 설정 (`PrivateTmp` 등) | **서비스의 동작 방식과 실행 환경**을 정의합니다. `Type`은 프로세스 모델(예: `simple`, `forking`, `notify`)을 지정하며, `ExecStart`는 실제 실행할 명령어를 기술합니다. `Restart`는 자가 치유(self-healing) 정책을, `User`는 최소 권한 원칙을 적용하는 데 사용됩니다. |
| **`[Install]`** | `WantedBy`, `RequiredBy`, `Alias` | **`systemctl enable` 명령어가 실행될 때의 동작**을 정의합니다. 예를 들어 `WantedBy=multi-user.target`은 `systemctl enable myapp.service` 실행 시 `/etc/systemd/system/multi-user.target.wants/myapp.service` 심볼릭 링크를 생성하여 부팅 시 자동으로 시작되도록 설정합니다. |

#### 서비스 타입(`Type=`)의 이해

* `Type=simple`: `ExecStart`로 지정된 프로세스가 주 프로세스입니다. 가장 일반적이고 권장되는 방식입니다.
* `Type=forking`: `ExecStart`로 지정된 프로세스가 자식 프로세스를 생성한 후 자신은 종료되는 전통적인 데몬 모델입니다. 이 경우 systemd가 주 프로세스를 추적할 수 있도록 `PIDFile=`을 함께 명시해야 합니다.
* `Type=notify`: 서비스가 준비 완료 시점을 `sd_notify()` 함수를 통해 systemd에게 명시적으로 알리는 방식입니다. 초기화에 시간이 걸리는 서비스의 시작 완료 시점을 정확히 파악할 수 있어, 의존성 관리에 매우 유용합니다.
* `Type=oneshot`: 한 번만 실행되고 바로 종료되는 스크립트 등에 사용됩니다. 타이머와 함께 사용되는 경우가 많습니다.

***

### Part 4: 고급 주제 및 실전 활용법

#### 사용자 서비스(User Instance)와 Lingering

* **개념**: Systemd는 시스템 전체를 관리하는 인스턴스(PID 1) 외에, 각 사용자별로 독립된 `systemd --user` 인스턴스를 실행할 수 있습니다.
* **유닛 파일 위치**: 사용자 유닛은 `~/.config/systemd/user/` 디렉터리에 위치합니다.
* **Lingering 활성화**: 기본적으로 사용자 인스턴스는 해당 사용자가 로그인해 있는 동안에만 활성화됩니다. 사용자가 로그아웃한 후에도 서비스가 계속 실행되게 하려면, `sudo loginctl enable-linger <username>` 명령으로 'lingering'을 활성화해야 합니다. 이는 개발 환경의 백그라운드 서비스나 멀티테넌트 환경에 매우 유용합니다.

#### 소켓 활성화 (Socket Activation)

* **개념**: 서비스가 항상 실행되어 메모리를 차지하는 대신, systemd가 먼저 서비스의 소켓(예: TCP 포트)을 열어두고 있다가 첫 연결 요청이 들어오면 그제야 해당 서비스를 활성화하는 방식입니다.
* **장점**: 시스템 리소스를 매우 효율적으로 사용할 수 있으며, 서비스가 재시작되는 동안에도 클라이언트의 연결 요청을 유실하지 않고 큐에 보관할 수 있어 가용성을 높입니다.
* **구현**: `.socket` 유닛과 `.service` 유닛을 쌍으로 만들어 사용합니다. 서비스는 `sd_notify(READY=1)`를 통해 준비 완료 신호를 보내야 합니다.

#### 타이머 (Timer Units)

* **개념**: `cron`을 대체하는 현대적인 방식의 작업 스케줄러입니다. `.timer` 유닛이 `.service` 유닛을 정해진 시간에 활성화합니다.
* **주요 기능**:
    * `OnCalendar=`: `cron`과 유사한 형식으로 실행 시간을 지정합니다.
    * `Persistent=true`: 서버가 꺼져 있어서 정해진 시간에 작업을 실행하지 못했다면, 다음 부팅 시 즉시 실행해 줍니다.
    * `RandomizedDelaySec=`: 여러 서버에서 동시에 작업이 시작되어 부하가 몰리는 것을 방지하기 위해 임의의 지연 시간을 추가합니다.

#### 컨테이너 환경에서의 Systemd (PID 1)

* **문제점**: Docker와 같은 컨테이너 환경에서 `docker stop` 명령은 기본적으로 `SIGTERM` 신호를 PID 1에게 보냅니다. 하지만 systemd는 이 신호를 일반적인 종료 신호로 해석하지 않습니다. `SIGKILL`은 하위 서비스를 정리할 시간을 주지 않고 즉시 종료시켜 데이터 무결성을 해칠 수 있습니다.
* **해결책**: Systemd를 PID 1로 사용하는 컨테이너의 경우, 정상적인 종료 시퀀스를 트리거하기 위해 **`SIGRTMIN+3`** 신호를 사용해야 합니다. Dockerfile에 `STOPSIGNAL SIGRTMIN+3`을 명시하면 `docker stop` 시 이 신호가 전달됩니다. 이 신호를 받은 systemd는 `systemctl stop`과 유사하게 모든 유닛을 순차적이고 안전하게 종료합니다.
* **실행 환경**: 컨테이너 내에서 systemd를 완전한 기능으로 사용하려면, cgroup 파일 시스템에 접근 권한이 필요합니다 (예: `docker run --privileged -v /sys/fs/cgroup:/sys/fs/cgroup:rw`).

이처럼 systemd는 단순한 서비스 실행 도구를 넘어, 현대 리눅스 시스템의 안정성, 성능, 관리 효율성을 극대화하는 핵심적인 프레임워크입니다. 제공해주신 자료를 통해 이미 깊이 이해하고 계시며, 이 설명을 통해 개념들이 더욱 명확하게 정리되셨기를 바랍니다.

## 기타

- [How to create a systemd service in Linux](https://linuxhandbook.com/create-systemd-services/)
- [Managing systemd Services](https://documentation.suse.com/smart/systems-management/html/systemd-management/index.html)
