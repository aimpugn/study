# 리베이스 중 사라진 커밋을 되살리기 — 빈 커밋 드롭과 수술적 복원

## 목차

- [1. 증상 — 리베이스는 성공했는데 커밋이 하나 없다](#1-증상--리베이스는-성공했는데-커밋이-하나-없다)
- [2. 직접 원인 — 빈 커밋은 버려진다](#2-직접-원인--빈-커밋은-버려진다)
- [3. `--empty` 로는 막을 수 없다](#3---empty-로는-막을-수-없다)
- [4. 진단 — 정말 없는지, 어디에 넣어야 하는지](#4-진단--정말-없는지-어디에-넣어야-하는지)
- [5. 원본을 cherry-pick 하지 않는 이유](#5-원본을-cherry-pick-하지-않는-이유)
- [6. 어디에 끼우나 — 자리가 결과를 바꾸지 않는 경우](#6-어디에-끼우나--자리가-결과를-바꾸지-않는-경우)
- [7. 되쓸 수 있는지 먼저 확인한다](#7-되쓸-수-있는지-먼저-확인한다)
- [8. 복원 절차](#8-복원-절차)
- [9. 검증](#9-검증)
- [10. 되돌리기](#10-되돌리기)
- [11. 한계와 주의](#11-한계와-주의)
- [출처](#출처)
- [관련 노트](#관련-노트)

## 1. 증상 — 리베이스는 성공했는데 커밋이 하나 없다

리베이스가 이렇게 끝났습니다.

```text
Successfully rebased and updated refs/heads/feature.
```

종료 코드도 0입니다. 그런데 커밋을 세 보면 9개를 올렸는데 8개만 있습니다.

```console
$ git rev-list --count master..feature
8
$ git log --oneline master..feature --grep='5 허용'
```

`코드 길이 5 허용`이라는 제목의 커밋이 브랜치에 하나도 없습니다. 검색 결과가 비었습니다.

**git은 이 사실을 한 줄도 알려 주지 않았습니다.** 경고도, 요약도, 0이 아닌 종료 코드도 없습니다. 그래서 커밋 수를 세 보지 않으면 모르고 지나갑니다.

이 문서는 그렇게 사라진 작업을 되살리는 절차입니다. 무엇이 사라졌는지 확인하는 단계, 원본을 그대로 가져오면 안 되는 이유, 어디에 끼울지 정하는 근거, 그리고 되쓰기 전 안전 점검까지 다룹니다. 실험 저장소와 판정 명령은 [`git/verify_rebase_result.md`](verify_rebase_result.md) 2장의 것을 그대로 씁니다.

## 2. 직접 원인 — 빈 커밋은 버려진다

원인은 충돌 해소 방식입니다. 그 커밋과 새 기반이 같은 줄을 고쳤고, 충돌에서 **새 기반 쪽을 택했습니다.**

```bash
git checkout --ours -- src/app/validator.py
git add src/app/validator.py
git rebase --continue
```

리베이스 중 `--ours`는 이미 적용된 쪽, 곧 새 기반인 master 쪽입니다. 얹히려던 내 커밋이 `--theirs`입니다. 머지할 때의 감각과 반대라 헷갈리는 지점입니다.

master 쪽을 택했으니 그 커밋이 만들려던 변경이 하나도 남지 않았습니다. 남길 것이 없는 커밋은 빈 커밋이고, git은 빈 커밋을 버립니다. 문서가 그 기본값을 명시합니다.

```text
--empty=(drop|keep|stop)
    How to handle commits that are not empty to start and are not
    clean cherry-picks of any upstream commit, but which become
    empty after rebasing ...

    drop
        The commit will be dropped. This is the default behavior.
```

여기까지는 문서대로입니다. 그런데 같은 문서가 이어서 이렇게 적습니다.

```text
    stop
        The rebase will halt when the commit is applied ...
        This option is implied when `-i`/`--interactive` is specified.
```

`-i`를 쓰면 멈춘다고 읽힙니다. 그러면 대화형으로 하면 예방되는 것처럼 보입니다. **실제로는 그렇지 않습니다.**

## 3. `--empty` 로는 막을 수 없다

같은 상황을 세 가지로 돌려 봤습니다. 결과가 모두 같습니다.

| 실행 | 결과 커밋 수 | 멈췄나 | 알렸나 |
|---|---|---|---|
| `git rebase master` | 8 | 아니오 | 아니오 |
| `git rebase -i master` | 8 | 아니오 | 아니오 |
| `git rebase --empty=stop master` | 8 | 아니오 | 아니오 |

`--empty=stop`을 명시해도 멈추지 않습니다. 세 경우 모두 `Successfully rebased`만 출력하고 종료 코드 0으로 끝났습니다.

문서를 다시 읽으면 이유가 보입니다. `--empty`가 다루는 대상은 **"become empty after rebasing (because they contain a subset of already upstream changes)"**, 곧 *내용이 이미 상위 브랜치에 있어서* 재적용 과정에서 저절로 비게 된 커밋입니다. 우리 경우는 다릅니다. **사람이 충돌을 해소하면서 그 커밋의 변경을 손으로 버렸습니다.** 그건 `--empty`의 관할이 아닙니다.

그래서 실무적 결론은 이렇습니다. **옵션으로 예방할 수 없고, 끝난 뒤 세어 보는 것이 유일한 방어입니다.** 리베이스 직후에 커밋 수와 제목 집합을 대조하는 습관이 이 문서를 안 읽게 해 주는 유일한 장치입니다. 세는 방법은 [`git/verify_rebase_result.md`](verify_rebase_result.md) 10~11장에 있습니다.

## 4. 진단 — 정말 없는지, 어디에 넣어야 하는지

되살리기 전에 세 가지를 확인합니다. 각각이 뒤 단계의 결정을 바꿉니다.

**첫째, 정말 없는지.** 제목으로 찾습니다.

```console
$ git log --oneline master..feature --grep='5 허용'
```

출력이 비면 브랜치에 그 제목의 커밋이 없습니다. 사본이 둘 있었다면 둘 다 사라졌는지도 이렇게 확인합니다.

**둘째, 그 파일의 브랜치 이력이 어디까지인지.**

```console
$ git log --oneline master..feature -- src/app/validator.py
```

역시 비었습니다. 리베이스 구간 8커밋 중 이 파일을 건드리는 커밋이 **하나도 없습니다.** 6장에서 이 사실이 결정적으로 쓰입니다.

**셋째, 그 파일을 마지막으로 바꾼 커밋이 공유 이력인지.**

```console
$ LAST=$(git log --format='%h' -1 feature -- src/app/validator.py)
$ echo $LAST
d46d180
$ git branch -a --contains $LAST
* feature
  master
  remotes/origin/master
```

출력에 `master`와 `origin/master`가 있습니다. **이미 공유된 이력**이라는 뜻입니다. 그래서 이 커밋에 `--fixup`이나 `--squash`로 녹여 넣으면 안 됩니다. 남이 이미 가진 커밋의 SHA를 바꾸는 일이 되니까요.

`git branch -a --contains <커밋>`은 그 커밋을 조상으로 갖는 모든 브랜치를 나열합니다. `-a`는 원격 추적 브랜치까지 포함합니다. **`-a`를 빼면 로컬만 보므로, 원격에만 있는 공유 사실을 놓칩니다.**

세 확인을 합치면 결론이 나옵니다. 흡수할 커밋이 없으니 **새로 만들어 끼워야** 하고, 파일 이력이 비어 있으니 **자리는 비교적 자유**롭습니다.

## 5. 원본을 cherry-pick 하지 않는 이유

가장 먼저 떠오르는 방법은 원본 커밋을 cherry-pick하는 것입니다. 그러면 안 됩니다. 원본 커밋의 stat을 보면 이유가 보입니다.

```console
$ git show --stat --format='' feature-orig
 src/app/validator.py | 18 +++++++++---------
 1 file changed, 9 insertions(+), 9 deletions(-)
```

9줄이 바뀐 것처럼 보입니다. 공백을 무시하면 달라집니다.

```console
$ git show -w --format='' feature-orig
@@ -7,5 +7,5 @@ def check_code(code):
     # rule 6
     # rule 7
     # rule 8
-    if len(code) not in (7, 8):
+    if len(code) not in (5, 7, 8):
         raise ValueError(code)
```

**실제 변경은 딱 한 줄**이고, 나머지 8줄은 행 끝 공백 제거입니다. 그 8줄을 함께 가져오면 두 가지가 생깁니다.

첫째, **원래 났던 충돌을 다시 만듭니다.** 공백 hunk의 컨텍스트가 새 기반의 변경과 어긋납니다. 실제로 2장의 리베이스에서 이렇게 됐습니다.

```console
$ git diff        # 충돌 중
@@@ -1,11 -1,11 +1,23 @@@
  def check_code(code):
++<<<<<<< HEAD
 +    # rule 1
 +    # rule 2
 ... (8줄)
 +    if len(code) not in (7, 8) or not code.isdigit():
++=======
+     # rule 1
+     # rule 2
+ ... (8줄)
+     if len(code) not in (5, 7, 8):
++>>>>>>> beeebbc (코드 길이 5 허용)
          raise ValueError(code)
```

가운데 `rule` 줄들은 자리를 줄여 적었고 실제 출력에는 여덟 줄이 다 나옵니다. 그리고 화면에서는 안 보이지만 **`HEAD` 쪽 여덟 줄은 끝에 공백 세 칸이 붙어 있고 아래쪽 여덟 줄은 붙어 있지 않습니다.** 눈에 안 보이는 그 차이가 이 hunk를 만든 것입니다.

master는 `rule` 줄을 건드린 적이 없습니다. 그런데도 여덟 줄이 통째로 충돌 블록에 끌려 들어왔습니다. 공백 제거가 그 줄들을 변경으로 만들어 놓았고, 실제로 충돌하는 마지막 한 줄과 붙어 있어 하나의 hunk로 묶인 탓입니다.

둘째, **통합 diff가 오염됩니다.** 해소하고 나면 기능을 하나도 바꾸지 않는 8줄이 결과 diff에 섞입니다. 나중에 이 브랜치를 리뷰하는 사람이 무엇이 통합 작업이고 무엇이 잡음인지 구별할 수 없습니다.

그래서 **실제 변경 한 줄만 담은 커밋을 원본의 제목·작성자·날짜로 다시 만듭니다.** 원본을 복사하는 것이 아니라 원본이 의도한 것을 다시 진술하는 것입니다.

## 6. 어디에 끼우나 — 자리가 결과를 바꾸지 않는 경우

4장 둘째 확인에서 **리베이스 구간의 어떤 커밋도 그 파일을 건드리지 않는다**는 것을 봤습니다. blob 해시로도 같은 답이 나옵니다.

```console
$ FIRST=$(git rev-list master..feature | tail -1)
$ git rev-list ${FIRST}^..feature | while read c; do git rev-parse "$c:src/app/validator.py"; done | sort -u | wc -l
1
```

종류가 1, 곧 8커밋 내내 같은 내용입니다. 이것이 뜻하는 바가 중요합니다. **어느 자리에 끼워도 뒤 커밋들이 그 파일을 만지지 않으니 충돌 없이 재적용되고, 최종 트리도 자리와 무관하게 같습니다.**

주장으로 두지 않고 확인했습니다. 같은 복원을 세 번째 pick 뒤와 여덟 번째 pick 뒤에 각각 넣고 결과 트리를 비교했습니다.

```console
$ git rev-list --count master..fix-mid
9
$ git rev-list --count master..fix-end
9
$ git rev-parse fix-mid^{tree}
a23d31fb7b105fcedbd5dcd97a63237b6f259725
$ git rev-parse fix-end^{tree}
a23d31fb7b105fcedbd5dcd97a63237b6f259725
```

**트리 해시가 같습니다.** `<커밋>^{tree}`는 그 커밋이 가리키는 최상위 tree 객체이고, tree 해시가 같다는 것은 저장소 전체 내용이 한 바이트도 다르지 않다는 뜻입니다.

두 브랜치의 커밋 순서는 실제로 다릅니다. 복원한 커밋이 한쪽은 네 번째, 한쪽은 맨 위입니다.

```console
$ git log --oneline master..fix-mid | tail -7
751ce78 parser 헬퍼 1 추가
db165d2 문서 보강 5
801ed3b 문서 보강 4
00f5469 코드 길이 5 허용
ee69273 문서 보강 3
3df4f3b 문서 보강 2
b9a4e13 문서 보강 1
$ git log --oneline master..fix-end | head -3
6507c5d 코드 길이 5 허용
2d275a6 parser 헬퍼 3 추가
abd6964 parser 헬퍼 2 추가
```

두 출력의 SHA를 견줘 보면 한 가지가 더 보입니다. `fix-mid` 쪽의 `문서 보강 4`·`문서 보강 5`·`parser 헬퍼 1` 은 `801ed3b`·`db165d2`·`751ce78` 인데, 끼우지 않은 원래 브랜치에서는 `9fb838b`·`618933e`·`2ffea74` 였습니다. **중간에 커밋을 넣으면 그 뒤 커밋의 SHA가 전부 다시 쓰입니다.** 커밋 SHA에 부모가 들어가기 때문입니다. 그래서 7장의 되쓰기 점검을 통과하지 못한 브랜치에서는 이 작업 자체를 하면 안 됩니다.

**이력은 다르고 최종 내용은 같습니다.** 그래서 자리 선택은 "무엇이 맞는가"가 아니라 "이력을 읽는 사람에게 무엇이 자연스러운가"의 문제입니다. 원본 계보에서 그 커밋 앞에 있던 작업의 리베이스 사본을 찾아 그 뒤에 넣으면, 나중에 이력을 읽는 사람에게 처음부터 그 자리에 있던 작업으로 보입니다.

**반대로 파일 이력이 비어 있지 않으면 이 절의 전제가 깨집니다.** 종류가 2 이상이면 그 파일을 건드리는 커밋이 있다는 뜻이고, 그때는 끼우는 자리에 따라 뒤 커밋이 충돌합니다. 그 경우에는 그 파일을 마지막으로 건드린 커밋 바로 뒤가 가장 안전한 자리입니다.

## 7. 되쓸 수 있는지 먼저 확인한다

복원은 이력을 다시 쓰는 작업입니다. 남이 그 이력을 이미 가져갔다면 되쓰면 안 됩니다. 두 곳을 봅니다.

```console
$ git for-each-ref --format='%(refname:short) upstream=[%(upstream:short)]' refs/heads/feature
feature upstream=[]
$ git ls-remote --heads origin feature
$ echo $?
0
```

첫째는 로컬 설정(그 브랜치에 upstream이 지정돼 있는지), 둘째는 원격의 실제 상태입니다. **둘을 다 봐야 합니다.** 설정만 보면 `--set-upstream` 없이 손으로 push한 브랜치를 놓칩니다. 실측으로 확인한 것입니다.

아래는 그 사실을 보이기 위한 시연입니다. **실험 저장소에서만 하십시오.** 브랜치를 원격에 올리는 순간 그 이력은 공유된 것이 되고, 이 문서가 하려는 되쓰기를 해서는 안 되는 상태가 됩니다. 시연 뒤에는 `git push origin --delete feature` 로 되돌려 놓습니다.

```console
$ git push origin feature                    # upstream 설정 없이 push
$ git for-each-ref --format='%(upstream:short)' refs/heads/feature
                                             # 여전히 비어 있다
$ git ls-remote --heads origin feature
2d275a60ab8fdba8ad1e172e33a07fc61d85d145        refs/heads/feature
```

`for-each-ref`는 비었는데 `ls-remote`는 브랜치를 찾아냅니다. 설정만 믿었다면 "로컬 전용"이라고 오판했을 상황입니다. (위 출력에서 SHA와 ref 이름 사이는 실제로는 탭 문자 하나이고, 여기서는 읽기 편하도록 공백으로 펴 두었습니다. 스크립트에서 자를 때는 탭을 구분자로 삼아야 합니다.)

`ls-remote`의 출력을 읽을 때 세 경우를 구분해야 합니다. 실측값입니다.

| 상황 | 출력 | 종료 코드 |
|---|---|---|
| 원격이 설정돼 있지 않음 | `fatal: 'origin' does not appear to be a git repository` | 128 |
| 원격은 있고 그 브랜치가 없음 | (빈 출력) | 0 |
| 원격에 그 브랜치가 있음 | `<sha>` 와 `refs/heads/<브랜치>` (사이는 탭 문자) | 0 |

**"출력이 비면 로컬 전용"으로 읽으면 첫째 경우를 놓칩니다.** 원격 설정 자체가 없어 조회에 실패한 것과 조회에 성공했는데 없는 것은 다릅니다. 종료 코드를 함께 봐야 갈립니다.

작업트리가 깨끗한지도 확인합니다.

```bash
git status --short      # 출력이 비어야 함
```

## 8. 복원 절차

이력을 되쓸 수 있고, 자리를 정했고, 무엇을 만들지 정했습니다. 이제 그 자리에 멈춰 커밋을 하나 끼웁니다.

`break`는 대화형 리베이스의 todo 명령으로, **그 지점에서 리베이스를 멈춥니다.** 멈춘 상태에서 원하는 커밋을 만들고 `--continue`하면 뒤 커밋들이 그 위에 재적용됩니다.

아래 SHA를 그대로 재현하려면 [`git/verify_rebase_result.md`](verify_rebase_result.md) 2장에서 쓴 committer 날짜 고정을 유지한 채 진행합니다.

```bash
export GIT_COMMITTER_DATE="2025-12-20 10:00:00 +0900"
```

```bash
git rebase -i master
```

편집기가 열리면 멈추고 싶은 `pick` 줄 **다음**에 `break` 한 줄을 넣고 저장합니다. 세 번째 커밋 뒤에서 멈추려면 이렇게 됩니다.

```text
pick b9a4e13 문서 보강 1
pick 3df4f3b 문서 보강 2
pick ee69273 문서 보강 3
break
pick 9fb838b 문서 보강 4
...
```

리베이스가 그 지점에서 멈춥니다. 거기서 실제 변경 한 줄만 만들고 커밋합니다.

```bash
sed -i 's/not in (7, 8)/not in (5, 7, 8)/' src/app/validator.py
git add src/app/validator.py
git commit -m "코드 길이 5 허용" \
  --author="bob <bob@example.com>" \
  --date="2025-12-15 13:14:48 +0900"
git rebase --continue
```

`--author`와 `--date`는 원본 커밋에서 그대로 가져온 값입니다. 원본에서 뽑는 방법은 이렇습니다.

```bash
git log -1 --format='%an <%ae>%n%ad' <원본커밋>
```

`--date`는 author 날짜만 설정합니다. committer 날짜까지 맞추려면 환경 변수를 함께 씁니다.

```bash
GIT_COMMITTER_DATE="2025-12-15 13:14:48 +0900" git commit -m "..." --author="..." --date="..."
```

이렇게 준 committer 날짜는 **`--continue` 를 지나도 그대로 남습니다.** 실측으로 확인했습니다 — `break` 지점에서 만든 커밋은 이미 새 기반 위에 있으므로 재적용 대상이 아니고, `--continue` 는 그 뒤의 `pick` 만 처리합니다.

그러니 맞추고 싶으면 맞출 수 있습니다. **다만 맞추는 것이 항상 나은 선택은 아닙니다.** author 날짜는 "그 변경을 언제 작성했나"이고 committer 날짜는 "이 커밋 객체를 언제 만들었나"입니다. 복원 작업을 오늘 했다면 committer 날짜가 오늘인 것이 사실에 맞고, 나중에 "이 커밋이 언제 이 브랜치에 들어왔나"를 추적할 단서가 됩니다. author 날짜만 원본으로 두고 committer 날짜는 건드리지 않는 편을 기본으로 두겠습니다.

편집기를 열지 않고 스크립트로 처리하려면 시퀀스 편집기를 지정합니다. 세 번째 줄 뒤에 `break`를 넣는 형태입니다.

```bash
GIT_SEQUENCE_EDITOR="sed -i '3a break'" git rebase -i master
```

이 방식은 자동화에 편하지만 **todo 목록을 눈으로 보지 않고 줄 번호를 믿는 것**이라, 목록을 먼저 확인하지 않으면 엉뚱한 자리에 멈춥니다. `3a` 가 세 번째 `pick` 뒤를 뜻하는 것은 todo 파일이 첫 줄부터 `pick` 으로 시작할 때뿐입니다(주석은 파일 끝에 붙습니다). 목록만 보려면 이렇게 합니다.

```bash
GIT_SEQUENCE_EDITOR='cat' git rebase -i master
```

## 9. 검증

복원이 끝나면 세 가지를 셉니다.

```console
$ git rev-list --count master..feature
9
```

8이 9가 됐습니다. 사라진 하나가 돌아왔습니다.

```console
$ git log --format='%h | %an <%ae> | %ad | %s' --date=iso feature --grep='5 허용'
00f5469 | bob <bob@example.com> | 2025-12-15 13:14:48 +0900 | 코드 길이 5 허용
```

작성자와 날짜가 원본 값으로 들어갔습니다.

```console
$ git diff --stat feature feature-orig -- src
 src/app/validator.py | 18 +++++++++---------
 1 file changed, 9 insertions(+), 9 deletions(-)
$ git diff -w --stat feature feature-orig -- src
 src/app/validator.py | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
```

원본과 비교하면 차이가 남아 있습니다. **이것이 정상입니다.** 두 가지가 겹쳐 있습니다.

1. 공백 8줄 — 5장에서 의도적으로 가져오지 않은 것입니다. `-w`를 통과시키면 사라집니다.
2. 남은 한 줄 — 새 기반이 스스로 추가한 변경입니다. 원본 브랜치는 그 변경보다 앞선 시점의 것이니 당연히 없습니다.

```console
$ git show feature:src/app/validator.py | grep 'len(code)'
    if len(code) not in (5, 7, 8) or not code.isdigit():
$ git show feature-orig:src/app/validator.py | grep 'len(code)'
    if len(code) not in (5, 7, 8):
```

`(5, 7, 8)`은 복원됐고, `or not code.isdigit()`는 새 기반의 것입니다. 둘이 합쳐진 상태가 우리가 원한 결과입니다. **"원본과 diff가 0"을 목표로 삼으면 안 됩니다.** 목표는 "복원하려던 변경이 들어갔고, 새 기반의 변경도 살아 있다"입니다.

## 10. 되돌리기

리베이스 전 위치가 reflog에 남아 있습니다. 이 문서의 복원까지 마친 뒤의 실제 출력입니다.

```console
$ git reflog show feature
ced4d66 feature@{0}: rebase (finish): refs/heads/feature onto fbbaab82dcb3c14e73c103e71fe676a5a6e38d82
2d275a6 feature@{1}: rebase (finish): refs/heads/feature onto fbbaab82dcb3c14e73c103e71fe676a5a6e38d82
beeebbc feature@{2}: commit: 코드 길이 5 허용
9bab2f1 feature@{3}: commit: parser 헬퍼 3 추가
```

`@{0}`이 복원 리베이스의 결과이고, `@{1}`이 그 직전 위치, 곧 8커밋이던 상태입니다. 그 위치로 되돌리면 복원을 취소할 수 있습니다.

```bash
git reset --hard feature@{1}
```

**`@{1}`을 기계적으로 쓰지 마십시오.** 위 출력에 `rebase (finish)`가 **두 줄** 있는 것을 보십시오. `@{1}`은 2장의 첫 리베이스 결과이고, 그보다 더 뒤로 가려면 `@{2}`(원본 브랜치의 tip)입니다. `@{n}`은 "그 ref의 n번째 이전 reflog 항목"일 뿐이므로, 중간에 무슨 조작이 있었는지에 따라 가리키는 곳이 달라집니다.

그래서 순서는 이렇습니다. 먼저 읽고, 눈으로 확인한 SHA로 리셋합니다.

```bash
git reflog show feature          # 먼저 읽는다
git reset --hard 2d275a6         # 확인한 SHA 로 되돌린다
```

`--hard`는 작업트리까지 되돌리므로 커밋하지 않은 변경이 사라집니다. 남길 것이 있으면 `git stash`를 먼저 하거나 `--keep`을 씁니다.

reflog 항목은 영구히 남지 않습니다. 일반 항목은 `gc.reflogExpire` 기본값 90일이지만, **`git rebase`나 `git commit --amend` 로 생긴 항목은 `gc.reflogExpireUnreachable` 의 관할이고 기본값이 30일입니다.** git 문서가 그 이유를 직접 적어 둡니다 — 그런 항목은 현재 프로젝트의 일부가 아니니 더 빨리 지우는 편이 낫다는 것입니다. 그러니 이 되돌리기는 **바로 쓸 때 확실하고 한 달이 지나면 보장되지 않습니다.**

## 11. 한계와 주의

- **`--empty` 옵션으로 예방할 수 없습니다.** 3장의 실측입니다. 예방책은 리베이스 직후 커밋 수를 세는 것뿐입니다.
- **6장의 "자리 무관"은 조건부입니다.** 리베이스 구간에서 그 파일을 건드리는 커밋이 하나도 없을 때만 성립합니다. 조건을 확인하지 않고 아무 자리에 끼우면 뒤 커밋들이 충돌합니다.
- **`--fixup`/`--squash`로 녹이는 것은 대상 커밋이 공유 이력이 아닐 때만 됩니다.** 4장 셋째 확인이 그 판정입니다.
- **사라진 커밋이 둘 이상일 수 있습니다.** 원본에 중복 사본이 있었다면 둘 다 빈 커밋이 됩니다. 그 경우 복원은 하나만 하면 되지만, 제목 집합 비교로 개수를 세면 원본 2개가 결과 0개가 된 것을 알아야 합니다. 집합 비교가 중복을 접는 문제는 [`git/verify_rebase_result.md`](verify_rebase_result.md) 10장에 있습니다.
- **이 문서의 복원은 이력 되쓰기입니다.** 7장의 두 확인을 통과하지 못하면 하지 마십시오. 공유된 이력을 되쓰면 그 이력을 가진 다른 사람의 저장소가 어긋납니다.
- **`break` 없이 `edit`를 쓰는 방법도 있습니다.** `edit`는 그 커밋을 적용한 뒤 멈추므로 `--amend`로 그 커밋 자체를 고칠 때 맞고, 새 커밋을 별도로 끼우려면 `break`가 의도를 더 정확히 드러냅니다.

## 출처

- 이 문서의 모든 출력값, 해시, 커밋 수, 종료 코드는 [`git/verify_rebase_result.md`](verify_rebase_result.md) 2장 스크립트로 만든 저장소에서 직접 실행해 얻었습니다. 환경은 `git version 2.54.0.windows.1`입니다.
- `git-rebase.adoc:259` 이하 — `--empty=(drop|keep|stop)` 의 정의와 대상 범위("become empty after rebasing (because they contain a subset of already upstream changes)"), `drop` 이 기본값이라는 것, `stop` 이 `-i` 에서 함의된다는 것. Git for Windows 설치본의 `mingw64/share/doc/git-doc/`.
- `git-rebase.adoc` INTERACTIVE MODE 절 — `break` 명령의 의미. 같은 디렉터리.
- `git-config.html` 의 `gc.reflogExpire`·`gc.reflogExpireUnreachable` 항목 — 기본값 90일과 30일, 그리고 후자가 "generally created as a result of using git commit --amend or git rebase" 인 항목을 대상으로 한다는 설명. 같은 디렉터리.
- **문서 근거가 아니라 실측**: 3장의 표. `-i` 와 `--empty=stop` 이 이 경우에 멈추지 않는다는 것은 세 번 실행해 확인한 결과이고, 문서가 `--empty` 의 대상을 "이미 상위에 있는 변경의 부분집합" 으로 한정한 것과 부합하는 해석입니다. 다만 문서가 "사람이 충돌 해소로 비운 커밋" 을 명시적으로 제외한다고 적어 둔 것은 아니므로, 이 해석은 실측에 근거한 추론입니다.
- 6장의 트리 해시 일치는 두 위치에 각각 복원해 `rev-parse ^{tree}` 로 비교한 결과입니다.

## 관련 노트

- [`git/verify_rebase_result.md`](verify_rebase_result.md) — 무엇이 사라졌는지 판정하는 명령들, 실험 저장소
- [`git/rebase_replay_order.md`](rebase_replay_order.md) — 왜 그 충돌이 났는지, 갈래와 적용 순서
- [`git/rebase_cases.md`](rebase_cases.md) — 리베이스 사례 모음
- [`git/git_rebase.md`](git_rebase.md) — rebase 명령과 옵션 전반
- [`git/reflog.md`](reflog.md) — 10장에서 쓴 reflog
- [`git/cherry_pick.md`](cherry_pick.md) — 5장에서 쓰지 않기로 한 cherry-pick
- [`git/conflict.md`](conflict.md) — 충돌 해소 전반
