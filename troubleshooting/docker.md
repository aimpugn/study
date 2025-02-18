# Docker

- [Docker](#docker)
    - [이미지를 깃헙 패키지에서 가져오지 못하는 이슈](#이미지를-깃헙-패키지에서-가져오지-못하는-이슈)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [At least one invalid signature was encountered](#at-least-one-invalid-signature-was-encountered)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [No space left on device](#no-space-left-on-device)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [postgres](#postgres)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [Err:5 `http://ports.ubuntu.com/ubuntu-ports` kinetic Release](#err5-httpportsubuntucomubuntu-ports-kinetic-release)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [Failed to create /init.scope control group: Read-only file system](#failed-to-create-initscope-control-group-read-only-file-system)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
        - [참고](#참고)

## 이미지를 깃헙 패키지에서 가져오지 못하는 이슈

### 문제

```shell
❯ docker pull ghcr\.io/some-orgcore_interface-builder:latest

^^^^ 아무것도 출력 안되고 hang
```

### 원인

알수없음. 깃헙 사이트에서도 접근이 불가

### 해결

이미지 파일을 받아서 `docker load` CLI로 직접 반영

```shell
❯ docker load --input ./core_interface-builder
e7a0eed9531e: Loading layer [==================================================>]  66.76MB/66.76MB
7d935b9bf332: Loading layer [==================================================>]   17.5MB/17.5MB
bfe5659d13eb: Loading layer [==================================================>]  8.692MB/8.692MB
9604c5dbb6f8: Loading layer [==================================================>]  13.83MB/13.83MB
c361bb48cf30: Loading layer [==================================================>]  797.7kB/797.7kB
5f70bf18a086: Loading layer [==================================================>]  1.024kB/1.024kB
8d0410f27670: Loading layer [==================================================>]  25.09kB/25.09kB
b5e8c381d2c9: Loading layer [==================================================>]  574.5kB/574.5kB
ede6abd2ac04: Loading layer [==================================================>]  63.91MB/63.91MB
103aa0be0f14: Loading layer [==================================================>]  13.82kB/13.82kB
Loaded image: ghcr\.io/some-orgcore_interface-builder:latest
```

## At least one invalid signature was encountered

### 문제

```log
╰─ docker build . --tag sandbox

... 생략 ...
                                                                              0.2s
 => ERROR [2/6] RUN apt-get update -y
------
 > [2/6] RUN apt-get update  -y:

0.858 Get:1 http://ports.ubuntu.com/ubuntu-ports jammy InRelease [270 kB]
2.368 Err:1 http://ports.ubuntu.com/ubuntu-ports jammy InRelease
2.368   At least one invalid signature was encountered.
2.627 Get:2 http://ports.ubuntu.com/ubuntu-ports jammy-updates InRelease [119 kB]
2.760 Err:2 http://ports.ubuntu.com/ubuntu-ports jammy-updates InRelease
2.760   At least one invalid signature was encountered.
3.008 Get:3 http://ports.ubuntu.com/ubuntu-ports jammy-backports InRelease [109 kB]
3.165 Err:3 http://ports.ubuntu.com/ubuntu-ports jammy-backports InRelease
3.165   At least one invalid signature was encountered.
3.409 Get:4 http://ports.ubuntu.com/ubuntu-ports jammy-security InRelease [110 kB]
3.574 Err:4 http://ports.ubuntu.com/ubuntu-ports jammy-security InRelease
3.574   At least one invalid signature was encountered.
3.578 Reading package lists...
3.594 W: GPG error: http://ports.ubuntu.com/ubuntu-ports jammy InRelease: At least one invalid signature was encountered.
3.594 E: The repository 'http://ports.ubuntu.com/ubuntu-ports jammy InRelease' is not signed.
3.594 W: GPG error: http://ports.ubuntu.com/ubuntu-ports jammy-updates InRelease: At least one invalid signature was encountered.
3.594 E: The repository 'http://ports.ubuntu.com/ubuntu-ports jammy-updates InRelease' is not signed.
3.594 W: GPG error: http://ports.ubuntu.com/ubuntu-ports jammy-backports InRelease: At least one invalid signature was encountered.
3.594 E: The repository 'http://ports.ubuntu.com/ubuntu-ports jammy-backports InRelease' is not signed.
3.594 W: GPG error: http://ports.ubuntu.com/ubuntu-ports jammy-security InRelease: At least one invalid signature was encountered.
3.594 E: The repository 'http://ports.ubuntu.com/ubuntu-ports jammy-security InRelease' is not signed.

```

### 원인

루트 파티션이 꽉 차서(아마도 apt를 통해 패키지를 다운로드하려고 너무 많이 시도한 것 같음) sudo apt clean을 실행하면 문제가 해결되었습니다.

### 해결

```shell
docker system df # which can show disk usage and size of 'Build Cache'
docker image prune # add -f or --force to not prompt for confirmation
docker container prune # add -f or --force to not prompt for confirmation
```

```shell
docker system prune --force
```

## No space left on device

### 문제

```log
2024-01-11 15:23:58 2024-01-11 06:23:58.560 UTC [41] FATAL:  terminating connection due to administrator command
2024-01-11 15:23:58 2024-01-11 06:23:58.561 UTC [32] FATAL:  terminating connection due to administrator command
2024-01-11 15:23:58 2024-01-11 06:23:58.581 UTC [1] LOG:  background worker "logical replication launcher" (PID 31) exited with exit code 1
2024-01-11 15:23:58 2024-01-11 06:23:58.583 UTC [26] LOG:  shutting down
2024-01-11 15:23:58 2024-01-11 06:23:58.585 UTC [26] LOG:  checkpoint starting: shutdown immediate
2024-01-11 15:23:58 2024-01-11 06:23:58.586 UTC [26] PANIC:  could not write to file "pg_logical/replorigin_checkpoint.tmp": No space left on device
2024-01-11 15:23:58 2024-01-11 06:23:58.588 UTC [1] LOG:  checkpointer process (PID 26) was terminated by signal 6: Aborted
2024-01-11 15:23:58 2024-01-11 06:23:58.588 UTC [1] LOG:  terminating any other active server processes
2024-01-11 15:23:58 2024-01-11 06:23:58.588 UTC [1] LOG:  abnormal database system shutdown
2024-01-11 15:23:58 2024-01-11 06:23:58.593 UTC [1] LOG:  database system is shut down
2024-01-11 15:29:14 2024-01-11 06:29:14.259 UTC [1] FATAL:  could not write lock file "postmaster.pid": No space left on device
2024-01-11 15:32:15 2024-01-11 06:32:15.032 UTC [1] FATAL:  could not write lock file "postmaster.pid": No space left on device
2024-01-11 15:32:17 2024-01-11 06:32:17.488 UTC [1] FATAL:  could not write lock file "postmaster.pid": No space left on device
```

그리고 testcontainers 실행할 때도 아래와 같은 에러 발생

```log
Data page checksums are disabled.

fixing permissions on existing directory /var/lib/postgresql/data ... ok
initdb: error: could not create directory "/var/lib/postgresql/data/pg_wal": No space left on device
initdb: removing contents of data directory "/var/lib/postgresql/data"
```

### 원인

### 해결

```bash
# 자원 정리: https://forums.docker.com/t/docker-no-space-left-on-device/69205/3
docker system prune --all --force
```

## postgres

### 문제

```log
Error: Database is uninitialized and superuser password is not specified.
       You must specify POSTGRES_PASSWORD to a non-empty value for the
       superuser. For example, "-e POSTGRES_PASSWORD=password" on "docker run".

       You may also use "POSTGRES_HOST_AUTH_METHOD=trust" to allow all
       connections without a password. This is *not* recommended.

       See PostgreSQL documentation about "trust":
       https://www.postgresql.org/docs/current/auth-trust.html
```

### 원인

이 에러는 PostgreSQL Docker 컨테이너를 초기화하는 과정에서 발생한다.
PostgreSQL Docker 이미지를 사용할 때, 데이터베이스의 슈퍼유저(보통 `postgres` 유저)에 대한 비밀번호를 설정하지 않았기 때문에 이 문제가 발생. PostgreSQL은 보안상의 이유로 슈퍼유저 비밀번호를 반드시 설정하도록 요구한다.

1. Docker 컨테이너를 실행할 때 슈퍼유저의 비밀번호를 설정하거나,
2. 비밀번호 없이 접근을 허용하는 환경 변수를 설정

### 해결

```bash
docker run -e POSTGRES_PASSWORD=password -d postgres
```

```bash
# 비밀번호 없이 접근 허용
docker run -e POSTGRES_HOST_AUTH_METHOD=trust -d postgres
```

## Err:5 `http://ports.ubuntu.com/ubuntu-ports` kinetic Release

### 문제

```sh
 > [2/9] RUN apt update -y     && apt-get install whiptail -y:
0.081
0.081 WARNING: apt does not have a stable CLI interface. Use with caution in scripts.
0.081
0.941 Ign:1 http://ports.ubuntu.com/ubuntu-ports kinetic InRelease
1.246 Ign:2 http://ports.ubuntu.com/ubuntu-ports kinetic-updates InRelease
1.512 Ign:3 http://ports.ubuntu.com/ubuntu-ports kinetic-backports InRelease
1.860 Ign:4 http://ports.ubuntu.com/ubuntu-ports kinetic-security InRelease
2.167 Err:5 http://ports.ubuntu.com/ubuntu-ports kinetic Release
2.167   404  Not Found [IP: 185.125.190.36 80]
2.475 Err:6 http://ports.ubuntu.com/ubuntu-ports kinetic-updates Release
2.475   404  Not Found [IP: 185.125.190.36 80]
2.783 Err:7 http://ports.ubuntu.com/ubuntu-ports kinetic-backports Release
2.783   404  Not Found [IP: 185.125.190.36 80]
3.049 Err:8 http://ports.ubuntu.com/ubuntu-ports kinetic-security Release
3.049   404  Not Found [IP: 185.125.190.36 80]
3.052 Reading package lists...
3.062 E: The repository 'http://ports.ubuntu.com/ubuntu-ports kinetic Release' does not have a Release file.
3.062 E: The repository 'http://ports.ubuntu.com/ubuntu-ports kinetic-updates Release' does not have a Release file.
3.062 E: The repository 'http://ports.ubuntu.com/ubuntu-ports kinetic-backports Release' does not have a Release file.
3.062 E: The repository 'http://ports.ubuntu.com/ubuntu-ports kinetic-security Release' does not have a Release file.
------
Dockerfile:9
--------------------
   8 |     # default settings
   9 | >>> RUN apt update -y \
  10 | >>>     # to prevent `debconf: (No usable dialog-like program is installed` prompt
  11 | >>>     && apt-get install whiptail -y
  12 |
--------------------
ERROR: failed to solve: process "/bin/sh -c apt update -y     && apt-get install whiptail -y" did not complete successfully: exit code: 100
```

### 원인

[Ubuntu 22.10 Kinetic can't update archive.ubuntu.com and ignored it [duplicate]](https://askubuntu.com/questions/1458489/ubuntu-22-10-kinetic-cant-update-archive-ubuntu-com-and-ignored-it)

22.10이 Dockerfile FROM 지시자에 떠서 설치해보려고 했는데,
[Release](https://wiki.ubuntu.com/Releases) 페이지를 확인해보니 22.10이 없다!

### 해결

`24.04`로 수정합니다.

```Dockerfile
FROM ubuntu:24.04
```

## Failed to create /init.scope control group: Read-only file system

### 문제

MacOS에서 systemd로 초기화되는 이미지를 생성합니다.

<details>
<summary>Dockerfile</summary>

```Dockerfile
FROM ubuntu:latest

USER root

RUN apt-get update
RUN yes | unminimize
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y \
    init systemd systemd-sysv sudo locales \
    man-db \
    build-essential glibc-doc \
    wget curl git zsh vim \
    && apt-get autoremove

RUN locale-gen --no-archive ko_KR.UTF-8

ENV LANG=ko_KR.UTF-8 LC_ALL=ko_KR.UTF-8

ENV USER=ubuntu
ENV HOME="/home/${USER}"

RUN usermod -aG sudo ${USER} \
    && echo "${USER} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/ubuntu

USER ${USER}
WORKDIR ${HOME}
RUN rm -rf ~/.oh-my-zsh
RUN yes | sh -c "$(wget https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)"

USER root
STOPSIGNAL SIGRTMIN+3
CMD ["/lib/systemd/systemd"]
```

</details>

그리고 `--privileged` 옵션 없이 실행하려고 하면 다음과 같은 에러가 발생합니다.

```log
systemd 255.4-1ubuntu8.5 running in system mode (+PAM +AUDIT +SELINUX +APPARMOR +IMA +SMACK +SECCOMP +GCRYPT -GNUTLS +OPENSSL +ACL +BLKID +CURL +ELFUTILS +FIDO2 +IDN2 -IDN +IPTC +KMOD +LIBCRYPTSETUP +LIBFDISK +PCRE2 -PWQUALITY +P11KIT +QRENCODE +TPM2 +BZIP2 +LZ4 +XZ +ZLIB +ZSTD -BPF_FRAMEWORK -XKBCOMMON +UTMP +SYSVINIT default-hierarchy=unified)
Detected virtualization docker.
Detected architecture arm64.

Welcome to Ubuntu 24.04.2 LTS!

Failed to create /init.scope control group: Read-only file system
Failed to allocate manager object: Read-only file system
[!!!!!!] Failed to allocate manager object.
Exiting PID 1...
```

### 원인

`init.scope`가 무엇인지 알기 위해서는 우선 `systemd`에서 유닛이 무엇인지 알아야 합니다.
- 유닛은 `systemd`가 관리하는 개별적인 객체를 의미합니다.
- 유닛의 종류
    - 서비스 유닛(`.service`): `nginx.service`, `ssh.service` 등 데몬 실행
    - 타겟 유닛(`.target`): `multi-user.target`, `graphical.target` 등 부팅 레벨 설정
    - 마운트 유닛(`.mount`): `/var.mount` 등 파일 시스템 마운트
    - 스왑 유닛(`.swap`): `swapfile.swap` 등 스왑 설정
    - 디바이스 유닛(`.device`): `/dev/sda1.device` 등 디바이스 관리
    - 스코프 유닛(`.scope`): `init.scope`(시스템의 최상위 cgroup을 관리하는 스코프), `user-1000.scope`(특정 사용자의 세션 프로세스를 관리하는 스코프) 등 외부에서 생성된 프로세스 그룹 관리

서비스(`.service`)는 `systemd`가 직접 생성하는 프로세스를 관리하지만, [스코프(`.scope`)](https://www.freedesktop.org/software/systemd/man/latest/systemd.scope.html)는 외부에서 생성된 프로세스를 그룹화하는 역할을 합니다.

`init.scope`는 `systemd` 자체가 실행되는 최상위 컨트롤 그룹으로, 시스템 초기화와 관련된 프로세스들을 관리하기 위한 특별한 scope unit입니다. [이 scope unit은 시스템과 서비스 관리자(PID 1)가 상주하는 곳](https://man7.org/linux/man-pages/man7/systemd.special.7.html)입니다. 시스템이 실행되는 동안 활성화되어 있습니다.

`systemd`는 각 프로세스를 `cgroup`을 통해 관리하며, 이를 통해 프로세스의 자원 사용을 추적하고 제어합니다.
`init.scope`는 시스템 초기화와 관련된 프로세스들을 포함하는 `cgroup`으로, 시스템 부팅 시 `systemd`에 의해 생성됩니다.
일반적으로 `/sys/fs/cgroup` 디렉토리에 마운트되며, 각 프로세스는 해당 디렉토리 내의 특정 서브디렉토리에 위치하게 됩니다.

따라서 컨테이너 환경에서 `systemd`를 실행할 때 컨테이너 내부에서 `cgroup` 파일 시스템에 쓰기 권한이 필요하므로, 다음과 같은 경우 `systemd`는 `cgroup`을 생성하지 못합니다.
- 컨테이너의 파일 시스템이 읽기 전용으로 마운트되어 있는 경우
- `/sys/fs/cgroup` 디렉토리에 대한 쓰기 권한이 제한되어 있는 경우

현재 `systemd`로 초기화되는 컨테이너를 테스트하는 환경은 MacOS & OrbStack 입니다.
MacOS는 Linux 커널을 제공하지 않으며, [OrbStack은 WSL 2와 비슷한 Linux VM을 사용](https://docs.orbstack.dev/architecture)한다고 합니다. 문제가 발생한 `/init.scope`는 Linux VM의 `/sys/fs/cgroup`입니다.

### 해결

앞서 확인했듯이, `systemd`는 PID 1로 실행될 때 자체 `cgroup` 네임스페이스를 생성하려 시도합니다.

```log
Failed to create /init.scope control group: Read-only file system
```

이를 위해서는 `cgroup`에 대한 쓰기가 가능해야 하는데, [`--privileged` 옵션](https://docs.docker.com/reference/cli/docker/container/run/#privileged)을 사용하거나 [Running systemd in a non-privileged container](https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container#) 설명처럼 `--privileged` 옵션 없이 `--volume`과 `--tmpfs` 옵션 등을 사용할 수 있습니다.

`--privileged` 옵션을 사용하면 컨테이너가 호스트와 거의 동일한 권한을 가지기 때문에, 컨테이너 내부의 프로세스가 호스트 시스템을 변경할 가능성이 있습니다.
- 모든 리눅스 커널 기능 활성화
- 기본 [`seccomp`(secure computing mode)](https://man7.org/linux/man-pages/man2/seccomp.2.html) profile 비활성화
- 기본 [AppArmor](https://ubuntu.com/server/docs/apparmor) profile 비활성화
- SELinux process label 비활성화
- 모든 호스트 장치에 대한 액세스 권한 부여
- `/sys` 읽고 쓰기 가능
- cgroups 마운트 읽고 쓰기 가능

컨테이너 격리가 약해지고, 악의적인 코드가 실행될 경우 호스트 시스템까지 영향을 줄 가능성이 있습니다. 따라서 보안상의 이유로 권장되지 않습니다.

`--privileged` 옵션 없는 방식을 따르려 하는데, 맥북에서 OrbStack을 사용해서인지 가이드대로 동작하지 않습니다.
대신 다음 옵션들을 사용합니다.
- `--tmpfs /sys/fs/cgroup:rw`
    - 컨테이너 내의 `/sys/fs/cgroup`을 메모리 기반 파일 시스템(`tmpfs`)으로 마운트합니다.
    - `systemd`가 사용하는 `/sys/fs/cgroup/systemd`를 `cgroup2`로 마운트하는 것은 [docker-entrypoint.sh](../docker/Dockerfiles/ubuntu/docker-entrypoint.sh)에서 수행합니다.
- `--cap-add SYS_ADMIN`: 시스템 관리자 권한을 부여합니다. 권한을 부여하지 않으면 `systemd-resolved.service` 시작에 실패합니다.

    ```log
    ➜  ~ systemctl status systemd-resolved.service
    × systemd-resolved.service - Network Name Resolution
         Loaded: loaded (/usr/lib/systemd/system/systemd-resolved.service; enabled; preset: enabled)
         Active: failed (Result: exit-code) since Tue 2025-02-18 14:56:53 UTC; 1s ago
           Docs: man:systemd-resolved.service(8)
                 man:org.freedesktop.resolve1(5)
                 https://www.freedesktop.org/wiki/Software/systemd/    writing-network-configuration-managers
                 https://www.freedesktop.org/wiki/Software/systemd/writing-resolver-clients
        Process: 78 ExecStart=/usr/lib/systemd/systemd-resolved (code=exited, status=217/USER)
       Main PID: 78 (code=exited, status=217/USER)
            CPU: 1ms

     2월 18 14:56:53 8ec7c9848d0a systemd[1]: systemd-resolved.service: Scheduled restart job, restart counter is at 5.
     2월 18 14:56:53 8ec7c9848d0a systemd[1]: systemd-resolved.service: Start request repeated too quickly.
     2월 18 14:56:53 8ec7c9848d0a systemd[1]: systemd-resolved.service: Failed with result 'exit-code'.
     2월 18 14:56:53 8ec7c9848d0a systemd[1]: Failed to start systemd-resolved.service - Network Name Resolution.
    ```

    반면, `SYS_ADMIN` 권한을 부여하면 다음과 같이 정상적으로 실행됩니다.

    ```log
    ➜  ~ systemctl status systemd-resolved.service
    ● systemd-resolved.service - Network Name Resolution
         Loaded: loaded (/usr/lib/systemd/system/systemd-resolved.service; enabled; preset: enabled)
         Active: active (running) since Tue 2025-02-18 15:21:03 UTC; 6s ago
           Docs: man:systemd-resolved.service(8)
                 man:org.freedesktop.resolve1(5)
                 https://www.freedesktop.org/wiki/Software/systemd/writing-network-configuration-managers
                 https://www.freedesktop.org/wiki/Software/systemd/writing-resolver-clients
       Main PID: 63 (systemd-resolve)
         Status: "Processing requests..."
            CPU: 33ms
         CGroup: /docker/7e199e73ef26bf4112269b96b501f01d42c65554d8d907ec38ab551b7bcc1784/system.slice/    systemd-resolved.service
                 └─63 /usr/lib/systemd/systemd-resolved

     2월 18 15:21:03 7e199e73ef26 systemd[1]: Starting systemd-resolved.service - Network Name Resolution...
     2월 18 15:21:03 7e199e73ef26 systemd-resolved[63]: Positive Trust Anchors:
     2월 18 15:21:03 7e199e73ef26 systemd-resolved[63]: . IN DS 20326 8 2 e06d44b80b8f1d39a95c0b0d7c65d08458e880409bbc683457104237c7f8ec8d
     2월 18 15:21:03 7e199e73ef26 systemd-resolved[63]: Negative trust anchors: home.arpa 10.in-addr.arpa 16.172.in-addr.arpa 17.172.in-a>
     2월 18 15:21:03 7e199e73ef26 systemd-resolved[63]: Using system hostname '7e199e73ef26'.
     2월 18 15:21:03 7e199e73ef26 systemd[1]: Started systemd-resolved.service - Network Name Resolution.
    ```

최종적으로는 다음과 같이 백그라운드에서 컨테이너를 실행하고 `docker exec -u ubuntu -it ubuntu-env /bin/zsh`으로 컨테이너에 진입합니다. (자세한 내용은 [Dockerfile](../docker/Dockerfiles/ubuntu/Dockerfile), [run.sh](../docker/Dockerfiles/ubuntu/run.sh), [docker-entrypoint.sh](../docker/Dockerfiles/ubuntu/docker-entrypoint.sh) 등 참고)

```sh
docker run \
    --name ubuntu-env \
    --tmpfs /run \
    --tmpfs /sys/fs/cgroup:rw \
    --cap-add SYS_ADMIN \
    -d
    ubuntu-env
```

실행되는 순서는 다음과 같습니다:
1. `docker run` 실행 시 `--tmpfs /sys/fs/cgroup`과 같은 마운트 옵션이 적용됩니다.
2. 컨테이너 내의 `/sys/fs/cgroup`을 메모리 기반 파일 시스템(`tmpfs`)으로 마운트됩니다.
3. `ENTRYPOINT`에서 지정한 실행 파일(예: `docker-entrypoint.sh`)이 실행되며, 이 프로세스가 1번 프로세스가 됩니다.
4. `docker-entrypoint.sh`에서 `/sys/fs/cgroup/systemd` 디렉토리를 생성하고 `cgroup2` 파일 시스템을 마운트합니다.
5. `exec "$@"`를 실행하여 `CMD`에 지정된 `/lib/systemd/systemd`이 실행되고, `/lib/systemd/systemd` 프로세스가 1번 프로세스가 됩니다.

### 참고

- [Container 격리 기술 이해하기](https://mokpolar.tistory.com/60)
- [Running systemd in a non-privileged container](https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container#)
- [The New Control Group Interfaces](https://systemd.io/CONTROL_GROUP_INTERFACE/)
