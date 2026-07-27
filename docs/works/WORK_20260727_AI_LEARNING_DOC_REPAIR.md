# AI 학습 문서 두 편 정확성·재현성 수리

## 0. Meta

- 작업 ID: `AI-DOC-REPAIR-20260727`
- whole-request objective: `WR-AI-DOC-REPAIR-20260727`
- stage / tranche: `DOC-REPAIR / T1`
- stage/tranche registry source: 이 WORK 문서
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형 / 깊이: `refactor_docs / full`
- 대상:
  - `ai/tool_calling.md`
  - `ai/prompt_caching.md`
- 실행 기록과 Claude 인계:
  - `docs/works/WORK_20260727_AI_LEARNING_DOC_REPAIR.md`
- finish: `test+commit`
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`

## 1. Request Normalization

### 목표

독립 검수에서 확인된 사실 오류, 논리 도약, 재현 불일치, 검증 라벨 불일치와 두 문서 사이의 드리프트를 실제 문서에서 수리한다. 수정 이유와 보존해야 할 경계를 Claude가 다시 확인할 수 있도록 이 문서에 남긴다.

### 고정 범위

- 표준 causal decoder의 층별 K/V 의존 논증과 생략 전제를 명시한다.
- Anthropic의 현재 tool use·prompt caching 레퍼런스를 다시 대조한다.
- 두 부록 A의 코드를 문서에서 직접 추출해 실행하고 기대값과 맞춘다.
- 본문, “확인한 것과 확인하지 못한 것”, 부록 B의 신뢰 수준을 일치시킨다.
- 등장 배경, 실제 artifact, 오해 교정, 검증 경로, teach-back을 보존·보강한다.
- 무효화 표, 20-block lookback, `tools → system → messages`, 상호 링크를 동기화한다.
- 자연스럽지 않은 직역과 내부 작업용 표현이 독자 문장에 새지 않도록 고친다.

### 비범위

모델 학습 과정, 샘플링과 제약 디코딩, 구조화 출력, 임베딩과 검색, 양자화, 에이전트 루프 일반론, 프롬프트 인젝션 상세, Claude Code 같은 하네스 구현은 확장하지 않는다. 기존 문장이 이 경계를 건드리는 곳은 범위를 표시하는 한두 문장만 남긴다.

### 변경 허용 목록

- `ai/tool_calling.md`
- `ai/prompt_caching.md`
- `docs/works/WORK_20260727_AI_LEARNING_DOC_REPAIR.md`

다른 dirty worktree 파일은 사용자 작업으로 간주하며 stage하거나 수정하지 않는다.

## 2. 적용 계약과 품질 하한

- 적용 문서:
  - 전역·로컬 `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `PROJECT_INTENT.md`
  - `USECASE.md`
  - `ai/authoring/LEARNING_DOC_GUIDE.md`
- `TERMINOLOGY.md`: 저장소에 없음
- 적용 스킬:
  - `study-explanation`: 실제 artifact, 배경, guided trace, 오해 교정, replay, teach-back
  - `rigorous-task`: 근거 수준, 반례, 검증, 상태 정직성
  - `review-kernel`: 독립 검수 결과를 수정 가능한 주장 단위로 수리
- primary exemplar: `computer_architecture/ostep/xv6-riscv/kernel/entry.S`
  - 참고한 원리: 첫 상태와 숨은 전제를 먼저 고정하고, 아래층 조건이 상위 결론을 어떻게 지탱하는지 추적
  - 따라 하지 않은 형태: 현재 주제에 필요하지 않은 과도한 line-by-line 주석
- secondary exemplar: `algorithms/dynamic_programming.md`
  - 참고한 원리: 작은 입력과 상태 관계에서 일반 원리로 올라가는 방식
  - 따라 하지 않은 형태: 정의와 분류가 초반을 길게 차지하는 구성

## 3. Root-First Framing과 결정

### 근본 문제

두 문서는 설명의 뼈대와 예제는 강했지만, AI 작성물의 전형적인 실패가 하중을 지탱하는 문장에 남아 있었다.

