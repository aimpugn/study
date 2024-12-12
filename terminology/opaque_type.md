# opaque type

- [opaque type](#opaque-type)
    - [불투명 타입 (Opaque Type)](#불투명-타입-opaque-type)
    - [불투명 타입의 주요 특징](#불투명-타입의-주요-특징)
    - [불투명 타입의 사용 사례](#불투명-타입의-사용-사례)
        - [C와 C++](#c와-c)
        - [Rust](#rust)
        - [Java](#java)
        - [PHP](#php)
        - [Go](#go)

## [불투명 타입 (Opaque Type)](https://en.wikipedia.org/wiki/Opaque_data_type)

불투명 타입(opaque type)은 데이터 타입의 내부 구현 세부 사항을 외부로부터 숨기고, 외부에는 해당 타입의 참조만을 제공하는 방식입니다.
- 내부 구현 숨기기: 데이터 타입의 내부 구조를 외부에서 볼 수 없도록 하여, 구현 세부 사항을 보호합니다.
- API를 통한 간접 접근: 사용자는 제공된 API를 통해서만 데이터를 조작할 수 있어, 데이터 무결성을 유지합니다.

사용자는 이 타입의 데이터 구조나 내용을 직접 접근하거나 수정할 수 없으며,
제공된 API를 통해서만 간접적으로 작업할 수 있습니다.

불투명 타입과 반대 개념은 투명 타입 (Transparent Type)입니다.
이 타입은 데이터의 내부 구조가 외부에서 완전히 보입니다.
예를 들어, 일반적인 C의 `struct`이 이에 해당합니다.

## 불투명 타입의 주요 특징

1. 캡슐화와 추상화:
    - 불투명 타입은 데이터의 내부 구조를 숨겨 외부에서는 접근할 수 없게 합니다.
    - 외부에서 데이터와 상호작용하려면 타입이 제공하는 명시적인 함수나 메서드를 사용해야 합니다.

2. 안정성과 모듈성:
    - 내부 구현을 캡슐화함으로써 시스템 모듈화를 높이고, 내부 구현이 변경되더라도 외부 인터페이스는 변하지 않아 코드 안정성을 제공합니다.

3. 의존성 감소:
    - 외부 코드가 내부 구조에 의존하지 않으므로, 내부 구현 변경 시 외부 코드를 수정할 필요가 없습니다.

4. 주요 형태:
    - 불투명 포인터 (Opaque Pointer): C에서 흔히 사용되며, 내부 데이터 구조를 숨기는 방식입니다.
    - 언어 내 특수한 타입 선언: Swift, Rust, TypeScript 등에서 불투명 타입을 공식적으로 지원합니다.

## 불투명 타입의 사용 사례

### C와 C++

내부 데이터 구조를 숨기는 데 사용됩니다.

```c
// opaque.h: 헤더 파일
typedef struct OpaqueStruct* OpaqueType;

OpaqueType createOpaqueType();
void useOpaqueType(OpaqueType obj);
void destroyOpaqueType(OpaqueType obj);
```

```c
// opaque.c: 구현 파일
#include "opaque.h"
#include <stdlib.h>

struct OpaqueStruct {
    int internalData;
};

OpaqueType createOpaqueType() {
    OpaqueType obj = malloc(sizeof(struct OpaqueStruct));
    obj->internalData = 42;
    return obj;
}

void useOpaqueType(OpaqueType obj) {
    // 내부 데이터에 접근 가능
    printf("Internal Data: %d\n", obj->internalData);
}

void destroyOpaqueType(OpaqueType obj) {
    free(obj);
}
```

```c
// 사용하는 코드
#include "opaque.h"

int main() {
    OpaqueType obj = createOpaqueType();
    useOpaqueType(obj);
    destroyOpaqueType(obj);
    return 0;
}
```

외부에서는 `OpaqueStruct`의 내부 내용을 알 수 없고, 제공된 함수만으로 작업합니다.

### Rust

Rust의 `impl Trait`는 불투명 타입으로 동작합니다.

```rust
fn get_value() -> impl Iterator<Item = i32> {
    vec![1, 2, 3].into_iter()
}
```

반환 타입 `impl Iterator<Item = i32>`는 구체적인 이터레이터 타입을 숨기고, 인터페이스만 노출합니다.

### Java

Java는 캡슐화와 객체 지향 설계를 통해 불투명 타입의 개념을 구현합니다.
즉, 클래스와 인터페이스가 불투명 타입으로 사용됩니다.
내부 구현 세부 사항을 외부로부터 숨기고, 외부는 공개된 인터페이스(API)만 사용할 수 있습니다.

```java
public class OpaqueExample {
    private int secretData; // 외부에서 직접 접근 불가

    // 외부와의 상호작용
    public OpaqueExample(int value) {
        this.secretData = value;
    }

    // 외부와의 상호작용
    public int getData() {
        return this.secretData;
    }

    // 외부와의 상호작용
    public void setData(int value) {
        this.secretData = value;
    }
}
```

### PHP

불투명 타입을 직접적으로 지원하지는 않지만, 객체 지향 설계를 통해 가능합니다.

```php
class OpaqueExample {
    private $secretData;

    public function __construct($value) {
        $this->secretData = $value;
    }

    public function getSecretData() {
        return $this->secretData;
    }

    public function setSecretData($value) {
        $this->secretData = $value;
    }
}

$example = new OpaqueExample(42);
echo $example->getSecretData(); // 42
```

또는 PHP의 클로저도 불투명성을 제공할 수 있습니다.

```php
<?php

function createCounter($initialValue = 0) {
    $count = $initialValue;

    return [
        'increment' => function() use (&$count) {
            $count++;
        },
        'decrement' => function() use (&$count) {
            $count--;
        },
        'getCount' => function() use (&$count) {
            return $count;
        }
    ];
}

/**
 * $counter array{
 *   increment: Closure,
 *   decrement: Closure,
 *   getCount: Closure,
 * }
 */
$counter = createCounter();

// 상태 조작
$counter['increment']();
$counter['increment']();
echo $counter['getCount'](); // 2

$counter['decrement']();
echo $counter['getCount'](); // 1
```

### Go

구조체의 필드를 비공개로 선언하고, 해당 구조체를 인터페이스나 공개된 메서드로 접근할 수 있습니다.

```go
package opaque

type OpaqueType struct {
    secretData int
}

// NewOpaqueType creates an instance of OpaqueType
func NewOpaqueType(value int) *OpaqueType {
    return &OpaqueType{secretData: value}
}

// GetSecretData returns the secret data
func (o *OpaqueType) GetSecretData() int {
    return o.secretData
}

// SetSecretData sets the secret data
func (o *OpaqueType) SetSecretData(value int) {
    o.secretData = value
}
```

외부에서는 `secretData`에 직접 접근할 수 없고, `GetSecretData`와 같은 메서드를 사용해야만 데이터에 접근 가능합니다.

```go
package main

import (
    "fmt"
    "example/opaque"
)

func main() {
    obj := opaque.NewOpaqueType(42)
    fmt.Println(obj.GetSecretData()) // 42
}
```
