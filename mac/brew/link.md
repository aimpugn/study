# link

- [link](#link)
    - [Homebrew의 `brew link` 명령어](#homebrew의-brew-link-명령어)
    - [`brew link` 명령어의 기본 사용법](#brew-link-명령어의-기본-사용법)
    - [주요 옵션](#주요-옵션)
    - [`brew link`의 역할](#brew-link의-역할)
    - [`brew link`와 `PATH`의 관계](#brew-link와-path의-관계)
    - [주의사항](#주의사항)
    - [예제](#예제)
        - [`link` 옵션 사용 예](#link-옵션-사용-예)
        - [`unlink` 옵션 사용 예](#unlink-옵션-사용-예)
        - [Warning: Already linked 라며 덮어쓰지 않는 경우](#warning-already-linked-라며-덮어쓰지-않는-경우)

## Homebrew의 `brew link` 명령어

Homebrew는 macOS에서 다양한 소프트웨어 패키지를 설치하고 관리하는 도구입니다.
`brew link` 명령어는 설치된 패키지의 파일과 디렉터리를 시스템의 표준 위치에 심볼릭 링크(Symbolic Link, 또는 Symlink)를 생성하는 역할을 합니다.

## `brew link` 명령어의 기본 사용법

```bash
brew link [formula_name] [options]
```

- `formula_name`: 링크할 Homebrew 패키지의 이름입니다.
- `[options]`: 링크 동작을 제어하는 옵션입니다.

## 주요 옵션

- `--overwrite`: 기존에 존재하는 파일을 삭제하면서 링크를 생성합니다.
- `-n`, `--dry-run`: 실제로 파일을 링크하거나 삭제하지 않고, 어떤 파일들이 링크되거나 삭제될 것인지 보여줍니다.
- `-f`, `--force`: Keg-only formula를 링크할 수 있게 합니다.
- `--HEAD`: 설치된 formula의 HEAD 버전을 링크합니다.

## `brew link`의 역할

주로 실행 파일, 라이브러리, 헤더 파일 등을 링크하여 쉽게 실행할 수 있도록 합니다.

설치된 패키지의 파일과 디렉터리를 다음과 같은 시스템의 표준 위치에 심볼릭 링크를 생성합니다.
- `/usr/local/bin`
- `/usr/local/lib`
- `/usr/local/include`

예를 들어, Homebrew로 PHP를 설치하면 실제 파일은 `/usr/local/Cellar/php/{version}` 디렉터리에 위치하지만,

`brew link` 명령을 사용하면 `/usr/local/bin/php`와 같은 위치에 심볼릭 링크를 생성하여 php 명령을 쉽게 실행할 수 있게 됩니다.

단, Homebrew는 Apple Silicon Mac에서 `/usr/local` 대신 `/opt/homebrew` 디렉토리를 기본 설치 경로로 사용합니다.
이는 `/usr/local` 경로가 기존의 x86_64 환경에서 사용되기 때문에 Apple Silicon (ARM) 환경과 충돌을 피하기 위함입니다.

## `brew link`와 `PATH`의 관계

`brew link` 명령을 사용하면, `PATH` 환경 변수에 나열된 디렉터리 중 하나에 심볼릭 링크를 생성합니다.

예를 들어, `/usr/local/bin`이 `PATH` 환경 변수에 포함되어 있고 `brew link php` 명령을 실행하면,
`/usr/local/bin/php` 심볼릭 링크가 생성되어 php 명령어를 쉽게 실행할 수 있게 됩니다.

Apple Silicon Mac에서 `brew link` 명령을 사용할 때 패키지의 심볼릭 링크는 `/usr/local/bin` 대신 `/opt/homebrew/bin`에 생성됩니다.

심볼릭 링크 생성 경로:
- **Intel 기반 Mac**: `/usr/local/bin`
- **Apple Silicon 기반 Mac**: `/opt/homebrew/bin`

따라서 PATH 환경 변수에 `/opt/homebrew/bin` 디렉터리가 포함되어 있지 않다면, 시스템은 해당 디렉터리에 있는 명령어를 찾지 못하게 됩니다.

```shell
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 주의사항

이미 링크된 패키지는 `brew link` 명령을 통해 덮어쓰지 않습니다.
이미 링크된 패키지를 재연결하려면 먼저 `brew unlink` 명령을 사용해야 합니다.

`brew link` 명령을 사용할 때는 패키지가 이미 설치되어 있는지 확인하고,
필요한 경우 `--overwrite` 옵션을 사용하여 기존 파일을 삭제하면서 링크를 생성할 수 있습니다.

## 예제

### `link` 옵션 사용 예

```bash
brew link --force php@8.1
```

이 명령은 PHP 8.1을 강제로 링크합니다. 이후 PHP 8.1을 사용하려면 `PATH` 환경 변수에 `/opt/homebrew/opt/php@8.1/bin`과 `/opt/homebrew/opt/php@8.1/sbin`이 포함되어 있어야 합니다.

### `unlink` 옵션 사용 예

```bash
brew unlink php@5.6
```

이 명령은 PHP 5.6의 심볼릭 링크를 제거합니다.

### Warning: Already linked 라며 덮어쓰지 않는 경우

```bash
brew link --force --overwrite "php@8.1"
```

이 명령은 PHP 8.1 버전을 링크하려고 시도하지만, 이미 PHP 8.1이 링크되어 있기 때문에 아무런 작업도 수행하지 않습니다. 이 경우, 먼저 `brew unlink` 명령을 사용하여 기존 링크를 제거한 다음, 다시 `brew link` 명령을 실행해야 합니다.

```bash
brew unlink php@8.1 && brew link --force php@8.1
```

이 명령은 먼저 PHP 8.1의 심볼릭 링크를 제거하고, 그 다음에 다시 링크를 생성합니다.
