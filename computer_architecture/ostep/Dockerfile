# multi pltform image:
# https://docs.docker.com/build/building/multi-platform/
# FROM --platform=linux/amd64 ubuntu:22.04
FROM ubuntu:24.04

USER root

# default settings
RUN apt update -y \
    # to prevent `debconf: (No usable dialog-like program is installed` prompt
    && apt-get install whiptail -y

RUN apt-get upgrade -y
RUN apt-get install -y \
    curl \
    wget \
    zsh \
    fzf \
    git \
    build-essential

# Add user and group
ENV USER=ostep
ENV GROUP=ostep
ENV HOME_DIR=/home/$USER
ENV WORK_DIR=$HOME_DIR/ostep

RUN groupadd $GROUP
RUN useradd $USER \
    # create home
    -m \
    # specify directory
    -d /home/$USER \
    -p 12345 \
    -g $GROUP

# If want to set the user as sudoer
# RUN apt-get install sudo -y
# RUN usermod \
#     # append
#     -a \
#     # append the user to the supplemental GROUPS
#     # without removing the user from other groups
#     -G sudo $USER

# switch to the USER
USER $USER
WORKDIR $HOME_DIR

# install rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs -o rustup-init.sh \
    && chmod +x ./rustup-init.sh
RUN ./rustup-init.sh -y && rm -f ./rustup-init.sh

# 컨테이너가 시작될 때 실행되는 기본 명령을 지정
# `bash` 셸을 시작하여 컨테이너가 셸 세션으로 시작되도록 설정
# 사용자가 컨테이너 내에서 셸을 통해 직접 명령을 입력하고 싶을 때 유용하다
CMD ["bash"]