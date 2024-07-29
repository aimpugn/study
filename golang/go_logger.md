# go logger

## logging with context

`zap` 라이브러리는 고성능 로깅을 위해 설계된 Go 언어용 로깅 라이브러리입니다. `zap` 자체는 `context`를 인자로 받지 않지만, 웹 요청에서 컨텍스트에 포함된 트레이스 아이디(B3 헤더, Amazon Trace ID 등)를 로깅에 포함시키는 방법은 여러 가지가 있습니다. 이러한 정보를 로깅에 포함시키는 베스트 프랙티스는 로깅을 호출할 때 컨텍스트 정보를 추출하여 로거에 추가하는 것입니다.

### 베스트 프랙티스: 컨텍스트 정보를 로거에 포함시키기

1. **컨텍스트에서 트레이스 정보를 추출**:
   - 웹 요청의 컨텍스트에서 트레이스 아이디를 추출합니다.

2. **로거에 필드로 추가**:
   - 추출한 트레이스 아이디를 `zap` 로거의 필드로 추가하여 로그 메시지에 포함시킵니다.

### 예시 코드

다음은 `zap` 로거를 사용하여 컨텍스트 정보를 로깅에 포함시키는 예시입니다.

#### 1. 컨텍스트 키 정의

먼저, 컨텍스트에 저장될 키를 정의합니다.

```go
package main

import (
    "context"
)

type ContextKey string

const (
    B3TraceIDKey ContextKey = "b3-trace-id"
    AmznTraceIDKey ContextKey = "amzn-trace-id"
)
```

#### 2. 로깅 유틸리티 함수

컨텍스트 정보를 추출하여 `zap` 로거에 필드로 추가하는 유틸리티 함수를 작성합니다.

```go
package main

import (
    "context"
    "go.uber.org/zap"
)

// LogWithContext는 컨텍스트 정보를 포함하여 로그를 출력합니다.
func LogWithContext(ctx context.Context, logger *zap.Logger, message string) {
    fields := []zap.Field{}

    // B3 Trace ID 추가
    if traceID, ok := ctx.Value(B3TraceIDKey).(string); ok {
        fields = append(fields, zap.String(string(B3TraceIDKey), traceID))
    }

    // Amazon Trace ID 추가
    if traceID, ok := ctx.Value(AmznTraceIDKey).(string); ok {
        fields = append(fields, zap.String(string(AmznTraceIDKey), traceID))
    }

    logger.With(fields...).Info(message)
}
```

#### 3. 사용 예시

컨텍스트에 트레이스 아이디를 설정하고, 로깅 유틸리티 함수를 사용하여 로그를 출력합니다.

```go
package main

import (
    "context"
    "go.uber.org/zap"
)

func main() {
    // zap 로거 초기화
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 컨텍스트 생성 및 값 설정
    ctx := context.Background()
    ctx = context.WithValue(ctx, B3TraceIDKey, "b3-trace-id-example")
    ctx = context.WithValue(ctx, AmznTraceIDKey, "amzn-trace-id-example")

    // 로깅 유틸리티 사용
    LogWithContext(ctx, logger, "This is a log message with context")
}
```

### 설명

1. **컨텍스트 키 정의**:
   - `ContextKey` 타입을 정의하여 컨텍스트에 저장될 키를 관리합니다. 이는 문자열 키를 사용하는 것보다 타입 안전성을 높입니다.

2. **로깅 유틸리티 함수**:
   - `LogWithContext` 함수는 `context.Context` 인자와 `zap.Logger` 인자를 받아서 로그 메시지에 컨텍스트 정보를 포함시킵니다.
   - `ctx.Value` 메서드를 사용하여 컨텍스트에서 값을 추출하고, 이를 `zap` 로거의 필드로 추가합니다.

3. **사용 예시**:
   - `context.Background()`를 사용하여 기본 컨텍스트를 생성하고, `context.WithValue`를 사용하여 컨텍스트에 값을 설정합니다.
   - `LogWithContext` 함수를 호출하여 컨텍스트 정보를 포함한 로그 메시지를 출력합니다.

