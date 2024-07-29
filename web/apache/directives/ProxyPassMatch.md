# ProxyPassMatch

## `ProxyPassMatch`와 `.htaccess`

`ProxyPassMatch` 사용 시 `.htaccess`의 rewrite 규칙이 적용이 안 되는 것으로 보입니다.

반면 [FilesMatch 사용](https://stackoverflow.com/questions/34350248/php-fpm-filesmatch-and-proxypassmatch-interchangeability)을 `.htaccess`의 rewrite 규칙이 적용됩니다.([`FilesMatch` 적용](https://ma.ttias.be/apache-2-4-proxypass-for-php-taking-precedence-over-filesfilesmatch-in-htaccess/) 참고)

`.htaccess`의 rewrite rule이 적용되지 않고, `serving URL`을 보면 `/var/www/someservice//index.php`로 요청이 들어갑니다.
반면 php가 아닌 경우에는 `ProxyPassMatch`에 안 걸려서 rewrite가 동작합니다.

```log
  http_request.c(394): [client 172.19.0.1:59234] Headers received from client:
  http_request.c(398): [client 172.19.0.1:59234]   Host: someservice.localhost:8001
  http_request.c(398): [client 172.19.0.1:59234]   Connection: keep-alive
  http_request.c(398): [client 172.19.0.1:59234]   Cache-Control: max-age=0
  http_request.c(398): [client 172.19.0.1:59234]   Upgrade-Insecure-Requests: 1
  http_request.c(398): [client 172.19.0.1:59234]   User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36
  http_request.c(398): [client 172.19.0.1:59234]   Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
  http_request.c(398): [client 172.19.0.1:59234]   Accept-Encoding: gzip, deflate
  http_request.c(398): [client 172.19.0.1:59234]   Accept-Language: en-US,en;q=0.9,ko;q=0.8
  http_request.c(398): [client 172.19.0.1:59234]   Cookie: _ga=GA1.2.118852574.1649855783; _gcl_au=1.1.59898551.1649990172
mod_authz_core.c(835): [client 172.19.0.1:59234] AH01628: authorization result: granted (no directives)
       request.c(312): [client 172.19.0.1:59234] request authorized without authentication by access_checker_ex hook: /index.php
 mod_proxy_fcgi.c(52): [client 172.19.0.1:59234] canonicalising URL //localhost/var/www/someservice//index.php
 mod_proxy_fcgi.c(84): [client 172.19.0.1:59234] AH01060: set r->filename to proxy:fcgi://localhost/var/www/someservice//index.php
   proxy_util.c(1963): [client 172.19.0.1:59234] fcgi: found worker fcgi://localhost/var/www/someservice/ for fcgi://localhost/var/www/someservice//index.php
    mod_proxy.c(1161): [client 172.19.0.1:59234] AH01143: Running scheme fcgi handler (attempt 0)
mod_proxy_fcgi.c(879): [client 172.19.0.1:59234] AH01076: url: fcgi://localhost/var/www/someservice//index.php proxyname: (null) proxyport: 0
mod_proxy_fcgi.c(886): [client 172.19.0.1:59234] AH01078: serving URL fcgi://localhost/var/www/someservice//index.php
   proxy_util.c(2157): AH00942: FCGI: has acquired connection for (localhost)
   proxy_util.c(2210): [client 172.19.0.1:59234] AH00944: connecting fcgi://localhost/var/www/someservice//index.php to localhost:8000
   proxy_util.c(2247): [client 172.19.0.1:59234] AH02545: fcgi: has determined UDS as /var/run/php/php5.6-fpm.sock
   proxy_util.c(2419): [client 172.19.0.1:59234] AH00947: connected /var/www/someservice//index.php to httpd-UDS:0
   proxy_util.c(2698): AH02823: FCGI: connection established with Unix domain socket /var/run/php/php5.6-fpm.sock (localhost)
```
