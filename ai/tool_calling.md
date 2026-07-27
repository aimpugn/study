# 툴 콜링은 실제로 어떻게 이뤄지는가

## 짧은 답

**모델은 도구를 실행하지 않습니다.** 모델은 어떤 도구를 어떤 인자로 써야 하는지 표현한 출력을 만들 뿐입니다. 사용자가 정의하고 자기 프로그램에서 실행하는 도구(client tool)는 호출한 애플리케이션이 실행하고, Anthropic 서버가 실행하는 도구(server tool)는 Anthropic이 실행합니다. 실행 주체는 달라도 모델이 직접 함수를 실행하지 않는다는 경계는 같습니다.

"에이전트가 알아서 도구를 호출해준다"는 설명이 부족한 이유가 여기 있습니다. 그 문장에는 **누가 실행하는가**가 빠져 있고, 그래서 도구가 실패하면 무엇을 돌려줘야 하는지, 도구 설명문이 왜 중요한지, 왜 매 요청마다 도구 정의가 다시 실려 가는지를 설명할 수 없습니다.

## 왜 이런 구조가 생겼는가

언어 모델은 입력을 읽고 다음 토큰을 만드는 계산 장치입니다. 이 계산만으로는 현재 날씨나 사내 데이터베이스를 조회할 수 없고, 결제 승인이나 파일 쓰기처럼 외부 상태를 바꾸는 일도 수행할 수 없습니다. 그래서 모델에게 실행 권한을 직접 주는 대신, 모델은 **실행 요청을 구조화해서 만들고 신뢰할 수 있는 프로그램이 그 요청을 검사·실행한 뒤 결과를 돌려주는 경계**가 필요해졌습니다.

Anthropic은 2024년 5월 30일 Claude의 tool use를 정식 기능으로 발표했습니다. 이 API는 모델의 선택과 외부 프로그램의 실행을 `tool_use`와 `tool_result`라는 블록으로 연결했습니다. 이 문서는 그중 사용자가 정의한 **client tool의 한 차례 왕복을 주 경로**로 설명하고, 실행 주체가 다른 server tool은 경계가 달라지는 지점에서 따로 표시합니다.

## 이 문서가 답하는 것

- 도구 정의는 어떤 모습으로 모델에게 전달되는가
- 모델이 도구를 쓰겠다고 결정하면 응답이 어떤 상태로 멈추는가
- 실행 결과는 어떤 경로로 대화에 다시 들어가는가
- 도구가 실패하면 무엇을 돌려줘야 하는가
- 여러 도구를 한 번에 호출하면 결과를 어떻게 돌려줘야 하는가
- 도구 설명문이 왜 프롬프트 엔지니어링인가
- 이 모든 것이 프롬프트 캐시와 어디서 만나는가

## 목차

