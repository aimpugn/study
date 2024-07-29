# cherry-pick

- [cherry-pick](#cherry-pick)
    - [서로 다른 repo의 커밋 cherry-pick](#서로-다른-repo의-커밋-cherry-pick)
        - [체리픽 외의 다른 방법들](#체리픽-외의-다른-방법들)
    - [특정 커밋들만 다른 리파지토리에 붙여넣기](#특정-커밋들만-다른-리파지토리에-붙여넣기)

## 서로 다른 repo의 커밋 cherry-pick

`repoA`의 커밋을 `repoB`로 체리픽하는 것은 가능하지만 `repoB`에는 `repoA`의 커밋 해시가 없기 때문에, 단순히 커밋 해시를 사용하여 체리픽을 수행할 수는 없습니다. 대신, 두 저장소 간에 커밋을 공유하려면 몇 가지 추가적인 단계를 수행해야 합니다.

> **주의사항**
> - 체리픽은 충돌을 일으킬 수 있으므로, 충돌이 발생하면 수동으로 해결해야 합니다.
> - 체리픽은 특정 커밋만 선택적으로 적용할 때 유용하지만, 두 프로젝트 간에 지속적으로 많은 변경 사항을 공유해야 하는 경우 다른 방법을 고려하는 것이 좋습니다.

1. **`repoA`를 `repoB`의 리모트로 추가**: 먼저 `repoB`의 로컬 복사본에서 `repoA`를 리모트 저장소로 추가합니다.

   ```bash
   git remote add repoA https://github.com/org/repoA.git
   git fetch repoA
   ```

2. **체리픽 실행**: 이제 `repoA`의 커밋을 `repoB`로 체리픽할 수 있습니다. `repoA`에서 원하는 커밋의 해시를 찾은 후, 해당 커밋을 `repoB`에 체리픽합니다.

   ```bash
   git cherry-pick <commit-hash>
   ```

### 체리픽 외의 다른 방법들

1. **패치 파일 생성 및 적용**:

    `repoA`에서 변경 사항을 패치 파일로 생성하고, 이를 `repoB`에 적용할 수 있습니다.

   ```bash
   # repoA에서 패치 생성
   git format-patch -1 <commit-hash> --stdout > change.patch
   
   # repoB에서 패치 적용
   cd path/to/repoB
   git apply /path/to/change.patch
   ```

2. **서브모듈 사용**:
    - 공통 코드를 별도의 저장소로 분리하고, 두 프로젝트에서 서브모듈로 추가하는 방법입니다. 이 방법은 코드 중복을 줄이고 관리를 용이하게 합니다.

    ```bash
    # 공통 코드를 별도의 저장소로 분리 후
    git submodule add https://github.com/org/common-repo.git common
    git commit -m "Add common submodule"
    ```

3. **리포지토리 병합**:
    - `repoA`와 `repoB`가 매우 유사하고 지속적으로 동기화가 필요한 경우, 두 리포지토리를 하나로 병합하는 것을 고려할 수 있습니다.

4. **스크립트 또는 자동화 도구 사용**:
    - 두 리포지토리 간의 동기화를 자동화하기 위해 스크립트를 작성하거나 CI/CD 파이프라인을 구성할 수 있습니다.

## 특정 커밋들만 다른 리파지토리에 붙여넣기

아래는 this 리파지토리의 커밋 히스토리입니다.
이중에서 SomeUtil 작업에 해당하는 8eee8e0f1, 63dcd4d3c, f62ddec36 세 개 커밋만
다른 that 리파지토리에 그대로 갖다 붙이고 싶습니다.
그러니까 세 개의 커밋 8eee8e0f1, 63dcd4d3c, f62ddec36을 순서 그대로
this 리파지토리에서 that 리파지토리로 붙여 넣고 싶습니다.

```bash
- 8eee8e0f1 (HEAD -> feat/aaaa) SomeUtil: some3
- 8ae5dc302 (origin/feat/aaaa) Fix: another things
- 5e6b836f1 AnotherService: added something
- 63dcd4d3c SomeUtil: some2
- f62ddec36 SomeUtil: some1
- ad42b18ef ThatComponent: added that thing
```

Git의 'cherry-pick' 기능을 사용할 수 있습니다. 다음은 요청하신 작업을 수행하는 단계별 가이드입니다:

1. 먼저 'that' 리포지토리로 이동합니다:

    ```bash
    cd path/to/that/repository
    ```

2. 'this' 리포지토리를 원격 저장소로 추가합니다 (아직 추가되지 않은 경우):

    ```bash
    # git remote add this_repo path/to/this/repository

    git remote add some_cloned_repo ~/IdeaProjects/some_cloned_repo
    ```

3. 'this' 리포지토리의 변경사항을 가져옵니다:

    ```bash
    # git fetch this_repo

    git fetch some_cloned_repo
    ```

    만약 로컬 some_cloned_repo 에서 커밋을 수정한 후에 fetch 하면,
    변경 사항들을 잘 가져옵니다.

    ```bash
    ❯ git fetch some_cloned_repo
    remote: Enumerating objects: 17, done.
    remote: Counting objects: 100% (17/17), done.
    remote: Compressing objects: 100% (9/9), done.
    remote: Total 9 (delta 8), reused 0 (delta 0), pack-reused 0
    Unpacking objects: 100% (9/9), 3.03 KiB | 3.03 MiB/s, done.
    From /Users/rody/IdeaProjects/some_cloned_repo
    - 8eee8e0f1...9257895f5 feat/something-apply -> some_cloned_repo/feat/something-apply  (forced update)
    ```

4. 원하는 커밋들을 순서대로 cherry-pick 합니다:

    ```bash
    git cherry-pick f62ddec36
    git cherry-pick 63dcd4d3c
    git cherry-pick 8eee8e0f1
    ```

5. 충돌이 발생하면 해결하고 계속 진행합니다:

   ```bash
   # 충돌 해결 후
   git add .
   git cherry-pick --continue
   ```

6. 모든 cherry-pick이 완료되면, 변경사항을 'that' 리포지토리의 원격 저장소에 푸시합니다:

   ```bash
   git push origin branch_name
   ```

주의사항:
- cherry-pick 과정에서 충돌이 발생할 수 있습니다. 이 경우 수동으로 충돌을 해결해야 합니다.
- 커밋 해시는 리포지토리마다 다를 수 있으므로, 'this' 리포지토리에서 정확한 커밋 해시를 사용해야 합니다.
- cherry-pick은 새로운 커밋을 생성합니다. 따라서 'that' 리포지토리에서 새로운 커밋 해시가 생성됩니다.

이 방법을 사용하면 원하는 세 개의 커밋만 'that' 리포지토리로 가져올 수 있습니다. 작업 중 어려움이 있다면 추가적인 도움을 드릴 수 있습니다.
