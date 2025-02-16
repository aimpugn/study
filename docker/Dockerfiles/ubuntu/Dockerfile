FROM ubuntu:latest

USER root

# 업데이트
RUN apt-get update && apt-get install -y wget curl git zsh sudo locales
# 한글 입력시 `<ffffffff><ffffffff>...` 입력되는 이슈 해결
# https://beomi.github.io/2017/07/10/Ubuntu-Locale-to-ko_KR/
RUN locale-gen --no-archive ko_KR.UTF-8

# Docker는 로그인 과정 없이 바로 실행되므로, 일반적인 로그인 세션처럼 동작하지 않습니다.
# 따라서 로케일 정보를 명시적으로 설정합니다.
ENV LANG=ko_KR.UTF-8 LC_ALL=ko_KR.UTF-8

# ubuntu 사용자를 sudo 그룹에 추가합니다.
RUN usermod -aG sudo ubuntu \
    # sudo 실행 시 패스워드를 묻지 않도록 설정합니다.
    && echo "ubuntu ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/ubuntu

ENV USER=ubuntu
ENV HOME="/home/${USER}"

USER ${USER}
WORKDIR ${HOME}

RUN rm -rf ~/.oh-my-zsh

RUN yes | sh -c "$(wget https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)"

# 컨테이너가 시작될 때 zsh 사용
CMD ["/bin/zsh"]
