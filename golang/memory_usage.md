# Memory usage

- [Memory usage](#memory-usage)
    - [메모리 사용 로그 출력](#메모리-사용-로그-출력)

## 메모리 사용 로그 출력

```go
func printMemUsage(msg string) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    mib := m.Alloc / 1024 / 1024
    fmt.Println(msg, mib, "MiB")
}
```
