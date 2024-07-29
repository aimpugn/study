# Test Double

- [Test Double](#test-double)
    - [Test Double 패턴](#test-double-패턴)
    - [역사](#역사)
    - [테스트 더블의 종류](#테스트-더블의-종류)
    - [더미(Dummy)](#더미dummy)
    - [스텁(Stub)](#스텁stub)
    - [스파이(Spy)](#스파이spy)
    - [모의 객체(Mock)](#모의-객체mock)
    - [부분 목(Partial Mock)](#부분-목partial-mock)
    - [가짜 객체(Fake)](#가짜-객체fake)
    - [다른 관련 용어들](#다른-관련-용어들)

## Test Double 패턴

테스트 더블 패턴은 소프트웨어 테스트에서 실제 객체를 대신하여 사용되는 객체를 생성하는 기법입니다.

*테스트 더블*은 소프트웨어 테스팅에서 *실제 객체 대신 사용되는 객체*를 총칭하는 용어입니다.
테스트 더블이라는 용어는 영화에서 위험한 장면을 대신 수행하는 스턴트 더블(stunt double)에서 유래했습니다.

이 패턴의 주요 목적:
- 테스트의 복잡성을 줄이기
- 테스트 대상 코드를 격리하여 보다 효과적이고 효율적인 단위 테스트를 가능하게 하기
- 실제 객체의 행동을 모방하거나 대체하여, 테스트 중인 코드가 예상대로 동작하는지 확인하기

실제 객체의 복잡성, 의존성, 또는 실행 시간으로 인한 문제를 해결하고, 테스트의 제어와 관찰을 용이하게 합니다.
특정 컴포넌트나 기능을 격리하여 테스트할 수 있으며, 다양한 시나리오와 엣지 케이스를 쉽게 시뮬레이션할 수 있습니다.

- 장점

    1. 테스트 대상 코드 격리
    2. 테스트 속도 향상
    3. 예측 가능한 테스트 환경 제공
    4. 복잡한 의존성 처리 용이
    5. 엣지 케이스 및 오류 상황 시뮬레이션 가능

- 단점

    1. 실제 객체와의 차이로 인한 테스트 신뢰성 저하 가능성
    2. 과도한 모킹으로 인한 테스트 복잡성 증가
    3. 실제 통합 문제 발견의 어려움
    4. 테스트 더블 유지보수에 따른 추가 작업

테스트 더블을 효과적으로 사용하면 코드의 품질을 향상시키고, 버그를 조기에 발견하며, 전반적인 개발 프로세스를 개선할 수 있습니다.
그러나 과도한 사용이나 부적절한 적용은 오히려 테스트의 가치를 떨어뜨릴 수 있으므로 주의가 필요합니다.

## 역사

테스트 더블의 개념은 오래전부터 존재했지만, 이 용어를 공식화하고 체계화한 것은 Gerard Meszaros입니다.
2007년 그의 저서 "xUnit Test Patterns: Refactoring Test Code"에서 테스트 더블이라는 용어를 처음 소개했습니다.

- 1990년대 말: 단위 테스트와 테스트 주도 개발(TDD)의 부상
- 2000년대 초: 모의 객체(Mock Objects) 개념의 등장
- 2007년: Gerard Meszaros가 "테스트 더블" 용어 공식화
- 2010년대: 다양한 모킹 프레임워크와 도구의 발전

## 테스트 더블의 종류

1. 더미 객체 (Dummy Objects):

    단순히 인자를 채우기 위해 사용되는 객체입니다.
    단순히 인스턴스화된 객체로, 실제로는 사용되지 않지만 메서드 호출 시 필요한 *매개변수 목록을 채우는 용도*로 사용됩니다.

2. 페이크 객체 (Fake Objects):

    실제 구현의 단순화된 버전입니다.
    실제 객체의 복잡한 로직을 단순화하여 구현한 객체로, 주로 외부 의존성을 대체할 때 사용됩니다.

    예를 들어, 실제 데이터베이스 대신 메모리 내 데이터베이스를 사용할 수 있습니다.

3. 스텁 (Stubs):

    테스트 중에 호출될 때 *미리 준비된 답변*을 제공합니다.
    일반적으로 테스트를 위해 프로그래밍된 것 외에는 응답하지 않습니다.
    호출에 대해 미리 정의된 응답을 반환하며, 주로 상태 기반 테스트에 사용됩니다.

4. 부분 스텁 (Partial Stubs):

    부분 목과 유사하지만, 오버라이드된 메서드가 *미리 정의된 응답만을 반환*합니다.
    실제 객체의 일부 동작만 스텁으로 대체하고 나머지는 원래 구현을 사용합니다.

5. 스파이 (Spies):

    스텁과 유사하지만, 호출에 대한 정보도 기록합니다.
    실제 객체의 메서드를 호출하면서 동시에 해당 호출에 대한 정보를 기록합니다.

    예를 들어, 메서드가 몇 번 호출되었는지 등을 추적할 수 있습니다.

6. 목 객체 (Mock Objects):

    호출될 것으로 예상되는 메서드와 반환할 결과를 미리 프로그래밍합니다.
    예상과 다른 방식으로 사용되면 예외를 발생시킬 수 있습니다.

    특정 메서드 호출의 발생 여부, 횟수, 순서 등을 검증하는 행위 기반 테스트에 사용됩니다.

7. 부분 목 (Partial Mocks):

    원래 클래스의 특정 메서드만 오버라이드하고, 나머지 메서드는 원래 클래스의 구현을 사용하는 테스트 객체입니다.
    대부분의 메서드는 원래 구현을 유지하고, 특정 메서드만 목 동작을 수행합니다.

    호출된 메서드를 기록하고 검증할 수 있습니다.

    특정 메서드를 오버라이드하면서, 이 메서드가 호출되었는지 여부를 검증하는 기능을 추가한다면 부분 목의 특성을 가지고 있습니다.

## 더미(Dummy)

```go
type User struct {
    ID   int
    Name string
}

type UserService struct {
    // ... 다른 필드들
}

func (s *UserService) CreateUser(user User) error {
    // 실제 구현
    return nil
}

// 테스트
func TestCreateUser(t *testing.T) {
    dummyUser := User{} // 더미 객체
    service := &UserService{}
    err := service.CreateUser(dummyUser)
    if err != nil {
        t.Errorf("예상치 못한 오류 발생: %v", err)
    }
}
```

## 스텁(Stub)

테스트 스텁은 테스트 더블의 한 유형으로, 특정 메서드에 대해 미리 정의된 응답을 제공하는 객체입니다.
주요 특징은 다음과 같습니다:
- 호출된 메서드가 고정된 결과를 반환하도록 설정합니다.
- 외부 의존성을 제거하고 테스트를 단순화합니다.
- 주로 간단한 대체 구현을 제공하는 데 사용됩니다.

```go
type UserRepository interface {
    GetUser(id int) (*User, error)
}

type StubUserRepository struct{}

func (s *StubUserRepository) GetUser(id int) (*User, error) {
    // 항상 동일한 사용자 반환
    return &User{ID: id, Name: "Test User"}, nil
}

func TestGetUserName(t *testing.T) {
    stub := &StubUserRepository{}
    user, err := stub.GetUser(1)
    if err != nil {
        t.Fatalf("오류 발생: %v", err)
    }
    if user.Name != "Test User" {
        t.Errorf("예상 이름: Test User, 실제 이름: %s", user.Name)
    }
}
```

## 스파이(Spy)

```go
type SpyUserRepository struct {
    GetUserCalls int
    Users        map[int]*User
}

func (s *SpyUserRepository) GetUser(id int) (*User, error) {
    s.GetUserCalls++
    user, ok := s.Users[id]
    if !ok {
        return nil, fmt.Errorf("사용자를 찾을 수 없음")
    }
    return user, nil
}

func TestGetUserCalls(t *testing.T) {
    spy := &SpyUserRepository{
        Users: map[int]*User{
            1: {ID: 1, Name: "Alice"},
        },
    }
    
    _, _ = spy.GetUser(1)
    _, _ = spy.GetUser(1)
    
    if spy.GetUserCalls != 2 {
        t.Errorf("GetUser 호출 횟수 예상: 2, 실제: %d", spy.GetUserCalls)
    }
}
```

테스트 스텁의 한계

1. **특정성의 한계**:

    스텁은 주로 간단한 대체 구현을 제공합니다.
    복잡한 상황, 예를 들어 일부 메서드만 오버라이드하고 나머지는 원래 구현을 유지해야 하는 경우에는 적합하지 않을 수 있습니다.

2. **기능의 제한**:

    스텁은 일반적으로 미리 정의된 고정 응답만을 제공합니다.
    동적인 상황이나 복잡한 로직이 필요한 경우, 스텁만으로는 충분하지 않을 수 있습니다.

3. **상태 변경 추적의 어려움**:

    스텁은 주로 응답 제공에 초점을 맞추기 때문에, 객체의 상태 변화를 추적하거나 검증하는 데는 적합하지 않습니다.

4. **상호작용 검증의 한계**:

    스텁은 메서드 호출 여부나 순서 등을 검증하는 기능이 없어, 복잡한 상호작용을 테스트하는 데 제한이 있습니다.

## 모의 객체(Mock)

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetUser(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

func TestGetUserWithMock(t *testing.T) {
    mockRepo := new(MockUserRepository)
    
    // 기대값 설정
    mockRepo.On("GetUser", 1).Return(&User{ID: 1, Name: "Mock User"}, nil)
    
    user, err := mockRepo.GetUser(1)
    
    if err != nil {
        t.Fatalf("예상치 못한 오류: %v", err)
    }
    
    if user.Name != "Mock User" {
        t.Errorf("예상 이름: Mock User, 실제 이름: %s", user.Name)
    }
    
    // 모든 기대값이 충족되었는지 확인
    mockRepo.AssertExpectations(t)
}
```

## 부분 목(Partial Mock)

```java
public class CriticalController {
    protected void doSomethingCustom() {
        // 원래 구현은 비어 있음
    }

    public void criticalMethod() {
        // 일반적인 구현
    }
}

// 실제 객체의 일부 메서드만 오버라이드하여 테스트용으로 사용하는 객체입니다.
// 대부분의 메서드는 원래 클래스의 구현을 그대로 사용하고, 특정 메서드만 테스트를 위해 다르게 구현합니다.
public class TestController extends CriticalController {
    @Override
    protected void doSomethingCustom() {
        // 테스트를 위한 구현
        System.out.println("Test output from doSomethingCustom");
    }
    // criticalMethod는 오버라이드하지 않음
}
```

## 가짜 객체(Fake)

```go
type FakeUserRepository struct {
    users map[int]*User
}

func NewFakeUserRepository() *FakeUserRepository {
    return &FakeUserRepository{
        users: make(map[int]*User),
    }
}

func (f *FakeUserRepository) GetUser(id int) (*User, error) {
    user, ok := f.users[id]
    if !ok {
        return nil, fmt.Errorf("사용자를 찾을 수 없음")
    }
    return user, nil
}

func (f *FakeUserRepository) AddUser(user *User) {
    f.users[user.ID] = user
}

func TestGetUserWithFake(t *testing.T) {
    fakeRepo := NewFakeUserRepository()
    fakeRepo.AddUser(&User{ID: 1, Name: "Fake User"})
    
    user, err := fakeRepo.GetUser(1)
    
    if err != nil {
        t.Fatalf("예상치 못한 오류: %v", err)
    }
    
    if user.Name != "Fake User" {
        t.Errorf("예상 이름: Fake User, 실제 이름: %s", user.Name)
    }
}
```

## 다른 관련 용어들

1. **프록시 컨트롤러 (Proxy Controller)**:
   - 주로 디자인 패턴에서 사용되는 용어로, 실제 객체에 대한 접근을 제어하는 객체를 의미합니다.
   - 테스트 컨텍스트에서는 직접적으로 사용되지 않지만, 유사한 개념이 적용될 수 있습니다.

2. **인터셉터 (Interceptor)**:
   - 메서드 호출을 가로채고 추가 로직을 실행하는 객체를 의미합니다.
   - 테스트 더블의 맥락에서는 직접적으로 사용되지 않지만, 스파이나 목 객체의 일부 기능과 유사한 점이 있습니다.