- Qwen 템플릿 관측을 상용 API 내부 구현으로 일반화했다.
- 장난감 모형의 숫자를 일반 정리의 증명처럼 사용했다.
- 실제 스크립트가 만들지 않는 `18번째 문자` 값을 직접 측정처럼 적었다.
- Anthropic API의 바뀐 기능과 필드를 누락하거나 오래된 값으로 고정했다.
- 문자 접두사, 토큰 접두사, 실제 API cache hit을 같은 검증으로 취급했다.

### 고려한 두 접근

1. 전면 재작성
   - 장점: 구조를 처음부터 통일할 수 있다.
   - 탈락 이유: 이미 좋은 실제 artifact, 짧은 직답, 왕복 trace, replay 질문을 불필요하게 잃고 새 오류를 만들 가능성이 크다.
2. 하중을 지탱하는 주장 중심의 동기화 수리
   - 장점: 좋은 teaching spine은 보존하면서 사실·추론·관측 경계를 정확히 고칠 수 있다.
   - 채택 이유: 독립 검수 finding을 위치별로 역추적하고 두 문서의 겹치는 표와 값을 함께 고칠 수 있다.

### decision envelope

- decision type: `review findings → document repair`
- support tier: 저장소 스크립트 실행과 공식 레퍼런스는 `T1`; L1 층별 논증은 명시한 전제 아래의 `T2 strong inference`
- admission lane: `APPLY`
- success: 재현값 일치, 현재 공식 사실 일치, 과잉 일반화 제거, 교차 표·링크 일치
- failure: live API를 호출하지 않고 적중을 실측했다고 쓰거나, toy output을 일반 증명으로 남기거나, 현재 레퍼런스와 다른 값을 남김
- verification path: 문서 코드 블록 직접 실행, 공식 URL·내부 anchor 확인, 두 표의 행 단위 비교, 독립 critic 확인

## 4. Frozen Checklist

검수 보고의 발견을 아래 체크리스트로 고정했다. 구현 중 삭제하거나 완화하지 않는다.

- C-01 L1 층별 논증
  - PASS: causal mask, 동일 모델·토큰·설정, 기존 접두사 영역의 같은 attention 연결, position id와 실제 position encoding/scaling 보존, 미래 정보 혼합 없음이 전제로 드러나며 수치 재현성 경계를 설명한다.
- C-02 장난감 모형의 증거 범위
  - PASS: masked 실행은 `7, 0, 8`이고 equality assertion이 모두 통과하며, `apply_causal_mask=False`인 tail 비교는 첫 값을 정확히 `7 → 0`으로 바꾼다. 이 결과는 예시로 남고 일반 결론은 층별 논증이 지탱한다. `same_prefix_len == 0`과 “전 위치가 다름”을 구분한다.
- C-03 tool calling 재현
  - PASS: 고정 Qwen revision에서 문자 `757/975`, 토큰 `188/248`, 설명문 `339`, 시각 `56/57`이 문서 코드 그대로 재현된다.
- C-04 Anthropic tool use
  - PASS: client/server 실행 주체를 섞지 않고, 모든 client `tool_result`를 바로 다음의 한 user 메시지에 함께 넣으며 일반 text보다 먼저 둔다는 계약, 결과 누락의 요청 거절 위험과 메시지 분할의 병렬성 저하, 현재 stop reasons가 공식 레퍼런스와 맞는다.
- C-05 Anthropic prompt caching
  - PASS: 자동/명시적 캐싱, 전체 breakpoint 4개와 자동 방식의 1개 슬롯 점유, 5분/1시간 TTL, 가격 배수, 최소 토큰, usage, 20-block 탐색, 최신 invalidation과 effort 기본값 예외가 공식 레퍼런스와 맞는다. server tool 결과 자동 breakpoint는 `cache_control` marker가 있어야 생기고 항상 5분임을 구분한다.
- C-06 사실·추론·미검증 라벨
  - PASS: Qwen 직접 관측, L1 논증, Anthropic 문서 확인, live API 미실측을 분리한다.
- C-07 교차 일관성
  - PASS: 무효화 표가 행 단위로 같고, 20-block 의미와 `tools → system → messages`의 성격이 같으며 상호 링크가 실제 heading을 가리킨다.
- C-08 학습 문서 계약
  - PASS: 직답, 실제 artifact, 등장 배경, guided trace, 오해 교정, replay, 근거, teach-back이 남아 있다.
