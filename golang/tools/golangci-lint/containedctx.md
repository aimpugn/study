# containedctx

## [`containedctx`](https://github.com/sivchari/containedctx)

`containedctx` is a linter that detects struct contained `context.Context` field.
This is discouraged technique in favour of passing context as first argument of method or function.

For rationale please read [Contexts and structs](https://go.dev/blog/context-and-structs) the Go blog post.

## [Contexts and structs](https://go.dev/blog/context-and-structs)

다른 타입에 저장하는 것보다 컨텍스트를 파라미터로 전달하는 것이 왜 중요한지 그 이유를 예시를 통해 설명합니다.
또한 구조체 유형에 컨텍스트를 저장하는 것이 합리적일 수 있는 드문 경우와 안전하게 저장하는 방법을 강조합니다.

```go
// Worker fetches and adds works to a remote work orchestration server.
type Worker struct { /* … */ }

type Work struct { /* … */ }

func New() *Worker {
  return &Worker{}
}

func (w *Worker) Fetch(ctx context.Context) (*Work, error) {
  _ = ctx // A per-call ctx is used for cancellation, deadlines, and metadata.
}

func (w *Worker) Process(ctx context.Context, work *Work) error {
  _ = ctx // A per-call ctx is used for cancellation, deadlines, and metadata.
}
```

이렇게 함수의 인자로 컨텍스트를 넘기면 다음과 같은 장점이 있습니다:
- 사용자는 호출별 deadline, cancellation 및 메타데이터를 설정할 수 있습니다.
- 각 메서드에 전달된 `context.Context`가 어떻게 사용되는지도 명확합니다. 즉, 한 메서드에 전달된 `context.Context`가 다른 메서드에서 사용될 것이라고 예상할 수 없습니다.

```go
type Worker struct {
  ctx context.Context
}

func New(ctx context.Context) *Worker {
  return &Worker{ctx: ctx}
}

func (w *Worker) Fetch() (*Work, error) {
  _ = w.ctx // A shared w.ctx is used for cancellation, deadlines, and metadata.
}

func (w *Worker) Process(work *Work) error {
  _ = w.ctx // A shared w.ctx is used for cancellation, deadlines, and metadata.
}
```

반면 이렇게 구조체의 필드로 컨텍스트를 저장하고 재사용할 때 다음과 같은 단점이 있습니다.
- 컨텍스트가 구조체 안에 가려져 있기 때문에, 함수를 호출하는 사용자에게 컨텍스트의 생애주기가 모호해집니다.
