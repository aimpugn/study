# systemd

- [systemd](#systemd)
    - [Synopsis](#synopsis)
    - [systemd.generator](#systemdgenerator)

## Synopsis

```plaintext
/path/to/generator normal-dir [early-dir] [late-dir]

/run/systemd/system-generators/*
/etc/systemd/system-generators/*
/usr/local/lib/systemd/system-generators/*
/usr/lib/systemd/system-generators/*
/run/systemd/user-generators/*
/etc/systemd/user-generators/*
/usr/local/lib/systemd/user-generators/*
/usr/lib/systemd/user-generators/*
```

## [systemd.generator](https://www.freedesktop.org/software/systemd/man/latest/systemd.generator.html)

generator는 `/usr/lib/systemd/system-generators/` 및 위에 나열된 기타 디렉토리에 있는 작은 실행 파일입니다.
[`systemd(1)`](https://www.freedesktop.org/software/systemd/man/latest/systemd.html#)는 부팅 시와 구성 리로드 시 유닛 파일이 로드되기 전에 이러한 바이너리를 매우 일찍 실행합니다.
이들의 주요 목적은 서비스 매니저에 기본이 아닌 구성 및 실행 컨텍스트 매개변수를 동적으로 생성된 유닛 파일, 심볼릭 링크 또는 유닛 파일 드롭인으로 변환하여 서비스 매니저가 나중에 로드하고 작동하는 유닛 파일 계층 구조를 확장하는 것입니다.

`systemd`는 generator 출력에 사용할 세 개의 디렉토리 경로로 각 generator를 호출합니다.
이 세 디렉토리에서 generator는 유닛 파일(일반 파일, 인스턴스, 템플릿), 유닛 파일 `.d/` 드롭인을 동적으로 생성하고 유닛 파일에 대한 심볼릭 링크를 생성하여 추가 종속성을 추가하거나 별칭을 만들거나 기존 템플릿을 인스턴스화할 수 있습니다.
이러한 디렉토리는 단위 로드 경로(unit load path)에 포함되므로 생성된 구성이 기존 정의를 확장하거나 재정의할 수 있습니다.
테스트의 경우 하나의 argument로 generator가 호출할 수 있으며, 이 경우 generator는 세 경로가 모두 동일하다고 가정해야 합니다.

generator 출력의 디렉토리 경로는 우선순위에 따라 다릅니다.
- `.../generator.early`는 `/etc/`의 관리자 구성보다 우선순위가 높고,
- `.../generator`는 `/etc/`보다 우선순위가 낮지만 `/usr/`의 벤더 구성보다 우선순위가 높으며,
- `.../generator.late`는 다른 모든 구성보다 우선순위가 낮습니다.

다음 섹션과 systemd.unit(5)의 단위 로드 경로 및 단위 재정의에 대한 설명을 참조하세요.

generator는 위에 나열된 대로 컴파일 중에 결정된 경로 집합에서 로드됩니다.
시스템 및 사용자 generator는 이름이 각각 `system-generators/` 및 `user-generators/`로 끝나는 디렉토리에서 로드됩니다.
앞서 나열된 디렉토리에 있는 generator는 목록 아래 디렉토리에 있는 같은 이름의 generator를 재정의합니다.
`/dev/null` 또는 빈 파일에 대한 심볼릭 링크를 사용하여 generator를 마스킹하여 실행되지 않도록 할 수 있습니다.
단위 로드 경로(unit load path)와 관련하여 우선순위가 가장 높은 두 디렉토리의 순서가 뒤바뀌며 `/run/`의 generator가 `/etc/`의 generator를 덮어씁니다.

새 generator를 설치하거나 구성을 업데이트한 후 `systemctl daemon-reload`가 실행될 수 있습니다. 그러면:
- generator가 생성한 이전 구성이 삭제되고
- 모든 generator가 다시 실행되며
- `systemd`가 디스크에서 유닛을 다시 로드합니다.

자세한 내용은 [`systemd(1)`](https://www.freedesktop.org/software/systemd/man/latest/systemd.html#)을 참조하세요.

> `.d/` 드롭인
>
> TL;DR
> **systemd**에서 유닛 파일(예: 서비스, 타이머, 스냅샷 등)의 구성을 확장하거나 수정하기 위해 사용하는 디렉토리입니다.
> 이 디렉토리 내에 추가 설정 파일을 배치하여 기존 유닛 파일의 동작을 오버라이드하거나 보완할 수 있습니다.
>
> `systemd`에서 각 유닛(예: `ssh.service`)은 일반적으로 `/usr/lib/systemd/system/` 디렉토리에 정의된 기본 유닛 파일로 구성됩니다. 그러나 경우에 따라 사용자는 이 기본 설정을 수정하거나 추가 설정을 적용하고 싶을 수 있습니다. 이를 위해 직접 유닛 파일을 수정하는 대신, **`.d/` 드롭인 디렉토리**를 사용하여 보다 유연하게 설정을 변경할 수 있습니다.
>
> 각 유닛 파일은 해당 유닛 파일 이름에 `.d`를 추가한 디렉토리를 가질 수 있습니다. 이 디렉토리 내에는 `override.conf`와 같은 파일을 추가할 수 있으며, 이 파일은 원래 유닛 파일의 설정을 오버라이드하거나 추가 설정을 정의합니다.
>
> 예를 들어, `ssh.service` 유닛 파일에 대해 `/usr/lib/systemd/system/ssh.service.d/` 디렉토리가 존재할 수 있습니다.
>
> 이 방식은 다음과 같은 장점이 있습니다:
> - `.d/` 드롭인은 유닛 파일을 직접 수정하지 않기 때문에, 시스템 업데이트나 패키지 업데이트로 인한 유닛 파일 변경 시에도 사용자의 설정이 유지되는 장점이 있습니다.
> - 또한, 여러 개의 드롭인 파일을 추가할 수 있기 때문에, 복잡한 설정을 여러 파일로 나누어 관리할 수도 있습니다.
