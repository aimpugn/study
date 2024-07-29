# php-src

## Dockerfile

php-src 를 빌드하기 위한 Dockerfile 입니다.
php-src 를 클론하고 해당 디렉토리에서 빌드합니다.

```Dockerfile
FROM ubuntu:18.04

RUN apt update && apt install -y \
    software-properties-common \
    git \
    wget \
    pkg-config \
    build-essential \
    autoconf \
    flex \
    libxml2-dev \
    libsqlite3-dev \
    libtool \
    openssl \
    libcurl4-openssl-dev \
    libgd-dev \
    vim

WORKDIR /temp

# Bison 2.7 설치
RUN wget https://ftp.gnu.org/gnu/bison/bison-2.7.tar.gz && tar -xzf bison-2.7.tar.gz
RUN (cd bison-2.7 && ./configure && make && make install)

# Re2c 설치
RUN wget https://sourceforge.net/projects/re2c/files/1.0.1/re2c-1.0.1.tar.gz && tar -xzf re2c-1.0.1.tar.gz
RUN (cd re2c-1.0.1 && ./configure && make && make install)

# Automake 최신 버전 설치
RUN wget https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz && tar -xzf automake-1.16.tar.gz
RUN (cd automake-1.16 && ./configure && make && make install)

# PPA 추가 및 특정 버전의 GCC 설치
RUN add-apt-repository ppa:ubuntu-toolchain-r/test
RUN apt update && apt install -y g++-5

WORKDIR /php-src

# build 시에 실행해도 되고, 도커 안에서 실행해도 된다
# RUN ./configure --disable-all && make

COPY . .

CMD ["bash"]
```