- C-09 표현
  - PASS: “질문을 닫는다”, “납작한” 같은 내부/직역 표현이 없고, 필요한 영어 용어는 쉬운 한국어와 연결된다.
- C-10 범위·worktree·closure
  - PASS: 허용 파일만 변경·stage하며 최종 검증과 독립 critic 확인 뒤 로컬 commit한다. push하지 않는다.

체크리스트 품질 검수: 각 항목은 사용자 범위에 1:1로 대응하고 실행 결과나 문서 diff로 PASS/FAIL을 판정할 수 있다. 독립 critic이 client/server 결과 계약과 최신 캐시 예외를 더 구체적으로 고정하라고 지적해 C-04·C-05를 강화했다. freeze 버전은 `v2`, freeze 시점은 2026-07-27이다.

## 5. Evidence / Claim Ledger

### E-01 Qwen 렌더링

- 근거: 고정 revision `a09a35458c702b33eeacc393d103063234e8bc28`의 `AutoTokenizer.apply_chat_template`
- 닫힌 주장: Qwen에서 도구 정의와 결과가 어떤 문자열·토큰 접두사를 만드는지
- 닫히지 않는 주장: Anthropic 내부 직렬화, 모델의 실제 도구 선택률, API cache hit

### E-02 causal 의존 구조

- 근거: causal self-attention의 층별 귀납 논증과 NumPy 최소 모형
- 닫힌 주장: 명시한 전제 아래 접두사 뒤에 덧붙인 토큰은 기존 위치의 K/V에 영향을 주지 않는다.
- 반례 경계: 위치 재배치, 미래 정보 혼합, 설정 변경, 비표준 attention 경계
- 모형의 한계: 무작위 가중치, 작은 차원, residual/MLP/normalization 생략

### E-03 Anthropic 현재 구현

- 근거: 2026-07-27에 확인한 Claude Platform 공식 문서
- 확인한 표면:
  - tool use 동작·정의·결과·병렬 처리
  - stop reasons
  - prompt caching·cache diagnostics·pricing
  - tool use with prompt caching
- 닫히지 않는 주장: 유료 live API의 실제 적중과 latency 개선 폭

### 소스 충돌 처리

Anthropic의 이전 prompt caching 발표 페이지는 현재 마이그레이션된 페이지에 2025년 날짜가 표시되지만 본문에는 2024년 12월 GA 업데이트가 남아 있다. 등장 시점은 날짜가 명시된 Claude Platform release notes의 2024-08-14 beta 항목을 근거로 썼다.

## 6. 실제 수정

### `ai/tool_calling.md`

- client tool과 server tool의 실행 주체를 분리했다.
- Qwen의 열린 템플릿 관측과 Anthropic의 공개 API 계약을 분리했다.
- `tool_result`가 user 메시지 안에서도 typed block으로 구분된다는 점과 신뢰 경계가 별개라는 점을 바로잡았다.
- `max_tokens` 처리와 `model_context_window_exceeded`를 현재 stop reason 기준으로 고쳤다.
- 설명문 효과를 미측정 성능 향상처럼 쓰지 않고 선택 근거를 제공한다는 수준으로 낮췄다.
- 병렬 결과가 가중치를 다시 “학습”시키는 것처럼 보이던 문장을 현재 대화 문맥의 영향으로 고쳤다.
- Qwen revision을 고정하고 문자·토큰 접두사와 두 mutation을 한 스크립트에서 측정한다.
- 20-block 규칙을 “누적 20블록”이 아니라 과거 write를 최대 20개 위치에서 찾는 규칙으로 고쳤다.

### `ai/prompt_caching.md`

- KV cache를 기술적으로 끌 수 없다는 주장을 제거하고 엔진 설정과 API 제어 경계를 분리했다.
- prefill/decode 병목을 하드웨어·배치에 따른 일반적 경향으로 한정했다.
- L1 결론의 전제와 position encoding 경계를 명시하고 층별 귀납 논증을 보강했다.
- 수학적 동일성과 GPU의 비트 단위 재현성을 구분했다.
- toy output을 증명이 아니라 예시와 반대 조건 실험으로 낮추고 전 위치 equality assertion을 추가했다.
- 상용 API 내부 렌더링 diff 대신 canonical request diff와 usage 확인을 검증 경로로 제시했다.
- 자동 캐싱, 최신 invalidation 항목, server tool 결과 캐시, 정확한 손익분기 표현을 추가했다.
- 자동 캐싱이 네 breakpoint 슬롯 중 하나를 쓰는 점과, server tool 결과의 자동 breakpoint가 marker를 요구하고 항상 5분을 쓰는 점을 명시했다.

