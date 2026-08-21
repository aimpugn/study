# 리베이스 결과가 원본과 같은지 어떻게 증명하는가 — blob 해시로 판정하기

## 목차

- [1. 시작 질문 — 커밋마다 diff를 돌려야 하나](#1-시작-질문--커밋마다-diff를-돌려야-하나)
- [2. 실험 저장소](#2-실험-저장소)
- [3. 구간을 먼저 확정한다](#3-구간을-먼저-확정한다)
- [4. 그 시점에 파일이 있었나](#4-그-시점에-파일이-있었나)
- [5. 핵심 — blob 해시는 내용의 지문이다](#5-핵심--blob-해시는-내용의-지문이다)
- [6. 구간에서 몇 번 바뀌었나](#6-구간에서-몇-번-바뀌었나)
- [7. 다른 각도로 같은 사실을 확인한다](#7-다른-각도로-같은-사실을-확인한다)
- [8. 공백만 바꾼 커밋 골라내기](#8-공백만-바꾼-커밋-골라내기)
- [9. 두 결과를 맞대 보기](#9-두-결과를-맞대-보기)
- [10. 제목으로 세는 방법과 그 함정](#10-제목으로-세는-방법과-그-함정)
- [11. 선형인지 확인하기](#11-선형인지-확인하기)
- [12. 셸별 형태](#12-셸별-형태)
- [13. 판정표](#13-판정표)
- [출처](#출처)
- [관련 노트](#관련-노트)

## 1. 시작 질문 — 커밋마다 diff를 돌려야 하나

리베이스로 커밋 수십 개를 새 기반 위에 올렸습니다. 그 구간에서 어떤 파일이 한 번이라도 바뀌었는지 알아야 합니다. 순진한 방법은 커밋마다 `git diff`를 돌려 그 파일이 나오는지 보는 것입니다. 커밋 수만큼 돌리고, 그만큼의 출력을 눈으로 읽고, 눈이 놓친 것은 놓친 채로 결론을 내립니다.

그럴 필요가 없습니다. **git은 파일 내용을 blob 객체로 저장하고 그 해시를 이름으로 씁니다. 내용이 한 바이트라도 다르면 해시가 다릅니다.** 그래서 구간의 커밋에서 그 파일의 blob 해시만 뽑아 종류를 세면 됩니다. 종류가 1이면 그 구간에서 한 번도 바뀌지 않았다는 뜻입니다. diff를 한 번도 돌리지 않고, 눈으로 읽는 판단이 끼어들 자리도 없습니다.

이 문서는 그 판정을 포함해, "리베이스 결과가 내가 의도한 것인지"를 확인하는 명령들을 다룹니다. 각 명령이 무엇을 근거로 답하는지, 어디서 틀리는지까지 봅니다. 아래 모든 출력값은 2장의 스크립트로 만든 저장소에서 실제로 실행해 얻은 것입니다. 실험 저장소는 손으로 따라갈 수 있도록 커밋 9개로 줄였지만, 세는 방법은 구간이 몇 개든 같습니다.

## 2. 실험 저장소

먼저 저장소를 만듭니다. 이후 모든 절의 출력값을 이 저장소에서 재현할 수 있습니다.

커밋 SHA는 커밋 객체 전체의 해시이고 그 안에 author·committer 날짜가 들어갑니다. 그래서 날짜를 고정하지 않으면 돌릴 때마다 SHA가 달라져 이 문서의 값과 대조할 수 없습니다. 아래 스크립트가 모든 커밋의 두 날짜를 고정하는 이유입니다. **고정했으므로 이 문서에 적힌 SHA가 그대로 재현됩니다.**

```bash
#!/usr/bin/env bash
set -eu
rm -rf expA && mkdir expA && cd expA

git init -q -b master
git config user.name  "alice"
git config user.email "alice@example.com"
git config core.autocrlf false   # 8장의 공백 실험이 줄끝 변환에 오염되지 않게

D="2025-11-20 10:00:00 +0900"
ci() { GIT_AUTHOR_DATE="$D" GIT_COMMITTER_DATE="$D" git commit -q "$@"; }

mkdir -p src/app

# 기반 커밋. rule 줄 끝에 공백 3칸을 일부러 남긴다 (8장 재료)
{
  echo 'def check_code(code):'
  for i in 1 2 3 4 5 6 7 8; do printf '    # rule %s   \n' "$i"; done
  echo '    if len(code) not in (7, 8):'
  echo '        raise ValueError(code)'
} > src/app/validator.py
printf 'def parse(line):\n    return line.split(",")\n' > src/app/parser.py
printf '# demo app\n' > README.md
git add -A
ci -m "초기 커밋: validator, parser, README"

# feature 브랜치: 기반에서 갈라져 9커밋
git checkout -q -b feature
for i in 1 2 3 4 5; do
  printf 'step %s\n' "$i" >> README.md
  ci -am "문서 보강 $i"
done
for i in 1 2 3; do                      # parser.py 를 3번 건드린다 (대조군)
  printf 'def helper%s():\n    pass\n' "$i" >> src/app/parser.py
  ci -am "parser 헬퍼 $i 추가"
done
sed -i 's/[[:space:]]*$//' src/app/validator.py           # 뒤 공백 제거 (8줄)
sed -i 's/not in (7, 8)/not in (5, 7, 8)/' src/app/validator.py   # 실제 변경 (1줄)
GIT_AUTHOR_DATE="2025-12-15 13:14:48 +0900" GIT_COMMITTER_DATE="2025-12-15 13:14:48 +0900" \
  git commit -q -am "코드 길이 5 허용" --author="bob <bob@example.com>"

# master 가 그 사이 전진. feature 와 "같은 줄"을 고쳐 충돌을 만든다
git checkout -q master
sed -i 's/not in (7, 8)/not in (7, 8) or not code.isdigit()/' src/app/validator.py
ci -am "숫자 여부 검사 추가"
mkdir -p docs && printf 'usage\n' > docs/usage.md
git add docs/usage.md && ci -m "사용법 문서 추가"

# 7장에서 쓸 원격
cd .. && rm -rf remote.git && git init -q --bare remote.git && cd expA
git remote add origin ../remote.git
git push -q origin master
```

여기까지 돌리면 `master`가 `fbbaab8`, `feature`가 `beeebbc`입니다. 다르게 나오면 날짜 고정이 빠진 것이니 위 스크립트를 다시 확인합니다.

이제 원본을 남겨 두고 리베이스합니다. 충돌은 **master 쪽을 택해** 해소합니다. 리베이스가 만드는 커밋의 committer 날짜도 고정해야 결과 SHA가 재현됩니다.

```bash
cd expA
export GIT_COMMITTER_DATE="2025-12-20 10:00:00 +0900"

git checkout -q feature
git branch -f feature-orig feature      # 원본 보존
git rebase master                       # validator.py 에서 충돌
git checkout --ours -- src/app/validator.py
git add src/app/validator.py
git rebase --continue
```

`--ours`는 리베이스 중에는 **이미 적용된 쪽**, 즉 새 기반인 master 쪽을 뜻합니다. 리베이스는 master를 체크아웃한 뒤 내 커밋을 그 위에 얹으므로, 얹히는 쪽이 `--theirs`가 됩니다. 머지할 때의 감각과 반대라 헷갈리는 지점입니다.

결과가 이렇습니다.

```text
Successfully rebased and updated refs/heads/feature.
```

그런데 커밋을 세 보면 9개가 아니라 8개입니다.

```console
$ git rev-list --count master..feature
8
$ git log --oneline master..feature
2d275a6 parser 헬퍼 3 추가
abd6964 parser 헬퍼 2 추가
2ffea74 parser 헬퍼 1 추가
618933e 문서 보강 5
9fb838b 문서 보강 4
ee69273 문서 보강 3
3df4f3b 문서 보강 2
b9a4e13 문서 보강 1
```

`코드 길이 5 허용`이 없습니다. 충돌에서 master 쪽을 택하니 그 커밋의 패치가 아무것도 남기지 않게 되어, git이 빈 커밋으로 보고 떨어뜨렸습니다. **`git rebase --continue`는 이 사실을 한 줄도 알려 주지 않았습니다.** 종료 코드도 0입니다. 이 침묵이 아래 판정들이 필요한 이유입니다. 커밋이 사라졌을 때 어떻게 되살리는지는 [`git/rebase_lost_commit_recovery.md`](rebase_lost_commit_recovery.md)에서 다룹니다.

## 3. 구간을 먼저 확정한다

무엇을 검사할지 정하지 않으면 판정 결과를 믿을 수 없습니다. "리베이스로 올린 커밋 전부"가 정확히 어디부터 어디까지인지 먼저 고정합니다.

`A..B`는 **B에서 부모를 거슬러 닿는데 A에서는 닿지 않는 커밋**을 뜻합니다. 그래서 `master..feature`는 feature에만 있는 커밋입니다. 이 표기에서 `A` 자신은 빠집니다.

구간의 첫 커밋을 잡고, 그 부모가 master tip인지 확인합니다.

```console
$ FIRST=$(git rev-list master..feature | tail -1)   # 목록은 최신순이므로 마지막이 가장 오래된 것
$ git rev-parse --short $FIRST
b9a4e13
$ git rev-parse --short ${FIRST}^
fbbaab8
$ git rev-parse --short master
fbbaab8
```

`^`는 첫 부모를 가리킵니다. 두 값이 같으므로 `b9a4e13`이 master 직후 첫 커밋이고, 구간 경계가 맞습니다. 만약 달랐다면 중간에 다른 커밋이 끼어 있다는 뜻이고, 그 상태로 아래 판정을 돌리면 세는 대상이 어긋납니다.

첫 커밋 자신을 포함해 세려면 왼쪽에 `FIRST^`를 둡니다.

```console
$ git rev-list --count ${FIRST}^..feature
8
```

`master..feature`와 같은 값이 나옵니다. 경계가 맞다는 것을 다시 확인해 주는 셈입니다.

## 4. 그 시점에 파일이 있었나

내용을 비교하기 전에 존재부터 확인합니다. `<커밋>:<경로>` 표기는 **그 커밋 시점의 그 경로**를 가리킵니다.

```console
$ git cat-file -e feature:src/app/validator.py
$ echo $?
0
```

`-e`는 존재 여부만 보고 종료 코드로 답합니다. **출력이 없는 것이 정상**입니다. 없으면 이렇습니다.

```console
$ git cat-file -e feature:src/app/nope.py
$ echo $?
128
```

여기서 한 가지 주의할 점이 있습니다. 없을 때의 종료 코드는 `1`이 아니라 **`128`**입니다. `if [ $? -eq 1 ]` 같은 조건을 쓰면 없는데도 없다고 판정하지 못합니다. 있음/없음만 가리려면 `0`인지만 봅니다.

## 5. 핵심 — blob 해시는 내용의 지문이다

`git rev-parse <커밋>:<경로>`는 그 시점 그 파일의 blob 해시를 출력합니다.

```console
$ git rev-parse master:src/app/validator.py
3387904af9e08819e3edfcde64160d904e64624d
$ git rev-parse feature:src/app/validator.py
3387904af9e08819e3edfcde64160d904e64624d
```

두 해시가 같습니다. 그래서 **리베이스된 feature의 validator.py는 master의 것과 내용이 완전히 같습니다.** diff를 돌리지 않았고, 출력을 눈으로 읽지도 않았습니다.

이것이 성립하는 이유는 git의 저장 방식입니다. git은 파일 내용을 blob 객체로 저장하고 그 내용의 해시를 객체 이름으로 씁니다. 같은 내용이면 저장소 어디에 있든 같은 blob 하나를 공유하고, 한 바이트라도 다르면 다른 blob이 됩니다. 그러니 **해시 일치는 내용 일치와 같은 말**입니다.

거꾸로도 유용합니다. 2장에서 `코드 길이 5 허용`이 사라진 것을 커밋 목록으로 알았지만, 해시 비교는 그 결과를 파일 수준에서 다시 확인해 줍니다. 그 커밋이 살아 있었다면 feature의 validator.py는 master의 것과 달랐을 테니까요.

## 6. 구간에서 몇 번 바뀌었나

해시를 구간 전체에서 모아 종류를 세면 변경 여부가 나옵니다.

```console
$ git rev-list ${FIRST}^..feature | while read c; do git rev-parse "$c:src/app/validator.py"; done | sort -u | wc -l
1
$ git rev-list ${FIRST}^..feature | while read c; do git rev-parse "$c:src/app/parser.py"; done | sort -u | wc -l
4
```

`validator.py`는 종류가 **1**입니다. 8개 커밋에서 뽑은 해시가 전부 같은 값이라는 뜻이고, 곧 **이 구간에서 이 파일은 한 번도 바뀌지 않았습니다.**

`parser.py`는 **4**입니다. 여기서 셈법을 정확히 잡아야 합니다. 이 값은 변경 횟수가 아니라 **구간 안에서 관측된 내용의 종류**입니다. 실험 저장소에서 parser.py는 커밋 3개에서 바뀌었고, 그 앞 5개 커밋은 기반 시점의 내용을 그대로 갖고 있습니다. 그래서 기반 상태 1종에 변경 결과 3종을 더해 4종입니다.

정리하면 **N번 바뀌었으면 종류는 N+1**이고, 바뀌지 않았으면 1입니다. "1이면 안 바뀜"만 기억하고 "3이면 3번 바뀜"으로 읽으면 하나씩 어긋납니다.

## 7. 다른 각도로 같은 사실을 확인한다

같은 질문에 다른 원리로 답하는 명령이 있습니다. `--` 뒤는 경로 제한자이고, 그 경로를 건드린 커밋만 남깁니다.

```console
$ git log --oneline ${FIRST}^..feature -- src/app/validator.py
$ git log --oneline ${FIRST}^..feature -- src/app/parser.py
2d275a6 parser 헬퍼 3 추가
abd6964 parser 헬퍼 2 추가
2ffea74 parser 헬퍼 1 추가
```

validator.py는 출력이 비었고, parser.py는 3줄입니다. 6장의 결과와 맞습니다 — validator는 종류 1(안 바뀜), parser는 종류 4(3번 바뀜).

두 명령을 다 돌리는 이유는 서로를 검산하기 때문입니다. 6장은 트리에 기록된 blob을 직접 비교하고, 7장은 git의 이력 순회가 그 경로를 변경점으로 잡는지 봅니다. **어긋나면 둘 중 한쪽 명령을 잘못 쓴 것**이니 그때는 결론을 내리지 말고 명령부터 다시 봅니다. 특히 `--` 를 빼먹으면 git이 그 인자를 경로가 아니라 리비전으로 해석하려 들어 전혀 다른 답이 나옵니다.

## 8. 공백만 바꾼 커밋 골라내기

리베이스 결과를 볼 때 "이 커밋은 열여덟 줄을 바꿨다"는 인상이 판단을 흐립니다. 실험 저장소의 원본 커밋을 보면 이렇습니다.

```console
$ git show --stat --format='' feature-orig
 src/app/validator.py | 18 +++++++++---------
 1 file changed, 9 insertions(+), 9 deletions(-)
```

9줄이 바뀐 것처럼 보입니다. `-w`를 붙이면 달라집니다.

```console
$ git show -w --stat --format='' feature-orig
 src/app/validator.py | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
```

`-w`는 모든 공백 차이를 무시합니다. 그러니 이 커밋의 **실제 변경은 한 줄**이고 나머지 8줄은 행 끝 공백 제거입니다. 내용을 보면 분명합니다.

```console
$ git show -w --format='' feature-orig
diff --git a/src/app/validator.py b/src/app/validator.py
index cdb13db..421c886 100644
--- a/src/app/validator.py
+++ b/src/app/validator.py
@@ -7,5 +7,5 @@ def check_code(code):
     # rule 6
     # rule 7
     # rule 8
-    if len(code) not in (7, 8):
+    if len(code) not in (5, 7, 8):
         raise ValueError(code)
```

`-w`를 붙인 것과 붙이지 않은 것을 나란히 비교하면 "공백만 바꾼 커밋"과 "공백 정리가 섞인 커밋"을 골라낼 수 있습니다. 둘의 stat이 같으면 공백 변경이 없고, 다르면 그 차이만큼이 공백입니다.

이것이 실무에서 왜 중요한가 하면, 공백 hunk가 **충돌 범위를 넓히기** 때문입니다. 2장의 리베이스에서 실제로 이렇게 됐습니다.

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

가운데 `rule` 줄들은 자리를 줄여 적었고, 실제 출력에는 여덟 줄이 모두 나옵니다. 그리고 화면에서는 안 보이지만 **`HEAD` 쪽 여덟 줄은 끝에 공백 세 칸이 붙어 있고 아래쪽 여덟 줄은 붙어 있지 않습니다.** 그 차이가 이 hunk의 정체입니다.

master는 `rule` 줄을 건드린 적이 없습니다. 그런데도 여덟 줄이 통째로 충돌 블록에 끌려 들어왔습니다. 공백 제거가 그 줄들을 변경으로 만들어 놓았고, 실제로 충돌하는 마지막 한 줄과 인접해 있어 하나의 hunk로 묶인 탓입니다.

그래서 이런 커밋을 다시 적용할 때는 원본을 그대로 가져오는 것보다 **실제 변경 한 줄만 담은 커밋을 다시 만드는 편**이 낫습니다. 그 절차는 [`git/rebase_lost_commit_recovery.md`](rebase_lost_commit_recovery.md)에 있습니다.

## 9. 두 결과를 맞대 보기

같은 작업을 두 번 리베이스했거나, 남이 한 결과와 내 결과를 비교할 때 씁니다. 커밋 SHA는 리베이스가 다시 쓰므로 대응이 안 되고, 볼 것은 최종 트리입니다.

```console
$ git diff --quiet feature feature-orig -- src
$ echo $?
1
```

`--quiet`는 출력 없이 종료 코드만 냅니다. `0`이면 차이 없음, `1`이면 차이 있음입니다. 스크립트에서 조건으로 쓰기 좋습니다.

어느 파일이 다른지는 `--stat`으로 바꿔 봅니다.

```console
$ git diff --stat feature feature-orig -- src
 src/app/validator.py | 18 +++++++++---------
 1 file changed, 9 insertions(+), 9 deletions(-)
```

validator.py 하나만 다릅니다. 사라진 커밋이 그 파일만 건드렸으니 예상과 맞습니다. 여기에 `-w`를 더하면 그 차이 중 실제 변경이 얼마인지 갈립니다.

```console
$ git diff -w --stat feature feature-orig -- src
 src/app/validator.py | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
```

두 결과의 실질 차이는 한 줄이고, 나머지 8줄은 공백입니다. **파일 목록만 보고 "많이 다르다"고 판단하기 전에 `-w`를 한 번 통과시키는 것**이 이 절의 요령입니다.

경로를 좁히는 것도 중요합니다. `-- src`를 빼면 `docs/`나 빌드 산출물 차이까지 섞여 들어와 판단이 흐려집니다.

## 10. 제목으로 세는 방법과 그 함정

리베이스는 SHA를 다시 쓰지만 커밋 제목은 보존합니다. 그래서 "무엇이 빠졌나"는 제목 집합의 차집합으로 셀 수 있습니다.

```console
$ git log --format='%s' master..feature      | sort > /tmp/a.txt
$ git log --format='%s' master..feature-orig | sort > /tmp/b.txt
$ comm -13 /tmp/a.txt /tmp/b.txt     # 원본에만 있는 것
코드 길이 5 허용
$ comm -23 /tmp/a.txt /tmp/b.txt     # 결과에만 있는 것
```

사라진 커밋을 이름으로 짚어 냈습니다. `comm`은 정렬된 두 파일을 비교해 3열(A만/B만/공통)로 나누고, `-13`은 1열과 3열을 숨겨 **B에만 있는 것**을 남깁니다. `-23`은 반대입니다. 그래서 두 방향을 다 봐야 "빠진 것"과 "새로 생긴 것"을 모두 잡습니다.

**함정은 이 방식이 중복을 접는다는 것입니다.** 같은 제목이 두 번 있어도 집합에는 하나로 들어갑니다.

```console
$ git commit --allow-empty -m "parser 헬퍼 3 추가"     # 같은 제목을 하나 더
$ git log --format='%s' master..feature | sort -u | wc -l
8
$ git rev-list --count master..feature
9
```

종류는 8인데 커밋은 9개입니다. 원본에 두 번 있던 제목이 결과에 한 번만 남은 경우를 이 방식으로는 못 잡습니다. 개수까지 보려면 제목별로 세야 합니다.

```console
$ git log --format='%s' master..feature | sort | uniq -c | awk '$1>1'
      2 parser 헬퍼 3 추가
```

이 함정을 보이려고 빈 커밋을 하나 넣었으니, 뒤 절을 이어서 재현하려면 되돌려 놓습니다.

```bash
git reset --hard HEAD~1
```

리베이스 전후로 이 표를 각각 만들어 비교하면 개수 차이까지 드러납니다. 갈래가 둘인 이력을 리베이스한 경우 같은 제목의 사본이 갈래마다 하나씩 생기므로, 이 함정이 실제로 걸립니다. 왜 사본이 생기는지는 [`git/rebase_replay_order.md`](rebase_replay_order.md)에서 다룹니다.

## 11. 선형인지 확인하기

리베이스 결과는 보통 선형이어야 합니다.

```console
$ git rev-list --count --merges master..feature
0
```

`0`이면 구간에 머지 커밋이 없다는 뜻입니다. `0`이 아니라면 두 가지 중 하나입니다. `--rebase-merges`로 머지 구조를 보존했거나, 리베이스가 끝난 뒤 누군가 머지를 넣었거나.

기본 리베이스는 머지 커밋을 그냥 버립니다. 그래서 원본에 머지가 있었다면 결과의 커밋 수가 줄어듭니다. 원본과 결과의 머지 수를 함께 세 두면 그 감소가 설명됩니다.

```bash
git rev-list --count --merges master..feature-orig    # 원본의 머지 수
git rev-list --count --merges master..feature         # 결과의 머지 수
```

## 12. 셸별 형태

bash 형태는 위 본문에 있습니다. 여기서는 다른 셸의 대응 형태를 둡니다.

PowerShell에서는 종료 코드가 `$LASTEXITCODE`이고, 파이프가 문자열이 아니라 객체를 넘깁니다. 아래는 PowerShell 7.6.5에서 실행해 확인한 것입니다.

```powershell
# 존재 확인 — 있으면 0, 없으면 128
git cat-file -e "feature:src/app/validator.py"; $LASTEXITCODE

# 구간에서 blob 해시 종류 세기
$first = (git rev-list master..feature | Select-Object -Last 1)
(git rev-list "$first^..feature" |
  ForEach-Object { git rev-parse "${_}:src/app/validator.py" } |
  Select-Object -Unique).Count

# 제목별 개수
git log --format='%s' master..feature | Group-Object |
  Where-Object Count -gt 1 | ForEach-Object { "{0}  {1}" -f $_.Count, $_.Name }

# 차이 유무
git diff --quiet feature feature-orig -- src; $LASTEXITCODE
```

`${_}:경로` 형태로 감싸는 이유는, 감싸지 않으면 PowerShell이 `$_:` 까지를 변수 이름으로 읽으려 하기 때문입니다.

Nushell 형태는 아래와 같습니다. **다만 이 환경에 nushell이 설치돼 있지 않아 실행으로 확인하지 못했습니다.** 버전에 따라 조정이 필요할 수 있습니다.

```nu
# 존재 확인 — 종료 코드를 값으로 받으려면 complete 를 쓴다
(do { git cat-file -e feature:src/app/validator.py } | complete).exit_code

# 구간에서 blob 해시 종류 세기
let first = (git rev-list master..feature | lines | last)
git rev-list $"($first)^..feature" | lines | each { |c|
  git rev-parse $"($c):src/app/validator.py"
} | uniq | length

# 제목별 개수
git log --format='%s' master..feature | lines
  | group-by { |x| $x }
  | items { |k, v| {제목: $k, 수: ($v | length)} }
  | where 수 > 1
```

Nushell에는 `comm`이 없으니 10장의 차집합은 컬렉션 연산으로 대신합니다. 이 형태도 미검증입니다.

```nu
let a = (git log --format='%s' master..feature | lines)
let b = (git log --format='%s' master..feature-orig | lines)
$b | where { |x| $x not-in $a }
```

`uniq`와 위 차집합은 모두 중복을 접으므로 10장의 함정이 그대로 적용됩니다.

## 13. 판정표

| 알고 싶은 것 | 명령 | 판정 |
|---|---|---|
| 구간 경계가 맞나 | `git rev-parse <first>^` 와 `master` 비교 | 두 값이 같으면 맞음 |
| 파일이 그 시점에 있나 | `git cat-file -e <c>:<path>` | 종료 코드 0 (없으면 128) |
| 내용이 같나 (diff 없이) | `git rev-parse <c>:<path>` | 해시 일치 |
| 구간에서 몇 번 바뀌었나 | 위 해시를 모아 종류 세기 | 1 = 안 바뀜, N+1 = N번 |
| 누가 그 파일을 건드렸나 | `git log <range> -- <path>` | 출력 비면 없음 |
| 실제 변경이 몇 줄인가 | `git show -w --stat <c>` | 공백 제외한 값 |
| 두 결과가 다른가 | `git diff --quiet A B -- <path>` | 0 = 차이 없음 |
| 무엇이 빠졌나 | 제목 집합 차집합 | 중복은 못 잡음 (10장) |
| 선형인가 | `git rev-list --count --merges <range>` | 0 = 선형 |

이 중 축은 세 번째 줄입니다. **blob 해시는 내용의 지문이므로, 구간의 커밋에서 해시를 모아 종류를 세는 것만으로 "이 구간에서 이 파일은 안 바뀌었다"를 증명할 수 있습니다.** diff를 구간 길이만큼 돌리는 것보다 빠르고, 눈으로 읽는 판단이 끼어들지 않습니다.

## 출처

- 이 문서의 모든 출력값, 해시, 종료 코드는 2장 스크립트로 만든 저장소에서 직접 실행해 얻었습니다. 환경은 `git version 2.54.0.windows.1`입니다.
- `git-cat-file.adoc:49` — `-e` 의 규정. "Exit with zero status if `<object>` exists and is a valid object." Git for Windows 설치본의 `mingw64/share/doc/git-doc/`.
- `revisions.adoc:217` — `<rev>:<path>` 는 "the blob or tree at the given path in the tree-ish object named by the part before the colon". 같은 디렉터리. `gitrevisions.adoc` 은 이 파일을 include 하는 껍데기입니다.
- `revisions.adoc:370` — `<rev1>..<rev2>` 는 "reachable from <rev2> but exclude those that are reachable from <rev1>". `:153` — 접미사 `^` 는 첫 부모. 같은 디렉터리.
- `diff-options.adoc:807` — `--exit-code` 는 "exits with 1 if there were differences and 0 means no differences". `:812` — `--quiet` 는 `--exit-code` 를 함의. 같은 디렉터리.
- PowerShell 형태는 PowerShell 7.6.5에서 실행해 확인했습니다.
- **미검증**: 12장의 nushell 형태는 이 환경에 nushell이 없어 실행하지 못했습니다. 원리는 같지만 문법은 버전에 따라 다를 수 있습니다.

## 관련 노트

- [`git/branch_merge_detection.md`](branch_merge_detection.md) — 도달성과 patch-id로 반영 여부를 판정. 이 문서의 blob 해시 비교가 그 3층에 해당
- [`git/rebase_replay_order.md`](rebase_replay_order.md) — 리베이스가 커밋을 재적용하는 순서, 같은 제목의 사본이 생기는 이유
- [`git/rebase_lost_commit_recovery.md`](rebase_lost_commit_recovery.md) — 이 문서에서 사라진 것으로 판정한 커밋을 되살리는 절차
- [`git/rev-parse.md`](rev-parse.md) — `rev-parse` 의 다른 용도
- [`git/diff.md`](diff.md) — diff 옵션 전반
- [`git/git_tree.md`](git_tree.md) — 객체 저장소와 tree/blob 구조
