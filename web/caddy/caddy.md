# caddy

- [caddy](#caddy)
    - [Caddy 서버란?](#caddy-서버란)
    - [Examples](#examples)
        - [`caddy run --adapter caddyfile --config -`](#caddy-run---adapter-caddyfile---config--)

## [Caddy 서버](https://caddyserver.com/)란?

Caddy는 자동 HTTPS를 지원하는 오픈 소스 웹 서버입니다.
Caddy는 기본적으로 Let's Encrypt를 사용하여 SSL/TLS 인증서를 자동으로 발급하고 갱신합니다.
이는 웹 사이트를 HTTPS로 쉽게 전환할 수 있게 해주며, 보안과 SEO(검색 엔진 최적화) 측면에서 큰 이점을 제공합니다.

## Examples

### `caddy run --adapter caddyfile --config -`

```bash
caddy run --adapter caddyfile --config - <<'EOF'
go.somehost.co
tls cert.pem key.pem
reverse_proxy http://localhost:8080
EOF
```

- `caddy run`: Caddy 서버를 실행하는 명령입니다.
- `--adapter caddyfile`: 이 옵션은 Caddy의 구성 파일 형식을 지정합니다. `caddyfile`은 Caddy의 기본 구성 파일 형식입니다.
- `--config -`: 이 옵션은 구성을 표준 입력(`-`, stdin)에서 직접 읽어들이라는 의미입니다.
- `<<'EOF' ... EOF`: heredoc을 사용하여 여러 줄에 걸친 문자열을 입력으로 제공합니다.

구성 파일의 내용은 다음과 같습니다:
- `go.somehost.co`: Caddy가 서비스할 도메인 이름을 지정합니다.
- `tls cert.pem key.pem`:

    TLS(전송 계층 보안)를 위한 인증서와 키 파일을 지정하여 HTTPS 연결을 위한 암호화 설정을 합니다.

- `reverse_proxy http://localhost:8080`:

    리버스 프록시 설정을 정의하여 모든 요청을 `http://localhost:8080`으로 전달하도록 Caddy를 구성합니다.
    일반적으로 내부 네트워크에 있는 다른 서버나 서비스로 트래픽을 라우팅하는 데 사용됩니다.
