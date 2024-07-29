# learngitbranching

## level intro3: `git merge`

### level intro3: 문제

1. Make a new branch called `bugFix`
2. Checkout the `bugFix` branch with git checkout bugFix

    ```shell
    git checkout -b bugFix
    ```

3. Commit once

    ```shell
    git commit
    ```

4. Go back to `main` with `git checkout`

    ```shell
    git checkout main
    ```

5. Commit another time

    ```shell
    git commit
    ```

6. Merge the branch `bugFix` into `main` with `git merge`

    ```shell
    git merge bugFix
    ```

## level intro4: `git rebase`

> Rebasing essentially takes a set of commits, "copies" them, and plops them down somewhere else
> `rebase`는 본질적으로 커밋 집합을 가져와서 복사하고 다른 곳에 버린다

리베이스의 장점은 커밋의 멋진 선형 시퀀스를 만드는 게 사용할 수 있다는 것이고,
따라서 리파지토리의 커밋 로그/히스토리가 더 깔끔해진다

```shell
# bugFix 브랜치의 작업(work)을 main 브랜치의 작업(work)으로 이동시키려면
# 현재 bugFix 브랜치일 때,
git rebase main
# 이러면, 기존 bugFix에 있던 C3 커밋은 어딘가에 존재하고,
# 새로운 C3' 라는 커밋은 main 브랜치로 rebase 된 "복사본(copy)"이다 
```

유일한 문제점은 아직 `main` 브랜치도 업데이트 되지 않았다는 것, 이를 업데이트 하려면

```shell
git checkout main
git rebase bugFix
# 앞서 `bugFix` 브랜치에서 `main` 브랜치를 리베이스 했기 때문에,
# 이제 `main` 브랜치가 `bugFix` 브랜치의 부모가 되고,
# git은 단순히 `main` 브랜치 참조(reference)를 history상에서 전진 시킨다(move forward)
```

### level intro4: 문제

1. Checkout a new branch named `bugFix`

```shell
git checkout -b bugFix
```

2. Commit once

```shell
git commit
```

3. Go back to main and commit again

```shell
git checkout main; git commit
```

4. Check out `bugFix` again and rebase onto `main`

```shell
git checkout bugFix; git rebase main
```

## level rampup1

프로젝트를 나타내는 커밋 트리를 따라 움직이는 다른 방법을 이해하는 것 중요하다.
편하게 돌아다질 수 있다면, 다른 깃 커맨드의 힘이 더 증폭된다

### level rampup1 설명

- `HEAD`:
    - 현재 체크아웃한 커밋에 대한 symbolic name
    - 항상 작업 트리에 반영된 가장 최근 커밋을 가리킨다
    - 일반적으로 `bugFix` 같은 브랜치 이름을 가리킨다
    - 커밋하면, `bugFix`의 상태가 변경되고, 이 변경 사항은 `HEAD` 통해서 볼 수 있다
- `Detaching HEAD`
    - 브랜치 대신 커밋에 `HEAD`를 붙인다는 의미
        1. `HEAD -> main -> C1`와 같은 모습이라고 할 때,
        2. `git checkout C1` 명령어 통해서 C1 커밋으로 체크아웃 하면,
        3. `HEAD -> C1`가 된다

### level rampup1 문제

1. To complete this level, let's detach `HEAD` from `bugFix` and attach it to the commit instead.
    - Specify this commit by its hash.
    - The hash for each commit is displayed on the circle that represents the commit.

```shell
git checkout C4
```

## level rampup2

### level rampup3 설명

- 현실에서는 나이스 한 커밋 트리 시각화가 없을 것이므로, 해시를 보기 위해 `git log`를 사용해야 할 것
- 실제 커밋 해시는 훨씬 길지만, 커밋을 유니크하게 식별할 정도로 충분한 해시 문자를 지정하는 것으로 충분.
    - `fed2da64c0efc5293610bdd892f82a58e8cbc5d8` 대신, `fed2`을 지정하는 것도 가능
- 해시로 커밋을 지정한느 것은 편한 방법이 아니므로, 깃은 relative refs 제공
- `Relative Refs`
    - 이를 통해 어딘가 기억할 수 있는 곳(`bugFix` 또는 `main` 브랜치)에서 시작하고 작업할 수 있다.
    - `^`: 한 번에 하나의 커밋 위로 이동
        - ref 이름에 붙일 때마다, 해당 커밋의 부모를 찾으라는 것
        - `main^`: `main`의 첫번째 부모
        - `main^^`: `main`의 조부모
    - `~<num>`: 숫자만큼 커밋 위로 이동

### level rampup3 문제

1. To complete this level, check out the parent commit of `bugFix`. This will detach `HEAD`. You can specify the hash if you want, but try using relative refs instead!

```shell
git checkout bugFix^
```

## level rampup3

### level rampup3: 설명

