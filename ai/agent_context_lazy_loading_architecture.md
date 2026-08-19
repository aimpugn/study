# 범용 에이전트 컨텍스트 & 점진적 로딩(Lazy Loading) 아키텍처

> **문서 목적**: 특정 서비스나 도메인 종속성을 완전히 배제하고, AI 에이전트 환경에서 **프롬프트 비대화(Prompt Bloat)와 비결정적 환각을 방지하며, 컴팩션(Compaction) 시에도 규칙을 100% 보존하는 컨텍스트 관리 및 점진적 로딩 아키텍처의 순수 뼈대**를 정의한다.

---

## 1. 해결하고자 하는 핵심 문제 (Problem Definition)

대규모 룰셋, 도메인 가이드라인, 코딩 지침을 에이전트에게 적용할 때 다음 4가지 실패가 발생한다:

1. **프롬프트 비대화 (Prompt Bloat)**: 수십 개의 룰 파일과 수천 줄의 분류표를 첫 턴부터 시스템 프롬프트에 전부 넣으면 컨텍스트 윈도우 비용과 지연 시간이 급증함.
2. **주의력 분산 (Attention Dilution / Lost in the Middle)**: 프롬프트가 과도하게 길어지면 LLM이 사용자의 현재 지시사항과 핵심 불변식을 망각함.
3. **비결정적 로딩 (Non-deterministic Loading)**: 파일 도구(`view_file`) 호출을 순수 LLM의 자율에만 맡기면, 필요한 규칙을 읽지 않고 기억(추측)에 의존해 환각을 일으킴.
4. **컴팩션 유실 및 망각 (Compaction Loss & Amnesia)**: 대화가 길어져 컨텍스트 요약(Compaction)이 발생할 때 룰과 지침이 함께 요약·삭제되어, 이후 턴에서 과거 지침을 추측으로 실행함.

---

## 2. 5계층 아키텍처 (The 5-Tier Architecture)

이 아키텍처는 **"상시 주입 불변식 -> 스코프 라우터 -> 점진적 로딩 -> 결정론적 하네스 검증 -> 피드백 캘리브레이션"**의 5단계 파이프라인으로 동작한다.

```mermaid
flowchart TD
    subgraph Tier0["Tier 0: 상시 주입 계층 (Eager Invariants)"]
        G["글로벌 불변식 & 런타임 규칙"]
        W["워크스페이스 오버레이 (AGENTS.md)"]
    end

    subgraph Tier1["Tier 1: 선행 게이트 & 스코프 라우터 (Scope Router)"]
        S["스킬 메타데이터 (SKILL.md Frontmatter)"]
        TF["범용 작업 프레임 (Universal Task Frame)"]
    end

    subgraph Tier2["Tier 2: 점진적 지연 로딩 계층 (Lazy Loading via Tool)"]
        REF["세부 참조/스펙 문서 (view_file로 온디맨드 로드)"]
    end

    subgraph Tier3["Tier 3: 결정론적 하네스 계층 (Deterministic Harness)"]
        H["프로그램 기반 검증기 (CLI Scripts / Linter / Tests)"]
    end

    subgraph Tier4["Tier 4: 피드백 캘리브레이션 계층 (Calibration & Promotion)"]
        LEDGER["피드백 원장 (feedback_ledger.md)"]
        PROMOTE["5단계 일반화 검증 -> Tier 0 / Tier 2 승격"]
    end

    Tier0 -->|런타임 매 턴 강제 주입| Tier1
    Tier1 -->|작업 매칭 시 view_file 온디맨드 호출| Tier2
    Tier2 --> LLM["LLM 추론 및 초안 생성"]
    LLM --> Tier3
    Tier3 -->|Exit Code 0 (PASS)| Output["최종 결과물 확정"]
    Tier3 -->|Exit Code != 0 (FAIL) + 진단 로그| LLM
    Output -.->|사용자 피드백 발생 시| Tier4
    LEDGER --> PROMOTE
    PROMOTE -.->|글로벌 룰 승격| Tier0
    PROMOTE -.->|도메인 스펙 승격| Tier2
```

---

## 3. 계층별 상세 명세

