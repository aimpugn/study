# go env

- [go env](#go-env)
    - [`GO111MODULE`](#go111module)
        - [options](#options)
    - [`GOARCH`](#goarch)
    - [`GOBIN`](#gobin)
    - [`GOCACHE`](#gocache)
    - [`GOENV`](#goenv)
    - [`GOEXE`](#goexe)
    - [`GOEXPERIMENT`](#goexperiment)
    - [`GOFLAGS`](#goflags)
    - [`GOHOSTARCH`](#gohostarch)
    - [`GOHOSTOS`](#gohostos)
    - [`GOINSECURE`](#goinsecure)
    - [`GOMODCACHE`](#gomodcache)
    - [`GONOPROXY`](#gonoproxy)
    - [`GONOSUMDB`](#gonosumdb)
    - [`GOOS`](#goos)
    - [`GOPATH`](#gopath)
    - [`GOPRIVATE`](#goprivate)
    - [`GOPROXY`](#goproxy)
    - [`GOROOT`](#goroot)
    - [`GOSUMDB`](#gosumdb)
    - [`GOTMPDIR`](#gotmpdir)
    - [`GOTOOLDIR`](#gotooldir)
    - [`GOVCS`](#govcs)
    - [`GOVERSION`](#goversion)
    - [`GCCGO`](#gccgo)
    - [`AR`](#ar)
    - [`CC`](#cc)
    - [`CXX`](#cxx)
    - [`CGO_ENABLED`](#cgo_enabled)
    - [`GOMOD`](#gomod)
    - [`GOWORK`](#gowork)
    - [`CGO_CFLAGS`](#cgo_cflags)
    - [`CGO_CPPFLAGS`](#cgo_cppflags)
    - [`CGO_CXXFLAGS=`](#cgo_cxxflags)
    - [`CGO_FFLAGS`](#cgo_fflags)
    - [`CGO_LDFLAGS`](#cgo_ldflags)
    - [`PKG_CONFIG`](#pkg_config)
    - [`GOGCCFLAGS`](#gogccflags)

## `GO111MODULE`

```bash
GO111MODULE=""
```

- Go 모듈의 사용 여부를 설정
- Go 1.16 이전에서는 모듈 지원이 기본적으로 활성화되지 않지만, Go 1.16 이후부터는 모듈 지원이 기본값으로 설정되었다
- 따라서, 모듈이 필요 없는 단일 파일 프로그램을 실행할 때 `GO111MODULE`을 "off"로 설정해야 한다
- `GO111MODULE`이 `"on"` 또는 `"auto"`로 설정되어 있고, 현재 작업 디렉토리에 go.mod 파일이 없다면, Go 시스템은 모듈의 위치를 찾으려고 시도하고, 이 과정에서 오류가 발생할 수 있다

### options

- "on"
    - `GOPATH`가 무시되고, 모든 Go 명령은 모듈 모드에서 실행
    - 모듈 모드에서는 `go.mod` 파일이 프로젝트의 루트에 있어야 하고, 이 파일은 프로젝트의 의존성과 다른 모듈 설정을 관리한다
    - 프로젝트에 go.mod 파일이 없는 경우, go mod init 명령을 사용하여 생성해야 한다
- 빈 문자열("") 또는 "auto"
    - 모듈 사용 여부가 자동으로 결정
    - Go 1.16
        - 이전에는 `$GOPATH/src` 안에 있지 **않은** 프로젝트에 대해서만 모듈 모드를 활성화. `$GOPATH/src` 안에 있는 프로젝트는 여전히 고전적인 `GOPATH` 방식을 사용
        - 이후에는 `$GOPATH/src` 안에 있더라도 모듈 모드가 기본으로 설정
    - auto 모드는 기존 프로젝트와 새 프로젝트 간의 호환성을 유지하고자 할 때 유용하다
- "off"
    - Go 시스템은 모듈 지원을 비활성화하고 `GOPATH` 모드로 돌아간다
    - 이 모드에서는 `go.mod` 파일이 무시된다
    - 단일 파일 스크립트나 간단한 프로젝트에서 유용할 수 있다

## `GOARCH`

```bash
GOARCH="arm64"
```

- Go 프로그램이 빌드될 때 타겟이 되는 아키텍처
- 여기서 "arm64"는 ARM 64비트 아키텍처를 의미

## `GOBIN`

```bash
GOBIN=""
```

- `go install` 명령어로 빌드된 실행 파일이 저장되는 디렉토리
- 빈 문자열("")은 기본값인 `$GOPATH/bin`을 사용함을 의미

## `GOCACHE`

```bash
GOCACHE="/Users/rody/Library/Caches/go-build"
```

- 컴파일된 Go 파일들이 캐시되는 디렉토리의 경로

## `GOENV`

```bash
GOENV="/Users/rody/Library/Application Support/go/env"
```

- Go 환경 설정 파일의 위치

## `GOEXE`

```bash
GOEXE=""
```

- Go 실행 파일의 확장자
    - Windows: ".exe`
    - macOS/Linux: 빈 문자열("")

## `GOEXPERIMENT`

```bash
GOEXPERIMENT=""
```

- Go 실험 기능의 사용 여부를 나타낸다

## `GOFLAGS`

```bash
GOFLAGS=""
```

- go 명령어에 기본적으로 적용되는 플래그

## `GOHOSTARCH`

```bash
GOHOSTARCH="arm64"
```

- GO HOST ARCH
- 현재 사용 중인 호스트 시스템의 아키텍처

## `GOHOSTOS`

```bash
GOHOSTOS="darwin"
```

- GO HOST OS
- 현재 사용 중인 호스트 시스템의 운영 체제

## `GOINSECURE`

```bash
GOINSECURE=""
```

- 보안되지 않은 방식으로 다운로드할 수 있는 패키지 목록

## `GOMODCACHE`

```bash
GOMODCACHE="/Users/rody/go/pkg/mod"
```

- 모듈 캐시 디렉토리의 경로입

## `GONOPROXY`

```bash
GONOPROXY="github.com/some-org/*"
```

- GOPROXY를 우회하여 직접 다운로드해야 하는 모듈의 목록

## `GONOSUMDB`

```bash
GONOSUMDB="github.com/some-org/*"
```

- 체크섬 데이터베이스를 확인하지 않을 모듈의 목록

## `GOOS`

```bash
GOOS="darwin"
```

- Go 프로그램이 빌드될 때 타겟이 되는 운영 체제
- 여기서 "darwin"은 macOS를 의미

## `GOPATH`

```bash
GOPATH="/Users/rody/go"
```

- Go의 작업 공간 디렉토리
- 소스 코드, 의존성, 빌드 파일 등이 저장된다

## `GOPRIVATE`

```bash
GOPRIVATE="github.com/some-org/*"
```

- 프라이빗 모듈의 목록. 특정한 패키지 경로들을 비공개로 취급할 때 사용한다.
- 비공개 또는 내부 저장소에 호스팅되는 Go 패키지들에 대해 Go 도구가 이들을 [`공개 모듈 프록시`](https://goproxy.io/docs/introduction.html)나 [`체크섬 데이터베이스`](https://go.dev/ref/mod#checksum-database)에 요청하지 않도록 지시한다
- 이러한 패키지들은 외부 공개 저장소에 존재하지 않으므로, Go 도구는 이들을 직접적으로 해당 소스 위치에서 가져와야 한다.

## `GOPROXY`

```bash
GOPROXY="https://proxy.golang.org,direct"
```

- 모듈을 다운로드할 때 사용하는 프록시 서버의 주소

## `GOROOT`

```bash
GOROOT="/opt/homebrew/Cellar/go/1.20.3/libexec"
```

- Go 언어와 표준 라이브러리가 설치된 디렉토리의 경로

## `GOSUMDB`

```bash
GOSUMDB="sum.golang.org"
```

- 모듈의 체크섬을 확인하는 데 사용되는 데이터베이스

## `GOTMPDIR`

```bash
GOTMPDIR=""
```

- 임시 파일을 저장할 디렉토리

## `GOTOOLDIR`

```bash
GOTOOLDIR="/opt/homebrew/Cellar/go/1.20.3/libexec/pkg/tool/darwin_arm64"
```

- Go 도구(컴파일러, 링커 등)가 설치된 디렉토리

## `GOVCS`

```bash
GOVCS=""
```

## `GOVERSION`

```bash
GOVERSION="go1.20.3"
```

- 현재 설치된 Go 버전

## `GCCGO`

```bash
GCCGO="gccgo"
```

- Go 소스 코드를 컴파일하는 데 사용되는 GCC 기반의 Go 컴파일러

## `AR`

```bash
AR="ar"
```

- 아카이브 파일을 만드는 데 사용되는 유틸리티

## `CC`

```bash
CC="cc"
```

- C 컴파일러

## `CXX`

```bash
CXX="c++"
```

- C++ 컴파일

## `CGO_ENABLED`

```bash
CGO_ENABLED="0"
```

- cgo가 활성화되어 있는지 여부를 나타낸다
- 여기서 "0"은 비활성화됨을 의미

## `GOMOD`

```bash
GOMOD="/dev/null"
```

- 현재 모듈의 `go.mod` 파일의 경로
- "/dev/null"은 현재 작업 중인 디렉토리에 go.mod 파일이 없음을 의미

## `GOWORK`

```bash
GOWORK=""
```

- Go 1.18부터 도입된 작업 공간 파일의 경로

## `CGO_CFLAGS`

```bash
CGO_CFLAGS="-O2 -g"
```

- Cgo를 사용할 때 적용되는 각종 컴파일러와 링커 플래그

## `CGO_CPPFLAGS`

```bash
CGO_CPPFLAGS=""
```

- Cgo를 사용할 때 적용되는 각종 컴파일러와 링커 플래그

## `CGO_CXXFLAGS=`

```bash
CGO_CXXFLAGS="-O2 -g"
```

- Cgo를 사용할 때 적용되는 각종 컴파일러와 링커 플래그

## `CGO_FFLAGS`

```bash
CGO_FFLAGS="-O2 -g"
```

- Cgo를 사용할 때 적용되는 각종 컴파일러와 링커 플래그

## `CGO_LDFLAGS`

```bash
CGO_LDFLAGS="-O2 -g"
```

- Cgo를 사용할 때 적용되는 각종 컴파일러와 링커 플래그

## `PKG_CONFIG`

```bash
PKG_CONFIG="pkg-config"
```

- pkg-config 유틸리티의 경로

## `GOGCCFLAGS`

```bash
GOGCCFLAGS="-fPIC -arch arm64 -fno-caret-diagnostics -Qunused-arguments -fmessage-length=0 -fdebug-prefix-map=/var/folders/9x/8djp1ylj221bk02dp8zqsps00000gn/T/go-build3109827647=/tmp/go-build -gno-record-gcc-switches -fno-common"
```

- Go 컴파일러가 사용하는 GCC 플래그
