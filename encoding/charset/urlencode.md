# URL encode

- [URL encode](#url-encode)
    - [UTF-8 → EUC-KR → URL encode](#utf-8--euc-kr--url-encode)
        - [1. UTF-8에서 EUC-KR로 인코딩 변환](#1-utf-8에서-euc-kr로-인코딩-변환)
        - [2. URL 인코딩](#2-url-인코딩)
        - [변환 과정 요약](#변환-과정-요약)

## UTF-8 → EUC-KR → URL encode

UTF-8에서 "홍길동" 문자열을 EUC-KR 문자셋으로 변환한 후 URL 인코딩하는 과정은 다음 단계를 포함합니다.
이 과정은 문자열의 인코딩 변환과 이후의 URL 인코딩으로 나눌 수 있습니다.

### 1. UTF-8에서 EUC-KR로 인코딩 변환

"홍길동"이라는 문자열은 UTF-8에서 특정 바이트 시퀀스로 표현됩니다. EUC-KR로 변환하려면, 이 UTF-8 바이트 시퀀스를 EUC-KR에서 해당하는 문자의 바이트 시퀀스로 매핑해야 합니다.

예를 들어, UTF-8에서 "홍길동"은 다음과 같은 바이트 시퀀스를 가질 수 있습니다:
- 홍: `\xed\x99\x8d`
- 길: `\xea\xb8\xb8`
- 동: `\xeb\x8f\x99`

이 바이트 시퀀스를 EUC-KR로 변환하면, 각 문자가 EUC-KR에서 정의된 해당하는 바이트로 매핑됩니다. EUC-KR은 한글을 2바이트로 표현합니다. 변환 후의 바이트 시퀀스는 예시와 같이 다를 수 있습니다 (실제 바이트 값은 문자에 따라 다름):
- 홍: `\xc8\xab`
- 길: `\xb1\xe6`
- 동: `\xb5\xee`

### 2. URL 인코딩

EUC-KR로 인코딩된 바이트 시퀀스를 URL 인코딩하려면, 각 바이트를 `%` 다음에 2자리 16진수로 표현합니다. URL 인코딩은 URL에서 사용할 수 없는 문자나 바이트를 표현하는 데 사용됩니다.

따라서, 위에서 EUC-KR로 변환된 "홍길동"의 바이트 시퀀스를 URL 인코딩하면 다음과 같이 변환될 수 있습니다:
- 홍: `%C8%AB`
- 길: `%B1%E6`
- 동: `%B5%BF`

결과적으로, "홍길동"은 URL 인코딩된 EUC-KR 문자열 `%C8%AB%B1%E6%B5%EE`로 표현됩니다.

### 변환 과정 요약

1. **인코딩 변환**: UTF-8로 인코딩된 "홍길동"을 EUC-KR로 변환합니다.
2. **URL 인코딩**: 변환된 EUC-KR 바이트 시퀀스를 URL 인코딩합니다.

이 과정은 특정 프로그래밍 언어나 라이브러리의 함수를 사용하여 프로그래밍적으로 수행할 수 있습니다.
Rust에서는 `encoding_rs` 크레이트와 같은 라이브러리를 사용하여 인코딩 변환을 수행할 수 있으며, URL 인코딩은 표준 라이브러리의 URL 인코딩 함수나 `percent-encoding` 크레이트를 사용할 수 있습니다.