### 결론

`zap` 라이브러리는 `context`를 직접 인자로 받지 않지만, 컨텍스트 정보를 추출하여 로거에 필드로 추가하는 방식으로 트레이스 아이디 등을 로깅에 포함시킬 수 있습니다. 이를 통해 로그 메시지에 더 많은 컨텍스트 정보를 포함시켜 디버깅과 문제 해결을 용이하게 할 수 있습니다. 이 방법은 로깅 유틸리티 함수를 사용하여 일관된 방식으로 컨텍스트 정보를 로깅에 포함시키는 베스트 프랙티스입니다.

### 컨텍스트를 로깅 유틸 함수에 넘기는 것의 타당성 검토

컨텍스트를 로깅 유틸 함수에 넘기는 것은 여러 가지 장점이 있지만, 몇 가지 고려사항도 있습니다. 이를 엄격하게 검토해보겠습니다.

#### 장점

1. **추적 가능성**:
   - 컨텍스트를 통해 요청 ID, 사용자 ID, 트랜잭션 ID 등을 로그에 포함시킬 수 있어, 로그 메시지를 특정 요청이나 사용자와 연관지어 추적할 수 있습니다.

2. **일관성**:
   - 모든 로그 메시지에 일관된 컨텍스트 정보를 포함시킬 수 있어, 로그 분석 및 모니터링 도구에서 로그를 더 쉽게 필터링하고 검색할 수 있습니다.

3. **유연성**:
   - 다양한 상황에 맞게 로그 메시지를 동적으로 생성할 수 있습니다. 예를 들어, 특정 요청에 대한 추가 정보를 로그에 포함시킬 수 있습니다.

#### 단점 및 고려사항

1. **복잡성 증가**:
   - 모든 함수에 컨텍스트를 전달해야 하므로 코드가 복잡해질 수 있습니다. 특히, 컨텍스트가 필요 없는 함수에도 컨텍스트를 전달해야 하는 경우가 발생할 수 있습니다.

2. **성능 오버헤드**:
   - 컨텍스트를 계속 전달하고 값을 추출하는 과정에서 약간의 성능 오버헤드가 발생할 수 있습니다. 하지만 이는 대부분의 경우 무시할 수 있는 수준입니다.

3. **의존성 증가**:
   - 로깅 유틸리티가 컨텍스트에 의존하게 되므로, 컨텍스트가 없는 환경에서는 로깅 유틸리티를 사용하기 어려울 수 있습니다.

### Go 표준 및 커뮤니티 논의

Go 커뮤니티에서는 컨텍스트를 함수에 전달하는 것이 일반적인 관행으로 받아들여지고 있습니다. 특히, 웹 요청과 같은 상황에서 컨텍스트를 사용하여 요청의 수명 주기를 관리하고, 요청과 관련된 메타데이터를 전달하는 것이 권장됩니다.

#### 커뮤니티 논의 예시

- **Stack Overflow**:
    - Go에서 컨텍스트를 함수에 전달하는 것이 일반적인 관행이며, 주로 함수의 첫 번째 인자로 전달됩니다. 이는 함수 간의 의존성을 명확히 하고, 요청의 수명 주기를 관리하는 데 유용합니다 [2].

- **Gopher Academy 블로그**:
    - 컨텍스트를 사용하여 로깅을 포함한 여러 가지 교차 관심사를 처리하는 것이 유용하다고 설명합니다. 이는 로깅, 프로파일링, 예외 처리 등을 일관되게 관리할 수 있게 합니다 [4].

### 베스트 프랙티스

컨텍스트를 로깅 유틸 함수에 넘기는 것은 타당하며, 특히 웹 요청과 같은 상황에서 유용합니다. 이를 통해 로그 메시지에 더 많은 컨텍스트 정보를 포함시켜 디버깅과 문제 해결을 용이하게 할 수 있습니다.

