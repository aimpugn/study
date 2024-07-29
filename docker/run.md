# [docker run](https://docs.docker.com/engine/reference/commandline/run/)

## 기본적인 실행

### 1. 빌드

```shell
docker build . -t $SOME_NAME
```

- `-t`: 태그

### 2. 실행

```shell
docker run -it $SOME_NAME
```

- `-i`: interactive
- `-t`: pseudo-TTY 할당

## `container_id`로 컨테이너 시작

```shell
/usr/bin/docker start $CONTAINER_ID
```
