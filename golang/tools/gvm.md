# gvm

- [gvm](#gvm)
    - [설치](#설치)
        - [macos X requirements](#macos-x-requirements)
        - [install](#install)
    - [brew 통해 설치한 go 삭제](#brew-통해-설치한-go-삭제)
    - [사용법](#사용법)
    - [1.5+ 설치시 1.4 먼저 설치](#15-설치시-14-먼저-설치)
        - [설치 가능한 Go 버전 확인](#설치-가능한-go-버전-확인)
        - [설치 가능한 Go 버전 중에서 특정 버전 설치](#설치-가능한-go-버전-중에서-특정-버전-설치)

> [`asdf`](../../linux/commands/asdf.md) 명령어를 사용한다.

## 설치

### macos X requirements

```bash
xcode-select --install
brew update
brew install mercurial
```

### install

```bash
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
```

```log
❯ bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
Cloning from https://github.com/moovweb/gvm.git to /Users/rody/.gvm
macOS detected. User shell is: /bin/zsh
No existing Go versions detected
Installed GVM v1.0.22

Please restart your terminal session or to get started right away run
 `source /Users/rody/.gvm/scripts/gvm`
```

## brew 통해 설치한 go 삭제

```bash
brew uninstall go

# 강제 삭제
brew uninstall --force go

# dependency 있는 경우
brew uninstall --force go --ignore-dependencies
```

## 사용법

## 1.5+ 설치시 1.4 먼저 설치

> Go 1.5+ removed the C compilers from the toolchain and replaced them with one written in Go. Obviously, this creates a bootstrapping problem if you don't already have a working Go install. In order to compile Go 1.5+, make sure Go 1.4 is installed first. If Go 1.4 won't install try a later version (e.g. go1.5), just make sure you have the -B option after the version number.

### 설치 가능한 Go 버전 확인

```bash
gvm listall
```

### 설치 가능한 Go 버전 중에서 특정 버전 설치

```bash
gvm install go1.20.3
```

네, 그래프 형식의 다이어그램과 상태 다이어그램(state diagram)은 목적과 표현 방식에서 다릅니다.

1. **그래프 형식 다이어그램**:
   - 이는 일반적으로 노드(nodes)와 엣지(edges)를 사용하여 시스템이나 네트워크의 구조를 나타냅니다.
   - 노드는 시스템의 구성 요소를, 엣지는 구성 요소 간의 관계나 연결을 나타냅니다.
   - 예를 들어, 컴퓨터 네트워크의 물리적 또는 논리적 구조를 나타낼 때 사용됩니다.

2. **상태 다이어그램(State Diagram)**:
   - 상태 다이어그램은 시스템이나 객체의 상태와 그 상태에서의 이벤트에 의한 전환을 나타내는 데 초점을 맞춥니다.
   - 각 노드는 시스템의 특정 상태를 나타내며, 화살표는 한 상태에서 다른 상태로의 전환(이벤트나 조건에 의해 발생)을 나타냅니다.
   - 예를 들어, 소프트웨어 응용 프로그램의 상태(로그인, 로그아웃, 대기, 처리 중 등)와 상태 변경을 나타낼 때 사용됩니다.

두 다이어그램 유형은 서로 다른 정보와 관점을 제공합니다. 그래프 형식 다이어그램은 구조적 관계에 중점을 두는 반면, 상태 다이어그램은 동적인 상태 변화와 시스템의 행동에 더 많은 초점을 맞춥니다. Mermaid와 같은 도구는 이 두 유형의 다이어그램 모두를 지원하여, 복잡한 시스템이나 프로세스를 시각적으로 표현하는 데 사용될 수 있습니다.
