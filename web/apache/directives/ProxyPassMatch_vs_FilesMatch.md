# ProxyPassMatch vs FilesMatch — 왜 한쪽에서만 `.htaccess` rewrite가 무시되나

## 짧게 답하면

같은 `.php` 요청을 PHP-FPM으로 넘기는 두 방법인데, **프록시로 보내기로 결정하는 시점이 요청 처리 파이프라인의 양 끝에 있습니다.** `ProxyPassMatch`는 URL을 파일 경로로 변환하는 첫 단계(translate)에서 요청을 가로채 파일시스템 매핑 자체를 건너뜁니다 — 디렉터리 워크가 일어나지 않으니 `.htaccess`는 읽히지도 않고, 그 안의 rewrite와 authz가 통째로 무시됩니다. 반면 `<FilesMatch>` + `SetHandler`는 매핑 → `.htaccess` 병합 → per-dir rewrite가 모두 끝난 뒤 마지막 핸들러 단계에서 작동하므로, rewrite가 끝난 최종 경로가 FPM에 전달됩니다. 그래서 `.htaccess`의 rewrite에 의존하는 앱(CakePHP류)은 FilesMatch 방식이어야 동작합니다.

## 무대 — 무엇을 실험했나

`/var/www/someservice`에 배치된 CakePHP 2.x 구조의 앱입니다. 루트의 `.htaccess`가 모든 요청을 `app/webroot/`로 보내는 rewrite를 갖고 있고(아래 로그에 패턴 `^$`, `(.*)`, 조건 `!-d`, `!-f`가 그대로 찍힙니다), PHP는 유닉스 도메인 소켓의 php5.6-fpm입니다. 동일한 `GET /index.php` 요청을 두 가지 구성으로 처리시키고 `LogLevel` trace 로그를 대조했습니다. 실험에 쓴 vhost는 [some_service.conf](../sites/some_service.conf)에 보존돼 있습니다 — 주석 처리된 `ProxyPassMatch` 두 줄이 기각된 구성이고, 채택된 것은 `<FilesMatch .php$> SetHandler ...`입니다.

## 요청이 FPM에 닿기까지 — 두 갈래 길

Apache는 요청 하나를 여러 단계(훅)로 나눠 처리합니다. 두 디렉티브가 끼어드는 지점이 다음과 같이 다릅니다.

```text
GET /index.php
  |
  v
[translate]  URL -> 파일 경로 변환
  |   * 서버/vhost 컨텍스트의 mod_rewrite가 도는 단계
  |   * ProxyPassMatch가 도는 단계  <- 패턴이 맞으면 r->filename = "proxy:fcgi://..."
  |     이후의 파일시스템 매핑 전체를 건너뛴다
  v
[map_to_storage]  DocumentRoot 아래로 디렉터리 워크
  |   * 경로상 각 디렉터리의 .htaccess를 읽어 병합 (AllowOverride 허용 범위에서)
  |   * proxy:로 바뀐 요청은 여기 올 일이 없다 -> .htaccess가 읽히지 않는 이유
  v
[authz]  병합된 Require 지시어 평가
  v
[fixup]  per-dir(.htaccess) mod_rewrite가 도는 단계
  |   * 경로가 바뀌면 내부 리다이렉트로 파이프라인 처음부터 재진입
  v
[handler]  최종 파일을 누가 서빙할지 결정
      * SetHandler "proxy:unix:...|fcgi://..." 가 발동하는 단계
      * rewrite가 끝난 최종 경로가 SCRIPT_FILENAME으로 FPM에 전달된다
```

mod_rewrite 공식 문서가 이 비대칭을 명시합니다 — 서버 컨텍스트 규칙은 URL-to-filename 단계에서, per-directory 규칙은 "최종 데이터 디렉터리가 정해진 뒤"인 fixup 단계에서 적용됩니다. fixup 시점에는 URL이 이미 파일 경로로 매핑된 뒤이므로, 규칙이 경로를 바꾸면 그 결과를 다시 매핑하기 위해 파이프라인 처음부터 재진입하게 됩니다(아래 로그의 INTERNAL REDIRECT가 그 재진입입니다).

