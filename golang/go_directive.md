# Go directive

- [Go directive](#go-directive)
    - [지시자(directive)](#지시자directive)
    - [Go에서의 지시자들](#go에서의-지시자들)

## 지시자(directive)

지시자는 주석의 특별한 형태입니다.
코드로 해석되지는 않지만, 정적 분석 도구에 의해 인식됩니다.

일반 주석은 개발자를 위한 설명이지만,
지시자는 도구나 컴파일러에게 특정 동작을 지시합니다.

## Go에서의 지시자들

- `//go:generate`: 코드 생성 도구 실행
- `//go:build`: 빌드 제약 조건 정의
- `//go:linkname`: 심볼 링킹 제어
- `//go:nosplit`: 스택 분할 비활성화
- `//lint:file-ignore`: 파일 전체에 대한 린트 규칙 무시

```go
//go:generate protoc --go_out=. myproto.proto

//go:build linux

//go:linkname runtime_procPin runtime.procPin
```
