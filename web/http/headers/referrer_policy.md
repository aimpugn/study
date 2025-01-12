# Referrer-Policy

- [Referrer-Policy](#referrer-policy)
    - [Referrer-Policy](#referrer-policy-1)
    - [Directives](#directives)
        - [`no-referrer`](#no-referrer)
        - [`no-referrer-when-downgrade` (기본값)](#no-referrer-when-downgrade-기본값)
        - [`origin`](#origin)
        - [`origin-when-cross-origin`](#origin-when-cross-origin)
        - [`same-origin`](#same-origin)
        - [`strict-origin`](#strict-origin)
        - [`strict-origin-when-cross-origin`(기본값)](#strict-origin-when-cross-origin기본값)
        - [`unsafe-url`](#unsafe-url)
    - [기타](#기타)

## [Referrer-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy)

`Referrer-Policy` 응답 헤더는 얼마나 많은 레퍼러 정보가 요청에 포함되어야 하는지 제어합니다.
HTTP 응답 헤더를 통해 브라우저가 Referrer 정보를 어떻게 처리할지 결정하게 합니다.

```http
HTTP/1.1 200 OK
Referrer-Policy: strict-origin
```

> Referrer?
>
> 현재 페이지에서 다른 페이지로 이동할 때 브라우저가 요청에 포함하는 참조 URL입니다.

HTTP 헤더 외에 HTML에서 정책을 설정할 수도 있습니다.

```html
<meta name="referrer" content="strict-origin" />
```

## Directives

### `no-referrer`

Referrer 정보를 요청 헤더에 포함하지 않습니다.

- `https://example.com/page-a` ->  `https://another.com/page-b`

    ```http
    Referrer: (없음)
    ```

### `no-referrer-when-downgrade` (기본값)

HTTPS -> HTTP로 이동하는 경우 Referrer를 제외하지만,
같은 프로토콜(HTTPS -> HTTPS 또는 HTTP -> HTTP)에서는 Referrer 정보를 포함합니다.

- 같은 프로토콜: `https://example.com/page-a` ->  `https://another.com/page-b`

    ```http
    Referrer: https://example.com/page-a
    ```

- 다운그레이드: `https://example.com/page-a` -> `http://another.com/page-b`

    ```http
    Referrer: (없음)
    ```

### `origin`

Referrer 정보에 페이지 경로를 제외하고 도메인(origin)만 포함합니다.

- `https://example.com/page-a` ->  `https://another.com/page-b`

    ```http
    Referrer: https://example.com
    ```

### `origin-when-cross-origin`

같은 도메인(origin) 요청 시 전체 Referrer 포함하지만,
교차 도메인 요청 시 도메인(origin)만 포함합니다.

- 같은 도메인: `https://example.com/page-a` -> `https://example.com/page-b`

    ```http
    Referrer: https://example.com/page-a
    ```

- 교차 도메인: `https://example.com/page-a` -> `https://another.com/page-b`

    ```http
    Referrer: https://example.com
    ```

### `same-origin`

같은 도메인(origin) 요청에만 Referrer 포함하고, 교차 도메인 요청에는 Referrer를 제외합니다.

- 같은 도메인: `https://example.com/page-a` -> `https://example.com/page-b`

    ```http
    Referrer: https://example.com/page-a
    ```

- 교차 도메인: `https://example.com/page-a` -> `https://another.com/page-b`

    ```http
    Referrer: (없음)
    ```

### `strict-origin`

같은 프로토콜(HTTPS -> HTTPS, HTTP -> HTTP)일 경우 Referrer에 도메인(origin)만 포함합니다.
하지만 HTTPS -> HTTP로 이동 시 Referrer 제외합니다.

- 같은 프로토콜: `https://example.com/page-a` -> `https://another.com/page-b`

    ```http
    Referrer: https://example.com
    ```

- 다운그레이드: `https://example.com/page-a` -> `http://another.com/page-b`

    ```http
    Referrer: (없음)
    ```

### `strict-origin-when-cross-origin`(기본값)

같은 도메인 요청 시 전체 Referrer 포함하고, 교차 도메인 요청 시 도메인(origin)만 포함합니다.
HTTPS -> HTTP로 다운그레이드 시에는 Referrer 제외합니다.

같은 도메인에서는 전체 Referrer를 포함해 편리하고,
교차 도메인 요청 시 도메인만 포함하여 민감한 정보 보호하여,
대부분의 상황에서 권장됩니다.

- 같은 도메인: `https://example.com/page-a` -> `https://example.com/page-b`

    ```http
    Referrer: https://example.com/page-a
    ```

- 교차 도메인: `https://example.com/page-a` -> `https://another.com/page-b`

    ```http
    Referrer: https://example.com
    ```

- 다운그레이드: `https://example.com/page-a` -> `http://another.com/page-b`

    ```http
    Referrer: (없음)
    ```

### `unsafe-url`

모든 요청에서 Referrer 정보를 포함하되, 페이지 경로까지 모두 포함합니다.

- `https://example.com/page-a` -> `https://another.com/page-b`

    ```http
    Referrer: https://example.com/page-a
    ```

## 기타

- [w3c - Referrer Policy](https://www.w3.org/TR/referrer-policy/)
