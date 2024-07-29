# git restore

- [git restore](#git-restore)
    - [`restore`?](#restore)
    - [작업 디렉토리와 스테이징 영역에 복원](#작업-디렉토리와-스테이징-영역에-복원)

## `restore`?

- 워킹 디렉토리와 인덱스(스테이지)의 상태를 조작

## 작업 디렉토리와 스테이징 영역에 복원

```bash
# Drop commit 시에 실행되는 IntelliJ git tool box 커맨드
# go.work.sum 파일을 HEAD의 상태로 작업 디렉토리와 스테이징 영역에 복원
git \
    # 경로명을 ASCII가 아닌 문자로 표시할 때의 문제를 방지
    -c core.quotepath=false \
    # 커밋 로그에 서명 정보를 표시하지 않도록 설정
    -c log.showSignature=false \
    # Git 2.23 이상에서 도입된 명령어로, 파일을 특정 상태로 복원
    restore \
    # 인덱스(스테이지) 상태를 복원
    --staged \
    # 작업 트리의 파일을 복원
    --worktree \
    # 복원할 기준을 현재 HEAD로 지정
    --source=HEAD \
    -- go.work.sum
```