### [Tier 0] 상시 주입 계층 (Eager System Invariants)
* **역할**: 대화 컴팩션(요약)과 무관하게 항상 유지되어야 하는 최상위 불변식.
* **주입 방식**: 런타임 전처리기(Runtime Preprocessor)가 시스템 프롬프트 최상단에 강제 주입 (LLM의 호출 여부와 무관하게 100% 결정적 로드).
* **내용**:
  * **사실 우선 원칙 (Fact > Speculation)**: 공식 문서, 코드, 테스트 결과가 추측보다 항상 우선.
  * **가설 검증 및 실패 복구 루프 (Reasoning & Reality Loop)**.
  * **컴팩션 복원력 (Compaction Resilience)**: 긴 대화로 이전 턴의 세부 규칙이 요약·유실되었을 때 추측하지 않고 도구로 재로드하는 프로토콜.
  * **안전성 및 파괴적 작업 차단 경계**.

### [Tier 1] 선행 게이트 & 스코프 라우터 (Scope Routing & Task Framing)
* **역할**: 작업의 유형과 목적을 판정하여 불필요한 규칙이 활성화되는 것을 방지.
* **Universal Task Frame (범용 작업 프레임 5대 요소)**:
  1. `Task Type`: 신규 생성(Create) | 수정/리팩토링(Modify) | 검수/진단(Audit) | 캘리브레이션(Calibrate)
  2. `Target / Consumer`: 결과물을 소비하는 주체 (컴파일러, 런타임, API 클라이언트, 동료 개발자, 채용관 등)
  3. `Contract & Schema`: 준수해야 하는 형식 (JSON Schema, Markdown Spec, API 인터페이스, 문법 규칙)
  4. `Evidence Boundary`: 추측이 금지된 인정된 사실의 범위 (공식 문서, 소스 코드, 로그, 명시적 요구사항)
  5. `Non-goals & Invariants`: 시스템을 망가뜨리거나 범위를 벗어나는 절대 금지 영역
* **스킬 메타데이터**: 2~3줄의 얇은 설명(Name, Description)만 상시 노출하여 프롬프트 토큰 절약.

### [Tier 2] 점진적 공개 계층 (Progressive Disclosure / Lazy Loading)
* **역할**: 대용량 참조 문서, 상세 룰셋, API 명세, 어휘 사전을 **필요한 순간에만** 워킹 메모리에 로드.
* **온디맨드 도구 호출 프로토콜 (On-demand Tool-calling Protocol)**:
  > *"특정 스킬이나 도메인 작업에 착수할 때, 에이전트는 지침에 명시된 참조 파일(`references/deep-rule-ssot.md`)을 `view_file` 도구로 완전히 읽은 후에만 출력을 생성해야 한다."*
* **효과**: 시스템 프롬프트 오염을 막으면서도, 모델이 자의적 기억 대신 최신 파일 스펙을 직접 읽고 정확하게 추론함.

### [Tier 3] 결정론적 하네스 계층 (Deterministic Verification)
* **역할**: LLM의 주관적 완료 선언("완료되었습니다")을 불신하고, **순수 프로그램 코드로 결과물을 검증**.
* **하네스 유형**:
  * **패턴/정적 검증기**: 금지된 추상 표현, 미완성 플레이스홀더(`TODO`, `FIXME`) 정규식 검사
  * **제약/예산 검증기**: 글자 수, 라인 수, 토큰 상하한 엄격 검사
  * **구조/스키마 검증기**: JSON Schema, Markdown 헤더 구조, AST 문법 검증
* **종결 기준 및 피드백**:
  * 스크립트의 리턴 코드가 `Exit Code 0`일 때만 작업을 종료.
  * 실패 시 단순 중단이 아닌 **위반 라인 번호와 수정 가이드(Actionable Diagnostic Logs)**를 표준 에러로 출력하여 LLM의 Self-Correction 유도.

