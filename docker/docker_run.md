# [docker run](https://docs.docker.com/engine/reference/commandline/run/)

- [docker run](#docker-run)
    - [기본적인 실행](#기본적인-실행)
    - [`container_id`로 컨테이너 시작](#container_id로-컨테이너-시작)
    - [options](#options)
        - [`-d`,`--detach`](#-d--detach)
        - [Mount volume `-v`](#mount-volume--v)

## 기본적인 실행

```shell
# 이미지 빌드
docker build . -t $SOME_NAME

# 빌드된 이미지로 컨테이너 실행
docker run -it $SOME_NAME
```

- `-i`: interactive
- `-t`: pseudo-TTY 할당

## `container_id`로 컨테이너 시작

```shell
/usr/bin/docker start $CONTAINER_ID
```

## options

### `-d`,`--detach`

> Run container in background and print container ID

```sh
docker run --name test --net mynet -d nginx:alpine
```

### Mount volume [`-v`](https://docs.docker.com/reference/cli/docker/container/run/#volume)

- `-v /some-data:/data` 옵션

    호스트 디렉토리 `/some-data`를 컨테이너 파일 시스템의 `/data` 경로에 마운트하라는 지시합니다.

    **볼륨 해석**:
    Docker는 호스트 머신에 `/some-data` 디렉토리가 존재하는지 확인합니다.
    만약 존재하지 않으면, Docker는 이 디렉토리를 자동으로 생성합니다.

    **마운팅**:
    Docker는 호스트 디렉토리 `/some-data`를 컨테이너의 `/data` 디렉토리에 마운트합니다.
    이 마운트 작업은 운영 체제에서 관리되며, `/data` 디렉토리에서 이루어진 모든 변경 사항은 호스트의 `/some-data` 디렉토리에 반영됩니다.

볼륨 마운트를 통해 데이터는 컨테이너의 라이프사이클과 관계없이 지속됩니다.
즉, 컨테이너가 중지되거나 삭제되더라도, 호스트 디렉토리인 `/some-data`에 있는 데이터는 유지됩니다.
이를 통해 컨테이너를 재시작하거나 재구성할 때 데이터 손실 없이 작업을 지속할 수 있습니다.
