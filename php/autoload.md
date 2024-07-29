# Autoload

- [Autoload](#autoload)
    - [오토로딩](#오토로딩)
    - [Fully Qualified Class Name (FQCN)](#fully-qualified-class-name-fqcn)
    - [`PSR-0`](#psr-0)
    - [`PSR-4`](#psr-4)
        - [참고](#참고)
    - [composer `dump-autoload`](#composer-dump-autoload)
        - [예제](#예제)
        - [참고](#참고-1)
    - [오토로딩 관련 유용한 함수](#오토로딩-관련-유용한-함수)

## 오토로딩

PHP에서 클래스가 필요할 때 그 클래스가 아직 로드되지 않았다면, 오토로더(autoloader)가 자동으로 해당 클래스 파일을 찾아 로드합니다.
이는 `spl_autoload_register()` 함수를 사용하여 구현할 수 있습니다.
PSR-4와 Composer와 같은 표준을 사용하면 더 효율적이고 일관된 오토로딩을 구현할 수 있습니다.

오토로딩은 클래스의 인스턴스 생성이나 정적 메소드 호출 등 런타임에 클래스가 실제로 필요할 때만 발동합니다.

> [실제로 필요할 때](./class.md#클래스가-실제로-사용될-때)란?
>
> 클래스를 "실제로 사용한다"는 것은 클래스를 메모리에 로드하고, 그와 관련된 작업을 수행하는 것을 의미합니다.
> 이는 *인스턴스화*, *정적 메서드 호출*, *정적 속성 접근*, *클래스 상속* 및 *인터페이스 구현* 등을 포함합니다.
> 이러한 과정은 컴파일 타임이 아닌 런타임에서 이루어지며, PHP의 오토로딩 메커니즘과 메모리 관리에 의해 처리됩니다.
>
> - 메모리 로드
>
>   클래스 정의가 메모리에 로드됩니다.
>   이는 메모리의 특정 영역에 클래스의 메타데이터와 코드가 저장되는 것을 의미합니다.
>
> - 인스턴스 생성
>
>   객체가 생성되면, 이는 메모리에서 할당된 객체의 인스턴스가 초기화된다는 것을 의미합니다.
>
> - 메서드 실행 및 속성 접근
>
>   메서드가 호출되거나 속성에 접근하면, CPU는 해당 명령을 실행하여 메모리에서 데이터를 읽거나 씁니다.

오토로딩은 다음과 같은 원리로 동작합니다:

1. **오토로더 등록**:

    `spl_autoload_register()` 함수를 사용하여 하나 이상의 오토로더를 등록합니다.
    이 함수는 콜백 함수를 인자로 받습니다.

2. **클래스 필요 시점**

    클래스가 인스턴스화되거나 정적 메소드가 호출되는 등의 경우에 PHP는 해당 클래스가 로드되었는지 확인합니다.

3. **오토로더 호출**

    클래스가 로드되지 않은 경우, PHP는 등록된 오토로더를 순차적으로 호출하여 클래스를 로드하려 시도합니다.

4. **클래스 로드**

    오토로더가 클래스를 찾으면 해당 파일을 포함하고 클래스를 로드합니다.
    그렇지 않으면, 다음 오토로더가 호출됩니다.
    모든 오토로더가 실패하면 오류가 발생합니다.

다음은 오토로딩이 발생하는 예제입니다:

```php
<?php
spl_autoload_register(function ($class) {
    echo "Attempting to load $class.\n";
    include $class . '.php';
});

// 클래스 이름을 가져오기
echo Does\Not\Exist::class;  // 출력: Does\Not\Exist

// 실제 클래스 로드 시도
new Does\Not\Exist();  // 여기서 오토로더가 작동하며, 클래스 파일을 찾지 못하면 오류 발생

class_exists('Does\Not\Exist'); // 이 또한 오토로더를 트리거함
```

이 예제에서는 `class_exists` 함수가 호출되면서 오토로딩이 트리거되어 `Does\Not\Exist` 클래스의 존재 여부를 확인하려 합니다.
만약 클래스가 존재하지 않으면, 오토로더가 이를 처리하려 시도하고, 최종적으로 클래스 파일을 찾지 못하면 오류가 발생합니다.

## Fully Qualified Class Name (FQCN)

Fully Qualified Class Name(FQCN)은 클래스의 전체 경로를 네임스페이스를 포함하여 표현한 이름입니다.
이는 클래스가 정의된 정확한 위치를 지정하기 때문에, 동일한 이름의 클래스가 다른 네임스페이스에 있을 때 충돌을 방지할 수 있습니다.

```php
// 클래스를 로드할 때 FQCN 사용
$controller = new \App\Controller\HomeController();

// 정적 메서드를 호출할 때 FQCN 사용
// `<NamespaceName>(\<SubNamespaceNames>)*\<ClassName>`
\App\Controller\HomeController::index();
```

- `<NamespaceName>`

    FQCN에는 반드시 최상위 네임스페이스(top-level namespace)가 포함되어야 합니다.
    이는 보통 벤더 네임스페이스로, 라이브러리나 프로젝트의 고유한 식별자 역할을 합니다.

    > **vendor namespace**
    >
    > 주로 라이브러리, 프레임워크, 애플리케이션의 고유한 식별자로 사용됩니다
    > 이는 프로젝트나 라이브러리를 제공하는 "벤더" 또는 "제공자"를 나타냅니다.

- `(\<SubNamespaceNames>)*`

    FQCN에는 하나 이상의 하위 네임스페이스(sub-namespace)가 포함될 수 있습니다.
    이는 클래스를 논리적으로 그룹화하고 계층 구조를 형성하는 데 사용됩니다.

    ***하위 디렉토리의 이름의 대소문자*는 반드시 *하위 네임스페이스 이름의 대소문자*가 일치**해야 한다.

- `<ClassName>`

    FQCN의 마지막 부분은 반드시 클래스 이름이어야 합니다.(최종 클래스 이름, terminating class name)
    이는 네임스페이스와 함께 클래스를 고유하게 식별합니다.

    마지막의 클래스 이름은 `.php`로 끝나는 파일명에 상응해야 합니다.
    **파일의 대소문자는 반드시 해당 클래스 이름의 대소문자와 일치**해야 합니다.

- 언더스코어(`_`)는 의미 없음

    네임스페이스나 클래스 이름에서 사용되는 언더스코어(`_`)는 특별한 의미가 없으며, 단순히 문자로 취급됩니다.

> **네임스페이스의 목적**
>
> - 이름 충돌 방지
>
>   네임스페이스는 동일한 이름을 가진 클래스가 서로 다른 네임스페이스에 존재할 수 있게 하여 이름 충돌을 방지합니다.
>
> - 조직화
>
>   네임스페이스는 코드를 논리적으로 그룹화하여 코드의 구조와 가독성을 향상시킵니다.
>
> - 자동 로딩
>
>   PSR-4와 같은 오토로딩 표준과 결합하여 파일을 자동으로 로드할 수 있습니다.
>

## `PSR-0`

*네임스페이스와 클래스 이름*이 *디렉토리 구조와 일치*해야 합니다.

예를 들어, `App\Controller\HomeController` 클래스는 `src/App/Controller/HomeController.php` 파일에 있어야 합니다.

## [`PSR-4`](https://www.php-fig.org/psr/psr-4/)

PSR-4는 PHP Framework Interop Group (PHP-FIG)에서 제정한 오토로딩 표준입니다.
이 표준은 *네임스페이스와 파일 경로 간의 일관성을 유지*하여 클래스 파일을 자동으로 로드할 수 있도록 규칙을 정의합니다.
이를 위해 *네임스페이스의 루트*와 *디렉토리의 루트*를 매핑합니다.

예를 들어, `composer.json` 파일에서 `"App\\": "src/"`와 같이 설정할 수 있습니다.

```json
{
    "autoload": {
        "psr-4": {
            "App\\": "src/"
        }
    }
}
```

이때 `App\Controller\HomeController` 클래스의 네임스페이스는 `App\Controller`가 되고,
이 클래스는 `src/Controller/HomeController.php` 파일에 있습니다.

```php
// app/Controller/HomeController.php
namespace App\Controller;

class HomeController {
    // ...
}
```

`spl_autoload_register` 함수는 PHP에서 클래스 오토로더를 등록하는 기능을 제공합니다.
이 함수는 하나 이상의 콜백 함수를 등록하여, *클래스가 필요할 때 자동으로 해당 콜백 함수를 호출*하여 클래스를 로드합니다.
오토로딩의 표준을 정의하는 사양 PSR-4은 `spl_autoload_register` 함수를 통해 아래와 같이 구현될 수 있습니다.

파일 로딩 규칙은 PSR-4 표준에 의해 정의된 네임스페이스와 디렉토리 구조를 일치시키는 방법입니다.

```php
<?php
spl_autoload_register(function ($class) { // `App\Controller\HomeController`
    // 1. 네임스페이스 접두사와 기본 디렉토리 매핑
    //     
    //     네임스페이스 접두사는 네임스페이스 체계의 최상위 부분으로, 코드의 위치를 결정하는 기본 디렉토리와 연결됩니다.
    //     여기서 `App`은 네임스페이스 접두사입니다. 
    //    `App` 접두사는 프로젝트의 특정 디렉토리와 매핑됩니다. 
    //    `composer.json` 파일에서 이를 설정할 수 있습니다:
    $prefix = 'App\\';
    $base_dir = __DIR__ . '/src/';
    
    $len = strlen($prefix); // 4
    if (strncmp($prefix, $class, $len) !== 0) {
        return;
    }

    // 2. 하위 네임스페이스와 디렉토리 구조 일치
    //
    //    네임스페이스 접두사 뒤에 오는 하위 네임스페이스는 기본 디렉토리의 하위 디렉토리로 변환됩니다. 
    //    여기서 대소문자가 일치해야 합니다.
    //    - `App` 네임스페이스 접두사는 `src/` 디렉토리로 변환됩니다.
    //    - `Controller` 하위 네임스페이스는 `src/Controller/` 디렉토리로 변환됩니다.
    // 
    $relative_class = substr($class, $len); // Controller\HomeController
    // `src/` 디렉토리 내의 파일에서 찾습니다. 
    $file = $base_dir
        . str_replace('\\', '/', $relative_class) // Controller/HomeController
        // 3. 클래스 이름과 파일 확장자
        // 
        //    네임스페이스의 마지막 부분은 클래스 이름이며, 이 클래스 이름은 `.php` 확장자를 가진 파일이 됩니다.
        //    - 네임스페이스: `App\Controller`
        //    - 클래스 이름: `HomeController`
        // 
        //    따라서 파일 경로는 `src/Controller/HomeController.php`입니다.
        . '.php';
    
    // 이 규칙에 따라 PHP 오토로더는 
    // `App\Controller\HomeController` 클래스를 찾을 때 
    // `__DIR__/src/Controller/HomeController.php` 파일을 로드합니다.
    if (file_exists($file)) {
        require $file;
    }
});
```

### 참고

- [How Composer’s PSR-4 autoloading works](https://sajadtorkamani.com/how-composers-psr-4-autoloading-works/)
- [example file](https://github.com/php-fig/fig-standards/blob/master/accepted/PSR-4-autoloader-examples.md)
- [Composer PSR-4 Autoloader Class not found](https://stackoverflow.com/questions/44468991/composer-psr-4-autoloader-class-not-found)

## [composer `dump-autoload`](https://getcomposer.org/doc/03-cli.md#dump-autoload-dumpautoload)

PHP의 의존성 관리 도구인 Composer는 자동으로 오토로더를 생성하여 프로젝트의 클래스 파일을 효율적으로 로드할 수 있도록 합니다.
Composer는 PHP의 의존성 관리 도구로, PSR-4 규칙을 자동으로 처리하는 오토로더를 생성합니다.

`composer.json` 파일에 다음과 같이 오토로딩 설정을 추가합니다:

```json
{
    "autoload": {
        "psr-4": {
            "App\\": "src/"
        }
    }
}
```

그리고 나서, `composer install` 또는 `composer dump-autoload` 명령어를 실행하여 오토로더 파일을 생성합니다.

```sh
composer dump-autoload
```

이제 Composer가 자동으로 `src/` 디렉토리 내의 클래스를 로드합니다.

### [예제](https://sajadtorkamani.com/how-composers-psr-4-autoloading-works/)

```json
{
  "name": "sajadtorkamani/understanding-psr4-autoloading",
  "autoload": {
    "psr-4": {
      "App\\": "src"
    }
  },
  "require": {}
}
```

```tree
├── main.php
├── src
│   └── Utils
│       └── Logger.php
```

```php
<?php

namespace App\Utils;

class Logger
{
    public function info(string $msg): void
    {
        echo "INFO: $msg";
    }

    public function error(string $msg): void
    {
        echo "ERROR: $msg";
    }
}
```

- `App`: namespace prefix
- `Utils`: sub-namespace
- `Logger`: terminating class name
- `src/`: base directory

또한 가령 `\Aura\Web\Response\Status`이고 `/path/to/aura-web/src/Response/Status.php` 경로에 위치할 경우 아래처럼 정리할 수 있다.

- `Aura\Web`: namespace prefix
- `Response`: sub-namespace
- `Status`: terminating class name
- `/path/to/aura-web/src/`: base directory

### 참고

- [composer autoload property](https://getcomposer.org/doc/04-schema.md#autoload)

## 오토로딩 관련 유용한 함수

- **spl_autoload_register**: 오토로더를 등록합니다.
- **spl_autoload_unregister**: 등록된 오토로더를 해제합니다.
- **spl_autoload_functions**: 현재 등록된 모든 오토로더를 반환합니다.
- **spl_autoload_call**: 지정된 클래스 이름으로 오토로더를 실행합니다.
