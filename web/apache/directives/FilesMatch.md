# FilesMatch

## `FilesMatch`와 `.htaccess`

[FilesMatch 사용](https://stackoverflow.com/questions/34350248/php-fpm-filesmatch-and-proxypassmatch-interchangeability)을 `.htaccess`의 rewrite 규칙이 적용됩니다.([`FilesMatch` 적용](https://ma.ttias.be/apache-2-4-proxypass-for-php-taking-precedence-over-filesfilesmatch-in-htaccess/) 참고)

`serving URL`을 보면 `/var/www/someservice/app/webroot/index.php`로 요청이 들어갑니다.

```log
    http_request.c(394): [client 172.19.0.1:59276] Headers received from client:
    http_request.c(398): [client 172.19.0.1:59276]   Host: someservice.localhost:8001
    http_request.c(398): [client 172.19.0.1:59276]   Connection: keep-alive
    http_request.c(398): [client 172.19.0.1:59276]   Cache-Control: max-age=0
    http_request.c(398): [client 172.19.0.1:59276]   Upgrade-Insecure-Requests: 1
    http_request.c(398): [client 172.19.0.1:59276]   User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36
    http_request.c(398): [client 172.19.0.1:59276]   Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
    http_request.c(398): [client 172.19.0.1:59276]   Accept-Encoding: gzip, deflate
    http_request.c(398): [client 172.19.0.1:59276]   Accept-Language: en-US,en;q=0.9,ko;q=0.8
    http_request.c(398): [client 172.19.0.1:59276]   Cookie: _ga=GA1.2.118852574.1649855783; _gcl_au=1.1.59898551.1649990172
  mod_authz_core.c(809): [client 172.19.0.1:59276] AH01626: authorization result of Require all granted: granted
  mod_authz_core.c(809): [client 172.19.0.1:59276] AH01626: authorization result of <RequireAny>: granted
         request.c(312): [client 172.19.0.1:59276] request authorized without authentication by access_checker_ex hook: /index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] strip per-dir prefix: /var/www/someservice/index.php -> index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] applying pattern '^$' to uri 'index.php'
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] strip per-dir prefix: /var/www/someservice/index.php -> index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] applying pattern '(.*)' to uri 'index.php'
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] rewrite 'index.php' -> 'app/webroot/index.php'
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] add per-dir prefix: app/webroot/index.php -> /var/www/someservice/app/webroot/index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] strip document_root prefix: /var/www/someservice/app/webroot/index.php -> /app/webroot/index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945e50a0/initial] [perdir /var/www/someservice/] internal redirect with /app/webroot/index.php [INTERNAL REDIRECT]
  mod_authz_core.c(809): [client 172.19.0.1:59276] AH01626: authorization result of Require all granted: granted
  mod_authz_core.c(809): [client 172.19.0.1:59276] AH01626: authorization result of <RequireAny>: granted
         request.c(312): [client 172.19.0.1:59276] request authorized without authentication by access_checker_ex hook: /app/webroot/index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945dfc60/initial/redir#1] [perdir /var/www/someservice/app/webroot/] strip per-dir prefix: /var/www/someservice/app/webroot/index.php -> index.php
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945dfc60/initial/redir#1] [perdir /var/www/someservice/app/webroot/] applying pattern '^' to uri 'index.php'
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945dfc60/initial/redir#1] [perdir /var/www/someservice/app/webroot/] RewriteCond: input='/var/www/someservice/app/webroot/index.php' pattern='!-d' => matched
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945dfc60/initial/redir#1] [perdir /var/www/someservice/app/webroot/] RewriteCond: input='/var/www/someservice/app/webroot/index.php' pattern='!-f' => not-matched
     mod_rewrite.c(476): [client 172.19.0.1:59276] 172.19.0.1 - - [someservice.localhost/sid#ffff9769cea0][rid#ffff945dfc60/initial/redir#1] [perdir /var/www/someservice/app/webroot/] pass through /var/www/someservice/app/webroot/index.php
     proxy_util.c(1986): [client 172.19.0.1:59276] *: using default reverse proxy worker for unix:/var/run/php/php5.6-fpm.sock|fcgi://localhost/var/www/someservice/app/webroot/index.php (no keepalive)
     proxy_util.c(1942): [client 172.19.0.1:59276] *: rewrite of url due to UDS(/var/run/php/php5.6-fpm.sock): fcgi://localhost/var/www/someservice/app/webroot/index.php (proxy:fcgi://localhost/var/www/someservice/app/webroot/index.php)
      mod_proxy.c(1161): [client 172.19.0.1:59276] AH01143: Running scheme unix handler (attempt 0)
  mod_proxy_fcgi.c(879): [client 172.19.0.1:59276] AH01076: url: fcgi://localhost/var/www/someservice/app/webroot/index.php proxyname: (null) proxyport: 0
  mod_proxy_fcgi.c(886): [client 172.19.0.1:59276] AH01078: serving URL fcgi://localhost/var/www/someservice/app/webroot/index.php
     proxy_util.c(2157): AH00942: FCGI: has acquired connection for (*)
     proxy_util.c(2210): [client 172.19.0.1:59276] AH00944: connecting fcgi://localhost/var/www/someservice/app/webroot/index.php to localhost:8000
     proxy_util.c(2247): [client 172.19.0.1:59276] AH02545: fcgi: has determined UDS as /var/run/php/php5.6-fpm.sock
     proxy_util.c(2419): [client 172.19.0.1:59276] AH00947: connected /var/www/someservice/app/webroot/index.php to httpd-UDS:0
     proxy_util.c(2698): AH02823: FCGI: connection established with Unix domain socket /var/run/php/php5.6-fpm.sock (*)
```
