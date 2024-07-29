# rust concurrency

## [`Send`와 `Sync` trait](https://doc.rust-lang.org/book/ch16-04-extensible-concurrency-sync-and-send.html)

### `Send` 트레이트

`Send` 트레이트는 타입이 스레드 간에 안전하게 전송될 수 있음을 나타냅니다.

```rust
pub unsafe auto trait Send { }
```

- 대부분의 타입은 자동으로 `Send`를 구현합니다.
- 포인터나 참조를 포함하는 타입은 주의가 필요합니다.
- `Rc<T>`와 같은 일부 타입은 의도적으로 `Send`를 구현하지 않습니다.

```rust
// Send를 구현하는 타입
struct SafeToSend;

// 다른 스레드로 안전하게 전송 가능
fn send_to_thread(value: SafeToSend) {
    std::thread::spawn(move || {
        // value 사용
    });
}
```

### `Sync` 트레이트

`Sync` 트레이트는 타입이 여러 스레드에서 동시에 안전하게 참조될 수 있음을 나타냅니다.

```rust
pub unsafe auto trait Sync { }
```

- `T: Sync`이면 `&T: Send`입니다.
- 내부 가변성을 가진 타입(예: `RefCell<T>`)은 일반적으로 `Sync`가 아닙니다.
- `Mutex<T>`와 같은 동기화 프리미티브는 `Sync`를 구현합니다.

```rust
use std::sync::Arc;

struct ThreadSafeData;

// Sync를 구현하는 타입은 Arc를 통해 여러 스레드에서 안전하게 공유 가능
fn share_across_threads(data: Arc<ThreadSafeData>) {
    for _ in 0..10 {
        let data_clone = Arc::clone(&data);
        std::thread::spawn(move || {
            // data_clone 사용
        });
    }
}
```

### Send와 Sync의 중요성

1. 컴파일 시간 안전성 보장:
   - 데이터 레이스와 같은 동시성 버그를 방지합니다.
   - 잘못된 스레드 간 데이터 공유를 컴파일 시에 감지합니다.

2. 추상화 수준 향상:
   - 라이브러리 작성자가 스레드 안전한 API를 설계할 수 있게 합니다.
   - 사용자는 구체적인 동시성 메커니즘을 알 필요 없이 안전하게 코드를 작성할 수 있습니다.

3. 성능 최적화:
   - 런타임 체크 없이 동시성 안전성을 보장하므로, 성능 오버헤드가 없습니다.

4. 명시적인 안전성 표현:
   - 타입 시스템을 통해 동시성 안전성을 명확히 표현합니다.

```rust
use std::sync::Mutex;

// T가 Send이면 Mutex<T>는 Send와 Sync를 모두 구현
struct ThreadSafeCounter<T: Send> {
    count: Mutex<T>,
}

impl<T: Send> ThreadSafeCounter<T> {
    fn new(initial: T) -> Self {
        ThreadSafeCounter {
            count: Mutex::new(initial),
        }
    }

    fn increment(&self, value: T) where T: std::ops::AddAssign {
        let mut count = self.count.lock().unwrap();
        *count += value;
    }
}

// 사용 예
let counter = ThreadSafeCounter::new(0);
let counter_arc = std::sync::Arc::new(counter);

for _ in 0..10 {
    let counter_clone = Arc::clone(&counter_arc);
    std::thread::spawn(move || {
        counter_clone.increment(1);
    });
}
```

이 예시에서 `ThreadSafeCounter<T>`는 `T: Send`일 때 자동으로 `Send`와 `Sync`를 구현합니다. 이를 통해 여러 스레드에서 안전하게 공유하고 수정할 수 있습니다.

## Troubleshooting

### `dyn std::error::Error` cannot be sent between threads safely

`Send` 트레이트 관련 문제로, 에러가 스레드 간에 전송될 수 있는지에 대한 것입니다.
`Send`는 소유권 이전에 관한 것입니다.