### 두 문서 공통

- 무효화 표와 부록 B를 현재 공식 레퍼런스에 맞춰 동기화했다.
- `tools → system → messages`를 Anthropic의 공개된 캐시 접두사 계층이라고 표시했다.
- 20-block 예시를 같은 의미로 맞추고 상호 링크를 실제 heading에 연결했다.
- 역사·등장 배경과 주요 근거를 보강했다.

## 7. Verification Log

### 완료

- 문서의 `render.py` 블록을 추출해 실행:
  - `True (757자 / 975자)`
  - `True (188토큰 / 248토큰)`
  - `339자`
  - `56자 공통, 57번째 문자부터 다름`
- 문서의 `causal.py` 블록을 추출해 실행:
  - `7, 0, 8`
  - causal mask 미적용: 첫 값 `7 → 0`
  - 세 equality assertion 통과
- 내부 Markdown 파일·anchor 검사: 73개 Markdown link PASS
- 외부 공식/논문 링크 17개 HTTP 확인: PASS
- 두 무효화 표의 행 단위 출력 비교: 동일
- `git diff --check` 대상 3개: PASS
- Axiom 최종 확인: L1 전제, 장난감 모형의 증거 범위, 두 스크립트 결과 모두 `CONFIRMED`
- Atlas 최종 확인: Anthropic L3 사실과 부록 B의 행별 근거 모두 `CONFIRMED`
- Sentry 최종 확인: 작업 정의, 성공 기준, frozen checklist v2, 실행 결과·Claude 인계 모두 `CONFIRMED`
- stage allowlist: `ai/tool_calling.md`, `ai/prompt_caching.md`, 이 WORK 문서만 포함. PASS

### 남은 검증

- 없음

## 8. Multi-Agent Deliberation Ledger

역할 roster:

- Axiom: L1 논증·재현 실험 critic
- Atlas: Anthropic 현재 API 사실 critic
- Sentry: 저장소 학습 계약·검증 라벨·교차 일관성 sentinel
- Orchestrator: finding 수리, 충돌 중재, 최종 검증과 closure

라운드:

1. 최초 독립 검수: 세 역할이 원문을 분리 검토했고, BLOCKER/MAJOR finding을 위치와 근거로 합성했다. 완료.
2. 작업 정의·성공 기준·체크리스트 공격: Sentry가 최신 사용자 수정 권한을 확인하고 client/server 결과 계약과 C-04·C-05의 빠진 판정 조건을 공격했다. 권한 의문은 최신 사용자 메시지로 해소했고 checklist를 v2로 강화했다. 완료.
3. L1·재현 결과 공격과 수리: Axiom이 position encoding scaling과 접두사 attention 연결 전제, unmasked `7 → 0` 고정 조건을 지적했다. 본문과 C-01·C-02를 수리한 뒤 네 artifact를 `CONFIRMED`했다. 완료.
4. 벤더 사실·교차 표 공격과 수리: Atlas가 automatic breakpoint 슬롯, server tool marker·5분 TTL, effort 기본값, web fetch, stop reason beta 경계를 지적했다. 두 표와 인계문을 동기화한 뒤 네 artifact를 `CONFIRMED`했다. 완료.
5. 최종 결과·인계문·closure 공격: Sentry가 client-only `tool_result`, 하네스 비범위, MQA 역사 인과, causal 과장, Claude 인계 누락을 공격했다. 수리 뒤 스크립트·표·링크·라벨·학습 계약을 다시 확인하고 네 artifact를 `CONFIRMED`했다. 완료.

cost / rigor receipt: 두 AI 작성 문서가 같은 수치와 벤더 사실을 서로 재인용하고 있어 한 사람이 한 번 읽는 방식으로는 상호 강화된 오류를 놓치기 쉽다. 독립 역할 비용은 L1 일반화 오류, 최신 API drift, 검증 라벨 과장을 서로 다른 근거로 공격하기 위해 사용한다.