## 증거 — 같은 요청, 두 로그

로그는 가독성을 위해 타임스탬프와 클라이언트 필드를 생략한 발췌입니다(전체 로그는 git 이력의 원 노트에 있습니다).

### ProxyPassMatch 구성 — rewrite 흔적이 아예 없다

```text
mod_authz_core.c(835): authorization result: granted (no directives)
    # per-dir 설정이 병합되지 않았다는 부수 증거.
    # 아래 FilesMatch 로그에서는 같은 자리에 "Require all granted" 평가가 찍힌다.
mod_proxy_fcgi.c(52):  canonicalising URL //localhost/var/www/someservice//index.php
mod_proxy_fcgi.c(84):  AH01060: set r->filename to proxy:fcgi://localhost/var/www/someservice//index.php
    # translate 단계에서 r->filename이 곧장 proxy:... 로 설정됐다.
    # 파일시스템 경로를 거치지 않았으므로 mod_rewrite [perdir] 줄이 로그 전체에 0개.
mod_proxy_fcgi.c(886): AH01078: serving URL fcgi://localhost/var/www/someservice//index.php
    # .htaccess의 app/webroot/ rewrite가 반영되지 않은 경로가 그대로 FPM으로 갔다.
    # 겹슬래시(//index.php)는 아래 별도 절에서 다룬다.
```

같은 노트의 관찰 하나가 메커니즘을 한 번 더 확인해 줍니다 — **php가 아닌 요청은 `ProxyPassMatch` 패턴에 걸리지 않으므로** 정상적으로 파일시스템 매핑을 타고, 그때는 `.htaccess` rewrite가 동작합니다. 즉 "rewrite가 고장난" 것이 아니라, 패턴에 걸린 요청만 파이프라인 앞단에서 빠져나간 것입니다.

### FilesMatch + SetHandler 구성 — rewrite 2-pass 후 핸들러에서 프록시

```text
mod_authz_core.c(809): AH01626: authorization result of Require all granted: granted
    # per-dir 설정이 병합돼 평가되고 있다 (.htaccess가 읽혔다는 증거).
mod_rewrite.c(476): [perdir /var/www/someservice/] applying pattern '(.*)' to uri 'index.php'
mod_rewrite.c(476): [perdir /var/www/someservice/] rewrite 'index.php' -> 'app/webroot/index.php'
    # 루트 .htaccess의 CakePHP 규칙이 fixup 단계에서 적용됐다.
mod_rewrite.c(476): [perdir /var/www/someservice/] internal redirect with /app/webroot/index.php [INTERNAL REDIRECT]
    # per-dir rewrite는 결과 경로로 파이프라인을 처음부터 다시 탄다 (2-pass).
mod_rewrite.c(476): [perdir .../app/webroot/] RewriteCond input='...' pattern='!-d' => matched
mod_rewrite.c(476): [perdir .../app/webroot/] RewriteCond input='...' pattern='!-f' => not-matched
    # 2번째 pass: webroot .htaccess의 "디렉터리도 파일도 아닐 때만 프런트 컨트롤러로" 조건 중
    # !-f 쪽이 깨졌다(실제 파일이므로) -> rewrite 없이 pass through.
proxy_util.c(1986): *: using default reverse proxy worker for unix:/var/run/php/php5.6-fpm.sock|fcgi://localhost/... (no keepalive)
    # 핸들러 단계에 와서야 SetHandler의 프록시가 발동했다.
    # "no keepalive"는 mod_proxy_fcgi가 기본으로 커넥션 재사용을 끄기 때문 (아래 비교표 참고).
mod_proxy_fcgi.c(886): AH01078: serving URL fcgi://localhost/var/www/someservice/app/webroot/index.php
    # rewrite가 완료된 최종 경로가 FPM에 전달됐다.
```

