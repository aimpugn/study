# Lint

## [elided-lifetimes-in-paths](https://doc.rust-lang.org/rustc/lints/listing/allowed-by-default.html#elided-lifetimes-in-paths)

> Elided lifetime parameters can make it difficult to see at a glance that borrowing is occurring. This lint ensures that lifetime parameters are always explicitly stated, even if it is the ['_ placeholder](https://doc.rust-lang.org/reference/lifetime-elision.html#lifetime-elision-in-functions) lifetime.
>
> This lint is "allow" by default because it has some known issues, and may require a significant transition for old code.

borrowing이 발생하는 것을 파악하기 어렵게 만든다
이 린트는 파라미터의 lifetime이 명시적으로 선언되는 것을 보장한다

- [What does "missing lifetime specifier" mean when storing a &str in a structure?](https://stackoverflow.com/a/33485351)
- [Why is Vec<&str> missing lifetime specifier here?](https://stackoverflow.com/a/57267192)

```rs
// 컴파일러에게 문자열이 최소한 `StrDisplayable` struct가 존재하는 동안 지속된다고 명시적으로 알린다
struct StrDisplayable<'a>(Vec<&'a str>);
```

```rs
// 이 경우 정의되지 않은 lifetime을 참조하게 된다
struct StrDisplayable(Vec<&'a str>); // Error!
```

[Error code E0261](https://doc.rust-lang.org/error_codes/E0261.html)

```rs
impl Scheduler<'a> {
              ^^^^ 
              error[E0261]: use of undeclared lifetime name `'a`
}
```

```rs
impl<'a> Scheduler<'a> {
    ^^^^ 
    correct
    Impl blocks declare lifetime parameters separately
}
```
