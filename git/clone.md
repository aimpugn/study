# git clone

## 특정 브랜치 클론하기

- [How do I clone a specific Git branch? [duplicate]](https://stackoverflow.com/a/4568323)

```shell
# 모든 브랜치 가져와서 해당 브랜치로 체크아웃

git clone -b <branch> <remote_repo>
git clone --branch <branchname> url

# Example:
# git clone -b my-branch git@github.com:user/myproject.git
```

```shell
git clone --single-branch --branch <branchname> <remote-repo>
# Example:
# git clone -b opencv-2.4 --single-branch https://github.com/Itseez/opencv.git
```
