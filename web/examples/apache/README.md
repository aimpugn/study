# Apache 베스트 프랙티스 conf 예제

복사해서 바로 쓸 수 있는 Apache 2.4 설정 모음입니다. 각 `.conf`는 디렉티브마다 **왜 그게 베스트 프랙티스이고, 요청 처리의 어느 단계에서 무슨 역할을 하는지**를 한글 주석으로 답니다. 그대로 운영에 쓰기보다, 주석을 읽고 자기 환경에 맞게 값을 조정하는 학습·출발점 용도입니다.

## 파일 구성과 적용 순서

요청이 도착해 응답이 나가기까지의 흐름 순서로 배치했습니다.

1. [`security-hardening.conf`](security-hardening.conf) — **서버 전역**. 정보 노출 차단, 위험 메서드 차단, 느린 요청(Slowloris) 방어, keep-alive·타임아웃. `<VirtualHost>` 밖 server config 컨텍스트에 둡니다.
2. [`vhost-http.conf`](vhost-http.conf) — **:80**. 인증서 발급(ACME) 경로만 열고 나머지는 HTTPS로 영구 리다이렉트.
3. [`vhost-https.conf`](vhost-https.conf) — **:443**. TLS, HSTS, 보안 헤더, HTTP/2, 압축·캐싱, 접근 제어, (선택) PHP-FPM 프록시.

적용 절차:

```bash
# 1) 필요한 모듈 활성화 (Debian/Ubuntu 기준)
sudo a2enmod headers reqtimeout ssl http2 deflate expires remoteip rewrite
# 2) 인증서 발급 (vhost-http.conf의 ACME 경로가 먼저 살아 있어야 함)
sudo certbot certonly --webroot -w /var/www/example/public -d example.com
# 3) conf 배치 후 문법 검사 — 통과(Syntax OK) 전에는 reload 금지
sudo apachectl configtest
# 4) 무중단 반영
sudo systemctl reload apache2
```

## 배치 환경별 주의 (단독 / WAF·LB 뒤 / 리버스 프록시)

각 conf 안에 케이스별 주석을 달았지만, 핵심만 먼저 정리합니다.

- **단독 웹서버(인터넷 직접 노출)**: 추가 설정 없이 그대로. TLS 종단·보안 헤더를 이 서버가 직접 책임집니다.
- **WAF·L4/L7 LB 뒤(예: BIG-IP)**: 클라이언트 실제 IP가 `X-Forwarded-For`로 들어오므로 `mod_remoteip`로 **신뢰 프록시 대역을 좁게** 지정해야 로그·접근제어가 실제 IP로 동작합니다. 신뢰 대역을 넓게 두면 공격자가 헤더를 위조해 IP 기반 제어를 우회합니다. LB가 TLS를 종단하면 리다이렉트 판단은 `X-Forwarded-Proto`로 합니다.
- **리버스 프록시 앞단**: 이 Apache가 프록시면 `ProxyPass`와 프록시 헤더(`X-Forwarded-*`)를 백엔드로 넘깁니다.

## 검증

적용 후 다음으로 확인합니다(자세한 명령은 보안 하드닝 문서 참고).

```bash
curl -sI https://example.com | grep -iE 'server|strict-transport|x-content-type|x-frame'
curl -i -X TRACE https://example.com        # 405 Method Not Allowed 면 정상
```

WAF 뒤라면 origin IP로 직접(`--resolve example.com:443:<ORIGIN_IP>`) 같은 검사를 한 번 더 해, origin 자체가 막고 있는지 확인합니다.

## 출처

- [Mozilla Server Side TLS (Intermediate)](https://wiki.mozilla.org/Security/Server_Side_TLS) — TLS 버전·cipher·HSTS 기준
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/) — 버전별 최신 설정 생성
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/) — 보안 응답 헤더 권고
- [Apache `mod_ssl`](https://httpd.apache.org/docs/2.4/mod/mod_ssl.html), [`core`](https://httpd.apache.org/docs/2.4/mod/core.html), [`mod_headers`](https://httpd.apache.org/docs/2.4/mod/mod_headers.html)
