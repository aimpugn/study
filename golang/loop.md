# loop

- [loop](#loop)
    - [Go 1.22 버전 전 반복문](#go-122-버전-전-반복문)
    - [Go 1.22 버전 후 반복문](#go-122-버전-후-반복문)

## Go 1.22 버전 전 반복문

Go 1.22 버전 전에서는 `info`는 한 메모리 공간을 계속 재사용했습니다.
따라서 반복문 안에서 `info`를 그대로 사용하면 결과값이 오염될 수 있었습니다.

```go
func main() {
    done := make(chan bool)

    values := []string{"a", "b", "c"}
    for _, v := range values {
        go func() {
            // they usually print “c”, “c”, “c”, 
            // instead of printing “a”, “b”, and “c” in some order.
            fmt.Println(v)
            done <- true
        }()
    }

    // wait for all goroutines to complete before exiting
    for _ = range values {
        <-done
    }
}
```

출력되는 결과:

```sh
❯ go version
go version go1.20.3 darwin/arm64

❯ go test -run TestLoop
c
c
c
PASS
ok      github.com/aimpugn/snippets/golang/examples/builtin     0.180s
```

이를 해결하기 위해 기존에는 변수에 명시적으로 할당하여 새로운 메모리 공간을 확보하는 식으로 사용해야 했습니다.

```go
for _, info := range values[SHORTENED] {
    var someCode, someCodeName string
    var stdCode *string

    if info.StdCode != nil {
        someCode = *info.StdCode
    }
    fmt.Println(info.someCode)
    someCodeName = info.someCodeName
    stdCode = &info.someCode

    flip = append(flip, SomeJSON{
        someCode:     someCode,
        someCodeName: someCodeName,
        StdCode:    stdCode,
    })
}
```

## Go 1.22 버전 후 반복문

> 1.21 버전 경우에는 `GOEXPERIMENT=loopvar` 통해 사용 가능

```go
func main() {
    done := make(chan bool)

    values := []string{"a", "b", "c"}
    for _, v := range values {
        go func() {
            // they usually print “c”, “c”, “c”, 
            // instead of printing “a”, “b”, and “c” in some order.
            fmt.Println(v)
            done <- true
        }()
    }

    // wait for all goroutines to complete before exiting
    for _ = range values {
        <-done
    }
}
```

코드를 수정한 게 없지만, 1.22 버전 전과 다르게 쉐도잉하지 않아도 의도한대로 작동합니다.

```sh
❯ asdf shell golang 1.22.4     
❯ export GOROOT="$(asdf where golang 1.22.4)/go"
❯ go version
go version go1.22.4 darwin/arm64

❯ go test -run TestLoop                         
a
c
b
PASS
ok      github.com/aimpugn/snippets/golang/examples/builtin     0.341s
```
