# hex

- [hex](#hex)
    - [`hex.DecodeString`](#hexdecodestring)
        - [`[]byte(s)`와의 비교](#bytes와의-비교)

## `hex.DecodeString`

이 함수는 16진수 문자열을 바이너리 데이터로 디코딩합니다.
즉, 2자리 16진수 문자열(16진수 쌍)을 하나의 바이트로 변환합니다.

예를 들어, "d5bc55fd"는 `0xd5`, `0xbc`, `0x55`, `0xfd`와 같은 네 개의 바이트로 변환됩니다.

### `[]byte(s)`와의 비교

문자열을 `[]byte`로 변환하면, 각 문자의 ASCII 코드 값을 바이트 배열로 변환합니다.

예를 들어, 문자열 "d5"는 ASCII 코드 값인 `0x64`(100)와 `0x35`(53)로 변환됩니다.

- `hex.DecodeString`은 16진수 문자열을 실제 바이트 데이터로 변환합니다.
- `[]byte` 캐스팅은 문자열의 각 문자를 ASCII 값으로 변환합니다.

```go
package main

import (
    "encoding/hex"
    "fmt"
    "strings"
)

func main() {
    // 1. UUID 문자열을 정의합니다.
    uuidWithHyphens := "d5bc55fd-367b-4490-aed7-2b2e72be4c85"
    fmt.Println("UUID:", uuidWithHyphens)

    // 2. UUID 문자열에서 하이픈을 제거합니다.
    uuidWithHyphens = strings.ReplaceAll(uuidWithHyphens, "-", "")
    fmt.Println("UUID without hyphens:", uuidWithHyphens)

    // 3. 16진수 문자열을 바이너리 데이터로 변환합니다.
    binaryData, err := hex.DecodeString(uuidWithHyphens)
    if err != nil {
        fmt.Println("Error decoding hex string:", err)
        return
    }
    fmt.Println("Binary data:", binaryData)

    // 문자열을 바이트 배열로 캐스팅한 결과를 출력합니다.
    fmt.Println("Binary data by casting to []byte:", []byte(uuidWithHyphens))
}
```

```bash
Binary data: [213 188 85 253 54 123 68 144 174 215 43 46 114 190 76 133]
Binary data by casting to []byte: [100 53 98 99 53 53 102 100 51 54 55 98 52 52 57 48 97 101 100 55 50 98 50 101 55 50 98 101 52 99 56 53]
```

**Binary data**:

- 여기서 `[213 188 85 253 54 123 68 144 174 215 43 46 114 190 76 133]`는 `d5bc55fd367b4490aed72b2e72be4c85`를 16진수로 디코딩한 결과입니다.
- `d5`는 213, `bc`는 188, `55`는 85, `fd`는 253로 변환됩니다. 이와 같은 방식으로 모든 16진수 쌍이 변환됩니다.

**Binary data by casting to []byte**:

- 여기서 `[100 53 98 99 53 53 102 100 51 54 55 98 52 52 57 48 97 101 100 55 50 98 50 101 55 50 98 101 52 99 56 53]`는 각 문자의 ASCII 값을 나타냅니다.
- 예를 들어, `d`는 100, `5`는 53, `b`는 98, `c`는 99 등으로 변환됩니다.
