# Localstack

- [Localstack](#localstack)
    - [install](#install)
        - [cli](#cli)
        - [cli options](#cli-options)
    - [docker-compose](#docker-compose)
    - [integrations](#integrations)
        - [aws-cli](#aws-cli)

## [install](https://docs.localstack.cloud/getting-started/installation/)

### cli

```shell
brew install localstack/tap/localstack-cli
```

`docker-compose.yaml`로 실행한다고 해도, `localstack config validate`로 config 체크하기 위해 설치

### cli options

```shell
❯ localstack -h
Usage: localstack [OPTIONS] COMMAND [ARGS]...

  The LocalStack Command Line Interface (CLI)

Options:
  -v, --version       Show the version and exit.
  -d, --debug         Enable CLI debugging mode
  -p, --profile TEXT  Set the configuration profile
  -h, --help          Show this message and exit.

Commands:
  auth        (Beta) Authenticate with your LocalStack account
  completion  CLI shell completion
  config      Manage your LocalStack config
  infra       (Deprecated) Manage LocalStack infrastructure
  license     (Beta) Manage and verify your LocalStack license
  login       Login to the LocalStack Platform
  logs        Show LocalStack logs
  pod         Manage the state of your instance via Cloud Pods
  ssh         Obtain a shell in LocalStack
  start       Start LocalStack
  state       (Beta) Manage the persistence state of localstack
  status      Query status info
  stop        Stop LocalStack
  update      Update LocalStack
  wait        Wait for LocalStack
```

## docker-compose

```yaml
version: "3.8"

services:
    localstack:
        container_name: "${LOCALSTACK_DOCKER_NAME-localstack_main}"
        image: localstack/localstack
        ports:
            - "127.0.0.1:4566:4566" # LocalStack Gateway
            - "127.0.0.1:4510-4559:4510-4559" # external services port range
        environment:
            - DEBUG=${DEBUG-}
            - DOCKER_HOST=unix:///var/run/docker.sock
        volumes:
            - "${LOCALSTACK_VOLUME_DIR:-./volume}:/var/lib/localstack"
            - "/var/run/docker.sock:/var/run/docker.sock"
```

- This command pulls the current nightly build from the master branch (if you don’t have the image locally) and **not the latest supported version**.
- If you are using LocalStack with an [API key](https://docs.localstack.cloud/getting-started/api-key/), you need to specify the image tag as `localstack/localstack-pro` in the `docker-compose.yml` file.
- This command reuses the image if it’s already on your machine, i.e. it will **not** pull the latest image automatically from Docker Hub.
- Mounting the Docker socket `/var/run/docker.sock` as a volume is required for the Lambda service. Check out the [Lambda providers](https://docs.localstack.cloud/user-guide/aws/lambda/) documentation for more information.
- To facilitate interoperability, configuration variables can be prefixed with `LOCALSTACK_` in docker.
    - For instance, setting `LOCALSTACK_PERSISTENCE=1` is equivalent to `PERSISTENCE=1`.
- If using the Docker default bridge network using `network_mode: bridge`, container name resolution will not work inside your containers. Please consider removing it, if this functionality is needed.

Please note that there are a few pitfalls when configuring your stack manually via `docker-compose` (e.g., required container name, Docker network, volume mounts, and environment variables). We recommend using the LocalStack CLI to validate your configuration, which will print warning messages in case it detects any potential misconfigurations:

```shell
localstack config validate
```

```shell
❯ localstack config validate -f ./docker-compose.corewebhook.yaml
✔ config valid
```

## integrations

### [aws-cli](https://docs.localstack.cloud/user-guide/integrations/aws-cli/)
