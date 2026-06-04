# Nginx 베스트 프랙티스 conf 예제

복사해서 바로 쓸 수 있는 Nginx 설정 모음입니다. 각 파일은 디렉티브마다 **왜 베스트 프랙티스이고, 요청 처리의 어느 단계에서 무슨 역할을 하는지**를 한글 주석으로 답니다. 그대로 쓰기보다 주석을 읽고 환경에 맞게 조정하는 학습·출발점 용도입니다.

## 파일 구성과 적용 순서

1. [`nginx.conf`](nginx.conf) — **main + http 블록 전역**. 워커·이벤트 모델, 정보 노출 차단, TLS 공통(모든 server가 상속), 압축, 타임아웃, rate limit zone, 로그 포맷. `conf.d/*.conf`로 server 블록을 include.
2. [`server-http.conf`](server-http.conf) — **:80**. ACME 경로만 평문으로 열고 나머지는 HTTPS로 리다이렉트.
3. [`server-https.conf`](server-https.conf) — **:443**. TLS(인증서만; 공통 설정은 http에서 상속), HSTS·보안 헤더, HTTP/2, 정적 캐싱, rate limit 적용, PHP-FPM 또는 프록시.

적용 절차:

```bash
# 1) 인증서 (server-http.conf의 ACME 경로가 먼저 살아 있어야 함)
sudo certbot certonly --webroot -w /var/www/certbot -d example.com
# 2) 문법 검사 — Syntax OK 전에는 reload 금지
sudo nginx -t
# 3) 무중단 반영
sudo nginx -s reload
```

## 꼭 기억할 nginx 함정 — `add_header` 상속

nginx의 `add_header`는 **하위 블록(`location`)에 `add_header`가 하나라도 있으면, 상위(`server`)의 `add_header`를 상속하지 않고 그 블록에서 통째로 버립니다.** 즉 `server`에 보안 헤더 6줄을 두고 어떤 `location`에서 `add_header Cache-Control ...` 한 줄만 써도, 그 `location` 응답에는 HSTS·CSP 같은 보안 헤더가 사라집니다.

대응:
- 보안 헤더가 필요한 모든 `location`에서 헤더를 다시 선언하거나,
- 공통 헤더를 별도 파일로 만들어 각 `location`에서 `include` 하거나,
- `add_header`를 쓰는 `location`을 최소화합니다.

이 함정은 `server-https.conf`의 정적 캐싱 `location` 주석에서 다시 짚습니다.

## 배치 환경별 주의 (단독 / WAF·LB 뒤 / 리버스 프록시)

- **단독**: 추가 설정 없이 그대로.
- **WAF·LB 뒤(예: BIG-IP)**: 실제 클라이언트 IP가 `X-Forwarded-For`로 오므로 `set_real_ip_from`(신뢰 LB 대역만 좁게) + `real_ip_header`로 `$remote_addr`를 실제 IP로 복원해야 로그·rate limit이 제대로 동작합니다. 대역을 넓히면 헤더 위조로 우회됩니다. LB가 TLS를 종단하면 리다이렉트는 `$http_x_forwarded_proto`로 판단합니다.
- **리버스 프록시 앞단**: `proxy_pass` + `proxy_set_header`로 `Host`·`X-Forwarded-*`를 백엔드에 넘깁니다.

## 검증

```bash
curl -sI https://example.com | grep -iE 'server|strict-transport|content-security'
curl -i -X TRACE https://example.com    # nginx는 기본적으로 405 (TRACE 미지원)
```

## 출처

- [Mozilla Server Side TLS (Intermediate)](https://wiki.mozilla.org/Security/Server_Side_TLS)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [nginx `ngx_http_headers_module` (add_header 상속 규칙)](https://nginx.org/en/docs/http/ngx_http_headers_module.html)
- [nginx `ngx_http_ssl_module`](https://nginx.org/en/docs/http/ngx_http_ssl_module.html)
