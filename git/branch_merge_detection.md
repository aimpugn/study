# 브랜치가 정말 병합됐는지 어떻게 아는가 — 도달성과 patch-id

## 목차

- [1. 시작 질문 — `--merged` 목록에 없으면 안 병합된 것인가](#1-시작-질문----merged-목록에-없으면-안-병합된-것인가)
- [2. 가장 작은 벽돌 — 커밋은 부모를 가리키는 포인터다](#2-가장-작은-벽돌--커밋은-부모를-가리키는-포인터다)
- [3. 머지 커밋은 부모가 둘이다](#3-머지-커밋은-부모가-둘이다)
- [4. 판정의 1층 — 도달성](#4-판정의-1층--도달성)
- [5. 도달성이 끊기는 세 가지 방식](#5-도달성이-끊기는-세-가지-방식)
- [6. 판정의 2층 — patch-id](#6-판정의-2층--patch-id)
- [7. squash는 patch-id로도 안 잡힌다](#7-squash는-patch-id로도-안-잡힌다)
- [8. 판정의 3층 — 파일 내용 비교](#8-판정의-3층--파일-내용-비교)
- [9. `..`와 `...`는 다른 것을 센다](#9-와-는-다른-것을-센다)
- [10. 다축 비교](#10-다축-비교)
- [11. 한계와 실패 모드](#11-한계와-실패-모드)
- [12. 직접 확인해 보기](#12-직접-확인해-보기)
- [13. 실무 판정 순서](#13-실무-판정-순서)
- [출처](#출처)
- [관련 노트](#관련-노트)

## 1. 시작 질문 — `--merged` 목록에 없으면 안 병합된 것인가

아닙니다. `git branch --merged master`에 나오지 않아도 그 브랜치의 작업이 master에 이미 들어있는 경우가 있습니다. 병합 방식이 커밋의 SHA를 바꿔 버렸기 때문입니다.

판정은 한 층이 아니라 **세 층**입니다. 아래로 갈수록 더 많은 경우를 잡지만 더 비쌉니다.

1. **도달성(reachability)** — 브랜치 tip이 master에서 부모를 거슬러 닿는가. `git branch --merged`, `git merge-base --is-ancestor`, `git rev-list A..B`가 모두 이 하나를 봅니다.
2. **patch-id** — SHA는 달라도 변경 내용의 해시가 같은가. `git cherry`가 이것을 봅니다. rebase와 cherry-pick을 잡습니다.
3. **파일 내용 비교** — 브랜치가 건드린 파일이 master에서 같은 내용인가. squash까지 잡습니다.

세 층이 필요한 이유는 이 문서의 실험이 보여 줍니다. 브랜치 여덟 개를 서로 다른 방식으로 반영해 놓고 각 명령을 돌렸을 때, **squash로 반영한 브랜치는 1층과 2층 모두 "미반영"으로 판정했고 3층에서만 반영으로 드러났습니다.**

## 2. 가장 작은 벽돌 — 커밋은 부모를 가리키는 포인터다

판정 원리를 이해하려면 커밋 객체의 실물부터 봐야 합니다. `git cat-file -p`는 커밋 객체를 그대로 출력합니다.

미반영 브랜치의 tip 커밋입니다.

```text
tree d0f9c4ebddcb47fc43b08b2a0aac28623e1f97e3
parent 41c85888a8fd623cec7e970e6dea86e29d3a2167
author Test <t@e.com> 1767193200 +0900
committer Test <t@e.com> 1767193200 +0900

G1: 아직 미반영
```

줄마다 하중이 다릅니다.

- `tree` — 그 커밋 시점의 디렉터리 스냅샷입니다. 커밋은 변경분(diff)을 저장하지 않고 **전체 스냅샷**을 가리킵니다. diff는 두 tree를 비교해 그때그때 계산됩니다. 6장의 patch-id가 "저장된 값"이 아니라 "계산되는 값"인 이유가 여기 있습니다.
- `parent` — **이 한 줄이 판정의 전부입니다.** 커밋은 자기 앞 커밋을 가리키고, 그 앞 커밋도 또 앞을 가리킵니다. 그래서 커밋 하나를 잡으면 그 아래 이력 전체가 딸려 옵니다.
- `author` / `committer` — 두 사람이 다를 수 있습니다. rebase나 cherry-pick은 author를 보존하고 committer만 바꿉니다. 5장에서 SHA가 달라지는 이유의 일부입니다.

브랜치는 커밋이 아니라 **커밋을 가리키는 파일 한 개**입니다.

```bash
cat .git/refs/heads/feat/normal-merge
```

```text
ac470be4f1071deca2d8899fd9b2359b89567fca
```

40자 SHA 한 줄이 전부입니다. 그래서 브랜치를 지우는 것은 이 파일을 지우는 일이고, 커밋을 지우는 일이 아닙니다. 11장에서 이 사실이 "삭제해도 안전한가"의 답이 됩니다.

## 3. 머지 커밋은 부모가 둘이다

머지 커밋이 새로 추가되는 커밋인데도 그 이전 커밋들이 master에 반영된 것으로 판정되는 이유가 여기서 갈립니다. 같은 저장소의 머지 커밋을 열어 봅니다.

```bash
git cat-file -p 82bd6d7
```

```text
tree 43cfc7affcd3712d6fe0bcded844341886206cc8
parent 0cedc0d5fd9657ac9ef132b48664839d2f3d440d
parent ac470be4f1071deca2d8899fd9b2359b89567fca
author Test <t@e.com> 1767193200 +0900
committer Test <t@e.com> 1767193200 +0900

Merge branch 'feat/normal-merge'
```

일반 커밋과 딱 한 가지가 다릅니다. **`parent`가 두 줄입니다.**

- 첫 `parent` `0cedc0d`는 머지하기 직전의 master입니다(first parent).
- 둘째 `parent` `ac470be`는 병합해 들여온 브랜치의 tip입니다(second parent).

머지 커밋은 "브랜치의 커밋들을 복사해 오는" 것이 아닙니다. **브랜치 tip을 두 번째 부모로 가리키기만 합니다.** 그 한 줄의 포인터로 브랜치 이력 전체가 master의 조상 집합에 편입됩니다.

브랜치가 만든 첫 커밋 `A1`(`6f23c8a`)이 master에서 실제로 닿는지 확인해 봅니다.

```bash
git merge-base --is-ancestor 6f23c8a master; echo "exit=$?"
git rev-list master | grep -c 6f23c8a
```

```text
exit=0
1
```

`exit=0`은 조상이라는 뜻이고, `rev-list`가 1건을 세었다는 것은 master에서 부모를 거슬러 올라가다 그 커밋을 실제로 만났다는 뜻입니다. `A1`을 master가 직접 부모로 가리키는 곳은 어디에도 없습니다. 경로는 이렇게 이어집니다.

```text
master(82bd6d7 머지커밋) --parent2--> ac470be(A2) --parent--> 6f23c8a(A1) --parent--> 0cedc0d(base)
```

**이것이 "머지 커밋 이전 커밋들이 어떻게 반영됐다고 판정되는가"의 답입니다.** 반영 여부는 커밋을 복사했는지가 아니라, **그래프에서 걸어갈 수 있는지**로 정의됩니다. 이 걸어갈 수 있음을 도달성(reachability)이라고 부릅니다.

fast-forward 머지는 부모가 둘인 커밋조차 만들지 않습니다. master ref가 브랜치 tip으로 옮겨가기만 하므로 브랜치 tip이 곧 master가 되고, 도달성은 자동으로 성립합니다.

## 4. 판정의 1층 — 도달성

도달성을 묻는 명령이 셋 있는데, **세 명령은 같은 것을 계산하고 표면만 다릅니다.**

### 4.1 `git merge-base --is-ancestor`

가장 직접적입니다. 공식 문서가 정의를 이렇게 적습니다.

> Check if the first `<commit>` is an ancestor of the second `<commit>`, and exit with status 0 if true, or with status 1 if not. Errors are signaled by a non-zero status that is not 1.
> — `git-merge-base.adoc`

출력이 없고 **종료 코드로만** 답한다는 점이 중요합니다. 셸에 따라 종료 코드를 꺼내는 방법이 달라 이 명령이 가장 자주 걸립니다(12장에 셸별 형태를 정리했습니다).

### 4.2 `git rev-list --count master..branch`

같은 판정을 **숫자로** 내놓습니다. `A..B`는 "B에서 도달 가능하고 A에서는 도달 불가능한 커밋"의 집합이므로, 그 개수가 0이면 브랜치의 모든 커밋이 master에서 도달 가능하다는 뜻입니다.

종료 코드가 아니라 stdout으로 답하므로 셸 문법에 걸리지 않습니다. 실무에서는 이 형태가 가장 다루기 쉽습니다.

### 4.3 `git branch --merged master`

브랜치 목록 전체를 한 번에 훑습니다. 내부적으로는 각 브랜치 tip에 대해 4.1과 같은 조상 판정을 합니다.

세 명령이 같은 것을 본다는 사실은 실험 결과가 그대로 보여 줍니다. 아래 표에서 `--merged`가 `YES`인 행과 `rev-list..`가 `0`인 행이 정확히 일치합니다.

| 브랜치 | 반영 방식 | `--merged` | `rev-list master..B` |
|---|---|---|---|
| `feat/normal-merge` | 머지 커밋 | YES | 0 |
| `feat/ff` | fast-forward | YES | 0 |
| `feat/squashed` | squash | no | 2 |
| `feat/rebased` | rebase 후 머지 | no | 1 |
| `feat/cherrypicked` | cherry-pick | no | 1 |
| `feat/not-merged` | 반영 안 함 | no | 1 |
| `feat/orphan` | 무관한 이력 | no | 1 |
| `feat/edited-pick` | 반영 중 수정 | no | 1 |

여기서 문제가 드러납니다. **`feat/squashed`, `feat/rebased`, `feat/cherrypicked`는 내용이 master에 실제로 들어가 있는데도 `feat/not-merged`와 똑같이 "미반영"으로 나옵니다.** 1층만으로는 이 넷을 가를 수 없습니다.

## 5. 도달성이 끊기는 세 가지 방식

왜 끊기는지는 커밋 SHA가 무엇으로 결정되는지에서 나옵니다. 커밋 SHA는 2장에서 본 객체 내용 전체(tree + parent + author + committer + 메시지)의 해시입니다. **그 중 하나라도 다르면 다른 SHA가 됩니다.**

- **rebase** — 커밋을 새 부모 위에 다시 적용합니다. `parent`가 바뀌므로 SHA가 바뀝니다. 변경 내용과 author는 그대로입니다.
- **cherry-pick** — 다른 브랜치의 변경을 현재 위치에 적용합니다. 역시 `parent`가 달라 SHA가 바뀝니다.
- **squash** — 여러 커밋을 하나로 합칩니다. `parent`뿐 아니라 **커밋의 개수 자체가 N에서 1로 줄어듭니다.** 7장에서 이 차이가 결정적으로 작용합니다.

실험에서 rebase된 커밋의 원본과 사본을 나란히 보면 이렇습니다.

```text
feat/rebased 의 tip : 6ef097f  "D1"
master 쪽 사본      : bb612c2  "D1"
```

같은 변경, 같은 메시지, 다른 SHA입니다. `feat/rebased`라는 ref는 여전히 옛 SHA `6ef097f`를 가리키고 master는 `bb612c2`를 가지므로, 그래프상 두 커밋은 별개입니다. 그래서 도달성 판정이 "미반영"이라고 답합니다. **판정이 틀린 것이 아니라, 도달성이라는 질문에 정직하게 답한 것입니다.** 질문을 바꿔야 합니다.

GitHub과 GitLab의 "Squash and merge", "Rebase and merge" 버튼이 정확히 이 상황을 만듭니다. PR은 병합됐다고 표시되는데 로컬에서 `git branch --merged`를 돌리면 그 브랜치가 안 나옵니다.

## 6. 판정의 2층 — patch-id

질문을 "같은 커밋인가"에서 "같은 변경인가"로 바꾸면 SHA에 의존하지 않을 수 있습니다. 그 답이 patch-id입니다.

### 6.1 patch-id란 무엇인가

공식 문서의 정의입니다.

> A "patch ID" is nothing but a sum of SHA-1 of the file diffs associated with a patch, with line numbers ignored. As such, it's "reasonably stable", but at the same time also reasonably unique, i.e., two patches that have the same "patch ID" are almost guaranteed to be the same thing.
> — `git-patch-id.adoc`

두 가지를 짚어야 합니다.

- **파일 diff들의 SHA-1을 합한 값**입니다. 커밋 객체의 해시가 아니라 **변경 내용의 해시**입니다. 그래서 부모가 바뀌어도 값이 유지됩니다.
- **줄 번호를 무시합니다.** 같은 변경이 파일의 다른 위치에 적용돼도 같은 값이 나옵니다. 문서가 "reasonably stable"이라고 표현한 것이 이 성질입니다.

직접 계산해 볼 수 있습니다. `git show`의 출력을 `git patch-id`에 넘기면 됩니다.

```bash
git show feat/rebased | git patch-id --stable
```

실험에서 원본과 사본의 값을 뽑아 보면 이렇습니다.

```text
rebase       feat/rebased  D1 : 84c1e8004705
             master 쪽 사본 D1 : 84c1e8004705   <- 같다

cherry-pick  feat/cherrypicked J1 : 9d5a741708b7
             master 쪽 사본     J1 : 9d5a741708b7   <- 같다
```

SHA는 달랐지만 patch-id는 동일합니다. **이것이 rebase와 cherry-pick을 잡아내는 근거입니다.**

### 6.2 `git cherry`가 하는 일

`git cherry`가 그 비교를 자동으로 합니다. 문서의 설명입니다.

> The equivalence test is based on the diff, after removing whitespace and line numbers. git-cherry therefore detects when commits have been "copied" by means of git-cherry-pick, git-am or git-rebase.
>
> Outputs the SHA1 of every commit in `<limit>..<head>`, prefixed with `-` for commits that have an equivalent in `<upstream>`, and `+` for commits that do not.
> — `git-cherry.adoc`

출력 규약이 명확합니다. `-`는 upstream에 동등한 것이 있다(반영됨), `+`는 없다(미반영)입니다.

rebase된 브랜치에 돌려 봅니다.

```bash
git cherry --verbose master feat/rebased
```

```text
- 6ef097ffcc7bf47811443733c500deadc877dd5b D1
```

`-`가 붙었습니다. 도달성 판정은 "미반영"이라고 했지만 patch-id 판정은 "반영됨"이라고 답합니다. **2층이 1층의 사각지대를 덮었습니다.**

### 6.3 문서가 그어 놓은 경계

위 인용문에서 반드시 읽어야 할 대목은 탐지 대상을 나열한 부분입니다. **`git-cherry-pick`, `git-am`, `git-rebase` 셋만 적혀 있고 squash는 없습니다.** 이 누락은 실수가 아니라 원리상의 경계이며, 다음 장이 그 이유입니다.

## 7. squash는 patch-id로도 안 잡힌다

squash로 반영한 브랜치에 같은 명령을 돌리면 결과가 뒤집힙니다.

```bash
git cherry --verbose master feat/squashed
```

```text
+ a7bf9f1c69065dcbcfc0c116a3bc0cab4a97fbd0 C1
+ 0f5db9952337d4ad0fa8d242ff200b96fcec66e1 C2
```

둘 다 `+`, 즉 미반영입니다. 그런데 그 내용은 master에 분명히 들어 있습니다(8장에서 확인합니다). patch-id를 직접 찍어 보면 이유가 그대로 보입니다.

```text
feat/squashed  C1        : bba6ec24dff1
feat/squashed  C2        : b25916c5ceea
master  squash 커밋       : 41d81c47894b
```

**셋이 전부 다릅니다.** 6.1의 정의가 이 결과를 설명합니다. patch-id는 "그 패치에 딸린 파일 diff들의 SHA-1 합"입니다.

- `C1`의 patch-id는 `C1`이 만든 diff의 해시입니다.
- `C2`의 patch-id는 `C2`가 만든 diff의 해시입니다.
- squash 커밋의 patch-id는 **`C1`과 `C2`를 합친 diff**의 해시입니다.

합친 diff는 각각의 diff와 다른 텍스트이므로 해시도 다릅니다. patch-id는 커밋 대 커밋의 1:1 대응을 전제로 설계됐고, squash는 **N:1 변환**이라 그 전제가 깨집니다. rebase와 cherry-pick이 잡히는 것은 둘이 1:1 변환이기 때문입니다.

이 결과로 5장의 세 방식이 두 부류로 갈립니다.

| 변환 | 커밋 대응 | SHA | patch-id | `git cherry` |
|---|---|---|---|---|
| rebase | 1:1 | 바뀜 | **유지** | 잡는다 (`-`) |
| cherry-pick | 1:1 | 바뀜 | **유지** | 잡는다 (`-`) |
| squash | **N:1** | 바뀜 | **바뀜** | **못 잡는다 (`+`)** |

squash 대상 커밋이 **하나뿐**이었다면 N:1이 1:1이 되므로 patch-id가 유지되고 `git cherry`가 잡습니다. 즉 "squash는 항상 못 잡는다"가 아니라 **"두 개 이상을 합쳤을 때 못 잡는다"**가 정확한 진술입니다.

## 8. 판정의 3층 — 파일 내용 비교

커밋 단위 판정이 통하지 않으면 커밋을 포기하고 **파일 내용**을 봅니다. 브랜치가 건드린 파일 목록을 뽑아, 그 파일들만 master와 비교하는 방식입니다.

```bash
git diff --stat master feat/squashed -- $(git diff --name-only master...feat/squashed)
```

두 부분으로 나뉩니다.

- `git diff --name-only master...feat/squashed` — 점 **세 개**입니다. 브랜치가 갈라진 뒤 **브랜치에서** 건드린 파일 목록을 냅니다(9장 참조).
- `git diff --stat master feat/squashed -- <그 파일들>` — 점 **없이** 두 ref를 직접 비교합니다. 즉 두 tree의 실제 차이를 봅니다. 범위를 그 파일들로 좁혔기 때문에 master가 다른 곳에서 앞서간 변경은 섞이지 않습니다.

네 브랜치에 돌린 결과입니다.

```text
feat/squashed    : 내용 일치 -> 반영됨
feat/rebased     : 내용 일치 -> 반영됨
feat/not-merged  : 차이 있음 -> 미반영
                     g.txt | 1 +
feat/edited-pick : 차이 있음 -> 미반영
                     i.txt | 2 +-
```

**`feat/squashed`가 드디어 "반영됨"으로 판정됐습니다.** 1층과 2층이 모두 놓친 경우입니다.

`feat/edited-pick`이 "미반영"으로 나온 것도 정확한 판정입니다. 그 브랜치는 반영 과정에서 한 줄이 수정됐으므로 브랜치의 내용과 master의 내용이 실제로 다릅니다. 이때 "반영됐다"고 답하는 것이 오히려 틀린 답입니다.

### 8.1 보조 도구 — `git range-diff`

커밋 대응 관계를 사람이 눈으로 확인하려면 `git range-diff`가 유용합니다. 두 커밋 범위를 짝지어 보여 줍니다.

```bash
git range-diff <merge-base>..feat/rebased <merge-base>..master
```

```text
-:  ------- > 1:  f5a31c5 E1: master 선행
1:  6ef097f = 2:  bb612c2 D1
-:  ------- > 3:  41c8588 F1
```

가운데 줄의 `=`가 "이 두 커밋은 같은 패치"라는 표시입니다. 왼쪽 `6ef097f`(브랜치 원본)와 오른쪽 `bb612c2`(master 사본)가 짝지어졌습니다. squash된 브랜치에 같은 것을 돌리면 `=`가 하나도 나오지 않습니다.

`range-diff`는 판정을 자동화하기보다 **왜 그렇게 판정됐는지 눈으로 확인**할 때 씁니다.

## 9. `..`와 `...`는 다른 것을 센다

앞 장들에서 점 두 개와 세 개가 섞여 나왔습니다. 이 둘을 혼동하면 판정 자체가 어긋나므로 여기서 갈라 둡니다. 더 까다로운 점은 **`git log`/`git rev-list` 계열과 `git diff`에서 `...`의 의미가 서로 다르다**는 것입니다.

| 표기 | `git rev-list` / `git log` | `git diff` |
|---|---|---|
| `A..B` | B에서 도달 가능하고 A에서 불가능한 커밋 | A와 B **두 tree의 직접 비교**(`A B`와 동일) |
| `A...B` | 한쪽에서만 도달 가능한 커밋 전부(대칭차) | `merge-base(A,B)`와 B의 비교 |

그래서 같은 `...`를 써도 두 명령이 다르게 답합니다.

- `git rev-list --left-right --count master...branch` -> `왼쪽<TAB>오른쪽`. 왼쪽은 master에만 있는 커밋 수, 오른쪽은 브랜치에만 있는 커밋 수입니다. **오른쪽이 0인 것이 도달성 판정입니다.**
- `git diff --shortstat master...branch` -> 브랜치가 갈라진 뒤 **브랜치에서 한 작업**의 크기입니다.

여기서 흔한 오해가 하나 생깁니다. **`git diff --shortstat master...branch`의 출력이 비었다는 것은 "반영됐다"는 뜻이 아닙니다.** 브랜치가 merge-base 이후 아무 작업도 하지 않았다는 뜻일 뿐입니다. 두 진술이 실무에서 자주 함께 성립하는 이유는, 브랜치가 이미 완전히 반영됐으면 merge-base가 브랜치 tip 자신이 되어 diff가 0이 되기 때문입니다. 그러나 squash된 브랜치는 반영됐는데도 이 출력이 비지 않습니다.

실험이 그 반례입니다. `feat/squashed`는 내용이 master에 있는데도 3-dot diff가 `1 file changed, 1 insertion(+)`을 냅니다. 반영 여부 판정에 3-dot diff를 쓰면 안 되는 이유입니다.

## 10. 다축 비교

네 가지 판정 방법을 축별로 갈라 놓습니다. 각 칸은 12장의 실험 실행 결과에서 나온 값입니다.

| 축 | `branch --merged` | `rev-list --count A..B` | `git cherry` | 파일별 2-dot diff |
|---|---|---|---|---|
| 무엇을 비교하나 | 커밋 SHA(도달성) | 커밋 SHA(도달성) | patch-id(변경 내용) | 파일 내용(tree) |
| 답의 형태 | 브랜치 목록 | 숫자 | 커밋별 `+`/`-` | diff 유무 |
| 일반 merge | 잡는다 | 잡는다 | 잡는다 | 잡는다 |
| fast-forward | 잡는다 | 잡는다 | 잡는다 | 잡는다 |
| rebase 후 반영 | **놓친다** | **놓친다** | 잡는다 | 잡는다 |
| cherry-pick 반영 | **놓친다** | **놓친다** | 잡는다 | 잡는다 |
| squash 반영(2커밋 이상) | **놓친다** | **놓친다** | **놓친다** | 잡는다 |
| 반영 중 내용 수정 | 미반영 | 미반영 | 미반영 | 미반영(정확) |
| 공통 조상 없음 | 미반영 | 미반영 | 동작하나 무의미 | 3-dot이 `fatal` |
| 비용 | 브랜치 수에 비례 | 커밋 수에 비례 | **커밋마다 diff 계산** | 파일 수에 비례 |
| 한 번에 전체 훑기 | 가능 | 브랜치별 반복 | 브랜치별 반복 | 브랜치별 반복 |
| 셸 문법 취약점 | 없음 | 없음 | 없음 | 명령 치환 필요 |

비용 축을 덧붙이면 실무 순서가 정해집니다. `git cherry`는 범위 안의 **커밋마다 diff를 계산**하므로 커밋이 수백 개인 브랜치에서는 눈에 띄게 느립니다. 그래서 `--merged`로 대부분을 먼저 걸러내고, 남은 것에만 `cherry`를 쓰는 순서가 낫습니다(13장).

## 11. 한계와 실패 모드

### 11.1 공통 조상이 없는 브랜치

`--orphan`으로 만든 브랜치나 이력을 재작성한 저장소에서는 두 이력이 공통 조상을 갖지 않습니다. 이때 3-dot을 쓰는 명령이 실패합니다.

```text
fatal: master...feat/orphan: no merge base
```

`git cherry`는 오류 없이 값을 내지만, 비교 기준점이 없으므로 브랜치의 모든 커밋을 `+`로 표시합니다. **이 숫자를 "미반영 N건"으로 읽으면 안 됩니다.** 판정 자체가 성립하지 않는 상태입니다.

이 경우 쓸 수 있는 것은 tree 비교뿐입니다. `git diff master branch`로 두 tree의 실제 차이를 보는 것이 유일하게 의미 있는 질문입니다.

공통 조상 유무는 이렇게 확인합니다.

```bash
git merge-base master feat/orphan; echo "exit=$?"
```

종료 코드가 1이면 공통 조상이 없습니다.

### 11.2 반영 중 내용이 수정된 경우

충돌을 해소하면서 한 줄이라도 바뀌면 patch-id가 달라집니다.

```text
feat/edited-pick 원본 : d26661446878
master 쪽 수정된 사본  : 8ca14939cc9d
```

`git cherry`는 이것을 미반영으로 판정합니다. 그리고 그 판정은 **틀린 것이 아닙니다.** 내용이 실제로 다르므로, "이 브랜치를 지워도 잃는 것이 없는가"라는 실무 질문에는 "그 한 줄 차이만큼은 잃는다"가 정확한 답입니다. 무엇이 달라졌는지는 파일 비교로 확인해야 합니다.

### 11.3 되돌려진(revert) 작업

브랜치의 작업이 master에 들어왔다가 `git revert`로 되돌려진 경우, 도달성은 여전히 성립하므로 `--merged`가 "반영됨"으로 답합니다. 그러나 최종 tree에는 그 변경이 없습니다. 즉 **"반영됐다"와 "지금 코드에 있다"는 다른 질문입니다.** 세 층 중 어느 것도 후자를 자동으로 답해 주지 않으므로, 되돌림이 의심되면 파일 내용을 직접 확인해야 합니다.

이 문서의 실험에는 revert 시나리오가 없으므로, 위 진술은 도달성의 정의에서 나온 **추론**입니다. 직접 관측으로 확인하지 않았습니다.

### 11.4 `git cherry`의 인자 순서

`git cherry <upstream> <head>`입니다. 순서를 뒤집으면 정반대 질문이 됩니다. `git cherry master feat/x`는 "feat/x의 커밋 중 master에 없는 것"을 묻고, `git cherry feat/x master`는 "master의 커밋 중 feat/x에 없는 것"을 묻습니다. 후자는 거의 항상 큰 수가 나오는데 그것을 미반영으로 읽으면 판정이 뒤집힙니다.

### 11.5 셸이 옵션을 가로채는 문제

판정 명령 자체와 무관하지만 실제로 자주 막히는 지점입니다.

PowerShell에서 `-v`를 넘기면 `-version`으로 확장돼 git이 `--version`으로 받습니다.

```text
error: unknown option `version'
usage: git cherry [-v] [<upstream> [<head> [<limit>]]]
```

긴 이름 `--verbose`를 쓰거나 stop-parsing 토큰 `--%`를 씁니다.

Nushell에서는 `&&`가 없습니다.

```text
Error: nu::parser::shell_andand
  × The '&&' operator is not supported in Nushell
  help: use ';' instead of the shell '&&', or 'and' instead of the boolean '&&'
```

그래서 종료 코드로만 답하는 `--is-ancestor`를 Nushell에서 쓰려면 손이 더 갑니다. **stdout으로 답하는 `rev-list --count` 쪽이 셸에 덜 의존합니다.**

## 12. 직접 확인해 보기

이 문서의 모든 관측값은 아래 스크립트로 만든 저장소에서 나왔습니다. 그대로 실행하면 같은 판정 결과를 재현할 수 있습니다(SHA는 실행 환경의 날짜·이름에 따라 달라집니다).

측정 환경은 `git version 2.54.0.windows.1`, Git Bash입니다.

### 12.1 실험 저장소 만들기

```bash
set -e
export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=t@e.com
export GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=t@e.com
export GIT_AUTHOR_DATE="2026-01-01T00:00:00+0900"
export GIT_COMMITTER_DATE="2026-01-01T00:00:00+0900"

git init -q -b master merge-detect-demo && cd merge-detect-demo
git config core.autocrlf false

c() { printf '%s\n' "$2" > "$1"; git add "$1"; git commit -q -m "$3"; }
c base.txt base "base: 최초 커밋"

# (1) 일반 merge
git switch -qc feat/normal-merge; c a.txt a1 "A1"; c a.txt a2 "A2"
git switch -q master; git merge --no-ff -m "Merge branch 'feat/normal-merge'" feat/normal-merge >/dev/null

# (2) fast-forward
git switch -qc feat/ff; c b.txt b1 "B1"
git switch -q master; git merge --ff-only feat/ff >/dev/null

# (3) squash (커밋 2개를 1개로)
git switch -qc feat/squashed; c c.txt c1 "C1"; c c.txt c2 "C2"
git switch -q master; git merge --squash feat/squashed >/dev/null
git commit -q -m "기능C를 한 커밋으로 (squash)"

# (4) rebase 후 머지, 원 브랜치 ref 는 옛 SHA 유지
git switch -qc feat/rebased; c d.txt d1 "D1"
git switch -q master; c e.txt e1 "E1: master 선행"
git switch -qc _tmp feat/rebased; git rebase master >/dev/null 2>&1
git switch -q master; git merge --ff-only _tmp >/dev/null; git branch -D _tmp >/dev/null

# (5) cherry-pick (부모가 달라 새 SHA 가 되도록 master 를 먼저 진행)
git switch -qc feat/cherrypicked; c j.txt j1 "J1"
git switch -q master; c k.txt k1 "K1: master 선행2"
git cherry-pick feat/cherrypicked >/dev/null 2>&1

# (6) 미반영
git switch -qc feat/not-merged; c g.txt g1 "G1: 아직 미반영"

# (7) orphan
git switch -q master; git switch -q --orphan feat/orphan; c h.txt h1 "H1: 무관한 이력"

# (8) 반영 중 내용 수정
git switch -q master; git switch -qc feat/edited-pick; c i.txt i1 "I1"
git switch -q master; git cherry-pick feat/edited-pick >/dev/null 2>&1
printf '%s\n' "i1-fixed" > i.txt; git add i.txt; git commit -q --amend -m "I1 (반영 중 한 줄 수정)"
git switch -q master
```

### 12.2 판정 표 재현

```bash
MERGED=$(git branch --merged master --format='%(refname:short)' | tr '\n' ' ')
for B in feat/normal-merge feat/ff feat/squashed feat/rebased \
         feat/cherrypicked feat/not-merged feat/orphan feat/edited-pick; do
  case " $MERGED " in *" $B "*) M=YES;; *) M=no;; esac
  R=$(git rev-list --count "master..$B")
  P=$(git cherry master "$B" 2>/dev/null | grep -c '^+' || true)
  printf '%-22s merged=%-4s ahead=%-3s cherry+=%s\n' "$B" "$M" "$R" "$P"
done
```

PASS 기준은 이렇습니다.

- `feat/normal-merge`, `feat/ff` -> `merged=YES ahead=0 cherry+=0`
- `feat/rebased`, `feat/cherrypicked` -> `merged=no ahead=1 cherry+=0` (도달성은 놓치고 patch-id가 잡음)
- `feat/squashed` -> `merged=no ahead=2 cherry+=2` (**두 층 모두 놓침**)
- `feat/not-merged`, `feat/edited-pick` -> `merged=no ahead=1 cherry+=1`

`feat/squashed`의 `cherry+`가 `0`으로 나오면 squash가 커밋 하나만 합친 것이므로, 12.1의 (3)에서 `C1`과 `C2`를 모두 만들었는지 확인해야 합니다.

### 12.3 patch-id 직접 계산

```bash
pid() { git show "$1" | git patch-id --stable | cut -d' ' -f1 | cut -c1-12; }
echo "rebase 원본 : $(pid feat/rebased)"
echo "rebase 사본 : $(pid $(git log --format=%H -1 --grep='^D1$' master))"
echo "squash C1   : $(pid $(git log --format=%H -1 --grep='^C1$' feat/squashed))"
echo "squash C2   : $(pid $(git log --format=%H -1 --grep='^C2$' feat/squashed))"
echo "squash 커밋 : $(pid $(git log --format=%H -1 --grep='squash' master))"
```

PASS 기준은 rebase 두 줄이 같은 값, squash 세 줄이 서로 다른 값입니다. rebase 두 줄이 다르게 나오면 rebase 과정에서 충돌 해소로 내용이 바뀐 것입니다.

### 12.4 셸별 형태

`--is-ancestor`는 종료 코드로만 답하므로 셸마다 다릅니다.

Bash / Zsh:

```bash
git merge-base --is-ancestor origin/feat/x master && echo MERGED || echo NOT
```

PowerShell (`&&`는 되지만 `-v` 같은 단문자 옵션은 확장에 걸립니다):

```powershell
git merge-base --is-ancestor origin/feat/x master; if ($LASTEXITCODE -eq 0) { "MERGED" } else { "NOT" }
```

Nushell (`&&` 없음):

```nu
git merge-base --is-ancestor origin/feat/x master | complete | get exit_code
```

셸 의존을 피하려면 종료 코드 대신 숫자를 내는 쪽을 씁니다. 어느 셸에서든 같습니다.

```nu
git rev-list --count "master..origin/feat/x"
```

Nushell에서 `..`와 `...`를 따옴표로 감싸는 것은 범위 연산자로 해석되는 것을 막기 위해서입니다.

## 13. 실무 판정 순서

비용이 낮고 잡는 범위가 좁은 것부터 올라갑니다.

1. **전체를 한 번 훑습니다.**

    ```bash
    git branch -r --merged master
    ```

    여기 나온 브랜치는 도달성이 성립하므로 더 볼 것이 없습니다. 이력이 그대로 master에 있고 브랜치 ref만 남은 상태입니다.

2. **목록에 없는 브랜치만 patch-id로 재봅니다.**

    ```bash
    git cherry --verbose master origin/feat/x
    ```

    전부 `-`면 rebase나 cherry-pick으로 반영된 것입니다. `+`가 있으면 다음 단계로 갑니다.

3. **`+`가 남은 브랜치는 파일 내용으로 확인합니다.**

    ```bash
    git diff --stat master origin/feat/x -- $(git diff --name-only "master...origin/feat/x")
    ```

    출력이 비면 squash로 반영된 경우입니다. 차이가 남으면 그 차이가 실제 미반영분입니다.

4. **공통 조상이 없으면 앞 단계들이 무의미하므로 tree 비교로 갑니다.**

    ```bash
    git merge-base master origin/feat/x
    ```

    실패하면 `git diff master origin/feat/x`만이 의미 있는 질문입니다.

5. **지우기 전에는 로컬 브랜치의 안전장치를 씁니다.**

    ```bash
    git branch -d feat/x
    ```

    소문자 `-d`는 머지되지 않은 브랜치를 거부합니다. 다만 이 검사도 **도달성만** 보므로 squash로 반영된 브랜치는 거부합니다. 3번으로 확인한 뒤에는 `-D`로 지워야 하는데, 그때는 검사가 없으므로 3번을 실제로 통과했는지가 유일한 근거입니다.

브랜치 ref를 지워도 커밋은 사라지지 않습니다. 2장에서 본 대로 ref는 SHA 한 줄이 든 파일이고, master가 그 커밋들을 계속 가리키는 동안에는 정리 대상이 되지 않습니다. 잃는 것은 "이 작업 묶음의 이름"이며, 일반 merge로 들어온 경우에는 그 이름조차 `Merge branch 'feat/x'` 커밋 메시지에 남습니다.

## 출처

1차 출처를 우선합니다.

- 이 문서의 모든 판정 결과, patch-id 값, 커밋 객체 출력은 12장 스크립트로 만든 저장소에서 직접 실행해 얻었습니다. 환경은 `git version 2.54.0.windows.1`입니다.
- `git-cherry.adoc` — 동등성 판정 기준(diff 기반, 공백·줄번호 제거)과 탐지 대상(`git-cherry-pick`, `git-am`, `git-rebase`), `+`/`-` 출력 규약. Git for Windows 설치본의 `mingw64/share/doc/git-doc/git-cherry.adoc`.
- `git-patch-id.adoc` — patch ID의 정의("파일 diff들의 SHA-1 합, 줄 번호 무시")와 안정성·유일성 성격. 같은 디렉터리.
- `git-merge-base.adoc` — `--is-ancestor`의 종료 코드 규약(0=조상, 1=아님, 그 외=오류). 같은 디렉터리.
- PowerShell의 `-v` 확장과 Nushell의 `&&` 미지원은 각 셸이 출력한 오류 메시지를 그대로 인용했습니다.

추론이며 직접 관측하지 않은 것은 11.3의 revert 시나리오 하나입니다. 도달성의 정의에서 따라 나오는 결론이지만 실험으로 확인하지 않았습니다.

## 관련 노트

- `git/merge-base.md` — `merge-base` 자체의 동작과 조상 확인
- `git/git_merge.md` — 머지 전략과 옵션
- `git/cherry_pick.md` — cherry-pick 사용법과 옵션
- `git/git_rebase.md` — rebase가 커밋을 다시 적용하는 방식
- `git/internals.md` — 객체 저장소와 ref 구조
- `git/orphan.md` — 공통 조상이 없는 브랜치
- `git/verify_rebase_result.md` — blob 해시로 내용 동일성을 판정 (이 문서 8장의 3층에 해당)
- `git/rebase_replay_order.md` — 리베이스가 커밋을 재적용하는 순서와 갈래
- `git/rebase_lost_commit_recovery.md` — 리베이스 중 빈 커밋으로 사라진 작업 복원
