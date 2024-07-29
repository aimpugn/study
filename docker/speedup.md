```shell
COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose -p test up --build
```

#### docker-compose 옵션

- [Faster builds in Docker Compose 1.25.1 thanks to BuildKit Support](https://www.docker.com/blog/faster-builds-in-compose-thanks-to-buildkit-support/)

##### `COMPOSE_DOCKER_CLI_BUILD`

This instruction tells docker-compose to use the Docker CLI when executing a build. 
You should see the same build output, but starting with the experimental warning.

As docker-compose passes its environment variables to the Docker CLI, we can also tell the CLI to use BuildKit instead of the default builder. 
To accomplish that, we can execute this:

##### `DOCKER_BUILDKIT`

- [Build images with BuildKit](https://docs.docker.com/develop/develop-images/build_enhancements/)

##### 빌드 캐시 제거

```shell
docker builder prune -a
```