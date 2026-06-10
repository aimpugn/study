# 권한 격리와 최소 권한 서비스 계정

한 계정에게 "자기 작업 공간 안에서만 강한 권한, 그 밖에선 무력"을 주고 싶을 때, 리눅스에서 그게 정확히 어떤 메커니즘으로 가능한지 정리합니다. 출발 질문은 구체적입니다. **"root로 접속해 `nexus` 계정을 만들고, 그 계정이 Nexus와 Maven을 설치·운영하도록 하되, 자기 공간을 벗어나면 권한이 없게 하고 싶다. 그리고 그 계정 로그인은 공개키 방식으로 하고 싶다. 가능한가?"**

이 문서는 명령어 사용법 문서가 아니라 개념·원리 허브입니다. 개별 명령 사용법은 이미 있는 [`commands/sudo.md`](commands/sudo.md), [`commands/useradd.md`](commands/useradd.md), [`commands/su.md`](commands/su.md), [`systemd/systemd.md`](systemd/systemd.md), [`namespaces.md`](namespaces.md)로 연결하고, 여기서는 그것들이 어떻게 하나의 설계로 엮이는지를 닫습니다.

## 목차

- [한 줄 답](#한-줄-답)
- [핵심 오해 권한 상승과 격리는 다른 축이다](#핵심-오해-권한-상승과-격리는-다른-축이다)
- [먼저 쌓는 기초 사용자와 권한과 프로세스](#먼저-쌓는-기초-사용자와-권한과-프로세스)
- [sudo 의 실제 권한 모델](#sudo-의-실제-권한-모델)
- [왜 sudo 로 작업 공간에만 가둘 수 없는가](#왜-sudo-로-작업-공간에만-가둘-수-없는가)
- [좁게 줘도 새는 길 sudo 권한 상승 함정](#좁게-줘도-새는-길-sudo-권한-상승-함정)
- [진짜로 가두는 방법 격리의 스펙트럼](#진짜로-가두는-방법-격리의-스펙트럼)
- [키 기반 로그인 설정과 ssh keygen](#키-기반-로그인-설정과-ssh-keygen)
- [권장 설계 nexus 와 maven 레시피](#권장-설계-nexus-와-maven-레시피)
- [경계가 지켜지는지 검증하기](#경계가-지켜지는지-검증하기)
- [정리와 teach back](#정리와-teach-back)
- [실제 사례](#실제-사례)
- [Nexus 운영 토폴로지와 데이터베이스](#nexus-운영-토폴로지와-데이터베이스)
- [근거와 더 읽을거리](#근거와-더-읽을거리)

## 한 줄 답

sudo만으로 "자기 작업 공간 안에서만 root, 밖에선 무력"은 만들 수 없습니다. 하지만 원래 목표인 "`nexus` 계정이 Nexus·Maven을 설치·운영하되 시스템 나머지는 못 건드림"은 아주 쉽게 달성됩니다. 핵심은 **root를 가두는 게 아니라, 애초에 권한을 주지 않고 소유권으로 작업 공간만 내주는 것**입니다.

이게 성립하는 이유는 한 문장으로 줄일 수 있습니다. **Nexus도 Maven도 설치·실행에 root가 거의 필요 없습니다.** 둘 다 자바 애플리케이션이라 자기 홈 디렉터리에 압축을 풀고 비특권 포트(Nexus 기본값 8081)로 실행하면 끝입니다. 그러면 "공간 밖에선 권한 없음"은 별도 장치 없이 그냥 따라옵니다. 비특권 사용자는 원래 남의 파일을 못 건드리니까요.

root가 진짜 필요한 일은 손에 꼽습니다. 부팅 시 자동 시작(systemd 유닛 등록), 파일 디스크립터 상한 상향, 방화벽 포트 개방, 1024 미만 포트 바인딩 정도입니다. 이건 관리자가 한 번 해 주거나, 그 몇 개 명령만 좁게 위임하면 됩니다.

## 핵심 오해 권한 상승과 격리는 다른 축이다

질문에 섞여 있는 두 개념을 먼저 분리해야 합니다. 이 둘은 전혀 다른 메커니즘이고, sudo는 그중 한쪽만 담당합니다.

1. 권한 상승(privilege elevation)

    "이 명령을 더 높은 권한으로 실행한다." sudo, su, setuid 바이너리가 여기에 속합니다. sudo가 답하는 질문은 오직 **"누가, 어떤 명령을, 어떤 대상 사용자로, 어느 호스트에서 실행해도 되는가"**입니다.

2. 격리/가두기(confinement)

    "이 프로세스가 볼 수 있고 건드릴 수 있는 범위를 제한한다." 파일 소유권과 퍼미션(DAC), Linux capabilities, 네임스페이스/컨테이너, systemd 샌드박싱, SELinux/AppArmor가 여기에 속합니다.

"자기 공간 안에서만 강하다"는 격리 쪽 개념인데, sudo는 권한 상승 쪽 도구입니다. **sudo의 정책에는 "어느 디렉터리 안에서만"이라는 축 자체가 없습니다.** 그래서 sudo로는 그 요구를 표현할 방법이 없습니다. 일단 어떤 명령이 sudo로 root(UID 0)가 되어 실행되면, 그 프로세스는 디렉터리와 무관하게 완전한 root입니다.

흥미롭게도 "공간 안에선 root처럼 강하고 밖에선 무력"이라는 직관에 정확히 대응하는 메커니즘이 따로 있습니다. **네임스페이스와 컨테이너**입니다. 컨테이너 안의 root는 호스트에서는 비특권 사용자로 매핑됩니다. 그 커널 원리는 이미 [`namespaces.md`](namespaces.md)의 사용자 네임스페이스 절에 정리돼 있습니다. 즉 원하는 그림 자체는 리눅스에 존재하지만, 그걸 만드는 도구가 sudo가 아닐 뿐입니다.

## 먼저 쌓는 기초 사용자와 권한과 프로세스

뒤의 설계를 이해하려면 세 가지 기초가 필요합니다. 사용자가 무엇인지, 파일 권한이 어떻게 경계를 만드는지, 프로세스가 권한을 어떻게 들고 다니는지입니다.

### 사용자와 UID 그리고 root

리눅스에서 사용자는 사람 이름이 아니라 숫자(UID)입니다. 이름은 `/etc/passwd`가 UID로 번역해 주는 라벨일 뿐입니다. 그룹도 마찬가지로 GID라는 숫자입니다.

```sh
id nexus
# uid=1001(nexus) gid=1001(nexus) groups=1001(nexus)
```

여기서 결정적인 사실 하나. **root는 "관리자라는 역할 이름"이 아니라 UID가 0인 사용자**입니다. 커널은 UID가 0인지만 봅니다. 거의 모든 권한 검사는 "이 프로세스의 effective UID가 0인가"로 갈립니다. 그래서 UID 0이 되면 그 즉시 시스템 전체에 대한 권한을 얻습니다. "root인데 일부만"이라는 상태는 기본 권한 모델에는 없습니다. 그 "일부만"을 만들려면 뒤에 나오는 capabilities나 네임스페이스 같은 별도 장치가 필요합니다.

### 파일 권한 DAC 가 작업 공간 경계를 만든다

평범한 사용자가 "남의 영역을 못 건드리는" 이유가 바로 작업 공간 경계의 정체입니다. 이건 임의 접근 제어(DAC, Discretionary Access Control)라고 부르며, 모든 파일이 소유자(user), 소유 그룹(group), 그 외(others)에 대해 각각 읽기·쓰기·실행(rwx) 비트를 가집니다.

```sh
ls -ld /opt/nexus /etc
# drwxr-x---. 5 nexus nexus  4096 ... /opt/nexus
# drwxr-xr-x. 1 root  root  12288 ... /etc
```

`/opt/nexus`는 소유자가 `nexus`라서 nexus가 자유롭게 읽고 씁니다. 반면 `/etc`는 소유자가 root이고 others에는 쓰기 비트가 없습니다. 그래서 nexus가 `/etc`에 파일을 만들려 하면 커널이 막습니다.

```text
$ touch /etc/test.txt
touch: cannot touch '/etc/test.txt': Permission denied
```

바로 이 `Permission denied`가 "작업 공간 밖에선 권한 없음"의 실제 구현입니다. 별도 설정이 아니라, 그냥 nexus가 그 디렉터리들의 소유자가 아니기 때문입니다. 그러니 우리가 할 일은 거꾸로 단순합니다. **nexus가 일할 디렉터리만 nexus 소유로 만들어 주면, 나머지는 기본 DAC가 알아서 막아 줍니다.**

여기에 특수 비트 세 개가 더 있습니다. setuid, setgid, sticky bit입니다. 이 중 다음 절의 권한 상승과 직접 얽히는 건 setuid입니다.

### 프로세스 자격증명과 setuid

프로세스는 자기를 실행한 사용자의 UID를 들고 실행됩니다. 더 정확히는 real UID와 effective UID를 따로 가지는데, 권한 검사에 쓰이는 건 effective UID입니다.

그런데 어떻게 일반 사용자가 `passwd` 명령으로 root 소유인 `/etc/shadow`를 바꿀 수 있을까요? `passwd` 실행 파일에 **setuid 비트**가 걸려 있기 때문입니다. setuid 비트가 있는 파일을 실행하면, 프로세스의 effective UID가 실행자가 아니라 그 파일의 소유자로 바뀝니다.

```sh
ls -l /usr/bin/passwd /usr/bin/sudo
# -rwsr-xr-x. 1 root root ... /usr/bin/passwd
# -rwsr-xr-x. 1 root root ... /usr/bin/sudo
```

소유자 실행 비트 자리에 `x` 대신 `s`가 보이는 게 setuid입니다. 둘 다 소유자가 root이므로 실행되는 순간 effective UID가 0이 됩니다. `sudo`도 정확히 이 원리로 동작합니다. sudo 자체가 root 소유의 setuid 바이너리라서, 일반 사용자가 실행해도 내부에서 root 권한을 얻은 뒤 정책을 확인하고 대상 명령을 실행합니다.

이 사실이 다음 절의 핵심으로 이어집니다. sudo는 권한을 "0이냐 아니냐"로 다루지, "0이되 어디까지"로 다루지 않습니다.

## sudo 의 실제 권한 모델

sudo의 정책은 `/etc/sudoers`와 `/etc/sudoers.d/*` 파일에 적힙니다. 한 줄의 규칙은 다음 형태입니다.

```conf
# 사용자  호스트=(대상사용자:대상그룹)  태그:  명령목록
nexus     ALL=(root)                   NOPASSWD: /usr/bin/systemctl restart nexus
```

이 한 줄에서 sudo가 제어할 수 있는 축을 모두 읽을 수 있습니다.

- 누가: `nexus`. 사용자나 `%group` 또는 별칭.
- 어느 호스트에서: `ALL`. 한 sudoers를 여러 머신에 배포할 때만 의미가 큽니다.
- 어떤 대상 사용자로: `(root)`. 생략하면 root. `(postgres)`처럼 다른 사용자로 한정할 수도 있습니다.
- 비밀번호 필요 여부: `NOPASSWD` 또는 기본(필요).
- 어떤 명령을: `/usr/bin/systemctl restart nexus`. 절대 경로와 인자까지 못박을 수 있습니다.

핵심은 이 목록에 **"어느 경로 안에서만"이라는 항목이 없다**는 것입니다. sudo는 "어떤 명령을 어떤 신분으로 실행해도 되는가"까지만 표현합니다. 그 명령이 root로 실행된 뒤 어디를 건드리는지는 sudo의 관심사가 아닙니다.

안전하게 쓰기 위한 기본값도 함께 정리해 둡니다. 이건 [`commands/sudo.md`](commands/sudo.md)의 `env_reset`, `secure_path` 설명과 이어집니다.

- 직접 편집은 `visudo`로만 합니다. 문법 오류가 있는 sudoers는 sudo 전체를 잠가 버릴 수 있는데, `visudo`는 저장 전에 문법을 검사합니다. 드롭인 파일도 `visudo -f /etc/sudoers.d/nexus`로 편집합니다.
- 드롭인 파일은 소유자 root, 권한 0440이어야 하고 그렇지 않으면 무시됩니다.
- `env_reset`(기본 on)으로 호출자의 환경 변수를 비웁니다. `secure_path`로 PATH를 고정해, 공격자가 PATH를 조작해 가짜 바이너리를 끼워 넣는 걸 막습니다.

## 왜 sudo 로 작업 공간에만 가둘 수 없는가

이제 핵심 질문에 정면으로 답합니다. "nexus에게 `/opt/nexus` 안에서만 root를 주자"가 왜 불가능한지를 메커니즘으로 보입니다.

가장 흔한 시도는 이렇습니다.

```conf
# 의도: nexus 가 자기 공간에서 뭐든 root 로 하게 해 주자
nexus ALL=(root) NOPASSWD: ALL
```

이 줄은 "공간 안에서만"이 아니라 그냥 **무제한 root**입니다. nexus는 즉시 다음을 할 수 있습니다.

```sh
sudo -i                      # 완전한 root 셸
sudo cat /etc/shadow         # 모든 비밀번호 해시 열람
sudo visudo                  # 자기 권한을 영구히 확장
sudo su -                    # 사실상 시스템 장악
```

"그럼 명령을 `/opt/nexus` 안의 것으로 제한하면?" 이것도 경계가 되지 못합니다. root로 실행되는 프로그램은 인자로 받은 경로 밖을 얼마든지 건드릴 수 있기 때문입니다. 예를 들어 `sudo /opt/nexus/bin/something` 류를 허용했는데 그 스크립트가 임의 명령을 부르거나, 애초에 root 권한 프로세스가 `/etc`를 쓰는 걸 sudo가 막을 방법이 없습니다.

sudo에 디렉터리 비슷한 옵션이 아예 없는 건 아닙니다. 정직하게 짚어 둡니다. sudo 1.9.3부터 `--chroot`(`-R`)와 sudoers의 `runchroot`, 그리고 작업 디렉터리를 정하는 `runcwd`가 있습니다. 하지만 **chroot는 root에 대한 보안 경계가 아닙니다.** chroot는 보이는 루트 디렉터리만 바꿀 뿐이고, root 권한 프로세스는 잘 알려진 기법으로 chroot를 빠져나옵니다. 그래서 이 옵션들은 편의 기능이지 "root를 가두는" 용도가 못 됩니다.

정리하면 이렇습니다. **sudo로 표현할 수 있는 가장 안전한 형태는 "디렉터리 제한"이 아니라 "꼭 필요한 명령 몇 개만 정확히 위임"입니다.** 디렉터리 경계는 sudo가 아니라 소유권(DAC)과 뒤의 격리 장치들이 만듭니다.

## 좁게 줘도 새는 길 sudo 권한 상승 함정

명령을 좁게 화이트리스트했다고 끝이 아닙니다. 어떤 명령은 그 자체로 셸 탈출구(shell escape)를 품고 있어서, 한 개만 허용해도 full root로 번질 수 있습니다. 설계할 때 반드시 피해야 하는 패턴들입니다.

- 페이저를 띄우는 명령. `sudo systemctl status nexus`는 출력이 길면 페이저(`less`)를 root로 띄웁니다. `less` 안에서 `!sh`를 치면 root 셸이 열립니다. 그래서 status를 위임해야 한다면 반드시 `--no-pager`를 붙이거나 `SYSTEMD_PAGER=` 같은 환경을 고정합니다.
- 편집기를 여는 명령. `sudo vi`, `sudo systemctl edit`처럼 편집기를 root로 열 수 있으면 `:!sh`로 탈출합니다. 파일을 안전하게 편집해야 하면 `sudoedit`(`sudo -e`)를 씁니다. 이건 편집기를 호출자 권한으로 실행하고 결과만 권한으로 반영합니다.
- 임의 하위 명령을 실행하는 도구. `find ... -exec`, `tar --to-command`, `awk`의 `system()`, 패키지 매니저의 훅 스크립트 등은 내부에서 다른 명령을 부를 수 있습니다.
- 와일드카드. sudoers의 `*`는 직관과 다르게 매칭됩니다. `systemctl restart *`는 `systemctl restart sshd`도 통과시킵니다. 와일드카드 대신 유닛 이름까지 박습니다.

추가 방어 장치도 있습니다. sudoers의 `NOEXEC` 태그는 위임된 프로그램이 다른 프로그램을 새로 실행하지 못하게 막습니다(동적 링크 바이너리에 한해 `noexec` 라이브러리를 끼우는 방식). 만능은 아니지만 셸 탈출 표면을 줄입니다.

이 함정들 때문에라도, 위임은 "정말 필요한 동사 + 정확한 대상"까지만 좁히는 게 정석입니다.

## 진짜로 가두는 방법 격리의 스펙트럼

"공간 안에서 충분히 일하고 밖에선 무력"을 실제로 만드는 장치들을, 약하고 단순한 것부터 강하고 복잡한 것 순으로 정리합니다. nexus 같은 평범한 자바 서비스에는 보통 1번과 2번, 그리고 4번이면 충분합니다.

### 권한을 주지 않는다 최소 권한과 소유권 경계

가장 강력하면서 가장 단순한 방법입니다. nexus에 sudo를 아예 주지 않고, 일할 디렉터리만 nexus 소유로 만든 뒤, 모든 설치를 그 안에서 합니다. 그러면 경계는 DAC가 공짜로 만들어 줍니다. 이게 권장 기본값이고, 뒤의 레시피가 이 방식입니다.

### 좁은 명령 화이트리스트 sudo

서비스 시작·정지처럼 root가 꼭 필요한 동작이 몇 개 있다면, 그것만 sudoers 드롭인으로 위임합니다. 앞 절의 함정을 피해 동사와 대상을 못박습니다.

```conf
# /etc/sudoers.d/nexus  (visudo -f 로 편집, 권한 0440)
nexus ALL=(root) NOPASSWD: /usr/bin/systemctl start nexus, \
                           /usr/bin/systemctl stop nexus, \
                           /usr/bin/systemctl restart nexus, \
                           /usr/bin/systemctl --no-pager status nexus
```

### Linux capabilities 한 조각 권한

root의 전능함은 사실 여러 개의 capability로 쪼개져 있습니다. 예를 들어 1024 미만 포트 바인딩은 `CAP_NET_BIND_SERVICE` 하나입니다. full root 대신 이 한 조각만 주면, "80 포트는 열되 나머지는 일반 사용자"가 됩니다.

```sh
# 특정 바이너리에 직접 부여 (공유 바이너리에는 신중)
sudo setcap 'cap_net_bind_service=+ep' /path/to/binary
```

다만 Nexus는 8081(비특권 포트)을 쓰므로 이 기능이 필요 없습니다. 80/443으로 노출하고 싶다면 capability보다 리버스 프록시([`commands/caddy.md`](commands/caddy.md))를 앞에 두거나, 다음 항목의 systemd `AmbientCapabilities`를 쓰는 편이 깔끔합니다.

### systemd 서비스 샌드박싱

서비스를 systemd 유닛으로 띄우면, 그 프로세스가 볼 수 있는 파일시스템과 가질 수 있는 권한을 유닛 파일에서 직접 좁힐 수 있습니다. 이게 "공간 안에서만"을 서비스 레벨에서 구현하는 정석입니다. 디렉티브 상세는 [`systemd/systemd.md`](systemd/systemd.md)에 있고, 핵심만 추리면 다음과 같습니다.

```ini
[Service]
User=nexus
Group=nexus
# 파일시스템 대부분을 읽기 전용으로, 그리고 일부만 쓰기 허용
ProtectSystem=strict
ReadWritePaths=/opt/sonatype-work
ProtectHome=true
PrivateTmp=true
# 권한 상승 자체를 봉인: 이 프로세스와 자식은 setuid 로도 권한을 못 올림
NoNewPrivileges=true
# 가질 수 있는 capability 집합을 비움
CapabilityBoundingSet=
```

`NoNewPrivileges=true`가 특히 강력합니다. 이게 켜지면 그 서비스 트리 안에서는 setuid 바이너리나 sudo로도 권한을 올릴 수 없습니다. 즉 위에서 본 셸 탈출이 일어나도 root가 되지 못합니다.

### 네임스페이스와 루트리스 컨테이너

"안에선 root, 밖에선 무력"의 직관에 정확히 대응하는 방법입니다. 컨테이너는 네임스페이스로 프로세스에게 독립된 파일시스템·네트워크·PID 공간을 만들어 줍니다. 특히 루트리스(rootless) Podman에서는 사용자 네임스페이스가 컨테이너 안의 UID 0을 호스트의 비특권 UID로 매핑합니다. 컨테이너 안에서는 root처럼 굴지만 호스트에는 일반 사용자의 흔적만 남습니다. 커널 원리는 [`namespaces.md`](namespaces.md)의 사용자 네임스페이스 절을 보세요.

Nexus는 공식 컨테이너 이미지가 있어서, 운영 환경에서는 이 방식도 흔합니다. 다만 이번 요청은 "계정을 만들어 그 계정이 설치"라는 호스트 설치 시나리오이므로, 아래 레시피는 호스트 설치를 기본으로 하고 컨테이너는 대안으로만 언급합니다.

### MAC SELinux 와 AppArmor

지금까지는 모두 임의 접근 제어(DAC)였습니다. 그 위에 강제 접근 제어(MAC)가 한 겹 더 있습니다. SELinux(RHEL 계열 기본)와 AppArmor(Debian/Ubuntu 계열 기본)는 root조차 정책으로 묶습니다. 정책상 허용되지 않으면 root여도 막힙니다. 강력하지만 정책 작성·디버깅 비용이 큽니다. 대상 서버가 RHEL 계열이면 SELinux가 enforcing일 가능성이 높으므로, 뒤 레시피에서 포트 라벨링 같은 최소 대응만 다룹니다.

### 방법 비교표

| 방법 | 만드는 보장 | 복잡도 | nexus 적합도 |
| :--- | :--- | :--- | :--- |
| 최소 권한 + 소유권 | 비특권이라 남의 영역 접근 불가 | 매우 낮음 | 기본값으로 권장 |
| 좁은 sudo 화이트리스트 | 지정한 명령만 root | 낮음 | 서비스 제어용으로 보조 |
| capabilities | root 권한을 조각 단위로 | 중간 | 8081 쓰면 불필요 |
| systemd 샌드박싱 | 서비스가 볼/쓸 범위 제한 | 중간 | 운영 시 권장 |
| 네임스페이스/컨테이너 | 독립된 공간, 호스트와 분리 | 중상 | 컨테이너 운영 시 |
| SELinux/AppArmor | root까지 정책으로 강제 | 높음 | 기본 정책 유지 정도 |

## 키 기반 로그인 설정과 ssh keygen

이제 nexus 계정 로그인을 공개키 인증으로 설정합니다. 먼저 용어를 정확히 합니다. 앞서 root 접속을 다룰 때 탐지 로그에서 본 `kex`(key exchange)는 **연결 암호화용 세션 키를 만드는 단계**로, 인증 방식과 무관하게 항상 일어납니다. 반면 지금 설정하려는 건 **공개키 인증(public-key authentication)**으로, 사용자가 비밀번호 대신 키로 신원을 증명하는 별개의 단계입니다.

또 하나 자주 헷갈리는 구분이 있습니다. 키에는 방향이 둘 있습니다.

- 호스트 키: 서버가 자기 신원을 클라이언트에게 증명하는 키. 처음 접속할 때 보는 지문(`SHA256:...` 형태의 값)이 이것입니다.
- 사용자 키: 사용자가 자기 신원을 서버에게 증명하는 키. 지금 만들 것이 이쪽입니다.

### 공개키 인증의 동작 원리

공개키 인증은 비대칭 키 한 쌍으로 동작합니다. 개인키(절대 외부로 나가지 않음)와 공개키(서버에 등록)입니다. 로그인 흐름은 다음과 같습니다.

1. 클라이언트가 "이 공개키로 로그인하겠다"고 제안합니다.
2. 서버는 그 공개키가 대상 계정의 `~/.ssh/authorized_keys`에 있는지 확인합니다.
3. 있으면 서버가 임의의 도전값(challenge)을 보냅니다.
4. 클라이언트가 개인키로 그 도전에 서명해 돌려보냅니다.
5. 서버가 등록된 공개키로 서명을 검증합니다. 맞으면 로그인 성공.

비밀번호와 달리 **비밀(개인키)이 네트워크로 전송되지 않습니다.** 서버가 털려도 공개키만 노출되고, 공개키로는 로그인할 수 없습니다. 이게 키 방식이 더 안전한 이유입니다.

### ssh keygen 으로 키 쌍 만들기

클라이언트(여기서는 Windows PC)에서 키를 만듭니다. 서버는 OpenSSH_8.7이고 호스트 키도 ed25519로 관측됐으므로, ed25519를 씁니다. RSA보다 짧고 빠르며 안전합니다.

```sh
ssh-keygen -t ed25519 -C "rody@nexus-login" -f $HOME/.ssh/nexus_ed25519
```

- `-t ed25519`: 키 종류. 아주 오래된 서버만 RSA를 요구하며, 그때만 `-t rsa -b 4096`.
- `-C "..."`: 주석. 키를 사람이 식별하기 위한 라벨일 뿐 보안과 무관합니다.
- `-f ...`: 저장 경로. 개인키가 `nexus_ed25519`, 공개키가 `nexus_ed25519.pub`로 생깁니다. 용도별로 키를 나누면 나중에 폐기·교체가 쉽습니다.

실행하면 패스프레이즈를 묻습니다. **비워 두지 말고 설정하길 권합니다.** 패스프레이즈는 개인키 파일 자체를 한 번 더 암호화해서, 파일이 유출돼도 바로 쓰이지 않게 합니다. 매번 입력하기 번거로우면 ssh-agent에 한 번만 풀어 두면 됩니다.

```text
$ ssh-keygen -t ed25519 -C "rody@nexus-login" -f $HOME/.ssh/nexus_ed25519
Generating public/private ed25519 key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in .../nexus_ed25519
Your public key has been saved in .../nexus_ed25519.pub
The key fingerprint is:
SHA256:....
```

참고로 앞 작업에서 만든 `n`/`n.pub`도 ed25519 키 쌍입니다. 그걸 그대로 nexus 로그인에 재사용해도 동작은 하지만, 용도가 다른 키는 분리하는 편이 관리·폐기에 유리합니다.

Windows에서 한 가지 주의. OpenSSH는 개인키 파일이 다른 사용자에게 열려 있으면 사용을 거부합니다. 리눅스에서는 권한이 너무 열렸다는 뜻의 다음 경고가 뜨고, Windows에서는 ACL 기준으로 같은 검사를 합니다.

```text
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@         WARNING: UNPROTECTED PRIVATE KEY FILE!          @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
Permissions for 'nexus_ed25519' are too open.
```

Windows PowerShell에서 현재 사용자에게만 권한을 남기는 방법입니다.

```powershell
icacls "$env:USERPROFILE\.ssh\nexus_ed25519" /inheritance:r /grant:r "$($env:USERNAME):R"
```

### 서버의 nexus 계정에 공개키 등록하기

이제 공개키를 서버의 nexus 계정에 심습니다. nexus는 아직 비밀번호 로그인이 안 되도록 잠가 둘 것이므로, 등록은 **root 세션에서** 직접 해 줍니다(다음 레시피의 계정 생성 단계와 이어집니다).

리눅스라면 보통 `ssh-copy-id`로 끝나지만, 그건 대상 계정에 비밀번호 로그인이 열려 있을 때 편한 방식입니다. 서비스 계정은 비밀번호 로그인을 막는 게 정석이라, root가 파일을 직접 놓는 방법을 기본으로 합니다.

```sh
# (서버에서 root 로) nexus 의 .ssh 준비
install -d -m 700 -o nexus -g nexus /home/nexus/.ssh

# 클라이언트에서 복사해 온 공개키 한 줄을 authorized_keys 에 기록
#   - 공개키 내용은 'ssh-ed25519 AAAA... rody@nexus-login' 형태의 한 줄
cat >> /home/nexus/.ssh/authorized_keys <<'EOF'
ssh-ed25519 AAAA...여기에_공개키_한_줄... rody@nexus-login
EOF

chown nexus:nexus /home/nexus/.ssh/authorized_keys
chmod 600 /home/nexus/.ssh/authorized_keys
```

권한이 핵심입니다. sshd는 StrictModes 검사로 다음을 요구합니다. 어기면 키가 무시되고 로그인이 실패합니다.

- 홈 디렉터리가 그룹·기타 쓰기 불가
- `~/.ssh`는 700
- `~/.ssh/authorized_keys`는 600, 소유자는 그 계정 본인

실패하면 클라이언트는 계속 비밀번호를 묻거나 `Permission denied (publickey)`를 내고, 서버 로그(`/var/log/secure` 또는 `journalctl -u sshd`)에 `Authentication refused: bad ownership or modes for directory ...`가 남습니다. 이 로그가 권한 문제의 결정적 신호입니다.

클라이언트에서 공개키를 서버로 옮길 때, Windows PowerShell이라면 이렇게 한 줄로 보낼 수 있습니다(이때는 root 또는 비밀번호 가능한 계정으로 한 번 접속). 서비스 계정 흐름에서는 위의 root 직접 등록을 권장합니다.

```powershell
Get-Content "$env:USERPROFILE\.ssh\nexus_ed25519.pub" | ssh root@<host> "cat >> /home/nexus/.ssh/authorized_keys"
```

### 로그인과 ssh config 별칭

이제 키로 접속합니다.

```sh
ssh -i $HOME/.ssh/nexus_ed25519 nexus@<host>
```

매번 `-i`와 IP를 치기 번거로우면 `~/.ssh/config`에 별칭을 둡니다.

```conf
Host nexus-server
    HostName <host>
    User nexus
    IdentityFile ~/.ssh/nexus_ed25519
    IdentitiesOnly yes
```

이러면 `ssh nexus-server`로 끝납니다. `IdentitiesOnly yes`는 등록된 다른 키들을 마구 시도하지 않고 이 키만 쓰게 해서, 키가 많을 때 생기는 인증 실패를 줄입니다.

### 비밀번호 로그인 잠그기

서비스 계정은 비밀번호 로그인을 막는 게 안전합니다. 두 층위가 있습니다.

- 계정 자체의 비밀번호 잠금: `passwd -l nexus` 또는 처음부터 비밀번호를 설정하지 않음. 이러면 비밀번호로는 못 들어오지만 키 로그인과 `sudo -iu nexus`(root가 전환)는 됩니다.
- sshd 전역 정책: `/etc/ssh/sshd_config`에서 `PasswordAuthentication no`. 단 이건 서버 전체에 영향을 주므로, 본인 root 접속까지 키로 바꾼 뒤에 적용해야 스스로 잠기는 사고를 피합니다.

## 권장 설계 nexus 와 maven 레시피

지금까지의 원리를 실제 절차로 묶습니다. 명령은 대상 서버(사내 GitLab 호스트)의 **root 세션에서 실행**한다고 가정합니다. 비밀번호 로그인이 대화형이라 명령을 대신 실행해 주지는 못하므로, 각 단계의 기대 결과와 실패 신호를 함께 적습니다.

관측된 사실과 추론을 분리합니다. 이 서버에서 직접 관측된 것은 SSH가 OpenSSH_8.7이고 publickey와 password 인증을 허용한다는 점입니다. 배포판이 RHEL 9 계열이라는 건 OpenSSH 8.7이 그 계열에서 출하됐다는 사실에 근거한 추론일 뿐이므로, 첫 단계에서 반드시 확인합니다.

### 환경 확인

```sh
cat /etc/os-release          # 배포판과 버전 (dnf 계열인지 apt 계열인지)
getenforce                   # SELinux 상태 (Enforcing/Permissive/Disabled)
java -version                # 시스템에 JDK 가 이미 있는지
firewall-cmd --state         # firewalld 사용 여부
```

이 출력으로 뒤 단계의 패키지 매니저, SELinux 대응 여부, 방화벽 명령이 갈립니다.

### 계정과 작업 공간 생성

```sh
# RHEL 계열에서 adduser 는 useradd 의 심볼릭 링크라, Debian 의 대화형 adduser 와 다릅니다.
useradd -m -d /home/nexus -s /bin/bash nexus
passwd -l nexus              # 비밀번호 로그인 잠금 (키로만 접속)

# 작업 공간을 nexus 소유로. 이 소유권이 '공간 경계'의 실체입니다.
install -d -o nexus -g nexus /opt/nexus /opt/sonatype-work
```

기대 결과: `id nexus`가 새 UID를 보여주고, `ls -ld /opt/nexus`의 소유자가 nexus입니다. 이 시점에서 nexus는 이미 `/etc` 등 남의 영역을 못 건드립니다(별도 설정 불필요).

### 키 로그인 설정

앞 절의 "서버의 nexus 계정에 공개키 등록하기"를 그대로 적용합니다. 등록 후 클라이언트에서 `ssh -i ~/.ssh/nexus_ed25519 nexus@<host>`로 비밀번호 없이 들어오면 성공입니다.

### 사용자 공간 설치 (root 불필요)

여기서부터는 nexus 계정으로 전환해 진행합니다(`sudo -iu nexus` 또는 키 로그인). 모든 설치가 nexus 소유 디렉터리 안에서 일어나므로 root가 필요 없습니다.

JDK와 Maven은 SDKMAN으로 받으면 전부 `~/.sdkman` 아래에 들어가, 시스템을 전혀 건드리지 않습니다.

```sh
# nexus 로 로그인한 상태
curl -s "https://get.sdkman.io" | bash
source "$HOME/.sdkman/bin/sdkman-init.sh"

sdk install java        # 적절한 LTS JDK
sdk install maven
java -version && mvn -version    # 둘 다 버전이 찍히면 성공
```

Nexus는 배포 tarball을 받아 `/opt/nexus`에 풀고 실행합니다.

```sh
# 버전 번호는 Sonatype 다운로드 페이지에서 최신 안정판 확인
cd /opt/nexus
curl -L -O https://download.sonatype.com/nexus/3/nexus-<버전>-unix.tar.gz
tar xzf nexus-<버전>-unix.tar.gz

# 데이터 디렉터리를 /opt/sonatype-work 로 분리 (기본 레이아웃과 동일)
./nexus-<버전>/bin/nexus run      # 포그라운드로 먼저 떠 보기
```

기대 결과: 로그에 `Started Sonatype Nexus`가 뜨고 `curl -I http://localhost:8081`이 응답합니다. 첫 기동 시 초기 admin 비밀번호가 `/opt/sonatype-work/nexus3/admin.password`에 생깁니다.

확인이 필요한 지점: 받는 Nexus 버전이 자체 JRE를 번들하는지, 아니면 외부 JDK를 요구하는지는 버전마다 다릅니다. tarball 안에 `jre/`가 있는지, 릴리스 노트의 자바 요구 버전을 함께 확인합니다. 이 부분은 버전 의존적이라 단정하지 않습니다.

### root 가 꼭 필요한 일만 (최소 위임)

부팅 자동 시작, 파일 디스크립터 상한, 방화벽은 root가 필요합니다. systemd 유닛으로 한 번에 처리하고, 동시에 앞에서 본 샌드박싱을 입힙니다.

```ini
# /etc/systemd/system/nexus.service   (root 로 작성)
[Unit]
Description=Sonatype Nexus Repository
After=network.target

[Service]
Type=forking
User=nexus
Group=nexus
LimitNOFILE=65536
ExecStart=/opt/nexus/nexus-<버전>/bin/nexus start
ExecStop=/opt/nexus/nexus-<버전>/bin/nexus stop
Restart=on-abort

# 샌드박싱: 쓰기는 데이터 디렉터리에만, 나머지는 읽기 전용
ProtectSystem=strict
ReadWritePaths=/opt/sonatype-work
ProtectHome=true
PrivateTmp=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

```sh
systemctl daemon-reload
systemctl enable --now nexus
firewall-cmd --add-port=8081/tcp --permanent && firewall-cmd --reload   # firewalld 쓰는 경우
```

SELinux가 Enforcing이고 8081이 표준 라벨에 없어 거부되면, 포트 라벨을 추가합니다. 막혔는지 여부는 `ausearch -m avc -ts recent`로 확인합니다.

```sh
semanage port -a -t http_port_t -p tcp 8081
```

서비스 제어를 nexus 계정 스스로 하게 하려면, 그때만 앞 절의 좁은 sudoers 드롭인을 추가합니다. 운영을 관리자가 한다면 이 위임도 생략하는 게 가장 안전합니다.

위 유닛은 디렉티브의 의미를 보이는 예시입니다. 한 가지 주의가 있습니다. 자바를 nexus 홈의 SDKMAN에 설치하면 `ProtectHome=true`가 그 자바까지 가려서 서비스가 뜨지 못합니다. 그래서 홈을 보호하려면 자바를 홈 밖(시스템 JDK)에 두거나, `Environment=INSTALL4J_JAVA_HOME=...`로 자바 위치를 명시해야 합니다. 실제로 동작하는 전체 구성은 아래 실제 사례 절에 있습니다.

## 경계가 지켜지는지 검증하기

설계가 "공간 밖에선 무력"을 실제로 만족하는지 직접 확인합니다. 이건 설명이 아니라 손으로 돌려 보는 단계입니다.

```sh
# 1) nexus 의 sudo 권한이 의도한 것만인지 (root 로 조회)
sudo -l -U nexus
#   기대: 위임한 systemctl 몇 줄만 보이거나, 아무것도 없음

# 2) 공간 밖 쓰기가 거부되는지 (nexus 로)
touch /etc/should-fail        # 기대: Permission denied
touch /opt/sonatype-work/ok   # 기대: 성공

# 3) 서비스가 root 가 아닌 nexus 로 도는지
ps -o pid,user,comm -C java   # 기대: USER 열이 nexus

# 4) 샌드박싱 노출도 점수
systemd-analyze security nexus
#   기대: 점수가 낮을수록(노출 적음) 좋음. 항목별로 무엇이 열려 있는지 보여줌
```

2번에서 `/etc` 쓰기가 거부되고 `/opt/sonatype-work` 쓰기가 성공하면, "자기 공간에서만"이 DAC 수준에서 성립한다는 직접 증거입니다. 1번에서 위임 목록이 좁게 유지되는지, 4번에서 노출 점수가 낮아지는지가 서비스 레벨 격리의 증거입니다.

## 정리와 teach back

이 문서를 덮고 다음을 자기 말로 설명할 수 있으면 핵심을 잡은 것입니다.

- sudo가 제어하는 축과 제어하지 못하는 축은 각각 무엇인가. 왜 "디렉터리 안에서만 root"가 sudo로 표현되지 않는가.
- "공간 밖에선 권한 없음"을 nexus에게 만들어 주는 실제 메커니즘은 무엇인가. 왜 그게 추가 설정이 아니라 비특권 사용자의 기본 성질인가.
- 키 교환(KEX)과 공개키 인증은 어떻게 다른가. 공개키 인증에서 네트워크로 전송되지 않는 것은 무엇인가.
- `~/.ssh`와 `authorized_keys`의 권한이 틀렸을 때 어떤 신호가 어디에 남는가.

작은 실험 하나로 굳힐 수 있습니다. 테스트 사용자를 만들어 `NOPASSWD: /usr/bin/systemctl --no-pager status sshd`만 위임한 뒤, 그 사용자로 `/etc`에 파일을 만들어 거부되는지, 그리고 위임된 status 명령은 되는지 확인해 보세요. 권한 상승과 격리가 서로 다른 축이라는 게 한 번에 체감됩니다.

## 실제 사례

앞의 권장 설계 레시피가 원리와 판단을 설명했다면, 이 절은 그대로 복사해 실행하는 런북입니다. 역할이 다릅니다. 목표는 `nexus` 계정을 베스트 프랙티스로 만들어, Nexus와 Maven을 그 계정의 공간 안에서 설치·운영하고, 로그인은 공개키로만 하는 것입니다.

설계 의도를 먼저 고정합니다. 자바·Maven·Nexus를 모두 nexus 계정의 홈과 `/opt` 아래에 두어 이 계정이 자기 공간에서 자급자족하게 합니다. root가 필요한 일은 부팅 자동 시작(systemd)과 방화벽뿐이고, 서비스 제어 위임(sudoers)은 선택입니다.

전제와 주의는 다음과 같습니다.

- 서버에 root로 접속한 상태에서 시작합니다. 배포판은 dnf 계열(RHEL/Rocky/Alma 9 추정)로 가정하며, 다른 계열이면 방화벽·SELinux 명령만 바뀝니다.
- 다운로드 단계는 서버에서 외부 인터넷이 나가야 합니다. 폐쇄망이면 SDKMAN 대신 tarball을 미리 받아 옮깁니다.
- 1번부터 4번 블록은 같은 root 세션에서 차례로 실행하면 거기서 정한 변수가 유지됩니다. 5번은 nexus 세션, 6번은 다시 root입니다.

### 0 클라이언트에서 키 생성

클라이언트(Windows PowerShell 등)에서 한 번만 합니다. 이미 만든 `n`/`n.pub`를 재사용해도 동작하지만, 용도별로 키를 나누는 편이 폐기·교체에 유리합니다.

```sh
ssh-keygen -t ed25519 -C "rody@nexus-login" -f $HOME/.ssh/nexus_ed25519
cat $HOME/.ssh/nexus_ed25519.pub      # 출력된 한 줄을 복사 (3번에서 사용)
```

### 1 그룹과 계정 생성

서버 root에서 실행합니다.

```sh
# --- 변수: 한 번만 정하면 4번까지 재사용 ---
NEXUS_USER=nexus
NEXUS_HOME=/home/nexus
APP_DIR=/opt/nexus
DATA_DIR=/opt/sonatype-work

# 전용 그룹 (없을 때만 생성)
getent group "$NEXUS_USER" >/dev/null || groupadd "$NEXUS_USER"

# 전용 사용자: 홈 생성(-m), 로그인 셸 bash(-s), 주 그룹 지정(-g)
#   관리 위해 로그인 가능한 계정으로 만들고 인증은 키로만 합니다.
#   순수 데몬 전용이면 'useradd -r ... -s /sbin/nologin' 으로 더 잠급니다.
id "$NEXUS_USER" >/dev/null 2>&1 || \
  useradd -m -d "$NEXUS_HOME" -s /bin/bash -g "$NEXUS_USER" "$NEXUS_USER"

# 비밀번호 로그인 잠금 (키 인증만)
passwd -l "$NEXUS_USER"

# 홈을 본인만 접근하도록 (sshd StrictModes 도 충족)
chmod 750 "$NEXUS_HOME"

id "$NEXUS_USER"        # 확인: uid/gid/그룹 출력
```

### 2 작업 공간 디렉터리

서버 root에서 실행합니다. 이 소유권이 작업 공간 경계의 실체입니다.

```sh
install -d -m 750 -o "$NEXUS_USER" -g "$NEXUS_USER" "$APP_DIR" "$DATA_DIR"
ls -ld "$APP_DIR" "$DATA_DIR"     # 확인: 소유자 열이 nexus
```

### 3 공개키 등록

서버 root에서 실행합니다.

```sh
install -d -m 700 -o "$NEXUS_USER" -g "$NEXUS_USER" "$NEXUS_HOME/.ssh"

# 0번에서 복사한 공개키 한 줄을 아래 따옴표 안에 붙여넣기
PUBKEY='ssh-ed25519 AAAA...여기에_공개키_한_줄... rody@nexus-login'
printf '%s\n' "$PUBKEY" > "$NEXUS_HOME/.ssh/authorized_keys"

chown "$NEXUS_USER:$NEXUS_USER" "$NEXUS_HOME/.ssh/authorized_keys"
chmod 600 "$NEXUS_HOME/.ssh/authorized_keys"
```

확인하려면 클라이언트에서 `ssh -i ~/.ssh/nexus_ed25519 nexus@<host>`가 비밀번호 없이 접속되면 성공입니다. 실패하면 서버 `journalctl -u sshd | tail`의 `bad ownership or modes`가 권한 문제 신호입니다.

### 4 서비스 제어 위임은 좁게

서버 root에서 실행하는 선택 단계입니다. nexus가 스스로 서비스를 켜고 끄게 하려면 이 드롭인만 추가하고, 운영을 관리자가 하면 생략합니다.

```sh
SYSTEMCTL=$(command -v systemctl)     # 보통 /usr/bin/systemctl
cat > /etc/sudoers.d/nexus <<EOF
nexus ALL=(root) NOPASSWD: $SYSTEMCTL start nexus, $SYSTEMCTL stop nexus, $SYSTEMCTL restart nexus, $SYSTEMCTL --no-pager status nexus
EOF
chmod 0440 /etc/sudoers.d/nexus
visudo -cf /etc/sudoers.d/nexus       # 'parsed OK' 출력되어야 함
```

여기 heredoc은 `$SYSTEMCTL`을 확장해야 하므로 따옴표 없는 `EOF`를 씁니다. 3번 공개키 블록이 따옴표 있는 `'EOF'`를 쓴 것과 반대인데, 그쪽은 키 안의 문자를 그대로 둬야 하기 때문입니다.

### 5 자바와 maven과 nexus 설치

root에서 `su - nexus`로 전환하거나, 방금 설정한 키로 nexus에 로그인한 뒤 실행합니다. 전부 nexus 소유 공간에 설치되어 root가 필요 없습니다.

```sh
# JDK + Maven: 모두 ~/.sdkman 아래에 설치 (시스템 무수정)
curl -s "https://get.sdkman.io" | bash
source "$HOME/.sdkman/bin/sdkman-init.sh"
sdk install java          # Nexus 가 요구하는 LTS 로 (버전은 릴리스 노트 확인)
sdk install maven
java -version && mvn -version        # 둘 다 버전이 찍히면 성공

# Nexus tarball
NEXUS_VERSION=3.xx.y-zz              # 다운로드 페이지의 최신 안정판으로 교체
cd /opt/nexus
curl -L -O "https://download.sonatype.com/nexus/3/nexus-${NEXUS_VERSION}-unix.tar.gz"
tar xzf "nexus-${NEXUS_VERSION}-unix.tar.gz"
ln -sfn "/opt/nexus/nexus-${NEXUS_VERSION}" /opt/nexus/current    # 버전 독립 경로

# 포그라운드로 한 번 확인 (Ctrl+C 로 종료)
/opt/nexus/current/bin/nexus run
```

기대 결과는 로그의 `Started Sonatype Nexus`와, 다른 터미널에서 `curl -I http://localhost:8081`의 응답입니다. 첫 기동 시 초기 비밀번호가 `/opt/sonatype-work/nexus3/admin.password`에 생깁니다.

### 6 부팅 자동 시작과 방화벽

서버 root에서 실행합니다. 자바를 nexus 홈의 SDKMAN에 두었으므로 유닛에서 그 위치를 직접 지정하고, 홈을 가리는 `ProtectHome`은 켜지 않습니다.

```ini
# /etc/systemd/system/nexus.service
[Unit]
Description=Sonatype Nexus Repository
After=network.target

[Service]
Type=forking
User=nexus
Group=nexus
LimitNOFILE=65536
TimeoutSec=600
Environment=INSTALL4J_JAVA_HOME=/home/nexus/.sdkman/candidates/java/current
ExecStart=/opt/nexus/current/bin/nexus start
ExecStop=/opt/nexus/current/bin/nexus stop
Restart=on-abort
ProtectSystem=full
PrivateTmp=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

```sh
systemctl daemon-reload
systemctl enable --now nexus
systemctl --no-pager status nexus        # active (running) 확인

firewall-cmd --add-port=8081/tcp --permanent && firewall-cmd --reload   # firewalld 쓰는 경우
```

SELinux가 Enforcing이고 8081이 거부되면 `semanage port -a -t http_port_t -p tcp 8081`로 라벨을 추가하고, 막힘 여부는 `ausearch -m avc -ts recent`로 확인합니다. 더 강하게 조이려면 `ProtectSystem=strict`에 `ReadWritePaths=/opt/sonatype-work`를 더하되, 변경 후 반드시 재기동해 nexus가 정상 기동하는지 확인합니다.

### 7 경계 검증

```sh
sudo -l -U nexus                      # (root) nexus 가 가진 sudo 가 의도한 것뿐인지
su - nexus -c 'touch /etc/should-fail; touch /opt/sonatype-work/ok; ls -l /opt/sonatype-work/ok'
ps -o pid,user,comm -C java           # USER 열이 nexus 인지
systemd-analyze security nexus        # 샌드박싱 노출도 (낮을수록 좋음)
```

기대 결과는 `/etc/should-fail`이 `Permission denied`로 거부되고, `/opt/sonatype-work/ok`는 생성에 성공하며, `java` 프로세스의 USER가 nexus로 보이는 것입니다. 앞의 두 가지가 자기 공간에서만 권한이 작동한다는 직접 증거입니다.

## Nexus 운영 토폴로지와 데이터베이스

production 수준으로 Nexus를 운영하려면 데이터베이스를 어떻게 둘지부터 정해야 한다. 이 절은 H2와 PostgreSQL의 실제 차이, 그리고 이미 GitLab용 PostgreSQL이 도는 호스트에서 어떻게 배치하는지를 공식 문서 근거와 함께 정리한다. 앞 실제 사례 런북은 설명을 줄이려고 기본 H2로 동작하며, production은 이 절의 PostgreSQL 토폴로지를 따른다.

### H2 와 PostgreSQL 의 실제 차이

먼저 흔한 오해부터 푼다. H2를 쓴다고 재시작 시 데이터가 초기화되지 않는다. Nexus는 H2를 file-based 모드로 디스크에 저장하므로 설정, 저장소, 컴포넌트 메타데이터가 그대로 유지된다. 실제로 사라지는 경우는 H2의 성질이 아니라 데이터 디렉터리가 휘발성일 때(예: 컨테이너에 볼륨 미연결)뿐이다.

진짜 차이는 한도와 내구성과 확장성이다. 공식 문서 기준으로 정리하면 다음과 같다.

- H2는 신규 설치의 기본 임베디드 DB이고, 문서가 제시하는 한도는 하루 20만 요청 또는 컴포넌트 10만 개다. 그 이상 워크로드는 지원되지 않는다.
- H2는 비정상 종료, 정전, 대용량에서 데이터 손상 위험이 높아지고, 고가용성(HA)을 지원하지 않으며, 컨테이너 기반 배포도 지원하지 않는다.
- Sonatype의 포괄 권장은 모든 배포에 외부 PostgreSQL을 쓰라는 것이며, 미션 크리티컬이나 대규모는 PostgreSQL이다.
- 무료 Community Edition에서 PostgreSQL은 3.77.0 이후 릴리스부터 가능하다. 그 이전 무료 에디션은 H2만 쓴다.
- 버전은 PostgreSQL이 현재 지원하는 메이저면 되고, 레퍼런스 아키텍처는 14 이상을 권장한다.

정리하면 H2는 소규모, 단일 팀, VM 단일 노드, 위 한도 이내에서 공식 지원되는 옵션이고, 그 밖의 경우(미션 크리티컬, 대규모, HA, 컨테이너)는 PostgreSQL이다.

### 배포 토폴로지 어디에 둘 것인가

Sonatype 공식 문서는 PostgreSQL 설치 방법을 처방하지 않고 PostgreSQL 사이트 링크만 준다. 배포 토폴로지가 운영 선택이기 때문이다. 다만 레퍼런스 아키텍처는 방향을 분명히 보여준다. 단일 노드 Nexus라도 PostgreSQL은 자체 자원을 가진 외부 전용 컴포넌트로 분리한다.

그래서 핵심 원칙은 하나다. Nexus의 PostgreSQL은 라이프사이클을 직접 통제하는 전용 인스턴스로, 앱과도 다른 서비스와도 분리해 둔다. 우선순위는 다음과 같다.

1. 별도 호스트 또는 관리형 PostgreSQL

    레퍼런스 아키텍처의 external database가 이것이다. 독립 백업, 튜닝, HA가 가능해 production 1순위다.

2. 같은 호스트의 루트리스 컨테이너

    같은 박스에 둬야 하면 전용 포트와 볼륨을 가진 컨테이너로 띄운다. 호스트와 격리되고 버전을 직접 고르며 무권한으로 돈다. 단일 호스트에서 가장 깔끔하다.

3. dnf 시스템 PostgreSQL을 비표준 포트에

    전통적 방식이다. 동작하지만 일회성 root가 들고 OS와 더 엉킨다.

컨테이너에는 한 가지 함의가 있다. H2는 컨테이너 배포가 지원되지 않으므로, Nexus를 컨테이너로 띄우면 자동으로 PostgreSQL이 강제된다. 즉 Nexus 컨테이너와 PostgreSQL 컨테이너 조합은 production용 DB와 최소 권한을 동시에 만족하는 일관된 선택이다.

### 이미 GitLab PostgreSQL 이 도는 호스트라면

GitLab Omnibus는 자체 PostgreSQL을 번들해 `gitlab-psql` 계정으로 돌린다. 이 인스턴스는 재사용하지 않는다. gitlab-ctl이 관리하고 reconfigure가 수동 변경을 덮어쓰며, 버전이 GitLab 업그레이드에 묶이고, 장애가 형상관리와 CI로 전파되기 때문이다.

포트 충돌을 걱정할 수 있는데, GitLab Omnibus PostgreSQL은 기본적으로 Unix 소켓 전용이라 TCP를 열지 않는 경우가 많다. `listen_addresses`가 빈 값이면 TCP로는 듣지 않는다는 뜻이고(포트 번호 5432는 소켓 파일 이름과 기본값일 뿐이다), 다음으로 실제 점유를 확인할 수 있다.

```sh
gitlab-psql -c 'SHOW listen_addresses;'   # 빈 값이면 TCP 미청취 (소켓 전용)
ss -ltnp | grep -E ':(5432|5433)'         # 실제 TCP 점유 확인
```

그래도 Nexus용 PostgreSQL은 전용 포트(예: 5433)와 localhost 바인드로 두는 편이 혼동을 없앤다. TCP와 Unix 소켓이라는 두 접속 통로, `listen_addresses`의 의미, 같은 호스트 다중 인스턴스 공존 규칙의 일반 원리는 [database/postgresql/connections.md](../database/postgresql/connections.md)에 정리되어 있다.

### PostgreSQL 준비와 Nexus 연결

공식 가이드가 요구하는 사전 객체와 연결 설정이다. 사전에 만들 객체는 전용 role(LOGIN과 비밀번호), 그 role을 소유자로 하는 UTF-8 데이터베이스, 그리고 트라이그램 확장 `pg_trgm`이다. 문서는 DB 유저가 데이터베이스 소유자여야 한다고 명시한다. 업그레이드나 스키마 변경이 소유 권한을 요구하기 때문이며, 비소유 유저는 지원되지 않는다.

```sql
CREATE ROLE nexus WITH LOGIN PASSWORD 'change-me-strong';
CREATE DATABASE nexus OWNER nexus ENCODING 'UTF8';
\connect nexus
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

Nexus에는 데이터 디렉터리 아래 설정 파일로 접속 정보를 준다.

```properties
# <data-dir>/etc/fabric/nexus-store.properties
username=nexus
password=change-me-strong
jdbcUrl=jdbc:postgresql://127.0.0.1:5433/nexus
```

또는 환경변수 `NEXUS_DATASTORE_NEXUS_JDBCURL`, `NEXUS_DATASTORE_NEXUS_USERNAME`, `NEXUS_DATASTORE_NEXUS_PASSWORD`로도 된다.

### Maven 의 위치

흔한 혼동 하나를 짚는다. Maven은 Nexus 서버에 설치하는 구성 요소가 아니다. Nexus는 저장소 서버이고, Maven은 개발자 PC와 CI 러너에서 돌며 `settings.xml`로 이 Nexus를 미러나 프록시로 가리킬 뿐이다. 그래서 서버 쪽 스택은 Nexus와 PostgreSQL, 그리고 선택적 리버스 프록시가 전부다.

## 근거와 더 읽을거리

- sudo의 정책 모델과 옵션: `man sudoers`, `man sudo`. 저장소 내 [`commands/sudo.md`](commands/sudo.md), [`commands/su.md`](commands/su.md).
- 사용자 생성 옵션: `man useradd`. 저장소 내 [`commands/useradd.md`](commands/useradd.md).
- 파일 권한과 setuid: `man chmod`, `man 7 credentials`.
- Linux capabilities: `man 7 capabilities`.
- systemd 샌드박싱 디렉티브: `man systemd.exec`, `systemd-analyze security`. 저장소 내 [`systemd/systemd.md`](systemd/systemd.md).
- 네임스페이스와 사용자 네임스페이스: `man 7 namespaces`, `man 7 user_namespaces`. 저장소 내 [`namespaces.md`](namespaces.md).
- SSH 공개키 인증과 키 생성: `man ssh-keygen`, `man sshd`(AUTHORIZED_KEYS, StrictModes). 저장소 내 [`ssh.md`](ssh.md), [`../troubleshooting/ssh.md`](../troubleshooting/ssh.md).
- Nexus 설치와 자바 요구 버전: Sonatype 공식 문서(버전별로 다르므로 설치 시점 릴리스 노트 확인).
- Nexus 데이터베이스 옵션과 H2 한도, PostgreSQL 권장: [database-options](https://help.sonatype.com/en/database-options.html), [system requirements](https://help.sonatype.com/en/sonatype-nexus-repository-system-requirements.html).
- PostgreSQL 구성(소유자 role, pg_trgm, nexus-store.properties): [Install with PostgreSQL](https://help.sonatype.com/en/install-nexus-repository-with-a-postgresql-database.html).
- 배포 토폴로지(외부 전용 PostgreSQL, 14+ 권장): [Reference Architecture 2](https://help.sonatype.com/en/nexus-repository-reference-architecture-2.html).

근거 수준 메모. sudo의 축 구성, setuid 동작, DAC 거부, 공개키 인증 흐름, systemd 디렉티브 의미는 man page와 직접 관측으로 닫히는 사실입니다. 대상 서버가 RHEL 9 계열이라는 것과 Nexus의 자바 번들 여부는 추론 또는 버전 의존이라 본문에서 확인 단계로 남겨 두었습니다.
