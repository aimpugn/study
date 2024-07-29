
# Ubuntu 22.04 & node & java

```Dockerfile
# Ubuntu 22.04 이미지를 기반으로 합니다.
FROM ubuntu:22.04

# 패키지 목록을 업데이트하고, `curl`, `gnupg`, `openjdk-17-jdk`를 설치합니다. 설치 후 불필요한 캐시 파일을 제거합니다.
RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    openjdk-17-jdk \
    && apt-get clean

# Node.js 및 npm 설치 (NodeSource를 통해 설치)
# NodeSource의 설치 스크립트를 사용하여 Node.js 20.x 및 npm을 설치합니다.
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean

# pnpm 설치
RUN npm install -g pnpm

# Java 환경 변수 설정
# Java 실행 파일을 포함한 경로를 설정하여, Java 명령을 사용할 수 있도록 환경 변수를 설정합니다.
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH=$JAVA_HOME/bin:$PATH

# 작업 디렉토리 설정
WORKDIR /usr/src/app

# 의존성 파일 복사 및 설치
# `package.json`과 `pnpm-lock.yaml` 파일을 복사하고, pnpm을 사용하여 의존성을 설치합니다.
COPY package.json pnpm-lock.yaml ./
RUN pnpm install

# 애플리케이션 소스 복사
COPY . .

# 파일 시스템 권한 설정
RUN chown -R node:node /usr/src/app

# 기본 사용자 변경
USER node

# 기본 명령 설정
CMD ["tail", "-f", "/dev/null"]
```
