# Apache 정보 노출과 위험 메서드 하드닝 — 무엇을, 왜, 어디서 막나

짧게 답하면, 두 가지를 막는 이야기입니다. 하나는 **서버가 응답 헤더와 에러 페이지에 자기 정체(버전·OS·모듈)를 떠벌리는 것**이고, 다른 하나는 **`TRACE` 같은 진단용 메서드가 켜져 있는 것**입니다. 둘 다 본질은 같습니다 — *공격자의 정찰(reconnaissance)을 도와주는 표면을 줄인다*. 그리고 마지막 질문, "origin(Apache)에서 막을까 앞단(WAF)에서 막을까"는 사실 둘 중 하나를 고르는 문제가 아니라 **신뢰 경계를 어디에 긋고 어느 계층에 무슨 책임을 줄까**의 문제입니다.

이 문서는 먼저 "왜 이게 위험으로 분류되는가"를 짚고(그 위험이 과장된 부분도 같이), 각 설정이 Apache 내부에서 어떻게 동작하는지, 그리고 왜 어떤 것은 origin에서, 어떤 것은 앞단에서 막는 게 더 자연스러운지를 설명합니다. 바로 복사해 쓸 최소 설정은 맨 끝에 정리해 둡니다.

## 큰 그림 — 정보는 어디서 새고, 어디서 막히나

요청 하나가 클라이언트에서 origin까지 갔다 오는 경로를 펼치면, 정보 노출과 메서드 처리의 "지점"이 보입니다.

```text
[클라이언트]
   |   응답 헤더의 Server:, X-Powered-By: 를 본다
   |   에러 페이지 푸터의 서버 서명을 본다
   v
[BIG-IP / WAF]   <- 1차 차단·헤더 가공이 가능한 지점 (중앙)
   |   여기서 TRACE를 끊거나 Server 헤더를 지울 수 있다
   v
[Apache (origin)]  <- 정보의 발원지, 최종 방어선
   |   Server 헤더를 만들어 붙이는 곳
   |   TRACE를 코어에서 처리하는 곳
   v
[애플리케이션 / PHP]  <- X-Powered-By 같은 헤더의 또 다른 발원지
```

핵심은 **정보가 새어 나오는 발원지는 origin과 그 위의 앱이고, 그것을 가릴 수 있는 지점은 발원지와 앞단 둘 다**라는 점입니다. 이 비대칭이 뒤에서 "어디서 막을까" 논의의 출발점이 됩니다.

## 왜 정보 노출이 위험으로 분류되나 — 그리고 그 한계

공격은 보통 정찰에서 시작합니다. 대상이 `Apache/2.4.49`라고 스스로 밝히면, 공격자는 그 버전에 걸리는 알려진 취약점(CVE)을 곧장 매칭해 볼 수 있습니다. `mod_ssl`, `OpenSSL`, `PHP` 버전까지 흘리면 매칭 표면은 더 넓어집니다. 자동화된 스캐너와 봇은 이 배너 문자열만 보고 공격 대상 목록을 추립니다.

그런데 여기서 균형을 잡아야 합니다. **버전을 숨긴다고 취약점이 사라지지는 않습니다.** 버전 헤더 가리기는 전형적인 "security through obscurity"이고, 그 자체가 방어의 본질일 수 없습니다. 본질은 패치입니다. 그러면 왜 그래도 하느냐 —

- **비용이 거의 0**입니다. 설정 한두 줄로, 잃을 게 없습니다.
- **자동화 노이즈를 줄입니다.** 버전 기반으로 무차별 매칭하는 봇·스캐너의 1차 필터를 피합니다(표적 공격은 못 막지만, 대량 스캔의 노이즈는 줄어듭니다).
- **컴플라이언스 항목**입니다. PCI-DSS 등 보안 점검에서 배너 노출은 단골 지적 사항이라, 점검을 통과하려면 어차피 꺼야 합니다.

즉 정보 노출 차단은 "이걸로 안전해진다"가 아니라 **"공짜로 줄일 수 있는 표면은 줄인다"**는 위생(hygiene) 차원으로 이해하는 게 정확합니다.

## Server 헤더는 무엇이고 어떻게 구성되나 — `ServerTokens`

`Server` 응답 헤더는 Apache가 모든 응답에 붙이는 자기소개입니다. 기본값(`Full`)에서는 이렇게 나갑니다.

