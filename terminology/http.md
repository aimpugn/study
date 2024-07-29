# HTTP

- [HTTP](#http)
    - [HTTP](#http-1)
    - [Headers](#headers)
        - [`Content-Security-Policy`](#content-security-policy)
    - [Response Status Code](#response-status-code)

## HTTP

## Headers

### `Content-Security-Policy`

- 웹 페이지에 대한 보안 정책을 브라우저에 알려, 공격자가 페이지를 악용하는 것을 방지하는데 사용
- 각 지시어(directive)는 특정 유형의 리소스에 대해 허용되는 출처를 지정
- CSP를 사용하면 XSS(Cross-Site Scripting) 공격과 데이터 주입 공격에 대한 보호를 강화할 수 있다
- 이 태그는 페이지 콘텐츠가 로드될 수 있는 출처를 제한함으로써 보안을 향상시키는 데 목적이 있다

```html
<meta 
    http-equiv="Content-Security-Policy" 
    content="default-src 'self' data: https://ssl.gstatic.com 'unsafe-eval'; 
            style-src 'self' 'unsafe-inline'; 
            media-src *; 
            img-src 'self' data: content:;">
```

- `default-src 'self' data: https://ssl.gstatic.com 'unsafe-eval'`:
    - `default-src`: 기본적으로 모든 리소스 타입에 대한 기본 정책을 설정
    - `'self'`: 리소스가 현재 출처(같은 도메인)에서만 로드될 수 있음을 나타낸다
    - `data:`: 데이터 URI를 통한 리소스 로드를 허용
    - `https://ssl.gstatic.com`: `ssl.gstatic.com` 도메인에서 리소스를 로드할 수 있다
    - `'unsafe-eval'`: 스크립트에서 `eval()`과 같은 방법을 사용하여 문자열을 코드로 변환하는 것을 허용한다. 이는 보안 위험을 증가시킬 수 있어 주의가 필요함.
- `style-src 'self' 'unsafe-inline'`:
    - `style-src`: 스타일 시트의 출처를 지정
    - `'unsafe-inline'`: 인라인 스타일(`<style>` 또는 `style` 속성) 사용을 허용. 이 역시 보안 취약점을 야기할 수 있다
- `media-src *`:
    - `media-src`: 미디어 파일(오디오, 비디오)의 출처를 지정
    - `*`: 모든 도메인에서 미디어 리소스를 로드할 수 있음을 나타낸다
- `img-src 'self' data: content:`:
    - `img-src`: 이미지 리소스의 출처를 지정한다
    - `content:`: 이는 표준 CSP 지시어가 아니며, 잘못 입력되었을 가능성이 있습니다. 올바르게는 `content:`가 아닌 해야 합니다.

## [Response Status Code](https://datatracker.ietf.org/doc/html/rfc7231#section-6)

- [Hypertext Transfer Protocol (HTTP) Status Code Registry](https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml)
