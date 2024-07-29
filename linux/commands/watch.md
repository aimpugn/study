# watch

- [watch](#watch)
    - [사용례](#사용례)
        - [grep 제외하고 특정 프로세스 grep하여 보기](#grep-제외하고-특정-프로세스-grep하여-보기)
        - [`-n`은 인터벌](#-n은-인터벌)

## 사용례

### grep 제외하고 특정 프로세스 grep하여 보기

```bash
watch 'ps -ef | grep cake | grep -v grep'
```

### `-n`은 인터벌

```bash
watch -n 0.5 'ps -ef | grep cake | grep -v grep'
```
