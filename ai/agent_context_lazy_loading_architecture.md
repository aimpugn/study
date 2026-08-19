# 범용 에이전트 컨텍스트 & 점진적 로딩(Lazy Loading) 아키텍처

> **문서 목적**: 특정 서비스나 도메인 종속성을 완전히 배제하고, AI 에이전트 환경에서 **프롬프트 비대화(Prompt Bloat)와 비결정적 환각을 방지하며, 컴팩션(Compaction) 시에도 규칙을 100% 보존하는 컨텍스트 관리 및 점진적 로딩 아키텍처의 순수 뼈대**를 정의한다.

---

## 1. 해결하고자 하는 핵심 문제 (Problem Definition)

대규모 룰셋, 도메인 가이드라인, 코딩 지침을 에이전트에게 적용할 때 다음 4가지 실패가 발생한다:

1. **프롬프트 비대화 (Prompt Bloat)**: 수십 개의 룰 파일과 수천 줄의 분류표를 첫 턴부터 시스템 프롬프트에 전부 넣으면 컨텍스트 윈도우 비용과 지연 시간이 급증함.
2. **주의력 분산 (Attention Dilution / Lost in the Middle)**: 프롬프트가 과도하게 길어지면 LLM이 사용자의 현재 지시사항과 핵심 불변식을 망각함.
3. **비결정적 로딩 (Non-deterministic Execution)**: 파일 도구(`view_file`) 호출을 순수 LLM의 자율에만 맡기면, 필요한 규칙을 읽지 않고 기억(추측)에 의존해 환각을 일으킴.
4. **컴팩션 유실 (Compaction Loss)**: 대화가 길어져 컨텍스트 요약(Compaction)이 발생할 때 룰과 지침이 함께 요약·삭제되어버림.

---

## 2. 5계층 아키텍처 (The 5-Tier Architecture)

이 아키텍처는 **"상시 주입 불변식 -> 스코프 라우터 -> 점진적 로딩 -> 결정론적 하네스 검증 -> 피드백 캘리브레이션"**의 5단계 파이프라인으로 동작한다.

```mermaid
flowchart TD
    subgraph Tier0["Tier 0: 상시 주입 계층 (Eager Invariants)"]
        G["글로벌 불변식 (Global Core Doctrine)"]
        W["워크스페이스 오버레이 (AGENTS.md)"]
    end

    subgraph Tier1["Tier 1: 선행 게이트 & 스코프 라우터 (Fast Scope Router)"]
        F["작업 프레임 판정 (Task Frame: 목적/독자/장르/제약)"]
        S["스킬 메타데이터 (Thin 2-3 Line Protocol)"]
    end

    subgraph Tier2["Tier 2: 점진적 공개 계층 (Progressive Disclosure / Lazy Loading)"]
        D["세부 참조 문서/분류 체계 (view_file로 온디맨드 로드)"]
    end

    subgraph Tier3["Tier 3: 결정론적 하네스 계층 (Deterministic Harness)"]
        H["프로그램 기반 검증기 (CLI Scripts / Linter / Tests)"]
    end

    subgraph Tier4["Tier 4: 피드백 캘리브레이션 계층 (Generalization Loop)"]
        C["피드백 분석 -> 5단계 일반화 검증 -> 룰셋 영구 승격"]
    end

    Tier0 -->|매 턴 100% 강제 주입| Tier1
    Tier1 -->|조건 부합 시에만 호출| Tier2
    Tier2 --> LLM["LLM 추론 및 초안 생성"]
    LLM --> Tier3
    Tier3 -->|Exit Code 0 (PASS)| Output["최종 결과물 확정"]
    Tier3 -->|Exit Code != 0 (FAIL)| Tier2
    Output -.->|사용자 피드백 발생 시| Tier4
    Tier4 -.->|일반화된 룰 영구 반영| Tier0
```

---

## 3. 계층별 상세 명세

### [Tier 0] 상시 주입 계층 (Eager System Invariants)
* **역할**: 대화 컴팩션(요약)과 무관하게 항상 유지되어야 하는 최상위 불변식.
* **주입 방식**: 런타임 전처리기(Runtime Preprocessor)가 시스템 프롬프트 최상단에 강제 주입 (LLM의 호출 여부와 무관하게 100% 결정적 로드).
* **내용**:
  * 사실 우선 원칙 (Fact > User Opinion / Speculation)
  * 가설 검증 및 실패 복구 루프 (Reasoning & Reality Loop)
  * 안전성, 프라이버시, 파괴적 작업 차단 경계