```http
Server: Apache/2.4.41 (Unix) OpenSSL/1.1.1 PHP/7.4.3
```

버전, OS, 그리고 컴파일된 모듈까지 다 들어 있습니다. `ServerTokens` 디렉티브가 이 문자열의 상세도를 결정합니다. 값에 따른 출력은 다음과 같습니다(공식 문서 기준).

| 값 | `Server` 헤더 출력 | 노출 정보 |
| --- | --- | --- |
| `Full` (기본값) | `Apache/2.4.X (Unix) OpenSSL/... PHP/...` | 버전 + OS + 모듈 |
| `OS` | `Apache/2.4.X (Unix)` | 버전 + OS |
| `Minimal` | `Apache/2.4.X` | 전체 버전 번호 |
| `Minor` | `Apache/2.4` | 메이저.마이너 |
| `Major` | `Apache/2` | 메이저 버전만 |
| `Prod` | `Apache` | 제품명만 |

정보량이 많은 순서로 `Full > OS > Minimal > Minor > Major > Prod`입니다. (`Minimal`이 전체 버전 `2.4.X`를, `Minor`가 `2.4`까지만 노출하므로 `Minimal`이 더 많이 흘립니다 — 이름 때문에 헷갈리기 쉬운 부분입니다.) 권장은 가장 적게 노출하는 `Prod`입니다.

```apache
ServerTokens Prod
```

**왜 이 디렉티브는 `<VirtualHost>` 안에서 안 먹나.** `ServerTokens`의 적용 컨텍스트는 **server config 전용**입니다. 가상 호스트 블록 안에 넣어도 무시됩니다. 이유를 추론하면, `Server` 배너는 서버가 뜰 때 전역으로 한 번 구성되는 토큰이라 요청이 어느 가상 호스트로 가는지와 무관하게 하나로 고정되기 때문입니다. 그래서 반드시 메인 설정(`httpd.conf` / `apache2.conf`)의 전역 영역에 둬야 합니다.

**함정 하나.** `Prod`로 줄여도 `Server: Apache`라는 토큰은 끝까지 남습니다. `ServerTokens`만으로는 이 헤더를 완전히 없앨 수 없습니다. 완전 제거는 뒤의 "mod_headers의 한계" 절에서 다룹니다.

## 에러 페이지의 서명은 별개 표면 — `ServerSignature`

`ServerTokens`가 헤더를 다룬다면, `ServerSignature`는 **본문**을 다룹니다. Apache가 스스로 만들어 내는 문서(404 같은 에러 페이지, 디렉터리 목록 등)의 맨 아래에 찍히는 한 줄짜리 서명입니다.

```text
Apache/2.4.41 (Unix) Server at example.com Port 80
```

이건 헤더와 다른 경로로 새는 정보라 따로 꺼야 합니다.

```apache
ServerSignature Off
```

값은 `On`(서명 표시), `Off`(숨김), `EMail`(서명 + 관리자 mailto 링크) 셋입니다. Apache 2.4 기본값은 `Off`지만, 배포판이나 기존 설정에서 `On`으로 바뀌어 있을 수 있으니 명시적으로 꺼 두는 게 안전합니다. `ServerSignature`는 `ServerTokens`와 달리 server config·virtual host·directory·`.htaccess`에서 모두 설정할 수 있는데(FileInfo override), 이는 서명이 본문에 찍히는 것이라 디렉터리 단위로 다르게 둘 여지가 있기 때문입니다.

## Server 헤더를 "완전히" 지우려면 — `mod_headers`의 한계

많은 사람이 여기서 막힙니다. 직관적으로는 이렇게 하면 될 것 같습니다.

```apache
# 동작하지 않는다
Header unset Server
Header always unset Server
```

그런데 `mod_headers`로는 `Server` 헤더가 지워지지 않습니다. 이유는 **헤더가 응답에 추가되는 순서(파이프라인 시점)** 에 있습니다.

```text
요청 처리
  -> 핸들러가 본문 생성
  -> mod_headers의 fixup 단계: 여기서 Header unset/set 이 실행됨
  -> ... 그 다음 ...
  -> 코어 출력 필터가 Server: 헤더를 붙임   <- mod_headers보다 나중!
  -> 클라이언트로 전송
```

