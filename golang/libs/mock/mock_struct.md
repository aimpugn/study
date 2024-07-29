# Mock struct

- [Mock struct](#mock-struct)
    - [struct 사용시 테스트를 위한 모의 객체 생성 방법](#struct-사용시-테스트를-위한-모의-객체-생성-방법)
        - [테스트용 구현체 생성](#테스트용-구현체-생성)
        - [의존성 주입 활용](#의존성-주입-활용)
        - [테스트 더블(Test Double) 패턴 사용](#테스트-더블test-double-패턴-사용)
        - [함수형 옵션 패턴](#함수형-옵션-패턴)
        - [내장 필드 활용](#내장-필드-활용)
        - [테스트용 서브패키지 생성](#테스트용-서브패키지-생성)
    - [순수한 구조체 하나만 유지할 경우 테스트 인스턴스](#순수한-구조체-하나만-유지할-경우-테스트-인스턴스)

## struct 사용시 테스트를 위한 모의 객체 생성 방법

struct를 사용하면서 테스트를 위한 모의 객체를 생성하는 방법들은 주로 외부 도구나 라이브러리에 의존하지 않고 순수한 Go 코드로 구현하는 접근 방식입니다.
이러한 방법들의 주요 목적과 특징은 다음과 같습니다:

1. [테스트 더블](../../../testing/testing_double.md)의 원리 이해: 개발자가 테스트 더블의 작동 원리를 깊이 이해하고 필요에 따라 커스터마이즈할 수 있도록 합니다.
2. 의존성 최소화: 외부 라이브러리나 패키지에 의존하지 않아 프로젝트의 유지보수성을 높입니다.
3. 기본 개념 학습: 모킹의 기본 개념과 다양한 구현 방법을 직접 경험하며 학습할 수 있습니다.
4. 유연성 확보: 특정 도구나 패키지에 종속되지 않아 다양한 상황과 요구사항에 맞게 유연하게 적용할 수 있습니다.
5. 깊이 있는 이해: 기본적인 모킹 기법을 직접 구현함으로써 모킹의 원리와 테스트 더블 패턴에 대한 깊은 이해를 얻을 수 있습니다.

주요 구현 방법:

1. [테스트용 구현체 생성](#테스트용-구현체-생성):

    실제 구현체와 동일한 구조를 가진 테스트용 구조체를 생성합니다.
    실제 구조체와 유사한 구조를 유지하면서 테스트에 필요한 동작을 구현할 수 있습니다.

2. [의존성 주입 활용](#의존성-주입-활용):

    구조체의 필드로 의존성을 주입받아 테스트 시 모의 객체를 주입합니다.
    실제 구현체를 변경하지 않고 테스트할 수 있으며, 의존성을 쉽게 교체할 수 있습니다.

3. [테스트 더블(Test Double) 패턴 사용](#테스트-더블test-double-패턴-사용):

    Stub, Spy, Fake 객체 등을 사용하여 실제 객체를 대체합니다.
    다양한 시나리오를 쉽게 테스트할 수 있고, 실제 객체의 복잡한 동작을 단순화할 수 있습니다.

4. [함수형 옵션 패턴](#함수형-옵션-패턴):

    구조체 생성 시 함수형 옵션을 사용하여 테스트에 필요한 동작을 주입합니다.
    매우 유연한 테스트 구성이 가능하며, 실제 구현체를 수정하지 않고 특정 동작만 변경할 수 있습니다.

5. [내장 필드 활용](#내장-필드-활용):

    Go의 내장 필드(embedded field) 기능을 활용하여 테스트용 구조체를 만듭니다.
    실제 구현체의 일부 메서드만 오버라이드하여 사용할 수 있어 유연성이 높습니다.

6. [테스트용 서브패키지 생성](#테스트용-서브패키지-생성):

    테스트용 서브패키지를 생성하여 테스트에 필요한 모의 객체들을 관리합니다.
    테스트 관련 코드를 명확하게 분리할 수 있고, 여러 테스트에서 재사용할 수 있는 모의 객체를 쉽게 관리할 수 있습니다.

### 테스트용 구현체 생성

실제 구현체와 동일한 구조를 가진 테스트용 구조체를 생성하여 사용합니다.

```go
// 실제 구현
type UserService struct {
    DB *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
    // 실제 데이터베이스 조회 로직
}

// 테스트용 구현
type TestUserService struct {
    Users map[int]*User
}

func (s *TestUserService) GetUser(id int) (*User, error) {
    user, exists := s.Users[id]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}

// 테스트 코드
func TestSomeFunction(t *testing.T) {
    testService := &TestUserService{
        Users: map[int]*User{
            1: {ID: 1, Name: "Test User"},
        },
    }
    
    // testService를 사용하여 테스트 수행
}
```

장점:
- 실제 구현체와 유사한 구조를 유지하면서 테스트에 필요한 동작을 쉽게 구현할 수 있습니다.
- 테스트 코드가 직관적이고 이해하기 쉽습니다.

단점:
- 실제 구조체와 테스트용 구조체를 동기화해야 하는 부담이 있습니다.
- 구조체의 메서드가 많을 경우 모든 메서드를 구현해야 할 수 있습니다.

### 의존성 주입 활용

구조체의 필드로 의존성을 주입받아 테스트 시 모의 객체를 주입합니다.

```go
type UserRepository interface {
    GetUser(id int) (*User, error)
}

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// 테스트 코드
func TestUserService(t *testing.T) {
    mockRepo := &MockUserRepository{
        Users: map[int]*User{
            1: {ID: 1, Name: "Test User"},
        },
    }
    service := NewUserService(mockRepo)
    
    // service를 사용하여 테스트 수행
}

type MockUserRepository struct {
    Users map[int]*User
}

func (m *MockUserRepository) GetUser(id int) (*User, error) {
    user, exists := m.Users[id]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}
```

장점:
- 의존성을 쉽게 교체할 수 있어 유연한 테스트가 가능합니다.
- 실제 구현체를 변경하지 않고 테스트할 수 있습니다.

단점:
- 인터페이스를 정의해야 하므로 코드량이 약간 증가할 수 있습니다.
- 구조체의 모든 의존성에 대해 이 방식을 적용해야 할 수 있습니다.

### 테스트 더블(Test Double) 패턴 사용

테스트 더블 패턴을 사용하여 실제 객체를 대체합니다. 주로 Stub, Spy, Fake 객체를 활용합니다.

```go
type UserService struct {
    DB *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
    // 실제 구현
}

// Stub 구현
type StubUserService struct {
    UserService
    StubUser *User
    StubErr  error
}

func (s *StubUserService) GetUser(id int) (*User, error) {
    return s.StubUser, s.StubErr
}

// 테스트 코드
func TestSomeFunction(t *testing.T) {
    stub := &StubUserService{
        StubUser: &User{ID: 1, Name: "Test User"},
        StubErr:  nil,
    }
    
    // stub을 사용하여 테스트 수행
}
```

장점:
- 다양한 시나리오를 쉽게 테스트할 수 있습니다.
- 실제 객체의 복잡한 동작을 단순화하여 테스트할 수 있습니다.

단점:
- 테스트 더블 객체를 별도로 구현해야 합니다.
- 실제 객체와 테스트 더블 간의 동기화를 유지해야 합니다.

### 함수형 옵션 패턴

구조체 생성 시 함수형 옵션을 사용하여 테스트에 필요한 동작을 주입합니다.

```go
type UserService struct {
    db        *sql.DB
    getUser   func(id int) (*User, error)
}

func NewUserService(db *sql.DB, opts ...func(*UserService)) *UserService {
    us := &UserService{
        db: db,
        getUser: func(id int) (*User, error) {
            // 기본 구현
        },
    }
    for _, opt := range opts {
        opt(us)
    }
    return us
}

// 테스트 코드
func TestUserService(t *testing.T) {
    testUser := &User{ID: 1, Name: "Test User"}
    us := NewUserService(nil, func(us *UserService) {
        us.getUser = func(id int) (*User, error) {
            return testUser, nil
        }
    })
    
    // us를 사용하여 테스트 수행
}
```

장점:
- 매우 유연한 테스트 구성이 가능합니다.
- 실제 구현체를 수정하지 않고도 특정 동작만 변경할 수 있습니다.

단점:
- 함수형 프로그래밍에 익숙하지 않은 개발자에게는 복잡해 보일 수 있습니다.
- 남용하면 코드의 가독성이 떨어질 수 있습니다.

### 내장 필드 활용

Go의 내장 필드(embedded field) 기능을 활용하여 테스트용 구조체를 만듭니다.

```go
type UserService struct {
    DB *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
    // 실제 구현
}

// 테스트용 구조체
type TestUserService struct {
    UserService
    MockGetUser func(id int) (*User, error)
}

func (s *TestUserService) GetUser(id int) (*User, error) {
    if s.MockGetUser != nil {
        return s.MockGetUser(id)
    }
    return s.UserService.GetUser(id)
}

// 테스트 코드
func TestSomeFunction(t *testing.T) {
    testService := &TestUserService{
        MockGetUser: func(id int) (*User, error) {
            return &User{ID: id, Name: "Test User"}, nil
        },
    }
    
    // testService를 사용하여 테스트 수행
}
```

장점:
- 실제 구현체의 일부 메서드만 오버라이드하여 사용할 수 있습니다.
- 기존 구조체의 기능을 그대로 유지하면서 테스트에 필요한 부분만 수정할 수 있습니다.

단점:
- 내장 필드의 개념을 이해해야 합니다.
- 복잡한 구조의 경우 관리가 어려울 수 있습니다.

### 테스트용 서브패키지 생성

테스트용 서브패키지를 생성하여 테스트에 필요한 모의 객체들을 관리합니다.

```go
// user/service.go
package user

type UserService struct {
    DB *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
    // 실제 구현
}

// user/testing/mock_service.go
package testing

import "path/to/user"

type MockUserService struct {
    user.UserService
    MockGetUser func(id int) (*user.User, error)
}

func (s *MockUserService) GetUser(id int) (*user.User, error) {
    if s.MockGetUser != nil {
        return s.MockGetUser(id)
    }
    return s.UserService.GetUser(id)
}

// 테스트 코드 (다른 패키지에서)
import "path/to/user/testing"

func TestSomeFunction(t *testing.T) {
    mockService := &testing.MockUserService{
        MockGetUser: func(id int) (*user.User, error) {
            return &user.User{ID: id, Name: "Test User"}, nil
        },
    }
    
    // mockService를 사용하여 테스트 수행
}
```

장점:
- 테스트 관련 코드를 명확하게 분리할 수 있습니다.
- 여러 테스트에서 재사용할 수 있는 모의 객체를 쉽게 관리할 수 있습니다.

단점:
- 추가적인 패키지 구조가 필요합니다.
- 실제 구현체와 모의 객체 간의 동기화를 주의깊게 관리해야 합니다.

## 순수한 구조체 하나만 유지할 경우 테스트 인스턴스

단순히 테스트를 위해 불필요한 인터페이스를 만드는 것은 좋은 설계 방식이 아닙니다.
구조체 하나만으로 충분한 상황에서 순수성을 유지하면서도 테스트 가능성을 높이는 방법들이 있습니다.

1. 의존성 주입 패턴 사용:
   구조체가 직접 DB와 통신하는 대신, DB 연결이나 쿼리 실행 함수를 매개변수로 받도록 설계합니다.

   ```go
   type UserService struct {
       db DBExecutor
   }

   type DBExecutor interface {
       Execute(query string, args ...interface{}) (interface{}, error)
   }

   func NewUserService(db DBExecutor) *UserService {
       return &UserService{db: db}
   }

   func (s *UserService) GetUser(id int) (*User, error) {
       user, err := s.db.Execute("SELECT * FROM users WHERE id = ?", id)
       // ... 결과 처리
   }

   // 실제 사용 시
   realDB := &RealDBExecutor{}
   service := NewUserService(realDB)

   // 테스트 시
   mockDB := &MockDBExecutor{}
   testService := NewUserService(mockDB)
   ```

   이 방식은 구조체의 순수성을 유지하면서도 테스트 시 DB 통신을 모의할 수 있게 합니다.

2. 기능적 옵션 패턴:
   구조체 생성 시 옵션 함수를 통해 내부 동작을 커스터마이즈할 수 있게 합니다.

   ```go
   type UserService struct {
       getUser func(id int) (*User, error)
   }

   func NewUserService(opts ...func(*UserService)) *UserService {
       us := &UserService{
           getUser: defaultGetUser,
       }
       for _, opt := range opts {
           opt(us)
       }
       return us
   }

   func defaultGetUser(id int) (*User, error) {
       // 실제 DB 쿼리 로직
   }

   // 테스트 시
   mockGetUser := func(id int) (*User, error) {
       return &User{ID: id, Name: "Test User"}, nil
   }
   testService := NewUserService(func(us *UserService) {
       us.getUser = mockGetUser
   })
   ```

   이 방식은 구조체의 기본 동작을 유지하면서 테스트 시 필요한 부분만 교체할 수 있습니다.

3. 내부 상태를 통한 테스트 모드:
   구조체 내부에 테스트 모드 플래그를 두어, 테스트 시 DB 통신을 우회합니다.

   ```go
   type UserService struct {
       db           *sql.DB
       testMode     bool
       testUserData map[int]*User
   }

   func (s *UserService) GetUser(id int) (*User, error) {
       if s.testMode {
           user, exists := s.testUserData[id]
           if !exists {
               return nil, errors.New("user not found")
           }
           return user, nil
       }
       // 실제 DB 쿼리 로직
   }

   // 테스트 시
   testService := &UserService{testMode: true, testUserData: map[int]*User{
       1: {ID: 1, Name: "Test User"},
   }}
   ```

   이 방식은 구조체 내부에 테스트 로직이 포함되지만, 실제 운영 코드와 테스트 코드를 명확히 분리할 수 있습니다.

4. 테스트 전용 빌더 패턴:
   테스트를 위한 특별한 빌더 함수를 제공하여 테스트에 적합한 구조체 인스턴스를 생성합니다.

   ```go
   func NewUserServiceForTesting() *UserService {
       return &UserService{
           db: &mockDB{},
       }
   }

   type mockDB struct{}

   func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
       // 모의 쿼리 결과 반환
   }
   ```

   이 방식은 테스트 코드와 실제 코드를 완전히 분리할 수 있지만, 테스트 전용 코드가 실제 코드베이스에 포함된다는 단점이 있습니다.

5. 테스트 시 수행되는 조건부 로직:
   구조체 메서드 내에서 테스트 환경을 감지하여 다르게 동작하도록 합니다.

   ```go
   func (s *UserService) GetUser(id int) (*User, error) {
       if os.Getenv("TEST_MODE") == "true" {
           // 테스트용 로직
           return &User{ID: id, Name: "Test User"}, nil
       }
       // 실제 DB 쿼리 로직
   }
   ```

   이 방식은 구현이 간단하지만, 실제 코드에 테스트 로직이 섞인다는 단점이 있습니다.
