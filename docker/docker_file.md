# Dockerfile

- [Dockerfile](#dockerfile)
    - [`EXPOSE`](#expose)
    - [`HEALTHCHECK`](#healthcheck)
    - [`CMD`](#cmd)
        - [사용 예시](#사용-예시)
    - [`ENTRYPOINT`](#entrypoint)
        - [사용 예시](#사용-예시-1)
    - [`CMD`와 `ENTRYPOINT`의 차이](#cmd와-entrypoint의-차이)
    - [컨테이너가 바로 종료되는 이유](#컨테이너가-바로-종료되는-이유)
        - [방법 1: 대화형 셸 유지](#방법-1-대화형-셸-유지)
        - [방법 2: `tail -f /dev/null` 사용](#방법-2-tail--f-devnull-사용)
        - [컨테이너 실행 및 bash 셸 접속](#컨테이너-실행-및-bash-셸-접속)
        - [요약](#요약)

## [`EXPOSE`](https://docs.docker.com/reference/dockerfile/#expose)

`EXPOSE` 명령어는 Dockerfile에서 *컨테이너가 수신할 수 있는 네트워크 포트*를 지정합니다.

그러나 `EXPOSE` 명령어 자체는 *포트를 실제로 노출하거나 외부에서 접근 가능하게 만들지는 않습니다*.

대신, 이는 이미지를 빌드하는 사람과 컨테이너를 실행하는 사람 사이에 어떤 포트를 게시할 것인지에 대한 일종의 문서 역할을 합니다.

- **문서화**:

  `EXPOSE` 명령어는 Dockerfile을 읽는 사람에게 이 컨테이너가 어떤 포트를 사용하고 있는지 알려줍니다.

  이는 컨테이너를 사용하는 개발자나 운영자가 컨테이너의 네트워크 설정을 이해하는 데 도움이 됩니다.
- **네트워크 설정 힌트**:

  Docker는 `EXPOSE`된 포트를 기반으로 컨테이너 간의 네트워크 통신을 설정할 때 힌트를 얻습니다.

  예를 들어, Docker Compose는 `EXPOSE`된 포트를 자동으로 인식하고 네트워크 설정을 구성할 수 있습니다.
- **실제 포트 노출**:

  `EXPOSE` 명령어는 컨테이너 내부에서 포트를 열지만, 외부에서 접근 가능하게 하려면:
    - `docker run` 명령어에서 `-p` 옵션을 사용
    - 또는 `docker-compose.yml` 파일에서 `ports` 옵션을 사용

```Dockerfile
... 생략 ...
# Expose the port that the service listens on
EXPOSE 8001
... 생략 ...
```

```yml
# docker-compose.yml
version: "3"

services:
  traefik:
    ... 생략 ...
    ports:
      - "8001:80"
      - "8002:8080"

  servicea:
    build:
      context: ./servicea
      dockerfile: Dockerfile
    restart: always
    working_dir: /usr/src/app
    volumes:
      - ./servicea/examples/standard:/usr/src/app
    command: ["pnpm", "start"]
    networks:
      - backend
    ports:
      - "8001:5432"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.servicea.rule=Host(`servicea.local`)"
      - "traefik.http.services.servicea.loadbalancer.server.port=5432"
```

servicea.local:8001

-> 호스트의 8001 포트

-> Traefik 컨테이너의 80 포트

-> Traefik 라우팅

-> servicea 컨테이너의 8001 포트

1. **클라이언트**

   클라이언트가 `servicea.local` 도메인으로 요청을 보냅니다.

   예를 들어, 브라우저에서 `http://servicea.local`로 접근합니다.
2. **Traefik**:

   Traefik은 리버스 프록시 역할을 하며, `servicea.local` 도메인으로 들어오는 요청을 처리합니다.

   Traefik은 `docker-compose.yml` 파일에서 정의된 포트 `8001`과 `8002`를 통해 외부와 통신합니다.

   `ports` 설정에 따라, Traefik은

   - 호스트의 `8001` 포트를 Traefik 컨테이너의 `80` 포트로 맵핑
   - 호스트의 `8002` 포트를 Traefik 컨테이너의 `8080` 포트로 맵핑
3. **Traefik 라우팅**:

   Traefik은 `servicea.local` 도메인으로 들어오는 요청을 `servicea` 서비스로 라우팅합니다.

   - `traefik.http.routers.servicea.rule=Host('servicea.local')` 라벨에 따라 Traefik은 해당 요청을 `servicea` 컨테이너로 전달합니다.
   - `traefik.http.services.servicea.loadbalancer.server.port=5432` 라벨에 따라 Traefik은 `servicea` 컨테이너의 `5432` 포트로 요청을 전달합니다.
4. **Servicea 컨테이너**:

   `servicea` 컨테이너는 `Dockerfile`에서 `EXPOSE 5432`로 정의된 포트를 통해 요청을 수신합니다.

   컨테이너 내부에서 `pnpm start` 명령어가 실행되며, 애플리케이션이 `5432` 포트에서 요청을 처리합니다.

## [`HEALTHCHECK`](https://docs.docker.com/reference/dockerfile/#healthcheck)

## [`CMD`](https://docs.docker.com/reference/dockerfile/#cmd)

- Docker 컨테이너가 시작될 때 실행되는 기본 명령을 지정
- `CMD`는 *컨테이너가 시작할 때만 실행*되며, 컨테이너가 이미 실행 중일 때는 적용되지 않는다
- 특징:
    - `docker run` 시에 명령줄 인수를 통해 덮어쓸 수 있다
    - Dockerfile 내에서 한 번만 사용할 수 있으며, 여러 번 사용되면 **마지막 `CMD`만 유효**하다
    - `CMD`가 제공하는 명령은 `ENTRYPOINT`가 정의되지 않았을 때 직접 실행된다

### 사용 예시

```Dockerfile
# 컨테이너가 시작될 때 "Hello, World!"를 출력
CMD ["echo", "Hello, World!"]
```

```Dockerfile
# 컨테이너가 시작될 때 Python 스크립트를 실행
CMD ["python", "script.py"]
```

## `ENTRYPOINT`

- 컨테이너가 시작될 때 항상 실행되어야 하는 명령을 설정
- `ENTRYPOINT`는 *컨테이너의 메인 실행 명령*으로 생각할 수 있으며, 일반적으로 변경되어서는 안 된다
- 특징:
    - `ENTRYPOINT`는 `docker run` 명령어의 인수를 `ENTRYPOINT`의 인수로 사용한다
    - `ENTRYPOINT`와 `CMD`를 함께 사용할 수 있으며, 이 경우 `CMD`는 `ENTRYPOINT`의 기본 인수로 사용된다

### 사용 예시

```Dockerfile
# 컨테이너가 시작할 때 "Hello World"를 출력
ENTRYPOINT ["echo", "Hello"]
CMD ["World"]
```

```Dockerfile
# 컨테이너가 시작될 때 항상 Python 스크립트를 실행
ENTRYPOINT ["python", "script.py"]
```

## `CMD`와 `ENTRYPOINT`의 차이

- CMD
    - 기본 명령을 제공(컨테이너의 기본값을 제공)
    - `docker run` 시에 인수를 통해 쉽게 덮어쓸 수 있다
    - 컨테이너가 실행될 때만 사용되며, 동작이 종료되면 컨테이너도 종료된다
- ENTRYPOINT
    - 컨테이너가 시작될 때 항상 실행되어야 하는 메인 명령을 설정(컨테이너가 실행해야 할 메인 프로세스를 정의)
    - 이를 통해 컨테이너를 특정 명령을 실행하는 실행 파일처럼 사용할 수 있다
    - `ENTRYPOINT`는 일반적으로 변경되어서는 안 되는 중요한 명령을 지정하는데 사용된다

## 컨테이너가 바로 종료되는 이유

- 컨테이너는, `CMD`나 `ENTRYPOINT`가 지정되지 않은 경우, 실행할 명령이 없으므로 바로 종료된다
- 예를 들어, `CMD ["echo", "Hello, World!"]`와 같이 설정하면, "Hello, World!"를 출력하고 컨테이너는 종료

`CMD ["/bin/bash"]`를 Dockerfile에 설정하면, 컨테이너는 bash 셸을 실행하고 나서 즉시 종료됩니다.

이는 `bash`가 대화형 모드로 실행되지 않기 때문입니다. Docker 컨테이너가 계속 실행되도록 하기 위해 대화형 셸을 유지하거나, 일반적으로 사용하는 방법은 컨테이너에서 무한 루프나 대기 상태를 유지하는 것입니다.

### 방법 1: 대화형 셸 유지

컨테이너가 계속 실행되도록 하기 위해 Dockerfile에 다음과 같이 설정할 수 있습니다:

```Dockerfile
FROM node:20.13.1-alpine3.20

# bash 설치
RUN apk update && apk add bash

# 기본 명령을 bash로 설정 (대화형 모드로)
CMD ["/bin/bash", "-c", "while :; do sleep 1; done"]
```

### 방법 2: `tail -f /dev/null` 사용

또 다른 방법으로는 `tail -f /dev/null` 명령을 사용하여 컨테이너가 종료되지 않도록 하는 것입니다:

```Dockerfile
FROM node:20.13.1-alpine3.20

# bash 설치
RUN apk update && apk add bash

# 기본 명령을 tail로 설정
CMD ["tail", "-f", "/dev/null"]
```

이 설정으로 컨테이너는 `tail -f /dev/null` 명령을 실행하면서 무기한 대기 상태를 유지합니다. 이 상태에서 `docker exec` 명령을 사용하여 bash 셸에 접속할 수 있습니다.

### 컨테이너 실행 및 bash 셸 접속

1. Dockerfile을 기반으로 이미지를 빌드합니다.

   ```sh
   docker build -t my-node-app .
   ```

2. 컨테이너를 실행합니다.

   ```sh
   docker run -d --name my-node-container my-node-app
   ```

3. 실행 중인 컨테이너에 bash 셸로 접속합니다.

   ```sh
   docker exec -it my-node-container /bin/bash
   ```

### 요약

- `CMD ["/bin/bash"]`는 대화형 셸을 유지하지 않으므로 컨테이너가 즉시 종료됩니다.
- 컨테이너를 계속 실행 상태로 유지하기 위해서는 `while` 루프나 `tail -f /dev/null`을 사용하여 대기 상태를 유지할 수 있습니다.
- 이렇게 설정하면 `docker exec -it container_name /bin/bash` 명령으로 컨테이너에 접속하여 bash 셸을 사용할 수 있습니다.
