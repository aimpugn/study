# Model

## model이란?

### Model

모델은 어플리케이션의 도메인(domain) 또는 비즈니스 로직을 표현하는 개념이다.
모델은 데이터의 구조뿐만 아니라, 그 데이터에 적용할 수 있는 행위나 규칙, 비즈니스 로직 등을 포함한다.
모델은 소프트웨어가 해결하려는 문제의 실제 세계적인 개념을 반영한다.

예를 들어, 은행 어플리케이션에서 계좌(Account) 모델은 계좌의 번호, 소유자, 잔액 등의 데이터를 저장할 뿐만 아니라, 입금(deposit)이나 출금(withdraw) 같은 행위도 정의할 수 있다.

## DTO와 Model의 역할 비교

- **DTO**: 데이터 전송에 초점을 맞추며, 계층 간의 데이터 교환을 위해 사용됩니다. DTO는 로직을 포함하지 않고, 데이터 구조와 그 데이터를 전달하는 역할만 수행합니다.

- **Model**: 어플리케이션의 비즈니스 로직과 데이터의 구조를 정의합니다. 모델은 실제 세계의 엔티티를 소프트웨어 내에서 추상화하고, 해당 엔티티의 데이터와 행위를 포함합니다.

## `model/doc.go`

```go
// Package model contains the core domain models used throughout the application.
// These models encapsulate the core business logic, data structures, and
// functions necessary for the application's operation. The models defined
// here are used to represent and manipulate the application's core data entities
// and their relationships.
//
// Overview:
// The model package is central to the application's domain-driven design,
// providing a clear separation between the application's core business logic
// and other layers such as the presentation and data access layers. This
// approach facilitates maintainability and scalability by isolating the
// business logic in a single, cohesive package.
//
// Entities vs. Models:
// Within this package, entities refer to objects that have a distinct identity
// and lifecycle, typically mirroring the application's database entities. Models,
// on the other hand, may encompass a wider range of objects including value
// objects, aggregates, and domain services that perform operations on or between
// entities.
//
// Usage:
// The models defined in this package should be instantiated and manipulated
// through the application's service layer or directly within the domain layer,
// depending on the architecture. It's recommended to use repositories for
// data persistence operations to maintain a clean separation of concerns.
//
// This package also includes utilities and functions that support the manipulation
// and validation of model instances, ensuring that the application's data integrity
// and business rules are enforced consistently.
//
// Note:
// This package's design is inspired by the principles of Domain-Driven Design (DDD),
// aiming to place the application's core business concepts at the forefront of development.
package model
```