- [1. 모델이 실제로 받는 것은 JSON이 아니다](#1-모델이-실제로-받는-것은-json이-아니다)
- [2. Qwen에서는 도구 정의가 시스템 메시지 안에 들어간다](#2-qwen에서는-도구-정의가-시스템-메시지-안에-들어간다)
- [3. 왕복 한 바퀴 전체](#3-왕복-한-바퀴-전체)
- [4. 모델이 멈추는 지점](#4-모델이-멈추는-지점)
- [5. 결과가 돌아가는 길](#5-결과가-돌아가는-길)
- [6. 도구 설명문이 프롬프트인 이유](#6-도구-설명문이-프롬프트인-이유)
- [7. 도구가 실패하면](#7-도구가-실패하면)
- [8. 여러 client tool을 한 번에 호출하면](#8-여러-client-tool을-한-번에-호출하면)
- [9. 프롬프트 캐시와 만나는 지점](#9-프롬프트-캐시와-만나는-지점)
- [10. 흔한 오해](#10-흔한-오해)
- [부록 A. 재현 방법](#부록-a-재현-방법)
- [부록 B. 확인 시점과 출처](#부록-b-확인-시점과-출처)
- [주요 근거](#주요-근거)
- [스스로 확인할 질문](#스스로-확인할-질문)

---

## 1. 모델이 실제로 받는 것은 JSON이 아니다

우리가 API로 보내는 건 이렇게 생긴 구조화된 데이터입니다.

```json
{
  "messages": [
    { "role": "user", "content": "서울 날씨 알려줘" }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "특정 도시의 현재 날씨를 조회한다.",
        "parameters": {
          "type": "object",
          "properties": {
            "city": { "type": "string", "description": "도시 이름. 예: 서울" }
          },
          "required": ["city"]
        }
      }
    }
  ]
}
```

`role`이라는 필드가 있고, `tools`라는 별도 배열이 있습니다. 마치 도구가 대화와는 다른 통로로 전달되는 것처럼 보입니다. 이 예제는 Qwen이 받는 OpenAI 호환 형식입니다. 뒤에서 다루는 Anthropic client tool은 `name`, `description`, `input_schema`를 `tools` 배열의 최상위 필드로 둡니다.

Qwen 같은 텍스트 기반 열린 모델은 이 JSON 객체 자체를 읽지 않습니다. **채팅 템플릿**이 role, 도구 정의, 메시지를 하나의 문자열로 직렬화하고, 토크나이저가 그 문자열을 토큰 번호 배열로 바꿉니다. 열린 가중치 모델에서는 보통 `tokenizer_config.json`의 `chat_template` 필드에서 이 변환을 확인할 수 있습니다.

이 관측을 모든 API의 내부 표현으로 일반화하면 안 됩니다. Anthropic은 `tools`로부터 별도 system prompt를 구성한다는 사실과 그 개략적인 틀을 공개하지만, 실제 요청 하나의 내부 직렬화 전문은 제공하지 않습니다. 따라서 공식 레퍼런스로 요청·응답 블록의 계약, system prompt 구성 원리, 캐시 접두사 계층은 확인할 수 있어도 Qwen과 같은 실제 문자열을 복원했다고 말할 수는 없습니다.

Qwen2.5-7B-Instruct의 템플릿에 위 JSON을 넣고 돌리면 아래 757자가 나옵니다. 토크나이저는 이 문자열을 188개 토큰 번호로 바꿔 모델에 전달합니다.

```text
<|im_start|>system
You are Qwen, created by Alibaba Cloud. You are a helpful assistant.

# Tools

You may call one or more functions to assist with the user query.

You are provided with function signatures within <tools></tools> XML tags:
<tools>
{"type": "function", "function": {"name": "get_weather", "description": "특정 도시의 현재 날씨를 조회한다.", "parameters": {"type": "object", "properties": {"city": {"type": "string", "description": "도시 이름. 예: 서울"}}, "required": ["city"]}}}
</tools>

For each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:
<tool_call>
{"name": <function-name>, "arguments": <args-json-object>}
</tool_call><|im_end|>
<|im_start|>user
서울 날씨 알려줘<|im_end|>
<|im_start|>assistant
```

이 Qwen 템플릿에서는 도구 정의가 별도 통로로 남지 않고 **시스템 턴 안의 JSON 텍스트**로 들어갑니다. 이것은 열린 모델에서 직접 확인한 한 구현이지, 모든 상용 API의 내부 직렬화가 같다는 증거는 아닙니다.

---

## 2. Qwen에서는 도구 정의가 시스템 메시지 안에 들어간다

위 출력을 네 조각으로 나눠서 각각이 무슨 일을 하는지 보겠습니다.

**첫째, 역할 구분자.** `<|im_start|>` 와 `<|im_end|>` 는 특수 토큰입니다. 토크나이저가 이 문자열을 하나의 토큰 번호로 바꿉니다. 이 두 표시 덕분에 모델은 "여기부터 시스템, 여기부터 사용자"를 구분합니다. `role: "system"`이라는 JSON 필드가 텍스트 세계에서 어떤 모습이 되는지가 여기 있습니다 — 필드가 아니라 **구분자 토큰**입니다.

**둘째, 도구가 있다는 안내문.**

```text
# Tools

You may call one or more functions to assist with the user query.

You are provided with function signatures within <tools></tools> XML tags:
```

평범한 영어 문장입니다. Qwen이 배포한 템플릿은 이 안내문 뒤에 도구 목록을 배치합니다. 따라서 모델이 받는 입력에서는 이 문장과 도구 정의가 함께 도구 사용 조건을 설명하는 신호로 작동합니다. 이 문서의 렌더링 실험은 그 입력 구조까지만 확인하며, 실제 모델의 도구 선택률은 측정하지 않습니다.

**셋째, 도구 정의 그 자체.** `<tools>` 와 `</tools>` 사이에 우리가 보낸 JSON이 `tojson` 필터를 거쳐 한 줄 문자열로 직렬화돼 들어갑니다. `description` 필드에 우리가 쓴 한국어 문장 "특정 도시의 현재 날씨를 조회한다."가 그대로 보입니다. **도구 설명문은 프롬프트 본문입니다.** 6절에서 이 사실의 결과를 다룹니다.

**넷째, 출력 형식 지시.**

```text
For each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:
<tool_call>
{"name": <function-name>, "arguments": <args-json-object>}
</tool_call>
```

여기가 결정적입니다. 이건 **명령이지 강제가 아닙니다.** 모델은 "이런 모양으로 써 달라"는 부탁을 받았을 뿐이고, 실제로 그 모양을 지킬지는 학습된 확률 분포에 달려 있습니다. 형식을 진짜로 강제하려면 샘플링 단계에서 문법에 맞지 않는 토큰을 아예 못 고르게 막아야 하는데, 그건 이 문서의 범위 밖입니다.

**마지막 줄에 주목하십시오.** 출력이 `<|im_start|>assistant\n` 로 끝납니다. 이건 모델의 응답이 아니라 **모델이 이어서 쓰도록 미리 깔아 둔 발판**입니다. 모델이 하는 일은 이 문자열 뒤에 올 다음 토큰을 하나씩 예측하는 것뿐입니다. 그러니 "도구를 호출한다"는 건 정확히 말해 **`<tool_call>` 로 시작하는 문자열을 이어 쓴다**는 뜻입니다.

---

## 3. 왕복 한 바퀴 전체

이제 Qwen에서 관측할 수 있는 client tool 왕복으로 전체 흐름을 봅니다. 세로 점선이 모델과 바깥 세계의 경계입니다.

```text
  내 코드 / 하네스              │ 모델
                               │
 ① 요청 JSON 조립               │
    messages + tools           │
         │                     │
 ② 채팅 템플릿으로 렌더링        │
    JSON ──> 문자열 하나        │
         │                     │
         └───── 토큰 배열 ─────>│ ③ 이 텍스트를 처음부터 읽는다
                               │        │
                               │        ▼
         <──── 생성 텍스트 ─────┤ ④ 다음 토큰을 하나씩 이어 쓴다
         │                     │    <tool_call>{"name": ...}</tool_call>
 ⑤ 파싱                        │
    문자열 ──> 구조화 블록       │
         │                     │
 ⑥ 응답 JSON을 받는다            │
    stop_reason = tool_use     │
         │                     │
 ⑦ 내가 함수를 실행한다          │   ← 모델은 여기에 관여하지 않는다
         │                     │
 ⑧ 결과를 대화 끝에 붙여서       │
    ①로 돌아간다                │
```

두 가지를 확인하십시오.

**사용자가 정의한 client tool에서 모델이 하는 일은 ③→④뿐입니다.** 텍스트를 읽고 도구 요청을 표현한 출력을 만듭니다. ②와 ⑤는 하네스 또는 API 서버가 담당하고, ⑦은 호출한 애플리케이션의 코드가 담당합니다.

Qwen 예제에서 **②와 ⑤는 서로 반대 방향의 변환**입니다. `tools` 배열은 템플릿 안의 JSON 텍스트가 되고, 모델이 생성한 `<tool_call>` 텍스트는 다시 구조화된 호출로 파싱됩니다. Anthropic API도 요청·응답 양쪽에 구조화된 블록을 제공하지만 내부 직렬화 형식은 공개하지 않으므로, Qwen과 같은 평문 템플릿이라고 단정하지 않습니다.

Anthropic의 server tool은 ⑦과 ⑧의 소유자가 다릅니다. 웹 검색, 웹 가져오기, 코드 실행 같은 server tool은 Anthropic 서버가 실행하고 결과를 모델에게 다시 공급합니다. 애플리케이션은 보통 그 도구의 `tool_result`를 직접 만들지 않습니다. 따라서 이 절의 ASCII 흐름은 **client tool의 수동 왕복**이며, 모든 tool use의 실행 위치를 뜻하지 않습니다.

한 가지 중요한 성질이 여기서 나옵니다. **Messages API의 다음 요청은 앞선 요청에서 생략한 대화를 자동으로 복원하지 않습니다.** ⑧에서 다시 ①로 갈 때 이전 대화와 도구 결과를 `messages`에 다시 넣어 보내야 합니다. 도구 실행 결과를 빼면 모델이 그 결과를 다음 응답의 입력으로 사용할 수 없습니다.

---

## 4. 모델이 멈추는 지점

모델이 도구를 쓰겠다고 결정하면 응답이 끝나면서 **왜 끝났는지를 알려주는 값**이 같이 옵니다. Anthropic API에서는 `stop_reason`이라는 필드입니다.

| `stop_reason` | 의미 | 내가 할 일 |
|---|---|---|
| `end_turn` | 할 말을 다 하고 정상 종료 | 응답을 사용자에게 보여준다 |
| `tool_use` | 도구를 쓰겠다고 멈춤 | 실행하고 결과를 붙여 다시 호출 |
| `max_tokens` | 출력 길이 상한에 걸림 | 상한을 올리거나 이어서 생성 |
| `model_context_window_exceeded` | 모델의 context window를 모두 사용 | 잘린 응답으로 처리하거나 입력을 줄임 |
| `stop_sequence` | 지정한 중단 문자열을 만남 | 상황에 따라 |
| `pause_turn` | 서버 측 도구가 반복 상한에 걸려 일시 정지 | 그대로 다시 보내면 이어서 진행 |
| `refusal` | 안전상의 이유로 거절 | 내용을 읽기 전에 먼저 이 값을 확인 |

`model_context_window_exceeded`는 Sonnet 4.5 이상에서 일반 응답으로 제공됩니다. 이전 모델이나 SDK 버전에서는 beta header와 beta namespace가 필요할 수 있으므로 사용하는 모델과 SDK의 레퍼런스를 함께 확인해야 합니다.

`tool_use`로 멈췄을 때 응답 본문에는 `tool_use` 블록이 들어 있습니다.

```json
{
  "stop_reason": "tool_use",
  "content": [
    { "type": "text", "text": "날씨를 확인해 보겠습니다." },
    {
      "type": "tool_use",
      "id": "toolu_abc123",
      "name": "get_weather",
      "input": { "city": "서울" }
    }
  ]
}
```

`id`를 기억해 두십시오. 결과를 돌려줄 때 이 값으로 짝을 맞춥니다.

`content`가 배열이고 `text` 블록과 `tool_use` 블록이 같이 들어 있다는 점도 봐 둘 만합니다. 따라서 `stop_reason == "tool_use"`라고 해서 응답의 모든 자연어 블록을 버리면 안 됩니다. Anthropic이 이 두 블록을 내부에서 어떤 문자열로 표현하는지는 공개하지 않으므로, 응답 배열만 보고 내부 파싱 과정을 단정하지 않습니다.

---

## 5. 결과가 돌아가는 길

도구를 실행했으면 결과를 대화에 붙여서 다시 보냅니다. Anthropic API에서는 `tool_result` 블록을 **사용자 메시지 안에** 넣습니다.

```json
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_abc123",
      "content": "{\"city\": \"서울\", \"temp_c\": 29, \"condition\": \"흐림\"}"
    }
  ]
}
```

`tool_use_id`가 4절의 `id`와 같아야 합니다. 짝이 맞지 않으면 API가 거절합니다.

이 user 메시지는 assistant의 `tool_use` 메시지 바로 뒤에 와야 합니다. `content`에 일반 text도 함께 넣어야 한다면 모든 `tool_result`를 먼저 두고 text를 그 뒤에 둡니다. assistant가 아직 끝나지 않은 server tool과 client tool을 함께 요청한 경우에는 client `tool_result`만 보내고 일반 text를 섞지 않아야 합니다.

여기서 눈여겨볼 게 있습니다. **역할이 `user`입니다.** `tool`이라는 별도 역할이 아닙니다. 왜 그런지는 Qwen 템플릿을 다시 렌더링해 보면 눈으로 확인됩니다. 도구 호출과 실행 결과까지 붙여서 2차 요청을 만들면 이렇게 됩니다 (앞부분 시스템 메시지는 동일하므로 생략).

```text
<|im_start|>user
서울 날씨 알려줘<|im_end|>
<|im_start|>assistant
<tool_call>
{"name": "get_weather", "arguments": {"city": "서울"}}
</tool_call><|im_end|>
<|im_start|>user
<tool_response>
{"city": "서울", "temp_c": 29, "condition": "흐림"}
</tool_response><|im_end|>
<|im_start|>assistant
```

도구 결과는 `<|im_start|>user` 아래에 놓이지만 `<tool_response>` 태그로 표시됩니다. Anthropic도 `role: "user"` 메시지 안에 결과를 넣되 `type: "tool_result"`인 콘텐츠 블록으로 일반 사용자 `text`와 구분합니다. **같은 user 메시지 봉투를 쓰는 것과 사용자 발화로 위장하는 것은 다릅니다.**

다만 `tool_result` 안의 웹 페이지나 데이터베이스 값은 다음 추론 입력에 포함되는 외부 데이터입니다. 구조화된 블록으로 구분되더라도 내용을 신뢰할 수 있다는 뜻은 아니므로, Anthropic도 통제할 수 없는 텍스트를 일반 user text나 system prompt가 아니라 `tool_result` 안에 유지하라고 권고합니다. 프롬프트 인젝션의 상세 방어는 이 문서의 범위를 넘으므로 여기서는 **타입 경계와 신뢰 경계가 다르다**는 점까지만 고정합니다.

---

## 6. 도구 설명문이 프롬프트인 이유

2절에서 확인한 대로 Qwen에서는 `description`이 프롬프트 본문에 그대로 찍힙니다. Anthropic도 도구 설명이 무엇을 하고 언제 써야 하는지를 자세히 적을수록 Claude가 도구를 선택하고 인자를 구성하는 데 도움이 된다고 안내합니다. 이 문서에서 직접 확인한 것은 **설명문이 모델 입력에 포함된다는 사실**이며, 설명 방식별 선택률은 측정하지 않았습니다.

결과적으로 아래 두 정의는 모델에게 **완전히 다른 프롬프트**입니다.

```json
{ "name": "search", "description": "검색한다." }
```

```json
{ "name": "search_internal_docs",
  "description": "사내 위키와 설계 문서를 전문 검색한다. 사용자가 우리 팀의 결정 사항, 운영 절차, 과거 장애 기록을 물을 때 사용한다. 공개된 일반 지식에는 쓰지 않는다." }
```

두 번째 정의는 무엇을 검색하고 언제 쓰지 말아야 하는지까지 판단 기준을 제공합니다. 그래서 모델이 `search`라는 짧은 이름만 보고 용도를 추측해야 하는 첫 번째 정의보다 선택 근거가 명확합니다. 실제 개선 폭은 모델과 도구 집합에 따라 달라지므로 대표 질의로 선택 정확도를 따로 측정해야 합니다.

"언제 쓰는가"를 적으면 모델이 판단해야 할 조건이 입력에 명시됩니다. 예를 들어 "사용자가 최근 사건이나 현재 가격을 물으면 호출하라"는 문장은 도구의 기능 설명만 있을 때보다 선택 기준을 구체적으로 제공합니다. 추론이 사라지는 것은 아니며, 모델은 여전히 사용자 요청과 그 조건을 비교해야 합니다.

반대 방향의 함정도 있습니다. `CRITICAL: 반드시 이 도구를 써야 한다` 같은 강한 표현은 실제 발동 조건보다 넓게 해석되어 불필요한 호출을 늘릴 수 있습니다. 설명문은 프롬프트의 일부이므로, 대표적인 호출·비호출 질의를 함께 두고 선택 결과를 확인해야 합니다.

---

## 7. 도구가 실패하면

client tool 실행이 실패해도 **결과 블록은 반드시 돌려줘야 합니다.** 응답을 생략하면 안 됩니다. client `tool_use` 블록 하나에 `tool_result` 하나가 짝을 이루지 않으면 API가 다음 요청을 거절합니다.

Anthropic API에서는 `is_error` 플래그를 씁니다.

```json
{
  "type": "tool_result",
  "tool_use_id": "toolu_abc123",
  "content": "Error: 'xyz'라는 도시를 찾을 수 없습니다. 올바른 도시 이름을 지정하세요.",
  "is_error": true
}
```

오류 메시지를 성의 있게 쓰는 게 실질적으로 중요합니다. 5절에서 봤듯이 이 문자열은 다음 요청의 프롬프트 안으로 들어가고, 모델은 이걸 읽고 다음 행동을 정합니다. `Error: 500`만 돌려주면 모델이 할 수 있는 게 없습니다. 무엇이 잘못됐고 어떻게 고치면 되는지를 적으면 모델이 인자를 고쳐서 다시 호출하거나 사용자에게 되물을 수 있습니다.

**오류 메시지도 프롬프트입니다.** 6절과 같은 원리입니다.

---

## 8. 여러 client tool을 한 번에 호출하면

모델은 한 응답에 client `tool_use` 블록을 여러 개 담을 수 있습니다. Anthropic API에서 이 동작은 기본으로 켜져 있습니다.

이때 다음 요청의 형식을 지켜야 합니다.

> **모든 client `tool_result` 블록을 바로 다음의 한 user 메시지에 함께 담고, 일반 text가 있다면 result 블록을 먼저 놓아야 합니다.**

일부 `tool_use_id`의 결과를 빠뜨리면 요청이 거절될 수 있습니다. 결과를 여러 user 메시지로 나누는 형식은 공식 예제에서 병렬 호출을 줄이는 잘못된 형식으로 분류됩니다. Anthropic은 모든 결과를 한꺼번에 돌려주는 형식이 이후에도 병렬 호출을 유지하는 데 중요하다고 안내합니다. 이는 가중치를 다시 학습시키는 규칙이 아니라, **API 계약과 현재 대화 문맥을 함께 보존하는 규칙**입니다.

```json
{
  "role": "user",
  "content": [
    { "type": "tool_result", "tool_use_id": "toolu_001", "content": "..." },
    { "type": "tool_result", "tool_use_id": "toolu_002", "content": "..." },
    { "type": "tool_result", "tool_use_id": "toolu_003", "content": "..." }
  ]
}
```

호출된 client tool 가운데 하나만 실패했더라도 그 블록을 빼면 안 됩니다. `is_error: true`를 붙여서 같이 담습니다.

병렬 호출을 끄고 싶으면 `tool_choice`에 `disable_parallel_tool_use: true`를 넣습니다. `tool_choice`는 네 가지 값을 가집니다.

| 값 | 동작 |
|---|---|
| `{"type": "auto"}` | 모델이 알아서 판단 (기본값) |
| `{"type": "any"}` | 반드시 도구 중 하나를 쓴다 |
| `{"type": "tool", "name": "..."}` | 지정한 도구를 반드시 쓴다 |
| `{"type": "none"}` | 도구를 쓸 수 없다 |

---

## 9. 프롬프트 캐시와 만나는 지점

툴 콜링과 프롬프트 캐시는 별개 주제처럼 보이지만 한 지점에서 정면으로 만납니다. 이 절은 그 지점만 다룹니다. 캐시의 원리와 층별 논증은 [프롬프트 캐시 문서](prompt_caching.md#4-왜-접두사여야만-하는가)에서 이어집니다.

### 도구 정의는 프롬프트 맨 앞에 있다

Qwen 템플릿에서는 도구 정의가 시스템 턴 안의 앞부분에 들어갔습니다. Anthropic은 내부 렌더링 전문 대신 캐시가 접두사를 구성하는 계층을 공개합니다. 그 순서는 다음과 같습니다.

```text
[ tools ] ──> [ system ] ──> [ messages ]
   맨 앞                          맨 뒤
```

프롬프트 캐시는 **접두사가 완전히 같을 때만** 재사용됩니다. 정확히는 캐시 지점까지의 토큰화된 입력과 캐시를 만드는 설정이 같아야 합니다. 도구 정의가 캐시 계층의 맨 앞에 있다는 사실과 이 규칙을 합치면 결론이 바로 나옵니다.

**도구 정의를 건드리면 캐시 전체가 날아갑니다.**

### 직접 측정한 결과

Qwen 템플릿으로 세 가지를 측정했습니다.

| 측정 | 결과 |
|---|---|
| 1차 요청(757자)이 2차 요청(975자)의 접두사인가 | **참** |
| 도구 설명문의 마지막 마침표를 느낌표로 교체 | 공통 접두사 757자 → **339자** |
| 시스템 프롬프트의 시각을 1초 변경 | 공통 접두사 **56자**, 57번째 문자부터 달라짐 |

첫 줄이 좋은 소식입니다. **client tool 왕복은 앞선 대화를 수정하지 않고 결과를 끝에 붙인다면 캐시 친화적입니다.** 이 Qwen revision에서는 1차 요청의 757자뿐 아니라 188개 토큰도 2차 요청의 접두사로 그대로 남았습니다. 상용 API에서 실제로 계산을 건너뛰었는지는 `cache_read_input_tokens` 같은 usage 값으로 따로 확인해야 합니다.

둘째 줄이 나쁜 소식입니다. 도구 설명문의 구두점 하나를 교체했을 뿐인데 문자 공통 접두사가 절반 이하로 줄었습니다. 도구 정의가 앞에 있으므로 그 뒤의 입력은 이전 도구 접두사와 더 이상 같지 않습니다. 도구 목록을 사용자마다 다르게 조립하거나 대화 중간에 추가하면 공유할 수 있는 캐시 접두사가 크게 줄어듭니다.

셋째 줄은 가장 흔한 실수입니다. 고정한 예제에서 시스템 프롬프트의 초만 바꿔도 57번째 문자부터 달라졌습니다. 정확한 위치는 앞 문장과 템플릿에 따라 달라지지만, **요청마다 달라지는 값을 앞쪽에 두면 그 뒤의 긴 고정 구간도 같은 접두사로 쓸 수 없다**는 결론은 같습니다.

### 무효화에는 층위가 있다

모든 변경이 전부를 날리는 건 아닙니다. Anthropic API의 캐시는 세 층으로 나뉘고, 변경은 자기 층과 그 아래만 무효화합니다.

| 바꾼 것 | tools 캐시 | system 캐시 | messages 캐시 |
|---|:---:|:---:|:---:|
| 도구 정의 추가·삭제·순서 변경 | 무효 | 무효 | 무효 |
| 모델 교체 | 무효 | 무효 | 무효 |
| 시스템 프롬프트 내용 | 유지 | 무효 | 무효 |
| 웹 검색·웹 가져오기·인용(citation)·속도(speed) 설정 켜고 끄기 | 유지 | 무효 | 무효 |
| `tool_choice`, `disable_parallel_tool_use`, 이미지 켜고 끄기 | 유지 | 유지 | 무효 |
| 사고(thinking) 설정 변경 | 모델별 | 모델별 | 무효 |
| 추론 노력(effort) 설정 변경 | 모델별 | 모델별 | 무효. 단, 모델 기본값을 명시하는 것은 생략과 동등 |
| 이전 메시지 수정·삭제·재정렬 | 유지 | 유지 | 무효 |
| 메시지를 끝에 추가 | 유지 | 유지 | 무효화 안 됨. 20-block 탐색 범위 안이면 기존 접두사 적중 |

**`tool_choice`를 매 요청 바꿔도 도구 정의와 시스템 프롬프트 캐시는 살아 있습니다.** 반면 도구 정의를 바꾸면 전부 다시 계산합니다. 이 표를 알면 "모드를 바꾸려고 도구 세트를 갈아 끼우는" 설계가 왜 비싼지, 대신 무엇을 해야 하는지가 보입니다 — 도구 세트를 고정해 두고 모드는 메시지 내용으로 전달하는 편이 훨씬 쌉니다.

### 블록이 빠르게 늘어날 때 생기는 함정: 20블록 되돌아보기

캐시를 쓸 위치(cache breakpoint)는 이전 요청에서 실제로 작성된 캐시 항목(cache entry)을 찾을 때 **현재 지점을 포함해 최대 20개 위치만 거슬러 올라갑니다.** 고정된 텍스트가 있다는 사실만으로는 부족하며, 그 위치에 과거 요청이 캐시 항목을 써 두었어야 합니다.

대화 전체가 20블록을 넘었다고 바로 미스가 나는 것은 아닙니다. 예를 들어 이전 요청이 10번 블록에 캐시 항목을 썼고 다음 요청의 breakpoint가 15번이면, 다섯 위치를 되돌아가 10번 항목을 찾을 수 있습니다. 반면 breakpoint가 35번인데 최근 항목이 여전히 10번뿐이면 35번부터 16번까지만 검사하므로 10번을 찾지 못합니다.

병렬 client tool은 한 왕복에서 `tool_use`와 `tool_result` 블록을 여러 개 추가하므로 이 간격을 빠르게 넓힐 수 있습니다. 하지만 같은 문제는 도구를 쓰지 않더라도 한 요청에 콘텐츠 블록을 많이 추가하면 생깁니다.

대응은 재사용할 고정 접두사 끝에 breakpoint를 미리 두고, 다음 요청에서 과거에 쓴 항목이 20개 위치의 검색 범위 안에 남도록 설계하는 것입니다. 명시적 breakpoint는 최대 네 개이므로 고정 간격으로 무조건 추가하기보다 실제 블록 증가량과 usage 값을 보고 배치해야 합니다.

### 캐시가 걸렸는지 확인하는 법

응답의 `usage`를 봅니다.

| 필드 | 의미 |
|---|---|
| `cache_creation_input_tokens` | 이번에 캐시에 쓴 토큰 수 |
| `cache_read_input_tokens` | 이번에 캐시에서 읽은 토큰 수 |
| `input_tokens` | 캐시를 못 쓰고 새로 처리한 토큰 수 |

같은 접두사로 반복 요청했는데 `cache_read_input_tokens`가 계속 0이면 접두사 불일치, 최소 크기 미달, cache write 실패, TTL 만료, 관련 설정 변경을 차례로 확인합니다.

주의할 점 하나. `input_tokens`는 **캐시에 걸리지 않은 나머지**입니다. 전체 프롬프트 크기가 아닙니다. 세 값을 더해야 전체입니다.

---

## 10. 흔한 오해

**"모델이 함수를 실행한다."** 아닙니다. 모델은 호출 요청을 표현한 출력을 만들고 멈춥니다. client tool은 호출한 애플리케이션이, server tool은 Anthropic 서버가 실행합니다. 3절.

**"도구 정의는 한 번만 보내면 된다."** 아닙니다. Messages API는 다음 요청에서 앞선 도구 정의와 대화를 자동으로 복원하지 않으므로 매 요청에 필요한 내용을 다시 보냅니다. 대화가 길어질수록 도구 정의도 매번 같이 실려 갑니다 — 캐시가 중요한 이유입니다. 3절, 9절.

**"도구 결과는 일반 사용자 발화와 구분되지 않는다."** 그렇지 않습니다. Anthropic에서는 user 메시지 안의 `tool_result` 블록으로, Qwen 예제에서는 `<tool_response>` 태그로 구분됩니다. 다만 구조화된 결과라고 해서 그 안의 외부 데이터까지 신뢰할 수 있는 것은 아닙니다. 5절.

**"도구 설명문은 문서화용 주석이다."** 아닙니다. 프롬프트 본문입니다. 6절.

**"client tool이 실패하면 결과를 안 보내면 된다."** 아닙니다. `is_error: true`를 붙여서 반드시 보냅니다. server tool 결과는 Anthropic 서버가 처리합니다. 7절.

**"병렬 client tool 결과는 여러 user 메시지로 나눠도 효과가 같다."** 아닙니다. 모든 결과를 바로 다음의 한 user 메시지에 담고, 일반 text가 있다면 result 블록 뒤에 둡니다. 결과 누락은 요청 거절 사유가 될 수 있고, 메시지 분할은 이후 병렬 호출을 줄일 수 있습니다. 8절.

**"JSON 스키마를 주면 유효한 JSON이 보장된다."** 도구 정의만으로는 보장되지 않습니다. 2절에서 본 대로 템플릿은 형식을 *부탁*합니다. 진짜 보장은 샘플링 단계에서 문법에 안 맞는 토큰을 막아야 나오고, 그건 별도 주제입니다.

---

## 부록 A. 재현 방법

이 문서의 렌더링 결과와 측정값을 다시 만드는 방법입니다. 모델 가중치는 필요 없지만, 템플릿과 토크나이저가 바뀌지 않도록 모델 revision을 고정합니다.

```bash
uv run --with transformers --with jinja2 python render.py
```

```python
# render.py
from copy import deepcopy
from transformers import AutoTokenizer

MODEL_ID = "Qwen/Qwen2.5-7B-Instruct"
REVISION = "a09a35458c702b33eeacc393d103063234e8bc28"
tokenizer = AutoTokenizer.from_pretrained(MODEL_ID, revision=REVISION)

TOOLS = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "특정 도시의 현재 날씨를 조회한다.",
        "parameters": {
            "type": "object",
            "properties": {"city": {"type": "string", "description": "도시 이름. 예: 서울"}},
            "required": ["city"],
        },
    },
}]

turn1 = [{"role": "user", "content": "서울 날씨 알려줘"}]
turn2 = turn1 + [
    {"role": "assistant", "content": "",
     "tool_calls": [{"function": {"name": "get_weather", "arguments": {"city": "서울"}}}]},
    {"role": "tool", "content": '{"city": "서울", "temp_c": 29, "condition": "흐림"}'},
]


def apply(messages, tools=TOOLS, tokenize=False):
    result = tokenizer.apply_chat_template(
        messages,
        tools=tools,
        add_generation_prompt=True,
        tokenize=tokenize,
        return_dict=tokenize,
    )
    return result["input_ids"] if tokenize else result


def common_prefix_len(a, b):
    return next(
        (i for i, (left, right) in enumerate(zip(a, b)) if left != right),
        min(len(a), len(b)),
    )


a_text, b_text = apply(turn1), apply(turn2)
a_tokens = apply(turn1, tokenize=True)
b_tokens = apply(turn2, tokenize=True)

assert "<tools>" in a_text and '"name": "get_weather"' in a_text
assert a_text.endswith("<|im_start|>assistant\n")
print("문자 접두사:", b_text.startswith(a_text),
      f"({len(a_text)}자 / {len(b_text)}자)")
print("토큰 접두사:", b_tokens[:len(a_tokens)] == a_tokens,
      f"({len(a_tokens)}토큰 / {len(b_tokens)}토큰)")

changed_tools = deepcopy(TOOLS)
changed_tools[0]["function"]["description"] = "특정 도시의 현재 날씨를 조회한다!"
changed_text = apply(turn1, tools=changed_tools)
print("설명 마침표를 느낌표로 교체:",
      common_prefix_len(a_text, changed_text), "자")

system_a = [{
    "role": "system",
    "content": "당신은 도우미입니다. 현재 시각은 2026-07-27 12:00:00입니다.",
}, *turn1]
system_b = [{
    "role": "system",
    "content": "당신은 도우미입니다. 현재 시각은 2026-07-27 12:00:01입니다.",
}, *turn1]
time_prefix = common_prefix_len(apply(system_a), apply(system_b))
print("시각 1초 변경:", time_prefix, "자 공통,",
      time_prefix + 1, "번째 문자부터 다름")
```

**PASS 조건**: 아래 네 줄이 차례로 나옵니다.

```text
문자 접두사: True (757자 / 975자)
토큰 접두사: True (188토큰 / 248토큰)
설명 마침표를 느낌표로 교체: 339 자
시각 1초 변경: 56 자 공통, 57 번째 문자부터 다름
```

**FAIL 신호**: revision이 다른데 숫자가 같을 것이라고 가정하거나, 문자 접두사만 확인하고 API 캐시 적중까지 증명했다고 해석하면 안 됩니다. 실제 캐시 적중은 해당 API의 usage 필드로 따로 확인해야 합니다.

`chat_template`이 없는 모델도 있습니다. 그런 경우 도구 호출을 지원하지 않거나 템플릿을 다른 방식으로 제공할 수 있으므로 모델 문서를 확인해야 합니다.

### 확인한 것과 확인하지 못한 것

| 항목 | 상태 |
|---|---|
| 도구 정의가 시스템 메시지 안에 JSON 텍스트로 들어간다 | **직접 관측** (Qwen2.5-7B-Instruct 템플릿 렌더링) |
| 도구 결과가 `user` 턴의 `<tool_response>`로 들어간다 | **직접 관측** (같은 템플릿) |
| 1차 요청이 2차 요청의 문자·토큰 접두사다 | **직접 측정** (문자 757/975, 토큰 188/248) |
| 설명문의 마지막 마침표를 느낌표로 바꾸면 문자 접두사가 757자에서 339자로 줄어든다 | **직접 측정** |
| 시스템 프롬프트의 시각을 1초 바꾸면 57번째 문자부터 달라진다 | **직접 측정** |
| 모델이 실제로 뱉는 원문 토큰 | **미검증.** 가중치를 돌리지 않았습니다. 템플릿이 지시하는 형식은 읽었지만 모델이 그 형식을 실제로 지키는지는 확인하지 못했습니다 |
| 상용 API 내부의 렌더링 결과 | **관측 불가.** 채팅 템플릿을 공개하지 않습니다. 변환은 서버 안에서 일어나고 우리는 양쪽 끝의 JSON만 봅니다 |
| Anthropic API의 실제 캐시 적중 | **미검증.** API를 호출하지 않았습니다. 캐시 규칙과 usage 필드는 공식 문서로 확인했습니다 |

마지막 두 줄이 중요합니다. **2절과 5절에서 눈으로 본 텍스트는 열린 가중치 모델의 것입니다.** 상용 API의 구체적인 내부 직렬화는 공개되지 않으므로, 로컬 템플릿 관측으로 대신 증명할 수 없습니다.

---

## 부록 B. 확인 시점과 출처

본문의 Qwen 렌더링 결과는 고정한 revision에서 직접 관측한 값입니다. client tool과 server tool의 실행 경계, Anthropic 요청·응답 필드, stop reason, 캐시 정책은 벤더 구현이라 바뀔 수 있습니다.

**확인 시점: 2026-07-27**
**확인 방법: 아래 Anthropic Claude API 공식 레퍼런스를 항목별로 대조**

| 사실 | 현재 값 | 공식 레퍼런스 |
|---|---|---|
| client tool의 핵심 필드 | `name`, `description`, `input_schema`가 `tools` 배열의 최상위에 놓임. `input_examples` 등 선택 필드는 별도 지원 | [도구 정의](https://platform.claude.com/docs/en/agents-and-tools/tool-use/define-tools) |
| 모델의 client tool 요청 | `{"type": "tool_use", "id", "name", "input"}` | [tool use 처리](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls) |
| client/server 실행 주체 | client tool은 호출 애플리케이션, server tool은 Anthropic 서버가 실행 | [tool use 동작](https://platform.claude.com/docs/en/agents-and-tools/tool-use/how-tool-use-works) |
| 멈춤 사유 | `end_turn`, `max_tokens`, `stop_sequence`, `tool_use`, `pause_turn`, `refusal`, `model_context_window_exceeded` | [stop reason 처리](https://platform.claude.com/docs/en/build-with-claude/handling-stop-reasons) |
| 결과 블록 | `{"type": "tool_result", "tool_use_id", "content"}`, 실패 시 `"is_error": true`; user 메시지 안에 둠 | [tool use 처리](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls) |
| 병렬 호출 | 기본 활성. 모든 결과를 **하나의** user 메시지에 담음. `disable_parallel_tool_use`로 제한 가능 | [병렬 tool use](https://platform.claude.com/docs/en/agents-and-tools/tool-use/parallel-tool-use) |
| 캐시 지정 | 최상위 `cache_control`을 쓰는 자동 방식과 콘텐츠 블록에 두는 명시적 방식 지원. 기본 TTL 5분, 선택적 1시간 | [prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| 캐시 지점 수 | 요청당 전체 breakpoint 슬롯은 4개. 자동 방식도 1개를 쓰므로 자동 방식과 함께라면 명시적 breakpoint는 최대 3개 | [prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| 캐시 접두사 계층 | `tools` → `system` → `messages` | [tool use와 caching](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-use-with-prompt-caching) |
| server tool 결과 캐시 | 요청에 `cache_control` marker가 하나 이상 있을 때 서버가 다음 내부 반복 전에 결과에 breakpoint를 자동 배치. 사용자가 지정한 TTL과 관계없이 항상 5분이며 `usage.cache_creation.ephemeral_5m_input_tokens`에 집계 | [tool use와 caching](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-use-with-prompt-caching) |
| 되돌아보기 창 | 지정 지점부터 최대 20개 블록 위치에서 과거 캐시 항목을 탐색 | [prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| 캐시 비용 | 읽기 0.1배, 쓰기 1.25배(5분 TTL) / 2배(1시간 TTL) | [가격표](https://platform.claude.com/docs/en/about-claude/pricing) |
| 단순 손익분기 | 5분은 총 2회 호출(1회 재사용), 1시간은 총 3회 호출(2회 재사용)부터 이득 | 위 가격으로 계산 |
| 최소 캐시 크기 | 모델마다 512 / 1024 / 2048 / 4096 토큰이며 세대순으로 단조롭지 않음 | [prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| 확인 필드 | `usage.cache_read_input_tokens`, `usage.cache_creation_input_tokens`, `usage.input_tokens` | [cache diagnostics](https://platform.claude.com/docs/en/build-with-claude/cache-diagnostics) |

최소 캐시 크기가 세대순이 아니라는 점은 직관에 어긋나니 적어 둡니다. 최신 모델이 512 토큰인데 한두 세대 전 모델이 4096인 경우가 있습니다. 3천 토큰짜리 프롬프트가 어떤 모델에서는 캐시되고 어떤 모델에서는 조용히 캐시되지 않습니다. 오류도 안 납니다.

또 하나. 이 표의 값은 **Anthropic 기준**입니다. 다른 벤더는 필드 이름이 다릅니다 (예: 결과를 담는 역할이 `tool`인 경우도 있습니다). 형태는 같지만 이름은 다르다는 게 이 계층의 성질입니다.

---

## 주요 근거

- [Anthropic tool use 정식 기능 발표, 2024-05-30](https://www.anthropic.com/news/tool-use-ga)
- [Anthropic tool use 동작](https://platform.claude.com/docs/en/agents-and-tools/tool-use/how-tool-use-works)
- [Anthropic tool use 처리](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls)
- [Anthropic prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
- [Qwen2.5-7B-Instruct, 고정 revision](https://huggingface.co/Qwen/Qwen2.5-7B-Instruct/tree/a09a35458c702b33eeacc393d103063234e8bc28)

---

## 스스로 확인할 질문

문서를 덮고 답해 보십시오.

1. "모델이 도구를 호출한다"는 문장에서 실제로 모델이 하는 일은 정확히 무엇인가. 나머지는 누가 하는가.
2. 도구 정의는 매 요청마다 전송되는가. 그렇다면 왜 그런가.
3. 도구 실행 결과는 어떤 역할과 콘텐츠 블록으로 대화에 들어가는가. 타입 경계와 신뢰 경계는 왜 다른가.
4. client tool 세 개를 한 번에 요청받았다. 다음 user 메시지의 content 블록을 어떤 순서와 묶음으로 구성해야 하는가.
5. 대화만 길어지는 경우와 도구 설명문을 한 글자 고치는 경우 중, 프롬프트 캐시에 더 큰 타격을 주는 쪽은 어디이며 그 이유는 무엇인가.
