# lazy

- [lazy](#lazy)
    - [늦은 평가(Lazy Evaluation)](#늦은-평가lazy-evaluation)
    - [Historical Context and Evolution](#historical-context-and-evolution)
    - [용어들](#용어들)
    - [이론적 근거](#이론적-근거)
        - [Mathematical Model](#mathematical-model)
        - [Algorithmic Representation](#algorithmic-representation)
    - [구현 예제](#구현-예제)
        - [Rust 예제](#rust-예제)
        - [Go 예제](#go-예제)
        - [Java 예제](#java-예제)
        - [Best Practices and Design Patterns](#best-practices-and-design-patterns)
    - [늦은 평가 사용시 속도가 빠른 이유](#늦은-평가-사용시-속도가-빠른-이유)
    - [늦은 평가 사용시 단점](#늦은-평가-사용시-단점)
    - [기타](#기타)

## 늦은 평가(Lazy Evaluation)

늦은 평가(Lazy Evaluation)는 표현식의 결과값이 실제로 필요할 때까지 계산(평가)을 미루는 평가 전략입니다.
당장 불필요한 계산을 필요한 시점으로 미룸으로써 불필요한 계산을 피하고, 잠재적으로 무한한 데이터 구조를 다룰 수 있게 하며, 특정 상황에서 성능을 향상시킬 수 있습니다.

반대로 [즉시 평가(Eager Evaluation, 또는 엄격한 평가)](https://ko.wikipedia.org/wiki/%EC%A1%B0%EA%B8%89%ED%95%9C_%EA%B3%84%EC%82%B0%EB%B2%95)는 즉시 계산(평가)이 이뤄집니다.

늦은 평가는 주로 함수형 프로그래밍 언어에서 널리 사용됩니다.
Rust와 같은 시스템 프로그래밍 언어에서도 효율성을 높이기 위해 사용됩니다.

Apache Spark에서 지연 변환(Lazy Transformation), Django의 QuerySet, 게임 엔진에서 복잡한 게임 월드의 효율적인 렌더링 등의 경우에 지연 평가를 사용합니다.

## Historical Context and Evolution

늦은 평가의 개념은 1970년대에 함수형 프로그래밍 언어의 발전과 함께 등장했습니다.
Peter Henderson과 James H. Morris Jr.가 1976년에 발표한 논문 ["A Lazy Evaluator"](https://dl.acm.org/doi/10.1145/800168.811543)에서 이 개념을 공식화했습니다.
이후 Haskell과 같은 순수 함수형 언어에서 기본 평가 전략으로 채택되었습니다.

## 용어들

- 평가 전략(Evaluation Strategy): 프로그램의 표현식을 언제, 어떻게 평가할지 결정하는 규칙
- 클로저(Closure): 함수와 그 함수가 참조하는 환경을 포함하는 데이터 구조
- 메모이제이션(Memoization): 이전에 계산한 결과를 저장하여 재사용하는 최적화 기법
- [Thunk](https://wiki.haskell.org/Thunk): 평가되지 않은 값

## 이론적 근거

늦은 평가는 람다 대수(Lambda Calculus)와 범주론(Category Theory)의 개념을 기반으로 합니다.
특히, 모나드(Monad)와 같은 추상적인 구조를 통해 형식화됩니다.

### Mathematical Model

늦은 평가의 수학적 모델은 다음과 같이 표현될 수 있습니다:

$\text{lazy} \; e = \lambda(). \; e$

$\text{force} \; (\text{lazy} \; e) = e$

- $e$: 표현식입니다.
- $\lambda()$:

    람다 대수(Lambda calculus)에서 함수를 나타냅니다.
    여기서는 지연 함수 생성자를 가리킵니다.

- force: 지연된 계산을 강제로 실행하는 함수입니다.

이 모델에서 `lazy e`는 표현식 `e`의 평가를 지연시키는 함수를 생성합니다. `force`는 이 지연된 계산을 실제로 수행합니다.

늦은 평가의 의미론(Semantics)을 형식화하면 다음과 같습니다:

```
E⟦x⟧ρ = ρ(x)
E⟦λx.e⟧ρ = λv.E⟦e⟧ρ[x↦v]
E⟦e1 e2⟧ρ = (E⟦e1⟧ρ) (λ().E⟦e2⟧ρ)

where:
E⟦-⟧: 표현식의 의미를 나타내는 함수
ρ: 환경 (변수에 값을 매핑)
λv.-: 지연된 인자를 나타내는 함수 추상화
```

이 의미론은 함수 적용 시 인자의 평가를 지연시키는 늦은 평가의 핵심 아이디어를 형식화합니다.

### Algorithmic Representation

늦은 평가의 기본 알고리즘은 다음과 같습니다:

```python
from functools import wraps

# Thunk 클래스로 연된 계산을 명시적으로 표현
class Thunk:
    def __init__(self, func, *args, **kwargs):
        self.func = func
        self.args = args
        self.kwargs = kwargs
        self._result = None
        self._evaluated = False

    def force(self):
        if not self._evaluated:
            self._result = self.func(*self.args, **self.kwargs)
            self._evaluated = True
        return self._result

def lazy(func):
    # `functools.wraps`를 사용하여 원본 함수의 메타데이터를 보존
    @wraps(func)
    def wrapper(*args, **kwargs):
        return Thunk(func, *args, **kwargs)
    return wrapper

# 사용 예
@lazy
def expensive_computation(x, y):
    print("Computing...")
    return x + y

result = expensive_computation(1, 2)  # 아직 계산되지 않음
print(result.force())  # 이 시점에 계산됨
print(result.force())  # 캐시된 결과 반환
```

## 구현 예제

늦은 평가의 핵심 아이디어는 다음과 같습니다:
1. 계산을 즉시 수행하지 않고 함수나 클로저 형태로 저장
2. 실제 값이 필요할 때까지 계산을 지연
3. 한 번 계산된 값은 저장하여 재사용(메모이제이션)

이를 통해 불필요한 계산을 피하고, 리소스를 효율적으로 사용할 수 있습니다.
특히 계산 비용이 높거나 결과가 항상 사용되지 않는 경우에 유용합니다.

### Rust 예제

```rust
use std::cell::RefCell;

// 늦은 평가를 위한 구조체
struct Lazy<T> {
    calculation: RefCell<Option<Box<dyn Fn() -> T>>>,
    value: RefCell<Option<T>>,
}

impl<T> Lazy<T> {
    // 새로운 Lazy 인스턴스 생성
    fn new<F: Fn() -> T + 'static>(f: F) -> Lazy<T> {
        Lazy {
            calculation: RefCell::new(Some(Box::new(f))),
            value: RefCell::new(None),
        }
    }

    // 값을 가져오거나 계산
    fn get(&self) -> T
    where
        T: Clone,
    {
        if self.value.borrow().is_none() {
            // 값이 아직 계산되지 않았다면 계산을 수행
            let calc = self.calculation.borrow_mut().take().unwrap();
            let result = calc();
            *self.value.borrow_mut() = Some(result);
        }
        // 계산된 값을 반환
        self.value.borrow().clone().unwrap()
    }
}

fn main() {
    // 늦은 평가 인스턴스 생성
    // 이 시점에서는 계산이 수행되지 않음
    let lazy_value = Lazy::new(|| {
        println!("Expensive calculation performed!");
        42
    });

    println!("Lazy value created");

    // 첫 번째 get() 호출 시 계산 수행
    println!("First access: {}", lazy_value.get());

    // 두 번째 get() 호출 시 이미 계산된 값을 반환
    println!("Second access: {}", lazy_value.get());
}
```

Rust에서의 늦은 평가:
1. `Lazy` 구조체를 사용하여 계산을 지연시킵니다.
2. `calculation`은 실제 계산을 수행할 함수를 저장합니다.
3. `value`는 계산된 결과를 캐시합니다.
4. `get()` 메서드가 호출될 때까지 실제 계산은 수행되지 않습니다.
5. 첫 `get()` 호출 시 계산이 수행되고, 이후 호출에서는 캐시된 값을 반환합니다.

### Go 예제

```go
package main

import (
    "fmt"
    "sync"
)

// Lazy 구조체 정의
type Lazy struct {
    once     sync.Once
    value    interface{}
    producer func() interface{}
}

// 새로운 Lazy 인스턴스 생성
func NewLazy(producer func() interface{}) *Lazy {
    return &Lazy{producer: producer}
}

// 값을 가져오거나 계산
func (l *Lazy) Get() interface{} {
    // sync.Once를 사용하여 한 번만 계산이 수행되도록 보장
    l.once.Do(func() {
        l.value = l.producer()
    })
    return l.value
}

func main() {
    // 늦은 평가 인스턴스 생성
    // 이 시점에서는 계산이 수행되지 않음
    lazyValue := NewLazy(func() interface{} {
        fmt.Println("Expensive calculation performed!")
        return 42
    })

    fmt.Println("Lazy value created")

    // 첫 번째 Get() 호출 시 계산 수행
    fmt.Println("First access:", lazyValue.Get())

    // 두 번째 Get() 호출 시 이미 계산된 값을 반환
    fmt.Println("Second access:", lazyValue.Get())
}
```

Go에서의 늦은 평가:
1. `Lazy` 구조체를 사용하여 계산을 지연시킵니다.
2. `sync.Once`를 사용하여 계산이 정확히 한 번만 수행되도록 보장합니다.
3. `producer`는 실제 계산을 수행할 함수를 저장합니다.
4. `Get()` 메서드가 호출될 때까지 실제 계산은 수행되지 않습니다.
5. 첫 `Get()` 호출 시 계산이 수행되고, 이후 호출에서는 저장된 값을 반환합니다.

### Java 예제

```java
import java.util.function.Supplier;

public class Main {
    // Lazy 클래스 정의
    static class Lazy<T> {
        private Supplier<T> producer;
        private T value;
        private boolean initialized = false;

        // 생성자
        public Lazy(Supplier<T> producer) {
            this.producer = producer;
        }

        // 값을 가져오거나 계산
        public T get() {
            if (!initialized) {
                // 값이 아직 계산되지 않았다면 계산을 수행
                value = producer.get();
                initialized = true;
            }
            return value;
        }
    }

    public static void main(String[] args) {
        // 늦은 평가 인스턴스 생성
        // 이 시점에서는 계산이 수행되지 않음
        Lazy<Integer> lazyValue = new Lazy<>(() -> {
            System.out.println("Expensive calculation performed!");
            return 42;
        });

        System.out.println("Lazy value created");

        // 첫 번째 get() 호출 시 계산 수행
        System.out.println("First access: " + lazyValue.get());

        // 두 번째 get() 호출 시 이미 계산된 값을 반환
        System.out.println("Second access: " + lazyValue.get());
    }
}
```

Java에서의 늦은 평가:
1. `Lazy` 클래스를 사용하여 계산을 지연시킵니다.
2. `Supplier` 인터페이스를 사용하여 계산을 표현합니다.
3. `initialized` 플래그를 사용하여 계산이 수행되었는지 추적합니다.
4. `get()` 메서드가 호출될 때까지 실제 계산은 수행되지 않습니다.
5. 첫 `get()` 호출 시 계산이 수행되고, 이후 호출에서는 저장된 값을 반환합니다.

### Best Practices and Design Patterns

1. 메모이제이션 사용: 한 번 계산된 값을 저장하여 재사용합니다.
2. 스트림과 제너레이터 활용: 무한한 데이터 구조를 효율적으로 다룹니다.
3. 부작용 최소화: 순수 함수를 사용하여 예측 가능성을 높입니다.

## 늦은 평가 사용시 속도가 빠른 이유

- 불필요한 계산 회피: 사용되지 않는 값은 계산하지 않습니다.

    필요한 순간까지 평가를 미룸으로써, 필요하지 않은 경우에는 아예 계산을 하지 않으므로, 전체적인 계산 양이 줄어듭니다.

    필요한 계산만 수행하여 자원을 절약하고 성능을 높일 수 있습니다.
    또한 불필요한 계산을 하지 않음으로써 코드가 더 단순해질 수 있습니다.

- 메모리 사용 최적화: 필요한 시점에만 메모리를 할당합니다.

    모든 표현식을 즉시 평가하지 않으므로, 메모리 사용량이 줄어들고, 프로그램이 더 많은 메모리 공간을 효율적으로 사용할 수 있습니다.

- 병렬성 활용: 독립적인 지연 계산을 병렬로 수행할 수 있습니다.

    특정 작업이 필요한 순간까지 지연되므로, 다른 작업들이 먼저 실행될 수 있습니다.
    이는 프로그램의 병렬성을 높이는 데 기여합니다.

    프로그램 로직의 흐름에 따라 계산 순서를 유연하게 조절할 수 있습니다.

## 늦은 평가 사용시 단점

- 평가 시점이 코드의 작성 순서와 다를 수 있기 때문에, 예상치 못한 시점에 버그가 발생할 수 있습니다. 이로 인해 디버깅이 어려울 수 있습니다.
- 잘못 사용시 참조가 유지된 상태에서 평가되지 않은 표현식이 메모리에 남아 있을 수 있습니다. 이 때문에 메모리 누수 가능성이 있습니다.

## 기타

- [Functional Programming in Scala](https://www.manning.com/books/functional-programming-in-scala), Paul Chiusano and Rúnar Bjarnason
- [Introduction to Functional Programming](https://www.amazon.com/Introduction-Functional-Programming-Calculus-Mathematics/dp/0201178141), Richard Bird and Philip Wadler
- [Purely Functional Data Structures](https://www.amazon.com/Purely-Functional-Structures-Chris-Okasaki/dp/0521663504), Chris Okasaki
-
