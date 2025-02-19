#!/bin/bash

set -e

NAME=ubuntu-env

# to stop and delete running container
echo "Remove container: $NAME"
docker rm -f "$NAME"

# build
echo "Build container: $NAME"
docker build -t "$NAME" .

# ex: docker run --name asm-linux-env -it asm-linux-env
RUN_OPTIONS=(
    --name "$NAME"
    # `/run`은 운영체제의 런타임 데이터(PID 파일, 소켓 파일, 서비스 상태 파일 등)를 저장하는 임시 디렉토리로, 시스템 부팅 후 생성됩니다.
    # systemd는 `/run` 디렉토리 내에서 PID 파일, 소켓 파일 등을 관리합니다.
    # 재부팅 후에도 유지될 필요가 없기 때문에 메모리 기반 파일 시스템 [`tmpfs`](https://docs.docker.com/engine/storage/tmpfs/)을 사용하면 자동으로 초기화됩니다.
    --tmpfs /run
    # mount volume
    --volume "$(pwd)/tmp:/home/$NAME"
    # 일반적인 컨테이너 런타임(예: containerd, Docker)은 컨테이너 내부에서
    # cgroup을 자체적으로 마운트하며, 이를 위한 cgroup 네임스페이스를 자동으로 생성합니다.
    # 따라서 `/sys/fs/cgroup`을 호스트와 컨테이너 간에 직접 볼륨으로 매핑하는 것은
    # 권장되지 않습니다.
    # 그리고 OrbStack 같은 경우 WSL 2같은 경량화된 Linux VM를 사용하고,
    # Docker 경우 LinuxKit 기반의 가상 머신을 사용합니다.
    # 이 가상 머신 내에서 Docker는 Linux와 동일한 방식으로 cgroup을 관리하므로,
    # 별도로 `/sys/fs/cgroup`을 마운트할 필요가 없습니다.
    # - https://github.com/containerd/nerdctl/discussions/1659#discussioncomment-6336493
    # - https://docs.orbstack.dev/architecture#architecture
    #
    # 하지만, systemd를 사용할 경우 init.scope 등 필요한 객체들을 cgroup에 생성하는데,
    # 이를 위해서 쓰기 권한을 설정합니다.
    # 여기서 `/sys/fs/cgroup`는 OrbStack이 사용하는 Linux VM의 `/sys/fs/cgroup`입니다.
    # --volume /sys/fs/cgroup/docker:/sys/fs/cgroup/docker:rw

    # 일반적인 컨테이너 런타임은 컨테이너별 독립적인 cgroup 네임스페이스를 생성합니다.
    # 각 컨테이너는 자신의 cgroups만 볼 수 있으므로, 다른 컨테이너나 호스트에 영향을 주지 않습니다.
    # 하지만 독립적인 cgroups 초기화 및 마운트 작업 필요합니다.
    #
    # 하지만 systemd가 컨테이너 내에서 실행될 때, cgroup을 생성 및 관리할 수 있도록
    # 컨테이너가 호스트의 cgroup 네임스페이스를 공유하도록 설정합니다.
    # --cgroupns host

    # `--volume /sys/fs/cgroup/docker:/sys/fs/cgroup/docker:rw` 옵션과
    # `--cgroupns host` 옵션 사용하지 않는 방법:
    #
    # 호스트의 `/sys/fs/cgroup`을 컨테이너와 공유하지 않고, 컨테이너가 독립적인
    # 메모리 기반 파일 시스템(tmpfs)으로 `/sys/fs/cgroup`를 마운트하도록 설정합니다.
    # 이를 통해 컨테이너 내부에서 `/sys/fs/cgroup`을 자유롭게 읽고 쓸 수 있도록 합니다.
    # cgroup2 마운트는 `docker-entrypoint.sh`에서 수행합니다.
    --tmpfs /sys/fs/cgroup:rw

    # cgroup을 생성하고 관리하는 데 필요한 권한을 부여합니다.
    # 권한이 없으면 `systemd-resolved.service` 실행에 실패합니다.
    # - https://docs.docker.com/engine/containers/run#runtime-privilege-and-linux-capabilities
    --cap-add SYS_ADMIN
    -d
    "$NAME" # tag name of image to run
)

DOCKER_RUN_CMD="docker run ${RUN_OPTIONS[*]}"
DOCKER_EXEC_CMD="docker exec -u ubuntu -it $NAME /bin/zsh"

for cmd in "$DOCKER_RUN_CMD" "$DOCKER_EXEC_CMD"
do
    echo "Execute '$cmd'"
    sh -c "$cmd"
done
