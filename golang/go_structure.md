# Package structure

- [Package structure](#package-structure)
    - [패키지 구조와 타입 정의 요약](#패키지-구조와-타입-정의-요약)
    - [package oriented design](#package-oriented-design)
    - [import cycle not allowed: import stack](#import-cycle-not-allowed-import-stack)
        - [해결 방안](#해결-방안)
    - [dao, model, dto](#dao-model-dto)
        - [Model 패키지](#model-패키지)
        - [DAO 패키지](#dao-패키지)
        - [DTO (Data Transfer Object) 패키지](#dto-data-transfer-object-패키지)
        - [관리 전략 및 컨벤션](#관리-전략-및-컨벤션)
        - [Example](#example)
    - [패키지 간의 의존성](#패키지-간의-의존성)
        - [의존성 방향](#의존성-방향)
        - [1. 단일 방향 의존성](#1-단일-방향-의존성)
        - [2. 공통 유틸리티 패키지](#2-공통-유틸리티-패키지)
        - [3. 순환 의존성](#3-순환-의존성)
        - [4. 인터페이스를 통한 의존성 역전](#4-인터페이스를-통한-의존성-역전)
    - [패키지 설계와 순환 참조](#패키지-설계와-순환-참조)
        - ["고수준(High-level)"과 "저수준(Low-level)" 모듈](#고수준high-level과-저수준low-level-모듈)
        - [부모/자식 패키지의 관계](#부모자식-패키지의-관계)
    - ["의존"한다는 것과 import 한다는 것](#의존한다는-것과-import-한다는-것)
        - [의존한다는 의미](#의존한다는-의미)
        - [DIP 관점에서의 의존성](#dip-관점에서의-의존성)
    - [패키지 설계 원칙](#패키지-설계-원칙)
    - [Examples](#examples)
        - [공통적으로 사용되는 `models` 패키지](#공통적으로-사용되는-models-패키지)
        - [서로가 서로를 사용하게 될 경우](#서로가-서로를-사용하게-될-경우)
        - [import repositories 사용 vs each/package/repository.go](#import-repositories-사용-vs-eachpackagerepositorygo)
            - [Layered Architecture](#layered-architecture)
            - [Package Oriented Design](#package-oriented-design-1)
            - [정리](#정리)
        - [의존 관계 및 방향성 설명](#의존-관계-및-방향성-설명)
            - [의존성 흐름](#의존성-흐름)
            - [왜 `repositories`는 `api`나 `utils`를 import하지 않아야 하는가?](#왜-repositories는-api나-utils를-import하지-않아야-하는가)
        - [인터페이스에 의존하는 경우에도 순환 참조 발생할 수 있는 경우](#인터페이스에-의존하는-경우에도-순환-참조-발생할-수-있는-경우)

## 패키지 구조와 타입 정의 요약

- Go는 간결함, 효율성, 명확한 의존성 관리를 강조한다
- **일방향 의존성 유지**: 패키지 간의 의존성은 항상 일방향으로 유지하여 순환 의존성을 방지한다.
- **상호 참조 방지**: 패키지가 서로를 참조하지 않도록 하여 결합도를 낮춘다.
- **타입 정의 위치**: 타입은 그 타입을 직접 사용하는 가장 가까운 곳에 정의한다.
- **의존성 역전 적용**: 고수준 모듈이 저수준 모듈에 의존하지 않도록, 구체적인 구현보다는 인터페이스나 추상 타입에 의존하여 결합도를 낮추고 유연성을 증가시킨다.
- **기능 분리**: 관련 기능과 타입을 기능적, 도메인적 측면에서 분리하여 조직함으로써 관리 용이성을 향상시킨다.
- **익스포트 최소화**: 필요한 최소한의 타입과 함수만을 외부에 노출하여 API의 안정성과 예측 가능성을 보장한다.
- 의존하지 말고 복사하라(Not dependent, but copying)
- 포함하지 말고 제공하가(Not contain, but provide)
- 필요로 하는 곳에서 타입을 정의하고, 데이터를 주입 받는다.

## [package oriented design](https://www.youtube.com/watch?v=spKM5CyBwJA)

패키지는 "방화벽"을 만드는 것과 같다.
- 다양한 프로그램 조각이 서로 분리
- 큰 팀이 큰 프로젝트에 작업
- 컴파일러와 언어 자체가 지원하고 도울 수 있음

언어 메커니즘
패키징은 지금껏 배웠던 다른 언어에서 소스 코드를 조직하는 방법과 직접적으로 충돌
여러 폴더에 클래스를 위치시킨다면 프로젝트가 커질수록 대규모로 사용하기 어려워진다.
우리가 하는 일은 API를 구성하는 것

소스 코드는 신경쓰지 말고 어떤 컴포넌트가 필요한지 생각해야 한다

패키징은 마이크로서비스 집합을 만드는 것과 같으며, 따라서 프로젝트 레벨에서, 소스 코드 레벨에서 대부분이 마이크로서비스
반면 다른 언어에서는 소스 트리가 하나의 모놀리틱 애플리케이션을 구성

모든 패키지는 first class. 유일한 계층 구조는 프로젝트의 소스 트리에서 정의한 계층 구조.
all packages are "first class," and the only hierarchy is what you define in the source tree for your project.

sub-package?
두 패키지는 상호 임포트할 수 없음
무슨 API를 개발중이며 API 간의 관계가 어떤지 먼저 생각을 해야 한다.

패턴과 프레임워크가 싫음... 이런 것들은 생각하지 않게 한다. design을 해야 한다.

사용성과 이식성. 패키지는 제공해야 하지, 포함하면 안된다.

http, io는 명확하다. 하지만 common 같은 것은 smell. 패키지 이름은 구체적이고 목적이 드러나야 한다.

패키지를 사용하는 사람의 관점에서 설계해야 한다.
올바른 API는 테스트 가능해야 한다? 아니다! 올바른 API는 사용 가능해야 한다.
따라서 API는 반드시 직관적이고 간단해야 한다.
사용하는 사람을 존중해야 한다.
변경사항이 있을 때 그 변경 사항이 cascading 되면 안되게 해야 한다.

에러 인터페이스를 사용. 인터페이스 레벨에서 에러를 핸들링하도록 해야 한다.

타입 단언을 수행하는 코드를 볼 때마다 it raises a flag
인터페이스가 있고, 디커플링 됐고, 구체적인 코드로 돌아가면, 계단식 변화에 대비할 수 있다.
you just set yourself up for a cascading change

패키지는 코드 베이스르 반드시 줄이고, 작제 만들고, 간단하게 만들어야 한다.

패키지는 합리적이고 실용적인 경우 의견 수렴을 줄여야 한다.
Packages must reduce taking on opinions when it's reasonable and practical.

패키지가 단일 종속 지점이 되어서는 안 된다.
packages must not become a single point of dependency.

가장 안 좋은 것은 패키지가 common types라고 불리는 것
다른 패키지에서도 이 타입이 필요할 것이라 생각해서 이를 공통 타입으로 만들게 된다.
근데 이렇게 하면 이를 깨뜨릴 수 없다.
you are not going to be able to break it up

다른 패키지에서 타입을 중복시켜라.
A 패키지가 A라는 목적이 있고, B 패키지가 B라는 목적이 있다면, 타입은 API에 대한 아티팩트.
that's the types are an artifact to the API, not the other way around.
따라서 어떤 타입은 중복을 해야 한다.

개발자가 Go 프로젝트 내에서 패키지가 속한 위치와 패키지가 준수해야 하는 디자인 가이드라인을 식별할 수 있.
Allows a developer to identify where a pacakge belongs inside a Go project and the design guidelines the package must respect.

log 패키지와 cfg 패키지는 서로 임포트하지 않는다. configuration이 필요하다면, 이 구성은 API의 일부가 되어야 한다

소스 트리를 사용해서 import 관계를 드러낸다

```text
github.com/servi-io/api
cmd/
    servi/
        cmdupdate/
        cmdquery/
        servi.go
    servid/
        handlers/
        routes/
        tests/
        servid.go
internal/ -> 특별한 수준의 컴파일러 보호. 다른 패키지가 internal 패키지를 import하지 않도록 한다.
    attachments/
    locations/
    orders/
        customers/
        items/
        tags/
        orders.go
    registrations/
    platform/
        crypto/
        mongo/
        json/
```

orders.go가 customers를 import한다
하지만 같은 레벨에서 패키지 간의 import는 허용되지 않는다.

handlers와 routes가 서로 import하지 않아야 한다.

attachments가 locations을 import하는 것은 big big flag
attachments가 locations을 필요로 한다면?
- attachments하위에 locations가 위치
- locations가 internal이 아닌 하나의 특정한 바이너리로 command package로 자유롭게 import

I'm always gonna question in an import to another package if you're doing it just to gain access to that package's types . That's a big red flag.

you better really be able to prove that we can't make a copy of the types you need and keep these two packages more portable

## import cycle not allowed: import stack

```log
import cycle not allowed: import stack: [users, mysql, common, payments, mysql]
```

이 에러 메시지는 Go에서 패키지 간 순환 의존성(circular dependency)이 발생했을 때 나타난다.
순환 의존성이란 두 개 이상의 패키지가 서로를 직접적이거나 간접적으로 임포트함으로써 생기는 의존성의 순환 말한다.

1. 패키지 `users` 가 패키지 `mysql` 를 임포트.
2. 패키지 `mysql` 가 패키지 `common` 를 임포트.
3. 패키지 `common` 가 패키지 `payments` 를 임포트합니다.
4. 패키지 `payments` 가 다시 패키지 `mysql` 를 임포트하려고 합니다.

여기서 문제는 패키지 `payments` 가 패키지 `mysql` 를 임포트하려고 시도하는 부분이다.
이로 인해 `mysql -> common -> payments -> mysql`와 같은 순환 의존성이 발생한다.
이러한 순환은 패키지 `mysql` 에서 시작하여 다시 `mysql`로 돌아오는 경로를 형성한다.

### 해결 방안

1. **인터페이스 도입**: 순환 의존성이 발생하는 패키지 중 하나에서 인터페이스를 정의하고, 해당 인터페이스를 다른 패키지에서 구현함으로써 의존성의 방향을 한쪽으로만 유지할 수 있습니다.

2. **공통 패키지 생성**: 순환 의존성을 유발하는 타입이나 함수를 별도의 공통 패키지로 분리합니다. 이렇게 함으로써 모든 관련 패키지가 공통 패키지만을 임포트하도록 할 수 있습니다.

3. **의존성 역전(Dependency Inversion) 적용**: 고수준 모듈이 저수준 모듈에 의존하지 않도록 하고, 둘 다 추상화에 의존하게 함으로써 의존성의 방향을 역전시킬 수 있습니다.

순환 의존성 문제를 해결하기 위해서는 코드 구조를 면밀히 분석하고, 설계를 재검토할 필요가 있습니다. 때로는 상당한 리팩토링이 필요할 수도 있으며, 이 과정에서 코드의 설계와 아키텍처에 대한 깊은 이해가 요구됩니다.

## dao, model, dto

- DAO는 데이터 액세스 로직과 데이터베이스와의 상호작용을 담당
- Model은 데이터의 구조(스키마)와 비즈니스 로직을 정의

### Model 패키지

Model 패키지는 애플리케이션에서 사용되는 데이터의 구조를 정의한다.
이는 보통 구조체(Struct)를 사용하여 데이터베이스 테이블의 스키마를 Go의 타입 시스템으로 매핑한다.
Model은 순수한 데이터 구조로만 구성될 수 있으며, 때로는 데이터를 조작하는 메서드(비즈니스 로직)를 포함할 수도 있다.

```go
package model

// User 구조체는 users 테이블의 스키마를 나타냅니다.
type User struct {
    ID        int
    Username  string
    Email     string
    CreatedAt time.Time
}
```

### DAO 패키지

DAO 패키지는 데이터베이스와의 상호작용을 담당한다.
이는 CRUD(Create, Read, Update, Delete) 작업과 같은 데이터 액세스 로직을 포함하며, Model 패키지에서 정의된 구조체를 사용하여 데이터를 조작한다.
DAO는 데이터베이스 연결과 쿼리 실행, 결과 처리 등을 담당하며, 데이터베이스와의 모든 상호작용을 캡슐화한다.

```go
package dao

import (
    "database/sql"
    "app/model" // model 패키지를 임포트
)

type UserDao struct {
    db *sql.DB
}

// NewUserDao는 UserDao의 새 인스턴스를 생성합니다.
func NewUserDao(db *sql.DB) *UserDao {
    return &UserDao{db: db}
}

// GetUserByID는 주어진 ID에 해당하는 사용자를 조회합니다.
func (dao *UserDao) GetUserByID(id int) (*model.User, error) {
    var user model.User
    err := dao.db.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### DTO (Data Transfer Object) 패키지

DTO는 주로 계층 간 데이터 전송에 사용되는 객체로, `sql.NullInt32,` `sql.NullString` 등을 사용하여 데이터베이스의 결과를 담을 수 있다. DTO는 데이터베이스의 로우 데이터를 애플리케이션 내부에서 사용하기 위한 더 고수준의 구조체로 변환하는 데 사용된다.

```go
// 예: DTO 구조체
package dto

type UserDTO struct {
    ID    sql.NullInt32
    Name  sql.NullString
    Email sql.NullString
}

// 예: Repository 인터페이스
package repository

import "context"

type UserRepository interface {
    GetUserByID(ctx context.Context, id int) (dto.UserDTO, error)
}

// 예: Model 구조체
package model

type User struct {
    ID    int
    Name  string
    Email string
}
```

### 관리 전략 및 컨벤션

- **책임의 분리**: 모델과 DAO를 분리함으로써 데이터베이스 구조와 비즈니스 로직 사이의 결합도를 낮춥니다. 이는 변경 사항이 한 부분에만 국한되도록 하여 유지보수성을 향상시킵니다.
- **재사용성 및 확장성**: 잘 설계된 DAO는 다양한 모델에 대해 재사용할 수 있으며, 애플리케이션의 데이터 액세스 계층을 쉽게 확장할 수 있게 해줍니다.
- **단위 테스트 용이성**: 분리된 구조는 모델이나 DAO를 독립적으로 테스트하기 쉽게 만들어 줍니다. 특히, DAO의 경우 데이터베이스와의 상호작용을 목(mock)을 사용하여 테스트할 수 있습니다.

### Example

```go
func (r *userRepositoryImpl) GetUserByAccessToken(ctx context.Context, accessToken string) (*model.V1User, error) {
    row := r.Stmt(userByAccessToken).
        QueryRowContext(ctx, accessToken, expiredInDateTimeFormat())

    if err := row.Err(); err != nil {
        return nil, errors.Wrap(err, errGetUserByAccessTokenQueryRowContext)
    }

    var dto dto.User // 계층 간 데이터 전송을 위해 사용되는 객체로, 비즈니스 로직을 포함하지 않고 단순히 데이터를 담는 컨테이너 역할
    err := row.Scan(
        &dto.ID,
        &dto.Code,
    )
```

## 패키지 간의 의존성

Go에서 *패키지 간 의존성*은 패키지가 다른 패키지의 타입, 변수, 상수, 함수 등을 사용할 때 발생한다.
Go의 패키지 시스템은 명시적인 의존성 관리를 통해 코드의 모듈성과 재사용성을 높이도록 설계되었다.

일반적으로, Go에서는 패키지 간의 의존성을 가능한 한 단방향으로 유지하는 것이 좋다.
Go에서는 한 패키지가 서로를 직접 또는 간접적으로 임포트하는 순환 참조를 허용하지 않는다.
Go 컴파일러는 순환 참조를 허용하지 않으며, 이는 코드의 의존성을 명확하게 관리하기 위함이다.

> 순환 참조: 한 패키지가 다른 패키지를 참조하고, 그 참조되는 패키지가 다시 첫 번째 패키지를 참조하는 상황

```text
proj/
  subproj1/
    queries
      request_query.go
  repositories/ (common pkg for several subproj)
    mysql/
      some_table_repo.go
```

따라서 부모 자식 pkg가 있다고 할 때 부모가 자식에 정의한 타입은 사용하지만 자식이 부모에 정의된 타입을 사용하는 것은 피해야 한다.
이러한 패키지 간의 상하 관계(부모/자식 관계)는 물리적인 디렉토리 구조에 의해 결정되는 것이 아니라, 패키지 간의 의존성과 참조 관계에 의해 결정된다.

근데 이렇게 `queries/request_query.go`의 부모인 `subproj1`와 같은 레벨에 위치하는 `repositories` 하위의 `mysql/request_query.go`을 사용하려고 하는 게 일반적일까?

여러분이 설명한 구조에서 `repositories/mysql/some_table_repo.go`가 `subproj1/queries/request_query.go`에 선언된 타입을 사용하는 것은 여러 가지 측면에서 고려해야 할 사항이 있습니다:

1. **의존성의 방향**

    일반적으로 공통된 라이브러리나 패키지(`repositories` 같은)는 상위 수준의 추상화를 제공하며,
    구체적인 애플리케이션 논리(`subproj1/queries`와 같은)에 의존해서는 안 된다.

2. **순환 참조 문제**

    `subproj1`이 `repositories`를 참조하고, `repositories`가 다시 `subproj1`을 참조한다면, 순환 참조가 발생할 수 있다. 이는 Go에서 컴파일 에러를 발생시킨다.

3. **디자인 원칙**:

    소프트웨어 설계의 일반적인 원칙 중 하나는 **의존성 역전 원칙(Dependency Inversion Principle, DIP)**이다.
    이 원칙에 따르면, **고수준 모듈은 저수준 모듈에 의존하지 않아야** 하며, 둘 모두 추상화에 의존해야 한다.
    이 원칙을 적용하면, `repositories`와 `queries` 사이에 추상화 계층(인터페이스 등)을 두어, 각각이 추상화에만 의존하도록 설계할 수 있다.

이에 대해 아래와 같은 대안이 있을 수 있다.

- **공통 인터페이스 정의**:

    `repositories`와 `subproj1/queries` 사이에 공통 인터페이스를 정의하여, 이 인터페이스를 통해 서로를 참조할 수 있도록 한다.
    이 방법은 DIP를 따르는 좋은 예시다.

- **DTO(Data Transfer Object) 패턴 사용**:

    공통 데이터 구조를 정의하는 별도의 패키지를 생성하고, 이를 `repositories`와 `subproj1/queries`에서 모두 임포트하여 사용한다.
    이렇게 하면 공통 데이터 타입을 공유하면서도 순환 참조를 피할 수 있다.

- **서비스 계층 도입**:

    비즈니스 로직과 데이터 액세스 로직 사이에 서비스 계층을 도입한다.
    서비스 계층은 데이터를 처리하고 변환하는 로직을 캡슐화하여, 데이터 액세스 계층(`repositories`)과 비즈니스 로직(`queries`) 사이의 의존성을 관리한다.

### 의존성 방향

소프트웨어 설계에서 일반적으로 권장되는 의존성 방향은 "고수준에서 저수준으로"다.
이는 고수준 모듈이 저수준 모듈에 의존해야 하며, 반대로 저수준 모듈이 고수준 모듈에 의존하는 것을 피해야 한다는 의미이다.
이 "고수준"과 "저수준"에서의 "고/저" 관계는 패키지가 위치하는 폴더 구조의 상하와는 다르며, *추상화의 수준*과 *의존성의 방향*에 따라 결정된다.

이 원칙을 의존성 역전 원칙(Dependency Inversion Principle, DIP)과 함께 사용하면, 고수준 모듈과 저수준 모듈 모두 추상화에 의존하게 되어, 모듈 간의 결합도를 낮추고 유연성을 높일 수 있다.

- **단방향 의존성**:
  
    패키지 간의 의존성이 단방향으로 유지되어야 한다.
    예를 들어, 공통적으로 사용되는 기능이나 타입은 공통 패키지에 정의하고, 이를 여러 패키지에서 참조하는 구조가 바람직하다.

- **의존성 역전 원칙(Dependency Inversion Principle)**

    고수준 모듈이 저수준 모듈에 의존하지 않도록, 추상화를 통해 의존성을 역전시키는 원칙을 적용할 수 있다.
    이를 통해 패키지 간의 결합도를 낮출 수 있다.

### 1. 단일 방향 의존성

가장 기본적인 형태는 한 패키지가 다른 패키지의 기능을 사용하는 경우다.
예를 들어, `net/http` 패키지의 기능을 사용하여 웹 서버를 구현하는 경우:

```go
package main

import (
    "net/http" // `main` 패키지는 `net/http` 패키지에 의존하고 있다
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    // `main` 패키지는 `http` 패키지의 `HandleFunc`과 `ListenAndServe` 함수를 사용하여 웹 서버를 구성한다.
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

### 2. 공통 유틸리티 패키지

```go
// 공통으로 사용되는 유틸리티 함수나 타입을 별도의 패키지로 분리
// `utils/strings.go`:
package utils

func Reverse(s string) string {
    // 문자열을 뒤집는 로직
}
```

다른 패키지에서 `utils` 패키지 사용하기:

```go
package main

import (
    "fmt"
    "myproject/utils"
)

func main() {
    fmt.Println(utils.Reverse("hello"))
}
```

### 3. 순환 의존성

Go에서는 순환 의존성을 허용하지 않는다.
따라서 아래처럼 패키지 A가 패키지 B를 import하고, 동시에 패키지 B가 패키지 A를 import하는 경우 컴파일 에러가 발생한다.

```go
// `packageA/a.go`:
package packageA

import "packageB"

func AFunc() {
    packageB.BFunc()
}
```

```go
// `packageB/b.go`:
package packageB

import "packageA"

func BFunc() {
    packageA.AFunc()
}
```

### 4. 인터페이스를 통한 의존성 역전

인터페이스를 사용하여 의존성을 관리할 수 있다.
예를 들어, 고수준 모듈이 인터페이스를 정의하고, 저수준 모듈이 이 인터페이스를 구현함으로써, 고수준 모듈이 저수준 모듈에 의존하지 않도록 할 수 있다.
이는 의존성 역전 원칙(Dependency Inversion Principle)의 핵심이다.

```go
// 고수준 모듈
// "데이터 저장"이라는 추상적인 작업을 정의하는 인터페이스(`Storer`)를 제공하기 때문
// 이 인터페이스는 실제 데이터 저장 방식(디스크, 메모리, 네트워크 등)에 대한 세부 사항을 추상화한다.

// `storage/interface.go`
package storage
type Storer interface {
    Save(data string) error
}
```

```go
//저수준 모듈
// 이 모듈은 `Storer` 인터페이스를 구현함으로써, 실제로 데이터를 디스크에 저장하는 구체적인 작업을 수행한다.
// `disk` 패키지(저수준)가 `storage` 패키지(고수준)에 정의된 `Storer` 인터페이스를 구현함으로써
// `disk` 패키지는 `storage` 패키지의 **추상화에 의존**하게 된다.
// 고수준 모듈(`storage`)이 저수준 모듈(`disk`)에 의존하지 않고, 대신 추상화(`Storer` 인터페이스)에 의존하므로, 의존성 역전 원칙(DIP)

// `disk/disk.go`
package disk

import "storage"

type DiskStorer struct{}

func (ds DiskStorer) Save(data string) error {
    // 데이터를 디스크에 저장하는 로직
    return nil
}
```

이런 방식으로, `disk` 패키지는 `storage` 패키지의 `Storer` 인터페이스를 구현함으로써, 두 패키지 간의 결합도를 낮추고 유연성을 높일 수 있다.

## 패키지 설계와 순환 참조

Go에서는 패키지 A가 패키지 B를 임포트(import)하고, 동시에 패키지 B가 패키지 A를 임포트하는 패키지 간 순환 참조를 허용하지 않는다.
이는 프로그램의 설계 오류를 나타내며, 컴파일 시점에 오류를 발생시킨다.

패키지 설계 시 순환 참조를 피하기 위해서는 의존성의 방향을 한쪽으로만 흐르도록 해야 한다.
보통 더 추상적인 개념(고수준 모듈)이 더 구체적인 구현(저수준 모듈)에 의존하지 않도록 구성한다.

### "고수준(High-level)"과 "저수준(Low-level)" 모듈

고수준 모듈과 저수준 모듈의 구분은 *추상화의 수준*에 따라 달라진다. 즉, 패키지가 서로 다른 추상화 수준에 있음을 나타낸다.
따라서 "고수준(High-level)"과 "저수준(Low-level)" 관계를 이해하는 데 있어 중요한 개념은 "추상화 수준"과 "의존성 방향"이다.
이 용어들은 *패키지가 다른 패키지에 제공하는 추상화의 정도와 의존성의 흐름을 기반*으로 한다.

- **고수준(High-level) 패키지**

    보통 더 높은 추상화 수준을 제공하는 패키지를 의미한다.
    보통 애플리케이션의 정책, 규칙, 비즈니스 로직 등 더 추상적인 개념을 다룬다.
    복잡한 로직이나 운영을 숨기고, 사용하기 쉬운 인터페이스를 제공하여 개발자가 세부적인 구현에 신경 쓰지 않고도 기능을 사용할 수 있도록 한다.
    고수준 패키지는 종종 다양한 저수준 구현을 추상화하여, 실제 구현 세부 사항에서 분리된다.

- **저수준(Low-level) 패키지**

    고수준 모듈에서 정의한 추상적인 개념을 구체적으로 구현한다.
    실제 작업을 수행하는 구체적인 구현을 제공하는 패키지를 말한다.
    이러한 패키지는 파일 시스템 접근, 네트워크 통신, 데이터 저장과 같은 구체적인 작업을 담당한다.
    저수준 패키지는 보통 한정된 기능에 집중하며, 특정 작업을 수행하기 위한 세부적인 로직을 포함한다.

### 부모/자식 패키지의 관계

Go에서 부모/자식 패키지의 개념은 파일 시스템의 디렉토리 구조에 더 가깝다.
예를 들어, `parent/child` 구조에서 `child`는 `parent`의 하위 디렉토리에 위치한다.
하지만, 이러한 **디렉토리 구조는 패키지 간 의존성과는 직접적인 관계가 없다**.

패키지 간 의존성은 `import` 문을 통해 결정된다.
부모 패키지가 자식 패키지를 임포트할 수도 있고, 그 반대의 경우도 가능하다.
중요한 것은 순환 참조를 피하고, 의존성의 방향이 명확하게 한 방향으로 흐르도록 하는 것이다.

## "의존"한다는 것과 import 한다는 것

### 의존한다는 의미

- **논리적 의존성**

    일반적으로 A 모듈이 B 모듈의 기능이나 데이터를 사용할 때, A 모듈은 B 모듈에 "의존"한다.
    이는 곧 *사용하는 모듈* A의 동작이 *사용 당하는 모듈* B에 정의된 기능이나 데이터에 의존적임을 의미한다.

- **물리적 의존성 (import)**

    프로그래밍 언어 차원에서의 "import"는 물리적 의존성을 만든다.
    즉, 한 파일이 다른 파일을 import 함으로써, 컴파일러에게 필요한 외부 코드의 위치를 알려주고, 해당 코드를 사용할 수 있게 된다.

### DIP 관점에서의 의존성

DIP에서는 고수준 모듈과 저수준 모듈 모두가 추상화에 의존해야 합니다. 이 경우:

- **저수준 모듈의 의존성**

    저수준 모듈이 고수준 모듈에 "의존"한다는 것은, 보통 저수준 모듈이 고수준 모듈에 정의된 인터페이스를 구현하거나 충족시킨다는 것을 의미한다.
    여기서 의존성의 방향은 고수준 모듈의 추상화된 인터페이스나 요구사항에 대한 것이다.

- **물리적 의존성과의 차이**

    저수준 모듈이 고수준 모듈을 `import`한다고 자동으로 고수준 모듈에 의존한다는 의미는 아니다.
    DIP에서 말하는 의존성은 추상화된 인터페이스나 계약에 대한 것이며, 이는 구현 세부사항이 아닌 추상화된 인터페이스에 대한 의존성을 의미한다.

## 패키지 설계 원칙

- **재사용성**

    `models`와 같은 패키지가 애플리케이션의 다른 부분에서 널리 사용되는 타입을 정의한다면, 이는 고수준의 재사용성을 가진다고 볼 수 있다.
    이러한 패키지는 일반적으로 다른 많은 패키지에 의해 참조된다

- **의존성의 최소화**

    가능한 한 패키지 간의 의존성을 최소화하는 것이 좋다.
    `models` 패키지가 다른 패키지를 임포트하지 않고, 독립적으로 존재한다면, 이는 의존성을 줄이는 좋은 예이다.

- **단방향 의존성**

    패키지 간 의존성은 가급적이면 단방향으로 유지하는 것이 좋다.
    즉, 순환 참조가 없어야 하며, 가능한 한 의존성의 방향이 명확해야 한다.

## Examples

### 공통적으로 사용되는 `models` 패키지

```text
service/
  models/
    my_model.go
  subpkg1/
    requests/
      some_requests.go
```

부모 패키지에서 자식 패키지에 정의된 것을 사용하고, 자식 패키지에서는 부모 패키지에 정의된 것을 사용하지 않는 방식은 Go에서 순환 참조를 피하기 위한 일반적인 규칙 중 하나다.

그러나, `models` 패키지를 다른 하위 패키지에서 사용하는 것이 이러한 방향성과 배치되는지 여부는 패키지의 역할과 구조에 따라 달라진다.
`models` 패키지가 도메인 모델이나 데이터 구조를 정의하는 곳이라면, 이는 애플리케이션의 여러 부분에서 공통적으로 사용될 수 있는, 재사용 가능한 코드의 집합이다. 이 경우 `models` 패키지는 애플리케이션의 핵심 구성 요소를 나타내며, 다른 "하위" 패키지들이 `models` 패키지를 사용하는 것은 자연스러운 구조다.

`models` 패키지와 같은 공통 패키지가 다른 패키지들에 의해 사용되는 것은 일반적인 패턴이며, 이는 부모와 자식 패키지 간의 일반적인 의존성 규칙과 필연적으로 배치되는 것은 아니다. 중요한 것은 패키지 간의 의존성이 명확하고, 순환 참조를 피하며, 코드의 재사용성과 유지보수성을 최대화하는 것이다.
`models` 패키지는 애플리케이션의 여러 부분에서 공통으로 사용되는 타입을 정의하기 때문에, 이를 다른 패키지에서 사용하는 것은 구조적으로 의미가 있으며, 패키지 설계 원칙에 부합한다.

### 서로가 서로를 사용하게 될 경우

```text
.
└── monorepo
    ├── service1
    │   ├── api
    │   │   ├── admin
    │   │   ├── boards
    │   │   ├── products
    │   │   └── users
    │   ├── models
    │   ├── repositories
    │   │   ├── grpc
    │   │   └── mysql
    │   └── utils
    ├── service2
    └── utils
```

- 모델의 분리
    - 데이터 모델을 중앙화하여 관리
    - 재사용성을 높임
    - 모델 변경 시 여러 서비스에 걸친 일관된 업데이트 가능
- 도메인별 api 구조
    - api 아래에 admin, boards, products, users 등 도메인별로 구분한 것은 DDD 원칙에 부합한다.
    - 각 기능별로 코드를 조직화하여 관리하기 용이하다.
- 레포지토리 패턴 사용
    - repositories 아래에 데이터 소스와의 상호 작용을 추상화하는 레포지토리 패턴을 적용한 것은 유연성과 테스트 용이성을 높여준다.

> `service/api/products/some_logic.go`에서 `service1/models/products.go` 모델을 import 해서 사용한다고 할 때, `service/api/` 패키지와 같은 레벨에 존재하는 `service1/models/` 패키지의 파일을 import하게 된다. 이처럼 `some_logic.go`가 `models` 패키지의 `products.go`에 의존하는 게 올바른 구조인가?

`service/api/products/some_logic.go`에서 `service1/models/products.go` 모델을 import하여 사용하는 것은 일반적으로 올바른 구조입니다. 이는 도메인 로직에서 사용되는 구조체를 명확히 정의하고, 이를 재사용함으로써 코드의 일관성과 유지보수성을 높이는 데 도움이 됩니다.

`some_logic.go`에서 `models/products.go`를 import하는 것은 반적으로 올바른 구조.
모델은 다양한 서비스 로직에서 사용될 수 있으며, 이렇게 재사용 가능한 컴포넌트로서 모델을 설계하는 것은 일반적인 접근 방식이다. 중요한 것은 모델이 비즈니스 로직으로부터 충분히 분리되어 있어야 한다는 것이다.

`api` 패키지 내의 도메인이 `models` 패키지에 의존하는 구조는 일반적으로 문제가 되지 않는다.
하지만, 반대로 `models` 패키지가 `api` 패키지에 의존하는 경우, 순환 참조 문제가 발생할 수 있으므로 패키지 간 의존성을 잘 관리해야 한다.

> 모델은 이 애플리케이션 내에서 서비스 로직에서 사용되는 구조체를 의미한다. 그런데 이 로직을 모두 마친 후에 클라이언트로 응답하는 구조체는 별도의 비즈니스 로직 없이 오직 컨테이너 역할만 하는 product_response.go를 정의했을 때, 이 product_response.go는 admin에서도 사용될 수 있다. 가령 상품을 등록한다고 했을 때 등록된 해당 상품 정보를 클라이언트에 반환할 때 product api에서 반환하는 것과 같은 포맷의 json으로 응답 받도록 재사용할 수 있다. 이런 경우 product_response.go는 어디에 위치해야 하나?

`product_response.go`는 여러 도메인에서 재사용될 수 있는 응답 구조체다.
이를 관리하는 가장 올바른 방법은, 이를 *공통으로 사용할 수 있는 패키지* 내에 위치시키는 것이다.
이를 위해, `service1` 바로 아래에 `responses` 패키지를 생성하고, 여기에 `product_response.go`를 위치시키는 것이 좋다. 이렇게 하면, `api/product` 패키지나 `api/admin` 패키지 등에서 쉽게 참조할 수 있으며, 순환 참조 문제를 피할 수 있다.

### import repositories 사용 vs each/package/repository.go

```text
api/
    product/ <- import repositories.ProductRepository
    users/ <- import repositories.UsersRepository
    admin/ <- import repositories.AdminRepository
    orders/ <- import repositories.OrdersRepository
repositories/
    product_repository.go
    users_repository.go
    admin_repository.go
    orders_repository.go
```

```text
api/
    product/
        repository.go
    users
        repository.go
    admin/
        repository.go
    orders/
        repository.go
```

#### Layered Architecture

Layered Architecture는 소프트웨어를 명확히 구분된 층(layer)으로 분리하여, 각 층이 자신의 역할에만 집중할 수 있도록 하는 구조다.
이 방식은 일반적으로 다음과 같은 장점을 제공한다.

- **명확한 역할 분리**: 각 레이어는 독립된 역할과 책임을 갖는다.
- **재사용성**: 공통 로직이나 모델을 여러 곳에서 재사용할 수 있다.
- **교체 용이성**: 특정 레이어의 구현을 변경해도 다른 레이어에 영향을 주지 않는다.

그러나, 공유 모델이 변경될 때 여러 레이어에 걸쳐 영향을 미치는 등의 단점도 있습니다.

#### Package Oriented Design

Package Oriented Design은 각 패키지를 독립적인 서비스처럼 취급하여, 내부 구현을 숨기고 필요한 인터페이스만 노출하는 방식이다.
이 접근 방식은 다음과 같은 장점을 갖는다.

- **높은 응집도와 낮은 결합도**: 각 패키지가 독립적으로 기능하기 때문에, 한 영역의 변경이 다른 영역에 미치는 영향이 적다.
- **변경 용이성**: 특정 도메인의 사양 변경이 다른 도메인에 영향을 미치지 않는다.
- **이해하기 쉬운 구조**: 각 패키지가 독립적인 서비스처럼 작동하기 때문에, 각각을 개별적으로 이해할 수 있다.

단점으로는, 코드 중복이 발생할 수 있고, 공통 기능의 중복 구현으로 인한 유지보수 비용이 증가할 수 있다.

#### 정리

Go 언어의 철학과 패키지 시스템은 Package Oriented Design과 잘 맞는다.
Go는 간결하고 읽기 쉬운 코드를 선호하며, 강력한 타입 시스템과 인터페이스를 통해 느슨한 결합과 응집력 있는 코드 구조를 장려한다.
또한, Go의 작은 인터페이스 원칙은 각 패키지가 필요한 최소한의 인터페이스만 노출하도록 격려한다.

장기적으로 보았을 때, 특히 Go 언어를 사용하는 경우, Package Oriented Design이 더 많은 이점을 제공할 수 있다.
이는 각 패키지의 독립성을 유지하고, 변화에 더 유연하게 대응할 수 있게 해주며, Go의 철학과도 잘 부합한다.
하지만, 코드 중복을 최소화하고 공통 기능의 재사용을 극대화하기 위해 일부 공통 인터페이스나 유틸리티 함수를 공유할 수 있는 방안을 고려하는 것도 중요하다.

### 의존 관계 및 방향성 설명

```text
~/tmp/go_pkg_test/monorepo/service1                                                                                        at 01:03:41
❯ tree . -d
.
├── api
│   ├── admin
│   ├── boards
│   ├── products
│   ├── requests
│   ├── responses
│   └── users
├── models
├── repositories
│   ├── grpc
│   └── mysql
└── utils
```

Go는 간결성, 명확성, 그리고 패키지 간의 느슨한 결합을 중시한다.
이를 위해 의존성은 한 방향으로 흐르도록 설계해야 한다.

#### 의존성 흐름

1. **`models` 패키지**:
    `models`는 데이터 구조를 정의하는 패키지로, 비즈니스 로직이나 데이터 접근 로직에서 사용된다.

    `models`는 다른 패키지를 import하지 않아야 한다.
    이는 모델이 최하위 레벨에 위치해야 함을 의미하며, 순환 참조의 위험을 최소화한다.

2. **`repositories` 패키지 (`grpc`, `mysql` 하위 포함)**:
    `repositories`는 데이터 소스와의 상호작용을 추상화 하므로, `models`를 import하여 데이터 소스와 모델 간의 매핑을 담당할 수 있다.

    `repositories`는 `api`나 `utils`를 import하지 않아야 한다.
    데이터 접근 로직이 상위 레벨의 API 로직에 의존하지 않도록 해야 한다.

3. **`utils` 패키지**:
    `utils`는 애플리케이션 전반에서 재사용될 수 있는 범용적인 유틸리티 함수나 컴포넌트를 제공한다.

    다른 패키지에 의해 널리 사용될 수 있으나, `utils` 자체는 가능한 한 의존성을 최소화해야 한다.
    유틸리티 기능이 특정 비즈니스 로직이나 데이터 모델에 종속되지 않도록 다른 패키지에 의존하지 않도록 한다.

4. **`api` 패키지 (`admin`, `boards`, `products`, `requests`, `responses`, `users` 하위 포함)**:
    `api`와 그 하위 패키지는 비즈니스 로직과 클라이언트 요청 처리를 담당한다.
    `api` 패키지는 `models`, `repositories`, 그리고 필요한 경우 `utils`를 import할 수 있다.
    하지만, `api` 내부의 패키지 간에는 순환 참조가 없도록 주의해야 한다.

    `requests`와 `responses`는 데이터 교환 포맷을 정의하므로, 다른 `api` 하위 패키지나 `models` 패키지와 독립적일 수 있다. 그러나 `models`에서 `responses`나 `requests`를 import하는 것은 피해야 한다.
    데이터 변환 로직은 `api` 패키지 내에서 처리하는 것이 좋다.

- **허용되는 Import**:
    - `repositories` -> `models`
    - `api` -> `models`, `repositories`, `utils`, `requests`, `responses`
    - `api/admin`, `api/boards`, 등 -> `models`, `repositories`, `utils`, `requests`, `responses`

- **피해야 하는 Import**:
    - `models` -> 다른 어떤 패키지도 import하지 않음
    - `repositories` -> `api`, `utils`
    - `utils` -> `api`, `models`, `repositories` (가능한 한 의존성을 피함)
    - `requests`, `responses` -> `models` (변환 로직은 `api`에서 처리)

#### 왜 `repositories`는 `api`나 `utils`를 import하지 않아야 하는가?

`repositories` 패키지가 `api`나 `utils` 패키지를 import하지 않아야 하는 이유는 몇 가지 설계 원칙과 구조적 측면에서 기인합니다. 아래에서 이러한 이유들을 자세히 살펴보겠습니다:

- **책임의 분리(Separation of Concerns, SoC):**

    `repositories` 패키지는 데이터 소스와의 상호작용을 담당하는 데이터 접근 계층(Data Access Layer, DAL)의 역할을 수행한다. 이는 데이터베이스 쿼리, 외부 API 호출 등 데이터 저장소와의 모든 통신을 캡슐화한다.

    `api` 패키지는 비즈니스 로직과 클라이언트 요청 처리를 담당한다.
    이는 사용자 요청을 받아 적절한 로직을 실행하고 응답을 반환하는 레이어다.

    이 두 영역은 각각 다른 책임을 지닌다. 따라서 각 계층의 독립성을 유지하고, 변경에 대한 유연성을 높이기 위해 `repositories`가 `api`를 직접 알거나 의존하지 않도록 하는 것이 중요하다.

- **의존성 역전 원칙(Dependency Inversion Principle, DIP):**

    의존성 역전 원칙은 고수준 모듈이 저수준 모듈에 의존해서는 안 되며, 둘 다 추상화에 의존해야 한다.

    `api` 계층(고수준)은 `repositories` 계층(저수준)의 구현에 직접 의존하지 않으며, 대신 인터페이스나 추상화된 계약을 통해 상호작용한다.
    이렇게 함으로써, 데이터 저장 방식의 변경이나 `repositories`의 내부 구현 변경이 `api` 계층에 영향을 미치지 않도록 보장할 수 있다.

- **유지보수성과 테스트 용이성:**

    `repositories`가 `api`나 `utils`에 의존하게 되면, 데이터 접근 로직이 비즈니스 로직이나 범용 유틸리티 함수의 변경에 영향을 받게 된다. 이는 유지보수와 테스트를 복잡하게 만들 수 있다.

    예를 들어, `utils` 내의 한 함수가 변경되었을 때, `repositories` 또한 영향을 받게 되며, 이는 예상치 못한 부작용을 초래할 수 있다. 따라서, 각 계층이나 영역의 의존성을 최소화하는 것이 안정성을 높이는 데 중요하다.

- **결합도(Coupling) 감소:**

    `repositories`가 `api`나 `utils`를 직접 import하게 되면, 시스템의 결합도가 증가한다.
    결합도가 높은 시스템은 한 부분의 변경이 전체 시스템에 파급 효과를 미칠 가능성이 높으며, 이는 유지보수와 확장성에 부정적인 영향을 미친다.

    시스템의 결합도를 낮추고 응집력을 높이는 것은 장기적으로 시스템의 안정성과 유지보수성을 보장하는 핵심 요소다.

### 인터페이스에 의존하는 경우에도 순환 참조 발생할 수 있는 경우

models 패키지에 SomeType을 정의했을 때

```go
// models/some.go
type SomeType interface {
  ToMySQLField() string
  ToOpenSearchField() string
}
```

이 도메인을 presentation 레이어에서

```go
// presentation/some.go
type SomeString string
func (r SomeString) ToMySQLField() string {
   return mysql.FieldA
}
func (r SomeString) ToOpenSearchField() string {
   return opensearch.FieldA
}

const (
  SomeStringA SomeString = "A"
  SomeStringB SomeString = "B"
)
```

그리고 models/some.go 파일을 repositories에서 import합니다.

```go
// infra/mysql/repsotiroy.go
import "models/some.go"

const (
  FieldA = "tbl.A"
)

func ExportedFunc(someType models.SomeType) {

}
```

```go
// infra/mysql/repsotiroy.go
import "opensearch/some.go"

const (
  FieldA = "data.A"
)

func ExportedFunc(someType models.SomeType) {

}
```

그리고 이에 대한 테스트 코드를 작성합니다.

```go
// repositories/mysql/repsotiroy_test.go
import "presentation/some.go"

func TestExportedFunc(t *testing.T) {
    // 테스트로 presenation.SomeStringA가 넘어오는 경우 테스트 위해 
    // presentation 패키지를 import
    input := presenation.SomeStringA
}
```

이런 식으로 들어오는 요청이 models.SomeType을 구현하게 하되, 변환하는 로직은 mysql 패키지나 opensearch에 정의된 상수를 사용해서 각 데이터소스에 맞는 필드로 변환하도록 했습니다.

그런데 이러면 presentation 패키지가 mysql 패키지를 사용하고, mysql 패키지에서 input으로 presenation.SomeStringA가 들어오는 경우를 테스트 하기 위해 presenation.SomeStringA를 사용하면 presentation 이 mysql을 import하고  mysql이 presentation을 import해서 순환 참조가 발생합니다.

이런 경우에는 presenation.SomeStringA를 가져와서 테스트 하는 게 아니라, 테스트용으로 다시 구현해야 하나요? 아니면 패키지 설계가 잘못된 걸까요? 그것도 아니라면 이런 경우에는 보통 어떻게 해야 하는 게 가장 올바른 방법일까요?
