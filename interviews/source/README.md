# Interview Source Reservoir

이 디렉터리는 인터뷰 준비 문서의 원재료를 보관합니다.
앞으로의 학습과 정리는 `interviews/` 바로 아래의 10개 대주제 문서를 기준으로 진행하고, 이 디렉터리의 파일은 원문 확인과 재생성 검증을 위한 source reservoir로 유지합니다.

이 디렉터리의 raw Markdown은 문장을 예쁘게 다듬는 대상이 아닙니다. 원문 표현, 중복, 거친 질문 흐름까지 보존해야 나중에 source span, SHA-256, alias를 다시 검증할 수 있습니다. 기술 학습 흐름 개선은 raw 파일을 직접 고치는 대신, root curriculum 문서나 정식 deep dive 문서로 승격할 때 적용합니다.

승격할 때는 먼저 질문을 하나의 판단 축으로 좁힙니다. 그다음 독자가 기억할 정리, 숨은 상태 이동, 비교축, 실패 모드, 검증 anchor를 붙여야 합니다. 예를 들어 메시징 질문은 producer request가 broker의 queue/log/offset/ack 상태로 남는 흐름을, DB 질문은 transaction/log/page/buffer/lock 상태가 어디서 바뀌는지를 보여 주는 식으로 바꿉니다.

## Raw Sources

- `interview_questions.md`
- `interview_questions2.md`
- `interview_questions3.md`
- `interview_questions4.md`
- `interview_s4.md`
- `interviews.md`
- `interviews2.md`

## Generated Source Context

- `_source-context-and-question-bank.md`: 기술 대주제 본문에 바로 넣기보다 질문 은행, 면접 시나리오, 생성 맥락으로 보존해야 하는 원문 chunk입니다.

## Regeneration

아래 명령은 `source/`의 raw 파일을 읽어 `interviews/` 바로 아래의 대주제 문서, `_question-index.md`, `_curriculum_manifest.json`, 그리고 이 디렉터리의 `_source-context-and-question-bank.md`를 다시 생성합니다.

```sh
python3 interviews/tools/build_interview_curriculum.py
```
