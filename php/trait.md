# trait

- [trait](#trait)
    - [trait?이란?](#trait이란)
    - [PHP에서의 Trait](#php에서의-trait)
    - [우선순위](#우선순위)
    - [예제](#예제)
    - [각 언어에서의 Trait](#각-언어에서의-trait)
        - [Rust](#rust)
        - [Haskell](#haskell)

## [trait](https://www.php.net/manual/en/language.oop5.traits.php)?이란?

**Trait**는 객체 지향 프로그래밍에서 *클래스들이 공통적으로 가져야 할 메서드들을 정의하는 방법*입니다.
- 클래스에 메서드를 추가하는 방식으로 코드 재사용성을 높이는 것을 목표로 합니다.
- Trait은 독립적인 메서드 집합으로, 특정 클래스에 혼합(mix)될 수 있습니다.

다른 프로그래밍 언어에서는 인터페이스(interface)로 알려져 있습니다.

Trait은 일종의 계약으로, 특정 메서드들을 구현해야 하는지 명시합니다.
이로 인해 여러 클래스들이 동일한 동작을 공유할 수 있습니다.

## PHP에서의 Trait

Trait은 PHP와 같은 단일 상속 언어에서 단일 상속의 한계를 극복하고, 복잡성을 줄이면서 코드 재사용성을 극대화하는 중요한 도구입니다.
클래스는 단일 부모 클래스만 상속받을 수 있지만, 여러 Trait을 포함하여 다양한 기능을 조합할 수 있습니다.
즉, Trait을 통해 클래스 계층 구조와 상관없이 다양한 클래스에서 공통된 기능을 손쉽게 공유할 수 있습니다.

Trait은 개발자가 서로 다른 클래스 계층에 속하는 여러 독립적인 클래스에서 메서드 집합을 자유롭게 재사용할 수 있도록 합니다.
트레잇과 클래스 조합의 의미론은 복잡성을 줄이고, 다중 상속과 Mixin에서 발생하는 일반적인 문제를 피하도록 정의되어 있습니다.

> **트레잇과 클래스 조합의 의미론**(Semantics of the combination of Traits and classes)
>
> 트레잇과 클래스의 조합이 어떻게 동작하는지에 대한 의미론적 정의를 말합니다.
> 의미론(semantics)은 프로그래밍 언어에서 코드의 의미와 동작을 설명하는 개념입니다.
> 이는 코드가 어떻게 실행되고, 어떤 결과를 초래하는지를 정의합니다.
> 의미론은 프로그래밍 언어의 구문(syntax)과는 달리, 코드의 의미와 그 실행 결과에 초점을 맞춥니다
>
> 트레잇과 클래스의 조합에서 의미론은 다음과 같은 방식으로 문제를 해결합니다:
>
> - **코드 플래트닝**: 트레잇의 메서드와 속성이 클래스에 직접 포함되어, 다중 상속의 복잡성을 피할 수 있습니다.
> - **메서드 우선순위**: 동일한 이름의 메서드가 있을 때 명시적으로 우선순위를 정하거나 충돌을 해결할 수 있습니다.
> - **추상 메서드 요구사항**: 트레잇이 요구하는 기능을 명확히 정의하여, 클래스가 이를 구현하도록 강제합니다.
> - **일관된 동작**: 클래스 연산자가 트레잇 내에서 일관되게 동작하여 예측 가능한 코드를 작성할 수 있습니다.
>
> 이러한 의미론적 정의는 프로그래밍 언어의 설계, 타입 이론, 컴파일러 및 인터프리터, 프로그램 검증 및 모델 검증 등과 밀접한 관련이 있습니다.
> 의미론은 프로그래밍 언어의 동작을 수학적으로 모델링하여, 코드의 정확성과 일관성을 보장하는 데 중요한 역할을 합니다.

PHP에서 Trait은 `trait` 키워드를 사용하여 정의되며, 클래스 내에서 `use` 키워드를 사용하여 포함됩니다.

PHP에서의 트레잇(trait)은 소프트웨어 엔지니어링에서 일반적으로 정의된 트레잇 개념에 부합합니다.
트레잇은 여러 언어에서 코드 재사용을 촉진하고 다중 상속의 복잡성을 피하기 위해 사용됩니다.
PHP의 트레잇은 다른 언어에서의 트레잇과 유사한 목적과 기능을 가지고 있지만, 각 언어의 특성에 따라 구현 방식과 사용 방법에 차이가 있을 수 있습니다.

- **코드 재사용**

    Trait는 클래스로 정의된 기능의 집합으로, 클래스가 이를 포함하여 사용할 수 있습니다.
    여러 클래스에서 동일한 기능을 필요로 할 때 Trait을 사용하면 여러 클래스에서 공통 기능을 재사용할 수 있습니다.

- **인스턴스화 불가**

    Trait는 클래스처럼 보이지만, 자체적으로 인스턴스화할 수 없습니다.
    기능을 세밀하고 일관된 방식으로 그룹화하는 데만 사용됩니다.
    이는 전통적인 상속에 대한 추가 기능으로, 상속 없이 클래스 멤버를 적용할 수 있는 수평적 행동 구성을 가능하게 합니다.

- **복잡성 감소**

    다중 상속이나 Mixin을 사용하면 발생할 수 있는 복잡한 문제를 피하면서도, 코드의 일관성과 재사용성을 유지합니다.

- **유연한 설계**

    클래스 계층 구조를 단순하게 유지하면서도 다양한 기능을 추가할 수 있습니다.

- **다중 상속 문제 해결**

    PHP는 단일 상속만을 지원하지만, Trait을 사용하면 여러 기능을 조합할 수 있어 다중 상속의 장점을 얻을 수 있습니다.

- **코드의 모듈화**

    공통 기능을 Trait으로 분리하여 모듈화된 코드 구조를 만들 수 있습니다.

- **상태를 가질 수 있음**

    Trait는 상태(속성)를 가질 수 있지만, 주로 메서드를 그룹화하는 데 사용됩니다.

- **테스트 용이성**

    공통 기능을 별도의 Trait으로 분리하여 독립적으로 테스트할 수 있습니다.

- **코드 플래트닝**

    트레잇은 클래스에 포함될 때, 마치 클래스의 일부인 것처럼 동작합니다.
    이는 트레잇의 메서드와 속성이 클래스에 직접 포함되는 것처럼 보이게 합니다.
    따라서 트레잇을 사용하는 클래스는 트레잇을 사용하지 않는 클래스와 런타임에서 구별되지 않습니다

    예를 들어, `TraitA`와 `ClassB`가 있을 때, `ClassB`가 `TraitA`를 사용하면
    `TraitA`의 메서드와 속성이 `ClassB`의 일부가 됩니다.

    ```php
    trait TraitA {
        public function methodA() {
            echo "Method A";
        }
    }

    class ClassB {
        use TraitA;
    }

    $obj = new ClassB();
    $obj->methodA(); // "Method A" 출력
    ```

- **메서드 우선순위**

    트레잇과 클래스가 동일한 이름의 메서드를 가질 경우, 클래스의 메서드가 우선합니다.
    트레잇 간의 충돌이 발생하면 이를 명시적으로 해결해야 합니다.
    그렇지 않으면 충돌하는 메서드가 제외됩니다.
    PHP에서는 메서드 충돌 해결을 위해 `insteadof`와 `as` 키워드를 제공합니다.

    예를 들어, `TraitA`와 `ClassB`가 동일한 `methodA`를 가질 때, `ClassB`의 `methodA`가 호출됩니다.

    ```php
    class ClassBase {
        public function methodA() {
            echo "ClassBase Method A\n";
        }
    }

    trait TraitA {
        public function methodA() {
            echo "TraitA Method A\n";
        }
    }

    class ClassB extends ClassBase {
        use TraitA;

        public function methodA() {
            echo "ClassB Method A\n";
        }
    }

    $obj = new ClassB();
    $obj->methodA(); // "ClassB Method A" 출력
    ```

- **추상 메서드 요구사항**

    트레잇은 추상 메서드를 정의할 수 있으며, 이를 사용하는 클래스는 해당 메서드를 구현해야 합니다.
    이는 트레잇이 특정 기능을 요구할 수 있게 합니다
  
    예를 들어, `TraitA`가 추상 메서드 `methodA`를 정의하면, 이를 사용하는 `ClassB`는 `methodA`를 구현해야 합니다.

    ```php
    trait TraitA {
        abstract public function methodA();
    }

    class ClassB {
        use TraitA;

        public function methodA() {
            echo "Implemented Method A";
        }
    }

    $obj = new ClassB();
    $obj->methodA(); // "Implemented Method A" 출력
    ```

- **클래스 연산자**

    트레잇 내에서 `parent::`, `self::`, `$this`와 같은 클래스 연산자는 트레잇이 포함된 클래스에서 정의된 것처럼 동작합니다

    예를 들어, `TraitA`가 `self::methodB`를 호출하면, 이는 `ClassB`의 `methodB`를 호출하는 것과 동일하게 동작합니다.

    ```php
    trait TraitA {
        public function methodA() {
            self::methodB();
        }

        public function methodB() {
            echo "TraitA Method B";
        }
    }

    class ClassB {
        use TraitA;

        public function methodB() {
            echo "ClassB Method B";
        }
    }

    $obj = new ClassB();
    $obj->methodA(); // "ClassB Method B" 출력
    ```

## 우선순위

- 현재 클래스의 메서드가 Trait 메서드를 재정의하고, 이 메서드가 다시 부모 클래스의 메서드를 재정의합니다.

    부모 클래스의 메서드를 상속하지만, trait이 자신의 메서드처럼 포함되므로 trait의 메서드가 우선한다.
    하지만 자신의 클래스에서 method를 재정의하면, 그 재정의한 메서드가 trait의 메서드보다 우선한다.

    ```php
    <?php
    class Base {
        public function sayHello() {
            echo 'Hello ';
        }
    }

    trait SayWorld {
        public function sayHello() {
            parent::sayHello();
            echo 'World!';
        }
    }

    // 현재 클래스의 메서드가 Trait 메서드를 재정의
    class MyHelloWorld extends Base {
        use SayWorld;
        // SayWorld::sayHello()가 포함되며,
        // 부모 클래스인 Base 클래스의 메서드를 override
    }

    $o = new MyHelloWorld();
    $o->sayHello(); // Hello World!
    ```

## 예제

```php
trait Logger {
    public function log($message) {
        echo "Logging message: $message";
    }
}

trait FileHandler {
    public function saveToFile($filename, $data) {
        file_put_contents($filename, $data);
        echo "Data saved to $filename";
    }
}

class Application {
    use Logger, FileHandler;

    public function run() {
        $this->log("Application started.");
        $this->saveToFile("log.txt", "Application log data");
    }
}

$app = new Application();
$app->run();


class Base
{
    public function sayHello()
    {
        echo 'Hello ';
    }
}

trait SayWorld
{
    public function sayHello()
    {
        parent::sayHello();
        echo 'World!' . PHP_EOL;
    }
}

// 현재 클래스의 메서드가 Trait 메서드를 재정의
class MyHelloWorld extends Base
{
    use SayWorld;

    // SayWorld::sayHello()가 포함되며,
    // 부모 클래스인 Base 클래스의 메서드를 재정의
}

$o = new MyHelloWorld();
$o->sayHello(); // Hello World


class ClassBase {
    public function methodA() {
        echo "ClassBase Method A\n";
    }
}

trait TraitA {
    public function methodA() {
        echo "TraitA Method A\n";
    }
}

class ClassB extends ClassBase{
    use TraitA;

    public function methodA() {
        echo "ClassB Method A\n";
    }
}

$obj = new ClassB();
$obj->methodA(); // "ClassB Method A" 출력
```

## 각 언어에서의 Trait

### Rust

**Rust**에서의 Trait은 인터페이스와 유사한 개념입니다.
Trait은 특정 타입이 가져야 할 메서드 시그니처를 정의하며, 이를 통해 다형성을 구현할 수 있습니다.

인터페이스와 유사한 역할을 하며, 다형성을 지원합니다.
Rust에서는 Trait을 구현할 때 `impl` 키워드를 사용합니다.

- **인터페이스 역할**

    특정 타입이 구현해야 할 메서드를 정의합니다.

- **다형성**

    여러 타입이 동일한 Trait을 구현함으로써 다형성을 지원합니다.

```rust
trait Greet {
    fn greet(&self);
}

struct Person {
    name: String,
}

impl Greet for Person {
    fn greet(&self) {
        println!("Hello, my name is {}", self.name);
    }
}

fn main() {
    let person = Person { name: String::from("Alice") };
    person.greet();
}
```

### Haskell

**Haskell**에서의 타입 클래스는 Rust의 Trait과 유사한 개념입니다.
타입 클래스는 특정 타입이 수행할 수 있는 연산을 정의하는 인터페이스와 같은 구조를 가집니다.

타입 클래스는 함수형 프로그래밍에서 다형성을 다루는 주요 방법입니다.
타입 클래스는 특정 연산을 수행할 수 있는 타입들을 그룹화하여 범용적인 코드를 작성할 수 있게 합니다.

- **다형성**

    다양한 타입이 동일한 연산을 수행할 수 있게 합니다.

- **함수형 프로그래밍 지원**

    함수형 프로그래밍에서 다형성을 다루는 데 사용됩니다.

```haskell
class Eq a where
    (==) :: a -> a -> Bool

instance Eq Int where
    (==) x y = x Prelude.== y

instance Eq Bool where
    (==) True True = True
    (==) False False = True
    (==) _ _ = False
```
