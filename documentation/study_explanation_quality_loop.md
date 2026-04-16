# Study Explanation Quality Loop

## 개요

이 문서는 외부 skill 경로 `/Users/rody/.codex/skills/study-explanation` 에 있는 `study-explanation`을 앞으로 어떻게 강화하고 검증할지 정리한 repo-side 추적 문서입니다.

핵심 목표는 단순합니다.
`study-explanation`이 현재 exemplar를 흉내 내는 데서 멈추지 않고, 현재 exemplar 수준 이상으로 설명을 만들고, 더 강한 새 문서를 다시 benchmark peer나 future exemplar 후보로 삼을 수 있게 만드는 것입니다.

즉 이 문서는 "어떻게 잘 쓰게 할 것인가"뿐 아니라 "무엇이 실패 출력이고, 무엇이 stronger-than-exemplar이며, 무엇을 future exemplar로 승격할 수 있는가"를 함께 다룹니다.

## 1. 세 층으로 본 skill 구조

현재 `study-explanation`은 세 층으로 보는 편이 가장 정확합니다.

1. 생성 계약

    문서를 어떤 순서로 열고, 어떤 artifact를 언제 보여 주고, 어디서부터 execution path로 내려갈지를 정합니다.
    이 층의 중심 reference는 다음 두 파일입니다.

    - [/Users/rody/.codex/skills/study-explanation/SKILL.md](/Users/rody/.codex/skills/study-explanation/SKILL.md)
    - [/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md](/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md)

2. 평가 계약

    draft가 실제로 어디까지 닫혔는지 판정합니다.
    여기서는 task completion state가 아니라 explanation quality grade를 봅니다.
    중심 reference는 다음 파일입니다.

    - [/Users/rody/.codex/skills/study-explanation/references/output-quality-gates.md](/Users/rody/.codex/skills/study-explanation/references/output-quality-gates.md)

3. benchmark / promotion 계약

    skill 개정이 정말로 stronger result를 만들었는지, 그리고 새 문서를 future exemplar 후보로 삼을 수 있는지를 정합니다.
    중심 reference는 다음 파일들입니다.

    - [/Users/rody/.codex/skills/study-explanation/references/benchmark-task-bank.md](/Users/rody/.codex/skills/study-explanation/references/benchmark-task-bank.md)
    - [/Users/rody/.codex/skills/study-explanation/references/exemplar-promotion-loop.md](/Users/rody/.codex/skills/study-explanation/references/exemplar-promotion-loop.md)

이렇게 나누어 두면, "잘 쓰는 법"만 강화되고 "무엇이 실패인지"는 흐려지는 문제를 줄일 수 있습니다.

## 2. Output Quality Grade

여기서 말하는 grade는 task completion state와 다릅니다.
전역 `AGENTS.md`의 `COMPLETE / PARTIAL / BLOCKED`는 작업 상태이고, 여기의 등급은 설명 산출물의 품질입니다.

현재는 아래 순서로 보는 편이 좋습니다.

- `PARTIAL`
  - 방향은 맞지만 base gate 또는 route-specific gate가 열려 있다.
- `COMPLETE`
  - 현재 질문을 미래 독자가 다시 복원하고 replay 할 수 있을 정도로 닫혀 있다.
- `PERFECT`
  - `COMPLETE`를 넘어서 opening, artifact timing, lower-layer closure, contrast, verification이 특히 강하다.
- `OVERACHIEVING`
  - current benchmark peer와 비교했을 때 structure 전체 차원에서 더 강한 문서다.
- `EXEMPLAR_CANDIDATE`
  - `OVERACHIEVING`이며 promotion loop를 돌릴 가치가 있다.

중요한 점은 이 등급이 감상 비평이 아니라 gate 기반 판정이어야 한다는 것입니다.

## 3. Benchmark Task Bank

skill을 실제로 개정했다면 적어도 일부 benchmark task를 다시 보는 편이 좋습니다.
현재 bank는 아래 route를 커버합니다.

- `dynamic_programming`
  - concept-first 알고리즘 문서
- `optimal_solution`
  - baseline에서 더 강한 방법으로 일반화하는 문서
- `git rebase`
  - tool / operational 문서
- `java synchronized`
  - runtime / language / hardware를 함께 연결하는 문서
- `entry.S`
  - code-first low-level entry path 문서
- raw material promotion
  - `.tmp/interview*.md` 또는 `interviews/*` cluster를 정식 문서로 승격하는 문서

핵심은 같은 route만 반복해 만족하지 않는 것입니다.
예를 들어 code-first topic만 계속 좋아졌다고 해서 skill 전체가 stronger해졌다고 말할 수는 없습니다.

## 4. Promotion 원칙

새 문서를 exemplar로 다룰 때는 아래를 지키는 편이 좋습니다.

- 더 길다고 자동 승격하지 않는다
- 더 최근에 썼다고 자동 승격하지 않는다
- 피드백을 많이 반영했다고 자동 승격하지 않는다
- current benchmark peer와 같은 추상화 수준에서 비교한다
- 기존 exemplar의 강점도 같이 적는다
- replacement보다 additive promotion을 먼저 고려한다

즉 "이 문서가 더 좋다"는 감상보다, "어떤 route에서 어떤 축이 더 강해졌는가"를 남기는 것이 더 중요합니다.

## 5. 이 루프가 필요한 이유

현재 exemplar 문서들은 매우 유용하지만, 완전히 균일한 품질 집합은 아닙니다.
어떤 문서는 opening이 강하고, 어떤 문서는 baseline-to-generalization이 강하고, 어떤 문서는 code-first explanation이 강합니다.

그래서 `study-explanation`이 해야 하는 일은 exemplar 전체를 평면적으로 복제하는 것이 아닙니다.
각 exemplar의 strongest trait를 추출하고, 현재 주제에 맞게 재구성하고, 더 강한 새 문서가 나오면 그것을 다시 benchmark set에 편입할 수 있게 만드는 것입니다.

이 루프가 없으면 skill은 계속 "좋은 문장 규칙 묶음"으로 남고, 시간이 갈수록 더 강해지는 시스템으로 발전하기 어렵습니다.

## 6. 운영 메모

skill 수정 후 확인 경로는 보통 아래 순서가 좋습니다.

1. external skill diff review
2. `agents/openai.yaml` regeneration
3. skill validator 실행
4. benchmark task bank 기준 self-audit
5. 필요한 경우 promotion 후보 비교

현재 validator 경로:

- [/Users/rody/.codex/skills/.system/skill-creator/scripts/generate_openai_yaml.py](/Users/rody/.codex/skills/.system/skill-creator/scripts/generate_openai_yaml.py)
- [/Users/rody/.codex/skills/.system/skill-creator/scripts/quick_validate.py](/Users/rody/.codex/skills/.system/skill-creator/scripts/quick_validate.py)

이 문서는 external skill path를 대신하는 SSOT는 아닙니다.
다만 앞으로 `study` 저장소 안에서 skill 강화 이력을 다시 따라갈 때, 무엇을 왜 바꿨는지 빠르게 복원할 수 있게 하는 보조 추적 문서로 둡니다.
