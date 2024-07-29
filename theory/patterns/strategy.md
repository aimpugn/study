# Strategy

- [Strategy](#strategy)
    - [전략 패턴?](#전략-패턴)
    - [Examples](#examples)
        - [(Rust) `Context`](#rust-context)
        - [(Java) 핵심 로직 외의 필수적인 로직들을 강제](#java-핵심-로직-외의-필수적인-로직들을-강제)
        - [(Go) 핵심 로직 외의 필수적인 로직들을 강제](#go-핵심-로직-외의-필수적인-로직들을-강제)

## 전략 패턴?

스트래티지 패턴은 알고리즘의 교체 가능성을 목적으로 한다.
즉, 동일한 문제를 해결할 수 있는 여러 알고리즘(전략)이 있을 때, 이들 중에서 선택하여 사용할 수 있는 구조를 제공한다.
이 패턴은 객체의 행동 또는 알고리즘을 캡슐화하고, 이를 객체의 필드로서 사용한다.
스트래티지 패턴은 주로 "행동"의 변화에 초점을 맞추며, 이를 통해 알고리즘의 교체나 추가가 용이하게 한다.

스트래티지 패턴은 실행 시간(runtime)에 알고리즘(이 경우 파싱 로직)을 선택할 수 있도록 한다.

## Examples

### (Rust) `Context`

이 방법은 파서 로직을 쉽게 교체할 수 있도록 하여 유연성을 제공한다.

```rs
struct Context {
    parser: Box<dyn Parser>,
}

impl Context {
    fn new(parser: Box<dyn Parser>) -> Self {
        Context { parser }
    }

    fn parse(&self, input: &str) -> Value {
        self.parser.parse(input)
    }
}
```

### (Java) 핵심 로직 외의 필수적인 로직들을 강제

```java
// 전략 인터페이스
interface PaymentStrategy {
    public void pay(int amount);
}

// 구체적인 전략 1
class CreditCardPaymentStrategy implements PaymentStrategy {
    private String name;
    private String cardNumber;

    public CreditCardPaymentStrategy(String nm, String ccNum) {
        this.name = nm;
        this.cardNumber = ccNum;
    }

    @Override
    public void pay(int amount) {
        System.out.println(amount +" paid with credit/debit card.");
    }
}

// 구체적인 전략 2
class PaypalPaymentStrategy implements PaymentStrategy {
    private String emailId;

    public PaypalPaymentStrategy(String email) {
        this.emailId = email;
    }

    @Override
    public void pay(int amount) {
        System.out.println(amount + " paid using Paypal.");
    }
}

// 클라이언트 클래스
public class StrategyPatternDemo {
    public static void main(String[] args) {
        PaymentStrategy creditCardPayment = new CreditCardPaymentStrategy("John Doe", "123456789");
        creditCardPayment.pay(100);

        PaymentStrategy paypalPayment = new PaypalPaymentStrategy("myemail@example.com");
        paypalPayment.pay(200);
    }
}
```

### (Go) 핵심 로직 외의 필수적인 로직들을 강제

전략 패턴은 Go에서 인터페이스를 사용하여 쉽게 구현할 수 있다.
인터페이스를 통해 다양한 전략을 정의하고, 구조체를 사용하여 이 인터페이스를 구현할 수 있다.

```go
package main

import "fmt"

// PaymentStrategy 인터페이스는 pay 메서드를 정의합니다.
type PaymentStrategy interface {
    pay(amount int)
}

// CreditCardPaymentStrategy 구조체는 PaymentStrategy 인터페이스를 구현합니다.
type CreditCardPaymentStrategy struct {}

func (c *CreditCardPaymentStrategy) pay(amount int) {
    fmt.Printf("%d paid with credit/debit card.\n", amount)
}

// PaypalPaymentStrategy 구조체는 PaymentStrategy 인터페이스를 구현합니다.
type PaypalPaymentStrategy struct {}

func (p *PaypalPaymentStrategy) pay(amount int) {
    fmt.Printf("%d paid using Paypal.\n", amount)
}

// Client 코드에서는 인터페이스를 통해 다양한 결제 전략을 사용할 수 있습니다.
func main() {
    var payment PaymentStrategy

    payment = &CreditCardPaymentStrategy{}
    payment.pay(100)

    payment = &PaypalPaymentStrategy{}
    payment.pay(200)
}
```
