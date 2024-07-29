# Go Package Design

- [Go Package Design](#go-package-design)
    - [Go 패키지 구조의 의미](#go-패키지-구조의-의미)
    - [Package Oriented Design](#package-oriented-design)
    - [클린 아키텍처 기반 구조](#클린-아키텍처-기반-구조)
    - [모듈형 모노리스 구조](#모듈형-모노리스-구조)
    - [기능과 레이어의 혼합 구조](#기능과-레이어의-혼합-구조)
    - [DDD(Domain-Driven Design) 기반 구조](#ddddomain-driven-design-기반-구조)

## Go 패키지 구조의 의미

Go 언어에서 패키지 구조는 *코드 조직화*, *모듈화*, 그리고 *의존성 관리*에 중요한 역할을 합니다.
패키지는 Go의 코드 구성 단위로, 관련된 기능을 그룹화하고 재사용성을 향상시키며 프로그램의 구조를 명확하게 만듭니다.

패키지 구조의 주요 의미:

1. 코드 조직화: 관련 기능을 논리적으로 그룹화합니다.
2. 캡슐화: 패키지 내부 구현을 숨기고 공개 인터페이스만 노출합니다.
3. 재사용성: 독립적인 기능 단위로 다른 프로젝트에서 재사용할 수 있습니다.
4. 의존성 관리: 패키지 간의 의존 관계를 명확히 하고 순환 의존성을 방지합니다.
5. 컴파일 최적화: Go 컴파일러가 패키지 단위로 최적화를 수행할 수 있습니다.

## Package Oriented Design

비즈니스 도메인이나 기능을 중심으로 패키지를 구성하는 방식을 따릅니다.

Package Oriented Design은 소프트웨어 아키텍처 접근 방식으로, 관련된 기능과 책임을 논리적으로 그룹화하여 독립적인 패키지로 구성하는 방법입니다.
이 설계 방식은 코드의 모듈성, 재사용성, 그리고 유지보수성을 향상시키는 것을 목표로 합니다.

주요 특징:
- 비즈니스 도메인이나 기능을 기준으로 패키지를 구성
- 각 패키지는 독립적으로 작동할 수 있도록 가능한 한 완전한 기능 단위를 포함.
- 패키지 간 의존성을 최소화하고 명확히 정의

```sh
project/
├── cmd/
│   ├── main.go
│   └── ...
├── user/
│   ├── model/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   └── ...
├── product/
│   ├── model/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   └── ...
└── utils/
    ├── logger/
    ├── config/
    └── ...
```

Package Oriented Design은 여러 소프트웨어 설계 원칙과 패턴에 기반을 두고 있습니다:

- 모듈화 (Modularity): 관련 기능을 논리적 단위로 그룹화하여 복잡성을 관리

    각 비즈니스 도메인이나 기능을 별도의 패키지로 분리합니다.
    `user`, `product`, `order` 등의 패키지로 분리하여 각 도메인의 책임을 명확히 합니다.

    ```go
    // user/service/user_service.go
    package service

    type UserService struct {
        // 사용자 관련 비즈니스 로직
    }
    ```

- 정보 은닉 (Information Hiding):  패키지 내부 구현을 캡슐화하고 필요한 인터페이스만 외부에 노출

    패키지 내부 구현을 숨기고 필요한 인터페이스만 노출합니다.
    소문자로 시작하는 비공개 함수/변수를 사용하여 패키지 내부 구현을 숨깁니다.

    ```go
    // user/repository/user_repository.go
    package repository

    type UserRepository interface {
        FindByID(id int) (*User, error)
    }

    type userRepositoryImpl struct {
        // 비공개 구현
    }

    func NewUserRepository() UserRepository {
        return &userRepositoryImpl{}
    }
    ```

- 단일 책임 원칙 (Single Responsibility Principle): 각 패키지가 하나의 주요 책임 또는 관심사에 집중합니다.

    각 패키지와 그 내부 구조가 하나의 주요 책임에 집중합니다.
    `user` 패키지는 사용자 관리에만 집중하고, `product` 패키지는 제품 관리에만 집중합니다.

    ```go
    // user/handler/user_handler.go
    package handler

    type UserHandler struct {
        // 사용자 관련 HTTP 요청 처리만 담당
    }
    ```

- 의존성 역전 원칙 (Dependency Inversion Principle): 고수준 모듈(예: service)이 저수준 모듈(예: repository)에 의존하지 않도록 설계합니다.

    고수준 모듈(service)이 저수준 모듈(repository)에 *직접* 의존하지 않도록 인터페이스를 사용합니다.
    서비스 레이어가 리포지토리 인터페이스에 의존하도록 설계합니다.

    ```go
    // user/service/user_service.go
    package service

    type UserService struct {
        repo repository.UserRepository
    }

    func NewUserService(repo repository.UserRepository) *UserService {
        return &UserService{repo: repo}
    }
   ```

- 컨웨이의 법칙 (Conway's Law): 조직 구조가 소프트웨어 구조에 반영됩니다.

    "시스템을 설계하는 조직은 그 조직의 의사소통 구조와 일치하는 구조의 설계를 만들어낼 수밖에 없다."

    사용자 관리팀, 제품 관리팀, 주문 처리팀이 있다면, 이에 대응하는 `user`, `product`, `order` 패키지를 만듭니다.

- 도메인 주도 설계 (Domain-Driven Design): 비즈니스 도메인을 중심으로 소프트웨어를 구조화

    비즈니스 도메인을 중심으로 패키지를 구성합니다.
    각 패키지 내에 도메인 모델, 리포지토리, 서비스 등을 포함시켜 도메인 중심의 구조를 만듭니다.

    DDD는 광범위한 소프트웨어 설계 철학이며, 문제를 해결하기 위한 일종의 전략/전술입니다.
    따라서 DDD 기반으로 애플리케이션 개발시 일반적으로 layered architecture를 따르지만,
    사실 DDD 자체는 특정 아키텍처 스타일을 강제하지 않습니다.

    여기서 DDD에 근거한다는 것은 모든 package oriented design이 DDD을 따른다는 말이 아니고,
    DDD의 일부 원칙을 구현하는 한 방법으로 package oriented design을 사용함을 의미합니다:
    - 각 패키지는 하나의 바운디드 컨텍스트를 나타낼 수 있습니다.
    - 도메인 모델, 값 객체, 엔티티 등 DDD의 핵심 개념을 각 패키지 내에서 구현할 수 있습니다.

    ```go
    user/
    ├── model/
    │   └── user.go
    ├── repository/
    │   └── user_repository.go
    ├── service/
    │   └── user_service.go
    └── handler/
        └── user_handler.go
    ```

**장점**:
1. 높은 모듈성: 관련 기능이 하나의 패키지에 모여 있어 코드 이해와 유지보수가 용이합니다.
2. 명확한 책임 분리: 각 패키지가 특정 도메인이나 기능에 집중하여 책임이 명확합니다.
3. 병렬 개발 용이성: 팀원들이 서로 다른 패키지에서 독립적으로 작업할 수 있습니다.
4. 재사용성 향상: 패키지 단위로 다른 프로젝트에서 쉽게 재사용할 수 있습니다.
5. 확장성: 새로운 기능이나 도메인을 추가할 때 기존 구조를 크게 변경하지 않고 새 패키지를 추가할 수 있습니다.

**단점**:
1. 중복 코드 가능성: 유사한 기능이 다른 패키지에서 중복될 수 있습니다.
2. 패키지 간 의존성 관리의 어려움: 패키지가 많아질수록 의존성 관리가 복잡해질 수 있습니다.
3. 과도한 분리: 작은 프로젝트에서는 불필요하게 복잡한 구조가 될 수 있습니다.
4. 일관성 유지의 어려움: 각 패키지가 독립적으로 발전하면서 전체적인 일관성을 유지하기 어려울 수 있습니다.
5. 횡단 관심사 처리의 어려움: 로깅, 인증 등 여러 패키지에 걸친 기능을 구현하기 어려울 수 있습니다.

Package Oriented Design은 중간 규모 이상의 프로젝트나 마이크로서비스 아키텍처를 고려하는 경우에 특히 유용할 수 있습니다.

## 레이어 중심의 패키지 구조

애플리케이션의 각 계층을 별도의 패키지로 분리합니다.

```sh
project/
├── cmd/
│   ├── main.go
│   └── ...
├── model/
│   ├── user.go
│   ├── product.go
│   └── ...
├── repository/
│   ├── user_repository.go
│   ├── product_repository.go
│   └── ...
├── service/
│   ├── user_service.go
│   ├── product_service.go
│   └── ...
├── handler/
│   ├── user_handler.go
│   ├── product_handler.go
│   └── ...
└── utils/
    ├── logger/
    ├── config/
    └── ...
```

## 클린 아키텍처 기반 구조

의존성 규칙을 엄격히 적용하여 내부 로직과 외부 인터페이스를 분리합니다.

```sh
project/
├── domain/
├── usecase/
├── interface/
└── infrastructure/
```

## 모듈형 모노리스 구조

마이크로서비스의 장점을 살리면서도 모노리스 아키텍처의 단순성을 유지합니다.

```sh
project/
├── core/
├── moduleA/
├── moduleB/
└── shared/
```

## 기능과 레이어의 혼합 구조

기능별 패키지 내에서 레이어를 구분합니다.

```sh
project/
├── user/
│   ├── model/
│   ├── repository/
│   └── service/
├── product/
│   ├── model/
│   ├── repository/
│   └── service/
└── utils/
```

## DDD(Domain-Driven Design) 기반 구조

도메인 모델을 중심으로 구조를 설계합니다.

```sh
project/
├── domain/
├── application/
├── infrastructure/
└── interfaces/
```
