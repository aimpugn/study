# curl

- [curl](#curl)
    - [`curl --version`](#curl---version)
    - [`--resolve`](#--resolve)
    - [`--connected-to`](#--connected-to)
        - [예제](#예제)
    - [`--location`](#--location)
        - [예제](#예제-1)
    - [`--cookie-jar` \&\& POST 로그인 하기](#--cookie-jar--post-로그인-하기)
    - [POST JSON body](#post-json-body)
    - [리모트 서버의 SSL/TLS 버전 확인하기](#리모트-서버의-ssltls-버전-확인하기)
    - [`Access-Control-Allow-Origin` 헤더 확인하기](#access-control-allow-origin-헤더-확인하기)
        - [Sending a preflight request using cUrl](#sending-a-preflight-request-using-curl)
    - [예제](#예제-2)
        - [액세스 토큰 발급 받고 API 호출하기](#액세스-토큰-발급-받고-api-호출하기)

## `curl --version`

```shell
curl --version
curl 7.47.0 (x86_64-pc-linux-gnu) libcurl/7.47.0 GnuTLS/3.4.10 zlib/1.2.8 libidn/1.32 librtmp/2.3
Protocols: dict file ftp ftps gopher http https imap imaps ldap ldaps pop3 pop3s rtmp rtsp smb smbs smtp smtps telnet tftp
Features: AsynchDNS IDN IPv6 Largefile GSS-API Kerberos SPNEGO NTLM NTLM_WB SSL libz TLS-SRP UnixSockets
```

- `curl 7.47.0`(사용 중인 cURL의 버전.): 이 버전은 cURL의 기능 및 지원 프로토콜에 대한 정보를 제공한다.
- `libcurl/7.47.0 GnuTLS/3.4.10`(라이브러리 및 프로토콜 지원):
    - cURL이 `GnuTLS` 라이브러리를 사용 의미
    - `GnuTLS`는 SSL/TLS 프로토콜을 구현하는 라이브러리로, 보안 연결을 제공한다.
    - 이 버전의 `GnuTLS`는 TLS 1.2를 지원
- `Protocols`(지원하는 프로토콜):
    - cURL이 지원하는 다양한 네트워크 프로토콜들
    - 예를 들어 `https`, `ftps`는 각각 HTTP와 FTP에 SSL/TLS 보안을 추가한 버전
- `Features`(기능)
    - cURL이 제공하는 추가적인 기능들
    - 예를 들어 `SSL`, `TLS-SRP`, `IPv6`, `AsynchDNS` 등이 있다.
    - 이 중 `SSL`은 SSL 프로토콜을 지원함을 의미하고, `TLS-SRP`는 TLS-Secure Remote Password 프로토콜을 지원한다는 것을 나타낸다

## `--resolve`

- 특정 도메인을 특정 IP 주소로 해석하도록 curl에 지시한다

```bash
--resolve <[+]host:port:addr[,addr]...>
```

Provide a custom address for a specific host and port pair. Using this, you can make the curl requests(s) use a specified address and prevent the otherwise normally resolved address to be used. Consider it a sort of
/etc/hosts alternative provided on the command line. The port number should be the number used for the specific protocol the host is used for. It means you need several entries if you want to provide address for the same
host but different ports.

By specifying '*' as host you can tell curl to resolve any host and specific port pair to the specified address. Wildcard is resolved last so any `--resolve` with a specific host and port is used first.

The provided address set by this option is used even if -4, --ipv4 or -6, --ipv6 is set to make curl use another IP version.

By prefixing the host with a '+' you can make the entry time out after curl's default timeout (1 minute). Note that this only makes sense for long running parallel transfers with a lot of files. In such cases, if this
option is used curl tries to resolve the host as it normally would once the timeout has expired.

Support for providing the IP address within [brackets] was added in 7.57.0.

Support for providing multiple IP addresses per entry was added in 7.59.0.

Support for resolving with wildcard was added in 7.64.0.

Support for the '+' prefix was was added in 7.75.0.

This option can be used many times to add many host names to resolve.

`--resolve` can be used several times in a command line

Example:

```bash
curl --resolve example.com:443:127.0.0.1 https://example.com

curl --resolve example.com:443:10.0.6.xxx https://example.com
```

See also `--connect-to` and `--alt-svc`.

```log
# <DOMAIN:PORT:ADDRESS> 형식이어야 하는데, 뒤에 443이 잘못 붙음
curl: (49) Couldn't parse CURLOPT_RESOLVE entry 'api.server.com:443:xxx:443'
```

## `--connected-to`

> --connect-to 'HOST1:PORT1:HOST2:PORT2'

특정 호스트와 포트를 다른 주소와 포트로 리디렉션하는 데 사용된다. 특정 서버에 요청을 보내고자 할 때 유용하며, 특히 클러스터의 특정 노드에 요청을 보내는 데 사용될 수 있다.
네트워크 연결을 설정하는 데만 사용되며, TLS/SSL (예: SNI, 인증서 검증) 또는 애플리케이션 프로토콜에 사용되는 호스트/포트에는 영향을 미치지 않는다.

For a request to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead.  This option is suitable to direct requests at a specific server, e.g. at a specific cluster node in a cluster of servers. This option is only
used to establish the network connection. It does NOT affect the hostname/port that is used for TLS/SSL (e.g. SNI, certificate verification) or for the application protocols. "HOST1" and "PORT1" may be the empty string,
meaning "any host/port". "HOST2" and "PORT2" may also be the empty string, meaning "use the request's original host/port".

A "host" specified to this option is compared as a string, so it needs to match the name used in request URL. It can be either numerical such as "127.0.0.1" or the full host name such as "example.org".

--connect-to can be used several times in a command line

### 예제

```bash
curl --connect-to example.com:443:example.net:8443 https://example.com

curl --connect-to example.com:443:10.0.6.xxx:443 https://example.com
```

See also --resolve and -H, --header.

## `--location`

HTTP 리다이렉션을 따르도록 지시한다. 즉, 서버가 요청을 다른 URL로 리다이렉트하더라도 `curl`은 그 위치로 이동하여 요청을 계속 진행한다.
서버가 리다이렉트를 시도하면 `curl`은 그 위치로 이동하여 요청을 계속 진행하고, 원래 요청된 URL이 아닌 리다이렉트된 URL에 대한 응답을 받게 된다.

> **HTTP 리다이렉션?**
>
> 서버가 클라이언트에게 요청된 리소스가 다른 위치에 있음을 알리는 방법.
> 일반적으로 웹 사이트의 구조를 변경하거나, 리소스를 다른 서버로 이동하는 등의 상황에서 사용된다.
> 리다이렉션은 3xx HTTP 상태 코드를 사용하여 표시된다.

### 예제

```bash
# `http://example.com/redirect` URL이 `http://example.com/new-location`로 리다이렉트되도록 설정되어 있다면,
# 자동으로 `http://example.com/new-location`로 이동하여 요청을 계속 진행
curl --location http://example.com/redirect
```

## `--cookie-jar` && POST 로그인 하기

- [쿠키를 사용할 경우](https://stackoverflow.com/a/45755598/8562273)

```shell
curl 'https://classic-admin.some-qwerty-org.io/users/login' \
    --cookie-jar cookies.txt \
    --form 'data[User][email]="demo.ksnet@legacy.email.com"' \
    --form 'data[User][password]="dkdlavhxm"' -v -L \
    | grep "결제승인내역"
```

- `cat cookies.txt`

```log
# Netscape HTTP Cookie File
# https://curl.se/docs/http-cookies.html
# This file was generated by libcurl! Edit at your own risk.

#HttpOnly_classic-admin.some-qwerty-org.io    FALSE    /    FALSE    1679178260    CAKEPHP    ni0ujm9597j9q5f5tfp3fam9b4
classic-admin.some-qwerty-org.io    FALSE    /    TRUE    1679768661    AWSALBCORS    0r/CCpIzQDKRPV3rjIfeY2c/CZAqiLe+nt2deDrToKb42u4AghIfWCPWN9wcOHnH89obCCnlgdziCElnBofH3iNXzAa2wtK44pNvZmMqjtqUk/tD/8h6hS3GnCxx
classic-admin.some-qwerty-org.io    FALSE    /    FALSE    1679768661    AWSALB    0r/CCpIzQDKRPV3rjIfeY2c/CZAqiLe+nt2deDrToKb42u4AghIfWCPWN9wcOHnH89obCCnlgdziCElnBofH3iNXzAa2wtK44pNvZmMqjtqUk/tD/8h6hS3GnCxx
```

## POST JSON body

```shell
curl -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <token>" \
    -d '{"key1":null ,"key2": 1000, "key3": "value3"}' \
    https://host.com/api/end/point
```

## 리모트 서버의 SSL/TLS 버전 확인하기

```shell
# TLS 1.0 사용
curl -v --tls-max 1.0 https://yourserver.com

# TLS 1.1 사용
curl -v --tls-max 1.1 https://yourserver.com

# TLS 1.2 사용
curl -v --tls-max 1.2 https://yourserver.com

# TLS 1.3 사용
curl -v --tls-max 1.3 https://yourserver.com
```

## `Access-Control-Allow-Origin` 헤더 확인하기

```bash
curl -I -H "Origin: http://example.com" \
     --verbose \
    -H "Access-Control-Request-Method: GET" \
    'https://cdn.some_name.kr/path/to/sdk.js'
```

### [Sending a preflight request using cUrl](https://stackoverflow.com/a/12179364)

```bash
curl -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: X-Requested-With" \
  -X OPTIONS --verbose \
  https://www.googleapis.com/discovery/v1/apis?fields=
```

## 예제

### 액세스 토큰 발급 받고 API 호출하기

```shell
TOKEN=$(curl -s -H 'Content-Type: application/json' \
  -d '{"apiKey":"apiKeyValue","apiSecret":"apiSecretValue"}' \
  https://core-api.dev.some_name.co/users/getToken)

ACCESS_TOKEN=$(jq -r -n --arg token "$TOKEN" '$token | fromjson | .response.access_token')

echo "access_token: $ACCESS_TOKEN"

curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  'https://core-api.dev.some_name.co/path/to/api?from=1704034800&to=1706713200&page=1&limit=100' |
  jq
```
