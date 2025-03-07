# git revert

- [git revert](#git-revert)
    - [revert?](#revert)
    - [Revert/롤백 필요한 경우](#revert롤백-필요한-경우)

## revert?

- `revert` 명령은 주어진 파일에 대해 이전의 변경 사항을 되돌리는 것이 아니라, 특정 커밋을 되돌리는 데 사용
- 파일의 이전 상태로 되돌리려면 checkout 명령을 사용해야 한다

## Revert/롤백 필요한 경우

> `develop` → → → → → →  `release/yyyymmdd` → `main`
>> ↘️ `A-branch` → `A PR` ↗️
>>> commit1  
> > > commit2
>
>> ↘️ `B-branch` →→→→→→→→→→→

- 시나리오들
    - [Git flow - get rid of a particular feature](https://stackoverflow.com/questions/28988952/git-flow-get-rid-of-a-particular-feature)

      > `F1`, `F2`, `F3` 기능 개발 완료  
      → `develop`으로 머지  
      → `release` 브랜치 생성  
      → `F3` 기능 제거 하려면?
      >
        - [`git revert`](https://git-scm.com/docs/git-revert) 로 머지된 커밋을 되돌리는 커밋을 생성  
          → `develop` 에 다시 머지할 때 **revert도 같이 머지** 된다는 문제 발생

            ```bash
            git checkout <release-branch>
            # --mainline 1: 변경 사항이 머지된 첫번째 부모 브랜치를 기준으로, 
            #               머지된 커밋의 변경 사항을 되돌린다. 
            #               이 케이스에서 `develop` 브랜치
            git revert --mainline 1 <hash-of-f3-merge-commit>
            ```

        - `develop` 대신 최신의 `master`로부터 `release` 브랜치를 생성하고, `[cherry-pick](https://git-scm.com/docs/git-cherry-pick)` 으로 포함시키고 싶은 피처들을 포함시키거나 `[merge](https://git-scm.com/docs/git-merge)` 사용해서 피처 브랜치를 병합시킨다  
          → 각 릴리즈가 개발된 피처의 작은 부분만 포함한다면 나쁘지 않지만, **많은 수의 기능을 포함한다면 쓰기 힘들다**.

            ```bash
            git cherry-pick --mainline 1 <hash-of-f1-merge-commit>
            git cherry-pick --mainline 1 <hash-of-f2-merge-commit>
            ```

    - [Git workflow - Reverting a feature branch from release branch](https://stackoverflow.com/questions/22013346/git-workflow-reverting-a-feature-branch-from-release-branch)

      > -F1---Commit1---Commit2---------------------------------------------  
      -F2--------Commit1--------------------------------------------------  
      -F3-------------------------Commit1-----Commit2---------------------  
      -Development-------Merge-F1------------Merge-F2---------------------  
      -Staging-------------------------------------------Merge-Development  
      `F1`을 제거하고, `F2` 와 `F3` 를 리뷰 위해 남기려면?

        1. `F1 revert` → `Development` → `Staging`: `F1` 피처는 이제 `Development` 과 `Staging` 모두에서 사라진다  
           → 머지 커밋을 `revert` 한다는 것은,  
           → 해당 머지로 인한 트리 변경 사항을 원하지 않음 의미하고,  
           → 이전에 **리버트 된 머지의 조상이 아닌 커밋**에 의해 도입된 트리 변경 사항만 가져온다  
           → `F1` 을 다시 도입하고 싶다면, `F1` 커밋을 **리베이스** 해야 한다
        2. `Staging` → reset to before `F1`, then merge `F2` & `F3`: `Staging` 은 `Development`의 하위 집합이 아니게 되며, `F1`은 `Staging` 에서만 사라진다