## 9. Claude에게 전달할 인계문

아래 블록은 Claude에게 그대로 전달할 수 있는 요약이다.

> `ai/tool_calling.md`와 `ai/prompt_caching.md`를 독립 검수 결과에 따라 수리했습니다. 핵심 목적은 문장을 더 그럴듯하게 만드는 것이 아니라, 직접 관측·수학적 추론·공식 문서 확인·미실측을 서로 다른 신뢰 수준으로 분리하는 것이었습니다.
>
> `tool_calling.md`에서는 client tool과 server tool의 실행 주체를 분리하고, Qwen 템플릿 관측을 Anthropic 내부 렌더링 증거처럼 쓰지 않도록 고쳤습니다. 모든 `tool_result` 의무는 client tool 계약이며, 병렬 결과는 바로 다음의 한 user 메시지에서 일반 text보다 먼저 와야 합니다. `tool_result`는 user 메시지 안에서도 typed block으로 구분되며, 외부 데이터의 신뢰 문제는 그 타입 구분과 별개라고 정리했습니다. 재현 스크립트는 Qwen revision을 고정하며 문자 `757/975`, 토큰 `188/248`, 설명문 mutation `339`, 시각 mutation `56/57`을 실제로 출력합니다.
>
> `prompt_caching.md`에서는 “모든 엔진에서 KV cache를 끌 수 없다”는 문장을 제거했습니다. L1 결론은 causal mask만 언급하고 끝내지 않고, 동일 모델·토큰·설정, 기존 position id 보존, 미래 정보를 섞는 연산 없음이라는 전제를 둔 층별 귀납 논증으로 고쳤습니다. 위치 인코딩과 GPU 수치 재현성의 경계도 추가했습니다. 어휘 20·차원 8·3층 모형의 `7/0/8`은 일반 증명이 아니라 예시와 negative control이며, `same_prefix_len == 0`만으로 전 위치가 다르다고 하지 않도록 equality assertion을 넣었습니다.
>
> Anthropic 표는 2026-07-27 공식 레퍼런스로 다시 맞췄습니다. 자동/명시적 prompt caching, 5분·1시간 TTL, 1.25x·2x·0.1x 가격, 전체 4개 breakpoint 슬롯, 20-block lookback, 현재 stop reasons, 공개된 논리적 캐시 계층 `tools → system → messages`, thinking/effort의 모델별 invalidation, server tool 결과 자동 cache breakpoint를 반영했습니다. 자동 방식도 한 슬롯을 사용하며, server tool 결과의 자동 지점은 요청에 marker가 있을 때만 생기고 항상 5분 TTL을 씁니다. effort는 모델 기본값을 명시한 경우 생략과 동등하고, web fetch도 system/messages 무효화 항목이며, `model_context_window_exceeded`는 이전 모델·SDK에서 beta 경계가 있을 수 있습니다. 손익분기는 5분 TTL이 총 2회 호출(1회 재사용), 1시간 TTL이 총 3회 호출(2회 재사용)부터 이득이라고 명확히 썼습니다.
>
> 이후 수정할 때 지켜야 할 경계는 세 가지입니다. Qwen 문자열을 Anthropic 내부 문자열로 일반화하지 말 것, toy 숫자를 L1 정리의 증명으로 승격하지 말 것, live API를 호출하지 않았으므로 `cache_read_input_tokens`와 latency를 실측했다고 쓰지 말 것입니다. 현재 공식 사양이 바뀌면 두 문서의 부록 B와 무효화 표를 함께 갱신해야 합니다.

## 10. Closure

- requested closure scope: 두 문서 수리, 검증, Claude 인계, 로컬 commit
- achieved closure scope: 두 문서 수리, 스크립트·링크·교차 표 최종 검증, 세 독립 critic-confirmation, Claude 인계 완료
- whole-request status: `WHOLE_COMPLETE`
- remaining count: 0
- next immediate target: 없음
- remaining open tranche: 없음
- downstream impact: 로컬 문서와 commit만 바뀌며 push·배포·외부 쓰기는 없다. `git diff`로 되돌릴 수 있다.
- 최종 상태: `COMPLETE`
- 완료 게이트: `PASS`
- commit: 이 문서를 포함한 로컬 commit. 실제 hash는 `git log -1 --oneline`으로 확인
