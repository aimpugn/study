# asdf

## asdf?

asdf is a CLI tool that can manage multiple language runtime versions on a per-project basis. It is like `nvm` for Node.js, `rbenv` for Ruby, `pyenv` for Python, `phpenv` for PHP, and so on.

## 설치

### macos

```bash
brew install asdf
```

```log
To use asdf, add the following line (or equivalent) to your shell profile
e.g. ~/.profile or ~/.zshrc:
  . /opt/homebrew/opt/asdf/libexec/asdf.sh
e.g. ~/.config/fish/config.fish
  source /opt/homebrew/opt/asdf/libexec/asdf.fish
Restart your terminal for the settings to take effect.

zsh completions have been installed to:
  /opt/homebrew/share/zsh/site-functions
```

## plugins

### go

```bash
asdf plugin-add golang
```

```bash
asdf plugin-add golang https://github.com/kennyp/asdf-golang.git
```

## golang 설치

### 설치 가능한 버전 확인하기

```bash
asdf list all golang
```

### 1.20.3 설치하기

```log
❯ asdf install golang 1.20.3
Platform 'darwin' supported!
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 92.3M  100 92.3M    0     0  6576k      0  0:00:14  0:00:14 --:--:-- 6875k
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    64  100    64    0     0    220      0 --:--:-- --:--:-- --:--:--   219
verifying checksum
/Users/rody/.asdf/downloads/golang/1.20.3/archive.tar.gz: OK
checksum verified
```

```bash
# apply globally
asdf global golang 1.20.3
```

### go env

```log
❯ go env
GO111MODULE=""
GOARCH="arm64"
GOBIN=""
GOCACHE="/Users/rody/Library/Caches/go-build"
GOENV="/Users/rody/Library/Application Support/go/env"
GOEXE=""
GOEXPERIMENT=""
GOFLAGS=""
GOHOSTARCH="arm64"
GOHOSTOS="darwin"
GOINSECURE=""
GOMODCACHE="/Users/rody/.asdf/installs/golang/1.20.3/packages/pkg/mod"
GONOPROXY="github.com/some-qwerty-org.io/*"
GONOSUMDB="github.com/some-qwerty-org.io/*"
GOOS="darwin"
GOPATH="/Users/rody/.asdf/installs/golang/1.20.3/packages"
GOPRIVATE="github.com/some-qwerty-org.io/*"
GOPROXY="https://proxy.golang.org,direct"
GOROOT="/Users/rody/.asdf/installs/golang/1.20.3/go"
GOSUMDB="sum.golang.org"
GOTMPDIR=""
GOTOOLDIR="/Users/rody/.asdf/installs/golang/1.20.3/go/pkg/tool/darwin_arm64"
GOVCS=""
GOVERSION="go1.20.3"
GCCGO="gccgo"
AR="ar"
CC="clang"
CXX="clang++"
CGO_ENABLED="0"
GOMOD="/dev/null"
GOWORK=""
CGO_CFLAGS="-O2 -g"
CGO_CPPFLAGS=""
CGO_CXXFLAGS="-O2 -g"
CGO_FFLAGS="-O2 -g"
CGO_LDFLAGS="-O2 -g"
PKG_CONFIG="pkg-config"
GOGCCFLAGS="-fPIC -arch arm64 -fno-caret-diagnostics -Qunused-arguments -fmessage-length=0 -fdebug-prefix-map=/var/folders/9x/8djp1ylj221bk02dp8zqsps00000gn/T/go-build2641947923=/tmp/go-build -gno-record-gcc-switches -fno-common"
```

### vscode에 반영하기

VSCode에서 Go 환경을 설정할 때 `GOROOT`와 `GOPATH` 환경 변수가 올바르게 설정되어 있지 않으면 위와 같은 오류 메시지가 나타날 수 있습니다. `asdf`를 사용하여 여러 버전의 Go를 관리하는 경우, 각 버전마다 `GOROOT`와 `GOPATH`가 달라질 수 있으므로, 이를 자동으로 관리하는 방법이 필요합니다.

#### 자동 설정 방법

1. **셸 초기화 파일 수정**: 사용하는 셸의 초기화 파일(`.bashrc`, `.zshrc` 등)에 `GOROOT`와 `GOPATH`를 동적으로 설정하는 코드를 추가할 수 있습니다. `asdf`는 현재 선택된 Go 버전에 따라 `GOROOT`를 자동으로 설정하지만, `GOPATH`는 사용자가 직접 설정해야 합니다. 다음은 `zsh`를 사용하는 경우 `.zshrc` 파일에 추가할 수 있는 예시입니다:

   ```bash
   export GOPATH="$HOME/go"
   export PATH="$GOPATH/bin:$PATH"
   ```

   이렇게 설정하면, `GOPATH`는 사용자의 홈 디렉토리 아래에 고정되며, 여러 Go 버전을 전환해도 `GOPATH`는 변경되지 않습니다.

2. **VSCode 설정**: VSCode에서 `GOROOT`와 `GOPATH`를 자동으로 인식하도록 하려면, VSCode의 설정(`settings.json`)에 다음 항목을 추가합니다:

   ```json
   "go.goroot": "${env:GOROOT}",
   "go.gopath": "${env:GOPATH}"
   ```

   이렇게 설정하면, VSCode가 환경 변수에서 `GOROOT`와 `GOPATH` 값을 읽어와서 사용합니다.

3. **환경 변수 확인**: VSCode를 재시작한 후, 터미널에서 `echo $GOROOT`와 `echo $GOPATH`를 실행하여 환경 변수가 올바르게 설정되었는지 확인합니다.

4. **VSCode 재시작**: 설정을 변경한 후에는 VSCode를 재시작하거나, "Reload Window" 명령을 실행하여 변경사항을 적용해야 합니다.

#### 주의사항

- `GOPATH`는 Go 모듈이 활성화되기 전(1.11 버전 이전)에 주로 사용되던 환경 변수입니다. Go 1.11 이상에서는 모듈을 사용하여 프로젝트를 관리하는 것이 권장되며, 이 경우 `GOPATH` 설정은 덜 중요해집니다.
- `asdf`를 사용할 때는 `asdf reshim golang` 명령을 실행하여 셸에 설치된 Go 버전의 변경사항을 반영해야 할 수도 있습니다.

이러한 설정을 통해 `asdf`로 관리하는 Go 버전을 변경할 때마다 `GOROOT`와 `GOPATH`를 수동으로 조정할 필요 없이, VSCode에서 자동으로 올바른 환경을 인식하도록 할 수 있습니다.

## `$PATH` 변수

```bash
export GOPATH="$HOME/go"
export GOROOT="$(asdf where golang)/go"

export PATH="\
$HOME/fvm/default/bin: \
$(go env GOPATH)/bin: \
$HOMEBREW_PREFIX/opt/python@3.11/libexec/bin: \
$HOME/.pub-cache/bin: \
$HOME/.composer/vendor/bin: \
$HOME/go/bin: \
/opt/homebrew/opt/mysql-client/bin: \
/opt/homebrew/bin: \
/opt/homebrew/opt/php@8.1/bin: \
/opt/homebrew/opt/bison/bin: \
${PATH}"
```
