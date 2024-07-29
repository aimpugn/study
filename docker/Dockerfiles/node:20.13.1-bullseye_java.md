# node:20.13.1-bullseye & java

```Dockerfile
FROM node:20.13.1-bullseye

# 패키지 목록 업데이트 및 필수 패키지 설치
RUN apt-get update && apt-get install -y \
    # packages/v1-legacy-sdk/closure-compiler-v20180101.jar 실행 위해 jdk 설치
    openjdk-17-jdk \
    && apt-get clean

# install pnpm tsx
RUN npm install -g pnpm tsx

WORKDIR /usr/src/app
# 컨테이너 내에서 의존성 설치되는 디렉토리 생성
RUN mkdir /usr/src/app/node_modules/
# /usr/src/app 하위 파일 및 디렉토리의 소유권을 node:node로 변경
RUN chown -R node:node /usr/src/app

# 기본 사용자 변경
USER node

WORKDIR /usr/src/app/examples/standard


# Corepack이 패키지 매니저 버전을 엄격하게 강제하지 않도록 하는 설정.
# ERR_PNPM_BAD_PM_VERSION 에러로 인한 실패를 방지합니다.
ENV COREPACK_ENABLE_STRICT=0

CMD pnpm install && pnpm start:local --host
```
