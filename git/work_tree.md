# Git Work Tree

- [Git Work Tree](#git-work-tree)
    - [Work Tree?](#work-tree)
    - [`GIT_WORK_TREE` 옵션](#git_work_tree-옵션)
        - [`/home/user/some_dir/myrepo.git`의 구조?](#homeusersome_dirmyrepogit의-구조)
    - [기타](#기타)

## Work Tree?

## `GIT_WORK_TREE` 옵션

```shell
cd /home/user/some_dir/myrepo.git



sudo -Hu www-data GIT_WORK_TREE=/path/to/source/cloned/dir git checkout -f main
```

### `/home/user/some_dir/myrepo.git`의 구조?

```tree
.
├── FETCH_HEAD
├── HEAD
├── branches
├── config
├── description
├── hooks
│   ├── applypatch-msg.sample
│   ├── commit-msg.sample
│   ├── post-update.sample
│   ├── pre-applypatch.sample
│   ├── pre-commit.sample
│   ├── pre-push.sample
│   ├── pre-rebase.sample
│   ├── prepare-commit-msg.sample
│   └── update.sample
├── index
├── info
│   ├── exclude
│   └── refs
├── logs
│   └── HEAD
├── objects
│   ├── 24
│   │   └── d34a03c72cf5bc21cca8aa8831655d5f109ea2
│   ................... 생략 ........................
│   ├── fe
│   │   └── cdeede1a3ec586dea216955358a4f89549fbc4
│   ├── info
│   │   └── packs
│   └── pack
│       ├── pack-5376e7dffa6795b3a36410dbc0b14e03d8ffb76e.idx
│       ├── pack-5376e7dffa6795b3a36410dbc0b14e03d8ffb76e.pack
│       ├── pack-d34e74a4568e49c354503723131d0b53306233c3.idx
│       └── pack-d34e74a4568e49c354503723131d0b53306233c3.pack
├── packed-refs
└── refs
    ├── heads
    │   └── test-phpseclib
    │       ├── polyfill56
    │       └── with-polyfill-php56
    ├── pull
    └── tags
```

## 기타

- [What is GIT_WORK_TREE, why have I never needed to set this ENV var, why now?](https://stackoverflow.com/questions/5283262/what-is-git-work-tree-why-have-i-never-needed-to-set-this-env-var-why-now)
