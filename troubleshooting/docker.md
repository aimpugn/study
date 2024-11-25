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
