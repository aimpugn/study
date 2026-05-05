# Interview Source Reservoir

이 디렉터리는 인터뷰 준비 문서의 원재료를 보관합니다.
앞으로의 학습과 정리는 `interviews/` 바로 아래의 10개 대주제 문서를 기준으로 진행하고, 이 디렉터리의 파일은 원문 확인과 재생성 검증을 위한 source reservoir로 유지합니다.

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
