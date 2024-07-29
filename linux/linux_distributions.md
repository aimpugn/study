# linux distributions

## 리눅스 배포판이란?

리눅스 배포판은 리눅스 커널을 기반으로 하여 다양한 응용 프로그램, 데스크탑 환경, 패키지 관리 시스템 등을 포함하는 운영 체제입니다.
각 배포판은 특정 용도나 사용자 그룹을 타겟으로 하여 최적화되어 있으며, 다양한 하드웨어와 소프트웨어 요구 사항을 만족시키기 위해 개발되었습니다.

## Docker 이미지의 리눅스 배포판

### Node.js Docker 이미지 태그

각각의 Node.js Docker 이미지 태그는 기반이 되는 리눅스 배포판과 그 버전에 따라 차이가 있습니다.

1. **`20.13.1`**:
    - **기반**: Debian
    - **특징**: 이 이미지는 특정 Debian 버전을 명시하지 않으므로, 기본적으로 최신 안정 버전의 Debian을 사용합니다. 보통은 최신 LTS 버전이 사용됩니다.
    - **용도**: 일반적인 리눅스 환경과 유사하며, 다양한 패키지를 쉽게 설치할 수 있습니다. 안정적이며, 보안 업데이트도 적시에 제공됩니다.

2. **`20.13.1-bullseye`**:
    - **기반**: Debian 11 (Bullseye)
    - **특징**: 이 이미지는 Debian 11을 기반으로 합니다. Debian 11은 2021년 8월에 릴리스된 안정 버전으로, 최신 안정 패키지와 보안 업데이트를 제공합니다.
    - **용도**: 최신 안정 버전의 Debian을 필요로 하는 환경에 적합합니다. Debian 11의 모든 최신 기능과 업데이트를 포함합니다.

3. **`20.13.1-bookworm`**:
    - **기반**: Debian 12 (Bookworm)
    - **특징**: 이 이미지는 Debian 12를 기반으로 합니다. Debian 12는 2023년에 릴리스된 최신 안정 버전으로, 더 최신의 패키지와 기능을 제공합니다.
    - **용도**: 가장 최신의 기능과 패키지를 필요로 하는 환경에 적합합니다. 최신 보안 업데이트와 기능을 포함하여 더욱 진보된 환경을 제공합니다.

4. **`20.13.1-buster`**:
    - **기반**: Debian 10 (Buster)
    - **특징**: 이전의 안정된 Debian 릴리스로, Bullseye보다 오래된 패키지 버전을 포함하고 있습니다.
    - **용도**: 오래된 버전으로, 보안 업데이트가 곧 종료될 예정입니다. 새로운 프로젝트에는 권장되지 않습니다.

5. **`20.13.1-alpine`**:
    - **기반**: Alpine Linux
    - **특징**: 매우 경량이며, 빠른 부팅과 높은 보안성을 제공하지만, glibc 대신 musl libc를 사용하여 glibc 기반 소프트웨어와 호환성 문제가 있을 수 있습니다.
    - **용도**: 경량의 컨테이너 이미지를 필요로 하는 경우에 적합합니다.

6. **`20.13.1-alpine3.20`**:
    - **기반**: Alpine Linux 3.20
    - **특징**: Alpine 3.20 릴리스를 사용하여 최신 패키지와 보안 업데이트를 포함합니다.
    - **용도**: 최신 기능과 패키지를 사용할 수 있는 경량 환경을 원하는 사용자에게 적합합니다.

- 세부적인 차이점:

    1. **패키지 버전**:
        - **Bullseye**: 안정적인 패키지를 제공하며, 최신 패키지보다는 검증된 버전을 포함합니다.
        - **Bookworm**: 최신 패키지를 제공하여 최신 기능과 개선 사항을 사용할 수 있습니다.

    2. **보안 업데이트**:
        - **Bullseye**: 보안 업데이트가 활발히 이루어지며, 안정성과 보안을 중시하는 사용자에게 적합합니다.
        - **Bookworm**: 최신 보안 업데이트를 제공하며, 더 최신의 보안 기능을 포함합니다.

    3. **하드웨어 지원**:
        - **Bullseye**: 다양한 하드웨어를 안정적으로 지원합니다.
        - **Bookworm**: 최신 하드웨어 지원을 포함하여 더 많은 기능과 최적화를 제공합니다.

- **`20.13.1`**: 특정 Debian 버전을 명시하지 않으며, 최신 안정 버전의 Debian을 사용합니다.
- **`20.13.1-bullseye`**: Debian 11 (Bullseye)을 기반으로 하여 안정성과 보안성을 중시하는 환경에 적합합니다.
- **`20.13.1-bookworm`**: Debian 12 (Bookworm)을 기반으로 하여 최신 기능과 패키지를 필요로 하는 환경에 적합합니다.

### 추천 태그

일반적인 리눅스 환경에서 JDK나 curl을 설치하여 사용하려면 Debian 기반 이미지를 사용하는 것이 좋습니다. 특히, 최신 안정 버전인 `bookworm`을 사용하는 것이 좋습니다.

### 최적화된 Dockerfile 예시

다음은 `20.13.1-bookworm` 태그를 사용하여 JDK와 curl을 설치하는 최적화된 Dockerfile 예시입니다:

```Dockerfile
# Node.js 20.13.1 이미지 사용 (Debian 12 Bookworm 기반)
FROM node:20.13.1-bookworm

# 패키지 목록 업데이트 및 필수 패키지 설치
RUN apt-get update && apt-get install -y \
    curl \
    openjdk-17-jdk \
    && apt-get clean

# pnpm 및 tsx 설치
RUN npm install -g pnpm tsx

# node 사용자 및 그룹 추가
RUN groupadd -g 1000 node && \
    useradd -u 1000 -g node -s /bin/bash node

WORKDIR /usr/src/app

# 컨테이너 내에서 의존성 설치되는 디렉토리 생성 및 소유권 변경
RUN mkdir /usr/src/app/node_modules/ && \
    chown -R node:node /usr/src/app

# 기본 사용자 변경
USER node

WORKDIR /usr/src/app/examples/standard

# pnpm 환경 변수 설정
ENV COREPACK_ENABLE_STRICT=0

# 애플리케이션 실행
CMD ["sh", "-c", "pnpm install && pnpm start:local --host"]
```

이 Dockerfile은 Debian 12 (Bookworm) 기반의 Node.js 이미지를 사용하여 JDK와 curl을 설치하고, 필요한 설정을 추가합니다. 이를 통해 일반적인 리눅스 환경에서 필요한 패키지를 설치하고 사용할 수 있습니다.