### 예시 코드

다음은 `zap` 로거를 사용하여 컨텍스트 정보를 로깅에 포함시키는 예시입니다.

```go
package main

import (
    "context"
    "go.uber.org/zap"
)

type ContextKey string

const (
    B3TraceIDKey ContextKey = "b3-trace-id"
    AmznTraceIDKey ContextKey = "amzn-trace-id"
)

// LogWithContext는 컨텍스트 정보를 포함하여 로그를 출력합니다.
func LogWithContext(ctx context.Context, logger *zap.Logger, message string) {
    fields := []zap.Field{}

    // B3 Trace ID 추가
    if traceID, ok := ctx.Value(B3TraceIDKey).(string); ok {
        fields = append(fields, zap.String(string(B3TraceIDKey), traceID))
    }

    // Amazon Trace ID 추가
    if traceID, ok := ctx.Value(AmznTraceIDKey).(string); ok {
        fields = append(fields, zap.String(string(AmznTraceIDKey), traceID))
    }

    logger.With(fields...).Info(message)
}

func main() {
    // zap 로거 초기화
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 컨텍스트 생성 및 값 설정
    ctx := context.Background()
    ctx = context.WithValue(ctx, B3TraceIDKey, "b3-trace-id-example")
    ctx = context.WithValue(ctx, AmznTraceIDKey, "amzn-trace-id-example")

    // 로깅 유틸리티 사용
    LogWithContext(ctx, logger, "This is a log message with context")
}
```

### 결론

컨텍스트를 로깅 유틸 함수에 넘기는 것은 Go 커뮤니티에서 일반적으로 받아들여지는 관행이며, 특히 웹 요청과 같은 상황에서 유용합니다. 이를 통해 로그 메시지에 더 많은 컨텍스트 정보를 포함시켜 디버깅과 문제 해결을 용이하게 할 수 있습니다. 따라서, 컨텍스트를 로깅 유틸 함수에 포함시키는 것은 좋은 설계라고 할 수 있습니다.

### 참고 자료