```rs
fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let encoding = EncodingType::from_str(&args.encoding)?;

    let input_file = File::open(&args.input)?;
    let reader = BufReader::new(input_file);
    let mut output_file = File::create(&args.output)?;

    reader
        .lines()
        .par_bridge()
        .try_for_each(|line| -> Result<(), Box<dyn Error + Send + Sync>> {
            let line = line?;
            let decoded = decode_input(&line, encoding)?;
                                                      ^^^
            let json = parse_to_json(&decoded)?;
                                             ^^^
            let json_string = serde_json::to_string(&json)?;

            // 각 스레드에서 결과를 문자열로 반환
            Ok(json_string)
        })
        .and_then(|results| {
            // 모든 결과를 하나의 스레드에서 파일에 쓰기
            for json_line in results {
                writeln!(output_file, "{}", json_line)?;
            }
            Ok(())
        })?;

    Ok(())
}
```

```sh
`dyn std::error::Error` cannot be sent between threads safely
the trait `Send` is not implemented for `dyn std::error::Error`, which is required by `Result<(), Box<dyn std::error::Error + Send + Sync>>: FromResidual<Result<Infallible, Box<dyn std::error::Error>>>`
the following other types implement trait `FromResidual<R>`:
  <Result<T, F> as FromResidual<Yeet<E>>>
  <Result<T, F> as FromResidual<Result<Infallible, E>>>
required for `Unique<dyn std::error::Error>` to implement `Send`
required for `Box<dyn std::error::Error + Send + Sync>` to implement `From<Box<dyn std::error::Error>>`
required for `Result<(), Box<dyn std::error::Error + Send + Sync>>` to implement `FromResidual<Result<Infallible, Box<dyn std::error::Error>>>`
```

### `dyn std::error::Error` cannot be shared between threads safely

`Sync` 트레이트 관련 문제로, 에러가 여러 스레드에서 동시에 안전하게 참조될 수 있는지에 대한 것입니다.
`Sync`는 공유 참조에 관한 것입니다.

```rs
type ThreadSafeError = Arc<dyn Error + Send + Sync>;

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let encoding = EncodingType::from_str(&args.encoding)?;

    let input_file = File::open(&args.input)?;
    let reader = BufReader::new(input_file);
    let mut output_file = File::create(&args.output)?;

    reader
        .lines()
        .par_bridge()
        .try_for_each(|line| -> Result<(), ThreadSafeError> {
            let line = line.map_err(|e| Arc::new(e) as ThreadSafeError)?;
            let decoded =
                decode_input(&line, encoding).map_err(|e| Arc::new(e) as ThreadSafeError)?;
            let json = parse_to_json(&decoded).map_err(|e| Arc::new(e) as ThreadSafeError)?;
            let json_string =
                serde_json::to_string(&json).map_err(|e| Arc::new(e) as ThreadSafeError)?;

            // 각 스레드에서 결과를 문자열로 반환
            Ok(())
        })?;
    Ok(())
}
```

`Arc<T>`를 사용할 때 `T`는 반드시 `Sync`여야 하므로, 이 문제가 발생했습니다.

