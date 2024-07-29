# static

- [static](#static)
    - [`trait`의 static method 호출하기](#trait의-static-method-호출하기)
    - [정적 메서드 호출 시 생성자가 호출되는가?](#정적-메서드-호출-시-생성자가-호출되는가)
    - [`$this` 내에서 `static` 속성 또는 함수 사용](#this-내에서-static-속성-또는-함수-사용)
        - [`$this`와 `self`](#this와-self)
        - [PHP와 Java의 차이점](#php와-java의-차이점)
        - [PHP의 메모리 관리](#php의-메모리-관리)
        - [정적 멤버의 인스턴스 접근](#정적-멤버의-인스턴스-접근)
        - [언제 인스턴스 메서드를 쓰고 언제 정적 메서드를 사용해야 할까](#언제-인스턴스-메서드를-쓰고-언제-정적-메서드를-사용해야-할까)
            - [정적 메서드 사용의 경우](#정적-메서드-사용의-경우)
            - [인스턴스 메서드 사용의 경우](#인스턴스-메서드-사용의-경우)
            - [PHP 내부 구현과 메모리 관리](#php-내부-구현과-메모리-관리)
            - [베스트 프랙티스](#베스트-프랙티스)
    - [일반 메서드에서 정적 메서드 사용](#일반-메서드에서-정적-메서드-사용)
        - [정적 메서드를 사용하는 것에 대한 의견](#정적-메서드를-사용하는-것에-대한-의견)
        - [PHP 내부 구현 및 성능 관점](#php-내부-구현-및-성능-관점)

## `trait`의 static method 호출하기

```php
trait AnyTrait {
    public static function traiMethod($params) {
        return "something";
    }
}

class SomeClass {
    use AnyTrait;
    
    public static function someMethod($params) {
        $resultOfTrait = (new SomeClass)->traiMethod($params);
        print_r($resultOfTrait);
    }
}

SomeClass::someMethod("test");
// 결과
// something
```

## 정적 메서드 호출 시 생성자가 호출되는가?

```php
class MyClass {
    private static $client;

    public function __construct() {
        if (self::$client === null) {
            self::$client = new Client();
        }
    }

    public static function test($path)
    {
        self::$client->post($path, []);
    }
}

MyClass::test('/path/to/endpoint');
```

- PHP에서 생성자는 클래스의 인스턴스가 생성될 때 자동으로 호출되는 특별한 메서드
- 정적 메서드는 클래스의 **인스턴스를 생성하지 않고도 호출**할 수 있기 때문에, `__construct` 생성자는 이 경우 호출되지 않는다
- 따라서, `MyClass::test('/path/to/endpoint');`와 같이 정적 메서드를 호출하는 경우, `self::$client`가 초기화되지 않았다면 `null` 상태일 것이며, 이로 인해 오류가 발생할 가능성이 있다
- 이를 해결하려면,
    1. 클라이언트를 초기화하는 별도의 정적 메서드를 만들거나
    2. 클라이언트 인스턴스를 매번 정적 메서드 내에서 직접 생성하는 방법이 있다

```php
class MyClass {
    private static $client;

    // 클라이언트를 초기화하는 별도의 정적 메서드 사용 예시
    private static function initializeClient() {
        if (self::$client === null) {
            self::$client = new Client();
        }
    }

    public static function test($path) {
        self::initializeClient();
        self::$client->post($path, []);
    }
}
```

## `$this` 내에서 `static` 속성 또는 함수 사용

### `$this`와 `self`

- `$this` 키워드
    - 인스턴스(객체) 메소드 내에서 사용되며, **현재 인스턴스**를 가리킨다
    - 이를 통해 인스턴스 변수에 접근하거나 인스턴스 메소드를 호출할 수 있다
- `self` 키워드
    - 클래스 내에서 사용되며, **현재 클래스 자체**를 가리킨다
    - `self`는 주로 정적(static) 메소드나 속성에 접근할 때 사용된다

### PHP와 Java의 차이점

- PHP
    - PHP는 상대적으로 느슨한 타입 체크와 개방적인 접근 방식을 갖는다
    - PHP에서는 정적 메소드나 속성을 인스턴스 컨텍스트(`$this`) 내에서도 접근할 수 있다.
- Java
    - Java는 보다 엄격한 언어로, 정적 메소드나 속성은 클래스 레벨에서만 접근 가능하다
    - 인스턴스 메소드 내에서 직접적인 정적 멤버 접근은 허용되지 않는다

### PHP의 메모리 관리

- PHP의 메모리 관리는 주로 내부적인 참조 카운팅과 가비지 컬렉션에 의해 이루어진다.
- 객체와 클래스의 메모리 관리는
    - 인스턴스 객체
        - 객체가 생성될 때, PHP는 메모리에 해당 객체의 인스턴스를 할당한다.
        - `$this`는 이 메모리 주소를 가리키는 참조
    - 정적 멤버
        - 정적 메소드와 속성은 클래스 레벨에서 관리된다
        - 즉, 이들은 클래스의 모든 인스턴스 간에 공유되며, 클래스가 로드될 때 메모리에 할당된다

### 정적 멤버의 인스턴스 접근

PHP에서 `$this` 내에서 `self::$staticFunction` 같은 방식으로 정적 멤버에 접근할 수 있는 것은 PHP의 유연한 언어 특성 때문이다. 이는 **객체의 컨텍스트 내에서도 클래스 레벨의 멤버에 접근**할 수 있음을 의미한다. 하지만, 이런 접근 방식은 객체지향 프로그래밍의 일반적인 원칙과는 다소 거리가 있으며, 때로는 혼란을 야기할 수도 있다.

### 언제 인스턴스 메서드를 쓰고 언제 정적 메서드를 사용해야 할까

#### 정적 메서드 사용의 경우

1. 상태 독립성: 만약 메서드가 객체의 상태(즉, 인스턴스 변수)에 의존하지 않고 동작한다면, 그 메서드는 정적으로 선언될 수 있다
2. 유틸리티 함수: 객체의 인스턴스 없이도 의미가 있는, 일반적인 유틸리티 함수(예: 수학 함수, 변환 함수)는 정적 메서드로 구현될 수 있다
3. 객체 생성 없이 접근 필요: 객체를 생성하지 않고도 접근해야 하는 메서드는 정적 메서드로 선언된다

#### 인스턴스 메서드 사용의 경우

1. 객체의 상태에 의존: 메서드가 객체의 특정 상태(인스턴스 변수)에 의존하는 경우, 이는 일반 인스턴스 메서드로 구현
2. 오버라이드가 필요한 경우: 상속과 다형성을 활용하여 메서드를 오버라이드할 필요가 있는 경우, 인스턴스 메서드를 사용

#### PHP 내부 구현과 메모리 관리

- 메모리 관리
    - PHP는 내부적으로 참조 카운팅과 가비지 컬렉션을 사용하여 메모리를 관리
    - 인스턴스 메서드는 객체당 별도의 메모리 공간을 사용하지만, 정적 메서드는 클래스당 한 번만 메모리에 할당된다
- Zend Engine
    - PHP의 핵심 구성 요소인 Zend Engine은 PHP 스크립트를 중간 코드로 컴파일하고, 이를 실행시킨다
    - 정적 메서드와 인스턴스 메서드는 Zend Engine에 의해 다르게 처리된다
- FPM (FastCGI Process Manager)
    - 웹 서버와 PHP 간의 인터페이스를 관리한다
    - FPM 자체는 메서드 타입(정적 vs 인스턴스)에 대한 선택에 직접적인 영향을 주지 않는다

#### 베스트 프랙티스

1. 단일 책임 원칙: 클래스와 메서드는 하나의 책임만 가져야 하며, 이 원칙에 따라 메서드의 역할을 결정한다.
2. 캡슐화 유지: 객체의 상태와 행동을 캡슐화하여 외부에서 직접 접근하지 못하도록 한다.
3. 재사용성과 유지보수: 코드의 재사용성과 유지보수를 고려하여, 공통 기능은 정적 메서드로 만들고, 객체의 특정 상태와 관련된 기능은 인스턴스 메서드로 구현한다

```php
class MyClass {
    public function test(){
        self::innerStaticFunction();
    }

    private static function innerStaticFunction() {}
}
```

예시로 든 `MyClass`에서 `innerStaticFunction`이 인스턴스의 상태에 의존하지 않는다면, 이를 정적 메서드로 선언하는 것이 타당할 수 있다. 하지만, 이러한 결정은 전체적인 애플리케이션의 설계와 컨텍스트에 따라 달라질 수 있다

## 일반 메서드에서 정적 메서드 사용

```php
<?php
App::uses('Component', 'Controller');

class GrpcClientComponent extends Component {
    private $grpcClient;

    public function __construct(ComponentCollection $collection, $settings = []) {
        parent::__construct($collection, $settings);

        if (!empty($settings['grpcClient'])) {
            $this->grpcClient = $settings['grpcClient'];
        } else {
            // 실제 gRPC 클라이언트를 초기화합니다.
            $this->grpcClient = new ActualGrpcClient(self::getGRPCHost(), [/*gRPC option*/]);
        }
    }

    // gRPC 클라이언트를 사용하는 메서드를 추가합니다.
    /** @return string */
    private static function getGRPCHost()
    {
        return $_ENV['GRPC_HOST'];
    }
}
```

1. 정적 메서드의 사용

    정적 메서드는 클래스의 인스턴스와 무관하게 호출할 수 있습니다.
    따라서 *상태를 가지지 않는* 유틸리티 함수나 공통 기능을 제공할 때 유용합니다.

    `getGRPCHost` 메서드는 환경 변수를 반환하는 단순한 함수이므로, 정적 메서드로 선언하는 것이 적절할 수 있습니다.

2. 일반 메서드에서 정적 메서드 호출

    일반 메서드에서 정적 메서드를 호출하는 것은 일반적인 패턴입니다.
    이는 *클래스의 인스턴스 메서드가 클래스 수준의 기능을 필요로 할 때 유용*합니다.

정적 메서드를 사용하는 것이 적절한 경우도 있지만, 환경 변수를 가져오는 단순한 작업이라면 정적 메서드로 선언하지 않고, 인스턴스 메서드로 선언하는 것도 고려해볼 수 있습니다. 이는 코드의 일관성을 유지하는 데 도움이 됩니다.

```php
// gRPC 클라이언트를 사용하는 메서드를 추가합니다.
/** @return string */
private function getGRPCHost() {
    return $_ENV['GRPC_HOST'];
}
```

### 정적 메서드를 사용하는 것에 대한 의견

1. 정적 메서드의 장점

   상태를 가지지 않는 유틸리티 함수나 공통 기능을 제공할 때 유용합니다.

   클래스의 인스턴스와 무관하게 호출할 수 있어, 코드의 재사용성을 높일 수 있습니다.

2. 정적 메서드의 단점

   테스트하기 어려울 수 있습니다.
   특히, 정적 메서드는 모킹(mocking)이 어렵기 때문에, 테스트 코드 작성 시 불편함을 초래할 수 있습니다.

   객체 지향 프로그래밍의 원칙을 위반할 수 있습니다.
   정적 메서드는 클래스의 상태를 가지지 않기 때문에, 객체 지향 설계의 장점을 활용하지 못할 수 있습니다.

### PHP 내부 구현 및 성능 관점

PHP는 인터프리터 언어로, 코드를 기계어로 직접 컴파일하지 않습니다.
대신, PHP 코드는 바이트코드로 변환되어 Zend Engine에 의해 실행됩니다.
정적 메서드 호출과 인스턴스 메서드 호출의 성능 차이는 미미하지만, 다음과 같은 점을 고려할 수 있습니다:

1. 정적 메서드 호출

   클래스의 인스턴스가 필요 없으므로, 메모리 사용이 약간 더 효율적일 수 있습니다.
   호출 시점에 클래스의 상태를 고려하지 않기 때문에, 약간 더 빠를 수 있습니다.

2. 인스턴스 메서드 호출

   클래스의 인스턴스가 필요하며, 호출 시점에 클래스의 상태를 고려합니다.
   객체 지향 설계의 장점을 활용할 수 있습니다.

정적 메서드를 사용하는 것은 특정 상황에서 유용할 수 있지만, 객체 지향 설계의 장점을 최대한 활용하기 위해서는 인스턴스 메서드를 사용하는 것이 더 나을 수 있습니다.
특히, 테스트 가능성과 코드의 일관성을 고려할 때, 인스턴스 메서드를 사용하는 것이 좋습니다.
성능 차이는 미미하므로, 코드의 가독성과 유지보수성을 우선시하는 것이 좋습니다.
