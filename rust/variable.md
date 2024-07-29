# Variable

## trouble shooting

### creates a temporary which is freed while still in use

이 문의에 대한 [답변](https://users.rust-lang.org/t/creates-a-temporary-which-is-freed-while-still-in-use-again/29211/2)에 의하면 expression이 변수에 할당되는 것과 다른 expression의 일부로 사용되는 것은 서로 다른 것이라 한다.

```rs
let process = Command::new(location_test).arg(address);

// 아래 코드는 위의 코드와 다르다.
// `Command`를 own할 수 있다
let process = Command::new(location_test);
process.arg(address);
```
