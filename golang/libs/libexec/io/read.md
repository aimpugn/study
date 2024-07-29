# Read

- [Read](#read)
    - [`ReadAll`](#readall)
        - [메모리 사용](#메모리-사용)
    - [`ReadFull`](#readfull)
        - [`EOF`와 `ErrUnexpectedEOF`의 차이?](#eof와-errunexpectedeof의-차이)
        - [만약 `len(buf)`보다 적은 수의 바이트를 읽게 되면?](#만약-lenbuf보다-적은-수의-바이트를-읽게-되면)
        - [`len(buf)` 바이트 이상 읽었지만 그 과정에서 오류가 있었다면?](#lenbuf-바이트-이상-읽었지만-그-과정에서-오류가-있었다면)
    - [예제](#예제)
    - [LimitReader](#limitreader)
        - [메모리 사용](#메모리-사용-1)

## `ReadAll`

- `ReadAll` reads from `r` until an error or `EOF` and returns the data it read.
- A successful call returns `err == nil`, not `err == EOF`.
- Because `ReadAll` is defined to read from src until `EOF`, it does not treat an `EOF` from `Read` as an error to be reported.

```go
r := strings.NewReader("Go is a general-purpose language designed with systems programming in mind.")  
b, err := io.ReadAll(r) 
if err != nil {     
    log.Fatal(err) 
}  
fmt.Printf("%s", b)
```

### 메모리 사용

```log
before Do Alloc 8 MiB
after Do Alloc 8 MiB
after ReadAll 67 MiB
ReadAll err is <nil>
ReadAll err is io.ErrUnexpectedEOF false
ReadAll err is io.EOF false
ReadAll length is 58487996
```

## `ReadFull`

- `ReadFull` reads exactly `len(buf)` bytes from `r` into `buf`.
- It returns the number of bytes copied and an error **if fewer bytes were read**.
- The error is EOF only if no bytes were read.
- If an EOF happens after reading some but not all the bytes, ReadFull returns `ErrUnexpectedEOF`.
- On return, `n == len(buf)` **if and only if** `err == nil`. ([동치](https://article2.tistory.com/1346))
- If `r` returns an error having read at least `len(buf)` bytes, the error is dropped.

### `EOF`와 `ErrUnexpectedEOF`의 차이?

- `EOF` (End Of File) 에러는 파일이나 데이터 스트림의 끝에 도달했음을 의미. 끝에 도달했다는 의미라서 에러라기보다는 정상으로 간주되기도 한다.
- `ErrUnexpectedEOF`는 예상보다 일찍 데이터 스트림의 끝에 도달했음을 의미. 즉, 더 많은 데이터를 기대했지만 더 이상 읽을 데이터가 없는 상황에서 발생

### 만약 `len(buf)`보다 적은 수의 바이트를 읽게 되면?

1. 아무 바이트도 읽지 않았다면 `EOF` 에러 반환
2. 일부 바이트는 읽었지만 len(buf)보다 적게 읽은 경우에는 `ErrUnexpectedEOF` 에러 반환

### `len(buf)` 바이트 이상 읽었지만 그 과정에서 오류가 있었다면?

- `r`이 `len(buf)` 바이트 이상을 읽은 상태에서 오류를 반환하면, 그 오류는 무시된다
- 즉, 원하는 바이트 수를 성공적으로 읽었다면 그 이후에 발생하는 오류는 무시

## 예제

```go
r := strings.NewReader("some io.Reader stream to be read\n")  
buf := make([]byte, 4) 
if _, err := io.ReadFull(r, buf); err != nil {     
    log.Fatal(err) 
} 
fmt.Printf("%s\n", buf)  

// minimal read size bigger than io.Reader stream 
longBuf := make([]byte, 64) 
if _, err := io.ReadFull(r, longBuf); err != nil {     
    fmt.Println("error:", err) 
}
```

## LimitReader

```go
reader := io.LimitReader(resp.Body, maxBodyLength)
read, err := io.ReadAll(reader)
// read, err := io.ReadAll(resp.Body)
if err != nil {
    return nil, err
}
```

### 메모리 사용

```log
before Do Alloc 7 MiB
after Do Alloc 8 MiB
after LimitReader 8 MiB
after LimitReader ReadAll 8 MiB
LimitReader err is <nil>
LimitReader err is io.ErrUnexpectedEOF false
LimitReader err is io.EOF false
LimitReader length is 1024
```
