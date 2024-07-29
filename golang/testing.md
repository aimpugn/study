# testing

- [testing](#testing)
    - [테스트 커버리지](#테스트-커버리지)
    - [`assert.IsTypef` 함수와 `assert.IsType` 함수](#assertistypef-함수와-assertistype-함수)
    - [`sqlmock`과 repository 테스트](#sqlmock과-repository-테스트)
        - [`sqlmock` examples](#sqlmock-examples)
            - [Example1](#example1)
            - [Example 2](#example-2)
    - [assertions](#assertions)
        - [`github.com/stretchr/testify/assert`](#githubcomstretchrtestifyassert)
        - [`"gotest.tools/assert"`](#gotesttoolsassert)
        - [왜 별도의 assert를 사용?](#왜-별도의-assert를-사용)
    - [병렬 테스트](#병렬-테스트)
        - [병렬 테스트 사용 방법](#병렬-테스트-사용-방법)
        - [설정](#설정)
        - [병렬 실행의 작동 방식](#병렬-실행의-작동-방식)
        - [`t.Run(name, func(t *testing.T))`와 `t.Parallel`](#trunname-funct-testingt와-tparallel)
        - [주의 사항](#주의-사항)
        - [On Using Go's `t.Parallel()`](#on-using-gos-tparallel)

## 테스트 커버리지

Go 테스트 커버리지를 확인하고, 어떤 코드가 커버되지 않았는지 구체적으로 확인하려면, 테스트 커버리지 데이터를 파일로 생성하고 해당 파일을 분석하는 과정을 거칩니다. 다음은 이 과정을 수행하는 방법을 단계별로 설명합니다:

1. **커버리지 데이터 파일 생성**: 먼저, `go test` 명령어를 사용하여 테스트를 실행하고, `-coverprofile` 옵션을 추가하여 커버리지 데이터를 파일로 저장합니다. 예를 들어, 커버리지 데이터를 `coverage.out` 파일에 저장하려면 다음과 같이 실행합니다:

    ```sh
    # 현재 디렉토리와 모든 하위 디렉토리에 대한 테스트를 실행하고, 테스트 커버리지 결과를 `coverage.out` 파일에 저장
    go test -coverprofile=coverage.out ./...
    ```

2. **커버리지 데이터 분석**: `coverage.out` 파일에 저장된 커버리지 데이터를 분석하여 어떤 코드가 테스트에 의해 커버되지 않았는지 확인할 수 있습니다. Go는 이를 위한 `go tool cover` 명령어를 제공합니다. `-html` 옵션을 사용하여 커버리지 데이터를 HTML 형식으로 변환하고 웹 브라우저에서 확인할 수 있습니다:

    ```sh
    # coverage.out 파일에 저장된 커버리지 데이터를 읽어서 HTML 형식으로 변환 후 기본 웹 브라우저에서 결과를 열어 보여준다.
    go tool cover -html=coverage.out
    ```

## `assert.IsTypef` 함수와 `assert.IsType` 함수

- **`assert.IsType`**: This function checks if the object is of the expected type. It's useful when you want to ensure that a variable or return value from a function is of a specific type. The function signature is usually something like `IsType(t *testing.T, expectedType interface{}, object interface{}, msgAndArgs ...interface{})`, where `expectedType` is the type you're expecting, and `object` is the actual variable or value you're testing.

- **`assert.IsTypef`**: This variant typically does not exist in standard assertion libraries. In Go's testing patterns, functions with an `f` suffix, like `Errorf`, `Fatalf`, etc., indicate that the function accepts a format string followed by values (similar to `printf` in C). These are used to format error messages or test failure messages dynamically. If you're seeing a reference to `assert.IsTypef`, it might be a misunderstanding, a custom extension of an assertion library, or a typo.

Given the context, the difference you're asking about seems to stem from a confusion with functions that do have an `f` variant for formatted string support, which is common in logging or error reporting functions in Go. However, in the context of `testify` or similar testing frameworks, `assert.IsType` is the standard function used to assert the type of a variable without the formatted message support. If you need to include a custom message with an assertion, `assert.IsType` usually allows passing additional arguments at the end for this purpose, but it doesn't use the `f` suffix pattern as formatting functions do.

Without direct documentation or code from a specific testing library that includes `assert.IsTypef`, the most accurate advice is to refer to the documentation of the testing framework you're using. For standard uses, relying on `assert.IsType` and understanding its parameters for including custom failure messages would be the way to go.

## `sqlmock`과 repository 테스트

네, `sqlmock` 라이브러리를 사용하는 것은 Go 언어에서 데이터베이스 상호작용을 테스트할 때 매우 일반적인 방법이다.
`sqlmock`을 사용하면 실제 데이터베이스에 연결하지 않고 SQL 데이터베이스와의 상호작용을 모의(Mock)할 수 있으므로, 테스트 실행 속도를 높이고 테스트 환경을 보다 일관되게 유지할 수 있다.
이 방법을 사용하면 리포지토리의 로직을 효과적으로 검증할 수 있으며, 테스트 중 발생하는 외부 의존성 문제를 최소화할 수 있다.

`sqlmock`을 사용하여 리포지토리 메서드에 대한 테스트를 작성하는 기본적인 절차는 다음과 같다:

1. **Mock DB 생성:** `db, mock, err := sqlmock.New()`를 호출하여 mock 데이터베이스 커넥션과 관련된 mock 객체를 생성.
2. **예상되는 동작 설정:** `mock.ExpectQuery`, `mock.ExpectExec` 등의 메서드를 사용하여 데이터베이스에 대한 특정 쿼리 또는 명령이 실행될 때 예상되는 동작을 설정
3. **테스트 대상 함수 실행:** 테스트하려는 리포지토리 메서드를 실행한다. 이 때, 실제 데이터베이스 커넥션 대신 mock 데이터베이스 커넥션을 사용한다.
4. **결과 및 동작 검증:** 실행 결과를 검증하고 `sqlmock`에 설정된 예상 동작이 올바르게 수행되었는지 확인.

### `sqlmock` examples

#### Example1

```go
package users

import (
    "context"
    "testing"
    "time"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/require"
)

func TestGetUserByAccessToken(t *testing.T) {
    // `sqlmock.New()`를 호출하여 mock DB 커넥션을 생성
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    sqlxDB := sqlx.NewDb(db, "sqlmock")

    repo := &mysqlRepositoryV1{db: sqlxDB}

    // `ExpectQuery` 메서드를 사용하여 특정 SQL 쿼리가 실행될 때 반환될 예상 결과를 설정
    mock.ExpectQuery("^SELECT (.+) FROM users WHERE access_token = ? AND expired < ?$").
        // 두 번째 인자는 expiredInDateTimeFormat 함수의 반환값이므로, 정확한 값을 알 수 없는 경우 sqlmock.AnyArg() 사용
        WithArgs("validToken", sqlmock.AnyArg()).
        WillReturnRows(
            sqlmock.NewRows([]string{"id", "code"}).
                AddRow(1, "UserCode")
        )

    // 실제 메서드를 호출하고 반환값을 검증
    user, err := repo.GetUserByAccessToken(context.Background(), "validToken")

    // 결과 및 동작 검증
    require.NoError(t, err)
    require.NotNil(t, user)
    require.Equal(t, int64(1), user.Id)
    require.Equal(t, "UserCode", *user.Code)

    // sqlmock에 설정된 모든 예상이 충족되었는지 확인
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }
}
```

#### Example 2

```go

// 예를 들어, 이 함수는 QueryRowContext를 사용하여 데이터베이스에서 사용자의 이름을 조회합니다.
func GetUserNameByID(ctx context.Context, db *sqlx.DB, userID int) (string, error) {
    var name string
    err := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", userID).Scan(&name)
    if err != nil {
        return "", err
    }
    return name, nil
}

func TestGetUserNameByID(t *testing.T) {
    // sqlmock 데이터베이스 연결과 mock 객체 생성
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    // 예상되는 쿼리 및 반환값 설정
    rows := sqlmock.NewRows([]string{"name"}).AddRow("John Doe")
    mock.ExpectQuery("SELECT name FROM users WHERE id = ?").
        WithArgs(sqlmock.AnyArg()).
        WillReturnRows(rows)

    sqlxDB := sqlx.NewDb(db, "mysql")

    // 함수 호출
    name, err := GetUserNameByID(context.Background(), sqlxDB, 1)
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", name)

    // 모든 예상이 충족되었는지 확인
    err = mock.ExpectationsWereMet()
    assert.NoError(t, err)
}
```

## assertions

`"gotest.tools/assert"`와 `"github.com/stretchr/testify/assert"`처럼 Go 언어에서 테스트를 작성할 때 사용할 수 있는 여러 assertion 라이브러리들이 있다.
이들은 테스트 케이스에서 조건을 검증하기 위해 사용되며, 테스트 코드의 가독성과 유지 보수성을 향상시키는 데 도움을 준다.
비록 비슷한 기능을 제공하지만, 각각의 라이브러리는 독특한 특징과 사용 방법을 가지고 있다.

### `github.com/stretchr/testify/assert`

이 라이브러리는 테스트 실패 시 더 나은 추적 정보를 제공한다.
이는 함수를 조합할 때 특히 유용하며, 에러 경로를 더 잘 이해할 수 있게 해준다.
그러나, 특히 Protobuf 값을 비교할 때 발생하는 문제점이 있다.
이 문제로 인해 `gotest.tools/assert`를 사용하는 것이 권장된다.

- **널리 사용됨**: `testify`는 Go 커뮤니티에서 널리 사용되며, 많은 오픈 소스 프로젝트에서 찾아볼 수 있습니다. 이는 라이브러리가 안정적이고 신뢰할 수 있음을 의미합니다.
- **풍부한 기능**: `testify`는 다양한 assertion 함수를 제공합니다. 이는 개발자가 다양한 유형의 검증을 쉽게 수행할 수 있게 해줍니다.
- **Mocking 지원**: `testify`는 `mock` 패키지를 통해 mocking과 테스트 더블을 쉽게 생성하고 사용할 수 있는 기능을 제공합니다. 이는 단위 테스트를 작성할 때 특히 유용합니다.
- **사용자 친화적인 오류 메시지**: 실패한 assertion에 대해 자세하고 이해하기 쉬운 오류 메시지를 제공합니다. 이는 문제를 빠르게 진단하고 수정하는 데 도움이 됩니다.

### `"gotest.tools/assert"`

이 라이브러리는 Go의 표준 테스팅 패키지(`testing`)를 기반으로 하며, 테스트 실패 시 더 나은 "diff" 출력을 제공한다.
이는 값을 비교할 때 유용하며, 특히 복잡한 데이터 구조를 비교할 때 더 명확한 차이점을 보여준다.
또한, 테스트 실패 시 발생하는 에러 경로에 대한 정보를 제공하는 데 유용하다. 이는 함수를 조합할 때 특히 유용하다.

- **간결함과 직관성**: `gotest.tools/assert`는 간결하고 직관적인 API를 제공합니다. 이는 테스트 코드를 더욱 읽기 쉽고 이해하기 쉽게 만들어 줍니다.
- **커스텀 검증기**: 사용자가 커스텀 검증 함수를 쉽게 작성하고 사용할 수 있게 해주는 기능을 제공합니다. 이는 표준 라이브러리에서 제공하지 않는 특정 조건을 검증해야 할 때 유용합니다.
- **포커스된 기능**: `gotest.tools/assert`는 assertion 기능에 초점을 맞추고 있으며, mocking과 같은 추가적인 기능을 제공하지 않습니다. 이는 라이브러리가 더 가볍고 특정 목적에 맞춰져 있음을 의미합니다.

### 왜 별도의 assert를 사용?

- **프로젝트 요구사항**:
    - 프로젝트의 특정 요구사항이나 개발자의 선호도에 따라 더 적합한 라이브러리를 선택할 수 있습니다.
- **개인적 선호도**:
    - 개발자마다 코드 스타일과 선호하는 도구가 다를 수 있다.
    - 어떤 개발자는 `testify`의 풍부한 기능을 선호할 수 있다
    - 다른 개발자는 `gotest.tools/assert`의 간결함을 선호할 수 있다
- **기존 코드베이스**:
    - 기존 프로젝트에서 이미 사용 중인 라이브러리가 있다면, 일관성을 유지하기 위해 동일한 라이브러리를 사용하는 것이 좋다.

## 병렬 테스트

### 병렬 테스트 사용 방법

`t.Parallel()`은 테스트 함수 내에서 호출되며, 테스트 케이스가 병렬로 실행될 수 있도록 Go 테스트 러너에게 알립니다. 이 메서드를 호출하면, Go 테스트 러너는 여러 테스트를 동시에 실행할 수 있습니다.
이는 테스트 함수가 다른 병렬 테스트 함수들과 함께 동시에 실행되어야 함을 의미합니다.
그러나 실제로 병렬 실행이 시작되는 시점은 테스트 프레임워크의 스케줄링에 따라 결정됩니다.

```go
func TestA(t *testing.T) {
    t.Parallel()
    // 테스트 로직
}

func TestB(t *testing.T) {
    t.Parallel()
    // 테스트 로직
}
```

위 예제에서 `TestA`와 `TestB`는 병렬로 실행될 수 있습니다. 각 테스트는 독립적인 고루틴에서 실행되므로, 테스트 간에 상태를 공유하지 않도록 주의해야 합니다.

### 설정

Go 테스트 도구는 `-parallel` 플래그를 사용하여 동시에 실행할 테스트의 최대 수를 설정할 수 있습니다. 예를 들어, 동시에 최대 4개의 테스트를 실행하려면 다음과 같이 실행할 수 있습니다:

```sh
go test -parallel 4
```

병렬 테스트는 테스트의 실행 시간을 단축시키고, 자원의 효율적 사용을 도모할 수 있지만, 테스트의 설계와 환경 설정에 주의가 필요합니다. 테스트가 서로에게 영향을 주지 않도록 격리를 잘 해야 하며, 시스템의 자원을 적절히 관리하는 것이 중요합니다.

### 병렬 실행의 작동 방식

1. **테스트 함수 호출**: `t.Parallel()`이 호출된 테스트 함수는 시작되지만, 실제 작업은 일시 중지됩니다. 이는 함수가 병렬로 처리될 준비가 되었음을 테스팅 프레임워크에 알리는 것입니다.

2. **일시 중지 및 대기**: `t.Parallel()`을 호출한 테스트 함수는 다른 테스트 함수들이 실행 완료될 때까지 일시 중지됩니다. 이는 병렬로 실행되어야 할 테스트들이 서로 간섭하지 않도록 조정하는 과정입니다.

3. **병렬 실행**: 일반 테스트 함수들이 모두 완료되면, `t.Parallel()`을 호출한 테스트 함수들이 병렬로 실행을 재개합니다. 이때, 동시에 실행될 수 있는 테스트의 수는 `GOMAXPROCS` 값 또는 `-parallel` 플래그로 설정된 값에 의해 제한됩니다.

`GOMAXPROCS`의 기본값과 `-parallel` 플래그의 기본값은 다음과 같습니다:

1. **GOMAXPROCS의 기본값**

   Go 1.5 버전부터 `GOMAXPROCS`의 기본값은 시스템에서 사용 가능한 CPU 코어의 수로 설정됩니다. 이는 Go 프로그램이 멀티코어 환경에서 자동으로 병렬 처리를 최대화할 수 있도록 설정된 것입니다. 이전 버전에서는 기본값이 1이었습니다. 따라서, 현재 Go 버전에서는 `GOMAXPROCS`를 별도로 설정하지 않으면, 사용 가능한 모든 CPU 코어를 사용하도록 설정됩니다

2. **-parallel 플래그의 기본값**

   `-parallel` 플래그는 `go test` 명령어를 사용할 때, 테스트를 병렬로 실행할 수 있는 최대 고루틴(goroutine)의 수를 설정합니다. 이 플래그의 기본값은 `GOMAXPROCS`의 값과 동일하게 설정됩니다. 즉, 시스템의 CPU 코어 수에 따라 결정되며, 이는 테스트 실행 시 병렬 처리를 통해 테스트의 실행 시간을 단축시킬 수 있습니다

### `t.Run(name, func(t *testing.T))`와 `t.Parallel`

```go
func TestAAAAA(t *testing.T) {
    t.Parallel()

    .... 이하 생략 ...

    for testName, testCase := range testCases {
        t.Run(testName, func(t *testing.T) {
            t.Parallel()
```

1. **테스트 함수 최상단에서의 호출**:

   `t.Parallel()`을 테스트 함수의 시작 부분에 호출하면, 해당 테스트 함수는 병렬로 실행될 준비가 됩니다. 이는 함수 전체가 다른 테스트와 동시에 실행될 수 있음을 의미합니다.

2. **t.Run() 내부에서의 호출**:

   s`t.Run()`을 사용하여 서브테스트를 정의할 때 각 서브테스트 내에서 `t.Parallel()`을 호출하면, 각 서브테스트가 별도의 고루틴에서 병렬로 실행될 수 있습니다.
   이는 서브테스트 간에도 동시성을 확보할 수 있음을 의미하며, 특히 다수의 서브테스트가 있는 경우 전체 테스트 시간을 크게 단축시킬 수 있습니다.

`t.Run` 사용하는 경우, 각각의 서브테스트(subtest)도 독립적으로 병렬로 실행되도록 합니다.
이를 통해 각각의 서브테스트(subtest)가 독립적으로 병렬로 실행되도록 합니다.
이 구조를 사용하면, 각 서브테스트는 서로에게 영향을 주지 않고 동시에 실행될 수 있습니다.

```go
func TestParent(t *testing.T) {
    t.Parallel() // 부모 테스트를 병렬로 실행
    testCases := map[string]struct{
        // 테스트 케이스 정의
    }{
        "case1": {},
        "case2": {},
    }

    for testName, testCase := range testCases {
        t.Run(testName, func(t *testing.T) {
            t.Parallel() // 각 서브테스트도 병렬로 실행
            // 서브테스트 로직
        })
    }
}
```

위에서 제공된 코드에서, `TestParent` 함수 내에서 `t.Parallel()`을 호출하면, `TestParent` 테스트는 다른 테스트와 병렬로 실행됩니다.
그리고 각 `t.Run()` 내에서 `t.Parallel()`을 호출함으로써, `case1`, `case2` 등의 각 서브테스트도 독립적으로 병렬로 실행됩니다.
이렇게 설정함으로써, 각 서브테스트는 서로의 실행에 영향을 주지 않고 동시에 실행될 수 있습니다.

### 주의 사항

- **자원 공유에 대한 주의**

    병렬 테스트를 사용할 때는 테스트 간에 공유하는 외부 자원(예: 파일 시스템, 데이터베이스)에 대한 접근시 데이터 경합이 발생할 수 있으므로, 동기화하거나 격리해야 합니다.
    자원 충돌을 방지하기 위해 각 테스트가 독립적인 자원을 사용하도록 설계하는 것이 좋습니다.

- **종속성 관리**

    테스트 간에 종속성이 있는 경우, 병렬 실행이 올바른 결과를 보장하지 않을 수 있습니다.
    따라서 테스트의 독립성을 확보하는 것이 중요합니다.

- **테스트 격리**

    병렬 테스트는 테스트의 독립성을 가정합니다.
    하나의 테스트가 다른 테스트의 실행 환경(예: 전역 변수, 싱글턴 객체)을 변경하지 않도록 주의해야 합니다.

    또한 각 서브테스트 내에서 호출함으로써, 각 서브 테스트 케이스는 별도의 고루틴에서 실행됩니다.
    이는 테스트 간의 상호 작용을 최소화하고, 각 테스트가 독립적인 환경에서 실행되도록 보장합니다.
    특히 공유 자원에 대한 접근이 필요한 테스트에서 중요합니다.

- **성능과 부하**

    병렬 테스트는 시스템에 높은 부하를 줄 수 있습니다.
    시스템의 자원(CPU, 메모리)을 고려하여 적절한 수준에서 병렬 실행을 조절해야 합니다.

    부모 테스트에서 `t.Parallel()`을 호출하는 것만으로는 모든 서브테스트가 자동으로 병렬로 실행되지 않습니다.
    부모 테스트의 `t.Parallel()` 호출은 해당 테스트 함수가 다른 테스트와 병렬로 실행될 수 있음을 의미하지만, 서브테스트의 병렬 실행은 각각 `t.Parallel()`을 호출함으로써 명시적으로 설정해야 합니다.

### [On Using Go's `t.Parallel()`](https://brandur.org/t-parallel)

- [A `TestTx` helper in Go using `t.Cleanup`](https://brandur.org/fragments/go-test-tx-using-t-cleanup)
