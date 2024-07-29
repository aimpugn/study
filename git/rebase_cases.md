# git rebase cases

- [git rebase cases](#git-rebase-cases)
    - [Git Rebase 사례 1: feature 브랜치를 develop 위로 리베이스하기](#git-rebase-사례-1-feature-브랜치를-develop-위로-리베이스하기)
    - [Git Rebase 사례 2: 충돌 해결을 포함한 리베이스](#git-rebase-사례-2-충돌-해결을-포함한-리베이스)
    - [Git Rebase 사례 3: 복잡한 리베이스 시나리오](#git-rebase-사례-3-복잡한-리베이스-시나리오)

이해했습니다. Git 로그와 커밋 해시를 포함한 상세한 rebase 과정을 포함하여 정리하겠습니다.

## Git Rebase 사례 1: feature 브랜치를 develop 위로 리베이스하기

- 단계 1: 현재 상태 확인

    `feature/test1` 브랜치에서 시작합니다.
    현재 상태를 확인하여 `origin/develop`에 반영된 변경사항이 로컬 브랜치에 반영되지 않은 것을 확인합니다.

    ```shell
    git checkout feature/test1
    git log --oneline --graph
    ```

    현재 `feature/test1` 브랜치의 로그는 다음과 같습니다:

    ```shell
    * 3104fef (HEAD -> feature/test1) docs: add git log related files
    *   91cbc8f (origin/main, main, develop) Merge pull request #3
    *   60ac45e Merge pull request #2
    *   8641718 (tag: v0.0.5, tag: v0.0.4) Merge pull request #1
    *   4bb53ba (tag: v0.0.1) 코드 정리 및 추가
    *   ae0e96e 스크립트 태그를 동적으로 추가하는 코드
    ```

- 단계 2: develop 브랜치 업데이트

    `develop` 브랜치로 체크아웃하고 최신 변경사항을 가져옵니다.
    이를 통해 `ba9aab0` 커밋이 로컬 `develop` 브랜치에 반영됩니다.

    ```shell
    git checkout develop
    git pull
    ```

    이제 `develop` 브랜치는 최신 상태가 되었습니다.

    ```shell
    *   ba9aab0 (HEAD -> develop, origin/develop) Merge pull request #4 from snippets/220817
    *   5c15700 docs: add commit result message
    *   0788756 docs: git related commands
    *   af32326 snippet: 빌드 코드 수정
    *   178383b add snippets
    *   91cbc8f (origin/main, main) Merge pull request #3
    ```

- 단계 3: feature 브랜치를 develop 위로 리베이스

    이제 `feature/test1` 브랜치를 `develop` 브랜치 위로 리베이스합니다.

    ```shell
    git checkout feature/test1
    git rebase develop
    ```

    리베이스가 완료된 후 `feature/test1` 브랜치의 로그는 다음과 같습니다:

    ```shell
    * 3104fef (HEAD -> feature/test1) docs: add git log related files
    * ba9aab0 (develop, origin/develop) Merge pull request #4 from snippets/220817
    * 5c15700 docs: add commit result message
    * 0788756 docs: git related commands
    * af32326 snippet: 빌드 코드 수정
    * 178383b add snippets
    * 91cbc8f (origin/main, main) Merge pull request #3
    * 60ac45e Merge pull request #2
    * 8641718 (tag: v0.0.5, tag: v0.0.4) Merge pull request #1
    * 4bb53ba (tag: v0.0.1) 코드 정리 및 추가
    * ae0e96e 스크립트 태그를 동적으로 추가하는 코드
    ```

    *`feature/test1` 브랜치가 `develop` 브랜치의 최신 커밋 위에 재배치*된 것을 확인할 수 있습니다.

- 단계 4: 원격 저장소와 동기화

    리베이스된 브랜치를 강제로 푸시하여 원격 저장소와 동기화합니다.

    ```shell
    git push --force-with-lease
    ```

`feature/test1` 브랜치가 최신 `develop` 브랜치의 변경 사항을 포함하게 되어, 충돌 가능성을 줄이고, 코드의 일관성을 유지합니다.

## Git Rebase 사례 2: 충돌 해결을 포함한 리베이스

- 단계 1: 현재 상태 확인

    `docs/learning-git` 브랜치에서 시작합니다.
    현재 상태를 확인합니다.

    ```shell
    git checkout docs/learning-git
    git log --oneline --graph
    ```

    현재 `docs/learning-git` 브랜치의 로그는 다음과 같습니다:

    ```shell
    * 323c56b (HEAD -> docs/learning-git) docs: rebase 과정 정리 문서
    * 6973e22 (origin/feature/test1, feature/test1) add git commands
    * d02daac Merge remote-tracking branch 'origin/feature/test1' into feature/test1
    * 3104fef docs: add git log related files
    * ba9aab0 (origin/develop, develop) Merge pull request #4 from snippets/220817
    * 5c15700 docs: add commit result message
    * 0788756 docs: git related commands
    * af32326 snippet: 빌드 코드 수정
    * 178383b add snippets
    * 91cbc8f (origin/main, main) Merge pull request #3 from snippets/220729
    * 4ca0c46 snippet: 빌드 코드 수정
    * 60ac45e Merge pull request #2 from snippets/220729
    * 779be86 snippet: shell 스니펫
    * 5cca721 snippet: php 스니펫
    * 05c37aa snippet: form -> json 스니펫
    * 9e7c275 snippet: git 관련 스니펫
    * a20a52f snippet: php7.4 & fpm & nginx 1.22.0 스택의 컨테이너
    * 8641718 (tag: v0.0.5, tag: v0.0.4) Merge pull request #1 from snippets/220728-add-snippets
    * b41ef3e snippets 추가
    * 4bb53ba (tag: v0.0.1) 코드 정리 및 추가
    * ae0e96e 스크립트 태그를 동적으로 추가하는 코드
    ```

- 단계 2: develop 브랜치와 리베이스

    `docs/learning-git` 브랜치를 `develop` 브랜치 위로 리베이스합니다.

    ```shell
    git rebase develop
    ```

    리베이스 도중 충돌이 발생할 수 있습니다. 예를 들어, 다음과 같은 메시지가 나타날 수 있습니다:

    ```shell
    error: could not apply 323c56b... docs: rebase 과정 정리 문서
    Resolve all conflicts manually, mark them as resolved with
    "git add <conflicted_files>" and run "git rebase --continue".
    You can instead skip this commit: run "git rebase --skip".
    To abort and get back to the state before "git rebase", run "git rebase --abort".
    ```

    충돌이 발생하면 `git status` 명령어를 사용하여 충돌 파일을 확인합니다.

    ```shell
    git status
    ```

    충돌 파일을 수정하고 스테이징한 후 리베이스를 계속합니다.

    ```shell
    git add <conflicted_file>
    git rebase --continue
    ```

- 단계 3: 리베이스 완료 후 상태 확인

    리베이스가 완료된 후 `docs/learning-git` 브랜치의 로그는 다음과 같습니다:

    ```shell
    * 4df27c7 (HEAD -> docs/learning-git) docs: rebase 과정 정리 문서
    * ba9aab0 (origin/develop, develop) Merge pull request #4 from snippets/220817
    * 5c15700 docs: add commit result message
    * 0788756 docs: git related commands
    * af32326 snippet: 빌드 코드 수정
    * 178383b add snippets
    * 91cbc8f (origin/main, main) Merge pull request #3 from snippets/220729
    * 4ca0c46 snippet: 빌드 코드 수정
    * 60ac45e Merge pull request #2 from snippets/220729
    * 779be86 snippet: shell 스니펫
    * 5cca721 snippet: php 스니펫
    * 05c37aa snippet: form -> json 스니펫
    * 9e7c275 snippet: git 관련 스니펫
    * a20a52f snippet: php7.4 & fpm & nginx 1.22.0 스택의 컨테이너
    * 8641718 (tag: v0.0.5, tag: v0.0.4) Merge pull request #1 from snippets/220728-add-snippets
    * b41ef3e snippets 추가
    * 4bb53ba (tag: v0.0.1) 코드 정리 및 추가
    * ae0e96e 스크립트 태그를 동적으로 추가하는 코드
    ```

- 단계 4: 원격 저장소와 동기화

    변경된 브랜치를 원격 저장소에 강제로 푸시하여 동기화합니다.

    ```shell
    git push --force-with-lease
    ```

이러한 리베이스 과정을 거치면 다음과 같은 이점이 있습니다:
- 리베이스 과정에서 발생하는 충돌을 해결함으로써 커밋 히스토리를 깔끔하게 유지할 수 있습니다.
- 병합 커밋 없이 깔끔한 커밋 히스토리를 유지할 수 있습니다.
- `docs/learning-git` 브랜치가 최신 `develop` 브랜치의 변경 사항을 포함하게 되어, 충돌 가능성을 줄이고, 코드의 일관성을 유지합니다.

## Git Rebase 사례 3: 복잡한 리베이스 시나리오

1. 초기 상태:

    ```mermaid
    gitGraph
    commit id: "3dd37f8" tag: "main"
    commit id: "459594a"
    commit id: "11a11d6"
    branch feature
    commit id: "ad82e81"
    commit id: "ee064d2"
    commit id: "82946d5"
    commit id: "d72e614"
    commit id: "9c5941d"
    commit id: "0e4db7c"
    commit id: "e221c0c"
    commit id: "3e6d741"
    commit id: "ce7eeb1"
    commit id: "9045770"
    commit id: "54a376e"
    commit id: "867f5ea"
    commit id: "dd693b4"
    commit id: "40175a9"
    branch origin/feat/another
    commit id: "02a68cf"
    ```

    - `main` 브랜치의 최신 커밋은 `11a11d6`입니다.
    - `feature` 브랜치가 `main`에서 분기되어 여러 커밋이 추가되었습니다.
    - `origin/feat/another` 브랜치가 별도로 존재합니다.

2. Rebase 시작 및 새 커밋 생성:

    ```mermaid
    gitGraph
    commit id: "3dd37f8" tag: "main"
    commit id: "459594a"
    commit id: "11a11d6"
    branch feature
    commit id: "781afcc" type: HIGHLIGHT
    commit id: "19f0953"
    commit id: "7fba669"
    commit id: "2c2d8a3"
    commit id: "8b60c7e"
    commit id: "b66187a"
    commit id: "3f1254c"
    commit id: "80c9ca6"
    commit id: "5b9dfc6"
    commit id: "63ecf8c"
    commit id: "666d0fb"
    commit id: "1dbac3b"
    commit id: "0e93188"
    commit id: "e974e26" type: REVERSE
    branch origin/feat/another
    commit id: "02a68cf"
    ```

    - `feature` 브랜치에서 `git rebase main`을 실행합니다.
    - 새 커밋 `781afcc` ("remove error in some provider")가 생성되었습니다. 이는 rebase 과정에서 충돌 해결 또는 변경사항 적용으로 인해 만들어진 것으로 보입니다.
    - 기존 커밋들의 해시가 변경되었습니다
        - `ad82e81` → `781afcc`
        - `ee064d2` → `19f0953` 등
    - *커밋 순서가 유지되었지만, 각 커밋의 내용이 새로운 베이스에 맞게 조정*되었습니다.

3. 충돌 발생 및 rebase 중단:

    ```mermaid
    gitGraph
    commit id: "3dd37f8" tag: "main"
    commit id: "459594a"
    commit id: "11a11d6"
    branch feature
    commit id: "781afcc"
    commit id: "19f0953"
    commit id: "7fba669"
    commit id: "2c2d8a3"
    commit id: "8b60c7e"
    commit id: "b66187a"
    commit id: "3f1254c"
    commit id: "80c9ca6"
    commit id: "5b9dfc6"
    commit id: "63ecf8c"
    commit id: "666d0fb"
    commit id: "1dbac3b"
    commit id: "0e93188"
    commit id: "e974e26" type: REVERSE
    branch origin/feat/another
    commit id: "02a68cf"
    ```

    - `e974e26` ("add some repo") 커밋에서 충돌이 발생하여 rebase 과정이 중단되었습니다. 이 시점에서 사용자의 수동 개입이 필요합니다.
    - `origin/feat/another` 브랜치의 `02a68cf` 커밋은 아직 적용되지 않은 상태입니다.

4. 충돌 해결

    충돌 해결 후 `git rebase --continue`를 실행하여 남은 커밋들을 적용해야 합니다.
    `origin/feat/another` 브랜치의 커밋도 적용될 것입니다.

    Rebase 완료 후 원격 저장소에 강제로 푸시하여 동기화합니다.

    ```shell
    git push --force-with-lease
    ```
