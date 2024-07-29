# [new](https://pkg.go.dev/builtin#new)

## `new` function?

> The new built-in function allocates memory.
> The first argument is a type, not a value, and the value returned is a pointer to a newly allocated zero value of that type.

## 스택 메모리에서 포인터 사용 대신 `new` 함수 사용

스택 메모리, 힙 메모리, 그리고 포인터를 사용하는 방법에 대한 이해는 메모리 관리와 프로그램의 성능에 직접적인 영향을 미친다.

### 스택 메모리 vs. 힙 메모리

- **스택 메모리**

    함수 호출과 함께 생성되는 지역 변수들이 저장되는 영역이다.
    스택은 자동으로 관리되며, 함수가 종료될 때 그 함수의 지역 변수들이 자동으로 해제된다.
    스택 메모리는 할당과 해제가 빠르다는 장점이 있지만, 고정된 크기를 가지고 있고, 함수 호출의 컨텍스트 내에서만 생존한다.
  
- **힙 메모리**

    동적 메모리 할당에 사용되며, 프로그램의 런타임 중에 할당되고 해제된다.
    힙 메모리는 프로그래머에 의해 직접 관리되거나, 가비지 컬렉터에 의해 관리된다.
    힙에 할당된 메모리는 프로그램의 어느 곳에서나 접근할 수 있으며, 사용이 끝난 후에는 명시적으로 해제하거나 가비지 컬렉터가 자동으로 해제한다.

### `new` 사용과 포인터

`new` 키워드는 힙 메모리에 새로운 변수를 할당하고, 해당 변수의 포인터를 반환한다.
이 방법을 사용하면, 할당된 메모리는 함수의 생명 주기를 넘어서서도 생존할 수 있으며, 반환된 포인터를 통해 해당 변수를 참조하고 수정할 수 있다.

```go
someType := new(SomeType) // 힙 메모리에 SomeType 인스턴스를 할당하고, 그 포인터를 someType에 저장합니다.
```

### 스택에서 포인터 사용

함수 내에서 스택 메모리에 저장된 변수의 포인터를 다루는 것은 그 포인터가 가리키는 변수가 함수가 종료될 때 사라지기 때문에, 함수 밖으로 그 포인터를 반환하는 것은 위험할 수 있다. 함수 밖에서 그 포인터를 사용하려고 할 때 이미 해제된 메모리를 참조하게 되므로, 예상치 못한 동작이나 오류를 발생시킬 수 있다.

하지만 `new` 함수나 `&` 연산자와 함께 구조체 리터럴을 사용하여 변수를 초기화하고 포인터를 반환할 때, 해당 변수는 힙 메모리에 할당된다. 따라서 이러한 방법으로 생성된 포인터는 함수가 종료된 후에도 안전하게 사용할 수 있다.

### 제안된 코드의 차이점

제안된 코드에서는 `new`를 사용하여 `SomeType`와 `SubSomeType` 인스턴스를 명시적으로 힙에 할당한다.
이 방식은 반환된 포인터를 통해 해당 인스턴스에 안전하게 접근하고 수정할 수 있게 한다.
반면, 스택에 할당된 변수의 포인터를 사용하면, 그 변수가 더 이상 유효하지 않을 때 그 포인터를 사용하는 위험이 있다.

### 결론

"스택에서 포인터를 사용하는 것보다는 `new`를 사용하는 게 더 좋다"는 지침은 특히 함수 밖으로 변수의 포인터를 반환하거나, 함수의 생명 주기를 넘어서서 변수를 사용해야 할 때 중요합니다. `new`를 사용함으로써 안전하게 메모리를 할당하고, 해당 메모리에 대한 포인터를 반환하여, 함수 밖에서도 해당 변수를 안전하게 사용할 수 있게 됩니다. 이는 메모리 안전성과 프로그램의 안정성을 보장하는 중요한 방법입니다.

### Examples

#### `new` 또는 `&` 연산자 사용

```go
func AAAAA(ctx *fiber.Ctx) (*SomeType, error) {
    // SubSomeType 인스턴스를 생성합니다.
    sub := new(SubSomeType)

    // sub.parse(ctx)를 호출하여 파싱을 시도합니다.
    if err := sub.parse(ctx); err != nil {
        // 에러가 발생한 경우, 에러를 반환합니다.
        // 여기서는 SomeType 인스턴스 대신 nil을 반환합니다.
        return nil, errors.Wrap(err, "pReq.parseParams")
    }

    // 파싱에 성공했다면, SomeType 인스턴스를 생성합니다.
    someType := new(SomeType)
    someType.Sub = *sub, // sub의 값을 SomeType의 Sub 필드에 할당합니다.

    // SomeType 인스턴스와 nil 에러를 반환합니다.
    return someType, nil
}
```

```go
func AAAAA(ctx *fiber.Ctx) (*SomeType, error) {
    // SubSomeType 인스턴스를 생성합니다.
    sub := &SubSomeType{}

    // sub.parse(ctx)를 호출하여 파싱을 시도합니다.
    if err := sub.parse(ctx); err != nil {
        // 에러가 발생한 경우, 에러를 반환합니다.
        // 여기서는 SomeType 인스턴스 대신 nil을 반환합니다.
        return nil, errors.Wrap(err, "pReq.parseParams")
    }

    // 파싱에 성공했다면, SomeType 인스턴스를 생성합니다.
    someType := &SomeType{
        Sub: *sub, // sub의 값을 SomeType의 Sub 필드에 할당합니다.
    }

    // SomeType 인스턴스와 nil 에러를 반환합니다.
    return someType, nil
}
```

### 스택에서 변수 선언 후 나중에 `&` 연산자 사용

```go
func AAAAA(ctx *fiber.Ctx) (*SomeType, error) {
    // 스택 메모리에 값으로 초기화하고 `sub` 변수가 실제 데이터를 직접 포함
    sub := SubSomeType{}
    // 힙 메모리에 새롭게 할당된 것이 아니라, 기존에 스택에 할당된 인스턴스의 주소를 가져온다.
    pSub := &sub

    // sub.parse(ctx)를 호출하여 파싱을 시도합니다.
    if err := sub.parse(ctx); err != nil {
        // 에러가 발생한 경우, 에러를 반환합니다.
        // 여기서는 SomeType 인스턴스 대신 nil을 반환합니다.
        return nil, errors.Wrap(err, "pReq.parseParams")
    }

    // 스택 메모리에 값으로 초기화하고 `someType` 변수가 실제 데이터를 직접 포함
    someType := SomeType{
        Sub: *sub, // sub의 값을 SomeType의 Sub 필드에 할당합니다.
    }

    // 힙 메모리에 새롭게 할당된 것이 아니라, 기존에 스택에 할당된 인스턴스의 주소를 가져온다.
    // 함수가 반환될 때, 이 주소들은 여전히 유효한 상태를 유지한다.
    // Go의 컴파일러가 이스케이프 분석을 통해 해당 변수들이 스택 범위를 벗어나 사용되어야 한다고 판단하고, 
    // 그에 따라 생명 주기를 관리하기 때문이다.
    return &someType, nil 
}
```

Go 컴파일러는 이스케이프 분석을 통해 `sub`와 `someType`가 함수 범위를 벗어나 사용될 것으로 예상되는 경우(즉, 포인터가 함수 밖으로 반환되는 경우) 이들을 힙에 할당해야 할 필요가 있는지를 결정한다. 이 예에서는 함수에서 포인터를 반환하므로, 해당 변수들은 필요에 따라 자동으로 힙에 할당될 수 있다.
