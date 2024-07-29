# Function - Closure

- [Function - Closure](#function---closure)
    - [클로저?](#클로저)
    - [프로그래밍 계층 구조에서 클로저의 의미](#프로그래밍-계층-구조에서-클로저의-의미)
    - [클로저의 동작 원리](#클로저의-동작-원리)
        - [PHP에서 클로저의 동작 원리](#php에서-클로저의-동작-원리)
        - [클로저의 데이터 캡처 기능 예제](#클로저의-데이터-캡처-기능-예제)
    - [클로저, 로컬 변수, 그리고 클래스 속성](#클로저-로컬-변수-그리고-클래스-속성)
    - [코드 예제](#코드-예제)
        - [PHP](#php)
            - [테스트 코드](#테스트-코드)
    - [클로저는 함수와 그 함수가 선언된 환경을 함께 캡처한다](#클로저는-함수와-그-함수가-선언된-환경을-함께-캡처한다)
        - [클로저 내부의 데이터만 출력하는 방법](#클로저-내부의-데이터만-출력하는-방법)
        - [코드 예제](#코드-예제-1)
        - [클로저 출력 결과](#클로저-출력-결과)
    - [고급](#고급)
        - [returnValue(클로저 호출) 은 원하는대로 동작하지 않을 수 있다](#returnvalue클로저-호출-은-원하는대로-동작하지-않을-수-있다)

## 클로저?

클로저는 함수와 그 함수가 선언된 환경을 캡처하여 저장하는 객체입니다.

클로저(Closure)는 *함수가 선언된 렉시컬 범위(Lexical Scope) 외부에서 선언된 변수를 참조할 수 있는 함수*입니다.
- 함수가 선언된 렉시컬 범위(Lexical Scope)가 있고
- 그 범위 외부에서 선언된 변수를
- 참조할 수 있는 함수

클로저는 *함수와 그 함수가 선언될 때의 환경을 기억*하여, *함수 외부에서도 그 환경을 유지하며 동작*할 수 있습니다.
이는 함수가 자신이 정의된 스코프 밖에서도 사용될 수 있도록 합니다.

클로저의 주요 특징은 다음과 같습니다.
- 외부 변수에 접근할 수 있음.
- 변수를 캡처하고, 이 값을 유지함.
- 일급 객체로서 다른 함수에 인자로 전달되거나 반환값으로 사용될 수 있음.

PHP 5.3부터 클로저를 지원하며, `use` 키워드를 사용하여 외부 변수를 캡처할 수 있습니다.
클로저는 `Closure` 클래스의 인스턴스로 취급되며, 이를 통해 함수 객체를 생성하고, 이를 나중에 호출할 수 있습니다.
따라서 클로저를 출력하면 캡처된 변수들과 클로저가 정의된 객체의 정보를 모두 포함하여 복잡한 구조를 나타냅니다.

클로저를 실제로 사용하여 데이터를 출력하려면 클로저 객체를 호출하여 반환값을 얻는 방식으로 접근해야 합니다.

## 프로그래밍 계층 구조에서 클로저의 의미

1. **프로그래밍 패러다임**:

    클로저는 주로 함수형 프로그래밍 언어에서 중요하게 다루어집니다.
    하지만 객체지향 프로그래밍에서도 사용될 수 있습니다.

2. **프로그래밍 언어 구성 요소**:

    함수는 프로그래밍 언어의 기본 구성 요소 중 하나입니다.
    클로저는 일급 함수의 한 종류로, *함수가 선언된 환경을 기억하는 능력*을 갖습니다.
    [일급 함수](./function.md#일급-객체first-class-citizen의-특성)는 함수를 값처럼 다룰 수 있는 특성을 의미합니다.

3. **언어 기능 및 개념**:

    스코프는 변수의 유효 범위를 의미합니다.
    *렉시컬 스코프*는 *변수가 정의된 위치에 따라 그 유효 범위가 결정되는 것*을 의미합니다.
    클로저는 이러한 렉시컬 스코프를 활용하여 외부 변수에 접근하고 이를 유지하는 기능을 제공합니다.

## 클로저의 동작 원리

클로저(Closure)의 동작 원리와 내부 구조를 이해하기 위해서는 클로저가 *어떻게 함수와 그 함수가 정의된 환경을 캡처하고 유지*하는지에 대해 알아야 합니다.

클로저는 렉시컬 스코프(lexical scope)를 사용하여 외부 변수를 캡처합니다.
*렉시컬 스코프는 함수가 정의된 시점의 스코프를 참조*합니다.
이 스코프 내의 변수들은 클로저가 정의될 때 캡처되어 클로저 내부에서 사용할 수 있게 됩니다.
이를 통해 클로저는 정의될 때의 환경(변수들)을 기억하고, 나중에 호출될 때 그 환경에 접근할 수 있습니다.

- **렉시컬 스코프**:
    - 클로저는 자신이 정의된 시점의 스코프(렉시컬 스코프)를 참조합니다.
    - 이 스코프에 있는 변수들은 클로저 내부에서 사용할 수 있도록 캡처됩니다.

- **외부 변수 접근**:
    - `captureData` 메서드 내에서 정의된 클로저는 `$data` 변수를 캡처합니다.
    - 이 캡처된 변수는 클로저 내부에서 사용될 수 있으며, 클로저가 호출될 때 그 값을 참조합니다.

### PHP에서 클로저의 동작 원리

PHP에서 클로저는 *함수와 그 함수가 정의된 환경을 함께 저장*하는 객체입니다.
이를 통해 클로저는 *함수 외부의 변수를 참조*할 수 있습니다.

`use`를 사용하는 클로저(Closure)는 *클로저가 정의된 시점의 변수들을 클로저 내부에서 사용할 수 있도록 캡처(capture)하는 기능*을 제공합니다.
이를 통해 클로저는
- 클로저가 정의된 외부 스코프의 변수를 기억할 수 있습니다.
- 클로저가 나중에 호출될 때에도 이 변수들에 접근할 수 있습니다.

```php
/**
 * @param mixed $data
 * @return Closure
 */
private function captureData($data)
{
    //                `use ($data)` 구문을 통해 `$data` 변수를 캡처합니다. 
    return function () use ($data) {
        // 클로저는 `$data` 변수를 캡처하여 내부 상태로 유지합니다.
        return $data;
    };
}
```

```php
/**
 * @param $data
 * @return Closure
 */
private function captureData($data)
{
    // `use`를 사용하지 않기 때문에 클로저는 외부 스코프의 `$data` 변수에 접근할 수 없습니다. 
    return function () {
        // 이 클로저는 단순히 `null`을 반환합니다.
        return null;
    };
}
```

`use` 키워드를 사용하면 클로저는 *정의된 시점의 외부 변수들을 캡처하여 클로저 내부에서 사용*할 수 있습니다.
이 과정에서 다음과 같은 일이 일어납니다:

1. **변수 캡처**:

    클로저가 정의될 때, `use` 키워드를 사용하여 외부 변수의 값을 캡처합니다.
    이 캡처된 값은 클로저 내부에 저장됩니다.

2. **변수 유지**:

    클로저는 캡처된 변수를 자신의 상태로 유지합니다.
    클로저가 나중에 호출될 때, 캡처된 변수를 참조할 수 있습니다.

3. **변수 접근**:

    클로저 내부에서 캡처된 변수에 접근하여 값을 사용할 수 있습니다.

```php
// 클래스 범위:
// - 클래스 MyClass는 정의된 모든 멤버 변수와 메서드를 포함합니다.
// - 클래스 내부에서 정의된 모든 메서드는 클래스 스코프에 접근할 수 있습니다.
class MyClass {

    // 메서드 `captureData` 범위:
    // - 메서드 `captureData`는 `$data` 변수를 인자로 받습니다.
    // - `$data` 변수는 메서드 내부에서만 유효합니다.
    // - 메서드 내부에서 정의된 클로저는 `$data` 변수를 캡처할 수 있습니다.
    private function captureData($data) {

        // 클로저의 정의 위치:
        // - 클로저는 `captureData` 메서드 내부에서 정의됩니다.
        // - 따라서, 클로저는 `captureData` 메서드의 스코프를 참조합니다.
        // - 이 경우, 클로저는 `$data` 변수를 캡처합니다.
        return function () use ($data) { // 클로저가 정의되는 시점.
                                         // 이 스코프에서 접근 가능한 변수는 `$data`입니다. 
            return $data;
        };
    }

    public function test() {
        // 여러 클로저를 정의하고 이를 저장합니다.
        $closures = [];

        for ($i = 1; $i <= 3; $i++) {
            // 반복문 내에서의 클로저 정의:
            // - `$i` 값은 반복문을 통해 1에서 3까지 변화합니다.
            // - 각 반복마다 `captureData($i)`가 호출되어 `$data` 값이 1, 2, 3으로 설정됩니다.
            // - `captureData($i)`가 호출될 때마다 새로운 클로저가 생성되어 `$closures` 배열에 저장됩니다.
            // - 각 클로저는 그 정의된 시점의 `$data` 값을 캡처합니다.
            $closures[] = $this->captureData($i);
        }

        // 클로저를 호출하여 캡처된 값을 출력합니다.
        // - 각 클로저를 호출하면, 클로저는 정의된 시점에 캡처한 `$data` 값을 반환합니다.
        // - 따라서, 첫 번째 클로저는 1을 반환하고, 두 번째 클로저는 2를 반환하며, 세 번째 클로저는 3을 반환합니다.
        foreach ($closures as $closure) {
            echo $closure() . "\n"; // 호출 시점에서 캡처된 `$data` 변수는 클로저 내부에서 사용됩니다.
        }
    }
}

$myClass = new MyClass();
$myClass->test();
```

### 클로저의 데이터 캡처 기능 예제

이 예제는 클로저가 외부 변수의 현재 값을 캡처하고, 이후에 그 값을 어떻게 유지하는지를 보여줍니다.

```php
class MyClass {
    private function captureData($data) {
        return function () use ($data) {
            return $data;
        };
    }

    public function test() {
        // 여러 클로저를 정의하고 이를 저장합니다.
        $closures = [];

        for ($i = 1; $i <= 3; $i++) {
            $closures[] = $this->captureData($i);
        }

        // 클로저를 호출하여 캡처된 값을 출력합니다.
        foreach ($closures as $closure) {
            echo $closure() . "\n";
        }
    }
}

$myClass = new MyClass();
$myClass->test();
// 출력: 
// 1
// 2
// 3
```

1. **클로저 정의**:

    클로저는 함수가 선언된 시점의 환경(렉시컬 스코프)을 캡처하여 저장합니다.
    이는 함수가 나중에 호출될 때도 그 환경을 기억하고 사용할 수 있게 합니다.

    - `captureData` 메서드는 `$data` 변수를 캡처하는 클로저를 생성합니다.
    - `use ($data)`를 통해 `$data` 변수를 클로저 내부로 캡처합니다.

2. **변수 캡처**

   클로저는 `use` 키워드를 사용하여 함수 외부의 변수를 캡처합니다.
   이렇게 캡처된 변수는 클로저 내부에서 사용될 수 있습니다.

   - `test` 메서드에서는 반복문을 통해 `$i` 값을 `captureData` 메서드에 전달하여 클로저를 생성합니다.

3. **클로저 저장**:

   - 생성된 클로저들은 `$closures` 배열에 저장됩니다.
   - 여기서 중요한 점은, 각 클로저가 생성될 때의 `$i` 값을 캡처한다는 것입니다.
        - 각 클로저는 생성될 때의 `$i` 값을 캡처하고, 이후 호출될 때 그 값을 반환합니다.
        - 첫 번째 클로저는 `$i = 1`일 때 생성되어 1을 캡처합니다.
        - 두 번째 클로저는 `$i = 2`일 때 생성되어 2를 캡처합니다.
        - 세 번째 클로저는 `$i = 3`일 때 생성되어 3을 캡처합니다.

4. **클로저 호출**:

    PHP에서 클로저는 `Closure` 클래스의 인스턴스입니다.
    클로저 객체는 캡처된 변수를 `static` 속성으로 저장하며, 클로저가 정의된 객체를 `this` 속성으로 저장합니다.

   - 반복문을 통해 `$closures` 배열에 저장된 각 클로저를 호출합니다.
   - 각 클로저는 정의될 때 캡처한 값을 반환합니다.

5. **캡처된 변수 사용**:
   - `test` 메서드에서 클로저를 호출하면, 캡처된 `$data` 변수가 반환됩니다.
   - 이 예제에서는 각 클로저가 호출될 때 캡처된 값인 1, 2, 3이 출력됩니다.

```php
class MyComponent
{
    private $dataCallback;

    public function myFunction1($data)
    {
        # `dataCallback` 메서드에서 `function() use ($data)` 형태로 클로저를 정의할 때, 
        // `$data` 변수가 클로저 내부에 캡처됩니다. 
        // 이때 클로저는 `$data` 변수를 `static` 속성에 저장합니다.
        $this->dataCallback = function () use ($data) {
            return $data;
        };
    }

    public function myFunction2()
    {
        if (is_callable($this->dataCallback)) {
            // print_r($this->dataCallback);
            // `myFunction2` 메서드에서 클로저를 호출할 때, 
            // `call_user_func($this->dataCallback)`는 캡처된 `$data` 변수를 반환합니다. 
            // 이때 클로저는 캡처된 환경을 사용하여 `$data`를 반환합니다.
            $data = call_user_func($this->dataCallback);
            return $data;
        }
        return null;
    }
}
```

`print_r($this->dataCallback);` 코드로 클로저를 출력하면 다음과 같은 구조가 나타납니다:

```plaintext
# 클로저 자체는 `Closure` 클래스의 인스턴스입니다
Closure Object
(
    # `static` 속성은 클로저가 캡처한 변수들이 저장되는 배열입니다. 
    # 여기서는 `data` 변수를 캡처하여 저장하고 있습니다.
    [static] => Array
        (
            [data] => Array
                (
                    [key] => value
                )
        )
    # 클로저가 정의된 객체(`MyComponent`)를 가리킵니다. 
    # 이는 클로저가 그 객체의 메서드 내에서 정의되었음을 나타냅니다.
    [this] => MyComponent Object
        (
            # `dataCallback`은 다시 클로저 객체를 가리키며, 재귀적으로 구조가 반복됩니다.
            [dataCallback:MyComponent:private] => Closure Object
                (
                    [static] => Array
                        (
                            [data] => Array
                                (
                                    [key] => value
                                )
                        )
                    [this] => MyComponent Object
                        (
                            [dataCallback:MyComponent:private] => Closure Object
                                ...
                        )
                )
        )
)
```

다른 함수를 예로 들어보겠습니다.

```php
class MyComponent {
    public $someCallback;

    public function someMethod($saveData) {
        $this->someCallback = $this->dataClosure($saveData);
    }

    private function dataClosure($data) {
        return function() use ($data) {
            return $data;
        };
    }
}

// MyComponent 객체 생성
$component = new MyComponent();

// someMethod 호출, $saveData를 전달
$component->someMethod(array('key' => 'value'));

// 클로저 호출
$result = ($component->someCallback)(); // 클로저 실행
print_r($result); // array('key' => 'value') 출력
```

1. **someMethod 호출**:
    - `$saveData`라는 인자를 받아서 `someMethod` 메서드가 호출됩니다.
    - 예: `$component->someMethod(array('key' => 'value'));`

2. **dataClosure 호출**:
    - `someMethod` 메서드 내부에서 `dataClosure` 메서드를 호출하며, `$saveData`를 인자로 전달합니다.
    - 호출되는 메서드: `$this->dataClosure($saveData)`

3. **클로저 생성**:
    - `dataClosure` 메서드는 전달받은 `$data` 인자를 캡처하여 클로저를 생성합니다.
    - 클로저 생성 구문: `function() use ($data) { return $data; }`
    - 이 클로저는 `$data` 변수를 캡처하여, 클로저가 호출될 때마다 이 변수를 반환합니다.
    - 위의 예제에서 `$data`에는 `array('key' => 'value')`가 캡처됩니다.

4. **클로저 반환**:
    - `dataClosure` 메서드는 생성된 클로저를 반환합니다.
    - 반환되는 값은 `function() use ($data) { return $data; }` 형태의 익명 함수입니다.
    - 반환된 클로저는 `someMethod` 메서드로 돌아갑니다.

5. **someCallback에 클로저 저장**:
    - `someMethod` 메서드 내부에서 `dataClosure` 메서드가 반환한 클로저를 `$this->someCallback`에 저장합니다.
    - 예: `$this->someCallback = $this->dataClosure($saveData);`

## 클로저, 로컬 변수, 그리고 클래스 속성

testMyFunctions 함수 내에서 로컬 변수로 선언된 tmp 변수가 제대로 작동하지 않는 이유는 그 변수의 스코프와 생명 주기 때문입니다.
- 로컬 변수: *해당 함수의 실행이 끝나면 소멸*되기 때문에 다른 함수에서 이를 참조할 수 없습니다.
- 클래스 속성: 객체의 생명 주기 동안 유지되므로 여러 메서드에서 접근할 수 있습니다.

이러한 이유로, 테스트 코드에서 클로저를 여러 메서드 간에 공유하려면 클래스 속성으로 저장하는 것이 필요합니다.
로컬 변수를 사용하여 이를 구현하려면, 로컬 변수가 소멸되지 않도록 클로저를 함수 호출 시점에 직접 전달하거나 반환값으로 사용하는 등의 다른 접근 방식이 필요합니다.

1. **로컬 변수**:
   - 특정 함수 내에서만 존재합니다.
   - 함수 호출이 끝나면 메모리에서 해제됩니다.
   - 다른 함수에서는 접근할 수 없습니다.

2. **클래스 속성**:
   - 객체의 생명 주기 동안 존재합니다.
   - 동일한 객체 내의 여러 메서드에서 접근할 수 있습니다.
   - 객체가 소멸될 때 메모리에서 해제됩니다.

클로저를 로컬 변수로 지정하는 경우 작동하지 않을 수 있습니다.
이유는 `myFunction1`과 `myFunction2`가 실제로 호출될 때 `$tmp` 변수가 올바르게 설정되지 않거나 스코프 밖에 있기 때문입니다.

```php
public function testMyFunctions() {
    $tmp = null; // $tmp 변수를 함수 내에 선언

    // myFunction1을 모의하고 클로저를 설정합니다.
    $this->MyComponent->MyModel
        ->expects($this->once())
        ->method('myFunction1')
        ->with($this->callback(function($saveData) use (&$tmp) {
            $tmp = $this->dataClosure($saveData);
            return true;
        }));

    // myFunction2를 모의하고 클로저를 사용합니다.
    $this->MyComponent->MyModel
        ->expects($this->once())
        ->method('myFunction2')
        ->with($this->callback(function($anotherData) use (&$tmp) {
            if (is_callable($tmp)) {
                $savedData = $tmp();
                return $savedData === array('key' => 'value');
            }
            return false;
        }));

    // 실제 테스트 실행
    $this->MyComponent->MyModel->myFunction1(array('key' => 'value'));
    $this->MyComponent->MyModel->myFunction2(array('key' => 'value'));
}
```

반면 클래스 속성으로 클로저를 저장하는 경우  `myFunction1`이 호출된 후에도 클로저가 유지됩니다.
따라서 `myFunction2`에서도 동일한 클로저에 접근할 수 있습니다.

```php
class MyComponentTest extends CakeTestCase {
    public $MyComponent;
    private $tmp; // 클래스 속성으로 클로저를 저장하기 위한 변수

    public function setUp() {
        parent::setUp();
        $Collection = new ComponentCollection();
        $this->MyComponent = new MyComponent($Collection);
        $this->MyComponent->initialize(new Controller());
        $this->tmp = null; // 초기화
    }

    private function dataClosure($data) {
        return function() use ($data) {
            return $data;
        };
    }

    public function testMyFunctions() {
        // myFunction1을 모의하고 클로저를 설정합니다.
        $this->MyComponent->MyModel
            ->expects($this->once())
            ->method('myFunction1')
            ->with($this->callback(function($saveData) {
                $this->tmp = $this->dataClosure($saveData);
                return true;
            }));

        // myFunction2를 모의하고 클로저를 사용합니다.
        $this->MyComponent->MyModel
            ->expects($this->once())
            ->method('myFunction2')
            ->with($this->callback(function($anotherData) {
                if (is_callable($this->tmp)) {
                    $savedData = $this->tmp();
                    return $savedData === array('key' => 'value');
                }
                return false;
            }));

        // 실제 테스트 실행
        $this->MyComponent->MyModel->myFunction1(array('key' => 'value'));
        $this->MyComponent->MyModel->myFunction2(array('key' => 'value'));
    }
}
```

## 코드 예제

### PHP

```php
class MyComponent extends Component {
    public $saveDataCallback;

    public function initialize(Controller $controller) {
        parent::initialize($controller);
        $this->saveDataCallback = null;
    }

    public function myFunction1($data) {
        $this->saveDataCallback = function() use ($data) {
            return $data;
        };
        // 실제 작업 수행
    }

    public function myFunction2() {
        if (is_callable($this->saveDataCallback)) {
            $data = call_user_func($this->saveDataCallback);
            // myFunction1에서 전달된 데이터를 사용하여 작업 수행
            return $data;
        }
        return null;
    }
}
```

1. **클로저 정의**: `myFunction1` 메서드가 호출되면, `$data` 변수를 받아 클로저를 정의합니다.
   - 클로저는 `function() use ($data)` 형식으로 정의되며, 이는 `$data` 변수를 클로저 내부로 캡처하여 사용할 수 있게 합니다.
   - `function() use ($data)`는 `$data` 변수를 클로저 내부에서 접근할 수 있게 합니다.
   - 클로저는 `$this->saveDataCallback` 변수에 저장됩니다.

2. **클로저 호출**: `myFunction2` 메서드가 호출되면, `$this->saveDataCallback`이 클로저인지 확인합니다.
   - `is_callable($this->saveDataCallback)`를 사용하여 `$this->saveDataCallback`이 호출 가능한 함수인지 확인합니다.
   - 클로저가 정의되어 있고 호출 가능하다면, `call_user_func($this->saveDataCallback)`를 사용하여 클로저를 호출합니다.
   - 클로저가 호출되면, `myFunction1`에서 설정한 `$data` 변수를 반환합니다.

3. **클로저 동작**: 클로저가 호출되면서, `myFunction1`에서 전달된 `$data` 변수를 반환합니다.
   - `myFunction2`는 이 반환된 데이터를 받아서 필요한 작업을 수행할 수 있습니다.

#### 테스트 코드

```php
class MyComponentTest extends CakeTestCase {
    public $MyComponent;

    public function setUp() {
        parent::setUp();
        $Collection = new ComponentCollection();
        $this->MyComponent = new MyComponent($Collection);
        $this->MyComponent->initialize(new Controller());
    }

    public function testMyFunctions() {
        // myFunction1 호출
        $this->MyComponent->myFunction1(array('key' => 'value'));

        // myFunction2에서 $saveData가 제대로 전달되는지 확인
        $result = $this->MyComponent->myFunction2();
        $this->assertEquals(array('key' => 'value'), $result);
    }
}
```

## 클로저는 함수와 그 함수가 선언된 환경을 함께 캡처한다

클로저(Closure)는 함수와 그 함수가 선언된 환경(즉, 클로저 내부에서 사용되는 변수)을 함께 캡처하여 저장합니다.
따라서 클로저 객체 자체를 출력하면 *함수 객체*와 함께 클로저가 *캡처한 변수들에 대한 정보*도 함께 출력됩니다.
이 때문에 클로저 객체 자체를 출력하면 복잡한 정보가 함께 나타날 수 있습니다.

### 클로저 내부의 데이터만 출력하는 방법

클로저 객체 자체를 출력하는 대신, 클로저를 호출하여 그 결과를 출력하면 원하는 데이터를 얻을 수 있습니다. 아래는 이를 구현한 예제입니다.

### 코드 예제

```php
class MyComponent {
    public function dataClosure($data) {
        return function() use ($data) {
            return $data;
        };
    }
}

// MyComponent 객체 생성
$component = new MyComponent();

// 클로저 생성
$closure = $component->dataClosure(array('key' => 'value'));

// 클로저 호출하여 결과 출력
$result = $closure();
print_r($result);
```

1. **클로저 정의**: `dataClosure` 메서드는 `$data` 변수를 캡처하는 클로저를 반환합니다.
2. **클로저 생성**: `dataClosure` 메서드를 호출하여 클로저를 생성하고 이를 `$closure` 변수에 저장합니다.
3. **클로저 호출**: 클로저를 호출하여 캡처된 `$data` 변수를 반환받습니다.
4. **결과 출력**: 클로저를 호출한 결과를 출력하여 캡처된 데이터를 확인합니다.

### 클로저 출력 결과

위 코드에서 `$result`에는 클로저가 반환하는 데이터가 저장됩니다. `print_r($result)`를 사용하면 `$data` 변수가 출력됩니다.

이 방식으로 클로저 객체 자체를 출력하지 않고, 클로저를 호출하여 실제로 원하는 데이터를 얻을 수 있습니다. 이를 통해 복잡한 클로저 객체의 구조를 피하고, 캡처된 변수의 값을 쉽게 확인할 수 있습니다.

## 고급

### returnValue(클로저 호출) 은 원하는대로 동작하지 않을 수 있다

```php
... 생략 ...
// 캡처한 데이터를 반환합니다.
->will($this->returnCallback(function () {
    return call_user_func($this->paymentDataCallback);
}));


// Null을 반환합니다.
->will($this->returnValue(call_user_func($this->paymentDataCallback)))
```

이 둘의 차이는 **평가 시점**에 있습니다.

1. **`returnCallback`**: 지연 평가

    `returnCallback`은 인자로 전달된 콜백 함수를 호출하여 그 반환값을 돌려줍니다.
    콜백 함수는 *스텁(stubbed) 메서드가 호출될 때마다 실행*됩니다.

    ```php
    ->will($this->returnCallback(function () {
        return call_user_func($this->paymentDataCallback);
    }));
    ```

    즉, `will` 메서드가 호출될 때마다 익명 함수를 실행하고, `call_user_func($this->paymentDataCallback)`의 반환값을 돌려줍니다.

2. **`returnValue`**: 즉시 평가

    `returnValue`는 전달된 값을 그대로 반환합니다.
    *한 번 평가된 값이 스텁(stubbed) 메서드의 반환값으로 사용*됩니다.
    전달된 값을 즉시 평가(실행)하고 그 결과를 반환값으로 사용합니다.
    여기서 "평가"는 전달된 표현식이나 함수를 실행하여 그 결과를 얻는 것을 의미합니다.
    만약 전달된 값이 콜백 함수일 경우, 그 콜백 함수는 평가되지 않고 함수 자체가 반환됩니다.

   ```php
   ->will($this->returnValue(call_user_func($this->paymentDataCallback)));
   ```

   여기서 중요한 점은 `call_user_func($this->paymentDataCallback)`는 `will` 메서드가 호출될 때 **단 한 번 실행된다**는 것입니다.
   그 결과가 무엇이든지 간에, 그것이 평가된 값입니다.

   `will` 메서드가 호출될 때 `call_user_func($this->paymentDataCallback)`의 반환값을 한 번 평가합니다.
   그 평가된 값을 스텁 메서드의 반환값으로 사용합니다.
   따라서 `call_user_func($this->paymentDataCallback)`가 실제로 한 번 실행되고 그 결과가 `returnValue`에 전달됩니다.
   이 경우, `paymentDataCallback`의 결과가 `null`일 수 있습니다.