`mod_headers`가 헤더를 손보는 시점은 코어가 `Server`를 붙이기 *전*입니다. 아직 존재하지 않는 헤더를 unset 해 봐야 소용이 없고, 이후 코어가 다시 붙여 버립니다. (반면 `X-Powered-By`처럼 코어가 아니라 앱/모듈이 만든 헤더는 `mod_headers`로 잘 지워집니다. 발원지가 코어가 아니기 때문입니다.)

origin에서 `Server`를 끝까지 손보려면 보통 **ModSecurity**의 `SecServerSignature` 지시어를 씁니다. 이건 코어가 헤더를 붙인 *이후* 단계에서 값을 덮어쓰는 방식이라, 완전 제거는 아니어도 임의 문자열로 위장할 수 있습니다. 다만 이걸 위해 ModSecurity를 새로 들이는 건 배보다 배꼽입니다. **이 한계가, "그러면 차라리 앞단(BIG-IP)에서 지우자"는 결론으로 자연스럽게 이어집니다** — 뒤에서 다시 봅니다.

PHP를 쓴다면 `Server`와 별개로 `X-Powered-By: PHP/7.4.3`가 새는데, 이건 Apache가 아니라 PHP 설정에서 끕니다.

```ini
; php.ini
expose_php = Off
```

## `TRACE`는 무엇이고 왜 위험하다고 하나 — XST

`TRACE`를 이해하려면 HTTP 메서드의 성격부터 짚는 게 좋습니다. 메서드에는 *안전성(safe, 서버 상태를 안 바꿈)* 과 *멱등성(idempotent, 여러 번 해도 결과 같음)* 이라는 분류가 있습니다. `TRACE`는 안전하고 멱등한 **진단용** 메서드입니다. 원래 목적은 프록시 체인 디버깅입니다 — 클라이언트가 보낸 요청이 여러 프록시를 거치며 어떻게 변형되는지 보려고, 최종 서버가 **받은 요청을 그대로 응답 본문에 되돌려(echo)** 줍니다.

이 "그대로 echo" 특성이 공격에 악용된 게 **XST(Cross-Site Tracing)** 입니다. 2003년 Jeremiah Grossman이 발표했고, 당시 Microsoft가 IE6 SP1에 도입한 `HttpOnly`(자바스크립트가 쿠키를 못 읽게 하는 플래그)를 우회하려는 기법이었습니다. 메커니즘은 이렇습니다.

```text
1. 공격자가 XSS로 피해자 브라우저에서 자바스크립트를 실행시킨다.
2. 그 스크립트가 서버로 TRACE 요청을 보낸다.
3. 서버는 받은 요청 헤더를 그대로 응답 본문에 echo 한다.
   - 이 헤더에는 Cookie:, Authorization: 이 포함되어 있다.
4. 스크립트는 응답 "본문"을 읽는다 (쿠키를 직접 읽는 게 아니라서 HttpOnly 우회).
   -> HttpOnly 쿠키, 인증 헤더가 탈취된다.
```

`HttpOnly`는 "자바스크립트가 쿠키를 *읽는* 것"을 막지만, TRACE 응답에서는 쿠키가 *본문 텍스트*로 돌아오므로 그 방어를 우회한다는 게 핵심입니다.

**다만 현대의 실효 위험은 상당히 낮습니다.** 오늘날 브라우저는 자바스크립트(`fetch`/`XMLHttpRequest`)로 `TRACE`를 보내는 것을 금지된 메서드(forbidden method)로 막아 두었습니다. 그래서 위 시나리오의 2단계가 브라우저만으로는 성립하기 어렵습니다(과거 Java 플러그인 등 다른 경로가 발견된 적은 있으나, 그런 실행 환경 자체가 거의 사라졌습니다). 그럼에도 `TRACE`를 끄는 이유는 정보 노출과 동일합니다 — **컴플라이언스 점검의 단골 항목**이고, **쓰지도 않는 진단 기능이라 닫아 두는 게 위생적**이기 때문입니다. "현재 위험이 낮다"와 "그래도 닫는다"는 모순이 아닙니다.

## `TRACE` 차단 — `TraceEnable`과 그 대안

Apache에는 전용 디렉티브가 있습니다. 기본값이 `On`이므로 명시적으로 꺼야 합니다.