### [Tier 1] 선행 게이트 & 스코프 라우터 (Scope Routing & Task Framing)
* **역할**: 작업의 유형과 목적을 판정하여 불필요한 규칙이 활성화되는 것을 방지.
* **Writing Task Frame (작업 프레임 5대 요소)**:
  1. `Request Type`: 신규 작성(Draft) / 부분 수정(Rewrite) / 검수(Review) / 캘리브레이션(Calibrate)
  2. `Reader`: 누가 읽고 어떤 판단을 내릴 것인가?
  3. `Genre`: 지원서, 기술 보고서, API 명세, 설계서, 회고록 등
  4. `Evidence Boundary`: 현재 사용 가능한 참(True)인 팩트의 범위
  5. `Non-goals`: 목적을 흐리거나 추측에 기반한 금지 영역
* **스킬 메타데이터**: 2~3줄의 얇은 설명(Name, Description)만 상시 노출하여 프롬프트 토큰 절약.

### [Tier 2] 점진적 공개 계층 (Progressive Disclosure / Lazy Loading)
* **역할**: 800줄 이상의 대용량 참조 문서, 상세 룰셋, 어휘 사전을 **필요한 순간에만** 워킹 메모리에 로드.
* **결정론적 강제 프로토콜 (System Mandatory Directive)**:
  > *"특정 스킬이나 도메인 작업에 착수할 때, 에이전트는 지침에 명시된 참조 파일(`references/deep-rule.md`)을 `view_file` 도구로 완전히 읽은 후에만 출력을 생성해야 한다."*
* **효과**: 프롬프트 오염을 막으면서도, 모델이 자의적 기억 대신 최신 파일 스펙을 직접 읽고 정확하게 추론함.

### [Tier 3] 결정론적 하네스 계층 (Deterministic Verification)
* **역할**: LLM의 주관적 완료 선언("잘 작성되었습니다")을 불신하고, **순수 프로그램 코드로 결과물을 검증**.
* **하네스 유형**:
  * **코퍼스/패턴 검증기**: Ripgrep 기반 금지어/어색한 연어 실측 검사 (`check_collocation.sh`)
  * **제약/예산 검증기**: 글자 수, 라인 수, 토큰 상하한 엄격 검사 (`check_budget.py`)
  * **구조/스키마 검증기**: JSON Schema, Markdown 헤더 구조 검증 (`validate_schema.py`)
* **종결 기준**: 스크립트의 리턴 코드가 `Exit Code 0`일 때만 작업을 종료.

### [Tier 4] 피드백 캘리브레이션 계층 (Calibration & Generalization Loop)
* **역할**: 피드백이 발생했을 때 단순 문자열 치환에 머무르지 않고, **작성자의 판단 의사결정 프로세스(Decision Layer)를 일반화하여 영구 룰로 승격**.
* **5단계 검증 프로세스**:
  1. `Representative Case`: 최초에 실패했던 사례가 해결되는가?
  2. `Near Transfer Case`: 같은 장르의 다른 문서에서도 동일하게 자연스러운가?
  3. `Far Transfer Case`: 다른 장르(기술문서, 보고서 등)에 적용해도 망가지지 않는가?
  4. `Clean Control`: 원래 정상이었던 대조군 문장을 훼손하지 않는가?
  5. `Adversarial Control`: 과적합(과도한 억제/검열)이 발생하지 않는가?

---

## 4. 이식 가능한 최소 파일 구조 (Portable Template File Tree)

새 프로젝트에 이 아키텍처를 도입할 때는 아래 5개 파일/디렉토리 구조를 생성합니다:

```text
my-project/
├── .agents/
│   └── AGENTS.md                  <-- [Tier 0] 워크스페이스 핵심 불변식 및 오버레이
├── skills/
│   └── domain-skill/
│       ├── SKILL.md               <-- [Tier 1] 얇은 진입 메타데이터 & 필수 선행 로드 지시
│       └── references/
│           └── deep-rule-ssot.md  <-- [Tier 2] 필요 시에만 view_file로 읽는 세부 참조 문서
├── scripts/
│   └── verify_harness.py          <-- [Tier 3] 글자 수/규격/불변식 결정론적 검증기
└── calibration/
    └── feedback_ledger.md         <-- [Tier 4] 피드백 발생 시 일반화된 룰을 누적하는 원장
```

---

## 5. 템플릿 코드 및 명세 (Copy-Paste Blueprints)

### (1) `AGENTS.md` (Tier 0: 핵심 불변식 템플릿)

```markdown
# AGENTS.md — Universal Execution Invariants

## 1. 사실과 현실 우선 (Reality Over Speculation)
- 공식 문서, 소스 코드, 실행 사실, 테스트 결과가 사용자나 AI의 추측보다 항상 우선한다.
- 불확실하면 먼저 조사하고, 사실로 반박하며, 가정이 틀렸을 때는 즉시 방향을 수정한다.

## 2. 점진적 로딩 프로토콜 (Progressive Disclosure Protocol)
- 방대한 세부 참조 문서는 상시 프롬프트에 올리지 않는다.
- 작업 대상 장르나 스킬이 확정되면, 지침에 링크된 참조 파일을 도구(view_file)로 먼저 검사한 후 실행한다.

## 3. 결정론적 검증 게이트 (Deterministic Verification Gate)
- AI 자신의 주관적 판단만으로 작업을 완료로 선언하지 않는다.
- 지정된 검증 스크립트(scripts/verify_harness.py)를 실행하여 성공(Exit Code 0)할 때만 최종 완료한다.
```

