# docker-compose env

## Docker Compose에서 환경 변수를 주입

```yaml
services:
  app:
    ... 생략 ...
    environment:
      # change from development to production if needed
      NODE_ENV: development
      ADMIN_PASSWORD: ${PREFIX_ADMIN_PASSWORD}
      DB_CHARSET: ${PREFIX_DB_CHARSET:-utf8mb4}
      DB_HOST: postgres
      DB_NAME: ${PREFIX_DB_NAME:?}
      DB_PASS: ${PREFIX_DB_PASSWORD:?}
      DB_PORT: ${PREFIX_DB_PORT:-5432}
```

위와 같은 docker-compose.yaml 파일에서 사용되는 환경 변수를 주입하는 방법은 여러 가지가 있습니다.

1. [`.env` 파일](https://docs.docker.com/compose/environment-variables/variable-interpolation/#env-file) 사용:

    Docker Compose 파일과 같은 디렉토리에 `.env` 파일을 만들고 그 안에 환경 변수를 정의합니다.

    ```sh
    PREFIX_ADMIN_PASSWORD=mypassword
    PREFIX_DB_NAME=mydb
    PREFIX_DB_PASSWORD=dbpassword
    PREFIX_DB_PORT=5432
    ```

2. 명령줄에서 환경 변수 설정:

    docker-compose 명령을 실행할 때 환경 변수를 직접 설정할 수 있습니다.

    ```sh
    PREFIX_ADMIN_PASSWORD=mypassword docker-compose up
    ```

3. shell 환경 변수 사용:

    셸에서 환경 변수를 미리 설정한 후 `docker-compose`를 실행합니다.

    ```sh
    export PREFIX_ADMIN_PASSWORD=mypassword
    docker-compose up
    ```

4. `docker-compose.override.yml` 파일 사용:

    기본 `docker-compose.yml` 파일과 함께 사용되는 오버라이드 파일에 환경 변수를 정의할 수 있습니다.

    ```yaml
    version: '3'
    services:
        app:
        environment:
            ADMIN_PASSWORD: mypassword
    ```

5. [`env_file` 옵션](https://docs.docker.com/reference/compose-file/services/#env_file) 사용:

    docker-compose.yml 파일에서 env_file 옵션을 사용하여 환경 변수 파일을 지정할 수 있습니다.

    ```yaml
    services:
        app:
        env_file:
            - ./common.env
            - ./app.env
            - /opt/secrets.env
    ```

6. Docker Secrets 사용 (Swarm mode):

    Docker Swarm mode에서는 보안이 중요한 데이터를 위해 secrets를 사용할 수 있습니다.

[우선순위](https://docs.docker.com/compose/environment-variables/envvars-precedence/)는 다음과 같습니다:
1. 명령줄에서 설정된 환경 변수
2. 쉘 또는 `.env` 등을 통해 interpolated 된 환경 변수
3. Compose 파일에 [`environment`](https://docs.docker.com/compose/environment-variables/set-environment-variables/#use-the-environment-attribute) 속성에 하드코딩된 값
4. 변수의 기본값 (예: `${VARIABLE:-default}`)
5. Compose 파일의 `env_file` 속성
6. 컨테이너 이미지의 [`ENV` 지시자s](https://docs.docker.com/reference/dockerfile/#env)