```sh
`dyn std::error::Error` cannot be shared between threads safely
the trait `Sync` is not implemented for `dyn std::error::Error`, which is required by `Box<dyn std::error::Error>: Sync`
required for `Unique<dyn std::error::Error>` to implement `Sync`
required for the cast from `Arc<Box<dyn std::error::Error>>` to `Arc<(dyn std::error::Error + Send + Sync + 'static)>`rustcClick for full compiler diagnostic
boxed.rs(195, 12): required because it appears within the type `Box<dyn std::error::Error>`

`dyn std::error::Error` cannot be sent between threads safely
the trait `Send` is not implemented for `dyn std::error::Error`, which is required by `Box<dyn std::error::Error>: Send`
required for `Unique<dyn std::error::Error>` to implement `Send`
required for the cast from `Arc<Box<dyn std::error::Error>>` to `Arc<(dyn std::error::Error + Send + Sync + 'static)>`rustcClick for full compiler diagnostic
boxed.rs(195, 12): required because it appears within the type `Box<dyn std::error::Error>`

the size for values of type `dyn std::error::Error` cannot be known at compilation time
the trait `Sized` is not implemented for `dyn std::error::Error`, which is required by `Box<dyn std::error::Error>: std::error::Error`
the trait `std::error::Error` is implemented for `Box<T>`
required for `Box<dyn std::error::Error>` to implement `std::error::Error`
required for the cast from `Arc<Box<dyn std::error::Error>>` to `Arc<(dyn std::error::Error + Send + Sync + 'static)>`rustcClick for full compiler diagnostic
```

`dyn std::error::Error`가 `Sync` 트레이트를 구현하지 않았기에 발생한 에러입니다.

그렇다면 왜 이 문제가 발생했는가?
- `std::error::Error` 트레이트는 `Sync`를 요구하지 않습니다.
- `Arc<T>`는 `T: Sync`일 때만 `Sync`를 구현합니다.
- `ThreadSafeError`는 `Send + Sync`를 요구하지만, `Box<dyn Error>`는 이를 보장하지 않습니다.

이 문제를 해결하려면 다음 방법들을 고려합니다.

1. 구체적인 에러 타입 사용:

     가능하다면, `dyn Error` 대신 구체적인 에러 타입을 사용합니다.

    ```rust
    use std::io;

    fn decode_input(line: &str, encoding: Encoding) -> Result<String, io::Error> {
        // 구현...
    }

    // 사용
    decode_input(&line, encoding).map_err(|e| Arc::new(e) as ThreadSafeError)
    ```

    - 장점: 타입 안전성 높음, 성능 좋음
    - 단점: 유연성 떨어짐, 모든 에러 타입에 대해 개별 처리 필요

2. 에러를 감싸는 새로운 타입 정의:

    ```rust
    use std::error::Error;
    use std::fmt;

    struct ThreadSafeErrorWrapper(Box<dyn Error + Send + Sync>);

    impl Error for ThreadSafeErrorWrapper {
        fn source(&self) -> Option<&(dyn Error + 'static)> {
            self.0.source()
        }
    }

    impl fmt::Display for ThreadSafeErrorWrapper {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "{}", self.0)
        }
    }

    impl fmt::Debug for ThreadSafeErrorWrapper {
        fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
            write!(f, "ThreadSafeErrorWrapper({:?})", self.0)
        }
    }

    type ThreadSafeError = Arc<ThreadSafeErrorWrapper>;

    // 사용
    decode_input(&line, encoding).map_err(|e| Arc::new(ThreadSafeErrorWrapper(Box::new(e))))
    ```

    - 장점: 모든 에러 타입 처리 가능, 스레드 안전성 보장
    - 단점: 추가적인 코드 필요, 약간의 런타임 오버헤드

3. `Box<dyn Error + Send + Sync>` 사용:

    ```rust
    type ThreadSafeError = Arc<Box<dyn Error + Send + Sync>>;

    // 사용
    decode_input(&line, encoding).map_err(|e| Arc::new(Box::new(e) as Box<dyn Error + Send + Sync>))
    ```

    - 장점: 간단한 구현, 대부분의 경우 충분
    - 단점: 모든 에러 타입이 `Send + Sync`를 구현해야 함

상황에 따라 다르지만, 일반적으로 세 번째 방법인 `Box<dyn Error + Send + Sync>` 사용이 가장 균형 잡힌 접근 방식입니다.
이 방법은 대부분의 상황에서 잘 작동하며, 스레드 안전성을 보장합니다.

```rust
use std::error::Error;
use std::sync::Arc;

type ThreadSafeError = Arc<Box<dyn Error + Send + Sync>>;

fn main() -> Result<(), ThreadSafeError> {
    // ...

    reader
        .lines()
        .par_bridge()
        .try_for_each(|line| -> Result<(), ThreadSafeError> {
            let line = line.map_err(|e| Arc::new(Box::new(e) as Box<dyn Error + Send + Sync>))?;
            let decoded = decode_input(&line, encoding)
                .map_err(|e| Arc::new(Box::new(e) as Box<dyn Error + Send + Sync>))?;
            // ... 나머지 코드 ...

            Ok(())
        })?;

    Ok(())
}
```
