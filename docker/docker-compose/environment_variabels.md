# [Environment variables](https://docs.docker.com/compose/environment-variables/)

- [Environment variables](#environment-variables)
    - [개요](#개요)
    - [환경 변수 우선 순위](#환경-변수-우선-순위)

## 개요

- The various ways you can [set environment variables in Compose](https://docs.docker.com/compose/environment-variables/set-environment-variables/).
- [How environment variable precedence works](https://docs.docker.com/compose/environment-variables/envvars-precedence/).
- The correct syntax for an [environment file](https://docs.docker.com/compose/environment-variables/env-file/).
- Changing pre-defined [environment variables](https://docs.docker.com/compose/environment-variables/envvars/).

## 환경 변수 우선 순위

The order of precedence (highest to lowest) is as follows:

1. [CLI의 `docker compose run -e`](https://docs.docker.com/compose/environment-variables/set-environment-variables/#set-environment-variables-with-docker-compose-run---env)
2. [쉘의 환경변수](https://docs.docker.com/compose/environment-variables/set-environment-variables/#substitute-from-the-shell)
3. `docker-compose.yaml` 파일의 [`environment` 속성](https://docs.docker.com/compose/environment-variables/set-environment-variables/#use-the-environment-attribute)

    ```yaml
    web:
        environment:
            - DEBUG=1
    ```

4. [CLI의 `--env-file`](https://docs.docker.com/compose/environment-variables/set-environment-variables/#substitute-with---env-file)

    ```shell
    docker compose --env-file ./config/.env.dev up
    ```

5. `docker-compose.yaml` 파일의 [`env_file` 속성](https://docs.docker.com/compose/environment-variables/set-environment-variables/#use-the-env_file-attribute)

    ```yaml
    web:
        env_file:
            - web-variables.env
    ```

6. 프로젝트 디렉토리의 베이스에 위치한 [`.env` 파일](https://docs.docker.com/compose/environment-variables/set-environment-variables/#substitute-with-an-env-file)

    ```shell
    $ cat .env
    TAG=v1.5
    ```

    ```yaml
    // compose.yml
    services:
        web:
            image: "webapp:${TAG}"
    ```

7. 컨테이너 이미지 안의 [`ENV` 지시자](https://docs.docker.com/engine/reference/builder/#env). Docker파일에 `ARG` 또는 `ENV` 설정이 있으면, `environment`, `env_file` 또는  `run --env`에 대한 Docker Compose 항목(entry) 없는 경우에만 평가된다.
