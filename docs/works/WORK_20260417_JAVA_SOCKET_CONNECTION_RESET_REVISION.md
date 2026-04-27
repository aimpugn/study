# WORK_20260417_JAVA_SOCKET_CONNECTION_RESET_REVISION

## 0. Meta

- 작업 제목: 사용자 피드백 기준으로 `java_socket_connection_reset.md`와 `study-explanation` 규칙 수정
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260417_JAVA_SOCKET_CONNECTION_RESET_REVISION.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - Markdown 문법 TOC, 4-space indent, 더 자연스러운 문장, exemplar 수준의 밀도와 예제 반영
- 원문 사용자 요청:
  - TOC가 Markdown 문법이 아니고, indent가 4칸이 아니며, 문장이 어색하고, exemplar와 비슷해 보이지 않는다고 지적
- 대상 경로 / 자산:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - `/Users/rody/.codex/skills/study-explanation/references/output-quality-gates.md`
  - `/Users/rody/VscodeProjects/study/jvm/java/java_socket_connection_reset.md`
- 실행자: Codex
- 시작 일시: `2026-04-17 00:33:54 +0900`
- 종료 일시: `2026-04-17 00:37:42 +0900`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 사용자 피드백을 one-off 수정으로 끝내지 않고, reusable skill rule과 실제 문서 개정으로 함께 반영한다.
- must_keep:
  - 문서는 실제 Markdown TOC를 사용한다
  - 번호 목록 아래 prose는 4칸 indent를 지킨다
  - 문장은 bullet + 후행 해설이 아니라, 가능한 한 바로 이해되는 직접 문장으로 쓴다
  - troubleshooting 문서라도 exemplar 수준의 설명 밀도와 예제를 유지한다

## 2. Root-First Framing

- 근본 문제:
  - 이전 산출물은 troubleshooting route를 추가했지만, 형식 정확성, prose 직진성, 예제 밀도에서 exemplar 기준을 제대로 재현하지 못했다.
- 왜 지금 중요한가:
  - 이런 종류의 실패는 문서 하나의 문제가 아니라 skill이 여전히 exemplar 수준을 안정적으로 못 지키고 있다는 증거다.
- 작업 목표:
  - 사용자 피드백을 skill 규칙과 산출물 둘 다에 반영해 같은 실패가 재발하지 않게 한다.

## 3. Frozen Checklist

- [x] C-01 문서 TOC가 실제 Markdown 링크 목록으로 바뀐다
- [x] C-02 번호 목록 설명 indent가 4칸으로 정리된다
- [x] C-03 어색한 bullet + 후행 해설 문장을 자연스러운 직접 문장으로 고친다
- [x] C-04 문서가 summary-like troubleshooting memo가 아니라 exemplar 수준의 밀도와 예제를 갖도록 강화된다
- [x] C-05 이 피드백이 skill rule로 승격된다
- [x] C-06 skill validator PASS
- [x] C-07 final review + commit

## 4. Verification Plan

- target file re-read against user feedback bullets
- `rg`로 TOC / indent / scaffolding / awkward sentence remnants 점검
- skill metadata regeneration + validator
- commit only this revision round assets

## 5. Final Audit

- document side:
  - prose `목차:`를 제거하고 실제 Markdown TOC로 교체
  - numbered list prose를 4칸 indent로 정리
  - `PASS 신호:`처럼 문장을 끊는 패턴을 자연스러운 직접 문장으로 교체
  - stale keep-alive worked example, 계층별 예외 전파, 실제 확인 순서, replay 실험을 추가해 밀도를 높임
- skill side:
  - Markdown TOC를 prose 대체물로 바꾸지 말라는 rule 추가
  - numbered list prose 4-space indent rule 추가
  - troubleshooting 문서는 worked example 없이 얕은 memo로 끝내지 말라는 gate 추가
  - validator 결과: `Skill is valid!`
- 남은 리스크:
  - 이 문서가 실제로 route-local exemplar candidate인지 여부는 다음 benchmark comparison 라운드에서 별도로 판정해야 한다
