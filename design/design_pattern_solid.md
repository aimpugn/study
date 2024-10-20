# SOLID

- [SOLID](#solid)
    - [SOLID 개요](#solid-개요)
    - [SRP (Single Responsibility Principle) - 단일 책임 원칙](#srp-single-responsibility-principle---단일-책임-원칙)
    - [OCP (Open/Closed Principle) - 개방-폐쇄 원칙](#ocp-openclosed-principle---개방-폐쇄-원칙)
    - [LSP (Liskov Substitution Principle) - 리스코프 치환 원칙](#lsp-liskov-substitution-principle---리스코프-치환-원칙)
    - [ISP (Interface Segregation Principle) - 인터페이스 분리 원칙](#isp-interface-segregation-principle---인터페이스-분리-원칙)
    - [DIP (Dependency Inversion Principle) - 의존 역전 원칙](#dip-dependency-inversion-principle---의존-역전-원칙)

## SOLID 개요

SOLID는 객체 지향 프로그래밍(OOP)에서 유지보수성과 확장성을 높이기 위한 5가지 원칙의 약어입니다.
시스템의 모듈화와 결합도를 줄이고, 응집도를 높여 소프트웨어가 변경에 유연하게 대응할 수 있도록 합니다.
- 단일 책임(Single Responsibility)
- 개방/폐쇄(Open/Closed)
- 리스코프 치환(Liskov Substitution)
- 인터페이스 분리(Interface Segregation)
- 의존성 역전(Dependency Inversion)

소프트웨어의 구조를 모듈화하고, 각 모듈의 책임을 명확히 하며, *모듈 간의 [결합도를 낮추고 응집도를 높이는](./design_cohesion_coupling.md)* 방법을 제시합니다.
이를 통해 변화에 유연하게 대응할 수 있도록 하며, 코드의 재사용성, 유연성, 테스트 가능성을 높입니다.

SOLID 원칙은 다음과 같은 개념들에 기반합니다.
- 정보 은닉(Information Hiding): 내부 구현 세부사항을 외부로부터 숨기는 과정
- 추상화(Abstraction): 복잡한 시스템을 단순화하여 표현하는 과정
- 다형성(Polymorphism): 여러 타입의 객체가 동일한 인터페이스를 통해 다르게 동작하는 능력
- 계약에 의한 설계(Design by Contract): 소프트웨어 구성 요소 간의 책임과 의무를 명확히 정의하는 방법

## SRP (Single Responsibility Principle) - 단일 책임 원칙

단일 책임 원칙(SRP)은 클래스는 하나의 명확한 책임만 가져야 한다는 원칙입니다.
즉, 클래스는 오직 하나의 목적을 위해 존재하며, 그 목적에 대한 변경 사항만 반영해야 합니다.

SRP를 위반하면, 하나의 클래스가 여러 역할을 담당하게 되며, 한 역할의 변경이 다른 역할에 불필요한 영향을 줄 수 있습니다.
이는 유지보수성과 확장성을 저하시킵니다.

책임을 분리하면 코드 변경 시 그 변화가 제한된 범위에만 영향을 미칩니다.
다만, 책임의 경계를 정의하는 것이 때로는 주관적일 수 있다는 것에 주의합니다.

```kotlin
// SRP를 위반하는 예
// `Employee` 클래스는 세 가지 책임을 가지고 있습니다.
// - 급여 계산
// - 근무 시간 보고
// - 데이터베이스 저장이라
class Employee {
    fun calculatePay(): Double {
        // 급여 계산 로직
    }

    fun reportHours(): String {
        // 근무 시간 보고 로직
    }

    fun save() {
        // 데이터베이스에 저장하는 로직
    }
}

// SRP를 준수하는 예
class Employee(val name: String, val position: String)

// 각 책임을 별도의 클래스로 분리합니다.
// 이를 통해 각 클래스는 단일 책임을 가지게 되어, 변경 사유가 하나로 제한됩니다.
class PayCalculator {
    fun calculatePay(employee: Employee): Double {
        // 급여 계산 로직
    }
}

class HoursReporter {
    fun reportHours(employee: Employee): String {
        // 근무 시간 보고 로직
    }
}

class EmployeeRepository {
    fun save(employee: Employee) {
        // 데이터베이스에 저장하는 로직
    }
}
```

## OCP (Open/Closed Principle) - 개방-폐쇄 원칙

OCP는 기존의 코드를 수정하지 않고 새로운 기능을 추가할 수 있어야 한다는 원칙입니다.
소프트웨어 엔티티(클래스, 모듈, 함수 등)는 확장에는 열려(Open) 있어야 하지만, 변경에는 닫혀(Closed) 있어야 한다는 원칙입니다.
인터페이스나 추상 클래스를 사용하여 확장 지점을 만들고, 새로운 기능은 이를 구현하거나 상속받아 추가합니다.

OCP를 위반하면 새로운 기능을 추가할 때마다 기존 코드를 수정해야 합니다.
이는 시스템에 대한 테스트와 안정성에 악영향을 미치며, 작은 변경이 전체 시스템에 파급효과를 일으킬 수 있습니다.

다만 때로는 추상화로 인한 초기 설계의 복잡성 증가하거나 과도한 추상화로 이어질 수 있음에 주의합니다.

```kotlin
// OCP 위반 코드
class Discount {
    fun applyDiscount(type: String, price: Double): Double {
        return when (type) {
            "seasonal" -> price * 0.9
            "holiday" -> price * 0.8
            else -> price
        }
    }
}

// OCP를 지킨 코드
interface DiscountStrategy {
    fun applyDiscount(price: Double): Double
}

class SeasonalDiscount : DiscountStrategy {
    override fun applyDiscount(price: Double) = price * 0.9
}

class HolidayDiscount : DiscountStrategy {
    override fun applyDiscount(price: Double) = price * 0.8
}

// DiscountContext는 구체적인 할인 클래스에 의존하지 않고,
// DiscountStrategy 인터페이스에만 의존합니다.
// 새로운 할인 클래스가 추가되어도 DiscountContext를 수정할 필요가 없습니다.
class DiscountContext(private val strategy: DiscountStrategy) {
    fun applyDiscount(price: Double) = strategy.applyDiscount(price)
}

fun main() {
    val discount = DiscountContext(SeasonalDiscount())
    println(discount.applyDiscount(100.0))  // Seasonal Discount 적용
}
```

## LSP (Liskov Substitution Principle) - 리스코프 치환 원칙

자식 클래스(서브타입)는 언제나 부모 클래스(기반 타입)를 대체할 수 있어야 한다는 원칙입니다.
LSP는 상속 관계에 있는 객체들이 서로 대체 가능해야 함을 강조합니다.

상속 관계에서 자식 클래스가 부모 클래스의 동작을 완전히 준수해야 하며, 자식 클래스가 부모 클래스 대신 사용되더라도 코드가 정상적으로 작동해야 한다는 의미입니다.
즉, 부모 클래스의 인스턴스가 사용되는 모든 곳에서 자식 클래스의 인스턴스로 대체하더라도 프로그램의 정확성이 유지되어야 합니다.
이는 상속을 올바르게 사용하는 방법을 제시합니다.

LSP를 위반하면 다형성이 깨집니다.
부모 클래스를 기대하는 곳에서 자식 클래스를 사용했을 때 의도치 않은 동작을 일으킬 수 있습니다.

단, 때로는 복잡한 계층 구조를 만들어낼 수 있으며, 상속 관계를 설계할 때 너무 많은 고민을 하게 될 수 있음에 주의합니다.

```kotlin
// LSP를 위반한 코드
open class Bird {
    open fun fly() {
        println("Flying")
    }
}

class Penguin : Bird() {
    // 펭귄이 날 수 없다는 특성 때문에 `fly()` 메소드를 호출하면 예외가 발생합니다.
    override fun fly() {
        throw UnsupportedOperationException("Penguins can't fly")
    }
}
```

LSP를 위반한 예에서 `Penguin` 클래스는 `fly()` 메서드를 호출하면 익셉션이 발생합니다.
이는 서브타입이 부모 타입의 계약을 지키지 않는 경우를 보여줍니다.

이를 해결하기 위해, 날 수 있는 새와 그렇지 않은 새를 구분하여 `FlyingBird` 클래스를 도입함으로써 문제를 해결할 수 있습니다.
`Penguin` 클래스는 더 이상 잘못된 기능을 오버라이드하지 않으므로, LSP를 준수하게 됩니다.

```kotlin
// LSP를 지킨 코드
open class Bird

open class FlyingBird : Bird() {
    open fun fly() {
        println("Flying")
    }
}

class Penguin : Bird()

class Sparrow : FlyingBird() {
    override fun fly() {
        println("Sparrow is flying")
    }
}
```

## ISP (Interface Segregation Principle) - 인터페이스 분리 원칙

클라이언트는 자신이 사용하지 않는 메서드에 의존해서는 안 됩니다.
클라이언트가 불필요한 의존성을 가지지 않도록 작고 명확한 인터페이스를 설계하는 것이 중요하다는 것을 강조합니다.
다르게 말하면, 범용 인터페이스 하나보다 특정 클라이언트를 위한 인터페이스 여러 개가 낫습니다.

ISP를 위반하면 클라이언트가 자신이 사용하지 않는 메소드까지 의존하게 되어, 불필요한 변경에 영향을 받을 수 있습니다.
인터페이스가 지나치게 크면 구현체가 인터페이스의 모든 메소드를 구현해야 하는 부담이 생깁니다.

단, 인터페이스의 수가 증가하여 관리해야 할 요소가 많아질 수 있고 때로는 인터페이스 간의 관계가 복잡해질 수 있음에 주의합니다.

```kotlin
// ISP를 위반하는 예
interface Worker {
    fun work()
    fun eat()
    fun sleep()
}

class Human : Worker {
    override fun work() {
        println("Human is working")
    }

    override fun eat() {
        println("Human is eating")
    }

    override fun sleep() {
        println("Human is sleeping")
    }
}

class Robot : Worker {
    override fun work() {
        println("Robot is working")
    }

    override fun eat() {
        // 로봇은 먹지 않음
        throw UnsupportedOperationException("Robots don't eat")
    }

    override fun sleep() {
        // 로봇은 자지 않음
        throw UnsupportedOperationException("Robots don't sleep")
    }
}
```

ISP를 위반하는 예에서 `Robot` 클래스는 사용하지 않는 `eat()`와 `sleep()` 메서드를 구현해야 했습니다. 이는 불필요한 의존성을 만들어냅니다.

아래의 ISP를 준수하는 예에서는 `Worker` 인터페이스를 `Workable`, `Eatable`, `Sleepable`로 분리하여 각 클래스가 필요한 인터페이스만 구현할 수 있도록 합니다.
이렇게 하면 `Robot` 클래스는 `Workable` 인터페이스만 구현하면 되므로, 불필요한 메서드에 대한 의존성이 제거됩니다.

```kotlin
// ISP를 준수하는 예
interface Workable {
    fun work()
}

interface Eatable {
    fun eat()
}

interface Sleepable {
    fun sleep()
}

class Human : Workable, Eatable, Sleepable {
    override fun work() {
        println("Human is working")
    }

    override fun eat() {
        println("Human is eating")
    }

    override fun sleep() {
        println("Human is sleeping")
    }
}

class Robot : Workable {
    override fun work() {
        println("Robot is working")
    }
}
```

## DIP (Dependency Inversion Principle) - 의존 역전 원칙

*고수준 모듈(정책)*은 *저수준 모듈(세부 사항)*에 의존해서는 안 되며, 추상화된 인터페이스나 추상 클래스에 의존해야 합니다.
추상화는 세부 사항에 의존해서는 안 되며, 세부 사항이 추상화에 의존해야 합니다.

DIP는 시스템의 결합도를 낮추고 유연성을 높이는 데 중점을 둡니다.
이 원칙은 의존성 주입(Dependency Injection)과 밀접한 관련이 있으며, 인터페이스나 추상 클래스를 통해 구현됩니다.
고수준 모듈과 저수준 모듈 사이에 추상화 계층을 두어 직접적인 의존성을 제거합니다.

DIP를 위반하면 구체적인 구현에 의존하게 되어, 변경에 취약해지고, 코드의 재사용성이 떨어집니다. 이를 방지하기 위해, 클래스들은 구체적인 클래스보다는 인터페이스나 추상 클래스와 같은 추상화된 것에 의존해야 합니다.

단, 추가적인 추상화 계층으로 인한 복잡성 증가할 수 있고 때로는 과도한 추상화로 이어질 수 있다는 점을 주의합니다.

> IoC(Inversion of Control)과 DI의 관계
>
> IoC의 구현 방식 중 하나가 *의존성 주입(Dependency Injection, DI)*입니다.
>
> IoC는 객체 간의 제어 흐름을 반전시키는 디자인 패턴이나 원칙을 설명하는 개념입니다.
> 전통적인 방식에서는 상위 모듈이 하위 모듈을 직접 제어하지만,
> IoC를 적용하면 제어의 흐름을 외부 컨테이너나 프레임워크가 관리하게 되어,
> 상위 모듈과 하위 모듈 간의 결합도가 낮아집니다.
>
> DI는 객체가 필요한 의존성을 스스로 생성하지 않고, 외부에서 주입받는 방식으로 제어를 역전시키는 패턴입니다.
>
> DI 외에는 다음과 같은 방식들로 구현할 수 있습니다.
> - 서비스 로케이터 (Service Locator):
>
>   객체가 필요한 의존성을 전역 서비스나 레지스트리에서 직접 가져오는 방식입니다.
>   이 레지스트리는 객체들이 필요로 하는 서비스들을 조회할 수 있는 일종의 저장소 역할을 합니다.
>
> - 콜백 (Callback):
>
>   어떤 작업을 수행할 때, 콜백 함수를 외부에서 전달받아 그 함수가 특정 시점에서 호출되는 방식입니다.
>
> - 템플릿 메서드 패턴 (Template Method Pattern):
>
>   상위 클래스에서 일련의 알고리즘 흐름을 정의하고, 일부 구체적인 작업은 하위 클래스에서 정의하도록 하는 방식입니다.
>   상위 클래스가 제어 흐름을 결정하지만, 구체적인 구현은 하위 클래스에 위임하는 방식입니다.

```kotlin
// DIP를 위반하는 예
class LightBulb {
    fun turnOn() {
        println("LightBulb: Bulb turned on...")
    }

    fun turnOff() {
        println("LightBulb: Bulb turned off...")
    }
}

class ElectricPowerSwitch(private val bulb: LightBulb) {
    private var isOn = false

    fun press() {
        isOn = !isOn
        if (isOn) {
            bulb.turnOn()
        } else {
            bulb.turnOff()
        }
    }
}
```

DIP를 위반하는 예에서 `ElectricPowerSwitch`는 `LightBulb`에 직접 의존합니다.
이는 `ElectricPowerSwitch`의 재사용성을 제한하고, 다른 종류의 장치에 대해 사용할 수 없게 만듭니다.

DIP를 준수하는 예에서는 `Switchable` 인터페이스를 도입하여 `ElectricPowerSwitch`가 구체적인 구현이 아닌 추상화에 의존하도록 합니다.
이를 통해 `ElectricPowerSwitch`는 `LightBulb`, `Fan` 등 `Switchable` 인터페이스를 구현하는 어떤 장치와도 작동할 수 있게 됩니다.

```kotlin
// DIP를 준수하는 예
interface Switchable {
    fun turnOn()
    fun turnOff()
}

class LightBulb : Switchable {
    override fun turnOn() {
        println("LightBulb: Bulb turned on...")
    }

    override fun turnOff() {
        println("LightBulb: Bulb turned off...")
    }
}

class Fan : Switchable {
    override fun turnOn() {
        println("Fan: Fan started...")
    }

    override fun turnOff() {
        println("Fan: Fan stopped...")
    }
}

class ElectricPowerSwitch(private val device: Switchable) {
    private var isOn = false

    fun press() {
        isOn = !isOn
        if (isOn) {
            device.turnOn()
        } else {
            device.turnOff()
        }
    }
}

// 사용 예
fun main() {
    val bulb = LightBulb()
    val bulbPowerSwitch = ElectricPowerSwitch(bulb)
    bulbPowerSwitch.press()
    bulbPowerSwitch.press()

    val fan = Fan()
    val fanPowerSwitch = ElectricPowerSwitch(fan)
    fanPowerSwitch.press()
    fanPowerSwitch.press()
}
```
