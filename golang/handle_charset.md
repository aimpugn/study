# Handle charset

- [Handle charset](#handle-charset)
    - [`EUC-KR` -\> `UTF-8`](#euc-kr---utf-8)
    - [테스트](#테스트)

## `EUC-KR` -> `UTF-8`

Go에서 `EUC-KR` 인코딩된 문자열이 URL 인코딩 되었는지 확인하려면, 먼저 URL 디코딩을 수행한 후 `EUC-KR`에서 `UTF-8`로의 변환을 시도해볼 수 있다.

Go의 표준 라이브러리에는 직접적으로 `EUC-KR`을 지원하지 않기 때문에, `golang.org/x/text/encoding/korean` 패키지를 사용하여 `EUC-KR`을 처리할 수 있다.

> NOTE: Go의 표준 라이브러리에 포함되어 있지 않은 `golang.org/x/text` 패키지가 필요하다.
>
> ```sh
> go get golang.org/x/text
> ```

```go
package main

import (
    "fmt"
    "net/url"

    "golang.org/x/net/html/charset"
    "golang.org/x/text/encoding/korean"
    "golang.org/x/text/transform"
    "strings"
    "io"
)

func main() {
    // URL 인코딩된 `EUC-KR` 문자열
    encodedStr := "%C8%AB%B1%E6%B5%BF"

    // URL 디코딩
    decodedStr, err := url.QueryUnescape(encodedStr)
    if err != nil {
        fmt.Println("Error decoding URL:", err)
        return
    }

    // `EUC-KR`에서 `UTF-8`로 변환
    utf8Str, err := eucKrToUtf8(decodedStr)
    if err != nil {
        fmt.Println("Error converting from EUC-KR to UTF-8:", err)
        return
    }

    fmt.Println("Original EUC-KR & URL encoded string:", encodedStr)
    fmt.Println("Decoded UTF-8 string:", utf8Str)
}

func eucKrToUtf8(str string) (string, error) {
    // EUC-KR 디코더를 사용하여 UTF-8로 변환
    reader := transform.NewReader(strings.NewReader(str), korean.EUCKR.NewDecoder())
    utf8Bytes, err := io.ReadAll(reader)
    if err != nil {
        return "", err
    }
    return string(utf8Bytes), nil
}
```

## 테스트

```go
import (
    "fmt"
    "io"
    "net/url"
    "strings"
    "testing"

    "github.com/portone-io/gophplib/v2"
    "github.com/stretchr/testify/require"
    "golang.org/x/text/encoding/korean"
    "golang.org/x/text/transform"
)

func TestDecodeString(t *testing.T) {
    // URL-encoded EUC-KR string
    encodedStr := "name=%C8%AB%B1%E6%B5%BF&age=30&gender=male"
    decodedStr1, err := url.QueryUnescape("%C8%AB%B1%E6%B5%BF")
    fmt.Println("decodedStr", decodedStr1)
    tmp1 := transform.NewReader(strings.NewReader(decodedStr1), korean.EUCKR.NewDecoder())
    result, err := io.ReadAll(tmp1)
    fmt.Printf("0. %s\n", result)

    tmp := gophplib.ParseStr(encodedStr)
    fmt.Println(tmp)
    nameInAny, ok := tmp.Get("name")
    require.True(t, ok)
    name := nameInAny.(string)
    fmt.Println("name:", name)
    decodedStr2, err := url.QueryUnescape(name)

    tmp2 := transform.NewReader(strings.NewReader(decodedStr2), korean.EUCKR.NewDecoder())
    result, err = io.ReadAll(tmp2)
    fmt.Printf("1. %s\n", result)
    fmt.Println("1.", decodedStr1 == decodedStr2)

    // Decode the URL-encoded string
    decodedStr3, err := url.QueryUnescape(encodedStr)
    if err != nil {
        fmt.Printf("Error decoding the string: %v", err)
        return
    }

    // Convert EUC-KR to UTF-8
    r := strings.NewReader(decodedStr3)
    tr := transform.NewReader(r, korean.EUCKR.NewDecoder())
    result, err = io.ReadAll(tr)
    if err != nil {
        fmt.Printf("Error converting EUC-KR to UTF-8: %v", err)
        return
    }

    fmt.Printf("Decoded and converted string: %s\n", result)
}
```

```bash
decodedStr ȫ�浿
0. 홍길동
{map[age:0x1400010f620 gender:0x1400010f650 name:0x1400010f5f0] {{0x1400010f5f0 0x1400010f650 <nil> <nil>}}}
name: ȫ�浿
1. 홍占썸동
1. false
Decoded and converted string: name=홍길동&age=30&gender=male
```
