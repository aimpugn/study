# WriteFile

- [WriteFile](#writefile)
    - [파일 쓰기](#파일-쓰기)

## 파일 쓰기

```go
// `fs.ModeAppend` 등 모드 가능.
// 아니면 직접 0644 명시
var someBytes []byte
someBytes = resp.Hits.Hits // ES response 중 Hits 속성(`json.RawMessage`)
temp := os.WriteFile("result.json", someBytes, 0644)
if temp != nil {
    fmt.Println("WriteFile err", temp)
}
```
