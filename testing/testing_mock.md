# Testing mock

- [Testing mock](#testing-mock)
    - [mocking?](#mocking)
    - [Mock 객체 생성의 내부 동작](#mock-객체-생성의-내부-동작)
        - [1. 런타임 코드 생성](#1-런타임-코드-생성)
        - [2. 리플렉션 (Reflection)](#2-리플렉션-reflection)
        - [3. 프록시 패턴](#3-프록시-패턴)
        - [4. 메소드 인터셉션](#4-메소드-인터셉션)
        - [실제 구현 예시 (Java)](#실제-구현-예시-java)
    - [모킹 없는 테스트](#모킹-없는-테스트)

## mocking?

`FakeRepository`는 엄밀히 말하면 "mocking"의 한 형태입니다. 다만, 일반적으로 우리가 얘기하는 "mocking 라이브러리를 사용한 동적 mock 객체 생성"과는 다른 접근 방식입니다. 테스트 더블(Test Double)의 종류를 살펴보면 이해가 더 쉬울 것 같습니다:

1. Dummy: 아무 기능도 하지 않는 객체. 단순히 인터페이스를 만족시키기 위해 사용됩니다.

2. Stub: 미리 준비된 응답을 제공하는 객체. 주로 상태 기반 테스트에 사용됩니다.

3. Spy: 자신에게 어떤 메서드가 호출되었는지 기록하는 객체. 행위 검증에 사용됩니다.

4. Mock: 기대되는 호출을 명시하고, 그에 따라 동작하도록 프로그래밍된 객체. 주로 행위 검증에 사용됩니다.

5. Fake: 실제 구현의 단순화된 버전. 실제 객체와 유사하게 동작하지만, 프로덕션에는 적합하지 않은 구현입니다.

이 분류에 따르면, 예제의 `FakeRepository`는 "Fake" 카테고리에 해당합니다. 이는 mocking의 한 형태이지만, 일반적으로 "mocking"이라고 할 때 떠올리는 동적으로 생성되는 mock 객체와는 다릅니다.

주요 차이점:

1. 구현 방식:
   - `FakeRepository`는 수동으로 구현된 클래스입니다.
   - 일반적인 mocking 라이브러리를 사용한 mock은 런타임에 동적으로 생성됩니다.

2. 유연성:
   - `FakeRepository`는 고정된 동작을 가집니다. 변경하려면 코드를 수정해야 합니다.
   - 동적 mock은 테스트 케이스마다 다른 동작을 쉽게 정의할 수 있습니다.

3. 사용 목적:
   - `FakeRepository`는 주로 상태 기반 테스트에 사용됩니다.
   - 동적 mock은 주로 행위 검증(특정 메서드가 호출되었는지 등)에 사용됩니다.

4. 유지보수:
   - `FakeRepository`는 실제 코드와 별도로 관리되어야 합니다.
   - 동적 mock은 테스트 코드 내에서 직접 정의되므로 관리가 더 쉬울 수 있습니다.

따라서, 이 예제가 "모킹을 사용하지 않는 방법"이라고 설명된 것은 완전히 정확하지는 않습니다. 더 정확한 설명은 다음과 같을 것입니다:

"이 예제는 동적 mocking 라이브러리를 사용하지 않고, 수동으로 구현된 Fake 객체를 활용한 테스트 방법을 보여줍니다."

이 접근 방식은 여전히 mocking의 한 형태이지만, 일반적으로 얘기하는 동적 mocking과는 다른 특성을 가지고 있습니다. 각 방식은 상황에 따라 장단점이 있으며, 프로젝트의 요구사항과 팀의 선호도에 따라 선택할 수 있습니다.

## Mock 객체 생성의 내부 동작

Mock 객체 생성의 내부 동작을 이해하는 것은 단위 테스트와 테스트 주도 개발(TDD)에 매우 중요합니다.

Mock 객체 생성의 핵심 원리:

1. 런타임 코드 생성
2. 리플렉션 (Reflection)
3. 프록시 패턴
4. 메소드 인터셉션

### 1. 런타임 코드 생성

Mock 프레임워크는 주어진 인터페이스나 클래스를 기반으로 실행 시간에 새로운 클래스를 동적으로 생성합니다.

예를 들어, Java에서는 ByteBuddy나 CGLib 같은 라이브러리를 사용하여 런타임에 바이트코드를 생성합니다.
Go에서는 `reflect` 패키지를 사용하여 런타임에 새로운 구조체(struct)를 만듭니다.

```go
// Go에서의 런타임 타입 생성 예시
type Interface interface {
    Method() string
}

func CreateMock() Interface {
    return &struct{ Interface }{} // 익명 구조체 생성
}
```

### 2. 리플렉션 (Reflection)

리플렉션을 사용하여 Mock 대상 타입(인터페이스나 클래스)의 메소드와 속성을 분석합니다.
이를 통해 Mock 객체가 어떤 메소드를 구현해야 하는지 결정합니다.

```java
// Java에서의 리플렉션 사용 예시
Method[] methods = interfaceToMock.getDeclaredMethods();
for (Method method : methods) {
    // 각 메소드에 대한 Mock 구현 생성
}
```

### 3. 프록시 패턴

대부분의 Mock 프레임워크는 프록시 패턴을 사용합니다.
실제 객체 대신 프록시 객체를 생성하여 메소드 호출을 가로챕니다.

```java
// Java에서의 동적 프록시 생성 예시
Interface proxy = (Interface) Proxy.newProxyInstance(
    Interface.class.getClassLoader(),
    new Class<?>[] { Interface.class },
    new InvocationHandler() {
        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            // Mock 동작 구현
        }
    }
);
```

### 4. 메소드 인터셉션

생성된 Mock 객체는 원본 메소드 호출을 가로채고, 대신 미리 정의된 동작을 수행합니다.
이를 통해 테스트에 필요한 특정 동작을 시뮬레이션할 수 있습니다.

```java
// Mockito를 사용한 Java에서의 메소드 인터셉션 예시
when(mockObject.someMethod()).thenReturn("Mocked Result");
```

### 실제 구현 예시 (Java)

간단한 Mock 프레임워크의 핵심 부분을 Java로 구현해보겠습니다:

```java
import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.HashMap;
import java.util.Map;

public class SimpleMockFramework {

    public static <T> T mock(Class<T> interfaceToMock) {
        InvocationHandler handler = new MockInvocationHandler();
        return (T) Proxy.newProxyInstance(
            interfaceToMock.getClassLoader(),
            new Class<?>[] { interfaceToMock },
            handler
        );
    }

    private static class MockInvocationHandler implements InvocationHandler {
        private Map<Method, Object> stubbedMethods = new HashMap<>();

        public void when(Method method, Object returnValue) {
            stubbedMethods.put(method, returnValue);
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            if (stubbedMethods.containsKey(method)) {
                return stubbedMethods.get(method);
            }
            // 기본 동작: null 반환
            return null;
        }
    }

    public static <T> OngoingStubbing<T> when(T methodCall) {
        // 실제 구현에서는 마지막으로 호출된 메소드를 추적해야 합니다.
        return new OngoingStubbing<>();
    }

    public static class OngoingStubbing<T> {
        public void thenReturn(T returnValue) {
            // 실제 구현에서는 추적된 메소드에 대해 반환값을 설정해야 합니다.
        }
    }
}
```

이 간단한 구현에서:

1. `mock` 메소드는 주어진 인터페이스의 Mock 객체를 생성합니다.
2. `MockInvocationHandler`는 메소드 호출을 인터셉트하고 미리 정의된 동작을 수행합니다.
3. `when` 메소드와 `OngoingStubbing` 클래스는 Mockito 스타일의 API를 모방하여 Mock 동작을 정의합니다.

실제 사용 예시:

```java
interface MyInterface {
    String doSomething();
}

MyInterface mock = SimpleMockFramework.mock(MyInterface.class);
SimpleMockFramework.when(mock.doSomething()).thenReturn("Mocked Result");

String result = mock.doSomething(); // "Mocked Result" 반환
```

이것은 매우 기본적인 구현이며, 실제 Mock 프레임워크는 훨씬 더 복잡하고 다양한 기능을 제공합니다.
예를 들어, 인자 매칭, 호출 횟수 검증, 예외 던지기 등의 기능을 지원합니다.

Go에서의 mockery나 PHP의 PHPUnit과 같은 프레임워크들도 이와 유사한 원리로 동작하지만,
각 언어의 특성에 맞게 구현 details가 다릅니다.

예를 들어, Go는 인터페이스 기반 다형성을 사용하므로 Mock 생성 시 해당 인터페이스를 구현하는 새로운 구조체를 만듭니다.

## 모킹 없는 테스트

"클래식한 테스트 방식" 또는 "상태 기반 테스트"라고 부르기도 합니다.

예시:

```go
// main.go
package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type User struct {
    ID   int
    Name string
}

type UserService struct {
    db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) CreateUser(name string) (int, error) {
    result, err := s.db.Exec("INSERT INTO users (name) VALUES (?)", name)
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    return int(id), err
}

func (s *UserService) GetUser(id int) (*User, error) {
    var user User
    err := s.db.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// main_test.go
package main

import (
    "database/sql"
    "testing"
    _ "github.com/go-sql-driver/mysql"
)

func TestUserService(t *testing.T) {
    // 실제 테스트 데이터베이스 연결
    db, err := sql.Open("mysql", "user:password@/testdb")
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    defer db.Close()

    // 테스트를 위한 테이블 생성
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(50) NOT NULL
    )`)
    if err != nil {
        t.Fatalf("Failed to create test table: %v", err)
    }

    // 테스트 후 정리를 위한 defer 함수
    defer func() {
        _, err := db.Exec("DROP TABLE users")
        if err != nil {
            t.Errorf("Failed to drop test table: %v", err)
        }
    }()

    service := NewUserService(db)

    // CreateUser 테스트
    id, err := service.CreateUser("John Doe")
    if err != nil {
        t.Errorf("Failed to create user: %v", err)
    }
    if id <= 0 {
        t.Errorf("Expected positive ID, got %d", id)
    }

    // GetUser 테스트
    user, err := service.GetUser(id)
    if err != nil {
        t.Errorf("Failed to get user: %v", err)
    }
    if user.Name != "John Doe" {
        t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
    }
}
```

이 방식의 이점:

1. 실제 동작 테스트: 모킹 없이 실제 데이터베이스와의 상호작용을 테스트하므로, 프로덕션 환경과 가장 유사한 조건에서 테스트할 수 있습니다.

2. 간단성: 모킹 라이브러리나 복잡한 설정 없이 직관적인 테스트 코드를 작성할 수 있습니다.

3. 통합 테스트의 성격: 데이터베이스 연결, SQL 쿼리 실행 등 전체 흐름을 테스트할 수 있어, 통합 테스트의 성격을 가집니다.

4. 신뢰성: 실제 데이터베이스를 사용하므로, 데이터베이스 관련 문제(예: 트랜잭션, 동시성 등)를 조기에 발견할 수 있습니다.

5. 학습 곡선: 추가적인 모킹 기술을 배울 필요가 없어, 초보 개발자도 쉽게 테스트를 작성할 수 있습니다.

6. 리팩토링 안전성: 내부 구현이 변경되어도 테스트가 깨지지 않습니다. 최종 결과만을 검증하기 때문입니다.

주의할 점:

1. 성능: 실제 데이터베이스 연결을 사용하므로 테스트 실행 속도가 느릴 수 있습니다.

2. 격리성: 테스트 간 데이터 충돌이 발생할 수 있으므로, 각 테스트 전후로 데이터를 정리해야 합니다.

3. 환경 의존성: 테스트 데이터베이스 설정이 필요하며, CI/CD 파이프라인에서 실행할 때 추가 설정이 필요할 수 있습니다.

4. 병렬 실행의 어려움: 동일한 데이터베이스를 사용하는 테스트들을 병렬로 실행하기 어려울 수 있습니다.

이 방식은 특히 작은 프로젝트나 데이터베이스 상호작용이 중요한 부분에서 유용할 수 있습니다.
그러나 대규모 프로젝트나 복잡한 비즈니스 로직을 테스트할 때는 모킹을 사용한 단위 테스트와 이러한 통합 테스트를 적절히 조합하는 것이 좋습니다.
