# pkg design

## Golang에서 서브 패키지에 대부분의 구현체가 위치한 경우

Golang에서는 패키지 구조가 매우 중요하며, *각 패키지가 명확한 책임을 가지도록 설계되어야* 합니다.
각 기능별로 패키지를 분리하고, 인터페이스와 구현체를 명확하게 나누는 것이 중요합니다.

그렇기 때문에 만약 서브 패키지에 대부분의 구현체가 집중되어 있으면 다음과 같은 문제가 발생할 수 있습니다:

- **복잡성 증가**: 코드가 여러 서브 패키지로 나뉘어져 있지만, 실제로는 특정 서브 패키지에 모든 중요한 구현이 집중되어 있어, 의존성이 복잡해지고 코드를 이해하기 어려워질 수 있습니다.
- **캡슐화와 정보 은닉이 제대로 이루어지지 않음**: 서브 패키지에 너무 많은 기능이 집중되면, 패키지 간의 경계가 모호해지고, 외부 패키지에서 서브 패키지의 내부 구현에 지나치게 의존하게 될 수 있습니다.
- **단일 책임 원칙(SRP)의 위반**: 하나의 서브 패키지가 여러 책임을 가지게 되면, 코드의 모듈성이 떨어지고, 유지보수가 어려워집니다.

```plaintext
project/
│
├── core/
│   ├── service.go
│   ├── repository.go
│   └── controller.go
│
└── core/impl/
    ├── service_impl.go
    ├── repository_impl.go
    └── controller_impl.go
```

위 코드 구조에서는 `core/impl/` 패키지에 모든 구현체가 집중되어 있습니다.
이 경우의 문제는 다음과 같습니다:
- **기능이 모호**: `core/impl/` 패키지에 다양한 구현체가 혼재해 있어, 코드의 기능과 역할이 모호해집니다.
- **테스트 어려움**: 특정 구현체만 테스트하려고 할 때, 다른 구현체와의 의존성 때문에 테스트가 복잡해질 수 있습니다.

프로젝트의 확장성과 유지보수성을 높이기 위해 다음과 같은 구조로 개선할 수 있습니다:

```plaintext
project/
│
├── core/
│   ├── service/
│   │   ├── service.go
│   │   └── service_impl.go
│   │
│   ├── repository/
│   │   ├── repository.go
│   │   └── repository_impl.go
│   │
│   └── controller/
│       ├── controller.go
│       └── controller_impl.go
```

- **기능별 패키지 분리**

    각 기능(서비스, 리포지토리, 컨트롤러)이 별도의 패키지로 나뉘어져 있습니다.
    이를 통해 각 패키지의 책임이 명확해지고, 코드의 가독성과 관리가 용이해집니다.

    ```go
    // core/service/service.go
    package service

    type Service interface {
        DoSomething() error
    }
    ```

    ```go
    // core/service/service_impl.go
    package service

    type DefaultService struct{}

    func (s *DefaultService) DoSomething() error {
        // 구체적인 서비스 로직 구현
        return nil
    }
    ```

    위 코드에서 `service.go` 파일에는 인터페이스가 정의되어 있고, `service_impl.go` 파일에는 해당 인터페이스를 구현한 구체적인 서비스 로직이 위치해 있습니다.

- **구현체와 인터페이스 분리**

    인터페이스(`service.go`, `repository.go`, `controller.go`)와 구체적인 구현(`service_impl.go`, `repository_impl.go`, `controller_impl.go`)이 명확하게 분리되어 있습니다.
    이 구조는 테스트와 유지보수를 용이하게 하고, 의존성을 줄이는 데 도움을 줍니다.

- **단일 책임 원칙 준수**

    각 패키지는 단일한 책임만을 가지며, 패키지 간의 역할이 명확히 구분되어 있습니다.

    예를 들어, `service` 패키지는 서비스 로직만을, `repository` 패키지는 데이터 접근 로직만을 담당합니다.

실제 코드가 변화하는 과정을 보면 다음과 같습니다:

```go
// core/impl/service_impl.go
package impl

type DefaultService struct{}

func (s *DefaultService) DoSomething() error {
    // 서비스 로직
    return nil
}

// core/impl/repository_impl.go
package impl

type DefaultRepository struct{}

func (r *DefaultRepository) GetData() (string, error) {
    // 데이터 접근 로직
    return "data", nil
}
```

위 코드에서는 `impl` 패키지에 서비스와 리포지토리의 구현체가 모두 포함되어 있습니다.
이는 패키지의 책임이 혼재되어 있으며, 다른 패키지에서 이 구현체를 참조할 때 의존성이 복잡해질 수 있습니다.

- core/service

    ```go
    // core/service/service.go
    package service

    type Service interface {
        DoSomething() error
    }

    // core/service/service_impl.go
    package service

    type DefaultService struct{}

    func (s *DefaultService) DoSomething() error {
        // 서비스 로직
        return nil
    }
    ```

- core/repository

    ```go
    // core/repository/repository.go
    package repository

    type Repository interface {
        GetData() (string, error)
    }

    // core/repository/repository_impl.go
    package repository

    type DefaultRepository struct{}

    func (r *DefaultRepository) GetData() (string, error) {
        // 데이터 접근 로직
        return "data", nil
    }
    ```