### [Tier 4] 피드백 캘리브레이션 계층 (Calibration & Promotion Loop)
* **역할**: 사용자 피드백이나 하네스 실패가 발생했을 때 단순 1회성 땜질에 머무르지 않고, **판단 의사결정 프로세스(Decision Layer)를 일반화하여 적절한 계층으로 승격(Promote)**.
* **5단계 일반화 검증 프로세스**:
  1. `Representative Case`: 최초에 실패했던 사례가 해결되는가?
  2. `Near Transfer Case`: 같은 도메인의 유사한 작업에서도 동일하게 유효한가?
  3. `Far Transfer Case`: 다른 도메인/장르에 적용해도 부작용이 없는가?
  4. `Clean Control`: 원래 정상이었던 대조군을 훼손하지 않는가?
  5. `Adversarial Control`: 과적합(과도한 억제/검열)이 발생하지 않는가?
* **승격 규칙 (Promotion Policy)**:
  * **Tier 0 승격**: 도메인 무관 전역 헌법 가치 (예: 사실 우선, 추측 금지, 하네스 필수, 컴팩션 복원)
  * **Tier 2 승격**: 특정 도메인/스킬 내의 세부 설계 규칙, 용어집, 스키마 제약
  * **Tier 3 승격**: 스크립트/AST/정규식으로 기계적 검증이 가능한 불변식

---

## 4. 이식 가능한 최소 파일 구조 (Portable Template File Tree)

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
│   └── verify_harness.py          <-- [Tier 3] 규격/제약/불변식 결정론적 검증기 (Stdin/File 지원)
└── calibration/
    └── feedback_ledger.md         <-- [Tier 4] 피드백 분석 및 룰 승격 원장
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
- 방대한 세부 참조 문서는 상시 시스템 프롬프트에 올리지 않는다.
- 작업 대상 도메인이나 스킬이 확정되면, 스킬 지침에 명시된 참조 파일을 도구(`view_file`)로 먼저 읽은 후 실행한다.

## 3. 컴팩션 복원력 (Compaction Resilience)
- 긴 대화로 인해 이전 턴에서 로드했던 세부 스펙이나 룰이 요약·유실되었다고 판단되면, 기억에 의존해 추측하지 말고 도구(`view_file`)로 참조 문서를 재로드한다.

## 4. 결정론적 검증 게이트 (Deterministic Verification Gate)
- AI 자신의 주관적 판단만으로 작업을 완료로 선언하지 않는다.
- 지정된 검증 스크립트(`scripts/verify_harness.py`)를 실행하여 성공(Exit Code 0)할 때만 최종 완료한다.
- 실패 시 하네스가 출력한 진단 로그(위반 라인 및 원인)를 바탕으로 즉시 자체 교정(Self-Correction)한다.
```

### (2) `SKILL.md` (Tier 1: 얇은 진입 프로토콜 템플릿)

```markdown
---
name: domain-processor
description: Process domain-specific tasks according to strict quality, architecture, and schema constraints.
---

# Domain Processor Protocol

When this skill is activated:
1. **Mandatory Load**: You MUST read the detailed specification at `skills/domain-skill/references/deep-rule-ssot.md` using `view_file` BEFORE generating any output or draft.
2. **Formulate Task Frame**: Establish the 5-element Universal Task Frame (Task Type, Target/Consumer, Contract/Schema, Evidence Boundary, Non-goals).
3. **Execute & Draft**: Generate the solution strictly bounded by the admitted facts and loaded specification.
4. **Deterministic Gate**: Run `python3 scripts/verify_harness.py <target_file>` (or stream via stdin) and ensure Exit Code 0.
```

### (3) `scripts/verify_harness.py` (Tier 3: 결정론적 검증기 템플릿)

```python
#!/usr/bin/env python3
"""
Universal Deterministic Quality & Constraint Harness
Returns Exit Code 0 on success, 1 on failure with actionable diagnostic logs.
Supports both file path arguments and standard input streaming.
"""
import sys
import re
from pathlib import Path

