# tag

- [tag](#tag)
    - [git fetch tag](#git-fetch-tag)
    - [show tags](#show-tags)
    - [tag sort by refname in reverse](#tag-sort-by-refname-in-reverse)
    - [git tag prune](#git-tag-prune)
    - [delete tag](#delete-tag)

## git fetch tag

```shell
git fetch --all --tags
```

## show tags

```shell
git tag
```

## tag sort by refname in reverse

```shell
git tag --sort=-v:refname
```

```shell
git tag -l --sort=-version:refname <pattern>
```

## [git tag prune](https://stackoverflow.com/a/54297675)

```shell
git fetch --prune origin "+refs/tags/*:refs/tags/*"

git fetch --prune --prune-tags
```

## delete tag

- [How To Delete Local and Remote Tags on Git](https://devconnected.com/how-to-delete-local-and-remote-tags-on-git/)

```bash
# 로컬에서 태그를 지운다
g tag -d <TAG_NAME>
```

```bash
# origin 에서 태그를 지운다
g push --delete origin <TAG_NAME>
```

```bash
# 삭제 반영
git fetch --prune --prune-tags
```

```bash
# :refs/tags/<tag>
$ git push origin :refs/tags/v1.0

To https://github.com/SCHKN/repo.git
 - [deleted]         v1.0
```