```apache
TraceEnable off
```

적용 컨텍스트는 server config와 virtual host입니다. 끄면 `TRACE` 요청에 `405 Method Not Allowed`를 돌려줍니다.

**왜 `mod_rewrite`보다 이게 정석인가.** 인터넷에는 이런 식의 예제도 많습니다.

```apache
RewriteCond %{REQUEST_METHOD} ^TRACE
RewriteRule .* - [F]
```

동작은 하지만, `TRACE`는 Apache **코어가 특수하게 처리하는** 메서드입니다. `TraceEnable off`는 코어 레벨에서 끊으므로 rewrite 핸들러에 도달하기 전에 확실히 차단됩니다. 핸들러 기반 우회보다 누수 가능성이 적습니다.

값 중 `extended`는 디버깅용으로 `TRACE`에 메시지 본문을 허용하는 옵션인데, **보안상 절대 운영에 쓰지 않습니다.** 운영 값은 `off` 하나입니다.

`TRACK`(IIS 계열의 `TRACE` 변형)도 스캐너가 같이 지적하곤 하는데, Apache는 `TRACK`을 모르는 메서드로 처리하므로 별도 조치는 보통 불필요합니다.

허용 메서드 자체를 화이트리스트로 좁히고 싶다면 `mod_allowmethods`나 `<LimitExcept>`를 쓸 수 있습니다(예: `GET POST HEAD`만 허용). 단 `TRACE`는 코어가 처리하므로 `TraceEnable off`가 우선이고, 이쪽은 그 위에 다른 메서드까지 통제하고 싶을 때의 확장입니다.

## 어디서 막을 것인가 — origin vs WAF, 그리고 계층 방어

이제 마지막 질문입니다. BIG-IP WAF가 앞에 있는데, 거기서 막는 게 더 올바른가?

**결론: 둘 중 하나가 아니라 둘 다입니다(defense in depth). 다만 계층마다 역할이 다릅니다.** "WAF가 있으니 origin은 그냥 둬도 된다"는 권장하지 않습니다.

핵심 개념은 **신뢰 경계(trust boundary)** 입니다. origin이 "내 앞에는 항상 WAF가 있으니 안전하다"고 가정하는 순간, 그 가정이 깨지는 모든 경로가 무방비가 됩니다. WAF는 우회 가능한 계층이기 때문입니다.

- **내부망·관리망에서 BIG-IP를 거치지 않고 Apache에 직접 접근**하는 경로가 흔히 존재합니다(모니터링, 배포, 동일 서브넷의 다른 호스트).
- **BIG-IP 정책 변경·마이그레이션·장애** 시 origin이 그대로 노출됩니다.
- **컴플라이언스 스캐너가 origin을 직접 겨냥**하면 WAF 차단과 무관하게 origin 설정을 봅니다.

그래서 **`TraceEnable off`는 비용이 사실상 0(한 줄)이고 origin 자체를 안전하게 만드는 최종 방어선이므로 무조건 둡니다.** 그 위에 WAF 차단을 얹는 게 방어 심층화입니다.

반대로 WAF에서 막는 것의 고유한 가치도 분명합니다.

- 백엔드에 **도달하기 전에 차단**해 공격 표면과 부하를 줄입니다.
- 백엔드가 여러 대여도 **정책·로깅을 한 곳에서** 관리합니다(중앙 집중).

흥미로운 비대칭이 하나 있습니다. 앞서 본 **`Server` 헤더 완전 제거는 오히려 WAF가 더 잘합니다.** origin의 `mod_headers`로는 코어가 붙이는 `Server`를 못 지운다고 했는데, BIG-IP는 응답이 origin을 떠난 *뒤* 클라이언트로 가기 전에 가공하므로 그 한계가 없습니다. 그래서 역할을 이렇게 나누는 게 가장 깔끔합니다.

| 항목 | Apache (origin) | BIG-IP (앞단) |
| --- | --- | --- |
| 버전 정보 축소 | `ServerTokens Prod` / `ServerSignature Off` | — |
| `Server` 헤더 완전 제거 | (코어 한계로 어려움) | iRule / 프로파일로 깔끔하게 제거 |
| `TRACE` 차단 | `TraceEnable off` (최종 방어선) | ASM Allowed Methods 또는 iRule (1차 차단) |

