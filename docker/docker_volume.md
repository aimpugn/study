# docker volume

- [docker volume](#docker-volume)
    - [도커 볼륨](#도커-볼륨)
    - [`-v` 또는 `--mount` flag 선택](#-v-또는---mount-flag-선택)
        - [`-v` or `--volume`](#-v-or---volume)
        - [`--mount`](#--mount)
    - [바인드 마운트](#바인드-마운트)
        - [바인드 마운트와 도커 볼륨의 차이점 및 사용 이유](#바인드-마운트와-도커-볼륨의-차이점-및-사용-이유)
    - [`VOLUME` instruction](#volume-instruction)
    - [`docker run -v`](#docker-run--v)
    - [`docker volume create`](#docker-volume-create)
    - [`docker volume prune`](#docker-volume-prune)
    - [기타](#기타)

## [도커 볼륨](https://docs.docker.com/engine/storage/volumes/)

Volumes are the preferred mechanism for persisting data generated by and used by Docker containers. While bind mounts are dependent on the directory structure and OS of the host machine, volumes are completely managed by Docker. Volumes have several advantages over bind mounts:

볼륨은 Docker 컨테이너에서 생성하고 사용되는 데이터를 저장하기 위해 선호되는 메커니즘입니다.
[바인드 마운트](https://docs.docker.com/engine/storage/bind-mounts/)는 호스트 머신의 디렉토리 구조와 OS에 따라 달라지지만, 볼륨은 Docker에서 완전히 관리합니다.
볼륨은 바인드 마운트보다 몇 가지 장점이 있습니다:

- 볼륨은 바인드 마운트보다 백업 또는 마이그레이션이 더 쉽습니다.
- 볼륨은 Docker CLI 명령어 또는 Docker API를 사용하여 관리할 수 있습니다.
- 볼륨은 Linux와 Windows 컨테이너 모두에서 작동합니다.
- 볼륨은 여러 컨테이너 간에 보다 안전하게 공유할 수 있습니다.
- 볼륨 드라이버를 사용하면
    - 원격 호스트 또는 클라우드 공급자에 볼륨을 저장하거나,
    - 볼륨의 콘텐츠를 암호화하거나,
    - 다른 기능을 추가할 수 있습니다.
- 새 볼륨은 컨테이너에 의해 미리 채워진(pre-populated) 컨텐츠를 가질 수 있습니다.
- Docker 데스크톱의 볼륨은 Mac 및 Windows 호스트의 바인드 마운트보다 훨씬 더 높은 성능을 제공합니다.

> **컨테이너에 의해 미리 채워진(pre-populated) 컨텐츠를 가질 수 있다**?
>
> 데이터가 이미 존재하는 컨테이너의 특정 디렉토리에 새로운 Docker 볼륨이 마운트되면,
> Docker는 해당 디렉토리에 있던 파일들을 자동으로 볼륨에 복사합니다.
> 이를 통해 볼륨을 빈 상태로 시작하는 대신, 필요한 내용이 이미 포함된 상태로 시작할 수 있습니다.
>
> 예를 들어, 웹 애플리케이션을 위한 Docker 이미지가 초기 설정 파일들을 포함한 디렉토리를 가지고 있다고 가정해봅시다.
> 이 디렉토리에 볼륨을 마운트하면, Docker는 볼륨이 동일한 초기 설정 파일들을 갖추도록 합니다.
> 즉, 볼륨이 자동으로 미리 데이터를 가지고 시작하는 것입니다.

## `-v` 또는 `--mount` flag 선택

일반적으로 `--mount`는 더 명시적이고 장황합니다.
- `-v` 구문은 모든 옵션을 하나의 필드에 함께 결합
- `--mount` 구문은 옵션을 분리

볼륨 드라이버 옵션을 지정해야 하는 경우 `--mount`를 사용해야 합니다.

### `-v` or `--volume`

콜론(`:`)으로 구별되는 세 개의 필드로 구성됩니다.
필드는 반드시 올바른 순서여야 하고, 각 필드의 의미가 즉시 명확하지는 않습니다.

- 첫번째 필드는 볼륨의 이름입니다.
    - 이름이 있는 볼륨(named volume)인 경우, 그 이름은 주어진 호스트 머신에서 유일해야 합니다.
    - 익명 볼륨 첫번째 필드는 생략됩니다.
- 두 번째 필드는 컨테이너에 파일 또는 디렉토리가 마운트되는 경로입니다.
- 세번째 필드는 선택 사항으로, `ro` 등 콤마(`,`)로 구별되는 옵션 목록입니다.

### `--mount`

여러 `<key>=<value>` 튜플들로 구성되며 콤마(`,`)로 구별됩니다.
`--mount` 옵션의 문법은 `-v` 옵션보다 장황하지만 key의 순서는 중요하지 않고, 그 값은 이해하기 더 쉽습니다.

- `type`:
    - 값으로 [`bind`](https://docs.docker.com/engine/storage/bind-mounts/), `volume` 또는 [`tmpfs`](https://docs.docker.com/engine/storage/tmpfs/)가 올 수 있습니다.
    - 이 항목에서는 볼륨에 대해 설명하므로 유형은 항상 볼륨입니다.
- `source`:
    - 명명된 볼륨의 경우 볼륨의 이름
    - 익명 볼륨의 경우 이 필드는 생략됩니다.
    - `source` 또는 `src`로 지정할 수 있습니다.
- `destination`:
    - 컨테이너에서 파일 또는 디렉토리가 마운트되는 경로를 값으로 사용합니다.
    - `destination`, `dst`, 또는 `target`으로 지정할 수 있습니다.
- `volume-subpath` 옵션
    - 컨테이너에 마운트할 볼륨 내 하위 디렉토리의 경로를 사용합니다.
    - 볼륨을 컨테이너에 마운트하기 전에 볼륨에 하위 디렉토리가 존재해야 합니다.
    - [볼륨 하위 디렉토리 마운트하기](https://docs.docker.com/engine/storage/volumes/#mount-a-volume-subdirectory)를 참조하세요.
- `readonly` 옵션
    - 바인드 마운트가 [컨테이너에 읽기 전용으로 마운트](https://docs.docker.com/engine/storage/volumes/#use-a-read-only-volume)됩니다.
    - `readonly` 또는`ro`로 지정할 수 있습니다.
- `volume-opt` 옵션
    - 한 번 이상 지정할 수 있으며 옵션 이름과 해당 값으로 구성된 키-값 쌍을 사용합니다.

## 바인드 마운트

바인드 마운트는 특정 호스트 디렉토리를 컨테이너 내의 디렉토리에 직접 연결합니다.
이 방식은 컨테이너가 삭제되더라도 호스트 시스템의 데이터가 유지되며, 컨테이너 내에서 변경된 내용이 즉시 호스트 시스템에 반영됩니다.

가령 `/path/to/some/application` 경로에 있는 소스 코드와 파일들을 컨테이너의 `/home/user` 디렉토리로 마운트할 때 사용합니다.

Dockerfile로 이미지를 빌드하고 컨테이너를 실행할 때 바인드 마운트를 사용할 수 있습니다.

```sh
docker build -t ostep .

docker run --name ostep -it \
    -v /path/to/some/application:/home/ostep \
    ostep
```

- `-v /path/to/some/application:/home/ostep`:
    `/path/to/some/application` 디렉토리를 컨테이너의 `/home/ostep` 디렉토리로 마운트합니다.

    `/path/to/some/application`의 모든 변경 사항은 컨테이너 내에서 바로 적용되며, 컨테이너가 삭제되더라도 데이터는 호스트 시스템에 그대로 남아 있습니다.

    그리고 새로운 익명의 볼륨이 생성되지 않습니다.

### 바인드 마운트와 도커 볼륨의 차이점 및 사용 이유

- **바인드 마운트**는 호스트의 특정 디렉토리를 컨테이너의 디렉토리와 연결합니다. 마운트 된 호스트의 디렉토리나 파일을 컨테이너 내에서 그대로 접근하고 사용할 수 있습니다.

- **도커 볼륨**은 도커가 관리하는 특정 위치에 저장되며, 컨테이너 간에 데이터를 공유하거나 지속성 있는 데이터를 보관할 때 사용됩니다. 도커 볼륨은 기본적으로 익명으로 생성될 수 있으며, 삭제하기 전까지 데이터가 보관됩니다.

바인드 마운트를 사용하면 호스트 시스템의 특정 디렉토리를 컨테이너 내부의 특정 디렉토리로 직접 연결할 수 있습니다. 이 방식은 다음과 같은 장점이 있습니다:
- **실시간 데이터 동기화**: 컨테이너 내에서 변경된 사항이 즉시 호스트 디렉토리에 반영됩니다.
- **지속성**: 컨테이너를 제거하더라도 호스트에 있는 데이터는 유지됩니다.
- **익명 볼륨 방지**: 바인드 마운트를 사용하면 익명 볼륨이 생성되지 않으므로, 볼륨 관리가 간단해집니다.

## [`VOLUME`](https://docs.docker.com/reference/dockerfile/#volume) instruction

`VOLUME` 명령은 지정된 이름의 마운트 포인트를 생성하고 외부의 네이티브 호스트 또는 다른 컨테이너에서  마운트된 볼륨을 보관하는 것으로 표시합니다.
`VOLUME` 명령어는 컨테이너에서 해당 디렉터리를 익명 볼륨으로 자동으로 마운트하겠다는 것을 의미합니다.
따라서 컨테이너가 시작될 때마다 Docker가 자동으로 새로운 익명 볼륨을 생성하여 이 디렉터리를 채우게 됩니다.

다음과 같이 사용할 수 있습니다:

```Dockerfile
# JSON 배열 형식
VOLUME ["/var/log/"]
# 평문 형식으로 단일 값
VOLUME "/var/log/"
# 평문 형식으로 여러 값
VOLUME "/var/log" "/var/db"
```

이렇게 볼륨이 선언된 채 생성된 이미지의 경우 `docker run`으로 컨테이너가 실행될 때 자동으로 해당 경로를 host에 연결합니다.
그 경로는 `/var/lib/docker/volumes/{volume_name}` 에 만들어집니다.

```Dockerfile
FROM ubuntu
RUN mkdir /myvol
RUN echo "hello world" > /myvol/greeting
VOLUME /myvol
```

- `docker run`이 `/myvol`에 새 마운트 지점을 생성
- `greeting` 파일을 새로 생성된 볼륨에 복사하도록 하는 이미지를 생성

## `docker run -v`

```sh
docker run -itd -v /host/some/where:/container/some/where ubuntu
```

호스트 머신의 파일 시스템과 container의 파일 시스템이 연결됩니다.
혹시 Dockerfile에 정의된 volume과 같은 경로를 잡는다면 `docker run -v`에서 `-v` 옵션의 값이 image level의 volume을 오버라이드 합니다.

`docker run`시에 자유롭게 지정하여 사용할 수 있지만, `docker volume ls` 같은 명령어로 추적이 안됩니다. container를 run 시킨 사람이 알아서 잘 관리해줘야 합니다.

## [`docker volume create`](https://docs.docker.com/reference/cli/docker/volume/create/)

```sh
docker volume create hello

hello

docker run -d -v hello:/world busybox ls /world
```

마운트는 컨테이너의 `/world` 디렉토리 내에 생성됩니다.
Docker는 컨테이너 내부의 마운트 지점에 대한 상대 경로를 지원하지 않습니다.

여러 컨테이너가 동일한 볼륨을 사용할 수 있습니다.
이 기능은 두 컨테이너가 공유 데이터에 액세스해야 하는 경우에 유용합니다.
예를 들어 한 컨테이너는 데이터를 쓰고 다른 컨테이너는 데이터를 읽는 경우입니다.

```sh
docker volume create \
    -d local \
    ostep
```

## [`docker volume prune`](https://docs.docker.com/reference/cli/docker/volume/prune/)

## 기타

- [docker volume의 사용방법과 차이점](https://darkrasid.github.io/docker/container/volume/2017/05/10/docker-volumes.html)
