# git commands

## git shortcuts of oh my zsh

- [oh my zsh git](https://kapeli.com/cheat_sheets/Oh-My-Zsh_Git.docset/Contents/Resources/Documents/index)
- [oh my zsh ig.plugin.zsh](https://github.com/ohmyzsh/ohmyzsh/blob/master/plugins/git/git.plugin.zsh)

## git restore

```shell
# restore uncommitted changes
git restore <파일명>
```

## git fetch

```shell
# update to date
git fetch --all
```

## git log

```shell
# gloga
# --all: `HEAD`와 함께, `refs/`의 모든 refs가 커밋처럼 커맨드라인에 리스팅 된다
git log --oneline --decorate --graph --all
```

## git status

```shell
# show changes of files
git status -s
```

## git add

```shell
# git add
# -A: 
git add --ignore-errors -A -f -- \
  shell/tls \
  .gitignore \
  javascript/query.js \
  shell/scp \
  git/fetch \
  git/log \
  git/commands.md \
  git/learngitbranching.md \
  git/rebase \
  shell/sudo \
  shell/tls_chain \
  docs/payments/tax/VAT.md
git commit -F \
  /private/var/folders/9x/8djp1ylj221bk02dp8zqsps00000gn/T/git-commit-msg-.txt -- 
  
# [snippets/220817 af32326] snippet: 빌드 코드 수정
#  12 files changed, 566 insertions(+), 3 deletions(-)

# --progress: standard error 스트림에 진행 상태가 보고 된다. 
# --porcelain: machine-readable output 생성. 각 ref에 대한 output status line이 탭으로 구별되며 stdout으로 전달된다.
# --set-upstream: 
#   ㄴ 최신이거나 성공적으로 푸시된 모든 브랜치에 대해서, upstream(tracking) 참조를 추가한다.
#   ㄴ 이 upstream 참조는 인자 없는 `git-pull`과 다른 커맨드에서 사용된다
git push --progress --porcelain origin refs/heads/snippets/220817:refs/heads/snippets/220817 --set-upstream

# Total 29 (delta 7), reused 0 (delta 0), pack-reused 0
# remote: Resolving deltas:   0% (0/7)        
# remote: Resolving deltas:  14% (1/7)        
# remote: Resolving deltas:  28% (2/7)        
# remote: Resolving deltas:  42% (3/7)        
# remote: Resolving deltas:  57% (4/7)        
# remote: Resolving deltas:  71% (5/7)        
# remote: Resolving deltas:  85% (6/7)        
# remote: Resolving deltas: 100% (7/7)        
# remote: Resolving deltas: 100% (7/7), completed with 4 local objects.        
# remote: 
# remote: Create a pull request for 'snippets/220817' on GitHub by visiting:        
# remote:      https://github.com/aimpugn/snippets/pull/new/snippets/220817        
# remote: 
# To https://github.com/aimpugn/snippets.git
# * refs/heads/snippets/220817:refs/heads/snippets/220817 [new branch]
# Branch 'snippets/220817' set up to track remote branch 'snippets/220817' from 'origin'.
# Done
```

## rerere

- git rebase 시에 충돌 해결된 내역을 기억하고 있다가 나중에 자동으로 적용해주는 기능, 이라고 한다

## grbc

## grbi

## git bisect

## git 로컬 태그 prune

```shell
git fetch --prune --prune-tags
```

```shell
git fetch --prune origin "+refs/tags/*:refs/tags/*"
```
