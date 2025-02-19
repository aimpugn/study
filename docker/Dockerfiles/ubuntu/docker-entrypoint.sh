#!/bin/bash

if [ ! -d /sys/fs/cgroup/systemd ]; then
    mkdir -p /sys/fs/cgroup/systemd
fi

# - mountpoint
#    `/sys/fs/cgroup/systemd`가 이미 마운트되어 있는지 확인합니다.
#    마운트되어 있지 않다면, cgroup2를 마운트합니다.
# - mount
#    cgroup2는 리눅스 커널이 내부적으로 관리하는 "가상 파일 시스템"이므로, 일반적인 물리 디바이스를
#    마운트할 필요가 없기 때문에, `none`을 소스로 지정할 수 있습니다.
#    `none`을 사용하면 커널이 자동으로 cgroup2 인터페이스를 포함한 디렉토리 구조를 생성합니다.
#    `none`은 실제 디바이스가 필요 없는 "가상 파일 시스템" 마운트에 사용됩니다.
mountpoint -q /sys/fs/cgroup/systemd \
    || mount -t cgroup2 none /sys/fs/cgroup/systemd

# `exec`을 사용하여 현재 쉘 프로세스를 종료하고, 지정된 명령어를 실행합니다.
#
# `ENTRYPOINT` 지시자가 실행되면 이 쉘 스크립트가 1번 프로세스가 됩니다.
# PID 1은 init 프로세스로, 시스템 부팅시 가장 먼저 실행되어 다른 모든 프로세스의 부모 역할을 합니다.
# 다음과 같은 역할들을 담당합니다:
# - 다른 프로세스의 종료
# - 신호 처리(`SIGTERM`, `SIGKILL`)
# - 좀비 프로세스 재수거(리핑, reaping):
#   좀비 프로세스란? 리눅스 및 유닉스 계열 운영 체제에서 프로세스가 종료되었지만, 부모 프로세스가
#   `wait()` 또는 `waitpid()`를 호출하지 않아 "프로세스 테이블에 남아 있는 상태"를 의미합니다.
# - 기타 등
#
# 일반적인 쉘 스크립트는 부모 프로세스로 전달된 신호를 하위 프로세스에 자동으로 전달하지 않습니다.
# 가령 컨테이너가 `docker stop`으로 종료될 때 `SIGTERM`이 전달되지만,
# 쉘 스크립트가 명시적으로 처리하지 않으면 하위 프로세스들는 신호를 받지 못하고 계속 실행될 수 있습니다.
# ```
# #!/bin/bash
# trap 'kill -TERM $PID' TERM INT
# sleep 1000 &
# PID=$!
# # `SIGTERM`을 받으면 sleep 1000도 종료됩니다.
# wait $PID
# ```
# 이 경우 컨테이너가 정상적으로 종료되지 않고 `docker kill`로 강제 종료해야 할 수도 있습니다.
#
# 또한 PID 1이 된 프로세스는 자식 프로세스가 종료될 때 `SIGCHLD` 신호를 수신하는데,
# 이를 수거(`wait()` 호출)하지 않으면 좀비 프로세스가 발생할 수 있습니다.
# 일반적인 쉘 스크립트는 `SIGCHLD`를 처리하는 로직이 없으며, `wait` 호출이 없을 경우
# 좀비 프로세스가 지속적으로 증가할 수 있습니다.
#
# 따라서 이 쉘 프로세스 대신 `/lib/systemd/systemd`가 1번 프로세스가 되도록 합니다.
# `exec`을 사용하면 현재 실행 중인 프로세스가 새로운 프로세스로 대체됩니다.
# 이 경우 이 `docker-entrypoint.sh`의 프로스세가 `CMD`의 프로세스로 대체됩니다.
# `CMD`나 `docker run` 명령어를 통해 전달된 인자가 그대로 전달되어 "$@"로 사용할 수 있습니다.
#
# ```
# ENTRYPOINT ["/entrypoint.sh"]
# CMD ["echo", "Hello, World!"]
# # /entrypoint.sh echo "Hello, World!"
# ```
#
# - https://docs.docker.com/reference/dockerfile/#exec-form-entrypoint-example
# - https://brunch.co.kr/@growthminder/142
# - [SP - 1.5 Fundamentals of Shell & Signal](https://velog.io/@junttang/SP-1.5-%EC%8B%9C%EA%B7%B8%EB%84%90Signal-1)
# - [Zombie process reaping 에 대하여, Container에서 고려할 부분들](https://blog.hyojun.me/4)
exec "$@"
