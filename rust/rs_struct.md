# Struct

- [Struct](#struct)

## [Using Structs to structure related data](https://doc.rust-lang.org/book/ch05-00-structs.html)

> A `struct`, or `structure`, is a custom data type that lets you package together and name multiple related values that make up a meaningful group.

### 함께 사용되는 구조체의 가시성(visibility)

당연한 얘기겠지만, 외부에서 사용되는 struct에서 사용중인 다른 struct 있는 경우, 그 다른 struct의 가시성도 `pub`이어야 합니다.

```rs
pub struct Scheduler {
    pub proc_info: HashMap<i32, ProcInfo>,
                                ^^^^^^^^ `ProcInfo`이 외부에 노출된다
    ... 생략 ...
}

pub struct ProcInfo {
^^^ `Scheduler` 통해서 외부에 노출되므로, pub여야 한다
    ... 생략 ...
}
```
