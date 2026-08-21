# 리베이스는 커밋을 어떤 순서로 다시 적용하는가 — 위상 정렬과 갈래

## 목차

- [1. 시작 질문 — 왜 11-26 커밋이 12-15 커밋보다 나중에 적용되나](#1-시작-질문--왜-11-26-커밋이-12-15-커밋보다-나중에-적용되나)
- [2. 실험 저장소](#2-실험-저장소)
- [3. 커밋 이력은 DAG다](#3-커밋-이력은-dag다)
- [4. 위상 정렬은 하나가 아니다](#4-위상-정렬은-하나가-아니다)
- [5. 두 정렬을 나란히 놓기](#5-두-정렬을-나란히-놓기)
- [6. 리베이스는 어느 순서를 쓰나](#6-리베이스는-어느-순서를-쓰나)
- [7. 머지 커밋은 버려진다](#7-머지-커밋은-버려진다)
- [8. 그래서 무슨 충돌이 생기나](#8-그래서-무슨-충돌이-생기나)
- [9. 3-way 화면의 중앙은 base다](#9-3-way-화면의-중앙은-base다)
- [10. 같은 제목의 사본이 갈래마다 하나씩](#10-같은-제목의-사본이-갈래마다-하나씩)
- [11. 실무 대응](#11-실무-대응)
- [12. 한계와 주의](#12-한계와-주의)
- [출처](#출처)
- [관련 노트](#관련-노트)

## 1. 시작 질문 — 왜 11-26 커밋이 12-15 커밋보다 나중에 적용되나

갈래가 둘인 이력을 리베이스하면 이런 일이 벌어집니다. 12월 15일 커밋이 먼저 적용되고, 그 뒤에 11월 26일 커밋이 적용됩니다. 날짜만 보면 거꾸로입니다. 그리고 그 11월 커밋이 **이미 고쳐 놓은 것을 되돌리려 해서** 충돌합니다.

버그가 아닙니다. 리베이스는 커밋을 날짜 순으로 적용하지 않고 **갈래를 하나씩 통째로 쏟아냅니다.** 갈래 A의 커밋을 그 갈래 안에서 순서대로 다 내보낸 뒤 갈래 B로 넘어갑니다. A의 12월 커밋이 B의 11월 커밋보다 먼저 나오는 것은 그 결과입니다.

이 문서는 그 순서가 무엇으로 정해지는지, 왜 그렇게 정하는 것이 합당한지, 그리고 그 때문에 생기는 충돌을 어떻게 읽는지를 다룹니다. 아래 모든 출력값은 2장 스크립트로 만든 저장소에서 실제로 실행해 얻었습니다.

## 2. 실험 저장소

갈래 둘이 같은 파일을 건드리고 머지된 이력을 만듭니다. 갈래 A는 필드를 만들고 나중에 이름을 일반화하고, 갈래 B는 같은 작업을 따로 하되 일반화는 하지 않은 상태로 남습니다.

```bash
#!/usr/bin/env bash
set -eu
rm -rf expB && mkdir expB && cd expB

git init -q -b master
git config user.name "alice"; git config user.email "alice@example.com"
git config core.autocrlf false

c() {  # c <날짜> <제목>  — author/committer 날짜를 함께 고정
  GIT_AUTHOR_DATE="$1" GIT_COMMITTER_DATE="$1" git commit -qam "$2"
}

printf 'class Handler:\n    pass\n' > handler.py
printf 'x = 1\n' > other.py
git add -A
GIT_AUTHOR_DATE="2025-11-20 10:00:00 +0900" GIT_COMMITTER_DATE="2025-11-20 10:00:00 +0900" \
  git commit -qm "초기 커밋"
M0=$(git rev-parse HEAD)

# 갈래 A (본선): 만들고 -> 이름을 일반화하고 -> 다른 일을 더 함
git checkout -q -b trackA
printf 'class Handler:\n    special_common = None\n' > handler.py
c "2025-11-25 09:00:00 +0900" "핸들러 생성"
printf 'class Handler:\n    common = None\n' > handler.py
c "2025-12-09 09:00:00 +0900" "필드명 일반화"
printf 'y = 2\n' >> other.py
c "2025-12-15 09:00:00 +0900" "other 보강"

# 갈래 B (M0 에서 갈라져 같은 작업을 따로 함, 일반화는 없음)
git checkout -q -b trackB "$M0"
printf 'class Handler:\n    special_common = None\n    retry = 0\n' > handler.py
c "2025-11-26 09:00:00 +0900" "핸들러 생성"
printf 'z = 3\n' >> other.py
c "2025-12-10 09:00:00 +0900" "other 정리"

# 머지 (충돌은 일반화된 쪽으로 해소)
git checkout -q trackA
git merge --no-ff -q trackB -m "갈래 B 병합" || {
  printf 'class Handler:\n    common = None\n    retry = 0\n' > handler.py
  printf 'x = 1\ny = 2\nz = 3\n' > other.py
  git add handler.py other.py
  GIT_AUTHOR_DATE="2025-12-16 09:00:00 +0900" GIT_COMMITTER_DATE="2025-12-16 09:00:00 +0900" \
    git commit -q --no-edit -m "갈래 B 병합"
}

# master 가 전진
git checkout -q master
printf 'README\n' > README.md
git add README.md
GIT_AUTHOR_DATE="2025-12-20 09:00:00 +0900" GIT_COMMITTER_DATE="2025-12-20 09:00:00 +0900" \
  git commit -qm "README 추가"
```

만들어진 이력입니다.

```console
$ git log --graph --oneline --format='%h %ad %s' --date=short trackA
*   572e121 2025-12-16 갈래 B 병합
|\
| * b811471 2025-12-10 other 정리
| * 09e7c47 2025-11-26 핸들러 생성
* | 38b7c51 2025-12-15 other 보강
* | bad26bb 2025-12-09 필드명 일반화
* | 436ce35 2025-11-25 핸들러 생성
|/
* bbcc5e5 2025-11-20 초기 커밋
```

`핸들러 생성`이라는 제목이 두 번 나옵니다. `436ce35`(11-25, 갈래 A)와 `09e7c47`(11-26, 갈래 B)입니다. 같은 일을 두 갈래에서 따로 한 결과이고, 10장에서 이 중복이 문제가 됩니다.

모든 커밋의 author·committer 날짜를 고정했으므로 이 문서의 SHA는 그대로 재현됩니다. 여기까지 돌리면 `master`가 `3abd5a7`, `trackA`가 `572e121`, `trackB`가 `b811471`입니다. 다르게 나오면 날짜 고정이 빠진 것입니다.

## 3. 커밋 이력은 DAG다

순서를 이야기하려면 대상의 모양부터 정해야 합니다. git의 커밋 이력은 **DAG**, 곧 방향이 있고 순환이 없는 그래프입니다. 각 커밋이 부모를 가리키고, 자식이 부모의 조상이 되는 일은 없습니다.

위 그래프에서 확인할 수 있는 사실 하나가 이 문서 전체의 열쇠입니다. **`38b7c51`(A의 12-15)과 `09e7c47`(B의 11-26) 사이에는 부모-자식 관계가 없습니다.** 서로 조상이 아닙니다. 둘 다 `bbcc5e5`를 조상으로 두지만, 서로를 거슬러 닿지는 못합니다.

```console
$ git merge-base --is-ancestor 09e7c47 38b7c51; echo $?
1
$ git merge-base --is-ancestor 38b7c51 09e7c47; echo $?
1
```

양쪽 다 `1`, 곧 "조상이 아니다"입니다. 그래서 이 둘 사이에는 **지켜야 할 순서 제약이 아예 없습니다.** 어느 쪽을 먼저 놓아도 그래프의 규칙을 위반하지 않습니다.

## 4. 위상 정렬은 하나가 아니다

위상 정렬(topological order)은 DAG를 한 줄로 세우는 것이고, 조건은 하나입니다. **부모가 자식보다 먼저 나온다.**

그런데 그 조건을 만족하는 나열이 **여러 개** 있습니다. 3장에서 본 것처럼 서로 조상이 아닌 커밋 쌍은 순서가 자유롭기 때문입니다. 그중 어느 것을 고르느냐가 `--topo-order`와 `--date-order`의 차이입니다.

git 문서가 두 옵션을 이렇게 정의합니다.

```text
--date-order
    Show no parents before all of its children are shown, but
    otherwise show commits in the commit timestamp order.

--topo-order
    Show no parents before all of its children are shown, and
    avoid showing commits on multiple lines of history
    intermixed.
```

앞 절은 같습니다. 둘 다 위상 정렬입니다. 뒤가 다릅니다. `--date-order`는 남은 자유를 **타임스탬프 순**으로 쓰고, `--topo-order`는 **갈래가 섞이지 않는 쪽**으로 씁니다.

문서가 든 예시가 결정적입니다.

```text
    ---1----2----4----7
        \              \
         3----5----6----8---
```

숫자가 커밋 타임스탬프 순서일 때, `--date-order`는 `8 7 6 5 4 3 2 1`로, `--topo-order`는 `8 6 5 3 7 4 2 1`(또는 `8 7 4 2 6 5 3 1`)로 보여 줍니다. 그리고 문서가 그 이유를 직접 말합니다.

> some older commits are shown before newer ones in order to avoid showing the commits from two parallel development track mixed together.

**"오래된 커밋이 새 커밋보다 먼저 나온다"**를 문서가 스스로 적어 둔 것입니다. 그것이 부작용이 아니라 목적입니다. 두 갈래를 섞지 않는 대가로 날짜 순서를 포기합니다.

`--topo-order`의 답이 둘로 적혀 있는 것("또는")도 의미가 있습니다. 갈래를 섞지 않는다는 조건만으로는 **어느 갈래를 먼저 쏟을지가 안 정해집니다.** 그래서 구현이 정하는 부분이 남습니다.

## 5. 두 정렬을 나란히 놓기

실험 저장소에서 같은 구간을 두 순서로 나열해 봅니다.

```console
$ git log --date-order --format='%h %ad %s' --date=short bbcc5e5..trackA
572e121 2025-12-16 갈래 B 병합
38b7c51 2025-12-15 other 보강
b811471 2025-12-10 other 정리
bad26bb 2025-12-09 필드명 일반화
09e7c47 2025-11-26 핸들러 생성
436ce35 2025-11-25 핸들러 생성
```

날짜가 12-16, 12-15, 12-10, 12-09, 11-26, 11-25로 **엄격하게 내림차순**입니다. 그리고 갈래가 한 줄씩 번갈아 나옵니다 — A(12-15), B(12-10), A(12-09), B(11-26), A(11-25).

```console
$ git log --topo-order --format='%h %ad %s' --date=short bbcc5e5..trackA
572e121 2025-12-16 갈래 B 병합
b811471 2025-12-10 other 정리
09e7c47 2025-11-26 핸들러 생성
38b7c51 2025-12-15 other 보강
bad26bb 2025-12-09 필드명 일반화
436ce35 2025-11-25 핸들러 생성
```

이쪽은 B를 통째로(12-10, 11-26) 낸 뒤 A를 통째로(12-15, 12-09, 11-25) 냅니다. 그래서 **11-26이 12-15보다 먼저** 나옵니다. 4장에서 문서가 예고한 그 현상입니다.

두 목록의 커밋 집합은 완전히 같습니다. 순서만 다릅니다. 그런데 실무에서 이 차이가 큰 이유는, 같은 제목의 사본이 갈래마다 있을 때 **어느 갈래의 사본이 먼저 나오는지가 바뀌기** 때문입니다. 제목만 보고 목록을 대조하면 서로 다른 커밋을 같은 것으로 착각합니다.

## 6. 리베이스는 어느 순서를 쓰나

여기서 사실과 추측을 갈라야 합니다. git 문서는 리베이스의 순서를 이렇게만 적습니다.

```text
3. Replay the commits, one by one, in order. This is similar to running
   `git cherry-pick <commit>` for each commit.
```

"in order"라고만 하고 **어느 순서인지 명시하지 않습니다.** 그러니 문서 근거로는 "위상 순서를 쓴다"고 단정할 수 없습니다. 대신 직접 관측할 수 있습니다. `git rebase -i`가 만드는 todo 목록이 곧 적용 순서이고, 시퀀스 편집기를 `cat`으로 두면 그 목록을 그대로 볼 수 있습니다.

```console
$ GIT_SEQUENCE_EDITOR='cat' git rebase -i master
pick 436ce35 # 핸들러 생성
pick bad26bb # 필드명 일반화
pick 38b7c51 # other 보강
pick 09e7c47 # 핸들러 생성
pick b811471 # other 정리
```

이것을 위상 순서를 뒤집은 목록과 대조합니다.

```console
$ git rev-list --reverse --topo-order master..trackA --no-merges |
    while read c; do git log --format='%h %ad %s' --date=short -1 $c; done
436ce35 2025-11-25 핸들러 생성
bad26bb 2025-12-09 필드명 일반화
38b7c51 2025-12-15 other 보강
09e7c47 2025-11-26 핸들러 생성
b811471 2025-12-10 other 정리
```

**완전히 일치합니다.** git이 스스로 만든 적용 순서가 위상 순서(를 뒤집은 것)와 같습니다. `--reverse`를 쓰는 이유는 `rev-list`가 최신순으로 내보내는데 적용은 오래된 것부터 해야 하기 때문입니다. `--no-merges`를 쓰는 이유는 리베이스가 머지 커밋을 버리기 때문이고, 그 이야기는 7장에서 합니다.

**여기서 방향을 혼동하기 쉽습니다.** 5장의 `--topo-order` 목록은 갈래 B로 시작했는데 위 적용 순서는 갈래 A로 시작합니다. 모순이 아니라 같은 목록을 반대로 읽은 것입니다. `git log`는 최신순으로 내보내므로 목록의 **끝**이 가장 먼저 적용될 커밋이고, 뒤집으면 A가 앞에 옵니다. 순서를 이야기할 때는 "목록의 위"인지 "적용의 처음"인지 매번 정해 두어야 합니다.

정리하면 이렇습니다. 위상 순서를 쓴다는 것은 **문서가 보장한 규정이 아니라 관측된 동작**이고, 여기서는 서로 독립적인 두 방법(todo 목록과 `rev-list` 대조)이 같은 답을 냈습니다. 이 정도면 실무 판단의 근거로 충분하지만, 버전이 바뀌면 다시 확인할 대상입니다.

그리고 이 순서에서 갈래 A가 먼저 나온 것은 4장에서 본 "구현이 정하는 부분"입니다. **어느 갈래가 먼저인지를 규칙으로 기억하면 안 됩니다.** 확인하려면 위 todo 목록을 보는 것이 가장 확실합니다.

## 7. 머지 커밋은 버려진다

6장의 todo 목록에 없는 것이 하나 있습니다. `572e121 갈래 B 병합`입니다. 기본 리베이스는 머지 커밋을 재적용하지 않고 버립니다. 그래서 결과가 선형이 됩니다.

5장과 6장의 출력은 리베이스 **전**에 얻은 것입니다. 아래를 재현하려면 먼저 원본을 남겨 두고 리베이스합니다. 충돌은 8장에서 다루는 그 충돌이고, 여기서는 이미 적용된 결과(`ours`)를 택해 넘어갑니다.

```bash
export GIT_COMMITTER_DATE="2025-12-25 10:00:00 +0900"   # 결과 SHA 를 재현 가능하게

git branch -f trackA-orig trackA        # 원본 보존
git checkout -q trackA
git rebase master                       # handler.py, other.py 에서 충돌
```

```console
$ git rev-list --count master..trackA-orig      # 원본
6
$ git rev-list --count master..trackA           # 결과
5
$ git rev-list --count --merges master..trackA-orig
1
$ git rev-list --count --merges master..trackA
0
```

6커밋(머지 1개 포함)이 5커밋 0머지가 됐습니다. **줄어든 하나가 머지 커밋입니다.** 커밋 수가 안 맞을 때 먼저 이 값을 세 보면 설명이 붙습니다.

결과 이력을 보면 1장에서 말한 그 모양이 그대로 나옵니다.

```console
$ git log --format='%h %ad %s' --date=short master..trackA
b91a08a 2025-12-10 other 정리
2efb5f0 2025-11-26 핸들러 생성
a8c4cd8 2025-12-15 other 보강
2d89001 2025-12-09 필드명 일반화
52fe428 2025-11-25 핸들러 생성
```

아래에서 위로 읽으면 적용 순서입니다 — 11-25, 12-09, 12-15(갈래 A) 다음에 11-26, 12-10(갈래 B). **한 줄로 누웠는데도 날짜가 단조롭지 않습니다.** 갈래 경계가 그 지점입니다.

머지가 버려진다는 것은 순서 문제와 직결됩니다. 머지 커밋은 두 갈래가 만난 지점이고, **거기서 어떻게 합쳤는지가 그 커밋의 내용**입니다. 그것을 버리면 두 갈래를 한 줄로 다시 늘어놓아야 하고, 그 늘어놓는 방식이 위상 순서입니다.

머지 구조를 보존하려면 `--rebase-merges`를 씁니다. 대가는 결과가 선형이 아니게 되고, todo 목록이 `label`·`reset`·`merge` 명령을 포함해 훨씬 복잡해지는 것입니다. 선형 이력을 원해서 리베이스하는 경우가 대부분이니 기본 동작이 맞는 선택인 때가 많지만, **머지에서 손으로 해소한 내용이 있었다면 그 해소가 사라집니다.** 그 해소를 다시 하게 되는 것이 8장의 충돌입니다.

## 8. 그래서 무슨 충돌이 생기나

지금까지의 조각을 합치면 충돌의 정체가 나옵니다. 갈래 A 안에서는 시간 순서가 지켜집니다.

```text
갈래 A:  11-25 생성(special_common)  ->  12-09 일반화(common)  ->  12-15 다른 작업
```

그 다음 갈래 B가 통째로 옵니다.

```text
갈래 B:  11-26 생성 사본(special_common)   <- 이미 common 인데 special_common 으로 되돌리려 함
         12-10 다른 작업
```

**정리가 끝난 상태 위로 "정리 이전" 사본이 오는 것**이 충돌의 정체입니다. 실제로 그렇게 됩니다. 아래와 9장의 출력은 7장의 리베이스가 **멈춰 있는 동안** 관측한 것입니다. 7장을 이미 끝까지 진행했다면 원본에서 다시 시작해 첫 충돌 지점에서 멈춰야 같은 것을 볼 수 있습니다.

```console
$ git rebase master
Rebasing (1/5)Rebasing (2/5)Rebasing (3/5)Rebasing (4/5)
Auto-merging handler.py
CONFLICT (content): Merge conflict in handler.py
error: could not apply 09e7c47... 핸들러 생성
```

넷째 커밋에서 멈췄고, 그 커밋이 `09e7c47`, 곧 **11-26 커밋**입니다. 그 앞의 셋은 갈래 A의 11-25, 12-09, 12-15였습니다.

```console
$ git log --format='%h %ad %s' --date=short master..HEAD
a8c4cd8 2025-12-15 other 보강
2d89001 2025-12-09 필드명 일반화
52fe428 2025-11-25 핸들러 생성
```

12-15까지 적용된 상태에서 11-26을 얹으려다 부딪힌 것입니다. 날짜만 보면 거꾸로인데 그게 정상입니다.

## 9. 3-way 화면의 중앙은 base다

충돌 화면에서 사람을 가장 많이 헷갈리게 하는 것이 중앙 열입니다. 충돌 중 세 자리의 실제 내용을 꺼내 보면 분명해집니다. git은 충돌 중 index의 stage 1/2/3에 세 버전을 담아 둡니다.

```console
$ git show :1:handler.py        # stage 1 = base (중앙)
class Handler:
    pass
$ git show :2:handler.py        # stage 2 = ours (지금까지 적용된 결과)
class Handler:
    common = None
$ git show :3:handler.py        # stage 3 = theirs (적용하려는 커밋)
class Handler:
    special_common = None
    retry = 0
```

중앙이 `pass`입니다. 좌우 어느 쪽 내용도 아닙니다. **중앙은 적용하려는 커밋의 부모 시점**, 곧 갈래 B가 갈라져 나온 `초기 커밋`의 상태입니다. 그래서 중앙은 과거이고, 앞에서 무엇을 선택했든 바뀌지 않습니다.

여기서 `ours`와 `theirs`가 리베이스 중에는 뒤집혀 보인다는 점도 함께 짚어야 합니다. 리베이스는 새 기반을 체크아웃한 뒤 내 커밋을 그 위에 얹습니다. 그래서 **`ours`가 새 기반 쪽(지금까지 적용된 결과)이고, 얹히는 내 커밋이 `theirs`입니다.** 머지할 때의 감각과 반대입니다.

이 구조에서 실무적으로 중요한 결과가 나옵니다. **좌우가 서로 같은데 중앙만 다른 hunk가 생길 수 있습니다.** 갈래 B에도 일반화 커밋의 사본이 있는 경우가 그렇습니다. 그 커밋이 하려는 일(일반화)과 이미 적용된 결과(일반화됨)가 같으니 좌우가 같아지고, 중앙만 일반화 이전 상태로 남습니다. 이런 hunk는 어느 쪽을 눌러도 결과가 같습니다.

git 명령줄에서는 이런 경우 애초에 충돌로 보고되지 않고 조용히 적용되거나, 패치가 빌 때는 커밋이 버려집니다. **좌우가 같은 hunk를 화면에서 보게 되는 것은 IDE의 머지 도구가 충돌하지 않는 hunk까지 함께 보여 주기 때문입니다.** 예컨대 IntelliJ 계열의 머지 대화상자는 비충돌 hunk도 나열하고 `Apply non-conflicting changes: All` 버튼을 제공합니다. 그것을 먼저 눌러 비충돌 hunk를 걷어 내면 진짜 충돌만 남습니다.

## 10. 같은 제목의 사본이 갈래마다 하나씩

리베이스 결과의 제목을 세 보면 이렇습니다.

```console
$ git log --format='%s' master..trackA | sort | uniq -c | sort -rn
      2 핸들러 생성
      1 필드명 일반화
      1 other 정리
      1 other 보강
```

`핸들러 생성`이 두 개입니다. 갈래마다 하나씩 있던 사본이 한 줄로 늘어섰기 때문입니다. 이것이 두 가지 실무 문제를 만듭니다.

첫째, **제목으로 목록을 대조하면 서로 다른 커밋을 같은 것으로 착각합니다.** 5장에서 본 것처럼 정렬 방식에 따라 어느 갈래의 사본이 먼저 나오는지가 바뀌므로, 같은 제목을 보고 "이 커밋"이라고 지목하면 다른 갈래의 것을 가리킬 수 있습니다. 확인은 제목이 아니라 SHA나 내용으로 해야 합니다.

둘째, **제목 집합의 차집합으로 누락을 세는 방법이 중복을 접습니다.** 원본에 두 번 있던 제목이 결과에 한 번만 남아도 집합 비교로는 안 잡힙니다. 개수까지 세는 방법은 [`git/verify_rebase_result.md`](verify_rebase_result.md) 10장에 있습니다.

## 11. 실무 대응

순서를 이해하고 나면 충돌을 훨씬 빠르게 밀 수 있습니다.

1. **먼저 적용 순서를 확인합니다.** `GIT_SEQUENCE_EDITOR='cat' git rebase -i <upstream>`로 todo 목록을 뜨면 어느 갈래가 먼저인지, 어디쯤에서 갈래가 바뀌는지 보입니다. 목록만 보고 나가려면 그대로 종료하거나 `git rebase --abort`를 씁니다.
2. **갈래가 바뀌는 지점 이후의 충돌은 대개 "되돌림"입니다.** 정리가 끝난 상태 위로 정리 이전 사본이 오는 것이니, 그 사본이 하려는 되돌림을 받지 않는 쪽(이미 적용된 결과, 곧 `ours`)을 택합니다.
3. **다만 그것을 규칙으로 굳히지 않습니다.** 뒤에 오는 갈래에만 있는 **새 작업**이 `theirs`에 올 수도 있고, 그때는 그쪽이 정답입니다. 되돌림인지 새 작업인지는 최종본과 대조해 가립니다.

    ```bash
    # 최종본에서 그 파일이 어떤 상태인지 확인한다
    git show origin/master:path/to/file.py | grep -n "관심 있는 식별자"
    ```

    파일이 새 기반에 아예 없는 경우(그 갈래에서 처음 만든 파일)는 최종본이 곧 정답이므로 이 조회로 바로 결론이 납니다.

4. **IDE라면 비충돌 hunk를 먼저 일괄 적용합니다.** 9장에서 본 좌우가 같은 hunk들이 그것으로 지나갑니다.
5. **끝나면 결과를 검증합니다.** 커밋 수, 머지 수, 파일 단위 차이를 세는 방법은 [`git/verify_rebase_result.md`](verify_rebase_result.md)에 있습니다.

## 12. 한계와 주의

- **어느 갈래가 먼저인지는 규정이 아닙니다.** 4장에서 문서가 답을 둘로 적어 둔 그 자유입니다. todo 목록으로 확인하는 것 말고 안전한 방법은 없습니다.
- **리베이스가 위상 순서를 쓴다는 것도 문서에 명시돼 있지 않습니다.** 6장의 두 관측이 근거이고, git 2.54.0에서 확인한 것입니다.
- **`--date-order`는 커밋 타임스탬프를 봅니다.** author 날짜로 정렬하려면 `--author-date-order`가 따로 있습니다. 리베이스는 커밋을 다시 만들면서 committer 날짜를 새로 쓰고 author 날짜는 보존하므로, 리베이스 결과를 볼 때 두 날짜가 크게 벌어집니다. 어느 쪽을 보고 있는지 헷갈리면 `--format='%ad / %cd'`로 둘을 같이 찍어 봅니다. 2장 스크립트가 둘을 같은 값으로 고정한 것은 이 문서에서 날짜 혼동을 없애기 위한 것이고, 실제 저장소에서는 두 값이 다릅니다.
- **`--rebase-merges`를 쓰면 이 문서의 전제가 달라집니다.** 실제로 돌려 보면 todo 목록이 `label onto` / `reset onto` / `pick` / `merge -C 572e121 trackB` 형태가 되어 머지가 보존되고, 두 갈래가 각각 따로 재적용됩니다. 그래서 **8장의 "되돌림" 충돌은 나타나지 않고**, 대신 `merge` 단계에서 원래 머지 때 났던 충돌을 다시 해소하게 됩니다. 부수적으로 이때는 갈래 B가 먼저 재적용됐습니다 — 여기서도 갈래 순서는 규정이 아닙니다.
- **실험 저장소는 갈래가 둘입니다.** 셋 이상이면 갈래가 바뀌는 지점도 여러 번 나타납니다. 원리는 같지만 todo 목록을 더 주의해 읽어야 합니다.

## 출처

- 이 문서의 모든 그래프, 목록, 커밋 수, stage 내용은 2장 스크립트로 만든 저장소에서 직접 실행해 얻었습니다. 환경은 `git version 2.54.0.windows.1`입니다.
- `rev-list-options.adoc:880` 이하 `Commit Ordering` 절 — `--date-order`·`--author-date-order`·`--topo-order`의 정의, 4장의 예시 그래프와 `8 6 5 3 7 4 2 1`, 그리고 "some older commits are shown before newer ones in order to avoid showing the commits from two parallel development track mixed together." 문장. Git for Windows 설치본의 `mingw64/share/doc/git-doc/`.
- `git-rebase.adoc:78` — "Replay the commits, one by one, in order." 어느 순서인지는 명시되지 않습니다. 같은 디렉터리.
- **문서 근거가 아니라 실측**: 리베이스가 위상 순서로 적용한다는 것(6장), 갈래 A가 먼저 나온다는 것(5·6장), 그리고 12장의 `--rebase-merges` 동작. 앞의 둘은 todo 목록과 `rev-list --reverse --topo-order` 대조로, 마지막은 `--rebase-merges` 를 실제로 돌려 확인했습니다.
- **미검증**: 9장에서 IntelliJ 머지 대화상자가 비충돌 hunk를 함께 보여 준다는 것은 이 저장소의 [`git/rebase_intellij.md`](rebase_intellij.md)에 기록된 사용 경험에 따른 것이고, 이 문서를 쓰면서 다시 확인하지는 않았습니다.

## 관련 노트

- [`git/verify_rebase_result.md`](verify_rebase_result.md) — 리베이스 결과가 의도한 것인지 판정하는 명령들
- [`git/rebase_lost_commit_recovery.md`](rebase_lost_commit_recovery.md) — 이 순서 때문에 빈 커밋이 되어 사라진 작업을 되살리기
- [`git/branch_merge_detection.md`](branch_merge_detection.md) — 리베이스가 SHA를 바꿔 도달성 판정이 어긋나는 문제
- [`git/git_rebase.md`](git_rebase.md) — rebase 명령과 옵션 전반
- [`git/rebase_intellij.md`](rebase_intellij.md) — IDE 머지 도구에서의 충돌 해소
- [`git/log.md`](log.md) — `git log` 옵션
- [`git/conflict.md`](conflict.md) — 충돌 해소 전반
- [`git/merge-base.md`](merge-base.md) — 3장에서 쓴 `--is-ancestor`