## 겹슬래시는 어디서 왔나

ProxyPassMatch 쪽 serving URL은 `/var/www/someservice//index.php`입니다. 겹슬래시는 정규식 캡처와 타깃 문자열의 결합 아티팩트입니다 — 타깃이 `/`로 끝나는데 캡처 그룹에 선행 `/`까지 포함되면 `…someservice/` + `/index.php`로 이어 붙습니다. 실험 당시의 정확한 설정 줄은 보존돼 있지 않아 이 결합 형태는 추정이지만, 겹슬래시 자체는 로그에 찍힌 관측 사실입니다. 로그상 FPM 연결까지는 진행됐으므로 경로 정규화는 FPM 쪽 처리에 맡겨진 셈인데, 정규식 치환으로 백엔드 경로를 조립하는 방식이 이런 오타·중복 슬래시에 취약하다는 신호로 읽으면 됩니다. SetHandler 방식은 서버가 매핑한 `r->filename`을 그대로 쓰므로 이 문제가 구조적으로 없습니다. 참고로 같은 vhost의 `AllowEncodedSlashes NoDecode`는 인코딩된 슬래시(`%2F`)를 디코딩하지 않고 보존하는 별개 층위의 설정이라, 여기서 관찰된 리터럴 겹슬래시(`//`)와는 무관합니다 — 이름이 비슷해 혼동하기 쉬운 지점입니다.

## 축별 비교

| 축 | ProxyPassMatch | FilesMatch + SetHandler |
| --- | --- | --- |
| 프록시 결정 시점 | translate (파이프라인 최전단) | handler (최후단) |
| 파일시스템 매핑 | 안 함 — URL 패턴만 본다 | 함 — DocumentRoot 기준 실제 경로 |
| `.htaccess` 읽힘 | 안 읽힘 (디렉터리 워크 없음) | 정상 병합 (`AllowOverride` 전제) |
| per-dir rewrite | 미적용 | 적용 (fixup, 내부 리다이렉트 2-pass) |
| per-dir authz | 미적용 — 로그에 `granted (no directives)` | 적용 — 로그에 `Require all granted` 평가 |
| 백엔드에 주는 경로 | 정규식 치환 결과 (겹슬래시·오타 위험) | 매핑이 끝난 `r->filename` — 공식 문서도 "가장 정확한 PATH_INFO 계산"을 이 방식의 장점으로 명시 |
| 버전 요구 | 구형 2.4에서도 가능 | `SetHandler proxy:` 는 2.4.10+, UDS는 2.4.9+ |
| FPM 커넥션 재사용 | 기본 꺼짐 (mod_proxy_fcgi 공통) | 기본 꺼짐 — 로그의 `(no keepalive)` |
| 적합한 용도 | 로컬 콘텐츠 없는 패턴 단위 패스스루 | `.htaccess` 의존 앱, 일반 권장 |

커넥션 재사용 축의 보충: 켜려면 명시적 워커에 `enablereuse=on`을 줘야 하는데, 공식 문서가 PHP-FPM 조합의 함정을 경고합니다 — FPM 워커 수를 잘못 잡으면 **유휴 지속 커넥션을 물고 있는 것만으로 FPM 워커가 전부 "busy"가 되어** 신규 요청이 타임아웃으로 쌓일 수 있습니다. 재사용을 켤 때는 httpd 자식 수 × 커넥션 풀과 FPM `pm.max_children`을 같이 계산해야 합니다.

## 실전 반영과 선택 기준

[some_service.conf](../sites/some_service.conf)가 이 실험의 결론입니다.

```apache
# ProxyPassMatch ^/(.*\.php(/.*)?)$ fcgi://127.0.0.1:9000/var/www/vanilla/$1   # 기각 — .htaccess 무시
<FilesMatch .php$>
  SetHandler "proxy:unix:/var/run/php/php5.6-fpm.sock|fcgi://localhost/"       # 채택
</FilesMatch>
<Directory /PATH/TO/SERVICE>
    AllowOverride All        # .htaccess 병합의 전제 조건 — 이게 None이면 FilesMatch 방식이어도 rewrite가 안 돈다
</Directory>
```

