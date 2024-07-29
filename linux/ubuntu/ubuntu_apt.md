# Ubuntu apt

- [Ubuntu apt](#ubuntu-apt)
    - [`apt`(Advanced Package Tool)](#aptadvanced-package-tool)
    - [스크립트 사용법 및 다른 APT 도구와의 차이점](#스크립트-사용법-및-다른-apt-도구와의-차이점)
    - [`apt update`](#apt-update)

## `apt`(Advanced Package Tool)

`apt`는 패키지 관리 시스템을 위한 높은 수준의 명령줄 인터페이스를 제공합니다.
이는 최종 사용자 인터페이스를 위한 것으로, 기본적으로 `apt-get(8)` 및 `apt-cache(8)`와 같은 보다 전문적인 APT 도구에 비해 *대화형 사용에 더 적합*한 일부 옵션을 사용할 수 있습니다.

관련 man 페이지
- /usr/share/doc/apt-doc/*
- `apt-get(8)`
- `apt-cache(8)`
- `sources.list(5)`
- `apt.conf(5)`
- `apt-config(8)`
- `apt-patterns(7)`,
- The APT User's guide in `/usr/share/doc/apt/examples`, `apt_preferences(5)`

## 스크립트 사용법 및 다른 APT 도구와의 차이점

`apt(8)` 명령줄은 최종 사용자 도구로 설계되었으며 버전 간에 동작이 변경될 수 있습니다.
이전 버전과의 호환성을 깨지 않으려고 노력하지만 대화형 사용에 도움이 되는 변경 사항도 보장되지 않습니다.

`apt(8)`의 모든 기능은 `apt-get(8)` 및 `apt-cache(8)`와 같은 전용 APT 도구에서도 사용할 수 있습니다.
`apt(8)`는 일부 옵션의 기본값만 변경합니다(`apt.conf(5)`, 특히 Binary 범위 참조).
따라서 이전 버전과의 호환성을 최대한 유지하려면 스크립트에서 이러한 명령(일부 추가 옵션을 활성화한 상태)을 사용하는 것이 좋습니다.

`apt`는 일반 사용자를 위한 더 간단하고 사용자 친화적인 인터페이스 제공합니다.
- 더 현대적이고 간결한 명령어 구조
- 진행 상황 표시 바, 컬러 출력 등 시각적 피드백 제공
- 가장 자주 사용되는 apt-get 및 apt-cache 명령어들의 기능 통합

`apt-get`은 더 오래되고 스크립트 친화적인 명령어입니다.

주요 명령어를 비교하면 다음과 같습니다.

```sh
# apt
apt update
apt install
apt remove
apt upgrade
apt full-upgrade
apt search
apt show

# apt-get & apt-cache
apt-get update
apt-get install
apt-get remove
apt-get upgrade
apt-get dist-upgrade
apt-cache search
apt-cache show
```

apt는 일반 사용자에게 더 친숙하고 간편한 인터페이스를 제공하며, apt-get은 더 세밀한 제어와 스크립팅에 적합합니다.
최신 Ubuntu나 Debian 시스템에서는 apt를 사용하는 것이 권장됩니다.

## `apt update`

`apt update`는 `/etc/apt/sources.list`와 `/etc/apt/sources.list.d/` 디렉토리에 정의된 저장소에서 패키지 정보를 가져오고,
이 정보를 로컬 데이터베이스에 저장합니다.
