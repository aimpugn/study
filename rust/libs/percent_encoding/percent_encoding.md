# [percent_encoding](https://crates.io/crates/percent-encoding)

- [percent\_encoding](#percent_encoding)
    - [`percent_encoding` crate?](#percent_encoding-crate)
    - [`percent_decode` function](#percent_decode-function)
        - [`percent_decode`와 `form_urlencoded::parse`의 차이](#percent_decode와-form_urlencodedparse의-차이)
    - [VS `form_urlencoded::parse`](#vs-form_urlencodedparse)

## `percent_encoding` crate?

URL은 요청의 부분을 나타내기 위해 특수 문자를 사용합니다.

예를 들어, `?`는 경로의 끝과 쿼리 스트링의 시작을 표시하는데,
`?`가 `va?lue`처럼 경로 내부에 존재하기 위해서는 다르게 인코딩되어야 합니다.([RFC_3986.md의 퍼센트 인코딩 참고](./../../../standards/RFC_3986.md))

```url
https://subdomain.domain.com/path/to/endpoint?key=va?lue
```

퍼센트 인코딩은 예약된 문자들을 `%` 이스케이프 문자와 뒤따르는 두 개의 16진수로 표현된 바이트 값으로 대체합니다.
- `%20` -> ASCII 공백(`' '`)으로 대체됩니다.

    URL에서 공백 문자(`' '`)는 예약된 문자이므로, 직접 사용될 수 없습니다.
    따라서 퍼센트 인코딩을 사용하여 `%20`으로 표현됩니다.

    예를 들어, "Hello World"라는 문자열이 URL의 일부로 사용되어야 할 경우, "Hello%20World"로 인코딩됩니다.

- `%3f` -> 물음표(`?`)로 변환됩니다.

## `percent_decode` function

퍼센트 인코딩된 바이트 시퀀스를 디코딩합니다.
이 함수는 주로 URL이나 기타 퍼센트 인코딩된 데이터를 원래의 형식으로 변환하는 데 사용됩니다.

`percent_decode` 함수의 주요 특징은 다음과 같습니다:
- ASCII 문자와 일부 안전한 문자('~'와 같은)는 인코딩되지 않습니다.
- 공백은 `+` 또는 `%20`으로 인코딩될 수 있으며, 둘 다 디코딩됩니다.
- UTF-8로 인코딩된 유니코드 문자를 올바르게 처리합니다.
- 잘못된 퍼센트 인코딩(예: `%2`와 같이 불완전한 인코딩)은 그대로 유지됩니다.
- `decode_utf8()` 메소드 경우 디코딩 과정에서 오류가 발생하면 (예: 유효하지 않은 UTF-8 시퀀스) 에러를 반환할 수 있습니다.

퍼센트 인코딩된 바이트 시퀀스 (`&[u8]`)를 입력으로 받을 경우:
- `%XX` 형식의 시퀀스를 해당하는 단일 바이트로 변환(예: `%20`은 공백(ASCII 32)으로 변환)합니다.
- 인코딩되지 않은 문자(ASCII 문자)는 그대로 유지합니다.

그 결과 디코딩된 바이트 시퀀스를 나타내는 `PercentDecode<'_>` 타입 (이는 `Iterator<Item = u8>`를 구현함)을 반환합니다.
- `Into<Cow<u8>>` 트레잇 구현: 디코딩이 필요 없는 경우 입력을 그대로 반환
- `Iterator<Item = u8>` 구현: 결과를 다양한 컬렉션 타입으로 변환 가능 (예: `Vec<u8>`, `String` 등)

이 함수는 다음과 같은 경우들에 사용됩니다.
- URL의 개별 구성 요소 (경로, 쿼리 파라미터 등) 디코딩
- 퍼센트 인코딩된 단일 문자열 디코딩하여 원본 형식으로 변환
- `application/x-www-form-urlencoded` 형식의 데이터 처리 (단, 키-값 쌍 분리는 별도로 처리해야 함)

하지만 UTF-8 유효성 검사를 수행하지 않음. 필요 시 `decode_utf8` 등의 별도 함수 사용 필요합니다.
또한 잘못된 퍼센트 인코딩 (예: `%2G`)은 그대로 유지되므로 주의해야 합니다.

### `percent_decode`와 `form_urlencoded::parse`의 차이

`percent_decode`와 `form_urlencoded::parse`는 모두 URL 인코딩된 데이터를 디코딩하는 데 사용되지만, 처리하는 데이터의 유형과 사용 사례에서 차이가 있습니다.

- `percent_decode`:
    - 단일 문자열 또는 바이트 시퀀스 디코딩
    - 순수 디코딩만 수행
    - URL 구성 요소, 단일 인코딩된 문자열 처리

- `form_urlencoded::parse`:
    - 전체 폼 데이터 (키-값 쌍) 파싱 및 디코딩
    - 디코딩 + 키-값 쌍 분리
    - HTML 폼 제출 데이터, 쿼리 문자열 전체 처리

```rust
use percent_encoding::percent_decode;

fn main() {
    // python3 -c "import urllib.parse; print(urllib.parse.quote('Hello, 世界! @#$%^&*'))"
    // Hello%2C%20%E4%B8%96%E7%95%8C%21%20%40%23%24%25%5E%26%2A
    let encoded = "Hello%2C%20%E4%B8%96%E7%95%8C%21%20%40%23%24%25%5E%26%2A";
    let decoded = percent_decode(encoded.as_bytes()).decode_utf8().unwrap();
    println!("Decoded: {}", decoded);
    // 출력: Decoded: Hello, 世界! @#$%^&*
}
```

1. URL 인코딩 전

    원본 문자열: `"Hello, 世界! @#$%^&*"`

2. URL 인코딩 후

    인코딩된 문자열: `""`

    ```sh
          ',' ' '    '世'      '界'    '!' ' ' '@' '#' '$' '%' '^' '&' '*'
    Hello|%2C|%20|%E4%B8%96|%E7%95%8C|%21|%20|%40|%23|%24|%25|%5E|%26|%2A
    ```

    인코딩 과정 설명:
    - `Hello`: ASCII 알파벳은 인코딩되지 않습니다.
    - `,` (쉼표): `%2C`로 인코딩됩니다.
    - 공백: `%20`으로 인코딩됩니다.
    - `世界` (한자): UTF-8로 인코딩된 후 각 바이트가 퍼센트 인코딩됩니다.
        - '世': UTF-8에서 `E4 B8 96` -> `%E4%B8%96`
        - '界': UTF-8에서 `E7 95 8C` -> `%E7%95%8C`
    - `!`: `%21`로 인코딩됩니다.
    - `@`: `%40`으로 인코딩됩니다.
    - `#`: `%23`으로 인코딩됩니다.
    - `$`: `%24`로 인코딩됩니다.
    - `%`: `%25`로 인코딩됩니다.
    - `^`: `%5E`로 인코딩됩니다.
    - `&`: `%26`으로 인코딩됩니다.
    - `*`: `%2A`로 인코딩됩니다.

3. percent_decode를 통한 디코딩

    `percent_decode` 함수의 디코딩 과정을 살펴보겠습니다:

    ```rust
    use percent_encoding::percent_decode;

    fn main() {
        let encoded = "Hello%2C%20%E4%B8%96%E7%95%8C%21%20%40%23%24%25%5E%26%2A";
        let decoded = percent_decode(encoded.as_bytes()).decode_utf8().unwrap();
        println!("Decoded: {}", decoded);
    }
    ```

    1. `percent_decode` 함수는 입력된 바이트 시퀀스를 순회합니다.
    2. `%` 문자를 만나면, 그 다음 두 문자를 16진수로 해석하여 해당하는 바이트로 변환합니다:
        - `%2C` -> `,` (ASCII 44)
        - `%20` -> `` (공백, ASCII 32)
        - `%E4%B8%96` -> `世`의 UTF-8 바이트 시퀀스
        - `%E7%95%8C` -> `界`의 UTF-8 바이트 시퀀스
        - `%21` -> `!` (ASCII 33)
        - `%40` -> `@` (ASCII 64)
        - `%23` -> `#` (ASCII 35)
        - `%24` -> `$` (ASCII 36)
        - `%25` -> `%` (ASCII 37)
        - `%5E` -> `^` (ASCII 94)
        - `%26` -> `&` (ASCII 38)
        - `%2A` -> `*` (ASCII 42)
    3. `%`로 시작하지 않는 문자는 그대로 유지됩니다 (예: `Hello`의 각 문자).
    4. 디코딩된 바이트 시퀀스는 `decode_utf8()` 메소드를 통해 유효한 UTF-8 문자열로 해석됩니다.
    5. 최종 결과로 원본 문자열 `"Hello, 世界! @#$%^&*"`가 복원됩니다.

## VS `form_urlencoded::parse`

`form_urlencoded::parse` 함수는 특히 `application/x-www-form-urlencoded` 미디어 타입의 데이터를 파싱하고 디코딩하는 데 사용됩니다.
이 함수는 입력된 바이트 스트림을 키-값 쌍으로 분리하고, 각 쌍의 키와 값 모두에 대해 퍼센트 디코딩을 수행합니다.
그 결과 키-값 쌍의 이터레이터가 반환됩니다.

주로 HTML 폼 데이터 또는 HTTP GET 요청의 쿼리 스트링을 파싱하고 디코딩하는 데 사용됩니다.
이 형식은 키-값 쌍을 `&`로 연결하고, 각 키와 값은 `=`로 연결된 문자열입니다.

`url` 크레이트의 `form_urlencoded` 모듈에서 제공됩니다.