선택 기준을 일반화하면: 앱이 자체 `.htaccess` rewrite·인증 규칙에 의존하면 두 방식 중에서는 FilesMatch + SetHandler여야 하고(translate에서 가로채는 방식으로는 그 규칙들이 아예 실행되지 않으므로), 로컬 콘텐츠가 전혀 없는 순수 패스스루(예: 특정 패턴을 통째로 별도 백엔드에 위임)라면 ProxyPassMatch도 충분합니다.

## 한계와 주의

- `AllowOverride All`은 공짜가 아닙니다. 요청 경로상 모든 디렉터리에서 매 요청마다 `.htaccess`를 탐색·해석하고, 운영 중 누구든 파일 하나로 서버 동작을 바꿀 수 있는 표면이 됩니다. 공식 문서의 일반 권고는 가능하면 규칙을 서버 설정의 `<Directory>` 블록으로 옮기고 `AllowOverride None`으로 닫는 것입니다 — 이 실험 구성은 `.htaccess`를 앱이 들고 다니는 CakePHP 관례를 따르느라 All을 받아들인 경우입니다.
- 이 문서의 로그는 httpd 2.4 + php5.6-fpm(UDS) 환경의 실측입니다. 단계 구조(translate/fixup/handler)는 2.4 전반에 적용되는 구조적 사실이지만, 로그 포맷·줄 번호는 버전에 따라 다를 수 있습니다.

## 직접 확인해 보기

자기 서버가 어느 쪽으로 동작하는지 가르는 판정 경로입니다.

1. trace 로그를 켭니다: `LogLevel debug rewrite:trace4 proxy:trace2` (운영에서는 일시적으로만).
2. `.php` 요청을 하나 보냅니다: `curl -s http://HOST/index.php -o /dev/null`.
3. error 로그에서 세 가지를 봅니다.

| 관찰 | 해석 |
| --- | --- |
| `[perdir ...]` rewrite 줄이 있다 | `.htaccess`가 적용되는 구성 (FilesMatch 경로) — PASS |
| rewrite 줄이 없고 `set r->filename to proxy:...`가 바로 찍힌다 | translate에서 가로채는 구성 (ProxyPassMatch 경로) — `.htaccess` 의존 앱이면 FAIL |
| `AH01078: serving URL ...`의 최종 경로 | rewrite 반영 여부를 한 줄로 판정 (기대 경로 = rewrite 결과 경로) |

## 출처

- [mod_proxy_fcgi — 예제·SetHandler 방식·enablereuse](https://httpd.apache.org/docs/2.4/mod/mod_proxy_fcgi.html) — SetHandler proxy 2.4.10+, UDS 2.4.9+, 커넥션 재사용 기본 꺼짐, FPM busy 경고
- [mod_rewrite Technical Details — API Phases](https://httpd.apache.org/docs/2.4/rewrite/tech.html) — 서버 컨텍스트 = URL-to-filename, per-dir = fixup, 내부 리다이렉트의 이유
- [mod_proxy — ProxyPassMatch](https://httpd.apache.org/docs/2.4/mod/mod_proxy.html#proxypassmatch)
- [.htaccess How-To — 성능·병합 비용](https://httpd.apache.org/docs/2.4/howto/htaccess.html)
- 원 관찰 기록의 참고 링크: [Stack Overflow — PHP-FPM FilesMatch and ProxyPassMatch interchangeability](https://stackoverflow.com/questions/34350248/php-fpm-filesmatch-and-proxypassmatch-interchangeability), [ma.ttias.be — ProxyPass for PHP taking precedence over Files/FilesMatch in htaccess](https://ma.ttias.be/apache-2-4-proxypass-for-php-taking-precedence-over-filesfilesmatch-in-htaccess/)
