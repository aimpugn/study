# DTO, DAO, entity, model

- [DTO, DAO, entity, model](#dto-dao-entity-model)
    - [DTO (Data Transfer Object)](#dto-data-transfer-object)
    - [Entity](#entity)
    - [Model](#model)
    - [DAO (Data Access Object)](#dao-data-access-object)

## DTO (Data Transfer Object)

```go
type User struct {
    ID   int32          // users.id
    Code sql.NullString // users.code
}

func Test() {
   row := db.QueryRowContext(params...)

    var dto User
    err := row.Scan(
        &dto.ID,
        &dto.Code,
    )
}
```

DTO는 계층 간의 데이터 전송 또는 교환을 목적으로 사용되는 객체다.

위의 코드에서 `User` 구조체가 데이터베이스로부터 데이터를 읽어와 애플리케이션 내 다른 부분으로 전달하는 데 사용되므로 DTO라고 볼 수 있다.
DTO는 주로 *데이터의 전송*에 초점을 맞추며, 복잡한 비즈니스 로직을 포함하지 않는 것이 일반적이다.

## Entity

Entity는 주로 데이터베이스 테이블의 레코드와 직접적으로 매핑되어 데이터를 표현하는 객체를 말한다.
ORM(Object-Relational Mapping) 패턴에서 사용되는 용어로, `User` 구조체가 데이터베이스의 `users` 테이블과 매핑되어 있다면, entity라고 볼 수 있다.

## Model

```go
type User struct {
    ID   int32
    Code string
}

// IsValidCode는 User의 Code가 유효한지 여부를 검사하는 메서드입니다.
func (u *User) IsValidCode() bool {
    // 여기서는 단순화를 위해 Code가 비어 있지 않으면 유효하다고 가정합니다.
    return u.Code != ""
}
```

모델은 애플리케이션의 비즈니스 도메인을 표현하는 객체로, 데이터와 해당 데이터에 대한 비즈니스 로직을 포함할 수 있습니다. `User` 구조체가 애플리케이션의 비즈니스 로직이나 규칙을 반영하고, 데이터베이스의 사용자 테이블을 표현한다면, 이는 모델로 간주될 수 있습니다.

## DAO (Data Access Object)

```go
type UserDAO struct {
    db *sql.DB
}

// NewUserDAO는 새 UserDAO 인스턴스를 생성합니다.
func NewUserDAO(db *sql.DB) *UserDAO {
    return &UserDAO{db: db}
}

// GetUserByID는 주어진 ID로 사용자를 조회합니다.
func (dao *UserDAO) GetUserByID(id int32) (*User, error) {
    row := dao.db.QueryRow("SELECT id, code FROM users WHERE id = ?", id)

    var user User
    err := row.Scan(&user.ID, &user.Code)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 여기에 더 많은 CRUD 작업을 추가할 수 있습니다.
```

DAO는 데이터베이스나 다른 저장소의 데이터에 접근하기 위한 객체로, 데이터베이스 CRUD (생성, 읽기, 업데이트, 삭제) 작업을 캡슐화한다.
`User` 구조체 자체보다는, `User` 구조체를 사용하여 데이터베이스와 상호작용하는 메서드(예: `row.Scan`)를 포함한 클래스나 구조체가 DAO의 역할을 수행한다. 따라서 `User` 구조체는 DAO가 아니며, DAO는 `User` 구조체를 사용하는 메서드를 포함할 수 있는 더 큰 컨텍스트다.