### (2) `SKILL.md` (Tier 1: 얇은 진입 프로토콜 템플릿)

```markdown
---
name: domain-processor
description: Process domain-specific tasks according to strict quality and style constraints.
---

# Domain Processor Protocol

When this skill is activated:
1. You MUST read the detailed specification at `references/deep-rule-ssot.md` using `view_file` BEFORE generating any output.
2. Formulate the Task Frame (Request Type, Reader, Genre, Evidence Boundary, Non-goals).
3. Draft output strictly based on the admitted facts.
4. Execute `python3 scripts/verify_harness.py` to deterministically validate output constraints.
```

### (3) `scripts/verify_harness.py` (Tier 3: 결정론적 검증기 템플릿)

```python
#!/usr/bin/env python3
"""
Deterministic Quality & Boundary Harness
Returns 0 on success, 1 on failure.
"""
import sys
import re
from pathlib import Path

def verify_file(file_path: Path) -> bool:
    if not file_path.exists():
        print(f"FAIL: File {file_path} not found.")
        return False

    content = file_path.read_text(encoding="utf-8")
    
    # 1. 길이 및 예산 제약 검증 (공백 포함 글자 수)
    char_count = len(content)
    if char_count < 300 or char_count > 1000:
        print(f"FAIL: Length constraint violated (300 <= {char_count} <= 1000)")
        return False

    # 2. 금지된 추상/과장 패턴 검증
    banned_patterns = [r"압도적인", r"혁신적인", r"시너지", r"오너십", r"드라이브", r"흐름을 닫다"]
    for pat in banned_patterns:
        if re.search(pat, content):
            print(f"FAIL: Banned pattern detected: {pat}")
            return False

    # 3. 필수 마크다운 헤더 구조 검증
    if not re.search(r"^#+\s+", content, re.MULTILINE):
        print("FAIL: Missing structured markdown headers.")
        return False

    print(f"PASS: {file_path} ({char_count} chars) passed all deterministic gates.")
    return True

if __name__ == "__main__":
    target = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("output.md")
    sys.exit(0 if verify_file(target) else 1)
```

### (4) `calibration/feedback_ledger.md` (Tier 4: 피드백 일반화 원장 템플릿)

```markdown
# Feedback Calibration Ledger

## Format
- **Date**: YYYY-MM-DD
- **Episode**: Rejected candidate vs User corrected candidate
- **Failing Layer**: Stance / Slot Mismatch / Over-assembly / Collocation
- **General Principle**: Abstract rule derived (Never a simple keyword ban)
- **Scope**: Applied genre and role
```

---

## 6. 멀티 런타임 연결 및 이식 가이드

각 AI 도구 환경에서 Tier 0 최상위 룰을 인식시키는 방법:

| AI 런타임 | 글로벌 SSOT 연결 방식 | 워크스페이스 로컬 오버레이 |
| :--- | :--- | :--- |
| **Codex** | `~/.codex/AGENTS.md` -> 글로벌 `AGENTS.md` 심링크 | 프로젝트 루트의 `AGENTS.md` 자동 로드 |
| **Claude Code** | `~/.claude/CLAUDE.md` 내에 `@/path/to/global/AGENTS.md` 선언 | 프로젝트 루트의 `CLAUDE.md` 자동 로드 |
| **Gemini / Antigravity** | `~/.gemini/GEMINI.md` 내에 `@/path/to/global/AGENTS.md` 선언 | `<RULE[workspace/AGENTS.md]>` 자동 주입 |

---

## 7. 동작 테스트 절차

1. 새 저장소에 위 **최소 파일 구조(5개 파일)**를 생성합니다.
2. 최상위 `AGENTS.md`에 **3대 불변식**을 배치합니다.
3. 세부 지침 문서를 `references/deep-rule-ssot.md`에 배치하고, `SKILL.md`에 **선행 읽기 강제 조항**을 작성합니다.
4. 에이전트에게 작업을 요청하면:
   - 프롬프트 낭비 없이 스킬 메타데이터만 인식하고 있다가,
   - 작업 시작 시 `view_file`로 세부 문서를 정확히 로드한 뒤,
   - 결과물 작성 후 `verify_harness.py`를 실행하여 통과 여부를 검증하는 완전한 라이프사이클이 작동합니다.
