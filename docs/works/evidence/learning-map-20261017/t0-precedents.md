# 공통 이벤트 형식 선례 조사 (T0)

확인 날짜: 2026-10-17
목적: place / run / event / link / note 후보 표를 설계하기 전에, 기존 형식에서 무엇을 빌리고 무엇을 버릴지 1차 자료로 확인한다.

## 0. 읽는 법과 신뢰 한계

표시는 세 가지다.

- **[확인]**: 이번 조사에서 해당 페이지를 가져와 그 내용이 나온 것을 봤다.
- **[추론]**: 확인한 사실에서 내가 끌어낸 판단이다. 원문이 말하지 않는다.
- **[못 봄]**: 가져오지 못했거나, 요약기가 찾지 못했다. 원문에 없다고 단정하지 않는다.

신뢰 한계는 아래와 같다.

1. **모든 인용과 앵커는 WebFetch 요약기를 거쳤다.** 원본 HTML을 직접 대조하지 않았다. 요약기가 "앵커는 제목에서 추정했다"고 밝힌 곳은 해당 줄에 **(앵커 추정)** 이라고 적었다. 앵커가 `#...`로 적혀 있어도 HTML 소스와 맞춰 보지 않았으므로, 문서에 인용하기 전에 한 번 열어 확인해야 한다.
2. 인용은 소스당 1개, 15단어 이하로 맞췄다. 요약기가 의역한 문장은 인용으로 쓰지 않고 "요약"으로 적었다.
3. 필드 이름과 필드 번호는 요약기가 원문대로 옮겼다고 보고한 것만 적었다. 주석은 의역이다.
4. 사용한 도구는 WebFetch와 WebSearch뿐이다. 브라우저 도구, git, 파일 다운로드는 쓰지 않았다. 쓴 파일은 이 파일 하나다.

---

## 1. OpenTelemetry

### S1. Trace API 명세

- 제목·URL: Tracing API — https://opentelemetry.io/docs/specs/otel/trace/api/ (절: `#span`, `#spancontext`, `#span-creation`, `#spankind`, `#link`, `#add-events`, `#set-status`, `#span-lifetime`)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** 단 이 페이지에는 "Links between spans" 개요 절이 없다. 그 절은 S2에서 봤다.

확인한 내용 (필드 이름은 원문 그대로):

- Span이 가진 것: span name, `SpanContext`(불변), parent(`Span`·`SpanContext`·null 중 하나), `SpanKind`, start timestamp, end timestamp, `Attributes`, `Events`(timestamped 목록), `Links`(다른 Span으로의 목록), `Status`.
- `SpanContext`: `TraceId`(16바이트, 0이 아닌 바이트가 하나 이상), `SpanId`(8바이트), `TraceFlags`, `TraceState`.
- 생성 시 parent는 `Context` 전체여야 하고 Span이나 SpanContext를 직접 받지 않는다. `SpanKind`를 생략하면 `Internal`이다.
- 시간 정밀도: 최소 밀리초, 최대 나노초.
- `Status`: `Unset`, `Ok`, `Error`. 요약기가 "Ok > Error > Unset 전순서"라고 보고했다.
- Link: 같은 trace의 SpanContext도, 다른 trace의 SpanContext도 가리킬 수 있고 attributes를 가진다.

