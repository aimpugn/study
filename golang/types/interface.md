# interface

## 인터페이스와 값과 포인터

```go
func NewClient(cfg Config, serviceName string) (Client, error) {
    // 이 포인터는 `Client` 인터페이스를 구현하기 때문에, `Client` 타입으로 리턴될 수 있다.
    return &clientImpl{
        client: httpClient,
    }, nil
}
```

### 인터페이스와 구조체 포인터

Go에서 인터페이스는 *메서드의 시그니처 집합*으로 정의된다.
어떤 타입이 특정 인터페이스를 구현하기 위해서는 그 인터페이스가 요구하는 모든 메서드를 구현해야 한다.
인터페이스의 모든 메서드를 구현한다면, 그 구조체의 인스턴스(값 또는 포인터)는 해당 인터페이스 타입으로 사용될 수 있다.

### 값 vs. 포인터

구조체 인스턴스에 대해 값을 리턴할 것인지, 아니면 포인터를 리턴할 것인지는 중요한 선택이다.
- 값으로 리턴하면, 리턴될 때 해당 구조체의 복사본이 생성된다
- 반면, 포인터로 리턴하면, 원본 구조체에 대한 참조가 리턴되어, 추가적인 메모리 할당 없이 원본 구조체를 조작할 수 있다.

### 사용하는 측에서의 포인터 사용

```go
package aaa

func NewClient(cfg Config, serviceName string) (Client, error) {
    // 이 포인터는 `Client` 인터페이스를 구현하기 때문에, `Client` 타입으로 리턴될 수 있다.
    return &clientImpl{
        client: httpClient,
    }, nil
}
```

```go
package bbb

func SomeLogic(cfg Config, serviceName string){
    client := aaa.NewClient(cfg, serviceName)
    newClient := NewService(&client)
    ... do something ...
}

func NewService(client *aaa.Client) Service {
  return &serviceImpl{
    client: client
  }
}
```

bbb 패키지에서는 aaa.Client 인터페이스로 리턴하는 것으로 보이므로, `&`를 다시 붙여서 `NewService(&client)`처럼 사용하게 될 수도 있을 거 같다. 이래도 문제 없을까? 아니면 새로 인스턴스 생성하는 함수를 타고 들어가서 포인터로 리턴하는지 여부 등을 직접 확인해 봐야 할까?

Go에서 *인터페이스 자체는 이미 일종의 참조 타입으로 동작*하기 때문에, 인터페이스 변수에 할당된 값이 구조체 포인터라 할지라도, 추가로 포인터를 붙여서 사용하는 것(`&client`)은 올바른 사용 방식이 아니다.

`&client`의 사용은 `client`가 이미 포인터 타입이 아닐 것으로 예상될 때 일반적으로 사용된다.
하지만 `client`는 `aaa.NewClient`에 의해 반환된 `Client` 인터페이스 타입이며, 이 경우 이미 포인터(`*clientImpl`)다. 따라서 `&client`는 포인터의 포인터(`**Client`)를 시도하는 것이 되며, 이는 컴파일 에러를 발생시킬 것이다.

포인터 등을 신경쓰지 말고, **인터페이스 타입 직접 사용**하자. `NewService` 함수가 인터페이스 `aaa.Client`를 직접 사용하도록 수정한다. 이는 `client` 변수가 이미 포인터를 저장하고 있기 때문에, 추가적인 포인터 사용이 불필요함을 의미한다.

인터페이스인 경우에는 굳이 포인터를 지정할 필요가 없다.

```go
// aaa.Client 인터페이스를 직접 받아들임으로써, 포인터로의 추가 변환을 방지합니다.
func NewService(client aaa.Client) Service {
    return &serviceImpl{
        client: client,
    }
}
```