def verify_content(content: str, source_name: str = "stream") -> bool:
    errors = []
    lines = content.splitlines()

    # 1. 길이 및 분량 제약 검증
    char_count = len(content)
    if char_count < 100:
        errors.append(f"[LENGTH_ERROR] Content too short ({char_count} chars < minimum 100 chars).")

    # 2. 범용 금지 패턴 및 미해결 플레이스홀더 검증
    banned_patterns = [
        (r"TODO:", "Unresolved TODO placeholder detected"),
        (r"FIXME:", "Unresolved FIXME placeholder detected"),
        (r"<INSERT_.*?_HERE>", "Template placeholder was not replaced"),
    ]
    for line_no, line in enumerate(lines, 1):
        for pattern, reason in banned_patterns:
            if re.search(pattern, line, re.IGNORECASE):
                errors.append(f"[PATTERN_VIOLATION] Line {line_no}: {reason} -> '{line.strip()}'")

    # 3. 구조적 무결성 검증 (헤더 등 필수 마크다운/스키마 구조)
    if not any(line.strip().startswith("#") for line in lines):
        errors.append("[STRUCTURE_ERROR] Missing structured Markdown headers (must start with #).")

    # 결과 판정 및 피드백 출력
    if errors:
        print(f"❌ HARNESS VERIFICATION FAILED for '{source_name}':", file=sys.stderr)
        for err in errors:
            print(f"  - {err}", file=sys.stderr)
        print("\nAction Required: Fix the violations above and re-run verification before finalizing.", file=sys.stderr)
        return False

    print(f"✅ HARNESS PASSED: '{source_name}' ({char_count} chars, {len(lines)} lines) satisfies all invariants.")
    return True

def main():
    if len(sys.argv) > 1:
        target_path = Path(sys.argv[1])
        if not target_path.exists():
            print(f"❌ ERROR: Target file '{target_path}' not found.", file=sys.stderr)
            sys.exit(1)
        content = target_path.read_text(encoding="utf-8")
        success = verify_content(content, str(target_path))
    else:
        if sys.stdin.isatty():
            print("Usage: python3 verify_harness.py <file_path> OR cat <file> | python3 verify_harness.py", file=sys.stderr)
            sys.exit(1)
        content = sys.stdin.read()
        success = verify_content(content, "STDIN")

    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
```

### (4) `calibration/feedback_ledger.md` (Tier 4: 피드백 원장 및 승격 정책 템플릿)

```markdown
# Feedback Calibration Ledger & Promotion Policy

## 1. Feedback Entry Schema

- **Date**: YYYY-MM-DD
- **Trigger**: [User Correction | Harness Failure | Regression]
- **Episode**:
  - Rejected: `[실패했던 출력/코드/표현]`
  - Corrected: `[사용자가 수정한 올바른 결과]`
- **Root Cause Layer**: [System Boundary | Spec Misunderstanding | Slot Mismatch | False Assumption]
- **General Principle**: 단순 키워드 밴이 아닌, 의사결정 계층에서 도출된 추상화된 판단 규칙
- **Promotion Target**: [Tier 0 (Global AGENTS.md) | Tier 2 (Skill SSOT) | Tier 3 (Harness Code)]

---

## 2. Promotion Policy (규칙 승격 기준)

1. **Tier 0 (AGENTS.md) 승격**:
   - 도메인과 무관하게 모든 에이전트 작업에 공통 적용되는 헌법적 가치 (예: 사실 우선, 추측 금지, 하네스 필수, 컴팩션 복원)
2. **Tier 2 (references/*.md) 승격**:
   - 특정 도메인/스킬 내에서만 유효한 상세 아키텍처 규칙, 용어집, 스키마 제약
3. **Tier 3 (scripts/verify_*.py) 승격**:
   - 정규식, AST 분석, 린터, 단위 테스트 등으로 기계적 검증이 가능한 불변식
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
2. 최상위 `AGENTS.md`에 **핵심 불변식**을 배치합니다.
3. 세부 지침 문서를 `references/deep-rule-ssot.md`에 배치하고, `SKILL.md`에 **선행 읽기 강제 조항**을 작성합니다.
4. 에이전트에게 작업을 요청하면:
   - 프롬프트 낭비 없이 스킬 메타데이터만 인식하고 있다가,
   - 작업 시작 시 `view_file`로 세부 문서를 정확히 로드한 뒤,
   - 결과물 작성 후 `verify_harness.py`를 실행하여 통과 여부를 검증하고, 실패 시 진단 로그를 바탕으로 자체 교정하는 완전한 라이프사이클이 작동합니다.
