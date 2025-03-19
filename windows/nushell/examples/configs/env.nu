# env.nu
#
# Installed by:
# version = "0.102.0"
#
# Previously, environment variables were typically configured in `env.nu`.
# In general, most configuration can and should be performed in `config.nu`
# or one of the autoload directories.
#
# This file is generated for backwards compatibility for now.
# It is loaded before config.nu and login.nu
#
# See https://www.nushell.sh/book/configuration.html
#
# Also see `help config env` for more options.
#
# You can remove these comments if you want or leave
# them for future reference.

# Path 변수에 임의의 경로 추가하기
# `echo $env.PATH` 실행하여 적용됐는지 확인할 수 있습니다.
let home = $env.USERPROFILE # C:\Users\username
let toolsBinDir = [ $home "dev" "tools" "bin" ] | path join
$env.PATH = ($env.PATH | append $toolsBinDir)

# Oh My Zsh 스타일 Git alias
alias g = git
alias ga = git add
alias gaa = git add --all
alias gam = git am
alias gama = git am --abort
alias gamc = git am --continue
alias gams = git am --skip
alias gamscp = git am --show-current-patch
alias gap = git apply
alias gapa = git add --patch
alias gapt = git apply --3way
alias gau = git add --update
alias gav = git add --verbose
alias gb = git branch
alias gbD = git branch --delete --force
alias gba = git branch --all
alias gbd = git branch --delete
alias gbl = git blame -w
alias gbm = git branch --move
alias gbnm = git branch --no-merged
alias gbr = git branch --remote
alias gbs = git bisect
alias gbsb = git bisect bad
alias gbsg = git bisect good
alias gbsn = git bisect new
alias gbso = git bisect old
alias gbsr = git bisect reset
alias gbss = git bisect start
alias gc = git commit --verbose
alias gcB = git checkout -B
alias gca = git commit --verbose --all
alias gcam = git commit --all --message
alias gcas = git commit --all --signoff
alias gcasm = git commit --all --signoff --message
alias gcb = git checkout -b
alias gcf = git config --list
alias gcl = git clone --recurse-submodules
alias gclean = git clean --interactive -d
alias gclf = git clone --recursive --shallow-submodules --filter=blob:none --also-filter-submodules
alias gcmsg = git commit --message
alias gcn = git commit --verbose --no-edit
alias gco = git checkout
alias gcor = git checkout --recurse-submodules
alias gcount = git shortlog --summary --numbered
alias gcp = git cherry-pick
alias gcpa = git cherry-pick --abort
alias gcpc = git cherry-pick --continue
alias gcs = git commit --gpg-sign
alias gcsm = git commit --signoff --message
alias gcss = git commit --gpg-sign --signoff
alias gcssm = git commit --gpg-sign --signoff --message
alias gd = git diff
alias gdca = git diff --cached
alias gdcw = git diff --cached --word-diff
alias gds = git diff --staged
alias gdt = git diff-tree --no-commit-id --name-only -r
alias gdup = git diff @{upstream}
alias gdw = git diff --word-diff
alias gf = git fetch
alias gfa = git fetch --all --tags --prune --jobs=10
alias gfo = git fetch origin
alias gg = git gui citool
alias gga = git gui citool --amend
alias ghh = git help
alias gignore = git update-index --assume-unchanged
alias gl = git pull
alias glg = git log --stat
alias glgg = git log --graph
alias glgga = git log --graph --decorate --all
alias glgm = git log --graph --max-count=10
alias glgp = git log --stat --patch
alias glo = git log --oneline --decorate
alias glod = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ad) %C(bold blue)<%an>%Creset"
alias glods = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ad) %C(bold blue)<%an>%Creset" --date=short
alias glog = git log --oneline --decorate --graph
alias gloga = git log --oneline --decorate --graph --all
alias glol = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset"
alias glola = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset" --all
alias glols = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset" --stat
alias gm = git merge
alias gma = git merge --abort
alias gmc = git merge --continue
alias gmff = git merge --ff-only
alias gms = git merge --squash
alias gmtl = git mergetool --no-prompt
alias gmtlvim = git mergetool --no-prompt --tool=vimdiff
alias gp = git push
alias gpd = git push --dry-run
alias gpf = git push --force-with-lease --force-if-includes
alias gpoat = git push origin --all and git push origin --tags
alias gpod = git push origin --delete
alias gpr = git pull --rebase
alias gpra = git pull --rebase --autostash
alias gprav = git pull --rebase --autostash -v
alias gpristine = git reset --hard and git clean --force -dfx
alias gprv = git pull --rebase -v
alias gpu = git push upstream
alias gpv = git push --verbose
alias gr = git remote
alias gra = git remote add
alias grb = git rebase
alias grba = git rebase --abort
alias grbc = git rebase --continue
alias grbi = git rebase --interactive
alias grbo = git rebase --onto
alias grbs = git rebase --skip
alias grev = git revert
alias greva = git revert --abort
alias grevc = git revert --continue
alias grf = git reflog
alias grh = git reset
alias grhh = git reset --hard
alias grhk = git reset --keep
alias grhs = git reset --soft
alias grm = git rm
alias grmc = git rm --cached
alias grmv = git remote rename
alias grrm = git remote remove
alias grs = git restore
alias grset = git remote set-url
alias grss = git restore --source
alias grst = git restore --staged
alias gru = git reset --
alias grup = git remote update
alias grv = git remote --verbose
alias gsb = git status --short --branch
alias gsd = git svn dcommit
alias gsh = git show
alias gsi = git submodule init
alias gsps = git show --pretty=short --show-signature
alias gsr = git svn rebase
alias gss = git status --short
alias gst = git status
alias gsta = git stash push
alias gstaa = git stash apply
alias gstall = git stash --all
alias gstc = git stash clear
alias gstd = git stash drop
alias gstl = git stash list
alias gstp = git stash pop
alias gsts = git stash show --patch
alias gsu = git submodule update
alias gsw = git switch
alias gswc = git switch --create
alias gta = git tag --annotate
alias gts = git tag --sign
alias gtv = git tag | sort -V
alias gunignore = git update-index --no-assume-unchanged
alias gwch = git whatchanged -p --abbrev-commit --pretty=medium
alias gwipe = git reset --hard and git clean --force -df
alias gwt = git worktree
alias gwta = git worktree add
alias gwtls = git worktree list
alias gwtmv = git worktree move
alias gwtrm = git worktree remove
