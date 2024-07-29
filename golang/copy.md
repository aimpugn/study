# copy

## `map[string]interface{}`

- `map`은 참조 타입 (Reference Type)
- 변수에 할당될 때 메모리 주소가 복사된다
- [Copying all elements of a map into another](https://stackoverflow.com/questions/7436864/copying-all-elements-of-a-map-into-another)

```go
dst := make(map[K]V, len(src))

for k, v := range src {
    dst[k] = v
}
```

```go
// Starting with Go 1.18, we can create a generic solution:
func MapCopy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
    for k, v := range src {
        dst[k] = v
    }
}
```

```go
// Pre 1.18 "generic" solution follows:
// If performance is not an issue (e.g. you're working with small maps), 
// a general solution may be created using the reflect package:
func MapCopy(dst, src interface{}) {
    dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

    for _, k := range sv.MapKeys() {
        dv.SetMapIndex(k, sv.MapIndex(k))
    }
}
```
