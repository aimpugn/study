# docker-compose

- [docker-compose](#docker-compose)
    - [실행](#실행)
        - [이미지 빌드하고 컨테이너 생성만](#이미지-빌드하고-컨테이너-생성만)
        - [백그라운드에서 컨테이너 실행](#백그라운드에서-컨테이너-실행)
    - [`docker-compose.yaml` 구조](#docker-composeyaml-구조)
    - [`services`](#services)
        - [`build`](#build)
            - [속성 옵션](#속성-옵션)
        - [`depends_on`](#depends_on)
        - [`healthcheck`](#healthcheck)
    - [`networks`](#networks)
    - [`volumes`](#volumes)
    - [`configs`](#configs)
    - [`secrets`](#secrets)
    - [응용](#응용)
        - [`healthcheck`와 `depends_on`으로 다른 서비스 기다리기](#healthcheck와-depends_on으로-다른-서비스-기다리기)

## 실행

### 이미지 빌드하고 컨테이너 생성만

```shell
docker-compose up --no-start
```

### 백그라운드에서 컨테이너 실행

```shell
docker-compose up -d
```

## `docker-compose.yaml` 구조

```yaml
# Docker Compose 파일의 버전. 예: version: '3' 또는 version: '3.8'
# 버전에 따라 사용할 수 있는 키와 기능이 다르다
version: '3'

name: $SOME_NAME

# 실행할 컨테이너와 그 설정을 정의한다.
# 이 섹션에서 각 서비스는 하나 이상의 컨테이너를 가질 수 있다
services:
  web:
    image: nginx:alpine
    # restart:
    #  컨테이너가 중지될 때마다 자동으로 다시 시작하도록 설정하는 옵션
    #  컨테이너가 예기치 않게 종료되더라도 자동으로 다시 시작한다.
    restart: always # `no`, `on-failure:5`, `always`, `unless-stopped`
    deploy:
      replicas: 3
      update_config:
        parallelism: 2
    ports:
      - "80:80"
    networks:
      - my_network
  db:
    image: postgres:latest
    environment:
      POSTGRES_DB: my_db
    networks:
      - my_network
    volumes:
      - my_volume:/var/lib/postgresql/data

# 사용자 정의 네트워크를 정의한다. 네트워크는 컨테이너 간의 통신을 가능하게 한다
# 컨테이너 간의 통신 규칙을 설정할 수 있다
networks:
  my_network:
    driver: bridge

# 데이터를 저장할 볼륨을 정의한다
# 이는 컨테이너가 삭제되더라도 데이터를 유지할 수 있게 해준다
volumes:
  my_volume:
    driver: local

# 설정 파일을 정의하며, 이는 서비스에 사용될 수 있다
configs:
    my_config:
        file: ./my_config.txt

# 비밀을 정의한다. 비밀은 민감한 데이터를 보호하는 데 사용된다
secrets:
  my_secret:
    file: ./my_secret.txt

# 배포 옵션을 정의한다. 이는 스웜 모드에서 사용된다
# deploy:
```

- `restart` 상세 설명

    1. **`no`**:
       - 컨테이너가 자동으로 다시 시작되지 않습니다. 기본값입니다.
       - 예시:

         ```yml
         restart: no
         ```

    2. **`on-failure[:max-retries]`**:
       - 컨테이너가 비정상 종료(즉, 종료 코드가 0이 아닌 경우)될 때만 다시 시작합니다.
       - 특정 오류 상황에서만 재시작이 필요할 때 유용합니다.
       - `max-retries`를 지정하면 최대 재시도 횟수를 설정할 수 있습니다.
       - 예시:

         ```yml
         restart: on-failure
         ```

         또는

         ```yml
         restart: on-failure:5
         ```

    3. **`always`**:
       - 컨테이너가 중지될 때마다 항상 다시 시작합니다. 수동으로 중지한 경우에도 Docker 데몬이 다시 시작되면 컨테이너가 다시 시작됩니다.
       - 단, 컨테이너가 시작 후 10초 이내에 종료되면, Docker는 무한 재시작 루프를 방지하기 위해 재시작을 시도하지 않습니다
       - 예시:

         ```yml
         restart: always
         ```

    4. **`unless-stopped`**:
       - `always`와 유사하지만, 수동으로 중지한 경우 Docker 데몬이 다시 시작되더라도 컨테이너가 다시 시작되지 않습니다.
       - 수동으로 중지한 컨테이너를 다시 시작하지 않으려는 경우 유용합니다.
       - 예시:

         ```yml
         restart: unless-stopped
         ```

## [`services`](https://docs.docker.com/compose/compose-file/05-services/)

- `image`
    - 사용할 Docker 이미지를 지정
- `build`
    - 이미지를 로컬에서 빌드할 경로 또는 설정을 지정
- `ports`
    - 포트 매핑을 정의
    - 호스트와 컨테이너 간의 포트를 매핑
- `environment`
    - 환경 변수를 설정
- `networks`
    - 이 서비스가 사용할 네트워크를 지정
- `volumes`
    - 볼륨을 마운트

### `build`

`docker-compose.yml` 파일에서 `build` 속성은 Docker 이미지를 빌드할 때 사용됩니다.
이 속성을 사용하여 빌드 컨텍스트, Dockerfile 경로, 빌드 argument 등을 지정할 수 있습니다.

`build` 속성은 Docker 이미지를 빌드할 때 사용할 빌드 컨텍스트, Dockerfile, 빌드 아르기먼트 등을 지정하는 데 사용됩니다.

#### 속성 옵션

```yaml
version: '3.8'

services:
    app:
        build:
        # Docker 빌드 컨텍스트를 지정합니다.
        # 이는 Docker 빌드가 실행되는 디렉토리를 의미합니다.
        context: ./sub-directory
        # 사용할 Dockerfile의 경로를 지정합니다.
        # 기본값은 빌드 컨텍스트의 루트에 있는 `Dockerfile`입니다.
        dockerfile: Dockerfile
        # Docker 빌드 중 사용할 빌드 아르기먼트를 지정합니다.
        # 이는 Dockerfile 내에서 `ARG` 지시어를 통해 접근할 수 있습니다.
        args:
            build_no: 1
        # 이전에 빌드한 이미지를 캐시 소스로 지정하여 빌드 시간을 단축할 수 있습니다.
        cache_from:
            - my-app:cache
        # 다중 스테이지 빌드에서 특정 빌드 스테이지를 지정합니다.
        target: build-stage
        # 빌드된 이미지에 레이블을 추가합니다.
        labels:
            com.example.description: "This is an example app"
        # 빌드할 때 사용할 네트워크 모드를 지정합니다.
        network: host
        # `/dev/shm` 디렉토리의 크기를 지정합니다.
        shm_size: '256m'
        # 빌드 시 사용할 격리 기술을 지정합니다.
        isolation: hyperv
        # 빌드할 때 사용할 추가 호스트 엔트리를 지정합니다.
        extra_hosts:
            - "somehost:162.242.195.82"
            - "otherhost:50.31.209.229"
        image: my-app:latest
        ports:
        - "80:80"
```

### `depends_on`

### [`healthcheck`](https://docs.docker.com/compose/compose-file/05-services/#healthcheck)

## [`networks`](https://docs.docker.com/compose/compose-file/06-networks/)

- `driver`
    - 네트워크 드라이버를 지정
    - 기본값은 `bridge`
- `ipam`
    - IP 주소 관리 설정을 할 수 있다

## [`volumes`](https://docs.docker.com/reference/compose-file/volumes/)

볼륨은 컨테이너 엔진에 의해 구현된 영구 데이터 저장소입니다.
Compose는 서비스가 볼륨을 마운트하는 중립적인 방법과 인프라에 볼륨을 할당하는 구성 매개변수를 제공합니다.
최상위 볼륨 선언을 사용하면 여러 서비스에서 재사용할 수 있는 네임드 볼륨을 구성할 수 있습니다.

- `driver`
    - 볼륨 드라이버를 지정
    - 기본값은 `local`

## [`configs`](https://docs.docker.com/compose/compose-file/08-configs/)

## [`secrets`](https://docs.docker.com/compose/compose-file/09-secrets/)

## 응용

### `healthcheck`와 `depends_on`으로 다른 서비스 기다리기

Docker Compose에서 `command`나 `entrypoint`가 실행되는 동안 기다리도록 하는 방법을 찾고 있다면,
보통 `depends_on` 옵션과 `healthcheck`를 사용하여 서비스의 상태를 확인하고, 모든 서비스가 준비될 때까지 기다리게 할 수 있습니다.

하지만, `depends_on`은 기본적으로 컨테이너 시작 순서를 보장하지만, 실제로 서비스가 준비되었는지 여부는 보장하지 않습니다.

이를 해결하려면 `healthcheck`를 설정하여 서비스가 준비된 상태를 확인하고, `depends_on`에 `condition: service_healthy`를 사용하여 종속성을 설정할 수 있습니다.

아래는 `pnpm` 설치가 완료될 때까지 기다린 후 다른 서비스를 시작하도록 설정하는 예제입니다.

```yaml
version: '3.8'

services:
  app:
    build:
      context: ../browser-sdk/examples/standard
      dockerfile: Dockerfile
    restart: always
    working_dir: /usr/src/app/examples/standard
    command: /bin/sh -c "pnpm install && pnpm start:local"
    volumes:
      - ../browser-sdk:/usr/src/app
      - /usr/src/app/node_modules
    networks:
      - backend
    healthcheck:
      test: ["CMD-SHELL", "pnpm --version || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  another_service:
    image: some_image
    depends_on:
      app:
        condition: service_healthy
    networks:
      - backend

networks:
  backend:
```

1. **`app` 서비스**:

    `command`로 `pnpm install && pnpm start:local`을 실행합니다.

    `healthcheck`를 설정하여 `pnpm --version` 명령이 성공하는지 확인합니다.
    이 명령이 성공하면 `pnpm`이 설치된 것으로 간주합니다.

    `healthcheck`의 `interval`, `timeout`, `retries`, `start_period` 옵션을 사용하여 주기적으로 상태를 확인합니다.

2. **`another_service` 서비스**:

    `depends_on` 옵션을 사용하여 `app` 서비스의 상태를 확인합니다.
    `condition: service_healthy`를 사용하여 `app` 서비스가 건강한 상태(healthcheck 통과)일 때까지 기다립니다.

Docker Compose 파일을 작성한 후 다음 명령을 사용하여 서비스를 실행합니다:

```sh
docker-compose -p core up -d --build
```

이 명령은 다음을 수행합니다:
- Docker Compose 프로젝트 이름을 `core`로 설정하여 서비스를 빌드하고 백그라운드에서 실행합니다.
- `app` 서비스가 `pnpm install` 명령을 실행하고, `healthcheck`를 통과할 때까지 기다립니다.
- `another_service`는 `app` 서비스가 준비된 후에 시작됩니다.

이 설정을 통해 `pnpm` 설치가 완료될 때까지 기다린 후 다른 서비스가 시작되도록 할 수 있습니다.
