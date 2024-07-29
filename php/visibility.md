# Visibility

- [Visibility](#visibility)
    - [`protected`](#protected)
    - [visibility와 상속](#visibility와-상속)
        - [왜 `new ProtectedClass()`를 통해 생성된 인스턴스의 `protected` 속성에도 접근할 수 있을까요?](#왜-new-protectedclass를-통해-생성된-인스턴스의-protected-속성에도-접근할-수-있을까요)
        - [코드 분석](#코드-분석)
        - [왜 둘 다 접근 가능한가?](#왜-둘-다-접근-가능한가)
        - [주의할 점](#주의할-점)
        - [결론](#결론)
    - [kotlin과 PHP의 인스턴스화 차이](#kotlin과-php의-인스턴스화-차이)
        - [Kotlin과 PHP의 인스턴스화 차이](#kotlin과-php의-인스턴스화-차이-1)
            - [1. Kotlin에서의 인스턴스화와 접근 제어](#1-kotlin에서의-인스턴스화와-접근-제어)
            - [예제 (Kotlin)](#예제-kotlin)
            - [2. PHP에서의 인스턴스화와 접근 제어](#2-php에서의-인스턴스화와-접근-제어)
            - [예제 (PHP)](#예제-php)
        - [차이의 기저 원인](#차이의-기저-원인)
    - [동적으로 이미 생성된 인스턴스의 protected 속성에 접근하기](#동적으로-이미-생성된-인스턴스의-protected-속성에-접근하기)
        - [코드 예시](#코드-예시)
        - [실무 적용](#실무-적용)
        - [최신 트렌드](#최신-트렌드)

## `protected`

제공된 코드 예제를 다시 살펴보겠습니다. 여기서는 `PublicClass`가 `ProtectedClass`를 상속받고 있으며, `publicFunction` 메소드 내에서 두 가지 방식으로 `protectedProperty`에 접근하고 있습니다:

1. `$this->protectedProperty`를 통한 접근
2. `new ProtectedClass()`를 통해 생성된 인스턴스를 통한 접근

이 코드의 출력 결과는 다음과 같습니다:

```plaintext
Array
(
    [1] => protected
    [2] => protected
)
```

이는 `PublicClass`가 `ProtectedClass`를 상속받았기 때문에, `PublicClass` 내부에서는 `ProtectedClass`의 `protected` 속성에 접근할 수 있음을 보여줍니다. 그러나 여기서 중요한 점은 `new ProtectedClass()`를 통해 생성된 인스턴스에 대한 접근입니다.

## visibility와 상속

java나 kotlin 등과 다르게, PHP는 extends를 하면 인스턴스화 된 자신(`$this->부모속성`) 뿐만 아니라, 새로운 인스턴스(`new ProtectedClass()`)의 `protected` 필드(`$protected->protectedProperty`)에도 접근이 가능하다.

### 왜 `new ProtectedClass()`를 통해 생성된 인스턴스의 `protected` 속성에도 접근할 수 있을까요?

이 질문에 대한 답은 PHP의 객체 지향 프로그래밍 모델과 접근 제한자의 동작 방식에 있습니다. `protected` 속성은 선언된 클래스와 그 클래스를 상속받은 자식 클래스에서만 접근 가능합니다. 그러나 여기서 주목해야 할 점은 `PublicClass`가 `ProtectedClass`를 상속받았기 때문에, `PublicClass` 내부에서는 `ProtectedClass`의 인스턴스를 생성하고 그 인스턴스의 `protected` 속성에 접근할 수 있다는 것입니다.

이는 PHP에서 클래스의 상속과 접근 제한자가 어떻게 동작하는지에 대한 중요한 개념입니다. 즉, 자식 클래스(`PublicClass`)는 부모 클래스(`ProtectedClass`)의 `protected` 속성에 접근할 수 있으며, 이는 부모 클래스의 인스턴스를 직접 생성하더라도 마찬가지입니다.

### 코드 분석

```php
class PublicClass extends ProtectedClass {
    public function publicFunction() {
        // new ProtectedClass()를 통해 생성된 객체도 
        // PublicClass 내에서는 protected 속성에 접근할 수 있습니다. 
        // 이는 PublicClass가 ProtectedClass의 상속 구조 안에 있기 때문입니다.
        $protected = new ProtectedClass();
        print_r([
            $protected->protectedProperty,
            $this->protectedProperty,
        ]);
    }
}
```

위 코드에서 `$protected = new ProtectedClass();`와 `$this->protectedProperty` 모두 `protectedProperty`에 접근합니다. 첫 번째 경우는 `PublicClass`가 `ProtectedClass`를 상속받았기 때문에 가능하며, 두 번째 경우는 `$this`가 현재 객체(`PublicClass`의 인스턴스)를 참조하기 때문에 가능합니다.

결론적으로, 이 코드 예제에서 보여주는 바와 같이, PHP에서는 자식 클래스가 부모 클래스의 `protected` 속성에 접근할 수 있으며, 이는 부모 클래스의 인스턴스를 직접 생성하더라도 마찬가지입니다. 이는 PHP의 객체 지향 프로그래밍 모델과 접근 제한자의 설계 철학에 따른 것입니다.

Citations:

### 왜 둘 다 접근 가능한가?

이 현상은 PHP의 독특한 특성 때문입니다. PHP에서는 같은 클래스의 다른 인스턴스의 protected 멤버에 접근할 수 있습니다. 이는 다음과 같은 원리로 작동합니다:

1. **클래스 범위 접근**: PHP에서 `protected` 멤버는 해당 클래스와 그 자식 클래스에서 접근 가능합니다. 그러나 PHP는 이를 "클래스 범위"로 해석합니다. 즉, 같은 클래스의 다른 인스턴스에서도 `protected` 멤버에 접근할 수 있습니다.

2. **상속 관계**: `PublicClass`가 `ProtectedClass`를 상속받았기 때문에, `PublicClass`는 `ProtectedClass`의 `protected` 멤버에 접근할 수 있는 권한을 가집니다.

3. **객체 지향 설계의 유연성**: 이러한 접근 방식은 PHP의 객체 지향 모델에서 유연성을 제공합니다. 같은 클래스의 다른 인스턴스와 상호작용할 때 유용할 수 있습니다.

### 주의할 점

이 동작은 PHP의 특성이며, 다른 언어들과는 다를 수 있습니다. 예를 들어, Java나 C++에서는 이러한 접근이 허용되지 않습니다. 따라서 이 특성을 사용할 때는 주의가 필요합니다:

1. **캡슐화 약화**: 이러한 접근 방식은 캡슐화를 약화시킬 수 있습니다. `protected` 멤버가 예상보다 더 넓은 범위에서 접근 가능해집니다.

2. **코드 이식성**: 다른 언어로 코드를 포팅할 때 문제가 될 수 있습니다.

3. **의도하지 않은 접근**: 개발자가 의도하지 않은 방식으로 `protected` 멤버에 접근할 수 있어, 버그의 원인이 될 수 있습니다.

### 결론

PHP에서 이러한 동작은 의도적으로 설계된 것입니다. 이는 PHP의 객체 모델의 특성을 반영하며, 개발자에게 더 많은 유연성을 제공합니다. 하지만 이 특성을 사용할 때는 코드의 명확성과 유지보수성을 고려해야 합니다. 또한, 다른 언어와의 차이점을 인식하고 있어야 합니다.

## kotlin과 PHP의 인스턴스화 차이

Kotlin과 PHP의 인스턴스화가 실제로 어떻게 다르고, 이러한 차이가 왜 발생하는지 구체적으로 설명하겠습니다.

### Kotlin과 PHP의 인스턴스화 차이

#### 1. Kotlin에서의 인스턴스화와 접근 제어

Kotlin에서는 객체 지향 프로그래밍의 원칙에 따라 접근 제어를 엄격하게 적용합니다. `protected` 속성은 해당 클래스와 그 하위 클래스에서만 접근할 수 있으며, 하위 클래스가 상위 클래스의 인스턴스를 통해 `protected` 속성에 접근하는 것은 불가능합니다. 이는 다음과 같은 이유 때문입니다:

- **캡슐화 원칙**: 객체 지향 프로그래밍에서 캡슐화는 데이터 은닉을 통해 객체의 내부 상태를 보호합니다. 상위 클래스의 `protected` 속성은 하위 클래스 내에서만 접근할 수 있으며, 다른 인스턴스를 통해 접근하는 것은 허용되지 않습니다.
- **런타임 안전성**: 각 객체는 자신의 상태를 관리해야 하며, 다른 객체의 내부 상태에 직접 접근하는 것은 그 객체의 무결성을 해칠 수 있습니다.

#### 예제 (Kotlin)

```kotlin
open class ProtectedClass {
    protected val protectedProperty = "protected"
}

class PublicClass : ProtectedClass() {
    fun publicFunction() {
        // println(ProtectedClass().protectedProperty) // 컴파일 오류 발생
        println(this.protectedProperty) // 허용
    }
}

fun main() {
    val publicClass = PublicClass()
    publicClass.publicFunction()
}
```

위 코드에서 `ProtectedClass`의 새로운 인스턴스를 통해 `protected` 속성에 접근하려고 하면 컴파일 오류가 발생합니다.

#### 2. PHP에서의 인스턴스화와 접근 제어

PHP에서는 `protected` 속성이 선언된 클래스와 그 하위 클래스에서 접근 가능하다는 점은 동일하지만, PHP는 클래스 내부에서 생성된 새로운 인스턴스를 통해서도 `protected` 속성에 접근할 수 있도록 허용합니다. 이는 다음과 같은 이유 때문입니다:

- **PHP의 유연성**: PHP는 동적 타이핑 언어로, 유연성과 접근성을 중시합니다. 이러한 설계 철학은 객체 지향 원칙을 약간 덜 엄격하게 적용하는 경향이 있습니다.
- **클래스 범위 접근**: PHP는 `protected` 속성을 "클래스 범위"로 해석하여, 같은 클래스의 다른 인스턴스에서도 `protected` 속성에 접근할 수 있도록 허용합니다.

#### 예제 (PHP)

```php
<?php

class ProtectedClass
{
    protected $protectedProperty = 'protected';
}

class PublicClass extends ProtectedClass
{
    public function publicFunction()
    {
        $protected = new ProtectedClass();
        echo $protected->protectedProperty; // 허용
        echo $this->protectedProperty; // 허용
    }
}

$publicClass = new PublicClass();
$publicClass->publicFunction();
```

위 코드에서 `ProtectedClass`의 새로운 인스턴스를 통해 `protected` 속성에 접근하는 것이 허용됩니다.

### 차이의 기저 원인

- **캡슐화와 무결성**: Kotlin은 캡슐화와 객체 무결성을 강하게 유지하는 반면, PHP는 더 유연한 접근 방식을 취합니다. 이러한 차이는 두 언어의 설계 철학과 사용 사례에 따라 다릅니다.
- **컴파일 타임 vs 런타임 검사**: Kotlin은 정적 타입 언어로 컴파일 타임에 많은 검사를 수행하여 오류를 방지합니다. PHP는 동적 타입 언어로 런타임에 많은 검사를 수행하며, 유연한 접근 방식을 허용합니다.
- **객체 지향 원칙의 적용**: Kotlin은 엄격하게 객체 지향 원칙을 따르는 반면, PHP는 현실적인 유연성을 중시합니다.

이 차이로 인해, PHP에서는 같은 클래스 내에서 생성된 새로운 인스턴스를 통해 `protected` 속성에 접근할 수 있지만, Kotlin에서는 이러한 접근이 제한됩니다.

## 동적으로 이미 생성된 인스턴스의 protected 속성에 접근하기

PHP의 객체 지향 프로그래밍 기능은 PHP 4에서 처음 도입되었고, PHP 5에서 크게 개선되었습니다.
접근 제어자(`public`, `protected`, `private`)는 PHP 5에서 추가되어 캡슐화를 강화했습니다.
`protected` 속성은 해당 클래스와 자식 클래스 내에서만 접근 가능하도록 설계되었습니다.

PHP에서 `protected` 속성에 직접적으로 외부에서 접근하는 것은 불가능합니다.
이는 객체 지향 프로그래밍의 캡슐화 원칙을 지키기 위함입니다.
그러나 몇 가지 우회적인 방법이 있습니다:

1. Reflection API 사용: PHP의 Reflection API를 사용하면 런타임에 클래스의 내부 구조에 접근할 수 있습니다.

2. 매직 메서드 활용: `__get()`,`__set()` 등의 매직 메서드를 사용하여 `protected` 속성에 대한 접근을 제어할 수 있습니다.

3. 클로저 사용: PHP 5.3 이상에서는 클로저를 사용하여 `protected` 멤버에 접근할 수 있습니다.

### 코드 예시

1. Reflection API 사용:

    ```php
    class MyClass {
        protected $protectedProperty = 'Protected Value';
    }

    $obj = new MyClass();
    $reflection = new ReflectionClass($obj);
    $property = $reflection->getProperty('protectedProperty');
    $property->setAccessible(true);
    echo $property->getValue($obj);  // 출력: Protected Value
    ```

2. 매직 메서드 활용:

    ```php
    class MyClass {
        protected $protectedProperty = 'Protected Value';
        
        public function __get($name) {
            if (property_exists($this, $name)) {
                return $this->$name;
            }
        }
    }

    $obj = new MyClass();
    echo $obj->protectedProperty;  // 출력: Protected Value
    ```

3. 클로저 사용:

    ```php
    class MyClass {
        protected $protectedProperty = 'Protected Value';
        
        public function getAccessor() {
            return function() {
                return $this->protectedProperty;
            };
        }
    }

    $obj = new MyClass();
    $accessor = $obj->getAccessor();
    echo $accessor->call($obj);  // 출력: Protected Value
    ```

### 실무 적용

실제 개발에서는 protected 속성에 직접 접근하는 것보다 getter/setter 메서드를 통해 접근하는 것이 좋습니다.
이는 코드의 유지보수성과 안정성을 높이는 데 도움이 됩니다.
Reflection API나 매직 메서드를 사용한 방법은 테스트나 디버깅 목적으로 제한적으로 사용하는 것이 좋습니다.

### 최신 트렌드

최근 PHP 개발에서는 캡슐화와 정보 은닉을 더욱 강조하는 추세입니다.
PHP 8.0에서 도입된 Constructor Property Promotion과 같은 기능은 속성 선언과 초기화를 더 간결하게 만들어주며, 이는 `protected` 속성의 사용을 더욱 편리하게 만듭니다.

`protected` 속성에 대한 접근을 제어하는 것은 객체 지향 설계의 중요한 부분입니다.
직접적인 접근보다는 잘 설계된 `public` 인터페이스를 통해 객체와 상호작용하는 것이 바람직합니다.
