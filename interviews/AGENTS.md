# AGENTS.md - interviews nested learning overlay

> 적용 범위: `study/interviews/`
> 상위 계약: `study/AGENTS.md`
> 프로젝트 사실 문서: [`PROJECT_INTENT.md`](PROJECT_INTENT.md), [`USECASE.md`](USECASE.md)
> WORK 템플릿: 상위 `study/AGENTS_WORK_TEMPLATE.md`

이 하위 경로는 경력 기술 인터뷰 준비를 위한 학습 작업 공간입니다.
상위 `study` 계약은 그대로 적용하되, `interviews` 안의 작업은 "질문을 많이 모으는 것"보다 질문을 받았을 때 짧게 답하고, 꼬리 질문에서 깊게 내려가며, 작은 기술 단위를 큰 시스템 판단으로 조립하는 능력을 만드는 쪽을 우선합니다.

## 1. 적용 순서

1. 현재 런타임에서 보이는 전역 `AGENTS.md`와 `AGENTS_WORK_TEMPLATE.md`
2. `study/AGENTS.md`와 `study/AGENTS_WORK_TEMPLATE.md`
3. 이 파일
4. 이 경로의 `PROJECT_INTENT.md`와 `USECASE.md`
5. 현재 작업이 직접 읽어야 하는 `source/`, root curriculum, topic 문서, 검증 스크립트

전역 오버레이는 존재하면 반드시 따르지만, 이 경로의 문서도 혼자 읽었을 때 작업 목적과 품질 기준이 닫혀야 합니다.
사용자가 자연어로 "정리", "다시 설명", "면접 준비용으로 만들어 달라"고 말하면, 표면 문장을 실행 상한으로 보지 말고 질문 의도, 면접 사용 장면, 기대 깊이, benchmark floor, 위험한 얕은 답변을 먼저 판정합니다.

## 2. 산출물 기준

정식 인터뷰 답변 자산은 보통 두 층을 함께 가져야 합니다.

- 먼저 면접장에서 바로 말할 수 있는 짧은 직답
- 그 직답을 지탱하는 메커니즘, 예시, 실패 모드, 꼬리 질문, 검증 경로

다만 모든 문서에 같은 표면 구조를 강제하지 않습니다.
큰 커리큘럼, source material 정리, database/network/runtime deep dive처럼 사용자가 깊은 재구성 문서를 원한 경우에는 30초 답변 양식을 억지로 끼워 넣지 말고, 복사해 다른 곳에 붙여도 문맥이 살아 있는 bridge section, trace, 비교표, ASCII 흐름, 검증 관점으로 깊이를 닫습니다.

## 3. Source Reservoir와 승격

`source/` 아래의 원재료는 버릴 자료가 아니라 승격 대기 자료입니다.
하지만 원재료 자체가 품질 기준선은 아닙니다.
정식 자산으로 승격하려면 질문 cluster가 좁혀지고, 핵심 기술 단위가 드러나며, 짧게 말하기와 깊게 설명하기가 함께 가능해야 합니다.

승격 또는 재작성 작업에서는 최소한 아래를 확인합니다.

- 어떤 질문 또는 판단 축을 보호하는가
- 어떤 작은 기술 단위가 중심인가
- 어떤 계층까지 내려가야 꼬리 질문에 견디는가
- 예시, trace, 반례, 실패 모드, 검증 경로가 필요한가
- 기존 root curriculum과 중복되거나 서로 다른 말을 하지 않는가

## 4. 품질 하한과 검증

구조 validator나 heading 수는 필요조건일 수 있지만 설명 품질의 충분조건이 아닙니다.
예를 들어 database deep dive류 작업에서는 `validate_interview_database_deep_dive.py` 같은 구조 검증이 통과하더라도, 내용이 `mvcc.md`급 benchmark floor보다 얕거나 사람이 다시 설명할 수 없으면 완료가 아닙니다.

완료 전에 다음 질문에 답할 수 있어야 합니다.

- 처음 보는 개발자가 핵심 결론을 짧게 말할 수 있는가
- 꼬리 질문을 받으면 메커니즘으로 내려갈 수 있는가
- 실패 모드와 반례를 설명할 수 있는가
- 작은 기술 단위를 시스템 설계 판단으로 다시 조립할 수 있는가
- 구조 검증과 실제 설명 품질 증거를 구분했는가

