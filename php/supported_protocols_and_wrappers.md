# [Supported Protocols and Wrappers](https://www.php.net/manual/en/wrappers.php)

- [Supported Protocols and Wrappers](#supported-protocols-and-wrappers)
    - [wrappers?](#wrappers)
    - [지원하는 프로토콜 리스트](#지원하는-프로토콜-리스트)
    - [`php://`](#php)
        - [`php://input`이란?](#phpinput이란)
        - [동작 방식](#동작-방식)
        - [사용 방식](#사용-방식)
        - [값이 있거나 없는 경우](#값이-있거나-없는-경우)
        - [주의사항](#주의사항)
        - [`$_POST`와의 차이점](#_post와의-차이점)
        - [`$HTTP_RAW_POST_DATA`와의 차이점](#http_raw_post_data와의-차이점)

## wrappers?

- PHP에는 `fopen()`, `copy()`, `file_exists()`, `filesize()`와 같은 파일시스템 함수와 함께 사용할 수 있는 다양한 URL 스타일 프로토콜을 위한 많은 기본 제공 래퍼가 있다.
- 이러한 래퍼 외에도 `stream_wrapper_register()` 함수를 사용하여 사용자 정의 래퍼를 등록할 수 있다

## 지원하는 프로토콜 리스트

- [`file://`](https://www.php.net/manual/en/wrappers.file.php) — Accessing local filesystem
- [`http://`](https://www.php.net/manual/en/wrappers.http.php) — Accessing HTTP(s) URLs
- [`ftp://`](https://www.php.net/manual/en/wrappers.ftp.php) — Accessing FTP(s) URLs
- [`php://`](https://www.php.net/manual/en/wrappers.php.php) — Accessing various I/O streams
- [`zlib://`](https://www.php.net/manual/en/wrappers.compression.php) — Compression Streams
- [`data://`](https://www.php.net/manual/en/wrappers.data.php) — Data (RFC 2397)
- [`glob://`](https://www.php.net/manual/en/wrappers.glob.php) — Find pathnames matching pattern
- [`phar://`](https://www.php.net/manual/en/wrappers.phar.php) — PHP Archive
- [`ssh2://`](https://www.php.net/manual/en/wrappers.ssh2.php) — Secure Shell 2
- [`rar://`](https://www.php.net/manual/en/wrappers.rar.php) — RAR
- [`ogg://`](https://www.php.net/manual/en/wrappers.audio.php) — Audio streams
- [`expect://`](https://www.php.net/manual/en/wrappers.expect.php) — Process Interaction Streams

## `php://`

- HTTP 요청 본문의 내용을 읽기 위한 일반적인 방법

### `php://input`이란?

- `php://input`은 PHP에서 제공하는 스트림 래퍼이며, 원시 HTTP 요청의 본문을 읽는 데 사용
- 이 스트림은 읽기 전용이며, POST나 PUT 요청 등을 통해 전송된 원시 데이터에 접근할 수 있다

### 동작 방식

1. HTTP 요청 수신: 클라이언트(브라우저, 모바일 앱 등)에서 서버로 HTTP 요청이 전송
2. 원시 데이터 접근: `file_get_contents('php://input')`은 이 요청의 본문에 들어 있는 원시 데이터에 접근한다. 이 데이터는 일반적으로 JSON, XML, 혹은 URL 인코딩된 문자열 등의 형태일 수 있다
3. 데이터 처리: 서버는 이 원시 데이터를 읽고, 필요에 따라 파싱하거나 처리

### 사용 방식

- 비표준 콘텐츠-타입 처리
    - `application/x-www-form-urlencoded`나 `multipart/form-data`와 같은 표준 콘텐츠 타입이 아닌 경우 (예: `application/json`), PHP가 자동으로 파싱하지 않는 데이터를 처리할 때 유용하다
    - XML이나 JSON과 같은 데이터 형식
    - SOAPXML-RPC나 요청을 처리할 때, `Content-type`이 `text/xml`로 설정되어 있기 때문에 `$_POST` 배열에는 데이터가 포함되지 않는다. 이런 경우 `file_get_contents('php://input')`를 사용하여 원시 XML 데이터에 접근할 수 있다.
- PUT, PATCH 요청 처리
    - POST가 아닌 HTTP 메소드 (예: PUT, PATCH)로 전송된 데이터를 처리할 때 자주 사용된다

### 값이 있거나 없는 경우

- 값이 있는 경우
    - HTTP 요청 본문에 데이터가 포함되어 있고, `php://input`을 통해 이 데이터에 접근할 수 있는 경우
- 값이 없는 경우: 몇 가지 이유가 있을 수 있습니다.
    - HTTP 요청 본문이 비어있는 경우.
    - `php://input`은 `multipart/form-data`나 `application/x-www-form-urlencoded`로 인코딩된 데이터에 대해서는 접근할 수 없다. 이러한 콘텐츠 타입의 경우 PHP가 자동으로 데이터를 파싱하고 `$_POST` 배열에 저장한다.
    - 요청 메소드가 GET이나 HEAD와 같이 본문을 포함하지 않는 경우

### 주의사항

- `file_get_contents('php://input')`으로 데이터를 한 번 읽으면, 그 데이터는 다시 읽을 수 없다. PHP 5.6 이후 버전에서는 이 데이터를 재사용할 수 있도록 변경되었다
- 이 스트림은 `enctype`이 `multipart/form-data`나 `application/x-www-form-urlencoded`인 POST 요청에는 사용할 수 없다
- 보안 측면에서, 원시 데이터를 처리할 때는 항상 주의가 필요하다. 입력 데이터 검증 및 적절한 처리가 필수적이다

### `$_POST`와의 차이점

- `$_POST` 배열은 `application/x-www-form-urlencoded`나 `multipart/form-data` 형태의 데이터를 파싱하여 저장한다
- 반면, `php://input`은 요청의 원시 데이터를 그대로 제공한다. 따라서 `Content-type`이 위 두 가지가 아닌 경우, `$_POST`에는 데이터가 없을 수 있다

### `$HTTP_RAW_POST_DATA`와의 차이점

- 과거에는 `$HTTP_RAW_POST_DATA` 변수를 사용하여 원시 POST 데이터에 접근할 수 있었으나, 이는 PHP 5.6 이후 버전에서 비권장되었다
- 현대 PHP 버전에서는 `php://input`이 권장되는 방식
