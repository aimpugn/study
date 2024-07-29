# docker build

- [docker build](#docker-build)
    - [docker build 실행 경로](#docker-build-실행-경로)
    - [예시](#예시)
    - [Options](#options)
        - [`-t`, `--tag` stringArray](#-t---tag-stringarray)
        - [`--platform` 옵션](#--platform-옵션)
            - [x86 (32비트) 아키텍처로 빌드하기](#x86-32비트-아키텍처로-빌드하기)
            - [x64 (64비트) 아키텍처로 빌드하기](#x64-64비트-아키텍처로-빌드하기)
            - [Docker 명령어에서 아키텍처 지정하기](#docker-명령어에서-아키텍처-지정하기)
    - [ETC](#etc)
        - [mysql in github actions](#mysql-in-github-actions)

## docker build 실행 경로

Docker 빌드 컨텍스트는 `docker build` 명령어의 마지막 인자로 지정된 디렉토리이며, 이 컨텍스트 내에서만 파일을 참조할 수 있습니다.

부모 디렉토리에서 빌드를 실행하면 빌드 컨텍스트가 달라져서 `COPY` 명령어가 실패하게 됩니다.
이를 해결하려면 빌드 컨텍스트를 올바르게 설정해야 합니다.
상대 경로에서 실행하면 안되고, Dockerfile이 있는 곳에서 실행해야 한다.

다음과 같은 디렉토리 구조가 있다고 가정해봅시다.

```bash
❯ tree . -d                       
.
├── subdir
│   ├── docker-commands/
│   ├── plugins/
│   └── Dockerfile
...생략...
```

```dockerfile
FROM debian:buster-slim

... 생략 ...

COPY plugins /workspace/plugins
COPY docker-commands /workspace/bin

WORKDIR /workspace
ENTRYPOINT ["/workspace/bin/entrypoint.sh"]
```

이때 만약 부모 디렉토리에서 `subdir` 디렉토리 내에 있는 `Dockerfile`을 빌드하려고 할 때,
부모 디렉토리에서 다음과 같이 명령어를 실행하면 아래와 같은 에러가 발생합니다.

```bash
docker build -f subdir/Dockerfile .
```

```bash
 => ERROR [ 9/11] COPY plugins /workspace/plugins                                                                                                                                                                             0.0s
 => ERROR [10/11] COPY docker-commands /workspace/bin                                                                                                                                                                         0.0s
------
 > [ 9/11] COPY plugins /workspace/plugins:
------
------
 > [10/11] COPY docker-commands /workspace/bin:
------
Dockerfile:48
--------------------
  46 |     
  47 |     COPY plugins /workspace/plugins
  48 | >>> COPY docker-commands /workspace/bin
  49 |     
  50 |     WORKDIR /workspace
--------------------
```

이는 *Docker 빌드 컨텍스트*와 관련된 문제 때문에 발생합니다.
Docker 빌드를 실행할 때, Docker는 빌드 컨텍스트를 설정하고, 이 컨텍스트 내에서만 파일을 참조할 수 있습니다.
빌드 컨텍스트는 `docker build` 명령어의 마지막 인자로 지정된 디렉토리입니다.
이 디렉토리의 내용이 Docker 데몬으로 전송되어 빌드가 진행됩니다.

빌드 컨텍스트는 현재 디렉토리(`.`)가 됩니다.
Docker는 현재 디렉토리의 내용을 빌드 컨텍스트로 사용하여 Docker 데몬으로 전송합니다.
그러나 `Dockerfile`은 `subdir` 디렉토리 내에 있으므로, `COPY` 명령어가 상대 경로를 사용하여 `plugins`와 `docker-commands` 디렉토리를 참조하려고 할 때 문제가 발생합니다.

`COPY` 명령어는 빌드 컨텍스트 내에서 파일을 찾기 때문에, `subdir` 디렉토리 내의 파일을 참조할 수 없습니다.
 따라서 `COPY` 명령어가 실패하게 됩니다.

이 문제를 해결하려면, 빌드 컨텍스트를 `subdir` 디렉토리로 설정해야 합니다. 즉, `docker build` 명령어를 `subdir` 디렉토리에서 실행하거나, 빌드 컨텍스트를 명시적으로 지정해야 합니다:

```bash
cd subdir
docker build -t myimage .
```

또는

```bash
docker build -t myimage -f subdir/Dockerfile subdir
```

이렇게 하면 빌드 컨텍스트가 `subdir` 디렉토리가 되어, `COPY` 명령어가 올바르게 작동하게 됩니다.

## 예시

```bash
docker build -t temp-builder -f "$CURRENT_DIR/Dockerfile" .
docker run \
  --rm \
  -v "$CURRENT_DIR/../../$INTERFACE/protobuf":/workspace/interface:ro \
  -v "$CURRENT_DIR/../../$CORE":/workspace/core:rw \
  temp-builder \
  "$SERVICE" \
  "$PHP_TARGET_PATH"  \
  ;
```

## Options

### `-t`, `--tag` stringArray

> Name and optionally a tag (format: "name:tag")

### `--platform` 옵션

Docker에서 특정 아키텍처(x86, x64)로 이미지를 빌드하려면, `--platform` 플래그를 사용하는 것이 일반적인 방법입니다.
이 플래그는 `docker build` 명령어에 사용되며, 빌드하려는 타깃 아키텍처를 지정하는 데 사용됩니다.

`--platform` 플래그는 Docker의 실험적 기능일 수 있으므로, 사용하려면 Docker의 버전이 최신인지 확인해야 합니다.

일부 시스템에서는 특정 아키텍처를 에뮬레이션하는 데 추가 설정이 필요할 수 있습니다.
예를 들어, ARM 기반 시스템에서 x86 이미지를 빌드하려면 QEMU 등의 에뮬레이터가 필요할 수 있습니다.

모든 베이스 이미지가 모든 아키텍처를 지원하는 것은 아니므로,
사용하려는 베이스 이미지가 원하는 아키텍처를 지원하는지 확인해야 합니다..

#### x86 (32비트) 아키텍처로 빌드하기

x86 아키텍처는 32비트 시스템을 나타냅니다. 하지만 대부분의 최신 시스템과 Docker 이미지는 64비트(x64) 기반으로 되어 있으므로, 32비트 이미지를 찾기 어려울 수 있습니다. 일반적으로 `linux/386` 또는 `linux/i386` 플랫폼을 지정하여 32비트 이미지를 빌드할 수 있습니다.

```dockerfile
# 예시
FROM --platform=linux/386 ubuntu:20.04

# 이후 Dockerfile의 내용
```

#### x64 (64비트) 아키텍처로 빌드하기

x64 아키텍처는 64비트 시스템을 나타냅니다. 대부분의 현대적인 시스템과 Docker 이미지는 기본적으로 64비트 기반이므로, 이는 가장 일반적인 케이스입니다. `linux/amd64` 플랫폼을 지정하여 64비트 이미지를 빌드할 수 있습니다.

```dockerfile
# 예시
FROM --platform=linux/amd64 ubuntu:20.04

# 이후 Dockerfile의 내용
```

#### Docker 명령어에서 아키텍처 지정하기

`docker build` 명령어를 사용할 때도 `--platform` 플래그를 사용하여 특정 아키텍처를 강제할 수 있습니다.

```bash
docker build --platform=linux/amd64 -t myimage:latest .
```

이 명령은 현재 디렉토리의 Dockerfile을 사용하여 `myimage:latest`라는 태그의 이미지를 `linux/amd64` 플랫폼으로 빌드합니다.

## ETC

### mysql in github actions

```shell
/usr/bin/docker create \
    --name 085b8ed81ae246b88a0e98941c2a74e8_mysql5559_8bb4a7 \
    --label 6f1554 \
    --network github_network_ad434fcff3894859b5667212834694fc \
    --network-alias mysql \
    -p ${CUSTOM_PORT}:3306 \
    --health-cmd="mysqladmin ping" \
    --health-interval=10s  \
    --health-timeout=5s \
    --health-retries=3 \
        -e "MYSQL_DATABASE=DBNAME" \
        -e "MYSQL_HOST=127.0.0.1" \
        -e "MYSQL_USER=${DB_USER}" \
        -e "MYSQL_PASSWORD=${DB_PASSWORD}" \
        -e "MYSQL_ROOT_PASSWORD=${DB_ROOT_PASSWORD}" \
        -e GITHUB_ACTIONS=true \
        -e CI=true mysql:${VERSION}
```

실행 결과fㅗ `container_id`가 출력된다

```shell
1df07d1efe02be0c726c0c78b5009851853a17f06ef9be2073005e1394002a7b
```
