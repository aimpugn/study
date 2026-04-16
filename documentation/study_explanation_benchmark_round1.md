# Study Explanation Benchmark Round 1

## 목적

이 문서는 `study-explanation`을 실제 저장소 문서 batch에 대입해 보면서, 어떤 실패 패턴이 반복되는지 확인하고 그 결과를 external skill 규칙으로 되먹인 1차 benchmark 기록입니다.

이번 라운드의 핵심은 "느낌상 더 좋아 보인다"가 아니라, 실제 corpus에서 같은 문제가 반복되면 그것을 skill rule로 승격해야 한다는 점을 검증하는 것이었습니다.

## 본 benchmark에서 본 문서

- [Algorithms Time Complexity](/Users/rody/VscodeProjects/study/algorithms/algorithms_time_complexity.md)
- [nohup](/Users/rody/VscodeProjects/study/linux/commands/nohup.md)
- [Java Socket](/Users/rody/VscodeProjects/study/jvm/java/java_socket.md)
- [Interviews2](/Users/rody/VscodeProjects/study/.tmp/interviews2.md)
- [Go troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/go.md)
- [Terminal troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/terminal.md)
- [Rust troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/rust.md)

## 반복해서 드러난 문제

### 1. 채팅 답변의 잔여 문장이 문서 본문으로 남는다

[Algorithms Time Complexity](/Users/rody/VscodeProjects/study/algorithms/algorithms_time_complexity.md)와 [Java Socket](/Users/rody/VscodeProjects/study/jvm/java/java_socket.md)는 opening에 `좋습니다`, `설명드릴게요`, `완료되면 알려드리겠습니다` 같은 assistant-presence scaffolding가 남아 있었습니다.

이 패턴은 단순히 문체가 어색한 문제가 아닙니다. opening이 topic-first direct statement여야 한다는 계약을 약하게 만들고, 문서의 실제 질문과 설명자의 진행 멘트를 섞어 버립니다. 그래서 이번 라운드에서는 이 패턴을 explicit fail pattern으로 승격했습니다.

### 2. troubleshooting 문서를 위한 독립 route가 필요하다

[Go troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/go.md), [Terminal troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/terminal.md), [Rust troubleshooting](/Users/rody/VscodeProjects/study/troubleshooting/rust.md), 그리고 [Java Socket](/Users/rody/VscodeProjects/study/jvm/java/java_socket.md)는 공통적으로 `증상 -> 원인 -> 해결` 구조를 가집니다.

하지만 기존 `study-explanation`에는 `code-first`, `concept-first`, `tool/operational`, `raw-material conversion`만 있었고 `troubleshooting` route가 없었습니다. 그래서 문제 해결형 문서가 broad concept note와 fix memo 사이에서 흔들릴 여지가 있었습니다.

이번 라운드에서는 troubleshooting route를 generation skeleton, route-specific quality gates, benchmark task bank에 모두 추가했습니다. 이제 이 route에서는 `증상`, `직접 원인`, `근본 원인`, `수정이 통하는 이유`, `replay path`, `재발 방지`를 분리해서 보게 됩니다.

### 3. 제목, opening, 실제 범위가 어긋날 수 있다

[Algorithms Time Complexity](/Users/rody/VscodeProjects/study/algorithms/algorithms_time_complexity.md)는 broad title을 달고 있지만, opening과 본문은 실제로는 `150만 건 정렬을 1GHz CPU에서 어떻게 추정할 것인가`라는 훨씬 좁은 질문을 다룹니다.

이 문제는 opening clarity와는 다릅니다. opening이 빨리 나온다고 해서 title/opening/body 범위가 서로 맞는 것은 아닙니다. 그래서 이번 라운드에서는 `title/opening/scope alignment`를 base gate로 추가했습니다.

## 긍정 신호도 있었다

[nohup](/Users/rody/VscodeProjects/study/linux/commands/nohup.md)은 tool/operational route의 positive peer였습니다. `nohup`이 무엇을 해결하는지, 왜 `&`만으로는 충분하지 않은지, 세션과 `SIGHUP` 메커니즘이 어떻게 이어지는지를 꽤 잘 설명합니다.

[Interviews2](/Users/rody/VscodeProjects/study/.tmp/interviews2.md)는 raw-material route가 계속 필요하다는 근거를 보여 줬습니다. 질문 cluster가 매우 크고 이질적이어서, 이 파일을 그대로 다듬는 것보다 cluster를 뽑아 정식 문서로 승격하는 접근이 더 안전합니다.

## 이번 라운드에서 바뀐 것

external skill 변경은 저장소 밖에서 일어났지만, 내용상으로는 다음이 추가되었습니다.

- `troubleshooting` route를 generation contract에 추가
- troubleshooting route용 route-specific quality gates 추가
- `assistant-presence scaffolding`를 explicit fail pattern으로 추가
- `title/opening/scope alignment`를 base gate로 추가
- troubleshooting benchmark task 추가
- canonical exemplar가 희박한 route에서는 weak peer를 ceiling으로 삼지 말고 gates와 benchmark를 더 강하게 보라는 guidance 추가

관련 WORK ledger는 [WORK_20260416_STUDY_EXPLANATION_BENCHMARK_ROUND1.md](/Users/rody/VscodeProjects/study/WORK_20260416_STUDY_EXPLANATION_BENCHMARK_ROUND1.md)입니다.

## 현재 판단

이번 라운드의 결과는 `benchmark-driven COMPLETE`에 가깝습니다.

의미는 이렇습니다. 이전 라운드에서 만든 quality loop가 실제 corpus failure를 받아 규칙으로 다시 강화되는지 확인했고, 그 루프는 이번에 실제로 작동했습니다. 다만 troubleshooting route의 새 exemplar candidate를 직접 생산해서 route-local benchmark로 승격하는 단계는 아직 남아 있으므로, 아직 `OVERACHIEVING evidence`까지 쌓였다고 말할 단계는 아닙니다.

다음 loop에서는 troubleshooting 문서 하나를 실제로 골라, 개정된 route와 gates를 적용해 rewrite 한 뒤 route-local benchmark candidate 여부까지 판단하는 것이 가장 자연스러운 다음 단계입니다.