- 커밋 트리에서 많은 수준을 이동하고 싶을 경우, `~`(tilde) 연산자에 숫자를 붙여서 그만큼의 부모로 타고 올라갈 수 있다
- `relative refs`를 사용하는 가장 일반적인 방법중 하나는, 브랜치를 이동하는 것(move branches around.)
    - `git branch -f main HEAD~3`: 강제로(`-f`) `main` 브랜치를 `HEAD`의 세 수준 위의 부모로 이동시킨다

### level rampup3: 문제

> Now that you have seen `relative refs` and branch forcing in combination, let's use them to solve the next level.
> To complete this level, move `HEAD`, `main`, and `bugFix` to their goal destinations shown.

```shell
git checkout C1

git branch -f bugFix HEAD^
git branch -f main C6
```

## level rampup4

깃에는 변경 사항을 되돌리는 여러 방법이 있고, 커밋과 마찬가지로, 변경 사항을 되돌리는 것은 두 방법이 있다

1. low-level 컴포넌트(개별 파일 stating 또는 chunks): `git reset`
2. high-level 컴포넌트(변경 사항이 실제로 어떻게 되돌려지는지): `git revert`

### level rampup4: 설명

1. `git reset`
    - 브랜치 참조를 시간상 이전 커밋으로 이동시켜서 변경 사항을 되돌린다
    - 그런 의미에서 "기록을 다시 쓴다(rewriting history)"라고 생각할 수 있다
    - 애초에 커밋이 발생하지 않은 것처럼 브랜치를 되돌린다
    - `git reset HEAD~1`:
        - 현재 체크아웃한 커밋의 부모로 브랜치 참조 이동하고
        - 로컬 리파지토리는 현재 커밋이 없었던 것과 같은 상태가 된다
2. `git revert`
    - "기록을 다시 쓰는 것"은 다른 사람이 사용하는 remote branch에는 작동하지 않는다
    - 변경 사항을 되돌리고 다른 사람과 공유하려면, `git revert` 사용해야 한다
        - 이러면 되돌리고 싶은 커밋 다음에 새로운 커밋이 생긴다.
        - **커밋을 되돌린다**는 변경사항을 가져오기 때문

### level rampup4: 문제

> To complete this level, reverse the most recent commit on both `local` and `pushed`. You will revert two commits total (one per branch).  
> Keep in mind that pushed is a remote branch and local is a local branch -- that should help you choose your methods.

1. `local` 브랜치 되돌리기

```shell
git reset local^
```

2. remote의 `pushed` 브랜치 되돌리기

```shell
git checkout pushed
git revert pushed
```

## level move1

"moving work around"(이 작업을 여기로, 저 작업을 저기로)에 대해 배워보자

### level move1: 설명

1. `git cherry-pick <commit1> <commit2> <...>`
    - 원하는 커밋과 해시를 아는 경우 유용
    - 현재 위치(`HEAD`) 아래의 일련의 커밋을 복사
    - 시나리오
        1. `side` 브랜치에서 작업하고 이를 `main` 브랜치로 복사하려고 할 때,
        2. rebase도 가능하지만, cherry-pick도 가능:
           `git checkout main`
           `git cherry-pick C2 C4`: 체크아웃한 main 다음에 C2, C4 복사한 신규 커밋 생성

### level move1: 문제

> To complete this level, simply copy some work from the three branches shown into main. You can see which commits we want by looking at the goal visualization.

1. `bugFix`(`C3` 커밋) 브랜치

```shell
git cherry-pick C3
```

2. `side`(`C5` 커밋)의 부모 커밋 `C4`

```shell
git cherry-pick C4
```

3. `another`(`C7` 커밋)

```shell
git cherry-pick C7
```

결국 한 명령어로 보자면,

```shell
git cherry-pick C3 C4 C7
```

rebase로 문제를 푼다면?

```shell
git rebase C3
```

## level move2

### level move2: 설명

- 어떤 커밋을 원하는지 모를 경우? interactive rebase 사용해서 리베이스하려는 일련의 커밋 리뷰 가능
- `git rebase -i`
    - 리베이스 대상 아래로 어떤 커밋이 복사되는지 보여주는 UI가 나타난다
    - 커밋 해시, 메시지 등 또한 보여준다.
    - 할 수 있는 것
        1. 커밋의 순서를 바꿀 수 있다
        2. 모든 커밋을 유지하거나, 특정 커밋을 드랍할 수 있다.
        3. squashing(combining) commits
        4. amending commit messages
        5. commit 자체를 수정
- `git rebase -i HEAD~4`
    1. 커밋의 순서를 바꾼다
    2. 그러면 동일한 커밋을 복사한 새로운 커밋이 새로운 순서로 생성

### level move2: 문제

> To finish this level, do an interactive rebase and achieve the order shown in the goal visualization. Remember you can always `undo` or `reset` to fix mistakes :D

- AS-IS: `C0 <- C1 <- C2 <- C3 <- C4 <- C5(main*)`(화살표는 부모를 가리킨다)
- TO-BE: `C0 <- C1 <- C3' <- C5' <- C4'(main*)`

```shell
git rebase -i HEAD~4
```

## level mixed1

### level mixed1: 설명

### level mixed1: 문제
