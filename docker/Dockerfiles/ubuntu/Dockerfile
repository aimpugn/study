FROM ubuntu:latest

USER root

# `RUN` 명령어마다 레이어를 생성하며, 이전 레이어를 캐싱합니다.
# - 명령어가 변경되지 않았고, 이전 레이어가 존재하면 캐시를 재사용합니다.
# - 명령어가 변경되면 해당 명령어 명령어가 변경되면 해당 명령어 "이후"의 모든 레이어는 다시 빌드됩니다.의 모든 레이어는 다시 빌드됩니다.
# 업데이트
RUN apt-get update

# Ubuntu의 최소화된 이미지를 다시 복구합니다.
# - [unminimize](https://git.launchpad.net/unminimize/tree/unminimize)
RUN yes | unminimize

# 사용할 패키지들을 설치합니다.
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y \
    # - [`init`](https://wiki.debian.org/Init): 시스템이 부팅된 후 처음 실행되는 프로그램
    # - `systemd`:  systemctl을 통해 서비스 관리를 수행
    # - `systemd-sysv`: 과거 `/etc/init.d/` 스크립트 기반 부팅 시스템 호환 위해 필요
    init systemd systemd-sysv sudo locales \
    man-db glibc-doc \
    # 컴파일 및 Debian/Ubuntu 패키지 빌드에 필수적인 도구 모음(gcc, make, libc6-dev 등)
    build-essential \
    wget curl git zsh vim \
    # netstat, ifconfig
    # - https://github.com/ecki/net-tools
    net-tools \
    # - https://launchpad.net/ubuntu/+source/iputils
    iputils-ping \
    # 다른 패키지의 의존성으로 자동 설치되었으나, 현재는 필요 없는 패키지 제거
    && apt-get autoremove

# 한글 입력시 `<ffffffff><ffffffff>...` 입력되는 이슈 해결
# - https://beomi.github.io/2017/07/10/Ubuntu-Locale-to-ko_KR/
RUN locale-gen --no-archive ko_KR.UTF-8

# Docker는 로그인 과정 없이 바로 실행되므로, 일반적인 로그인 세션처럼 동작하지 않습니다.
# 따라서 로케일 정보를 명시적으로 설정합니다.
ENV LANG=ko_KR.UTF-8 LC_ALL=ko_KR.UTF-8

# Ubuntu 유저 설정 시작
ENV USER=ubuntu
ENV HOME="/home/${USER}"

# ubuntu 사용자를 sudo 그룹에 추가하고, sudo 실행 시 패스워드를 묻지 않도록 설정합니다.
RUN usermod -aG sudo ${USER} \
    && echo "${USER} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/ubuntu

USER ${USER}
WORKDIR ${HOME}

RUN rm -rf ~/.oh-my-zsh

RUN yes | sh -c "$(wget https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)"

# Ubuntu 유저 설정 끝

# systemd 실행 위해 root 유저로 변경
USER root

# `systemd`가 컨테이너에서 PID 1 프로세스로 동작 중이라면, 컨테이너 종료시 올바른 종료 신호가 필요합니다. 그런데 `systemd`는 `SIGTERM`이나 `SIGKILL`로 제대로 종료되지 않을 수 있습니다.
# - Systemd does not exit on sigterm.  Systemd defines that shutdown signal as `SIGRTMIN+3`, docker upstream should send this signal when user does a docker stop.
# - `SIGKILL`(`kill -9`): `systemd`는 즉시 종료되지만, 모든 하위 서비스는 제대로 정리되지 않고 그대로 남아 있을 수 있습니다.
#
# `SIGRTMIN+3`신호를 받으면 `systemd`는 내부적으로 `systemctl stop`과 유사한 과정을 거쳐 모든 데몬 및 서비스를 순차적으로 종료함으로써 데이터 손실이나 불안정성을 최소화할 수 있습니다.
#
# References:
# - https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container#
# - https://raby.sh/sigterm-and-pid-1-why-does-a-container-linger-after-receiving-a-sigterm.html
STOPSIGNAL SIGRTMIN+3

COPY docker-entrypoint.sh /usr/local/bin/

# 컨테이너가 시작될 때 항상 실행되어야 하는 명령을 정의합니다.
ENTRYPOINT [ "docker-entrypoint.sh" ]

# systemd를 PID 1로 실행합니다.
# init 시스템이 PID 1로 실행되지 않는 경우 다음과 같은 문제들이 있을 수 있습니다.
# - 1번 프로세스가 `SIGCHLD`를 몰라서 컨테이너 내에 무수히 많은 좀비 프로세스가 생성될 수 있고
# - 보통 `/dev/syslog`를 리스닝하는 프로세스가 존재하지 않아서 어떤 프로세스가 `syslog`에 로그를 남길 경우 해당 로그들은 사라지고
# - 자체적인 Init 커맨드를 빌드해야 합니다.
# `systemd`를 사용하면 서비스 실행 위한 표준 리눅스 API를 제공합니다.
CMD ["/lib/systemd/systemd"]
