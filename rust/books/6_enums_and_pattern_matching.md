# 6. Enums and Pattern MAtching

## [6.3. Concise Control Flow with if let](https://doc.rust-lang.org/book/ch06-03-if-let.html#concise-control-flow-with-if-let)

`if let` 구문을 사용하면 `if`와 `let`을 결합하여 하나의 패턴과 일치하는 값을 처리하고 나머지는 무시하는 덜 장황한 방식으로 처리할 수 있다.
목록 6-6의 프로그램이 `config_max` 변수의 `Option<u8>` 값과 일치하지만 값이 `Some`인 경우에만 코드를 실행하려고 한다고 가정해 보자.

값이 `Some`인 경우 패턴의 `max` 변수에 값을 바인딩하여 `Some` variant의 값을 출력한다.
`None` 값으로는 아무 작업도 수행하지 않으려 합니다. 일치 표현식을 충족하려면 하나의 변형만 처리한 후 _ => ()를 추가해야 하는데, 이는 추가하기 번거로운 상용구 코드입니다.

```rs
let config_max = Some(3u8);
match config_max {
    Some(max) => println!("The maximum is configured to be {}", max),
    //   ^^^ 3u8 값을 바인딩
    _ => (),
}
```

`if let`을 사용해서 더 간단하게 작성할 수 있다.

```rs
let config_max = Some(3u8);
if let Some(max) = config_max {
    println!("The maximum is configured to be {}", max);
}
```
