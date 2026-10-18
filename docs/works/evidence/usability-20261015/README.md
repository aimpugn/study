# 실험 5 화면 사용성 시험의 원 보고

[트랜잭션 관측 WORK](../../WORK_20261014_TRANSACTION_OBSERVATION.md)의 「실험 5 화면 사용성 시험」 사전 등록과 1~4차 결과가 가리키는 참가자 보고입니다. 참가자는 3년 차 백엔드 개발자 역할을 맡은 AI입니다. 1·2차는 클로드 하위 에이전트(sonnet, haiku)이고, 3·4차는 클로드가 아닌 모델(Gemini 3.8 Flash, GPT-6-Luna)을 각 CLI로 돌렸습니다. 4차는 갈래가 셋입니다: S(3차와 같은 도구, v2 화면), V(스크린샷·좌표 클릭만, v2 화면), Vold(V와 같은 도구, 3차 화면). 파일마다 그 에이전트의 마지막 보고를 그대로 옮겼습니다.

| 파일 | 판 | 참가자 | 단계 | 대상 커밋 |
| --- | --- | --- | --- | --- |
| `r1-sonnet-A.md`, `r1-sonnet-B.md` | 1차 | sonnet | A(자유 사용), B(질문) | `a084feb53` |
| `r1-haiku-A.md`, `r1-haiku-B.md` | 1차 | haiku | A, B | `a084feb53` |
| `r2-sonnet-A.md`, `r2-sonnet-B.md` | 2차 | sonnet(새 인스턴스) | A, B | `aac957767` |
| `r2-haiku-A.md`, `r2-haiku-B.md` | 2차 | haiku(새 인스턴스) | A, B | `aac957767` |
| `r3-gemini-A.md`, `r3-gemini-B.md` | 3차 | gemini-3.8-flash-high(agy) | A, B | `180410b9f` |
| `r3-gpt-A.md`, `r3-gpt-B.md` | 3차 | gpt-6-luna high(codex) | A, B | `180410b9f` |
| `r3-prompt-A.md`, `r3-prompt-B.md` | 3차 | 두 모델에 준 지시문 | A, B | |
| `r4-gemini-<갈래>-A.md`, `r4-gemini-<갈래>-B.md` | 4차 | gemini-3.8-flash-high(agy), 갈래 S·V·Vold | A, B | S·V는 `3b39039af`, Vold는 `180410b9f` |
| `r4-gpt-<갈래>-A.md`, `r4-gpt-<갈래>-B.md` | 4차 | gpt-6-luna high(codex), 갈래 S·V·Vold | A, B | S·V는 `3b39039af`, Vold는 `180410b9f` |
| `r4-prompt-<갈래>-A.md`, `r4-prompt-<갈래>-B.md` | 4차 | 갈래마다 준 지시문 | A, B | |

판정은 WORK의 기대 답 표에 맞춰 메인이 했습니다. 판정을 다시 하려면 B 보고의 C1~C5와 M1~M5를 그 표와 견주면 됩니다.
