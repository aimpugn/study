# Detect Encoding

## chardetect?

## `uchardet`

`uchardet`는 Mozilla의 Universal Charset Detector 라이브러리를 기반으로 한 도구입니다.
다양한 인코딩을 감지하는 데 사용됩니다.

### 설치

Debian/Ubuntu 기반 시스템에서는 다음 명령어로 설치할 수 있습니다:

```sh
sudo apt-get install uchardet
```

macOS에서는 Homebrew를 사용하여 설치할 수 있습니다:

```sh
brew install uchardet
```

### help

```bash
❯ uchardet --help

uchardet Command Line Tool
Version 0.0.8

Authors: BYVoid, Jehan
Bug Report: https://gitlab.freedesktop.org/uchardet/uchardet/-/issues

Usage:
 uchardet [Options] [File]...

Options:
 -v, --version         Print version and build information.
 -h, --help            Print this help.
```

### 사용 예시

```sh
uchardet filename
```

예를 들어, `example.txt` 파일의 인코딩을 확인하려면:

```sh
uchardet example.txt
```

출력 예시:

```sh
UTF-8
```