1. [Stack Overflow: Best way to pass context](https://stackoverflow.com/questions/72007434/best-way-to-pass-context)
2. [Gopher Academy Blog: Using Go's 'context' library for making your logs make sense](https://blog.gopheracademy.com/advent-2016/context-logging/)
3. [Better Stack Community Guide: Adding global context to a Logger](https://betterstack.com/community/guides/logging/zerolog/)

### 코드 설명

```go
package myappcontext

import (
    "context"

    "github.com/uber-go/zap"
)

type correlationIdType int
const (
    requestIdKey correlationIdType = iota
    sessionIdKey
)

var logger zap.Logger

func init() {
    // a fallback/root logger for events without context
    logger = zap.New(
        zap.NewJSONEncoder(zap.TimeFormatter(TimestampField)),
        zap.Fields(zap.Int("pid", os.Getpid()),
            zap.String("exe", path.Base(os.Args[0]))),
    )
}

// WithRqId returns a context which knows its request ID
func WithRqId(ctx context.Context, rqId string) context.Context {
    return context.WithValue(ctx, requestIdKey, requestId)
}

// WithSessionId returns a context which knows its session ID
func WithSessionId(ctx context.Context, sessionId string) context.Context {
    return context.WithValue(ctx, sessionIdKey, sessionId)
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) zap.Logger {
    newLogger := logger
    if ctx != nil {
        if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
            newLogger = newLogger.With(zap.String("rqId", ctxRqId))
        }
        if ctxSessionId, ok := ctx.Value(sessionIdKey).(string); ok {
            newLogger = newLogger.With(zap.String("sessionId", ctxSessionId))
        }
    }
    return newLogger
}
```

전역 변수 `logger`를 두고 `newLogger`로 다시 할당하는 것은 컨텍스트 정보를 포함한 새로운 로거 인스턴스를 생성하기 위한 방법입니다. 이를 통해 각 요청마다 다른 컨텍스트 정보를 포함한 로거를 생성할 수 있으며, 기본 로거 설정을 변경하지 않고 유연하게 로깅을 관리할 수 있습니다.

#### 전역 변수 `logger`

전역 변수 `logger`는 기본 로거로, 애플리케이션 전체에서 사용할 수 있는 기본 로거 인스턴스를 저장합니다. 이 로거는 초기화 시점에 설정된 기본 필드와 함께 생성됩니다.

```go
var logger zap.Logger

func init() {
    // a fallback/root logger for events without context
    logger = zap.New(
        zap.NewJSONEncoder(zap.TimeFormatter(TimestampField)),
        zap.Fields(zap.Int("pid", os.Getpid()),
            zap.String("exe", path.Base(os.Args[1]))),
    )
}
```

#### `Logger` 함수

`Logger` 함수는 주어진 컨텍스트를 기반으로 새로운 로거 인스턴스를 생성합니다. 이 함수는 다음과 같은 단계를 거칩니다:

1. **기본 로거 복사**:
   - `newLogger` 변수에 전역 변수 `logger`를 복사합니다. 이는 기본 로거를 기반으로 새로운 로거를 생성하기 위한 첫 단계입니다.

    ```go
    newLogger := logger
    ```

2. **컨텍스트 정보 추가**:
   - 주어진 컨텍스트에서 `requestIdKey`와 `sessionIdKey` 값을 추출하여, 이를 로거의 필드로 추가합니다.
   - 컨텍스트에 `requestIdKey`가 존재하면, `rqId` 필드를 추가합니다.
   - 컨텍스트에 `sessionIdKey`가 존재하면, `sessionId` 필드를 추가합니다.

    ```go
    if ctx != nil {
        if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
            newLogger = newLogger.With(zap.String("rqId", ctxRqId))
        }
        if ctxSessionId, ok := ctx.Value(sessionIdKey).(string); ok {
            newLogger = newLogger.With(zap.String("sessionId", ctxSessionId))
        }
    }
    ```

3. **새로운 로거 반환**:
   - 컨텍스트 정보를 포함한 새로운 로거 인스턴스를 반환합니다.

    ```go
    return newLogger
    ```

### 의미와 장점

1. **컨텍스트 기반 로깅**:
   - 각 요청마다 다른 컨텍스트 정보를 포함한 로거를 생성할 수 있습니다. 이를 통해 로그 메시지에 더 많은 컨텍스트 정보를 포함시켜 디버깅과 문제 해결을 용이하게 할 수 있습니다.

2. **기본 로거 유지**:
   - 전역 변수 `logger`는 기본 로거로 유지되며, 이를 기반으로 새로운 로거를 생성합니다. 이는 기본 로거 설정을 변경하지 않고, 각 요청마다 동적으로 로거를 생성할 수 있게 합니다.

3. **유연성**:
   - 컨텍스트에 따라 로거에 추가할 필드를 동적으로 결정할 수 있습니다. 이는 다양한 상황에 맞게 로거를 구성할 수 있게 합니다.

### 예시

다음은 `Logger` 함수를 사용하여 컨텍스트 정보를 포함한 로거를 생성하고 로그 메시지를 출력하는 예시입니다.

```go
package main

import (
    "context"
    "myappcontext"
    "go.uber.org/zap"
)

func main() {
    // zap 로거 초기화
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 컨텍스트 생성 및 값 설정
    ctx := context.Background()
    ctx = myappcontext.WithRqId(ctx, "request-id-example")
    ctx = myappcontext.WithSessionId(ctx, "session-id-example")

    // 로깅 유틸리티 사용
    loggerWithCtx := myappcontext.Logger(ctx)
    loggerWithCtx.Info("This is a log message with context")
}
```

## 기타

- [context를 인자로 넘기를 로깅 함수](https://blog.gopheracademy.com/advent-2016/context-logging/)
