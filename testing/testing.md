# testing

## 테스트의 목적

테스트의 목적은 크게 두 가지로 나눌 수 있습니다:

- 기능적 정확성 검증: 코드가 예상대로 동작하는지 확인
- 변경 감지: 코드 변경이 기존 동작에 영향을 미치는지 확인

## 구현 세부사항 vs 계약(Contract)

- 구현 세부사항: 내부적으로 어떻게 동작하는지에 대한 것
- 계약: 외부에서 볼 때 어떻게 동작해야 하는지에 대한 것

## 트랜잭션 관리 방식 테스트의 장단점

중요한 내부 동작의 변경을 감지하는 것은 분명 가치가 있습니다.
다만, 이를 어떻게 테스트에 반영할지는 상황에 따라 신중히 결정해야 합니다.
공개 API 테스트와 중요 내부 동작 테스트를 적절히 분리하고 조합하는 것이 좋은 접근 방식이 될 수 있습니다.

장점:
- 변경 감지: 트랜잭션 관리 방식의 변경을 즉시 감지할 수 있습니다.
- 정확성 보장: 트랜잭션이 예상대로 시작, 커밋 또는 롤백되는지 확인할 수 있습니다.

단점:
- 유연성 저하: 구현 방식을 조금만 바꿔도 테스트가 실패할 수 있습니다.
- 리팩토링 어려움: 내부 로직을 변경할 때마다 테스트도 수정해야 할 수 있습니다.

이상적인 접근 방식은 다음과 같은 균형을 잡는 것입니다:

1. 공개 API(public interface)의 동작을 테스트하여 계약을 검증합니다.
2. 중요한 내부 동작(예: 트랜잭션 관리)에 대해서는 별도의 테스트를 작성합니다.

예를 들어:

```go
type UserService struct {
    db Database
}

func (s *UserService) CreateUser(user User) error {
    return s.db.WithTransaction(func(tx Transaction) error {
        return tx.InsertUser(user)
    })
}

// Database 인터페이스
type Database interface {
    WithTransaction(fn func(Transaction) error) error
}

type Transaction interface {
    InsertUser(user User) error
}

// 테스트 코드
func TestCreateUser(t *testing.T) {
    // 1. 공개 API 테스트
    mockDB := &MockDatabase{}
    service := &UserService{db: mockDB}

    mockDB.On("WithTransaction", mock.AnythingOfType("func(Transaction) error")).Return(nil)

    err := service.CreateUser(User{Name: "John"})

    assert.NoError(t, err)
    mockDB.AssertCalled(t, "WithTransaction", mock.AnythingOfType("func(Transaction) error"))

    // 2. 트랜잭션 동작 테스트
    mockTx := &MockTransaction{}
    mockDB.On("WithTransaction", mock.AnythingOfType("func(Transaction) error")).Run(func(args mock.Arguments) {
        fn := args.Get(0).(func(Transaction) error)
        fn(mockTx)
    }).Return(nil)

    mockTx.On("InsertUser", mock.AnythingOfType("User")).Return(nil)

    err = service.CreateUser(User{Name: "John"})

    assert.NoError(t, err)
    mockTx.AssertCalled(t, "InsertUser", mock.AnythingOfType("User"))
}
```

이 접근 방식의 장점:

1. 공개 API 테스트: `CreateUser` 메서드가 예상대로 동작하는지 확인합니다.
2. 트랜잭션 사용 확인: 트랜잭션 내에서 작업이 수행되는지 확인합니다.
3. 유연성: 트랜잭션 관리 방식이 변경되어도 (예: `WithTransaction` 대신 직접 `Begin`, `Commit` 사용) 첫 번째 테스트는 여전히 통과할 것입니다.
4. 세부 동작 검증: 두 번째 테스트에서 트랜잭션 내부 동작을 검증합니다.

이렇게 함으로써:
- 공개 API의 동작을 보장하면서도
- 중요한 내부 구현(트랜잭션 사용)을 검증하고
- 동시에 일정 수준의 유연성을 유지할 수 있습니다.