인용: "Linked SpanContexts can be from the same or a different trace" (`#link`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | span = (이름, 시작, 끝, attributes, parent) 한 레코드 | 한 호출을 한 행으로 두면 enter/exit 두 줄을 나중에 짝지을 필요가 없다. [추론] |
| 빌릴 것 | Event(시각이 붙은 점)와 Link(다른 span 참조)를 span 본체와 구분하는 개념 | 우리의 instant와 link에 대응한다. 단 대응 방식은 D4, D5에서 따로 정한다. |
| 버릴 것 | 16바이트 `TraceId`와 8바이트 `SpanId` | 분산 환경에서 조정 없이 유일성을 얻으려는 장치다. 우리는 한 번에 한 기록을 오프라인으로 적재하므로 정수 대리키로 충분하다. [추론] |
| 버릴 것 | `TraceFlags`(sampled), `TraceState` | 샘플링 전파용이다. 우리는 이미 기록한 것을 적재한다. [추론] |
| 버릴 것 | Span에 붙은 Event를 독립 이벤트로 쓰지 않는 구조 | OTel Event는 Span에 매달려 있다. 우리 instant(syscall 하나, 메시지 하나)는 감싸는 span 없이도 존재해야 한다. [추론] |

### S2. OpenTelemetry 개요 (Links between spans)

- 제목·URL: Overview — https://opentelemetry.io/docs/specs/otel/overview/ (절: `#links-between-spans`, `#spans`, `#spancontext`, `#context-propagation`)
- 확인일: 2026-10-17
- 확인 결과: **확인함.**

확인한 내용:

- Link는 부모-자식이 아닌 인과 관계를 표현한다. 요약: 한 span이 "0개 이상의 다른 Span과 연결될 수 있다".
- 쓰임새 세 가지(요약): 배치 처리처럼 여러 입력이 한 span을 시작하는 경우, 서로 다른 trace를 잇는 경우(들어온 컨텍스트를 신뢰하지 않고 새 trace를 시작할 때), scatter/gather 같은 fan-in.
- 요약: parent 필드는 "단일 부모" 시나리오를 나타내고, 많은 경우 부모 span이 자식 span을 완전히 감싼다. fan-in에서는 이 전제가 성립하지 않는다.
- `SpanContext`는 자식 Span과 프로세스 경계를 넘어 전파되어야 한다(MUST).

인용: "the parent Span fully encloses the child Span" (`#links-between-spans`, 요약기가 잘라낸 구절)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 「parent는 포함 관계, link는 인과 관계」라는 구분 | 우리의 호출 트리는 포함 관계(parent)이고, 메시지 송수신·syscall→커널 진입은 서로 다른 place 사이의 인과 관계(link)다. 이 구분이 link 표를 따로 두는 근거가 된다. [추론] |
| 빌릴 것 | link의 개수 제한이 없다는 점(0개 이상) | 한 메시지가 여러 곳에서 소비되는 경우를 행 여러 개로 표현할 수 있다. [추론] |
| 버릴 것 | 「link는 걸린 span에만 저장되고 반대편은 모른다」는 단방향 저장 | 우리 지도는 양쪽 어느 쪽에서 봐도 상대편이 보여야 한다. 표 하나에 (src, dst) 쌍으로 두는 편이 낫다. [추론] |

### S3. Context Propagators

- 제목·URL: Propagators API — https://opentelemetry.io/docs/specs/otel/context/api-propagators/ (절: `#inject`, `#extract` 확인 / TextMap Propagator, Global Propagators, W3C Trace Context Requirements **(앵커 추정)**)
- 확인일: 2026-10-17
- 확인 결과: **확인함.**

확인한 내용:

- 상태를 다음 프로세스로 옮기는 장치가 `Propagator`다. `Inject`는 Context의 값을 carrier(예: HTTP 헤더)에 쓰고, `Extract`는 carrier에서 읽어 Context에 넣는다.
- `Extract`는 값을 파싱하지 못해도 예외를 던져서는 안 된다(MUST NOT).
- 기본 전파 형식은 W3C Trace Context의 `traceparent`와 `tracestate`다.

인용: "Injects the value into a carrier." (`#inject`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 「양쪽 프로세스가 같은 ID를 공유하면 link가 확정된다」는 원리를 link의 한 가지 근거로만 | 직접 계측한 구간에서는 ID 전파로 link를 확정할 수 있다. |
| 버릴 것 | 전파 메커니즘 자체 | 우리는 계측을 코드에 넣어 ID를 실어 보내는 쪽이 아니라, 이미 일어난 일을 사후에 짝짓는 쪽이다. PostgreSQL 와이어 프로토콜에 trace 컨텍스트를 실을 자리가 있는지는 이번에 보지 않았다. [못 봄] 따라서 link마다 「어떻게 짝지었는지」를 남기는 열이 필요하다는 것이 [추론]이다 (D5). |

### S4. OTLP trace.proto

- 제목·URL: opentelemetry-proto `opentelemetry/proto/trace/v1/trace.proto` — https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/trace/v1/trace.proto (GitHub 화면 주소: https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** `main` 브랜치를 가져왔고 커밋 해시는 기록하지 못했다. 요약기가 주석을 의역했다고 밝혔으므로 필드 이름·번호·타입만 신뢰한다.

확인한 내용:

- 중첩: `TracesData.resource_spans` → `ResourceSpans`(`resource`, `scope_spans`, `schema_url`) → `ScopeSpans`(`scope`, `spans`, `schema_url`).
- `Span`의 필드와 번호:

| 필드 | 번호 | 타입 |
| --- | --- | --- |
| `trace_id` | 1 | bytes (16바이트) |
| `span_id` | 2 | bytes (8바이트) |
| `trace_state` | 3 | string |
| `parent_span_id` | 4 | bytes (8바이트) |
| `flags` | 16 | fixed32 |
| `name` | 5 | string |
| `kind` | 6 | SpanKind |
| `start_time_unix_nano` | 7 | fixed64 |
| `end_time_unix_nano` | 8 | fixed64 |
| `attributes` | 9 | repeated KeyValue |
| `dropped_attributes_count` | 10 | uint32 |
| `events` | 11 | repeated Span.Event |
| `dropped_events_count` | 12 | uint32 |
| `links` | 13 | repeated Span.Link |
| `dropped_links_count` | 14 | uint32 |
| `status` | 15 | Status |

- `SpanKind` 값: `SPAN_KIND_UNSPECIFIED = 0`, `SPAN_KIND_INTERNAL = 1`, `SPAN_KIND_SERVER = 2`, `SPAN_KIND_CLIENT = 3`, `SPAN_KIND_PRODUCER = 4`, `SPAN_KIND_CONSUMER = 5`.
- `Span.Event`: `time_unix_nano`, `name`, `attributes`, `dropped_attributes_count`.
- `Span.Link`: `trace_id`, `span_id`, `trace_state`, `attributes`, `dropped_attributes_count`, `flags`.
- `Status`: `message`, `code`(`STATUS_CODE_UNSET = 0`, `STATUS_CODE_OK = 1`, `STATUS_CODE_ERROR = 2`).
- 시간 필드 주석(요약): UNIX epoch 이후 나노초. 속성 키는 유일해야 한다.

인용: "If this is a root span, then this field must be empty." (`Span.parent_span_id` 주석)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | parent를 depth가 아니라 포인터(`parent_span_id`)로만 저장하고, 루트는 빈 값 | OTLP는 깊이를 저장하지 않는다. 포인터가 진실이고 깊이는 파생값이다. |
| 빌릴 것 | 시각을 정수 나노초(`fixed64`)로 두는 방식 | 정수라서 정렬·차이 계산에 오차가 없다. 단 epoch 기준은 D1에서 따로 따진다. |
| 빌릴 것 | `attributes` = 키가 유일한 키-값 목록 | 소스마다 필드가 다른 부분(syscall 인자, 와이어 메시지 필드)을 구조 열이 아니라 확장 열에 둔다. |
| 빌릴 것 | `SpanKind`의 `PRODUCER`/`CONSUMER`/`CLIENT`/`SERVER` 어휘 | 메시지 송신·수신 link의 방향 표식 후보다. 원래 의미는 span의 성격이지 link 방향이 아니므로 적용은 [추론]이다. |
| 버릴 것 | `dropped_attributes_count`, `dropped_events_count`, `dropped_links_count` | 전송 중 버려진 수를 알리는 필드다. 우리는 기록 전체를 보관한다. 단 우리 수집기가 값을 잘라 담으면 그 사실을 attributes에 남겨야 한다. [추론] |
| 버릴 것 | `trace_state`, `flags`, `schema_url` | 분산 전파·스키마 버전용이다. |
| 버릴 것 | `ResourceSpans → ScopeSpans` 고정 2단계 | 우리 place는 system→host→process→component→function으로 더 깊고 가변이다. |

---

## 2. Perfetto

### S5. PerfettoSQL 시작하기 (트랙·슬라이스 개념)

- 제목·URL: Getting Started with PerfettoSQL — https://perfetto.dev/docs/analysis/perfetto-sql-getting-started (절: `#tracks`, Events 아래 Slices, `#thread-and-process-identifiers`, `#timestamps-are-in-nanoseconds`, `#operator-tables`)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** Slices 소절에는 요약기가 앵커를 보고하지 않았다.

확인한 내용:

- 슬라이스는 시간 구간이다. 트랙은 "이벤트의 이름 붙은 분할"이고 스레드·프로세스·CPU 같은 맥락에 이벤트를 묶는다.
- 요약: 스레드 트랙은 스레드 하나에 속하고 슬라이스를 담는다. 동기 userspace 슬라이스는 이벤트를 낸 스레드마다 트랙이 하나 있다.
- 요약: 같은 트랙 안에서 `ancestor_slice`는 조상 슬라이스를, `descendant_slice`는 중첩된 후손을 돌려준다. 조상·후손 계산이 **같은 트랙 안에서만** 이뤄진다.
- 요약: OS의 `tid`·`pid`는 trace 안에서 유일한 식별자로 쓸 수 없어 `utid`·`upid`를 쓴다.
- 모든 `ts`와 `dur`는 나노초다 (`#timestamps-are-in-nanoseconds`).
- `DIRECTLY_CONNECTED_FLOW`, `FOLLOWING_FLOW`, `PRECEDING_FLOW` 연산자 표가 있다(Operator tables 아래). flow 표 컬럼은 S6에서 봤다.

인용: "A track is a named partition of events" (`#tracks`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 「중첩의 범위는 트랙 하나」라는 규칙 | 우리 event의 parent·depth는 같은 실행 맥락(스레드에 해당하는 place) 안에서만 의미가 있다고 명시한다. [추론] |
| 빌릴 것 | OS id(pid/tid)를 키로 쓰지 않고 별도 대리키(`utid`/`upid`)를 두는 규칙 | place 표의 핵심 근거다. D7에서 쓴다. |
| 빌릴 것 | 연결 추적 연산자 형태 | 「이 syscall에서 이어진 커널 호출 트리 전체」 같은 질의가 필요하다. 구현은 우리 몫이다. [추론] |
| 버릴 것 | 「이 UI의 시각적 행은 trace processor 트랙과 1:1이 아니다」 쪽의 UI 전용 트랙 개념 | 렌더링 문제다. 저장 모델에 끌어오지 않는다. |

### S6. PerfettoSQL Prelude (slice/thread/process/flow 컬럼)

- 제목·URL: PerfettoSQL Prelude — https://perfetto.dev/docs/analysis/sql-tables (절: `#slice`, `#thread`, `#process`, `#flow`)
- 확인일: 2026-10-17
- 확인 결과: **일부.** 이 페이지에는 `thread_track`·`track`·`counter` 정의가 없다고 요약기가 보고했다. 그 표는 S7에서 봤다.

확인한 내용 (컬럼은 요약기가 원문 그대로 옮겼다고 보고한 것):

- `slice`: `id`, `ts`, `dur`, `track_id`, `category`, `name`, `depth`, `parent_id`, `arg_set_id`, `thread_ts`, `thread_dur`, `thread_instruction_count`, `thread_instruction_delta`
- `thread`: `utid`, `tid`, `name`, `start_ts`, `end_ts`, `upid`, `is_main_thread`, `is_idle`, `machine_id`, `arg_set_id`
- `process`: `upid`, `pid`, `name`, `start_ts`, `end_ts`, `parent_upid`, `uid`, `android_appid`, `android_user_id`, `cmdline`, `arg_set_id`, `machine_id`
- `flow`: `id`, `slice_out`, `slice_in`, `trace_id`, `arg_set_id`

인용: "The depth of the slice in the current stack of slices." (`slice.depth`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | **`parent_id`와 `depth`를 둘 다 저장** | Perfetto 슬라이스가 두 열을 모두 가진다. 포인터가 진실이고 깊이는 「줌 수준별 필터」용 비정규화 열이다. [확인: 두 열이 함께 있음 / 추론: 용도] |
| 빌릴 것 | `flow` = (`slice_out`, `slice_in`) 쌍 + `arg_set_id` | 방향이 있는 쌍 표다. 우리 link의 가장 가까운 선례다. |
| 빌릴 것 | `process.parent_upid`, `thread.upid` 같은 FK 계층과 `start_ts`/`end_ts` | place 계층을 외래키로 표현하고 생존 구간을 둔다. |
| 빌릴 것 | `arg_set_id` = 인자 묶음을 가리키는 간접 참조 | 인자를 별도 표로 빼는 방식이다. 같은 묶음을 공유하는지는 읽은 범위에 없다. [못 봄] |
| 버릴 것 | `thread_ts`, `thread_dur`, `thread_instruction_*` | 스레드 CPU 시간·명령어 카운터 수집이 있어야 채워진다. 우리 기록에는 소스가 없다. |
| 버릴 것 | `android_*`, `machine_id` 등 Android·다중 머신 전용 열 | 현재 범위 밖이다. |
| 버릴 것 | process·thread·track 분리 표 구조 | 우리 계층은 시스템부터 함수까지 깊어서 place 한 표(인접 목록)가 더 자연스럽다. [추론] |

### S7. PerfettoSQL 표준 라이브러리 (prelude 패키지)

- 제목·URL: PerfettoSQL standard library — https://perfetto.dev/docs/analysis/stdlib-docs (절: `#package-prelude`)
- 확인일: 2026-10-17
- 확인 결과: **일부.** 페이지 전체가 548,997자이고 앞 100,000자만 읽었다. prelude의 표별 앵커(`#slice` 등)는 페이지에서 보이지 않는다고 요약기가 보고했다.

확인한 내용:

- `track`: `id`, `name`, `type`, `dimension_arg_set_id`, `parent_id`, `source_arg_set_id`, `machine_id`, `track_group_id`
- `thread_track`: `id`, `name`, `type`, `parent_id`, `source_arg_set_id`, `machine_id`, `utid`
- `flow`: 슬라이스를 잇는 흐름 이벤트를 담는 표라고 설명한다. (요약)
- `counter`: 트레이스를 파싱하는 동안 트랙에 놓인 값을 담는 표라고 설명한다. (요약)
- 단위: `ts`·`dur`는 나노초. 요약: `ts`의 실제 기준은 TraceConfig의 `primary_trace_clock`에 따라 다르다.
- `slice.dur`가 끝나지 않은 슬라이스에서 어떤 값인지는 이 페이지에 없다. 센티널(-1)은 S9의 strace 절에서만 나온다.

인용: "The timestamp at the start of the slice in nanoseconds." (`slice.ts`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 시각의 기준 시계를 trace 설정에 둔다는 구조 | 소스마다 시계가 다르다(S9: strace는 epoch 초, S16: ftrace는 마이크로초 소수). run에 기준 시계를 기록해 두는 근거다. [추론] |
| 빌릴 것 | `counter` = 트랙에 놓인 값의 열 | `pg_stat_activity` 샘플 같은 관측의 선례 후보다. D10. |
| 버릴 것 | `track_group_id`, `dimension_arg_set_id` | UI 묶기와 트랙 구분용이다. |

### S8. Trace Processor 문서 (일부)

- 제목·URL: Trace Processor — https://perfetto.dev/docs/analysis/trace-processor (절: `#shell`, `#running-the-shell`, `#embedding`, `#executing-queries`)
- 확인일: 2026-10-17
- 확인 결과: **일부.** 요청한 `docs/concepts/trace-processor`는 **404**였다. 위 페이지에서는 `slice`와 `counter` 표 예시만 나왔고 `thread_track`·`process`·`thread`는 설명하지 않는다고 요약기가 보고했다. 해당 표의 정의는 S6, S7에서 확인했다.

확인한 내용: 라이브러리가 "일관된 표 집합으로 SQL 인터페이스를 노출한다"는 취지의 문장, `SELECT ts, dur, name FROM slice LIMIT 10;`, `SELECT ts, value FROM counter LIMIT 10;` 예시.

인용: "exposes an SQL interface for querying them through a consistent set of tables" (페이지 도입부)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 「파서가 여러 입력을 같은 표 집합에 넣고 질의는 SQL로」라는 구조 | 우리 사정과 같다. 소스가 다섯 가지여도 질의 쪽은 표 다섯 개만 본다. |

### S9. Perfetto 외부 형식 가져오기

- 제목·URL: Visualizing external trace formats with Perfetto — https://perfetto.dev/docs/getting-started/other-formats (절: `#chrome-json-format`, `#strace-format`, `#collapsed-stack-format`, `#linux-ftrace-textual-format`)
- 확인일: 2026-10-17
- 확인 결과: **확인함** (위 네 절만 읽음).

확인한 내용:

- **Chrome JSON**: UI에서 바로 열린다. 파서가 `slice`, `track`, `process`, `thread`, `counter`, `args` 표로 채운다. `B/E`와 `X`는 슬라이스로, `I`(대문자라고 요약기가 보고)는 시간선의 점으로, `C`는 카운터 트랙으로, `M`은 프로세스·스레드 이름으로, `s/t/f`는 슬라이스를 잇는 화살표로 가져온다. `ts`는 마이크로초. **기간 이벤트(B/E, X)는 중첩되어야 하고, 겹쳐서 중첩되지 않으면 별도의 overflow 트랙에 놓인다.**
- **strace** (`#strace-format`): `-ttt`(epoch 초), `-f`(pid 출력)가 필수, `-T`는 권장. 완료된 호출은 `strace` 카테고리의 슬라이스로 호출한 스레드의 트랙에 놓인다. `-T`가 없으면 호출 진입 시각만 남아 길이 0의 표식이 된다. 블로킹 호출은 `<unfinished ...>`과 `<... resumed>` 줄 사이가 하나의 슬라이스가 된다. **끝나지 않은 호출은 `dur` -1로 열린 채 남는다.**
- **Collapsed Stack** (`#collapsed-stack-format`): 한 줄이 스택 하나, 「루트 프레임이 먼저, 리프가 마지막」, 줄 끝에 공백으로 구분된 양의 정수 횟수, `#`으로 시작하면 주석. 시간 보존 여부와 대상 표는 페이지에 없다고 요약기가 보고했다. 결과는 flamegraph로 보인다.
- **ftrace 텍스트** (`#linux-ftrace-textual-format`): 「호환성 목적, best-effort」로 지원한다고 한다. `sched_switch` 등을 다루고, `function_graph` 지원은 이 절에서 확인할 수 없다.

인용 (strace 절): "A call still unfinished when the log ends stays open (a dur of -1)."

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | **열린 구간(끝나지 않은 호출)을 표현할 수 있어야 한다** | strace 가져오기가 그렇게 한다. 센티널은 `dur = -1`이다. 우리는 NULL이 더 명확하다고 본다 (D2). |
| 빌릴 것 | 진입 시각만 있는 호출을 길이 0 표식으로 취급 | 같은 syscall이 `-T` 유무에 따라 span이 되기도 점이 되기도 한다. 한 `event` 표에서 span과 instant를 같은 행으로 다뤄야 하는 이유다 (D4). |
| 빌릴 것 | Chrome JSON으로 내보내 Perfetto UI에서 열기 | 적재와 중첩 계산이 맞는지 보는 외부 검증 도구로 쓸 수 있다. D13. |
| 빌릴 것 | 접힌 스택(루트→리프, 횟수)을 **내보내기 형식**으로 | 저장 형식이 아니라 파생 출력이다 (D12). |
| 버릴 것 | 겹치는 비중첩 기간 이벤트를 허용하지 않는 제약 | 호출 트리는 중첩이지만 메시지 전달 구간은 비중첩이다. 후자는 span이 아니라 link로 둔다. [추론] |
| 버릴 것 | ftrace 텍스트 가져오기 경로 | `function_graph`를 지원한다는 근거가 이 절에 없다. 우리는 자체 파서가 필요하다. [못 봄] |

---

## 3. Chrome Trace Event Format

### S10. Trace Event Format (Google 문서)

- 제목·URL: Trace Event Format — https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU/mobilebasic (S9가 같은 문서를 `/preview` 주소로 가리킨다)
- 확인일: 2026-10-17
- 확인 결과: **확인함. 단 `/mobilebasic` 주소로만.** 접근을 시도한 네 가지 주소는 다음과 같다.

| 시도한 주소 | 결과 |
| --- | --- |
| `/preview` | 제목만 반환되고 본문이 없었다 |
| `/export?format=txt` | 307 리다이렉트 후 리다이렉트 대상에서 **HTTP 400** |
| `/pub` | **HTTP 401** |
| `/mobilebasic` | **성공**. 이 절의 모든 내용은 이 주소에서 나왔다 |

확인한 내용 (필드 이름은 원문 그대로):

- 형식 두 가지: **JSON Array Format**(이벤트 객체의 배열)과 **JSON Object Format**(필수 `traceEvents`, 선택 `displayTimeUnit`, `otherData`, `stackFrames`, `samples`).
- 공통 필드: `name`, `cat`, `ph`, `ts`, `tts`(선택), `pid`, `tid`, `args`, 선택 `cname`. `dur`은 Complete 이벤트의 추가 매개변수.
- `ts` 단위는 마이크로초. 예시에 소수가 있다: `"ts": 1234523.3`. `displayTimeUnit`은 표시용이다.
- Phase: 기간 `B`/`E`, 완전 `X`, 인스턴트 `i`(대문자 `I`는 Deprecated 표시), 카운터 `C`, 비동기 `b`/`n`/`e`(중첩 가능), 흐름 `s`/`t`/`f`, 메타데이터 `M`, 샘플·객체·메모리 덤프 등.
- 기간 이벤트(`B`/`E`)는 **반드시 중첩되어야 하고**(요약기 보고), 한 스레드 안에서 타임스탬프가 증가해야 한다.
- Complete(`X`)는 `B`/`E` 쌍을 하나로 합친 것이며 `dur`을 가진다.
- Instant의 범위는 `s` 필드: 글로벌 `g`, 프로세스 `p`, 스레드 `t`. 예시: `{"name": "OutOfMemory", "ph": "i", "ts": 1234523.3, "pid": 2343, "tid": 2347, "s": "g"}`. 문서가 「네 가지 범위」라고 쓰면서 셋만 나열한다고 요약기가 보고했다.
- Flow: `s`(시작), `t`(단계), `f`(끝). 바인딩: `s`와 `t`는 포함하는 슬라이스에 붙는다. `f`는 다음 슬라이스에 붙는데, `bp="e"`면 포함하는 슬라이스에 붙는다. 서로 어떻게 연결되는지(같은 `id` 등)에 대한 서술은 요약기가 찾지 못했다. [못 봄 — 원문에 없다고 단정하지 않는다]
- Async 이벤트는 같은 category와 `id`를 가진 것끼리 묶인다. `id`는 기본적으로 프로세스 간에 전역이고 `scope`로 충돌을 피한다. 객체 이벤트의 `id`는 기본적으로 프로세스 지역이다. `id2`가 추가되어 지역·전역을 명시한다.
- Metadata(`M`): `process_name`, `process_labels`, `process_sort_index`, `thread_name`, `thread_sort_index`. 예시: `{"name": "thread_name", "ph": "M", "pid": 2343, "tid": 2347, "args": {"name" : "RendererThread"}}`.

인용: "The timestamps are provided at microsecond granularity."

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | **instant와 span이 별개 phase** (`i` vs `B/E/X`) | 점과 구간이 개념적으로 다름을 형식이 보여 준다. 다만 저장은 한 표에 두고 `kind` 열로 구분한다 (D4). |
| 빌릴 것 | instant의 범위 `s`(g/p/t) | 점이 어느 층위에 속하는지가 곧 place다. 우리는 범위를 별도 열로 두지 않고 `place_id`로 표현한다. [추론] |
| 빌릴 것 | `X` = 시작과 길이를 한 행에 | 저장 단위는 한 행이다. 입력이 `B`/`E` 쌍이어도 적재 때 합친다. |
| 빌릴 것 | `args` = 자유 형식 객체 | attributes 확장 열의 선례다. |
| 빌릴 것 | `M` 메타데이터로 pid/tid의 이름을 따로 붙이는 방식 | 이름·표시 정보를 이벤트와 분리해 place에 둔다. |
| 빌릴 것 | flow에서 「끝점이 어느 슬라이스에 붙는가」가 `bp`에 따라 달라진다는 점이 보여 주는 위험 | 우리는 link 양끝을 적재 시점에 event id로 확정해 이 모호함을 피한다. [추론] |
| 버릴 것 | `pid`/`tid` 정수 쌍을 장소의 정체성으로 쓰는 것 | 재사용되고 컨테이너 간 충돌한다. S5의 `utid` 근거와 같다. |
| 버릴 것 | 마이크로초 `ts` | 정밀도가 ftrace·JFR에 못 미칠 수 있다 (D1). |
| 버릴 것 | `cat` 콤마 구분 문자열 | 필터용 문자열이다. 질의에는 단일값 열이 낫다. [추론] |
| 버릴 것 | `tts`, `cname`, 샘플·객체·메모리 덤프 등 | 우리 소스에 대응물이 없거나 렌더링용이다. |

---

## 4. 플레임 그래프

### S11. Flame Graphs (Brendan Gregg)

- 제목·URL: Flame Graphs — https://www.brendangregg.com/flamegraphs.html (절: `#Summary`, `#Variations`)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** 접힌 스택 입력 형식은 이 페이지에 없다고 요약기가 보고했다. 그 부분은 S12에서 봤다. 별도의 "Flame Charts" 절은 없고 Variations 아래에서 다룬다.

확인한 내용:

- x축은 알파벳 순으로 정렬되며 시간의 흐름이 아니다. y축은 스택 깊이이며 맨 아래가 0이다.
- 플레임 **차트**는 x축에 알파벳 대신 시간의 흐름을 둔다.

인용: "sorted alphabetically (it is not the passage of time)" (`#Summary`)

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 「플레임 그래프는 집계, 플레임 차트는 시간」이라는 구분 | 우리 `event` 표는 시간순 개별 호출이므로 **플레임 차트의 원자료**다. 플레임 그래프는 여기서 파생한 집계 뷰다. |
| 버릴 것 | 플레임 그래프 모양을 저장 형식으로 쓰는 것 | 집계하면 시간·순서·개별 호출이 사라진다. 지도는 개별 호출의 맥락이 필요하다. [추론] |

### S12. FlameGraph 저장소 README

- 제목·URL: brendangregg/FlameGraph README — https://raw.githubusercontent.com/brendangregg/FlameGraph/master/README.md (GitHub 화면 주소: https://github.com/brendangregg/FlameGraph)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** `master`를 가져왔고 커밋 해시는 기록하지 않았다. 접힌 형식은 README의 예시 줄에서만 확인했다. 요약기가 「형식을 한 문장으로 정의한 곳은 없다」고 보고했다.

확인한 내용:

- 3단계: 스택 캡처 → `stackcollapse*.pl`로 한 줄씩 접기 → `flamegraph.pl`로 SVG 생성.
- 접힌 줄의 모양(예시에서 읽음): 프레임을 `;`로 이어 쓰고, 공백 뒤에 횟수. 예: ``unix`_sys_sysenter_post_swapgs;genunix`close 5``
- `flamegraph.pl` 옵션 중 `--flamechart`는 「시간순 정렬, 스택 병합 없음」(요약기 표현)이다.

인용: "Use the stackcollapse programs to fold stack samples into single lines."

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | 접힌 스택 = 내보내기 대상 | `event`의 parent 사슬을 따라 루트→리프 경로를 만들고 같은 경로를 세면 접힌 줄이 된다. 변환이 단순하다. [추론] |
| 빌릴 것 | `--flamechart` 모드의 존재 | 같은 도구가 병합 없는 시간순 뷰를 지원한다는 증거다. 시간순 원자료가 먼저라는 판단의 보강이다. |
| 버릴 것 | 접힌 줄을 1차 저장 형식으로 쓰는 것 | S11의 이유와 같다. |

---

## 5. git 내용 주소 저장

### S13. Pro Git — Git Objects

- 제목·URL: Git Internals - Git Objects — https://git-scm.com/book/en/v2/Git-Internals-Git-Objects (절: `#git-objects`, `#_tree_objects`, `#_commit_objects`, `#_object_storage` — 마지막 세 앵커는 **(앵커 추정)**)
- 확인일: 2026-10-17
- 확인 결과: **일부.** 「같은 내용은 한 번만 저장된다」는 문장은 이 페이지에서 찾지 못했다. 요약기가 「설계에서 함축될 뿐 명시한 곳이 없다」고 보고했다.

확인한 내용:

- Git은 내용 주소 파일시스템이며 핵심은 단순한 키-값 저장소다.
- `git hash-object`의 키는 「저장하려는 내용에 헤더를 붙인 것」의 체크섬이다. 헤더는 `blob <size>\0`이다 (`#_object_storage`의 예시에서 `blob 16\0what is up, doc?`).
- `-w`로 쓰면 `.git/objects/` 아래 해시 앞 2글자 디렉터리 안에 저장된다.
- 객체 종류: blob(파일 내용만, 파일명 없음), tree(항목마다 mode·type·이름), commit(최상위 tree·부모·메타데이터).
- 저장 전에 zlib으로 압축한다.

인용: "Git is a content-addressable filesystem."

### S14. Pro Git — Git이란 무엇인가

- 제목·URL: Getting Started - What is Git? — https://git-scm.com/book/en/v2/Getting-Started-What-is-Git%3F (절: `#snapshots-not-differences`, `#git-has-integrity`)
- 확인일: 2026-10-17
- 확인 결과: **확인함.** S13에서 못 찾은 문장이 여기에 있었다.

확인한 내용:

- 파일이 바뀌지 않았으면 Git은 그 파일을 다시 저장하지 않고, 이미 저장한 같은 파일로의 링크만 둔다.
- Git은 모든 것을 저장하기 전에 체크섬을 낸다(SHA-1). 데이터베이스에는 파일 이름이 아니라 내용의 해시값으로 저장한다.

인용: "if files have not changed, Git doesn't store the file again" (`#snapshots-not-differences`)

### S15. git 용어집

- 제목·URL: gitglossary — https://git-scm.com/docs/gitglossary (절: `#def_object`, `#def_object_database`, `#def_object_name`, `#def_blob_object`)
- 확인일: 2026-10-17
- 확인 결과: **일부.** "content-addressable"이라는 단어와 중복 제거에 대한 직접 서술은 없다고 요약기가 보고했다.

확인한 내용: 객체는 저장 단위이고 내용의 SHA-1로 유일하게 식별된다. 따라서 바꿀 수 없다. 객체 이름은 40글자 16진수 문자열이다. blob은 파일 내용 같은 형식 없는 객체다.

인용: "It is uniquely identified by the SHA-1 of its contents." (`#def_object`)

**S13~S15 종합 — 우리 형식 매핑:**

| 구분 | 내용 | 표시 |
| --- | --- | --- |
| 확인된 사실 | 키가 내용에서만 파생된다 (S13, S15) | [확인] |
| 확인된 사실 | 변경 없는 파일은 다시 저장하지 않는다 (S14) | [확인] |
| 추론 | 「서로 다른 경로에 있는 동일한 내용도 같은 키라 한 번만 저장된다」 | [추론] 키가 내용에서만 나온다는 사실에서 따른다. 위 세 소스 어디에도 이 문장 그대로는 없다. |
| 빌릴 것 | 「같은 내용이면 같은 키」 원리를 **정적 place**(함수 심볼, 커널 함수)와 **note의 재사용 키**에 쓴다 | [추론] D9, D11 |
| 버릴 것 | 해시를 `event` 행의 기본키로 쓰는 것 | 같은 호출이 여러 번, 같은 내용으로 일어난다. 호출 한 번이 한 행이어야 한다. 키 정책은 [추론]. |
| 버릴 것 | SHA-1 자체 | 알고리즘 선택은 별개 문제다. SHA-1에서 다른 해시로의 전환 문서는 이번에 읽지 않았다. [못 봄] |

---

## 6. 선택 항목 — ftrace function_graph

### S16. Linux ftrace 문서

- 제목·URL: ftrace - Function Tracer — https://docs.kernel.org/trace/ftrace.html (절: `#function-graph-tracer`)
- 확인일: 2026-10-17
- 확인 결과: **일부.** 페이지는 155,141자이고 첫 호출은 `function` 트레이서 절 앞에서 잘렸다. 그래서 offset 60,000과 100,000으로 다시 읽어 `function graph tracer` 절을 찾았다. 앞쪽 일부는 읽지 않았을 수 있다.

확인한 내용:

- 출력 열: `CPU`, `DURATION`, `FUNCTION CALLS`.
- 호출을 담은 함수는 `{`로 열고 `}`로 닫는다. 호출이 없는 함수는 한 줄에 `name();`으로 끝난다. 샘플:

```
 0)               |  sys_open() {
 0)               |    do_sys_open() {
 0)               |      getname() {
 0)               |        kmem_cache_alloc() {
 0)   1.382 us    |          __might_sleep();
 0)   2.478 us    |        }
```

- 지속 시간은 닫는 괄호 줄에 표시된다.
- 느린 호출 표식: `+` 10 usecs 초과, `!` 100 usecs 초과, `#` 1000 usecs 초과, `*` 10 ms, `@` 100 ms, `$` 1초 초과.
- 깊이: 샘플에서 중첩 호출이 들여쓰기로 보인다. 요약기는 「들여쓰기 규칙을 원문이 명시하지 않는다」고 보고했다. [추론 — 샘플에서 읽음]
- `us`가 마이크로초라는 정의는 페이지에 명시되지 않았다. 표식 범례가 "usecs"라고 쓰므로 마이크로초로 읽는 것은 가벼운 [추론]이다.

인용: "displayed on the closing bracket line of a function"

우리 형식 매핑:

| 판정 | 후보 | 이유 |
| --- | --- | --- |
| 빌릴 것 | `{`/`}`는 진입·종료 한 쌍이고 `name();`은 진입과 종료가 한 줄인 리프다 | 한 호출을 한 `event` 행으로 합칠 때의 가장 작은 단위다. 파서는 괄호 짝을 맞춰 parent를 만든다. [추론] |
| 빌릴 것 | 깊이는 별도 필드가 아니라 중첩 상태에서 나온다 | ftrace 출력은 깊이를 열로 갖지 않고 들여쓰기로만 보인다. 우리가 `depth`를 저장한다면 파생 열임이 분명하다 (D3). |
| 빌릴 것 | 소수 마이크로초(예: 1.382) 지속 시간 | ns 정수로 올려 담을 수 있다 (D1). |
| 버릴 것 | `+`/`!` 같은 느린 호출 표식 | 표시용 가공이다. 임계값은 뷰에서 계산한다. |
| 버릴 것 | CPU 열을 장소로 쓰는 것 | 커널 안의 호출은 CPU가 아니라 task(프로세스·스레드)에 귀속해야 syscall과 이어진다. [추론 — 이번에 ftrace의 task 표시 열(`funcgraph-proc`)은 요약 수준으로만 봤다] |

---

## 7. note 표의 선례

이번에 가져온 어떤 소스에서도 「독자를 위한 설명이 정규화된 범위에 붙는」 모델을 찾지 못했다. OTel `Events`·`attributes`, Chrome `args`, Perfetto `args`는 모두 **기록자가 남긴 사실**이다. note는 **독자·AI가 사후에 붙이는 설명**이라서 성격이 다르다. 선례를 억지로 끌어오지 않는다. Perfetto UI의 메모 기능 같은 것이 있는지는 이번에 보지 않았다. [못 봄]

---

## 8. T1에 적용할 결정 후보

각 결정에서 「근거」는 [확인]한 사실이고 「판단」은 [추론]이다.

### D1. 시각의 단위와 기준

- 근거 [확인]: OTLP는 UNIX epoch 이후 나노초(`fixed64`)다 (S4). Perfetto는 나노초이고 기준 시계는 `primary_trace_clock`에 따른다 (S5, S7). Chrome은 마이크로초이고 예시에 소수가 있다 (S10). strace 가져오기는 epoch 초(`-ttt`)를 요구한다 (S9). ftrace 샘플은 소수 마이크로초를 보여 준다 (S16).
- 판단 [추론]: 저장은 **정수 나노초 `start_ns`·`end_ns`(run 시작 기준 상대값)** 로 하고, run에 `start_epoch_ns`와 시계 종류를 둔다. 이유 둘.
  1. 마이크로초 소스는 나노초로 무손실로 올라간다. 나노초 소스는 마이크로초로 내리면 손실이 생긴다.
  2. 2026년 epoch 나노초는 약 180경(1,800,000,000,000,000,000) 근처라서, 브라우저 JavaScript 숫자가 정수를 정확히 담는 한도(2^53 = 9,007,199,254,740,992, 약 9,007조)를 훌쩍 넘는다. run 시작 기준 상대 나노초는 약 104일(9,007,199초 ÷ 86,400초)까지 정확하다 (산술).
- 내보낼 때(Chrome JSON): `start_ns`를 1,000으로 나눈 값을 소수 마이크로초로 적는다. 이 값을 Perfetto가 소수 마이크로초로 읽는지는 확인하지 않았다. [못 봄]

### D2. span의 저장 단위: 한 행, 끝은 NULL 가능

- 근거 [확인]: OTLP는 시작·종료를 한 레코드에 둔다 (S4). Chrome의 `X`는 `B/E` 쌍을 합친 것이고 Perfetto는 `B/E`와 `X`를 같은 슬라이스로 가져온다 (S9, S10). 끝나지 않은 호출은 Perfetto strace 가져오기에서 `dur = -1`이다 (S9). ftrace는 `{`와 `}`가 짝이다 (S16).
- 판단 [추론]: 적재 때 enter/exit 쌍을 한 행으로 합친다. 필드는 `start_ns`, `end_ns`. **끝을 알 수 없으면 `end_ns` NULL**로 한다. Perfetto의 `-1` 센티널은 `dur = 0`(점)과 의미가 한 열에 섞이므로 따르지 않는다.
- 열린 span이 생기는 이유 후보: `-finstrument-functions` 기록에서 진입 훅만 남고 종료 훅이 빠지는 경우. PostgreSQL의 오류 경로가 `longjmp`를 쓰는지는 이번에 읽지 않았다. [못 봄 — 실험으로 확인할 것]

### D3. parent와 depth

- 근거 [확인]: OTLP는 `parent_span_id`만 저장하고 깊이는 없다 (S4). Perfetto는 `parent_id`와 `depth`를 모두 저장한다 (S6). ftrace는 깊이를 열이 아니라 들여쓰기로 보인다 (S16). Perfetto의 조상·후손 계산은 같은 트랙 안에서만 이뤄진다 (S5).
- 판단 [추론]: `parent_event_id`가 진실이고 `depth`는 적재 때 계산해 넣는 비정규화 열이다. 「깊이 ≤ N만 보기」 같은 줌 수준 필터가 한 번의 범위 조회로 끝난다. 중첩 범위는 「같은 run의 같은 실행 맥락 place」로 문서에 못 박는다.

### D4. span과 instant

- 근거 [확인]: Chrome은 `i`(범위 `s`)와 `B/E/X`를 phase로 구분한다 (S10). Perfetto의 strace 가져오기는 `-T`가 없으면 진입 시각만 있는 호출을 길이 0 표식으로 둔다 (S9). OTel `Events`는 Span에 매달려 있고 독립 이벤트가 아니다 (S1, S4).
- 판단 [추론]: `event` 표 하나에 `kind` 열(`span` / `instant` / `sample`)을 둔다. instant는 `end_ns = start_ns`가 아니라 **NULL**을 쓰고 kind로 구분한다. 그래야 「끝을 못 본 span」(`kind = span`, `end_ns` NULL)과 「애초에 점」(`kind = instant`)이 섞이지 않는다. instant의 감싸는 span은 `parent_event_id`로 가리킬 수 있다. instant의 범위(`g/p/t`)는 열을 따로 두지 않고 `place_id`의 층위로 나타낸다.

### D5. link: 쌍, 종류, 근거

- 근거 [확인]: Perfetto `flow`는 (`slice_out`, `slice_in`, `trace_id`, `arg_set_id`) 쌍이다 (S6). Chrome flow는 `s/t/f` 세 phase이고 끝점 바인딩이 `bp`에 따라 달라진다 (S10). OTel Link는 같은 trace이든 다른 trace이든 가리킬 수 있고, 부모-자식으로 표현할 수 없는 인과와 fan-in에 쓴다 (S1, S2).
- 판단 [추론]: `link(run_id, src_event_id, dst_event_id, kind, basis, attrs)`.
  - `kind`는 어휘를 닫는다. 예: `message`(송신→수신), `syscall_entry`(사용자 영역 syscall→커널 호출 트리 루트). OTel `SpanKind`의 `PRODUCER`/`CONSUMER`는 message의 방향 표식 후보로만 참고한다.
  - 중간 단계(`t`)는 두지 않는다. 사슬은 연속된 쌍 여러 행으로 쓴다.
  - 양끝은 적재 시점에 event id로 확정한다 (`bp`의 모호함을 피함).
  - **`basis`** 는 링크를 어떻게 확정했는지(`propagated_id` / `matched_by_bytes` / `matched_by_order` / `direct_entry` 등)를 적는 열이다. OTel은 ID를 전파해 링크를 확정하지만 우리는 대개 사후 짝짓기다 (S3). 확신도를 질의로 거를 수 있게 한다.
  - fan-out은 같은 `src`로 여러 행을 둔다. OTel처럼 span당 링크 개수에 제한을 두지 않는다 (S2).

### D6. attributes: 확장 열 하나, 구조 열은 최소

- 근거 [확인]: OTLP `attributes`는 유일한 키-값 목록이고 `dropped_*_count`를 따로 둔다 (S4). Chrome `args`는 자유 객체다 (S10). Perfetto는 `arg_set_id`로 별도 표를 가리킨다 (S6, S9).
- 판단 [추론]: 구조 열은 `id`, `run_id`, `place_id`, `kind`, `name`, `start_ns`, `end_ns`, `parent_event_id`, `depth`, `source`로 둔다. 소스별 내용(syscall 인자, 와이어 메시지 필드, JFR 프레임 정보)은 `attrs`(키-값) 한 열에 둔다. `dropped_*_count`는 버리되, 수집기가 값을 잘라 담았으면 `attrs`에 `truncated`를 남긴다. `category`/`cat` 대신 값이 하나인 `source`(어느 수집기에서 왔는가)를 둔다.

### D7. place: 대리키, 자기 참조 계층, OS id는 속성

- 근거 [확인]: Perfetto는 OS id가 trace 안에서 유일한 식별자가 못 된다고 하고 `utid`/`upid`를 쓴다 (S5, S6). `process.parent_upid`와 `thread.upid`로 계층을 외래키로 표현하고 `start_ts`/`end_ts`로 생존 구간을 둔다 (S6). Chrome은 `M` 이벤트로 이름을 붙인다 (S10). OTLP는 `ResourceSpans → ScopeSpans` 고정 2단계다 (S4).
- 판단 [추론]: `place(id, parent_place_id, kind, name, attrs)` 하나로 system→host/container→process→component→function과 host→kernel→kernel function을 인접 목록으로 표현한다. pid·tid·컨테이너 id는 `attrs`에 두고 키로 쓰지 않는다. 생존 구간은 필요할 때만 둔다.

### D8. 정적 place와 동적 place를 구분할지

- 근거 [확인]: Perfetto는 「어디서 실행됐는가」(트랙·스레드)와 「무엇이었는가」(슬라이스 `name`)를 분리한다 (S5, S6). git은 내용에서 키를 만들면 같은 내용이 같은 키가 된다 (S13~S15).
- 판단 [추론, 열린 질문]: 함수·커널 함수는 한 호출마다 새로 생기는 것이 아니라 심볼로 이미 정해진 장소다. 반면 프로세스·스레드는 run마다 다르다. 두 부류의 정체성 정책이 다를 수 있다. 정적: (빌드 id, 심볼) 같은 자연키로 run을 넘어 공유. 동적: run 소속. 한 표에 `scope`(run/global) 열을 두는 방법이 있다. T1에서 정할 일이고 이 조사에서는 결론을 내지 않는다.

### D9. run

- 근거 [확인]: Perfetto 표에는 run에 해당하는 표가 없고 trace 하나가 처리 단위다 (S5~S8). OTel `trace_id`는 요청 단위 상관 ID다 (S1). Chrome `otherData`는 trace 수준 메타데이터다 (S10).
- 판단 [추론]: `run`은 우리가 새로 두는 표다. 기록 한 번(시작 epoch, 시계 종류, 수집기 목록, 환경 설명)을 담고 `event`·`link`가 `run_id`를 가진다. OTel `trace_id`는 run 키가 아니라 `attrs`로만 다룬다.

### D10. 관측(`pg_stat_activity` 샘플)

- 근거 [확인]: Perfetto `counter`는 「트랙에 놓인 값」이다 (S7). Chrome에 `C`(카운터) phase가 있고 Perfetto가 카운터 트랙으로 가져온다 (S9, S10).
- 판단 [추론, 열린 질문]: 후보 둘. (a) `kind = sample`인 `event`에 값은 `attrs`. 표가 늘지 않는다. (b) 수치 계열은 카운터 같은 별도 표. `pg_stat_activity`는 행이 문자열·상태 등 다양해서 (a)가 맞아 보이지만, 수치 시계열(대기 이벤트 수 등)이 많아지면 (b)가 낫다. T1 시점에서 실제 샘플을 보고 정한다.

### D11. note

- 근거 [못 봄]: 가져온 소스에 선례가 없다 (7절).
- 판단 [추론]: 정규화된 범위(예: 동일한 호출 경로)에 붙는 설명의 재사용 키는 「내용에서 파생한 키」(S13~S15의 원리)로 만들 수 있다. 같은 경로면 같은 키라 설명이 재사용된다. 키에 무엇을 넣을지(함수 이름 경로만 / 인자 모양 포함)는 T1에서 따로 정한다.

### D12. 집계 뷰는 파생물

- 근거 [확인]: 플레임 그래프의 x축은 알파벳 순으로 시간이 아니고 플레임 차트가 시간순이다 (S11). 접힌 스택은 루트→리프 + 횟수이고 `--flamechart`는 병합하지 않는다 (S9, S12).
- 판단 [추론]: 저장은 시간순 개별 호출(`event`)만 한다. 접힌 스택·플레임 그래프는 질의 결과나 내보내기로만 만든다.

### D13. ID와 외부 검증 경로

- 근거 [확인]: OTLP ID는 16바이트·8바이트 바이트열이다 (S1, S4). Perfetto는 정수 `id`를 쓴다 (S6). Chrome JSON은 Perfetto UI에서 바로 열리고, 중첩되지 않는 기간 이벤트는 overflow 트랙에 놓인다 (S9).
- 판단 [추론]: 단일 적재기를 가진 오프라인 저장소이므로 정수 대리키로 충분하다. 개발 중에는 적재 결과를 Chrome JSON으로 내보내 Perfetto UI에서 열면, enter/exit 짝이 틀린 곳이 overflow 트랙으로 눈에 띈다. 이 경로는 T1 이후에 시험해야 한다.

---

## 9. 「다시 볼 조건」

| 조건 | 다시 볼 것 |
| --- | --- |
| 소스끼리 충돌한 지점 (Chrome instant의 phase 글자) | S9(Perfetto `#chrome-json-format`)는 instant를 대문자 `I`로 적고, S10(Google 문서 Phases 표와 예시)은 소문자 `i`를 쓰며 `I`는 Deprecated로 표시한다. 둘 다 요약기 경유다. 소문자 `i`를 쓰는 쪽이 명세 원문이고 Perfetto 문서의 대문자 표기는 요약기 오기이거나 Perfetto 문서의 표기일 수 있다. 원문을 직접 열어 판정할 것 |
| 이 문서의 인용·앵커를 다른 문서에 옮겨 쓸 때 | 모든 인용과 앵커는 요약기 경유다. 원문을 직접 열어 대조한다. 특히 **(앵커 추정)** 표시가 붙은 곳과, S10의 「네 가지 범위인데 셋만 나열」 같은 요약기 보고 |
| `end_ns` NULL 대신 센티널이 필요해질 때 | Perfetto `dur = -1`의 정의를 slice 표 쪽에서 찾을 것 (이번에는 strace 절에서만 확인) |
| Perfetto prelude를 더 깊게 인용할 때 | `stdlib-docs`는 549k자 중 100k자만 읽었다. 나머지와 `thread_track`·`track` 표의 앵커를 확인 |
| 와이어 프로토콜 link를 설계할 때 | PostgreSQL 와이어 프로토콜 문서(trace 컨텍스트를 실을 자리가 있는지), 그리고 `longjmp` 경로에서 종료 훅이 빠지는지 (D2). 둘 다 이번에 읽지 않았다 |
| 다중 머신이나 다중 시계를 병합할 때 | Perfetto의 시계 동기화 문서(이번에 못 봄)와 OTel `Resource` 의미 규약 |
| ftrace 입력을 실제 파서로 만들 때 | `function_graph` 출력 전체 문법(`funcgraph-*` 옵션을 켰을 때 달라지는 열 배치, 컨텍스트 전환 때 나오는 프로세스 이름 줄). 이번에는 대표 샘플과 표식만 확인 |
| 해시 키를 쓰기로 할 때 | git의 해시 함수 전환 문서(SHA-1 → SHA-256). 이번에 못 봄 |
| Chrome JSON 내보내기를 실제로 쓸 때 | flow `id` 연결 규칙과 소수 마이크로초 `ts` 수용 여부 (S10에서 요약기가 못 찾았거나 예시만 확인) |
| 이벤트 수가 run당 수백만 단위가 될 때 | `depth` 같은 비정규화 열의 비용, 질의 성능, 저장소 선택 (이번 조사 범위 밖) |