BIG-IP 쪽 구현은 세 갈래입니다.

- **ASM(WAF) 모듈이 켜져 있다면** — 보안 정책의 Allowed Methods(허용 HTTP 메서드) 목록이 정공법입니다. `TRACE`는 기본적으로 비허용이라, 정책이 enforcement(차단) 모드인지만 확인하면 됩니다.
- **LTM iRule로 메서드 차단** (버전에 따라 문법 차이가 있을 수 있어 예시 수준입니다):

  ```tcl
  when HTTP_REQUEST {
      if { [string toupper [HTTP::method]] equals "TRACE" } {
          HTTP::respond 405 content "Method Not Allowed"
      }
  }
  ```

- **LTM iRule로 응답 헤더 제거**:

  ```tcl
  when HTTP_RESPONSE {
      HTTP::header remove "Server"
      HTTP::header remove "X-Powered-By"
  }
  ```

  (헤더 가공은 iRule 외에 HTTP profile의 헤더 삽입/삭제 기능으로도 가능합니다.)

**운영상 주의 하나.** 앞단에서 헤더를 손대거나 메서드를 차단하면 헬스 모니터나 기존 연동이 영향을 받을 수 있습니다. 예컨대 일부 모니터가 특정 메서드 응답이나 `Server` 헤더를 검사하도록 설정돼 있을 수 있습니다. 적용 전 스테이징에서 검증하는 게 안전합니다.

## 정리 — 최소 설정 요약

학습 맥락을 걷어내고 바로 적용할 것만 추리면 이렇습니다.

origin(Apache), 메인 설정의 전역 영역:

```apache
ServerTokens Prod
ServerSignature Off
TraceEnable off
```

PHP를 쓴다면 `php.ini`:

```ini
expose_php = Off
```

앞단(BIG-IP)에서 추가 계층으로:

- `Server`·`X-Powered-By` 응답 헤더 제거 (origin이 못 지우는 `Server`를 여기서 마무리)
- `TRACE`는 ASM Allowed Methods 또는 iRule로 1차 차단

한 줄 요약: **origin 하드닝은 "잃을 게 없는 위생 + 최종 방어선"이라 무조건 하고, WAF는 그 앞에 중앙 집중 차단·헤더 가공 계층으로 얹는다. 어느 한쪽만 믿지 않는다.**

## 직접 확인해 볼 수 있는 포인트

- 현재 노출 상태 확인: `curl -I https://대상/` 로 `Server`, `X-Powered-By` 헤더를 봅니다.
- `TRACE` 상태 확인: `curl -i -X TRACE https://대상/` — 차단되면 `405`, 열려 있으면 `200`과 함께 요청이 echo 됩니다.
- origin 직접 확인: WAF를 우회해 origin IP로 직접 같은 요청을 보내, 앞단이 아니라 origin 자체가 막고 있는지 검증합니다(이것이 "최종 방어선" 여부를 가르는 테스트입니다).

## 출처

- [Apache `core` — ServerTokens](https://httpd.apache.org/docs/2.4/mod/core.html#servertokens)
- [Apache `core` — ServerSignature](https://httpd.apache.org/docs/2.4/mod/core.html#serversignature)
- [Apache `core` — TraceEnable](https://httpd.apache.org/docs/2.4/mod/core.html#traceenable)
- [Apache `mod_headers`](https://httpd.apache.org/docs/2.4/mod/mod_headers.html)
- [OWASP — Cross Site Tracing (XST)](https://owasp.org/www-community/attacks/Cross_Site_Tracing)
- [OWASP WSTG — Test HTTP Methods](https://owasp.org/www-project-web-security-testing-guide/stable/4-Web_Application_Security_Testing/02-Configuration_and_Deployment_Management_Testing/06-Test_HTTP_Methods)
- [RFC 9110 — HTTP Semantics (TRACE, §9.3.8)](https://www.rfc-editor.org/rfc/rfc9110.html#name-trace)
- [F5 iRules — HTTP::respond](https://clouddocs.f5.com/api/irules/HTTP__respond.html)
- [F5 iRules — HTTP::header](https://clouddocs.f5.com/api/irules/HTTP__header.html)
- ModSecurity `SecServerSignature` (Reference Manual) — `Server` 헤더 위장이 필요할 때
