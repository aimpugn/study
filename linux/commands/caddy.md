# caddy

## caddy?

Caddy는 Go 언어로 작성된 확장 가능한 서버 플랫폼이다.
이는 특히 웹 서버, 리버스 프록시, 그리고 HTTPS를 자동화하는 데 사용한다.
고성능, 확장성, 보안성을 갖춘 웹 서버 및 리버스 프록시 서버로서, 다양한 웹 애플리케이션과 서비스에 적합한 플랫폼이다.

## Caddy의 핵심 설계 철학과 모듈 시스템

> Modules are plugged in statically at compile-time to provide useful functionality.
> 모듈이 컴파일 타임에 정적으로 플러그인 된다

1. 구성 관리: Caddy의 핵심 기능으로, Caddy의 주요 기능은 구성을 관리하는 것입니다. 이는 Caddy가 사용자가 정의한 구성 파일을 해석하고, 그에 따라 서버의 동작을 조정한다는 의미입니다. 예를 들어, HTTPS 설정, 리버스 프록시 규칙, 로드 밸런싱 설정 등을 구성 파일을 통해 관리합니다.

2. 모듈 시스템 - 컴파일 타임에 정적 플러그인: Caddy는 확장 가능한 서버 플랫폼으로, 다양한 모듈(기능 확장을 위한 추가 구성 요소)을 지원합니다. "모듈이 컴파일 타임에 정적으로 플러그인 된다"는 것은 Caddy의 실행 파일이 컴파일 될 때, 필요한 모듈들이 프로그램에 포함되어 정적으로 통합된다는 의미입니다. 즉, 서버가 실행되기 전에, 컴파일 과정에서 필요한 모듈들이 Caddy 실행 파일에 직접 포함되어 컴파일되는 것을 의미합니다.

이러한 접근 방식의 장점은 성능과 안정성입니다. 정적으로 플러그인된 모듈은 실행 중에 추가하거나 제거할 필요가 없으며, 이는 실행 중인 서버의 안정성을 향상시킬 수 있습니다. 또한, 필요한 모듈만 포함시켜 더 가벼운 실행 파일을 만들 수 있으며, 이는 리소스 사용을 최적화하는 데 도움이 됩니다.

### Caddy의 핵심 기능

- 구성 관리
    - Caddy의 핵심은 구성을 관리하는 것
    - 사용자가 정의한 구성에 따라 웹 서버, 리버스 프록시 등의 역할을 수행한다.
- 모듈 기반 확장성
    - 컴파일 시점에 정적으로 플러그인되는 모듈을 통해 다양한 기능을 제공한다.
    - 이러한 모듈은 HTTP 서비스, TLS, PKI 애플리케이션 등을 지원합니다.
- 인증서 자동화
    - Caddy는 표준 배포판에서 HTTPS 인증서를 자동으로 관리하고 갱신하는 기능을 포함한다
    - Let's Encrypt와 같은 `ACME` 프로토콜을 사용하여 SSL/TLS 인증서를 자동으로 발급하고 갱신한다.

### Caddy의 사용 방법

- `caddy run`: Caddy를 전경에서 실행합니다. 이 방식은 서버를 직접 관리하고 모니터링하는 데 권장됩니다.
- `caddy start`: Caddy를 백그라운드에서 시작합니다. 서버를 중단하려면 해당 터미널 창을 열어두고 `caddy stop`을 실행해야 합니다.

Caddy가 시작되면 로컬로 바인딩된 관리 소켓을 열고, 여기에 [RESTful HTTP API](https://caddyserver.com/docs/api)를 통해 구성을 POST할 수 있습니다.

### Caddy를 리버스 프록시로 사용할 때의 동작

리버스 프록시로서 Caddy는 클라이언트로부터 들어오는 요청을 받아 실제 서버(백엔드 서버)로 전달하고, 그 서버로부터의 응답을 다시 클라이언트에게 전달합니다. 이 과정에서 Caddy는 여러 가지 추가적인 기능을 수행할 수 있습니다:

1. 부하 분산: Caddy는 여러 백엔드 서버 사이에서 들어오는 요청을 분산시킬 수 있습니다. 이를 통해 트래픽이 많은 환경에서도 안정적인 서비스를 제공할 수 있습니다.

2. HTTPS 자동화: 리버스 프록시를 사용할 때 Caddy는 HTTPS를 통한 보안 통신을 자동으로 설정합니다. 클라이언트와 Caddy 사이의 통신은 SSL/TLS로 암호화되어 보안이 강화됩니다.

3. 캐싱 및 압축: Caddy는 정적 콘텐츠를 캐싱하고, HTTP 응답을 압축하여 빠른 페이지 로딩 속도와 효율적인 대역폭 사용을 지원합니다.

4. 헤더 조작 및 리다이렉션: Caddy는 요청 또는 응답 헤더를 조작하거나, 특정 조건에 따라 요청을 다른 URL로 리다이렉션할 수 있습니다.

5. 로그 관리 및 분석: Caddy는 요청과 응답에 대한 로그를 기록하여 서버의 성능과 문제를 분석하는 데 도움을 줍니다.

Caddy의 강력한 기능 중 하나는 사용자가 쉽게 구성할 수 있는 선언적 구성 파일을 통해 이러한 기능을 설정하고 관리할 수 있다는 점입니다. 구성 파일을 통해 리버스 프록시 규칙, SSL/TLS 인증서 관리, 로드 밸런싱 설정 등을 간편하게 정의할 수 있습니다.

Caddy의 리버스 프록시 설정은 매우 직관적이며, 몇 줄의 구성만으로도 강력한 리버스 프록시 서버를 구축할 수 있습니다. 예를 들어, 특정 호스트로 들어오는 요청을 다른 서버로 전달하는 등의 동작을 쉽게 설정할 수 있습니다. 또한, Caddy는 기본적으로 SSL/TLS를 지원하므로 HTTPS를 통한 안전한 연결을 자동으로 설정하며, 필요에 따라 HTTP/2 및 HTTP/3도 지원합니다.

## 사용법

To run Caddy, use:

- `caddy run` to run Caddy in the foreground (recommended).
- `caddy start` to start Caddy in the background; only do this if you will be keeping the terminal window open until you run `caddy stop` to close the server.

When Caddy is started, it opens a locally-bound administrative socket to which configuration can be POSTed via a [restful HTTP API](https://caddyserver.com/docs/api).

Caddy's native configuration format is JSON. However, config adapters can be used to convert other config formats to JSON when Caddy receives its configuration. The Caddyfile is a built-in config adapter that is popular for hand-written configurations due to its [straightforward syntax](https://caddyserver.com/docs/caddyfile). Many [third-party adapters](https://caddyserver.com/docs/config-adapters) are available.  Use 'caddy adapt' to see how a config translates to JSON.

For convenience, the CLI can act as an HTTP client to give Caddy its initial configuration for you. If a file named Caddyfile is in the current working directory, it will do this automatically. Otherwise, you can use the `--config` flag to specify the path to a config file.

Some special-purpose subcommands build and load a configuration file for you directly from command line input; for example:

- `caddy file-server`
- `caddy reverse-proxy`
- `caddy respond`

These commands disable the administration endpoint because their configuration is specified solely on the command line.

In general, the most common way to run Caddy is simply:

```bash
caddy run
```

### reverse proxy

```bash
caddy run --adapter caddyfile `--config` - <<'EOF'
go.some_name.co
tls cert.pem key.pem
reverse_proxy http://localhost:8080
EOF
```
